import type { normalizeViewRowPatchV1 } from "@cartulary/view-contracts";
import { acceptWorkbookRowObservation } from "./acceptWorkbookRowObservation";
import type { WorkbookQueryRow, WorkbookReadScope } from "./WorkbookQueryRow";
import { workbookRowIsAdmissible } from "./workbookRowObservation";

export function applyWorkbookQueryRowPatch(
  row: WorkbookQueryRow,
  patch: ReturnType<typeof normalizeViewRowPatchV1>,
  scope: WorkbookReadScope | null = null,
): WorkbookQueryRow {
  const next = {
    ...row,
    row_version: patch.rowVersion,
    cells: { ...row.cells, ...patch.cells },
    ...(patch.groupValues === undefined
      ? {}
      : {
          group_values: {
            ...(row.group_values ?? {}),
            ...patch.groupValues,
          },
        }),
  };
  return workbookRowIsAdmissible(row, row.record_id, scope)
    ? acceptWorkbookRowObservation(next, scope)
    : next;
}
