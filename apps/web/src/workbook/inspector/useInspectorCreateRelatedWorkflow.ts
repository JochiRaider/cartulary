import {
  getViewContract,
  type InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useReducer, useRef } from "react";
import { genericCreateMinimumMessage } from "../models/genericWorkbookModel";
import type { TimelineRelatedRecordPort } from "../mutations/workbookMutationCommandPorts";
import {
  buildInspectorRelatedRecordDraft,
  type InspectorRelatedRecordWorkflowAction,
  inspectorRelatedRecordWorkflowReducer,
} from "./inspectorRelatedRecordModel";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorFeedback,
  workbookInspectorLocalErrorPresentation,
  workbookInspectorMessageFeedback,
} from "./workbookInspectorErrorModel";
import type { WorkbookInspectorLiveRowBinding } from "./workbookInspectorSubject";

export function useInspectorCreateRelatedWorkflow({
  beginMutation,
  currentUserId,
  mutationCommands,
  onCreated,
  onFeedback,
  selectedSubject,
}: {
  readonly beginMutation: () => () => void;
  readonly currentUserId: string | null;
  readonly mutationCommands: TimelineRelatedRecordPort;
  readonly onCreated: () => Promise<void> | void;
  readonly onFeedback: (feedback: WorkbookInspectorFeedback | null) => void;
  readonly selectedSubject: WorkbookInspectorLiveRowBinding | null;
}) {
  const [workflow, reactDispatch] = useReducer(
    inspectorRelatedRecordWorkflowReducer,
    null,
  );
  const workflowRef = useRef(workflow);
  workflowRef.current = workflow;
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
      subject: selectedSubject?.subject ?? null,
    });
  }, [dispatchWorkflow, selectedSubject]);

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
        onFeedback(
          workbookInspectorLocalErrorFeedback(
            "The target view does not allow row creation.",
          ),
        );
        return true;
      }
      if (selectedSubject === null) {
        onFeedback(
          workbookInspectorLocalErrorFeedback(
            "Select a saved row before creating a related record.",
          ),
        );
        return true;
      }
      const result = buildInspectorRelatedRecordDraft({
        currentUserId,
        featureGroup,
        subject: selectedSubject,
        targetContract,
      });
      if (result.kind === "invalid_target") {
        onFeedback(
          workbookInspectorLocalErrorFeedback(
            "The target view does not allow row creation.",
          ),
        );
        return true;
      }
      dispatchWorkflow({
        type: "begin",
        draft: result.draft,
        featureGroup,
        subject: selectedSubject.subject,
        targetContract,
        workflowId: Symbol("inspector-create-related-workflow"),
      });
      onFeedback(null);
      return true;
    },
    [currentUserId, dispatchWorkflow, onFeedback, selectedSubject],
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
    const finish = beginMutation();
    try {
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
        onFeedback(
          workbookInspectorMessageFeedback(
            `Created ${active.targetContract.title} record ${outcome.value.recordId}.`,
            "none",
          ),
        );
      }
      await onCreated();
    } finally {
      finish();
    }
  }, [
    beginMutation,
    dispatchWorkflow,
    mutationCommands,
    onCreated,
    onFeedback,
  ]);

  return {
    commands: { begin, cancel, submit, updateDraft },
    snapshot: { workflow },
  };
}
