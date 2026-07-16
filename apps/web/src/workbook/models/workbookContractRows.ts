import type { GridColumn, GridDataRow } from "@cartulary/grid-adapter";
import {
  gridRowTestId,
  gridSortHeaderTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import {
  type NormalizedViewRowV1,
  normalizeViewRowV1,
  resolveHeaderSortFieldKey,
  type ViewContract,
  type ViewFieldContract,
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

function materializeWorkbookViewRow(
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

export function workbookGridRows<Row>({
  getRecordId,
  getRowVersion,
  rows,
  surface,
}: {
  readonly getRecordId: (row: Row) => string;
  readonly getRowVersion: (row: Row) => number;
  readonly rows: readonly Row[];
  readonly surface: WorkbookSurface;
}): readonly GridDataRow<Row>[] {
  return rows.map((row) => {
    const recordId = getRecordId(row);
    return {
      kind: "data" as const,
      mutationIdentity: {
        kind: "core_row_version" as const,
        baseRowVersion: getRowVersion(row),
      },
      rowIdentity: { kind: "core_record" as const, recordId },
      data: row,
      testId: gridRowTestId(surface, recordId),
    };
  });
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
  return contract.fields.map((field) => ({
    fieldKey: field.fieldKey,
    headerTestId: gridSortHeaderTestId(surface, field.fieldKey),
    label: field.label,
    width: widthForField?.(field) ?? 220,
    renderCell: () => null,
    sortableFieldKey: resolveHeaderSortFieldKey(contract, field.fieldKey),
  }));
}
