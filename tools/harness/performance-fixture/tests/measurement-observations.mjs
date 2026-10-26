import { readFileSync } from "node:fs";
import { createHash } from "node:crypto";
import path from "node:path";
import { canonicalJSONString } from "../../contract/index.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
export function buildEvidenceInspectorObservation() {
    const contract = JSON.parse(readFileSync(path.join(root, "contracts/performance/ac045.v2.json"), "utf8"));
    const profile = { fixture_profile_id: "evidence_heavy_snapshot_v2" };
    const sample = (index) => ({ sample_index: index, warmup: index === 0, total_ms: 100, accepted_at: index * 1000 + 10, metadata_started_at: index * 1000 + 11, metadata_completed_at: index * 1000 + 50, first_frame_at: index * 1000 + 90, second_frame_at: index * 1000 + 110 });
    const proof = { schema_id: "cartulary.evidence_inspector_measurement_proof.v1", kind: "evidence_inspector", consecutive_frames: 2, all_content_visible: true, bindings_stable: true, fresh_metadata_each_open: true,
        target: { record_id: "00000000-0000-0000-0000-000000000100", row_version: 2 },
        items: Array.from({ length: 5 }, (_, index) => ({ record_id: `00000000-0000-0000-0000-${String(index + 1).padStart(12, "0")}`, row_version: 1 })),
        viewport_width: 1920, viewport_height: 1080, device_scale_factor: 1, owning_viewport: "inspector", preview_preflight_records: 100, preview_preflight_body_verification: "complete_size_and_sha256", measured_access_handle_requests: 0, samples: Array.from({ length: 101 }, (_, i) => sample(i)) };
    const observation = { schema_id: "cartulary.frontend_measurement_observation.v4", fixture_profile_id: profile.fixture_profile_id, criterion_id: contract.criterion_id, predicate_id: contract.predicate_id,
        fixture_id: contract.fixture.fixture_id, fixture_digest: `sha256:${createHash("sha256").update(canonicalJSONString(contract.fixture)).digest("hex")}`, measurement_policy_id: contract.measurement_policy_id, threshold_ms: contract.threshold_ms,
        warmup_samples: 1, measured_samples: 100, statistics: { kind: "percentile", percentile: 95, p50_ms: 100, p95_ms: 100 }, outcome: "passed",
        traffic: { counts: [{ fact_id: "analyst_sessions", value: 1 }, { fact_id: "background_sessions", value: 0 }], rates: [], conditions: [] },
        samples: proof.samples.map(({ sample_index, warmup, total_ms }) => ({ sample_index, warmup, total_ms, stages_ms: [{ fact_id: "accepted_to_visible_paint", value: total_ms }] })), predicate_evidence: proof };
    return observation;
}
