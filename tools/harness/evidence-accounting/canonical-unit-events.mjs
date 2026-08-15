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
      validateSchemaSync("cartulary.harness_unit_event.v1", event);
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
  let runStarted = null;
  let runCompleted = null;
  let eventCount = 0;
  let finalMonotonicMs = 0;
  for await (const event of readCanonicalUnitEvents(file, options)) {
    eventCount += 1;
    finalMonotonicMs = event.monotonic_ms;
    if (event.event === "run_started") {
      if (runStarted !== null) throw new Error(`${file} has duplicate run_started events`);
      runStarted = event;
    }
    if (event.event === "run_completed") {
      if (runCompleted !== null) throw new Error(`${file} has duplicate run_completed events`);
      runCompleted = event;
    }
    if (selected !== null && !selected.has(event.unit_id)) continue;
    if (event.event === "started") {
      if (starts.has(event.unit_id)) throw new Error(`${file} has duplicate start event for ${event.unit_id}`);
      starts.set(event.unit_id, event.monotonic_ms);
    }
    if (terminalEventNames.has(event.event)) {
      if (event.event !== "skipped" && !starts.has(event.unit_id)) {
        throw new Error(`${file} terminal event for ${event.unit_id} occurs before start`);
      }
      if (terminals.has(event.unit_id)) {
        throw new Error(`${file} has duplicate terminal event for ${event.unit_id}`);
      }
      terminals.set(event.unit_id, event);
    }
  }
  return {
    eventCount,
    finalMonotonicMs,
    runStarted,
    runCompleted,
    starts,
    terminals,
  };
}
