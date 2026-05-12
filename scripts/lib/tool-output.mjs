import path from "node:path";

import {
  redactValue,
  resolveOutputMode as resolveHarnessOutputMode,
} from "./harness-contract.mjs";
import {
  defaultReasonForFailureClass,
  failureClassOrder,
  failureReasonOrder,
  normalizeFailureClass,
  normalizeFailureReason,
  publicExitCodeForFailures,
} from "./failure-taxonomy.mjs";

export const toolRunSummarySchemaID = "cartulary.tool_run_summary.v2";

export const failureClasses = failureClassOrder;
export const failureReasons = failureReasonOrder;

export function normalizeOutputMode(env = process.env) {
  return resolveHarnessOutputMode(env, env.CARTULARY_TEST_TARGET || "");
}

export function quietLikeOutput(env = process.env) {
  const mode = normalizeOutputMode(env);
  return (
    mode === "quiet" ||
    mode === "summary" ||
    mode === "ci" ||
    mode === "machine"
  );
}

export function machineOutput(env = process.env) {
  return normalizeOutputMode(env) === "machine";
}

export function verboseOutput(env = process.env) {
  const mode = normalizeOutputMode(env);
  return mode === "verbose" || mode === "debug";
}

export function suppressChildSuccess(env = process.env) {
  return env.CARTULARY_SUPPRESS_CHILD_SUCCESS === "1";
}

export function artifactRef(role, file, kind = "json") {
  if (!file) {
    return null;
  }
  return { role, kind, path: file };
}

function normalizeArtifactPath(value) {
  return String(value ?? "").replaceAll("\\", "/");
}

export function terminalArtifactPath(runRoot, file) {
  if (!file) {
    return "-";
  }
  const root = normalizeArtifactPath(runRoot).replace(/\/+$/u, "");
  const target = normalizeArtifactPath(file);
  if (root && target === `${root}/tool-run-summary.json`) {
    return "tool-run-summary.json";
  }
  if (root && target.startsWith(`${root}/`)) {
    return target.slice(root.length + 1);
  }
  return target;
}

export function compactDurationMs(value) {
  if (!Number.isFinite(value) || value < 0) {
    return 0;
  }
  return Math.round(value);
}

function normalizeTimestamp(value, fallback = null) {
  const candidate = typeof value === "string" ? value.trim() : "";
  if (candidate && !Number.isNaN(Date.parse(candidate))) {
    return candidate;
  }
  if (fallback) {
    return normalizeTimestamp(fallback);
  }
  return new Date().toISOString();
}

export function slowestTargetRef(summary) {
  if (!summary) {
    return null;
  }
  if (summary.slowest_target?.target) {
    return {
      id: summary.slowest_target.target,
      duration_ms: compactDurationMs(
        summary.slowest_target.critical_path_wall_duration_ms,
      ),
      kind: "target",
    };
  }
  const totals = summary.totals ?? summary;
  if (totals.slowest_lifecycle_bucket?.name) {
    return {
      id: totals.slowest_lifecycle_bucket.target
        ? `${totals.slowest_lifecycle_bucket.target}:${totals.slowest_lifecycle_bucket.name}`
        : totals.slowest_lifecycle_bucket.name,
      duration_ms: compactDurationMs(
        totals.slowest_lifecycle_bucket.duration_ms,
      ),
      kind: "lifecycle_bucket",
    };
  }
  return null;
}

export function slowestText(entry) {
  if (!entry) {
    return "none";
  }
  return `${entry.id}:${entry.duration_ms}`;
}

export function phaseAccountingFromCounts(counts = {}) {
  return {
    authoritative: counts.authoritative ?? 0,
    support: counts.support ?? 0,
    raw: counts.raw ?? 0,
    tooling_support: counts.tooling_support ?? 0,
    unowned_regression: counts.unowned_regression ?? 0,
    unmapped: counts.unmapped ?? 0,
    authoritative_failed: counts.authoritative_failed ?? 0,
    support_failed: counts.support_failed ?? 0,
    raw_failed: counts.raw_failed ?? 0,
    tooling_support_failed: counts.tooling_support_failed ?? 0,
    unowned_regression_failed: counts.unowned_regression_failed ?? 0,
    unmapped_failed: counts.unmapped_failed ?? 0,
    missing: counts.missing ?? 0,
  };
}

export function countsForToolSummary(counts = {}) {
  return {
    phases: counts.phases ?? 0,
    tests: counts.tests ?? 0,
    failed: counts.failed ?? 0,
    non_test: counts.non_test ?? 0,
    non_test_failed: counts.non_test_failed ?? 0,
    packages: counts.packages ?? 0,
  };
}

function hasApplicableCount(counts = {}, key) {
  return (
    Object.hasOwn(counts, key) &&
    counts[key] !== null &&
    counts[key] !== undefined
  );
}

export function commandInfo({
  target,
  command = null,
  cwd = process.cwd(),
  env = process.env,
} = {}) {
  const argv = Array.isArray(command)
    ? command.map((entry) => String(entry))
    : command
      ? [String(command)]
      : process.argv.slice(1);
  return {
    cwd,
    argv: redactValue(argv),
    make_target: target ?? null,
    env: redactValue({
      CARTULARY_OUTPUT_MODE: env.CARTULARY_OUTPUT_MODE ?? null,
      VERBOSE: env.VERBOSE ?? null,
      CI_VERBOSE: env.CI_VERBOSE ?? null,
      CARTULARY_TEST_RESULTS_DIR: env.CARTULARY_TEST_RESULTS_DIR ?? null,
      CARTULARY_TEST_RUN_ID: env.CARTULARY_TEST_RUN_ID ?? null,
    }),
  };
}

function artifactSortKey(artifact) {
  return `${artifact.role}\0${artifact.kind}\0${artifact.path}`;
}

function sortArtifacts(artifacts = []) {
  return artifacts
    .filter(Boolean)
    .slice()
    .sort((left, right) =>
      artifactSortKey(left).localeCompare(artifactSortKey(right)),
    );
}

function sortWorkUnits(workUnits = []) {
  return workUnits
    .map((unit) => {
      const normalized = { ...unit };
      if (!normalized.aborted_after) {
        delete normalized.aborted_after;
      }
      return normalized;
    })
    .slice()
    .sort((left, right) =>
      String(left.id ?? "").localeCompare(String(right.id ?? "")),
    );
}

function sortTargetRefs(targets = []) {
  return targets
    .slice()
    .sort((left, right) =>
      String(left.target ?? "").localeCompare(String(right.target ?? "")),
    );
}

function failureSortKey(failure) {
  return [
    failure.failure_class ?? "",
    failure.failure_reason ?? "",
    failure.target ?? "",
    failure.work_unit ?? "",
    failure.child_target ?? "",
    failure.label ?? "",
    failure.kind ?? "",
    failure.message ?? failure.headline ?? "",
  ].join("\0");
}

function sortFailures(failures = []) {
  return failures
    .slice()
    .sort((left, right) =>
      failureSortKey(left).localeCompare(failureSortKey(right)),
    );
}

function sortSlowest(slowest = []) {
  return slowest
    .slice()
    .sort(
      (left, right) =>
        (right.duration_ms ?? 0) - (left.duration_ms ?? 0) ||
        String(left.id ?? "").localeCompare(String(right.id ?? "")),
    );
}

export function buildToolRunSummary({
  target,
  command = null,
  status,
  exitCode = status === "pass" ? 0 : 1,
  startedAt,
  completedAt,
  durationMs,
  outputMode = normalizeOutputMode(),
  resultRoot = "",
  runId = "",
  runRoot,
  summaryArtifacts = [],
  logArtifacts = [],
  workUnits = [],
  evidenceTargets = [],
  helperUnits = [],
  counts = {},
  phaseAccounting = {},
  failureClass = null,
  failureReason = null,
  failures = [],
  slowest = [],
  warnings = [],
  rerunCommands = [],
  extensions = {},
}) {
  const normalizedFailureClass = normalizeFailureClass(failureClass, "");
  const normalizedFailureReason = normalizedFailureClass
    ? normalizeFailureReason(
        failureReason ??
          failures.find((failure) => failure?.failure_reason)?.failure_reason,
        defaultReasonForFailureClass(normalizedFailureClass),
      )
    : null;
  const normalizedFailures = sortFailures(failures);
  const normalizedExitCode =
    status === "pass"
      ? 0
      : publicExitCodeForFailures(failures, {
          failure_class: normalizedFailureClass || "unknown",
          failure_reason: normalizedFailureReason || "unknown_failure",
        });
  const normalizedCompletedAt = normalizeTimestamp(completedAt, startedAt);
  const normalizedStartedAt = normalizeTimestamp(
    startedAt,
    normalizedCompletedAt,
  );
  return {
    schema_id: toolRunSummarySchemaID,
    target,
    command: commandInfo({ target, command }),
    status,
    exit_code: normalizedExitCode || exitCode,
    started_at: normalizedStartedAt,
    completed_at: normalizedCompletedAt,
    duration_ms: compactDurationMs(durationMs),
    output_mode: outputMode,
    result_root: resultRoot,
    run_id: runId,
    run_root: runRoot,
    summary_artifacts: sortArtifacts(summaryArtifacts),
    log_artifacts: sortArtifacts(logArtifacts),
    work_units: sortWorkUnits(workUnits),
    evidence_targets: sortTargetRefs(evidenceTargets),
    helper_units: sortTargetRefs(helperUnits),
    counts: countsForToolSummary(counts),
    phase_accounting: {
      ...phaseAccountingFromCounts(counts),
      ...phaseAccounting,
    },
    failure_class: normalizedFailureClass,
    failure_reason: normalizedFailureReason,
    failures: normalizedFailures,
    slowest: sortSlowest(slowest),
    warnings,
    rerun_commands: rerunCommands,
    extensions,
  };
}

export function toolSummaryPath(runRoot) {
  return path.join(runRoot, "tool-run-summary.json");
}

export function resultLine(summary, summaryJsonPath) {
  const counts = summary.counts ?? {};
  const phase = summary.phase_accounting ?? {};
  const workUnits =
    summary.work_units?.[0]?.total !== undefined
      ? `${summary.work_units[0].completed}/${summary.work_units[0].total}`
      : "-";
  const tests =
    hasApplicableCount(counts, "tests") && counts.tests > 0
      ? counts.tests
      : "-";
  const phaseApplicable = [
    "authoritative",
    "support",
    "raw",
    "tooling_support",
    "unowned_regression",
    "unmapped",
    "missing",
  ].some((key) => Number(phase[key] ?? 0) > 0);
  const missing = phaseApplicable ? (phase.missing ?? 0) : "-";
  const unmapped = phaseApplicable && phase.unmapped > 0 ? phase.unmapped : "-";
  return `[RESULT] target=${summary.target} status=${summary.status} duration_ms=${summary.duration_ms} work_units=${workUnits} tests=${tests} failed=${counts.failed ?? 0} missing=${missing} unmapped=${unmapped} slowest=${slowestText(summary.slowest?.[0])} run_root=${summary.run_root} summary_json=${terminalArtifactPath(summary.run_root, summaryJsonPath)}\n`;
}

export function artifactLine(
  summary,
  summaryJsonPath,
  { investigate = null, log = null, extraFields = [] } = {},
) {
  const fields = [
    `[ARTIFACTS] target=${summary.target}`,
    `root=${summary.run_root}`,
    `summary_json=${terminalArtifactPath(summary.run_root, summaryJsonPath)}`,
    `logs=${terminalArtifactPath(summary.run_root, log)}`,
    ...extraFields,
  ];
  if (investigate) {
    fields.push(`investigate="${investigate}"`);
  }
  return `${fields.join(" ")}\n`;
}
