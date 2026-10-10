import { intervalUnion } from "../evidence-accounting/index.mjs";
import { analyzeExecution } from "./execution-analysis.mjs";

export function analyzePerformance(run, target, { comparison, resources, envelope, execution = null }) {
  const summary = run.targetSummaries.get(target || run.manifest.target);
  if (!summary) throw new Error("selected target has no canonical target summary");
  const members = new Set(summary.unit_ids);
  const phases = [...run.phases.values()].filter((phase) => members.has(phase.unit_id));
  const byPhase = new Map();
  for (const phase of phases) {
    if (!byPhase.has(phase.phase)) byPhase.set(phase.phase, []);
    if (phase.end_ms !== null) byPhase.get(phase.phase).push({ start: phase.start_ms, end: phase.end_ms });
  }
  const selectedIntervals = [...run.intervals].filter(([id]) => members.has(id));
  const units = selectedIntervals.map(([id, interval]) => ({ unit_id: id, kind: run.registrations.get(id).kind,
    queue_ms: interval.queue_ms, execution_ms: interval.end - interval.start,
    on_invocation_dependency_path: run.summary.critical_path.includes(id) }))
    .sort((a, b) => (b.queue_ms + b.execution_ms) - (a.queue_ms + a.execution_ms) || a.unit_id.localeCompare(b.unit_id));
  const covered = intervalUnion(phases.filter((phase) => phase.end_ms !== null).map((phase) => ({ start: phase.start_ms, end: phase.end_ms })));
  return {
    comparison, schema_id: "cartulary.harness_performance_explanation.v2", run_id: run.manifest.run_id,
    execution: analyzeExecution(execution),
    target: summary.target, status: summary.status, canonical_wall_ms: run.summary.wall_duration_ms,
    inclusive_wall_ms: summary.inclusive_wall_ms, exclusive_wall_ms: summary.exclusive_wall_ms,
    dependency_path_ms: summary.actual_dependency_critical_path_ms,
    invocation_path_queue_ms: run.summary.critical_path.reduce((sum, id) => sum + (run.intervals.get(id)?.queue_ms ?? 0), 0),
    phase_union_ms: Object.fromEntries([...byPhase].sort(([a], [b]) => a.localeCompare(b)).map(([name, values]) => [name, intervalUnion(values)])),
    canonical_time_outside_selected_phases_ms: Math.max(0, run.summary.wall_duration_ms - covered),
    phase_overlap_ms: [...byPhase.values()].reduce((sum, values) => sum + intervalUnion(values), 0) - covered,
    units: units.slice(0, 20), units_omitted: Math.max(0, units.length - 20),
    reservations: { ...run.summary.resource_pressure,
      blocked_units: run.summary.resource_pressure.blocked_units.slice(0, 20),
      resource_holders: run.summary.resource_pressure.resource_holders.slice(0, 20),
      blocked_units_omitted: Math.max(0, run.summary.resource_pressure.blocked_units.length - 20),
      resource_holders_omitted: Math.max(0, run.summary.resource_pressure.resource_holders.length - 20) }, resources, envelope,
    limitations: ["unit_execution_includes_harness_work", "phase_categories_may_overlap", "reservations_are_not_consumption", "resource_coverage_is_observed_partial", "guest_is_not_windows_host"],
  };
}
