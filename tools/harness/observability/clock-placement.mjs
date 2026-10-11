// Pure interpretation of producer-owned clock evidence. No UTC, filesystem or
// fallback ordering is available here, and local duration is never recomputed.
function offsets(anchors, identity) {
  if (anchors?.length !== 2) return { failure: "missing_anchor" };
  const [first, last] = anchors;
  if (anchors.some((a) => a.clock_identity !== identity || a.boot_identity !== first.boot_identity)) return { failure: "incompatible_clock" };
  if (anchors.some((a) => a.local_after_ms < a.local_before_ms) || last.local_before_ms < first.local_after_ms || last.boot_ms < first.boot_ms) return { failure: "discontinuity" };
  const lower = Math.max(...anchors.map((a) => a.boot_ms - a.local_after_ms));
  const upper = Math.min(...anchors.map((a) => a.boot_ms + a.boot_resolution_ms - a.local_before_ms));
  if (lower > upper) return { failure: "discontinuity" };
  return { lower, upper, identity: first.boot_identity };
}

export function placeActivity(activity, graphClock, maximumUncertaintyMS = 25) {
  const unavailable = (availability) => ({ activity_ref: activity.ref, availability,
    start_bounds_ms: null, end_bounds_ms: null, uncertainty_ms: null });
  if (activity.outcome === "incomplete") return unavailable("incomplete");
  if (activity.availability !== "available" || activity.local_start_ms === null || activity.local_end_ms === null || !graphClock) return unavailable("missing_anchor");
  const producer = offsets(activity.anchors, activity.clock_identity), scheduler = offsets(graphClock.anchors, graphClock.clock_identity);
  if (producer.failure || scheduler.failure) return unavailable(producer.failure ?? scheduler.failure);
  if (producer.identity !== scheduler.identity) return unavailable("incompatible_clock");
  if (activity.local_end_ms < activity.local_start_ms || Math.abs(activity.local_end_ms - activity.local_start_ms - activity.duration_ms) > Math.max(0.00001, 2 * activity.resolution_ms)) return unavailable("discontinuity");
  if (activity.local_start_ms < activity.anchors[0].local_before_ms || activity.local_end_ms > activity.anchors[1].local_after_ms) return unavailable("discontinuity");
  const lower = producer.lower - scheduler.upper, upper = producer.upper - scheduler.lower;
  const width = upper - lower;
  if (width > maximumUncertaintyMS) return unavailable("uncertainty_exceeded");
  const place = (value) => [value + lower - graphClock.epoch_ms, value + upper - graphClock.epoch_ms];
  return { activity_ref: activity.ref, availability: "available", start_bounds_ms: place(activity.local_start_ms),
    end_bounds_ms: place(activity.local_end_ms), uncertainty_ms: width };
}
