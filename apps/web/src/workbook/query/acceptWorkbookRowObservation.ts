import type { WorkbookQueryRow, WorkbookReadScope } from "./WorkbookQueryRow";

/** Only a source read/mutation owner may call this, at acceptance of captured evidence. */
export function acceptWorkbookRowObservation<Row extends WorkbookQueryRow>(
  row: Row,
  scope: WorkbookReadScope | null,
): Row & Pick<WorkbookQueryRow, "observation"> {
  if (!scope) return row;
  return {
    ...row,
    observation: {
      recordId: row.record_id,
      rowVersion: row.row_version,
      scope: { ...scope },
    },
  };
}
