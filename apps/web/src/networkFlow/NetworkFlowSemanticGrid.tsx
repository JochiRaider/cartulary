import {
  type GridCellAnchor,
  type GridCellRange,
  type GridDataRow,
  type GridDataState,
  type GridHandle,
  type GridSortEntry,
  GridViewport,
  SemanticDataGrid,
} from "@cartulary/grid-adapter";
import {
  networkAnalysisColumnActionTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type RefObject,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowRow,
  NetworkFlowSort,
  NetworkFlowTable,
} from "../services/networkFlowContractAdapter";
import {
  compileNetworkFlowColumns,
  localizedNetworkFlowDiagnosticMessage,
  type NetworkFlowGridSchemaId,
  networkFlowClipboardValue,
  networkFlowColumnLabel,
  networkFlowContributorsForGrid,
  networkFlowDiagnosticsForGrid,
  networkFlowGridSurface,
  networkFlowPresentationColumns,
  networkFlowRowsForGrid,
} from "./networkFlowPresentation";
import { useNetworkFlowGridLayout } from "./useNetworkFlowGridLayout";
import type { NetworkFlowQueryLoadState } from "./useNetworkFlowPagedQuery";

export function NetworkFlowAcceptedGrid({
  filtered,
  loadGenerationKey = 0,
  loadState,
  onResetQuery,
  onRetry,
  onSortChange,
  resetKey,
  rows,
  sort,
  onSelectionChange,
}: {
  readonly filtered: boolean;
  readonly loadGenerationKey?: string | number | undefined;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly onResetQuery: () => void;
  readonly onRetry: () => void;
  readonly onSortChange: (sort: readonly NetworkFlowSort[]) => void;
  readonly resetKey: string;
  readonly rows: readonly NetworkFlowRow[];
  readonly sort: readonly NetworkFlowSort[];
  readonly onSelectionChange: (
    activeAnchor: GridCellAnchor | null,
    cellRange: GridCellRange | null,
  ) => void;
}) {
  const layout = useNetworkFlowGridLayout("network_flow.accepted_rows.v1");
  const columns = useMemo(
    () =>
      compileNetworkFlowColumns<NetworkFlowRow>({
        gridSchemaId: "network_flow.accepted_rows.v1",
        orderedVisibleFieldKeys: layout.orderedVisibleFieldKeys,
        widths: layout.columnWidths,
      }),
    [layout.columnWidths, layout.orderedVisibleFieldKeys],
  );
  const dataRows = useStableGridProjection(rows, networkFlowRowsForGrid);
  const sortableFieldKeys = useMemo(
    () =>
      new Set(
        networkFlowPresentationColumns("network_flow.accepted_rows.v1")
          .filter((column) => column.sortable)
          .map((column) => column.field_key),
      ),
    [],
  );
  return (
    <NetworkFlowGridFrame
      columnsControl={layout}
      dataState={networkFlowGridDataState({
        filtered,
        itemCount: rows.length,
        loadGenerationKey,
        loadState,
        onResetQuery,
        onRetry,
        surfaceLabel: "accepted Network Flow rows",
      })}
      gridSchemaId="network_flow.accepted_rows.v1"
      onSelectionChange={onSelectionChange}
      resetKey={resetKey}
      rows={rows}
    >
      {({
        activeAnchor,
        cellRange,
        gridRef,
        onActiveAnchorChange,
        onCellRangeChange,
      }) => (
        <SemanticDataGrid
          ref={gridRef}
          accessibleLabel="Accepted Network Flow rows"
          activeRowIdentity={activeAnchor?.rowIdentity ?? null}
          cellRange={cellRange}
          columns={columns}
          columnWidths={layout.columnWidths}
          dataRows={dataRows}
          dataState={networkFlowGridDataState({
            filtered,
            itemCount: rows.length,
            loadGenerationKey,
            loadState,
            onResetQuery,
            onRetry,
            surfaceLabel: "accepted Network Flow rows",
          })}
          density="default"
          fillViewportInline
          interactionMode={{
            kind: "read_only",
            label:
              "Network Flow rows are read-only. Range selection and copy are available.",
          }}
          onActiveCellChange={onActiveAnchorChange}
          onCellRangeChange={onCellRangeChange}
          onColumnReorder={layout.onColumnReorder}
          onColumnWidthChange={layout.onColumnWidthChange}
          onSortChange={(nextSort) =>
            onSortChange(
              nextSort.flatMap((entry) =>
                sortableFieldKeys.has(entry.fieldKey)
                  ? [
                      {
                        field_key:
                          entry.fieldKey as NetworkFlowSort["field_key"],
                        direction: entry.direction,
                      },
                    ]
                  : [],
              ),
            )
          }
          rowGutter={{ label: "Source row", minWidth: 56, width: 64 }}
          sort={sort.map(
            (entry): GridSortEntry => ({
              fieldKey: entry.field_key,
              direction: entry.direction,
            }),
          )}
          surface={networkFlowGridSurface("network_flow.accepted_rows.v1")}
        />
      )}
    </NetworkFlowGridFrame>
  );
}

export function NetworkFlowRejectedGrid({
  diagnostics,
  filtered,
  loadGenerationKey = 0,
  loadState,
  onResetQuery,
  onRetry,
  resetKey,
}: {
  readonly diagnostics: readonly NetworkFlowDiagnostic[];
  readonly filtered: boolean;
  readonly loadGenerationKey?: string | number | undefined;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly onResetQuery: () => void;
  readonly onRetry: () => void;
  readonly resetKey: string;
}) {
  const layout = useNetworkFlowGridLayout("network_flow.rejected_rows.v1");
  const columns = useMemo(
    () =>
      compileNetworkFlowColumns<NetworkFlowDiagnostic>({
        gridSchemaId: "network_flow.rejected_rows.v1",
        orderedVisibleFieldKeys: layout.orderedVisibleFieldKeys,
        widths: layout.columnWidths,
      }),
    [layout.columnWidths, layout.orderedVisibleFieldKeys],
  );
  const dataRows = useStableGridProjection(
    diagnostics,
    networkFlowDiagnosticsForGrid,
  );
  return (
    <NetworkFlowGridFrame
      columnsControl={layout}
      dataState={networkFlowGridDataState({
        filtered,
        itemCount: diagnostics.length,
        loadGenerationKey,
        loadState,
        onResetQuery,
        onRetry,
        surfaceLabel: "rejected-row diagnostics",
      })}
      gridSchemaId="network_flow.rejected_rows.v1"
      resetKey={resetKey}
      rows={diagnostics}
    >
      {({
        activeAnchor,
        cellRange,
        gridRef,
        onActiveAnchorChange,
        onCellRangeChange,
      }) => (
        <SemanticDataGrid
          ref={gridRef}
          accessibleLabel="Rejected Network Flow diagnostics"
          activeRowIdentity={activeAnchor?.rowIdentity ?? null}
          cellRange={cellRange}
          columns={columns}
          columnWidths={layout.columnWidths}
          dataRows={dataRows}
          dataState={networkFlowGridDataState({
            filtered,
            itemCount: diagnostics.length,
            loadGenerationKey,
            loadState,
            onResetQuery,
            onRetry,
            surfaceLabel: "rejected-row diagnostics",
          })}
          density="default"
          fillViewportInline
          interactionMode={{
            kind: "read_only",
            label:
              "Rejected-row diagnostics are read-only. Range selection and copy are available.",
          }}
          onActiveCellChange={onActiveAnchorChange}
          onCellRangeChange={onCellRangeChange}
          onColumnReorder={layout.onColumnReorder}
          onColumnWidthChange={layout.onColumnWidthChange}
          surface={networkFlowGridSurface("network_flow.rejected_rows.v1")}
        />
      )}
    </NetworkFlowGridFrame>
  );
}

export function NetworkFlowContributorGrid({
  contributors,
  loadGenerationKey = 0,
  loadState,
  onRetry,
  tables,
}: {
  readonly contributors: readonly NetworkFlowContributor[];
  readonly loadGenerationKey?: string | number | undefined;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly onRetry: () => void;
  readonly tables: readonly NetworkFlowTable[];
}) {
  const layout = useNetworkFlowGridLayout("network_flow.graph_contributors.v1");
  const columns = useMemo(
    () =>
      compileNetworkFlowColumns<NetworkFlowRow>({
        gridSchemaId: "network_flow.graph_contributors.v1",
        orderedVisibleFieldKeys: layout.orderedVisibleFieldKeys,
        widths: layout.columnWidths,
      }),
    [layout.columnWidths, layout.orderedVisibleFieldKeys],
  );
  const tableNames = useMemo(
    () =>
      new Map(
        tables.map((table) => [
          table.network_flow_table_id,
          table.display_name,
        ]),
      ),
    [tables],
  );
  const dataRows = useStableGridProjection(
    contributors,
    networkFlowContributorsForGrid,
  );
  const grouping = useMemo(
    () => ({
      fieldKey: "network_flow_table_id",
      label: "Table",
      getValue: (row: NetworkFlowRow) => row.network_flow_table_id,
      formatLabel: (value: boolean | number | string | null) =>
        typeof value === "string"
          ? (tableNames.get(value) ?? "Unavailable table")
          : null,
    }),
    [tableNames],
  );
  return (
    <div style={gridFrameStyle}>
      <ColumnLayoutControls control={layout} />
      <GridViewport
        blockSizing="fill"
        style={gridViewportStyle}
        testId={networkAnalysisTestId("contributor-grid")}
      >
        <SemanticDataGrid
          accessibleLabel="Network Flow graph contributors"
          columns={columns}
          columnWidths={layout.columnWidths}
          dataRows={dataRows}
          dataState={networkFlowGridDataState({
            filtered: false,
            itemCount: contributors.length,
            loadGenerationKey,
            loadState,
            onResetQuery: () => undefined,
            onRetry,
            surfaceLabel: "graph contributors",
          })}
          density="default"
          fillViewportInline
          grouping={grouping}
          interactionMode={{
            kind: "read_only",
            label:
              "Network Flow contributors are read-only and preserve server order within workspace table groups.",
          }}
          onColumnReorder={layout.onColumnReorder}
          onColumnWidthChange={layout.onColumnWidthChange}
          rowGutter={{ label: "Source row", minWidth: 56, width: 64 }}
          surface={networkFlowGridSurface("network_flow.graph_contributors.v1")}
        />
      </GridViewport>
    </div>
  );
}

function NetworkFlowGridFrame<Row extends object>({
  children,
  columnsControl,
  dataState,
  gridSchemaId,
  onSelectionChange,
  resetKey,
  rows,
}: {
  readonly children: (state: {
    readonly activeAnchor: GridCellAnchor | null;
    readonly cellRange: GridCellRange | null;
    readonly gridRef: RefObject<GridHandle | null>;
    readonly onActiveAnchorChange: (anchor: GridCellAnchor | null) => void;
    readonly onCellRangeChange: (range: GridCellRange | null) => void;
  }) => React.ReactNode;
  readonly columnsControl: ReturnType<typeof useNetworkFlowGridLayout>;
  readonly dataState: GridDataState;
  readonly gridSchemaId: Exclude<
    NetworkFlowGridSchemaId,
    "network_flow.graph_contributors.v1"
  >;
  readonly onSelectionChange?:
    | ((
        activeAnchor: GridCellAnchor | null,
        cellRange: GridCellRange | null,
      ) => void)
    | undefined;
  readonly resetKey: string;
  readonly rows: readonly Row[];
}) {
  const [activeAnchor, setActiveAnchor] = useState<GridCellAnchor | null>(null);
  const [cellRange, setCellRange] = useState<GridCellRange | null>(null);
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const gridRef = useRef<GridHandle | null>(null);
  const focusRestorationRef = useRef(false);
  const lastRowIndexRef = useRef(0);
  const restoreGridAnchor = useCallback((anchor: GridCellAnchor) => {
    focusRestorationRef.current = true;
    queueMicrotask(() => {
      if (gridRef.current?.focusAnchor(anchor) !== true) {
        gridRef.current?.focusRoot();
      }
      window.setTimeout(() => {
        focusRestorationRef.current = false;
      }, 0);
    });
  }, []);
  const rowResourceIds = useMemo(
    () =>
      rows.flatMap((row) => {
        const resourceId = resourceIdForRow(row, gridSchemaId);
        return resourceId === null ? [] : [resourceId];
      }),
    [gridSchemaId, rows],
  );
  const rowResourceKey = rowResourceIds.join("\u0000");
  const priorGridStateRef = useRef({ resetKey, rowResourceKey });
  const handleActiveAnchorChange = useCallback(
    (anchor: GridCellAnchor | null) => {
      setActiveAnchor(anchor);
      if (anchor !== null) {
        const resourceId =
          anchor.rowIdentity.kind === "extension_resource"
            ? anchor.rowIdentity.resourceId
            : null;
        const rowIndex =
          resourceId === null ? -1 : rowResourceIds.indexOf(resourceId);
        if (rowIndex >= 0) lastRowIndexRef.current = rowIndex;
        if (!focusRestorationRef.current) setInspectorOpen(true);
      }
    },
    [rowResourceIds],
  );
  useEffect(() => {
    const previous = priorGridStateRef.current;
    const resetChanged = previous.resetKey !== resetKey;
    const rowsChanged = previous.rowResourceKey !== rowResourceKey;
    priorGridStateRef.current = { resetKey, rowResourceKey };
    if (resetChanged) {
      const hadSemanticSelection =
        activeAnchor !== null || cellRange !== null || inspectorOpen;
      setActiveAnchor(null);
      setCellRange(null);
      setInspectorOpen(false);
      if (hadSemanticSelection) focusGridRoot(gridRef);
      return;
    }
    if (!rowsChanged || activeAnchor === null) return;
    const activeResourceId =
      activeAnchor.rowIdentity.kind === "extension_resource"
        ? activeAnchor.rowIdentity.resourceId
        : null;
    const exactIndex =
      activeResourceId === null ? -1 : rowResourceIds.indexOf(activeResourceId);
    if (exactIndex >= 0) {
      lastRowIndexRef.current = exactIndex;
      restoreGridAnchor(activeAnchor);
      return;
    }
    const nearestResourceId =
      rowResourceIds[
        Math.min(lastRowIndexRef.current, rowResourceIds.length - 1)
      ];
    if (
      nearestResourceId === undefined ||
      activeAnchor.rowIdentity.kind !== "extension_resource"
    ) {
      setActiveAnchor(null);
      setCellRange(null);
      setInspectorOpen(false);
      focusGridRoot(gridRef);
      return;
    }
    const nextAnchor: GridCellAnchor = {
      ...activeAnchor,
      rowIdentity: {
        ...activeAnchor.rowIdentity,
        resourceId: nearestResourceId,
      },
    };
    lastRowIndexRef.current = rowResourceIds.indexOf(nearestResourceId);
    setActiveAnchor(nextAnchor);
    setCellRange({ end: nextAnchor, start: nextAnchor });
    restoreGridAnchor(nextAnchor);
  }, [
    activeAnchor,
    cellRange,
    inspectorOpen,
    resetKey,
    restoreGridAnchor,
    rowResourceIds,
    rowResourceKey,
  ]);
  const visibleFieldKey = columnsControl.orderedVisibleFieldKeys.join("\u0000");
  useEffect(() => {
    void visibleFieldKey;
    if (
      activeAnchor === null ||
      columnsControl.orderedVisibleFieldKeys.includes(activeAnchor.fieldKey)
    ) {
      return;
    }
    setActiveAnchor(null);
    setCellRange(null);
    setInspectorOpen(false);
    focusGridRoot(gridRef);
  }, [activeAnchor, columnsControl.orderedVisibleFieldKeys, visibleFieldKey]);
  useEffect(() => {
    onSelectionChange?.(activeAnchor, cellRange);
  }, [activeAnchor, cellRange, onSelectionChange]);
  const activeRow = activeAnchor
    ? (rows.find(
        (row) =>
          resourceIdForRow(row, gridSchemaId) ===
          (activeAnchor.rowIdentity.kind === "extension_resource"
            ? activeAnchor.rowIdentity.resourceId
            : null),
      ) ?? null)
    : null;

  return (
    <div style={gridFrameStyle}>
      <ColumnLayoutControls control={columnsControl} />
      <div style={gridAndInspectorStyle}>
        <GridViewport
          blockSizing="fill"
          style={gridViewportStyle}
          testId={networkAnalysisTestId(
            gridSchemaId === "network_flow.accepted_rows.v1"
              ? "accepted-grid"
              : "rejected-grid",
          )}
        >
          {children({
            activeAnchor,
            cellRange,
            gridRef,
            onActiveAnchorChange: handleActiveAnchorChange,
            onCellRangeChange: setCellRange,
          })}
        </GridViewport>
        {activeRow === null ||
        activeAnchor === null ||
        !inspectorOpen ? null : (
          <NetworkFlowInspector
            anchor={activeAnchor}
            gridSchemaId={gridSchemaId}
            row={activeRow}
            onClose={() => {
              setInspectorOpen(false);
              restoreGridAnchor(activeAnchor);
            }}
          />
        )}
      </div>
      <span aria-live="polite" style={visuallyHiddenStyle}>
        {dataState.kind === "ready"
          ? "Network Flow grid ready"
          : dataState.kind}
      </span>
    </div>
  );
}

function ColumnLayoutControls({
  control,
}: {
  readonly control: ReturnType<typeof useNetworkFlowGridLayout>;
}) {
  const [announcement, setAnnouncement] = useState("");
  return (
    <div style={layoutToolbarStyle}>
      <details data-testid={networkAnalysisTestId("column-menu")}>
        <summary>Columns</summary>
        <div style={columnMenuStyle}>
          {control.allColumns.map((column) => {
            const checked = control.visibleFieldKeys.has(column.field_key);
            const visibleIndex = control.orderedVisibleFieldKeys.indexOf(
              column.field_key,
            );
            return (
              <div key={column.field_key} style={columnControlRowStyle}>
                <label style={columnToggleStyle}>
                  <input
                    checked={checked}
                    data-testid={networkAnalysisColumnActionTestId(
                      column.field_key,
                      "toggle",
                    )}
                    disabled={checked && control.visibleFieldKeys.size === 1}
                    type="checkbox"
                    onChange={(event) => {
                      control.setColumnVisible(
                        column.field_key,
                        event.currentTarget.checked,
                      );
                      setAnnouncement(
                        `${networkFlowColumnLabel(column.label_key)} column ${event.currentTarget.checked ? "shown" : "hidden"}.`,
                      );
                    }}
                  />
                  {networkFlowColumnLabel(column.label_key)}
                </label>
                {checked ? (
                  <span style={columnMoveActionsStyle}>
                    <button
                      aria-label={`Move ${networkFlowColumnLabel(column.label_key)} earlier`}
                      data-testid={networkAnalysisColumnActionTestId(
                        column.field_key,
                        "move-earlier",
                      )}
                      disabled={visibleIndex <= 0}
                      type="button"
                      onClick={() => {
                        const target =
                          control.orderedVisibleFieldKeys[visibleIndex - 1];
                        if (target === undefined) return;
                        control.onColumnReorder(column.field_key, target);
                        setAnnouncement(
                          `${networkFlowColumnLabel(column.label_key)} column moved earlier.`,
                        );
                      }}
                    >
                      Earlier
                    </button>
                    <button
                      aria-label={`Move ${networkFlowColumnLabel(column.label_key)} later`}
                      data-testid={networkAnalysisColumnActionTestId(
                        column.field_key,
                        "move-later",
                      )}
                      disabled={
                        visibleIndex < 0 ||
                        visibleIndex ===
                          control.orderedVisibleFieldKeys.length - 1
                      }
                      type="button"
                      onClick={() => {
                        const target =
                          control.orderedVisibleFieldKeys[visibleIndex + 1];
                        if (target === undefined) return;
                        control.onColumnReorder(column.field_key, target);
                        setAnnouncement(
                          `${networkFlowColumnLabel(column.label_key)} column moved later.`,
                        );
                      }}
                    >
                      Later
                    </button>
                  </span>
                ) : null}
              </div>
            );
          })}
        </div>
      </details>
      <button
        data-testid={networkAnalysisTestId("layout-reset")}
        type="button"
        onClick={() => {
          control.reset();
          setAnnouncement("Network Flow column layout reset to defaults.");
        }}
      >
        Reset layout
      </button>
      <span aria-live="polite" style={visuallyHiddenStyle}>
        {announcement}
      </span>
    </div>
  );
}

function NetworkFlowInspector<Row extends object>({
  anchor,
  gridSchemaId,
  onClose,
  row,
}: {
  readonly anchor: GridCellAnchor;
  readonly gridSchemaId:
    | "network_flow.accepted_rows.v1"
    | "network_flow.rejected_rows.v1";
  readonly onClose: () => void;
  readonly row: Row;
}) {
  const metadata = networkFlowPresentationColumns(gridSchemaId);
  return (
    <aside
      aria-label="Network Flow cell inspector"
      data-testid={networkAnalysisTestId("inspector")}
      style={inspectorStyle}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          onClose();
        }
      }}
    >
      <div style={inspectorHeaderStyle}>
        <div>
          <strong>Cell inspector</strong>
          <div style={mutedStyle}>
            {networkFlowColumnLabel(
              metadata.find((column) => column.field_key === anchor.fieldKey)
                ?.label_key ?? anchor.fieldKey,
            )}
          </div>
        </div>
        <button
          aria-label="Close cell inspector"
          data-testid={networkAnalysisTestId("inspector-close")}
          type="button"
          onClick={onClose}
        >
          Close
        </button>
      </div>
      <dl style={inspectorListStyle}>
        {metadata.map((column) => {
          const value =
            column.renderer_kind === "diagnostic_message"
              ? localizedNetworkFlowDiagnosticMessage(
                  row as NetworkFlowDiagnostic,
                )
              : networkFlowClipboardValue(row, column);
          return (
            <div key={column.field_key} style={inspectorEntryStyle}>
              <dt style={mutedStyle}>
                {networkFlowColumnLabel(column.label_key)}
              </dt>
              <dd style={inspectorValueStyle}>{value === "" ? "—" : value}</dd>
            </div>
          );
        })}
      </dl>
    </aside>
  );
}

function focusGridRoot(gridRef: RefObject<GridHandle | null>) {
  queueMicrotask(() => gridRef.current?.focusRoot());
}

function useStableGridProjection<Owner, Row>(
  owners: readonly Owner[],
  project: (
    owners: readonly Owner[],
    previous: readonly GridDataRow<Row>[],
  ) => readonly GridDataRow<Row>[],
): readonly GridDataRow<Row>[] {
  const previousRef = useRef<readonly GridDataRow<Row>[]>([]);
  const projected = useMemo(
    () => project(owners, previousRef.current),
    [owners, project],
  );
  useEffect(() => {
    previousRef.current = projected;
  }, [projected]);
  return projected;
}

function networkFlowGridDataState(options: {
  readonly filtered: boolean;
  readonly itemCount: number;
  readonly loadGenerationKey: string | number;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly onResetQuery: () => void;
  readonly onRetry: () => void;
  readonly surfaceLabel: string;
}): GridDataState {
  if (options.loadState === "loading") {
    return {
      generationKey: options.loadGenerationKey,
      kind: "initial_loading",
      surfaceLabel: options.surfaceLabel,
    };
  }
  if (options.loadState === "refreshing") {
    return { kind: "refreshing", surfaceLabel: options.surfaceLabel };
  }
  if (options.loadState === "error") {
    return options.itemCount > 0
      ? {
          kind: "stale_error",
          message: "Refresh failed. Previously loaded rows may be stale.",
          action: { label: "Retry", onInvoke: options.onRetry },
        }
      : {
          kind: "unavailable",
          message: "Network Flow rows are unavailable.",
          action: { label: "Retry", onInvoke: options.onRetry },
        };
  }
  if (options.itemCount === 0) {
    return options.filtered
      ? {
          kind: "filtered_empty",
          action: { label: "Clear filters", onInvoke: options.onResetQuery },
        }
      : { kind: "empty", message: `No ${options.surfaceLabel}.` };
  }
  return { kind: "ready" };
}

function resourceIdForRow(
  row: object,
  gridSchemaId:
    | "network_flow.accepted_rows.v1"
    | "network_flow.rejected_rows.v1",
): string | null {
  const key =
    gridSchemaId === "network_flow.accepted_rows.v1"
      ? "network_flow_row_id"
      : "diagnostic_id";
  const value = (row as Record<string, unknown>)[key];
  return typeof value === "string" ? value : null;
}

const gridFrameStyle = {
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;

const gridAndInspectorStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;

const gridViewportStyle = {
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;

const layoutToolbarStyle = {
  alignItems: "center",
  background: "var(--ct-colors-surface-1)",
  borderBlockEnd: "var(--ct-border-hairline)",
  display: "flex",
  gap: "var(--ct-spacing-sm)",
  justifyContent: "flex-end",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-md)",
} satisfies CSSProperties;

const columnMenuStyle = {
  background: "var(--ct-colors-surface-1)",
  border: "var(--ct-border-hairline)",
  boxShadow: "var(--ct-shadow-popover)",
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  maxBlockSize: "20rem",
  overflow: "auto",
  padding: "var(--ct-spacing-sm)",
  position: "absolute",
  zIndex: 5,
} satisfies CSSProperties;

const columnToggleStyle = {
  alignItems: "center",
  display: "flex",
  gap: "var(--ct-spacing-xs)",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const columnControlRowStyle = {
  alignItems: "center",
  display: "flex",
  gap: "var(--ct-spacing-sm)",
  justifyContent: "space-between",
} satisfies CSSProperties;

const columnMoveActionsStyle = {
  display: "inline-flex",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const inspectorStyle = {
  background: "var(--ct-colors-surface-1)",
  borderInlineStart: "var(--ct-border-hairline)",
  inlineSize: "min(24rem, 35vw)",
  overflow: "auto",
  padding: "var(--ct-spacing-md)",
} satisfies CSSProperties;

const inspectorHeaderStyle = {
  alignItems: "flex-start",
  display: "flex",
  gap: "var(--ct-spacing-md)",
  justifyContent: "space-between",
} satisfies CSSProperties;

const inspectorListStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  margin: "var(--ct-spacing-md) 0 0",
} satisfies CSSProperties;

const inspectorEntryStyle = {
  borderBlockEnd: "var(--ct-border-hairline)",
  display: "grid",
  gap: "0.125rem",
  paddingBlockEnd: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const inspectorValueStyle = {
  margin: 0,
  overflowWrap: "anywhere",
  whiteSpace: "pre-wrap",
} satisfies CSSProperties;

const mutedStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.75rem",
} satisfies CSSProperties;

const visuallyHiddenStyle = {
  blockSize: 1,
  clip: "rect(0 0 0 0)",
  clipPath: "inset(50%)",
  inlineSize: 1,
  overflow: "hidden",
  position: "absolute",
  whiteSpace: "nowrap",
} satisfies CSSProperties;
