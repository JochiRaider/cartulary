import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, writeFileSync, readFileSync, rmSync, symlinkSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import { Worker } from "node:worker_threads";
import { actualCriticalPath, intervalUnion } from "../../evidence-accounting/index.mjs";
import { reduceCanonicalUnitIntervals } from "../../evidence-accounting/index.mjs";
import { buildWorkGraph } from "../../scheduler/work-graph/model.mjs";
import { runWorkGraph } from "../../scheduler/work-graph/scheduler.mjs";
import { parseProcessStat, linuxUnits, parsePSI, createLinuxResourceAdapter } from "../resource-linux.mjs";
import { resourceProjection } from "../resource-projection.mjs";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { otlpProjection } from "../otlp-projection.mjs";
import { validateCanonicalRun } from "../canonical-evidence.mjs";
import { performanceExplanation, loadRetainedObservability } from "../observability.mjs";
import { formatPerformanceExplanation } from "../performance-presentation.mjs";
import { exportRetainedObservability } from "../otel-export-cli.mjs";
import { hostClaims, processIdentity, transactAdmission } from "../../runtime/host-admission.mjs";
import { createResourceCollector, instrumentationPolicy } from "../resource-collector.mjs";
import { executionRecordsDigest } from "../../runtime/execution-observations.mjs";

const digest = `sha256:${"a".repeat(64)}`;
function temporary(t) {
  const dir = mkdtempSync(path.join(os.tmpdir(), "cartulary-instrumentation-"));
  t.after(() => rmSync(dir, { recursive: true, force: true })); return dir;
}
const unit = (id, needs = []) => ({ unit_id: id, owner_id: "harness.evidence_accounting", kind: "runner",
  command: { executable: "true", args: [], environment: {} }, needs,
  resource_claims: { cpu: 1, process: 1 }, fixture_lease: "none", service_dependencies: [],
  cache_policy: "none", timeout_ms: 1000, current_run_evidence_outputs: [],
  failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" }, estimated_work_ms: 10 });

test("queue time changes the dependency path; overlapping wall time is a union", () => {
  const graph = { units: [unit("long"), unit("a"), unit("b", ["a"])] };
  const intervals = new Map([
    ["long", { start: 0, end: 50, queue_ms: 0 }],
    ["a", { start: 20, end: 30, queue_ms: 20 }],
    ["b", { start: 50, end: 61, queue_ms: 20 }],
  ]);
  assert.deepEqual(actualCriticalPath(graph, intervals), ["a", "b"]);
  assert.equal(intervals.get("b").queue_ms, 20);
  assert.equal(intervalUnion([...intervals.values()]), 61);
  assert.deepEqual(actualCriticalPath(graph, new Map()), ["a"]);
});

test("scheduler phases preserve unit outcomes and reconstruct from the canonical reader", async (t) => {
  const graph = buildWorkGraph([unit("a"), unit("b", ["a"])]);
  const result = await runWorkGraph({ graph, capacities: new Map([["cpu", 2], ["process", 2]]),
    hostAdmission: async () => ({ release: async () => {} }),
    executeUnit: async () => ({ status: "passed", exit_code: 0 }) });
  const dir = temporary(t), file = path.join(dir, "unit-events.ndjson");
  writeFileSync(file, result.events.map((entry) => JSON.stringify(entry)).join("\n") + "\n");
  const reduced = await reduceCanonicalUnitIntervals(file);
  assert.equal(reduced.registrations.size, 2);
  assert.equal(reduced.starts.size, 2);
  for (const phase of reduced.phases.values()) assert.ok(phase.end_ms >= phase.start_ms);
  assert.ok([...reduced.phases.values()].some((phase) => phase.phase === "host_admission"));
  const duplicate = result.events.find((entry) => entry.event === "phase_started");
  const malformed = [...result.events.slice(0, duplicate.seq), { ...duplicate, seq: duplicate.seq + 1 }];
  writeFileSync(file, malformed.map((entry) => JSON.stringify(entry)).join("\n") + "\n");
  await assert.rejects(reduceCanonicalUnitIntervals(file), /phase start/u);
});

test("Linux counter adapter handles command parentheses, units, PID reuse and unavailable I/O", () => {
  const fields = Array(22).fill("0");
  Object.assign(fields, { 0: "S", 1: "1", 11: "100", 12: "25", 17: "4", 19: "900", 21: "10" });
  const stat = () => `123 (private ) command) ${fields.join(" ")}`;
  assert.equal(parseProcessStat(stat()).start, "900");
  assert.equal(parseProcessStat(stat()).user_ticks, 100);
  const auxv = Buffer.alloc(48);
  auxv.writeBigUInt64LE(6n, 0); auxv.writeBigUInt64LE(4096n, 8);
  auxv.writeBigUInt64LE(17n, 16); auxv.writeBigUInt64LE(100n, 24);
  assert.deepEqual(linuxUnits(auxv), { pageBytes: 4096, ticksPerSecond: 100 });
  assert.deepEqual(parsePSI("some avg10=0.00 avg60=0.00 avg300=0.00 total=123\n"), { some_us: 123 });
  const adapter = createLinuxResourceAdapter({ link: () => "pid:[1]", read: (file) => {
    if (file.endsWith("boot_id")) return "private-boot";
    if (file.endsWith("auxv")) return auxv;
    if (file.endsWith("/stat")) return stat();
    if (file.endsWith("/status")) return "VmHWM:\t64 kB\n";
    throw Object.assign(new Error("denied"), { code: "EACCES" });
  } });
  const proof = adapter.proof(123), sample = adapter.processSample(proof);
  assert.equal(sample.metrics.cpu_user_us.value, 1_000_000);
  assert.equal(sample.metrics.rss_bytes.value, 40960);
  assert.deepEqual(sample.metrics.io_read_bytes, { value: null, availability: "permission_denied" });
  fields[19] = "901";
  assert.equal(adapter.processSample(proof).availability, "process_gone");
});

function retainedSamples(t, records) {
  const dir = temporary(t), bytes = records.map((record, index) => JSON.stringify({
    schema_id: "cartulary.harness_resource_sample.v1", seq: index + 1, segment: 0,
    scope: "process", scope_ref: "process:1", identity_digest: digest, availability: "available", ...record,
  })).join("\n") + "\n";
  const identity = { run_id: "synthetic", source_digest: digest, graph_digest: digest, policy_digest: digest };
  const index = { schema_id: "cartulary.harness_resource_index.v2", ...identity,
    status: "complete", clock: "graph_process_monotonic", coverage: "observed_partial",
    samples: records.length, sample_bytes: Buffer.byteLength(bytes), sample_digest: `sha256:${createHash("sha256").update(bytes).digest("hex")}`,
    omitted_observations: 0, discovery_truncations: 0, failed_sweeps: 0, sweeps: records.length,
    processes: [{ process_ref: "process:1", identity_digest: digest }],
    observer: { cpu_user_us: 0, cpu_system_us: 0, heap_peak_bytes: 0, read_bytes: 0, write_bytes: 0, maximum_sweep_ms: 0, shutdown_ms: 0, gate_cpu: "not_observed" } };
  atomicLocalFile(path.join(dir, "diagnostics/resource-samples.ndjson"), bytes);
  atomicLocalFile(path.join(dir, "diagnostics/resource-index.json"), JSON.stringify(index));
  return { dir, run: { manifest: { ...identity, instrumentation: { mode: "basic", policy_digest: digest } }, registrations: new Map([["a", {}]]) } };
}

test("process relationship proof must resolve and conflicting registration excludes attributed totals", (t) => {
  const f = retainedSamples(t, [{ elapsed_ms: 0, metrics: { cpu_user_us: metric(0), cpu_system_us: metric(0) } },
    { elapsed_ms: 1000, metrics: { cpu_user_us: metric(100), cpu_system_us: metric(10) } }]);
  f.run.registrations = new Map([["a", { row_ids: [] }], ["b", { row_ids: [] }]]);
  const relationships = ["a", "b"].map((unit_id) => {
    const fact = { kind: "relationship", identity_digest: digest, unit_id, allocation_ref: null, invocation_id: null, provenance: "registered" };
    return { ...fact, ref: `relationship:${createHash("sha256").update(JSON.stringify(fact)).digest("hex")}`, observed_elapsed_ms: 0 };
  });
  const execution = { records: relationships };
  assert.deepEqual(resourceProjection(f.dir, f.run, execution).unit_cpu_lower_bounds, []);
  execution.records = [relationships[0]];
  assert.deepEqual(resourceProjection(f.dir, f.run, execution).unit_cpu_lower_bounds, [{ unit_id: "a", observed_attributed_cpu_us: 110 }]);
  execution.records[0] = { ...relationships[0], identity_digest: `sha256:${"b".repeat(64)}` };
  assert.throws(() => resourceProjection(f.dir, f.run, execution), /physical proof/u);
});
const metric = (value) => ({ value, availability: "available" });
test("streaming resource projection handles resets and measurement gaps deterministically", (t) => {
  const f = retainedSamples(t, [
    { elapsed_ms: 0, metrics: { cpu_user_us: metric(100), rss_bytes: metric(1000) } },
    { elapsed_ms: 2000, metrics: { cpu_user_us: metric(300), rss_bytes: metric(2000) } },
    { elapsed_ms: 4000, metrics: { cpu_user_us: metric(10) } },
    { elapsed_ms: 6000, segment: 1, metrics: { cpu_user_us: metric(10000) } },
  ]);
  const first = resourceProjection(f.dir, f.run);
  assert.deepEqual(resourceProjection(f.dir, f.run), first);
  assert.equal(first.scopes[0].metrics.cpu_user_us.observed_delta, 200);
  assert.equal(first.scopes[0].resets, 1);
  assert.equal(first.sampled_simultaneous_rss_sum_bytes, 2000);
  assert.equal(first.scopes[0].cpu_percent_one_core, null);
  const file = path.join(f.dir, "diagnostics/resource-samples.ndjson");
  rmSync(file); symlinkSync("/dev/null", file);
  assert.throws(() => resourceProjection(f.dir, f.run));
});

test("an omitted resource counter breaks its delta interval", (t) => {
  const f = retainedSamples(t, [
    { elapsed_ms: 0, metrics: { io_read_bytes: metric(100), rss_bytes: metric(10) } },
    { elapsed_ms: 2000, metrics: { rss_bytes: metric(20) } },
    { elapsed_ms: 4000, metrics: { io_read_bytes: metric(900), rss_bytes: metric(10) } },
  ]);
  const io = resourceProjection(f.dir, f.run).scopes[0].metrics.io_read_bytes;
  assert.equal(io.observed_delta, null);
  assert.equal(io.observed_interval_ms, 0);
  assert.equal(io.maximum, 900);
});

test("OTLP has invocation parentage and closed safe metric dimensions", () => {
  const run = { manifest: { run_id: "private-run", target: "test-slice", command_id: "cartulary.harness.command.test_slice.v1", source_digest: digest, started_at: "2026-10-09T00:00:00.000Z" },
    summary: { status: "pass", wall_duration_ms: 10, actual_dependency_critical_path_ms: 8, timing_accounting: { resource_blocking_ms: 2, unattributed_ms: 1 } },
    terminal: new Map([["raw_test_symbol", { event: "completed", monotonic_ms: 8, status: "passed" }]]), started: new Map([["raw_test_symbol", 1]]),
    registrations: new Map([["raw_test_symbol", { kind: "runner" }]]), intervals: new Map([["raw_test_symbol", { queue_ms: 1 }]]) };
  const value = otlpProjection(run), text = JSON.stringify(value);
  assert.equal(text.includes("private-run"), false); assert.equal(text.includes("raw_test_symbol"), false);
  assert.ok(text.includes("service.name"));
  assert.equal(value.traceOTLP.resourceSpans[0].resource.attributes[0].value.stringValue, "cartulary.harness");
  const spans = value.traceOTLP.resourceSpans[0].scopeSpans[0].spans;
  assert.equal(spans[1].parentSpanId, spans[0].spanId);
  assert.equal(value.metricsOTLP.resourceMetrics[0].scopeMetrics[0].metrics.length, 5);
});

test("terminating an observer thread cannot strand its measurement fence", async (t) => {
  const root = temporary(t);
  const worker = new Worker(`
    const { parentPort, workerData } = require('node:worker_threads');
    import(workerData.gate).then(async ({ createObservationGate }) => {
      const gate = createObservationGate({ root: workerData.root, capacities: { cpu: 1, process: 1 } });
      await gate.acquire({ signal: new AbortController().signal });
      parentPort.postMessage('fenced');
      parentPort.on('message', () => {});
    });
  `, { eval: true, workerData: { root, gate: new URL("../../runtime/observation-gate.mjs", import.meta.url).href } });
  t.after(() => worker.terminate());
  assert.equal(await new Promise((resolve, reject) => { worker.once("message", resolve); worker.once("error", reject); }), "fenced");
  const request = { root, operation: "admit", token: "b".repeat(32), owner: processIdentity(),
    mode: "exclusive", claims: hostClaims(), capacities: hostClaims({ cpu: 1, process: 1 }), parent: null };
  assert.equal(transactAdmission(request), false);
  await worker.terminate();
  assert.equal(transactAdmission(request), true);
});

test("four sibling collectors remain quiet during exclusive measurement and resume with new segments", { timeout: 20000 }, async (t) => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-observer-quietness-")), admissionRoot = path.join(root, "admission"), collectors = [];
  const capacities = hostClaims({ cpu: 4, process: 4, memory_mb: 4096 });
  const request = { root: admissionRoot, operation: "admit", token: "c".repeat(32), owner: processIdentity(),
    mode: "exclusive", claims: hostClaims(), capacities, parent: null };
  t.after(async () => {
    try { transactAdmission({ root: admissionRoot, operation: "release", token: request.token }); }
    finally {
      await Promise.all(collectors.map((collector) => collector.stop()));
      rmSync(root, { recursive: true, force: true });
    }
  });
  assert.equal(transactAdmission(request), true);
  const { digest: policyDigest } = instrumentationPolicy();
  const epoch = process.hrtime.bigint();
  for (let index = 0; index < 4; index++) collectors.push(createResourceCollector({
    runRoot: path.join(root, `run-${index}`), capacities, epoch, admissionRoot,
    manifest: { run_id: `synthetic-${index}`, source_digest: digest, graph_digest: digest,
      instrumentation: { mode: "basic", policy_digest: policyDigest } },
  }));
  const waitUntil = async (predicate) => {
    const deadline = performance.now() + 7000;
    while (!predicate()) {
      assert.ok(performance.now() < deadline, "collector transition exceeded bounded test deadline");
      await new Promise((resolve) => setTimeout(resolve, 50));
    }
  };
  await waitUntil(() => collectors.every((collector) => collector.snapshot().availability === "paused_measurement"));
  for (const collector of collectors) assert.equal(collector.snapshot().sampled_elapsed_ms, null);
  transactAdmission({ root: admissionRoot, operation: "release", token: request.token });
  await waitUntil(() => collectors.every((collector) => collector.snapshot().availability === "available"));
  await Promise.all(collectors.map((collector) => collector.pause()));
  const paused = collectors.map((collector) => collector.snapshot());
  assert.ok(paused.every((snapshot) => snapshot.availability === "paused_measurement" && snapshot.rss_bytes === null));
  // Acknowledged explicit pause also drains the fence, so measurement admission
  // succeeds while all four observers remain alive.
  assert.equal(transactAdmission(request), true);
  await new Promise((resolve) => setTimeout(resolve, 2100));
  assert.deepEqual(collectors.map((collector) => collector.snapshot().sampled_elapsed_ms), paused.map((snapshot) => snapshot.sampled_elapsed_ms));
  transactAdmission({ root: admissionRoot, operation: "release", token: request.token });
  collectors.forEach((collector) => collector.resume());
  await waitUntil(() => collectors.every((collector, index) => collector.snapshot().sampled_elapsed_ms > paused[index].sampled_elapsed_ms));
  await Promise.all(collectors.map((collector) => collector.stop()));
  for (let index = 0; index < collectors.length; index++) {
    const result = resourceProjection(path.join(root, `run-${index}`), { manifest: {
      run_id: `synthetic-${index}`, source_digest: digest, graph_digest: digest,
      instrumentation: { mode: "basic", policy_digest: policyDigest },
    }, registrations: new Map() });
    assert.equal(result.availability, "available");
    assert.ok(result.samples > 0);
    for (const scope of result.scopes) if (scope.scope === "process") assert.equal(scope.cpu_percent_one_core, null, "no CPU delta crosses the measurement pause");
  }
});

test("retained validation reconstructs path and target fields and rejects tampered summaries", async (t) => {
  const dir = temporary(t), graph = { units: [unit("a")] };
  const wait = { wait_reason: "capacity", blocking_resources: [], blocking_unit_ids: [] };
  const events = [
    ["run_started", 0], ["queued", 0, { unit_kind: "runner", row_ids: [] }],
    ["eligible", 0], ["wait_started", 0, wait], ["wait_ended", 2, wait],
    ["admitted", 2], ["started", 2], ["completed", 9],
    ["cleanup_started", 9], ["cleanup_completed", 10], ["run_completed", 10],
  ].map(([event, monotonic_ms, extra = {}], index) => ({ schema_id: "cartulary.harness_unit_event.v3",
    seq: index + 1, monotonic_ms, event, unit_id: /^(run_|cleanup_)/u.test(event) ? "harness:run" : "a",
    status: ["completed", "run_completed", "cleanup_completed"].includes(event) ? "passed" : "running",
    needs: [], resource_claims: /^(run_|cleanup_)/u.test(event) ? {} : { cpu: 1, process: 1 }, service_dependencies: [], ...extra }));
  const capability = { schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024,
    process_slots: 2, io_tokens: 1, postgres_lanes: 1, object_store_lanes: 1,
    writable_volume: true, port_lanes: 1, services: {}, sources: {}, snapshot_digest: digest };
  const manifest = { schema_id: "cartulary.harness_run_manifest.v2", run_id: "synthetic", target: "test-slice",
    command_id: "cartulary.harness.command.test_slice.v2", declared_inputs: {}, source_commit: "a".repeat(40), source_state: "dirty",
    source_digest: digest, toolchain_digest: digest, system_digest: digest, graph_digest: digest, capability_snapshot: capability,
    cache_mode: "off", instrumentation: { mode: "basic", policy_digest: digest }, started_at: "2026-10-09T00:00:00.000Z" };
  const timing = { setup_ms: 2, fixture_ms: 0, execution_ms: 7, collation_ms: 0, wrapper_ms: 1, unattributed_ms: 0, resource_blocking_ms: 2, process_count: 1 };
  const summary = { schema_id: "cartulary.harness_run_summary.v1", run_id: "synthetic", target: "test-slice", status: "pass",
    failure_class: null, failure_reason: null, unit_counts: { total: 1, passed: 1, failed: 0, skipped: 0, cancelled: 0 },
    wall_duration_ms: 10, critical_path: ["a"], actual_dependency_critical_path_ms: 9, timing_accounting: timing,
    resource_pressure: {
      requested_capacity: { browser_stack: 0, cpu: 1, io: 0, memory_mb: 0, object_store: 0, port_lane: 0, postgres: 0, process: 1, service_stack: 0, volume: 0 },
      resolved_capacity: { browser_stack: 1, cpu: 2, io: 1, memory_mb: 1, object_store: 1, port_lane: 1, postgres: 1, process: 2, service_stack: 1, volume: 1 },
      peak_use: { cpu: 1, process: 1 }, saturation_ms: {}, wait_events: 1, blocked_units: ["a"], resource_holders: [],
    }, cache: {}, artifact_refs: ["target-summaries/test-slice.json"] };
  const target = { schema_id: "cartulary.harness_target_summary.v1", target: "test-slice", command_id: manifest.command_id,
    status: "pass", failure_class: null, failure_reason: null, workload_digest: digest, unit_ids: ["a"],
    inclusive_wall_ms: 7, exclusive_wall_ms: 7, actual_dependency_critical_path_ms: 9,
    timing_accounting: { ...timing, setup_ms: 0, wrapper_ms: 0 }, children: [], evidence_refs: [] };
  atomicLocalFile(path.join(dir, "run-manifest.json"), JSON.stringify(manifest));
  atomicLocalFile(path.join(dir, "unit-events.ndjson"), events.map((event) => JSON.stringify(event)).join("\n") + "\n");
  atomicLocalFile(path.join(dir, "run-summary.json"), JSON.stringify(summary));
  atomicLocalFile(path.join(dir, "target-summaries/test-slice.json"), JSON.stringify(target));
  const records = [{ kind: "activity", ref: "activity:example", activity: "report_parse", unit_id: "a", invocation_id: null,
    allocation_ref: null, clock: "node_monotonic", clock_identity: "clock:example", resolution_ms: 0.000001, duration_ms: 5,
    availability: "available", outcome: "failed" }];
  atomicLocalFile(path.join(dir, "diagnostics/execution-index.json"), JSON.stringify({ schema_id: "cartulary.harness_execution_index.v1",
    run_id: manifest.run_id, source_digest: digest, graph_digest: digest, policy_digest: digest, status: "complete", omitted_records: 0,
    omitted_refs: [], records, records_digest: executionRecordsDigest(records) }));
  await validateCanonicalRun(dir);
  const explained = await performanceExplanation(dir);
  assert.equal(explained.dependency_path_ms, 9);
  assert.equal(explained.invocation_path_queue_ms, 2);
  assert.deepEqual(JSON.parse(formatPerformanceExplanation(explained, { json: true })), explained);
  const human = formatPerformanceExplanation(explained);
  assert.ok(human.includes("[UNIT] a queue_ms=2 execution_ms=7 invocation_path=true"));
  assert.ok(human.includes("activity=report_parse outcome=failed duration_ms=5"));
  assert.ok(Buffer.byteLength(human) < 8192);
  const retained = await loadRetainedObservability(dir);
  assert.equal(Object.hasOwn(retained, "built"), false);
  const payloads = [];
  assert.deepEqual(await exportRetainedObservability({ retained, endpoint: new URL("https://example.com"), headers: {} }, {
    fetchImpl: async (url, options) => { payloads.push([String(url), JSON.parse(options.body)]); return { ok: true, status: 200, redirected: false }; },
  }), { invocations: 1, signals: 2 });
  assert.ok(payloads[0][0].endsWith("/v1/traces")); assert.ok(payloads[1][0].endsWith("/v1/metrics"));
  assert.equal(JSON.stringify(payloads).includes("resource-samples"), false);
  assert.equal(JSON.stringify(payloads).includes("activity:example"), false);
  const cli = new URL("../observability-check-cli.mjs", import.meta.url);
  const check = (args) => spawnSync(process.execPath, [cli.pathname, ...args], { encoding: "utf8" });
  assert.equal(check(["--results-dir", dir]).status, 0);
  assert.equal(check(["--results-dir", path.join(dir, "missing")]).status, 2);
  const isolated = new Worker(`
    const { registerHooks } = require('node:module');
    const { parentPort, workerData } = require('node:worker_threads');
    registerHooks({ resolve(specifier, context, next) {
      if (/otlp-projection|otel-export/.test(specifier)) throw new Error('ordinary reader imported export');
      return next(specifier, context);
    } });
    import(workerData.facade).then(async (facade) => {
      await facade.loadRetainedObservability(workerData.dir);
      const value = await facade.performanceExplanation(workerData.dir);
      parentPort.postMessage(value.dependency_path_ms);
    });
  `, { eval: true, workerData: { facade: new URL("../observability.mjs", import.meta.url).href, dir } });
  t.after(() => isolated.terminate());
  assert.equal(await new Promise((resolve, reject) => { isolated.once("message", resolve); isolated.once("error", reject); }), 9);

  summary.actual_dependency_critical_path_ms = 7;
  atomicLocalFile(path.join(dir, "run-summary.json"), JSON.stringify(summary), { replace: true });
  await assert.rejects(validateCanonicalRun(dir), /critical duration/u);
  assert.equal(check(["--results-dir", dir]).status, 11);
});


test("boot clock is shared by helpers, quantized, and unavailable without kernel evidence", async () => {
  const { readBootClock } = await import("../../runtime/boot-clock.mjs");
  const read = (name) => name.endsWith("boot_id") ? "boot\n" : "123.45 777.12\n";
  const first = readBootClock({ read, link: () => "time:[123]" });
  assert.equal(first.monotonic_ms, 123450);
  assert.equal(first.clock, "linux_boottime_10ms");
  assert.deepEqual(readBootClock({ read, link: () => "time:[123]" }), first);
  assert.notEqual(readBootClock({ read, link: () => "time:[124]" }).clock_identity, first.clock_identity);
  assert.deepEqual(readBootClock({ read: () => { throw new Error("denied"); } }), {
    clock: "unavailable", clock_identity: null, monotonic_ms: null,
  });
});


test("run comparison rejects workload and outcome differences while allowing only the declared treatment", async () => {
  const { compareRunMetadata } = await import("../run-comparison.mjs");
  const run = { manifest: { command_id: "command", target: "slice", source_commit: "commit", source_state: "dirty", source_digest: digest,
    graph_digest: digest, toolchain_digest: digest, system_digest: digest, cache_mode: "off", capability_snapshot: {},
    declared_inputs: { ROWS: "a", HARNESS_DIAGNOSTICS: "off" }, instrumentation: { mode: "off", policy_digest: digest } },
    summary: { status: "pass", wall_duration_ms: 100, cache: { hit: 0 } }, terminal: new Map([["a", { status: "passed" }]]) };
  const treatment = structuredClone(run);
  treatment.manifest.instrumentation.mode = "basic"; treatment.manifest.declared_inputs.HARNESS_DIAGNOSTICS = "basic";
  treatment.summary.wall_duration_ms = 103;
  assert.equal(compareRunMetadata(treatment, run, "instrumentation").wall_delta_ms, 3);
  assert.deepEqual(compareRunMetadata(treatment, run).reasons, ["instrumentation_mode"]);
  treatment.manifest.declared_inputs.ROWS = "b";
  assert.equal(compareRunMetadata(treatment, run, "instrumentation").wall_delta_ms, null);
  treatment.summary.status = "fail";
  assert.ok(compareRunMetadata(treatment, run, "instrumentation").reasons.includes("incomplete_or_failed_run"));
  assert.equal(compareRunMetadata(run, run).qualification, "descriptive_only");
});

test("controlled resource streams distinguish CPU, waiting, allocation and I/O without equating them", (t) => {
  for (const scenario of [
    { cpu: 2_000_000, rss: 100, io: 0, percent: 100 },
    { cpu: 0, rss: 100, io: 0, percent: 0 },
    { cpu: 200_000, rss: 1_000_000, io: 0, percent: 10 },
    { cpu: 20_000, rss: 100, io: 8_388_608, percent: 1 },
  ]) {
    const f = retainedSamples(t, [
      { elapsed_ms: 0, metrics: { cpu_user_us: metric(0), cpu_system_us: metric(0), rss_bytes: metric(100), io_write_bytes: metric(0) } },
      { elapsed_ms: 2000, metrics: { cpu_user_us: metric(scenario.cpu), cpu_system_us: metric(0), rss_bytes: metric(scenario.rss), io_write_bytes: metric(scenario.io) } },
    ]);
    const scope = resourceProjection(f.dir, f.run).scopes[0];
    assert.equal(scope.cpu_percent_one_core, scenario.percent);
    assert.equal(scope.metrics.rss_bytes.maximum, scenario.rss);
    assert.equal(scope.metrics.io_write_bytes.observed_delta, scenario.io);
  }
});

test("equal CPU counter durations over different windows do not imply a joint utilization", (t) => {
  const f = retainedSamples(t, [
    { elapsed_ms: 0, metrics: { cpu_user_us: metric(0) } },
    { elapsed_ms: 1000, metrics: { cpu_user_us: metric(1_000_000), cpu_system_us: metric(0) } },
    { elapsed_ms: 2000, metrics: { cpu_system_us: metric(1_000_000) } },
  ]);
  assert.equal(resourceProjection(f.dir, f.run).scopes[0].cpu_percent_one_core, null);
});

test("lease correlation preserves shared allocation identity and cannot change fixture cleanup", async () => {
  const { FixtureBroker } = await import("../../scheduler/fixture-broker/index.mjs");
  const correlations = [];
  let released = 0;
  const broker = new FixtureBroker({
    providers: { browser_stack: { acquire: async () => ({ ownership: "owned", resource_ids: [], release: async () => { released += 1; } }) } },
    observeLease: (record) => { correlations.push(record); throw new Error("collector unavailable"); },
  });
  const a = await broker.acquire("browser_stack", { affinityKey: "shared", unitID: "a" });
  const b = await broker.acquire("browser_stack", { affinityKey: "shared", unitID: "b" });
  assert.equal(correlations[0].allocation_ref, correlations[1].allocation_ref);
  assert.notEqual(correlations[0].lease_ref, correlations[1].lease_ref);
  await a.release(); assert.equal(released, 0);
  await b.release(); assert.equal(released, 1);
  await broker.close(); assert.equal(released, 1);
});

test("instrumentation schemas validate without loading the runtime AJV compiler", async (t) => {
  const worker = new Worker(`
    const Module = require('node:module');
    const { parentPort, workerData } = require('node:worker_threads');
    const original = Module._load;
    Module._load = function(name, ...args) {
      if (String(name).includes('ajv')) throw new Error('runtime compiler was loaded');
      return original.call(this, name, ...args);
    };
    import(workerData.contract).then(({ validateSchemaSync }) => {
      const policy = JSON.parse(require('node:fs').readFileSync(new URL(workerData.policy), 'utf8'));
      validateSchemaSync(policy.schema_id, policy);
      validateSchemaSync('cartulary.harness_resource_sample.v1', {
        schema_id: 'cartulary.harness_resource_sample.v1', seq: 1, elapsed_ms: 0, segment: 0,
        scope: 'guest_context', scope_ref: 'guest:cpu', identity_digest: null,
        availability: 'unsupported', metrics: {},
      });
      for (const id of ['constructor', 'toString']) {
        let rejected = false;
        try { validateSchemaSync(id, {}); } catch { rejected = true; }
        if (!rejected) throw new Error('prototype member accepted as a schema');
      }
      parentPort.postMessage('validated');
    });
  `, { eval: true, workerData: {
    contract: new URL("../../contract/index.mjs", import.meta.url).href,
    policy: new URL("../../../harness_instrumentation_policy.json", import.meta.url).href,
  } });
  t.after(() => worker.terminate());
  assert.equal(await new Promise((resolve, reject) => { worker.once("message", resolve); worker.once("error", reject); }), "validated");
});

test("secure resource loading rejects malformed, oversized, mismatched and digest-tampered artifacts", (t) => {
  for (const mutation of ["malformed", "oversized", "identity", "digest", "forbidden_field"]) {
    const f = retainedSamples(t, [{ elapsed_ms: 0, metrics: { rss_bytes: metric(1) } }]);
    const file = path.join(f.dir, "diagnostics/resource-samples.ndjson");
    if (mutation === "malformed") writeFileSync(file, "{\n");
    if (mutation === "oversized") writeFileSync(file, " ".repeat(65537));
    if (mutation === "identity") f.run.manifest.run_id = "foreign";
    if (mutation === "digest") writeFileSync(file, readFileSync(file, "utf8").replace('"value":1', '"value":2'));
    if (mutation === "forbidden_field") writeFileSync(file, readFileSync(file, "utf8").replace('"metrics":{', '"secret":"forbidden","metrics":{'));
    assert.throws(() => resourceProjection(f.dir, f.run), undefined, mutation);
  }
});
