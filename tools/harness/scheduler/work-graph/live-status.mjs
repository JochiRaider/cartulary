import path from "node:path";
import {
  compareObservationIDs,
  liveSnapshotBytes,
  liveSnapshotSchema,
  manifestObservationIdentity,
  validateLiveSnapshot,
} from "../../contract/live-observation.mjs";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";

// A diagnostic sink, not an event reducer. The scheduler supplies coherent
// state; publication failures cannot reject scheduling or replace its outcome.
export function createLiveStatusPublisher({
  runRoot, manifest, graph, projections,
  now = () => performance.now(), utc = () => new Date().toISOString(),
  write = atomicLocalFile, timers = { setTimeout, clearTimeout },
  maximumBytes = liveSnapshotBytes,
}) {
  const started = now();
  const identity = manifestObservationIdentity(manifest);
  const targets = Object.entries(projections).map(([target, units]) => ({
    target, unit_ids: [...new Set(units)].sort(compareObservationIDs),
  })).sort((a, b) => compareObservationIDs(a.target, b.target));
  let view = {
    phase: "executing", last_transition_seq: 0, sampled_elapsed_ms: 0,
    units: graph.units.map((unit) => ({
      unit_id: unit.unit_id, status: "pending", activity: "pending",
      started_elapsed_ms: null, wait_reason: null, blocking_resources: [],
      blocking_unit_ids: [], cache_disposition: "not_observed",
      failure_class: null, failure_reason: null,
    })).sort((a, b) => compareObservationIDs(a.unit_id, b.unit_id)),
  };
  let revision = 0;
  let lastPublication = -Infinity;
  let viewReceivedAt = started;
  let schedulerOffset = null;
  let dirty = false;
  let stopped = false;
  let pauses = 0;
  let timer = null;
  let publicationError = null;
  const clear = () => { if (timer !== null) timers.clearTimeout(timer); timer = null; };
  function publish() {
    lastPublication = now();
    dirty = false;
    if (revision >= Number.MAX_SAFE_INTEGER) return;
    try {
      const snapshot = {
        schema_id: liveSnapshotSchema, ...identity, revision: revision + 1,
        sampled_at: utc(),
        sampled_elapsed_ms: view.sampled_elapsed_ms + Math.max(0, Math.floor(now() - viewReceivedAt)),
        last_transition_seq: view.last_transition_seq,
        publication_mode: pauses ? "paused_measurement" : "active",
        phase: view.phase, availability: "available", unavailable_reason: null,
        targets, units: view.units,
      };
      let bytes = `${JSON.stringify(snapshot)}\n`;
      if (Buffer.byteLength(bytes) > maximumBytes) {
        delete snapshot.targets;
        delete snapshot.units;
        snapshot.availability = "unavailable";
        snapshot.unavailable_reason = "snapshot_too_large";
        bytes = `${JSON.stringify(snapshot)}\n`;
      }
      validateLiveSnapshot(snapshot, manifest);
      write(path.join(runRoot, "diagnostics/live-status.json"), bytes, {
        replace: true, maximumReplacementBytes: liveSnapshotBytes,
      });
      revision = snapshot.revision;
      publicationError = null;
    } catch (error) { publicationError = error; }
  }
  function schedule() {
    clear();
    if (stopped || pauses) return;
    const interval = dirty ? 1000 : 5000;
    timer = timers.setTimeout(() => { timer = null; publish(); schedule(); }, Math.max(0, interval - (now() - lastPublication)));
    timer?.unref?.();
  }
  publish();
  schedule();
  return {
    update(next) {
      if (stopped) return;
      viewReceivedAt = now();
      // Scheduler elapsed time begins after publisher admission/setup. Bind its
      // epoch once so neither sample time nor execution durations move backward.
      schedulerOffset ??= Math.max(0, Math.floor(viewReceivedAt - started) - next.sampled_elapsed_ms);
      view = {
        ...next, sampled_elapsed_ms: next.sampled_elapsed_ms + schedulerOffset,
        units: next.units.map((unit) => ({ ...unit, started_elapsed_ms: unit.started_elapsed_ms === null ? null : unit.started_elapsed_ms + schedulerOffset })),
      };
      dirty = true;
      schedule();
    },
    pause() {
      if (stopped) return;
      clear();
      if (pauses++ === 0) publish();
    },
    resume() {
      if (stopped || pauses === 0) return;
      if (--pauses === 0) { publish(); schedule(); }
    },
    stop() {
      if (stopped) return;
      stopped = true;
      clear();
      if (!pauses) publish();
    },
    get publicationError() { return publicationError; },
  };
}
