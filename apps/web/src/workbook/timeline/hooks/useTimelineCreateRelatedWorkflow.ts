import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useReducer, useRef } from "react";
import {
  buildInspectorRelatedRecordDraft,
  type InspectorRelatedRecordWorkflowAction,
  inspectorRelatedRecordWorkflowReducer,
} from "../../inspector/inspectorRelatedRecordModel";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  updateWorkbookInspectorSubject,
  type WorkbookInspectorLiveSubject,
  workbookInspectorSubjectsEqual,
} from "../../inspector/workbookInspectorSubject";
import { genericCreateMinimumMessage } from "../../models/genericWorkbookModel";
import { evidenceViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type {
  TimelineRelatedEvidenceLinked,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookRow } from "../models/workbookTimelineModel";

export function useTimelineCreateRelatedWorkflow({
  applyAcceptedRowMutation,
  currentUserId,
  loadRows,
  mutationCommands,
  selectedRow,
  selectedSubject,
  setInspectorMessage,
  targetContracts,
}: {
  readonly applyAcceptedRowMutation: (
    rowKey: string,
    accepted: Pick<TimelineRelatedEvidenceLinked, "row" | "viewSchemaId">,
  ) => unknown;
  readonly currentUserId: string | null;
  readonly loadRows: (options: {
    readonly showLoading: boolean;
  }) => Promise<void>;
  readonly mutationCommands: TimelineRelatedRecordPort;
  readonly selectedRow: WorkbookRow | null;
  readonly selectedSubject: WorkbookInspectorLiveSubject | null;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly targetContracts: ReadonlyMap<string, ViewContract>;
}) {
  const [workflow, reactDispatch] = useReducer(
    inspectorRelatedRecordWorkflowReducer,
    null,
  );
  const workflowRef = useRef(workflow);
  workflowRef.current = workflow;
  const currentWorkflowIdentityRef =
    useRef<CurrentTimelineRelatedWorkflowIdentity | null>(null);
  const capturedOwnerSequenceRef = useRef(false);
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
    const currentIdentity = currentWorkflowIdentityRef.current;
    if (
      currentIdentity !== null &&
      !workbookInspectorSubjectsEqual(currentIdentity.subject, selectedSubject)
    ) {
      currentWorkflowIdentityRef.current = null;
    }
    const active = workflowRef.current;
    if (active === null) return;
    dispatchWorkflow({
      type: "retarget",
      workflowId: active.workflowId,
      subject: selectedSubject,
    });
  }, [dispatchWorkflow, selectedSubject]);

  const cancelWorkflow = useCallback(
    (reason: "owner_action" | "lifecycle" = "owner_action") => {
      if (reason === "lifecycle" && capturedOwnerSequenceRef.current) return;
      capturedOwnerSequenceRef.current = false;
      currentWorkflowIdentityRef.current = null;
      const active = workflowRef.current;
      if (active === null) return;
      dispatchWorkflow({ type: "cancel", workflowId: active.workflowId });
    },
    [dispatchWorkflow],
  );

  const beginWorkflow = useCallback(
    (featureGroup: InspectorFeatureGroup) => {
      if (
        featureGroup.routeBinding.kind !== "view_row_create" ||
        featureGroup.routeBinding.owner !== "view_row_create_route"
      ) {
        setInspectorMessage(
          workbookInspectorMessageFeedback(
            "Inspector action is unavailable.",
            "none",
          ),
        );
        return;
      }
      const targetViewSchemaId = featureGroup.routeBinding.targetViewSchemaId;
      const targetContract =
        targetViewSchemaId === undefined
          ? undefined
          : targetContracts.get(targetViewSchemaId);
      if (targetContract === undefined) {
        setInspectorMessage(
          workbookInspectorMessageFeedback(
            "Inspector action is unavailable.",
            "none",
          ),
        );
        return;
      }
      if (selectedRow?.recordId == null || selectedSubject === null) {
        setInspectorMessage(
          workbookInspectorMessageFeedback(
            "Select a row before creating a related record.",
            "none",
          ),
        );
        return;
      }
      const result = buildInspectorRelatedRecordDraft({
        currentUserId,
        featureGroup,
        subject: {
          cells: selectedRow.rawRow?.cells ?? {},
          subject: selectedSubject,
        },
        targetContract,
      });
      if (result.kind === "invalid_target") {
        setInspectorMessage(
          workbookInspectorMessageFeedback(
            "The target view does not allow row creation.",
            "none",
          ),
        );
        return;
      }
      const workflowId = Symbol("timeline-create-related-workflow");
      dispatchWorkflow({
        type: "begin",
        featureGroup,
        targetContract,
        subject: selectedSubject,
        draft: result.draft,
        workflowId,
      });
      currentWorkflowIdentityRef.current = {
        subject: selectedSubject,
        workflowId,
      };
      capturedOwnerSequenceRef.current = false;
      setInspectorMessage(null);
    },
    [
      currentUserId,
      dispatchWorkflow,
      selectedRow,
      selectedSubject,
      setInspectorMessage,
      targetContracts,
    ],
  );

  const updateWorkflowDraft = useCallback(
    (featureGroupKey: string, fieldKey: string, value: string) => {
      const active = workflowRef.current;
      if (
        active === null ||
        active.featureGroup.featureGroupKey !== featureGroupKey
      ) {
        return;
      }
      dispatchWorkflow({
        type: "update",
        fieldKey,
        value,
        workflowId: active.workflowId,
      });
    },
    [dispatchWorkflow],
  );

  const submitWorkflow = useCallback(async () => {
    const activeWorkflow = workflowRef.current;
    const sourceRow = selectedRow;
    if (
      activeWorkflow === null ||
      activeWorkflow.phase !== "editing" ||
      sourceRow === null ||
      sourceRow.recordId === null ||
      sourceRow.recordId !== activeWorkflow.subject.recordId ||
      sourceRow.rowVersion !== activeWorkflow.subject.rowVersion
    ) {
      return;
    }
    const submitted = dispatchWorkflow({
      type: "submit",
      workflowId: activeWorkflow.workflowId,
    });
    if (submitted?.workflowId !== activeWorkflow.workflowId) return;
    const createResult = await mutationCommands.createRelatedRecord({
      contract: activeWorkflow.targetContract,
      draft: activeWorkflow.draft,
      featureGroupKey: activeWorkflow.featureGroup.featureGroupKey,
    });
    if (createResult.kind === "rejected") {
      dispatchWorkflow({
        type: "reject",
        workflowId: activeWorkflow.workflowId,
        error:
          createResult.failure.kind === "validation"
            ? workbookInspectorLocalErrorPresentation(
                genericCreateMinimumMessage(activeWorkflow.targetContract),
              )
            : workbookInspectorErrorPresentation(createResult.failure),
      });
      return;
    }
    const createdRecordId = createResult.value.recordId;

    if (activeWorkflow.targetContract.viewSchemaId === evidenceViewSchemaId) {
      const patchResult = await mutationCommands.linkCreatedEvidence({
        sourceRow,
        createdRecordId,
      });
      if (patchResult.kind === "rejected") {
        dispatchWorkflow({
          type: "reject",
          workflowId: activeWorkflow.workflowId,
          error: workbookInspectorErrorPresentation(patchResult.failure),
        });
        return;
      }
      if (
        currentWorkflowIdentityRef.current?.workflowId ===
          activeWorkflow.workflowId &&
        workflowRef.current?.workflowId === activeWorkflow.workflowId
      ) {
        dispatchWorkflow({
          type: "complete",
          workflowId: activeWorkflow.workflowId,
        });
        const acceptedSubject = updateWorkbookInspectorSubject(
          activeWorkflow.subject,
          {
            kind: "live",
            recordId: patchResult.value.row.record_id,
            rowVersion: patchResult.value.row.row_version,
          },
        );
        currentWorkflowIdentityRef.current =
          acceptedSubject?.kind === "live"
            ? {
                subject: acceptedSubject,
                workflowId: activeWorkflow.workflowId,
              }
            : null;
        capturedOwnerSequenceRef.current = true;
      }
      applyAcceptedRowMutation(sourceRow.key, patchResult.value);
      await loadRows({ showLoading: false });
      if (
        currentWorkflowIdentityRef.current?.workflowId ===
        activeWorkflow.workflowId
      ) {
        setInspectorMessage(
          workbookInspectorMessageFeedback(
            `Created and linked evidence ${createdRecordId}.`,
            "none",
          ),
        );
        currentWorkflowIdentityRef.current = null;
      }
      capturedOwnerSequenceRef.current = false;
      return;
    }

    if (
      currentWorkflowIdentityRef.current?.workflowId ===
        activeWorkflow.workflowId &&
      workflowRef.current?.workflowId === activeWorkflow.workflowId
    ) {
      dispatchWorkflow({
        type: "complete",
        workflowId: activeWorkflow.workflowId,
      });
      setInspectorMessage(
        workbookInspectorMessageFeedback(
          `Created related ${activeWorkflow.targetContract.viewSchemaId} row ${createdRecordId}.`,
          "none",
        ),
      );
      currentWorkflowIdentityRef.current = null;
    }
  }, [
    applyAcceptedRowMutation,
    dispatchWorkflow,
    loadRows,
    mutationCommands,
    selectedRow,
    setInspectorMessage,
  ]);

  return {
    beginWorkflow,
    cancelWorkflow,
    submitWorkflow,
    updateWorkflowDraft,
    workflow,
  };
}

type CurrentTimelineRelatedWorkflowIdentity = {
  readonly subject: WorkbookInspectorLiveSubject;
  readonly workflowId: symbol;
};
