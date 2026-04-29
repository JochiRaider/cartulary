export const failureClassOrder = ["test", "infra", "timing", "artifact", "helper"];

const failureClassSet = new Set(failureClassOrder);

export function createFailureClassCounts() {
  return Object.fromEntries(failureClassOrder.map((failureClass) => [failureClass, 0]));
}

export function normalizeFailureClass(value, fallback = "helper") {
  const normalized = String(value ?? "").trim().toLowerCase();
  if (failureClassSet.has(normalized)) {
    return normalized;
  }
  if (fallback === "" || fallback === null) {
    return null;
  }
  return failureClassSet.has(fallback) ? fallback : "helper";
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
  const failureClass = normalizeFailureClass(record.failure_class ?? defaults.failure_class);
  return {
    failure_class: failureClass,
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
  for (const failure of normalized) {
    failureClasses[failure.failure_class] += 1;
  }
  const failureClass = primaryFailureClass(failureClasses);
  const headline = failureClass
    ? formatFailureHeadline({ failureClass, failures: normalized, counts })
    : "";
  return {
    failure_class: failureClass,
    failure_classes: failureClasses,
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
    return "test";
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
  return "helper";
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
  return `${testsPassedPrefix}${normalizedClass} ${kind}failure: ${detail}`;
}
