import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, useContext, useSyncExternalStore } from "react";
import { GenericMutationControl } from "../components/GenericMutationControl";
import { ContextualCreateContext } from "../features/coordination/ContextualCreateContext";
import { ContextualCreateForm } from "../features/coordination/ContextualCreateForm";
import { CoordinationCreateContext } from "../features/coordination/CoordinationCreateContext";
import { CoordinationCreateForm } from "../features/coordination/CoordinationCreateForm";
import { isContextualCreateFeature } from "../features/coordination/contextualCreateModel";
import { coordinationVariant } from "../features/coordination/coordinationCreateModel";
import { TimelineRelatedEvidenceContext } from "../features/evidence/TimelineRelatedEvidenceContext";
import { TimelineRelatedEvidenceForm } from "../features/evidence/TimelineRelatedEvidenceForm";
import { relatedEvidenceFeature } from "../features/evidence/timelineRelatedEvidenceModel";
import { NoteCreateContext } from "../features/notes/NoteCreateContext";
import { NoteCreateForm } from "../features/notes/NoteCreateForm";
import { noteCreateFeature } from "../features/notes/noteCreateModel";
import type { InspectorRelatedRecordWorkflowState } from "./inspectorRelatedRecordModel";
import { WorkbookInspectorActionButton } from "./presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "./presentation/WorkbookInspectorFeedback";

export function InspectorCreateRelatedWorkflow({
  state,
  onCancel,
  onSubmit,
  onUpdateDraft,
}: {
  readonly state: InspectorRelatedRecordWorkflowState;
  readonly onCancel: () => void;
  readonly onSubmit: () => void;
  readonly onUpdateDraft: (fieldKey: string, value: string) => void;
}) {
  const note = useContext(NoteCreateContext);
  const coordination = useContext(CoordinationCreateContext);
  const context = useContext(ContextualCreateContext);
  const evidence = useContext(TimelineRelatedEvidenceContext);
  if (coordinationVariant(state.featureGroup.featureGroupKey))
    return coordination ? (
      <CoordinationCreateForm
        owner={coordination.owner}
        attachment={state.workflowId}
        onSubmit={() => void coordination.owner.submit(state.workflowId)}
      />
    ) : null;
  if (state.featureGroup.featureGroupKey === noteCreateFeature)
    return note ? (
      <NoteCreateForm
        owner={note.owner}
        attachment={state.workflowId}
        onSubmit={() => void note.owner.submit(state.workflowId)}
      />
    ) : null;
  if (state.featureGroup.featureGroupKey === relatedEvidenceFeature)
    return evidence ? (
      <TimelineRelatedEvidenceForm
        owner={evidence.owner}
        attachment={state.workflowId}
        onSubmit={() => void evidence.owner.submit(state.workflowId)}
        onReview={() => void evidence.owner.review()}
      />
    ) : null;
  if (isContextualCreateFeature(state.featureGroup.featureGroupKey))
    return context ? (
      <RetainedContextualForm
        owner={context.owner}
        attachment={state.workflowId}
      />
    ) : null;
  const createFields = state.targetContract.fields.filter(
    (field) => field.createWritable,
  );
  return (
    <section aria-label={state.featureGroup.label} style={workflowStyle}>
      <p style={messageStyle}>Create in {state.targetContract.title}</p>
      {createFields.map((field) => {
        const controlId = `inspector-related-${field.fieldKey}`;
        return (
          <label htmlFor={controlId} key={field.fieldKey} style={labelStyle}>
            {field.label}
            <GenericMutationControl
              collectionMode="add"
              field={field}
              id={controlId}
              testId={genericCreateFieldTestId(field.fieldKey)}
              value={state.draft[field.fieldKey] ?? ""}
              onChange={(value) => onUpdateDraft(field.fieldKey, value)}
            />
          </label>
        );
      })}
      {state.error === null ? null : (
        <WorkbookInspectorPublicError error={state.error} />
      )}
      <div style={actionsStyle}>
        <WorkbookInspectorActionButton
          data-testid={genericCreateSubmitTestId(
            state.targetContract.viewSchemaId,
          )}
          disabled={state.phase === "submitting"}
          tone="primary"
          onClick={onSubmit}
        >
          Create related row
        </WorkbookInspectorActionButton>
        <WorkbookInspectorActionButton
          disabled={state.phase === "submitting"}
          tone="secondary"
          onClick={onCancel}
        >
          Cancel
        </WorkbookInspectorActionButton>
      </div>
    </section>
  );
}

function RetainedContextualForm({
  owner,
  attachment,
}: {
  readonly owner: NonNullable<
    React.ContextType<typeof ContextualCreateContext>
  >["owner"];
  readonly attachment: symbol;
}) {
  useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  return (
    <ContextualCreateForm
      owner={owner}
      attachment={attachment}
      disabled={owner.busy}
      onSubmit={() => void owner.submit(attachment)}
    />
  );
}

const workflowStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  paddingBlock: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const labelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const messageStyle = { margin: 0 } satisfies CSSProperties;

const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
