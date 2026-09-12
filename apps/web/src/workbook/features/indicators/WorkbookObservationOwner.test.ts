import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import {
  observationAuthority,
  observationCreateIntent,
  observationOwnerFixture,
  observationSource,
  observationTargetId,
  observationTargetRow,
  testObservation,
  testObservationReceipt,
} from "../../../testing/observationTestSupport";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { WorkbookMutationRuntimeRegistry } from "../../runtime/WorkbookMutationRuntimeRegistry";
import type { ObservationOutcome } from "./observationOperation";

afterEach(() => vi.useRealTimers());
it("Observation runtime retains uncertainty across shell recovery with save status and retires account state", async () => {
  const t = observationOwnerFixture(),
    registry = new WorkbookMutationRuntimeRegistry(),
    scope = {
      incidentId: observationAuthority.incidentId,
      clientInstanceId: "tab",
    };
  const create = () =>
    new WorkbookMutationRuntime(scope, t.ids, { execute: vi.fn() });
  const runtime = registry.acquire(scope, create),
    owner = runtime.indicatorObservations;
  owner.configure(t.reader, t.transport);
  owner.setAuthority(observationAuthority);
  owner.drafts.ensure("private");
  t.transport.send.mockResolvedValueOnce({ kind: "uncertain" });
  const a = owner.admit(observationCreateIntent, t.binding);
  if (!a) throw new Error("admission");
  await owner.execute(a);
  expect(runtime.getSnapshot().primaryLabel).toBe("Conflict");
  registry.sessionUnavailable();
  expect(owner.getSnapshot().entries).toEqual([]);
  expect(registry.acquire(scope, create)).toBe(runtime);
  owner.setAuthority(observationAuthority);
  expect(owner.getSnapshot().entries[0]?.attempt).toBe(a);
  await owner.replay(a.id);
  expect(runtime.history.latestVersion(observationSource.recordId)).toBe(5);
  expect(runtime.resolveSocketClientTxn(a.id)).toBe(true);
  registry.replaceAccount();
  expect(owner.getSnapshot().entries).toEqual([]);
  expect(owner.drafts.get("private")).toBeUndefined();
  registry.dispose();
});
it("Observation admission synchronously reserves one operation and rechecks preparation races", async () => {
  const t = observationOwnerFixture(),
    pending = deferred<boolean>();
  t.binding.prepare.mockReturnValue(pending.promise);
  const a = t.admit();
  expect(t.owner.admit(observationCreateIntent, t.binding)).toBeNull();
  const run = t.owner.execute(a);
  await t.owner.execute(a);
  t.binding.matchesDraft.mockReturnValue(false);
  pending.resolve(true);
  await run;
  expect(t.transport.send).not.toHaveBeenCalled();
  expect(t.entry()?.phase).toBe("rejected");
  expect(t.ids.create).toHaveBeenCalledTimes(1);
  t.binding.matchesDraft.mockReturnValue(true);
  t.binding.prepare.mockResolvedValue(true);
  await t.owner.execute(t.admit());
  await t.owner.execute(t.admit());
  expect(t.transport.send).toHaveBeenCalledTimes(2);
  expect(t.transport.send.mock.calls[0]?.[0].id).not.toBe(
    t.transport.send.mock.calls[1]?.[0].id,
  );
});
it("Observation exact replay preserves immutable route body identity through closure and source changes", async () => {
  const t = observationOwnerFixture();
  t.transport.send.mockResolvedValueOnce({ kind: "uncertain" });
  const a = t.admit();
  await t.owner.execute(a);
  expect(Object.isFrozen(a.intent)).toBe(true);
  expect(Object.isFrozen(a.authority)).toBe(true);
  t.owner.acceptVersion(observationSource.recordId, 99);
  t.binding.isCurrent.mockReturnValue(false);
  t.owner.closeIncident();
  await t.owner.replay(a.id);
  expect(t.transport.send.mock.calls.map(([attempt]) => attempt)).toEqual([
    a,
    a,
  ]);
  expect(t.ids.create).toHaveBeenCalledTimes(1);
  expect(t.binding.prepare).toHaveBeenCalledTimes(1);
  expect(t.binding.reconcile).not.toHaveBeenCalled();
  expect(t.owner.latestVersion(observationSource.recordId)).toBe(99);
});
it("Observation authority loss hides retained uncertainty and account retirement removes protected data", async () => {
  const t = observationOwnerFixture();
  t.transport.send.mockResolvedValueOnce({ kind: "uncertain" });
  const a = t.admit();
  await t.owner.execute(a);
  t.owner.setAuthority({ ...observationAuthority, role: "viewer" });
  await t.owner.replay(a.id);
  expect(t.transport.send).toHaveBeenCalledTimes(1);
  t.owner.setAuthority(observationAuthority);
  t.transport.send.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "authorization_lost", message: "Access lost" },
  });
  await t.owner.replay(a.id);
  expect(t.owner.getSnapshot().entries).toEqual([]);
  t.owner.setAuthority(observationAuthority);
  expect(t.entry()?.phase).toBe("uncertain");
  t.owner.drafts.ensure("private");
  t.owner.setAuthority({
    ...observationAuthority,
    actorId: observationTargetId,
  });
  expect(t.owner.getSnapshot().entries).toEqual([]);
  expect(t.owner.drafts.get("private")).toBeUndefined();
});
it("Observation fresh target discovery is independent paged and stale child or target dispatches nothing", async () => {
  const t = observationOwnerFixture();
  t.reader.records.mockResolvedValueOnce({
    kind: "accepted",
    value: {
      items: [
        { ...observationTargetRow, record_id: observationSource.recordId },
      ],
      hasMore: true,
      nextCursor: "opaque",
    },
  });
  await t.owner.execute(
    t.admit({
      action: "resolve",
      observation: testObservation,
      targetId: observationTargetId,
    }),
  );
  expect(t.reader.records.mock.calls.map((call) => call[2])).toEqual([
    null,
    "opaque",
  ]);
  expect(
    JSON.parse(t.transport.send.mock.calls[0]?.[0].body ?? "{}")
      .base_row_version,
  ).toBe(testObservation.row_version);
  t.reader.observations.mockResolvedValue({
    kind: "accepted",
    value: {
      items: [{ ...testObservation, row_version: 8 }],
      hasMore: false,
      nextCursor: null,
    },
  });
  await t.owner.execute(
    t.admit({ action: "dismiss", observation: testObservation }),
  );
  expect(t.transport.send).toHaveBeenCalledTimes(1);
  t.reader.records.mockResolvedValue({
    kind: "accepted",
    value: { items: [], hasMore: false, nextCursor: null },
  });
  await t.owner.execute(
    t.admit({ ...observationCreateIntent, targetId: observationTargetId }),
  );
  expect(t.transport.send).toHaveBeenCalledTimes(1);
});
it("Observation accepts before refresh and retries only reads while retaining all affected versions", async () => {
  const t = observationOwnerFixture();
  const receipt = {
    ...testObservationReceipt,
    affected_records: [
      ...testObservationReceipt.affected_records,
      { record_id: observationTargetId, row_version: 91 },
    ],
  };
  t.transport.send.mockResolvedValue({ kind: "acknowledged", receipt });
  t.projections.mockRejectedValueOnce(new Error("projection unavailable"));
  const a = t.admit();
  await t.owner.execute(a);
  expect(t.entry()).toMatchObject({
    phase: "acknowledged",
    receipt,
    reconciliation: "required",
  });
  expect(t.accepted.mock.invocationCallOrder[0]).toBeLessThan(
    t.projections.mock.invocationCallOrder[0] ?? 0,
  );
  t.owner.acceptVersion(observationTargetId, 92);
  await t.owner.refresh(a.id);
  expect(t.transport.send).toHaveBeenCalledTimes(1);
  expect(t.entry()?.reconciliation).toBe("complete");
  expect(t.owner.latestVersion(observationTargetId)).toBe(92);
});
it("Observation late settlement survives Inspector disposal but cannot complete a retired account", async () => {
  vi.useFakeTimers();
  const t = observationOwnerFixture(),
    response = deferred<ObservationOutcome>();
  t.transport.send.mockReturnValue(response.promise);
  const a = t.admit(),
    run = t.owner.execute(a);
  await vi.advanceTimersByTimeAsync(30_001);
  await run;
  expect(t.entry()).toMatchObject({
    phase: "uncertain",
    transportPending: true,
  });
  t.binding.isCurrent.mockReturnValue(false);
  await t.owner.replay(a.id);
  expect(t.transport.send).toHaveBeenCalledTimes(1);
  response.resolve({ kind: "acknowledged", receipt: testObservationReceipt });
  await vi.advanceTimersByTimeAsync(0);
  expect(t.entry()?.phase).toBe("acknowledged");
  expect(t.binding.reconcile).not.toHaveBeenCalled();
  const u = observationOwnerFixture(),
    late = deferred<ObservationOutcome>();
  u.transport.send.mockReturnValue(late.promise);
  const b = u.admit(),
    running = u.owner.execute(b);
  await vi.advanceTimersByTimeAsync(0);
  u.owner.retire();
  late.resolve({ kind: "acknowledged", receipt: testObservationReceipt });
  await running;
  expect(u.accepted).not.toHaveBeenCalled();
  expect(u.owner.getSnapshot().entries).toEqual([]);
});

it("Observation stale child and unavailable targets stay local while incident access loss suspends recovery", async () => {
  for (const code of [
    "row_version_conflict",
    "resolved_indicator_not_found",
    "indicator_observation_not_found",
    "incident_not_found",
  ]) {
    const t = observationOwnerFixture(),
      lost = vi.fn();
    t.owner.configure(t.reader, t.transport, lost);
    t.transport.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: {
        kind: "stale_target",
        publicCode: code,
        message: "Review current state",
      },
    });
    const attempt = t.admit({
      action: "resolve",
      observation: testObservation,
      targetId: observationTargetId,
    });
    await t.owner.execute(attempt);
    expect(lost).toHaveBeenCalledTimes(code === "incident_not_found" ? 1 : 0);
    expect(t.owner.getSnapshot().authority !== null).toBe(
      code !== "incident_not_found",
    );
    if (code !== "incident_not_found")
      expect(t.entry()?.phase).toBe("rejected");
  }
});
