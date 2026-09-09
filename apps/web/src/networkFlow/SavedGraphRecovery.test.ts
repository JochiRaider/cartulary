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
  const receipt = vi.fn<() => ReturnType<SavedGraphTransport["submit"]>>(
    async () => savedGraphAccepted(a),
  );
  const transport = {
    observationLimit: vi.fn(async () => 4),
    list: vi.fn<SavedGraphTransport["list"]>(async () => [a, b]),
    get: vi.fn<SavedGraphTransport["get"]>(async () => a),
    readJob: vi.fn<SavedGraphTransport["readJob"]>(async () => {
      throw new Error("Job unobserved");
    }),
    receipt,
    submit: vi.fn<SavedGraphTransport["submit"]>(
      async (_attempt, _signal, dispatch) => {
        dispatch();
        return receipt();
      },
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
  it("reconciles retirement removal after it interrupts the acknowledged list read", async () => {
    const { owner, transport, a, setAuthority } = await setup();
    setAuthority({ role: "admin" });
    transport.list.mockResolvedValue([a]);
    await owner.loadGraphs();
    const interrupted = deferredSavedGraph<NetworkFlowSavedGraph[]>();
    transport.list
      .mockReturnValueOnce(interrupted.promise)
      .mockResolvedValue([]);
    transport.receipt.mockResolvedValue({ kind: "retired" });
    owner.openAction("retire");
    expect(await owner.submit()).toBe(true);
    await owner.onResourceChange({
      resourceKind: "network_flow_graph_view",
      resourceId: a.graph_view_id,
      changeKind: "remove",
      reasonCode: "retired",
    });
    interrupted.resolve([a]);
    await vi.waitFor(() => expect(owner.getSnapshot().listState).toBe("ready"));
    expect(owner.getSnapshot().graphs).toEqual([]);
    expect(owner.getSnapshot().navigation.result).toBeNull();
  });

  it("keeps a queued unsent abort definite and requires review instead of replay", async () => {
    const { owner, transport, setAuthority } = await setup();
    const queue = deferredSavedGraph<void>();
    transport.submit.mockImplementation(async (_attempt, signal, dispatch) => {
      await queue.promise;
      signal.throwIfAborted();
      dispatch();
      return savedGraphAccepted();
    });
    owner.openAction("rename");
    owner.setDraft("Queued draft");
    const writing = owner.submit();
    setAuthority({ role: "viewer" });
    queue.resolve();
    await writing;
    expect(owner.getSnapshot().operation?.phase).toBe("awaiting_review");
    expect(owner.getSnapshot().operation?.failure?.certainty).toBe("rejected");
    expect(owner.getSnapshot().operation?.failure?.category).toBe(
      "authorization",
    );
    expect(owner.getSnapshot().operation?.draft).toBe("Queued draft");
  });

  it("resumes the captured create query after definite failure and dialog closure", async () => {
    const { owner, transport, a } = await setup();
    owner.openAction("create", a.semantic_query);
    owner.setDraft("Retained create");
    transport.receipt.mockRejectedValueOnce(
      failure("network_flow_graph_view_limit_exceeded", 409),
    );
    await owner.submit();
    owner.closeDialog();
    owner.openAction("create", {
      ...a.semantic_query,
      selected_table_ids: [`nft_${"d".repeat(32)}`],
    });
    expect(owner.getSnapshot().operation?.intent.query).toEqual(
      a.semantic_query,
    );
    expect(owner.getSnapshot().operation?.draft).toBe("Retained create");
  });

  it("replaces a rejected captured query only through explicit current-query preparation", async () => {
    const { owner, transport, a } = await setup();
    owner.openAction("create", a.semantic_query);
    owner.setDraft("Same name");
    transport.receipt.mockRejectedValueOnce(
      failure("network_flow_graph_view_limit_exceeded", 409),
    );
    await owner.submit();
    const original = owner.getSnapshot().operation?.attempt;
    const query: typeof a.semantic_query = {
      ...a.semantic_query,
      selected_table_ids: [`nft_${"d".repeat(32)}`],
    };
    owner.prepareCurrentQuery(query);
    expect(owner.getSnapshot().operation?.draft).toBe("Same name");
    expect(owner.getSnapshot().operation?.intent.query).toEqual(query);
    transport.receipt.mockRejectedValueOnce(new Error("Lost acknowledgement"));
    await owner.submit();
    const uncertain = owner.getSnapshot().operation?.attempt;
    expect(uncertain?.transactionId).not.toBe(original?.transactionId);
    owner.prepareCurrentQuery(a.semantic_query);
    expect(owner.getSnapshot().operation?.attempt).toBe(uncertain);
  });

  it("hides paused session drafts and fences old acknowledgements until current authorization returns", async () => {
    const { owner, transport, a, setAuthority } = await setup();
    owner.openAction("create", a.semantic_query);
    owner.setDraft("Paused draft");
    const pending = deferredSavedGraph<ReturnType<typeof savedGraphAccepted>>();
    transport.receipt.mockReturnValueOnce(pending.promise);
    const writing = owner.submit();
    const attempt = owner.getSnapshot().operation?.attempt;
    owner.pauseSession();
    expect(owner.getSnapshot().operation).toBeNull();
    expect(owner.getSnapshot().navigation.result).toBeNull();
    expect(owner.getSnapshot().graphs).toEqual([]);
    pending.resolve(savedGraphAccepted(a));
    await writing;
    expect(owner.getSnapshot().operation).toBeNull();
    setAuthority({ session: {}, sessionResolved: true });
    expect(owner.getSnapshot().operation?.phase).toBe("uncertain");
    expect(owner.getSnapshot().operation?.attempt).toBe(attempt);
    expect(transport.submit).toHaveBeenCalledTimes(1);
    await owner.replay();
    expect(transport.submit.mock.calls.at(-1)?.[0]).toBe(attempt);
    expect(owner.getSnapshot().operation?.phase).toBe("acknowledged");
  });

  it("retains a copyable write-loss draft but purges it when the profile is withdrawn", async () => {
    const { owner, setAuthority } = await setup();
    owner.openAction("rename");
    owner.setDraft("Copyable draft");
    setAuthority({ role: "viewer" });
    expect(owner.getSnapshot().operation?.draft).toBe("Copyable draft");
    expect(owner.getSnapshot().operation?.phase).toBe("awaiting_review");
    expect(await owner.submit()).toBe(false);
    setAuthority({ available: false, profileAvailable: true, open: false });
    expect(owner.getSnapshot().operation?.draft).toBe("Copyable draft");
    expect(owner.getSnapshot().navigation.result).toBeNull();
    setAuthority({ available: false, profileAvailable: false });
    expect(owner.getSnapshot().operation).toBeNull();
    expect(owner.getSnapshot().graphs).toEqual([]);
  });

  it("withdraws saved exposure for the canonical unclaimed profile response", async () => {
    const { owner, transport } = await setup();
    transport.list.mockRejectedValueOnce(
      failure("extension_profile_not_claimed", 404),
    );
    await owner.loadGraphs();
    expect(owner.getSnapshot().graphs).toEqual([]);
    expect(owner.getSnapshot().navigation.result).toBeNull();
  });

  it("publishes selection and immutable result exposure atomically to subscribers", async () => {
    const { owner, b } = await setup();
    const inconsistent: string[] = [];
    const unsubscribe = owner.subscribe(() => {
      const snapshot = owner.getSnapshot();
      const resultGraph =
        snapshot.navigation.result?.result.graph_projection_result
          .graph_view_id;
      if (resultGraph && resultGraph !== snapshot.selectedGraphViewId)
        inconsistent.push(resultGraph);
    });
    owner.selectGraphView(b.graph_view_id);
    unsubscribe();
    expect(inconsistent).toEqual([]);
  });

  it("recovers an accepted job and declaration by their operation target after selection and read failure", async () => {
    const { owner, transport, a, b } = await setup();
    owner.openAction("refresh");
    transport.get.mockRejectedValueOnce(new Error("Declaration unavailable"));
    await owner.submit();
    const receipt = owner.getSnapshot().operation?.receipt;
    if (receipt?.kind !== "accepted") throw new Error("Expected receipt");
    const jobId = receipt.value.job.job_id;
    await vi.waitFor(() =>
      expect(owner.getSnapshot().observations[jobId]?.state).toBe("paused"),
    );
    owner.selectGraphView(b.graph_view_id);
    const count = transport.readJob.mock.calls.length;
    owner.resumeOperationObservation();
    await vi.waitFor(() =>
      expect(transport.readJob.mock.calls.length).toBe(count + 1),
    );
    expect(transport.readJob.mock.calls.at(-1)?.[0]).toMatchObject({
      jobId,
      graphId: a.graph_view_id,
    });
    await owner.reloadOperationGraph();
    expect(transport.get.mock.calls.at(-1)?.[0]).toBe(a.graph_view_id);
    expect(owner.getSnapshot().selectedGraphViewId).toBe(b.graph_view_id);
    expect(owner.getSnapshot().operation?.receipt).toBe(receipt);
    expect(Object.isFrozen(receipt)).toBe(true);
  });

  it("publishes source and scope withdrawal without intermediate protected exposure", async () => {
    for (const kind of ["source", "scope"] as const) {
      const { owner, a } = await setup();
      const exposed: unknown[] = [];
      const unsubscribe = owner.subscribe(() => {
        if (owner.getSnapshot().navigation.result !== null)
          exposed.push(owner.getSnapshot().navigation.result);
      });
      await owner.onResourceChange(
        kind === "scope"
          ? {
              resourceKind: "*",
              resourceId: "*",
              changeKind: "remove",
              reasonCode: "authorization_lost",
            }
          : {
              resourceKind: "network_flow_table",
              resourceId: a.semantic_query.selected_table_ids[0],
              changeKind: "remove",
              reasonCode: "soft_deleted",
            },
      );
      unsubscribe();
      expect(exposed).toEqual([]);
    }
  });

  it("purges resource drafts and all protected operation state on explicit removal", async () => {
    const { owner, a } = await setup();
    owner.openAction("rename");
    owner.setDraft("Protected name");
    owner.removeGraph(a.graph_view_id);
    expect(owner.getSnapshot().operation).toBeNull();
    owner.selectGraphView(owner.getSnapshot().graphs[0]?.graph_view_id ?? null);
    owner.openAction("rename");
    await owner.onResourceChange({
      resourceKind: "*",
      resourceId: "*",
      changeKind: "remove",
      reasonCode: "authorization_lost",
    });
    expect(owner.getSnapshot().operation).toBeNull();
    expect(owner.getSnapshot().notice).toBeNull();
    expect(owner.getSnapshot().observations).toEqual({});
    const other = await setup();
    other.owner.openAction("rename");
    other.transport.get.mockRejectedValueOnce(
      failure("authorization_denied", 403),
    );
    await other.owner.reconcileGraph(other.a.graph_view_id);
    expect(other.owner.getSnapshot().declarationErrors).toEqual({});
    expect(other.owner.getSnapshot().operation).toBeNull();
  });

  it("keeps receipt declarations out of the current catalog when follow-up reads fail", async () => {
    const { owner, transport, a } = await setup();
    owner.openAction("rename");
    owner.setDraft("Acknowledged name");
    transport.receipt.mockResolvedValueOnce({
      kind: "renamed",
      value: { ...a, display_name: "Acknowledged name", graph_view_version: 2 },
    });
    transport.get.mockRejectedValue(new Error("Read unavailable"));
    transport.list.mockRejectedValue(new Error("List unavailable"));
    await owner.submit();
    expect(owner.getSnapshot().operation?.phase).toBe("acknowledged");
    expect(owner.selectedGraph?.display_name).toBe(a.display_name);
  });

  it("retains original uncertainty after a denied replay and resolves only the same attempt", async () => {
    const { owner, transport, a } = await setup();
    owner.openAction("create", a.semantic_query);
    owner.setDraft("Created graph");
    transport.receipt.mockRejectedValueOnce(new Error("Lost acknowledgement"));
    await owner.submit();
    const attempt = owner.getSnapshot().operation?.attempt;
    transport.receipt.mockRejectedValueOnce(
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
    transport.receipt.mockResolvedValueOnce(savedGraphAccepted(a));
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
    transport.receipt.mockRejectedValueOnce(new Error("Lost acknowledgement"));
    await owner.submit();
    await owner.loadGraphs();
    const result = owner.getSnapshot().navigation.result;
    transport.list.mockRejectedValue(new Error("Follow-up unavailable"));
    transport.receipt.mockResolvedValueOnce(
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
    transport.receipt.mockRejectedValueOnce(new Error("Lost acknowledgement"));
    await owner.submit();
    owner.selectGraphView(b.graph_view_id);
    await owner.replay();
    expect(owner.getSnapshot().selectedGraphViewId).toBe(b.graph_view_id);
  });

  it("fences conflict review after retirement without resurrecting a declaration", async () => {
    const { owner, transport, a } = await setup();
    owner.openAction("rename");
    owner.setDraft("Retained draft");
    transport.receipt.mockRejectedValueOnce(
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
    expect(owner.getSnapshot().operation).toBeNull();
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
    transport.receipt.mockReturnValueOnce(pending.promise);
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
    expect(other.owner.getSnapshot().operation).toBeNull();
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
      expect(
        owner.getSnapshot().navigation.result?.result.graph_projection_result
          .projection_result_id,
      ).toBe(changed.selected_result_binding?.projection_result_id),
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
    transport.receipt.mockReturnValueOnce(pending.promise);
    owner.openAction("create", a.semantic_query);
    owner.setDraft(a.display_name);
    const writing = owner.submit();
    await vi.advanceTimersByTimeAsync(30_000);
    await writing;
    transport.receipt.mockRejectedValueOnce(
      failure("client_txn_conflict", 409),
    );
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
