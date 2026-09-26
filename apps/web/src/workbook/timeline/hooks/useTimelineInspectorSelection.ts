import type { InspectorPanelId } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type MutableRefObject,
  type SetStateAction,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { WorkbookContinuityAnchor } from "../../continuity/workbookContinuityPort";
import { useRetainedInspectorRow } from "../../inspector/useRetainedInspectorRow";
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
import type { WorkbookReadScope } from "../../query/WorkbookQueryRow";
import type { MentionSubject } from "../actions/timelineMentionOperationModel";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import type { LocalConflictState } from "../models/timelineConflictState";
import type {
  DisclosureReviewNavigationScope,
  TimelineCommittedInspectorRecords,
} from "../models/timelineControllerPorts";
import type { CollectionFieldKey } from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";
import {
  buildInspectorMentions,
  type DismissedMention,
} from "../models/workbookMentionChips";

type TimelineRowsRef = {
  readonly current: readonly WorkbookRow[];
};

export function useTimelineInspectorSelection({
  committedRecords,
  readScope,
  currentIncidentRole,
  dismissedMentionsByRow,
  observedMentions,
  rows,
  selectedMentionRef,
}: {
  readonly committedRecords?: TimelineCommittedInspectorRecords;
  readonly readScope: WorkbookReadScope | null;
  readonly currentIncidentRole: string | null | undefined;
  readonly dismissedMentionsByRow: Record<string, DismissedMention[]>;
  readonly observedMentions: readonly MentionSubject[];
  readonly rows: readonly WorkbookRow[];
  readonly selectedMentionRef: string | null;
}) {
  const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
  const [inspectorMessage, setInspectorMessage] =
    useState<WorkbookInspectorFeedback | null>(null);
  useLayoutEffect(() => {
    committedRecords?.retainInspectorRecord(selectedRowId);
    return () => committedRecords?.retainInspectorRecord(null);
  }, [committedRecords, selectedRowId]);
  const selectedRow = useRetainedInspectorRow({
    recordId: selectedRowId,
    rows: [
      selectedRowId
        ? committedRecords?.currentCommittedTimelineRow(selectedRowId)
        : null,
      rows.find(
        (row) => row.recordId === selectedRowId && row.recordId !== null,
      ),
    ],
    sourceRow: (row) => row.rawRow,
    // Role changes invalidate action authority independently. An accepted
    // incident reader still owns the same source projection.
    scope: readScope,
    readable: !!currentIncidentRole,
  });
  const draftRow = useMemo(
    () => rows.find((row) => row.recordId === null) ?? null,
    [rows],
  );
  const dismissedForSelectedRow = selectedRow?.recordId
    ? (dismissedMentionsByRow[selectedRow.recordId] ?? [])
    : [];
  const inspectorMentions = useMemo(
    () =>
      buildInspectorMentions(
        selectedRow ?? undefined,
        dismissedForSelectedRow,
        observedMentions,
      ),
    [dismissedForSelectedRow, selectedRow, observedMentions],
  );
  const selectedMention =
    (selectedMentionRef === null
      ? inspectorMentions[0]
      : inspectorMentions.find(
          (item) => item.itemRef === selectedMentionRef,
        )) ?? null;
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
  currentCommittedRow,
  elementRegistry,
  publishViewingPresence,
  rowsRef,
  selectedRowId,
  setInspectorMessage,
  setIsInspectorOpen,
  setSelectedMentionRef,
  setSelectedRowId,
}: {
  readonly currentCommittedRow: (recordId: string) => WorkbookRow | null;
  readonly elementRegistry: TimelineInspectorElementRegistry;
  readonly publishViewingPresence: (recordId: string) => void;
  readonly rowsRef: TimelineRowsRef;
  readonly selectedRowId: string | null;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly setIsInspectorOpen: Dispatch<SetStateAction<boolean>>;
  readonly setSelectedMentionRef: Dispatch<SetStateAction<string | null>>;
  readonly setSelectedRowId: Dispatch<SetStateAction<string | null>>;
}) {
  const [pendingPanelFocus, setPendingPanelFocus] = useState<{
    readonly recordId: string;
    readonly panel: InspectorPanelId;
  } | null>(null);
  const requestRowPanelFocus = useCallback(
    (recordId: string, panel: InspectorPanelId) => {
      setPendingPanelFocus({ recordId, panel });
    },
    [],
  );
  useLayoutEffect(() => {
    if (!pendingPanelFocus) return;
    const row = rowsRef.current.find(
      (row) => row.recordId === pendingPanelFocus.recordId,
    );
    if (
      !row ||
      row.rowVersion === null ||
      selectedRowId !== pendingPanelFocus.recordId
    ) {
      setPendingPanelFocus(null);
      return;
    }
    if (
      elementRegistry.focusPanel(
        {
          recordId: pendingPanelFocus.recordId,
          rowVersion: row.rowVersion,
          viewSchemaId: timelineViewSchemaId,
        },
        pendingPanelFocus.panel,
      )
    )
      setPendingPanelFocus(null);
  });
  useEffect(() => {
    if (!pendingPanelFocus) return;
    const cancel = () => setPendingPanelFocus(null);
    document.addEventListener("pointerdown", cancel, true);
    document.addEventListener("keydown", cancel, true);
    document.addEventListener("focusin", cancel, true);
    return () => {
      document.removeEventListener("pointerdown", cancel, true);
      document.removeEventListener("keydown", cancel, true);
      document.removeEventListener("focusin", cancel, true);
    };
  }, [pendingPanelFocus]);
  const [pendingMentionFocus, setPendingMentionFocus] = useState<{
    readonly identity: {
      readonly recordId: string;
      readonly rowVersion: number;
      readonly viewSchemaId: string;
    };
    readonly itemRef: string;
    readonly sourceRecordId: string;
    readonly fieldKey?: CollectionFieldKey;
    readonly reviewScope?: DisclosureReviewNavigationScope;
  } | null>(null);

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

  const handleSelectMention = useCallback(
    (
      rowRecordId: string,
      itemRef: string,
      reviewScope?: DisclosureReviewNavigationScope,
    ) => {
      const rowVersion =
        currentCommittedRow(rowRecordId)?.rowVersion ??
        rowsRef.current.find((row) => row.recordId === rowRecordId)
          ?.rowVersion ??
        null;
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
          ...(reviewScope ? { reviewScope } : {}),
        });
      } else reviewScope?.settle(false);
    },
    [
      currentCommittedRow,
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
    const { reviewScope } = pendingMentionFocus;
    if (reviewScope && !reviewScope.isCurrent()) {
      setPendingMentionFocus(null);
      reviewScope.settle(false);
      return;
    }
    const row =
      currentCommittedRow(pendingMentionFocus.identity.recordId) ??
      rowsRef.current.find(
        (candidate) =>
          candidate.recordId === pendingMentionFocus.identity.recordId,
      );
    if (
      selectedRowId !== pendingMentionFocus.identity.recordId ||
      row?.rowVersion !== pendingMentionFocus.identity.rowVersion
    ) {
      setPendingMentionFocus(null);
      reviewScope?.settle(false);
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
    const focused =
      reviewScope === undefined
        ? elementRegistry.containsActiveElement() ||
          elementRegistry.focusMention(
            pendingMentionFocus.identity,
            pendingMentionFocus.sourceRecordId,
            pendingMentionFocus.itemRef,
          )
        : reviewScope.runOwnedFocus(() =>
            elementRegistry.focusMention(
              pendingMentionFocus.identity,
              pendingMentionFocus.sourceRecordId,
              pendingMentionFocus.itemRef,
            ),
          );
    if (focused) {
      setPendingMentionFocus(null);
      reviewScope?.settle(true);
    }
  });

  return {
    commands: {
      handleInspectCollection,
      handleSelectMention,
      handleSelectRow,
      openInspectorForRow,
      requestRowPanelFocus,
      timelineRowForEventTarget,
    },
    snapshot: {
      selectedRowId,
    },
  };
}

export function useTimelineInspectorLifecycle({
  clearRowHistory,
  inspectorInvalidationCause,
  inspectorMentions,
  inspectorInvalidationGeneration,
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
}: {
  readonly clearRowHistory: () => void;
  readonly gridShellRef: MutableRefObject<HTMLDivElement | null>;
  readonly inspectorInvalidationCause: WorkbookInspectorState["invalidationCause"];
  readonly inspectorMentions: readonly { readonly itemRef: string }[];
  readonly inspectorInvalidationGeneration: number;
  readonly restoreTimelineFocusAnchor: (
    anchor: WorkbookContinuityAnchor,
  ) => Promise<boolean>;
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
    if (
      selectedRowId === null ||
      rows.some((row) => row.recordId === selectedRowId) ||
      rowHistoryData?.deleted !== true ||
      rowHistoryData.record_id !== selectedRowId
    )
      return;
    setSelectedRowId(null);
    setSelectedMentionRef(null);
    setSelectedResolveTargetId("");
    dispatchRowHistory({ type: "cancel" });
    setInspectorMessage(
      workbookInspectorMessageFeedback("Selected row was deleted.", "none"),
    );
  }, [
    dispatchRowHistory,
    rowHistoryData,
    rows,
    selectedRowId,
    setInspectorMessage,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
    setSelectedRowId,
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

  const previousInvalidationGenerationRef = useRef({
    generation: inspectorInvalidationGeneration,
    recordId: selectedRowId,
  });
  useLayoutEffect(() => {
    if (
      previousInvalidationGenerationRef.current.generation ===
      inspectorInvalidationGeneration
    ) {
      return;
    }
    const sameSourceRetarget =
      inspectorInvalidationCause === "retarget" &&
      previousInvalidationGenerationRef.current.recordId === selectedRowId;
    previousInvalidationGenerationRef.current = {
      generation: inspectorInvalidationGeneration,
      recordId: selectedRowId,
    };
    if (!sameSourceRetarget) {
      setSelectedMentionRef(null);
      setSelectedResolveTargetId("");
    }
    dispatchRowHistory({ type: "cancel" });
    if (inspectorInvalidationCause !== "retarget") {
      clearRowHistory();
    }
  }, [
    clearRowHistory,
    dispatchRowHistory,
    inspectorInvalidationCause,
    inspectorInvalidationGeneration,
    selectedRowId,
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
  ) => Promise<boolean>;
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
        void restoreTimelineFocusAnchor(anchor);
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
