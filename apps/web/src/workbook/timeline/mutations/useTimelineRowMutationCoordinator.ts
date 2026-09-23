import { useCallback, useEffect, useMemo, useRef } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import {
  useWorkbookRecoveryActivation,
  useWorkbookRecoveryNavigation,
} from "../../../shared/WorkbookRecoveryBoundary";
import { workbookConflictRecoveryKey } from "../../../shared/workbookRecoveryNavigation";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import { useWorkbookMutationConflicts } from "../../runtime/useWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { createTimelineSocketTransactionAdapter } from "../adapters/createTimelineSocketTransactionAdapter";
import { commitTimelineProjection } from "../adapters/timelineProjectionCommitAdapter";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { useTimelineConflictProjectionAdapter } from "../hooks/useTimelineConflictProjectionAdapter";
import { useTimelineConflicts } from "../hooks/useTimelineConflicts";
import { useTimelineSaveStatePresentation } from "../hooks/useTimelineSaveStatePresentation";
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
  TimelineRowMutationEditorPort,
  TimelineRowStoreCommands,
} from "../models/timelineControllerPorts";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import {
  rowFromApi,
  type TimelineApiRow,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../models/timelineRowModel";

type TimelineMutationApplyOptions = {
  readonly continueOnFreshDraft?: boolean;
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

/**
 * Single Timeline owner for row-version admission and transitions between
 * query, fresh/replayed mutations, live patches, conflicts, and continuity.
 */
export function useTimelineRowMutationCoordinator({
  committedRows,
  sheetRef,
  completeAcceptedViewportContinuity,
  createdRowPresentationScopeKey,
  editorDraftRegistry,
  editorPort,
  mutationRuntime,
  nextDraftIndex,
  pendingSavesRefs,
  rowsRef,
  selectedRowId,
  rowStoreCommands,
  setSelectedRowId,
}: {
  readonly committedRows: ReturnType<
    typeof useTimelineCommittedRows
  >["commands"];
  readonly completeAcceptedViewportContinuity: (
    token: number | undefined,
    continuity: TimelineAcceptedContinuity,
  ) => void;
  readonly createdRowPresentationScopeKey: string;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly editorPort: TimelineRowMutationEditorPort;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly sheetRef: SheetRef;
  readonly nextDraftIndex: () => number;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly selectedRowId: string | null;
  readonly rowStoreCommands: TimelineRowStoreCommands;
  readonly setSelectedRowId: (recordId: string | null) => void;
}) {
  const { updateRows } = rowStoreCommands;
  const selectedRowIdRef = useRef(selectedRowId);
  selectedRowIdRef.current = selectedRowId;
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
  const commonConflicts = useWorkbookMutationConflicts(
    mutationRuntime.statusSource,
    timelineViewSchemaId,
  );
  const { activeConflictKey, conflictQueue } = conflicts.snapshot;
  const { setActiveConflictKey, setConflictQueueState } = conflicts.commands;

  useEffect(() => {
    const commonKeys = new Set(
      commonConflicts
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
  }, [commonConflicts, setActiveConflictKey, setConflictQueueState]);

  const saveState = useTimelineSaveStatePresentation({
    sheetRef,
    mutationRuntime,
    pendingSavesRefs,
  });
  const { beginRefreshInFlight, publishSaveStatePresentation } =
    saveState.commands;

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
  } = committedRows;

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
      if (
        editorDraftRegistry.deferWhileComposing(rowKey, () => {
          if (mountedRef.current)
            applyAcceptedRowMutation(rowKey, mutation, options);
        })
      )
        return committed;
      const captureEditor = editorDraftRegistry.captureEditor(
        rowKey,
        committed,
      );
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
      }, true);
      if (projection === undefined) {
        throw new Error("Timeline accepted projection was not committed.");
      }
      const effects = planTimelineAcceptedMutationEffects({
        committed,
        continueOnFreshDraft: options.continueOnFreshDraft === true,
        projection,
        promoteToCommittedRowInspect:
          options.promoteToCommittedRowInspect === true,
        selectedRowId: selectedRowIdRef.current,
      });
      if (effects.selectionUpdate !== null) {
        setSelectedRowId(effects.selectionUpdate.recordId);
      }
      if (effects.createdRecordId !== null) {
        createdRowPresentationRef.current = {
          recordId: effects.createdRecordId,
          scopeKey: createdRowPresentationScopeKey,
        };
      }
      if (
        captureEditor !== null &&
        committed.recordId !== null &&
        options.continueOnFreshDraft !== true
      ) {
        const recordId = committed.recordId;
        // Finish the identity handoff before a waiting Enter/Tab completion
        // applies newer navigation, so a late editor mount cannot steal focus.
        commitTimelineProjection(() => {
          if (captureEditor.collectionFocusKey) {
            const input = editorDraftRegistry.inputElementForFocusKey(
              captureEditor.collectionFocusKey,
            );
            if (input && !input.readOnly) {
              input.focus({ preventScroll: true });
              input.setSelectionRange(
                captureEditor.selectionRange.start,
                captureEditor.selectionRange.end,
              );
            }
          } else editorPort.activateEdit({ ...captureEditor, recordId });
        }, true);
      }
      completeAcceptedViewportContinuity(
        options.viewportContinuityToken,
        effects.continuity,
      );
      recordWorkbookTiming("apply_row_mutation_end", { kind: "row_mutation" });
      return committed;
    },
    [
      acceptCommittedTimelineRow,
      completeAcceptedViewportContinuity,
      createdRowPresentationScopeKey,
      editorPort,
      editorDraftRegistry,
      nextDraftIndex,
      rowsRef,
      updateRows,
      setSelectedRowId,
    ],
  );

  const applyAcceptedBatchRows = useCallback(
    (rows: readonly TimelineApiRow[]) => {
      const committed = rows.map(
        (row) => acceptCommittedTimelineRow(rowFromApi(row)).row,
      );
      if (!mountedRef.current || committed.length === 0) return;
      // One receipt is one presentation commit. Per-row flushes can exhaust
      // React's nested update limit for a virtualized selected rectangle.
      commitTimelineProjection(() => {
        updateRows((current) => {
          let projected = current;
          for (const row of committed) {
            const projection = projectAcceptedTimelineRow({
              committed: row,
              currentRows: projected,
              nextDraftIndex,
              rowKey: row.key,
            });
            projected = projection.rows;
          }
          rowsRef.current = projected;
          return projected;
        });
      }, true);
    },
    [acceptCommittedTimelineRow, nextDraftIndex, rowsRef, updateRows],
  );

  const socketTransactions = useMemo(
    () => createTimelineSocketTransactionAdapter(mutationRuntime),
    [mutationRuntime],
  );
  const trackPendingSocketTxn = socketTransactions.track;
  const resolvePendingSocketTxn = socketTransactions.resolve;

  const enqueueOrderedRead = useCallback(
    (work: () => Promise<void>) => {
      pendingSavesRefs.saveQueueRef.current =
        pendingSavesRefs.saveQueueRef.current.catch(() => undefined).then(work);
    },
    [pendingSavesRefs],
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

  const recoveryNavigation = useWorkbookRecoveryNavigation();
  const activateRecovery = useWorkbookRecoveryActivation();
  const activateConflict = useCallback(
    (key: string | null) => {
      if (key !== null) {
        const conflict = mutationRuntime
          .getSnapshot()
          .conflicts.find((entry) => entry.key === key);
        activateRecovery(
          workbookConflictRecoveryKey(
            recoveryNavigation?.getSnapshot().entries ?? [],
            conflict ?? { key },
          ),
        );
      }
      setActiveConflictKey(key);
    },
    [
      mutationRuntime,
      setActiveConflictKey,
      recoveryNavigation,
      activateRecovery,
    ],
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
      applyAcceptedBatchRows,
      beginRefreshInFlight,
      currentCommittedTimelineRow,
      enqueueOrderedRead,
      hasLoadedRows,
      isCurrentLoadSequence,
      knownTimelineRowVersion,
      latestCommittedRowVersion,
      latestCommittedTimelineRow,
      markRowsLoaded,
      publishSaveStatePresentation,
      registerSameFieldConflict,
      resolvePendingSocketTxn,
      setActiveConflictKey,
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
      commonConflicts,
      conflictQueue,
    },
  };
}
