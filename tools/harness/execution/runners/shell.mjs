import { operationDefinition, readOperationOwner } from "../../scheduler/work-graph/operations.mjs";

export function buildShellInvocations(rows, root = process.cwd()) {
  const owner = readOperationOwner(root);
  return rows.map((row) => ({
    command: process.execPath,
    args: ["tools/harness/execution/operation-runner-cli.mjs", (() => {
      const id = row.selector.operation_id;
      operationDefinition(owner, id);
      return id;
    })()],
    rows: [{ row_id: row.row_id, selectors: [row.selector.operation_id] }],
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
