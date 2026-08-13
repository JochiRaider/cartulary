export const runnerContract = Object.freeze({
  runner: "playwright",
  selector_kind: "playwright_exact_scenarios",
});

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function normalizeFile(value) {
  const normalized = String(value ?? "").replaceAll("\\", "/").replace(/^.*?apps\/web\//u, "apps/web/");
  return normalized.startsWith("apps/web/") || normalized === "" ? normalized : `apps/web/e2e/${normalized}`;
}

function flattenSuites(suites, specs = [], inheritedFile = "") {
  for (const suite of suites ?? []) {
    const suiteFile = suite.file || inheritedFile;
    flattenSuites(suite.suites, specs, suiteFile);
    for (const spec of suite.specs ?? []) specs.push({ ...spec, file: spec.file || suiteFile });
  }
  return specs;
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
  if (tests.length !== 1) return { status: "ambiguous", durationMs: 0, exitCode: 11 };
  const test = tests[0];
  const statuses = (test.results ?? []).map((result) => result.status);
  const durationMs = (test.results ?? []).reduce(
    (total, result) => total + (Number.isFinite(result.duration) ? result.duration : 0),
    0,
  );
  const status = resultStatus(test);
  if (status === "passed") return { status, durationMs, exitCode: 0 };
  if (status === "failed") return { status, durationMs, exitCode: 10 };
  if (status === "timedOut" || statuses.includes("timedOut")) {
    return { status: "timed_out", durationMs, exitCode: 13 };
  }
  if (status === "interrupted" || statuses.includes("interrupted")) {
    return { status: "interrupted", durationMs, exitCode: 130 };
  }
  if (status === "skipped") return { status, durationMs, exitCode: 11 };
  return { status: "missing", durationMs, exitCode: 11 };
}

export function adaptPlaywrightReport(rows, report, processStatus = 0) {
  const specs = flattenSuites(report?.suites).map((spec) => ({
    ...spec,
    normalizedFile: normalizeFile(spec.file),
  }));
  return [...rows]
    .sort((left, right) => asciiCompare(left.row_id, right.row_id))
    .map((row) => {
      const observations = row.selector.titles.map((title) => {
        const matches = specs.filter(
          (spec) => spec.normalizedFile === normalizeFile(row.selector.file) && spec.title === title,
        );
        return matches.length === 1
          ? selectorObservation(matches[0])
          : { status: matches.length === 0 ? "missing" : "ambiguous", durationMs: 0, exitCode: 11 };
      });
      const timedOut = observations.some((entry) => entry.status === "timed_out");
      const interrupted = observations.some((entry) => entry.status === "interrupted");
      const missing = observations.some((entry) => ["missing", "ambiguous", "skipped"].includes(entry.status));
      const failed = observations.some((entry) => entry.status === "failed");
      const terminalState = interrupted || timedOut || missing
        ? "infrastructure_failed"
        : failed
          ? "failed"
          : "passed";
      return {
        row_id: row.row_id,
        terminal_state: terminalState,
        duration_ms: observations.reduce((total, entry) => total + entry.durationMs, 0),
        exit_code: interrupted
          ? 130
          : timedOut
            ? 13
            : terminalState === "passed"
              ? 0
              : terminalState === "failed"
                ? 10
                : processStatus || 11,
        failure_class: interrupted
          ? "interrupted"
          : timedOut || missing
            ? "infra"
            : failed
              ? "product"
              : null,
        failure_reason: interrupted
          ? "interrupted"
          : timedOut
            ? "timeout"
            : missing
              ? "missing_or_ambiguous_selector_result"
              : failed
                ? "test_assertion_failure"
                : null,
        failure_diagnostic: null,
      };
    });
}
