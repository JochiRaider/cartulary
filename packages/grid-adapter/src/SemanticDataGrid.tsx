import {
  gridScrollportClassName,
  workbookGridRowHeightPx,
} from "@cartulary/ui-contracts";
import type {
  CSSProperties,
  ForwardedRef,
  MutableRefObject,
  ReactElement,
  RefAttributes,
} from "react";
import {
  forwardRef,
  useCallback,
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
  buildGridPresentationRows,
  formatGridClipboardTSV,
  type GridCellAnchor,
  type GridCellRange,
  type GridDataRow,
  type GridDraftRow,
  type GridGroupingScalar,
  type GridHandle,
  type GridRowIdentity,
  type GridRowStateInput,
  type GridSortEntry,
  type GridSurfaceIdentity,
  gridClipboardDimensions,
  gridRowIdentitiesEqual,
  gridRowIdentityKey,
  gridSurfaceIdentitiesEqual,
  gridSurfaceIdentityKey,
  gridUnassignedGroupLabel,
  isGridColumnEditable,
  resolveGridCellRange,
  resolveGridPasteTargets,
  type SemanticDataGridProps,
} from "./core";
import {
  compileGridColumns,
  type GridCompiledBulkSelection,
  gridActionsColumnKey,
  gridRowGutterColumnKey,
  gridSelectionColumnKey,
} from "./rdgCompiler";
import {
  type GridResolvedSemanticState,
  gridSemanticStateClassNames,
  mergeGridSemanticState,
  resolveGridSemanticState,
} from "./semanticState";
import { isGridVirtualizationEnabled } from "./virtualizationDiagnostics";

const emptySelectedRecordIds: ReadonlySet<string> = new Set();

type GroupMetadata = {
  readonly label: string | null;
  readonly value: GridGroupingScalar;
};

type GridVendorPosition = {
  readonly idx: number;
  readonly rowIdx: number;
};

type GridSemanticPositionMap = {
  readonly fieldKeys: readonly string[];
  readonly positions: ReadonlyMap<string, GridVendorPosition>;
  readonly rowIdentities: readonly GridRowIdentity[];
  readonly surface: GridSurfaceIdentity;
};

type PendingEditorSeed = {
  readonly anchor: GridCellAnchor;
  readonly baseRowVersion: number;
  readonly value: unknown;
};

type SemanticCellRegistration = {
  readonly cell: HTMLElement;
  readonly token: object;
};

function gridAnchorKey(anchor: GridCellAnchor): string {
  return `${gridSurfaceIdentityKey(anchor.surface)}\u0000${gridRowIdentityKey(anchor.rowIdentity)}\u0000${anchor.fieldKey}`;
}

function coreRecordId<Row>(row: GridDataRow<Row>): string | null {
  return row.rowIdentity.kind === "core_record"
    ? row.rowIdentity.recordId
    : null;
}

function coreRowVersion<Row>(row: GridDataRow<Row>): number | null {
  return row.mutationIdentity?.kind === "core_row_version"
    ? row.mutationIdentity.baseRowVersion
    : null;
}

function isCoreRecordRow<Row>(
  row: GridDataRow<Row>,
): row is GridDataRow<Row> & {
  readonly rowIdentity: Extract<
    GridRowIdentity,
    { readonly kind: "core_record" }
  >;
} {
  return row.rowIdentity.kind === "core_record";
}

function emptySemanticPositionMap(
  surface: GridSurfaceIdentity,
): GridSemanticPositionMap {
  return {
    fieldKeys: [],
    positions: new Map(),
    rowIdentities: [],
    surface,
  };
}

function sameGridCellAnchor(
  left: GridCellAnchor | null,
  right: GridCellAnchor | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      left.fieldKey === right.fieldKey &&
      gridRowIdentitiesEqual(left.rowIdentity, right.rowIdentity) &&
      gridSurfaceIdentitiesEqual(left.surface, right.surface))
  );
}

function sameGridCellRange(
  left: GridCellRange | null,
  right: GridCellRange | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      sameGridCellAnchor(left.start, right.start) &&
      sameGridCellAnchor(left.end, right.end))
  );
}

function useStableStringArray(values: readonly string[]): readonly string[] {
  const stableRef = useRef(values);
  if (
    stableRef.current.length !== values.length ||
    stableRef.current.some((value, index) => value !== values[index])
  ) {
    stableRef.current = values;
  }
  return stableRef.current;
}

function useSemanticDataGrid<Row>(
  props: SemanticDataGridProps<Row>,
  ref: ForwardedRef<GridHandle>,
) {
  const {
    activeRowIdentity = null,
    allowPasteCreateRows = false,
    actionsColumn,
    coreRecordBulkSelection,
    cellRange: controlledCellRange,
    columns,
    columnWidths,
    dataState = { kind: "ready" },
    density = "default",
    draftRow,
    fillViewportInline = false,
    getCellState,
    getRowState,
    grouping = null,
    interactionMode,
    onActiveCellChange,
    onCellRangeChange,
    onColumnReorder,
    onColumnWidthChange,
    onCopyCell,
    onEditCell,
    onFillCells,
    onPasteCell,
    onSelectRow,
    onSortChange,
    dataRows,
    rowGutter,
    sort = [],
    surface,
  } = props;
  if (
    surface.kind === "extension_grid" &&
    (actionsColumn !== undefined ||
      allowPasteCreateRows ||
      coreRecordBulkSelection !== undefined ||
      draftRow !== undefined ||
      interactionMode?.kind === "editable" ||
      onEditCell !== undefined ||
      onFillCells !== undefined ||
      onPasteCell !== undefined)
  ) {
    throw new Error(
      "Extension grid surfaces cannot enable Core mutation or bulk-selection capabilities.",
    );
  }
  const effectiveInteractionMode =
    interactionMode ??
    (surface.kind === "extension_grid"
      ? { kind: "read_only" as const, label: "Read only" }
      : { kind: "editable" as const });
  const editable = effectiveInteractionMode.kind === "editable";
  const vendorHandle = useRef<DataGridHandle>(null);
  const semanticPositionMapRef = useRef<GridSemanticPositionMap>(
    emptySemanticPositionMap(surface),
  );
  const semanticCellElementsRef = useRef(
    new Map<string, SemanticCellRegistration>(),
  );
  const pendingEditorSeedRef = useRef<PendingEditorSeed | null>(null);
  const pendingRangeEndRef = useRef<GridCellAnchor | null>(null);
  const [keyboardAnnouncement, setKeyboardAnnouncement] = useState("");
  const gridBusy =
    dataState.kind === "initial_loading" || dataState.kind === "refreshing";
  useEffect(() => {
    const element = vendorHandle.current?.element;
    if (element === null || element === undefined) return;
    element.setAttribute("aria-busy", String(gridBusy));
    element.setAttribute("aria-readonly", String(!editable));
  }, [editable, gridBusy]);
  const selectionAnchorRecordId = useRef<string | null>(null);
  const [internalCellRange, setInternalCellRange] =
    useState<GridCellRange | null>(null);
  const [activeCellAnchor, setActiveCellAnchor] =
    useState<GridCellAnchor | null>(null);
  const activeCellAnchorRef = useRef<GridCellAnchor | null>(null);
  const pendingFillIntent = useRef<
    Parameters<NonNullable<typeof onFillCells>>[0] | null
  >(null);
  const fillDispatchScheduled = useRef(false);
  const onFillCellsRef = useRef(onFillCells);
  onFillCellsRef.current = onFillCells;
  const cellRange =
    controlledCellRange === undefined ? internalCellRange : controlledCellRange;
  const cellRangeRef = useRef(cellRange);
  cellRangeRef.current = cellRange;
  const updateCellRange = useCallback(
    (range: GridCellRange | null) => {
      if (sameGridCellRange(cellRangeRef.current, range)) return;
      cellRangeRef.current = range;
      if (controlledCellRange === undefined) setInternalCellRange(range);
      onCellRangeChange?.(range);
    },
    [controlledCellRange, onCellRangeChange],
  );
  const consumeEditorSeed = useCallback(
    (
      target: PendingEditorSeed["anchor"] & {
        readonly mutationIdentity: {
          readonly kind: "core_row_version";
          readonly baseRowVersion: number;
        };
      },
    ) => {
      const pending = pendingEditorSeedRef.current;
      if (
        pending === null ||
        pending.baseRowVersion !== target.mutationIdentity.baseRowVersion ||
        !sameGridCellAnchor(pending.anchor, target)
      ) {
        return null;
      }
      queueMicrotask(() => {
        if (pendingEditorSeedRef.current === pending) {
          pendingEditorSeedRef.current = null;
        }
      });
      return { value: pending.value };
    },
    [],
  );
  const handleEditorKeyboardAction = useCallback(
    (
      target: PendingEditorSeed["anchor"],
      action:
        | { readonly kind: "exit"; readonly backwards: boolean }
        | { readonly kind: "move"; readonly rowDelta: -1 | 1 },
    ) => {
      if (action.kind === "exit") {
        focusAdjacentOutsideGrid(
          vendorHandle.current?.element ?? null,
          action.backwards,
        );
        return;
      }
      const next = moveSemanticAnchor(
        semanticPositionMapRef.current,
        target,
        0,
        action.rowDelta,
      );
      focusOrScrollAnchor({
        anchor: next ?? target,
        cellElements: semanticCellElementsRef.current,
        focus: true,
        positionMap: semanticPositionMapRef.current,
        vendorHandle: vendorHandle.current,
      });
    },
    [],
  );
  assertGridRows(dataRows);

  const presentationRows = useMemo(
    () => buildGridPresentationRows({ grouping, rows: dataRows }),
    [grouping, dataRows],
  );

  const handleSemanticPaste = useCallback(
    (
      row: GridDataRow<Row>,
      fieldKey: string,
      clipboardText: string,
    ): boolean => {
      if (!editable) return true;
      const target = semanticTarget(row, fieldKey, columns, surface);
      if (target === null) return false;
      const dimensions = gridClipboardDimensions(clipboardText);
      const targetResolution = resolveGridPasteTargets({
        allowCreateRows: allowPasteCreateRows && grouping === null,
        columns,
        current: target,
        pastedColumnCount: dimensions.columnCount,
        pastedRowCount: dimensions.rowCount,
        presentationRows,
      });
      const targetsAreEditable = targetResolution?.columns.every(
        (targetFieldKey) =>
          columns.some(
            (candidate) =>
              candidate.fieldKey === targetFieldKey &&
              candidate.contractWritable === true &&
              candidate.editor !== undefined,
          ),
      );
      if (targetResolution === null || targetsAreEditable !== true) return true;
      const lastFieldKey = targetResolution.columns.at(-1);
      const lastRowTarget = targetResolution.rowTargets.at(-1);
      const range: GridCellRange | undefined =
        lastFieldKey !== undefined && lastRowTarget?.kind === "record"
          ? {
              start: target,
              end: {
                fieldKey: lastFieldKey,
                rowIdentity: lastRowTarget.rowIdentity,
                surface,
              },
            }
          : undefined;
      if (range !== undefined) updateCellRange(range);
      onPasteCell?.({ clipboardText, range, target, targetResolution });
      return true;
    },
    [
      allowPasteCreateRows,
      columns,
      editable,
      grouping,
      onPasteCell,
      presentationRows,
      updateCellRange,
      surface,
    ],
  );

  const selectableRows = useMemo(
    () =>
      coreRecordBulkSelection === undefined || !editable
        ? []
        : dataRows.filter(
            (
              row,
            ): row is GridDataRow<Row> & {
              readonly rowIdentity: Extract<
                GridRowIdentity,
                { readonly kind: "core_record" }
              >;
            } =>
              isCoreRecordRow(row) &&
              coreRecordBulkSelection.isRecordSelectable?.(row) !== false,
          ),
    [coreRecordBulkSelection, dataRows, editable],
  );
  const compiledBulkSelection = useMemo<
    GridCompiledBulkSelection<Row> | undefined
  >(() => {
    if (coreRecordBulkSelection === undefined || !editable) return undefined;
    const selectableIds = selectableRows.map((row) => row.rowIdentity.recordId);
    const selectedOnPage = selectableIds.filter((recordId) =>
      coreRecordBulkSelection.selectedRecordIds.has(recordId),
    );
    return {
      allSelected:
        selectableIds.length > 0 &&
        selectedOnPage.length === selectableIds.length,
      partiallySelected:
        selectedOnPage.length > 0 &&
        selectedOnPage.length < selectableIds.length,
      selectedRecordIds: coreRecordBulkSelection.selectedRecordIds,
      selectableRecordCount: selectableIds.length,
      isRecordSelectable: (row) =>
        isCoreRecordRow(row) &&
        coreRecordBulkSelection.isRecordSelectable?.(row) !== false,
      onSelectAll: () => {
        selectionAnchorRecordId.current = null;
        coreRecordBulkSelection.onSelectedRecordIdsChange(
          selectedOnPage.length === selectableIds.length
            ? new Set()
            : new Set(selectableIds),
        );
      },
      onToggleRecord: (row, shiftKey) => {
        if (!isCoreRecordRow(row)) return;
        const next = new Set(coreRecordBulkSelection.selectedRecordIds);
        const anchorIndex = selectableRows.findIndex(
          (candidate) =>
            candidate.rowIdentity.recordId === selectionAnchorRecordId.current,
        );
        const rowIndex = selectableRows.findIndex(
          (candidate) =>
            candidate.rowIdentity.recordId === row.rowIdentity.recordId,
        );
        if (shiftKey && anchorIndex >= 0 && rowIndex >= 0) {
          const start = Math.min(anchorIndex, rowIndex);
          const end = Math.max(anchorIndex, rowIndex);
          for (const candidate of selectableRows.slice(start, end + 1)) {
            next.add(candidate.rowIdentity.recordId);
          }
        } else if (next.has(row.rowIdentity.recordId)) {
          next.delete(row.rowIdentity.recordId);
        } else {
          next.add(row.rowIdentity.recordId);
        }
        selectionAnchorRecordId.current = row.rowIdentity.recordId;
        coreRecordBulkSelection.onSelectedRecordIdsChange(next);
      },
    };
  }, [coreRecordBulkSelection, editable, selectableRows]);

  const selectedRows = editable
    ? (coreRecordBulkSelection?.selectedRecordIds ?? emptySelectedRecordIds)
    : emptySelectedRecordIds;
  const stale = dataState.kind === "stale_error";
  const rowStateFor = useCallback(
    (row: GridDataRow<Row>) =>
      mergeGridSemanticState(getRowState?.(row), {
        active:
          activeCellAnchor !== null &&
          gridRowIdentitiesEqual(activeCellAnchor.rowIdentity, row.rowIdentity),
        bulkSelected:
          coreRecordId(row) !== null &&
          selectedRows.has(coreRecordId(row) ?? ""),
        inspectorActive:
          activeRowIdentity !== null &&
          gridRowIdentitiesEqual(activeRowIdentity, row.rowIdentity),
        readOnlyOrDerived: !editable,
        saved: true,
        stale,
      }),
    [
      activeCellAnchor,
      activeRowIdentity,
      editable,
      getRowState,
      selectedRows,
      stale,
    ],
  );
  const cellStateFor = useCallback(
    (row: GridDataRow<Row>, column: (typeof columns)[number]) => {
      const anchor = {
        fieldKey: column.fieldKey,
        rowIdentity: row.rowIdentity,
        surface,
      };
      const rowState = rowStateFor(row);
      return mergeGridSemanticState(
        getCellState?.({
          anchor,
          mutationIdentity: row.mutationIdentity,
          row: row.data,
        }),
        {
          active:
            activeCellAnchor !== null &&
            gridRowIdentitiesEqual(
              activeCellAnchor.rowIdentity,
              row.rowIdentity,
            ) &&
            activeCellAnchor.fieldKey === column.fieldKey,
          bulkSelected: rowState.bulkSelected,
          inspectorActive: rowState.inspectorActive,
          // Row-level pending/stale states are rendered once on the row. They
          // must not manufacture a marker in every cell; an owner can still
          // opt a specific cell into either state through getCellState.
          pending: false,
          readOnlyOrDerived: !editable || !isGridColumnEditable(column),
          saved: true,
          stale: false,
        },
      );
    },
    [activeCellAnchor, editable, getCellState, rowStateFor, surface],
  );
  const isCellRangeSelected = useCallback(
    (row: GridDataRow<Row>, column: (typeof columns)[number]) =>
      gridCellRangeContains(
        semanticPositionMapRef.current,
        cellRangeRef.current,
        {
          fieldKey: column.fieldKey,
          rowIdentity: row.rowIdentity,
          surface,
        },
      ),
    [surface],
  );
  const registerSemanticCell = useCallback(
    (anchor: GridCellAnchor, cell: HTMLElement | null, token: object) => {
      const key = gridAnchorKey(anchor);
      if (cell === null) {
        if (semanticCellElementsRef.current.get(key)?.token === token) {
          semanticCellElementsRef.current.delete(key);
        }
      } else {
        semanticCellElementsRef.current.set(key, { cell, token });
      }
    },
    [],
  );

  const compiledColumns = useMemo(
    () =>
      compileGridColumns({
        actionsColumn,
        bulkSelection: compiledBulkSelection,
        cellStateFor,
        columns,
        consumeEditorSeed,
        editable,
        isCellRangeSelected,
        onEditorKeyboardAction: handleEditorKeyboardAction,
        onPasteCellContent:
          onPasteCell === undefined ? undefined : handleSemanticPaste,
        registerSemanticCell,
        rowGutter,
        surface,
      }),
    [
      actionsColumn,
      cellStateFor,
      columns,
      compiledBulkSelection,
      consumeEditorSeed,
      editable,
      handleEditorKeyboardAction,
      handleSemanticPaste,
      isCellRangeSelected,
      onPasteCell,
      registerSemanticCell,
      rowGutter,
      surface,
    ],
  );
  const compiledColumnKeys = useStableStringArray(
    compiledColumns.map((column) => column.key),
  );
  const semanticFieldKeys = useStableStringArray(
    columns.map((column) => column.fieldKey),
  );
  const ungroupedPositionMap = useMemo(
    () =>
      grouping === null
        ? buildSemanticPositionMap({
            columnKeys: compiledColumnKeys,
            fieldKeys: semanticFieldKeys,
            dataRows,
            surface,
          })
        : null,
    [compiledColumnKeys, dataRows, grouping, semanticFieldKeys, surface],
  );
  if (ungroupedPositionMap !== null) {
    semanticPositionMapRef.current = ungroupedPositionMap;
  }
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
          cellElements: semanticCellElementsRef.current,
          positionMap: semanticPositionMapRef.current,
          vendorHandle: vendorHandle.current,
          focus: true,
        }),
      getScrollElement: () => vendorHandle.current?.element ?? null,
      scrollToAnchor: (anchor) =>
        focusOrScrollAnchor({
          anchor,
          cellElements: semanticCellElementsRef.current,
          positionMap: semanticPositionMapRef.current,
          vendorHandle: vendorHandle.current,
          focus: false,
        }),
    }),
    [],
  );
  const sharedProps: DataGridProps<
    GridDataRow<Row>,
    GridDraftRow<Row>,
    string
  > = {
    // RDG calls this private renderer channel a bottom summary. Cartulary uses
    // it only to realize the recordless create draft; it is not a public
    // summary-row or aggregation capability.
    bottomSummaryRows:
      !editable || draftRow === undefined ? undefined : [draftRow],
    className: `${gridScrollportClassName()} cartulary-grid rdg-dark`,
    columnWidths: rdgColumnWidths,
    columns: compiledColumns,
    // Production grids always use RDG's row and column virtualization. A
    // result-size threshold would create two interaction runtimes and let
    // small fixtures miss focus, range, editor, and scrolling defects.
    enableVirtualization: isGridVirtualizationEnabled(),
    headerRowHeight: 32,
    headerRowClass: "cartulary-grid-header-row",
    onCellClick: ({ column, row }) => {
      onSelectRow?.(row.rowIdentity);
      const anchor = semanticAnchor(row, column.key, columns, surface);
      if (anchor === null) return;
      pendingRangeEndRef.current = null;
      if (!sameGridCellAnchor(activeCellAnchorRef.current, anchor)) {
        activeCellAnchorRef.current = anchor;
        setActiveCellAnchor(anchor);
        onActiveCellChange?.(anchor);
      }
      updateCellRange({ end: anchor, start: anchor });
    },
    onCellCopy: ({ column, row }, event) => {
      const anchor = semanticAnchor(row, column.key, columns, surface);
      if (anchor === null) return;
      const range = cellRange ?? { end: anchor, start: anchor };
      const expandedRange = resolveVisibleGridCellRange({
        columns,
        positionMap: semanticPositionMapRef.current,
        dataRows,
        range,
      });
      if (expandedRange === null) return;
      const values = expandedRange.rowIdentities.map((rowIdentity) => {
        const record = dataRows.find((candidate) =>
          gridRowIdentitiesEqual(candidate.rowIdentity, rowIdentity),
        );
        return expandedRange.fieldKeys.map((fieldKey) => {
          const semanticColumn = columns.find(
            (candidate) => candidate.fieldKey === fieldKey,
          );
          return record === undefined
            ? ""
            : (semanticColumn?.getClipboardValue?.(record.data) ?? "");
        });
      });
      event.clipboardData?.setData(
        "text/plain",
        formatGridClipboardTSV(values),
      );
      event.preventDefault();
      onCopyCell?.({ anchor, expandedRange, range });
    },
    onCellDoubleClick: ({ column, row }) => {
      if (!editable) return;
      const target = semanticTarget(row, column.key, columns, surface);
      if (target !== null) onEditCell?.({ target });
    },
    onCellKeyDown: (args, event) => {
      if (args.mode !== "SELECT" || !isDataRow<Row>(args.row)) return;
      const anchor = semanticAnchor(
        args.row,
        args.column.key,
        columns,
        surface,
      );
      if (anchor === null) return;

      if (event.key === "Tab") {
        event.preventDefault();
        event.preventGridDefault();
        focusAdjacentOutsideGrid(
          vendorHandle.current?.element ?? null,
          event.shiftKey,
        );
        return;
      }

      const semanticColumn = columns.find(
        (column) => column.fieldKey === anchor.fieldKey,
      );
      const canEdit =
        editable &&
        semanticColumn !== undefined &&
        isGridColumnEditable(semanticColumn);
      const startEditor = (seed: unknown, hasSeed: boolean) => {
        event.preventDefault();
        event.preventGridDefault();
        if (!canEdit || semanticColumn === undefined) {
          setKeyboardAnnouncement(
            editable
              ? `${semanticColumn?.label ?? "Cell"} is read-only.`
              : effectiveInteractionMode.kind === "read_only"
                ? effectiveInteractionMode.label
                : "This workbook is read-only.",
          );
          return;
        }
        const baseRowVersion = coreRowVersion(args.row);
        if (baseRowVersion === null) {
          setKeyboardAnnouncement("This row cannot be edited.");
          return;
        }
        pendingEditorSeedRef.current = hasSeed
          ? {
              anchor,
              baseRowVersion,
              value: seed,
            }
          : null;
        const target = semanticTarget(
          args.row,
          anchor.fieldKey,
          columns,
          surface,
        );
        if (target !== null) onEditCell?.({ target });
        args.selectCell(
          { idx: args.column.idx, rowIdx: args.rowIdx },
          { enableEditor: true, shouldFocusCell: true },
        );
      };

      if (event.key === "Enter") {
        startEditor(undefined, false);
        return;
      }
      if (isPrintableGridEntry(event)) {
        startEditor(event.key, true);
        return;
      }
      if (event.key === "Backspace" || event.key === "Delete") {
        if (semanticColumn?.editor?.clearDraftValue === undefined) {
          event.preventDefault();
          event.preventGridDefault();
          setKeyboardAnnouncement(
            `${semanticColumn?.label ?? "This field"} cannot be cleared.`,
          );
          return;
        }
        startEditor(semanticColumn.editor.clearDraftValue, true);
        return;
      }
      if (!isSemanticNavigationKey(event.key)) return;

      event.preventDefault();
      event.preventGridDefault();
      const next = navigateSemanticAnchor(
        semanticPositionMapRef.current,
        anchor,
        {
          ctrlOrMetaKey: event.ctrlKey || event.metaKey,
          key: event.key,
          pageSize: visibleGridPageSize(
            vendorHandle.current?.element ?? null,
            workbookGridRowHeightPx(density),
          ),
        },
      );
      if (next === null) return;
      const nextPosition = semanticPositionMapRef.current.positions.get(
        gridAnchorKey(next),
      );
      if (nextPosition === undefined) return;
      if (event.shiftKey && event.key.startsWith("Arrow")) {
        const range = {
          start: cellRangeRef.current?.start ?? anchor,
          end: next,
        };
        pendingRangeEndRef.current = next;
        updateCellRange(range);
      }
      args.selectCell(nextPosition, { shouldFocusCell: true });
    },
    onCellPaste: ({ column, row }, event) => {
      const clipboardText = event.clipboardData?.getData("text/plain") ?? "";
      if (!editable || handleSemanticPaste(row, column.key, clipboardText)) {
        event.preventDefault();
      }
      return row;
    },
    onColumnResize: (
      column: CalculatedColumn<GridDataRow<Row>, GridDraftRow<Row>>,
      width: number,
    ) => {
      if (columns.some((candidate) => candidate.fieldKey === column.key)) {
        onColumnWidthChange?.(column.key, width);
      }
    },
    onColumnsReorder: (sourceKey: string, targetKey: string) => {
      if (
        columns.some((column) => column.fieldKey === sourceKey) &&
        columns.some((column) => column.fieldKey === targetKey)
      ) {
        onColumnReorder?.(sourceKey, targetKey);
      }
    },
    onFill:
      editable && grouping === null && onFillCells !== undefined
        ? ({ columnKey, sourceRow, targetRow }) => {
            const source = semanticTarget(
              sourceRow,
              columnKey,
              columns,
              surface,
            );
            const target = semanticTarget(
              targetRow,
              columnKey,
              columns,
              surface,
            );
            if (source !== null && target !== null) {
              const column = columns.find(
                (candidate) => candidate.fieldKey === columnKey,
              );
              const range = { start: source, end: target };
              const expanded = resolveGridCellRange({
                columns,
                presentationRows,
                range,
              });
              if (
                column?.contractWritable === true &&
                column.editor !== undefined &&
                column.valueKind !== "collection" &&
                expanded !== null
              ) {
                const targets = expanded.rowIdentities.flatMap(
                  (rowIdentity) => {
                    const dataRow = dataRows.find((candidate) =>
                      gridRowIdentitiesEqual(
                        candidate.rowIdentity,
                        rowIdentity,
                      ),
                    );
                    return dataRow?.mutationIdentity === undefined
                      ? []
                      : [
                          {
                            fieldKey: columnKey,
                            mutationIdentity: dataRow.mutationIdentity,
                            rowIdentity,
                            surface,
                          },
                        ];
                  },
                );
                if (targets.length !== expanded.rowIdentities.length) {
                  return targetRow;
                }
                const pending = pendingFillIntent.current;
                pendingFillIntent.current =
                  pending !== null &&
                  gridRowIdentitiesEqual(
                    pending.source.rowIdentity,
                    source.rowIdentity,
                  ) &&
                  pending.source.fieldKey === source.fieldKey
                    ? {
                        range,
                        source,
                        target,
                        targets: dedupeGridTargets([
                          ...pending.targets,
                          ...targets,
                        ]),
                      }
                    : { range, source, target, targets };
                updateCellRange(range);
                if (!fillDispatchScheduled.current) {
                  fillDispatchScheduled.current = true;
                  queueMicrotask(() => {
                    fillDispatchScheduled.current = false;
                    const intent = pendingFillIntent.current;
                    pendingFillIntent.current = null;
                    if (intent !== null) onFillCellsRef.current?.(intent);
                  });
                }
              }
            }
            return targetRow;
          }
        : undefined,
    onRowsChange:
      !editable ||
      (onPasteCell === undefined &&
        !columns.some((column) => column.editor !== undefined))
        ? undefined
        : () => {},
    onSelectedCellChange: ({ column, row }) => {
      const anchor =
        row === undefined
          ? null
          : semanticAnchor(row, column.key, columns, surface);
      const preserveRange = sameGridCellAnchor(
        pendingRangeEndRef.current,
        anchor,
      );
      pendingRangeEndRef.current = null;
      if (!sameGridCellAnchor(activeCellAnchorRef.current, anchor)) {
        activeCellAnchorRef.current = anchor;
        setActiveCellAnchor(anchor);
        onActiveCellChange?.(anchor);
      }
      if (!preserveRange) {
        updateCellRange(
          anchor === null ? null : { end: anchor, start: anchor },
        );
      }
    },
    onSortColumnsChange: (next: SortColumn[]) => {
      const semanticSort: readonly GridSortEntry[] = next.map((entry) => ({
        direction: entry.direction === "ASC" ? "asc" : "desc",
        fieldKey: entry.columnKey,
      }));
      onSortChange?.(semanticSort);
    },
    renderers: {
      // RDG 7 suppresses its private bottom-summary channel whenever
      // noRowsFallback is present and the committed collection is empty. The
      // recordless create draft must remain usable at zero committed rows, so
      // reserve the fallback for grids without a draft.
      noRowsFallback: undefined,
      renderRow: (
        key,
        rowProps: RenderRowProps<GridDataRow<Row>, GridDraftRow<Row>>,
      ) => {
        const semanticState = resolveGridSemanticState(
          rowStateFor(rowProps.row),
          "data row",
        );
        return (
          <Row
            {...rowProps}
            {...semanticRowAttributes(semanticState)}
            data-cartulary-grid-row-kind="data"
            data-inspector-active={
              activeRowIdentity !== null &&
              gridRowIdentitiesEqual(
                rowProps.row.rowIdentity,
                activeRowIdentity,
              )
                ? "true"
                : undefined
            }
            data-grid-row-identity-kind={rowProps.row.rowIdentity.kind}
            data-grid-record-id={
              rowProps.row.rowIdentity.kind === "core_record"
                ? rowProps.row.rowIdentity.recordId
                : undefined
            }
            data-testid={rowProps.row.testId}
            key={key}
          />
        );
      },
    },
    rowClass: (row) =>
      gridSemanticStateClassNames(
        "row",
        resolveGridSemanticState(rowStateFor(row), "data row"),
      ),
    rowKeyGetter: (row: GridDataRow<Row>) =>
      gridRowIdentityKey(row.rowIdentity),
    ref: vendorHandle,
    rows: dataRows,
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

  const grid =
    grouping === null ? (
      <DataGrid {...sharedProps} rowHeight={workbookGridRowHeightPx(density)} />
    ) : (
      <GroupedSemanticDataGrid
        {...props}
        density={density}
        positionMapRef={semanticPositionMapRef}
        rowStateFor={rowStateFor}
        sharedProps={sharedProps}
      />
    );

  return (
    <div className="cartulary-grid-state-frame" style={gridStateFrameStyle}>
      {grid}
      <GridStatePresentation
        dataState={dataState}
        interactionMode={effectiveInteractionMode}
      />
      {coreRecordBulkSelection === undefined ? null : (
        <span
          aria-live="polite"
          className="cartulary-grid-live-region"
          role="status"
        >
          {selectedRows.size === 0
            ? "No records selected."
            : `${selectedRows.size} ${selectedRows.size === 1 ? "record" : "records"} selected.`}
        </span>
      )}
      <GridRangeAnnouncement
        columns={columns}
        dataRows={dataRows}
        positionMap={semanticPositionMapRef.current}
        range={cellRange}
      />
      {keyboardAnnouncement === "" ? null : (
        <span
          aria-live="assertive"
          className="cartulary-grid-live-region"
          role="alert"
        >
          {keyboardAnnouncement}
        </span>
      )}
    </div>
  );
}

function GridRangeAnnouncement<Row>({
  columns,
  dataRows,
  positionMap,
  range,
}: {
  readonly columns: readonly SemanticDataGridProps<Row>["columns"][number][];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly positionMap: GridSemanticPositionMap;
  readonly range: GridCellRange | null;
}) {
  if (range === null) return null;
  const expanded = resolveVisibleGridCellRange({
    columns,
    dataRows,
    positionMap,
    range,
  });
  if (
    expanded === null ||
    (expanded.fieldKeys.length === 1 && expanded.rowIdentities.length === 1)
  ) {
    return null;
  }
  return (
    <span
      aria-live="polite"
      className="cartulary-grid-live-region"
      role="status"
    >
      Selected {expanded.rowIdentities.length} rows by{" "}
      {expanded.fieldKeys.length} columns.
    </span>
  );
}

function GridStatePresentation({
  dataState,
  interactionMode,
}: {
  readonly dataState: NonNullable<SemanticDataGridProps<unknown>["dataState"]>;
  readonly interactionMode: NonNullable<
    SemanticDataGridProps<unknown>["interactionMode"]
  >;
}) {
  const presentation = gridDataStatePresentation(dataState);
  return (
    <>
      {presentation === null ? null : (
        <div
          aria-live={presentation.live}
          className={`cartulary-grid-state cartulary-grid-state-${dataState.kind}`}
          data-grid-data-state={dataState.kind}
          role={presentation.role}
          style={{
            ...gridStateStyle,
            ...(presentation.blocking ? gridBlockingStateStyle : null),
          }}
        >
          <span>{presentation.message}</span>
          {presentation.action === undefined ? null : (
            <button
              style={gridStateActionStyle}
              type="button"
              onClick={presentation.action.onInvoke}
            >
              {presentation.action.label}
            </button>
          )}
        </div>
      )}
      {interactionMode.kind === "read_only" ? (
        <div
          aria-live="polite"
          className="cartulary-grid-interaction-state"
          data-grid-interaction-mode="read_only"
          role="status"
          style={gridInteractionStateStyle}
        >
          {interactionMode.label}
        </div>
      ) : null}
    </>
  );
}

function gridDataStatePresentation(
  state: NonNullable<SemanticDataGridProps<unknown>["dataState"]>,
): {
  readonly action?: { readonly label: string; readonly onInvoke: () => void };
  readonly blocking: boolean;
  readonly live: "assertive" | "polite";
  readonly message: string;
  readonly role: "alert" | "status";
} | null {
  switch (state.kind) {
    case "ready":
      return null;
    case "initial_loading":
      return {
        blocking: true,
        live: "polite",
        message: `Loading ${state.surfaceLabel}…`,
        role: "status",
      };
    case "refreshing":
      return {
        blocking: false,
        live: "polite",
        message: `Refreshing ${state.surfaceLabel}…`,
        role: "status",
      };
    case "empty":
      return {
        ...(state.action === undefined ? {} : { action: state.action }),
        blocking: false,
        live: "polite",
        message: state.message,
        role: "status",
      };
    case "filtered_empty":
      return {
        action: state.action,
        blocking: false,
        live: "polite",
        message: "No rows match the current filters.",
        role: "status",
      };
    case "stale_error":
      return {
        ...(state.action === undefined ? {} : { action: state.action }),
        blocking: false,
        live: "assertive",
        message: `${state.message} Previously loaded rows may be stale.`,
        role: "alert",
      };
    case "unavailable":
      return {
        ...(state.action === undefined ? {} : { action: state.action }),
        blocking: true,
        live: "assertive",
        message: state.message,
        role: "alert",
      };
    case "permission_denied":
      return {
        blocking: true,
        live: "assertive",
        message: state.message ?? "You no longer have access to this workbook.",
        role: "alert",
      };
  }
}

const gridStateFrameStyle = {
  blockSize: "100%",
  minBlockSize: 0,
  position: "relative",
} satisfies CSSProperties;

const gridStateStyle = {
  alignItems: "center",
  background: "var(--ct-colors-surface-2)",
  border: "var(--ct-border-hairline)",
  display: "flex",
  gap: "0.75rem",
  insetBlockStart: "0.5rem",
  insetInline: "0.5rem",
  justifyContent: "space-between",
  padding: "0.65rem 0.8rem",
  pointerEvents: "none",
  position: "absolute",
  zIndex: 5,
} satisfies CSSProperties;

const gridStateActionStyle = {
  pointerEvents: "auto",
} satisfies CSSProperties;

const gridBlockingStateStyle = {
  insetBlock: 0,
  justifyContent: "center",
} satisfies CSSProperties;

const gridInteractionStateStyle = {
  background: "var(--ct-colors-surface-3)",
  border: "var(--ct-border-hairline)",
  insetBlockEnd: "0.5rem",
  insetInlineEnd: "0.5rem",
  padding: "0.35rem 0.55rem",
  position: "absolute",
  zIndex: 4,
} satisfies CSSProperties;

function SemanticDataGridInner<Row>(
  props: SemanticDataGridProps<Row>,
  ref: ForwardedRef<GridHandle>,
) {
  // biome-ignore lint/correctness/useHookAtTopLevel: this generic function is passed directly to React.forwardRef below.
  return useSemanticDataGrid(props, ref);
}

export const SemanticDataGrid = forwardRef(SemanticDataGridInner) as <Row>(
  props: SemanticDataGridProps<Row> & RefAttributes<GridHandle>,
) => ReactElement;

function focusOrScrollAnchor(options: {
  readonly anchor: GridCellAnchor;
  readonly cellElements: ReadonlyMap<string, SemanticCellRegistration>;
  readonly focus: boolean;
  readonly positionMap: GridSemanticPositionMap;
  readonly vendorHandle: DataGridHandle | null;
}): boolean {
  const { anchor, cellElements, focus, positionMap, vendorHandle } = options;
  if (
    !gridSurfaceIdentitiesEqual(anchor.surface, positionMap.surface) ||
    vendorHandle === null
  )
    return false;
  const position = positionMap.positions.get(gridAnchorKey(anchor));
  if (position === undefined) return false;
  if (focus) {
    vendorHandle.selectCell(position, { shouldFocusCell: true });
    focusSemanticCellAfterRender(cellElements, anchor);
  } else {
    vendorHandle.scrollToCell(position);
  }
  return true;
}

function focusSemanticCellAfterRender(
  cellElements: ReadonlyMap<string, SemanticCellRegistration>,
  anchor: GridCellAnchor,
) {
  // beta.59 does not honor shouldFocusCell when selectCell receives the
  // already-selected position. Keep the workaround private and resolve the
  // element from Cartulary's private semantic registry, never RDG coordinates
  // or generated class names.
  const focusRegisteredCell = () => {
    const cell = cellElements.get(gridAnchorKey(anchor))?.cell;
    if (cell === undefined) return null;
    if (!cell.hasAttribute("tabindex")) cell.tabIndex = -1;
    cell.focus({ preventScroll: true });
    return cell;
  };
  window.setTimeout(() => {
    const focusedCell = focusRegisteredCell();
    // Focusing away from an inspector/editor may synchronously publish its
    // final clean draft and replace the rendered cell. Re-resolve once after
    // that render, but never reclaim focus after the user has moved to a
    // different live control.
    window.setTimeout(() => {
      const activeElement = document.activeElement;
      if (activeElement !== focusedCell && activeElement !== document.body) {
        return;
      }
      focusRegisteredCell();
    }, 0);
  }, 0);
}

function buildSemanticPositionMap<Row>({
  columnKeys,
  dataRows,
  fieldKeys,
  rowIndexes,
  surface,
}: {
  readonly columnKeys: readonly string[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly fieldKeys: readonly string[];
  readonly rowIndexes?: ReadonlyMap<string, number> | undefined;
  readonly surface: GridSurfaceIdentity;
}): GridSemanticPositionMap {
  const positions = new Map<string, GridVendorPosition>();
  const columnIndexes = new Map(
    columnKeys.map((columnKey, index) => [columnKey, index]),
  );
  const rowIdentities: GridRowIdentity[] = [];
  for (const [fallbackRowIdx, row] of dataRows.entries()) {
    const rowKey = gridRowIdentityKey(row.rowIdentity);
    const rowIdx = rowIndexes?.get(rowKey) ?? fallbackRowIdx;
    if (rowIdx < 0) continue;
    rowIdentities.push(row.rowIdentity);
    for (const fieldKey of fieldKeys) {
      const idx = columnIndexes.get(fieldKey);
      if (idx === undefined) continue;
      const anchor = { fieldKey, rowIdentity: row.rowIdentity, surface };
      positions.set(gridAnchorKey(anchor), { idx, rowIdx });
    }
  }
  return { fieldKeys, positions, rowIdentities, surface };
}

function moveSemanticAnchor(
  positionMap: GridSemanticPositionMap,
  anchor: GridCellAnchor,
  columnDelta: number,
  rowDelta: number,
): GridCellAnchor | null {
  const columnIndex = positionMap.fieldKeys.indexOf(anchor.fieldKey);
  const rowIndex = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, anchor.rowIdentity),
  );
  const fieldKey = positionMap.fieldKeys[columnIndex + columnDelta];
  const rowIdentity = positionMap.rowIdentities[rowIndex + rowDelta];
  if (
    columnIndex < 0 ||
    rowIndex < 0 ||
    fieldKey === undefined ||
    rowIdentity === undefined
  ) {
    return null;
  }
  return { fieldKey, rowIdentity, surface: positionMap.surface };
}

function navigateSemanticAnchor(
  positionMap: GridSemanticPositionMap,
  anchor: GridCellAnchor,
  intent: {
    readonly ctrlOrMetaKey: boolean;
    readonly key:
      | "ArrowDown"
      | "ArrowLeft"
      | "ArrowRight"
      | "ArrowUp"
      | "End"
      | "Home"
      | "PageDown"
      | "PageUp";
    readonly pageSize: number;
  },
): GridCellAnchor | null {
  const columnIndex = positionMap.fieldKeys.indexOf(anchor.fieldKey);
  const rowIndex = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, anchor.rowIdentity),
  );
  if (columnIndex < 0 || rowIndex < 0) return null;
  let nextColumnIndex = columnIndex;
  let nextRowIndex = rowIndex;
  switch (intent.key) {
    case "ArrowDown":
      nextRowIndex += 1;
      break;
    case "ArrowLeft":
      nextColumnIndex -= 1;
      break;
    case "ArrowRight":
      nextColumnIndex += 1;
      break;
    case "ArrowUp":
      nextRowIndex -= 1;
      break;
    case "PageDown":
      nextRowIndex = Math.min(
        positionMap.rowIdentities.length - 1,
        rowIndex + intent.pageSize,
      );
      break;
    case "PageUp":
      nextRowIndex = Math.max(0, rowIndex - intent.pageSize);
      break;
    case "Home":
      nextColumnIndex = 0;
      if (intent.ctrlOrMetaKey) nextRowIndex = 0;
      break;
    case "End":
      nextColumnIndex = positionMap.fieldKeys.length - 1;
      if (intent.ctrlOrMetaKey)
        nextRowIndex = positionMap.rowIdentities.length - 1;
      break;
  }
  const fieldKey = positionMap.fieldKeys[nextColumnIndex];
  const rowIdentity = positionMap.rowIdentities[nextRowIndex];
  if (fieldKey === undefined || rowIdentity === undefined) return null;
  return { fieldKey, rowIdentity, surface: positionMap.surface };
}

function isSemanticNavigationKey(
  key: string,
): key is
  | "ArrowDown"
  | "ArrowLeft"
  | "ArrowRight"
  | "ArrowUp"
  | "End"
  | "Home"
  | "PageDown"
  | "PageUp" {
  return (
    key === "ArrowDown" ||
    key === "ArrowLeft" ||
    key === "ArrowRight" ||
    key === "ArrowUp" ||
    key === "End" ||
    key === "Home" ||
    key === "PageDown" ||
    key === "PageUp"
  );
}

function isPrintableGridEntry(event: {
  readonly altKey: boolean;
  readonly ctrlKey: boolean;
  readonly key: string;
  readonly metaKey: boolean;
}): boolean {
  return (
    event.key.length === 1 && !event.altKey && !event.ctrlKey && !event.metaKey
  );
}

function visibleGridPageSize(
  gridElement: HTMLDivElement | null,
  rowHeight: number,
): number {
  if (gridElement === null || rowHeight <= 0) return 1;
  return Math.max(1, Math.floor(gridElement.clientHeight / rowHeight) - 2);
}

function focusAdjacentOutsideGrid(
  gridElement: HTMLDivElement | null,
  backwards: boolean,
): boolean {
  if (gridElement === null) return false;
  const focusable = Array.from(
    document.querySelectorAll<HTMLElement>(
      "a[href], button, input, select, textarea, [tabindex]",
    ),
  ).filter(
    (element) =>
      !element.hasAttribute("disabled") &&
      element.getAttribute("aria-hidden") !== "true" &&
      element.tabIndex >= 0,
  );
  const gridIndexes = focusable.flatMap((element, index) =>
    element === gridElement || gridElement.contains(element) ? [index] : [],
  );
  if (gridIndexes.length === 0) return false;
  const targetIndex = backwards
    ? Math.min(...gridIndexes) - 1
    : Math.max(...gridIndexes) + 1;
  const target = focusable[targetIndex];
  if (target === undefined) {
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
    return !gridElement.contains(document.activeElement);
  }
  target.focus();
  return document.activeElement === target;
}

function resolveVisibleGridCellRange<Row>({
  columns,
  dataRows,
  positionMap,
  range,
}: {
  readonly columns: readonly SemanticDataGridProps<Row>["columns"][number][];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly positionMap: GridSemanticPositionMap;
  readonly range: GridCellRange;
}) {
  if (
    !gridSurfaceIdentitiesEqual(range.start.surface, positionMap.surface) ||
    !gridSurfaceIdentitiesEqual(range.end.surface, positionMap.surface)
  ) {
    return null;
  }
  const startColumn = positionMap.fieldKeys.indexOf(range.start.fieldKey);
  const endColumn = positionMap.fieldKeys.indexOf(range.end.fieldKey);
  const startRow = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, range.start.rowIdentity),
  );
  const endRow = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, range.end.rowIdentity),
  );
  if (startColumn < 0 || endColumn < 0 || startRow < 0 || endRow < 0) {
    return null;
  }
  const fieldKeys = positionMap.fieldKeys.slice(
    Math.min(startColumn, endColumn),
    Math.max(startColumn, endColumn) + 1,
  );
  if (
    !fieldKeys.every((fieldKey) =>
      columns.some((column) => column.fieldKey === fieldKey),
    )
  ) {
    return null;
  }
  const rowIdentities = positionMap.rowIdentities.slice(
    Math.min(startRow, endRow),
    Math.max(startRow, endRow) + 1,
  );
  if (
    rowIdentities.some(
      (rowIdentity) =>
        !dataRows.some((row) =>
          gridRowIdentitiesEqual(row.rowIdentity, rowIdentity),
        ),
    )
  ) {
    return null;
  }
  return { fieldKeys, rowIdentities };
}

function gridCellRangeContains(
  positionMap: GridSemanticPositionMap,
  range: GridCellRange | null,
  anchor: GridCellAnchor,
): boolean {
  if (range === null) return false;
  if (sameGridCellAnchor(range.start, range.end)) return false;
  const startColumn = positionMap.fieldKeys.indexOf(range.start.fieldKey);
  const endColumn = positionMap.fieldKeys.indexOf(range.end.fieldKey);
  const column = positionMap.fieldKeys.indexOf(anchor.fieldKey);
  const startRow = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, range.start.rowIdentity),
  );
  const endRow = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, range.end.rowIdentity),
  );
  const row = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, anchor.rowIdentity),
  );
  if (
    startColumn < 0 ||
    endColumn < 0 ||
    column < 0 ||
    startRow < 0 ||
    endRow < 0 ||
    row < 0
  ) {
    return false;
  }
  return (
    column >= Math.min(startColumn, endColumn) &&
    column <= Math.max(startColumn, endColumn) &&
    row >= Math.min(startRow, endRow) &&
    row <= Math.max(startRow, endRow)
  );
}

function semanticAnchor<Row>(
  row: GridDataRow<Row>,
  fieldKey: string,
  columns: readonly SemanticDataGridProps<Row>["columns"][number][],
  surface: GridSurfaceIdentity,
): GridCellAnchor | null {
  if (!isDataRow(row)) return null;
  if (!columns.some((column) => column.fieldKey === fieldKey)) return null;
  return { fieldKey, rowIdentity: row.rowIdentity, surface };
}

function semanticTarget<Row>(
  row: GridDataRow<Row>,
  fieldKey: string,
  columns: readonly SemanticDataGridProps<Row>["columns"][number][],
  surface: GridSurfaceIdentity,
) {
  if (
    !isDataRow(row) ||
    surface.kind !== "view_schema" ||
    row.rowIdentity.kind !== "core_record" ||
    row.mutationIdentity === undefined
  ) {
    return null;
  }
  const column = columns.find((candidate) => candidate.fieldKey === fieldKey);
  if (column?.contractWritable !== true || column.editor === undefined) {
    return null;
  }
  return {
    fieldKey,
    mutationIdentity: row.mutationIdentity,
    rowIdentity: row.rowIdentity,
    surface,
  };
}

function dedupeGridTargets(
  targets: readonly {
    readonly fieldKey: string;
    readonly mutationIdentity: {
      readonly kind: "core_row_version";
      readonly baseRowVersion: number;
    };
    readonly rowIdentity: GridRowIdentity;
    readonly surface: GridSurfaceIdentity;
  }[],
) {
  const seen = new Set<string>();
  return targets.filter((target) => {
    const key = gridAnchorKey(target);
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

function semanticRowAttributes(state: GridResolvedSemanticState) {
  return {
    "aria-busy": state.stateIds.includes("pending") || undefined,
    "aria-current": state.stateIds.includes("inspector-active") || undefined,
    "aria-description": state.description,
    "aria-invalid": state.stateIds.includes("invalid") || undefined,
    "aria-selected": state.stateIds.includes("bulk-selected") || undefined,
    "data-grid-primary-state": state.primary,
    "data-grid-semantic-states": state.stateIds.join(" "),
  } as const;
}

function isDataRow<Row>(candidate: unknown): candidate is GridDataRow<Row> {
  if (typeof candidate !== "object" || candidate === null) return false;
  const row = candidate as Partial<GridDataRow<Row>>;
  return row.kind === "data" && row.rowIdentity !== undefined;
}

function GroupedSemanticDataGrid<Row>({
  grouping,
  density = "default",
  positionMapRef,
  rowStateFor,
  sharedProps,
  surface,
}: SemanticDataGridProps<Row> & {
  readonly sharedProps: DataGridProps<
    GridDataRow<Row>,
    GridDraftRow<Row>,
    string
  >;
  readonly positionMapRef: MutableRefObject<GridSemanticPositionMap>;
  readonly rowStateFor: (row: GridDataRow<Row>) => GridRowStateInput;
}) {
  if (grouping === null || grouping === undefined) {
    throw new Error("Grouped grid requires a grouping descriptor.");
  }
  const [collapsedGroupIdsByScope, setCollapsedGroupIdsByScope] = useState<
    ReadonlyMap<string, ReadonlySet<string>>
  >(() => new Map());
  const expansionScope = JSON.stringify([
    gridSurfaceIdentityKey(surface),
    grouping.fieldKey,
  ]);
  const { groupIds, metadata, rowGroups } = useMemo(() => {
    const groups: Record<string, Array<GridDataRow<Row>>> = {};
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
    GridDataRow<Row>,
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
  const groupColumn: Column<GridDataRow<Row>, GridDraftRow<Row>> = {
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
              semanticRow.classList.add("cartulary-grid-group-row");
              semanticRow.dataset.gridRowKind = "group";
              semanticRow.dataset.gridPrimaryState = "group";
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
  const selectionColumn = compiledColumns.find(
    (column) => column.key === gridSelectionColumnKey,
  );
  const groupedColumns = [
    ...(selectionColumn === undefined ? [] : [selectionColumn]),
    groupColumn,
    ...compiledColumns.filter(
      (column) =>
        column.key !== gridRowGutterColumnKey &&
        column.key !== gridSelectionColumnKey,
    ),
  ];
  const vendorGroupedColumns = [
    groupColumn,
    ...groupedColumns.filter((column) => column.key !== groupColumn.key),
  ];
  const vendorGroupedColumnKeys = useStableStringArray(
    vendorGroupedColumns.map((column) => column.key),
  );
  const groupedFieldKeys = useStableStringArray(
    compiledColumns.flatMap((column) =>
      column.key === gridSelectionColumnKey ||
      column.key === gridRowGutterColumnKey ||
      column.key === gridActionsColumnKey
        ? []
        : [column.key],
    ),
  );
  const { rowIndexes, visibleRows } = useMemo(() => {
    const nextVisibleRows: Array<GridDataRow<Row>> = [];
    const nextRowIndexes = new Map<string, number>();
    let rowIdx = 0;
    for (const groupId of groupIds) {
      rowIdx += 1;
      if (!expandedGroupIds.has(groupId)) continue;
      for (const row of rowGroups[groupId] ?? []) {
        nextRowIndexes.set(gridRowIdentityKey(row.rowIdentity), rowIdx);
        nextVisibleRows.push(row);
        rowIdx += 1;
      }
    }
    return { rowIndexes: nextRowIndexes, visibleRows: nextVisibleRows };
  }, [expandedGroupIds, groupIds, rowGroups]);
  const groupedPositionMap = useMemo(
    () =>
      buildSemanticPositionMap({
        columnKeys: vendorGroupedColumnKeys,
        fieldKeys: groupedFieldKeys,
        dataRows: visibleRows,
        rowIndexes,
        surface,
      }),
    [
      groupedFieldKeys,
      rowIndexes,
      vendorGroupedColumnKeys,
      surface,
      visibleRows,
    ],
  );
  positionMapRef.current = groupedPositionMap;

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
          rowProps: RenderRowProps<GridDataRow<Row>, GridDraftRow<Row>>,
        ) => {
          const semanticState = resolveGridSemanticState(
            rowStateFor(rowProps.row),
            "data row",
          );
          return (
            <Row
              {...rowProps}
              {...semanticRowAttributes(semanticState)}
              data-cartulary-grid-row-kind="data"
              data-grid-row-identity-kind={rowProps.row.rowIdentity.kind}
              data-grid-record-id={
                rowProps.row.rowIdentity.kind === "core_record"
                  ? rowProps.row.rowIdentity.recordId
                  : undefined
              }
              data-inspector-active={
                semanticState.stateIds.includes("inspector-active")
                  ? "true"
                  : undefined
              }
              data-testid={rowProps.row.testId}
              key={key}
              viewportColumns={rowProps.viewportColumns.map((column) =>
                column.key === groupColumn.key &&
                typeof gutterRenderCell === "function"
                  ? { ...column, renderCell: gutterRenderCell }
                  : column,
              )}
            />
          );
        },
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
