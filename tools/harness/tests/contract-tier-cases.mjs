import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { validateSchemaSync } from "../contract/index.mjs";
import { loadExecutionTopology, executionTopologySchemaID } from "../generated-artifacts/execution-topology.mjs";
import { WorkGraphCompiler, validateWorkGraph } from "../scheduler/work-graph/index.mjs";

// Reviewed migration selection, independent of the tier values being tested.
const fullTierSelfTests = [
  "harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131",
  "harness.browser.boundary_support.atomic_private_harness_state_publication_and_val_6e8cd66ca6",
  "harness.browser.boundary_support.autosave_feedback_diagnostics",
  "harness.browser.boundary_support.cleanup_suite_9036a615b1",
  "harness.browser.boundary_support.evidenceupload_suite_6f27a1c9e4",
  "harness.browser.boundary_support.fixtureidentity_suite_827c446bfe",
  "harness.browser.boundary_support.fixtures_suite_b570a8a829",
  "harness.browser.boundary_support.harnessstate_suite_a4c7c5d935",
  "harness.browser.boundary_support.incidentsocket_suite_6db375c4ec",
  "harness.browser.boundary_support.installed_only_images",
  "harness.browser.boundary_support.mentions_suite_e2e13939bd",
  "harness.browser.boundary_support.mutationanchors_suite_ba8fab1244",
  "harness.browser.boundary_support.ownercharacterization_suite_3242ca020b",
  "harness.browser.boundary_support.public_http_diagnostic_privacy",
  "harness.browser.boundary_support.publicjsonclient_suite_a3aedb76f9",
  "harness.browser.boundary_support.record_stable_collaboration_summary_polling_604f761c28",
  "harness.browser.boundary_support.savedviews_suite_79e8036a1a",
  "harness.browser.boundary_support.sessions_suite_4d7bc21669",
  "harness.browser.boundary_support.testcontrolclient_suite_f60081d8f5",
  "harness.browser.boundary_support.timeline_investigation_recipe",
  "harness.browser.boundary_support.timingsupport_dom_suite_6614d14ea5",
  "harness.browser.boundary_support.ui_review_artifacts",
  "harness.browser.boundary_support.ui_review_contract",
  "harness.browser.boundary_support.ui_review_execution",
  "harness.browser.boundary_support.ui_review_lifecycle",
  "harness.browser.boundary_support.ui_review_presentation",
  "harness.browser.boundary_support.ui_review_workflow",
  "harness.browser.boundary_support.visual_preference_isolation",
  "harness.browser.boundary_support.visual_presentation_profiles",
  "harness.browser.integration.object_store_fixture_roundtrip",
  "harness.browser.integration.postgres_cleanup_target_scoped_coordination",
  "harness.browser.integration.ui_review_seeded_default",
  "harness.browser.integration.ui_review_seeded_network_flow_claimed",
  "harness.browser.unit.fixture_tls",
  "harness.browser.unit.object_store_fixture_admission",
  "harness.browser.unit.object_store_secure_origins",
  "harness.browser.unit.performance_fixture_lifecycle",
  "harness.browser.unit.performance_fixture_runtime_and_assembler",
  "harness.browser.unit.performance_fixture_snapshot_key",
  "harness.browser.unit.recovery_fixture_server_lifecycle",
  "harness.browser.unit.source_owner_contribution_assembler",
  "harness.browser.unit.testservices_lifecycle_contract",
  "harness.command_surface.behavior.public_registry_parity",
  "harness.evidence_accounting.behavior.current_epoch_evidence",
  "harness.test_catalog.behavior.routing_contract",
  "package.test_utils.boundary_support.index_suite_466a5263a0"
];
const rawGroups = [
  "backend-unit-configtest",
  "backend-unit-suiteservices",
  "backend-integration-testutil",
  "backend-process-processtest",
];
const retainedRows = [
  "harness.command_surface.unit.embedded_asset_layout",
  "harness.browser.integration.runtime_reset_recovery_purpose_contract",
  "module.projections.testsupport.projection_capability_caller_matrix_6f21615b92",
  "module.database_migrations.unit.test_harness_targeted_operation_validation",
  "app.server.integration.http_harness_restart_storage",
  "app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831",
  "platform.postgres.integration.fresh_fixture_admission",
  "platform.postgres.support_unit.certificate_fixture_lifecycle",
];
const retainedPolicyTargets = [
  "check-harness-smoke", "json-shape-check", "test-catalog-check",
  "generate-drift", "generated-artifact-policy-check", "migration-input-drift",
  "backend-module-boundary-check", "frontend-import-boundary-check",
  "go-gosec-targeted", "go-vulncheck",
];

function rowIDs(graph) {
  return graph.units.flatMap((unit) => unit.current_run_evidence_outputs
    .filter((output) => output.startsWith("rows/"))
    .map((output) => output.slice("rows/".length, -".json".length))).sort();
}

export function assertHarnessTierRouting(context) {
  const compiler = context.compiler;
  assert.equal(fullTierSelfTests.length, 46);
  for (const rowID of fullTierSelfTests) {
    assert.equal(compiler.catalog.rowByID.get(rowID)?.minimum_tier, "full", rowID);
  }
  assert.deepEqual(compiler.rawGoAggregates.map((entry) => entry.id).sort(), [...rawGroups].sort());
  assert.ok(compiler.rawGoAggregates.every((entry) => entry.minimum_tier === "full"));
  const plans = new Map();
  for (const target of ["test-fast", "check", "test", "ci", "release-check"]) {
    const plan = compiler.compileAggregatePlan(target);
    plans.set(target, plan);
    validateWorkGraph(plan.graph);
    const selected = rowIDs(plan.graph);
    assert.equal(new Set(selected).size, selected.length, `${target}: unique row evidence`);
    const includesFull = ["test", "ci", "release-check"].includes(target);
    for (const rowID of fullTierSelfTests) {
      assert.equal(selected.includes(rowID), includesFull, `${target}: ${rowID}`);
    }
    for (const id of rawGroups) {
      assert.equal(plan.graph.units.some((unit) => unit.unit_id === `raw_go:${id}`), includesFull, `${target}: ${id}`);
    }
    // Retain every other ordinary tier-selected row, including product tests
    // co-located with helper tests. Policy-added evidence may widen this set.
    const rank = compiler.owner.tier_order.indexOf(compiler.owner.aggregate_tiers[target]);
    for (const row of compiler.catalog.rows.filter((entry) => compiler.owner.tier_order.indexOf(entry.minimum_tier) <= rank)) {
      assert.ok(selected.includes(row.row_id), `${target} lost ${row.row_id}`);
    }
    assert.deepEqual(plan.projections[target], plan.graph.units.map((unit) => unit.unit_id));
    const byID = new Map(plan.graph.units.map((unit) => [unit.unit_id, unit]));
    for (const [projection, members] of Object.entries(plan.projections)) {
      assert.equal(new Set(members).size, members.length, projection);
      for (const id of members) {
        assert.ok(byID.has(id), `${projection}: dangling ${id}`);
        assert.ok(byID.get(id).needs.every((dependency) => members.includes(dependency)), `${projection}: dependency closure`);
      }
    }
    for (const rowID of fullTierSelfTests) {
      const projection = plan.projections[compiler.rowTargets.get(rowID)] ?? [];
      const members = plan.graph.units.filter((unit) => unit.current_run_evidence_outputs.includes(`rows/${rowID}.json`));
      assert.ok(members.every((unit) => projection.includes(unit.unit_id)), `${target}: row projection ${rowID}`);
    }
    for (const entry of compiler.rawGoAggregates) {
      assert.equal((plan.projections[entry.target] ?? []).includes(`raw_go:${entry.id}`), includesFull, `${target}: raw projection ${entry.id}`);
    }
  }
  const check = plans.get("check");
  for (const rowID of retainedRows) assert.ok(rowIDs(check.graph).includes(rowID), rowID);
  for (const target of retainedPolicyTargets) assert.ok(check.projections[target]?.length > 0, target);
  assert.equal(check.projections["harness-contract"], undefined);
  const smoke = compiler.taskSurface.harness_tiers.fast.checks;
  assert.equal(smoke.length, 3);
  assert.deepEqual(smoke.map((name) => compiler.taskSurface.harness_checks.find((entry) => entry.name === name).gate_smoke_role).sort(), [
    "fixture_broker_semantic", "public_make_wrapper", "work_graph_scheduler_semantic",
  ]);
  assert.deepEqual(rowIDs(compiler.compileRows(fullTierSelfTests)), fullTierSelfTests);
  const owners = new Set(fullTierSelfTests.map((id) => compiler.catalog.rowByID.get(id).owner_id));
  for (const owner of owners) {
    const explicit = rowIDs(compiler.compileOwner(owner));
    for (const rowID of fullTierSelfTests.filter((id) => compiler.catalog.rowByID.get(id).owner_id === owner)) {
      assert.ok(explicit.includes(rowID), `${owner}: ${rowID}`);
    }
  }
  for (const target of ["backend-unit", "backend-integration", "backend-process"]) {
    const direct = compiler.compileTarget(target);
    for (const entry of compiler.rawGoAggregates.filter((value) => value.target === target)) {
      const id = `raw_go:${entry.id}`;
      assert.deepEqual(direct.units.find((unit) => unit.unit_id === id), plans.get("test").graph.units.find((unit) => unit.unit_id === id));
    }
  }
}

export function assertRawTierValidation(context) {
  const directory = mkdtempSync(path.join(tmpdir(), "cartulary-raw-tier-"));
  const manifestPath = path.join(directory, "tools/execution_topology_manifest.json");
  try {
    mkdirSync(path.join(directory, "tools"));
    writeFileSync(path.join(directory, "tools/task_surface_owner.json"), JSON.stringify(context.taskSurface));
    writeFileSync(path.join(directory, "tools/harness_work_graph_owner.json"), JSON.stringify(context.compiler.owner));
    const invalid = [];
    for (const tier of [undefined, null, "", "unknown", "FULL", false, 1, ["full"]]) {
      const candidate = structuredClone(context.topology);
      if (tier === undefined) delete candidate.go_targets.raw_go_aggregates[0].minimum_tier;
      else candidate.go_targets.raw_go_aggregates[0].minimum_tier = tier;
      invalid.push(candidate);
    }
    const omitted = structuredClone(context.topology);
    delete omitted.go_targets.raw_go_aggregates;
    invalid.push(omitted);
    const legacy = { ...context.topology, schema_id: "cartulary.execution_topology.v8" };
    invalid.push(legacy);
    for (const candidate of invalid) {
      writeFileSync(manifestPath, JSON.stringify(candidate));
      assert.throws(() => validateSchemaSync(executionTopologySchemaID, candidate), /validation failed/u);
      assert.throws(() => loadExecutionTopology({ root: directory, manifestPath }), /validation failed|must declare schema_id/u);
      assert.throws(() => new WorkGraphCompiler(directory), /validation failed|must declare schema_id/u);
    }
    for (const tier of ["fast", "standard", "full", "release"]) {
      const candidate = structuredClone(context.topology);
      candidate.go_targets.raw_go_aggregates.forEach((entry) => { entry.minimum_tier = tier; });
      writeFileSync(manifestPath, JSON.stringify(candidate));
      const normalized = loadExecutionTopology({ root: directory, manifestPath });
      assert.ok(normalized.goTargets.rawAggregates.every((entry) => entry.minimumTier === tier));
      assert.ok(new WorkGraphCompiler(directory).rawGoAggregates.every((entry) => entry.minimum_tier === tier));
    }
    for (const mutate of [
      (candidate) => candidate.go_targets.raw_go_aggregates.push(candidate.go_targets.raw_go_aggregates[0]),
      (candidate) => { candidate.go_targets.raw_go_aggregates[0].target = "backend-absent"; },
    ]) {
      const candidate = structuredClone(context.topology);
      mutate(candidate);
      writeFileSync(manifestPath, JSON.stringify(candidate));
      assert.throws(() => loadExecutionTopology({ root: directory, manifestPath }), /duplicate|unknown backend target/u);
      assert.throws(() => new WorkGraphCompiler(directory), /duplicate|unknown backend target/u);
    }
    const empty = structuredClone(context.topology);
    empty.go_targets.raw_go_aggregates = [];
    writeFileSync(manifestPath, JSON.stringify(empty));
    assert.deepEqual(new WorkGraphCompiler(directory).rawGoAggregates, []);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}

export function assertIndependentRawTierSelection(context) {
  const compiler = context.compiler;
  compiler.ensureCatalog();
  // No catalog rows or policy roots can incidentally admit a backend target.
  compiler._catalog = { ...compiler.catalog, rows: [], rowByID: new Map() };
  compiler._targetRows = new Map();
  compiler._rowTargets = new Map();
  for (const target of Object.keys(compiler.owner.aggregate_tiers)) {
    compiler.owner.aggregate_policy_roots[target] = [];
    compiler.owner.aggregate_policy_inherits[target] = [];
  }
  compiler.rawGoAggregates = compiler.rawGoAggregates.map((entry, index) => ({
    ...entry, minimum_tier: compiler.owner.tier_order[index],
  }));
  for (const [target, count] of [["test-fast", 1], ["check", 2], ["test", 3], ["ci", 3], ["release-check", 4]]) {
    const plan = compiler.compileAggregatePlan(target);
    assert.deepEqual(plan, compiler.compileAggregatePlan(target), `${target}: deterministic`);
    validateWorkGraph(plan.graph);
    assert.deepEqual(plan.graph.units.filter((unit) => unit.unit_id.startsWith("raw_go:")).map((unit) => unit.unit_id).sort(),
      compiler.rawGoAggregates.slice(0, count).map((entry) => `raw_go:${entry.id}`).sort());
    assert.equal(plan.graph.units.some((unit) => unit.unit_id === "target:test-service-images"), count >= 3,
      `${target}: only selected service consumers introduce readiness`);
  }
  compiler.rawGoAggregates = [];
  assert.deepEqual(compiler.compileAggregate("release-check").units, []);
}
