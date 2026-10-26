import assert from "node:assert/strict";
import test from "node:test";
import { resolvePrivateSelection } from "../command-selection.mjs";

// Only the explicit inputs consumed by selection expansion; no compiler or I/O.
function selectionFixture() {
  const rows = [
    { row_id: "owner.unit", runner: "go", selector: {}, aggregate_placement: "fast" },
    { row_id: "owner.integration", runner: "go", selector: {}, aggregate_placement: "full" },
    { row_id: "owner.support", runner: "go", selector: {}, aggregate_placement: "full" },
    { row_id: "owner.visual", runner: "playwright", selector: { stage: "visual" }, aggregate_placement: "explicit_only" },
  ];
  return {
    owner: { operations: { "harness.prepare": {} }, target_members: {
      "harness.one": [{ row_id: "owner.unit" }, "harness.prepare"],
      "harness.union": ["harness.one", "harness.prepare", { row_id: "owner.unit" }],
    } },
    catalog: { rows, rowByID: new Map(rows.map(row => [row.row_id, row])) },
    rowTargets: new Map([["owner.unit", "backend-unit"], ["owner.integration", "backend-integration"], ["owner.support", "backend-integration-support"]]),
  };
}

test("private unions deduplicate exact rows and operations independently of reporting names", () => {
  const fixture = selectionFixture();
  assert.deepEqual(resolvePrivateSelection(fixture, "harness.union"), { rows: ["owner.unit"], operations: ["harness.prepare"] });
  fixture.owner.target_members["harness.renamed"] = fixture.owner.target_members["harness.union"];
  assert.deepEqual(resolvePrivateSelection(fixture, "harness.renamed"), resolvePrivateSelection(fixture, "harness.union"));
});

test("catalog expansion includes integration support and preserves explicit browser membership", () => {
  const fixture = selectionFixture();
  assert.deepEqual(resolvePrivateSelection(fixture, { catalog: { runner: "go", go_suite: "integration" } }),
    { rows: ["owner.integration", "owner.support"], operations: [] });
  assert.deepEqual(resolvePrivateSelection(fixture, { catalog: { runner: "playwright", stages: ["visual"] } }),
    { rows: ["owner.visual"], operations: [] });
});

test("invalid or cyclic private selections fail before any execution", () => {
  const fixture = selectionFixture();
  fixture.owner.target_members["harness.cycle"] = ["harness.cycle"];
  for (const invalid of ["harness.cycle", "harness.absent", "public-target", {}, { row_id: "absent" },
    { catalog: { runner: "playwright", stages: ["absent"] } }]) {
    assert.throws(() => resolvePrivateSelection(fixture, invalid), /cycle|unknown|canonical identity|no active rows/u);
  }
});
