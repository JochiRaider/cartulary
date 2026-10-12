import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { canonicalTimingAccounting, projectResourcePressure, reduceCanonicalUnitIntervals } from "../../evidence-accounting/index.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";
import { resolveExactRunDir } from "../observability.mjs";
import { deliver, headersFromFile, validatedEndpoint } from "../otel-export-cli.mjs";
import { renderInstrumentationDefinitions } from "../../generated-artifacts/instrumentation-definitions.mjs";
import { readFileSync } from "node:fs";
import { createResourceAccumulator } from "../resource-accumulator.mjs";
import Ajv2020 from "ajv/dist/2020.js";
import { createSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { createLaunchContext, readLaunchContext } from "../../runtime/launch-context.mjs";
import { openExecutionChannel } from "../../test-support/execution-channel-fixture.mjs";
import { readExecutionIndex } from "../execution-reader.mjs";
import { instrumentationPolicy } from "../instrumentation-policy.mjs";
import { activityTiming, beginActivity } from "../../runtime/local-activity.mjs";

test("qualification roster resolves exact public owner rows without changing the frozen treatments", () => {
  const root = path.resolve(import.meta.dirname, "../../../..");
  const plan = JSON.parse(readFileSync(path.join(root, "tools/harness_instrumentation_qualification_plan.json"), "utf8"));
  assert.deepEqual(plan.reader_counts, [0, 1, 4]);
  assert.deepEqual(plan.modes, ["off", "basic"]);
  assert.equal(plan.reader_cadence_ms, 30000);
  assert.equal(plan.memory_cadence_ms, 100);
  assert.equal(plan.seed, 20261009);
  assert.deepEqual(plan.steady_fraction, [0.1, 0.9]);
  assert.deepEqual(plan.budgets, { elapsed_median_percent: 2, elapsed_p90_percent: 3,
    cpu_mean_percent_one_core: 2.5, steady_rss_bytes: 33554432, finalization_ms: 500 });
  assert.deepEqual(plan.workloads.map((workload) => workload.id), ["go", "vitest", "shell", "service", "browser"]);
  const runners = [];
  for (const workload of plan.workloads) {
    const family = JSON.parse(readFileSync(path.join(root, `tools/test_families/${workload.owner}.json`), "utf8"));
    const selected = family.rows.filter((row) => workload.rows.includes(row.row_id));
    assert.equal(selected.length, 1);
    assert.equal(selected[0].status, "active");
    assert.equal(selected[0].owner_id, workload.owner);
    assert.equal(workload.target === "service-backed-test-slice", selected[0].service_dependencies.length > 0);
    runners.push(selected[0].runner);
  }
  assert.deepEqual(runners, ["go", "vitest", "shell", "go", "playwright"]);
});

test("local activities retain independent overlapping durations, unavailable clocks and incomplete ends", () => {
  const options = { repoRoot: ".", environment: {}, activity: "report_parse" };
  const tick = (value, identity = "clock:a") => ({ clock: "node_monotonic", clock_identity: identity, resolution_ms: 0.000001, monotonic_ms: value });
  let now = 10;
  const read = () => tick(now);
  const outer = beginActivity(options, read);
  now = 12; const inner = beginActivity(options, read);
  now = 15; assert.equal(inner.finish("failed").duration_ms, 3);
  now = 20; assert.equal(outer.finish("passed").duration_ms, 10);
  assert.equal(outer.finish("failed"), null);
  for (const end of [tick(5), tick(20, "clock:b"), null]) {
    assert.equal(activityTiming(tick(10), end).duration_ms, null);
  }
  const unavailable = beginActivity(options, () => { throw new Error("clock unavailable"); }).finish("cancelled");
  assert.equal(unavailable.availability, "unavailable");
  assert.equal(unavailable.duration_ms, null);
});

test("neutral launch records retain packed parentage, unlaunched failure and one terminal outcome", (t) => {
  const repoRoot = path.resolve(import.meta.dirname, "../../../..");
  const runRoot = temporary(t), runID = path.basename(runRoot);
  const runtime = createSuiteRuntime({ repoRoot, runRoot, runID });
  t.after(() => runtime.close());
  const digest = `sha256:${"a".repeat(64)}`;
  const { policy, digest: policyDigest } = instrumentationPolicy();
  const manifest = { run_id: runID, source_digest: digest, graph_digest: digest, instrumentation: { mode: "basic", policy_digest: policyDigest } };
  const session = openExecutionChannel(runtime);
  const environment = { CARTULARY_TEST_RESULTS_DIR: path.dirname(runRoot), CARTULARY_TEST_RUN_ID: runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID, ...session.environment };
  const parent = createLaunchContext({ repoRoot, environment, unitID: "u", rowIDs: ["a", "b"], producer: "unit" });
  parent.spawned();
  const child = createLaunchContext({ repoRoot, environment: { ...environment, ...parent.environment }, producer: "go" });
  const activity = beginActivity({ repoRoot, environment, launch: child.identity, activity: "migration" });
  activity.finish("failed");
  beginActivity({ repoRoot, environment, launch: parent.identity, activity: "fixture_reset" });
  child.settle("spawn_failed"); child.settle("passed"); parent.settle("passed");
  session.close();
  createLaunchContext({ repoRoot, environment, unitID: "u" }).settle("passed");
  const index = {records:session.records().map(v=>v.record??v)};
  assert.equal(index.records.length, 4, "closed producers cannot publish late records");
  assert.equal(index.records.find((record) => record.activity === "migration").outcome, "failed");
  assert.equal(index.records.find((record) => record.activity === "fixture_reset").outcome, "incomplete");
  const recorded = index.records.find((record) => record.kind === "launch" && record.invocation_id === child.identity.invocation_id);
  assert.deepEqual(recorded.row_ids, ["a", "b"]);
  assert.equal(recorded.parent_invocation_id, parent.identity.invocation_id);
  assert.equal(recorded.outcome, "spawn_failed");
  assert.equal(recorded.launched, false);
  const stale = { ...environment, ...parent.environment, CARTULARY_TEST_RUN_ID: "other" };
  assert.throws(() => readLaunchContext(stale), /run mismatch/u);
  const fresh = createLaunchContext({ repoRoot, environment: stale });
  assert.equal(fresh.identity.parent_invocation_id, null);
  assert.equal(fresh.identity.run_id, "other");
  assert.notEqual(fresh.identity.invocation_id, parent.identity.invocation_id);
});

test("late attribution counts only complete jointly observed CPU intervals after registration", () => {
  const accumulator = createResourceAccumulator();
  const roster = new Map([["process:1", { unit_id: "a", allocation_ref: null, attribution_start_ms: 500 }]]);
  for (const [elapsed_ms, user, system] of [[0, 100, 20], [1000, 1000, 30], [2000, 1200, 40]]) {
    accumulator.consume({ scope: "process", scope_ref: "process:1", identity_digest: "proof", elapsed_ms, segment: 0, availability: "available",
      metrics: { cpu_user_us: { value: user, availability: "available" }, cpu_system_us: { value: system, availability: "available" } } }, roster);
  }
  assert.deepEqual([...accumulator.finish().unitCPU], [["a", 210]]);
});

const temporary = (t) => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-foundation-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  return root;
};
const graph = { units: [
  { unit_id: "a", kind: "runner", needs: [], resource_claims: { cpu: 1 } },
  { unit_id: "b", kind: "finalizer", needs: [], resource_claims: { cpu: 1 } },
] };
const facts = {
  starts: new Map([["a", 2], ["b", 5]]),
  terminals: new Map([["a", { event: "completed", monotonic_ms: 9, seq: 11 }], ["b", { event: "failed", monotonic_ms: 10, seq: 12 }]]),
  waits: new Map([["a", { start: 0, end: 2, blocking_unit_ids: [] }], ["b", { start: 3, end: 5, blocking_unit_ids: ["a"] }]]),
  admissions: new Map([["a", { time: 2, seq: 4 }], ["b", { time: 5, seq: 9 }]]),
  fixtureAcquired: new Map([["a", 4]]), cleanupStarted: 10, finalMonotonicMs: 12,
};

test("independent overlapping accounting expectations include precedence and reservations", () => {
  assert.deepEqual(canonicalTimingAccounting(facts, graph, 12), {
    setup_ms: 2, fixture_ms: 2, execution_ms: 1, collation_ms: 5,
    wrapper_ms: 2, unattributed_ms: 0, resource_blocking_ms: 4, process_count: 2,
  });
  assert.deepEqual(canonicalTimingAccounting(facts, graph, 7, { selectedUnitIDs: ["a"], includeRunEnvelope: false }), {
    setup_ms: 0, fixture_ms: 2, execution_ms: 5, collation_ms: 0,
    wrapper_ms: 0, unattributed_ms: 0, resource_blocking_ms: 2, process_count: 1,
  });
  const pressure = projectResourcePressure(facts, graph, { cpu: 2, process: 2, io: 1, memory_mb: 1, postgres: 1, object_store: 1, port_lane: 1, browser_stack: 1, service_stack: 1, volume: 1 });
  assert.deepEqual(pressure.peak_use, { cpu: 2 });
  assert.deepEqual(pressure.saturation_ms, { cpu: 4 });
  assert.equal(pressure.wait_events, 2);
  assert.deepEqual(pressure.blocked_units, ["a", "b"]);
  assert.deepEqual(pressure.resource_holders, ["a"]);
});

test("canonical phases retain incomplete ends and reject the 129th start", async (t) => {
  const root = temporary(t), file = path.join(root, "events.ndjson");
  const entries = [{ event: "queued", unit_kind: "runner", row_ids: [], status: "pending" }];
  for (let i = 1; i <= 128; i++) entries.push({ event: "phase_started", interval_id: `phase:${i}`, phase: "runner" });
  const publish = () => writeFileSync(file, entries.map((entry, i) => JSON.stringify({
    schema_id: "cartulary.harness_unit_event.v3", seq: i + 1, monotonic_ms: i,
    unit_id: "a", status: "running", needs: [], resource_claims: {}, service_dependencies: [], ...entry,
  })).join("\n") + "\n");
  publish();
  const facts = await reduceCanonicalUnitIntervals(file);
  assert.equal(facts.phases.size, 128);
  assert.ok([...facts.phases.values()].every((phase) => phase.end_ms === null));
  entries.push({ event: "phase_started", interval_id: "phase:129", phase: "runner" });
  publish();
  await assert.rejects(reduceCanonicalUnitIntervals(file), /phase start/u);
});

test("sample acceptance preserves nullable availability and closed vocabulary", () => {
  const sample = { schema_id: "cartulary.harness_resource_sample.v2", seq: 1, elapsed_ms: 0,
    segment: 0, scope: "process", scope_ref: "process:1", identity_digest: null,
    availability: "available", metrics: {} };
  for (const availability of ["available", "unsupported", "permission_denied", "not_observed", "process_gone", "counter_reset", "paused_measurement", "truncated", "collector_failed"]) {
    const value = { ...sample, metrics: { cpu_user_us: { value: availability === "available" ? 0 : null, availability } } };
    validateSchemaSync(sample.schema_id, value);
    value.metrics.cpu_user_us.value = availability === "available" ? null : 0;
    assert.throws(() => validateSchemaSync(sample.schema_id, value));
  }
  for (const invalid of [{ ...sample, unknown: 1 }, { ...sample, metrics: { mystery: { value: 0, availability: "available" } } }, { ...sample, scope: "container" }]) {
    assert.throws(() => validateSchemaSync(sample.schema_id, invalid));
  }
});

test("exact selection never chooses a newest sibling", (t) => {
  const root = temporary(t);
  assert.throws(() => resolveExactRunDir(root), /RUN_ID/u);
  assert.throws(() => resolveExactRunDir(root, "missing"), /RUN_ID/u);
});

test("explicit export validates endpoints, private headers and transport failures", async (t) => {
  assert.equal(validatedEndpoint("http://127.0.0.1:4318").protocol, "http:");
  for (const endpoint of ["http://example.com", "https://user:secret@example.com", "https://example.com/?secret=x"]) assert.throws(() => validatedEndpoint(endpoint));
  const file = path.join(temporary(t), "headers.json");
  writeFileSync(file, JSON.stringify({ authorization: "test-only" }), { mode: 0o600 });
  assert.deepEqual(headersFromFile(file), { authorization: "test-only" });
  writeFileSync(file, JSON.stringify({ host: "forbidden" }));
  assert.throws(() => headersFromFile(file));
  let requests = 0;
  await assert.rejects(deliver(new URL("https://example.com"), {}, {}, 100, async (_url, options) => {
    requests++;
    assert.equal(options.redirect, "error");
    assert.ok(options.signal instanceof AbortSignal);
    return { redirected: false, status: 302, ok: false };
  }), /redirect/u);
  assert.equal(requests, 1);
});

test("generated definitions preserve acceptance and admit isolated ordinary extensions", async () => {
  const catalog = JSON.parse(readFileSync(new URL("../../../harness_instrumentation_definitions.json", import.meta.url), "utf8"));
  const lease = JSON.parse(readFileSync(new URL("../../../schemas/cartulary.harness_fixture_lease.v4.schema.json", import.meta.url), "utf8"));
  const ajv = new Ajv2020({ strict: false, validateFormats: false });
  ajv.addSchema(lease);
  const original = renderInstrumentationDefinitions(catalog);
  ajv.addSchema(original.schema);
  const value = ajv.compile({ $ref: `${original.schema.$id}#/$defs/sample_metrics` });
  for (const signal of catalog.signals) {
    assert.equal(value({ [signal.name]: { availability: "available", value: 0 } }), true);
    assert.equal(value({ [signal.name]: { availability: "not_observed", value: null } }), true);
    assert.equal(value({ [signal.name]: { availability: "available", value: null } }), false);
  }
  assert.equal(value({ synthetic_counter: { availability: "available", value: 0 } }), false);
  catalog.phases.push("synthetic_phase");
  catalog.signals.push({ name: "synthetic_counter", kind: "counter", unit: "count", producer_scope: "process", aggregation: "segmented_delta" },
    { name: "synthetic_gauge", kind: "gauge", unit: "count", producer_scope: "process", aggregation: "sampled_maximum" });
  ajv.removeSchema(original.schema.$id);
  const extended = renderInstrumentationDefinitions(catalog);
  ajv.addSchema(extended.schema);
  const extendedValue = ajv.compile({ $ref: `${extended.schema.$id}#/$defs/sample_metrics` });
  assert.equal(extendedValue({ synthetic_counter: { availability: "available", value: 2 }, synthetic_gauge: { availability: "available", value: 3 } }), true);
  assert.ok(extended.schema.$defs.phase.enum.includes("synthetic_phase"));
  const descriptors = await import(`data:text/javascript;base64,${Buffer.from(extended.runtime).toString("base64")}`);
  const accumulator = createResourceAccumulator({ signals: descriptors.signals });
  for (const [elapsed_ms, counter, gauge] of [[0, 2, 12], [10, 7, 8]]) accumulator.consume({
    scope_ref: "guest:synthetic", scope: "guest_context", availability: "available", segment: 0, identity_digest: null, elapsed_ms,
    metrics: { synthetic_counter: { value: counter, availability: "available" }, synthetic_gauge: { value: gauge, availability: "available" } },
  }, new Map());
  const metrics = accumulator.finish().rows[0].metrics;
  assert.deepEqual(metrics.synthetic_counter, { maximum: 7, observed_delta: 5, observed_interval_ms: 10, unavailable: 0 });
  assert.deepEqual(metrics.synthetic_gauge, { maximum: 12, observed_delta: null, observed_interval_ms: 0, unavailable: 0 });

  const duplicate = structuredClone(catalog); duplicate.signals.push(duplicate.signals[0]);
  assert.throws(() => renderInstrumentationDefinitions(duplicate), /duplicate/u);
});

test("foundation generation compiles before publication and rolls back partial installation", async (t) => {
  const { cpSync, mkdirSync, renameSync } = await import("node:fs");
  const { generateFoundationArtifacts } = await import("../../generated-artifacts/generate-foundation-schema-validators.mjs");
  const { publishGeneratedTransaction } = await import("../../generated-artifacts/generated-transaction.mjs");
  const root = temporary(t), repo = path.resolve(import.meta.dirname, "../../../..");
  for (const relative of ["tools/schemas", "tools/harness_schema_attachments.json", "tools/harness_instrumentation_definitions.json", "tools/harness_lifecycle_policy.json", "tools/task_surface_owner.json", "internal/testutil/harnessdiag/testdata/schema-probes.json"]) {
    mkdirSync(path.dirname(path.join(root, relative)), { recursive: true });
    cpSync(path.join(repo, relative), path.join(root, relative), { recursive: true });
  }
  generateFoundationArtifacts(root);
  const shared = path.join(root, "tools/schemas/cartulary.harness_instrumentation_defs.v2.schema.json");
  const before = readFileSync(shared);
  const catalogFile = path.join(root, "tools/harness_instrumentation_definitions.json");
  const catalog = JSON.parse(readFileSync(catalogFile)); catalog.phases.push("synthetic_phase"); writeFileSync(catalogFile, JSON.stringify(catalog));
  // Catalog is valid, but a dependent foundation schema cannot compile.
  writeFileSync(path.join(root, "tools/schemas/cartulary.harness_resource_sample.v2.schema.json"), "{");
  assert.throws(() => generateFoundationArtifacts(root));
  assert.deepEqual(readFileSync(shared), before);
  const rendered = temporary(t), relative = "tools/transaction-fixture.json";
  writeFileSync(path.join(root, relative), "old"); mkdirSync(path.join(rendered, "tools")); writeFileSync(path.join(rendered, relative), "new");
  assert.throws(() => publishGeneratedTransaction({ repoRoot: root, renderedRoot: rendered, generatedPaths: [relative],
    rename: (from, to) => { if (from.includes(`${path.sep}staging${path.sep}`)) throw new Error("install failed"); renameSync(from, to); },
  }), /install failed/u);
  assert.equal(readFileSync(path.join(root, relative), "utf8"), "old");
});
