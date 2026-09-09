import {
  boundedRead,
  type ObservationClock,
} from "../services/asyncObservation";
import { equalJSONResource } from "../services/commonJobContract";
import {
  type NetworkFlowContributor,
  type NetworkFlowGraphSelector,
  type NetworkFlowSavedGraph,
  type NetworkFlowSavedGraphContributorResult,
  type NetworkFlowSavedGraphResult,
  savedGraphBindingIdentity,
} from "../services/networkFlowContractAdapter";
import type { NetworkFlowRequestError } from "./networkFlowErrors";
import type { SavedGraphLoadState } from "./savedGraphOperation";
import {
  savedGraphReadDisposition,
  savedGraphReadFailure,
} from "./savedGraphReadFailure";

export type SavedGraphResultState = {
  readonly identity: string | null;
  readonly result: Pick<NetworkFlowSavedGraphResult, "result"> | null;
  readonly resultState: SavedGraphLoadState;
  readonly resultError: NetworkFlowRequestError | null;
  readonly selection: NetworkFlowGraphSelector | null;
  readonly contributors: readonly NetworkFlowContributor[];
  readonly contributorState: SavedGraphLoadState;
  readonly contributorError: NetworkFlowRequestError | null;
  readonly nextContributorCursor: string | null;
  readonly contributorPage: number;
  readonly vertexPage: number;
  readonly edgePage: number;
  readonly bucketIndex: number;
};
export type SavedGraphResultPort = {
  readonly result: (
    graphId: string,
    signal: AbortSignal,
  ) => Promise<NetworkFlowSavedGraphResult>;
  readonly contributors: (
    graph: NetworkFlowSavedGraph,
    selector: NetworkFlowGraphSelector,
    cursor: string | undefined,
    signal: AbortSignal,
  ) => Promise<NetworkFlowSavedGraphContributorResult>;
};
export const initialSavedGraphResultState = (): SavedGraphResultState => ({
  identity: null,
  result: null,
  resultState: "idle",
  resultError: null,
  selection: null,
  contributors: [],
  contributorState: "idle",
  contributorError: null,
  nextContributorCursor: null,
  contributorPage: 0,
  vertexPage: 0,
  edgePage: 0,
  bucketIndex: 0,
});

/** Navigation only exposes bytes selected by the operation owner's current authority. */
export class SavedGraphResultNavigation {
  private state = initialSavedGraphResultState();
  private graph: NetworkFlowSavedGraph | null = null;
  private port: SavedGraphResultPort | null = null;
  private resultRequest: AbortController | null = null;
  private contributorRequest: AbortController | null = null;
  private resultGeneration = 0;
  private contributorGeneration = 0;
  private active = false;
  private valid: () => boolean = () => false;
  constructor(
    private readonly publish: (state: SavedGraphResultState) => void,
    private readonly clock?: ObservationClock,
    private readonly invalidated?: (graphId: string) => void,
    private readonly readFailed?: (
      graphId: string,
      error: NetworkFlowRequestError,
      surface: "result" | "contributors",
    ) => void,
  ) {}
  bind(port: SavedGraphResultPort): void {
    this.port = port;
  }
  authorize(
    graph: NetworkFlowSavedGraph | null,
    publicationIsCurrent: () => boolean,
  ): void {
    this.graph = graph;
    this.valid = publicationIsCurrent;
    const identity = graph === null ? null : savedGraphBindingIdentity(graph);
    if (identity === this.state.identity) {
      if (
        identity !== null &&
        this.active &&
        this.state.result === null &&
        this.state.resultState === "idle"
      )
        void this.loadResult();
      else if (
        identity !== null &&
        this.active &&
        this.state.selection !== null &&
        this.state.contributorState === "idle"
      )
        void this.loadContributors();
      return;
    }
    this.clear();
    this.update({ identity });
    if (identity !== null && this.active) void this.loadResult();
  }
  setActive(active: boolean): void {
    this.active = active;
    if (!active) this.stopReads();
    else if (
      this.state.identity !== null &&
      this.state.result === null &&
      this.state.resultState === "idle"
    )
      void this.loadResult();
  }
  clear(): void {
    this.stopRequests();
    this.update(initialSavedGraphResultState());
  }
  private stopRequests(): void {
    this.resultGeneration++;
    this.contributorGeneration++;
    this.resultRequest?.abort();
    this.contributorRequest?.abort();
    this.resultRequest = null;
    this.contributorRequest = null;
  }
  stopReads(): void {
    this.stopRequests();
    this.update({
      resultState:
        this.state.resultState === "loading" ||
        this.state.resultState === "refreshing"
          ? this.state.result === null
            ? "idle"
            : "ready"
          : this.state.resultState,
      contributorState:
        this.state.contributorState === "loading" ||
        this.state.contributorState === "refreshing"
          ? this.state.contributors.length
            ? "ready"
            : "idle"
          : this.state.contributorState,
    });
  }
  private update(patch: Partial<SavedGraphResultState>): void {
    this.state = { ...this.state, ...patch };
    this.publish(this.state);
  }
  readonly setPage = (
    kind: "vertexPage" | "edgePage" | "bucketIndex",
    page: number,
  ): void => {
    if (!Number.isSafeInteger(page) || page < 0 || this.state.result === null)
      return;
    const result = this.state.result.result;
    const maximum =
      kind === "vertexPage"
        ? Math.max(
            0,
            Math.ceil(result.graph_projection_result.vertices.length / 500) - 1,
          )
        : kind === "edgePage"
          ? Math.max(
              0,
              Math.ceil(result.graph_projection_result.edges.length / 1000) - 1,
            )
          : result.result_variant.kind === "time_bucket_v1"
            ? result.result_variant.time_buckets.length - 1
            : 0;
    if (page > maximum) return;
    this.update({
      [kind]: page,
      ...(kind === "bucketIndex" ? { vertexPage: 0, edgePage: 0 } : {}),
    });
  };
  readonly loadResult = async (): Promise<void> => {
    const graph = this.graph,
      port = this.port,
      identity = this.state.identity,
      valid = this.valid;
    if (graph === null || port === null || identity === null || !valid())
      return;
    this.resultRequest?.abort();
    const request = new AbortController();
    this.resultRequest = request;
    const generation = ++this.resultGeneration;
    this.update({
      resultState: this.state.result === null ? "loading" : "refreshing",
      resultError: null,
    });
    try {
      const result = await boundedRead(
        (signal) => port.result(graph.graph_view_id, signal),
        request.signal,
        30_000,
        this.clock,
      );
      if (
        !valid() ||
        generation !== this.resultGeneration ||
        identity !== this.state.identity
      )
        return;
      if (
        result.graph_view.state !== "active" ||
        result.graph_view.incident_id !== graph.incident_id ||
        result.graph_view.graph_view_id !== graph.graph_view_id ||
        savedGraphBindingIdentity(result.graph_view) !== identity
      ) {
        this.clear();
        this.invalidated?.(graph.graph_view_id);
        this.update({
          resultState: "error",
          resultError: savedGraphReadFailure(
            new SyntaxError(),
            "The result binding changed. Reload the current declaration.",
          ),
        });
        return;
      }
      // Unchanged results keep the original object and all navigation context.
      this.update({
        result: this.state.result ?? { result: result.result },
        resultState: "ready",
      });
    } catch (caught) {
      if (
        !valid() ||
        generation !== this.resultGeneration ||
        identity !== this.state.identity
      )
        return;
      const error = savedGraphReadFailure(
        caught,
        "The selected immutable result could not be loaded. Retry the result or reload its declaration.",
      );
      if (savedGraphReadDisposition(error, "result") !== "retain") {
        this.clear();
        if (!this.readFailed) this.invalidated?.(graph.graph_view_id);
      }
      this.readFailed?.(graph.graph_view_id, error, "result");
      this.update({ resultState: "error", resultError: error });
    } finally {
      if (this.resultRequest === request) this.resultRequest = null;
    }
  };
  readonly selectObject = async (
    selector: NetworkFlowGraphSelector | null,
  ): Promise<void> => {
    if (equalJSONResource(selector, this.state.selection)) return;
    this.contributorRequest?.abort();
    this.contributorGeneration++;
    this.update({
      selection: selector,
      contributors: [],
      contributorState: "idle",
      contributorError: null,
      contributorPage: 0,
      nextContributorCursor: null,
    });
    if (selector !== null) await this.loadContributors();
  };
  readonly loadContributors = async (advance = false): Promise<void> => {
    const graph = this.graph,
      port = this.port,
      identity = this.state.identity,
      selector = this.state.selection,
      valid = this.valid;
    if (
      graph?.selected_result_binding == null ||
      port === null ||
      identity === null ||
      selector === null ||
      !valid() ||
      (advance && this.state.nextContributorCursor === null)
    )
      return;
    const cursor = advance
      ? (this.state.nextContributorCursor ?? undefined)
      : undefined;
    const page = advance ? this.state.contributorPage + 1 : 0;
    this.contributorRequest?.abort();
    const request = new AbortController();
    this.contributorRequest = request;
    const generation = ++this.contributorGeneration;
    this.update({
      contributorState: this.state.contributors.length
        ? "refreshing"
        : "loading",
      contributorError: null,
    });
    try {
      const result = await boundedRead(
        (signal) => port.contributors(graph, selector, cursor, signal),
        request.signal,
        30_000,
        this.clock,
      );
      if (
        !valid() ||
        generation !== this.contributorGeneration ||
        identity !== this.state.identity ||
        !equalJSONResource(selector, this.state.selection)
      )
        return;
      if (
        result.graph_view_id !== graph.graph_view_id ||
        result.projection_result_id !==
          graph.selected_result_binding.projection_result_id ||
        !equalJSONResource(result.selector, selector) ||
        result.contributors.length > 100
      )
        throw new SyntaxError(
          "Contributors do not match the selected immutable result.",
        );
      this.update({
        contributors: result.contributors,
        contributorState: "ready",
        nextContributorCursor: result.meta.paging.next_cursor_token,
        contributorPage: page,
      });
    } catch (caught) {
      if (
        !valid() ||
        generation !== this.contributorGeneration ||
        identity !== this.state.identity ||
        !equalJSONResource(selector, this.state.selection)
      )
        return;
      const error = savedGraphReadFailure(
        caught,
        "Contributors could not be loaded. Restart from the current immutable result.",
      );
      if (savedGraphReadDisposition(error, "contributors") !== "retain") {
        this.clear();
        if (!this.readFailed) this.invalidated?.(graph.graph_view_id);
      }
      this.readFailed?.(graph.graph_view_id, error, "contributors");
      this.update({
        contributorState: "error",
        contributorError: error,
        ...(error.code === "network_flow_cursor_invalid"
          ? { nextContributorCursor: null }
          : {}),
      });
    } finally {
      if (this.contributorRequest === request) this.contributorRequest = null;
    }
  };
}
