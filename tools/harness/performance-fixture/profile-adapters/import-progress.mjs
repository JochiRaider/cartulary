import { fixtureDigest, readPerformanceContract, validateLargeGridTraffic } from "./large-grid.mjs";

export function validateImportProgressObservation(root, profile, observation) {
  const contracts = [
    ["discovery", readPerformanceContract(root, "contracts/performance/import-discovery.v1.json", "cartulary.performance_import_discovery_contract.v1")],
    ["apply", readPerformanceContract(root, "contracts/performance/import-apply.v1.json", "cartulary.performance_import_apply_contract.v1")],
  ];
  const admitted = contracts.find(([, value]) => value.predicates.some(predicate => predicate.predicate_id === observation.predicate_id));
  if (!admitted) throw new Error("Import predicate is not admitted");
  const [purpose, contract] = admitted;
  const shared = readPerformanceContract(root, "contracts/performance/large-grid.v1.json", "cartulary.performance_large_grid_contract.v1");
  const predicate = contract.predicates.find(value => value.predicate_id === observation.predicate_id);
  if (!predicate || profile.fixture_profile_id !== contract.fixture_profile_id || observation.fixture_profile_id !== profile.fixture_profile_id ||
      contract.fixture_contract_id !== shared.contract_id || observation.criterion_id !== contract.criterion_id ||
      observation.fixture_id !== shared.fixture.fixture_id || observation.fixture_digest !== fixtureDigest(shared.fixture) ||
      observation.measurement_policy_id !== contract.measurement_policy_id || observation.threshold_ms !== contract.threshold_ms ||
      observation.statistics.kind !== "single" || observation.warmup_samples !== 0) {
    throw new Error("Import progress identity, fixture, policy or statistics diverge from the predicate contract");
  }
  validateLargeGridTraffic(observation, shared.fixture);
  if (["passed", "threshold_failed"].includes(observation.outcome)) {
    const elapsed = observation.samples[0]?.total_ms;
    if (observation.samples.length !== 1 || observation.measured_samples !== 1 ||
        observation.samples[0].warmup || observation.samples[0].sample_index !== 0 ||
        observation.statistics.elapsed_ms !== elapsed ||
        observation.outcome !== (elapsed <= contract.threshold_ms ? "passed" : "threshold_failed")) {
      throw new Error("Import progress single observation, threshold or outcome is inconsistent");
    }
    const proof = observation.predicate_evidence;
    if (proof?.kind !== `import_${purpose}` || proof.source_format !== predicate.kind) {
      throw new Error("Import progress format or job proof is inconsistent");
    }
  }
}
