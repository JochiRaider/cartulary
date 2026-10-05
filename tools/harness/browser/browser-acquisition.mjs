import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { spawn, spawnSync } from "../workspace/child-process.mjs";
import { createHash, randomUUID } from "node:crypto";
import { existsSync, mkdirSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { createSecureWriteStream, parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { CommandFailure, publishCommandFailure, readCommandFailure } from "../runtime/command-failure.mjs";
import { ownedProcess, stopOwnedProcess } from "../runtime/owned-process.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile, removePrivateFile, removePrivateTree } from "../runtime/secure-local-files.mjs";

const schemaID = "cartulary.browser_acquisition.v1";
const bytes = (value) => `${JSON.stringify(value)}\n`;
const digest = (value) => createHash("sha256").update(value).digest("hex");
const read = (file) => parseStrictJSON(readLocalFile(file, { maximum: 1048576 }).toString());
const invalid = (cause) => new CommandFailure("invalid browser acquisition ownership", { failure_class: "artifact", failure_reason: "artifact_error" }, { cause });

function inside(parent, child) {
  const relative = path.relative(parent, child);
  if (!relative || relative.startsWith("..") || path.isAbsolute(relative)) throw invalid();
}

function suiteProof(bytes, runID) {
  try {
    const suite = parseStrictJSON(bytes.toString());
    validateSchemaSync("cartulary.test_services.lease.v1", suite);
    if (suite.run_id !== runID || !path.isAbsolute(suite.result_root) ||
        suite.run_root !== path.join(suite.result_root, suite.run_id)) throw invalid();
    return suite;
  } catch (cause) { throw invalid(cause); }
}

// The immutable attempt precedes launch. Independent resource records avoid a
// shared read/modify/write journal between the producer and detached services.
export function createBrowserAcquisition({ runtime, sessionRoot, leaseFile, suiteLease }) {
  inside(runtime.root, sessionRoot);
  inside(sessionRoot, leaseFile);
  inside(runtime.root, suiteLease);
  const suiteBytes = readLocalFile(suiteLease, { maximum: 1048576 });
  const suite = suiteProof(suiteBytes, runtime.runID);
  const file = path.join(sessionRoot, "acquisition", "owner.json");
  const value = { schema_id: schemaID, attempt_id: randomUUID(), run_id: runtime.runID,
    runtime_lease_id: runtime.leaseID, runtime_root: runtime.root, session_root: sessionRoot,
    ready_lease: leaseFile, suite_lease: suiteLease, suite_digest: digest(suiteBytes), suite_id: suite.suite_id };
  validateSchemaSync(schemaID, value);
  atomicLocalFile(file, bytes(value));
  return file;
}

export function readBrowserAcquisition(file, runtime) {
  try {
    const value = read(file);
    validateSchemaSync(schemaID, value);
    if (path.resolve(file) !== path.join(value.session_root, "acquisition", "owner.json")) throw invalid();
    inside(value.runtime_root, value.session_root);
    inside(value.session_root, value.ready_lease);
    inside(value.runtime_root, value.suite_lease);
    const owner = read(path.join(value.runtime_root, "runtime-owner.json"));
    if (owner.run_id !== value.run_id || owner.lease_id !== value.runtime_lease_id ||
        (runtime && (runtime.root !== value.runtime_root || runtime.runID !== value.run_id || runtime.leaseID !== value.runtime_lease_id))) throw invalid();
    return value;
  } catch (cause) { throw invalid(cause); }
}

function resourceFile(file, name) { return path.join(path.dirname(file), name); }
function receipt(file) {
  try {
    const result = read(resourceFile(file, "settlement.json"));
    const owner = readBrowserAcquisition(file);
    if (Object.keys(result).sort().join(",") !== "attempt_id,state" || result.attempt_id !== owner.attempt_id || result.state !== "released") throw invalid();
    return result;
  } catch (error) { if (error.code === "ENOENT") return null; throw error; }
}

function writeResource(file, name, resource, { duringStop = false } = {}) {
  const owner = readBrowserAcquisition(file);
  if (receipt(file) || (!duringStop && stopping(file))) throw invalid();
  const value = { attempt_id: owner.attempt_id, ...resource };
  try { atomicLocalFile(resourceFile(file, name), bytes(value)); }
  catch (error) {
    if (error.code !== "EEXIST" || bytes(read(resourceFile(file, name))) !== bytes(value)) throw error;
  }
}

const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u;
const roles = ["producer", "backend", "frontend"];
function stopping(file) {
  try {
    const value = read(resourceFile(file, "stopping.json"));
    if (Object.keys(value).join(",") !== "attempt_id" || value.attempt_id !== readBrowserAcquisition(file).attempt_id) throw invalid();
    return true;
  } catch (error) { if (error.code === "ENOENT") return false; throw error; }
}
function launchRecord(file, launchID) {
  try {
    const owner = readBrowserAcquisition(file);
    if (!uuid.test(launchID)) throw invalid();
    const value = read(resourceFile(file, `launch-${launchID}.json`));
    if (Object.keys(value).sort().join(",") !== "attempt_id,launch_id,role" ||
        value.attempt_id !== owner.attempt_id || value.launch_id !== launchID || !roles.includes(value.role)) throw invalid();
    return value;
  } catch (cause) { throw invalid(cause); }
}
export function createAcquisitionLaunch(file, role) {
  if (!roles.includes(role)) throw invalid();
  const launchID = randomUUID();
  writeResource(file, `launch-${launchID}.json`, { launch_id: launchID, role });
  return launchID;
}
// Only the creator calls this after proving no child was created or reaping
// its exact process group. Recovery never infers it from an absent PID journal.
export function closeAcquisitionLaunch(file, launchID) {
  launchRecord(file, launchID);
  if (receipt(file)) return;
  writeResource(file, `closed-${launchID}.json`, { launch_id: launchID }, { duringStop: true });
}
export function recordAcquisitionProcess(file, launchID, pid = process.pid) {
  const { role } = launchRecord(file, launchID);
  const proof = ownedProcess(pid);
  writeResource(file, `process-${launchID}-${pid}.json`, { launch_id: launchID, role, proof }, { duringStop: true });
  return proof;
}

export function recordAcquisitionPort(file, directory) {
  const owner = readBrowserAcquisition(file);
  // This is called only after exclusive reservation, before the lease is used.
  // The marker binds subsequent transfers to this attempt, not a reused PID.
  const marker = { attempt_id: owner.attempt_id };
  writeResource(file, `port-${digest(directory)}.json`, { directory });
  atomicLocalFile(path.join(directory, "acquisition.json"), bytes(marker));
}

export function reserveAcquisitionPort(file, directory, pid) {
  // Publish intent first. Interruption between mkdir and the marker is an
  // unresolved reservation, never permission to infer absence or delete it.
  const journal = resourceFile(file, `port-${digest(directory)}.json`);
  const prior = existsSync(journal);
  writeResource(file, `port-${digest(directory)}.json`, { directory });
  try { mkdirSync(directory, { mode: 0o700 }); }
  catch (error) {
    if (error.code !== "EEXIST") throw error;
    if (prior) {
      const marker = read(path.join(directory, "acquisition.json"));
      if (Object.keys(marker).join(",") !== "attempt_id" || marker.attempt_id !== readBrowserAcquisition(file).attempt_id) throw invalid();
      return true;
    }
    removePrivateFile(journal);
    return false;
  }
  const owner = readBrowserAcquisition(file);
  atomicLocalFile(path.join(directory, "acquisition.json"), bytes({ attempt_id: owner.attempt_id }));
  atomicLocalFile(path.join(directory, "pid"), `${pid}\n`);
  return true;
}

function resources(file) {
  const owner = readBrowserAcquisition(file);
  const names = readdirSync(path.dirname(file));
  if (names.length > 256) throw invalid();
  const result = { launches: [], closed: new Set(), processes: [], ports: [] };
  for (const name of names) {
    if (name === "owner.json" || name === "settlement.json" || name === "stopping.json" || name.startsWith(".publishing-")) continue;
    const value = read(resourceFile(file, name));
    if (value.attempt_id !== owner.attempt_id) throw invalid();
    if (name.startsWith("launch-")) {
      const launch = launchRecord(file, value.launch_id);
      if (name !== `launch-${launch.launch_id}.json`) throw invalid();
      result.launches.push(launch);
    } else if (name.startsWith("closed-")) {
      launchRecord(file, value.launch_id);
      if (Object.keys(value).sort().join(",") !== "attempt_id,launch_id" || name !== `closed-${value.launch_id}.json`) throw invalid();
      result.closed.add(value.launch_id);
    } else if (name.startsWith("process-")) {
      const p = value.proof;
      if (Object.keys(value).sort().join(",") !== "attempt_id,launch_id,proof,role" || launchRecord(file, value.launch_id).role !== value.role ||
          !p || Object.keys(p).sort().join(",") !== "boot,group,pid,start" || !Number.isSafeInteger(p.pid) || p.pid < 2 ||
          typeof p.boot !== "string" || typeof p.group !== "boolean" || !/^\d+$/u.test(p.start) || name !== `process-${value.launch_id}-${p.pid}.json`) throw invalid();
      result.processes.push(value);
    } else if (name.startsWith("port-")) {
      if (Object.keys(value).sort().join(",") !== "attempt_id,directory" || typeof value.directory !== "string" ||
          !path.isAbsolute(value.directory) || name !== `port-${digest(value.directory)}.json`) throw invalid();
      result.ports.push(value);
    } else throw invalid();
  }
  if (new Set(result.processes.map((entry) => entry.launch_id)).size !== result.processes.length) throw invalid();
  return result;
}

export async function stopAcquisitionProcesses(file, { producerActive = false, stop = stopOwnedProcess } = {}) {
  const stopped = new Set();
  // Stopping fences new launches. First stop the producer, then rescan because
  // a child can publish its exact identity while its creator is being reaped.
  for (let pass = 0; pass < 4; pass++) {
    const records = resources(file);
    const pending = records.processes.filter((entry) => !(producerActive && entry.role === "producer") &&
      !stopped.has(`${entry.launch_id}:${entry.proof.pid}`)).sort((a, b) => Number(b.role === "producer") - Number(a.role === "producer"));
    if (pending.length) {
      const failures = [];
      for (const entry of pending) {
        try { await stop(entry.proof); stopped.add(`${entry.launch_id}:${entry.proof.pid}`); }
        catch (error) { failures.push(error); }
      }
      if (failures.length) throw new AggregateError(failures, "browser process cleanup failed");
      continue;
    }
    if (records.launches.some((entry) => !records.closed.has(entry.launch_id) &&
        !records.processes.some((p) => p.launch_id === entry.launch_id))) {
      throw new CommandFailure("browser process launch has unresolved ownership", { failure_class: "harness", failure_reason: "cleanup_error" });
    }
    return;
  }
  throw new CommandFailure("browser processes did not settle", { failure_class: "harness", failure_reason: "cleanup_error" });
}

export function releaseAcquisitionPorts(file) {
  const owner = readBrowserAcquisition(file);
  const failures = [];
  for (const { directory } of resources(file).ports) {
    try {
      if (!existsSync(directory)) continue;
      const marker = read(path.join(directory, "acquisition.json"));
      if (Object.keys(marker).join(",") !== "attempt_id" || marker.attempt_id !== owner.attempt_id) throw invalid();
      removePrivateTree(directory);
    } catch (error) { failures.push(error); }
  }
  if (failures.length) throw new AggregateError(failures, "browser port cleanup failed");
}

export function acknowledgeBrowserSettlement(file) {
  const owner = readBrowserAcquisition(file);
  if (!receipt(file)) try {
    atomicLocalFile(resourceFile(file, "settlement.json"), bytes({ attempt_id: owner.attempt_id, state: "released" }));
  } catch (error) { if (error.code !== "EEXIST" || !receipt(file)) throw error; }
}

// Service data remains owned by the exact managed suite, including allocations
// interrupted before fixture metadata publication. Releasing this dependent
// permits that suite's existing ledger/container finalizer to run; it never
// deletes or reconstructs database/bucket identities here.
async function settle({ file, runtime, root, environment = {}, producerActive = false, recovery = false,
  stop = stopOwnedProcess, run = spawnSync } = {}) {
  const owner = readBrowserAcquisition(file, runtime);
  if (receipt(file)) {
    await stopAcquisitionProcesses(file, { producerActive, stop });
    return { state: "released" };
  }
  const suiteBytes = readLocalFile(owner.suite_lease, { maximum: 1048576 });
  if (digest(suiteBytes) !== owner.suite_digest) throw invalid();
  const suite = suiteProof(suiteBytes, owner.run_id);
  if (suite.suite_id !== owner.suite_id) throw invalid();
  writeResource(file, "stopping.json", {}, { duringStop: true });
  await stopAcquisitionProcesses(file, { producerActive, stop });
  releaseAcquisitionPorts(file);
  const metadata = path.join(owner.session_root, "runtime-root", "test-services-web-e2e.json");
  if (existsSync(metadata)) {
    readLocalFile(metadata, { maximum: 1048576 });
    // Ordinary retirement belongs to the original run, even when the provider
    // has only tool/service environment. Later recovery must never backfill that
    // run or collide with immutable output from an earlier recovery attempt.
    const resultsRoot = recovery
      ? path.join(owner.runtime_root, "lifecycle", `browser-recovery-${randomUUID()}`)
      : suite.result_root;
    if (recovery) privateDirectory(path.join(resultsRoot, owner.run_id));
    const result = run(path.join(root, `${workspaceLayout.toolbin}/cartulary-test-services`), ["cleanup-web-e2e", "--metadata-file", metadata], {
      cwd: root, env: { ...environment, CARTULARY_TEST_SUITE_ID: owner.suite_id, CARTULARY_TEST_SERVICES_CALL_MODE: "attach",
        CARTULARY_TEST_RESULTS_DIR: resultsRoot, CARTULARY_TEST_RUN_ID: owner.run_id, CARTULARY_TEST_TARGET: suite.target,
        CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: owner.runtime_root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: owner.runtime_lease_id,
        CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: owner.run_id, CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE: "1" },
      timeout: 120000, killSignal: "SIGKILL", stdio: "ignore",
    });
    if (result.error || result.status !== 0) throw new CommandFailure("browser fixture retirement failed", {
      failure_class: "harness", failure_reason: "cleanup_error",
    }, { cause: result.error });
    removePrivateFile(metadata);
  }
  removePrivateTree(path.join(owner.session_root, "runtime-root"));
  // An exit trap can release services but cannot prove its own producer has
  // exited. Only the external owner may publish terminal settlement after reaping.
  if (producerActive) return { state: "recovery_required", acquisition_file: file };
  acknowledgeBrowserSettlement(file);
  return { state: "released" };
}

const settlements = new Map();
export function settleBrowserAcquisition(options) {
  const key = path.resolve(options.file);
  if (!settlements.has(key)) {
    const attempt = settle(options).catch((error) => {
      error.settlement = { state: "recovery_required", acquisition_file: key };
      throw error;
    }).finally(() => settlements.delete(key));
    settlements.set(key, attempt);
  }
  return settlements.get(key);
}

async function launch(file, launchID, command, args) {
  recordAcquisitionProcess(file, launchID);
  if (stopping(file)) return;
  // Install before spawning. A group signal reaches the command as well; keep
  // its supervisor alive to await the command's exit trap and reap it.
  const signals = new Map(["SIGINT", "SIGTERM"].map((name) => [name, () => {}]));
  for (const [name, listener] of signals) process.on(name, listener);
  const child = spawn(command, args, { stdio: "inherit", env: process.env });
  await new Promise((resolve, reject) => { child.once("error", reject); child.once("close", (code, signal) => {
    process.exitCode = code ?? (signal === "SIGINT" ? 130 : signal === "SIGTERM" ? 143 : 1); resolve();
  }); });
  for (const [name, listener] of signals) process.removeListener(name, listener);
}

// The service launcher lives in the producer group until it has published the
// detached child's exact identity. Its ticket survives interruption before that
// publication; the detached wrapper also journals before starting any payload.
async function spawnService(file, role, logFile, command, args) {
  const launchID = createAcquisitionLaunch(file, role);
  let output, child, proof, primary;
  try {
    output = createSecureWriteStream(logFile);
    if (stopping(file)) throw invalid();
    child = spawn(process.execPath, ["--", fileURLToPath(import.meta.url), "launch", file, launchID, command, ...args], {
      detached: true, stdio: ["ignore", output.fd, output.fd], env: process.env,
    });
    // Attach before returning to the event loop, including failed exec/spawn.
    const spawned = new Promise((resolve, reject) => { child.once("spawn", resolve); child.once("error", reject); });
    spawned.catch(() => {});
    if (child.pid) {
      proof = ownedProcess(child.pid);
      recordAcquisitionProcess(file, launchID, child.pid);
    }
    await spawned;
    if (stopping(file)) throw invalid();
    process.stdout.write(`${child.pid}\n`);
    child.unref();
  } catch (error) {
    primary = error;
    try {
      if (proof) await stopOwnedProcess(proof);
      if (!child?.pid || proof) closeAcquisitionLaunch(file, launchID);
    } catch (cleanupError) { (primary.cleanupFailures ??= []).push(cleanupError); }
  } finally {
    if (output) try {
      await new Promise((resolve, reject) => { output.once("error", reject); output.once("close", resolve); output.end(); });
    } catch (error) { if (primary) (primary.cleanupFailures ??= []).push(error); else primary = error; }
  }
  if (primary) throw primary;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const [operation, file, ...args] = process.argv.slice(2);
  const root = path.resolve(import.meta.dirname, "../../..");
  try {
    if (operation === "launch") await launch(file, args[0], args[1], args.slice(2));
    else if (operation === "port") recordAcquisitionPort(file, args[0]);
    else if (operation === "reserve-port") process.exitCode = reserveAcquisitionPort(file, args[0], Number(args[1])) ? 0 : 1;
    else if (operation === "spawn") await spawnService(file, args[0], args[1], args[2], args.slice(3));
    else if (operation === "settle") await settleBrowserAcquisition({ file, root, environment: process.env, producerActive: args[0] === "producer-active" });
    else throw invalid();
  } catch (error) {
    process.stderr.write(`browser acquisition ${operation} failed\n`);
    process.exitCode = error.failure_reason === "artifact_error" ? 11 : 12;
    // Startup owners publish their known cause through the existing private
    // channel. An exit-trap settlement is secondary to the producer's outcome.
    if (operation !== "settle" && error instanceof CommandFailure) {
      try { if (!readCommandFailure(root)) publishCommandFailure(root, error); }
      catch { process.exitCode = 11; }
    }
  }
}
