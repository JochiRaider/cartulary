import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import {
  lifecycleAuthority,
  lifecycleDraft,
  lifecycleIndicator,
  lifecycleReceipt,
  lifecycleRow,
} from "../../../testing/indicatorLifecycleTestSupport";
import { createIndicatorLifecycleAdapter } from "../../adapters/createIndicatorLifecycleAdapter";
import type {
  IndicatorLifecycleTransportPort,
  LifecycleOutcome,
} from "./indicatorLifecycleOperation";
import { WorkbookIndicatorLifecycleOwner } from "./WorkbookIndicatorLifecycleOwner";

function setup() {
  let n = 0;
  const ids = { create: vi.fn(() => `interval-${++n}`) },
    coordinate = vi.fn(async () => true),
    accepted = vi.fn();
  const owner = new WorkbookIndicatorLifecycleOwner(
    lifecycleAuthority.incidentId,
    ids,
    { canReserve: () => true, coordinate, accepted },
  );
  const send = vi.fn<IndicatorLifecycleTransportPort["send"]>(async () => ({
    kind: "acknowledged",
    receipt: lifecycleReceipt(),
  }));
  const records = vi.fn<IndicatorLifecycleTransportPort["records"]>(
    async () => ({
      kind: "accepted",
      value: { items: [lifecycleRow(2)], hasMore: false, nextCursor: null },
    }),
  );
  owner.configure({
    ...createIndicatorLifecycleAdapter({
      apiBase: "https://original.test",
      incidentId: lifecycleAuthority.incidentId,
    }),
    send,
    records,
  });
  owner.setAuthority(lifecycleAuthority);
  owner.acceptRow(lifecycleRow());
  owner.drafts.open(lifecycleIndicator, "example.test", 1);
  owner.drafts.update(lifecycleIndicator, lifecycleDraft().values);
  const current = vi.fn(() => true),
    reconcile = vi.fn(async () => {}),
    projections = vi.fn(async () => {});
  owner.registerReconciliation(projections);
  const binding = { isCurrent: current, reconcile };
  const draft = () => {
    const value = owner.drafts.get(lifecycleIndicator);
    if (!value) throw new Error("draft missing");
    return value;
  };
  const admit = () => {
    const a = owner.admit(draft(), binding);
    if (!a) throw new Error("admission missing");
    return a;
  };
  return {
    owner,
    ids,
    coordinate,
    accepted,
    send,
    records,
    current,
    reconcile,
    projections,
    binding,
    draft,
    admit,
    entry: () => owner.getSnapshot().entries[0],
  };
}
afterEach(() => vi.useRealTimers());
it("Lifecycle admission reserves synchronously and rejects changed versions before dispatch", async () => {
  const t = setup(),
    earlier = deferred<boolean>();
  t.coordinate.mockReturnValue(earlier.promise);
  const a = t.admit();
  expect(t.owner.admit(t.draft(), t.binding)).toBeNull();
  const run = t.owner.execute(a);
  await t.owner.execute(a);
  t.owner.acceptVersion(lifecycleIndicator, 2);
  earlier.resolve(true);
  await run;
  expect(t.send).not.toHaveBeenCalled();
  expect(t.entry()?.phase).toBe("rejected");
  expect(t.ids.create).toHaveBeenCalledTimes(1);
  expect(t.draft().baseRowVersion).toBe(1);
  expect(t.owner.admit(t.draft(), t.binding)).toBeNull();
  expect(await t.owner.review(lifecycleIndicator)).toBe(true);
  expect(t.draft().baseRowVersion).toBe(2);
  expect(t.draft().values).toEqual(a.draft.values);
});
it("Lifecycle exact replay retains bytes identity and ordered support through closure and newer versions", async () => {
  const t = setup();
  t.send.mockResolvedValueOnce({ kind: "uncertain" });
  t.owner.drafts.update(lifecycleIndicator, {
    ...t.draft().values,
    support: [
      {
        recordId: lifecycleAuthority.actorId,
        label: "Second",
        viewSchemaId: "cartulary.view.timeline.v1",
      },
      {
        recordId: lifecycleAuthority.incidentId,
        label: "First",
        viewSchemaId: "cartulary.view.timeline.v1",
      },
    ],
  });
  const a = t.admit();
  await t.owner.execute(a);
  expect(Object.isFrozen(a.values.support_refs)).toBe(true);
  expect(Object.isFrozen(a.draft.values)).toBe(true);
  t.owner.acceptVersion(lifecycleIndicator, 99);
  t.current.mockReturnValue(false);
  t.owner.closeIncident();
  t.owner.drafts.update(lifecycleIndicator, {
    ...t.draft().values,
    rationale: "Changed form",
  });
  await t.owner.replay(a.id);
  expect(t.send.mock.calls.map(([attempt]) => attempt)).toEqual([a, a]);
  expect(t.ids.create).toHaveBeenCalledTimes(1);
  expect(t.coordinate).toHaveBeenCalledTimes(1);
  expect(JSON.parse(a.body).support_refs).toEqual([
    lifecycleAuthority.actorId,
    lifecycleAuthority.incidentId,
  ]);
  expect(t.owner.latestVersion(lifecycleIndicator)).toBe(99);
  expect(t.reconcile).not.toHaveBeenCalled();
});
it("Lifecycle replay uses current authority and access rejection cannot erase uncertainty", async () => {
  const t = setup();
  t.send.mockResolvedValueOnce({ kind: "uncertain" });
  const a = t.admit();
  await t.owner.execute(a);
  t.owner.setAuthority({ ...lifecycleAuthority, role: "viewer" });
  await t.owner.replay(a.id);
  expect(t.send).toHaveBeenCalledTimes(1);
  t.owner.setAuthority(lifecycleAuthority);
  t.send.mockResolvedValueOnce({
    kind: "rejected",
    failure: {
      kind: "authorization_lost",
      message: "Lost access",
      publicCode: "forbidden",
    },
  });
  await t.owner.replay(a.id);
  expect(t.owner.getSnapshot().entries).toEqual([]);
  t.owner.setAuthority(lifecycleAuthority);
  expect(t.entry()?.phase).toBe("uncertain");
  expect(t.draft().values).toEqual(a.draft.values);
  t.owner.setAuthority({
    ...lifecycleAuthority,
    actorId: lifecycleAuthority.incidentId,
  });
  expect(t.owner.getSnapshot().entries).toEqual([]);
  expect(t.owner.drafts.get(lifecycleIndicator)).toBeNull();
});
it("Lifecycle fresh dispatch rechecks role closure subject and secure identity without discarding drafts", async () => {
  for (const change of ["role", "closed", "subject", "draft"]) {
    const t = setup(),
      earlier = deferred<boolean>();
    t.coordinate.mockReturnValue(earlier.promise);
    const a = t.admit();
    const run = t.owner.execute(a);
    if (change === "role")
      t.owner.setAuthority({ ...lifecycleAuthority, role: "viewer" });
    if (change === "closed") t.owner.closeIncident();
    if (change === "subject") t.current.mockReturnValue(false);
    if (change === "draft")
      t.owner.drafts.update(lifecycleIndicator, {
        ...t.draft().values,
        rationale: "New",
      });
    earlier.resolve(true);
    await run;
    expect(t.send).not.toHaveBeenCalled();
    expect(t.owner.drafts.get(lifecycleIndicator)).not.toBeNull();
  }
  const t = setup();
  t.ids.create.mockImplementation(() => {
    throw new Error("Crypto unavailable");
  });
  expect(t.owner.admit(t.draft(), t.binding)).toBeNull();
  expect(t.owner.getSnapshot().entries).toEqual([]);
  expect(t.send).not.toHaveBeenCalled();
  expect(t.owner.blocksRecord(lifecycleIndicator)).toBe(false);
});
it("Lifecycle late settlement survives detachment while retired accounts cannot materialize receipts", async () => {
  vi.useFakeTimers();
  for (const retire of [false, true]) {
    const t = setup(),
      late = deferred<LifecycleOutcome>();
    t.send.mockReturnValue(late.promise);
    const a = t.admit(),
      run = t.owner.execute(a);
    await vi.advanceTimersByTimeAsync(30001);
    await run;
    expect(t.entry()?.phase).toBe("uncertain");
    expect(t.entry()?.transportPending).toBe(true);
    await t.owner.replay(a.id);
    expect(t.send).toHaveBeenCalledTimes(1);
    t.owner.suspend();
    if (retire) t.owner.retire();
    late.resolve({ kind: "acknowledged", receipt: lifecycleReceipt() });
    await vi.advanceTimersByTimeAsync(0);
    expect(t.owner.getSnapshot().entries).toEqual([]);
    t.owner.setAuthority(lifecycleAuthority);
    expect(t.entry()?.phase).toBe(retire ? undefined : "acknowledged");
    expect(t.accepted).toHaveBeenCalledTimes(retire ? 0 : 1);
    expect(t.reconcile).not.toHaveBeenCalled();
  }
});
it("Lifecycle receipt precedes projection failure and refresh recovery never sends a mutation", async () => {
  const t = setup();
  t.projections.mockImplementation(async () => {
    expect(t.entry()?.phase).toBe("acknowledged");
    expect(t.owner.latestVersion(lifecycleIndicator)).toBe(2);
    throw new Error("read failed");
  });
  const a = t.admit();
  await t.owner.execute(a);
  expect(t.entry()?.reconciliation).toBe("required");
  expect(t.entry()).toMatchObject({
    phase: "acknowledged",
    receipt: lifecycleReceipt(),
  });
  t.projections.mockResolvedValue(undefined);
  await t.owner.refresh(a.id);
  expect(t.entry()?.reconciliation).toBe("complete");
  expect(t.send).toHaveBeenCalledTimes(1);
  expect(t.accepted).toHaveBeenCalledTimes(1);
});
