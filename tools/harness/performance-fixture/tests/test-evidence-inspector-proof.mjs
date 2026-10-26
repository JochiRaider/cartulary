import { validateEvidenceInspectorCapabilities } from "../../generated-artifacts/performance/evidence-inspector-contract.mjs";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { validateSchemaSync } from "../../contract/index.mjs";
import { validateProfileMeasurementObservation } from "../profile-adapters/index.mjs";
import test from "node:test";
import { buildEvidenceInspectorObservation } from "./measurement-observations.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
test("Evidence inspector contract binds workload and view capabilities", async (t) => {
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
        const invalid = structuredClone(contract);
        mutate(invalid);
        assert.throws(() => validate(invalid));
    }
    const unavailableFilter = structuredClone(view);
    unavailableFilter.fields.find((field) => field.field_key === "timeline.has_evidence").filter_ops = [];
    assert.throws(() => validateEvidenceInspectorCapabilities(contract, unavailableFilter, policy), /admitted view capabilities/u);
    assert.throws(() => validateEvidenceInspectorCapabilities(contract, view, { ...policy, measured_samples: 99 }), /sampling differs/u);
});
test("Evidence inspector proof admits exact samples and rejects malformed identities and timing", async (t) => {
    const contract = JSON.parse(readFileSync(path.join(root, "contracts/performance/ac045.v2.json"), "utf8"));
    const profile = { fixture_profile_id: "evidence_heavy_snapshot_v2" };
    const observation = buildEvidenceInspectorObservation();
    const validate = value => { validateSchemaSync(value.schema_id, value); validateProfileMeasurementObservation(root, profile, value); };
    validate(observation);
    for (const mutate of [
        v => { v.predicate_id = "perf.evidence_inspector.metadata_shell.v1"; },
        v => { v.fixture_id = "cartulary.perf.evidence_heavy.v1"; },
        v => { v.fixture_digest = `sha256:${"0".repeat(64)}`; },
        v => { v.traffic.counts[0].value = 25; },
        v => { v.predicate_evidence = null; },
        v => { v.predicate_evidence.preview_preflight_records = 5; },
        v => { v.predicate_evidence.measured_access_handle_requests = 1; },
        v => { v.predicate_evidence.fresh_metadata_each_open = false; },
        v => { v.predicate_evidence.consecutive_frames = 1; },
        v => { v.predicate_evidence.owning_viewport = "grid"; },
        v => { v.predicate_evidence.target.row_version = 1; },
        v => { v.predicate_evidence.items[0].row_version = 2; },
        v => { v.predicate_evidence.items[1] = v.predicate_evidence.items[0]; },
        v => { v.predicate_evidence.samples[1].metadata_started_at = 1; },
        v => { v.predicate_evidence.samples[1].metadata_completed_at = 99999; },
        v => { v.predicate_evidence.samples[1].first_frame_at = v.predicate_evidence.samples[1].second_frame_at; },
        v => { v.predicate_evidence.samples[1].total_ms = 99; },
        v => { v.predicate_evidence.samples[1].warmup = true; },
        v => { v.samples.pop(); },
        v => { v.statistics.p95_ms = 99; },
        v => { v.outcome = "threshold_failed"; },
        v => { v.predicate_evidence.retry_to_green = true; },
    ]) {
        const invalid = structuredClone(observation);
        mutate(invalid);
        assert.throws(() => validate(invalid));
    }
    for (const [elapsed, outcome] of [[300, "passed"], [300.01, "threshold_failed"]]) {
        const boundary = structuredClone(observation);
        boundary.outcome = outcome;
        boundary.statistics.p50_ms = elapsed;
        boundary.statistics.p95_ms = elapsed;
        for (const item of boundary.predicate_evidence.samples) {
            item.second_frame_at = item.accepted_at + elapsed;
            item.total_ms = item.second_frame_at - item.accepted_at;
        }
        boundary.samples = boundary.predicate_evidence.samples.map(({ sample_index, warmup, total_ms }) => ({ sample_index, warmup, total_ms, stages_ms: [{ fact_id: "accepted_to_visible_paint", value: total_ms }] }));
        const values = boundary.samples.slice(1).map(s => s.total_ms).sort((a, b) => a - b);
        boundary.statistics.p50_ms = values[49];
        boundary.statistics.p95_ms = values[94];
        validate(boundary);
    }
});
