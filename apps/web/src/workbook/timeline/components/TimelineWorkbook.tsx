import type {
  GridCellAnchor,
  GridCellPasteIntent,
  GridCellStateInput,
  GridColumn,
  GridDataRow,
  GridHandle,
  GridRowStateInput,
} from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  draftCellTestId,
  gridGroupRowTestId,
  timelineMutationSubstrateReadyTestId,
  workbookConflictSummaryTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  type SetStateAction,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { type SheetRef, sheetRefKey } from "../../../shared/sheetRef";
import { WorkbookGridControls } from "../../components/WorkbookGridControls";
import { WorkbookRowGutterContent } from "../../components/WorkbookPresenceMarkers";
import { WorkbookStatusStrip } from "../../components/WorkbookStatusStrip";
import { WorkbookViewBar } from "../../components/WorkbookViewBar";
import type {
  WorkbookContinuityAnchor,
  WorkbookContinuityToken,
} from "../../continuity/workbookContinuityPort";
import { IndicatorInspectorWorkflow } from "../../features/indicators/IndicatorInspectorWorkflow";
import { useIncidentMemberReferenceOptions } from "../../hooks/useOwnerReferenceOptions";
import { useWorkbookInspectorCoordinator } from "../../inspector/useWorkbookInspectorCoordinator";
import { WorkbookSurfaceLayout } from "../../layout/WorkbookSurfaceLayout";
import { applyWorkbookLayoutToColumns } from "../../layout/workbookColumnLayout";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../../models/workbookGridState";
import { selectInspectorConfig } from "../../models/workbookInspectorModel";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
  removeFilterField,
} from "../../models/workbookQuery";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { workbookClipboardPasteContract } from "../../utils/workbookClipboard";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import { createTimelineClipboardPasteAdapter } from "../adapters/createTimelineClipboardPasteAdapter";
import { createTimelineEvidenceAttachmentAdapter } from "../adapters/createTimelineEvidenceAttachmentAdapter";
import { createTimelineHistoryAdapter } from "../adapters/createTimelineHistoryAdapter";
import { createTimelineMentionAdapter } from "../adapters/createTimelineMentionAdapter";
import { createTimelineRecordActionAdapter } from "../adapters/createTimelineRecordActionAdapter";
import { createTimelineRowMutationEditorAdapter } from "../adapters/createTimelineRowMutationEditorAdapter";
import { useTimelineBulkTagController } from "../bulk/useTimelineBulkTagController";
import { useTimelineFillController } from "../bulk/useTimelineFillController";
import { TimelineCollaborationBoundary } from "../collaboration/TimelineCollaborationBoundary";
import { useTimelineCollaborationBindings } from "../collaboration/useTimelineCollaborationBindings";
import { useTimelinePresenceController } from "../collaboration/useTimelinePresenceController";
import { useTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineClipboardPasteController } from "../hooks/useTimelineClipboardPasteController";
import { useTimelineCreateRelatedWorkflow } from "../hooks/useTimelineCreateRelatedWorkflow";
import { useTimelineEvidenceActions } from "../hooks/useTimelineEvidenceActions";
import { useTimelineEvidenceAttach } from "../hooks/useTimelineEvidenceAttach";
import { useTimelineGridAnchorController } from "../hooks/useTimelineGridAnchorController";
import {
  type TimelineGridInteractionRefs,
  useTimelineGridInteractions,
} from "../hooks/useTimelineGridInteractions";
import { useTimelineHistoryActions } from "../hooks/useTimelineHistoryActions";
import { useTimelineHistoryState } from "../hooks/useTimelineHistoryState";
import { useTimelineInspectorFeatureController } from "../hooks/useTimelineInspectorFeatureController";
import {
  useTimelineInspectorEscape,
  useTimelineInspectorLifecycle,
  useTimelineInspectorRowInteractions,
  useTimelineInspectorSelection,
} from "../hooks/useTimelineInspectorSelection";
import { useTimelineKeyboardController } from "../hooks/useTimelineKeyboardController";
import { useTimelineMentionActions } from "../hooks/useTimelineMentionActions";
import { useTimelineMentions } from "../hooks/useTimelineMentions";
import { useTimelineMutationCommands } from "../hooks/useTimelineMutationCommands";
import { useTimelinePendingReplayController } from "../hooks/useTimelinePendingReplayController";
import { useTimelinePendingSaves } from "../hooks/useTimelinePendingSaves";
import { useTimelineRows } from "../hooks/useTimelineRows";
import {
  type LoadRowsOptions,
  useTimelineRowsLoader,
} from "../hooks/useTimelineRowsLoader";
import {
  useTimelineViewportContinuityController,
  type TimelineViewportContinuityRequest as ViewportContinuityRequest,
} from "../hooks/useTimelineViewportContinuityController";
import { useTimelineWorkbookRuntime } from "../hooks/useTimelineWorkbookRuntime";
import type { PendingReplayRuntimeMeta } from "../models/timelineControllerPorts";
import { buildTimelineGridRows } from "../models/timelineRowsModel";
import { timelineCreateRelatedTargetContracts } from "../models/timelineWorkbookFeaturePolicy";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";
import { reconcileDismissedMentionsForRow } from "../models/workbookMentionChips";
import {
  createDraftRow,
  inputFocusKey,
  type RowValues,
  type TimelineScalarEditorSurface,
  timelineGroupLabel,
  timelineRelationshipLabel,
  timelineScalarBindingForField,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import { useTimelineRowMutationCoordinator } from "../mutations/useTimelineRowMutationCoordinator";

function gridCoreRecordId(
  anchor: GridCellAnchor | GridCellPasteIntent["target"],
): string | null {
  return anchor.rowIdentity.kind === "core_record"
    ? anchor.rowIdentity.recordId
    : null;
}

import { DraftRowCreateButton } from "./TimelineCellEditors";
import { TimelineGridSurface } from "./TimelineGridSurface";
import { TimelineRowContextMenu } from "./TimelineRowActions";
import { TimelineWorkbookInspector } from "./TimelineWorkbookInspector";
import { useTimelineWorkbookInspectorSections } from "./TimelineWorkbookInspectorSections";
import {
  TimelineWorkbookNotices,
  timelinePendingQueueMessage,
} from "./TimelineWorkbookNotices";
import { useTimelineWorkbookRenderers } from "./TimelineWorkbookRenderers";
import {
  timelineGridShellStyle,
  timelineRowGutterWidth,
} from "./TimelineWorkbookStyles";

const timelineContract = requireViewContract(timelineViewSchemaId);
const timelineInspectorConfig = selectInspectorConfig(timelineContract);
const timelineObservationSourceFields = [
  { fieldKey: "timeline.date_entered_text", label: "Date Entered" },
  { fieldKey: "timeline.analyst_text", label: "Analyst" },
  { fieldKey: "timeline.mitre_stage_text", label: "MITRE Stage" },
  { fieldKey: "timeline.device_object_text", label: "Device/Object" },
  { fieldKey: "timeline.ip_address_text", label: "IP Address" },
  { fieldKey: "timeline.activity_utc_text", label: "Activity UTC" },
  { fieldKey: "timeline.activity_local_text", label: "Activity Local" },
  { fieldKey: "timeline.raw_activity_text", label: "Raw Activity" },
  { fieldKey: "timeline.activity_synopsis_text", label: "Activity Synopsis" },
  { fieldKey: "timeline.data_source_text", label: "Data Source" },
] as const;
const bulkActionFieldsetStyle = {
  border: 0,
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  margin: 0,
  padding: 0,
};

export type TimelineWorkbookProps = {
  readonly runtime: TimelineWorkbookSurfaceRuntime;
};

function recordWorkbookTiming(
  name: string,
  details: Record<string, unknown> = {},
) {
  if (typeof performance === "undefined") {
    return;
  }
  performance.mark(`cartulary.workbook.${name}`, { detail: details });
}

function TimelineWorkbookContent({
  collaborationProjection,
  indicatorWorkflow,
  mutationCommands,
  mutationRuntime,
  pendingMutationPort,
  incident,
  query,
  entities,
  layout,
  onIncidentAccessLost,
}: TimelineWorkbookSurfaceRuntime) {
  const {
    id: incidentId,
    apiBase,
    continuityResetKey,
    currentUserId,
    incidentPort,
    sheetRef,
    inspectorResetKey,
    reloadToken,
    currentRole: currentIncidentRole,
  } = incident;
  const {
    filterDraft: shellFilterDraft,
    setFilterDraft: setShellFilterDraft,
    state: shellQueryState,
    setState: setShellQueryState,
    renderInlineControls: renderInlineQueryControls,
    savedViewSelector,
    viewQuery,
    viewBarQueryControls,
  } = query;
  const {
    hosts: hostEntities,
    identities: identityEntities,
    index: entityIndex,
    refresh: onRefreshEntities,
  } = entities;
  const timelineHistory = useMemo(
    () => createTimelineHistoryAdapter({ apiBase }),
    [apiBase],
  );
  const timelineRecordActions = useMemo(
    () => createTimelineRecordActionAdapter({ apiBase }),
    [apiBase],
  );
  const timelineMentionPort = useMemo(
    () => createTimelineMentionAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const timelineClipboardPastePort = useMemo(
    () => createTimelineClipboardPasteAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const timelineEvidenceAttachment = useMemo(
    () =>
      createTimelineEvidenceAttachmentAdapter({
        apiBase,
        createClientTxnId: mutationCommands.identity.createLogicalActionId,
        incidentId,
      }),
    [apiBase, incidentId, mutationCommands.identity],
  );
  const {
    commands: {
      onColumnHiddenChange: handleColumnHiddenChange,
      onColumnMove: handleColumnMove,
      onColumnReorder: handleColumnReorder,
      onColumnWidthChange: handleColumnWidthChange,
      onResetColumns: handleResetColumns,
    },
    snapshot: {
      chromeMode,
      density,
      interactionMode,
      showStatusPresence,
      state: layoutState,
    },
  } = layout;
  const clientInstanceId = mutationRuntime.scope.clientInstanceId;
  const timelineRuntime = useTimelineWorkbookRuntime({
    filterDraft: shellFilterDraft,
    queryState: shellQueryState,
    setFilterDraft: setShellFilterDraft,
    setQueryState: setShellQueryState,
  });
  const {
    isInitialLoading,
    isRefreshing,
    loadError,
    refreshError,
    setIsInitialLoading,
    setIsRefreshing,
    setLoadError,
    setRefreshError,
  } = timelineRuntime.lifecycle;
  const [loadAccessLost, setLoadAccessLost] = useState(false);
  const [initialLoadGenerationKey, setInitialLoadGenerationKey] = useState(0);
  const {
    applyQueryFilter,
    filterDraft,
    handleQueryGroupByChange,
    handleQuerySortChange,
    queryState,
    setFilterDraft,
    setQueryState,
  } = timelineRuntime.query;
  const initialTimelineRows = useMemo(() => [createDraftRow(1)], []);
  const rowsRef = useRef<WorkbookRow[]>(initialTimelineRows);
  const draftCounterRef = useRef(2);
  const timelineRows = useTimelineRows({ draftCounterRef, rowsRef });
  const { rows } = timelineRows.snapshot;
  const { setRows } = timelineRows.commands;
  const canBulkTag =
    interactionMode.kind === "editable" &&
    (currentIncidentRole === "editor" ||
      currentIncidentRole === "reviewer" ||
      currentIncidentRole === "admin");
  const getTimelineRowState = useCallback(
    (row: GridDataRow<WorkbookRow>): GridRowStateInput => ({
      pending: row.data.pendingSignature !== null,
    }),
    [],
  );
  const timelineMentions = useTimelineMentions();
  const {
    autoResolutionNotices,
    dismissedMentionsByRow,
    selectedMentionRef,
    selectedResolveTargetId,
  } = timelineMentions.snapshot;
  const {
    setAutoResolutionNotices,
    setDismissedMentionsByRow,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
  } = timelineMentions.commands;
  const timelineEvidenceActions = useTimelineEvidenceActions();
  const { inspectorMessage } = timelineEvidenceActions.snapshot;
  const { setInspectorMessage } = timelineEvidenceActions.commands;
  const timelinePendingSaves =
    useTimelinePendingSaves<PendingReplayRuntimeMeta>({
      mutationRuntime,
    });
  const { pendingQueueSnapshot } = timelinePendingSaves.snapshot;
  const { setPendingQueueSnapshot } = timelinePendingSaves.commands;
  const pendingSavesRefsRef = useRef(timelinePendingSaves.refs);
  pendingSavesRefsRef.current = timelinePendingSaves.refs;
  const workbookFocusAnchorRef = useRef<WorkbookContinuityAnchor | null>(null);
  const inspectorContinuityTokenRef = useRef<WorkbookContinuityToken | null>(
    null,
  );
  const editorDraftRegistry =
    useTimelineEditorDraftRegistry(timelineViewSchemaId);
  const timelineAnchorColumnsRef = useRef<readonly GridColumn<WorkbookRow>[]>(
    [],
  );
  const timelineGridHandleRef = useRef<GridHandle | null>(null);
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const recoveryPanelRef = useRef<HTMLDivElement | null>(null);
  const discardBlockedEditRef = useRef<(unitId: string) => boolean>(
    () => false,
  );
  const schedulePendingReplayRuntimeRef = useRef<() => void>(() => undefined);
  const [timelineGridShellWidth, setTimelineGridShellWidth] = useState(0);
  const viewportContinuityTokenRef = useRef(1);
  const timelineGridInteractionRefs: TimelineGridInteractionRefs = {
    gridHandleRef: timelineGridHandleRef,
    gridShellRef,
    timelineAnchorColumnsRef,
    viewportContinuityTokenRef,
    workbookFocusAnchorRef,
  };
  const timelineGridInteractions =
    useTimelineGridInteractions<ViewportContinuityRequest>({
      continuityResetKey,
      refs: timelineGridInteractionRefs,
    });
  const { continuityPort } = timelineGridInteractions;
  const { viewportContinuityRequest, workbookFocusAnchor } =
    timelineGridInteractions.snapshot;
  const {
    setViewportContinuityRequest,
    updateTimelineFocusAnchor,
    updateWorkbookFocusAnchor,
  } = timelineGridInteractions.commands;
  useLayoutEffect(() => {
    const gridShell = gridShellRef.current;
    if (gridShell === null) {
      return;
    }
    const updateGridShellWidth = (width: number) => {
      const measuredWidth = Math.max(0, Math.floor(width));
      setTimelineGridShellWidth((current) =>
        current === measuredWidth ? current : measuredWidth,
      );
    };
    const updateFromElement = () => {
      updateGridShellWidth(gridShell.clientWidth);
    };
    const scheduleUpdateFromElement = () => {
      updateFromElement();
      window.requestAnimationFrame(updateFromElement);
    };

    updateFromElement();
    window.addEventListener("resize", scheduleUpdateFromElement);
    window.visualViewport?.addEventListener(
      "resize",
      scheduleUpdateFromElement,
    );

    if (typeof ResizeObserver === "undefined") {
      return () => {
        window.removeEventListener("resize", scheduleUpdateFromElement);
        window.visualViewport?.removeEventListener(
          "resize",
          scheduleUpdateFromElement,
        );
      };
    }

    const observer = new ResizeObserver(() => {
      scheduleUpdateFromElement();
    });
    observer.observe(gridShell);
    observer.observe(document.documentElement);
    return () => {
      window.removeEventListener("resize", scheduleUpdateFromElement);
      window.visualViewport?.removeEventListener(
        "resize",
        scheduleUpdateFromElement,
      );
      observer.disconnect();
    };
  }, []);

  useLayoutEffect(() => {
    const gridShell = gridShellRef.current;
    if (gridShell === null) {
      return;
    }
    const measuredWidth = Math.max(0, Math.floor(gridShell.clientWidth));
    setTimelineGridShellWidth((current) =>
      current === measuredWidth ? current : measuredWidth,
    );
  }, []);

  const [activeCollectionInputKey, setActiveCollectionInputKey] = useState<
    string | null
  >(null);
  const timelineInspectorSelection = useTimelineInspectorSelection({
    currentIncidentRole,
    dismissedMentionsByRow,
    rows,
    selectedMentionRef,
  });
  const {
    canManageMentions,
    draftRow,
    inspectorMentions,
    selectedMention,
    selectedRow,
    selectedRowId,
    selectedRowWorkflowKey,
  } = timelineInspectorSelection.snapshot;
  const { setSelectedRowId } = timelineInspectorSelection.commands;
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      clearLifecycleState: () => {
        continuityPort.clear();
      },
      clearSelection: () => {
        setSelectedRowId(null);
      },
      restoreFocus: () => {
        const token = inspectorContinuityTokenRef.current;
        inspectorContinuityTokenRef.current = null;
        if (token !== null) {
          continuityPort.restore(token);
        }
      },
    },
    config: timelineInspectorConfig,
    lifecycleKey: inspectorResetKey,
    subject:
      selectedRow?.recordId === null ||
      selectedRow?.recordId === undefined ||
      selectedRow.rowVersion === null
        ? null
        : {
            recordId: selectedRow.recordId,
            rowVersion: selectedRow.rowVersion,
            viewSchemaId: timelineViewSchemaId,
          },
  });
  const isInspectorOpen = inspector.snapshot.isOpen;
  const setIsInspectorOpen = useCallback(
    (next: SetStateAction<boolean>) => {
      const nextOpen =
        typeof next === "function" ? next(inspector.snapshot.isOpen) : next;
      if (nextOpen && !inspector.snapshot.isOpen) {
        inspectorContinuityTokenRef.current = continuityPort.capture();
      } else if (!nextOpen && inspector.snapshot.isOpen) {
        inspectorContinuityTokenRef.current = continuityPort.capture(
          workbookFocusAnchorRef.current,
        );
      }
      inspector.commands.setOpen(next);
    },
    [continuityPort, inspector.commands, inspector.snapshot.isOpen],
  );
  const timelineHistoryState = useTimelineHistoryState({
    draftRow,
    selectedRow,
  });
  const {
    activeHistoryLiveRecordId,
    currentHistoryDeleted,
    currentHistoryRecordId,
    currentHistoryRowVersion,
    inspectorHistorySubject,
    rowHistory,
    rowHistoryPendingAction,
  } = timelineHistoryState.snapshot;
  const {
    beginRowHistoryRequest,
    cancelRowHistoryRequests,
    clearRowHistory,
    currentHistoryRecordIdMatches,
    rowHistoryRequestIsCurrent,
    setRowHistory,
    setRowHistoryPendingAction,
  } = timelineHistoryState.commands;
  const loadRowsRef = useRef<(options: LoadRowsOptions) => Promise<void>>(
    async () => undefined,
  );

  const activeSheetRef = useMemo<SheetRef>(
    () => sheetRef ?? { kind: "view_schema", id: timelineViewSchemaId },
    [sheetRef],
  );
  const updateTimelineSurfaceFocusAnchor = useCallback(
    (recordId: string | null, fieldKey: string) => {
      updateTimelineFocusAnchor(recordId, fieldKey, timelineViewSchemaId);
    },
    [updateTimelineFocusAnchor],
  );

  const nextDraftIndex = useCallback(() => {
    const value = draftCounterRef.current;
    draftCounterRef.current += 1;
    return value;
  }, []);

  const handleResolveTargetChange = useCallback(
    (value: string) => {
      setSelectedResolveTargetId(value);
      if (value !== "") {
        setInspectorMessage(`Selected ${value}`);
      }
    },
    [setSelectedResolveTargetId, setInspectorMessage],
  );

  const timelineViewportContinuity = useTimelineViewportContinuityController({
    editorDraftRegistry,
    gridHandleRef: timelineGridHandleRef,
    gridShellRef,
    setViewportContinuityRequest,
    viewportContinuityRequest,
    viewportContinuityTokenRef,
  });
  const {
    advanceViewportContinuity,
    beginViewportContinuity,
    clearViewportContinuity,
    failViewportContinuity,
    requireViewportContinuitySourceRecord,
    resolveInputElement,
    settleViewportContinuityFollowUp,
  } = timelineViewportContinuity.commands;
  const timelineMutationEditorPort = useMemo(
    () =>
      createTimelineRowMutationEditorAdapter({
        continuityPort,
        focusInput: (focusKey) => {
          resolveInputElement(focusKey)?.focus({ preventScroll: true });
        },
        gridHandleRef: timelineGridHandleRef,
      }),
    [continuityPort, resolveInputElement],
  );
  const timelineRowMutations = useTimelineRowMutationCoordinator({
    advanceViewportContinuity,
    clearViewportContinuity,
    discardBlockedEditRef,
    editorDraftRegistry,
    editorPort: timelineMutationEditorPort,
    loadRowsRef,
    mutationRuntime,
    nextDraftIndex,
    pendingQueueSnapshot,
    pendingSavesRefsRef,
    rowsRef,
    schedulePendingReplayRef: schedulePendingReplayRuntimeRef,
    selectedRowId,
    setActiveCollectionInputKey,
    setAutoResolutionNotices,
    setDismissedMentionsByRow,
    setPendingQueueSnapshot,
    setRows,
    setSelectedRowId,
  });
  const { activeConflict, commonMutationSnapshot, conflictQueue } =
    timelineRowMutations.snapshot;
  const { collaborationAdmission: timelineCollaborationAdmission } =
    timelineRowMutations.ports;
  const {
    beginLoad: beginTimelineRowsLoad,
    committedRowsChangedSince,
    currentCommittedTimelineRow,
    hasLoadedRows,
    isCurrentLoadSequence,
    knownTimelineRowVersion,
    markRowsLoaded,
  } = timelineRowMutations.ports.queryAdmission;
  const {
    acceptCommittedTimelineRows,
    acceptTimelineActionResult,
    acceptTimelineRecordVersion,
    activateConflict,
    applyAcceptedRowMutation: applyNormalizedRowMutation,
    applyClipboardResponseRows: applyTimelineClipboardResponseRows,
    beginRefreshInFlight,
    beginSave,
    enqueueSaveWork: enqueueTimelineSaveWork,
    finishRefreshInFlight,
    finishSave,
    latestCommittedTimelineRow,
    pruneAutoResolutionNoticesForRows,
    publishPendingQueueState,
    publishSaveStatePresentation,
    reconcileDiscardedPendingUnit,
    registerSameFieldConflict,
    resolvePendingSocketTxn,
    setActiveConflictKey,
    setPasteConflictGroup,
    trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  } = timelineRowMutations.commands;
  const { conflictQueueRef } = timelineRowMutations.refs;
  const conflictCellKeys = useMemo(
    () =>
      new Set(
        Object.values(conflictQueue).map(
          (entry) => `${entry.anchor.record_id}\u0000${entry.anchor.field_key}`,
        ),
      ),
    [conflictQueue],
  );
  const getTimelineCellState = useCallback(
    ({
      fieldKey,
      recordId,
    }: {
      readonly fieldKey: string;
      readonly recordId: string;
    }): GridCellStateInput => ({
      conflicted: conflictCellKeys.has(`${recordId}\u0000${fieldKey}`),
    }),
    [conflictCellKeys],
  );
  useEffect(() => {
    if (
      viewportContinuityRequest === null ||
      !commonMutationSnapshot.conflicts.some(
        (entry) => entry.origin.viewSchemaId === timelineViewSchemaId,
      )
    ) {
      return;
    }
    clearViewportContinuity(viewportContinuityRequest.token);
  }, [
    clearViewportContinuity,
    commonMutationSnapshot.conflicts,
    viewportContinuityRequest,
  ]);
  const {
    currentTimelineAnchorFor,
    navigateTimelineFocusAnchor,
    resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor,
  } = useTimelineGridAnchorController({
    continuityPort,
    gridHandleRef: timelineGridHandleRef,
    rowsRef,
    timelineAnchorColumnsRef,
    updateTimelineSurfaceFocusAnchor,
    updateWorkbookFocusAnchor,
  });

  const { loadRows } = useTimelineRowsLoader({
    acceptCommittedTimelineRows,
    advanceViewportContinuity,
    beginRefreshInFlight,
    beginTimelineRowsLoad,
    committedRowsChangedSince,
    currentCommittedTimelineRow,
    finishRefreshInFlight,
    failViewportContinuity,
    hasLoadedRows,
    isCurrentLoadSequence,
    knownTimelineRowVersion,
    loadRowsRef,
    markRowsLoaded,
    nextDraftIndex,
    onIncidentAccessLost,
    pendingSavesRefsRef,
    pruneAutoResolutionNoticesForRows,
    pruneDismissedMentionsForRow: reconcileDismissedMentionsForRow,
    publishSaveStatePresentation,
    queryState,
    rowsRef,
    editorDraftRegistry,
    setDismissedMentionsByRow,
    setIsInitialLoading,
    setInitialLoadGenerationKey,
    setIsRefreshing,
    setLoadAccessLost,
    setLoadError,
    setRefreshError,
    setRows,
    viewQuery,
  });
  const {
    beginWorkflow: beginCreateRelatedWorkflow,
    cancelWorkflow: cancelCreateRelatedWorkflow,
    submitWorkflow: submitCreateRelatedWorkflow,
    updateWorkflowDraft: updateCreateRelatedWorkflowDraft,
    workflow: createRelatedWorkflow,
  } = useTimelineCreateRelatedWorkflow({
    applyAcceptedRowMutation: applyNormalizedRowMutation,
    currentUserId,
    loadRows: loadRowsRef.current,
    mutationCommands: mutationCommands.related,
    selectedRow,
    selectedRowWorkflowKey,
    setInspectorMessage,
    targetContracts: timelineCreateRelatedTargetContracts,
  });
  const timelineInspectorFeatures = useTimelineInspectorFeatureController({
    beginCreateRelatedWorkflow,
    cancelCreateRelatedWorkflow,
    lifecycle: {
      authorizationKey: `${currentIncidentRole ?? "none"}:${loadAccessLost}`,
      invalidationGeneration: inspector.snapshot.invalidationGeneration,
      lifecycleKey: `${inspectorResetKey}:${continuityResetKey}`,
      subjectKey: selectedRowWorkflowKey,
      surfaceKey: sheetRefKey(activeSheetRef),
    },
    setInspectorMessage,
  });
  const {
    cancelFeatureAction: cancelInspectorFeatureAction,
    handleFeatureAction: handleInspectorFeatureAction,
    isFeatureActionSupported: supportsTimelineInspectorFeature,
  } = timelineInspectorFeatures.commands;
  const { indicatorHandler: indicatorInspectorHandler } =
    timelineInspectorFeatures.snapshot;
  const createRelatedNeedsIncidentMembers =
    createRelatedWorkflow?.targetContract.fields.some(
      (field) =>
        field.directReferenceContractId === "incident_member_user_ref_v1",
    ) ?? false;
  const { options: timelineIncidentMemberOptions } =
    useIncidentMemberReferenceOptions({
      enabled: createRelatedNeedsIncidentMembers,
      incidentPort,
      onIncidentAccessLost,
    });
  const timelineCreateRelatedReferenceOptions = useMemo(
    () => ({
      ...emptyGenericReferenceOptions(),
      incidentMembers: timelineIncidentMemberOptions,
    }),
    [timelineIncidentMemberOptions],
  );

  const refreshRowsForCollaboration = useCallback(
    () => loadRowsRef.current({ showLoading: false }),
    [],
  );
  const timelineCollaboration = useTimelineCollaborationBindings({
    activeSheetRef,
    admission: timelineCollaborationAdmission,
    beginRowsLoad: beginTimelineRowsLoad,
    collaborationProjection,
    refreshRows: refreshRowsForCollaboration,
    resolveClientTxn: resolvePendingSocketTxn,
    rowsRef,
    setRows,
  });

  useEffect(() => {
    // Keep saved-view reselection and shell-level refreshes observable here even
    // though loadRows owns the actual query inputs.
    void reloadToken;
    void loadRows({ showLoading: true });
  }, [loadRows, reloadToken]);

  useTimelineInspectorLifecycle({
    cancelRowHistoryRequests,
    clearRowHistory,
    gridShellRef,
    inspectorInvalidationCause: inspector.snapshot.invalidationCause,
    inspectorMentions,
    inspectorInvalidationGeneration: inspector.snapshot.invalidationGeneration,
    restoreTimelineFocusAnchor,
    rowHistory,
    rows,
    selectedMentionRef,
    selectedRowId,
    setInspectorMessage,
    setRowHistory,
    setRowHistoryPendingAction,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
    setSelectedRowId,
    workbookFocusAnchorRef,
  });

  const nextClientTxnId = useCallback(() => {
    return mutationCommands.identity.createLogicalActionId();
  }, [mutationCommands.identity]);

  const refreshRowsForBulkTag = useCallback(
    () => loadRows({ showLoading: false }),
    [loadRows],
  );
  const timelineBulkTag = useTimelineBulkTagController({
    canAssign: canBulkTag,
    port: mutationCommands.bulk,
    refreshRows: refreshRowsForBulkTag,
    rows,
    rowsRef,
  });
  const {
    canSubmit: canSubmitBulkTag,
    gridSelection: timelineBulkSelection,
    message: bulkTagMessage,
    selectedRecordIds: selectedTimelineRecordIds,
    tagName: bulkTagName,
  } = timelineBulkTag.snapshot;
  const { assignTag: assignTagToSelectedRows, changeTagName: setBulkTagName } =
    timelineBulkTag.commands;

  const {
    discardBlockedEdit,
    enqueuePendingReplayUnit,
    schedulePendingReplay,
  } = useTimelinePendingReplayController({
    applyAcceptedRowMutation: applyNormalizedRowMutation,
    clearSubmittedScalarEditorDraftValuesForRow:
      editorDraftRegistry.clearSubmittedRow,
    clearViewportContinuity,
    conflictQueueRef,
    registerMutationConflict: (
      conflict,
      rowKey,
      focusField,
      surface,
      refresh,
    ) => {
      registerSameFieldConflict(
        conflict,
        inputFocusKey(rowKey, focusField, surface),
        surface,
        refresh,
      );
      return true;
    },
    latestCommittedTimelineRow,
    loadRowsRef,
    mutationCommands: mutationCommands.identity,
    mutationRuntime,
    pendingSavesRefsRef,
    postMutationQueryRefreshRequired:
      queryState.filters.length > 0 ||
      queryState.sort.length > 0 ||
      queryState.groupBy !== null,
    publishPendingQueueState,
    reconcileDiscardedPendingUnit,
    recordWorkbookTiming,
    resolvePendingSocketTxn,
    rowsRef,
    requestAuthorizationRecovery: () =>
      timelineCollaboration.commands.requestAuthorizationRecovery(),
    setRefreshError,
    setRows,
    pendingMutationPort,
    trackPendingSocketTxn,
  });
  discardBlockedEditRef.current = discardBlockedEdit;
  schedulePendingReplayRuntimeRef.current = schedulePendingReplay;

  const activeSheetPresenceRecords =
    timelineCollaboration.snapshot.activeSheetPresenceRecords;
  const timelinePresence = useTimelinePresenceController({
    presenceRecords: activeSheetPresenceRecords,
    publishPresence: timelineCollaboration.commands.publishPresence,
    resetKey: `${continuityResetKey}:${
      loadAccessLost ? "access-lost" : "authorized"
    }`,
  });
  const { editingPresenceForCell, presenceForRow } = timelinePresence.snapshot;
  const {
    publishEditModePresence: handleEditModePresence,
    publishViewingPresence,
  } = timelinePresence.commands;

  const timelineInspectorRowInteractions = useTimelineInspectorRowInteractions({
    publishViewingPresence,
    rows,
    rowsRef,
    selectedRowId,
    setInspectorMessage,
    setIsInspectorOpen,
    setSelectedMentionRef,
    setSelectedRowId,
  });
  const { activeRowContextMenuRow, rowContextMenu } =
    timelineInspectorRowInteractions.snapshot;
  const {
    closeRowContextMenu,
    handleSelectMention,
    handleSelectRow,
    handleTimelineGridContextKeyDown,
    handleTimelineGridContextMenu,
    openInspectorForRow,
    timelineRowForEventTarget,
  } = timelineInspectorRowInteractions.commands;

  const timelineMutationCommands = useTimelineMutationCommands({
    acceptTimelineActionResult,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    clientInstanceId,
    conflictQueueRef,
    editorDraftRegistry,
    enqueueSaveWork: enqueueTimelineSaveWork,
    enqueuePendingReplayUnit,
    finishSave,
    incidentId,
    latestCommittedTimelineRow,
    loadRowsRef,
    nextClientTxnId,
    pendingSavesRefsRef,
    recordActionPort: timelineRecordActions,
    resolvePendingSocketTxn,
    rowsRef,
    setRows,
    trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  });
  const { replacementDrafts } = timelineMutationCommands.snapshot;
  const {
    changeReplacementDraft,
    commitScalarGridEdit,
    queueAction,
    queueCollectionSave,
    queueScalarSave,
  } = timelineMutationCommands.commands;

  const {
    confirmRowHistoryPendingAction,
    openRowHistory,
    previewRowHistoryDeleteRestore,
    previewRowHistoryRollback,
  } = useTimelineHistoryActions({
    acceptTimelineRecordVersion,
    activeHistoryLiveRecordId,
    beginRowHistoryRequest,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    currentHistoryRecordId,
    currentHistoryRecordIdMatches,
    currentHistoryRowVersion,
    enqueueSaveWork: enqueueTimelineSaveWork,
    finishSave,
    historyPort: timelineHistory,
    loadRows: loadRowsRef.current,
    nextClientTxnId,
    resolvePendingSocketTxn,
    rowHistory,
    rowHistoryPendingAction,
    rowHistoryRequestIsCurrent,
    selectedRowRecordId: selectedRow?.recordId ?? null,
    setIsInspectorOpen,
    setRowHistory,
    setRowHistoryPendingAction,
    setSelectedRowId,
    trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  });

  const {
    createEntityFromMention,
    handleUndoAutoResolutionNotice,
    submitMentionAction,
  } = useTimelineMentionActions({
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    enqueueSaveWork: enqueueTimelineSaveWork,
    finishSave,
    loadRows: loadRowsRef.current,
    mentionPort: timelineMentionPort,
    nextClientTxnId,
    onRefreshEntities,
    requireViewportContinuitySourceRecord,
    resolvePendingSocketTxn,
    rowsRef,
    setDismissedMentionsByRow,
    setInspectorMessage,
    settleViewportContinuityFollowUp,
    trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  });

  const { handleTimelineEvidenceFiles } = useTimelineEvidenceAttach({
    applyAcceptedRowMutation: applyNormalizedRowMutation,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    enqueueSaveWork: enqueueTimelineSaveWork,
    evidenceAttachmentPort: timelineEvidenceAttachment,
    finishSave,
    resolvePendingSocketTxn,
    rowsRef,
    setInspectorMessage,
    trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  });

  const handleBlur = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
      currentValue: string,
    ) => {
      queueScalarSave(
        rowKey,
        focusField,
        {
          continueOnFreshDraft: false,
          preserveInputFocus: false,
          surface,
        },
        currentValue,
      );
    },
    [queueScalarSave],
  );

  const {
    onCollectionEditorKeyDown: handleCollectionKeyDown,
    onScalarEditorKeyDown: handleKeyDown,
    onWorkAreaKeyDown: handleTimelineWorkAreaKeyDown,
  } = useTimelineKeyboardController({
    clearRowHistory,
    currentTimelineAnchorFor,
    handleTimelineGridContextKeyDown,
    navigateTimelineFocusAnchor,
    openRowHistory,
    queueCollectionSave,
    queueScalarSave,
    recordTiming: recordWorkbookTiming,
    restoreTimelineFocusAnchor,
    rowHistory,
    selectedRowId,
    setInspectorMessage,
    setIsInspectorOpen,
    setSelectedMentionRef,
    setSelectedRowId,
    timelineRowForEventTarget,
    workbookFocusAnchorRef,
  }).commands;

  const { handleGridPaste, handlePaste } = useTimelineClipboardPasteController({
    applyResponseRows: applyTimelineClipboardResponseRows,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    editorDraftRegistry,
    finishSave,
    loadRowsRef,
    nextClientTxnId,
    pendingSavesRefsRef,
    queueScalarSave,
    registerSameFieldConflict,
    resolvePendingSocketTxn,
    resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor,
    setActiveConflictKey,
    setPasteConflictGroup,
    timelineClipboardPaste: timelineClipboardPastePort,
    trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  }).commands;

  const handleTimelineGridPaste = useCallback(
    (intent: Parameters<typeof handleGridPaste>[0]) => {
      if (intent.input.kind === "scalar") {
        const binding = timelineScalarBindingForField(intent.target.fieldKey);
        if (binding === null) return;
        void commitScalarGridEdit(
          gridCoreRecordId(intent.target) ?? "",
          binding.key,
          intent.input.value,
        ).then((outcome) => {
          if (outcome.kind !== "accepted") setRefreshError(outcome.message);
        });
        return;
      }
      handleGridPaste(intent);
    },
    [commitScalarGridEdit, handleGridPaste, setRefreshError],
  );
  const timelineClipboardPaste = useMemo(
    () => workbookClipboardPasteContract(handleTimelineGridPaste),
    [handleTimelineGridPaste],
  );

  const { onFillCells: handleFillCells } = useTimelineFillController({
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    contract: timelineContract,
    enqueueSaveWork: enqueueTimelineSaveWork,
    finishSave,
    getVisibleFieldKeys: () =>
      new Set(
        timelineAnchorColumnsRef.current.map((column) => column.fieldKey),
      ),
    groupBy: queryState.groupBy,
    interactionMode,
    loadRows: loadRowsRef.current,
    port: mutationCommands.bulk,
    resolvePendingSocketTxn,
    restoreFocusAnchor: restoreTimelineFocusAnchor,
    rowsRef,
    setError: setRefreshError,
    trackPendingSocketTxn,
  }).commands;

  const focusDraftRow = useCallback(() => {
    const draftSummary = document.querySelector<HTMLInputElement>(
      dataTestIdSelector(draftCellTestId("timeline.activity_synopsis_text")),
    );
    draftSummary?.focus({ preventScroll: false });
  }, []);

  const handleCreateBlankDraftRow = useCallback(
    (row: WorkbookRow) => {
      const activeRow =
        rowsRef.current.find((candidate) => candidate.key === row.key) ?? row;
      queueScalarSave(activeRow.key, "activitySynopsisText", {
        allowZeroFieldCreate: true,
        continueOnFreshDraft: true,
        preserveInputFocus: false,
        surface: "grid",
      });
    },
    [queueScalarSave],
  );

  const handleCollectionInputChange = useCallback(
    (focusKey: string, value: string) => {
      if (
        pendingSavesRefsRef.current.collectionKeyboardCommitRef.current.get(
          focusKey,
        ) !== value
      ) {
        pendingSavesRefsRef.current.collectionKeyboardCommitRef.current.delete(
          focusKey,
        );
      }
    },
    [],
  );

  const {
    renderTimelineCollectionInput,
    renderTimelineInspectorEditor,
    timelineColumns,
  } = useTimelineWorkbookRenderers({
    activeCollectionInputKey,
    commitScalarGridEdit,
    conflictQueue,
    editorDraftRegistry,
    editingPresenceForCell,
    entityIndex,
    gridShellWidth: timelineGridShellWidth,
    handleBlur,
    handleCollectionInputChange,
    handleCollectionKeyDown,
    handleEditModePresence,
    handleKeyDown,
    handlePaste,
    handleSelectMention,
    handleSelectRow,
    openInspectorForRow,
    queueCollectionSave,
    readOnly: interactionMode.kind === "read_only",
    rowGutterWidth: timelineRowGutterWidth,
    setActiveCollectionInputKey,
    setActiveConflictKey,
    timelineContract,
    updateTimelineSurfaceFocusAnchor,
  });
  const visibleTimelineColumns = useMemo(
    () =>
      applyWorkbookLayoutToColumns(
        timelineContract,
        timelineColumns,
        layoutState,
      ),
    [layoutState, timelineColumns],
  );

  const timelineRowGutter = useMemo(
    () => ({
      label: "",
      width: timelineRowGutterWidth,
      minWidth: timelineRowGutterWidth,
    }),
    [],
  );

  const timelineGrid = useMemo(
    () =>
      buildTimelineGridRows({
        presenceForRow,
        renderDraftGutterContent: (row) => (
          <DraftRowCreateButton
            row={row}
            onCreate={handleCreateBlankDraftRow}
            onFilesSelected={handleTimelineEvidenceFiles}
          />
        ),
        renderSavedGutterContent: ({ ordinal, presences, recordId }) => (
          <WorkbookRowGutterContent
            ordinal={ordinal}
            presences={presences}
            recordId={recordId}
          />
        ),
        rows,
      }),
    [
      handleCreateBlankDraftRow,
      handleTimelineEvidenceFiles,
      presenceForRow,
      rows,
    ],
  );
  const timelineGridRows = timelineGrid.recordRows;
  const timelineDraftRow = timelineGrid.draftRow;
  useLayoutEffect(() => {
    timelineAnchorColumnsRef.current = visibleTimelineColumns;
  }, [visibleTimelineColumns]);

  const getTimelineGroupLabel = useCallback(
    (row: WorkbookRow, fieldKey: string) => timelineGroupLabel(row, fieldKey),
    [],
  );
  const getTimelineGroupRowTestId = useCallback(
    (fieldKey: string, value: string) =>
      gridGroupRowTestId(timelineViewSchemaId, fieldKey, value),
    [],
  );

  const {
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditors: renderInspectorRelationshipEditors,
    renderRowHistorySection,
    renderWorkflowSection: renderCreateRelatedWorkflowSection,
  } = useTimelineWorkbookInspectorSections({
    cancelCreateRelatedWorkflow: cancelInspectorFeatureAction,
    confirmRowHistoryPendingAction,
    createRelatedWorkflow,
    currentHistoryRecordId,
    handleTimelineEvidenceFiles,
    inspectorHistorySubject,
    openRowHistory,
    previewRowHistoryDeleteRestore,
    previewRowHistoryRollback,
    renderTimelineCollectionInput,
    renderTimelineInspectorEditor,
    rowHistory,
    rowHistoryPendingAction,
    setRowHistoryPendingAction,
    submitCreateRelatedWorkflow,
    timelineCreateRelatedReferenceOptions,
    updateCreateRelatedWorkflowDraft,
  });

  useTimelineInspectorEscape({
    activeConflict,
    clearRowHistory,
    isInspectorOpen,
    restoreTimelineFocusAnchor,
    setInspectorMessage,
    setIsInspectorOpen,
    setSelectedMentionRef,
    setSelectedRowId,
    workbookFocusAnchorRef,
  });

  const pendingQueueDisplayMessage =
    timelinePendingQueueMessage(pendingQueueSnapshot);
  const conflictStatusIsActionable =
    pendingQueueSnapshot.blockedEdit !== null ||
    pendingQueueSnapshot.overflowMessage !== null ||
    Object.keys(conflictQueue).length > 0;
  const activateConflictStatus = useCallback(() => {
    if (
      pendingQueueSnapshot.blockedEdit !== null ||
      pendingQueueSnapshot.overflowMessage !== null
    ) {
      recoveryPanelRef.current?.focus();
      return;
    }
    const firstConflictKey = Object.keys(conflictQueue)[0];
    if (firstConflictKey === undefined) {
      return;
    }
    activateConflict(firstConflictKey);
    window.requestAnimationFrame(() => {
      const summary = document.querySelector<HTMLElement>(
        dataTestIdSelector(workbookConflictSummaryTestId()),
      );
      summary?.focus();
    });
  }, [
    conflictQueue,
    activateConflict,
    pendingQueueSnapshot.blockedEdit,
    pendingQueueSnapshot.overflowMessage,
  ]);
  const visibleRefreshError =
    refreshError !== null && refreshError !== pendingQueueDisplayMessage
      ? refreshError
      : null;
  const timelineLoadState: WorkbookQueryLoadState = isInitialLoading
    ? { generationKey: initialLoadGenerationKey, kind: "initial_loading" }
    : loadError !== null
      ? loadAccessLost
        ? { kind: "permission_denied", message: loadError }
        : { kind: "unavailable", message: loadError }
      : isRefreshing
        ? { kind: "refreshing" }
        : visibleRefreshError !== null
          ? { kind: "stale_error", message: visibleRefreshError }
          : { kind: "ready" };
  const timelineDataState = workbookGridDataState({
    emptyAction:
      interactionMode.kind === "editable"
        ? { label: "Add row", onInvoke: focusDraftRow }
        : undefined,
    emptyMessage: "No Timeline records have been added.",
    loadState: timelineLoadState,
    onClearFilters: () => {
      setQueryState(emptyWorkbookQueryState());
      setFilterDraft(defaultFilterDraft(timelineContract));
    },
    onRetry: () => void loadRows({ showLoading: true }),
    queryState,
    rowCount: timelineGridRows.length,
    surfaceLabel: timelineContract.title,
  });

  return (
    <WorkbookSurfaceLayout
      chromeMode={chromeMode}
      inspector={
        isInspectorOpen ? (
          <TimelineWorkbookInspector
            canManageMentions={canManageMentions}
            currentHistoryDeleted={currentHistoryDeleted}
            entityIndex={entityIndex}
            getRelationshipLabel={timelineRelationshipLabel}
            hostEntities={hostEntities}
            identityEntities={identityEntities}
            inspectorConfig={timelineInspectorConfig}
            inspectorMessage={inspectorMessage}
            inspectorMentions={inspectorMentions}
            onClose={() => {
              setIsInspectorOpen(false);
              clearRowHistory();
              cancelInspectorFeatureAction();
            }}
            onFeatureAction={handleInspectorFeatureAction}
            isFeatureActionSupported={supportsTimelineInspectorFeature}
            onResolveTargetChange={handleResolveTargetChange}
            onSelectMention={handleSelectMention}
            onSetInspectorMessage={setInspectorMessage}
            onCreateEntityFromMention={createEntityFromMention}
            onSubmitMentionAction={submitMentionAction}
            renderEvidenceAttachSection={renderEvidenceAttachSection}
            renderInspectorFieldEditors={renderInspectorFieldEditors}
            renderRelationshipEditors={renderInspectorRelationshipEditors}
            renderWorkflowSection={() => renderCreateRelatedWorkflowSection()}
            renderPanelSupplement={(panelId) =>
              indicatorInspectorHandler?.panelId === panelId &&
              selectedRow?.recordId &&
              selectedRow.rowVersion !== null ? (
                <IndicatorInspectorWorkflow
                  action={indicatorInspectorHandler.action}
                  port={indicatorWorkflow}
                  rowVersion={selectedRow.rowVersion}
                  sourceFields={timelineObservationSourceFields}
                  sourceRecordId={selectedRow.recordId}
                  onMutationCommitted={() =>
                    loadRowsRef.current({ showLoading: false })
                  }
                />
              ) : null
            }
            renderRowHistorySection={renderRowHistorySection}
            rowHistoryRecordId={
              currentHistoryDeleted ? currentHistoryRecordId : null
            }
            selectedMention={selectedMention}
            selectedResolveTargetId={selectedResolveTargetId}
            selectedRow={selectedRow}
          />
        ) : undefined
      }
      onRequestInspectorClose={() => {
        setIsInspectorOpen(false);
        clearRowHistory();
        cancelInspectorFeatureAction();
      }}
      primaryGrid={
        <TimelineGridSurface
          activeRecordId={selectedRowId}
          bulkSelection={timelineBulkSelection}
          clipboardPaste={timelineClipboardPaste}
          columns={visibleTimelineColumns}
          columnWidths={layoutState.columnWidths}
          dataState={timelineDataState}
          density={density}
          getCellState={getTimelineCellState}
          getGroupLabel={getTimelineGroupLabel}
          getGroupRowTestId={getTimelineGroupRowTestId}
          getRowState={getTimelineRowState}
          groupBy={queryState.groupBy}
          interactionMode={interactionMode}
          onActiveCellChange={(anchor) =>
            updateTimelineSurfaceFocusAnchor(
              anchor === null ? null : gridCoreRecordId(anchor),
              anchor?.fieldKey ?? "",
            )
          }
          onColumnReorder={handleColumnReorder}
          onColumnWidthChange={handleColumnWidthChange}
          onFillCells={handleFillCells}
          onSortChange={handleQuerySortChange}
          ref={timelineGridHandleRef}
          rowGutter={timelineRowGutter}
          rows={rows}
          sort={queryState.sort}
          style={timelineGridShellStyle}
          timelineDraftRow={timelineDraftRow}
          timelineGridRows={timelineGridRows}
          shellRef={gridShellRef}
          onSelectRecord={handleSelectRow}
        />
      }
      statusStrip={
        <WorkbookStatusStrip
          activeSheetPresenceRecords={activeSheetPresenceRecords}
          inFlightCount={pendingQueueSnapshot.inFlightCount}
          queuedCount={pendingQueueSnapshot.queuedCount}
          saveState={commonMutationSnapshot.primaryLabel}
          saveStateSecondaryMessage={commonMutationSnapshot.secondaryMessage}
          showPresence={showStatusPresence}
          onActivateConflict={
            conflictStatusIsActionable ? activateConflictStatus : undefined
          }
          workbookFocusAnchor={workbookFocusAnchor}
        />
      }
      testId={timelineMutationSubstrateReadyTestId()}
      viewBar={
        <WorkbookViewBar
          addRowDisabled={interactionMode.kind === "read_only"}
          queryControls={
            renderInlineQueryControls ? (
              <WorkbookGridControls
                chromeMode={chromeMode}
                contract={timelineContract}
                defaultFilterPopoverOpen
                filterDraft={filterDraft}
                layoutState={layoutState}
                onApplyFilter={applyQueryFilter}
                onClearAll={() => {
                  setQueryState(emptyWorkbookQueryState());
                  setFilterDraft(defaultFilterDraft(timelineContract));
                }}
                onFilterDraftChange={setFilterDraft}
                onGroupByChange={handleQueryGroupByChange}
                onColumnHiddenChange={handleColumnHiddenChange}
                onColumnMove={handleColumnMove}
                onResetColumns={handleResetColumns}
                onRemoveFilter={(fieldKey) => {
                  setQueryState((current) =>
                    removeFilterField(current, fieldKey),
                  );
                }}
                onSortChange={handleQuerySortChange}
                queryState={queryState}
                surface={timelineViewSchemaId}
              />
            ) : (
              viewBarQueryControls
            )
          }
          savedViewControls={savedViewSelector}
          supplementalControls={
            selectedTimelineRecordIds.size === 0 ? undefined : (
              <fieldset style={bulkActionFieldsetStyle}>
                <legend style={visuallyHiddenStyle}>
                  Timeline bulk record actions
                </legend>
                <span aria-live="polite">
                  {selectedTimelineRecordIds.size} selected
                </span>
                <input
                  aria-label="Tag for selected Timeline records"
                  disabled={!canBulkTag || selectedTimelineRecordIds.size === 0}
                  placeholder="Tag selected"
                  type="text"
                  value={bulkTagName}
                  onChange={(event) => {
                    setBulkTagName(event.target.value);
                  }}
                />
                <button
                  disabled={!canSubmitBulkTag}
                  type="button"
                  onClick={() => {
                    void assignTagToSelectedRows();
                  }}
                >
                  Assign tag
                </button>
                {bulkTagMessage === null ? null : (
                  <span
                    aria-live={
                      bulkTagMessage.kind === "error" ? "assertive" : "polite"
                    }
                    role={bulkTagMessage.kind === "error" ? "alert" : "status"}
                  >
                    {bulkTagMessage.message}
                  </span>
                )}
              </fieldset>
            )
          }
          onAddRow={focusDraftRow}
          onInspectorToggle={() => {
            setIsInspectorOpen(true);
          }}
          surface={timelineViewSchemaId}
        />
      }
      viewSchemaId={timelineViewSchemaId}
      workAreaAriaLabel="Timeline row interaction layer"
      workAreaOverlays={
        <>
          <TimelineWorkbookNotices
            autoResolutionNotices={autoResolutionNotices}
            entityIndex={entityIndex}
            inspectorOpen={isInspectorOpen}
            onReviewAutoResolution={handleSelectMention}
            onUndoAutoResolution={handleUndoAutoResolutionNotice}
            pendingQueueSnapshot={pendingQueueSnapshot}
            recoveryPanelRef={recoveryPanelRef}
          />
          {rowContextMenu === null ? null : (
            <TimelineRowContextMenu
              position={rowContextMenu.position}
              replacementDraft={
                activeRowContextMenuRow === null
                  ? ""
                  : (replacementDrafts[activeRowContextMenuRow.key] ?? "")
              }
              row={activeRowContextMenuRow}
              onClose={closeRowContextMenu}
              onInspectRow={openInspectorForRow}
              onMarkReviewed={(rowKey) => {
                queueAction(rowKey, "mark-reviewed");
              }}
              onOpenHistory={openRowHistory}
              onReplacementDraftChange={(rowKey, value) => {
                changeReplacementDraft(rowKey, value);
              }}
              onSupersede={(rowKey) => {
                queueAction(rowKey, "supersede");
              }}
            />
          )}
        </>
      }
      onWorkAreaContextMenu={handleTimelineGridContextMenu}
      onWorkAreaKeyDown={handleTimelineWorkAreaKeyDown}
    />
  );
}

export function TimelineWorkbook(props: TimelineWorkbookProps) {
  const { runtime } = props;
  return (
    <TimelineCollaborationBoundary
      apiBase={runtime.incident.apiBase}
      attachSession={runtime.attachCollaborationSession}
      incidentId={runtime.incident.id}
      projection={runtime.collaborationProjection}
      sheetRef={runtime.incident.sheetRef}
    >
      <TimelineWorkbookContent {...runtime} key={runtime.incident.id} />
    </TimelineCollaborationBoundary>
  );
}
