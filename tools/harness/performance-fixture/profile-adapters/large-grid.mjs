import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";
import { canonicalJSONString, parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";

export function readPerformanceContract(root, relativePath, schemaID) {
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

export function fixtureDigest(fixture) {
  return `sha256:${createHash("sha256").update(canonicalJSONString(fixtureProjection(fixture))).digest("hex")}`;
}

function exactFacts(actual, expected, label) {
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(`Large-grid ${label} facts diverge from Core 04`);
  }
}

export function validateLargeGridTraffic(observation, fixture) {
  exactFacts(observation.traffic.counts, [
    { fact_id: "analyst_sessions", value: fixture.analyst_sessions },
    { fact_id: "background_sessions", value: fixture.background_sessions },
  ], "traffic count");
  exactFacts(observation.traffic.rates, [
    {
      fact_id: "background_updates_per_second",
      value: fixture.background_updates_per_second,
    },
  ], "traffic rate");
  exactFacts(observation.traffic.conditions, [
    { fact_id: "presence_enabled", value: true },
    { fact_id: "target_row_excluded", value: true },
  ], "traffic condition");
}
