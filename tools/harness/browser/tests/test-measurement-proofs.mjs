import { buildEvidenceInspectorObservation } from "../../performance-fixture/tests/measurement-observations.mjs";
import { referencePackBundle } from "../reference-pack-fixture.mjs";
import { validateViewChangeCapabilities } from "../../generated-artifacts/performance/performance-contracts.mjs";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { activePerformanceFixtureProfile, loadPerformanceFixtureSnapshotRegistry } from "../../performance-fixture/index.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";
import { validateProfileMeasurementObservation } from "../../performance-fixture/profile-adapters/index.mjs";
import { fixtureDigest } from "../../performance-fixture/profile-adapters/large-grid.mjs";
import test from "node:test";
import { measurementContext, measurementObservation, buildPasteObservation } from "./measurement-fixtures.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
test("paste proof preserves exact synthetic identities and threshold", async (t) => {
    const { fixtureProfile, collect, collectWrittenSummary, persist } = measurementContext(t);
    const paste = buildPasteObservation();
    const syntheticProofObservations = [];
    const validatePaste = (value) => {
        validateSchemaSync("cartulary.frontend_measurement_observation.v4", value);
        validateProfileMeasurementObservation(root, fixtureProfile, value);
    };
    validatePaste(paste);
    assert.equal(collect(paste).length, 1, "only typed synthetic insertion identities are retained");
    syntheticProofObservations.push(paste);
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
        const invalid = structuredClone(paste);
        mutate(invalid);
        assert.throws(() => validatePaste(invalid));
    }
    const atThreshold = structuredClone(paste);
    atThreshold.samples[0].total_ms = 2000;
    atThreshold.statistics.elapsed_ms = 2000;
    atThreshold.outcome = "threshold_failed";
    validatePaste(atThreshold);
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
test("presence proof binds each surface row and cell observation", async (t) => {
    const { fixtureProfile, collect, collectWrittenSummary, persist } = measurementContext(t);
    const paste = buildPasteObservation();
    const syntheticProofObservations = [];
    const presenceContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/presence.v2.json"), "utf8"));
    validateSchemaSync("cartulary.performance_presence_contract.v2", presenceContract);
    for (const mutate of [
        value => { value.criterion_ids = ["AC-008"]; },
        value => { value.criterion_ids = ["AC-008", "AC-008"]; },
        value => { value.criterion_ids.reverse(); },
        value => { value.criterion_id = "AC-008"; },
        value => { value.schema_id = "cartulary.performance_presence_contract.v1"; },
    ]) {
        const invalid = structuredClone(presenceContract);
        mutate(invalid);
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
            const invalid = structuredClone(presence);
            mutate(invalid);
            assert.throws(() => validatePresence(invalid));
        }
        const slow = structuredClone(presence);
        slow.samples[0].total_ms = 1001;
        slow.statistics.elapsed_ms = 1001;
        slow.outcome = "threshold_failed";
        validatePresence(slow);
    }
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
test("anchor trace proof rejects changed geometry bindings and drafts", async (t) => {
    const { fixtureProfile, collect, collectWrittenSummary, persist } = measurementContext(t);
    const paste = buildPasteObservation();
    const syntheticProofObservations = [];
    const anchorContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/ac047.v2.json"), "utf8"));
    const trace = { ...paste, criterion_id: "AC-047", predicate_id: anchorContract.predicate_id,
        measurement_policy_id: anchorContract.measurement_policy_id, threshold_ms: null, measured_samples: 0, samples: [],
        statistics: { kind: "seeded_trace", completed_steps: anchorContract.steps.length, expected_steps: anchorContract.steps.length },
        predicate_evidence: { schema_id: "cartulary.anchor_trace_proof.v1", kind: "timeline_anchor_trace", record_id: "00000000-0000-0000-0000-000000000001", request_record_id: "00000000-0000-0000-0000-000000000001",
            initial_row_version: 1, committed_row_version: 2, final_row_version: 3, max_mounted_rows: 35, frame_count: 22, grid_replacements: 0, violations: [],
            checkpoints: anchorContract.steps.map((step, index) => ({ step, frame_count: 2, row_version: index >= 9 ? 3 : index >= 7 ? 2 : 1, mounted_rows: 35 })),
        },
    };
    const validateTrace = value => { validateSchemaSync("cartulary.frontend_measurement_observation.v4", value); validateProfileMeasurementObservation(root, fixtureProfile, value); };
    validateTrace(trace);
    assert.equal(collect(trace).length, 1);
    syntheticProofObservations.push(trace);
    for (const mutate of [
        value => { value.threshold_ms = 2000; },
        value => { value.statistics.kind = "single"; },
        value => { value.statistics.completed_steps = 10; },
        value => { value.statistics.expected_steps = 12; },
        value => { value.predicate_evidence.checkpoints.reverse(); },
        value => { value.predicate_evidence.checkpoints[0].frame_count = 1; },
        value => { value.predicate_evidence.frame_count = 1; },
        value => { value.predicate_evidence.request_record_id = null; },
        value => { value.predicate_evidence.request_record_id = "00000000-0000-0000-0000-000000000002"; },
        value => { value.predicate_evidence.committed_row_version = 1; },
        value => { value.predicate_evidence.final_row_version = 2; },
        value => { value.predicate_evidence.checkpoints[8].row_version = 1; },
        value => { value.predicate_evidence.max_mounted_rows = 20000; },
        value => { value.predicate_evidence.grid_replacements = -1; },
        value => { value.predicate_evidence.violations.push("draft_changed"); },
        value => { value.predicate_evidence = null; },
        value => { value.samples = [{ sample_index: 0, warmup: false, total_ms: 100, stages_ms: [] }]; },
        value => { value.traffic.counts[0].value = 24; },
        value => { value.outcome = "threshold_failed"; },
        value => { value.schema_id = "cartulary.frontend_measurement_observation.v3"; },
    ]) {
        const invalid = structuredClone(trace);
        mutate(invalid);
        assert.throws(() => validateTrace(invalid));
    }
    const failedTrace = structuredClone(trace);
    failedTrace.outcome = "predicate_failed";
    failedTrace.predicate_evidence.violations.push("draft_changed");
    validateTrace(failedTrace);
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
test("import proofs bind discovery and apply to admitted source formats", async (t) => {
    const { fixtureProfile, collect, collectWrittenSummary, persist } = measurementContext(t);
    const paste = buildPasteObservation();
    const syntheticProofObservations = [];
    for (const purpose of ["discovery", "apply"]) {
        const importContract = JSON.parse(readFileSync(path.join(root, `contracts/performance/import-${purpose}.v1.json`), "utf8"));
        const importProfile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), importContract.fixture_profile_id);
        for (const predicate of importContract.predicates) {
            const discovery = { ...paste, fixture_profile_id: importProfile.fixture_profile_id, criterion_id: "AC-027", predicate_id: predicate.predicate_id, threshold_ms: 1000,
                statistics: { kind: "single", elapsed_ms: 1000 }, samples: [{ sample_index: 0, warmup: false, total_ms: 1000, stages_ms: [] }],
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
            ]) {
                const invalid = structuredClone(discovery);
                mutate(invalid);
                assert.throws(() => validateDiscovery(invalid));
            }
            const slow = structuredClone(discovery);
            slow.samples[0].total_ms = 1000.01;
            slow.statistics.elapsed_ms = 1000.01;
            slow.outcome = "threshold_failed";
            validateDiscovery(slow);
        }
    }
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
test("verification proof binds active work cancellation and local draft", async (t) => {
    const { collect, collectWrittenSummary, persist } = measurementContext(t);
    const evidenceInspectorObservation = buildEvidenceInspectorObservation();
    const syntheticProofObservations = [];
    const verificationContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/evidence-verification.v1.json"), "utf8"));
    const verificationProfile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), verificationContract.fixture_profile_id);
    const verificationObservation = { ...evidenceInspectorObservation, criterion_id: "AC-016", predicate_id: verificationContract.predicate_id,
        measurement_policy_id: verificationContract.measurement_policy_id, threshold_ms: 1000, warmup_samples: 0, measured_samples: 1,
        statistics: { kind: "single", elapsed_ms: 1000 }, samples: [{ sample_index: 0, warmup: false, total_ms: 1000, stages_ms: [] }],
        predicate_evidence: { schema_id: "cartulary.evidence_verification_measurement_proof.v1", kind: "evidence_verification", job_purpose: "evidence_verification",
            job_id: "00000000-0000-4000-8000-000000000001", job_status: "running", object_blob_id: "00000000-0000-4000-8000-000000000002",
            record_id: "00000000-0000-4000-8000-000000000003", source_record_id: "00000000-0000-4000-8000-000000000004", row_version: 1,
            view_schema_id: "cartulary.view.timeline.v2", field_key: "timeline.analyst_text", clipboard_driver: "synthesized_file_event", uploaded_bytes: 2097152,
            preview_preflight_records: 100, inspector_open: false, attachment_details_expanded: false, cancel_enabled: true, grid_edit_accepted: true, draft_uncommitted: true,
            consecutive_frames: 2, binding_stable: true, geometry_unchanged: true, scroll_unchanged: true, action_count: 1, viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1 } };
    const validateVerification = value => { validateSchemaSync("cartulary.frontend_measurement_observation.v4", value); validateProfileMeasurementObservation(root, verificationProfile, value); };
    validateVerification(verificationObservation);
    assert.equal(collect(verificationObservation).length, 1);
    syntheticProofObservations.push(verificationObservation);
    for (const mutate of [
        v => { v.predicate_evidence.source_record_id = v.predicate_evidence.record_id; }, v => { v.predicate_evidence.object_blob_id = null; },
        v => { v.predicate_evidence.clipboard_driver = "platform_clipboard"; }, v => { v.predicate_evidence.uploaded_bytes = 1; },
        v => { v.predicate_evidence.preview_preflight_records = 99; }, v => { v.predicate_evidence.job_status = "succeeded"; },
        v => { v.predicate_evidence.inspector_open = true; }, v => { v.predicate_evidence.attachment_details_expanded = true; },
        v => { v.predicate_evidence.cancel_enabled = false; }, v => { v.predicate_evidence.grid_edit_accepted = false; }, v => { v.predicate_evidence.consecutive_frames = 1; },
        v => { v.predicate_evidence.action_count = 2; }, v => { v.predicate_evidence.scroll_unchanged = false; }, v => { v.predicate_evidence = null; },
        v => { v.traffic.counts[0].value = 25; }, v => { v.samples[0].total_ms = 1000.01; v.statistics.elapsed_ms = 1000.01; }, v => { v.warmup_samples = 1; },
    ]) {
        const invalid = structuredClone(verificationObservation);
        mutate(invalid);
        assert.throws(() => validateVerification(invalid));
    }
    const slowVerification = structuredClone(verificationObservation);
    slowVerification.samples[0].total_ms = 1000.01;
    slowVerification.statistics.elapsed_ms = 1000.01;
    slowVerification.outcome = "threshold_failed";
    validateVerification(slowVerification);
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
test("reference pack proofs bind bundle identity and foreground operation", async (t) => {
    const { collect, collectWrittenSummary, persist } = measurementContext(t);
    const paste = buildPasteObservation();
    const syntheticProofObservations = [];
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
        validate(observation);
        assert.equal(collect(observation).length, 1);
        syntheticProofObservations.push(observation);
        for (const mutate of [
            v => { v.predicate_id = "perf.job_progress.visible_with_cancel.v1"; }, v => { v.predicate_evidence.job_purpose = "other"; },
            v => { v.predicate_evidence.bundle_sha256 = "0".repeat(64); }, v => { v.predicate_evidence.bundle_bytes++; },
            v => { v.predicate_evidence.bundle_preparation_date = "2026-10-17"; }, v => { v.predicate_evidence.bundle_preparation_date = "2026-99-99"; },
            v => { v.predicate_evidence.payload_sha256 = "0".repeat(64); }, v => { v.predicate_evidence.pack_version = "other"; },
            v => { v.predicate_evidence.grid_location = "current"; }, v => { v.predicate_evidence.companion_consecutive_frames = 1; },
            v => { v.predicate_evidence.job_status = "succeeded"; }, v => { v.predicate_evidence.cancel_enabled = false; },
            v => { v.predicate_evidence.grid_edit_accepted = false; }, v => { v.predicate_evidence.action_count = 2; },
            v => { v.predicate_evidence.scroll_unchanged = false; }, v => { v.predicate_evidence.ordinary_security = false; },
            v => { v.predicate_evidence.same_foreground_session = false; }, v => { v.predicate_evidence = null; },
            v => { v.traffic.counts[0].value = 26; }, v => { v.samples[0].total_ms = 1000.01; v.statistics.elapsed_ms = 1000.01; },
        ]) {
            const invalid = structuredClone(observation);
            mutate(invalid);
            assert.throws(() => validate(invalid));
        }
        const slow = structuredClone(observation);
        slow.samples[0].total_ms = 1000.01;
        slow.statistics.elapsed_ms = 1000.01;
        slow.outcome = "threshold_failed";
        validate(slow);
    }
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
test("reporting proofs bind snapshot and render provenance", async (t) => {
    const { collect, collectWrittenSummary, persist } = measurementContext(t);
    const paste = buildPasteObservation();
    const syntheticProofObservations = [];
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
        validate(observation);
        assert.equal(collect(observation).length, 1);
        syntheticProofObservations.push(observation);
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
        ]) {
            const invalid = structuredClone(observation);
            mutate(invalid);
            assert.throws(() => validate(invalid));
        }
        const slow = structuredClone(observation);
        slow.samples[0].total_ms = 1000.01;
        slow.statistics.elapsed_ms = 1000.01;
        slow.outcome = "threshold_failed";
        validate(slow);
    }
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
test("history proofs bind native action and attributed version changes", async (t) => {
    const { fixtureProfile, collect, collectWrittenSummary, persist } = measurementContext(t);
    const paste = buildPasteObservation();
    const syntheticProofObservations = [];
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
        ]) {
            const invalid = structuredClone(history);
            mutate(invalid);
            assert.throws(() => validateHistory(invalid));
        }
        const slow = structuredClone(history);
        slow.samples[0].total_ms = 2000;
        slow.statistics.elapsed_ms = 2000;
        slow.outcome = "threshold_failed";
        validateHistory(slow);
    }
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
test("view change proofs bind capabilities exact policy and sample counts", async (t) => {
    const { fixtureProfile, collectWrittenSummary, persist } = measurementContext(t);
    const syntheticProofObservations = [];
    const viewChangeContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/ac044.v1.json"), "utf8"));
    const viewSchema = JSON.parse(readFileSync(path.join(root, "contracts/view-schemas/cartulary.view.timeline.v2.json"), "utf8"));
    validateViewChangeCapabilities(viewChangeContract, viewSchema);
    for (const mutate of [
        value => { value.workloads.filter.field_key = "timeline.activity_synopsis_text"; },
        value => { value.workloads.filter.operator = "contains"; },
        value => { value.workloads.group.field_key = "timeline.activity_synopsis_text"; },
        value => { value.workloads.sort.field_key = "timeline.tags"; },
    ]) {
        const invalid = structuredClone(viewChangeContract);
        mutate(invalid);
        assert.throws(() => validateViewChangeCapabilities(invalid, viewSchema), /admitted view capabilities/u);
    }
    const sharedContract = JSON.parse(readFileSync(path.join(root, "contracts/performance/large-grid.v1.json"), "utf8"));
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
                const invalid = structuredClone(viewChange);
                mutate(invalid);
                assert.throws(() => validateProfileMeasurementObservation(root, fixtureProfile, invalid));
            }
        }
    }
    assert.throws(() => validateProfileMeasurementObservation(root, fixtureProfile, { ...percentile, outcome: "threshold_failed" }), /outcome differ/u);
    assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations);
});
function assertRetainedProofs({ persist, collectWrittenSummary }, syntheticProofObservations) {
    for (const observation of syntheticProofObservations) {
        const profile = activePerformanceFixtureProfile(loadPerformanceFixtureSnapshotRegistry(root), observation.fixture_profile_id);
        const context = { profile, ...(observation.predicate_evidence.kind === "reference_pack_progress" ? { rowID: `module.reference_data.measurement.${observation.predicate_evidence.job_purpose}_progress` } : {}), ...(observation.predicate_evidence.kind === "evidence_verification" ? { rowID: "module.evidence.measurement.verification_progress" } : {}), ...(observation.predicate_evidence.kind === "reporting_progress" ? { rowID: `module.reporting.measurement.${observation.predicate_evidence.job_purpose === "snapshot" ? "snapshot" : "render"}_progress` } : {}), ...(["import_discovery", "import_apply"].includes(observation.predicate_evidence.kind) ? { rowID: `module.imports.measurement.${observation.predicate_evidence.source_format}_${observation.predicate_evidence.kind === "import_apply" ? "apply" : "discovery"}` } : {}) };
        const validateObservation = value => validateProfileMeasurementObservation(root, profile, value);
        const retained = persist({ ...context, observationOverrides: observation });
        const [qualified] = collectWrittenSummary(retained, validateObservation);
        assert.equal(qualified.summary.rollup.predicate_id, observation.predicate_id, "retained finalization must admit the same closed synthetic proof as collection");
        for (const extra of [{ payload: "private" }, { record_id: "00000000-0000-0000-0000-000000000001" }]) {
            const invalid = persist({
                ...context, observationOverrides: { ...observation, ...extra },
            });
            assert.throws(() => collectWrittenSummary(invalid, validateObservation), /additional properties/u, "admitting a proof never admits sensitive top-level fields");
        }
        const invalidProof = persist({
            ...context, observationOverrides: { ...observation, predicate_evidence: { ...observation.predicate_evidence, payload: "private" } },
        });
        assert.throws(() => collectWrittenSummary(invalidProof, validateObservation), /validation failed/u);
    }
}
