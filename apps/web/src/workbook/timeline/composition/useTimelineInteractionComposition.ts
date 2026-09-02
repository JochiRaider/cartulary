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
import { workbookClipboardPasteContract } from "../../utils/workbookClipboard";
import { useTimelineBulkTagController } from "../bulk/useTimelineBulkTagController";
import { useTimelineFillController } from "../bulk/useTimelineFillController";
import { useTimelineClipboardPasteController } from "../hooks/useTimelineClipboardPasteController";
import { useTimelineKeyboardController } from "../hooks/useTimelineKeyboardController";
import type { useTimelineMutationCommands } from "../hooks/useTimelineMutationCommands";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";
import {
  timelineScalarBindingForField,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

const timelineContract = requireViewContract(timelineViewSchemaId);

type BulkInput = Parameters<typeof useTimelineBulkTagController>[0];
type ClipboardInput = Parameters<typeof useTimelineClipboardPasteController>[0];
type FillInput = Parameters<typeof useTimelineFillController>[0];
type KeyboardInput = Parameters<typeof useTimelineKeyboardController>[0];
type MutationCommandOutput = ReturnType<
  typeof useTimelineMutationCommands
>["commands"];

type TimelineInteractionCompositionInput = {
  readonly foundation: {
    readonly activateCollectionInput: (focusKey: string) => void;
    readonly activeCollectionInputKey: string | null;
    readonly clipboardPastePort: ClipboardInput["timelineClipboardPaste"];
    readonly deactivateCollectionInput: (focusKey: string) => void;
    readonly editorDraftRegistry: ClipboardInput["editorDraftRegistry"];
    readonly pendingSavesRefs: ClipboardInput["pendingSavesRefs"];
    readonly recordTiming: KeyboardInput["recordTiming"];
    readonly rows: readonly WorkbookRow[];
    readonly rowsRef: BulkInput["rowsRef"];
    readonly setRefreshError: (message: string | null) => void;
    readonly setSelectedMentionRef: KeyboardInput["setSelectedMentionRef"];
  };
  readonly grid: {
    readonly beginViewportContinuity: ClipboardInput["beginViewportContinuity"];
    readonly clearViewportContinuity: ClipboardInput["clearViewportContinuity"];
    readonly currentTimelineAnchorFor: KeyboardInput["currentTimelineAnchorFor"];
    readonly focusDraftRow: () => void;
    readonly navigateTimelineFocusAnchor: KeyboardInput["navigateTimelineFocusAnchor"];
    readonly resolveTimelinePasteTargetResolution: ClipboardInput["resolveTimelinePasteTargetResolution"];
    readonly restoreTimelineFocusAnchor: (
      anchor: GridCellAnchor | WorkbookContinuityAnchor,
    ) => boolean;
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
  readonly mutation: {
    readonly applyClipboardResponseRows: ClipboardInput["applyResponseRows"];
    readonly beginSave: ClipboardInput["beginSave"];
    readonly commitScalarGridEdit: MutationCommandOutput["commitScalarGridEdit"];
    readonly enqueueSaveWork: FillInput["enqueueSaveWork"];
    readonly finishSave: ClipboardInput["finishSave"];
    readonly loadRows: ClipboardInput["loadRows"];
    readonly mutationCommands: TimelineWorkbookSurfaceRuntime["mutationCommands"];
    readonly nextClientTxnId: ClipboardInput["nextClientTxnId"];
    readonly queueCollectionSave: KeyboardInput["queueCollectionSave"];
    readonly queueScalarSave: KeyboardInput["queueScalarSave"];
    readonly registerSameFieldConflict: ClipboardInput["registerSameFieldConflict"];
    readonly resolvePendingSocketTxn: ClipboardInput["resolvePendingSocketTxn"];
    readonly setActiveConflictKey: ClipboardInput["setActiveConflictKey"];
    readonly setPasteConflictGroup: ClipboardInput["setPasteConflictGroup"];
    readonly trackPendingSocketTxn: ClipboardInput["trackPendingSocketTxn"];
    readonly waitForCommittedRecordIdle: ClipboardInput["waitForCommittedRecordIdle"];
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
  mutation,
  queryState,
  role,
  workflow,
}: TimelineInteractionCompositionInput) {
  const canBulkTag =
    interactionMode.kind === "editable" &&
    (role === "editor" || role === "reviewer" || role === "admin");
  const refreshRowsForBulkTag = useCallback(
    () => mutation.loadRows({ showLoading: false }),
    [mutation.loadRows],
  );
  const bulk = useTimelineBulkTagController({
    canAssign: canBulkTag,
    port: mutation.mutationCommands.bulk,
    refreshRows: refreshRowsForBulkTag,
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
  const clipboard = useTimelineClipboardPasteController({
    applyResponseRows: mutation.applyClipboardResponseRows,
    beginSave: mutation.beginSave,
    beginViewportContinuity: grid.beginViewportContinuity,
    clearViewportContinuity: grid.clearViewportContinuity,
    editorDraftRegistry: foundation.editorDraftRegistry,
    finishSave: mutation.finishSave,
    loadRows: mutation.loadRows,
    nextClientTxnId: mutation.nextClientTxnId,
    pendingSavesRefs: foundation.pendingSavesRefs,
    queueScalarSave: mutation.queueScalarSave,
    registerSameFieldConflict: mutation.registerSameFieldConflict,
    resolvePendingSocketTxn: mutation.resolvePendingSocketTxn,
    resolveTimelinePasteTargetResolution:
      grid.resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor: grid.restoreTimelineFocusAnchor,
    setActiveConflictKey: mutation.setActiveConflictKey,
    setPasteConflictGroup: mutation.setPasteConflictGroup,
    timelineClipboardPaste: foundation.clipboardPastePort,
    trackPendingSocketTxn: mutation.trackPendingSocketTxn,
    waitForCommittedRecordIdle: mutation.waitForCommittedRecordIdle,
  }).commands;
  const handleTimelineGridPaste = useCallback(
    (intent: Parameters<typeof clipboard.handleGridPaste>[0]) => {
      if (intent.input.kind === "scalar") {
        const binding = timelineScalarBindingForField(intent.target.fieldKey);
        if (binding === null) return;
        void mutation
          .commitScalarGridEdit(
            gridCoreRecordId(intent.target) ?? "",
            binding.key,
            intent.input.value,
          )
          .then((outcome) => {
            if (outcome.kind !== "accepted") {
              foundation.setRefreshError(outcome.message ?? "Save failed.");
            }
          });
        return;
      }
      clipboard.handleGridPaste(intent);
    },
    [clipboard.handleGridPaste, foundation.setRefreshError, mutation],
  );
  const clipboardPaste = useMemo(
    () => workbookClipboardPasteContract(handleTimelineGridPaste),
    [handleTimelineGridPaste],
  );
  const fill = useTimelineFillController({
    beginSave: mutation.beginSave,
    beginViewportContinuity: grid.beginViewportContinuity,
    clearViewportContinuity: grid.clearViewportContinuity,
    contract: timelineContract,
    enqueueSaveWork: mutation.enqueueSaveWork,
    finishSave: mutation.finishSave,
    getVisibleFieldKeys: () =>
      new Set(
        grid.timelineAnchorColumnsRef.current.map((column) => column.fieldKey),
      ),
    groupBy: queryState.groupBy,
    interactionMode,
    loadRows: mutation.loadRows,
    port: mutation.mutationCommands.bulk,
    resolvePendingSocketTxn: mutation.resolvePendingSocketTxn,
    restoreFocusAnchor: grid.restoreTimelineFocusAnchor,
    rowsRef: foundation.rowsRef,
    setError: foundation.setRefreshError,
    trackPendingSocketTxn: mutation.trackPendingSocketTxn,
  }).commands;
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
  const collectionKeyboardCommitRef =
    foundation.pendingSavesRefs.collectionKeyboardCommitRef;
  const handleCollectionInputChange = useCallback(
    (focusKey: string, value: string) => {
      if (collectionKeyboardCommitRef.current.get(focusKey) !== value) {
        collectionKeyboardCommitRef.current.delete(focusKey);
      }
    },
    [collectionKeyboardCommitRef],
  );
  return {
    commands: {
      bulk: bulk.commands,
      editor: {
        activateConflictCell: mutation.setActiveConflictKey,
        activateCollectionInput: foundation.activateCollectionInput,
        commitScalarGridEdit: mutation.commitScalarGridEdit,
        deactivateCollectionInput: foundation.deactivateCollectionInput,
        handleBlur,
        handleCollectionInputChange,
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
        handleWorkAreaKeyDown: keyboard.onWorkAreaKeyDown,
      },
    },
    ports: {},
    snapshot: {
      bulk: bulk.snapshot,
      editor: {
        activeCollectionInputKey: foundation.activeCollectionInputKey,
      },
    },
  };
}
