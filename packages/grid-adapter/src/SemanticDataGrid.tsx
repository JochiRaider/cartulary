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
  KeyboardEvent as ReactKeyboardEvent,
  MouseEvent as ReactMouseEvent,
  RefAttributes,
} from "react";
import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  type Column,
  DataGrid,
  type DataGridHandle,
  type DataGridProps,
  type RenderRowProps,
  Row,
  type SortColumn,
  TreeDataGrid,
} from "react-data-grid";
import {
  type ClipboardRepresentations,
  clipboardRepresentations,
} from "./clipboardCodec";
import { normalizeMeasuredColumnWidth } from "./columnSizing";
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
  visibleGridPageSize,
} from "./domInteraction";
import {
  type ActiveEditorSession,
  editorSeedForTarget,
  type PendingEditorSeed,
} from "./editorSessionPolicy";
import { GridOperationalStatePlane } from "./GridOperationalStatePlane";
import type { RegisteredGridCell } from "./gridInteractionDom";
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
import { createSemanticCellNavigation } from "./semanticCellNavigation";
import { captureSemanticClear, isClearNavigationKey } from "./semanticClear";
import {
  mergeSemanticFillIntents,
  planSemanticCopy,
  planSemanticFill,
  planSemanticFillFromRange,
  planSemanticPaste,
} from "./semanticClipboardPolicy";
import {
  gridDataStateBlocksInteraction,
  gridDataStatePresentsAuthorizedRows,
  gridDataStatePresentsDraft,
} from "./semanticDataState";
import { createSemanticFocusRequests } from "./semanticFocusRequest";
import {
  decideSemanticGridKey,
  decideSpreadsheetNavigation,
  normalizeGridKey,
  type SemanticGridDecision,
} from "./semanticKeyboardPolicy";
import {
  buildSemanticGroupBuckets,
  coreRecordId,
  coreRowVersion,
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
  semanticPresentationContainsAnchor,
  semanticTarget,
} from "./semanticPresentation";
import { createGridPresentationPort } from "./semanticPresentationPort";
import {
  extendSemanticCellRange,
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
import { useFrozenDataColumns } from "./useFrozenDataColumns";
import { useGridColumnSizing } from "./useGridColumnSizing";
import { useGridInteraction } from "./useGridInteraction";
import { elementCssScale, revealGridCell } from "./viewportGeometry";

const emptySelectedRecordIds: ReadonlySet<string> = new Set();

type SemanticCellRegistration = RegisteredGridCell;

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
  useLayoutEffect(() => {
    if (!bulkSelection || !state) return;
    const visible = new Set(state.selectableIds);
    const retained = new Set(
      [...bulkSelection.selectedRecordIds].filter((id) => visible.has(id)),
    );
    if (retained.size !== bulkSelection.selectedRecordIds.size)
      bulkSelection.onSelectedRecordIdsChange(retained);
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
  presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>,
  vendorHandle: MutableRefObject<DataGridHandle | null>,
  draftFieldKeysRef: MutableRefObject<readonly string[]>,
  editable: boolean,
) {
  const cellElementsRef = useRef(new Map<string, SemanticCellRegistration>());
  const draftFocusTargetsRef = useRef(new Map<string, GridEditorFocusTarget>());
  const focusRequests = useMemo(
    () =>
      createSemanticFocusRequests({
        prepare: (target) => {
          if (target.kind === "draft") {
            const idx = presentationRef.current.columnKeys.indexOf(
              target.fieldKey,
            );
            if (idx >= 0 && draftFieldKeysRef.current.includes(target.fieldKey))
              vendorHandle.current?.scrollToCell({ idx });
          } else if (target.kind === "cell") {
            const position = presentationRef.current.positions.get(
              gridAnchorKey(target.anchor),
            );
            if (position) {
              const cell = cellElementsRef.current.get(
                gridAnchorKey(target.anchor),
              )?.cell;
              // A rendered target can scroll synchronously. The vendor's
              // virtual scroll sentinel persists until intersection, and would
              // otherwise override a caller's restored viewport on later renders.
              if (cell?.isConnected && vendorHandle.current?.element)
                revealGridCell(vendorHandle.current.element, cell);
              else vendorHandle.current?.scrollToCell(position);
              vendorHandle.current?.selectCell(position);
            }
          }
        },
        resolve: (target) => {
          if (target.kind === "draft") {
            if (
              !draftFieldKeysRef.current.includes(target.fieldKey) ||
              !presentationRef.current.columnKeys.includes(target.fieldKey)
            )
              return { kind: "unavailable" };
            const element = draftFocusTargetsRef.current.get(target.fieldKey);
            if (element instanceof HTMLElement && vendorHandle.current?.element)
              revealGridCell(vendorHandle.current.element, element);
            return element ? { kind: "target", element } : { kind: "pending" };
          }
          if (target.kind === "cell") {
            if (
              !gridSurfaceIdentitiesEqual(
                target.anchor.surface,
                presentationRef.current.surface,
              ) ||
              !presentationRef.current.positions.has(
                gridAnchorKey(target.anchor),
              )
            )
              return { kind: "unavailable" };
            const element = cellElementsRef.current.get(
              gridAnchorKey(target.anchor),
            )?.cell;
            if (element instanceof HTMLElement && vendorHandle.current?.element)
              revealGridCell(vendorHandle.current.element, element);
            return element ? { kind: "target", element } : { kind: "pending" };
          }
          const element = vendorHandle.current?.element;
          return element ? { kind: "target", element } : { kind: "pending" };
        },
      }),
    [presentationRef, vendorHandle, draftFieldKeysRef],
  );
  useLayoutEffect(() => {
    focusRequests.refresh();
  });
  useLayoutEffect(() => {
    if (!editable) focusRequests.cancel();
    return () => focusRequests.cancel();
  }, [focusRequests, editable]);
  const registerSemanticCell = useCallback(
    (anchor: GridCellAnchor, cell: HTMLElement | null, token: object) => {
      const key = gridAnchorKey(anchor);
      if (cell === null) {
        if (cellElementsRef.current.get(key)?.token === token) {
          cellElementsRef.current.delete(key);
        }
        return;
      }
      cellElementsRef.current.set(key, { anchor, cell, token });
      focusRequests.refresh();
    },
    [focusRequests],
  );
  const draftFocusTargetRef = useCallback(
    (fieldKey: string) =>
      createDraftFocusTargetRef(
        draftFocusTargetsRef.current,
        fieldKey,
        focusRequests.refresh,
      ),
    [focusRequests],
  );
  const requestFocus = focusRequests.requestFocus;
  const focusDraftCell = useCallback(
    (fieldKey: string) => {
      void requestFocus({ kind: "draft", fieldKey });
    },
    [requestFocus],
  );
  return {
    cellElementsRef,
    requestFocus,
    focusDraftCell,
    draftFocusTargetRef,
    registerSemanticCell,
  };
}

function useGridEditorController() {
  const pendingEditorSeedRef = useRef<PendingEditorSeed | null>(null);
  const activeEditorSessionRef = useRef<ActiveEditorSession | null>(null);
  const readEditorSeed = useCallback(
    (
      target: Parameters<typeof editorSeedForTarget>[1],
      retainAcrossVersions = false,
    ) => {
      const pending = pendingEditorSeedRef.current;
      return editorSeedForTarget(
        pending !== null && retainAcrossVersions
          ? { ...pending, retained: true }
          : pending,
        target,
      );
    },
    [],
  );
  const retainEditorDraft = useCallback(
    (
      target: Parameters<typeof editorSeedForTarget>[1],
      value: unknown,
      selectionRange: GridEditorActivation["selectionRange"],
    ) => {
      const pending = pendingEditorSeedRef.current;
      if (pending === null || !sameGridCellAnchor(pending.anchor, target))
        return;
      pendingEditorSeedRef.current = {
        ...pending,
        baseRowVersion: target.mutationIdentity.baseRowVersion,
        hasValue: true,
        retained: true,
        value,
        activation: {
          ...pending.activation,
          initialSelection: "end",
          selectionRange,
        },
      };
    },
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
  return {
    activeEditorSessionRef,
    clearEditorSeed,
    pendingEditorSeedRef,
    readEditorSeed,
    retainEditorDraft,
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
  announce,
}: Pick<
  SemanticDataGridProps<Row>,
  "clipboardPaste" | "columns" | "surface"
> & {
  readonly editable: boolean;
  readonly presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>;
  readonly announce: (message: string) => void;
  readonly updateRange: (range: GridCellRange | null) => void;
}) {
  const delivered = useRef(new WeakSet<object>());
  return useCallback(
    (
      row: GridDataRow<Row>,
      fieldKey: string,
      clipboardText: ClipboardRepresentations,
      delivery?: object,
    ): boolean => {
      if (!editable) {
        announce("This workbook is read-only.");
        return true;
      }
      const event =
        delivery &&
        "nativeEvent" in delivery &&
        typeof delivery.nativeEvent === "object" &&
        delivery.nativeEvent !== null
          ? delivery.nativeEvent
          : delivery;
      if (event && delivered.current.has(event)) return true;
      const target = semanticTarget(row, fieldKey, columns, surface);
      if (target === null) return false;
      if (clipboardPaste === undefined) return true;
      if (event) delivered.current.add(event);
      const decoded = clipboardPaste.decode(clipboardText);
      if (decoded.kind === "noop") return true;
      if (decoded.kind === "failure") {
        clipboardPaste.onError?.(decoded.message);
        announce(decoded.message);
        return true;
      }
      const intent = planSemanticPaste({
        input: decoded,
        model: presentationRef.current,
        target,
      });
      if (intent === null) {
        const message =
          "The clipboard does not fit the available writable cells.";
        clipboardPaste.onError?.(message);
        announce(message);
        return true;
      }
      const admitted = clipboardPaste.onPaste(intent);
      if (admitted === true || admitted === undefined)
        updateRange(intent.range);
      else if (admitted instanceof Promise)
        void admitted.catch(() => announce("Paste could not be applied."));
      return true;
    },
    [
      clipboardPaste,
      columns,
      editable,
      presentationRef,
      surface,
      updateRange,
      announce,
    ],
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

function useSemanticDataGrid<Row>(
  props: SemanticDataGridProps<Row>,
  ref: ForwardedRef<GridHandle>,
  enableVirtualization: boolean,
) {
  const deliveredClearEvents = useRef(new WeakSet<object>());
  const {
    accessibleLabel,
    activeRowIdentity = null,
    allowPasteCreateRows = false,
    actionsColumn,
    clipboardPaste,
    cellRangeSelection,
    coreRecordBulkSelection,
    cellRange: controlledCellRange,
    columns,
    dataState = { kind: "ready" },
    density = "default",
    draftRow: ownerDraftRow,
    fillViewportInline = false,
    getCellState,
    getRowState,
    grouping = null,
    keyboardNavigation = "region",
    onActiveCellChange,
    onCellRangeChange,
    onColumnReorder,
    onColumnSizingIntent,
    onCopyCell,
    onFillCells,
    onSelectRow,
    onSortChange,
    dataRows: ownerDataRows,
    rowGutter,
    sort = [],
    surface,
  } = props;
  const dataRows = gridDataStatePresentsAuthorizedRows(dataState)
    ? ownerDataRows
    : [];
  const draftRow = gridDataStatePresentsDraft(dataState)
    ? ownerDraftRow
    : undefined;
  const capabilities = resolveSemanticGridCapabilities(props);
  const effectiveInteractionMode = capabilities.interactionMode;
  const editable = capabilities.editable;
  const vendorHandle = useRef<DataGridHandle>(null);
  const freezing = useFrozenDataColumns({
    getRoot: () => vendorHandle.current?.element ?? null,
    columns,
    prefix: props.frozenDataColumnPrefix,
    structuralCount:
      (coreRecordBulkSelection && editable ? 1 : 0) +
      (grouping || rowGutter ? 1 : 0),
  });
  const sizing = useGridColumnSizing({
    getRoot: () => vendorHandle.current?.element ?? null,
    columns,
    rows: dataRows,
    density,
    dataState: dataState.kind,
    interactionMode: effectiveInteractionMode.kind,
    surfaceKey: gridSurfaceIdentityKey(surface),
    groupingFieldKey: grouping?.fieldKey ?? null,
  });
  const onHeaderSizingKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (
      !onColumnSizingIntent ||
      !(event.ctrlKey || event.metaKey) ||
      event.altKey ||
      !["ArrowLeft", "ArrowRight"].includes(event.key)
    )
      return;
    const header =
      event.target instanceof HTMLElement
        ? event.target.closest<HTMLElement>('[role="columnheader"]')
        : null;
    const column = columns.find(
      (entry) => entry.fieldKey === header?.dataset.gridFieldKey,
    );
    if (!header || !column) return;
    event.preventDefault();
    event.stopPropagation();
    sizing.invalidate();
    const direction = getComputedStyle(header).direction === "rtl" ? -1 : 1;
    const widthPx = normalizeMeasuredColumnWidth(
      header.getBoundingClientRect().width / elementCssScale(header) +
        direction *
          (event.key === "ArrowRight" ? 1 : -1) *
          cartularyDesignPresentation.workbookColumnSizing.keyboardStepPx,
      column,
    );
    if (widthPx !== null)
      onColumnSizingIntent({
        kind: "set_width",
        fieldKey: column.fieldKey,
        widthPx,
      });
  };
  const dataRowsRef = useRef(dataRows);
  dataRowsRef.current = dataRows;
  const semanticPresentationRef = useRef<GridRdgPresentationModel<Row>>(
    emptyRdgPresentationModel(surface),
  );
  const presentationPort = useMemo(() => createGridPresentationPort(), []);
  const publishPresentation = useCallback(
    (model: GridSemanticPresentationModel<Row>) => {
      presentationPort.publish(model);
    },
    [presentationPort],
  );
  const [keyboardAnnouncement, setKeyboardAnnouncement] = useState("");
  const [contextDescription, setAccessibleDescription] = useState<
    string | undefined
  >();
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
  const draftFieldKeysRef = useRef<readonly string[]>([]);
  draftFieldKeysRef.current =
    editable && draftRow !== undefined
      ? columns
          .filter(
            (column) =>
              column.renderDraftCell !== undefined &&
              column.draftWritable === true,
          )
          .map((column) => column.fieldKey)
      : [];
  const {
    cellElementsRef: semanticCellElementsRef,
    requestFocus: requestRegisteredFocus,
    draftFocusTargetRef,
    focusDraftCell,
    registerSemanticCell,
  } = useGridRegistration(
    semanticPresentationRef,
    vendorHandle,
    draftFieldKeysRef,
    editable,
  );
  const cancelInteractionRef = useRef<() => void>(() => {});
  const requestFocus = useCallback<GridHandle["requestFocus"]>(
    (target, options) => {
      if (options?.signal?.aborted) return Promise.resolve("cancelled");
      cancelInteractionRef.current();
      if (
        target.kind === "cell" &&
        semanticPresentationContainsAnchor(
          semanticPresentationRef.current,
          target.anchor,
        )
      ) {
        if (
          !options?.preserveSelection &&
          !(cellRangeSelection?.keyboardEntry === "cycle"
            ? gridCellRangeContains(
                semanticPresentationRef.current,
                cellRangeRef.current,
                target.anchor,
              )
            : sameGridCellAnchor(
                cellRangeRef.current?.end ?? null,
                target.anchor,
              ))
        )
          updateCellRange({ start: target.anchor, end: target.anchor });
      } else if (target.kind === "draft") updateCellRange(null);
      return requestRegisteredFocus(target, options);
    },
    [
      cellRangeRef,
      cellRangeSelection?.keyboardEntry,
      requestRegisteredFocus,
      updateCellRange,
    ],
  );
  const isCellRangeSelected = useCallback(
    (
      row: GridDataRow<Row>,
      column: SemanticDataGridProps<Row>["columns"][number],
    ) =>
      gridCellRangeContains(semanticPresentationRef.current, cellRange, {
        surface,
        rowIdentity: row.rowIdentity,
        fieldKey: column.fieldKey,
      }),
    [cellRange, surface],
  );
  const {
    activeEditorSessionRef,
    clearEditorSeed,
    pendingEditorSeedRef,
    readEditorSeed,
    retainEditorDraft,
    registerEditorSession,
  } = useGridEditorController();
  const interactionScopeKey = JSON.stringify([
    gridSurfaceIdentityKey(surface),
    cellRangeSelection?.scopeKey,
    grouping?.fieldKey,
  ]);
  const interaction = useGridInteraction({
    read: () => ({
      active: activeEditorSessionRef.current?.target ?? activeCellAnchor,
      editor: activeEditorSessionRef.current,
      model: semanticPresentationRef.current,
      range: cellRangeRef.current,
      enabled: cellRangeSelection?.kind === "contiguous",
      available:
        !gridDataStateBlocksInteraction(dataState) && dataRows.length > 0,
      scopeKey: interactionScopeKey,
      authorityKey: effectiveInteractionMode.kind,
    }),
    root: () => vendorHandle.current?.element ?? null,
    cells: () => semanticCellElementsRef.current,
    changeRange: updateCellRange,
    announce: setKeyboardAnnouncement,
    accept: (range, edit, inspect) => {
      const anchor = range.end;
      const position = semanticPresentationRef.current.positions.get(
        gridAnchorKey(anchor),
      );
      const row = dataRowsRef.current.find((candidate) =>
        gridRowIdentitiesEqual(candidate.rowIdentity, anchor.rowIdentity),
      );
      if (position === undefined || row === undefined) return;
      const editing =
        edit &&
        prepareEditorActivation(row, anchor.fieldKey, {
          initialSelection: "end",
          source: "pointer",
        });
      updateCellRange(editing ? null : range);
      publishActiveCell(anchor);
      if (inspect) onSelectRow?.(anchor.rowIdentity);
      vendorHandle.current?.selectCell(position, {
        enableEditor: editing,
        shouldFocusCell: true,
      });
      // RDG may elide selecting its already-active position. The semantic
      // focus owner still owes this admitted click/range a focus transition.
      if (!editing) void requestFocus({ kind: "cell", anchor });
    },
  });
  cancelInteractionRef.current = interaction.controller.cancel;
  const navigationState = useRef({
    available: !gridDataStateBlocksInteraction(dataState),
    authority: effectiveInteractionMode.kind,
    scope: interactionScopeKey,
  });
  navigationState.current = {
    available: !gridDataStateBlocksInteraction(dataState),
    authority: effectiveInteractionMode.kind,
    scope: interactionScopeKey,
  };
  const cellNavigation = useMemo(
    () =>
      createSemanticCellNavigation({
        read: () => ({
          ...navigationState.current,
          presentation: presentationPort.getSnapshot(),
          editor: activeEditorSessionRef.current,
          range: cellRangeRef.current,
        }),
        focus: requestFocus,
        changeRange: updateCellRange,
        exit: (backwards) =>
          focusAdjacentOutsideGrid(
            vendorHandle.current?.element ?? null,
            backwards,
          ),
        cancelInteraction: () => cancelInteractionRef.current(),
      }),
    [
      activeEditorSessionRef,
      cellRangeRef,
      presentationPort,
      requestFocus,
      updateCellRange,
    ],
  );
  const prepareEditorKeyboardAction = useCallback(
    (
      target: GridCellAnchor,
      action:
        | { readonly kind: "exit"; readonly backwards: boolean }
        | { readonly kind: "move"; readonly rowDelta: -1 | 1 },
    ) => {
      const editor = activeEditorSessionRef.current;
      if (editor !== null && !sameGridCellAnchor(editor.target, target))
        return null;
      const model = semanticPresentationRef.current;
      const decision =
        keyboardNavigation === "spreadsheet"
          ? decideSpreadsheetNavigation(
              model,
              target,
              {
                key: action.kind === "exit" ? "Tab" : "Enter",
                shiftKey:
                  action.kind === "exit"
                    ? action.backwards
                    : action.rowDelta < 0,
              },
              draftFieldKeysRef.current,
              cellRangeSelection?.keyboardEntry === "cycle"
                ? cellRangeRef.current
                : null,
            )
          : action.kind === "exit"
            ? { kind: "exit_grid" as const, backwards: action.backwards }
            : {
                kind: "navigate" as const,
                range: null,
                target:
                  navigateSemanticPresentation(model, target, {
                    key: action.rowDelta < 0 ? "ArrowUp" : "ArrowDown",
                  }) ?? target,
              };
      const capturedPresentation = presentationPort.getSnapshot();
      const capturedState = navigationState.current;
      return () => {
        if (
          !navigationState.current.available ||
          navigationState.current.authority !== capturedState.authority ||
          navigationState.current.scope !== capturedState.scope ||
          presentationPort.getSnapshot() !== capturedPresentation
        )
          return;
        if (decision.kind === "focus_draft") {
          void cellNavigation.depart(
            { kind: "draft", fieldKey: decision.fieldKey },
            {
              isCurrent: () =>
                draftFieldKeysRef.current.includes(decision.fieldKey),
            },
          );
        } else if (decision.kind === "exit_grid") {
          void cellNavigation.depart({
            kind: "exit",
            backwards: decision.backwards,
          });
        } else {
          void cellNavigation.depart(
            {
              kind: "cell",
              anchor: decision.kind === "navigate" ? decision.target : target,
            },
            {
              range:
                decision.kind === "navigate"
                  ? (decision.range ?? undefined)
                  : undefined,
            },
          );
        }
      };
    },
    [
      presentationPort,
      activeEditorSessionRef,
      cellNavigation,
      cellRangeRef,
      cellRangeSelection?.keyboardEntry,
      keyboardNavigation,
    ],
  );
  const handleEditorKeyboardAction = useCallback(
    (
      target: GridCellAnchor,
      action:
        | { readonly kind: "exit"; readonly backwards: boolean }
        | { readonly kind: "move"; readonly rowDelta: -1 | 1 },
    ) => {
      prepareEditorKeyboardAction(target, action)?.();
    },
    [prepareEditorKeyboardAction],
  );
  useLayoutEffect(() => {
    cellNavigation.reconcile();
  });
  useLayoutEffect(() => {
    const externalFocus = (event: FocusEvent) => {
      if (
        event.target instanceof Node &&
        !vendorHandle.current?.element?.contains(event.target)
      )
        cellNavigation.cancel();
    };
    document.addEventListener("pointerdown", cellNavigation.cancel, true);
    document.addEventListener("keydown", cellNavigation.cancel, true);
    document.addEventListener("input", cellNavigation.cancel, true);
    document.addEventListener("compositionstart", cellNavigation.cancel, true);
    document.addEventListener("focusin", externalFocus, true);
    return () => {
      document.removeEventListener("pointerdown", cellNavigation.cancel, true);
      document.removeEventListener("keydown", cellNavigation.cancel, true);
      document.removeEventListener("input", cellNavigation.cancel, true);
      document.removeEventListener(
        "compositionstart",
        cellNavigation.cancel,
        true,
      );
      document.removeEventListener("focusin", externalFocus, true);
      cellNavigation.cancel();
    };
  }, [cellNavigation]);
  useLayoutEffect(() => () => presentationPort.retire(), [presentationPort]);
  const isCellRangePreview = useCallback(
    (
      row: GridDataRow<Row>,
      column: SemanticDataGridProps<Row>["columns"][number],
    ) => {
      const anchor = {
        surface,
        rowIdentity: row.rowIdentity,
        fieldKey: column.fieldKey,
      };
      return (
        interaction.preview !== null &&
        (sameGridCellAnchor(interaction.preview.start, anchor) ||
          gridCellRangeContains(
            semanticPresentationRef.current,
            interaction.preview,
            anchor,
          ))
      );
    },
    [interaction.preview, surface],
  );
  const onDoubleClickCapture = (event: ReactMouseEvent<HTMLDivElement>) => {
    if (!isGridFillHandleTarget(event.target)) return;
    event.preventDefault();
    event.stopPropagation();
  };
  useLayoutEffect(() => {
    if (!gridDataStatePresentsAuthorizedRows(dataState)) clearEditorSeed();
  }, [clearEditorSeed, dataState]);
  assertGridRows(ownerDataRows);

  const handleSemanticPaste = useGridPasteController({
    announce: setKeyboardAnnouncement,
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
        frozenDataColumnPrefix: freezing.activePrefix,
        onColumnSizingIntent,
        onColumnSizingStart: sizing.invalidate,
        actionsColumn,
        bulkSelection: compiledBulkSelection,
        clearEditorSeed,
        cellStateFor,
        columns,
        readEditorSeed,
        retainEditorDraft,
        editable,
        isCellRangeSelected,
        isCellRangePreview,
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
      freezing.activePrefix,
      onColumnSizingIntent,
      sizing.invalidate,
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
      isCellRangePreview,
      registerEditorSession,
      registerSemanticCell,
      readEditorSeed,
      retainEditorDraft,
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
  useLayoutEffect(() => {
    publishPresentation(semanticPresentationRef.current);
  });
  const previousPresentationRef = useRef(semanticPresentationRef.current);
  const detachEditorPresentation = activeEditorSessionRef.current?.detach;
  useLayoutEffect(() => {
    const seed = pendingEditorSeedRef.current;
    if (seed === null) return;
    const key = gridAnchorKey(seed.anchor);
    const position = semanticPresentationRef.current.positions.get(key);
    if (editable && position !== undefined) return;
    // Eligibility loss detaches presentation. Escape is the separate source
    // discard action; returning to this target must require explicit activation.
    activeEditorSessionRef.current = null;
    clearEditorSeed();
    detachEditorPresentation?.();
  });
  const retainedAnchor =
    pendingEditorSeedRef.current?.anchor ?? activeCellAnchor;
  const retainedCell =
    retainedAnchor === null
      ? undefined
      : semanticCellElementsRef.current.get(gridAnchorKey(retainedAnchor))
          ?.cell;
  // Capture focus before RDG commits a reordered projection. Its selected
  // position can become another record, or disappear when the page shrinks.
  const retainedCellHadFocus =
    retainedCell?.contains(document.activeElement) === true;
  const focusedControl = document.activeElement;
  const retainedSelection =
    focusedControl instanceof HTMLInputElement ||
    focusedControl instanceof HTMLTextAreaElement
      ? {
          start: focusedControl.selectionStart ?? 0,
          end: focusedControl.selectionEnd ?? 0,
        }
      : undefined;
  useLayoutEffect(() => {
    const previous = previousPresentationRef.current;
    const current = semanticPresentationRef.current;
    previousPresentationRef.current = current;
    if (!retainedCellHadFocus || retainedAnchor === null) return;
    const key = gridAnchorKey(retainedAnchor);
    const before = previous.positions.get(key);
    const after = current.positions.get(key);
    if (before !== undefined && after === undefined) {
      // Extension surfaces retain their own page-replacement focus policy.
      if (retainedAnchor.rowIdentity.kind !== "core_record") return;
      const fallbackRow = current.rowIdentities[0];
      const fieldKey = current.fieldKeys.includes(retainedAnchor.fieldKey)
        ? retainedAnchor.fieldKey
        : current.fieldKeys[0];
      const anchor =
        fallbackRow && fieldKey
          ? { ...retainedAnchor, rowIdentity: fallbackRow, fieldKey }
          : null;
      publishActiveCell(anchor);
      const priorFocus = document.activeElement;
      queueMicrotask(() => {
        if (
          document.activeElement !== priorFocus &&
          document.activeElement !== document.body
        )
          return;
        void requestFocus(anchor ? { kind: "cell", anchor } : { kind: "root" });
      });
      return;
    }
    if (
      before === undefined ||
      after === undefined ||
      (before.rowIdx === after.rowIdx && before.idx === after.idx)
    )
      return;
    const grid = vendorHandle.current;
    if (grid === null) return;
    const element = grid.element;
    if (element !== null) {
      element.scrollTop +=
        (after.rowIdx - before.rowIdx) * workbookGridRowHeightPx(density);
    }
    const seed = pendingEditorSeedRef.current;
    if (seed !== null && sameGridCellAnchor(seed.anchor, retainedAnchor)) {
      pendingEditorSeedRef.current = {
        ...seed,
        activation: { ...seed.activation, selectionRange: retainedSelection },
      };
    }
    const restoreSeed = pendingEditorSeedRef.current;
    const activeElement = document.activeElement;
    // RDG resets an out-of-bounds editor during its own render. Selecting in
    // this layout effect would first commit that obsolete vendor position.
    queueMicrotask(() => {
      if (
        pendingEditorSeedRef.current !== restoreSeed ||
        (document.activeElement !== activeElement &&
          document.activeElement !== document.body)
      )
        return;
      vendorHandle.current?.selectCell(after, {
        enableEditor:
          editable &&
          seed !== null &&
          sameGridCellAnchor(seed.anchor, retainedAnchor),
        shouldFocusCell: true,
      });
    });
  });
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
      if (
        activation.source === "pointer" &&
        fieldKey === "timeline.activity_synopsis_text"
      )
        performance.mark("cartulary.workbook.focus_edit_accepted", {
          detail: { field: fieldKey, surface: "grid" },
        });
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
  useImperativeHandle(
    ref,
    () => ({
      setAccessibleDescription,
      captureClearIntent: (delivery) =>
        interaction.controller.pointerId !== null
          ? null
          : captureSemanticClear(
              semanticPresentationRef.current,
              activeCellAnchor,
              cellRangeRef.current,
              delivery,
            ),
      columnSizing: sizing.port,
      frozenColumns: freezing.port,
      presentation: presentationPort,
      navigateToCell: cellNavigation.navigate,
      activateEdit: (anchor, seed) => {
        cellNavigation.cancel();
        interaction.controller.cancel();
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
            {
              initialSelection: seed === undefined ? "all" : "end",
              source: "programmatic",
              selectionRange: seed?.selectionRange,
            },
            seed,
          )
        ) {
          return false;
        }
        vendorHandle.current?.scrollToCell(position);
        updateCellRange(null);
        vendorHandle.current?.selectCell(position, { enableEditor: true });
        return true;
      },
      cancelEdit: (anchor) => {
        cellNavigation.cancel();
        const session = activeEditorSessionRef.current;
        if (
          session === null ||
          gridAnchorKey(session.target) !== gridAnchorKey(anchor)
        ) {
          return false;
        }
        session.cancel(
          vendorHandle.current?.element?.contains(document.activeElement) ===
            true,
        );
        return true;
      },
      detachEdit: () => {
        cellNavigation.cancel();
        interaction.controller.cancel();
        const session = activeEditorSessionRef.current;
        activeEditorSessionRef.current = null;
        clearEditorSeed();
        session?.detach();
      },
      ownsNavigationFocus: (target) =>
        target instanceof HTMLElement &&
        target.getAttribute("role") === "gridcell" &&
        vendorHandle.current?.element?.contains(target) === true,
      getActiveCell: () =>
        pendingEditorSeedRef.current?.anchor ?? activeCellAnchor,
      requestFocus: (target, options) => {
        cellNavigation.cancel();
        return requestFocus(target, options);
      },
      focusAdjacentRegion: (backwards) =>
        focusAdjacentOutsideGrid(
          vendorHandle.current?.element ?? null,
          backwards,
        ),
      getAnchorRect: (anchor) =>
        registeredSemanticAnchorRect(
          semanticCellElementsRef.current,
          semanticPresentationRef.current,
          anchor,
        ),
      getScrollElement: () => vendorHandle.current?.element ?? null,
      isAnchorRendered: (anchor) =>
        semanticCellElementsRef.current.has(gridAnchorKey(anchor)),
      prepareNavigation: (current, intent) => {
        if (intent.key !== "Enter" && intent.key !== "Tab") return null;
        return prepareEditorKeyboardAction(
          current,
          intent.key === "Tab"
            ? { kind: "exit", backwards: intent.shiftKey === true }
            : { kind: "move", rowDelta: intent.shiftKey ? -1 : 1 },
        );
      },
      moveFocus: (current, intent) => {
        if (
          keyboardNavigation === "spreadsheet" &&
          (intent.key === "Enter" || intent.key === "Tab")
        ) {
          const session = activeEditorSessionRef.current;
          if (session !== null) {
            if (!sameGridCellAnchor(session.target, current)) return null;
            session.cancel(false);
          }
          clearEditorSeed();
          handleEditorKeyboardAction(
            current,
            intent.key === "Tab"
              ? { kind: "exit", backwards: intent.shiftKey === true }
              : { kind: "move", rowDelta: intent.shiftKey ? -1 : 1 },
          );
          return null;
        }
        const next = navigateSemanticPresentation(
          semanticPresentationRef.current,
          current,
          intent,
        );
        if (next === null) return null;
        updateCellRange(
          intent.shiftKey === true && intent.key.startsWith("Arrow")
            ? extendSemanticCellRange(
                semanticPresentationRef.current,
                current,
                cellRangeRef.current,
                next,
              )
            : { start: next, end: next },
        );
        void requestFocus({ kind: "cell", anchor: next });
        return next;
      },
      planPasteTargets: (current, dimensions) =>
        planSemanticPasteTargets(
          semanticPresentationRef.current,
          current,
          dimensions,
        ),
      scrollToAnchor: (anchor) =>
        scrollToSemanticAnchor({
          anchor,
          positionMap: semanticPresentationRef.current,
          vendorHandle: vendorHandle.current,
          cell: semanticCellElementsRef.current.get(gridAnchorKey(anchor))
            ?.cell,
        }),
    }),
    [
      activeEditorSessionRef,
      interaction.controller,
      updateCellRange,
      cellRangeRef,
      sizing.port,
      freezing.port,
      presentationPort,
      cellNavigation,
      activeCellAnchor,
      pendingEditorSeedRef,
      clearEditorSeed,
      requestFocus,
      semanticCellElementsRef,
      keyboardNavigation,
      handleEditorKeyboardAction,
      prepareEditorKeyboardAction,
    ],
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
    "aria-description":
      [
        contextDescription,
        props.onClearCells
          ? "Delete clears selected cell contents. Backspace edits the active cell."
          : undefined,
        cellRangeSelection?.keyboardEntry === "cycle"
          ? "Within a selected range, Tab moves across rows and Enter moves down columns, wrapping at the edge. Shift reverses direction. Escape returns to a single active cell. F2 edits the active cell."
          : undefined,
      ]
        .filter(Boolean)
        .join(" ") || undefined,
    className: `${gridScrollportClassName()} cartulary-grid rdg-dark`,
    columns: compiledColumns,
    // Production grids always use RDG's row and column virtualization. A
    // result-size threshold would create two interaction runtimes and let
    // small fixtures miss focus, range, editor, and scrolling defects.
    enableVirtualization,
    headerRowHeight: workbookGridRowHeightPx(density),
    headerRowClass: "cartulary-grid-header-row",
    onCellMouseDown: (_args, event) => {
      // The interaction owner admits cell selection; embedded actions own themselves.
      event.preventGridDefault();
    },
    onCellClick: ({ column, row }, event) => {
      if (
        (interaction.binding.current?.handledPointer && event.detail !== 0) ||
        isInteractiveCellActionTarget(event.target) ||
        isGridFillHandleTarget(event.target) ||
        event.button !== 0 ||
        event.altKey ||
        event.ctrlKey ||
        event.metaKey
      )
        return;
      const anchor = semanticAnchor(row, column.key, columns, surface);
      if (anchor !== null) interaction.click(anchor, event.shiftKey);
    },
    onCellDoubleClick: (_args, event) => {
      event.preventGridDefault();
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
      if ("kind" in plan.representations) {
        event.preventDefault();
        setKeyboardAnnouncement(plan.representations.message);
        return;
      }
      for (const [type, value] of Object.entries(plan.representations)) {
        event.clipboardData?.setData(type, value);
      }
      event.preventDefault();
      onCopyCell?.(plan.intent);
    },
    onCellKeyDown: (args, event) => {
      if (args.mode !== "SELECT" || !isDataRow<Row>(args.row)) return;
      if (
        event.nativeEvent.isComposing ||
        isInteractiveCellActionTarget(event.target)
      ) {
        event.preventGridDefault();
        return;
      }
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
      if (props.onClearCells && isClearNavigationKey(event)) {
        event.preventGridDefault();
        if (window.getSelection()?.isCollapsed === false) return;
        event.preventDefault();
        event.stopPropagation();
        if (
          event.repeat ||
          interaction.controller.pointerId !== null ||
          deliveredClearEvents.current.has(event.nativeEvent)
        )
          return;
        deliveredClearEvents.current.add(event.nativeEvent);
        const intent = captureSemanticClear(
          semanticPresentationRef.current,
          anchor,
          cellRangeRef.current,
          event.nativeEvent,
        );
        if (intent) props.onClearCells(intent);
        else
          setKeyboardAnnouncement(
            "Clear contents requires an available committed cell selection.",
          );
        return;
      }
      const decision = decideSemanticGridKey({
        anchor,
        column: semanticColumn,
        editable,
        input: normalizeGridKey(event),
        keyboardNavigation,
        draftFieldKeys: draftFieldKeysRef.current,
        model: semanticPresentationRef.current,
        pageSize: visibleGridPageSize(
          vendorHandle.current?.element ?? null,
          workbookGridRowHeightPx(density),
        ),
        range: cellRangeRef.current,
        rangeKeyboardEntry: cellRangeSelection?.keyboardEntry,
        readOnlyLabel:
          effectiveInteractionMode.kind === "read_only"
            ? effectiveInteractionMode.label
            : "This workbook is read-only.",
        row: args.row,
      });
      if (decision.kind === "ignore") return;
      setKeyboardAnnouncement("");
      event.preventDefault();
      event.preventGridDefault();
      if (decision.kind === "collapse_range") event.stopPropagation();
      if (decision.kind === "focus_draft") {
        updateCellRange(null);
        focusDraftCell(decision.fieldKey);
        return;
      }
      executeSemanticKeyDecision({
        args,
        columns,
        dataRows,
        decision,
        dispatchFillIntent,
        gridElement: vendorHandle.current?.element ?? null,
        pendingEditorSeedRef,
        positionMap: semanticPresentationRef.current,
        setKeyboardAnnouncement,
        surface,
        updateCellRange,
      });
    },
    onCellPaste: ({ column, row }, event) => {
      const clipboardText = event.clipboardData
        ? clipboardRepresentations(event.clipboardData)
        : {};
      if (
        !editable ||
        handleSemanticPaste(row, column.key, clipboardText, event)
      ) {
        event.preventDefault();
      }
      return row;
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
      if (interaction.controller.active) return;
      const anchor =
        row === undefined || column === undefined
          ? null
          : semanticAnchor(row, column.key, columns, surface);
      // Vendor notifications publish focus only. Explicit semantic commands own selection.
      publishActiveCell(anchor);
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
            data-grid-row-version={coreRowVersion(rowProps.row) ?? undefined}
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
      accessibleLabel={accessibleLabel}
      bulkSelectionEnabled={coreRecordBulkSelection !== undefined}
      columns={columns}
      dataRows={dataRows}
      dataState={dataState}
      density={density}
      draftVisible={editable && draftRow !== undefined}
      interactionMode={effectiveInteractionMode}
      keyboardAnnouncement={keyboardAnnouncement}
      onDoubleClickCapture={onDoubleClickCapture}
      onKeyDownCapture={onHeaderSizingKeyDown}
      positionMap={semanticPresentationRef.current}
      range={cellRange}
      selectedRecordCount={selectedRows.size}
      surface={surface}
      requestFocus={requestFocus}
    >
      <ProductionGridBinding
        density={density}
        grouping={grouping}
        presentationRef={semanticPresentationRef}
        publishPresentation={publishPresentation}
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
  publishPresentation,
  props,
  rowStateFor,
  sharedProps,
}: {
  readonly density: GridDensity;
  readonly grouping: SemanticDataGridProps<Row>["grouping"];
  readonly presentationRef: MutableRefObject<GridRdgPresentationModel<Row>>;
  readonly publishPresentation: (
    model: GridSemanticPresentationModel<Row>,
  ) => void;
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
      publishPresentation={publishPresentation}
      rowStateFor={rowStateFor}
      sharedProps={sharedProps}
    />
  );
}

function GridBindingFrame<Row>({
  accessibleLabel,
  bulkSelectionEnabled,
  children,
  columns,
  dataRows,
  dataState,
  density,
  draftVisible,
  requestFocus,
  interactionMode,
  keyboardAnnouncement,
  onDoubleClickCapture,
  onKeyDownCapture,
  positionMap,
  range,
  selectedRecordCount,
  surface,
}: {
  readonly accessibleLabel: string | undefined;
  readonly bulkSelectionEnabled: boolean;
  readonly children: ReactElement;
  readonly columns: readonly SemanticDataGridProps<Row>["columns"][number][];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly dataState: NonNullable<SemanticDataGridProps<Row>["dataState"]>;
  readonly density: GridDensity;
  readonly draftVisible: boolean;
  readonly requestFocus: GridHandle["requestFocus"];
  readonly interactionMode: NonNullable<
    SemanticDataGridProps<Row>["interactionMode"]
  >;
  readonly keyboardAnnouncement: string;
  readonly onDoubleClickCapture: (
    event: ReactMouseEvent<HTMLDivElement>,
  ) => void;
  readonly onKeyDownCapture?:
    | ((event: ReactKeyboardEvent<HTMLDivElement>) => void)
    | undefined;
  readonly positionMap: GridSemanticPresentationModel<Row>;
  readonly range: GridCellRange | null;
  readonly selectedRecordCount: number;
  readonly surface: GridSurfaceIdentity;
}) {
  const blocking = gridDataStateBlocksInteraction(dataState);
  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: bridges already-consumed RDG Escape without adding an interactive surface.
    <div
      className="cartulary-grid-state-frame"
      style={
        {
          "--cartulary-grid-state-row-height": `${workbookGridRowHeightPx(density)}px`,
          "--cartulary-grid-state-draft-inset": draftVisible
            ? `${workbookGridRowHeightPx(density)}px`
            : "0px",
        } as CSSProperties
      }
      onDoubleClickCapture={onDoubleClickCapture}
      onKeyDownCapture={onKeyDownCapture}
      onKeyDown={(event) => {
        // RDG passes a copy of the React event to onCellKeyDown. Its native
        // cancellation survives, but React's propagation flag stays on the
        // copy. Bridge an owned Escape before it reaches the shell ladder.
        if (event.key === "Escape" && event.nativeEvent.defaultPrevented)
          event.stopPropagation();
      }}
    >
      <div
        className="cartulary-grid-binding-content"
        inert={blocking ? true : undefined}
      >
        {children}
      </div>
      <GridOperationalStatePlane
        accessibleLabel={accessibleLabel}
        dataState={dataState}
        requestFocus={requestFocus}
        interactionMode={interactionMode}
        surface={surface}
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
        range={
          keyboardAnnouncement === "Selection collapsed to the active cell."
            ? null
            : range
        }
      />
      {keyboardAnnouncement === "" ? null : (
        <span
          aria-live={
            keyboardAnnouncement === "Selection collapsed to the active cell."
              ? "polite"
              : "assertive"
          }
          className="cartulary-grid-live-region"
          role={
            keyboardAnnouncement === "Selection collapsed to the active cell."
              ? "status"
              : "alert"
          }
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
  readonly positionMap: GridRdgPresentationModel<Row>;
  readonly setKeyboardAnnouncement: (value: string) => void;
  readonly surface: GridSurfaceIdentity;
  readonly updateCellRange: (range: GridCellRange | null) => void;
}): void {
  switch (decision.kind) {
    case "ignore":
    case "focus_draft":
    case "copy":
    case "paste":
      return;
    case "reject":
      setKeyboardAnnouncement(decision.announcement);
      return;
    case "exit_grid":
      focusAdjacentOutsideGrid(gridElement, decision.backwards);
      return;
    case "collapse_range":
      updateCellRange({ start: decision.target, end: decision.target });
      setKeyboardAnnouncement("Selection collapsed to the active cell.");
      return;
    case "begin_edit":
      updateCellRange(decision.range);
      executeBeginEditDecision(decision, args, pendingEditorSeedRef, surface);
      return;
    case "navigate":
      executeNavigateDecision(
        decision,
        args,
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
  positionMap: GridRdgPresentationModel<Row>,
  surface: GridSurfaceIdentity,
  updateCellRange: (range: GridCellRange | null) => void,
): void {
  const nextPosition = positionMap.positions.get(
    gridAnchorKey(decision.target),
  );
  if (nextPosition === undefined) return;
  updateCellRange(
    decision.range ?? { start: decision.target, end: decision.target },
  );
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
  if (expanded === null) return null;
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

function scrollToSemanticAnchor(options: {
  readonly anchor: GridCellAnchor;
  readonly positionMap: GridRdgPositionMap;
  readonly vendorHandle: DataGridHandle | null;
  readonly cell: HTMLElement | undefined;
}): boolean {
  const { anchor, positionMap, vendorHandle } = options;
  if (
    !gridSurfaceIdentitiesEqual(anchor.surface, positionMap.surface) ||
    vendorHandle === null
  )
    return false;
  const position = positionMap.positions.get(gridAnchorKey(anchor));
  if (position === undefined) return false;
  if (options.cell?.isConnected && vendorHandle.element)
    revealGridCell(vendorHandle.element, options.cell);
  else vendorHandle.scrollToCell(position);
  return true;
}

function createDraftFocusTargetRef(
  registry: Map<string, GridEditorFocusTarget>,
  fieldKey: string,
  refresh: () => void,
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
    refresh();
  };
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
  publishPresentation,
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
  readonly publishPresentation: (
    model: GridSemanticPresentationModel<Row>,
  ) => void;
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
  useLayoutEffect(() => {
    publishPresentation(groupedPresentation);
  }, [groupedPresentation, publishPresentation]);

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
              data-grid-row-version={coreRowVersion(rowProps.row) ?? undefined}
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
