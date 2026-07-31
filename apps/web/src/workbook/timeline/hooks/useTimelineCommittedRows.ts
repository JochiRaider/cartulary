import { useCallback, useRef } from "react";
import { decideWorkbookRecordFreshness } from "../models/workbookRecordFreshness";
import type { RowValues, WorkbookRow } from "../models/workbookTimelineModel";
import type { TimelineRecordActionAccepted } from "../ports/TimelineRecordActionPort";

function committedProjection<Row extends WorkbookRow>(row: Row): Row {
  const scalarValuesAreCommitted = Object.entries(row.values).every(
    ([field, value]) => row.committedValues[field as keyof RowValues] === value,
  );
  if (
    row.pendingSignature === null &&
    scalarValuesAreCommitted &&
    row.collectionDrafts.hostRefs === "" &&
    row.collectionDrafts.identityRefs === "" &&
    row.collectionDrafts.tags === ""
  ) {
    return row;
  }
  return {
    ...row,
    collectionDrafts: { hostRefs: "", identityRefs: "", tags: "" },
    pendingSignature: null,
    values: { ...row.committedValues },
  } as Row;
}

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
    (recordId: string) => {
      const cached = committedTimelineRowsRef.current.get(recordId);
      if (cached !== undefined) return cached;
      const visible = rowsRef.current.find(
        (candidate) =>
          candidate.recordId === recordId && candidate.rowVersion !== null,
      );
      return visible === undefined ? null : committedProjection(visible);
    },
    [rowsRef],
  );

  const acceptCommittedTimelineRow = useCallback(
    (
      row: WorkbookRow,
    ): { row: WorkbookRow; accepted: boolean; stale: boolean } => {
      if (row.recordId === null || row.rowVersion === null) {
        return { row, accepted: false, stale: false };
      }
      const recordId = row.recordId;
      const rowVersion = row.rowVersion;
      const committed = committedProjection(row);
      const currentVersion = knownTimelineRowVersion(recordId);
      if (decideWorkbookRecordFreshness(committed, currentVersion).stale) {
        return {
          row: currentCommittedTimelineRow(recordId) ?? committed,
          accepted: false,
          stale: true,
        };
      }
      if (currentVersion !== rowVersion) {
        committedTimelineRowsEpochRef.current += 1;
      }
      committedTimelineRowVersionsRef.current.set(recordId, rowVersion);
      committedTimelineRowsRef.current.set(recordId, committed);
      return { row: committed, accepted: true, stale: false };
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
    (result: TimelineRecordActionAccepted) => {
      const existing =
        committedTimelineRowsRef.current.get(result.recordId) ??
        rowsRef.current.find((row) => row.recordId === result.recordId) ??
        null;
      if (existing === null) {
        acceptTimelineRecordVersion(result.recordId, result.rowVersion);
        return;
      }
      acceptCommittedTimelineRow({
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
        return acceptCommittedTimelineRow(visibleRow).row;
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
