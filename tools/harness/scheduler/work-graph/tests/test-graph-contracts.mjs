import assert from "node:assert/strict";
import path from "node:path";
import test from "node:test";
import { validateWorkGraph } from "../index.mjs";
import { createContractTestContext } from "../../../tests/contract-test-context.mjs";
import { assertCatalogTopologyValidation, assertCatalogTierSelection } from "./tier-selection.mjs";
const root = path.resolve(import.meta.dirname, "../../../../..");

test("catalog_topology_input_validation: catalog topology rejects retired raw selectors and malformed coverage inputs", async () => {
  const context = createContractTestContext(root);
  assertCatalogTopologyValidation(context);
});
test("catalog_independent_tier_selection: catalog tiers select before readiness without policy coupling", async () => {
  const context = createContractTestContext(root);
  assertCatalogTierSelection(context);
});
test("visual_semantic_partition: visual groups bound complete catalog claims without duplicate or missing execution", async () => {
  const context = createContractTestContext(root);
  const graph = context.compiler.compile({ kind: "target", target: "harness.suite_browser_visual" });
  const groups = graph.units.filter((unit) => unit.unit_id.startsWith("browser_group:visual:"));
  const rows = context.catalog.rows.filter((row) => row.runner === "playwright" && row.selector.stage === "visual");
  const selected = groups.flatMap((unit) => {
    const ids = unit.row_bindings.map((row) => row.row_id);
    assert.ok(ids.length > 0, "visual cohorts retain semantic row identities");
    const selectedRows = ids.map((id) => context.catalog.rowByID.get(id));
    if (ids.length > 1)
      assert.ok(selectedRows.every((row) => row.browser_state_scope === "test_owned"), "only explicitly test-owned visual rows may share a cohort");
    assert.equal(unit.owner_id, "harness.browser", "visual execution ownership is independent of reporting ownership");
    return ids;
  });
  assert.equal(new Set(selected).size, selected.length, "visual claims execute exactly once");
  assert.deepEqual(selected.sort(), rows.map((row) => row.row_id).sort(), "no active visual claim is omitted");
  for (const unit of graph.units) {
    assert.ok(Buffer.byteLength(`.tmp-${unit.unit_id}.json.2147483647`) <= 255, "private visual group/reset artifacts leave room for atomic-write names");
  }
  validateWorkGraph(graph);
});
test("measurement_semantic_identity: measurement groups preserve complete semantic identities independently of private storage", async () => {
  const context = createContractTestContext(root);
  const graph = context.compiler.compile({ kind: "target", target: "harness.suite_browser_measurement" });
  const rows = context.catalog.rows.filter((row) => row.selector?.file === "apps/web/e2e/measurement/timeline-grid.spec.ts");
  const ids = new Set(rows.map((row) => row.row_id));
  const groups = graph.units.filter((unit) => unit.unit_id.startsWith("browser_group:measurement:") &&
    unit.row_bindings.some((binding) => ids.has(binding.row_id)));
  assert.ok(rows.length > 0);
  assert.equal(groups.length, rows.length);
  assert.deepEqual(groups.map((unit) => unit.command.environment.CARTULARY_FIXTURE_ROW_ID).sort(), rows.map((row) => row.row_id).sort());
  for (const unit of groups) {
    assert.deepEqual(unit.row_bindings, [{ row_id: unit.command.environment.CARTULARY_FIXTURE_ROW_ID, runner: "playwright" }], "dedicated measurement retains its full catalog identity");
    assert.match(unit.unit_id, /^browser_group:measurement:cohort-[a-f0-9]{64}$/u, "bounded cohort identity binds the exact semantic selection");
  }
});
test("row_evidence_output_contract: row graphs declare exact current-run evidence outputs", async () => {
  const context = createContractTestContext(root);
  const row = context.catalog.rows.find((entry) => entry.runner === "go" && entry.fixture_capability === "none");
  const graph = context.compiler.compile({ kind: "rows", row_ids: [row.row_id] });
  assert.deepEqual(graph.units.flatMap((unit) => unit.current_run_evidence_outputs), [`rows/${row.row_id}.json`, `unit-results/go-${graph.units[0].unit_id.split(":").slice(1).join("-")}.json`]);
});

test("target_graph_validation: target graphs validate before execution", async () => {
  const context = createContractTestContext(root);
  const graph = context.compiler.compile({ kind: "target", target: "lint" });
  validateWorkGraph(graph);
  assert.ok(graph.units.length >= 7);
});
