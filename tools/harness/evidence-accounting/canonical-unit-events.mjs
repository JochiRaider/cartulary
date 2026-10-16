import { parseStrictJSON } from "../contract/strict-json.mjs";
import { closeSync, constants, createReadStream, fstatSync, openSync, readSync } from "node:fs";
import { createInterface } from "node:readline";
import path from "node:path";

import { assertMachineEvidencePath, maximumPhasesPerUnit, openDirectory, validateSchemaSync } from "../contract/index.mjs";

const defaultCanonicalEventLineLimit = 1024 * 1024;
const terminalEventNames = new Set(["completed", "failed", "skipped", "cancelled"]);
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
  const maxLineBytes = options.maxLineBytes ?? defaultCanonicalEventLineLimit;
  if (!Number.isSafeInteger(maxLineBytes) || maxLineBytes < 1024) {
    throw new Error("canonical event maxLineBytes must be a safe integer of at least 1024");
  }
  const prefix = options.prefix;
  if (prefix !== undefined) validateSchemaSync("cartulary.harness_unit_event_prefix.v1", prefix);
  assertMachineEvidencePath(file);
  let descriptor;
  let parent;
  try {
    parent = openDirectory(path.dirname(file));
    descriptor = openSync(`/proc/self/fd/${parent}/${path.basename(file)}`, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK);
  } catch (error) {
    if (error.code === "ENOENT") throw new Error(`${file} is required canonical event evidence`);
    throw error;
  } finally {
    if (parent !== undefined) closeSync(parent);
  }
  let input;
  let stat;
  try {
    stat = fstatSync(descriptor, { bigint: true });
    if (!stat.isFile() || stat.nlink !== 1n) throw new Error(`${file} must be an unaliased non-symlink regular canonical event file`);
    if (stat.size === 0n) throw new Error(`${file} is empty`);
    if (stat.size > BigInt(Number.MAX_SAFE_INTEGER)) throw new Error("canonical event file exceeds addressable size");
    if (prefix !== undefined) {
      if (prefix.file !== file || stat.dev.toString() !== prefix.device || stat.ino.toString() !== prefix.inode ||
          stat.size < BigInt(prefix.byte_length) || stat.uid !== BigInt(process.getuid()) ||
          stat.nlink !== 1n || (stat.mode & 0o777n) !== 0o600n) {
        throw new Error("canonical event prefix inode authority does not match");
      }
      const boundary = Buffer.alloc(1);
      if (readSync(descriptor, boundary, 0, 1, prefix.byte_length - 1) !== 1 || boundary[0] !== 10) {
        throw new Error("canonical event prefix must end at a complete record boundary");
      }
    }
    input = createReadStream(file, { fd: descriptor, autoClose: false, encoding: "utf8", start: 0, end: prefix === undefined ? Number(stat.size) - 1 : prefix.byte_length - 1 });
  } catch (error) {
    closeSync(descriptor);
    throw error;
  }
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
        event = parseStrictJSON(line, "canonical event");
      } catch (error) {
        throw new Error(`${file} line ${lineNumber} is invalid JSON: ${error.message}`);
      }
      validateSchemaSync("cartulary.harness_unit_event.v4", event);
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
    if (prefix !== undefined && previousSeq !== prefix.last_seq) {
      throw new Error("canonical event prefix final sequence does not match");
    }
    if (prefix === undefined) {
      const after = fstatSync(descriptor, { bigint: true });
      if (after.size !== stat.size || after.mtimeNs !== stat.mtimeNs || after.ctimeNs !== stat.ctimeNs || after.nlink !== 1n) {
        throw new Error("canonical event evidence changed during read");
      }
    }
  } finally {
    signal?.removeEventListener("abort", onAbort);
    lines.close();
    await new Promise((resolve, reject) => input.close((error) => error ? reject(error) : resolve()));
  }
}

export async function reduceCanonicalUnitIntervals(file, options = {}) {
  const selected = options.unitIDs === undefined ? null : new Set(options.unitIDs);
  const starts = new Map();
  const terminals = new Map();
  const eligible = new Map();
  const openWaits = new Map();
  const closedWaits = new Map();
  const admitted = new Set();
  const cacheOutcomes = new Map();
  const registrations = new Map();
  const phases = new Map();
  const phaseIDs = new Set();
  const phaseCounts = new Map();
  let runStarted = null;
  let runCompleted = null;
  let eventCount = 0;
  let finalMonotonicMs = 0;
  const waits = new Map();
  const admissions = new Map();
  const fixtureAcquired = new Map();
  const cache = { hit: 0, miss: 0, bypass: 0 };
  let cleanupStarted = null;
  for await (const event of readCanonicalUnitEvents(file, options)) {
    eventCount += 1;
    finalMonotonicMs = event.monotonic_ms;
    if (event.event === "cleanup_started") cleanupStarted = event.monotonic_ms;
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
    if (event.event === "queued") {
      if (registrations.has(event.unit_id)) throw new Error("duplicate unit registration");
      registrations.set(event.unit_id, { unit_id: event.unit_id, kind: event.unit_kind,
        row_ids: event.row_ids, needs: event.needs, resource_claims: event.resource_claims,
        service_dependencies: event.service_dependencies });
    }
    if (event.event === "phase_started") {
      const count = (phaseCounts.get(event.unit_id) ?? 0) + 1;
      if (count > maximumPhasesPerUnit || phaseIDs.has(event.interval_id) || !registrations.has(event.unit_id) || terminals.has(event.unit_id)) throw new Error("invalid phase start");
      phaseCounts.set(event.unit_id, count);
      phaseIDs.add(event.interval_id);
      phases.set(event.interval_id, { unit_id: event.unit_id, phase: event.phase, start_ms: event.monotonic_ms, end_ms: null, outcome: null });
    }
    if (event.event === "phase_finished") {
      const phase = phases.get(event.interval_id);
      if (!phase || phase.end_ms !== null || phase.unit_id !== event.unit_id || phase.phase !== event.phase || terminals.has(event.unit_id)) throw new Error("unmatched phase finish");
      phase.end_ms = event.monotonic_ms;
      phase.outcome = event.phase_outcome;
    }
    if (event.event === "eligible") {
      if (eligible.has(event.unit_id)) {
        throw new Error(`${file} has duplicate eligible event for ${event.unit_id}`);
      }
      eligible.set(event.unit_id, event.monotonic_ms);
    }
    if (event.event === "wait_started") {
      const prior = waits.get(event.unit_id);
      if (!eligible.has(event.unit_id) || openWaits.has(event.unit_id) || admitted.has(event.unit_id) || terminals.has(event.unit_id) ||
        event.monotonic_ms !== (closedWaits.get(event.unit_id) ?? eligible.get(event.unit_id))) {
        throw new Error(`${file} has invalid wait start for ${event.unit_id}`);
      }
      const wait = {
        start: event.monotonic_ms, end: null,
        wait_reason: event.wait_reason,
        blocking_resources: event.blocking_resources,
        blocking_unit_ids: event.blocking_unit_ids,
      };
      const blocked = ["resources", "exclusive_barrier"].includes(wait.wait_reason);
      if (blocked !== (wait.blocking_resources.length > 0) || blocked !== (wait.blocking_unit_ids.length > 0) ||
        [wait.blocking_resources, wait.blocking_unit_ids].some((ids) => JSON.stringify(ids) !== JSON.stringify([...ids].sort())) ||
        wait.blocking_unit_ids.includes(event.unit_id) ||
        (selected === null && wait.blocking_unit_ids.some((id) => !registrations.has(id)))) throw new Error("invalid wait authority");
      if (prior && prior.last.wait_reason === wait.wait_reason &&
        JSON.stringify(prior.last.blocking_resources) === JSON.stringify(wait.blocking_resources) &&
        JSON.stringify(prior.last.blocking_unit_ids) === JSON.stringify(wait.blocking_unit_ids)) throw new Error("unchanged wait segment");
      const summary = prior ?? { start: event.monotonic_ms, end: null, count: 0,
        by_reason: { resources: 0, exclusive_barrier: 0, rank: 0, scheduler_stop: 0 },
        resource_intervals: [], resource_holders: new Set() };
      summary.count += 1;
      summary.last = wait;
      waits.set(event.unit_id, summary);
      openWaits.set(event.unit_id, wait);
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
      opened.end = event.monotonic_ms;
      const summary = waits.get(event.unit_id);
      summary.end = opened.end;
      summary.by_reason[opened.wait_reason] += opened.end - opened.start;
      if (opened.wait_reason === "resources") {
        for (const holder of opened.blocking_unit_ids) summary.resource_holders.add(holder);
        if (opened.end > opened.start) {
          const prior = summary.resource_intervals.at(-1);
          if (prior?.end === opened.start) prior.end = opened.end;
          else summary.resource_intervals.push({ start: opened.start, end: opened.end });
        }
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
      if (admitted.has(event.unit_id)) throw new Error("duplicate unit admission");
      admitted.add(event.unit_id);
      admissions.set(event.unit_id, { time: event.monotonic_ms, seq: event.seq });
    }
    if (["cache_hit", "cache_miss", "cache_bypass"].includes(event.event)) {
      if (cacheOutcomes.has(event.unit_id) || starts.has(event.unit_id) || terminals.has(event.unit_id)) {
        throw new Error(`${file} has duplicate or late cache outcome for ${event.unit_id}`);
      }
      const outcome = event.event.slice("cache_".length);
      cacheOutcomes.set(event.unit_id, outcome);
      cache[outcome] += 1;
    }
    if (event.event === "fixture_acquired") fixtureAcquired.set(event.unit_id, event.monotonic_ms);
    if (event.event === "started") {
      if (cacheOutcomes.get(event.unit_id) === "hit") {
        throw new Error(`${file} starts cached unit ${event.unit_id}`);
      }
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
        cacheOutcomes.get(event.unit_id) !== "hit"
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
      terminals.set(event.unit_id, {
        ...Object.fromEntries([
          "unit_id", "event", "status", "seq", "monotonic_ms", "failure_class", "failure_reason",
        ].filter((key) => Object.hasOwn(event, key)).map((key) => [key, event[key]])),
        cache_status: cacheOutcomes.get(event.unit_id) ?? null,
      });
    }
  }
  if (openWaits.size > 0) {
    throw new Error(`${file} has unmatched wait start for ${[...openWaits.keys()].sort().join(", ")}`);
  }
  const intervals = new Map();
  for (const [id, terminal] of terminals) {
    const start = starts.get(id) ?? terminal.monotonic_ms;
    intervals.set(id, { start, end: terminal.monotonic_ms, queue_ms: start - waits.get(id).start });
  }
  return {
    intervals, waits, admissions, fixtureAcquired, cache, cleanupStarted,
    eventCount,
    finalMonotonicMs,
    runStarted,
    runCompleted,
    starts,
    terminals,
    registrations,
    phases,
  };
}
