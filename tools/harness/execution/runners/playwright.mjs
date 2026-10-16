import path from "node:path";
import { browserFixtureDiagnostic } from "../../contract/index.mjs";
import {
  primaryPublicFailure,
  publicExitCodeForFailure,
  publicExitCodeForFailures,
} from "../../contract/failure-taxonomy.mjs";

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function normalizeFile(value) {
  const normalized = String(value ?? "").replaceAll("\\", "/").replace(/^.*?apps\/web\//u, "apps/web/");
  return normalized.startsWith("apps/web/") || normalized === "" ? normalized : `apps/web/e2e/${normalized}`;
}

function flattenSuites(suites, specs = [], inheritedFile = "", ancestors = [], depth = 0) {
  for (const suite of suites ?? []) {
    const suiteFile = suite.file || inheritedFile;
    const titlePath = depth === 0 ? [] : [...ancestors, suite.title];
    flattenSuites(suite.suites, specs, suiteFile, titlePath, depth + 1);
    for (const spec of suite.specs ?? []) specs.push({ ...spec, file: spec.file || suiteFile, titlePath: [...titlePath, spec.title] });
  }
  return specs;
}

export function playwrightTestList(selection) {
  if (!selection.length) throw new Error("browser test list must not be empty");
  return selection.map(entry => {
    const file = path.posix.relative("apps/web/e2e", entry.file);
    const parts = [`[${entry.project_id}]`, file, ...entry.title_path];
    if (!file || file.startsWith("../") || parts.some(part => !part || part.trim() !== part || /[\r\n\u203a]/u.test(part))) {
      throw new Error("invalid exact browser test-list identity");
    }
    return parts.join(" › ");
  }).join("\n") + "\n";
}

function resultStatus(test) {
  const results = Array.isArray(test?.results) ? test.results : [];
  const terminal = results.at(-1)?.status;
  if (terminal) return terminal;
  if (test?.status === "expected") return "passed";
  if (test?.status === "unexpected") return "failed";
  return test?.status ?? "missing";
}

function selectorObservation(spec) {
  const tests = Array.isArray(spec?.tests) ? spec.tests : [];
  if (tests.length !== 1) return { status: "ambiguous", durationMs: 0 };
  const test = tests[0];
  const statuses = (test.results ?? []).map((result) => result.status);
  const durationMs = (test.results ?? []).reduce(
    (total, result) => total + (Number.isFinite(result.duration) ? result.duration : 0),
    0,
  );
  const status = resultStatus(test);
  const fixture = browserFixtureDiagnostic(test.results?.at(-1));
  if (fixture?.failures.length) return { status: "fixture_failed", durationMs, ...fixture };
  const firstError = test.results?.at(-1)?.errors?.[0];
  if (status === "failed" && /^Cartulary(?:FrontendArtifact|VisualCapture)Error:/u.test(firstError?.message ?? "")) {
    return { status: "artifact_failed", durationMs };
  }
  if (status === "passed") return { status, durationMs };
  if (status === "failed") return { status, durationMs };
  if (status === "timedOut" || statuses.includes("timedOut")) {
    return { status: "timed_out", durationMs };
  }
  if (status === "interrupted" || statuses.includes("interrupted")) {
    return { status: "interrupted", durationMs };
  }
  if (status === "skipped") return { status, durationMs };
  return { status: "missing", durationMs };
}

function observationFailure(observation) {
  if (observation.status === "artifact_failed") {
    return { failure_class: "artifact", failure_reason: "artifact_error" };
  }
  if (["failed", "timed_out"].includes(observation.status)) {
    return {
      failure_class: "product",
      failure_reason: "test_assertion_failure",
    };
  }
  if (observation.status === "interrupted") {
    return {
      failure_class: "interrupted",
      failure_reason: "cancelled_or_interrupted",
    };
  }
  if (["missing", "ambiguous", "skipped"].includes(observation.status)) {
    return {
      failure_class: "harness",
      failure_reason: "scheduler_accounting_error",
    };
  }
  return null;
}

function processFailure(processStatus, processSignal) {
  if (processSignal) {
    return {
      failure_class: "interrupted",
      failure_reason: "cancelled_or_interrupted",
    };
  }
  if (Number.isInteger(processStatus) && processStatus !== 0) {
    return {
      failure_class: "harness",
      failure_reason: "scheduler_accounting_error",
    };
  }
  return null;
}

function terminalStateForFailure(failure) {
  if (failure === null) return "passed";
  if (failure.failure_class === "interrupted") return "cancelled";
  if (failure.failure_class === "product") return "failed";
  return "infrastructure_failed";
}

function exitContext(processStatus, processSignal) {
  return {
    signal: processSignal,
    status: Number.isInteger(processStatus) ? processStatus : undefined,
  };
}

export function adaptPlaywrightReport(
  rows,
  report,
  processStatus = 0,
  processSignal = null,
  selection,
) {
  const specs = flattenSuites(report?.suites).map((spec) => ({
    ...spec,
    normalizedFile: normalizeFile(spec.file),
  }));
  if (!Array.isArray(selection) || rows.some(row => selection.filter(entry => entry.row_id === row.row_id).length !== row.selector.titles.length)) {
    throw new Error("browser report requires its exact resolved selection");
  }
  const selected = [...rows]
    .sort((left, right) => asciiCompare(left.row_id, right.row_id))
    .map((row) => {
      const observations = selection.filter(entry => entry.row_id === row.row_id).map((entry) => {
        const matches = specs.filter(
          (spec) => spec.normalizedFile === normalizeFile(entry.file) &&
            JSON.stringify(spec.titlePath) === JSON.stringify(entry.title_path) &&
            spec.tests?.length === 1 && spec.tests[0].projectName === entry.project_id,
        );
        return matches.length === 1
          ? selectorObservation(matches[0])
          : { status: matches.length === 0 ? "missing" : "ambiguous", durationMs: 0 };
      });
      return { observations, row };
    });
  const unexpected = specs.some(spec => !selection.some(entry =>
    spec.normalizedFile === normalizeFile(entry.file) && JSON.stringify(spec.titlePath) === JSON.stringify(entry.title_path) &&
    spec.tests?.length === 1 && spec.tests[0].projectName === entry.project_id));
  const childFailure = unexpected ? { failure_class: "harness", failure_reason: "scheduler_accounting_error" } : selected.every(({ observations }) =>
    observations.every((observation) => observation.status === "passed")
  )
    ? processFailure(processStatus, processSignal)
    : null;
  return selected.map(({ observations, row }) => {
      const failures = observations.flatMap((entry) => entry.failures ?? [observationFailure(entry)]).filter(Boolean);
      if (childFailure) failures.push(childFailure);
      const primaryFailure = primaryPublicFailure(failures);
      const terminalState = terminalStateForFailure(primaryFailure);
      return {
        row_id: row.row_id,
        terminal_state: terminalState,
        duration_ms: observations.reduce((total, entry) => total + entry.durationMs, 0),
        exit_code: primaryFailure
          ? publicExitCodeForFailure(
              primaryFailure,
              exitContext(processStatus, processSignal),
            )
          : 0,
        failure_class: primaryFailure?.failure_class ?? null,
        failure_reason: primaryFailure?.failure_reason ?? null,
        failure_diagnostic: observations.find((entry) => entry.diagnostic && entry.failures.some((failure) => failure.failure_class === primaryFailure?.failure_class && failure.failure_reason === primaryFailure?.failure_reason))?.diagnostic ?? null,
      };
    });
}

export function playwrightGroupExitCode(rowResults, child = {}) {
  const failures = rowResults
    .filter((row) => row.terminal_state !== "passed")
    .map((row) => ({
      failure_class: row.failure_class,
      failure_reason: row.failure_reason,
    }));
  const childFailure = processFailure(child.status, child.signal);
  if (childFailure && failures.length === 0) {
    failures.push(childFailure);
  }
  return publicExitCodeForFailures(
    failures,
    exitContext(child.status, child.signal),
  );
}
