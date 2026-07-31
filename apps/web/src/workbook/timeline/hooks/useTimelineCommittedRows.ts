import { useCallback, useRef } from "react";
import { decideWorkbookRecordFreshness } from "../models/workbookRecordFreshness";
import type { WorkbookRow } from "../models/workbookTimelineModel";

type TimelineActionResultRowVersion = {
  readonly capture_state: WorkbookRow["captureState"];
  readonly record_id: string;
  readonly row_version: number;
};

export function useTimelineCommittedRows({
  rowsRef,
}: {
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
}) {
  const committedTimelineRowsRef = useRef(new Map<string, WorkbookRow>());
  const committedTimelineRowVersionsRef = useRef(new Map<string, number>());
  const committedTimelineRowsEpochRef = useRef(0);
  const hasLoadedRowsRef = useRef(false);
  const loadSequenceRef = useRef(0);

  const knownTimelineRowVersion = useCallback(
    (recordId: string) => committedTimelineRowVersionsRef.current.get(recordId),
    [],
  );

  const currentCommittedTimelineRow = useCallback(
    (recordId: string) =>
      committedTimelineRowsRef.current.get(recordId) ??
      rowsRef.current.find(
        (candidate) =>
          candidate.recordId === recordId && candidate.rowVersion !== null,
      ) ??
      null,
    [rowsRef],
  );

  const acceptCommittedTimelineRow = useCallback(
    (
      row: WorkbookRow,
    ): { row: WorkbookRow; accepted: boolean; stale: boolean } => {
      if (row.recordId === null || row.rowVersion === null) {
        return { row, accepted: false, stale: false };
      }
      const currentVersion = knownTimelineRowVersion(row.recordId);
      if (decideWorkbookRecordFreshness(row, currentVersion).stale) {
        return {
          row: currentCommittedTimelineRow(row.recordId) ?? row,
          accepted: false,
          stale: true,
        };
      }
      if (currentVersion !== row.rowVersion) {
        committedTimelineRowsEpochRef.current += 1;
      }
      committedTimelineRowVersionsRef.current.set(row.recordId, row.rowVersion);
      committedTimelineRowsRef.current.set(row.recordId, row);
      return { row, accepted: true, stale: false };
    },
    [currentCommittedTimelineRow, knownTimelineRowVersion],
  );

  const acceptCommittedTimelineRows = useCallback(
    (committedRows: readonly WorkbookRow[]) => {
      for (const row of committedRows) {
        acceptCommittedTimelineRow(row);
      }
    },
    [acceptCommittedTimelineRow],
  );

  const isStaleTimelineRowVersion = useCallback(
    (recordId: string, rowVersion: number) =>
      decideWorkbookRecordFreshness(
        { recordId, rowVersion },
        knownTimelineRowVersion(recordId),
      ).stale,
    [knownTimelineRowVersion],
  );

  const acceptTimelineRecordVersion = useCallback(
    (recordId: string, rowVersion: number) => {
      if (isStaleTimelineRowVersion(recordId, rowVersion)) {
        return { accepted: false, stale: true };
      }
      const existing = currentCommittedTimelineRow(recordId);
      if (existing === null) {
        const currentVersion = knownTimelineRowVersion(recordId);
        if (currentVersion !== rowVersion) {
          committedTimelineRowsEpochRef.current += 1;
        }
        committedTimelineRowVersionsRef.current.set(recordId, rowVersion);
        return { accepted: true, stale: false };
      }
      const accepted = acceptCommittedTimelineRow({
        ...existing,
        rowVersion,
        rawRow:
          existing.rawRow === null
            ? null
            : {
                ...existing.rawRow,
                row_version: rowVersion,
              },
      });
      return { accepted: accepted.accepted, stale: accepted.stale };
    },
    [
      acceptCommittedTimelineRow,
      currentCommittedTimelineRow,
      isStaleTimelineRowVersion,
      knownTimelineRowVersion,
    ],
  );

  const acceptTimelineActionResult = useCallback(
    (result: TimelineActionResultRowVersion) => {
      const existing =
        committedTimelineRowsRef.current.get(result.record_id) ??
        rowsRef.current.find((row) => row.recordId === result.record_id) ??
        null;
      if (existing === null) {
        acceptTimelineRecordVersion(result.record_id, result.row_version);
        return;
      }
      acceptCommittedTimelineRow({
        ...existing,
        rowVersion: result.row_version,
        captureState: result.capture_state,
        rawRow:
          existing.rawRow === null
            ? null
            : {
                ...existing.rawRow,
                row_version: result.row_version,
                cells: {
                  ...existing.rawRow.cells,
                  "timeline.capture_state": {
                    value: result.capture_state,
                  },
                },
              },
      });
    },
    [acceptCommittedTimelineRow, acceptTimelineRecordVersion, rowsRef],
  );

  const latestCommittedTimelineRow = useCallback(
    (recordId: string) => {
      const visibleRow = rowsRef.current.find(
        (candidate) => candidate.recordId === recordId,
      );
      const knownVersion = knownTimelineRowVersion(recordId);
      if (
        visibleRow !== undefined &&
        visibleRow.rowVersion !== null &&
        (knownVersion === undefined || visibleRow.rowVersion >= knownVersion)
      ) {
        acceptCommittedTimelineRow(visibleRow);
        return visibleRow;
      }

      const committedRow = committedTimelineRowsRef.current.get(recordId);
      if (
        committedRow !== undefined &&
        committedRow.rowVersion !== null &&
        (knownVersion === undefined || committedRow.rowVersion >= knownVersion)
      ) {
        return committedRow;
      }
      return null;
    },
    [acceptCommittedTimelineRow, knownTimelineRowVersion, rowsRef],
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
      queryStartEpoch: committedTimelineRowsEpochRef.current,
      requestSequence,
    };
  }, []);

  const isCurrentLoadSequence = useCallback(
    (requestSequence: number) => requestSequence === loadSequenceRef.current,
    [],
  );

  const committedRowsChangedSince = useCallback(
    (queryStartEpoch: number) =>
      queryStartEpoch !== committedTimelineRowsEpochRef.current,
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
      committedRowsChangedSince,
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
