type RecordIdentity = {
  readonly recordId: string | null;
  readonly rowVersion?: number | null;
};

export function reconcileWorkbookRecordRows<Row extends RecordIdentity>(
  previousRows: readonly Row[],
  nextRows: readonly Row[],
): readonly Row[] {
  const previousByRecordId = new Map<string, Row>();
  for (const row of previousRows) {
    if (row.recordId !== null && row.recordId.trim() !== "") {
      previousByRecordId.set(row.recordId, row);
    }
  }
  return nextRows.map((row) => {
    if (row.recordId === null || row.recordId.trim() === "") {
      return row;
    }
    const previous = previousByRecordId.get(row.recordId);
    if (previous === undefined) return row;
    if (
      typeof previous.rowVersion === "number" &&
      previous.rowVersion === row.rowVersion
    ) {
      return previous;
    }
    return shallowEqualRecord(previous, row) ? previous : row;
  });
}

function shallowEqualRecord<Row extends object>(left: Row, right: Row) {
  const leftRecord = left as Record<string, unknown>;
  const rightRecord = right as Record<string, unknown>;
  const leftKeys = Object.keys(leftRecord);
  const rightKeys = Object.keys(rightRecord);
  return (
    leftKeys.length === rightKeys.length &&
    leftKeys.every((key) => Object.is(leftRecord[key], rightRecord[key]))
  );
}
