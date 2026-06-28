import { dataTestIdSelector } from "@cartulary/ui-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
} from "react";
import { apiPath } from "../../../services/browserApi";
import { fetchJSON, readEnvelope } from "../../../services/workbookApi";
import { sameFieldConflictQueueKey } from "../../utils/workbookPendingQueue";
import { parseSameFieldConflict } from "../models/timelineConflictModel";
import {
  type FocusFieldKey,
  inputFocusKey,
  type LocalConflictState,
  type PasteConflictGroupState,
  type SameFieldConflictPayload,
  type TimelineConflictResolution,
  type TimelineScalarEditorSurface,
  timelineScalarBindingForField,
  timelineScalarEditorSurfaces,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import {
  buildTimelineConflictResolutionPayload,
  type TimelineMutationEnvelope,
} from "../services/timelineMutationRequests";
import type { PendingReplayRuntimeMeta } from "./useTimelinePendingReplayController";
import type {
  TimelineMutableRef,
  TimelinePendingQueueRuntime,
  TimelinePendingSavesRefs,
} from "./useTimelinePendingSaves";

type TimelineSaveState = "Syncing" | "Saved" | "Conflict";

function saveStateConflictAnchorFromPayload(
  conflict: SameFieldConflictPayload,
) {
  return {
    record_id: conflict.record_id,
    field_key: conflict.field_key,
    base_row_version: conflict.base_row_version,
    current_row_version: conflict.current_row_version,
  };
}

export function useTimelineConflictResolverCoordinator({
  activeConflictKey,
  apiBase,
  applyRowMutation,
  beginViewportContinuity,
  conflictQueue,
  nextClientTxnId,
  pasteConflictGroup,
  pendingSavesRefsRef,
  publishSaveStatePresentation,
  resolveInputElement,
  rowsRef,
  scalarDraftValuesRef,
  schedulePendingReplayRef,
  setActiveConflictKey,
  setConflictQueueState,
  setPasteConflictGroup,
  setRows,
  setSaveState,
  setSaveStateSecondaryMessage,
}: {
  readonly activeConflictKey: string | null;
  readonly apiBase?: string | undefined;
  readonly applyRowMutation: (
    rowKey: string,
    envelope: TimelineMutationEnvelope,
    options?: {
      readonly viewportContinuityToken?: number;
    },
  ) => WorkbookRow;
  readonly beginViewportContinuity: (target: {
    readonly kind: "input";
    readonly focusKey: string;
  }) => number;
  readonly conflictQueue: Record<string, LocalConflictState>;
  readonly nextClientTxnId: () => string;
  readonly pasteConflictGroup: PasteConflictGroupState | null;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    TimelinePendingSavesRefs<PendingReplayRuntimeMeta>
  >;
  readonly publishSaveStatePresentation: (
    pending: TimelinePendingQueueRuntime<PendingReplayRuntimeMeta>,
    conflicts?: Record<string, LocalConflictState>,
  ) => unknown;
  readonly resolveInputElement: (focusKey: string) => HTMLElement | null;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly scalarDraftValuesRef: TimelineMutableRef<Map<string, string>>;
  readonly schedulePendingReplayRef: TimelineMutableRef<() => void>;
  readonly setActiveConflictKey: Dispatch<SetStateAction<string | null>>;
  readonly setConflictQueueState: (
    updater: (
      current: Record<string, LocalConflictState>,
    ) => Record<string, LocalConflictState>,
  ) => void;
  readonly setPasteConflictGroup: Dispatch<
    SetStateAction<PasteConflictGroupState | null>
  >;
  readonly setRows: Dispatch<SetStateAction<WorkbookRow[]>>;
  readonly setSaveState: (state: TimelineSaveState) => void;
  readonly setSaveStateSecondaryMessage: (message: string | null) => void;
}) {
  const restoreConflictFocus = useCallback(
    (focusKey: string) => {
      window.setTimeout(() => {
        resolveInputElement(focusKey)?.focus();
      }, 0);
    },
    [resolveInputElement],
  );

  const updateSaveStateForConflicts = useCallback(
    (nextQueue: Record<string, LocalConflictState>) => {
      publishSaveStatePresentation(
        pendingSavesRefsRef.current.pendingQueueRef.current,
        nextQueue,
      );
    },
    [pendingSavesRefsRef, publishSaveStatePresentation],
  );

  const registerSameFieldConflict = useCallback(
    (
      conflict: SameFieldConflictPayload,
      focusKey: string,
      surface: TimelineScalarEditorSurface,
    ) => {
      const queueKey = sameFieldConflictQueueKey(conflict);
      const binding = timelineScalarBindingForField(conflict.field_key);
      if (binding !== null && typeof conflict.client_value === "string") {
        scalarDraftValuesRef.current.set(
          inputFocusKey(conflict.record_id, binding.key, surface),
          conflict.client_value,
        );
      }
      if (binding !== null) {
        setRows((current) => {
          const nextRows = current.map((row) => {
            if (row.recordId !== conflict.record_id) {
              return row;
            }
            const serverText =
              typeof conflict.server_value === "string"
                ? conflict.server_value
                : "";
            return {
              ...row,
              rowVersion: conflict.current_row_version,
              values: { ...row.values, [binding.key]: serverText },
              committedValues: {
                ...row.committedValues,
                [binding.key]: serverText,
              },
              pendingSignature: null,
            };
          });
          rowsRef.current = nextRows;
          return nextRows;
        });
      }
      setConflictQueueState((current) => {
        const existing = current[queueKey];
        const mergedDraft =
          existing?.mergedDraft ??
          (typeof conflict.suggested_merged_value === "string"
            ? conflict.suggested_merged_value
            : typeof conflict.server_value === "string"
              ? conflict.server_value
              : "");
        const next = {
          ...current,
          [queueKey]: {
            key: queueKey,
            anchor: saveStateConflictAnchorFromPayload(conflict),
            conflict,
            focusKey,
            localValue:
              conflict.client_value === undefined
                ? existing?.localValue
                : conflict.client_value,
            mergedDraft,
          },
        };
        updateSaveStateForConflicts(next);
        return next;
      });
      setActiveConflictKey(queueKey);
    },
    [
      rowsRef,
      scalarDraftValuesRef,
      setActiveConflictKey,
      setConflictQueueState,
      setRows,
      updateSaveStateForConflicts,
    ],
  );

  const handleMutationConflict = useCallback(
    (
      payload: unknown,
      rowKey: string,
      focusField: FocusFieldKey,
      surface: TimelineScalarEditorSurface,
    ) => {
      const conflict = parseSameFieldConflict(payload);
      if (conflict === null) {
        return false;
      }
      registerSameFieldConflict(
        conflict,
        inputFocusKey(rowKey, focusField, surface),
        surface,
      );
      return true;
    },
    [registerSameFieldConflict],
  );

  const activeConflict = useMemo(
    () =>
      activeConflictKey === null
        ? null
        : (conflictQueue[activeConflictKey] ?? null),
    [activeConflictKey, conflictQueue],
  );
  const activePasteConflictKeys = useMemo(
    () => pasteConflictGroup?.keys.filter((key) => conflictQueue[key]) ?? [],
    [conflictQueue, pasteConflictGroup],
  );
  const activePasteConflictIndex =
    activeConflictKey === null
      ? -1
      : activePasteConflictKeys.indexOf(activeConflictKey);
  const showPasteConflictNavigator =
    activePasteConflictKeys.length > 1 && activePasteConflictIndex >= 0;

  useEffect(() => {
    if (activeConflict === null) {
      return;
    }
    window.setTimeout(() => {
      (
        document.querySelector(
          dataTestIdSelector("conflict-resolver-summary"),
        ) as HTMLElement | null
      )?.focus();
    }, 0);
  }, [activeConflict]);

  const closeConflictResolver = useCallback(
    (conflict: LocalConflictState) => {
      setActiveConflictKey(null);
      restoreConflictFocus(conflict.focusKey);
    },
    [restoreConflictFocus, setActiveConflictKey],
  );

  useEffect(() => {
    if (!activeConflict) {
      return;
    }
    const handleConflictResolverEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") {
        return;
      }
      event.preventDefault();
      closeConflictResolver(activeConflict);
    };
    document.addEventListener("keydown", handleConflictResolverEscape);
    return () => {
      document.removeEventListener("keydown", handleConflictResolverEscape);
    };
  }, [activeConflict, closeConflictResolver]);

  const clearLocalConflict = useCallback(
    (conflict: LocalConflictState) => {
      pendingSavesRefsRef.current.pendingQueueRef.current.model.clearSameFieldConflict(
        conflict.key,
      );
      setConflictQueueState((current) => {
        const next = { ...current };
        delete next[conflict.key];
        updateSaveStateForConflicts(next);
        return next;
      });
      setActiveConflictKey((current) =>
        current === conflict.key ? null : current,
      );
      setPasteConflictGroup((current) => {
        if (current === null || !current.keys.includes(conflict.key)) {
          return current;
        }
        const keys = current.keys.filter((key) => key !== conflict.key);
        return keys.length > 1 ? { keys } : null;
      });
      restoreConflictFocus(conflict.focusKey);
      schedulePendingReplayRef.current();
    },
    [
      pendingSavesRefsRef,
      restoreConflictFocus,
      schedulePendingReplayRef,
      setActiveConflictKey,
      setConflictQueueState,
      setPasteConflictGroup,
      updateSaveStateForConflicts,
    ],
  );

  const submitConflictResolution = useCallback(
    (
      conflict: LocalConflictState,
      resolutionKind: TimelineConflictResolution,
    ) => {
      const body = buildTimelineConflictResolutionPayload({
        clientTxnId: nextClientTxnId(),
        conflictResolutionClass: conflict.conflict.conflict_resolution_class,
        conflictToken: conflict.conflict.conflict_token,
        localValue: conflict.localValue,
        mergedDraft: conflict.mergedDraft,
        resolutionKind,
      });
      setSaveState("Syncing");
      setSaveStateSecondaryMessage("Workbook edits are syncing.");
      pendingSavesRefsRef.current.saveQueueRef.current =
        pendingSavesRefsRef.current.saveQueueRef.current
          .catch(() => undefined)
          .then(async () => {
            const result = await fetchJSON<TimelineMutationEnvelope>(
              apiPath(
                apiBase,
                `/api/v1/records/${conflict.conflict.record_id}/conflicts/${conflict.conflict.conflict_token}/resolve`,
              ),
              {
                method: "POST",
                body: JSON.stringify(body),
              },
            );
            if (!result.ok) {
              const refreshedConflict = parseSameFieldConflict(result.payload);
              if (refreshedConflict !== null) {
                registerSameFieldConflict(
                  refreshedConflict,
                  conflict.focusKey,
                  "grid",
                );
                setSaveState("Conflict");
                setSaveStateSecondaryMessage("Conflict requires review.");
                return;
              }
              setSaveState("Conflict");
              setSaveStateSecondaryMessage("Conflict requires review.");
              return;
            }
            const envelope = readEnvelope<TimelineMutationEnvelope>(
              result.payload,
            );
            scalarDraftValuesRef.current.delete(conflict.focusKey);
            const binding = timelineScalarBindingForField(
              conflict.conflict.field_key,
            );
            if (binding !== null) {
              for (const surface of timelineScalarEditorSurfaces) {
                scalarDraftValuesRef.current.delete(
                  inputFocusKey(
                    conflict.conflict.record_id,
                    binding.key,
                    surface,
                  ),
                );
              }
            }
            applyRowMutation(conflict.conflict.record_id, envelope, {
              viewportContinuityToken: beginViewportContinuity({
                kind: "input",
                focusKey: conflict.focusKey,
              }),
            });
            clearLocalConflict(conflict);
          });
    },
    [
      apiBase,
      applyRowMutation,
      beginViewportContinuity,
      clearLocalConflict,
      nextClientTxnId,
      pendingSavesRefsRef,
      registerSameFieldConflict,
      scalarDraftValuesRef,
      setSaveState,
      setSaveStateSecondaryMessage,
    ],
  );

  const handleConflictMergedDraftChange = useCallback(
    (conflictKey: string, value: string) => {
      setConflictQueueState((current) => {
        const conflict = current[conflictKey];
        if (conflict === undefined) {
          return current;
        }
        return {
          ...current,
          [conflictKey]: {
            ...conflict,
            mergedDraft: value,
          },
        };
      });
    },
    [setConflictQueueState],
  );

  return {
    commands: {
      clearLocalConflict,
      closeConflictResolver,
      handleConflictMergedDraftChange,
      handleMutationConflict,
      registerSameFieldConflict,
      submitConflictResolution,
    },
    snapshot: {
      activeConflict,
      activePasteConflictIndex,
      activePasteConflictKeys,
      showPasteConflictNavigator,
    },
  };
}
