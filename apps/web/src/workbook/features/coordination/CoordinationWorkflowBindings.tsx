import type { ViewContract } from "@cartulary/view-contracts";
import { useEffect, useState } from "react";
import { apiPath } from "../../../services/browserApi";
import { fetchWorkbookJSON } from "../../../services/workbookApi";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import {
  genericRowLabel,
  normalizeGenericTextValue,
} from "../../models/genericWorkbookModel";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import { taskRequestsViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { EntityApiRow } from "../../timeline/models/workbookTimelineModel";

type DecisionSupersedeEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    target_record_id: string;
    superseding_record_id: string;
    target_row_version: number;
    superseding_row_version: number;
    target_status: string;
    reason: string;
  };
};

type MutationPorts = Pick<
  GenericSurfaceMutationController,
  | "beginMutation"
  | "completeGenericMutation"
  | "rejectMutationPayload"
  | "setValidationError"
>;

export function CoordinationWorkflowBindings({
  apiBase,
  contract,
  disabled,
  mutation,
  ownerBindings,
  referenceOptions,
  resetKey,
  rows,
}: {
  readonly apiBase: string | undefined;
  readonly contract: ViewContract;
  readonly disabled: boolean;
  readonly mutation: MutationPorts;
  readonly ownerBindings: readonly WorkbookOwnerBinding[];
  readonly referenceOptions: GenericReferenceOptions;
  readonly resetKey: string;
  readonly rows: readonly EntityApiRow[];
}) {
  const [lifecycleRecordId, setLifecycleRecordId] = useState("");
  const [lifecycleStatus, setLifecycleStatus] = useState("blocked");
  const [lifecycleBlockedReason, setLifecycleBlockedReason] = useState("");
  const [supersedeTargetId, setSupersedeTargetId] = useState("");
  const [supersedeReplacementId, setSupersedeReplacementId] = useState("");
  const [supersedeReason, setSupersedeReason] = useState("");

  useEffect(() => {
    if (resetKey === "") {
      return;
    }
    setLifecycleRecordId("");
    setLifecycleBlockedReason("");
    setSupersedeTargetId("");
    setSupersedeReplacementId("");
    setSupersedeReason("");
  }, [resetKey]);

  const submitLifecyclePatch = async () => {
    const target = rows.find((row) => row.record_id === lifecycleRecordId);
    if (!target) {
      mutation.setValidationError("Select a task row.");
      return;
    }
    const changes: Array<Record<string, unknown>> = [
      { field_key: "task.status", value: lifecycleStatus },
    ];
    if (lifecycleStatus === "blocked") {
      const reason = normalizeGenericTextValue(lifecycleBlockedReason);
      if (reason === "") {
        mutation.setValidationError("Blocked tasks need a reason.");
        return;
      }
      changes.push({ field_key: "task.blocked_reason", value: reason });
    }
    mutation.beginMutation();
    const result = await fetchWorkbookJSON(
      apiPath(apiBase, `/api/v1/records/${target.record_id}`),
      {
        method: "PATCH",
        body: JSON.stringify({
          view_schema_id: taskRequestsViewSchemaId,
          base_row_version: target.row_version,
          client_txn_id: `task-lifecycle-${Date.now()}`,
          changes,
        }),
      },
    );
    if (!result.ok) {
      mutation.rejectMutationPayload(result.payload);
      return;
    }
    if (lifecycleStatus !== "blocked") {
      setLifecycleBlockedReason("");
    }
    await mutation.completeGenericMutation(result.payload);
  };

  const submitSupersede = async () => {
    const target = rows.find((row) => row.record_id === supersedeTargetId);
    if (!target || supersedeReplacementId === "") {
      mutation.setValidationError("Select target and superseding decisions.");
      return;
    }
    if (target.record_id === supersedeReplacementId) {
      mutation.setValidationError("Select a different superseding decision.");
      return;
    }
    const reason = normalizeGenericTextValue(supersedeReason);
    if (reason === "") {
      mutation.setValidationError("Reason is required.");
      return;
    }
    mutation.beginMutation();
    const result = await fetchWorkbookJSON<DecisionSupersedeEnvelope>(
      apiPath(apiBase, `/api/v1/records/${target.record_id}/supersede`),
      {
        method: "POST",
        body: JSON.stringify({
          base_row_version: target.row_version,
          client_txn_id: `decision-supersede-${Date.now()}`,
          replacement_record_id: supersedeReplacementId,
          reason,
        }),
      },
    );
    if (!result.ok) {
      mutation.rejectMutationPayload(result.payload);
      return;
    }
    setSupersedeReason("");
    await mutation.completeGenericMutation<DecisionSupersedeEnvelope>(
      result.payload,
    );
  };

  if (ownerBindings.includes("task_lifecycle") && rows.length > 0) {
    return (
      <div style={workflowRowStyle}>
        <select
          aria-label="Task lifecycle row"
          data-testid="task-lifecycle-target"
          style={selectStyle}
          value={lifecycleRecordId}
          onChange={(event) => setLifecycleRecordId(event.target.value)}
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
          data-testid="task-lifecycle-status"
          style={selectStyle}
          value={lifecycleStatus}
          onChange={(event) => setLifecycleStatus(event.target.value)}
        >
          <option value="open">open</option>
          <option value="in_progress">in_progress</option>
          <option value="blocked">blocked</option>
          <option value="done">done</option>
          <option value="canceled">canceled</option>
        </select>
        <input
          aria-label="Blocked reason"
          data-testid="task-lifecycle-blocked-reason"
          disabled={lifecycleStatus !== "blocked"}
          style={inputStyle}
          type="text"
          value={lifecycleBlockedReason}
          onChange={(event) => setLifecycleBlockedReason(event.target.value)}
        />
        <button
          data-testid="task-lifecycle-submit"
          disabled={disabled}
          style={actionButtonStyle}
          type="button"
          onClick={() => void submitLifecyclePatch()}
        >
          Apply task status
        </button>
      </div>
    );
  }

  if (ownerBindings.includes("decision_supersede") && rows.length > 1) {
    return (
      <div style={workflowRowStyle}>
        <select
          aria-label="Superseded decision"
          data-testid="decision-supersede-target"
          style={selectStyle}
          value={supersedeTargetId}
          onChange={(event) => setSupersedeTargetId(event.target.value)}
        >
          <option value="">Target</option>
          {rows.map((row) => (
            <option key={row.record_id} value={row.record_id}>
              {genericRowLabel(contract, row)}
            </option>
          ))}
        </select>
        <select
          aria-label="Superseding decision"
          data-testid="decision-supersede-replacement"
          style={selectStyle}
          value={supersedeReplacementId}
          onChange={(event) => setSupersedeReplacementId(event.target.value)}
        >
          <option value="">Superseding</option>
          {referenceOptions.decisions.map((option) => (
            <option key={option.recordId} value={option.recordId}>
              {option.label}
            </option>
          ))}
        </select>
        <input
          aria-label="Decision supersession reason"
          data-testid="decision-supersede-reason"
          style={inputStyle}
          type="text"
          value={supersedeReason}
          onChange={(event) => setSupersedeReason(event.target.value)}
        />
        <button
          data-testid="decision-supersede-submit"
          disabled={disabled}
          style={actionButtonStyle}
          type="button"
          onClick={() => void submitSupersede()}
        >
          Supersede decision
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
