import type {
  GridCellAnchor,
  GridNavigationIntent,
} from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { type KeyboardEvent as ReactKeyboardEvent, useCallback } from "react";
import type { WorkbookContinuityAnchor } from "../../continuity/workbookContinuityPort";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { mapWorkbookKeyboardCommand } from "../../utils/workbookKeyboard";
import type { TimelineScalarSaveOptions } from "../models/timelineControllerPorts";
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

export function useTimelineKeyboardController({
  clearRowHistory,
  currentTimelineAnchorFor,
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
  readonly rowHistory: {
    readonly recordId: string | null;
    readonly status: string;
  };
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
        mode: "editor",
        rowKind: anchor === null ? "draft" : "committed",
      });
      if (command.preventDefault) event.preventDefault();

      const adapterOwnsRange =
        surface === "grid" &&
        command.kind === "navigate" &&
        command.intent.shiftKey &&
        command.intent.key.startsWith("Arrow");
      if (adapterOwnsRange) return;

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
        if (
          command.intent.key === "Enter" &&
          focusField === "activitySynopsisText" &&
          surface === "grid"
        ) {
          recordTiming("blank_row_commit_accepted", {
            field: "timeline.activity_synopsis_text",
            surface,
          });
        }
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
      if (command.kind === "close-inspector") {
        closeInspectorFromEditor(anchor);
      }
    },
    [
      closeInspectorFromEditor,
      currentTimelineAnchorFor,
      navigateTimelineFocusAnchor,
      queueScalarSave,
      recordTiming,
      restoreTimelineFocusAnchor,
      rowHistory.recordId,
      rowHistory.status,
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
      const command = mapWorkbookKeyboardCommand(event, {
        closeInspector:
          anchor !== null &&
          (selectedRowId !== null ||
            rowHistory.recordId !== null ||
            rowHistory.status !== "idle"),
        mode: "editor",
        rowKind: anchor === null ? "draft" : "committed",
      });
      if (command.preventDefault) event.preventDefault();
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
      if (command.kind === "close-inspector") {
        closeInspectorFromEditor(anchor);
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
      closeInspectorFromEditor,
      currentTimelineAnchorFor,
      navigateTimelineFocusAnchor,
      queueCollectionSave,
      rowHistory.recordId,
      rowHistory.status,
      selectedRowId,
    ],
  );

  const focusInspectorSection = useCallback(
    (section: "evidence" | "history") => {
      window.requestAnimationFrame(() => {
        document
          .querySelector<HTMLElement>(
            dataTestIdSelector(timelineInspectorSectionTestId(section)),
          )
          ?.focus({ preventScroll: true });
      });
    },
    [],
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
      const recordId = row?.recordId ?? null;
      const command = mapWorkbookKeyboardCommand(event, {
        cell:
          fieldKey === undefined
            ? null
            : {
                linkResolveCapability:
                  fieldKey === "timeline.host_refs" ||
                  fieldKey === "timeline.identity_refs",
              },
        inspectorGroups: ["evidence", "history"],
        mode: "grid_navigation",
        committedRowIdentity: recordId,
        previewableEvidenceCount: 0,
        rowKind: recordId === null ? "none" : "committed",
      });
      const isInspectorShortcut =
        command.kind === "open-history" ||
        command.kind === "preview-linked-evidence" ||
        command.kind === "quick-link";
      if (!isInspectorShortcut || row === null || recordId === null) return;

      event.preventDefault();
      event.stopPropagation();
      if (command.kind === "open-history") {
        openRowHistory(recordId);
        focusInspectorSection("history");
        return;
      }
      if (command.kind === "preview-linked-evidence") {
        setSelectedRowId(recordId);
        setIsInspectorOpen(true);
        setInspectorMessage(null);
        focusInspectorSection("evidence");
        return;
      }

      const mention = [
        ...row.collectionValues.hostRefs,
        ...row.collectionValues.identityRefs,
      ].find((item) => item.itemKind !== "resolved_ref");
      setSelectedRowId(recordId);
      setIsInspectorOpen(true);
      if (mention === undefined) {
        setInspectorMessage(
          workbookInspectorMessageFeedback(
            "No unresolved mention is available for quick link.",
            "none",
          ),
        );
      } else {
        setSelectedMentionRef(mention.itemRef);
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
