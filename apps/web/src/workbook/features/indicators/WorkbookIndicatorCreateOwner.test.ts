import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import {
  canonicalCreateContract,
  canonicalCreateFixture,
  canonicalCreateValues,
} from "../../../testing/indicatorCreateTestSupport";
import {
  observationAuthority,
  observationOwnerFixture,
  observationTargetId,
  testObservation,
} from "../../../testing/observationTestSupport";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { WorkbookMutationRuntimeRegistry } from "../../runtime/WorkbookMutationRuntimeRegistry";
import type { IndicatorCreateOutcome } from "./indicatorCreateOperation";

afterEach(() => vi.useRealTimers());
it("Canonical create reserves admission synchronously and rejects invalid or changed observations before dispatch", async () => {
  const t = canonicalCreateFixture();
  expect(
    t.owner.admit(testObservation, canonicalCreateContract, {}, t.binding).kind,
  ).toBe("invalid");
  expect(
    t.owner.admit(
      { ...testObservation, resolution_status: "dismissed" },
      canonicalCreateContract,
      canonicalCreateValues,
      t.binding,
    ).kind,
  ).toBe("unavailable");
  const attempt = t.admit();
  expect(
    t.owner.admit(
      testObservation,
      canonicalCreateContract,
      canonicalCreateValues,
      t.binding,
    ).kind,
  ).toBe("unavailable");
  t.reader.observations.mockResolvedValue({
    kind: "accepted",
    value: {
      items: [{ ...testObservation, row_version: 2 }],
      hasMore: false,
      nextCursor: null,
    },
  });
  await t.owner.execute(attempt);
  expect(t.transport.send).not.toHaveBeenCalled();
  expect(t.entry()?.phase).toBe("rejected");
  expect(t.ids.create).toHaveBeenCalledTimes(1);
  t.reader.observations.mockResolvedValue({
    kind: "accepted",
    value: { items: [testObservation], hasMore: false, nextCursor: null },
  });
  await t.owner.execute(t.admit());
  expect(t.transport.send).toHaveBeenCalledTimes(1);
  expect(
    t.owner.admit(
      testObservation,
      canonicalCreateContract,
      canonicalCreateValues,
      t.binding,
    ).kind,
  ).toBe("unavailable");
});
it("Canonical uncertainty replays original bytes independently of drafts panel closure and incident closure", async () => {
  const t = canonicalCreateFixture();
  t.owner.drafts.ensure(testObservation);
  t.transport.send.mockResolvedValueOnce({ kind: "uncertain" });
  const attempt = t.admit();
  await t.owner.execute(attempt);
  t.binding.isCurrent.mockReturnValue(false);
  t.owner.drafts.update(
    testObservation.observation_id,
    "indicator.display_value",
    "other.example",
  );
  t.owner.closeIncident();
  await t.owner.replay(attempt.id);
  expect(t.transport.send.mock.calls.map(([value]) => value)).toEqual([
    attempt,
    attempt,
  ]);
  expect(JSON.parse(attempt.body)["indicator.display_value"]).toBe(
    "EXAMPLE[.]COM",
  );
  expect(attempt.observation).toEqual(testObservation);
  expect(Object.isFrozen(attempt.observation)).toBe(true);
  expect(t.ids.create).toHaveBeenCalledTimes(1);
  expect(t.entry()?.receipt).toEqual(t.receipt);
});
it("Canonical acceptance precedes refresh and read retry preserves newer materialization without another create", async () => {
  const t = canonicalCreateFixture();
  t.reconcile.mockImplementation(async () => {
    expect(t.entry()?.receipt).not.toBeNull();
    throw new Error("refresh failed");
  });
  t.owner.acceptVersion(observationTargetId, 99);
  const canonical = t.admit();
  await t.owner.execute(canonical);
  expect(t.entry()?.phase).toBe("accepted");
  expect(t.entry()?.refresh).toBe("required");
  expect(t.owner.latestVersion(observationTargetId)).toBe(99);
  expect(t.owner.acceptRow(t.receipt.row)).toBeNull();
  t.reconcile.mockResolvedValue();
  await t.owner.refresh(canonical.id);
  expect(t.entry()?.refresh).toBe("complete");
  expect(t.transport.send).toHaveBeenCalledTimes(1);
  expect(t.accepted).toHaveBeenCalledTimes(1);
});
it("Canonical late acceptance survives panel disposal but cannot attach to a retired account", async () => {
  vi.useFakeTimers();
  const t = canonicalCreateFixture(),
    response = deferred<IndicatorCreateOutcome>();
  t.transport.send.mockReturnValue(response.promise);
  const attempt = t.admit(),
    run = t.owner.execute(attempt);
  await vi.advanceTimersByTimeAsync(30_001);
  await run;
  expect(t.entry()?.phase).toBe("uncertain");
  expect(t.entry()?.transportPending).toBe(true);
  t.binding.isCurrent.mockReturnValue(false);
  await t.owner.replay(attempt.id);
  expect(t.transport.send).toHaveBeenCalledTimes(1);
  response.resolve({ kind: "accepted", receipt: t.receipt });
  await vi.advanceTimersByTimeAsync(0);
  expect(t.entry()?.phase).toBe("accepted");
  const other = canonicalCreateFixture(),
    late = deferred<IndicatorCreateOutcome>();
  other.transport.send.mockReturnValue(late.promise);
  const pending = other.owner.execute(other.admit());
  await vi.advanceTimersByTimeAsync(0);
  other.owner.setAuthority({
    ...observationAuthority,
    actorId: observationTargetId,
  });
  late.resolve({ kind: "accepted", receipt: other.receipt });
  await pending;
  expect(other.owner.getSnapshot().entries).toEqual([]);
  expect(other.accepted).not.toHaveBeenCalled();
});
it("Canonical authority gates replay and runtime lifetime retains uncertainty without browser persistence", async () => {
  const t = canonicalCreateFixture(),
    registry = new WorkbookMutationRuntimeRegistry(),
    scope = {
      incidentId: observationAuthority.incidentId,
      clientInstanceId: "tab",
    };
  const runtime = registry.acquire(
      scope,
      () => new WorkbookMutationRuntime(scope, t.ids, { execute: vi.fn() }),
    ),
    owner = runtime.indicatorCreate;
  owner.configure(t.reader, t.transport);
  owner.setAuthority(observationAuthority);
  owner.drafts.ensure(testObservation);
  t.transport.send.mockResolvedValueOnce({ kind: "uncertain" });
  const admission = owner.admit(
    testObservation,
    canonicalCreateContract,
    canonicalCreateValues,
    t.binding,
  );
  if (admission.kind !== "admitted") throw new Error("admission");
  await owner.execute(admission.attempt);
  expect(runtime.getSnapshot().primaryLabel).toBe("Conflict");
  registry.sessionUnavailable();
  expect(owner.getSnapshot().entries).toEqual([]);
  owner.setAuthority({ ...observationAuthority, role: "viewer" });
  await owner.replay(admission.attempt.id);
  expect(t.transport.send).toHaveBeenCalledTimes(1);
  owner.setAuthority(observationAuthority);
  t.transport.send.mockResolvedValueOnce({
    kind: "rejected",
    failure: {
      kind: "authorization_lost",
      publicCode: "authorization_denied",
      message: "Denied",
    },
  });
  await owner.replay(admission.attempt.id);
  owner.setAuthority(observationAuthority);
  expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
  registry.replaceAccount();
  expect(owner.drafts.get(testObservation.observation_id)).toBeUndefined();
  registry.dispose();
});
it("Canonical results associate explicit observation attempts and retain independent resolution recovery", async () => {
  const t = canonicalCreateFixture(),
    observations = observationOwnerFixture();
  const canonical = t.admit();
  await t.owner.execute(canonical);
  const attempt = observations.owner.admit(
    {
      action: "resolve",
      observation: testObservation,
      targetId: observationTargetId,
    },
    observations.binding,
  );
  if (!attempt) throw new Error("observation admission");
  t.owner.associateResolution(canonical.id, attempt);
  observations.transport.send.mockResolvedValueOnce({ kind: "uncertain" });
  await observations.owner.execute(attempt);
  expect(t.entry()?.resolutionAttemptIds).toEqual([attempt.id]);
  expect(t.entry()?.attempt.id).not.toBe(attempt.id);
  expect(t.entry()?.receipt).toEqual(t.receipt);
  await observations.owner.replay(attempt.id);
  expect(observations.transport.send).toHaveBeenCalledTimes(2);
  expect(t.transport.send).toHaveBeenCalledTimes(1);
});
