import {
  networkAnalysisEdgeTestId,
  networkAnalysisTestId,
  networkAnalysisVertexTestId,
} from "@cartulary/ui-contracts";
import { Link2, Network, RefreshCw, X } from "lucide-react";
import { type CSSProperties, useLayoutEffect, useMemo, useRef } from "react";
import {
  NetworkFlowButton,
  NetworkFlowIconButton,
} from "./NetworkFlowControls";
import {
  NetworkFlowQueryPagination,
  pageFailureFeedback,
} from "./NetworkFlowQueryPagination";
import { NetworkFlowContributorGrid } from "./NetworkFlowSemanticGrid";
import type {
  NetworkFlowContributor,
  NetworkFlowTable,
} from "./networkFlowClient";
import {
  type ExplorationAction,
  type ExplorationFocus,
  type ExplorationNavigation,
  explorationObjects,
  explorationPresentation,
  explorationSelectionId,
  explorationEdgeLimit as graphEdgeRenderLimit,
  explorationVertexLimit as graphVertexRenderLimit,
} from "./networkFlowExplorationNavigation";
import {
  networkFlowEdgeLinkCandidate,
  networkFlowVertexLinkCandidate,
} from "./networkFlowIndicatorLinkModel";
import type {
  NetworkFlowPageNavigation,
  NetworkFlowQueryLoadState,
} from "./useNetworkFlowPagedQuery";

export type ExplorationContributorPage = NetworkFlowPageNavigation & {
  readonly items: readonly NetworkFlowContributor[];
  readonly loadState: NetworkFlowQueryLoadState;
  readonly loadGenerationKey: string | number;
};
export function NetworkFlowExplorationPanel({
  navigation,
  contributorPage,
  status,
  tables,
  canLink,
  onNavigate,
  onRefreshGraph,
  onLinkEdge,
  onLinkVertex,
  isFocusCurrent,
  bindFocusRestoration,
}: {
  readonly navigation: ExplorationNavigation;
  readonly contributorPage: ExplorationContributorPage;
  readonly status: {
    readonly loadState: NetworkFlowQueryLoadState;
    readonly stale: boolean;
    readonly validationMessage: string | null;
  };
  readonly tables: readonly NetworkFlowTable[];
  readonly canLink: boolean;
  readonly onNavigate: (action: ExplorationAction) => void;
  readonly onRefreshGraph: () => void;
  readonly onLinkEdge: (
    fieldKey: "network_flow.src_ip" | "network_flow.dst_ip",
  ) => void;
  readonly onLinkVertex: () => void;
  readonly isFocusCurrent: (focus: ExplorationFocus) => boolean;
  readonly bindFocusRestoration: (restore: () => boolean) => () => void;
}) {
  const graph = navigation.result;
  const presentation = explorationPresentation(navigation);
  const {
    loadState: graphLoadState,
    stale: graphStale,
    validationMessage,
  } = status;
  const contributorLoadState = contributorPage.loadState;
  const contributorLoadGenerationKey = contributorPage.loadGenerationKey;
  const contributorError = contributorPage.error;
  const contributors = contributorPage.items;
  const selectedVertex =
    navigation.selection?.kind === "vertex"
      ? (navigation.index?.vertexById.get(navigation.selection.source_vertex_id)
          ?.object ?? null)
      : null;
  const selectedEdge =
    navigation.selection && navigation.selection.kind !== "vertex"
      ? (navigation.index?.edgeById.get(navigation.selection.source_edge_id)
          ?.object ?? null)
      : null;
  const selectedObject = selectedVertex ?? selectedEdge;
  const {
    vertexPage,
    edgePage,
    bucketIndex,
    bucket: selectedBucket,
  } = presentation;
  const timeBuckets = navigation.index?.buckets ?? [];
  const graphVertices = presentation.vertices;
  const graphEdges = presentation.edges;
  const { vertices: graphAllVertices, edges: graphAllEdges } =
    explorationObjects(navigation);
  const graphEndpointLabels = useMemo(
    () =>
      new Map(
        navigation.index?.vertices.map((entry) => [
          entry.selector.source_vertex_id,
          entry.label,
        ]) ?? [],
      ),
    [navigation.index],
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
  const root = useRef<HTMLElement | null>(null);
  const ownsFocus = useRef(false);
  const latestNavigation = useRef(navigation);
  latestNavigation.current = navigation;
  useLayoutEffect(
    () =>
      bindFocusRestoration(() => {
        const container = root.current;
        if (!container?.isConnected || !latestNavigation.current.active)
          return false;
        const selector = latestNavigation.current.selection;
        const target = selector
          ? container.querySelector<HTMLElement>(
              `[data-graph-selector-id="${explorationSelectionId(selector)}"]`,
            )
          : null;
        (target ?? container).focus();
        return true;
      }),
    [bindFocusRestoration],
  );
  useLayoutEffect(() => {
    const intent = navigation.focus;
    if (intent === null || !isFocusCurrent(intent)) return;
    const container = root.current;
    if (!container?.isConnected) return;
    const target =
      container.querySelector<HTMLElement>(
        `[data-graph-selector-id="${explorationSelectionId(intent.selector)}"]`,
      ) ??
      container.querySelector<HTMLElement>(
        `nav[aria-label="${intent.selector.kind === "vertex" ? "vertices" : "edges"} navigation"] button`,
      ) ??
      container.querySelector<HTMLElement>(
        'nav[aria-label="Time bucket navigation"] button',
      ) ??
      container;
    onNavigate({ type: "interact" });
    target.focus();
  }, [navigation.focus, isFocusCurrent, onNavigate]);
  useLayoutEffect(() => {
    // A removed control has no mounted semantic realization. Only restore the
    // region if focus belonged here and no newer control received it.
    const container = root.current;
    if (
      navigation.active &&
      ownsFocus.current &&
      container?.isConnected &&
      container.ownerDocument.activeElement === container.ownerDocument.body
    ) {
      container.focus();
    }
  }, [navigation]);
  return (
    <section
      ref={root}
      tabIndex={-1}
      onFocusCapture={() => {
        ownsFocus.current = true;
      }}
      onBlurCapture={(event) => {
        if (
          event.relatedTarget &&
          !event.currentTarget.contains(event.relatedTarget)
        )
          ownsFocus.current = false;
      }}
      onPointerDownCapture={() => onNavigate({ type: "interact" })}
      onKeyDownCapture={() => onNavigate({ type: "interact" })}
      aria-label="Network Flow graph"
      data-testid={networkAnalysisTestId("graph-panel")}
      style={{
        ...graphLayoutStyle,
        gridTemplateColumns: selectedObject
          ? "minmax(0, 1fr) min(45%, var(--ct-layout-inspectorDefaultWidth))"
          : "minmax(0, 1fr)",
      }}
    >
      <div style={graphTableStyle}>
        {validationMessage === null ? null : (
          <p className="network-flow-status" data-tone="error" role="alert">
            {validationMessage}
          </p>
        )}
        {selectedBucket === null ? null : (
          <nav
            aria-label="Time bucket navigation"
            style={boundedNavigationStyle}
          >
            <NetworkFlowButton
              aria-disabled={bucketIndex === 0}
              variant="secondary"
              onClick={() =>
                onNavigate({ type: "bucket", index: bucketIndex - 1 })
              }
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
              aria-disabled={bucketIndex + 1 >= timeBuckets.length}
              variant="secondary"
              onClick={() =>
                onNavigate({ type: "bucket", index: bucketIndex + 1 })
              }
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
              : graphStale
                ? "Graph sources changed. Recompute to continue."
                : graph
                  ? "Graph ready"
                  : graphLoadState === "error"
                    ? "Graph computation failed."
                    : "No graph"}
          </span>
          <span style={mutedTextStyle}>
            Query totals: {graph?.source_table_refs.length ?? 0} tables ·{" "}
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
              onPageChange={(page) =>
                onNavigate({ type: "page", kind: "vertices", page })
              }
            />
          ) : null}
          {graphAllEdges.length > graphEdgeRenderLimit ? (
            <BoundedGraphNavigation
              itemLabel="edges"
              page={edgePage}
              pageSize={graphEdgeRenderLimit}
              total={graphAllEdges.length}
              onPageChange={(page) =>
                onNavigate({ type: "page", kind: "edges", page })
              }
            />
          ) : null}
          {graphStale || graphLoadState === "error" ? (
            <NetworkFlowButton variant="secondary" onClick={onRefreshGraph}>
              <RefreshCw aria-hidden="true" size={14} />{" "}
              {graphStale ? "Recompute graph" : "Retry graph"}
            </NetworkFlowButton>
          ) : null}
        </div>
        <div style={tableScrollStyle}>
          {graph &&
          graphAllEdges.length === 0 &&
          graphAllVertices.length === 0 ? (
            <p className="network-flow-status">
              {presentation.temporal
                ? "No flows start in this bucket."
                : "This graph has no matching flows."}
            </p>
          ) : null}
          {presentation.temporal ? (
            <p style={contextStyle}>
              Vertices shown occur in this bucket; vertex flow counts and
              contributors cover the full query.
            </p>
          ) : null}
          <h3 style={graphSectionTitleStyle}>Vertices</h3>
          <table style={dataTableStyle}>
            <thead>
              <tr>
                <th style={thStyle}>Endpoint</th>
                <th style={thStyle}>Query-wide flows</th>
                <th style={thStyle}>Tables</th>
                <th style={thStyle}>Select</th>
              </tr>
            </thead>
            <tbody>
              {graphVertices.map((entry) => {
                const vertex = entry.object;
                const selector = entry.selector;
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
                        data-graph-selector-id={selector.source_vertex_id}
                        aria-label={`Select vertex ${entry.label}`}
                        aria-pressed={selectedVertex === vertex}
                        selected={selectedVertex === vertex}
                        variant="mode"
                        onClick={() => onNavigate({ type: "select", selector })}
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
                <th style={thStyle}>
                  {presentation.temporal ? "Bucket rows" : "Query-wide rows"}
                </th>
                <th style={thStyle}>Select</th>
              </tr>
            </thead>
            <tbody>
              {graphEdges.map((entry) => {
                const edge = entry.object;
                const edgeId = entry.selector.source_edge_id;
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
                      {graphScalar(edge.properties.flow_row_count)}
                    </td>
                    <td style={tdStyle}>
                      <NetworkFlowButton
                        data-graph-selector-id={entry.selector.source_edge_id}
                        aria-label={`Select edge ${entry.label}`}
                        aria-pressed={selectedEdge === edge}
                        selected={selectedEdge === edge}
                        variant="mode"
                        onClick={() =>
                          onNavigate({
                            type: "select",
                            selector: entry.selector,
                          })
                        }
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
            event.stopPropagation();
            onNavigate({ type: "close" });
          }}
        >
          <div style={drawerHeaderStyle}>
            <strong>
              {`${selectedVertex ? "Vertex" : "Edge"} ${presentation.selected?.label ?? ""}`}
            </strong>
            <NetworkFlowIconButton
              aria-label="Close graph contributors"
              data-testid={networkAnalysisTestId("contributor-close")}
              title="Close"
              onClick={() => {
                onNavigate({ type: "close" });
              }}
            >
              <X aria-hidden="true" size={15} />
            </NetworkFlowIconButton>
          </div>
          <div style={linkActionsStyle}>
            <p style={contextStyle}>
              {navigation.selection?.kind === "time_bucket_edge"
                ? `Contributors in bucket [${navigation.selection.bucket_start_utc}, ${navigation.selection.bucket_end_utc}).`
                : "Contributors across the full query range."}
            </p>
            {presentation.offPage ? (
              <div>
                <span>Selected object is off-page. </span>
                <NetworkFlowButton
                  variant="secondary"
                  onClick={() => onNavigate({ type: "reveal" })}
                >
                  Reveal selected object
                </NetworkFlowButton>
              </div>
            ) : null}
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
            onRetry={contributorPage.retry}
            pageFeedback={pageFailureFeedback(contributorPage)}
            semanticPageSelection
          />
          <NetworkFlowQueryPagination
            page={contributorPage}
            onRefreshResource={onRefreshGraph}
          />
        </aside>
      ) : null}
      <span
        aria-live="polite"
        data-testid={networkAnalysisTestId("graph-live-region")}
        style={visuallyHiddenStyle}
      >
        {navigation.notice ||
          (graphLoadState === "loading"
            ? "Loading Network Flow graph."
            : graphStale
              ? "Graph sources changed. Recompute to continue."
              : graphLoadState === "error"
                ? "Graph computation failed."
                : graph
                  ? "Network Flow graph ready. Select a vertex or edge to inspect contributors."
                  : "Network Flow graph unavailable.")}
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
        aria-disabled={page === 0}
        variant="secondary"
        onClick={() => onPageChange(page - 1)}
      >
        Previous
      </NetworkFlowButton>
      <NetworkFlowButton
        aria-disabled={page + 1 >= pageCount}
        variant="secondary"
        onClick={() => onPageChange(page + 1)}
      >
        Next
      </NetworkFlowButton>
    </nav>
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

function graphEndpointLabel(
  value: unknown,
  labels: ReadonlyMap<string, string>,
): string {
  const endpointId = graphString(value);
  return endpointId === null
    ? "—"
    : (labels.get(endpointId) ?? "Unavailable endpoint");
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

const tableScrollStyle = {
  flex: "1 0 12rem",
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
  minBlockSize: "24rem",
  minWidth: 0,
} satisfies CSSProperties;

const graphTableStyle = {
  display: "flex",
  flexDirection: "column",
  overflow: "auto",
  minBlockSize: 0,
  minWidth: 0,
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
  minBlockSize: 0,
  minInlineSize: 0,
  overflow: "auto",
  display: "grid",
  gridTemplateRows: "auto auto minmax(0, 1fr) auto",
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

const mutedTextStyle = {
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
const contextStyle = {
  margin: 0,
  padding: "var(--ct-spacing-xs)",
  overflowWrap: "anywhere",
  fontSize: "var(--ct-typography-metadata-fontSize)",
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
