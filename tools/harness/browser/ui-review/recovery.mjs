import { stopDiagnosticProcesses } from "./diagnostic-processes.mjs";
import path from "node:path";
import { closeSync, constants, existsSync, openSync, readdirSync } from "node:fs";
import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";
import { borrowSuiteRuntime, closeSuiteRuntimeRoot, scanRetainedRoot } from "../../runtime/suite-runtime.mjs";
import { processIdentityAlive } from "../../runtime/host-admission.mjs";
import { atomicLocalFile, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { recoverReviewPreparation } from "../review-preparation.mjs";
import { emptyCounts, failureRecord, result, ReviewFailure, validate } from "./contract.mjs";
import { readLocator, terminalResult, unregisterSession, updateSessionRecord } from "./session-files.mjs";
import { finishTerminal } from "./terminal.mjs";
import { ownedProcess, purgeReviewDetail, recordResource, recoveryResources, stopOwnedProcess } from "./ownership.mjs";
import { repoRoot } from "./policy.mjs";

// Kernel ownership serializes independent stop callers and survives death.
// Private context travels over IPC, never argv or captured child output.
export async function recoverSession(record) {
  const lock = `${record.socket}.recovery-lock`;
  try { atomicLocalFile(lock, ""); } catch (error) { if (error.code !== "EEXIST") throw error; }
  readLocalFile(lock, { maximum: 0 });
  const fd = openSync(lock, constants.O_RDONLY | constants.O_NOFOLLOW);
  const transport = Object.fromEntries(["HOME", "DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "XDG_RUNTIME_DIR"].filter((name) => process.env[name] !== undefined).map((name) => [name, process.env[name]]));
  const child = spawn("/usr/bin/flock", ["-x", "-w", "5", "/proc/self/fd/4", process.execPath, fileURLToPath(import.meta.url), "--recover"], { detached: true, stdio: ["ignore", "ignore", "ignore", "ipc", fd], env: { ...transport, PATH: "/usr/bin:/bin", LANG: "C.UTF-8", TZ: "UTC" } });
  closeSync(fd);
  return new Promise((resolve, reject) => {
    let response, expired = false;
    const proof = ownedProcess(child.pid);
    const timer = setTimeout(() => { expired = true; void stopOwnedProcess(proof); }, 330000);
    child.once("message", (message) => { response = message; });
    child.once("error", () => { clearTimeout(timer); reject(new ReviewFailure("cleanup_failed")); });
    child.once("exit", () => {
      clearTimeout(timer);
      try { if (!response || expired) throw new ReviewFailure(expired ? "cleanup_failed" : "capacity_exceeded"); resolve(validate("command_result", response)); }
      catch (error) { reject(error); }
    });
    child.send(record);
  });
}

async function recover(record) {
  if (processIdentityAlive(record.process)) throw new ReviewFailure("session_mismatch");
  const identity = readLocator(record.locator);
  if (identity.locator.session_id !== record.session_id || identity.runRoot !== path.dirname(path.dirname(record.locator))) throw new ReviewFailure("session_mismatch");
  const prior = ["closed", "failed"].includes(identity.locator.state) ? terminalResult(identity, "ui-review", (receipt, ref) => result("ui-review", { session_id: record.session_id, state: receipt.state, status: receipt.status, exit_code: receipt.exit_code, failures: receipt.failures, receipt: ref })) : null;
  const failures = [failureRecord(new ReviewFailure("session_lost"))]; let cleaned = true;
  const attempt = async (fn) => { try { await fn(); } catch { cleaned = false; failures.push(failureRecord(new ReviewFailure("cleanup_failed"))); } };
  if (existsSync(record.runtime.root)) {
    const runtime = borrowSuiteRuntime({ repoRoot, runRoot: identity.runRoot, environment: { CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: record.runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: record.runtime.lease_id, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: record.runtime.run_id } });
    for (const resource of recoveryResources(runtime).filter((entry) => entry.kind === "diagnostic_scope")) await attempt(async () => { await stopDiagnosticProcesses(resource.target); recordResource(runtime, { ...resource, state: "released" }); });
    const processes = recoveryResources(runtime).filter((entry) => entry.kind.endsWith("_process"));
    for (const resource of processes) await attempt(async () => { await stopOwnedProcess(resource.target); recordResource(runtime, { ...resource, state: "released" }); });
    await attempt(() => recoverReviewPreparation({ runtime, resources: recoveryResources(runtime), onReleased: (resource) => recordResource(runtime, { ...resource, state: "released" }) }));
    await attempt(() => purgeReviewDetail(runtime));
    if (cleaned) await attempt(() => { updateSessionRecord(record, { resources_released: true }); closeSuiteRuntimeRoot({ root: record.runtime.root, expectedLeaseID: record.runtime.lease_id }); });
  } else if (record.resources_released !== true) throw new ReviewFailure("unsafe_artifact");
  if (prior) {
    if (cleaned) {
      if ((await scanRetainedRoot(identity.runRoot, { removeUnsafe: true })).status !== "pass") throw new ReviewFailure("unsafe_artifact");
      unregisterSession(record);
    }
    return prior;
  }
  const counts = emptyCounts();
  try {
    const directory = path.join(identity.runRoot, "ui-review/operations");
    for (const operation of readdirSync(directory)) {
      if (!/^[1-9][0-9]*$/u.test(operation)) throw new ReviewFailure("invalid_artifact");
      const receipt = validate("receipt", JSON.parse(readLocalFile(path.join(directory, operation, "receipt.json"))));
      if (receipt.session_id !== record.session_id) throw new ReviewFailure("invalid_artifact");
      if (receipt.status === "ok") for (const key of Object.keys(counts)) counts[key] += receipt.counts[key];
    }
  } catch (error) { if (error.code !== "ENOENT") failures.push(failureRecord(new ReviewFailure("invalid_artifact"))); }
  const terminal = await finishTerminal({ record, locator: identity.locator, runRoot: identity.runRoot, counts, failures, exitCode: 3, duration: Math.floor(Number(process.hrtime.bigint() - BigInt(record.started_tick_ns)) / 1e6), cleanupFailed: !cleaned });
  return result("ui-review", { session_id: record.session_id, state: terminal.state, status: "error", exit_code: terminal.exitCode, failures: terminal.failures, receipt: terminal.receipt });
}

if (process.argv[2] === "--recover") {
  const record = await new Promise((resolve) => process.once("message", resolve));
  let value;
  try { value = await recover(record); }
  catch (error) { value = result("ui-review", { status: "error", exit_code: error.exitCode ?? 11, failures: [failureRecord(error)] }); }
  if (process.connected) process.send(value, () => process.disconnect());
}
