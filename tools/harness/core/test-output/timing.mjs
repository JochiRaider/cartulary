#!/usr/bin/env node

import {
  existsSync,
  readdirSync,
  readFileSync,
} from "node:fs";
import path from "node:path";
import {
  prettyJSONString,
  secureMkdir,
  secureWriteFile,
} from "../harness-contract.mjs";
import {
  repoRoot,
  resolveResultsRoot,
  resolveRunId,
  targetTimingSchemaID,
  timingBucketOrder,
  timingBucketSet,
} from "./context.mjs";

const resultsRoot = resolveResultsRoot();

const runId = resolveRunId();

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function relToRepo(value) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(repoRoot, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

function ensureDir(dir) {
  secureMkdir(dir);
}

function writeJson(file, value) {
  ensureDir(path.dirname(file));
  secureWriteFile(file, prettyJSONString(value));
}

function clampDurationMs(value) {
  if (!Number.isFinite(value) || value < 0) {
    return 0;
  }
  return value;
}

function normalizeAccountingMode(value) {
  if (value === "actual" || value === "reused" || value === "derived") {
    return value;
  }
  return "actual";
}

function normalizeTimingBucket(value, runner = "") {
  if (value && timingBucketSet.has(value)) {
    return value;
  }
  if (runner === "go_test" || runner === "vitest" || runner === "playwright") {
    return "test_command";
  }
  return "test_command";
}

function computeWindowDurationMs(startTime, endTime) {
  if (!startTime || !endTime) {
    return 0;
  }
  const startMs = Date.parse(startTime);
  const endMs = Date.parse(endTime);
  if (!Number.isFinite(startMs) || !Number.isFinite(endMs) || endMs < startMs) {
    return 0;
  }
  return endMs - startMs;
}

function requiredEnv(name) {
  const value = process.env[name];
  if (value === undefined || value === "") {
    throw new Error(`missing required environment variable ${name}`);
  }
  return value;
}

function optionalEnv(name, fallback = "") {
  return process.env[name] ?? fallback;
}

function parseInteger(name, fallback = 0) {
  const value = process.env[name];
  if (value === undefined || value === "") {
    return fallback;
  }
  const parsed = Number.parseInt(value, 10);
  if (Number.isNaN(parsed)) {
    throw new Error(`invalid integer ${name}=${value}`);
  }
  return parsed;
}

function slugifyLabel(label) {
  return label
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .replace(/--+/g, "-");
}

function timingSpanPath(target) {
  const targetDir = path.join(resultsRoot, runId, target);
  const label = optionalEnv("CARTULARY_TIMING_LABEL", "timing-span");
  const slug = slugifyLabel(label) || "timing-span";
  const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
  return path.join(
    targetDir,
    "timing-spans",
    `${timestamp}-${process.pid}-${slug}.json`,
  );
}

function createTargetOwnedTimingSpan(source = "target") {
  const target = optionalEnv("CARTULARY_TEST_TARGET", "");
  if (target === "") {
    return null;
  }
  const bucket = normalizeTimingBucket(optionalEnv("CARTULARY_TIMING_BUCKET"));
  const label = requiredEnv("CARTULARY_TIMING_LABEL");
  const durationMs = clampDurationMs(
    parseInteger("CARTULARY_TIMING_DURATION_MS", 0),
  );
  const startTime = optionalEnv("CARTULARY_TIMING_START_TIME");
  const endTime = optionalEnv("CARTULARY_TIMING_END_TIME");
  return {
    source,
    bucket,
    label,
    start_time: startTime,
    end_time: endTime,
    duration_ms: durationMs,
    status: optionalEnv("CARTULARY_TIMING_STATUS", "pass"),
  };
}

export function handleTimingSpan() {
  const span = createTargetOwnedTimingSpan();
  if (!span) {
    return 0;
  }
  writeJson(timingSpanPath(optionalEnv("CARTULARY_TEST_TARGET")), span);
  return 0;
}

function loadTargetOwnedTimingSpans(targetDir) {
  const spansDir = path.join(targetDir, "timing-spans");
  if (!existsSync(spansDir)) {
    return [];
  }
  const spans = [];
  for (const entry of readdirSync(spansDir, { withFileTypes: true })) {
    if (!entry.isFile() || !entry.name.endsWith(".json")) {
      continue;
    }
    const span = JSON.parse(
      readFileSync(path.join(spansDir, entry.name), "utf8"),
    );
    if (!span?.bucket || !timingBucketSet.has(span.bucket)) {
      continue;
    }
    spans.push({
      source: span.source ?? "target",
      bucket: span.bucket,
      label: span.label ?? "",
      start_time: span.start_time ?? "",
      end_time: span.end_time ?? "",
      duration_ms: clampDurationMs(span.duration_ms ?? 0),
      status: span.status ?? "",
    });
  }
  return spans;
}

function loadServiceTimingSpans(target) {
  const servicesRoot = path.join(
    resultsRoot,
    runId,
    "_shared",
    "test-services",
  );
  if (!existsSync(servicesRoot)) {
    return [];
  }
  const spans = [];
  const stack = [servicesRoot];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile() || !entry.name.endsWith(".json")) {
        continue;
      }
      let event;
      try {
        event = JSON.parse(readFileSync(next, "utf8"));
      } catch {
        continue;
      }
      if (event.type !== "timing-span") {
        continue;
      }
      const details = event.details ?? {};
      if (details.target !== target) {
        continue;
      }
      const bucket = normalizeTimingBucket(details.bucket);
      spans.push({
        source: "test_services",
        bucket,
        label: details.label ?? event.name ?? "test-services timing",
        start_time: details.start_time ?? "",
        end_time: details.end_time ?? event.timestamp ?? "",
        duration_ms: clampDurationMs(details.duration_ms ?? 0),
        status: details.status ?? event.status ?? "",
        janitorial: details.janitorial === true,
        startup_attempt: details.startup_attempt === true,
        service: details.service ?? event.service ?? "",
        attempt: details.attempt ?? 0,
        max_attempts: details.max_attempts ?? 0,
        retry_scheduled: details.retry_scheduled === true,
        retry_blocked_by_context: details.retry_blocked_by_context === true,
        artifact: relToRepo(path.dirname(path.dirname(next))),
      });
    }
  }
  return spans;
}

function phaseSummaryTimingSpan(summary) {
  return {
    source: "phase",
    bucket: normalizeTimingBucket(summary.timing_bucket, summary.runner),
    label: summary.label,
    runner: summary.runner,
    status: summary.status,
    accounting_mode: normalizeAccountingMode(summary.accounting_mode),
    start_time: summary.start_time ?? "",
    end_time: summary.end_time ?? "",
    duration_ms: clampDurationMs(
      summary.wall_duration_ms ??
        summary.logical_duration_ms ??
        summary.duration_ms ??
        0,
    ),
    logical_duration_ms: clampDurationMs(
      summary.logical_duration_ms ?? summary.duration_ms ?? 0,
    ),
    executed_duration_ms: clampDurationMs(summary.executed_duration_ms ?? 0),
    artifacts: summary.artifacts ?? {},
  };
}

function addTimingSpanToBuckets(buckets, span) {
  if (!span || !timingBucketSet.has(span.bucket)) {
    return;
  }
  if (!buckets.has(span.bucket)) {
    buckets.set(span.bucket, {
      name: span.bucket,
      duration_ms: 0,
      spans: [],
    });
  }
  const bucket = buckets.get(span.bucket);
  const durationMs = clampDurationMs(span.duration_ms ?? 0);
  bucket.spans.push({
    ...span,
    duration_ms: durationMs,
  });
}

export function disjointSpanDurationMs(spans) {
  if (spans.length === 1) {
    return clampDurationMs(spans[0]?.duration_ms ?? 0);
  }

  const intervals = [];
  let fallbackDurationMs = 0;
  for (const span of spans) {
    const durationMs = clampDurationMs(span.duration_ms ?? 0);
    const startMs = Date.parse(span.start_time ?? "");
    const endMs = Date.parse(span.end_time ?? "");
    if (
      Number.isFinite(startMs) &&
      Number.isFinite(endMs) &&
      endMs >= startMs
    ) {
      intervals.push([startMs, endMs]);
      continue;
    }
    fallbackDurationMs += durationMs;
  }
  intervals.sort((left, right) => left[0] - right[0] || left[1] - right[1]);
  let total = fallbackDurationMs;
  let currentStart;
  let currentEnd;
  for (const [startMs, endMs] of intervals) {
    if (currentStart === undefined) {
      currentStart = startMs;
      currentEnd = endMs;
      continue;
    }
    if (startMs <= currentEnd) {
      currentEnd = Math.max(currentEnd, endMs);
      continue;
    }
    total += currentEnd - currentStart;
    currentStart = startMs;
    currentEnd = endMs;
  }
  if (currentStart !== undefined) {
    total += currentEnd - currentStart;
  }
  return clampDurationMs(total);
}

export function lifecycleTimingSpans(target, targetDir) {
  return [
    ...loadTargetOwnedTimingSpans(targetDir),
    ...loadServiceTimingSpans(target),
  ].filter((span) => span.janitorial !== true);
}

export function janitorialTimingSpans(target) {
  return loadServiceTimingSpans(target).filter(
    (span) => span.janitorial === true,
  );
}

export function timingStatusFailed(status) {
  const normalized = String(status ?? "")
    .trim()
    .toLowerCase();
  if (
    normalized === "" ||
    normalized === "pass" ||
    normalized === "succeeded"
  ) {
    return false;
  }
  return true;
}

export function createDurationFields() {
  return {
    wall_duration_ms: 0,
    critical_path_wall_duration_ms: 0,
    executed_duration_ms: 0,
    logical_duration_ms: 0,
    reused_duration_ms: 0,
    derived_duration_ms: 0,
    teardown_duration_ms: 0,
  };
}

export function readSummaryDurationFields(
  summary,
  accountingMode = normalizeAccountingMode(summary?.accounting_mode),
) {
  const logicalDurationMs = clampDurationMs(
    summary?.logical_duration_ms ?? summary?.duration_ms ?? 0,
  );
  const wallDurationMs = clampDurationMs(
    summary?.wall_duration_ms ??
      (accountingMode === "actual" ? logicalDurationMs : 0),
  );
  return {
    wall_duration_ms: wallDurationMs,
    critical_path_wall_duration_ms: clampDurationMs(
      summary?.critical_path_wall_duration_ms ?? wallDurationMs,
    ),
    executed_duration_ms: clampDurationMs(
      summary?.executed_duration_ms ??
        (accountingMode === "actual" ? logicalDurationMs : 0),
    ),
    logical_duration_ms: logicalDurationMs,
    reused_duration_ms: clampDurationMs(
      summary?.reused_duration_ms ??
        (accountingMode === "reused" ? logicalDurationMs : 0),
    ),
    derived_duration_ms: clampDurationMs(
      summary?.derived_duration_ms ??
        (accountingMode === "derived" ? logicalDurationMs : 0),
    ),
    teardown_duration_ms: clampDurationMs(
      summary?.teardown_duration_ms ??
        (summary?.timing_bucket === "teardown" ? wallDurationMs : 0),
    ),
  };
}

export function addDurationFields(target, fields) {
  for (const key of Object.keys(createDurationFields())) {
    target[key] += clampDurationMs(fields?.[key] ?? 0);
  }
}

export function durationFieldsForJSON(fields, overrides = {}) {
  return {
    wall_duration_ms: clampDurationMs(
      overrides.wall_duration_ms ?? fields.wall_duration_ms,
    ),
    critical_path_wall_duration_ms: clampDurationMs(
      overrides.critical_path_wall_duration_ms ??
        fields.critical_path_wall_duration_ms,
    ),
    executed_duration_ms: clampDurationMs(
      overrides.executed_duration_ms ?? fields.executed_duration_ms,
    ),
    logical_duration_ms: clampDurationMs(
      overrides.logical_duration_ms ?? fields.logical_duration_ms,
    ),
    reused_duration_ms: clampDurationMs(
      overrides.reused_duration_ms ?? fields.reused_duration_ms,
    ),
    derived_duration_ms: clampDurationMs(
      overrides.derived_duration_ms ?? fields.derived_duration_ms,
    ),
    teardown_duration_ms: clampDurationMs(
      overrides.teardown_duration_ms ?? fields.teardown_duration_ms,
    ),
  };
}

function timingFailureReference(span) {
  return {
    source: span.source ?? "",
    bucket: span.bucket ?? "",
    label: span.label ?? "",
    status: span.status ?? "",
    start_time: span.start_time ?? "",
    end_time: span.end_time ?? "",
    wall_duration_ms: clampDurationMs(span.duration_ms ?? 0),
    artifact: span.artifact ?? "",
  };
}

function causalTimingFailureSpan(span) {
  const source = String(span?.source ?? "").toLowerCase();
  const bucket = String(span?.bucket ?? "").toLowerCase();
  const label = String(span?.label ?? "").toLowerCase();
  if (source === "target" && (bucket === "test_command" || bucket === "report_collation")) {
    return false;
  }
  if (source === "test_services") {
    return true;
  }
  if (bucket === "teardown" && (label.includes("cleanup") || label.includes("janitor") || label.includes("leak"))) {
    return true;
  }
  return (
    label.includes("timeout") ||
    label.includes("deadline") ||
    label.includes("watchdog") ||
    label.includes("lock") ||
    label.includes("duration-baseline-drift") ||
    label.includes("scheduler-summary-timing-drift")
  );
}

function retryScheduledStartupAttempt(span) {
  return (
    span?.startup_attempt === true &&
    span?.retry_scheduled === true &&
    span?.retry_blocked_by_context !== true
  );
}

export function timingFailuresFromSpans(spans) {
  return spans
    .filter(
      (span) =>
        timingStatusFailed(span.status) && !retryScheduledStartupAttempt(span),
    )
    .filter(causalTimingFailureSpan)
    .map(timingFailureReference);
}

export function teardownStatus(teardownDurationMs, teardownFailures) {
  if (teardownFailures.length > 0) {
    return "fail";
  }
  if (teardownDurationMs > 0) {
    return "pass";
  }
  return "none";
}

function accountableTargetWallSpan(span) {
  if (!span || span.bucket === "report_collation") {
    return false;
  }
  if (span.source === "phase") {
    return normalizeAccountingMode(span.accounting_mode) === "actual";
  }
  return true;
}

function summarizeAccountableTargetWindow(spans) {
  let accountableSpans = spans.filter(accountableTargetWallSpan);
  if (accountableSpans.some((span) => span.source === "scheduler")) {
    accountableSpans = accountableSpans.filter(
      (span) => span.source !== "phase",
    );
  }
  const summedDurationMs = accountableSpans.reduce(
    (total, span) => total + clampDurationMs(span.duration_ms ?? 0),
    0,
  );
  const startTimes = accountableSpans
    .map((span) => span.start_time)
    .filter((value) => Number.isFinite(Date.parse(value)))
    .sort((left, right) => Date.parse(left) - Date.parse(right));
  const endTimes = accountableSpans
    .map((span) => span.end_time)
    .filter((value) => Number.isFinite(Date.parse(value)))
    .sort((left, right) => Date.parse(left) - Date.parse(right));
  const startTime = startTimes[0] ?? "";
  const endTime = endTimes[endTimes.length - 1] ?? "";
  const windowDurationMs = computeWindowDurationMs(startTime, endTime);
  const wallDurationMs =
    accountableSpans.length === 1
      ? summedDurationMs
      : windowDurationMs > 0
        ? windowDurationMs
        : summedDurationMs;
  return {
    startTime,
    endTime,
    wallDurationMs,
  };
}

export function summarizeTargetTiming(
  target,
  targetDir,
  phaseSummaries,
  status,
  reportCollationSpan,
  lifecycleSpans = lifecycleTimingSpans(target, targetDir),
) {
  const buckets = new Map();
  const phaseSpans = phaseSummaries.map((summary) =>
    phaseSummaryTimingSpan(summary),
  );
  for (const span of phaseSpans) {
    addTimingSpanToBuckets(buckets, span);
  }
  for (const span of lifecycleSpans) {
    addTimingSpanToBuckets(buckets, span);
  }
  addTimingSpanToBuckets(buckets, reportCollationSpan);
  const accountableWindow = summarizeAccountableTargetWindow([
    ...phaseSpans,
    ...lifecycleSpans,
  ]);
  const summaryWindow = {
    ...accountableWindow,
    startTime: accountableWindow.startTime || reportCollationSpan.start_time,
    endTime: accountableWindow.endTime || reportCollationSpan.end_time,
  };

  const bucketList = timingBucketOrder
    .map((name) => buckets.get(name))
    .filter(Boolean)
    .map((bucket) => ({
      ...bucket,
      duration_ms: disjointSpanDurationMs(bucket.spans),
      spans: bucket.spans.sort((left, right) =>
        `${left.start_time ?? ""}:${left.label ?? ""}`.localeCompare(
          `${right.start_time ?? ""}:${right.label ?? ""}`,
        ),
      ),
    }));
  const slowest = bucketList.reduce((current, bucket) => {
    if (!current || bucket.duration_ms > current.duration_ms) {
      return { name: bucket.name, duration_ms: bucket.duration_ms };
    }
    return current;
  }, null);
  const timing = {
    schema_id: targetTimingSchemaID,
    target,
    status,
    generated_at: new Date().toISOString(),
    start_time: summaryWindow.startTime,
    end_time: summaryWindow.endTime,
    buckets: bucketList,
    slowest_lifecycle_bucket: slowest,
  };
  const timingPath = path.join(targetDir, "target-timing.json");
  writeJson(timingPath, timing);
  return { timing, timingPath, accountableWindow: summaryWindow };
}
