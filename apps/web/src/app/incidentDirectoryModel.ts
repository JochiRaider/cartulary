import {
  type APIError,
  extractError,
  type HTTPOperationResult,
} from "../services/browserApi";
import type {
  IncidentDirectoryPaging,
  IncidentDirectoryResource,
  ListVisibleIncidentsResponse,
} from "./api/publicHttpTypes";

export type IncidentDirectoryQuery = Readonly<{
  search: string;
  statusFilter: "all" | IncidentDirectoryResource["status"];
}>;
export type IncidentDirectoryFailure = Readonly<{
  error: APIError | null;
  message: string;
  scope: "replace" | "page";
  restart: boolean;
}>;
export type IncidentDirectoryState = Readonly<{
  query: IncidentDirectoryQuery;
  acceptedQuery: IncidentDirectoryQuery | null;
  incidents: readonly IncidentDirectoryResource[];
  paging: IncidentDirectoryPaging | null;
  phase:
    | "idle"
    | "debouncing"
    | "loading"
    | "refreshing"
    | "paging"
    | "ready"
    | "failed"
    | "forbidden";
  failure: IncidentDirectoryFailure | null;
}>;
export type IncidentDirectoryPorts = {
  list: (request: {
    query: IncidentDirectoryQuery;
    limit: number;
    cursorToken: string | null;
    signal: AbortSignal;
  }) => Promise<HTTPOperationResult<ListVisibleIncidentsResponse>>;
  isCurrentSession: (identity: string) => boolean;
  sessionLost: () => void;
};
type Request = {
  identity: string;
  generation: number;
  query: IncidentDirectoryQuery;
  scope: "replace" | "page";
  limit: number;
  cursorToken: string | null;
  controller: AbortController;
  timeout?: ReturnType<typeof setTimeout>;
};

const initialState = (
  query: IncidentDirectoryQuery = { search: "", statusFilter: "all" },
): IncidentDirectoryState => ({
  query,
  acceptedQuery: null,
  incidents: [],
  paging: null,
  phase: "idle",
  failure: null,
});
const sameQuery = (
  left: IncidentDirectoryQuery,
  right: IncidentDirectoryQuery,
) => left.search === right.search && left.statusFilter === right.statusFilter;

export function directoryIsLoading(state: IncidentDirectoryState) {
  return ["debouncing", "loading", "refreshing", "paging"].includes(
    state.phase,
  );
}

export function directoryCanLoadMore(state: IncidentDirectoryState) {
  return (
    state.paging?.has_more === true &&
    state.acceptedQuery !== null &&
    sameQuery(state.query, state.acceptedQuery) &&
    (state.phase === "ready" ||
      (state.phase === "failed" &&
        state.failure?.scope === "page" &&
        !state.failure.restart))
  );
}

// One directory lifetime owns every accepted list response. Abort is an
// optimization; request identity and generation checks enforce acceptance.
export class IncidentDirectoryController {
  private state = initialState();
  private listeners = new Set<() => void>();
  private identity: string | null = null;
  private active = false;
  private generation = 0;
  private request: Request | null = null;
  private debounce: ReturnType<typeof setTimeout> | undefined;
  private completedCursors = new Set<string>();

  constructor(private readonly ports: IncidentDirectoryPorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(state: IncidentDirectoryState) {
    this.state = state;
    for (const listener of this.listeners) listener();
  }
  private invalidate() {
    this.generation += 1;
    clearTimeout(this.debounce);
    this.debounce = undefined;
    const request = this.request;
    this.request = null;
    clearTimeout(request?.timeout);
    request?.controller.abort();
  }
  setSession(identity: string | null) {
    if (this.identity === identity) return;
    this.invalidate();
    this.identity = identity;
    this.active = false;
    this.completedCursors.clear();
    this.publish(initialState());
  }
  setActive(active: boolean) {
    if (this.active === active) return;
    this.active = active;
    if (active) {
      this.refresh();
    } else {
      this.invalidate();
      this.completedCursors.clear();
      this.publish(initialState(this.state.query));
    }
  }
  dispose() {
    this.invalidate();
    this.identity = null;
    this.active = false;
    this.completedCursors.clear();
    this.publish(initialState());
  }
  private available() {
    return (
      this.active &&
      this.identity !== null &&
      this.ports.isCurrentSession(this.identity)
    );
  }
  private current(request: Request) {
    return (
      this.available() &&
      this.identity === request.identity &&
      this.generation === request.generation &&
      this.request === request &&
      !request.controller.signal.aborted
    );
  }
  changeSearch = (search: string) => {
    if (this.identity === null || search === this.state.query.search) return;
    this.invalidate();
    this.publish({
      ...this.state,
      query: { ...this.state.query, search },
      phase: this.active ? "debouncing" : "idle",
      failure: null,
    });
    if (this.available())
      this.debounce = setTimeout(() => {
        this.debounce = undefined;
        this.submit();
      }, 180);
  };
  changeStatus = (statusFilter: IncidentDirectoryQuery["statusFilter"]) => {
    if (
      this.identity === null ||
      statusFilter === this.state.query.statusFilter
    )
      return;
    this.invalidate();
    this.publish({
      ...this.state,
      query: { ...this.state.query, statusFilter },
      phase: this.active ? "debouncing" : "idle",
      failure: null,
    });
    this.submit();
  };
  submit = () => {
    if (
      this.request?.scope === "replace" &&
      sameQuery(this.request.query, this.state.query)
    )
      return;
    this.refresh();
  };
  refresh = () => {
    if (!this.available()) return;
    this.invalidate();
    this.completedCursors.clear();
    this.start("replace", this.state.query, null, 100);
  };
  loadMore = () => {
    if (
      !this.available() ||
      this.request !== null ||
      !directoryCanLoadMore(this.state)
    )
      return;
    const { acceptedQuery, paging } = this.state;
    if (
      acceptedQuery === null ||
      !paging?.has_more ||
      paging.next_cursor === null
    )
      return;
    this.start("page", acceptedQuery, paging.next_cursor, paging.limit);
  };
  retry = () => {
    if (this.state.failure?.scope === "page" && !this.state.failure.restart)
      this.loadMore();
    else this.refresh();
  };
  private start(
    scope: Request["scope"],
    query: IncidentDirectoryQuery,
    cursorToken: string | null,
    limit: number,
  ) {
    if (this.identity === null) return;
    const request: Request = {
      scope,
      query,
      cursorToken,
      limit,
      identity: this.identity,
      generation: this.generation,
      controller: new AbortController(),
    };
    // The request is installed before publishing or calling the async port.
    this.request = request;
    request.timeout = setTimeout(() => {
      if (!this.current(request)) return;
      this.request = null;
      request.controller.abort();
      this.fail(
        request,
        null,
        "The incident directory request timed out. Try again.",
      );
    }, 30_000);
    this.publish({
      ...this.state,
      phase:
        scope === "page"
          ? "paging"
          : this.state.acceptedQuery === null
            ? "loading"
            : "refreshing",
      failure: null,
    });
    void this.execute(request);
  }
  private fail(
    request: Request,
    error: APIError | null,
    message: string,
    restart = false,
  ) {
    this.publish({
      ...this.state,
      phase: "failed",
      paging: restart ? null : this.state.paging,
      failure: { error, message, scope: request.scope, restart },
    });
  }
  private invalidContinuation(request: Request) {
    this.fail(
      request,
      {
        code: "invalid_public_contract_response",
        status: 502,
        retryable: true,
      },
      "The server returned inconsistent directory continuation data. Refresh the directory.",
      true,
    );
  }
  private async execute(request: Request) {
    try {
      const result = await this.ports.list({
        query: request.query,
        cursorToken: request.cursorToken,
        limit: request.limit,
        signal: request.controller.signal,
      });
      if (!this.current(request)) return;
      this.request = null;
      if (!result.ok) {
        const error = extractError(result.payload);
        if (result.status === 401) {
          this.setSession(null);
          this.ports.sessionLost();
        } else if (
          result.status === 403 ||
          error?.code === "authorization_denied"
        ) {
          this.completedCursors.clear();
          this.publish({
            ...initialState(this.state.query),
            phase: "forbidden",
            failure: {
              error,
              message:
                "The incident directory is unavailable for this session.",
              scope: "replace",
              restart: true,
            },
          });
        } else {
          this.fail(
            request,
            error,
            request.scope === "page"
              ? "Failed to load more incidents."
              : "Failed to load incidents.",
            error?.code === "invalid_pagination_request",
          );
        }
        return;
      }
      const paging = result.payload.meta.paging;
      if (
        paging.limit !== request.limit ||
        (paging.has_more &&
          (paging.next_cursor === request.cursorToken ||
            this.completedCursors.has(paging.next_cursor)))
      ) {
        this.invalidContinuation(request);
        return;
      }
      const rows = result.payload.data.incidents;
      // Live keyset pages can overlap after authorized sort-key edits. Only
      // accepted list responses may replace payloads; first-seen order is kept.
      const incidents = new Map<string, IncidentDirectoryResource>(
        (request.scope === "page" ? this.state.incidents : []).map(
          (incident) => [incident.incident_id, incident],
        ),
      );
      for (const incident of rows)
        incidents.set(incident.incident_id, incident);
      if (request.cursorToken !== null)
        this.completedCursors.add(request.cursorToken);
      this.publish({
        ...this.state,
        acceptedQuery: request.query,
        incidents: [...incidents.values()],
        paging,
        phase: "ready",
        failure: null,
      });
    } catch {
      if (!this.current(request)) return;
      this.request = null;
      this.fail(
        request,
        null,
        "The incident directory could not be reached. Try again.",
      );
    } finally {
      clearTimeout(request.timeout);
    }
  }
}
