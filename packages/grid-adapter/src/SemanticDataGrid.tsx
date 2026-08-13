import {
  cartularyDesignPresentation,
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
  formatGridClipboardTSV,
  type GridCellAnchor,
  type GridCellRange,
  type GridDataRow,
  type GridDraftRow,
  type GridEditorActivation,
  type GridHandle,
  type GridRowIdentity,
  type GridRowStateInput,
  type GridSortEntry,
  gridRowIdentitiesEqual,
  gridRowIdentityKey,
  gridSurfaceIdentitiesEqual,
  gridSurfaceIdentityKey,
  gridUnassignedGroupLabel,
  isGridColumnEditable,
  type SemanticDataGridProps,
} from "./core";
import {
  type ActiveEditorSession,
  editorSeedForTarget,
  type PendingEditorSeed,
} from "./editorSessionPolicy";
import {
  focusAdjacentOutsideGrid,
  isGridFillHandleTarget,
  isInteractiveCellActionTarget,
  isPrintableGridEntry,
  isSemanticNavigationKey,
  semanticAnchorFromDomTarget,
  visibleGridPageSize,
} from "./interactionPolicy";
import {
  compileGridColumns,
  type GridCompiledBulkSelection,
  gridActionsColumnKey,
  gridRowGutterColumnKey,
  gridSelectionColumnKey,
} from "./rdgCompiler";
import {
  buildSemanticGroupBuckets,
  buildSemanticPresentationModel,
  coreRecordId,
  coreRowVersion,
  dedupeGridTargets,
  emptySemanticPresentationModel,
  type GridSemanticPositionMap,
  type GridSemanticPresentationModel,
  gridAnchorKey,
  gridCellRangeContains,
  gridClipboardInputDimensions,
  isCoreRecordRow,
  isDataRow,
  navigateSemanticPresentation,
  planSemanticPasteTargets,
  resolveVisibleGridCellRange,
  sameGridCellAnchor,
  sameGridCellRange,
  semanticAnchor,
  semanticTarget,
} from "./semanticPresentation";
import {
  type GridResolvedSemanticState,
  gridSemanticStateClassNames,
  mergeGridSemanticState,
  resolveGridSemanticState,
} from "./semanticState";

const emptySelectedRecordIds: ReadonlySet<string> = new Set();

type SemanticCellRegistration = {
  readonly cell: HTMLElement;
  readonly token: object;
};

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
  enableVirtualization: boolean,
) {
  const {
    accessibleLabel,
    activeRowIdentity = null,
    allowPasteCreateRows = false,
    actionsColumn,
    clipboardPaste,
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
    onFillCells,
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
      onFillCells !== undefined ||
      clipboardPaste !== undefined)
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
  const dataRowsRef = useRef(dataRows);
  dataRowsRef.current = dataRows;
  const semanticPresentationRef = useRef<GridSemanticPresentationModel<Row>>(
    emptySemanticPresentationModel(surface),
  );
  const semanticCellElementsRef = useRef(
    new Map<string, SemanticCellRegistration>(),
  );
  const pendingEditorSeedRef = useRef<PendingEditorSeed | null>(null);
  const activeEditorSessionRef = useRef<ActiveEditorSession | null>(null);
  const pointerCellActivatorRef = useRef<{
    readonly activate: () => void;
    readonly anchorKey: string;
  } | null>(null);
  const pointerTransitionPendingRef = useRef(false);
  const pendingRangeEndRef = useRef<GridCellAnchor | null>(null);
  const [keyboardAnnouncement, setKeyboardAnnouncement] = useState("");
  const gridBusy =
    dataState.kind === "initial_loading" || dataState.kind === "refreshing";
  useEffect(() => {
    const element = vendorHandle.current?.element;
    if (element === null || element === undefined) return;
    element.setAttribute("aria-busy", String(gridBusy));
    element.setAttribute("aria-readonly", String(!editable));
    if (accessibleLabel === undefined) element.removeAttribute("aria-label");
    else element.setAttribute("aria-label", accessibleLabel);
  }, [accessibleLabel, editable, gridBusy]);
  useEffect(() => {
    const element = vendorHandle.current?.element;
    if (element === null || element === undefined) return;
    const labelFillHandles = () => {
      for (const handle of element.querySelectorAll(".rdg-cell-drag-handle")) {
        handle.setAttribute("aria-label", "Drag to fill this value");
        handle.setAttribute("data-cartulary-fill-handle", "true");
        handle.setAttribute("role", "img");
        handle.setAttribute("title", "Drag to fill this value");
      }
    };
    labelFillHandles();
    const observer = new MutationObserver(labelFillHandles);
    observer.observe(element, { childList: true, subtree: true });
    return () => observer.disconnect();
  }, []);
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
  const readEditorSeed = useCallback(
    (target: Parameters<typeof editorSeedForTarget>[1]) =>
      editorSeedForTarget(pendingEditorSeedRef.current, target),
    [],
  );
  const registerEditorSession = useCallback(
    (session: ActiveEditorSession | null) => {
      activeEditorSessionRef.current = session;
    },
    [],
  );
  const clearEditorSeed = useCallback(() => {
    pendingEditorSeedRef.current = null;
  }, []);
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
      const next = navigateSemanticPresentation(
        semanticPresentationRef.current,
        target,
        {
          key: action.rowDelta < 0 ? "ArrowUp" : "ArrowDown",
        },
      );
      focusOrScrollAnchor({
        anchor: next ?? target,
        cellElements: semanticCellElementsRef.current,
        focus: true,
        positionMap: semanticPresentationRef.current,
        vendorHandle: vendorHandle.current,
      });
    },
    [],
  );
  assertGridRows(dataRows);

  const handleSemanticPaste = useCallback(
    (
      row: GridDataRow<Row>,
      fieldKey: string,
      clipboardText: string,
    ): boolean => {
      if (!editable) return true;
      const target = semanticTarget(row, fieldKey, columns, surface);
      if (target === null) return false;
      if (clipboardPaste === undefined) return true;
      const input = clipboardPaste.decode(clipboardText);
      const dimensions = gridClipboardInputDimensions(input);
      if (dimensions === null) return true;
      const targetResolution = planSemanticPasteTargets(
        semanticPresentationRef.current,
        target,
        dimensions,
      );
      if (targetResolution === null) return true;
      const lastFieldKey = targetResolution.columns.at(-1);
      const lastRecordTarget = targetResolution.rowTargets
        .filter((candidate) => candidate.kind === "record")
        .at(-1);
      const range: GridCellRange =
        lastFieldKey !== undefined && lastRecordTarget !== undefined
          ? {
              start: target,
              end: {
                fieldKey: lastFieldKey,
                rowIdentity: lastRecordTarget.rowIdentity,
                surface,
              },
            }
          : { start: target, end: target };
      updateCellRange(range);
      clipboardPaste.onPaste({ input, range, target, targetResolution });
      return true;
    },
    [clipboardPaste, columns, editable, updateCellRange, surface],
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
        semanticPresentationRef.current,
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

  const dispatchSemanticFill = useCallback(
    (
      sourceRow: GridDataRow<Row>,
      targetRow: GridDataRow<Row>,
      columnKey: string,
    ): boolean => {
      const source = semanticTarget(sourceRow, columnKey, columns, surface);
      const target = semanticTarget(targetRow, columnKey, columns, surface);
      if (source === null || target === null) return false;
      const column = columns.find(
        (candidate) => candidate.fieldKey === columnKey,
      );
      const range: GridCellRange = {
        start: {
          fieldKey: source.fieldKey,
          rowIdentity: source.rowIdentity,
          surface: source.surface,
        },
        end: {
          fieldKey: target.fieldKey,
          rowIdentity: target.rowIdentity,
          surface: target.surface,
        },
      };
      const expanded = resolveVisibleGridCellRange({
        columns,
        dataRows,
        positionMap: semanticPresentationRef.current,
        range,
      });
      if (
        column?.contractWritable !== true ||
        column.editor === undefined ||
        column.valueKind === "collection" ||
        expanded === null ||
        expanded.fieldKeys.length !== 1
      ) {
        return false;
      }
      const targets = expanded.rowIdentities.flatMap((rowIdentity) => {
        if (gridRowIdentitiesEqual(rowIdentity, source.rowIdentity)) return [];
        const dataRow = dataRows.find((candidate) =>
          gridRowIdentitiesEqual(candidate.rowIdentity, rowIdentity),
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
      });
      if (
        targets.length === 0 ||
        targets.length !== expanded.rowIdentities.length - 1
      ) {
        return false;
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
              targets: dedupeGridTargets([...pending.targets, ...targets]),
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
      return true;
    },
    [columns, dataRows, surface, updateCellRange],
  );

  const compiledColumns = useMemo(
    () =>
      compileGridColumns({
        actionsColumn,
        bulkSelection: compiledBulkSelection,
        clearEditorSeed,
        cellStateFor,
        columns,
        readEditorSeed,
        editable,
        isCellRangeSelected,
        onEditorKeyboardAction: handleEditorKeyboardAction,
        onPasteCellContent:
          clipboardPaste === undefined ? undefined : handleSemanticPaste,
        registerEditorSession,
        registerSemanticCell,
        rowGutter,
        surface,
      }),
    [
      actionsColumn,
      cellStateFor,
      clearEditorSeed,
      clipboardPaste,
      columns,
      compiledBulkSelection,
      editable,
      handleEditorKeyboardAction,
      handleSemanticPaste,
      isCellRangeSelected,
      registerEditorSession,
      registerSemanticCell,
      readEditorSeed,
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
  const ungroupedPresentation = useMemo(
    () =>
      grouping === null
        ? buildSemanticPresentationModel({
            allowCreateRows: allowPasteCreateRows,
            columns,
            columnKeys: compiledColumnKeys,
            fieldKeys: semanticFieldKeys,
            dataRows,
            surface,
          })
        : null,
    [
      allowPasteCreateRows,
      columns,
      compiledColumnKeys,
      dataRows,
      grouping,
      semanticFieldKeys,
      surface,
    ],
  );
  if (ungroupedPresentation !== null) {
    semanticPresentationRef.current = ungroupedPresentation;
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
  const prepareEditorActivation = useCallback(
    (
      row: GridDataRow<Row>,
      fieldKey: string,
      activation: GridEditorActivation,
      seed?: { readonly value: unknown },
    ) => {
      if (!editable) return false;
      const target = semanticTarget(row, fieldKey, columns, surface);
      if (target === null) return false;
      pendingEditorSeedRef.current = {
        activation,
        anchor: target,
        baseRowVersion: target.mutationIdentity.baseRowVersion,
        hasValue: seed !== undefined,
        value: seed?.value,
      };
      return true;
    },
    [columns, editable, surface],
  );
  const prepareEditorActivationRef = useRef(prepareEditorActivation);
  prepareEditorActivationRef.current = prepareEditorActivation;
  useImperativeHandle(
    ref,
    () => ({
      activateEdit: (anchor, seed) => {
        const position = semanticPresentationRef.current.positions.get(
          gridAnchorKey(anchor),
        );
        const row = dataRowsRef.current.find((candidate) =>
          gridRowIdentitiesEqual(candidate.rowIdentity, anchor.rowIdentity),
        );
        if (
          position === undefined ||
          row === undefined ||
          !prepareEditorActivationRef.current(
            row,
            anchor.fieldKey,
            { initialSelection: "all", source: "programmatic" },
            seed,
          )
        ) {
          return false;
        }
        vendorHandle.current?.selectCell(position, { enableEditor: true });
        return true;
      },
      cancelEdit: (anchor) => {
        const session = activeEditorSessionRef.current;
        if (
          session === null ||
          gridAnchorKey(session.target) !== gridAnchorKey(anchor)
        ) {
          return false;
        }
        session.cancel();
        return true;
      },
      focusAnchor: (anchor) =>
        focusOrScrollAnchor({
          anchor,
          cellElements: semanticCellElementsRef.current,
          positionMap: semanticPresentationRef.current,
          vendorHandle: vendorHandle.current,
          focus: true,
        }),
      focusRoot: () => focusGridRoot(vendorHandle.current?.element ?? null),
      getScrollElement: () => vendorHandle.current?.element ?? null,
      isAnchorRendered: (anchor) =>
        semanticCellElementsRef.current.has(gridAnchorKey(anchor)),
      moveFocus: (current, intent) => {
        const next = navigateSemanticPresentation(
          semanticPresentationRef.current,
          current,
          intent,
        );
        if (next === null) return null;
        return focusOrScrollAnchor({
          anchor: next,
          cellElements: semanticCellElementsRef.current,
          positionMap: semanticPresentationRef.current,
          vendorHandle: vendorHandle.current,
          focus: true,
        })
          ? next
          : null;
      },
      planPasteTargets: (current, dimensions) =>
        planSemanticPasteTargets(
          semanticPresentationRef.current,
          current,
          dimensions,
        ),
      scrollToAnchor: (anchor) =>
        focusOrScrollAnchor({
          anchor,
          cellElements: semanticCellElementsRef.current,
          positionMap: semanticPresentationRef.current,
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
    enableVirtualization,
    headerRowHeight: 32,
    headerRowClass: "cartulary-grid-header-row",
    onCellMouseDown: (args, event) => {
      const session = activeEditorSessionRef.current;
      const destination = semanticAnchor(
        args.row,
        args.column.key,
        columns,
        surface,
      );
      if (session === null) {
        pointerCellActivatorRef.current =
          destination === null
            ? null
            : {
                activate: () => args.selectCell(true),
                anchorKey: gridAnchorKey(destination),
              };
        return;
      }
      if (
        destination === null ||
        sameGridCellAnchor(session.target, destination)
      ) {
        return;
      }
      event.preventDefault();
      event.preventGridDefault();
      if (pointerTransitionPendingRef.current) return;
      const destinationIsInteractive = isInteractiveCellActionTarget(
        event.target,
      );
      pointerTransitionPendingRef.current = true;
      void session.requestCommit().then((accepted) => {
        pointerTransitionPendingRef.current = false;
        if (!accepted) {
          session.focus();
          return;
        }
        const shouldEdit =
          !destinationIsInteractive &&
          prepareEditorActivation(args.row, args.column.key, {
            initialSelection: "end",
            source: "pointer",
          });
        args.selectCell(shouldEdit);
      });
    },
    onCellClick: ({ column, row, selectCell }, event) => {
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
      if (
        pointerTransitionPendingRef.current ||
        isInteractiveCellActionTarget(event.target)
      ) {
        return;
      }
      if (
        prepareEditorActivation(row, column.key, {
          initialSelection: "end",
          source: "pointer",
        })
      ) {
        selectCell(true);
      }
    },
    onCellCopy: ({ column, row }, event) => {
      const anchor = semanticAnchor(row, column.key, columns, surface);
      if (anchor === null) return;
      const range = cellRange ?? { end: anchor, start: anchor };
      const expandedRange = resolveVisibleGridCellRange({
        columns,
        positionMap: semanticPresentationRef.current,
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
    onCellKeyDown: (args, event) => {
      if (args.mode !== "SELECT" || !isDataRow<Row>(args.row)) return;
      const anchor = semanticAnchor(
        args.row,
        args.column.key,
        columns,
        surface,
      );
      if (anchor === null) return;

      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "d") {
        event.preventDefault();
        event.preventGridDefault();
        const range = cellRangeRef.current;
        const expanded =
          range === null
            ? null
            : resolveVisibleGridCellRange({
                columns,
                dataRows,
                positionMap: semanticPresentationRef.current,
                range,
              });
        const sourceRowIdentity = expanded?.rowIdentities[0];
        const targetRowIdentity = expanded?.rowIdentities.at(-1);
        const sourceRow = dataRows.find(
          (candidate) =>
            sourceRowIdentity !== undefined &&
            gridRowIdentitiesEqual(candidate.rowIdentity, sourceRowIdentity),
        );
        const targetRow = dataRows.find(
          (candidate) =>
            targetRowIdentity !== undefined &&
            gridRowIdentitiesEqual(candidate.rowIdentity, targetRowIdentity),
        );
        const filled =
          expanded?.fieldKeys.length === 1 &&
          expanded.rowIdentities.length > 1 &&
          sourceRow !== undefined &&
          targetRow !== undefined &&
          dispatchSemanticFill(
            sourceRow,
            targetRow,
            expanded.fieldKeys[0] ?? "",
          );
        setKeyboardAnnouncement(
          filled
            ? `Filled ${expanded.rowIdentities.length - 1} cells from the top selected cell.`
            : "Select a writable one-column range before using fill down.",
        );
        return;
      }

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
        const timelineSummaryEnter =
          event.key === "Enter" &&
          surface.kind === "view_schema" &&
          surface.viewSchemaId === "cartulary.view.timeline.v2" &&
          anchor.fieldKey === "timeline.activity_synopsis_text";
        pendingEditorSeedRef.current = {
          activation: {
            initialSelection: hasSeed
              ? "seed"
              : event.shiftKey || timelineSummaryEnter
                ? "end"
                : "all",
            source: hasSeed
              ? event.key === "Backspace" || event.key === "Delete"
                ? "clear"
                : "printable"
              : event.shiftKey
                ? "shift_enter"
                : "enter",
          },
          anchor,
          baseRowVersion,
          hasValue: hasSeed,
          value: seed,
        };
        if (timelineSummaryEnter) {
          performance.mark("cartulary.workbook.focus_edit_accepted", {
            detail: { field: anchor.fieldKey, surface: surface.viewSchemaId },
          });
        }
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
      const next = navigateSemanticPresentation(
        semanticPresentationRef.current,
        anchor,
        {
          ctrlOrMetaKey: event.ctrlKey || event.metaKey,
          key: event.key,
          pageSize: visibleGridPageSize(
            vendorHandle.current?.element ?? null,
            workbookGridRowHeightPx(density),
          ),
          shiftKey: event.shiftKey,
        },
      );
      if (next === null) return;
      const nextPosition = semanticPresentationRef.current.positions.get(
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
      if (
        event.key === "ArrowDown" &&
        !event.shiftKey &&
        !event.ctrlKey &&
        !event.metaKey &&
        surface.kind === "view_schema" &&
        surface.viewSchemaId === "cartulary.view.timeline.v2" &&
        anchor.fieldKey === "timeline.activity_synopsis_text"
      ) {
        performance.mark("cartulary.workbook.selection_change_accepted", {
          detail: { field: anchor.fieldKey, surface: surface.viewSchemaId },
        });
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
            dispatchSemanticFill(sourceRow, targetRow, columnKey);
            return targetRow;
          }
        : undefined,
    onRowsChange:
      !editable ||
      (clipboardPaste === undefined &&
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
        presentationRef={semanticPresentationRef}
        rowStateFor={rowStateFor}
        sharedProps={sharedProps}
      />
    );

  return (
    <div
      className="cartulary-grid-state-frame"
      style={gridStateFrameStyle}
      onMouseDownCapture={(event) => {
        pointerCellActivatorRef.current = null;
        if (
          event.button !== 0 ||
          event.altKey ||
          event.ctrlKey ||
          event.metaKey ||
          event.shiftKey ||
          isInteractiveCellActionTarget(event.target) ||
          isGridFillHandleTarget(event.target)
        ) {
          return;
        }
        const anchor = semanticAnchorFromDomTarget(event.target, surface);
        if (anchor === null) return;
        const startX = event.clientX;
        const startY = event.clientY;
        window.addEventListener(
          "mouseup",
          (mouseUpEvent) => {
            if (
              mouseUpEvent.button !== 0 ||
              Math.abs(mouseUpEvent.clientX - startX) > 4 ||
              Math.abs(mouseUpEvent.clientY - startY) > 4
            ) {
              return;
            }
            const row = dataRows.find((candidate) =>
              gridRowIdentitiesEqual(candidate.rowIdentity, anchor.rowIdentity),
            );
            const position = semanticPresentationRef.current.positions.get(
              gridAnchorKey(anchor),
            );
            if (row !== undefined && position !== undefined) {
              window.setTimeout(() => {
                if (
                  activeEditorSessionRef.current === null &&
                  prepareEditorActivation(row, anchor.fieldKey, {
                    initialSelection: "end",
                    source: "pointer",
                  })
                ) {
                  const pointerActivator = pointerCellActivatorRef.current;
                  pointerCellActivatorRef.current = null;
                  if (pointerActivator?.anchorKey === gridAnchorKey(anchor)) {
                    pointerActivator.activate();
                  } else {
                    vendorHandle.current?.selectCell(position, {
                      enableEditor: true,
                    });
                  }
                }
              }, 0);
            }
          },
          { capture: true, once: true },
        );
      }}
      onDoubleClickCapture={(event) => {
        if (!isGridFillHandleTarget(event.target)) return;
        event.preventDefault();
        event.stopPropagation();
      }}
    >
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
        positionMap={semanticPresentationRef.current}
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
  const delayedInitialLoading = useDelayedInitialLoading(dataState);
  const presentation = gridDataStatePresentation(
    dataState,
    delayedInitialLoading,
  );
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

function useDelayedInitialLoading(
  state: NonNullable<SemanticDataGridProps<unknown>["dataState"]>,
): boolean {
  const [delayed, setDelayed] = useState(false);
  const generationKey =
    state.kind === "initial_loading" ? state.generationKey : null;
  useEffect(() => {
    setDelayed(false);
    if (generationKey === null) return;
    const timeout = window.setTimeout(() => {
      setDelayed(true);
    }, cartularyDesignPresentation.initialLoading.delayMs);
    return () => window.clearTimeout(timeout);
  }, [generationKey]);
  return delayed;
}

function gridDataStatePresentation(
  state: NonNullable<SemanticDataGridProps<unknown>["dataState"]>,
  delayedInitialLoading = false,
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
        message: delayedInitialLoading
          ? cartularyDesignPresentation.initialLoading.message
          : `Loading ${state.surfaceLabel}…`,
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
  return useSemanticDataGrid(props, ref, true);
}

function SemanticDataGridDomUnitInner<Row>(
  props: SemanticDataGridProps<Row>,
  ref: ForwardedRef<GridHandle>,
) {
  // biome-ignore lint/correctness/useHookAtTopLevel: this generic function is passed directly to React.forwardRef below.
  return useSemanticDataGrid(props, ref, false);
}

export const SemanticDataGrid = forwardRef(SemanticDataGridInner) as <Row>(
  props: SemanticDataGridProps<Row> & RefAttributes<GridHandle>,
) => ReactElement;

export const SemanticDataGridDomUnitBinding = forwardRef(
  SemanticDataGridDomUnitInner,
) as <Row>(
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

function focusGridRoot(gridElement: HTMLDivElement | null): boolean {
  if (gridElement === null) return false;
  if (!gridElement.hasAttribute("tabindex")) gridElement.tabIndex = 0;
  gridElement.focus({ preventScroll: true });
  return document.activeElement === gridElement;
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

function GroupedSemanticDataGrid<Row>({
  columns,
  grouping,
  density = "default",
  presentationRef,
  rowStateFor,
  sharedProps,
  surface,
}: SemanticDataGridProps<Row> & {
  readonly sharedProps: DataGridProps<
    GridDataRow<Row>,
    GridDraftRow<Row>,
    string
  >;
  readonly presentationRef: MutableRefObject<
    GridSemanticPresentationModel<Row>
  >;
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
  const groupBuckets = useMemo(
    () => buildSemanticGroupBuckets(sharedProps.rows, grouping),
    [grouping, sharedProps.rows],
  );
  const { groupIds, metadata, rowGroups } = useMemo(
    () => ({
      groupIds: groupBuckets.map((bucket) => bucket.id),
      metadata: new Map(
        groupBuckets.map((bucket) => [
          bucket.id,
          { label: bucket.label, value: bucket.value },
        ]),
      ),
      rowGroups: Object.fromEntries(
        groupBuckets.map((bucket) => [bucket.id, bucket.rows]),
      ),
    }),
    [groupBuckets],
  );
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
  const groupedPresentation = useMemo(
    () =>
      buildSemanticPresentationModel({
        allowCreateRows: false,
        columns,
        columnKeys: vendorGroupedColumnKeys,
        fieldKeys: groupedFieldKeys,
        grouping: {
          buckets: groupBuckets,
          collapsedGroupIds,
          scope: expansionScope,
        },
        dataRows: visibleRows,
        rowIndexes,
        surface,
      }),
    [
      columns,
      collapsedGroupIds,
      expansionScope,
      groupBuckets,
      groupedFieldKeys,
      rowIndexes,
      vendorGroupedColumnKeys,
      surface,
      visibleRows,
    ],
  );
  presentationRef.current = groupedPresentation;

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
