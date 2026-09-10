import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import { useNetworkFlowPagedQuery } from "./useNetworkFlowPagedQuery";

type Request =
  | { readonly schema_id: "initial"; readonly query: string }
  | { readonly schema_id: "continuation"; readonly cursor_token: string };

describe("useNetworkFlowPagedQuery", () => {
  afterEach(() => vi.useRealTimers());

  it("retains the committed snapshot on failed automatic restart and never resubmits invalid cursors", async () => {
    const fetchPage = vi.fn(async () => page(["one"], "two"));
    const { result } = renderQueryHook(fetchPage);
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    fetchPage.mockResolvedValueOnce(page(["two"], "three"));
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.pageNumber).toBe(2));
    fetchPage
      .mockRejectedValueOnce(cursorError("expired"))
      .mockRejectedValueOnce(new Error("offline"));
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.loadState).toBe("error"));
    expect(result.current.items).toEqual(["two"]);
    expect(result.current.pageNumber).toBe(2);
    expect(result.current.recovery).toBe("restart");
    expect(result.current.canNext).toBe(false);
    expect(result.current.canPrevious).toBe(false);
    const calls = fetchPage.mock.calls.length;
    act(() => {
      result.current.nextPage();
      result.current.previousPage();
      result.current.refresh();
      result.current.retry();
    });
    expect(fetchPage).toHaveBeenCalledTimes(calls);
    act(() => result.current.restart());
    await waitFor(() => expect(result.current.pageNumber).toBe(1));
  });

  it("recovers expired backward navigation and stops after one automatic initial request", async () => {
    const fetchPage = vi.fn(async () => page(["one"], "two"));
    const { result } = renderQueryHook(fetchPage);
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    fetchPage.mockResolvedValueOnce(page(["two"], "three"));
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.pageNumber).toBe(2));
    fetchPage.mockResolvedValueOnce(page(["three"], null));
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.pageNumber).toBe(3));
    fetchPage.mockRejectedValue(cursorError("expired"));
    act(() => result.current.previousPage());
    await waitFor(() => expect(result.current.loadState).toBe("error"));
    expect(fetchPage).toHaveBeenCalledTimes(5);
    expect(result.current.pageNumber).toBe(3);
    expect(result.current.items).toEqual(["three"]);
    expect(result.current.failed?.automaticRestart).toBe(true);
  });

  it("gives protected-state loss precedence and never automatically restarts other cursor failures", async () => {
    for (const reason of [
      "authorization_lost",
      "actor_mismatch",
      "scope_stale",
      "malformed",
      "route_mismatch",
      "semantic_query_mismatch",
    ]) {
      const fetchPage = vi.fn(async () => page(["one"], "two"));
      const { result, unmount } = renderQueryHook(fetchPage);
      await waitFor(() => expect(result.current.loadState).toBe("ready"));
      fetchPage.mockRejectedValueOnce(cursorError(reason));
      act(() => result.current.nextPage());
      await waitFor(() => expect(result.current.loadState).toBe("error"));
      expect(fetchPage).toHaveBeenCalledTimes(2);
      if (
        ["authorization_lost", "actor_mismatch", "scope_stale"].includes(reason)
      ) {
        expect(result.current.items).toEqual([]);
        expect(result.current.pageNumber).toBeNull();
      } else expect(result.current.pageNumber).toBe(1);
      unmount();
    }
  });

  it("captures the initial request and fences success failure recovery and callbacks by current read authority", async () => {
    for (const outcome of ["success", "expired", "denied"]) {
      const pending = deferred<ReturnType<typeof page>>();
      const fetchPage = vi.fn(async () => page(["one"], "two"));
      const onQueryResult = vi.fn(),
        onError = vi.fn(),
        onIncidentAccessLost = vi.fn();
      let authority = "session-a";
      const hook = renderHook(
        ({ query }) => {
          const captured = authority;
          return useNetworkFlowPagedQuery<string, Request>({
            enabled: true,
            fetchPage,
            initialRequest: { schema_id: "initial", query },
            queryKey: "applied-query",
            readIdentity: captured,
            isCurrent: () => authority === captured,
            isContinuation: (r) => r.schema_id === "continuation",
            makeContinuation: (cursor_token) => ({
              schema_id: "continuation",
              cursor_token,
            }),
            onQueryResult,
            onError,
            onIncidentAccessLost,
            reconcile: (_p, next) => [...next],
          });
        },
        { initialProps: { query: "alpha" } },
      );
      await waitFor(() => expect(hook.result.current.loadState).toBe("ready"));
      fetchPage.mockImplementationOnce(() => pending.promise);
      act(() => hook.result.current.nextPage());
      hook.rerender({ query: "unapplied-new-request" });
      authority = "session-b";
      hook.rerender({ query: "unapplied-new-request" });
      await waitFor(() => expect(hook.result.current.loadState).toBe("ready"));
      const count = onError.mock.calls.length;
      await act(async () => {
        if (outcome === "success") pending.resolve(page(["late"], null));
        else
          pending.reject(
            cursorError(
              outcome === "expired" ? "expired" : "authorization_lost",
            ),
          );
      });
      expect(onError).toHaveBeenCalledTimes(count);
      expect(onQueryResult).toHaveBeenCalledTimes(2);
      expect(onIncidentAccessLost).not.toHaveBeenCalled();
      expect(fetchPage).toHaveBeenCalledTimes(3);
      expect(hook.result.current.pending).toBeNull();
      hook.unmount();
    }
  });

  it("uses the captured initial query when expiry races a newer mutable initial request", async () => {
    const pending = deferred<ReturnType<typeof page>>();
    const fetchPage = vi.fn(async (_request: Request) => page(["one"], "two"));
    const hook = renderHook(
      ({ query }) =>
        useNetworkFlowPagedQuery<string, Request>({
          enabled: true,
          fetchPage,
          initialRequest: { schema_id: "initial", query },
          queryKey: "stable",
          isContinuation: (r) => r.schema_id === "continuation",
          makeContinuation: (cursor_token) => ({
            schema_id: "continuation",
            cursor_token,
          }),
          onError: () => {},
          onIncidentAccessLost: undefined,
          reconcile: (_p, next) => [...next],
        }),
      { initialProps: { query: "alpha" } },
    );
    await waitFor(() => expect(hook.result.current.loadState).toBe("ready"));
    fetchPage.mockImplementationOnce(() => pending.promise);
    act(() => hook.result.current.nextPage());
    hook.rerender({ query: "bravo" });
    await act(async () => pending.reject(cursorError("expired")));
    await waitFor(() => expect(hook.result.current.loadState).toBe("ready"));
    expect(fetchPage.mock.calls.at(-1)?.[0]).toEqual({
      schema_id: "initial",
      query: "alpha",
    });
  });

  it("retries the failed destination independently of refreshing the displayed page", async () => {
    const fetchPage = vi.fn(async (request: Request) =>
      page([request.schema_id === "initial" ? "one" : "two"], "next"),
    );
    const { result } = renderQueryHook(fetchPage);
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    fetchPage.mockRejectedValueOnce(new Error("offline"));
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.loadState).toBe("error"));
    expect(result.current.failed?.destination).toBe(2);
    act(() => result.current.refresh());
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    expect(fetchPage.mock.calls.at(-1)?.[0]).toEqual({
      schema_id: "initial",
      query: "alpha",
    });
    expect(result.current.pageNumber).toBe(1);
    fetchPage.mockRejectedValueOnce(new Error("offline"));
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.loadState).toBe("error"));
    act(() => result.current.retry());
    await waitFor(() => expect(result.current.pageNumber).toBe(2));
    expect(fetchPage.mock.calls.at(-1)?.[0]).toEqual({
      schema_id: "continuation",
      cursor_token: "next",
    });
  });

  it("cancels pending navigation on explicit restart and coalesces duplicate restarts", async () => {
    const pending = deferred<ReturnType<typeof page>>();
    const fetchPage = vi.fn(async (_request: Request, _signal: AbortSignal) =>
      page(["one"], "two"),
    );
    const { result } = renderQueryHook(fetchPage);
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    fetchPage.mockImplementationOnce(() => pending.promise);
    act(() => result.current.nextPage());
    act(() => {
      result.current.restart();
      result.current.restart();
    });
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    expect(fetchPage).toHaveBeenCalledTimes(3);
    expect(fetchPage.mock.calls[1]?.[1].aborted).toBe(true);
    await act(async () => pending.resolve(page(["obsolete"], null)));
    expect(result.current.items).toEqual(["one"]);
    expect(result.current.pageNumber).toBe(1);
  });

  it("bounds a stopped read and distinguishes an accepted empty page from initial loading", async () => {
    const pending = deferred<ReturnType<typeof page>>();
    const fetchPage = vi.fn(async () => page([], null));
    const { result } = renderQueryHook(fetchPage);
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    expect(result.current.committed).not.toBeNull();
    expect(result.current.canNext).toBe(false);
    vi.useFakeTimers();
    fetchPage.mockImplementationOnce(() => pending.promise);
    act(() => result.current.refresh());
    expect(result.current.loadState).toBe("refreshing");
    await act(async () => vi.advanceTimersByTimeAsync(30_000));
    expect(result.current.loadState).toBe("error");
    expect(result.current.pending).toBeNull();
    expect(result.current.pageNumber).toBe(1);
    expect(result.current.recovery).toBe("retry");
    await act(async () => pending.resolve(page(["too late"], null)));
    expect(result.current.items).toEqual([]);
  });

  it("retains the committed page and history after failed forward and backward navigation", async () => {
    const fetchPage = vi.fn(async (request: Request) =>
      request.schema_id === "initial"
        ? page(["one"], "two")
        : page(["two"], "three"),
    );
    const { result } = renderQueryHook(fetchPage);
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    fetchPage.mockRejectedValueOnce(new Error("offline"));
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.loadState).toBe("error"));
    expect(result.current.items).toEqual(["one"]);
    expect(result.current.pageNumber).toBe(1);
    expect(result.current.canPrevious).toBe(false);
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.items).toEqual(["two"]));
    fetchPage.mockRejectedValueOnce(new Error("offline"));
    act(() => result.current.previousPage());
    await waitFor(() => expect(result.current.loadState).toBe("error"));
    expect(result.current.items).toEqual(["two"]);
    expect(result.current.pageNumber).toBe(2);
    expect(result.current.canPrevious).toBe(true);
  });

  it("admits only the first navigation command in the same turn", async () => {
    const pending = deferred<ReturnType<typeof page>>();
    const fetchPage = vi.fn(async () => page(["one"], "two"));
    const { result } = renderQueryHook(fetchPage);
    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    fetchPage.mockImplementation(() => pending.promise);
    act(() => {
      result.current.nextPage();
      result.current.nextPage();
      result.current.previousPage();
      result.current.refresh();
    });
    expect(fetchPage).toHaveBeenCalledTimes(2);
    expect(result.current.pageNumber).toBe(1);
    await act(async () => pending.resolve(page(["two"], null)));
    expect(result.current.pageNumber).toBe(2);
  });

  it("uses exact next continuations and replays the request that produced the previous page", async () => {
    const fetchPage = vi.fn(async (request: Request) =>
      request.schema_id === "initial"
        ? page(["page-one"], "cursor-2")
        : page(["page-two"], null),
    );
    const { result } = renderQueryHook(fetchPage);

    await waitFor(() => expect(result.current.loadState).toBe("ready"));
    act(() => result.current.nextPage());
    await waitFor(() => expect(result.current.items).toEqual(["page-two"]));
    act(() => result.current.previousPage());
    await waitFor(() => expect(result.current.items).toEqual(["page-one"]));

    expect(fetchPage.mock.calls.map(([request]) => request)).toEqual([
      { schema_id: "initial", query: "alpha" },
      { schema_id: "continuation", cursor_token: "cursor-2" },
      { schema_id: "initial", query: "alpha" },
    ]);
  });

  it("discards an invalid cursor and explains the automatic page-one restart", async () => {
    let initialCalls = 0;
    const fetchPage = vi.fn(async (request: Request) => {
      if (request.schema_id === "continuation") {
        throw new NetworkFlowRequestError({
          code: "network_flow_cursor_invalid",
          reasonCode: "expired",
          retryAction: "restart_query",
          retryable: false,
          safeMessage: "Invalid cursor.",
          status: 400,
        });
      }
      initialCalls += 1;
      return page(
        [initialCalls === 1 ? "first-version" : "restarted-version"],
        initialCalls === 1 ? "expired-cursor" : null,
      );
    });
    const { result } = renderQueryHook(fetchPage);

    await waitFor(() => expect(result.current.canNext).toBe(true));
    act(() => result.current.nextPage());
    await waitFor(() =>
      expect(result.current.items).toEqual(["restarted-version"]),
    );

    expect(result.current.pageNumber).toBe(1);
    expect(result.current.notice).toBe(
      "The page cursor expired. Results restarted at page one.",
    );
    expect(fetchPage.mock.calls.map(([request]) => request)).toEqual([
      { schema_id: "initial", query: "alpha" },
      { schema_id: "continuation", cursor_token: "expired-cursor" },
      { schema_id: "initial", query: "alpha" },
    ]);
  });

  it("aborts superseded requests and rejects late responses after a query change", async () => {
    for (const lateFailure of [false, true]) {
      const alpha = deferred<ReturnType<typeof page>>();
      const bravo = deferred<ReturnType<typeof page>>();
      const signals: AbortSignal[] = [];
      const onError = vi.fn();
      const onIncidentAccessLost = vi.fn();
      const fetchPage = vi.fn((request: Request, signal: AbortSignal) => {
        signals.push(signal);
        return request.schema_id === "initial" && request.query === "alpha"
          ? alpha.promise
          : bravo.promise;
      });
      const { result, rerender, unmount } = renderHook(
        ({ query }) =>
          useNetworkFlowPagedQuery<string, Request>({
            enabled: true,
            fetchPage,
            initialRequest: { schema_id: "initial", query },
            isContinuation: (request) => request.schema_id === "continuation",
            makeContinuation: (cursorToken) => ({
              schema_id: "continuation",
              cursor_token: cursorToken,
            }),
            onError,
            onIncidentAccessLost,
            queryKey: query,
            reconcile: (_previous, incoming) => [...incoming],
          }),
        { initialProps: { query: "alpha" } },
      );

      rerender({ query: "bravo" });
      expect(signals[0]?.aborted).toBe(true);
      await act(async () => bravo.resolve(page(["bravo"], null)));
      await waitFor(() => expect(result.current.items).toEqual(["bravo"]));
      const errorCount = onError.mock.calls.length;
      await act(async () => {
        if (lateFailure)
          alpha.reject(
            new NetworkFlowRequestError({
              code: "authorization_denied",
              safeMessage: "Access denied.",
              status: 403,
              retryAction: "do_not_retry",
              retryable: false,
            }),
          );
        else alpha.resolve(page(["late-alpha"], null));
      });
      expect(onError).toHaveBeenCalledTimes(errorCount);
      expect(onIncidentAccessLost).not.toHaveBeenCalled();
      expect(result.current.loadState).toBe("ready");
      expect(result.current.items).toEqual(["bravo"]);
      unmount();
    }
  });
});

function renderQueryHook(
  fetchPage: (
    request: Request,
    signal: AbortSignal,
  ) => Promise<ReturnType<typeof page>>,
) {
  return renderHook(() =>
    useNetworkFlowPagedQuery<string, Request>({
      enabled: true,
      fetchPage,
      initialRequest: { schema_id: "initial", query: "alpha" },
      isContinuation: (request) => request.schema_id === "continuation",
      makeContinuation: (cursorToken) => ({
        schema_id: "continuation",
        cursor_token: cursorToken,
      }),
      onError: vi.fn(),
      onIncidentAccessLost: undefined,
      queryKey: "alpha",
      reconcile: (_previous, incoming) => [...incoming],
    }),
  );
}

function page(items: readonly string[], nextCursorToken: string | null) {
  return {
    items,
    paging: {
      limit: 200,
      returned_count: items.length,
      next_cursor_token: nextCursorToken,
    },
  };
}

function deferred<T>() {
  let resolvePromise: (value: T) => void = () => undefined;
  let rejectPromise: (reason: unknown) => void = () => undefined;
  const promise = new Promise<T>((resolve, reject) => {
    resolvePromise = resolve;
    rejectPromise = reject;
  });
  return { promise, resolve: resolvePromise, reject: rejectPromise };
}

function cursorError(reasonCode: string) {
  return new NetworkFlowRequestError({
    code: "network_flow_cursor_invalid",
    reasonCode,
    retryAction: "restart_query",
    retryable: false,
    safeMessage: "Invalid cursor.",
    status: 400,
  });
}
