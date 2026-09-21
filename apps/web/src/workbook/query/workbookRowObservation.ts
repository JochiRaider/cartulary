import type { WorkbookQueryRow, WorkbookReadScope } from "./WorkbookQueryRow";

export function sameWorkbookReadScope(
  left: WorkbookReadScope | null | undefined,
  right: WorkbookReadScope | null | undefined,
): boolean {
  return (
    !!left &&
    !!right &&
    left.actorId === right.actorId &&
    left.sessionIdentity === right.sessionIdentity &&
    left.incidentId === right.incidentId &&
    left.epoch === right.epoch
  );
}
export function workbookRowIsAdmissible(
  row: WorkbookQueryRow,
  recordId: string,
  scope: WorkbookReadScope | null,
): boolean {
  const observation = row.observation;
  return (
    !!observation &&
    row.record_id === recordId &&
    observation.recordId === recordId &&
    Number.isSafeInteger(row.row_version) &&
    row.row_version > 0 &&
    observation.rowVersion === row.row_version &&
    sameWorkbookReadScope(observation.scope, scope)
  );
}
/** Conversions preserve evidence verbatim; they cannot create or relabel it. */
export function preserveWorkbookRowObservation<Row extends WorkbookQueryRow>(
  source: unknown,
  row: Row,
): Row {
  if (!source || typeof source !== "object" || !("observation" in source))
    return row;
  const observation = (source as WorkbookQueryRow).observation;
  if (
    !observation ||
    observation.recordId !== row.record_id ||
    observation.rowVersion !== row.row_version
  )
    return row;
  return { ...row, observation };
}
