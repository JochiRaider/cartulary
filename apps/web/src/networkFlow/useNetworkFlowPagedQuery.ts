import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { ExtensionAvailabilityUnavailableError } from "../extensions/extensionAvailability";
import { boundedRead, ObservationStopped } from "../services/asyncObservation";
import type { NetworkFlowPaging } from "../services/networkFlowContractAdapter";
import {
  isNetworkFlowAuthorizationLoss,
  isNetworkFlowProtectedStateLoss,
  type NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";

export type NetworkFlowQueryLoadState =
  | "idle"
  | "loading"
  | "refreshing"
  | "ready"
  | "error";
type PageCommand = "initial" | "next" | "previous" | "refresh" | "restart";
type NetworkFlowPage<Item, Metadata = unknown> = {
  readonly items: readonly Item[];
  readonly paging: NetworkFlowPaging;
  readonly metadata?: Metadata;
};
type CommittedNetworkFlowPage<
  Item,
  Request,
  Metadata = unknown,
> = NetworkFlowPage<Item, Metadata> & {
  readonly request: Request;
  readonly initialRequest: Request;
  readonly contextKey: string;
  readonly pageNumber: number;
};
type NetworkFlowPageAttempt<Request> = {
  readonly command: PageCommand;
  readonly destination: number;
  readonly request: Request;
  readonly initialRequest: Request;
  readonly contextKey: string;
  readonly generation: number;
  readonly notifyQuery: boolean;
  readonly automaticRestart: boolean;
};
type NetworkFlowPageRecovery =
  | "retry"
  | "restart"
  | "refresh_resource"
  | "correct_request"
  | "reduce_scope"
  | "none";

type Options<Item, Request, Metadata> = {
  readonly active?: boolean;
  readonly enabled: boolean;
  readonly fetchPage: (
    request: Request,
    signal: AbortSignal,
  ) => Promise<NetworkFlowPage<Item, Metadata>>;
  readonly initialRequest: Request;
  readonly isContinuation: (request: Request) => boolean;
  readonly makeContinuation: (cursorToken: string) => Request;
  readonly onError: (error: NetworkFlowRequestError | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly onQueryResult?: (error: NetworkFlowRequestError | null) => void;
  readonly onProtectedStateLoss?: (error: NetworkFlowRequestError) => void;
  readonly onGraphStale?: () => void;
  readonly queryKey: string;
  readonly readIdentity?: string | null | undefined;
  readonly isCurrent?: () => boolean;
  readonly validatePage?: (
    page: NetworkFlowPage<Item, Metadata>,
    previous: CommittedNetworkFlowPage<Item, Request, Metadata> | null,
    request: Request,
  ) => void;
  readonly reconcile: (
    previous: readonly Item[],
    incoming: readonly Item[],
  ) => Item[];
};
type State<Item, Request, Metadata> = {
  readonly contextKey: string;
  readonly committed: CommittedNetworkFlowPage<Item, Request, Metadata> | null;
  readonly pending: NetworkFlowPageAttempt<Request> | null;
  readonly failed: NetworkFlowPageAttempt<Request> | null;
  readonly history: readonly Request[];
  readonly cursorChainInvalid: boolean;
  readonly loadState: NetworkFlowQueryLoadState;
  readonly error: NetworkFlowRequestError | null;
  readonly notice: string | null;
};
const empty = <Item, Request, Metadata>(
  contextKey: string,
): State<Item, Request, Metadata> => ({
  contextKey,
  committed: null,
  pending: null,
  failed: null,
  history: [],
  cursorChainInvalid: false,
  loadState: "idle",
  error: null,
  notice: null,
});

/** A single committed page, request-only history, and one admitted read attempt. */
export function useNetworkFlowPagedQuery<Item, Request, Metadata = unknown>(
  options: Options<Item, Request, Metadata>,
) {
  const contextKey = JSON.stringify([options.queryKey, options.readIdentity]);
  const latest = useRef({ options, contextKey });
  latest.current = { options, contextKey };
  const generation = useRef(0);
  const controller = useRef<AbortController | null>(null);
  const paused = useRef(options.active === false);
  paused.current = options.active === false;
  const [stored, setStored] = useState(() =>
    empty<Item, Request, Metadata>(contextKey),
  );
  const current = useRef(stored);
  const publish = useCallback((state: State<Item, Request, Metadata>) => {
    current.current = state;
    setStored(state);
  }, []);
  const cancel = useCallback(() => {
    generation.current++;
    controller.current?.abort();
    controller.current = null;
  }, []);

  const execute = useCallback(
    async (
      input: Omit<NetworkFlowPageAttempt<Request>, "generation" | "contextKey">,
    ) => {
      const captured = latest.current;
      if (
        paused.current ||
        !captured.options.enabled ||
        captured.options.isCurrent?.() === false
      )
        return;
      const before = current.current;
      if (before.pending !== null) {
        if (input.command !== "restart" || before.pending.command === "restart")
          return;
        cancel();
      }
      const stop = new AbortController();
      controller.current = stop;
      const activeGeneration = ++generation.current;
      let attempt: NetworkFlowPageAttempt<Request> = {
        ...input,
        request: structuredClone(input.request),
        initialRequest: structuredClone(input.initialRequest),
        generation: activeGeneration,
        contextKey: captured.contextKey,
      };
      const accepts = () =>
        !paused.current &&
        !stop.signal.aborted &&
        generation.current === activeGeneration &&
        latest.current.contextKey === captured.contextKey &&
        latest.current.options.enabled &&
        captured.options.isCurrent?.() !== false;
      publish({
        ...before,
        contextKey: captured.contextKey,
        pending: attempt,
        failed: null,
        error: null,
        loadState: before.committed === null ? "loading" : "refreshing",
        notice:
          attempt.command === "restart"
            ? "Restarting the query at page one."
            : null,
      });
      captured.options.onError(null);
      for (;;) {
        try {
          const page = await boundedRead(
            (signal) => captured.options.fetchPage(attempt.request, signal),
            stop.signal,
          );
          if (!accepts()) return;
          const state = current.current;
          captured.options.validatePage?.(
            page,
            state.committed,
            attempt.request,
          );
          const items = captured.options.reconcile(
            state.committed?.items ?? [],
            page.items,
          );
          const history =
            attempt.destination === 1
              ? [attempt.request]
              : [
                  ...state.history.slice(0, attempt.destination - 1),
                  attempt.request,
                ];
          publish({
            ...state,
            committed: {
              ...page,
              items,
              request: attempt.request,
              initialRequest: attempt.initialRequest,
              contextKey: attempt.contextKey,
              pageNumber: attempt.destination,
            },
            history,
            pending: null,
            failed: null,
            error: null,
            loadState: "ready",
            cursorChainInvalid: false,
            notice: attempt.automaticRestart
              ? "The page cursor expired. Results restarted at page one."
              : attempt.command === "restart"
                ? "Results restarted at page one."
                : null,
          });
          if (attempt.notifyQuery && accepts())
            captured.options.onQueryResult?.(null);
          if (accepts()) captured.options.onError(null);
          return;
        } catch (caught) {
          if (!accepts()) return;
          if (
            caught instanceof ExtensionAvailabilityUnavailableError ||
            (caught instanceof ObservationStopped &&
              caught.reason === "aborted")
          ) {
            const state = current.current;
            publish({
              ...state,
              pending: null,
              loadState: state.committed === null ? "idle" : "ready",
            });
            return;
          }
          const error = networkFlowErrorFromUnknown(
            caught,
            caught instanceof ObservationStopped
              ? "The page read timed out. Try again."
              : "Network Flow query failed.",
          );
          const protectedLoss = isNetworkFlowProtectedStateLoss(error);
          const graphStale = error.code === "network_flow_graph_query_stale";
          const sourceStale =
            error.code === "network_flow_cursor_invalid" &&
            error.reasonCode === "scope_stale";
          if (protectedLoss || graphStale || sourceStale) {
            publish({
              ...empty<Item, Request, Metadata>(captured.contextKey),
              error,
              failed: attempt,
              loadState: "error",
            });
            if (graphStale) captured.options.onGraphStale?.();
            else captured.options.onProtectedStateLoss?.(error);
            if (isNetworkFlowAuthorizationLoss(error) && accepts())
              captured.options.onIncidentAccessLost?.();
            // The owner may synchronously invalidate this attempt while clearing exposure.
            if (accepts()) captured.options.onError(error);
            return;
          }
          const cursorInvalid = error.code === "network_flow_cursor_invalid";
          if (cursorInvalid)
            publish({ ...current.current, cursorChainInvalid: true });
          if (
            cursorInvalid &&
            error.reasonCode === "expired" &&
            error.retryAction === "restart_query" &&
            captured.options.isContinuation(attempt.request) &&
            !attempt.automaticRestart
          ) {
            attempt = {
              ...attempt,
              command: "restart",
              destination: 1,
              request: attempt.initialRequest,
              automaticRestart: true,
              notifyQuery: false,
            };
            publish({
              ...current.current,
              pending: attempt,
              notice: "The page cursor expired. Restarting at page one.",
            });
            if (!accepts()) return;
            continue;
          }
          publish({
            ...current.current,
            pending: null,
            failed: attempt,
            error,
            loadState: "error",
            notice:
              attempt.command === "restart"
                ? "The page-one restart failed."
                : null,
          });
          if (attempt.notifyQuery && accepts())
            captured.options.onQueryResult?.(error);
          if (accepts()) captured.options.onError(error);
          return;
        } finally {
          if (
            generation.current === activeGeneration &&
            !accepts() &&
            current.current.pending?.generation === activeGeneration
          ) {
            const state = current.current;
            publish({
              ...state,
              pending: null,
              loadState: state.committed === null ? "idle" : "ready",
            });
          }
        }
      }
    },
    [cancel, publish],
  );

  useLayoutEffect(() => {
    cancel();
    publish(empty<Item, Request, Metadata>(contextKey));
    if (options.enabled) {
      const initialRequest = latest.current.options.initialRequest;
      void execute({
        command: "initial",
        destination: 1,
        request: initialRequest,
        initialRequest,
        automaticRestart: false,
        notifyQuery: true,
      });
    }
    return cancel;
  }, [cancel, contextKey, execute, options.enabled, publish]);

  // Visibility suspends observation, independently of query/authority ownership.
  const pause = useCallback(() => {
    paused.current = true;
    cancel();
    const state = current.current;
    if (state.pending !== null)
      publish({
        ...state,
        pending: null,
        loadState: state.committed === null ? "idle" : "ready",
        notice: null,
      });
  }, [cancel, publish]);
  useLayoutEffect(() => {
    if (options.active === false) {
      pause();
      return;
    }
    paused.current = false;
    const state = current.current;
    if (
      options.enabled &&
      state.committed === null &&
      state.pending === null &&
      state.failed === null
    ) {
      const initialRequest = latest.current.options.initialRequest;
      void execute({
        command: "initial",
        destination: 1,
        request: initialRequest,
        initialRequest,
        automaticRestart: false,
        notifyQuery: true,
      });
    }
  }, [execute, options.active, options.enabled, pause]);

  const nextPage = useCallback(() => {
    const state = current.current;
    const cursor = state.committed?.paging.next_cursor_token;
    if (
      state.pending ||
      state.cursorChainInvalid ||
      (state.error !== null && pageRecovery(state.error) !== "retry") ||
      !cursor ||
      !state.committed
    )
      return;
    void execute({
      command: "next",
      destination: state.committed.pageNumber + 1,
      request: latest.current.options.makeContinuation(cursor),
      initialRequest: state.committed.initialRequest,
      automaticRestart: false,
      notifyQuery: false,
    });
  }, [execute]);
  const previousPage = useCallback(() => {
    const state = current.current;
    const destination = (state.committed?.pageNumber ?? 1) - 1;
    const request = state.history[destination - 1];
    if (
      state.pending ||
      state.cursorChainInvalid ||
      (state.error !== null && pageRecovery(state.error) !== "retry") ||
      destination < 1 ||
      request === undefined ||
      !state.committed
    )
      return;
    void execute({
      command: "previous",
      destination,
      request,
      initialRequest: state.committed.initialRequest,
      automaticRestart: false,
      notifyQuery: false,
    });
  }, [execute]);
  const refresh = useCallback(() => {
    const state = current.current;
    if (
      state.pending ||
      state.cursorChainInvalid ||
      (state.error !== null && pageRecovery(state.error) !== "retry") ||
      !state.committed
    )
      return;
    void execute({
      command: "refresh",
      destination: state.committed.pageNumber,
      request: state.committed.request,
      initialRequest: state.committed.initialRequest,
      automaticRestart: false,
      notifyQuery: false,
    });
  }, [execute]);
  const restart = useCallback(() => {
    const state = current.current;
    if (
      state.error !== null &&
      !["retry", "restart"].includes(
        pageRecovery(state.error, state.cursorChainInvalid),
      )
    )
      return;
    const initialRequest =
      state.committed?.initialRequest ??
      state.failed?.initialRequest ??
      latest.current.options.initialRequest;
    void execute({
      command: "restart",
      destination: 1,
      request: initialRequest,
      initialRequest,
      automaticRestart: false,
      notifyQuery:
        state.committed === null && (state.failed?.notifyQuery ?? true),
    });
  }, [execute]);
  const retry = useCallback(() => {
    const state = current.current;
    if (
      state.pending ||
      !state.failed ||
      pageRecovery(state.error, state.cursorChainInvalid) !== "retry"
    )
      return;
    void execute(state.failed);
  }, [execute]);
  const clear = useCallback(() => {
    cancel();
    publish(empty<Item, Request, Metadata>(latest.current.contextKey));
  }, [cancel, publish]);
  const state =
    stored.contextKey === contextKey &&
    options.enabled &&
    options.isCurrent?.() !== false
      ? stored
      : empty<Item, Request, Metadata>(contextKey);
  return {
    pause,
    committed: state.committed,
    pending: state.pending,
    failed: state.failed,
    canNext:
      !state.cursorChainInvalid &&
      (state.error === null || pageRecovery(state.error) === "retry") &&
      state.committed?.paging.next_cursor_token != null,
    canPrevious:
      !state.cursorChainInvalid &&
      (state.error === null || pageRecovery(state.error) === "retry") &&
      (state.committed?.pageNumber ?? 1) > 1,
    canRefresh:
      !state.cursorChainInvalid &&
      state.committed !== null &&
      (state.error === null || pageRecovery(state.error) === "retry"),
    canRestart:
      state.error === null ||
      ["retry", "restart"].includes(
        pageRecovery(state.error, state.cursorChainInvalid),
      ),
    clear,
    error: state.error,
    items: state.committed?.items ?? [],
    loadState: state.loadState,
    loadGenerationKey: generation.current,
    nextPage,
    notice: state.notice,
    pageNumber: state.committed?.pageNumber ?? null,
    paging: state.committed?.paging ?? null,
    previousPage,
    refresh,
    retry,
    restart,
    restartAfterResourceRefresh: () => {
      const failed = state.failed;
      if (
        failed === null ||
        current.current.failed !== failed ||
        pageRecovery(current.current.error) !== "refresh_resource"
      )
        return;
      void execute({
        ...failed,
        command: "restart",
        destination: 1,
        request: failed.initialRequest,
        automaticRestart: false,
      });
    },
    recovery: pageRecovery(state.error, state.cursorChainInvalid),
  };
}

function pageRecovery(
  error: NetworkFlowRequestError | null,
  invalidCursor = false,
): NetworkFlowPageRecovery {
  if (error === null || isNetworkFlowProtectedStateLoss(error)) return "none";
  if (
    error.code === "network_flow_cursor_invalid" &&
    error.reasonCode === "scope_stale"
  )
    return "refresh_resource";
  if (error.retryAction === "do_not_retry") return "none";
  if (
    invalidCursor &&
    (error.retryAction === "restart_query" ||
      error.retryAction === "retry_with_backoff")
  )
    return "restart";
  switch (error.retryAction) {
    case "restart_query":
      return "restart";
    case "retry_with_backoff":
      return "retry";
    case "refresh_resource":
      return "refresh_resource";
    case "correct_request":
      return "correct_request";
    case "reduce_scope_or_limits":
      return "reduce_scope";
  }
}

export type NetworkFlowPageNavigation = Pick<
  ReturnType<typeof useNetworkFlowPagedQuery>,
  | "pending"
  | "failed"
  | "error"
  | "notice"
  | "pageNumber"
  | "recovery"
  | "canNext"
  | "canPrevious"
  | "canRefresh"
  | "canRestart"
  | "nextPage"
  | "previousPage"
  | "refresh"
  | "retry"
  | "restart"
>;
