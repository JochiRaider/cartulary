#!/usr/bin/env node

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import {
  closeSync,
  mkdirSync,
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
  currentUnitEventFile,
  readMeasurementSchedulerEvidence,
} from "../frontend-measurement-evidence.mjs";
import {
  activePerformanceFixtureProfile,
  loadPerformanceFixtureSnapshotRegistry,
} from "../../performance-fixture/index.mjs";
import { publicExitCodeForFailure } from "../../contract/index.mjs";
import {
  executeUnitProcess,
  WorkGraphCompiler,
} from "../../scheduler/work-graph/index.mjs";
import { planBrowserFunctionalLanes } from "../../scheduler/work-graph/browser-functional-lanes.mjs";
import { simulateWorkGraph } from "../../scheduler/work-graph/scheduler.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const compiler = new WorkGraphCompiler(root);

const lifecycleFailureRoot = mkdtempSync(
  path.join(os.tmpdir(), "cartulary-browser-reset-failure."),
);
try {
  const runRoot = path.join(lifecycleFailureRoot, "run");
  mkdirSync(runRoot);
  writeFileSync(
    path.join(runRoot, "reset.attempt.json"),
    `${JSON.stringify({
      schema_id: "cartulary.browser_reset_attempt.v1",
      reset_id: "predecessor--before-successor",
      status: "fail",
      attempt: 1,
      duration_ms: 5,
      runtime_profile_id: "default",
      stages: [
        { stage: "allocation_validated", status: "pass" },
        { stage: "backend_stopped", status: "fail" },
        { stage: "database_reset", status: "skipped" },
        { stage: "object_store_reset", status: "skipped" },
        { stage: "browser_state_cleared", status: "skipped" },
        { stage: "replacement_backend_ready", status: "skipped" },
        { stage: "generation_published", status: "skipped" },
      ],
      backend_generation_before: 1,
      backend_generation_after: 1,
      failure_stage: "backend_stop_or_preflight",
      database_diagnostic_ref: null,
      database_stage: null,
      database_sqlstate: null,
      persistent_state_reset: false,
      browser_state_cleared: false,
      backend_ready: false,
      backend_generation_ref: null,
      tainted: true,
      failure_class: "harness",
      failure_reason: "fixture_error",
    }, null, 2)}\n`,
  );
  const failure = await executeUnitProcess(
    {
      kind: "lifecycle",
      command: {
        executable: process.execPath,
        args: ["--eval", "process.exit(1)"],
        environment: {},
      },
      current_run_evidence_outputs: ["reset.attempt.json"],
      timeout_ms: 1000,
    },
    {
      cwd: lifecycleFailureRoot,
      environment: {
        CARTULARY_TEST_RESULTS_DIR: ".",
        CARTULARY_TEST_RUN_ID: "run",
      },
    },
  );
  assert.equal(failure.failure_class, "harness");
  assert.equal(failure.failure_reason, "fixture_error");
  assert.equal(publicExitCodeForFailure(failure), 3);
} finally {
  rmSync(lifecycleFailureRoot, { recursive: true, force: true });
}
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
      schema_id: "cartulary.harness_unit_event.v2",
      seq,
      monotonic_ms: seq,
      resource_claims: {},
      needs: [],
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

  const previousLiveEventFile = process.env.CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE;
  const liveEventFile = path.join(schedulerStreamRoot, "unit-events.ndjson.tmp-123-0");
  writeFileSync(liveEventFile, readFileSync(eventFile), { mode: 0o600 });
  try {
    process.env.CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE = liveEventFile;
    assert.equal(currentUnitEventFile(schedulerStreamRoot), liveEventFile);
    process.env.CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE = path.join(os.tmpdir(), "unit-events.ndjson.tmp-escape");
    assert.throws(
      () => currentUnitEventFile(schedulerStreamRoot),
      /escapes its run/u,
      "a same-run finalizer must reject an external staging stream",
    );
  } finally {
    if (previousLiveEventFile === undefined) {
      delete process.env.CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE;
    } else {
      process.env.CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE = previousLiveEventFile;
    }
  }
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
    schema_id: "cartulary.harness_unit_event.v2",
    seq: index + 1,
    monotonic_ms: index + 1,
    resource_claims: {},
    needs: [],
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
      schema_id: "cartulary.harness_unit_event.v2",
      event: "skipped",
      failure_reason: "dependency_failure",
      monotonic_ms: 1,
      resource_claims: {},
      needs: [],
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
      schema_id: "cartulary.harness_unit_event.v2",
      event: "started",
      monotonic_ms: 1,
      resource_claims: {},
      needs: [],
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
  visualTargetFinalizer.current_run_evidence_outputs.includes(
    "browser-e2e-visual/frontend-visual-reconciliation.json",
  ),
  "visual target finalizer must retain the reconciliation artifact",
);
const stateful = compiler.compile({ kind: "target", target: "browser-e2e-stateful" });
assert.ok(stateful.units.some((unit) => unit.unit_id.startsWith("browser_reset:")), "stateful browser work must expose resets");

function assertResetTargetMatchesEvidence(graph, label) {
  const resets = graph.units.filter((unit) => unit.unit_id.startsWith("browser_reset:"));
  assert.ok(resets.length > 0, `${label} must contain a reset unit`);
  for (const reset of resets) {
    assert.equal(reset.current_run_evidence_outputs.length, 1);
    assert.equal(
      reset.command.environment.CARTULARY_TEST_TARGET,
      reset.current_run_evidence_outputs[0].split("/", 1)[0],
      `${reset.unit_id} must write evidence beneath its declared target`,
    );
    assert.equal(reset.command.environment.OWNER, "", `${reset.unit_id} must clear OWNER`);
    assert.equal(reset.command.environment.ROWS, "", `${reset.unit_id} must clear ROWS`);
    assert.equal(
      reset.command.environment.SERVICE_BACKED_ONLY,
      "",
      `${reset.unit_id} must clear SERVICE_BACKED_ONLY`,
    );
    assert.equal(
      reset.command.environment.CARTULARY_MAKE_INPUT_SOURCES,
      "",
      `${reset.unit_id} must clear inherited public input provenance`,
    );
  }
}

const collaborationStatefulOwner = compiler.compile({
  kind: "owner",
  owner_id: "module.collaboration",
});
assertResetTargetMatchesEvidence(collaborationStatefulOwner, "Collaboration owner selection");

const networkFlowStatefulOwner = compiler.compile({
  kind: "owner",
  owner_id: "module.networkflow",
});
const selectedNetworkFlowGroup = networkFlowStatefulOwner.units.find((unit) =>
  unit.unit_id === "browser_group:stateful:stateful-network-flow-claimed-network-flow"
);
assert.ok(selectedNetworkFlowGroup, "Network Flow owner selection must contain its stateful group");
assert.equal(
  networkFlowStatefulOwner.units.some((unit) => unit.unit_id.startsWith("browser_reset:stateful:")),
  false,
  "a first selected stateful group must not receive a reset inherited from the unfiltered graph",
);

const claimedStatefulChain = compiler.compile({
  kind: "rows",
  row_ids: [
    "module.extensions.browser_stateful.bc015_availability_continuity_d538000c38",
    "module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle",
    "module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf",
  ],
});
const claimedPredecessor = claimedStatefulChain.units.find((unit) =>
  unit.unit_id === "browser_group:stateful:stateful-network-flow-claimed-extensions-stateful"
);
const claimedSuccessor = claimedStatefulChain.units.find((unit) =>
  unit.unit_id === "browser_group:stateful:stateful-network-flow-claimed-network-flow"
);
const claimedResets = claimedStatefulChain.units.filter((unit) =>
  unit.unit_id.startsWith("browser_reset:stateful:")
);
assert.ok(claimedPredecessor && claimedSuccessor);
assert.equal(claimedResets.length, 1, "one selected affinity predecessor requires one reset");
assert.deepEqual(claimedResets[0].needs, [claimedPredecessor.unit_id]);
assert.deepEqual(claimedSuccessor.needs, [claimedResets[0].unit_id]);
assert.match(
  claimedResets[0].unit_id,
  /extensions-stateful--before-stateful-network-flow-claimed-network-flow/u,
);
assertResetTargetMatchesEvidence(claimedStatefulChain, "explicit row selection");

const incompatibleStatefulSelection = compiler.compile({
  kind: "rows",
  row_ids: [
    "module.extensions.browser_stateful.bc015_availability_continuity_d538000c38",
    "module.savedviews.browser_stateful.browser_saved_view_query_layout_state_user_home_cb9c681674",
  ],
});
assert.equal(
  incompatibleStatefulSelection.units.some((unit) => unit.unit_id.startsWith("browser_reset:stateful:")),
  false,
  "incompatible stateful affinities must never share a reset",
);
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
  const resets = units.filter((unit) => unit.unit_id.startsWith("browser_reset:"));
  const chainedResets = resets.filter((unit) =>
    unit.needs.some((dependency) => dependency.startsWith("browser_reset:")),
  );
  assert.equal(
    chainedResets.length,
    Math.max(resets.length - 1, 0),
    `${affinity} must propagate a reset failure across the remaining lifecycle chain`,
  );
}

const syntheticFunctionalGroups = [
  ["default-long", "default", 9000],
  ["default-next", "default", 8000],
  ["default-third", "default", 7000],
  ["default-fourth", "default", 6000],
  ["claimed-long", "network_flow_claimed", 4000],
  ["claimed-next", "network_flow_claimed", 3000],
].map(([name, runtimeProfileID, estimate]) => ({
  name,
  runtimeProfileID,
  resourceProfileID: "browser_functional",
  estimate,
}));
const laneProjection = (groups) =>
  planBrowserFunctionalLanes(groups, {
    estimateGroup: (group) => group.estimate,
    lanePrefix: "functional-test",
    maxLanes: 4,
  }).map((lane) => ({
    lane_id: lane.laneID,
    runtime_profile_id: lane.runtimeProfileID,
    estimated_work_ms: lane.estimatedWorkMs,
    groups: lane.groups.map((item) => [
      item.group.name,
      item.estimatedWorkMs,
      item.generation,
    ]),
  }));
const syntheticLanePlan = laneProjection(syntheticFunctionalGroups);
assert.deepEqual(
  syntheticLanePlan.map((lane) => lane.runtime_profile_id),
  ["default", "default", "default", "network_flow_claimed"],
  "four total lanes must be allocated without mixing immutable runtime profiles",
);
assert.deepEqual(
  syntheticLanePlan,
  laneProjection([...syntheticFunctionalGroups].reverse()),
  "functional lane planning must be independent of input order",
);
assert.deepEqual(
  syntheticLanePlan.flatMap((lane) => lane.groups.map(([name]) => name)).sort(),
  syntheticFunctionalGroups.map((group) => group.name).sort(),
  "stable LPT must assign every group exactly once",
);

const webserverBacked = compiler.compile({
  kind: "target",
  target: "browser-e2e-webserver-backed",
});
const functionalGroups = webserverBacked.units.filter((unit) =>
  unit.unit_id.startsWith("browser_group:webserver-backed:"),
);
const functionalResets = webserverBacked.units.filter((unit) =>
  unit.unit_id.startsWith("browser_reset:webserver-backed:"),
);
assert.equal(
  webserverBacked.units.some((unit) => unit.unit_id.startsWith("browser_lifecycle:")),
  false,
  "the first executable browser consumer must own its stack without a synthetic readiness handoff",
);
assert.equal(functionalGroups.length, 26);
assert.equal(
  functionalGroups.flatMap((unit) =>
    unit.current_run_evidence_outputs.filter((output) => output.startsWith("rows/")),
  ).length,
  83,
  "functional lane graph must preserve the exact current row closure",
);
assert.equal(
  new Set(functionalGroups.map((unit) => unit.affinity_key)).size,
  2,
  "functional groups must derive affinity from their two declared browser-session/runtime tuples",
);
assert.equal(
  functionalResets.length,
  functionalGroups.length - 2,
  "each selected browser affinity must reset only between adjacent groups",
);
for (const unit of [...functionalResets, ...functionalGroups]) {
  assert.equal(unit.resource_claims.postgres, 2, `${unit.unit_id} must claim two PostgreSQL tokens`);
  assert.equal(unit.resource_claims.browser_stack, 1);
  assert.equal(unit.resource_claims.port_lane, 1);
  assert.equal(
    unit.command.environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID,
    "browser_functional",
  );
}
for (const unit of functionalGroups) {
  const rowCount = unit.current_run_evidence_outputs.filter((output) =>
    output.startsWith("rows/"),
  ).length;
  assert.equal(
    unit.estimated_work_ms,
    rowCount * compiler.owner.evidence_estimates_ms.browser,
    `${unit.unit_id} cost must sum selected row estimates`,
  );
  assert.match(unit.command.environment.CARTULARY_BROWSER_FUNCTIONAL_LANE_ID, /^webserver-backed-/u);
  assert.match(unit.command.environment.CARTULARY_BROWSER_GROUP_GENERATION, /^[1-9][0-9]*$/u);
}
for (const reset of functionalResets) {
  assert.equal(reset.failure_policy.block_descendants, true);
  assert.equal(reset.command.args.includes("--renew-generation"), false);
  assert.match(reset.command.environment.CARTULARY_BROWSER_FUNCTIONAL_LANE_ID, /^webserver-backed-/u);
}
const functionalTargetFinalizer = webserverBacked.units.find(
  (unit) => unit.unit_id === "browser_target_summary:browser-e2e-webserver-backed",
);
assert.ok(functionalTargetFinalizer);
for (const reset of functionalResets) {
  assert.ok(
    functionalTargetFinalizer.needs.includes(reset.unit_id),
    "functional target finalizer must observe every reset",
  );
}
for (const affinity of new Set(functionalGroups.map((unit) => unit.affinity_key))) {
  const laneUnits = webserverBacked.units.filter((unit) => unit.affinity_key === affinity);
  const runtimeProfiles = new Set(
    laneUnits.map((unit) => unit.command.environment.CARTULARY_BROWSER_RUNTIME_PROFILE_ID),
  );
  assert.equal(runtimeProfiles.size, 1, `${affinity} must retain immutable runtime identity`);
  assert.equal(
    laneUnits.filter((unit) =>
      unit.command.environment.CARTULARY_BROWSER_RELEASE_AFFINITY === "1",
    ).length,
    1,
    `${affinity} must have one terminal stack releaser`,
  );
}

const simulationCapacities = new Map();
for (const unit of webserverBacked.units) {
  for (const resource of Object.keys(unit.resource_claims)) {
    simulationCapacities.set(resource, 100);
  }
}
const oneMsDurations = new Map(
  webserverBacked.units.map((unit) => [unit.unit_id, 1]),
);
const resetAfterProductFailure = functionalResets[0];
const failedProductUnitID = resetAfterProductFailure.needs[0];
const groupAfterProductFailure = functionalGroups.find((unit) =>
  unit.needs.includes(resetAfterProductFailure.unit_id),
);
const productFailureSimulation = simulateWorkGraph({
  graph: webserverBacked,
  capacities: simulationCapacities,
  durations: oneMsDurations,
  outcomes: new Map([[failedProductUnitID, "product"]]),
});
assert.equal(productFailureSimulation.status, "failed");
assert.equal(productFailureSimulation.states[resetAfterProductFailure.unit_id], "passed");
assert.equal(productFailureSimulation.states[groupAfterProductFailure.unit_id], "passed");
assert.equal(
  productFailureSimulation.admissions.filter((unitID) => unitID === failedProductUnitID).length,
  1,
  "product failures must not be retried",
);
const resetFailureSimulation = simulateWorkGraph({
  graph: webserverBacked,
  capacities: simulationCapacities,
  durations: oneMsDurations,
  outcomes: new Map([[resetAfterProductFailure.unit_id, "infra"]]),
});
assert.equal(resetFailureSimulation.status, "failed");
assert.equal(resetFailureSimulation.states[resetAfterProductFailure.unit_id], "failed");
assert.equal(
  resetFailureSimulation.states[groupAfterProductFailure.unit_id],
  "passed",
  "reset failure must allow later work to obtain a fresh allocation",
);

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
assert.deepEqual(snapshotBuilders[0].service_dependencies, ["postgres"]);
assert.deepEqual(snapshotBuilders[0].resource_claims, {
  cpu: 1,
  io: 1,
  memory_mb: 512,
  postgres: 1,
  process: 1,
});
assert.equal(
  snapshotBuilders[0].command.environment.CARTULARY_FIXTURE_BUILDER_RESOURCE_PROFILE_ID,
  "performance_fixture_builder",
);
for (const forbiddenResource of ["browser_stack", "object_store", "port_lane"]) {
  assert.equal(
    Object.hasOwn(snapshotBuilders[0].resource_claims, forbiddenResource),
    false,
    `fixture builder must not claim ${forbiddenResource}`,
  );
}
assert.deepEqual(snapshotBuilders[0].shared_locks, ["host_activity"]);
assert.deepEqual(snapshotBuilders[0].exclusive_locks, []);
assert.match(snapshotBuilders[0].snapshot_key, /^[a-f0-9]{64}$/u);
assert.deepEqual(snapshotBuilders[0].current_run_evidence_outputs, [
  `performance-fixtures/${snapshotBuilders[0].snapshot_key}/build-diagnostics.json`,
  `performance-fixtures/${snapshotBuilders[0].snapshot_key}/snapshot-build.json`,
]);
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
    summary.current_run_evidence_outputs[0],
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
for (const firstConsumer of measurement.units.filter((entry) =>
	entry.unit_id.startsWith("browser_group:measurement:measurement-measurement-timeline-grid-"),
)) {
	assert.ok(
		firstConsumer.needs.includes(snapshotBuilders[0].unit_id),
		firstConsumer.unit_id + " must depend on the shared snapshot builder",
	);
	assert.equal(firstConsumer.fixture_profile_id, "ac043_large_grid_snapshot_v1");
	assert.equal(firstConsumer.snapshot_key, snapshotBuilders[0].snapshot_key);
	assert.match(
		firstConsumer.command.environment.CARTULARY_FIXTURE_ROW_ID,
		/^module\.timeline\.measurement\./u,
	);
	assert.match(
		firstConsumer.command.environment.CARTULARY_FIXTURE_PREDICATE_ID,
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
    "browser_functional",
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
    unit.current_run_evidence_outputs
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
