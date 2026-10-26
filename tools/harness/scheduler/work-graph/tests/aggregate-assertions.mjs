import assert from "node:assert/strict";

export function selectedRows(graph) {
  return [...new Set(graph.units.flatMap((unit) => unit.row_bindings.map((row) => row.row_id)))].sort();
}

export function assertBlockingPolicy(compiler, plans) {
  const check = plans.get("check");
  assert.equal(check.projections["frontend-fallow-static"], undefined);
  assert.equal(check.projections["go-gosec-audit"], undefined);
  for (const name of ["go-gosec-targeted", "go-vulncheck", "harness.json_shape_check", "otel-conformance"]) {
    assert.ok(check.projections[name]?.length > 0, `local blocking policy: ${name}`);
  }
  for (const [target, plan] of plans) {
    const uiReviewRows = selectedRows(plan.graph).filter((row) =>
      /^harness\.browser\.(?:boundary_support\.ui_review_|integration\.ui_review_|unit\.native_review$)/u.test(row));
    assert.deepEqual(uiReviewRows, [], `${target} excludes optional UI-review skill qualification`);
    assert.ok(plan.graph.units.every((unit) => !unit.unit_id.includes("harness_ui_review")),
      `${target} cannot introduce UI-review operations through policy dependencies`);
    if (target === "test-fast")
      assert.deepEqual(compiler.aggregatePolicyRoots(target), []);
    for (const name of ["frontend-fallow-static", "go-gosec-audit", "harness.deployable_shape", "build"]) {
      if (["ci", "release-check"].includes(target))
        assert.ok(plan.projections[name]?.length > 0, `${target}: ${name}`);
    }
  }
  for (const name of ["build-server", "build-server-harness", "build-migrate", "build-operator"]) {
    assert.ok(!compiler.owner.aggregate_policy_roots.check.includes(name), "check never roots production qualification");
    const id = `operation:${`harness.${name.replaceAll("-", "_")}`}`;
    if (check.graph.units.some((unit) => unit.unit_id === id)) {
      assert.ok(check.graph.units.some((unit) => unit.needs.includes(id)), `selected consumer requires ${name}`);
    }
  }
  const scanner = compiler.compileTarget("go-gosec-targeted");
  const policy = compiler.owner.policy_units["harness.go_gosec_audit"];
  const needs = policy.needs;
  try {
    policy.needs = ["invalid-advisory-dependency"];
    assert.deepEqual(compiler.compileTarget("go-gosec-targeted"), scanner, "blocking scanner never borrows advisory success or dependencies");
  } finally { policy.needs = needs; }
}

export function assertSmokeContexts(compiler, plans) {
  const smoke = compiler.compileTarget("harness.contract_smoke");
  const rows = ["harness.browser.unit.fixture_broker_smoke", "harness.command_surface.unit.scheduler_smoke"];
  assert.deepEqual(selectedRows(smoke), rows);
  assert.equal(smoke.units.filter((unit) => unit.unit_id.startsWith("go:") || unit.unit_id === "operation:harness.public_wrapper_smoke").length, 3);
  const check = plans.get("check");
  const ci = plans.get("ci");
  for (const graph of [check.graph, ci.graph]) {
    for (const unit of smoke.units)
      assert.deepEqual(graph.units.find((candidate) => candidate.unit_id === unit.unit_id), unit);
    for (const rowID of rows) {
      assert.equal(graph.units.filter((unit) => unit.current_run_evidence_outputs.includes(`rows/${rowID}.json`)).length, 1);
      const direct = compiler.compileRows([rowID]);
      const assertions = direct.units.filter((unit) => unit.current_run_evidence_outputs.includes(`rows/${rowID}.json`));
      assert.equal(assertions.length, 1);
      assert.equal(assertions[0].resource_claims.cpu, 1);
      assert.equal(assertions[0].cache_policy, "none");
      assert.deepEqual(graph.units.find((unit) => unit.unit_id === direct.units[0].unit_id), direct.units[0]);
    }
  }
  for (const rowID of ["harness.command_surface.unit.native_runtime", "harness.browser.unit.native_fixtures", "harness.evidence_accounting.behavior.output_accounting"]) {
    assert.equal(compiler.catalog.rowByID.get(rowID).aggregate_placement, "full");
    assert.ok(!selectedRows(check.graph).includes(rowID));
    assert.ok(selectedRows(ci.graph).includes(rowID));
  }
  const roles = compiler.taskSurface.harness_checks.filter((entry) => entry.gate_smoke_role).map((entry) => entry.name);
  assert.equal(roles.length, 3);
  for (const role of roles) {
    const binding = compiler.taskSurface.harness_checks.find((entry) => entry.name === role).selection;
    const expected = compiler.compilePrivateSelection(binding);
    assert.deepEqual(check.projections[role], expected.units.map((unit) => unit.unit_id).sort(), "one truthful smoke view with its exact dependencies");
  }
  for (const target of ["harness-runtime-contract", "harness-fixture-contract"]) {
    const operation = compiler.owner.operations[`harness.${target.replaceAll("-", "_")}`];
    assert.ok(operation.execution.command.includes("-race"));
    assert.ok(!smoke.units.some((unit) => unit.unit_id === `operation:harness.${target.replaceAll("-", "_")}`));
  }
}

export function assertAssuranceUnions(compiler, plans) {
  // Public assurance responsibilities are stable. Private entrypoint filenames
  // may move; duplicate registration remains forbidden.
  const common = ["harness-boundary-contract", "harness-browser-assurance-contract", "harness-command-surface-contract", "harness-evidence-contract", "harness-graph-contract", "harness-local-session-contract", "harness-scheduler-contract"];
  const ci = plans.get("ci");
  const target = "harness.contract_default";
  const operations = compiler.owner.target_members[target].map((member) => compiler.owner.operations[typeof member === "string" ? member : member.operation_id]);
  const files = operations.flatMap((operation) => operation.execution.command).filter((arg) => /\.(?:mjs|sh)$/u.test(arg)).map((arg) => arg.replace(/^\.\//u, ""));
  assert.deepEqual(operations.map((operation) => operation.report_key).sort(), [...common, "harness-observation-contract", "harness-dev-services-lifecycle", "harness-smoke-frontend-import-boundaries"].sort());
  assert.equal(new Set(files).size, files.length, "no entrypoint is registered twice within the assurance union");
  const direct = compiler.compileTarget(target);
  const row = "harness.test_catalog.behavior.routing_contract";
  const canonical = compiler.compileRows([row]).units.find((unit) => unit.row_bindings.some((binding) => binding.row_id === row));
  for (const preset of ["harness.contract_execution", "harness.contract_regression", "harness.contract_full"]) {
    const selected = compiler.compileTarget(preset);
    assert.deepEqual(selected.units.filter((unit) => unit.row_bindings.some((binding) => binding.row_id === row)), [canonical], `${preset}: one canonical graph assurance unit`);
  }
  assert.deepEqual(ci.projections["harness-contract"], direct.units.map((unit) => unit.unit_id).sort(), "CI retains the complete default assurance view");
  const scannerRow = "harness.command_surface.behavior.go_scanner_wrappers";
  const scanner = compiler.compileRows([scannerRow]).units.find((unit) => unit.row_bindings.some((binding) => binding.row_id === scannerRow));
  assert.ok(!selectedRows(direct).includes(scannerRow), "default assurance remains bounded");
  for (const preset of ["harness.contract_regression", "harness.contract_full"]) {
    assert.deepEqual(compiler.compileTarget(preset).units.filter((unit) => unit.row_bindings.some((binding) => binding.row_id === scannerRow)), [scanner]);
  }
  assert.deepEqual(ci.graph.units.filter((unit) => unit.row_bindings.some((binding) => binding.row_id === scannerRow)), [scanner]);
  const outputCases = compiler.compilePrivateSelection("harness.assurance_tool_output_real_targets");
  const outputUnits = outputCases.units.filter((unit) => unit.unit_id.startsWith("operation:harness.assurance_output_"));
  const caseArguments = outputUnits.map((unit) => compiler.owner.operations[unit.unit_id.slice("operation:".length)].execution.command.at(-1));
  assert.deepEqual(caseArguments.sort(), ["backend-unit", "backend-unit-machine", "build", "build-machine", "check", "check-machine", "invalid-usage", "lint", "owner-json", "source-schema", "test-fast"]);
  assert.equal(new Set(outputUnits.map((unit) => unit.semantic_digest)).size, outputUnits.length, "each distinct output assertion is one separately bounded execution");
  const artifacts = compiler.compilePrivateSelection("harness.contract_artifact_consumers");
  assert.deepEqual(selectedRows(artifacts), ["harness.browser.boundary_support.frontend_artifact_lifetime", "package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b"]);
  assert.deepEqual(artifacts, compiler.compileRows(selectedRows(artifacts)), "artifact preset preserves exact row execution");
  for (const row of selectedRows(artifacts))
    assert.equal(ci.graph.units.filter((unit) => unit.current_run_evidence_outputs.includes(`rows/${row}.json`)).length, 1);
  for (const unit of direct.units)
    assert.deepEqual(ci.graph.units.filter((entry) => entry.unit_id === unit.unit_id), [unit], "one canonical definition across direct and CI selection");
  const check = plans.get("check");
  assert.equal(check.projections["harness-contract"], undefined, "partial selection cannot claim the full assurance union");
  assert.equal(compiler.taskSurface.make_recipes["test"], undefined, "the removed aggregate cannot become another policy");
}

export function assertAggregateResources(plans) {
  for (const graphTarget of ["test-fast", "check", "ci", "release-check"]) {
    const graph = plans.get(graphTarget).graph;
    assert.equal(
      graph.units.some((entry) =>
        ["test-fast", "check", "ci", "release-check"].includes(
          entry.command.args.at(-1),
        )),
      false,
      `${graphTarget} must not invoke a nested aggregate`,
    );
  }
  const releaseGraph = plans.get("release-check").graph;
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
