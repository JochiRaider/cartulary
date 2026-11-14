import assert from "node:assert/strict";
import { closeSync, fstatSync, openSync, readFileSync, writeSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { validateSchemaSync } from "../../contract/index.mjs";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import { spawn, spawnSync } from "node:child_process";
import { runOwnedCommand } from "../owned-command.mjs";

test("native command descriptor output retains caller ownership and exact exit versus signal facts", async t => {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "owned-command-output-"));
  const descriptors = [];
  t.after(() => { for (const fd of descriptors) closeSync(fd); rmSync(scratch, { recursive: true, force: true }); });
  const stdout = openSync(path.join(scratch, "stdout"), "wx", 0o600); descriptors.push(stdout);
  const stderr = openSync(path.join(scratch, "stderr"), "wx", 0o600); descriptors.push(stderr);
  const identities = descriptors.map(fd => ({ dev: fstatSync(fd).dev, ino: fstatSync(fd).ino }));
  for (const [script, code, signal] of [["exit 143", 143, null], ["kill -TERM $$", 143, "SIGTERM"], ["kill -34 $$", 162, "SIG34"]]) {
    const result = await runOwnedCommand("sh", ["-c", `printf output; printf diagnostic >&2; ${script}`], {
      cwd: scratch, environment: process.env, output: { stdout, stderr },
    });
    assert.deepEqual(result, { code, signal, started: true, settled: true, error: undefined });
  }
  assert.deepEqual(descriptors.map(fd => ({ dev: fstatSync(fd).dev, ino: fstatSync(fd).ino })), identities);
  writeSync(stdout, "caller"); writeSync(stderr, "owner");
  assert.equal(readFileSync(path.join(scratch, "stdout"), "utf8"), "outputoutputoutputcaller");
  assert.equal(readFileSync(path.join(scratch, "stderr"), "utf8"), "diagnosticdiagnosticdiagnosticowner");
  for (const output of ["pipe", {}, { stdout }, { stdout, stderr, extra: true }, { stdout: -1, stderr }, { stdout: "private-file", stderr }, [stdout, stderr]]) {
    const result = await runOwnedCommand("sleep", ["60"], { cwd: scratch, environment: process.env, output,
      onReady: () => assert.fail("invalid output policy launched child") });
    assert.match(result.error.message, /invalid owned command output policy/u);
    assert.equal(result.started, false); assert.equal(result.settled, true);
  }
  for (const [command, args] of [[null, []], ["", []], ["sleep", null], ["sleep", [60]], ["sleep\0", []], ["sleep", ["60\0"]]]) {
    const result = await runOwnedCommand(command, args, { cwd: scratch, environment: process.env,
      onReady: () => assert.fail("invalid invocation launched child") });
    assert.match(result.error.message, /invalid owned command invocation/u);
    assert.equal(result.started, false); assert.equal(result.settled, true);
  }
});

test("native-owned Node commands retain fast registration and settle descendants before returning", async t => {
  const { ownedProcess } = await import("../owned-process.mjs");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "owned-command-node-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const descendant = path.join(scratch, "descendant");
  let proof;
  const result = await runOwnedCommand("sh", ["-c", 'sleep 60 </dev/null >/dev/null 2>&1 & echo $! > "$1"; exit 17', "fixture", descendant], {
    cwd: scratch, environment: process.env, onReady: ready => { proof = ready; assert.deepEqual(ownedProcess(ready.pid), ready); },
  });
  assert.deepEqual(result, { code: 17, signal: null, started: true, settled: true, error: undefined });
  assert.throws(() => readFileSync(`/proc/${proof.pid}/stat`), { code: "ENOENT" }, "native workload was not reaped");
  const pid = Number(readFileSync(descendant, "utf8").trim());
  try { assert.match(readFileSync(`/proc/${pid}/stat`, "utf8").split(") ").at(-1), /^[ZX] /u); }
  catch (error) { if (error.code !== "ENOENT") throw error; }
});

test("native-owned Node cancellation and failed registration preserve settlement and unrelated work", async t => {
  const { ownedProcess } = await import("../owned-process.mjs");
  const peerAbort = new AbortController(), abort = new AbortController();
  let peerReady, workReady, peerProof, workProof;
  const peerStarted = new Promise(resolve => { peerReady = resolve; });
  const workStarted = new Promise(resolve => { workReady = resolve; });
  const peer = runOwnedCommand("sleep", ["60"], { cwd: process.cwd(), environment: process.env, signal: peerAbort.signal,
    onReady: proof => { peerProof = proof; peerReady(); } });
  const work = runOwnedCommand("sleep", ["60"], { cwd: process.cwd(), environment: process.env, signal: abort.signal,
    onReady: proof => { workProof = proof; workReady(); } });
  t.after(async () => { abort.abort(); peerAbort.abort(); await Promise.all([peer, work]); });
  await Promise.race([Promise.all([peerStarted, workStarted]), Promise.race([peer, work]).then(result => { throw result.error ?? new Error("command exited before registration"); })]);
  abort.abort(new Error("cancelled fixture"));
  assert.deepEqual(await work, { code: 143, signal: "SIGTERM", started: true, settled: true, error: undefined });
  assert.throws(() => readFileSync(`/proc/${workProof.pid}/stat`), { code: "ENOENT" });
  assert.deepEqual(ownedProcess(peerProof.pid), peerProof);
  const registrationFailure = new Error("registration failed");
  const rejected = await runOwnedCommand("sleep", ["60"], { cwd: process.cwd(), environment: process.env,
    onReady: () => { throw registrationFailure; } });
  assert.equal(rejected.error, registrationFailure);
  assert.equal(rejected.code, 143);
  assert.equal(rejected.settled, true);
  assert.equal(rejected.started, true);
  const earlyAbort = new AbortController();
  let earlyProof;
  const early = runOwnedCommand("sleep", ["60"], { cwd: process.cwd(), environment: process.env, signal: earlyAbort.signal,
    onReady: proof => {
      // Deliberately slow synchronous registration. An already queued abort
      // must not let the native owner reap this proof during the callback.
      Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 100);
      assert.deepEqual(ownedProcess(proof.pid), proof);
      earlyProof = proof;
    } });
  earlyAbort.abort();
  assert.deepEqual(await early, { code: 143, signal: "SIGTERM", started: true, settled: true, error: undefined });
  assert.throws(() => readFileSync(`/proc/${earlyProof.pid}/stat`), { code: "ENOENT" });
  const missing = await runOwnedCommand("/cartulary-absent-executable", [], { cwd: process.cwd(), environment: process.env,
    onReady: () => assert.fail("absent executable registered") });
  assert.equal(missing.error.code, "ENOENT");
  assert.equal(missing.started, false);
  assert.equal(missing.settled, true, "explicit failed acquisition must release its pending ticket");
});

test("Node command adapter rejects corrupt, contradictory or incomplete native receipts", async t => {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "owned-command-protocol-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const proof = { boot: "11111111-2222-3333-4444-555555555555", pid: 100, start: "1", group: true, pgid: 100, sid: 2 };
  const ready = JSON.stringify({ event: "ready", proof }) + "\n";
  const settled = JSON.stringify({ event: "settled", exit: 0, cause: "exited" }) + "\n";
  const cases = [
    [JSON.stringify({ event: "failed", cause: "cleanup" }) + "\n", 12],
    ["", 0], ["{", 0], [ready, 0], [settled, 0], [ready + ready + settled, 0],
    [ready + settled + "{}\n", 0], [ready + settled.trimEnd(), 0], [ready + settled, 1],
    [ready + '{"event":"settled","event":"settled","exit":0,"cause":"exited"}\n', 0],
    [ready + JSON.stringify({ event: "settled", exit: 0, cause: "exited", extra: true }) + "\n", 0],
    ...[null, 0, 1, 15, 65, "SIGTERM", "secret-sentinel"].map(signal => [ready + JSON.stringify({ event: "settled", exit: 0, cause: "exited", signal }) + "\n", 0]),
    [" ".repeat(8193), 12], [JSON.stringify({ event: "ready", proof: { ...proof, group: false } }) + "\n", 12],
    [ready + JSON.stringify({ event: "failed", cause: "cleanup", diagnostic: { secret: "sentinel" } }) + "\n", 12],
  ];
  for (const [index, [events, code]] of cases.entries()) {
    const binary = path.join(scratch, `controller-${index}`);
    writeFileSync(binary, `#!${process.execPath}\nimport {writeSync} from 'node:fs';process.stdin.on('error',()=>{});process.stdin.resume();writeSync(3,${JSON.stringify(events)});process.exit(${code});\n`, { mode: 0o700 });
    const result = await runOwnedCommand("unused", [], { cwd: scratch, environment: { ...process.env, CARTULARY_HARNESS_RUNTIME_BIN: binary } });
    assert.equal(result.settled, false, `fixture ${index} acknowledged settlement`);
    assert.equal(result.code, null);
    assert.equal(result.error.failure_reason, "cleanup_error");
    assert.equal(JSON.stringify(result.error.cleanupDiagnostic).includes("sentinel"), false);
  }
});
test("retired private process dispatch rejects without signaling a live owned process", async (t) => {
  const { once } = await import("node:events");
  const { createRuntime } = await import("../session-client.mjs");
  const { drainOwnedProcess, inspectOwnedProcess, ownedProcess, stopOwnedProcess } = await import("../owned-process.mjs");
  const repository = path.resolve(import.meta.dirname, "../../../..");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "retired-process-dispatch-"));
  const runRoot = path.join(scratch, "results", "retired-dispatch");
  let runtime, child, proof, resource, registered = false;
  let closed = Promise.resolve();
  t.after(async () => {
    if (proof) await stopOwnedProcess(proof, { environment: {
      ...process.env, ...runtime.producerEnvironment(), CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT: runRoot,
    } });
    else if (child && child.exitCode === null && child.signalCode === null) child.kill("SIGKILL");
    await closed;
    if (registered) runtime.record({ ...resource, state: "released" });
    runtime?.close();
    rmSync(scratch, { recursive: true, force: true });
  });
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  runtime = createRuntime({ repoRoot: repository, runRoot, runID: "retired-dispatch", scratchRoot: path.join(scratch, "private") });
  child = spawn(process.execPath, ["-e", "process.on('SIGTERM',()=>{});process.stdout.write('ready');setInterval(()=>{},1000)"], { detached: true, stdio: ["ignore", "pipe", "ignore"] });
  let spawnError;
  child.once("error", (error) => { spawnError = error; });
  closed = new Promise((resolve) => child.once("close", resolve));
  if (child.pid) proof = ownedProcess(child.pid);
  await Promise.race([once(child.stdout, "data"), closed.then(() => { throw spawnError ?? new Error("process fixture exited before readiness"); })]);
  resource = { kind: "helper_process", target: proof, state: "acquired", run_id: runtime.runID, lease_id: runtime.leaseID, depends_on: [], recovery_files: [] };
  validateSchemaSync("cartulary.harness_runtime_resource.v2", resource);
  runtime.record(resource);
  registered = true;
  assert.throws(() => validateSchemaSync("cartulary.harness_runtime_resource.v2", { ...resource, kind: "preparation_process" }));
  const result = spawnSync(process.execPath, [path.join(repository, "tools/harness/runtime/resource-stop-private.mjs")], {
    cwd: repository, env: { ...process.env, ...runtime.producerEnvironment(), CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT: runRoot }, input: JSON.stringify(resource), timeout: 10000, encoding: "utf8"
  });
  assert.equal(result.error, undefined);
  assert.equal(result.status, 12);
  assert.equal(child.exitCode, null);
  assert.doesNotThrow(() => process.kill(child.pid, 0), "retired dispatch signaled a process");
  const environment = { ...process.env, ...runtime.producerEnvironment(), CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT: runRoot };
  assert.equal(await inspectOwnedProcess(proof, { environment }), true);
  await assert.rejects(inspectOwnedProcess({ ...proof, start: "0" }, { environment }));
  await assert.rejects(inspectOwnedProcess({ ...proof, group: !proof.group }, { environment }));
  assert.equal(await inspectOwnedProcess(proof, { environment }), true, "rejected inspection changed the live process");
  await assert.rejects(drainOwnedProcess({ ...proof, start: "0" }, { environment, timeoutMS: 150 }));
  await assert.rejects(drainOwnedProcess(proof, { environment, timeoutMS: 150 }), error =>
    error.failure_class === "timing" && error.failure_reason === "timeout_failure" && error.cleanupDiagnostic.cause === "deadline");
  assert.equal(await inspectOwnedProcess(proof, { environment }), true, "graceful drain forced termination or claimed settlement");
  await stopOwnedProcess(proof, { environment });
  await closed;
  assert.equal(await inspectOwnedProcess(proof, { environment }), false, "settled inspection acquired signal authority");
  await drainOwnedProcess(proof, { environment, timeoutMS: 150 });
});
test("physical cleanup diagnostics remain bounded and cannot authorize success", async (t) => {
  const { stopOwnedProcess } = await import("../owned-process.mjs");
  const { findCleanupDiagnostic, withCleanupDiagnostic } = await import("../cleanup-diagnostic.mjs");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "cleanup-diagnostic-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const valid = { schema_id: "cartulary.harness_cleanup_diagnostic.v4", stage: "process_signal", cause: "io", errno: 13 };
  const cases = [
    [JSON.stringify(valid), "process_signal"],
    [JSON.stringify({ ...valid, process_id: 123 }), "process_protocol"],
    [JSON.stringify({ ...valid, cause: "credential-sentinel" }), "process_protocol"],
    [JSON.stringify(valid) + JSON.stringify(valid), "process_protocol"],
    [" ".repeat(4097), "process_protocol"],
  ];
  for (const [index, [stderr, stage]] of cases.entries()) {
    const binary = path.join(scratch, `physical-${index}`);
    writeFileSync(binary, `#!${process.execPath}\nprocess.stdin.resume();process.stdin.on("end",()=>{process.stderr.write(${JSON.stringify(stderr)});process.exitCode=12;});\n`, { mode: 0o700 });
    await assert.rejects(stopOwnedProcess({}, { environment: { ...process.env, CARTULARY_HARNESS_RUNTIME_BIN: binary } }), error => {
      const diagnostic = findCleanupDiagnostic(new AggregateError([error], "private-path-sentinel"));
      assert.equal(diagnostic.stage, stage);
      assert.equal(JSON.stringify(diagnostic).includes("sentinel"), false);
      return true;
    });
  }
  const privateFailure = new Error("private-path-sentinel");
  Object.defineProperty(privateFailure, "privateDiagnostic", { value: { reason: "unsafe", errno: null, process_id: 123, proof_file: "environ" } });
  const projected = findCleanupDiagnostic(withCleanupDiagnostic(privateFailure, "browser_acknowledgement"));
  assert.deepEqual(projected, { schema_id: valid.schema_id, stage: "browser_acknowledgement", cause: "unsafe", errno: 0 });
  const wrapped = new AggregateError([new Error("private wrapper", { cause: privateFailure })], "private outer");
  assert.deepEqual(findCleanupDiagnostic(withCleanupDiagnostic(wrapped, "recovery_proof")), projected);
  const unprojected = new Error("private-path-sentinel");
  Object.defineProperty(unprojected, "privateDiagnostic", { value: privateFailure.privateDiagnostic });
  const nested = new AggregateError([new Error("private wrapper", { cause: unprojected })], "private outer");
  assert.deepEqual(findCleanupDiagnostic(withCleanupDiagnostic(nested, "recovery_proof")), { schema_id: valid.schema_id, stage: "recovery_proof", cause: "unsafe", errno: 0 });
});
test("cleanup evidence retains each failure without leaking private errors", async (t) => {
  const { CleanupResults } = await import("../cleanup-results.mjs");
  const { withCleanupDiagnostic } = await import("../cleanup-diagnostic.mjs");
  const root = mkdtempSync(path.join(os.tmpdir(), "cleanup-results-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const results = new CleanupResults();
  const cause = new Error("credential-sentinel");
  Object.defineProperty(cause, "privateDiagnostic", { value: { reason: "io", errno: 13, proof_file: "credential-sentinel" } });
  results.record("fixture_release", { error: withCleanupDiagnostic(cause, "browser_namespace") });
  results.record("host_release", { error: new Error("another-private-path") });
  results.record("services_close", { blocked: true });
  results.record("runtime_close");
  results.publish(root, "cleanup-evidence");
  const bytes = readFileSync(path.join(root, "cleanup-results.json"), "utf8");
  assert.equal(bytes.includes("sentinel"), false);
  assert.equal(bytes.includes("private-path"), false);
  const record = JSON.parse(bytes);
  validateSchemaSync("cartulary.harness_cleanup_results.v2", record);
  assert.deepEqual(record.results.map(x => x.diagnostic?.stage ?? null), ["browser_namespace", "host_release", null, null]);
  assert.equal(record.results[0].diagnostic.errno, 13);
  for (const mutate of [
    r => { r.schema_id = "cartulary.harness_cleanup_results.v1"; },
    r => { r.results[0].diagnostic = null; },
    r => { r.results[0].diagnostic.schema_id = "cartulary.harness_cleanup_diagnostic.v2"; },
    r => { r.results[0].diagnostic.secret = "credential-sentinel"; },
    r => { r.results[0].diagnostic.stage = "arbitrary-private-path"; },
    r => { r.results[3].diagnostic = r.results[0].diagnostic; },
  ]) {
    const candidate = structuredClone(record);
    mutate(candidate);
    assert.throws(() => validateSchemaSync("cartulary.harness_cleanup_results.v2", candidate));
  }
});
