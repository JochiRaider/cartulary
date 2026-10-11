export const median = (values) => {
  if (!values.length) throw Error("empty qualification window");
  const sorted = [...values].sort((a, b) => a - b), middle = Math.floor(sorted.length / 2);
  return sorted.length % 2 ? sorted[middle] : (sorted[middle - 1] + sorted[middle]) / 2;
};
export const mean = (values) => {
  if (!values.length) throw Error("empty qualification window");
  return values.reduce((sum, value) => sum + value, 0) / values.length;
};
export const mad = (values) => { const center = median(values); return median(values.map((value) => Math.abs(value - center))); };
export const p90 = (values) => [...values].sort((a, b) => a - b)[Math.ceil(values.length * 0.9) - 1];
export function intervalUnion(intervals) {
  let end = 0, sum = 0;
  for (const entry of [...intervals].sort((a, b) => a.start_ms - b.start_ms)) {
    if (entry.end_ms < entry.start_ms) throw Error("reversed finalization interval");
    sum += Math.max(0, entry.end_ms - Math.max(end, entry.start_ms)); end = Math.max(end, entry.end_ms);
  }
  return sum;
}
export function eligible(result) {
  const m = result.measurement;
  return result.exit_code === 0 && result.failure === null && m && Object.values(m.coverage).every(Boolean) &&
    [m.cpu_error_us, m.rss_error_bytes, m.outer_ms, m.outer_error_ms, m.canonical_ms, m.steady_rss_bytes, m.sampled_peak_rss_bytes, m.workload_cpu_us, m.reader_cpu_us].every(Number.isFinite) &&
    m.cpu_error_us !== null && m.rss_error_bytes !== null && m.outer_ms > m.outer_error_ms && m.canonical_ms > 0 && m.steady_rss_bytes <= m.sampled_peak_rss_bytes &&
    m.memory_samples >= 2 && m.maximum_memory_gap_ms <= 200 && m.finalization_intervals.every((entry) => entry.start_ms <= entry.end_ms && entry.end_ms <= m.outer_ms);
}
export function compareCell(reference, candidate, budgets, finalization = candidate) {
  if (reference.length !== candidate.length || !candidate.length || [...reference, ...candidate].some((entry) => !eligible(entry))) return { eligible: false, passed: false, failures: ["ineligible"] };
  const pairs = candidate.map((value) => {
    const baseline = reference.find((entry) => entry.block === value.block); if (!baseline) throw Error("unmatched qualification block");
    const a = baseline.measurement, b = value.measurement, same = baseline === value;
    return { canonical: (b.canonical_ms - a.canonical_ms) / a.canonical_ms * 100, outer: (b.outer_ms - a.outer_ms) / a.outer_ms * 100,
      cpu: (b.workload_cpu_us + b.reader_cpu_us - a.workload_cpu_us - a.reader_cpu_us) / (a.outer_ms * 10),
      cpuError: same ? 0 : (a.cpu_error_us + b.cpu_error_us) / ((a.outer_ms - a.outer_error_ms) * 10) + Math.abs(b.workload_cpu_us + b.reader_cpu_us - a.workload_cpu_us - a.reader_cpu_us) * a.outer_error_ms / (10 * a.outer_ms * (a.outer_ms - a.outer_error_ms)),
      rss: b.steady_rss_bytes - a.steady_rss_bytes, rssError: same ? 0 : a.rss_error_bytes + b.rss_error_bytes };
  });
  const values = (key) => pairs.map((pair) => pair[key]);
  const stats = { canonical_median_percent: median(values("canonical")), canonical_p90_percent: p90(values("canonical")),
    outer_median_percent: median(values("outer")), outer_p90_percent: p90(values("outer")),
    cpu_mean_percent_one_core: mean(values("cpu")),
    cpu_median_percent_one_core: median(values("cpu")),
    cpu_mad_percent_one_core: mad(values("cpu")),
    cpu_error_bound_percent_one_core: Math.max(...values("cpuError")),
    cpu_upper_percent_one_core: median(values("cpu")) + 3 * mad(values("cpu")) + Math.max(...values("cpuError")),
    rss_upper_bytes: median(values("rss")) + 3 * mad(values("rss")) + Math.max(...values("rssError")),
    finalization_upper_ms: Math.max(...finalization.filter(eligible).map((entry) => intervalUnion(entry.measurement.finalization_intervals))) };
  if (Object.values(stats).some((value) => !Number.isFinite(value))) return { eligible: false, passed: false, failures: ["unbounded_arithmetic"] };
  const limits = { canonical_median_percent: budgets.elapsed_median_percent, canonical_p90_percent: budgets.elapsed_p90_percent,
    outer_median_percent: budgets.elapsed_median_percent, outer_p90_percent: budgets.elapsed_p90_percent,
    rss_upper_bytes: budgets.steady_rss_bytes, finalization_upper_ms: budgets.finalization_ms };
  const failures = Object.keys(limits).filter((name) => stats[name] > limits[name]);
  // CPU uses a strict arithmetic-mean gate. Dispersion/error remain reported
  // diagnostics; RSS and elapsed keep their independently adopted estimands.
  if (stats.cpu_mean_percent_one_core >= budgets.cpu_mean_percent_one_core) failures.push("cpu_mean_percent_one_core");
  return { eligible: true, passed: failures.length === 0, failures, stats, signed_pairs: pairs };
}
export function assessWorkload(attempts, plan) {
  const measured = attempts.filter((entry) => entry.phase === "measured"), reference = measured.filter((entry) => entry.mode === "off" && entry.readers === 0).sort((a, b) => a.block - b.block);
  const cells = [], incrementalCollection = [];
  let unstable = false;
  for (const mode of plan.modes) for (const readers of plan.reader_counts) {
    const candidate = measured.filter((entry) => entry.mode === mode && entry.readers === readers).sort((a, b) => a.block - b.block);
    const nonMeasured = attempts.filter((entry) => entry.phase !== "measured" && entry.mode === mode && entry.readers === readers);
    const assessment = candidate.length >= 5 && reference.length >= 5
      ? compareCell(reference, candidate, plan.budgets, [...candidate, ...nonMeasured])
      : { eligible: false, passed: false, failures: ["incomplete_blocks"] };
    if (assessment.eligible) {
      for (const window of [reference, candidate]) for (const field of ["canonical_ms", "outer_ms"]) {
        const values = window.map((entry) => entry.measurement[field]); if (mad(values) > median(values) * 0.1) unstable = true;
      }
      if (candidate.length >= 5) for (const entry of candidate) {
        const reduced = compareCell(reference.filter((value) => value.block !== entry.block), candidate.filter((value) => value.block !== entry.block), plan.budgets, [...candidate.filter((value) => value.block !== entry.block), ...nonMeasured]);
        // Every individual gate must retain its disposition, not only the combined result.
        if (reduced.failures.join() !== assessment.failures.join()) unstable = true;
      }
    }
    cells.push({ mode, readers, ...assessment });
    if (mode === "basic") incrementalCollection.push({ readers, ...compareCell(measured.filter((entry) => entry.mode === "off" && entry.readers === readers), candidate, plan.budgets) });
  }
  return { blocks: reference.length, unstable, needs_sixth: reference.length === 5 && unstable,
    passed: [5, 6].includes(reference.length) && !unstable && cells.every((cell) => cell.passed), cells, incremental_collection: incrementalCollection };
}
