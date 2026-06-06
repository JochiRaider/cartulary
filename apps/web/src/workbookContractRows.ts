import {
  type GridColumn,
} from "@cartulary/grid-adapter";
import {
  gridSortHeaderTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import {
  normalizeViewRowV1,
  resolveHeaderSortFieldKey,
  type NormalizedViewRowV1,
  type ViewContract,
  type ViewFieldContract,
  visibleFields,
} from "@cartulary/view-contracts";

export type WorkbookViewApiCell = {
  value: unknown;
};

export type WorkbookViewApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, WorkbookViewApiCell>;
  group_values?: Record<string, unknown>;
  view_schema_id?: string;
};

function materializeViewRowCells(
  cells: NormalizedViewRowV1["cells"],
): Record<string, WorkbookViewApiCell> {
  return Object.fromEntries(
    Object.entries(cells).map(([fieldKey, cell]) => [
      fieldKey,
      { value: cell.value },
    ]),
  );
}

export function materializeWorkbookViewRow(
  normalized: NormalizedViewRowV1,
): WorkbookViewApiRow {
  return {
    record_id: normalized.recordId,
    row_version: normalized.rowVersion,
    view_schema_id: normalized.viewSchemaId,
    cells: materializeViewRowCells(normalized.cells),
    ...(normalized.groupValues === undefined
      ? {}
      : { group_values: { ...normalized.groupValues } }),
  };
}

export function normalizeWorkbookViewRows(
  contract: ViewContract,
  rows: readonly unknown[],
  source: string,
): WorkbookViewApiRow[] {
  return rows.map((row, index) =>
    materializeWorkbookViewRow(
      normalizeViewRowV1(contract, row, `${source} rows[${index}]`),
    ),
  );
}

export function workbookContractColumns<Row>({
  contract,
  surface,
  widthForField,
}: {
  readonly contract: ViewContract;
  readonly surface: WorkbookSurface;
  readonly widthForField?: (field: ViewFieldContract) => number;
}): readonly GridColumn<Row>[] {
  return visibleFields(contract).map((field) => ({
    fieldKey: field.fieldKey,
    headerTestId: gridSortHeaderTestId(surface, field.fieldKey),
    label: field.label,
    width: widthForField?.(field) ?? 220,
    renderCell: () => null,
    sortableFieldKey: resolveHeaderSortFieldKey(contract, field.fieldKey),
  }));
}
