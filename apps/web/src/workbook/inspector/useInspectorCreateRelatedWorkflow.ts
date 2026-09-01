import {
  getViewContract,
  type InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useReducer, useRef } from "react";
import { genericCreateMinimumMessage } from "../models/genericWorkbookModel";
import type { TimelineRelatedRecordPort } from "../mutations/workbookMutationCommandPorts";
import {
  buildInspectorRelatedRecordDraft,
  type InspectorRelatedRecordSubjectKey,
  type InspectorRelatedRecordWorkflowAction,
  inspectorRelatedRecordWorkflowReducer,
} from "./inspectorRelatedRecordModel";
import {
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "./workbookInspectorErrorModel";

export type InspectorCreateRelatedSubject = {
  readonly cells: Readonly<Record<string, { readonly value: unknown }>>;
  readonly recordId: string;
  readonly rowVersion: number;
  readonly viewSchemaId: string;
};

export function useInspectorCreateRelatedWorkflow({
  currentUserId,
  mutationCommands,
  onCreated,
  onMessage,
  selectedSubject,
}: {
  readonly currentUserId: string | null;
  readonly mutationCommands: TimelineRelatedRecordPort;
  readonly onCreated: () => Promise<void> | void;
  readonly onMessage: (message: string | null) => void;
  readonly selectedSubject: InspectorCreateRelatedSubject | null;
}) {
  const [workflow, reactDispatch] = useReducer(
    inspectorRelatedRecordWorkflowReducer,
    null,
  );
  const workflowRef = useRef(workflow);
  workflowRef.current = workflow;
  const selectedSubjectRecordId = selectedSubject?.recordId ?? null;
  const selectedSubjectRowVersion = selectedSubject?.rowVersion ?? null;
  const selectedSubjectViewSchemaId = selectedSubject?.viewSchemaId ?? null;
  const selectedSubjectKey = useMemo<InspectorRelatedRecordSubjectKey | null>(
    () =>
      selectedSubjectRecordId === null ||
      selectedSubjectRowVersion === null ||
      selectedSubjectViewSchemaId === null
        ? null
        : {
            recordId: selectedSubjectRecordId,
            rowVersion: selectedSubjectRowVersion,
            viewSchemaId: selectedSubjectViewSchemaId,
          },
    [
      selectedSubjectRecordId,
      selectedSubjectRowVersion,
      selectedSubjectViewSchemaId,
    ],
  );
  const dispatchWorkflow = useCallback(
    (action: InspectorRelatedRecordWorkflowAction) => {
      workflowRef.current = inspectorRelatedRecordWorkflowReducer(
        workflowRef.current,
        action,
      );
      reactDispatch(action);
      return workflowRef.current;
    },
    [],
  );

  useEffect(() => {
    const active = workflowRef.current;
    if (active === null) return;
    dispatchWorkflow({
      type: "retarget",
      workflowId: active.workflowId,
      subjectKey: selectedSubjectKey,
    });
  }, [dispatchWorkflow, selectedSubjectKey]);

  const begin = useCallback(
    (featureGroup: InspectorFeatureGroup): boolean => {
      if (
        featureGroup.routeBinding.kind !== "view_row_create" ||
        featureGroup.routeBinding.owner !== "view_row_create_route" ||
        featureGroup.routeBinding.targetViewSchemaId === undefined
      ) {
        return false;
      }
      const targetContract = getViewContract(
        featureGroup.routeBinding.targetViewSchemaId,
      );
      if (targetContract === undefined) {
        onMessage("The target view does not allow row creation.");
        return true;
      }
      if (selectedSubject === null || selectedSubjectKey === null) {
        onMessage("Select a saved row before creating a related record.");
        return true;
      }
      const result = buildInspectorRelatedRecordDraft({
        currentUserId,
        featureGroup,
        subject: selectedSubject,
        targetContract,
      });
      if (result.kind === "invalid_target") {
        onMessage("The target view does not allow row creation.");
        return true;
      }
      dispatchWorkflow({
        type: "begin",
        draft: result.draft,
        featureGroup,
        subjectKey: selectedSubjectKey,
        targetContract,
        workflowId: Symbol("inspector-create-related-workflow"),
      });
      onMessage(null);
      return true;
    },
    [
      currentUserId,
      dispatchWorkflow,
      onMessage,
      selectedSubject,
      selectedSubjectKey,
    ],
  );

  const updateDraft = useCallback(
    (fieldKey: string, value: string) => {
      const active = workflowRef.current;
      if (active === null) return;
      dispatchWorkflow({
        type: "update",
        fieldKey,
        value,
        workflowId: active.workflowId,
      });
    },
    [dispatchWorkflow],
  );

  const cancel = useCallback(() => {
    const active = workflowRef.current;
    if (active === null) return;
    dispatchWorkflow({ type: "cancel", workflowId: active.workflowId });
  }, [dispatchWorkflow]);

  const submit = useCallback(async () => {
    const active = workflowRef.current;
    if (active === null || active.phase !== "editing") return;
    const submitted = dispatchWorkflow({
      type: "submit",
      workflowId: active.workflowId,
    });
    if (submitted?.workflowId !== active.workflowId) return;
    const outcome = await mutationCommands.createRelatedRecord({
      contract: active.targetContract,
      draft: active.draft,
      featureGroupKey: active.featureGroup.featureGroupKey,
    });
    if (outcome.kind === "rejected") {
      dispatchWorkflow({
        type: "reject",
        workflowId: active.workflowId,
        error:
          outcome.failure.kind === "validation"
            ? workbookInspectorLocalErrorPresentation(
                genericCreateMinimumMessage(active.targetContract),
              )
            : workbookInspectorErrorPresentation(outcome.failure),
      });
      return;
    }
    if (workflowRef.current?.workflowId === active.workflowId) {
      dispatchWorkflow({
        type: "complete",
        workflowId: active.workflowId,
      });
      onMessage(
        `Created ${active.targetContract.title} record ${outcome.value.recordId}.`,
      );
    }
    await onCreated();
  }, [dispatchWorkflow, mutationCommands, onCreated, onMessage]);

  return {
    commands: { begin, cancel, submit, updateDraft },
    snapshot: { workflow },
  };
}
