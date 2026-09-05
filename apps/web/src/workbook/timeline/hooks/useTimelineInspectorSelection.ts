import {
  type Dispatch,
  type MutableRefObject,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  type SetStateAction,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { WorkbookContinuityAnchor } from "../../continuity/workbookContinuityPort";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryState,
  workbookRecordHistoryLoadedData,
} from "../../inspector/workbookRecordHistoryModel";
import type { WorkbookInspectorState } from "../../models/workbookInspectorModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import type { LocalConflictState } from "../models/timelineConflictState";
import type { TimelineRowContextMenuPosition } from "../models/timelineControllerPorts";
import type { CollectionFieldKey } from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";
import {
  buildInspectorMentions,
  type DismissedMention,
} from "../models/workbookMentionChips";

type TimelineRowsRef = {
  readonly current: readonly WorkbookRow[];
};

type TimelineRowContextMenuState = {
  readonly position: TimelineRowContextMenuPosition;
  readonly recordId: string;
};

export function useTimelineInspectorSelection({
  currentIncidentRole,
  dismissedMentionsByRow,
  rows,
  selectedMentionRef,
}: {
  readonly currentIncidentRole: string | null | undefined;
  readonly dismissedMentionsByRow: Record<string, DismissedMention[]>;
  readonly rows: readonly WorkbookRow[];
  readonly selectedMentionRef: string | null;
}) {
  const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
  const [inspectorMessage, setInspectorMessage] =
    useState<WorkbookInspectorFeedback | null>(null);
  const selectedRow = useMemo(
    () =>
      rows.find(
        (row) => row.recordId !== null && row.recordId === selectedRowId,
      ) ?? null,
    [rows, selectedRowId],
  );
  const draftRow = useMemo(
    () => rows.find((row) => row.recordId === null) ?? null,
    [rows],
  );
  const dismissedForSelectedRow = selectedRow?.recordId
    ? (dismissedMentionsByRow[selectedRow.recordId] ?? [])
    : [];
  const inspectorMentions = useMemo(
    () =>
      buildInspectorMentions(selectedRow ?? undefined, dismissedForSelectedRow),
    [dismissedForSelectedRow, selectedRow],
  );
  const selectedMention =
    inspectorMentions.find((item) => item.itemRef === selectedMentionRef) ??
    inspectorMentions[0] ??
    null;
  const canManageMentions =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  return {
    commands: {
      setInspectorMessage,
      setSelectedRowId,
    },
    snapshot: {
      canManageMentions,
      draftRow,
      dismissedForSelectedRow,
      inspectorMentions,
      inspectorMessage,
      selectedMention,
      selectedRow,
      selectedRowId,
    },
  };
}

export function useTimelineInspectorRowInteractions({
  elementRegistry,
  publishViewingPresence,
  rows,
  rowsRef,
  selectedRowId,
  setInspectorMessage,
  setIsInspectorOpen,
  setSelectedMentionRef,
  setSelectedRowId,
}: {
  readonly elementRegistry: TimelineInspectorElementRegistry;
  readonly publishViewingPresence: (recordId: string) => void;
  readonly rows: readonly WorkbookRow[];
  readonly rowsRef: TimelineRowsRef;
  readonly selectedRowId: string | null;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly setIsInspectorOpen: Dispatch<SetStateAction<boolean>>;
  readonly setSelectedMentionRef: Dispatch<SetStateAction<string | null>>;
  readonly setSelectedRowId: Dispatch<SetStateAction<string | null>>;
}) {
  const [rowContextMenu, setRowContextMenu] =
    useState<TimelineRowContextMenuState | null>(null);
  const [pendingMentionFocus, setPendingMentionFocus] = useState<{
    readonly identity: {
      readonly recordId: string;
      readonly rowVersion: number;
      readonly viewSchemaId: string;
    };
    readonly itemRef: string;
    readonly sourceRecordId: string;
    readonly fieldKey?: CollectionFieldKey;
  } | null>(null);
  const rowContextMenuFallbackFocusRef = useRef<HTMLElement | null>(null);
  const rowContextMenuReturnFocusRef = useRef<HTMLElement | null>(null);

  const handleSelectRow = useCallback(
    (recordId: string) => {
      setSelectedRowId(recordId);
      setInspectorMessage(null);
      publishViewingPresence(recordId);
    },
    [publishViewingPresence, setInspectorMessage, setSelectedRowId],
  );

  const openInspectorForRow = useCallback(
    (recordId: string) => {
      handleSelectRow(recordId);
      setIsInspectorOpen(true);
    },
    [handleSelectRow, setIsInspectorOpen],
  );

  const timelineRowForEventTarget = useCallback(
    (target: EventTarget | null) => {
      if (!(target instanceof Element)) {
        return null;
      }
      const rowElement = target.closest<HTMLElement>("[data-grid-record-id]");
      const recordId = rowElement?.dataset.gridRecordId ?? "";
      if (recordId === "") {
        return null;
      }
      return (
        rowsRef.current.find((candidate) => candidate.recordId === recordId) ??
        null
      );
    },
    [rowsRef],
  );

  const openTimelineRowContextMenu = useCallback(
    (
      row: WorkbookRow,
      position: TimelineRowContextMenuPosition,
      returnFocusTarget: HTMLElement | null,
    ) => {
      if (row.recordId === null) {
        return;
      }
      handleSelectRow(row.recordId);
      rowContextMenuReturnFocusRef.current = returnFocusTarget;
      setRowContextMenu({
        position,
        recordId: row.recordId,
      });
    },
    [handleSelectRow],
  );

  const handleTimelineGridContextMenu = useCallback(
    (event: ReactMouseEvent<HTMLDivElement>) => {
      const row = timelineRowForEventTarget(event.target);
      if (row?.recordId === null || row?.recordId === undefined) {
        return;
      }
      event.preventDefault();
      event.stopPropagation();
      rowContextMenuFallbackFocusRef.current = event.currentTarget;
      openTimelineRowContextMenu(
        row,
        {
          x: event.clientX,
          y: event.clientY,
        },
        event.target instanceof HTMLElement
          ? event.target
          : event.currentTarget,
      );
    },
    [openTimelineRowContextMenu, timelineRowForEventTarget],
  );

  const handleTimelineGridContextKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLDivElement>) => {
      if (
        !(
          event.key === "ContextMenu" ||
          (event.key === "F10" && event.shiftKey)
        )
      ) {
        return;
      }
      const row = timelineRowForEventTarget(event.target);
      if (row?.recordId === null || row?.recordId === undefined) {
        return;
      }
      const targetElement =
        event.target instanceof Element
          ? event.target.closest<HTMLElement>(
              "[role='gridcell'], [role='rowheader'], [data-grid-record-id]",
            )
          : null;
      const targetRect = targetElement?.getBoundingClientRect();
      const fallbackRect = event.currentTarget.getBoundingClientRect();
      event.preventDefault();
      event.stopPropagation();
      rowContextMenuFallbackFocusRef.current = event.currentTarget;
      openTimelineRowContextMenu(
        row,
        {
          x: targetRect ? targetRect.left + 12 : fallbackRect.left + 16,
          y: targetRect ? targetRect.top + 12 : fallbackRect.top + 16,
        },
        targetElement ?? event.currentTarget,
      );
    },
    [openTimelineRowContextMenu, timelineRowForEventTarget],
  );

  useEffect(() => {
    if (rowContextMenu === null) {
      return;
    }
    if (!rows.some((row) => row.recordId === rowContextMenu.recordId)) {
      setRowContextMenu(null);
    }
  }, [rowContextMenu, rows]);

  const closeRowContextMenu = useCallback(() => {
    setRowContextMenu(null);
  }, []);

  const activeRowContextMenuRow = useMemo(
    () =>
      rowContextMenu === null
        ? null
        : (rows.find((row) => row.recordId === rowContextMenu.recordId) ??
          null),
    [rowContextMenu, rows],
  );

  const handleSelectMention = useCallback(
    (rowRecordId: string, itemRef: string) => {
      const rowVersion =
        rowsRef.current.find((row) => row.recordId === rowRecordId)
          ?.rowVersion ?? null;
      setSelectedRowId(rowRecordId);
      setSelectedMentionRef(itemRef);
      setInspectorMessage(null);
      setIsInspectorOpen(true);
      if (rowVersion !== null) {
        setPendingMentionFocus({
          identity: {
            recordId: rowRecordId,
            rowVersion,
            viewSchemaId: timelineViewSchemaId,
          },
          itemRef,
          sourceRecordId: rowRecordId,
        });
      }
    },
    [
      rowsRef,
      setInspectorMessage,
      setIsInspectorOpen,
      setSelectedMentionRef,
      setSelectedRowId,
    ],
  );

  const handleInspectCollection = useCallback(
    (recordId: string, fieldKey: CollectionFieldKey, itemRef: string) => {
      const row = rowsRef.current.find(
        (candidate) => candidate.recordId === recordId,
      );
      if (row?.rowVersion == null) return;
      const items =
        fieldKey === "timeline.tags"
          ? row.collectionValues.tags
          : fieldKey === "timeline.host_refs"
            ? row.collectionValues.hostRefs
            : row.collectionValues.identityRefs;
      if (!items.some((item) => item.itemRef === itemRef)) return;
      setSelectedRowId(recordId);
      if (fieldKey !== "timeline.tags") setSelectedMentionRef(itemRef);
      setInspectorMessage(null);
      setIsInspectorOpen(true);
      setPendingMentionFocus({
        identity: {
          recordId,
          rowVersion: row.rowVersion,
          viewSchemaId: timelineViewSchemaId,
        },
        itemRef,
        sourceRecordId: recordId,
        fieldKey,
      });
    },
    [
      rowsRef,
      setSelectedRowId,
      setSelectedMentionRef,
      setInspectorMessage,
      setIsInspectorOpen,
    ],
  );

  useLayoutEffect(() => {
    if (pendingMentionFocus === null) return;
    const row = rowsRef.current.find(
      (candidate) =>
        candidate.recordId === pendingMentionFocus.identity.recordId,
    );
    if (
      selectedRowId !== pendingMentionFocus.identity.recordId ||
      row?.rowVersion !== pendingMentionFocus.identity.rowVersion
    ) {
      setPendingMentionFocus(null);
      return;
    }
    if (pendingMentionFocus.fieldKey !== undefined) {
      if (
        elementRegistry.focusCollectionItem(
          pendingMentionFocus.identity,
          pendingMentionFocus.fieldKey,
          pendingMentionFocus.itemRef,
        )
      )
        setPendingMentionFocus(null);
      return;
    }
    if (
      elementRegistry.containsActiveElement() ||
      elementRegistry.focusMention(
        pendingMentionFocus.identity,
        pendingMentionFocus.sourceRecordId,
        pendingMentionFocus.itemRef,
      )
    )
      setPendingMentionFocus(null);
  });

  return {
    commands: {
      closeRowContextMenu,
      handleInspectCollection,
      handleSelectMention,
      handleSelectRow,
      handleTimelineGridContextKeyDown,
      handleTimelineGridContextMenu,
      openInspectorForRow,
      setRowContextMenu,
      timelineRowForEventTarget,
    },
    snapshot: {
      activeRowContextMenuRow,
      rowContextMenu,
      rowContextMenuFallbackFocusRef,
      rowContextMenuReturnFocusRef,
      selectedRowId,
    },
  };
}

export function useTimelineInspectorLifecycle({
  cancelRowHistoryRequests,
  clearRowHistory,
  gridShellRef,
  inspectorInvalidationCause,
  inspectorMentions,
  inspectorInvalidationGeneration,
  restoreTimelineFocusAnchor,
  rowHistory,
  rows,
  selectedMentionRef,
  selectedRowId,
  dispatchRowHistory,
  setInspectorMessage,
  setIsInspectorOpen,
  setSelectedMentionRef,
  setSelectedResolveTargetId,
  setSelectedRowId,
  workbookFocusAnchorRef,
}: {
  readonly cancelRowHistoryRequests: () => void;
  readonly clearRowHistory: () => void;
  readonly gridShellRef: MutableRefObject<HTMLDivElement | null>;
  readonly inspectorInvalidationCause: WorkbookInspectorState["invalidationCause"];
  readonly inspectorMentions: readonly { readonly itemRef: string }[];
  readonly inspectorInvalidationGeneration: number;
  readonly restoreTimelineFocusAnchor: (
    anchor: WorkbookContinuityAnchor,
  ) => boolean;
  readonly rowHistory: WorkbookRecordHistoryState;
  readonly rows: readonly WorkbookRow[];
  readonly selectedMentionRef: string | null;
  readonly selectedRowId: string | null;
  readonly dispatchRowHistory: (event: WorkbookRecordHistoryEvent) => void;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly setIsInspectorOpen: Dispatch<SetStateAction<boolean>>;
  readonly setSelectedMentionRef: Dispatch<SetStateAction<string | null>>;
  readonly setSelectedResolveTargetId: Dispatch<SetStateAction<string>>;
  readonly setSelectedRowId: Dispatch<SetStateAction<string | null>>;
  readonly workbookFocusAnchorRef: MutableRefObject<WorkbookContinuityAnchor | null>;
}) {
  const rowHistoryData = workbookRecordHistoryLoadedData(rowHistory);
  const closeInspector = useCallback(() => {
    setIsInspectorOpen(false);
  }, [setIsInspectorOpen]);

  useEffect(() => {
    if (selectedRowId === null) {
      return;
    }
    if (!rows.some((row) => row.recordId === selectedRowId)) {
      const deletedHistoryMatchesSelectedRow =
        rowHistoryData?.deleted === true &&
        rowHistoryData.record_id === selectedRowId;
      const previousAnchor = workbookFocusAnchorRef.current;
      setSelectedRowId(null);
      setSelectedMentionRef(null);
      setSelectedResolveTargetId("");
      if (
        rowHistory.subject?.recordId === selectedRowId &&
        rowHistoryData?.deleted !== true
      ) {
        cancelRowHistoryRequests();
        dispatchRowHistory({ type: "clear" });
      }
      dispatchRowHistory({ type: "cancel" });
      setInspectorMessage(
        workbookInspectorMessageFeedback(
          deletedHistoryMatchesSelectedRow
            ? "Selected row was deleted."
            : "Selected row is no longer available.",
          "none",
        ),
      );
      if (deletedHistoryMatchesSelectedRow) {
        return;
      }
      window.setTimeout(() => {
        const fallbackFieldKey =
          previousAnchor?.viewSchemaId === timelineViewSchemaId
            ? previousAnchor.fieldKey
            : "timeline.activity_synopsis_text";
        const fallbackRow = rows.find((row) => row.recordId !== null);
        if (fallbackRow?.recordId) {
          if (
            restoreTimelineFocusAnchor({
              fieldKey: fallbackFieldKey,
              recordId: fallbackRow.recordId,
              viewSchemaId: timelineViewSchemaId,
            })
          ) {
            return;
          }
        }
        const gridShell = gridShellRef.current;
        if (gridShell !== null) {
          if (!gridShell.hasAttribute("tabindex")) {
            gridShell.tabIndex = -1;
          }
          gridShell.focus({ preventScroll: true });
        }
      }, 0);
    }
  }, [
    cancelRowHistoryRequests,
    dispatchRowHistory,
    gridShellRef,
    restoreTimelineFocusAnchor,
    rowHistoryData,
    rowHistory.subject?.recordId,
    rows,
    selectedRowId,
    setInspectorMessage,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
    setSelectedRowId,
    workbookFocusAnchorRef,
  ]);

  useEffect(() => {
    if (inspectorMentions.length < 1) {
      if (selectedMentionRef !== null) {
        setSelectedMentionRef(null);
      }
      setSelectedResolveTargetId("");
      return;
    }
    if (
      selectedMentionRef !== null &&
      inspectorMentions.some((item) => item.itemRef === selectedMentionRef)
    ) {
      return;
    }
    const [firstMention] = inspectorMentions;
    if (firstMention) {
      setSelectedMentionRef(firstMention.itemRef);
    }
    setSelectedResolveTargetId("");
  }, [
    inspectorMentions,
    selectedMentionRef,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
  ]);

  const previousInvalidationGenerationRef = useRef(
    inspectorInvalidationGeneration,
  );
  useLayoutEffect(() => {
    if (
      previousInvalidationGenerationRef.current ===
      inspectorInvalidationGeneration
    ) {
      return;
    }
    previousInvalidationGenerationRef.current = inspectorInvalidationGeneration;
    setSelectedMentionRef(null);
    setSelectedResolveTargetId("");
    dispatchRowHistory({ type: "cancel" });
    if (inspectorInvalidationCause !== "retarget") {
      clearRowHistory();
    }
  }, [
    clearRowHistory,
    dispatchRowHistory,
    inspectorInvalidationCause,
    inspectorInvalidationGeneration,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
  ]);

  return { closeInspector };
}

export function useTimelineInspectorEscape({
  activeConflict,
  clearRowHistory,
  isInspectorOpen,
  restoreTimelineFocusAnchor,
  setInspectorMessage,
  setIsInspectorOpen,
  setSelectedMentionRef,
  setSelectedRowId,
  workbookFocusAnchorRef,
}: {
  readonly activeConflict: LocalConflictState | null;
  readonly clearRowHistory: () => void;
  readonly isInspectorOpen: boolean;
  readonly restoreTimelineFocusAnchor: (
    anchor: WorkbookContinuityAnchor,
  ) => boolean;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly setIsInspectorOpen: Dispatch<SetStateAction<boolean>>;
  readonly setSelectedMentionRef: Dispatch<SetStateAction<string | null>>;
  readonly setSelectedRowId: Dispatch<SetStateAction<string | null>>;
  readonly workbookFocusAnchorRef: MutableRefObject<WorkbookContinuityAnchor | null>;
}) {
  useEffect(() => {
    if (!isInspectorOpen) {
      return;
    }
    const handleTimelineInspectorEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape" || activeConflict !== null) {
        return;
      }
      const target = event.target instanceof HTMLElement ? event.target : null;
      if (
        target instanceof HTMLInputElement ||
        target instanceof HTMLTextAreaElement ||
        target instanceof HTMLSelectElement
      ) {
        return;
      }
      event.preventDefault();
      setIsInspectorOpen(false);
      setSelectedRowId(null);
      setSelectedMentionRef(null);
      setInspectorMessage(null);
      clearRowHistory();
      const anchor = workbookFocusAnchorRef.current;
      if (anchor?.viewSchemaId === timelineViewSchemaId) {
        restoreTimelineFocusAnchor(anchor);
      }
    };
    document.addEventListener("keydown", handleTimelineInspectorEscape);
    return () => {
      document.removeEventListener("keydown", handleTimelineInspectorEscape);
    };
  }, [
    activeConflict,
    clearRowHistory,
    isInspectorOpen,
    restoreTimelineFocusAnchor,
    setInspectorMessage,
    setIsInspectorOpen,
    setSelectedMentionRef,
    setSelectedRowId,
    workbookFocusAnchorRef,
  ]);
}
