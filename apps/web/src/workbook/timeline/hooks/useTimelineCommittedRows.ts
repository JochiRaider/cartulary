import { useCallback, useRef } from "react";
import { useWorkbookHistoryRuntime } from "../../history/WorkbookHistoryContext";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { timelineCaptureOwnerFor } from "../actions/timelineCaptureOwnerFor";
import { createTimelineCommittedVersionLedger } from "../models/timelineCommittedVersionLedger";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { TimelineRecordActionAccepted } from "../ports/TimelineRecordActionPort";

export function useTimelineCommittedRows({
  rowsRef,
  mutationRuntime,
}: {
  readonly mutationRuntime?: WorkbookMutationRuntime | undefined;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
}) {
  const history = useWorkbookHistoryRuntime()?.history;
  const capture = mutationRuntime
    ? timelineCaptureOwnerFor(mutationRuntime)
    : null;
  const ledgerRef = useRef(createTimelineCommittedVersionLedger());
  const hasLoadedRowsRef = useRef(false);
  const loadSequenceRef = useRef(0);

  const knownTimelineRowVersion = useCallback(
    (recordId: string) =>
      Math.max(
        ledgerRef.current.knownVersion(recordId) ?? 0,
        capture?.latestVersion(recordId) ?? 0,
      ) || undefined,
    [capture],
  );

  const currentCommittedTimelineRow = useCallback(
    (recordId: string) => ledgerRef.current.current(recordId, rowsRef.current),
    [rowsRef],
  );

  const acceptCommittedTimelineRow = useCallback(
    (row: WorkbookRow) => {
      if (row.recordId && row.rowVersion !== null) {
        if (row.rowVersion < (knownTimelineRowVersion(row.recordId) ?? 0))
          return {
            accepted: false,
            stale: true,
            row:
              ledgerRef.current.current(row.recordId, rowsRef.current) ?? row,
          };
        history?.acceptVersion(row.recordId, row.rowVersion);
        capture?.acceptVersion(row.recordId, row.rowVersion);
      }
      return ledgerRef.current.accept(row, rowsRef.current);
    },
    [rowsRef, history, capture, knownTimelineRowVersion],
  );

  const acceptCommittedTimelineRows = useCallback(
    (committedRows: readonly WorkbookRow[]) => {
      for (const row of committedRows) acceptCommittedTimelineRow(row);
    },
    [acceptCommittedTimelineRow],
  );

  const isStaleTimelineRowVersion = useCallback(
    (recordId: string, rowVersion: number) =>
      rowVersion < (knownTimelineRowVersion(recordId) ?? 0),
    [knownTimelineRowVersion],
  );

  const acceptTimelineRecordVersion = useCallback(
    (recordId: string, rowVersion: number) => {
      history?.acceptVersion(recordId, rowVersion);
      capture?.acceptVersion(recordId, rowVersion);
      return ledgerRef.current.acceptVersion(
        recordId,
        rowVersion,
        rowsRef.current,
      );
    },
    [rowsRef, history, capture],
  );

  const acceptTimelineActionResult = useCallback(
    (result: TimelineRecordActionAccepted) => {
      if (result.rowVersion < (knownTimelineRowVersion(result.recordId) ?? 0))
        return;
      capture?.acceptVersion(result.recordId, result.rowVersion);
      history?.acceptVersion(result.recordId, result.rowVersion);
      const existing = ledgerRef.current.current(
        result.recordId,
        rowsRef.current,
      );
      if (existing === null) {
        ledgerRef.current.acceptVersion(
          result.recordId,
          result.rowVersion,
          rowsRef.current,
        );
        return;
      }
      ledgerRef.current.accept(
        {
          ...existing,
          rowVersion: result.rowVersion,
          captureState: result.captureState,
          rawRow:
            existing.rawRow === null
              ? null
              : {
                  ...existing.rawRow,
                  row_version: result.rowVersion,
                  cells: {
                    ...existing.rawRow.cells,
                    "timeline.replacement_record_id": {
                      value: result.replacementRecordId,
                    },
                    "timeline.capture_state": {
                      value: result.captureState,
                    },
                  },
                },
        },
        rowsRef.current,
      );
    },
    [rowsRef, history, capture, knownTimelineRowVersion],
  );

  const latestCommittedTimelineRow = useCallback(
    (recordId: string) => ledgerRef.current.latest(recordId, rowsRef.current),
    [rowsRef],
  );

  const latestCommittedRowVersion = useCallback(
    (recordId: string) => {
      const row = latestCommittedTimelineRow(recordId);
      return knownTimelineRowVersion(recordId) ?? row?.rowVersion ?? null;
    },
    [knownTimelineRowVersion, latestCommittedTimelineRow],
  );

  const beginLoad = useCallback(() => {
    const requestSequence = loadSequenceRef.current + 1;
    loadSequenceRef.current = requestSequence;
    return {
      queryStartEpoch: ledgerRef.current.currentEpoch(),
      requestSequence,
    };
  }, []);

  const isCurrentLoadSequence = useCallback(
    (requestSequence: number) => requestSequence === loadSequenceRef.current,
    [],
  );
  const currentMutationEpoch = useCallback(
    () => ledgerRef.current.currentEpoch(),
    [],
  );
  const hasLoadedRows = useCallback(() => hasLoadedRowsRef.current, []);
  const markRowsLoaded = useCallback(() => {
    hasLoadedRowsRef.current = true;
  }, []);

  return {
    commands: {
      acceptCommittedTimelineRow,
      acceptCommittedTimelineRows,
      acceptTimelineActionResult,
      acceptTimelineRecordVersion,
      beginLoad,
      currentMutationEpoch,
      currentCommittedTimelineRow,
      hasLoadedRows,
      isCurrentLoadSequence,
      isStaleTimelineRowVersion,
      knownTimelineRowVersion,
      latestCommittedRowVersion,
      latestCommittedTimelineRow,
      markRowsLoaded,
    },
  };
}
