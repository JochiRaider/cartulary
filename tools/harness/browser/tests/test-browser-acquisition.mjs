import assert from "node:assert/strict";
import test from "node:test";
import { spawn } from "node:child_process";
import { createConnection } from "node:net";
import { once } from "node:events";
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createBrowserAcquisition, readBrowserAcquisition, recordAcquisitionPort, recordAcquisitionProcess, reserveAcquisitionPort, settleBrowserAcquisition } from "../browser-acquisition.mjs";
import { preparationOwnership, withReviewResources } from "../review-preparation.mjs";
import { createSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { recordRuntimeResource, runtimeRecoveryResources } from "../../runtime/resource-recovery.mjs";
import { ownedProcess, stopOwnedProcess } from "../../runtime/owned-process.mjs";
import { CommandFailure } from "../../runtime/command-failure.mjs";
import { productionFixtureProviders } from "../../scheduler/fixture-broker/providers.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
function fixture(t) {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "browser-acquisition-"));
  const runRoot = path.join(scratch, "run"); mkdirSync(runRoot, { mode: 0o700 });
  const runtime = createSuiteRuntime({ repoRoot: root, runRoot, runID: "acquisition-test", scratchRoot: path.join(scratch, "private") });
  const suiteLease = runtime.privatePath("test-services", "suite-lease.json");
  mkdirSync(path.dirname(suiteLease), { mode: 0o700 });
  writeFileSync(suiteLease, JSON.stringify({ schema_id: "cartulary.test_services.lease.v1", run_id: runtime.runID, suite_id: "test-suite" }), { mode: 0o600 });
  const sessionRoot = runtime.privatePath("browser-stack-leases", "attempt"); mkdirSync(sessionRoot, { recursive: true, mode: 0o700 });
  const file = createBrowserAcquisition({ runtime, sessionRoot, suiteLease, leaseFile: path.join(sessionRoot, "stack.lease") });
  const ownership = preparationOwnership(runtime);
  ownership.record({ kind: "managed_suite", target: suiteLease });
  ownership.record({ kind: "browser_stack", target: file, state: "pending" });
  const cleanup = [];
  t.after(async () => {
    for (const release of cleanup) await release();
    // These are synthetic service/controller records; real test processes are
    // reaped above before their fixture ownership can be retired.
    for (const record of runtimeRecoveryResources(runtime)) recordRuntimeResource(runtime, { ...record, state: "released" });
    runtime.close(); rmSync(scratch, { recursive: true, force: true });
  });
  return { scratch, runtime, suiteLease, sessionRoot, file, ownership, cleanup };
}

test("browser acquisition before launch settles without a ready lease", async (t) => {
  const f = fixture(t);
  assert.equal(existsSync(path.join(f.sessionRoot, "stack.lease")), false);
  assert.equal((await settleBrowserAcquisition({ ...f, root })).state, "released");
  f.ownership.record({ kind: "browser_stack", target: f.file, state: "released" });
  assert.deepEqual(f.ownership.outstanding().map((r) => r.kind), ["managed_suite"]);
  assert.equal(existsSync(f.suiteLease), true, "browser settlement never closes its borrowed suite");
  assert.equal((await settleBrowserAcquisition({ ...f, root })).state, "released");
});

test("borrowed runtime controller records do not hide or block preparation dependencies", (t) => {
  const f = fixture(t);
  recordRuntimeResource(f.runtime, { kind: "preparation_process", target: ownedProcess(process.pid) });
  assert.equal(f.ownership.outstanding().filter((r) => r.kind === "browser_stack").length, 1);
  f.ownership.record({ kind: "browser_stack", target: f.file, state: "released" });
  assert.deepEqual(f.ownership.outstanding().map((r) => r.kind), ["managed_suite"]);
  assert.equal(runtimeRecoveryResources(f.runtime).some((r) => r.kind === "preparation_process"), true);
});

test("failed registration acknowledgement retains dependency and exact acquisition proof", (t) => {
  const f = fixture(t);
  const ownership = preparationOwnership(f.runtime, (r) => { if (r.state === "released") throw new Error("acknowledgement failed"); });
  ownership.record({ kind: "browser_stack", target: f.file, state: "pending" });
  assert.throws(() => ownership.record({ kind: "browser_stack", target: f.file, state: "released" }), /acknowledgement/u);
  assert.equal(ownership.outstanding().length, 1);
  assert.equal(runtimeRecoveryResources(f.runtime).some((r) => r.target === f.file), true);
  assert.equal(readBrowserAcquisition(f.file, f.runtime).suite_lease, f.suiteLease);
});

test("pre-lease partial acquisition reaps real processes and exact ports before suite cleanup", async (t) => {
  const f = fixture(t);
  const child = spawn(process.execPath, ["-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore" });
  const closed = once(child, "close");
  const proof = ownedProcess(child.pid);
  f.cleanup.push(async () => { await stopOwnedProcess(proof); await closed; });
  recordAcquisitionProcess(f.file, "backend", child.pid);
  const ports = path.join(f.scratch, "ports"); mkdirSync(ports, { mode: 0o700 });
  const directory = path.join(ports, "port-19001"); mkdirSync(directory, { mode: 0o700 });
  recordAcquisitionPort(f.file, directory);
  const sibling = path.join(ports, "port-19002"); mkdirSync(sibling, { mode: 0o700 });
  await settleBrowserAcquisition({ ...f, root });
  await closed;
  assert.equal(existsSync(directory), false);
  assert.equal(existsSync(sibling), true);
  assert.equal(existsSync(f.suiteLease), true);
});

test("failed process cleanup retains acquisition and suite proof for explicit recovery", async (t) => {
  const f = fixture(t);
  recordAcquisitionProcess(f.file, "backend");
  const primary = new CommandFailure("startup sentinel", { failure_class: "infra", failure_reason: "service_start_error" });
  let caught;
  try {
    await withReviewResources({ acquire: async () => { throw primary; }, prepare: async () => {}, hold: async () => {},
      close: () => settleBrowserAcquisition({ ...f, root, stop: async () => { throw new Error("release sentinel"); } }), finish: async () => {} });
  } catch (error) { caught = error; }
  assert.equal(caught, primary);
  assert.equal(caught.cleanupFailures.length, 1);
  assert.equal(caught.cleanupFailures[0].settlement.state, "recovery_required");
  assert.equal(existsSync(path.join(path.dirname(f.file), "settlement.json")), false);
  assert.equal(runtimeRecoveryResources(f.runtime).length, 2);
  const calls = [];
  await settleBrowserAcquisition({ ...f, root, stop: async (proof) => { calls.push(proof.pid); } });
  assert.deepEqual(calls, [process.pid]);
});

test("wrong runtime or changed suite proof is rejected before destructive recovery", async (t) => {
  const f = fixture(t);
  let stopped = false;
  const stop = async () => { stopped = true; };
  recordAcquisitionProcess(f.file, "backend");
  await assert.rejects(settleBrowserAcquisition({ ...f, root, runtime: { ...f.runtime, runID: "wrong" }, stop }), (e) => e.failure_reason === "artifact_error");
  writeFileSync(f.suiteLease, "{}", { mode: 0o600 });
  await assert.rejects(settleBrowserAcquisition({ ...f, root, stop }), (e) => e.failure_reason === "artifact_error");
  assert.equal(stopped, false);
});

test("changed port proof preserves unrelated resources and remains recoverable", async (t) => {
  const f = fixture(t);
  const directory = path.join(f.scratch, "port-19001"); mkdirSync(directory, { mode: 0o700 });
  recordAcquisitionPort(f.file, directory);
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

test("producer death before ready publication leaves exact service process identity", async (t) => {
  const f = fixture(t);
  const child = spawn(process.execPath, ["--", path.join(root, "tools/harness/browser/browser-acquisition.mjs"), "launch", f.file, "producer",
    process.execPath, "-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore" });
  const closed = once(child, "close");
  const proof = ownedProcess(child.pid); f.cleanup.push(async () => { await stopOwnedProcess(proof); await closed; });
  const deadline = Date.now() + 10000;
  while (!readdirSync(path.dirname(f.file)).some((name) => name.startsWith("process-producer-"))) {
    assert.ok(Date.now() < deadline, "producer publishes ownership before work");
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  await settleBrowserAcquisition({ ...f, root });
  await closed;
  assert.equal(existsSync(path.join(f.sessionRoot, "stack.lease")), false);
});

test("concurrent settlement is shared and late acquisition cannot publish after cancellation", async (t) => {
  const f = fixture(t);
  recordAcquisitionProcess(f.file, "backend");
  const stopped = Promise.withResolvers();
  let count = 0;
  const options = { ...f, root, stop: async () => { count++; await stopped.promise; } };
  const first = settleBrowserAcquisition(options);
  const second = settleBrowserAcquisition(options);
  assert.equal(first, second);
  stopped.resolve();
  await Promise.all([first, second]);
  assert.equal(count, 1);
  assert.throws(() => recordAcquisitionProcess(f.file, "frontend"), (error) => error.failure_reason === "artifact_error");
  assert.equal(existsSync(path.join(f.sessionRoot, "stack.lease")), false);
});

for (const boundary of ["registration", "before-ready", "cancelled"]) test(`production provider settles ${boundary} acquisition failure without replacing its cause`, async (t) => {
  const f = fixture(t);
  // Retire the setup-only attempt; this case acquires through the real provider.
  f.ownership.record({ kind: "browser_stack", target: f.file, state: "released" });
  const lifecycle = path.join(f.scratch, "tools/harness/browser/start-web-e2e.sh");
  mkdirSync(path.dirname(lifecycle), { recursive: true, mode: 0o700 });
  const driver = path.join(root, "tools/harness/tests/fixture-stack-driver.mjs");
  writeFileSync(lifecycle, `#!/bin/sh\nexec "$NODE_BIN" -- '${driver}' "$@"\n`, { mode: 0o700 });
  const primary = new CommandFailure("registration sentinel", { failure_class: "artifact", failure_reason: "artifact_error" });
  let allocation, port, published = false;
  const controller = new AbortController();
  const provider = productionFixtureProviders({ root: f.scratch, suiteRuntime: f.runtime,
    signal: controller.signal,
    suiteController: { ensure: () => ({ environment: {}, leaseFile: f.suiteLease }) },
    runtimeEnvironment: { NODE_BIN: process.execPath, CARTULARY_FIXTURE_TEST_ACQUISITION_FAILURE: boundary === "cancelled" ? "late-ready" : "before-ready" },
    onOwnedResource: (record) => {
      f.ownership.record(record);
      if (record.state === "acquired") published = true;
      if (record.state === "released") {
        const portFile = path.join(path.dirname(path.dirname(record.target)), "port");
        if (existsSync(portFile)) port = Number(readFileSync(portFile, "utf8"));
      }
      if (record.kind === "browser_stack" && record.state === "pending") {
        allocation = record.target;
        f.cleanup.push(async () => { if (existsSync(allocation)) await settleBrowserAcquisition({ file: allocation, runtime: f.runtime, root: f.scratch }); });
        if (boundary === "registration") throw primary;
      }
    },
  }).browser_stack;
  const acquisition = provider.acquire({ affinityKey: "prelease", browserStage: "webserver-backed" });
  acquisition.catch(() => {});
  f.cleanup.unshift(async () => { controller.abort(primary); await acquisition.catch(() => {}); });
  if (boundary === "cancelled") {
    const deadline = Date.now() + 10000;
    while (!existsSync(path.join(path.dirname(path.dirname(allocation)), "cancellable"))) {
      assert.ok(Date.now() < deadline, "fixture installs its interruption handler");
      await new Promise((resolve) => setTimeout(resolve, 10));
    }
    controller.abort(primary);
  }
  await assert.rejects(acquisition, (error) => {
    if (boundary !== "before-ready") assert.equal(error, primary);
    else assert.match(error.message, /unclassified browser acquisition failure/u);
    assert.equal(error.cleanupFailures, undefined);
    return true;
  });
  assert.equal(existsSync(allocation), false, "the failed pending registration was settled");
  assert.deepEqual(f.ownership.outstanding().map((resource) => resource.kind), ["managed_suite"]);
  assert.equal(existsSync(f.suiteLease), true);
  assert.equal(published, false, "late completion cannot publish acquired readiness");
  if (boundary !== "registration") assert.ok(Number.isInteger(port) && port > 0, "partial acquisition created a real listener");
  if (port) assert.equal(await new Promise((resolve) => {
    const socket = createConnection({ host: "127.0.0.1", port });
    socket.on("connect", () => { socket.destroy(); resolve(true); });
    socket.on("error", () => resolve(false));
  }), false, "partial acquisition left no listener");
});
