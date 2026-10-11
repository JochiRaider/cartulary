import { instrumentationSignals } from "../contract/index.mjs";

// Format only the validated projection. Signal semantics remain catalog-owned;
// the physical roster supplies ownership, never per-unit cost attribution.
export function formatResourceScopes(resources) {
  const containers = new Map((resources.containers?.identities ?? []).map((entry) => [entry.container_ref, entry]));
  const lines = [];
  if (resources.scopes_omitted) lines.push(`[OMITTED] scopes=${resources.scopes_omitted}`);
  if (resources.unit_cpu_omitted) lines.push(`[OMITTED] unit_cpu_lower_bounds=${resources.unit_cpu_omitted}`);
  // Give the separately owned service observations room before verbose process
  // detail reaches the human byte limit. The selected facts themselves do not change.
  const scopes = [...(resources.scopes ?? [])].sort((a, b) => Number(b.scope === "container") - Number(a.scope === "container"));
  for (const scope of scopes) {
    const container = containers.get(scope.scope_ref);
    const fields = [`[SCOPE] ${scope.scope_ref}`, `scope=${scope.scope}`, `unit=${scope.unit_id ?? "context_or_harness"}`,
      `allocation=${scope.allocation_ref ?? container?.allocation_ref ?? "not_observed"}`,
      `samples=${scope.samples}`, `unavailable=${scope.unavailable}`, `resets=${scope.resets}`,
      `cpu_percent_one_core=${scope.cpu_percent_one_core ?? "not_observed"}`];
    if (scope.scope === "container") fields.push(`ownership=${container?.ownership ?? "not_observed"}`, "attribution=contextual");
    for (const [name, metric] of Object.entries(scope.metrics).sort(([a], [b]) => a.localeCompare(b))) {
      const signal = instrumentationSignals[name];
      if (!signal || signal.producer_scope !== scope.scope) continue;
      const unit = name.endsWith(`_${signal.unit}`) ? "" : `(${signal.unit})`;
      if (signal.aggregation === "sampled_maximum") fields.push(`sampled_max_${name}=${metric.maximum ?? "not_observed"}${unit}`);
      else fields.push(`observed_delta_${name}=${metric.observed_delta ?? "not_observed"}${unit}`, `${name}_interval_ms=${metric.observed_interval_ms}`);
      if (metric.unavailable) fields.push(`${name}_unavailable=${metric.unavailable}`);
    }
    lines.push(fields.join(" "));
  }
  return lines;
}

export function presentResources(index, mode, { rows, unitCPU, sampledSimultaneousRSS }) {
  return { availability: "available", completeness: index.status, coverage: index.coverage,
    mode, containers: index.containers, samples: index.samples, sample_bytes: index.sample_bytes,
    omitted_observations: index.omitted_observations, discovery_truncations: index.discovery_truncations,
    sampled_simultaneous_rss_sum_bytes: sampledSimultaneousRSS,
    maximum_sample_sweep_ms: index.observer.maximum_sweep_ms,
    unit_cpu_lower_bounds: [...unitCPU].sort(([a], [b]) => a.localeCompare(b)).slice(0, 20).map(([unit_id, observed_attributed_cpu_us]) => ({ unit_id, observed_attributed_cpu_us })),
    unit_cpu_omitted: Math.max(0, unitCPU.size - 20),
    rss_sum_semantics: "sweep_subset_shared_pages_may_repeat",
    scopes: rows.slice(0, 20), scopes_omitted: Math.max(0, rows.length - 20), observer: index.observer };
}
