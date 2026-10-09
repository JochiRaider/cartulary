import { closeSync, readdirSync } from "node:fs";
import path from "node:path";
import { setTimeout as delay } from "node:timers/promises";
import { parseStrictJSON, publicExitCodeForFailure, validateSchemaSync } from "../contract/index.mjs";
import {
  compareObservationIDs, liveSnapshotBytes, manifestObservationIdentity,
  observationBlockerLimit, observationBytes, observationListLimit,
  runObservationSchema, validateLiveSnapshot,
} from "../contract/live-observation.mjs";
import { openDirectory, readLocalFile } from "../runtime/secure-local-files.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../../..");
const token = /^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$/u;
const targetToken = /^[A-Za-z0-9_][A-Za-z0-9_.-]{0,254}$/u;
const baseLimitations = ["current_source_not_audited", "outer_exit_not_observed", "process_not_observed"];

function failure(failureClass, failureReason) {
  return Object.assign(new Error(failureReason), { failure_class: failureClass, failure_reason: failureReason });
}
const usageError = () => failure("config", "usage_error");
const artifactError = () => failure("artifact", "artifact_error");

export function emptyObservation() {
  return {
    schema_id: runObservationSchema, operation_status: "ok", operation_exit_code: 0,
    run: null, selected_target: null, observation: "not_published", revision: null,
    wait_result: "immediate", waited_ms: 0, live: null, terminal: null,
    unavailable_reason: null, limitations: [...baseLimitations],
    failure_class: null, failure_reason: null,
  };
}

export function observationFailure(error, observation = emptyObservation()) {
  const interrupted = error?.name === "AbortError" || error?.failure_class === "interrupted";
  const failureClass = interrupted ? "interrupted" : error?.failure_class ?? (error?.code === "MODULE_NOT_FOUND" ? "config" : "artifact");
  const failureReason = interrupted ? "cancelled_or_interrupted" : error?.failure_reason ?? (failureClass === "config" ? "configuration_error" : "artifact_error");
  return {
    ...observation, operation_status: "error",
    operation_exit_code: interrupted && error?.signal === "SIGTERM" ? 143 : interrupted && error?.signal === "SIGINT" ? 130 : publicExitCodeForFailure({ failure_class: failureClass, failure_reason: failureReason }),
    observation: "invalid", live: null, terminal: null,
    failure_class: failureClass, failure_reason: failureReason,
  };
}

function integer(value, fallback, maximum) {
  if (value === undefined || value === "") return fallback;
  if (!/^\d+$/u.test(String(value))) throw usageError();
  const number = Number(value);
  if (!Number.isSafeInteger(number) || number > maximum) throw usageError();
  return number;
}

export function normalizeObservationInputs({ resultsDir, runId = "", target = "", afterRevision, waitSeconds } = {}) {
  if (typeof resultsDir !== "string" || !resultsDir || resultsDir.length > 4096 || /[\x00-\x1f\\]/u.test(resultsDir) || resultsDir.split("/").includes("..")) throw usageError();
  if (runId && (!token.test(runId) || runId === "." || runId === "..")) throw usageError();
  if (target && !targetToken.test(target)) throw usageError();
  const runRoot = path.resolve(repoRoot, resultsDir, runId);
  const expectedRunID = runId || path.basename(runRoot);
  if (!token.test(expectedRunID) || runRoot.length > 4096) throw usageError();
  const after = integer(afterRevision, null, Number.MAX_SAFE_INTEGER);
  const wait = integer(waitSeconds, 0, 30);
  if (wait > 0 && after === null) throw usageError();
  return { runRoot, runId: expectedRunID, target, afterRevision: after, waitSeconds: wait };
}

function readJSON(file, schema, { optional = false, mutable = false } = {}) {
  let bytes;
  try { bytes = readLocalFile(file, { maximum: liveSnapshotBytes, allowReplacement: mutable }); }
  catch (error) { if (optional && error.code === "ENOENT") return null; throw artifactError(); }
  const value = parseStrictJSON(bytes.toString("utf8"));
  if (schema) validateSchemaSync(schema, value);
  return value;
}

function existingRun(runRoot) {
  let descriptor;
  try { descriptor = openDirectory(runRoot, { privateLeaf: true }); return true; }
  catch (error) { if (error.code === "ENOENT") return false; throw artifactError(); }
  finally { if (descriptor !== undefined) closeSync(descriptor); }
}

function toolTargets(runRoot) {
  const descriptor = openDirectory(runRoot, { privateLeaf: true });
  try {
    const entries = readdirSync(`/proc/self/fd/${descriptor}`, { withFileTypes: true });
    if (entries.length > 4096) throw artifactError();
    return entries.filter((entry) => entry.isDirectory() && targetToken.test(entry.name))
      .map((entry) => entry.name).sort(compareObservationIDs);
  } finally { closeSync(descriptor); }
}

function terminalObservation(result, summary, summaryRef) {
  result.observation = "terminal_summary";
  result.wait_result = "terminal";
  result.terminal = {
    reported_status: summary.status, summary_ref: summaryRef,
    failure_class: summary.failure_class, failure_reason: summary.failure_reason,
    normalized_exit_code: summary.exit_code ?? null, evidence_audit: "not_run",
  };
  return result;
}

function liveProjection(snapshot, target) {
  let units = snapshot.units;
  if (target) {
    const projection = snapshot.targets.find((entry) => entry.target === target);
    if (!projection) throw usageError();
    const selected = new Set(projection.unit_ids);
    units = units.filter((unit) => selected.has(unit.unit_id));
  }
  const unitCounts = { total: units.length, pending: 0, running: 0, passed: 0, failed: 0, skipped: 0, cancelled: 0 };
  const cacheCounts = { not_observed: 0, hit: 0, miss: 0, bypass: 0 };
  const running = [], waiting = [], failures = [];
  let runningCount = 0, waitingCount = 0, failureCount = 0;
  for (const unit of units) {
    unitCounts[unit.status]++;
    cacheCounts[unit.cache_disposition]++;
    if (unit.status === "running") {
      runningCount++;
      if (running.length < observationListLimit) running.push({ unit_id: unit.unit_id, activity: unit.activity, elapsed_ms: unit.started_elapsed_ms === null ? null : snapshot.sampled_elapsed_ms - unit.started_elapsed_ms });
    }
    if (unit.wait_reason !== null) {
      waitingCount++;
      if (waiting.length < observationListLimit) waiting.push({
        unit_id: unit.unit_id, activity: unit.activity, wait_reason: unit.wait_reason,
        blocking_resources: unit.blocking_resources.slice(0, observationBlockerLimit),
        blocking_unit_ids: unit.blocking_unit_ids.slice(0, observationBlockerLimit),
        blocking_resources_omitted: Math.max(0, unit.blocking_resources.length - observationBlockerLimit),
        blocking_unit_ids_omitted: Math.max(0, unit.blocking_unit_ids.length - observationBlockerLimit),
      });
    }
    if (unit.failure_class !== null) {
      failureCount++;
      if (failures.length < observationListLimit) failures.push({ unit_id: unit.unit_id, failure_class: unit.failure_class, failure_reason: unit.failure_reason });
    }
  }
  return {
    sampled_at: snapshot.sampled_at, sampled_elapsed_ms: snapshot.sampled_elapsed_ms,
    last_transition_seq: snapshot.last_transition_seq, publication_mode: snapshot.publication_mode,
    phase: snapshot.phase, unit_counts: unitCounts, cache_counts: cacheCounts,
    running, waiting, failures_observed: failures,
    running_omitted: runningCount - running.length, waiting_omitted: waitingCount - waiting.length,
    failures_omitted: failureCount - failures.length,
  };
}

function addLimitation(result, code) {
  result.limitations = [...new Set([...result.limitations, code])].sort(compareObservationIDs);
}

export function boundObservation(result) {
  const live = result.live;
  if (live) {
    const lists = [["failures_observed", "failures_omitted"], ["waiting", "waiting_omitted"], ["running", "running_omitted"]];
    if (lists.some(([, omitted]) => live[omitted] > 0) || live.waiting.some((entry) => entry.blocking_resources_omitted || entry.blocking_unit_ids_omitted)) addLimitation(result, "lists_truncated");
    while (Buffer.byteLength(JSON.stringify(result)) + 1 > observationBytes) {
      const list = lists.find(([name]) => live[name].length);
      if (!list) throw artifactError();
      live[list[0]].pop(); live[list[1]]++;
      addLimitation(result, "lists_truncated");
    }
  }
  validateSchemaSync(runObservationSchema, result);
  if (Buffer.byteLength(JSON.stringify(result)) + 1 > observationBytes) throw artifactError();
  return result;
}

function sample(options) {
  const { runRoot, runId, target } = options;
  const result = emptyObservation();
  result.run = { result_root: runRoot, run_id: runId, public_target: null, command_id: null, source_digest: null, graph_digest: null };
  result.selected_target = target || null;
  if (!existingRun(runRoot)) { addLimitation(result, "live_unavailable"); return result; }
  const manifest = readJSON(path.join(runRoot, "run-manifest.json"), "cartulary.harness_run_manifest.v1", { optional: true });
  if (manifest) {
    if (manifest.run_id !== runId) throw artifactError();
    Object.assign(result.run, manifestObservationIdentity(manifest));
  }
  const summary = readJSON(path.join(runRoot, "run-summary.json"), "cartulary.harness_run_summary.v1", { optional: true });
  if (summary) {
    if (!manifest || summary.run_id !== runId || summary.target !== manifest.target) throw artifactError();
    if (!target) return terminalObservation(result, summary, "run-summary.json");
    const ref = `target-summaries/${target}.json`;
    if (!summary.artifact_refs.includes(ref)) throw usageError();
    const projection = readJSON(path.join(runRoot, ref), "cartulary.harness_target_summary.v1");
    if (projection.target !== target) throw artifactError();
    return terminalObservation(result, projection, ref);
  }
  if (!manifest) {
    const candidates = target ? [target] : toolTargets(runRoot);
    const summaries = [];
    for (const candidate of candidates) {
      const ref = `${candidate}/tool-run-summary.json`;
      const tool = readJSON(path.join(runRoot, ref), "cartulary.tool_run_summary.v5", { optional: true });
      if (!tool) continue;
      if (tool.run_id !== runId || tool.target !== candidate || path.resolve(repoRoot, tool.run_root) !== runRoot || path.resolve(repoRoot, tool.result_root) !== path.dirname(runRoot)) throw artifactError();
      summaries.push({ tool, ref });
    }
    if (summaries.length > 1) throw usageError();
    if (summaries.length) {
      result.run.public_target = summaries[0].tool.target;
      return terminalObservation(result, summaries[0].tool, summaries[0].ref);
    }
    addLimitation(result, "live_unavailable");
    return result;
  }
  const snapshot = readJSON(path.join(runRoot, "diagnostics/live-status.json"), null, { optional: true, mutable: true });
  if (!snapshot) { addLimitation(result, "live_unavailable"); return result; }
  validateLiveSnapshot(snapshot, manifest);
  if (options.afterRevision !== null && options.afterRevision > snapshot.revision) throw usageError();
  result.revision = snapshot.revision;
  if (snapshot.publication_mode === "paused_measurement") {
    result.wait_result = "unavailable";
    addLimitation(result, "measurement_pause");
  }
  if (snapshot.availability === "unavailable") {
    result.observation = "unavailable";
    result.unavailable_reason = snapshot.unavailable_reason;
    result.wait_result = "unavailable";
    addLimitation(result, "live_unavailable");
  } else {
    result.observation = "live";
    result.live = liveProjection(snapshot, target);
  }
  return result;
}

export async function observeRun(inputs, { signal, now = () => performance.now(), sleep = (ms) => delay(ms, undefined, { signal }) } = {}) {
  const started = now();
  let result = emptyObservation();
  try {
    const options = normalizeObservationInputs(inputs);
    while (true) {
      signal?.throwIfAborted();
      result = sample(options);
      result.waited_ms = Math.max(0, Math.floor(now() - started));
      if (["terminal", "unavailable"].includes(result.wait_result)) break;
      if (options.afterRevision !== null && result.revision !== null && result.revision > options.afterRevision) { result.wait_result = "changed"; break; }
      if (options.waitSeconds === 0) break;
      const remaining = options.waitSeconds * 1000 - (now() - started);
      if (remaining <= 0) { result.wait_result = "wait_expired"; break; }
      await sleep(Math.min(1000, remaining));
      if (now() - started >= options.waitSeconds * 1000) {
        result.wait_result = "wait_expired";
        result.waited_ms = Math.max(0, Math.floor(now() - started));
        break;
      }
    }
    return boundObservation(result);
  } catch (error) {
    // timers/promises wraps the caller's signal reason in an AbortError.
    // Preserve the native signal separately from generic cancellation.
    return boundObservation(observationFailure(signal?.aborted ? signal.reason : error, { ...result, waited_ms: Math.max(0, Math.floor(now() - started)) }));
  }
}

export function renderObservation(result) {
  const lines = [
    `Observation: ${result.observation}; reader=${result.operation_status} code=${result.operation_exit_code}`,
    `Run: ${result.run?.result_root ?? "unavailable"}`,
    `Target: ${result.selected_target ?? result.run?.public_target ?? "unavailable"}; wait=${result.wait_result}`,
  ];
  if (result.live) {
    const live = result.live;
    lines.push(`Sample: ${live.sampled_at}; revision=${result.revision}; progress=${live.last_transition_seq}; ${live.publication_mode}; ${live.phase}`);
    lines.push(`Units: ${JSON.stringify(live.unit_counts)}`);
    for (const unit of live.running.slice(0, 4)) lines.push(`Running: ${unit.unit_id} (${unit.activity})`);
    for (const unit of live.waiting.slice(0, 4)) lines.push(`Waiting: ${unit.unit_id} (${unit.wait_reason})`);
    for (const unit of live.failures_observed.slice(0, 4)) lines.push(`Failure: ${unit.unit_id} ${unit.failure_class}/${unit.failure_reason}`);
    lines.push(`Omitted: running=${live.running_omitted + Math.max(0, live.running.length - 4)} waiting=${live.waiting_omitted + Math.max(0, live.waiting.length - 4)} failures=${live.failures_omitted + Math.max(0, live.failures_observed.length - 4)}`);
  }
  if (result.terminal) lines.push(`Reported result: ${result.terminal.reported_status}; ${result.terminal.summary_ref}; evidence audit not run`);
  if (result.failure_class) lines.push(`Reader failure: ${result.failure_class}/${result.failure_reason}`);
  lines.push(`Limitations: ${result.limitations.join(", ")}`);
  // Scalar paths and IDs have schema bounds; bound the human renderer as well.
  const output = [];
  let bytes = 0;
  for (const line of lines) {
    if (output.length >= 29 || bytes + Buffer.byteLength(line) + 1 > 8100) { output.push("Display truncated; use JSON=1 for bounded structured detail."); break; }
    output.push(line); bytes += Buffer.byteLength(line) + 1;
  }
  return `${output.join("\n")}\n`;
}
