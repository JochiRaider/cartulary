import { fixtureDigest, readPerformanceContract, validateLargeGridTraffic } from "./large-grid.mjs";

const profileID = "timeline_large_grid_snapshot_v2";

export function validateInteractiveMeasurementObservation(root, profile, observation, contractPath, schemaID) {
  if (profile.fixture_profile_id !== profileID) {
    throw new Error("Interactive observation adapter received another fixture profile");
  }
  const contract = readPerformanceContract(
    root,
    contractPath,
    schemaID,
  );
  const policy = readPerformanceContract(
    root,
    "tools/measurement_policy_owner.json",
    "cartulary.measurement_policy.v1",
  );
  const shared = readPerformanceContract(root, "contracts/performance/large-grid.v1.json", "cartulary.performance_large_grid_contract.v1");
  const predicate = contract.predicates.find(
    (candidate) => candidate.predicate_id === observation.predicate_id,
  );
  if (
    predicate === undefined ||
    contract.fixture_contract_id !== shared.contract_id ||
    observation.fixture_profile_id !== profile.fixture_profile_id ||
    observation.criterion_id !== contract.criterion_id ||
    observation.fixture_id !== shared.fixture.fixture_id ||
    observation.fixture_digest !== fixtureDigest(shared.fixture) ||
    observation.measurement_policy_id !== policy.policy_id ||
    observation.threshold_ms !== predicate.threshold_ms ||
    observation.statistics.kind !== "percentile" ||
    observation.statistics.percentile !== policy.percentile ||
    observation.predicate_evidence !== null
  ) {
    throw new Error("Interactive measurement observation diverges from Core 04 identity or threshold");
  }
  validateLargeGridTraffic(observation, shared.fixture);
  if (["passed", "threshold_failed"].includes(observation.outcome)) {
    const values = observation.samples.filter((sample) => !sample.warmup).map((sample) => sample.total_ms).sort((a, b) => a - b);
    const p50 = values[Math.ceil(0.5 * values.length) - 1];
    const p95 = values[Math.ceil(policy.percentile / 100 * values.length) - 1];
    if (observation.statistics.p50_ms !== p50 || observation.statistics.p95_ms !== p95 ||
      observation.outcome !== (p95 <= predicate.threshold_ms ? "passed" : "threshold_failed")) {
      throw new Error("Interactive statistics or outcome differ from retained samples");
    }
  }
  if (
    ["passed", "threshold_failed"].includes(observation.outcome) &&
    (observation.warmup_samples !== policy.warmup_passes ||
      observation.measured_samples !== policy.measured_samples ||
      observation.samples.length !== policy.warmup_passes + policy.measured_samples)
  ) {
    throw new Error("Interactive qualifying observation has incomplete sampling");
  }
}
