import assert from "node:assert/strict";
import test from "node:test";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import { mkdtempSync, writeFileSync, readFileSync, rmSync, symlinkSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import { Worker } from "node:worker_threads";
import { actualCriticalPath, intervalUnion } from "../../evidence-accounting/index.mjs";
import { reduceCanonicalUnitIntervals } from "../../evidence-accounting/index.mjs";
import { buildWorkGraph } from "../../scheduler/work-graph/model.mjs";
import { resourceProjection } from "../resource-projection.mjs";
import { readResourceObservations } from "../resource-reader.mjs";
import { atomicLocalFile, readLocalChunks, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { otlpProjection } from "../otlp-projection.mjs";
import { validateCanonicalRun } from "../canonical-evidence.mjs";
import { performanceExplanation, loadRetainedObservability, resolveExactRunDir } from "../observability.mjs";
import { formatPerformanceExplanation } from "../performance-presentation.mjs";
import { formatResourceScopes } from "../resource-presentation.mjs";
import { exportRetainedObservability, headersFromFile } from "../otel-export-cli.mjs";
import { hostClaims, processIdentity, transactAdmission } from "../../runtime/host-admission.mjs";
import { executionRecordsDigest } from "../../runtime/execution-observations.mjs";
import { canonicalJSONString, parseStrictJSON } from "../../contract/index.mjs";
import { createResourceAccumulator } from "../resource-accumulator.mjs";

test("strict JSON retains prototype-named members for closed schema rejection", () => {
  for (const text of [
    '{"__proto__":null,"constructor":1,"prototype":2}',
    '{"__proto__":{"inherited":true},"nested":{"__proto__":{"value":3}}}',
    '[{"__proto__":4},{"__proto__":[]}]',
  ]) {
    const value = parseStrictJSON(text);
    assert.deepEqual(value, JSON.parse(text));
    assert.equal(JSON.stringify(value), text);
    for (const object of Array.isArray(value) ? value : [value, ...(value.nested ? [value.nested] : [])]) {
      assert.equal(Object.getPrototypeOf(object), Object.prototype);
      assert.equal(Object.hasOwn(object, "__proto__"), true);
      assert.equal(Object.hasOwn(object, "inherited"), false);
    }
  }
  assert.throws(() => parseStrictJSON('{"__proto__":1,"\\u005f_proto__":2}'), /duplicate object member/u);
});

test("strict JSON preserves native values while rejecting non-interoperable and ambiguous input", () => {
  for (const text of [
    'null', 'true', 'false', '[]', '{}', ' [ 0, -1, 1.25, 2e-3, 9007199254740991 ] \r\n',
    '{"a":[],"b":{},"c":{"quote":"\\\"","slash":"\\\\","pair":"\\ud83d\\ude00","literal":"😀"}}',
    '{"0":0,"01":1,"toString":null,"hasOwnProperty":false}',
  ]) assert.deepEqual(parseStrictJSON(text), JSON.parse(text));
  for (const text of [
    '', ' ', 'undefined', 'NaN', 'Infinity', '1e400', '9007199254740992', '-0', '-0.0', '-1e-400',
    '01', '+1', '1.', '.1', '1e', '[1,]', '{"x":1,}', '{"x" 1}', '{"x":1} null',
    '{"x":1,"x":2}', '{"a":{"x":1,"\\u0078":2}}',
    '"\\ud800"', '"\\udc00"', '"\\ud800x"', '"\ud800"', '"\u0000"', '\u00a0{}',
  ]) assert.throws(() => parseStrictJSON(text), undefined, JSON.stringify(text));
});

const digest = `sha256:${"a".repeat(64)}`;
test("machine evidence rejects documentation paths before any filesystem access", async () => {
  const original = new Map(["openSync", "existsSync", "readFileSync"].map((name) => [name, fs[name]]));
  try {
    for (const name of original.keys()) fs[name] = () => { throw new Error("unexpected filesystem access"); };
    syncBuiltinESMExports();
    // Hostile tokens only; no documentation fixture is created or inspected.
    const forbiddenDirectory = "docs";
    for (const file of [path.join("/unavailable", forbiddenDirectory, "evidence.json"), "/unavailable/README", "/unavailable/evidence.md", "/unavailable/evidence.markdown"]) {
      assert.throws(() => readLocalFile(file), /documentation/u);
      assert.throws(() => resolveExactRunDir(file), /documentation/u);
      assert.throws(() => headersFromFile(file), /documentation/u);
      assert.throws(() => [...readLocalChunks(file, { maximum: 8 })], /documentation/u);
      await assert.rejects(validateCanonicalRun(file), /documentation/u);
    }
  } finally {
    for (const [name, originalMethod] of original) fs[name] = originalMethod;
    syncBuiltinESMExports();
  }
});
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

function retainedSamples(t, records) {
  const dir = temporary(t), bytes = records.map((record, index) => JSON.stringify({
    schema_id: "cartulary.harness_resource_sample.v2", seq: index + 1, segment: 0,
    scope: "process", scope_ref: "process:1", identity_digest: digest, availability: "available", ...record,
  })).join("\n") + "\n";
  const identity = { run_id: "synthetic", source_digest: digest, graph_digest: digest, policy_digest: digest };
  const index = { schema_id: "cartulary.harness_resource_index.v3", ...identity,
    status: "complete", clock: "graph_process_monotonic", coverage: "observed_partial",
    samples: records.length, sample_bytes: Buffer.byteLength(bytes), sample_digest: `sha256:${createHash("sha256").update(bytes).digest("hex")}`,
    omitted_observations: 0, discovery_truncations: 0, failed_sweeps: 0, sweeps: records.length,
    containers: { availability: "not_observed", identities: [], omitted_observations: 0, deadline_overruns: 0 },
    processes: [{ process_ref: "process:1", identity_digest: digest }],
    observer: { qualification: "component_diagnostic_only", collector_thread_cpu_user_us: 0, collector_thread_cpu_system_us: 0, collector_js_heap_peak_bytes: 0, read_bytes: 0, write_bytes: 0, maximum_sweep_ms: 0, shutdown_ms: 0, outside_component_accounting: "startup_parent_helpers_readers_publication" } };
  atomicLocalFile(path.join(dir, "diagnostics/resource-samples.ndjson"), bytes);
  atomicLocalFile(path.join(dir, "diagnostics/resource-index.json"), JSON.stringify(index));
  return { dir, run: { manifest: { ...identity, instrumentation: { mode: "basic", policy_digest: digest } }, registrations: new Map([["a", {}]]) } };
}

test("published resource indexes cannot hide unknown keys through prototype assignment", (t) => {
  const fixture = retainedSamples(t, [{ elapsed_ms: 0, metrics: { rss_bytes: { value: 1024, availability: "available" } } }]);
  assert.equal(resourceProjection(fixture.dir, fixture.run).availability, "available");
  const file = path.join(fixture.dir, "diagnostics/resource-index.json");
  const original = readFileSync(file, "utf8");
  writeFileSync(file, `${original.slice(0, -1)},"__proto__":null}`);
  assert.throws(() => resourceProjection(fixture.dir, fixture.run), /validation failed/u);
});

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

test("multiple samples within each segment retain CPU while excluding the pause", (t) => {
  const f = retainedSamples(t, [
    { elapsed_ms: 0, segment: 1, metrics: { cpu_user_us: metric(100000), cpu_system_us: metric(50000) } },
    { elapsed_ms: 1000, segment: 1, metrics: { cpu_user_us: metric(300000), cpu_system_us: metric(150000) } },
    { elapsed_ms: 4000, segment: 3, metrics: { cpu_user_us: metric(900000), cpu_system_us: metric(550000) } },
    { elapsed_ms: 5000, segment: 3, metrics: { cpu_user_us: metric(1200000), cpu_system_us: metric(650000) } },
  ]);
  const scope = resourceProjection(f.dir, f.run).scopes[0];
  // 300 ms before and 400 ms after the pause, over two observed seconds.
  // The one second of CPU accumulated during the three-second gap is excluded.
  assert.equal(scope.cpu_percent_one_core, 35);
  assert.equal(scope.metrics.cpu_user_us.observed_delta, 500000);
  assert.equal(scope.metrics.cpu_system_us.observed_delta, 200000);
  assert.equal(scope.metrics.cpu_user_us.observed_interval_ms, 2000);
  assert.equal(scope.metrics.cpu_system_us.observed_interval_ms, 2000);
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

test("retained validation reconstructs path and target fields and rejects tampered summaries", async (t) => {
  const dir = temporary(t), graph = { units: [unit("a")] };
  const wait = { wait_reason: "rank", blocking_resources: [], blocking_unit_ids: [] };
  const events = [
    ["run_started", 0], ["queued", 0, { unit_kind: "runner", row_ids: [] }],
    ["eligible", 0], ["wait_started", 0, wait], ["wait_ended", 2, wait],
    ["admitted", 2], ["started", 2], ["completed", 9],
    ["cleanup_started", 9], ["cleanup_completed", 10], ["run_completed", 10],
  ].map(([event, monotonic_ms, extra = {}], index) => ({ schema_id: "cartulary.harness_unit_event.v4",
    seq: index + 1, monotonic_ms, event, unit_id: /^(run_|cleanup_)/u.test(event) ? "harness:run" : "a",
    status: ["completed", "run_completed", "cleanup_completed"].includes(event) ? "passed" : "running",
    needs: [], resource_claims: /^(run_|cleanup_)/u.test(event) ? {} : { cpu: 1, process: 1 }, service_dependencies: [], ...extra }));
  const capability = { schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024,
    process_slots: 2, io_tokens: 1, postgres_lanes: 1, object_store_lanes: 1,
    writable_volume: true, port_lanes: 1, services: {}, sources: {}, snapshot_digest: digest };
  const manifest = { schema_id: "cartulary.harness_run_manifest.v3", run_id: "synthetic", target: "test-slice",
    command_id: "cartulary.harness.command.test_slice.v2", declared_inputs: {}, source_commit: "a".repeat(40), source_state: "dirty",
    source_digest: digest, toolchain_digest: digest, system_digest: digest, graph_digest: digest, capability_snapshot: capability, go_worker_budget: 2,
    cache_mode: "off", instrumentation: { mode: "basic", policy_digest: digest }, started_at: "2026-10-09T00:00:00.000Z" };
  const timing = { setup_ms: 2, fixture_ms: 0, execution_ms: 7, collation_ms: 0, wrapper_ms: 1, unattributed_ms: 0, resource_blocking_ms: 0, process_count: 1 };
  const summary = { schema_id: "cartulary.harness_run_summary.v1", run_id: "synthetic", target: "test-slice", status: "pass",
    failure_class: null, failure_reason: null, unit_counts: { total: 1, passed: 1, failed: 0, skipped: 0, cancelled: 0 },
    wall_duration_ms: 10, critical_path: ["a"], actual_dependency_critical_path_ms: 9, timing_accounting: timing,
    resource_pressure: {
      requested_capacity: { browser_stack: 0, cpu: 1, io: 0, memory_mb: 0, object_store: 0, port_lane: 0, postgres: 0, process: 1, service_stack: 0, volume: 0 },
      resolved_capacity: { browser_stack: 1, cpu: 2, io: 1, memory_mb: 1, object_store: 1, port_lane: 1, postgres: 1, process: 2, service_stack: 1, volume: 1 },
      peak_use: { cpu: 1, process: 1 }, saturation_ms: {}, wait_events: 1, blocked_units: [], resource_holders: [],
    }, cache: {}, artifact_refs: ["target-summaries/test-slice.json"] };
  const target = { schema_id: "cartulary.harness_target_summary.v1", target: "test-slice", command_id: manifest.command_id,
    status: "pass", failure_class: null, failure_reason: null, workload_digest: digest, unit_ids: ["a"],
    inclusive_wall_ms: 7, exclusive_wall_ms: 7, actual_dependency_critical_path_ms: 9,
    timing_accounting: { ...timing, setup_ms: 0, wrapper_ms: 0 }, children: [], evidence_refs: [] };
  atomicLocalFile(path.join(dir, "run-manifest.json"), JSON.stringify(manifest));
  atomicLocalFile(path.join(dir, "unit-events.ndjson"), events.map((event) => JSON.stringify(event)).join("\n") + "\n");
  atomicLocalFile(path.join(dir, "run-summary.json"), JSON.stringify(summary));
  atomicLocalFile(path.join(dir, "target-summaries/test-slice.json"), JSON.stringify(target));
  const records = [{ kind: "activity", local_start_ms: null, local_end_ms: null, anchors: [], ref: "activity:example", activity: "report_parse", unit_id: "a", invocation_id: null,
    allocation_ref: null, clock: "node_monotonic", clock_identity: "clock:example", resolution_ms: 0.000001, duration_ms: 5,
    availability: "available", outcome: "failed" }];
  atomicLocalFile(path.join(dir, "diagnostics/execution-index.json"), JSON.stringify({ schema_id: "cartulary.harness_execution_index.v2", graph_clock: null,
    run_id: manifest.run_id, source_digest: digest, graph_digest: digest, policy_digest: digest, status: "complete", omitted_records: 0,
    omitted_refs: [], records, records_digest: executionRecordsDigest(records) }));
  await validateCanonicalRun(dir);
  const forbiddenDirectory = "docs";
  for (const reference of ["../escaped.json", path.join(forbiddenDirectory, "evidence.json"), "README", "notes.md"]) {
    const poisoned = { ...summary, artifact_refs: [...summary.artifact_refs, reference] };
    atomicLocalFile(path.join(dir, "run-summary.json"), JSON.stringify(poisoned), { replace: true });
    await assert.rejects(validateCanonicalRun(dir), /normalized and relative|documentation/u);
  }
  atomicLocalFile(path.join(dir, "run-summary.json"), JSON.stringify(summary), { replace: true });
  const explained = await performanceExplanation(dir);
  assert.equal(explained.dependency_path_ms, 9);
  assert.equal(explained.invocation_path_queue_ms, 2);
  assert.deepEqual(JSON.parse(formatPerformanceExplanation(explained, { json: true })), explained);
  const human = formatPerformanceExplanation(explained);
  assert.ok(human.includes("[UNIT] a queue_ms=2 execution_ms=7 queue_by_reason_ms="));
  assert.deepEqual(explained.units[0].queue_by_reason_ms, { resources: 0, exclusive_barrier: 0, rank: 2, scheduler_stop: 0 });
  assert.ok(human.includes("activity=report_parse outcome=failed duration_ms=5"));
  assert.ok(Buffer.byteLength(human) < 8192);
  const many = structuredClone(explained);
  const aggregateMetric = { maximum: 4096, observed_delta: null, observed_interval_ms: 0, unavailable: 0 };
  many.resources = { availability: "available", mode: "basic", scopes_omitted: 3, unit_cpu_omitted: 2,
    scopes: Array.from({ length: 20 }, (_, i) => ({ scope: "container", scope_ref: `container:${i + 1}`, unit_id: null,
      allocation_ref: null, samples: 1, unavailable: 0, resets: 0, cpu_percent_one_core: null,
      metrics: Object.fromEntries(["container_memory_usage_bytes", "container_memory_limit_bytes", "container_pids", "container_cpu_us", "container_block_read_bytes", "container_block_write_bytes"].map((name) => [name, aggregateMetric])) })) };
  assert.deepEqual(JSON.parse(formatPerformanceExplanation(many, { json: true })), many);
  const bounded = formatPerformanceExplanation(many);
  assert.ok(Buffer.byteLength(bounded) < 8192);
  assert.match(bounded, /\[OMITTED\] lines=\d+/u);
  assert.match(bounded, /\[OMITTED\] scopes=3/u);
  const retained = await loadRetainedObservability(dir);
  assert.equal(Object.hasOwn(retained, "built"), false);
  assert.ok(Object.isFrozen(retained.resources));
  assert.throws(() => { retained.run.manifest.run_id = "forged"; }, TypeError);
  assert.equal(retained.run.terminal.set, undefined);
  assert.throws(() => { retained.run.terminal.get("a").status = "forged"; }, TypeError);
  const payloads = [];
  assert.deepEqual(await exportRetainedObservability({ retained, endpoint: new URL("https://example.com"), headers: {} }, {
    fetchImpl: async (url, options) => { payloads.push([String(url), JSON.parse(options.body)]); return { ok: true, status: 200, redirected: false }; },
  }), { invocations: 1, signals: 2 });
  assert.ok(payloads[0][0].endsWith("/v1/traces")); assert.ok(payloads[1][0].endsWith("/v1/metrics"));
  assert.equal(JSON.stringify(payloads).includes("resource-samples"), false);
  assert.equal(JSON.stringify(payloads).includes("activity:example"), false);
  const cli = new URL("../observability-check-cli.mjs", import.meta.url);
  const check = (args) => spawnSync(process.execPath, [cli.pathname, ...args], { encoding: "utf8" });
  assert.equal(check(["--results-dir", dir]).status, 11, "partial diagnostic fixture cannot establish full validity");
  assert.equal(check(["--results-dir", dir, "--check", "timing"]).status, 0);
  assert.equal(check(["--results-dir", path.join(dir, "missing")]).status, 2);
  const sampled = retainedSamples(t, [{ elapsed_ms: 0, metrics: { rss_bytes: metric(1024) } }]);
  for (const name of ["resource-index.json", "resource-samples.ndjson"]) {
    atomicLocalFile(path.join(dir, "diagnostics", name), readFileSync(path.join(sampled.dir, "diagnostics", name)));
  }
  const isolated = new Worker(`
    const { registerHooks, syncBuiltinESMExports } = require('node:module');
    const fs = require('node:fs');
    let traversals = 0;
    const open = fs.openSync;
    fs.openSync = (...args) => {
      if (String(args[0]).endsWith('/resource-samples.ndjson')) traversals++;
      return open(...args);
    };
    syncBuiltinESMExports();
    const { parentPort, workerData } = require('node:worker_threads');
    registerHooks({ resolve(specifier, context, next) {
      if (/otlp-projection|otel-export/.test(specifier)) throw new Error('ordinary reader imported export');
      return next(specifier, context);
    } });
    import(workerData.facade).then(async (facade) => {
      await facade.loadRetainedObservability(workerData.dir);
      const value = await facade.performanceExplanation(workerData.dir);
      await facade.performanceExplanation(workerData.dir, "", { comparisonDir: workerData.dir });
      process.argv = [process.execPath, workerData.cli, '--results-dir', workerData.dir, '--check', 'diagnostics'];
      await import(workerData.cli);
      parentPort.postMessage({ duration: value.dependency_path_ms, traversals, artifactExit: process.exitCode });
    });
  `, { eval: true, workerData: { facade: new URL("../observability.mjs", import.meta.url).href, cli: cli.href, dir } });
  t.after(() => isolated.terminate());
  assert.deepEqual(await new Promise((resolve, reject) => { isolated.once("message", resolve); isolated.once("error", reject); }), { duration: 9, traversals: 4, artifactExit: 11 });

  const envelopeFile = path.join(dir, "diagnostics/invocation-envelope.json");
  const envelope = { schema_id: "cartulary.harness_invocation_envelope.v2",
    run_id: "synthetic", source_digest: digest, graph_digest: digest,
    invocation_digest: null, boundary: "graph_main_entry_to_pre_scan", clock: "graph_process_monotonic", canonical_ref: "unit-events.ndjson",
    intervals: [{ phase: "configuration_catalog", start_ms: 0, end_ms: 2 }],
    exclusions: ["make_and_preflight", "module_imports", "retained_secret_scan", "command_return", "receipt_publication"] };
  atomicLocalFile(envelopeFile, JSON.stringify(envelope));
  assert.equal((await performanceExplanation(dir)).envelope.intervals[0].end_ms, 2);
  atomicLocalFile(envelopeFile, JSON.stringify({ ...envelope, run_id: "foreign" }), { replace: true });
  await assert.rejects(performanceExplanation(dir), /envelope identity/u);
  await assert.rejects(performanceExplanation(dir, "", { comparisonDir: dir }), /envelope identity/u);
  await t.test("published foreign envelope is also rejected by checker", async () => {
    await assert.rejects(loadRetainedObservability(dir), /envelope identity/u);
    assert.equal(check(["--results-dir", dir]).status, 11);
  });
  const duplicateEnvelope = JSON.stringify(envelope).replace('{"schema_id":', '{"run_id":"foreign","schema_id":');
  atomicLocalFile(envelopeFile, duplicateEnvelope, { replace: true });
  await t.test("duplicate envelope keys fail both retained readers", async () => {
    await assert.rejects(performanceExplanation(dir));
    await assert.rejects(loadRetainedObservability(dir));
  });
  rmSync(envelopeFile);

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
      validateSchemaSync('cartulary.harness_resource_sample.v2', {
        schema_id: 'cartulary.harness_resource_sample.v2', seq: 1, elapsed_ms: 0, segment: 0,
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


test("container counter resets stay contextual and do not contaminate the physical roster", (t) => {
  const f = retainedSamples(t, [0, 1000, 25, 125].map((cpu, i) => ({ elapsed_ms: i * 1000, scope: "container", scope_ref: "container:1",
    metrics: { container_cpu_us: metric(cpu), container_memory_usage_bytes: metric(4096) } })));
  const file = path.join(f.dir, "diagnostics/resource-index.json");
  const index = JSON.parse(readFileSync(file)); index.processes = [];
  index.containers = { availability: "available", identities: [{ container_ref: "container:1", identity_digest: digest, ownership: "borrowed", allocation_ref: null }], omitted_observations: 0, deadline_overruns: 0 };
  atomicLocalFile(file, JSON.stringify(index), { replace: true });
  const view = resourceProjection(f.dir, f.run);
  assert.equal(view.scopes[0].resets, 1); assert.equal(view.scopes[0].metrics.container_cpu_us.observed_delta, 1100);
  assert.equal(view.scopes[0].cpu_percent_one_core, 0.055);
  assert.equal(view.sampled_simultaneous_rss_sum_bytes, null); assert.deepEqual(view.unit_cpu_lower_bounds, []);
  assert.deepEqual(view.containers, index.containers);
});

test("resource reports expose sampled gauges and only valid counter intervals across every scope", () => {
  const accumulate = createResourceAccumulator();
  const record = (elapsed_ms, scope, scope_ref, metrics, extra = {}) => accumulate.consume({ elapsed_ms, scope, scope_ref,
    identity_digest: digest, segment: 0, availability: "available", metrics, ...extra }, new Map());
  const sample = (cpu, read, memory) => ({ container_cpu_us: metric(cpu), container_block_read_bytes: metric(read),
    container_block_write_bytes: metric(0), container_memory_usage_bytes: metric(memory),
    container_memory_limit_bytes: metric(8192), container_pids: metric(2) });
  record(0, "container", "container:shared", sample(1000, 10, 4096));
  const roster = { identities: [{ container_ref: "container:shared", ownership: "borrowed", allocation_ref: "allocation:shared" }] };
  let lines = formatResourceScopes({ scopes: accumulate.finish().rows, containers: roster });
  assert.match(lines[0], /cpu_percent_one_core=not_observed/u);
  assert.match(lines[0], /observed_delta_container_cpu_us=not_observed/u);
  assert.match(lines[0], /sampled_max_container_memory_usage_bytes=4096/u);
  assert.match(lines[0], /sampled_max_container_memory_limit_bytes=8192/u);
  assert.match(lines[0], /sampled_max_container_pids=2\(count\)/u);
  assert.match(lines[0], /allocation=allocation:shared.*ownership=borrowed attribution=contextual/u);
  assert.doesNotMatch(lines[0], /rss/u);
  record(1000, "container", "container:shared", sample(101000, 110, 2048));
  record(2000, "container", "container:shared", sample(10, 5, 1024));
  record(3000, "container", "container:shared", {}, { availability: "container_gone" });
  record(4000, "container", "container:shared", sample(1000000, 900, 4096));
  record(5000, "container", "container:shared", sample(1000000, 900, 4096), { segment: 1 });
  record(6000, "container", "container:shared", sample(1000000, 900, 4096), { segment: 1 });
  record(6000, "process", "process:1", { rss_bytes: metric(0), io_read_bytes: { value: null, availability: "permission_denied" } });
  record(6000, "cgroup_context", "cgroup:1", { cgroup_memory_bytes: metric(512), cgroup_pids: metric(3) });
  record(6000, "guest_context", "guest:1", { psi_some_us: metric(500) }, { identity_digest: null });
  lines = formatResourceScopes({ scopes: accumulate.finish().rows, containers: roster, scopes_omitted: 3, unit_cpu_omitted: 2 });
  const container = lines.find((line) => line.startsWith("[SCOPE] container:"));
  assert.match(container, /cpu_percent_one_core=5 /u);
  assert.match(container, /observed_delta_container_block_read_bytes=100 container_block_read_bytes_interval_ms=2000/u);
  assert.match(container, /observed_delta_container_block_write_bytes=0 /u);
  assert.match(container, /unavailable=1 resets=2 /u);
  const process = lines.find((line) => line.startsWith("[SCOPE] process:"));
  assert.match(process, /sampled_max_rss_bytes=0/u);
  assert.match(process, /observed_delta_io_read_bytes=not_observed/u);
  assert.match(lines.find((line) => line.includes("scope=cgroup_context")), /sampled_max_cgroup_memory_bytes=512/u);
  assert.match(lines.find((line) => line.includes("scope=guest_context")), /observed_delta_psi_some_us=not_observed/u);
  assert.deepEqual(lines.slice(0, 2), ["[OMITTED] scopes=3", "[OMITTED] unit_cpu_lower_bounds=2"]);
});


test("scoped observability checks cannot establish full validity and never repair evidence", async (t) => {
  const { checkRetainedObservability } = await import("../retained-checks.mjs");
  const { writeCompleteFixture } = await import("./retained-fixture.mjs");
  const dir = mkdtempSync(path.join(os.tmpdir(), "cartulary-scoped-evidence-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  const fixture = writeCompleteFixture(dir);
  const { fixtureExplanation } = await import("../fixture-explanation.mjs");
  assert.deepEqual((await fixtureExplanation(dir)).leases, []);
  await assert.rejects(fixtureExplanation(dir, "absent"), /absent/u);
  const before = new Map(fs.readdirSync(dir, { recursive: true, withFileTypes: true }).filter((e) => e.isFile()).map((e) => {
    const file = path.join(e.parentPath, e.name); return [file, readFileSync(file)];
  }));
  for (const check of ["events", "timing", "evidence", "diagnostics", "all"]) {
    const result = await checkRetainedObservability(dir, check);
    assert.equal(result.full_validity, check === "all");
    assert.equal(result.run_id, fixture.manifest.run_id);
  }
  for (const [file, bytes] of before) assert.deepEqual(readFileSync(file), bytes);
  fixture.write("run-manifest.json", { ...fixture.manifest, declared_inputs: { OWNER: "another.owner" } });
  assert.equal((await checkRetainedObservability(dir, "events")).full_validity, false);
  await assert.rejects(checkRetainedObservability(dir, "all"), /manifest binding/u);
  fixture.write("run-manifest.json", fixture.manifest);
  rmSync(path.join(dir, "unit-results/a.json"));
  assert.equal((await checkRetainedObservability(dir, "timing")).full_validity, false);
  await assert.rejects(checkRetainedObservability(dir, "evidence"));
  await assert.rejects(checkRetainedObservability(dir, "all"));
  await assert.rejects(checkRetainedObservability(dir, "unknown"), /invalid observability/u);
});

test("complete observability rejects duplicate, unsafe and incomplete evidence at its boundary", async (t) => {
  const { checkRetainedObservability } = await import("../retained-checks.mjs");
  const { writeCompleteFixture } = await import("./retained-fixture.mjs");
  for (const mutation of [
    (dir, f) => f.write("cleanup-results.json", { schema_id: "cartulary.harness_cleanup_results.v1", run_id: f.manifest.run_id, results: [] }),
    (dir, f) => f.write("execution-context.json", { ...f.context, graph: { ...f.context.graph, units: [...f.context.graph.units, ...f.context.graph.units] } }),
    (dir, f) => { rmSync(path.join(dir, "unit-results/a.json")); symlinkSync("/dev/null", path.join(dir, "unit-results/a.json")); },
    (dir, f) => f.write("unit-results/a.json", { schema_id: "cartulary.harness_unit_result.v1", unit_id: "a", semantic_digest: f.context.graph.units[0].semantic_digest,
      status: "failed", exit_code: 1, signal: null, failure_class: "harness", failure_reason: "tool_diagnostic_failure", evidence_outputs: [], missing_outputs: [] }),
  ]) {
    const dir = mkdtempSync(path.join(os.tmpdir(), "cartulary-invalid-evidence-"));
    t.after(() => rmSync(dir, { recursive: true, force: true }));
    const f = writeCompleteFixture(dir); mutation(dir, f);
    await assert.rejects(checkRetainedObservability(dir, "all"));
  }
  for (const budget of [undefined, 0, 1.5, 3, Number.MAX_SAFE_INTEGER + 1]) {
    const dir = temporary(t), f = writeCompleteFixture(dir);
    if (budget === undefined) delete f.manifest.go_worker_budget; else f.manifest.go_worker_budget = budget;
    // Unsafe integers reject during strict JSON decoding, before any digest can
    // be computed. Other cases must reach schema/semantic budget validation.
    if (budget !== Number.MAX_SAFE_INTEGER + 1) {
      f.context.manifest_digest = `sha256:${createHash("sha256").update(canonicalJSONString(f.manifest)).digest("hex")}`;
    }
    f.write("run-manifest.json", f.manifest); f.write("execution-context.json", f.context);
    await assert.rejects(checkRetainedObservability(dir, "all"), budget === Number.MAX_SAFE_INTEGER + 1
      ? /integer is outside the I-JSON interoperable range/u : /go_worker_budget|Go worker budget/u);
  }
});

test("fixture explanation preserves JSON filters, exact invocation and secure read-only journals", async (t) => {
  const { writeCompleteFixture } = await import("./retained-fixture.mjs");
  const dir = temporary(t); writeCompleteFixture(dir);
  const script = path.resolve(import.meta.dirname, "../../diagnostics/explain-run-cli.mjs");
  const invoke = (...args) => spawnSync(process.execPath, [script, "--results-dir", dir, "--detail", "fixtures", ...args], { encoding: "utf8" });
  const before = fs.readdirSync(dir, { recursive: true }).sort();
  const result = invoke("--json", "--fixture-threshold-ms", "17", "--fixture-top", "2");
  assert.equal(result.status, 0, result.stderr);
  const report = JSON.parse(result.stdout);
  assert.equal(report.schema_id, "cartulary.fixture_report.v1");
  assert.equal(report.run_id, path.basename(dir));
  assert.equal(report.threshold_ms, 17);
  assert.deepEqual(fs.readdirSync(dir, { recursive: true }).sort(), before);
  for (const args of [["--fixture-top", "0"], ["--fixture-threshold-ms", "2x"], ["--run-id", "another-run"], ["--target", "absent"]]) assert.notEqual(invoke(...args).status, 0);
  const parent = temporary(t);
  writeCompleteFixture(path.join(parent, "retained-run"));
  const ambiguous = spawnSync(process.execPath, [script, "--results-dir", parent, "--detail", "fixtures", "--json"], { encoding: "utf8" });
  assert.notEqual(ambiguous.status, 0, "a results parent cannot implicitly select its newest run");
  const journals = path.join(dir, "_shared", "test-services", "fixture", "journals");
  fs.mkdirSync(journals, { recursive: true, mode: 0o700 });
  const journal = path.join(journals, "example.ndjson");
  writeFileSync(journal, '{"seq":1,"seq":2}\n', { mode: 0o600 });
  assert.notEqual(invoke("--json").status, 0, "ambiguous journal records reject");
  rmSync(journal); symlinkSync(path.join(dir, "run-manifest.json"), journal);
  assert.notEqual(invoke("--json").status, 0, "journal links reject without following them");
});

test("complete retained evidence rejects altered context preimages independently of qualification", async (t) => {
  const { checkRetainedObservability } = await import("../retained-checks.mjs");
  const { writeCompleteFixture } = await import("./retained-fixture.mjs");
  for (const mutate of [
    (f) => { f.context.source_inputs[0].byte_digest = `sha256:${"b".repeat(64)}`; },
    (f) => { f.context.source_inputs.push(f.context.source_inputs[0]); },
    (f) => { f.context.instrumentation_policy_bytes += " "; },
    (f) => { f.context.unrecognized_context = true; },
    (f) => { f.context.projections["test-slice"].unit_ids = []; },
  ]) {
    const dir = temporary(t), f = writeCompleteFixture(dir);
    mutate(f); f.write("execution-context.json", f.context);
    await assert.rejects(checkRetainedObservability(dir, "all"));
  }
});
