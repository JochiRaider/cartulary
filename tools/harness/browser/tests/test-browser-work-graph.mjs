#!/usr/bin/env node

import assert from "node:assert/strict";
import path from "node:path";

import { WorkGraphCompiler } from "../../scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const compiler = new WorkGraphCompiler(root);
for (const target of [
  "browser-e2e-functional",
  "browser-e2e-stateful",
  "browser-e2e-measurement",
  "browser-e2e-a11y",
  "browser-e2e-visual",
]) {
  const graph = compiler.compile({ kind: "target", target });
  assert.ok(graph.units.some((unit) => unit.unit_id.startsWith("browser_group:")), `${target} must expose browser groups as units`);
  assert.ok(graph.units.some((unit) => unit.unit_id.startsWith("browser_target_summary:")), `${target} must expose its summary projection as a unit`);
  assert.equal(graph.units.some((unit) => unit.command.args.some((arg) => arg.includes("run-browser-e2e-batch"))), false);
}
const stateful = compiler.compile({ kind: "target", target: "browser-e2e-stateful" });
assert.ok(stateful.units.some((unit) => unit.unit_id.startsWith("browser_reset:")), "stateful browser work must expose resets");
for (const unit of stateful.units.filter((entry) => entry.fixture_lease === "browser_stack")) {
  assert.equal(unit.resource_claims.postgres, 4, `${unit.unit_id} must reserve a safe browser Postgres connection budget`);
}
const byAffinity = Map.groupBy(
  stateful.units.filter((unit) => unit.fixture_lease === "browser_stack"),
  (unit) => unit.affinity_key,
);
for (const [affinity, units] of byAffinity) {
  const terminal = units.filter((unit) =>
    unit.command.environment.CARTULARY_BROWSER_RELEASE_AFFINITY === "1",
  );
  assert.equal(terminal.length, 1, `${affinity} must have one terminal stack releaser`);
}

const workbookOwner = compiler.compile({
  kind: "owner",
  owner_id: "module.workbook",
});
const selectedWorkbookBrowserRows = compiler.catalog.rows
  .filter(
    (row) => row.owner_id === "module.workbook" && row.runner === "playwright",
  )
  .map((row) => row.row_id)
  .sort();
const projectedWorkbookBrowserRows = workbookOwner.units
  .flatMap((unit) =>
    unit.evidence_outputs
      .filter((output) => output.startsWith("rows/"))
      .map((output) => output.slice("rows/".length, -".json".length)),
  )
  .sort();
assert.deepEqual(
  projectedWorkbookBrowserRows,
  selectedWorkbookBrowserRows,
  "owner selections must project Playwright rows through browser group units",
);
assert.equal(
  workbookOwner.units.some(
    (unit) =>
      unit.unit_id.startsWith("row:") &&
      selectedWorkbookBrowserRows.includes(unit.unit_id.slice("row:".length)),
  ),
  false,
  "owner selections must not schedule Playwright row runners directly",
);
