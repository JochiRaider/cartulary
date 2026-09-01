import { useCallback, useMemo } from "react";
import { sheetRefKey } from "../../../shared/sheetRef";
import { useIncidentMemberReferenceOptions } from "../../hooks/useOwnerReferenceOptions";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import { useTimelineCreateRelatedWorkflow } from "../hooks/useTimelineCreateRelatedWorkflow";
import { useTimelineEvidenceAttach } from "../hooks/useTimelineEvidenceAttach";
import { useTimelineHistoryActions } from "../hooks/useTimelineHistoryActions";
import { useTimelineInspectorFeatureController } from "../hooks/useTimelineInspectorFeatureController";
import {
  useTimelineInspectorEscape,
  useTimelineInspectorLifecycle,
  useTimelineInspectorRowInteractions,
} from "../hooks/useTimelineInspectorSelection";
import { useTimelineMentionActions } from "../hooks/useTimelineMentionActions";
import { timelineCreateRelatedTargetContracts } from "../models/timelineWorkbookFeaturePolicy";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";

type CreateRelatedInput = Parameters<
  typeof useTimelineCreateRelatedWorkflow
>[0];
type EvidenceInput = Parameters<typeof useTimelineEvidenceAttach>[0];
type HistoryInput = Parameters<typeof useTimelineHistoryActions>[0];
type InspectorLifecycleInput = Parameters<
  typeof useTimelineInspectorLifecycle
>[0];
type InspectorRowInteractionsInput = Parameters<
  typeof useTimelineInspectorRowInteractions
>[0];
type MentionInput = Parameters<typeof useTimelineMentionActions>[0];

type TimelineInspectorWorkflowCompositionInput = {
  readonly foundation: {
    readonly evidenceAttachmentPort: EvidenceInput["evidenceAttachmentPort"];
    readonly historyPort: HistoryInput["historyPort"];
    readonly loadAccessLost: boolean;
    readonly mentionPort: MentionInput["mentionPort"];
    readonly rows: InspectorLifecycleInput["rows"];
    readonly rowsRef: MentionInput["rowsRef"];
    readonly selectedMentionRef: InspectorLifecycleInput["selectedMentionRef"];
    readonly setDismissedMentionsByRow: MentionInput["setDismissedMentionsByRow"];
    readonly setSelectedMentionRef: InspectorLifecycleInput["setSelectedMentionRef"];
    readonly setSelectedResolveTargetId: InspectorLifecycleInput["setSelectedResolveTargetId"];
  };
  readonly grid: {
    readonly beginViewportContinuity: HistoryInput["beginViewportContinuity"];
    readonly clearViewportContinuity: HistoryInput["clearViewportContinuity"];
    readonly gridShellRef: InspectorLifecycleInput["gridShellRef"];
    readonly requireViewportContinuitySourceRecord: MentionInput["requireViewportContinuitySourceRecord"];
    readonly restoreTimelineFocusAnchor: InspectorLifecycleInput["restoreTimelineFocusAnchor"];
    readonly settleViewportContinuityFollowUp: MentionInput["settleViewportContinuityFollowUp"];
    readonly workbookFocusAnchorRef: InspectorLifecycleInput["workbookFocusAnchorRef"];
  };
  readonly incident: Pick<
    TimelineWorkbookSurfaceRuntime["incident"],
    | "continuityResetKey"
    | "currentRole"
    | "currentUserId"
    | "incidentPort"
    | "inspectorResetKey"
  >;
  readonly inspector: {
    readonly history: {
      readonly commands: Pick<
        HistoryInput,
        | "beginRowHistoryRequest"
        | "currentHistoryRecordIdMatches"
        | "rowHistoryRequestIsCurrent"
        | "setRowHistory"
        | "setRowHistoryPendingAction"
      > &
        Pick<
          InspectorLifecycleInput,
          "cancelRowHistoryRequests" | "clearRowHistory"
        >;
      readonly snapshot: Pick<
        HistoryInput,
        | "activeHistoryLiveRecordId"
        | "currentHistoryRecordId"
        | "currentHistoryRowVersion"
        | "rowHistory"
        | "rowHistoryPendingAction"
      >;
    };
    readonly lifecycle: {
      readonly invalidationCause: InspectorLifecycleInput["inspectorInvalidationCause"];
      readonly invalidationGeneration: number;
      readonly isOpen: boolean;
    };
    readonly selection: {
      readonly inspectorMentions: InspectorLifecycleInput["inspectorMentions"];
      readonly selectedRow: CreateRelatedInput["selectedRow"];
      readonly selectedRowId: InspectorLifecycleInput["selectedRowId"];
      readonly selectedRowWorkflowKey: string;
    };
    readonly publishFeedback: CreateRelatedInput["setInspectorMessage"];
    readonly selectRow: InspectorLifecycleInput["setSelectedRowId"];
    readonly setOpen: InspectorLifecycleInput["setIsInspectorOpen"];
  };
  readonly mutation: {
    readonly activeConflict: Parameters<
      typeof useTimelineInspectorEscape
    >[0]["activeConflict"];
    readonly applyAcceptedRowMutation: EvidenceInput["applyAcceptedRowMutation"];
    readonly commands: Pick<
      HistoryInput,
      | "acceptTimelineRecordVersion"
      | "beginSave"
      | "enqueueSaveWork"
      | "finishSave"
      | "nextClientTxnId"
      | "resolvePendingSocketTxn"
      | "trackPendingSocketTxn"
    >;
    readonly waitForCommittedRecordIdle: EvidenceInput["waitForCommittedRecordIdle"];
    readonly loadRows: HistoryInput["loadRows"];
    readonly publishViewingPresence: InspectorRowInteractionsInput["publishViewingPresence"];
  };
  readonly mutationCommands: TimelineWorkbookSurfaceRuntime["mutationCommands"];
  readonly onIncidentAccessLost: TimelineWorkbookSurfaceRuntime["onIncidentAccessLost"];
  readonly onRefreshEntities: TimelineWorkbookSurfaceRuntime["entities"]["refresh"];
  readonly activeSheetRef: TimelineWorkbookSurfaceRuntime["incident"]["sheetRef"];
};

export function useTimelineInspectorWorkflowComposition({
  activeSheetRef,
  foundation,
  grid,
  incident,
  inspector,
  mutation,
  mutationCommands,
  onIncidentAccessLost,
  onRefreshEntities,
}: TimelineInspectorWorkflowCompositionInput) {
  const {
    beginWorkflow,
    cancelWorkflow,
    submitWorkflow,
    updateWorkflowDraft,
    workflow: createRelatedWorkflow,
  } = useTimelineCreateRelatedWorkflow({
    applyAcceptedRowMutation: mutation.applyAcceptedRowMutation,
    currentUserId: incident.currentUserId,
    loadRows: mutation.loadRows,
    mutationCommands: mutationCommands.related,
    selectedRow: inspector.selection.selectedRow,
    selectedRowWorkflowKey: inspector.selection.selectedRowWorkflowKey,
    setInspectorMessage: inspector.publishFeedback,
    targetContracts: timelineCreateRelatedTargetContracts,
  });
  const features = useTimelineInspectorFeatureController({
    beginCreateRelatedWorkflow: beginWorkflow,
    cancelCreateRelatedWorkflow: cancelWorkflow,
    lifecycle: {
      authorizationKey: `${incident.currentRole ?? "none"}:${foundation.loadAccessLost}`,
      invalidationGeneration: inspector.lifecycle.invalidationGeneration,
      lifecycleKey: `${incident.inspectorResetKey}:${incident.continuityResetKey}`,
      subjectKey: inspector.selection.selectedRowWorkflowKey,
      surfaceKey: sheetRefKey(activeSheetRef),
    },
    setInspectorMessage: inspector.publishFeedback,
  });
  const createRelatedNeedsIncidentMembers =
    createRelatedWorkflow?.targetContract.fields.some(
      (field) =>
        field.directReferenceContractId === "incident_member_user_ref_v1",
    ) ?? false;
  const { options: incidentMemberOptions } = useIncidentMemberReferenceOptions({
    enabled: createRelatedNeedsIncidentMembers,
    incidentPort: incident.incidentPort,
    onIncidentAccessLost,
  });
  const createRelatedReferenceOptions = useMemo(
    () => ({
      ...emptyGenericReferenceOptions(),
      incidentMembers: incidentMemberOptions,
    }),
    [incidentMemberOptions],
  );
  const rowInteractions = useTimelineInspectorRowInteractions({
    publishViewingPresence: mutation.publishViewingPresence,
    rows: foundation.rows,
    rowsRef: foundation.rowsRef,
    selectedRowId: inspector.selection.selectedRowId,
    setInspectorMessage: inspector.publishFeedback,
    setIsInspectorOpen: inspector.setOpen,
    setSelectedMentionRef: foundation.setSelectedMentionRef,
    setSelectedRowId: inspector.selectRow,
  });
  const close = useTimelineInspectorLifecycle({
    cancelRowHistoryRequests:
      inspector.history.commands.cancelRowHistoryRequests,
    clearRowHistory: inspector.history.commands.clearRowHistory,
    gridShellRef: grid.gridShellRef,
    inspectorInvalidationCause: inspector.lifecycle.invalidationCause,
    inspectorMentions: inspector.selection.inspectorMentions,
    inspectorInvalidationGeneration: inspector.lifecycle.invalidationGeneration,
    restoreTimelineFocusAnchor: grid.restoreTimelineFocusAnchor,
    rowHistory: inspector.history.snapshot.rowHistory,
    rows: foundation.rows,
    selectedMentionRef: foundation.selectedMentionRef,
    selectedRowId: inspector.selection.selectedRowId,
    setInspectorMessage: inspector.publishFeedback,
    setIsInspectorOpen: inspector.setOpen,
    setRowHistory: inspector.history.commands.setRowHistory,
    setRowHistoryPendingAction:
      inspector.history.commands.setRowHistoryPendingAction,
    setSelectedMentionRef: foundation.setSelectedMentionRef,
    setSelectedResolveTargetId: foundation.setSelectedResolveTargetId,
    setSelectedRowId: inspector.selectRow,
    workbookFocusAnchorRef: grid.workbookFocusAnchorRef,
  });
  const history = useTimelineHistoryActions({
    acceptTimelineRecordVersion: mutation.commands.acceptTimelineRecordVersion,
    activeHistoryLiveRecordId:
      inspector.history.snapshot.activeHistoryLiveRecordId,
    beginRowHistoryRequest: inspector.history.commands.beginRowHistoryRequest,
    beginSave: mutation.commands.beginSave,
    beginViewportContinuity: grid.beginViewportContinuity,
    clearViewportContinuity: grid.clearViewportContinuity,
    currentHistoryRecordId: inspector.history.snapshot.currentHistoryRecordId,
    currentHistoryRecordIdMatches:
      inspector.history.commands.currentHistoryRecordIdMatches,
    currentHistoryRowVersion:
      inspector.history.snapshot.currentHistoryRowVersion,
    enqueueSaveWork: mutation.commands.enqueueSaveWork,
    finishSave: mutation.commands.finishSave,
    historyPort: foundation.historyPort,
    loadRows: mutation.loadRows,
    nextClientTxnId: mutation.commands.nextClientTxnId,
    resolvePendingSocketTxn: mutation.commands.resolvePendingSocketTxn,
    rowHistory: inspector.history.snapshot.rowHistory,
    rowHistoryPendingAction: inspector.history.snapshot.rowHistoryPendingAction,
    rowHistoryRequestIsCurrent:
      inspector.history.commands.rowHistoryRequestIsCurrent,
    selectedRowRecordId: inspector.selection.selectedRow?.recordId ?? null,
    setIsInspectorOpen: inspector.setOpen,
    setRowHistory: inspector.history.commands.setRowHistory,
    setRowHistoryPendingAction:
      inspector.history.commands.setRowHistoryPendingAction,
    setSelectedRowId: inspector.selectRow,
    trackPendingSocketTxn: mutation.commands.trackPendingSocketTxn,
    waitForCommittedRecordIdle: mutation.waitForCommittedRecordIdle,
  });
  const mentions = useTimelineMentionActions({
    beginSave: mutation.commands.beginSave,
    beginViewportContinuity: grid.beginViewportContinuity,
    clearViewportContinuity: grid.clearViewportContinuity,
    enqueueSaveWork: mutation.commands.enqueueSaveWork,
    finishSave: mutation.commands.finishSave,
    loadRows: mutation.loadRows,
    mentionPort: foundation.mentionPort,
    nextClientTxnId: mutation.commands.nextClientTxnId,
    onRefreshEntities,
    requireViewportContinuitySourceRecord:
      grid.requireViewportContinuitySourceRecord,
    resolvePendingSocketTxn: mutation.commands.resolvePendingSocketTxn,
    rowsRef: foundation.rowsRef,
    setDismissedMentionsByRow: foundation.setDismissedMentionsByRow,
    setInspectorMessage: inspector.publishFeedback,
    settleViewportContinuityFollowUp: grid.settleViewportContinuityFollowUp,
    trackPendingSocketTxn: mutation.commands.trackPendingSocketTxn,
    waitForCommittedRecordIdle: mutation.waitForCommittedRecordIdle,
  });
  const evidence = useTimelineEvidenceAttach({
    applyAcceptedRowMutation: mutation.applyAcceptedRowMutation,
    beginSave: mutation.commands.beginSave,
    beginViewportContinuity: grid.beginViewportContinuity,
    clearViewportContinuity: grid.clearViewportContinuity,
    enqueueSaveWork: mutation.commands.enqueueSaveWork,
    evidenceAttachmentPort: foundation.evidenceAttachmentPort,
    finishSave: mutation.commands.finishSave,
    resolvePendingSocketTxn: mutation.commands.resolvePendingSocketTxn,
    rowsRef: foundation.rowsRef,
    setInspectorMessage: inspector.publishFeedback,
    trackPendingSocketTxn: mutation.commands.trackPendingSocketTxn,
    waitForCommittedRecordIdle: mutation.waitForCommittedRecordIdle,
  });
  const handleResolveTargetChange = useCallback(
    (value: string) => {
      foundation.setSelectedResolveTargetId(value);
      if (value !== "") {
        inspector.publishFeedback(`Selected ${value}`);
      }
    },
    [foundation.setSelectedResolveTargetId, inspector.publishFeedback],
  );

  useTimelineInspectorEscape({
    activeConflict: mutation.activeConflict,
    clearRowHistory: inspector.history.commands.clearRowHistory,
    isInspectorOpen: inspector.lifecycle.isOpen,
    restoreTimelineFocusAnchor: grid.restoreTimelineFocusAnchor,
    setInspectorMessage: inspector.publishFeedback,
    setIsInspectorOpen: inspector.setOpen,
    setSelectedMentionRef: foundation.setSelectedMentionRef,
    setSelectedRowId: inspector.selectRow,
    workbookFocusAnchorRef: grid.workbookFocusAnchorRef,
  });

  return {
    commands: {
      closeInspector: close.closeInspector,
      evidence: evidence.handleTimelineEvidenceFiles,
      feature: features.commands,
      history,
      mentions,
      resolveTargetChange: handleResolveTargetChange,
      rowInteractions: rowInteractions.commands,
      workflow: {
        submit: submitWorkflow,
        updateDraft: updateWorkflowDraft,
      },
    },
    ports: {},
    snapshot: {
      createRelatedReferenceOptions,
      createRelatedWorkflow,
      indicatorHandler: features.snapshot.indicatorHandler,
      rowInteractions: rowInteractions.snapshot,
    },
  };
}
