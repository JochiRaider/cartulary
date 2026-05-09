export const failureClassOrder = [
  "product",
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
  "child_target_failure",
  "scheduler_accounting_error",
  "artifact_error",
  "cleanup_error",
  "timeout_failure",
  "cancelled_or_interrupted",
  "unknown_failure",
];

const failureClassSet = new Set(failureClassOrder);
const failureReasonSet = new Set(failureReasonOrder);

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
  ["child_target_failure", "harness"],
  ["scheduler_accounting_error", "harness"],
  ["artifact_error", "artifact"],
  ["cleanup_error", "harness"],
  ["timeout_failure", "timing"],
  ["cancelled_or_interrupted", "interrupted"],
  ["unknown_failure", "unknown"],
]);

const classDefaultReasonMap = new Map([
  ["product", "test_assertion_failure"],
  ["config", "configuration_error"],
  ["infra", "preflight_error"],
  ["harness", "unknown_failure"],
  ["artifact", "artifact_error"],
  ["timing", "timeout_failure"],
  ["interrupted", "cancelled_or_interrupted"],
  ["unknown", "unknown_failure"],
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
  return {
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
  const failureClass = primaryFailureClass(failureClasses);
  const failureReason =
    normalized.find((failure) => failure.failure_class === failureClass)?.failure_reason ??
    null;
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
  if (joined.includes("duration-baseline-drift") || joined.includes("duration baseline drift")) {
    return "timing";
  }
  return "harness";
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
      label.includes("minio") ||
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
