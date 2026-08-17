import { createReadStream, existsSync, lstatSync } from "node:fs";
import { createInterface } from "node:readline";

import { validateSchemaSync } from "../contract/index.mjs";

export const defaultCanonicalEventLineLimit = 1024 * 1024;
export const terminalEventNames = new Set(["completed", "failed", "skipped", "cancelled"]);
const terminalStatuses = new Map([
  ["completed", "passed"],
  ["failed", "failed"],
  ["skipped", "skipped"],
  ["cancelled", "cancelled"],
]);

function abortError() {
  const error = new Error("canonical event iteration was cancelled");
  error.name = "AbortError";
  return error;
}

export async function* readCanonicalUnitEvents(file, options = {}) {
  if (!existsSync(file)) throw new Error(`${file} is required canonical event evidence`);
  const stat = lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error(`${file} must be a non-symlink regular canonical event file`);
  }
  const maxLineBytes = options.maxLineBytes ?? defaultCanonicalEventLineLimit;
  if (!Number.isSafeInteger(maxLineBytes) || maxLineBytes < 1024) {
    throw new Error("canonical event maxLineBytes must be a safe integer of at least 1024");
  }
  const input = createReadStream(file, { encoding: "utf8" });
  const lines = createInterface({ input, crlfDelay: Number.POSITIVE_INFINITY });
  const signal = options.signal;
  let previousSeq = 0;
  let previousMonotonicMs = 0;
  let lineNumber = 0;
  let eventCount = 0;
  const onAbort = () => input.destroy(abortError());
  if (signal?.aborted) onAbort();
  signal?.addEventListener("abort", onAbort, { once: true });
  try {
    for await (const line of lines) {
      lineNumber += 1;
      if (signal?.aborted) throw abortError();
      if (line === "") continue;
      const lineBytes = Buffer.byteLength(line, "utf8");
      if (lineBytes > maxLineBytes) {
        throw new Error(`${file} line ${lineNumber} exceeds ${maxLineBytes} bytes`);
      }
      let event;
      try {
        event = JSON.parse(line);
      } catch (error) {
        throw new Error(`${file} line ${lineNumber} is invalid JSON: ${error.message}`);
      }
      validateSchemaSync("cartulary.harness_unit_event.v2", event);
      if (
        terminalEventNames.has(event.event) &&
        event.status !== terminalStatuses.get(event.event)
      ) {
        throw new Error(`${file} terminal event ${event.event} has inconsistent status ${event.status}`);
      }
      if (event.seq !== previousSeq + 1) {
        throw new Error(`${file} sequence ${event.seq} is not contiguous at line ${lineNumber}`);
      }
      if (event.monotonic_ms < previousMonotonicMs) {
        throw new Error(`${file} monotonic time regresses at line ${lineNumber}`);
      }
      previousSeq = event.seq;
      previousMonotonicMs = event.monotonic_ms;
      eventCount += 1;
      yield event;
    }
    if (eventCount === 0) throw new Error(`${file} is empty`);
  } finally {
    signal?.removeEventListener("abort", onAbort);
    lines.close();
    input.destroy();
  }
}

export async function reduceCanonicalUnitIntervals(file, options = {}) {
  const selected = options.unitIDs === undefined ? null : new Set(options.unitIDs);
  const starts = new Map();
  const terminals = new Map();
  const eligible = new Set();
  const openWaits = new Map();
  const closedWaits = new Map();
  const admitted = new Set();
  const cacheHits = new Set();
  let runStarted = null;
  let runCompleted = null;
  let eventCount = 0;
  let finalMonotonicMs = 0;
  const projectedEvents = [];
  const projectedEventNames = new Set([
    "run_started",
    "wait_started",
    "wait_ended",
    "admitted",
    "started",
    "cache_hit",
    "cache_miss",
    "cache_bypass",
    "fixture_acquired",
    "fixture_released",
    "completed",
    "failed",
    "skipped",
    "cancelled",
    "cleanup_started",
    "cleanup_completed",
    "run_completed",
  ]);
  for await (const event of readCanonicalUnitEvents(file, options)) {
    eventCount += 1;
    finalMonotonicMs = event.monotonic_ms;
    if (options.retainProjection === true && projectedEventNames.has(event.event)) {
      projectedEvents.push(event);
    }
    if (event.event === "run_started") {
      if (runStarted !== null) throw new Error(`${file} has duplicate run_started events`);
      runStarted = event;
    }
    if (event.event === "run_completed") {
      if (runCompleted !== null) throw new Error(`${file} has duplicate run_completed events`);
      runCompleted = event;
    }
    if (selected !== null && !selected.has(event.unit_id)) continue;
    if (event.unit_id === "harness:run") continue;
    if (event.event === "eligible") {
      if (eligible.has(event.unit_id)) {
        throw new Error(`${file} has duplicate eligible event for ${event.unit_id}`);
      }
      eligible.add(event.unit_id);
    }
    if (event.event === "wait_started") {
      if (!eligible.has(event.unit_id) || openWaits.has(event.unit_id)) {
        throw new Error(`${file} has invalid wait start for ${event.unit_id}`);
      }
      openWaits.set(event.unit_id, {
        wait_reason: event.wait_reason,
        blocking_resources: event.blocking_resources,
        blocking_unit_ids: event.blocking_unit_ids,
      });
    }
    if (event.event === "wait_ended") {
      const opened = openWaits.get(event.unit_id);
      if (!opened) throw new Error(`${file} has unmatched wait end for ${event.unit_id}`);
      if (
        opened.wait_reason !== event.wait_reason ||
        JSON.stringify(opened.blocking_resources) !== JSON.stringify(event.blocking_resources) ||
        JSON.stringify(opened.blocking_unit_ids) !== JSON.stringify(event.blocking_unit_ids)
      ) {
        throw new Error(`${file} has mismatched wait end for ${event.unit_id}`);
      }
      openWaits.delete(event.unit_id);
      closedWaits.set(event.unit_id, event.monotonic_ms);
    }
    if (event.event === "admitted") {
      if (!eligible.has(event.unit_id) || openWaits.has(event.unit_id)) {
        throw new Error(`${file} admits ${event.unit_id} without a closed wait`);
      }
      if (closedWaits.get(event.unit_id) !== event.monotonic_ms) {
        throw new Error(`${file} admits ${event.unit_id} after its wait boundary`);
      }
      admitted.add(event.unit_id);
    }
    if (event.event === "cache_hit") cacheHits.add(event.unit_id);
    if (event.event === "started") {
      if (!admitted.has(event.unit_id)) {
        throw new Error(`${file} starts ${event.unit_id} before admission`);
      }
      if (starts.has(event.unit_id)) throw new Error(`${file} has duplicate start event for ${event.unit_id}`);
      starts.set(event.unit_id, event.monotonic_ms);
    }
    if (terminalEventNames.has(event.event)) {
      if (!eligible.has(event.unit_id) || openWaits.has(event.unit_id)) {
        throw new Error(`${file} terminates ${event.unit_id} without a closed wait`);
      }
      if (
        !new Set(["skipped", "cancelled"]).has(event.event) &&
        !starts.has(event.unit_id) &&
        !cacheHits.has(event.unit_id)
      ) {
        throw new Error(`${file} terminal event for ${event.unit_id} occurs before start`);
      }
      if (
        !starts.has(event.unit_id) &&
        closedWaits.get(event.unit_id) !== event.monotonic_ms
      ) {
        throw new Error(`${file} terminates ${event.unit_id} after its wait boundary`);
      }
      if (terminals.has(event.unit_id)) {
        throw new Error(`${file} has duplicate terminal event for ${event.unit_id}`);
      }
      terminals.set(event.unit_id, event);
    }
  }
  if (openWaits.size > 0) {
    throw new Error(`${file} has unmatched wait start for ${[...openWaits.keys()].sort().join(", ")}`);
  }
  return {
    eventCount,
    finalMonotonicMs,
    runStarted,
    runCompleted,
    starts,
    terminals,
    projectedEvents,
  };
}
