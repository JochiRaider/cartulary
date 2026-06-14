export type WorkbookVersionedRecord = {
  readonly recordId: string | null;
  readonly rowVersion: number | null;
};

export type WorkbookRecordFreshnessDecision = {
  readonly comparable: boolean;
  readonly stale: boolean;
};

export function decideWorkbookRecordFreshness(
  incoming: WorkbookVersionedRecord,
  knownRowVersion: number | null | undefined,
): WorkbookRecordFreshnessDecision {
  if (
    incoming.recordId === null ||
    incoming.rowVersion === null ||
    knownRowVersion === null ||
    knownRowVersion === undefined
  ) {
    return {
      comparable: false,
      stale: false,
    };
  }
  return {
    comparable: true,
    stale: incoming.rowVersion < knownRowVersion,
  };
}
