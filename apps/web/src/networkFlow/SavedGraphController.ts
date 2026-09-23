import {
  boundedRead,
  type ObservationClock,
} from "../services/asyncObservation";
import type { CommonJobResource } from "../services/commonJobContract";
import type {
  NetworkFlowGraphSemanticQuery,
  NetworkFlowSavedGraph,
} from "../services/networkFlowContractAdapter";
import type { NetworkFlowExtensionResourceChange } from "./networkFlowCollaborationInterpreter";
import type { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  initialSavedGraphResultState,
  SavedGraphResultNavigation,
  type SavedGraphResultPort,
  type SavedGraphResultState,
} from "./SavedGraphResultNavigation";
import {
  observeSavedGraphJob,
  type SavedGraphJobTarget,
  type SavedGraphObservation,
  savedGraphObservationTiming,
} from "./savedGraphObservation";
import {
  canMutateSavedGraph,
  canReadSavedGraphs,
  captureSavedGraphIntent,
  captureSavedGraphReceipt,
  prepareSavedGraphAttempt,
  type SavedGraphAction,
  type SavedGraphAttempt,
  type SavedGraphAuthority,
  type SavedGraphIntent,
  type SavedGraphLoadState,
  type SavedGraphReceipt,
  SavedGraphWriteError,
  sameSavedGraphAuthority,
  sameSavedGraphScope,
  savedGraphOperationGraphId,
  savedGraphWriteFailure,
} from "./savedGraphOperation";
import {
  type SavedGraphReadSurface,
  savedGraphReadDisposition,
  savedGraphReadFailure,
} from "./savedGraphReadFailure";

type SavedGraphPreparation = {
  readonly intent: SavedGraphIntent;
  readonly draft: string;
};
type SavedGraphOperation = SavedGraphPreparation &
  (
    | {
        readonly phase: "preparing";
        readonly attempt: null;
        readonly failure: null;
        readonly receipt: null;
      }
    | {
        readonly phase: "rejected" | "awaiting_review";
        readonly attempt: SavedGraphAttempt | null;
        readonly failure: SavedGraphWriteError | null;
        readonly receipt: null;
      }
    | {
        readonly phase: "submitting";
        readonly attempt: SavedGraphAttempt;
        readonly dispatch: "queued" | "sent";
        readonly failure: null;
        readonly receipt: null;
      }
    | {
        readonly phase: "uncertain";
        readonly attempt: SavedGraphAttempt;
        readonly failure: SavedGraphWriteError;
        readonly receipt: null;
      }
    | {
        readonly phase: "acknowledged";
        readonly attempt: SavedGraphAttempt;
        readonly failure: null;
        readonly receipt: SavedGraphReceipt;
      }
  );
type SavedGraphSnapshot = {
  readonly navigation: SavedGraphResultState;
  readonly observations: Readonly<Record<string, SavedGraphObservation>>;
  readonly graphs: readonly NetworkFlowSavedGraph[];
  readonly selectedGraphViewId: string | null;
  readonly listState: SavedGraphLoadState;
  readonly listError: NetworkFlowRequestError | null;
  readonly declarationErrors: Readonly<Record<string, NetworkFlowRequestError>>;
  readonly operation: SavedGraphOperation | null;
  readonly dialogOpen: boolean;
  readonly notice: string | null;
};
export type SavedGraphTransport = {
  readonly observationLimit: (signal: AbortSignal) => Promise<number>;
  readonly list: (signal: AbortSignal) => Promise<NetworkFlowSavedGraph[]>;
  readonly get: (
    graphId: string,
    signal: AbortSignal,
  ) => Promise<NetworkFlowSavedGraph>;
  readonly readJob: (
    target: SavedGraphJobTarget,
    signal: AbortSignal,
  ) => Promise<CommonJobResource>;
  readonly navigation: SavedGraphResultPort;
  readonly submit: (
    attempt: SavedGraphAttempt,
    signal: AbortSignal,
    authorizeDispatch: () => void,
  ) => Promise<SavedGraphReceipt>;
};
const initialSnapshot = (): SavedGraphSnapshot => ({
  navigation: initialSavedGraphResultState(),
  observations: {},
  graphs: [],
  selectedGraphViewId: null,
  listState: "idle",
  listError: null,
  declarationErrors: {},
  operation: null,
  dialogOpen: false,
  notice: null,
});

/** The incident workbook owns this in-memory declaration and operation lifetime. */
export class SavedGraphController {
  private state = initialSnapshot();
  readonly navigation: SavedGraphResultNavigation;
  private readonly observations = new Map<string, AbortController>();
  private readonly acceptedTargets = new Map<string, SavedGraphJobTarget>();
  private reconcilingNavigation = false;
  private publicationDepth = 0;
  private readonly listeners = new Set<() => void>();
  private authority: SavedGraphAuthority | null = null;
  private transport: SavedGraphTransport | null = null;
  private currentAuthority: (() => SavedGraphAuthority) | null = null;
  private active = false;
  private observationLimit = 1;
  private admission = false;
  private generation = 0;
  private selectionRevision = 0;
  private pendingCreationSelection: {
    graphId: string;
    revision: number;
    scope: SavedGraphAuthority;
  } | null = null;
  private listRequest: AbortController | null = null;
  private readonly graphRequests = new Map<string, AbortController>();
  private readonly pendingGraphs = new Set<string>();
  private pendingList = false;
  private writeRequest: AbortController | null = null;
  private attemptContext: {
    attempt: SavedGraphAttempt;
    selectionRevision: number;
    dispatched: boolean;
    uncertain: boolean;
  } | null = null;
  private readonly retired = new Set<string>();
  private readonly invalidatedGraphs = new Set<string>();
  private readonly removedSources = new Set<string>();
  private protectedReadWithdrawn = false;
  private authorityRemoved = false;
  private suspendedSession: {
    scope: SavedGraphAuthority;
    operation: SavedGraphOperation | null;
  } | null = null;
  private readonly drafts = new Map<string, SavedGraphPreparation>();
  constructor(
    private readonly options: {
      readonly clock?: ObservationClock;
      readonly identify?: (prefix: string) => string;
    } = {},
  ) {
    this.navigation = new SavedGraphResultNavigation(
      (navigation) => this.update({ navigation }),
      options.clock,
      (graphId) => this.withdrawGraphResult(graphId),
      (graphId, error, surface) =>
        this.handleReadFailure(error, surface, graphId),
    );
  }
  readonly subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  readonly getSnapshot = () => this.state;
  get selectedGraph(): NetworkFlowSavedGraph | null {
    return (
      this.state.graphs.find(
        (g) => g.graph_view_id === this.state.selectedGraphViewId,
      ) ?? null
    );
  }
  get mutationPending(): boolean {
    return this.admission || this.state.operation?.phase === "uncertain";
  }

  bind(
    transport: SavedGraphTransport,
    currentAuthority: () => SavedGraphAuthority,
  ): void {
    this.transport = transport;
    this.navigation.bind(transport.navigation);
    this.currentAuthority = currentAuthority;
    this.revalidateAuthority();
  }
  revalidateAuthority(): void {
    this.batch(() => this.reconcileAuthority());
  }
  private reconcileAuthority(): void {
    const next = this.currentAuthority?.() ?? null;
    if (next === null) {
      this.clear();
      return;
    }
    if (this.suspendedSession !== null) {
      const suspended = this.suspendedSession;
      if (
        next.incidentId !== suspended.scope.incidentId ||
        (next.actorId !== null && next.actorId !== suspended.scope.actorId)
      )
        this.clear();
      else if (
        next.session !== suspended.scope.session &&
        next.sessionResolved !== false &&
        next.actorId !== null &&
        !(next.profileAvailable ?? next.available)
      ) {
        this.clear();
        this.authority = next;
        return;
      } else if (
        next.session === suspended.scope.session ||
        !canReadSavedGraphs(next)
      )
        return;
      else {
        this.suspendedSession = null;
        this.authority = next;
        this.authorityRemoved = false;
        this.protectedReadWithdrawn = false;
        this.update({ operation: suspended.operation });
        if (this.active) void this.loadGraphs();
        return;
      }
    }
    if (next.sessionResolved === false || next.actorId === null) {
      this.pauseSession();
      return;
    }
    if (this.authority !== null && !sameSavedGraphScope(this.authority, next))
      this.clear();
    if (!(next.profileAvailable ?? next.available)) {
      this.clear();
      this.authority = next;
      return;
    }
    const previousAuthority = this.authority;
    const changed =
      this.authority !== null && !sameSavedGraphAuthority(this.authority, next);
    this.authority = next;
    if (changed) {
      if (canReadSavedGraphs(next)) this.authorityRemoved = false;
      this.writeRequest?.abort();
      if (
        this.attemptContext?.dispatched &&
        this.state.operation?.phase === "submitting"
      )
        this.attemptContext.uncertain = true;
      if (
        previousAuthority !== null &&
        !canReadSavedGraphs(previousAuthority) &&
        canReadSavedGraphs(next)
      )
        this.protectedReadWithdrawn = false;
      this.invalidateReads();
      this.stopObservations();
      this.navigation.stopReads();
      const operation = this.state.operation;
      if (operation !== null && operation.phase !== "acknowledged") {
        const uncertain =
          this.attemptContext?.uncertain === true && operation.attempt !== null;
        const failure = new SavedGraphWriteError(
          "authorization",
          uncertain ? "uncertain" : "rejected",
          "Authority changed. Review current access before continuing.",
        );
        this.update({
          operation:
            uncertain && operation.attempt !== null
              ? {
                  ...operation,
                  phase: "uncertain",
                  attempt: operation.attempt,
                  receipt: null,
                  failure,
                }
              : {
                  ...operation,
                  phase: "awaiting_review",
                  receipt: null,
                  failure,
                },
        });
      }
    }
    if (!canReadSavedGraphs(next)) {
      this.stopObservations();
      this.navigation.clear();
      this.invalidateReads();
      this.update({ graphs: [], selectedGraphViewId: null, listState: "idle" });
    } else if (changed) {
      this.update({ graphs: this.state.graphs });
      if (this.active) void this.loadGraphs();
    }
  }
  setActive(active: boolean): void {
    if (active === this.active) return;
    this.active = active;
    this.navigation.setActive(active);
    if (active) {
      void this.loadGraphs();
      this.observeCurrent();
    } else {
      this.invalidateReads();
      this.stopObservations();
    }
  }
  dispose(): void {
    this.clear();
    this.transport = null;
    this.currentAuthority = null;
    this.active = false;
  }
  private clear(): void {
    this.batch(() => this.clearState());
  }
  private clearState(): void {
    this.stopObservations();
    this.acceptedTargets.clear();
    this.navigation.clear();
    this.invalidateReads();
    this.writeRequest?.abort();
    this.writeRequest = null;
    this.attemptContext = null;
    this.suspendedSession = null;
    this.admission = false;
    this.authority = null;
    this.pendingGraphs.clear();
    this.pendingList = false;
    this.pendingCreationSelection = null;
    this.retired.clear();
    this.invalidatedGraphs.clear();
    this.removedSources.clear();
    this.protectedReadWithdrawn = false;
    this.authorityRemoved = false;
    this.observationLimit = 1;
    this.drafts.clear();
    this.selectionRevision++;
    this.update(initialSnapshot());
  }
  readonly pauseSession = (): void => {
    this.batch(() => this.suspendSession());
  };
  private suspendSession(): void {
    if (this.suspendedSession !== null || this.authority === null) return;
    const op = this.state.operation;
    let operation = op;
    if (op !== null && op.phase !== "acknowledged") {
      const uncertain =
        this.attemptContext?.uncertain === true ||
        (op.phase === "submitting" && this.attemptContext?.dispatched === true);
      if (this.attemptContext) this.attemptContext.uncertain = uncertain;
      const failure = new SavedGraphWriteError(
        "authorization",
        uncertain ? "uncertain" : "rejected",
        "Session paused. Reauthenticate and review current access before continuing.",
      );
      operation =
        uncertain && op.attempt !== null
          ? {
              ...op,
              phase: "uncertain",
              attempt: op.attempt,
              failure,
              receipt: null,
            }
          : { ...op, phase: "awaiting_review", failure, receipt: null };
    }
    this.suspendedSession = { scope: this.authority, operation };
    this.writeRequest?.abort();
    this.writeRequest = null;
    this.admission = false;
    this.invalidateReads();
    this.stopObservations();
    this.navigation.clear();
    this.update(initialSnapshot());
  }
  private purgeProtectedState(): void {
    this.batch(() => {
      const authority = this.authority;
      this.clear();
      this.authority = authority;
      this.protectedReadWithdrawn = true;
    });
  }
  private invalidateReads(): void {
    for (const request of this.graphRequests.values()) request.abort();
    this.graphRequests.clear();
    this.cancelListRead();
  }
  private cancelGraphRead(graphId: string): void {
    this.graphRequests.get(graphId)?.abort();
    this.graphRequests.delete(graphId);
    this.cancelListRead();
  }
  private cancelListRead(): void {
    this.generation++;
    if (this.listRequest !== null) {
      this.listRequest.abort();
      this.update({
        listState:
          this.state.listState === "refreshing"
            ? "ready"
            : this.state.listState === "loading"
              ? "idle"
              : this.state.listState,
      });
    }
    this.listRequest = null;
  }
  private batch(action: () => void): void {
    const previous = this.state;
    this.publicationDepth++;
    try {
      action();
    } finally {
      this.publicationDepth--;
      if (this.publicationDepth === 0 && previous !== this.state)
        for (const listener of this.listeners) listener();
    }
  }
  private update(patch: Partial<SavedGraphSnapshot>): void {
    this.state = { ...this.state, ...patch };
    if (
      !this.reconcilingNavigation &&
      (patch.graphs !== undefined || patch.selectedGraphViewId !== undefined)
    ) {
      this.reconcilingNavigation = true;
      const scope = this.authority;
      const graph = this.selectedGraph;
      const exposed =
        graph !== null &&
        !this.protectedReadWithdrawn &&
        !this.invalidatedGraphs.has(graph.graph_view_id) &&
        !graph.semantic_query.selected_table_ids.some((id) =>
          this.removedSources.has(id),
        );
      this.navigation.authorize(
        exposed ? graph : null,
        () => scope !== null && this.current(scope),
      );
      this.reconcilingNavigation = false;
    }
    if (this.reconcilingNavigation) return;
    if (this.publicationDepth === 0)
      for (const listener of this.listeners) listener();
    if (patch.graphs !== undefined || patch.selectedGraphViewId !== undefined)
      this.observeCurrent();
  }

  private readable(reconcileAuthority = false): SavedGraphAuthority | null {
    this.revalidateAuthority();
    return this.authority !== null &&
      this.suspendedSession === null &&
      !this.authorityRemoved &&
      (reconcileAuthority || !this.protectedReadWithdrawn) &&
      canReadSavedGraphs(this.authority)
      ? this.authority
      : null;
  }
  private current(
    captured: SavedGraphAuthority,
    reconcileAuthority = false,
  ): boolean {
    const now = this.currentAuthority?.();
    return (
      this.suspendedSession === null &&
      !this.authorityRemoved &&
      (reconcileAuthority || !this.protectedReadWithdrawn) &&
      now !== undefined &&
      sameSavedGraphAuthority(captured, now) &&
      canReadSavedGraphs(now)
    );
  }
  readonly selectGraphView = (graphId: string | null): void => {
    if (
      graphId !== null &&
      !this.state.graphs.some((graph) => graph.graph_view_id === graphId)
    )
      return;
    this.selectionRevision++;
    this.pendingCreationSelection = null;
    this.update({ selectedGraphViewId: graphId });
  };
  readonly loadGraphs = async (): Promise<void> => {
    const scope = this.readable(true),
      transport = this.transport;
    if (!this.active || scope === null || transport === null) return;
    if (this.admission) {
      this.pendingList = true;
      return;
    }
    this.invalidateReads();
    const generation = this.generation;
    const revision = this.selectionRevision;
    const request = new AbortController();
    this.listRequest = request;
    this.update({
      listState: this.state.graphs.length ? "refreshing" : "loading",
      listError: null,
    });
    try {
      const graphs = await boundedRead(
        async (signal) => {
          const [graphs, limit] = await Promise.allSettled([
            transport.list(signal),
            transport.observationLimit(signal),
          ]);
          if (graphs.status === "rejected") throw graphs.reason;
          if (generation === this.generation && this.current(scope, true))
            this.observationLimit =
              limit.status === "fulfilled" &&
              Number.isInteger(limit.value) &&
              limit.value >= 1
                ? Math.min(
                    limit.value,
                    savedGraphObservationTiming.maximumActive,
                  )
                : 1;
          return graphs.value;
        },
        request.signal,
        30_000,
        this.options.clock,
      );
      if (generation !== this.generation || !this.current(scope, true)) return;
      this.protectedReadWithdrawn = false;
      const active = sortSavedGraphs(
        graphs.filter((graph) => !this.retired.has(graph.graph_view_id)),
      );
      const selected = this.state.selectedGraphViewId;
      for (const graph of active)
        this.invalidatedGraphs.delete(graph.graph_view_id);
      this.update({
        graphs: active,
        listState: "ready",
        selectedGraphViewId:
          this.reconciledSelection(active) ??
          (selected !== null && active.some((g) => g.graph_view_id === selected)
            ? selected
            : revision === this.selectionRevision
              ? (active[0]?.graph_view_id ?? null)
              : null),
      });
    } catch (caught) {
      if (generation !== this.generation || !this.current(scope, true)) return;
      const error = savedGraphReadFailure(
        caught,
        "Saved graphs could not be read. Reload to recover current declarations.",
      );
      this.handleReadFailure(error, "list");
      this.update({ listState: "error", listError: error });
    } finally {
      if (this.listRequest === request) this.listRequest = null;
    }
  };
  readonly openAction = (
    kind: SavedGraphAction,
    query: NetworkFlowGraphSemanticQuery | null = null,
  ): void => {
    const scope = this.readable();
    if (scope === null || !canMutateSavedGraph(scope, kind)) return;
    const pending = this.state.operation;
    if (pending?.phase === "submitting" || pending?.phase === "uncertain") {
      this.update({ dialogOpen: true });
      return;
    }
    const target = kind === "create" ? null : this.selectedGraph;
    if (kind !== "create" && target === null) return;
    const key = `${kind}:${target?.graph_view_id ?? "new"}`;
    const retained = this.drafts.get(key);
    const intent =
      retained?.intent ?? captureSavedGraphIntent(kind, scope, target, query);
    const draft =
      retained?.draft ??
      (kind === "rename" ? (target?.display_name ?? "") : "");
    this.drafts.set(key, { intent, draft });
    this.update({
      dialogOpen: true,
      operation: {
        intent,
        draft,
        phase: "preparing",
        attempt: null,
        receipt: null,
        failure: null,
      },
    });
  };
  readonly prepareCurrentQuery = (
    query: NetworkFlowGraphSemanticQuery,
  ): void => {
    const op = this.state.operation,
      scope = this.readable();
    if (
      op?.intent.kind !== "create" ||
      this.mutationPending ||
      op.phase === "acknowledged" ||
      scope === null ||
      !canMutateSavedGraph(scope, "create")
    )
      return;
    const intent = captureSavedGraphIntent("create", scope, null, query);
    this.drafts.set("create:new", { intent, draft: op.draft });
    this.attemptContext = null;
    this.update({
      operation: {
        intent,
        draft: op.draft,
        phase: "preparing",
        attempt: null,
        receipt: null,
        failure: null,
      },
    });
  };
  readonly reopenOperation = (): void => {
    if (this.state.operation) this.update({ dialogOpen: true });
  };
  readonly closeDialog = (): void => {
    this.update({ dialogOpen: false });
  };
  readonly setDraft = (draft: string): void => {
    const op = this.state.operation;
    if (
      op === null ||
      op.phase === "submitting" ||
      op.phase === "uncertain" ||
      op.phase === "acknowledged"
    )
      return;
    this.drafts.set(
      `${op.intent.kind}:${op.intent.target?.graph_view_id ?? "new"}`,
      { intent: op.intent, draft },
    );
    this.update({
      operation: {
        ...op,
        draft,
        attempt: null,
        failure: null,
        phase: op.phase === "awaiting_review" ? op.phase : "preparing",
      },
    });
  };
  readonly submit = async (): Promise<boolean> => {
    if (
      this.admission ||
      this.state.operation?.phase === "uncertain" ||
      this.state.operation?.phase === "awaiting_review"
    )
      return false;
    const op = this.state.operation;
    if (
      op === null ||
      op.phase === "acknowledged" ||
      !this.admitIntent(op.intent)
    )
      return false;
    this.admission = true; // Synchronous, before ID generation or asynchronous work.
    let attempt: SavedGraphAttempt;
    try {
      attempt = prepareSavedGraphAttempt(
        op.intent,
        op.draft,
        this.options.identify,
      );
    } catch (caught) {
      this.admission = false;
      this.update({
        operation: {
          ...op,
          phase: "rejected",
          receipt: null,
          failure:
            caught instanceof SavedGraphWriteError
              ? caught
              : new SavedGraphWriteError(
                  "validation",
                  "rejected",
                  "A secure transaction identifier could not be generated. Nothing was submitted.",
                ),
        },
      });
      return false;
    }
    this.attemptContext = {
      attempt,
      selectionRevision: this.selectionRevision,
      dispatched: false,
      uncertain: false,
    };
    return this.send(attempt);
  };
  private admitIntent(intent: SavedGraphIntent): boolean {
    const authority = this.readable();
    const op = this.state.operation;
    if (op === null) return false;
    if (
      authority === null ||
      !sameSavedGraphAuthority(intent.authority, authority) ||
      !canMutateSavedGraph(authority, intent.kind) ||
      !this.intentContextCurrent(intent)
    ) {
      this.update({
        operation: {
          ...op,
          phase: "awaiting_review",
          receipt: null,
          failure: new SavedGraphWriteError(
            "version_conflict",
            "rejected",
            "The captured graph or authority changed. Review its current state before submitting.",
          ),
        },
      });
      return false;
    }
    return true;
  }
  private intentContextCurrent(intent: SavedGraphIntent): boolean {
    const sources =
      intent.query?.selected_table_ids ??
      intent.target?.semantic_query.selected_table_ids ??
      [];
    if (sources.some((id) => this.removedSources.has(id))) return false;
    if (intent.target === null) return true;
    const target = this.state.graphs.find(
      (g) => g.graph_view_id === intent.target?.graph_view_id,
    );
    return (
      target !== undefined &&
      !this.retired.has(target.graph_view_id) &&
      !this.invalidatedGraphs.has(target.graph_view_id) &&
      target.graph_view_version === intent.target.graph_view_version &&
      target.display_name === intent.target.display_name &&
      target.semantic_query_sha256 === intent.target.semantic_query_sha256
    );
  }
  readonly replay = async (): Promise<boolean> => {
    const op = this.state.operation;
    if (this.admission || op?.phase !== "uncertain" || op.attempt === null)
      return false;
    const authority = this.readable();
    if (
      authority === null ||
      op.intent.authority.incidentId !== authority.incidentId ||
      op.intent.authority.actorId !== authority.actorId ||
      !canMutateSavedGraph(authority, op.intent.kind)
    )
      return false;
    this.admission = true;
    return this.send(op.attempt);
  };
  private async send(attempt: SavedGraphAttempt): Promise<boolean> {
    const op = this.state.operation,
      transport = this.transport,
      authority = this.readable();
    if (op === null || transport === null || authority === null) {
      this.admission = false;
      return false;
    }
    this.invalidateReads();
    const context = this.attemptContext;
    if (context?.attempt !== attempt) {
      this.admission = false;
      return false;
    }
    const recovering = context.uncertain;
    this.update({
      operation: {
        ...op,
        attempt,
        phase: "submitting",
        dispatch: "queued",
        receipt: null,
        failure: null,
      },
    });
    const request = new AbortController();
    this.writeRequest = request;
    try {
      await boundedRead(
        async (signal) => {
          const receipt = await transport.submit(attempt, signal, () => {
            this.revalidateAuthority();
            signal.throwIfAborted();
            if (
              this.state.operation?.attempt !== attempt ||
              this.state.operation.phase !== "submitting" ||
              this.writeRequest !== request ||
              !this.current(authority) ||
              !canMutateSavedGraph(authority, attempt.intent.kind)
            )
              throw new SavedGraphWriteError(
                "authorization",
                "rejected",
                "The captured authority changed before dispatch. Review current access.",
              );
            if (!recovering && !this.intentContextCurrent(attempt.intent))
              throw new SavedGraphWriteError(
                "version_conflict",
                "rejected",
                "The captured target changed before dispatch. Review its current state.",
              );
            context.dispatched = true;
            this.update({
              operation: {
                intent: attempt.intent,
                draft: op.draft,
                attempt,
                phase: "submitting",
                dispatch: "sent",
                failure: null,
                receipt: null,
              },
            });
          });
          // A late acknowledgement can still resolve this exact uncertain attempt.
          if (
            this.state.operation?.attempt === attempt &&
            this.current(authority)
          )
            this.acceptReceipt(attempt, receipt, context.selectionRevision);
          return receipt;
        },
        request.signal,
        30_000,
        this.options.clock,
      );
      return (
        this.state.operation?.attempt === attempt &&
        this.state.operation.phase === "acknowledged"
      );
    } catch (caught) {
      const current = this.state.operation;
      if (current?.attempt !== attempt || current.phase === "acknowledged")
        return current?.phase === "acknowledged";
      if (this.writeRequest !== request) return false;
      const classified =
        !context.dispatched &&
        current.phase === "awaiting_review" &&
        current.failure !== null
          ? current.failure
          : savedGraphWriteFailure(caught);
      context.uncertain ||=
        recovering ||
        (context.dispatched && classified.certainty === "uncertain");
      const failure = new SavedGraphWriteError(
        classified.category,
        context.uncertain ? "uncertain" : "rejected",
        !context.dispatched && classified.certainty === "uncertain"
          ? "The request was stopped before dispatch. Nothing was submitted. Review current access before trying again."
          : classified.message,
        classified.detail,
      );
      this.update({
        operation: {
          ...current,
          attempt,
          receipt: null,
          failure,
          phase: context.uncertain
            ? "uncertain"
            : current.phase === "awaiting_review" ||
                failure.category === "version_conflict" ||
                failure.category === "authorization"
              ? "awaiting_review"
              : "rejected",
        },
      });
      return false;
    } finally {
      if (this.writeRequest === request) {
        this.writeRequest = null;
        this.admission = false;
        void this.reconcilePending();
      }
    }
  }
  private acceptReceipt(
    attempt: SavedGraphAttempt,
    receipt: SavedGraphReceipt,
    selectionRevision: number,
  ): void {
    const op = this.state.operation;
    if (op?.attempt !== attempt || op.phase === "acknowledged") return;
    receipt = captureSavedGraphReceipt(receipt);
    this.invalidateReads();
    this.admission = false;
    let graphs = this.state.graphs,
      selected = this.state.selectedGraphViewId;
    if (receipt.kind === "retired") {
      const id = attempt.intent.target?.graph_view_id;
      if (id !== undefined) {
        this.retired.add(id);
        for (const [jobId, target] of this.acceptedTargets)
          if (target.graphId === id) this.acceptedTargets.delete(jobId);
        graphs = graphs.filter((g) => g.graph_view_id !== id);
        if (selected === id) selected = null;
      }
    }
    if (receipt.kind === "accepted") {
      if (
        attempt.intent.kind === "create" &&
        selectionRevision === this.selectionRevision
      )
        this.pendingCreationSelection = {
          graphId: receipt.value.graph_view.graph_view_id,
          revision: selectionRevision,
          scope: attempt.intent.authority,
        };
      const target = {
        incidentId: attempt.intent.authority.incidentId,
        graphId: receipt.value.graph_view.graph_view_id,
        jobId: receipt.value.job.job_id,
        statusRoute: receipt.value.job.status_route,
      };
      this.acceptedTargets.set(target.jobId, target);
    }
    const name =
      receipt.kind === "retired"
        ? attempt.intent.target?.display_name
        : receipt.kind === "accepted"
          ? receipt.value.graph_view.display_name
          : receipt.value.display_name;
    this.update({
      graphs,
      selectedGraphViewId: selected,
      operation: {
        ...op,
        attempt,
        phase: "acknowledged",
        receipt,
        failure: null,
      },
      dialogOpen: false,
      notice:
        receipt.kind === "accepted"
          ? `${name}: ${attempt.intent.kind === "create" ? "saved graph created" : "refresh accepted"}.`
          : `${name}: ${receipt.kind === "retired" ? "retired" : "rename acknowledged"}.`,
    });
    this.drafts.delete(
      `${attempt.intent.kind}:${attempt.intent.target?.graph_view_id ?? "new"}`,
    );
    // Receipts acknowledge writes. Only current reads can install declarations.
    if (receipt.kind !== "retired") {
      const graph =
        receipt.kind === "accepted" ? receipt.value.graph_view : receipt.value;
      void this.reconcileGraph(graph.graph_view_id);
    } else void this.loadGraphs(); // Follow-up failure cannot change the committed receipt.
  }
  readonly reviewCurrent = async (): Promise<void> => {
    const op = this.state.operation,
      scope = this.readable(),
      transport = this.transport;
    if (op?.phase !== "awaiting_review" || scope === null || transport === null)
      return;
    try {
      const graph =
        op.intent.target === null
          ? null
          : await this.readGraph(op.intent.target.graph_view_id);
      if (
        this.state.operation !== op ||
        !this.current(scope) ||
        (op.intent.target !== null &&
          (graph === null ||
            this.retired.has(op.intent.target.graph_view_id) ||
            this.state.graphs.find(
              (g) => g.graph_view_id === graph.graph_view_id,
            ) !== graph))
      )
        return;
      const intent = captureSavedGraphIntent(
        op.intent.kind,
        scope,
        graph,
        op.intent.query,
      );
      this.drafts.set(
        `${intent.kind}:${intent.target?.graph_view_id ?? "new"}`,
        { intent, draft: op.draft },
      );
      this.update({
        operation: {
          ...op,
          intent,
          phase: "preparing",
          attempt: null,
          receipt: null,
          failure: null,
        },
        notice:
          "Current state loaded. Review the target and submit again to make a new attempt.",
      });
    } catch (caught) {
      if (this.state.operation === op && this.current(scope))
        this.update({
          operation: {
            ...op,
            failure: savedGraphWriteFailure(
              savedGraphReadFailure(
                caught,
                "Current graph state could not be read. Retry review.",
              ),
            ),
          },
        });
    }
  };
  private stopObservations(): void {
    for (const stop of this.observations.values()) stop.abort();
    this.observations.clear();
    this.update({
      observations: Object.fromEntries(
        Object.entries(this.state.observations).map(([id, observation]) => [
          id,
          observation.state === "observing"
            ? {
                ...observation,
                state: "paused" as const,
                message:
                  "Observation stopped. Server work may continue; resume when ready.",
              }
            : observation,
        ]),
      ),
    });
  }
  private observeCurrent(): void {
    if (
      !this.active ||
      this.authority === null ||
      !canReadSavedGraphs(this.authority)
    )
      return;
    const selected = this.selectedGraph;
    const targets = [...this.acceptedTargets.values()].filter(
      (target) => !this.retired.has(target.graphId),
    );
    if (selected?.latest_job_id)
      targets.unshift({
        incidentId: selected.incident_id,
        graphId: selected.graph_view_id,
        jobId: selected.latest_job_id,
        statusRoute: `/api/v1/jobs/${selected.latest_job_id}`,
      });
    for (const [jobId, stop] of this.observations) {
      if (!targets.some((target) => target.jobId === jobId)) {
        stop.abort();
        this.observations.delete(jobId);
        const observation = this.state.observations[jobId];
        if (observation)
          this.update({
            observations: {
              ...this.state.observations,
              [jobId]: {
                ...observation,
                state: "paused",
                message:
                  "Observation stopped after navigation. Server work may continue.",
              },
            },
          });
      }
    }
    const retainedJobs = new Set([
      ...this.state.graphs.flatMap((graph) =>
        graph.latest_job_id === null ? [] : [graph.latest_job_id],
      ),
      ...targets.map((target) => target.jobId),
    ]);
    const observations = Object.entries(this.state.observations).filter(
      ([id]) => retainedJobs.has(id),
    );
    if (observations.length !== Object.keys(this.state.observations).length)
      this.update({ observations: Object.fromEntries(observations) });
    for (const target of targets)
      if (this.state.observations[target.jobId] === undefined)
        void this.observe(target);
  }
  readonly reloadOperationGraph = async (): Promise<void> => {
    const graphId = savedGraphOperationGraphId(this.state.operation);
    if (graphId !== null) await this.reconcileGraph(graphId);
  };
  readonly resumeOperationObservation = (): void => {
    const receipt = this.state.operation?.receipt;
    if (receipt?.kind === "accepted")
      this.resumeObservation(receipt.value.job.job_id);
  };
  readonly resumeObservation = (jobId?: string): void => {
    const selected = this.selectedGraph;
    const id = jobId ?? selected?.latest_job_id;
    if (!id) return;
    const known = this.state.observations[id];
    if (known) void this.observe(known.target);
    else if (this.acceptedTargets.has(id)) {
      const target = this.acceptedTargets.get(id);
      if (target) void this.observe(target);
    } else if (selected?.latest_job_id === id)
      void this.observe({
        incidentId: selected.incident_id,
        graphId: selected.graph_view_id,
        jobId: id,
        statusRoute: `/api/v1/jobs/${id}`,
      });
  };
  private async observe(target: SavedGraphJobTarget): Promise<void> {
    const scope = this.authority,
      transport = this.transport;
    if (
      !this.active ||
      scope === null ||
      !this.current(scope) ||
      transport === null ||
      this.observations.has(target.jobId) ||
      this.observations.size >= this.observationLimit ||
      this.retired.has(target.graphId)
    )
      return;
    const stop = new AbortController();
    this.observations.set(target.jobId, stop);
    const previous = this.state.observations[target.jobId]?.job ?? null;
    this.update({
      observations: {
        ...this.state.observations,
        [target.jobId]: {
          target,
          state: "observing",
          job: previous,
          message: null,
        },
      },
    });
    const result = await observeSavedGraphJob({
      target,
      previous,
      read: (target, signal) => {
        if (!this.current(scope))
          throw new Error("Saved graph authority changed.");
        return transport.readJob(target, signal);
      },
      signal: stop.signal,
      clock: this.options.clock,
      publish: (job) => {
        if (this.observations.get(target.jobId) === stop && this.current(scope))
          this.update({
            observations: {
              ...this.state.observations,
              [target.jobId]: {
                target,
                state: "observing",
                job,
                message: null,
              },
            },
          });
      },
    });
    if (this.observations.get(target.jobId) !== stop || !this.current(scope))
      return;
    this.observations.delete(target.jobId);
    if (result.failure)
      this.handleReadFailure(result.failure, "job", target.graphId);
    if (!this.current(scope)) return;
    this.update({
      observations: { ...this.state.observations, [target.jobId]: result },
    });
    if (result.state === "terminal") {
      // Retain the immutable accepted target for recovery independently of selection.
      // Its terminal observation prevents automatic polling.
      // A terminal job is execution evidence; only the current declaration grants result exposure.
      await this.reconcileGraph(target.graphId);
    }
    this.observeCurrent(); // Fill a freed slot without restarting paused observations.
  }
  readonly reconcileGraph = async (graphId: string): Promise<void> => {
    await this.readGraph(graphId);
  };
  private async readGraph(
    graphId: string,
  ): Promise<NetworkFlowSavedGraph | null> {
    const scope = this.readable(),
      transport = this.transport;
    if (scope === null || transport === null || this.retired.has(graphId))
      return null;
    if (this.admission) {
      this.pendingGraphs.add(graphId);
      return null;
    }
    this.cancelGraphRead(graphId);
    const request = new AbortController();
    this.graphRequests.set(graphId, request);
    const declarationErrors = { ...this.state.declarationErrors };
    delete declarationErrors[graphId];
    this.update({ declarationErrors });
    const current = () =>
      this.graphRequests.get(graphId) === request &&
      this.current(scope) &&
      !this.retired.has(graphId);
    try {
      const graph = await boundedRead(
        (signal) => transport.get(graphId, signal),
        request.signal,
        30_000,
        this.options.clock,
      );
      if (!current()) return null;
      if (
        graph.graph_view_id !== graphId ||
        graph.incident_id !== scope.incidentId ||
        graph.state !== "active"
      )
        throw new SyntaxError(
          "Current declaration does not match its authorized target.",
        );
      this.cancelListRead();
      this.invalidatedGraphs.delete(graphId);
      for (const [jobId, target] of this.acceptedTargets) {
        if (
          target.graphId === graphId &&
          this.state.observations[jobId]?.state === "terminal"
        )
          this.acceptedTargets.delete(jobId);
      }
      this.update({
        graphs: sortSavedGraphs([
          ...this.state.graphs.filter((g) => g.graph_view_id !== graphId),
          graph,
        ]),
        selectedGraphViewId:
          this.reconciledSelection([graph]) ?? this.state.selectedGraphViewId,
        listError: null,
      });
      return graph;
    } catch (caught) {
      if (!current()) return null;
      const error = savedGraphReadFailure(
        caught,
        "The current declaration could not be read. Reload to recover.",
      );
      this.handleReadFailure(error, "declaration", graphId);
      const disposition = savedGraphReadDisposition(error, "declaration");
      if (disposition === "scope" || disposition === "declaration") return null;
      const op = this.state.operation;
      this.update({
        declarationErrors: {
          ...this.state.declarationErrors,
          [graphId]: error,
        },
        listError: error,
        listState: "error",
        ...(op?.phase === "awaiting_review" &&
        op.intent.target?.graph_view_id === graphId
          ? {
              operation: { ...op, failure: savedGraphWriteFailure(error) },
            }
          : {}),
      });
      return null;
    } finally {
      if (this.graphRequests.get(graphId) === request)
        this.graphRequests.delete(graphId);
    }
  }
  private reconciledSelection(
    graphs: readonly NetworkFlowSavedGraph[],
  ): string | null {
    const pending = this.pendingCreationSelection;
    if (
      pending === null ||
      pending.revision !== this.selectionRevision ||
      this.authority === null ||
      !sameSavedGraphScope(pending.scope, this.authority) ||
      !this.current(this.authority) ||
      !graphs.some((graph) => graph.graph_view_id === pending.graphId)
    )
      return null;
    this.pendingCreationSelection = null;
    return pending.graphId;
  }
  private async reconcilePending(): Promise<void> {
    if (this.admission) return;
    const list = this.pendingList;
    const targets = [...this.pendingGraphs];
    this.pendingList = false;
    this.pendingGraphs.clear();
    if (list) await this.loadGraphs();
    await Promise.all(targets.map((id) => this.reconcileGraph(id)));
  }
  readonly recoverResult = async (): Promise<void> => {
    const graphId = this.state.selectedGraphViewId;
    if (graphId === null) {
      await this.loadGraphs();
      return;
    }
    if (
      this.state.navigation.identity === null ||
      this.invalidatedGraphs.has(graphId)
    ) {
      await this.reconcileGraph(graphId); // Authorization installs and loads only the newly observed binding.
    } else await this.navigation.loadResult();
  };
  private handleReadFailure(
    error: NetworkFlowRequestError,
    surface: SavedGraphReadSurface,
    graphId?: string,
  ): void {
    this.batch(() => this.applyReadFailure(error, surface, graphId));
  }
  private applyReadFailure(
    error: NetworkFlowRequestError,
    surface: SavedGraphReadSurface,
    graphId?: string,
  ): void {
    switch (savedGraphReadDisposition(error, surface)) {
      case "scope":
        if (error.code === "session_required") this.pauseSession();
        else if (error.code === "incident_closed") {
          this.writeRequest?.abort();
          this.protectedReadWithdrawn = true;
          this.invalidateReads();
          this.stopObservations();
          this.navigation.clear();
          this.update({ graphs: [], selectedGraphViewId: null });
        } else this.purgeProtectedState();
        this.update({ listState: "error", listError: error });
        break;
      case "declaration":
        if (graphId !== undefined) this.removeGraph(graphId);
        break;
      case "binding":
        if (graphId !== undefined) this.withdrawGraphResult(graphId);
        break;
      case "retain":
        break;
    }
  }

  readonly onResourceChange = (
    change: NetworkFlowExtensionResourceChange,
  ): Promise<void> => {
    let pending: Promise<void> = Promise.resolve();
    this.batch(() => {
      pending = this.applyResourceChange(change);
    });
    return pending;
  };
  private async applyResourceChange(
    change: NetworkFlowExtensionResourceChange,
  ): Promise<void> {
    if (change.resourceKind === "*") {
      if (change.changeKind === "remove") {
        if (change.reasonCode === "session_revoked") this.pauseSession();
        else if (change.reasonCode === "incident_closed") {
          this.writeRequest?.abort();
          this.invalidateReads();
          this.stopObservations();
          this.navigation.clear();
          this.protectedReadWithdrawn = true;
          this.authorityRemoved = true;
          this.update({ graphs: [], selectedGraphViewId: null });
        } else {
          this.purgeProtectedState();
          this.authorityRemoved = true;
        }
        return;
      }
      this.invalidateReads();
      this.stopObservations();
      this.navigation.clear();
      await this.loadGraphs();
      return;
    }
    if (change.resourceKind === "network_flow_graph_view") {
      if (change.changeKind === "remove") {
        this.removeGraph(change.resourceId);
        await this.loadGraphs();
        return;
      }
      // An invalidation withdraws current exposure until the declaration is observed.
      if (change.reasonCode === "source_invalidated")
        this.withdrawGraphResult(change.resourceId);
      await this.reconcileGraph(change.resourceId);
      return;
    }
    const affected = this.state.graphs.filter((graph) =>
      graph.semantic_query.selected_table_ids.includes(change.resourceId),
    );
    if (change.changeKind === "remove") {
      this.removedSources.add(change.resourceId);
      for (const [key, draft] of this.drafts) {
        if (
          (
            draft.intent.query ?? draft.intent.target?.semantic_query
          )?.selected_table_ids.includes(change.resourceId)
        )
          this.drafts.delete(key);
      }
      const op = this.state.operation;
      if (
        op !== null &&
        (
          op.intent.query ?? op.intent.target?.semantic_query
        )?.selected_table_ids.includes(change.resourceId)
      ) {
        this.writeRequest?.abort();
        this.writeRequest = null;
        this.attemptContext = null;
        this.admission = false;
        this.update({ operation: null, dialogOpen: false, notice: null });
      }
      for (const graph of affected)
        this.withdrawGraphResult(graph.graph_view_id);
      await this.loadGraphs();
    }
    // Table rename changes only labels; immutable result and contributor identities remain valid.
  }
  private withdrawGraphResult(graphId: string): void {
    this.batch(() => {
      this.cancelGraphRead(graphId);
      this.invalidatedGraphs.add(graphId);
      if (this.state.selectedGraphViewId === graphId)
        this.navigation.authorize(null, () => false);
    });
  }

  readonly removeGraph = (id: string): void => {
    this.batch(() => this.removeGraphState(id));
  };
  private removeGraphState(id: string): void {
    this.retired.add(id);
    for (const [jobId, target] of this.acceptedTargets)
      if (target.graphId === id) this.acceptedTargets.delete(jobId);
    for (const [key, draft] of this.drafts) {
      if (draft.intent.target?.graph_view_id === id) this.drafts.delete(key);
    }
    const operationTargetsResource =
      savedGraphOperationGraphId(this.state.operation) === id;
    if (operationTargetsResource) {
      this.writeRequest?.abort();
      this.writeRequest = null;
      this.attemptContext = null;
      this.admission = false;
    }
    for (const [jobId, observation] of Object.entries(
      this.state.observations,
    )) {
      if (observation.target.graphId === id) {
        this.observations.get(jobId)?.abort();
        this.observations.delete(jobId);
      }
    }
    this.pendingGraphs.delete(id);
    this.cancelGraphRead(id);
    this.update({
      ...(operationTargetsResource
        ? { operation: null, dialogOpen: false, notice: null }
        : {}),
      observations: Object.fromEntries(
        Object.entries(this.state.observations).filter(
          ([, observation]) => observation.target.graphId !== id,
        ),
      ),
      declarationErrors: Object.fromEntries(
        Object.entries(this.state.declarationErrors).filter(
          ([graphId]) => graphId !== id,
        ),
      ),
      graphs: this.state.graphs.filter((g) => g.graph_view_id !== id),
      selectedGraphViewId:
        this.state.selectedGraphViewId === id
          ? null
          : this.state.selectedGraphViewId,
    });
  }
}

/** Unicode code-point ordering, independent of browser locale and UTF-16 pairs. */
export function sortSavedGraphs(
  graphs: readonly NetworkFlowSavedGraph[],
): NetworkFlowSavedGraph[] {
  const compare = (a: string, b: string): number => {
    const left = Array.from(a, (c) => c.codePointAt(0) ?? 0),
      right = Array.from(b, (c) => c.codePointAt(0) ?? 0);
    for (let i = 0; i < Math.min(left.length, right.length); i++) {
      const delta = (left[i] ?? 0) - (right[i] ?? 0);
      if (delta) return delta;
    }
    return left.length - right.length;
  };
  return [...graphs].sort(
    (a, b) =>
      compare(a.display_name, b.display_name) ||
      compare(a.graph_view_id, b.graph_view_id),
  );
}
