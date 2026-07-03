#!/usr/bin/env node

import {
  existsSync,
  readFileSync,
} from "node:fs";
import path from "node:path";
import { helperArtifactReferences } from "../artifact-discovery.mjs";
import {
  classifyExecutionFailure,
  failureFieldsForJSON,
  publicExitCodeForSummary,
} from "../failure-taxonomy.mjs";
import { combineFixtureSummaries } from "../fixture-reporting.mjs";
import {
  compactJSONString,
  prettyJSONString,
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../harness-contract.mjs";
import {
  artifactLine,
  artifactRef,
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
  repoRoot,
  resolveResultsRoot,
  resolveRunId,
  runSummarySchemaID,
  testCoverageBuckets,
} from "./context.mjs";
import { buildSharedExecutionGroups } from "./shared-execution.mjs";
import {
  addDurationFields,
  createDurationFields,
  durationFieldsForJSON,
  teardownStatus,
} from "./timing.mjs";
import {
  loadTargetSummary,
  parseTargetList,
  targetSummaryAccountingView,
  targetSummaryPath,
  writeFailureHeadline,
  writeFixtureLine,
  writeToolSummary,
} from "./target-summary.mjs";

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

function resolveAccountingModes(accountingModes, fallbackActualPhases = 0) {
  const modes = createAccountingModes();
  if (!accountingModes) {
    modes.actual = clampDurationMs(fallbackActualPhases);
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

function schedulerSummaryPath(target) {
  return path.join(resultsRoot, runId, target, "scheduler-summary.json");
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

function normalizeCounts(counts = {}) {
  const normalized = {
    phases: clampDurationMs(counts.phases ?? 0),
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

function parseSummaryGroupsSpec(value) {
  if (!value) {
    return [];
  }
  return value
    .split(";")
    .map((group) => group.trim())
    .filter((group) => group.length > 0)
    .map((group) => {
      const separator = group.indexOf("=");
      if (separator <= 0) {
        throw new Error(
          `invalid summary group ${group}; expected <name>=<target,target>`,
        );
      }
      const name = group.slice(0, separator).trim();
      const summaryTargets = parseTargetList(group.slice(separator + 1));
      if (summaryTargets.length === 0) {
        throw new Error(
          `invalid summary group ${name}; expected at least one target`,
        );
      }
      return { name, summaryTargets };
    });
}

function parseRunSummaryArgs(args) {
  const [
    label,
    requestedStatus = "pass",
    completedText = "0",
    totalText = "0",
    abortedAfter = "",
    ...remaining
  ] = args;
  if (!label) {
    throw new Error(
      "usage: test-output.mjs run-summary <label> <pass|fail> <completed> <total> <aborted_after|-> [--summary-groups <name=a,b;name=c>] [--skipped-after-failure <target,target>] [--helper-units <unit,unit>] [--completed-helper-units <unit,unit>] [--quiet-success] [--quiet-failure] [--suppress-machine-output] [summary_targets...]",
    );
  }
  const summaryTargets = [];
  const summaryGroups = [];
  const skippedAfterFailure = [];
  let helperUnits = [];
  let completedHelperUnits = [];
  let quietSuccess = false;
  let quietFailure = false;
  let suppressMachineOutput = false;
  while (remaining.length > 0) {
    const value = remaining.shift();
    if (value === "--quiet-success") {
      quietSuccess = true;
      continue;
    }
    if (value === "--quiet-failure") {
      quietFailure = true;
      continue;
    }
    if (value === "--suppress-machine-output") {
      suppressMachineOutput = true;
      continue;
    }
    if (value === "--summary-groups") {
      const spec = remaining.shift();
      if (spec === undefined) {
        throw new Error("--summary-groups requires <name=a,b;name=c>");
      }
      summaryGroups.push(...parseSummaryGroupsSpec(spec));
      continue;
    }
    if (value === "--skipped-after-failure") {
      const spec = remaining.shift();
      if (spec === undefined) {
        throw new Error("--skipped-after-failure requires <target,target>");
      }
      skippedAfterFailure.push(...parseTargetList(spec));
      continue;
    }
    if (value === "--helper-units") {
      const spec = remaining.shift();
      if (spec === undefined) {
        throw new Error("--helper-units requires <unit,unit>");
      }
      helperUnits = parseTargetList(spec);
      continue;
    }
    if (value === "--completed-helper-units") {
      const spec = remaining.shift();
      if (spec === undefined) {
        throw new Error("--completed-helper-units requires <unit,unit>");
      }
      completedHelperUnits = parseTargetList(spec);
      continue;
    }
    if (value.startsWith("--")) {
      throw new Error(`unknown run-summary option ${value}`);
    }
    summaryTargets.push(value);
  }
  return {
    label,
    requestedStatus,
    completedText,
    totalText,
    abortedAfter,
    summaryTargets,
    summaryGroups,
    skippedAfterFailure,
    helperUnits,
    completedHelperUnits,
    quietSuccess,
    quietFailure,
    suppressMachineOutput,
  };
}

function createDurationAggregate() {
  return {
    phases: 0,
    ...createCounts(),
    ...createDurationFields(),
  };
}

function addSummaryToAggregate(aggregate, accountingModes, summary) {
  const view = targetSummaryAccountingView(summary);
  addCounts(aggregate, view.counts);
  addDurationFields(aggregate, view);
  mergeAccountingModes(
    accountingModes,
    resolveAccountingModes(view.accounting_modes, view.counts?.phases ?? 0),
  );
}

function countsForJSON(aggregate) {
  const counts = {
    phases: aggregate.phases,
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

function summarizeTargetSummaries(
  summaries,
  missingTargetSummaries,
  requestedStatus = "pass",
  failureContext = {},
) {
  const aggregate = createDurationAggregate();
  const accountingModes = createAccountingModes();
  let failed = requestedStatus === "fail" || missingTargetSummaries.length > 0;
  let startTime = "";
  let endTime = "";
  const timingFailures = [];
  const teardownFailures = [];
  const failures = [];

  for (const summary of summaries) {
    const view = targetSummaryAccountingView(summary);
    addSummaryToAggregate(aggregate, accountingModes, summary);
    timingFailures.push(...(view.timing_failures ?? []));
    teardownFailures.push(...(view.teardown_failures ?? []));
    failures.push(...(view.failures ?? []));
    if (startTime === "" || (view.start_time && view.start_time < startTime)) {
      startTime = view.start_time ?? "";
    }
    if (endTime === "" || (view.end_time && view.end_time > endTime)) {
      endTime = view.end_time ?? "";
    }
    if (summary.status !== "pass") {
      failed = true;
    }
  }

  const abortedAfter = failureContext.abortedAfter ?? "";
  const recordsMissingTargetFailures =
    missingTargetSummaries.length > 0 &&
    !(requestedStatus === "fail" && abortedAfter && abortedAfter !== "-");
  if (
    requestedStatus === "fail" &&
    aggregate.failed === 0 &&
    recordsMissingTargetFailures
  ) {
    aggregate.phases += 1;
    aggregate.failed += 1;
    aggregate.non_test += 1;
    aggregate.non_test_failed += 1;
  }

  if (
    requestedStatus === "fail" &&
    failures.length === 0 &&
    !recordsMissingTargetFailures
  ) {
    if (aggregate.failed === 0) {
      aggregate.phases += 1;
      aggregate.failed += 1;
      aggregate.non_test += 1;
      aggregate.non_test_failed += 1;
    }
    const failureTarget =
      failureContext.abortedAfter && failureContext.abortedAfter !== "-"
        ? failureContext.abortedAfter
        : (failureContext.target ?? "");
    failures.push({
      failure_class: classifyExecutionFailure(
        failureTarget,
        failureContext.label ?? "",
        failureContext.command ?? "",
      ),
      kind: "failure",
      source: "run-summary",
      target: failureContext.target ?? "",
      label: failureTarget,
      message: failureTarget
        ? `${failureTarget} failed before a test failure was attributed`
        : "run failed before a test failure was attributed",
    });
  }
  for (const missingTarget of missingTargetSummaries) {
    if (requestedStatus === "fail" && abortedAfter && abortedAfter !== "-") {
      continue;
    }
    const abortedClass =
      abortedAfter && abortedAfter !== "-"
        ? classifyExecutionFailure(abortedAfter, failureContext.label ?? "")
        : "";
    const missingClass =
      abortedClass ||
      (classifyExecutionFailure(missingTarget) === "timing"
        ? "timing"
        : "artifact");
    failures.push({
      failure_class: missingClass,
      kind: missingClass === "timing" ? "timing" : "artifact",
      source: "run-summary",
      target: failureContext.target ?? "",
      label: missingTarget,
      message: `missing target summary: ${missingTarget}`,
      artifact: relToRepo(targetSummaryPath(missingTarget)),
    });
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
    aggregate,
    accountingModes,
    failed,
    startTime,
    endTime,
    wallDurationMs,
    criticalPathWallDurationMs,
    timingFailures,
    teardownFailures,
    ...failureFields,
  };
}

function buildSummaryGroups(summaryGroups, skippedAfterFailureSet = new Set()) {
  return summaryGroups.map((group) => {
    const groupSummaries = [];
    const missingSummaryTargets = [];
    const skippedAfterFailure = [];
    for (const summaryTarget of group.summaryTargets) {
      const summary = loadTargetSummary(summaryTarget);
      if (!summary) {
        if (skippedAfterFailureSet.has(summaryTarget)) {
          skippedAfterFailure.push(summaryTarget);
          continue;
        }
        missingSummaryTargets.push(summaryTarget);
        continue;
      }
      groupSummaries.push(summary);
    }
    const summarized = summarizeTargetSummaries(
      groupSummaries,
      missingSummaryTargets,
    );
    return {
      name: group.name,
      status: summarized.failed ? "fail" : "pass",
      summary_targets: group.summaryTargets,
      missing_summary_targets: missingSummaryTargets,
      skipped_after_failure: skippedAfterFailure,
      start_time: summarized.startTime,
      end_time: summarized.endTime,
      ...durationFieldsForJSON(summarized.aggregate, {
        wall_duration_ms: summarized.wallDurationMs,
        critical_path_wall_duration_ms: summarized.criticalPathWallDurationMs,
      }),
      accounting_modes: summarized.accountingModes,
      counts: countsForJSON(summarized.aggregate),
      failure_class: summarized.failure_class,
      failure_classes: summarized.failure_classes,
      failures: summarized.failures,
      failure_headline: summarized.failure_headline,
      timing_failures: summarized.timingFailures,
      teardown_status: teardownStatus(
        summarized.aggregate.teardown_duration_ms,
        summarized.teardownFailures,
      ),
      teardown_failures: summarized.teardownFailures,
    };
  });
}

function writeSummaryGroupLines(stream, label, summaryGroups) {
  for (const group of summaryGroups) {
    const missing =
      group.missing_summary_targets.length > 0
        ? ` missing_summary_targets=${group.missing_summary_targets.join(",")}`
        : "";
    const skipped =
      group.skipped_after_failure.length > 0
        ? ` skipped_after_failure=${group.skipped_after_failure.join(",")}`
        : "";
    const failureClass = group.failure_class
      ? ` failure_class=${group.failure_class}`
      : "";
    stream.write(
      `[GROUP] ${label} ${group.name} summary_targets=${group.summary_targets.join(",")} status=${group.status}${failureClass} ${formatDurationFields(group.wall_duration_ms, group.executed_duration_ms, group.logical_duration_ms, group.critical_path_wall_duration_ms, group.teardown_duration_ms)} ${formatAccountingModeFields(group.accounting_modes)}${missing}${skipped}\n`,
    );
  }
}

function writeSharedExecutionGroupLines(stream, label, sharedExecutionGroups) {
  for (const group of sharedExecutionGroups) {
    const failureClass = group.failure_class
      ? ` failure_class=${group.failure_class}`
      : "";
    stream.write(
      `[SHARED] ${label} ${group.name} status=${group.status}${failureClass} wall=${formatDuration(group.wall_duration_ms)} exec=${formatDuration(group.executed_duration_ms)} reports=${group.reports}\n`,
    );
  }
}

function runToolSummary(runSummary, summaryJsonPath) {
  const slowest = slowestTargetRef(runSummary);
  const schedulerTiming = runSummary.scheduler_timing ?? null;
  const targetDir = runSummaryTargetDir(runSummary);
  return buildToolRunSummary({
    target: runSummary.label,
    command: ["make", runSummary.label],
    status: runSummary.status,
    exitCode: runSummary.status === "pass" ? 0 : 1,
    startedAt: schedulerTiming?.scheduler_started_at ?? runSummary.start_time,
    completedAt: schedulerTiming?.scheduler_completed_at ?? runSummary.end_time,
    durationMs:
      schedulerTiming?.scheduler_total_duration_ms ??
      runSummary.wall_duration_ms,
    outputMode: resolveOutputMode(),
    resultRoot: relToRepo(resultsRoot),
    runId,
    runRoot: runSummary.artifacts?.dir ?? "",
    summaryArtifacts: [
      artifactRef("tool_run_summary", summaryJsonPath),
      artifactRef(
        "run_summary",
        path.join(runSummary.artifacts?.dir ?? "", "run-summary.json"),
      ),
      ...runTargetSummaryArtifacts(targetDir),
    ],
    logArtifacts: runTargetLogArtifacts(targetDir),
    workUnits: [
      {
        id: runSummary.label,
        completed: runSummary.work_units?.completed ?? 0,
        total: runSummary.work_units?.total ?? 0,
        aborted_after: runSummary.work_units?.aborted_after ?? "",
        status: runSummary.status,
      },
    ],
    evidenceTargets: (runSummary.evidence_targets?.present ?? []).map(
      (target) => ({ target }),
    ),
    helperUnits: (runSummary.helper_units?.names ?? []).map((name) => ({
      target: name,
    })),
    counts: runSummary.counts,
    phaseAccounting: {
      missing: runSummary.summary_targets?.missing?.length ?? 0,
    },
    failureClass: runSummary.failure_class,
    failureReason: runSummary.failure_reason,
    failures: runSummary.failures ?? [],
    slowest: slowest ? [slowest] : [],
    schedulerTiming,
    rerunCommands: [`make ${runSummary.label}`],
  });
}

function runSummaryTargetDir(runSummary) {
  const runDir = runSummary.artifacts?.dir ?? "";
  const target = runSummary.label ?? "";
  if (!runDir || !target) {
    return "";
  }
  const targetDir = path.join(runDir, target);
  return existsSync(targetDir) ? targetDir : "";
}

function runTargetSummaryArtifacts(targetDir) {
  if (!targetDir) {
    return [];
  }
  return [
    artifactRef("target_summary", path.join(targetDir, "target-summary.json")),
    artifactRef("target_timing", path.join(targetDir, "target-timing.json")),
    existsSync(path.join(targetDir, "scheduler-summary.json"))
      ? artifactRef(
          "scheduler_summary",
          path.join(targetDir, "scheduler-summary.json"),
        )
      : null,
    existsSync(path.join(targetDir, "scheduler-events.jsonl"))
      ? artifactRef(
          "scheduler_events",
          path.join(targetDir, "scheduler-events.jsonl"),
          "jsonl",
        )
      : null,
  ];
}

function runTargetLogArtifacts(targetDir) {
  if (!targetDir) {
    return [];
  }
  return [
    existsSync(path.join(targetDir, "progress-summary.log"))
      ? artifactRef(
          "scheduler_progress",
          path.join(targetDir, "progress-summary.log"),
          "log",
        )
      : null,
    existsSync(path.join(targetDir, "scheduler-logs"))
      ? artifactRef(
          "scheduler_logs",
          path.join(targetDir, "scheduler-logs"),
          "directory",
        )
      : null,
  ];
}

function writeRunResult(stream, runSummary, summaryJsonPath) {
  const summary = runToolSummary(runSummary, summaryJsonPath);
  stream.write(resultLine(summary, summaryJsonPath));
  stream.write(
    artifactLine(summary, summaryJsonPath, {
      investigate: `make explain-run RESULTS_DIR=${summary.run_root}`,
    }),
  );
}

function writeRunFailure(stream, runSummary, summaryJsonPath) {
  const summary = runToolSummary(runSummary, summaryJsonPath);
  const failureClass = summary.failure_class ?? "harness";
  const failureReason = summary.failure_reason ?? "unknown_failure";
  const headline = runSummary.failure_headline ?? "run failed";
  const abortedAfter = runSummary.work_units?.aborted_after || "";
  const failedTarget =
    abortedAfter ||
    runSummary.summary_targets?.missing?.[0] ||
    runSummary.failures?.[0]?.target ||
    "";
  const logArtifact = firstArtifactPath(
    runSummary.failures?.find((failure) => failure.artifact)?.artifact ??
      runSummary.failures?.find((failure) => failure.raw)?.raw,
  );
  stream.write(
    `[FAIL] target=${summary.target} exit_code=${summary.exit_code} failure_class=${failureClass} reason=${failureReason} work_unit=${abortedAfter || "-"} child_target=${failedTarget || "-"} duration_ms=${summary.duration_ms} headline="${headline}"\n`,
  );
  stream.write(
    `[ARTIFACTS] target=${summary.target} root=${summary.run_root} summary_json=${terminalArtifactPath(summary.run_root, summaryJsonPath)} log_artifact=${terminalArtifactPath(summary.run_root, logArtifact)} scheduler_json=- progress_log=-\n`,
  );
  stream.write(`[RERUN] command="make ${summary.target}"\n`);
  stream.write(
    `[INVESTIGATE] command="make explain-run RESULTS_DIR=${summary.run_root}"\n`,
  );
}

function findSlowestTarget(targetSummaries) {
  return targetSummaries.reduce((current, summary) => {
    const view = targetSummaryAccountingView(summary);
    const durationMs = clampDurationMs(
      view.critical_path_wall_duration_ms ??
        view.wall_duration_ms ??
        view.logical_duration_ms ??
        0,
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
    const bucket =
      targetSummaryAccountingView(summary).slowest_lifecycle_bucket;
    if (!bucket) {
      return current;
    }
    const candidate = {
      target: summary.target,
      name: bucket.name,
      duration_ms: clampDurationMs(bucket.duration_ms ?? 0),
    };
    if (!current || candidate.duration_ms > current.duration_ms) {
      return candidate;
    }
    return current;
  }, null);
}

export function handleRunSummary(args) {
  const {
    label,
    requestedStatus,
    completedText,
    totalText,
    abortedAfter,
    summaryTargets,
    summaryGroups,
    skippedAfterFailure,
    helperUnits,
    completedHelperUnits,
    quietSuccess,
    quietFailure,
    suppressMachineOutput,
  } = parseRunSummaryArgs(args);
  const completedWorkUnits = Number.parseInt(completedText, 10) || 0;
  const totalWorkUnits = Number.parseInt(totalText, 10) || 0;
  const skippedAfterFailureSet = new Set(skippedAfterFailure);
  const missingSummaryTargets = [];
  const evidenceTargetSummaries = [];

  for (const summaryTarget of summaryTargets) {
    const summary = loadTargetSummary(summaryTarget);
    if (!summary) {
      if (skippedAfterFailureSet.has(summaryTarget)) {
        continue;
      }
      missingSummaryTargets.push(summaryTarget);
      continue;
    }
    evidenceTargetSummaries.push(summary);
  }
  const summarized = summarizeTargetSummaries(
    evidenceTargetSummaries,
    missingSummaryTargets,
    requestedStatus,
    { target: label, label, abortedAfter },
  );
  const aggregate = summarized.aggregate;
  const accountingModes = summarized.accountingModes;
  const schedulerSummary = loadSchedulerSummary(label);
  const schedulerTiming = schedulerTimingFromSummary(schedulerSummary);
  const schedulerAccounting = schedulerAccountingFromSummary(schedulerSummary);
  if (schedulerAccounting) {
    addSchedulerAccountingDurations(aggregate, schedulerAccounting);
    mergeAccountingModes(accountingModes, schedulerAccounting.accounting_modes);
  }
  const wallDurationMs =
    schedulerTiming?.scheduler_total_duration_ms ?? summarized.wallDurationMs;
  const criticalPathWallDurationMs =
    schedulerTiming?.scheduler_total_duration_ms ??
    summarized.criticalPathWallDurationMs;
  const renderedSummaryGroups = buildSummaryGroups(
    summaryGroups,
    skippedAfterFailureSet,
  );
  const sharedExecutionGroups = buildSharedExecutionGroups();
  const runFailureFields = failureFieldsForJSON(
    [
      ...(summarized.failures ?? []),
      ...renderedSummaryGroups.flatMap((group) => group.failures ?? []),
      ...sharedExecutionGroups.flatMap((group) => group.failures ?? []),
    ],
    countsForJSON(aggregate),
  );
  const failed =
    summarized.failed ||
    renderedSummaryGroups.some((group) => group.status !== "pass") ||
    sharedExecutionGroups.some((group) => group.status !== "pass");
  const slowestTarget = findSlowestTarget(evidenceTargetSummaries);
  const slowestLifecycleBucket = findSlowestLifecycleBucket(
    evidenceTargetSummaries,
  );
  const runFixture = combineFixtureSummaries(
    label,
    null,
    evidenceTargetSummaries.map((summary) => ({
      fixture: targetSummaryAccountingView(summary, summary.target).fixture,
    })),
  );
  const workUnits = {
    completed: completedWorkUnits,
    total: totalWorkUnits,
    aborted_after: abortedAfter === "-" ? "" : abortedAfter,
  };
  const summaryTargetFields = {
    expected: summaryTargets,
    missing: missingSummaryTargets,
    skipped_after_failure: skippedAfterFailure,
  };
  const evidenceTargets = {
    present: evidenceTargetSummaries.map((summary) => summary.target),
    summaries: evidenceTargetSummaries,
  };
  const helperUnitFields = {
    total: helperUnits.length,
    completed: completedHelperUnits.length,
    names: helperUnits,
    artifacts: helperArtifactReferences(helperUnits, { root: repoRoot, runId }),
  };
  const runSummaryCollationTime = new Date().toISOString();

  const runSummary = {
    schema_id: runSummarySchemaID,
    label,
    status: failed ? "fail" : "pass",
    work_units: workUnits,
    start_time:
      schedulerTiming?.scheduler_started_at ||
      summarized.startTime ||
      runSummaryCollationTime,
    end_time:
      schedulerTiming?.scheduler_completed_at ||
      summarized.endTime ||
      runSummaryCollationTime,
    ...durationFieldsForJSON(aggregate, {
      wall_duration_ms: wallDurationMs,
      critical_path_wall_duration_ms: criticalPathWallDurationMs,
    }),
    accounting_modes: accountingModes,
    counts: countsForJSON(aggregate),
    ...runFailureFields,
    slowest_target: slowestTarget,
    slowest_lifecycle_bucket: slowestLifecycleBucket,
    timing_failures: summarized.timingFailures,
    teardown_status: teardownStatus(
      aggregate.teardown_duration_ms,
      summarized.teardownFailures,
    ),
    teardown_failures: summarized.teardownFailures,
    fixture: runFixture,
    artifacts: {
      dir: relToRepo(path.join(resultsRoot, runId)),
    },
    summary_targets: summaryTargetFields,
    evidence_targets: evidenceTargets,
    helper_units: helperUnitFields,
    summary_groups: renderedSummaryGroups,
    shared_execution_groups: sharedExecutionGroups,
    ...(schedulerTiming ? { scheduler_timing: schedulerTiming } : {}),
    ...(schedulerAccounting
      ? {
          extensions: {
            "cartulary.scheduler_accounting": schedulerAccounting,
          },
        }
      : {}),
  };
  writeValidatedJson(
    path.join(resultsRoot, runId, "run-summary.json"),
    runSummarySchemaID,
    runSummary,
  );
  const runToolSummaryFile = toolSummaryPath(path.join(resultsRoot, runId));
  const runToolSummaryRel = writeToolSummary(
    runToolSummaryFile,
    runToolSummary(runSummary, relToRepo(runToolSummaryFile)),
  );
  const shouldSuppressMachineOutput =
    suppressMachineOutput || suppressChildSuccess();

  if (!failed) {
    if (machineOutput()) {
      if (!shouldSuppressMachineOutput) {
        process.stdout.write(
          compactJSONString(runToolSummary(runSummary, runToolSummaryRel)),
        );
      }
      return 0;
    }
    if (quietSuccess && quietOutputMode()) {
      return 0;
    }
    writeRunResult(process.stdout, runSummary, runToolSummaryRel);
    if (verboseOutput()) {
      writeFixtureLine(process.stdout, runFixture);
      writeSummaryGroupLines(process.stdout, label, renderedSummaryGroups);
      writeSharedExecutionGroupLines(
        process.stdout,
        label,
        sharedExecutionGroups,
      );
    }
    return 0;
  }

  if (machineOutput()) {
    if (!shouldSuppressMachineOutput) {
      process.stdout.write(
        compactJSONString(runToolSummary(runSummary, runToolSummaryRel)),
      );
    }
    return publicExitCodeForSummary(
      runToolSummary(runSummary, runToolSummaryRel),
    );
  }
  if ((quietFailure || suppressChildSuccess()) && quietOutputMode()) {
    return publicExitCodeForSummary(
      runToolSummary(runSummary, runToolSummaryRel),
    );
  }
  writeRunFailure(process.stderr, runSummary, runToolSummaryRel);
  if (verboseOutput()) {
    writeFailureHeadline(process.stderr, label, runSummary);
    writeFixtureLine(process.stderr, runFixture);
    writeSummaryGroupLines(process.stderr, label, renderedSummaryGroups);
    writeSharedExecutionGroupLines(
      process.stderr,
      label,
      sharedExecutionGroups,
    );
  }
  return publicExitCodeForSummary(
    runToolSummary(runSummary, runToolSummaryRel),
  );
}
