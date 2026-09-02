import type {
  GridCellAnchor,
  GridNavigationIntent,
} from "@cartulary/grid-adapter";
import { type KeyboardEvent as ReactKeyboardEvent, useCallback } from "react";
import type { WorkbookContinuityAnchor } from "../../continuity/workbookContinuityPort";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookRecordHistoryState } from "../../inspector/workbookRecordHistoryModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import type { TimelineScalarSaveOptions } from "../models/timelineControllerPorts";
import {
  mapTimelineCollectionEditorIntent,
  mapTimelineScalarEditorIntent,
  mapTimelineWorkAreaInspectorIntent,
  type TimelineEditorKeyboardIntent,
} from "../models/timelineKeyboardIntentModel";
import {
  type CollectionDraftKey,
  type CollectionFieldKey,
  type RowValues,
  type TimelineScalarEditorSurface,
  timelineScalarBindings,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

type QueueScalarSave = (
  rowKey: string,
  focusField: keyof RowValues,
  options: TimelineScalarSaveOptions,
  currentValue?: string,
) => void;

type QueueCollectionSave = (
  rowKey: string,
  fieldKey: CollectionFieldKey,
  draftKey: CollectionDraftKey,
  currentValue?: string,
  source?: "keyboard" | "blur",
) => void;

function executeScalarEditorIntent({
  anchor,
  closeInspector,
  focusField,
  intent,
  navigate,
  priorGridAnchor,
  queueSave,
  recordTiming,
  restoreFocus,
  rowKey,
  surface,
  value,
}: {
  readonly anchor: GridCellAnchor | null;
  readonly closeInspector: (anchor: GridCellAnchor | null) => boolean;
  readonly focusField: keyof RowValues;
  readonly intent: TimelineEditorKeyboardIntent;
  readonly navigate: (
    anchor: GridCellAnchor,
    intent: GridNavigationIntent,
  ) => void;
  readonly priorGridAnchor: WorkbookContinuityAnchor | null;
  readonly queueSave: QueueScalarSave;
  readonly recordTiming: (
    name: string,
    details?: Record<string, unknown>,
  ) => void;
  readonly restoreFocus: (anchor: WorkbookContinuityAnchor) => unknown;
  readonly rowKey: string;
  readonly surface: TimelineScalarEditorSurface;
  readonly value: string;
}) {
  switch (intent.kind) {
    case "restore_prior_grid_focus":
      if (priorGridAnchor !== null) restoreFocus(priorGridAnchor);
      return;
    case "close_inspector":
      closeInspector(anchor);
      return;
    case "navigate":
      if (anchor !== null) navigate(anchor, intent.navigation);
      return;
    case "save":
      executeScalarSaveIntent({
        anchor,
        focusField,
        intent,
        navigate,
        queueSave,
        recordTiming,
        rowKey,
        surface,
        value,
      });
      return;
    case "none":
      return;
  }
}

function executeScalarSaveIntent({
  anchor,
  focusField,
  intent,
  navigate,
  queueSave,
  recordTiming,
  rowKey,
  surface,
  value,
}: {
  readonly anchor: GridCellAnchor | null;
  readonly focusField: keyof RowValues;
  readonly intent: Extract<TimelineEditorKeyboardIntent, { kind: "save" }>;
  readonly navigate: (
    anchor: GridCellAnchor,
    intent: GridNavigationIntent,
  ) => void;
  readonly queueSave: QueueScalarSave;
  readonly recordTiming: (
    name: string,
    details?: Record<string, unknown>,
  ) => void;
  readonly rowKey: string;
  readonly surface: TimelineScalarEditorSurface;
  readonly value: string;
}) {
  if (intent.recordBlankRowTiming) {
    recordTiming("blank_row_commit_accepted", {
      field: "timeline.activity_synopsis_text",
      surface,
    });
  }
  queueSave(
    rowKey,
    focusField,
    {
      continueOnFreshDraft: true,
      preserveInputFocus: intent.preserveInputFocus,
      surface,
    },
    value,
  );
  if (intent.navigateAfterSave !== null && anchor !== null) {
    navigate(anchor, intent.navigateAfterSave);
  }
}

export function useTimelineKeyboardController({
  clearRowHistory,
  currentTimelineAnchorFor,
  elementRegistry,
  handleTimelineGridContextKeyDown,
  navigateTimelineFocusAnchor,
  openRowHistory,
  queueCollectionSave,
  queueScalarSave,
  recordTiming,
  restoreTimelineFocusAnchor,
  rowHistory,
  selectedRowId,
  setInspectorMessage,
  setIsInspectorOpen,
  setSelectedMentionRef,
  setSelectedRowId,
  timelineRowForEventTarget,
  workbookFocusAnchorRef,
}: {
  readonly clearRowHistory: () => void;
  readonly currentTimelineAnchorFor: (
    rowKey: string,
    fieldKey: string,
  ) => GridCellAnchor | null;
  readonly elementRegistry: TimelineInspectorElementRegistry;
  readonly handleTimelineGridContextKeyDown: (
    event: ReactKeyboardEvent<HTMLDivElement>,
  ) => void;
  readonly navigateTimelineFocusAnchor: (
    anchor: GridCellAnchor,
    intent: GridNavigationIntent,
  ) => void;
  readonly openRowHistory: (recordId: string) => void;
  readonly queueCollectionSave: QueueCollectionSave;
  readonly queueScalarSave: QueueScalarSave;
  readonly recordTiming: (
    name: string,
    details?: Record<string, unknown>,
  ) => void;
  readonly restoreTimelineFocusAnchor: (
    anchor: GridCellAnchor | WorkbookContinuityAnchor,
  ) => unknown;
  readonly rowHistory: WorkbookRecordHistoryState;
  readonly selectedRowId: string | null;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly setIsInspectorOpen: (isOpen: boolean) => void;
  readonly setSelectedMentionRef: (itemRef: string | null) => void;
  readonly setSelectedRowId: (recordId: string | null) => void;
  readonly timelineRowForEventTarget: (target: Element) => WorkbookRow | null;
  readonly workbookFocusAnchorRef: {
    readonly current: WorkbookContinuityAnchor | null;
  };
}) {
  const closeInspectorFromEditor = useCallback(
    (anchor: GridCellAnchor | null) => {
      if (anchor === null) return false;
      setSelectedRowId(null);
      setSelectedMentionRef(null);
      setInspectorMessage(null);
      clearRowHistory();
      restoreTimelineFocusAnchor(anchor);
      return true;
    },
    [
      clearRowHistory,
      restoreTimelineFocusAnchor,
      setInspectorMessage,
      setSelectedMentionRef,
      setSelectedRowId,
    ],
  );

  const onScalarEditorKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
    ) => {
      const priorGridAnchor = workbookFocusAnchorRef.current;
      const binding = timelineScalarBindings.find(
        (candidate) => candidate.key === focusField,
      );
      const fieldKey = binding?.fieldKey ?? focusField;
      const anchor = currentTimelineAnchorFor(rowKey, fieldKey);
      const intent = mapTimelineScalarEditorIntent({
        event,
        focusField,
        hasCommittedAnchor: anchor !== null,
        inspectorCanClose:
          surface === "inspector" ||
          selectedRowId !== null ||
          rowHistory.subject !== null ||
          rowHistory.phase !== "idle",
        priorTimelineGridAnchor:
          priorGridAnchor?.viewSchemaId === timelineViewSchemaId,
        surface,
      });
      if (intent.preventDefault) event.preventDefault();
      executeScalarEditorIntent({
        anchor,
        closeInspector: closeInspectorFromEditor,
        focusField,
        intent,
        navigate: navigateTimelineFocusAnchor,
        priorGridAnchor,
        queueSave: queueScalarSave,
        recordTiming,
        restoreFocus: restoreTimelineFocusAnchor,
        rowKey,
        surface,
        value: event.currentTarget.value,
      });
    },
    [
      closeInspectorFromEditor,
      currentTimelineAnchorFor,
      navigateTimelineFocusAnchor,
      queueScalarSave,
      recordTiming,
      restoreTimelineFocusAnchor,
      rowHistory.phase,
      rowHistory.subject,
      selectedRowId,
      workbookFocusAnchorRef,
    ],
  );

  const onCollectionEditorKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement>,
      rowKey: string,
      fieldKey: CollectionFieldKey,
      draftKey: CollectionDraftKey,
    ) => {
      const anchor = currentTimelineAnchorFor(rowKey, fieldKey);
      const intent = mapTimelineCollectionEditorIntent({
        event,
        hasCommittedAnchor: anchor !== null,
        inspectorCanClose:
          selectedRowId !== null ||
          rowHistory.subject !== null ||
          rowHistory.phase !== "idle",
      });
      if (intent.preventDefault) event.preventDefault();
      if (intent.kind === "close_inspector") {
        closeInspectorFromEditor(anchor);
        return;
      }
      if (intent.kind === "navigate" && anchor !== null) {
        navigateTimelineFocusAnchor(anchor, intent.navigation);
        return;
      }
      if (intent.kind !== "save") return;
      queueCollectionSave(
        rowKey,
        fieldKey,
        draftKey,
        event.currentTarget.value,
        "keyboard",
      );
      if (intent.navigateAfterSave !== null && anchor !== null) {
        navigateTimelineFocusAnchor(anchor, intent.navigateAfterSave);
      }
    },
    [
      closeInspectorFromEditor,
      currentTimelineAnchorFor,
      navigateTimelineFocusAnchor,
      queueCollectionSave,
      rowHistory.phase,
      rowHistory.subject,
      selectedRowId,
    ],
  );

  const focusInspectorSection = useCallback(
    (section: "evidence" | "history", row: WorkbookRow) => {
      if (row.recordId === null || row.rowVersion === null) return;
      const identity = {
        recordId: row.recordId,
        rowVersion: row.rowVersion,
        viewSchemaId: timelineViewSchemaId,
      };
      window.requestAnimationFrame(() => {
        elementRegistry.focusPanel(identity, section);
      });
    },
    [elementRegistry],
  );

  const onWorkAreaKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLDivElement>) => {
      handleTimelineGridContextKeyDown(event);
      if (event.defaultPrevented || !(event.target instanceof Element)) return;
      if (
        event.target.closest(
          "input, textarea, select, button, a, [contenteditable='true'], [role='menu'], [role='dialog'], [role='listbox'], [role='option']",
        ) !== null
      ) {
        return;
      }
      const row = timelineRowForEventTarget(event.target);
      const fieldElement = event.target.closest<HTMLElement>(
        "[data-grid-field-key]",
      );
      const fieldKey = fieldElement?.dataset.gridFieldKey;
      const intent = mapTimelineWorkAreaInspectorIntent({
        event,
        fieldKey,
        row,
      });
      if (intent.kind === "none") return;

      event.preventDefault();
      event.stopPropagation();
      const recordId = intent.row.recordId;
      if (recordId === null) return;
      if (intent.kind === "open_panel" && intent.panelId === "history") {
        openRowHistory(recordId);
        focusInspectorSection("history", intent.row);
        return;
      }
      if (intent.kind === "open_panel") {
        setSelectedRowId(recordId);
        setIsInspectorOpen(true);
        setInspectorMessage(null);
        focusInspectorSection("evidence", intent.row);
        return;
      }

      setSelectedRowId(recordId);
      setIsInspectorOpen(true);
      if (intent.itemRef === null) {
        setInspectorMessage(
          workbookInspectorMessageFeedback(
            "No unresolved mention is available for quick link.",
            "none",
          ),
        );
      } else {
        setSelectedMentionRef(intent.itemRef);
        setInspectorMessage(null);
      }
    },
    [
      focusInspectorSection,
      handleTimelineGridContextKeyDown,
      openRowHistory,
      setInspectorMessage,
      setIsInspectorOpen,
      setSelectedMentionRef,
      setSelectedRowId,
      timelineRowForEventTarget,
    ],
  );

  return {
    commands: {
      onCollectionEditorKeyDown,
      onScalarEditorKeyDown,
      onWorkAreaKeyDown,
    },
  };
}
