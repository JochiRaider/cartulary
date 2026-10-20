import { parseStrictJSON } from "../contract/strict-json.mjs";
import { resourceCapacities } from "../scheduler/work-graph/index.mjs";
import path from "node:path";

import { assertRetainedArtifactFile, readRetainedArtifactBytes, validateSchemaSync } from "../contract/index.mjs";
import { reduceCanonicalUnitIntervals, notLaunchedUnitIDs } from "../evidence-accounting/index.mjs";

import { actualCriticalPath, intervalUnion, canonicalTimingAccounting, projectResourcePressure } from "../evidence-accounting/index.mjs";

function readJSON(file) {
  return parseStrictJSON(readRetainedArtifactBytes(path.dirname(file), path.basename(file), 16 * 1024 * 1024).toString("utf8"), "canonical evidence");
}

export async function validateCanonicalRun(runRoot, expectedTarget = "", { evidence = true } = {}) {
  const files = {
    manifest: path.join(runRoot, "run-manifest.json"),
    summary: path.join(runRoot, "run-summary.json"),
    events: path.join(runRoot, "unit-events.ndjson"),
  };
  const manifest = readJSON(files.manifest);
  const summary = readJSON(files.summary);
  validateSchemaSync("cartulary.harness_run_manifest.v3", manifest);
  validateSchemaSync("cartulary.harness_run_summary.v1", summary);
  if (manifest.run_id !== summary.run_id || manifest.target !== summary.target) {
    throw new Error(`${runRoot} manifest and summary identity do not close`);
  }
  if (expectedTarget && manifest.target !== expectedTarget) {
    throw new Error(`${runRoot} target ${manifest.target} does not match ${expectedTarget}`);
  }
  const eventState = await reduceCanonicalUnitIntervals(files.events);
  const terminal = eventState.terminals;
  const started = eventState.starts;
  const runStarted = eventState.runStarted;
  const runCompleted = eventState.runCompleted;
  const graph = { units: [...eventState.registrations.values()] };
  if (graph.units.length !== terminal.size) throw new Error("canonical registration roster does not close");
  const intervals = eventState.intervals;
  const criticalPath = actualCriticalPath(graph, intervals);
  const pathDuration = (ids) => ids.reduce((total, id) => {
    const interval = intervals.get(id);
    return total + (interval ? interval.end - interval.start + interval.queue_ms : 0);
  }, 0);
  const same = (left, right, label) => {
    const normalize = (v) => Array.isArray(v) ? v.map(normalize) : v && typeof v === "object" ? Object.fromEntries(Object.keys(v).sort().map((k) => [k, normalize(v[k])])) : v;
    if (JSON.stringify(normalize(left)) !== JSON.stringify(normalize(right))) throw new Error(`${label} does not reconstruct from canonical events`);
  };
  same(summary.critical_path, criticalPath, "critical path");
  same(summary.actual_dependency_critical_path_ms, pathDuration(criticalPath), "critical duration");
  same(summary.timing_accounting, canonicalTimingAccounting(eventState, graph, summary.wall_duration_ms), "timing accounting");
  same(summary.resource_pressure, projectResourcePressure(eventState, graph, Object.fromEntries(resourceCapacities(manifest.capability_snapshot))), "reservation pressure");
  if ([...eventState.phases.values()].some((phase) => phase.end_ms === null)) throw new Error("incomplete phase in finalized run");
  const counts = summary.unit_counts;
  const terminalCounts = { passed: 0, failed: 0, skipped: 0, cancelled: 0 };
  for (const event of terminal.values()) terminalCounts[event.status] += 1;
  if (
    terminal.size !== counts.total ||
    counts.passed + counts.failed + counts.skipped + counts.cancelled !== counts.total ||
    Object.entries(terminalCounts).some(([status, count]) => counts[status] !== count)
  ) {
    throw new Error(`${files.summary} unit roster does not close against terminal events`);
  }
  if (!runStarted || !runCompleted || runStarted.monotonic_ms !== 0) {
    throw new Error(`${files.events} does not close the canonical run interval`);
  }
  if (
    ({ pass: "passed", fail: "failed", cancelled: "cancelled" })[summary.status] !== runCompleted.status ||
    summary.wall_duration_ms !== runCompleted.monotonic_ms ||
    runCompleted.seq !== eventState.eventCount
  ) {
    throw new Error(`${files.summary} wall duration does not equal the canonical run interval`);
  }
  for (const artifact of evidence ? summary.artifact_refs : []) {
    assertRetainedArtifactFile(runRoot, artifact);
  }
  // Unit results are always required. Their declared outputs bind an absence
  // to canonical dispatch facts, rather than interpreting a filesystem error
  // or a failure string as evidence that work never started.
  const unitEvidence = new Map(), outputOwners = new Map();
  const notLaunched = notLaunchedUnitIDs(eventState);
  const unavailable = new Set(notLaunched);
  for (let changed = true; changed;) {
    changed = false;
    for (const [id, unit] of eventState.registrations) {
      if (!unavailable.has(id) && ["finalizer", "projection"].includes(unit.kind) && terminal.get(id).status !== "passed" && unit.needs.some((need) => unavailable.has(need))) {
        unavailable.add(id); changed = true;
      }
    }
  }
  const evidenceAbsences = new Map();
  const targetSummaries = new Map();
  for (const artifact of summary.artifact_refs.filter((value) => value.startsWith("target-summaries/"))) {
    const targetSummary = readJSON(path.join(runRoot, artifact));
    validateSchemaSync("cartulary.harness_target_summary.v1", targetSummary);
    if (targetSummary.unit_ids.some((unitID) => !terminal.has(unitID))) {
      throw new Error(`${artifact} references a unit outside the canonical terminal roster`);
    }
    for (const reference of evidence ? targetSummary.evidence_refs.filter((ref) => ref.startsWith("unit-results/")) : []) {
      if (unitEvidence.has(reference)) continue;
      const result = readJSON(path.join(runRoot, reference));
      validateSchemaSync("cartulary.harness_unit_result.v1", result);
      if (!targetSummary.unit_ids.includes(result.unit_id) || result.status !== terminal.get(result.unit_id)?.status) throw new Error("unit evidence contradicts canonical outcome");
      unitEvidence.set(reference, result);
      for (const output of result.evidence_outputs) {
        if (!outputOwners.has(output)) outputOwners.set(output, new Set());
        outputOwners.get(output).add(result.unit_id);
      }
    }
    for (const reference of evidence ? targetSummary.evidence_refs : []) {
      try { assertRetainedArtifactFile(runRoot, reference); }
      catch (error) {
        const owners = [...(outputOwners.get(reference) ?? [])];
        if (error.code !== "ENOENT" || /^(?:unit-results|rows)\//u.test(reference) || !owners.length || !owners.every((id) => unavailable.has(id))) throw error;
        evidenceAbsences.set(reference, { reference, unit_ids: owners.sort(), reason: owners.every((id) => notLaunched.has(id)) ? "not_launched" : "dependency_unavailable" });
      }
    }
    const selectedIntervals = targetSummary.unit_ids
      .filter((unitID) => started.has(unitID) && terminal.has(unitID))
      .map((unitID) => ({ start: started.get(unitID), end: terminal.get(unitID).monotonic_ms }));
    if (targetSummary.inclusive_wall_ms !== intervalUnion(selectedIntervals)) {
      throw new Error(`${artifact} inclusive interval union does not close`);
    }
    same(targetSummary.actual_dependency_critical_path_ms, pathDuration(actualCriticalPath(graph, intervals, targetSummary.unit_ids)), "target critical duration");
    same(targetSummary.timing_accounting, canonicalTimingAccounting(eventState, graph, targetSummary.inclusive_wall_ms, { includeRunEnvelope: false, selectedUnitIDs: targetSummary.unit_ids }), "target accounting");
    if (targetSummaries.has(targetSummary.target)) throw new Error("duplicate target summary");
    if (artifact !== `target-summaries/${targetSummary.target}.json`) throw new Error("target summary path mismatch");
    targetSummaries.set(targetSummary.target, targetSummary);
  }
  for (const target of targetSummaries.values()) {
    const childUnits = new Set(target.children.flatMap((child) => {
      if (!targetSummaries.has(child)) throw new Error("missing target child");
      const ids = targetSummaries.get(child).unit_ids;
      if (ids.some((id) => !target.unit_ids.includes(id))) throw new Error("invalid target child membership");
      return ids;
    }));
    same(target.exclusive_wall_ms, intervalUnion(target.unit_ids.filter((id) => !childUnits.has(id)).map((id) => intervals.get(id)).filter(Boolean)), "exclusive interval union");
  }
  return {
    notLaunched: [...notLaunched].sort(), evidenceAbsences: [...evidenceAbsences.values()],
    waits: eventState.waits, phases: eventState.phases, registrations: eventState.registrations, intervals,
    eventCount: eventState.eventCount,
    manifest,
    summary,
    terminal,
    started,
    targetSummaries,
  };
}
