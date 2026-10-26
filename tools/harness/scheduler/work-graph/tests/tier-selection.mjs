import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { validateSchemaSync } from "../../../contract/index.mjs";
import { loadExecutionTopology, executionTopologySchemaID } from "../../../generated-artifacts/execution-topology.mjs";
import { WorkGraphCompiler, validateWorkGraph } from "../index.mjs";
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
  "harness.browser.integration.object_store_fixture_roundtrip",
  "harness.browser.integration.postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277",
  "harness.browser.integration.postgre_sql_cleanup_rejects_invalid_input_before_3f68344cda",
  "harness.browser.integration.service_free_contract_assertions_for_postgres_cl_83b8534255",
  "harness.browser.unit.fixture_tls",
  "harness.browser.unit.managed_service_assertions_for_object_store_fixt_7b174abeda",
  "harness.browser.unit.service_free_contract_assertions_for_object_stor_aac986f098",
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
const retainedRows = [
  "harness.command_surface.unit.embedded_asset_layout",
  "harness.browser.integration.managed_service_assertions_for_runtime_reset_rec_79940f98ab",
  "harness.browser.integration.service_free_contract_assertions_for_runtime_res_04b3f3a218",
  "module.projections.testsupport.projection_capability_caller_matrix_6f21615b92",
  "module.database_migrations.unit.test_harness_targeted_operation_validation",
  "app.server.integration.http_harness_restart_storage",
  "app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831",
  "platform.postgres.integration.fresh_fixture_admission",
  "platform.postgres.support_unit.certificate_fixture_lifecycle",
];
const retainedPolicyTargets = [
  "harness.public_wrapper_smoke", "harness.json_shape_check", "harness.test_catalog_check",
  "harness.source_generated", "harness.generated_artifact_policy_check", "harness.migration_input_drift",
  "harness.backend_module_boundary_check", "harness.frontend_import_boundary_check",
  "go-gosec-targeted", "go-vulncheck",
];
function rowIDs(graph) {
  return graph.units.flatMap((unit) => unit.current_run_evidence_outputs
    .filter((output) => output.startsWith("rows/"))
    .map((output) => output.slice("rows/".length, -".json".length))).sort();
}
export function assertHarnessTierRouting(context) {
  const compiler = context.compiler;
  for (const rowID of fullTierSelfTests) {
    assert.equal(compiler.catalog.rowByID.get(rowID)?.aggregate_placement, "full", rowID);
  }
  const plans = new Map();
  for (const target of ["test-fast", "check", "ci", "release-check"]) {
    const plan = compiler.compileAggregatePlan(target);
    plans.set(target, plan);
    validateWorkGraph(plan.graph);
    const selected = rowIDs(plan.graph);
    assert.equal(new Set(selected).size, selected.length, `${target}: unique row evidence`);
    const includesFull = ["ci", "release-check"].includes(target);
    for (const rowID of fullTierSelfTests) {
      assert.equal(selected.includes(rowID), includesFull, `${target}: ${rowID}`);
    }
    assert.ok(plan.graph.units.every((unit) => !unit.unit_id.startsWith("raw_go:")), "raw package selection is retired");
    // Retain every other ordinary tier-selected row, including product tests
    // co-located with helper tests. Policy-added evidence may widen this set.
    const rank = compiler.owner.tier_order.indexOf(compiler.owner.aggregate_tiers[target]);
    for (const row of compiler.catalog.rows.filter((entry) => entry.aggregate_placement !== "explicit_only" && compiler.owner.tier_order.indexOf(entry.aggregate_placement) <= rank)) {
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
  }
  const check = plans.get("check");
  for (const rowID of retainedRows)
    assert.ok(rowIDs(check.graph).includes(rowID), rowID);
  for (const target of retainedPolicyTargets)
    assert.ok(check.projections[target]?.length > 0, target);
  assert.equal(check.projections["harness-contract"], undefined);
  const smoke = compiler.taskSurface.harness_tiers.fast.checks;
  assert.equal(smoke.length, 3);
  assert.deepEqual(smoke.map((name) => compiler.taskSurface.harness_checks.find((entry) => entry.name === name).gate_smoke_role).sort(), [
    "fixture_broker_semantic", "public_make_wrapper", "work_graph_scheduler_semantic",
  ]);
  assert.deepEqual(rowIDs(compiler.compileRows(fullTierSelfTests)), [...fullTierSelfTests].sort());
  const owners = new Set(fullTierSelfTests.map((id) => compiler.catalog.rowByID.get(id).owner_id));
  for (const owner of owners) {
    const explicit = rowIDs(compiler.compileOwner(owner));
    for (const rowID of fullTierSelfTests.filter((id) => compiler.catalog.rowByID.get(id).owner_id === owner)) {
      assert.ok(explicit.includes(rowID), `${owner}: ${rowID}`);
    }
  }
  for (const target of ["backend-unit", "backend-integration", "backend-process"]) {
    const direct = compiler.compileCommand("test-slice", { SUITE: target }).graph;
    assert.deepEqual(rowIDs(direct), [...compiler.targetRows.get(target), ...(target === "backend-integration" ? compiler.targetRows.get("backend-integration-support") : [])].filter((id) => compiler.catalog.rowByID.get(id).aggregate_placement !== "explicit_only").sort(), `${target}: complete direct catalog selection`);
  }
}
export function assertCatalogTopologyValidation(context) {
  const directory = mkdtempSync(path.join(tmpdir(), "cartulary-catalog-topology-"));
  const manifestPath = path.join(directory, "tools/execution_topology_manifest.json");
  try {
    mkdirSync(path.join(directory, "tools"));
    writeFileSync(path.join(directory, "tools/task_surface_owner.json"), JSON.stringify(context.taskSurface));
    writeFileSync(path.join(directory, "tools/harness_work_graph_owner.json"), JSON.stringify(context.compiler.owner));
    const candidates = [
      { ...context.topology, schema_id: "cartulary.execution_topology.v9" },
      { ...context.topology, schema_id: "cartulary.execution_topology.v10" },
      { ...context.topology, schema_id: "cartulary.execution_topology.v15" },
      { ...context.topology, go_targets: { targets: [], family_runtime_binaries: [] } },
      { ...context.topology, go_targets: { ...context.topology.go_targets, raw_go_aggregates: [] } },
      ...[null, "./internal/testutil/configtest", [], ["../foreign"], ["./same", "./same"]].map((value) => ({
        ...context.topology, go_targets: { ...context.topology.go_targets, catalog_packages: value },
      })),
    ];
    for (const candidate of candidates) {
      writeFileSync(manifestPath, JSON.stringify(candidate));
      assert.throws(() => validateSchemaSync(executionTopologySchemaID, JSON.parse(JSON.stringify(candidate))), /validation failed/u);
      assert.throws(() => loadExecutionTopology({ root: directory, manifestPath }), /validation failed|must declare schema_id/u);
      assert.throws(() => new WorkGraphCompiler(directory), /validation failed|must declare schema_id/u);
    }
  }
  finally {
    rmSync(directory, { recursive: true, force: true });
  }
}
export function assertCatalogTierSelection(context) {
  const compiler = context.compiler;
  compiler.ensureCatalog();
  const placements = [
    ["harness.command_surface.process.processtest_helper_contract", "fast"],
    ["harness.command_surface.unit.configtest_helper_contract", "standard"],
    ["harness.command_surface.unit.embedded_asset_layout", "full"],
    ["harness.browser.integration.wstest_helper_contract", "release"],
    ["harness.command_surface.unit.native_runtime", "explicit_only"],
  ];
  const rows = placements.map(([id, placement]) => ({ ...compiler.catalog.rowByID.get(id), aggregate_placement: placement }));
  assert.ok(rows.every((row) => row.row_id), "the authored fixture rows must exist");
  compiler._catalog = { ...compiler.catalog, rows, rowByID: new Map(rows.map((row) => [row.row_id, row])) };
  compiler.taskSurface.harness_checks = [];
  // This fixture isolates tier selection from policy and public command views.
  compiler.taskSurface.targets = compiler.taskSurface.targets.filter((entry) => compiler.owner.aggregate_tiers[entry.name]);
  for (const target of Object.keys(compiler.owner.aggregate_tiers)) {
    compiler.owner.aggregate_policy_roots[target] = [];
    compiler.owner.aggregate_policy_inherits[target] = [];
  }
  const expectedSelections = [
    ["test-fast", ["harness.command_surface.process.processtest_helper_contract"]],
    ["check", ["harness.command_surface.process.processtest_helper_contract", "harness.command_surface.unit.configtest_helper_contract"]],
    ["ci", ["harness.command_surface.process.processtest_helper_contract", "harness.command_surface.unit.configtest_helper_contract", "harness.command_surface.unit.embedded_asset_layout"]],
    ["release-check", ["harness.command_surface.process.processtest_helper_contract", "harness.command_surface.unit.configtest_helper_contract", "harness.command_surface.unit.embedded_asset_layout", "harness.browser.integration.wstest_helper_contract"]],
  ];
  for (const [target, expected] of expectedSelections) {
    const plan = compiler.compileAggregatePlan(target);
    assert.deepEqual(plan, compiler.compileAggregatePlan(target), `${target}: deterministic`);
    validateWorkGraph(plan.graph);
    assert.deepEqual(rowIDs(plan.graph), expected.toSorted());
    assert.equal(plan.graph.units.some((unit) => unit.unit_id === "operation:harness.test_service_images"), target === "release-check", `${target}: excluded service rows cannot introduce readiness`);
  }
  const explicit = "harness.command_surface.unit.native_runtime";
  assert.deepEqual(rowIDs(compiler.compileRows([explicit])), [explicit], "explicit selection ignores aggregate tier");
  rows[0].aggregate_placement = "unknown_future_tier";
  assert.throws(() => compiler.compileAggregatePlan("ci"), /invalid aggregate placement/u);
}
