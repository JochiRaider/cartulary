import type {
  GridCellAnchor,
  GridCellRange,
  GridHandle,
} from "@cartulary/grid-adapter";
import {
  networkAnalysisTableTabTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import { Link2, Network, RefreshCw, Table2, Upload } from "lucide-react";
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
  NetworkFlowChromeStyles,
  NetworkFlowIconButton,
  networkFlowChromeRootClassName,
} from "./NetworkFlowControls";
import { NetworkFlowExplorationPanel } from "./NetworkFlowExplorationPanel";
import type { NetworkFlowImportController } from "./NetworkFlowImportController";
import type { NetworkFlowIndicatorLinkController } from "./NetworkFlowIndicatorLinkController";
import {
  NetworkFlowAcceptedQueryControls,
  NetworkFlowGraphQueryControls,
  NetworkFlowRejectedQueryControls,
} from "./NetworkFlowQueryControls";
import {
  NetworkFlowQueryPagination,
  pageFailureFeedback,
} from "./NetworkFlowQueryPagination";
import { NetworkFlowSavedGraphPanel } from "./NetworkFlowSavedGraphPanel";
import {
  NetworkFlowAcceptedGrid,
  NetworkFlowRejectedGrid,
} from "./NetworkFlowSemanticGrid";
import type { NetworkFlowTableController } from "./NetworkFlowTableController";
import { TableLifecycleControls } from "./NetworkFlowTableLifecycle";
import type {
  NetworkFlowDiagnostic,
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
import { useNetworkFlowGraphController } from "./useNetworkFlowGraphController";
import { useNetworkFlowImportController } from "./useNetworkFlowImportController";
import { useNetworkFlowIndicatorLinkController } from "./useNetworkFlowIndicatorLinkController";
import type {
  NetworkFlowPageNavigation,
  NetworkFlowQueryLoadState,
} from "./useNetworkFlowPagedQuery";
import { useNetworkFlowQueryAuthoring } from "./useNetworkFlowQueryAuthoring";
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
  const queryAuthoring = useNetworkFlowQueryAuthoring({
    contextKey:
      protectedStateError !== null &&
      isNetworkFlowAuthorizationLoss(protectedStateError)
        ? null
        : tableOperation.queryContextIdentity(),
    activeTableId: tableController.activeTableId,
    mode,
    tables: tableController.tables,
  });
  const readIdentity = tableOperation.readContextIdentity();
  const isCurrentRead = useCallback(
    () =>
      readIdentity !== null &&
      tableOperation.readContextIdentity() === readIdentity,
    [readIdentity, tableOperation],
  );
  const onQueryProtectedStateLoss = useCallback(
    (error: NetworkFlowRequestError) => {
      handleWorkspaceError(error);
      tableOperation.onQueryFailure(
        error,
        tableOperation.getSnapshot().activeTableId ?? undefined,
      );
    },
    [handleWorkspaceError, tableOperation],
  );
  const rowsController = useNetworkFlowRowsController({
    readIdentity,
    isCurrentRead,
    onProtectedStateLoss: onQueryProtectedStateLoss,
    query: queryAuthoring.acceptedQuery,
    revision: queryAuthoring.acceptedRevision,
    onQueryResult: queryAuthoring.acceptedResult,
    activeTableId: tableController.activeTableId,
    availability: extensionAvailability,
    apiBase,
    enabled: mode === "rows" && queryAuthoring.contextKey !== null,
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
  });
  const rejectedRowsController = useNetworkFlowRejectedRowsController({
    readIdentity,
    isCurrentRead,
    onProtectedStateLoss: onQueryProtectedStateLoss,
    query: queryAuthoring.rejectedQuery,
    revision: queryAuthoring.rejectedRevision,
    onQueryResult: queryAuthoring.rejectedResult,
    activeTableId: tableController.activeTableId,
    availability: extensionAvailability,
    apiBase,
    enabled: mode === "rejected" && queryAuthoring.contextKey !== null,
    incidentId,
    onError: handleWorkspaceError,
    onIncidentAccessLost,
  });
  const graphSelectionChanged = useRef<(context: string) => void>(() => {});
  const graphController = useNetworkFlowGraphController({
    active: graphSurface === "explore" && mode === "graph",
    onNavigationContextChange: (context) =>
      graphSelectionChanged.current(context),
    readIdentity,
    isCurrentRead,
    onProtectedStateLoss: onQueryProtectedStateLoss,
    settings: queryAuthoring.graphSettings,
    applicationRevision: queryAuthoring.graphApplicationRevision,
    revision: queryAuthoring.acceptedRevision,
    onQueryResult: queryAuthoring.acceptedResult,
    tableLifecycle: tableOperation,
    activeTableId: tableController.activeTableId,
    availability: extensionAvailability,
    apiBase,
    enabled: mode === "graph" && queryAuthoring.contextKey !== null,
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
  const purgeQueries = queryAuthoring.purge;
  const clearResources = useCallback(() => {
    purgeQueries();
    clearRows();
    clearDiagnostics();
    clearGraph();
    setRowGridSelection({ activeAnchor: null, cellRange: null });
    setMode("rows");
    setGraphSurface("explore");
  }, [clearDiagnostics, clearGraph, clearRows, purgeQueries]);
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
  const graphFocusRestorer = useRef<(() => boolean) | null>(null);
  const bindGraphFocusRestoration = useCallback((restore: () => boolean) => {
    graphFocusRestorer.current = restore;
    return () => {
      if (graphFocusRestorer.current === restore)
        graphFocusRestorer.current = null;
    };
  }, []);
  useLayoutEffect(
    () =>
      indicatorLinkOperation.bindFocusRestoration(() => {
        if (mode === "graph" && graphSurface === "explore")
          return graphFocusRestorer.current?.() ?? false;
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
    [indicatorLinkOperation, rowGridSelection.activeAnchor, mode, graphSurface],
  );
  const workspaceSelectionContext = (
    graphContext: string,
    surface = graphSurface,
  ) =>
    JSON.stringify([
      incidentId,
      mode,
      surface,
      tableController.activeTableId,
      rowsController.query,
      rowGridSelection,
      graphContext,
      graphController.aggregationMode,
      graphController.bucketWidthSeconds,
      graphController.scopeMode,
      graphController.selectedTableIds,
      tableController.activeTable?.mapping_fingerprint,
      importController.state.write?.request,
      importController.state.draft,
    ]);
  graphSelectionChanged.current = (context) =>
    indicatorLinkOperation.setSelectionContext(
      workspaceSelectionContext(context),
    );
  const changeGraphSurface = (surface: NetworkFlowGraphSurface) => {
    graphController.setActive(surface === "explore");
    indicatorLinkOperation.setSelectionContext(
      workspaceSelectionContext(graphController.selectionContext, surface),
    );
    setGraphSurface(surface);
    setErrorMessage(null);
  };
  const indicatorLinkController = useNetworkFlowIndicatorLinkController({
    controller: indicatorLinkOperation,
    selectionContext: workspaceSelectionContext(
      graphController.selectionContext,
    ),
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
                draft={queryAuthoring.rejectedDraft}
                onDraftChange={queryAuthoring.setRejectedDraft}
                onApply={queryAuthoring.applyRejected}
                onClear={queryAuthoring.clearRejected}
                appliedQuery={queryAuthoring.appliedRejected}
                issues={queryAuthoring.rejectedIssues}
                status={queryAuthoring.rejectedStatus}
              />
            ) : mode === "graph" && graphSurface === "saved" ? (
              <span>
                Saved graphs use their saved queries. The unsaved exploration
                draft is retained.
              </span>
            ) : (
              <NetworkFlowAcceptedQueryControls
                graphMode={mode === "graph"}
                graphDirty={queryAuthoring.graphDirty}
                temporal={
                  mode === "graph" &&
                  queryAuthoring.graphDraft.aggregation.mode ===
                    "time_bucket_v1"
                }
                tableLabel={tableController.activeTable.display_name}
                draft={queryAuthoring.acceptedDraft}
                onDraftChange={queryAuthoring.setAcceptedDraft}
                onApply={queryAuthoring.applyAccepted}
                onClear={queryAuthoring.clearAccepted}
                appliedQuery={queryAuthoring.appliedAccepted}
                issues={queryAuthoring.acceptedIssues}
                status={queryAuthoring.acceptedStatus}
                graphControls={
                  <NetworkFlowGraphQueryControls
                    draft={queryAuthoring.graphDraft}
                    applied={queryAuthoring.appliedGraph}
                    onChange={queryAuthoring.setGraphDraft}
                    tables={tableController.tables}
                    activeTableId={tableController.activeTableId}
                    issues={queryAuthoring.acceptedIssues}
                  />
                }
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
                onClick={() => changeGraphSurface("explore")}
              >
                Unsaved exploration
              </NetworkFlowButton>
              <NetworkFlowButton
                aria-pressed={graphSurface === "saved"}
                data-testid={networkAnalysisTestId("graph-surface-saved")}
                selected={graphSurface === "saved"}
                variant="mode"
                onClick={() => changeGraphSurface("saved")}
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
                currentGraph={
                  queryAuthoring.acceptedStatus === "applied"
                    ? graphController.graph
                    : null
                }
              />
            ) : (
              <NetworkFlowExplorationPanel
                canLink={canLink && queryAuthoring.acceptedStatus === "applied"}
                contributorPage={graphController.contributorPage}
                navigation={graphController.navigation}
                status={{
                  loadState: graphController.graphLoadState,
                  stale: graphController.graphStale,
                  validationMessage: graphController.validationMessage,
                }}
                tables={tableController.tables}
                onNavigate={graphController.navigate}
                isFocusCurrent={graphController.isFocusCurrent}
                bindFocusRestoration={bindGraphFocusRestoration}
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
                onRefreshGraph={graphController.refreshGraph}
              />
            )}
          </section>
        ) : mode === "rejected" ? (
          <RejectedRowsPanel
            activeTable={tableController.activeTable}
            diagnostics={rejectedRowsController.diagnostics}
            error={rejectedRowsController.error}
            loadGenerationKey={rejectedRowsController.loadGenerationKey}
            loadState={rejectedRowsController.loadState}
            query={rejectedRowsController.query}
            onResetQuery={queryAuthoring.clearRejected}
            page={rejectedRowsController}
            onRefreshResource={() => {
              void tableController.loadTables().then((refreshed) => {
                if (refreshed && isCurrentRead())
                  rejectedRowsController.restartAfterResourceRefresh();
              });
            }}
          />
        ) : (
          <RowsPanel
            gridRef={acceptedGridRef}
            linkLimitError={indicatorLinkController.limitError}
            onRetryLinkLimit={indicatorLinkOperation.loadLimit}
            activeTable={tableController.activeTable}
            canLink={canLink && queryAuthoring.acceptedStatus === "applied"}
            loadGenerationKey={rowsController.loadGenerationKey}
            loadState={rowsController.loadState}
            error={rowsController.error}
            query={rowsController.query}
            rows={rowsController.rows}
            rowLinkSelection={rowLinkSelection}
            onBeginLink={() => {
              if (rowLinkSelection !== null)
                indicatorLinkOperation.openDraft(
                  networkFlowRowLinkCandidate(rowLinkSelection),
                );
            }}
            onResetQuery={queryAuthoring.clearAccepted}
            page={rowsController}
            onRefreshResource={() => {
              void tableController.loadTables().then((refreshed) => {
                if (refreshed && isCurrentRead())
                  rowsController.restartAfterResourceRefresh();
              });
            }}
            onSortChange={queryAuthoring.sortAccepted}
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
  page,
  onRefreshResource,
  canLink,
  error,
  loadState,
  loadGenerationKey,
  onBeginLink,
  onResetQuery,
  onSortChange,
  onSelectionChange,
  query,
  rows,
  rowLinkSelection,
}: {
  readonly gridRef: RefObject<GridHandle | null>;
  readonly linkLimitError: string | null;
  readonly onRetryLinkLimit: () => void;
  readonly page: NetworkFlowPageNavigation;
  readonly onRefreshResource: () => void;
  readonly activeTable: NetworkFlowTable | null;
  readonly canLink: boolean;
  readonly error: NetworkFlowRequestError | null;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly loadGenerationKey: string | number;
  readonly onBeginLink: () => void;
  readonly onResetQuery: () => void;
  readonly onSortChange: (sort: NetworkFlowAcceptedQuery["sort"]) => void;
  readonly onSelectionChange: (
    activeAnchor: GridCellAnchor | null,
    cellRange: GridCellRange | null,
  ) => void;
  readonly query: NetworkFlowAcceptedQuery;
  readonly rows: readonly NetworkFlowRow[];
  readonly rowLinkSelection: NetworkFlowRowLinkSelection | null;
}) {
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
        onRetry={page.retry}
        pageFeedback={pageFailureFeedback(page)}
        semanticPageSelection
        onSelectionChange={onSelectionChange}
        onSortChange={onSortChange}
      />
      <NetworkFlowQueryPagination
        page={page}
        onRefreshResource={onRefreshResource}
      />
    </section>
  );
}

function RejectedRowsPanel({
  activeTable,
  page,
  onRefreshResource,
  diagnostics,
  error,
  loadState,
  loadGenerationKey,
  onResetQuery,
  query,
}: {
  readonly page: NetworkFlowPageNavigation;
  readonly onRefreshResource: () => void;
  readonly activeTable: NetworkFlowTable | null;
  readonly diagnostics: readonly NetworkFlowDiagnostic[];
  readonly error: NetworkFlowRequestError | null;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly loadGenerationKey: string | number;
  readonly onResetQuery: () => void;
  readonly query: NetworkFlowRejectedQuery;
}) {
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
        onRetry={page.retry}
        pageFeedback={pageFailureFeedback(page)}
        semanticPageSelection
      />
      <NetworkFlowQueryPagination
        page={page}
        onRefreshResource={onRefreshResource}
      />
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

const panelHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-md)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
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

const linkActionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
  padding: "var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
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
