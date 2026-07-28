import {
  dataTestIdSelector,
  mentionItemTestId,
  timelineInspectorTestId,
} from "@cartulary/ui-contracts";
import {
  type Dispatch,
  type MutableRefObject,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useInspectorLifecycleReset } from "../../hooks/useInspectorLifecycleReset";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookPresenceDraft } from "../../runtime/workbookCollaborationMessages";
import type { WorkbookFocusAnchor } from "../../utils/workbookGridFocus";
import type {
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../components/TimelineHistoryPanel";
import {
  buildInspectorMentions,
  type DismissedMention,
} from "../models/workbookMentionChips";
import type {
  LocalConflictState,
  WorkbookRow,
} from "../models/workbookTimelineModel";

type TimelineRowsRef = {
  readonly current: readonly WorkbookRow[];
};

type TimelineRowContextMenuPosition = {
  readonly x: number;
  readonly y: number;
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
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
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
  const selectedRowWorkflowKey =
    selectedRow?.recordId && selectedRow.rowVersion !== null
      ? `${selectedRow.recordId}:${selectedRow.rowVersion}`
      : (selectedRow?.recordId ?? "");

  return {
    commands: {
      setIsInspectorOpen,
      setSelectedRowId,
    },
    snapshot: {
      canManageMentions,
      draftRow,
      dismissedForSelectedRow,
      isInspectorOpen,
      inspectorMentions,
      selectedMention,
      selectedRow,
      selectedRowId,
      selectedRowWorkflowKey,
    },
  };
}

export function useTimelineInspectorRowInteractions({
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
}: {
  readonly currentPresenceRef: MutableRefObject<WorkbookPresenceDraft>;
  readonly rows: readonly WorkbookRow[];
  readonly rowsRef: TimelineRowsRef;
  readonly selectedRowId: string | null;
  readonly sendPresenceUpdate: (presence: WorkbookPresenceDraft) => void;
  readonly setCurrentPresence: Dispatch<SetStateAction<WorkbookPresenceDraft>>;
  readonly setInspectorMessage: (message: string | null) => void;
  readonly setIsInspectorOpen: Dispatch<SetStateAction<boolean>>;
  readonly setSelectedMentionRef: Dispatch<SetStateAction<string | null>>;
  readonly setSelectedRowId: Dispatch<SetStateAction<string | null>>;
}) {
  const [rowContextMenu, setRowContextMenu] =
    useState<TimelineRowContextMenuState | null>(null);

  const handleSelectRow = useCallback(
    (recordId: string) => {
      setSelectedRowId(recordId);
      setInspectorMessage(null);
      if (currentPresenceRef.current.mode === "editing") {
        return;
      }
      const next = { fieldKey: null, mode: "viewing" as const, recordId };
      currentPresenceRef.current = next;
      setCurrentPresence(next);
      sendPresenceUpdate(next);
    },
    [
      currentPresenceRef,
      sendPresenceUpdate,
      setCurrentPresence,
      setInspectorMessage,
      setSelectedRowId,
    ],
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
    (row: WorkbookRow, position: TimelineRowContextMenuPosition) => {
      if (row.recordId === null) {
        return;
      }
      handleSelectRow(row.recordId);
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
      openTimelineRowContextMenu(row, {
        x: event.clientX,
        y: event.clientY,
      });
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
      openTimelineRowContextMenu(row, {
        x: targetRect ? targetRect.left + 12 : fallbackRect.left + 16,
        y: targetRect ? targetRect.top + 12 : fallbackRect.top + 16,
      });
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
      setSelectedRowId(rowRecordId);
      setSelectedMentionRef(itemRef);
      setInspectorMessage(null);
      setIsInspectorOpen(true);
      window.requestAnimationFrame(() => {
        window.requestAnimationFrame(() => {
          const mentionButton = document.querySelector<HTMLButtonElement>(
            dataTestIdSelector(mentionItemTestId(itemRef)),
          );
          if (mentionButton === null) {
            return;
          }
          const activeElement = document.activeElement;
          if (
            activeElement instanceof HTMLElement &&
            activeElement !== document.body &&
            activeElement !== mentionButton &&
            activeElement.closest(dataTestIdSelector(timelineInspectorTestId()))
          ) {
            return;
          }
          mentionButton.focus({ preventScroll: true });
        });
      });
    },
    [
      setInspectorMessage,
      setIsInspectorOpen,
      setSelectedMentionRef,
      setSelectedRowId,
    ],
  );

  return {
    commands: {
      closeRowContextMenu,
      handleSelectMention,
      handleSelectRow,
      handleTimelineGridContextKeyDown,
      handleTimelineGridContextMenu,
      openInspectorForRow,
      setRowContextMenu,
    },
    snapshot: {
      activeRowContextMenuRow,
      rowContextMenu,
      selectedRowId,
    },
  };
}

export function useTimelineInspectorLifecycle({
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
}: {
  readonly cancelCreateRelatedWorkflow: () => void;
  readonly cancelRowHistoryRequests: () => void;
  readonly clearRowHistory: () => void;
  readonly gridShellRef: MutableRefObject<HTMLDivElement | null>;
  readonly inspectorMentions: readonly { readonly itemRef: string }[];
  readonly inspectorResetKey: string | undefined;
  readonly restoreTimelineFocusAnchor: (
    anchor: Pick<WorkbookFocusAnchor, "fieldKey" | "recordId" | "viewSchemaId">,
  ) => boolean;
  readonly rowHistory: RecordHistoryState;
  readonly rows: readonly WorkbookRow[];
  readonly selectedMentionRef: string | null;
  readonly selectedRowId: string | null;
  readonly setInspectorMessage: (message: string | null) => void;
  readonly setIsInspectorOpen: Dispatch<SetStateAction<boolean>>;
  readonly setRowHistory: Dispatch<SetStateAction<RecordHistoryState>>;
  readonly setRowHistoryPendingAction: Dispatch<
    SetStateAction<RowHistoryPendingAction | null>
  >;
  readonly setSelectedMentionRef: Dispatch<SetStateAction<string | null>>;
  readonly setSelectedResolveTargetId: Dispatch<SetStateAction<string>>;
  readonly setSelectedRowId: Dispatch<SetStateAction<string | null>>;
  readonly workbookFocusAnchorRef: MutableRefObject<WorkbookFocusAnchor | null>;
}) {
  useEffect(() => {
    if (selectedRowId === null) {
      return;
    }
    if (!rows.some((row) => row.recordId === selectedRowId)) {
      const deletedHistoryMatchesSelectedRow =
        rowHistory.data?.deleted === true &&
        rowHistory.data.record_id === selectedRowId;
      const previousAnchor = workbookFocusAnchorRef.current;
      setSelectedRowId(null);
      setSelectedMentionRef(null);
      setSelectedResolveTargetId("");
      setRowHistory((current) => {
        if (
          current.recordId !== selectedRowId ||
          current.data?.deleted === true
        ) {
          return current;
        }
        cancelRowHistoryRequests();
        return {
          recordId: null,
          status: "idle",
          data: null,
          message: null,
        };
      });
      setRowHistoryPendingAction((current) =>
        current?.recordId === selectedRowId ? null : current,
      );
      setInspectorMessage(
        deletedHistoryMatchesSelectedRow
          ? "Selected row was deleted."
          : "Selected row is no longer available.",
      );
      if (deletedHistoryMatchesSelectedRow) {
        return;
      }
      window.setTimeout(() => {
        const fallbackFieldKey =
          previousAnchor?.surface === timelineViewSchemaId
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
    gridShellRef,
    restoreTimelineFocusAnchor,
    rowHistory.data,
    rows,
    selectedRowId,
    setInspectorMessage,
    setRowHistory,
    setRowHistoryPendingAction,
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

  useInspectorLifecycleReset(inspectorResetKey, () => {
    setIsInspectorOpen(false);
    setSelectedRowId(null);
    setSelectedMentionRef(null);
    setSelectedResolveTargetId("");
    setInspectorMessage(null);
    cancelCreateRelatedWorkflow();
    clearRowHistory();
  });
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
    anchor: Pick<WorkbookFocusAnchor, "fieldKey" | "recordId" | "viewSchemaId">,
  ) => boolean;
  readonly setInspectorMessage: (message: string | null) => void;
  readonly setIsInspectorOpen: Dispatch<SetStateAction<boolean>>;
  readonly setSelectedMentionRef: Dispatch<SetStateAction<string | null>>;
  readonly setSelectedRowId: Dispatch<SetStateAction<string | null>>;
  readonly workbookFocusAnchorRef: MutableRefObject<WorkbookFocusAnchor | null>;
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
      if (anchor?.surface === timelineViewSchemaId) {
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
