import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useMemo, useRef } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import { useWorkbookMutationRuntime } from "../../runtime/useWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { WorkbookPendingQueueSnapshot } from "../../runtime/workbookPendingReplayRuntime";
import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import { createTimelineSocketTransactionAdapter } from "../adapters/createTimelineSocketTransactionAdapter";
import { commitTimelineProjection } from "../adapters/timelineProjectionCommitAdapter";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { useTimelineConflictProjectionAdapter } from "../hooks/useTimelineConflictProjectionAdapter";
import { useTimelineConflicts } from "../hooks/useTimelineConflicts";
import { useTimelineSaveStatePresentation } from "../hooks/useTimelineSaveStatePresentation";
import type { TimelineViewportContinuityTarget } from "../hooks/useTimelineViewportContinuityController";
import {
  planTimelineAcceptedMutationEffects,
  type TimelineAcceptedContinuity,
} from "../models/timelineAcceptedMutationEffects";
import {
  projectAcceptedTimelineRow,
  type TimelineAcceptedProjection,
} from "../models/timelineAcceptedProjection";
import type { LocalConflictState } from "../models/timelineConflictState";
import type {
  TimelineMutableRef,
  TimelineReplayContext,
  TimelineRowMutationEditorPort,
  TimelineRowStoreCommands,
} from "../models/timelineControllerPorts";
import { reconcileDiscardedTimelineUnit } from "../models/timelineDiscardedReconciliation";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import {
  rowFromApi,
  type TimelineApiRow,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../models/timelineRowModel";
import type { DismissedMention } from "../models/workbookMentionChips";
import {
  type AutoResolutionNotice,
  reconcileDismissedMentionsForRow,
} from "../models/workbookMentionChips";

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

function appendAutoResolutionNotices(
  setAutoResolutionNotices: Dispatch<SetStateAction<AutoResolutionNotice[]>>,
  notices: readonly AutoResolutionNotice[],
) {
  if (notices.length < 1) return;
  setAutoResolutionNotices((current) => {
    const knownRefs = new Set(current.map((notice) => notice.itemRef));
    return [
      ...current,
      ...notices.filter((notice) => !knownRefs.has(notice.itemRef)),
    ];
  });
}

function completeAcceptedContinuity({
  advanceViewportContinuity,
  clearViewportContinuity,
  continuity,
  editorPort,
  viewportContinuityToken,
}: {
  readonly advanceViewportContinuity: (
    token?: number,
    options?: { readonly target?: TimelineViewportContinuityTarget | null },
  ) => void;
  readonly clearViewportContinuity: (token: number) => void;
  readonly continuity: TimelineAcceptedContinuity;
  readonly editorPort: TimelineRowMutationEditorPort;
  readonly viewportContinuityToken: number | undefined;
}) {
  if (continuity.kind === "fresh_draft") {
    if (viewportContinuityToken !== undefined) {
      clearViewportContinuity(viewportContinuityToken);
    }
    editorPort.reveal({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: continuity.recordId,
    });
    editorPort.focusInput(continuity.focusKey);
    return;
  }
  advanceViewportContinuity(viewportContinuityToken, {
    target: continuity.target,
  });
  if (continuity.target?.kind === "input") {
    editorPort.focusInput(continuity.target.focusKey);
  }
}

/**
 * Single Timeline owner for row-version admission and transitions between
 * query, fresh/replayed mutations, live patches, conflicts, and continuity.
 */
export function useTimelineRowMutationCoordinator({
  sheetRef,
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
  readonly sheetRef: SheetRef;
  readonly nextDraftIndex: () => number;
  readonly pendingQueueSnapshot: WorkbookPendingQueueSnapshot;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
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
  const mountedRef = useRef(true);
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);
  const conflicts = useTimelineConflicts({ conflictQueueRef });
  const commonMutationSnapshot = useWorkbookMutationRuntime(
    mutationRuntime,
    sheetRef,
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
    setConflictQueueState((current) => {
      const entries = Object.entries(current);
      const retained = entries.filter(([key]) => commonKeys.has(key));
      return retained.length === entries.length
        ? current
        : Object.fromEntries(retained);
    });
    setActiveConflictKey((current) =>
      current !== null && commonKeys.has(current) ? current : null,
    );
  }, [
    commonMutationSnapshot.conflicts,
    setActiveConflictKey,
    setConflictQueueState,
  ]);

  const saveState = useTimelineSaveStatePresentation({
    conflictQueue,
    sheetRef,
    mutationRuntime,
    pendingQueueSnapshot,
    pendingSavesRefs,
    setPendingQueueSnapshot,
  });
  const {
    beginRefreshInFlight,
    beginSave,
    publishPendingQueueState,
    publishSaveStatePresentation,
  } = saveState.commands;
  useEffect(() => {
    publishPendingQueueState();
  }, [publishPendingQueueState]);

  const committedRows = useTimelineCommittedRows({ rowsRef });
  const {
    acceptCommittedTimelineRow,
    acceptCommittedTimelineRows,
    acceptTimelineActionResult,
    acceptTimelineRecordVersion,
    beginLoad,
    currentCommittedTimelineRow,
    currentMutationEpoch,
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
      validateTimelineViewSchemaId(mutation.viewSchemaId, "mutation response");
      const responseRow = mutation.row;
      recordWorkbookTiming("apply_row_mutation_start", {
        kind: "row_mutation",
      });
      const accepted = acceptCommittedTimelineRow(rowFromApi(responseRow));
      const committed = accepted.row;
      // The FIFO owns settlement after navigation. A detached projection has
      // no React commit or viewport/focus effects to apply.
      if (!mountedRef.current) return committed;
      let projection: TimelineAcceptedProjection | undefined;
      commitTimelineProjection(() => {
        updateRows((current) => {
          projection = projectAcceptedTimelineRow({
            committed,
            currentRows: current,
            nextDraftIndex,
            rowKey,
          });
          rowsRef.current = projection.rows;
          return projection.rows;
        });
        if (options.clearActiveCollectionFocusKey !== undefined) {
          clearActiveCollectionInputKey(options.clearActiveCollectionFocusKey);
        }
      }, true);
      if (projection === undefined) {
        throw new Error("Timeline accepted projection was not committed.");
      }
      const effects = planTimelineAcceptedMutationEffects({
        committed,
        continueOnFreshDraft: options.continueOnFreshDraft === true,
        detectAutoResolution: options.detectAutoResolution !== false,
        projection,
        promoteToCommittedRowInspect:
          options.promoteToCommittedRowInspect === true,
        selectedRowId,
      });
      if (effects.reconcileDismissedMentions) {
        setDismissedMentionsByRow((current) =>
          reconcileDismissedMentionsForRow(current, committed),
        );
        pruneAutoResolutionNoticesForRows([committed]);
      }
      appendAutoResolutionNotices(
        setAutoResolutionNotices,
        effects.autoResolutionNotices,
      );
      if (effects.selectionUpdate !== null) {
        setSelectedRowId(effects.selectionUpdate.recordId);
      }
      if (effects.createdRecordId !== null) {
        createdRowPresentationRef.current = {
          recordId: effects.createdRecordId,
          scopeKey: createdRowPresentationScopeKey,
        };
      }
      completeAcceptedContinuity({
        advanceViewportContinuity,
        clearViewportContinuity,
        continuity: effects.continuity,
        editorPort,
        viewportContinuityToken: options.viewportContinuityToken,
      });
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

  const socketTransactions = useMemo(
    () => createTimelineSocketTransactionAdapter(mutationRuntime),
    [mutationRuntime],
  );
  const trackPendingSocketTxn = socketTransactions.track;
  const resolvePendingSocketTxn = socketTransactions.resolve;

  const enqueueSaveWork = useCallback(
    (work: () => Promise<void>) => {
      const finish = mutationRuntime.beginExplicitMutation();
      pendingSavesRefs.saveQueueRef.current =
        pendingSavesRefs.saveQueueRef.current
          .catch(() => undefined)
          .then(work)
          .finally(finish);
    },
    [mutationRuntime, pendingSavesRefs],
  );

  const reconcileDiscardedPendingUnit = useCallback(
    (
      discardedUnit: PendingReplayUnitState,
      remainingUnits: readonly PendingReplayUnitState[],
      contextByUnitId: ReadonlyMap<string, TimelineReplayContext>,
    ) => {
      const plan = reconcileDiscardedTimelineUnit({
        committedRow:
          discardedUnit.recordId === null
            ? null
            : latestCommittedTimelineRow(discardedUnit.recordId),
        contextByUnitId,
        currentRows: rowsRef.current,
        discardedUnit,
        nextDraftIndex,
        remainingUnits,
      });
      if (
        plan.discardedFocusKey !== null &&
        !plan.remainingFocusKeys.has(plan.discardedFocusKey)
      ) {
        editorDraftRegistry.deleteDraftForFocusKey(plan.discardedFocusKey);
      }
      editorDraftRegistry.clearScalarDraftsForRow(
        discardedUnit.rowKey,
        plan.remainingFocusKeys,
      );
      if (plan.cancelEdit !== null) {
        editorPort.cancelEdit(plan.cancelEdit);
      }
      if (plan.rows !== null) {
        rowsRef.current = [...plan.rows];
        replaceRows(plan.rows);
      }
    },
    [
      editorDraftRegistry,
      editorPort,
      latestCommittedTimelineRow,
      nextDraftIndex,
      rowsRef,
      replaceRows,
    ],
  );

  const conflictProjection = useTimelineConflictProjectionAdapter({
    sheetRef,
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
      currentCommittedTimelineRow,
      enqueueSaveWork,
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
        currentCreatedRowPresentationRecordId,
        currentCommittedTimelineRow,
        currentMutationEpoch,
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
