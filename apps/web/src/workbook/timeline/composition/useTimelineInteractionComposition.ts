import type {
  GridCellAnchor,
  GridCellPasteIntent,
} from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useMemo } from "react";
import type { WorkbookContinuityAnchor } from "../../continuity/workbookContinuityPort";
import type { WorkbookSurfaceLayoutOwner } from "../../layout/useWorkbookLayoutFacade";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationFeedback } from "../../runtime/workbookLifecycleModel";
import { useTimelineBulkTagController } from "../bulk/useTimelineBulkTagController";
import { useTimelineClearController } from "../bulk/useTimelineClearController";
import { useTimelineFillController } from "../bulk/useTimelineFillController";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineClipboardPasteController } from "../hooks/useTimelineClipboardPasteController";
import { useTimelineKeyboardController } from "../hooks/useTimelineKeyboardController";
import type { useTimelineMutationCommands } from "../hooks/useTimelineMutationCommands";
import { decodeTimelineClipboardInput } from "../models/timelineClipboardPastePlan";
import { timelineScalarBindingForField } from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";

const timelineContract = requireViewContract(timelineViewSchemaId);

type BulkInput = Parameters<typeof useTimelineBulkTagController>[0];
type ClipboardInput = Parameters<typeof useTimelineClipboardPasteController>[0];
type KeyboardInput = Parameters<typeof useTimelineKeyboardController>[0];
type MutationCommandOutput = ReturnType<
  typeof useTimelineMutationCommands
>["commands"];

type TimelineInteractionCompositionInput = {
  readonly foundation: {
    readonly editorDraftRegistry: TimelineEditorDraftRegistry;
    readonly bulkTagPort: BulkInput["port"];
    readonly bulkTagReadiness: BulkInput["readiness"];
    readonly clipboardPastePort: ClipboardInput["clipboardPaste"];
    readonly pendingSavesRefs: ClipboardInput["pendingSavesRefs"];
    readonly recordTiming: KeyboardInput["recordTiming"];
    readonly rows: readonly WorkbookRow[];
    readonly rowsRef: BulkInput["rowsRef"];
    readonly setOperationError: (
      feedback: WorkbookOperationFeedback | null,
    ) => void;
    readonly setSelectedMentionRef: KeyboardInput["setSelectedMentionRef"];
  };
  readonly grid: {
    readonly currentTimelineAnchorFor: KeyboardInput["currentTimelineAnchorFor"];
    readonly focusDraftRow: () => void;
    readonly prepareTimelineCollectionNavigation: KeyboardInput["prepareTimelineCollectionNavigation"];
    readonly navigateTimelineDraftFocus?: KeyboardInput["navigateTimelineDraftFocus"];
    readonly navigateTimelineFocusAnchor: KeyboardInput["navigateTimelineFocusAnchor"];
    readonly resolveTimelinePasteTargetResolution: ClipboardInput["resolveTimelinePasteTargetResolution"];
    readonly restoreTimelineFocusAnchor: (
      anchor: GridCellAnchor | WorkbookContinuityAnchor,
    ) => Promise<boolean>;
    readonly timelineAnchorColumnsRef: {
      readonly current: readonly { readonly fieldKey: string }[];
    };
    readonly updateTimelineSurfaceFocusAnchor: (
      recordId: string | null,
      fieldKey: string,
    ) => void;
    readonly workbookFocusAnchorRef: KeyboardInput["workbookFocusAnchorRef"];
  };
  readonly inspector: {
    readonly clearRowHistory: KeyboardInput["clearRowHistory"];
    readonly elementRegistry: KeyboardInput["elementRegistry"];
    readonly publishFeedback: KeyboardInput["setInspectorMessage"];
    readonly rowHistory: KeyboardInput["rowHistory"];
    readonly selectedRowId: KeyboardInput["selectedRowId"];
    readonly selectRow: KeyboardInput["setSelectedRowId"];
    readonly setOpen: KeyboardInput["setIsInspectorOpen"];
  };
  readonly interactionMode: WorkbookSurfaceLayoutOwner["snapshot"]["interactionMode"];
  readonly loadAccessLost: boolean;
  readonly mutation: {
    readonly activateConflict: (key: string | null) => void;
    readonly commitScalarGridEdit: MutationCommandOutput["commitScalarGridEdit"];
    readonly mutationCommands: TimelineWorkbookSurfaceRuntime["mutationCommands"];
    readonly queueCollectionSave: KeyboardInput["queueCollectionSave"];
    readonly queueScalarSave: KeyboardInput["queueScalarSave"];
  };
  readonly queryState: WorkbookQueryState;
  readonly role: TimelineWorkbookSurfaceRuntime["incident"]["currentRole"];
  readonly workflow: {
    readonly handleTimelineGridContextKeyDown: KeyboardInput["handleTimelineGridContextKeyDown"];
    readonly openRowHistory: KeyboardInput["openRowHistory"];
    readonly timelineRowForEventTarget: KeyboardInput["timelineRowForEventTarget"];
  };
};

function gridCoreRecordId(
  anchor: GridCellAnchor | GridCellPasteIntent["target"],
): string | null {
  return anchor.rowIdentity.kind === "core_record"
    ? anchor.rowIdentity.recordId
    : null;
}

export function useTimelineInteractionComposition({
  foundation,
  grid,
  inspector,
  interactionMode,
  loadAccessLost,
  mutation,
  queryState,
  role,
  workflow,
}: TimelineInteractionCompositionInput) {
  const canEdit =
    interactionMode.kind === "editable" &&
    (role === "editor" || role === "reviewer" || role === "admin");
  const canBulkTag = canEdit;
  const bulk = useTimelineBulkTagController({
    context: {
      authorized: canBulkTag && !loadAccessLost,
      capabilityAvailable: timelineBulkTagCapabilityAvailable,
    },
    port: foundation.bulkTagPort,
    readiness: foundation.bulkTagReadiness,
    precedingSaves: () => foundation.pendingSavesRefs.saveQueueRef.current,
    rows: foundation.rows,
    rowsRef: foundation.rowsRef,
  });
  const handleBlur = useCallback(
    (
      rowKey: string,
      focusField: Parameters<KeyboardInput["queueScalarSave"]>[1],
      surface: Parameters<KeyboardInput["queueScalarSave"]>[2]["surface"],
      currentValue: string,
    ) => {
      mutation.queueScalarSave(
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
    [mutation.queueScalarSave],
  );
  const { commands: keyboard } = useTimelineKeyboardController({
    editorDraftRegistry: foundation.editorDraftRegistry,
    prepareTimelineCollectionNavigation:
      grid.prepareTimelineCollectionNavigation,
    navigateTimelineDraftFocus: grid.navigateTimelineDraftFocus,
    clearRowHistory: inspector.clearRowHistory,
    currentTimelineAnchorFor: grid.currentTimelineAnchorFor,
    elementRegistry: inspector.elementRegistry,
    handleTimelineGridContextKeyDown: workflow.handleTimelineGridContextKeyDown,
    navigateTimelineFocusAnchor: grid.navigateTimelineFocusAnchor,
    openRowHistory: workflow.openRowHistory,
    queueCollectionSave: mutation.queueCollectionSave,
    queueScalarSave: mutation.queueScalarSave,
    recordTiming: foundation.recordTiming,
    restoreTimelineFocusAnchor: grid.restoreTimelineFocusAnchor,
    rowHistory: inspector.rowHistory,
    selectedRowId: inspector.selectedRowId,
    setInspectorMessage: inspector.publishFeedback,
    setIsInspectorOpen: inspector.setOpen,
    setSelectedMentionRef: foundation.setSelectedMentionRef,
    setSelectedRowId: inspector.selectRow,
    timelineRowForEventTarget: workflow.timelineRowForEventTarget,
    workbookFocusAnchorRef: grid.workbookFocusAnchorRef,
  });
  const setPasteError = useCallback(
    (message: string | null) =>
      foundation.setOperationError(
        message === null ? null : { family: "paste", message },
      ),
    [foundation.setOperationError],
  );
  const setFillError = useCallback(
    (message: string | null) =>
      foundation.setOperationError(
        message === null ? null : { family: "fill", message },
      ),
    [foundation.setOperationError],
  );
  const setClearError = useCallback(
    (message: string | null) =>
      foundation.setOperationError(
        message === null ? null : { family: "clear", message },
      ),
    [foundation.setOperationError],
  );
  const clipboard = useTimelineClipboardPasteController({
    canCreateRows: canEdit,
    clipboardPaste: foundation.clipboardPastePort,
    editable: canEdit,
    grouped: queryState.groupBy !== null,
    pendingSavesRefs: foundation.pendingSavesRefs,
    queueScalarSave: mutation.queueScalarSave,
    resolveTimelinePasteTargetResolution:
      grid.resolveTimelinePasteTargetResolution,
    setError: setPasteError,
  }).commands;
  const handleTimelineGridPaste = useCallback(
    (intent: Parameters<typeof clipboard.handleGridPaste>[0]) => {
      if (intent.input.kind === "scalar") {
        const binding = timelineScalarBindingForField(intent.target.fieldKey);
        if (binding === null) return false;
        return mutation
          .commitScalarGridEdit(
            gridCoreRecordId(intent.target) ?? "",
            binding.key,
            intent.input.value,
          )
          .then((outcome) => {
            if (outcome.kind !== "accepted") {
              setPasteError(outcome.message ?? "Save failed.");
            }
            return outcome.kind === "accepted";
          });
      }
      return clipboard.handleGridPaste(intent);
    },
    [clipboard.handleGridPaste, setPasteError, mutation.commitScalarGridEdit],
  );
  const clipboardPaste = useMemo(
    () => ({
      decode: decodeTimelineClipboardInput,
      onError: setPasteError,
      onPaste: handleTimelineGridPaste,
    }),
    [handleTimelineGridPaste, setPasteError],
  );
  const fill = useTimelineFillController({
    contract: timelineContract,
    precedingSaves: () => foundation.pendingSavesRefs.saveQueueRef.current,
    getVisibleFieldKeys: () =>
      new Set(
        grid.timelineAnchorColumnsRef.current.map((column) => column.fieldKey),
      ),
    groupBy: queryState.groupBy,
    interactionMode,
    port: mutation.mutationCommands.fill,
    rowsRef: foundation.rowsRef,
    setError: setFillError,
  }).commands;
  const clear = useTimelineClearController({
    authorized: canEdit && !loadAccessLost,
    contract: timelineContract,
    getVisibleFieldKeys: () =>
      new Set(
        grid.timelineAnchorColumnsRef.current.map((column) => column.fieldKey),
      ),
    hasUnsubmittedDraft: (row, fieldKey) => {
      const binding = timelineScalarBindingForField(fieldKey);
      if (!binding) return true;
      const pending = foundation.pendingSavesRefs;
      const admitted = pending.pendingQueueRef.current.model
        .snapshot()
        .units.flatMap((unit) => {
          const revisions = pending.replayContextByUnitId.get(
            unit.id,
          )?.draftRevisions;
          return revisions ? [revisions] : [];
        });
      return foundation.editorDraftRegistry.hasUnsubmittedScalarDraft(
        row.key,
        binding.key,
        admitted,
      );
    },
    port: mutation.mutationCommands.clear,
    precedingSaves: () => foundation.pendingSavesRefs.saveQueueRef.current,
    rowsRef: foundation.rowsRef,
    setError: setClearError,
  });
  const handleCreateBlankDraftRow = useCallback(
    (row: WorkbookRow) => {
      const currentRows = foundation.rowsRef.current;
      const activeRow =
        currentRows.find((candidate) => candidate.key === row.key) ?? row;
      mutation.queueScalarSave(activeRow.key, "activitySynopsisText", {
        allowZeroFieldCreate: true,
        continueOnFreshDraft: true,
        preserveInputFocus: false,
        surface: "grid",
      });
    },
    [foundation.rowsRef, mutation.queueScalarSave],
  );
  return {
    commands: {
      editor: {
        activateConflictCell: mutation.activateConflict,
        commitScalarGridEdit: mutation.commitScalarGridEdit,
        handleBlur,
        handleCollectionKeyDown: keyboard.onCollectionEditorKeyDown,
        handleKeyDown: keyboard.onScalarEditorKeyDown,
        handlePaste: clipboard.handlePaste,
        queueCollectionSave: mutation.queueCollectionSave,
      },
      grid: {
        clipboardPaste,
        focusDraftRow: grid.focusDraftRow,
        handleCreateBlankDraftRow,
        handleFillCells: fill.onFillCells,
        handleClearCells: clear,
        handleWorkAreaKeyDown: keyboard.onWorkAreaKeyDown,
      },
    },
    ports: { bulkTag: bulk.controls },
    snapshot: {
      bulk: bulk.snapshot,
      editor: {},
    },
  };
}

const timelineBulkTagCapabilityAvailable =
  timelineContract.fieldMap["timeline.tags"]?.writeKind === "action_payload";
