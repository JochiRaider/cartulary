import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  NetworkFlowQueryPagination,
  pageFailureFeedback,
} from "./NetworkFlowQueryPagination";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import type { NetworkFlowPageNavigation } from "./useNetworkFlowPagedQuery";

afterEach(cleanup);
describe("Network Flow page feedback", () => {
  it("names the committed page separately from pending destinations and successful recovery", () => {
    const page = navigation();
    const { rerender } = render(
      <NetworkFlowQueryPagination page={page} onRefreshResource={vi.fn()} />,
    );
    expect(screen.getByText("No page displayed")).toBeTruthy();
    const pending = {
      command: "next" as const,
      destination: 3,
      request: {},
      initialRequest: {},
      contextKey: "query",
      generation: 2,
      notifyQuery: false,
      automaticRestart: false,
    };
    rerender(
      <NetworkFlowQueryPagination
        page={{ ...page, pageNumber: 2, pending }}
        onRefreshResource={vi.fn()}
      />,
    );
    expect(
      screen.getByText("Page 2 · Loading page 3").getAttribute("aria-live"),
    ).toBe("polite");
    expect(
      screen
        .getByRole("button", { name: "Next" })
        .getAttribute("aria-disabled"),
    ).toBe("true");
    rerender(
      <NetworkFlowQueryPagination
        page={{
          ...page,
          pageNumber: 2,
          pending: { ...pending, command: "restart", destination: 1 },
          notice: "The page cursor expired. Restarting at page one.",
        }}
        onRefreshResource={vi.fn()}
      />,
    );
    expect(
      screen.getByText(
        "Page 2 · The page cursor expired. Restarting at page one.",
      ),
    ).toBeTruthy();
    rerender(
      <NetworkFlowQueryPagination
        page={{
          ...page,
          pageNumber: 1,
          notice: "The page cursor expired. Results restarted at page one.",
        }}
        onRefreshResource={vi.fn()}
      />,
    );
    expect(
      screen.getByText(
        "Page 1 · The page cursor expired. Results restarted at page one.",
      ),
    ).toBeTruthy();
  });

  it("separates failed destination retry from refresh and honors correction and resource recovery", () => {
    const base = navigation();
    const error = new NetworkFlowRequestError({
      code: "network_flow_transport_failed",
      status: 0,
      retryable: true,
      retryAction: "retry_with_backoff",
      safeMessage: "Connection failed.",
    });
    const page = {
      ...base,
      pageNumber: 2,
      canRefresh: true,
      error,
      recovery: "retry" as const,
      failed: {
        command: "previous" as const,
        destination: 1,
        request: {},
        initialRequest: {},
        contextKey: "query",
        generation: 2,
        notifyQuery: false,
        automaticRestart: false,
      },
    };
    const resource = vi.fn();
    const { rerender } = render(
      <NetworkFlowQueryPagination page={page} onRefreshResource={resource} />,
    );
    expect(pageFailureFeedback(page).message).toBe(
      "Could not load page 1. Showing page 2. Connection failed.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Retry page 1" }));
    expect(page.retry).toHaveBeenCalledOnce();
    expect(page.refresh).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Refresh page" }));
    expect(page.refresh).toHaveBeenCalledOnce();
    for (const recovery of [
      "correct_request",
      "reduce_scope",
      "none",
    ] as const) {
      rerender(
        <NetworkFlowQueryPagination
          page={{ ...page, canRefresh: false, canRestart: false, recovery }}
          onRefreshResource={resource}
        />,
      );
      expect(screen.queryByRole("button", { name: /Retry/ })).toBeNull();
      expect(
        screen
          .getByRole("button", { name: "Restart query" })
          .getAttribute("aria-disabled"),
      ).toBe("true");
    }
    rerender(
      <NetworkFlowQueryPagination
        page={{ ...page, recovery: "refresh_resource" }}
        onRefreshResource={resource}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Refresh tables" }));
    expect(resource).toHaveBeenCalledOnce();
  });
});

function navigation(): NetworkFlowPageNavigation {
  return {
    pending: null,
    failed: null,
    error: null,
    notice: null,
    pageNumber: null,
    recovery: "none",
    canNext: false,
    canPrevious: false,
    canRefresh: false,
    canRestart: true,
    nextPage: vi.fn(),
    previousPage: vi.fn(),
    refresh: vi.fn(),
    retry: vi.fn(),
    restart: vi.fn(),
  };
}
