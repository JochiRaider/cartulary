import { afterEach, expect, it, vi } from "vitest";
import {
  decisionAuthority,
  decisionReceipt,
  decisionReplacementId,
  decisionRequest,
  decisionReview,
  decisionRow,
  decisionTargetId,
} from "../../../testing/decisionSupersessionTestSupport";
import { createWorkbookDecisionSupersessionAdapter } from "../../adapters/createWorkbookDecisionSupersessionAdapter";
import type {
  DecisionSupersessionOutcome,
  DecisionSupersessionTransportPort,
} from "./decisionSupersessionOperation";
import { WorkbookDecisionSupersessionOwner } from "./WorkbookDecisionSupersessionOwner";

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
function setup() {
  let id = 0;
  const ids = { create: vi.fn(() => `decision-${++id}`) };
  const coordinate = vi.fn(async () => true);
  const owner = new WorkbookDecisionSupersessionOwner(
    decisionAuthority.incidentId,
    ids,
    { canReserve: () => true, coordinate },
  );
  const send = vi.fn<DecisionSupersessionTransportPort["send"]>(async () => ({
    kind: "acknowledged",
    receipt: decisionReceipt(),
  }));
  owner.configure({
    ...createWorkbookDecisionSupersessionAdapter({
      apiBase: undefined,
      incidentId: decisionAuthority.incidentId,
    }),
    send,
  });
  owner.setAuthority(decisionAuthority);
  const current = vi.fn(() => true),
    matches = vi.fn(() => true),
    reconcile = vi.fn(async () => {}),
    projections = vi.fn(async () => {});
  owner.registerReconciliation(projections);
  const binding = { isCurrent: current, matchesReview: matches, reconcile };
  const review = decisionReview("proposed", owner.getSnapshot().generation);
  const admit = () => {
    const attempt = owner.admit(review, binding);
    if (!attempt) throw new Error("Expected admission");
    return attempt;
  };
  return {
    owner,
    ids,
    send,
    coordinate,
    current,
    matches,
    reconcile,
    projections,
    binding,
    review,
    admit,
    entry: () => owner.getSnapshot().entries[0],
  };
}
afterEach(() => vi.useRealTimers());

it("Decision admission reserves both participants synchronously and rechecks earlier writes", async () => {
  const t = setup(),
    earlier = deferred<boolean>();
  t.coordinate.mockImplementation(() => earlier.promise);
  const attempt = t.admit();
  expect(t.owner.blocksRecord(decisionTargetId)).toBe(true);
  expect(t.owner.blocksRecord(decisionReplacementId)).toBe(true);
  expect(t.owner.admit(t.review, t.binding)).toBeNull();
  const first = t.owner.execute(attempt),
    second = t.owner.execute(attempt);
  expect(t.send).not.toHaveBeenCalled();
  t.owner.acceptVersion(decisionReplacementId, 7);
  earlier.resolve(true);
  await Promise.all([first, second]);
  expect(t.send).not.toHaveBeenCalled();
  expect(t.entry()?.phase).toBe("rejected");
  expect(t.ids.create).toHaveBeenCalledTimes(1);
  expect(JSON.parse(attempt.body)).toEqual(decisionRequest(attempt.id));
  expect(Object.isFrozen(attempt.review.replacement)).toBe(true);
});

it("Decision exact replay survives superseded state and never replaces an uncertain transaction", async () => {
  const t = setup();
  t.send.mockResolvedValueOnce({ kind: "uncertain" });
  const attempt = t.admit();
  await t.owner.execute(attempt);
  t.owner.acceptRow(decisionRow(decisionTargetId, "superseded", 5));
  t.current.mockReturnValue(false);
  t.matches.mockReturnValue(false);
  await t.owner.replay(attempt.id);
  expect(t.send).toHaveBeenCalledTimes(2);
  expect(t.send.mock.calls.map(([request]) => request)).toEqual([
    attempt,
    attempt,
  ]);
  expect(t.coordinate).toHaveBeenCalledTimes(1);
  expect(t.ids.create).toHaveBeenCalledTimes(1);
  expect(t.entry()?.receipt).toEqual(decisionReceipt());
  expect(t.reconcile).not.toHaveBeenCalled();
});

it("Decision replay rejection preserves uncertainty and current authorization controls dispatch", async () => {
  const t = setup();
  t.send.mockResolvedValue({ kind: "uncertain" });
  const attempt = t.admit();
  await t.owner.execute(attempt);
  t.owner.setAuthority({ ...decisionAuthority, role: "viewer" });
  await t.owner.replay(attempt.id);
  expect(t.send).toHaveBeenCalledTimes(1);
  t.owner.setAuthority({
    ...decisionAuthority,
    sessionIdentity: "replacement-session",
  });
  t.send.mockResolvedValue({
    kind: "rejected",
    failure: { kind: "client_txn_conflict", message: "Transaction conflict" },
  });
  await t.owner.replay(attempt.id);
  expect(t.entry()?.phase).toBe("uncertain");
  expect(t.entry()?.attempt).toBe(attempt);
  expect(t.ids.create).toHaveBeenCalledTimes(1);
});

it("Decision review cancellation closure and session changes fence preparation", async () => {
  for (const change of ["review", "closure", "session", "selection"] as const) {
    const t = setup(),
      earlier = deferred<boolean>();
    t.coordinate.mockImplementation(() => earlier.promise);
    const attempt = t.admit(),
      execution = t.owner.execute(attempt);
    if (change === "review") t.matches.mockReturnValue(false);
    if (change === "selection") t.current.mockReturnValue(false);
    if (change === "closure") t.owner.closeIncident();
    if (change === "session")
      t.owner.setAuthority({ ...decisionAuthority, sessionIdentity: "new" });
    earlier.resolve(true);
    await execution;
    expect(t.send).not.toHaveBeenCalled();
    expect(t.entry()?.phase).toBe("rejected");
  }
});

it("Decision timeout retains late acknowledgement while access loss conceals and retirement forgets", async () => {
  vi.useFakeTimers();
  const t = setup(),
    transport = deferred<DecisionSupersessionOutcome>();
  t.send.mockImplementation(() => transport.promise);
  const attempt = t.admit(),
    execution = t.owner.execute(attempt);
  await vi.advanceTimersByTimeAsync(30_001);
  await execution;
  expect(t.entry()).toMatchObject({
    phase: "uncertain",
    transportPending: true,
  });
  await t.owner.replay(attempt.id);
  expect(t.send).toHaveBeenCalledTimes(1);
  t.owner.suspend();
  expect(t.owner.getSnapshot().entries).toEqual([]);
  expect(t.owner.latestRow(decisionTargetId)).toBeNull();
  transport.resolve({ kind: "acknowledged", receipt: decisionReceipt() });
  await vi.advanceTimersByTimeAsync(1);
  expect(t.projections).not.toHaveBeenCalled();
  t.owner.setAuthority({
    ...decisionAuthority,
    sessionIdentity: "new-session",
  });
  expect(t.entry()?.receipt).toEqual(decisionReceipt());
  await t.owner.refresh(attempt.id);
  expect(t.entry()?.reconciliation).toBe("complete");
  expect(t.reconcile).not.toHaveBeenCalled();
  t.owner.setAuthority({ ...decisionAuthority, actorId: "other-account" });
  expect(t.owner.getSnapshot().entries).toEqual([]);
  expect(t.owner.latestVersion(decisionTargetId)).toBeNull();
});

it("Decision acknowledgement precedes refresh debt and refresh-only recovery sends no mutation", async () => {
  const t = setup();
  t.projections.mockImplementation(async () => {
    expect(t.entry()?.receipt).toEqual(decisionReceipt());
    throw new Error("read failed");
  });
  const attempt = t.admit();
  await t.owner.execute(attempt);
  expect(t.entry()).toMatchObject({
    phase: "acknowledged",
    reconciliation: "required",
  });
  t.owner.acceptVersion(decisionTargetId, 9);
  t.owner.acceptVersion(decisionTargetId, 5);
  expect(t.owner.latestVersion(decisionTargetId)).toBe(9);
  t.projections.mockResolvedValue();
  await t.owner.refresh(attempt.id);
  expect(t.entry()?.reconciliation).toBe("complete");
  expect(t.send).toHaveBeenCalledTimes(1);
  expect(t.owner.latestVersion(decisionReplacementId)).toBe(7);
});
