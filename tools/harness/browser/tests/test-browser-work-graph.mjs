#!/usr/bin/env node

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import {
  closeSync,
  mkdtempSync,
  openSync,
  readFileSync,
  rmSync,
  writeFileSync,
  writeSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";

import {
  classifyFrontendVisualGoldens,
  resolveRegisteredFixtures,
} from "../frontend-visual-reconciliation.mjs";
import {
  collectFinalizedMeasurementSummaries,
  collectFrontendMeasurementObservations,
  readMeasurementSchedulerEvidence,
} from "../frontend-measurement-evidence.mjs";
import {
  activePerformanceFixtureProfile,
  loadPerformanceFixtureSnapshotRegistry,
} from "../../performance-fixture/index.mjs";
import { WorkGraphCompiler } from "../../scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const compiler = new WorkGraphCompiler(root);
const fixtureProfile = activePerformanceFixtureProfile(
  loadPerformanceFixtureSnapshotRegistry(root),
  "ac043_large_grid_snapshot_v1",
);
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

function measurementObservation(overrides = {}) {
  return {
    schema_id: "cartulary.frontend_measurement_observation.v2",
    fixture_profile_id: "ac043_large_grid_snapshot_v1",
    criterion_id: "AC-043",
    predicate_id: "perf.timeline_summary_selection_down.v1",
    fixture_id: "cartulary.perf.large_grid.v1",
    fixture_digest: `sha256:${"a".repeat(64)}`,
    measurement_policy_id: "cartulary.measurement.interactive_p95.v1",
    threshold_ms: 100,
    warmup_samples: 1,
    measured_samples: 100,
    percentile: 95,
    p50_ms: 10,
    p95_ms: 20,
    outcome: "passed",
    traffic: {
      counts: [
        { fact_id: "analyst_sessions", value: 25 },
        { fact_id: "background_sessions", value: 24 },
      ],
      rates: [{ fact_id: "background_updates_per_second", value: 4.8 }],
      conditions: [
        { fact_id: "presence_enabled", value: true },
        { fact_id: "target_row_excluded", value: true },
      ],
    },
    samples: Array.from({ length: 101 }, (_, sampleIndex) => ({
      sample_index: sampleIndex,
      warmup: sampleIndex === 0,
      total_ms: 10,
      stages_ms: [{ fact_id: "apply_to_visible_paint", value: 10 }],
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
              name: `cartulary.frontend_measurement_observation.v2.${summary.predicate_id}`,
              body: Buffer.from(JSON.stringify(summary)).toString("base64"),
            }],
          }],
        }],
      }],
    }],
  };
}

function artifactDigest(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function writeMeasurementArtifact(rootDirectory, name, role, value) {
  const file = path.join(rootDirectory, name);
  const bytes = Buffer.from(`${JSON.stringify(value)}\n`);
  writeFileSync(file, bytes);
  return {
    ref: {
      role,
      path_kind: "file",
      format: "json",
      path: name,
      sha256: artifactDigest(bytes),
    },
    value,
  };
}

function finalizedMeasurementSummary(rootDirectory, options = {}) {
  const observation = measurementObservation(options.observationOverrides);
  const rowID =
    "module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3";
  const snapshotKey = "b".repeat(64);
  const builderUnitID = `fixture_snapshot:default:ac043_large_grid_snapshot_v1:${snapshotKey}`;
  const observationArtifact = writeMeasurementArtifact(
    rootDirectory,
    "frontend-measurement-observation.v2.json",
    "frontend_measurement_observation",
    observation,
  );
  const buildArtifact = writeMeasurementArtifact(
    rootDirectory,
    "snapshot-build.json",
    "performance_fixture_snapshot_build",
    {
      schema_id: "cartulary.performance_fixture_snapshot.v2",
      fixture_profile_id: fixtureProfile.fixture_profile_id,
      fixture_version: fixtureProfile.fixture_version,
      seed: fixtureProfile.seed,
      snapshot_key_schema_id: "cartulary.performance_fixture_snapshot_key.v2",
      snapshot_key: snapshotKey,
      migration_digest: "a".repeat(64),
      source_contract_digest: fixtureProfile.source_contract_digest,
      builder_unit_id: builderUnitID,
      build_ordinal: 1,
      state: "sealed",
      contribution_receipts: fixtureProfile.contributions.map((contribution) => ({
        contribution_id: contribution.contribution_id,
        version: contribution.version,
        owner_id: contribution.owner_id,
        counts: contribution.expected_receipt_counts.map((count) => ({
          count_id: count.count_id,
          actual: count.exact,
        })),
      })),
      semantic_validation_digest: `sha256:${"c".repeat(64)}`,
      validation: {
        counts: fixtureProfile.semantic_expectations.counts.map((expectation) => ({
          expectation_id: expectation.expectation_id,
          actual: expectation.exact,
          passed: true,
        })),
        conditions: fixtureProfile.semantic_expectations.conditions.map((expectation) => ({
          expectation_id: expectation.expectation_id,
          actual: expectation.required,
          passed: true,
        })),
        connections_closed: true,
      },
      redaction_policy_id: fixtureProfile.redaction_policy.policy_id,
      created_at: "2026-08-14T00:00:00.000Z",
    },
  );
  const leaseArtifact = writeMeasurementArtifact(
    rootDirectory,
    "snapshot-lease.json",
    "performance_fixture_snapshot_lease",
    {
      schema_id: "cartulary.performance_fixture_snapshot_lease.v2",
      fixture_profile_id: "ac043_large_grid_snapshot_v1",
      snapshot_key: snapshotKey,
      builder_unit_id: builderUnitID,
      row_id: rowID,
      predicate_id: observation.predicate_id,
      lease_identity: `sha256:${"d".repeat(64)}`,
      clone_ordinal: 1,
      creation_state: "created",
      isolation_result: "isolated",
      cleanup_results: [
        "bucket",
        "credential_copy",
        "database",
        "process",
        "session",
      ].map((resource_class) => ({ resource_class, outcome: "complete" })),
      cleanup_state: "complete",
      redaction_policy_id: "cartulary.performance_fixture_redaction.v1",
      finalized_at: "2026-08-14T00:01:00.000Z",
      ...options.leaseOverrides,
    },
  );
  return {
    schema_id: "cartulary.frontend_measurement_summary.v3",
    row_id: rowID,
    fixture_profile_id: "ac043_large_grid_snapshot_v1",
    snapshot_key: snapshotKey,
    observation_artifact: observationArtifact.ref,
    build_artifact: buildArtifact.ref,
    lease_artifact: leaseArtifact.ref,
    clone_ordinal: 1,
    scheduler_overlap_count: 0,
    rollup: {
      predicate_id: observation.predicate_id,
      sample_count: observation.measured_samples,
      p50_ms: observation.p50_ms,
      p95_ms: observation.p95_ms,
      threshold_ms: observation.threshold_ms,
      outcome: observation.outcome,
    },
    qualification_outcome: "qualified",
    ...options.summaryOverrides,
  };
}

const measurementEvidenceRoot = mkdtempSync(
  path.join(os.tmpdir(), "cartulary-measurement-evidence-"),
);
try {
  const reportPath = path.join(measurementEvidenceRoot, "playwright-report.json");
  const collect = (summary, expected = [summary.predicate_id]) => {
    writeFileSync(reportPath, `${JSON.stringify(measurementReport(summary))}\n`);
    return collectFrontendMeasurementObservations({
      expectedPredicateIDs: expected,
      observationSchemaID: "cartulary.frontend_measurement_observation.v2",
      reportPaths: [reportPath],
      runRoot: measurementEvidenceRoot,
    });
  };
  assert.equal(collect(measurementObservation()).length, 1);
  assert.equal(
    collect(measurementObservation({ outcome: "threshold_failed", p95_ms: 101 }))[0].outcome,
    "threshold_failed",
    "threshold failures must remain valid diagnostic evidence",
  );
  assert.equal(
    collect(measurementObservation({
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
    () => collect(measurementObservation({
      measured_samples: 99,
      outcome: "incomplete",
      p50_ms: null,
      p95_ms: null,
    })),
    /sample counts differ/u,
    "summary cardinality must match its retained samples",
  );
  assert.throws(
    () => collect(measurementObservation({
      samples: measurementObservation().samples.map((sample, sampleIndex) =>
        sampleIndex === 0
          ? {
              ...sample,
              stages_ms: [
                { fact_id: "apply_to_visible_paint", value: 1, record_id: 1 },
              ],
            }
          : sample),
    })),
    /additional properties/u,
  );
  assert.throws(
    () => collect(measurementObservation(), []),
    /observations differ/u,
    "missing expected evidence must fail closed",
  );
  const finalizedPath = path.join(
    measurementEvidenceRoot,
    "frontend-measurement-summary.v3.json",
  );
  const collectWrittenSummary = (summary) => {
    writeFileSync(finalizedPath, `${JSON.stringify(summary)}\n`);
    return collectFinalizedMeasurementSummaries({
      buildSchemaID: "cartulary.performance_fixture_snapshot.v2",
      expectedPredicateIDs: [
        "perf.timeline_summary_selection_down.v1",
      ],
      leaseSchemaID: "cartulary.performance_fixture_snapshot_lease.v2",
      observationSchemaID: "cartulary.frontend_measurement_observation.v2",
      runRoot: measurementEvidenceRoot,
      summarySchemaID: "cartulary.frontend_measurement_summary.v3",
      summaryPaths: [finalizedPath],
    });
  };
  const collectFinalized = (options = {}) => {
    const summary = finalizedMeasurementSummary(
      measurementEvidenceRoot,
      options,
    );
    return collectWrittenSummary(summary);
  };
  assert.equal(collectFinalized().length, 1);
  assert.throws(
    () =>
      collectFinalized({
        observationOverrides: {
            failure_reason: "setup failed",
            measured_samples: 0,
            outcome: "incomplete",
            p50_ms: null,
            p95_ms: null,
            warmup_samples: 0,
            samples: [],
        },
        summaryOverrides: {
          rollup: {
            predicate_id: "perf.timeline_summary_selection_down.v1",
            sample_count: 0,
            p50_ms: null,
            p95_ms: null,
            threshold_ms: 100,
            outcome: "incomplete",
          },
          qualification_outcome: "environment_not_qualified",
        },
      }),
    /not eligible for active qualification/u,
  );
  assert.throws(
    () =>
      collectFinalized({
        summaryOverrides: {
          qualification_outcome: "threshold_failed",
        },
      }),
    /inconsistent qualification outcome/u,
  );
  assert.throws(
    () =>
      collectFinalizedMeasurementSummaries({
        buildSchemaID: "cartulary.performance_fixture_snapshot.v2",
        expectedPredicateIDs: [
          "perf.timeline_summary_selection_down.v1",
        ],
        leaseSchemaID: "cartulary.performance_fixture_snapshot_lease.v2",
        observationSchemaID: "cartulary.frontend_measurement_observation.v2",
        runRoot: measurementEvidenceRoot,
        summarySchemaID: "cartulary.frontend_measurement_summary.v3",
        summaryPaths: [finalizedPath, finalizedPath],
      }),
    /paths are duplicated/u,
  );
  assert.throws(
    () => collectFinalized({
      summaryOverrides: {
        schema_id: "cartulary.frontend_measurement_summary.v1",
      },
    }),
    /cartulary.frontend_measurement_summary.v3/u,
    "historical summary v1 must not qualify current source",
  );
  const tamperedReference = finalizedMeasurementSummary(
    measurementEvidenceRoot,
  );
  tamperedReference.observation_artifact.sha256 = `sha256:${"e".repeat(64)}`;
  assert.throws(
    () => collectWrittenSummary(tamperedReference),
    /digest or size is invalid/u,
    "a mutated referenced artifact digest must fail qualification",
  );
  const wrongRole = finalizedMeasurementSummary(measurementEvidenceRoot);
  wrongRole.observation_artifact.role = "diagnostic_observation";
  assert.throws(
    () => collectWrittenSummary(wrongRole),
    /artifact reference is malformed/u,
    "artifact references must retain their exact provenance role",
  );
  assert.throws(
    () => collectFinalized({
      leaseOverrides: {
        cleanup_results: [
          "bucket",
          "bucket",
          "credential_copy",
          "database",
          "process",
          "session",
        ].map((resource_class) => ({ resource_class, outcome: "complete" })),
      },
    }),
    /duplicates cleanup class bucket/u,
    "duplicate cleanup classes must not make lease evidence ambiguous",
  );
} finally {
  rmSync(measurementEvidenceRoot, { force: true, recursive: true });
}

const measurementGroupResult = {
  group_id: "measurement-timeline-grid",
  stage_id: "measurement",
};
const measurementUnitID = "browser_group:measurement:measurement-timeline-grid";
const schedulerStreamRoot = mkdtempSync(
  path.join(os.tmpdir(), "cartulary-measurement-scheduler-stream-"),
);
try {
  const eventFile = path.join(schedulerStreamRoot, "unit-events.ndjson");
  const file = openSync(eventFile, "w", 0o600);
  let seq = 0;
  const writeEvent = (event) => {
    seq += 1;
    writeSync(file, `${JSON.stringify({
      schema_id: "cartulary.harness_unit_event.v1",
      seq,
      monotonic_ms: seq,
      resource_claims: {},
      service_dependencies: [],
      status: "running",
      ...event,
    })}\n`);
  };
  for (let index = 0; index < 20_000; index += 1) {
    writeEvent({ event: "queued", unit_id: `row:release-prefix-${index}` });
  }
  writeEvent({ event: "started", unit_id: measurementUnitID });
  for (let index = 0; index < 20_000; index += 1) {
    writeEvent({ event: "completed", status: "passed", unit_id: `row:release-prefix-${index}` });
  }
  writeEvent({ event: "completed", unit_id: measurementUnitID, status: "passed" });
  closeSync(file);
  const streamedEvidence = await readMeasurementSchedulerEvidence(
    eventFile,
    measurementGroupResult,
  );
  assert.equal(streamedEvidence.dependency_skipped, false);
  assert.equal(streamedEvidence.overlap_count, 0);
  assert.equal(streamedEvidence.start_seq, 20_001);
  assert.equal(streamedEvidence.end_seq, 40_002);

  const overlapFile = path.join(schedulerStreamRoot, "overlap-events.ndjson");
  const overlapEvents = [
    {
      event: "started",
      status: "running",
      unit_id: measurementUnitID,
    },
    { event: "started", status: "running", unit_id: "row:ordinary" },
    { event: "completed", status: "passed", unit_id: "row:ordinary" },
    { event: "completed", status: "passed", unit_id: measurementUnitID },
  ].map((entry, index) => ({
    schema_id: "cartulary.harness_unit_event.v1",
    seq: index + 1,
    monotonic_ms: index + 1,
    resource_claims: {},
    service_dependencies: [],
    ...entry,
  }));
  writeFileSync(
    overlapFile,
    `${overlapEvents.map((entry) => JSON.stringify(entry)).join("\n")}\n`,
  );
  assert.equal(
    (await readMeasurementSchedulerEvidence(overlapFile, measurementGroupResult))
      .overlap_count,
    1,
    "ordinary work beginning inside a measurement session must disqualify it",
  );

  const incompleteFile = path.join(schedulerStreamRoot, "incomplete-events.ndjson");
  writeFileSync(incompleteFile, `${JSON.stringify(overlapEvents[0])}\n`);
  await assert.rejects(
    readMeasurementSchedulerEvidence(incompleteFile, measurementGroupResult),
    /lacks a closed scheduler interval/u,
    "incomplete scheduler proof must fail closed",
  );

  const skippedFile = path.join(schedulerStreamRoot, "skipped-events.ndjson");
  writeFileSync(
    skippedFile,
    `${JSON.stringify({
      schema_id: "cartulary.harness_unit_event.v1",
      event: "skipped",
      failure_reason: "dependency_failure",
      monotonic_ms: 1,
      resource_claims: {},
      seq: 1,
      service_dependencies: [],
      status: "skipped",
      unit_id: measurementUnitID,
    })}\n`,
  );
  const skippedEvidence = await readMeasurementSchedulerEvidence(
    skippedFile,
    measurementGroupResult,
  );
  assert.equal(skippedEvidence.dependency_skipped, true);

  const invalidSequenceFile = path.join(
    schedulerStreamRoot,
    "invalid-sequence-events.ndjson",
  );
  writeFileSync(
    invalidSequenceFile,
    `${JSON.stringify({
      schema_id: "cartulary.harness_unit_event.v1",
      event: "started",
      monotonic_ms: 1,
      resource_claims: {},
      seq: 2,
      service_dependencies: [],
      status: "running",
      unit_id: measurementUnitID,
    })}\n`,
  );
  await assert.rejects(
    readMeasurementSchedulerEvidence(invalidSequenceFile, measurementGroupResult),
    /sequence 2 is not contiguous/u,
  );
} finally {
  rmSync(schedulerStreamRoot, { force: true, recursive: true });
}
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
  assert.equal(unit.resource_claims.object_store, 1, `${unit.unit_id} must reserve object-store capacity`);
  assert.deepEqual(unit.service_dependencies, ["object_store", "postgres"]);
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
const snapshotBuilders = measurement.units.filter((entry) =>
  entry.unit_id.startsWith("fixture_snapshot:default:ac043_large_grid_snapshot_v1:"),
);
assert.equal(snapshotBuilders.length, 1, "the four AC-043 rows must share one builder");
assert.equal(snapshotBuilders[0].kind, "fixture_builder");
assert.equal(snapshotBuilders[0].fixture_lease, "postgres_dedicated");
assert.deepEqual(snapshotBuilders[0].shared_locks, ["host_activity"]);
assert.deepEqual(snapshotBuilders[0].exclusive_locks, []);
assert.match(snapshotBuilders[0].snapshot_key, /^[a-f0-9]{64}$/u);
const profiledRowIDs = compiler.catalog.rows
  .filter((row) => row.fixture_profile_id === snapshotBuilders[0].fixture_profile_id)
  .map((row) => row.row_id)
  .sort();
const fixtureIdentity = (graph) => {
  const builders = graph.units.filter((unit) => unit.kind === "fixture_builder");
  assert.equal(builders.length, 1, "profiled route must resolve one fixture builder");
  return {
    builder_unit_id: builders[0].unit_id,
    fixture_profile_id: builders[0].fixture_profile_id,
    snapshot_key: builders[0].snapshot_key,
  };
};
const expectedFixtureIdentity = fixtureIdentity(measurement);
for (const graph of [
  compiler.compile({ kind: "rows", row_ids: [profiledRowIDs[0]] }),
  compiler.compile({ kind: "rows", row_ids: profiledRowIDs }),
  compiler.compile({ kind: "owner", owner_id: "module.timeline" }),
  compiler.compile({ kind: "aggregate", target: "release-check" }),
]) {
  assert.deepEqual(
    fixtureIdentity(graph),
    expectedFixtureIdentity,
    "direct, row, owner, aggregate, and release routes must share fixture identity",
  );
}
const measurementSummaries = measurement.units.filter((entry) =>
  entry.unit_id.startsWith("browser_measurement_summary:measurement:"),
);
assert.equal(
  measurementSummaries.length,
  4,
  "each AC-043 row must have one post-cleanup summary finalizer",
);
for (const summary of measurementSummaries) {
  assert.equal(summary.kind, "finalizer");
  assert.equal(summary.fixture_lease, "none");
  assert.ok(Object.keys(summary.resource_claims).length > 0);
  assert.deepEqual(summary.shared_locks, ["host_activity"]);
  assert.deepEqual(summary.exclusive_locks, []);
  assert.equal(summary.needs.length, 1);
  assert.ok(summary.needs[0].startsWith("browser_group:measurement:"));
  assert.match(
    summary.evidence_outputs[0],
    /frontend-measurement-summary\.v3\.json$/u,
  );
}
const measurementTargetFinalizer = measurement.units.find(
  (entry) =>
    entry.unit_id === "browser_target_summary:browser-e2e-measurement",
);
assert.ok(measurementTargetFinalizer);
for (const summary of measurementSummaries) {
  assert.ok(
    measurementTargetFinalizer.needs.includes(summary.unit_id),
    "the aggregate must wait for every row finalizer",
  );
}
for (const lifecycle of measurement.units.filter((entry) =>
  entry.unit_id.startsWith("browser_lifecycle:measurement-measurement-measurement-timeline-grid-"),
)) {
  assert.ok(
    lifecycle.needs.includes(snapshotBuilders[0].unit_id),
    lifecycle.unit_id + " must depend on the shared snapshot builder",
  );
  assert.equal(lifecycle.fixture_profile_id, "ac043_large_grid_snapshot_v1");
  assert.equal(lifecycle.snapshot_key, snapshotBuilders[0].snapshot_key);
  assert.match(
    lifecycle.command.environment.CARTULARY_FIXTURE_ROW_ID,
    /^module\.timeline\.measurement\./u,
  );
  assert.match(
    lifecycle.command.environment.CARTULARY_FIXTURE_PREDICATE_ID,
    /^perf\./u,
  );
}
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
