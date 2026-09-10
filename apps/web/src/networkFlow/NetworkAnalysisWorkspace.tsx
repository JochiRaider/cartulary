import type {
  GridCellAnchor,
  GridCellRange,
  GridHandle,
} from "@cartulary/grid-adapter";
import {
  networkAnalysisEdgeTestId,
  networkAnalysisTableTabTestId,
  networkAnalysisTestId,
  networkAnalysisVertexTestId,
} from "@cartulary/ui-contracts";
import { Link2, Network, RefreshCw, Table2, Upload, X } from "lucide-react";
import type { ReactNode } from "react";
import {
  type CSSProperties,
  type RefObject,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { IncidentCollaborationBoundary } from "../collaboration/IncidentCollaborationSession";
import { useExtensionAvailabilityController } from "../extensions/ExtensionAvailabilityContext";
import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import {
  NetworkFlowButton,
  NetworkFlowChoice,
  NetworkFlowChromeStyles,
  NetworkFlowField,
  NetworkFlowIconButton,
  NetworkFlowSelect,
  networkFlowChromeRootClassName,
} from "./NetworkFlowControls";
import type { NetworkFlowImportController } from "./NetworkFlowImportController";
import type { NetworkFlowIndicatorLinkController } from "./NetworkFlowIndicatorLinkController";
import {
  NetworkFlowAcceptedQueryControls,
  NetworkFlowRejectedQueryControls,
} from "./NetworkFlowQueryControls";
import { NetworkFlowSavedGraphPanel } from "./NetworkFlowSavedGraphPanel";
import {
  NetworkFlowAcceptedGrid,
  NetworkFlowContributorGrid,
  NetworkFlowRejectedGrid,
} from "./NetworkFlowSemanticGrid";
import type { NetworkFlowTableController } from "./NetworkFlowTableController";
import { TableLifecycleControls } from "./NetworkFlowTableLifecycle";
import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowEdgeAnnotation,
  NetworkFlowGraphEdge,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowGraphVertex,
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
import { networkFlowImportStatus } from "./networkFlowImportState";
import {
  type NetworkFlowRowLinkSelection,
  networkFlowEdgeLinkCandidate,
  networkFlowRowLinkCandidate,
  networkFlowVertexLinkCandidate,
  resolveNetworkFlowRowLinkSelection,
} from "./networkFlowIndicatorLinkModel";
import type {
  NetworkFlowAcceptedQuery,
  NetworkFlowRejectedQuery,
} from "./networkFlowQueryModel";
import { networkFlowWorkspaceStatus } from "./networkFlowWorkspaceStatus";
import type { SavedGraphController } from "./SavedGraphController";
import { useNetworkFlowCollaborationController } from "./useNetworkFlowCollaborationController";
import {
  type NetworkFlowGraphAggregationMode,
  type NetworkFlowGraphBucketWidth,
  type NetworkFlowGraphScopeMode,
  useNetworkFlowGraphController,
} from "./useNetworkFlowGraphController";
import { useNetworkFlowImportController } from "./useNetworkFlowImportController";
import { useNetworkFlowIndicatorLinkController } from "./useNetworkFlowIndicatorLinkController";
import type { NetworkFlowQueryLoadState } from "./useNetworkFlowPagedQuery";
import { useNetworkFlowRejectedRowsController } from "./useNetworkFlowRejectedRowsController";
import { useNetworkFlowRowsController } from "./useNetworkFlowRowsController";
import { useNetworkFlowSavedGraphController } from "./useNetworkFlowSavedGraphController";
import { useNetworkFlowTableController } from "./useNetworkFlowTableController";

type NetworkAnalysisMode = "rows" | "rejected" | "graph";
type NetworkFlowGraphSurface = "explore" | "saved";

export type NetworkAnalysisWorkspaceProps = {
  readonly tableController: NetworkFlowTableController;
  readonly indicatorLinkController: NetworkFlowIndicatorLinkController;
  readonly savedGraphController: SavedGraphController;
  readonly importController: NetworkFlowImportController;
  readonly workbookStatus?: ReactNode;
  readonly apiBase?: string | undefined;
  readonly currentUserId?: string | null | undefined;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
};

const activeTableScopeLabel = networkAnalysisSheetRef();
const graphVertexRenderLimit = 500;
const graphEdgeRenderLimit = 1_000;

function NetworkAnalysisWorkspaceContent({
  tableController: tableOperation,
  indicatorLinkController: indicatorLinkOperation,
  importController: importOperation,
  savedGraphController: savedGraphOperation,
  workbookStatus,
  apiBase,
  currentIncidentRole,
  incidentId,
  onIncidentAccessLost,
}: NetworkAnalysisWorkspaceProps) {
  const extensionAvailability = useExtensionAvailabilityController();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [mode, setMode] = useState<NetworkAnalysisMode>("rows");
  const [graphSurface, setGraphSurface] =
    useState<NetworkFlowGraphSurface>("explore");
  const [message, setMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] =
    useState<NetworkFlowWorkspaceError | null>(null);
  const [protectedStateError, setProtectedStateError] =
    useState<NetworkFlowRequestError | null>(null);
  const [rowGridSelection, setRowGridSelection] = useState<{
    readonly activeAnchor: GridCellAnchor | null;
    readonly cellRange: GridCellRange | null;
  }>({ activeAnchor: null, cellRange: null });
  const failedObservation = useRef(-1);
  const handleWorkspaceError = useCallback(
    (error: NetworkFlowWorkspaceError | null) => {
      setErrorMessage(error);
      if (
        error instanceof NetworkFlowRequestError &&
        isNetworkFlowProtectedStateLoss(error)
      ) {
        failedObservation.current =
          tableOperation.getSnapshot().observationRevision;
        setProtectedStateError(error);
      }
    },
    [tableOperation],
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
  const canLink =
    currentIncidentRole === "editor" || currentIncidentRole === "admin";
  const canDelete =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const canManageSavedGraphs =
    currentIncidentRole === "editor" || currentIncidentRole === "admin";
  const tableController = useNetworkFlowTableController({
    controller: tableOperation,
    enabled: canRead,
  });
  const rowsController = useNetworkFlowRowsController({
    activeTableId: tableController.activeTableId,
    availability: extensionAvailability,
    apiBase,
    enabled: mode === "rows",
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
  });
  const rejectedRowsController = useNetworkFlowRejectedRowsController({
    activeTableId: tableController.activeTableId,
    availability: extensionAvailability,
    apiBase,
    enabled: mode === "rejected",
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
  });
  const graphController = useNetworkFlowGraphController({
    tableLifecycle: tableOperation,
    activeTableId: tableController.activeTableId,
    availability: extensionAvailability,
    apiBase,
    enabled: mode === "graph",
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
    query: rowsController.query,
    tables: tableController.tables,
  });
  const savedGraphController = useNetworkFlowSavedGraphController({
    controller: savedGraphOperation,
    enabled: mode === "graph" && graphSurface === "saved" && canRead,
  });
  const importController = useNetworkFlowImportController({
    controller: importOperation,
    onImported: tableController.handoffImportedTable,
  });
  const clearRows = rowsController.clearRows;
  const clearDiagnostics = rejectedRowsController.clearDiagnostics;
  const clearGraph = graphController.clearGraph;
  const clearResources = useCallback(() => {
    clearRows();
    clearDiagnostics();
    clearGraph();
    setRowGridSelection({ activeAnchor: null, cellRange: null });
    setMode("rows");
    setGraphSurface("explore");
  }, [clearDiagnostics, clearGraph, clearRows]);
  const clearActiveTable = useCallback(() => {
    clearRows();
    clearDiagnostics();
    setRowGridSelection({ activeAnchor: null, cellRange: null });
  }, [clearRows, clearDiagnostics]);
  useNetworkFlowCollaborationController({
    controller: tableOperation,
    activeTableId: tableController.activeTableId,
    clearResources,
    clearActiveTable,
    onMessage: setMessage,
    onProtectedStateLoss: handleWorkspaceError,
  });
  const acceptedGridRef = useRef<GridHandle | null>(null);
  useLayoutEffect(
    () =>
      indicatorLinkOperation.bindFocusRestoration(() => {
        if (
          rowGridSelection.activeAnchor !== null &&
          acceptedGridRef.current?.focusAnchor(
            rowGridSelection.activeAnchor,
          ) === true
        )
          return true;
        if (acceptedGridRef.current !== null) {
          acceptedGridRef.current.focusRoot();
          return true;
        }
        return false;
      }),
    [indicatorLinkOperation, rowGridSelection.activeAnchor],
  );
  const indicatorLinkController = useNetworkFlowIndicatorLinkController({
    controller: indicatorLinkOperation,
    selectionContext: JSON.stringify([
      incidentId,
      mode,
      graphSurface,
      tableController.activeTableId,
      rowsController.query,
      rowGridSelection,
      graphController.graph?.graph_query_digest,
      graphController.selection,
      graphController.aggregationMode,
      graphController.bucketWidthSeconds,
      graphController.scopeMode,
      graphController.selectedTableIds,
      tableController.activeTable?.mapping_fingerprint,
      importController.state.write?.request,
      importController.state.draft,
    ]),
  });
  const effectiveStatus = networkFlowWorkspaceStatus({
    graphStale:
      mode === "graph" &&
      graphSurface === "explore" &&
      graphController.graphStale,
    importStatus: networkFlowImportStatus(importController.state),
    linkStatus: indicatorLinkController.status,
    graphState:
      mode === "graph" && graphSurface === "explore"
        ? graphController.graphLoadState
        : "idle",
    hasGraph:
      mode === "graph" &&
      graphSurface === "explore" &&
      graphController.graph !== null,
    rejectedRows: tableController.activeTable?.row_count_rejected ?? null,
  });
  const rowLinkSelection = useMemo(
    () =>
      resolveNetworkFlowRowLinkSelection({
        activeAnchor: rowGridSelection.activeAnchor,
        bindingSourceRowLimit: indicatorLinkController.sourceLimit,
        cellRange: rowGridSelection.cellRange,
        rows: rowsController.rows,
      }),
    [
      indicatorLinkController.sourceLimit,
      rowGridSelection.activeAnchor,
      rowGridSelection.cellRange,
      rowsController.rows,
    ],
  );
  const handleRowGridSelectionChange = useCallback(
    (activeAnchor: GridCellAnchor | null, cellRange: GridCellRange | null) => {
      setRowGridSelection({ activeAnchor, cellRange });
    },
    [],
  );
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
    if (
      isNetworkFlowAuthorizationLoss(error) ||
      error.code === "incident_closed"
    ) {
      clearResources();
      if (
        error.code === "incident_closed" ||
        error.status === 401 ||
        error.code === "session_required"
      )
        tableOperation.onProtectedFailure(error);
      else tableController.clearAuthorization();
      return;
    }
    if (isNetworkFlowLifecycleLoss(error)) {
      if (mode === "graph") graphController.markGraphStale();
      else clearActiveTable();
      void tableController.loadTables();
    }
  }, [
    clearResources,
    clearActiveTable,
    mode,
    graphController.markGraphStale,
    errorMessage,
    protectedStateError,
    tableOperation,
    tableController.clearAuthorization,
    tableController.error,
    tableController.loadTables,
  ]);
  useEffect(() => {
    if (
      protectedStateError !== null &&
      protectedStateError.code !== "incident_closed" &&
      (isNetworkFlowLifecycleLoss(protectedStateError) ||
        protectedStateError.status === 401) &&
      tableController.loadState === "ready" &&
      tableController.observationRevision > failedObservation.current &&
      (tableController.activeTableId !== null ||
        protectedStateError.status === 401)
    ) {
      setErrorMessage(null);
      setProtectedStateError(null);
    }
  }, [
    protectedStateError,
    tableController.activeTableId,
    tableController.observationRevision,
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
      className={networkFlowChromeRootClassName}
      data-extension-profile-id={activeTableScopeLabel.extension_profile_id}
      data-testid={networkAnalysisTestId("workspace")}
      data-workspace-key={activeTableScopeLabel.workspace_key}
      style={workspaceStyle}
      tabIndex={-1}
    >
      <NetworkFlowChromeStyles />
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
          <TableLifecycleControls controller={tableOperation} />
          <NetworkFlowIconButton
            aria-label="Refresh tables"
            data-testid={networkAnalysisTestId("refresh")}
            title="Refresh"
            onClick={() => {
              void tableController.loadTables();
            }}
          >
            <RefreshCw aria-hidden="true" size={16} />
          </NetworkFlowIconButton>
          <input
            ref={fileInputRef}
            accept=".csv,text/csv"
            data-testid={networkAnalysisTestId("import-input")}
            hidden
            type="file"
            onChange={importController.handleImportChange}
          />
          {canImport || importController.hasWork ? (
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("import-trigger")}
              variant="primary"
              disabled={
                !importController.state.canWrite && !importController.hasWork
              }
              onClick={() =>
                importController.hasWork
                  ? importOperation.setPresented(true)
                  : fileInputRef.current?.click()
              }
            >
              <Upload aria-hidden="true" size={16} />
              {importController.hasWork
                ? "Review current import"
                : "Import NetFlow CSV"}
            </NetworkFlowButton>
          ) : null}
        </div>
      </header>
      <div style={viewBarStyle}>
        <div aria-label="Network Flow tables" role="tablist" style={tabsStyle}>
          {tableController.tables.map((table, index) => {
            const selected =
              table.network_flow_table_id === tableController.activeTableId;
            return (
              <NetworkFlowButton
                key={table.network_flow_table_id}
                aria-controls="network-flow-work-area"
                data-network-flow-table-id={table.network_flow_table_id}
                aria-selected={selected}
                data-testid={networkAnalysisTableTabTestId(
                  table.network_flow_table_id,
                )}
                role="tab"
                selected={selected}
                tabIndex={selected ? 0 : -1}
                variant="mode"
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
                  tableController.selectTable(
                    targetTable.network_flow_table_id,
                  );
                  setMode("rows");
                  event.currentTarget.parentElement
                    ?.querySelectorAll<HTMLButtonElement>('[role="tab"]')
                    [targetIndex]?.focus();
                }}
                onClick={() => {
                  tableController.selectTable(table.network_flow_table_id);
                  setMode("rows");
                }}
              >
                <span className="network-flow-truncate">
                  {table.display_name}
                </span>
                <small style={tabCountStyle}>{index + 1}</small>
              </NetworkFlowButton>
            );
          })}
        </div>
      </div>

      <div style={modeBarStyle}>
        <NetworkFlowButton
          aria-pressed={mode === "rows"}
          data-testid={networkAnalysisTestId("mode-rows")}
          disabled={tableController.activeTable === null}
          selected={mode === "rows"}
          variant="mode"
          onClick={() => setMode("rows")}
        >
          <Table2 aria-hidden="true" size={15} />
          Rows
        </NetworkFlowButton>
        <NetworkFlowButton
          aria-pressed={mode === "rejected"}
          data-testid={networkAnalysisTestId("mode-rejected")}
          disabled={tableController.activeTable === null}
          selected={mode === "rejected"}
          variant="mode"
          onClick={() => setMode("rejected")}
        >
          Rejected
        </NetworkFlowButton>
        <NetworkFlowButton
          aria-pressed={mode === "graph"}
          data-testid={networkAnalysisTestId("mode-graph")}
          disabled={tableController.tables.length === 0}
          selected={mode === "graph"}
          variant="mode"
          onClick={() => setMode("graph")}
        >
          <Network aria-hidden="true" size={15} />
          Graph
        </NetworkFlowButton>
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
          <section
            aria-label="Network Flow graph workspace"
            style={graphWorkspaceStyle}
          >
            <fieldset style={graphSurfaceFieldsetStyle}>
              <legend style={visuallyHiddenStyle}>Graph workspace mode</legend>
              <NetworkFlowButton
                aria-pressed={graphSurface === "explore"}
                data-testid={networkAnalysisTestId("graph-surface-explore")}
                selected={graphSurface === "explore"}
                variant="mode"
                onClick={() => setGraphSurface("explore")}
              >
                Unsaved exploration
              </NetworkFlowButton>
              <NetworkFlowButton
                aria-pressed={graphSurface === "saved"}
                data-testid={networkAnalysisTestId("graph-surface-saved")}
                selected={graphSurface === "saved"}
                variant="mode"
                onClick={() => setGraphSurface("saved")}
              >
                Saved graphs
              </NetworkFlowButton>
            </fieldset>
            {graphSurface === "saved" ? (
              <NetworkFlowSavedGraphPanel
                canCreate={canManageSavedGraphs}
                canRetire={canDelete}
                controller={savedGraphController}
                tables={tableController.tables}
                currentGraph={graphController.graph}
              />
            ) : (
              <GraphPanel
                aggregationMode={graphController.aggregationMode}
                bucketWidthSeconds={graphController.bucketWidthSeconds}
                canLink={canLink}
                canNextContributorPage={graphController.canNextContributorPage}
                canPreviousContributorPage={
                  graphController.canPreviousContributorPage
                }
                contributorLoadState={graphController.contributorLoadState}
                contributorLoadGenerationKey={
                  graphController.contributorLoadGenerationKey
                }
                contributorError={graphController.contributorError}
                contributorPageNumber={graphController.contributorPageNumber}
                contributors={graphController.contributors}
                graph={graphController.graph}
                graphLoadState={graphController.graphLoadState}
                scopeMode={graphController.scopeMode}
                selectedEdge={graphController.selectedEdge}
                selectedTableIds={graphController.selectedTableIds}
                selectedVertex={graphController.selectedVertex}
                tables={tableController.tables}
                validationMessage={graphController.validationMessage}
                onCloseDrawer={() => graphController.selectGraphObject(null)}
                onLinkEdge={(fieldKey) => {
                  const candidate = networkFlowEdgeLinkCandidate({
                    edge: graphController.selectedEdge,
                    fieldKey,
                    graph: graphController.graph,
                  });
                  if (candidate !== null)
                    indicatorLinkOperation.openDraft(candidate);
                }}
                onLinkVertex={() => {
                  const candidate = networkFlowVertexLinkCandidate(
                    graphController.graph,
                    graphController.selectedVertex,
                  );
                  if (candidate !== null)
                    indicatorLinkOperation.openDraft(candidate);
                }}
                onNextContributorPage={graphController.nextContributorPage}
                onPreviousContributorPage={
                  graphController.previousContributorPage
                }
                onRefreshGraph={graphController.refreshGraph}
                onRetryContributors={graphController.retryContributorPage}
                onAggregationModeChange={graphController.setAggregationMode}
                onBucketWidthChange={graphController.setBucketWidthSeconds}
                onScopeModeChange={graphController.setScopeMode}
                onSelectEdge={graphController.selectGraphObject}
                onSelectTable={graphController.setTableSelected}
                onSelectVertex={graphController.selectGraphObject}
              />
            )}
          </section>
        ) : mode === "rejected" ? (
          <RejectedRowsPanel
            activeTable={tableController.activeTable}
            canNext={rejectedRowsController.canNext}
            canPrevious={rejectedRowsController.canPrevious}
            diagnostics={rejectedRowsController.diagnostics}
            error={rejectedRowsController.error}
            loadGenerationKey={rejectedRowsController.loadGenerationKey}
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
            gridRef={acceptedGridRef}
            linkLimitError={indicatorLinkController.limitError}
            onRetryLinkLimit={indicatorLinkOperation.loadLimit}
            activeTable={tableController.activeTable}
            canLink={canLink}
            canNext={rowsController.canNext}
            canPrevious={rowsController.canPrevious}
            loadGenerationKey={rowsController.loadGenerationKey}
            loadState={rowsController.loadState}
            error={rowsController.error}
            notice={rowsController.notice}
            pageNumber={rowsController.pageNumber}
            query={rowsController.query}
            rows={rowsController.rows}
            rowLinkSelection={rowLinkSelection}
            onBeginLink={() => {
              if (rowLinkSelection !== null)
                indicatorLinkOperation.openDraft(
                  networkFlowRowLinkCandidate(rowLinkSelection),
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
        {workbookStatus}
        <span
          style={{
            minWidth: 0,
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
          aria-atomic="true"
          aria-live="polite"
          role="status"
        >
          <span>
            {tableController.loadState === "loading"
              ? "Loading"
              : tableController.loadState === "refreshing"
                ? "Refreshing table metadata"
                : `${tableController.tables.length} active table${
                    tableController.tables.length === 1 ? "" : "s"
                  }`}
          </span>
          {tableController.notice || message ? (
            <span data-testid={networkAnalysisTestId("stale-state")}>
              {tableController.notice ?? message}
              {tableController.operation?.status === "acknowledged" &&
              tableController.loadState === "error"
                ? " Completed; current table metadata could not be refreshed."
                : null}
            </span>
          ) : null}
          {effectiveStatus !== null ? (
            <span data-network-flow-state={effectiveStatus}>
              {" · "}
              {effectiveStatus.replaceAll("_", " ")}
              {importController.hasWork &&
              networkFlowImportStatus(importController.state) !== null
                ? ` · ${importController.state.message}`
                : ""}
            </span>
          ) : null}
        </span>
        {visibleError ? (
          <span aria-atomic="true" role="alert" style={errorTextStyle}>
            Error: {networkFlowErrorMessage(visibleError)}
          </span>
        ) : null}
      </div>
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
        <NetworkFlowButton variant="secondary" onClick={onRetry}>
          Retry
        </NetworkFlowButton>
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
        <NetworkFlowButton
          pending={importing}
          variant="primary"
          onClick={onImport}
        >
          <Upload aria-hidden="true" size={16} />
          Import NetFlow CSV
        </NetworkFlowButton>
      ) : (
        <span style={mutedTextStyle}>No active Network Flow tables.</span>
      )}
    </section>
  );
}

function RowsPanel({
  gridRef,
  linkLimitError,
  onRetryLinkLimit,
  activeTable,
  canLink,
  canNext,
  canPrevious,
  error,
  loadState,
  loadGenerationKey,
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
  readonly gridRef: RefObject<GridHandle | null>;
  readonly linkLimitError: string | null;
  readonly onRetryLinkLimit: () => void;
  readonly activeTable: NetworkFlowTable | null;
  readonly canLink: boolean;
  readonly canNext: boolean;
  readonly canPrevious: boolean;
  readonly error: NetworkFlowRequestError | null;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly loadGenerationKey: string | number;
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
          {linkLimitError !== null ? (
            <>
              <span role="alert">{linkLimitError}</span>
              <NetworkFlowButton onClick={onRetryLinkLimit}>
                Retry link limits
              </NetworkFlowButton>
            </>
          ) : null}
          <NetworkFlowButton
            disabled={rowLinkSelection === null}
            variant="secondary"
            onClick={onBeginLink}
          >
            <Link2 aria-hidden="true" size={15} />
            {rowLinkSelection === null
              ? "Select one IP cell or same-value IP range"
              : `Link ${rowLinkSelection.rows.length} selected row${
                  rowLinkSelection.rows.length === 1 ? "" : "s"
                }`}
          </NetworkFlowButton>
        </div>
      ) : null}
      <NetworkFlowAcceptedGrid
        gridRef={gridRef}
        error={error}
        filtered={query.filters.length > 0 || query.timeWindow !== null}
        loadGenerationKey={loadGenerationKey}
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
  error,
  loadState,
  loadGenerationKey,
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
  readonly error: NetworkFlowRequestError | null;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly loadGenerationKey: string | number;
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
        error={error}
        filtered={
          query.errorCodes.length > 0 ||
          query.fieldKeys.length > 0 ||
          query.sourceRowRange !== null
        }
        loadState={loadState}
        loadGenerationKey={loadGenerationKey}
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
    <nav
      aria-label="Network Flow result pages"
      className="network-flow-pagination"
      style={paginationStyle}
    >
      <NetworkFlowButton
        data-testid={networkAnalysisTestId("page-previous")}
        disabled={!canPrevious || loading}
        variant="secondary"
        onClick={onPrevious}
      >
        Previous
      </NetworkFlowButton>
      <span
        aria-live="polite"
        data-testid={networkAnalysisTestId("page-status")}
      >
        {loading ? "Loading page" : `Page ${pageNumber}`}
      </span>
      <NetworkFlowButton
        data-testid={networkAnalysisTestId("page-next")}
        disabled={!canNext || loading}
        variant="secondary"
        onClick={onNext}
      >
        Next
      </NetworkFlowButton>
      {notice === null ? null : <span role="status">{notice}</span>}
    </nav>
  );
}

function GraphPanel({
  aggregationMode,
  bucketWidthSeconds,
  canLink,
  canNextContributorPage,
  canPreviousContributorPage,
  contributorLoadState,
  contributorLoadGenerationKey,
  contributorError,
  contributorPageNumber,
  contributors,
  graph,
  graphLoadState,
  scopeMode,
  selectedEdge,
  selectedTableIds,
  selectedVertex,
  tables,
  validationMessage,
  onCloseDrawer,
  onLinkEdge,
  onLinkVertex,
  onNextContributorPage,
  onPreviousContributorPage,
  onRefreshGraph,
  onRetryContributors,
  onAggregationModeChange,
  onBucketWidthChange,
  onScopeModeChange,
  onSelectEdge,
  onSelectTable,
  onSelectVertex,
}: {
  readonly aggregationMode: NetworkFlowGraphAggregationMode;
  readonly bucketWidthSeconds: NetworkFlowGraphBucketWidth;
  readonly canLink: boolean;
  readonly canNextContributorPage: boolean;
  readonly canPreviousContributorPage: boolean;
  readonly contributorLoadState: NetworkFlowQueryLoadState;
  readonly contributorLoadGenerationKey: string | number;
  readonly contributorError: NetworkFlowRequestError | null;
  readonly contributorPageNumber: number;
  readonly contributors: readonly NetworkFlowContributor[];
  readonly graph: NetworkFlowGraphResult | null;
  readonly graphLoadState: NetworkFlowQueryLoadState;
  readonly scopeMode: NetworkFlowGraphScopeMode;
  readonly selectedEdge: NetworkFlowGraphEdge | null;
  readonly selectedTableIds: readonly string[];
  readonly selectedVertex: NetworkFlowGraphVertex | null;
  readonly tables: readonly NetworkFlowTable[];
  readonly validationMessage: string | null;
  readonly onCloseDrawer: () => void;
  readonly onLinkEdge: (
    fieldKey: "network_flow.src_ip" | "network_flow.dst_ip",
  ) => void;
  readonly onLinkVertex: () => void;
  readonly onNextContributorPage: () => void;
  readonly onPreviousContributorPage: () => void;
  readonly onRefreshGraph: () => void;
  readonly onRetryContributors: () => void;
  readonly onAggregationModeChange: (
    mode: NetworkFlowGraphAggregationMode,
  ) => void;
  readonly onBucketWidthChange: (width: NetworkFlowGraphBucketWidth) => void;
  readonly onScopeModeChange: (mode: NetworkFlowGraphScopeMode) => void;
  readonly onSelectEdge: (selector: NetworkFlowGraphSelector) => void;
  readonly onSelectTable: (tableId: string, selected: boolean) => void;
  readonly onSelectVertex: (selector: NetworkFlowGraphSelector) => void;
}) {
  const selectedGraphButtonRef = useRef<HTMLButtonElement | null>(null);
  const [vertexPage, setVertexPage] = useState(0);
  const [edgePage, setEdgePage] = useState(0);
  const [bucketIndex, setBucketIndex] = useState(0);
  const selectedObject = selectedVertex ?? selectedEdge;
  const timeBuckets =
    graph?.result_variant.kind === "time_bucket_v1"
      ? graph.result_variant.time_buckets
      : [];
  const selectedBucket = timeBuckets[bucketIndex] ?? null;
  const visibleTemporalEdgeIDs = useMemo(
    () =>
      new Set(
        graph?.edge_annotations.flatMap((annotation) =>
          annotation.selector.kind === "time_bucket_edge" &&
          selectedBucket !== null &&
          annotation.selector.bucket_start_utc === selectedBucket.start_utc &&
          annotation.selector.bucket_end_utc === selectedBucket.end_utc
            ? [annotation.projected_edge_id]
            : [],
        ) ?? [],
      ),
    [graph, selectedBucket],
  );
  const visibleTemporalVertexIDs = useMemo(() => {
    const ids = new Set<string>();
    if (selectedBucket === null) return ids;
    for (const edge of graph?.graph_projection_result.edges ?? []) {
      if (visibleTemporalEdgeIDs.has(edge.edge_id)) {
        ids.add(edge.src_vertex_id);
        ids.add(edge.dst_vertex_id);
      }
    }
    return ids;
  }, [graph, selectedBucket, visibleTemporalEdgeIDs]);
  const graphAllVertices = useMemo(
    () =>
      [...(graph?.graph_projection_result.vertices ?? [])]
        .filter(
          (vertex) =>
            selectedBucket === null ||
            visibleTemporalVertexIDs.has(vertex.vertex_id),
        )
        .sort((left, right) =>
          graphVertexLabel(left).localeCompare(graphVertexLabel(right)),
        ),
    [graph, selectedBucket, visibleTemporalVertexIDs],
  );
  const graphVertices = graphAllVertices.slice(
    vertexPage * graphVertexRenderLimit,
    (vertexPage + 1) * graphVertexRenderLimit,
  );
  const graphEndpointLabels = useMemo(
    () =>
      new Map(
        graphAllVertices.flatMap((vertex) => {
          const endpointId = semanticGraphVertexId(graph, vertex);
          return endpointId === null
            ? []
            : ([[endpointId, graphVertexLabel(vertex)]] as const);
        }),
      ),
    [graph, graphAllVertices],
  );
  const graphAllEdges = useMemo(
    () =>
      [...(graph?.graph_projection_result.edges ?? [])]
        .filter(
          (edge) =>
            selectedBucket === null || visibleTemporalEdgeIDs.has(edge.edge_id),
        )
        .sort((left, right) =>
          graphEdgeLabel(left, graphEndpointLabels).localeCompare(
            graphEdgeLabel(right, graphEndpointLabels),
          ),
        ),
    [graph, graphEndpointLabels, selectedBucket, visibleTemporalEdgeIDs],
  );
  const graphEdges = graphAllEdges.slice(
    edgePage * graphEdgeRenderLimit,
    (edgePage + 1) * graphEdgeRenderLimit,
  );
  // biome-ignore lint/correctness/useExhaustiveDependencies: immutable result identity resets bounded navigation.
  useEffect(() => {
    setVertexPage(0);
    setEdgePage(0);
    setBucketIndex(0);
  }, [graph?.graph_projection_result.projection_result_id]);
  // biome-ignore lint/correctness/useExhaustiveDependencies: bucket navigation resets bounded object pages.
  useEffect(() => {
    setVertexPage(0);
    setEdgePage(0);
  }, [bucketIndex]);
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
            <label
              key={value}
              htmlFor={`network-flow-graph-scope-${value}`}
              style={inlineControlStyle}
            >
              <NetworkFlowChoice
                checked={scopeMode === value}
                id={`network-flow-graph-scope-${value}`}
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
                    htmlFor={`network-flow-graph-table-${table.network_flow_table_id}`}
                    style={inlineControlStyle}
                  >
                    <NetworkFlowChoice
                      checked={checked}
                      disabled={checked && selectedTableIds.length === 1}
                      id={`network-flow-graph-table-${table.network_flow_table_id}`}
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
        <fieldset style={graphScopeStyle}>
          <legend>Graph aggregation</legend>
          <label
            htmlFor="network-flow-graph-aggregation-default"
            style={inlineControlStyle}
          >
            <NetworkFlowChoice
              checked={aggregationMode === "default_flow_edge_v1"}
              id="network-flow-graph-aggregation-default"
              name="network-flow-graph-aggregation"
              type="radio"
              onChange={() => onAggregationModeChange("default_flow_edge_v1")}
            />
            Default flow edges
          </label>
          <label
            htmlFor="network-flow-graph-aggregation-time"
            style={inlineControlStyle}
          >
            <NetworkFlowChoice
              checked={aggregationMode === "time_bucket_v1"}
              id="network-flow-graph-aggregation-time"
              name="network-flow-graph-aggregation"
              type="radio"
              onChange={() => onAggregationModeChange("time_bucket_v1")}
            />
            Time buckets
          </label>
          {aggregationMode === "time_bucket_v1" ? (
            <NetworkFlowField
              htmlFor="network-flow-graph-bucket-width"
              label="Bucket width"
            >
              <NetworkFlowSelect
                aria-label="Bucket width"
                id="network-flow-graph-bucket-width"
                value={bucketWidthSeconds}
                onChange={(event) =>
                  onBucketWidthChange(
                    Number(
                      event.currentTarget.value,
                    ) as NetworkFlowGraphBucketWidth,
                  )
                }
              >
                <option value={60}>1 minute</option>
                <option value={300}>5 minutes</option>
                <option value={900}>15 minutes</option>
                <option value={3600}>1 hour</option>
                <option value={21600}>6 hours</option>
                <option value={86400}>1 day</option>
              </NetworkFlowSelect>
            </NetworkFlowField>
          ) : null}
          {validationMessage === null ? null : (
            <p className="network-flow-status" data-tone="error" role="alert">
              {validationMessage}
            </p>
          )}
        </fieldset>
        {selectedBucket === null ? null : (
          <nav
            aria-label="Time bucket navigation"
            style={boundedNavigationStyle}
          >
            <NetworkFlowButton
              disabled={bucketIndex === 0}
              variant="secondary"
              onClick={() => setBucketIndex((current) => current - 1)}
            >
              Previous bucket
            </NetworkFlowButton>
            <strong>
              Bucket {bucketIndex + 1} of {timeBuckets.length}
            </strong>
            <span>
              [{selectedBucket.start_utc}, {selectedBucket.end_utc}) ·{" "}
              {selectedBucket.unique_vertex_count} vertices ·{" "}
              {selectedBucket.edge_count} edges ·{" "}
              {selectedBucket.contributing_row_count} rows
            </span>
            <NetworkFlowButton
              disabled={bucketIndex + 1 >= timeBuckets.length}
              variant="secondary"
              onClick={() => setBucketIndex((current) => current + 1)}
            >
              Next bucket
            </NetworkFlowButton>
          </nav>
        )}
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
          {graphAllVertices.length > graphVertexRenderLimit ||
          graphAllEdges.length > graphEdgeRenderLimit ? (
            <span style={mutedTextStyle}>
              Large results are paged: at most {graphVertexRenderLimit} vertices
              and {graphEdgeRenderLimit} edges are mounted at once.
            </span>
          ) : null}
          {graphAllVertices.length > graphVertexRenderLimit ? (
            <BoundedGraphNavigation
              itemLabel="vertices"
              page={vertexPage}
              pageSize={graphVertexRenderLimit}
              total={graphAllVertices.length}
              onPageChange={setVertexPage}
            />
          ) : null}
          {graphAllEdges.length > graphEdgeRenderLimit ? (
            <BoundedGraphNavigation
              itemLabel="edges"
              page={edgePage}
              pageSize={graphEdgeRenderLimit}
              total={graphAllEdges.length}
              onPageChange={setEdgePage}
            />
          ) : null}
          {graphLoadState === "error" ? (
            <NetworkFlowButton variant="secondary" onClick={onRefreshGraph}>
              <RefreshCw aria-hidden="true" size={14} /> Retry
            </NetworkFlowButton>
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
                const selector = graphVertexSelector(graph, vertex);
                if (selector === null) {
                  return null;
                }
                return (
                  <tr
                    key={vertex.vertex_id}
                    data-testid={networkAnalysisVertexTestId(
                      selector.source_vertex_id,
                    )}
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
                      <NetworkFlowButton
                        ref={
                          selectedVertex === vertex
                            ? selectedGraphButtonRef
                            : undefined
                        }
                        aria-pressed={selectedVertex === vertex}
                        selected={selectedVertex === vertex}
                        variant="mode"
                        onClick={() => onSelectVertex(selector)}
                      >
                        Select vertex
                      </NetworkFlowButton>
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
                const annotation = graphEdgeAnnotation(graph, edge);
                if (annotation === null) {
                  return null;
                }
                const edgeId = annotation.selector.source_edge_id;
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
                      {annotation.example_refs_total_count ??
                        graphScalar(edge.properties.flow_row_count)}
                    </td>
                    <td style={tdStyle}>
                      <NetworkFlowButton
                        ref={
                          selectedEdge === edge
                            ? selectedGraphButtonRef
                            : undefined
                        }
                        aria-pressed={selectedEdge === edge}
                        selected={selectedEdge === edge}
                        variant="mode"
                        onClick={() => onSelectEdge(annotation.selector)}
                      >
                        Select edge
                      </NetworkFlowButton>
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
            <NetworkFlowIconButton
              aria-label="Close graph contributors"
              data-testid={networkAnalysisTestId("contributor-close")}
              title="Close"
              onClick={() => {
                const returnTarget = selectedGraphButtonRef.current;
                onCloseDrawer();
                queueMicrotask(() => returnTarget?.focus());
              }}
            >
              <X aria-hidden="true" size={15} />
            </NetworkFlowIconButton>
          </div>
          <div style={linkActionsStyle}>
            {canLink &&
            graphLoadState === "ready" &&
            networkFlowVertexLinkCandidate(graph, selectedVertex) !== null ? (
              <NetworkFlowButton variant="secondary" onClick={onLinkVertex}>
                <Link2 aria-hidden="true" size={15} />
                Link vertex
              </NetworkFlowButton>
            ) : null}
            {(["network_flow.src_ip", "network_flow.dst_ip"] as const).map(
              (fieldKey) =>
                canLink &&
                graphLoadState === "ready" &&
                networkFlowEdgeLinkCandidate({
                  graph,
                  edge: selectedEdge,
                  fieldKey,
                }) !== null ? (
                  <NetworkFlowButton
                    key={fieldKey}
                    variant="secondary"
                    onClick={() => onLinkEdge(fieldKey)}
                  >
                    <Link2 aria-hidden="true" size={15} />
                    {fieldKey === "network_flow.src_ip"
                      ? "Link source"
                      : "Link destination"}
                  </NetworkFlowButton>
                ) : null,
            )}
          </div>
          <NetworkFlowContributorGrid
            contributors={contributors}
            error={contributorError}
            loadGenerationKey={contributorLoadGenerationKey}
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

function BoundedGraphNavigation({
  itemLabel,
  onPageChange,
  page,
  pageSize,
  total,
}: {
  readonly itemLabel: string;
  readonly onPageChange: (page: number) => void;
  readonly page: number;
  readonly pageSize: number;
  readonly total: number;
}) {
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  return (
    <nav aria-label={`${itemLabel} navigation`} style={boundedNavigationStyle}>
      <span>
        {itemLabel} {page + 1}/{pageCount}
      </span>
      <NetworkFlowButton
        disabled={page === 0}
        variant="secondary"
        onClick={() => onPageChange(page - 1)}
      >
        Previous
      </NetworkFlowButton>
      <NetworkFlowButton
        disabled={page + 1 >= pageCount}
        variant="secondary"
        onClick={() => onPageChange(page + 1)}
      >
        Next
      </NetworkFlowButton>
    </nav>
  );
}

function graphVertexSelector(
  graph: NetworkFlowGraphResult | null,
  vertex: NetworkFlowGraphVertex,
): Extract<NetworkFlowGraphSelector, { readonly kind: "vertex" }> | null {
  return (
    graph?.vertex_selectors.find(
      (binding) => binding.projected_vertex_id === vertex.vertex_id,
    )?.selector ?? null
  );
}

function graphEdgeAnnotation(
  graph: NetworkFlowGraphResult | null,
  edge: NetworkFlowGraphEdge,
): NetworkFlowEdgeAnnotation | null {
  return (
    graph?.edge_annotations.find(
      (annotation) => annotation.projected_edge_id === edge.edge_id,
    ) ?? null
  );
}

function semanticGraphVertexId(
  graph: NetworkFlowGraphResult | null,
  vertex: NetworkFlowGraphVertex,
): string | null {
  return graphVertexSelector(graph, vertex)?.source_vertex_id ?? null;
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

const workspaceStyle = {
  display: "grid",
  gridTemplateRows:
    "auto var(--ct-layout-viewBarHeight) auto auto minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
  background: "var(--ct-colors-canvas)",
  color: "var(--ct-colors-ink)",
} satisfies CSSProperties;

const queryBandsStyle = {
  minWidth: 0,
} satisfies CSSProperties;

const workspaceHeaderStyle = {
  alignItems: "center",
  background: "var(--ct-colors-surface-1)",
  borderBlockEnd: "var(--ct-border-hairline)",
  display: "flex",
  flexWrap: "wrap",
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
  overflowWrap: "anywhere",
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

const tabCountStyle = {
  fontVariantNumeric: "tabular-nums",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const viewActionsStyle = {
  display: "flex",
  alignItems: "center",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
  justifyContent: "flex-end",
  minWidth: 0,
} satisfies CSSProperties;

const inlineControlStyle = {
  alignItems: "center",
  display: "inline-flex",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;

const modeBarStyle = {
  display: "flex",
  flexWrap: "wrap",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-md)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

const workAreaStyle = {
  minBlockSize: 0,
  minWidth: 0,
  overflow: "hidden",
} satisfies CSSProperties;

const graphWorkspaceStyle = {
  blockSize: "100%",
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  minBlockSize: 0,
  minWidth: 0,
  overflow: "auto",
} satisfies CSSProperties;

const graphSurfaceFieldsetStyle = {
  ...modeBarStyle,
  border: 0,
  margin: 0,
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
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 600,
} satisfies CSSProperties;

const tdStyle = {
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
  verticalAlign: "top",
} satisfies CSSProperties;

const tdMonoStyle = {
  ...tdStyle,
  fontFamily: "var(--ct-typography-mono-fontFamily)",
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
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

const boundedNavigationStyle = {
  alignItems: "center",
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
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
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.75rem",
  overflowX: "auto",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const mutedTextStyle = {
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const errorTextStyle = {
  color: "var(--ct-colors-semantic-destructive)",
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
