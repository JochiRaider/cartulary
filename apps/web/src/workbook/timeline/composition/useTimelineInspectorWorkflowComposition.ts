import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useMemo } from "react";
import { sheetRefKey } from "../../../shared/sheetRef";
import { useIncidentMemberReferenceOptions } from "../../hooks/useOwnerReferenceOptions";
import { workbookInspectorMessageFeedback } from "../../inspector/workbookInspectorErrorModel";
import {
  type WorkbookInspectorState,
  workbookInspectorStateIsOpen,
} from "../../models/workbookInspectorModel";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import { useTimelineCreateRelatedWorkflow } from "../hooks/useTimelineCreateRelatedWorkflow";
import { useTimelineEvidenceAttach } from "../hooks/useTimelineEvidenceAttach";
import { useTimelineHistoryActions } from "../hooks/useTimelineHistoryActions";
import type { useTimelineHistoryState } from "../hooks/useTimelineHistoryState";
import { useTimelineInspectorFeatureController } from "../hooks/useTimelineInspectorFeatureController";
import {
  useTimelineInspectorEscape,
  useTimelineInspectorLifecycle,
  useTimelineInspectorRowInteractions,
} from "../hooks/useTimelineInspectorSelection";
import { useTimelineMentionActions } from "../hooks/useTimelineMentionActions";
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
  readonly mentionOwner: MentionInput["owner"];
  readonly mentionCandidates: MentionInput["candidatePort"];
  readonly earlierSaves: MentionInput["earlierSaves"];
  readonly foundation: {
    readonly evidenceAttachmentPort: EvidenceInput["evidenceAttachmentPort"];
    readonly loadAccessLost: boolean;

    readonly rows: InspectorLifecycleInput["rows"];
    readonly rowsRef: MentionInput["rowsRef"];
    readonly selectedTargetId: string;
    readonly selectedMentionRef: InspectorLifecycleInput["selectedMentionRef"];

    readonly setSelectedMentionRef: InspectorLifecycleInput["setSelectedMentionRef"];
    readonly setSelectedResolveTargetId: InspectorLifecycleInput["setSelectedResolveTargetId"];
  };
  readonly grid: {
    readonly beginViewportContinuity: EvidenceInput["beginViewportContinuity"];
    readonly clearViewportContinuity: EvidenceInput["clearViewportContinuity"];
    readonly gridShellRef: InspectorLifecycleInput["gridShellRef"];
    readonly restoreTimelineFocusAnchor: InspectorLifecycleInput["restoreTimelineFocusAnchor"];
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
    readonly elementRegistry: TimelineInspectorElementRegistry;
    readonly history: ReturnType<typeof useTimelineHistoryState>;
    readonly lifecycle: WorkbookInspectorState;
    readonly selection: {
      readonly inspectorMentions: InspectorLifecycleInput["inspectorMentions"];
      readonly selectedMention: MentionInput["selectedMention"];
      readonly selectedRow: CreateRelatedInput["selectedRow"];
      readonly selectedRowId: InspectorLifecycleInput["selectedRowId"];
      readonly selectedRowWorkflowSubject: CreateRelatedInput["selectedSubject"];
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
      EvidenceInput,
      "enqueueSaveWork" | "resolvePendingSocketTxn" | "trackPendingSocketTxn"
    > &
      Pick<HistoryInput, "acceptTimelineRecordVersion">;
    readonly waitForCommittedRecordIdle: EvidenceInput["waitForCommittedRecordIdle"];
    readonly loadRows: HistoryInput["loadRows"];
    readonly publishViewingPresence: InspectorRowInteractionsInput["publishViewingPresence"];
  };
  readonly onAuthorityUncertain: TimelineWorkbookSurfaceRuntime["onAuthorityUncertain"];
  readonly activeSheetRef: TimelineWorkbookSurfaceRuntime["incident"]["sheetRef"];
};

export function useTimelineInspectorWorkflowComposition({
  activeSheetRef,
  mentionOwner,
  mentionCandidates,
  earlierSaves,
  foundation,
  grid,
  incident,
  inspector,
  mutation,
  onAuthorityUncertain,
}: TimelineInspectorWorkflowCompositionInput) {
  const actionContext = {
    authorized:
      !foundation.loadAccessLost &&
      (incident.currentRole === "editor" ||
        incident.currentRole === "reviewer" ||
        incident.currentRole === "admin"),
    surfaceKey: sheetRefKey(activeSheetRef),
  };
  const {
    beginWorkflow,
    cancelWorkflow,
    submitWorkflow,
    updateWorkflowDraft,
    workflow: createRelatedWorkflow,
  } = useTimelineCreateRelatedWorkflow({
    isInspectorOpen: inspector.lifecycle.phase !== "closed",
    selectedRow: inspector.selection.selectedRow,
    selectedSubject: inspector.selection.selectedRowWorkflowSubject,
    setInspectorMessage: inspector.publishFeedback,
  });
  const features = useTimelineInspectorFeatureController({
    beginCreateRelatedWorkflow: beginWorkflow,
    cancelCreateRelatedWorkflow: cancelWorkflow,
    lifecycle: {
      authorizationKey: `${incident.currentRole ?? "none"}:${foundation.loadAccessLost}`,
      invalidationGeneration: inspector.lifecycle.invalidationGeneration,
      invalidationCause: inspector.lifecycle.invalidationCause,
      isOpen: inspector.lifecycle.phase !== "closed",
      lifecycleKey: `${incident.inspectorResetKey}:${incident.continuityResetKey}`,
      subject: inspector.selection.selectedRowWorkflowSubject,
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
    onAuthorityUncertain,
  });
  const createRelatedReferenceOptions = useMemo(
    () => ({
      ...emptyGenericReferenceOptions(),
      incidentMembers: incidentMemberOptions,
    }),
    [incidentMemberOptions],
  );
  const rowInteractions = useTimelineInspectorRowInteractions({
    elementRegistry: inspector.elementRegistry,
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
    clearRowHistory: inspector.history.commands.clearRowHistory,
    dispatchRowHistory: inspector.history.commands.dispatchRowHistory,
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
    setSelectedMentionRef: foundation.setSelectedMentionRef,
    setSelectedResolveTargetId: foundation.setSelectedResolveTargetId,
    setSelectedRowId: inspector.selectRow,
    workbookFocusAnchorRef: grid.workbookFocusAnchorRef,
  });
  const history = useTimelineHistoryActions({
    presentationActive: workbookInspectorStateIsOpen(inspector.lifecycle),
    acceptTimelineRecordVersion: mutation.commands.acceptTimelineRecordVersion,
    activeHistorySubject: inspector.history.snapshot.inspectorHistorySubject,
    dispatchRowHistory: inspector.history.commands.dispatchRowHistory,
    enqueueSaveWork: mutation.commands.enqueueSaveWork,
    loadRows: mutation.loadRows,
    rowHistory: inspector.history.snapshot.rowHistory,
    setIsInspectorOpen: inspector.setOpen,
    setSelectedRowId: inspector.selectRow,
    waitForCommittedRecordIdle: mutation.waitForCommittedRecordIdle,
  });
  const mentions = useTimelineMentionActions({
    owner: mentionOwner,
    candidatePort: mentionCandidates,
    earlierSaves,
    selectedMention: inspector.selection.selectedMention,
    selectedTargetId: foundation.selectedTargetId,
    setSelectedTargetId: foundation.setSelectedResolveTargetId,
    presentationKey: `${actionContext.surfaceKey}:${incident.inspectorResetKey}:${inspector.selection.selectedRowId}`,
    presentationActive:
      workbookInspectorStateIsOpen(inspector.lifecycle) &&
      !foundation.loadAccessLost,
    restoreActionFocus: (recordId) => {
      grid.restoreTimelineFocusAnchor({
        recordId,
        fieldKey: "timeline.activity_synopsis_text",
        viewSchemaId: timelineViewSchemaId,
      });
    },
    refreshProjection: () =>
      mutation.loadRows({ showLoading: false, requireAcceptance: true }),
    rowsRef: foundation.rowsRef,
    setInspectorMessage: inspector.publishFeedback,
    waitForCommittedRecordIdle: mutation.waitForCommittedRecordIdle,
  });
  const evidence = useTimelineEvidenceAttach({
    actionContext: {
      ...actionContext,
      capabilityAvailable: timelineEvidenceCapabilityAvailable,
      selectedRowKey: inspector.selection.selectedRow?.key ?? null,
    },
    applyAcceptedRowMutation: mutation.applyAcceptedRowMutation,
    beginViewportContinuity: grid.beginViewportContinuity,
    clearViewportContinuity: grid.clearViewportContinuity,
    enqueueSaveWork: mutation.commands.enqueueSaveWork,
    evidenceAttachmentPort: foundation.evidenceAttachmentPort,
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
        inspector.publishFeedback(
          workbookInspectorMessageFeedback(`Selected ${value}`, "none"),
        );
      }
    },
    [foundation.setSelectedResolveTargetId, inspector.publishFeedback],
  );

  useTimelineInspectorEscape({
    activeConflict: mutation.activeConflict,
    clearRowHistory: inspector.history.commands.clearRowHistory,
    isInspectorOpen: workbookInspectorStateIsOpen(inspector.lifecycle),
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

const timelineContract = requireViewContract(timelineViewSchemaId);
const timelineEvidenceCapabilityAvailable =
  timelineContract.inspectorConfig.featureGroups.some(
    (group) => group.featureGroupKey === "evidence.attach_blob",
  );
