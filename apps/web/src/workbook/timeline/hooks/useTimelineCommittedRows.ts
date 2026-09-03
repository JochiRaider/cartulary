import { useCallback, useRef } from "react";
import { createTimelineCommittedVersionLedger } from "../models/timelineCommittedVersionLedger";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { TimelineRecordActionAccepted } from "../ports/TimelineRecordActionPort";

export function useTimelineCommittedRows({
  rowsRef,
}: {
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
}) {
  const ledgerRef = useRef(createTimelineCommittedVersionLedger());
  const hasLoadedRowsRef = useRef(false);
  const loadSequenceRef = useRef(0);

  const knownTimelineRowVersion = useCallback(
    (recordId: string) => ledgerRef.current.knownVersion(recordId),
    [],
  );

  const currentCommittedTimelineRow = useCallback(
    (recordId: string) => ledgerRef.current.current(recordId, rowsRef.current),
    [rowsRef],
  );

  const acceptCommittedTimelineRow = useCallback(
    (row: WorkbookRow) => ledgerRef.current.accept(row, rowsRef.current),
    [rowsRef],
  );

  const acceptCommittedTimelineRows = useCallback(
    (committedRows: readonly WorkbookRow[]) => {
      for (const row of committedRows) acceptCommittedTimelineRow(row);
    },
    [acceptCommittedTimelineRow],
  );

  const isStaleTimelineRowVersion = useCallback(
    (recordId: string, rowVersion: number) =>
      ledgerRef.current.isStale(recordId, rowVersion),
    [],
  );

  const acceptTimelineRecordVersion = useCallback(
    (recordId: string, rowVersion: number) =>
      ledgerRef.current.acceptVersion(recordId, rowVersion, rowsRef.current),
    [rowsRef],
  );

  const acceptTimelineActionResult = useCallback(
    (result: TimelineRecordActionAccepted) => {
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
                    "timeline.capture_state": {
                      value: result.captureState,
                    },
                  },
                },
        },
        rowsRef.current,
      );
    },
    [rowsRef],
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
