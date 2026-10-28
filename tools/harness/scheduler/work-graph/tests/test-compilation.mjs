import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import path from "node:path";
import { packInvocationCohorts } from "../cohorts.mjs";
import { operationExecutionDigest } from "../operations.mjs";
import { selectionFor } from "../preparation.mjs";

import {
  WorkGraphCompiler,
  availableGoWorkerBudget,
  resourceCapacities,
  compileRetainedRun,
  assertFixtureServiceDependencies,
  assertServiceDependencies,
  buildWorkGraph,
} from "../index.mjs";

const root = path.resolve(import.meta.dirname, "../../../../..");

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

function assertSizedGoPartition(units, rowIDs, fixture) {
  assert.ok(units.length > 0 && units.length <= rowIDs.length);
  assert.equal(new Set(units.map((entry) => entry.unit_id.split(":")[1])).size, 1,
    "compatible rows retain one compatibility partition across independently sized invocations");
  assert.deepEqual(units.flatMap((entry) => entry.row_bindings.map((row) => row.row_id)).sort(),
    [...rowIDs].sort(), "sizing executes exactly the selected rows without overlaps");
  for (const entry of units) {
    assert.equal(entry.fixture_lease, fixture);
    assertCanonicalCurrentRunOutputs(entry, entry.row_bindings.map((row) => row.row_id), `${fixture} invocation`);
    assert.equal(entry.command.environment.GOMAXPROCS, String(entry.resource_claims.cpu));
  }
}

test("retained selection preserves inherited capacity independently of reader environment", () => {
  assert.equal(availableGoWorkerBudget(19, {}), 19);
  assert.equal(availableGoWorkerBudget(19, { CARTULARY_UNIT_CPU_TOKENS: "4" }), 19, "unleased ambient values cannot alter planning");
  const parent = { CARTULARY_HOST_ADMISSION_LEASE: "opaque-parent", CARTULARY_UNIT_CPU_TOKENS: "4" };
  assert.equal(availableGoWorkerBudget(19, parent), 4);
  assert.equal(availableGoWorkerBudget(2, parent), 2);
  for (const budget of [undefined, "", "0", "-1", "1.5", "4x", "9007199254740992"]) {
    assert.throws(() => availableGoWorkerBudget(19, { ...parent, CARTULARY_UNIT_CPU_TOKENS: budget }), /inherited Go CPU budget/u);
  }
  const retainedInputs = { OWNER: "platform.telemetry" };
  const nested = new WorkGraphCompiler(root, { targetInputs: { "test-slice": retainedInputs } });
  nested.availableGoLanes = 4;
  nested.availablePostgresLanes = 2;
  const retainedManifest = {
    target: "test-slice",
    command_id: nested.taskSurface.targets.find((entry) => entry.name === "test-slice").command_id,
    declared_inputs: retainedInputs,
    capability_snapshot: { cpu_tokens: 19, postgres_lanes: 2,
      memory_bytes: 32 * 1024 ** 3, process_slots: 4096, io_tokens: 32,
      object_store_lanes: 8, port_lanes: 32, writable_volume: true,
      services: { browser: true, object_store: true, postgres: true, service_stack: true } },
    go_worker_budget: 4,
  };
  nested.capacities = Object.fromEntries(resourceCapacities(retainedManifest.capability_snapshot));
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
});

test("shell identity and audit inputs bind only their selected operation", () => {
  const compiler = new WorkGraphCompiler(root);
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
});

test("native boundary presets preserve catalog preparation and exact selection", () => {
  const compiler = new WorkGraphCompiler(root);
  const boundaryRows = compiler.catalog.rows.filter((row) => row.runner === "go" &&
    row.selector.package === "./internal/testutil/harnessruntime/boundary");
  assert.equal(boundaryRows.length, 4, "native boundary assurance has four independently owned row contracts");
  assert.deepEqual(boundaryRows.flatMap((row) => row.selector.tests).sort(), [
    "TestNativeBorrowedFixturePhysicalLifecycle", "TestNativeFixturePhysicalLifecycle",
    "TestNativeGraphServiceObservations", "TestNativeSelectedServiceLifecycle", "TestStandaloneNativeSession",
  ], "native boundary assurance must cover each package test exactly once");
  const cancellationRow = compiler.catalog.rows.find((row) => row.row_id === "harness.browser.integration.native_nested_cancellation");
  assert.equal(cancellationRow?.runner, "go");
  assert.deepEqual(cancellationRow.selector, {
    package: "./internal/testutil/harnessruntime", tests: ["TestRuntimeNestedServiceCancellation"],
  }, "native assurance must include the separate nested cancellation owner");
  const boundaryIDs = [...boundaryRows, cancellationRow].map((row) => row.row_id).sort();
  const boundaryGraph = compiler.compileRows(boundaryIDs);
  const boundaryCommand = new WorkGraphCompiler(root, { targetInputs: { "harness-contract": { SUITE: "runtime-boundary" } } });
  assert.deepEqual(boundaryCommand.compileTarget("harness-contract"), boundaryGraph,
    "direct runtime assurance must reuse catalog preparation, resource, isolation and deadline contracts");
  const boundaryConsumers = (graph) => graph.units.filter((unit) =>
    unit.row_bindings.some((row) => boundaryIDs.includes(row.row_id)));
  for (const selection of ["harness.contract_execution", "harness.contract_regression", "harness.contract_full"]) {
    const selected = compiler.compileTarget(selection);
    assert.deepEqual(boundaryConsumers(selected), boundaryConsumers(boundaryGraph),
      `${selection} must select the same canonical boundary invocations without widening or duplication`);
    assert.equal(selected.units.some((unit) => unit.unit_id === "operation:harness.harness_runtime_boundary_contract"), false);
  }
  assert.throws(() => compiler.compileTarget("harness.harness_runtime_boundary_contract"), /unknown public command/u);
  assert.throws(() => compiler.compileTarget("harness.assurance_work_graph_service_backed_target"), /unknown public command/u);
});

test("standalone renderer preparation follows evidence context", () => {
  const compiler = new WorkGraphCompiler(root);
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
      assert.deepEqual(consumer.row_bindings, [{ row_id: row.row_id, runner: "vitest" }]);
      assert.equal(consumer.command.args[1], "--unit-rows");
      assert.equal(consumer.command.environment.CARTULARY_TEST_ROWS, undefined);
    }
  }
});

test("cross-language and installed renderer preparation blocks consumers", () => {
  const compiler = new WorkGraphCompiler(root);
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
});

test("Vitest grouping preserves exact membership and isolation", () => {
  const compiler = new WorkGraphCompiler(root);
  const frontendRows = compiler.catalog.rows.filter((row) => row.runner === "vitest");
  const firstFrontend = frontendRows.find((row) => frontendRows.filter((other) => other.selector.file === row.selector.file &&
    other.owner_id === row.owner_id && other.evidence_class === row.evidence_class && other.resource_profile_id === row.resource_profile_id).length > 1);
  const pair = frontendRows.filter((row) => row.selector.file === firstFrontend.selector.file && row.owner_id === firstFrontend.owner_id &&
    row.evidence_class === firstFrontend.evidence_class && row.resource_profile_id === firstFrontend.resource_profile_id).slice(0, 2);
  const ids = pair.map((row) => row.row_id).sort();
  const consumers = (graph) => graph.units.filter((unit) => unit.current_run_evidence_outputs.some((output) => ids.some((id) => output === `rows/${id}.json`)));
  const sharedGraph = compiler.compileRows(ids), shared = consumers(sharedGraph);
  assert.equal(shared.length, 1, "compatible file rows must share one fresh process");
  assert.equal(shared[0].command.environment.VITEST_MAX_WORKERS, "1");
  assert.equal(shared[0].resource_claims.cpu, 2);
  assert.equal(shared[0].resource_claims.process, 4);
  assertCanonicalCurrentRunOutputs(shared[0], ids, "Vitest file group");
  assert.deepEqual(compiler.compileRows([...ids].reverse()), sharedGraph);
  assert.deepEqual(compiler.compileOwner(pair[0].owner_id, ids), sharedGraph);
  const single = consumers(compiler.compileRows([ids[0]]));
  assertCanonicalCurrentRunOutputs(single[0], [ids[0]], "Vitest subset");
  assert.notEqual(single[0].unit_id, shared[0].unit_id, "selected membership owns semantic identity");
  for (const mutation of [
    { process_isolation: "exclusive" }, { evidence_class: "security" },
    { runtime_profile_id: "default" }, { resource_profile_id: "harness_contract" },
    { selector: { ...pair[1].selector, project_id: pair[1].selector.project_id === "browser-unit" ? "harness-node" : "browser-unit" } },
  ]) {
    const prior = compiler.catalog.rowByID.get(pair[1].row_id);
    compiler.catalog.rowByID.set(prior.row_id, { ...prior, ...mutation });
    try { assert.equal(consumers(compiler.compileRows(ids)).length, 2, JSON.stringify(mutation)); }
    finally { compiler.catalog.rowByID.set(prior.row_id, prior); }
  }
  const changedOwner = { ...pair[1], owner_id: pair[1].owner_id === "web.architecture" ? "harness.browser" : "web.architecture" };
  compiler.catalog.rowByID.set(changedOwner.row_id, changedOwner);
  try {
    const mixed = consumers(compiler.compileRows(ids));
    assert.equal(mixed.length, 1, "reporting ownership cannot fragment a compatible invocation");
    assert.equal(mixed[0].owner_id, "harness.command_surface");
    assertCanonicalCurrentRunOutputs(mixed[0], ids, "mixed-owner Vitest group");
  } finally { compiler.catalog.rowByID.set(pair[1].row_id, pair[1]); }
  const authRows = frontendRows.filter((row) => row.selector.file === "apps/web/src/app/App.auth.test.tsx");
  assert.ok(authRows.length > 0);
  assert.deepEqual([...new Set(authRows.map((row) => row.owner_id))].sort(), ["module.auth", "web.application"]);
  const authIDs = authRows.map((row) => row.row_id).sort();
  const authGraph = compiler.compileRows(authIDs);
  const authUnits = authGraph.units.filter((unit) => unit.unit_id.startsWith("vitest_cohort:"));
  assert.equal(authUnits.length, 1, "App.auth rows across two owners execute in one isolated file worker");
  assert.equal(authUnits[0].owner_id, "harness.command_surface");
  assertCanonicalCurrentRunOutputs(authUnits[0], authIDs, "App.auth cohort");
  assert.deepEqual(compiler.compileRows([...authIDs].reverse()), authGraph);
  for (const owner of ["module.auth", "web.application"]) {
    const selected = authRows.filter((row) => row.owner_id === owner).map((row) => row.row_id).sort();
    const subset = compiler.compileOwner(owner, selected).units.filter((unit) => unit.unit_id.startsWith("vitest_cohort:"));
    assert.equal(subset.length, 1);
    assert.equal(subset[0].owner_id, owner);
    assertCanonicalCurrentRunOutputs(subset[0], selected, "App.auth exact owner subset");
    assert.throws(() => compiler.compileOwner(owner, authIDs), /does not belong/u);
  }
});

test("Go grouping preserves runtime binaries, isolation and reporting independence", () => {
  const compiler = new WorkGraphCompiler(root);
  const originalGoLanes = compiler.availableGoLanes;
  compiler.availableGoLanes = 1; // Prove compatibility independently of parallel sizing.
  const mixedGoIDs = ["module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2", "module.networkflow.integration.graph_result_cleanup"].sort();
  const mixedGo = compiler.compileRows(mixedGoIDs);
  const goConsumers = (graph) => graph.units.filter((unit) => unit.row_bindings.some((row) => mixedGoIDs.includes(row.row_id)));
  assert.equal(goConsumers(mixedGo).length, 1, "compatible reporting/networkflow rows share one Go package invocation");
  assert.equal(goConsumers(mixedGo)[0].owner_id, "harness.backend");
  assertCanonicalCurrentRunOutputs(goConsumers(mixedGo)[0], mixedGoIDs, "mixed-owner Go package");
  assert.deepEqual(compiler.compileRows([...mixedGoIDs].reverse()), mixedGo);
  const priorTarget = compiler.rowTargets.get(mixedGoIDs[0]);
  compiler.rowTargets.set(mixedGoIDs[0], "backend-store");
  try {
    const changedReporting = compiler.compileRows(mixedGoIDs);
    assert.deepEqual(goConsumers(changedReporting), goConsumers(mixedGo), "reporting targets do not change physical identity, costs or selection");
    const physical = goConsumers(changedReporting)[0];
    assert.equal(physical.command.environment.CARTULARY_TEST_TARGET, undefined);
    assert.equal(physical.command.environment.CARTULARY_TEST_ROWS, undefined);
    assert.deepEqual(physical.command.args.slice(1), ["--unit-rows"]);
  } finally { compiler.rowTargets.set(mixedGoIDs[0], priorTarget); }
  const goRow = compiler.catalog.rowByID.get(mixedGoIDs[0]);
  for (const mutation of [{ process_isolation: "exclusive" }, { postgres_isolation: "exclusive" },
    { preparation_operations: ["harness.playwright_install"] }, { resource_profile_id: "harness_contract" },
    { runtime_binary_ids: ["operator"] }]) {
    compiler.catalog.rowByID.set(goRow.row_id, { ...goRow, ...mutation });
    try { assert.equal(goConsumers(compiler.compileRows(mixedGoIDs)).length, 2, JSON.stringify(mutation)); }
    finally { compiler.catalog.rowByID.set(goRow.row_id, goRow); }
  }
  const reportingRow = compiler.catalog.rowByID.get(mixedGoIDs[0]);
  compiler.catalog.rowByID.set(reportingRow.row_id, { ...reportingRow, family_id: "app.operator.process", owner_id: "app.operator" });
  try {
    assert.deepEqual(goConsumers(compiler.compileRows(mixedGoIDs)), goConsumers(mixedGo),
      "reporting families cannot inject runtime binaries or build dependencies");
  } finally { compiler.catalog.rowByID.set(reportingRow.row_id, reportingRow); }
  for (const binary of compiler.topology.runtime_binaries) {
    assert.equal(goConsumers(mixedGo)[0].command.environment[binary.consumer_env], "",
      "undeclared ambient consumer paths are cleared before Go execution");
  }
  const operatorIDs = ["app.operator.process.object_store_init_process_transport_400591b90d",
    "platform.objectstore.support_integration.object_store_init_durable_effect_578bf55fdf"].sort();
  for (const ids of [operatorIDs, ...operatorIDs.map((id) => [id])]) {
    const selected = compiler.compileRows(ids);
    const consumers = selected.units.filter((unit) => unit.row_bindings.some((row) => ids.includes(row.row_id)));
    assert.equal(consumers.length, 1, "each exact selection closes its operator prerequisite; compatible targets may share execution");
    assert.deepEqual(consumers[0].row_bindings.map((row) => row.row_id).sort(), ids);
    assert.ok(consumers[0].needs.includes("operation:harness.build_operator"));
    assert.ok(selected.units.some((unit) => unit.unit_id === "operation:harness.build_operator"));
    const binary = compiler.topology.runtime_binaries.find((entry) => entry.id === "operator");
    assert.equal(consumers[0].command.environment.CARTULARY_OPERATOR_BIN,
      path.resolve(root, process.env[binary.output_make_variable] || binary.default_output_path));
    assert.equal(consumers[0].command.environment.CARTULARY_SERVER_HARNESS_BIN, "");
  }
  compiler.availableGoLanes = originalGoLanes;
});

test("compatible contract files share bounded invocations", () => {
  const compiler = new WorkGraphCompiler(root);
  const frontendRows = compiler.catalog.rows.filter((row) => row.runner === "vitest");
  const contractRows = frontendRows.filter((row) => row.owner_id === "package.view_contracts" && row.selector.project_id === "node-unit");
  const contractGraph = compiler.compileRows(contractRows.map((row) => row.row_id));
  const contractUnits = contractGraph.units.filter((unit) => unit.unit_id.startsWith("vitest_cohort:"));
  assert.ok(contractUnits.length < new Set(contractRows.map((row) => row.selector.file)).size, "compatible files amortize coordinator startup");
  assert.deepEqual(compiler.compileRows(contractRows.map((row) => row.row_id).reverse()), contractGraph);
});

test("cohort packing includes lifecycle and preserves oversized work", () => {
  const policy = { startup_ms: 10, cleanup_ms: 5 };
  const packed = packInvocationCohorts([{ id: "c", cost: 25 }, { id: "a", cost: 50 }, { id: "b", cost: 50 }], policy, 90);
  assert.deepEqual(packed.map((cohort) => [cohort.cost, cohort.items.map((item) => item.id)]), [[90, ["a", "c"]], [65, ["b"]]]);
  assert.deepEqual(packInvocationCohorts([{ id: "oversized", cost: 100 }], policy, 90).map((cohort) => cohort.cost), [115]);
  assert.deepEqual(packInvocationCohorts([], policy, 90), []);
});

test("service declarations and graph resources reject invalid models", () => {
  const compiler = new WorkGraphCompiler(root);
  const metadata = compiler.compileRows(["web.design.static.visual_fixture_registry"]);
  assert.ok(metadata.units.every((unit) => unit.fixture_lease === "none" && unit.service_dependencies.length === 0));
  assert.equal(metadata.units.length, 1, "static registry validation starts only its Node operation");
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
});

test("service-backed Go selections preserve resource and process isolation", () => {
  const compiler = new WorkGraphCompiler(root);
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
  assertSizedGoPartition(dedicatedUnits, dedicatedRows, "postgres_dedicated");

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
  assertSizedGoPartition(migrationUnits, migrationRows, "postgres_migration");

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
  const ordinaryProcessRows = managedProcessRows.map((rowID) => compiler.catalog.rowByID.get(rowID));
  try {
    // Exercise managed capability with explicit fixture-owned metadata.
    for (const row of ordinaryProcessRows) compiler.catalog.rowByID.set(row.row_id, {
      ...row, fixture_capability: "managed_process", runtime_profile_id: "default",
      resource_profile_id: "managed_process", service_dependencies: ["object_store", "postgres"], postgres_isolation: "exclusive",
    });
    const managedProcessGraph = compiler.compile({ kind: "rows", row_ids: managedProcessRows });
    const managedProcessUnits = managedProcessGraph.units.filter((entry) =>
      managedProcessRows.some((rowID) => entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`)),
    );
    assert.equal(managedProcessUnits.length, 2, "managed-process rows must remain process-isolated");
    assert.ok(managedProcessUnits.every((entry) => entry.fixture_lease === "managed_process"));
  } finally {
    for (const row of ordinaryProcessRows) compiler.catalog.rowByID.set(row.row_id, row);
  }

  const catalogPostgresUnit = compiler
    .compile({ kind: "target", target: "harness.suite_backend_integration" })
    .units.find((entry) => entry.current_run_evidence_outputs.includes("rows/harness.browser.integration.postgres_transaction_isolation.json"));
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

});
