import type {
  CSSProperties,
  MouseEventHandler,
  PropsWithChildren,
  ReactNode,
} from "react";

export type GridSortDirection = "asc" | "desc";

export type GridSortEntry = {
  readonly fieldKey: string;
  readonly direction: GridSortDirection;
};

export type GridColumn<Row> = {
  readonly fieldKey: string;
  readonly headerTestId?: string | undefined;
  readonly label: string;
  readonly renderCell: (row: Row) => ReactNode;
  readonly sortableFieldKey?: string | null;
  readonly sortDisabled?: boolean | undefined;
  readonly sortDisabledReason?: string | null | undefined;
  readonly align?: "left" | "center" | "right" | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridRow<Row> = {
  readonly key: string;
  readonly recordId: string | null;
  readonly data: Row;
  readonly onSelect?: MouseEventHandler<HTMLTableRowElement> | undefined;
  readonly selected?: boolean | undefined;
  readonly variant?: "default" | "draft" | undefined;
  readonly testId?: string | undefined;
};

export type GridActionsColumn<Row> = {
  readonly label: string;
  readonly renderCell: (row: GridRow<Row>) => ReactNode;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridViewportProps = PropsWithChildren<{
  readonly className?: string | undefined;
  readonly style?: CSSProperties | undefined;
  readonly testId?: string | undefined;
}>;

export type GridTableProps<Row> = {
  readonly actionsColumn?: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly emptyMessage?: ReactNode | undefined;
  readonly getGroupLabel?: (
    row: Row,
    fieldKey: string,
  ) => string | null | undefined;
  readonly getGroupRowTestId?: (
    fieldKey: string,
    value: string,
  ) => string | undefined;
  readonly groupBy?: string | null | undefined;
  readonly onToggleSort?: ((fieldKey: string) => void) | undefined;
  readonly rows: readonly GridRow<Row>[];
  readonly sort?: readonly GridSortEntry[] | undefined;
};

export type RecordIdentity = {
  readonly recordId: string | null;
};

export function assertGridRows<Row extends RecordIdentity>(
  rows: readonly Row[],
) {
  const seen = new Set<string>();
  for (const row of rows) {
    if (row.recordId === null) {
      continue;
    }
    const normalized = row.recordId.trim();
    if (normalized === "") {
      throw new Error(
        "Grid adapter invariant failed: missing record_id on a saved row.",
      );
    }
    if (seen.has(normalized)) {
      throw new Error(
        `Grid adapter invariant failed: duplicate record_id "${normalized}".`,
      );
    }
    seen.add(normalized);
  }
}

export function reconcileRecordRows<Row extends RecordIdentity>(
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
    if (previous === undefined) {
      return row;
    }
    if (shallowEqualRecord(previous, row)) {
      return previous;
    }
    return row;
  });
}

function shallowEqualRecord<Row extends object>(left: Row, right: Row) {
  const leftRecord = left as Record<string, unknown>;
  const rightRecord = right as Record<string, unknown>;
  const leftKeys = Object.keys(leftRecord);
  const rightKeys = Object.keys(rightRecord);
  if (leftKeys.length !== rightKeys.length) {
    return false;
  }
  return leftKeys.every((key) => Object.is(leftRecord[key], rightRecord[key]));
}
