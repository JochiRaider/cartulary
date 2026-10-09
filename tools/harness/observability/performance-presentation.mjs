import { validateSchemaSync } from "../contract/index.mjs";

export function formatPerformanceExplanation(value, { json = false, resourcesOnly = false } = {}) {
  validateSchemaSync(value.schema_id, value);
  const serialized = `${JSON.stringify(value)}\n`;
  if (Buffer.byteLength(serialized) > 262144) throw new Error("performance explanation exceeds byte limit");
  if (json) return serialized;
  else {
    const lines = [`target=${value.target} status=${value.status} canonical_wall_ms=${value.canonical_wall_ms} inclusive_wall_ms=${value.inclusive_wall_ms} exclusive_wall_ms=${value.exclusive_wall_ms} dependency_path_ms=${value.dependency_path_ms}`];
    if (!resourcesOnly) {
      lines.push(`phase_union_ms=${JSON.stringify(value.phase_union_ms)} uncovered_ms=${value.canonical_time_outside_selected_phases_ms} overlap_ms=${value.phase_overlap_ms}`);
      for (const unit of value.units) lines.push(`[UNIT] ${unit.unit_id} queue_ms=${unit.queue_ms} execution_ms=${unit.execution_ms} invocation_path=${unit.on_invocation_dependency_path}`);
    }
    if (value.comparison) lines.push(`[COMPARISON] ${JSON.stringify(value.comparison)}`);
    const resources = value.resources;
    lines.push(`[RESOURCES] availability=${resources.availability} coverage=${resources.coverage ?? "not_observed"} completeness=${resources.completeness ?? "not_observed"} samples=${resources.samples ?? "not_observed"}`);
    for (const unit of resources.unit_cpu_lower_bounds ?? []) lines.push(`[UNIT-CPU] ${unit.unit_id} observed_lifetime_cpu_us=${unit.observed_lifetime_cpu_us} coverage=lower_bound`);
    for (const scope of resources.scopes ?? []) lines.push(`[SCOPE] ${scope.scope_ref} scope=${scope.scope} unit=${scope.unit_id ?? "context_or_harness"} cpu_percent_one_core=${scope.cpu_percent_one_core ?? "not_observed"} sampled_rss_max_bytes=${scope.metrics.rss_bytes?.maximum ?? "not_observed"}`);
    lines.push(`limits=${value.limitations.join(",")}`);
    const selected = []; let bytes = 0;
    for (const line of lines) { if (selected.length >= 119 || bytes + Buffer.byteLength(line) + 1 > 7900) break; selected.push(line); bytes += Buffer.byteLength(line) + 1; }
    if (selected.length < lines.length) selected.push(`[OMITTED] lines=${lines.length - selected.length}`);
    return `${selected.join("\n")}\n`;
  }
}
