import {
  networkAnalysisSavedGraphEdgeTestId,
  networkAnalysisSavedGraphVertexTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import { useEffect, useMemo, useRef, useState } from "react";
import type {
  NetworkFlowGraphEdge,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowGraphVertex,
} from "./networkFlowClient";
import { useNetworkFlowModalFocus } from "./useNetworkFlowModalFocus";
import type { useNetworkFlowSavedGraphController } from "./useNetworkFlowSavedGraphController";

const vertexPageSize = 500;
const edgePageSize = 1_000;

export type NetworkFlowSavedGraphPanelController = ReturnType<
  typeof useNetworkFlowSavedGraphController
>;
type SavedGraphQueryResult = NonNullable<
  NetworkFlowSavedGraphPanelController["result"]
>["result"];

export function NetworkFlowSavedGraphPanel({
  canCreate,
  canRetire,
  controller,
  currentGraph,
}: {
  readonly canCreate: boolean;
  readonly canRetire: boolean;
  readonly controller: NetworkFlowSavedGraphPanelController;
  readonly currentGraph: NetworkFlowGraphResult | null;
}) {
  const [dialog, setDialog] = useState<"create" | "rename" | "retire" | null>(
    null,
  );
  const [displayName, setDisplayName] = useState("");
  const [vertexPage, setVertexPage] = useState(0);
  const [edgePage, setEdgePage] = useState(0);
  const [bucketIndex, setBucketIndex] = useState(0);
  const selectedObjectButtonRef = useRef<HTMLButtonElement | null>(null);
  const graphResult = controller.result?.result ?? null;
  const result = graphResult?.graph_projection_result ?? null;
  const timeBuckets =
    graphResult?.schema_id === "cartulary.network_flow.graph_query_result.v2" &&
    graphResult.result_variant.kind === "time_bucket_v1"
      ? graphResult.result_variant.time_buckets
      : [];
  const selectedBucket = timeBuckets[bucketIndex] ?? null;
  const visibleTemporalEdgeIDs = useMemo(
    () =>
      new Set(
        graphResult?.schema_id ===
          "cartulary.network_flow.graph_query_result.v2" &&
          selectedBucket !== null
          ? graphResult.edge_annotations.flatMap((annotation) =>
              annotation.selector.kind === "time_bucket_edge" &&
              annotation.selector.bucket_start_utc ===
                selectedBucket.start_utc &&
              annotation.selector.bucket_end_utc === selectedBucket.end_utc
                ? [annotation.projected_edge_id]
                : [],
            )
          : [],
      ),
    [graphResult, selectedBucket],
  );
  const visibleTemporalVertexIDs = useMemo(() => {
    const ids = new Set<string>();
    if (selectedBucket === null) return ids;
    for (const edge of result?.edges ?? []) {
      if (visibleTemporalEdgeIDs.has(edge.edge_id)) {
        ids.add(edge.src_vertex_id);
        ids.add(edge.dst_vertex_id);
      }
    }
    return ids;
  }, [result, selectedBucket, visibleTemporalEdgeIDs]);
  const vertices = useMemo(
    () =>
      [...(result?.vertices ?? [])]
        .filter(
          (vertex) =>
            selectedBucket === null ||
            visibleTemporalVertexIDs.has(vertex.vertex_id),
        )
        .sort((left, right) =>
          savedVertexLabel(left).localeCompare(savedVertexLabel(right)),
        ),
    [result, selectedBucket, visibleTemporalVertexIDs],
  );
  const endpointLabels = useMemo(
    () =>
      new Map(
        vertices.map(
          (vertex) => [vertex.vertex_id, savedVertexLabel(vertex)] as const,
        ),
      ),
    [vertices],
  );
  const edges = useMemo(
    () =>
      [...(result?.edges ?? [])]
        .filter(
          (edge) =>
            selectedBucket === null || visibleTemporalEdgeIDs.has(edge.edge_id),
        )
        .sort((left, right) =>
          savedEdgeLabel(left, endpointLabels).localeCompare(
            savedEdgeLabel(right, endpointLabels),
          ),
        ),
    [endpointLabels, result, selectedBucket, visibleTemporalEdgeIDs],
  );
  const visibleVertices = vertices.slice(
    vertexPage * vertexPageSize,
    (vertexPage + 1) * vertexPageSize,
  );
  const visibleEdges = edges.slice(
    edgePage * edgePageSize,
    (edgePage + 1) * edgePageSize,
  );

  // biome-ignore lint/correctness/useExhaustiveDependencies: result identity is the page-reset boundary.
  useEffect(() => {
    setVertexPage(0);
    setEdgePage(0);
    setBucketIndex(0);
  }, [result?.projection_result_id]);

  // biome-ignore lint/correctness/useExhaustiveDependencies: bucket navigation resets bounded object pages.
  useEffect(() => {
    setVertexPage(0);
    setEdgePage(0);
  }, [bucketIndex]);

  useEffect(() => {
    if (
      controller.selection === null &&
      selectedObjectButtonRef.current?.isConnected === true
    ) {
      selectedObjectButtonRef.current.focus({ preventScroll: true });
    }
  }, [controller.selection]);

  function closeDialog() {
    setDialog(null);
    setDisplayName("");
  }

  const modalFocus = useNetworkFlowModalFocus<HTMLDivElement>({
    dismissDisabled: controller.mutationPending,
    initialFocusTestId:
      dialog === "create" || dialog === "rename"
        ? networkAnalysisTestId("saved-graph-name")
        : undefined,
    onDismiss: closeDialog,
  });

  return (
    <section
      aria-label="Saved Network Flow graphs"
      data-testid={networkAnalysisTestId("saved-graphs")}
      style={panelStyle}
    >
      <header style={headerStyle}>
        <div>
          <h3 style={titleStyle}>Saved graphs</h3>
          <p style={mutedStyle}>
            Immutable results remain stable while refresh jobs produce a newer
            generation.
          </p>
        </div>
        {canCreate ? (
          <button
            data-testid={networkAnalysisTestId("saved-graph-create")}
            disabled={currentGraph === null || controller.mutationPending}
            type="button"
            onClick={() => {
              setDisplayName("");
              setDialog("create");
            }}
          >
            Save current graph
          </button>
        ) : null}
      </header>

      <div style={workspaceStyle}>
        <nav aria-label="Saved graphs" style={listStyle}>
          <div style={listHeaderStyle}>
            <strong>{controller.graphs.length} saved</strong>
            <button
              disabled={controller.listState === "loading"}
              type="button"
              onClick={() => void controller.loadGraphs()}
            >
              Reload
            </button>
          </div>
          {controller.listState === "loading" ? (
            <p role="status">Loading saved graphs…</p>
          ) : controller.listState === "error" ? (
            <p role="alert">Saved graphs are unavailable.</p>
          ) : controller.graphs.length === 0 ? (
            <p>No saved graphs yet.</p>
          ) : (
            <ul style={plainListStyle}>
              {controller.graphs.map((graph) => (
                <li key={graph.graph_view_id}>
                  <button
                    aria-current={
                      controller.selectedGraphViewId === graph.graph_view_id
                        ? "true"
                        : undefined
                    }
                    style={
                      controller.selectedGraphViewId === graph.graph_view_id
                        ? selectedListButtonStyle
                        : listButtonStyle
                    }
                    type="button"
                    onClick={() =>
                      controller.selectGraphView(graph.graph_view_id)
                    }
                  >
                    <span>{graph.display_name}</span>
                    <small>
                      {materializationLabel(graph.last_materialization_status)}
                    </small>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </nav>

        <div style={resultStyle}>
          {controller.selectedGraph === null ? (
            <div style={emptyStyle}>
              <strong>Select or create a saved graph.</strong>
              <span>
                Unsaved exploration remains available in the current graph view.
              </span>
            </div>
          ) : (
            <>
              <header style={resultHeaderStyle}>
                <div>
                  <h4
                    data-testid={networkAnalysisTestId("saved-graph-heading")}
                    style={titleStyle}
                  >
                    {controller.selectedGraph.display_name}
                  </h4>
                  <p aria-live="polite" style={mutedStyle}>
                    {savedGraphStatusMessage(controller.selectedGraph)}
                  </p>
                </div>
                <div style={actionsStyle}>
                  {canCreate ? (
                    <>
                      <button
                        disabled={controller.mutationPending}
                        type="button"
                        onClick={() => {
                          setDisplayName(
                            controller.selectedGraph?.display_name ?? "",
                          );
                          setDialog("rename");
                        }}
                      >
                        Rename
                      </button>
                      <button
                        disabled={controller.mutationPending}
                        type="button"
                        onClick={() => void controller.refreshGraph()}
                      >
                        Refresh
                      </button>
                    </>
                  ) : null}
                  {canRetire ? (
                    <button
                      disabled={controller.mutationPending}
                      type="button"
                      onClick={() => setDialog("retire")}
                    >
                      Retire
                    </button>
                  ) : null}
                </div>
              </header>

              {controller.notice ? (
                <p aria-live="polite" role="status" style={noticeStyle}>
                  {controller.notice}
                </p>
              ) : null}
              {controller.selectedGraph.selected_result !== null &&
              (controller.selectedGraph.last_materialization_status ===
                "queued" ||
                controller.selectedGraph.last_materialization_status ===
                  "running") ? (
                <p role="status" style={noticeStyle}>
                  Showing the last successful result while refresh continues.
                </p>
              ) : null}

              {controller.resultState === "loading" ? (
                <p role="status">Loading immutable graph result…</p>
              ) : controller.resultState === "error" ? (
                <div role="alert" style={emptyStyle}>
                  <span>The selected result could not be loaded.</span>
                  <button
                    type="button"
                    onClick={() => void controller.loadResult()}
                  >
                    Retry result
                  </button>
                </div>
              ) : result === null ? (
                <div style={emptyStyle}>
                  <strong>No materialized result yet.</strong>
                  <span>
                    {controller.selectedGraph.last_materialization_status ===
                    "failed"
                      ? `The last attempt failed${controller.selectedGraph.last_failure_code ? ` (${controller.selectedGraph.last_failure_code})` : ""}. Refresh to retry.`
                      : "The result will appear after the materialization job succeeds."}
                  </span>
                </div>
              ) : (
                <div data-testid={networkAnalysisTestId("saved-graph-result")}>
                  <p style={summaryStyle}>
                    Result {shortIdentity(result.projection_result_id)} ·{" "}
                    {vertices.length} vertices · {edges.length} edges
                  </p>
                  {vertices.length > vertexPageSize ||
                  edges.length > edgePageSize ? (
                    <p style={mutedStyle}>
                      Large results are paged: at most {vertexPageSize} vertices
                      and {edgePageSize} edges are mounted at once.
                    </p>
                  ) : null}
                  {selectedBucket === null ? null : (
                    <nav
                      aria-label="Saved graph time bucket navigation"
                      style={pagerStyle}
                    >
                      <button
                        disabled={bucketIndex === 0}
                        type="button"
                        onClick={() => setBucketIndex((current) => current - 1)}
                      >
                        Previous bucket
                      </button>
                      <strong>
                        Bucket {bucketIndex + 1} of {timeBuckets.length}
                      </strong>
                      <span>
                        [{selectedBucket.start_utc}, {selectedBucket.end_utc}) ·{" "}
                        {selectedBucket.unique_vertex_count} vertices ·{" "}
                        {selectedBucket.edge_count} edges ·{" "}
                        {selectedBucket.contributing_row_count} rows
                      </span>
                      <button
                        disabled={bucketIndex + 1 >= timeBuckets.length}
                        type="button"
                        onClick={() => setBucketIndex((current) => current + 1)}
                      >
                        Next bucket
                      </button>
                    </nav>
                  )}
                  <BoundedPager
                    itemLabel="vertices"
                    page={vertexPage}
                    pageSize={vertexPageSize}
                    total={vertices.length}
                    onPageChange={setVertexPage}
                  />
                  <ul aria-label="Saved graph vertices" style={objectListStyle}>
                    {visibleVertices.map((vertex) => {
                      const selector = savedVertexSelector(graphResult, vertex);
                      return (
                        <li
                          data-testid={networkAnalysisSavedGraphVertexTestId(
                            vertex.vertex_id,
                          )}
                          key={vertex.vertex_id}
                        >
                          <button
                            disabled={selector === null}
                            type="button"
                            onClick={(event) => {
                              selectedObjectButtonRef.current =
                                event.currentTarget;
                              void controller.selectObject(selector);
                            }}
                          >
                            {savedVertexLabel(vertex)}
                          </button>
                        </li>
                      );
                    })}
                  </ul>
                  <BoundedPager
                    itemLabel="edges"
                    page={edgePage}
                    pageSize={edgePageSize}
                    total={edges.length}
                    onPageChange={setEdgePage}
                  />
                  <ul aria-label="Saved graph edges" style={objectListStyle}>
                    {visibleEdges.map((edge) => {
                      const selector = savedEdgeSelector(
                        graphResult,
                        edge,
                        endpointLabels,
                      );
                      return (
                        <li
                          data-testid={networkAnalysisSavedGraphEdgeTestId(
                            edge.edge_id,
                          )}
                          key={edge.edge_id}
                        >
                          <button
                            disabled={selector === null}
                            type="button"
                            onClick={(event) => {
                              selectedObjectButtonRef.current =
                                event.currentTarget;
                              void controller.selectObject(selector);
                            }}
                          >
                            {savedEdgeLabel(edge, endpointLabels)}
                          </button>
                        </li>
                      );
                    })}
                  </ul>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {controller.selection ? (
        <aside
          aria-label="Saved graph contributors"
          data-testid={networkAnalysisTestId("saved-graph-contributors")}
          style={drawerStyle}
        >
          <header style={listHeaderStyle}>
            <strong>Contributors</strong>
            <button
              type="button"
              onClick={() => void controller.selectObject(null)}
            >
              Close
            </button>
          </header>
          {controller.contributorState === "loading" ? (
            <p role="status">Loading contributors…</p>
          ) : controller.contributorState === "error" ? (
            <p role="alert">Contributors are unavailable.</p>
          ) : controller.contributors.length === 0 ? (
            <p>No contributors were returned.</p>
          ) : (
            <ol>
              {controller.contributors.map((contributor) => (
                <li key={contributor.row_ref.network_flow_row_id}>
                  Row {contributor.row_ref.source_row_number} · table{" "}
                  {shortIdentity(contributor.row_ref.network_flow_table_id)}
                </li>
              ))}
            </ol>
          )}
        </aside>
      ) : null}

      {dialog ? (
        <div
          ref={modalFocus.dialogRef}
          aria-label={
            dialog === "create"
              ? "Save current graph"
              : dialog === "rename"
                ? "Rename saved graph"
                : "Retire saved graph"
          }
          aria-modal="true"
          data-testid={networkAnalysisTestId("saved-graph-dialog")}
          role="dialog"
          style={dialogStyle}
          onKeyDown={modalFocus.onKeyDown}
        >
          {dialog === "retire" ? (
            <>
              <h3>Retire {controller.selectedGraph?.display_name}?</h3>
              <p>
                The graph leaves the active list. Leased immutable results
                remain protected.
              </p>
              <div style={actionsStyle}>
                <button type="button" onClick={closeDialog}>
                  Cancel
                </button>
                <button
                  disabled={controller.mutationPending}
                  type="button"
                  onClick={() => {
                    void controller.retireGraph().then((succeeded) => {
                      if (succeeded) closeDialog();
                    });
                  }}
                >
                  Retire graph
                </button>
              </div>
            </>
          ) : (
            <form
              onSubmit={(event) => {
                event.preventDefault();
                const normalized = displayName.trim();
                if (normalized.length === 0) return;
                const operation =
                  dialog === "create" && currentGraph !== null
                    ? controller.createGraph(normalized, currentGraph)
                    : controller.renameGraph(normalized);
                void operation.then((succeeded) => {
                  if (succeeded) closeDialog();
                });
              }}
            >
              <h3>
                {dialog === "create"
                  ? "Save current graph"
                  : "Rename saved graph"}
              </h3>
              <label style={fieldStyle}>
                Display name
                <input
                  data-testid={networkAnalysisTestId("saved-graph-name")}
                  maxLength={64}
                  required
                  value={displayName}
                  onChange={(event) =>
                    setDisplayName(event.currentTarget.value)
                  }
                />
              </label>
              <div style={actionsStyle}>
                <button type="button" onClick={closeDialog}>
                  Cancel
                </button>
                <button
                  disabled={
                    controller.mutationPending ||
                    displayName.trim().length === 0
                  }
                  type="submit"
                >
                  {dialog === "create" ? "Save graph" : "Rename graph"}
                </button>
              </div>
            </form>
          )}
        </div>
      ) : null}
    </section>
  );
}

function BoundedPager({
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
    <nav aria-label={`${itemLabel} navigation`} style={pagerStyle}>
      <strong>{itemLabel}</strong>
      <span>
        Page {page + 1} of {pageCount}
      </span>
      <button
        disabled={page === 0}
        type="button"
        onClick={() => onPageChange(page - 1)}
      >
        Previous
      </button>
      <button
        disabled={page + 1 >= pageCount}
        type="button"
        onClick={() => onPageChange(page + 1)}
      >
        Next
      </button>
    </nav>
  );
}

function materializationLabel(status: string): string {
  return status.replaceAll("_", " ");
}

function savedGraphStatusMessage(
  graph: NetworkFlowSavedGraphPanelController["selectedGraph"],
): string {
  if (graph === null) return "";
  switch (graph.last_materialization_status) {
    case "queued":
      return "Materialization queued.";
    case "running":
      return "Materialization running.";
    case "succeeded":
      return "Materialization succeeded.";
    case "failed":
      return graph.selected_result === null
        ? "Materialization failed; no successful result is available."
        : "Refresh failed; the last successful result remains available.";
    case "cancelled":
      return graph.selected_result === null
        ? "Materialization was cancelled."
        : "Refresh was cancelled; the last successful result remains available.";
    default:
      return "Materialization has not started.";
  }
}

function savedVertexSelector(
  graph: SavedGraphQueryResult | null,
  vertex: NetworkFlowGraphVertex,
): Extract<NetworkFlowGraphSelector, { readonly kind: "vertex" }> | null {
  if (graph?.schema_id === "cartulary.network_flow.graph_query_result.v2") {
    return (
      graph.vertex_selectors.find(
        (binding) => binding.projected_vertex_id === vertex.vertex_id,
      )?.selector ?? null
    );
  }
  const sourceVertexID = vertex.source_entity_ref?.source_entity_id;
  const endpointValue = vertex.properties.endpoint_value;
  return typeof sourceVertexID === "string" &&
    typeof endpointValue === "string" &&
    endpointValue !== ""
    ? {
        kind: "vertex",
        source_vertex_id: sourceVertexID,
        endpoint_value: endpointValue,
      }
    : null;
}

function savedEdgeSelector(
  graph: SavedGraphQueryResult | null,
  edge: NetworkFlowGraphEdge,
  endpointLabels: ReadonlyMap<string, string>,
): Exclude<NetworkFlowGraphSelector, { readonly kind: "vertex" }> | null {
  if (graph?.schema_id === "cartulary.network_flow.graph_query_result.v2") {
    return (
      graph.edge_annotations.find(
        (annotation) => annotation.projected_edge_id === edge.edge_id,
      )?.selector ?? null
    );
  }
  const sourceEdgeID =
    typeof edge.properties.edge_id === "string"
      ? edge.properties.edge_id
      : edge.source_relationship_ref?.source_relationship_id;
  const sourceEndpoint = endpointLabels.get(edge.src_vertex_id);
  const destinationEndpoint = endpointLabels.get(edge.dst_vertex_id);
  const protocol = edge.properties.ip_protocol;
  const destinationPort = edge.properties.dst_port;
  if (
    typeof sourceEdgeID !== "string" ||
    typeof sourceEndpoint !== "string" ||
    typeof destinationEndpoint !== "string" ||
    typeof protocol !== "number"
  ) {
    return null;
  }
  return typeof destinationPort === "number"
    ? {
        kind: "default_edge",
        source_edge_id: sourceEdgeID,
        source_endpoint_value: sourceEndpoint,
        destination_endpoint_value: destinationEndpoint,
        protocol,
        destination_port_present: true,
        destination_port: destinationPort,
      }
    : {
        kind: "default_edge",
        source_edge_id: sourceEdgeID,
        source_endpoint_value: sourceEndpoint,
        destination_endpoint_value: destinationEndpoint,
        protocol,
        destination_port_present: false,
      };
}

function savedVertexLabel(vertex: NetworkFlowGraphVertex): string {
  const value = vertex.properties.endpoint_value;
  return typeof value === "string" && value.length > 0
    ? value
    : shortIdentity(vertex.vertex_id);
}

function savedEdgeLabel(
  edge: NetworkFlowGraphEdge,
  endpointLabels: ReadonlyMap<string, string>,
): string {
  const source =
    endpointLabels.get(edge.src_vertex_id) ?? shortIdentity(edge.src_vertex_id);
  const destination =
    endpointLabels.get(edge.dst_vertex_id) ?? shortIdentity(edge.dst_vertex_id);
  const protocol = edge.properties.ip_protocol;
  return `${source} → ${destination}${typeof protocol === "number" || typeof protocol === "string" ? ` · protocol ${protocol}` : ""}`;
}

function shortIdentity(value: string): string {
  return value.length <= 18
    ? value
    : `${value.slice(0, 10)}…${value.slice(-6)}`;
}

const panelStyle = { display: "grid", gap: 12, minHeight: 320 } as const;
const headerStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: 16,
  alignItems: "start",
} as const;
const titleStyle = { margin: 0 } as const;
const mutedStyle = {
  color: "var(--ct-colors-text-muted)",
  margin: "4px 0 0",
} as const;
const workspaceStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(180px, 240px) minmax(0, 1fr)",
  gap: 12,
} as const;
const listStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: 8,
  padding: 10,
} as const;
const listHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: 8,
} as const;
const plainListStyle = {
  listStyle: "none",
  margin: "10px 0 0",
  padding: 0,
  display: "grid",
  gap: 6,
} as const;
const listButtonStyle = {
  width: "100%",
  display: "grid",
  gap: 2,
  textAlign: "left",
} as const;
const selectedListButtonStyle = {
  ...listButtonStyle,
  outline: "var(--ct-component-focus-ring-border)",
} as const;
const resultStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: 8,
  padding: 12,
  minWidth: 0,
} as const;
const resultHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: 12,
  alignItems: "start",
} as const;
const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: 8,
  justifyContent: "flex-end",
} as const;
const noticeStyle = {
  padding: 8,
  borderRadius: 6,
  background: "var(--ct-colors-surface-2)",
} as const;
const emptyStyle = {
  minHeight: 120,
  display: "grid",
  placeContent: "center",
  gap: 8,
  textAlign: "center",
} as const;
const summaryStyle = {
  fontFamily: "var(--ct-font-family-mono)",
  color: "var(--ct-colors-text-muted)",
} as const;
const pagerStyle = {
  display: "flex",
  alignItems: "center",
  gap: 8,
  margin: "12px 0 6px",
} as const;
const objectListStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fill, minmax(190px, 1fr))",
  gap: 6,
  listStyle: "none",
  margin: 0,
  padding: 0,
  maxHeight: 280,
  overflow: "auto",
} as const;
const drawerStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: 8,
  padding: 12,
} as const;
const dialogStyle = {
  position: "fixed",
  inset: "50% auto auto 50%",
  transform: "translate(-50%, -50%)",
  width: "min(420px, calc(100vw - 32px))",
  zIndex: 20,
  padding: 18,
  borderRadius: 10,
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "0 16px 48px rgb(0 0 0 / 0.3)",
} as const;
const fieldStyle = { display: "grid", gap: 6, margin: "14px 0" } as const;
