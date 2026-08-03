#!/usr/bin/env node
import { repoRoot } from "../../contract/index.mjs";

import {
  existsSync,
  readdirSync,
  readFileSync,
} from "node:fs";
import path from "node:path";
import {
  artifactFailureRecord,
  classifyExecutionFailure,
  failureFieldsForJSON,
  failureHeadlineForSummary,
  timingFailureRecord,
} from "../../contract/failure-taxonomy.mjs";
import {
  combineFixtureSummaries,
  emptyFixtureSummary,
  fixtureSummaryLine,
  normalizeFixtureSummary,
  summarizeFixtureActivities,
} from "../../diagnostics/fixture-reporting.mjs";
import {
  compactJSONString,
  prettyJSONString,
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../../contract/harness-contract.mjs";
import {
  loadSummaryTopologyContext,
  summaryProjectionChildren,
} from "../../execution/summary-topology.mjs";
import { defaultTaskSurfaceManifestPath } from "../../generated-artifacts/task-surface/model.mjs";
import { finalizeObservabilitySafely, observabilityRequiredTarget } from "../../observability/observability.mjs";
import {
  artifactLine,
  directoryArtifactRef,
  fileArtifactRef,
  buildToolRunSummary,
  machineOutput,
  normalizeOutputMode,
  quietLikeOutput,
  resultLine,
  slowestTargetRef,
  suppressChildSuccess,
  terminalArtifactPath,
  toolSummaryPath,
  verboseOutput,
} from "../tool-output.mjs";
import {
  resolveResultsRoot,
  resolveRunId,
  targetSummarySchemaID,
  testCoverageBucketSet,
  testCoverageBuckets,
  validStepCountingModes,
} from "../../contract/test-output-context.mjs";
import {
  govulncheckArtifactFailure,
  loadGovulncheckFindingsFile,
} from "./security-diagnostics.mjs";
import {
  addDurationFields,
  createDurationFields,
  durationFieldsForJSON,
  janitorialTimingSpans,
  lifecycleTimingSpans,
  readSummaryDurationFields,
  summarizeTargetTiming,
  teardownStatus,
  timingFailuresFromSpans,
} from "./timing.mjs";

const resultsRoot = resolveResultsRoot();

const runId = resolveRunId();

function firstArtifactPath(value) {
  if (!value || value === "-") {
    return "-";
  }
  const paths = String(value)
    .split(";")
    .map((part) => part.trim())
    .filter(Boolean);
  return (
    paths.find((part) =>
      existsSync(path.isAbsolute(part) ? part : path.join(repoRoot, part)),
    ) ??
    paths[0] ??
    "-"
  );
}

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

function writeValidatedJson(file, schemaID, value) {
  validateSchemaSync(schemaID, value);
  writeJson(file, value);
}

function createCounts() {
  const counts = {
    tests: 0,
    failed: 0,
    non_test: 0,
    non_test_failed: 0,
    packages: 0,
  };
  for (const coverage of testCoverageBuckets) {
    counts[coverage] = 0;
    counts[`${coverage}_failed`] = 0;
  }
  return counts;
}

function normalizeTestCoverage(value, fallback = "unmapped") {
  const normalized = String(value ?? "").trim();
  if (testCoverageBucketSet.has(normalized)) {
    return normalized;
  }
  return testCoverageBucketSet.has(fallback) ? fallback : "unmapped";
}

function failedTestCount(counts = {}) {
  return testCoverageBuckets.reduce(
    (total, coverage) => total + (counts[`${coverage}_failed`] ?? 0),
    0,
  );
}

function formatCoverageCountFields(counts = {}, { failed = false } = {}) {
  return testCoverageBuckets
    .map((coverage) => {
      const key = failed ? `${coverage}_failed` : coverage;
      return `${key}=${counts[key] ?? 0}`;
    })
    .join(" ");
}

function clampDurationMs(value) {
  if (!Number.isFinite(value) || value < 0) {
    return 0;
  }
  return value;
}

function formatDuration(durationMs) {
  if (!Number.isFinite(durationMs) || durationMs < 0) {
    return "0ms";
  }
  if (durationMs < 1000) {
    return `${Math.round(durationMs)}ms`;
  }
  const seconds = durationMs / 1000;
  if (seconds < 60) {
    return `${seconds.toFixed(seconds >= 10 ? 1 : 2)}s`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainder = seconds - minutes * 60;
  return `${minutes}m${remainder.toFixed(1)}s`;
}

function normalizeAccountingMode(value) {
  if (value === "actual" || value === "reused" || value === "derived") {
    return value;
  }
  return "actual";
}

function normalizeStepCountingMode(value) {
  if (validStepCountingModes.has(value)) {
    return value;
  }
  return "counted";
}

function createAccountingModes() {
  return {
    actual: 0,
    reused: 0,
    derived: 0,
  };
}

function mergeAccountingModes(target, source) {
  for (const mode of Object.keys(target)) {
    target[mode] += clampDurationMs(source?.[mode] ?? 0);
  }
}

function resolveAccountingModes(accountingModes, fallbackActualSteps = 0) {
  const modes = createAccountingModes();
  if (!accountingModes) {
    modes.actual = clampDurationMs(fallbackActualSteps);
    return modes;
  }
  for (const mode of Object.keys(modes)) {
    modes[mode] = clampDurationMs(accountingModes[mode] ?? 0);
  }
  return modes;
}

function formatAccountingModeFields(accountingModes) {
  const modes = resolveAccountingModes(accountingModes);
  return `actual=${modes.actual ?? 0} reused=${modes.reused ?? 0} derived=${modes.derived ?? 0}`;
}

function formatDurationFields(
  wallDurationMs,
  executedDurationMs,
  logicalDurationMs = executedDurationMs,
  criticalPathWallDurationMs = wallDurationMs,
  teardownDurationMs = 0,
) {
  const effectiveLogical = clampDurationMs(logicalDurationMs);
  const effectiveExecuted = clampDurationMs(executedDurationMs);
  const effectiveWall = Number.isFinite(wallDurationMs)
    ? wallDurationMs
    : effectiveLogical;
  const effectiveCriticalPath = Number.isFinite(criticalPathWallDurationMs)
    ? criticalPathWallDurationMs
    : effectiveWall;
  const effectiveTeardown = clampDurationMs(teardownDurationMs);
  return `wall=${formatDuration(effectiveWall)} critical=${formatDuration(effectiveCriticalPath)} exec=${formatDuration(effectiveExecuted)} logical=${formatDuration(effectiveLogical)} teardown=${formatDuration(effectiveTeardown)}`;
}

function resolveOutputMode() {
  return normalizeOutputMode();
}

function quietOutputMode() {
  return quietLikeOutput();
}

function formatBucketSummary(bucket) {
  if (!bucket) {
    return "none";
  }
  return `${bucket.name}(${formatDuration(bucket.duration_ms)})`;
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

function resolveArtifactPath(value) {
  if (!value) {
    return "";
  }
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function uniqueSortedStrings(values = []) {
  return [
    ...new Set(
      values
        .filter((value) => typeof value === "string")
        .map((value) => value.trim())
        .filter(Boolean),
    ),
  ].sort((left, right) => left.localeCompare(right));
}

function arrayFromArtifactValue(value) {
  if (Array.isArray(value)) {
    return value.flatMap((entry) => arrayFromArtifactValue(entry));
  }
  if (!value) {
    return [];
  }
  return String(value)
    .split(";")
    .map((entry) => entry.trim())
    .filter(Boolean);
}

function isDiagnosticWrapperFailure(failure) {
  return (
    failure?.source === "shell" &&
    failure?.runner === "shell" &&
    failure?.label === "(shell command)" &&
    failure?.failure_class === "harness" &&
    failure?.failure_reason === "unknown_failure" &&
    String(failure?.message ?? "").startsWith("command exited with status ")
  );
}

function demoteDiagnosticWrapperFailures(failures) {
  const hasRicherFailure = failures.some(
    (failure) =>
      !isDiagnosticWrapperFailure(failure) &&
      failure?.failure_class &&
      failure?.failure_reason,
  );
  if (!hasRicherFailure) {
    return 0;
  }
  let writeIndex = 0;
  let demoted = 0;
  for (const failure of failures) {
    if (isDiagnosticWrapperFailure(failure)) {
      demoted += 1;
      continue;
    }
    failures[writeIndex] = failure;
    writeIndex += 1;
  }
  failures.length = writeIndex;
  return demoted;
}

function summarizeTargetDir(target) {
  const targetDir = path.join(resultsRoot, runId, target);
  const summaries = [];
  if (existsSync(targetDir)) {
    const stack = [targetDir];
    while (stack.length > 0) {
      const current = stack.pop();
      for (const entry of readdirSync(current, { withFileTypes: true })) {
        const next = path.join(current, entry.name);
        if (entry.isDirectory()) {
          stack.push(next);
          continue;
        }
        if (entry.isFile() && entry.name === "step-summary.json") {
          summaries.push(JSON.parse(readFileSync(next, "utf8")));
        }
      }
    }
  }
  summaries.sort((left, right) =>
    left.start_time.localeCompare(right.start_time),
  );

  const owners = new Set();
  const inventoryByCoverage = Object.fromEntries(
    testCoverageBuckets.map((coverage) => [coverage, []]),
  );
  const counts = {
    steps: summaries.length,
    ...createCounts(),
  };
  let startTime = "";
  let endTime = "";
  let actualStartTime = "";
  let actualEndTime = "";
  const durations = createDurationFields();
  const accountingModes = createAccountingModes();
  const failures = [];
  let failed = false;

  for (const summary of summaries) {
    const accountingMode = normalizeAccountingMode(summary.accounting_mode);
    const countingMode = normalizeStepCountingMode(
      summary.counting_mode ?? "counted",
    );
    const summaryDurations = readSummaryDurationFields(summary, accountingMode);
    if (startTime === "" || summary.start_time < startTime) {
      startTime = summary.start_time;
    }
    if (endTime === "" || summary.end_time > endTime) {
      endTime = summary.end_time;
    }
    if (accountingMode === "actual") {
      if (actualStartTime === "" || summary.start_time < actualStartTime) {
        actualStartTime = summary.start_time;
      }
      if (actualEndTime === "" || summary.end_time > actualEndTime) {
        actualEndTime = summary.end_time;
      }
    }
    addDurationFields(durations, summaryDurations);
    accountingModes[accountingMode] += 1;
    if (countingMode !== "none") {
      counts.tests += summary.counts?.tests ?? 0;
      for (const coverage of testCoverageBuckets) {
        counts[coverage] += summary.counts?.[coverage] ?? 0;
        counts[`${coverage}_failed`] +=
          summary.counts?.[`${coverage}_failed`] ?? 0;
      }
    }
    counts.failed += summary.counts?.failed ?? 0;
    counts.non_test += summary.counts?.non_test ?? 0;
    counts.non_test_failed += summary.counts?.non_test_failed ?? 0;
    failures.push(...(summary.failures ?? []));
    for (const owner of summary.owners ?? []) {
      owners.add(owner);
    }
    if (countingMode !== "none") {
      for (const item of summary.inventory ?? []) {
        const coverage = normalizeTestCoverage(item.coverage);
        if (inventoryByCoverage[coverage]) {
          inventoryByCoverage[coverage].push({ ...item, coverage });
        }
      }
    }
    if (summary.status !== "pass") {
      failed = true;
    }
  }
  const demotedFailureCount = demoteDiagnosticWrapperFailures(failures);
  if (demotedFailureCount > 0) {
    counts.failed = Math.max(0, counts.failed - demotedFailureCount);
    counts.non_test = Math.max(0, counts.non_test - demotedFailureCount);
    counts.non_test_failed = Math.max(
      0,
      counts.non_test_failed - demotedFailureCount,
    );
  }
  counts.packages = owners.size;

  const actualWindowWallDurationMs = computeWindowDurationMs(
    actualStartTime,
    actualEndTime,
  );
  const wallDurationMs =
    actualWindowWallDurationMs > 0
      ? actualWindowWallDurationMs
      : durations.wall_duration_ms;

  return {
    target,
    targetDir,
    summaries,
    counts,
    startTime,
    endTime,
    durations: durationFieldsForJSON(durations, {
      wall_duration_ms: wallDurationMs,
      critical_path_wall_duration_ms: wallDurationMs,
    }),
    executedDurationMs: durations.executed_duration_ms,
    logicalDurationMs: durations.logical_duration_ms,
    wallDurationMs,
    criticalPathWallDurationMs: wallDurationMs,
    reusedDurationMs: durations.reused_duration_ms,
    derivedDurationMs: durations.derived_duration_ms,
    teardownDurationMs: durations.teardown_duration_ms,
    accountingModes,
    failures,
    failed,
    inventoryByCoverage,
  };
}

function stepAlreadyReportedGovulncheckArtifactError(summary = {}) {
  return (summary.failures ?? []).some((failure) => {
    if (failure?.failure_reason !== "artifact_error") {
      return false;
    }
    const text = [
      failure.source,
      failure.label,
      failure.message,
      failure.artifact,
    ]
      .join("\n")
      .toLowerCase();
    return text.includes("govulncheck") || text.includes("go-vulncheck");
  });
}

function govulncheckRollupFromStepSummaries(target, summaries = []) {
  const artifacts = [];
  const artifactSet = new Set();
  const failures = [];
  let status = "pass";
  let findingCount = 0;
  let blockingCount = 0;
  let validArtifactCount = 0;
  const blockingIDs = [];

  for (const summary of summaries) {
    const values = arrayFromArtifactValue(summary.artifacts?.govulncheck_findings);
    for (const value of values) {
      const artifactPath = resolveArtifactPath(value);
      const artifact = relToRepo(artifactPath);
      if (artifactSet.has(artifact)) {
        continue;
      }
      artifactSet.add(artifact);
      artifacts.push(artifact);

      const result = loadGovulncheckFindingsFile(artifactPath);
      if (result.error) {
        if (!stepAlreadyReportedGovulncheckArtifactError(summary)) {
          failures.push(
            govulncheckArtifactFailure(artifactPath, result.error, {
              target,
              step: summary.step ?? "",
              runner: summary.runner ?? "",
            }),
          );
        }
        continue;
      }
      const findings = result.findings;
      if (!findings) {
        continue;
      }
      validArtifactCount += 1;
      findingCount += findings.counts?.finding_count ?? 0;
      blockingCount += findings.counts?.blocking_count ?? 0;
      blockingIDs.push(...(findings.blocking_vulnerability_ids ?? []));
      if (findings.status === "fail" || (findings.counts?.blocking_count ?? 0) > 0) {
        status = "fail";
      }
    }
  }

  const extension =
    validArtifactCount > 0
      ? {
          "cartulary.security": {
            govulncheck: {
              status,
              finding_count: findingCount,
              blocking_count: blockingCount,
              blocking_vulnerability_ids: uniqueSortedStrings(blockingIDs),
            },
          },
        }
      : {};

  return {
    artifacts: artifacts.length > 0 ? { govulncheck_findings: artifacts } : {},
    extensions: extension,
    failures,
  };
}

function printInventory(targetSummary) {
  if (process.env.CARTULARY_TEST_INVENTORY !== "1") {
    return;
  }
  const sections = testCoverageBuckets.map((coverage) => [
    coverage,
    targetSummary.inventoryByCoverage?.[coverage] ?? [],
  ]);
  for (const [coverage, items] of sections) {
    if (items.length === 0) {
      continue;
    }
    const uniqueItems = Array.from(
      new Map(
        items.map((item) => [
          `${item.coverage}::${item.step}::${item.id}::${item.package_or_file}::${item.symbol_or_title}`,
          item,
        ]),
      ).values(),
    );
    process.stdout.write(
      `[INVENTORY] ${targetSummary.target} ${coverage}=${uniqueItems.length}\n`,
    );
    const sorted = [...uniqueItems].sort((left, right) => {
      const leftKey = `${left.step}::${left.id}::${left.package_or_file}::${left.symbol_or_title}`;
      const rightKey = `${right.step}::${right.id}::${right.package_or_file}::${right.symbol_or_title}`;
      return leftKey.localeCompare(rightKey);
    });
    for (const item of sorted) {
      process.stdout.write(
        `${coverage} step=${item.step || "-"} id=${item.id || "-"} owner=${item.package_or_file} name=${item.symbol_or_title}\n`,
      );
    }
  }
}

export function parseTargetList(value) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter((item) => item.length > 0);
}

function parseTargetSummaryArgs(args) {
  const [target, ...rest] = args;
  if (!target) {
    throw new Error(
      "usage: test-output.mjs target-summary <target> [pass|fail] [--children <target,target,...>] [--projection <target>] [--skipped-from-child <target>] [--skipped-from-scheduler <target>] [--skipped-after-failure <target,target>] [--failed-dependency <target>] [--quiet-success] [--quiet-failure] [--suppress-machine-output] [--preserve-existing-tool-summary]",
    );
  }

  let requestedStatus = "pass";
  let projectionTarget = "";
  let quietSuccess = false;
  let quietFailure = false;
  let suppressMachineOutput = false;
  let preserveExistingToolSummary = false;
  let skippedAfterFailure = [];
  let skippedFromChildTargets = [];
  let skippedFromSchedulerTargets = [];
  let failedDependency = "";
  const remaining = [...rest];
  if (remaining.length > 0 && !remaining[0].startsWith("--")) {
    requestedStatus = remaining.shift();
  }

  const childTargetNames = [];
  while (remaining.length > 0) {
    const option = remaining.shift();
    if (option === "--quiet-success") {
      quietSuccess = true;
      continue;
    }
    if (option === "--quiet-failure") {
      quietFailure = true;
      continue;
    }
    if (option === "--suppress-machine-output") {
      suppressMachineOutput = true;
      continue;
    }
    if (option === "--preserve-existing-tool-summary") {
      preserveExistingToolSummary = true;
      continue;
    }
    if (option === "--projection") {
      projectionTarget = remaining.shift() ?? "";
      if (projectionTarget === "") {
        throw new Error("--projection requires a target name");
      }
      continue;
    }
    if (option === "--skipped-after-failure") {
      const value = remaining.shift();
      if (value === undefined) {
        throw new Error("--skipped-after-failure requires <target,target>");
      }
      skippedAfterFailure = skippedAfterFailure.concat(parseTargetList(value));
      continue;
    }
    if (option === "--skipped-from-child") {
      const value = remaining.shift();
      if (value === undefined) {
        throw new Error("--skipped-from-child requires <target>");
      }
      skippedFromChildTargets = skippedFromChildTargets.concat(
        parseTargetList(value),
      );
      continue;
    }
    if (option === "--skipped-from-scheduler") {
      const value = remaining.shift();
      if (value === undefined) {
        throw new Error("--skipped-from-scheduler requires <target>");
      }
      skippedFromSchedulerTargets = skippedFromSchedulerTargets.concat(
        parseTargetList(value),
      );
      continue;
    }
    if (option === "--failed-dependency") {
      failedDependency = remaining.shift() ?? "";
      if (failedDependency === "") {
        throw new Error("--failed-dependency requires a target name");
      }
      continue;
    }
    if (option !== "--children") {
      throw new Error(`unknown target-summary option ${option}`);
    }
    const value = remaining.shift();
    if (value === undefined) {
      throw new Error("--children requires a comma-separated target list");
    }
    childTargetNames.push(...parseTargetList(value));
  }

  if (projectionTarget && childTargetNames.length === 0) {
    const context = loadSummaryTopologyContext({
      taskSurfaceManifestPath:
        process.env.TASK_SURFACE_MANIFEST ?? defaultTaskSurfaceManifestPath,
      schedulerManifestPath: process.env.SCHEDULER_MANIFEST,
      browserBatchManifestPath: process.env.BROWSER_E2E_BATCH_MANIFEST,
    });
    childTargetNames.push(
      ...summaryProjectionChildren(context, projectionTarget),
    );
  }

  return {
    target,
    requestedStatus,
    childTargetNames,
    skippedAfterFailure,
    skippedFromChildTargets,
    skippedFromSchedulerTargets,
    failedDependency,
    quietSuccess,
    quietFailure,
    suppressMachineOutput,
    preserveExistingToolSummary,
  };
}

export function targetSummaryPath(target) {
  return path.join(resultsRoot, runId, target, "target-summary.json");
}

function schedulerSummaryPath(target) {
  return path.join(resultsRoot, runId, target, "scheduler-summary.json");
}

export function loadTargetSummary(target) {
  const file = targetSummaryPath(target);
  if (!existsSync(file)) {
    return undefined;
  }
  return JSON.parse(readFileSync(file, "utf8"));
}

function loadSchedulerSummary(target) {
  const file = schedulerSummaryPath(target);
  if (!existsSync(file)) {
    return undefined;
  }
  return JSON.parse(readFileSync(file, "utf8"));
}

function schedulerTimingFromSummary(summary) {
  if (!summary || !Number.isFinite(summary.scheduler_total_duration_ms)) {
    return null;
  }
  const startedAt =
    typeof summary.scheduler_started_at === "string"
      ? summary.scheduler_started_at
      : "";
  const completedAt =
    typeof summary.scheduler_completed_at === "string"
      ? summary.scheduler_completed_at
      : "";
  if (
    startedAt === "" ||
    completedAt === "" ||
    Number.isNaN(Date.parse(startedAt)) ||
    Number.isNaN(Date.parse(completedAt))
  ) {
    return null;
  }
  return {
    scheduler_kind: summary.scheduler_kind ?? "",
    scheduler_started_monotonic_ms: clampDurationMs(
      summary.scheduler_started_monotonic_ms ?? 0,
    ),
    scheduler_completed_monotonic_ms: clampDurationMs(
      summary.scheduler_completed_monotonic_ms ??
        summary.scheduler_total_duration_ms,
    ),
    scheduler_total_duration_ms: clampDurationMs(
      summary.scheduler_total_duration_ms,
    ),
    scheduler_started_at: startedAt,
    scheduler_completed_at: completedAt,
  };
}

function schedulerAccountingFromSummary(summary) {
  const accounting = summary?.extensions?.["cartulary.scheduler_accounting"];
  if (!accounting || typeof accounting !== "object") {
    return null;
  }
  return {
    reused_duration_ms: clampDurationMs(accounting.reused_duration_ms ?? 0),
    actual_duration_ms: clampDurationMs(accounting.actual_duration_ms ?? 0),
    accounting_modes: resolveAccountingModes(accounting.accounting_modes, 0),
    work_unit_accounting: Array.isArray(accounting.work_unit_accounting)
      ? accounting.work_unit_accounting
      : [],
  };
}

function schedulerFailureCoveredByChildTarget(failure, failedChildTargetSet) {
  const childTarget =
    typeof failure?.child_target === "string" && failure.child_target
      ? failure.child_target
      : typeof failure?.target === "string"
        ? failure.target
        : "";
  if (childTarget && failedChildTargetSet.has(childTarget)) {
    return true;
  }
  const workUnit =
    typeof failure?.work_unit === "string" ? failure.work_unit : "";
  if (workUnit === "") {
    return false;
  }
  for (const failedChildTarget of failedChildTargetSet) {
    if (
      workUnit === failedChildTarget ||
      workUnit.startsWith(`${failedChildTarget}:`)
    ) {
      return true;
    }
  }
  return false;
}

function addSchedulerAccountingDurations(target, accounting) {
  if (!accounting) {
    return;
  }
  target.reused_duration_ms = clampDurationMs(
    (target.reused_duration_ms ?? 0) + accounting.reused_duration_ms,
  );
  target.logical_duration_ms = clampDurationMs(
    (target.logical_duration_ms ?? 0) + accounting.reused_duration_ms,
  );
}

function schedulerTimingSpan(timing) {
  if (!timing) {
    return null;
  }
  return {
    source: "scheduler",
    bucket: "test_command",
    label: `${timing.scheduler_kind || "scheduler"} scheduler`,
    status: "pass",
    start_time: timing.scheduler_started_at,
    end_time: timing.scheduler_completed_at,
    duration_ms: timing.scheduler_total_duration_ms,
  };
}

function readJsonIfExists(file) {
  if (!existsSync(file)) {
    return null;
  }
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function normalizeCounts(counts = {}) {
  const normalized = {
    steps: clampDurationMs(counts.steps ?? 0),
    tests: clampDurationMs(counts.tests ?? 0),
    failed: clampDurationMs(counts.failed ?? 0),
    non_test: clampDurationMs(counts.non_test ?? 0),
    non_test_failed: clampDurationMs(counts.non_test_failed ?? 0),
    packages: clampDurationMs(counts.packages ?? 0),
  };
  for (const coverage of testCoverageBuckets) {
    normalized[coverage] = clampDurationMs(counts[coverage] ?? 0);
    normalized[`${coverage}_failed`] = clampDurationMs(
      counts[`${coverage}_failed`] ?? 0,
    );
  }
  return normalized;
}

function addCounts(target, source) {
  const normalized = normalizeCounts(source);
  for (const key of Object.keys(normalized)) {
    target[key] += normalized[key];
  }
}

function createStepAggregate() {
  return {
    steps: 0,
    ...createCounts(),
    ...createDurationFields(),
  };
}

function countsForJSON(aggregate) {
  const counts = {
    steps: aggregate.steps,
    tests: aggregate.tests,
    failed: aggregate.failed,
    non_test: aggregate.non_test,
    non_test_failed: aggregate.non_test_failed,
    packages: aggregate.packages,
  };
  for (const coverage of testCoverageBuckets) {
    counts[coverage] = aggregate[coverage];
    counts[`${coverage}_failed`] = aggregate[`${coverage}_failed`];
  }
  return counts;
}

function findSlowestTarget(targetSummaries) {
  return targetSummaries.reduce((current, summary) => {
    const view = targetSummaryAccountingView(summary);
    const durationMs = clampDurationMs(
      view.critical_path_wall_duration_ms ?? view.wall_duration_ms ?? view.logical_duration_ms ?? 0,
    );
    if (!current || durationMs > current.critical_path_wall_duration_ms) {
      return {
        target: summary.target,
        critical_path_wall_duration_ms: durationMs,
        basis: "critical_path_wall_duration_ms",
      };
    }
    return current;
  }, null);
}

function findSlowestLifecycleBucket(targetSummaries) {
  return targetSummaries.reduce((current, summary) => {
    const bucket = targetSummaryAccountingView(summary).slowest_lifecycle_bucket;
    if (!bucket) return current;
    const candidate = {
      target: summary.target,
      name: bucket.name,
      duration_ms: clampDurationMs(bucket.duration_ms ?? 0),
    };
    return !current || candidate.duration_ms > current.duration_ms ? candidate : current;
  }, null);
}

function sectionFromFlatSummary(summary, fallbackTarget) {
  const durations = readSummaryDurationFields(summary);
  const counts = normalizeCounts(summary?.counts ?? {});
  let failures = summary?.failures ?? [];
  if (failures.length === 0 && summary?.status && summary.status !== "pass") {
    const failedTests = failedTestCount(counts);
    failures =
      failedTests > 0
        ? [
            {
              failure_class: "product",
              kind: "test",
              target: summary?.target ?? fallbackTarget,
              message: "reported test failure",
            },
          ]
        : counts.non_test_failed > 0
          ? [
              {
                failure_class: classifyExecutionFailure(
                  summary?.target ?? fallbackTarget,
                ),
                kind: "failure",
                target: summary?.target ?? fallbackTarget,
                message: "reported non-test failure",
              },
            ]
          : [];
  }
  const failureFields = failureFieldsForJSON(failures, counts);
  return {
    target: summary?.target ?? fallbackTarget,
    status: summary?.status ?? "",
    start_time: summary?.start_time ?? "",
    end_time: summary?.end_time ?? "",
    ...durations,
    accounting_modes: resolveAccountingModes(
      summary?.accounting_modes,
      counts.steps,
    ),
    counts,
    failure_class: summary?.failure_class ?? failureFields.failure_class,
    failure_reason: summary?.failure_reason ?? failureFields.failure_reason,
    failure_classes: summary?.failure_classes ?? failureFields.failure_classes,
    failure_reasons:
      summary?.failure_reasons ?? failureFields.failure_reasons,
    failures: summary?.failures ?? failureFields.failures,
    failure_headline:
      summary?.failure_headline ?? failureFields.failure_headline,
    slowest_lifecycle_bucket: summary?.slowest_lifecycle_bucket ?? null,
    timing_failures: summary?.timing_failures ?? [],
    teardown_status:
      summary?.teardown_status ??
      teardownStatus(durations.teardown_duration_ms, []),
    teardown_failures: summary?.teardown_failures ?? [],
    fixture: normalizeFixtureSummary(
      summary?.target ?? fallbackTarget,
      summary?.fixture,
    ),
    artifacts: {
      dir:
        summary?.artifacts?.dir ??
        relToRepo(path.join(resultsRoot, runId, fallbackTarget)),
      timing_json: summary?.artifacts?.timing_json ?? "",
    },
  };
}

function targetSummarySection(summary, sectionName, fallbackTarget) {
  if (summary?.[sectionName]) {
    return sectionFromFlatSummary(
      summary[sectionName],
      summary.target ?? fallbackTarget,
    );
  }
  return sectionFromFlatSummary(summary, fallbackTarget);
}

export function targetSummaryAccountingView(
  summary,
  fallbackTarget = summary?.target ?? "",
) {
  return targetSummarySection(summary, "totals", fallbackTarget);
}

function toTargetSummaryReference(summary, fallbackTarget) {
  const own = targetSummarySection(summary, "own", fallbackTarget);
  const totals = targetSummaryAccountingView(summary, fallbackTarget);
  return {
    schema_id: summary.schema_id ?? "",
    kind: summary.kind ?? "leaf",
    target: summary.target ?? fallbackTarget,
    status: summary.status ?? "",
    start_time: totals.start_time,
    end_time: totals.end_time,
    ...durationFieldsForJSON(totals),
    counts: totals.counts,
    failure_class: totals.failure_class,
    failure_reason: totals.failure_reason,
    failure_classes: totals.failure_classes,
    failure_reasons: totals.failure_reasons,
    failures: totals.failures,
    failure_headline: totals.failure_headline,
    accounting_modes: totals.accounting_modes,
    slowest_lifecycle_bucket: totals.slowest_lifecycle_bucket,
    timing_failures: totals.timing_failures,
    teardown_status: totals.teardown_status,
    teardown_failures: totals.teardown_failures,
    fixture: totals.fixture,
    artifacts: own.artifacts,
    own,
    children: summary.children ?? {
      expected: [],
      present: [],
      missing: [],
      skipped: [],
      status: "pass",
      ...durationFieldsForJSON(createDurationFields()),
      accounting_modes: createAccountingModes(),
      counts: normalizeCounts(),
      ...failureFieldsForJSON([], normalizeCounts()),
      fixture: emptyFixtureSummary(summary.target ?? fallbackTarget),
      failed_targets: [],
    },
    totals,
  };
}

function loadChildTargetSummaries(childTargetNames) {
  const childTargets = [];
  const missingChildTargetSummaries = [];
  for (const childTarget of childTargetNames) {
    const summary = loadTargetSummary(childTarget);
    if (!summary) {
      missingChildTargetSummaries.push(childTarget);
      continue;
    }
    childTargets.push(toTargetSummaryReference(summary, childTarget));
  }
  return { childTargets, missingChildTargetSummaries };
}

function skippedChildTargetSummaries(
  parentTarget,
  missingChildTargetSummaries,
  explicitSkippedAfterFailure = [],
  skippedFromChildTargets = [],
  skippedFromSchedulerTargets = [],
  explicitFailedDependency = "",
) {
  const missing = new Set(missingChildTargetSummaries);
  if (missing.size === 0) {
    return [];
  }
  const skippedByTarget = new Map();
  for (const schedulerTarget of [
    parentTarget,
    ...skippedFromSchedulerTargets,
  ]) {
    const schedulerSummary = loadSchedulerSummary(schedulerTarget);
    if (!schedulerSummary) {
      continue;
    }
    for (const skipped of schedulerSummary.skipped_work_units ?? []) {
      const childTarget = skipped.aggregate_target;
      if (!missing.has(childTarget) || skippedByTarget.has(childTarget)) {
        continue;
      }
      skippedByTarget.set(childTarget, {
        target: childTarget,
        work_unit: skipped.label ?? skipped.id ?? childTarget,
        reason: skipped.reason ?? "unknown",
        failed_dependency:
          skipped.failed_dependency ?? schedulerSummary.failed_work_unit ?? "",
      });
    }
  }
  for (const sourceTarget of skippedFromChildTargets) {
    const sourceSummary = loadTargetSummary(sourceTarget);
    const skippedChildren = sourceSummary?.children?.skipped;
    if (!Array.isArray(skippedChildren)) {
      continue;
    }
    for (const skipped of skippedChildren) {
      const childTarget =
        typeof skipped?.target === "string" ? skipped.target : "";
      if (!missing.has(childTarget) || skippedByTarget.has(childTarget)) {
        continue;
      }
      skippedByTarget.set(childTarget, {
        target: childTarget,
        work_unit:
          typeof skipped.work_unit === "string" && skipped.work_unit
            ? skipped.work_unit
            : childTarget,
        reason:
          typeof skipped.reason === "string" && skipped.reason
            ? skipped.reason
            : "unknown",
        failed_dependency:
          typeof skipped.failed_dependency === "string"
            ? skipped.failed_dependency
            : "",
      });
    }
  }
  for (const childTarget of explicitSkippedAfterFailure) {
    if (!missing.has(childTarget) || skippedByTarget.has(childTarget)) {
      continue;
    }
    skippedByTarget.set(childTarget, {
      target: childTarget,
      work_unit: childTarget,
      reason: "schedule_stopped_after_failure",
      failed_dependency: explicitFailedDependency,
    });
  }
  return missingChildTargetSummaries
    .filter((childTarget) => skippedByTarget.has(childTarget))
    .map((childTarget) => skippedByTarget.get(childTarget));
}

function combineSummarySections(target, sections, status = "pass") {
  const aggregate = createStepAggregate();
  const accountingModes = createAccountingModes();
  const timingFailures = [];
  const teardownFailures = [];
  const failures = [];
  let startTime = "";
  let endTime = "";
  let failed = status !== "pass";

  for (const section of sections) {
    const view = targetSummaryAccountingView(
      section,
      section?.target ?? target,
    );
    const viewFailures = [...(view.failures ?? [])];
    if (view.status && view.status !== "pass" && viewFailures.length === 0) {
      viewFailures.push({
        failure_class: "harness",
        failure_reason: "child_target_failure",
        kind: "failure",
        source: "target-summary",
        target,
        child_target: view.target ?? section?.target ?? "",
        label: view.target ?? section?.target ?? "",
        message: "child target reported failure status without failure details",
        artifact: view.artifacts?.dir ?? "",
      });
    }
    addCounts(aggregate, view.counts);
    addDurationFields(aggregate, view);
    mergeAccountingModes(accountingModes, view.accounting_modes);
    timingFailures.push(...(view.timing_failures ?? []));
    teardownFailures.push(...(view.teardown_failures ?? []));
    failures.push(...viewFailures);
    if (startTime === "" || (view.start_time && view.start_time < startTime)) {
      startTime = view.start_time ?? "";
    }
    if (endTime === "" || (view.end_time && view.end_time > endTime)) {
      endTime = view.end_time ?? "";
    }
    if (view.status && view.status !== "pass") {
      failed = true;
    }
  }

  const windowWallDurationMs = computeWindowDurationMs(startTime, endTime);
  const wallDurationMs =
    windowWallDurationMs > 0
      ? windowWallDurationMs
      : aggregate.wall_duration_ms;
  const criticalPathWallDurationMs = wallDurationMs;
  const failureFields = failureFieldsForJSON(
    failures,
    countsForJSON(aggregate),
  );
  return {
    target,
    status: failed ? "fail" : "pass",
    start_time: startTime,
    end_time: endTime,
    ...durationFieldsForJSON(aggregate, {
      wall_duration_ms: wallDurationMs,
      critical_path_wall_duration_ms: criticalPathWallDurationMs,
    }),
    accounting_modes: accountingModes,
    counts: countsForJSON(aggregate),
    ...failureFields,
    slowest_lifecycle_bucket: findSlowestLifecycleBucket(sections),
    timing_failures: timingFailures,
    teardown_status: teardownStatus(
      aggregate.teardown_duration_ms,
      teardownFailures,
    ),
    teardown_failures: teardownFailures,
  };
}

function writeTargetLine(stream, label, targetSummary) {
  const target = targetSummary.target;
  const failureClassField = targetSummary.failure_class
    ? ` failure_class=${targetSummary.failure_class}`
    : "";
  if (targetSummary.kind === "aggregate") {
    const own = targetSummary.own;
    const children = targetSummary.children;
    const totals = targetSummary.totals;
    const slowestChild = findSlowestTarget(children.present ?? []);
    const slowestChildField = slowestChild
      ? `${slowestChild.target}(${formatDuration(slowestChild.critical_path_wall_duration_ms)})`
      : "none";
    const failedChildren =
      (children.failed_targets ?? []).length > 0
        ? children.failed_targets.join(",")
        : "none";
    stream.write(
      `${label} ${target} kind=aggregate${failureClassField} children=${children.present.length}/${children.expected.length} child_tests=${children.counts.tests} child_failed=${children.counts.failed} failed_children=${failedChildren} slowest_child=${slowestChildField} own_steps=${own.counts.steps} own_tests=${own.counts.tests} own_failed=${own.counts.failed} total_tests=${totals.counts.tests} total_failed=${totals.counts.failed} ${formatDurationFields(totals.wall_duration_ms, totals.executed_duration_ms, totals.logical_duration_ms, totals.critical_path_wall_duration_ms, totals.teardown_duration_ms)} ${formatAccountingModeFields(totals.accounting_modes)} own_fixture_count=${own.fixture.total_count} own_fixture_duration=${formatDuration(own.fixture.total_duration_ms)} child_fixture_count=${children.fixture.total_count} child_fixture_duration=${formatDuration(children.fixture.total_duration_ms)} total_fixture_count=${totals.fixture.total_count} total_fixture_duration=${formatDuration(totals.fixture.total_duration_ms)} slowest_lifecycle_bucket=${formatBucketSummary(totals.slowest_lifecycle_bucket)} artifacts=${targetSummary.own.artifacts.dir}\n`,
    );
    return;
  }

  const totals = targetSummary.totals;
  stream.write(
    `${label} ${target} kind=leaf${failureClassField} steps=${totals.counts.steps} tests=${totals.counts.tests} failed=${totals.counts.failed} ${formatCoverageCountFields(totals.counts)} packages=${totals.counts.packages} ${formatDurationFields(totals.wall_duration_ms, totals.executed_duration_ms, totals.logical_duration_ms, totals.critical_path_wall_duration_ms, totals.teardown_duration_ms)} ${formatAccountingModeFields(totals.accounting_modes)} fixture_count=${totals.fixture.total_count} fixture_duration=${formatDuration(totals.fixture.total_duration_ms)} slowest_lifecycle_bucket=${formatBucketSummary(totals.slowest_lifecycle_bucket)} artifacts=${targetSummary.own.artifacts.dir}\n`,
  );
}

function govulncheckFindingArtifactRefs(targetSummary) {
  return arrayFromArtifactValue(targetSummary.artifacts?.govulncheck_findings).map(
    (artifact) => fileArtifactRef("govulncheck_findings", artifact, "json"),
  );
}

function targetToolSummary(targetSummary, summaryJsonPath) {
  const totals = targetSummary.totals ?? {};
  const counts = totals.counts ?? {};
  const targetArtifactRoot =
    targetSummary.artifacts?.dir ?? targetSummary.own?.artifacts?.dir ?? "";
  const missingChildren = targetSummary.children?.missing ?? [];
  const runRunRoot = path.join(resultsRoot, runId);
  const runRoot = relToRepo(runRunRoot);
  const runSummaryFile = path.join(runRunRoot, "run-summary.json");
  const runToolSummaryFile = toolSummaryPath(runRunRoot);
  let ownsRunArtifacts = false;
  if (existsSync(runSummaryFile) && existsSync(runToolSummaryFile)) {
    const runToolSummary = JSON.parse(readFileSync(runToolSummaryFile, "utf8"));
    validateSchemaSync("cartulary.tool_run_summary.v5", runToolSummary);
    ownsRunArtifacts = runToolSummary.target === targetSummary.target;
  }
  const schedulerSummaryFile = schedulerSummaryPath(targetSummary.target);
  const schedulerSummary = existsSync(schedulerSummaryFile)
    ? loadSchedulerSummary(targetSummary.target)
    : null;
  const schedulerTiming = schedulerTimingFromSummary(schedulerSummary);
  const schedulerArtifacts = schedulerSummary?.artifacts ?? {};
  const browserArtifacts = browserOwnedStackArtifacts(targetArtifactRoot, targetSummary.target);
  const serviceMetadata = serviceSharedMetadata(runRunRoot);
  const workUnits =
    targetSummary.kind === "aggregate"
      ? [
          {
            id: targetSummary.target,
            completed: targetSummary.children?.present?.length ?? 0,
            total: targetSummary.children?.expected?.length ?? 0,
            status: targetSummary.status,
          },
        ]
      : [];
  const slowest = slowestTargetRef(targetSummary);
  return buildToolRunSummary({
    target: targetSummary.target,
    command: ["make", targetSummary.target],
    status: targetSummary.status,
    exitCode: targetSummary.status === "pass" ? 0 : 1,
    startedAt:
      targetSummary.start_time ?? schedulerTiming?.scheduler_started_at,
    completedAt:
      targetSummary.end_time ?? schedulerTiming?.scheduler_completed_at,
    durationMs:
      totals.wall_duration_ms ??
      targetSummary.wall_duration_ms ??
      schedulerTiming?.scheduler_total_duration_ms,
    outputMode: resolveOutputMode(),
    resultRoot: relToRepo(resultsRoot),
    runId,
    runRoot,
    summaryArtifacts: [
      fileArtifactRef("tool_run_summary", summaryJsonPath),
      fileArtifactRef(
        "target_summary",
        path.join(targetArtifactRoot, "target-summary.json"),
      ),
      fileArtifactRef("target_timing", targetSummary.artifacts?.timing_json),
      browserArtifacts.stackMetadata
        ? fileArtifactRef("browser_stack", browserArtifacts.stackMetadata)
        : null,
      browserArtifacts.startupDiagnostics
        ? fileArtifactRef(
            "browser_startup_diagnostics",
            browserArtifacts.startupDiagnostics,
          )
        : null,
      ownsRunArtifacts
        ? fileArtifactRef("run_summary", relToRepo(runSummaryFile))
        : null,
      ownsRunArtifacts
        ? fileArtifactRef("run_tool_run_summary", relToRepo(runToolSummaryFile))
        : null,
      existsSync(schedulerSummaryFile)
        ? fileArtifactRef("scheduler_summary", relToRepo(schedulerSummaryFile))
        : null,
      schedulerArtifacts.events_jsonl
        ? fileArtifactRef(
            "scheduler_events",
            schedulerArtifacts.events_jsonl,
            "jsonl",
          )
        : null,
      ...govulncheckFindingArtifactRefs(targetSummary),
      ...serviceSharedSummaryArtifacts(runRunRoot),
    ],
    logArtifacts: [
      schedulerArtifacts.progress_summary_log
        ? fileArtifactRef(
            "scheduler_progress",
            schedulerArtifacts.progress_summary_log,
            "log",
          )
        : null,
      schedulerArtifacts.scheduler_logs_dir
        ? directoryArtifactRef(
            "scheduler_logs",
            schedulerArtifacts.scheduler_logs_dir,
          )
        : null,
      ...serviceSharedLogArtifacts(runRunRoot),
    ],
    workUnits,
    evidenceTargets:
      targetSummary.kind === "aggregate"
        ? (targetSummary.children?.present ?? []).map((summary) => ({
            target: summary.target,
            status: summary.status,
            run_root: runRoot,
          }))
        : [
            {
              target: targetSummary.target,
              status: targetSummary.status,
              run_root: runRoot,
            },
          ],
    helperUnits: [],
    counts,
    stepAccounting: {
      missing: missingChildren.length,
    },
    failureClass: targetSummary.failure_class,
    failureReason: targetSummary.failure_reason,
    failures: targetSummary.failures ?? [],
    slowest: slowest ? [slowest] : [],
    rerunCommands: [`make ${targetSummary.target}`],
    schedulerTiming,
    extensions: {
      ...(targetSummary.extensions ?? {}),
      ...(serviceMetadata
        ? { "cartulary.service_backed": serviceMetadata }
        : {}),
    },
  });
}

function serviceSharedLogArtifacts(runRunRoot) {
  const serviceRoot = path.join(runRunRoot, "_shared", "test-services");
  if (!existsSync(serviceRoot)) {
    return [];
  }
  return readdirSync(serviceRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort((left, right) => left.localeCompare(right))
    .flatMap((suiteID) => {
      const suiteRoot = path.join(serviceRoot, suiteID);
      return [
        ["service_child_process_log", "child-process.log"],
        ["service_goose_log", "goose.log"],
      ]
        .map(([role, filename]) => {
          const file = path.join(suiteRoot, filename);
          return existsSync(file)
            ? fileArtifactRef(role, relToRepo(file), "log")
            : null;
        })
        .filter(Boolean);
    });
}

function serviceSharedSummaryArtifacts(runRunRoot) {
  const serviceRoot = path.join(runRunRoot, "_shared", "test-services");
  if (!existsSync(serviceRoot)) {
    return [];
  }
  return readdirSync(serviceRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort((left, right) => left.localeCompare(right))
    .flatMap((suiteID) => {
      const suiteRoot = path.join(serviceRoot, suiteID);
      return [
        ["service_lease", "service-lease.json"],
        ["service_scope", "service-scope.json"],
      ]
        .map(([role, filename]) => {
          const file = path.join(suiteRoot, filename);
          return existsSync(file) ? fileArtifactRef(role, relToRepo(file)) : null;
        })
        .filter(Boolean);
    });
}

function browserOwnedStackArtifacts(targetDir) {
  const ownedStackDir = path.join(targetDir, "owned-stack");
  const stackMetadata = path.join(ownedStackDir, "stack.json");
  const startupDiagnostics = path.join(ownedStackDir, "startup-diagnostics.json");
  return {
    stackMetadata: existsSync(stackMetadata) ? relToRepo(stackMetadata) : "",
    startupDiagnostics: existsSync(startupDiagnostics)
      ? relToRepo(startupDiagnostics)
      : "",
  };
}

function statusFromBooleans(values) {
  if (values.length === 0) {
    return "unknown";
  }
  return values.every(Boolean) ? "pass" : "fail";
}

function rollupStatuses(values) {
  if (values.length === 0) {
    return "unknown";
  }
  if (values.some((value) => value === "fail")) {
    return "fail";
  }
  if (values.every((value) => value === "pass")) {
    return "pass";
  }
  return "unknown";
}

function normalizeServiceExtensionStatus(status) {
  switch (status) {
    case "pass":
    case "succeeded":
    case "skipped_no_lease":
      return "pass";
    case "fail":
    case "failed":
    case "startup_failed":
    case "child_start_failed":
    case "cleanup_failed":
      return "fail";
    default:
      return "unknown";
  }
}

function serviceSharedMetadata(runRunRoot) {
  const serviceRoot = path.join(runRunRoot, "_shared", "test-services");
  if (!existsSync(serviceRoot)) {
    return null;
  }
  const suites = readdirSync(serviceRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort((left, right) => left.localeCompare(right))
    .map((suiteID) => {
      const scope = readJsonIfExists(
        path.join(serviceRoot, suiteID, "service-scope.json"),
      );
      const serviceNames = ["postgres", "object_store"].filter(
        (name) => scope?.[name]?.started === true,
      );
      const startupStatuses = serviceNames.map(
        (name) => scope?.[name]?.startup?.final_status === "pass",
      );
      const cleanupStatus = scope?.cleanup?.status ?? "unknown";
      const extensionStatus = normalizeServiceExtensionStatus(cleanupStatus);
      return {
        suite_id: suiteID,
        services: serviceNames,
        readiness_status: statusFromBooleans(startupStatuses),
        cleanup_status: cleanupStatus,
        teardown_status: extensionStatus,
        leak_status: extensionStatus,
      };
    });
  if (suites.length === 0) {
    return null;
  }
  return {
    suite_count: suites.length,
    services: [...new Set(suites.flatMap((suite) => suite.services))].sort(
      (left, right) => left.localeCompare(right),
    ),
    readiness_status: statusFromBooleans(
      suites.map((suite) => suite.readiness_status === "pass"),
    ),
    teardown_status: rollupStatuses(
      suites.map((suite) => suite.teardown_status),
    ),
    leak_status: rollupStatuses(suites.map((suite) => suite.leak_status)),
    suites,
  };
}

export function writeToolSummary(file, summary) {
  writeValidatedJson(file, summary.schema_id, summary);
  return relToRepo(file);
}

function writeTargetResult(stream, targetSummary, summaryJsonPath) {
  const summary = targetToolSummary(targetSummary, summaryJsonPath);
  stream.write(resultLine(summary, summaryJsonPath));
  if (targetSummary.status === "pass") {
    stream.write(
      artifactLine(summary, summaryJsonPath, {
        investigate: `make explain-target TARGET=${targetSummary.target} DETAIL=artifacts`,
      }),
    );
  }
}

function writeTargetFailure(stream, targetSummary, summaryJsonPath) {
  const summary = targetToolSummary(targetSummary, summaryJsonPath);
  const failureClass = summary.failure_class ?? "harness";
  const failureReason = summary.failure_reason ?? "unknown_failure";
  const headline = targetSummary.failure_headline ?? "target failed";
  const failedChildTarget = targetSummary.children?.failed_targets?.[0] ?? "";
  const schedulerSummary = loadSchedulerSummary(summary.target);
  const failedWorkUnitDetail =
    schedulerSummary?.failed_work_unit_detail ?? null;
  const failedWorkUnitAggregateTarget =
    typeof failedWorkUnitDetail?.aggregate_target === "string"
      ? failedWorkUnitDetail.aggregate_target
      : "";
  const failedWorkUnit =
    schedulerSummary?.failed_work_unit ??
    targetSummary.children?.skipped?.[0]?.failed_dependency ??
    targetSummary.failures?.[0]?.label ??
    "";
  const failureTarget = targetSummary.failures?.[0]?.target ?? "";
  const childTarget =
    failedChildTarget ||
    failedWorkUnitAggregateTarget ||
    (failureTarget && failureTarget !== summary.target ? failureTarget : "") ||
    targetSummary.children?.missing?.[0] ||
    "";
  const logArtifact = firstArtifactPath(
    targetSummary.failures?.find((failure) => failure.artifact)?.artifact ??
      targetSummary.failures?.find((failure) => failure.raw)?.raw ??
      summary.log_artifacts?.find(
        (artifact) => artifact.role === "scheduler_progress",
      )?.path ??
      summary.log_artifacts?.[0]?.path,
  );
  const progressLog =
    summary.log_artifacts?.find(
      (artifact) => artifact.role === "scheduler_progress",
    )?.path ?? "-";
  const schedulerJson = schedulerSummary
    ? relToRepo(schedulerSummaryPath(summary.target))
    : "-";
  stream.write(
    `[FAIL] target=${summary.target} exit_code=${summary.exit_code} failure_class=${failureClass} reason=${failureReason} work_unit=${failedWorkUnit || "-"} child_target=${childTarget || "-"} duration_ms=${summary.duration_ms} headline="${headline}"\n`,
  );
  stream.write(
    `[ARTIFACTS] target=${summary.target} root=${summary.run_root} summary_json=${terminalArtifactPath(summary.run_root, summaryJsonPath)} log_artifact=${terminalArtifactPath(summary.run_root, logArtifact)} scheduler_json=${terminalArtifactPath(summary.run_root, schedulerJson)} progress_log=${terminalArtifactPath(summary.run_root, progressLog)}\n`,
  );
  stream.write(`[RERUN] command="make ${summary.target}"\n`);
  stream.write(
    `[INVESTIGATE] command="make explain-target TARGET=${summary.target} DETAIL=artifacts"\n`,
  );
}

export function writeFailureHeadline(stream, label, summary) {
  const headline = failureHeadlineForSummary(summary);
  if (headline) {
    stream.write(`[FAILURE] ${label} ${headline}\n`);
  }
}

function fixtureLineOptions() {
  return {
    thresholdMs:
      process.env.FIXTURE_THRESHOLD_MS ??
      process.env.CARTULARY_FIXTURE_THRESHOLD_MS,
    top: process.env.FIXTURE_TOP ?? process.env.CARTULARY_FIXTURE_TOP,
  };
}

export function writeFixtureLine(stream, fixture) {
  const line = fixtureSummaryLine(fixture, fixtureLineOptions());
  if (line) {
    stream.write(`${line}\n`);
  }
}

function writeChildTargetLines(
  stream,
  parentTarget,
  childTargets,
  missingChildTargetSummaries,
) {
  for (const child of childTargets) {
    const totals =
      child.totals ?? targetSummaryAccountingView(child, child.target);
    const failureClass = child.failure_class
      ? ` failure_class=${child.failure_class}`
      : "";
    stream.write(
      `[CHILD] ${parentTarget} ${child.target} status=${child.status}${failureClass} steps=${totals.counts?.steps ?? 0} tests=${totals.counts?.tests ?? 0} failed=${totals.counts?.failed ?? 0} ${formatDurationFields(totals.wall_duration_ms, totals.executed_duration_ms, totals.logical_duration_ms, totals.critical_path_wall_duration_ms, totals.teardown_duration_ms)} ${formatAccountingModeFields(totals.accounting_modes)} artifacts=${child.artifacts?.dir ?? ""}\n`,
    );
  }
  for (const childTarget of missingChildTargetSummaries) {
    stream.write(
      `[CHILD-MISSING] ${parentTarget} ${childTarget} artifacts=${relToRepo(targetSummaryPath(childTarget))}\n`,
    );
  }
}

function writeSkippedChildTargetLines(
  stream,
  parentTarget,
  skippedChildTargets,
) {
  for (const child of skippedChildTargets) {
    stream.write(
      `[CHILD-SKIPPED] ${parentTarget} ${child.target} reason=${child.reason} failed_dependency=${child.failed_dependency || "unknown"} work_unit=${child.work_unit}\n`,
    );
  }
}

function testAccountingUnmappedFailures(section, target) {
  const count = normalizeCounts(section?.counts ?? {}).unmapped;
  if (count <= 0) {
    return [];
  }
  return [
    {
      failure_class: "harness",
      failure_reason: "test_accounting_unmapped",
      kind: "failure",
      source: "test-accounting",
      target,
      label: "unmapped test accounting",
      message: `${count} executed test(s) are unmapped; map conformance evidence or classify intentional residual coverage`,
    },
  ];
}

function appendTestAccountingFailures(section, failures) {
  if (failures.length === 0) {
    return;
  }
  section.status = "fail";
  section.counts = normalizeCounts(section.counts);
  section.counts.failed += failures.length;
  section.counts.non_test += failures.length;
  section.counts.non_test_failed += failures.length;
  const failureFields = failureFieldsForJSON(
    [...(section.failures ?? []), ...failures],
    section.counts,
  );
  Object.assign(section, failureFields);
}

export function handleTargetSummary(args) {
  const reportCollationStartMs = Date.now();
  const reportCollationStartTime = new Date(
    reportCollationStartMs,
  ).toISOString();
  const {
    target,
    requestedStatus,
    childTargetNames,
    skippedAfterFailure,
    skippedFromChildTargets,
    skippedFromSchedulerTargets,
    failedDependency,
    quietSuccess,
    quietFailure,
    suppressMachineOutput,
    preserveExistingToolSummary,
  } = parseTargetSummaryArgs(args);
  const summary = summarizeTargetDir(target);
  const securityRollup = govulncheckRollupFromStepSummaries(
    target,
    summary.summaries,
  );
  const lifecycleSpans = lifecycleTimingSpans(target, summary.targetDir);
  const schedulerSummary = loadSchedulerSummary(target);
  const schedulerTiming = schedulerTimingFromSummary(schedulerSummary);
  const schedulerAccounting = schedulerAccountingFromSummary(schedulerSummary);
  const timingFailures = timingFailuresFromSpans(lifecycleSpans);
  const teardownFailures = timingFailures.filter(
    (failure) => failure.bucket === "teardown",
  );
  const { childTargets, missingChildTargetSummaries } =
    loadChildTargetSummaries(childTargetNames);
  const skippedChildTargets = skippedChildTargetSummaries(
    target,
    missingChildTargetSummaries,
    skippedAfterFailure,
    skippedFromChildTargets,
    skippedFromSchedulerTargets,
    failedDependency,
  );
  const skippedChildTargetNames = new Set(
    skippedChildTargets.map((child) => child.target),
  );
  const unresolvedMissingChildTargetSummaries =
    missingChildTargetSummaries.filter(
      (childTarget) => !skippedChildTargetNames.has(childTarget),
    );
  const failedChildTargets = childTargets
    .filter((child) => child.status !== "pass")
    .map((child) => child.target);
  const failedChildTargetSet = new Set(failedChildTargets);
  const schedulerFailuresForOwnSection =
    schedulerSummary?.status === "fail" && Array.isArray(schedulerSummary.failures)
      ? schedulerSummary.failures.filter(
          (failure) =>
            !schedulerFailureCoveredByChildTarget(
              failure,
              failedChildTargetSet,
            ),
        )
      : [];
  const schedulerNonProductFailures = schedulerFailuresForOwnSection.filter(
    (failure) => failure.failure_class !== "product",
  );
  if (timingFailures.length > 0) {
    summary.counts.failed += timingFailures.length;
    summary.counts.non_test += timingFailures.length;
    summary.counts.non_test_failed += timingFailures.length;
  }
  if (unresolvedMissingChildTargetSummaries.length > 0) {
    summary.counts.failed += unresolvedMissingChildTargetSummaries.length;
    summary.counts.non_test += unresolvedMissingChildTargetSummaries.length;
    summary.counts.non_test_failed +=
      unresolvedMissingChildTargetSummaries.length;
  }
  if (securityRollup.failures.length > 0) {
    summary.counts.failed += securityRollup.failures.length;
    summary.counts.non_test += securityRollup.failures.length;
    summary.counts.non_test_failed += securityRollup.failures.length;
  }
  if (schedulerNonProductFailures.length > 0) {
    summary.counts.failed += schedulerNonProductFailures.length;
    summary.counts.non_test += schedulerNonProductFailures.length;
    summary.counts.non_test_failed += schedulerNonProductFailures.length;
  }
  if (
    requestedStatus === "fail" &&
    summary.failed === false &&
    timingFailures.length === 0 &&
    unresolvedMissingChildTargetSummaries.length === 0 &&
    securityRollup.failures.length === 0 &&
    failedChildTargets.length === 0 &&
    skippedChildTargets.length === 0
  ) {
    summary.counts.failed += 1;
    summary.counts.non_test += 1;
    summary.counts.non_test_failed += 1;
  }
  const requestedFallbackFailure =
    requestedStatus === "fail" &&
    summary.failed === false &&
    timingFailures.length === 0 &&
    unresolvedMissingChildTargetSummaries.length === 0 &&
    securityRollup.failures.length === 0 &&
    failedChildTargets.length === 0 &&
    skippedChildTargets.length === 0
      ? [
          {
            failure_class: classifyExecutionFailure(target),
            kind: "failure",
            source: "target-summary",
            target,
            message: `${target} failed before a test or child failure was attributed`,
          },
        ]
      : [];
  const ownFailureFields = failureFieldsForJSON(
    [
      ...summary.failures,
      ...timingFailures.map((failure) =>
        timingFailureRecord(failure, { target }),
      ),
      ...unresolvedMissingChildTargetSummaries.map((childTarget) =>
        artifactFailureRecord(`missing child target summary: ${childTarget}`, {
          target,
          source: "target-summary",
        }),
      ),
      ...securityRollup.failures,
      ...schedulerFailuresForOwnSection,
      ...requestedFallbackFailure,
    ],
    normalizeCounts(summary.counts),
  );
  const ownFailed =
    summary.failed ||
    timingFailures.length > 0 ||
    unresolvedMissingChildTargetSummaries.length > 0 ||
    securityRollup.failures.length > 0 ||
    (requestedStatus === "fail" &&
      failedChildTargets.length === 0 &&
      skippedChildTargets.length === 0);
  const status =
    ownFailed || failedChildTargets.length > 0 || requestedStatus === "fail"
      ? "FAIL"
      : "PASS";
  const reportCollationEndMs = Date.now();
  const reportCollationEndTime = new Date(reportCollationEndMs).toISOString();
  const accountableTimingSpans = [
    ...lifecycleSpans,
    schedulerTimingSpan(schedulerTiming),
  ].filter(Boolean);
  const { timing, timingPath, accountableWindow } = summarizeTargetTiming(
    target,
    summary.targetDir,
    summary.summaries,
    status.toLowerCase(),
    {
      source: "target",
      bucket: "report_collation",
      label: "target summary collation",
      start_time: reportCollationStartTime,
      end_time: reportCollationEndTime,
      duration_ms: clampDurationMs(
        reportCollationEndMs - reportCollationStartMs,
      ),
      status: status.toLowerCase(),
    },
    accountableTimingSpans,
  );
  summary.wallDurationMs = accountableWindow.wallDurationMs;
  summary.criticalPathWallDurationMs = accountableWindow.wallDurationMs;
  summary.startTime = accountableWindow.startTime;
  summary.endTime = accountableWindow.endTime;
  summary.slowestLifecycleBucket = timing.slowest_lifecycle_bucket;
  const ownFixture = summarizeFixtureActivities(target, {
    resultsRoot,
    runId,
  });
  const childFixture = combineFixtureSummaries(target, null, childTargets);
  const totalFixture = combineFixtureSummaries(
    target,
    ownFixture,
    childTargets,
  );
  summary.teardownDurationMs = clampDurationMs(
    timing.buckets.find((bucket) => bucket.name === "teardown")?.duration_ms ??
      0,
  );
  summary.durations = durationFieldsForJSON(summary.durations, {
    wall_duration_ms: summary.wallDurationMs,
    critical_path_wall_duration_ms: summary.criticalPathWallDurationMs,
    teardown_duration_ms: summary.teardownDurationMs,
  });
  const ownSection = {
    target,
    status: ownFailed ? "fail" : "pass",
    start_time: summary.startTime,
    end_time: summary.endTime,
    ...summary.durations,
    accounting_modes: summary.accountingModes,
    counts: normalizeCounts(summary.counts),
    ...ownFailureFields,
    slowest_lifecycle_bucket: timing.slowest_lifecycle_bucket,
    timing_failures: timingFailures,
    janitorial_timing: janitorialTimingSpans(target),
    teardown_status: teardownStatus(
      summary.teardownDurationMs,
      teardownFailures,
    ),
    teardown_failures: teardownFailures,
    fixture: ownFixture,
    inventory_by_coverage: summary.inventoryByCoverage,
    artifacts: {
      dir: relToRepo(summary.targetDir),
      timing_json: relToRepo(timingPath),
    },
  };
  const childrenRollup = combineSummarySections(target, childTargets);
  const missingChildFailures = unresolvedMissingChildTargetSummaries.map(
    (childTarget) =>
      artifactFailureRecord(`missing child target summary: ${childTarget}`, {
        target,
        source: "target-summary",
      }),
  );
  const childrenFailureFields = failureFieldsForJSON(
    [...(childrenRollup.failures ?? []), ...missingChildFailures],
    normalizeCounts(childrenRollup.counts),
  );
  const childrenSection = {
    target,
    status:
      failedChildTargets.length > 0 ||
      unresolvedMissingChildTargetSummaries.length > 0 ||
      skippedChildTargets.length > 0
        ? "fail"
        : "pass",
    expected: childTargetNames,
    present: childTargets,
    missing: unresolvedMissingChildTargetSummaries,
    skipped: skippedChildTargets,
    failed_targets: failedChildTargets,
    start_time: childrenRollup.start_time,
    end_time: childrenRollup.end_time,
    ...durationFieldsForJSON(childrenRollup),
    accounting_modes: childrenRollup.accounting_modes,
    counts: childrenRollup.counts,
    ...childrenFailureFields,
    slowest_lifecycle_bucket: childrenRollup.slowest_lifecycle_bucket,
    timing_failures: childrenRollup.timing_failures,
    teardown_status: childrenRollup.teardown_status,
    teardown_failures: childrenRollup.teardown_failures,
    fixture: childFixture,
  };
  const totalRollup =
    childTargetNames.length === 0
      ? ownSection
      : combineSummarySections(
          target,
          [ownSection, childrenSection],
          status.toLowerCase(),
        );
  const totalsSection =
    childTargetNames.length === 0
      ? { ...ownSection, fixture: ownFixture }
      : {
          ...totalRollup,
          status: status.toLowerCase(),
          fixture: totalFixture,
        };
  if (schedulerTiming) {
    totalsSection.start_time = ownSection.start_time;
    totalsSection.end_time = ownSection.end_time;
    totalsSection.wall_duration_ms = ownSection.wall_duration_ms;
    totalsSection.critical_path_wall_duration_ms =
      ownSection.critical_path_wall_duration_ms;
  }
  if (schedulerAccounting) {
    addSchedulerAccountingDurations(totalsSection, schedulerAccounting);
    const modes = resolveAccountingModes(totalsSection.accounting_modes, 0);
    mergeAccountingModes(modes, schedulerAccounting.accounting_modes);
    totalsSection.accounting_modes = modes;
  }
  if (securityRollup.artifacts.govulncheck_findings?.length > 0) {
    ownSection.artifacts.govulncheck_findings =
      securityRollup.artifacts.govulncheck_findings;
    if (totalsSection.artifacts && typeof totalsSection.artifacts === "object") {
      totalsSection.artifacts.govulncheck_findings =
        securityRollup.artifacts.govulncheck_findings;
    }
  }
  const schedulerFailureOverride =
    status === "FAIL" &&
    schedulerSummary?.status === "fail" &&
    schedulerSummary.failure_class &&
    schedulerSummary.failure_reason
      ? schedulerSummary
      : null;
  const topLevelTiming = {
    start_time:
      totalsSection.start_time ??
      summary.startTime ??
      schedulerTiming?.scheduler_started_at ??
      reportCollationStartTime,
    end_time:
      totalsSection.end_time ??
      summary.endTime ??
      schedulerTiming?.scheduler_completed_at ??
      reportCollationEndTime,
    ...durationFieldsForJSON(totalsSection),
  };
  const browserArtifacts = browserOwnedStackArtifacts(summary.targetDir);
  if (browserArtifacts.stackMetadata) {
    ownSection.artifacts.browser_stack = browserArtifacts.stackMetadata;
    if (totalsSection.artifacts && typeof totalsSection.artifacts === "object") {
      totalsSection.artifacts.browser_stack = browserArtifacts.stackMetadata;
    }
  }
  if (browserArtifacts.startupDiagnostics) {
    ownSection.artifacts.browser_startup_diagnostics =
      browserArtifacts.startupDiagnostics;
    if (totalsSection.artifacts && typeof totalsSection.artifacts === "object") {
      totalsSection.artifacts.browser_startup_diagnostics =
        browserArtifacts.startupDiagnostics;
    }
  }
  const testAccountingFailures =
    status === "PASS" ? testAccountingUnmappedFailures(ownSection, target) : [];
  appendTestAccountingFailures(ownSection, testAccountingFailures);
  appendTestAccountingFailures(totalsSection, testAccountingFailures);
  const finalStatus = testAccountingFailures.length > 0 ? "FAIL" : status;
  const targetExtensions = {
    ...securityRollup.extensions,
    ...(schedulerAccounting
      ? {
          "cartulary.scheduler_accounting": schedulerAccounting,
        }
      : {}),
  };
  const targetSummary = {
    schema_id: targetSummarySchemaID,
    target,
    kind: childTargetNames.length > 0 ? "aggregate" : "leaf",
    status: finalStatus.toLowerCase(),
    ...topLevelTiming,
    accounting_modes: totalsSection.accounting_modes,
    failure_class:
      schedulerFailureOverride?.failure_class ?? totalsSection.failure_class,
    failure_reason:
      schedulerFailureOverride?.failure_reason ?? totalsSection.failure_reason,
    failure_classes: totalsSection.failure_classes,
    failure_reasons: totalsSection.failure_reasons,
    failures: totalsSection.failures,
    failure_headline:
      schedulerFailureOverride?.failure_headline ??
      totalsSection.failure_headline,
    artifacts: ownSection.artifacts,
    own: ownSection,
    children: childrenSection,
    totals: {
      ...totalsSection,
      start_time: topLevelTiming.start_time,
      end_time: topLevelTiming.end_time,
      ...durationFieldsForJSON(totalsSection),
    },
    scheduler_timing: schedulerTiming,
    ...(Object.keys(targetExtensions).length > 0
      ? { extensions: targetExtensions }
      : {}),
  };
  writeValidatedJson(
    path.join(summary.targetDir, "target-summary.json"),
    targetSummarySchemaID,
    targetSummary,
  );
  const targetToolSummaryFile = toolSummaryPath(summary.targetDir);
  let targetToolSummaryRel = relToRepo(targetToolSummaryFile);
  if (
    !preserveExistingToolSummary ||
    !existsSync(targetToolSummaryFile)
  ) {
    targetToolSummaryRel = writeToolSummary(
      targetToolSummaryFile,
      targetToolSummary(targetSummary, relToRepo(targetToolSummaryFile)),
    );
  } else {
    const existingToolSummary = JSON.parse(
      readFileSync(targetToolSummaryFile, "utf8"),
    );
    validateSchemaSync("cartulary.tool_run_summary.v5", existingToolSummary);
  }
  const shouldSuppressMachineOutput =
    suppressMachineOutput || suppressChildSuccess();
  if (
    !suppressChildSuccess() &&
    process.env.CARTULARY_DEFER_OBSERVABILITY_FINALIZE !== "1" &&
    observabilityRequiredTarget(targetSummary.target)
  ) {
    const observability = finalizeObservabilitySafely(path.join(resultsRoot, runId), {
      target: targetSummary.target,
      status: targetSummary.status === "pass" ? "passed" : "failed",
    });
    if (observability.status === "partial") {
      const toolSummary = JSON.parse(readFileSync(targetToolSummaryFile, "utf8"));
      toolSummary.warnings = [
        ...(toolSummary.warnings ?? []),
        {
          kind: "harness_observability",
          status: "partial",
          diagnostic: observability.diagnostic,
        },
      ];
      writeToolSummary(targetToolSummaryFile, toolSummary);
    }
  }

  if (finalStatus === "PASS") {
    if (machineOutput()) {
      if (!shouldSuppressMachineOutput) {
        process.stdout.write(
          compactJSONString(
            JSON.parse(readFileSync(targetToolSummaryFile, "utf8")),
          ),
        );
      }
      return 0;
    }
    if ((quietSuccess || suppressChildSuccess()) && quietOutputMode()) {
      return 0;
    }
    writeTargetResult(process.stdout, targetSummary, targetToolSummaryRel);
    if (verboseOutput()) {
      writeFixtureLine(process.stdout, targetSummary.totals.fixture);
      writeChildTargetLines(
        process.stdout,
        target,
        childTargets,
        unresolvedMissingChildTargetSummaries,
      );
      writeSkippedChildTargetLines(process.stdout, target, skippedChildTargets);
      printInventory(summary);
    }
    return 0;
  }

  if (machineOutput()) {
    if (!shouldSuppressMachineOutput) {
      process.stdout.write(
        compactJSONString(
          JSON.parse(readFileSync(targetToolSummaryFile, "utf8")),
        ),
      );
    }
    return 0;
  }
  if ((quietFailure || suppressChildSuccess()) && quietOutputMode()) {
    return 0;
  }
  writeTargetFailure(process.stderr, targetSummary, targetToolSummaryRel);
  if (verboseOutput()) {
    writeTargetLine(process.stderr, "[FAIL]", targetSummary);
    writeFailureHeadline(process.stderr, target, targetSummary);
    writeFixtureLine(process.stderr, targetSummary.totals.fixture);
    writeChildTargetLines(
      process.stderr,
      target,
      childTargets,
      unresolvedMissingChildTargetSummaries,
    );
    writeSkippedChildTargetLines(process.stderr, target, skippedChildTargets);
  }
  return testAccountingFailures.length > 0 ? 1 : 0;
}
