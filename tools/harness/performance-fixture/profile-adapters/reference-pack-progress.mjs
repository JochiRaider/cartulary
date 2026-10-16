import { referencePackBundle } from "../../browser/index.mjs";
import { fixtureDigest, readPerformanceContract, validateLargeGridTraffic } from "./large-grid.mjs";

export function validateReferencePackProgressObservation(root, profile, observation) {
  const contract = readPerformanceContract(root, "contracts/performance/reference-pack-progress.v1.json", "cartulary.performance_reference_pack_progress_contract.v1");
  const shared = readPerformanceContract(root, "contracts/performance/large-grid.v1.json", "cartulary.performance_large_grid_contract.v1");
  const predicate = contract.predicates.find(value => value.predicate_id === observation.predicate_id);
  if (!predicate || profile.fixture_profile_id !== contract.fixture_profile_id || observation.fixture_profile_id !== profile.fixture_profile_id ||
      contract.fixture_contract_id !== shared.contract_id || observation.criterion_id !== contract.criterion_ids[0] ||
      observation.fixture_id !== shared.fixture.fixture_id || observation.fixture_digest !== fixtureDigest(shared.fixture) ||
      observation.measurement_policy_id !== contract.measurement_policy_id || observation.threshold_ms !== contract.threshold_ms ||
      observation.statistics.kind !== "single" || observation.warmup_samples !== 0) {
    throw new Error("Reference Pack progress identity, fixture, policy or statistics differ from the predicate contract");
  }
  validateLargeGridTraffic(observation, shared.fixture);
  if (["passed", "threshold_failed"].includes(observation.outcome)) {
    const elapsed = observation.samples[0]?.total_ms;
    if (observation.samples.length !== 1 || observation.measured_samples !== 1 || observation.samples[0].warmup || observation.samples[0].sample_index !== 0 ||
        observation.statistics.elapsed_ms !== elapsed || observation.outcome !== (elapsed <= contract.threshold_ms ? "passed" : "threshold_failed"))
      throw new Error("Reference Pack single observation, threshold or outcome is inconsistent");
    const proof = observation.predicate_evidence;
    if (proof?.kind !== "reference_pack_progress" || proof.job_purpose !== predicate.kind)
      throw new Error("Reference Pack progress purpose or proof is inconsistent");
    const bundle = referencePackBundle({ preparationDate: proof.bundle_preparation_date });
    if (bundle.baseVersion !== contract.workload.base_version || bundle.payloadSha256 !== contract.workload.payload_sha256 ||
        proof.bundle_sha256 !== bundle.sha256 || proof.bundle_bytes !== bundle.upload.buffer.length ||
        proof.pack_key !== bundle.key || proof.pack_version !== bundle.version)
      throw new Error("Reference Pack input differs from the exact prepared bundle workload");
  }
}
