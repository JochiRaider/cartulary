import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { FixtureBroker, createSuiteController, CleanupResults, dedicatedPoolProvider, DedicatedResourcePool, productionFixtureProviders } from "../scheduler/fixture-broker/index.mjs";
import { readBrowserAcquisition, settleBrowserAcquisition } from "../browser/browser-acquisition.mjs";
import { buildWorkGraph, runWorkGraph } from "../scheduler/work-graph/index.mjs";
import { createSuiteRuntime } from "../runtime/suite-runtime.mjs";
import { recordRuntimeResource, runtimeRecoveryResources } from "../runtime/resource-recovery.mjs";
import { mkdtempSync, mkdirSync, writeFileSync, existsSync, rmSync, readFileSync, symlinkSync } from "node:fs";
import { createConnection } from "node:net";
import os from "node:os";
import path from "node:path";
import { primaryPublicFailure, validateSchemaSync } from "../contract/index.mjs";

const owned = (release) => ({ ownership: "owned", resource_ids: ["fixture:owned"], release });
const brokerFor = (release, options = {}) => new FixtureBroker({ providers: {
  browser_stack: { acquire: async () => owned(release) },
}, ...options });

export async function assertFixtureCleanupCases() {
  for (const thrown of [undefined, null, 0, ""]) {
    let calls = 0;
    const broker = brokerFor(() => { calls += 1; throw thrown; });
    const lease = await broker.acquire("browser_stack");
    await assert.rejects(lease.release());
    await assert.rejects(broker.close());
    assert.equal(calls, 1);
    assert.equal(broker.hasUnresolvedCleanup(), true, "a falsy thrown value is still failed physical settlement");
    const results = new CleanupResults();
    const failure = await results.attempt("provider_close", () => Promise.reject(thrown));
    assert.equal(failure.failure_reason, "cleanup_error");
    assert.equal(results.results[0].outcome, "failed");
  }
  for (const asynchronous of [false, true]) for (const failed of [false, true]) {
    let calls = 0;
    const failure = new Error("deliberate cleanup failure");
    const release = ({ healthy }) => {
      assert.equal(healthy, false); calls += 1;
      if (failed) { if (asynchronous) return Promise.reject(failure); throw failure; }
      return asynchronous ? Promise.resolve() : undefined;
    };
    const broker = brokerFor(release);
    const lease = await broker.acquire("browser_stack");
    const first = lease.release({ healthy: false });
    assert.equal(first, lease.release({ healthy: false }), "release joins the original promise");
    if (failed) {
      await assert.rejects(first);
      await assert.rejects(lease.release({ healthy: false }));
      const close = broker.close();
      assert.equal(close, broker.close());
      await assert.rejects(close);
      assert.equal(broker.hasUnresolvedCleanup(), true);
    } else {
      await first; await broker.close();
      assert.equal(broker.hasUnresolvedCleanup(), false);
    }
    assert.equal(calls, 1);
    await assert.rejects(lease.release(), /contradictory/u);
    assert.equal(calls, 1);
  }
  let allocations = 0;
  const dispositions = [];
  const shared = new FixtureBroker({ providers: { browser_stack: { acquire: async () => ({
    ...owned(({ healthy }) => dispositions.push(healthy)), resource: ++allocations,
  }) } } });
  const a = await shared.acquire("browser_stack", { affinityKey: "shared" });
  const b = await shared.acquire("browser_stack", { affinityKey: "shared" });
  await a.release({ healthy: false });
  const replacement = await shared.acquire("browser_stack", { affinityKey: "shared" });
  assert.notEqual(a.resource, replacement.resource);
  await b.release({ retainWarm: true });
  assert.deepEqual(dispositions, [false], "healthy final borrower cannot recycle a tainted allocation");
  await replacement.release({ retainWarm: true });
  await shared.close();
  assert.deepEqual(dispositions, [false, true]);

  let independent = 0;
  const failing = new FixtureBroker({ providers: {
    managed_process: { acquire: async ({ unitID }) => owned(() => {
      if (unitID === "bad") throw new Error("bad resource"); independent += 1;
    }) },
  } });
  await failing.acquire("managed_process", { unitID: "good" });
  await failing.acquire("managed_process", { unitID: "bad" });
  await assert.rejects(failing.close());
  assert.equal(independent, 1, "one failed release cannot strand an independent resource");
  assert.equal(failing.cleanupResults.results.length, 2);

  let cleaned = 0;
  const publication = brokerFor(() => { cleaned += 1; }, { recordSink: () => { throw new Error("publication"); } });
  await assert.rejects(publication.acquire("browser_stack"), (error) => error.failure_reason === "artifact_error");
  await assert.rejects(publication.close());
  assert.equal(cleaned, 1, "failed publication still releases acquired ownership once");
  assert.equal(publication.hasUnresolvedCleanup(), false);
  const failedTransfer = new FixtureBroker({ providers: { browser_stack: { acquire: () => {
    throw Object.assign(new Error("publication"), { failure_class: "artifact", failure_reason: "artifact_error",
      cleanupFailures: [new Error("rollback failure"), new Error("recovery publication failure")] });
  } } } });
  await assert.rejects(failedTransfer.acquire("browser_stack"), (error) => error.failure_reason === "artifact_error");
  await assert.rejects(failedTransfer.close());
  assert.equal(failedTransfer.cleanupResults.results[0].operation, "fixture_acquisition_cleanup");
  assert.equal(failedTransfer.cleanupResults.results[0].failures.length, 2);
  assert.equal(failedTransfer.hasUnresolvedCleanup(), true);
  let finishAcquisition;
  const acquiring = new FixtureBroker({ providers: { browser_stack: { acquire: async () => {
    await new Promise((resolve) => { finishAcquisition = resolve; });
    throw Object.assign(new Error("publication"), { cleanupFailures: [new Error("rollback")] });
  } } } });
  const acquisition = acquiring.acquire("browser_stack");
  const acquisitionRejection = assert.rejects(acquisition);
  const acquisitionClose = acquiring.close();
  finishAcquisition();
  await acquisitionRejection;
  await assert.rejects(acquisitionClose, "close retains cleanup failures that settle during admission drain");
  let sharedStops = 0, rejectPublication = true;
  const sharedPublication = brokerFor(() => { sharedStops += 1; }, { recordSink: (record) => {
    if (rejectPublication && record.state === "released") { rejectPublication = false; throw new Error("logical release publication"); }
  } });
  const firstBorrower = await sharedPublication.acquire("browser_stack", { affinityKey: "publication" });
  const lastBorrower = await sharedPublication.acquire("browser_stack", { affinityKey: "publication" });
  await assert.rejects(firstBorrower.release(), (error) => error.failure_reason === "artifact_error");
  await assert.rejects(lastBorrower.release(), (error) => error.failure_reason === "artifact_error");
  await assert.rejects(sharedPublication.close());
  assert.equal(sharedStops, 1);
  assert.equal(sharedPublication.hasUnresolvedCleanup(), false, "publication failure does not manufacture a live allocation");
  assert.equal(sharedPublication.cleanupResults.results[0].failures[0].failure_reason, "artifact_error");
  for (const ownership of ["owned", "borrowed"]) {
    const invalid = new FixtureBroker({ providers: { managed_process: { acquire: async () => ({ ownership, resource_ids: [] }) } } });
    await assert.rejects(invalid.acquire("managed_process"), /requires/u);
    await assert.rejects(invalid.close());
  }
  let detached = 0;
  const borrowed = new FixtureBroker({ providers: { managed_process: { acquire: async () => ({ ownership: "borrowed", resource_ids: [], detach() { detached += 1; } }) } } });
  await (await borrowed.acquire("managed_process")).release({ healthy: false });
  await borrowed.close(); assert.equal(detached, 1);
  const none = new FixtureBroker(); await (await none.acquire("none")).release(); await none.close();
  let suiteStops = 0;
  const controller = createSuiteController(() => ({ close() { suiteStops += 1; throw new Error("suite cleanup"); } }));
  controller.ensure(); const suiteClose = controller.close();
  assert.equal(suiteClose, controller.close()); await assert.rejects(suiteClose);
  await assert.rejects(controller.close()); assert.equal(suiteStops, 1);

  const poolDestroyed = [];
  const pool = new DedicatedResourcePool({ create: async () => ({}), healthy: async () => true,
    reset: async () => {}, destroy: async (resource) => poolDestroyed.push(resource) });
  const pooled = new FixtureBroker({ providers: { postgres_dedicated: dedicatedPoolProvider(pool) } });
  await (await pooled.acquire("postgres_dedicated")).release({ healthy: false });
  await pooled.close(); assert.ok(poolDestroyed.length >= 1);
  let resetReady;
  const resetting = new Promise((resolve) => { resetReady = resolve; });
  const drained = [];
  const drainingPool = new DedicatedResourcePool({ create: async () => ({}), healthy: async () => true,
    reset: () => resetting, destroy: (resource) => drained.push(resource) });
  const resource = await drainingPool.acquire();
  const releasing = drainingPool.release(resource);
  const closing = drainingPool.close();
  resetReady(); await releasing; await closing;
  assert.equal(drainingPool.ready.length, 0);
  assert.equal(drainingPool.leased.size, 0);
  assert.equal(drained.filter((entry) => entry === resource).length, 1);
}

function workUnit() {
  return { unit_id: "cleanup:work", owner_id: "harness.command_surface", kind: "runner",
    command: { executable: "true", args: [], environment: { CARTULARY_BROWSER_RELEASE_AFFINITY: "1" } },
    needs: [], resource_claims: { cpu: 1 }, fixture_lease: "browser_stack", service_dependencies: [],
    cache_policy: "none", timeout_ms: 1000, current_run_evidence_outputs: [],
    failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" }, estimated_work_ms: 1 };
}

export async function assertSchedulerCleanupCases() {
  assert.equal(primaryPublicFailure([
    { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" },
    { failure_class: "timing", failure_reason: "timeout_failure", lifecycle_step: "cleanup_finalizers" },
  ]).failure_class, "interrupted");
  assert.equal(primaryPublicFailure([
    { failure_class: "timing", failure_reason: "timeout_failure", lifecycle_step: "cleanup_finalizers", scheduler_event_sequence: 0 },
    { failure_class: "harness", failure_reason: "cleanup_error", scheduler_event_sequence: 1 },
  ]).failure_reason, "timeout_failure");
  for (const productFailed of [false, true]) for (const cleanupFailed of [false, true]) {
    const order = [];
    const broker = brokerFor(() => { order.push("fixture"); if (cleanupFailed) throw new Error("private sentinel must not enter receipt"); });
    let cacheStores = 0;
    const result = await runWorkGraph({ graph: buildWorkGraph([workUnit()]), capacities: new Map([["cpu", 1]]),
      cwd: process.cwd(), environment: {}, fixtureBroker: broker,
      executeUnit: async () => { order.push("execute"); return productFailed
        ? { status: "failed", exit_code: 10, failure_class: "product", failure_reason: "test_assertion_failure" }
        : { status: "passed", exit_code: 0 }; },
      onUnitTerminal: () => order.push("publish"),
      cache: { validateGraph() {}, lookup: async () => ({ outcome: "miss" }), store: async () => { cacheStores += 1; order.push("cache"); } },
      cleanup: () => order.push("services"), finalize: () => order.push("runtime"),
    });
    if (!productFailed && !cleanupFailed) assert.equal(result.events.at(-1).failure_class, undefined);
    const terminal = result.unit_results["cleanup:work"];
    assert.equal(terminal.failure_reason ?? null, productFailed ? "test_assertion_failure" : cleanupFailed ? "cleanup_error" : null);
    assert.equal(terminal.exit_code, productFailed ? 10 : cleanupFailed ? 12 : 0);
    assert.equal(cacheStores, !productFailed && !cleanupFailed ? 1 : 0);
    assert.ok(order.indexOf("fixture") < order.indexOf("publish"));
    if (!cleanupFailed) assert.ok(order.indexOf("fixture") < order.indexOf("services"));
    else assert.equal(order.includes("services"), false);
    assert.equal(JSON.stringify(result.cleanup_results).includes("private sentinel"), false);
  }
  const broker = brokerFor(() => { throw new Error("cleanup"); });
  const result = await runWorkGraph({ graph: buildWorkGraph([workUnit()]), capacities: new Map([["cpu", 1]]),
    cwd: process.cwd(), environment: {}, fixtureBroker: broker,
    executeUnit: async () => { throw Object.assign(new Error("execution"), { failure_class: "infra", failure_reason: "preflight_error" }); },
    finalize: async () => { throw new Error("independent finalizer"); },
  });
  assert.equal(result.unit_results["cleanup:work"].failure_reason, "preflight_error");
  assert.equal(result.cleanup_error.errors.length, 2);
  let warmCacheStores = 0;
  const warm = brokerFor(() => { throw new Error("warm close failure"); });
  const warmUnit = workUnit(); warmUnit.command.environment = {};
  const warmResult = await runWorkGraph({ graph: buildWorkGraph([warmUnit]), capacities: new Map([["cpu", 1]]),
    cwd: process.cwd(), environment: {}, fixtureBroker: warm,
    executeUnit: async () => ({ status: "passed", exit_code: 0 }),
    cache: { validateGraph() {}, lookup: async () => ({ outcome: "miss" }), store: async () => { warmCacheStores += 1; } },
  });
  assert.equal(warmResult.status, "failed");
  assert.equal(warmCacheStores, 0, "a failed retained allocation cannot admit a passing cache result");
  const hostResult = await runWorkGraph({ graph: buildWorkGraph([workUnit()]), capacities: new Map([["cpu", 1]]),
    cwd: process.cwd(), environment: {}, fixtureBroker: brokerFor(() => {}),
    hostAdmission: async () => ({ token: "opaque", release: () => { throw Object.assign(new Error("deadline"), { code: "ETIMEDOUT" }); } }),
    executeUnit: async () => ({ status: "passed", exit_code: 0 }),
  });
  assert.equal(hostResult.cleanup_error.failure_reason, "timeout_failure");
  assert.equal(hostResult.unit_results["cleanup:work"].exit_code, 13);
  assert.equal(hostResult.cleanup_results.find((step) => step.operation === "host_release").failure_reason, "timeout_failure");
}

export function assertRecoveryCases(repoRoot) {
  const runRoot = mkdtempSync(path.join(os.tmpdir(), "fixture-cleanup-results-"));
  const runID = path.basename(runRoot);
  const runtime = createSuiteRuntime({ repoRoot, runRoot, runID });
  try {
    const proof = runtime.privatePath("owned", "lease.json");
    mkdirSync(path.dirname(proof), { mode: 0o700 });
    writeFileSync(proof, "{}", { mode: 0o600 });
    recordRuntimeResource(runtime, { kind: "browser_stack", target: proof });
    assert.throws(() => runtime.close(), /unresolved/u);
    const detail = runtime.privatePath("unit-output"); mkdirSync(detail, { mode: 0o700 });
    writeFileSync(path.join(detail, "private.json"), "credential-sentinel", { mode: 0o600 });
    runtime.preserveRecovery();
    assert.equal(existsSync(detail), false);
    assert.equal(existsSync(proof), true);
    assert.equal(runtimeRecoveryResources(runtime).length, 1);
    const receipt = new CleanupResults(); receipt.record("fixture_release", { leaseID: "lease-1", error: new Error("credential-sentinel") });
    receipt.publish(runRoot, runID);
    const payload = JSON.parse(readFileSync(path.join(runRoot, "cleanup-results.json"), "utf8"));
    for (const invalid of [
      { ...payload.results[0], unit_id: "/private/path" },
      { ...payload.results[0], failure_class: "timing", failure_reason: "cleanup_error" },
      { ...payload.results[0], private_message: "credential-sentinel" },
      { ...payload.results[0], artifact_refs: ["../private.json"] },
    ]) assert.throws(() => validateSchemaSync(payload.schema_id, { ...payload, results: [invalid] }));
    assert.equal(readFileSync(path.join(runRoot, "cleanup-results.json"), "utf8").includes("credential-sentinel"), false);
    const output = execFileSync(process.execPath, [path.join(repoRoot, "tools/harness/diagnostics/explain-run-cli.mjs"), "--results-dir", runRoot], { encoding: "utf8" });
    assert.match(output, /failed=1/u); assert.match(output, /reason=cleanup_error/u);
    assert.equal(output.includes("credential-sentinel"), false);
    assert.throws(() => recordRuntimeResource(runtime, { kind: "browser_stack", target: path.join(runRoot, "outside.json") }));
    const link = runtime.privatePath("symlink"); symlinkSync(runRoot, link);
    assert.throws(() => runtime.preserveRecovery()); rmSync(link);
    recordRuntimeResource(runtime, { kind: "browser_stack", target: proof, state: "released" });
  } finally { runtime.close(); rmSync(runRoot, { recursive: true, force: true }); }
}

export async function assertProductionCleanupCases(repoRoot) {
  const cases = [
    ...[false, true].flatMap((publicationFailed) => [false, true].map((cleanupFailed) => ({ publicationFailed, cleanupFailed, releasePublicationFailed: false }))),
    { publicationFailed: false, cleanupFailed: false, releasePublicationFailed: true },
  ];
  for (const { publicationFailed, cleanupFailed, releasePublicationFailed } of cases) {
    const fixtureRoot = mkdtempSync(path.join(os.tmpdir(), "fixture-production-cleanup-"));
    const runtime = createSuiteRuntime({ repoRoot, runRoot: fixtureRoot, runID: "production-cleanup-test" });
    const lifecycle = path.join(fixtureRoot, "tools/harness/browser/start-web-e2e.sh");
    mkdirSync(path.dirname(lifecycle), { recursive: true, mode: 0o700 });
    const driver = path.join(repoRoot, "tools/harness/tests/fixture-stack-driver.mjs").replaceAll("'", "'\"'\"'");
    writeFileSync(lifecycle, `#!/bin/sh\nexec "$NODE_BIN" -- '${driver}' "$@"\n`, { mode: 0o700 });
    const environment = { NODE_BIN: process.execPath };
    const suiteLease = runtime.privatePath("test-services", "suite.json");
    mkdirSync(path.dirname(suiteLease), { mode: 0o700 });
    const resultRoot = path.join(fixtureRoot, "results");
    writeFileSync(suiteLease, JSON.stringify({ schema_id: "cartulary.test_services.lease.v1", run_id: runtime.runID, suite_id: "controlled-suite",
      lease_id: runtime.leaseID, result_root: resultRoot, run_root: path.join(resultRoot, runtime.runID), target: "browser-e2e",
      mode: "owned", ownership_mode: "owned", owner_pid: process.pid, created_at: new Date().toISOString(),
      resources: [], proof_labels: {}, proof_prefixes: {}, cleanup_state: "not_started" }), { mode: 0o600 });
    recordRuntimeResource(runtime, { kind: "managed_suite", target: suiteLease });
    let stopCount = 0;
    const closedDetail = runtime.privatePath("closed-consumer");
    mkdirSync(closedDetail, { mode: 0o700 });
    writeFileSync(path.join(closedDetail, "credentials.json"), "private-sentinel", { mode: 0o600 });
    let leaseFile, port;
    const broker = new FixtureBroker({ providers: productionFixtureProviders({ root: fixtureRoot,
      suiteRuntime: runtime, suiteController: { ensure: () => ({ environment: {}, leaseFile: suiteLease }) },
      settleBrowser: async (options) => { stopCount++; if (cleanupFailed) throw new Error("deliberate owner stop failure"); return settleBrowserAcquisition(options); },
      runtimeEnvironment: environment, onOwnedResource: (resource) => {
        if (resource.kind === "browser_stack" && resource.state === "released" && releasePublicationFailed) {
          throw new Error("release publication failure");
        }
        recordRuntimeResource(runtime, resource);
        if (resource.kind === "browser_stack" && resource.state === "acquired") {
          leaseFile = resource.target;
          port = JSON.parse(readFileSync(readBrowserAcquisition(leaseFile).ready_lease, "utf8")).port;
          if (publicationFailed) throw new Error("acquisition publication failure");
        }
      },
    }) });
    try {
      const result = await runWorkGraph({ graph: buildWorkGraph([workUnit()]), capacities: new Map([["cpu", 1]]),
        cwd: fixtureRoot, environment: {}, fixtureBroker: broker,
        executeUnit: async (_unit, { fixtureLease }) => {
          port = JSON.parse(readFileSync(fixtureLease.resource.environment.CARTULARY_WEB_E2E_SESSION_LEASE_FILE, "utf8")).port;
          return { status: "failed", failure_class: "product", failure_reason: "test_assertion_failure", exit_code: 10 };
        },
        finalize: ({ unresolved }) => { if (unresolved) runtime.preserveRecovery(); },
      });
      assert.equal(result.unit_results["cleanup:work"].failure_reason, publicationFailed ? "artifact_error" : "test_assertion_failure");
      assert.equal(stopCount, 1);
      const unresolved = cleanupFailed || releasePublicationFailed;
      assert.equal(existsSync(leaseFile), unresolved);
      assert.equal(broker.hasUnresolvedCleanup(), unresolved);
      if (unresolved) {
        assert.equal(runtimeRecoveryResources(runtime).length, 2);
        assert.equal(existsSync(closedDetail), false, "closed-consumer credentials are removed while exact recovery proof remains");
        if (releasePublicationFailed) {
          assert.equal(result.cleanup_error.failure_reason, "artifact_error");
          await assert.rejects(broker.close());
          assert.equal(stopCount, 1, "reobserving publication failure cannot repeat physical cleanup");
        }
        // Explicit resource-owner recovery is separate from the memoized broker
        // lifetime. It consumes the still-live exact lease, with no heuristic.
        await settleBrowserAcquisition({ root: fixtureRoot, file: leaseFile, runtime, environment, recovery: true });
        recordRuntimeResource(runtime, { kind: "browser_stack", target: leaseFile, state: "released" });
      }
      const connected = await new Promise((resolve) => {
        const socket = createConnection({ host: "127.0.0.1", port });
        socket.on("connect", () => { socket.destroy(); resolve(true); });
        socket.on("error", () => resolve(false));
      });
      assert.equal(connected, false, "the exact owned listener must be gone after cleanup/recovery");
    } finally {
      if (leaseFile && existsSync(leaseFile)) {
        await settleBrowserAcquisition({ root: fixtureRoot, file: leaseFile, runtime, environment, recovery: true });
        recordRuntimeResource(runtime, { kind: "browser_stack", target: leaseFile, state: "released" });
      }
      recordRuntimeResource(runtime, { kind: "managed_suite", target: suiteLease, state: "released" });
      runtime.close(); rmSync(fixtureRoot, { recursive: true, force: true });
    }
  }
}
