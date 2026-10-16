import { readFileSync } from "node:fs";
import path from "node:path";
import { parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
import { fixtureDigest, validateLargeGridTraffic } from "./large-grid.mjs";

export function validateTimelinePasteObservation(root, profile, observation) {
  const read = (file, schema) => {
    const value = parseStrictJSON(readFileSync(path.join(root, file), "utf8"), file);
    validateSchemaSync(schema, value);
    return value;
  };
  const contract = read("contracts/performance/ac003.v2.json", "cartulary.performance_paste_contract.v1");
  const shared = read("contracts/performance/large-grid.v1.json", "cartulary.performance_large_grid_contract.v1");
  if (observation.fixture_profile_id !== profile.fixture_profile_id ||
      contract.fixture_contract_id !== shared.contract_id ||
      observation.criterion_id !== contract.criterion_id || observation.predicate_id !== contract.predicate_id ||
      observation.fixture_id !== shared.fixture.fixture_id || observation.fixture_digest !== fixtureDigest(shared.fixture) ||
      observation.measurement_policy_id !== contract.measurement_policy_id || observation.threshold_ms !== contract.threshold_ms ||
      observation.statistics.kind !== "single" || observation.warmup_samples !== 0) {
    throw new Error("AC-003 identity, fixture, policy or statistics diverge from the predicate contract");
  }
  validateLargeGridTraffic(observation, shared.fixture);
  if (["passed", "threshold_failed"].includes(observation.outcome)) {
    const proof = observation.predicate_evidence;
    const elapsed = observation.samples[0]?.total_ms;
    if (observation.samples.length !== 1 || observation.measured_samples !== 1 ||
        observation.samples[0].warmup || observation.samples[0].sample_index !== 0 ||
        observation.statistics.elapsed_ms !== elapsed ||
        observation.outcome !== (elapsed < contract.threshold_ms ? "passed" : "threshold_failed")) {
      throw new Error("AC-003 single observation, strict threshold or outcome is inconsistent");
    }
    if (proof?.kind !== "timeline_paste_20x5" || proof.created_records.length !== 20 ||
        new Set(proof.created_records.map((record) => record.record_id)).size !== 20) {
      throw new Error("AC-003 insertion proof does not contain 20 distinct new records");
    }
  }
}
