export const runnerContract = Object.freeze({
  runner: "shell",
  selector_kind: "shell_registered_command",
});

export function buildShellInvocations(rows, command = process.env.MAKE || "make") {
  return rows.map((row) => ({
    command,
    args: ["--silent", "--no-print-directory", row.target_name],
    rows: [{ row_id: row.row_id, selectors: [row.selector.command_id] }],
  }));
}

export function adaptShellInvocation(invocation, result) {
  const failure = result.commandFailure;
  const terminalState = result.status === 0 && !failure ? "passed" : "failed";
  return invocation.rows.map((row) => ({
    row_id: row.row_id,
    terminal_state: terminalState,
    duration_ms: 0,
    exit_code: result.status,
    failure_class: terminalState === "passed" ? null : result.status === 0 ? "harness" : failure?.failure_class ?? "product",
    failure_reason: terminalState === "passed" ? null : result.status === 0 ? "scheduler_accounting_error" : failure?.failure_reason ?? "test_assertion_failure",
    failure_diagnostic: null,
  }));
}
