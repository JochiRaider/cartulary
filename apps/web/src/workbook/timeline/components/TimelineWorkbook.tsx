import type {
  GridCellAnchor,
  GridCellMutationIntent,
  GridCellStateInput,
  GridColumn,
  GridDataRow,
  GridDensity,
  GridFillIntent,
  GridHandle,
  GridInteractionMode,
  GridRowStateInput,
} from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  draftCellTestId,
  gridGroupRowTestId,
  timelineMutationSubstrateReadyTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import {
  type Dispatch,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactNode,
  type SetStateAction,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { flushSync } from "react-dom";
import {
  IncidentCollaborationBoundary,
  useIncidentCollaborationSession,
} from "../../../collaboration/IncidentCollaborationSession";
import { apiPath } from "../../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  workbookLoadFailureIsAccessLoss,
} from "../../../services/workbookApi";
import { WorkbookGridControls } from "../../components/WorkbookGridControls";
import { WorkbookSheetToolbar } from "../../components/WorkbookSheetToolbar";
import { WorkbookStatusStrip } from "../../components/WorkbookStatusStrip";
import { WorkbookSurfaceFrame } from "../../components/WorkbookSurfaceFrame";
import { useIncidentMemberReferenceOptions } from "../../hooks/useOwnerReferenceOptions";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../../models/workbookGridState";
import { selectInspectorConfig } from "../../models/workbookInspectorModel";
import {
  applyWorkbookLayoutToColumns,
  defaultWorkbookLayoutState,
  moveWorkbookColumn,
  reorderWorkbookColumns,
  setWorkbookColumnHidden,
  setWorkbookColumnWidth,
  type WorkbookResolvedLayoutState,
} from "../../models/workbookLayout";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  removeFilterField,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { WorkbookSheetRef } from "../../models/workbookStartup";
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
import { parseClipboardTable } from "../../utils/workbookClipboard";
import type { WorkbookFocusAnchor } from "../../utils/workbookGridFocus";
import {
  mapWorkbookKeyboardCommand,
  type WorkbookKeyboardCommand,
} from "../../utils/workbookKeyboard";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import { useTimelineClipboardPasteController } from "../hooks/useTimelineClipboardPasteController";
import { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { useTimelineConflictResolverCoordinator } from "../hooks/useTimelineConflictResolverCoordinator";
import { useTimelineConflicts } from "../hooks/useTimelineConflicts";
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
import { useTimelineLiveUpdateController } from "../hooks/useTimelineLiveUpdateController";
import { useTimelineLiveUpdates } from "../hooks/useTimelineLiveUpdates";
import { useTimelineMentionActions } from "../hooks/useTimelineMentionActions";
import { useTimelineMentions } from "../hooks/useTimelineMentions";
import { useTimelineMutationCommands } from "../hooks/useTimelineMutationCommands";
import { useTimelinePendingReplayController } from "../hooks/useTimelinePendingReplayController";
import { useTimelinePendingSaves } from "../hooks/useTimelinePendingSaves";
import { useTimelinePresenceProjection } from "../hooks/useTimelinePresenceProjection";
import { useTimelineRows } from "../hooks/useTimelineRows";
import {
  type LoadRowsOptions,
  useTimelineRowsLoader,
} from "../hooks/useTimelineRowsLoader";
import { useTimelineSaveStatePresentation } from "../hooks/useTimelineSaveStatePresentation";
import {
  useTimelineViewportContinuityController,
  type TimelineViewportContinuityRequest as ViewportContinuityRequest,
  type TimelineViewportContinuityTarget as ViewportContinuityTarget,
} from "../hooks/useTimelineViewportContinuityController";
import { useTimelineWorkbookRuntime } from "../hooks/useTimelineWorkbookRuntime";
import type {
  PendingReplayRuntimeMeta,
  TimelineLiveUpdateRefs,
} from "../models/timelineControllerPorts";
import {
  createTimelinePendingQueueRuntime,
  refreshBlocksTimelinePendingRecord,
} from "../models/timelinePendingReplayModel";
import { buildTimelineGridRows } from "../models/timelineRowsModel";
import type { TimelineEntityCatalogInput } from "../models/timelineViewportContinuityModel";
import {
  type AutoResolutionNotice,
  buildAutoResolutionNotices,
  type DismissedMention,
} from "../models/workbookMentionChips";
import {
  applyViewRowPatch,
  type CollectionDraftKey,
  type CollectionFieldKey,
  createDraftRow,
  type EntityApiRow,
  type FocusFieldKey,
  inputFocusKey,
  type LocalConflictState,
  normalizeTimelineFullRow,
  normalizeTimelinePatchCells,
  type RowValues,
  readTimelineCellValue,
  rowFromApi,
  type TimelinePatchCells,
  type TimelineScalarEditorSurface,
  timelineGroupLabel,
  timelineRelationshipLabel,
  timelineScalarBindingForField,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

function gridCoreRecordId(
  anchor: GridCellAnchor | GridCellMutationIntent["target"],
): string | null {
  return anchor.rowIdentity.kind === "core_record"
    ? anchor.rowIdentity.recordId
    : null;
}

import {
  createTimelineCollaborationState,
  type TimelineCollaborationAction,
  type TimelineCollaborationEffect,
} from "../services/timelineCollaborationEffects";
import type { TimelineMutationEnvelope } from "../services/timelineMutationRequests";
import type {
  RecordChangedPayload,
  TimelinePresenceDraft,
} from "../services/workbookCollaborationMessages";
import {
  DraftRowCreateButton,
  mentionChipStateForItem,
} from "./TimelineCellEditors";
import { TimelineConflictResolver } from "./TimelineConflictResolver";
import { TimelineGridSurface } from "./TimelineGridSurface";
import { TimelineRowGutterContent } from "./TimelinePresenceMarkers";
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
type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;
export type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";

const bulkActionFieldsetStyle = {
  border: 0,
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  margin: 0,
  padding: 0,
};

function rowStillHasAutoResolvedNotice(
  row: WorkbookRow,
  notice: AutoResolutionNotice,
) {
  if (row.recordId !== notice.rowRecordId) {
    return false;
  }
  const item = [
    ...row.collectionValues.hostRefs,
    ...row.collectionValues.identityRefs,
  ].find((candidate) => candidate.itemRef === notice.itemRef);
  return (
    item?.itemKind === "resolved_ref" &&
    item.autoResolved &&
    item.resolvedRecordId !== null
  );
}

export type TimelineWorkbookProps = {
  incidentId: string;
  apiBase?: string | undefined;
  currentUserId?: string | null | undefined;
  sheetRef?: WorkbookSheetRef | undefined;
  inspectorResetKey?: string | undefined;
  reloadToken?: number | undefined;
  renderInlineQueryControls?: boolean | undefined;
  savedViewSelector?: ReactNode | undefined;
  filterDraft?: FilterDraft | undefined;
  onFilterDraftChange?: FilterDraftSetter | undefined;
  onQueryStateChange?: WorkbookQueryStateSetter | undefined;
  queryState?: WorkbookQueryState | undefined;
  hostEntities?: EntityRow[];
  identityEntities?: EntityRow[];
  entityIndex?: Record<string, EntityRow>;
  currentIncidentRole?: IncidentRole | null;
  density?: GridDensity | undefined;
  layoutState?: WorkbookResolvedLayoutState | undefined;
  onColumnHiddenChange?:
    | ((fieldKey: string, hidden: boolean) => void)
    | undefined;
  onColumnMove?:
    | ((fieldKey: string, direction: "earlier" | "later") => void)
    | undefined;
  onColumnReorder?:
    | ((sourceFieldKey: string, targetFieldKey: string) => void)
    | undefined;
  onColumnWidthChange?: ((fieldKey: string, width: number) => void) | undefined;
  onResetColumns?: (() => void) | undefined;
  onRefreshEntities?: () => Promise<void> | void;
  interactionMode?: GridInteractionMode | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
};

type EntityRow = {
  entityType: "host" | "identity";
  recordId: string;
  rowVersion: number;
  label: string;
  secondaryText: string;
  state: string;
  aliasTexts: string[];
  linkedEventCount: number;
  rawRow: EntityApiRow;
  identifiers: Array<{
    key: string;
    label: string;
    value: string;
  }>;
};

function recordWorkbookTiming(
  name: string,
  details: Record<string, unknown> = {},
) {
  const probe =
    typeof window === "undefined"
      ? undefined
      : window.__cartularyWorkbookTimingProbe;
  if (probe === undefined) {
    return;
  }
  const event = {
    at: performance.now(),
    name,
    ...details,
  };
  probe.events.push(event);
  probe.mark?.(event);
}

function ensureDraftRowWithFreshIndex(
  rows: WorkbookRow[],
  nextDraftIndex: () => number,
): {
  rows: WorkbookRow[];
  draftSummaryKey: string | null;
} {
  if (rows.some((row) => row.recordId === null)) {
    return {
      rows,
      draftSummaryKey: null,
    };
  }

  const draftIndex = nextDraftIndex();
  return {
    rows: [...rows, createDraftRow(draftIndex)],
    draftSummaryKey: inputFocusKey(
      `draft-${draftIndex}`,
      "activitySynopsisText",
    ),
  };
}

function rowWithMaterializedScalarDrafts(
  row: WorkbookRow,
  draftValueForFocusKey: (focusKey: string) => string | undefined,
  preferred?: {
    readonly field: keyof RowValues;
    readonly value: string | undefined;
  },
): WorkbookRow {
  let nextValues: RowValues | null = null;
  for (const binding of timelineScalarBindings) {
    let draftValue =
      preferred?.field === binding.key ? preferred.value : undefined;
    if (draftValue === undefined) {
      for (const surface of timelineScalarEditorSurfaces) {
        draftValue = draftValueForFocusKey(
          inputFocusKey(row.key, binding.key, surface),
        );
        if (draftValue !== undefined) {
          break;
        }
      }
    }
    if (draftValue === undefined || draftValue === row.values[binding.key]) {
      continue;
    }
    nextValues ??= { ...row.values };
    nextValues[binding.key] = draftValue;
  }
  return nextValues === null ? row : { ...row, values: nextValues };
}

function pruneDismissedMentions(
  dismissedMentionsByRow: Record<string, DismissedMention[]>,
  row: WorkbookRow,
) {
  if (row.recordId === null) {
    return dismissedMentionsByRow;
  }

  const activeRefs = new Set(
    [
      ...row.collectionValues.hostRefs,
      ...row.collectionValues.identityRefs,
    ].map((item) => item.itemRef),
  );
  const next = { ...dismissedMentionsByRow };
  const current = next[row.recordId] ?? [];
  const remaining = current.filter((item) => !activeRefs.has(item.itemRef));
  if (remaining.length < 1) {
    delete next[row.recordId];
    return next;
  }
  next[row.recordId] = remaining;
  return next;
}

function TimelineWorkbookContent({
  incidentId,
  apiBase,
  currentUserId = null,
  sheetRef,
  inspectorResetKey,
  reloadToken = 0,
  renderInlineQueryControls = true,
  savedViewSelector,
  filterDraft: controlledFilterDraft,
  onFilterDraftChange,
  onQueryStateChange,
  queryState: controlledQueryState,
  hostEntities = [],
  identityEntities = [],
  entityIndex = {},
  currentIncidentRole = "",
  density = "compact",
  layoutState: controlledLayoutState,
  onColumnHiddenChange: controlledColumnHiddenChange,
  onColumnMove: controlledColumnMove,
  onColumnReorder: controlledColumnReorder,
  onColumnWidthChange: controlledColumnWidthChange,
  onResetColumns: controlledResetColumns,
  onRefreshEntities,
  interactionMode = { kind: "editable" },
  onIncidentAccessLost,
}: TimelineWorkbookProps) {
  const { clientInstanceId, connectionId } = useIncidentCollaborationSession();
  const entityCatalogInput = useMemo(
    () =>
      ({
        hostEntities,
        identityEntities,
      }) satisfies TimelineEntityCatalogInput,
    [hostEntities, identityEntities],
  );
  const timelineRuntime = useTimelineWorkbookRuntime({
    controlledFilterDraft,
    controlledQueryState,
    onFilterDraftChange,
    onQueryStateChange,
  });
  const {
    isInitialLoading,
    isRefreshing,
    loadError,
    refreshError,
    saveState,
    saveStateSecondaryMessage,
    setIsInitialLoading,
    setIsRefreshing,
    setLoadError,
    setRefreshError,
    setSaveState,
    setSaveStateSecondaryMessage,
  } = timelineRuntime.lifecycle;
  const {
    applyQueryFilter,
    filterDraft,
    handleQueryGroupByChange,
    handleQuerySortChange,
    queryState,
    setFilterDraft,
    setQueryState,
  } = timelineRuntime.query;
  const [uncontrolledLayoutState, setUncontrolledLayoutState] =
    useState<WorkbookResolvedLayoutState>(() =>
      defaultWorkbookLayoutState(timelineContract),
    );
  const layoutState = controlledLayoutState ?? uncontrolledLayoutState;
  const handleColumnHiddenChange = useCallback(
    (fieldKey: string, hidden: boolean) => {
      if (controlledColumnHiddenChange !== undefined) {
        controlledColumnHiddenChange(fieldKey, hidden);
        return;
      }
      setUncontrolledLayoutState((current) =>
        setWorkbookColumnHidden(timelineContract, current, fieldKey, hidden),
      );
    },
    [controlledColumnHiddenChange],
  );
  const handleColumnMove = useCallback(
    (fieldKey: string, direction: "earlier" | "later") => {
      if (controlledColumnMove !== undefined) {
        controlledColumnMove(fieldKey, direction);
        return;
      }
      setUncontrolledLayoutState((current) =>
        moveWorkbookColumn(timelineContract, current, fieldKey, direction),
      );
    },
    [controlledColumnMove],
  );
  const handleColumnReorder = useCallback(
    (sourceFieldKey: string, targetFieldKey: string) => {
      if (controlledColumnReorder !== undefined) {
        controlledColumnReorder(sourceFieldKey, targetFieldKey);
        return;
      }
      setUncontrolledLayoutState((current) =>
        reorderWorkbookColumns(
          timelineContract,
          current,
          sourceFieldKey,
          targetFieldKey,
        ),
      );
    },
    [controlledColumnReorder],
  );
  const handleColumnWidthChange = useCallback(
    (fieldKey: string, width: number) => {
      if (controlledColumnWidthChange !== undefined) {
        controlledColumnWidthChange(fieldKey, width);
        return;
      }
      setUncontrolledLayoutState((current) =>
        setWorkbookColumnWidth(timelineContract, current, fieldKey, width),
      );
    },
    [controlledColumnWidthChange],
  );
  const handleResetColumns = useCallback(() => {
    if (controlledResetColumns !== undefined) {
      controlledResetColumns();
      return;
    }
    setUncontrolledLayoutState(defaultWorkbookLayoutState(timelineContract));
  }, [controlledResetColumns]);
  const initialTimelineRows = useMemo(() => [createDraftRow(1)], []);
  const rowsRef = useRef<WorkbookRow[]>(initialTimelineRows);
  const draftCounterRef = useRef(2);
  const timelineRows = useTimelineRows({ draftCounterRef, rowsRef });
  const { rows } = timelineRows.snapshot;
  const { setRows } = timelineRows.commands;
  const [selectedTimelineRecordIds, setSelectedTimelineRecordIds] = useState<
    ReadonlySet<string>
  >(() => new Set());
  const [bulkTagName, setBulkTagName] = useState("");
  const [bulkTagMessage, setBulkTagMessage] = useState<{
    readonly kind: "error" | "success";
    readonly message: string;
  } | null>(null);
  const [bulkTagSubmitting, setBulkTagSubmitting] = useState(false);
  const canBulkTag =
    interactionMode.kind === "editable" &&
    (currentIncidentRole === "editor" ||
      currentIncidentRole === "reviewer" ||
      currentIncidentRole === "admin");
  useEffect(() => {
    const selectableIds = new Set(
      canBulkTag
        ? rows.flatMap((row) =>
            row.recordId !== null &&
            row.rowVersion !== null &&
            row.pendingSignature === null
              ? [row.recordId]
              : [],
          )
        : [],
    );
    setSelectedTimelineRecordIds((current) => {
      const next = new Set(
        [...current].filter((recordId) => selectableIds.has(recordId)),
      );
      return next.size === current.size ? current : next;
    });
  }, [canBulkTag, rows]);
  const conflictQueueRef = useRef<Record<string, LocalConflictState>>({});
  const timelineConflicts = useTimelineConflicts({ conflictQueueRef });
  const { activeConflictKey, conflictQueue, pasteConflictGroup } =
    timelineConflicts.snapshot;
  const { setActiveConflictKey, setConflictQueueState, setPasteConflictGroup } =
    timelineConflicts.commands;
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
  const collaborationStateRef = useRef(createTimelineCollaborationState());
  const presenceUpdateTimerRef = useRef<number | null>(null);
  const currentPresenceRef = useRef<TimelinePresenceDraft>({
    fieldKey: null,
    mode: "viewing",
    recordId: null,
  });
  const socketReconnectAfterAuthRef = useRef<(() => void) | null>(null);
  const dispatchCollaborationRef = useRef<
    (
      action: TimelineCollaborationAction,
    ) => readonly TimelineCollaborationEffect[]
  >(() => []);
  const timelineLiveUpdateRefs: TimelineLiveUpdateRefs = {
    collaborationStateRef,
    currentPresenceRef,
    dispatchCollaborationRef,
    presenceUpdateTimerRef,
    socketReconnectAfterAuthRef,
  };
  const timelineLiveUpdates = useTimelineLiveUpdates({
    refs: timelineLiveUpdateRefs,
  });
  const { currentPresence, presenceRecords } = timelineLiveUpdates.snapshot;
  const { setCurrentPresence, setPresenceRecords } =
    timelineLiveUpdates.commands;
  const timelinePendingSaves =
    useTimelinePendingSaves<PendingReplayRuntimeMeta>({
      clientInstanceId,
      incidentId,
    });
  const { pendingQueueSnapshot } = timelinePendingSaves.snapshot;
  const { setPendingQueueSnapshot } = timelinePendingSaves.commands;
  const pendingSavesRefsRef = useRef(timelinePendingSaves.refs);
  pendingSavesRefsRef.current = timelinePendingSaves.refs;
  const workbookFocusAnchorRef = useRef<WorkbookFocusAnchor | null>(null);
  const rowInputRefs = useRef(
    new Map<string, HTMLInputElement | HTMLTextAreaElement>(),
  );
  const rowInputTestIdsRef = useRef(new Map<string, string>());
  const timelineAnchorColumnsRef = useRef<readonly GridColumn<WorkbookRow>[]>(
    [],
  );
  const timelineAnchorRowsRef = useRef<readonly GridDataRow<WorkbookRow>[]>([]);
  const timelineGridHandleRef = useRef<GridHandle | null>(null);
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const [timelineGridShellWidth, setTimelineGridShellWidth] = useState(0);
  const viewportContinuityTokenRef = useRef(1);
  const timelineGridInteractionRefs: TimelineGridInteractionRefs = {
    gridHandleRef: timelineGridHandleRef,
    gridShellRef,
    rowInputRefs,
    rowInputTestIdsRef,
    timelineAnchorColumnsRef,
    timelineAnchorRowsRef,
    viewportContinuityTokenRef,
    workbookFocusAnchorRef,
  };
  const timelineGridInteractions =
    useTimelineGridInteractions<ViewportContinuityRequest>({
      refs: timelineGridInteractionRefs,
    });
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

  const [replacementDrafts, setReplacementDrafts] = useState<
    Record<string, string>
  >({});
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
    isInspectorOpen,
    inspectorMentions,
    selectedMention,
    selectedRow,
    selectedRowId,
    selectedRowWorkflowKey,
  } = timelineInspectorSelection.snapshot;
  const { setIsInspectorOpen, setSelectedRowId } =
    timelineInspectorSelection.commands;
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
  const clientTxnRef = useRef(1);
  const timelineCommittedRows = useTimelineCommittedRows({ rowsRef });
  const {
    acceptCommittedTimelineRow,
    acceptCommittedTimelineRows,
    acceptTimelineActionResult,
    acceptTimelineRecordVersion,
    beginLoad: beginTimelineRowsLoad,
    committedRowsChangedSince,
    currentCommittedTimelineRow,
    hasLoadedRows,
    isCurrentLoadSequence,
    isStaleTimelineRowVersion,
    knownTimelineRowVersion,
    latestCommittedRowVersion,
    latestCommittedTimelineRow,
    markRowsLoaded,
  } = timelineCommittedRows.commands;
  const scalarDraftValuesRef = useRef(new Map<string, string>());
  const loadRowsRef = useRef<(options: LoadRowsOptions) => Promise<void>>(
    async () => undefined,
  );

  const pruneAutoResolutionNoticesForRows = useCallback(
    (committedRows: readonly WorkbookRow[]) => {
      if (committedRows.length < 1) {
        return;
      }
      setAutoResolutionNotices((current) =>
        current.filter((notice) => {
          const row = committedRows.find(
            (candidate) => candidate.recordId === notice.rowRecordId,
          );
          return (
            row === undefined || rowStillHasAutoResolvedNotice(row, notice)
          );
        }),
      );
    },
    [setAutoResolutionNotices],
  );

  const activeSheetRef = useMemo<WorkbookSheetRef>(
    () => sheetRef ?? { kind: "view_schema", id: timelineViewSchemaId },
    [sheetRef],
  );
  const activeSheetRuntimeRef = useRef(activeSheetRef);
  activeSheetRuntimeRef.current = activeSheetRef;

  const updateTimelineSurfaceFocusAnchor = useCallback(
    (recordId: string | null, fieldKey: string) => {
      updateTimelineFocusAnchor(recordId, fieldKey, timelineViewSchemaId);
    },
    [updateTimelineFocusAnchor],
  );

  const timelineSaveStatePresentation =
    useTimelineSaveStatePresentation<PendingReplayRuntimeMeta>({
      conflictQueue,
      conflictQueueRef,
      pendingQueueSnapshot,
      pendingSavesRefsRef,
      setPendingQueueSnapshot,
      setSaveState,
      setSaveStateSecondaryMessage,
    });
  const {
    beginRefreshInFlight,
    beginSave,
    finishRefreshInFlight,
    finishSave,
    publishPendingQueueState,
    publishSaveStatePresentation,
  } = timelineSaveStatePresentation.commands;
  const {
    beginRefreshInFlightRef,
    finishRefreshInFlightRef,
    publishPendingQueueStateRef,
  } = timelineSaveStatePresentation.refs;

  useEffect(() => {
    const scope =
      pendingSavesRefsRef.current.pendingQueueRef.current.model.scope;
    if (
      scope.incidentId === incidentId &&
      scope.clientInstanceId === clientInstanceId
    ) {
      return;
    }
    pendingSavesRefsRef.current.pendingQueueRef.current =
      createTimelinePendingQueueRuntime({
        incidentId,
        clientInstanceId,
      });
    collaborationStateRef.current = createTimelineCollaborationState();
    pendingSavesRefsRef.current.pendingSignaturesRef.current.clear();
    publishPendingQueueState();
  }, [clientInstanceId, incidentId, publishPendingQueueState]);

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
    entityCatalogInput,
    gridHandleRef: timelineGridHandleRef,
    gridShellRef,
    rowInputRefs,
    rowInputTestIdsRef,
    setViewportContinuityRequest,
    viewportContinuityRequest,
    viewportContinuityTokenRef,
  });
  const {
    advanceViewportContinuity,
    beginViewportContinuity,
    clearViewportContinuity,
    resolveInputElement,
    resolveViewportContinuityElement,
    scrollToViewportContinuityTarget,
    settleViewportContinuityBarrier,
  } = timelineViewportContinuity.commands;
  const { advanceViewportContinuityRef, beginViewportContinuityRef } =
    timelineViewportContinuity.refs;

  const trackPendingSocketTxn = useCallback((clientTxnId: string) => {
    const existingTimeout =
      pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current.get(
        clientTxnId,
      );
    if (existingTimeout !== undefined) {
      window.clearTimeout(existingTimeout);
    }
    const timeoutId = window.setTimeout(() => {
      pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current.delete(
        clientTxnId,
      );
    }, 30_000);
    pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current.set(
      clientTxnId,
      timeoutId,
    );
  }, []);

  const resolvePendingSocketTxn = useCallback(
    (clientTxnId: string | null | undefined) => {
      if (!clientTxnId) {
        return false;
      }

      const timeoutId =
        pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current.get(
          clientTxnId,
        );
      if (timeoutId === undefined) {
        return false;
      }

      window.clearTimeout(timeoutId);
      pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current.delete(
        clientTxnId,
      );
      return true;
    },
    [],
  );
  const resolvePendingSocketTxnRef = useRef(resolvePendingSocketTxn);
  resolvePendingSocketTxnRef.current = resolvePendingSocketTxn;

  const {
    currentTimelineAnchorFor,
    navigateTimelineFocusAnchor,
    resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor,
  } = useTimelineGridAnchorController({
    gridHandleRef: timelineGridHandleRef,
    groupBy: queryState.groupBy,
    rowsRef,
    timelineAnchorColumnsRef,
    timelineAnchorRowsRef,
    updateTimelineSurfaceFocusAnchor,
    updateWorkbookFocusAnchor,
  });

  const applyRowMutation = useCallback(
    (
      rowKey: string,
      envelope: TimelineMutationEnvelope,
      options: {
        continueOnFreshDraft?: boolean;
        clearActiveCollectionFocusKey?: string | undefined;
        detectAutoResolution?: boolean;
        promoteToCommittedRowInspect?: boolean;
        viewportContinuityToken?: number;
      } = {},
    ) => {
      let previousRow = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      validateTimelineViewSchemaId(
        envelope.data.view_schema_id,
        "mutation response",
      );
      const responseRow = normalizeTimelineFullRow(
        envelope.data.row,
        "mutation response row",
      );
      recordWorkbookTiming("apply_row_mutation_start", {
        rowKey,
        recordId: responseRow.record_id,
        rowVersion: responseRow.row_version,
      });
      const incomingCommitted = rowFromApi(responseRow);
      const accepted = acceptCommittedTimelineRow(incomingCommitted);
      const committed = accepted.row;
      let draftSummaryKey: string | null = null;
      flushSync(() => {
        setRows((current) => {
          previousRow =
            current.find((candidate) => candidate.key === rowKey) ??
            current.find(
              (candidate) => candidate.recordId === committed.recordId,
            ) ??
            previousRow;
          let replaced = false;
          let nextRows = current.map((row) => {
            if (
              row.key !== rowKey &&
              (committed.recordId === null ||
                row.recordId !== committed.recordId)
            ) {
              return row;
            }
            replaced = true;
            return committed;
          });
          if (!replaced) {
            const draftIndex = nextRows.findIndex(
              (row) => row.recordId === null,
            );
            nextRows =
              draftIndex === -1
                ? [...nextRows, committed]
                : [
                    ...nextRows.slice(0, draftIndex),
                    committed,
                    ...nextRows.slice(draftIndex),
                  ];
          }
          const hydrated = ensureDraftRowWithFreshIndex(
            nextRows,
            nextDraftIndex,
          );
          draftSummaryKey = hydrated.draftSummaryKey;
          rowsRef.current = hydrated.rows;
          return hydrated.rows;
        });
        if (options.clearActiveCollectionFocusKey !== undefined) {
          setActiveCollectionInputKey((current) =>
            current === options.clearActiveCollectionFocusKey ? null : current,
          );
        }
      });

      if (committed.recordId !== null) {
        setDismissedMentionsByRow((current) =>
          pruneDismissedMentions(current, committed),
        );
        pruneAutoResolutionNoticesForRows([committed]);
      }
      if (options.detectAutoResolution !== false) {
        const notices = buildAutoResolutionNotices(previousRow, committed);
        if (notices.length > 0) {
          setAutoResolutionNotices((current) => {
            const knownRefs = new Set(current.map((notice) => notice.itemRef));
            return [
              ...current,
              ...notices.filter((notice) => !knownRefs.has(notice.itemRef)),
            ];
          });
        }
      }
      if (
        selectedRowId !== null &&
        previousRow?.recordId !== null &&
        previousRow !== undefined &&
        previousRow.recordId === selectedRowId
      ) {
        setSelectedRowId(committed.recordId);
      }
      const nextViewportTarget =
        options.continueOnFreshDraft && draftSummaryKey
          ? ({
              kind: "input",
              focusKey: draftSummaryKey,
            } satisfies ViewportContinuityTarget)
          : options.promoteToCommittedRowInspect && committed.recordId !== null
            ? ({
                kind: "row-inspect",
                recordId: committed.recordId,
              } satisfies ViewportContinuityTarget)
            : null;
      advanceViewportContinuity(options.viewportContinuityToken, {
        target: nextViewportTarget,
      });
      if (nextViewportTarget?.kind === "input") {
        resolveInputElement(nextViewportTarget.focusKey)?.focus({
          preventScroll: true,
        });
      }
      recordWorkbookTiming("apply_row_mutation_end", {
        rowKey,
        recordId: committed.recordId,
        rowVersion: committed.rowVersion,
      });
      return committed;
    },
    [
      acceptCommittedTimelineRow,
      advanceViewportContinuity,
      nextDraftIndex,
      pruneAutoResolutionNoticesForRows,
      resolveInputElement,
      selectedRowId,
      setDismissedMentionsByRow,
      setAutoResolutionNotices,
      setRows,
      setSelectedRowId,
    ],
  );

  const applyTimelineClipboardResponseRows = useCallback(
    (responseRows: readonly unknown[]) => {
      for (const [index, responseRow] of responseRows.entries()) {
        const row = normalizeTimelineFullRow(
          responseRow,
          `clipboard paste response rows[${index}]`,
        );
        applyRowMutation(row.record_id, {
          data: {
            row,
            view_schema_id: timelineViewSchemaId,
          },
        });
      }
    },
    [applyRowMutation],
  );

  const waitForCommittedRecordIdle = useCallback(
    async (
      recordId: string,
      options: {
        readonly fallbackRowVersion?: number | null | undefined;
        readonly refreshIfMissing?: boolean;
      } = {},
    ): Promise<{ row: WorkbookRow | null; rowVersion: number } | null> => {
      let attemptedRefresh = false;
      for (;;) {
        const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
        const snapshot = pending.model.snapshot();
        const hasPendingRecordWork = snapshot.units.some(
          (unit) => unit.recordId === recordId,
        );
        if (
          snapshot.authPaused ||
          snapshot.halted !== null ||
          snapshot.overflow !== null ||
          snapshot.sameFieldConflicts.length > 0 ||
          Object.keys(conflictQueueRef.current).length > 0
        ) {
          return null;
        }
        if (
          !hasPendingRecordWork &&
          !refreshBlocksTimelinePendingRecord(pending, recordId)
        ) {
          const row = latestCommittedTimelineRow(recordId);
          const rowVersion =
            latestCommittedRowVersion(recordId) ?? options.fallbackRowVersion;
          if (typeof rowVersion === "number") {
            return { row, rowVersion };
          }
          if (
            options.refreshIfMissing !== false &&
            attemptedRefresh === false
          ) {
            attemptedRefresh = true;
            await loadRowsRef.current({ showLoading: false });
            continue;
          }
          return null;
        }
        await new Promise((resolve) => window.setTimeout(resolve, 16));
      }
    },
    [latestCommittedRowVersion, latestCommittedTimelineRow],
  );

  const { loadRows } = useTimelineRowsLoader({
    acceptCommittedTimelineRows,
    advanceViewportContinuity,
    apiBase,
    beginRefreshInFlight,
    beginTimelineRowsLoad,
    clearViewportContinuity,
    committedRowsChangedSince,
    currentCommittedTimelineRow,
    finishRefreshInFlight,
    hasLoadedRows,
    incidentId,
    isCurrentLoadSequence,
    knownTimelineRowVersion,
    loadRowsRef,
    markRowsLoaded,
    nextDraftIndex,
    onIncidentAccessLost,
    pendingSavesRefsRef,
    pruneAutoResolutionNoticesForRows,
    pruneDismissedMentionsForRow: pruneDismissedMentions,
    publishSaveStatePresentation,
    queryState,
    rowsRef,
    scalarDraftValuesRef,
    setDismissedMentionsByRow,
    setIsInitialLoading,
    setIsRefreshing,
    setLoadError,
    setRefreshError,
    setRows,
    timelineContract,
  });

  const {
    beginWorkflow: beginCreateRelatedWorkflow,
    cancelWorkflow: cancelCreateRelatedWorkflow,
    submitWorkflow: submitCreateRelatedWorkflow,
    updateWorkflowDraft: updateCreateRelatedWorkflowDraft,
    workflow: createRelatedWorkflow,
  } = useTimelineCreateRelatedWorkflow({
    apiBase,
    applyRowMutation,
    currentUserId,
    incidentId,
    loadRows: loadRowsRef.current,
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
      apiBase,
      enabled: createRelatedNeedsIncidentMembers,
      incidentId,
    });
  const timelineCreateRelatedReferenceOptions = useMemo(
    () => ({
      ...emptyGenericReferenceOptions(),
      incidentMembers: timelineIncidentMemberOptions,
    }),
    [timelineIncidentMemberOptions],
  );

  const applyRecordChangedPatch = useCallback(
    (payload: RecordChangedPayload) => {
      if (isStaleTimelineRowVersion(payload.record_id, payload.row_version)) {
        return true;
      }
      acceptTimelineRecordVersion(payload.record_id, payload.row_version);
      const affectedView = payload.affected_views.find(
        (view) => view.view_schema_id === timelineViewSchemaId,
      );
      if (
        affectedView?.change_kind !== "patch" ||
        affectedView.patch_cells === undefined
      ) {
        return false;
      }

      let patch: TimelinePatchCells;
      try {
        patch = normalizeTimelinePatchCells(
          affectedView.patch_cells,
          "record_changed patch_cells",
        );
      } catch {
        return false;
      }
      if (patch.record_id !== payload.record_id) {
        return false;
      }
      if (isStaleTimelineRowVersion(patch.record_id, patch.row_version)) {
        return true;
      }
      acceptTimelineRecordVersion(patch.record_id, patch.row_version);
      let patched = false;
      const nextRows = rowsRef.current.map((row) => {
        if (row.recordId !== patch.record_id || row.rawRow === null) {
          return row;
        }
        patched = true;
        const nextRawRow = applyViewRowPatch(row.rawRow, patch);
        const committed = rowFromApi(nextRawRow);
        const accepted = acceptCommittedTimelineRow(committed);
        return {
          ...accepted.row,
          collectionDrafts: row.collectionDrafts,
          pendingSignature: row.pendingSignature,
        };
      });
      if (!patched) {
        return false;
      }
      rowsRef.current = nextRows;
      setRows(nextRows);
      return true;
    },
    [
      acceptCommittedTimelineRow,
      acceptTimelineRecordVersion,
      isStaleTimelineRowVersion,
      setRows,
    ],
  );
  const applyRecordChangedPatchRef = useRef(applyRecordChangedPatch);
  applyRecordChangedPatchRef.current = applyRecordChangedPatch;

  useEffect(() => {
    // Keep saved-view reselection and shell-level refreshes observable here even
    // though loadRows owns the actual query inputs.
    void reloadToken;
    void loadRows({ showLoading: true });
  }, [loadRows, reloadToken]);

  useEffect(() => {
    return () => {
      for (const timeoutId of pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current.values()) {
        window.clearTimeout(timeoutId);
      }
      pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current.clear();
      if (pendingSavesRefsRef.current.pendingReplayTimerRef.current !== null) {
        window.clearTimeout(
          pendingSavesRefsRef.current.pendingReplayTimerRef.current,
        );
        pendingSavesRefsRef.current.pendingReplayTimerRef.current = null;
      }
      if (
        pendingSavesRefsRef.current.pendingReplayAuthRetryRef.current !== null
      ) {
        window.clearTimeout(
          pendingSavesRefsRef.current.pendingReplayAuthRetryRef.current,
        );
        pendingSavesRefsRef.current.pendingReplayAuthRetryRef.current = null;
      }
    };
  }, []);

  useTimelineInspectorLifecycle({
    cancelCreateRelatedWorkflow,
    cancelRowHistoryRequests,
    clearRowHistory,
    gridShellRef,
    inspectorMentions,
    inspectorResetKey,
    restoreTimelineFocusAnchor,
    rowHistory,
    rows,
    selectedMentionRef,
    selectedRowId,
    setInspectorMessage,
    setIsInspectorOpen,
    setRowHistory,
    setRowHistoryPendingAction,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
    setSelectedRowId,
    workbookFocusAnchorRef,
  });

  const nextClientTxnId = useCallback(() => {
    const value = clientTxnRef.current;
    clientTxnRef.current += 1;
    return `timeline-client-${value}`;
  }, []);

  const assignTagToSelectedRows = useCallback(async () => {
    const tagName = bulkTagName.trim();
    if (!canBulkTag || tagName === "" || selectedTimelineRecordIds.size === 0) {
      return;
    }
    const selectedRows = rowsRef.current.filter(
      (row) =>
        row.recordId !== null &&
        selectedTimelineRecordIds.has(row.recordId) &&
        row.rowVersion !== null &&
        row.pendingSignature === null,
    );
    if (selectedRows.length !== selectedTimelineRecordIds.size) {
      setBulkTagMessage({
        kind: "error",
        message:
          "Selection changed before the command could be submitted. Review the selected rows and try again.",
      });
      return;
    }
    setBulkTagSubmitting(true);
    setBulkTagMessage(null);
    const result = await fetchWorkbookJSON<Record<string, unknown>>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
      ),
      {
        method: "POST",
        body: JSON.stringify({
          view_schema_id: timelineViewSchemaId,
          client_txn_id: nextClientTxnId(),
          kind: "multi_row_tag_assignment_v1",
          tag_name: tagName,
          targets: selectedRows.map((row) => ({
            record_id: row.recordId,
            base_row_version: row.rowVersion,
          })),
        }),
      },
    );
    if (!result.ok) {
      setBulkTagSubmitting(false);
      setBulkTagMessage({
        kind: "error",
        message: parseErrorMessage(result.payload),
      });
      return;
    }
    await loadRows({ showLoading: false });
    setBulkTagSubmitting(false);
    setBulkTagMessage({
      kind: "success",
      message: `Assigned tag to ${selectedRows.length} selected record${selectedRows.length === 1 ? "" : "s"}.`,
    });
  }, [
    apiBase,
    bulkTagName,
    canBulkTag,
    incidentId,
    loadRows,
    nextClientTxnId,
    selectedTimelineRecordIds,
  ]);

  const schedulePendingReplayRuntimeRef = useRef<() => void>(() => undefined);
  const timelineConflictResolverCoordinator =
    useTimelineConflictResolverCoordinator({
      activeConflictKey,
      activateGridEditor: (recordId, fieldKey, draftValue) =>
        timelineGridHandleRef.current?.activateEdit(
          {
            fieldKey,
            rowIdentity: { kind: "core_record", recordId },
            surface: {
              kind: "view_schema",
              viewSchemaId: timelineViewSchemaId,
            },
          },
          { value: draftValue },
        ) ?? false,
      apiBase,
      applyRowMutation,
      beginViewportContinuity,
      cancelGridEditor: (recordId, fieldKey) =>
        timelineGridHandleRef.current?.cancelEdit({
          fieldKey,
          rowIdentity: { kind: "core_record", recordId },
          surface: {
            kind: "view_schema",
            viewSchemaId: timelineViewSchemaId,
          },
        }) ?? false,
      conflictQueue,
      nextClientTxnId,
      pasteConflictGroup,
      pendingSavesRefsRef,
      publishSaveStatePresentation,
      resolveInputElement,
      rowsRef,
      scalarDraftValuesRef,
      schedulePendingReplayRef: schedulePendingReplayRuntimeRef,
      setActiveConflictKey,
      setConflictQueueState,
      setPasteConflictGroup,
      setRows,
      setSaveState,
      setSaveStateSecondaryMessage,
    });
  const {
    activeConflict,
    activePasteConflictIndex,
    activePasteConflictKeys,
    showPasteConflictNavigator,
  } = timelineConflictResolverCoordinator.snapshot;
  const {
    closeConflictResolver,
    handleConflictMergedDraftChange,
    handleMutationConflict,
    registerSameFieldConflict,
    submitConflictResolution,
  } = timelineConflictResolverCoordinator.commands;

  const setScalarEditorDraftValue = useCallback(
    (
      rowKey: string,
      field: keyof RowValues,
      surface: TimelineScalarEditorSurface,
      value: string,
    ) => {
      scalarDraftValuesRef.current.set(
        inputFocusKey(rowKey, field, surface),
        value,
      );
    },
    [],
  );

  const rowWithScalarEditorDrafts = useCallback(
    (
      row: WorkbookRow,
      preferred?: {
        readonly field: keyof RowValues;
        readonly value: string | undefined;
      },
    ): WorkbookRow => {
      return rowWithMaterializedScalarDrafts(
        row,
        (focusKey) => scalarDraftValuesRef.current.get(focusKey),
        preferred,
      );
    },
    [],
  );

  const clearSubmittedScalarEditorDraftValuesForRow = useCallback(
    (rowKey: string, submittedValues: RowValues) => {
      for (const binding of timelineScalarBindings) {
        for (const surface of timelineScalarEditorSurfaces) {
          const focusKey = inputFocusKey(rowKey, binding.key, surface);
          if (
            scalarDraftValuesRef.current.get(focusKey) ===
            submittedValues[binding.key]
          ) {
            scalarDraftValuesRef.current.delete(focusKey);
          }
        }
      }
    },
    [],
  );

  const registerInput = useCallback(
    (
      rowKey: string,
      field: FocusFieldKey,
      surface: TimelineScalarEditorSurface,
      dataTestId: string,
      element: HTMLInputElement | HTMLTextAreaElement | null,
    ) => {
      const key = inputFocusKey(rowKey, field, surface);
      if (element === null) {
        rowInputRefs.current.delete(key);
        rowInputTestIdsRef.current.delete(key);
        return;
      }
      rowInputTestIdsRef.current.set(key, dataTestId);
      rowInputRefs.current.set(key, element);
    },
    [],
  );

  const {
    enqueuePendingReplayUnit,
    scheduleAuthRecoveryProbeRef,
    schedulePendingReplay,
  } = useTimelinePendingReplayController({
    apiBase,
    applyRowMutation,
    clearSubmittedScalarEditorDraftValuesForRow,
    clearViewportContinuity,
    conflictQueueRef,
    handleMutationConflict,
    latestCommittedTimelineRow,
    pendingSavesRefsRef,
    publishPendingQueueState,
    recordWorkbookTiming,
    resolvePendingSocketTxn,
    rowsRef,
    scheduleSocketReconnectAfterAuthRef: socketReconnectAfterAuthRef,
    setRefreshError,
    setRows,
    trackPendingSocketTxn,
  });
  schedulePendingReplayRuntimeRef.current = schedulePendingReplay;

  const { sendPresenceUpdate } = useTimelineLiveUpdateController({
    activeSheetRuntimeRef,
    advanceViewportContinuityRef,
    apiBase,
    applyRecordChangedPatchRef,
    beginRefreshInFlightRef,
    beginViewportContinuityRef,
    currentPresence,
    finishRefreshInFlightRef,
    incidentId,
    liveUpdateRefs: timelineLiveUpdateRefs,
    loadRowsRef,
    pendingSavesRefsRef,
    publishPendingQueueStateRef,
    resolvePendingSocketTxnRef,
    scheduleAuthRecoveryProbeRef,
    setPresenceRecords,
    setRefreshError,
  });

  const {
    activeSheetPresenceRecords,
    editingPresenceForCell,
    handleEditModePresence,
    presenceForRow,
  } = useTimelinePresenceProjection({
    activeSheetRef,
    connectionId,
    currentPresenceRef,
    presenceRecords,
    sendPresenceUpdate,
    setCurrentPresence,
  });

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

  const {
    commitScalarGridEdit,
    queueAction,
    queueCollectionSave,
    queueScalarSave,
  } = useTimelineMutationCommands({
    acceptTimelineActionResult,
    apiBase,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    clientInstanceId,
    conflictQueueRef,
    enqueuePendingReplayUnit,
    finishSave,
    incidentId,
    latestCommittedTimelineRow,
    loadRowsRef,
    nextClientTxnId,
    pendingSavesRefsRef,
    replacementDrafts,
    resolvePendingSocketTxn,
    rowWithScalarEditorDrafts,
    rowsRef,
    scalarDraftValuesRef,
    setRows,
    trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  });

  const enqueueTimelineSaveWork = useCallback((work: () => Promise<void>) => {
    pendingSavesRefsRef.current.saveQueueRef.current =
      pendingSavesRefsRef.current.saveQueueRef.current
        .catch(() => undefined)
        .then(work);
  }, []);

  const {
    confirmRowHistoryPendingAction,
    openRowHistory,
    previewRowHistoryDeleteRestore,
    previewRowHistoryRollback,
  } = useTimelineHistoryActions({
    acceptTimelineRecordVersion,
    activeHistoryLiveRecordId,
    apiBase,
    beginRowHistoryRequest,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    currentHistoryRecordId,
    currentHistoryRecordIdMatches,
    currentHistoryRowVersion,
    enqueueSaveWork: enqueueTimelineSaveWork,
    finishSave,
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
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork: enqueueTimelineSaveWork,
      entityCatalogInput,
      finishSave,
      incidentId,
      loadRows: loadRowsRef.current,
      nextClientTxnId,
      onRefreshEntities,
      resolvePendingSocketTxn,
      resolveViewportContinuityElement,
      scrollToViewportContinuityTarget,
      rowsRef,
      setDismissedMentionsByRow,
      setInspectorMessage,
      settleViewportContinuityBarrier,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    });

  const { handleTimelineEvidenceFiles } = useTimelineEvidenceAttach({
    apiBase,
    applyRowMutation,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    enqueueSaveWork: enqueueTimelineSaveWork,
    finishSave,
    incidentId,
    nextClientTxnId,
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
        priorGridAnchor?.surface === timelineViewSchemaId
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
    apiBase,
    applyResponseRows: applyTimelineClipboardResponseRows,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    finishSave,
    incidentId,
    loadRowsRef,
    nextClientTxnId,
    pendingSavesRefsRef,
    queueScalarSave,
    registerSameFieldConflict,
    resolvePendingSocketTxn,
    resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor,
    rowInputRefs,
    setActiveConflictKey,
    setPasteConflictGroup,
    setScalarEditorDraftValue,
    trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  }).commands;

  const handleTimelineGridPaste = useCallback(
    (intent: Parameters<typeof handleGridPaste>[0]) => {
      const values = parseClipboardTable(intent.clipboardText ?? "");
      if (values.length === 1 && values[0]?.length === 1) {
        const binding = timelineScalarBindingForField(intent.target.fieldKey);
        if (binding === null) return;
        void commitScalarGridEdit(
          gridCoreRecordId(intent.target) ?? "",
          binding.key,
          values[0]?.[0] ?? "",
        ).then((outcome) => {
          if (outcome.kind !== "accepted") setRefreshError(outcome.message);
        });
        return;
      }
      handleGridPaste(intent);
    },
    [commitScalarGridEdit, handleGridPaste, setRefreshError],
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
      const clientTxnId = nextClientTxnId();
      const viewportContinuityToken = beginViewportContinuity({
        kind: "scroll-only",
      });
      beginSave();
      pendingSavesRefsRef.current.saveQueueRef.current =
        pendingSavesRefsRef.current.saveQueueRef.current
          .catch(() => undefined)
          .then(async () => {
            trackPendingSocketTxn(clientTxnId);
            const result = await fetchWorkbookJSON<unknown>(
              apiPath(
                apiBase,
                `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
              ),
              {
                method: "POST",
                body: JSON.stringify({
                  view_schema_id: timelineViewSchemaId,
                  client_txn_id: clientTxnId,
                  kind: "fill_down_v1",
                  field_key: intent.source.fieldKey,
                  value,
                  targets: targets.map((target) => ({
                    record_id: gridCoreRecordId(target) ?? "",
                    base_row_version: target.mutationIdentity.baseRowVersion,
                  })),
                }),
              },
            );
            resolvePendingSocketTxn(clientTxnId);
            if (!result.ok) {
              clearViewportContinuity(viewportContinuityToken);
              setRefreshError(parseErrorMessage(result.payload));
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
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      finishSave,
      incidentId,
      nextClientTxnId,
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
    timelineBindingLabel,
    timelineColumns,
  } = useTimelineWorkbookRenderers({
    activeCollectionInputKey,
    commitScalarGridEdit,
    conflictQueue,
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
    registerInput,
    rowGutterWidth: timelineRowGutterWidth,
    scalarDraftValuesRef,
    setActiveCollectionInputKey,
    setActiveConflictKey,
    setScalarEditorDraftValue,
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
          <TimelineRowGutterContent
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
  const timelineBulkSelection = useMemo(
    () => ({
      isRecordSelectable: (row: GridDataRow<WorkbookRow>) =>
        canBulkTag && row.data.pendingSignature === null,
      onSelectedRecordIdsChange: (recordIds: ReadonlySet<string>) => {
        setSelectedTimelineRecordIds(new Set(recordIds));
        setBulkTagMessage(null);
      },
      selectedRecordIds: selectedTimelineRecordIds,
    }),
    [canBulkTag, selectedTimelineRecordIds],
  );

  useLayoutEffect(() => {
    timelineAnchorColumnsRef.current = visibleTimelineColumns;
    timelineAnchorRowsRef.current = timelineGridRows;
  }, [timelineGridRows, visibleTimelineColumns]);

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
  const visibleRefreshError =
    refreshError !== null && refreshError !== pendingQueueDisplayMessage
      ? refreshError
      : null;
  const timelineLoadState: WorkbookQueryLoadState = isInitialLoading
    ? { kind: "initial_loading" }
    : loadError !== null
      ? workbookLoadFailureIsAccessLoss(loadError)
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
    <WorkbookSurfaceFrame
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
      primaryGrid={
        <TimelineGridSurface
          activeRecordId={selectedRowId}
          bulkSelection={timelineBulkSelection}
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
          onPasteCell={handleTimelineGridPaste}
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
          saveState={saveState}
          saveStateSecondaryMessage={saveStateSecondaryMessage}
          workbookFocusAnchor={workbookFocusAnchor}
        />
      }
      testId={timelineMutationSubstrateReadyTestId()}
      viewBar={
        <WorkbookSheetToolbar
          addRowDisabled={interactionMode.kind === "read_only"}
          leading={
            <>
              {savedViewSelector}
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
                    setBulkTagMessage(null);
                  }}
                />
                <button
                  disabled={
                    !canBulkTag ||
                    bulkTagSubmitting ||
                    selectedTimelineRecordIds.size === 0 ||
                    bulkTagName.trim() === ""
                  }
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
              {renderInlineQueryControls ? (
                <WorkbookGridControls
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
              ) : null}
            </>
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
                setReplacementDrafts((current) => ({
                  ...current,
                  [rowKey]: value,
                }));
              }}
              onSupersede={(rowKey) => {
                queueAction(rowKey, "supersede");
              }}
            />
          )}
          {activeConflict ? (
            <TimelineConflictResolver
              activeConflict={activeConflict}
              activeConflictKey={activeConflictKey}
              activePasteConflictIndex={activePasteConflictIndex}
              activePasteConflictKeys={activePasteConflictKeys}
              conflictQueue={conflictQueue}
              getFieldLabel={timelineBindingLabel}
              onClose={closeConflictResolver}
              onMergedDraftChange={handleConflictMergedDraftChange}
              onSelectConflictKey={(conflictKey) => {
                setActiveConflictKey(conflictKey);
              }}
              onSubmit={submitConflictResolution}
              showPasteConflictNavigator={showPasteConflictNavigator}
            />
          ) : null}
        </>
      }
      onWorkAreaContextMenu={handleTimelineGridContextMenu}
      onWorkAreaKeyDown={handleTimelineGridContextKeyDown}
    />
  );
}

export function TimelineWorkbook(props: TimelineWorkbookProps) {
  return (
    <IncidentCollaborationBoundary
      apiBase={props.apiBase}
      incidentId={props.incidentId}
      initialPresence={{
        sheet_ref: props.sheetRef ?? {
          kind: "view_schema",
          id: timelineViewSchemaId,
        },
        mode: "viewing",
      }}
    >
      <TimelineWorkbookContent {...props} key={props.incidentId} />
    </IncidentCollaborationBoundary>
  );
}
