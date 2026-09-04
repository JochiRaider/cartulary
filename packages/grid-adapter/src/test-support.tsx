// biome-ignore-all lint/a11y/noNoninteractiveElementToInteractiveRole: The deterministic test grid preserves the public semantic role surface.
// biome-ignore-all lint/a11y/noRedundantRoles: Explicit roles keep workbook tests independent of native accessibility-role inference.
// biome-ignore-all lint/a11y/useFocusableInteractive: This test renderer provides semantic selector compatibility, not production interaction mechanics.
import {
  gridScrollportClassName,
  workbookGridRowHeightPx,
} from "@cartulary/ui-contracts";
import {
  type ClipboardEvent,
  type CSSProperties,
  type FocusEvent,
  type ForwardedRef,
  forwardRef,
  type KeyboardEvent,
  type MutableRefObject,
  type ReactElement,
  type ReactNode,
  type RefAttributes,
  useCallback,
  useImperativeHandle,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

import "./styles.css";

import {
  assertGridRows,
  type GridActionsColumn,
  type GridCellAnchor,
  type GridCellRange,
  type GridColumn,
  type GridCoreRecordBulkSelection,
  type GridDataRow,
  type GridDraftRow,
  type GridEditCommitOutcome,
  type GridEditorAdapter,
  type GridEditorFocusTarget,
  type GridHandle,
  type GridInteractionMode,
  type GridRowGutter,
  type GridRowIdentity,
  type GridSemanticStateInput,
  type GridSortEntry,
  type GridViewportProps,
  gridRowIdentitiesEqual,
  gridRowIdentityKey,
  gridSurfaceIdentitiesEqual,
  gridUnassignedGroupLabel,
  type SemanticDataGridProps,
} from "./core";
import { GridOperationalStatePlane } from "./GridOperationalStatePlane";
import { decideSemanticActiveCellTransition } from "./semanticActiveCellPolicy";
import { resolveSemanticGridCapabilities } from "./semanticCapabilities";
import {
  planSemanticCopy,
  planSemanticFillFromRange,
  planSemanticPaste,
} from "./semanticClipboardPolicy";
import {
  gridDataStateBlocksInteraction,
  gridDataStatePresentsAuthorizedRows,
  gridDataStatePresentsDraft,
} from "./semanticDataState";
import {
  decideSemanticGridKey,
  normalizeGridKey,
  type SemanticGridDecision,
} from "./semanticKeyboardPolicy";
import {
  buildSemanticGroupBuckets,
  buildSemanticPresentationModel,
  gridAnchorKey,
  navigateSemanticPresentation,
  planSemanticPasteTargets,
  sameGridCellAnchor,
  sameGridCellRange,
  semanticPresentationContainsAnchor,
  semanticTarget,
} from "./semanticPresentation";
import {
  nextSemanticSort,
  resolveSemanticBulkSelection,
  type SemanticBulkSelectionState,
  toggleAllSemanticRecords,
  toggleSemanticRecordRange,
} from "./semanticSelectionPolicy";
import {
  gridSemanticStateClassNames,
  mergeGridSemanticState,
  resolveGridSemanticState,
} from "./semanticState";
import { resolveGridViewportStyle } from "./viewportStyle";

function coreRecordId<Row>(row: GridDataRow<Row>): string | null {
  return row.rowIdentity.kind === "core_record"
    ? row.rowIdentity.recordId
    : null;
}

type TestActiveEditor = {
  readonly fieldKey: string;
  readonly rowIdentity: GridRowIdentity;
};

type TestPresentationRow<Row> =
  | {
      readonly gridRow: GridDataRow<Row>;
      readonly key: string;
      readonly kind: "data";
    }
  | {
      readonly groupLabel: string | null;
      readonly key: string;
      readonly kind: "group";
      readonly testId: string | undefined;
    };

function buildTestPresentationRows<Row>(
  dataRows: readonly GridDataRow<Row>[],
  grouping: SemanticDataGridProps<Row>["grouping"],
): readonly TestPresentationRow<Row>[] {
  if (grouping === null || grouping === undefined) {
    return dataRows.map((gridRow) => ({
      gridRow,
      key: gridRowIdentityKey(gridRow.rowIdentity),
      kind: "data",
    }));
  }
  return buildSemanticGroupBuckets(dataRows, grouping).flatMap((bucket) => [
    {
      groupLabel: bucket.label,
      key: `group:${grouping.fieldKey}:${bucket.id}`,
      kind: "group" as const,
      testId: grouping.getTestId?.(
        grouping.fieldKey,
        bucket.value,
        bucket.label,
      ),
    },
    ...bucket.rows.map((gridRow) => ({
      gridRow,
      key: gridRowIdentityKey(gridRow.rowIdentity),
      kind: "data" as const,
    })),
  ]);
}

function useTestSupportFocus<Row>(
  presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>,
  cellElements: MutableRefObject<Map<string, HTMLTableCellElement>>,
  onActiveCellChange: ((anchor: GridCellAnchor | null) => void) | undefined,
) {
  const activeCellRef = useRef<GridCellAnchor | null>(null);
  const publishActiveCell = useCallback(
    (anchor: GridCellAnchor | null) => {
      const transition = decideSemanticActiveCellTransition(
        activeCellRef.current,
        anchor,
      );
      if (transition.kind === "no_change") return;
      activeCellRef.current = transition.anchor;
      onActiveCellChange?.(transition.anchor);
    },
    [onActiveCellChange],
  );
  const focusSemanticAnchor = useCallback(
    (anchor: GridCellAnchor) => {
      if (!semanticPresentationContainsAnchor(presentation, anchor))
        return false;
      const element = cellElements.current.get(gridAnchorKey(anchor));
      if (element === undefined) return false;
      element.focus();
      return document.activeElement === element;
    },
    [cellElements, presentation],
  );
  return { focusSemanticAnchor, publishActiveCell };
}

function useTestSupportRange(
  controlledRange: GridCellRange | null | undefined,
  onRangeChange: ((range: GridCellRange | null) => void) | undefined,
) {
  const rangeRef = useRef<GridCellRange | null>(controlledRange ?? null);
  if (controlledRange !== undefined) rangeRef.current = controlledRange;
  const updateRange = useCallback(
    (next: GridCellRange | null) => {
      if (sameGridCellRange(rangeRef.current, next)) return;
      rangeRef.current = next;
      onRangeChange?.(next);
    },
    [onRangeChange],
  );
  return { rangeRef, updateRange };
}

function useTestSupportGridHandle<Row>({
  activeEditor,
  cellElements,
  columns,
  dataRows,
  draftFocusTargets,
  editable,
  focusSemanticAnchor,
  presentation,
  ref,
  scrollElement,
  setActiveEditor,
  surface,
}: {
  readonly activeEditor: TestActiveEditor | null;
  readonly cellElements: MutableRefObject<Map<string, HTMLTableCellElement>>;
  readonly columns: readonly GridColumn<Row>[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly draftFocusTargets: MutableRefObject<
    Map<string, GridEditorFocusTarget>
  >;
  readonly editable: boolean;
  readonly focusSemanticAnchor: (anchor: GridCellAnchor) => boolean;
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly ref: ForwardedRef<GridHandle>;
  readonly scrollElement: MutableRefObject<HTMLDivElement | null>;
  readonly setActiveEditor: (editor: TestActiveEditor | null) => void;
  readonly surface: SemanticDataGridProps<Row>["surface"];
}): void {
  useImperativeHandle(
    ref,
    () => ({
      activateEdit: (anchor) => {
        const row = dataRows.find((candidate) =>
          gridRowIdentitiesEqual(candidate.rowIdentity, anchor.rowIdentity),
        );
        if (
          !editable ||
          row === undefined ||
          semanticTarget(row, anchor.fieldKey, columns, surface) === null
        ) {
          return false;
        }
        setActiveEditor({
          fieldKey: anchor.fieldKey,
          rowIdentity: anchor.rowIdentity,
        });
        return true;
      },
      cancelEdit: (anchor) => {
        if (!activeEditorsEqual(activeEditor, anchor, surface)) return false;
        setActiveEditor(null);
        return true;
      },
      focusAnchor: focusSemanticAnchor,
      focusDraftCell: (fieldKey) =>
        focusTestDraftCell(draftFocusTargets.current, fieldKey),
      focusRoot: () => focusTestRoot(scrollElement.current),
      getScrollElement: () => scrollElement.current,
      getAnchorRect: (anchor) =>
        testAnchorRect(cellElements.current, surface, anchor),
      isAnchorRendered: (anchor) =>
        semanticPresentationContainsAnchor(presentation, anchor),
      moveFocus: (current, intent) => {
        const next = navigateSemanticPresentation(
          presentation,
          current,
          intent,
        );
        return next !== null && focusSemanticAnchor(next) ? next : null;
      },
      planPasteTargets: (current, dimensions) =>
        planSemanticPasteTargets(presentation, current, dimensions),
      scrollToAnchor: (anchor) =>
        semanticPresentationContainsAnchor(presentation, anchor),
    }),
    [
      activeEditor,
      cellElements,
      columns,
      dataRows,
      draftFocusTargets,
      editable,
      focusSemanticAnchor,
      presentation,
      scrollElement,
      setActiveEditor,
      surface,
    ],
  );
}

function activeEditorsEqual(
  activeEditor: TestActiveEditor | null,
  anchor: GridCellAnchor,
  surface: SemanticDataGridProps<unknown>["surface"],
): boolean {
  return (
    activeEditor !== null &&
    activeEditor.fieldKey === anchor.fieldKey &&
    gridRowIdentitiesEqual(activeEditor.rowIdentity, anchor.rowIdentity) &&
    gridSurfaceIdentitiesEqual(surface, anchor.surface)
  );
}

function focusTestRoot(element: HTMLDivElement | null): boolean {
  if (element === null) return false;
  element.focus();
  return document.activeElement === element;
}

function createTestSupportSemanticState<Row>({
  activeRowIdentity,
  bulkSelection,
  dataState,
  editable,
  getCellState,
  getRowState,
  rows,
  surface,
}: Pick<
  SemanticDataGridProps<Row>,
  "activeRowIdentity" | "dataState" | "getCellState" | "getRowState" | "surface"
> & {
  readonly bulkSelection: SemanticDataGridProps<Row>["coreRecordBulkSelection"];
  readonly editable: boolean;
  readonly rows: readonly GridDataRow<Row>[];
}) {
  const bulkSelectionState =
    bulkSelection === undefined
      ? null
      : resolveSemanticBulkSelection(
          rows,
          bulkSelection.selectedRecordIds,
          bulkSelection.isRecordSelectable,
        );
  const rowStateFor = (row: GridDataRow<Row>): GridSemanticStateInput =>
    mergeGridSemanticState(getRowState?.(row), {
      bulkSelected:
        row.rowIdentity.kind === "core_record" &&
        bulkSelection?.selectedRecordIds.has(row.rowIdentity.recordId) === true,
      inspectorActive:
        activeRowIdentity !== null &&
        activeRowIdentity !== undefined &&
        gridRowIdentitiesEqual(row.rowIdentity, activeRowIdentity),
      readOnlyOrDerived: !editable,
      saved: true,
      stale: dataState?.kind === "stale_error",
    });
  const cellStateFor = (row: GridDataRow<Row>, column: GridColumn<Row>) => {
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
        bulkSelected: rowState.bulkSelected,
        inspectorActive: rowState.inspectorActive,
        pending: false,
        readOnlyOrDerived:
          !editable ||
          column.contractWritable !== true ||
          column.editor === undefined,
        saved: true,
        stale: false,
      },
    );
  };
  return { bulkSelectionState, cellStateFor, rowStateFor };
}

export type {
  GridActionsColumn,
  GridBlockSizing,
  GridCellAnchor,
  GridCellCopyIntent,
  GridCellPasteIntent,
  GridCellRange,
  GridCellRenderContext,
  GridCellStateContext,
  GridCellStateInput,
  GridCellTarget,
  GridChrome,
  GridClipboardDimensions,
  GridClipboardInput,
  GridClipboardPasteContract,
  GridColumn,
  GridCoreRecordBulkSelection,
  GridDataRow,
  GridDataState,
  GridDataStateAction,
  GridDensity,
  GridDraftCellRenderContext,
  GridDraftRow,
  GridEditCommitIntent,
  GridEditCommitOutcome,
  GridEditorActivation,
  GridEditorAdapter,
  GridEditorFocusTarget,
  GridEditorRenderContext,
  GridExpandedCellRange,
  GridFillIntent,
  GridGroupingDescriptor,
  GridGroupingScalar,
  GridHandle,
  GridInteractionMode,
  GridMutationIdentity,
  GridNavigationIntent,
  GridNavigationKey,
  GridPasteRowTarget,
  GridPasteTargetResolution,
  GridRowGutter,
  GridRowIdentity,
  GridRowStateInput,
  GridSemanticStateInput,
  GridSortDirection,
  GridSortEntry,
  GridStateValidation,
  GridSurfaceIdentity,
  GridViewportProps,
  SemanticDataGridProps,
} from "./core";
export function GridViewport({
  blockSizing,
  children,
  chrome,
  className,
  style,
  testId,
}: GridViewportProps) {
  return (
    <div
      className={className}
      data-testid={testId}
      style={resolveGridViewportStyle(style, chrome, blockSizing)}
    >
      {children}
    </div>
  );
}

function TestGridHeader<Row>({
  actionsColumn,
  bulkSelection,
  bulkSelectionState,
  columns,
  onSelectAll,
  onSortChange,
  rowGutter,
  sort,
}: {
  readonly actionsColumn: GridActionsColumn<Row> | undefined;
  readonly bulkSelection: GridCoreRecordBulkSelection<Row> | undefined;
  readonly bulkSelectionState: SemanticBulkSelectionState<Row> | null;
  readonly columns: readonly GridColumn<Row>[];
  readonly onSelectAll: () => void;
  readonly onSortChange: ((sort: readonly GridSortEntry[]) => void) | undefined;
  readonly rowGutter: GridRowGutter | undefined;
  readonly sort: readonly GridSortEntry[];
}) {
  return (
    <thead>
      <tr role="row">
        {bulkSelection === undefined ? null : (
          <th role="columnheader" scope="col">
            <input
              aria-label="Select all records on this page"
              checked={bulkSelectionState?.allSelected === true}
              disabled={bulkSelectionState?.selectableIds.length === 0}
              ref={(node) => {
                if (node !== null) {
                  node.indeterminate =
                    bulkSelectionState?.partiallySelected === true;
                }
              }}
              readOnly
              type="checkbox"
              onClick={onSelectAll}
            />
          </th>
        )}
        {rowGutter === undefined ? null : (
          <th role="columnheader" scope="col">
            <span data-testid={rowGutter.headerTestId}>
              {rowGutter.label ?? ""}
            </span>
          </th>
        )}
        {columns.map((column) => (
          <TestGridColumnHeader
            column={column}
            key={column.fieldKey}
            onSortChange={onSortChange}
            sort={sort}
          />
        ))}
        {actionsColumn === undefined ? null : (
          <th role="columnheader" scope="col">
            <span>{actionsColumn.label}</span>
          </th>
        )}
      </tr>
    </thead>
  );
}

function TestGridColumnHeader<Row>({
  column,
  onSortChange,
  sort,
}: {
  readonly column: GridColumn<Row>;
  readonly onSortChange: ((sort: readonly GridSortEntry[]) => void) | undefined;
  readonly sort: readonly GridSortEntry[];
}) {
  const canToggleSort =
    onSortChange !== undefined &&
    column.sortableFieldKey !== null &&
    column.sortableFieldKey !== undefined &&
    !column.sortDisabled;
  const fieldKey = column.sortableFieldKey ?? column.fieldKey;
  const sortState = sort.find((entry) => entry.fieldKey === fieldKey);
  return (
    <th role="columnheader" scope="col">
      <button
        data-grid-field-key={column.fieldKey}
        data-testid={column.headerTestId}
        disabled={!canToggleSort}
        title={column.sortDisabledReason ?? undefined}
        type="button"
        onClick={(event) =>
          onSortChange?.(
            nextSemanticSort(sort, fieldKey, event.ctrlKey || event.metaKey),
          )
        }
      >
        <span>{column.label}</span>
        {canToggleSort ? (
          <span>
            {sortState === undefined
              ? "Sort"
              : sortState.direction === "asc"
                ? "Asc"
                : "Desc"}
          </span>
        ) : null}
      </button>
    </th>
  );
}

function TestGridBody<Row>({
  actionsColumn,
  activeEditor,
  bulkSelection,
  bulkSelectionState,
  cellElements,
  cellStateFor,
  clipboardPaste,
  columns,
  draftFocusTargets,
  draftRow,
  editable,
  focusSemanticAnchor,
  interactionMode,
  onCopyCell,
  onFillCells,
  onSelectRecord,
  onSelectRow,
  pendingRangeEnd,
  presentation,
  publishActiveCell,
  rangeRef,
  renderedRows,
  rowGutter,
  rowStateFor,
  setActiveEditor,
  setKeyboardAnnouncement,
  surface,
  totalColumnCount,
  updateRange,
}: {
  readonly actionsColumn: GridActionsColumn<Row> | undefined;
  readonly activeEditor: TestActiveEditor | null;
  readonly bulkSelection: GridCoreRecordBulkSelection<Row> | undefined;
  readonly bulkSelectionState: SemanticBulkSelectionState<Row> | null;
  readonly cellElements: MutableRefObject<Map<string, HTMLTableCellElement>>;
  readonly cellStateFor: (
    row: GridDataRow<Row>,
    column: GridColumn<Row>,
  ) => GridSemanticStateInput;
  readonly clipboardPaste: SemanticDataGridProps<Row>["clipboardPaste"];
  readonly columns: readonly GridColumn<Row>[];
  readonly draftFocusTargets: MutableRefObject<
    Map<string, GridEditorFocusTarget>
  >;
  readonly draftRow: GridDraftRow<Row> | undefined;
  readonly editable: boolean;
  readonly focusSemanticAnchor: (anchor: GridCellAnchor) => boolean;
  readonly interactionMode: GridInteractionMode;
  readonly onCopyCell: SemanticDataGridProps<Row>["onCopyCell"];
  readonly onFillCells: SemanticDataGridProps<Row>["onFillCells"];
  readonly onSelectRecord: (row: GridDataRow<Row>, shiftKey: boolean) => void;
  readonly onSelectRow: ((rowIdentity: GridRowIdentity) => void) | undefined;
  readonly pendingRangeEnd: MutableRefObject<GridCellAnchor | null>;
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly publishActiveCell: (anchor: GridCellAnchor | null) => void;
  readonly rangeRef: MutableRefObject<GridCellRange | null>;
  readonly renderedRows: readonly TestPresentationRow<Row>[];
  readonly rowGutter: GridRowGutter | undefined;
  readonly rowStateFor: (row: GridDataRow<Row>) => GridSemanticStateInput;
  readonly setActiveEditor: (editor: TestActiveEditor | null) => void;
  readonly setKeyboardAnnouncement: (announcement: string) => void;
  readonly surface: SemanticDataGridProps<Row>["surface"];
  readonly totalColumnCount: number;
  readonly updateRange: (range: GridCellRange | null) => void;
}) {
  return (
    <tbody>
      {renderedRows.map((row) =>
        row.kind === "group" ? (
          <TestGridGroupRow
            groupLabel={row.groupLabel}
            key={row.key}
            testId={row.testId}
            totalColumnCount={totalColumnCount}
          />
        ) : (
          <TestGridDataRow
            actionsColumn={actionsColumn}
            activeEditor={activeEditor}
            bulkSelection={bulkSelection}
            bulkSelectionState={bulkSelectionState}
            cellElements={cellElements}
            cellStateFor={cellStateFor}
            clipboardPaste={clipboardPaste}
            columns={columns}
            editable={editable}
            focusSemanticAnchor={focusSemanticAnchor}
            gridRow={row.gridRow}
            interactionMode={interactionMode}
            key={row.key}
            onCopyCell={onCopyCell}
            onFillCells={onFillCells}
            onSelectRecord={onSelectRecord}
            onSelectRow={onSelectRow}
            pendingRangeEnd={pendingRangeEnd}
            presentation={presentation}
            publishActiveCell={publishActiveCell}
            rangeRef={rangeRef}
            rowGutter={rowGutter}
            rowState={rowStateFor(row.gridRow)}
            setActiveEditor={setActiveEditor}
            setKeyboardAnnouncement={setKeyboardAnnouncement}
            surface={surface}
            updateRange={updateRange}
          />
        ),
      )}
      {draftRow === undefined ? null : (
        <TestGridDraftRow
          actionsColumn={actionsColumn}
          bulkSelection={bulkSelection}
          columns={columns}
          draftFocusTargets={draftFocusTargets}
          draftRow={draftRow}
          rowGutter={rowGutter}
          surface={surface}
        />
      )}
    </tbody>
  );
}

function TestGridGroupRow({
  groupLabel,
  testId,
  totalColumnCount,
}: {
  readonly groupLabel: string | null;
  readonly testId: string | undefined;
  readonly totalColumnCount: number;
}) {
  return (
    <tr role="row">
      <td colSpan={totalColumnCount} role="gridcell">
        <strong data-testid={testId}>
          {groupLabel ?? gridUnassignedGroupLabel}
        </strong>
      </td>
    </tr>
  );
}

function TestGridDataRow<Row>({
  actionsColumn,
  activeEditor,
  bulkSelection,
  bulkSelectionState,
  cellElements,
  cellStateFor,
  clipboardPaste,
  columns,
  editable,
  focusSemanticAnchor,
  gridRow,
  interactionMode,
  onCopyCell,
  onFillCells,
  onSelectRecord,
  onSelectRow,
  pendingRangeEnd,
  presentation,
  publishActiveCell,
  rangeRef,
  rowGutter,
  rowState,
  setActiveEditor,
  setKeyboardAnnouncement,
  surface,
  updateRange,
}: {
  readonly actionsColumn: GridActionsColumn<Row> | undefined;
  readonly activeEditor: TestActiveEditor | null;
  readonly bulkSelection: GridCoreRecordBulkSelection<Row> | undefined;
  readonly bulkSelectionState: SemanticBulkSelectionState<Row> | null;
  readonly cellElements: MutableRefObject<Map<string, HTMLTableCellElement>>;
  readonly cellStateFor: (
    row: GridDataRow<Row>,
    column: GridColumn<Row>,
  ) => GridSemanticStateInput;
  readonly clipboardPaste: SemanticDataGridProps<Row>["clipboardPaste"];
  readonly columns: readonly GridColumn<Row>[];
  readonly editable: boolean;
  readonly focusSemanticAnchor: (anchor: GridCellAnchor) => boolean;
  readonly gridRow: GridDataRow<Row>;
  readonly interactionMode: GridInteractionMode;
  readonly onCopyCell: SemanticDataGridProps<Row>["onCopyCell"];
  readonly onFillCells: SemanticDataGridProps<Row>["onFillCells"];
  readonly onSelectRecord: (row: GridDataRow<Row>, shiftKey: boolean) => void;
  readonly onSelectRow: ((rowIdentity: GridRowIdentity) => void) | undefined;
  readonly pendingRangeEnd: MutableRefObject<GridCellAnchor | null>;
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly publishActiveCell: (anchor: GridCellAnchor | null) => void;
  readonly rangeRef: MutableRefObject<GridCellRange | null>;
  readonly rowGutter: GridRowGutter | undefined;
  readonly rowState: GridSemanticStateInput;
  readonly setActiveEditor: (editor: TestActiveEditor | null) => void;
  readonly setKeyboardAnnouncement: (announcement: string) => void;
  readonly surface: SemanticDataGridProps<Row>["surface"];
  readonly updateRange: (range: GridCellRange | null) => void;
}) {
  return (
    <tr
      {...testSemanticAttributes("row", rowState, "data row")}
      data-grid-row-identity-kind={gridRow.rowIdentity.kind}
      data-grid-record-id={coreRecordId(gridRow) ?? undefined}
      data-testid={gridRow.testId}
      role="row"
      tabIndex={onSelectRow === undefined ? undefined : 0}
      onClick={(event) => {
        if (!isInteractiveTestTarget(event.target)) {
          onSelectRow?.(gridRow.rowIdentity);
        }
      }}
      onKeyDown={(event) => {
        if (
          event.target === event.currentTarget &&
          (event.key === "Enter" || event.key === " ")
        ) {
          onSelectRow?.(gridRow.rowIdentity);
        }
      }}
    >
      {bulkSelection === undefined ? null : (
        <td role="gridcell">
          {gridRow.rowIdentity.kind !== "core_record" ||
          bulkSelection.isRecordSelectable?.(gridRow) === false ? null : (
            <input
              aria-label={`Select record ${gridRow.rowIdentity.recordId}`}
              checked={bulkSelection.selectedRecordIds.has(
                gridRow.rowIdentity.recordId,
              )}
              disabled={bulkSelectionState === null}
              readOnly
              type="checkbox"
              onClick={(event) => {
                event.stopPropagation();
                onSelectRecord(gridRow, event.shiftKey);
              }}
            />
          )}
        </td>
      )}
      {rowGutter === undefined ? null : (
        <th
          data-grid-field-key="__cartulary_row_gutter__"
          data-testid={gridRow.gutterTestId}
          scope="row"
        >
          {gridRow.gutterContent ?? gridRow.gutterLabel ?? ""}
        </th>
      )}
      {columns.map((column) => (
        <TestGridDataCell
          activeEditor={activeEditor}
          cellElements={cellElements}
          cellState={cellStateFor(gridRow, column)}
          clipboardPaste={clipboardPaste}
          column={column}
          columns={columns}
          editable={editable}
          focusSemanticAnchor={focusSemanticAnchor}
          gridRow={gridRow}
          interactionMode={interactionMode}
          key={column.fieldKey}
          onCopyCell={onCopyCell}
          onFillCells={onFillCells}
          pendingRangeEnd={pendingRangeEnd}
          presentation={presentation}
          publishActiveCell={publishActiveCell}
          rangeRef={rangeRef}
          setActiveEditor={setActiveEditor}
          setKeyboardAnnouncement={setKeyboardAnnouncement}
          surface={surface}
          updateRange={updateRange}
        />
      ))}
      {actionsColumn === undefined ? null : (
        <td role="gridcell">{actionsColumn.renderCell(gridRow)}</td>
      )}
    </tr>
  );
}

function TestGridDataCell<Row>({
  activeEditor,
  cellElements,
  cellState,
  clipboardPaste,
  column,
  columns,
  editable,
  focusSemanticAnchor,
  gridRow,
  interactionMode,
  onCopyCell,
  onFillCells,
  pendingRangeEnd,
  presentation,
  publishActiveCell,
  rangeRef,
  setActiveEditor,
  setKeyboardAnnouncement,
  surface,
  updateRange,
}: {
  readonly activeEditor: TestActiveEditor | null;
  readonly cellElements: MutableRefObject<Map<string, HTMLTableCellElement>>;
  readonly cellState: GridSemanticStateInput;
  readonly clipboardPaste: SemanticDataGridProps<Row>["clipboardPaste"];
  readonly column: GridColumn<Row>;
  readonly columns: readonly GridColumn<Row>[];
  readonly editable: boolean;
  readonly focusSemanticAnchor: (anchor: GridCellAnchor) => boolean;
  readonly gridRow: GridDataRow<Row>;
  readonly interactionMode: GridInteractionMode;
  readonly onCopyCell: SemanticDataGridProps<Row>["onCopyCell"];
  readonly onFillCells: SemanticDataGridProps<Row>["onFillCells"];
  readonly pendingRangeEnd: MutableRefObject<GridCellAnchor | null>;
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly publishActiveCell: (anchor: GridCellAnchor | null) => void;
  readonly rangeRef: MutableRefObject<GridCellRange | null>;
  readonly setActiveEditor: (editor: TestActiveEditor | null) => void;
  readonly setKeyboardAnnouncement: (announcement: string) => void;
  readonly surface: SemanticDataGridProps<Row>["surface"];
  readonly updateRange: (range: GridCellRange | null) => void;
}) {
  const anchor = {
    fieldKey: column.fieldKey,
    rowIdentity: gridRow.rowIdentity,
    surface,
  };
  const semanticState = resolveGridSemanticState(cellState, column.label);
  const editorActive = activeEditorsEqual(activeEditor, anchor, surface);
  return (
    <td
      {...testSemanticAttributes("cell", cellState, column.label)}
      data-grid-field-key={column.fieldKey}
      role="gridcell"
      tabIndex={-1}
      ref={(element) => registerTestCell(cellElements.current, anchor, element)}
      onCopy={(event) =>
        handleTestCellCopy({
          anchor,
          columns,
          dataRows: presentation.dataRows,
          event,
          onCopyCell,
          presentation,
          range: rangeRef.current,
        })
      }
      onFocus={() => {
        const preserveRange = sameGridCellAnchor(
          pendingRangeEnd.current,
          anchor,
        );
        pendingRangeEnd.current = null;
        publishActiveCell(anchor);
        if (!preserveRange) updateRange({ end: anchor, start: anchor });
      }}
      onMouseDown={() => {
        publishActiveCell(anchor);
        updateRange({ end: anchor, start: anchor });
      }}
      onClick={(event) => {
        if (
          canBeginTestEdit({
            column,
            editable,
            eventTarget: event.target,
            gridRow,
            surface,
          })
        ) {
          setActiveEditor({
            fieldKey: column.fieldKey,
            rowIdentity: gridRow.rowIdentity,
          });
        }
      }}
      onKeyDown={(event) =>
        handleTestCellKeyDown({
          anchor,
          cell: event.currentTarget,
          column,
          editable,
          event,
          focusSemanticAnchor,
          gridRow,
          interactionMode,
          onFillCells,
          pendingRangeEnd,
          presentation,
          rangeRef,
          setActiveEditor,
          setKeyboardAnnouncement,
          surface,
          updateRange,
        })
      }
      onPaste={(event) =>
        handleTestCellPaste({
          clipboardPaste,
          columns,
          editable,
          event,
          gridRow,
          presentation,
          surface,
          updateRange,
        })
      }
    >
      {semanticState.markers.map((marker) => (
        <span
          aria-label={marker.accessibleLabel}
          data-grid-state-marker={marker.kind}
          key={marker.kind}
          role="img"
        >
          {marker.glyph}
        </span>
      ))}
      {editorActive &&
      surface.kind === "view_schema" &&
      gridRow.mutationIdentity !== undefined &&
      column.editor !== undefined ? (
        <TestGridEditor
          adapter={column.editor}
          row={gridRow.data}
          target={{
            fieldKey: column.fieldKey,
            mutationIdentity: gridRow.mutationIdentity,
            rowIdentity: gridRow.rowIdentity,
            surface,
          }}
          onClose={() => setActiveEditor(null)}
        />
      ) : (
        column.renderCell({ anchor, row: gridRow.data })
      )}
    </td>
  );
}

function TestGridDraftRow<Row>({
  actionsColumn,
  bulkSelection,
  columns,
  draftFocusTargets,
  draftRow,
  rowGutter,
  surface,
}: {
  readonly actionsColumn: GridActionsColumn<Row> | undefined;
  readonly bulkSelection: GridCoreRecordBulkSelection<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly draftFocusTargets: MutableRefObject<
    Map<string, GridEditorFocusTarget>
  >;
  readonly draftRow: GridDraftRow<Row>;
  readonly rowGutter: GridRowGutter | undefined;
  readonly surface: SemanticDataGridProps<Row>["surface"];
}) {
  return (
    <tr
      data-cartulary-grid-draft-row="true"
      data-testid={draftRow.testId}
      role="row"
    >
      {bulkSelection === undefined ? null : <td role="gridcell" />}
      {rowGutter === undefined ? null : (
        <th data-grid-field-key="__cartulary_row_gutter__" scope="row">
          {draftRow.gutterContent ?? draftRow.gutterLabel ?? ""}
        </th>
      )}
      {columns.map((column) => (
        <td
          data-grid-field-key={column.fieldKey}
          key={column.fieldKey}
          role="gridcell"
        >
          {surface.kind === "view_schema"
            ? (column.renderDraftCell?.({
                fieldKey: column.fieldKey,
                focusTargetRef: createTestDraftFocusTargetRef(
                  draftFocusTargets.current,
                  column.fieldKey,
                ),
                row: draftRow.data,
                surface,
              }) ?? null)
            : null}
        </td>
      ))}
      {actionsColumn === undefined ? null : (
        <td role="gridcell">{actionsColumn.renderDraftCell?.(draftRow)}</td>
      )}
    </tr>
  );
}

function registerTestCell(
  registry: Map<string, HTMLTableCellElement>,
  anchor: GridCellAnchor,
  element: HTMLTableCellElement | null,
): void {
  const key = gridAnchorKey(anchor);
  if (element === null) registry.delete(key);
  else registry.set(key, element);
}

function isInteractiveTestTarget(target: EventTarget): boolean {
  return (
    target instanceof Element &&
    target.closest(
      "a,button,input,select,textarea,[role='button'],[role='link']",
    ) !== null
  );
}

function canBeginTestEdit<Row>({
  column,
  editable,
  eventTarget,
  gridRow,
  surface,
}: {
  readonly column: GridColumn<Row>;
  readonly editable: boolean;
  readonly eventTarget: EventTarget;
  readonly gridRow: GridDataRow<Row>;
  readonly surface: SemanticDataGridProps<Row>["surface"];
}): boolean {
  return (
    editable &&
    column.contractWritable === true &&
    column.editor !== undefined &&
    column.valueKind !== "collection" &&
    semanticTarget(gridRow, column.fieldKey, [column], surface) !== null &&
    !(
      eventTarget instanceof Element &&
      eventTarget.closest(
        "button, a, input, select, textarea, [role='button'], [data-grid-prevent-cell-edit='true']",
      ) !== null
    )
  );
}

function handleTestCellKeyDown<Row>({
  anchor,
  cell,
  column,
  editable,
  event,
  focusSemanticAnchor,
  gridRow,
  interactionMode,
  onFillCells,
  pendingRangeEnd,
  presentation,
  rangeRef,
  setActiveEditor,
  setKeyboardAnnouncement,
  surface,
  updateRange,
}: {
  readonly anchor: GridCellAnchor;
  readonly cell: HTMLTableCellElement;
  readonly column: GridColumn<Row>;
  readonly editable: boolean;
  readonly event: KeyboardEvent<HTMLTableCellElement>;
  readonly focusSemanticAnchor: (anchor: GridCellAnchor) => boolean;
  readonly gridRow: GridDataRow<Row>;
  readonly interactionMode: GridInteractionMode;
  readonly onFillCells: SemanticDataGridProps<Row>["onFillCells"];
  readonly pendingRangeEnd: MutableRefObject<GridCellAnchor | null>;
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly rangeRef: MutableRefObject<GridCellRange | null>;
  readonly setActiveEditor: (editor: TestActiveEditor | null) => void;
  readonly setKeyboardAnnouncement: (announcement: string) => void;
  readonly surface: SemanticDataGridProps<Row>["surface"];
  readonly updateRange: (range: GridCellRange | null) => void;
}): void {
  if (event.key === "Enter" && event.target !== event.currentTarget) {
    const next = navigateSemanticPresentation(presentation, anchor, {
      key: "Enter",
      shiftKey: event.shiftKey,
    });
    if (next !== null) focusSemanticAnchor(next);
    return;
  }
  if (event.target !== event.currentTarget && event.key !== "Tab") return;
  const decision = decideSemanticGridKey({
    anchor,
    column,
    editable,
    input: normalizeGridKey(event),
    model: presentation,
    pageSize: 10,
    range: rangeRef.current,
    readOnlyLabel:
      interactionMode.kind === "read_only"
        ? interactionMode.label
        : "This workbook is read-only.",
    row: gridRow,
  });
  if (
    executeTestSemanticKeyDecision({
      cell,
      decision,
      focusSemanticAnchor,
      onFillCells,
      pendingRangeEnd,
      presentation,
      setActiveEditor,
      setKeyboardAnnouncement,
      surface,
      updateRange,
    })
  ) {
    event.preventDefault();
  }
}

function handleTestCellCopy<Row>({
  anchor,
  columns,
  dataRows,
  event,
  onCopyCell,
  presentation,
  range,
}: {
  readonly anchor: GridCellAnchor;
  readonly columns: readonly GridColumn<Row>[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly event: ClipboardEvent<HTMLTableCellElement>;
  readonly onCopyCell: SemanticDataGridProps<Row>["onCopyCell"];
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly range: GridCellRange | null;
}): void {
  const plan = planSemanticCopy({
    anchor,
    columns,
    dataRows,
    model: presentation,
    range: range ?? { end: anchor, start: anchor },
  });
  if (plan === null) return;
  event.clipboardData?.setData("text/plain", plan.text);
  event.preventDefault();
  onCopyCell?.(plan.intent);
}

function handleTestCellPaste<Row>({
  clipboardPaste,
  columns,
  editable,
  event,
  gridRow,
  presentation,
  surface,
  updateRange,
}: {
  readonly clipboardPaste: SemanticDataGridProps<Row>["clipboardPaste"];
  readonly columns: readonly GridColumn<Row>[];
  readonly editable: boolean;
  readonly event: ClipboardEvent<HTMLTableCellElement>;
  readonly gridRow: GridDataRow<Row>;
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly surface: SemanticDataGridProps<Row>["surface"];
  readonly updateRange: (range: GridCellRange | null) => void;
}): void {
  if (
    !editable ||
    clipboardPaste === undefined ||
    event.target instanceof HTMLInputElement ||
    event.target instanceof HTMLTextAreaElement ||
    event.target instanceof HTMLSelectElement
  ) {
    return;
  }
  const input = clipboardPaste.decode(
    event.clipboardData?.getData("text/plain") ?? "",
  );
  const intent = planSemanticPaste({
    input,
    model: presentation,
    target: semanticTarget(
      gridRow,
      event.currentTarget.dataset.gridFieldKey ?? "",
      columns,
      surface,
    ),
  });
  if (intent === null) return;
  event.preventDefault();
  updateRange(intent.range);
  clipboardPaste.onPaste(intent);
}

function useSemanticDataGridTestSupport<Row>(
  {
    accessibleLabel,
    activeRowIdentity = null,
    allowPasteCreateRows = false,
    actionsColumn,
    cellRange: controlledCellRange,
    coreRecordBulkSelection,
    columns,
    dataState = { kind: "ready" },
    density = "default",
    draftRow: ownerDraftRow,
    fillViewportInline = false,
    getCellState,
    getRowState,
    grouping = null,
    interactionMode,
    clipboardPaste,
    onActiveCellChange,
    onCellRangeChange,
    onCopyCell,
    onFillCells,
    onSelectRow,
    onSortChange,
    dataRows: ownerDataRows,
    rowGutter,
    sort = [],
    surface,
  }: SemanticDataGridProps<Row>,
  ref: ForwardedRef<GridHandle>,
) {
  const capabilities = resolveSemanticGridCapabilities({
    actionsColumn,
    allowPasteCreateRows,
    clipboardPaste,
    coreRecordBulkSelection,
    draftRow: ownerDraftRow,
    interactionMode,
    onFillCells,
    surface,
  });
  const effectiveInteractionMode = capabilities.interactionMode;
  const editable = capabilities.editable;
  const effectiveBulkSelection = capabilities.bulkSelection;
  const dataRows = gridDataStatePresentsAuthorizedRows(dataState)
    ? ownerDataRows
    : [];
  const effectiveDraftRow = gridDataStatePresentsDraft(dataState)
    ? capabilities.draftRow
    : undefined;
  const cellElements = useRef(new Map<string, HTMLTableCellElement>());
  const draftFocusTargets = useRef(new Map<string, GridEditorFocusTarget>());
  const scrollElement = useRef<HTMLDivElement>(null);
  const pendingRangeEnd = useRef<GridCellAnchor | null>(null);
  const selectionAnchorRecordId = useRef<string | null>(null);
  const [activeEditor, setActiveEditor] = useState<TestActiveEditor | null>(
    null,
  );
  const [keyboardAnnouncement, setKeyboardAnnouncement] = useState("");
  const { rangeRef, updateRange } = useTestSupportRange(
    controlledCellRange,
    onCellRangeChange,
  );
  const renderedRows = buildTestPresentationRows(dataRows, grouping);
  const semanticPresentation = buildSemanticPresentationModel({
    allowCreateRows: allowPasteCreateRows && grouping === null,
    columns,
    dataRows,
    fieldKeys: columns.map((column) => column.fieldKey),
    surface,
  });
  const { focusSemanticAnchor, publishActiveCell } = useTestSupportFocus(
    semanticPresentation,
    cellElements,
    onActiveCellChange,
  );
  useTestSupportGridHandle({
    activeEditor,
    cellElements,
    columns,
    dataRows,
    draftFocusTargets,
    editable,
    focusSemanticAnchor,
    presentation: semanticPresentation,
    ref,
    scrollElement,
    setActiveEditor,
    surface,
  });
  assertGridRows(ownerDataRows);

  const { bulkSelectionState, cellStateFor, rowStateFor } =
    createTestSupportSemanticState({
      activeRowIdentity,
      bulkSelection: effectiveBulkSelection,
      dataState,
      editable,
      getCellState,
      getRowState,
      rows: dataRows,
      surface,
    });

  const totalColumnCount =
    columns.length +
    (effectiveBulkSelection === undefined ? 0 : 1) +
    (rowGutter === undefined ? 0 : 1) +
    (actionsColumn === undefined ? 0 : 1);

  return (
    <div
      className="cartulary-grid-state-frame"
      style={
        {
          "--cartulary-grid-state-row-height": `${workbookGridRowHeightPx(density)}px`,
          "--cartulary-grid-state-draft-inset":
            editable && effectiveDraftRow !== undefined
              ? `${workbookGridRowHeightPx(density)}px`
              : "0px",
        } as CSSProperties
      }
    >
      <div
        className="cartulary-grid-binding-content"
        inert={gridDataStateBlocksInteraction(dataState) ? true : undefined}
      >
        <div
          className={gridScrollportClassName()}
          ref={scrollElement}
          style={
            fillViewportInline ? { minWidth: 0, width: "100%" } : undefined
          }
          tabIndex={-1}
        >
          <table
            aria-label={accessibleLabel}
            aria-busy={
              dataState.kind === "initial_loading" ||
              dataState.kind === "refreshing"
            }
            aria-readonly={!editable}
            role="grid"
            style={
              fillViewportInline ? { minWidth: 0, width: "100%" } : undefined
            }
          >
            <TestGridHeader
              actionsColumn={actionsColumn}
              bulkSelection={effectiveBulkSelection}
              bulkSelectionState={bulkSelectionState}
              columns={columns}
              onSelectAll={() => {
                selectionAnchorRecordId.current = null;
                if (
                  effectiveBulkSelection !== undefined &&
                  bulkSelectionState !== null
                ) {
                  effectiveBulkSelection.onSelectedRecordIdsChange(
                    toggleAllSemanticRecords(bulkSelectionState),
                  );
                }
              }}
              onSortChange={onSortChange}
              rowGutter={rowGutter}
              sort={sort}
            />
            <TestGridBody
              actionsColumn={actionsColumn}
              activeEditor={activeEditor}
              bulkSelection={effectiveBulkSelection}
              bulkSelectionState={bulkSelectionState}
              cellElements={cellElements}
              cellStateFor={cellStateFor}
              clipboardPaste={clipboardPaste}
              columns={columns}
              draftFocusTargets={draftFocusTargets}
              draftRow={effectiveDraftRow}
              editable={editable}
              focusSemanticAnchor={focusSemanticAnchor}
              interactionMode={effectiveInteractionMode}
              onCopyCell={onCopyCell}
              onFillCells={onFillCells}
              onSelectRecord={(gridRow, shiftKey) => {
                const recordId = coreRecordId(gridRow);
                if (
                  recordId === null ||
                  effectiveBulkSelection === undefined ||
                  bulkSelectionState === null
                ) {
                  return;
                }
                const next = toggleSemanticRecordRange({
                  anchorRecordId: selectionAnchorRecordId.current,
                  recordId,
                  selectableRows: bulkSelectionState.selectableRows,
                  selectedRecordIds: effectiveBulkSelection.selectedRecordIds,
                  shiftKey,
                });
                selectionAnchorRecordId.current = recordId;
                effectiveBulkSelection.onSelectedRecordIdsChange(next);
              }}
              onSelectRow={onSelectRow}
              pendingRangeEnd={pendingRangeEnd}
              presentation={semanticPresentation}
              publishActiveCell={publishActiveCell}
              rangeRef={rangeRef}
              renderedRows={renderedRows}
              rowGutter={rowGutter}
              rowStateFor={rowStateFor}
              setActiveEditor={setActiveEditor}
              setKeyboardAnnouncement={setKeyboardAnnouncement}
              surface={surface}
              totalColumnCount={totalColumnCount}
              updateRange={updateRange}
            />
          </table>
        </div>
      </div>
      <GridOperationalStatePlane
        accessibleLabel={accessibleLabel}
        dataState={dataState}
        focusRoot={() => focusTestRoot(scrollElement.current)}
        interactionMode={effectiveInteractionMode}
        surface={surface}
      />
      {keyboardAnnouncement === "" ? null : (
        <span aria-live="assertive" role="alert">
          {keyboardAnnouncement}
        </span>
      )}
    </div>
  );
}

function executeTestSemanticKeyDecision<Row>({
  cell,
  decision,
  focusSemanticAnchor,
  onFillCells,
  pendingRangeEnd,
  presentation,
  setActiveEditor,
  setKeyboardAnnouncement,
  surface,
  updateRange,
}: {
  readonly cell: HTMLTableCellElement;
  readonly decision: SemanticGridDecision;
  readonly focusSemanticAnchor: (anchor: GridCellAnchor) => boolean;
  readonly onFillCells: SemanticDataGridProps<Row>["onFillCells"];
  readonly pendingRangeEnd: MutableRefObject<GridCellAnchor | null>;
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly setActiveEditor: (value: {
    readonly fieldKey: string;
    readonly rowIdentity: GridRowIdentity;
  }) => void;
  readonly setKeyboardAnnouncement: (value: string) => void;
  readonly surface: SemanticDataGridProps<Row>["surface"];
  readonly updateRange: (range: GridCellRange | null) => void;
}): boolean {
  switch (decision.kind) {
    case "ignore":
    case "copy":
    case "paste":
      return false;
    case "reject":
      setKeyboardAnnouncement(decision.announcement);
      return true;
    case "exit_grid":
      cell.blur();
      return true;
    case "begin_edit":
      setActiveEditor({
        fieldKey: decision.seed.anchor.fieldKey,
        rowIdentity: decision.seed.anchor.rowIdentity,
      });
      return true;
    case "navigate":
      return executeTestNavigateDecision(
        decision,
        focusSemanticAnchor,
        pendingRangeEnd,
        updateRange,
      );
    case "fill":
      return executeTestFillDecision({
        decision,
        onFillCells,
        presentation,
        setKeyboardAnnouncement,
        surface,
        updateRange,
      });
  }
}

function executeTestNavigateDecision(
  decision: Extract<SemanticGridDecision, { readonly kind: "navigate" }>,
  focusSemanticAnchor: (anchor: GridCellAnchor) => boolean,
  pendingRangeEnd: MutableRefObject<GridCellAnchor | null>,
  updateRange: (range: GridCellRange | null) => void,
): boolean {
  if (decision.range !== null) pendingRangeEnd.current = decision.target;
  const focused = focusSemanticAnchor(decision.target);
  if (!focused) pendingRangeEnd.current = null;
  if (focused && decision.range !== null) updateRange(decision.range);
  return focused;
}

function executeTestFillDecision<Row>({
  decision,
  onFillCells,
  presentation,
  setKeyboardAnnouncement,
  surface,
  updateRange,
}: {
  readonly decision: Extract<SemanticGridDecision, { readonly kind: "fill" }>;
  readonly onFillCells: SemanticDataGridProps<Row>["onFillCells"];
  readonly presentation: ReturnType<typeof buildSemanticPresentationModel<Row>>;
  readonly setKeyboardAnnouncement: (value: string) => void;
  readonly surface: SemanticDataGridProps<Row>["surface"];
  readonly updateRange: (range: GridCellRange | null) => void;
}): boolean {
  const intent = planSemanticFillFromRange({
    columns: presentation.columns,
    dataRows: presentation.dataRows,
    model: presentation,
    range: decision.range,
    surface,
  });
  if (intent === null) {
    setKeyboardAnnouncement(
      "Select a writable one-column range before using fill down.",
    );
    return true;
  }
  updateRange(intent.range);
  onFillCells?.(intent);
  setKeyboardAnnouncement(
    `Filled ${intent.targets.length} cells from the top selected cell.`,
  );
  return true;
}

function createTestDraftFocusTargetRef(
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

function focusTestDraftCell(
  registry: ReadonlyMap<string, GridEditorFocusTarget>,
  fieldKey: string,
): boolean {
  const element = registry.get(fieldKey);
  if (
    element === undefined ||
    !element.isConnected ||
    element.disabled ||
    element.hidden ||
    element.closest("[hidden], [aria-hidden='true']") !== null
  ) {
    return false;
  }
  element.focus({ preventScroll: true });
  return document.activeElement === element;
}

function testAnchorRect(
  cellElements: ReadonlyMap<string, HTMLTableCellElement>,
  surface: SemanticDataGridProps<unknown>["surface"],
  anchor: GridCellAnchor,
): DOMRectReadOnly | null {
  if (!gridSurfaceIdentitiesEqual(surface, anchor.surface)) return null;
  const element = cellElements.get(gridAnchorKey(anchor));
  return element?.isConnected === true &&
    element.closest("[hidden], [aria-hidden='true']") === null
    ? element.getBoundingClientRect()
    : null;
}

function SemanticDataGridInner<Row>(
  props: SemanticDataGridProps<Row>,
  ref: ForwardedRef<GridHandle>,
) {
  // biome-ignore lint/correctness/useHookAtTopLevel: this generic function is passed directly to React.forwardRef below.
  return useSemanticDataGridTestSupport(props, ref);
}

export const SemanticDataGrid = forwardRef(SemanticDataGridInner) as <Row>(
  props: SemanticDataGridProps<Row> & RefAttributes<GridHandle>,
) => ReactElement;

function TestGridEditor<Row>({
  adapter,
  row,
  target,
  onClose,
}: {
  readonly adapter: GridEditorAdapter<Row>;
  readonly row: Row;
  readonly target: Parameters<GridEditorAdapter<Row>["commit"]>[0]["target"];
  readonly onClose: () => void;
}) {
  const [draftValue, setDraftValue] = useState(() =>
    adapter.initialDraftValue(row),
  );
  const [outcome, setOutcome] = useState<GridEditCommitOutcome | null>(null);
  const [pending, setPending] = useState(false);
  const focusTarget = useRef<GridEditorFocusTarget | null>(null);
  const commitPromises = useRef(
    new Map<string, Promise<GridEditCommitOutcome>>(),
  );
  const latestCommitSequence = useRef(0);
  const focusTargetRef = useCallback(
    (element: GridEditorFocusTarget | null) => {
      focusTarget.current = element;
    },
    [],
  );
  useLayoutEffect(() => {
    const element = focusTarget.current;
    if (element === null) return;
    element.focus();
    if (
      element instanceof HTMLInputElement ||
      element instanceof HTMLTextAreaElement
    ) {
      const end = element.value.length;
      element.setSelectionRange(end, end);
    }
  }, []);
  const commit = useCallback(
    (draftValueOverride?: unknown) => {
      const requestedDraft =
        draftValueOverride === undefined ? draftValue : draftValueOverride;
      const draftKey = testGridEditorDraftKey(requestedDraft);
      const duplicate = commitPromises.current.get(draftKey);
      if (duplicate !== undefined) return duplicate;
      const sequence = latestCommitSequence.current + 1;
      latestCommitSequence.current = sequence;
      setPending(true);
      setOutcome(null);
      const request = adapter
        .commit({
          draftValue: requestedDraft,
          row,
          target,
        })
        .then((next) => {
          commitPromises.current.delete(draftKey);
          setPending(commitPromises.current.size > 0);
          const isLatest = latestCommitSequence.current === sequence;
          if (isLatest) setOutcome(next);
          if (
            next.kind === "accepted" &&
            isLatest &&
            commitPromises.current.size === 0
          ) {
            onClose();
          }
          return next;
        });
      commitPromises.current.set(draftKey, request);
      return request;
    },
    [adapter, draftValue, onClose, row, target],
  );
  const handleBlur = (event: FocusEvent<HTMLDivElement>) => {
    if (event.currentTarget.contains(event.relatedTarget as Node | null)) {
      return;
    }
    void commit();
  };
  return (
    <div
      onBlurCapture={handleBlur}
      onKeyDownCapture={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          onClose();
        }
      }}
    >
      {
        adapter.renderEditor({
          activation: { initialSelection: "end", source: "pointer" },
          cancel: onClose,
          commit: async (draftValueOverride) => {
            await commit(draftValueOverride);
          },
          draftValue,
          focusTargetRef,
          outcome,
          pending,
          row,
          setDraftValue,
          target,
        }) as ReactNode
      }
    </div>
  );
}

function testGridEditorDraftKey(value: unknown): string {
  if (value === undefined) return "undefined";
  if (typeof value === "string") return `string:${value}`;
  if (typeof value === "number") return `number:${String(value)}`;
  if (typeof value === "boolean") return `boolean:${String(value)}`;
  try {
    return `json:${JSON.stringify(value)}`;
  } catch {
    return `opaque:${String(value)}`;
  }
}

function testSemanticAttributes(
  scope: "cell" | "row",
  input: GridSemanticStateInput,
  label: string,
) {
  const state = resolveGridSemanticState(input, label);
  return {
    "aria-busy": state.stateIds.includes("pending") || undefined,
    "aria-current":
      scope === "row" && state.stateIds.includes("inspector-active")
        ? true
        : undefined,
    "aria-description": state.description,
    "aria-invalid": state.stateIds.includes("invalid") || undefined,
    "aria-readonly":
      scope === "cell" && state.stateIds.includes("read-only")
        ? true
        : undefined,
    "aria-selected":
      scope === "row" && state.stateIds.includes("bulk-selected")
        ? true
        : undefined,
    className: gridSemanticStateClassNames(scope, state),
    "data-grid-primary-state": state.primary,
    "data-grid-semantic-states": state.stateIds.join(" "),
  } as const;
}
