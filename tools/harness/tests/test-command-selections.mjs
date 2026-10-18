import assert from "node:assert/strict";
import test from "node:test";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";
import { resolveCommandInputs } from "../contract/index.mjs";
import { validateCommandSelectionBinding } from "../generated-artifacts/task-surface/input-validation.mjs";
import { buildTargetPlan, readPlanInputs, renderTargetPlan } from "../diagnostics/target-plan.mjs";
import { buildOwnerTaskGuide } from "../diagnostics/owner-diagnostics.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import fs, { mkdtempSync, writeFileSync, rmSync, symlinkSync } from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import { tmpdir } from "node:os";
import path from "node:path";
import { buildOperationSequence, operationBindingNames, readOperationOwner } from "../scheduler/work-graph/operations.mjs";
import { admitCommandSelection } from "../scheduler/work-graph/command-selection.mjs";

const root = new URL("../../../", import.meta.url).pathname;
const input = (name, type, extra = {}) => ({ name, type, binding: "make_variable", allowed_sources: ["make_command_line"],
  required: false, default: null, empty_string: "invalid", normalization: "trim", invalid_reason: "usage_error", summary_emission: "value", child_forwarding: "argv", ...extra });
function bind(compiler, target, selector, variants, options = {}) {
  const entry = compiler.taskSurface.targets.find((value) => value.name === target);
  entry.selection_binding = { kind: options.kind ?? "selector", input: selector, variants };
  entry.input_contract = { undeclared_make_command_line: "usage_error", undeclared_inherited_env: "ignore",
    inputs: [input(selector, "enum", { values: Object.keys(variants), ...options.input }), ...(options.inputs ?? [])], constraints: options.constraints ?? [] };
  const errors = [];
  validateCommandSelectionBinding(errors, entry);
  assert.deepEqual(errors, []);
  return entry;
}
const rows = (graph) => [...new Set(graph.units.flatMap((unit) => unit.current_run_evidence_outputs)
  .filter((file) => file.startsWith("rows/")).map((file) => file.slice(5, -5)))].sort();

test("command selectors resolve private operations without changing their identity", () => {
  const compiler = new WorkGraphCompiler(root);
  const variants = { production: "harness.build_production", server: "harness.build_server", "server-harness": "harness.build_server_harness",
    migrate: "harness.build_migrate", operator: "harness.build_operator", web: "harness.build_web", "web-measurement": "harness.build_web_measurement",
    "report-renderer": "harness.build_report_renderer" };
  bind(compiler, "build", "COMPONENT", variants, { input: { default: "production" } });
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
  bind(compiler, "test-slice", "SUITE", variants, { kind: "owner_or_suite",
    inputs: [input("OWNER", "owner_id"), input("ROWS", "row_ids"), input("SCOPE", "enum", { values: ["all", "services"], default: "all" })],
    constraints: [{ kind: "exactly_one", inputs: ["OWNER", "SUITE"] }, { kind: "requires", input: "ROWS", required_input: "OWNER" }] });
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

test("browser selection requires a suite and directly selects every suite row", () => {
  const compiler = new WorkGraphCompiler(root);
  const variants = Object.fromEntries(["all", "functional", "stateful", "measurement", "accessibility", "visual", "support"].map((name) => [name, `harness.suite_browser_${name}`]));
  bind(compiler, "browser-e2e", "SUITE", variants, { input: { required: true } });
  assert.throws(() => resolveCommandInputs("browser-e2e", {}, compiler.taskSurface), /required/u);
  for (const suite of Object.keys(variants)) {
    const selected = rows(compiler.compileCommand("browser-e2e", { SUITE: suite }).graph);
    const stage = suite === "functional" ? "webserver_backed" : suite;
    const expected = compiler.catalog.rows.filter((row) => row.runner === "playwright" && (suite === "all" || row.selector.stage === stage)).map((row) => row.row_id).sort();
    assert.deepEqual(selected, expected);
  }
});

test("package inspection union shares its one producer and assurance presets have explicit private selections", () => {
  const compiler = new WorkGraphCompiler(root);
  bind(compiler, "package-inspect", "CHECK", { archive: "harness.package_archive", startup: "harness.package_startup", "reference-pack": "harness.package_reference_pack",
    recovery: "harness.package_recovery", all: "harness.package_all" }, { input: { default: "archive" } });
  const graph = compiler.compileCommand("package-inspect", { CHECK: "all" }).graph;
  assert.equal(graph.units.filter((unit) => unit.unit_id === "operation:harness.package_release").length, 1);
  const variants = Object.fromEntries(["default", "smoke", "execution", "regression", "lifecycle", "full", "runtime", "fixtures", "runtime_boundary", "artifact_consumers", "ui_review"]
    .map((name) => [name === "default" ? "contract" : name.replaceAll("_", "-"), `harness.contract_${name}`]));
  bind(compiler, "harness-contract", "SUITE", variants, { input: { default: "contract" } });
  for (const suite of Object.keys(variants)) assert.ok(compiler.compileCommand("harness-contract", { SUITE: suite }).graph.units.length > 0, suite);
  const full = compiler.compileCommand("harness-contract", { SUITE: "full" }).graph;
  assert.equal(full.units.filter((unit) => unit.current_run_evidence_outputs.includes("rows/harness.browser.unit.fixture_broker_smoke.json")).length, 1);
});

test("lint and source checks have closed selections and source validation cannot repair inputs", () => {
  const compiler = new WorkGraphCompiler(root);
  const variants = { standard: "harness.lint_standard", go: "harness.lint_go", types: "harness.frontend_typecheck",
    frontend: "harness.lint_biome", scripts: "harness.lint_scripts", shell: "harness.lint_shell",
    "frontend-boundary": "harness.frontend_import_boundary_check", "backend-boundary": "harness.backend_module_boundary_check", protocol: "harness.lint_protocol" };
  bind(compiler, "lint", "CHECK", variants, { input: { default: "standard" } });
  for (const check of Object.keys(variants)) assert.ok(compiler.compileCommand("lint", { CHECK: check }).graph.units.length > 0, check);
  const source = structuredClone(compiler.taskSurface.targets.find((entry) => entry.name === "lint"));
  source.name = "source-check";
  source.command_id = "cartulary.harness.command.source_check.v1";
  compiler.taskSurface.targets.push(source);
  const selections = Object.fromEntries(["generated", "schema", "catalog", "surface", "identity", "toolchain", "migration-input"]
    .map((name) => [name, `harness.source_${name.replaceAll("-", "_")}_selection`]));
  selections.all = "harness.source_all";
  bind(compiler, "source-check", "CHECK", selections, { input: { default: "all" } });
  for (const check of Object.keys(selections)) {
    const graph = compiler.compileCommand("source-check", { CHECK: check }).graph;
    assert.ok(graph.units.length > 0, check);
    assert.ok(graph.units.every((unit) => !unit.unit_id.startsWith("go:") && !unit.unit_id.startsWith("playwright:")));
    assert.ok(graph.units.every((unit) => !["operation:harness.generated_refresh", "operation:harness.generate_artifacts"].includes(unit.unit_id)));
  }
  assert.equal(compiler.owner.operations["harness.source_generated"].execution.env.CARTULARY_GENERATE_DRIFT_REFRESH, "0");
});

test("planning validates exact declared input files through the execution resolver", (t) => {
  const scratch = mkdtempSync(path.join(tmpdir(), "target-plan-inputs-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const file = path.join(scratch, "inputs.json");
  const compiler = new WorkGraphCompiler(root);
  bind(compiler, "build", "COMPONENT", { web: "harness.build_web", server: "harness.build_server" }, { input: { default: "server" } });
  writeFileSync(file, '{"COMPONENT":"web"}');
  const inputs = readPlanInputs(root, file);
  const plan = buildTargetPlan(root, "build", inputs, compiler);
  validateSchemaSync(plan.schema_id, plan);
  assert.equal(plan.graph_digest, compiler.compileCommand("build", inputs).graph.graph_digest);
  assert.deepEqual(plan.inputs, { COMPONENT: "web" });
  for (const detail of ["summary", "graph", "rows", "artifacts"]) assert.ok(renderTargetPlan(plan, detail));
  for (const invalid of ['{"COMPONENT":"web","COMPONENT":"server"}', '[]', 'null']) {
    writeFileSync(file, invalid);
    assert.throws(() => readPlanInputs(root, file));
  }
  symlinkSync(file, path.join(scratch, "linked.json"));
  assert.throws(() => readPlanInputs(root, path.join(scratch, "linked.json")));
  for (const invalid of [{ ARGS: "--unsafe" }, { COMPONENT: "unknown" }, { COMPONENT: 1 }]) {
    assert.throws(() => buildTargetPlan(root, "build", invalid, compiler));
  }
  assert.throws(() => buildTargetPlan(root, "bootstrap-tool-installations", {}, compiler), /unknown public target/u);
  assert.throws(() => buildTargetPlan(root, "absent", {}, compiler), /unknown public target/u);
  const guide = buildOwnerTaskGuide(root, "harness.command_surface");
  validateSchemaSync(guide.schema_id, guide);
  assert.equal(guide.rows.length, guide.coverage.row_count);
  assert.ok(guide.rows.some((row) => row.row_id.endsWith("scheduler_smoke") && row.minimum_tier === "fast"));
  assert.ok(guide.rows.some((row) => row.row_id.endsWith("native_runtime") && row.minimum_tier === "full"));
  assert.equal(Object.hasOwn(guide, "role"), false);
});

test("planning excludes documentation before filesystem access", () => {
  const methods = new Map(["readFileSync", "openSync", "statSync", "lstatSync"].map((name) => [name, fs[name]]));
  for (const name of methods.keys()) fs[name] = () => { throw new Error("unexpected filesystem access"); };
  syncBuiltinESMExports();
  try {
    const directory = "docs";
    for (const file of [path.join("/unavailable", directory, "plan.json"), "/unavailable/README.json", "/unavailable/plan.md"]) {
      assert.throws(() => readPlanInputs(root, file), /documentation|JSON file/u);
    }
  } finally {
    for (const [name, method] of methods) fs[name] = method;
    syncBuiltinESMExports();
  }
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
