import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useReducer, useRef } from "react";
import { useContextualCreateAttachment } from "../../features/coordination/useContextualCreateAttachment";
import { useTimelineRelatedEvidenceAttachment } from "../../features/evidence/useTimelineRelatedEvidenceAttachment";
import {
  buildInspectorRelatedRecordDraft,
  type InspectorRelatedRecordWorkflowAction,
  type InspectorRelatedRecordWorkflowState,
  inspectorRelatedRecordWorkflowReducer,
} from "../../inspector/inspectorRelatedRecordModel";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  type WorkbookInspectorLiveSubject,
  workbookInspectorSubjectsEqual,
} from "../../inspector/workbookInspectorSubject";
import { genericCreateMinimumMessage } from "../../models/genericWorkbookModel";
import type { TimelineRelatedRecordPort } from "../../mutations/workbookMutationCommandPorts";
import {
  planTimelineRelatedSubmission,
  type TimelineRelatedActionContext,
  type TimelineRelatedSubmissionPlan,
  type TimelineRelatedWorkflowIdentity,
  timelineRelatedWorkflowIdentity,
  timelineRelatedWorkflowIsCurrent,
} from "../models/timelineRelatedRecordWorkflow";
import type { WorkbookRow } from "../models/timelineRowModel";

type TimelineCreateRelatedWorkflowInput = {
  readonly isInspectorOpen?: boolean;
  readonly actionContext: TimelineRelatedActionContext;
  readonly currentUserId: string | null;
  readonly mutationCommands: TimelineRelatedRecordPort;
  readonly selectedRow: WorkbookRow | null;
  readonly selectedSubject: WorkbookInspectorLiveSubject | null;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly targetContracts: ReadonlyMap<string, ViewContract>;
};

type TimelineRelatedWorkflowRuntime = {
  readonly capturedOwnerSequenceRef: { current: boolean };
  readonly currentWorkflowIdentityRef: {
    current: TimelineRelatedWorkflowIdentity | null;
  };
  readonly dispatchWorkflow: (
    action: InspectorRelatedRecordWorkflowAction,
  ) => InspectorRelatedRecordWorkflowState | null;
  readonly inputRef: { readonly current: TimelineCreateRelatedWorkflowInput };
  readonly workflowRef: {
    readonly current: InspectorRelatedRecordWorkflowState | null;
  };
};

export function useTimelineCreateRelatedWorkflow(
  input: TimelineCreateRelatedWorkflowInput,
) {
  const evidence = useTimelineRelatedEvidenceAttachment(
    input.selectedSubject && input.selectedRow
      ? {
          subject: input.selectedSubject,
          cells: input.selectedRow.rawRow?.cells ?? {},
        }
      : null,
    input.isInspectorOpen ?? true,
  );
  const {
    workflow: contextualWorkflow,
    begin: contextualBegin,
    detach: contextualDetach,
    update: contextualUpdate,
  } = useContextualCreateAttachment(
    input.selectedSubject && input.selectedRow
      ? {
          subject: input.selectedSubject,
          cells: input.selectedRow.rawRow?.cells ?? {},
        }
      : null,
  );
  const [workflow, reactDispatch] = useReducer(
    inspectorRelatedRecordWorkflowReducer,
    null,
  );
  const inputRef = useRef(input);
  const workflowRef = useRef(workflow);
  inputRef.current = input;
  workflowRef.current = workflow;
  const currentWorkflowIdentityRef =
    useRef<TimelineRelatedWorkflowIdentity | null>(null);
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
  const runtime = useMemo<TimelineRelatedWorkflowRuntime>(
    () => ({
      capturedOwnerSequenceRef,
      currentWorkflowIdentityRef,
      dispatchWorkflow,
      inputRef,
      workflowRef,
    }),
    [dispatchWorkflow],
  );

  useEffect(() => {
    const identity = currentWorkflowIdentityRef.current;
    if (
      identity !== null &&
      (identity.surfaceKey !== input.actionContext.surfaceKey ||
        !workbookInspectorSubjectsEqual(
          identity.subject,
          input.selectedSubject,
        ))
    ) {
      currentWorkflowIdentityRef.current = null;
    }
    const active = workflowRef.current;
    if (active !== null) {
      dispatchWorkflow({
        type: "retarget",
        workflowId: active.workflowId,
        subject: input.selectedSubject,
      });
    }
  }, [dispatchWorkflow, input.actionContext.surfaceKey, input.selectedSubject]);

  const cancelWorkflow = useCallback(
    (reason: "owner_action" | "lifecycle" = "owner_action") => {
      contextualDetach();
      if (reason === "owner_action") evidence.detach();
      if (reason === "lifecycle" && capturedOwnerSequenceRef.current) return;
      capturedOwnerSequenceRef.current = false;
      currentWorkflowIdentityRef.current = null;
      const active = workflowRef.current;
      if (active !== null) {
        dispatchWorkflow({ type: "cancel", workflowId: active.workflowId });
      }
    },
    [contextualDetach, evidence.detach, dispatchWorkflow],
  );

  const beginWorkflow = useCallback(
    (featureGroup: InspectorFeatureGroup) => {
      const activeFeature =
        evidence.workflow?.featureGroup.featureGroupKey ??
        contextualWorkflow?.featureGroup.featureGroupKey ??
        workflowRef.current?.featureGroup.featureGroupKey;
      if (activeFeature && activeFeature !== featureGroup.featureGroupKey)
        cancelWorkflow();
      if (contextualBegin(featureGroup)) return;
      if (evidence.begin(featureGroup)) {
        const message = evidence.notice();
        if (message)
          inputRef.current.setInspectorMessage(
            workbookInspectorMessageFeedback(message, "none"),
          );
        return;
      }
      beginTimelineRelatedWorkflow(runtime, featureGroup);
    },
    [
      contextualBegin,
      contextualWorkflow?.featureGroup.featureGroupKey,
      evidence.begin,
      evidence.notice,
      evidence.workflow?.featureGroup.featureGroupKey,
      cancelWorkflow,
      runtime,
    ],
  );
  const updateWorkflowDraft = useCallback(
    (featureGroupKey: string, fieldKey: string, value: string) => {
      if (contextualWorkflow) {
        contextualUpdate(fieldKey, value);
        return;
      }
      const active = workflowRef.current;
      if (
        active !== null &&
        active.featureGroup.featureGroupKey === featureGroupKey
      ) {
        dispatchWorkflow({
          fieldKey,
          type: "update",
          value,
          workflowId: active.workflowId,
        });
      }
    },
    [contextualWorkflow, contextualUpdate, dispatchWorkflow],
  );
  const submitWorkflow = useCallback(async () => {
    const identity = currentWorkflowIdentityRef.current;
    if (identity === null) return;
    const current = inputRef.current;
    const plan = planTimelineRelatedSubmission({
      context: current.actionContext,
      identity,
      selectedRow: current.selectedRow,
      selectedSubject: current.selectedSubject,
      targetContracts: current.targetContracts,
      workflow: workflowRef.current,
    });
    if (plan.kind === "reject") {
      publishTimelineRelatedRejection(current, plan.reason);
      return;
    }
    const submitted = dispatchWorkflow({
      type: "submit",
      workflowId: identity.workflowId,
    });
    if (submitted?.workflowId !== identity.workflowId) return;
    await executeTimelineRelatedSubmission(runtime, plan);
  }, [dispatchWorkflow, runtime]);

  return {
    beginWorkflow,
    cancelWorkflow,
    submitWorkflow,
    updateWorkflowDraft,
    workflow: evidence.workflow ?? contextualWorkflow ?? workflow,
  };
}

function beginTimelineRelatedWorkflow(
  runtime: TimelineRelatedWorkflowRuntime,
  featureGroup: InspectorFeatureGroup,
): void {
  const input = runtime.inputRef.current;
  const targetId =
    featureGroup.routeBinding.kind === "view_row_create" &&
    featureGroup.routeBinding.owner === "view_row_create_route"
      ? featureGroup.routeBinding.targetViewSchemaId
      : undefined;
  const targetContract =
    targetId === undefined ? undefined : input.targetContracts.get(targetId);
  if (!input.actionContext.authorized || targetContract === undefined) {
    input.setInspectorMessage(
      workbookInspectorMessageFeedback(
        "Inspector action is unavailable.",
        "none",
      ),
    );
    return;
  }
  if (input.selectedRow?.recordId == null || input.selectedSubject === null) {
    input.setInspectorMessage(
      workbookInspectorMessageFeedback(
        "Select a row before creating a related record.",
        "none",
      ),
    );
    return;
  }
  const draft = buildInspectorRelatedRecordDraft({
    currentUserId: input.currentUserId,
    featureGroup,
    subject: {
      cells: input.selectedRow.rawRow?.cells ?? {},
      subject: input.selectedSubject,
    },
    targetContract,
  });
  if (draft.kind === "invalid_target") {
    input.setInspectorMessage(
      workbookInspectorMessageFeedback(
        "The target view does not allow row creation.",
        "none",
      ),
    );
    return;
  }
  const workflowId = Symbol("timeline-create-related-workflow");
  const state = runtime.dispatchWorkflow({
    draft: draft.draft,
    featureGroup,
    subject: input.selectedSubject,
    targetContract,
    type: "begin",
    workflowId,
  });
  if (state !== null) {
    runtime.currentWorkflowIdentityRef.current =
      timelineRelatedWorkflowIdentity(state, input.actionContext.surfaceKey);
  }
  runtime.capturedOwnerSequenceRef.current = false;
  input.setInspectorMessage(null);
}

async function executeTimelineRelatedSubmission(
  runtime: TimelineRelatedWorkflowRuntime,
  plan: Extract<TimelineRelatedSubmissionPlan, { kind: "dispatch" }>,
): Promise<void> {
  const created =
    await runtime.inputRef.current.mutationCommands.createRelatedRecord({
      contract: plan.contract,
      draft: plan.draft,
      featureGroupKey: plan.featureGroupKey,
    });
  if (created.kind === "rejected") {
    rejectTimelineRelatedOperation(runtime, plan, created.failure);
    return;
  }
  completeRelatedRecord(runtime, plan, created.value.recordId);
}

function currentRelatedSourceRow(
  runtime: TimelineRelatedWorkflowRuntime,
  identity: TimelineRelatedWorkflowIdentity,
): WorkbookRow | null {
  const input = runtime.inputRef.current;
  return timelineRelatedWorkflowIsCurrent({
    context: input.actionContext,
    identity,
    selectedRow: input.selectedRow,
    selectedSubject: input.selectedSubject,
    workflow: runtime.workflowRef.current,
  })
    ? input.selectedRow
    : null;
}

function completeRelatedRecord(
  runtime: TimelineRelatedWorkflowRuntime,
  plan: Extract<TimelineRelatedSubmissionPlan, { kind: "dispatch" }>,
  createdRecordId: string,
): void {
  if (currentRelatedSourceRow(runtime, plan.identity) === null) return;
  runtime.dispatchWorkflow({
    type: "complete",
    workflowId: plan.identity.workflowId,
  });
  runtime.inputRef.current.setInspectorMessage(
    workbookInspectorMessageFeedback(
      `Created related ${plan.contract.viewSchemaId} row ${createdRecordId}.`,
      "none",
    ),
  );
  runtime.currentWorkflowIdentityRef.current = null;
}

function rejectTimelineRelatedOperation(
  runtime: TimelineRelatedWorkflowRuntime,
  plan: Extract<TimelineRelatedSubmissionPlan, { kind: "dispatch" }>,
  failure: Parameters<typeof workbookInspectorErrorPresentation>[0],
): void {
  if (currentRelatedSourceRow(runtime, plan.identity) === null) return;
  runtime.dispatchWorkflow({
    error:
      failure.kind === "validation"
        ? workbookInspectorLocalErrorPresentation(
            genericCreateMinimumMessage(plan.contract),
          )
        : workbookInspectorErrorPresentation(failure),
    type: "reject",
    workflowId: plan.identity.workflowId,
  });
}

function publishTimelineRelatedRejection(
  input: TimelineCreateRelatedWorkflowInput,
  reason: Extract<TimelineRelatedSubmissionPlan, { kind: "reject" }>["reason"],
): void {
  if (reason === "workflow_unavailable") return;
  input.setInspectorMessage(
    workbookInspectorMessageFeedback(
      reason === "capability_unavailable"
        ? "Inspector action is unavailable."
        : "The selected Timeline row is no longer available.",
      "none",
    ),
  );
}
