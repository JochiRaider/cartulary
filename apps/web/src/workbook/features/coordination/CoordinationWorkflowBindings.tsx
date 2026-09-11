import { coordinationWorkflowTestId } from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { genericRowLabel } from "../../models/genericWorkbookModel";
import type {
  CoordinationMutationCommandPort,
  TaskLifecycleStatus,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import {
  type CoordinationWorkflowMutationPorts,
  useCoordinationWorkflowController,
} from "./useCoordinationWorkflowController";

function parseTaskLifecycleStatus(value: string): TaskLifecycleStatus | null {
  switch (value) {
    case "blocked":
    case "canceled":
    case "done":
    case "in_progress":
    case "open":
      return value;
  }
  return null;
}

export function CoordinationWorkflowBindings({
  contract,
  disabled,
  mutation,
  mutationCommands,
  ownerBindings,
  resetKey,
  rows,
}: {
  readonly contract: ViewContract;
  readonly disabled: boolean;
  readonly mutation: CoordinationWorkflowMutationPorts;
  readonly mutationCommands: CoordinationMutationCommandPort;
  readonly ownerBindings: readonly WorkbookOwnerBinding[];
  readonly resetKey: string;
  readonly rows: readonly WorkbookQueryRow[];
}) {
  const workflow = useCoordinationWorkflowController({
    mutation,
    mutationCommands,
    resetKey,
    rows,
  });

  if (ownerBindings.includes("task_lifecycle") && rows.length > 0) {
    return (
      <div style={workflowRowStyle}>
        <select
          aria-label="Task lifecycle row"
          data-testid={coordinationWorkflowTestId("task-target")}
          style={selectStyle}
          value={workflow.lifecycle.recordId}
          onChange={(event) =>
            workflow.lifecycle.setRecordId(event.target.value)
          }
        >
          <option value="">Task</option>
          {rows.map((row) => (
            <option key={row.record_id} value={row.record_id}>
              {genericRowLabel(contract, row)}
            </option>
          ))}
        </select>
        <select
          aria-label="Task lifecycle status"
          data-testid={coordinationWorkflowTestId("task-status")}
          style={selectStyle}
          value={workflow.lifecycle.status}
          onChange={(event) => {
            const status = parseTaskLifecycleStatus(event.target.value);
            if (status !== null) workflow.lifecycle.setStatus(status);
          }}
        >
          <option value="open">open</option>
          <option value="in_progress">in_progress</option>
          <option value="blocked">blocked</option>
          <option value="done">done</option>
          <option value="canceled">canceled</option>
        </select>
        <input
          aria-label="Blocked reason"
          data-testid={coordinationWorkflowTestId("task-blocked-reason")}
          disabled={workflow.lifecycle.status !== "blocked"}
          style={inputStyle}
          type="text"
          value={workflow.lifecycle.blockedReason}
          onChange={(event) =>
            workflow.lifecycle.setBlockedReason(event.target.value)
          }
        />
        <button
          data-testid={coordinationWorkflowTestId("task-submit")}
          disabled={disabled}
          style={actionButtonStyle}
          type="button"
          onClick={() => void workflow.lifecycle.submit()}
        >
          Apply task status
        </button>
      </div>
    );
  }

  return null;
}

const workflowRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  gap: "0.6rem",
  alignItems: "stretch",
};

const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};

const selectStyle = { ...inputStyle, appearance: "auto" as const };

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};
