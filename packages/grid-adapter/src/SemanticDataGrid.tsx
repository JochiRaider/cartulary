import {
  cartularyDesignPresentation,
  gridScrollportClassName,
  workbookGridDensityMetrics,
  workbookGridRowHeightPx,
} from "@cartulary/ui-contracts";
import type {
  CSSProperties,
  ForwardedRef,
  MutableRefObject,
  ReactElement,
  MouseEvent as ReactMouseEvent,
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
  type GridCellAnchor,
  type GridCellRange,
  type GridDataRow,
  type GridDensity,
  type GridDraftRow,
  type GridEditorActivation,
  type GridEditorFocusTarget,
  type GridFillIntent,
  type GridHandle,
  type GridRowStateInput,
  type GridSortEntry,
  type GridSurfaceIdentity,
  gridRowIdentitiesEqual,
  gridRowIdentityKey,
  gridSurfaceIdentitiesEqual,
  gridSurfaceIdentityKey,
  gridUnassignedGroupLabel,
  isGridColumnEditable,
  type SemanticDataGridProps,
} from "./core";
import {
  focusAdjacentOutsideGrid,
  isGridFillHandleTarget,
  isInteractiveCellActionTarget,
  semanticAnchorFromDomTarget,
  visibleGridPageSize,
} from "./domInteraction";
import {
  type ActiveEditorSession,
  editorSeedForTarget,
  type PendingEditorSeed,
} from "./editorSessionPolicy";
import {
  compileGridColumns,
  type GridCompiledBulkSelection,
  gridActionsColumnKey,
  gridRowGutterColumnKey,
  gridSelectionColumnKey,
} from "./rdgCompiler";
import {
  buildRdgPresentationModel,
  emptyRdgPresentationModel,
  type GridRdgPositionMap,
  type GridRdgPresentationModel,
} from "./rdgPositionMap";
import { decideSemanticActiveCellTransition } from "./semanticActiveCellPolicy";
import { resolveSemanticGridCapabilities } from "./semanticCapabilities";
import {
  mergeSemanticFillIntents,
  planSemanticCopy,
  planSemanticFill,
  planSemanticFillFromRange,
  planSemanticPaste,
} from "./semanticClipboardPolicy";
import { resolveGridDataStatePresentation } from "./semanticDataState";
import {
  decideSemanticGridKey,
  normalizeGridKey,
  type SemanticGridDecision,
} from "./semanticKeyboardPolicy";
import {
  buildSemanticGroupBuckets,
  coreRecordId,
  type GridSemanticPresentationModel,
  gridAnchorKey,
  gridCellRangeContains,
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
  resolveSemanticBulkSelection,
  toggleAllSemanticRecords,
  toggleSemanticRecordRange,
} from "./semanticSelectionPolicy";
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

function useGridDomPresentation(
  vendorHandle: MutableRefObject<DataGridHandle | null>,
  accessibleLabel: string | undefined,
  busy: boolean,
  editable: boolean,
): void {
  useEffect(() => {
    const element = vendorHandle.current?.element;
    if (element === null || element === undefined) return;
    element.setAttribute("aria-busy", String(busy));
    element.setAttribute("aria-readonly", String(!editable));
    if (accessibleLabel === undefined) element.removeAttribute("aria-label");
    else element.setAttribute("aria-label", accessibleLabel);
  }, [accessibleLabel, busy, editable, vendorHandle]);
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
  }, [vendorHandle]);
}

function useGridRangeController(
  controlledRange: GridCellRange | null | undefined,
  onRangeChange: ((range: GridCellRange | null) => void) | undefined,
) {
  const [internalRange, setInternalRange] = useState<GridCellRange | null>(
    controlledRange ?? null,
  );
  const range = controlledRange === undefined ? internalRange : controlledRange;
  const rangeRef = useRef(range);
  rangeRef.current = range;
  const updateRange = useCallback(
    (next: GridCellRange | null) => {
      if (sameGridCellRange(rangeRef.current, next)) return;
      rangeRef.current = next;
      if (controlledRange === undefined) setInternalRange(next);
      onRangeChange?.(next);
    },
    [controlledRange, onRangeChange],
  );
  return { range, rangeRef, updateRange };
}

function useActiveCellController(
  onChange: ((anchor: GridCellAnchor | null) => void) | undefined,
) {
  const [activeCell, setActiveCell] = useState<GridCellAnchor | null>(null);
  const activeCellRef = useRef<GridCellAnchor | null>(null);
  const publishActiveCell = useCallback(
    (next: GridCellAnchor | null) => {
      const transition = decideSemanticActiveCellTransition(
        activeCellRef.current,
        next,
      );
      if (transition.kind === "no_change") return false;
      activeCellRef.current = transition.anchor;
      setActiveCell(transition.anchor);
      onChange?.(transition.anchor);
      return true;
    },
    [onChange],
  );
  return { activeCell, publishActiveCell };
}

function useGridBulkSelection<Row>(
  rows: readonly GridDataRow<Row>[],
  bulkSelection: SemanticDataGridProps<Row>["coreRecordBulkSelection"],
  editable: boolean,
): {
  readonly compiled: GridCompiledBulkSelection<Row> | undefined;
  readonly selectedRows: ReadonlySet<string>;
} {
  const anchorRecordId = useRef<string | null>(null);
  const state = useMemo(
    () =>
      bulkSelection === undefined || !editable
        ? null
        : resolveSemanticBulkSelection(
            rows,
            bulkSelection.selectedRecordIds,
            bulkSelection.isRecordSelectable,
          ),
    [bulkSelection, editable, rows],
  );
  const compiled = useMemo<GridCompiledBulkSelection<Row> | undefined>(() => {
    if (bulkSelection === undefined || state === null) return undefined;
    return {
      allSelected: state.allSelected,
      partiallySelected: state.partiallySelected,
      selectedRecordIds: bulkSelection.selectedRecordIds,
      selectableRecordCount: state.selectableIds.length,
      isRecordSelectable: (row) =>
        isCoreRecordRow(row) &&
        bulkSelection.isRecordSelectable?.(row) !== false,
      onSelectAll: () => {
        anchorRecordId.current = null;
        bulkSelection.onSelectedRecordIdsChange(
          toggleAllSemanticRecords(state),
        );
      },
      onToggleRecord: (row, shiftKey) => {
        if (!isCoreRecordRow(row)) return;
        const next = toggleSemanticRecordRange({
          anchorRecordId: anchorRecordId.current,
          recordId: row.rowIdentity.recordId,
          selectableRows: state.selectableRows,
          selectedRecordIds: bulkSelection.selectedRecordIds,
          shiftKey,
        });
        anchorRecordId.current = row.rowIdentity.recordId;
        bulkSelection.onSelectedRecordIdsChange(next);
      },
    };
  }, [bulkSelection, state]);
  return {
    compiled,
    selectedRows:
      editable && bulkSelection !== undefined
        ? bulkSelection.selectedRecordIds
        : emptySelectedRecordIds,
  };
}

function useGridSemanticState<Row>({
  activeCellAnchor,
  activeRowIdentity,
  columns,
  dataState,
  editable,
  getCellState,
  getRowState,
  selectedRows,
  surface,
}: Pick<
  SemanticDataGridProps<Row>,
  | "activeRowIdentity"
  | "columns"
  | "dataState"
  | "getCellState"
  | "getRowState"
  | "surface"
> & {
  readonly activeCellAnchor: GridCellAnchor | null;
  readonly editable: boolean;
  readonly selectedRows: ReadonlySet<string>;
}) {
  const stale = dataState?.kind === "stale_error";
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
          activeRowIdentity !== undefined &&
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
      const rowState = rowStateFor(row);
      return mergeGridSemanticState(
        getCellState?.({
          anchor: {
            fieldKey: column.fieldKey,
            rowIdentity: row.rowIdentity,
            surface,
          },
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
          pending: false,
          readOnlyOrDerived: !editable || !isGridColumnEditable(column),
          saved: true,
          stale: false,
        },
      );
    },
    [activeCellAnchor, editable, getCellState, rowStateFor, surface],
  );
  return { cellStateFor, rowStateFor };
}

function useGridRegistration<Row>(
  surface: GridSurfaceIdentity,
  rangeRef: MutableRefObject<GridCellRange | null>,
  presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>,
) {
  const cellElementsRef = useRef(new Map<string, SemanticCellRegistration>());
  const draftFocusTargetsRef = useRef(new Map<string, GridEditorFocusTarget>());
  const isCellRangeSelected = useCallback(
    (
      row: GridDataRow<Row>,
      column: SemanticDataGridProps<Row>["columns"][number],
    ) =>
      gridCellRangeContains(presentationRef.current, rangeRef.current, {
        fieldKey: column.fieldKey,
        rowIdentity: row.rowIdentity,
        surface,
      }),
    [presentationRef, rangeRef, surface],
  );
  const registerSemanticCell = useCallback(
    (anchor: GridCellAnchor, cell: HTMLElement | null, token: object) => {
      const key = gridAnchorKey(anchor);
      if (cell === null) {
        if (cellElementsRef.current.get(key)?.token === token) {
          cellElementsRef.current.delete(key);
        }
        return;
      }
      cellElementsRef.current.set(key, { cell, token });
    },
    [],
  );
  const draftFocusTargetRef = useCallback(
    (fieldKey: string) =>
      createDraftFocusTargetRef(draftFocusTargetsRef.current, fieldKey),
    [],
  );
  return {
    cellElementsRef,
    draftFocusTargetRef,
    draftFocusTargetsRef,
    isCellRangeSelected,
    registerSemanticCell,
  };
}

function useGridEditorController<Row>(
  presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>,
  vendorHandle: MutableRefObject<DataGridHandle | null>,
  cellElementsRef: MutableRefObject<Map<string, SemanticCellRegistration>>,
) {
  const pendingEditorSeedRef = useRef<PendingEditorSeed | null>(null);
  const activeEditorSessionRef = useRef<ActiveEditorSession | null>(null);
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
        presentationRef.current,
        target,
        {
          key: action.rowDelta < 0 ? "ArrowUp" : "ArrowDown",
        },
      );
      focusOrScrollAnchor({
        anchor: next ?? target,
        cellElements: cellElementsRef.current,
        focus: true,
        positionMap: presentationRef.current,
        vendorHandle: vendorHandle.current,
      });
    },
    [cellElementsRef, presentationRef, vendorHandle],
  );
  return {
    activeEditorSessionRef,
    clearEditorSeed,
    handleEditorKeyboardAction,
    pendingEditorSeedRef,
    readEditorSeed,
    registerEditorSession,
  };
}

function useGridPasteController<Row>({
  clipboardPaste,
  columns,
  editable,
  presentationRef,
  surface,
  updateRange,
}: Pick<
  SemanticDataGridProps<Row>,
  "clipboardPaste" | "columns" | "surface"
> & {
  readonly editable: boolean;
  readonly presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>;
  readonly updateRange: (range: GridCellRange | null) => void;
}) {
  return useCallback(
    (
      row: GridDataRow<Row>,
      fieldKey: string,
      clipboardText: string,
    ): boolean => {
      if (!editable) return true;
      const target = semanticTarget(row, fieldKey, columns, surface);
      if (target === null) return false;
      if (clipboardPaste === undefined) return true;
      const intent = planSemanticPaste({
        input: clipboardPaste.decode(clipboardText),
        model: presentationRef.current,
        target,
      });
      if (intent === null) return true;
      updateRange(intent.range);
      clipboardPaste.onPaste(intent);
      return true;
    },
    [clipboardPaste, columns, editable, presentationRef, surface, updateRange],
  );
}

function useGridFillController<Row>({
  columns,
  dataRows,
  onFillCells,
  presentationRef,
  surface,
  updateRange,
}: Pick<
  SemanticDataGridProps<Row>,
  "columns" | "dataRows" | "onFillCells" | "surface"
> & {
  readonly presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>;
  readonly updateRange: (range: GridCellRange | null) => void;
}) {
  const pendingIntent = useRef<GridFillIntent | null>(null);
  const dispatchScheduled = useRef(false);
  const onFillRef = useRef(onFillCells);
  onFillRef.current = onFillCells;
  const dispatchFillIntent = useCallback(
    (intent: GridFillIntent) => {
      pendingIntent.current = mergeSemanticFillIntents(
        pendingIntent.current,
        intent,
      );
      updateRange(intent.range);
      if (dispatchScheduled.current) return;
      dispatchScheduled.current = true;
      queueMicrotask(() => {
        dispatchScheduled.current = false;
        const next = pendingIntent.current;
        pendingIntent.current = null;
        if (next !== null) onFillRef.current?.(next);
      });
    },
    [updateRange],
  );
  const dispatchSemanticFill = useCallback(
    (
      sourceRow: GridDataRow<Row>,
      targetRow: GridDataRow<Row>,
      columnKey: string,
    ): boolean => {
      const intent = planSemanticFill({
        columnKey,
        columns,
        dataRows,
        model: presentationRef.current,
        sourceRow,
        surface,
        targetRow,
      });
      if (intent === null) return false;
      dispatchFillIntent(intent);
      return true;
    },
    [columns, dataRows, dispatchFillIntent, presentationRef, surface],
  );
  return { dispatchFillIntent, dispatchSemanticFill };
}

function useGridPointerController<Row>({
  activeEditorSessionRef,
  dataRows,
  prepareEditorActivation,
  presentationRef,
  surface,
  vendorHandle,
}: {
  readonly activeEditorSessionRef: MutableRefObject<ActiveEditorSession | null>;
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly prepareEditorActivation: (
    row: GridDataRow<Row>,
    fieldKey: string,
    activation: GridEditorActivation,
  ) => boolean;
  readonly presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>;
  readonly surface: GridSurfaceIdentity;
  readonly vendorHandle: MutableRefObject<DataGridHandle | null>;
}) {
  const cellActivatorRef = useRef<{
    readonly activate: () => void;
    readonly anchorKey: string;
  } | null>(null);
  const transitionPendingRef = useRef(false);
  const onMouseDownCapture = useCallback(
    (event: ReactMouseEvent<HTMLDivElement>) => {
      cellActivatorRef.current = null;
      if (!isUnmodifiedCellPointer(event)) return;
      const anchor = semanticAnchorFromDomTarget(event.target, surface);
      if (anchor === null) return;
      const origin = { x: event.clientX, y: event.clientY };
      window.addEventListener(
        "mouseup",
        (mouseUpEvent) => {
          if (!isStationaryPrimaryRelease(mouseUpEvent, origin)) return;
          const row = dataRows.find((candidate) =>
            gridRowIdentitiesEqual(candidate.rowIdentity, anchor.rowIdentity),
          );
          const position = presentationRef.current.positions.get(
            gridAnchorKey(anchor),
          );
          if (row === undefined || position === undefined) return;
          window.setTimeout(() => {
            if (
              activeEditorSessionRef.current !== null ||
              !prepareEditorActivation(row, anchor.fieldKey, {
                initialSelection: "end",
                source: "pointer",
              })
            ) {
              return;
            }
            const activator = cellActivatorRef.current;
            cellActivatorRef.current = null;
            if (activator?.anchorKey === gridAnchorKey(anchor)) {
              activator.activate();
              return;
            }
            vendorHandle.current?.selectCell(position, { enableEditor: true });
          }, 0);
        },
        { capture: true, once: true },
      );
    },
    [
      activeEditorSessionRef,
      dataRows,
      prepareEditorActivation,
      presentationRef,
      surface,
      vendorHandle,
    ],
  );
  const onDoubleClickCapture = useCallback(
    (event: ReactMouseEvent<HTMLDivElement>) => {
      if (!isGridFillHandleTarget(event.target)) return;
      event.preventDefault();
      event.stopPropagation();
    },
    [],
  );
  return {
    cellActivatorRef,
    onDoubleClickCapture,
    onMouseDownCapture,
    transitionPendingRef,
  };
}

function isUnmodifiedCellPointer(
  event: ReactMouseEvent<HTMLDivElement>,
): boolean {
  return (
    event.button === 0 &&
    !event.altKey &&
    !event.ctrlKey &&
    !event.metaKey &&
    !event.shiftKey &&
    !isInteractiveCellActionTarget(event.target) &&
    !isGridFillHandleTarget(event.target)
  );
}

function isStationaryPrimaryRelease(
  event: MouseEvent,
  origin: { readonly x: number; readonly y: number },
): boolean {
  return (
    event.button === 0 &&
    Math.abs(event.clientX - origin.x) <= 4 &&
    Math.abs(event.clientY - origin.y) <= 4
  );
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
  const capabilities = resolveSemanticGridCapabilities(props);
  const effectiveInteractionMode = capabilities.interactionMode;
  const editable = capabilities.editable;
  const vendorHandle = useRef<DataGridHandle>(null);
  const dataRowsRef = useRef(dataRows);
  dataRowsRef.current = dataRows;
  const semanticPresentationRef = useRef<GridRdgPresentationModel<Row>>(
    emptyRdgPresentationModel(surface),
  );
  const pendingRangeEndRef = useRef<GridCellAnchor | null>(null);
  const [keyboardAnnouncement, setKeyboardAnnouncement] = useState("");
  const gridBusy =
    dataState.kind === "initial_loading" || dataState.kind === "refreshing";
  useGridDomPresentation(vendorHandle, accessibleLabel, gridBusy, editable);
  const { activeCell: activeCellAnchor, publishActiveCell } =
    useActiveCellController(onActiveCellChange);
  const {
    range: cellRange,
    rangeRef: cellRangeRef,
    updateRange: updateCellRange,
  } = useGridRangeController(controlledCellRange, onCellRangeChange);
  const {
    cellElementsRef: semanticCellElementsRef,
    draftFocusTargetRef,
    draftFocusTargetsRef,
    isCellRangeSelected,
    registerSemanticCell,
  } = useGridRegistration(surface, cellRangeRef, semanticPresentationRef);
  const {
    activeEditorSessionRef,
    clearEditorSeed,
    handleEditorKeyboardAction,
    pendingEditorSeedRef,
    readEditorSeed,
    registerEditorSession,
  } = useGridEditorController(
    semanticPresentationRef,
    vendorHandle,
    semanticCellElementsRef,
  );
  assertGridRows(dataRows);

  const handleSemanticPaste = useGridPasteController({
    clipboardPaste,
    columns,
    editable,
    presentationRef: semanticPresentationRef,
    surface,
    updateRange: updateCellRange,
  });

  const { compiled: compiledBulkSelection, selectedRows } =
    useGridBulkSelection(dataRows, coreRecordBulkSelection, editable);
  const { cellStateFor, rowStateFor } = useGridSemanticState({
    activeCellAnchor,
    activeRowIdentity,
    columns,
    dataState,
    editable,
    getCellState,
    getRowState,
    selectedRows,
    surface,
  });
  const { dispatchFillIntent, dispatchSemanticFill } = useGridFillController({
    columns,
    dataRows,
    onFillCells,
    presentationRef: semanticPresentationRef,
    surface,
    updateRange: updateCellRange,
  });

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
        draftFocusTargetRef,
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
      draftFocusTargetRef,
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
        ? buildRdgPresentationModel({
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
    [columns, editable, pendingEditorSeedRef, surface],
  );
  const prepareEditorActivationRef = useRef(prepareEditorActivation);
  prepareEditorActivationRef.current = prepareEditorActivation;
  const {
    cellActivatorRef: pointerCellActivatorRef,
    onDoubleClickCapture,
    onMouseDownCapture,
    transitionPendingRef: pointerTransitionPendingRef,
  } = useGridPointerController({
    activeEditorSessionRef,
    dataRows,
    prepareEditorActivation,
    presentationRef: semanticPresentationRef,
    surface,
    vendorHandle,
  });
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
      focusDraftCell: (fieldKey) =>
        focusRegisteredDraftCell(draftFocusTargetsRef.current, fieldKey),
      focusRoot: () => focusGridRoot(vendorHandle.current?.element ?? null),
      getAnchorRect: (anchor) =>
        registeredSemanticAnchorRect(
          semanticCellElementsRef.current,
          semanticPresentationRef.current,
          anchor,
        ),
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
    [activeEditorSessionRef, draftFocusTargetsRef, semanticCellElementsRef],
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
    headerRowHeight: workbookGridRowHeightPx(density),
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
      publishActiveCell(anchor);
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
      const plan = planSemanticCopy({
        anchor,
        columns,
        dataRows,
        model: semanticPresentationRef.current,
        range,
      });
      if (plan === null) return;
      event.clipboardData?.setData("text/plain", plan.text);
      event.preventDefault();
      onCopyCell?.(plan.intent);
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
      const semanticColumn = columns.find(
        (column) => column.fieldKey === anchor.fieldKey,
      );
      const decision = decideSemanticGridKey({
        anchor,
        column: semanticColumn,
        editable,
        input: normalizeGridKey(event),
        model: semanticPresentationRef.current,
        pageSize: visibleGridPageSize(
          vendorHandle.current?.element ?? null,
          workbookGridRowHeightPx(density),
        ),
        range: cellRangeRef.current,
        readOnlyLabel:
          effectiveInteractionMode.kind === "read_only"
            ? effectiveInteractionMode.label
            : "This workbook is read-only.",
        row: args.row,
      });
      if (decision.kind === "ignore") return;
      event.preventDefault();
      event.preventGridDefault();
      executeSemanticKeyDecision({
        args,
        columns,
        dataRows,
        decision,
        dispatchFillIntent,
        gridElement: vendorHandle.current?.element ?? null,
        pendingEditorSeedRef,
        pendingRangeEndRef,
        positionMap: semanticPresentationRef.current,
        setKeyboardAnnouncement,
        surface,
        updateCellRange,
      });
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
      publishActiveCell(anchor);
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
      "--cartulary-grid-cell-padding-block": `${workbookGridDensityMetrics(density).cellPaddingBlockCssPx}px`,
      "--cartulary-grid-cell-padding-inline": `${workbookGridDensityMetrics(density).cellPaddingInlineCssPx}px`,
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

  return (
    <GridBindingFrame
      bulkSelectionEnabled={coreRecordBulkSelection !== undefined}
      columns={columns}
      dataRows={dataRows}
      dataState={dataState}
      interactionMode={effectiveInteractionMode}
      keyboardAnnouncement={keyboardAnnouncement}
      onDoubleClickCapture={onDoubleClickCapture}
      onMouseDownCapture={onMouseDownCapture}
      positionMap={semanticPresentationRef.current}
      range={cellRange}
      selectedRecordCount={selectedRows.size}
    >
      <ProductionGridBinding
        density={density}
        grouping={grouping}
        presentationRef={semanticPresentationRef}
        props={props}
        rowStateFor={rowStateFor}
        sharedProps={sharedProps}
      />
    </GridBindingFrame>
  );
}

function ProductionGridBinding<Row>({
  density,
  grouping,
  presentationRef,
  props,
  rowStateFor,
  sharedProps,
}: {
  readonly density: GridDensity;
  readonly grouping: SemanticDataGridProps<Row>["grouping"];
  readonly presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>;
  readonly props: SemanticDataGridProps<Row>;
  readonly rowStateFor: (row: GridDataRow<Row>) => GridRowStateInput;
  readonly sharedProps: DataGridProps<
    GridDataRow<Row>,
    GridDraftRow<Row>,
    string
  >;
}) {
  if (grouping === null || grouping === undefined) {
    return (
      <DataGrid {...sharedProps} rowHeight={workbookGridRowHeightPx(density)} />
    );
  }
  return (
    <GroupedSemanticDataGrid
      {...props}
      density={density}
      presentationRef={presentationRef}
      rowStateFor={rowStateFor}
      sharedProps={sharedProps}
    />
  );
}

function GridBindingFrame<Row>({
  bulkSelectionEnabled,
  children,
  columns,
  dataRows,
  dataState,
  interactionMode,
  keyboardAnnouncement,
  onDoubleClickCapture,
  onMouseDownCapture,
  positionMap,
  range,
  selectedRecordCount,
}: {
  readonly bulkSelectionEnabled: boolean;
  readonly children: ReactElement;
  readonly columns: readonly SemanticDataGridProps<Row>["columns"][number][];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly dataState: NonNullable<SemanticDataGridProps<Row>["dataState"]>;
  readonly interactionMode: NonNullable<
    SemanticDataGridProps<Row>["interactionMode"]
  >;
  readonly keyboardAnnouncement: string;
  readonly onDoubleClickCapture: (
    event: ReactMouseEvent<HTMLDivElement>,
  ) => void;
  readonly onMouseDownCapture: (event: ReactMouseEvent<HTMLDivElement>) => void;
  readonly positionMap: GridSemanticPresentationModel<Row>;
  readonly range: GridCellRange | null;
  readonly selectedRecordCount: number;
}) {
  return (
    <div
      className="cartulary-grid-state-frame"
      style={gridStateFrameStyle}
      onDoubleClickCapture={onDoubleClickCapture}
      onMouseDownCapture={onMouseDownCapture}
    >
      {children}
      <GridStatePresentation
        dataState={dataState}
        interactionMode={interactionMode}
      />
      {bulkSelectionEnabled ? (
        <span
          aria-live="polite"
          className="cartulary-grid-live-region"
          role="status"
        >
          {bulkSelectionMessage(selectedRecordCount)}
        </span>
      ) : null}
      <GridRangeAnnouncement
        columns={columns}
        dataRows={dataRows}
        positionMap={positionMap}
        range={range}
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

function bulkSelectionMessage(selectedRecordCount: number): string {
  if (selectedRecordCount === 0) return "No records selected.";
  return `${selectedRecordCount} ${selectedRecordCount === 1 ? "record" : "records"} selected.`;
}

function executeSemanticKeyDecision<Row>({
  args,
  columns,
  dataRows,
  decision,
  dispatchFillIntent,
  gridElement,
  pendingEditorSeedRef,
  pendingRangeEndRef,
  positionMap,
  setKeyboardAnnouncement,
  surface,
  updateCellRange,
}: {
  readonly args: {
    readonly column: { readonly idx: number };
    readonly rowIdx: number;
    readonly selectCell: (
      position: { readonly idx: number; readonly rowIdx: number },
      options?: {
        readonly enableEditor?: boolean;
        readonly shouldFocusCell?: boolean;
      },
    ) => void;
  };
  readonly columns: readonly SemanticDataGridProps<Row>["columns"][number][];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly decision: SemanticGridDecision;
  readonly dispatchFillIntent: (intent: GridFillIntent) => void;
  readonly gridElement: HTMLDivElement | null;
  readonly pendingEditorSeedRef: MutableRefObject<PendingEditorSeed | null>;
  readonly pendingRangeEndRef: MutableRefObject<GridCellAnchor | null>;
  readonly positionMap: GridRdgPresentationModel<Row>;
  readonly setKeyboardAnnouncement: (value: string) => void;
  readonly surface: GridSurfaceIdentity;
  readonly updateCellRange: (range: GridCellRange | null) => void;
}): void {
  switch (decision.kind) {
    case "ignore":
    case "copy":
    case "paste":
      return;
    case "reject":
      setKeyboardAnnouncement(decision.announcement);
      return;
    case "exit_grid":
      focusAdjacentOutsideGrid(gridElement, decision.backwards);
      return;
    case "begin_edit":
      executeBeginEditDecision(decision, args, pendingEditorSeedRef, surface);
      return;
    case "navigate":
      executeNavigateDecision(
        decision,
        args,
        pendingRangeEndRef,
        positionMap,
        surface,
        updateCellRange,
      );
      return;
    case "fill":
      executeFillDecision(
        decision,
        columns,
        dataRows,
        dispatchFillIntent,
        positionMap,
        setKeyboardAnnouncement,
        surface,
      );
  }
}

type GridSelectCellArgs = {
  readonly column: { readonly idx: number };
  readonly rowIdx: number;
  readonly selectCell: (
    position: { readonly idx: number; readonly rowIdx: number },
    options?: {
      readonly enableEditor?: boolean;
      readonly shouldFocusCell?: boolean;
    },
  ) => void;
};

function executeBeginEditDecision(
  decision: Extract<SemanticGridDecision, { readonly kind: "begin_edit" }>,
  args: GridSelectCellArgs,
  pendingEditorSeedRef: MutableRefObject<PendingEditorSeed | null>,
  surface: GridSurfaceIdentity,
): void {
  pendingEditorSeedRef.current = decision.seed;
  if (decision.timelineMeasurement && surface.kind === "view_schema") {
    performance.mark("cartulary.workbook.focus_edit_accepted", {
      detail: {
        field: decision.seed.anchor.fieldKey,
        surface: surface.viewSchemaId,
      },
    });
  }
  args.selectCell(
    { idx: args.column.idx, rowIdx: args.rowIdx },
    { enableEditor: true, shouldFocusCell: true },
  );
}

function executeNavigateDecision<Row>(
  decision: Extract<SemanticGridDecision, { readonly kind: "navigate" }>,
  args: GridSelectCellArgs,
  pendingRangeEndRef: MutableRefObject<GridCellAnchor | null>,
  positionMap: GridRdgPresentationModel<Row>,
  surface: GridSurfaceIdentity,
  updateCellRange: (range: GridCellRange | null) => void,
): void {
  const nextPosition = positionMap.positions.get(
    gridAnchorKey(decision.target),
  );
  if (nextPosition === undefined) return;
  if (decision.range !== null) {
    pendingRangeEndRef.current = decision.target;
    updateCellRange(decision.range);
  }
  if (decision.timelineMeasurement && surface.kind === "view_schema") {
    performance.mark("cartulary.workbook.selection_change_accepted", {
      detail: {
        field: decision.target.fieldKey,
        surface: surface.viewSchemaId,
      },
    });
  }
  args.selectCell(nextPosition, { shouldFocusCell: true });
}

function executeFillDecision<Row>(
  decision: Extract<SemanticGridDecision, { readonly kind: "fill" }>,
  columns: readonly SemanticDataGridProps<Row>["columns"][number][],
  dataRows: readonly GridDataRow<Row>[],
  dispatchFillIntent: (intent: GridFillIntent) => void,
  positionMap: GridRdgPresentationModel<Row>,
  setKeyboardAnnouncement: (value: string) => void,
  surface: GridSurfaceIdentity,
): void {
  const intent = planSemanticFillFromRange({
    columns,
    dataRows,
    model: positionMap,
    range: decision.range,
    surface,
  });
  if (intent === null) {
    setKeyboardAnnouncement(
      "Select a writable one-column range before using fill down.",
    );
    return;
  }
  dispatchFillIntent(intent);
  setKeyboardAnnouncement(
    `Filled ${intent.targets.length} cells from the top selected cell.`,
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
  readonly positionMap: GridSemanticPresentationModel<Row>;
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
  const presentation = resolveGridDataStatePresentation(
    dataState,
    delayedInitialLoading
      ? cartularyDesignPresentation.initialLoading.message
      : undefined,
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
  readonly positionMap: GridRdgPositionMap;
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
  const invocationActiveElement = document.activeElement;
  const focusRegisteredCell = () => {
    const cell = cellElements.get(gridAnchorKey(anchor))?.cell;
    if (cell === undefined) return null;
    if (!cell.hasAttribute("tabindex")) cell.tabIndex = -1;
    cell.focus({ preventScroll: true });
    return cell;
  };
  window.setTimeout(() => {
    const activeElement = document.activeElement;
    const registeredCell = cellElements.get(gridAnchorKey(anchor))?.cell;
    if (
      activeElement !== invocationActiveElement &&
      activeElement !== document.body &&
      activeElement !== registeredCell
    ) {
      return;
    }
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

function isFocusableSemanticElement(
  element: GridEditorFocusTarget | undefined,
): element is GridEditorFocusTarget {
  return (
    element?.isConnected === true &&
    !element.disabled &&
    isMeasurableSemanticElement(element)
  );
}

function createDraftFocusTargetRef(
  registry: Map<string, GridEditorFocusTarget>,
  fieldKey: string,
) {
  let registeredElement: GridEditorFocusTarget | null = null;
  return (element: GridEditorFocusTarget | null) => {
    if (element === null) {
      if (registry.get(fieldKey) === registeredElement) {
        registry.delete(fieldKey);
      }
      registeredElement = null;
      return;
    }
    registeredElement = element;
    registry.set(fieldKey, element);
  };
}

function focusRegisteredDraftCell(
  registry: ReadonlyMap<string, GridEditorFocusTarget>,
  fieldKey: string,
): boolean {
  const element = registry.get(fieldKey);
  if (!isFocusableSemanticElement(element)) return false;
  element.focus({ preventScroll: true });
  return document.activeElement === element;
}

function registeredSemanticAnchorRect<Row>(
  registry: ReadonlyMap<string, SemanticCellRegistration>,
  presentation: GridSemanticPresentationModel<Row>,
  anchor: GridCellAnchor,
): DOMRectReadOnly | null {
  if (!gridSurfaceIdentitiesEqual(anchor.surface, presentation.surface)) {
    return null;
  }
  const cell = registry.get(gridAnchorKey(anchor))?.cell;
  return isMeasurableSemanticElement(cell)
    ? cell.getBoundingClientRect()
    : null;
}

function isMeasurableSemanticElement(
  element: HTMLElement | undefined,
): element is HTMLElement {
  if (element === undefined || !element.isConnected || element.hidden) {
    return false;
  }
  if (element.closest("[hidden], [aria-hidden='true']") !== null) return false;
  const style = window.getComputedStyle(element);
  return style.display !== "none" && style.visibility !== "hidden";
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
      buildRdgPresentationModel({
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
