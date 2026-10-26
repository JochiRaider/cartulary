import { fixtureDigest } from "../../performance-fixture/profile-adapters/large-grid.mjs";
import { createHash } from "node:crypto";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { collectFinalizedMeasurementSummaries, collectFrontendMeasurementObservations } from "../frontend-measurement-evidence.mjs";
import { activePerformanceFixtureProfile, loadPerformanceFixtureSnapshotRegistry } from "../../performance-fixture/index.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
export function measurementObservation(overrides = {}) {
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
export function measurementReport(summary) {
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
export function artifactDigest(bytes) {
    return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}
function measurementArtifact(files, name, role, value) {
    const bytes = Buffer.from(`${JSON.stringify(value)}\n`);
    files[name] = bytes;
    return { ref: { role, path_kind: "file", format: "json", path: name, sha256: artifactDigest(bytes) }, value };
}
export function buildMeasurementFixture(profile, options = {}) {
    const files = {};
    const observation = measurementObservation(options.observationOverrides);
    const rowID = options.rowID ??
        "module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3";
    const snapshotKey = "b".repeat(64);
    const builderUnitID = `fixture_snapshot:default:${profile.fixture_profile_id}:${snapshotKey}`;
    const observationArtifact = measurementArtifact(files, "frontend-measurement-observation.v4.json", "frontend_measurement_observation", observation);
    const buildArtifact = measurementArtifact(files, "snapshot-build.json", "performance_fixture_snapshot_build", {
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
    });
    const leaseArtifact = measurementArtifact(files, "snapshot-lease.json", "performance_fixture_snapshot_lease", {
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
    });
    const summary = {
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
    return { summary, files };
}
export function persistMeasurementFixture(directory, fixture) {
    for (const [name, bytes] of Object.entries(fixture.files))
        writeFileSync(path.join(directory, name), bytes);
    return fixture.summary;
}
export function writeMeasurementArtifact(directory, name, role, value) {
    const files = {};
    const artifact = measurementArtifact(files, name, role, value);
    for (const [file, bytes] of Object.entries(files))
        writeFileSync(path.join(directory, file), bytes);
    return artifact;
}
export function measurementContext(t) {
    const measurementEvidenceRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-measurement-evidence-"));
    t.after(() => rmSync(measurementEvidenceRoot, { force: true, recursive: true }));
    const fixtureProfile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), "timeline_large_grid_snapshot_v2");
    const reportPath = path.join(measurementEvidenceRoot, "playwright-report.json");
    const collect = (summary, expected = [summary.predicate_id]) => {
        writeFileSync(reportPath, `${JSON.stringify(measurementReport(summary))}\n`);
        return collectFrontendMeasurementObservations({ expectedPredicateIDs: expected,
            observationSchemaID: "cartulary.frontend_measurement_observation.v4", reportPaths: [reportPath], runRoot: measurementEvidenceRoot });
    };
    const finalizedPath = path.join(measurementEvidenceRoot, "frontend-measurement-summary.v5.json");
    const collectWrittenSummary = (summary, validateObservation) => {
        writeFileSync(finalizedPath, `${JSON.stringify(summary)}\n`);
        const profile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), summary.fixture_profile_id);
        return collectFinalizedMeasurementSummaries({ buildSchemaID: "cartulary.performance_fixture_snapshot.v3",
            objectPopulation: profile.object_population, expectedPredicateIDs: [summary.rollup.predicate_id],
            leaseSchemaID: "cartulary.performance_fixture_snapshot_lease.v3", observationSchemaID: "cartulary.frontend_measurement_observation.v4",
            runRoot: measurementEvidenceRoot, summarySchemaID: "cartulary.frontend_measurement_summary.v5", summaryPaths: [finalizedPath], validateObservation });
    };
    const persist = (options = {}) => persistMeasurementFixture(measurementEvidenceRoot, buildMeasurementFixture(options.profile ?? fixtureProfile, options));
    const collectFinalized = (options = {}) => collectWrittenSummary(persist(options));
    return { measurementEvidenceRoot, fixtureProfile, reportPath, finalizedPath, collect, collectWrittenSummary, persist, collectFinalized };
}
export function buildPasteObservation() {
    const sharedContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/large-grid.v1.json"), "utf8"));
    return measurementObservation({
        criterion_id: "AC-003", predicate_id: "perf.timeline_paste_20x5.v2", measurement_policy_id: "cartulary.measurement.single_observation.v1",
        fixture_digest: fixtureDigest(sharedContract.fixture), threshold_ms: 2000, warmup_samples: 0, measured_samples: 1,
        statistics: { kind: "single", elapsed_ms: 1999 }, samples: [{ sample_index: 0, warmup: false, total_ms: 1999, stages_ms: [] }],
        predicate_evidence: { kind: "timeline_paste_20x5", clipboard_format: "tsv", clipboard_mime: "text/plain", clipboard_delivery: "synthesized_event", viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1, row_count: 20, cell_count: 100, consecutive_frames: 2, all_new_ids: true, geometry_unchanged: true, target_bindings_stable: true, created_records: Array.from({ length: 20 }, (_, index) => ({ record_id: `00000000-0000-0000-0000-${String(index + 1).padStart(12, "0")}`, row_version: 1 })) },
    });
}
