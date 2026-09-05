import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { HTTPOperationResult } from "../services/browserApi";
import { incidentResource } from "../testing/appShellTestSupport";
import { deferred, jsonResponse } from "../testing/fetchMockTestSupport";
import type {
  IncidentDirectoryPaging,
  IncidentDirectoryResource,
  ListVisibleIncidentsResponse,
} from "./api/publicHttpTypes";
import {
  directoryCanLoadMore,
  IncidentDirectoryController,
  type IncidentDirectoryPorts,
} from "./incidentDirectoryModel";
import { useIncidentDirectory } from "./useIncidentDirectory";

const alpha = incidentResource(
  "00000000-0000-4000-8000-000000004101",
  "IR-A",
  "Alpha",
);
const beta = incidentResource(
  "00000000-0000-4000-8000-000000004102",
  "IR-B",
  "Beta",
);
const terminal: IncidentDirectoryPaging = {
  limit: 100,
  has_more: false,
  next_cursor: null,
};
const continuing: IncidentDirectoryPaging = {
  limit: 100,
  has_more: true,
  next_cursor: "page-two",
};
const success = (
  incidents: IncidentDirectoryResource[] = [alpha],
  paging: IncidentDirectoryPaging = terminal,
): HTTPOperationResult<ListVisibleIncidentsResponse> => ({
  ok: true,
  status: 200,
  payload: {
    data: { incidents },
    meta: { request_id: "request-directory", paging },
  },
});
const failure = (
  status: number,
  code: string,
): HTTPOperationResult<ListVisibleIncidentsResponse> => ({
  ok: false,
  status,
  payload: { error: { code, status } },
});
const controllers: IncidentDirectoryController[] = [];
function setup() {
  vi.useFakeTimers();
  const pending: ReturnType<
    typeof deferred<HTTPOperationResult<ListVisibleIncidentsResponse>>
  >[] = [];
  const ports = {
    list: vi.fn<IncidentDirectoryPorts["list"]>().mockImplementation(() => {
      const next =
        deferred<HTTPOperationResult<ListVisibleIncidentsResponse>>();
      pending.push(next);
      return next.promise;
    }),
    isCurrentSession: vi.fn().mockReturnValue(true),
    sessionLost: vi.fn(),
  };
  const controller = new IncidentDirectoryController(ports);
  controllers.push(controller);
  controller.setSession("account:session-one");
  controller.setActive(true);
  return { controller, ports, pending };
}
async function settle() {
  for (let step = 0; step < 5; step += 1) await Promise.resolve();
}
afterEach(() => {
  for (const controller of controllers.splice(0)) controller.dispose();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("incident directory operation", () => {
  it("preserves typing during initial load and rejects its obsolete response", async () => {
    const { controller, ports, pending } = setup();
    controller.changeSearch("Beta");
    expect(ports.list.mock.calls[0]?.[0].signal.aborted).toBe(true);
    pending[0]?.resolve(success());
    await settle();
    expect(controller.getSnapshot()).toMatchObject({
      query: { search: "Beta" },
      acceptedQuery: null,
      incidents: [],
      phase: "debouncing",
    });
    await vi.advanceTimersByTimeAsync(180);
    expect(ports.list).toHaveBeenCalledTimes(2);
    expect(ports.list.mock.calls[1]?.[0]).toMatchObject({
      query: { search: "Beta", statusFilter: "all" },
      limit: 100,
      cursorToken: null,
    });
    pending[1]?.resolve(success([beta]));
    await settle();
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [beta],
      phase: "ready",
    });
  });
  it("coalesces rapid edits and Enter without a second debounce request", async () => {
    const { controller, ports } = setup();
    controller.changeSearch("A");
    await vi.advanceTimersByTimeAsync(100);
    controller.changeSearch("Beta");
    controller.submit();
    controller.submit();
    await vi.advanceTimersByTimeAsync(180);
    expect(ports.list).toHaveBeenCalledTimes(2);
    expect(ports.list.mock.calls[1]?.[0].query.search).toBe("Beta");
  });
  it("submits status immediately with the current search and cancels debounce", async () => {
    const { controller, ports } = setup();
    controller.changeSearch("Beta");
    controller.changeStatus("closed");
    await vi.advanceTimersByTimeAsync(180);
    expect(ports.list).toHaveBeenCalledTimes(2);
    expect(ports.list.mock.calls[1]?.[0].query).toEqual({
      search: "Beta",
      statusFilter: "closed",
    });
  });
  it("retains previous accepted rows during replacement and ordinary failure", async () => {
    const { controller, pending } = setup();
    pending[0]?.resolve(success());
    await settle();
    controller.changeSearch("Beta");
    controller.submit();
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [alpha],
      acceptedQuery: { search: "" },
      phase: "refreshing",
    });
    pending[1]?.resolve(failure(500, "internal_error"));
    await settle();
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [alpha],
      phase: "failed",
      failure: { scope: "replace" },
    });
    expect(directoryCanLoadMore(controller.getSnapshot())).toBe(false);
    controller.retry();
    pending[2]?.resolve(success([]));
    await settle();
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [],
      acceptedQuery: { search: "Beta" },
      phase: "ready",
      failure: null,
    });
  });
  it("locks same-tick pagination and uses accepted query cursor and limit", async () => {
    const { controller, ports, pending } = setup();
    pending[0]?.resolve(success([alpha], continuing));
    await settle();
    controller.loadMore();
    controller.loadMore();
    expect(ports.list).toHaveBeenCalledTimes(2);
    expect(ports.list.mock.calls[1]?.[0]).toMatchObject({
      query: { search: "", statusFilter: "all" },
      cursorToken: "page-two",
      limit: 100,
    });
    controller.changeSearch("Beta");
    controller.loadMore();
    expect(ports.list).toHaveBeenCalledTimes(2);
    expect(ports.list.mock.calls[1]?.[0].signal.aborted).toBe(true);
    pending[1]?.resolve(success([beta]));
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([alpha]);
  });
  it("accumulates live overlapping pages by identity without sorting", async () => {
    const { controller, pending } = setup();
    pending[0]?.resolve(success([alpha], continuing));
    await settle();
    controller.loadMore();
    const updated = { ...alpha, incident_version: 2, title: "Updated" };
    pending[1]?.resolve(success([beta, updated]));
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([updated, beta]);
    expect(directoryCanLoadMore(controller.getSnapshot())).toBe(false);
  });
  it("uses continuation metadata for an empty nonterminal page", async () => {
    const { controller, pending } = setup();
    pending[0]?.resolve(success([], continuing));
    await settle();
    expect(directoryCanLoadMore(controller.getSnapshot())).toBe(true);
    controller.loadMore();
    pending[1]?.resolve(success([beta]));
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([beta]);
  });
  it("retries a failed page without changing its accepted cursor", async () => {
    const { controller, ports, pending } = setup();
    pending[0]?.resolve(success([alpha], continuing));
    await settle();
    controller.loadMore();
    pending[1]?.reject(new Error("offline"));
    await settle();
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [alpha],
      failure: { scope: "page", restart: false },
    });
    controller.retry();
    expect(ports.list.mock.calls[2]?.[0].cursorToken).toBe("page-two");
    pending[2]?.resolve(success([beta]));
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([alpha, beta]);
  });
  it("restarts page one only on explicit retry after an invalid cursor", async () => {
    const { controller, ports, pending } = setup();
    pending[0]?.resolve(success([alpha], continuing));
    await settle();
    controller.loadMore();
    pending[1]?.resolve(failure(400, "invalid_pagination_request"));
    await settle();
    expect(directoryCanLoadMore(controller.getSnapshot())).toBe(false);
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [alpha],
      paging: null,
      failure: { restart: true },
    });
    expect(ports.list).toHaveBeenCalledTimes(2);
    controller.retry();
    expect(ports.list.mock.calls[2]?.[0].cursorToken).toBeNull();
  });
  it("rejects cursor cycles and changed effective limits", async () => {
    const { controller, pending } = setup();
    pending[0]?.resolve(success([alpha], continuing));
    await settle();
    controller.loadMore();
    pending[1]?.resolve(success([beta], continuing));
    await settle();
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [alpha],
      phase: "failed",
      failure: { restart: true },
    });
    controller.retry();
    pending[2]?.resolve(success([beta], { ...terminal, limit: 50 }));
    await settle();
    expect(controller.getSnapshot().failure?.error?.code).toBe(
      "invalid_public_contract_response",
    );
  });
  it("supersedes refresh and rejects old successes and errors", async () => {
    const { controller, ports, pending } = setup();
    controller.refresh();
    controller.refresh();
    pending[2]?.resolve(success([beta]));
    await settle();
    pending[0]?.resolve(success());
    pending[1]?.resolve(failure(401, "session_required"));
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([beta]);
    expect(ports.sessionLost).not.toHaveBeenCalled();
  });
  it("clears materialization on departure and refreshes the retained query on return", async () => {
    const { controller, ports, pending } = setup();
    controller.changeSearch("Beta");
    controller.changeStatus("closed");
    pending[1]?.resolve(success([beta], continuing));
    await settle();
    controller.loadMore();
    controller.setActive(false);
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [],
      paging: null,
      query: { search: "Beta", statusFilter: "closed" },
      phase: "idle",
    });
    pending[2]?.resolve(success([alpha]));
    await settle();
    controller.setActive(true);
    expect(ports.list.mock.calls[3]?.[0]).toMatchObject({
      query: { search: "Beta", statusFilter: "closed" },
      cursorToken: null,
    });
    expect(controller.getSnapshot().incidents).toEqual([]);
  });
  it("clears account state before late responses from replaced authentication", async () => {
    const { controller, ports, pending } = setup();
    controller.changeSearch("private");
    controller.submit();
    controller.setSession("account:session-two");
    expect(controller.getSnapshot()).toMatchObject({
      query: { search: "", statusFilter: "all" },
      incidents: [],
      acceptedQuery: null,
    });
    controller.setActive(true);
    pending[0]?.resolve(success());
    pending[1]?.resolve(failure(401, "session_required"));
    await settle();
    expect(ports.sessionLost).not.toHaveBeenCalled();
    expect(controller.getSnapshot().incidents).toEqual([]);
    pending[2]?.resolve(success([beta]));
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([beta]);
  });
  it("clears protected rows on denial without terminating authentication", async () => {
    const { controller, ports, pending } = setup();
    pending[0]?.resolve(success([alpha], continuing));
    await settle();
    controller.loadMore();
    pending[1]?.resolve(failure(403, "authorization_denied"));
    await settle();
    expect(controller.getSnapshot()).toMatchObject({
      incidents: [],
      paging: null,
      acceptedQuery: null,
      phase: "forbidden",
    });
    expect(ports.sessionLost).not.toHaveBeenCalled();
    controller.retry();
    expect(ports.list).toHaveBeenCalledTimes(3);
  });
  it("clears all directory state before notifying current session loss", async () => {
    const { controller, ports, pending } = setup();
    controller.changeSearch("private");
    controller.submit();
    ports.sessionLost.mockImplementation(() =>
      expect(controller.getSnapshot()).toEqual({
        query: { search: "", statusFilter: "all" },
        acceptedQuery: null,
        incidents: [],
        paging: null,
        phase: "idle",
        failure: null,
      }),
    );
    pending[1]?.resolve(failure(401, "session_required"));
    await settle();
    expect(ports.sessionLost).toHaveBeenCalledOnce();
    pending[0]?.resolve(success());
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([]);
  });
  it("times out observation without automatic retry or late acceptance", async () => {
    const { controller, ports, pending } = setup();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(ports.list.mock.calls[0]?.[0].signal.aborted).toBe(true);
    expect(controller.getSnapshot()).toMatchObject({
      phase: "failed",
      failure: {
        scope: "replace",
        message: expect.stringContaining("timed out"),
      },
    });
    pending[0]?.resolve(success());
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([]);
    expect(ports.list).toHaveBeenCalledOnce();
    controller.retry();
    expect(ports.list).toHaveBeenCalledTimes(2);
  });
  it("invalidates disposal and independently checks the current authentication port", async () => {
    const { controller, ports, pending } = setup();
    ports.isCurrentSession.mockReturnValue(false);
    pending[0]?.resolve(success());
    await settle();
    expect(controller.getSnapshot().incidents).toEqual([]);
    controller.dispose();
    controller.changeSearch("ignored");
    controller.refresh();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(controller.getSnapshot()).toMatchObject({
      phase: "idle",
      query: { search: "" },
    });
    expect(ports.list).toHaveBeenCalledOnce();
  });
});

describe("incident directory React binding", () => {
  it("survives strict effect replay and cancels the retired observation", async () => {
    const signals: AbortSignal[] = [];
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((_input, init) => {
        signals.push(init.signal);
        return Promise.resolve(jsonResponse(success().payload));
      }),
    );
    const { result, unmount } = renderHook(
      () =>
        useIncidentDirectory({
          active: true,
          sessionIdentity: "account:session",
          sessionLost: vi.fn(),
        }),
      { reactStrictMode: true },
    );
    await act(settle);
    expect(signals[0]?.aborted).toBe(true);
    expect(result.current.state.incidents).toEqual([alpha]);
    unmount();
    expect(result.current.controller.getSnapshot().incidents).toEqual([]);
  });
  it("retains query drafts on inactivity but clears them on authentication replacement", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockImplementation(() =>
          Promise.resolve(jsonResponse(success().payload)),
        ),
    );
    const { result, rerender, unmount } = renderHook(
      (props: { active: boolean; sessionIdentity: string | null }) =>
        useIncidentDirectory({ ...props, sessionLost: vi.fn() }),
      {
        initialProps: {
          active: true,
          sessionIdentity: "account:one" as string | null,
        },
      },
    );
    await act(settle);
    act(() => result.current.controller.changeSearch("Beta"));
    rerender({ active: false, sessionIdentity: "account:one" });
    expect(result.current.state).toMatchObject({
      query: { search: "Beta" },
      incidents: [],
    });
    rerender({ active: true, sessionIdentity: "account:two" });
    expect(result.current.state.query.search).toBe("");
    await act(settle);
    rerender({ active: false, sessionIdentity: null });
    expect(result.current.state.incidents).toEqual([]);
    unmount();
  });
});
