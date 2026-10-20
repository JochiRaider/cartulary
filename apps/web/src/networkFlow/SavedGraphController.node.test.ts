import { afterEach, describe, expect, it, vi } from "vitest";
import {
  SavedGraphController,
  type SavedGraphTransport,
  sortSavedGraphs,
} from "./SavedGraphController";
import {
  type SavedGraphAuthority,
  type SavedGraphReceipt,
  SavedGraphWriteError,
} from "./savedGraphOperation";
import {
  deferredSavedGraph,
  savedGraphAccepted,
  savedGraphBindingFixture,
  savedGraphFixture,
  type savedGraphJobFixture,
  savedGraphResultFixture,
  savedGraphTestAuthority,
} from "./savedGraphTestFixtures";

async function setup() {
  let authority: SavedGraphAuthority = {
    ...savedGraphTestAuthority,
    session: {},
  };
  const a = savedGraphFixture(),
    b = savedGraphFixture({
      graph_view_id: `nfgv_${"b".repeat(32)}`,
      display_name: "Graph B",
    });
  const receipt = vi.fn<() => ReturnType<SavedGraphTransport["submit"]>>(
    async () => savedGraphAccepted(a),
  );
  const transport = {
    observationLimit: vi.fn(async () => 4),
    readJob: vi.fn<SavedGraphTransport["readJob"]>(async () => {
      throw new Error("No job read in operation tests");
    }),
    navigation: {
      result: vi.fn<SavedGraphTransport["navigation"]["result"]>(async () => {
        throw new Error("No result in operation tests");
      }),
      contributors: vi.fn<SavedGraphTransport["navigation"]["contributors"]>(
        async () => {
          throw new Error("No contributors in operation tests");
        },
      ),
    },
    list: vi.fn(async () => [a, b]),
    get: vi.fn(async () => a),
    receipt,
    submit: vi.fn<SavedGraphTransport["submit"]>(
      async (_attempt, _signal, dispatch) => {
        dispatch();
        return receipt();
      },
    ),
  };
  let id = 0;
  const identify = vi.fn(() => `attempt-${++id}`);
  const controller = new SavedGraphController({ identify });
  controller.bind(transport, () => authority);
  controller.setActive(true);
  await controller.loadGraphs();
  return {
    controller,
    transport,
    identify,
    a,
    b,
    authority,
    setAuthority: (next: SavedGraphAuthority) => {
      authority = next;
      controller.revalidateAuthority();
    },
  };
}
afterEach(() => {
  vi.useRealTimers();
});
describe("Saved graph captured operations", () => {
  it("fills a freed observation slot while retaining explicit recovery for the paused job", async () => {
    const { controller, transport, a, b } = await setup();
    controller.setActive(false);
    const pending =
      deferredSavedGraph<ReturnType<typeof savedGraphJobFixture>>();
    const other = {
      ...b,
      latest_job_id: "00000000-0000-4000-8000-000000000005",
    };
    transport.observationLimit.mockResolvedValue(1);
    transport.list.mockResolvedValue([]);
    controller.setActive(true);
    await controller.loadGraphs();
    const created = {
      ...a,
      latest_job_id: "00000000-0000-4000-8000-000000000006",
    };
    transport.receipt.mockResolvedValue(savedGraphAccepted(created));
    const reads: string[] = [];
    transport.readJob.mockImplementation((...args: unknown[]) => {
      const target = args[0] as { jobId: string };
      reads.push(target.jobId);
      return pending.promise;
    });
    controller.openAction("create", a.semantic_query);
    controller.setDraft(a.display_name);
    await controller.submit();
    transport.list.mockResolvedValue([created, other]);
    await controller.loadGraphs();
    controller.selectGraphView(other.graph_view_id);
    expect(reads).toEqual([created.latest_job_id]);
    pending.reject(new Error("job read unavailable"));
    await vi.waitFor(() =>
      expect(reads).toEqual([created.latest_job_id, other.latest_job_id]),
    );
    expect(
      controller.getSnapshot().observations[created.latest_job_id]?.state,
    ).toBe("paused");
    controller.dispose();
  });

  it("admits one same-tick submission and keeps the exact confirmed target after selection changes", async () => {
    const { controller, transport, identify, a, b } = await setup();
    const pending = deferredSavedGraph<SavedGraphReceipt>();
    transport.receipt.mockReturnValue(pending.promise);
    controller.openAction("rename");
    controller.setDraft("Renamed A");
    controller.selectGraphView(b.graph_view_id);
    const first = controller.submit(),
      second = controller.submit();
    await Promise.resolve();
    expect(transport.submit).toHaveBeenCalledTimes(1);
    expect(identify).toHaveBeenCalledTimes(1);
    expect(
      transport.submit.mock.calls[0]?.[0].intent.target?.graph_view_id,
    ).toBe(a.graph_view_id);
    pending.resolve({
      kind: "renamed",
      value: { ...a, display_name: "Renamed A", graph_view_version: 2 },
    });
    expect(await first).toBe(true);
    expect(await second).toBe(false);
    expect(controller.getSnapshot().selectedGraphViewId).toBe(b.graph_view_id);
  });
  it("requires current-state review after a metadata or version conflict and retains the draft", async () => {
    const { controller, transport, a } = await setup();
    controller.openAction("rename");
    controller.setDraft("My draft");
    const newer = {
      ...a,
      graph_view_version: 2,
      display_name: "Remote rename",
    };
    transport.list.mockResolvedValue([newer]);
    await controller.loadGraphs();
    expect(await controller.submit()).toBe(false);
    expect(transport.submit).not.toHaveBeenCalled();
    expect(controller.getSnapshot().operation?.phase).toBe("awaiting_review");
    transport.get.mockResolvedValue(newer);
    await controller.reviewCurrent();
    expect(
      controller.getSnapshot().operation?.intent.target?.graph_view_version,
    ).toBe(2);
    expect(controller.getSnapshot().operation?.draft).toBe("My draft");
    expect(transport.submit).not.toHaveBeenCalled();
    transport.receipt.mockRejectedValueOnce(
      new SavedGraphWriteError(
        "version_conflict",
        "rejected",
        "Version conflict",
      ),
    );
    expect(await controller.submit()).toBe(false);
    expect(controller.getSnapshot().operation?.phase).toBe("awaiting_review");
  });
  it("keeps uncertain bytes and transaction identity for explicit replay and blocks changed intent", async () => {
    const { controller, transport, identify, a } = await setup();
    controller.openAction("create", a.semantic_query);
    controller.setDraft("New graph");
    transport.receipt.mockRejectedValueOnce(new Error("lost acknowledgement"));
    await controller.submit();
    const attempt = controller.getSnapshot().operation?.attempt;
    expect(controller.getSnapshot().operation?.phase).toBe("uncertain");
    controller.closeDialog();
    controller.setDraft("Changed");
    controller.openAction("refresh");
    expect(controller.getSnapshot().operation?.attempt).toBe(attempt);
    expect(await controller.submit()).toBe(false);
    transport.receipt.mockResolvedValue(savedGraphAccepted());
    await controller.replay();
    expect(transport.submit.mock.calls[1]?.[0]).toBe(attempt);
    expect(identify).toHaveBeenCalledTimes(1);
    expect(controller.getSnapshot().operation?.phase).toBe("acknowledged");
  });
  it("keeps acknowledgement independent of follow-up list failure and dialog closure", async () => {
    const { controller, transport, a } = await setup();
    controller.openAction("create", a.semantic_query);
    controller.setDraft("New graph");
    transport.list.mockRejectedValue(new Error("read unavailable"));
    controller.closeDialog();
    expect(await controller.submit()).toBe(true);
    await controller.loadGraphs();
    await vi.waitFor(() =>
      expect(controller.getSnapshot().listState).toBe("error"),
    );
    expect(controller.getSnapshot().operation?.phase).toBe("acknowledged");
    expect(controller.getSnapshot().dialogOpen).toBe(false);
  });
  it("resolves the actual uncertain attempt from a late acknowledgement after observation deadline", async () => {
    const { controller, transport, a } = await setup();
    vi.useFakeTimers();
    const pending = deferredSavedGraph<SavedGraphReceipt>();
    transport.receipt.mockReturnValue(pending.promise);
    controller.openAction("create", a.semantic_query);
    controller.setDraft("New graph");
    const submitting = controller.submit();
    await vi.advanceTimersByTimeAsync(30_001);
    expect(await submitting).toBe(false);
    expect(controller.getSnapshot().operation?.phase).toBe("uncertain");
    pending.resolve(savedGraphAccepted());
    await Promise.resolve();
    await Promise.resolve();
    expect(controller.getSnapshot().operation?.phase).toBe("acknowledged");
  });
  it("fences stale lists and receipts after retirement and protected-scope replacement", async () => {
    const { controller, transport, a, authority, setAuthority } = await setup();
    const stale = deferredSavedGraph<ReturnType<typeof savedGraphFixture>[]>();
    transport.list.mockReturnValueOnce(stale.promise);
    const loading = controller.loadGraphs();
    controller.removeGraph(a.graph_view_id);
    stale.resolve([a]);
    await loading;
    expect(
      controller
        .getSnapshot()
        .graphs.some((g) => g.graph_view_id === a.graph_view_id),
    ).toBe(false);
    const pending = deferredSavedGraph<SavedGraphReceipt>();
    transport.receipt.mockReturnValue(pending.promise);
    controller.openAction("create", a.semantic_query);
    controller.setDraft("Draft");
    const submitting = controller.submit();
    setAuthority({ ...authority, session: {} });
    pending.resolve(savedGraphAccepted());
    await submitting;
    expect(controller.getSnapshot().operation).toBeNull();
    expect(controller.getSnapshot().graphs).toEqual([]);
  });
  it("withdraws only affected source bindings and preserves retirement tombstones across stale reads", async () => {
    const { controller, transport, a, b } = await setup();
    const materialized = {
      ...a,
      selected_result_binding: savedGraphBindingFixture,
    };
    transport.list.mockResolvedValue([materialized, b]);
    transport.navigation.result.mockResolvedValue(
      savedGraphResultFixture(materialized),
    );
    await controller.loadGraphs();
    await vi.waitFor(() =>
      expect(controller.getSnapshot().navigation.result).not.toBeNull(),
    );
    const first = controller.getSnapshot().navigation.result;
    await controller.onResourceChange({
      resourceKind: "network_flow_table",
      resourceId: a.semantic_query.selected_table_ids[0] ?? "",
      changeKind: "invalidate",
      reasonCode: "renamed",
    });
    expect(controller.getSnapshot().navigation.result).toBe(first);
    await controller.onResourceChange({
      resourceKind: "network_flow_table",
      resourceId: a.semantic_query.selected_table_ids[0] ?? "",
      changeKind: "remove",
      reasonCode: "soft_deleted",
    });
    expect(controller.getSnapshot().navigation.result).toBeNull();
    await controller.onResourceChange({
      resourceKind: "network_flow_graph_view",
      resourceId: a.graph_view_id,
      changeKind: "remove",
      reasonCode: "soft_deleted",
    });
    await controller.loadGraphs();
    expect(
      controller
        .getSnapshot()
        .graphs.some((graph) => graph.graph_view_id === a.graph_view_id),
    ).toBe(false);
    await controller.onResourceChange({
      resourceKind: "*",
      resourceId: "*",
      changeKind: "remove",
      reasonCode: "authorization_lost",
    });
    await controller.loadGraphs();
    expect(controller.getSnapshot().graphs).toEqual([]);
  });
  it("enforces exact roles at dispatch and preserves duplicate names in code-point order", async () => {
    const { controller, transport, authority, setAuthority, a, b } =
      await setup();
    for (const role of ["viewer", "reviewer", "", null]) {
      setAuthority({ ...authority, role });
      controller.openAction("create", a.semantic_query);
      expect(controller.getSnapshot().operation).toBeNull();
    }
    setAuthority({ ...authority, role: "editor" });
    controller.openAction("retire");
    expect(controller.getSnapshot().operation).toBeNull();
    expect(transport.submit).not.toHaveBeenCalled();
    expect(
      sortSavedGraphs([
        { ...b, display_name: "Same" },
        { ...a, display_name: "Same" },
      ]).map((g) => g.graph_view_id),
    ).toEqual([a.graph_view_id, b.graph_view_id]);
    expect(
      sortSavedGraphs([
        { ...a, display_name: "😀" },
        { ...b, display_name: "\ue000" },
      ])[0]?.graph_view_id,
    ).toBe(b.graph_view_id);
  });
});
