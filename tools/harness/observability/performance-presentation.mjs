import { validateSchemaSync } from "../contract/index.mjs";
import { formatResourceScopes } from "./resource-presentation.mjs";

export function formatPerformanceExplanation(value, { json = false, resourcesOnly = false } = {}) {
  validateSchemaSync(value.schema_id, value);
  const serialized = `${JSON.stringify(value)}\n`;
  if (Buffer.byteLength(serialized) > 262144) throw new Error("performance explanation exceeds byte limit");
  if (json) return serialized;
  else {
    const lines = [`target=${value.target} status=${value.status} canonical_wall_ms=${value.canonical_wall_ms} inclusive_wall_ms=${value.inclusive_wall_ms} exclusive_wall_ms=${value.exclusive_wall_ms} dependency_path_ms=${value.dependency_path_ms}`];
    if (value.terminal_receipt) {
      const receipt = value.terminal_receipt;
      lines.push(`[INVOCATION] identity=${receipt.invocation_digest} outer_duration_ms=${receipt.outer_duration_ms} child_exit_code=${receipt.child_exit_code} outcome=${receipt.outcome} graphs=${receipt.graph_memberships.length} shared_outer_duration=1 publication_start_ms=${receipt.publication_start_ms}`);
    } else lines.push("[INVOCATION] terminal_receipt=unavailable");
    if (!resourcesOnly) {
      lines.push(`phase_union_ms=${JSON.stringify(value.phase_union_ms)} uncovered_ms=${value.canonical_time_outside_selected_phases_ms} overlap_ms=${value.phase_overlap_ms}`);
      for (const id of value.not_launched_units) lines.push(`[UNIT] ${id} launch=not_launched`);
      for (const absence of value.evidence_absences) lines.push(`[EVIDENCE] ${absence.reference} availability=${absence.reason}`);
      for (const unit of value.units) lines.push(`[UNIT] ${unit.unit_id} queue_ms=${unit.queue_ms} execution_ms=${unit.execution_ms} queue_by_reason_ms=${JSON.stringify(unit.queue_by_reason_ms)} wait_segments=${unit.wait_segments} invocation_path=${unit.on_invocation_dependency_path}`);
    }
    if (value.comparison) lines.push(`[COMPARISON] ${JSON.stringify(value.comparison)}`);
    const resources = value.resources;
    lines.push(`[RESOURCES] availability=${resources.availability} coverage=${resources.coverage ?? "not_observed"} completeness=${resources.completeness ?? "not_observed"} samples=${resources.samples ?? "not_observed"}`);
    for (const unit of resources.unit_cpu_lower_bounds ?? []) lines.push(`[UNIT-CPU] ${unit.unit_id} observed_attributed_cpu_us=${unit.observed_attributed_cpu_us} coverage=lower_bound`);
    lines.push(...formatResourceScopes(resources));
    lines.push(`limits=${value.limitations.join(",")}`);
    const execution = value.execution;
    lines.push(`[EXECUTION] availability=${execution.availability} completeness=${execution.completeness ?? "not_observed"} omitted=${execution.omitted_records ?? "unknown"} scope=whole_invocation local_activities_are_not_additive=1`);
    for (const name of ["launches", "allocations", "leases", "activities", "relationships"]) {
      for (const value of execution[name]?.records ?? []) {
        const record = value.fact ?? value;
        const details = name === "activities" ? `activity=${record.activity} outcome=${record.outcome} duration_ms=${record.duration_ms ?? "unavailable"} clock=${record.clock} clock_identity=${record.clock_identity ?? "unavailable"} resolution_ms=${record.resolution_ms ?? "unavailable"}`
          : name === "relationships" ? `attribution=${value.attribution} provenance=${record.provenance} proof=${record.identity_digest}`
          : `outcome=${record.outcome ?? "leased"} parent=${record.parent_invocation_id ?? "none"} ownership=${record.ownership ?? "unavailable"} launched=${record.launched ?? "unavailable"}`;
        lines.push(`[${name.toUpperCase()}] ${record.ref} unit=${record.unit_id ?? "context"} invocation=${record.invocation_id ?? "none"} allocation=${record.allocation_ref ?? "none"} ${details}`);
      }
      if (execution[name]?.omitted) lines.push(`[OMITTED] ${name}=${execution[name].omitted}`);
    }
    for (const placement of execution.placements ?? []) lines.push(`[PLACEMENT] ${placement.activity_ref} availability=${placement.availability} start_bounds_ms=${JSON.stringify(placement.start_bounds_ms)} end_bounds_ms=${JSON.stringify(placement.end_bounds_ms)} uncertainty_ms=${placement.uncertainty_ms ?? "unavailable"}`);
    const selected = []; let bytes = 0;
    for (const line of lines) { if (selected.length >= 119 || bytes + Buffer.byteLength(line) + 1 > 7900) break; selected.push(line); bytes += Buffer.byteLength(line) + 1; }
    if (selected.length < lines.length) selected.push(`[OMITTED] lines=${lines.length - selected.length}`);
    return `${selected.join("\n")}\n`;
  }
}
