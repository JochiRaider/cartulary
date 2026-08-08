import { useCallback, useEffect, useRef, useState } from "react";
import type { NetworkFlowPaging } from "../services/networkFlowContractAdapter";
import {
  isNetworkFlowAuthorizationLoss,
  isNetworkFlowCursorInvalid,
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

export function useNetworkFlowPagedQuery<Item, Request>(options: {
  readonly enabled: boolean;
  readonly fetchPage: (
    request: Request,
    signal: AbortSignal,
  ) => Promise<{
    readonly items: readonly Item[];
    readonly paging: NetworkFlowPaging;
  }>;
  readonly initialRequest: Request;
  readonly isContinuation: (request: Request) => boolean;
  readonly makeContinuation: (cursorToken: string) => Request;
  readonly onError: (error: NetworkFlowRequestError | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly queryKey: string;
  readonly reconcile: (
    previous: readonly Item[],
    incoming: readonly Item[],
  ) => Item[];
}) {
  const generation = useRef(0);
  const controller = useRef<AbortController | null>(null);
  const itemsRef = useRef<readonly Item[]>([]);
  const pagingRef = useRef<NetworkFlowPaging | null>(null);
  const historyRef = useRef<Request[]>([options.initialRequest]);
  const pageIndexRef = useRef(0);
  const queryKeyRef = useRef(options.queryKey);
  const initialRequestRef = useRef(options.initialRequest);
  const fetchPageRef = useRef(options.fetchPage);
  const isContinuationRef = useRef(options.isContinuation);
  const makeContinuationRef = useRef(options.makeContinuation);
  const onErrorRef = useRef(options.onError);
  const onIncidentAccessLostRef = useRef(options.onIncidentAccessLost);
  const reconcileRef = useRef(options.reconcile);
  initialRequestRef.current = options.initialRequest;
  fetchPageRef.current = options.fetchPage;
  isContinuationRef.current = options.isContinuation;
  makeContinuationRef.current = options.makeContinuation;
  onErrorRef.current = options.onError;
  onIncidentAccessLostRef.current = options.onIncidentAccessLost;
  reconcileRef.current = options.reconcile;

  const [items, setItems] = useState<readonly Item[]>([]);
  const [paging, setPaging] = useState<NetworkFlowPaging | null>(null);
  const [pageIndex, setPageIndex] = useState(0);
  const [loadState, setLoadState] = useState<NetworkFlowQueryLoadState>("idle");
  const [loadGenerationKey, setLoadGenerationKey] = useState(0);
  const [error, setError] = useState<NetworkFlowRequestError | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const execute = useCallback(async (requested: Request) => {
    controller.current?.abort();
    const activeController = new AbortController();
    controller.current = activeController;
    generation.current += 1;
    const activeGeneration = generation.current;
    setLoadGenerationKey(activeGeneration);
    setLoadState(itemsRef.current.length === 0 ? "loading" : "refreshing");
    setError(null);
    onErrorRef.current(null);

    let request = requested;
    let recoveredCursor = false;
    for (;;) {
      try {
        const result = await fetchPageRef.current(
          request,
          activeController.signal,
        );
        if (
          activeController.signal.aborted ||
          generation.current !== activeGeneration
        ) {
          return;
        }
        const nextItems = reconcileRef.current(itemsRef.current, result.items);
        itemsRef.current = nextItems;
        pagingRef.current = result.paging;
        setItems(nextItems);
        setPaging(result.paging);
        setLoadState("ready");
        setError(null);
        onErrorRef.current(null);
        return;
      } catch (caught) {
        if (activeController.signal.aborted) {
          return;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow query failed.",
        );
        if (
          !recoveredCursor &&
          isContinuationRef.current(request) &&
          !isNetworkFlowProtectedStateLoss(requestError) &&
          isNetworkFlowCursorInvalid(requestError)
        ) {
          recoveredCursor = true;
          request = initialRequestRef.current;
          historyRef.current = [request];
          pageIndexRef.current = 0;
          itemsRef.current = [];
          pagingRef.current = null;
          setItems([]);
          setPaging(null);
          setPageIndex(0);
          setNotice(cursorResetNotice(requestError));
          setLoadState("loading");
          continue;
        }
        if (isNetworkFlowProtectedStateLoss(requestError)) {
          historyRef.current = [initialRequestRef.current];
          pageIndexRef.current = 0;
          itemsRef.current = [];
          pagingRef.current = null;
          setItems([]);
          setPaging(null);
          setPageIndex(0);
          setNotice(null);
          if (isNetworkFlowAuthorizationLoss(requestError)) {
            onIncidentAccessLostRef.current?.();
          }
        }
        setError(requestError);
        setLoadState("error");
        onErrorRef.current(requestError);
        return;
      }
    }
  }, []);

  useEffect(() => {
    const queryChanged = queryKeyRef.current !== options.queryKey;
    queryKeyRef.current = options.queryKey;
    controller.current?.abort();
    generation.current += 1;
    const initialRequest = initialRequestRef.current;
    historyRef.current = [initialRequest];
    pageIndexRef.current = 0;
    itemsRef.current = [];
    pagingRef.current = null;
    setItems([]);
    setPaging(null);
    setPageIndex(0);
    setError(null);
    if (queryChanged || !options.enabled) {
      setNotice(null);
    }
    onErrorRef.current(null);
    if (!options.enabled) {
      setLoadState("idle");
      return;
    }
    void execute(initialRequest);
    return () => controller.current?.abort();
  }, [execute, options.enabled, options.queryKey]);

  const nextPage = useCallback(() => {
    const cursorToken = pagingRef.current?.next_cursor_token ?? null;
    if (
      cursorToken === null ||
      loadState === "loading" ||
      loadState === "refreshing"
    ) {
      return;
    }
    const request = makeContinuationRef.current(cursorToken);
    const nextIndex = pageIndexRef.current + 1;
    historyRef.current = [...historyRef.current.slice(0, nextIndex), request];
    pageIndexRef.current = nextIndex;
    setPageIndex(nextIndex);
    setNotice(null);
    void execute(request);
  }, [execute, loadState]);

  const previousPage = useCallback(() => {
    if (pageIndexRef.current === 0) {
      return;
    }
    const previousIndex = pageIndexRef.current - 1;
    const request = historyRef.current[previousIndex];
    if (request === undefined) {
      return;
    }
    pageIndexRef.current = previousIndex;
    setPageIndex(previousIndex);
    setNotice(null);
    void execute(request);
  }, [execute]);

  const refresh = useCallback(() => {
    const request = historyRef.current[pageIndexRef.current];
    if (request !== undefined) {
      setNotice(null);
      void execute(request);
    }
  }, [execute]);

  const clear = useCallback(() => {
    controller.current?.abort();
    generation.current += 1;
    const initialRequest = initialRequestRef.current;
    historyRef.current = [initialRequest];
    pageIndexRef.current = 0;
    itemsRef.current = [];
    pagingRef.current = null;
    setItems([]);
    setPaging(null);
    setPageIndex(0);
    setLoadState("idle");
    setError(null);
    setNotice(null);
  }, []);

  return {
    canNext: paging?.next_cursor_token !== null && paging !== null,
    canPrevious: pageIndex > 0,
    clear,
    error,
    items,
    loadState,
    loadGenerationKey,
    nextPage,
    notice,
    pageNumber: pageIndex + 1,
    paging,
    previousPage,
    refresh,
  };
}

function cursorResetNotice(error: NetworkFlowRequestError): string {
  switch (error.reasonCode) {
    case "expired":
      return "The page cursor expired. Results restarted at page one.";
    case "scope_stale":
    case "semantic_query_mismatch":
      return "The query changed on the server. Results restarted at page one.";
    default:
      return "The page cursor is no longer valid. Results restarted at page one.";
  }
}
