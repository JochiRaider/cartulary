import type { GridCellAnchor } from "@cartulary/grid-adapter";
import { type ClipboardEvent as ReactClipboardEvent, useCallback } from "react";
import { apiPath } from "../../../services/browserApi";
import { fetchWorkbookJSON, readEnvelope } from "../../../services/workbookApi";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { sameFieldConflictQueueKey } from "../../utils/workbookPendingQueue";
import {
  inputFocusKey,
  type RowValues,
  type SameFieldConflictPayload,
  type TimelineScalarEditorSurface,
  timelineScalarBindingForField,
  timelineScalarBindings,
} from "../models/workbookTimelineModel";
import type { TimelinePasteTargetResolution } from "./useTimelineGridAnchorController";
import type { TimelineScalarSaveOptions } from "./useTimelineMutationCommands";
import type { PendingReplayRuntimeMeta } from "./useTimelinePendingReplayController";
import type {
  TimelineMutableRef,
  TimelinePendingSavesRefs,
} from "./useTimelinePendingSaves";

type TimelineClipboardPasteEnvelope = {
  data: {
    view_schema_id: string;
    rows: unknown[];
    conflicts?: SameFieldConflictPayload[];
  };
};

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
  apiBase,
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
}: {
  readonly apiBase?: string | undefined;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    target:
      | { readonly kind: "input"; readonly focusKey: string }
      | { readonly kind: "row-inspect"; readonly recordId: string }
      | { readonly kind: "scroll-only" },
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly finishSave: (nextState: "Conflict" | "Saved" | "Syncing") => void;
  readonly incidentId: string;
  readonly loadRowsRef: TimelineMutableRef<TimelineLoadRowsForPaste>;
  readonly nextClientTxnId: () => string;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    TimelinePendingSavesRefs<PendingReplayRuntimeMeta>
  >;
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
    clipboardText: string,
  ) => TimelinePasteTargetResolution | null;
  readonly restoreTimelineFocusAnchor: (anchor: GridCellAnchor) => boolean;
  readonly rowInputRefs: TimelineMutableRef<
    Map<string, HTMLInputElement | HTMLTextAreaElement>
  >;
  readonly setActiveConflictKey: (key: string | null) => void;
  readonly setPasteConflictGroup: (group: { keys: string[] } | null) => void;
  readonly setScalarEditorDraftValue: (
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
    value: string,
  ) => void;
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
                const result =
                  await fetchWorkbookJSON<TimelineClipboardPasteEnvelope>(
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
    ],
  );

  return {
    commands: {
      handlePaste,
    },
  };
}
