import { createLaunchContext } from "../../runtime/launch-context.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { spawn, spawnSync } from "node:child_process";
import { createConnection } from "node:net";
import { once } from "node:events";
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { browserAcquisitionLaunchArguments, createAcquisitionLaunch, closeAcquisitionLaunch, createBrowserAcquisition, readBrowserAcquisition, recordAcquisitionProcess, reserveAcquisitionPort, releaseAcquisitionPorts, settleAcquisitionService, settleBrowserAcquisition } from "../browser-acquisition.mjs";
import { createRuntime, openRecoveryAttempt } from "../../runtime/session-client.mjs";
import { processIdentityAlive } from "../../runtime/host-admission.mjs";
import { ownedProcess, stopOwnedProcess } from "../../runtime/owned-process.mjs";
import { CommandFailure, createCommandFailureContext, publishCommandFailure } from "../../runtime/command-failure.mjs";
import { findCleanupDiagnostic } from "../../runtime/cleanup-diagnostic.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
function fixture(t, services = []) {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "browser-acquisition-"));
  const runRoot = path.join(scratch, "results", "acquisition-test"); mkdirSync(runRoot, { recursive:true, mode: 0o700 });
  const privateRoot=mkdtempSync(path.join(os.tmpdir(),"browser-acquisition-private-"));
  const runtime = createRuntime({ repoRoot: root, runRoot, runID: "acquisition-test", scratchRoot: privateRoot });
  const suiteLease = runtime.recoveryPath("test-services", "suite-lease.json");
  mkdirSync(path.dirname(suiteLease), { mode: 0o700 });
  const resultRoot = path.join(scratch, "results");
  writeFileSync(suiteLease, JSON.stringify({ schema_id: "cartulary.test_services.lease.v3", run_id: runtime.runID, suite_id: "test-suite",
    lease_id: runtime.leaseID, result_root: resultRoot, run_root: path.join(resultRoot, runtime.runID), target: "browser-e2e",
    mode: "owned", ownership_mode: "owned", owner_pid: process.pid, created_at: new Date().toISOString(),
    postgres_provisioning: null, resources: services.map(service => ({kind:"container", service, container_id:`${service}-fixture`, image:"fixture",
      labels:{"cartulary.test-services.managed":"true", "cartulary.test-services.run-id":runtime.runID,
        "cartulary.test-services.suite-id":"test-suite", "cartulary.test-services.service":service}})),
    proof_labels: {}, proof_prefixes: {}, cleanup_state: "not_started" }), { mode: 0o600 });
  const sessionRoot = runtime.recoveryPath("browser-stack-leases", "attempt"); mkdirSync(sessionRoot, { recursive: true, mode: 0o700 });
  const file = path.join(sessionRoot, "acquisition", "owner.json");
  const suiteRef = runtime.record({ kind: "managed_suite", target: suiteLease, depends_on: [] });
  runtime.record({ kind: "browser_stack", target: file, state: "pending", depends_on: [suiteRef] });
  assert.equal(createBrowserAcquisition({ runtime, sessionRoot, suiteLease, leaseFile: path.join(sessionRoot, "stack.lease") }), file);
  const cleanup = [];
  t.after(async () => {
    for (const release of cleanup) await release();
    // These are synthetic service/controller records; real test processes are
    // reaped above before their fixture ownership can be retired.
    for (const record of [{ kind: "managed_suite", target: suiteLease }, { kind: "browser_stack", target: file }]) runtime.record({ ...record, state: "released" });
    runtime.close(); rmSync(scratch, { recursive: true, force: true }); rmSync(privateRoot,{recursive:true,force:true});
  });
  return { scratch, runtime, suiteLease, sessionRoot, file, cleanup };
}

function recoveryOutput(f) {
 if (f.attempt) f.attempt.finish();
 const attempt = openRecoveryAttempt({repoRoot:root,runRoot:path.join(f.scratch,"results",f.runtime.runID),root:f.runtime.root,runID:f.runtime.runID,leaseID:f.runtime.leaseID});
 f.attempt=attempt;
 f.cleanup.push(()=>{if(f.attempt===attempt){attempt.finish();f.attempt=null;}});
 return path.dirname(attempt.allocateOutput());
}

test("recovery inspection reports a bounded diagnostic without private proof", (t) => {
  const f = fixture(t), original = readFileSync(f.file);
  const secret = "private-proof-payload-must-not-be-reported";
  writeFileSync(f.file, JSON.stringify({ private_payload: secret }));
  try {
    const environment = { ...process.env, CARTULARY_TEST_RESULTS_DIR: path.join(f.scratch, "results"), CARTULARY_TEST_RUN_ID: f.runtime.runID,
      CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: f.runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: f.runtime.runID,
      CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: f.runtime.leaseID };
    const result = spawnSync(process.execPath, [path.join(root, "tools/harness/browser/fixture-lifecycle.mjs"), "inspect", f.file],
      { cwd: root, env: environment, encoding: "utf8", timeout: 5000, maxBuffer: 4096 });
    assert.equal(result.error, undefined);
    assert.equal(result.status, 11);
    assert.deepEqual(JSON.parse(result.stdout), { failure_class: "artifact", failure_reason: "artifact_error" });
    const diagnostic = JSON.parse(result.stderr);
    assert.equal(diagnostic.schema_id, "cartulary.harness_cleanup_diagnostic.v3");
    assert.equal(diagnostic.stage, "recovery_proof");
    assert.deepEqual(Object.keys(diagnostic).sort(), ["cause", "errno", "schema_id", "stage"]);
    assert.equal((result.stdout + result.stderr).includes(secret), false);
    assert.equal((result.stdout + result.stderr).includes(f.file), false);
  } finally { writeFileSync(f.file, original); }
});

test("service supervisor retains its identity after payload exit until physical settlement", async (t) => {
  const f = fixture(t);
  const launchID = createAcquisitionLaunch(f.file, "frontend");
  const environment = { ...process.env, CARTULARY_TEST_RESULTS_DIR: path.join(f.scratch, "results"), CARTULARY_TEST_RUN_ID: f.runtime.runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: f.runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: f.runtime.runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: f.runtime.leaseID };
  delete environment.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
  const marker = path.join(f.scratch, "payload.json");
  const payload = `const {spawn}=require('node:child_process');const fs=require('node:fs');const child=spawn(process.execPath,['-e','process.stdout.write("ready");setInterval(()=>{},1000)'],{stdio:['ignore','pipe','ignore']});child.stdout.once('data',()=>{fs.writeFileSync(${JSON.stringify(marker)},JSON.stringify({payload:process.pid,descendant:child.pid}));process.exit(0);});`;
  const child = spawn(process.execPath, browserAcquisitionLaunchArguments(f.file, launchID, process.execPath, ["-e", payload]), {
    cwd: root, env: environment, detached: true, stdio: "ignore",
  });
  const closed = once(child, "close");
  const proof = ownedProcess(child.pid);
  f.cleanup.push(async () => { if (child.exitCode === null && child.signalCode === null) child.kill("SIGKILL"); await closed; });
  let processes;
  const deadline = Date.now() + 5000;
  while (Date.now() < deadline) {
    if (existsSync(marker)) processes = JSON.parse(readFileSync(marker, "utf8"));
    if (processes && !existsSync(`/proc/${processes.payload}`)) break;
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  assert.ok(processes && !existsSync(`/proc/${processes.payload}`), "payload must already be reaped");
  const descendant = ownedProcess(processes.descendant);
  // Independent fallback authority keeps a deliberately failing regression
  // from depending on the broken wrapper's group anchor for its own teardown.
  f.runtime.record({ kind: "helper_process", target: descendant, depends_on: [] });
  f.cleanup.unshift(async () => {
    await stopOwnedProcess(descendant, { environment });
    f.runtime.record({ kind: "helper_process", target: descendant, state: "released" });
  });
  // Give a prematurely returning wrapper time to exit. The invariant is that
  // no amount of idle time can retire the anchor before physical settlement.
  await new Promise((resolve) => setTimeout(resolve, 100));
  assert.equal(processIdentityAlive(proof), true, "payload completion must not discard the group anchor");
  await stopOwnedProcess(proof);
  await closed;
  assert.equal(processIdentityAlive(descendant), false, "physical settlement must include surviving descendants");
});

test("browser acquisition before launch settles without a ready lease", async (t) => {
  const f = fixture(t);
  const environment = { ...process.env, CARTULARY_TEST_RESULTS_DIR: path.join(f.scratch, "results"), CARTULARY_TEST_RUN_ID: f.runtime.runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: f.runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: f.runtime.runID, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: f.runtime.leaseID };
  delete environment.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
  const before = readFileSync(f.file);
  for (const [script, operation, args, status] of [
    ["fixture-lifecycle.mjs", "reaped", [f.file, "00000000-0000-4000-8000-000000000000"], 12],
    ["browser-acquisition.mjs", "port", [f.file, path.join(f.scratch, "retired-port")], 11],
  ]) {
    const result = spawnSync(process.execPath, [path.join(root, "tools/harness/browser", script), operation, ...args], { cwd: root, env: environment, timeout: 5000, maxBuffer: 4096 });
    assert.equal(result.error, undefined);
    assert.equal(result.status, status);
    assert.ok(result.stderr.length < 1024);
  }
  assert.deepEqual(readFileSync(f.file), before);
  assert.equal(existsSync(path.join(f.scratch, "retired-port")), false);
  assert.equal(existsSync(path.join(f.sessionRoot, "stack.lease")), false);
  assert.equal((await settleBrowserAcquisition({ ...f, root })).state, "released");
  f.runtime.record({ kind: "browser_stack", target: f.file, state: "released" });
  assert.equal(existsSync(f.suiteLease), true, "browser settlement never closes its borrowed suite");
  assert.equal((await settleBrowserAcquisition({ ...f, root })).state, "released");
});

for (const operation of ["retire", "reset"]) test(`orphaned service ${operation} uses its explicit runtime authority`, async (t) => {
  const f = fixture(t);
  const environment = { ...process.env, CARTULARY_TEST_RESULTS_DIR: path.join(f.scratch, "results"), CARTULARY_TEST_RUN_ID: f.runtime.runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: f.runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: f.runtime.runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: f.runtime.leaseID };
  delete environment.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
  delete environment.CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT;
  const launched = spawnSync(process.execPath, ["--", path.join(root, "tools/harness/browser/browser-acquisition.mjs"),
    "spawn", f.file, "frontend", path.join(f.scratch, "service.log"), process.execPath, "-e", "setInterval(()=>{},1000)"],
  { cwd: root, env: environment, encoding: "utf8", timeout: 10000, maxBuffer: 4096 });
  assert.equal(launched.error, undefined);
  assert.equal(launched.status, 0);
  const pid = Number(launched.stdout.trim());
  const proof = ownedProcess(pid);
  // The launcher has exited, so ancestry cannot authorize cleanup. Independent
  // journal-bound teardown must also work when the assertion intentionally fails.
  f.cleanup.unshift(() => stopOwnedProcess(proof, { environment }));
  assert.equal(processIdentityAlive(proof), true);
  if (operation === "reset") await settleAcquisitionService(f.file, pid, { environment });
  else await settleBrowserAcquisition({ ...f, root, environment });
  assert.equal(processIdentityAlive(proof), false);
  assert.equal(existsSync(f.suiteLease), true, "service settlement preserves the borrowed suite");
});

for (const recovery of [false, true]) test(`browser fixture retirement binds ${recovery ? "recovery" : "ordinary cleanup"} output independently of caller environment`, async (t) => {
  const f = fixture(t);
  const metadata = path.join(f.sessionRoot, "runtime-root", "test-services-web-e2e.json");
  mkdirSync(path.dirname(metadata), { mode: 0o700 });
  writeFileSync(metadata, "{}", { mode: 0o600 });
  const suite = JSON.parse(readFileSync(f.suiteLease, "utf8"));
  let calls = 0;
  await settleBrowserAcquisition({ ...f, root, recovery, recoveryResultsRoot: recovery ? recoveryOutput(f) : undefined,
    environment: { ...(f.attempt?.environment() ?? {}), CARTULARY_TEST_RESULTS_DIR: "/unrelated-results", CARTULARY_TEST_RUN_ID: "unrelated-run",
      CARTULARY_TEST_TARGET: "unrelated-target", CARTULARY_TEST_SUITE_ID: "unrelated-suite" },
    run: (command, args, options) => {
      calls++;
      assert.equal(command, path.join(root, ".cache/cartulary/toolbin/cartulary-test-services"));
      assert.deepEqual(args, ["cleanup-web-e2e", "--metadata-file", metadata]);
      assert.equal(options.env.CARTULARY_TEST_RUN_ID, f.runtime.runID);
      assert.equal(options.env.CARTULARY_TEST_SUITE_ID, suite.suite_id);
      assert.equal(options.env.CARTULARY_TEST_TARGET, suite.target);
      const output = options.env.CARTULARY_TEST_RESULTS_DIR;
      if (recovery) {
        assert.equal(path.dirname(path.dirname(output)), path.join(f.runtime.root,"recovery-detail",f.attempt.id));
        assert.match(path.basename(output), /^[0-9a-f-]{36}$/u);
      } else assert.equal(output, suite.result_root);
      return { status: 0 };
    } });
  assert.equal(calls, 1);
  assert.equal(existsSync(metadata), false);
});

test("explicit browser recovery attempts cannot overwrite earlier immutable retirement output", async (t) => {
  const f = fixture(t);
  const metadata = path.join(f.sessionRoot, "runtime-root", "test-services-web-e2e.json");
  mkdirSync(path.dirname(metadata), { mode: 0o700 });
  writeFileSync(metadata, "{}", { mode: 0o600 });
  const outputs = [];
  const options = { ...f, root, recovery: true, recoveryResultsRoot: recoveryOutput(f), environment: f.attempt.environment(), run: (_command, _args, { env }) => {
    const output = env.CARTULARY_TEST_RESULTS_DIR;
    assert.equal(typeof output, "string");
    assert.equal(path.dirname(path.dirname(output)), path.join(f.runtime.root,"recovery-detail",f.attempt.id));
    assert.ok(!outputs.includes(output), "each explicit attempt has independent immutable output");
    outputs.push(output);
    return { status: outputs.length === 1 ? 1 : 0 };
  } };
  await assert.rejects(settleBrowserAcquisition(options), (error) => error.failure_reason === "cleanup_error");
  assert.equal(existsSync(metadata), true);
  await assert.rejects(settleBrowserAcquisition(options), (error) => error.failure_reason === "artifact_error");
  assert.equal(outputs.length, 1, "a repeated output cannot reach the physical writer");
  options.recoveryResultsRoot = recoveryOutput(f);
  options.environment = f.attempt.environment();
  assert.equal((await settleBrowserAcquisition(options)).state, "released");
  assert.equal(outputs.length, 2);
});

test("browser retirement imports only bounded child diagnostics and preserves failed ownership", async (t) => {
  const valid = { schema_id: "cartulary.harness_cleanup_diagnostic.v3", stage: "browser_namespace", cause: "io", errno: 13 };
  for (const [label, stderr, expected] of [
    ["typed", JSON.stringify(valid), valid],
    ["private", JSON.stringify({ ...valid, credential: "private-sentinel" }), null],
    ["unstructured", "private-sentinel", null],
    ["oversized", "x".repeat(4097), null],
  ]) await t.test(label, async (t) => {
    const f = fixture(t), metadata = path.join(f.sessionRoot, "runtime-root", "test-services-web-e2e.json");
    mkdirSync(path.dirname(metadata), { mode: 0o700 });
    writeFileSync(metadata, "{}", { mode: 0o600 });
    await assert.rejects(settleBrowserAcquisition({ ...f, root, run: (_command, _args, options) => {
      assert.equal(options.maxBuffer, 4096);
      assert.deepEqual(options.stdio, ["ignore", "ignore", "pipe"]);
      return { status: 1, stderr: Buffer.from(stderr) };
    } }), error => {
      assert.equal(error.failure_reason, "cleanup_error");
      const diagnostic = findCleanupDiagnostic(error);
      assert.deepEqual(diagnostic, expected ?? { ...valid, stage: "browser_data_retirement", cause: "internal", errno: 0 });
      assert.equal(JSON.stringify(error).includes("private-sentinel"), false);
      return true;
    });
    assert.equal(existsSync(metadata), true);
    assert.equal(existsSync(path.join(path.dirname(f.file), "settlement.json")), false);
  });
});

test("whole-run recovery delegates ordinary data only to exact owned physical parents", async (t) => {
  for (const variant of ["owned", "borrowed", "partial", "foreign", "measurement", "ordinary", "process_failure"]) await t.test(variant, async t => {
    const f = fixture(t, ["postgres", "object_store"]);
    const parentFile = f.runtime.recoveryPath("test-services", "physical-parent.json");
    const parent = JSON.parse(readFileSync(f.suiteLease));
    if (variant === "borrowed") parent.ownership_mode = "attach";
    if (variant === "partial") parent.resources.pop();
    if (variant === "foreign") parent.resources[0].container_id = "foreign-container";
    writeFileSync(parentFile, JSON.stringify(parent), { mode: 0o600 });
    const metadata = path.join(f.sessionRoot, "runtime-root", "test-services-web-e2e.json");
    mkdirSync(path.dirname(metadata), { mode: 0o700 });
    writeFileSync(metadata, JSON.stringify({database_name:"ct_fixture",bucket:"ct-fixture-web-e2e",
      ...(variant === "measurement" ? {fixture_profile_id:"measurement-fixture"} : {})}), {mode:0o600});
    if (variant === "process_failure") {
      const launch = createAcquisitionLaunch(f.file, "backend");
      recordAcquisitionProcess(f.file, launch);
    }
    const recovery = variant !== "ordinary";
    let calls = 0;
    const options = {...f, root, recovery, ownedServices:recovery ? [parentFile] : [],
      recoveryResultsRoot:recovery ? recoveryOutput(f) : undefined, environment:f.attempt?.environment() ?? {},
      stop:async () => { throw new Error("unsettled process"); }, run:() => { calls++; return {status:0}; }};
    if (["borrowed", "partial", "foreign", "process_failure"].includes(variant)) {
      await assert.rejects(settleBrowserAcquisition(options));
      assert.equal(calls, 0);
      assert.equal(existsSync(metadata), true);
    } else {
      assert.equal((await settleBrowserAcquisition(options)).state, "released");
      assert.equal(calls, variant === "owned" ? 0 : 1);
      assert.equal(existsSync(metadata), false);
    }
    assert.equal(existsSync(parentFile), true, "browser settlement cannot consume physical parent proof");
  });
});


test("producer exit cleanup cannot publish release before the external owner reaps it", async (t) => {
  const f = fixture(t);
  recordAcquisitionProcess(f.file, createAcquisitionLaunch(f.file, "producer"));
  const stopped = [];
  const stop = async (proof) => { stopped.push(proof.pid); };
  const pending = await settleBrowserAcquisition({ ...f, root, producerActive: true, stop });
  assert.equal(pending.state, "recovery_required");
  assert.deepEqual(stopped, []);
  assert.equal(existsSync(path.join(path.dirname(f.file), "settlement.json")), false);
  assert.equal((await settleBrowserAcquisition({ ...f, root, stop })).state, "released");
  assert.deepEqual(stopped, [process.pid]);
});


test("pre-lease partial acquisition reaps real processes and exact ports before suite cleanup", async (t) => {
  const f = fixture(t);
  const child = spawn(process.execPath, ["-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore" });
  const closed = once(child, "close");
  const proof = ownedProcess(child.pid);
  f.cleanup.push(async () => { await stopOwnedProcess(proof); await closed; });
  recordAcquisitionProcess(f.file, createAcquisitionLaunch(f.file, "backend"), child.pid);
  const ports = path.join(f.scratch, "ports"); mkdirSync(ports, { mode: 0o700 });
  const directory = path.join(ports, "port-19001");
  assert.equal(reserveAcquisitionPort(f.file, directory, process.pid), true);
  const sibling = path.join(ports, "port-19002"); mkdirSync(sibling, { mode: 0o700 });
  await settleBrowserAcquisition({ ...f, root });
  await closed;
  assert.equal(existsSync(directory), false);
  assert.equal(existsSync(sibling), true);
  assert.equal(existsSync(f.suiteLease), true);
});

test("failed process cleanup retains acquisition and suite proof for explicit recovery", async (t) => {
  const f = fixture(t);
  recordAcquisitionProcess(f.file, createAcquisitionLaunch(f.file, "backend"));
  await assert.rejects(settleBrowserAcquisition({ ...f, root, stop: async () => { throw new Error("release sentinel"); } }),
    (error) => error.settlement.state === "recovery_required");
  assert.equal(existsSync(path.join(path.dirname(f.file), "settlement.json")), false);
  const calls = [];
  await settleBrowserAcquisition({ ...f, root, stop: async (proof) => { calls.push(proof.pid); } });
  assert.deepEqual(calls, [process.pid]);
});

test("confirmed launch settlement survives later fixture failure and terminal repetition", async (t) => {
  const f = fixture(t);
  const launchID = createAcquisitionLaunch(f.file, "backend");
  recordAcquisitionProcess(f.file, launchID);
  const metadata = path.join(f.sessionRoot, "runtime-root", "test-services-web-e2e.json");
  mkdirSync(path.dirname(metadata), { mode: 0o700 });
  writeFileSync(metadata, "{}", { mode: 0o600 });
  let stops = 0;
  const stop = async () => {
    assert.equal(++stops, 1, "settled process identity must never be signaled again, even after PID reuse");
  };
  await assert.rejects(settleBrowserAcquisition({ ...f, root, stop, run: () => ({ status: 1 }) }));
  assert.equal((await settleBrowserAcquisition({ ...f, root, stop, run: () => ({ status: 0 }) })).state, "released");
  const receipt = readFileSync(path.join(path.dirname(f.file), "settlement.json"));
  assert.equal((await settleBrowserAcquisition({ ...f, root, stop })).state, "released");
  assert.equal(stops, 1);
  assert.deepEqual(readFileSync(path.join(path.dirname(f.file), "settlement.json")), receipt);
});

test("creator's confirmed process settlement retires its published launch identity", async (t) => {
  const f = fixture(t);
  const launchID = createAcquisitionLaunch(f.file, "backend");
  recordAcquisitionProcess(f.file, launchID);
  closeAcquisitionLaunch(f.file, launchID);
  assert.throws(() => recordAcquisitionProcess(f.file, launchID));
  assert.equal((await settleBrowserAcquisition({ ...f, root,
    stop: async () => assert.fail("closed launch must not reauthorize process control") })).state, "released");
});

test("reset settlement resolves only an open service launch and preserves failed authority", async (t) => {
  const f = fixture(t);
  const launchID = createAcquisitionLaunch(f.file, "backend");
  recordAcquisitionProcess(f.file, launchID);
  let calls = 0;
  await assert.rejects(settleAcquisitionService(f.file, 1, { stop: async () => { calls++; } }));
  await assert.rejects(settleAcquisitionService(f.file, process.pid, { stop: async () => { calls++; throw new Error("unresolved"); } }));
  assert.equal(existsSync(path.join(path.dirname(f.file), `closed-${launchID}.json`)), false);
  await settleAcquisitionService(f.file, process.pid, { stop: async (proof) => { calls++; assert.equal(proof.pid, process.pid); } });
  assert.equal(calls, 2);
  await assert.rejects(settleAcquisitionService(f.file, process.pid, { stop: async () => assert.fail("closed authority") }));
  assert.equal((await settleBrowserAcquisition({ ...f, root, stop: async () => assert.fail("retired reset authority") })).state, "released");
});

test("wrong runtime or changed suite proof is rejected before destructive recovery", async (t) => {
  const f = fixture(t);
  let stopped = false;
  const stop = async () => { stopped = true; };
  recordAcquisitionProcess(f.file, createAcquisitionLaunch(f.file, "backend"));
  await assert.rejects(settleBrowserAcquisition({ ...f, root, runtime: { ...f.runtime, runID: "wrong" }, stop }), (e) => e.failure_reason === "artifact_error");
  writeFileSync(f.suiteLease, "{}", { mode: 0o600 });
  await assert.rejects(settleBrowserAcquisition({ ...f, root, stop }), (e) => e.failure_reason === "artifact_error");
  assert.equal(stopped, false);
});


test("acquisition CLI preserves classified startup proof failures and prior diagnostics", (t) => {
  const f = fixture(t);
  mkdirSync(path.join(f.scratch, f.runtime.runID), { mode: 0o700 });
  const environment = { ...process.env, CARTULARY_TEST_RESULTS_DIR: path.join(f.scratch, "results"), CARTULARY_TEST_RUN_ID: f.runtime.runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: f.runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: f.runtime.runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: f.runtime.leaseID };
  writeFileSync(f.file, "{}", { mode: 0o600 });
  for (const prior of [null, { failure_class: "infra", failure_reason: "service_start_error" }]) {
    const context = createCommandFailureContext({ ...{ repoRoot: root, environment,
      unitID: "review:browser_stack", commandID: "cartulary.harness.command.ui_review.v1" }, launch: createLaunchContext({ repoRoot: root, environment,
      unitID: "review:browser_stack", commandID: "cartulary.harness.command.ui_review.v1" }) });
    try {
      const childEnvironment = { ...environment, ...context.environment };
      if (prior) publishCommandFailure(root, prior, childEnvironment);
      const child = spawnSync(process.execPath, ["--", path.join(root, "tools/harness/browser/browser-acquisition.mjs"),
        "launch", f.file, "00000000-0000-4000-8000-000000000000", process.execPath, "-e", "process.exit(77)"],
      { env: childEnvironment, timeout: 10000, killSignal: "SIGKILL", stdio: "ignore" });
      assert.equal(child.error, undefined);
      assert.equal(child.status, 11);
      assert.deepEqual(context.read(), prior ?? { failure_class: "artifact", failure_reason: "artifact_error" });
    } finally { context.close(); }
  }
});

test("changed port proof preserves unrelated resources and remains recoverable", async (t) => {
  const f = fixture(t);
  const directory = path.join(f.scratch, "port-19001");
  assert.equal(reserveAcquisitionPort(f.file, directory, process.pid), true);
  const marker = path.join(directory, "acquisition.json");
  const prior = readFileSync(marker);
  writeFileSync(marker, JSON.stringify({ attempt_id: "other" }));
  await assert.rejects(settleBrowserAcquisition({ ...f, root }));
  assert.equal(existsSync(directory), true);
  writeFileSync(marker, prior);
  await settleBrowserAcquisition({ ...f, root });
  assert.equal(existsSync(directory), false);
});

test("port reservation is idempotent and collision cannot retire an earlier ownership record", async (t) => {
  const f = fixture(t), directory = path.join(f.scratch, "port-19003");
  assert.equal(reserveAcquisitionPort(f.file, directory, process.pid), true);
  assert.equal(reserveAcquisitionPort(f.file, directory, process.pid), true);
  const marker = path.join(directory, "acquisition.json"), original = readFileSync(marker);
  writeFileSync(marker, JSON.stringify({ attempt_id: "other" }));
  assert.throws(() => reserveAcquisitionPort(f.file, directory, process.pid));
  await assert.rejects(settleBrowserAcquisition({ ...f, root }));
  writeFileSync(marker, original);
  await settleBrowserAcquisition({ ...f, root });
  assert.equal(existsSync(directory), false);
});

test("initial foreign port collision retains intent through failed settlement and recovery", async (t) => {
  const f = fixture(t), directory = path.join(f.scratch, "occupied-port");
  mkdirSync(directory, { mode: 0o700 });
  const marker = path.join(directory, "acquisition.json");
  writeFileSync(marker, JSON.stringify({ attempt_id: "foreign" }), { mode: 0o600 });
  const original = readFileSync(marker);
  assert.equal(reserveAcquisitionPort(f.file, directory, process.pid), false);
  const intent = readdirSync(path.dirname(f.file)).find(name => name.startsWith("port-") && !name.startsWith("port-outcome-"));
  assert.ok(intent, "declared recovery intent must remain present after a collision");
  const intentBytes = readFileSync(path.join(path.dirname(f.file), intent));
  // Even if the first occupant leaves, a declined intent cannot silently turn
  // into a new reservation on a later call by this acquisition.
  rmSync(directory, { recursive: true });
  assert.equal(reserveAcquisitionPort(f.file, directory, process.pid), false);
  assert.equal(existsSync(directory), false);
  mkdirSync(directory, { mode: 0o700 }); writeFileSync(marker, original, { mode: 0o600 });
  recordAcquisitionProcess(f.file, createAcquisitionLaunch(f.file, "backend"));
  await assert.rejects(settleBrowserAcquisition({ ...f, root, stop: async () => { throw new Error("failed stop"); } }));
  assert.deepEqual(readFileSync(path.join(path.dirname(f.file), intent)), intentBytes);
  assert.equal((await settleBrowserAcquisition({ ...f, root, recovery: true, recoveryResultsRoot: recoveryOutput(f), environment: f.attempt.environment(), stop: async () => {} })).state, "released");
  assert.deepEqual(readFileSync(marker), original);
});

test("released port authority stays retired through later failure and foreign reuse", async (t) => {
  const f = fixture(t), directory = path.join(f.scratch, "reused-port");
  assert.equal(reserveAcquisitionPort(f.file, directory, process.pid), true);
  releaseAcquisitionPorts(f.file);
  assert.equal(existsSync(directory), false);
  assert.throws(() => reserveAcquisitionPort(f.file, directory, process.pid));
  mkdirSync(directory, { mode: 0o700 });
  const marker = path.join(directory, "acquisition.json");
  writeFileSync(marker, JSON.stringify({ attempt_id: "new-owner" }), { mode: 0o600 });
  const original = readFileSync(marker);
  const metadata = path.join(f.sessionRoot, "runtime-root", "test-services-web-e2e.json");
  mkdirSync(path.dirname(metadata), { mode: 0o700 }); writeFileSync(metadata, "{}", { mode: 0o600 });
  await assert.rejects(settleBrowserAcquisition({ ...f, root, run: () => ({ status: 1 }) }));
  assert.equal((await settleBrowserAcquisition({ ...f, root, recovery: true, recoveryResultsRoot: recoveryOutput(f), environment: f.attempt.environment(), run: () => ({ status: 0 }) })).state, "released");
  assert.deepEqual(readFileSync(marker), original);
});

test("invalid terminal port outcomes cannot excuse missing or foreign authority", async (t) => {
  for (const variant of ["foreign", "malformed", "orphan", "unsettled_absence"]) {
    await t.test(variant, async (t) => {
      const f = fixture(t), directory = path.join(f.scratch, "port");
      assert.equal(reserveAcquisitionPort(f.file, directory, process.pid), true);
      const intent = readdirSync(path.dirname(f.file)).find(name => name.startsWith("port-"));
      if (variant === "unsettled_absence") rmSync(directory, { recursive: true });
      else {
        const value = JSON.parse(readFileSync(path.join(path.dirname(f.file), intent), "utf8"));
        value.state = "released";
        if (variant === "foreign") value.attempt_id = "foreign";
        if (variant === "malformed") value.state = "unknown";
        writeFileSync(path.join(path.dirname(f.file), intent.replace("port-", "port-outcome-")), JSON.stringify(value), { mode: 0o600 });
        if (variant === "orphan") rmSync(path.join(path.dirname(f.file), intent));
      }
      await assert.rejects(settleBrowserAcquisition({ ...f, root }));
      assert.equal(existsSync(path.join(path.dirname(f.file), "settlement.json")), false);
      if (variant !== "unsettled_absence") assert.equal(existsSync(directory), true);
    });
  }
});

test("producer death before ready publication leaves exact service process identity", async (t) => {
  const f = fixture(t);
  const child = spawn(process.execPath, ["--", path.join(root, "tools/harness/browser/browser-acquisition.mjs"), "launch", f.file, createAcquisitionLaunch(f.file, "producer"),
    process.execPath, "-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore" });
  const closed = once(child, "close");
  const proof = ownedProcess(child.pid); f.cleanup.push(async () => { await stopOwnedProcess(proof); await closed; });
  const deadline = Date.now() + 10000;
  while (!readdirSync(path.dirname(f.file)).some((name) => name.startsWith("process-"))) {
    assert.ok(Date.now() < deadline, "producer publishes ownership before work");
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  await settleBrowserAcquisition({ ...f, root });
  await closed;
  assert.equal(existsSync(path.join(f.sessionRoot, "stack.lease")), false);
});

test("concurrent settlement is shared and late acquisition cannot publish after cancellation", async (t) => {
  const f = fixture(t);
  recordAcquisitionProcess(f.file, createAcquisitionLaunch(f.file, "backend"));
  const stopped = Promise.withResolvers();
  let count = 0;
  const options = { ...f, root, stop: async () => { count++; await stopped.promise; } };
  const first = settleBrowserAcquisition(options);
  const second = settleBrowserAcquisition(options);
  assert.equal(first, second);
  stopped.resolve();
  await Promise.all([first, second]);
  assert.equal(count, 1);
  assert.throws(() => recordAcquisitionProcess(f.file, createAcquisitionLaunch(f.file, "frontend")), (error) => error.failure_reason === "artifact_error");
  assert.equal(existsSync(path.join(f.sessionRoot, "stack.lease")), false);
});

test("unbound launch intent cannot settle until its creator proves no child remains", async (t) => {
  const f = fixture(t);
  const launchID = createAcquisitionLaunch(f.file, "backend");
  await assert.rejects(settleBrowserAcquisition({ ...f, root }), (error) => error.settlement?.state === "recovery_required");
  assert.equal(existsSync(path.join(path.dirname(f.file), "settlement.json")), false);
  assert.throws(() => createAcquisitionLaunch(f.file, "frontend"));
  closeAcquisitionLaunch(f.file, launchID);
  assert.equal((await settleBrowserAcquisition({ ...f, root })).state, "released");
});

test("a late launcher journals after stopping without starting its payload", async (t) => {
  const f = fixture(t);
  const launchID = createAcquisitionLaunch(f.file, "backend");
  await assert.rejects(settleBrowserAcquisition({ ...f, root }));
  const payload = path.join(f.scratch, "payload-started");
  const result = spawnSync(process.execPath, ["--", path.join(root, "tools/harness/browser/browser-acquisition.mjs"),
    "launch", f.file, launchID, process.execPath, "-e", `require("node:fs").writeFileSync(${JSON.stringify(payload)}, "started")`],
  { timeout: 10000, killSignal: "SIGKILL", stdio: "ignore" });
  assert.equal(result.error, undefined);
  assert.equal(result.status, 0);
  assert.equal(existsSync(payload), false);
  assert.equal(readdirSync(path.dirname(f.file)).filter((name) => name.startsWith("process-")).length, 1);
  assert.equal((await settleBrowserAcquisition({ ...f, root })).state, "released");
});
