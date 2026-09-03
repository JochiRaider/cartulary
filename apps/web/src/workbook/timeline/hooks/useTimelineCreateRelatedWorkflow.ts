import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useReducer, useRef } from "react";
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
  readonly actionContext: TimelineRelatedActionContext;
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
      if (reason === "lifecycle" && capturedOwnerSequenceRef.current) return;
      capturedOwnerSequenceRef.current = false;
      currentWorkflowIdentityRef.current = null;
      const active = workflowRef.current;
      if (active !== null) {
        dispatchWorkflow({ type: "cancel", workflowId: active.workflowId });
      }
    },
    [dispatchWorkflow],
  );

  const beginWorkflow = useCallback(
    (featureGroup: InspectorFeatureGroup) =>
      beginTimelineRelatedWorkflow(runtime, featureGroup),
    [runtime],
  );
  const updateWorkflowDraft = useCallback(
    (featureGroupKey: string, fieldKey: string, value: string) => {
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
    [dispatchWorkflow],
  );
  const submitWorkflow = useCallback(async () => {
    const identity = currentWorkflowIdentityRef.current;
    if (identity === null) return;
    const current = inputRef.current;
    const plan = planTimelineRelatedSubmission({
      context: current.actionContext,
      evidenceViewSchemaId,
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
    workflow,
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
  if (!plan.evidenceLinkRequired) {
    completeRelatedRecord(runtime, plan, created.value.recordId);
    return;
  }
  const sourceRow = currentRelatedSourceRow(runtime, plan.identity);
  if (sourceRow === null) return;
  const linked =
    await runtime.inputRef.current.mutationCommands.linkCreatedEvidence({
      createdRecordId: created.value.recordId,
      sourceRow,
    });
  if (linked.kind === "rejected") {
    rejectTimelineRelatedOperation(runtime, plan, linked.failure);
    return;
  }
  await completeRelatedEvidence(
    runtime,
    plan,
    sourceRow,
    created.value.recordId,
    linked.value,
  );
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

async function completeRelatedEvidence(
  runtime: TimelineRelatedWorkflowRuntime,
  plan: Extract<TimelineRelatedSubmissionPlan, { kind: "dispatch" }>,
  sourceRow: WorkbookRow,
  createdRecordId: string,
  linked: TimelineRelatedEvidenceLinked,
): Promise<void> {
  if (currentRelatedSourceRow(runtime, plan.identity) === null) return;
  runtime.dispatchWorkflow({
    type: "complete",
    workflowId: plan.identity.workflowId,
  });
  const acceptedSubject = updateWorkbookInspectorSubject(
    plan.identity.subject,
    {
      kind: "live",
      recordId: linked.row.record_id,
      rowVersion: linked.row.row_version,
    },
  );
  runtime.currentWorkflowIdentityRef.current =
    acceptedSubject?.kind === "live"
      ? { ...plan.identity, subject: acceptedSubject }
      : null;
  runtime.capturedOwnerSequenceRef.current = true;
  const input = runtime.inputRef.current;
  input.applyAcceptedRowMutation(sourceRow.key, linked);
  try {
    await input.loadRows({ showLoading: false });
    const currentIdentity = runtime.currentWorkflowIdentityRef.current;
    const currentInput = runtime.inputRef.current;
    if (
      currentIdentity?.workflowId === plan.identity.workflowId &&
      currentIdentity.surfaceKey === currentInput.actionContext.surfaceKey &&
      currentInput.actionContext.authorized
    ) {
      currentInput.setInspectorMessage(
        workbookInspectorMessageFeedback(
          `Created and linked evidence ${createdRecordId}.`,
          "none",
        ),
      );
      runtime.currentWorkflowIdentityRef.current = null;
    }
  } finally {
    runtime.capturedOwnerSequenceRef.current = false;
  }
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
