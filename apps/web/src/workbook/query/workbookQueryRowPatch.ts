import type { normalizeViewRowPatchV1 } from "@cartulary/view-contracts";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";

export function applyWorkbookQueryRowPatch(
  row: WorkbookQueryRow,
  patch: ReturnType<typeof normalizeViewRowPatchV1>,
): WorkbookQueryRow {
  return {
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
}
