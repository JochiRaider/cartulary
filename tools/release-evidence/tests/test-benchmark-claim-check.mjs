import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import os from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { loadBenchmarkProfile, validateBenchmarkManifest } from "../check-benchmark-claim.mjs";
import { resolveHarnessConfig } from "../../harness/contract/harness-contract.mjs";
import { loadPerformanceFixtureSnapshotRegistry, postgresMigrationDigest, snapshotKey } from "../../harness/performance-fixture/index.mjs";
import { fixtureDigest } from "../../harness/performance-fixture/profile-adapters/large-grid.mjs";
import { loadTestCatalog } from "../../harness/test-catalog/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const checker = path.join(root, "tools/release-evidence/check-benchmark-claim.mjs");
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), "cartulary-benchmark-contract-"));
const digest = bytes => `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
const json = value => Buffer.from(`${JSON.stringify(value)}\n`);
const read = relative => JSON.parse(fs.readFileSync(path.join(root, relative), "utf8"));
const profile = loadBenchmarkProfile(root);
const registry = loadPerformanceFixtureSnapshotRegistry(root);
const fixture = registry.profiles.get("timeline_large_grid_snapshot_v2");
const catalog = loadTestCatalog(root);
const large = read("contracts/performance/large-grid.v1.json").fixture;
const interaction = read("contracts/performance/ac043.v2.json");
const paste = read("contracts/performance/ac003.v2.json");
const presence = read("contracts/performance/presence.v2.json");
const migration = postgresMigrationDigest(root);
const key = snapshotKey(fixture, migration);
const source = `sha256:${"a".repeat(64)}`;
let ordinal = 0;

// Synthetic validator fixtures only: no browser or product performance claim.
function packet({ criterion = "AC-043", mutate = {} } = {}) {
  const directory = path.join(temporary, String(++ordinal));
  const runRoot = path.join(directory, "run");
  fs.mkdirSync(runRoot, { recursive: true });
  const change = (kind, value) => { mutate[kind]?.(value); return value; };
  function write(relative, value) {
    const file = path.join(runRoot, relative);
    fs.mkdirSync(path.dirname(file), { recursive: true });
    const bytes = json(value);
    fs.writeFileSync(file, bytes);
    return { path: relative, sha256: digest(bytes) };
  }
  const refs = [];
  const isPresence = presence.criterion_ids.includes(criterion);
  const selected = isPresence ? presence : criterion === "AC-003" ? paste : interaction;
  for (const [index, predicate] of (selected.predicates ?? [selected]).entries()) {
    const prefix = `measurement-${index}`;
    const row = catalog.rows.find(row => row.fixture_profile_id === fixture.fixture_profile_id && row.verification_ids.some(id =>
      registry.verificationBindings.get(id)?.predicate_id === predicate.predicate_id));
    assert.ok(row);
    const single = criterion !== "AC-043";
    const threshold = predicate.threshold_ms ?? selected.threshold_ms;
    const observation = change("observation", {
      schema_id: "cartulary.frontend_measurement_observation.v4", fixture_profile_id: fixture.fixture_profile_id,
      criterion_id: isPresence ? presence.criterion_ids[0] : criterion, predicate_id: predicate.predicate_id, fixture_id: large.fixture_id, fixture_digest: fixtureDigest(large),
      measurement_policy_id: selected.measurement_policy_id,
      threshold_ms: threshold, warmup_samples: single ? 0 : 1, measured_samples: single ? 1 : 100,
      outcome: "passed", statistics: single ? { kind: "single", elapsed_ms: 10 } : { kind: "percentile", percentile: 95, p50_ms: 10, p95_ms: 10 },
      traffic: { counts: [{ fact_id: "analyst_sessions", value: large.analyst_sessions }, { fact_id: "background_sessions", value: large.background_sessions }],
        rates: [{ fact_id: "background_updates_per_second", value: large.background_updates_per_second }],
        conditions: [{ fact_id: "presence_enabled", value: true }, { fact_id: "target_row_excluded", value: true }] },
      samples: Array.from({ length: single ? 1 : 101 }, (_, sample_index) => ({ sample_index, warmup: !single && sample_index === 0, total_ms: 10, stages_ms: [{ fact_id: "accepted_to_visible_paint", value: 10 }] })),
      predicate_evidence: isPresence ? {
        schema_id: "cartulary.presence_measurement_proof.v1", kind: "timeline_presence", presence_kind: predicate.kind,
        actor_user_id: "00000000-0000-4000-8000-000000000001", view_schema_id: presence.view_schema_id,
        record_id: predicate.kind === "surface" ? null : "00000000-0000-4000-8000-000000000002", row_version: predicate.kind === "surface" ? null : 1,
        field_key: predicate.kind === "cell" ? presence.field_key : null, consecutive_frames: 2,
        clock_domain: "browser_monotonic_time_origin", origins_unchanged: true, initially_absent: true,
        binding_stable: true, action_count: 1, viewport_width: presence.viewport.width,
        viewport_height: presence.viewport.height, device_scale_factor: presence.viewport.device_scale_factor,
      } : single ? {
        kind: "timeline_paste_20x5", clipboard_format: "tsv", clipboard_mime: "text/plain", clipboard_delivery: "synthesized_event",
        viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1, row_count: 20, cell_count: 100,
        consecutive_frames: 2, all_new_ids: true, geometry_unchanged: true, target_bindings_stable: true,
        created_records: Array.from({ length: 20 }, (_, i) => ({ record_id: `00000000-0000-4000-8000-${String(i + 1).padStart(12, "0")}`, row_version: 1 })),
      } : null,
    });
    const observationRef = { role: "frontend_measurement_observation", path_kind: "file", format: "json", ...write(`${prefix}/observation.json`, observation) };
    const objectSnapshot = { manifest_sha256: source, object_count: fixture.object_population.object_count, total_bytes: fixture.object_population.total_bytes };
    const build = change("build", {
      schema_id: "cartulary.performance_fixture_snapshot.v3", fixture_profile_id: fixture.fixture_profile_id,
      fixture_version: fixture.fixture_version, seed: fixture.seed, snapshot_key_schema_id: "cartulary.performance_fixture_snapshot_key.v2", snapshot_key: key,
      migration_digest: migration, source_contract_digest: fixture.source_contract_digest,
      builder_unit_id: `fixture_snapshot:default:${fixture.fixture_profile_id}:${key}`, build_ordinal: 1, state: "sealed", object_snapshot: objectSnapshot,
      contribution_receipts: fixture.contributions.map(c => ({ contribution_id: c.contribution_id, version: c.version, owner_id: c.owner_id, counts: c.expected_receipt_counts.map(n => ({ count_id: n.count_id, actual: n.exact })) })),
      semantic_validation_digest: source,
      validation: { counts: fixture.semantic_expectations.counts.map(n => ({ expectation_id: n.expectation_id, actual: n.exact, passed: true })),
        conditions: fixture.semantic_expectations.conditions.map(n => ({ expectation_id: n.expectation_id, actual: n.required, passed: true })), connections_closed: true },
      redaction_policy_id: fixture.redaction_policy.policy_id, created_at: "2026-10-14T00:00:00.000Z",
    });
    const buildRef = { role: "performance_fixture_snapshot_build", path_kind: "file", format: "json", ...write(`${prefix}/build.json`, build) };
    const lease = change("lease", {
      schema_id: "cartulary.performance_fixture_snapshot_lease.v3", fixture_profile_id: fixture.fixture_profile_id, snapshot_key: key,
      builder_unit_id: build.builder_unit_id, row_id: row.row_id, predicate_id: predicate.predicate_id, lease_identity: source,
      clone_ordinal: 1, creation_state: "created", object_snapshot: objectSnapshot, isolation_result: "isolated",
      cleanup_results: ["bucket", "credential_copy", "database", "process", "session"].map(resource_class => ({ resource_class, outcome: "complete" })),
      cleanup_state: "complete", redaction_policy_id: fixture.redaction_policy.policy_id, finalized_at: "2026-10-14T00:01:00.000Z",
    });
    const leaseRef = { role: "performance_fixture_snapshot_lease", path_kind: "file", format: "json", ...write(`${prefix}/lease.json`, lease) };
    const summary = change("summary", {
      schema_id: "cartulary.frontend_measurement_summary.v5", row_id: row.row_id, fixture_profile_id: fixture.fixture_profile_id, snapshot_key: key,
      observation_artifact: observationRef, build_artifact: buildRef, lease_artifact: leaseRef, clone_ordinal: 1, scheduler_overlap_count: 0,
      rollup: { predicate_id: predicate.predicate_id, sample_count: observation.measured_samples, statistics: observation.statistics, threshold_ms: observation.threshold_ms, outcome: observation.outcome },
      qualification_outcome: "qualified",
    });
    const environment = change("environment", {
      schema_id: "cartulary.browser_execution_environment.v1", browser_engine: "chromium", browser_build: profile.environment.browser_build,
      browser_mode: "headed", playwright_version: profile.environment.benchmark_harness_version,
      node_version: `v${read("tools/toolchain_pins.json").node_version}`, platform: "linux", architecture: "x64",
    });
    const report = change("report", { config: { version: environment.playwright_version }, suites: [{ file: row.selector.file, specs: [{ file: row.selector.file, title: row.selector.titles[0], tests: [{ projectName: row.selector.project_id, results: [{ status: "passed", retry: 0, attachments: [
      { name: `${observation.schema_id}.${observation.predicate_id}`, body: json(observation).toString("base64") },
      { name: environment.schema_id, body: json(environment).toString("base64") },
    ] }] }] }] }] });
    refs.push({ summary: write(`${prefix}/summary.json`, summary), report: write(`${prefix}/report.json`, report), browser_environment: write(`${prefix}/environment.json`, environment) });
  }
  const runManifest = change("runManifest", {
    schema_id: "cartulary.harness_run_manifest.v3", run_id: `synthetic-${ordinal}`, target: "test-slice", command_id: "cartulary.harness.command.test_slice.v2",
    declared_inputs: {}, source_commit: "a".repeat(40), source_state: "clean", source_digest: source, toolchain_digest: source, system_digest: source, graph_digest: source,
    go_worker_budget: 2, capability_snapshot: { schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024, process_slots: 2, io_tokens: 1,
      postgres_lanes: 1, object_store_lanes: 1, writable_volume: true, port_lanes: 1, services: {}, sources: {}, snapshot_digest: source },
    cache_mode: "off", instrumentation: { mode: "off", policy_digest: source }, started_at: "2026-10-14T00:00:00.000Z",
  });
  const runSummary = change("runSummary", {
    schema_id: "cartulary.harness_run_summary.v1", run_id: runManifest.run_id, target: runManifest.target, status: "pass", failure_class: null, failure_reason: null,
    unit_counts: { total: refs.length, passed: refs.length, failed: 0, skipped: 0, cancelled: 0 }, wall_duration_ms: 1000, critical_path: [], actual_dependency_critical_path_ms: 1000,
    timing_accounting: { setup_ms: 0, fixture_ms: 0, execution_ms: 1000, collation_ms: 0, wrapper_ms: 0, unattributed_ms: 0, resource_blocking_ms: 0, process_count: 1 },
    resource_pressure: {}, cache: { bypass: refs.length, hit: 0, miss: 0 }, artifact_refs: [],
  });
  const bundle = change("bundle", { schema_id: "cartulary.benchmark_evidence_bundle.v1", runs: [{ root: "run", manifest: write("run-manifest.json", runManifest), summary: write("run-summary.json", runSummary), measurements: refs }] });
  const bundleBytes = json(bundle);
  fs.writeFileSync(path.join(directory, "bundle.json"), bundleBytes);
  const manifest = change("manifest", {
    schema_id: "cartulary.benchmark_manifest.v2", claim_bearing: true, benchmark_profile_id: profile.benchmark_profile_id,
    environment: structuredClone(profile.environment), source_digest: source, criterion_ids: [criterion],
    security_controls_state: Object.fromEntries(["authentication", "session_handling", "csrf_protection", "sanitization", "safe_preview_restrictions", "integrity_checks"].map(name => [name, true])),
    bundle: { path: "bundle.json", sha256: digest(bundleBytes) },
  });
  const file = path.join(directory, "manifest.json");
  fs.writeFileSync(file, json(manifest));
  return { directory, runRoot, file, manifest, bundle };
}

function errors(packet) { return validateBenchmarkManifest(packet.manifest, packet.directory); }
function rejected(mutate, pattern, criterion) {
  const candidate = packet({ mutate, ...(criterion ? { criterion } : {}) });
  assert.match(errors(candidate).join("; "), pattern);
}
function cli(file) { return spawnSync(process.execPath, [checker, ...(file ? [file] : [])], { cwd: root, encoding: "utf8" }); }
try {
  const percentile = packet();
  assert.deepEqual(errors(percentile), []);
  assert.equal(cli(percentile.file).status, 0);
  assert.deepEqual(errors(packet({ criterion: "AC-003" })), [], "single observation must not inherit p95 warm-up or sample count");
  for (const criterion of presence.criterion_ids) assert.deepEqual(errors(packet({ criterion })), []);
  assert.deepEqual(errors(packet({ criterion: "AC-132", mutate: { manifest: m => { m.criterion_ids = presence.criterion_ids; } } })), [],
    "one predicate set supplies both criteria without duplicate execution");
  rejected({ bundle: b => { b.runs[0].measurements.pop(); } }, /missing or extra/u, "AC-132");
  rejected({ observation: o => { o.criterion_id = "AC-132"; } }, /diverge/u, "AC-132");
  rejected({ manifest: m => { m.schema_id = "cartulary.benchmark_manifest.v1"; } }, /constant/u);
  for (const id of ["cartulary.perf.desktop_ref.v1", "cartulary.perf.desktop_ref.v2", "unknown"]) rejected({ manifest: m => { m.benchmark_profile_id = id; } }, /constant/u);
  rejected({ manifest: m => { m.unknown = true; } }, /additional/u);
  rejected({ manifest: m => { m.artifact_bundle_metadata_path = "self-asserted.json"; } }, /additional/u);
  rejected({ manifest: m => { m.criterion_ids = ["AC-000"]; } }, /no admitted/u);
  // AC-016 has its own admitted verification predicate; AC-043 observations
  // cannot qualify it merely by changing the claim's criterion label.
  rejected({ manifest: m => { m.criterion_ids = ["AC-016"]; } }, /undeclared or duplicate measurement predicate/u);
  rejected({ manifest: m => { m.security_controls_state.csrf_protection = false; } }, /constant/u);
  rejected({ manifest: m => { m.environment.browser_build = "wrong"; } }, /exact current/u);
  rejected({ bundle: b => { b.runs[0].measurements.pop(); } }, /missing or extra/u);
  rejected({ bundle: b => { b.runs[0].measurements.push(b.runs[0].measurements[0]); } }, /duplicate measurement/u);
  rejected({ bundle: b => { b.runs.push(b.runs[0]); } }, /duplicate benchmark/u);
  rejected({ runManifest: r => { r.source_digest = `sha256:${"b".repeat(64)}`; } }, /source-incompatible/u);
  rejected({ runSummary: r => { r.cache.hit = 1; } }, /cached/u);
  rejected({ runSummary: r => { r.status = "fail"; } }, /failed/u);
  rejected({ runSummary: r => { r.unit_counts.passed--; r.unit_counts.skipped++; } }, /incomplete/u);
  for (const state of ["failed", "skipped", "cancelled"]) rejected({ runSummary: r => { r.unit_counts[state] = 1; } }, /incomplete/u);
  rejected({ runSummary: r => { r.unit_counts.total = 0; r.unit_counts.passed = 0; } }, /incomplete/u);
  rejected({ environment: e => { e.browser_mode = "headless"; } }, /actual browser/u);
  rejected({ environment: e => { e.browser_build = "wrong"; } }, /actual browser/u);
  rejected({ environment: e => { e.node_version = "v0.0.0"; } }, /Node version/u);
  rejected({ report: r => { r.suites[0].specs[0].title += " suffix"; } }, /exact catalog selection/u);
  rejected({ report: r => { r.suites[0].specs[0].tests[0].projectName = "other"; } }, /exact catalog selection/u);
  rejected({ report: r => { r.suites[0].specs[0].file = "wrong.spec.ts"; } }, /exact catalog selection/u);
  rejected({ report: r => { r.suites[0].specs[0].tests[0].results[0].retry = 1; } }, /without retry/u);
  rejected({ report: r => { r.suites[0].specs[0].tests[0].results[0].attachments.pop(); } }, /not bound/u);
  rejected({ report: r => { r.suites[0].specs[0].tests[0].results[0].attachments[0].body = json({}).toString("base64"); } }, /required/u);
  rejected({ summary: s => { s.scheduler_overlap_count = 1; } }, /environment_not_qualified/u);
  rejected({ build: b => { b.source_contract_digest = "0".repeat(64); } }, /current owner/u);
  rejected({ build: b => { b.contribution_receipts[0].counts[0].actual++; } }, /current owner/u);
  rejected({ build: b => { b.validation.conditions[0].passed = false; } }, /current owner/u);
  rejected({ lease: l => { l.cleanup_state = "failed"; l.failure_code = "cleanup_failed"; l.cleanup_results[0].outcome = "failed"; } }, /allowed values|environment_not_qualified/u);
  rejected({ observation: o => { o.measured_samples = 99; o.samples.pop(); } }, /incomplete sampling/u);
  rejected({ observation: o => { o.statistics.p95_ms = 0; } }, /statistics/u);
  rejected({ observation: o => { o.fixture_id = "cartulary.perf.wrong.v1"; } }, /diverge/u);
  rejected({ observation: o => { o.samples[0].total_ms = 2000; o.statistics.elapsed_ms = 2000; } }, /threshold/u, "AC-003");
  rejected({ observation: o => { o.predicate_evidence.created_records[1].record_id = o.predicate_evidence.created_records[0].record_id; } }, /distinct/u, "AC-003");
  for (const relative of ["../bundle.json", "/absolute.json", "./bundle.json", "nested//bundle.json"]) rejected({ manifest: m => { m.bundle.path = relative; } }, /normalized and relative/u);
  const tampered = packet(); fs.appendFileSync(path.join(tampered.directory, "bundle.json"), " "); assert.match(errors(tampered).join(), /digest mismatch/u);
  const linked = packet(); fs.renameSync(path.join(linked.directory, "bundle.json"), path.join(linked.directory, "real.json")); fs.symlinkSync("real.json", path.join(linked.directory, "bundle.json")); assert.match(errors(linked).join(), /ELOOP/u);
  const hardlinked = packet(); fs.linkSync(path.join(hardlinked.directory, "bundle.json"), path.join(hardlinked.directory, "alias.json")); assert.match(errors(hardlinked).join(), /unaliased/u);
  const parent = packet(); fs.renameSync(parent.runRoot, path.join(parent.directory, "actual")); fs.symlinkSync("actual", parent.runRoot); assert.match(errors(parent).join(), /ENOTDIR|ELOOP/u);
  const originalOpen = fs.openSync;
  let forbiddenOpens = 0;
  try {
    const forbiddenDirectory = "docs";
    fs.openSync = (file, ...args) => {
      if (String(file).split(path.sep).includes(forbiddenDirectory) || /\.(?:md|markdown)$/iu.test(String(file))) { forbiddenOpens++; throw new Error("unexpected forbidden open"); }
      return originalOpen(file, ...args);
    };
    syncBuiltinESMExports();
    for (const name of [`${forbiddenDirectory}/bundle.json`, "README", "nested/source.MD", "nested/source.markdown"]) rejected({ manifest: m => { m.bundle.path = name; } }, /documentation/u);
    assert.equal(forbiddenOpens, 0);
  } finally { fs.openSync = originalOpen; syncBuiltinESMExports(); }
  const forbiddenDirectory = "docs";
  assert.match(cli(path.join(temporary, forbiddenDirectory, "manifest.json")).stderr, /documentation/u);
  assert.match(cli(path.join(temporary, "absent.json")).stderr, /benchmark manifest missing/u);
  const duplicate = path.join(temporary, "duplicate.json"); fs.writeFileSync(duplicate, '{"schema_id":"one","schema_id":"two"}'); assert.notEqual(cli(duplicate).status, 0);
  if (!fs.existsSync(path.join(root, ".cartulary/benchmark/benchmark_manifest.json"))) assert.match(cli().stdout, /no claim-bearing/u);
  const base = { HOME: process.env.HOME, CARTULARY_TEST_RUN_ID: "benchmark-input-contract" };
  const configured = { ...base, BENCHMARK_MANIFEST: "tmp/selected-benchmark.json", CARTULARY_MAKE_INPUT_SOURCES: "BENCHMARK_MANIFEST=cli" };
  const resolve = env => resolveHarnessConfig("benchmark-claim-check", env, { prepareRetainedArtifacts: false });
  assert.equal(resolve(base).variables.target_inputs.BENCHMARK_MANIFEST.value, ".cartulary/benchmark/benchmark_manifest.json");
  assert.equal(resolve(configured).variables.target_inputs.BENCHMARK_MANIFEST.value, configured.BENCHMARK_MANIFEST);
  assert.throws(() => resolve({ ...configured, BENCHMARK_MANIFEST: "" }), /must not be empty/u);
  assert.throws(() => resolveHarnessConfig("lint", configured), /not declared for target lint/u);
  console.log("benchmark publication contract passed");
} finally { fs.rmSync(temporary, { recursive: true, force: true }); }
