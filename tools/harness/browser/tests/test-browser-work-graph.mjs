#!/usr/bin/env node

import assert from "node:assert/strict";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import {
  classifyFrontendVisualGoldens,
  resolveRegisteredFixtures,
} from "../frontend-visual-reconciliation.mjs";
import {
  collectFrontendMeasurementSummaries,
  measurementSchedulerOverlapCount,
} from "../frontend-measurement-evidence.mjs";
import { WorkGraphCompiler } from "../../scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const compiler = new WorkGraphCompiler(root);
const timelineMeasurementSource = readFileSync(
  path.join(root, "apps/web/e2e/measurement/timeline-grid.spec.ts"),
  "utf8",
);
assert.match(
  timelineMeasurementSource,
  /interactiveMeasurementSamplePolicy\.totalSamples/u,
  "AC-043 scenarios must consume the typed operation count",
);
assert.doesNotMatch(
  timelineMeasurementSource,
  /toBeLessThanOrEqual\((?:100|150)\)/u,
  "AC-043 scenarios must not substitute test-local thresholds",
);
assert.doesNotMatch(
  timelineMeasurementSource,
  /ordinaryMeasurementSamplePolicy|retry/iu,
  "AC-043 scenarios must not substitute a local sample policy or retry path",
);

function measurementSummary(overrides = {}) {
  return {
    schema_id: "cartulary.frontend_measurement_summary.v1",
    criterion_id: "AC-043",
    predicate_id: "perf.timeline_summary_selection_down.v1",
    fixture_id: "cartulary.perf.timeline_supported_envelope.v1",
    fixture_digest: `sha256:${"a".repeat(64)}`,
    measurement_policy_id: "cartulary.measurement.interactive_p95.v1",
    threshold_ms: 100,
    warmup_samples: 1,
    measured_samples: 100,
    percentile: 95,
    p50_ms: 10,
    p95_ms: 20,
    outcome: "passed",
    qualification: {
      quiet_profile_id: "browser_measurement_quiet",
      scheduler_overlap_count: 0,
      analyst_sessions: 25,
      background_updates_per_second: 4.8,
    },
    samples: Array.from({ length: 101 }, (_, sampleIndex) => ({
      sample_index: sampleIndex,
      warmup: sampleIndex === 0,
      total_ms: 10,
      stages_ms: { apply_to_visible_paint: 10 },
    })),
    ...overrides,
  };
}

function measurementReport(summary) {
  return {
    suites: [{
      specs: [{
        tests: [{
          results: [{
            attachments: [{
              name: `cartulary.frontend_measurement_summary.v1.${summary.predicate_id}`,
              body: Buffer.from(JSON.stringify(summary)).toString("base64"),
            }],
          }],
        }],
      }],
    }],
  };
}

const measurementEvidenceRoot = mkdtempSync(
  path.join(os.tmpdir(), "cartulary-measurement-evidence-"),
);
try {
  const reportPath = path.join(measurementEvidenceRoot, "playwright-report.json");
  const collect = (summary, expected = [summary.predicate_id]) => {
    writeFileSync(reportPath, `${JSON.stringify(measurementReport(summary))}\n`);
    return collectFrontendMeasurementSummaries({
      expectedPredicateIDs: expected,
      reportPaths: [reportPath],
      runRoot: measurementEvidenceRoot,
    });
  };
  assert.equal(collect(measurementSummary()).length, 1);
  assert.equal(
    collect(measurementSummary({ outcome: "threshold_failed", p95_ms: 101 }))[0].outcome,
    "threshold_failed",
    "threshold failures must remain valid diagnostic evidence",
  );
  assert.equal(
    collect(measurementSummary({
      failure_reason: "setup failed",
      measured_samples: 0,
      outcome: "incomplete",
      p50_ms: null,
      p95_ms: null,
      warmup_samples: 0,
      samples: [],
    }))[0].outcome,
    "incomplete",
    "setup failures must retain safe partial evidence",
  );
  assert.throws(
    () => collect(measurementSummary({
      measured_samples: 99,
      outcome: "incomplete",
      p50_ms: null,
      p95_ms: null,
    })),
    /sample counts differ/u,
    "summary cardinality must match its retained samples",
  );
  assert.throws(
    () => collect(measurementSummary({
      qualification: {
        quiet_profile_id: "browser_measurement_quiet",
        scheduler_overlap_count: 1,
        analyst_sessions: 25,
        background_updates_per_second: 4.8,
      },
    })),
    /environment_not_qualified/u,
  );
  assert.throws(
    () => collect(measurementSummary({
      samples: [{
        sample_index: 0,
        warmup: true,
        total_ms: 10,
        stages_ms: { record_id: 1 },
      }],
    })),
    /forbidden key record_id/u,
  );
  assert.throws(
    () => collect(measurementSummary(), []),
    /summaries differ/u,
    "missing expected evidence must fail closed",
  );
} finally {
  rmSync(measurementEvidenceRoot, { force: true, recursive: true });
}

const measurementGroupResult = {
  group_id: "measurement-timeline-grid",
  stage_id: "measurement",
};
const measurementUnitID = "browser_group:measurement:measurement-timeline-grid";
const quietEvents = [
  { event: "started", monotonic_ms: 100, seq: 1, unit_id: measurementUnitID },
  { event: "completed", monotonic_ms: 200, seq: 3, unit_id: measurementUnitID },
  {
    event: "started",
    monotonic_ms: 200,
    seq: 4,
    unit_id: "browser_target_summary:browser-e2e-measurement",
  },
];
assert.equal(
  measurementSchedulerOverlapCount(quietEvents, [measurementGroupResult]),
  0,
);
assert.equal(
  measurementSchedulerOverlapCount(
    [
      ...quietEvents,
      { event: "started", monotonic_ms: 150, seq: 2, unit_id: "row:ordinary" },
    ],
    [measurementGroupResult],
  ),
  1,
  "ordinary work beginning inside a measurement session must disqualify it",
);
assert.throws(
  () => measurementSchedulerOverlapCount(quietEvents.slice(0, 1), [measurementGroupResult]),
  /lacks a closed scheduler interval/u,
  "incomplete scheduler proof must fail closed",
);
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

const measurement = compiler.compile({
  kind: "target",
  target: "browser-e2e-measurement",
});
for (const unit of measurement.units) {
  if (Object.keys(unit.resource_claims).length === 0) continue;
  assert.equal(
    unit.shared_locks.includes("host_activity") +
      unit.exclusive_locks.includes("host_activity"),
    1,
    `${unit.unit_id} must declare exactly one host_activity mode`,
  );
}
for (const unit of measurement.units.filter((entry) =>
  entry.unit_id.startsWith("browser_lifecycle:measurement-") ||
  entry.unit_id.startsWith("browser_group:measurement:") ||
  entry.unit_id === "browser_target_summary:browser-e2e-measurement"
)) {
  assert.ok(
    unit.exclusive_locks.includes("host_activity"),
    `${unit.unit_id} must hold the quiet profile exclusively`,
  );
}
for (const unit of measurement.units.filter((entry) =>
  entry.unit_id.startsWith("browser_group:measurement:")
)) {
  assert.equal(
    unit.command.environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID,
    "browser_measurement_quiet",
  );
  assert.equal(unit.timeout_ms, 3_600_000);
  assert.equal(
    unit.failure_policy.block_descendants,
    false,
    `${unit.unit_id} must allow the evidence finalizer to run after failure`,
  );
}
const functional = compiler.compile({
  kind: "target",
  target: "browser-e2e-functional",
});
for (const unit of functional.units.filter((entry) =>
  entry.unit_id.startsWith("browser_group:")
)) {
  assert.ok(unit.shared_locks.includes("host_activity"));
  assert.equal(
    unit.command.environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID,
    "browser_isolated",
  );
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
  .filter((unit) => unit.unit_id.startsWith("browser_group:"))
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
