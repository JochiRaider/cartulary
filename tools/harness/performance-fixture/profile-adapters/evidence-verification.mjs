import { createHash } from "node:crypto";
import { canonicalJSONString, validateSchemaSync } from "../../contract/index.mjs";
import { readPerformanceContract } from "./large-grid.mjs";

export function validateEvidenceVerificationObservation(root, profile, observation) {
  const contract = readPerformanceContract(root, "contracts/performance/evidence-verification.v1.json", "cartulary.performance_evidence_verification_contract.v1");
  const shared = readPerformanceContract(root, "contracts/performance/ac045.v2.json", "cartulary.performance_evidence_inspector_contract.v1");
  const digest = `sha256:${createHash("sha256").update(canonicalJSONString(shared.fixture)).digest("hex")}`;
  if (profile.fixture_profile_id !== contract.fixture_profile_id || observation.fixture_profile_id !== profile.fixture_profile_id ||
      contract.fixture_contract_id !== shared.contract_id || observation.criterion_id !== contract.criterion_ids[0] ||
      observation.predicate_id !== contract.predicate_id || observation.fixture_id !== shared.fixture.fixture_id ||
      observation.fixture_digest !== digest || observation.measurement_policy_id !== contract.measurement_policy_id ||
      observation.threshold_ms !== contract.threshold_ms || observation.statistics.kind !== "single" || observation.warmup_samples !== 0 ||
      JSON.stringify(observation.traffic) !== JSON.stringify({counts:[{fact_id:"analyst_sessions",value:1},{fact_id:"background_sessions",value:0}],rates:[],conditions:[]})) {
    throw new Error("Evidence verification identity, fixture, traffic or policy differs from its workload");
  }
  const proof = observation.predicate_evidence;
  if (proof !== null) {
    validateSchemaSync("cartulary.evidence_verification_measurement_proof.v1", proof);
    if (proof.source_record_id === proof.record_id) throw new Error("Evidence verification must preserve a distinct grid edit target");
  }
  if (["passed", "threshold_failed"].includes(observation.outcome)) {
    const elapsed = observation.samples[0]?.total_ms;
    if (!proof || observation.samples.length !== 1 || observation.measured_samples !== 1 ||
        observation.samples[0].warmup || observation.samples[0].sample_index !== 0 || observation.statistics.elapsed_ms !== elapsed ||
        observation.outcome !== (elapsed <= contract.threshold_ms ? "passed" : "threshold_failed")) {
      throw new Error("Evidence verification single observation, proof or threshold is inconsistent");
    }
  }
}
