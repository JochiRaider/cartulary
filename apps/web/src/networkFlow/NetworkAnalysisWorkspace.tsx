import type { GridCellAnchor, GridCellRange } from "@cartulary/grid-adapter";
import {
  networkAnalysisEdgeTestId,
  networkAnalysisTableTabTestId,
  networkAnalysisTestId,
  networkAnalysisVertexTestId,
} from "@cartulary/ui-contracts";
import {
  Link2,
  Network,
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
  useMemo,
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
  NetworkFlowContributorGrid,
  NetworkFlowRejectedGrid,
} from "./NetworkFlowSemanticGrid";
import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowGraphEdge,
  NetworkFlowGraphResult,
  NetworkFlowGraphVertex,
  NetworkFlowIndicatorTarget,
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
import {
  type NetworkFlowRowLinkSelection,
  networkFlowRowLinkCandidate,
  resolveNetworkFlowRowLinkSelection,
} from "./networkFlowIndicatorLinkModel";
import type {
  NetworkFlowAcceptedQuery,
  NetworkFlowRejectedQuery,
} from "./networkFlowQueryModel";
import { useNetworkFlowCollaborationController } from "./useNetworkFlowCollaborationController";
import {
  type NetworkFlowGraphScopeMode,
  useNetworkFlowGraphController,
} from "./useNetworkFlowGraphController";
import { useNetworkFlowImportController } from "./useNetworkFlowImportController";
import {
  type NetworkFlowIndicatorLinkCandidate,
  useNetworkFlowIndicatorLinkController,
} from "./useNetworkFlowIndicatorLinkController";
import { useNetworkFlowModalFocus } from "./useNetworkFlowModalFocus";
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
  const [rowGridSelection, setRowGridSelection] = useState<{
    readonly activeAnchor: GridCellAnchor | null;
    readonly cellRange: GridCellRange | null;
  }>({ activeAnchor: null, cellRange: null });
  const [linkCandidate, setLinkCandidate] =
    useState<NetworkFlowIndicatorLinkCandidate | null>(null);
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
    activeTableId: tableController.activeTableId,
    apiBase,
    enabled: mode === "graph",
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
    query: rowsController.query,
    tables: tableController.tables,
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
    setLinkCandidate(null);
    setRowGridSelection({ activeAnchor: null, cellRange: null });
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
    activeCandidateKey: linkCandidate?.key ?? null,
    apiBase,
    enabled: canLink,
    incidentId,
    onError: handleWorkspaceError,
    onGraphStale: graphController.markGraphStale,
    onMessage: setMessage,
  });
  const rowLinkSelection = useMemo(
    () =>
      resolveNetworkFlowRowLinkSelection({
        activeAnchor: rowGridSelection.activeAnchor,
        bindingSourceRowLimit: indicatorLinkController.bindingSourceRowLimit,
        cellRange: rowGridSelection.cellRange,
        rows: rowsController.rows,
      }),
    [
      indicatorLinkController.bindingSourceRowLimit,
      rowGridSelection.activeAnchor,
      rowGridSelection.cellRange,
      rowsController.rows,
    ],
  );
  const handleRowGridSelectionChange = useCallback(
    (activeAnchor: GridCellAnchor | null, cellRange: GridCellRange | null) => {
      setRowGridSelection({ activeAnchor, cellRange });
      setLinkCandidate(null);
    },
    [],
  );
  useEffect(() => {
    if (!canImport) {
      resetImport();
    }
  }, [canImport, resetImport]);
  useEffect(() => {
    if (!canLink) {
      setLinkCandidate(null);
    }
  }, [canLink]);
  useEffect(() => {
    if (
      linkCandidate?.selector.kind === "graph_vertex" ||
      linkCandidate?.selector.kind === "graph_edge"
    ) {
      if (
        graphController.graph?.graph_query_digest !==
        linkCandidate.selector.graph_query_digest
      ) {
        setLinkCandidate(null);
      }
    }
  }, [graphController.graph, linkCandidate]);
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
      tabIndex={-1}
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
                aria-controls="network-flow-work-area"
                aria-selected={selected}
                data-testid={networkAnalysisTableTabTestId(
                  table.network_flow_table_id,
                )}
                role="tab"
                style={{
                  ...innerTabStyle,
                  ...(selected ? innerTabActiveStyle : null),
                }}
                tabIndex={selected ? 0 : -1}
                type="button"
                onKeyDown={(event) => {
                  if (
                    event.key !== "ArrowLeft" &&
                    event.key !== "ArrowRight" &&
                    event.key !== "Home" &&
                    event.key !== "End"
                  ) {
                    return;
                  }
                  event.preventDefault();
                  const targetIndex =
                    event.key === "Home"
                      ? 0
                      : event.key === "End"
                        ? tableController.tables.length - 1
                        : (index +
                            (event.key === "ArrowLeft" ? -1 : 1) +
                            tableController.tables.length) %
                          tableController.tables.length;
                  const targetTable = tableController.tables[targetIndex];
                  if (targetTable === undefined) return;
                  tableController.dispatch({
                    type: "select_table",
                    tableId: targetTable.network_flow_table_id,
                  });
                  setMode("rows");
                  event.currentTarget.parentElement
                    ?.querySelectorAll<HTMLButtonElement>('[role="tab"]')
                    [targetIndex]?.focus();
                }}
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

      <div id="network-flow-work-area" style={workAreaStyle}>
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
            canNextContributorPage={graphController.canNextContributorPage}
            canPreviousContributorPage={
              graphController.canPreviousContributorPage
            }
            contributorLoadState={graphController.contributorLoadState}
            contributorPageNumber={graphController.contributorPageNumber}
            contributors={graphController.contributors}
            firstContributor={graphController.firstContributor}
            graph={graphController.graph}
            graphLoadState={graphController.graphLoadState}
            scopeMode={graphController.scopeMode}
            selectedEdge={graphController.selectedEdge}
            selectedTableIds={graphController.selectedTableIds}
            selectedVertex={graphController.selectedVertex}
            tables={tableController.tables}
            onCloseDrawer={() => graphController.selectGraphObject(null)}
            onLinkEdge={(fieldKey) => {
              const candidate = networkFlowEdgeLinkCandidate({
                edge: graphController.selectedEdge,
                fieldKey,
                firstContributor: graphController.firstContributor,
                graph: graphController.graph,
              });
              setLinkCandidate(candidate);
            }}
            onLinkVertex={() => {
              setLinkCandidate(
                networkFlowVertexLinkCandidate(
                  graphController.graph,
                  graphController.selectedVertex,
                ),
              );
            }}
            onNextContributorPage={graphController.nextContributorPage}
            onPreviousContributorPage={graphController.previousContributorPage}
            onRefreshGraph={graphController.refreshGraph}
            onRetryContributors={graphController.retryContributorPage}
            onScopeModeChange={graphController.setScopeMode}
            onSelectEdge={(edgeId) =>
              graphController.selectGraphObject({
                kind: "edge",
                edge_id: edgeId,
              })
            }
            onSelectTable={graphController.setTableSelected}
            onSelectVertex={(vertexId) =>
              graphController.selectGraphObject({
                kind: "vertex",
                vertex_id: vertexId,
              })
            }
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
            canLink={canLink}
            canNext={rowsController.canNext}
            canPrevious={rowsController.canPrevious}
            loadState={rowsController.loadState}
            notice={rowsController.notice}
            pageNumber={rowsController.pageNumber}
            query={rowsController.query}
            rows={rowsController.rows}
            rowLinkSelection={rowLinkSelection}
            onBeginLink={() => {
              setLinkCandidate(
                rowLinkSelection === null
                  ? null
                  : networkFlowRowLinkCandidate(rowLinkSelection),
              );
            }}
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
            onSelectionChange={handleRowGridSelectionChange}
          />
        )}
      </div>

      <div
        data-testid={networkAnalysisTestId("status-strip")}
        style={statusStripStyle}
      >
        <span aria-atomic="true" aria-live="polite" role="status">
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
        </span>
        {visibleError ? (
          <span aria-atomic="true" role="alert" style={errorTextStyle}>
            Error: {networkFlowErrorMessage(visibleError)}
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
      {linkCandidate !== null ? (
        <IndicatorLinkDialog
          candidate={linkCandidate}
          linking={indicatorLinkController.linking}
          onCancel={() => setLinkCandidate(null)}
          onSubmit={async ({ confirmExactValue, target }) => {
            const linked = await indicatorLinkController.link({
              candidate: linkCandidate,
              confirmExactValue,
              target,
            });
            if (linked) {
              setLinkCandidate(null);
            }
            return linked;
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
        <RenameTableDialog
          busy={busy}
          renameValue={renameValue}
          renaming={mutationState.kind === "renaming"}
          onCancel={() => setDialog(null)}
          onRenameValueChange={setRenameValue}
          onSubmit={() => {
            void onRename(table, renameValue).then((renamed) => {
              if (renamed) setDialog(null);
            });
          }}
        />
      ) : null}
      {dialog === "delete" && canDelete ? (
        <DeleteTableDialog
          busy={busy}
          confirmation={deleteConfirmation}
          deleting={mutationState.kind === "deleting"}
          table={table}
          onCancel={() => setDialog(null)}
          onConfirmationChange={setDeleteConfirmation}
          onSubmit={() => {
            void onDelete(table).then((deleted) => {
              if (deleted) setDialog(null);
            });
          }}
        />
      ) : null}
    </>
  );
}

function RenameTableDialog({
  busy,
  onCancel,
  onRenameValueChange,
  onSubmit,
  renameValue,
  renaming,
}: {
  readonly busy: boolean;
  readonly onCancel: () => void;
  readonly onRenameValueChange: (value: string) => void;
  readonly onSubmit: () => void;
  readonly renameValue: string;
  readonly renaming: boolean;
}) {
  const modalFocus = useNetworkFlowModalFocus<HTMLFormElement>({
    dismissDisabled: busy,
    initialFocusTestId: networkAnalysisTestId("rename-input"),
    onDismiss: onCancel,
  });
  return (
    <div style={commandDialogBackdropStyle}>
      <form
        ref={modalFocus.dialogRef}
        aria-labelledby="network-flow-rename-title"
        aria-modal="true"
        data-testid={networkAnalysisTestId("rename-dialog")}
        role="dialog"
        style={commandDialogStyle}
        onKeyDown={modalFocus.onKeyDown}
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit();
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
            onChange={(event) => onRenameValueChange(event.currentTarget.value)}
          />
        </label>
        <div style={dialogActionsStyle}>
          <button
            data-testid={networkAnalysisTestId("rename-cancel")}
            disabled={busy}
            type="button"
            onClick={onCancel}
          >
            Cancel
          </button>
          <button
            data-testid={networkAnalysisTestId("rename-submit")}
            disabled={busy || renameValue.trim() === ""}
            type="submit"
          >
            {renaming ? "Renaming…" : "Rename"}
          </button>
        </div>
      </form>
    </div>
  );
}

function DeleteTableDialog({
  busy,
  confirmation,
  deleting,
  onCancel,
  onConfirmationChange,
  onSubmit,
  table,
}: {
  readonly busy: boolean;
  readonly confirmation: string;
  readonly deleting: boolean;
  readonly onCancel: () => void;
  readonly onConfirmationChange: (value: string) => void;
  readonly onSubmit: () => void;
  readonly table: NetworkFlowTable;
}) {
  const modalFocus = useNetworkFlowModalFocus<HTMLFormElement>({
    dismissDisabled: busy,
    initialFocusTestId: networkAnalysisTestId("delete-confirmation"),
    onDismiss: onCancel,
  });
  return (
    <div style={commandDialogBackdropStyle}>
      <form
        ref={modalFocus.dialogRef}
        aria-describedby="network-flow-delete-description"
        aria-labelledby="network-flow-delete-title"
        aria-modal="true"
        data-testid={networkAnalysisTestId("delete-dialog")}
        role="alertdialog"
        style={commandDialogStyle}
        onKeyDown={modalFocus.onKeyDown}
        onSubmit={(event) => {
          event.preventDefault();
          if (confirmation === table.display_name) onSubmit();
        }}
      >
        <h3 id="network-flow-delete-title">Delete Network Flow table</h3>
        <p id="network-flow-delete-description">
          This soft-deletes <strong>{table.display_name}</strong> and makes its
          rows, diagnostics, graph results, and cursors unavailable. Type the
          exact table name to confirm.
        </p>
        <label style={fieldLabelStyle}>
          Confirm table name
          <input
            data-testid={networkAnalysisTestId("delete-confirmation")}
            value={confirmation}
            onChange={(event) =>
              onConfirmationChange(event.currentTarget.value)
            }
          />
        </label>
        <div style={dialogActionsStyle}>
          <button
            data-testid={networkAnalysisTestId("delete-cancel")}
            disabled={busy}
            type="button"
            onClick={onCancel}
          >
            Cancel
          </button>
          <button
            data-testid={networkAnalysisTestId("delete-confirm")}
            disabled={busy || confirmation !== table.display_name}
            style={dangerButtonStyle}
            type="submit"
          >
            {deleting ? "Deleting…" : "Delete table"}
          </button>
        </div>
      </form>
    </div>
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
      aria-live={state.kind === "loading" ? "polite" : "assertive"}
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
  canLink,
  canNext,
  canPrevious,
  loadState,
  notice,
  onBeginLink,
  onNext,
  onPrevious,
  onResetQuery,
  onRetry,
  onSortChange,
  onSelectionChange,
  pageNumber,
  query,
  rows,
  rowLinkSelection,
}: {
  readonly activeTable: NetworkFlowTable | null;
  readonly canLink: boolean;
  readonly canNext: boolean;
  readonly canPrevious: boolean;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly notice: string | null;
  readonly onBeginLink: () => void;
  readonly onNext: () => void;
  readonly onPrevious: () => void;
  readonly onResetQuery: () => void;
  readonly onRetry: () => void;
  readonly onSortChange: (sort: NetworkFlowAcceptedQuery["sort"]) => void;
  readonly onSelectionChange: (
    activeAnchor: GridCellAnchor | null,
    cellRange: GridCellRange | null,
  ) => void;
  readonly pageNumber: number;
  readonly query: NetworkFlowAcceptedQuery;
  readonly rows: readonly NetworkFlowRow[];
  readonly rowLinkSelection: NetworkFlowRowLinkSelection | null;
}) {
  const loading = loadState === "loading" || loadState === "refreshing";
  return (
    <section
      aria-label="Network Flow table rows"
      data-testid={networkAnalysisTestId("table-panel")}
      style={canLink ? panelGridWithActionsStyle : panelGridStyle}
    >
      <PanelHeader table={activeTable} />
      {canLink ? (
        <div style={linkActionsStyle}>
          <button
            disabled={rowLinkSelection === null}
            style={commandButtonStyle}
            type="button"
            onClick={onBeginLink}
          >
            <Link2 aria-hidden="true" size={15} />
            {rowLinkSelection === null
              ? "Select one IP cell or same-value IP range"
              : `Link ${rowLinkSelection.rows.length} selected row${
                  rowLinkSelection.rows.length === 1 ? "" : "s"
                }`}
          </button>
        </div>
      ) : null}
      <NetworkFlowAcceptedGrid
        filtered={query.filters.length > 0 || query.timeWindow !== null}
        loadState={loadState}
        resetKey={`${activeTable?.network_flow_table_id ?? "none"}:${JSON.stringify(query)}`}
        rows={rows}
        sort={query.sort}
        onResetQuery={onResetQuery}
        onRetry={onRetry}
        onSelectionChange={onSelectionChange}
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
  canNextContributorPage,
  canPreviousContributorPage,
  contributorLoadState,
  contributorPageNumber,
  contributors,
  firstContributor,
  graph,
  graphLoadState,
  scopeMode,
  selectedEdge,
  selectedTableIds,
  selectedVertex,
  tables,
  onCloseDrawer,
  onLinkEdge,
  onLinkVertex,
  onNextContributorPage,
  onPreviousContributorPage,
  onRefreshGraph,
  onRetryContributors,
  onScopeModeChange,
  onSelectEdge,
  onSelectTable,
  onSelectVertex,
}: {
  readonly canLink: boolean;
  readonly canNextContributorPage: boolean;
  readonly canPreviousContributorPage: boolean;
  readonly contributorLoadState: NetworkFlowQueryLoadState;
  readonly contributorPageNumber: number;
  readonly contributors: readonly NetworkFlowContributor[];
  readonly firstContributor: NetworkFlowContributor | null;
  readonly graph: NetworkFlowGraphResult | null;
  readonly graphLoadState: NetworkFlowQueryLoadState;
  readonly scopeMode: NetworkFlowGraphScopeMode;
  readonly selectedEdge: NetworkFlowGraphEdge | null;
  readonly selectedTableIds: readonly string[];
  readonly selectedVertex: NetworkFlowGraphVertex | null;
  readonly tables: readonly NetworkFlowTable[];
  readonly onCloseDrawer: () => void;
  readonly onLinkEdge: (
    fieldKey: "network_flow.src_ip" | "network_flow.dst_ip",
  ) => void;
  readonly onLinkVertex: () => void;
  readonly onNextContributorPage: () => void;
  readonly onPreviousContributorPage: () => void;
  readonly onRefreshGraph: () => void;
  readonly onRetryContributors: () => void;
  readonly onScopeModeChange: (mode: NetworkFlowGraphScopeMode) => void;
  readonly onSelectEdge: (edgeId: string) => void;
  readonly onSelectTable: (tableId: string, selected: boolean) => void;
  readonly onSelectVertex: (vertexId: string) => void;
}) {
  const selectedGraphButtonRef = useRef<HTMLButtonElement | null>(null);
  const selectedObject = selectedVertex ?? selectedEdge;
  const graphVertices = useMemo(
    () =>
      [...(graph?.graph_projection_result.vertices ?? [])].sort((left, right) =>
        graphVertexLabel(left).localeCompare(graphVertexLabel(right)),
      ),
    [graph],
  );
  const graphEndpointLabels = useMemo(
    () =>
      new Map(
        graphVertices.flatMap((vertex) => {
          const endpointId = semanticGraphVertexId(vertex);
          return endpointId === null
            ? []
            : ([[endpointId, graphVertexLabel(vertex)]] as const);
        }),
      ),
    [graphVertices],
  );
  const graphEdges = useMemo(
    () =>
      [...(graph?.graph_projection_result.edges ?? [])].sort((left, right) =>
        graphEdgeLabel(left, graphEndpointLabels).localeCompare(
          graphEdgeLabel(right, graphEndpointLabels),
        ),
      ),
    [graph, graphEndpointLabels],
  );
  const graphTableLabels = useMemo(
    () =>
      new Map(
        tables.map((table) => [
          table.network_flow_table_id,
          table.display_name,
        ]),
      ),
    [tables],
  );
  return (
    <section
      aria-label="Network Flow graph"
      data-testid={networkAnalysisTestId("graph-panel")}
      style={graphLayoutStyle}
    >
      <div style={graphTableStyle}>
        <fieldset
          data-testid={networkAnalysisTestId("graph-scope")}
          style={graphScopeStyle}
        >
          <legend>Graph scope</legend>
          {(
            [
              ["active_table", "Active table"],
              ["selected_tables", "Selected tables"],
              ["all_active_tables", "All active tables"],
            ] as const
          ).map(([value, label]) => (
            <label key={value} style={inlineControlStyle}>
              <input
                checked={scopeMode === value}
                name="network-flow-graph-scope"
                type="radio"
                value={value}
                onChange={() => onScopeModeChange(value)}
              />
              {label}
            </label>
          ))}
          {scopeMode === "selected_tables" ? (
            <div style={graphTableSelectionStyle}>
              {tables.map((table) => {
                const checked = selectedTableIds.includes(
                  table.network_flow_table_id,
                );
                return (
                  <label
                    key={table.network_flow_table_id}
                    style={inlineControlStyle}
                  >
                    <input
                      checked={checked}
                      disabled={checked && selectedTableIds.length === 1}
                      type="checkbox"
                      onChange={(event) =>
                        onSelectTable(
                          table.network_flow_table_id,
                          event.currentTarget.checked,
                        )
                      }
                    />
                    {table.display_name}
                  </label>
                );
              })}
            </div>
          ) : null}
        </fieldset>
        <div style={graphSummaryStyle}>
          <Network aria-hidden="true" size={18} />
          <span>
            {graphLoadState === "loading"
              ? "Loading graph…"
              : graph
                ? "Graph ready"
                : "No graph"}
          </span>
          <span style={mutedTextStyle}>
            {graph?.source_table_refs.length ?? 0} tables ·{" "}
            {graph?.graph_projection_result.vertices.length ?? 0} vertices ·{" "}
            {graph?.graph_projection_result.edges.length ?? 0} edges
          </span>
          {graphLoadState === "error" ? (
            <button type="button" onClick={onRefreshGraph}>
              <RefreshCw aria-hidden="true" size={14} /> Retry
            </button>
          ) : null}
        </div>
        <div style={tableScrollStyle}>
          <h3 style={graphSectionTitleStyle}>Vertices</h3>
          <table style={dataTableStyle}>
            <thead>
              <tr>
                <th style={thStyle}>Endpoint</th>
                <th style={thStyle}>Flows</th>
                <th style={thStyle}>Tables</th>
                <th style={thStyle}>Select</th>
              </tr>
            </thead>
            <tbody>
              {graphVertices.map((vertex) => {
                const vertexId = semanticGraphVertexId(vertex);
                if (vertexId === null) {
                  return null;
                }
                return (
                  <tr
                    key={vertex.vertex_id}
                    data-testid={networkAnalysisVertexTestId(vertexId)}
                  >
                    <td style={tdMonoStyle}>
                      {graphScalar(vertex.properties.endpoint_value)}
                    </td>
                    <td style={tdMonoStyle}>
                      {graphScalar(vertex.properties.flow_row_count)}
                    </td>
                    <td style={tdMonoStyle}>
                      {graphTableList(
                        vertex.properties.contributing_table_ids,
                        graphTableLabels,
                      )}
                    </td>
                    <td style={tdStyle}>
                      <button
                        ref={
                          selectedVertex === vertex
                            ? selectedGraphButtonRef
                            : undefined
                        }
                        aria-pressed={selectedVertex === vertex}
                        type="button"
                        onClick={() => onSelectVertex(vertexId)}
                      >
                        Select vertex
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          <h3 style={graphSectionTitleStyle}>Edges</h3>
          <table style={dataTableStyle}>
            <thead>
              <tr>
                <th style={thStyle}>Source</th>
                <th style={thStyle}>Destination</th>
                <th style={thStyle}>Protocol</th>
                <th style={thStyle}>Rows</th>
                <th style={thStyle}>Select</th>
              </tr>
            </thead>
            <tbody>
              {graphEdges.map((edge) => {
                const edgeId = semanticGraphEdgeId(edge);
                if (edgeId === null) {
                  return null;
                }
                const annotation = graph?.edge_annotations.find(
                  (candidate) => candidate.edge_id === edgeId,
                );
                return (
                  <tr
                    key={edge.edge_id}
                    data-testid={networkAnalysisEdgeTestId(edgeId)}
                  >
                    <td style={tdMonoStyle}>
                      {graphEndpointLabel(
                        edge.properties.src_endpoint_id,
                        graphEndpointLabels,
                      )}
                    </td>
                    <td style={tdMonoStyle}>
                      {graphEndpointLabel(
                        edge.properties.dst_endpoint_id,
                        graphEndpointLabels,
                      )}
                    </td>
                    <td style={tdMonoStyle}>
                      {graphScalar(edge.properties.ip_protocol)}
                    </td>
                    <td style={tdMonoStyle}>
                      {annotation?.example_refs_total_count ??
                        graphScalar(edge.properties.flow_row_count)}
                    </td>
                    <td style={tdStyle}>
                      <button
                        ref={
                          selectedEdge === edge
                            ? selectedGraphButtonRef
                            : undefined
                        }
                        aria-pressed={selectedEdge === edge}
                        type="button"
                        onClick={() => onSelectEdge(edgeId)}
                      >
                        Select edge
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
      {selectedObject ? (
        <aside
          aria-label="Graph contributors"
          data-testid={networkAnalysisTestId("contributor-drawer")}
          style={drawerStyle}
          onKeyDown={(event) => {
            if (event.key !== "Escape") return;
            event.preventDefault();
            const returnTarget = selectedGraphButtonRef.current;
            onCloseDrawer();
            queueMicrotask(() => returnTarget?.focus());
          }}
        >
          <div style={drawerHeaderStyle}>
            <strong>
              {selectedVertex
                ? `Vertex ${graphVertexLabel(selectedVertex)}`
                : `Edge ${graphEdgeLabel(selectedEdge as NetworkFlowGraphEdge, graphEndpointLabels)}`}
            </strong>
            <button
              data-testid={networkAnalysisTestId("contributor-close")}
              style={iconButtonStyle}
              title="Close"
              type="button"
              onClick={() => {
                const returnTarget = selectedGraphButtonRef.current;
                onCloseDrawer();
                queueMicrotask(() => returnTarget?.focus());
              }}
            >
              <X aria-hidden="true" size={15} />
            </button>
          </div>
          <div style={linkActionsStyle}>
            {canLink && selectedVertex ? (
              <button
                style={commandButtonStyle}
                type="button"
                onClick={onLinkVertex}
              >
                <Link2 aria-hidden="true" size={15} />
                Link vertex
              </button>
            ) : null}
            {canLink && selectedEdge && firstContributor ? (
              <>
                <button
                  style={commandButtonStyle}
                  type="button"
                  onClick={() => onLinkEdge("network_flow.src_ip")}
                >
                  <Link2 aria-hidden="true" size={15} />
                  Link source
                </button>
                <button
                  style={commandButtonStyle}
                  type="button"
                  onClick={() => onLinkEdge("network_flow.dst_ip")}
                >
                  <Link2 aria-hidden="true" size={15} />
                  Link destination
                </button>
              </>
            ) : null}
          </div>
          <NetworkFlowContributorGrid
            contributors={contributors}
            loadState={contributorLoadState}
            tables={tables}
            onRetry={onRetryContributors}
          />
          <QueryPagination
            canNext={canNextContributorPage}
            canPrevious={canPreviousContributorPage}
            loading={
              contributorLoadState === "loading" ||
              contributorLoadState === "refreshing"
            }
            notice={null}
            pageNumber={contributorPageNumber}
            onNext={onNextContributorPage}
            onPrevious={onPreviousContributorPage}
          />
        </aside>
      ) : null}
      <span
        aria-live="polite"
        data-testid={networkAnalysisTestId("graph-live-region")}
        style={visuallyHiddenStyle}
      >
        {selectedVertex
          ? `Vertex selected. ${contributors.length} contributors on page ${contributorPageNumber}.`
          : selectedEdge
            ? `Edge selected. ${contributors.length} contributors on page ${contributorPageNumber}.`
            : graphLoadState === "loading"
              ? "Loading Network Flow graph."
              : graph
                ? "Network Flow graph ready. Select a vertex or edge to inspect contributors."
                : "Network Flow graph unavailable."}
      </span>
    </section>
  );
}

function networkFlowVertexLinkCandidate(
  graph: NetworkFlowGraphResult | null,
  vertex: NetworkFlowGraphVertex | null,
): NetworkFlowIndicatorLinkCandidate | null {
  const vertexId = vertex === null ? null : semanticGraphVertexId(vertex);
  const candidateValue =
    vertex === null
      ? null
      : graphString(vertex.properties.indicator_candidate_value);
  if (graph === null || vertexId === null || candidateValue === null) {
    return null;
  }
  return {
    candidateValue,
    key: `${graph.graph_query_digest}:vertex:${vertexId}:${candidateValue}`,
    label: `graph vertex ${compactID(vertexId)}`,
    selector: {
      kind: "graph_vertex",
      graph_query: graph.semantic_query,
      graph_query_digest: graph.graph_query_digest,
      vertex_id: vertexId,
    },
  };
}

function networkFlowEdgeLinkCandidate(options: {
  readonly edge: NetworkFlowGraphEdge | null;
  readonly fieldKey: "network_flow.src_ip" | "network_flow.dst_ip";
  readonly firstContributor: NetworkFlowContributor | null;
  readonly graph: NetworkFlowGraphResult | null;
}): NetworkFlowIndicatorLinkCandidate | null {
  const edgeId =
    options.edge === null ? null : semanticGraphEdgeId(options.edge);
  const candidateValue =
    options.firstContributor?.row[options.fieldKey] ?? null;
  if (
    options.graph === null ||
    edgeId === null ||
    typeof candidateValue !== "string"
  ) {
    return null;
  }
  return {
    candidateValue,
    key: `${options.graph.graph_query_digest}:edge:${edgeId}:${options.fieldKey}:${candidateValue}`,
    label: `${options.fieldKey === "network_flow.src_ip" ? "source" : "destination"} endpoint on graph edge ${compactID(edgeId)}`,
    selector: {
      kind: "graph_edge",
      graph_query: options.graph.semantic_query,
      graph_query_digest: options.graph.graph_query_digest,
      edge_id: edgeId,
      field_key: options.fieldKey,
    },
  };
}

function semanticGraphVertexId(vertex: NetworkFlowGraphVertex): string | null {
  return vertex.source_entity_ref?.source_entity_id ?? null;
}

function semanticGraphEdgeId(edge: NetworkFlowGraphEdge): string | null {
  return (
    graphString(edge.properties.edge_id) ??
    edge.source_relationship_ref?.source_relationship_id ??
    null
  );
}

function graphString(value: unknown): string | null {
  return typeof value === "string" && value !== "" ? value : null;
}

function graphScalar(value: unknown): string {
  return typeof value === "string" || typeof value === "number"
    ? String(value)
    : "—";
}

function graphVertexLabel(vertex: NetworkFlowGraphVertex): string {
  return graphScalar(vertex.properties.endpoint_value);
}

function graphEndpointLabel(
  value: unknown,
  labels: ReadonlyMap<string, string>,
): string {
  const endpointId = graphString(value);
  return endpointId === null
    ? "—"
    : (labels.get(endpointId) ?? "Unavailable endpoint");
}

function graphEdgeLabel(
  edge: NetworkFlowGraphEdge,
  endpointLabels: ReadonlyMap<string, string>,
): string {
  const source = graphEndpointLabel(
    edge.properties.src_endpoint_id,
    endpointLabels,
  );
  const destination = graphEndpointLabel(
    edge.properties.dst_endpoint_id,
    endpointLabels,
  );
  const protocol = graphScalar(edge.properties.ip_protocol);
  const port = graphScalar(edge.properties.dst_port);
  return `${source} → ${destination} · protocol ${protocol} · port ${port}`;
}

function graphTableList(
  value: unknown,
  labels: ReadonlyMap<string, string>,
): string {
  if (!Array.isArray(value)) return "—";
  return value
    .filter((item): item is string => typeof item === "string")
    .map((tableId) => labels.get(tableId) ?? "Unavailable table")
    .join(", ");
}

function IndicatorLinkDialog({
  candidate,
  linking,
  onCancel,
  onSubmit,
}: {
  readonly candidate: NetworkFlowIndicatorLinkCandidate;
  readonly linking: boolean;
  readonly onCancel: () => void;
  readonly onSubmit: (options: {
    readonly confirmExactValue: string;
    readonly target: NetworkFlowIndicatorTarget;
  }) => Promise<boolean>;
}) {
  const [mode, setMode] = useState<"create" | "existing">("create");
  const [existingIndicatorId, setExistingIndicatorId] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const modalFocus = useNetworkFlowModalFocus<HTMLFormElement>({
    dismissDisabled: linking,
    initialFocusTestId: networkAnalysisTestId("indicator-link-confirmation"),
    onDismiss: onCancel,
  });
  const target: NetworkFlowIndicatorTarget | null =
    mode === "create"
      ? {
          mode: "create_indicator",
          indicator_type: candidate.candidateValue.includes(":")
            ? "ipv6_addr"
            : "ipv4_addr",
        }
      : existingIndicatorId.trim() === ""
        ? null
        : {
            mode: "existing_indicator",
            indicator_id: existingIndicatorId.trim(),
          };
  return (
    <div style={commandDialogBackdropStyle}>
      <form
        ref={modalFocus.dialogRef}
        aria-labelledby="network-flow-indicator-link-title"
        aria-modal="true"
        data-testid={networkAnalysisTestId("indicator-link-dialog")}
        role="dialog"
        style={commandDialogStyle}
        onKeyDown={modalFocus.onKeyDown}
        onSubmit={(event) => {
          event.preventDefault();
          if (target !== null && confirmation === candidate.candidateValue) {
            void onSubmit({ confirmExactValue: confirmation, target });
          }
        }}
      >
        <h3 id="network-flow-indicator-link-title">Link Core Indicator</h3>
        <p>
          Link <strong>{candidate.label}</strong>. Confirm the exact canonical
          IP value; display labels and positions are never used as targets.
        </p>
        <output style={monoTextStyle}>{candidate.candidateValue}</output>
        <fieldset style={dialogFieldsetStyle}>
          <legend>Indicator target</legend>
          <label style={inlineControlStyle}>
            <input
              checked={mode === "create"}
              name="network-flow-indicator-target"
              type="radio"
              onChange={() => setMode("create")}
            />
            Create Indicator
          </label>
          <label style={inlineControlStyle}>
            <input
              checked={mode === "existing"}
              name="network-flow-indicator-target"
              type="radio"
              onChange={() => setMode("existing")}
            />
            Existing Indicator
          </label>
        </fieldset>
        {mode === "existing" ? (
          <label style={fieldLabelStyle}>
            Existing Indicator ID
            <input
              data-testid={networkAnalysisTestId("indicator-link-existing-id")}
              required
              value={existingIndicatorId}
              onChange={(event) =>
                setExistingIndicatorId(event.currentTarget.value)
              }
            />
          </label>
        ) : null}
        <label style={fieldLabelStyle}>
          Confirm exact canonical value
          <input
            data-testid={networkAnalysisTestId("indicator-link-confirmation")}
            required
            value={confirmation}
            onChange={(event) => setConfirmation(event.currentTarget.value)}
          />
        </label>
        <div style={dialogActionsStyle}>
          <button
            data-testid={networkAnalysisTestId("indicator-link-cancel")}
            disabled={linking}
            type="button"
            onClick={onCancel}
          >
            Cancel
          </button>
          <button
            data-testid={networkAnalysisTestId("indicator-link-submit")}
            disabled={
              linking ||
              target === null ||
              confirmation !== candidate.candidateValue
            }
            type="submit"
          >
            {linking ? "Linking…" : "Link Indicator"}
          </button>
        </div>
      </form>
    </div>
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

const dialogFieldsetStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const inlineControlStyle = {
  alignItems: "center",
  display: "inline-flex",
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

const panelGridWithActionsStyle = {
  ...panelGridStyle,
  gridTemplateRows: "auto auto minmax(0, 1fr) auto",
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
  gridTemplateRows: "auto auto minmax(0, 1fr)",
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;

const graphScopeStyle = {
  alignItems: "center",
  background: "var(--ct-colors-surface-1)",
  border: 0,
  borderBlockEnd: "var(--ct-border-hairline)",
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-md)",
  margin: 0,
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
} satisfies CSSProperties;

const graphTableSelectionStyle = {
  alignItems: "center",
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const graphSummaryStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

const graphSectionTitleStyle = {
  background: "var(--ct-colors-surface-2)",
  borderBlockEnd: "var(--ct-border-hairline)",
  fontSize: "0.8125rem",
  margin: 0,
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
} satisfies CSSProperties;

const drawerStyle = {
  position: "absolute",
  insetBlock: 0,
  insetInlineEnd: 0,
  display: "grid",
  gridTemplateRows: "auto auto minmax(0, 1fr) auto",
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

const visuallyHiddenStyle = {
  blockSize: 1,
  clip: "rect(0 0 0 0)",
  clipPath: "inset(50%)",
  inlineSize: 1,
  overflow: "hidden",
  position: "absolute",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const monoTextStyle = {
  fontFamily: "var(--ct-font-family-mono)",
  fontVariantNumeric: "tabular-nums",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
} satisfies CSSProperties;
