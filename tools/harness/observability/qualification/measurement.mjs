import assert from "node:assert/strict";
// The estimand is the time-weighted sampled cohort RSS, not an unsampled peak or
// an assertion about allocation between samples. Each population is counted once.
export function summarizeMemory(samples, outer, { bounded = false, pageBytes = 0 } = {}) {
  assert.ok(Number.isFinite(outer) && outer > 0);
  const left = outer * 0.1, right = outer * 0.9;
  let integral = 0, coverage = 0, maximumGap = 0, maximumCount = 0, peak = null, previous = -1, timingErrorIntegral = 0, available = true;
  for (const [index, sample] of samples.entries()) {
    assert.ok(Number.isFinite(sample.at_ms) && sample.at_ms >= previous && sample.at_ms <= outer);
    assert.ok(Number.isSafeInteger(sample.count) && sample.count >= 0 && Number.isFinite(sample.span_ms) && sample.span_ms >= 0 && sample.span_ms <= sample.at_ms);
    assert.ok(Number.isSafeInteger(sample.transition_bytes) && sample.transition_bytes >= 0);
    assert.ok(["available", "unavailable"].includes(sample.availability));
    assert.equal(sample.rss === null, sample.availability === "unavailable");
    previous = sample.at_ms;
    const end = samples[index + 1]?.at_ms ?? outer, start = Math.max(left, sample.at_ms), stop = Math.min(right, end);
    const span = Math.max(0, Math.min(right, sample.at_ms) - Math.max(left, sample.at_ms - sample.span_ms));
    if (stop > start || span > 0) maximumCount = Math.max(maximumCount, sample.count);
    if (sample.availability === "unavailable") { if (stop > start || span > 0) available = false; continue; }
    assert.ok(Number.isSafeInteger(sample.rss) && sample.rss >= 0);
    peak = Math.max(peak ?? 0, sample.rss);
    const prior = index ? samples[index - 1].rss : 0;
    if (prior !== null) assert.ok(sample.transition_bytes >= Math.abs(sample.rss - prior) && sample.transition_bytes <= sample.rss + prior, "resident transition bound mismatch");
    // Per-identity held values can update anywhere within this sweep. Summed
    // absolute transitions retain opposing allocations and process replacement.
    if (prior === null && span > 0) available = false;
    timingErrorIntegral += sample.transition_bytes * span;
    if (stop > start) { integral += sample.rss * (stop - start); coverage += stop - start; maximumGap = Math.max(maximumGap, end - sample.at_ms + sample.span_ms); }
  }
  const count = samples.filter((sample) => sample.availability === "available" && sample.at_ms >= left && sample.at_ms <= right).length;
  const covered = available && Math.abs(coverage - (right - left)) < 0.00001;
  return { steady_rss_bytes: covered ? integral / (right - left) : null, sampled_peak_rss_bytes: peak, memory_samples: count,
    maximum_memory_gap_ms: maximumGap, complete: covered && count >= 2 && maximumGap <= 200,
    // This bounds timing of the sampled hold-value estimator. It makes no claim
    // about physical allocations between samples or an unsampled peak.
    rss_error_bytes: covered && bounded && Number.isSafeInteger(pageBytes) && pageBytes > 0 ? maximumCount * pageBytes + timingErrorIntegral / (right - left) : null, maximum_processes: maximumCount };
}
export function finalizationIntervals(raw, retained, mode) {
  if (mode === "off") return { intervals: [], complete: true };
  const receipt = retained.terminalReceipt, envelope = retained.envelope;
  if (!receipt || !envelope || raw.envelope_seen_ms < 0) return { intervals: [], complete: false };
  const phase = envelope.intervals.find((entry) => entry.phase === "collector_shutdown");
  if (!phase) return { intervals: [], complete: false };
  const duration = phase.end_ms - phase.start_ms;
  // The scan-bound envelope can appear at any time between these external
  // observations. The resulting interval includes its publication and drain.
  const drain = { start_ms: Math.max(0, raw.envelope_previous_ms - duration), end_ms: raw.envelope_seen_ms };
  // Outer time before wrapper entry is conservatively included. BOOTTIME's
  // published quantization is 10 ms; receipt publication ends before return.
  const publication = { start_ms: Math.max(0, receipt.outer_duration_ms - 10), end_ms: raw.outer_ms };
  if (drain.end_ms > raw.outer_ms || publication.start_ms > publication.end_ms) return { intervals: [], complete: false };
  return { intervals: [drain, publication], complete: true };
}
export function assembleMeasurement(raw, retained, mode) {
  if (mode === "basic") requireBasicTreatment(retained);
  const memory = summarizeMemory(raw.samples, raw.outer_ms, { bounded: raw.rss_bounded, pageBytes: raw.page_bytes }), finalization = finalizationIntervals(raw, retained, mode);
  return { outer_ms: raw.outer_ms, canonical_ms: retained.run.summary.wall_duration_ms,
    workload_cpu_us: raw.workload_cpu_us, reader_cpu_us: raw.reader_cpu_us, cpu_error_us: raw.cpu_error_us,
    steady_rss_bytes: memory.steady_rss_bytes, sampled_peak_rss_bytes: memory.sampled_peak_rss_bytes,
    rss_error_bytes: memory.rss_error_bytes, outer_error_ms: raw.outer_error_ms,
    memory_samples: memory.memory_samples, maximum_memory_gap_ms: memory.maximum_memory_gap_ms,
    finalization_intervals: finalization.intervals,
    coverage: { waited_descendants: raw.waited_descendants, reader_lifecycle: raw.reader_lifecycle,
      memory_samples: memory.complete, bounded_error: memory.rss_error_bytes !== null, exclusive_quietness: raw.reader_lifecycle, complete_finalization: finalization.complete },
    support_cpu_us: raw.support_cpu_us, support_cpu_error_us: raw.support_cpu_error_us, support_peak_rss_bytes: raw.support_peak_rss_bytes, support_read_bytes: raw.support_read_bytes, support_write_bytes: raw.support_write_bytes };
}
export function measurementFailure(raw, measurement) {
  if (raw.cancelled) return "cancelled";
  if (raw.exit_code !== 0) return "command_failed";
  if (!measurement.coverage.waited_descendants || !measurement.coverage.reader_lifecycle || !measurement.coverage.complete_finalization) return "measurement_lost";
  if (!measurement.coverage.memory_samples) return "insufficient_samples";
  if (!measurement.coverage.bounded_error) return "unbounded_error";
  return null;
}
export function requireBasicTreatment(retained) {
  // A cheap failed/absent observer is not the requested basic treatment, even
  // when external process accounting and a terminal receipt are complete.
  assert.ok(retained.execution && retained.envelope && retained.terminalReceipt &&
    retained.resourceIndex?.failed_sweeps === 0 &&
    retained.resourceIndex.containers.deadline_overruns === 0 &&
    retained.resourceIndex.containers.availability !== "collector_failed" &&
    retained.resources?.mode === "basic" && retained.resources.availability === "available" &&
    ["complete", "partial"].includes(retained.resources.completeness), "basic observer treatment is incomplete");
}
