import assert from "node:assert/strict";
import test from "node:test";
import { captureClockAnchor } from "../../runtime/boot-clock.mjs";
import { placeActivity } from "../clock-placement.mjs";

const boot = `sha256:${"1".repeat(64)}`;
const anchor = (id, before, after, value) => ({ clock_identity: id, boot_identity: boot,
  local_before_ms: before, local_after_ms: after, boot_ms: value, boot_resolution_ms: 10 });
const scheduler = () => ({ clock_identity: "clock:s", epoch_ms: 1000,
  anchors: [anchor("clock:s", 1000, 1001, 5000), anchor("clock:s", 1100, 1101, 5100)] });
const activity = () => ({ ref: "activity:a", clock_identity: "clock:p", resolution_ms: 0.000001,
  availability: "available", outcome: "passed", duration_ms: 30, local_start_ms: 210, local_end_ms: 240,
  anchors: [anchor("clock:p", 200, 202, 5050), anchor("clock:p", 250, 252, 5100)] });

test("clock brackets enclose a boot read and preserve its quantization", () => {
  const calls = [], coordinates = [200, 202];
  const value = captureClockAnchor(() => { calls.push("local"); return { clock_identity: "clock:p", monotonic_ms: coordinates.shift() }; },
    () => { calls.push("boot"); return { clock_identity: boot, monotonic_ms: 5050 }; });
  assert.deepEqual(calls, ["local", "boot", "local"]);
  assert.deepEqual(value, anchor("clock:p", 200, 202, 5050));
  assert.equal(captureClockAnchor(() => { throw Error("unavailable"); }), null);
  const reversed = [202, 200];
  assert.equal(captureClockAnchor(() => ({ clock_identity: "clock:p", monotonic_ms: reversed.shift() }),
    () => ({ clock_identity: boot, monotonic_ms: 5050 })), null);
});

test("placement uses independently calculated interval arithmetic and preserves local duration", () => {
  // Scheduler offset [3999,4010], producer [4848,4860]; composed [838,861].
  for (const outcome of ["passed", "failed", "cancelled"]) {
    const local = { ...activity(), outcome }, original = structuredClone(local);
    assert.deepEqual(placeActivity(local, scheduler()), { activity_ref: "activity:a", availability: "available",
      start_bounds_ms: [48, 71], end_bounds_ms: [78, 101], uncertainty_ms: 23 });
    assert.deepEqual(local, original);
  }
});

test("placement rejects unsupported relationships without losing valid local measurements", () => {
  const cases = [
    [(a) => { a.anchors = []; }, "missing_anchor"],
    [(a) => { a.anchors[1].boot_identity = `sha256:${"2".repeat(64)}`; }, "incompatible_clock"],
    [(a) => { a.anchors.forEach((v) => { v.boot_identity = `sha256:${"2".repeat(64)}`; }); }, "incompatible_clock"],
    [(a) => { a.anchors[1].local_after_ms = 249; }, "discontinuity"],
    [(a) => { a.anchors[1].boot_ms += 10000; }, "discontinuity"],
    [(a) => { a.local_end_ms = 209; }, "discontinuity"],
    [(a) => { a.anchors.forEach((v) => { v.local_after_ms += 3; }); }, "uncertainty_exceeded"],
    [(a) => { a.outcome = "incomplete"; a.local_end_ms = null; }, "incomplete"],
  ];
  for (const [mutate, expected] of cases) {
    const local = activity(); mutate(local);
    const result = placeActivity(local, scheduler());
    assert.equal(result.availability, expected);
    assert.equal(result.start_bounds_ms, null);
    assert.equal(local.duration_ms, 30);
  }
  assert.equal(placeActivity(activity(), null).availability, "missing_anchor");
  const boundary = activity(); boundary.anchors.forEach((v) => { v.local_after_ms += 2; });
  assert.equal(placeActivity(boundary, scheduler()).uncertainty_ms, 25);
});
