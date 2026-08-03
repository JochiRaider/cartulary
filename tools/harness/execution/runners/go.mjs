export const runnerContract = Object.freeze({
  runner: "go",
  selector_kind: "go_exact_tests",
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

export function buildGoInvocations(rows, workers, command = process.env.GO || "go") {
  const byPackage = new Map();
  for (const row of rows) {
    const values = byPackage.get(row.selector.package) ?? [];
    values.push(row);
    byPackage.set(row.selector.package, values);
  }
  return [...byPackage.entries()]
    .sort(([left], [right]) => asciiCompare(left, right))
    .map(([packageName, packageRows]) => ({
      command,
      args: [
        "test",
        "-json",
        "-count=1",
        `-p=${Math.min(workers, 16)}`,
        "-run",
        exactAlternation(
          [...new Set(packageRows.flatMap((row) => row.selector.tests))].sort(asciiCompare),
        ),
        packageName,
      ],
      rows: packageRows.map((row) => ({
        row_id: row.row_id,
        selectors: [...row.selector.tests],
      })),
    }));
}

function parseEvents(stdout) {
  const terminal = new Map();
  for (const line of String(stdout).split("\n")) {
    if (line.trim() === "") continue;
    let event;
    try {
      event = JSON.parse(line);
    } catch {
      continue;
    }
    if (
      typeof event.Test === "string" &&
      ["pass", "fail", "skip"].includes(event.Action)
    ) {
      terminal.set(event.Test, {
        action: event.Action,
        duration_ms: Math.max(0, Math.round(Number(event.Elapsed ?? 0) * 1000)),
      });
    }
  }
  return terminal;
}

export function adaptGoInvocation(invocation, result) {
  const terminal = parseEvents(result.stdout);
  return invocation.rows.map((row) => {
    const observations = row.selectors.map((selector) => terminal.get(selector));
    const missing = observations.some((entry) => entry === undefined);
    const skipped = observations.some((entry) => entry?.action === "skip");
    const failed = observations.some((entry) => entry?.action === "fail");
    const terminalState = missing || skipped
      ? "infrastructure_failed"
      : failed
        ? "failed"
        : "passed";
    return {
      row_id: row.row_id,
      terminal_state: terminalState,
      duration_ms: observations.reduce((sum, entry) => sum + (entry?.duration_ms ?? 0), 0),
      exit_code: terminalState === "passed" ? 0 : terminalState === "failed" ? 10 : 3,
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
