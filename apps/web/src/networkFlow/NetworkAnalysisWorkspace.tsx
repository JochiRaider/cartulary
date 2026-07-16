import {
  networkAnalysisEdgeTestId,
  networkAnalysisRowTestId,
  networkAnalysisTableTabTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import {
  Link2,
  Network,
  PanelRightOpen,
  RefreshCw,
  Table2,
  Upload,
  X,
} from "lucide-react";
import { type CSSProperties, useCallback, useRef, useState } from "react";
import { IncidentCollaborationBoundary } from "../collaboration/IncidentCollaborationSession";
import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import { NetworkFlowMappingModal } from "./NetworkFlowMappingModal";
import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowEdgeAnnotation,
  NetworkFlowGraphResult,
  NetworkFlowRow,
  NetworkFlowTable,
} from "./networkFlowClient";
import { networkAnalysisSheetRef } from "./networkFlowClient";
import {
  type NetworkFlowWorkspaceError,
  networkFlowErrorMessage,
} from "./networkFlowErrors";
import { useNetworkFlowCollaborationController } from "./useNetworkFlowCollaborationController";
import { useNetworkFlowGraphController } from "./useNetworkFlowGraphController";
import { useNetworkFlowImportController } from "./useNetworkFlowImportController";
import { useNetworkFlowIndicatorLinkController } from "./useNetworkFlowIndicatorLinkController";
import { useNetworkFlowRejectedRowsController } from "./useNetworkFlowRejectedRowsController";
import { useNetworkFlowRowsController } from "./useNetworkFlowRowsController";
import { useNetworkFlowTableController } from "./useNetworkFlowTableController";

type NetworkAnalysisMode = "rows" | "rejected" | "graph";

export type NetworkAnalysisWorkspaceProps = {
  readonly apiBase?: string | undefined;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
};

const activeTableScopeLabel = networkAnalysisSheetRef();

function NetworkAnalysisWorkspaceContent({
  apiBase,
  currentIncidentRole,
  incidentId,
  onIncidentAccessLost,
}: NetworkAnalysisWorkspaceProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [mode, setMode] = useState<NetworkAnalysisMode>("rows");
  const [message, setMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] =
    useState<NetworkFlowWorkspaceError | null>(null);
  const canImport =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const canLink =
    currentIncidentRole === "editor" || currentIncidentRole === "admin";
  const tableController = useNetworkFlowTableController({
    apiBase,
    incidentId,
    onIncidentAccessLost,
  });
  const rowsController = useNetworkFlowRowsController({
    activeTableId: tableController.activeTableId,
    apiBase,
    enabled: mode === "rows",
    incidentId,
    onError: setErrorMessage,
    onIncidentAccessLost,
  });
  const rejectedRowsController = useNetworkFlowRejectedRowsController({
    activeTableId: tableController.activeTableId,
    apiBase,
    enabled: mode === "rejected",
    incidentId,
    onError: setErrorMessage,
    onIncidentAccessLost,
  });
  const graphController = useNetworkFlowGraphController({
    apiBase,
    enabled: mode === "graph",
    incidentId,
    onError: setErrorMessage,
    onIncidentAccessLost,
    tableIds: tableController.tableIds,
  });
  const clearResources = useCallback(() => {
    rowsController.clearRows();
    rejectedRowsController.clearDiagnostics();
    graphController.clearGraph();
  }, [graphController, rejectedRowsController, rowsController]);
  useNetworkFlowCollaborationController({
    apiBase,
    clearResources,
    dispatchTableAction: tableController.dispatch,
    incidentId,
    loadTables: tableController.loadTables,
    onMessage: setMessage,
  });
  const importController = useNetworkFlowImportController({
    apiBase,
    canImport,
    incidentId,
    onError: setErrorMessage,
    onImported: tableController.loadTables,
    onMessage: setMessage,
  });
  const indicatorLinkController = useNetworkFlowIndicatorLinkController({
    apiBase,
    firstContributor: graphController.firstContributor,
    graph: graphController.graph,
    incidentId,
    onError: setErrorMessage,
    onMessage: setMessage,
    selectedEdge: graphController.selectedEdge,
  });
  const empty =
    tableController.loadState === "ready" &&
    tableController.tables.length === 0;
  const visibleError = errorMessage ?? tableController.error;

  return (
    <section
      aria-label="Network Analysis"
      data-extension-profile-id={activeTableScopeLabel.extension_profile_id}
      data-testid={networkAnalysisTestId("workspace")}
      data-workspace-key={activeTableScopeLabel.workspace_key}
      style={workspaceStyle}
    >
      <div style={viewBarStyle}>
        <div aria-label="Network Flow tables" role="tablist" style={tabsStyle}>
          {tableController.tables.map((table, index) => {
            const selected =
              table.network_flow_table_id === tableController.activeTableId;
            return (
              <button
                key={table.network_flow_table_id}
                aria-selected={selected}
                data-testid={networkAnalysisTableTabTestId(
                  table.network_flow_table_id,
                )}
                role="tab"
                style={{
                  ...innerTabStyle,
                  ...(selected ? innerTabActiveStyle : null),
                }}
                type="button"
                onClick={() => {
                  tableController.dispatch({
                    type: "select_table",
                    tableId: table.network_flow_table_id,
                  });
                  setMode("rows");
                }}
              >
                <span>{table.display_name}</span>
                <small style={tabCountStyle}>{index + 1}</small>
              </button>
            );
          })}
        </div>
        <div style={viewActionsStyle}>
          <button
            data-testid={networkAnalysisTestId("refresh")}
            style={iconButtonStyle}
            title="Refresh"
            type="button"
            onClick={() => {
              void tableController.loadTables();
            }}
          >
            <RefreshCw aria-hidden="true" size={16} />
          </button>
          <input
            ref={fileInputRef}
            accept=".csv,text/csv"
            data-testid={networkAnalysisTestId("import-input")}
            hidden
            type="file"
            onChange={importController.handleImportChange}
          />
          {canImport ? (
            <button
              data-testid={networkAnalysisTestId("import-trigger")}
              disabled={importController.importing}
              style={commandButtonStyle}
              type="button"
              onClick={() => fileInputRef.current?.click()}
            >
              <Upload aria-hidden="true" size={16} />
              Import NetFlow CSV
            </button>
          ) : null}
        </div>
      </div>

      <div style={modeBarStyle}>
        <button
          aria-pressed={mode === "rows"}
          data-testid={networkAnalysisTestId("mode-rows")}
          disabled={tableController.activeTable === null}
          style={{
            ...modeButtonStyle,
            ...(mode === "rows" ? modeButtonActiveStyle : null),
          }}
          type="button"
          onClick={() => setMode("rows")}
        >
          <Table2 aria-hidden="true" size={15} />
          Rows
        </button>
        <button
          aria-pressed={mode === "rejected"}
          data-testid={networkAnalysisTestId("mode-rejected")}
          disabled={tableController.activeTable === null}
          style={{
            ...modeButtonStyle,
            ...(mode === "rejected" ? modeButtonActiveStyle : null),
          }}
          type="button"
          onClick={() => setMode("rejected")}
        >
          Rejected
        </button>
        <button
          aria-pressed={mode === "graph"}
          data-testid={networkAnalysisTestId("mode-graph")}
          disabled={tableController.tables.length === 0}
          style={{
            ...modeButtonStyle,
            ...(mode === "graph" ? modeButtonActiveStyle : null),
          }}
          type="button"
          onClick={() => setMode("graph")}
        >
          <Network aria-hidden="true" size={15} />
          Graph
        </button>
      </div>

      <div style={workAreaStyle}>
        {empty ? (
          <EmptyNetworkAnalysisState
            canImport={canImport}
            importing={importController.importing}
            onImport={() => fileInputRef.current?.click()}
          />
        ) : mode === "graph" ? (
          <GraphPanel
            canLink={canLink}
            contributors={graphController.contributors}
            firstContributor={graphController.firstContributor}
            graph={graphController.graph}
            selectedEdge={graphController.selectedEdge}
            onCloseDrawer={() => {
              graphController.setSelectedEdgeId(null);
            }}
            onLinkDst={() =>
              void indicatorLinkController.linkEdge("network_flow.dst_ip")
            }
            onLinkSrc={() =>
              void indicatorLinkController.linkEdge("network_flow.src_ip")
            }
            onSelectEdge={graphController.setSelectedEdgeId}
          />
        ) : mode === "rejected" ? (
          <RejectedRowsPanel
            activeTable={tableController.activeTable}
            canNext={rejectedRowsController.canNext}
            canPrevious={rejectedRowsController.canPrevious}
            diagnostics={rejectedRowsController.diagnostics}
            loading={
              rejectedRowsController.loadState === "loading" ||
              rejectedRowsController.loadState === "refreshing"
            }
            notice={rejectedRowsController.notice}
            pageNumber={rejectedRowsController.pageNumber}
            onNext={rejectedRowsController.nextPage}
            onPrevious={rejectedRowsController.previousPage}
          />
        ) : (
          <RowsPanel
            activeTable={tableController.activeTable}
            canNext={rowsController.canNext}
            canPrevious={rowsController.canPrevious}
            loading={
              rowsController.loadState === "loading" ||
              rowsController.loadState === "refreshing"
            }
            notice={rowsController.notice}
            pageNumber={rowsController.pageNumber}
            rows={rowsController.rows}
            onNext={rowsController.nextPage}
            onPrevious={rowsController.previousPage}
          />
        )}
      </div>

      <div style={statusStripStyle}>
        <span>
          {tableController.loadState === "loading"
            ? "Loading"
            : `${tableController.tables.length} active table${
                tableController.tables.length === 1 ? "" : "s"
              }`}
        </span>
        {message ? (
          <span data-testid={networkAnalysisTestId("stale-state")}>
            {message}
          </span>
        ) : null}
        {visibleError ? (
          <span style={errorTextStyle}>
            {networkFlowErrorMessage(visibleError)}
          </span>
        ) : null}
      </div>
      {importController.mappingOpen &&
      importController.discovery !== null &&
      importController.draft !== null ? (
        <NetworkFlowMappingModal
          canApply={importController.canApply}
          discovery={importController.discovery}
          draft={importController.draft}
          preview={importController.preview}
          stage={importController.stage}
          onApply={() => {
            void importController.apply();
          }}
          onCancel={importController.reset}
          onDraftChange={importController.updateDraft}
          onPreview={() => {
            void importController.requestPreview();
          }}
        />
      ) : null}
    </section>
  );
}

export function NetworkAnalysisWorkspace(props: NetworkAnalysisWorkspaceProps) {
  return (
    <IncidentCollaborationBoundary
      apiBase={props.apiBase}
      incidentId={props.incidentId}
      initialPresence={{
        sheet_ref: networkAnalysisSheetRef(),
        mode: "viewing",
      }}
    >
      <NetworkAnalysisWorkspaceContent {...props} />
    </IncidentCollaborationBoundary>
  );
}

function EmptyNetworkAnalysisState({
  canImport,
  importing,
  onImport,
}: {
  readonly canImport: boolean;
  readonly importing: boolean;
  readonly onImport: () => void;
}) {
  return (
    <section
      aria-label="Empty Network Analysis workspace"
      style={emptyStateStyle}
    >
      {canImport ? (
        <button
          disabled={importing}
          style={commandButtonStyle}
          type="button"
          onClick={onImport}
        >
          <Upload aria-hidden="true" size={16} />
          Import NetFlow CSV
        </button>
      ) : (
        <span style={mutedTextStyle}>No active Network Flow tables.</span>
      )}
    </section>
  );
}

function RowsPanel({
  activeTable,
  canNext,
  canPrevious,
  loading,
  notice,
  onNext,
  onPrevious,
  pageNumber,
  rows,
}: {
  readonly activeTable: NetworkFlowTable | null;
  readonly canNext: boolean;
  readonly canPrevious: boolean;
  readonly loading: boolean;
  readonly notice: string | null;
  readonly onNext: () => void;
  readonly onPrevious: () => void;
  readonly pageNumber: number;
  readonly rows: readonly NetworkFlowRow[];
}) {
  return (
    <section
      aria-label="Network Flow table rows"
      data-testid={networkAnalysisTestId("table-panel")}
      style={panelGridStyle}
    >
      <PanelHeader table={activeTable} />
      <div style={tableScrollStyle}>
        <table style={dataTableStyle}>
          <thead>
            <tr>
              <th style={thStyle}>Row</th>
              <th style={thStyle}>Start</th>
              <th style={thStyle}>Source</th>
              <th style={thStyle}>Destination</th>
              <th style={thStyle}>Protocol</th>
              <th style={thStyle}>Bytes</th>
              <th style={thStyle}>Packets</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr
                key={row.network_flow_row_id}
                data-testid={networkAnalysisRowTestId(row.network_flow_row_id)}
              >
                <td style={tdMonoStyle}>{row.source_row_number}</td>
                <td style={tdStyle}>
                  {shortTimestamp(row["network_flow.flow_start_utc"])}
                </td>
                <td style={tdMonoStyle}>{endpointLabel(row, "src")}</td>
                <td style={tdMonoStyle}>{endpointLabel(row, "dst")}</td>
                <td style={tdMonoStyle}>{row["network_flow.ip_protocol"]}</td>
                <td style={tdMonoStyle}>{row["network_flow.bytes_count"]}</td>
                <td style={tdMonoStyle}>{row["network_flow.packets_count"]}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <QueryPagination
        canNext={canNext}
        canPrevious={canPrevious}
        loading={loading}
        notice={notice}
        pageNumber={pageNumber}
        onNext={onNext}
        onPrevious={onPrevious}
      />
    </section>
  );
}

function RejectedRowsPanel({
  activeTable,
  canNext,
  canPrevious,
  diagnostics,
  loading,
  notice,
  onNext,
  onPrevious,
  pageNumber,
}: {
  readonly activeTable: NetworkFlowTable | null;
  readonly canNext: boolean;
  readonly canPrevious: boolean;
  readonly diagnostics: readonly NetworkFlowDiagnostic[];
  readonly loading: boolean;
  readonly notice: string | null;
  readonly onNext: () => void;
  readonly onPrevious: () => void;
  readonly pageNumber: number;
}) {
  return (
    <section
      aria-label="Network Flow rejected rows"
      data-testid={networkAnalysisTestId("table-panel")}
      style={panelGridStyle}
    >
      <PanelHeader table={activeTable} />
      <div style={tableScrollStyle}>
        <table style={dataTableStyle}>
          <thead>
            <tr>
              <th style={thStyle}>Row</th>
              <th style={thStyle}>Column</th>
              <th style={thStyle}>Field</th>
              <th style={thStyle}>Reason</th>
              <th style={thStyle}>Diagnostic</th>
            </tr>
          </thead>
          <tbody>
            {diagnostics.map((diagnostic) => (
              <tr key={diagnostic.diagnostic_id}>
                <td style={tdMonoStyle}>{diagnostic.source_row_number}</td>
                <td style={tdMonoStyle}>
                  {diagnostic.source_column_ordinal ?? ""}
                </td>
                <td style={tdMonoStyle}>{diagnostic.field_key ?? ""}</td>
                <td style={tdMonoStyle}>{diagnostic.reason_code}</td>
                <td style={tdStyle}>{diagnostic.message}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <QueryPagination
        canNext={canNext}
        canPrevious={canPrevious}
        loading={loading}
        notice={notice}
        pageNumber={pageNumber}
        onNext={onNext}
        onPrevious={onPrevious}
      />
    </section>
  );
}

function QueryPagination({
  canNext,
  canPrevious,
  loading,
  notice,
  onNext,
  onPrevious,
  pageNumber,
}: {
  readonly canNext: boolean;
  readonly canPrevious: boolean;
  readonly loading: boolean;
  readonly notice: string | null;
  readonly onNext: () => void;
  readonly onPrevious: () => void;
  readonly pageNumber: number;
}) {
  return (
    <nav aria-label="Network Flow result pages" style={paginationStyle}>
      <button
        data-testid={networkAnalysisTestId("page-previous")}
        disabled={!canPrevious || loading}
        type="button"
        onClick={onPrevious}
      >
        Previous
      </button>
      <span
        aria-live="polite"
        data-testid={networkAnalysisTestId("page-status")}
      >
        {loading ? "Loading page" : `Page ${pageNumber}`}
      </span>
      <button
        data-testid={networkAnalysisTestId("page-next")}
        disabled={!canNext || loading}
        type="button"
        onClick={onNext}
      >
        Next
      </button>
      {notice === null ? null : <span role="status">{notice}</span>}
    </nav>
  );
}

function GraphPanel({
  canLink,
  contributors,
  firstContributor,
  graph,
  selectedEdge,
  onCloseDrawer,
  onLinkDst,
  onLinkSrc,
  onSelectEdge,
}: {
  readonly canLink: boolean;
  readonly contributors: readonly NetworkFlowContributor[];
  readonly firstContributor: NetworkFlowContributor | null;
  readonly graph: NetworkFlowGraphResult | null;
  readonly selectedEdge: NetworkFlowEdgeAnnotation | null;
  readonly onCloseDrawer: () => void;
  readonly onLinkDst: () => void;
  readonly onLinkSrc: () => void;
  readonly onSelectEdge: (edgeId: string) => void;
}) {
  return (
    <section
      aria-label="Network Flow graph"
      data-testid={networkAnalysisTestId("graph-panel")}
      style={graphLayoutStyle}
    >
      <div style={graphTableStyle}>
        <div style={graphSummaryStyle}>
          <Network aria-hidden="true" size={18} />
          <span>
            {graph ? compactID(graph.graph_query_digest) : "No graph"}
          </span>
          <span style={mutedTextStyle}>
            {graph?.source_table_refs.length ?? 0} tables
          </span>
        </div>
        <div style={tableScrollStyle}>
          <table style={dataTableStyle}>
            <thead>
              <tr>
                <th style={thStyle}>Edge</th>
                <th style={thStyle}>Rows</th>
                <th style={thStyle}>Examples</th>
                <th style={thStyle}>Open</th>
              </tr>
            </thead>
            <tbody>
              {(graph?.edge_annotations ?? []).map((edge) => (
                <tr
                  key={edge.edge_id}
                  data-testid={networkAnalysisEdgeTestId(edge.edge_id)}
                >
                  <td style={tdMonoStyle}>{compactID(edge.edge_id)}</td>
                  <td style={tdMonoStyle}>{edge.example_refs_total_count}</td>
                  <td style={tdMonoStyle}>{edge.example_row_refs.length}</td>
                  <td style={tdStyle}>
                    <button
                      style={iconButtonStyle}
                      title="Open contributors"
                      type="button"
                      onClick={() => onSelectEdge(edge.edge_id)}
                    >
                      <PanelRightOpen aria-hidden="true" size={15} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      {selectedEdge ? (
        <aside
          aria-label="Graph contributors"
          data-testid={networkAnalysisTestId("contributor-drawer")}
          style={drawerStyle}
        >
          <div style={drawerHeaderStyle}>
            <strong>{compactID(selectedEdge.edge_id)}</strong>
            <button
              style={iconButtonStyle}
              title="Close"
              type="button"
              onClick={onCloseDrawer}
            >
              <X aria-hidden="true" size={15} />
            </button>
          </div>
          {canLink && firstContributor ? (
            <div style={linkActionsStyle}>
              <button
                style={commandButtonStyle}
                type="button"
                onClick={onLinkSrc}
              >
                <Link2 aria-hidden="true" size={15} />
                Link source
              </button>
              <button
                style={commandButtonStyle}
                type="button"
                onClick={onLinkDst}
              >
                <Link2 aria-hidden="true" size={15} />
                Link destination
              </button>
            </div>
          ) : null}
          <div style={drawerRowsStyle}>
            {contributors.map((contributor) => (
              <div
                key={contributor.row_ref.network_flow_row_id}
                style={drawerRowStyle}
              >
                <span style={monoTextStyle}>
                  {contributor.row_ref.source_row_number}
                </span>
                <span style={monoTextStyle}>
                  {endpointLabel(contributor.row, "src")}
                </span>
                <span style={monoTextStyle}>
                  {endpointLabel(contributor.row, "dst")}
                </span>
              </div>
            ))}
          </div>
        </aside>
      ) : null}
    </section>
  );
}

function PanelHeader({ table }: { readonly table: NetworkFlowTable | null }) {
  return (
    <div style={panelHeaderStyle}>
      <strong>{table?.display_name ?? "No active table"}</strong>
      {table ? (
        <span style={mutedTextStyle}>
          {table.row_count_accepted} accepted / {table.row_count_rejected}{" "}
          rejected
        </span>
      ) : null}
    </div>
  );
}

function endpointLabel(row: NetworkFlowRow, kind: "src" | "dst"): string {
  const ip =
    kind === "src" ? row["network_flow.src_ip"] : row["network_flow.dst_ip"];
  const port =
    kind === "src"
      ? row["network_flow.src_port"]
      : row["network_flow.dst_port"];
  return port === null ? ip : `${ip}:${port}`;
}

function shortTimestamp(value: string): string {
  return value.replace("T", " ").replace(/(?:\.[0-9]+)?Z$/u, "Z");
}

function compactID(value: string): string {
  if (value.length <= 18) {
    return value;
  }
  return `${value.slice(0, 10)}...${value.slice(-6)}`;
}

const workspaceStyle = {
  display: "grid",
  gridTemplateRows:
    "var(--ct-layout-viewBarHeight) auto minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
  background: "var(--ct-colors-canvas)",
  color: "var(--ct-colors-text-primary)",
} satisfies CSSProperties;

const viewBarStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
  paddingInline: "var(--ct-spacing-md)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  minWidth: 0,
} satisfies CSSProperties;

const tabsStyle = {
  display: "flex",
  gap: "var(--ct-spacing-xs)",
  minWidth: 0,
  overflowX: "auto",
} satisfies CSSProperties;

const innerTabStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
  minBlockSize: "2rem",
  maxInlineSize: "18rem",
  paddingInline: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-text-primary)",
  whiteSpace: "nowrap",
  overflow: "hidden",
  textOverflow: "ellipsis",
} satisfies CSSProperties;

const innerTabActiveStyle = {
  borderColor: "var(--ct-colors-accent-strong)",
  background: "var(--ct-colors-surface-2)",
} satisfies CSSProperties;

const tabCountStyle = {
  fontVariantNumeric: "tabular-nums",
  color: "var(--ct-colors-text-muted)",
} satisfies CSSProperties;

const viewActionsStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const commandButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "var(--ct-spacing-xs)",
  minBlockSize: "2rem",
  paddingInline: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-text-primary)",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const iconButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  inlineSize: "2rem",
  blockSize: "2rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-text-primary)",
} satisfies CSSProperties;

const modeBarStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-md)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

const modeButtonStyle = {
  ...commandButtonStyle,
  minBlockSize: "1.75rem",
} satisfies CSSProperties;

const modeButtonActiveStyle = {
  borderColor: "var(--ct-colors-accent-strong)",
  background: "var(--ct-colors-surface-2)",
} satisfies CSSProperties;

const workAreaStyle = {
  minBlockSize: 0,
  minWidth: 0,
  overflow: "hidden",
} satisfies CSSProperties;

const emptyStateStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  blockSize: "100%",
  minBlockSize: "12rem",
  borderBlockEnd: "var(--ct-border-hairline)",
} satisfies CSSProperties;

const panelGridStyle = {
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr) auto",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;

const paginationStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-end",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-md)",
  borderBlockStart: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  fontSize: "0.8125rem",
} satisfies CSSProperties;

const panelHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

const tableScrollStyle = {
  minBlockSize: 0,
  minWidth: 0,
  overflow: "auto",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

const dataTableStyle = {
  borderCollapse: "collapse",
  inlineSize: "100%",
  minInlineSize: "48rem",
  fontSize: "0.8125rem",
} satisfies CSSProperties;

const thStyle = {
  position: "sticky",
  top: 0,
  zIndex: 1,
  textAlign: "start",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-text-muted)",
  fontWeight: 600,
} satisfies CSSProperties;

const tdStyle = {
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
  verticalAlign: "top",
} satisfies CSSProperties;

const tdMonoStyle = {
  ...tdStyle,
  fontFamily: "var(--ct-font-family-mono)",
  fontVariantNumeric: "tabular-nums",
} satisfies CSSProperties;

const graphLayoutStyle = {
  position: "relative",
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;

const graphTableStyle = {
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;

const graphSummaryStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

const drawerStyle = {
  position: "absolute",
  insetBlock: 0,
  insetInlineEnd: 0,
  display: "grid",
  gridTemplateRows: "auto auto minmax(0, 1fr)",
  inlineSize:
    "min(var(--ct-layout-inspectorDefaultWidth), calc(100% - var(--ct-spacing-xl)))",
  minInlineSize:
    "min(var(--ct-layout-inspectorMinWidth), calc(100% - var(--ct-spacing-xl)))",
  borderInlineStart: "var(--ct-border-hairline)",
  background: "var(--ct-component-inspector-backgroundColor)",
  boxShadow: "var(--ct-elevation-drawer)",
  zIndex: 3,
} satisfies CSSProperties;

const drawerHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
} satisfies CSSProperties;

const linkActionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
  padding: "var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
} satisfies CSSProperties;

const drawerRowsStyle = {
  overflow: "auto",
  minBlockSize: 0,
} satisfies CSSProperties;

const drawerRowStyle = {
  display: "grid",
  gridTemplateColumns: "4rem minmax(0, 1fr) minmax(0, 1fr)",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
} satisfies CSSProperties;

const statusStripStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-md)",
  minBlockSize: "var(--ct-layout-statusStripHeight)",
  paddingInline: "var(--ct-spacing-md)",
  borderBlockStart: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-text-muted)",
  fontSize: "0.75rem",
} satisfies CSSProperties;

const mutedTextStyle = {
  color: "var(--ct-colors-text-muted)",
} satisfies CSSProperties;

const errorTextStyle = {
  color: "var(--ct-colors-danger-strong)",
} satisfies CSSProperties;

const monoTextStyle = {
  fontFamily: "var(--ct-font-family-mono)",
  fontVariantNumeric: "tabular-nums",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
} satisfies CSSProperties;
