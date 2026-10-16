import { fixtureDigest, readPerformanceContract, validateLargeGridTraffic } from "./large-grid.mjs";

export function validateAnchorTraceObservation(root, profile, observation) {
  const contract = readPerformanceContract(root, "contracts/performance/ac047.v2.json", "cartulary.performance_anchor_trace_contract.v1");
  const shared = readPerformanceContract(root, "contracts/performance/large-grid.v1.json", "cartulary.performance_large_grid_contract.v1");
  if (observation.predicate_id !== contract.predicate_id || observation.fixture_profile_id !== profile.fixture_profile_id ||
      observation.criterion_id !== contract.criterion_id || observation.fixture_id !== shared.fixture.fixture_id ||
      observation.fixture_digest !== fixtureDigest(shared.fixture) || observation.measurement_policy_id !== contract.measurement_policy_id ||
      observation.threshold_ms !== null || observation.statistics.kind !== "seeded_trace" ||
      observation.statistics.expected_steps !== contract.steps.length || observation.warmup_samples !== 0 ||
      observation.measured_samples !== 0 || observation.samples.length !== 0) {
    throw new Error("Anchor trace identity, fixture or non-timed policy is inconsistent");
  }
  validateLargeGridTraffic(observation, shared.fixture);
  const proof = observation.predicate_evidence;
  if (proof === null) {
    if (!["incomplete", "environment_not_qualified"].includes(observation.outcome) || observation.statistics.completed_steps !== 0)
      throw new Error("Anchor trace lacks its required proof");
    return;
  }
  if (proof.kind !== "timeline_anchor_trace" || proof.checkpoints.length !== observation.statistics.completed_steps ||
      JSON.stringify(proof.checkpoints.map(step => step.step)) !== JSON.stringify(contract.steps.slice(0, proof.checkpoints.length)) ||
      proof.checkpoints.some(step => step.frame_count < 2 || step.mounted_rows > proof.max_mounted_rows) ||
      proof.frame_count < proof.checkpoints.reduce((total, step) => total + step.frame_count, 0)) {
    throw new Error("Anchor trace step order, frame coverage or mount counts are inconsistent");
  }
  if (observation.outcome === "passed") {
    if (proof.checkpoints.length !== contract.steps.length || proof.violations.length !== 0 ||
        proof.request_record_id !== proof.record_id || proof.max_mounted_rows >= shared.fixture.timeline_rows ||
        proof.committed_row_version !== proof.initial_row_version + 1 || proof.final_row_version !== proof.committed_row_version + 1 ||
        proof.checkpoints.at(-1)?.row_version !== proof.final_row_version ||
        proof.checkpoints.some((step, index) => step.row_version !== (index < 7 ? proof.initial_row_version : index < 9 ? proof.committed_row_version : proof.final_row_version)) ||
        proof.checkpoints.some((step, index) => step.row_version < (proof.checkpoints[index - 1]?.row_version ?? proof.initial_row_version))) {
      throw new Error("Anchor trace does not establish identity, monotonic versions or bounded rendering");
    }
  } else if (observation.outcome === "predicate_failed" && proof.violations.length === 0) {
    throw new Error("Failed anchor trace must retain its violated invariant");
  }
}
