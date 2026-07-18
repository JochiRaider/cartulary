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
  const terminalState = result.status === 0 ? "passed" : "failed";
  return invocation.rows.map((row) => ({
    row_id: row.row_id,
    terminal_state: terminalState,
    duration_ms: 0,
    exit_code: result.status,
    failure_reason: terminalState === "passed" ? null : "test_assertion_failure",
  }));
}
