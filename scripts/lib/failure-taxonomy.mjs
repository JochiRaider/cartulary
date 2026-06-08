export const failureClassOrder = [
  "product",
  "security",
  "config",
  "infra",
  "harness",
  "artifact",
  "timing",
  "interrupted",
  "unknown",
];

export const failureReasonOrder = [
  "usage_error",
  "configuration_error",
  "preflight_error",
  "service_start_error",
  "service_readiness_timeout",
  "fixture_error",
  "resource_conflict",
  "test_assertion_failure",
  "security_finding",
  "child_target_failure",
  "tool_diagnostic_failure",
  "scheduler_accounting_error",
  "frontend_row_accounting",
  "artifact_error",
  "cleanup_error",
  "duration_baseline_drift",
  "timeout_failure",
  "cancelled_or_interrupted",
  "unknown_failure",
];

export const commandLifecycleOrder = [
  "wrapper_identity",
  "output_mode_resolution",
  "configuration_resolution",
  "result_root_run_id_resolution",
  "redaction_initialization",
  "semantic_target_behavior",
  "artifact_validation",
  "cleanup_finalizers",
  "public_output_emission",
];

const failureClassSet = new Set(failureClassOrder);
const failureReasonSet = new Set(failureReasonOrder);
const commandLifecycleRank = new Map(
  commandLifecycleOrder.map((step, index) => [step, index]),
);

const legacyFailureClassMap = new Map([
  ["test", "product"],
  ["helper", "harness"],
  ["infrastructure", "infra"],
]);

const reasonClassMap = new Map([
  ["usage_error", "config"],
  ["configuration_error", "config"],
  ["preflight_error", "infra"],
  ["service_start_error", "infra"],
  ["service_readiness_timeout", "infra"],
  ["fixture_error", "harness"],
  ["resource_conflict", "infra"],
  ["test_assertion_failure", "product"],
  ["security_finding", "security"],
  ["child_target_failure", "harness"],
  ["tool_diagnostic_failure", "harness"],
  ["scheduler_accounting_error", "harness"],
  ["frontend_row_accounting", "harness"],
  ["artifact_error", "artifact"],
  ["cleanup_error", "harness"],
  ["duration_baseline_drift", "timing"],
  ["timeout_failure", "timing"],
  ["cancelled_or_interrupted", "interrupted"],
  ["unknown_failure", "unknown"],
]);

const classDefaultReasonMap = new Map([
  ["product", "test_assertion_failure"],
  ["security", "security_finding"],
  ["config", "configuration_error"],
  ["infra", "preflight_error"],
  ["harness", "unknown_failure"],
  ["artifact", "artifact_error"],
  ["timing", "timeout_failure"],
  ["interrupted", "cancelled_or_interrupted"],
  ["unknown", "unknown_failure"],
]);

const reasonPublicExitCodeMap = new Map([
  ["usage_error", 2],
  ["configuration_error", 2],
  ["preflight_error", 3],
  ["service_start_error", 3],
  ["service_readiness_timeout", 3],
  ["fixture_error", 3],
  ["resource_conflict", 4],
  ["test_assertion_failure", 10],
  ["security_finding", 1],
  ["tool_diagnostic_failure", 1],
  ["scheduler_accounting_error", 11],
  ["frontend_row_accounting", 11],
  ["artifact_error", 11],
  ["cleanup_error", 12],
  ["duration_baseline_drift", 13],
  ["timeout_failure", 13],
  ["unknown_failure", 1],
]);

export function createFailureClassCounts() {
  return Object.fromEntries(failureClassOrder.map((failureClass) => [failureClass, 0]));
}

export function createFailureReasonCounts() {
  return Object.fromEntries(failureReasonOrder.map((reason) => [reason, 0]));
}

export function normalizeFailureClass(value, fallback = "helper") {
  const normalized = String(value ?? "").trim().toLowerCase();
  if (failureClassSet.has(normalized)) {
    return normalized;
  }
  if (legacyFailureClassMap.has(normalized)) {
    return legacyFailureClassMap.get(normalized);
  }
  if (fallback === "" || fallback === null) {
    return null;
  }
  const normalizedFallback = String(fallback).trim().toLowerCase();
  if (failureClassSet.has(normalizedFallback)) {
    return normalizedFallback;
  }
  return legacyFailureClassMap.get(normalizedFallback) ?? "harness";
}

export function normalizeFailureReason(value, fallback = "unknown_failure") {
  const normalized = String(value ?? "").trim().toLowerCase();
  if (failureReasonSet.has(normalized)) {
    return normalized;
  }
  if (fallback === "" || fallback === null) {
    return null;
  }
  return failureReasonSet.has(fallback) ? fallback : "unknown_failure";
}

export function failureClassForReason(reason, fallback = "unknown") {
  return reasonClassMap.get(normalizeFailureReason(reason)) ?? normalizeFailureClass(fallback, "unknown");
}

export function defaultReasonForFailureClass(failureClass) {
  return classDefaultReasonMap.get(normalizeFailureClass(failureClass, "unknown")) ?? "unknown_failure";
}

export function primaryFailureClass(failureClasses = createFailureClassCounts()) {
  for (const failureClass of failureClassOrder) {
    if ((failureClasses[failureClass] ?? 0) > 0) {
      return failureClass;
    }
  }
  return null;
}

export function normalizeFailureRecord(record = {}, defaults = {}) {
  const explicitReason = record.failure_reason ?? record.reason ?? defaults.failure_reason;
  const failureReason = normalizeFailureReason(explicitReason, "");
  const failureClass = normalizeFailureClass(
    record.failure_class ?? defaults.failure_class ?? failureClassForReason(failureReason, "unknown"),
  );
  const normalizedReason = failureReason ?? defaultReasonForFailureClass(failureClass);
  const normalized = {
    failure_class: failureClass,
    failure_reason: normalizedReason,
    kind: String(record.kind ?? defaults.kind ?? "failure"),
    source: String(record.source ?? defaults.source ?? ""),
    target: String(record.target ?? defaults.target ?? ""),
    phase: String(record.phase ?? defaults.phase ?? ""),
    runner: String(record.runner ?? defaults.runner ?? ""),
    label: String(record.label ?? defaults.label ?? ""),
    message: String(record.message ?? defaults.message ?? ""),
    artifact: String(record.artifact ?? defaults.artifact ?? ""),
  };
  const lifecycleStep = String(
    record.lifecycle_step ??
      record.lifecycle ??
      record.lifecycleStep ??
      defaults.lifecycle_step ??
      "",
  ).trim();
  if (lifecycleStep) {
    normalized.lifecycle_step = lifecycleStep;
  }
  for (const [field, aliases] of [
    ["scheduler_event_sequence", ["scheduler_event_sequence", "scheduler_seq", "event_sequence", "seq"]],
    ["child_registry_order", ["child_registry_order", "child_target_order", "target_registry_order"]],
  ]) {
    for (const alias of aliases) {
      const value = record[alias] ?? defaults[alias];
      if (Number.isInteger(value) && value >= 0) {
        normalized[field] = value;
        break;
      }
    }
  }
  const childTarget = String(record.child_target ?? defaults.child_target ?? "");
  const workUnit = String(record.work_unit ?? defaults.work_unit ?? "");
  if (childTarget) {
    normalized.child_target = childTarget;
  }
  if (workUnit) {
    normalized.work_unit = workUnit;
  }
  return normalized;
}

function classRank(failure) {
  const rank = failureClassOrder.indexOf(failure.failure_class);
  return rank === -1 ? Number.MAX_SAFE_INTEGER : rank;
}

function lifecycleRank(failure) {
  if (!failure.lifecycle_step) {
    return Number.MAX_SAFE_INTEGER;
  }
  return commandLifecycleRank.get(failure.lifecycle_step) ?? Number.MAX_SAFE_INTEGER;
}

function numericTie(value) {
  return Number.isInteger(value) ? value : Number.MAX_SAFE_INTEGER;
}

function compareFailureRecords(left, right) {
  return classRank(left) - classRank(right) ||
    lifecycleRank(left) - lifecycleRank(right) ||
    numericTie(left.scheduler_event_sequence) - numericTie(right.scheduler_event_sequence) ||
    numericTie(left.child_registry_order) - numericTie(right.child_registry_order) ||
    String(left.artifact ?? "").localeCompare(String(right.artifact ?? "")) ||
    String(left.failure_reason ?? "").localeCompare(String(right.failure_reason ?? "")) ||
    numericTie(left.__input_order) - numericTie(right.__input_order);
}

function publicExitCodeForInterrupted(context = {}) {
  const signal = String(context.signal ?? context.termsig ?? context.term_signal ?? "").trim().toUpperCase();
  if (signal === "SIGINT") {
    return 130;
  }
  if (signal === "SIGTERM") {
    return 143;
  }
  const status = Number(context.status ?? context.exit_status ?? context.exitCode ?? context.exit_code);
  if (status === 130 || status === 143) {
    return status;
  }
  return 15;
}

function childSummaryCandidates(record = {}, context = {}) {
  const candidates = context.childSummaries ?? context.child_summaries ?? [];
  if (!Array.isArray(candidates)) {
    return [];
  }
  const childTarget = String(record.child_target || record.target || record.label || "").trim();
  if (!childTarget) {
    return candidates;
  }
  return candidates.filter((summary) => {
    const target = String(summary?.target ?? summary?.label ?? "").trim();
    return target === childTarget;
  });
}

function delegatedChildFailure(record = {}, context = {}) {
  for (const summary of childSummaryCandidates(record, context)) {
    const reason = normalizeFailureReason(summary?.failure_reason, "");
    if (reason && reason !== "child_target_failure") {
      return {
        failure_reason: reason,
        failure_class: normalizeFailureClass(summary?.failure_class ?? failureClassForReason(reason), "unknown"),
      };
    }
    const failure = primaryPublicFailure(summary?.failures ?? []);
    if (failure && failure.failure_reason !== "child_target_failure") {
      return failure;
    }
  }
  return null;
}

export function publicExitCodeForFailure(record = {}, context = {}) {
  const failure = normalizeFailureRecord(record, {
    failure_reason: context.failure_reason,
    failure_class: context.failure_class,
  });
  if (failure.failure_reason === "child_target_failure") {
    const delegated = delegatedChildFailure(failure, context);
    return delegated
      ? publicExitCodeForFailure(delegated, context)
      : reasonPublicExitCodeMap.get("unknown_failure");
  }
  if (failure.failure_reason === "cancelled_or_interrupted") {
    return publicExitCodeForInterrupted(context);
  }
  return reasonPublicExitCodeMap.get(failure.failure_reason) ?? reasonPublicExitCodeMap.get("unknown_failure");
}

export function primaryPublicFailure(failures = [], fallback = null) {
  const normalized = failures
    .filter(Boolean)
    .map((failure, index) => ({
      ...normalizeFailureRecord(failure),
      __input_order: index,
    }))
    .filter((failure) => failure.failure_class);
  const selectByClassPrecedence = (candidates) => {
    const selected = candidates.slice().sort(compareFailureRecords)[0] ?? null;
    if (!selected) {
      return null;
    }
    const { __input_order: _inputOrder, ...publicFailure } = selected;
    return publicFailure;
  };
  const nonCleanup = normalized.filter(
    (failure) => failure.failure_reason !== "cleanup_error",
  );
  if (nonCleanup.length > 0) {
    return selectByClassPrecedence(nonCleanup);
  }
  if (normalized.length > 0) {
    return selectByClassPrecedence(normalized);
  }
  if (fallback?.failure_reason || fallback?.failure_class) {
    return normalizeFailureRecord(fallback);
  }
  return null;
}

export function publicExitCodeForFailures(failures = [], context = {}) {
  const failure = primaryPublicFailure(failures, context);
  if (!failure) {
    return 0;
  }
  return publicExitCodeForFailure(failure, context);
}

export function publicExitCodeForSummary(summary = {}, context = {}) {
  if (!summary || summary.status === "pass") {
    return 0;
  }
  return publicExitCodeForFailures(summary.failures ?? [], {
    ...context,
    failure_class: summary.failure_class ?? context.failure_class,
    failure_reason: summary.failure_reason ?? context.failure_reason,
    childSummaries: context.childSummaries ?? summary.child_summaries,
  });
}

export function summarizeFailures(failures = [], counts = {}) {
  const normalized = failures
    .filter(Boolean)
    .map((failure) => normalizeFailureRecord(failure))
    .filter((failure) => failure.failure_class);
  const failureClasses = createFailureClassCounts();
  const failureReasons = createFailureReasonCounts();
  for (const failure of normalized) {
    failureClasses[failure.failure_class] += 1;
    failureReasons[failure.failure_reason] += 1;
  }
  const primaryFailure = primaryPublicFailure(normalized);
  const failureClass = primaryFailure?.failure_class ?? primaryFailureClass(failureClasses);
  const failureReason = primaryFailure?.failure_reason ?? null;
  const headline = failureClass
    ? formatFailureHeadline({ failureClass, failures: normalized, counts })
    : "";
  return {
    failure_class: failureClass,
    failure_reason: failureReason,
    failure_classes: failureClasses,
    failure_reasons: failureReasons,
    failures: normalized,
    failure_headline: headline,
  };
}

export function failureFieldsForJSON(failures = [], counts = {}) {
  return summarizeFailures(failures, counts);
}

export function mergeFailureFields(...values) {
  return values.flatMap((value) => value?.failures ?? []);
}

export function classifyDossierFailure(dossier = {}, context = {}) {
  if (dossier.failure_class) {
    return normalizeFailureClass(dossier.failure_class);
  }
  if (dossier.coverage && dossier.coverage !== "non_test") {
    return "product";
  }
  return classifyExecutionFailure(context.target ?? "", context.label ?? "", context.command ?? "");
}

export function failuresFromDossiers(dossiers = [], context = {}) {
  return dossiers.map((dossier) =>
    normalizeFailureRecord(
      {
        failure_class: classifyDossierFailure(dossier, context),
        failure_reason: dossier.failure_reason ?? context.failure_reason,
        kind: dossier.kind ?? "failure",
        source: dossier.source ?? dossier.runner ?? context.runner ?? "",
        target: context.target ?? "",
        phase: dossier.phase ?? context.phase ?? "",
        runner: dossier.runner ?? context.runner ?? "",
        label: dossier.symbol_or_title ?? dossier.package_or_file ?? context.label ?? "",
        message: dossier.message ?? "",
        artifact: dossier.raw ?? "",
      },
      { target: context.target ?? "", runner: context.runner ?? "" },
    ),
  );
}

export function classifyExecutionFailure(target = "", label = "", command = "") {
  const joined = `${target} ${label} ${command}`.toLowerCase();
  if (joined.includes("deployable-shape")) {
    return "artifact";
  }
  if (
    joined.includes("duration-baseline-drift") ||
    joined.includes("duration baseline drift") ||
    joined.includes("scheduler-summary-timing-drift") ||
    joined.includes("scheduler summary timing drift")
  ) {
    return "timing";
  }
  return "harness";
}

export function classifyExecutionFailureReason(target = "", label = "", command = "") {
  const joined = `${target} ${label} ${command}`.toLowerCase();
  if (
    joined.includes("duration-baseline-drift") ||
    joined.includes("duration baseline drift") ||
    joined.includes("scheduler-summary-timing-drift") ||
    joined.includes("scheduler summary timing drift")
  ) {
    return "duration_baseline_drift";
  }
  if (joined.includes("postgres-fixture-shape")) {
    return "fixture_error";
  }
  return defaultReasonForFailureClass(classifyExecutionFailure(target, label, command));
}

export function classifyTimingFailure(span = {}) {
  const source = String(span.source ?? "").toLowerCase();
  const bucket = String(span.bucket ?? "").toLowerCase();
  const label = String(span.label ?? "").toLowerCase();
  if (source === "test_services") {
    if (bucket === "teardown" || label.includes("cleanup") || label.includes("janitor") || label.includes("leak")) {
      return "artifact";
    }
    if (
      bucket === "service_wait" ||
      bucket === "migration" ||
      bucket === "server_startup" ||
      bucket === "frontend_startup" ||
      label.includes("postgres") ||
      label.includes("object-store") ||
      label.includes("service")
    ) {
      return "infra";
    }
  }
  return "timing";
}

export function timingFailureRecord(span = {}, defaults = {}) {
  return normalizeFailureRecord(
    {
      failure_class: classifyTimingFailure(span),
      kind: "timing",
      source: span.source ?? "timing",
      target: defaults.target ?? "",
      label: span.label ?? span.bucket ?? "timing",
      message: span.label ?? span.bucket ?? "timing failure",
      artifact: span.artifact ?? "",
    },
    defaults,
  );
}

export function artifactFailureRecord(message, defaults = {}) {
  return normalizeFailureRecord(
    {
      failure_class: "artifact",
      failure_reason: "artifact_error",
      kind: "artifact",
      message,
      label: message,
      source: defaults.source ?? "summary",
    },
    defaults,
  );
}

export function manifestMismatchFailureRecord(mismatch = {}, defaults = {}) {
  const missing = mismatch.missing_ids?.length ?? 0;
  const unexpected = mismatch.unexpected_ids?.length ?? 0;
  const message =
    missing > 0 || unexpected > 0
      ? `manifest mismatch missing=${missing} unexpected=${unexpected}`
      : "manifest mismatch";
  return artifactFailureRecord(message, { ...defaults, source: "manifest" });
}

export function failureHeadlineForSummary(summary = {}) {
  return summary.failure_headline || formatFailureHeadline({
    failureClass: summary.failure_class,
    failures: summary.failures ?? [],
    counts: summary.counts ?? {},
  });
}

export function formatFailureHeadline({ failureClass, failures = [], counts = {} } = {}) {
  const normalizedClass = normalizeFailureClass(failureClass, "");
  if (!normalizedClass) {
    return "";
  }
  const primary =
    failures.find((failure) => failure.failure_class === normalizedClass) ??
    failures[0] ??
    normalizeFailureRecord({ failure_class: normalizedClass, message: "unclassified failure" });
  const failedTests =
    (counts.authoritative_failed ?? 0) +
    (counts.support_failed ?? 0) +
    (counts.raw_failed ?? 0) +
    (counts.tooling_support_failed ?? 0) +
    (counts.unowned_regression_failed ?? 0) +
    (counts.unmapped_failed ?? 0);
  const testsPassedPrefix = (counts.tests ?? 0) > 0 && failedTests === 0 ? "tests passed; " : "";
  const kind = primary.kind && primary.kind !== "failure" ? `${primary.kind} ` : "";
  const detail =
    primary.message ||
    primary.label ||
    primary.source ||
    primary.target ||
    "unclassified failure";
  const reason = primary.failure_reason ? ` reason=${primary.failure_reason}` : "";
  return `${testsPassedPrefix}${normalizedClass}${reason} ${kind}failure: ${detail}`;
}
