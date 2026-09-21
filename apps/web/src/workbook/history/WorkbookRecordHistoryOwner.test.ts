import { waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { historyDiffFixture } from "../../testing/workbookHistoryTestSupport";
import type { RecordHistoryData } from "../adapters/workbookHistoryResponse";
import { WorkbookRecordHistoryOwner } from "./WorkbookRecordHistoryOwner";
import type { WorkbookRecordHistoryPendingAction } from "./workbookHistoryItem";
import {
  type HistoryAuthority,
  type HistoryBinding,
  type HistoryReceipt,
  historyActionPermitted,
  type WorkbookRecordHistoryPort,
} from "./workbookHistoryOperation";

type HistoryTransportOutcome = Awaited<
  ReturnType<WorkbookRecordHistoryPort["send"]>
>;

const authority: HistoryAuthority = {
  actorId: "actor",
  incidentId: "incident",
  role: "reviewer",
  closed: false,
};
const subject = {
  kind: "live" as const,
  recordId: "record",
  rowVersion: 4,
  label: "A row",
  stateLabel: "Deleted",
  surfaceLabel: "Timeline",
  viewSchemaId: "cartulary.view.timeline.v2",
};
const item = {
  actor_user_id: "actor",
  committed_at: "2026-09-10T00:00:00Z",
  history_item_ref: "item",
  history_entry_ref: "opaque-selector",
  change_set_id: "change",
  revision_no: 2,
  reversible: true,
  operation: "patch",
  diff_summary: historyDiffFixture("Changed fields"),
  available_rollback_actions: [
    "history_entry",
    "change_set",
    "row_restore",
  ] as const,
};
function setup(
  pending: WorkbookRecordHistoryPendingAction = {
    kind: "destructive",
    operation: "delete",
    recordId: "record",
    rowVersion: 4,
  },
) {
  let history: RecordHistoryData = {
    record_id: "record",
    incident_id: "incident",
    row_version: 4,
    deleted: pending.kind === "destructive" && pending.operation === "restore",
    representation_generation: "cartulary.history.1",
    items: [item],
  };
  const ids = { create: vi.fn(() => `txn-${ids.create.mock.calls.length}`) };
  const owner = new WorkbookRecordHistoryOwner("incident", ids);
  owner.setAuthority(authority);
  const receipt: HistoryReceipt =
    pending.kind === "rollback"
      ? {
          kind: "rollback",
          incidentId: "incident",
          recordId: "record",
          rowVersion: 5,
          changeSetId: "rollback-change",
          target: pending.target,
          affectedRecordIds: ["other", "record"],
        }
      : {
          kind: pending.operation,
          incidentId: "incident",
          recordId: "record",
          rowVersion: 5,
          changeSetId: "change-result",
          deleted: pending.operation === "delete",
          deletedAt:
            pending.operation === "delete" ? "2026-09-10T00:00:00Z" : null,
          deletedByUserId: pending.operation === "delete" ? "actor" : null,
        };
  const send = vi.fn<WorkbookRecordHistoryPort["send"]>(async () => {
    history = {
      ...history,
      row_version: 5,
      deleted: receipt.kind === "delete",
    };
    return { kind: "acknowledged", receipt };
  });
  const load = vi.fn(async () => ({
    kind: "accepted" as const,
    value: {
      ...history,
      paging: { limit: 100, has_more: false as const, next_cursor: null },
    },
  }));
  owner.configure({ send, load });
  const binding: HistoryBinding = {
    isCurrent: () => true,
    coordinate: vi.fn(async () => 4),
    acknowledged: vi.fn(),
    reconcile: vi.fn(async () => {}),
  };
  const intent = {
    pending,
    subject: {
      ...subject,
      kind: history.deleted ? ("deleted" as const) : ("live" as const),
    },
  };
  return {
    owner,
    ids,
    send,
    load,
    binding,
    intent,
    receipt,
    setHistory: (data: Partial<RecordHistoryData>) => {
      history = { ...history, ...data };
    },
  };
}
afterEach(() => vi.useRealTimers());

describe("Workbook history operation owner", () => {
  it("keeps history authority for an unavailable related projection and still suspends genuine session loss", async () => {
    const t = setup();
    const recover = vi.fn();
    const load = vi
      .fn<WorkbookRecordHistoryPort["load"]>()
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "terminal",
          publicCode: "record_not_found",
          message: "Unavailable record",
        },
      })
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "stale_target",
          publicCode: "record_not_found",
          message: "Unavailable record",
        },
      })
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "authentication_required",
          message: "Session expired",
        },
      });
    t.owner.configure({ load, send: t.send }, recover);
    expect((await t.owner.loadProjection("related")).kind).toBe("rejected");
    expect(t.owner.readable).toBe(true);
    expect((await t.owner.load("record")).kind).toBe("rejected");
    expect(t.owner.readable).toBe(true);
    expect(recover).not.toHaveBeenCalled();
    expect((await t.owner.loadProjection("related")).kind).toBe("rejected");
    expect(t.owner.readable).toBe(false);
    expect(recover).toHaveBeenCalledOnce();
    expect(t.send).not.toHaveBeenCalled();
  });
  it("retains admitted work after presentation closes without detached effects", async () => {
    const t = setup();
    let current = true;
    const binding = { ...t.binding, isCurrent: () => current };
    const attempt = t.owner.admit(t.intent, binding);
    expect(attempt).not.toBeNull();
    if (!attempt) return;
    current = false;
    await t.owner.execute(attempt);
    expect(t.send).toHaveBeenCalledOnce();
    expect(binding.coordinate).toHaveBeenCalledOnce();
    expect(binding.acknowledged).not.toHaveBeenCalled();
    expect(binding.reconcile).not.toHaveBeenCalled();
    expect(t.owner.getSnapshot()[0]?.phase).toBe("acknowledged");
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("required"),
    );
  });
  it("enforces operation-specific roles and incident closure", () => {
    for (const role of ["viewer", "editor", "reviewer", "admin"] as const) {
      expect(historyActionPermitted({ ...authority, role }, "delete")).toBe(
        role !== "viewer",
      );
      for (const operation of ["restore", "rollback"] as const)
        expect(historyActionPermitted({ ...authority, role }, operation)).toBe(
          role === "reviewer" || role === "admin",
        );
      expect(
        historyActionPermitted({ ...authority, role, closed: true }, "delete"),
      ).toBe(false);
    }
  });
  it.each([
    "delete",
    "restore",
    "history_entry",
    "change_set",
    "row_restore",
  ] as const)("retains exact identity and acknowledgement for %s", async (kind) => {
    const pending: WorkbookRecordHistoryPendingAction =
      kind === "delete" || kind === "restore"
        ? {
            kind: "destructive",
            operation: kind,
            recordId: "record",
            rowVersion: 4,
          }
        : {
            kind: "rollback",
            action: kind,
            historyItemRef: "item",
            recordId: "record",
            rowVersion: 4,
            target:
              kind === "history_entry"
                ? { kind, history_entry_ref: "opaque-selector" }
                : kind === "change_set"
                  ? { kind, change_set_id: "change" }
                  : { kind, restore_to_revision_no: 2 },
          };
    const t = setup(pending);
    t.send.mockResolvedValueOnce({ kind: "uncertain" });
    const attempt = t.owner.admit(t.intent, t.binding);
    expect(attempt).not.toBeNull();
    if (!attempt) return;
    expect(t.owner.admit(t.intent, t.binding)).toBeNull();
    await t.owner.execute(attempt);
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.transportPending).toBe(false),
    );
    expect(t.owner.getSnapshot()[0]?.phase).toBe("uncertain");
    await Promise.all([t.owner.replay(attempt.id), t.owner.replay(attempt.id)]);
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.phase).toBe("acknowledged"),
    );
    expect(t.send).toHaveBeenCalledTimes(2);
    expect(t.ids.create).toHaveBeenCalledTimes(1);
    expect(t.send.mock.calls[0]?.[0]).toBe(t.send.mock.calls[1]?.[0]);
    expect(t.owner.getSnapshot()[0]?.receipt).toEqual(t.receipt);
  });
  it("invalidates confirmed versions and selectors instead of rebasing", async () => {
    const t = setup();
    const attempt = t.owner.admit(t.intent, t.binding);
    if (!attempt) throw new Error("admission");
    t.setHistory({ row_version: 6 });
    await t.owner.execute(attempt);
    expect(t.send).not.toHaveBeenCalled();
    expect(t.owner.getSnapshot()[0]?.dispatched).toBe(false);
    expect(JSON.parse(attempt.body).base_row_version).toBe(4);
  });
  it("reauthorizes after sequencing and never resumes automatically after reopening", async () => {
    const t = setup();
    let release!: (version: number) => void;
    const binding = {
      ...t.binding,
      coordinate: () =>
        new Promise<number>((resolve) => {
          release = resolve;
        }),
    };
    const attempt = t.owner.admit(t.intent, binding);
    if (!attempt) throw new Error("admission");
    const completion = t.owner.execute(attempt);
    t.owner.closeIncident();
    t.owner.setAuthority(authority);
    release(4);
    await completion;
    t.owner.setAuthority(authority);
    expect(t.send).not.toHaveBeenCalled();
  });
  it("acknowledges before failed refresh and refresh recovery never resends", async () => {
    const t = setup();
    let release!: () => void;
    const reconcile = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        }),
    );
    const attempt = t.owner.admit(t.intent, { ...t.binding, reconcile });
    if (!attempt) throw new Error("admission");
    await t.owner.execute(attempt);
    await waitFor(() => expect(reconcile).toHaveBeenCalledOnce());
    expect(t.owner.getSnapshot()[0]?.phase).toBe("acknowledged");
    expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("refreshing");
    release();
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("complete"),
    );
    await t.owner.refresh(attempt.id);
    expect(t.send).toHaveBeenCalledOnce();
  });
  it("retains receipt on refresh failure and does not regress a newer record", async () => {
    const t = setup();
    const binding = {
      ...t.binding,
      reconcile: vi.fn(async () => {
        throw new Error("refresh");
      }),
    };
    t.owner.acceptVersion("record", 9);
    t.send.mockResolvedValueOnce({ kind: "uncertain" });
    // Admit at a reviewed version before observing the newer accepted version.
    const u = setup();
    u.send.mockResolvedValueOnce({ kind: "uncertain" });
    const attempt = u.owner.admit(u.intent, binding);
    if (!attempt) throw new Error("admission");
    await u.owner.execute(attempt);
    await waitFor(() =>
      expect(u.owner.getSnapshot()[0]?.transportPending).toBe(false),
    );
    u.owner.acceptVersion("record", 9);
    u.setHistory({ row_version: 9 });
    u.send.mockResolvedValue({ kind: "acknowledged", receipt: u.receipt });
    await u.owner.replay(attempt.id);
    await waitFor(() =>
      expect(u.owner.getSnapshot()[0]?.reconciliation).toBe("required"),
    );
    expect(u.owner.getSnapshot()[0]?.receipt).toEqual(u.receipt);
    expect(u.owner.latestVersion("record")).toBe(9);
    expect(binding.acknowledged).not.toHaveBeenCalled();
  });
  it("retains same-account uncertainty while hiding it during session loss and retires another actor", async () => {
    const t = setup();
    t.send.mockResolvedValue({ kind: "uncertain" });
    const attempt = t.owner.admit(t.intent, t.binding);
    if (!attempt) throw new Error("admission");
    await t.owner.execute(attempt);
    t.owner.suspend();
    expect(t.owner.getSnapshot()).toEqual([]);
    t.owner.setAuthority(authority);
    expect(t.owner.getSnapshot()[0]?.attempt).toBe(attempt);
    t.owner.setAuthority({ ...authority, actorId: "someone-else" });
    expect(t.owner.getSnapshot()).toEqual([]);
  });
  it("retains acknowledgement after same-actor session replacement without old callbacks", async () => {
    const t = setup();
    t.owner.setAuthority({ ...authority, sessionIdentity: "session-one" });
    let respond!: (outcome: HistoryTransportOutcome) => void;
    t.send.mockImplementation(
      () =>
        new Promise((resolve) => {
          respond = resolve;
        }),
    );
    const attempt = t.owner.admit(t.intent, t.binding);
    if (!attempt) throw new Error("admission");
    const completion = t.owner.execute(attempt);
    await waitFor(() => expect(t.send).toHaveBeenCalledOnce());
    t.owner.suspend();
    t.owner.setAuthority({ ...authority, sessionIdentity: "session-two" });
    t.setHistory({ row_version: 5, deleted: true });
    respond({ kind: "acknowledged", receipt: t.receipt });
    await completion;
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("required"),
    );
    expect(t.owner.getSnapshot()[0]?.receipt).toEqual(t.receipt);
    expect(t.binding.acknowledged).not.toHaveBeenCalled();
    expect(t.binding.reconcile).not.toHaveBeenCalled();
  });
  it("requires explicit conflict review before allocating a replacement ID", async () => {
    const t = setup();
    t.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "client_txn_conflict", message: "Conflict" },
    });
    const attempt = t.owner.admit(t.intent, t.binding);
    if (!attempt) throw new Error("admission");
    await t.owner.execute(attempt);
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.transportPending).toBe(false),
    );
    await t.owner.retryWithNewId(attempt.id);
    expect(t.ids.create).toHaveBeenCalledOnce();
    await t.owner.review(attempt.id);
    await t.owner.retryWithNewId(attempt.id);
    expect(t.ids.create).toHaveBeenCalledTimes(2);
    expect(t.send).toHaveBeenCalledTimes(2);
    const stale = setup();
    stale.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "client_txn_conflict", message: "Conflict" },
    });
    const reserved = stale.owner.admit(stale.intent, stale.binding);
    if (!reserved) throw new Error("admission");
    await stale.owner.execute(reserved);
    await waitFor(() =>
      expect(stale.owner.getSnapshot()[0]?.transportPending).toBe(false),
    );
    stale.owner.acceptVersion("record", reserved.pending.rowVersion + 1);
    await stale.owner.review(reserved.id);
    expect(stale.owner.getSnapshot()[0]?.reviewState?.phase).toBe("changed");
    expect(stale.owner.canReplace(reserved.id)).toBe(false);
    await stale.owner.retryWithNewId(reserved.id);
    expect(stale.ids.create).toHaveBeenCalledOnce();
    expect(stale.send).toHaveBeenCalledOnce();
  });
  it("bounds observation retains a late receipt and excludes concurrent transport", async () => {
    vi.useFakeTimers();
    const t = setup();
    let respond!: (value: HistoryTransportOutcome) => void;
    t.send.mockImplementation(
      () =>
        new Promise((resolve) => {
          respond = resolve;
        }),
    );
    const attempt = t.owner.admit(t.intent, t.binding);
    if (!attempt) throw new Error("admission");
    const completion = t.owner.execute(attempt);
    await vi.advanceTimersByTimeAsync(1);
    await vi.advanceTimersByTimeAsync(30_000);
    await completion;
    expect(t.owner.getSnapshot()[0]?.phase).toBe("uncertain");
    await t.owner.replay(attempt.id);
    expect(t.send).toHaveBeenCalledOnce();
    respond({ kind: "acknowledged", receipt: t.receipt });
    await vi.advanceTimersByTimeAsync(1);
    expect(t.owner.getSnapshot()[0]?.phase).toBe("acknowledged");
  });
  it("bounds blocked sequencing before dispatch", async () => {
    vi.useFakeTimers();
    const t = setup();
    const attempt = t.owner.admit(t.intent, {
      ...t.binding,
      coordinate: () => new Promise(() => {}),
    });
    if (!attempt) throw new Error("admission");
    const completion = t.owner.execute(attempt);
    await vi.advanceTimersByTimeAsync(30_001);
    await completion;
    expect(t.owner.getSnapshot()[0]?.phase).toBe("rejected");
    expect(t.owner.getSnapshot()[0]?.dispatched).toBe(false);
    expect(t.send).not.toHaveBeenCalled();
  });
  it("fences timed-out reconciliation effects when refresh is retried", async () => {
    vi.useFakeTimers();
    const t = setup();
    let finish!: () => void;
    const firstRefresh = new Promise<void>((resolve) => {
      finish = resolve;
    });
    const effect = vi.fn();
    const reconcile = vi.fn(
      async (_receipt: HistoryReceipt, current: () => boolean) => {
        if (reconcile.mock.calls.length === 1) await firstRefresh;
        if (current()) effect();
      },
    );
    const attempt = t.owner.admit(t.intent, { ...t.binding, reconcile });
    if (!attempt) throw new Error("admission");
    await t.owner.execute(attempt);
    await vi.advanceTimersByTimeAsync(30_001);
    expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("required");
    await t.owner.refresh(attempt.id);
    expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("complete");
    finish();
    await vi.advanceTimersByTimeAsync(1);
    expect(effect).toHaveBeenCalledOnce();
    expect(t.send).toHaveBeenCalledOnce();
  });
  it("invalidates peer changes to advertised eligibility for every rollback selector", async () => {
    for (const action of [
      "history_entry",
      "change_set",
      "row_restore",
    ] as const) {
      const target =
        action === "history_entry"
          ? { kind: action, history_entry_ref: "opaque-selector" }
          : action === "change_set"
            ? { kind: action, change_set_id: "change" }
            : { kind: action, restore_to_revision_no: 2 };
      const t = setup({
        kind: "rollback",
        action,
        target,
        historyItemRef: "item",
        recordId: "record",
        rowVersion: 4,
      });
      const attempt = t.owner.admit(t.intent, t.binding);
      if (!attempt) throw new Error("admission");
      t.setHistory({ items: [{ ...item, available_rollback_actions: [] }] });
      await t.owner.execute(attempt);
      expect(t.send).not.toHaveBeenCalled();
      expect(t.owner.getSnapshot()[0]?.failure?.kind).toBe("stale_target");
    }
  });
  it("keeps uncertainty local to its record and fails locally without secure identity", async () => {
    const t = setup();
    t.send.mockResolvedValue({ kind: "uncertain" });
    const attempt = t.owner.admit(t.intent, t.binding);
    if (!attempt) throw new Error("admission");
    await t.owner.execute(attempt);
    expect(t.owner.admit(t.intent, t.binding)).toBeNull();
    expect(
      t.owner.admit(
        {
          subject: { ...t.intent.subject, recordId: "other" },
          pending: { ...t.intent.pending, recordId: "other" },
        },
        t.binding,
      ),
    ).not.toBeNull();
    const insecure = new WorkbookRecordHistoryOwner("incident", {
      create: () => {
        throw new Error("unavailable");
      },
    });
    insecure.setAuthority(authority);
    expect(insecure.admit(t.intent, t.binding)).toBeNull();
    expect(insecure.getSnapshot()).toEqual([]);
  });
  it("keeps multi-record acknowledgement when a related source projection fails", async () => {
    const t = setup({
      kind: "rollback",
      action: "change_set",
      target: { kind: "change_set", change_set_id: "change" },
      historyItemRef: "item",
      recordId: "record",
      rowVersion: 4,
    });
    const related = vi
      .fn(async () => {})
      .mockRejectedValueOnce(new Error("related projection"));
    t.owner.registerRelatedProjectionRefresh(related);
    const attempt = t.owner.admit(t.intent, t.binding);
    if (!attempt) throw new Error("admission");
    await t.owner.execute(attempt);
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("required"),
    );
    expect(t.owner.getSnapshot()[0]?.receipt).toEqual(t.receipt);
    expect(t.binding.reconcile).not.toHaveBeenCalled();
    await t.owner.refresh(attempt.id);
    expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("complete");
    expect(related).toHaveBeenCalledTimes(2);
    expect(t.binding.reconcile).toHaveBeenCalledTimes(1);
    expect(t.send).toHaveBeenCalledTimes(1);
  });
  it("retains authorized closed-incident reads and hides history on access loss", async () => {
    const t = setup();
    t.owner.closeIncident();
    expect((await t.owner.load("record")).kind).toBe("accepted");
    expect(t.owner.permitted("restore")).toBe(false);
    t.owner.configure({
      send: t.send,
      load: async () => ({
        kind: "rejected",
        failure: { kind: "authorization_lost", message: "Access required" },
      }),
    });
    await t.owner.load("record");
    expect(t.owner.readable).toBe(false);
    expect(t.owner.getSnapshot()).toEqual([]);
  });
  it("retires actor and incident callbacks while an admitted response is pending", async () => {
    const t = setup();
    let respond!: (outcome: HistoryTransportOutcome) => void;
    t.send.mockImplementation(
      () =>
        new Promise((resolve) => {
          respond = resolve;
        }),
    );
    const attempt = t.owner.admit(t.intent, t.binding);
    if (!attempt) throw new Error("admission");
    const completion = t.owner.execute(attempt);
    await waitFor(() => expect(t.send).toHaveBeenCalledOnce());
    t.owner.retire();
    t.owner.setAuthority({ ...authority, actorId: "new-actor" });
    respond({ kind: "acknowledged", receipt: t.receipt });
    await completion;
    expect(t.owner.getSnapshot()).toEqual([]);
    expect(t.binding.acknowledged).not.toHaveBeenCalled();
    expect(t.binding.reconcile).not.toHaveBeenCalled();
  });
});
