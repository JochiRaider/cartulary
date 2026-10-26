import { buildEvidenceInspectorObservation } from "../../performance-fixture/tests/measurement-observations.mjs";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { collectFinalizedMeasurementSummaries, collectFrontendMeasurementObservations } from "../frontend-measurement-evidence.mjs";
import { activePerformanceFixtureProfile, loadPerformanceFixtureSnapshotRegistry } from "../../performance-fixture/index.mjs";
import { validateProfileMeasurementObservation } from "../../performance-fixture/profile-adapters/index.mjs";
import test from "node:test";
import { measurementContext, measurementObservation, measurementReport } from "./measurement-fixtures.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
test("observation collection rejects retries duplicate members missing predicates and incomplete samples", async (t) => {
    const { measurementEvidenceRoot, reportPath, collect } = measurementContext(t);
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
    assert.equal(collect(measurementObservation({ outcome: "threshold_failed", threshold_ms: 19 }))[0].outcome, "threshold_failed", "threshold failures must remain valid diagnostic evidence");
    assert.equal(collect(measurementObservation({
        failure_reason: "setup failed",
        measured_samples: 0,
        outcome: "incomplete",
        statistics: { kind: "percentile", percentile: 95, p50_ms: null, p95_ms: null },
        warmup_samples: 0,
        samples: [],
    }))[0].outcome, "incomplete", "setup failures must retain safe partial evidence");
    assert.throws(() => collect(measurementObservation({
        measured_samples: 99,
        outcome: "incomplete",
        statistics: { kind: "percentile", percentile: 95, p50_ms: null, p95_ms: null },
    })), /sample counts differ/u, "summary cardinality must match its retained samples");
    assert.throws(() => collect(measurementObservation({
        samples: measurementObservation().samples.map((sample, sampleIndex) => sampleIndex === 0
            ? {
                ...sample,
                stages_ms: [
                    { fact_id: "apply_to_visible_paint", value: 1, record_id: 1 },
                ],
            }
            : sample),
    })), /additional properties/u);
    assert.throws(() => collect(measurementObservation(), []), /observations differ/u, "missing expected evidence must fail closed");
});
test("Evidence inspector observation and retained proof reject extra private fields", async (t) => {
    const { collect, collectWrittenSummary, persist } = measurementContext(t);
    const evidenceInspectorObservation = buildEvidenceInspectorObservation();
    assert.equal(collect(evidenceInspectorObservation).length, 1, "collection admits only the exact Evidence target and item identities");
    const evidenceProfile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), "evidence_heavy_snapshot_v2");
    const evidenceSummary = persist({ profile: evidenceProfile, rowID: "module.timeline.measurement.evidence_inspector_v2", observationOverrides: evidenceInspectorObservation });
    assert.equal(collectWrittenSummary(evidenceSummary, value => validateProfileMeasurementObservation(root, evidenceProfile, value)).length, 1);
    for (const extra of [{ record_id: "00000000-0000-0000-0000-000000000001" }, { payload: "private" }]) {
        assert.throws(() => collect({ ...evidenceInspectorObservation, ...extra }), /additional properties/u);
        assert.throws(() => collect({ ...evidenceInspectorObservation, predicate_evidence: { ...evidenceInspectorObservation.predicate_evidence, ...extra } }), /validation failed/u);
    }
});
test("retained measurement admission binds provenance cleanup and current generation", async (t) => {
    const { measurementEvidenceRoot, fixtureProfile, finalizedPath, collectWrittenSummary, persist, collectFinalized } = measurementContext(t);
    assert.equal(collectFinalized().length, 1);
    assert.throws(() => collectFinalized({
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
    }), /not eligible for active qualification/u);
    assert.throws(() => collectFinalized({
        summaryOverrides: {
            qualification_outcome: "threshold_failed",
        },
    }), /inconsistent qualification outcome/u);
    assert.throws(() => collectFinalizedMeasurementSummaries({
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
    }), /paths are duplicated/u);
    assert.throws(() => collectFinalized({
        summaryOverrides: {
            schema_id: "cartulary.frontend_measurement_summary.v1",
        },
    }), /cartulary.frontend_measurement_summary.v5/u, "historical summary v1 must not qualify current source");
    for (const object_snapshot of [null, undefined, { manifest_sha256: "sha256:" + "a".repeat(64), object_count: 0, total_bytes: 0 }, { manifest_sha256: "sha256:" + "f".repeat(64), object_count: 1, total_bytes: 12 }]) {
        assert.throws(() => collectFinalized({ leaseOverrides: { object_snapshot } }), /object_snapshot|provenance/u, "missing or mismatched object snapshot proof must not qualify");
    }
    for (const role of ["observation_artifact", "build_artifact", "lease_artifact"]) {
        const duplicateSummary = persist();
        const reference = duplicateSummary[role];
        const artifactPath = path.join(measurementEvidenceRoot, reference.path);
        const original = JSON.parse(readFileSync(artifactPath, "utf8"));
        const bytes = Buffer.from(`{"schema_id":${JSON.stringify(original.schema_id)},${JSON.stringify(original).slice(1)}`);
        writeFileSync(artifactPath, bytes);
        reference.sha256 = `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
        assert.throws(() => collectWrittenSummary(duplicateSummary), /duplicate/u, "digest-valid duplicate members cannot qualify evidence");
    }
    const tamperedReference = persist();
    tamperedReference.observation_artifact.sha256 = `sha256:${"e".repeat(64)}`;
    assert.throws(() => collectWrittenSummary(tamperedReference), /digest or size is invalid/u, "a mutated referenced artifact digest must fail qualification");
    const wrongRole = persist();
    wrongRole.observation_artifact.role = "diagnostic_observation";
    assert.throws(() => collectWrittenSummary(wrongRole), /artifact reference is malformed/u, "artifact references must retain their exact provenance role");
    assert.throws(() => collectFinalized({
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
    }), /duplicates cleanup class bucket/u, "duplicate cleanup classes must not make lease evidence ambiguous");
});
