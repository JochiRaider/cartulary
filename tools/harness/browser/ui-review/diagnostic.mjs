import { openSync, closeSync, constants, readFileSync, readdirSync, lstatSync } from "node:fs";
import { spawn } from "../../workspace/child-process.mjs";
import { fileURLToPath } from "node:url";
import path from "node:path";
import { chromium } from "playwright";
import { acquireHostAdmission } from "../../runtime/host-admission.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile, removePrivateTree } from "../../runtime/secure-local-files.mjs";
import { resolvePlaywrightPackages } from "../../readiness/playwright-packages.mjs";
import { parseStrictJSON } from "../../contract/index.mjs";
import { diagnosticProcessScope, diagnosticProcesses, stopDiagnosticProcesses, bindDiagnosticProcess } from "./diagnostic-processes.mjs";
import { boundedCleanup, ownedProcess } from "./ownership.mjs";
import { workClaims } from "./work-policy.mjs";
import { ReviewFailure, limits, validate } from "./contract.mjs";
import { repoRoot } from "./policy.mjs";

export const diagnosticScratchLimit = 16 * 1024 ** 2;
export function checkDiagnosticScratch(directory) {
  let bytes = 0, entries = 0;
  const visit = (current, depth) => {
    if (depth > 16) throw new ReviewFailure("observation_limit");
    for (const entry of readdirSync(current)) {
      if (++entries > 256) throw new ReviewFailure("observation_limit");
      const file = path.join(current, entry), stat = lstatSync(file);
      if (stat.isSymbolicLink()) throw new ReviewFailure("unsafe_artifact");
      if (stat.isDirectory()) visit(file, depth + 1);
      else if (stat.isFile() && (bytes += stat.size) > diagnosticScratchLimit) throw new ReviewFailure("observation_limit");
    }
  };
  try { visit(directory, 0); } catch (error) { if (error.code !== "ENOENT") throw error; }
}

export function diagnosticEnvironment(directory, token) {
  return { PATH: `${path.dirname(process.execPath)}:/usr/bin:/bin`, HOME: path.join(directory, "home"),
    XDG_CACHE_HOME: path.join(directory, "cache"), XDG_CONFIG_HOME: path.join(directory, "config"),
    XDG_DATA_HOME: path.join(directory, "data"), TMPDIR: path.join(directory, "tmp"),
    LANG: "C.UTF-8", TZ: "UTC", CARTULARY_DIAGNOSTIC_TOKEN: token };
}
export function parseDiagnosticOutput(bytes) {
  if (bytes.length > limits.snapshot) throw new ReviewFailure("observation_limit");
  try {
    const result = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
    if (!result || typeof result !== "object" || Array.isArray(result) || result.isError || result.result?.isError) throw new Error("diagnostic failed");
    return result;
  } catch (cause) { throw new ReviewFailure("target_unavailable", { cause }); }
}
async function pageState(owner) {
  if (owner.browser.contexts().length !== 1 || owner.context.pages().length !== 1 || owner.context.pages()[0] !== owner.page) throw new ReviewFailure("target_unavailable");
  return { epoch: owner.epoch, generation: owner.generation, url: owner.page.url(), viewport: owner.page.viewportSize(),
    position: await owner.page.evaluate(() => ({ x: scrollX, y: scrollY, zoom: document.documentElement.style.zoom,
      scroll: [...document.querySelectorAll("*")].filter((node) => node.scrollLeft || node.scrollTop).map((node) => [node.tagName, node.id, node.scrollLeft, node.scrollTop]) })) };
}
async function targetID(context, page) {
  const cdp = await context.newCDPSession(page);
  try { return (await cdp.send("Target.getTargetInfo")).targetInfo.targetId; }
  finally { await cdp.detach(); }
}

export async function diagnosticSnapshot({ owner, runtime, operationID, signal, capacities, parentLease, onResource, spawnProcess = spawn }) {
  const scope = diagnosticProcessScope();
  const directory = runtime.privatePath("artifacts", "diagnostics", String(operationID));
  const environment = diagnosticEnvironment(directory, scope.token);
  const controller = new AbortController();
  const combined = AbortSignal.any([signal, controller.signal]);
  const timer = setTimeout(() => controller.abort(new ReviewFailure("operation_expired")), limits.operation);
  const resourceTimer = setInterval(() => {
    try {
      const processes = diagnosticProcesses(scope);
      const memory = processes.reduce((sum, proof) => {
        try { return sum + Number(readFileSync(`/proc/${proof.pid}/statm`, "utf8").split(" ")[1]) * 4096; }
        catch (error) { if (["ENOENT", "ESRCH"].includes(error.code)) return sum; throw error; }
      }, 0);
      if (processes.length > workClaims.diagnostic.process || memory > workClaims.diagnostic.memory_mb * 1048576) controller.abort(new ReviewFailure("capacity_exceeded"));
      checkDiagnosticScratch(directory);
    } catch (error) { controller.abort(error instanceof ReviewFailure ? error : new ReviewFailure("target_unavailable")); }
  }, 50);
  let stage = "admission";
  let temporaryFD, lease, bound = false, registered = false, reaped = false, failure, attached = false, endpoint, focus;
  const oldCache = process.env.XDG_CACHE_HOME;
  const session = "cartulary-diagnostic";
  let executable;
  let lastCommand = null;
  const proofs = new Map();
  const config = path.join(directory, "diagnostic.config.json");
  const snapshotFile = path.join(directory, "snapshot.yml");
  const run = async (command, cancellation = combined) => {
    cancellation.throwIfAborted();
    const args = command === "attach" ? ["attach", `--endpoint=${endpoint}`, `--config=${config}`]
      : command === "snapshot" ? ["snapshot", `--filename=${snapshotFile}`] : command === "detach" ? ["detach"] : null;
    if (!args) throw new ReviewFailure("invalid_request");
    const child = spawnProcess(process.execPath, [fileURLToPath(new URL("./diagnostic-cli.cjs", import.meta.url)), executable, "cli", `--session=${session}`, "--json", ...args],
      { cwd: directory, env: environment, detached: true, stdio: ["ignore", "pipe", "pipe"] });
    let size = 0, overflow = false; const stdout = [], stderr = [];
    const collect = (target) => (bytes) => { size += bytes.length; if (size > limits.snapshot) { overflow = true; abort(); } else target.push(bytes); };
    child.stdout.on("data", collect(stdout)); child.stderr.on("data", collect(stderr));
    const exit = new Promise((resolve, reject) => {
      let spawnError;
      child.once("error", (error) => { spawnError = error; });
      child.once("close", (status) => { if (spawnError) reject(spawnError); else resolve(status); });
    });
    // Publication can fail before awaiting completion. Observe rejection now;
    // report spawn errors after close without replacing an earlier failure.
    exit.catch(() => {});
    let killTimer;
    const abort = () => { child.kill("SIGTERM"); killTimer = setTimeout(() => child.kill("SIGKILL"), 250); };
    cancellation.addEventListener("abort", abort, { once: true });
    let proof;
    try {
      if (!Number.isSafeInteger(child.pid) || child.pid < 2) {
        await exit;
        throw new ReviewFailure("target_unavailable");
      }
      try { proof = ownedProcess(child.pid); }
      catch (error) { if (!["ENOENT", "ESRCH"].includes(error.code)) throw error; }
      if (proof) {
        proofs.set(proof.pid, proof); await onResource({ kind: "helper_process", target: proof });
        await bindDiagnosticProcess(lease, proof);
      }
      const status = await exit;
      lastCommand = { command, status, stdout: Buffer.concat(stdout).subarray(0, 4096).toString("utf8"), stderr: Buffer.concat(stderr).subarray(0, 4096).toString("utf8") };
      atomicLocalFile(path.join(directory, `${command}.stdout`), Buffer.concat(stdout));
      atomicLocalFile(path.join(directory, `${command}.stderr`), Buffer.concat(stderr));
      cancellation.throwIfAborted();
      if (overflow) throw new ReviewFailure("observation_limit");
      if (status !== 0) throw new ReviewFailure("target_unavailable");
      return parseDiagnosticOutput(Buffer.concat(stdout));
    } finally { clearTimeout(killTimer); cancellation.removeEventListener("abort", abort); if (proof && child.exitCode !== null) await onResource({ kind: "helper_process", target: proof, state: "released" }); }
  };
  try {
    combined.throwIfAborted();
    executable = path.join(resolvePlaywrightPackages(repoRoot).playwrightPath, "cli.js");
    try { lease = await acquireHostAdmission({ claims: workClaims.diagnostic, capacities, parent: parentLease, signal: combined, timeoutMs: limits.lock }); }
    catch (cause) { combined.throwIfAborted(); throw new ReviewFailure("capacity_exceeded", { cause }); }
    privateDirectory(directory);
    for (const name of ["home", "cache", "config", "data", "tmp", "output"]) privateDirectory(path.join(directory, name));
    // A directory FD keeps Unix socket names below Linux's 108-byte limit even
    // when the Make-owned runtime root is long. Bytes remain in private storage.
    temporaryFD = openSync(environment.TMPDIR, constants.O_RDONLY | constants.O_DIRECTORY | constants.O_NOFOLLOW);
    environment.TMPDIR = `/proc/${process.pid}/fd/${temporaryFD}`;
    atomicLocalFile(config, JSON.stringify({ browser: { browserName: "chromium" }, outputDir: path.join(directory, "output"), outputMaxSize: diagnosticScratchLimit,
      saveSession: false, imageResponses: "omit", allowUnrestrictedFileAccess: false, snapshot: { mode: "full", boxes: false },
      timeouts: { action: limits.action, navigation: limits.action }, console: { level: "error" } }));
    await onResource({ kind: "diagnostic_scope", target: scope }); registered = true;
    stage = "state";
    const before = await pageState(owner);
    focus = await owner.page.evaluateHandle(() => document.activeElement);
    const expectedTarget = await targetID(owner.context, owner.page);
    // Upstream caches its registry root on first use. Keep a stable private
    // session root, separate from disposable per-operation CLI scratch.
    const registryRoot = runtime.privatePath("artifacts", "diagnostic-registry");
    privateDirectory(registryRoot);
    process.env.XDG_CACHE_HOME = registryRoot;
    stage = "binding";
    bound = true;
    ({ endpoint } = await owner.browser.bind(session, { host: "127.0.0.1", port: 0, workspaceDir: directory }));
    if (!/^ws:\/\/127\.0\.0\.1:\d+\/[A-Za-z0-9_-]+$/u.test(endpoint)) throw new ReviewFailure("target_unavailable");
    stage = "probe";
    const probe = await chromium.connect(endpoint, { timeout: limits.action });
    try {
      const contexts = probe.contexts();
      if (contexts.length !== 1 || contexts[0].pages().length !== 1 || await targetID(contexts[0], contexts[0].pages()[0]) !== expectedTarget) throw new ReviewFailure("target_unavailable");
    } finally { await probe.close(); }
    stage = "attach";
    const attachment = await run("attach"); attached = true;
    if (attachment.session !== session || attachment.endpoint !== endpoint || !diagnosticProcesses(scope).some((proof) => proof.pid === attachment.pid)) throw new ReviewFailure("target_unavailable");
    for (const proof of diagnosticProcesses(scope)) { proofs.set(proof.pid, proof); await onResource({ kind: "helper_process", target: proof }); await bindDiagnosticProcess(lease, proof); }
    stage = "snapshot";
    const result = await run("snapshot");
    if (typeof result.snapshot?.file !== "string" || path.resolve(directory, result.snapshot.file) !== snapshotFile) throw new ReviewFailure("target_unavailable");
    const snapshot = new TextDecoder("utf-8", { fatal: true }).decode(readLocalFile(snapshotFile, { maximum: limits.snapshot }));
    if (!snapshot.trim()) throw new ReviewFailure("target_unavailable");
    stage = "detach";
    const detached = await run("detach"); attached = false;
    if (detached.session !== session || detached.status !== "detached") throw new ReviewFailure("cleanup_failed");
    stage = "invariants";
    if (JSON.stringify(await pageState(owner)) !== JSON.stringify(before) || !await focus.evaluate((node) => node === document.activeElement)) throw new ReviewFailure("unstable_capture");
    owner.checkFault(); combined.throwIfAborted();
    return validate("observations", { ...await owner.observe([]), accessibility_snapshot: snapshot });
  } catch (error) {
    failure = error instanceof ReviewFailure ? error : new ReviewFailure("target_unavailable", { cause: error }); failure.diagnosticStage = stage;
    const cause = error.cause ?? error;
    const detail = { operation_id: operationID, stage, diagnostic_code: failure.diagnostic,
      cause_name: ["Error", "TimeoutError", "AbortError", "SyntaxError"].includes(cause.name) ? cause.name : "Other",
      cause_code: ["EAGAIN", "ENOENT", "ESRCH", "EACCES", "EPIPE", "ECONNRESET", "ECONNREFUSED", "ETIMEDOUT"].includes(cause.code) ? cause.code : null,
      message: String(cause.message ?? "").slice(0, 2048), last_command: lastCommand };
    let bytes = JSON.stringify(detail);
    if (Buffer.byteLength(bytes) > limits.request) bytes = JSON.stringify({ ...detail, message: "", last_command: null });
    try { atomicLocalFile(runtime.privatePath("artifacts", "diagnostic-failure.json"), bytes, { replace: true }); }
    catch (reportError) { failure.privateDiagnosticFailure = reportError; }
    throw failure;
  }
  finally {
    clearTimeout(timer); clearInterval(resourceTimer); const errors = [];
    const attempt = async (release) => { try { await release(); } catch (error) { errors.push(error); } };
    if (attached) await attempt(() => boundedCleanup(() => run("detach", AbortSignal.timeout(3000)), 4000));
    if (registered) await attempt(async () => { await stopDiagnosticProcesses(scope); reaped = true; for (const proof of proofs.values()) await onResource({ kind: "helper_process", target: proof, state: "released" }); await onResource({ kind: "diagnostic_scope", target: scope, state: "released" }); });
    if (bound) await attempt(() => boundedCleanup(() => owner.browser.unbind(), 5000));
    if (oldCache === undefined) delete process.env.XDG_CACHE_HOME; else process.env.XDG_CACHE_HOME = oldCache;
    await attempt(() => focus?.dispose());
    if (temporaryFD !== undefined) await attempt(() => closeSync(temporaryFD));
    if (!registered || reaped) { await attempt(() => removePrivateTree(directory)); await attempt(() => lease?.release()); }
    if (errors.length) {
      if (failure) (failure.cleanupFailures ??= []).push(...errors);
      else throw Object.assign(new ReviewFailure("cleanup_failed"), { cleanupFailures: errors });
    }
  }
}
