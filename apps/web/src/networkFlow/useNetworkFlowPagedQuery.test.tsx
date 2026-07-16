import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import { useNetworkFlowPagedQuery } from "./useNetworkFlowPagedQuery";

type Request =
  | { readonly schema_id: "initial"; readonly query: string }
  | { readonly schema_id: "continuation"; readonly cursor_token: string };

describe("useNetworkFlowPagedQuery", () => {
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
    const alpha = deferred<ReturnType<typeof page>>();
    const bravo = deferred<ReturnType<typeof page>>();
    const signals: AbortSignal[] = [];
    const fetchPage = vi.fn((request: Request, signal: AbortSignal) => {
      signals.push(signal);
      return request.schema_id === "initial" && request.query === "alpha"
        ? alpha.promise
        : bravo.promise;
    });
    const { result, rerender } = renderHook(
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
          onError: vi.fn(),
          onIncidentAccessLost: undefined,
          queryKey: query,
          reconcile: (_previous, incoming) => [...incoming],
        }),
      { initialProps: { query: "alpha" } },
    );

    rerender({ query: "bravo" });
    expect(signals[0]?.aborted).toBe(true);
    await act(async () => bravo.resolve(page(["bravo"], null)));
    await waitFor(() => expect(result.current.items).toEqual(["bravo"]));
    await act(async () => alpha.resolve(page(["late-alpha"], null)));
    expect(result.current.items).toEqual(["bravo"]);
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
  const promise = new Promise<T>((resolve) => {
    resolvePromise = resolve;
  });
  return { promise, resolve: resolvePromise };
}
