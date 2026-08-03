import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function loadEvents(file) {
  return readFileSync(file, "utf8").split(/\r?\n/u).filter(Boolean).map(JSON.parse);
}

function containedArtifact(runRoot, relative) {
  if (path.isAbsolute(relative) || relative.split(/[\\/]/u).includes("..")) {
    throw new Error(`${relative} is not a contained canonical artifact reference`);
  }
  const resolved = path.resolve(runRoot, relative);
  if (resolved !== runRoot && !resolved.startsWith(`${runRoot}${path.sep}`)) {
    throw new Error(`${relative} escapes the canonical run root`);
  }
  return resolved;
}

function intervalUnion(intervals) {
  const sorted = intervals.sort((left, right) => left.start - right.start || left.end - right.end);
  let active = null;
  let total = 0;
  for (const interval of sorted) {
    if (!active || interval.start > active.end) {
      if (active) total += active.end - active.start;
      active = { ...interval };
    } else active.end = Math.max(active.end, interval.end);
  }
  if (active) total += active.end - active.start;
  return total;
}

export function validateCanonicalRun(runRoot, expectedTarget = "") {
  const files = {
    manifest: path.join(runRoot, "run-manifest.json"),
    summary: path.join(runRoot, "run-summary.json"),
    events: path.join(runRoot, "unit-events.ndjson"),
  };
  for (const file of Object.values(files)) {
    if (!existsSync(file)) throw new Error(`${file} is required canonical evidence`);
  }
  const manifest = readJSON(files.manifest);
  const summary = readJSON(files.summary);
  validateSchemaSync("cartulary.harness_run_manifest.v1", manifest);
  validateSchemaSync("cartulary.harness_run_summary.v1", summary);
  if (manifest.run_id !== summary.run_id || manifest.target !== summary.target) {
    throw new Error(`${runRoot} manifest and summary identity do not close`);
  }
  if (expectedTarget && manifest.target !== expectedTarget) {
    throw new Error(`${runRoot} target ${manifest.target} does not match ${expectedTarget}`);
  }
  const events = loadEvents(files.events);
  if (events.length === 0) throw new Error(`${files.events} is empty`);
  const terminal = new Map();
  const started = new Map();
  let runStarted = null;
  let runCompleted = null;
  let previousMs = 0;
  for (const [index, event] of events.entries()) {
    validateSchemaSync("cartulary.harness_unit_event.v1", event);
    if (event.seq !== index + 1 || event.monotonic_ms < previousMs) {
      throw new Error(`${files.events} is not a contiguous monotonic event stream`);
    }
    previousMs = event.monotonic_ms;
    if (event.event === "run_started") {
      if (runStarted) throw new Error(`${files.events} has duplicate run_started events`);
      runStarted = event;
    }
    if (event.event === "run_completed") {
      if (runCompleted) throw new Error(`${files.events} has duplicate run_completed events`);
      runCompleted = event;
    }
    if (event.event === "started") started.set(event.unit_id, event.monotonic_ms);
    if (["completed", "failed", "skipped", "cancelled"].includes(event.event)) {
      if (terminal.has(event.unit_id)) throw new Error(`${files.events} has duplicate terminal event for ${event.unit_id}`);
      terminal.set(event.unit_id, event);
    }
  }
  const counts = summary.unit_counts;
  if (terminal.size !== counts.total || counts.passed + counts.failed + counts.skipped + counts.cancelled !== counts.total) {
    throw new Error(`${files.summary} unit roster does not close against terminal events`);
  }
  if (!runStarted || !runCompleted || runStarted.monotonic_ms !== 0) {
    throw new Error(`${files.events} does not close the canonical run interval`);
  }
  if (summary.wall_duration_ms !== runCompleted.monotonic_ms || runCompleted !== events.at(-1)) {
    throw new Error(`${files.summary} wall duration does not equal the canonical run interval`);
  }
  for (const artifact of summary.artifact_refs) {
    if (!existsSync(containedArtifact(runRoot, artifact))) throw new Error(`${runRoot} is missing declared artifact ${artifact}`);
  }
  const targetSummaries = new Map();
  for (const artifact of summary.artifact_refs.filter((value) => value.startsWith("target-summaries/"))) {
    const targetSummary = readJSON(path.join(runRoot, artifact));
    validateSchemaSync("cartulary.harness_target_summary.v1", targetSummary);
    if (targetSummary.unit_ids.some((unitID) => !terminal.has(unitID))) {
      throw new Error(`${artifact} references a unit outside the canonical terminal roster`);
    }
    for (const evidence of targetSummary.evidence_refs) {
      if (!existsSync(containedArtifact(runRoot, evidence))) {
        throw new Error(`${artifact} is missing declared unit evidence ${evidence}`);
      }
    }
    const intervals = targetSummary.unit_ids
      .filter((unitID) => started.has(unitID) && terminal.has(unitID))
      .map((unitID) => ({ start: started.get(unitID), end: terminal.get(unitID).monotonic_ms }));
    if (targetSummary.inclusive_wall_ms !== intervalUnion(intervals)) {
      throw new Error(`${artifact} inclusive interval union does not close`);
    }
    targetSummaries.set(targetSummary.target, targetSummary);
  }
  return { manifest, summary, events, terminal, started, targetSummaries };
}
