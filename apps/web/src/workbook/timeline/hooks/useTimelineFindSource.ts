import type {
  GridCellAnchor,
  GridCellNavigationOptions,
  GridEditCommitOutcome,
} from "@cartulary/grid-adapter";
import { useCallback, useRef } from "react";
import {
  useWorkbookFind,
  type WorkbookFindFocusLoan,
} from "../../find/useWorkbookFind";
import type { WorkbookQueryBrowser } from "../../query/WorkbookQueryBrowser";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type {
  TimelineQueueCollectionSave,
  TimelineQueueScalarSave,
} from "../models/timelineControllerPorts";
import { timelineCollectionBindings } from "../models/timelineFieldRegistry";
import { timelineFindText } from "../models/timelineFindText";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";

/** Timeline supplies authority, accepted membership, text and source editor settlement. */
export function useTimelineFindSource(input: {
  runtime: {
    incident: Pick<
      TimelineWorkbookSurfaceRuntime["incident"],
      "id" | "continuityResetKey" | "sheetRef" | "currentRole"
    >;
    query: Pick<TimelineWorkbookSurfaceRuntime["query"], "state">;
    layout: {
      snapshot: Pick<
        TimelineWorkbookSurfaceRuntime["layout"]["snapshot"],
        "state"
      >;
    };
    entities: Pick<TimelineWorkbookSurfaceRuntime["entities"], "index">;
    collaborationProjection: Pick<
      TimelineWorkbookSurfaceRuntime["collaborationProjection"],
      "subscribe" | "getReadAuthorization"
    >;
  };
  browser: WorkbookQueryBrowser | undefined;
  rows: readonly WorkbookRow[];
  registry: TimelineEditorDraftRegistry;
  accessLost: boolean;
  stale: boolean;
  interruptViewportContinuity: () => void;
  queueScalarSave: TimelineQueueScalarSave;
  queueCollectionSave: TimelineQueueCollectionSave;
}) {
  const latest = useRef(input);
  latest.current = input;
  const sourceSettlement = useRef<{
    key: string;
    promise: Promise<GridEditCommitOutcome>;
  } | null>(null);
  const captureFocus = useCallback(
    (target?: EventTarget | null): WorkbookFindFocusLoan | null => {
      if (
        target instanceof HTMLElement &&
        target.closest("[data-inspector-editor-field]")
      ) {
        return {
          ownsFocus: (focused) => target === focused && target.isConnected,
          restore: () => {
            if (!target.isConnected) return false;
            target.focus({ preventScroll: true });
            return document.activeElement === target;
          },
          settle: async () => "accepted",
        };
      }
      const editor = latest.current.registry.activeInput(target);
      if (!editor) return null;
      return {
        ownsFocus: (focused) =>
          latest.current.registry.inputElementForFocusKey(editor.focusKey) ===
          focused,
        restore: () => {
          const element = latest.current.registry.inputElementForFocusKey(
            editor.focusKey,
          );
          if (!element) return false;
          element.focus();
          return document.activeElement === element;
        },
        settle: async (options: GridCellNavigationOptions) => {
          const current = latest.current;
          const row =
            editor &&
            current.rows.find(
              (row) =>
                row.key === current.registry.resolveRowKey(editor.rowKey),
            );
          // Recordless authoring stays retained; Find must never be its creation trigger.
          if (editor && row?.recordId) {
            const collection = timelineCollectionBindings.find(
              (binding) => binding.draftKey === editor.field,
            );
            if (collection) {
              const element = current.registry.inputElementForFocusKey(
                editor.focusKey,
              );
              const key = `${editor.focusKey}:${current.registry.revisionForFocusKey(editor.focusKey)}`;
              if (sourceSettlement.current?.key !== key) {
                const promise = new Promise<GridEditCommitOutcome>(
                  (resolve) => {
                    if (collection) {
                      const value =
                        current.registry.draftValue(editor) ??
                        element?.value ??
                        "";
                      if (value === "") {
                        resolve({ kind: "accepted" });
                        return;
                      }
                      current.queueCollectionSave(
                        editor.rowKey,
                        collection.fieldKey,
                        collection.draftKey,
                        value,
                        editor.surface,
                        resolve,
                      );
                    }
                  },
                ).finally(() => {
                  if (sourceSettlement.current?.promise === promise)
                    sourceSettlement.current = null;
                });
                sourceSettlement.current = { key, promise };
              }
              const outcome = await new Promise<GridEditCommitOutcome | null>(
                (resolve) => {
                  const abort = () => resolve(null);
                  if (options.signal?.aborted) return abort();
                  options.signal?.addEventListener("abort", abort, {
                    once: true,
                  });
                  void sourceSettlement.current?.promise.then((result) => {
                    options.signal?.removeEventListener("abort", abort);
                    resolve(result);
                  });
                },
              );
              if (
                !outcome ||
                options.signal?.aborted ||
                options.isCurrent?.() === false
              )
                return "cancelled";
              if (outcome.kind !== "accepted") {
                current.registry
                  .inputElementForFocusKey(editor.focusKey)
                  ?.focus();
                return "rejected";
              }
            }
          }
          if (options.signal?.aborted || options.isCurrent?.() === false)
            return "cancelled";
          return "accepted";
        },
      };
    },
    [],
  );
  const readText = useCallback(
    (anchor: GridCellAnchor) => {
      if (anchor.rowIdentity.kind !== "core_record") return [];
      const id = anchor.rowIdentity.recordId;
      const row = input.rows.find((row) => row.recordId === id);
      return row
        ? timelineFindText(row, anchor.fieldKey, input.runtime.entities.index)
        : [];
    },
    [input.rows, input.runtime.entities.index],
  );
  return useWorkbookFind({
    lifetimeKey: `${input.runtime.incident.id}:timeline`,
    configurationKey: JSON.stringify([
      input.runtime.incident.continuityResetKey,
      input.runtime.incident.sheetRef,
      input.runtime.query.state,
      input.runtime.layout.snapshot.state,
    ]),
    browser: input.browser,
    authorization: input.runtime.collaborationProjection,
    readable: !input.accessLost && !!input.runtime.incident.currentRole,
    stale: input.stale,
    readText,
    captureFocus,
    beforeFocus: input.interruptViewportContinuity,
  });
}
