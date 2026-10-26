import assert from "node:assert/strict";
import test from "node:test";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";
import { resolveCommandInputs } from "../contract/index.mjs";
import path from "node:path";
import { buildOperationSequence, operationBindingNames, readOperationOwner } from "../scheduler/work-graph/operations.mjs";
import { admitCommandSelection } from "../scheduler/work-graph/command-selection.mjs";

const root = new URL("../../../", import.meta.url).pathname;
const rows = (graph) => [...new Set(graph.units.flatMap((unit) => unit.current_run_evidence_outputs)
  .filter((file) => file.startsWith("rows/")).map((file) => file.slice(5, -5)))].sort();

test("command selectors resolve private operations without changing their identity", () => {
  const compiler = new WorkGraphCompiler(root);
  const variants = { production: "harness.build_production", server: "harness.build_server", "server-harness": "harness.build_server_harness",
    migrate: "harness.build_migrate", operator: "harness.build_operator", web: "harness.build_web", "web-measurement": "harness.build_web_measurement",
    "report-renderer": "harness.build_report_renderer" };
  for (const [component, selection] of Object.entries(variants)) {
    const result = compiler.compileCommand("build", { COMPONENT: component });
    assert.ok(result.graph.units.length > 0);
    if (component !== "production") assert.ok(result.graph.units.some((unit) => unit.unit_id === `operation:${selection}`));
  }
  assert.deepEqual(compiler.compileCommand("build", {}), compiler.compileCommand("build", { COMPONENT: "production" }));
  for (const invalid of [{ COMPONENT: "unknown" }, { COMPONENT: "" }, { COMPONENT: 1 }, { MAKE: "false" }, { GO_CACHE_DIR: "/tmp/foreign" }, { ARGS: "--override" }]) {
    assert.throws(() => compiler.compileCommand("build", invalid), /must (?:not )?be|not declared/u);
  }
  const renamed = structuredClone(compiler.taskSurface.targets.find((entry) => entry.name === "build"));
  renamed.name = "renamed-build";
  renamed.command_id = "cartulary.harness.command.renamed_build.v1";
  compiler.taskSurface.targets.push(renamed);
  assert.deepEqual(compiler.compileCommand("renamed-build", { COMPONENT: "web" }).graph, compiler.compileCommand("build", { COMPONENT: "web" }).graph);
});

test("owner and suite selection rejects conflicts and preserves support evidence and service scope", () => {
  const compiler = new WorkGraphCompiler(root);
  const variants = Object.fromEntries(["unit", "store", "integration", "process"].map((name) => [`backend-${name}`, `harness.suite_backend_${name}`]));
  variants["frontend-unit"] = "harness.suite_frontend_unit";
  for (const suite of Object.keys(variants)) assert.ok(rows(compiler.compileCommand("test-slice", { SUITE: suite }).graph).length > 0);
  const integration = rows(compiler.compileCommand("test-slice", { SUITE: "backend-integration" }).graph);
  const support = compiler.catalog.rows.filter((row) => compiler.rowTargets.get(row.row_id) === "backend-integration-support");
  assert.ok(support.length > 0);
  for (const row of support) assert.ok(integration.includes(row.row_id));
  const rowID = "harness.command_surface.unit.scheduler_smoke";
  const valid = { OWNER: "harness.command_surface", ROWS: rowID };
  const equivalent = { ...valid, ROWS: `${rowID},harness.command_surface.unit.native_runtime` };
  const reverse = { ...equivalent, ROWS: equivalent.ROWS.split(",").reverse().join(",") };
  assert.deepEqual(resolveCommandInputs("test-slice", equivalent, compiler.taskSurface), resolveCommandInputs("test-slice", reverse, compiler.taskSurface));
  assert.equal(compiler.compileCommand("test-slice", equivalent).graph.graph_digest, compiler.compileCommand("test-slice", reverse).graph.graph_digest);
  assert.deepEqual(rows(compiler.compileCommand("test-slice", valid).graph), [rowID]);
  for (const invalid of [{}, { ...valid, SUITE: "backend-unit" }, { SUITE: "backend-unit", ROWS: rowID }, { ...valid, ROWS: `${rowID},${rowID}` },
    { ...valid, SCOPE: "services" }, { ...valid, ROWS: "harness.browser.unit.fixture_broker_smoke" }, { ...valid, ROWS: "harness.command_surface.unit.absent" }]) {
    assert.throws(() => compiler.compileCommand("test-slice", invalid), /exactly one|requires|unique|scope|does not belong/u);
  }
  const scoped = compiler.compileCommand("test-slice", { OWNER: "harness.browser", SCOPE: "services" });
  assert.ok(rows(scoped.graph).every((id) => compiler.catalog.rowByID.get(id).service_dependencies.length > 0));
  compiler.compileRows = compiler.compileOperation = () => { throw new Error("selection admission attempted preparation"); };
  const supplied = { OWNER: { value: valid.OWNER, source: "make_command_line" }, ROWS: { value: rowID, source: "make_command_line" }, SCOPE: { value: "all", source: "default" } };
  assert.deepEqual(admitCommandSelection(compiler, "test-slice", supplied).rows, [rowID]);
  assert.throws(() => admitCommandSelection(compiler, "test-slice", { ...supplied, SCOPE: { value: "services", source: "make_command_line" } }),
    (error) => error.exit_code === 2 && error.failure_reason === "usage_error");
});

test("browser selection requires a suite and selects its rows and static prerequisites exactly once", () => {
  const compiler = new WorkGraphCompiler(root);
  const variants = Object.fromEntries(["functional", "stateful", "measurement", "accessibility", "visual", "support"].map((name) => [name, `harness.suite_browser_${name}`]));
  assert.throws(() => resolveCommandInputs("browser-e2e", {}, compiler.taskSurface), /required/u);
  for (const suite of Object.keys(variants)) {
    const graph = compiler.compileCommand("browser-e2e", { SUITE: suite }).graph;
    const selected = rows(graph);
    const stage = suite === "functional" ? "webserver_backed" : suite;
    const expected = compiler.catalog.rows.filter((row) => suite === "visual" ? row.aggregate_placement === "explicit_only" && (row.evidence_class === "visual" || row.row_id.includes(".visual_")) : row.runner === "playwright" && row.selector.stage === stage && row.aggregate_placement !== "explicit_only").map((row) => row.row_id).sort();
    if (suite === "visual") {
      const registry = "web.design.static.visual_fixture_registry";
      expected.push(registry); expected.sort();
      assert.equal(graph.units.filter((unit) => unit.row_bindings.some((row) => row.row_id === registry)).length, 1);
    }
    assert.deepEqual(selected, expected);
  }
});

test("package inspection union shares its one producer and assurance presets have explicit private selections", () => {
  const compiler = new WorkGraphCompiler(root);
  const graph = compiler.compileCommand("package-inspect", { CHECK: "all" }).graph;
  assert.equal(graph.units.filter((unit) => unit.unit_id === "operation:harness.package_release").length, 1);
  const variants = Object.fromEntries(["default", "smoke", "execution", "regression", "lifecycle", "full", "runtime", "fixtures", "runtime_boundary", "artifact_consumers", "ui_review"]
    .map((name) => [name === "default" ? "contract" : name.replaceAll("_", "-"), `harness.contract_${name}`]));
  for (const suite of Object.keys(variants)) {
    const selected = compiler.compileCommand("harness-contract", { SUITE: suite }).graph;
    assert.ok(selected.units.length > 0, suite);
    if (suite !== "ui-review") {
      assert.deepEqual(rows(selected).filter((row) =>
        /^harness\.browser\.(?:boundary_support\.ui_review_|integration\.ui_review_|unit\.native_review$)/u.test(row)),
      [], `${suite} excludes optional UI-review skill qualification`);
      assert.ok(selected.units.every((unit) => !unit.unit_id.includes("harness_ui_review")), suite);
    }
  }
  const full = compiler.compileCommand("harness-contract", { SUITE: "full" }).graph;
  assert.equal(full.units.filter((unit) => unit.current_run_evidence_outputs.includes("rows/harness.browser.unit.fixture_broker_smoke.json")).length, 1);
});

test("lint and source checks have closed selections and source validation cannot repair inputs", () => {
  const compiler = new WorkGraphCompiler(root);
  const variants = { standard: "harness.lint_standard", go: "harness.lint_go", types: "harness.frontend_typecheck",
    frontend: "harness.lint_biome", scripts: "harness.lint_scripts", shell: "harness.lint_shell",
    "frontend-boundary": "harness.frontend_import_boundary_check", "backend-boundary": "harness.backend_module_boundary_check", protocol: "harness.lint_protocol" };
  for (const check of Object.keys(variants)) assert.ok(compiler.compileCommand("lint", { CHECK: check }).graph.units.length > 0, check);
  const selections = Object.fromEntries(["generated", "schema", "catalog", "surface", "identity", "toolchain", "migration-input"]
    .map((name) => [name, `harness.source_${name.replaceAll("-", "_")}_selection`]));
  selections.all = "harness.source_all";
  for (const check of Object.keys(selections)) {
    const graph = compiler.compileCommand("source-check", { CHECK: check }).graph;
    assert.ok(graph.units.length > 0, check);
    assert.ok(graph.units.every((unit) => !unit.unit_id.startsWith("go:") && !unit.unit_id.startsWith("playwright:")));
    assert.ok(graph.units.every((unit) => !["operation:harness.generated_refresh", "operation:harness.generate_artifacts"].includes(unit.unit_id)));
  }
  assert.equal(compiler.owner.operations["harness.source_generated"].execution.env.CARTULARY_GENERATE_DRIFT_REFRESH, "0");
});



test("installation operations validate all file producers before acquisition", () => {
  const owner = readOperationOwner(root);
  const environment = Object.fromEntries(operationBindingNames(owner).map((name) => [name, `/installed/${name}`]));
  environment.MAKE = "make";
  for (const part of ["node", "frontend", "browser", "test_images", "full"]) {
    const invocations = buildOperationSequence(owner, `harness.bootstrap_${part}`, environment);
    assert.equal(invocations[0].command, "make");
    assert.ok(invocations[0].args.slice(3).every((file) => path.isAbsolute(file)));
    assert.equal(invocations.length, ["test_images", "full"].includes(part) ? 2 : 1);
  }
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_frontend", { ...environment, FRONTEND_INSTALL_STAMP: "retired-wrapper" }), /absolute file targets/u);
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_full", { ...environment, GO_TOOLCHAIN: undefined }), /binding|template/u);
});



test("static-analysis exact, preset and aggregate selections share canonical assurance units", () => {
  const compiler = new WorkGraphCompiler(root);
  const selections = [
    ["web.architecture.support.frontend_import_boundary_checker_assurance_393c407a70", "harness.assurance_frontend_import_boundaries"],
    ["web.architecture.support.fallow_static_and_protocol_analyzer_assurance_0794130901", "harness.assurance_fallow_static"],
  ];
  const exact = compiler.compileCommand("test-slice", { OWNER: "web.architecture", ROWS: selections.map(([id]) => id).join(",") }).graph;
  assert.deepEqual(rows(exact), selections.map(([id]) => id).sort());
  const graphs = [
    ...["contract", "regression", "full"].map((suite) => [suite, compiler.compileCommand("harness-contract", { SUITE: suite }).graph]),
    ["ci", compiler.compileAggregatePlan("ci").graph],
  ];
  for (const [selection, graph] of graphs) {
    for (const [row, operation] of selections) {
      const units = graph.units.filter((unit) => unit.unit_id === `row:${row}` || unit.unit_id === `operation:${operation}`);
      const expected = selection === "contract" && operation === "harness.assurance_fallow_static" ? 0 : 1;
      assert.equal(units.length, expected, `${selection}: ${operation} runs once when selected`);
      if (!expected) continue;
      const reference = exact.units.find((unit) => unit.unit_id === `row:${row}`);
      assert.equal(units[0].unit_id, reference.unit_id);
      assert.equal(units[0].semantic_digest, reference.semantic_digest);
      assert.equal(graph.units.filter((unit) => unit.current_run_evidence_outputs.includes(`rows/${row}.json`)).length, 1);
    }
  }
});
