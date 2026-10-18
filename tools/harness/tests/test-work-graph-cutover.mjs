#!/usr/bin/env node

import assert from "node:assert/strict";
import { existsSync, mkdtempSync, readFileSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { operationExecutionDigest } from "../scheduler/work-graph/operations.mjs";
import { selectionFor } from "../scheduler/work-graph/preparation.mjs";

import {
  WorkGraphCompiler,
  availableGoWorkerBudget,
  compileRetainedRun,
  assertFixtureServiceDependencies,
  assertServiceDependencies,
  buildWorkGraph,
} from "../scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const mode = process.argv[2] ?? "matrix";
assert.ok(["fast", "matrix"].includes(mode), `unknown graph compiler test mode ${mode}`);

function unit(unitID, claims, work, needs = []) {
  return {
    unit_id: unitID,
    owner_id: "harness.command_surface",
    kind: "runner",
    command: { executable: "true", args: [], environment: {} },
    needs,
    resource_claims: claims,
    fixture_lease: "none",
    service_dependencies: [],
    cache_policy: "none",
    timeout_ms: 1000,
    current_run_evidence_outputs: [],
    row_bindings: [],
    failure_policy: {
      block_descendants: true,
      continue_independent: true,
      aggregate_effect: "required",
    },
    estimated_work_ms: work,
  };
}

function expectedCurrentRunOutputs(unitValue, rowIDs) {
  const safeUnitID = unitValue.unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
  return [...rowIDs.map((rowID) => `rows/${rowID}.json`), `unit-results/${safeUnitID}.json`].sort();
}

function assertCanonicalCurrentRunOutputs(unitValue, rowIDs, label) {
  assert.deepEqual(
    unitValue.current_run_evidence_outputs,
    expectedCurrentRunOutputs(unitValue, rowIDs),
    `${label} must publish only selected row evidence and one canonical unit result`,
  );
  assert.equal(
    unitValue.current_run_evidence_outputs.filter((output) => output.startsWith("unit-results/"))
      .length,
    1,
    `${label} must publish exactly one canonical unit result`,
  );
  assert.equal(
    new Set(unitValue.current_run_evidence_outputs).size,
    unitValue.current_run_evidence_outputs.length,
    `${label} outputs must be unique`,
  );
}

function assertGraphCutover() {
  assert.equal(availableGoWorkerBudget(19, {}), 19);
  assert.equal(availableGoWorkerBudget(19, { CARTULARY_UNIT_CPU_TOKENS: "4" }), 19, "unleased ambient values cannot alter planning");
  const parent = { CARTULARY_HOST_ADMISSION_LEASE: "opaque-parent", CARTULARY_UNIT_CPU_TOKENS: "4" };
  assert.equal(availableGoWorkerBudget(19, parent), 4);
  assert.equal(availableGoWorkerBudget(2, parent), 2);
  for (const budget of [undefined, "", "0", "-1", "1.5", "4x", "9007199254740992"]) {
    assert.throws(() => availableGoWorkerBudget(19, { ...parent, CARTULARY_UNIT_CPU_TOKENS: budget }), /inherited Go CPU budget/u);
  }
  const compiler = new WorkGraphCompiler(root);
  const retainedInputs = { OWNER: "platform.telemetry" };
  const nested = new WorkGraphCompiler(root, { targetInputs: { "test-slice": retainedInputs } });
  nested.availableGoLanes = 4;
  nested.availablePostgresLanes = 2;
  const retainedManifest = {
    target: "test-slice",
    command_id: nested.taskSurface.targets.find((entry) => entry.name === "test-slice").command_id,
    declared_inputs: retainedInputs,
    capability_snapshot: { cpu_tokens: 19, postgres_lanes: 2 },
    go_worker_budget: 4,
  };
  assert.deepEqual(compileRetainedRun(root, retainedManifest), selectionFor({ selection: "command", target: "test-slice" }, nested), "retained owner selection must reconstruct its inherited worker budget, not physical host capacity");
  const savedEnvironment = { lease: process.env.CARTULARY_HOST_ADMISSION_LEASE, budget: process.env.CARTULARY_UNIT_CPU_TOKENS };
  try {
    process.env.CARTULARY_HOST_ADMISSION_LEASE = "unrelated-reader-parent";
    process.env.CARTULARY_UNIT_CPU_TOKENS = "invalid-reader-budget";
    const row = nested.catalog.rows.find((value) => value.owner_id === "platform.telemetry");
    for (const selection of [retainedInputs, { ...retainedInputs, ROWS: row.row_id }]) {
      for (const budget of [2, 4, 19]) {
        nested.targetInputs["test-slice"] = selection;
        nested.availableGoLanes = budget;
        assert.deepEqual(compileRetainedRun(root, { ...retainedManifest, declared_inputs: selection, go_worker_budget: budget }),
          selectionFor({ selection: "command", target: "test-slice" }, nested), "direct and nested retained graphs preserve exact rows and worker claims independently of reader environment");
      }
    }
    for (const budget of [undefined, null, "4", 0, -1, 1.5, 20, Number.MAX_SAFE_INTEGER + 1]) {
      assert.throws(() => compileRetainedRun(root, { ...retainedManifest, go_worker_budget: budget }), /retained Go worker budget/u);
    }
  } finally {
    for (const [name, value] of [["CARTULARY_HOST_ADMISSION_LEASE", savedEnvironment.lease], ["CARTULARY_UNIT_CPU_TOKENS", savedEnvironment.budget]]) {
      if (value === undefined) delete process.env[name]; else process.env[name] = value;
    }
  }
  const shellRow = compiler.catalog.rows.find((row) => row.runner === "shell");
  const shellOperation = compiler.owner.operations[shellRow.selector.operation_id];
  const shellUnit = () => compiler.compileRows([shellRow.row_id]).units.find((value) => value.unit_id === `row:${shellRow.row_id}`);
  const originalShellUnit = shellUnit();
  assert.equal(originalShellUnit.command.environment.CARTULARY_OPERATION_DIGEST, operationExecutionDigest(shellOperation), "shell dispatch binds its own operation independently of the enclosing operation");
  const originalExecution = shellOperation.execution;
  shellOperation.execution = { ...originalExecution, env: { ...originalExecution.env, EXECUTION_CHARACTERIZATION: "changed" } };
  assert.notEqual(shellUnit().semantic_digest, originalShellUnit.semantic_digest, "shell execution changes invalidate compiled identity");
  shellOperation.execution = originalExecution;
  const inputs = new WorkGraphCompiler(root, { targetInputs: {
    "test-evidence-audit": { OWNER: "platform.pagination", EVIDENCE_ROOTS_FILE: ".cartulary/audit-inputs/selected.json" },
  } });
  const invalidInputs = new WorkGraphCompiler(root, { targetInputs: { "test-evidence-audit": { UNDECLARED: "must-not-forward" } } });
  assert.throws(() => invalidInputs.compileTarget("test-evidence-audit"), /not declared/u);
  const audit = inputs.compileTarget("test-evidence-audit").units.find((unit) => unit.unit_id === "operation:harness.test_evidence_audit");
  assert.equal(audit.command.executable, "node");
  assert.deepEqual(audit.command.args, ["tools/harness/execution/operation-runner-cli.mjs", "harness.test_evidence_audit"]);
  assert.equal(audit.command.environment.OWNER, "platform.pagination");
  assert.equal(audit.command.environment.EVIDENCE_ROOTS_FILE, ".cartulary/audit-inputs/selected.json");
  assert.equal(JSON.stringify(audit).includes("must-not-forward"), false);
  assert.notEqual(audit.semantic_digest, new WorkGraphCompiler(root, { targetInputs: { "test-evidence-audit": { OWNER: "harness.browser", EVIDENCE_ROOTS_FILE: ".cartulary/audit-inputs/other.json" } } }).compileTarget("test-evidence-audit").units.find((unit) => unit.unit_id === audit.unit_id).semantic_digest);
  assert.deepEqual(inputs.compileTarget("harness.frontend_install"), compiler.compileTarget("harness.frontend_install"), "explicit audit input must not reach another command");
  const rendererRows = compiler.catalog.rows.filter((row) => row.runner === "vitest" && row.selector.project_id === "renderer-node");
  assert.ok(rendererRows.some((row) => row.evidence_class === "integration"));
  assert.ok(rendererRows.some((row) => row.evidence_class === "unit"));
  for (const row of rendererRows) {
    const graph = compiler.compileRows([row.row_id]);
    assert.ok(graph.units.some((unit) => unit.unit_id === "operation:harness.frontend_install"));
    assert.equal(graph.units.some((unit) => unit.unit_id === "operation:harness.playwright_install"), row.evidence_class === "integration");
    assert.ok(graph.units.every((unit) => unit.service_dependencies.length === 0), "standalone renderer must not acquire product services");
    const consumer = graph.units.find((unit) => unit.current_run_evidence_outputs.includes(`rows/${row.row_id}.json`));
    assert.equal(consumer.fixture_lease, "none");
    assert.equal(consumer.resource_claims.postgres, undefined);
    if (row.evidence_class === "integration") {
      assert.equal(consumer.resource_claims.browser_stack, 1);
      assert.equal(consumer.command.environment.CARTULARY_TEST_ROWS, row.row_id);
    }
  }
  const frontendRows = compiler.catalog.rows.filter((row) => row.runner === "vitest");
  const prepared = compiler.catalog.rows.find((row) => row.runner === "go" && row.preparation_operations?.includes("harness.playwright_install"));
  assert.ok(prepared, "cross-language renderer has an explicit preparation contract");
  const preparedGraph = compiler.compileRows([prepared.row_id]);
  const preparedConsumer = preparedGraph.units.find((unit) => unit.current_run_evidence_outputs.includes(`rows/${prepared.row_id}.json`));
  for (const target of prepared.preparation_operations) {
    assert.ok(preparedGraph.units.some((unit) => unit.unit_id === `operation:${target}`));
    assert.ok(preparedConsumer.needs.includes(`operation:${target}`), "preparation failure must block the consumer");
  }
  compiler.catalog.rowByID.set(prepared.row_id, { ...prepared, preparation_operations: undefined });
  try {
    const unprepared = compiler.compileRows([prepared.row_id]);
    const consumer = unprepared.units.find((unit) => unit.current_run_evidence_outputs.includes(`rows/${prepared.row_id}.json`));
    assert.notEqual(consumer.semantic_digest, preparedConsumer.semantic_digest, "preparation changes execution identity");
    assert.equal(unprepared.units.some((unit) => unit.unit_id === "operation:harness.playwright_install"), false, "capacity must not infer preparation");
  } finally { compiler.catalog.rowByID.set(prepared.row_id, prepared); }
  const installedRows = compiler.catalog.rows.filter((row) => row.preparation_operations?.includes("harness.build_report_renderer"));
  assert.deepEqual([...new Set(installedRows.map((row) => row.runner))].sort(), ["go", "vitest"]);
  for (const installed of installedRows) {
    const installedGraph = compiler.compileRows([installed.row_id]);
    const installation = installedGraph.units.find((unit) => unit.unit_id === "operation:harness.build_report_renderer");
    const installedConsumer = installedGraph.units.find((unit) => unit.current_run_evidence_outputs.includes(`rows/${installed.row_id}.json`));
    const dependencies = new Set();
    const visit = (id) => { if (dependencies.has(id)) return; dependencies.add(id); for (const next of installedGraph.units.find((unit) => unit.unit_id === id).needs) visit(next); };
    visit(installation.unit_id);
    assert.ok(dependencies.has("operation:harness.frontend_install"));
    assert.ok(installation.needs.includes("operation:harness.playwright_install"));
    assert.ok(installedConsumer.needs.includes(installation.unit_id));
    assert.ok(installedGraph.units.every((unit) => unit.service_dependencies.length === 0));
  }
  const firstFrontend = frontendRows.find((row) => frontendRows.filter((other) => other.selector.file === row.selector.file &&
    other.owner_id === row.owner_id && other.evidence_class === row.evidence_class && other.resource_profile_id === row.resource_profile_id).length > 1);
  const pair = frontendRows.filter((row) => row.selector.file === firstFrontend.selector.file && row.owner_id === firstFrontend.owner_id &&
    row.evidence_class === firstFrontend.evidence_class && row.resource_profile_id === firstFrontend.resource_profile_id).slice(0, 2);
  const ids = pair.map((row) => row.row_id).sort();
  const consumers = (graph) => graph.units.filter((unit) => unit.current_run_evidence_outputs.some((output) => ids.some((id) => output === `rows/${id}.json`)));
  const sharedGraph = compiler.compileRows(ids), shared = consumers(sharedGraph);
  assert.equal(shared.length, 1, "compatible file rows must share one fresh process");
  assert.equal(shared[0].command.environment.VITEST_MAX_WORKERS, "1");
  assert.equal(shared[0].resource_claims.cpu, 1);
  assertCanonicalCurrentRunOutputs(shared[0], ids, "Vitest file group");
  assert.deepEqual(compiler.compileRows([...ids].reverse()), sharedGraph);
  assert.deepEqual(compiler.compileOwner(pair[0].owner_id, ids), sharedGraph);
  const single = consumers(compiler.compileRows([ids[0]]));
  assertCanonicalCurrentRunOutputs(single[0], [ids[0]], "Vitest subset");
  assert.notEqual(single[0].unit_id, shared[0].unit_id, "selected membership owns semantic identity");
  for (const mutation of [
    { process_isolation: "exclusive" }, { evidence_class: "security" },
    { runtime_profile_id: "default" }, { resource_profile_id: "harness_contract" },
    { owner_id: pair[1].owner_id === "web.architecture" ? "harness.browser" : "web.architecture" },
    { selector: { ...pair[1].selector, project_id: pair[1].selector.project_id === "browser-unit" ? "harness-node" : "browser-unit" } },
  ]) {
    const prior = compiler.catalog.rowByID.get(pair[1].row_id);
    compiler.catalog.rowByID.set(prior.row_id, { ...prior, ...mutation });
    try { assert.equal(consumers(compiler.compileRows(ids)).length, 2, JSON.stringify(mutation)); }
    finally { compiler.catalog.rowByID.set(prior.row_id, prior); }
  }
  const topology = JSON.parse(
    readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"),
  );
  assert.throws(
    () => assertServiceDependencies(topology, "default", ["postgres", "postgres"], "duplicate"),
    /duplicate/u,
  );
  assert.throws(
    () => assertServiceDependencies(topology, "default", ["postgres", "object_store"], "unsorted"),
    /sorted/u,
  );
  assert.throws(
    () => assertServiceDependencies(topology, "default", ["message_bus"], "unknown"),
    /unknown/u,
  );
  assert.throws(
    () => assertServiceDependencies(topology, "none", ["postgres"], "runtime-incompatible"),
    /unavailable/u,
  );
  assert.throws(
    () => assertFixtureServiceDependencies("postgres_dedicated", [], "omitted"),
    /requires service dependency postgres/u,
  );
  assertCanonicalCurrentRunOutputs(buildWorkGraph([unit("bounded", { cpu: 1 }, 1)]).units[0], [], "unit-only fixture");
  assert.throws(
    () => buildWorkGraph([{ ...unit("omitted", { cpu: 1 }, 1), service_dependencies: undefined }]),
    /service_dependencies is required/u,
  );
  assert.throws(
    () => buildWorkGraph([unit("unbounded", {}, 1)]),
    /resource_claims must bound executable work/u,
  );
  const affectedRowID = "module.entities.integration.the_mention_resolve_route_persists_durable_resol_d68c1befd6";
  const affectedDirect = compiler.compile({ kind: "rows", row_ids: [affectedRowID] });
  const affectedOwner = compiler.compile({
    kind: "owner",
    owner_id: "module.entities",
    row_ids: [affectedRowID],
  });
  assert.deepEqual(affectedDirect, affectedOwner, "equivalent row selectors must compile identically");
  const affectedUnit = affectedDirect.units.find((entry) =>
    entry.current_run_evidence_outputs.includes(`rows/${affectedRowID}.json`),
  );
  assert.equal(affectedUnit.fixture_lease, "postgres_dedicated");
  assert.deepEqual(affectedUnit.service_dependencies, ["object_store", "postgres"]);
  assert.equal(affectedUnit.resource_claims.object_store, 1);
  assert.equal(affectedUnit.resource_claims.cpu, 1);
  assert.equal(affectedUnit.resource_claims.postgres, 1);
  assert.equal(affectedUnit.command.environment.CARTULARY_UNIT_CPU_TOKENS, "1");
  assert.equal(affectedUnit.command.environment.GOMAXPROCS, "1");

  const dedicatedRows = [
    "module.entities.integration.host_and_identity_create_routes_reuse_exact_matc_f443e2591f",
    "module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044",
  ];
  const dedicatedGraph = compiler.compile({ kind: "rows", row_ids: dedicatedRows });
  const dedicatedUnits = dedicatedGraph.units.filter((entry) =>
    dedicatedRows.some((rowID) =>
      entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`),
    ),
  );
  assert.equal(dedicatedUnits.length, 1, "compatible dedicated rows must share one Go process");
  assert.equal(dedicatedUnits[0].fixture_lease, "postgres_dedicated");
  assertCanonicalCurrentRunOutputs(dedicatedUnits[0], dedicatedRows, "dedicated Go batch");
  assert.equal(dedicatedUnits[0].command.environment.GOMAXPROCS, "1");

  const migrationRows = [
    "module.database_migrations.integration.production_ddl_v2_recurrence",
    "module.database_migrations.integration.production_preflight_state_matrix",
  ];
  const migrationGraph = compiler.compile({ kind: "rows", row_ids: migrationRows });
  const migrationUnits = migrationGraph.units.filter((entry) =>
    migrationRows.some((rowID) =>
      entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`),
    ),
  );
  assert.equal(migrationUnits.length, 1, "compatible migration rows must share one Go process");
  assert.equal(migrationUnits[0].fixture_lease, "postgres_migration");
  assertCanonicalCurrentRunOutputs(migrationUnits[0], migrationRows, "migration Go batch");

  const processGlobalRows = [
    "platform.jobs.integration.claim_recovery_and_publication",
    "platform.jobs.integration.operational_telemetry",
  ];
  const processGlobalGraph = compiler.compile({ kind: "rows", row_ids: processGlobalRows });
  const processGlobalUnits = processGlobalGraph.units.filter((entry) =>
    processGlobalRows.some((rowID) =>
      entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`),
    ),
  );
  assert.equal(
    processGlobalUnits.length,
    2,
    "an explicit process-global assertion must not share a Go process",
  );

  const managedProcessRows = [
    "module.extensions.integration.bc011_deadline_precedence_ef23af86ac",
    "module.extensions.integration.bc015_browser_availability_e0a71bee5d",
  ];
  const managedProcessGraph = compiler.compile({
    kind: "rows",
    row_ids: managedProcessRows,
  });
  const managedProcessUnits = managedProcessGraph.units.filter((entry) =>
    managedProcessRows.some((rowID) =>
      entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`),
    ),
  );
  assert.equal(managedProcessUnits.length, 2, "managed-process rows must remain process-isolated");
  assert.ok(managedProcessUnits.every((entry) => entry.fixture_lease === "managed_process"));

  const catalogPostgresUnit = compiler
    .compile({ kind: "target", target: "harness.suite_backend_integration" })
    .units.find((entry) => entry.current_run_evidence_outputs.includes("rows/harness.browser.integration.wstest_helper_contract.json"));
  assert.equal(catalogPostgresUnit.resource_claims.cpu, 1);
  assert.equal(catalogPostgresUnit.resource_claims.postgres, 1);
  assert.equal(catalogPostgresUnit.command.environment.GOMAXPROCS, "1");

  const ioHeavy = compiler.compile({
    kind: "rows",
    row_ids: ["platform.audit.integration.transactional_immutable_persistence"],
  }).units.find((entry) =>
    entry.current_run_evidence_outputs.includes(
      "rows/platform.audit.integration.transactional_immutable_persistence.json",
    ),
  );
  assert.equal(ioHeavy.resource_claims.io, 2, "io_heavy must be topology-authoritative");

  const first = compiler.compileAggregatePlan("test-fast");
  const second = compiler.compileAggregatePlan("test-fast");
  assert.deepEqual(first, second, "same-source graph compilation must be byte-stable");
  assert.ok(first.graph.units.length > 0);
  assert.ok(first.projections["test-fast"].length > 0);
  for (const graphTarget of ["test-fast", "check", "ci", "release-check"]) {
    const graph = compiler.compileAggregatePlan(graphTarget).graph;
    assert.equal(
      graph.units.some((entry) =>
        ["test-fast", "check", "ci", "release-check"].includes(
          entry.command.args.at(-1),
        )),
      false,
      `${graphTarget} must not invoke a nested aggregate`,
    );
  }
  const releaseGraph = compiler.compileAggregatePlan("release-check").graph;
  for (const entry of releaseGraph.units) {
    if (Object.keys(entry.resource_claims).length === 0) continue;
    assert.equal(
      entry.shared_locks.includes("host_activity") +
        entry.exclusive_locks.includes("host_activity"),
      1,
      `${entry.unit_id} must have exactly one host_activity lock mode`,
    );
  }
  const compatibility = releaseGraph.units.find(
    (entry) => entry.unit_id === "operation:harness.seaweedfs_compatibility",
  );
  assert.equal(compatibility?.fixture_lease, "object_store_namespace");
  assert.equal(compatibility?.resource_claims.object_store, 1);
  assert.equal(
    releaseGraph.units.some((entry) => entry.unit_id === "operation:harness.object_store_init"),
    false,
    "release compatibility must not start the local-development object store",
  );
}

assertGraphCutover();
