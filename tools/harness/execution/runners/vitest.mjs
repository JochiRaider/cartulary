import path from "node:path";

export const runnerContract = Object.freeze({
  runner: "vitest",
  selector_kind: "vitest_exact_titles",
});

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function regexEscape(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
}

function exactAlternation(values) {
  return `^(?:${values.map(regexEscape).join("|")})$`;
}

export function buildVitestInvocations(root, rows, workers, command) {
  const byFile = new Map();
  for (const row of rows) {
    const values = byFile.get(row.selector.file) ?? [];
    values.push(row);
    byFile.set(row.selector.file, values);
  }
  return [...byFile.entries()]
    .sort(([left], [right]) => asciiCompare(left, right))
    .map(([file, fileRows]) => ({
      command,
      args: [
        "--dir",
        "apps/web",
        "exec",
        "vitest",
        "run",
        path.resolve(root, file),
        "-t",
        exactAlternation(
          [...new Set(fileRows.flatMap((row) => row.selector.titles))].sort(asciiCompare),
        ),
        "--reporter=json",
        `--maxWorkers=${workers}`,
      ],
      rows: fileRows.map((row) => ({
        row_id: row.row_id,
        selectors: [...row.selector.titles],
      })),
    }));
}

function parseAssertions(stdout) {
  let report;
  try {
    report = JSON.parse(String(stdout).trim());
  } catch {
    return null;
  }
  const assertions = new Map();
  for (const suite of report.testResults ?? []) {
    for (const assertion of suite.assertionResults ?? []) {
      const title = assertion.fullName ?? assertion.title;
      if (typeof title === "string") assertions.set(title, assertion.status);
    }
  }
  return assertions;
}

export function adaptVitestInvocation(invocation, result) {
  const assertions = parseAssertions(result.stdout);
  return invocation.rows.map((row) => {
    const statuses = row.selectors.map((selector) => assertions?.get(selector));
    const missing = !assertions || statuses.some((entry) => entry === undefined);
    const skipped = statuses.some((entry) => ["pending", "skipped", "todo"].includes(entry));
    const failed = statuses.some((entry) => entry === "failed");
    const terminalState = missing || skipped
      ? "infrastructure_failed"
      : failed
        ? "failed"
        : "passed";
    return {
      row_id: row.row_id,
      terminal_state: terminalState,
      duration_ms: 0,
      exit_code: terminalState === "passed" ? 0 : result.status,
      failure_reason: missing
        ? "missing_selector_result"
        : skipped
          ? "unauthorized_skip"
          : failed
            ? "test_assertion_failure"
            : null,
    };
  });
}
