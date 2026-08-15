import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";

import {
  canonicalJSONString,
  parseStrictJSON,
  validateSchemaSync,
} from "../../contract/index.mjs";

const profileID = "ac043_large_grid_snapshot_v1";

function readContract(root, relativePath, schemaID) {
  const file = path.join(root, relativePath);
  const value = parseStrictJSON(readFileSync(file, "utf8"), file);
  validateSchemaSync(schemaID, value);
  return value;
}

function fixtureProjection(fixture) {
  return {
    fixtureId: fixture.fixture_id,
    seed: fixture.seed,
    timelineRows: fixture.timeline_rows,
    hostRows: fixture.host_rows,
    identityRows: fixture.identity_rows,
    tagAssignments: fixture.tag_assignments,
    mentionAssignments: fixture.mention_assignments,
    linkAssignments: fixture.link_assignments,
    analystSessions: fixture.analyst_sessions,
    backgroundSessions: fixture.background_sessions,
    backgroundUpdateIntervalMs: fixture.background_update_interval_ms,
    backgroundUpdatesPerSecond: fixture.background_updates_per_second,
    trafficTraceId: fixture.traffic_trace_id,
  };
}

function fixtureDigest(fixture) {
  return `sha256:${createHash("sha256").update(canonicalJSONString(fixtureProjection(fixture))).digest("hex")}`;
}

function exactFacts(actual, expected, label) {
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(`AC-043 ${label} facts diverge from Core 04`);
  }
}

export function validateAc043MeasurementObservation(root, profile, observation) {
  if (profile.fixture_profile_id !== profileID) {
    throw new Error("AC-043 observation adapter received another fixture profile");
  }
  const contract = readContract(
    root,
    "contracts/performance/ac043.v1.json",
    "cartulary.performance_interaction_contract.v1",
  );
  const policy = readContract(
    root,
    "tools/measurement_policy_owner.json",
    "cartulary.measurement_policy.v1",
  );
  const predicate = contract.predicates.find(
    (candidate) => candidate.predicate_id === observation.predicate_id,
  );
  if (
    predicate === undefined ||
    observation.fixture_profile_id !== profile.fixture_profile_id ||
    observation.criterion_id !== contract.criterion_id ||
    observation.fixture_id !== contract.fixture.fixture_id ||
    observation.fixture_digest !== fixtureDigest(contract.fixture) ||
    observation.measurement_policy_id !== policy.policy_id ||
    observation.threshold_ms !== predicate.threshold_ms ||
    observation.percentile !== policy.percentile
  ) {
    throw new Error("AC-043 measurement observation diverges from Core 04 identity or threshold");
  }
  exactFacts(observation.traffic.counts, [
    { fact_id: "analyst_sessions", value: contract.fixture.analyst_sessions },
    { fact_id: "background_sessions", value: contract.fixture.background_sessions },
  ], "traffic count");
  exactFacts(observation.traffic.rates, [
    {
      fact_id: "background_updates_per_second",
      value: contract.fixture.background_updates_per_second,
    },
  ], "traffic rate");
  exactFacts(observation.traffic.conditions, [
    { fact_id: "presence_enabled", value: true },
    { fact_id: "target_row_excluded", value: true },
  ], "traffic condition");
  if (
    ["passed", "threshold_failed"].includes(observation.outcome) &&
    (observation.warmup_samples !== policy.warmup_passes ||
      observation.measured_samples !== policy.measured_samples ||
      observation.samples.length !== policy.warmup_passes + policy.measured_samples)
  ) {
    throw new Error("AC-043 qualifying observation has incomplete sampling");
  }
}
