export function presentResources(index, mode, { rows, unitCPU, sampledSimultaneousRSS }) {
  return { availability: "available", completeness: index.status, coverage: index.coverage,
    mode, samples: index.samples, sample_bytes: index.sample_bytes,
    omitted_observations: index.omitted_observations, discovery_truncations: index.discovery_truncations,
    sampled_simultaneous_rss_sum_bytes: sampledSimultaneousRSS,
    maximum_sample_sweep_ms: index.observer.maximum_sweep_ms,
    unit_cpu_lower_bounds: [...unitCPU].sort(([a], [b]) => a.localeCompare(b)).slice(0, 20).map(([unit_id, observed_attributed_cpu_us]) => ({ unit_id, observed_attributed_cpu_us })),
    unit_cpu_omitted: Math.max(0, unitCPU.size - 20),
    rss_sum_semantics: "sweep_subset_shared_pages_may_repeat",
    scopes: rows.slice(0, 20), scopes_omitted: Math.max(0, rows.length - 20), observer: index.observer };
}
