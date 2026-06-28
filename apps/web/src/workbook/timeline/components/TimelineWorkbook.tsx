import type {
  GridCellAnchor,
  GridColumn,
  GridDensity,
  GridRow,
} from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  draftCellTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridGroupRowTestId,
  gridRowGutterTestId,
  gridScrollportSelector,
  rowCellTestId,
  timelineInspectorSectionTestId,
  timelineMutationSubstrateReadyTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import {
  type Dispatch,
  type ClipboardEvent as ReactClipboardEvent,
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
import { apiPath } from "../../../services/browserApi";
import { fetchJSON, readEnvelope } from "../../../services/workbookApi";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import { WorkbookGridControls } from "../../components/WorkbookGridControls";
import { WorkbookSheetToolbar } from "../../components/WorkbookSheetToolbar";
import { WorkbookStatusStrip } from "../../components/WorkbookStatusStrip";
import { WorkbookSurfaceFrame } from "../../components/WorkbookSurfaceFrame";
import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import { selectInspectorConfig } from "../../models/workbookInspectorModel";
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
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "../../utils/workbookContinuity";
import type { WorkbookFocusAnchor } from "../../utils/workbookGridFocus";
import {
  mapWorkbookKeyboardCommand,
  type WorkbookKeyboardCommand,
} from "../../utils/workbookKeyboard";
import {
  deriveWorkbookSaveState,
  sameFieldConflictQueueKey,
  type WorkbookSaveStateConflictAnchor,
} from "../../utils/workbookPendingQueue";
import { presenceMatchesSheet } from "../../utils/workbookPresence";
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
import {
  type TimelineLiveUpdateRefs,
  type TimelinePresenceDraft,
  useTimelineLiveUpdates,
} from "../hooks/useTimelineLiveUpdates";
import { useTimelineMentionActions } from "../hooks/useTimelineMentionActions";
import { useTimelineMentions } from "../hooks/useTimelineMentions";
import { useTimelineMutationCommands } from "../hooks/useTimelineMutationCommands";
import {
  type PendingReplayRuntimeMeta,
  useTimelinePendingReplayController,
} from "../hooks/useTimelinePendingReplayController";
import {
  beginTimelinePendingRefreshBlock,
  createTimelinePendingQueueRuntime,
  ensureTimelineTabClientInstanceId,
  finishTimelinePendingRefreshBlock,
  refreshBlocksTimelinePendingRecord,
  type TimelinePendingQueueRuntime,
  type TimelinePendingRefreshBlockScope,
  timelinePendingQueueSnapshot,
  useTimelinePendingSaves,
} from "../hooks/useTimelinePendingSaves";
import { useTimelineRows } from "../hooks/useTimelineRows";
import {
  type LoadRowsOptions,
  useTimelineRowsLoader,
} from "../hooks/useTimelineRowsLoader";
import { useTimelineWorkbookRuntime } from "../hooks/useTimelineWorkbookRuntime";
import { buildTimelineGridRows } from "../models/timelineRowsModel";
import {
  settleTimelineViewportContinuityBarrier,
  type TimelineEntityCatalogInput,
  type TimelineEntityRefreshSettleState,
  type TimelineViewportContinuityBarrier,
  timelineViewportContinuityBarrierSatisfied,
} from "../models/timelineViewportContinuityModel";
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
  type SameFieldConflictPayload,
  type TimelinePatchCells,
  type TimelineScalarEditorSurface,
  timelineCollectionBindings,
  timelineGroupLabel,
  timelineInspectorBindings,
  timelineRelationshipLabel,
  timelineScalarBindingForField,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import type { TimelineMutationEnvelope } from "../services/timelineMutationRequests";
import type { RecordChangedPayload } from "../services/workbookCollaborationMessages";
import {
  createWorkbookSocketLifecycleState,
  type WorkbookSocketLifecycleAction,
  type WorkbookSocketLifecycleEffect,
} from "../services/workbookSocketLifecycle";
import {
  DraftRowCreateButton,
  mentionChipStateForItem,
} from "./TimelineCellEditors";
import { TimelineConflictResolver } from "./TimelineConflictResolver";
import { TimelineEvidencePanel } from "./TimelineEvidencePanel";
import { TimelineGridSurface } from "./TimelineGridSurface";
import { TimelineHistoryPanel } from "./TimelineHistoryPanel";
import { TimelineRowGutterContent } from "./TimelinePresenceMarkers";
import { TimelineRowContextMenu } from "./TimelineRowActions";
import { TimelineWorkbookInspector } from "./TimelineWorkbookInspector";
import {
  TimelineWorkbookNotices,
  timelinePendingQueueMessage,
} from "./TimelineWorkbookNotices";
import { useTimelineWorkbookRenderers } from "./TimelineWorkbookRenderers";
import {
  actionButtonStyle,
  bodyStyle,
  eyebrowStyle,
  headlineStyle,
  inlineButtonRowStyle,
  inspectorActionStackStyle,
  inspectorSectionStyle,
  labelStyle,
  panelStyle,
  secondaryActionButtonStyle,
  sectionTitleStyle,
  timelineGridShellStyle,
  timelineNoticeOverlayStyle,
  timelineRowGutterWidth,
} from "./TimelineWorkbookStyles";

const timelineContract = requireViewContract(timelineViewSchemaId);
const timelineInspectorConfig = selectInspectorConfig(timelineContract);
const createRelatedTargetContracts = new Map<string, ViewContract>(
  [
    notesViewSchemaId,
    taskRequestsViewSchemaId,
    decisionsViewSchemaId,
    evidenceViewSchemaId,
    commLogViewSchemaId,
    handoffViewSchemaId,
    statusReviewViewSchemaId,
    lessonViewSchemaId,
  ].map((viewSchemaId) => [viewSchemaId, requireViewContract(viewSchemaId)]),
);
const timelineCreateRelatedReferenceOptions = emptyGenericReferenceOptions();

export type SaveState = "Syncing" | "Saved" | "Conflict";
type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;
export type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";

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
  onRefreshEntities?: () => Promise<void> | void;
};

type TimelineClipboardPasteEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id?: string;
    rows: unknown[];
    conflicts?: SameFieldConflictPayload[];
  };
};

type PendingQueueRuntime =
  TimelinePendingQueueRuntime<PendingReplayRuntimeMeta>;

function saveStateConflictAnchorsFromLocalConflicts(
  conflicts: Record<string, LocalConflictState>,
): WorkbookSaveStateConflictAnchor[] {
  return Object.values(conflicts).map((entry) => ({ ...entry.anchor }));
}

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

type ViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

type ViewportContinuityRequest = {
  token: number;
  attemptVersion: number;
  target: ViewportContinuityTarget;
  preservedViewport: ViewportSnapshot | null;
  barrier: TimelineViewportContinuityBarrier;
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

function resolveGridScrollElement(
  element: HTMLElement,
  surface: WorkbookSurface,
): HTMLElement {
  const selector = gridScrollportSelector();
  const scrollports = Array.from(
    element.querySelectorAll<HTMLElement>(selector),
  );
  if (scrollports.length !== 1) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${selector} scrollport, received ${scrollports.length}`,
    );
  }
  const scrollport = scrollports[0];
  if (scrollport === undefined) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${selector} scrollport, received 0`,
    );
  }
  return scrollport;
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

export function TimelineWorkbook({
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
  onRefreshEntities,
}: TimelineWorkbookProps) {
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
    loadError,
    refreshError,
    saveState,
    saveStateSecondaryMessage,
    setIsInitialLoading,
    setLoadError,
    setRefreshError,
    setSaveState,
    setSaveStateSecondaryMessage,
  } = timelineRuntime.lifecycle;
  const {
    applyQueryFilter,
    filterDraft,
    handleQueryGroupByChange,
    handleQuerySortToggle,
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
  const conflictQueueRef = useRef<Record<string, LocalConflictState>>({});
  const timelineConflicts = useTimelineConflicts({ conflictQueueRef });
  const { activeConflictKey, conflictQueue, pasteConflictGroup } =
    timelineConflicts.snapshot;
  const { setActiveConflictKey, setConflictQueueState, setPasteConflictGroup } =
    timelineConflicts.commands;
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
  const activeSocketRef = useRef<WebSocket | null>(null);
  const socketLifecycleRef = useRef(createWorkbookSocketLifecycleState());
  const socketEstablishedRef = useRef(false);
  const socketConnectionIDRef = useRef<string | null>(null);
  const presenceUpdateTimerRef = useRef<number | null>(null);
  const currentPresenceRef = useRef<TimelinePresenceDraft>({
    fieldKey: null,
    mode: "viewing",
    recordId: null,
  });
  const socketReconnectAfterAuthRef = useRef<(() => void) | null>(null);
  const socketResumeTokenRef = useRef<string | null>(null);
  const socketLastSeenStreamSeqRef = useRef(0);
  const dispatchSocketLifecycleRef = useRef<
    (action: WorkbookSocketLifecycleAction) => WorkbookSocketLifecycleEffect[]
  >(() => []);
  const timelineLiveUpdateRefs: TimelineLiveUpdateRefs = {
    activeSocketRef,
    currentPresenceRef,
    dispatchSocketLifecycleRef,
    presenceUpdateTimerRef,
    socketConnectionIDRef,
    socketEstablishedRef,
    socketLastSeenStreamSeqRef,
    socketLifecycleRef,
    socketReconnectAfterAuthRef,
    socketResumeTokenRef,
  };
  const timelineLiveUpdates = useTimelineLiveUpdates({
    refs: timelineLiveUpdateRefs,
  });
  const { currentPresence, presenceRecords } = timelineLiveUpdates.snapshot;
  const {
    dispatchSocketLifecycle,
    setCurrentPresence,
    setPresenceRecords,
    syncSocketLifecycleRefs,
  } = timelineLiveUpdates.commands;
  const timelinePendingSaves =
    useTimelinePendingSaves<PendingReplayRuntimeMeta>({
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
  const timelineAnchorRowsRef = useRef<readonly GridRow<WorkbookRow>[]>([]);
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const [timelineGridShellWidth, setTimelineGridShellWidth] = useState(0);
  const viewportContinuityTokenRef = useRef(1);
  const timelineGridInteractionRefs: TimelineGridInteractionRefs = {
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

  const computeSaveStatePresentation = useCallback(
    (
      pending: PendingQueueRuntime,
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ) => {
      const snapshot = pending.model.snapshot();
      return deriveWorkbookSaveState({
        authPaused: snapshot.authPaused,
        halted: snapshot.halted,
        overflow: snapshot.overflow,
        sameFieldConflicts: snapshot.sameFieldConflicts,
        localDraftConflicts:
          saveStateConflictAnchorsFromLocalConflicts(conflicts),
        queuedCount: snapshot.queuedCount,
        inFlightCount: snapshot.inFlightCount,
        refreshPaused: pending.resetRefreshInFlight,
        pendingMutationCount: pendingSavesRefsRef.current.pendingOpsRef.current,
      });
    },
    [],
  );

  const publishSaveStatePresentation = useCallback(
    (
      pending: PendingQueueRuntime,
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ) => {
      const presentation = computeSaveStatePresentation(pending, conflicts);
      setSaveState(presentation.primaryLabel);
      setSaveStateSecondaryMessage(presentation.secondaryMessage);
      return presentation;
    },
    [computeSaveStatePresentation, setSaveState, setSaveStateSecondaryMessage],
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

  const publishPendingQueueState = useCallback(() => {
    const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
    setPendingQueueSnapshot(timelinePendingQueueSnapshot(pending));
    publishSaveStatePresentation(pending);
  }, [publishSaveStatePresentation, setPendingQueueSnapshot]);
  const publishPendingQueueStateRef = useRef(publishPendingQueueState);
  publishPendingQueueStateRef.current = publishPendingQueueState;

  const beginRefreshInFlight = useCallback(
    (scope: TimelinePendingRefreshBlockScope) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      beginTimelinePendingRefreshBlock(pending, scope);
      publishPendingQueueState();
    },
    [publishPendingQueueState],
  );

  const finishRefreshInFlight = useCallback(
    (scope: TimelinePendingRefreshBlockScope) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      finishTimelinePendingRefreshBlock(pending, scope);
      publishPendingQueueState();
      pendingSavesRefsRef.current.schedulePendingReplayRef.current();
    },
    [publishPendingQueueState],
  );
  const beginRefreshInFlightRef = useRef(beginRefreshInFlight);
  beginRefreshInFlightRef.current = beginRefreshInFlight;
  const finishRefreshInFlightRef = useRef(finishRefreshInFlight);
  finishRefreshInFlightRef.current = finishRefreshInFlight;

  useEffect(() => {
    const clientInstanceId = ensureTimelineTabClientInstanceId(
      pendingSavesRefsRef.current.socketClientInstanceIdRef,
    );
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
    socketLifecycleRef.current = createWorkbookSocketLifecycleState();
    syncSocketLifecycleRefs();
    pendingSavesRefsRef.current.pendingSignaturesRef.current.clear();
    publishPendingQueueState();
  }, [incidentId, publishPendingQueueState, syncSocketLifecycleRefs]);

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
  const activeSheetPresenceRecords = useMemo(
    () =>
      [...presenceRecords]
        .filter((presence) => presenceMatchesSheet(presence, activeSheetRef))
        .filter(
          (presence) =>
            presence.connection_id !== socketConnectionIDRef.current,
        )
        .sort((left, right) => {
          const byName = left.display_name.localeCompare(right.display_name);
          return byName === 0
            ? left.connection_id.localeCompare(right.connection_id)
            : byName;
        }),
    [activeSheetRef, presenceRecords],
  );

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

  const currentGridScrollSnapshot = useCallback(() => {
    const element = gridShellRef.current;
    if (!element) {
      return null;
    }
    const scrollElement = resolveGridScrollElement(element, "timeline");
    return {
      top: scrollElement.scrollTop,
      left: scrollElement.scrollLeft,
    };
  }, []);

  const currentGridViewportSnapshot = useCallback(
    (target: HTMLElement | null = null): ViewportSnapshot | null => {
      const gridShell = gridShellRef.current;
      const scroll = currentGridScrollSnapshot();
      if (gridShell === null || scroll === null) {
        return null;
      }
      return {
        scroll,
        anchor:
          target === null
            ? null
            : captureViewportAnchor(
                resolveGridScrollElement(
                  gridShell,
                  "timeline",
                ).getBoundingClientRect(),
                target.getBoundingClientRect(),
              ),
      };
    },
    [currentGridScrollSnapshot],
  );

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

  const restoreGridScroll = useCallback(
    (preservedScroll: ScrollPosition | null) => {
      const gridShell = gridShellRef.current;
      if (gridShell === null || preservedScroll === null) {
        return;
      }
      const scrollElement = resolveGridScrollElement(gridShell, "timeline");
      scrollElement.scrollTop = preservedScroll.top;
      scrollElement.scrollLeft = preservedScroll.left;
      window.requestAnimationFrame(() => {
        const currentGridShell = gridShellRef.current;
        if (currentGridShell === null) {
          return;
        }
        const currentScrollElement = resolveGridScrollElement(
          currentGridShell,
          "timeline",
        );
        currentScrollElement.scrollTop = preservedScroll.top;
        currentScrollElement.scrollLeft = preservedScroll.left;
      });
    },
    [],
  );

  const restoreGridViewportForElement = useCallback(
    (
      resolveElement: () => HTMLElement | null,
      preservedViewport: ViewportSnapshot | null,
    ) => {
      // Continuity restores the previous scroll position first, then applies
      // only the extra delta needed to keep the target fully visible.
      const currentViewport =
        preservedViewport ??
        ({
          scroll: currentGridScrollSnapshot(),
          anchor: null,
        } satisfies ViewportSnapshot);
      const preservedScroll = currentViewport.scroll;
      const focusResolvedElement = () => {
        const element = resolveElement();
        if (element === null || !element.isConnected) {
          return false;
        }
        if (!element.hasAttribute("tabindex")) {
          element.tabIndex = -1;
        }
        element.focus({ preventScroll: true });
        return document.activeElement === element;
      };
      window.focus();
      const focusedNow = focusResolvedElement();
      restoreGridScroll(preservedScroll);
      const restoreViewportGeometryNow = () => {
        const currentGridShell = gridShellRef.current;
        const currentElement = resolveElement();
        if (
          currentGridShell === null ||
          preservedScroll === null ||
          currentElement === null ||
          !currentElement.isConnected
        ) {
          return false;
        }
        const scrollElement = resolveGridScrollElement(
          currentGridShell,
          "timeline",
        );
        const restoredScroll = computeRestoredViewportScroll({
          preservedScroll,
          currentScroll: {
            top: scrollElement.scrollTop,
            left: scrollElement.scrollLeft,
          },
          preservedAnchor: currentViewport.anchor,
          containerRect: scrollElement.getBoundingClientRect(),
          elementRect: currentElement.getBoundingClientRect(),
        });
        restoreGridScroll(restoredScroll);
        const updatedGridShell = gridShellRef.current;
        const updatedElement = resolveElement();
        if (
          updatedGridShell === null ||
          updatedElement === null ||
          !updatedElement.isConnected
        ) {
          return false;
        }
        const fullyVisible = isRectFullyVisibleWithinContainer(
          resolveGridScrollElement(
            updatedGridShell,
            "timeline",
          ).getBoundingClientRect(),
          updatedElement.getBoundingClientRect(),
        );
        return focusResolvedElement() && fullyVisible;
      };
      const restoredNow = restoreViewportGeometryNow();
      const restoreViewportGeometry = (attempt: number) => {
        window.requestAnimationFrame(() => {
          if (restoreViewportGeometryNow()) {
            return;
          }
          if (attempt < 6) {
            restoreViewportGeometry(attempt + 1);
          }
        });
      };
      restoreViewportGeometry(0);
      return focusedNow && restoredNow;
    },
    [currentGridScrollSnapshot, restoreGridScroll],
  );

  const resolveInputElement = useCallback((focusKey: string) => {
    const selectorTestId = rowInputTestIdsRef.current.get(focusKey) ?? null;
    const selector =
      selectorTestId === null
        ? null
        : document.querySelector<HTMLInputElement | HTMLTextAreaElement>(
            dataTestIdSelector(selectorTestId),
          );
    if (selector !== null) {
      return selector;
    }
    const [rowKey, fieldKey, surface] = focusKey.split(":");
    const scalarBinding = timelineScalarBindings.find(
      (binding) => binding.key === fieldKey,
    );
    if (
      rowKey !== undefined &&
      surface === "grid" &&
      scalarBinding !== undefined
    ) {
      const fallbackTestId = rowKey.startsWith("draft-")
        ? draftCellTestId(scalarBinding.fieldKey)
        : rowCellTestId(rowKey, scalarBinding.fieldKey);
      const fallback = document.querySelector<
        HTMLInputElement | HTMLTextAreaElement
      >(dataTestIdSelector(fallbackTestId));
      if (fallback !== null) {
        return fallback;
      }
    }
    return rowInputRefs.current.get(focusKey) ?? null;
  }, []);

  const {
    currentTimelineAnchorFor,
    navigateTimelineFocusAnchor,
    resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor,
  } = useTimelineGridAnchorController({
    groupBy: queryState.groupBy,
    resolveInputElement,
    rowsRef,
    timelineAnchorColumnsRef,
    timelineAnchorRowsRef,
    updateTimelineSurfaceFocusAnchor,
    updateWorkbookFocusAnchor,
  });

  const resolveViewportContinuityElement = useCallback(
    (target: ViewportContinuityTarget) => {
      switch (target.kind) {
        case "row-inspect":
          return (
            document.querySelector<HTMLElement>(
              dataTestIdSelector(
                rowCellTestId(
                  target.recordId,
                  "timeline.activity_synopsis_text",
                ),
              ),
            ) ??
            document.querySelector<HTMLElement>(
              dataTestIdSelector(
                gridRowGutterTestId(timelineViewSchemaId, target.recordId),
              ),
            )
          );
        case "input":
          return resolveInputElement(target.focusKey);
        case "scroll-only":
          return null;
      }
    },
    [resolveInputElement],
  );

  const beginViewportContinuity = useCallback(
    (
      target: ViewportContinuityTarget,
      options: { barrier?: TimelineViewportContinuityBarrier } = {},
    ) => {
      const token = viewportContinuityTokenRef.current;
      viewportContinuityTokenRef.current += 1;
      setViewportContinuityRequest({
        token,
        attemptVersion: 0,
        target,
        preservedViewport: currentGridViewportSnapshot(
          resolveViewportContinuityElement(target),
        ),
        barrier: options.barrier ?? null,
      });
      return token;
    },
    [
      currentGridViewportSnapshot,
      resolveViewportContinuityElement,
      setViewportContinuityRequest,
    ],
  );
  const beginViewportContinuityRef = useRef(beginViewportContinuity);
  beginViewportContinuityRef.current = beginViewportContinuity;

  const settleViewportContinuityBarrier = useCallback(
    (token: number, refreshState: TimelineEntityRefreshSettleState) => {
      setViewportContinuityRequest((current) => {
        if (!current || current.token !== token) {
          return current;
        }
        return {
          ...current,
          barrier: settleTimelineViewportContinuityBarrier(
            current.barrier,
            refreshState,
          ),
          attemptVersion: current.attemptVersion + 1,
        };
      });
    },
    [setViewportContinuityRequest],
  );

  const clearViewportContinuity = useCallback(
    (token: number) => {
      setViewportContinuityRequest((current) =>
        current?.token === token ? null : current,
      );
    },
    [setViewportContinuityRequest],
  );

  const advanceViewportContinuity = useCallback(
    (
      token: number | undefined,
      options: {
        barrier?: TimelineViewportContinuityBarrier;
        target?: ViewportContinuityTarget | null;
      } = {},
    ) => {
      if (token === undefined) {
        return;
      }
      setViewportContinuityRequest((current) => {
        if (current === null || current.token !== token) {
          return current;
        }
        return {
          ...current,
          attemptVersion: current.attemptVersion + 1,
          barrier:
            options.barrier === undefined ? current.barrier : options.barrier,
          target: options.target ?? current.target,
        };
      });
    },
    [setViewportContinuityRequest],
  );
  const advanceViewportContinuityRef = useRef(advanceViewportContinuity);
  advanceViewportContinuityRef.current = advanceViewportContinuity;

  const tryRestoreViewportContinuity = useCallback(
    (continuity: ViewportContinuityRequest) => {
      if (continuity.target.kind === "scroll-only") {
        restoreGridScroll(continuity.preservedViewport?.scroll ?? null);
        return true;
      }
      return restoreGridViewportForElement(
        () => resolveViewportContinuityElement(continuity.target),
        continuity.preservedViewport,
      );
    },
    [
      resolveViewportContinuityElement,
      restoreGridScroll,
      restoreGridViewportForElement,
    ],
  );

  const shouldHoldViewportContinuity = useCallback(
    (continuity: ViewportContinuityRequest) => {
      return !timelineViewportContinuityBarrierSatisfied(
        continuity.barrier,
        entityCatalogInput,
      );
    },
    [entityCatalogInput],
  );

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
    pendingSavesRefsRef,
    pruneAutoResolutionNoticesForRows,
    pruneDismissedMentionsForRow: pruneDismissedMentions,
    publishSaveStatePresentation,
    queryState,
    rowsRef,
    scalarDraftValuesRef,
    setDismissedMentionsByRow,
    setIsInitialLoading,
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

  useLayoutEffect(() => {
    if (
      viewportContinuityRequest === null ||
      viewportContinuityRequest.attemptVersion < 1
    ) {
      return;
    }
    let cancelled = false;
    const restoreTarget = (attempt: number) => {
      if (cancelled) {
        return;
      }
      if (!tryRestoreViewportContinuity(viewportContinuityRequest)) {
        if (attempt < 60) {
          window.setTimeout(() => {
            restoreTarget(attempt + 1);
          }, 50);
        }
        return;
      }
      if (shouldHoldViewportContinuity(viewportContinuityRequest)) {
        return;
      }
      window.requestAnimationFrame(() => {
        if (cancelled) {
          return;
        }
        if (shouldHoldViewportContinuity(viewportContinuityRequest)) {
          return;
        }
        if (!tryRestoreViewportContinuity(viewportContinuityRequest)) {
          if (attempt < 60) {
            window.setTimeout(() => {
              restoreTarget(attempt + 1);
            }, 50);
          }
          return;
        }
        clearViewportContinuity(viewportContinuityRequest.token);
      });
    };
    restoreTarget(0);
    return () => {
      cancelled = true;
    };
  }, [
    clearViewportContinuity,
    shouldHoldViewportContinuity,
    tryRestoreViewportContinuity,
    viewportContinuityRequest,
  ]);

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

  useEffect(() => {
    const hasUnsavedRuntimeWork =
      pendingQueueSnapshot.queuedCount > 0 ||
      pendingQueueSnapshot.inFlightCount > 0 ||
      Object.keys(conflictQueue).length > 0;
    if (!hasUnsavedRuntimeWork) {
      return;
    }
    const warnBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warnBeforeUnload);
    return () => {
      window.removeEventListener("beforeunload", warnBeforeUnload);
    };
  }, [
    conflictQueue,
    pendingQueueSnapshot.inFlightCount,
    pendingQueueSnapshot.queuedCount,
  ]);

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

  const beginSave = useCallback(() => {
    pendingSavesRefsRef.current.pendingOpsRef.current += 1;
    publishSaveStatePresentation(
      pendingSavesRefsRef.current.pendingQueueRef.current,
    );
  }, [publishSaveStatePresentation]);

  const finishSave = useCallback(
    (nextState: SaveState) => {
      pendingSavesRefsRef.current.pendingOpsRef.current = Math.max(
        0,
        pendingSavesRefsRef.current.pendingOpsRef.current - 1,
      );
      if (nextState === "Conflict") {
        setSaveState("Conflict");
        setSaveStateSecondaryMessage("Conflict requires review.");
        return;
      }
      publishSaveStatePresentation(
        pendingSavesRefsRef.current.pendingQueueRef.current,
      );
    },
    [publishSaveStatePresentation, setSaveState, setSaveStateSecondaryMessage],
  );

  const schedulePendingReplayRuntimeRef = useRef<() => void>(() => undefined);
  const timelineConflictResolverCoordinator =
    useTimelineConflictResolverCoordinator({
      activeConflictKey,
      apiBase,
      applyRowMutation,
      beginViewportContinuity,
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
    dispatchSocketLifecycle,
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

  const { queueAction, queueCollectionSave, queueScalarSave } =
    useTimelineMutationCommands({
      acceptTimelineActionResult,
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
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

  const { submitMentionAction } = useTimelineMentionActions({
    advanceViewportContinuity,
    apiBase,
    applyRowMutation,
    beginSave,
    beginViewportContinuity,
    clearViewportContinuity,
    enqueueSaveWork: enqueueTimelineSaveWork,
    entityCatalogInput,
    finishSave,
    loadRows: loadRowsRef.current,
    nextClientTxnId,
    onRefreshEntities,
    resolvePendingSocketTxn,
    resolveViewportContinuityElement,
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
      if (command.kind === "open-history") {
        openRowHistory(anchor.recordId);
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
        setSelectedRowId(anchor.recordId);
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
            recordId: anchor.recordId,
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
        const row = rowsRef.current.find(
          (candidate) => candidate.recordId === anchor.recordId,
        );
        const mention =
          row === undefined
            ? undefined
            : [
                ...row.collectionValues.hostRefs,
                ...row.collectionValues.identityRefs,
              ].find((item) => item.itemKind !== "resolved_ref");
        if (mention !== undefined) {
          setSelectedRowId(anchor.recordId);
          setIsInspectorOpen(true);
          setSelectedMentionRef(mention.itemRef);
          setInspectorMessage(null);
        } else {
          setSelectedRowId(anchor.recordId);
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

  const handlePaste = useCallback(
    (
      event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
    ) => {
      const clipboardText = event.clipboardData?.getData("text/plain") ?? "";
      const binding = timelineScalarBindings.find(
        (candidate) => candidate.key === focusField,
      );
      const fieldKey = binding?.fieldKey ?? focusField;
      if (surface === "grid" && binding !== undefined) {
        const pasteTargetResolution = resolveTimelinePasteTargetResolution(
          rowKey,
          fieldKey,
          clipboardText,
        );
        if (pasteTargetResolution !== null) {
          event.preventDefault();
          const { anchor, targetResolution } = pasteTargetResolution;
          const clientTxnId = nextClientTxnId();
          const viewportContinuityToken = beginViewportContinuity(
            anchor === null
              ? { kind: "scroll-only" }
              : {
                  kind: "input",
                  focusKey: inputFocusKey(rowKey, focusField, surface),
                },
          );
          beginSave();
          pendingSavesRefsRef.current.saveQueueRef.current =
            pendingSavesRefsRef.current.saveQueueRef.current
              .catch(() => undefined)
              .then(async () => {
                const rowTargetPayload: Array<
                  | { readonly kind: "create" }
                  | {
                      readonly base_row_version: number;
                      readonly kind: "record";
                      readonly record_id: string;
                    }
                > = [];
                for (const target of targetResolution.rowTargets) {
                  if (target.kind === "create") {
                    rowTargetPayload.push({ kind: "create" });
                    continue;
                  }
                  const idleRecord = await waitForCommittedRecordIdle(
                    target.recordId,
                  );
                  if (idleRecord === null) {
                    clearViewportContinuity(viewportContinuityToken);
                    finishSave("Conflict");
                    return;
                  }
                  rowTargetPayload.push({
                    kind: "record",
                    record_id: target.recordId,
                    base_row_version: idleRecord.rowVersion,
                  });
                }
                trackPendingSocketTxn(clientTxnId);
                const result = await fetchJSON<TimelineClipboardPasteEnvelope>(
                  apiPath(
                    apiBase,
                    `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
                  ),
                  {
                    method: "POST",
                    body: JSON.stringify({
                      view_schema_id: timelineViewSchemaId,
                      client_txn_id: clientTxnId,
                      clipboard_text: clipboardText,
                      format: clipboardText.includes("\t") ? "tsv" : "csv",
                      start_field_key: fieldKey,
                      columns: targetResolution.columns,
                      targets: rowTargetPayload,
                    }),
                  },
                );
                resolvePendingSocketTxn(clientTxnId);
                if (!result.ok) {
                  clearViewportContinuity(viewportContinuityToken);
                  finishSave("Conflict");
                  return;
                }
                const envelope = readEnvelope<TimelineClipboardPasteEnvelope>(
                  result.payload,
                );
                const pasteConflictKeys: string[] = [];
                for (const conflict of envelope.data.conflicts ?? []) {
                  const conflictBinding = timelineScalarBindingForField(
                    conflict.field_key,
                  );
                  const queueKey = sameFieldConflictQueueKey(conflict);
                  pasteConflictKeys.push(queueKey);
                  registerSameFieldConflict(
                    conflict,
                    inputFocusKey(
                      conflict.record_id,
                      conflictBinding?.key ?? focusField,
                      "grid",
                    ),
                    "grid",
                  );
                }
                if (pasteConflictKeys.length > 1) {
                  setPasteConflictGroup({ keys: pasteConflictKeys });
                  setActiveConflictKey(pasteConflictKeys[0] ?? null);
                } else if (pasteConflictKeys.length === 0) {
                  setPasteConflictGroup(null);
                }
                await loadRowsRef.current({
                  showLoading: false,
                  viewportContinuityToken,
                });
                if (anchor !== null) {
                  restoreTimelineFocusAnchor(anchor);
                }
                finishSave(
                  envelope.data.conflicts && envelope.data.conflicts.length > 0
                    ? "Conflict"
                    : "Saved",
                );
              });
          return;
        }
      }
      window.setTimeout(() => {
        const editor = rowInputRefs.current.get(
          inputFocusKey(rowKey, focusField, surface),
        );
        if (editor) {
          setScalarEditorDraftValue(rowKey, focusField, surface, editor.value);
        }
        queueScalarSave(
          rowKey,
          focusField,
          {
            continueOnFreshDraft: false,
            preserveInputFocus: true,
            surface,
          },
          editor?.value,
        );
      }, 0);
    },
    [
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      finishSave,
      incidentId,
      nextClientTxnId,
      queueScalarSave,
      registerSameFieldConflict,
      resolvePendingSocketTxn,
      resolveTimelinePasteTargetResolution,
      restoreTimelineFocusAnchor,
      setScalarEditorDraftValue,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
      setActiveConflictKey,
      setPasteConflictGroup,
    ],
  );

  const focusDraftRow = useCallback(() => {
    const draftSummary = document.querySelector<HTMLInputElement>(
      dataTestIdSelector(draftCellTestId("timeline.activity_synopsis_text")),
    );
    draftSummary?.focus({ preventScroll: false });
  }, []);

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
    [sendPresenceUpdate, setCurrentPresence],
  );

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
    registerInput,
    rowGutterWidth: timelineRowGutterWidth,
    scalarDraftValuesRef,
    setActiveCollectionInputKey,
    setActiveConflictKey,
    setScalarEditorDraftValue,
    timelineContract,
    updateTimelineSurfaceFocusAnchor,
  });

  const timelineRowGutter = useMemo(
    () => ({
      label: "",
      width: timelineRowGutterWidth,
      minWidth: timelineRowGutterWidth,
    }),
    [],
  );

  const timelineGridRows = useMemo<readonly GridRow<WorkbookRow>[]>(
    () =>
      buildTimelineGridRows({
        onSelectRow: handleSelectRow,
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
        selectedRowId,
      }),
    [
      handleCreateBlankDraftRow,
      handleTimelineEvidenceFiles,
      handleSelectRow,
      presenceForRow,
      rows,
      selectedRowId,
    ],
  );

  useLayoutEffect(() => {
    timelineAnchorColumnsRef.current = timelineColumns;
    timelineAnchorRowsRef.current = timelineGridRows;
  }, [timelineColumns, timelineGridRows]);

  const getTimelineGroupLabel = useCallback(
    (row: WorkbookRow, fieldKey: string) => timelineGroupLabel(row, fieldKey),
    [],
  );
  const getTimelineGroupRowTestId = useCallback(
    (fieldKey: string, value: string) =>
      gridGroupRowTestId(timelineViewSchemaId, fieldKey, value),
    [],
  );

  function renderInspectorFieldEditors(row: WorkbookRow) {
    return (
      <section
        data-testid={timelineInspectorSectionTestId("operational-text")}
        style={inspectorSectionStyle}
      >
        <h3 style={sectionTitleStyle}>Operational Text</h3>
        <div style={inspectorActionStackStyle}>
          {timelineInspectorBindings.map((binding) =>
            renderTimelineInspectorEditor(row, binding),
          )}
        </div>
      </section>
    );
  }

  function renderInspectorRelationshipEditors(row: WorkbookRow) {
    return (
      <div style={inspectorActionStackStyle}>
        {timelineCollectionBindings.map((binding) =>
          renderTimelineCollectionInput(row, binding),
        )}
      </div>
    );
  }

  function renderEvidenceAttachSection(row: WorkbookRow) {
    const countDisplay = buildEvidenceCountDisplayViewModel({
      projectedCount: readTimelineCellValue(
        row.rawRow,
        "timeline.evidence_count",
      ),
      projectedHasEvidence: readTimelineCellValue(
        row.rawRow,
        "timeline.has_evidence",
      ),
    });
    return (
      <TimelineEvidencePanel
        countDisplay={countDisplay}
        row={row}
        onFilesSelected={handleTimelineEvidenceFiles}
      />
    );
  }

  function renderCreateRelatedWorkflowSection() {
    if (createRelatedWorkflow === null) {
      return (
        <p style={bodyStyle}>
          Select a workflow action to create a related row.
        </p>
      );
    }
    const workflow = createRelatedWorkflow;
    const writableFields = workflow.targetContract.fields.filter(
      (field) => field.writeKind !== "read_only",
    );
    return (
      <div style={inspectorActionStackStyle}>
        <p style={bodyStyle}>{workflow.targetContract.viewSchemaId}</p>
        {writableFields.map((field) => {
          const controlId = `timeline-create-related-${workflow.featureGroup.featureGroupKey}-${field.fieldKey}`;
          return (
            <label htmlFor={controlId} key={field.fieldKey} style={labelStyle}>
              {field.label}
              <GenericMutationControl
                collectionMode="add"
                field={field}
                id={controlId}
                referenceOptions={timelineCreateRelatedReferenceOptions}
                testId={genericCreateFieldTestId(field.fieldKey)}
                value={workflow.draft[field.fieldKey] ?? ""}
                onChange={(value) => {
                  updateCreateRelatedWorkflowDraft(
                    workflow.featureGroup.featureGroupKey,
                    field.fieldKey,
                    value,
                  );
                }}
              />
            </label>
          );
        })}
        <div style={inlineButtonRowStyle}>
          <button
            data-testid={genericCreateSubmitTestId(
              workflow.targetContract.viewSchemaId,
            )}
            disabled={workflow.isSubmitting}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => {
              void submitCreateRelatedWorkflow();
            }}
          >
            Create related row
          </button>
          <button
            disabled={workflow.isSubmitting}
            style={actionButtonStyle}
            type="button"
            onClick={() => {
              cancelCreateRelatedWorkflow();
            }}
          >
            Cancel
          </button>
        </div>
        {workflow.message ? (
          <p role="alert" style={bodyStyle}>
            {workflow.message}
          </p>
        ) : null}
      </div>
    );
  }

  function renderRowHistorySection() {
    return (
      <TimelineHistoryPanel
        currentRecordId={currentHistoryRecordId}
        history={rowHistory}
        pendingAction={rowHistoryPendingAction}
        selectedActiveRowRecordId={
          inspectorHistorySubject.kind === "live"
            ? inspectorHistorySubject.recordId
            : null
        }
        onCancelPendingAction={() => {
          setRowHistoryPendingAction(null);
        }}
        onConfirmPendingAction={confirmRowHistoryPendingAction}
        onOpenHistory={openRowHistory}
        onPreviewDeleteRestore={previewRowHistoryDeleteRestore}
        onPreviewRollback={previewRowHistoryRollback}
      />
    );
  }

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

  if (isInitialLoading) {
    return (
      <section style={panelStyle}>
        <p style={eyebrowStyle}>Timeline</p>
        <h1 style={headlineStyle}>Loading projection-backed rows.</h1>
      </section>
    );
  }

  if (loadError !== null) {
    return (
      <section style={panelStyle}>
        <p style={eyebrowStyle}>Timeline</p>
        <h1 style={headlineStyle}>Timeline load failed.</h1>
        <p style={bodyStyle}>{loadError}</p>
      </section>
    );
  }

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
          columns={timelineColumns}
          density={density}
          getGroupLabel={getTimelineGroupLabel}
          getGroupRowTestId={getTimelineGroupRowTestId}
          groupBy={queryState.groupBy}
          onToggleSort={handleQuerySortToggle}
          ref={gridShellRef}
          rowGutter={timelineRowGutter}
          rows={rows}
          sort={queryState.sort}
          style={timelineGridShellStyle}
          timelineGridRows={timelineGridRows}
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
          leading={
            <>
              {savedViewSelector}
              {renderInlineQueryControls ? (
                <WorkbookGridControls
                  contract={timelineContract}
                  defaultFilterPopoverOpen
                  filterDraft={filterDraft}
                  onApplyFilter={applyQueryFilter}
                  onClearAll={() => {
                    setQueryState(emptyWorkbookQueryState());
                    setFilterDraft(defaultFilterDraft(timelineContract));
                  }}
                  onFilterDraftChange={setFilterDraft}
                  onGroupByChange={handleQueryGroupByChange}
                  onRemoveFilter={(fieldKey) => {
                    setQueryState((current) =>
                      removeFilterField(current, fieldKey),
                    );
                  }}
                  onToggleSort={handleQuerySortToggle}
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
          {visibleRefreshError !== null ? (
            <aside
              data-testid="timeline-refresh-error"
              role="status"
              style={timelineNoticeOverlayStyle}
            >
              <p style={bodyStyle}>{visibleRefreshError}</p>
            </aside>
          ) : null}
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
