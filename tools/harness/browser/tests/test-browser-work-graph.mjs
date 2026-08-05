#!/usr/bin/env node

import assert from "node:assert/strict";
import path from "node:path";

import {
  classifyFrontendVisualGoldens,
  resolveRegisteredFixtures,
} from "../frontend-visual-reconciliation.mjs";
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
const visualGraph = compiler.compile({
  kind: "target",
  target: "browser-e2e-visual",
});
const visualTargetFinalizer = visualGraph.units.find(
  (unit) => unit.unit_id === "browser_target_summary:browser-e2e-visual",
);
assert.ok(visualTargetFinalizer, "visual target must expose its finalizer");
assert.ok(
  visualTargetFinalizer.evidence_outputs.includes(
    "browser-e2e-visual/frontend-visual-reconciliation.json",
  ),
  "visual target finalizer must retain the reconciliation artifact",
);
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

const capture = (captureID, goldenPath) => ({
  capture_id: captureID,
  capture_intent: captureID,
  owner_id: "harness.browser",
  row_id: `row.${captureID}`,
  scenario_id: `scenario.${captureID}`,
  project_id: "chromium",
  expected_golden_path: goldenPath,
});
const classifiedGoldens = classifyFrontendVisualGoldens({
  captureIntents: [
    capture("active", "snapshots/active.png"),
    capture("missing", "snapshots/missing.png"),
    capture("ambiguous-a", "snapshots/ambiguous.png"),
    capture("ambiguous-b", "snapshots/ambiguous.png"),
  ],
  committedGoldens: new Map([
    ["snapshots/active.png", "a".repeat(64)],
    ["snapshots/ambiguous.png", "b".repeat(64)],
    ["snapshots/orphan.png", "c".repeat(64)],
  ]),
  fixtures: [
    {
      fixture_id: "visual.fixture.active",
      golden_artifacts: ["snapshots/active.png"],
    },
  ],
});
assert.deepEqual(
  Object.fromEntries(
    classifiedGoldens.map((golden) => [
      golden.golden_path,
      golden.classification,
    ]),
  ),
  {
    "snapshots/active.png": "active",
    "snapshots/ambiguous.png": "ambiguous_mapping",
    "snapshots/missing.png": "missing_golden",
    "snapshots/orphan.png": "orphan",
  },
  "visual reconciliation must distinguish active, ambiguous, missing, and orphan goldens",
);

const fixtureCatalogEntries = [
  {
    row_id: "row.selected",
    title: "Selected fixture",
  },
  {
    row_id: "row.unselected",
    title: "Unselected fixture",
  },
];
const fixtureGoldens = [
  {
    golden_path: "snapshots/selected.png",
    sha256: "d".repeat(64),
    catalog_row_ids: ["row.selected"],
  },
  {
    golden_path: "snapshots/unselected.png",
    sha256: "e".repeat(64),
    catalog_row_ids: [],
  },
];
assert.deepEqual(
  resolveRegisteredFixtures(
    [
      {
        fixture_id: "visual.fixture.selected",
        catalog_row_ids: ["row.selected"],
        playwright_scenario_title: "Selected fixture",
        golden_artifacts: ["snapshots/selected.png"],
      },
      {
        fixture_id: "visual.fixture.unselected",
        catalog_row_ids: ["row.unselected"],
        playwright_scenario_title: "Unselected fixture",
        golden_artifacts: ["snapshots/unselected.png"],
      },
    ],
    fixtureGoldens,
    fixtureCatalogEntries,
    [
      {
        row_id: "row.selected",
      },
    ],
  ),
  [],
  "an owner slice must validate selected fixtures from capture intent and unselected fixtures from exact catalog/path existence",
);
assert.deepEqual(
  resolveRegisteredFixtures(
    [
      {
        fixture_id: "visual.fixture.selected",
        catalog_row_ids: ["row.selected"],
        playwright_scenario_title: "Selected fixture",
        golden_artifacts: ["snapshots/unselected.png"],
      },
    ],
    fixtureGoldens,
    fixtureCatalogEntries,
    [
      {
        row_id: "row.selected",
      },
    ],
  ).map((fixture) => fixture.fixture_id),
  ["visual.fixture.selected"],
  "a selected fixture must resolve its exact runtime catalog row",
);
