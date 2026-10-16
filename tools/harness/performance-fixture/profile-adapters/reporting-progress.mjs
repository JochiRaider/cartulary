import { fixtureDigest, readPerformanceContract, validateLargeGridTraffic } from "./large-grid.mjs";

export function validateReportingProgressObservation(root, profile, observation) {
  const contract = readPerformanceContract(root, "contracts/performance/reporting-progress.v1.json", "cartulary.performance_reporting_progress_contract.v1");
  const shared = readPerformanceContract(root, "contracts/performance/large-grid.v1.json", "cartulary.performance_large_grid_contract.v1");
  const predicate = contract.predicates.find(value => value.predicate_id === observation.predicate_id);
  if (!predicate || profile.fixture_profile_id !== contract.fixture_profile_id || observation.fixture_profile_id !== profile.fixture_profile_id ||
      contract.fixture_contract_id !== shared.contract_id || observation.criterion_id !== contract.criterion_id ||
      observation.fixture_id !== shared.fixture.fixture_id || observation.fixture_digest !== fixtureDigest(shared.fixture) ||
      observation.measurement_policy_id !== contract.measurement_policy_id || observation.threshold_ms !== contract.threshold_ms ||
      observation.statistics.kind !== "single" || observation.warmup_samples !== 0) {
    throw new Error("Reporting progress identity, fixture, policy or statistics differ from the predicate contract");
  }
  validateLargeGridTraffic(observation, shared.fixture);
  if (["passed", "threshold_failed"].includes(observation.outcome)) {
    const elapsed = observation.samples[0]?.total_ms;
    if (observation.samples.length !== 1 || observation.measured_samples !== 1 ||
        observation.samples[0].warmup || observation.samples[0].sample_index !== 0 ||
        observation.statistics.elapsed_ms !== elapsed ||
        observation.outcome !== (elapsed <= contract.threshold_ms ? "passed" : "threshold_failed")) {
      throw new Error("Reporting progress single observation, threshold or outcome is inconsistent");
    }
    const proof = observation.predicate_evidence;
    if (proof?.kind !== "reporting_progress" || proof.job_purpose !== predicate.kind) {
      throw new Error("Reporting progress job purpose or proof is inconsistent");
    }
  }
}
