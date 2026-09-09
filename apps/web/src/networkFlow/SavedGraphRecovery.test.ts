import { afterEach, describe, expect, it, vi } from "vitest";
import type { NetworkFlowSavedGraph } from "../services/networkFlowContractAdapter";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  SavedGraphController,
  type SavedGraphTransport,
} from "./SavedGraphController";
import type { SavedGraphAuthority } from "./savedGraphOperation";
import {
  deferredSavedGraph,
  savedGraphAccepted,
  savedGraphBindingFixture,
  savedGraphFixture,
  savedGraphResultFixture,
  savedGraphSelectorFixture,
  savedGraphTestAuthority,
} from "./savedGraphTestFixtures";

const owners: SavedGraphController[] = [];
afterEach(() => {
  for (const owner of owners.splice(0)) owner.dispose();
  vi.useRealTimers();
});
function failure(code: string, status: number) {
  return new NetworkFlowRequestError({
    code,
    status,
    safeMessage: code,
    retryAction: "refresh_resource",
    retryable: false,
  });
}
async function setup() {
  let authority: SavedGraphAuthority = {
    ...savedGraphTestAuthority,
    session: {},
  };
  const a = savedGraphFixture({
    selected_result_binding: savedGraphBindingFixture,
  });
  const b = savedGraphFixture({
    graph_view_id: `nfgv_${"b".repeat(32)}`,
    display_name: "Graph B",
  });
  const transport = {
    observationLimit: vi.fn(async () => 4),
    list: vi.fn<SavedGraphTransport["list"]>(async () => [a, b]),
    get: vi.fn<SavedGraphTransport["get"]>(async () => a),
    readJob: vi.fn<SavedGraphTransport["readJob"]>(async () => {
      throw new Error("Job unobserved");
    }),
    submit: vi.fn<SavedGraphTransport["submit"]>(async () =>
      savedGraphAccepted(a),
    ),
    navigation: {
      result: vi.fn<SavedGraphTransport["navigation"]["result"]>(async () =>
        savedGraphResultFixture(a),
      ),
      contributors: vi.fn<SavedGraphTransport["navigation"]["contributors"]>(
        async () => ({
          schema_id:
            "cartulary.network_flow.graph_view_contributor_query_result.v2",
          graph_view_id: a.graph_view_id,
          projection_result_id: savedGraphBindingFixture.projection_result_id,
          selector: savedGraphSelectorFixture,
          contributors: [],
          meta: {
            paging: {
              limit: 100,
              returned_count: 0,
              next_cursor_token: "next",
            },
          },
        }),
      ),
    },
  };
  let sequence = 0;
  const owner = new SavedGraphController({
    identify: () => `attempt-${++sequence}`,
  });
  owners.push(owner);
  owner.bind(transport, () => authority);
  owner.setActive(true);
  await owner.loadGraphs();
  await vi.waitFor(() =>
    expect(owner.getSnapshot().navigation.result).not.toBeNull(),
  );
  return {
    owner,
    transport,
    a,
    b,
    setAuthority(patch: Partial<SavedGraphAuthority>) {
      authority = { ...authority, ...patch };
      owner.revalidateAuthority();
    },
  };
}

describe("Saved graph recovery boundaries", () => {
  it("retains original uncertainty after a denied replay and resolves only the same attempt", async () => {
    const { owner, transport, a } = await setup();
    owner.openAction("create", a.semantic_query);
    owner.setDraft("Created graph");
    transport.submit.mockRejectedValueOnce(new Error("Lost acknowledgement"));
    await owner.submit();
    const attempt = owner.getSnapshot().operation?.attempt;
    transport.submit.mockRejectedValueOnce(
      failure("authorization_denied", 403),
    );
    await owner.replay();
    expect(owner.getSnapshot().operation?.phase).toBe("uncertain");
    owner.setDraft("Replacement");
    owner.openAction("refresh");
    expect(owner.getSnapshot().operation?.attempt).toBe(attempt);
    expect(await owner.submit()).toBe(false);
    owner.selectGraphView(null);
    const declaration = deferredSavedGraph<NetworkFlowSavedGraph>();
    transport.get.mockReturnValueOnce(declaration.promise);
    transport.submit.mockResolvedValueOnce(savedGraphAccepted(a));
    await owner.replay();
    await owner.reconcileGraph(a.graph_view_id);
    declaration.resolve(a);

    expect(transport.submit.mock.calls.at(-1)?.[0]).toBe(attempt);
    expect(owner.getSnapshot().operation?.phase).toBe("acknowledged");
  });

  it("does not replace a materialized declaration with its historical creation receipt", async () => {
    const { owner, transport, a } = await setup();
    owner.openAction("create", a.semantic_query);
    owner.setDraft(a.display_name);
    transport.submit.mockRejectedValueOnce(new Error("Lost acknowledgement"));
    await owner.submit();
    await owner.loadGraphs();
    const result = owner.getSnapshot().navigation.result;
    transport.list.mockRejectedValue(new Error("Follow-up unavailable"));
    transport.submit.mockResolvedValueOnce(
      savedGraphAccepted({ ...a, selected_result_binding: null }),
    );
    await owner.replay();
    expect(owner.selectedGraph?.selected_result_binding).toEqual(
      savedGraphBindingFixture,
    );
    expect(owner.getSnapshot().navigation.result).toBe(result);
    expect(owner.getSnapshot().operation?.phase).toBe("acknowledged");
  });

  it("preserves selection made after the original attempt across explicit replay", async () => {
    const { owner, transport, a, b } = await setup();
    owner.openAction("create", a.semantic_query);
    owner.setDraft(a.display_name);
    transport.submit.mockRejectedValueOnce(new Error("Lost acknowledgement"));
    await owner.submit();
    owner.selectGraphView(b.graph_view_id);
    await owner.replay();
    expect(owner.getSnapshot().selectedGraphViewId).toBe(b.graph_view_id);
  });

  it("fences conflict review after retirement without resurrecting a declaration", async () => {
    const { owner, transport, a } = await setup();
    owner.openAction("rename");
    owner.setDraft("Retained draft");
    transport.submit.mockRejectedValueOnce(
      failure("network_flow_graph_view_version_conflict", 409),
    );
    await owner.submit();
    const deferred = deferredSavedGraph<NetworkFlowSavedGraph>();
    transport.get.mockReturnValueOnce(deferred.promise);
    const reviewing = owner.reviewCurrent();
    owner.removeGraph(a.graph_view_id);
    deferred.resolve(a);
    await reviewing;
    expect(
      owner
        .getSnapshot()
        .graphs.some((graph) => graph.graph_view_id === a.graph_view_id),
    ).toBe(false);
    expect(owner.getSnapshot().operation?.draft).toBe("Retained draft");
    expect(owner.getSnapshot().operation?.phase).toBe("awaiting_review");
  });

  it("reconciles independent graphs without dropping either accepted read", async () => {
    const { owner, transport, a, b } = await setup();
    const first = deferredSavedGraph<NetworkFlowSavedGraph>();
    const second = deferredSavedGraph<NetworkFlowSavedGraph>();
    transport.get.mockImplementation((id) =>
      id === a.graph_view_id ? first.promise : second.promise,
    );
    const readingA = owner.reconcileGraph(a.graph_view_id);
    const readingB = owner.reconcileGraph(b.graph_view_id);
    second.resolve({ ...b, display_name: "Changed B", graph_view_version: 2 });
    await readingB;
    first.resolve({ ...a, display_name: "Changed A", graph_view_version: 2 });
    await readingA;
    expect(
      owner.getSnapshot().graphs.map((graph) => graph.display_name),
    ).toEqual(["Changed A", "Changed B"]);
  });

  it("reconciles invalidations received while a rejected mutation holds admission", async () => {
    const { owner, transport, a } = await setup();
    const pending = deferredSavedGraph<never>();
    transport.submit.mockReturnValueOnce(pending.promise);
    transport.get.mockResolvedValue({
      ...a,
      display_name: "Remote rename",
      graph_view_version: 2,
    });
    owner.openAction("rename");
    owner.setDraft("Local name");
    const writing = owner.submit();
    await owner.onResourceChange({
      resourceKind: "network_flow_graph_view",
      resourceId: a.graph_view_id,
      changeKind: "invalidate",
      reasonCode: "renamed",
    });
    pending.reject(failure("network_flow_graph_view_version_conflict", 409));
    await writing;
    await vi.waitFor(() =>
      expect(owner.selectedGraph?.display_name).toBe("Remote rename"),
    );
  });

  it("withdraws cached exposure on an authoritative list denial and permits explicit current-state recovery", async () => {
    const { owner, transport } = await setup();
    transport.list.mockRejectedValueOnce(failure("authorization_denied", 403));
    await owner.loadGraphs();
    expect(owner.getSnapshot().navigation.result).toBeNull();
    expect(owner.getSnapshot().graphs).toEqual([]);
    await owner.loadGraphs();
    await vi.waitFor(() =>
      expect(owner.getSnapshot().navigation.result).not.toBeNull(),
    );
  });

  it("withdraws result and contributors when a contributor request reports lost read authority", async () => {
    const { owner, transport } = await setup();
    await owner.navigation.selectObject(savedGraphSelectorFixture);
    transport.navigation.contributors.mockRejectedValueOnce(
      failure("authorization_denied", 403),
    );
    await owner.navigation.loadContributors(true);
    expect(owner.getSnapshot().navigation.result).toBeNull();
    expect(owner.getSnapshot().navigation.selection).toBeNull();
    expect(owner.getSnapshot().navigation.contributors).toEqual([]);
  });

  it("aborts a queued mutation when its captured authority changes before dispatch", async () => {
    const { owner, transport, setAuthority } = await setup();
    const queue = deferredSavedGraph<void>();
    const dispatched = vi.fn();
    transport.submit.mockImplementation(async (_attempt, signal) => {
      await queue.promise;
      signal.throwIfAborted();
      dispatched();
      return savedGraphAccepted();
    });
    owner.openAction("rename");
    owner.setDraft("Queued rename");
    const writing = owner.submit();
    setAuthority({ role: "admin" });
    queue.resolve();
    await writing;
    expect(dispatched).not.toHaveBeenCalled();
    const other = await setup();
    const waiting = deferredSavedGraph<void>();
    const sending = vi.fn();
    other.transport.submit.mockImplementation(
      async (_attempt, _signal, guard) => {
        await waiting.promise;
        guard();
        sending();
        return savedGraphAccepted();
      },
    );
    other.owner.openAction("rename");
    other.owner.setDraft("Retired while queued");
    const queued = other.owner.submit();
    other.owner.removeGraph(other.a.graph_view_id);
    waiting.resolve();
    await queued;
    expect(sending).not.toHaveBeenCalled();
    expect(other.owner.getSnapshot().operation?.phase).toBe("awaiting_review");
  });
  it("recovers a withdrawn result through its current declaration before loading authorized bytes", async () => {
    const { owner, transport, a } = await setup();
    const changed = {
      ...a,
      selected_result_binding: {
        ...savedGraphBindingFixture,
        canonical_output_sha256: "e".repeat(64),
      },
    };
    transport.navigation.result.mockResolvedValueOnce(
      savedGraphResultFixture(changed),
    );
    await owner.navigation.loadResult();
    expect(owner.getSnapshot().navigation.identity).toBeNull();
    expect(owner.getSnapshot().navigation.resultError?.code).toBe(
      "network_flow_invalid_response",
    );
    transport.get.mockResolvedValue(changed);
    transport.navigation.result.mockResolvedValue(
      savedGraphResultFixture(changed),
    );
    await owner.recoverResult();
    await vi.waitFor(() =>
      expect(owner.getSnapshot().navigation.result?.graph_view).toEqual(
        changed,
      ),
    );
    expect(transport.get).toHaveBeenCalledWith(
      a.graph_view_id,
      expect.any(AbortSignal),
    );
  });

  it("restarts an interrupted initial result read after valid authority reconciliation", async () => {
    const { owner, transport, a, setAuthority } = await setup();
    owner.selectGraphView(null);
    const late =
      deferredSavedGraph<ReturnType<typeof savedGraphResultFixture>>();
    transport.navigation.result.mockReturnValueOnce(late.promise);
    owner.selectGraphView(a.graph_view_id);
    expect(owner.getSnapshot().navigation.resultState).toBe("loading");
    setAuthority({ role: "admin" });
    await vi.waitFor(() =>
      expect(owner.getSnapshot().navigation.result).not.toBeNull(),
    );
    const current = owner.getSnapshot().navigation.result;
    late.resolve(savedGraphResultFixture({ ...a, display_name: "Historical" }));
    await Promise.resolve();
    expect(owner.getSnapshot().navigation.result).toBe(current);
  });

  it("keeps authorized contributor context during transport failure and withdraws only its invalidated binding", async () => {
    const { owner, transport, a, b } = await setup();
    await owner.navigation.selectObject(savedGraphSelectorFixture);
    const before = owner.getSnapshot().navigation;
    transport.navigation.contributors.mockRejectedValueOnce(
      new Error("Disconnected"),
    );
    await owner.navigation.loadContributors(true);
    const transient = owner.getSnapshot().navigation;
    expect(transient.result).toBe(before.result);
    expect(transient.selection).toBe(before.selection);
    expect(transient.nextContributorCursor).toBe(before.nextContributorCursor);
    expect(transient.contributorError?.code).toBe(
      "network_flow_transport_failed",
    );
    transport.navigation.contributors.mockRejectedValueOnce(
      failure("network_flow_graph_query_stale", 409),
    );
    await owner.navigation.loadContributors();
    expect(owner.getSnapshot().navigation.result).toBeNull();
    expect(
      owner.getSnapshot().graphs.map((graph) => graph.graph_view_id),
    ).toEqual([a.graph_view_id, b.graph_view_id]);
    await owner.recoverResult();
    await vi.waitFor(() =>
      expect(owner.getSnapshot().navigation.result).not.toBeNull(),
    );
  });

  it("preserves typed job read failures without treating missing jobs as lost incident access", async () => {
    const { owner, transport, a } = await setup();
    const result = owner.getSnapshot().navigation.result;
    transport.readJob.mockRejectedValueOnce(failure("job_not_found", 404));
    owner.resumeObservation();
    await vi.waitFor(() =>
      expect(
        owner.getSnapshot().observations[a.latest_job_id ?? ""]?.failure?.code,
      ).toBe("job_not_found"),
    );
    expect(owner.getSnapshot().navigation.result).toBe(result);
    transport.readJob.mockRejectedValueOnce(
      failure("authorization_denied", 403),
    );
    owner.resumeObservation();
    await vi.waitFor(() => expect(owner.getSnapshot().graphs).toEqual([]));
    expect(owner.getSnapshot().navigation.result).toBeNull();
  });

  it("accepts a late original receipt after replay denial without installing the historical binding", async () => {
    const { owner, transport, a } = await setup();
    vi.useFakeTimers();
    const pending = deferredSavedGraph<ReturnType<typeof savedGraphAccepted>>();
    transport.submit.mockReturnValueOnce(pending.promise);
    owner.openAction("create", a.semantic_query);
    owner.setDraft(a.display_name);
    const writing = owner.submit();
    await vi.advanceTimersByTimeAsync(30_000);
    await writing;
    transport.submit.mockRejectedValueOnce(failure("client_txn_conflict", 409));
    await owner.replay();
    expect(owner.getSnapshot().operation?.phase).toBe("uncertain");
    pending.resolve(
      savedGraphAccepted({ ...a, selected_result_binding: null }),
    );
    await vi.advanceTimersByTimeAsync(0);
    expect(owner.getSnapshot().operation?.phase).toBe("acknowledged");
    expect(owner.selectedGraph?.selected_result_binding).toEqual(
      savedGraphBindingFixture,
    );
  });
});
