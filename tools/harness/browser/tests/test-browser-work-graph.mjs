#!/usr/bin/env node
import { referencePackBundle } from "../reference-pack-fixture.mjs";
import { validateViewChangeCapabilities } from "../../generated-artifacts/performance/performance-contracts.mjs";
import { validateEvidenceInspectorCapabilities } from "../../generated-artifacts/performance/evidence-inspector-contract.mjs";

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import "./test-frontend-artifact.mjs";
import { createHash } from "node:crypto";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import {
  chmodSync,
  closeSync,
  lstatSync,
  linkSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  openSync,
  readFileSync,
  renameSync,
  rmSync,
  symlinkSync,
  writeFileSync,
  writeSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";

import {
  classifyFrontendVisualGoldens,
  validateCapturePayload,
  resolveRegisteredFixtures,
} from "../frontend-visual-reconciliation.mjs";
import {
  buildFrontendVisualGoldenManifest,
  validateFrontendVisualGoldenManifest,
} from "../frontend-visual-golden-manifest.mjs";
import {
  assertVisualRendererEnvironmentIsPrivate,
  loadVisualRendererProfile,
} from "../visual-renderer-lease.mjs";
import { promoteVisualSnapshotCandidate, stageVisualSnapshotCandidate } from "../visual-snapshot-promotion.mjs";
import {
  collectFinalizedMeasurementSummaries,
  collectFrontendMeasurementObservations,
  currentUnitEventSource,
  readMeasurementSchedulerEvidence,
} from "../frontend-measurement-evidence.mjs";
import {
  activePerformanceFixtureProfile,
  loadPerformanceFixtureSnapshotRegistry,
} from "../../performance-fixture/index.mjs";
import { canonicalJSONString, assertRetainedArtifactFile, publicExitCodeForFailure, readRetainedArtifactBytes, validateSchemaSync } from "../../contract/index.mjs";
import { validateProfileMeasurementObservation } from "../../performance-fixture/profile-adapters/index.mjs";
import { fixtureDigest } from "../../performance-fixture/profile-adapters/large-grid.mjs";
import {
  WorkGraphCompiler,
} from "../../scheduler/work-graph/index.mjs";

const artifactReaderRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-evidence-reader-"));
try {
  const admitted = path.join(artifactReaderRoot, "admitted");
  const displaced = path.join(artifactReaderRoot, "displaced");
  const outside = path.join(artifactReaderRoot, "outside");
  mkdirSync(admitted); mkdirSync(outside);
  writeFileSync(path.join(admitted, "evidence.json"), "original");
  writeFileSync(path.join(outside, "evidence.json"), "redirected");
  const read = () => readRetainedArtifactBytes(artifactReaderRoot, "admitted/evidence.json", 8).toString();
  assert.equal(read(), "original");
  assertRetainedArtifactFile(artifactReaderRoot, "admitted/evidence.json");
  assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, "admitted/evidence.json", 7), /bounded/u);
  assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, "admitted", 8), /regular file/u);
  assert.throws(() => assertRetainedArtifactFile(artifactReaderRoot, "admitted"), /regular file/u);
  for (const relative of ["../outside", "admitted/../outside", "/absolute", "admitted//evidence.json", "admitted\\evidence.json"]) {
    assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, relative, 8), /normalized and relative/u);
  }
  symlinkSync(admitted, path.join(artifactReaderRoot, "alias"));
  symlinkSync(path.join(admitted, "evidence.json"), path.join(admitted, "alias.json"));
  for (const relative of ["alias/evidence.json", "admitted/alias.json"]) {
    assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, relative, 8), /ELOOP|ENOTDIR/u);
    assert.throws(() => assertRetainedArtifactFile(artifactReaderRoot, relative), /ELOOP|ENOTDIR/u);
  }
  assert.throws(() => readRetainedArtifactBytes(path.join(artifactReaderRoot, "alias"), "evidence.json", 8), /ELOOP|ENOTDIR/u);
  linkSync(path.join(admitted, "evidence.json"), path.join(admitted, "hardlink.json"));
  assert.throws(read, /unaliased/u);
  assert.throws(() => assertRetainedArtifactFile(artifactReaderRoot, "admitted/hardlink.json"), /unaliased/u);
  rmSync(path.join(admitted, "hardlink.json"));

  const originalOpen = fs.openSync;
  const originalRead = fs.readSync;
  const originalClose = fs.closeSync;
  const opened = new Set();
  try {
    // No forbidden documentation component may even reach the filesystem.
    fs.openSync = () => { throw new Error("unexpected filesystem traversal"); };
    syncBuiltinESMExports();
    // These are hostile path tokens, not references to repository documents.
    const forbiddenDirectory = "docs";
    for (const relative of [`${forbiddenDirectory}/evidence.json`, "evidence.md", "nested/README", "nested/readme.txt"]) {
      assert.throws(() => readRetainedArtifactBytes(artifactReaderRoot, relative, 8), /documentation/u);
      assert.throws(() => assertRetainedArtifactFile(artifactReaderRoot, relative), /documentation/u);
    }
    assert.throws(() => readRetainedArtifactBytes(path.join(artifactReaderRoot, forbiddenDirectory), "evidence.json", 8), /documentation/u);

    fs.openSync = (file, flags) => {
      if (file.endsWith("/evidence.json")) {
        renameSync(admitted, displaced);
        symlinkSync(outside, admitted);
      }
      const descriptor = originalOpen(file, flags);
      opened.add(descriptor);
      return descriptor;
    };
    fs.closeSync = (descriptor) => { opened.delete(descriptor); return originalClose(descriptor); };
    syncBuiltinESMExports();
    assert.equal(read(), "original", "renaming an ancestor cannot redirect a descriptor-relative read");
    assert.equal(opened.size, 0);
    rmSync(admitted); renameSync(displaced, admitted);

    fs.openSync = (file, flags) => {
      const descriptor = originalOpen(file, flags); opened.add(descriptor); return descriptor;
    };
    let mutated = false;
    fs.readSync = (...args) => {
      if (!mutated) { mutated = true; writeFileSync(path.join(admitted, "evidence.json"), "changed"); }
      return originalRead(...args);
    };
    syncBuiltinESMExports();
    assert.throws(read, /changed during read/u);
    assert.equal(opened.size, 0, "failed reads close every descriptor");
  } finally {
    fs.openSync = originalOpen; fs.readSync = originalRead; fs.closeSync = originalClose;
    syncBuiltinESMExports();
  }
} finally {
  rmSync(artifactReaderRoot, { recursive: true, force: true });
}
import { planBrowserFunctionalLanes } from "../../scheduler/work-graph/browser-functional-lanes.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
// Malformed machine owners must reject inadmissible source paths before any
// filesystem probe. Links are never a way to admit a second contract location.
const registryBoundaryRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-registry-boundary-"));
try {
  const registry = JSON.parse(readFileSync(path.join(root, "tools/performance_fixture_snapshot_owner.json"), "utf8"));
  const registryPath = path.join(registryBoundaryRoot, "registry.json");
  function selectSource(relative) {
    const value = structuredClone(registry);
    value.profiles[0].source_contract_refs = [{ ...value.profiles[0].source_contract_refs[0], path: relative }];
    writeFileSync(registryPath, JSON.stringify(value));
  }
  const forbiddenDirectory = "docs";
  const forbidden = [`${forbiddenDirectory}/source.json`, `contracts/${forbiddenDirectory}/source.json`, "contracts/source.MD", "contracts/README"];
  const probes = ["openSync", "statSync", "lstatSync", "realpathSync", "readdirSync"];
  const originals = new Map(probes.map(name => [name, fs[name]]));
  let forbiddenProbes = 0;
  try {
    for (const name of probes) fs[name] = (file, ...args) => {
      const token = String(file);
      if (token.split(path.sep).includes(forbiddenDirectory) || /(?:\.(?:md|markdown)$|(?:^|\/)readme(?:\.|$))/iu.test(token)) {
        forbiddenProbes++;
        throw new Error("unexpected forbidden filesystem probe");
      }
      return originals.get(name)(file, ...args);
    };
    syncBuiltinESMExports();
    for (const relative of forbidden) {
      selectSource(relative);
      assert.throws(() => loadPerformanceFixtureSnapshotRegistry(root, { registryPath }), /under contracts|documentation/u);
    }
    assert.equal(forbiddenProbes, 0);
  } finally {
    for (const [name, original] of originals) fs[name] = original;
    syncBuiltinESMExports();
  }
  mkdirSync(path.join(registryBoundaryRoot, "contracts"));
  writeFileSync(path.join(registryBoundaryRoot, "source.json"), "{}");
  symlinkSync("../source.json", path.join(registryBoundaryRoot, "contracts/alias.json"));
  selectSource("contracts/alias.json");
  assert.throws(() => loadPerformanceFixtureSnapshotRegistry(registryBoundaryRoot, { registryPath }), /ELOOP/u);
  symlinkSync(registryBoundaryRoot, path.join(registryBoundaryRoot, "contracts/alias"));
  selectSource("contracts/alias/source.json");
  assert.throws(() => loadPerformanceFixtureSnapshotRegistry(registryBoundaryRoot, { registryPath }), /ENOTDIR|ELOOP/u);
  linkSync(path.join(registryBoundaryRoot, "source.json"), path.join(registryBoundaryRoot, "contracts/hardlink.json"));
  selectSource("contracts/hardlink.json");
  assert.throws(() => loadPerformanceFixtureSnapshotRegistry(registryBoundaryRoot, { registryPath }), /unaliased/u);
} finally {
  rmSync(registryBoundaryRoot, { recursive: true, force: true });
}
let evidenceInspectorObservation;
{
  const contract = JSON.parse(readFileSync(path.join(root, "contracts/performance/ac045.v2.json"), "utf8"));
  const view = JSON.parse(readFileSync(path.join(root, "contracts/view-schemas/cartulary.view.timeline.v2.json"), "utf8"));
  const policy = JSON.parse(readFileSync(path.join(root, "tools/measurement_policy_owner.json"), "utf8"));
  const validate = (value) => {
    validateSchemaSync("cartulary.performance_evidence_inspector_contract.v1", value);
    validateEvidenceInspectorCapabilities(value, view, policy);
  };
  validate(contract);
  for (const mutate of [
    (value) => { value.predicate_id = "perf.evidence_inspector.metadata_shell.v1"; },
    (value) => { value.fixture.fixture_id = "cartulary.perf.evidence_heavy.v1"; },
    (value) => { value.fixture.object_count = 9999; },
    (value) => { value.fixture.object_bytes = 0; },
    (value) => { value.fixture.total_object_bytes--; },
    (value) => { value.fixture.stub_blob_bytes = true; },
    (value) => { value.workload.preview_preflight_records = 5; },
    (value) => { value.workload.measured_access_handle_requests = 1; },
    (value) => { value.workload.fresh_metadata_request_each_open = false; },
    (value) => { value.workload.consecutive_frames = 1; },
    (value) => { value.workload.owning_viewport = "grid"; },
    (value) => { value.workload.filter_field_key = "timeline.activity_synopsis_text"; },
    (value) => { value.measured_samples = 99; },
    (value) => { value.threshold_ms = 301; },
    (value) => { value.product_retries = 1; },
    (value) => { delete value.workload; },
  ]) {
    const invalid = structuredClone(contract); mutate(invalid);
    assert.throws(() => validate(invalid));
  }
  const unavailableFilter = structuredClone(view);
  unavailableFilter.fields.find((field) => field.field_key === "timeline.has_evidence").filter_ops = [];
  assert.throws(() => validateEvidenceInspectorCapabilities(contract, unavailableFilter, policy), /admitted view capabilities/u);
  assert.throws(() => validateEvidenceInspectorCapabilities(contract, view, { ...policy, measured_samples: 99 }), /sampling differs/u);
}
{
  const contract=JSON.parse(readFileSync(path.join(root,"contracts/performance/ac045.v2.json"),"utf8"));
  const profile={fixture_profile_id:"evidence_heavy_snapshot_v2"};
  const sample=(index)=>({sample_index:index,warmup:index===0,total_ms:100,accepted_at:index*1000+10,metadata_started_at:index*1000+11,metadata_completed_at:index*1000+50,first_frame_at:index*1000+90,second_frame_at:index*1000+110});
  const proof={schema_id:"cartulary.evidence_inspector_measurement_proof.v1",kind:"evidence_inspector",consecutive_frames:2,all_content_visible:true,bindings_stable:true,fresh_metadata_each_open:true,
    target:{record_id:"00000000-0000-0000-0000-000000000100",row_version:2},
    items:Array.from({length:5},(_,index)=>({record_id:`00000000-0000-0000-0000-${String(index+1).padStart(12,"0")}`,row_version:1})),
    viewport_width:1920,viewport_height:1080,device_scale_factor:1,owning_viewport:"inspector",preview_preflight_records:100,preview_preflight_body_verification:"complete_size_and_sha256",measured_access_handle_requests:0,samples:Array.from({length:101},(_,i)=>sample(i))};
  const observation={schema_id:"cartulary.frontend_measurement_observation.v4",fixture_profile_id:profile.fixture_profile_id,criterion_id:contract.criterion_id,predicate_id:contract.predicate_id,
    fixture_id:contract.fixture.fixture_id,fixture_digest:`sha256:${createHash("sha256").update(canonicalJSONString(contract.fixture)).digest("hex")}`,measurement_policy_id:contract.measurement_policy_id,threshold_ms:contract.threshold_ms,
    warmup_samples:1,measured_samples:100,statistics:{kind:"percentile",percentile:95,p50_ms:100,p95_ms:100},outcome:"passed",
    traffic:{counts:[{fact_id:"analyst_sessions",value:1},{fact_id:"background_sessions",value:0}],rates:[],conditions:[]},
    samples:proof.samples.map(({sample_index,warmup,total_ms})=>({sample_index,warmup,total_ms,stages_ms:[{fact_id:"accepted_to_visible_paint",value:total_ms}]})),predicate_evidence:proof};
  const validate=value=>{validateSchemaSync(value.schema_id,value);validateProfileMeasurementObservation(root,profile,value);};
  validate(observation);
  evidenceInspectorObservation=structuredClone(observation);
  for(const mutate of [
    v=>{v.predicate_id="perf.evidence_inspector.metadata_shell.v1";},
    v=>{v.fixture_id="cartulary.perf.evidence_heavy.v1";},
    v=>{v.fixture_digest=`sha256:${"0".repeat(64)}`;},
    v=>{v.traffic.counts[0].value=25;},
    v=>{v.predicate_evidence=null;},
    v=>{v.predicate_evidence.preview_preflight_records=5;},
    v=>{v.predicate_evidence.measured_access_handle_requests=1;},
    v=>{v.predicate_evidence.fresh_metadata_each_open=false;},
    v=>{v.predicate_evidence.consecutive_frames=1;},
    v=>{v.predicate_evidence.owning_viewport="grid";},
    v=>{v.predicate_evidence.target.row_version=1;},
    v=>{v.predicate_evidence.items[0].row_version=2;},
    v=>{v.predicate_evidence.items[1]=v.predicate_evidence.items[0];},
    v=>{v.predicate_evidence.samples[1].metadata_started_at=1;},
    v=>{v.predicate_evidence.samples[1].metadata_completed_at=99999;},
    v=>{v.predicate_evidence.samples[1].first_frame_at=v.predicate_evidence.samples[1].second_frame_at;},
    v=>{v.predicate_evidence.samples[1].total_ms=99;},
    v=>{v.predicate_evidence.samples[1].warmup=true;},
    v=>{v.samples.pop();},
    v=>{v.statistics.p95_ms=99;},
    v=>{v.outcome="threshold_failed";},
    v=>{v.predicate_evidence.retry_to_green=true;},
  ]) {const invalid=structuredClone(observation);mutate(invalid);assert.throws(()=>validate(invalid));}
  for (const [elapsed,outcome] of [[300,"passed"],[300.01,"threshold_failed"]]) {
    const boundary=structuredClone(observation);boundary.outcome=outcome;boundary.statistics.p50_ms=elapsed;boundary.statistics.p95_ms=elapsed;
    for(const item of boundary.predicate_evidence.samples) {item.second_frame_at=item.accepted_at+elapsed;item.total_ms=item.second_frame_at-item.accepted_at;}
    boundary.samples=boundary.predicate_evidence.samples.map(({sample_index,warmup,total_ms})=>({sample_index,warmup,total_ms,stages_ms:[{fact_id:"accepted_to_visible_paint",value:total_ms}]}));
    const values=boundary.samples.slice(1).map(s=>s.total_ms).sort((a,b)=>a-b);boundary.statistics.p50_ms=values[49];boundary.statistics.p95_ms=values[94];
    validate(boundary);
  }
}
const compiler = new WorkGraphCompiler(root);
const resourceProfiles = new Map(JSON.parse(readFileSync(
  path.join(root, "tools/execution_topology_manifest.json"), "utf8",
)).resource_profiles.map((profile) => [profile.id, profile.resource_claims]));

const fixtureProfile = activePerformanceFixtureProfile(
  loadPerformanceFixtureSnapshotRegistry(root),
  "timeline_large_grid_snapshot_v2",
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
    schema_id: "cartulary.frontend_measurement_observation.v4",
    fixture_profile_id: "timeline_large_grid_snapshot_v2",
    criterion_id: "AC-043",
    predicate_id: "perf.timeline_summary_selection_down.v1",
    fixture_id: "cartulary.perf.large_grid.v1",
    fixture_digest: `sha256:${"a".repeat(64)}`,
    measurement_policy_id: "cartulary.measurement.interactive_p95.v1",
    threshold_ms: 100,
    warmup_samples: 1,
    measured_samples: 100,
    statistics: { kind: "percentile", percentile: 95, p50_ms: 10, p95_ms: 20 },
    predicate_evidence: null,
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
      total_ms: sampleIndex >= 95 ? 20 : 10,
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
              name: `cartulary.frontend_measurement_observation.v4.${summary.predicate_id}`,
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
  const profile=options.profile ?? fixtureProfile;
  const observation = measurementObservation(options.observationOverrides);
  const rowID = options.rowID ??
    "module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3";
  const snapshotKey = "b".repeat(64);
  const builderUnitID = `fixture_snapshot:default:${profile.fixture_profile_id}:${snapshotKey}`;
  const observationArtifact = writeMeasurementArtifact(
    rootDirectory,
    "frontend-measurement-observation.v4.json",
    "frontend_measurement_observation",
    observation,
  );
  const buildArtifact = writeMeasurementArtifact(
    rootDirectory,
    "snapshot-build.json",
    "performance_fixture_snapshot_build",
    {
      schema_id: "cartulary.performance_fixture_snapshot.v3",
      fixture_profile_id: profile.fixture_profile_id,
      fixture_version: profile.fixture_version,
      seed: profile.seed,
      snapshot_key_schema_id: "cartulary.performance_fixture_snapshot_key.v2",
      snapshot_key: snapshotKey,
      migration_digest: "a".repeat(64),
      source_contract_digest: profile.source_contract_digest,
      builder_unit_id: builderUnitID,
      build_ordinal: 1,
      state: "sealed",
 object_snapshot: { manifest_sha256: "sha256:" + "f".repeat(64), object_count: profile.object_population.object_count, total_bytes: profile.object_population.total_bytes },
      contribution_receipts: profile.contributions.map((contribution) => ({
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
        counts: profile.semantic_expectations.counts.map((expectation) => ({
          expectation_id: expectation.expectation_id,
          actual: expectation.exact,
          passed: true,
        })),
        conditions: profile.semantic_expectations.conditions.map((expectation) => ({
          expectation_id: expectation.expectation_id,
          actual: expectation.required,
          passed: true,
        })),
        connections_closed: true,
      },
      redaction_policy_id: profile.redaction_policy.policy_id,
      created_at: "2026-08-14T00:00:00.000Z",
    },
  );
  const leaseArtifact = writeMeasurementArtifact(
    rootDirectory,
    "snapshot-lease.json",
    "performance_fixture_snapshot_lease",
    {
      schema_id: "cartulary.performance_fixture_snapshot_lease.v3",
      fixture_profile_id: profile.fixture_profile_id,
      snapshot_key: snapshotKey,
      builder_unit_id: builderUnitID,
      row_id: rowID,
      predicate_id: observation.predicate_id,
      lease_identity: `sha256:${"d".repeat(64)}`,
      clone_ordinal: 1,
      creation_state: "created",
 object_snapshot: { manifest_sha256: "sha256:" + "f".repeat(64), object_count: profile.object_population.object_count, total_bytes: profile.object_population.total_bytes },
      isolation_result: "isolated",
      cleanup_results: [
        "bucket",
        "credential_copy",
        "database",
        "process",
        "session",
      ].map((resource_class) => ({ resource_class, outcome: "complete" })),
      cleanup_state: "complete",
      redaction_policy_id: "cartulary.performance_fixture_redaction.v2",
      finalized_at: "2026-08-14T00:01:00.000Z",
      ...options.leaseOverrides,
    },
  );
  return {
    schema_id: "cartulary.frontend_measurement_summary.v5",
    row_id: rowID,
    fixture_profile_id: profile.fixture_profile_id,
    snapshot_key: snapshotKey,
    observation_artifact: observationArtifact.ref,
    build_artifact: buildArtifact.ref,
    lease_artifact: leaseArtifact.ref,
    clone_ordinal: 1,
    scheduler_overlap_count: 0,
    rollup: {
      predicate_id: observation.predicate_id,
      sample_count: observation.measured_samples,
      statistics: observation.statistics,
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
  const summary = finalizedMeasurementSummary(measurementEvidenceRoot);
  const summaryArtifact = writeMeasurementArtifact(measurementEvidenceRoot, "inspection-summary.json", "frontend_measurement_summary", summary);
  const aggregateDirectory = path.join(measurementEvidenceRoot, "browser-e2e-measurement");
  mkdirSync(aggregateDirectory, { mode: 0o700 });
  const aggregateFile = path.join(aggregateDirectory, "frontend-measurement-aggregate.json");
  const aggregate = { schema_id: "cartulary.frontend_measurement_aggregate.v5", target_id: "browser-e2e-measurement", status: "qualified", profile_groups: [{
    fixture_profile_id: summary.fixture_profile_id, snapshot_key: summary.snapshot_key, builder_count: 1, clone_count: 1, scheduler_overlap_count: 0,
    build_artifact: summary.build_artifact, summary_artifacts: [summaryArtifact.ref], rollups: [{predicate_id: summary.rollup.predicate_id, sample_count: summary.rollup.sample_count, statistics: summary.rollup.statistics, threshold_ms: summary.rollup.threshold_ms, outcome: summary.rollup.outcome}],
  }] };
  const targetEvidence = {
    schema_id: "cartulary.browser_target_result.v7", target_id: aggregate.target_id, status: "pass", generated_at: "2026-10-14T00:00:00Z",
    group_results: [{ group_id: "measurement", browser_session_id: "measurement", ref: "group.json", sha256: summaryArtifact.ref.sha256 }],
    sessions: [{ browser_session_id: "measurement", runtime_profile_id: "measurement", service_requirement: "test-services", artifacts: [
      { kind: "stack_v7", ref: "stack.json", sha256: summaryArtifact.ref.sha256 },
      { kind: "startup_diagnostics_v2", ref: "startup.json", sha256: summaryArtifact.ref.sha256 },
    ] }],
    artifacts: [{ kind: `frontend_measurement_aggregate_${aggregate.schema_id.split(".").at(-1)}`, ref: "frontend-measurement-aggregate.json", sha256: summaryArtifact.ref.sha256 }],
  };
  validateSchemaSync(targetEvidence.schema_id, targetEvidence);
  assert.throws(() => validateSchemaSync(targetEvidence.schema_id, { ...targetEvidence, schema_id: "cartulary.browser_target_result.v4" }), /validation failed/u);
  assert.throws(() => validateSchemaSync(targetEvidence.schema_id, { ...targetEvidence, artifacts: [{ ...targetEvidence.artifacts[0], kind: "frontend_measurement_aggregate_v3" }] }), /validation failed/u);
  const inspect = () => spawnSync(process.execPath, ["tools/harness/diagnostics/explain-run-cli.mjs", "--results-dir", path.dirname(measurementEvidenceRoot), "--run-id", path.basename(measurementEvidenceRoot), "--detail", "measurement"], { cwd: root, encoding: "utf8", timeout: 30000 });
  writeFileSync(aggregateFile, JSON.stringify(aggregate));
  const current = inspect();
  assert.equal(current.status, 0, current.stderr);
  assert.match(current.stdout, /generation=current/u);
  assert.match(current.stdout, /MEASUREMENT-STAGE/u);
  for (const schema_id of ["cartulary.frontend_measurement_aggregate.v1", "cartulary.frontend_measurement_aggregate.v2", "unsupported-private-value"]) {
    const bytes = JSON.stringify({ schema_id, summaries: [] });
    writeFileSync(aggregateFile, bytes);
    const rejected = inspect();
    assert.equal(rejected.status, 11, rejected.stderr);
    assert.equal(rejected.stdout, "");
    assert.match(rejected.stderr, /^unsupported measurement artifact;/u);
    assert.ok(rejected.stderr.length < 160);
    assert.equal(rejected.stderr.includes(schema_id), false);
    assert.equal(readFileSync(aggregateFile, "utf8"), bytes, "inspection changed retired bytes");
  }
  rmSync(aggregateDirectory, { recursive: true });
  const reportPath = path.join(measurementEvidenceRoot, "playwright-report.json");
  const collect = (summary, expected = [summary.predicate_id]) => {
    writeFileSync(reportPath, `${JSON.stringify(measurementReport(summary))}\n`);
    return collectFrontendMeasurementObservations({
      expectedPredicateIDs: expected,
      observationSchemaID: "cartulary.frontend_measurement_observation.v4",
      reportPaths: [reportPath],
      runRoot: measurementEvidenceRoot,
    });
  };
  assert.equal(collect(measurementObservation()).length, 1);
  assert.throws(() => collect(measurementObservation({ statistics: { kind: "percentile", percentile: 95, p50_ms: 10, p95_ms: 1 } })), /statistics differ/u);
  const retried = measurementReport(measurementObservation());
  retried.suites[0].specs[0].tests[0].results[0].retry = 1;
  writeFileSync(reportPath, JSON.stringify(retried));
  assert.throws(() => collectFrontendMeasurementObservations({ expectedPredicateIDs: ["perf.timeline_summary_selection_down.v1"], observationSchemaID: "cartulary.frontend_measurement_observation.v4", reportPaths: [reportPath], runRoot: measurementEvidenceRoot }), /retried browser measurement/u);
  const duplicateAttachment = measurementReport(measurementObservation());
  const attachment = duplicateAttachment.suites[0].specs[0].tests[0].results[0].attachments[0];
  const originalAttachment = Buffer.from(attachment.body, "base64").toString("utf8");
  attachment.body = Buffer.from(`{"schema_id":"cartulary.frontend_measurement_observation.v4",${originalAttachment.slice(1)}`).toString("base64");
  writeFileSync(reportPath, JSON.stringify(duplicateAttachment));
  assert.throws(() => collectFrontendMeasurementObservations({ expectedPredicateIDs: ["perf.timeline_summary_selection_down.v1"], observationSchemaID: "cartulary.frontend_measurement_observation.v4", reportPaths: [reportPath], runRoot: measurementEvidenceRoot }), /duplicate/u);
  const sharedContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/large-grid.v1.json"), "utf8"));
  const paste = measurementObservation({
    criterion_id: "AC-003", predicate_id: "perf.timeline_paste_20x5.v2", measurement_policy_id: "cartulary.measurement.single_observation.v1",
    fixture_digest: fixtureDigest(sharedContract.fixture), threshold_ms: 2000, warmup_samples: 0, measured_samples: 1,
    statistics: { kind: "single", elapsed_ms: 1999 }, samples: [{ sample_index: 0, warmup: false, total_ms: 1999, stages_ms: [] }],
    predicate_evidence: { kind: "timeline_paste_20x5", clipboard_format: "tsv", clipboard_mime: "text/plain", clipboard_delivery: "synthesized_event", viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1, row_count: 20, cell_count: 100, consecutive_frames: 2, all_new_ids: true, geometry_unchanged: true, target_bindings_stable: true, created_records: Array.from({ length: 20 }, (_, index) => ({ record_id: `00000000-0000-0000-0000-${String(index + 1).padStart(12, "0")}`, row_version: 1 })) },
  });
  const validatePaste = (value) => {
    validateSchemaSync("cartulary.frontend_measurement_observation.v4", value);
    validateProfileMeasurementObservation(root, fixtureProfile, value);
  };
  validatePaste(paste);
  assert.equal(collect(paste).length, 1, "only typed synthetic insertion identities are retained");
  const syntheticProofObservations = [paste];
  for (const mutate of [
    (value) => { value.predicate_evidence.created_records[1].record_id = value.predicate_evidence.created_records[0].record_id; },
    (value) => { value.predicate_evidence.cell_count = 99; },
    (value) => { value.predicate_evidence.row_count = 19; },
    (value) => { value.predicate_evidence.consecutive_frames = 1; },
    (value) => { value.predicate_evidence.target_bindings_stable = false; },
    (value) => { value.predicate_evidence.geometry_unchanged = false; },
    (value) => { value.traffic.conditions[1].value = false; },
    (value) => { value.statistics.elapsed_ms = 10; },
    (value) => { value.samples[0].total_ms = 2000; value.statistics.elapsed_ms = 2000; },
    (value) => { value.predicate_evidence = null; },
    (value) => { value.schema_id = "cartulary.frontend_measurement_observation.v2"; },
  ]) {
    const invalid = structuredClone(paste); mutate(invalid);
    assert.throws(() => validatePaste(invalid));
  }
  const atThreshold = structuredClone(paste);
  atThreshold.samples[0].total_ms = 2000; atThreshold.statistics.elapsed_ms = 2000; atThreshold.outcome = "threshold_failed";
  validatePaste(atThreshold);
  const presenceContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/presence.v2.json"), "utf8"));
  validateSchemaSync("cartulary.performance_presence_contract.v2", presenceContract);
  for (const mutate of [
    value => { value.criterion_ids = ["AC-008"]; },
    value => { value.criterion_ids = ["AC-008", "AC-008"]; },
    value => { value.criterion_ids.reverse(); },
    value => { value.criterion_id = "AC-008"; },
    value => { value.schema_id = "cartulary.performance_presence_contract.v1"; },
  ]) {
    const invalid = structuredClone(presenceContract); mutate(invalid);
    assert.throws(() => validateSchemaSync("cartulary.performance_presence_contract.v2", invalid));
  }
  for (const kind of ["surface", "row", "cell"]) {
    const presence = {
      ...paste, criterion_id: "AC-008", predicate_id: `perf.timeline_presence_${kind}.rendered.v1`, threshold_ms: 1000,
      statistics: { kind: "single", elapsed_ms: 1000 }, samples: [{ sample_index: 0, warmup: false, total_ms: 1000, stages_ms: [] }],
      predicate_evidence: {
        schema_id: "cartulary.presence_measurement_proof.v1", kind: "timeline_presence", presence_kind: kind,
        actor_user_id: "00000000-0000-0000-0000-000000000001", view_schema_id: "cartulary.view.timeline.v2",
        record_id: kind === "surface" ? null : "00000000-0000-0000-0000-000000000002", row_version: kind === "surface" ? null : 1,
        field_key: kind === "cell" ? "timeline.activity_synopsis_text" : null,
        consecutive_frames: 2, clock_domain: "browser_monotonic_time_origin", origins_unchanged: true,
        initially_absent: true, binding_stable: true, action_count: 1,
        viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1,
      },
    };
    const validatePresence = value => {
      validateSchemaSync("cartulary.frontend_measurement_observation.v4", value);
      validateProfileMeasurementObservation(root, fixtureProfile, value);
    };
    validatePresence(presence);
    syntheticProofObservations.push(presence);
    assert.equal(collect(presence).length, 1, "closed presence proof retains only its admitted synthetic target identity");
    assert.throws(() => collect({ ...presence, predicate_evidence: { ...presence.predicate_evidence, payload: "private" } }));
    for (const mutate of [
      value => { value.predicate_id = "perf.presence_delta.rendered.v1"; },
      value => { value.predicate_evidence.origins_unchanged = false; },
      value => { value.predicate_evidence.initially_absent = false; },
      value => { value.predicate_evidence.consecutive_frames = 1; },
      value => { value.predicate_evidence.action_count = 2; },
      value => { value.predicate_evidence.clock_domain = "driver_clock"; },
      value => { value.predicate_evidence.row_version = kind === "surface" ? 1 : null; },
      value => { value.predicate_evidence.field_key = kind === "cell" ? null : "timeline.activity_synopsis_text"; },
      value => { value.predicate_evidence = null; },
      value => { value.samples[0].total_ms = 1001; value.statistics.elapsed_ms = 1001; },
      value => { value.samples[0].warmup = true; },
      value => { value.statistics.elapsed_ms = 1; },
      value => { value.traffic.counts[0].value = 26; },
    ]) {
      const invalid = structuredClone(presence); mutate(invalid);
      assert.throws(() => validatePresence(invalid));
    }
    const slow = structuredClone(presence);
    slow.samples[0].total_ms = 1001; slow.statistics.elapsed_ms = 1001; slow.outcome = "threshold_failed";
    validatePresence(slow);
  }
  const anchorContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/ac047.v2.json"), "utf8"));
  const trace = { ...paste, criterion_id: "AC-047", predicate_id: anchorContract.predicate_id,
    measurement_policy_id: anchorContract.measurement_policy_id, threshold_ms: null, measured_samples: 0, samples: [],
    statistics: {kind: "seeded_trace", completed_steps: anchorContract.steps.length, expected_steps: anchorContract.steps.length},
    predicate_evidence: {schema_id: "cartulary.anchor_trace_proof.v1", kind: "timeline_anchor_trace", record_id: "00000000-0000-0000-0000-000000000001", request_record_id: "00000000-0000-0000-0000-000000000001",
      initial_row_version: 1, committed_row_version: 2, final_row_version: 3, max_mounted_rows: 35, frame_count: 22, grid_replacements: 0, violations: [],
      checkpoints: anchorContract.steps.map((step,index) => ({step, frame_count: 2, row_version: index >= 9 ? 3 : index >= 7 ? 2 : 1, mounted_rows: 35})),
    },
  };
  const validateTrace = value => {validateSchemaSync("cartulary.frontend_measurement_observation.v4",value);validateProfileMeasurementObservation(root,fixtureProfile,value);};
  validateTrace(trace); assert.equal(collect(trace).length,1); syntheticProofObservations.push(trace);
  for (const mutate of [
    value => {value.threshold_ms = 2000;},
    value => {value.statistics.kind = "single";},
    value => {value.statistics.completed_steps = 10;},
    value => {value.statistics.expected_steps = 12;},
    value => {value.predicate_evidence.checkpoints.reverse();},
    value => {value.predicate_evidence.checkpoints[0].frame_count = 1;},
    value => {value.predicate_evidence.frame_count = 1;},
    value => {value.predicate_evidence.request_record_id = null;},
    value => {value.predicate_evidence.request_record_id = "00000000-0000-0000-0000-000000000002";},
    value => {value.predicate_evidence.committed_row_version = 1;},
    value => {value.predicate_evidence.final_row_version = 2;},
    value => {value.predicate_evidence.checkpoints[8].row_version = 1;},
    value => {value.predicate_evidence.max_mounted_rows = 20000;},
    value => {value.predicate_evidence.grid_replacements = -1;},
    value => {value.predicate_evidence.violations.push("draft_changed");},
    value => {value.predicate_evidence = null;},
    value => {value.samples = [{sample_index:0,warmup:false,total_ms:100,stages_ms:[]}];},
    value => {value.traffic.counts[0].value = 24;},
    value => {value.outcome = "threshold_failed";},
    value => {value.schema_id = "cartulary.frontend_measurement_observation.v3";},
  ]) {const invalid=structuredClone(trace);mutate(invalid);assert.throws(() => validateTrace(invalid));}
  const failedTrace=structuredClone(trace);failedTrace.outcome="predicate_failed";failedTrace.predicate_evidence.violations.push("draft_changed");validateTrace(failedTrace);
  for (const purpose of ["discovery", "apply"]) {
  const importContract = JSON.parse(readFileSync(path.join(root, `contracts/performance/import-${purpose}.v1.json`), "utf8"));
  const importProfile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), importContract.fixture_profile_id);
  for (const predicate of importContract.predicates) {
    const discovery = { ...paste, fixture_profile_id: importProfile.fixture_profile_id, criterion_id: "AC-027", predicate_id: predicate.predicate_id, threshold_ms: 1000,
      statistics: {kind: "single", elapsed_ms: 1000}, samples: [{sample_index:0,warmup:false,total_ms:1000,stages_ms:[]}],
      predicate_evidence: {
        schema_id: `cartulary.import_${purpose}_measurement_proof.v1`, kind: `import_${purpose}`, source_format: predicate.kind,
        ...(purpose === "apply" ? { selected_units: 1, mapped_column_ordinal: 1, mapped_field_key: "timeline.activity_synopsis_text", unmapped_column_ordinal: 2 } : {}),
        job_id: "00000000-0000-4000-8000-000000000001", job_status: "running",
        record_id: "00000000-0000-4000-8000-000000000002", row_version: 3,
        view_schema_id: "cartulary.view.timeline.v2", field_key: "timeline.analyst_text",
        data_rows: 20000, columns: 2, header_rows: 1, cancel_enabled: true, grid_edit_accepted: true, draft_uncommitted: true,
        consecutive_frames: 2, binding_stable: true, geometry_unchanged: true, scroll_unchanged: true,
        action_count: 1, viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1,
      },
    };
    const validateDiscovery = value => { validateSchemaSync("cartulary.frontend_measurement_observation.v4", value); validateProfileMeasurementObservation(root, importProfile, value); };
    validateDiscovery(discovery);
    assert.equal(collect(discovery).length, 1);
    syntheticProofObservations.push(discovery);
    for (const mutate of [
      value => { value.fixture_profile_id = fixtureProfile.fixture_profile_id; },
      value => { value.predicate_id = "perf.job_progress.visible_with_cancel.v1"; },
      value => { value.predicate_evidence.kind = purpose === "apply" ? "import_discovery" : "import_apply"; },
      value => { value.predicate_evidence.mapped_column_ordinal = 0; },
      value => { value.predicate_evidence.unknown = true; },
      value => { value.predicate_evidence.source_format = predicate.kind === "csv" ? "xlsx" : "csv"; },
      value => { value.predicate_evidence.job_status = "succeeded"; },
      value => { value.predicate_evidence.cancel_enabled = false; },
      value => { value.predicate_evidence.grid_edit_accepted = false; },
      value => { value.predicate_evidence.draft_uncommitted = false; },
      value => { value.predicate_evidence.data_rows = 1; },
      value => { value.predicate_evidence.columns = 1; },
      value => { value.predicate_evidence.consecutive_frames = 1; },
      value => { value.predicate_evidence.action_count = 2; },
      value => { value.predicate_evidence.scroll_unchanged = false; },
      value => { value.predicate_evidence = null; },
      value => { value.traffic.counts[0].value = 24; },
      value => { value.samples[0].total_ms = 1000.01; value.statistics.elapsed_ms = 1000.01; },
      value => { value.warmup_samples = 1; },
    ]) { const invalid=structuredClone(discovery); mutate(invalid); assert.throws(() => validateDiscovery(invalid)); }
    const slow=structuredClone(discovery); slow.samples[0].total_ms=1000.01; slow.statistics.elapsed_ms=1000.01; slow.outcome="threshold_failed"; validateDiscovery(slow);
  }
  }
  const verificationContract = JSON.parse(readFileSync(path.join(root,"contracts/performance/evidence-verification.v1.json"),"utf8"));
  const verificationProfile=activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root),verificationContract.fixture_profile_id);
  const verificationObservation={...evidenceInspectorObservation,criterion_id:"AC-016",predicate_id:verificationContract.predicate_id,
    measurement_policy_id:verificationContract.measurement_policy_id,threshold_ms:1000,warmup_samples:0,measured_samples:1,
    statistics:{kind:"single",elapsed_ms:1000},samples:[{sample_index:0,warmup:false,total_ms:1000,stages_ms:[]}],
    predicate_evidence:{schema_id:"cartulary.evidence_verification_measurement_proof.v1",kind:"evidence_verification",job_purpose:"evidence_verification",
      job_id:"00000000-0000-4000-8000-000000000001",job_status:"running",object_blob_id:"00000000-0000-4000-8000-000000000002",
      record_id:"00000000-0000-4000-8000-000000000003",source_record_id:"00000000-0000-4000-8000-000000000004",row_version:1,
      view_schema_id:"cartulary.view.timeline.v2",field_key:"timeline.analyst_text",clipboard_driver:"synthesized_file_event",uploaded_bytes:2097152,
      preview_preflight_records:100,inspector_open:false,attachment_details_expanded:false,cancel_enabled:true,grid_edit_accepted:true,draft_uncommitted:true,
      consecutive_frames:2,binding_stable:true,geometry_unchanged:true,scroll_unchanged:true,action_count:1,viewport_width:1920,viewport_height:1080,device_scale_factor:1}};
  const validateVerification=value=>{validateSchemaSync("cartulary.frontend_measurement_observation.v4",value);validateProfileMeasurementObservation(root,verificationProfile,value);};
  validateVerification(verificationObservation);assert.equal(collect(verificationObservation).length,1);syntheticProofObservations.push(verificationObservation);
  for(const mutate of [
    v=>{v.predicate_evidence.source_record_id=v.predicate_evidence.record_id;},v=>{v.predicate_evidence.object_blob_id=null;},
    v=>{v.predicate_evidence.clipboard_driver="platform_clipboard";},v=>{v.predicate_evidence.uploaded_bytes=1;},
    v=>{v.predicate_evidence.preview_preflight_records=99;},v=>{v.predicate_evidence.job_status="succeeded";},
    v=>{v.predicate_evidence.inspector_open=true;},v=>{v.predicate_evidence.attachment_details_expanded=true;},
    v=>{v.predicate_evidence.cancel_enabled=false;},v=>{v.predicate_evidence.grid_edit_accepted=false;},v=>{v.predicate_evidence.consecutive_frames=1;},
    v=>{v.predicate_evidence.action_count=2;},v=>{v.predicate_evidence.scroll_unchanged=false;},v=>{v.predicate_evidence=null;},
    v=>{v.traffic.counts[0].value=25;},v=>{v.samples[0].total_ms=1000.01;v.statistics.elapsed_ms=1000.01;},v=>{v.warmup_samples=1;},
  ]){const invalid=structuredClone(verificationObservation);mutate(invalid);assert.throws(()=>validateVerification(invalid));}
  const slowVerification=structuredClone(verificationObservation);slowVerification.samples[0].total_ms=1000.01;slowVerification.statistics.elapsed_ms=1000.01;slowVerification.outcome="threshold_failed";validateVerification(slowVerification);
  const referenceContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/reference-pack-progress.v1.json"), "utf8"));
  const referenceProfile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), referenceContract.fixture_profile_id);
  const referenceBundle = referencePackBundle({ preparationDate: "2026-10-16" });
  for (const predicate of referenceContract.predicates) {
    const observation = { ...paste, fixture_profile_id: referenceProfile.fixture_profile_id, criterion_id: "AC-033", predicate_id: predicate.predicate_id, threshold_ms: 1000,
      statistics: { kind: "single", elapsed_ms: 1000 }, samples: [{ sample_index: 0, warmup: false, total_ms: 1000, stages_ms: [] }],
      predicate_evidence: { schema_id: "cartulary.reference_pack_progress_measurement_proof.v1", kind: "reference_pack_progress", job_purpose: predicate.kind,
        job_id: "00000000-0000-4000-8000-000000000001", job_status: "running", record_id: "00000000-0000-4000-8000-000000000002", row_version: 3,
        pack_key: referenceBundle.key, pack_version: referenceBundle.version, payload_sha256: referenceBundle.payloadSha256,
        bundle_sha256: referenceBundle.sha256, bundle_bytes: referenceBundle.upload.buffer.length, bundle_preparation_date: referenceBundle.preparationDate,
        grid_location: "opener", companion_consecutive_frames: 2, same_foreground_session: true, ordinary_security: true,
        view_schema_id: "cartulary.view.timeline.v2", field_key: "timeline.analyst_text", cancel_enabled: true, grid_edit_accepted: true, draft_uncommitted: true,
        consecutive_frames: 2, binding_stable: true, geometry_unchanged: true, scroll_unchanged: true, action_count: 1, viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1 },
    };
    const validate = value => { validateSchemaSync("cartulary.frontend_measurement_observation.v4", value); validateProfileMeasurementObservation(root, referenceProfile, value); };
    validate(observation); assert.equal(collect(observation).length, 1); syntheticProofObservations.push(observation);
    for (const mutate of [
      v => {v.predicate_id="perf.job_progress.visible_with_cancel.v1";}, v => {v.predicate_evidence.job_purpose="other";},
      v => {v.predicate_evidence.bundle_sha256="0".repeat(64);}, v => {v.predicate_evidence.bundle_bytes++;},
      v => {v.predicate_evidence.bundle_preparation_date="2026-10-17";}, v => {v.predicate_evidence.bundle_preparation_date="2026-99-99";},
      v => {v.predicate_evidence.payload_sha256="0".repeat(64);}, v => {v.predicate_evidence.pack_version="other";},
      v => {v.predicate_evidence.grid_location="current";}, v => {v.predicate_evidence.companion_consecutive_frames=1;},
      v => {v.predicate_evidence.job_status="succeeded";}, v => {v.predicate_evidence.cancel_enabled=false;},
      v => {v.predicate_evidence.grid_edit_accepted=false;}, v => {v.predicate_evidence.action_count=2;},
      v => {v.predicate_evidence.scroll_unchanged=false;}, v => {v.predicate_evidence.ordinary_security=false;},
      v => {v.predicate_evidence.same_foreground_session=false;}, v => {v.predicate_evidence=null;},
      v => {v.traffic.counts[0].value=26;}, v => {v.samples[0].total_ms=1000.01;v.statistics.elapsed_ms=1000.01;},
    ]) { const invalid=structuredClone(observation); mutate(invalid); assert.throws(()=>validate(invalid)); }
    const slow=structuredClone(observation);slow.samples[0].total_ms=1000.01;slow.statistics.elapsed_ms=1000.01;slow.outcome="threshold_failed";validate(slow);
  }
  const reportingContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/reporting-progress.v1.json"), "utf8"));
  const reportingProfile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), reportingContract.fixture_profile_id);
  for (const predicate of reportingContract.predicates) {
    const observation = { ...paste, fixture_profile_id: reportingProfile.fixture_profile_id, criterion_id: "AC-030", predicate_id: predicate.predicate_id, threshold_ms: 1000,
      statistics: { kind: "single", elapsed_ms: 1000 }, samples: [{ sample_index: 0, warmup: false, total_ms: 1000, stages_ms: [] }],
      predicate_evidence: { schema_id: "cartulary.reporting_progress_measurement_proof.v1", kind: "reporting_progress", job_purpose: predicate.kind,
        source_snapshot_id: predicate.kind === "snapshot" ? null : "00000000-0000-4000-8000-000000000003",
        job_id: "00000000-0000-4000-8000-000000000001", job_status: "running",
        record_id: "00000000-0000-4000-8000-000000000002", row_version: 3,
        view_schema_id: "cartulary.view.timeline.v2", field_key: "timeline.analyst_text", cancel_enabled: true,
        grid_edit_accepted: true, draft_uncommitted: true, consecutive_frames: 2, binding_stable: true, geometry_unchanged: true,
        scroll_unchanged: true, action_count: 1, viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1 },
    };
    const validate = value => { validateSchemaSync("cartulary.frontend_measurement_observation.v4", value); validateProfileMeasurementObservation(root, reportingProfile, value); };
    validate(observation); assert.equal(collect(observation).length, 1); syntheticProofObservations.push(observation);
    for (const mutate of [
      value => { value.predicate_id = "perf.job_progress.visible_with_cancel.v1"; },
      value => { value.predicate_evidence.job_purpose = predicate.kind === "snapshot" ? "report" : "snapshot"; },
      value => { value.predicate_evidence.source_snapshot_id = predicate.kind === "snapshot" ? "00000000-0000-4000-8000-000000000003" : null; },
      value => { value.predicate_evidence.kind = "import_apply"; },
      value => { value.predicate_evidence.unknown = true; },
      value => { value.predicate_evidence.job_status = "succeeded"; },
      value => { value.predicate_evidence.cancel_enabled = false; },
      value => { value.predicate_evidence.grid_edit_accepted = false; },
      value => { value.predicate_evidence.consecutive_frames = 1; },
      value => { value.predicate_evidence.action_count = 2; },
      value => { value.predicate_evidence.scroll_unchanged = false; },
      value => { value.predicate_evidence = null; },
      value => { value.traffic.counts[0].value = 24; },
      value => { value.samples[0].total_ms = 1000.01; value.statistics.elapsed_ms = 1000.01; },
      value => { value.warmup_samples = 1; },
    ]) { const invalid = structuredClone(observation); mutate(invalid); assert.throws(() => validate(invalid)); }
    const slow = structuredClone(observation); slow.samples[0].total_ms = 1000.01; slow.statistics.elapsed_ms = 1000.01; slow.outcome = "threshold_failed"; validate(slow);
  }
  const historyContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/history.v2.json"), "utf8"));
  for (const predicate of historyContract.predicates) {
    const history = { ...paste, criterion_id: "AC-011", predicate_id: predicate.predicate_id,
      predicate_evidence: {
        schema_id: "cartulary.history_measurement_proof.v2", kind: "timeline_history", action_kind: predicate.kind,
        record_id: "00000000-0000-0000-0000-000000000001", actor_user_id: "00000000-0000-0000-0000-000000000002",
        prior_row_version: 3, row_version: 4, view_schema_id: "cartulary.view.timeline.v2", visible_cells: 2,
        consecutive_frames: 2, new_history_entry: true, attribution_verified: true, binding_stable: true,
        geometry_unchanged: true, grid_scroll_unchanged: true, action_driver: "keyboard_enter", inspector_scroll_policy: "product_managed", action_count: 1, viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1,
      },
    };
    const validateHistory = value => { validateSchemaSync("cartulary.frontend_measurement_observation.v4", value); validateProfileMeasurementObservation(root, fixtureProfile, value); };
    validateHistory(history);
    assert.equal(collect(history).length, 1);
    syntheticProofObservations.push(history);
    for (const mutate of [
      value => { value.predicate_id = "perf.rollback_or_row_restore.rendered.v1"; },
      value => { value.predicate_id = value.predicate_id.replace(".v2", ".v1"); },
      value => { value.predicate_evidence.schema_id = "cartulary.history_measurement_proof.v1"; },
      value => { value.predicate_evidence.grid_scroll_unchanged = false; },
      value => { value.predicate_evidence.action_driver = "script_click"; },
      value => { value.predicate_evidence.inspector_scroll_policy = "driver_managed"; },
      value => { value.predicate_evidence.row_version = 3; },
      value => { value.predicate_evidence.row_version = 5; },
      value => { value.predicate_evidence.action_kind = "unknown"; },
      value => { value.predicate_evidence.new_history_entry = false; },
      value => { value.predicate_evidence.attribution_verified = false; },
      value => { value.predicate_evidence.visible_cells = 1; },
      value => { value.predicate_evidence.consecutive_frames = 1; },
      value => { value.predicate_evidence.geometry_unchanged = false; },
      value => { value.predicate_evidence.action_count = 2; },
      value => { value.samples[0].total_ms = 2000; value.statistics.elapsed_ms = 2000; },
      value => { value.warmup_samples = 1; },
      value => { value.predicate_evidence = null; },
    ]) { const invalid = structuredClone(history); mutate(invalid); assert.throws(() => validateHistory(invalid)); }
    const slow = structuredClone(history);
    slow.samples[0].total_ms = 2000; slow.statistics.elapsed_ms = 2000; slow.outcome = "threshold_failed";
    validateHistory(slow);
  }
  const viewChangeContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/ac044.v1.json"), "utf8"));
  const viewSchema = JSON.parse(readFileSync(path.join(root, "contracts/view-schemas/cartulary.view.timeline.v2.json"), "utf8"));
  validateViewChangeCapabilities(viewChangeContract, viewSchema);
  for (const mutate of [
    value => { value.workloads.filter.field_key = "timeline.activity_synopsis_text"; },
    value => { value.workloads.filter.operator = "contains"; },
    value => { value.workloads.group.field_key = "timeline.activity_synopsis_text"; },
    value => { value.workloads.sort.field_key = "timeline.tags"; },
  ]) {
    const invalid = structuredClone(viewChangeContract); mutate(invalid);
    assert.throws(() => validateViewChangeCapabilities(invalid, viewSchema), /admitted view capabilities/u);
  }
  const percentile = measurementObservation({ fixture_digest: fixtureDigest(sharedContract.fixture) });
  validateProfileMeasurementObservation(root, fixtureProfile, percentile);
  for (const operation of ["sort", "filter", "group"]) {
    for (const completion of ["first_useful_viewport", "stable_viewport"]) {
      const viewChange = { ...percentile, criterion_id: "AC-044", predicate_id: `perf.timeline_${operation}.${completion}.v1`, threshold_ms: completion === "stable_viewport" ? 1000 : 250 };
      validateProfileMeasurementObservation(root, fixtureProfile, viewChange);
      for (const mutate of [
        value => { value.criterion_id = "AC-043"; },
        value => { value.threshold_ms++; },
        value => { value.predicate_id = "perf.view_change.first_useful_viewport.v1"; },
        value => { value.measured_samples = 99; value.samples.pop(); },
        value => { value.statistics.p95_ms = 1; },
      ]) {
        const invalid = structuredClone(viewChange); mutate(invalid);
        assert.throws(() => validateProfileMeasurementObservation(root, fixtureProfile, invalid));
      }
    }
  }
  assert.throws(() => validateProfileMeasurementObservation(root, fixtureProfile, { ...percentile, outcome: "threshold_failed" }), /outcome differ/u);
  assert.equal(
    collect(measurementObservation({ outcome: "threshold_failed", threshold_ms: 19 }))[0].outcome,
    "threshold_failed",
    "threshold failures must remain valid diagnostic evidence",
  );
  assert.equal(
    collect(measurementObservation({
      failure_reason: "setup failed",
      measured_samples: 0,
      outcome: "incomplete",
      statistics: { kind: "percentile", percentile: 95, p50_ms: null, p95_ms: null },
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
      statistics: { kind: "percentile", percentile: 95, p50_ms: null, p95_ms: null },
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
    "frontend-measurement-summary.v5.json",
  );
  const collectWrittenSummary = (summary, validateObservation) => {
    writeFileSync(finalizedPath, `${JSON.stringify(summary)}\n`);
    return collectFinalizedMeasurementSummaries({
      buildSchemaID: "cartulary.performance_fixture_snapshot.v3",
      objectPopulation: compiler.catalog.fixtureProfiles.profiles.get(summary.fixture_profile_id)?.object_population,
      expectedPredicateIDs: [summary.rollup.predicate_id],
      leaseSchemaID: "cartulary.performance_fixture_snapshot_lease.v3",
      observationSchemaID: "cartulary.frontend_measurement_observation.v4",
      runRoot: measurementEvidenceRoot,
      summarySchemaID: "cartulary.frontend_measurement_summary.v5",
      summaryPaths: [finalizedPath],
      validateObservation,
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
  assert.equal(collect(evidenceInspectorObservation).length,1,"collection admits only the exact Evidence target and item identities");
  const evidenceProfile=activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root),"evidence_heavy_snapshot_v2");
  const evidenceSummary=finalizedMeasurementSummary(measurementEvidenceRoot,{profile:evidenceProfile,rowID:"module.timeline.measurement.evidence_inspector_v2",observationOverrides:evidenceInspectorObservation});
  assert.equal(collectWrittenSummary(evidenceSummary,value=>validateProfileMeasurementObservation(root,evidenceProfile,value)).length,1);
  for(const extra of [{record_id:"00000000-0000-0000-0000-000000000001"},{payload:"private"}]) {
    assert.throws(()=>collect({...evidenceInspectorObservation,...extra}),/additional properties/u);
    assert.throws(()=>collect({...evidenceInspectorObservation,predicate_evidence:{...evidenceInspectorObservation.predicate_evidence,...extra}}),/validation failed/u);
  }

  for (const observation of syntheticProofObservations) {
    const profile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), observation.fixture_profile_id);
    const context = { profile, ...(observation.predicate_evidence.kind === "reference_pack_progress" ? {rowID: `module.reference_data.measurement.${observation.predicate_evidence.job_purpose}_progress`} : {}), ...(observation.predicate_evidence.kind === "evidence_verification" ? {rowID:"module.evidence.measurement.verification_progress"} : {}), ...(observation.predicate_evidence.kind === "reporting_progress" ? { rowID: `module.reporting.measurement.${observation.predicate_evidence.job_purpose === "snapshot" ? "snapshot" : "render"}_progress` } : {}), ...(["import_discovery", "import_apply"].includes(observation.predicate_evidence.kind) ? { rowID: `module.imports.measurement.${observation.predicate_evidence.source_format}_${observation.predicate_evidence.kind === "import_apply" ? "apply" : "discovery"}` } : {}) };
    const validateObservation = value => validateProfileMeasurementObservation(root, profile, value);
    const retained = finalizedMeasurementSummary(measurementEvidenceRoot, { ...context, observationOverrides: observation });
    const [qualified] = collectWrittenSummary(retained, validateObservation);
    assert.equal(qualified.summary.rollup.predicate_id, observation.predicate_id,
      "retained finalization must admit the same closed synthetic proof as collection");
    for (const extra of [{ payload: "private" }, { record_id: "00000000-0000-0000-0000-000000000001" }]) {
      const invalid = finalizedMeasurementSummary(measurementEvidenceRoot, {
        ...context, observationOverrides: { ...observation, ...extra },
      });
      assert.throws(() => collectWrittenSummary(invalid, validateObservation),
        /additional properties/u, "admitting a proof never admits sensitive top-level fields");
    }
    const invalidProof = finalizedMeasurementSummary(measurementEvidenceRoot, {
      ...context, observationOverrides: { ...observation, predicate_evidence: { ...observation.predicate_evidence, payload: "private" } },
    });
    assert.throws(() => collectWrittenSummary(invalidProof, validateObservation), /validation failed/u);
  }
  assert.throws(
    () =>
      collectFinalized({
        observationOverrides: {
            failure_reason: "setup failed",
            measured_samples: 0,
            outcome: "incomplete",
            statistics: { kind: "percentile", percentile: 95, p50_ms: null, p95_ms: null },
            warmup_samples: 0,
            samples: [],
        },
        summaryOverrides: {
          rollup: {
            predicate_id: "perf.timeline_summary_selection_down.v1",
            sample_count: 0,
            statistics: { kind: "percentile", percentile: 95, p50_ms: null, p95_ms: null },
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
        buildSchemaID: "cartulary.performance_fixture_snapshot.v3",
      objectPopulation: fixtureProfile.object_population,
        expectedPredicateIDs: [
          "perf.timeline_summary_selection_down.v1",
        ],
        leaseSchemaID: "cartulary.performance_fixture_snapshot_lease.v3",
        observationSchemaID: "cartulary.frontend_measurement_observation.v4",
        runRoot: measurementEvidenceRoot,
        summarySchemaID: "cartulary.frontend_measurement_summary.v5",
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
    /cartulary.frontend_measurement_summary.v5/u,
    "historical summary v1 must not qualify current source",
  );
  for (const object_snapshot of [null, undefined, { manifest_sha256: "sha256:" + "a".repeat(64), object_count: 0, total_bytes: 0 }, { manifest_sha256: "sha256:" + "f".repeat(64), object_count: 1, total_bytes: 12 }]) {
    assert.throws(() => collectFinalized({ leaseOverrides: { object_snapshot } }), /object_snapshot|provenance/u, "missing or mismatched object snapshot proof must not qualify");
  }
  for (const role of ["observation_artifact", "build_artifact", "lease_artifact"]) {
    const duplicateSummary = finalizedMeasurementSummary(measurementEvidenceRoot);
    const reference = duplicateSummary[role];
    const artifactPath = path.join(measurementEvidenceRoot, reference.path);
    const original = JSON.parse(readFileSync(artifactPath, "utf8"));
    const bytes = Buffer.from(`{"schema_id":${JSON.stringify(original.schema_id)},${JSON.stringify(original).slice(1)}`);
    writeFileSync(artifactPath, bytes);
    reference.sha256 = `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
    assert.throws(() => collectWrittenSummary(duplicateSummary), /duplicate/u, "digest-valid duplicate members cannot qualify evidence");
  }
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
      schema_id: "cartulary.harness_unit_event.v4",
      seq,
      monotonic_ms: seq,
      resource_claims: {},
      needs: [],
      service_dependencies: [],
      status: "running",
      ...(event.event === "queued" ? { unit_kind: "runner", row_ids: [] } : {}),
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
    { file: eventFile },
    measurementGroupResult,
  );

  const previousPrefix = process.env.CARTULARY_HARNESS_UNIT_EVENT_PREFIX;
  const previousRunID = process.env.CARTULARY_TEST_RUN_ID;
  const liveEventFile = path.join(schedulerStreamRoot, ".opaque-publisher-stream");
  writeFileSync(liveEventFile, readFileSync(eventFile), { mode: 0o600 });
  const stat = lstatSync(liveEventFile, { bigint: true });
  const prefix = { schema_id: "cartulary.harness_unit_event_prefix.v1", run_id: "prefix-fixture",
    file: liveEventFile, device: stat.dev.toString(), inode: stat.ino.toString(),
    byte_length: Number(stat.size), last_seq: seq };
  const configure = (change = {}) => { process.env.CARTULARY_HARNESS_UNIT_EVENT_PREFIX = JSON.stringify({ ...prefix, ...change }); };
  try {
    process.env.CARTULARY_TEST_RUN_ID = prefix.run_id;
    configure();
    assert.deepEqual(currentUnitEventSource(schedulerStreamRoot), { file: liveEventFile, prefix });
    writeFileSync(liveEventFile, '{"partial":', { flag: "a" });
    assert.deepEqual(await readMeasurementSchedulerEvidence(currentUnitEventSource(schedulerStreamRoot), measurementGroupResult), streamedEvidence);
    await assert.rejects(readMeasurementSchedulerEvidence({ file: liveEventFile }, measurementGroupResult), /invalid JSON/u);
    configure({ run_id: "foreign" });
    assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /another run/u);
    configure();
    chmodSync(liveEventFile, 0o644);
    assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /private regular file/u);
    chmodSync(liveEventFile, 0o600);
    const alias = path.join(schedulerStreamRoot, "stream-alias");
    linkSync(liveEventFile, alias);
    assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /private regular file/u);
    rmSync(alias);
    symlinkSync(liveEventFile, alias);
    configure({ file: alias });
    assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /private regular file/u);
    rmSync(alias);
    configure({ file: path.join(schedulerStreamRoot, "nested", "stream") });
    assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /escapes its run/u);
    configure({ file: "relative-stream" });
    assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /must be absolute/u);
    configure({ file: path.join(os.tmpdir(), ".foreign-stream") });
    assert.throws(() => currentUnitEventSource(schedulerStreamRoot), /escapes its run/u);
    delete process.env.CARTULARY_HARNESS_UNIT_EVENT_PREFIX;
    assert.deepEqual(currentUnitEventSource(schedulerStreamRoot), { file: eventFile });
  } finally {
    for (const [key, value] of [["CARTULARY_HARNESS_UNIT_EVENT_PREFIX", previousPrefix], ["CARTULARY_TEST_RUN_ID", previousRunID]]) {
      if (value === undefined) delete process.env[key]; else process.env[key] = value;
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
    schema_id: "cartulary.harness_unit_event.v4",
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
    (await readMeasurementSchedulerEvidence({ file: overlapFile }, measurementGroupResult))
      .overlap_count,
    1,
    "ordinary work beginning inside a measurement session must disqualify it",
  );

  const incompleteFile = path.join(schedulerStreamRoot, "incomplete-events.ndjson");
  writeFileSync(incompleteFile, `${JSON.stringify(overlapEvents[0])}\n`);
  await assert.rejects(
    readMeasurementSchedulerEvidence({ file: incompleteFile }, measurementGroupResult),
    /lacks a closed scheduler interval/u,
    "incomplete scheduler proof must fail closed",
  );

  const skippedFile = path.join(schedulerStreamRoot, "skipped-events.ndjson");
  writeFileSync(
    skippedFile,
    `${JSON.stringify({
      schema_id: "cartulary.harness_unit_event.v4",
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
    { file: skippedFile },
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
      schema_id: "cartulary.harness_unit_event.v4",
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
    readMeasurementSchedulerEvidence({ file: invalidSequenceFile }, measurementGroupResult),
    /sequence 2 is not contiguous/u,
  );
} finally {
  rmSync(schedulerStreamRoot, { force: true, recursive: true });
}
for (const target of [
  "browser-e2e-webserver-backed",
  "browser-e2e-stateful",
  "browser-e2e-measurement",
  "browser-e2e-a11y",
  "browser-e2e-visual",
]) {
  const suite = ({ "browser-e2e-webserver-backed": "functional", "browser-e2e-a11y": "accessibility" })[target] ?? target.slice("browser-e2e-".length);
  const graph = compiler.compileCommand("browser-e2e", { SUITE: suite }).graph;
  assert.ok(graph.units.some((unit) => unit.unit_id.startsWith("browser_group:")), `${target} must expose browser groups as units`);
  assert.ok(graph.units.some((unit) => unit.unit_id.startsWith("browser_target_summary:")), `${target} must expose its summary projection as a unit`);
  assert.equal(graph.units.some((unit) => unit.command.args.some((arg) => arg.includes("run-browser-e2e-batch"))), false);
}
const visualGraph = compiler.compile({
  kind: "target",
  target: "harness.suite_browser_visual",
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
for (const unit of visualGraph.units.filter((entry) =>
  entry.unit_id.startsWith("browser_group:visual:"),
)) {
  assert.ok(
    unit.current_run_evidence_outputs.some((output) =>
      output.endsWith("/renderer-profile-attestation.json"),
    ),
    `${unit.unit_id} must retain renderer attestation`,
  );
}
const visualUpdateGraph = compiler.compile({
  kind: "target",
  target: "browser-e2e-visual-update",
});
const visualUpdateFinalizer = visualUpdateGraph.units.find(
  (unit) => unit.unit_id === "browser_target_summary:browser-e2e-visual-update",
);
assert.ok(
  visualUpdateFinalizer?.current_run_evidence_outputs.includes(
    "browser-e2e-visual-update/frontend-visual-reconciliation.json",
  ),
  "visual update must reconcile its complete scratch candidate before promotion",
);
const rendererProfile = loadVisualRendererProfile(root);
assert.equal(rendererProfile.playwright_version, JSON.parse(readFileSync(path.join(root, "tools/toolchain_pins.json"), "utf8")).ui_review.playwright);
assert.throws(
  () =>
    assertVisualRendererEnvironmentIsPrivate({
      CARTULARY_VISUAL_RENDERER_WS_ENDPOINT: "ws://127.0.0.1:1/host-input",
    }),
  /harness-private/u,
  "visual renderer endpoints must not be inherited user inputs",
);
assert.doesNotThrow(() => assertVisualRendererEnvironmentIsPrivate({}));
const currentGoldenManifest = buildFrontendVisualGoldenManifest({ root });
assert.deepEqual(
  validateFrontendVisualGoldenManifest(currentGoldenManifest, { root }),
  [],
  "generated golden manifest must exhaustively match committed PNG identities",
);
assert.match(
  validateFrontendVisualGoldenManifest(
    {
      ...currentGoldenManifest,
      goldens: currentGoldenManifest.goldens.map((golden, index) =>
        index === 0 ? { ...golden, sha256: "0".repeat(64) } : golden,
      ),
    },
    { root },
  ).join("\n"),
  /do not match committed PNGs/u,
  "edited PNG identity must fail golden-manifest validation",
);
const promotionRoot = mkdtempSync(
  path.join(os.tmpdir(), "cartulary-visual-promotion."),
);
try {
  const sourceSnapshots = path.join(promotionRoot, "source-snapshots");
  const candidateSnapshots = path.join(promotionRoot, "candidate-snapshots");
  const sourceManifest = path.join(promotionRoot, "source-manifest.json");
  const candidateManifest = path.join(promotionRoot, "candidate-manifest.json");
  const snapshotBackup = path.join(promotionRoot, "snapshot-backup");
  const manifestBackup = path.join(promotionRoot, "manifest-backup.json");
  mkdirSync(sourceSnapshots);
  writeFileSync(path.join(sourceSnapshots, "fixture.png"), "old-golden", { mode: 0o644 });
  chmodSync(path.join(sourceSnapshots, "fixture.png"), 0o644);
  stageVisualSnapshotCandidate(sourceSnapshots, candidateSnapshots);
  assert.equal(lstatSync(path.join(candidateSnapshots, "fixture.png")).mode & 0o777, 0o600);
  assert.equal(readFileSync(path.join(candidateSnapshots, "fixture.png"), "utf8"), "old-golden");
  stageVisualSnapshotCandidate(sourceSnapshots, candidateSnapshots);
  writeFileSync(path.join(candidateSnapshots, "fixture.png"), "new-golden");
  writeFileSync(sourceManifest, "old-manifest");
  writeFileSync(candidateManifest, "new-manifest");
  let renameCount = 0;
  assert.throws(
    () =>
      promoteVisualSnapshotCandidate(
        {
          sourceSnapshots,
          candidateSnapshots,
          sourceManifest,
          candidateManifest,
          snapshotBackup,
          manifestBackup,
        },
        {
          existsSync,
          renameSync(from, to) {
            renameCount += 1;
            if (renameCount === 4) throw new Error("injected manifest promotion failure");
            renameSync(from, to);
          },
          rmSync,
        },
      ),
    /injected manifest promotion failure/u,
  );
  assert.equal(
    readFileSync(path.join(sourceSnapshots, "fixture.png"), "utf8"),
    "old-golden",
  );
  assert.equal(readFileSync(sourceManifest, "utf8"), "old-manifest");
  assert.equal(
    readFileSync(path.join(candidateSnapshots, "fixture.png"), "utf8"),
    "new-golden",
  );
  assert.equal(readFileSync(candidateManifest, "utf8"), "new-manifest");
  assert.equal(existsSync(snapshotBackup), false);
  assert.equal(existsSync(manifestBackup), false);

  promoteVisualSnapshotCandidate({
    sourceSnapshots,
    candidateSnapshots,
    sourceManifest,
    candidateManifest,
    snapshotBackup,
    manifestBackup,
  });
  assert.equal(
    readFileSync(path.join(sourceSnapshots, "fixture.png"), "utf8"),
    "new-golden",
  );
  assert.equal(readFileSync(sourceManifest, "utf8"), "new-manifest");
  assert.equal(existsSync(snapshotBackup), false);
  assert.equal(existsSync(manifestBackup), false);
} finally {
  rmSync(promotionRoot, { force: true, recursive: true });
}
const visualFinalizerSource = readFileSync(
  path.join(root, "tools/harness/browser/browser-target-finalizer-cli.mjs"),
  "utf8",
);
assert.ok(
  visualFinalizerSource.lastIndexOf("secureWriteFile(output") <
    visualFinalizerSource.lastIndexOf("promoteVisualSnapshotCandidate"),
  "visual promotion must be the final fallible target operation after schema and artifact writes",
);
const stateful = compiler.compile({ kind: "target", target: "harness.suite_browser_stateful" });
assert.ok(stateful.units.some((unit) => unit.unit_id.startsWith("browser_reset:")), "stateful browser work must expose resets");

function assertResetTargetMatchesEvidence(graph, label) {
  const resets = graph.units.filter((unit) => unit.unit_id.startsWith("browser_reset:"));
  assert.ok(resets.length > 0, `${label} must contain a reset unit`);
  for (const reset of resets) {
    assert.equal(reset.current_run_evidence_outputs.length, 2);
    assert.ok(reset.current_run_evidence_outputs.includes(
      `unit-results/${reset.unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-")}.json`,
    ));
    const attempt = reset.current_run_evidence_outputs.find((output) => output.includes("/reset-boundary/"));
    assert.ok(attempt?.endsWith(".attempt.json"), `${reset.unit_id} must declare its reset attempt`);
    assert.equal(
      reset.command.environment.CARTULARY_TEST_TARGET,
      attempt.split("/", 1)[0],
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
const singleNetworkFlowGroup = compiler.compile({
  kind: "rows",
  row_ids: ["module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle"],
});
assert.equal(
  singleNetworkFlowGroup.units.some((unit) => unit.unit_id.startsWith("browser_reset:stateful:")),
  false,
  "a first selected stateful group must not receive a reset inherited from the unfiltered graph",
);
const networkFlowStatefulGroups = networkFlowStatefulOwner.units.filter((unit) => unit.unit_id.startsWith("browser_group:stateful:"));
assert.equal(
  networkFlowStatefulOwner.units.filter((unit) => unit.unit_id.startsWith("browser_reset:stateful:")).length,
  networkFlowStatefulGroups.length - new Set(networkFlowStatefulGroups.map((unit) => unit.affinity_key)).size,
  "owner selection must reset exactly between selected groups sharing an affinity",
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
  assert.equal(unit.resource_claims.postgres,
    resourceProfiles.get(unit.command.environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID).postgres,
    `${unit.unit_id} must reserve its authored browser Postgres connection budget`);
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
  target: "harness.suite_browser_functional",
});
const functionalContracts = new Map(JSON.parse(readFileSync(
  path.join(root, "tools/browser_e2e_batch_manifest.json"), "utf8",
)).stages.find((stage) => stage.name === "webserver-backed").groups.map(
  (group) => [group.name, group.browser_session_group],
));
const functionalGroups = webserverBacked.units.filter((unit) =>
  unit.unit_id.startsWith("browser_group:webserver-backed:"),
);
const functionalResets = webserverBacked.units.filter((unit) =>
  unit.unit_id.startsWith("browser_reset:webserver-backed:"),
);
const functionalAffinityCount = Math.min(4, functionalGroups.length);
assert.equal(resourceProfiles.get("browser_functional").postgres, 2, "ordinary functional work reserves the adopted two PostgreSQL tokens");
assert.equal(
  webserverBacked.units.some((unit) => unit.unit_id.startsWith("browser_lifecycle:")),
  false,
  "the first executable browser consumer must own its stack without a synthetic readiness handoff",
);
const functionalRows = compiler.catalog.rows.filter((row) =>
  row.runner === "playwright" && row.selector.stage === "webserver_backed",
);
assert.equal(functionalGroups.length, new Set(functionalRows.map((row) =>
  `${row.runtime_profile_id}:${row.selector.file}`,
)).size);
assert.deepEqual(
  functionalGroups.flatMap((unit) =>
    unit.current_run_evidence_outputs.filter((output) => output.startsWith("rows/")),
  ).sort(),
  functionalRows.map((row) => `rows/${row.row_id}.json`).sort(),
  "functional lane graph must preserve the exact current row closure",
);
assert.equal(
  new Set(functionalGroups.map((unit) => unit.affinity_key)).size,
  functionalAffinityCount,
  "functional groups must use at most four distinct runtime-bound lanes",
);
assert.equal(
  functionalResets.length,
  functionalGroups.length - functionalAffinityCount,
  "each selected browser affinity must reset only between adjacent groups",
);
for (const unit of [...functionalResets, ...functionalGroups]) {
  assert.equal(unit.resource_claims.postgres,
    resourceProfiles.get(unit.command.environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID).postgres,
    `${unit.unit_id} must claim its authored profile's PostgreSQL tokens`);
  assert.equal(unit.resource_claims.browser_stack, 1);
  assert.equal(unit.resource_claims.port_lane, 1);
}
for (const unit of functionalGroups) {
  assert.equal(unit.command.environment.CARTULARY_BROWSER_SESSION_CONTRACT,
    functionalContracts.get(unit.unit_id.split(":").at(-1)),
    "lane isolation must preserve the authored attachment contract");
  assert.notEqual(unit.command.environment.CARTULARY_BROWSER_SESSION_CONTRACT,
    unit.command.environment.CARTULARY_BROWSER_FUNCTIONAL_LANE_ID);
  assert.equal(unit.command.environment.CARTULARY_BROWSER_RESOURCE_PROFILE_ID, "browser_functional");
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
  assert.equal(reset.failure_policy.block_descendants, false);
  assert.equal(reset.failure_policy.aggregate_effect, "required");
  assert.equal(reset.command.args.includes("--renew-generation"), false);
  assert.match(reset.command.environment.CARTULARY_BROWSER_FUNCTIONAL_LANE_ID, /^webserver-backed-/u);
}

const laneRow = functionalRows.find((row) => row.resource_profile_id === "browser_functional");
const laneSelection = compiler.compile({ kind: "rows", row_ids: [laneRow.row_id] });
const laneGroup = laneSelection.units.find((unit) => unit.unit_id.startsWith("browser_group:webserver-backed:"));
assert.match(laneGroup.command.environment.CARTULARY_BROWSER_FUNCTIONAL_LANE_ID, /^webserver-backed-/u);
assert.equal(laneGroup.command.environment.CARTULARY_BROWSER_GROUP_GENERATION, "1");
assert.equal(laneGroup.estimated_work_ms, compiler.owner.evidence_estimates_ms[laneRow.evidence_class]);
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

// The compiler must preserve failure continuation; actual execution, no retry,
// taint and fresh allocation are tested through the native runtime owner.
for (const group of functionalGroups) assert.equal(group.failure_policy.block_descendants, false);
for (const reset of functionalResets) assert.equal(reset.failure_policy.block_descendants, false);

const measurement = compiler.compile({
  kind: "target",
  target: "harness.suite_browser_measurement",
});
const snapshotBuilders = measurement.units.filter((entry) =>
  entry.unit_id.startsWith("fixture_snapshot:default:timeline_large_grid_snapshot_v2:"),
);
assert.equal(snapshotBuilders.length, 1, "measurement rows sharing the large-grid profile use one builder");
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
  `unit-results/${snapshotBuilders[0].unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-")}.json`,
]);
const activeFixtureProfiles = [...compiler.catalog.fixtureProfiles.profiles.values()].filter(profile=>profile.status === "active");
assert.deepEqual(measurement.units.filter(unit=>unit.kind === "fixture_builder").map(unit=>unit.fixture_profile_id).sort(),activeFixtureProfiles.map(profile=>profile.fixture_profile_id).sort());
for (const profile of activeFixtureProfiles) {
  const profiledRowIDs=compiler.catalog.rows.filter(row=>row.fixture_profile_id===profile.fixture_profile_id).map(row=>row.row_id).sort();
  assert.ok(profiledRowIDs.length>0);
  const fixtureIdentity=graph=>{
    const builders=graph.units.filter(unit=>unit.kind === "fixture_builder" && unit.fixture_profile_id===profile.fixture_profile_id);
    assert.equal(builders.length,1,"profiled route must resolve exactly one builder for each selected profile");
    return {builder_unit_id:builders[0].unit_id,fixture_profile_id:builders[0].fixture_profile_id,snapshot_key:builders[0].snapshot_key};
  };
  const expectedFixtureIdentity=fixtureIdentity(measurement);
  const owners=[...new Set(compiler.catalog.rows.filter(row=>row.fixture_profile_id===profile.fixture_profile_id).map(row=>row.owner_id))];
  for (const graph of [
    compiler.compile({kind:"rows",row_ids:[profiledRowIDs[0]]}),
    compiler.compile({kind:"rows",row_ids:profiledRowIDs}),
    ...owners.map(owner_id=>compiler.compile({kind:"owner",owner_id})),
    compiler.compile({kind:"aggregate",target:"release-check"}),
  ]) assert.deepEqual(fixtureIdentity(graph),expectedFixtureIdentity,"direct, row, owner and release routes must preserve each fixture identity");
}
const measurementSummaries = measurement.units.filter((entry) =>
  entry.unit_id.startsWith("browser_measurement_summary:measurement:"),
);
assert.equal(
  measurementSummaries.length,
  activeFixtureProfiles.reduce((count,profile)=>count+profile.verification_bindings.length,0),
  "each admitted measurement row has one post-cleanup summary finalizer",
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
    /frontend-measurement-summary\.v5\.json$/u,
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
	assert.equal(firstConsumer.fixture_profile_id, "timeline_large_grid_snapshot_v2");
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
  target: "harness.suite_browser_functional",
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

const validCapturePayload = {
  schema_id: "cartulary.frontend_visual_capture_intent.v2", capture_id: "visual.capture." + "a".repeat(20),
  capture_intent: "test", expected_golden_path: "apps/web/e2e/workbook.visual.spec.ts-snapshots/test.png",
  project_id: "chromium", renderer_profile_id: rendererProfile.profile_id, screenshot_assertion_location: "test",
  test_file: "apps/web/e2e/workbook.visual.spec.ts", test_title: "test",
  capture_profile: { browser_zoom_percent: 100, color_scheme: "dark", device_scale_factor: 1, reduced_motion: true, project_id: "chromium", snapshot_path_template: "{snapshotDir}/{testFileDir}/{testFileName}-snapshots/{arg}{-snapshotSuffix}{ext}", snapshot_suffix: "linux", viewport_css_px: "1280x720", surface_kind: "workbook_shell", density_id: "compact", expected_density_id: "compact", theme_id: "dark_graphite", expected_theme_id: "dark_graphite" },
};
validateCapturePayload(validCapturePayload);
for (const density of [undefined, null, "", "unknown", "comfortable"]) {
  assert.throws(() => validateCapturePayload({ ...validCapturePayload, capture_profile: { ...validCapturePayload.capture_profile, density_id: density } }));
}
validateCapturePayload({ ...validCapturePayload, capture_profile: { ...validCapturePayload.capture_profile, surface_kind: "application_shell", density_id: null, expected_density_id: null } });
assert.throws(() => validateCapturePayload({ ...validCapturePayload, schema_id: "cartulary.frontend_visual_capture_intent.v1" }));
