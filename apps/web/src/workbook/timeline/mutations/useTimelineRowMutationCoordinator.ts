import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useMemo, useRef } from "react";
import { flushSync } from "react-dom";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { useWorkbookMutationRuntime } from "../../runtime/useWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type {
  WorkbookPendingQueueSnapshot,
  WorkbookPendingSavesRefs,
} from "../../runtime/workbookPendingReplayRuntime";
import { refreshBlocksWorkbookPendingRecord } from "../../runtime/workbookPendingReplayRuntime";
import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { useTimelineConflictProjectionAdapter } from "../hooks/useTimelineConflictProjectionAdapter";
import { useTimelineConflicts } from "../hooks/useTimelineConflicts";
import type { LoadRowsOptions } from "../hooks/useTimelineRowsLoader";
import { useTimelineSaveStatePresentation } from "../hooks/useTimelineSaveStatePresentation";
import type { TimelineViewportContinuityTarget } from "../hooks/useTimelineViewportContinuityController";
import type {
  PendingReplayRuntimeMeta,
  TimelineMutableRef,
  TimelineRowMutationEditorPort,
} from "../models/timelineControllerPorts";
import type { DismissedMention } from "../models/workbookMentionChips";
import {
  type AutoResolutionNotice,
  buildAutoResolutionNotices,
  reconcileDismissedMentionsForRow,
} from "../models/workbookMentionChips";
import {
  type CollectionDraftKey,
  createDraftRow,
  type FocusFieldKey,
  inputFocusKey,
  type LocalConflictState,
  rowFromApi,
  type TimelineApiRow,
  timelineFieldBinding,
  timelineScalarBindingForField,
  timelineScalarBindings,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import type {
  TimelinePendingMutationAccepted,
  TimelinePendingMutationPort,
} from "../ports/TimelinePendingMutationPort";

type TimelineMutationApplyOptions = {
  readonly clearActiveCollectionFocusKey?: string | undefined;
  readonly continueOnFreshDraft?: boolean;
  readonly detectAutoResolution?: boolean;
  readonly promoteToCommittedRowInspect?: boolean;
  readonly viewportContinuityToken?: number;
};

function recordWorkbookTiming(
  name: string,
  details: Record<string, unknown> = {},
) {
  if (typeof performance === "undefined") return;
  performance.mark(`cartulary.workbook.${name}`, { detail: details });
}

function rowStillHasAutoResolvedNotice(
  row: WorkbookRow,
  notice: AutoResolutionNotice,
) {
  if (row.recordId !== notice.rowRecordId) return false;
  const item = [
    ...row.collectionValues.hostRefs,
    ...row.collectionValues.identityRefs,
  ].find((candidate) => candidate.itemRef === notice.itemRef);
  return (
    item?.itemKind === "resolved_ref" &&
    item.autoResolved &&
    item.resolvedRecordId !== null
  );
}

function ensureDraftRowWithFreshIndex(
  rows: WorkbookRow[],
  nextDraftIndex: () => number,
) {
  if (rows.some((row) => row.recordId === null)) {
    return { rows, draftSummaryKey: null };
  }
  const draftIndex = nextDraftIndex();
  return {
    rows: [...rows, createDraftRow(draftIndex)],
    draftSummaryKey: inputFocusKey(
      `draft-${draftIndex}`,
      "activitySynopsisText",
    ),
  };
}

function isCollectionDraftKey(
  field: FocusFieldKey,
): field is CollectionDraftKey {
  return field === "hostRefs" || field === "identityRefs" || field === "tags";
}

/**
 * Single Timeline owner for row-version admission and transitions between
 * query, fresh/replayed mutations, live patches, conflicts, and continuity.
 */
export function useTimelineRowMutationCoordinator({
  advanceViewportContinuity,
  discardBlockedEditRef,
  editorDraftRegistry,
  editorPort,
  loadRowsRef,
  mutationRuntime,
  nextDraftIndex,
  pendingQueueSnapshot,
  pendingSavesRefsRef,
  rowsRef,
  schedulePendingReplayRef,
  selectedRowId,
  setActiveCollectionInputKey,
  setAutoResolutionNotices,
  setDismissedMentionsByRow,
  setPendingQueueSnapshot,
  setRows,
  setSelectedRowId,
  timelinePendingMutation,
}: {
  readonly advanceViewportContinuity: (
    token?: number,
    options?: { readonly target?: TimelineViewportContinuityTarget | null },
  ) => void;
  readonly discardBlockedEditRef: TimelineMutableRef<
    (unitId: string) => boolean
  >;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly editorPort: TimelineRowMutationEditorPort;
  readonly loadRowsRef: TimelineMutableRef<
    (options: LoadRowsOptions) => Promise<void>
  >;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly nextDraftIndex: () => number;
  readonly pendingQueueSnapshot: WorkbookPendingQueueSnapshot;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    WorkbookPendingSavesRefs<PendingReplayRuntimeMeta>
  >;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly schedulePendingReplayRef: TimelineMutableRef<() => void>;
  readonly selectedRowId: string | null;
  readonly setActiveCollectionInputKey: Dispatch<SetStateAction<string | null>>;
  readonly setAutoResolutionNotices: Dispatch<
    SetStateAction<AutoResolutionNotice[]>
  >;
  readonly setDismissedMentionsByRow: Dispatch<
    SetStateAction<Record<string, DismissedMention[]>>
  >;
  readonly setPendingQueueSnapshot: (
    snapshot: WorkbookPendingQueueSnapshot,
  ) => void;
  readonly setRows: Dispatch<SetStateAction<WorkbookRow[]>>;
  readonly setSelectedRowId: (recordId: string | null) => void;
  readonly timelinePendingMutation: TimelinePendingMutationPort;
}) {
  const conflictQueueRef = useRef<Record<string, LocalConflictState>>({});
  const conflicts = useTimelineConflicts({ conflictQueueRef });
  const commonMutationSnapshot = useWorkbookMutationRuntime(mutationRuntime);
  const { activeConflictKey, conflictQueue, pasteConflictGroup } =
    conflicts.snapshot;
  const { setActiveConflictKey, setConflictQueueState, setPasteConflictGroup } =
    conflicts.commands;

  useEffect(() => {
    const commonKeys = new Set(
      commonMutationSnapshot.conflicts
        .filter((entry) => entry.origin.viewSchemaId === timelineViewSchemaId)
        .map((entry) => entry.key),
    );
    setConflictQueueState((current) =>
      Object.fromEntries(
        Object.entries(current).filter(([key]) => commonKeys.has(key)),
      ),
    );
    setActiveConflictKey((current) =>
      current !== null && commonKeys.has(current) ? current : null,
    );
  }, [
    commonMutationSnapshot.conflicts,
    setActiveConflictKey,
    setConflictQueueState,
  ]);

  const saveState = useTimelineSaveStatePresentation<PendingReplayRuntimeMeta>({
    conflictQueue,
    conflictQueueRef,
    mutationRuntime,
    pendingQueueSnapshot,
    pendingSavesRefsRef,
    setPendingQueueSnapshot,
  });
  const {
    beginRefreshInFlight,
    beginSave,
    finishRefreshInFlight,
    finishSave,
    publishPendingQueueState,
    publishSaveStatePresentation,
  } = saveState.commands;
  useEffect(() => {
    publishPendingQueueState(conflictQueue);
  }, [conflictQueue, publishPendingQueueState]);

  const committedRows = useTimelineCommittedRows({ rowsRef });
  const {
    acceptCommittedTimelineRow,
    acceptCommittedTimelineRows,
    acceptTimelineActionResult,
    acceptTimelineRecordVersion,
    beginLoad,
    committedRowsChangedSince,
    currentCommittedTimelineRow,
    hasLoadedRows,
    isCurrentLoadSequence,
    isStaleTimelineRowVersion,
    knownTimelineRowVersion,
    latestCommittedRowVersion,
    latestCommittedTimelineRow,
    markRowsLoaded,
  } = committedRows.commands;

  const pruneAutoResolutionNoticesForRows = useCallback(
    (committed: readonly WorkbookRow[]) => {
      if (committed.length < 1) return;
      setAutoResolutionNotices((current) =>
        current.filter((notice) => {
          const row = committed.find(
            (candidate) => candidate.recordId === notice.rowRecordId,
          );
          return (
            row === undefined || rowStillHasAutoResolvedNotice(row, notice)
          );
        }),
      );
    },
    [setAutoResolutionNotices],
  );

  const applyAcceptedRowMutation = useCallback(
    (
      rowKey: string,
      mutation: Pick<TimelinePendingMutationAccepted, "row" | "viewSchemaId">,
      options: TimelineMutationApplyOptions = {},
    ) => {
      let previousRow = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      validateTimelineViewSchemaId(mutation.viewSchemaId, "mutation response");
      const responseRow = mutation.row;
      recordWorkbookTiming("apply_row_mutation_start", {
        rowKey,
        recordId: responseRow.record_id,
        rowVersion: responseRow.row_version,
      });
      const accepted = acceptCommittedTimelineRow(rowFromApi(responseRow));
      const committed = accepted.row;
      let draftSummaryKey: string | null = null;
      flushSync(() => {
        setRows((current) => {
          previousRow =
            current.find((candidate) => candidate.key === rowKey) ??
            current.find(
              (candidate) => candidate.recordId === committed.recordId,
            ) ??
            previousRow;
          let replaced = false;
          let nextRows = current.map((row) => {
            if (
              row.key !== rowKey &&
              (committed.recordId === null ||
                row.recordId !== committed.recordId)
            ) {
              return row;
            }
            replaced = true;
            return committed;
          });
          if (!replaced) {
            const draftIndex = nextRows.findIndex(
              (row) => row.recordId === null,
            );
            nextRows =
              draftIndex === -1
                ? [...nextRows, committed]
                : [
                    ...nextRows.slice(0, draftIndex),
                    committed,
                    ...nextRows.slice(draftIndex),
                  ];
          }
          const hydrated = ensureDraftRowWithFreshIndex(
            nextRows,
            nextDraftIndex,
          );
          draftSummaryKey = hydrated.draftSummaryKey;
          rowsRef.current = hydrated.rows;
          return hydrated.rows;
        });
        if (options.clearActiveCollectionFocusKey !== undefined) {
          setActiveCollectionInputKey((current) =>
            current === options.clearActiveCollectionFocusKey ? null : current,
          );
        }
      });

      if (committed.recordId !== null) {
        setDismissedMentionsByRow((current) =>
          reconcileDismissedMentionsForRow(current, committed),
        );
        pruneAutoResolutionNoticesForRows([committed]);
      }
      if (options.detectAutoResolution !== false) {
        const notices = buildAutoResolutionNotices(previousRow, committed);
        if (notices.length > 0) {
          setAutoResolutionNotices((current) => {
            const knownRefs = new Set(current.map((notice) => notice.itemRef));
            return [
              ...current,
              ...notices.filter((notice) => !knownRefs.has(notice.itemRef)),
            ];
          });
        }
      }
      if (
        selectedRowId !== null &&
        previousRow?.recordId !== null &&
        previousRow !== undefined &&
        previousRow.recordId === selectedRowId
      ) {
        setSelectedRowId(committed.recordId);
      }
      const target =
        options.continueOnFreshDraft && draftSummaryKey !== null
          ? ({ kind: "input", focusKey: draftSummaryKey } as const)
          : options.promoteToCommittedRowInspect && committed.recordId !== null
            ? ({ kind: "row-inspect", recordId: committed.recordId } as const)
            : null;
      advanceViewportContinuity(options.viewportContinuityToken, { target });
      if (target?.kind === "input") editorPort.focusInput(target.focusKey);
      recordWorkbookTiming("apply_row_mutation_end", {
        rowKey,
        recordId: committed.recordId,
        rowVersion: committed.rowVersion,
      });
      return committed;
    },
    [
      acceptCommittedTimelineRow,
      advanceViewportContinuity,
      editorPort,
      nextDraftIndex,
      pruneAutoResolutionNoticesForRows,
      rowsRef,
      selectedRowId,
      setActiveCollectionInputKey,
      setAutoResolutionNotices,
      setDismissedMentionsByRow,
      setRows,
      setSelectedRowId,
    ],
  );

  const applyClipboardResponseRows = useCallback(
    (responseRows: readonly TimelineApiRow[]) => {
      for (const row of responseRows) {
        applyAcceptedRowMutation(row.record_id, {
          row,
          viewSchemaId: timelineViewSchemaId,
        });
      }
    },
    [applyAcceptedRowMutation],
  );

  const waitForCommittedRecordIdle = useCallback(
    async (
      recordId: string,
      options: {
        readonly fallbackRowVersion?: number | null | undefined;
        readonly refreshIfMissing?: boolean;
      } = {},
    ): Promise<{ row: WorkbookRow | null; rowVersion: number } | null> => {
      let attemptedRefresh = false;
      for (;;) {
        const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
        const snapshot = pending.model.snapshot();
        const hasPendingRecordWork = snapshot.units.some(
          (unit) => unit.recordId === recordId,
        );
        if (
          snapshot.authPaused ||
          snapshot.halted !== null ||
          snapshot.overflow !== null ||
          snapshot.sameFieldConflicts.length > 0 ||
          Object.keys(conflictQueueRef.current).length > 0
        ) {
          return null;
        }
        if (
          !hasPendingRecordWork &&
          !refreshBlocksWorkbookPendingRecord(pending, recordId)
        ) {
          const row = latestCommittedTimelineRow(recordId);
          const rowVersion =
            latestCommittedRowVersion(recordId) ?? options.fallbackRowVersion;
          if (typeof rowVersion === "number") return { row, rowVersion };
          if (
            options.refreshIfMissing !== false &&
            attemptedRefresh === false
          ) {
            attemptedRefresh = true;
            await loadRowsRef.current({ showLoading: false });
            continue;
          }
          return null;
        }
        await new Promise((resolve) => window.setTimeout(resolve, 16));
      }
    },
    [
      latestCommittedRowVersion,
      latestCommittedTimelineRow,
      loadRowsRef,
      pendingSavesRefsRef,
    ],
  );

  const trackPendingSocketTxn = useCallback(
    (clientTxnId: string) => {
      const timeouts =
        pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current;
      const existingTimeout = timeouts.get(clientTxnId);
      if (existingTimeout !== undefined) window.clearTimeout(existingTimeout);
      const timeoutId = window.setTimeout(() => {
        timeouts.delete(clientTxnId);
      }, 30_000);
      timeouts.set(clientTxnId, timeoutId);
    },
    [pendingSavesRefsRef],
  );

  const resolvePendingSocketTxn = useCallback(
    (clientTxnId: string | null | undefined) => {
      if (!clientTxnId) return false;
      const timeouts =
        pendingSavesRefsRef.current.pendingSocketTxnTimeoutsRef.current;
      const timeoutId = timeouts.get(clientTxnId);
      if (timeoutId === undefined) return false;
      window.clearTimeout(timeoutId);
      timeouts.delete(clientTxnId);
      return true;
    },
    [pendingSavesRefsRef],
  );

  const enqueueSaveWork = useCallback(
    (work: () => Promise<void>) => {
      pendingSavesRefsRef.current.saveQueueRef.current =
        pendingSavesRefsRef.current.saveQueueRef.current
          .catch(() => undefined)
          .then(work);
    },
    [pendingSavesRefsRef],
  );

  const reconcileDiscardedPendingUnit = useCallback(
    (
      discardedUnit: PendingReplayUnitState,
      remainingUnits: readonly PendingReplayUnitState[],
    ) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      const discardedMeta = pending.metaByUnitId.get(discardedUnit.id);
      const remainingForRow = remainingUnits
        .filter(
          (unit) =>
            unit.rowKey === discardedUnit.rowKey ||
            (discardedUnit.recordId !== null &&
              unit.recordId === discardedUnit.recordId),
        )
        .sort((left, right) => left.enqueueOrder - right.enqueueOrder);
      const remainingFocusKeys = new Set(
        remainingForRow
          .map((unit) => pending.metaByUnitId.get(unit.id)?.focusKey)
          .filter((focusKey): focusKey is string => focusKey !== undefined),
      );
      if (
        discardedMeta !== undefined &&
        !remainingFocusKeys.has(discardedMeta.focusKey)
      ) {
        editorDraftRegistry.deleteDraftForFocusKey(discardedMeta.focusKey);
      }
      editorDraftRegistry.clearScalarDraftsForRow(
        discardedUnit.rowKey,
        remainingFocusKeys,
      );
      const discardedFocusField = discardedMeta?.focusField;
      const discardedScalarBinding =
        discardedFocusField === undefined ||
        isCollectionDraftKey(discardedFocusField)
          ? null
          : timelineScalarBindings.find(
              (binding) => binding.key === discardedFocusField,
            );
      if (
        discardedUnit.recordId !== null &&
        discardedScalarBinding !== null &&
        discardedScalarBinding !== undefined
      ) {
        editorPort.cancelEdit({
          fieldKey: discardedScalarBinding.fieldKey,
          recordId: discardedUnit.recordId,
        });
      }

      if (discardedUnit.kind === "create") {
        const nextRows = rowsRef.current.filter(
          (row) => row.key !== discardedUnit.rowKey,
        );
        if (!nextRows.some((row) => row.recordId === null)) {
          nextRows.push(createDraftRow(nextDraftIndex()));
        }
        rowsRef.current = nextRows;
        setRows(nextRows);
        return;
      }

      const currentRow = rowsRef.current.find(
        (row) =>
          row.key === discardedUnit.rowKey ||
          row.recordId === discardedUnit.recordId,
      );
      const committedRow =
        discardedUnit.recordId === null
          ? null
          : latestCommittedTimelineRow(discardedUnit.recordId);
      if (committedRow === null && currentRow === undefined) return;
      const baseRow = committedRow ?? currentRow;
      if (baseRow === undefined) return;
      let reconciled: WorkbookRow = {
        ...baseRow,
        key: discardedUnit.rowKey,
        values: { ...baseRow.committedValues },
        pendingSignature: remainingForRow.at(-1)?.mutationSignature ?? null,
      };
      if (reconciled.rawRow !== null) {
        const committedScalarCells = Object.fromEntries(
          timelineScalarBindings.map((binding) => [
            binding.fieldKey,
            { value: reconciled.committedValues[binding.key] },
          ]),
        );
        reconciled = {
          ...reconciled,
          rawRow: {
            ...reconciled.rawRow,
            cells: { ...reconciled.rawRow.cells, ...committedScalarCells },
          },
        };
      }

      for (const unit of remainingForRow) {
        const meta = pending.metaByUnitId.get(unit.id);
        const changes = Array.isArray(unit.payloadIntent.changes)
          ? unit.payloadIntent.changes
          : [];
        for (const change of changes) {
          if (change === null || typeof change !== "object") continue;
          const candidate = change as Record<string, unknown>;
          if (typeof candidate.field_key !== "string") continue;
          const binding = timelineFieldBinding(candidate.field_key);
          if (binding.kind === "scalar" && "value" in candidate) {
            const value =
              typeof candidate.value === "string" ? candidate.value : "";
            reconciled = {
              ...reconciled,
              values: { ...reconciled.values, [binding.key]: value },
              rawRow:
                reconciled.rawRow === null
                  ? null
                  : {
                      ...reconciled.rawRow,
                      cells: {
                        ...reconciled.rawRow.cells,
                        [binding.fieldKey]: { value: candidate.value },
                      },
                    },
            };
          }
        }
        if (meta !== undefined && isCollectionDraftKey(meta.focusField)) {
          reconciled = {
            ...reconciled,
            collectionDrafts: {
              ...reconciled.collectionDrafts,
              [meta.focusField]:
                meta.rowSnapshot.collectionDrafts[meta.focusField],
            },
          };
        }
      }

      const nextRows = rowsRef.current.map((row) =>
        row.key === discardedUnit.rowKey ||
        row.recordId === discardedUnit.recordId
          ? reconciled
          : row,
      );
      rowsRef.current = nextRows;
      setRows(nextRows);
    },
    [
      editorDraftRegistry,
      editorPort,
      latestCommittedTimelineRow,
      nextDraftIndex,
      pendingSavesRefsRef,
      rowsRef,
      setRows,
    ],
  );

  useEffect(
    () => () => {
      const pendingRefs = pendingSavesRefsRef.current;
      for (const timeoutId of pendingRefs.pendingSocketTxnTimeoutsRef.current.values()) {
        window.clearTimeout(timeoutId);
      }
      pendingRefs.pendingSocketTxnTimeoutsRef.current.clear();
      if (pendingRefs.pendingReplayTimerRef.current !== null) {
        window.clearTimeout(pendingRefs.pendingReplayTimerRef.current);
        pendingRefs.pendingReplayTimerRef.current = null;
      }
    },
    [pendingSavesRefsRef],
  );

  useEffect(
    () =>
      mutationRuntime.registerDrainer(() => schedulePendingReplayRef.current()),
    [mutationRuntime, schedulePendingReplayRef],
  );

  const conflictProjection = useTimelineConflictProjectionAdapter({
    acceptCommittedRow: acceptCommittedTimelineRow,
    activeConflictKey,
    conflictQueue,
    editorDraftRegistry,
    mutationRuntime,
    rowsRef,
    setActiveConflictKey,
    setConflictQueueState,
    setRows,
  });
  const { activeConflict } = conflictProjection.snapshot;
  const { registerSameFieldConflict } = conflictProjection.commands;

  useEffect(
    () =>
      mutationRuntime.registerSurface(
        timelineViewSchemaId,
        () => loadRowsRef.current({ showLoading: false }),
        async (mutation, conflict) => {
          const recordId = conflict.conflict.record_id;
          const binding = timelineScalarBindingForField(
            conflict.conflict.field_key,
          );
          if (binding !== null) {
            editorDraftRegistry.clearScalarDraftsForField(
              recordId,
              binding.key,
            );
            editorPort.cancelEdit({
              fieldKey: binding.fieldKey,
              recordId,
            });
          }
          const outcome = timelinePendingMutation.normalizeResolvedConflict({
            expectedRecordId: recordId,
            row: mutation.row,
            viewSchemaId: mutation.viewSchemaId,
          });
          if (outcome.kind === "accepted") {
            applyAcceptedRowMutation(recordId, outcome.value);
          } else {
            await loadRowsRef.current({ showLoading: false });
          }
          if (binding !== null) {
            window.setTimeout(() => {
              editorPort.focus({ fieldKey: binding.fieldKey, recordId });
            }, 0);
          }
        },
        (conflict) => {
          window.setTimeout(() => {
            const existingEditor =
              conflict.focusKey === null
                ? null
                : editorDraftRegistry.inputElementForFocusKey(
                    conflict.focusKey,
                  );
            if (existingEditor !== null) {
              existingEditor.focus({ preventScroll: true });
              return;
            }
            editorPort.activateEdit({
              fieldKey: conflict.conflict.field_key,
              recordId: conflict.conflict.record_id,
              value: conflict.localValue,
            });
          }, 0);
        },
        (unitId) => discardBlockedEditRef.current(unitId),
      ),
    [
      applyAcceptedRowMutation,
      discardBlockedEditRef,
      editorDraftRegistry,
      editorPort,
      loadRowsRef,
      mutationRuntime,
      timelinePendingMutation,
    ],
  );

  const collaborationAdmission = useMemo(
    () => ({
      acceptCommittedRow: acceptCommittedTimelineRow,
      acceptRecordVersion: acceptTimelineRecordVersion,
      isStaleRecordVersion: isStaleTimelineRowVersion,
    }),
    [
      acceptCommittedTimelineRow,
      acceptTimelineRecordVersion,
      isStaleTimelineRowVersion,
    ],
  );

  const activateConflict = useCallback(
    (key: string) => {
      mutationRuntime.activateConflict();
      setActiveConflictKey(key);
    },
    [mutationRuntime, setActiveConflictKey],
  );

  return {
    commands: {
      acceptCommittedTimelineRow,
      acceptCommittedTimelineRows,
      acceptTimelineActionResult,
      acceptTimelineRecordVersion,
      activateConflict,
      applyAcceptedRowMutation,
      applyClipboardResponseRows,
      beginRefreshInFlight,
      beginSave,
      committedRowsChangedSince,
      currentCommittedTimelineRow,
      enqueueSaveWork,
      finishRefreshInFlight,
      finishSave,
      hasLoadedRows,
      isCurrentLoadSequence,
      knownTimelineRowVersion,
      latestCommittedRowVersion,
      latestCommittedTimelineRow,
      markRowsLoaded,
      pruneAutoResolutionNoticesForRows,
      publishPendingQueueState,
      publishSaveStatePresentation,
      reconcileDiscardedPendingUnit,
      registerSameFieldConflict,
      resolvePendingSocketTxn,
      setActiveConflictKey,
      setPasteConflictGroup,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    },
    ports: {
      collaborationAdmission,
      queryAdmission: {
        beginLoad,
        committedRowsChangedSince,
        currentCommittedTimelineRow,
        hasLoadedRows,
        isCurrentLoadSequence,
        knownTimelineRowVersion,
        markRowsLoaded,
      },
    },
    refs: { conflictQueueRef },
    snapshot: {
      activeConflict,
      activeConflictKey,
      commonMutationSnapshot,
      conflictQueue,
      pasteConflictGroup,
    },
  };
}
