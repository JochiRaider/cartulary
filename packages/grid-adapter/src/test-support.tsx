// biome-ignore-all lint/a11y/noNoninteractiveElementToInteractiveRole: The test grid intentionally preserves RDG role attributes for selector compatibility.
// biome-ignore-all lint/a11y/noRedundantRoles: Explicit roles keep workbook tests independent of native accessibility-role inference.
// biome-ignore-all lint/a11y/useFocusableInteractive: This test renderer mirrors RDG's query surface, not a production interaction model.
import { gridScrollportClassName } from "@cartulary/ui-contracts";
import {
  type FocusEvent,
  Fragment,
  type KeyboardEvent,
  type ReactNode,
  type Ref,
  useCallback,
  useImperativeHandle,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

import {
  assertGridRows,
  buildGridPresentationRows,
  type GridColumn,
  type GridDataRow,
  type GridEditCommitOutcome,
  type GridEditorAdapter,
  type GridEditorFocusTarget,
  type GridHandle,
  type GridRowIdentity,
  type GridSemanticStateInput,
  type GridViewportProps,
  gridClipboardDimensions,
  gridRowIdentitiesEqual,
  gridSurfaceIdentitiesEqual,
  gridUnassignedGroupLabel,
  resolveGridPasteTargets,
  type SemanticDataGridProps,
} from "./core";
import {
  gridSemanticStateClassNames,
  mergeGridSemanticState,
  resolveGridSemanticState,
} from "./semanticState";

function coreRecordId<Row>(row: GridDataRow<Row>): string | null {
  return row.rowIdentity.kind === "core_record"
    ? row.rowIdentity.recordId
    : null;
}

export {
  assertGridRows,
  buildGridPresentationRows,
  formatGridClipboardTSV,
  type GridActionsColumn,
  type GridBlockSizing,
  type GridCellAnchor,
  type GridCellCopyIntent,
  type GridCellMutationIntent,
  type GridCellRange,
  type GridCellSelection,
  type GridCellStateContext,
  type GridCellStateInput,
  type GridCellTarget,
  type GridColumn,
  type GridCoreRecordBulkSelection,
  type GridDataRow,
  type GridDataState,
  type GridDataStateAction,
  type GridDraftRow,
  type GridEditCommitIntent,
  type GridEditCommitOutcome,
  type GridEditorAdapter,
  type GridEditorRenderContext,
  type GridExpandedCellRange,
  type GridFillIntent,
  type GridGroupingDescriptor,
  type GridGroupingScalar,
  type GridHandle,
  type GridInteractionMode,
  type GridMutationIdentity,
  type GridNavigationIntent,
  type GridNavigationKey,
  type GridRowGutter,
  type GridRowIdentity,
  type GridRowStateInput,
  type GridSemanticStateInput,
  type GridSortDirection,
  type GridSortEntry,
  type GridStateValidation,
  type GridSurfaceIdentity,
  type GridViewportProps,
  gridClipboardDimensions,
  navigateGridCellAnchor,
  parseGridClipboardTable,
  resolveGridCellAnchor,
  resolveGridCellRange,
  type SemanticDataGridProps,
} from "./core";
export { setGridVirtualizationDisabledForDiagnostics } from "./virtualizationDiagnostics";

export function GridViewport({
  children,
  className,
  style,
  testId,
}: GridViewportProps) {
  return (
    <div className={className} data-testid={testId} style={style}>
      {children}
    </div>
  );
}

export function SemanticDataGrid<Row>({
  accessibleLabel,
  activeRowIdentity = null,
  allowPasteCreateRows = false,
  actionsColumn,
  coreRecordBulkSelection,
  columns,
  dataState = { kind: "ready" },
  draftRow,
  fillViewportInline = false,
  getCellState,
  getRowState,
  grouping = null,
  interactionMode,
  onPasteCell,
  onSelectRow,
  onSortChange,
  dataRows,
  rowGutter,
  sort = [],
  surface,
  ref,
}: SemanticDataGridProps<Row> & { readonly ref?: Ref<GridHandle> }) {
  const effectiveInteractionMode =
    interactionMode ??
    (surface.kind === "extension_grid"
      ? { kind: "read_only" as const, label: "Read only" }
      : { kind: "editable" as const });
  const editable = effectiveInteractionMode.kind === "editable";
  const effectiveBulkSelection = editable ? coreRecordBulkSelection : undefined;
  const effectiveDraftRow = editable ? draftRow : undefined;
  const selectionAnchorRecordId = useRef<string | null>(null);
  const [activeEditor, setActiveEditor] = useState<{
    readonly fieldKey: string;
    readonly rowIdentity: GridRowIdentity;
  } | null>(null);
  useImperativeHandle(
    ref,
    () => ({
      activateEdit: (anchor) => {
        if (!gridSurfaceIdentitiesEqual(surface, anchor.surface)) return false;
        setActiveEditor({
          fieldKey: anchor.fieldKey,
          rowIdentity: anchor.rowIdentity,
        });
        return true;
      },
      cancelEdit: (anchor) => {
        if (
          activeEditor === null ||
          activeEditor.fieldKey !== anchor.fieldKey ||
          !gridRowIdentitiesEqual(
            activeEditor.rowIdentity,
            anchor.rowIdentity,
          ) ||
          !gridSurfaceIdentitiesEqual(surface, anchor.surface)
        ) {
          return false;
        }
        setActiveEditor(null);
        return true;
      },
      focusAnchor: () => false,
      focusRoot: () => false,
      getScrollElement: () => null,
      scrollToAnchor: () => false,
    }),
    [activeEditor, surface],
  );
  assertGridRows(dataRows);

  const selectableRows =
    effectiveBulkSelection === undefined
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
            row.rowIdentity.kind === "core_record" &&
            effectiveBulkSelection.isRecordSelectable?.(row) !== false,
        );
  const selectableIds = selectableRows.map((row) => row.rowIdentity.recordId);
  const selectedOnPage = selectableIds.filter((recordId) =>
    effectiveBulkSelection?.selectedRecordIds.has(recordId),
  );
  const rowStateFor = (row: GridDataRow<Row>): GridSemanticStateInput =>
    mergeGridSemanticState(getRowState?.(row), {
      bulkSelected:
        row.rowIdentity.kind === "core_record" &&
        effectiveBulkSelection?.selectedRecordIds.has(
          row.rowIdentity.recordId,
        ) === true,
      inspectorActive:
        activeRowIdentity !== null &&
        gridRowIdentitiesEqual(row.rowIdentity, activeRowIdentity),
      readOnlyOrDerived: !editable,
      saved: true,
      stale: dataState.kind === "stale_error",
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

  const renderedRows = buildGridPresentationRows({
    grouping,
    rows: dataRows,
  });
  const totalColumnCount =
    columns.length +
    (effectiveBulkSelection === undefined ? 0 : 1) +
    (rowGutter === undefined ? 0 : 1) +
    (actionsColumn === undefined ? 0 : 1);

  return (
    <>
      <TestGridStatePresentation
        dataState={dataState}
        interactionMode={effectiveInteractionMode}
      />
      <table
        aria-label={accessibleLabel}
        aria-busy={
          dataState.kind === "initial_loading" ||
          dataState.kind === "refreshing"
        }
        aria-readonly={!editable}
        className={gridScrollportClassName()}
        role="grid"
        style={fillViewportInline ? { minWidth: 0, width: "100%" } : undefined}
      >
        <thead>
          <tr role="row">
            {effectiveBulkSelection === undefined ? null : (
              <th role="columnheader" scope="col">
                <input
                  aria-label="Select all records on this page"
                  checked={
                    selectableIds.length > 0 &&
                    selectedOnPage.length === selectableIds.length
                  }
                  disabled={selectableIds.length === 0}
                  ref={(node) => {
                    if (node !== null) {
                      node.indeterminate =
                        selectedOnPage.length > 0 &&
                        selectedOnPage.length < selectableIds.length;
                    }
                  }}
                  readOnly
                  type="checkbox"
                  onClick={() => {
                    selectionAnchorRecordId.current = null;
                    effectiveBulkSelection.onSelectedRecordIdsChange(
                      selectedOnPage.length === selectableIds.length
                        ? new Set()
                        : new Set(selectableIds),
                    );
                  }}
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
            {columns.map((column) => {
              const canToggleSort =
                onSortChange !== undefined &&
                column.sortableFieldKey !== null &&
                column.sortableFieldKey !== undefined &&
                !column.sortDisabled;
              const sortState = sort.find(
                (entry) =>
                  entry.fieldKey ===
                  (column.sortableFieldKey ?? column.fieldKey),
              );
              return (
                <th key={column.fieldKey} role="columnheader" scope="col">
                  <button
                    data-grid-field-key={column.fieldKey}
                    data-testid={column.headerTestId}
                    disabled={!canToggleSort}
                    title={column.sortDisabledReason ?? undefined}
                    type="button"
                    onClick={(event) => {
                      const fieldKey =
                        column.sortableFieldKey ?? column.fieldKey;
                      const currentIndex = sort.findIndex(
                        (entry) => entry.fieldKey === fieldKey,
                      );
                      const current = sort[currentIndex];
                      const nextEntry =
                        current === undefined
                          ? { fieldKey, direction: "asc" as const }
                          : current.direction === "asc"
                            ? { fieldKey, direction: "desc" as const }
                            : null;
                      const additive = event.ctrlKey || event.metaKey;
                      if (!additive) {
                        onSortChange?.(nextEntry === null ? [] : [nextEntry]);
                        return;
                      }
                      const next = sort.filter(
                        (entry) => entry.fieldKey !== fieldKey,
                      );
                      onSortChange?.(
                        nextEntry === null ? next : [...next, nextEntry],
                      );
                    }}
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
            })}
            {actionsColumn === undefined ? null : (
              <th role="columnheader" scope="col">
                <span>{actionsColumn.label}</span>
              </th>
            )}
          </tr>
        </thead>
        <tbody>
          {renderedRows.length === 0 &&
          effectiveDraftRow === undefined ? null : (
            <Fragment>
              {renderedRows.map((row) =>
                row.kind === "group" ? (
                  <tr key={row.key} role="row">
                    <td colSpan={totalColumnCount} role="gridcell">
                      <strong data-testid={row.testId}>
                        {row.groupLabel ?? gridUnassignedGroupLabel}
                      </strong>
                    </td>
                  </tr>
                ) : (
                  <tr
                    {...testSemanticAttributes(
                      "row",
                      rowStateFor(row.gridRow),
                      "data row",
                    )}
                    data-grid-row-identity-kind={row.gridRow.rowIdentity.kind}
                    data-grid-record-id={coreRecordId(row.gridRow) ?? undefined}
                    data-testid={row.gridRow.testId}
                    key={row.key}
                    role="row"
                    tabIndex={onSelectRow === undefined ? undefined : 0}
                    onClick={() => {
                      onSelectRow?.(row.gridRow.rowIdentity);
                    }}
                    onKeyDown={(event: KeyboardEvent<HTMLTableRowElement>) => {
                      if (event.key === "Enter" || event.key === " ") {
                        onSelectRow?.(row.gridRow.rowIdentity);
                      }
                    }}
                  >
                    {effectiveBulkSelection === undefined ? null : (
                      <td role="gridcell">
                        {row.gridRow.rowIdentity.kind !== "core_record" ||
                        effectiveBulkSelection.isRecordSelectable?.(
                          row.gridRow,
                        ) === false ? null : (
                          <input
                            aria-label={`Select record ${row.gridRow.rowIdentity.recordId}`}
                            checked={effectiveBulkSelection.selectedRecordIds.has(
                              row.gridRow.rowIdentity.recordId,
                            )}
                            readOnly
                            type="checkbox"
                            onClick={(event) => {
                              event.stopPropagation();
                              const next = new Set(
                                effectiveBulkSelection.selectedRecordIds,
                              );
                              const anchorIndex = selectableRows.findIndex(
                                (candidate) =>
                                  candidate.rowIdentity.recordId ===
                                  selectionAnchorRecordId.current,
                              );
                              const rowIndex = selectableRows.findIndex(
                                (candidate) =>
                                  candidate.rowIdentity.recordId ===
                                  (coreRecordId(row.gridRow) ?? ""),
                              );
                              if (
                                event.shiftKey &&
                                anchorIndex >= 0 &&
                                rowIndex >= 0
                              ) {
                                const start = Math.min(anchorIndex, rowIndex);
                                const end = Math.max(anchorIndex, rowIndex);
                                for (const candidate of selectableRows.slice(
                                  start,
                                  end + 1,
                                )) {
                                  next.add(candidate.rowIdentity.recordId);
                                }
                              } else if (
                                next.has(coreRecordId(row.gridRow) ?? "")
                              ) {
                                next.delete(coreRecordId(row.gridRow) ?? "");
                              } else {
                                next.add(coreRecordId(row.gridRow) ?? "");
                              }
                              selectionAnchorRecordId.current = coreRecordId(
                                row.gridRow,
                              );
                              effectiveBulkSelection.onSelectedRecordIdsChange(
                                next,
                              );
                            }}
                          />
                        )}
                      </td>
                    )}
                    {rowGutter === undefined ? null : (
                      <th
                        data-grid-field-key="__cartulary_row_gutter__"
                        data-testid={row.gridRow.gutterTestId}
                        scope="row"
                      >
                        {row.gridRow.gutterContent ??
                          row.gridRow.gutterLabel ??
                          ""}
                      </th>
                    )}
                    {columns.map((column) => {
                      const semanticState = resolveGridSemanticState(
                        cellStateFor(row.gridRow, column),
                        column.label,
                      );
                      return (
                        <td
                          {...testSemanticAttributes(
                            "cell",
                            cellStateFor(row.gridRow, column),
                            column.label,
                          )}
                          data-grid-field-key={column.fieldKey}
                          key={column.fieldKey}
                          role="gridcell"
                          onClick={(event) => {
                            if (
                              !editable ||
                              column.contractWritable !== true ||
                              column.editor === undefined ||
                              column.valueKind === "collection" ||
                              (event.target instanceof Element &&
                                event.target.closest(
                                  "button, a, input, select, textarea, [role='button'], [data-grid-prevent-cell-edit='true']",
                                ) !== null)
                            ) {
                              return;
                            }
                            setActiveEditor({
                              fieldKey: column.fieldKey,
                              rowIdentity: row.gridRow.rowIdentity,
                            });
                          }}
                          onKeyDown={(event) => {
                            if (
                              event.key !== "Enter" ||
                              !editable ||
                              column.contractWritable !== true ||
                              column.editor === undefined ||
                              column.valueKind === "collection"
                            ) {
                              return;
                            }
                            setActiveEditor({
                              fieldKey: column.fieldKey,
                              rowIdentity: row.gridRow.rowIdentity,
                            });
                          }}
                          onPaste={(event) => {
                            if (!editable || onPasteCell === undefined) return;
                            if (
                              event.target instanceof HTMLInputElement ||
                              event.target instanceof HTMLTextAreaElement ||
                              event.target instanceof HTMLSelectElement
                            ) {
                              return;
                            }
                            const clipboardText =
                              event.clipboardData?.getData("text/plain") ?? "";
                            const dimensions =
                              gridClipboardDimensions(clipboardText);
                            const targetResolution = resolveGridPasteTargets({
                              allowCreateRows:
                                allowPasteCreateRows && grouping === null,
                              columns,
                              current: {
                                fieldKey: column.fieldKey,
                                rowIdentity: row.gridRow.rowIdentity,
                                surface,
                              },
                              pastedColumnCount: dimensions.columnCount,
                              pastedRowCount: dimensions.rowCount,
                              presentationRows: renderedRows,
                            });
                            if (
                              targetResolution === null ||
                              !targetResolution.columns.every((fieldKey) =>
                                columns.some(
                                  (candidate) =>
                                    candidate.fieldKey === fieldKey &&
                                    candidate.contractWritable === true &&
                                    candidate.editor !== undefined,
                                ),
                              )
                            ) {
                              return;
                            }
                            event.preventDefault();
                            if (row.gridRow.mutationIdentity === undefined) {
                              return;
                            }
                            onPasteCell({
                              clipboardText,
                              target: {
                                fieldKey: column.fieldKey,
                                mutationIdentity: row.gridRow.mutationIdentity,
                                rowIdentity: row.gridRow.rowIdentity,
                                surface,
                              },
                              targetResolution,
                            });
                          }}
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
                          {activeEditor?.fieldKey === column.fieldKey &&
                          gridRowIdentitiesEqual(
                            activeEditor.rowIdentity,
                            row.gridRow.rowIdentity,
                          ) &&
                          surface.kind === "view_schema" &&
                          row.gridRow.mutationIdentity !== undefined &&
                          column.editor !== undefined ? (
                            <TestGridEditor
                              adapter={column.editor}
                              row={row.gridRow.data}
                              target={{
                                fieldKey: column.fieldKey,
                                mutationIdentity: row.gridRow.mutationIdentity,
                                rowIdentity: row.gridRow.rowIdentity,
                                surface,
                              }}
                              onClose={() => setActiveEditor(null)}
                            />
                          ) : (
                            column.renderCell({
                              anchor: {
                                fieldKey: column.fieldKey,
                                rowIdentity: row.gridRow.rowIdentity,
                                surface,
                              },
                              row: row.gridRow.data,
                            })
                          )}
                        </td>
                      );
                    })}
                    {actionsColumn === undefined ? null : (
                      <td role="gridcell">
                        {actionsColumn.renderCell(row.gridRow)}
                      </td>
                    )}
                  </tr>
                ),
              )}
              {effectiveDraftRow === undefined ? null : (
                <tr
                  data-cartulary-grid-draft-row="true"
                  data-testid={effectiveDraftRow.testId}
                  role="row"
                >
                  {effectiveBulkSelection === undefined ? null : (
                    <td role="gridcell" />
                  )}
                  {rowGutter === undefined ? null : (
                    <th
                      data-grid-field-key="__cartulary_row_gutter__"
                      scope="row"
                    >
                      {effectiveDraftRow.gutterContent ??
                        effectiveDraftRow.gutterLabel ??
                        ""}
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
                            row: effectiveDraftRow.data,
                            surface,
                          }) ?? null)
                        : null}
                    </td>
                  ))}
                  {actionsColumn === undefined ? null : (
                    <td role="gridcell">
                      {actionsColumn.renderDraftCell?.(effectiveDraftRow)}
                    </td>
                  )}
                </tr>
              )}
            </Fragment>
          )}
        </tbody>
      </table>
    </>
  );
}

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

function TestGridStatePresentation({
  dataState,
  interactionMode,
}: {
  readonly dataState: NonNullable<SemanticDataGridProps<unknown>["dataState"]>;
  readonly interactionMode: NonNullable<
    SemanticDataGridProps<unknown>["interactionMode"]
  >;
}) {
  let message: string | null = null;
  if (dataState.kind === "initial_loading") {
    message = `Loading ${dataState.surfaceLabel}…`;
  } else if (dataState.kind === "refreshing") {
    message = `Refreshing ${dataState.surfaceLabel}…`;
  } else if (dataState.kind === "empty") {
    message = dataState.message;
  } else if (dataState.kind === "filtered_empty") {
    message = "No rows match the current filters.";
  } else if (dataState.kind === "stale_error") {
    message = `${dataState.message} Previously loaded rows may be stale.`;
  } else if (dataState.kind === "unavailable") {
    message = dataState.message;
  } else if (dataState.kind === "permission_denied") {
    message =
      dataState.message ?? "You no longer have access to this workbook.";
  }
  const action =
    "action" in dataState && dataState.action !== undefined
      ? dataState.action
      : undefined;
  return (
    <>
      {message === null ? null : (
        <div data-grid-data-state={dataState.kind} role="status">
          {message}
          {action === undefined ? null : (
            <button type="button" onClick={action.onInvoke}>
              {action.label}
            </button>
          )}
        </div>
      )}
      {interactionMode.kind === "read_only" ? (
        <div data-grid-interaction-mode="read_only" role="status">
          {interactionMode.label}
        </div>
      ) : null}
    </>
  );
}
