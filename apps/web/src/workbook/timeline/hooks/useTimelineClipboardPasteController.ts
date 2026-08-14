import type {
  GridCellAnchor,
  GridCellPasteIntent,
  GridClipboardInput,
} from "@cartulary/grid-adapter";
import { type ClipboardEvent as ReactClipboardEvent, useCallback } from "react";
import type { WorkbookPendingSavesRefs } from "../../runtime/workbookPendingReplayRuntime";
import { decodeWorkbookClipboardInput } from "../../utils/workbookClipboard";
import { sameFieldConflictQueueKey } from "../../utils/workbookPendingQueue";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type {
  PendingReplayRuntimeMeta,
  TimelinePasteTargetResolution,
  TimelineScalarSaveOptions,
} from "../models/timelineControllerPorts";
import {
  inputFocusKey,
  type RowValues,
  type SameFieldConflictPayload,
  type TimelineApiRow,
  type TimelineScalarEditorSurface,
  timelineScalarBindingForField,
  timelineScalarBindings,
} from "../models/workbookTimelineModel";
import type {
  TimelineClipboardPastePort,
  TimelineClipboardPasteTarget,
} from "../ports/TimelineClipboardPastePort";

type TimelineLoadRowsForPaste = (options: {
  readonly showLoading: boolean;
  readonly viewportContinuityToken?: number;
}) => Promise<void>;

type TimelineQueueScalarSave = (
  rowKey: string,
  focusField: keyof RowValues,
  options: TimelineScalarSaveOptions,
  currentValue?: string,
) => void;

type TimelineCommittedRecordIdle = {
  readonly rowVersion: number;
};

export function useTimelineClipboardPasteController({
  applyResponseRows,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  editorDraftRegistry,
  finishSave,
  loadRows,
  nextClientTxnId,
  pendingSavesRefs,
  queueScalarSave,
  registerSameFieldConflict,
  resolvePendingSocketTxn,
  resolveTimelinePasteTargetResolution,
  restoreTimelineFocusAnchor,
  setActiveConflictKey,
  setPasteConflictGroup,
  timelineClipboardPaste,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly applyResponseRows: (rows: readonly TimelineApiRow[]) => void;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    target:
      | { readonly kind: "input"; readonly focusKey: string }
      | { readonly kind: "row-inspect"; readonly recordId: string }
      | { readonly kind: "scroll-only" },
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly finishSave: (nextState: "Conflict" | "Saved" | "Syncing") => void;
  readonly loadRows: TimelineLoadRowsForPaste;
  readonly nextClientTxnId: () => string;
  readonly pendingSavesRefs: WorkbookPendingSavesRefs<PendingReplayRuntimeMeta>;
  readonly queueScalarSave: TimelineQueueScalarSave;
  readonly registerSameFieldConflict: (
    conflict: SameFieldConflictPayload,
    focusKey: string,
    surface: TimelineScalarEditorSurface,
  ) => void;
  readonly resolvePendingSocketTxn: (
    clientTxnId: string | null | undefined,
  ) => boolean;
  readonly resolveTimelinePasteTargetResolution: (
    rowKey: string,
    fieldKey: string,
    input: GridClipboardInput,
  ) => TimelinePasteTargetResolution | null;
  readonly restoreTimelineFocusAnchor: (anchor: GridCellAnchor) => boolean;
  readonly setActiveConflictKey: (key: string | null) => void;
  readonly setPasteConflictGroup: (group: { keys: string[] } | null) => void;
  readonly timelineClipboardPaste: TimelineClipboardPastePort;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<TimelineCommittedRecordIdle | null>;
}) {
  const handlePaste = useCallback(
    (
      event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
      prepared?:
        | (TimelinePasteTargetResolution & {
            readonly input: GridClipboardInput;
          })
        | undefined,
    ) => {
      const clipboardText =
        prepared?.input.rawText ??
        event.clipboardData?.getData("text/plain") ??
        "";
      const clipboardInput =
        prepared?.input ?? decodeWorkbookClipboardInput(clipboardText);
      const binding = timelineScalarBindings.find(
        (candidate) => candidate.key === focusField,
      );
      const fieldKey = binding?.fieldKey ?? focusField;
      if (surface === "grid" && binding !== undefined) {
        const pasteTargetResolution =
          prepared ??
          resolveTimelinePasteTargetResolution(
            rowKey,
            fieldKey,
            clipboardInput,
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
          pendingSavesRefs.saveQueueRef.current =
            pendingSavesRefs.saveQueueRef.current
              .catch(() => undefined)
              .then(async () => {
                const rowTargetPayload: TimelineClipboardPasteTarget[] = [];
                for (const target of targetResolution.rowTargets) {
                  if (target.kind === "create") {
                    rowTargetPayload.push({ kind: "create" });
                    continue;
                  }
                  const idleRecord = await waitForCommittedRecordIdle(
                    target.rowIdentity.recordId,
                  );
                  if (idleRecord === null) {
                    clearViewportContinuity(viewportContinuityToken);
                    finishSave("Conflict");
                    return;
                  }
                  rowTargetPayload.push({
                    kind: "record",
                    recordId: target.rowIdentity.recordId,
                    baseRowVersion: idleRecord.rowVersion,
                  });
                }
                trackPendingSocketTxn(clientTxnId);
                const result = await timelineClipboardPaste.paste({
                  clientTxnId,
                  clipboardText,
                  format:
                    clipboardInput.kind === "table"
                      ? clipboardInput.format
                      : "csv",
                  startFieldKey: fieldKey,
                  columns: targetResolution.columns,
                  targets: rowTargetPayload,
                });
                resolvePendingSocketTxn(clientTxnId);
                if (result.kind === "rejected") {
                  clearViewportContinuity(viewportContinuityToken);
                  finishSave("Conflict");
                  return;
                }
                const pasteConflictKeys: string[] = [];
                for (const conflict of result.value.conflicts) {
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
                applyResponseRows(result.value.rows);
                await loadRows({
                  showLoading: false,
                  viewportContinuityToken,
                });
                if (anchor !== null) {
                  restoreTimelineFocusAnchor(anchor);
                }
                finishSave(
                  result.value.conflicts.length > 0 ? "Conflict" : "Saved",
                );
              });
          return;
        }
      }
      window.setTimeout(() => {
        const editor = editorDraftRegistry.inputElementForFocusKey(
          inputFocusKey(rowKey, focusField, surface),
        );
        if (editor) {
          editorDraftRegistry.setDraft(
            { field: focusField, rowKey, surface },
            editor.value,
          );
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
      applyResponseRows,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      editorDraftRegistry,
      finishSave,
      loadRows,
      nextClientTxnId,
      pendingSavesRefs,
      queueScalarSave,
      registerSameFieldConflict,
      resolvePendingSocketTxn,
      resolveTimelinePasteTargetResolution,
      restoreTimelineFocusAnchor,
      setActiveConflictKey,
      setPasteConflictGroup,
      timelineClipboardPaste,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  const handleGridPaste = useCallback(
    (intent: GridCellPasteIntent) => {
      const clipboardText = intent.input.rawText;
      const binding = timelineScalarBindingForField(intent.target.fieldKey);
      if (binding === null) return;
      handlePaste(
        {
          clipboardData: {
            getData: () => clipboardText,
          },
          preventDefault: () => undefined,
        } as unknown as ReactClipboardEvent<HTMLInputElement>,
        intent.target.rowIdentity.kind === "core_record"
          ? intent.target.rowIdentity.recordId
          : "",
        binding.key,
        "grid",
        {
          anchor: intent.target,
          input: intent.input,
          targetResolution: intent.targetResolution,
        },
      );
    },
    [handlePaste],
  );

  return {
    commands: {
      handleGridPaste,
      handlePaste,
    },
  };
}
