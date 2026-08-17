import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useMemo, useRef } from "react";
import { flushSync } from "react-dom";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import { useWorkbookMutationRuntime } from "../../runtime/useWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type {
  WorkbookPendingQueueSnapshot,
  WorkbookPendingSavesRefs,
} from "../../runtime/workbookPendingReplayRuntime";
import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { useTimelineConflictProjectionAdapter } from "../hooks/useTimelineConflictProjectionAdapter";
import { useTimelineConflicts } from "../hooks/useTimelineConflicts";
import { useTimelineSaveStatePresentation } from "../hooks/useTimelineSaveStatePresentation";
import type { TimelineViewportContinuityTarget } from "../hooks/useTimelineViewportContinuityController";
import type {
  PendingReplayRuntimeMeta,
  TimelineMutableRef,
  TimelineRowMutationEditorPort,
  TimelineRowStoreCommands,
} from "../models/timelineControllerPorts";
import { ensureTimelineDraftRow } from "../models/timelineRowsModel";
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
  type LocalConflictState,
  rowFromApi,
  type TimelineApiRow,
  timelineFieldBinding,
  timelineScalarBindings,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

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
  clearActiveCollectionInputKey,
  clearViewportContinuity,
  createdRowPresentationScopeKey,
  editorDraftRegistry,
  editorPort,
  mutationRuntime,
  nextDraftIndex,
  pendingQueueSnapshot,
  pendingSavesRefs,
  rowsRef,
  selectedRowId,
  setAutoResolutionNotices,
  setDismissedMentionsByRow,
  setPendingQueueSnapshot,
  rowStoreCommands,
  setSelectedRowId,
}: {
  readonly advanceViewportContinuity: (
    token?: number,
    options?: { readonly target?: TimelineViewportContinuityTarget | null },
  ) => void;
  readonly clearActiveCollectionInputKey: (focusKey: string) => void;
  readonly clearViewportContinuity: (token: number) => void;
  readonly createdRowPresentationScopeKey: string;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly editorPort: TimelineRowMutationEditorPort;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly nextDraftIndex: () => number;
  readonly pendingQueueSnapshot: WorkbookPendingQueueSnapshot;
  readonly pendingSavesRefs: WorkbookPendingSavesRefs<PendingReplayRuntimeMeta>;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly selectedRowId: string | null;
  readonly setAutoResolutionNotices: Dispatch<
    SetStateAction<AutoResolutionNotice[]>
  >;
  readonly setDismissedMentionsByRow: Dispatch<
    SetStateAction<Record<string, DismissedMention[]>>
  >;
  readonly setPendingQueueSnapshot: (
    snapshot: WorkbookPendingQueueSnapshot,
  ) => void;
  readonly rowStoreCommands: TimelineRowStoreCommands;
  readonly setSelectedRowId: (recordId: string | null) => void;
}) {
  const { replaceRows, updateRows } = rowStoreCommands;
  const conflictQueueRef = useRef<Record<string, LocalConflictState>>({});
  const createdRowPresentationRef = useRef<{
    recordId: string | null;
    scopeKey: string;
  }>({ recordId: null, scopeKey: createdRowPresentationScopeKey });
  const conflicts = useTimelineConflicts({ conflictQueueRef });
  const commonMutationSnapshot = useWorkbookMutationRuntime(
    mutationRuntime,
    timelineViewSchemaId,
  );
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
    pendingSavesRefs,
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
      mutation: Pick<WorkbookPendingMutationAccepted, "row" | "viewSchemaId">,
      options: TimelineMutationApplyOptions = {},
    ) => {
      let previousRow = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      validateTimelineViewSchemaId(mutation.viewSchemaId, "mutation response");
      const responseRow = mutation.row;
      recordWorkbookTiming("apply_row_mutation_start", {
        kind: "row_mutation",
      });
      const accepted = acceptCommittedTimelineRow(rowFromApi(responseRow));
      const committed = accepted.row;
      let draftSummaryKey: string | null = null;
      flushSync(() => {
        updateRows((current) => {
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
          const hydrated = ensureTimelineDraftRow({
            nextDraftIndex,
            rows: nextRows,
          });
          draftSummaryKey = hydrated.draftFocusKey;
          rowsRef.current = hydrated.rows;
          return hydrated.rows;
        });
        if (options.clearActiveCollectionFocusKey !== undefined) {
          clearActiveCollectionInputKey(options.clearActiveCollectionFocusKey);
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
      const createdFromDraft = previousRow?.recordId === null;
      if (createdFromDraft && committed.recordId !== null) {
        createdRowPresentationRef.current = {
          recordId: committed.recordId,
          scopeKey: createdRowPresentationScopeKey,
        };
      }
      if (
        createdFromDraft &&
        options.continueOnFreshDraft &&
        draftSummaryKey !== null &&
        committed.recordId !== null
      ) {
        if (options.viewportContinuityToken !== undefined) {
          clearViewportContinuity(options.viewportContinuityToken);
        }
        editorPort.reveal({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: committed.recordId,
        });
        editorPort.focusInput(draftSummaryKey);
      } else {
        const target =
          options.continueOnFreshDraft && draftSummaryKey !== null
            ? ({ kind: "input", focusKey: draftSummaryKey } as const)
            : options.promoteToCommittedRowInspect &&
                committed.recordId !== null
              ? ({
                  kind: "row-inspect",
                  recordId: committed.recordId,
                } as const)
              : null;
        advanceViewportContinuity(options.viewportContinuityToken, { target });
        if (target?.kind === "input") editorPort.focusInput(target.focusKey);
      }
      recordWorkbookTiming("apply_row_mutation_end", { kind: "row_mutation" });
      return committed;
    },
    [
      acceptCommittedTimelineRow,
      advanceViewportContinuity,
      clearViewportContinuity,
      createdRowPresentationScopeKey,
      editorPort,
      nextDraftIndex,
      pruneAutoResolutionNoticesForRows,
      rowsRef,
      selectedRowId,
      clearActiveCollectionInputKey,
      setAutoResolutionNotices,
      setDismissedMentionsByRow,
      updateRows,
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

  const trackPendingSocketTxn = useCallback(
    (clientTxnId: string) => {
      const timeouts = pendingSavesRefs.pendingSocketTxnTimeoutsRef.current;
      const existingTimeout = timeouts.get(clientTxnId);
      if (existingTimeout !== undefined) window.clearTimeout(existingTimeout);
      const timeoutId = window.setTimeout(() => {
        timeouts.delete(clientTxnId);
      }, 30_000);
      timeouts.set(clientTxnId, timeoutId);
    },
    [pendingSavesRefs],
  );

  const resolvePendingSocketTxn = useCallback(
    (clientTxnId: string | null | undefined) => {
      if (!clientTxnId) return false;
      const timeouts = pendingSavesRefs.pendingSocketTxnTimeoutsRef.current;
      const timeoutId = timeouts.get(clientTxnId);
      if (timeoutId === undefined) return false;
      window.clearTimeout(timeoutId);
      timeouts.delete(clientTxnId);
      return true;
    },
    [pendingSavesRefs],
  );

  const enqueueSaveWork = useCallback(
    (work: () => Promise<void>) => {
      pendingSavesRefs.saveQueueRef.current =
        pendingSavesRefs.saveQueueRef.current.catch(() => undefined).then(work);
    },
    [pendingSavesRefs],
  );

  const reconcileDiscardedPendingUnit = useCallback(
    (
      discardedUnit: PendingReplayUnitState,
      remainingUnits: readonly PendingReplayUnitState[],
    ) => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
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
        replaceRows(nextRows);
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
      replaceRows(nextRows);
    },
    [
      editorDraftRegistry,
      editorPort,
      latestCommittedTimelineRow,
      nextDraftIndex,
      pendingSavesRefs,
      rowsRef,
      replaceRows,
    ],
  );

  useEffect(
    () => () => {
      const pendingRefs = pendingSavesRefs;
      for (const timeoutId of pendingRefs.pendingSocketTxnTimeoutsRef.current.values()) {
        window.clearTimeout(timeoutId);
      }
      pendingRefs.pendingSocketTxnTimeoutsRef.current.clear();
      if (pendingRefs.pendingReplayTimerRef.current !== null) {
        window.clearTimeout(pendingRefs.pendingReplayTimerRef.current);
        pendingRefs.pendingReplayTimerRef.current = null;
      }
    },
    [pendingSavesRefs],
  );

  const conflictProjection = useTimelineConflictProjectionAdapter({
    acceptCommittedRow: acceptCommittedTimelineRow,
    activeConflictKey,
    conflictQueue,
    editorDraftRegistry,
    mutationRuntime,
    rowsRef,
    rowStoreCommands,
    setActiveConflictKey,
    setConflictQueueState,
  });
  const { activeConflict } = conflictProjection.snapshot;
  const { registerSameFieldConflict } = conflictProjection.commands;

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

  const currentCreatedRowPresentationRecordId = useCallback(
    () =>
      createdRowPresentationRef.current.scopeKey ===
      createdRowPresentationScopeKey
        ? createdRowPresentationRef.current.recordId
        : null,
    [createdRowPresentationScopeKey],
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
    },
    ports: {
      collaborationAdmission,
      queryAdmission: {
        beginLoad,
        committedRowsChangedSince,
        currentCreatedRowPresentationRecordId,
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
