import {
  type FocusEvent,
  type KeyboardEvent,
  type ReactNode,
  useCallback,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { Column } from "react-data-grid";
import type {
  GridActionsColumn,
  GridCellAnchor,
  GridColumn,
  GridDataRow,
  GridDraftRow,
  GridEditCommitOutcome,
  GridEditorActivation,
  GridEditorAdapter,
  GridEditorFocusTarget,
  GridRowGutter,
  GridSemanticStateInput,
  GridSurfaceIdentity,
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
  readonly onToggleRecord: (row: GridDataRow<Row>, shiftKey: boolean) => void;
  readonly isRecordSelectable: (row: GridDataRow<Row>) => boolean;
};

type CompileGridColumnsInput<Row> = {
  readonly actionsColumn: GridActionsColumn<Row> | undefined;
  readonly bulkSelection: GridCompiledBulkSelection<Row> | undefined;
  readonly clearEditorSeed: () => void;
  readonly columns: readonly GridColumn<Row>[];
  readonly cellStateFor: (
    row: GridDataRow<Row>,
    column: GridColumn<Row>,
  ) => GridSemanticStateInput;
  readonly editable: boolean;
  readonly isCellRangeSelected: (
    row: GridDataRow<Row>,
    column: GridColumn<Row>,
  ) => boolean;
  readonly readEditorSeed: (
    target: Parameters<GridEditorAdapter<Row>["commit"]>[0]["target"],
  ) => {
    readonly activation: GridEditorActivation;
    readonly hasValue: boolean;
    readonly value: unknown;
  } | null;
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
  readonly draftFocusTargetRef: (
    fieldKey: string,
  ) => (element: GridEditorFocusTarget | null) => void;
  readonly registerEditorSession: (
    session: {
      readonly cancel: () => void;
      readonly focus: () => void;
      readonly requestCommit: () => Promise<boolean>;
      readonly target: Parameters<
        GridEditorAdapter<Row>["commit"]
      >[0]["target"];
    } | null,
  ) => void;
  readonly rowGutter: GridRowGutter | undefined;
  readonly surface: GridSurfaceIdentity;
  readonly onPasteCellContent:
    | ((
        row: GridDataRow<Row>,
        fieldKey: string,
        clipboardText: string,
      ) => boolean)
    | undefined;
};

function gridMutationTarget<Row>(
  row: GridDataRow<Row>,
  fieldKey: string,
  surface: GridSurfaceIdentity,
) {
  if (
    surface.kind !== "view_schema" ||
    row.rowIdentity.kind !== "core_record" ||
    row.mutationIdentity === undefined
  ) {
    return null;
  }
  return {
    fieldKey,
    mutationIdentity: row.mutationIdentity,
    rowIdentity: row.rowIdentity,
    surface,
  };
}

export function compileGridColumns<Row>({
  actionsColumn,
  bulkSelection,
  clearEditorSeed,
  cellStateFor,
  columns,
  readEditorSeed,
  editable,
  draftFocusTargetRef,
  isCellRangeSelected,
  onEditorKeyboardAction,
  onPasteCellContent,
  registerEditorSession,
  registerSemanticCell,
  rowGutter,
  surface,
}: CompileGridColumnsInput<Row>): readonly Column<
  GridDataRow<Row>,
  GridDraftRow<Row>
>[] {
  const firstColumnKey =
    rowGutter !== undefined
      ? gridRowGutterColumnKey
      : (columns[0]?.fieldKey ??
        (actionsColumn === undefined ? undefined : gridActionsColumnKey));
  const compiled: Array<Column<GridDataRow<Row>, GridDraftRow<Row>>> = [];

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
        <span
          className="cartulary-grid-header-content cartulary-grid-selection-content"
          data-grid-field-key={gridSelectionColumnKey}
          ref={(node) =>
            markSemanticHeaderCell(node, {
              accessibleLabel: "Select records",
              fieldKey: gridSelectionColumnKey,
              testId: undefined,
            })
          }
        >
          <input
            aria-label="Select all records on this page"
            checked={bulkSelection.allSelected}
            disabled={bulkSelection.selectableRecordCount === 0}
            ref={(node) => {
              if (node !== null) {
                node.indeterminate = bulkSelection.partiallySelected;
              }
            }}
            readOnly
            type="checkbox"
            onClick={(event) => {
              event.stopPropagation();
              bulkSelection.onSelectAll();
            }}
          />
        </span>
      ),
      renderCell: ({ row }) =>
        row.rowIdentity.kind === "core_record" &&
        bulkSelection.isRecordSelectable(row) ? (
          <span
            className="cartulary-grid-selection-content"
            data-grid-field-key={gridSelectionColumnKey}
          >
            <input
              aria-label={`Select record ${row.rowIdentity.recordId}`}
              checked={bulkSelection.selectedRecordIds.has(
                row.rowIdentity.recordId,
              )}
              data-grid-field-key={gridSelectionColumnKey}
              readOnly
              type="checkbox"
              onClick={(event) => {
                event.stopPropagation();
                bulkSelection.onToggleRecord(row, event.shiftKey);
              }}
            />
          </span>
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
          className="cartulary-grid-header-content cartulary-grid-gutter-content"
          data-grid-field-key={gridRowGutterColumnKey}
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
          className="cartulary-grid-gutter-content"
          data-grid-field-key={gridRowGutterColumnKey}
          ref={(node) => markSemanticGutterCell(node, row.gutterTestId)}
        >
          {row.gutterContent ?? row.gutterLabel ?? ""}
        </span>
      ),
      renderSummaryCell: ({ row }) => (
        <span
          {...draftMarker(row, gridRowGutterColumnKey, firstColumnKey)}
          className="cartulary-grid-gutter-content"
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
        column.editor !== undefined &&
        column.valueKind !== "collection",
      key: column.fieldKey,
      headerCellClass: "cartulary-grid-header-cell",
      minWidth: column.minWidth,
      name: column.label,
      renderHeaderCell: ({ sortDirection }) => (
        <span
          className="cartulary-grid-header-content"
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
          rowIdentity: row.rowIdentity,
          surface,
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
        column.editor !== undefined &&
        column.valueKind !== "collection"
          ? ({ onClose, row }) => {
              const target = gridMutationTarget(row, column.fieldKey, surface);
              const editorSeed =
                target === null ? null : readEditorSeed(target);
              return target === null ? null : (
                <SemanticGridEditor
                  adapter={column.editor as GridEditorAdapter<Row>}
                  baseState={cellStateFor(row, column)}
                  editorSeed={editorSeed}
                  fieldLabel={column.label}
                  registerEditorSession={registerEditorSession}
                  registerSemanticCell={registerSemanticCell}
                  row={row.data}
                  target={target}
                  onClose={(accepted, shouldFocusCell) => {
                    // Clear the semantic session before RDG schedules the
                    // editor unmount. A pointer transition in that interval
                    // must not negotiate with an editor that has already
                    // accepted cancellation or commit.
                    registerEditorSession(null);
                    clearEditorSeed();
                    onClose(accepted, shouldFocusCell);
                  }}
                  onKeyboardAction={onEditorKeyboardAction}
                />
              );
            }
          : undefined,
      editorOptions: {
        closeOnExternalRowChange: false,
        commitOnOutsideClick: false,
      },
      renderSummaryCell: ({ row }) => (
        // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns this private bottom-draft cell; the wrapper only prevents duplicate vendor navigation after a nested control handles the key.
        <span
          {...draftMarker(row, column.fieldKey, firstColumnKey)}
          className="cartulary-grid-cell-content"
          data-grid-field-key={column.fieldKey}
          onKeyDown={stopVendorNavigationForInteractiveContent}
          role="presentation"
        >
          {surface.kind === "view_schema"
            ? (column.renderDraftCell?.({
                fieldKey: column.fieldKey,
                focusTargetRef: draftFocusTargetRef(column.fieldKey),
                row: row.data,
                surface,
              }) ?? null)
            : null}
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
          className="cartulary-grid-header-content cartulary-grid-actions-content"
          data-grid-field-key={gridActionsColumnKey}
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
  registerEditorSession,
  registerSemanticCell,
  row,
  target,
  onClose,
  onKeyboardAction,
}: {
  readonly adapter: GridEditorAdapter<Row>;
  readonly baseState: GridSemanticStateInput;
  readonly editorSeed: {
    readonly activation: GridEditorActivation;
    readonly hasValue: boolean;
    readonly value: unknown;
  } | null;
  readonly fieldLabel: string;
  readonly registerEditorSession: (
    session: {
      readonly cancel: () => void;
      readonly focus: () => void;
      readonly requestCommit: () => Promise<boolean>;
      readonly target: Parameters<
        GridEditorAdapter<Row>["commit"]
      >[0]["target"];
    } | null,
  ) => void;
  readonly registerSemanticCell: (
    anchor: GridCellAnchor,
    cell: HTMLElement | null,
    token: object,
  ) => void;
  readonly row: Row;
  readonly target: Parameters<GridEditorAdapter<Row>["commit"]>[0]["target"];
  readonly onClose: (accepted: boolean, shouldFocusCell: boolean) => void;
  readonly onKeyboardAction: (
    target: Parameters<GridEditorAdapter<Row>["commit"]>[0]["target"],
    action:
      | { readonly kind: "exit"; readonly backwards: boolean }
      | { readonly kind: "move"; readonly rowDelta: -1 | 1 },
  ) => void;
}) {
  const activation = editorSeed?.activation ?? {
    initialSelection: "all" as const,
    source: "enter" as const,
  };
  const [draftValue, setDraftValue] = useState(() =>
    editorSeed?.hasValue === true
      ? editorSeed.value
      : adapter.initialDraftValue(row),
  );
  const [outcome, setOutcome] = useState<GridEditCommitOutcome | null>(null);
  const [pending, setPending] = useState(false);
  const commitPromisesRef = useRef(
    new Map<string, Promise<GridEditCommitOutcome>>(),
  );
  const latestCommitSequenceRef = useRef(0);
  const focusTargetRef = useRef<GridEditorFocusTarget | null>(null);
  const registerFocusTarget = useCallback(
    (element: GridEditorFocusTarget | null) => {
      focusTargetRef.current = element;
    },
    [],
  );
  const focusEditor = useCallback(() => {
    const element = focusTargetRef.current;
    if (element === null) return;
    element.focus({ preventScroll: true });
  }, []);
  useLayoutEffect(() => {
    const element = focusTargetRef.current;
    if (element === null) return;
    element.focus({ preventScroll: true });
    if (
      element instanceof HTMLInputElement ||
      element instanceof HTMLTextAreaElement
    ) {
      if (activation.initialSelection === "all") {
        element.select();
      } else if (activation.initialSelection === "end") {
        const end = element.value.length;
        element.setSelectionRange(end, end);
      }
    }
  }, [activation.initialSelection]);
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
  const commitDraft = useCallback(
    (
      draftValueOverride?: unknown,
      shouldFocusCell = true,
    ): Promise<GridEditCommitOutcome> => {
      const requestedDraft =
        draftValueOverride === undefined ? draftValue : draftValueOverride;
      const draftKey = gridEditorDraftKey(requestedDraft);
      const duplicate = commitPromisesRef.current.get(draftKey);
      if (duplicate !== undefined) return duplicate;
      const sequence = latestCommitSequenceRef.current + 1;
      latestCommitSequenceRef.current = sequence;
      setPending(true);
      setOutcome(null);
      const request = (async (): Promise<GridEditCommitOutcome> => {
        let next: GridEditCommitOutcome;
        try {
          next = await adapter.commit({
            draftValue: requestedDraft,
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
        commitPromisesRef.current.delete(draftKey);
        setPending(commitPromisesRef.current.size > 0);
        const isLatest = latestCommitSequenceRef.current === sequence;
        if (isLatest) setOutcome(next);
        if (
          next.kind === "accepted" &&
          isLatest &&
          commitPromisesRef.current.size === 0
        ) {
          onClose(true, shouldFocusCell);
        }
        return next;
      })();
      commitPromisesRef.current.set(draftKey, request);
      return request;
    },
    [adapter, draftValue, onClose, row, target],
  );
  useLayoutEffect(() => {
    registerEditorSession({
      cancel: () => onClose(false, true),
      focus: focusEditor,
      requestCommit: async () =>
        (await commitDraft(undefined, false)).kind === "accepted",
      target,
    });
    return () => registerEditorSession(null);
  }, [commitDraft, focusEditor, onClose, registerEditorSession, target]);
  const commit = async (draftValueOverride?: unknown) => {
    await commitDraft(draftValueOverride);
  };
  const handleKeyboardAction = async (
    action:
      | { readonly kind: "exit"; readonly backwards: boolean }
      | { readonly kind: "move"; readonly rowDelta: -1 | 1 },
    draftValueOverride?: unknown,
  ) => {
    const next = await commitDraft(draftValueOverride);
    if (next?.kind === "accepted") {
      queueMicrotask(() => onKeyboardAction(target, action));
    }
  };
  return (
    <SemanticGridCellContent
      anchor={target}
      editing
      fieldKey={target.fieldKey}
      registerSemanticCell={registerSemanticCell}
      semanticState={semanticState}
      onKeyDownCapture={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          onClose(false, true);
          return;
        }
        if (event.key === "Tab") {
          event.preventDefault();
          event.stopPropagation();
          void handleKeyboardAction(
            {
              backwards: event.shiftKey,
              kind: "exit",
            },
            editorControlValue(event.target),
          );
          return;
        }
        if (event.key === "Enter") {
          event.preventDefault();
          event.stopPropagation();
          void handleKeyboardAction(
            {
              kind: "move",
              rowDelta: event.shiftKey ? -1 : 1,
            },
            editorControlValue(event.target),
          );
        }
      }}
      onBlurCapture={(event) => {
        if (event.currentTarget.contains(event.relatedTarget as Node | null)) {
          return;
        }
        if (
          event.relatedTarget instanceof Element &&
          event.relatedTarget.closest(
            '[data-grid-editor-external-action="true"]',
          ) !== null
        ) {
          return;
        }
        void commitDraft(undefined, false).then((next) => {
          if (next.kind !== "accepted") focusEditor();
        });
      }}
    >
      {adapter.renderEditor({
        activation,
        cancel: () => onClose(false, true),
        commit,
        draftValue,
        focusTargetRef: registerFocusTarget,
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

function gridEditorDraftKey(value: unknown): string {
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

function SemanticGridCellContent({
  anchor,
  children,
  editing = false,
  fieldKey,
  onBlurCapture,
  onKeyDownCapture,
  onPaste,
  rangeSelected = false,
  registerSemanticCell,
  semanticState,
}: {
  readonly anchor: GridCellAnchor;
  readonly children: ReactNode;
  readonly editing?: boolean | undefined;
  readonly fieldKey: string;
  readonly onBlurCapture?:
    | ((event: FocusEvent<HTMLSpanElement>) => void)
    | undefined;
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
      data-grid-editing={editing ? "true" : undefined}
      data-grid-field-key={fieldKey}
      data-grid-primary-state={semanticState.primary}
      ref={(node) => {
        const cell = markSemanticDataCell(node, semanticState, rangeSelected);
        registerSemanticCell(anchor, cell, registrationToken);
      }}
      onKeyDown={(event) =>
        stopVendorNavigationForInteractiveContent(event, true)
      }
      onBlurCapture={onBlurCapture}
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

function editorControlValue(target: EventTarget): unknown {
  if (target instanceof HTMLInputElement) {
    return target.type === "checkbox" ? target.checked : target.value;
  }
  if (
    target instanceof HTMLSelectElement ||
    target instanceof HTMLTextAreaElement
  ) {
    return target.value;
  }
  return undefined;
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
