import {
  networkAnalysisEdgeTestId,
  networkAnalysisTableTabTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import {
  Link2,
  Network,
  PanelRightOpen,
  Pencil,
  RefreshCw,
  Table2,
  Trash2,
  Upload,
  X,
} from "lucide-react";
import {
  type CSSProperties,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { IncidentCollaborationBoundary } from "../collaboration/IncidentCollaborationSession";
import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import { NetworkFlowMappingModal } from "./NetworkFlowMappingModal";
import {
  NetworkFlowAcceptedQueryControls,
  NetworkFlowRejectedQueryControls,
} from "./NetworkFlowQueryControls";
import {
  NetworkFlowAcceptedGrid,
  NetworkFlowRejectedGrid,
} from "./NetworkFlowSemanticGrid";
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
  isNetworkFlowAuthorizationLoss,
  isNetworkFlowLifecycleLoss,
  isNetworkFlowProtectedStateLoss,
  NetworkFlowRequestError,
  type NetworkFlowWorkspaceError,
  networkFlowErrorMessage,
} from "./networkFlowErrors";
import type {
  NetworkFlowAcceptedQuery,
  NetworkFlowRejectedQuery,
} from "./networkFlowQueryModel";
import { useNetworkFlowCollaborationController } from "./useNetworkFlowCollaborationController";
import { useNetworkFlowGraphController } from "./useNetworkFlowGraphController";
import { useNetworkFlowImportController } from "./useNetworkFlowImportController";
import { useNetworkFlowIndicatorLinkController } from "./useNetworkFlowIndicatorLinkController";
import type { NetworkFlowQueryLoadState } from "./useNetworkFlowPagedQuery";
import { useNetworkFlowRejectedRowsController } from "./useNetworkFlowRejectedRowsController";
import { useNetworkFlowRowsController } from "./useNetworkFlowRowsController";
import {
  type NetworkFlowTableMutationState,
  useNetworkFlowTableController,
} from "./useNetworkFlowTableController";

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
  const [protectedStateError, setProtectedStateError] =
    useState<NetworkFlowRequestError | null>(null);
  const handleWorkspaceError = useCallback(
    (error: NetworkFlowWorkspaceError | null) => {
      setErrorMessage(error);
      if (
        error instanceof NetworkFlowRequestError &&
        isNetworkFlowProtectedStateLoss(error)
      ) {
        setProtectedStateError(error);
      }
    },
    [],
  );
  const canRead =
    currentIncidentRole === "viewer" ||
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const canImport =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const canRename = canImport;
  const canLink = canImport;
  const canDelete =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const tableController = useNetworkFlowTableController({
    apiBase,
    enabled: canRead,
    incidentId,
    onIncidentAccessLost,
  });
  const rowsController = useNetworkFlowRowsController({
    activeTableId: tableController.activeTableId,
    apiBase,
    enabled: mode === "rows",
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
  });
  const rejectedRowsController = useNetworkFlowRejectedRowsController({
    activeTableId: tableController.activeTableId,
    apiBase,
    enabled: mode === "rejected",
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
  });
  const graphController = useNetworkFlowGraphController({
    apiBase,
    enabled: mode === "graph",
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
    tableIds: tableController.tableIds,
  });
  const importController = useNetworkFlowImportController({
    apiBase,
    canImport,
    incidentId,
    onError: handleWorkspaceError,
    onImported: tableController.loadTables,
    onMessage: setMessage,
  });
  const clearRows = rowsController.clearRows;
  const clearDiagnostics = rejectedRowsController.clearDiagnostics;
  const clearGraph = graphController.clearGraph;
  const resetImport = importController.reset;
  const clearResources = useCallback(() => {
    clearRows();
    clearDiagnostics();
    clearGraph();
    resetImport();
    setMode("rows");
  }, [clearDiagnostics, clearGraph, clearRows, resetImport]);
  useNetworkFlowCollaborationController({
    apiBase,
    clearResources,
    dispatchTableAction: tableController.dispatch,
    incidentId,
    loadTables: tableController.loadTables,
    onMessage: setMessage,
    onProtectedStateLoss: handleWorkspaceError,
  });
  const indicatorLinkController = useNetworkFlowIndicatorLinkController({
    apiBase,
    firstContributor: graphController.firstContributor,
    graph: graphController.graph,
    incidentId,
    onError: handleWorkspaceError,
    onMessage: setMessage,
    selectedEdge: graphController.selectedEdge,
  });
  useEffect(() => {
    if (!canImport) {
      resetImport();
    }
  }, [canImport, resetImport]);
  useEffect(() => {
    void incidentId;
    setErrorMessage(null);
    setProtectedStateError(null);
  }, [incidentId]);
  useEffect(() => {
    const error =
      protectedStateError ??
      (errorMessage instanceof NetworkFlowRequestError
        ? errorMessage
        : tableController.error);
    if (error === null || !isNetworkFlowProtectedStateLoss(error)) {
      return;
    }
    clearResources();
    if (
      isNetworkFlowAuthorizationLoss(error) ||
      error.code === "incident_closed"
    ) {
      tableController.clearAuthorization();
      return;
    }
    if (isNetworkFlowLifecycleLoss(error)) {
      void tableController.loadTables();
    }
  }, [
    clearResources,
    errorMessage,
    protectedStateError,
    tableController.clearAuthorization,
    tableController.error,
    tableController.loadTables,
  ]);
  useEffect(() => {
    if (
      protectedStateError !== null &&
      protectedStateError.code !== "incident_closed" &&
      isNetworkFlowLifecycleLoss(protectedStateError) &&
      tableController.loadState === "ready" &&
      tableController.activeTableId !== null
    ) {
      setErrorMessage(null);
      setProtectedStateError(null);
    }
  }, [
    protectedStateError,
    tableController.activeTableId,
    tableController.loadState,
  ]);
  const empty =
    tableController.loadState === "ready" &&
    tableController.tables.length === 0;
  const visibleError =
    protectedStateError ?? errorMessage ?? tableController.error;
  const blockingState = networkFlowWorkspaceBlockingState({
    canRead,
    error: visibleError,
    loadState: tableController.loadState,
    tableCount: tableController.tables.length,
  });

  return (
    <section
      aria-label="Network Analysis"
      data-extension-profile-id={activeTableScopeLabel.extension_profile_id}
      data-testid={networkAnalysisTestId("workspace")}
      data-workspace-key={activeTableScopeLabel.workspace_key}
      style={workspaceStyle}
    >
      <header
        data-testid={networkAnalysisTestId("workspace-header")}
        style={workspaceHeaderStyle}
      >
        <div>
          <h2 style={workspaceTitleStyle}>Network Analysis</h2>
          <span style={mutedTextStyle}>
            {tableController.tables.length} active table
            {tableController.tables.length === 1 ? "" : "s"}
            {tableController.activeTable === null
              ? ""
              : ` · ${tableController.activeTable.display_name}`}
          </span>
        </div>
        <div style={viewActionsStyle}>
          <TableLifecycleControls
            canDelete={canDelete}
            canRename={canRename}
            mutationState={tableController.mutationState}
            table={tableController.activeTable}
            onDelete={async (table) => {
              const deleted = await tableController.softDeleteTable({
                baseTableVersion: table.table_version,
                tableId: table.network_flow_table_id,
              });
              if (deleted) {
                clearResources();
                setMessage(`${table.display_name} was deleted.`);
              }
              return deleted;
            }}
            onRename={async (table, displayName) => {
              const renamed = await tableController.renameTable({
                baseTableVersion: table.table_version,
                displayName,
                tableId: table.network_flow_table_id,
              });
              if (renamed) {
                setMessage(
                  displayName.trim() === table.display_name
                    ? "The table name is unchanged."
                    : "The table was renamed.",
                );
              }
              return renamed;
            }}
          />
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
      </header>
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

      <div style={queryBandsStyle}>
        {tableController.activeTable === null ? null : (
          <>
            <DiagnosticsSummary table={tableController.activeTable} />
            {mode === "rejected" ? (
              <NetworkFlowRejectedQueryControls
                query={rejectedRowsController.query}
                onChange={rejectedRowsController.setQuery}
              />
            ) : (
              <NetworkFlowAcceptedQueryControls
                graphMode={mode === "graph"}
                query={rowsController.query}
                onChange={rowsController.setQuery}
              />
            )}
          </>
        )}
      </div>

      <div style={workAreaStyle}>
        {blockingState !== null ? (
          <NetworkFlowBlockingState
            state={blockingState}
            onRetry={() => {
              void tableController.loadTables();
            }}
          />
        ) : empty ? (
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
            loadState={rejectedRowsController.loadState}
            notice={rejectedRowsController.notice}
            pageNumber={rejectedRowsController.pageNumber}
            query={rejectedRowsController.query}
            onNext={rejectedRowsController.nextPage}
            onPrevious={rejectedRowsController.previousPage}
            onResetQuery={() =>
              rejectedRowsController.setQuery({
                errorCodes: [],
                fieldKeys: [],
                sourceRowRange: null,
              })
            }
            onRetry={rejectedRowsController.refresh}
          />
        ) : (
          <RowsPanel
            activeTable={tableController.activeTable}
            canNext={rowsController.canNext}
            canPrevious={rowsController.canPrevious}
            loadState={rowsController.loadState}
            notice={rowsController.notice}
            pageNumber={rowsController.pageNumber}
            query={rowsController.query}
            rows={rowsController.rows}
            onNext={rowsController.nextPage}
            onPrevious={rowsController.previousPage}
            onResetQuery={() =>
              rowsController.setQuery({
                filters: [],
                sort: [],
                timeWindow: null,
              })
            }
            onRetry={rowsController.refresh}
            onSortChange={(sort) =>
              rowsController.setQuery((current) => ({ ...current, sort }))
            }
          />
        )}
      </div>

      <div style={statusStripStyle}>
        <span>
          {tableController.loadState === "loading"
            ? "Loading"
            : tableController.loadState === "refreshing"
              ? "Refreshing table metadata"
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

function TableLifecycleControls({
  canDelete,
  canRename,
  mutationState,
  onDelete,
  onRename,
  table,
}: {
  readonly canDelete: boolean;
  readonly canRename: boolean;
  readonly mutationState: NetworkFlowTableMutationState;
  readonly onDelete: (table: NetworkFlowTable) => Promise<boolean>;
  readonly onRename: (
    table: NetworkFlowTable,
    displayName: string,
  ) => Promise<boolean>;
  readonly table: NetworkFlowTable | null;
}) {
  const tableId = table?.network_flow_table_id ?? null;
  const [dialog, setDialog] = useState<"delete" | "rename" | null>(null);
  const [renameValue, setRenameValue] = useState("");
  const [deleteConfirmation, setDeleteConfirmation] = useState("");
  useEffect(() => {
    void tableId;
    setDialog(null);
    setRenameValue("");
    setDeleteConfirmation("");
  }, [tableId]);
  useEffect(() => {
    if (
      (dialog === "rename" && !canRename) ||
      (dialog === "delete" && !canDelete)
    ) {
      setDialog(null);
      setRenameValue("");
      setDeleteConfirmation("");
    }
  }, [canDelete, canRename, dialog]);
  if (table === null) {
    return null;
  }
  const busy = mutationState.kind !== "idle";
  return (
    <>
      {canRename ? (
        <button
          data-testid={networkAnalysisTestId("rename-trigger")}
          disabled={busy}
          style={iconButtonStyle}
          title="Rename active table"
          type="button"
          onClick={() => {
            setRenameValue(table.display_name);
            setDialog("rename");
          }}
        >
          <Pencil aria-hidden="true" size={16} />
          Rename
        </button>
      ) : null}
      {canDelete ? (
        <button
          data-testid={networkAnalysisTestId("delete-trigger")}
          disabled={busy}
          style={dangerButtonStyle}
          type="button"
          onClick={() => {
            setDeleteConfirmation("");
            setDialog("delete");
          }}
        >
          <Trash2 aria-hidden="true" size={16} />
          Delete
        </button>
      ) : null}
      {dialog === "rename" && canRename ? (
        <div style={commandDialogBackdropStyle}>
          <form
            aria-labelledby="network-flow-rename-title"
            aria-modal="true"
            data-testid={networkAnalysisTestId("rename-dialog")}
            role="dialog"
            style={commandDialogStyle}
            onSubmit={(event) => {
              event.preventDefault();
              void onRename(table, renameValue).then((renamed) => {
                if (renamed) {
                  setDialog(null);
                }
              });
            }}
          >
            <h3 id="network-flow-rename-title">Rename Network Flow table</h3>
            <label style={fieldLabelStyle}>
              Display name
              <input
                data-testid={networkAnalysisTestId("rename-input")}
                maxLength={64}
                required
                value={renameValue}
                onChange={(event) => setRenameValue(event.currentTarget.value)}
              />
            </label>
            <div style={dialogActionsStyle}>
              <button
                data-testid={networkAnalysisTestId("rename-cancel")}
                disabled={busy}
                type="button"
                onClick={() => setDialog(null)}
              >
                Cancel
              </button>
              <button
                data-testid={networkAnalysisTestId("rename-submit")}
                disabled={busy || renameValue.trim() === ""}
                type="submit"
              >
                {mutationState.kind === "renaming" ? "Renaming…" : "Rename"}
              </button>
            </div>
          </form>
        </div>
      ) : null}
      {dialog === "delete" && canDelete ? (
        <div style={commandDialogBackdropStyle}>
          <section
            aria-describedby="network-flow-delete-description"
            aria-labelledby="network-flow-delete-title"
            aria-modal="true"
            data-testid={networkAnalysisTestId("delete-dialog")}
            role="alertdialog"
            style={commandDialogStyle}
          >
            <h3 id="network-flow-delete-title">Delete Network Flow table</h3>
            <p id="network-flow-delete-description">
              This soft-deletes <strong>{table.display_name}</strong> and makes
              its rows, diagnostics, graph results, and cursors unavailable.
              Type the exact table name to confirm.
            </p>
            <label style={fieldLabelStyle}>
              Confirm table name
              <input
                data-testid={networkAnalysisTestId("delete-confirmation")}
                value={deleteConfirmation}
                onChange={(event) =>
                  setDeleteConfirmation(event.currentTarget.value)
                }
              />
            </label>
            <div style={dialogActionsStyle}>
              <button
                data-testid={networkAnalysisTestId("delete-cancel")}
                disabled={busy}
                type="button"
                onClick={() => setDialog(null)}
              >
                Cancel
              </button>
              <button
                data-testid={networkAnalysisTestId("delete-confirm")}
                disabled={busy || deleteConfirmation !== table.display_name}
                style={dangerButtonStyle}
                type="button"
                onClick={() => {
                  void onDelete(table).then((deleted) => {
                    if (deleted) {
                      setDialog(null);
                    }
                  });
                }}
              >
                {mutationState.kind === "deleting"
                  ? "Deleting…"
                  : "Delete table"}
              </button>
            </div>
          </section>
        </div>
      ) : null}
    </>
  );
}

type NetworkFlowBlockingState =
  | { readonly kind: "lifecycle"; readonly message: string }
  | { readonly kind: "loading"; readonly message: string }
  | { readonly kind: "permission"; readonly message: string }
  | { readonly kind: "unavailable"; readonly message: string };

function networkFlowWorkspaceBlockingState(options: {
  readonly canRead: boolean;
  readonly error: NetworkFlowWorkspaceError | null;
  readonly loadState: "error" | "loading" | "ready" | "refreshing";
  readonly tableCount: number;
}): NetworkFlowBlockingState | null {
  if (!options.canRead) {
    return {
      kind: "permission",
      message:
        "You no longer have access to Network Analysis for this incident.",
    };
  }
  if (options.error instanceof NetworkFlowRequestError) {
    if (isNetworkFlowAuthorizationLoss(options.error)) {
      return { kind: "permission", message: options.error.message };
    }
    if (isNetworkFlowLifecycleLoss(options.error)) {
      return { kind: "lifecycle", message: options.error.message };
    }
  }
  if (
    options.tableCount === 0 &&
    (options.loadState === "loading" || options.loadState === "refreshing")
  ) {
    return { kind: "loading", message: "Loading Network Analysis…" };
  }
  if (options.tableCount === 0 && options.loadState === "error") {
    return {
      kind: "unavailable",
      message:
        networkFlowErrorMessage(options.error) ??
        "Network Analysis is unavailable.",
    };
  }
  return null;
}

function NetworkFlowBlockingState({
  onRetry,
  state,
}: {
  readonly onRetry: () => void;
  readonly state: NetworkFlowBlockingState;
}) {
  return (
    <section
      aria-label={`Network Analysis ${state.kind} state`}
      style={emptyStateStyle}
    >
      <strong>
        {state.kind === "permission"
          ? "Access changed"
          : state.kind === "lifecycle"
            ? "Resource unavailable"
            : state.kind === "loading"
              ? "Loading"
              : "Unavailable"}
      </strong>
      <span>{state.message}</span>
      {state.kind === "unavailable" ? (
        <button type="button" onClick={onRetry}>
          Retry
        </button>
      ) : null}
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
  canNext,
  canPrevious,
  loadState,
  notice,
  onNext,
  onPrevious,
  onResetQuery,
  onRetry,
  onSortChange,
  pageNumber,
  query,
  rows,
}: {
  readonly activeTable: NetworkFlowTable | null;
  readonly canNext: boolean;
  readonly canPrevious: boolean;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly notice: string | null;
  readonly onNext: () => void;
  readonly onPrevious: () => void;
  readonly onResetQuery: () => void;
  readonly onRetry: () => void;
  readonly onSortChange: (sort: NetworkFlowAcceptedQuery["sort"]) => void;
  readonly pageNumber: number;
  readonly query: NetworkFlowAcceptedQuery;
  readonly rows: readonly NetworkFlowRow[];
}) {
  const loading = loadState === "loading" || loadState === "refreshing";
  return (
    <section
      aria-label="Network Flow table rows"
      data-testid={networkAnalysisTestId("table-panel")}
      style={panelGridStyle}
    >
      <PanelHeader table={activeTable} />
      <NetworkFlowAcceptedGrid
        filtered={query.filters.length > 0 || query.timeWindow !== null}
        loadState={loadState}
        resetKey={`${activeTable?.network_flow_table_id ?? "none"}:${JSON.stringify(query)}`}
        rows={rows}
        sort={query.sort}
        onResetQuery={onResetQuery}
        onRetry={onRetry}
        onSortChange={onSortChange}
      />
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
  loadState,
  notice,
  onNext,
  onPrevious,
  onResetQuery,
  onRetry,
  pageNumber,
  query,
}: {
  readonly activeTable: NetworkFlowTable | null;
  readonly canNext: boolean;
  readonly canPrevious: boolean;
  readonly diagnostics: readonly NetworkFlowDiagnostic[];
  readonly loadState: NetworkFlowQueryLoadState;
  readonly notice: string | null;
  readonly onNext: () => void;
  readonly onPrevious: () => void;
  readonly onResetQuery: () => void;
  readonly onRetry: () => void;
  readonly pageNumber: number;
  readonly query: NetworkFlowRejectedQuery;
}) {
  const loading = loadState === "loading" || loadState === "refreshing";
  return (
    <section
      aria-label="Network Flow rejected rows"
      data-testid={networkAnalysisTestId("table-panel")}
      style={panelGridStyle}
    >
      <PanelHeader table={activeTable} />
      <NetworkFlowRejectedGrid
        diagnostics={diagnostics}
        filtered={
          query.errorCodes.length > 0 ||
          query.fieldKeys.length > 0 ||
          query.sourceRowRange !== null
        }
        loadState={loadState}
        resetKey={`${activeTable?.network_flow_table_id ?? "none"}:${JSON.stringify(query)}`}
        onResetQuery={onResetQuery}
        onRetry={onRetry}
      />
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

function DiagnosticsSummary({ table }: { readonly table: NetworkFlowTable }) {
  return (
    <section
      aria-label="Network Flow diagnostics summary"
      data-testid={networkAnalysisTestId("diagnostics-summary")}
      style={diagnosticsSummaryStyle}
    >
      <span>
        <strong>{table.row_count_accepted}</strong> accepted
      </span>
      <span>
        <strong>{table.row_count_rejected}</strong> rejected
      </span>
      <span>
        Mapping <code>{table.source_profile_id}</code>
      </span>
      <span>
        Parser <code>{table.parser_profile_id}</code>
      </span>
      <span>
        Source <strong>{table.source_filename_display}</strong>
      </span>
      <span>
        Diagnostics {table.diagnostics_truncated ? "truncated" : "complete"}
      </span>
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

function compactID(value: string): string {
  if (value.length <= 18) {
    return value;
  }
  return `${value.slice(0, 10)}...${value.slice(-6)}`;
}

const workspaceStyle = {
  display: "grid",
  gridTemplateRows:
    "auto var(--ct-layout-viewBarHeight) auto auto minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
  background: "var(--ct-colors-canvas)",
  color: "var(--ct-colors-text-primary)",
} satisfies CSSProperties;

const queryBandsStyle = {
  minWidth: 0,
} satisfies CSSProperties;

const workspaceHeaderStyle = {
  alignItems: "center",
  background: "var(--ct-colors-surface-1)",
  borderBlockEnd: "var(--ct-border-hairline)",
  display: "flex",
  gap: "var(--ct-spacing-md)",
  justifyContent: "space-between",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
} satisfies CSSProperties;

const workspaceTitleStyle = {
  fontSize: "1rem",
  margin: 0,
} satisfies CSSProperties;

const diagnosticsSummaryStyle = {
  alignItems: "center",
  background: "var(--ct-colors-surface-1)",
  borderBlockEnd: "var(--ct-border-hairline)",
  display: "flex",
  flexWrap: "wrap",
  fontSize: "0.75rem",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-md)",
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

const dangerButtonStyle = {
  ...commandButtonStyle,
  borderColor: "var(--ct-colors-danger-strong)",
  color: "var(--ct-colors-danger-strong)",
} satisfies CSSProperties;

const commandDialogBackdropStyle = {
  alignItems: "center",
  background: "color-mix(in srgb, var(--ct-colors-canvas) 72%, transparent)",
  display: "flex",
  inset: 0,
  justifyContent: "center",
  position: "fixed",
  zIndex: 20,
} satisfies CSSProperties;

const commandDialogStyle = {
  background: "var(--ct-colors-surface-1)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  boxShadow: "var(--ct-elevation-drawer)",
  display: "grid",
  gap: "var(--ct-spacing-md)",
  inlineSize: "min(32rem, calc(100vw - 2rem))",
  padding: "var(--ct-spacing-lg)",
} satisfies CSSProperties;

const fieldLabelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const dialogActionsStyle = {
  display: "flex",
  gap: "var(--ct-spacing-sm)",
  justifyContent: "flex-end",
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
  flexDirection: "column",
  gap: "var(--ct-spacing-sm)",
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
