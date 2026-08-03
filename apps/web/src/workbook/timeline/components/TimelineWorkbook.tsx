import type {
  GridCellAnchor,
  GridCellPasteIntent,
  GridCellStateInput,
  GridColumn,
  GridDataRow,
  GridFillIntent,
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
import {
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import {
  type KeyboardEvent as ReactKeyboardEvent,
  type SetStateAction,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import { WorkbookGridControls } from "../../components/WorkbookGridControls";
import { WorkbookRowGutterContent } from "../../components/WorkbookPresenceMarkers";
import { WorkbookStatusStrip } from "../../components/WorkbookStatusStrip";
import { WorkbookViewBar } from "../../components/WorkbookViewBar";
import type {
  WorkbookContinuityAnchor,
  WorkbookContinuityToken,
} from "../../continuity/workbookContinuityPort";
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
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import { workbookClipboardPasteContract } from "../../utils/workbookClipboard";
import {
  mapWorkbookKeyboardCommand,
  type WorkbookKeyboardCommand,
} from "../../utils/workbookKeyboard";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import { createTimelineClipboardPasteAdapter } from "../adapters/createTimelineClipboardPasteAdapter";
import { createTimelineEvidenceAttachmentAdapter } from "../adapters/createTimelineEvidenceAttachmentAdapter";
import { createTimelineHistoryAdapter } from "../adapters/createTimelineHistoryAdapter";
import { createTimelineMentionAdapter } from "../adapters/createTimelineMentionAdapter";
import { createTimelineRecordActionAdapter } from "../adapters/createTimelineRecordActionAdapter";
import { createTimelineRowMutationEditorAdapter } from "../adapters/createTimelineRowMutationEditorAdapter";
import { useTimelineBulkTagController } from "../bulk/useTimelineBulkTagController";
import { TimelineCollaborationBoundary } from "../collaboration/TimelineCollaborationBoundary";
import { useTimelineCollaborationBindings } from "../collaboration/useTimelineCollaborationBindings";
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
import {
  useTimelineInspectorEscape,
  useTimelineInspectorLifecycle,
  useTimelineInspectorRowInteractions,
  useTimelineInspectorSelection,
} from "../hooks/useTimelineInspectorSelection";
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
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";
import {
  type AutoResolutionNotice,
  reconcileDismissedMentionsForRow,
} from "../models/workbookMentionChips";
import {
  type CollectionDraftKey,
  type CollectionFieldKey,
  createDraftRow,
  inputFocusKey,
  type RowValues,
  readTimelineCellValue,
  type TimelineScalarEditorSurface,
  timelineGroupLabel,
  timelineRelationshipLabel,
  timelineScalarBindingForField,
  timelineScalarBindings,
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

import type { WorkbookPresenceDraft } from "../../collaboration/workbookCollaborationMessages";
import {
  DraftRowCreateButton,
  mentionChipStateForItem,
} from "./TimelineCellEditors";
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
const createRelatedTargetViewSchemaIds = [
  notesViewSchemaId,
  taskRequestsViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  commLogViewSchemaId,
  handoffViewSchemaId,
  statusReviewViewSchemaId,
  lessonViewSchemaId,
] as const;
const createRelatedTargetContracts = new Map<string, ViewContract>(
  createRelatedTargetViewSchemaIds.map((viewSchemaId) => [
    viewSchemaId,
    requireViewContract(viewSchemaId),
  ]),
);
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
  const [currentPresence, setCurrentPresence] = useState<WorkbookPresenceDraft>(
    {
      fieldKey: null,
      mode: "viewing",
      recordId: null,
    },
  );
  const currentPresenceRef = useRef<WorkbookPresenceDraft>({
    fieldKey: null,
    mode: "viewing",
    recordId: null,
  });
  currentPresenceRef.current = currentPresence;
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
    targetContracts: createRelatedTargetContracts,
  });
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
    cancelCreateRelatedWorkflow,
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

  const sendPresenceUpdate = timelineCollaboration.commands.publishPresence;
  const activeSheetPresenceRecords =
    timelineCollaboration.snapshot.activeSheetPresenceRecords;
  const presenceForRow = useCallback(
    (recordId: string | null) =>
      recordId === null
        ? []
        : activeSheetPresenceRecords.filter(
            (presence) => presence.record_id === recordId,
          ),
    [activeSheetPresenceRecords],
  );
  const editingPresenceForCell = useCallback(
    (recordId: string | null, fieldKey: string) =>
      recordId === null
        ? []
        : activeSheetPresenceRecords.filter(
            (presence) =>
              presence.record_id === recordId &&
              presence.field_key === fieldKey &&
              presence.mode === "editing",
          ),
    [activeSheetPresenceRecords],
  );
  const handleEditModePresence = useCallback(
    (recordId: string | null, fieldKey: string, editing: boolean) => {
      const next = editing
        ? { fieldKey, mode: "editing" as const, recordId }
        : {
            fieldKey: null,
            mode: "viewing" as const,
            recordId: recordId ?? currentPresenceRef.current.recordId,
          };
      currentPresenceRef.current = next;
      setCurrentPresence(next);
      sendPresenceUpdate(next);
    },
    [sendPresenceUpdate],
  );

  const timelineInspectorRowInteractions = useTimelineInspectorRowInteractions({
    currentPresenceRef,
    rows,
    rowsRef,
    selectedRowId,
    sendPresenceUpdate,
    setCurrentPresence,
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

  const { createEntityFromMention, submitMentionAction } =
    useTimelineMentionActions({
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

  const handleTimelineCommonKeyboardCommand = useCallback(
    (command: WorkbookKeyboardCommand, anchor: GridCellAnchor | null) => {
      if (anchor === null) {
        return false;
      }
      const recordId = gridCoreRecordId(anchor);
      if (recordId === null) return false;
      if (command.kind === "open-history") {
        openRowHistory(recordId);
        return true;
      }
      if (command.kind === "close-inspector") {
        setSelectedRowId(null);
        setSelectedMentionRef(null);
        setInspectorMessage(null);
        clearRowHistory();
        restoreTimelineFocusAnchor(anchor);
        return true;
      }
      if (command.kind === "preview-linked-evidence") {
        setSelectedRowId(recordId);
        setIsInspectorOpen(true);
        setInspectorMessage(
          "Linked evidence preview is unavailable for this row.",
        );
        restoreTimelineFocusAnchor(anchor);
        return true;
      }
      return false;
    },
    [
      clearRowHistory,
      openRowHistory,
      restoreTimelineFocusAnchor,
      setIsInspectorOpen,
      setSelectedMentionRef,
      setInspectorMessage,
      setSelectedRowId,
    ],
  );

  const handleKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
    ) => {
      const priorGridAnchor = workbookFocusAnchorRef.current;
      if (
        surface === "inspector" &&
        event.key === "Escape" &&
        priorGridAnchor?.viewSchemaId === timelineViewSchemaId
      ) {
        event.preventDefault();
        restoreTimelineFocusAnchor(priorGridAnchor);
        return;
      }
      const binding = timelineScalarBindings.find(
        (candidate) => candidate.key === focusField,
      );
      const fieldKey = binding?.fieldKey ?? focusField;
      const anchor = currentTimelineAnchorFor(rowKey, fieldKey);
      const command = mapWorkbookKeyboardCommand(event, {
        closeInspector:
          anchor !== null &&
          (surface === "inspector" ||
            selectedRowId !== null ||
            rowHistory.recordId !== null ||
            rowHistory.status !== "idle"),
        history: anchor !== null,
        previewLinkedEvidence:
          anchor !== null && event.currentTarget.value === "",
      });
      if (command.preventDefault) {
        event.preventDefault();
      }
      const adapterOwnsRange =
        surface === "grid" &&
        command.kind === "navigate" &&
        command.intent.shiftKey &&
        command.intent.key.startsWith("Arrow");
      if (adapterOwnsRange) {
        return;
      }
      if (
        surface === "grid" &&
        command.kind === "navigate" &&
        anchor !== null &&
        command.intent.key === "Tab"
      ) {
        queueScalarSave(
          rowKey,
          focusField,
          {
            continueOnFreshDraft: true,
            preserveInputFocus: false,
            surface,
          },
          event.currentTarget.value,
        );
        return;
      }
      if (
        command.kind === "navigate" &&
        anchor === null &&
        (command.intent.key === "Enter" || command.intent.key === "Tab")
      ) {
        recordWorkbookTiming("key_commit_accepted", {
          field: focusField,
          key: command.intent.key,
          rowKey,
          surface,
        });
        queueScalarSave(
          rowKey,
          focusField,
          {
            continueOnFreshDraft: true,
            preserveInputFocus: true,
            surface,
          },
          event.currentTarget.value,
        );
        return;
      }
      if (command.kind === "navigate" && anchor !== null) {
        if (command.intent.key === "Enter" || command.intent.key === "Tab") {
          recordWorkbookTiming("key_commit_accepted", {
            field: focusField,
            key: command.intent.key,
            recordId: gridCoreRecordId(anchor) ?? "",
            rowKey,
            surface,
          });
          queueScalarSave(
            rowKey,
            focusField,
            {
              continueOnFreshDraft: true,
              preserveInputFocus: false,
              surface,
            },
            event.currentTarget.value,
          );
        }
        navigateTimelineFocusAnchor(anchor, command.intent);
        return;
      }
      if (handleTimelineCommonKeyboardCommand(command, anchor)) {
        return;
      }
    },
    [
      currentTimelineAnchorFor,
      handleTimelineCommonKeyboardCommand,
      navigateTimelineFocusAnchor,
      queueScalarSave,
      rowHistory.recordId,
      rowHistory.status,
      restoreTimelineFocusAnchor,
      selectedRowId,
    ],
  );

  const handleCollectionKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement>,
      rowKey: string,
      fieldKey: CollectionFieldKey,
      draftKey: CollectionDraftKey,
    ) => {
      const anchor = currentTimelineAnchorFor(rowKey, fieldKey);
      const command = mapWorkbookKeyboardCommand(event, {
        closeInspector:
          anchor !== null &&
          (selectedRowId !== null ||
            rowHistory.recordId !== null ||
            rowHistory.status !== "idle"),
        history: anchor !== null,
        previewLinkedEvidence:
          anchor !== null && event.currentTarget.value === "",
        quickLink:
          anchor !== null &&
          (fieldKey === "timeline.host_refs" ||
            fieldKey === "timeline.identity_refs"),
      });
      if (command.preventDefault) {
        event.preventDefault();
      }
      if (
        command.kind === "navigate" &&
        command.intent.shiftKey &&
        command.intent.key.startsWith("Arrow")
      ) {
        return;
      }
      if (
        command.kind === "navigate" &&
        command.intent.key === "Tab" &&
        anchor !== null
      ) {
        queueCollectionSave(
          rowKey,
          fieldKey,
          draftKey,
          event.currentTarget.value,
          "keyboard",
        );
        return;
      }
      if (command.kind === "navigate" && anchor !== null) {
        if (command.intent.key === "Enter" || command.intent.key === "Tab") {
          queueCollectionSave(
            rowKey,
            fieldKey,
            draftKey,
            event.currentTarget.value,
            "keyboard",
          );
        }
        navigateTimelineFocusAnchor(anchor, command.intent);
        return;
      }
      if (handleTimelineCommonKeyboardCommand(command, anchor)) {
        return;
      }
      if (command.kind === "quick-link" && anchor !== null) {
        const recordId = gridCoreRecordId(anchor);
        if (recordId === null) return;
        const row = rowsRef.current.find(
          (candidate) => candidate.recordId === recordId,
        );
        const mention =
          row === undefined
            ? undefined
            : [
                ...row.collectionValues.hostRefs,
                ...row.collectionValues.identityRefs,
              ].find((item) => item.itemKind !== "resolved_ref");
        if (mention !== undefined) {
          setSelectedRowId(recordId);
          setIsInspectorOpen(true);
          setSelectedMentionRef(mention.itemRef);
          setInspectorMessage(null);
        } else {
          setSelectedRowId(recordId);
          setIsInspectorOpen(true);
          setInspectorMessage(
            "No unresolved mention is available for quick link.",
          );
        }
        return;
      }
      if (event.key === "Enter" || event.key === "Tab") {
        event.preventDefault();
        queueCollectionSave(
          rowKey,
          fieldKey,
          draftKey,
          event.currentTarget.value,
          "keyboard",
        );
      }
    },
    [
      currentTimelineAnchorFor,
      handleTimelineCommonKeyboardCommand,
      navigateTimelineFocusAnchor,
      queueCollectionSave,
      rowHistory.recordId,
      rowHistory.status,
      selectedRowId,
      setIsInspectorOpen,
      setSelectedMentionRef,
      setInspectorMessage,
      setSelectedRowId,
    ],
  );

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

  const handleFillCells = useCallback(
    (intent: GridFillIntent) => {
      const field = timelineContract.fieldMap[intent.source.fieldKey];
      const binding = timelineScalarBindingForField(intent.source.fieldKey);
      const sourceRow = rowsRef.current.find(
        (row) => row.recordId === gridCoreRecordId(intent.source),
      );
      const targets = intent.targets.filter(
        (target) =>
          gridCoreRecordId(target) !== gridCoreRecordId(intent.source),
      );
      const validTargets =
        queryState.groupBy === null &&
        field?.gridEditable === true &&
        binding !== null &&
        sourceRow !== undefined &&
        sourceRow.rowVersion ===
          intent.source.mutationIdentity.baseRowVersion &&
        sourceRow.pendingSignature === null &&
        targets.length > 0 &&
        targets.every((target) => {
          const row = rowsRef.current.find(
            (candidate) => candidate.recordId === gridCoreRecordId(target),
          );
          return (
            target.surface.kind === "view_schema" &&
            target.surface.viewSchemaId === timelineViewSchemaId &&
            target.fieldKey === intent.source.fieldKey &&
            row !== undefined &&
            row.rowVersion === target.mutationIdentity.baseRowVersion &&
            row.pendingSignature === null
          );
        });
      if (!validTargets || binding === null || sourceRow === undefined) {
        setRefreshError(
          "Fill was rejected because one or more targets are unavailable or stale.",
        );
        return;
      }
      const value = stringifyGridValue(
        readTimelineCellValue(sourceRow.rawRow, binding.fieldKey),
      );
      const viewportContinuityToken = beginViewportContinuity({
        kind: "scroll-only",
      });
      beginSave();
      enqueueTimelineSaveWork(async () => {
        const result = await mutationCommands.bulk.fillDown({
          fieldKey: intent.source.fieldKey,
          onClientTxnId: trackPendingSocketTxn,
          value,
          targets: targets.map((target) => ({
            recordId: gridCoreRecordId(target) ?? "",
            baseRowVersion: target.mutationIdentity.baseRowVersion,
          })),
        });
        resolvePendingSocketTxn(result.clientTxnId);
        if (result.outcome.kind === "rejected") {
          clearViewportContinuity(viewportContinuityToken);
          setRefreshError(result.outcome.failure.message);
          finishSave("Conflict");
          return;
        }
        await loadRowsRef.current({
          showLoading: false,
          viewportContinuityToken,
        });
        restoreTimelineFocusAnchor(intent.source);
        finishSave("Saved");
      });
    },
    [
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueTimelineSaveWork,
      finishSave,
      mutationCommands.bulk,
      queryState.groupBy,
      resolvePendingSocketTxn,
      restoreTimelineFocusAnchor,
      setRefreshError,
      trackPendingSocketTxn,
    ],
  );

  const focusDraftRow = useCallback(() => {
    const draftSummary = document.querySelector<HTMLInputElement>(
      dataTestIdSelector(draftCellTestId("timeline.activity_synopsis_text")),
    );
    draftSummary?.focus({ preventScroll: false });
  }, []);

  function handleUndoAutoResolutionNotice(notice: AutoResolutionNotice) {
    const row = rowsRef.current.find(
      (candidate) => candidate.recordId === notice.rowRecordId,
    );
    const activeItem =
      row === undefined
        ? undefined
        : [
            ...row.collectionValues.hostRefs,
            ...row.collectionValues.identityRefs,
          ].find(
            (item) =>
              item.itemRef === notice.itemRef &&
              item.itemKind === "resolved_ref",
          );
    if (activeItem && row?.recordId) {
      submitMentionAction(
        {
          rowRecordId: row.recordId,
          fieldKey: notice.fieldKey,
          entityType: activeItem.entityType,
          itemRef: activeItem.itemRef,
          rawText: activeItem.rawText,
          resolvedRecordId: activeItem.resolvedRecordId,
          mentionRowVersion: activeItem.mentionRowVersion,
          resolutionMethod: activeItem.resolutionMethod,
          autoResolved: activeItem.autoResolved,
          status: "resolved",
          chipState: mentionChipStateForItem(activeItem),
          anchor: {
            recordId: row.recordId,
            fieldKey: notice.fieldKey,
            itemRef: activeItem.itemRef,
            entityMentionId: activeItem.itemRef.startsWith("entity_mention:")
              ? activeItem.itemRef.slice("entity_mention:".length) || null
              : null,
            targetEntityRecordId: activeItem.resolvedRecordId,
          },
          sourceKind: "entity_mention",
          isActiveRelationshipValue: true,
          priorTargetEntityRecordId: null,
          displayText: activeItem.displayText,
          provenance: activeItem.provenance,
          confidence: activeItem.confidence,
          matchedAliasText: activeItem.matchedAliasText,
        },
        "revert_to_unresolved",
      );
    }
  }

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
    cancelCreateRelatedWorkflow,
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
    ? { kind: "initial_loading" }
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
              cancelCreateRelatedWorkflow();
            }}
            onFeatureAction={beginCreateRelatedWorkflow}
            onResolveTargetChange={handleResolveTargetChange}
            onSelectMention={handleSelectMention}
            onSetInspectorMessage={setInspectorMessage}
            onCreateEntityFromMention={createEntityFromMention}
            onSubmitMentionAction={submitMentionAction}
            renderEvidenceAttachSection={renderEvidenceAttachSection}
            renderInspectorFieldEditors={renderInspectorFieldEditors}
            renderRelationshipEditors={renderInspectorRelationshipEditors}
            renderWorkflowSection={renderCreateRelatedWorkflowSection}
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
        cancelCreateRelatedWorkflow();
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
      onWorkAreaKeyDown={handleTimelineGridContextKeyDown}
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
