import {
  type KeyboardEvent,
  type ReactNode,
  useMemo,
  useRef,
  useState,
} from "react";
import type { Column } from "react-data-grid";
import type {
  GridActionsColumn,
  GridCellAnchor,
  GridColumn,
  GridDraftRow,
  GridEditCommitOutcome,
  GridEditorAdapter,
  GridRecordRow,
  GridRowGutter,
  GridSemanticStateInput,
} from "./core";
import {
  type GridResolvedSemanticState,
  gridSemanticStateClassNames,
  mergeGridSemanticState,
  resolveGridSemanticState,
} from "./semanticState";

export const gridActionsColumnKey = "__cartulary_actions__";
export const gridRowGutterColumnKey = "__cartulary_row_gutter__";
export const gridSelectionColumnKey = "__cartulary_selection__";

export type GridCompiledBulkSelection<Row> = {
  readonly allSelected: boolean;
  readonly partiallySelected: boolean;
  readonly selectedRecordIds: ReadonlySet<string>;
  readonly selectableRecordCount: number;
  readonly onSelectAll: () => void;
  readonly onToggleRecord: (row: GridRecordRow<Row>, shiftKey: boolean) => void;
  readonly isRecordSelectable: (row: GridRecordRow<Row>) => boolean;
};

type CompileGridColumnsInput<Row> = {
  readonly actionsColumn: GridActionsColumn<Row> | undefined;
  readonly bulkSelection: GridCompiledBulkSelection<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly cellStateFor: (
    row: GridRecordRow<Row>,
    column: GridColumn<Row>,
  ) => GridSemanticStateInput;
  readonly editable: boolean;
  readonly isCellRangeSelected: (
    row: GridRecordRow<Row>,
    column: GridColumn<Row>,
  ) => boolean;
  readonly consumeEditorSeed: (
    target: Parameters<GridEditorAdapter<Row>["commit"]>[0]["target"],
  ) => { readonly value: unknown } | null;
  readonly onEditorKeyboardAction: (
    target: Parameters<GridEditorAdapter<Row>["commit"]>[0]["target"],
    action:
      | { readonly kind: "exit"; readonly backwards: boolean }
      | { readonly kind: "move"; readonly rowDelta: -1 | 1 },
  ) => void;
  readonly registerSemanticCell: (
    anchor: GridCellAnchor,
    cell: HTMLElement | null,
    token: object,
  ) => void;
  readonly rowGutter: GridRowGutter | undefined;
  readonly viewSchemaId: string;
  readonly onPasteCellContent:
    | ((
        row: GridRecordRow<Row>,
        fieldKey: string,
        clipboardText: string,
      ) => boolean)
    | undefined;
};

export function compileGridColumns<Row>({
  actionsColumn,
  bulkSelection,
  cellStateFor,
  columns,
  consumeEditorSeed,
  editable,
  isCellRangeSelected,
  onEditorKeyboardAction,
  onPasteCellContent,
  registerSemanticCell,
  rowGutter,
  viewSchemaId,
}: CompileGridColumnsInput<Row>): readonly Column<
  GridRecordRow<Row>,
  GridDraftRow<Row>
>[] {
  const firstColumnKey =
    rowGutter !== undefined
      ? gridRowGutterColumnKey
      : (columns[0]?.fieldKey ??
        (actionsColumn === undefined ? undefined : gridActionsColumnKey));
  const compiled: Array<Column<GridRecordRow<Row>, GridDraftRow<Row>>> = [];

  if (bulkSelection !== undefined) {
    compiled.push({
      cellClass: "cartulary-grid-selection-cell",
      frozen: true,
      headerCellClass:
        "cartulary-grid-header-cell cartulary-grid-selection-header-cell",
      key: gridSelectionColumnKey,
      minWidth: 42,
      name: "Select records",
      renderHeaderCell: () => (
        <input
          aria-label="Select all records on this page"
          checked={bulkSelection.allSelected}
          disabled={bulkSelection.selectableRecordCount === 0}
          ref={(node) => {
            if (node !== null) {
              node.indeterminate = bulkSelection.partiallySelected;
            }
            markSemanticHeaderCell(node, {
              accessibleLabel: "Select records",
              fieldKey: gridSelectionColumnKey,
              testId: undefined,
            });
          }}
          readOnly
          type="checkbox"
          onClick={(event) => {
            event.stopPropagation();
            bulkSelection.onSelectAll();
          }}
        />
      ),
      renderCell: ({ row }) =>
        bulkSelection.isRecordSelectable(row) ? (
          <input
            aria-label={`Select record ${row.recordId}`}
            checked={bulkSelection.selectedRecordIds.has(row.recordId)}
            data-grid-field-key={gridSelectionColumnKey}
            readOnly
            type="checkbox"
            onClick={(event) => {
              event.stopPropagation();
              bulkSelection.onToggleRecord(row, event.shiftKey);
            }}
          />
        ) : null,
      renderSummaryCell: () => null,
      resizable: false,
      sortable: false,
      width: 42,
    });
  }

  if (rowGutter !== undefined) {
    compiled.push({
      cellClass: "cartulary-grid-gutter-cell",
      frozen: true,
      headerCellClass:
        "cartulary-grid-header-cell cartulary-grid-gutter-header-cell",
      key: gridRowGutterColumnKey,
      minWidth: rowGutter.minWidth,
      name: (
        <span
          ref={(node) =>
            markSemanticHeaderCell(node, {
              accessibleLabel: "Row gutter",
              fieldKey: gridRowGutterColumnKey,
              testId: rowGutter.headerTestId,
            })
          }
        >
          {rowGutter.label}
        </span>
      ),
      renderCell: ({ row }) => (
        <span
          data-grid-field-key={gridRowGutterColumnKey}
          ref={(node) => markSemanticGutterCell(node, row.gutterTestId)}
        >
          {row.gutterContent ?? row.gutterLabel ?? ""}
        </span>
      ),
      renderSummaryCell: ({ row }) => (
        <span
          {...draftMarker(row, gridRowGutterColumnKey, firstColumnKey)}
          data-grid-field-key={gridRowGutterColumnKey}
        >
          {row.gutterContent ?? row.gutterLabel ?? ""}
        </span>
      ),
      resizable: false,
      width: rowGutter.width,
    });
  }

  for (const column of columns) {
    compiled.push({
      cellClass: (row) =>
        [
          dataCellClass(
            column.align,
            resolveGridSemanticState(cellStateFor(row, column), column.label),
          ),
          isCellRangeSelected(row, column)
            ? "cartulary-grid-cell-is-range-selected"
            : undefined,
        ]
          .filter((value) => value !== undefined)
          .join(" "),
      editable:
        editable &&
        column.contractWritable === true &&
        column.editor !== undefined,
      key: column.fieldKey,
      headerCellClass: "cartulary-grid-header-cell",
      minWidth: column.minWidth,
      name: column.label,
      renderHeaderCell: ({ sortDirection }) => (
        <span
          data-grid-field-key={column.fieldKey}
          ref={(node) =>
            markSemanticHeaderCell(node, {
              accessibleLabel: column.label,
              fieldKey: column.fieldKey,
              testId: column.headerTestId,
              title: column.sortDisabledReason ?? undefined,
            })
          }
        >
          {column.label}
          {sortDirection === "ASC" ? " Asc" : null}
          {sortDirection === "DESC" ? " Desc" : null}
        </span>
      ),
      renderCell: ({ row }) => {
        const semanticState = resolveGridSemanticState(
          cellStateFor(row, column),
          column.label,
        );
        const anchor = {
          fieldKey: column.fieldKey,
          recordId: row.recordId,
          viewSchemaId,
        };
        return (
          <SemanticGridCellContent
            anchor={anchor}
            fieldKey={column.fieldKey}
            rangeSelected={isCellRangeSelected(row, column)}
            registerSemanticCell={registerSemanticCell}
            semanticState={semanticState}
            onPaste={(clipboardText) =>
              onPasteCellContent?.(row, column.fieldKey, clipboardText) === true
            }
          >
            {column.renderCell({
              anchor,
              row: row.data,
            })}
          </SemanticGridCellContent>
        );
      },
      renderEditCell:
        editable &&
        column.contractWritable === true &&
        column.editor !== undefined
          ? ({ onClose, row }) => (
              <SemanticGridEditor
                adapter={column.editor as GridEditorAdapter<Row>}
                baseState={cellStateFor(row, column)}
                editorSeed={consumeEditorSeed({
                  baseRowVersion: row.rowVersion,
                  fieldKey: column.fieldKey,
                  recordId: row.recordId,
                  viewSchemaId,
                })}
                fieldLabel={column.label}
                registerSemanticCell={registerSemanticCell}
                row={row.data}
                target={{
                  baseRowVersion: row.rowVersion,
                  fieldKey: column.fieldKey,
                  recordId: row.recordId,
                  viewSchemaId,
                }}
                onClose={(accepted) => onClose(accepted, true)}
                onKeyboardAction={onEditorKeyboardAction}
              />
            )
          : undefined,
      renderSummaryCell: ({ row }) => (
        // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns this private bottom-draft cell; the wrapper only prevents duplicate vendor navigation after a nested control handles the key.
        <span
          {...draftMarker(row, column.fieldKey, firstColumnKey)}
          className="cartulary-grid-cell-content"
          data-grid-field-key={column.fieldKey}
          onKeyDown={stopVendorNavigationForInteractiveContent}
          role="presentation"
        >
          {column.renderDraftCell?.({
            fieldKey: column.fieldKey,
            row: row.data,
            viewSchemaId,
          }) ?? null}
        </span>
      ),
      resizable: true,
      summaryCellClass: "cartulary-grid-data-cell cartulary-grid-draft-cell",
      sortable:
        column.sortableFieldKey !== null &&
        column.sortableFieldKey !== undefined &&
        !column.sortDisabled,
      width: column.width,
    });
  }

  if (actionsColumn !== undefined) {
    compiled.push({
      cellClass: "cartulary-grid-actions-cell",
      headerCellClass:
        "cartulary-grid-header-cell cartulary-grid-actions-header-cell",
      key: gridActionsColumnKey,
      minWidth: actionsColumn.minWidth,
      name: (
        <span
          ref={(node) =>
            markSemanticHeaderCell(node, {
              accessibleLabel: actionsColumn.label,
              fieldKey: gridActionsColumnKey,
              testId: actionsColumn.headerTestId,
            })
          }
        >
          {actionsColumn.label}
        </span>
      ),
      renderCell: ({ row }) => (
        // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns the gridcell; this wrapper only prevents duplicate vendor navigation after a nested action handles the key.
        <span
          className="cartulary-grid-cell-content"
          data-grid-field-key={gridActionsColumnKey}
          onKeyDown={stopVendorNavigationForInteractiveContent}
          role="presentation"
        >
          {actionsColumn.renderCell(row)}
        </span>
      ),
      renderSummaryCell: ({ row }) => (
        // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns this private bottom-draft cell; the wrapper only prevents duplicate vendor navigation after a nested action handles the key.
        <span
          {...draftMarker(row, gridActionsColumnKey, firstColumnKey)}
          className="cartulary-grid-cell-content"
          data-grid-field-key={gridActionsColumnKey}
          onKeyDown={stopVendorNavigationForInteractiveContent}
          role="presentation"
        >
          {actionsColumn.renderDraftCell?.(row)}
        </span>
      ),
      resizable: false,
      width: actionsColumn.width,
    });
  }

  return compiled;
}

function SemanticGridEditor<Row>({
  adapter,
  baseState,
  editorSeed,
  fieldLabel,
  registerSemanticCell,
  row,
  target,
  onClose,
  onKeyboardAction,
}: {
  readonly adapter: GridEditorAdapter<Row>;
  readonly baseState: GridSemanticStateInput;
  readonly editorSeed: { readonly value: unknown } | null;
  readonly fieldLabel: string;
  readonly registerSemanticCell: (
    anchor: GridCellAnchor,
    cell: HTMLElement | null,
    token: object,
  ) => void;
  readonly row: Row;
  readonly target: Parameters<GridEditorAdapter<Row>["commit"]>[0]["target"];
  readonly onClose: (accepted: boolean) => void;
  readonly onKeyboardAction: (
    target: Parameters<GridEditorAdapter<Row>["commit"]>[0]["target"],
    action:
      | { readonly kind: "exit"; readonly backwards: boolean }
      | { readonly kind: "move"; readonly rowDelta: -1 | 1 },
  ) => void;
}) {
  const [draftValue, setDraftValue] = useState(() =>
    editorSeed === null ? adapter.initialDraftValue(row) : editorSeed.value,
  );
  const [outcome, setOutcome] = useState<GridEditCommitOutcome | null>(null);
  const [pending, setPending] = useState(false);
  const pendingRef = useRef(false);
  const semanticState = useMemo(
    () =>
      resolveGridSemanticState(
        mergeGridSemanticState(baseState, {
          conflicted: outcome?.kind === "conflict",
          invalid: outcome?.kind === "validation_error" ? outcome : false,
          pending,
          stale: outcome?.kind === "stale_target",
        }),
        fieldLabel,
      ),
    [baseState, fieldLabel, outcome, pending],
  );
  const commitDraft = async (
    draftValueOverride?: unknown,
  ): Promise<GridEditCommitOutcome | null> => {
    if (pendingRef.current) return null;
    pendingRef.current = true;
    setPending(true);
    setOutcome(null);
    let next: GridEditCommitOutcome;
    try {
      next = await adapter.commit({
        draftValue:
          draftValueOverride === undefined ? draftValue : draftValueOverride,
        row,
        target,
      });
    } catch (error) {
      next = {
        kind: "rejected_mutation",
        message:
          error instanceof Error ? error.message : "Mutation was rejected.",
      };
    }
    pendingRef.current = false;
    setPending(false);
    setOutcome(next);
    if (next.kind === "accepted") {
      onClose(true);
    }
    return next;
  };
  const commit = async (draftValueOverride?: unknown) => {
    await commitDraft(draftValueOverride);
  };
  const handleKeyboardAction = async (
    action:
      | { readonly kind: "exit"; readonly backwards: boolean }
      | { readonly kind: "move"; readonly rowDelta: -1 | 1 },
  ) => {
    const next = await commitDraft();
    if (next?.kind === "accepted") {
      queueMicrotask(() => onKeyboardAction(target, action));
    }
  };
  return (
    <SemanticGridCellContent
      anchor={target}
      fieldKey={target.fieldKey}
      registerSemanticCell={registerSemanticCell}
      semanticState={semanticState}
      onKeyDownCapture={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          onClose(false);
          return;
        }
        if (event.key === "Tab") {
          event.preventDefault();
          event.stopPropagation();
          void handleKeyboardAction({
            backwards: event.shiftKey,
            kind: "exit",
          });
          return;
        }
        if (
          event.key === "Enter" &&
          !(event.shiftKey && event.target instanceof HTMLTextAreaElement)
        ) {
          event.preventDefault();
          event.stopPropagation();
          void handleKeyboardAction({
            kind: "move",
            rowDelta: event.shiftKey ? -1 : 1,
          });
        }
      }}
    >
      {adapter.renderEditor({
        cancel: () => onClose(false),
        commit,
        draftValue,
        outcome,
        pending,
        row,
        setDraftValue,
        target,
      })}
      {outcome === null || outcome.kind === "accepted" ? null : (
        <span aria-live="assertive" role="alert">
          {outcome.message}
        </span>
      )}
    </SemanticGridCellContent>
  );
}

function SemanticGridCellContent({
  anchor,
  children,
  fieldKey,
  onKeyDownCapture,
  onPaste,
  rangeSelected = false,
  registerSemanticCell,
  semanticState,
}: {
  readonly anchor: GridCellAnchor;
  readonly children: ReactNode;
  readonly fieldKey: string;
  readonly onKeyDownCapture?:
    | ((event: KeyboardEvent<HTMLSpanElement>) => void)
    | undefined;
  readonly onPaste?: ((clipboardText: string) => boolean) | undefined;
  readonly rangeSelected?: boolean | undefined;
  readonly registerSemanticCell: (
    anchor: GridCellAnchor,
    cell: HTMLElement | null,
    token: object,
  ) => void;
  readonly semanticState: GridResolvedSemanticState;
}) {
  const registrationToken = useRef({}).current;
  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns the gridcell; this wrapper provides semantic state and prevents duplicate vendor handling after a nested control handles the event.
    <span
      className="cartulary-grid-cell-content"
      data-grid-field-key={fieldKey}
      data-grid-primary-state={semanticState.primary}
      ref={(node) => {
        const cell = markSemanticDataCell(node, semanticState, rangeSelected);
        registerSemanticCell(anchor, cell, registrationToken);
      }}
      onKeyDown={(event) =>
        stopVendorNavigationForInteractiveContent(event, true)
      }
      onKeyDownCapture={onKeyDownCapture}
      onCopy={(event) => {
        if (isInteractiveEditorTarget(event.target)) {
          event.stopPropagation();
        }
      }}
      onPaste={(event) => {
        if (isInteractiveEditorTarget(event.target)) {
          event.stopPropagation();
          return;
        }
        if (onPaste === undefined) {
          return;
        }
        if (onPaste(event.clipboardData.getData("text/plain"))) {
          event.preventDefault();
          event.stopPropagation();
        }
      }}
      role="presentation"
    >
      {semanticState.markers.map((marker) => (
        <span
          aria-label={marker.accessibleLabel}
          className={`cartulary-grid-state-marker cartulary-grid-state-marker-${marker.kind}`}
          data-grid-state-marker={marker.kind}
          key={marker.kind}
          role="img"
          title={marker.accessibleLabel}
        >
          {marker.glyph}
        </span>
      ))}
      {children}
    </span>
  );
}

function isInteractiveEditorTarget(target: EventTarget): boolean {
  return (
    target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    target instanceof HTMLSelectElement
  );
}

function markSemanticDataCell(
  node: HTMLElement | null,
  state: GridResolvedSemanticState,
  rangeSelected: boolean,
) {
  const cell = node?.closest<HTMLElement>('[role="gridcell"]');
  if (cell === undefined || cell === null) return null;
  for (const className of [...cell.classList]) {
    if (
      className.startsWith("cartulary-grid-cell-state-") ||
      className.startsWith("cartulary-grid-cell-is-")
    ) {
      cell.classList.remove(className);
    }
  }
  cell.classList.add(
    ...gridSemanticStateClassNames("cell", state)
      .split(" ")
      .filter((className) => className !== "cartulary-grid-cell"),
  );
  cell.dataset.gridPrimaryState = state.primary;
  setOptionalAriaBoolean(cell, "aria-busy", state.stateIds.includes("pending"));
  setOptionalAriaBoolean(
    cell,
    "aria-invalid",
    state.stateIds.includes("invalid"),
  );
  cell.setAttribute(
    "aria-readonly",
    String(state.stateIds.includes("read-only")),
  );
  setOptionalAriaBoolean(cell, "aria-selected", rangeSelected);
  if (state.description === undefined) cell.removeAttribute("aria-description");
  else cell.setAttribute("aria-description", state.description);
  return cell;
}

function setOptionalAriaBoolean(
  element: HTMLElement,
  attribute: "aria-busy" | "aria-invalid" | "aria-selected",
  value: boolean,
) {
  if (value) element.setAttribute(attribute, "true");
  else element.removeAttribute(attribute);
}

function markSemanticHeaderCell(
  node: HTMLElement | null,
  options: {
    readonly accessibleLabel: string;
    readonly fieldKey: string;
    readonly testId: string | undefined;
    readonly title?: string | undefined;
  },
) {
  const header = node?.closest<HTMLElement>('[role="columnheader"]');
  if (header === undefined || header === null) return;
  header.dataset.gridFieldKey = options.fieldKey;
  if (options.testId !== undefined) header.dataset.testid = options.testId;
  header.setAttribute("aria-label", options.accessibleLabel);
  if (options.title !== undefined) header.title = options.title;
}

function draftMarker<Row>(
  row: GridDraftRow<Row>,
  columnKey: string,
  firstColumnKey: string | undefined,
) {
  const isMarkerCell = columnKey === firstColumnKey;
  return {
    "data-cartulary-grid-draft-row": isMarkerCell ? "true" : undefined,
    ref: (node: HTMLSpanElement | null) => {
      const semanticRow = node?.closest<HTMLElement>('[role="row"]');
      if (semanticRow === undefined || semanticRow === null) return;
      semanticRow.classList.add("cartulary-grid-draft-row");
      semanticRow.dataset.cartularyGridDraftRow = "true";
      semanticRow.dataset.gridPrimaryState = "draft";
      if (row.testId === undefined) {
        semanticRow.removeAttribute("data-testid");
      } else {
        semanticRow.dataset.testid = row.testId;
      }
    },
  };
}

function stopVendorNavigationForInteractiveContent(
  event: KeyboardEvent<HTMLSpanElement>,
  allowSemanticRangeAndExit = false,
) {
  if (
    event.target !== event.currentTarget &&
    ((event.key.startsWith("Arrow") &&
      !(allowSemanticRangeAndExit && event.shiftKey)) ||
      event.key === "Enter" ||
      (!allowSemanticRangeAndExit && event.key === "Tab"))
  ) {
    event.stopPropagation();
  }
}

function alignmentClass(align: GridColumn<unknown>["align"]) {
  return align === undefined ? undefined : `cartulary-grid-cell-${align}`;
}

function dataCellClass(
  align: GridColumn<unknown>["align"],
  state: GridResolvedSemanticState,
) {
  return [
    "cartulary-grid-data-cell",
    alignmentClass(align),
    gridSemanticStateClassNames("cell", state),
  ]
    .filter((value) => value !== undefined)
    .join(" ");
}

function markSemanticGutterCell(
  node: HTMLSpanElement | null,
  testId: string | undefined,
) {
  const semanticCell = node?.closest<HTMLElement>('[role="gridcell"]');
  if (semanticCell === undefined || semanticCell === null) return;
  semanticCell.dataset.gridFieldKey = gridRowGutterColumnKey;
  if (testId === undefined) {
    semanticCell.removeAttribute("data-testid");
  } else {
    semanticCell.dataset.testid = testId;
  }
}
