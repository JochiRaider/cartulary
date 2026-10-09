import { canonicalJSONString } from "../contract/index.mjs";

const same = (a, b) => canonicalJSONString(a) === canonicalJSONString(b);
const selection = (manifest) => Object.fromEntries(Object.entries(manifest.declared_inputs)
  .filter(([name]) => name !== "HARNESS_DIAGNOSTICS"));

// This is a descriptive compatibility check. It cannot establish matching
// unobserved host activity, application caches, or service history.
export function compareRunMetadata(current, reference, mode = "equivalent") {
  if (!["equivalent", "instrumentation"].includes(mode)) throw new Error("invalid comparison mode");
  const reasons = [];
  for (const key of ["command_id", "target", "source_commit", "source_state", "source_digest", "graph_digest", "toolchain_digest", "system_digest", "cache_mode"]) {
    if (!same(current.manifest[key], reference.manifest[key])) reasons.push(key);
  }
  if (!same(selection(current.manifest), selection(reference.manifest))) reasons.push("declared_inputs");
  if (!same(current.manifest.capability_snapshot, reference.manifest.capability_snapshot)) reasons.push("capability_snapshot");
  if (current.manifest.instrumentation.policy_digest !== reference.manifest.instrumentation.policy_digest) reasons.push("instrumentation_policy");
  if (mode === "equivalent" && current.manifest.instrumentation.mode !== reference.manifest.instrumentation.mode) reasons.push("instrumentation_mode");
  if (mode === "instrumentation" && current.manifest.instrumentation.mode === reference.manifest.instrumentation.mode) reasons.push("missing_treatment_difference");
  if (!same(current.summary.cache, reference.summary.cache)) reasons.push("observed_cache_counts");
  const outcomes = (run) => [...run.terminal].map(([id, event]) => [id, event.status, event.cache_status ?? null]).sort(([a], [b]) => a.localeCompare(b));
  if (!same(outcomes(current), outcomes(reference))) reasons.push("unit_outcomes");
  if (current.summary.status !== "pass" || reference.summary.status !== "pass") reasons.push("incomplete_or_failed_run");
  return {
    mode, reference_run_id: reference.manifest.run_id,
    compatible: reasons.length === 0, reasons, qualification: "descriptive_only",
    current_mode: current.manifest.instrumentation.mode, reference_mode: reference.manifest.instrumentation.mode,
    current_wall_ms: current.summary.wall_duration_ms, reference_wall_ms: reference.summary.wall_duration_ms,
    wall_delta_ms: reasons.length ? null : current.summary.wall_duration_ms - reference.summary.wall_duration_ms,
    limitations: ["one_pair_is_not_a_performance_gate", "unobserved_host_activity_and_service_history", "same_source_required", "canonical_phase_cost_is_present_in_both_modes"],
  };
}
