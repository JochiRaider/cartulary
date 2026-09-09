import {
  networkAnalysisSavedGraphEdgeTestId,
  networkAnalysisSavedGraphTestId,
  networkAnalysisSavedGraphVertexTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import { useEffect, useMemo, useRef } from "react";
import { normalizeSavedGraphDisplayName } from "../services/networkFlowContractAdapter";
import {
  NetworkFlowActionGroup,
  NetworkFlowButton,
  NetworkFlowField,
  NetworkFlowTextInput,
} from "./NetworkFlowControls";
import type {
  NetworkFlowGraphEdge,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowGraphVertex,
  NetworkFlowTable,
} from "./networkFlowClient";
import type { SavedGraphObservation } from "./savedGraphObservation";
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
  tables,
}: {
  readonly canCreate: boolean;
  readonly canRetire: boolean;
  readonly controller: NetworkFlowSavedGraphPanelController;
  readonly currentGraph: NetworkFlowGraphResult | null;
  readonly tables: readonly NetworkFlowTable[];
}) {
  const { vertexPage, edgePage, bucketIndex } = controller;
  const setVertexPage = (page: number) =>
    controller.setPage("vertexPage", page);
  const setEdgePage = (page: number) => controller.setPage("edgePage", page);
  const bindingWithdrawn =
    controller.selectedGraph?.selected_result_binding != null &&
    controller.identity === null;
  const selectedObservation = controller.selectedGraph?.latest_job_id
    ? controller.observations[controller.selectedGraph.latest_job_id]
    : undefined;
  const contributorGroups = useMemo(() => {
    const groups = new Map<
      string,
      (typeof controller.contributors)[number][]
    >();
    for (const contributor of controller.contributors) {
      const id = contributor.row_ref.network_flow_table_id;
      const group = groups.get(id) ?? [];
      group.push(contributor);
      groups.set(id, group);
    }
    return [...groups.entries()];
  }, [controller.contributors]);
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

  useEffect(() => {
    if (
      controller.selection === null &&
      selectedObjectButtonRef.current?.isConnected === true
    ) {
      selectedObjectButtonRef.current.focus({ preventScroll: true });
    }
  }, [controller.selection]);

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
          <NetworkFlowButton
            data-testid={networkAnalysisTestId("saved-graph-create")}
            disabled={currentGraph === null}
            pending={controller.mutationPending}
            variant="primary"
            onClick={() =>
              controller.openAction(
                "create",
                currentGraph?.semantic_query ?? null,
              )
            }
          >
            Save current graph
          </NetworkFlowButton>
        ) : null}
      </header>

      <div className="network-flow-saved-workspace">
        <nav aria-label="Saved graphs" className="network-flow-saved-list">
          <div style={listHeaderStyle}>
            <strong>{controller.graphs.length} saved</strong>
            <NetworkFlowButton
              data-testid={networkAnalysisTestId("saved-graph-reload")}
              aria-disabled={controller.listState === "loading"}
              aria-busy={controller.listState === "loading" || undefined}
              variant="secondary"
              onClick={() => {
                if (controller.listState !== "loading")
                  void controller.loadGraphs();
              }}
            >
              Reload
            </NetworkFlowButton>
          </div>
          {controller.listState === "loading" ? (
            <p role="status">Loading saved graphs…</p>
          ) : null}
          {controller.listError ? (
            <p role="alert">{controller.listError.message}</p>
          ) : null}
          {controller.graphs.length === 0 ? (
            controller.listState === "ready" ? (
              <p>No saved graphs yet.</p>
            ) : null
          ) : (
            <ul style={plainListStyle}>
              {controller.graphs.map((graph) => (
                <li key={graph.graph_view_id}>
                  <NetworkFlowButton
                    aria-current={
                      controller.selectedGraphViewId === graph.graph_view_id
                        ? "true"
                        : undefined
                    }
                    className="network-flow-list-action"
                    data-testid={networkAnalysisSavedGraphTestId(
                      graph.graph_view_id,
                    )}
                    variant="ghost"
                    onClick={() =>
                      controller.selectGraphView(graph.graph_view_id)
                    }
                  >
                    <span className="network-flow-truncate">
                      {graph.display_name}
                    </span>
                    <small>
                      {graph.latest_job_id
                        ? (controller.observations[
                            graph.latest_job_id
                          ]?.job?.status.replaceAll("_", " ") ??
                          "Status unobserved")
                        : graph.last_failure_code
                          ? "Source unavailable"
                          : "No result"}
                    </small>
                  </NetworkFlowButton>
                </li>
              ))}
            </ul>
          )}
        </nav>

        <div className="network-flow-saved-result">
          {controller.selectedGraph === null ? (
            <div style={emptyStyle}>
              <strong>
                {controller.listState === "loading" ||
                controller.listState === "idle"
                  ? "Loading saved graph declarations…"
                  : controller.listError
                    ? "Saved graphs are unavailable. Reload to review current access."
                    : "Select or create a saved graph."}
              </strong>
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
                    {savedGraphStatusMessage(
                      controller.selectedGraph,
                      selectedObservation,
                      bindingWithdrawn,
                    )}
                  </p>
                </div>
                <NetworkFlowActionGroup>
                  {canCreate ? (
                    <>
                      <NetworkFlowButton
                        disabled={controller.mutationPending}
                        variant="secondary"
                        onClick={() => controller.openAction("rename")}
                      >
                        Rename
                      </NetworkFlowButton>
                      <NetworkFlowButton
                        disabled={controller.mutationPending}
                        variant="secondary"
                        onClick={() => controller.openAction("refresh")}
                      >
                        Refresh
                      </NetworkFlowButton>
                    </>
                  ) : null}
                  {canRetire ? (
                    <NetworkFlowButton
                      disabled={controller.mutationPending}
                      variant="danger"
                      onClick={() => controller.openAction("retire")}
                    >
                      Retire
                    </NetworkFlowButton>
                  ) : null}
                </NetworkFlowActionGroup>
              </header>

              {controller.notice ? (
                <p
                  aria-live="polite"
                  className="network-flow-status"
                  role="status"
                  style={noticeStyle}
                >
                  {controller.notice}
                </p>
              ) : null}
              {controller.selectedGraph.selected_result_binding !== null &&
              selectedObservation?.state === "observing" &&
              selectedObservation.job?.status !== "succeeded" ? (
                <p
                  className="network-flow-status"
                  data-tone="stale"
                  role="status"
                  style={noticeStyle}
                >
                  Showing the last successful result while refresh continues.
                </p>
              ) : null}
              {selectedObservation?.state === "paused" ? (
                <div role="status" style={noticeStyle}>
                  <p>{selectedObservation.message}</p>
                  <NetworkFlowButton
                    onClick={() => controller.resumeObservation()}
                  >
                    Resume observation
                  </NetworkFlowButton>
                </div>
              ) : null}
              {controller.resultError || bindingWithdrawn ? (
                <div role="alert" style={noticeStyle}>
                  <p>
                    {controller.resultError?.message ??
                      controller.contributorError?.message ??
                      "Result access was withdrawn. Reload the declaration to review its current binding."}
                  </p>
                  <NetworkFlowButton
                    onClick={() => void controller.loadResult()}
                  >
                    {controller.identity === null
                      ? "Reload saved graph"
                      : "Retry result"}
                  </NetworkFlowButton>
                </div>
              ) : null}
              {result === null && controller.resultState === "loading" ? (
                <p role="status">Loading immutable graph result…</p>
              ) : result === null ? (
                <div style={emptyStyle}>
                  <strong>
                    {bindingWithdrawn
                      ? "Result access requires review."
                      : controller.resultError
                        ? "The immutable result could not be loaded."
                        : controller.selectedGraph.last_failure_code !== null ||
                            selectedObservation?.job?.status === "failed"
                          ? "Materialization failed; no successful result is available."
                          : "No materialized result yet."}
                  </strong>
                  <span>
                    {bindingWithdrawn || controller.resultError
                      ? "Use the recovery action above to read the authorized result."
                      : controller.selectedGraph.last_failure_code !== null
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
                      <NetworkFlowButton
                        disabled={bucketIndex === 0}
                        variant="secondary"
                        onClick={() =>
                          controller.setPage("bucketIndex", bucketIndex - 1)
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
                        disabled={bucketIndex + 1 >= timeBuckets.length}
                        variant="secondary"
                        onClick={() =>
                          controller.setPage("bucketIndex", bucketIndex + 1)
                        }
                      >
                        Next bucket
                      </NetworkFlowButton>
                    </nav>
                  )}
                  <BoundedPager
                    itemLabel="vertices"
                    page={vertexPage}
                    pageSize={vertexPageSize}
                    total={vertices.length}
                    onPageChange={setVertexPage}
                  />
                  <ul
                    aria-label="Saved graph vertices"
                    className="network-flow-object-list"
                    style={objectListStyle}
                  >
                    {visibleVertices.map((vertex) => {
                      const selector = savedVertexSelector(graphResult, vertex);
                      return (
                        <li
                          data-testid={networkAnalysisSavedGraphVertexTestId(
                            vertex.vertex_id,
                          )}
                          key={vertex.vertex_id}
                        >
                          <NetworkFlowButton
                            disabled={selector === null}
                            variant="secondary"
                            onClick={(event) => {
                              selectedObjectButtonRef.current =
                                event.currentTarget;
                              void controller.selectObject(selector);
                            }}
                          >
                            {savedVertexLabel(vertex)}
                          </NetworkFlowButton>
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
                  <ul
                    aria-label="Saved graph edges"
                    className="network-flow-object-list"
                    style={objectListStyle}
                  >
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
                          <NetworkFlowButton
                            disabled={selector === null}
                            variant="secondary"
                            onClick={(event) => {
                              selectedObjectButtonRef.current =
                                event.currentTarget;
                              void controller.selectObject(selector);
                            }}
                          >
                            {savedEdgeLabel(edge, endpointLabels)}
                          </NetworkFlowButton>
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
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              event.stopPropagation();
              void controller.selectObject(null);
            }
          }}
        >
          <header style={listHeaderStyle}>
            <strong>
              Contributors · {controller.selectedGraph?.display_name}
            </strong>
            <NetworkFlowButton
              variant="secondary"
              onClick={() => void controller.selectObject(null)}
            >
              Close
            </NetworkFlowButton>
          </header>
          {controller.contributorState === "loading" ? (
            <p role="status">Loading contributors…</p>
          ) : controller.contributors.length === 0 &&
            controller.contributorState !== "error" ? (
            <p>No contributors were returned.</p>
          ) : (
            <div>
              {contributorGroups.map(([id, contributors]) => (
                <section
                  key={id}
                  aria-label={`Contributors from ${tables.find((table) => table.network_flow_table_id === id)?.display_name ?? shortIdentity(id)}`}
                >
                  <h4>
                    {tables.find((table) => table.network_flow_table_id === id)
                      ?.display_name ?? `Table ${shortIdentity(id)}`}
                  </h4>
                  <ol>
                    {contributors.map((contributor) => (
                      <li key={contributor.row_ref.network_flow_row_id}>
                        Row {contributor.row_ref.source_row_number} · table{" "}
                        {shortIdentity(id)}
                      </li>
                    ))}
                  </ol>
                </section>
              ))}
            </div>
          )}
          {controller.contributorError ? (
            <p role="alert">{controller.contributorError.message}</p>
          ) : null}
          <nav aria-label="Contributor pages" style={pagerStyle}>
            <span>
              Page {controller.contributorPage + 1} · up to 100 contributors
            </span>
            <NetworkFlowButton
              disabled={
                controller.contributorState === "loading" ||
                controller.contributorState === "refreshing"
              }
              onClick={() => void controller.loadContributors()}
            >
              Restart contributors
            </NetworkFlowButton>
            <NetworkFlowButton
              disabled={
                controller.nextContributorCursor === null ||
                controller.contributorState === "loading" ||
                controller.contributorState === "refreshing"
              }
              onClick={() => void controller.loadContributors(true)}
            >
              Next contributors
            </NetworkFlowButton>
          </nav>
        </aside>
      ) : null}

      {controller.notice && controller.selectedGraph === null ? (
        <p
          className="network-flow-status"
          data-tone="info"
          role="status"
          style={noticeStyle}
        >
          {controller.notice}
        </p>
      ) : null}
      {controller.operation !== null &&
      !controller.dialogOpen &&
      controller.operation.phase !== "acknowledged" ? (
        <div role="status" style={noticeStyle}>
          <p>
            {controller.operation.intent.kind} ·{" "}
            {controller.operation.intent.target?.display_name ??
              controller.operation.draft}
            : {controller.operation.phase.replaceAll("_", " ")}. Closing a
            dialog does not cancel server work.
          </p>
          <NetworkFlowButton onClick={controller.reopenOperation}>
            Review saved graph operation
          </NetworkFlowButton>
        </div>
      ) : null}
      {controller.dialogOpen && controller.operation !== null ? (
        <SavedGraphDialog controller={controller} />
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

function savedGraphStatusMessage(
  graph: NetworkFlowSavedGraphPanelController["selectedGraph"],
  observation: SavedGraphObservation | undefined,
  bindingWithdrawn: boolean,
): string {
  if (graph === null) return "";
  if (bindingWithdrawn)
    return "The selected result is unavailable until its declaration is revalidated.";
  const retained = graph.selected_result_binding !== null;
  if (observation?.job?.status === "failed" || graph.last_failure_code !== null)
    return retained
      ? "Refresh failed; the last successful result remains available."
      : "Materialization failed; no successful result is available.";
  if (observation?.job?.status === "canceled")
    return retained
      ? "Refresh was canceled; the last successful result remains available."
      : "Materialization was canceled.";
  if (observation?.job?.status === "succeeded")
    return retained
      ? "Materialization succeeded."
      : "The job succeeded. Reload the declaration to observe its selected result.";
  if (observation?.job?.status === "running") return "Materialization running.";
  if (observation?.job?.status === "queued") return "Materialization queued.";
  return retained
    ? "Showing the selected immutable result. Job status is separate."
    : "No materialized result yet. An accepted request is not a completed result.";
}

function SavedGraphDialog({
  controller,
}: {
  readonly controller: NetworkFlowSavedGraphPanelController;
}) {
  const operation = controller.operation;
  const kind = operation?.intent.kind ?? "create";
  const label =
    kind === "create"
      ? "Save current graph"
      : kind === "rename"
        ? "Rename saved graph"
        : kind === "refresh"
          ? "Refresh saved graph"
          : "Retire saved graph";
  const focus = useNetworkFlowModalFocus<HTMLDivElement>({
    onDismiss: controller.closeDialog,
    initialFocusTestId:
      kind === "create" || kind === "rename"
        ? networkAnalysisTestId("saved-graph-name")
        : undefined,
    fallbackFocusTestId: networkAnalysisTestId("saved-graph-reload"),
  });
  if (operation === null) return null;
  const named = kind === "create" || kind === "rename";
  const name = normalizeSavedGraphDisplayName(operation.draft);
  const nameRejected =
    operation.failure?.detail?.code === "network_flow_invalid_display_name";
  const locked =
    operation.phase === "submitting" ||
    operation.phase === "uncertain" ||
    operation.phase === "acknowledged";
  const action =
    kind === "create"
      ? "Save graph"
      : kind === "rename"
        ? "Rename graph"
        : kind === "refresh"
          ? "Refresh graph"
          : "Retire graph";
  return (
    <div className="network-flow-dialog-backdrop">
      <div
        ref={focus.dialogRef}
        aria-label={label}
        aria-modal="true"
        className="network-flow-dialog"
        data-testid={networkAnalysisTestId("saved-graph-dialog")}
        role="dialog"
        onKeyDown={focus.onKeyDown}
      >
        <form
          className="network-flow-dialog-form"
          onSubmit={(event) => {
            event.preventDefault();
            void controller.submit();
          }}
        >
          <h3>{label}</h3>
          {operation.intent.target ? (
            <p>
              Target: {operation.intent.target.display_name} ·{" "}
              {shortIdentity(operation.intent.target.graph_view_id)} · version{" "}
              {operation.intent.target.graph_view_version}
            </p>
          ) : (
            <p>
              The captured query includes{" "}
              {operation.intent.query?.selected_table_ids.length ?? 0} source
              tables.
            </p>
          )}
          {kind === "retire" ? (
            <p>
              The declaration leaves the active list. Leased immutable results
              remain protected.
            </p>
          ) : kind === "refresh" ? (
            <p>
              Materialization runs on the server. The selected result remains
              available while a new result is produced.
            </p>
          ) : null}
          {named ? (
            <NetworkFlowField
              htmlFor="network-flow-saved-graph-name"
              label="Display name"
            >
              <NetworkFlowTextInput
                data-testid={networkAnalysisTestId("saved-graph-name")}
                id="network-flow-saved-graph-name"
                required
                value={operation.draft}
                disabled={locked}
                aria-invalid={!name.ok || nameRejected || undefined}
                aria-describedby={
                  nameRejected
                    ? "saved-graph-name-guidance saved-graph-operation-error"
                    : "saved-graph-name-guidance"
                }
                onChange={(event) =>
                  controller.setDraft(event.currentTarget.value)
                }
              />
              <span id="saved-graph-name-guidance" aria-live="polite">
                Up to 64 UTF-8 bytes after normalization. Duplicate names are
                allowed.
                {name.ok
                  ? ` ${new TextEncoder().encode(name.name).length} bytes.`
                  : name.reason === "display_name_too_long"
                    ? " This name exceeds 64 bytes."
                    : name.reason === "forbidden_control"
                      ? " Remove control characters."
                      : " Enter a nonempty name."}
              </span>
            </NetworkFlowField>
          ) : null}
          {operation.failure ? (
            <div
              id="saved-graph-operation-error"
              role="alert"
              className="network-flow-status"
            >
              <p>{operation.failure.message}</p>
              <p>
                {operation.failure.category.replaceAll("_", " ")}
                {operation.failure.detail?.reasonCode
                  ? ` · ${operation.failure.detail.reasonCode}`
                  : ""}
              </p>
            </div>
          ) : null}
          {operation.phase === "submitting" ? (
            <p role="status">
              Submitting the captured request. Closing this dialog does not
              cancel server work.
            </p>
          ) : null}
          {operation.phase === "uncertain" ? (
            <p role="status">
              The outcome is uncertain. Exact replay uses the original request
              and transaction ID.
            </p>
          ) : null}
          <NetworkFlowActionGroup>
            <NetworkFlowButton onClick={controller.closeDialog}>
              Close
            </NetworkFlowButton>
            {operation.phase === "uncertain" ? (
              <NetworkFlowButton onClick={() => void controller.replay()}>
                Replay exact attempt
              </NetworkFlowButton>
            ) : operation.phase === "awaiting_review" ? (
              <NetworkFlowButton
                onClick={() => void controller.reviewCurrent()}
              >
                Review current graph
              </NetworkFlowButton>
            ) : (
              <NetworkFlowButton
                type="submit"
                disabled={locked || (named && !name.ok)}
                pending={operation.phase === "submitting"}
                variant={kind === "retire" ? "danger" : "primary"}
              >
                {action}
              </NetworkFlowButton>
            )}
          </NetworkFlowActionGroup>
        </form>
      </div>
    </div>
  );
}

function savedVertexSelector(
  graph: SavedGraphQueryResult | null,
  vertex: NetworkFlowGraphVertex,
): Extract<NetworkFlowGraphSelector, { readonly kind: "vertex" }> | null {
  return (
    graph?.vertex_selectors.find(
      (binding) => binding.projected_vertex_id === vertex.vertex_id,
    )?.selector ?? null
  );
}
function savedEdgeSelector(
  graph: SavedGraphQueryResult | null,
  edge: NetworkFlowGraphEdge,
  _endpointLabels: ReadonlyMap<string, string>,
): Exclude<NetworkFlowGraphSelector, { readonly kind: "vertex" }> | null {
  return (
    graph?.edge_annotations.find(
      (annotation) => annotation.projected_edge_id === edge.edge_id,
    )?.selector ?? null
  );
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

const panelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-md)",
  minHeight: 320,
  minWidth: 0,
} as const;
const headerStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-lg)",
  alignItems: "start",
  flexWrap: "wrap",
} as const;
const titleStyle = { margin: 0 } as const;
const mutedStyle = {
  color: "var(--ct-colors-ink-muted)",
  margin: "var(--ct-spacing-xs) 0 0",
} as const;
const listHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-sm)",
} as const;
const plainListStyle = {
  listStyle: "none",
  margin: "var(--ct-spacing-sm) 0 0",
  padding: 0,
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} as const;
const resultHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-md)",
  alignItems: "start",
  flexWrap: "wrap",
} as const;
const noticeStyle = {
  padding: "var(--ct-spacing-sm)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
} as const;
const emptyStyle = {
  minHeight: 120,
  display: "grid",
  placeContent: "center",
  gap: "var(--ct-spacing-sm)",
  textAlign: "center",
} as const;
const summaryStyle = {
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  color: "var(--ct-colors-ink-muted)",
} as const;
const pagerStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
  margin: "var(--ct-spacing-md) 0 var(--ct-spacing-sm)",
  flexWrap: "wrap",
} as const;
const objectListStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fill, minmax(190px, 1fr))",
  gap: "var(--ct-spacing-xs)",
  listStyle: "none",
  margin: 0,
  padding: 0,
  maxHeight: 280,
  overflow: "auto",
} as const;
const drawerStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  padding: "var(--ct-spacing-md)",
} as const;
