import type { GridHandle } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { type RefObject, useCallback } from "react";
import { sheetRefKey } from "../../../shared/sheetRef";
import { workbookInspectorMessageFeedback } from "../../inspector/workbookInspectorErrorModel";
import {
  type WorkbookInspectorState,
  workbookInspectorStateIsOpen,
} from "../../models/workbookInspectorModel";
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
import { useTimelineRowActionMenu } from "../hooks/useTimelineRowActionMenu";
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
  readonly rowMenuScopeKey: string;
  readonly mentionOwner: MentionInput["owner"];
  readonly mentionCandidates: MentionInput["candidatePort"];
  readonly earlierSaves: MentionInput["earlierSaves"];
  readonly foundation: {
    readonly currentCommittedRow: NonNullable<
      MentionInput["currentCommittedRow"]
    >;
    readonly acceptDisclosureSource: NonNullable<
      MentionInput["acceptDisclosureSource"]
    >;
    readonly evidenceAttachmentPort: EvidenceInput["owner"];
    readonly loadAccessLost: boolean;

    readonly rows: InspectorLifecycleInput["rows"];
    readonly rowsRef: MentionInput["rowsRef"];
    readonly selectedTargetId: string;
    readonly selectedMentionRef: InspectorLifecycleInput["selectedMentionRef"];

    readonly setSelectedMentionRef: InspectorLifecycleInput["setSelectedMentionRef"];
    readonly setSelectedResolveTargetId: InspectorLifecycleInput["setSelectedResolveTargetId"];
  };
  readonly grid: {
    readonly gridHandleRef: RefObject<GridHandle | null>;
    readonly gridShellRef: InspectorLifecycleInput["gridShellRef"];
    readonly restoreTimelineFocusAnchor: InspectorLifecycleInput["restoreTimelineFocusAnchor"];
    readonly focusContinuity: NonNullable<MentionInput["focusContinuity"]>;
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
    readonly commands: Pick<HistoryInput, "enqueueOrderedRead"> &
      Pick<HistoryInput, "acceptTimelineRecordVersion">;
    readonly waitForCommittedRecordIdle: HistoryInput["waitForCommittedRecordIdle"];
    readonly loadRows: HistoryInput["loadRows"];
    readonly publishViewingPresence: InspectorRowInteractionsInput["publishViewingPresence"];
  };
  readonly onAuthorityUncertain: TimelineWorkbookSurfaceRuntime["onAuthorityUncertain"];
  readonly activeSheetRef: TimelineWorkbookSurfaceRuntime["incident"]["sheetRef"];
};

export function useTimelineInspectorWorkflowComposition({
  activeSheetRef,
  rowMenuScopeKey,
  mentionOwner,
  mentionCandidates,
  earlierSaves,
  foundation,
  grid,
  incident,
  inspector,
  mutation,
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
  const rowMenu = useTimelineRowActionMenu({
    gridHandleRef: grid.gridHandleRef,
    readable: !!incident.currentRole && !foundation.loadAccessLost,
    rows: foundation.rows,
    scopeKey: rowMenuScopeKey,
  });
  const rowInteractions = useTimelineInspectorRowInteractions({
    currentCommittedRow: foundation.currentCommittedRow,
    elementRegistry: inspector.elementRegistry,
    publishViewingPresence: mutation.publishViewingPresence,
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
    enqueueOrderedRead: mutation.commands.enqueueOrderedRead,
    loadRows: mutation.loadRows,
    rowHistory: inspector.history.snapshot.rowHistory,
    setIsInspectorOpen: inspector.setOpen,
    setSelectedRowId: inspector.selectRow,
    waitForCommittedRecordIdle: mutation.waitForCommittedRecordIdle,
  });
  const mentions = useTimelineMentionActions({
    currentCommittedRow: foundation.currentCommittedRow,
    acceptDisclosureSource: foundation.acceptDisclosureSource,
    owner: mentionOwner,
    candidatePort: mentionCandidates,
    earlierSaves,
    selectedMention: inspector.selection.selectedMention,
    selectedMentionRef: foundation.selectedMentionRef,
    selectedRowId: inspector.selection.selectedRowId,
    inspectorInvalidationGeneration: inspector.lifecycle.invalidationGeneration,
    reviewSurfaceKey: `${actionContext.surfaceKey}:${incident.inspectorResetKey}`,
    selectedTargetId: foundation.selectedTargetId,
    setSelectedTargetId: foundation.setSelectedResolveTargetId,
    presentationKey: `${actionContext.surfaceKey}:${incident.inspectorResetKey}:${inspector.selection.selectedRowId}`,
    presentationActive:
      workbookInspectorStateIsOpen(inspector.lifecycle) &&
      !foundation.loadAccessLost,
    focusContinuity: grid.focusContinuity,
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
    owner: foundation.evidenceAttachmentPort,
    setInspectorMessage: inspector.publishFeedback,
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
      rowMenu: rowMenu.commands,
    },
    ports: {},
    snapshot: {
      createRelatedWorkflow,
      indicatorHandler: features.snapshot.indicatorHandler,
      rowInteractions: rowInteractions.snapshot,
      rowMenu: rowMenu.menu,
    },
  };
}

const timelineContract = requireViewContract(timelineViewSchemaId);
const timelineEvidenceCapabilityAvailable =
  timelineContract.inspectorConfig.featureGroups.some(
    (group) => group.featureGroupKey === "evidence.attach_blob",
  );
