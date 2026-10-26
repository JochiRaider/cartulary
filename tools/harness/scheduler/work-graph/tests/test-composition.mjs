import assert from "node:assert/strict";
import test from "node:test";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import { validateStaticConfiguration } from "../../../static-analysis/index.mjs";
import { WorkGraphCompiler } from "../index.mjs";
import { buildOperationInvocation, operationExecutionDigest, projectOperationBindings } from "../operations.mjs";
import { policyDependencyRowIDs } from "../policy-rows.mjs";
import { commandInputContext, selectionFor } from "../preparation.mjs";
import { buildWorkGraph, unitSemanticDigest } from "../model.mjs";
import { selectedRows } from "./aggregate-assertions.mjs";

test("each sized invocation pays its lifecycle once without discounting serialized selected work", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const ids = [
    "module.revisions.store.history_query_repository_mapping_7b4d2e9c11",
    "module.revisions.store.revision_window_utc_ordering_51f2a6b3c5",
  ];
  const runners = () => compiler.compileRows(ids).units.filter((unit) => unit.kind === "runner");
  for (const lanes of [1, 4]) {
    compiler.availableGoLanes = lanes;
    const before = runners();
    assert.deepEqual(before.flatMap((unit) => unit.row_bindings.map((row) => row.row_id)).sort(), [...ids].sort());
    if (lanes === 1)
      assert.equal(before.length, 1);
    for (const unit of before) {
      const selected = unit.row_bindings.reduce((total, { row_id }) => {
        const row = compiler.catalog.rowByID.get(row_id);
        return total + compiler.owner.evidence_estimates_ms[row.evidence_class];
      }, 0);
      assert.equal(unit.estimated_work_ms, compiler.owner.invocation_policies.go.startup_ms +
        compiler.owner.invocation_policies.go.cleanup_ms + selected);
    }
    compiler.owner.invocation_policies.go.startup_ms += 317;
    const after = runners();
    assert.equal(after.length, before.length);
    for (const unit of before) {
      const changed = after.find((candidate) => JSON.stringify(candidate.row_bindings) === JSON.stringify(unit.row_bindings));
      assert.ok(changed);
      assert.equal(changed.estimated_work_ms - unit.estimated_work_ms, 317);
      assert.notEqual(changed.semantic_digest, unit.semantic_digest);
      assert.equal(changed.timeout_ms, unit.timeout_ms);
      assert.deepEqual(changed.resource_claims, unit.resource_claims);
    }
    assert.deepEqual(compiler.compileRows(ids), compiler.compileRows([...ids].reverse()));
  }
});
test("installed browser lifecycle parents reserve their dedicated footprint without unused outer services", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const topology = JSON.parse(fs.readFileSync(new URL("../../../../execution_topology_manifest.json", import.meta.url)));
  const physical = Object.values(topology.physical_resource_claims);
  for (const row of ["harness.browser.integration.native_fixture_lifecycle", "harness.browser.integration.ui_review_seeded_default", "harness.browser.integration.ui_review_seeded_network_flow_claimed"]) {
    const graph = compiler.compileRows([row]);
    const unit = graph.units.find((entry) => entry.row_bindings.some((binding) => binding.row_id === row));
    assert.deepEqual(unit.service_dependencies, []);
    assert.deepEqual(unit.physical_allocations, {});
    for (const resource of ["port_lane", "process", "memory_mb", "browser_stack", "object_store", "postgres"]) {
      const resident = physical.reduce((sum, claims) => sum + (claims[resource] ?? 0), 0);
      const activity = resource === "memory_mb" ? 512 : resource === "process" ? 1 : 0;
      assert.ok((unit.resource_claims[resource] ?? 0) >= resident + activity, `${row}: missing nested ${resource} entitlement`);
    }
  }
});
test("installed fixture consumers share graph preparation and cannot start before it settles", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const owned = "harness.browser.integration.native_fixture_lifecycle";
  const borrowed = "harness.browser.integration.native_borrowed_fixture_lifecycle";
  const graph = compiler.compileRows([owned, borrowed]);
  const consumer = (row) => graph.units.find((unit) => unit.row_bindings.some((binding) => binding.row_id === row));
  const prerequisites = (unit) => {
    const found = new Set();
    const visit = (id) => {
      if (found.has(id))
        return;
      found.add(id);
      for (const next of graph.units.find((entry) => entry.unit_id === id).needs)
        visit(next);
    };
    for (const id of unit.needs)
      visit(id);
    return found;
  };
  for (const operation of ["build_web", "embedded_web_assets", "build_server_harness", "build_migrate", "testservices_build", "test_service_images", "playwright_install"]) {
    const id = `operation:harness.${operation}`;
    assert.equal(graph.units.filter((unit) => unit.unit_id === id).length, 1);
    assert.ok(prerequisites(consumer(owned)).has(id), `${operation} must settle before owned acquisition`);
  }
  assert.ok(prerequisites(consumer(borrowed)).has("operation:harness.test_service_images"));
  assert.equal(prerequisites(consumer(borrowed)).has("operation:harness.build_server_harness"), false);
  for (const row of [owned, borrowed]) {
    assert.deepEqual(consumer(row).service_dependencies, []);
    assert.deepEqual(consumer(row).physical_allocations, {});
  }
});
test("private input forwarding follows typed metadata including future selectors", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const surface = structuredClone(compiler.taskSurface);
  const entry = surface.targets.find((target) => target.name === "harness-contract");
  const suite = entry.input_contract.inputs.find((input) => input.name === "SUITE");
  entry.input_contract.inputs.push({ ...suite, name: "FUTURE_SELECTOR", values: ["selected"], default: null });
  entry.input_contract.inputs.push({ ...suite, name: "RUNTIME_VALUE", values: ["admitted"], default: null,
    allowed_sources: ["environment"], child_forwarding: "runtime_env" });
  const flag = surface.targets.find((target) => target.name === "test-slice").input_contract.inputs.find((input) => input.name === "JSON");
  entry.input_contract.inputs.push({ ...flag, name: "RUNTIME_FLAG", child_forwarding: "runtime_env" });
  surface.targets.find((target) => target.name === "lint").input_contract.inputs.push({ ...suite,
    name: "FOREIGN_SELECTOR", values: ["foreign"], default: null });
  const environment = { SUITE: "runtime", FUTURE_SELECTOR: "selected", FOREIGN_SELECTOR: "foreign",
    RUNTIME_VALUE: " admitted ", RUNTIME_FLAG: "1", VITEST_MAX_WORKERS: "99",
    CARTULARY_MAKE_INPUT_SOURCES: "SUITE=cli FUTURE_SELECTOR=cli RUNTIME_VALUE=env RUNTIME_FLAG=env FOREIGN_SELECTOR=env" };
  const before = { ...environment };
  const context = commandInputContext(entry, {}, { environment, surface });
  assert.equal(context.inputs.SUITE, "runtime");
  assert.equal(context.inputs.FUTURE_SELECTOR, "selected");
  assert.equal(context.input_environment.forward.RUNTIME_VALUE, "admitted");
  assert.equal(context.input_environment.forward.RUNTIME_FLAG, "1");
  assert.equal(commandInputContext(entry, {}, { environment: { ...environment, RUNTIME_FLAG: undefined }, surface }).input_environment.forward.RUNTIME_FLAG, "");
  for (const name of ["SUITE", "FUTURE_SELECTOR", "FOREIGN_SELECTOR", "VITEST_MAX_WORKERS", "CARTULARY_MAKE_INPUT_SOURCES"]) {
    assert.ok(context.input_environment.clear.includes(name));
    assert.equal(Object.hasOwn(context.input_environment.forward, name), false, `${name} cannot leak into private work`);
  }
  assert.deepEqual(environment, before, "preparation does not mutate caller inputs");
  assert.throws(() => commandInputContext(entry, {}, { environment: { ...environment, FUTURE_SELECTOR: "unknown" }, surface }), /FUTURE_SELECTOR/u);
  entry.input_contract.inputs.find((input) => input.name === "RUNTIME_VALUE").summary_emission = "redacted_value";
  assert.throws(() => commandInputContext(entry, {}, { environment, surface }), /owned private runtime binding/u);
});
test("row bindings close exact outputs and affect physical execution identity", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const row = compiler.catalog.rows.find((row) => row.row_id === "harness.command_surface.unit.scheduler_smoke");
  const graph = compiler.compileRows([row.row_id]);
  const unit = graph.units.find((unit) => unit.row_bindings.some((binding) => binding.row_id === row.row_id));
  assert.deepEqual(unit.row_bindings, [{ row_id: row.row_id, runner: row.runner }]);
  const changed = structuredClone(unit);
  changed.row_bindings[0].runner = "shell";
  assert.notEqual(unitSemanticDigest(changed), unit.semantic_digest);
  assert.throws(() => buildWorkGraph([{ ...unit, row_bindings: [] }]), /row bindings/u);
  assert.throws(() => buildWorkGraph([{ ...unit, row_bindings: [...unit.row_bindings, ...unit.row_bindings] }]), /duplicate/u);
  assert.throws(() => buildWorkGraph([...graph.units, { ...unit, unit_id: "duplicate:row-owner" }]), /multiple graph owners/u);
});
test("operation deadlines are explicit, validated and part of execution identity", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const operationID = "harness.assurance_output_check";
  const policy = compiler.owner.policy_units[operationID];
  const compile = () => compiler.compileOperation(operationID).units[0];
  const declared = compile();
  assert.equal(declared.timeout_ms, 1800000);
  assert.equal(compiler.compileOperation("harness.assurance_output_check_machine").units[0].timeout_ms, declared.timeout_ms);
  delete policy.timeout_ms;
  const ordinary = compile();
  assert.equal(ordinary.timeout_ms, compiler.owner.default_timeout_ms);
  assert.notEqual(ordinary.semantic_digest, declared.semantic_digest);
  assert.equal(ordinary.unit_id, declared.unit_id);
  assert.deepEqual(ordinary.command, declared.command);
  for (const timeout of [null, 0, -1, 1.5, "1800000", 3600001]) {
    policy.timeout_ms = timeout;
    assert.throws(compile, /timeout_ms/u, `invalid operation deadline: ${timeout}`);
  }
});
test("blocking Fallow configuration validation stays read-only and rejects missing declared inputs", () => {
  const root = new URL("../../../../../", import.meta.url).pathname;
  const read = fs.readFileSync;
  const write = fs.writeFileSync;
  let invalid = false;
  fs.writeFileSync = () => { throw new Error("configuration validation attempted a write"); };
  fs.readFileSync = (file, ...args) => {
    const result = read(file, ...args);
    if (!invalid || String(file) !== `${root}tools/fallow/reachability_owner.json`)
      return result;
    const owner = JSON.parse(String(result));
    owner.executable_tooling_dependencies[0].owner_script = "tools/absent-owner-script.mjs";
    return JSON.stringify(owner);
  };
  syncBuiltinESMExports();
  try {
    validateStaticConfiguration(root);
    invalid = true;
    assert.throws(() => validateStaticConfiguration(root), /missing file tools\/absent-owner-script/u, "prior success cannot conceal a missing executable dependency declaration");
  }
  finally {
    fs.readFileSync = read;
    fs.writeFileSync = write;
    syncBuiltinESMExports();
  }
});


test("aggregate browser selection preserves tier rows without stage expansion or excluded preparation", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const measurement = compiler.catalog.rows.filter((row) => row.runner === "playwright" && compiler.rowTargets.get(row.row_id) === "browser-e2e-measurement");
  const full = measurement.filter((row) => row.aggregate_placement === "full").map((row) => row.row_id).sort();
  const release = measurement.filter((row) => row.aggregate_placement === "release").map((row) => row.row_id).sort();
  assert.ok(full.length > 0);
  assert.ok(release.length > 0);
  assert.deepEqual(selectedRows(compiler.compileTarget("harness.suite_browser_measurement")), [...full, ...release].sort());
  for (const rowID of release)
    compiler.catalog.rowByID.get(rowID).preparation_operations = ["forbidden-excluded-producer"];
  const graph = compiler.compileAggregate("ci");
  assert.deepEqual(selectedRows(graph).filter((id) => measurement.some((row) => row.row_id === id)), full);
  for (const unit of graph.units) {
    const filter = (unit.row_bindings ?? []).filter((row) => row.runner === "playwright");
    assert.ok(filter.every(({ row_id }) => !release.includes(row_id)), "runner receives the exact admitted rows");
  }
  assert.throws(() => compiler.compileRows([release[0]]), /unknown harness operation forbidden-excluded-producer/u, "negative preparation fixture reaches its boundary when explicitly selected");
});
test("conformance dependencies preserve required security rows without following owner growth", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const ids = policyDependencyRowIDs(compiler.owner, "harness.otel_conformance", compiler.catalog.rows);
  assert.equal(ids.length, 11);
  assert.throws(() => policyDependencyRowIDs(compiler.owner, "otel-conformance", compiler.catalog.rows), /unknown policy operation/u);
  assert.equal(ids.filter((id) => compiler.catalog.rowByID.get(id).aggregate_placement === "release").length, 7);
  const before = compiler.compileTarget("otel-conformance");
  assert.deepEqual(selectedRows(before), ids);
  const source = compiler.catalog.rowByID.get(ids[0]);
  const added = { ...source, row_id: "platform.telemetry.unit.unrelated_future_assertion" };
  compiler.catalog.rows.push(added);
  compiler.catalog.rowByID.set(added.row_id, added);
  compiler.rowTargets.set(added.row_id, compiler.rowTargets.get(source.row_id));
  assert.deepEqual(compiler.compileTarget("otel-conformance"), before);
  source.status = "inactive";
  assert.throws(() => compiler.compileTarget("otel-conformance"), /missing active required row/u);
  source.status = "active";
  compiler.owner.policy_units["harness.otel_conformance"].required_rows.push({ row_id: ids[0], rationale: "duplicate fixture" });
  assert.throws(() => compiler.compileTarget("otel-conformance"), /duplicate required row/u);
});
test("service scope rejects every invalid requested row before compiling preparation", () => {
  const rows = [
    { row_id: "owner.service", owner_id: "owner", service_dependencies: ["postgres"] },
    { row_id: "owner.unit", owner_id: "owner", service_dependencies: [] },
    { row_id: "other.service", owner_id: "other", service_dependencies: ["postgres"] },
  ];
  const compiler = { catalog: { rows }, compile() { throw new Error("preparation reached"); } };
  const options = { selection: "owner", target: "test-slice", owner: "owner", serviceBackedOnly: true };
  for (const selected of [["owner.unit"], ["owner.service", "owner.unit"], ["unknown.row"], ["other.service"], ["owner.service", "owner.service"]]) {
    assert.throws(() => selectionFor({ ...options, rows: selected }, compiler), /does not belong|outside service-backed scope|duplicate/u);
  }
});
test("command bindings preserve operation identity and executable changes invalidate it", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const original = compiler.compileCommand("build", { COMPONENT: "migrate" }).graph;
  const definition = compiler.taskSurface.targets.find((entry) => entry.name === "build");
  const renamed = structuredClone(definition);
  renamed.name = "renamed-command";
  renamed.command_id = "cartulary.harness.command.renamed.v1";
  compiler.taskSurface.targets.push(renamed);
  assert.deepEqual(compiler.compileCommand("renamed-command", { COMPONENT: "migrate" }).graph, original, "public spelling does not change physical work");
  renamed.selection_binding.variants.migrate = "harness.unknown";
  assert.throws(() => compiler.compileCommand("renamed-command", { COMPONENT: "migrate" }), /unknown private selection/u);
  const operation = compiler.owner.operations["harness.build_migrate"];
  const digest = operationExecutionDigest(operation);
  operation.execution.prerequisites = ["$(OTHER_BINARY)"];
  assert.notEqual(operationExecutionDigest(operation), digest, "changing the executable producer changes identity");
  const publication = compiler.owner.operations["harness.json_shape_check"];
  const publicationDigest = operationExecutionDigest(publication);
  const publicationGraph = compiler.compileOperation("harness.json_shape_check");
  publication.execution.success_summary = false;
  assert.notEqual(operationExecutionDigest(publication), publicationDigest, "removing the evidence writer changes operation identity");
  assert.notDeepEqual(compiler.compileOperation("harness.json_shape_check"), publicationGraph, "evidence publication changes the compiled unit identity");
  publication.execution.success_summary = true;
  publication.execution.step_label = "another-step";
  assert.notEqual(operationExecutionDigest(publication), publicationDigest, "the run-step argument is executable input");
  const producer = structuredClone(operation);
  const producerDigest = operationExecutionDigest(producer);
  delete producer.execution.test_target;
  assert.notEqual(operationExecutionDigest(producer), producerDigest, "artifact summary emission changes execution identity");
  producer.report_key = "another-report-view";
  const reportDigest = operationExecutionDigest(producer);
  producer.report_key = "yet-another-report-view";
  assert.equal(operationExecutionDigest(producer), reportDigest, "the separately bound report key is not an executable definition");
  assert.throws(() => projectOperationBindings({ make_recipes: { invalid: { type: "operation", operation_id: "harness.build_migrate", command: ["override"] } } }, compiler.owner), /cannot override executable/u);
});
test("operation argument transport never evaluates shell or Make expressions", () => {
  const owner = { operations: { "harness.fixture": { report_key: "fixture", execution: {
          type: "step_command", mode: "command", prerequisites: [], command: ["tool", "$(INPUT)"]
        },
      } } };
  const literal = "$(touch forbidden); `echo unexpected` with spaces";
  const invocation = buildOperationInvocation(owner, "harness.fixture", { INPUT: literal });
  assert.deepEqual(invocation.args, [literal]);
  assert.throws(() => buildOperationInvocation(owner, "harness.fixture", {}), /missing operation binding/u);
  owner.operations["harness.fixture"].execution.command = ["make", "retired-wrapper"];
  assert.throws(() => buildOperationInvocation(owner, "harness.fixture", {}), /Make may produce files only/u);
  owner.operations["harness.fixture"].execution = { type: "artifact_binding", prerequisites: ["retired-wrapper"] };
  assert.throws(() => buildOperationInvocation(owner, "harness.fixture", {}), /absolute file targets/u);
});
test("vulnerability operations preserve an explicit database and allow the aggregate default", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../../../", import.meta.url).pathname);
  const environment = { GO: "go", GO_CACHE_DIR: "/cache/build", GO_MOD_CACHE_DIR: "/cache/mod", GO_TMP_DIR: "/cache/tmp",
    GOVULNCHECK_BIN: "/tools/govulncheck", RUN_STEP_SCRIPT: "/tools/run-step.sh" };
  const defaults = buildOperationInvocation(compiler.owner, "harness.go_vulncheck", environment);
  assert.equal(Object.hasOwn(defaults.env, "GOVULNCHECK_DB"), false, "the scanner owns the omitted database default");
  const selected = buildOperationInvocation(compiler.owner, "harness.go_vulncheck", { ...environment, GOVULNCHECK_DB: "file:///database" });
  assert.equal(selected.env.GOVULNCHECK_DB, "file:///database");
  assert.deepEqual(defaults.args, selected.args);
});
