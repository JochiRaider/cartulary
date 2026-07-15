import {
  gridScrollportClassName,
  workbookGridRowHeightPx,
} from "@cartulary/ui-contracts";
import type {
  CSSProperties,
  ForwardedRef,
  ReactElement,
  RefAttributes,
} from "react";
import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  type CalculatedColumn,
  type Column,
  type ColumnWidths,
  DataGrid,
  type DataGridHandle,
  type DataGridProps,
  type RenderRowProps,
  Row,
  type SortColumn,
  TreeDataGrid,
} from "react-data-grid";
import {
  assertGridRows,
  type GridCellAnchor,
  type GridDraftRow,
  type GridGroupingScalar,
  type GridHandle,
  type GridRecordRow,
  type GridSortEntry,
  gridUnassignedGroupLabel,
  type WorkbookDataGridProps,
} from "./core";
import { compileGridColumns, gridRowGutterColumnKey } from "./rdgCompiler";

type GroupMetadata = {
  readonly label: string | null;
  readonly value: GridGroupingScalar;
};

// RDG exposes row and column virtualization through one switch. Query-sized
// workbooks keep every semantic field addressable by the stable selector
// facade; large result sets cross this boundary and receive bounded, fixed-row
// virtualization. The threshold is presentation-only and never affects query
// ownership, ordering, identity, or mutation targets.
const gridVirtualizationRowThreshold = 100;

function useWorkbookDataGrid<Row>(
  props: WorkbookDataGridProps<Row>,
  ref: ForwardedRef<GridHandle>,
) {
  const {
    actionsColumn,
    columns,
    columnWidths,
    density = "default",
    draftRow,
    emptyMessage = "No rows",
    fillViewportInline = false,
    grouping = null,
    onActiveCellChange,
    onColumnWidthChange,
    onCopyCell,
    onEditCell,
    onFillCells,
    onPasteCell,
    onSelectRecord,
    onSelectedRecordIdsChange,
    onSortChange,
    onToggleSort,
    recordRows,
    rowGutter,
    selectedRecordIds,
    sort = [],
    viewSchemaId,
  } = props;
  const vendorHandle = useRef<DataGridHandle>(null);
  assertGridRows(recordRows);

  const compiledColumns = useMemo(
    () =>
      compileGridColumns({
        actionsColumn,
        columns,
        rowGutter,
        viewSchemaId,
      }),
    [actionsColumn, columns, rowGutter, viewSchemaId],
  );
  const selectedRows = useMemo(
    () =>
      selectedRecordIds ??
      new Set(
        recordRows
          .filter((row) => row.selected === true)
          .map((row) => row.recordId),
      ),
    [recordRows, selectedRecordIds],
  );
  const rdgColumnWidths = useMemo<ColumnWidths | undefined>(
    () =>
      columnWidths === undefined
        ? undefined
        : new Map(
            Object.entries(columnWidths).map(([fieldKey, width]) => [
              fieldKey,
              { type: "resized" as const, width },
            ]),
          ),
    [columnWidths],
  );
  const sortColumns = useMemo(
    () =>
      sort.map((entry) => ({
        columnKey: entry.fieldKey,
        direction:
          entry.direction === "asc" ? ("ASC" as const) : ("DESC" as const),
      })),
    [sort],
  );
  useImperativeHandle(
    ref,
    () => ({
      focusAnchor: (anchor) =>
        focusOrScrollAnchor({
          anchor,
          columns: compiledColumns,
          grouping,
          recordRows,
          vendorHandle: vendorHandle.current,
          viewSchemaId,
          focus: true,
        }),
      getScrollElement: () => vendorHandle.current?.element ?? null,
      scrollToAnchor: (anchor) =>
        focusOrScrollAnchor({
          anchor,
          columns: compiledColumns,
          grouping,
          recordRows,
          vendorHandle: vendorHandle.current,
          viewSchemaId,
          focus: false,
        }),
    }),
    [compiledColumns, grouping, recordRows, viewSchemaId],
  );
  const sharedProps: DataGridProps<
    GridRecordRow<Row>,
    GridDraftRow<Row>,
    string
  > = {
    bottomSummaryRows: draftRow === undefined ? undefined : [draftRow],
    className: `${gridScrollportClassName()} cartulary-grid rdg-dark`,
    columnWidths: rdgColumnWidths,
    columns: compiledColumns,
    enableVirtualization: recordRows.length > gridVirtualizationRowThreshold,
    headerRowHeight: 32,
    onCellClick: ({ row }) => {
      onSelectRecord?.(row.recordId);
    },
    onCellCopy: ({ column, row }) => {
      const anchor = semanticAnchor(row, column.key, columns, viewSchemaId);
      if (anchor !== null) onCopyCell?.({ anchor });
    },
    onCellDoubleClick: ({ column, row }) => {
      const target = semanticTarget(row, column.key, columns, viewSchemaId);
      if (target !== null) onEditCell?.({ target });
    },
    onCellKeyDown: (_args, event) => {
      if (event.key === "Tab") {
        event.preventGridDefault();
      }
    },
    onCellPaste: ({ column, row }) => {
      const target = semanticTarget(row, column.key, columns, viewSchemaId);
      if (target !== null) onPasteCell?.({ target });
      return row;
    },
    onColumnResize: (
      column: CalculatedColumn<GridRecordRow<Row>, GridDraftRow<Row>>,
      width: number,
    ) => {
      onColumnWidthChange?.(column.key, width);
    },
    onFill:
      grouping === null
        ? ({ columnKey, sourceRow, targetRow }) => {
            const source = semanticTarget(
              sourceRow,
              columnKey,
              columns,
              viewSchemaId,
            );
            const target = semanticTarget(
              targetRow,
              columnKey,
              columns,
              viewSchemaId,
            );
            if (source !== null && target !== null) {
              onFillCells?.({ source, target });
            }
            return targetRow;
          }
        : undefined,
    onSelectedRowsChange: (next: Set<string>) => {
      onSelectedRecordIdsChange?.(next);
    },
    onRowsChange: onPasteCell === undefined ? undefined : () => {},
    onSelectedCellChange: ({ column, row }) => {
      onActiveCellChange?.(
        row === undefined
          ? null
          : semanticAnchor(row, column.key, columns, viewSchemaId),
      );
    },
    onSortColumnsChange: (next: SortColumn[]) => {
      const semanticSort: readonly GridSortEntry[] = next.map((entry) => ({
        direction: entry.direction === "ASC" ? "asc" : "desc",
        fieldKey: entry.columnKey,
      }));
      if (onSortChange !== undefined) {
        onSortChange(semanticSort);
        return;
      }
      const changed = semanticSort.find(
        (entry, index) =>
          entry.fieldKey !== sort[index]?.fieldKey ||
          entry.direction !== sort[index]?.direction,
      );
      if (changed !== undefined) {
        onToggleSort?.(changed.fieldKey);
      } else if (next.length === 0 && sort[0] !== undefined) {
        onToggleSort?.(sort[0].fieldKey);
      }
    },
    renderers: {
      // RDG 7 suppresses summary rows whenever noRowsFallback is present and
      // the committed row collection is empty. A draft is a real bottom
      // summary affordance, so reserve the fallback for truly empty grids.
      noRowsFallback:
        draftRow === undefined ? (
          <div className="cartulary-grid-empty" data-grid-empty="true">
            {emptyMessage}
          </div>
        ) : undefined,
      renderRow: (
        key,
        rowProps: RenderRowProps<GridRecordRow<Row>, GridDraftRow<Row>>,
      ) => (
        <Row
          {...rowProps}
          data-cartulary-grid-row-kind="record"
          data-grid-record-id={rowProps.row.recordId}
          data-testid={rowProps.row.testId}
          key={key}
        />
      ),
    },
    rowKeyGetter: (row: GridRecordRow<Row>) => row.recordId,
    ref: vendorHandle,
    rows: recordRows,
    selectedRows,
    sortColumns,
    style: {
      "--cartulary-grid-cell-padding": `var(--ct-density-${density}-cellPadding)`,
      "--cartulary-grid-density": density,
      "--cartulary-grid-font-size": `var(--ct-density-${density}-fontSize)`,
      "--cartulary-grid-line-height": `var(--ct-density-${density}-lineHeight)`,
      "--cartulary-grid-row-height": `var(--ct-density-${density}-rowHeight)`,
      blockSize: "100%",
      minWidth: fillViewportInline ? 0 : 1248,
      overflow: "auto",
      width: fillViewportInline ? "100%" : undefined,
    } as CSSProperties,
    summaryRowHeight: workbookGridRowHeightPx(density),
  };

  if (grouping === null) {
    return (
      <DataGrid {...sharedProps} rowHeight={workbookGridRowHeightPx(density)} />
    );
  }

  return (
    <GroupedWorkbookDataGrid
      {...props}
      density={density}
      sharedProps={sharedProps}
    />
  );
}

function WorkbookDataGridInner<Row>(
  props: WorkbookDataGridProps<Row>,
  ref: ForwardedRef<GridHandle>,
) {
  // biome-ignore lint/correctness/useHookAtTopLevel: this generic function is passed directly to React.forwardRef below.
  return useWorkbookDataGrid(props, ref);
}

export const WorkbookDataGrid = forwardRef(WorkbookDataGridInner) as <Row>(
  props: WorkbookDataGridProps<Row> & RefAttributes<GridHandle>,
) => ReactElement;

function focusOrScrollAnchor<Row>(options: {
  readonly anchor: GridCellAnchor;
  readonly columns: readonly Column<GridRecordRow<Row>, GridDraftRow<Row>>[];
  readonly focus: boolean;
  readonly grouping: WorkbookDataGridProps<Row>["grouping"];
  readonly recordRows: readonly GridRecordRow<Row>[];
  readonly vendorHandle: DataGridHandle | null;
  readonly viewSchemaId: string;
}): boolean {
  const {
    anchor,
    columns,
    focus,
    grouping,
    recordRows,
    vendorHandle,
    viewSchemaId,
  } = options;
  if (anchor.viewSchemaId !== viewSchemaId || vendorHandle === null)
    return false;
  const rowIdx = recordRows.findIndex(
    (row) => row.recordId === anchor.recordId,
  );
  const idx = columns.findIndex((column) => column.key === anchor.fieldKey);
  if (rowIdx < 0 || idx < 0) return false;
  if (grouping === null || grouping === undefined) {
    if (focus) vendorHandle.selectCell({ idx, rowIdx });
    else vendorHandle.scrollToCell({ idx, rowIdx });
    return true;
  }
  const row = Array.from(
    vendorHandle.element?.querySelectorAll<HTMLElement>(
      '[role="row"][data-grid-record-id]',
    ) ?? [],
  ).find((candidate) => candidate.dataset.gridRecordId === anchor.recordId);
  const marker = Array.from(
    row?.querySelectorAll<HTMLElement>("[data-grid-field-key]") ?? [],
  ).find((candidate) => candidate.dataset.gridFieldKey === anchor.fieldKey);
  const cell = marker?.closest<HTMLElement>('[role="gridcell"]');
  if (cell === undefined || cell === null) return false;
  cell.scrollIntoView({ block: "nearest", inline: "nearest" });
  if (focus) cell.focus({ preventScroll: true });
  return true;
}

function semanticAnchor<Row>(
  row: GridRecordRow<Row>,
  fieldKey: string,
  columns: readonly WorkbookDataGridProps<Row>["columns"][number][],
  viewSchemaId: string,
): GridCellAnchor | null {
  if (!isRecordRow(row)) return null;
  if (!columns.some((column) => column.fieldKey === fieldKey)) return null;
  return { fieldKey, recordId: row.recordId, viewSchemaId };
}

function semanticTarget<Row>(
  row: GridRecordRow<Row>,
  fieldKey: string,
  columns: readonly WorkbookDataGridProps<Row>["columns"][number][],
  viewSchemaId: string,
) {
  if (!isRecordRow(row)) return null;
  const column = columns.find((candidate) => candidate.fieldKey === fieldKey);
  if (
    column?.contractWritable !== true ||
    column.renderEditCell === undefined
  ) {
    return null;
  }
  return {
    baseRowVersion: row.rowVersion,
    fieldKey,
    recordId: row.recordId,
    viewSchemaId,
  };
}

function isRecordRow<Row>(candidate: unknown): candidate is GridRecordRow<Row> {
  if (typeof candidate !== "object" || candidate === null) return false;
  const row = candidate as Partial<GridRecordRow<Row>>;
  return (
    row.kind === "record" &&
    typeof row.recordId === "string" &&
    row.recordId.trim() !== "" &&
    typeof row.rowVersion === "number"
  );
}

function GroupedWorkbookDataGrid<Row>({
  grouping,
  density = "default",
  sharedProps,
  viewSchemaId,
}: WorkbookDataGridProps<Row> & {
  readonly sharedProps: DataGridProps<
    GridRecordRow<Row>,
    GridDraftRow<Row>,
    string
  >;
}) {
  if (grouping === null || grouping === undefined) {
    throw new Error("Grouped grid requires a grouping descriptor.");
  }
  const [collapsedGroupIdsByScope, setCollapsedGroupIdsByScope] = useState<
    ReadonlyMap<string, ReadonlySet<string>>
  >(() => new Map());
  const expansionScope = JSON.stringify([viewSchemaId, grouping.fieldKey]);
  const { groupIds, metadata, rowGroups } = useMemo(() => {
    const groups: Record<string, Array<GridRecordRow<Row>>> = {};
    const nextMetadata = new Map<string, GroupMetadata>();
    for (const row of sharedProps.rows) {
      const value = grouping.getValue(row.data);
      const id = encodeGroupValue(value);
      const groupRows = groups[id] ?? [];
      groupRows.push(row);
      groups[id] = groupRows;
      if (!nextMetadata.has(id)) {
        nextMetadata.set(id, {
          label: grouping.formatLabel(value),
          value,
        });
      }
    }
    return {
      groupIds: Object.keys(groups),
      metadata: nextMetadata,
      rowGroups: groups,
    };
  }, [grouping, sharedProps.rows]);
  const collapsedGroupIds =
    collapsedGroupIdsByScope.get(expansionScope) ?? emptyStringSet;
  useEffect(() => {
    setCollapsedGroupIdsByScope((current) => {
      const scoped = current.get(expansionScope);
      if (scoped === undefined) return current;
      const currentIds = new Set(groupIds);
      const reconciled = new Set(
        [...scoped].filter((groupId) => currentIds.has(groupId)),
      );
      if (reconciled.size === scoped.size) return current;
      const next = new Map(current);
      if (reconciled.size === 0) next.delete(expansionScope);
      else next.set(expansionScope, reconciled);
      return next;
    });
  }, [expansionScope, groupIds]);
  const expandedGroupIds = useMemo(
    () => new Set(groupIds.filter((id) => !collapsedGroupIds.has(id))),
    [collapsedGroupIds, groupIds],
  );
  const compiledColumns = sharedProps.columns as readonly Column<
    GridRecordRow<Row>,
    GridDraftRow<Row>
  >[];
  const gutterColumn = compiledColumns.find(
    (column) => column.key === gridRowGutterColumnKey,
  );
  const gutterRenderCell = gutterColumn?.renderCell;
  const groupColumnWidth = Math.max(
    typeof gutterColumn?.width === "number" ? gutterColumn.width : 48,
    128,
  );
  const groupColumn: Column<GridRecordRow<Row>, GridDraftRow<Row>> = {
    ...gutterColumn,
    frozen: true,
    key: "__cartulary_group__",
    minWidth: Math.max(gutterColumn?.minWidth ?? 48, 128),
    name: grouping.label ?? grouping.fieldKey,
    renderCell: gutterColumn?.renderCell ?? (() => null),
    renderGroupCell: ({ groupKey, isExpanded, toggleGroup }) => {
      const id = String(groupKey);
      const group = metadata.get(id);
      const label = group?.label ?? null;
      return (
        <button
          aria-expanded={isExpanded}
          data-cartulary-grid-group-id={id}
          data-testid={
            group === undefined
              ? undefined
              : grouping.getTestId?.(grouping.fieldKey, group.value, label)
          }
          ref={(node) => {
            const semanticRow = node?.closest<HTMLElement>('[role="row"]');
            if (semanticRow !== undefined && semanticRow !== null) {
              semanticRow.dataset.gridRowKind = "group";
            }
          }}
          type="button"
          onClick={toggleGroup}
        >
          {label ?? gridUnassignedGroupLabel}
        </button>
      );
    },
    width: groupColumnWidth,
  };
  const groupedColumns = [
    groupColumn,
    ...compiledColumns.filter(
      (column) => column.key !== gridRowGutterColumnKey,
    ),
  ];

  return (
    <TreeDataGrid
      {...sharedProps}
      columns={groupedColumns}
      expandedGroupIds={expandedGroupIds}
      groupBy={[groupColumn.key]}
      groupIdGetter={(groupKey) => groupKey}
      onExpandedGroupIdsChange={(next) => {
        setCollapsedGroupIdsByScope((current) => {
          const collapsed = new Set(
            groupIds.filter((groupId) => !next.has(groupId)),
          );
          const updated = new Map(current);
          if (collapsed.size === 0) updated.delete(expansionScope);
          else updated.set(expansionScope, collapsed);
          return updated;
        });
      }}
      renderers={{
        ...sharedProps.renderers,
        renderRow: (
          key,
          rowProps: RenderRowProps<GridRecordRow<Row>, GridDraftRow<Row>>,
        ) => (
          <Row
            {...rowProps}
            data-cartulary-grid-row-kind="record"
            data-grid-record-id={rowProps.row.recordId}
            data-testid={rowProps.row.testId}
            key={key}
            viewportColumns={rowProps.viewportColumns.map((column) =>
              column.key === groupColumn.key &&
              typeof gutterRenderCell === "function"
                ? { ...column, renderCell: gutterRenderCell }
                : column,
            )}
          />
        ),
      }}
      rowGrouper={() => rowGroups}
      rowHeight={workbookGridRowHeightPx(density)}
    />
  );
}

const emptyStringSet: ReadonlySet<string> = new Set();

function encodeGroupValue(value: GridGroupingScalar): string {
  if (value === null) return "n:null";
  if (typeof value === "boolean") return value ? "b:true" : "b:false";
  if (typeof value === "number") {
    if (!Number.isFinite(value)) {
      throw new Error("Grid grouping values must be finite numbers.");
    }
    return `d:${String(value)}`;
  }
  return `s:${value}`;
}
