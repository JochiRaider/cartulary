import { useContext, useSyncExternalStore } from "react";
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

export function InspectorCreateRelatedWorkflow({
  state,
}: {
  readonly state: InspectorRelatedRecordWorkflowState;
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
  return null;
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
