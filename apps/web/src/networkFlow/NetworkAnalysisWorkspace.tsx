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
import {
  type ChangeEvent,
  type CSSProperties,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import {
  importNetworkFlowCSV,
  linkNetworkFlowIndicator,
  listNetworkFlowTables,
  type NetworkFlowContributor,
  type NetworkFlowDiagnostic,
  type NetworkFlowEdgeAnnotation,
  type NetworkFlowGraphResult,
  type NetworkFlowRow,
  type NetworkFlowTable,
  networkAnalysisSheetRef,
  queryNetworkFlowContributors,
  queryNetworkFlowGraph,
  queryNetworkFlowRejectedRows,
  queryNetworkFlowTable,
} from "./networkFlowClient";
import {
  type NetworkFlowExtensionResourceChange,
  useNetworkFlowExtensionEvents,
} from "./useNetworkFlowExtensionEvents";

type NetworkAnalysisMode = "rows" | "rejected" | "graph";

export type NetworkAnalysisWorkspaceProps = {
  readonly apiBase?: string | undefined;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
};

const activeTableScopeLabel = networkAnalysisSheetRef();

export function NetworkAnalysisWorkspace({
  apiBase,
  currentIncidentRole,
  incidentId,
  onIncidentAccessLost,
}: NetworkAnalysisWorkspaceProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [tables, setTables] = useState<NetworkFlowTable[]>([]);
  const [activeTableId, setActiveTableId] = useState<string | null>(null);
  const [mode, setMode] = useState<NetworkAnalysisMode>("rows");
  const [rows, setRows] = useState<NetworkFlowRow[]>([]);
  const [diagnostics, setDiagnostics] = useState<NetworkFlowDiagnostic[]>([]);
  const [graph, setGraph] = useState<NetworkFlowGraphResult | null>(null);
  const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null);
  const [contributors, setContributors] = useState<NetworkFlowContributor[]>(
    [],
  );
  const [loadState, setLoadState] = useState<"loading" | "ready" | "error">(
    "loading",
  );
  const [message, setMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [importing, setImporting] = useState(false);

  const canImport =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const canLink =
    currentIncidentRole === "editor" || currentIncidentRole === "admin";

  const activeTable = useMemo(
    () =>
      tables.find((table) => table.network_flow_table_id === activeTableId) ??
      null,
    [activeTableId, tables],
  );
  const graphTableIds = useMemo(
    () => tables.map((table) => table.network_flow_table_id),
    [tables],
  );
  const selectedEdge = useMemo(
    () =>
      graph?.edge_annotations.find((edge) => edge.edge_id === selectedEdgeId) ??
      null,
    [graph, selectedEdgeId],
  );
  const firstContributor = contributors[0] ?? null;

  const loadTables = useCallback(async () => {
    setLoadState((current) => (current === "ready" ? current : "loading"));
    try {
      const nextTables = await listNetworkFlowTables({ apiBase, incidentId });
      setTables(nextTables);
      setActiveTableId((current) => {
        if (
          current !== null &&
          nextTables.some((table) => table.network_flow_table_id === current)
        ) {
          return current;
        }
        return nextTables[0]?.network_flow_table_id ?? null;
      });
      setLoadState("ready");
      setErrorMessage(null);
    } catch (error) {
      const nextMessage =
        error instanceof Error ? error.message : "load_failed";
      if (isAuthorizationLoss(nextMessage)) {
        onIncidentAccessLost?.();
      }
      setLoadState("error");
      setErrorMessage(nextMessage);
    }
  }, [apiBase, incidentId, onIncidentAccessLost]);

  useEffect(() => {
    void loadTables();
  }, [loadTables]);

  useEffect(() => {
    if (activeTableId === null || mode === "graph") {
      setRows([]);
      setDiagnostics([]);
      return;
    }
    let cancelled = false;
    const load = async () => {
      try {
        if (mode === "rows") {
          const result = await queryNetworkFlowTable({
            apiBase,
            incidentId,
            tableId: activeTableId,
          });
          if (!cancelled) {
            setRows(result.rows);
            setDiagnostics([]);
            setErrorMessage(null);
          }
          return;
        }
        const result = await queryNetworkFlowRejectedRows({
          apiBase,
          incidentId,
          tableId: activeTableId,
        });
        if (!cancelled) {
          setDiagnostics(result.diagnostics);
          setRows([]);
          setErrorMessage(null);
        }
      } catch (error) {
        if (!cancelled) {
          const nextMessage =
            error instanceof Error ? error.message : "query_failed";
          if (isAuthorizationLoss(nextMessage)) {
            onIncidentAccessLost?.();
          }
          setErrorMessage(nextMessage);
        }
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [activeTableId, apiBase, incidentId, mode, onIncidentAccessLost]);

  useEffect(() => {
    if (mode !== "graph" || graphTableIds.length === 0) {
      setGraph(null);
      setSelectedEdgeId(null);
      setContributors([]);
      return;
    }
    let cancelled = false;
    const loadGraph = async () => {
      try {
        const nextGraph = await queryNetworkFlowGraph({
          apiBase,
          incidentId,
          tableIds: graphTableIds,
        });
        if (!cancelled) {
          setGraph(nextGraph);
          setSelectedEdgeId((current) =>
            current !== null &&
            nextGraph.edge_annotations.some((edge) => edge.edge_id === current)
              ? current
              : (nextGraph.edge_annotations[0]?.edge_id ?? null),
          );
          setErrorMessage(null);
        }
      } catch (error) {
        if (!cancelled) {
          const nextMessage =
            error instanceof Error ? error.message : "graph_query_failed";
          if (isAuthorizationLoss(nextMessage)) {
            onIncidentAccessLost?.();
          }
          setErrorMessage(nextMessage);
        }
      }
    };
    void loadGraph();
    return () => {
      cancelled = true;
    };
  }, [apiBase, graphTableIds, incidentId, mode, onIncidentAccessLost]);

  useEffect(() => {
    if (graph === null || selectedEdgeId === null) {
      setContributors([]);
      return;
    }
    let cancelled = false;
    const loadContributors = async () => {
      try {
        const result = await queryNetworkFlowContributors({
          apiBase,
          incidentId,
          graph,
          selector: { kind: "edge", edge_id: selectedEdgeId },
        });
        if (!cancelled) {
          setContributors(result.contributors);
          setErrorMessage(null);
        }
      } catch (error) {
        if (!cancelled) {
          const nextMessage =
            error instanceof Error
              ? error.message
              : "contributors_query_failed";
          if (isAuthorizationLoss(nextMessage)) {
            onIncidentAccessLost?.();
          }
          setErrorMessage(nextMessage);
          setContributors([]);
        }
      }
    };
    void loadContributors();
    return () => {
      cancelled = true;
    };
  }, [apiBase, graph, incidentId, onIncidentAccessLost, selectedEdgeId]);

  const handleResourceChange = useCallback(
    (change: NetworkFlowExtensionResourceChange) => {
      setMessage("Network Analysis data changed.");
      if (
        change.reasonCode === "authorization_lost" ||
        (change.changeKind === "remove" && change.resourceId === "*")
      ) {
        setTables([]);
        setActiveTableId(null);
        setRows([]);
        setDiagnostics([]);
        setGraph(null);
        setSelectedEdgeId(null);
        setContributors([]);
        setMessage("Network Analysis access changed.");
        return;
      }
      if (change.changeKind === "remove") {
        setTables((currentTables) => {
          const nextTables = currentTables.filter(
            (table) => table.network_flow_table_id !== change.resourceId,
          );
          setActiveTableId((current) =>
            current === change.resourceId
              ? (nextTables[0]?.network_flow_table_id ?? null)
              : current,
          );
          return nextTables;
        });
        setRows([]);
        setDiagnostics([]);
        setGraph(null);
        setSelectedEdgeId(null);
        setContributors([]);
        return;
      }
      void loadTables();
    },
    [loadTables],
  );

  useNetworkFlowExtensionEvents({
    apiBase,
    enabled: true,
    incidentId,
    onResourceChange: handleResourceChange,
  });

  const handleImportChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] ?? null;
    event.target.value = "";
    if (file === null || importing || !canImport) {
      return;
    }
    setImporting(true);
    setErrorMessage(null);
    try {
      await importNetworkFlowCSV({
        apiBase,
        incidentId,
        file,
        onProgress: setMessage,
      });
      setMessage("Import applied.");
      await loadTables();
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "import_failed");
    } finally {
      setImporting(false);
    }
  };

  const handleLinkEdge = async (
    fieldKey: "network_flow.src_ip" | "network_flow.dst_ip",
  ) => {
    if (graph === null || selectedEdge === null || firstContributor === null) {
      return;
    }
    const confirmExactValue = firstContributor.row[fieldKey];
    if (confirmExactValue.trim() === "") {
      setErrorMessage("indicator_candidate_unavailable");
      return;
    }
    try {
      const result = await linkNetworkFlowIndicator({
        apiBase,
        incidentId,
        graph,
        edgeId: selectedEdge.edge_id,
        fieldKey,
        confirmExactValue,
      });
      setMessage(
        result.duplicate
          ? "Indicator link already exists."
          : "Indicator link created.",
      );
      setErrorMessage(null);
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : "indicator_link_failed",
      );
    }
  };

  const empty = loadState === "ready" && tables.length === 0;

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
          {tables.map((table, index) => {
            const selected = table.network_flow_table_id === activeTableId;
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
                  setActiveTableId(table.network_flow_table_id);
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
              void loadTables();
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
            onChange={handleImportChange}
          />
          {canImport ? (
            <button
              data-testid={networkAnalysisTestId("import-trigger")}
              disabled={importing}
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
          disabled={activeTable === null}
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
          disabled={activeTable === null}
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
          disabled={tables.length === 0}
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
            importing={importing}
            onImport={() => fileInputRef.current?.click()}
          />
        ) : mode === "graph" ? (
          <GraphPanel
            canLink={canLink}
            contributors={contributors}
            firstContributor={firstContributor}
            graph={graph}
            selectedEdge={selectedEdge}
            onCloseDrawer={() => {
              setSelectedEdgeId(null);
              setContributors([]);
            }}
            onLinkDst={() => void handleLinkEdge("network_flow.dst_ip")}
            onLinkSrc={() => void handleLinkEdge("network_flow.src_ip")}
            onSelectEdge={setSelectedEdgeId}
          />
        ) : mode === "rejected" ? (
          <RejectedRowsPanel
            activeTable={activeTable}
            diagnostics={diagnostics}
          />
        ) : (
          <RowsPanel activeTable={activeTable} rows={rows} />
        )}
      </div>

      <div style={statusStripStyle}>
        <span>
          {loadState === "loading"
            ? "Loading"
            : `${tables.length} active table${tables.length === 1 ? "" : "s"}`}
        </span>
        {message ? (
          <span data-testid={networkAnalysisTestId("stale-state")}>
            {message}
          </span>
        ) : null}
        {errorMessage ? (
          <span style={errorTextStyle}>{errorMessage}</span>
        ) : null}
      </div>
    </section>
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
  rows,
}: {
  readonly activeTable: NetworkFlowTable | null;
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
    </section>
  );
}

function RejectedRowsPanel({
  activeTable,
  diagnostics,
}: {
  readonly activeTable: NetworkFlowTable | null;
  readonly diagnostics: readonly NetworkFlowDiagnostic[];
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
    </section>
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

function isAuthorizationLoss(message: string): boolean {
  return (
    message.includes("authorization_denied") ||
    message.includes("incident_not_found") ||
    message.includes("session_required")
  );
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
  gridTemplateRows: "auto minmax(0, 1fr)",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
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
