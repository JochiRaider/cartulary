import {
  coordinationWorkflowTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorDisabledCondition,
  ViewContract,
} from "@cartulary/view-contracts";
import { getReferenceFieldContract } from "@cartulary/view-contracts";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { WorkbookReferenceControl } from "../../components/WorkbookReferenceControl";
import { admitCanonicalInspectorFeature } from "../../inspector/canonicalInspectorAdmission";
import { workbookInspectorDisabledReason } from "../../inspector/presentation/workbookInspectorPresentationModel";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import {
  type TaskLifecycleDraftStore,
  taskStatuses,
  taskTransitionAllowed,
  taskValue,
} from "./taskLifecycleModel";
import {
  type CoordinationWorkflowMutationPorts,
  useCoordinationWorkflowController,
} from "./useCoordinationWorkflowController";

export function CoordinationWorkflowBindings(props: {
  readonly contract: ViewContract;
  readonly disabled: boolean;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly mutation: CoordinationWorkflowMutationPorts;
  readonly drafts: TaskLifecycleDraftStore;
  readonly row: WorkbookQueryRow;
}) {
  const feature = admitCanonicalInspectorFeature(
    props.contract.inspectorConfig,
    "task.status.transition",
  );
  if (
    !feature ||
    feature.panelId !== "workflow" ||
    feature.routeBinding.kind !== "record_patch" ||
    feature.routeBinding.owner !== "record_patch_route" ||
    feature.requiresConfirmation
  )
    return null;
  const disabledReason =
    workbookInspectorDisabledReason({
      currentIncidentRole: props.currentIncidentRole,
      featureGroup: feature,
      stateTokens: props.disabledTokens,
    }) ??
    (props.disabled
      ? "Wait for the current workbook operation to finish."
      : null);
  return (
    <TaskLifecycleEditor
      {...props}
      disabled={disabledReason !== null}
      disabledReason={disabledReason}
    />
  );
}

function TaskLifecycleEditor(
  props: Parameters<typeof CoordinationWorkflowBindings>[0] & {
    readonly disabledReason: string | null;
  },
) {
  const editor = useCoordinationWorkflowController(props);
  const status = editor.value("task.status");
  const from = taskValue(props.row, "task.status");
  const previous = props.mutation.explicitPatches
    ?.getSnapshot()
    .entries.filter(
      (entry) => entry.intent.baseline.record_id === props.row.record_id,
    )
    .at(-1);
  const errors = [
    ...editor.errors,
    ...(previous?.failure?.kind === "validation" &&
    JSON.stringify(previous.intent.changes) === JSON.stringify(editor.changes)
      ? (previous.failure.fields ?? [])
      : []),
  ];
  const ownerReference = getReferenceFieldContract(
    props.contract.viewSchemaId,
    "task.owner_user_id",
  );
  if (!ownerReference) return null;
  const fieldError = (field: string) => (
    <span id={`task-error-${field}`}>
      {errors
        .filter((error) => error.field === field)
        .map((error) => error.message)
        .join(" ")}
    </span>
  );
  return (
    <fieldset
      disabled={props.disabled}
      style={groupStyle}
      aria-label="Task status transition"
    >
      <legend>Task status</legend>
      <p style={textStyle}>Saved status: {from}</p>
      {props.disabledReason ? (
        <p style={textStyle}>{props.disabledReason}</p>
      ) : null}
      <label style={labelStyle}>
        Status
        <select
          id={`task-lifecycle-${props.row.record_id}-task.status`}
          aria-label="Task lifecycle status"
          aria-describedby="task-transition-guidance task-error-task.status"
          data-testid={coordinationWorkflowTestId("task-status")}
          style={inputStyle}
          value={status}
          onChange={(event) => editor.update("task.status", event.target.value)}
        >
          {taskStatuses.map((next) => (
            <option
              key={next}
              value={next}
              disabled={!taskTransitionAllowed(from, next)}
            >
              {next}
            </option>
          ))}
        </select>
        {fieldError("task.status")}
      </label>
      <p id="task-transition-guidance" style={textStyle}>
        {from === "done" || from === "canceled"
          ? `Reopen to open, in_progress, or blocked. ${from === "done" ? "Canceled" : "Done"} is unavailable until reopened.`
          : "Active Tasks can move between open, in_progress, and blocked, or finish as done or canceled."}
      </p>
      <div style={labelStyle}>
        <span>Owner</span>
        <WorkbookReferenceControl
          field={ownerReference}
          label="Task lifecycle owner"
          value={editor.value("task.owner_user_id")}
          sourceRecordId={props.row.record_id}
          disabled={props.disabled}
          id={`task-lifecycle-${props.row.record_id}-task.owner_user_id`}
          testId="task-lifecycle-owner"
          describedBy="task-error-task.owner_user_id"
          onChange={(value) => editor.update("task.owner_user_id", value)}
          onAccept={(items) => {
            const id = items[0]?.identity.id;
            if (id) editor.update("task.owner_user_id", id);
          }}
        />
        {fieldError("task.owner_user_id")}
      </div>
      {status === "blocked" ? (
        <label style={labelStyle}>
          Blocked reason
          <input
            id={`task-lifecycle-${props.row.record_id}-task.blocked_reason`}
            aria-label="Blocked reason"
            aria-describedby="task-error-task.blocked_reason"
            data-testid={coordinationWorkflowTestId("task-blocked-reason")}
            style={inputStyle}
            value={editor.value("task.blocked_reason")}
            onChange={(event) =>
              editor.update("task.blocked_reason", event.target.value)
            }
          />
          {fieldError("task.blocked_reason")}
        </label>
      ) : null}
      {status === "done" ? (
        <label style={labelStyle}>
          Completion time (optional when entering done)
          <input
            id={`task-lifecycle-${props.row.record_id}-task.completed_at`}
            aria-label="Task completion time"
            aria-describedby="task-completion-guidance task-error-task.completed_at"
            style={inputStyle}
            value={editor.value("task.completed_at")}
            placeholder="2026-09-11T12:00:00Z"
            onChange={(event) =>
              editor.update("task.completed_at", event.target.value)
            }
          />
          <span id="task-completion-guidance">
            Leave empty when entering done to use the server commit time. An
            explicit time must include a timezone and cannot precede creation.
          </span>
          {fieldError("task.completed_at")}
        </label>
      ) : null}
      {from === "blocked" && status !== "blocked" ? (
        <p style={textStyle}>The saved blocked reason will be cleared.</p>
      ) : null}
      {from === "done" && status !== "done" ? (
        <p style={textStyle}>The saved completion time will be cleared.</p>
      ) : null}
      {editor.staleFields.map((field) => (
        <div key={field} role="status">
          Saved {props.contract.fieldMap[field]?.label ?? field} changed to{" "}
          {taskValue(props.row, field) || "empty"}. Your draft is retained.
          <button type="button" onClick={() => editor.review(field, false)}>
            Use saved {props.contract.fieldMap[field]?.label}
          </button>
          <button type="button" onClick={() => editor.review(field, true)}>
            Keep draft {props.contract.fieldMap[field]?.label}
          </button>
        </div>
      ))}
      <button
        data-testid={workbookInspectorFeatureActionTestId(
          props.contract.viewSchemaId,
          "task.status.transition",
        )}
        disabled={
          props.disabled ||
          editor.errors.length > 0 ||
          editor.staleFields.length > 0
        }
        style={buttonStyle}
        type="button"
        onClick={() => void editor.submit()}
      >
        Apply task status
      </button>
    </fieldset>
  );
}
const groupStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
  margin: 0,
  padding: 0,
  border: 0,
};
const labelStyle = { display: "grid", gap: "var(--ct-spacing-xs)" };
const textStyle = { margin: 0, color: "var(--ct-colors-ink-muted)" };
const inputStyle = {
  boxSizing: "border-box" as const,
  minWidth: 0,
  width: "100%",
  padding: "var(--ct-spacing-xs)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  font: "inherit",
};
const buttonStyle = {
  padding: "var(--ct-spacing-xs)",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-component-button-secondary-textColor)",
  font: "inherit",
};
