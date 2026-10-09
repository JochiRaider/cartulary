import { validateSchemaSync } from "./harness-contract.mjs";

export const liveSnapshotSchema = "cartulary.harness_live_snapshot.v2";
export const runObservationSchema = "cartulary.harness_run_observation.v2";
export const liveSnapshotBytes = 16 * 1024 * 1024;
export const observationBytes = 64 * 1024;
export const observationListLimit = 16;
export const observationBlockerLimit = 8;
export const compareObservationIDs = (left, right) => left < right ? -1 : left > right ? 1 : 0;

export function manifestObservationIdentity(manifest) {
  return {
    run_id: manifest.run_id,
    public_target: manifest.target,
    command_id: manifest.command_id,
    source_digest: manifest.source_digest,
    graph_digest: manifest.graph_digest,
  };
}

function sortedUnique(values) {
  return values.every((value, index) => index === 0 || compareObservationIDs(values[index - 1], value) < 0);
}

export function validateLiveSnapshot(snapshot, manifest) {
  validateSchemaSync(liveSnapshotSchema, snapshot);
  for (const [key, value] of Object.entries(manifestObservationIdentity(manifest))) {
    if (snapshot[key] !== value) throw new Error("live snapshot identity does not match manifest");
  }
  if (snapshot.availability === "unavailable") return;
  const ids = snapshot.units.map((unit) => unit.unit_id);
  const members = new Set(ids);
  if (!sortedUnique(ids) || !sortedUnique(snapshot.targets.map((target) => target.target))) throw new Error("live snapshot rosters must be sorted and unique");
  if (!snapshot.targets.some((target) => target.target === manifest.target)) throw new Error("live snapshot omits invocation target");
  for (const target of snapshot.targets) {
    if (!sortedUnique(target.unit_ids) || target.unit_ids.some((id) => !members.has(id))) throw new Error("live target membership is invalid");
  }
  for (const unit of snapshot.units) {
    if (!sortedUnique(unit.blocking_resources) || !sortedUnique(unit.blocking_unit_ids) || unit.blocking_unit_ids.some((id) => !members.has(id))) throw new Error("live blockers are invalid");
    if (unit.started_elapsed_ms !== null && unit.started_elapsed_ms > snapshot.sampled_elapsed_ms) throw new Error("live unit start is ahead of sample");
    if (unit.failure_class !== null && unit.failure_reason === null) throw new Error("live failure lacks its normalized reason");
  }
}
