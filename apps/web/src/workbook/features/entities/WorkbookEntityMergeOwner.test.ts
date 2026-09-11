import { afterEach, describe, expect, it, vi } from "vitest";
import {
  mergeAuthority,
  mergeIncidentId,
  mergeLoserId,
  mergeReceipt,
  mergeReview,
  mergeSurvivorId,
} from "../../../testing/entityMergeTestSupport";
import { createWorkbookEntityMergeAdapter } from "../../adapters/createWorkbookEntityMergeAdapter";
import type {
  EntityMergeBinding,
  EntityMergeTransportOutcome,
  WorkbookEntityMergePort,
} from "./entityMergeOperation";
import { WorkbookEntityMergeOwner } from "./WorkbookEntityMergeOwner";

function setup(type: "host" | "identity" = "host") {
  let sequence = 0;
  const ids = { create: vi.fn(() => `merge-${++sequence}`) };
  const coordinate = vi.fn(async () => true);
  const owner = new WorkbookEntityMergeOwner(mergeIncidentId, ids, {
    canReserve: () => true,
    coordinate,
  });
  const send = vi.fn<WorkbookEntityMergePort["send"]>(async () => ({
    kind: "acknowledged",
    receipt: mergeReceipt(type),
  }));
  owner.configure({
    ...createWorkbookEntityMergeAdapter({
      apiBase: undefined,
      incidentId: mergeIncidentId,
    }),
    send,
  });
  owner.setAuthority(mergeAuthority);
  const current = vi.fn(() => true);
  const matches = vi.fn(() => true);
  const acknowledged = vi.fn();
  const reconcile = vi.fn(async () => undefined);
  const projections = vi.fn(async () => undefined);
  owner.registerProjectionRefresh(projections);
  const binding: EntityMergeBinding = {
    isCurrent: current,
    matchesReview: matches,
    acknowledged,
    reconcile,
  };
  const review = mergeReview(type, owner.getSnapshot().generation);
  const admit = () => {
    const attempt = owner.admit(review, binding);
    expect(attempt).not.toBeNull();
    if (attempt === null) throw new Error("Expected merge admission");
    return attempt;
  };
  const entry = () => owner.getSnapshot().entries[0];
  return {
    owner,
    ids,
    send,
    coordinate,
    current,
    matches,
    acknowledged,
    reconcile,
    projections,
    binding,
    review,
    admit,
    entry,
  };
}
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
afterEach(() => {
  vi.useRealTimers();
});

describe("Workbook entity merge ownership", () => {
  it("reserves both records synchronously and requires the reviewed versions after coordination", async () => {
    for (const type of ["host", "identity"] as const) {
      const t = setup(type);
      const earlier = deferred<boolean>();
      t.coordinate.mockImplementation(() => earlier.promise);
      const attempt = t.admit();
      expect(Object.isFrozen(attempt.review.plan)).toBe(true);
      expect(t.owner.blocksRecord(mergeSurvivorId)).toBe(true);
      expect(t.owner.blocksRecord(mergeLoserId)).toBe(true);
      expect(t.owner.blocksRecord("unrelated")).toBe(false);
      expect(t.owner.admit(t.review, t.binding)).toBeNull();
      expect(t.ids.create).toHaveBeenCalledTimes(1);
      const run = t.owner.execute(attempt);
      expect(t.send).not.toHaveBeenCalled();
      t.owner.acceptVersion(mergeLoserId, 3);
      earlier.resolve(true);
      await run;
      expect(t.send).not.toHaveBeenCalled();
      expect(t.entry()).toMatchObject({
        phase: "rejected",
        failure: { kind: "stale_target" },
      });
      expect(t.owner.blocksRecord(mergeSurvivorId)).toBe(false);
    }
  });
  it("retains exact replay after uncertainty without requiring the loser or a current presentation", async () => {
    for (const type of ["host", "identity"] as const) {
      const t = setup(type);
      t.send.mockRejectedValueOnce(new TypeError("lost transport"));
      const attempt = t.admit();
      await t.owner.execute(attempt);
      expect(t.entry()?.phase).toBe("uncertain");
      const bytes = attempt.body;
      t.current.mockReturnValue(false);
      t.matches.mockReturnValue(false);
      t.owner.acceptVersion(mergeLoserId, 3);
      await Promise.all([
        t.owner.replay(attempt.id),
        t.owner.replay(attempt.id),
      ]);
      expect(t.send).toHaveBeenCalledTimes(2);
      expect(t.send.mock.calls[1]?.[0]).toBe(attempt);
      expect(attempt.body).toBe(bytes);
      expect(t.ids.create).toHaveBeenCalledTimes(1);
      expect(t.entry()).toMatchObject({
        phase: "acknowledged",
        receipt: mergeReceipt(type),
        reconciliation: "complete",
      });
      expect(t.acknowledged).not.toHaveBeenCalled();
      expect(t.reconcile).not.toHaveBeenCalled();
    }
  });
  it("preserves uncertainty across transaction conflict, stale version and lock rejection", async () => {
    for (const type of ["host", "identity"] as const)
      for (const failure of [
        {
          kind: "client_txn_conflict" as const,
          message: "Transaction conflict",
        },
        {
          kind: "stale_target" as const,
          publicCode: "row_version_conflict",
          message: "Version changed",
        },
        {
          kind: "retryable" as const,
          publicCode: "record_locked",
          message: "Locked",
        },
      ]) {
        const t = setup(type);
        t.send
          .mockResolvedValueOnce({ kind: "uncertain" })
          .mockResolvedValueOnce({ kind: "rejected", failure });
        const attempt = t.admit();
        await t.owner.execute(attempt);
        await t.owner.replay(attempt.id);
        expect(t.entry()).toMatchObject({ phase: "uncertain", failure });
        t.owner.dismiss(attempt.id);
        expect(t.entry()?.attempt).toBe(attempt);
        expect(t.owner.blocksRecord(mergeLoserId)).toBe(true);
        expect(t.ids.create).toHaveBeenCalledTimes(1);
      }
  });
  it("separates timeout from transport settlement and retains late acknowledgement", async () => {
    vi.useFakeTimers();
    const t = setup();
    const late = deferred<EntityMergeTransportOutcome>();
    t.send.mockImplementation(() => late.promise);
    const attempt = t.admit();
    const run = t.owner.execute(attempt);
    await vi.advanceTimersByTimeAsync(0);
    expect(t.send).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(30_001);
    await run;
    expect(t.entry()).toMatchObject({
      phase: "uncertain",
      transportPending: true,
    });
    await t.owner.replay(attempt.id);
    expect(t.send).toHaveBeenCalledTimes(1);
    t.current.mockReturnValue(false);
    late.resolve({ kind: "acknowledged", receipt: mergeReceipt() });
    await vi.advanceTimersByTimeAsync(0);
    expect(t.entry()).toMatchObject({
      phase: "acknowledged",
      transportPending: false,
      reconciliation: "complete",
    });
    expect(t.acknowledged).not.toHaveBeenCalled();
  });
  it("conceals suspended content, fences session callbacks and clears retired account attempts", async () => {
    for (const change of [
      "suspend",
      "session",
      "role",
      "actor",
      "retire",
    ] as const) {
      const t = setup();
      const late = deferred<EntityMergeTransportOutcome>();
      t.send.mockImplementation(() => late.promise);
      const attempt = t.admit();
      const run = t.owner.execute(attempt);
      await vi.waitFor(() => expect(t.send).toHaveBeenCalledTimes(1));
      if (change === "suspend") t.owner.suspend();
      else if (change === "retire") t.owner.retire();
      else
        t.owner.setAuthority({
          ...mergeAuthority,
          ...(change === "session" ? { sessionIdentity: "replacement" } : {}),
          ...(change === "role" ? { role: "viewer" as const } : {}),
          ...(change === "actor" ? { actorId: "replacement" } : {}),
        });
      if (["suspend", "retire", "actor"].includes(change))
        expect(t.owner.getSnapshot().entries).toEqual([]);
      late.resolve({ kind: "acknowledged", receipt: mergeReceipt() });
      await run;
      expect(t.acknowledged).not.toHaveBeenCalled();
      expect(t.reconcile).not.toHaveBeenCalled();
      if (change === "suspend") {
        expect(t.owner.getSnapshot().entries).toEqual([]);
        t.owner.setAuthority(mergeAuthority);
        expect(t.entry()?.receipt).toEqual(mergeReceipt());
      } else if (change === "retire" || change === "actor")
        expect(t.owner.getSnapshot().entries).toEqual([]);
    }
  });
  it("records the complete receipt before refresh and retries only reconciliation without regressing newer versions", async () => {
    for (const type of ["host", "identity"] as const) {
      const t = setup(type);
      t.projections.mockImplementationOnce(async () => {
        expect(t.entry()?.receipt).toEqual(mergeReceipt(type));
        throw new Error("Refresh failed");
      });
      t.owner.acceptVersion(mergeSurvivorId, 7);
      const attempt = t.admit();
      // A later rollback or write must not be overwritten by an old replay receipt.
      t.send.mockImplementation(async () => {
        t.owner.acceptVersion(mergeSurvivorId, 12);
        t.owner.acceptVersion(mergeLoserId, 9);
        return { kind: "acknowledged", receipt: mergeReceipt(type) };
      });
      await t.owner.execute(attempt);
      expect(t.entry()).toMatchObject({
        phase: "acknowledged",
        reconciliation: "required",
      });
      expect(t.owner.latestVersion(mergeSurvivorId)).toBe(12);
      expect(t.owner.latestVersion(mergeLoserId)).toBe(9);
      expect(t.acknowledged).not.toHaveBeenCalled();
      t.owner.dismiss(attempt.id);
      expect(t.entry()).toBeDefined();
      await t.owner.refresh(attempt.id);
      expect(t.send).toHaveBeenCalledTimes(1);
      expect(t.entry()?.reconciliation).toBe("complete");
      t.owner.dismiss(attempt.id);
      expect(t.owner.getSnapshot().entries).toEqual([]);

      const changed = setup(type);
      const completed = changed.admit();
      await changed.owner.execute(completed);
      const gate = deferred<undefined>();
      changed.projections.mockImplementationOnce(() => gate.promise);
      const refresh = changed.owner.refresh(completed.id, {
        requireAcceptance: true,
      });
      const rejectedRefresh = expect(refresh).rejects.toThrow(
        "did not establish current projections",
      );
      changed.owner.registerProjectionRefresh(async () => {});
      gate.resolve(undefined);
      await rejectedRefresh;
      expect(changed.entry()?.reconciliation).toBe("required");
      expect(changed.entry()?.receipt).toEqual(mergeReceipt(type));
      await changed.owner.refresh(completed.id, { requireAcceptance: true });
      expect(changed.entry()?.reconciliation).toBe("complete");
      expect(changed.send).toHaveBeenCalledTimes(1);
    }
  });
});
