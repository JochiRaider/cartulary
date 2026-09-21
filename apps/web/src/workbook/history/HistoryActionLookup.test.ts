import { describe, expect, it, vi } from "vitest";
import { historyDiffFixture } from "../../testing/workbookHistoryTestSupport";
import { HistoryActionLookup } from "./HistoryActionLookup";
import { HistoryPageLookup } from "./HistoryPageLookup";
import { WorkbookRecordHistoryOwner } from "./WorkbookRecordHistoryOwner";
import type { WorkbookRecordHistoryPendingAction } from "./workbookHistoryItem";
import type { HistoryAttempt } from "./workbookHistoryOperation";
import type {
  HistoryPage,
  HistoryPageProvenance,
  HistoryPageRequest,
} from "./workbookHistoryPage";

const scope = {
  actorId: "actor",
  incidentId: "incident",
  sessionIdentity: "session",
  epoch: 1,
};
const subject = {
  kind: "live" as const,
  recordId: "record",
  viewSchemaId: "view",
  rowVersion: 8,
  label: "Row",
  surfaceLabel: "Generic",
};
const item = {
  history_item_ref: "selected",
  actor_user_id: "actor",
  committed_at: "2026-09-10T00:00:00Z",
  operation: "patch",
  change_set_id: "change",
  reversible: true,
  available_rollback_actions: ["history_entry" as const],
  history_entry_ref: "opaque",
  diff_summary: historyDiffFixture("Edit"),
};
const pending = {
  kind: "rollback" as const,
  recordId: "record",
  rowVersion: 8,
  historyItemRef: "selected",
  action: "history_entry" as const,
  target: { kind: "history_entry" as const, history_entry_ref: "opaque" },
};
function page(next: string | null, selected = false): HistoryPage {
  return {
    record_id: "record",
    incident_id: "incident",
    row_version: 8,
    deleted: false,
    representation_generation: "cartulary.history.1",
    items: selected ? [item] : [],
    paging:
      next === null
        ? { limit: 100, has_more: false, next_cursor: null }
        : { limit: 100, has_more: true, next_cursor: next },
  };
}
function setup(
  pages: HistoryPage[],
  provenance?: HistoryPageProvenance,
  action: WorkbookRecordHistoryPendingAction = pending,
) {
  let current = scope;
  const read = vi.fn(
    async (_request: HistoryPageRequest, _signal: AbortSignal) => ({
      kind: "accepted" as const,
      value: required(pages.shift()),
    }),
  );
  const lookup = new HistoryActionLookup({
    scope,
    recordId: "record",
    viewSchemaId: "view",
    pending: action,
    ...(provenance ? { provenance } : {}),
    currentScope: () => current,
    latestVersion: () => 8,
    read,
  });
  return {
    lookup,
    read,
    replaceSession: () => {
      current = { ...scope, sessionIdentity: "replacement", epoch: 2 };
    },
  };
}

describe("History action lookup", () => {
  it("restarts a deployment transition without accepting a mixed generation action proof", async () => {
    const t = setup([
      page("a"),
      { ...page(null, true), representation_generation: "cartulary.history.2" },
      { ...page(null, true), representation_generation: "cartulary.history.2" },
    ]);
    expect((await t.lookup.run()).phase).toBe("restart_required");
    expect(t.lookup.snapshot.page).toBeNull();
    t.lookup.restart();
    expect((await t.lookup.run()).phase).toBe("matched");
    expect(t.read.mock.calls.at(-1)?.[0]).toEqual({});
  });
  it("publishes bounded navigation without retaining an extra result page payload", async () => {
    const onPage = vi.fn();
    const lookup = new HistoryPageLookup({
      scope,
      recordId: "record",
      viewSchemaId: "view",
      currentScope: () => scope,
      latestVersion: () => 8,
      maxRetainedPages: 3,
      retainResultPage: false,
      unavailable: { kind: "stale_target", message: "Change unavailable" },
      evaluate: () => null,
      onPage,
      read: vi
        .fn()
        .mockResolvedValueOnce({ kind: "accepted", value: page("cursor") })
        .mockResolvedValueOnce({
          kind: "accepted",
          value: page("cursor", true),
        }),
    });
    expect((await lookup.run()).phase).toBe("restart_required");
    expect(onPage).toHaveBeenCalledTimes(1);
    expect(lookup.snapshot.page).toBeNull();
    // Rejected continuation must not pin another raw response beside accepted browsing pages.
    expect(lookup.snapshot.pagesChecked).toBe(2);
  });
  it("pauses after three pages and resumes without claiming absence", async () => {
    const t = setup([page("a"), page("b"), page("c"), page(null, true)]);
    expect((await t.lookup.run()).phase).toBe("paused");
    expect(t.read).toHaveBeenCalledTimes(3);
    expect((await t.lookup.run()).phase).toBe("matched");
    expect(t.read.mock.calls.map(([request]) => request)).toEqual([
      {},
      { cursorToken: "a" },
      { cursorToken: "b" },
      { cursorToken: "c" },
    ]);
  });
  it("retries the failed request while retaining successful progress", async () => {
    const t = setup([page("a"), page(null, true)]);
    const read = required(t.read.getMockImplementation());
    t.read.mockImplementationOnce(read).mockImplementationOnce(async () => {
      throw new Error("transport");
    });
    expect((await t.lookup.run()).phase).toBe("failed");
    expect((await t.lookup.run()).phase).toBe("matched");
    expect(t.read.mock.calls.map(([request]) => request)).toEqual([
      {},
      { cursorToken: "a" },
      { cursorToken: "a" },
    ]);
  });
  it("refetches valid provenance and starts a fresh chain only after absence", async () => {
    const provenance = {
      scope,
      recordId: "record",
      viewSchemaId: "view",
      chainId: 1,
      effectiveLimit: 100,
      generation: 2,
      request: { cursorToken: "origin" },
    };
    const t = setup([page(null), page("new"), page(null, true)], provenance);
    expect((await t.lookup.run()).phase).toBe("matched");
    expect(t.read.mock.calls.map(([request]) => request)).toEqual([
      { cursorToken: "origin" },
      {},
      { cursorToken: "new" },
    ]);
  });
  it("requires explicit restart for non-progressing cursors", async () => {
    const t = setup([page("a"), page("a"), page(null, true)]);
    expect((await t.lookup.run()).phase).toBe("restart_required");
    await t.lookup.run();
    expect(t.read).toHaveBeenCalledTimes(2);
    t.lookup.restart();
    expect((await t.lookup.run()).phase).toBe("matched");
    expect(t.read.mock.calls[2]?.[0]).toEqual({});
    const unrelated = { ...item, history_item_ref: "unrelated" };
    for (const next of [unrelated, { ...unrelated, operation: "rewritten" }]) {
      const overlap = setup([
        { ...page("a"), items: [unrelated] },
        { ...page(null), items: [next] },
      ]);
      expect((await overlap.lookup.run()).phase).toBe("restart_required");
    }
  });
  it("fences duplicate activation and cancelled late responses", async () => {
    const t = setup([]);
    let finish!: (value: { kind: "accepted"; value: HistoryPage }) => void;
    t.read.mockImplementation(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    const checking = t.lookup.run();
    expect((await t.lookup.run()).phase).toBe("checking");
    t.lookup.cancel();
    finish({ kind: "accepted", value: page(null, true) });
    expect((await checking).phase).toBe("cancelled");
    expect(t.read).toHaveBeenCalledTimes(1);
    expect(t.read.mock.calls[0]?.[1].aborted).toBe(true);
  });
  it("rejects session replacement and changed current versions", async () => {
    const t = setup([page(null, true)]);
    t.replaceSession();
    expect((await t.lookup.run()).phase).toBe("failed");
    expect(t.read).not.toHaveBeenCalled();
    const changed = setup([{ ...page(null, true), row_version: 7 }]);
    expect((await changed.lookup.run()).phase).toBe("changed");
  });
  it("finds all advertised selectors and rejects changed eligibility or targets", async () => {
    for (const action of [
      "history_entry",
      "change_set",
      "row_restore",
    ] as const) {
      const target =
        action === "history_entry"
          ? pending.target
          : action === "change_set"
            ? { kind: action, change_set_id: "change" }
            : { kind: action, restore_to_revision_no: 2 };
      const selected = {
        ...item,
        available_rollback_actions: [action],
        revision_no: 2,
      };
      const good = { ...page(null), items: [selected] };
      expect(
        (
          await setup([page("later"), good], undefined, {
            ...pending,
            action,
            target,
          }).lookup.run()
        ).phase,
      ).toBe("matched");
      expect(
        (
          await setup(
            [
              {
                ...good,
                representation_generation: "cartulary.history.1",
                items: [
                  {
                    ...selected,
                    reversible: false,
                    available_rollback_actions: [],
                  },
                ],
              },
            ],
            undefined,
            { ...pending, action, target },
          ).lookup.run()
        ).phase,
      ).toBe("unavailable");
    }
    const changed = setup([
      {
        ...page(null, true),
        representation_generation: "cartulary.history.1",
        items: [{ ...item, history_entry_ref: "changed" }],
      },
    ]);
    expect((await changed.lookup.run()).phase).toBe("unavailable");
    expect((await setup([page(null)]).lookup.run()).phase).toBe("unavailable");
  });
  it("bounds a stalled lookup and resumes its exact unfinished page", async () => {
    vi.useFakeTimers();
    try {
      const t = setup([page("a"), page(null, true)]);
      const read = required(t.read.getMockImplementation());
      t.read
        .mockImplementationOnce(read)
        .mockImplementationOnce(() => new Promise(() => {}));
      const checking = t.lookup.run();
      await vi.advanceTimersByTimeAsync(30_000);
      expect((await checking).phase).toBe("failed");
      expect(t.lookup.snapshot.pagesChecked).toBe(1);
      expect((await t.lookup.run()).phase).toBe("matched");
      expect(t.read.mock.calls.map(([request]) => request)).toEqual([
        {},
        { cursorToken: "a" },
        { cursorToken: "a" },
      ]);
    } finally {
      vi.useRealTimers();
    }
  });

  it("retains one exact admission across batches and cancels before dispatch", async () => {
    const owner = new WorkbookRecordHistoryOwner("incident", {
      create: () => "transaction",
    });
    owner.setAuthority({ ...scope, role: "reviewer", closed: false });
    const pages = [page("a"), page("b"), page("c"), page(null, true)];
    const send = vi.fn(async (_attempt: HistoryAttempt) => ({
      kind: "uncertain" as const,
    }));
    owner.configure({
      load: async () => ({ kind: "accepted", value: required(pages.shift()) }),
      send,
    });
    const attempt = required(
      owner.admit(
        { subject, pending },
        {
          isCurrent: () => true,
          coordinate: async () => 8,
          acknowledged: vi.fn(),
          reconcile: async () => {},
        },
      ),
    );
    await owner.execute(attempt);
    expect(owner.getSnapshot()[0]).toMatchObject({
      phase: "preparing",
      dispatched: false,
      checking: { phase: "paused" },
    });
    expect(send).not.toHaveBeenCalled();
    owner.setAuthority({
      ...scope,
      sessionIdentity: "renewed",
      role: "reviewer",
      closed: false,
    });
    await Promise.all([
      owner.continueChecking(attempt.id),
      owner.continueChecking(attempt.id),
    ]);
    expect(send).toHaveBeenCalledOnce();
    expect(send.mock.calls[0]?.[0]).toBe(attempt);
    owner.retire();
    owner.setAuthority({ ...scope, role: "reviewer", closed: false });
    pages.push(page("a"), page("b"), page("c"));
    const cancelled = required(
      owner.admit(
        { subject, pending },
        {
          isCurrent: () => true,
          coordinate: async () => 8,
          acknowledged: vi.fn(),
          reconcile: async () => {},
        },
      ),
    );
    await owner.execute(cancelled);
    owner.cancelChecking(cancelled.id);
    await owner.continueChecking(cancelled.id);
    expect(owner.getSnapshot()[0]).toMatchObject({
      phase: "rejected",
      dispatched: false,
    });
    expect(send).toHaveBeenCalledOnce();
    owner.retire();
    owner.setAuthority({ ...scope, role: "reviewer", closed: false });
    pages.push(page("a"), page("b"), page("c"), {
      ...page(null, true),
      row_version: 9,
    });
    const stale = required(
      owner.admit(
        { subject, pending },
        {
          isCurrent: () => true,
          coordinate: async () => 8,
          acknowledged: vi.fn(),
          reconcile: async () => {},
        },
      ),
    );
    const capturedBody = stale.body;
    await owner.execute(stale);
    await owner.continueChecking(stale.id);
    expect(owner.getSnapshot()[0]).toMatchObject({
      phase: "rejected",
      dispatched: false,
      checking: { phase: "changed" },
    });
    expect(stale.body).toBe(capturedBody);
    expect(send).toHaveBeenCalledOnce();
  });
});

function required<T>(value: T | undefined | null): T {
  if (value === undefined || value === null)
    throw new Error("Missing history fixture value");
  return value;
}
