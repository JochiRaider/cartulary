import { afterEach, describe, expect, it, vi } from "vitest";
import {
  metadataDeferred as deferred,
  metadataIncident as incident,
  metadataActorId,
  metadataIncidentId,
} from "../testing/incidentMetadataTestSupport";
import type { LifecycleResult } from "./api/incidentLifecycleClient";
import {
  IncidentLifecycleController,
  type IncidentLifecyclePorts,
} from "./incidentLifecycleController";
import type { LifecycleAuthority } from "./incidentLifecycleModel";

const authority: LifecycleAuthority = {
  incidentId: metadataIncidentId,
  actorId: metadataActorId,
  lifetime: "session",
  role: "admin",
};
const closed = () =>
  incident({
    status: "closed",
    closed_at: "2026-08-02T00:00:00Z",
    incident_version: 2,
  });
const flush = async () => {
  for (let n = 0; n < 30; ++n) await Promise.resolve();
};
const controllers: IncidentLifecycleController[] = [];
function setup() {
  const read = vi.fn<IncidentLifecyclePorts["read"]>(async () => ({
    ok: true,
    resource: incident(),
  }));
  const mutate = vi.fn<IncidentLifecyclePorts["mutate"]>(async () => ({
    ok: true,
    resource: closed(),
  }));
  const recover = vi.fn<IncidentLifecyclePorts["recover"]>(async (a) => ({
    kind: "authorized",
    role: a.role,
    userId: a.actorId,
  }));
  const lost = vi.fn();
  const publishResource = vi.fn();
  const isCurrent = vi.fn(() => true);
  const controller = new IncidentLifecycleController({
    read,
    mutate,
    recover,
    lost,
    publishResource,
    isCurrent,
  });
  controllers.push(controller);
  controller.setAuthority(authority);
  controller.setActive(true);
  return {
    controller,
    read,
    mutate,
    recover,
    lost,
    publishResource,
    isCurrent,
  };
}
function close(
  controller: IncidentLifecycleController,
  reason = "  e\u0301\r\nreason  ",
) {
  controller.changeReason(reason);
  controller.propose("close");
  controller.confirm();
}
afterEach(() => {
  for (const c of controllers.splice(0)) c.dispose();
  vi.useRealTimers();
  vi.restoreAllMocks();
});
describe("Incident lifecycle ownership", () => {
  it("locks same-tick admission before authorization and uses Web Crypto", async () => {
    const { controller, recover, mutate } = setup();
    await flush();
    const access =
      deferred<Awaited<ReturnType<IncidentLifecyclePorts["recover"]>>>();
    recover.mockReturnValue(access.promise);
    close(controller);
    controller.confirm();
    controller.propose("close");
    controller.confirm();
    expect(controller.getSnapshot().operation.kind).toBe("pending");
    expect(mutate).not.toHaveBeenCalled();
    access.resolve({
      kind: "authorized",
      role: "admin",
      userId: metadataActorId,
    });
    await flush();
    expect(mutate).toHaveBeenCalledTimes(1);
    const attempt = mutate.mock.calls[0]?.[0].attempt;
    expect(attempt?.payload.reason).toBe("  e\u0301\r\nreason  ");
    expect(attempt?.payload.client_txn_id).toMatch(/^incident-close-/u);
    expect(Object.isFrozen(attempt)).toBe(true);
    expect(Object.isFrozen(attempt?.payload)).toBe(true);
  });
  it("requires current reviewed intent and version for a fresh action", async () => {
    const { controller, mutate } = setup();
    await flush();
    controller.changeReason("reason");
    controller.propose("close");
    controller.changeReason("newer");
    controller.confirm();
    expect(mutate).not.toHaveBeenCalled();
    controller.review();
    controller.acceptResource(incident({ incident_version: 2 }));
    controller.confirm();
    expect(mutate).not.toHaveBeenCalled();
    expect(controller.getSnapshot().review).toBeNull();
    controller.review();
    controller.confirm();
    await flush();
    expect(mutate).toHaveBeenCalledTimes(1);
    expect(
      mutate.mock.calls[0]?.[0].attempt.payload.base_incident_version,
    ).toBe(2);
  });
  it("replays the exact route actor lifetime key version and raw payload despite newer form input", async () => {
    const { controller, mutate } = setup();
    await flush();
    mutate.mockRejectedValueOnce(new TypeError("lost"));
    close(controller);
    await flush();
    const original = mutate.mock.calls[0]?.[0].attempt;
    controller.changeReason("a different new action");
    controller.acceptResource(closed());
    controller.propose("reopen");
    expect(mutate).toHaveBeenCalledTimes(1);
    controller.replay();
    await flush();
    expect(mutate.mock.calls[1]?.[0].attempt).toBe(original);
    expect(controller.getSnapshot().draft.reason).toBe(
      "a different new action",
    );
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
  });
  it("never publishes a replay receipt as current even without a newer local version", async () => {
    const { controller, mutate, read, publishResource } = setup();
    await flush();
    mutate.mockRejectedValueOnce(new Error("lost"));
    close(controller);
    await flush();
    publishResource.mockClear();
    read.mockResolvedValue({
      ok: false,
      status: 500,
      problem: { code: "internal_error" },
    });
    controller.replay();
    await flush();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(controller.getSnapshot().resource?.incident_version).toBe(1);
    expect(controller.getSnapshot().read).toBe("failed");
    expect(publishResource).not.toHaveBeenCalled();
    expect(controller.canReplay()).toBe(false);
  });
  it("keeps newer reopened state while acknowledging the original close", async () => {
    const { controller, mutate, read, publishResource } = setup();
    await flush();
    mutate.mockRejectedValueOnce(new Error("lost"));
    close(controller);
    await flush();
    const latest = incident({ incident_version: 3 });
    controller.acceptResource(latest);
    read.mockResolvedValue({ ok: true, resource: latest });
    publishResource.mockClear();
    controller.replay();
    await flush();
    expect(controller.getSnapshot().resource).toEqual(latest);
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      replay: true,
      receipt: closed(),
    });
    expect(publishResource).not.toHaveBeenCalled();
  });
  it("preserves newer reason input after delayed confirmation", async () => {
    const { controller, mutate } = setup();
    await flush();
    const response = deferred<LifecycleResult>();
    mutate.mockReturnValue(response.promise);
    close(controller);
    await flush();
    controller.changeReason("  newer\nreason  ");
    response.resolve({ ok: true, resource: closed() });
    await flush();
    expect(controller.getSnapshot().draft.reason).toBe("  newer\nreason  ");
  });
  it("retains confirmation when subsequent resource and authorization reads fail", async () => {
    const { controller, mutate, read, recover } = setup();
    await flush();
    read.mockResolvedValue({
      ok: false,
      status: 500,
      problem: { code: "internal_error" },
    });
    close(controller);
    await flush();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(controller.getSnapshot().read).toBe("failed");
    recover.mockResolvedValue({ kind: "unavailable", failure: "transient" });
    controller.refresh();
    await flush();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(controller.getSnapshot().access).toBe("unavailable");
    controller.replay();
    expect(mutate).toHaveBeenCalledTimes(1);
  });
  it("holds the unsettled transport lock beyond timeout and accepts a late current receipt", async () => {
    vi.useFakeTimers();
    const { controller, mutate } = setup();
    await flush();
    const response = deferred<LifecycleResult>();
    mutate.mockReturnValue(response.promise);
    close(controller);
    await flush();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(controller.getSnapshot()).toMatchObject({
      transportPending: true,
      operation: { kind: "uncertain" },
    });
    controller.replay();
    controller.forgetRecovery();
    close(controller);
    expect(mutate).toHaveBeenCalledTimes(1);
    response.resolve({ ok: true, resource: closed() });
    await flush();
    expect(controller.getSnapshot()).toMatchObject({
      transportPending: false,
      operation: { kind: "confirmed" },
    });
  });
  it("retains settled transport loss until explicit replay and never infers receipt from reads", async () => {
    const { controller, mutate, read } = setup();
    await flush();
    mutate.mockRejectedValue(new Error("lost"));
    close(controller);
    await flush();
    read.mockResolvedValue({ ok: true, resource: closed() });
    controller.refresh();
    await flush();
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    expect(mutate).toHaveBeenCalledTimes(1);
    expect(controller.canReplay()).toBe(true);
  });
  it("requires renewed review after version rejection and allocates a fresh key", async () => {
    const { controller, mutate, read } = setup();
    await flush();
    mutate.mockResolvedValueOnce({
      ok: false,
      status: 409,
      problem: { code: "incident_version_conflict" },
    });
    read.mockResolvedValue({
      ok: true,
      resource: incident({ incident_version: 2 }),
    });
    close(controller);
    await flush();
    expect(controller.getSnapshot().review).toBeNull();
    controller.confirm();
    expect(mutate).toHaveBeenCalledTimes(1);
    controller.review();
    controller.confirm();
    await flush();
    expect(
      mutate.mock.calls[1]?.[0].attempt.payload.base_incident_version,
    ).toBe(2);
    expect(mutate.mock.calls[1]?.[0].attempt.payload.client_txn_id).not.toBe(
      mutate.mock.calls[0]?.[0].attempt.payload.client_txn_id,
    );
  });
  it("keeps illegal transition and transaction conflict distinct without automatic replacement", async () => {
    for (const code of ["illegal_transition", "client_txn_conflict"] as const) {
      const { controller, mutate } = setup();
      await flush();
      mutate.mockResolvedValue({
        ok: false,
        status: 409,
        problem: { code, reason: "incident_already_closed" },
      });
      close(controller);
      await flush();
      expect(controller.getSnapshot().operation).toMatchObject({
        kind: "rejected",
        problem: { code },
      });
      expect(mutate).toHaveBeenCalledTimes(1);
      if (code === "client_txn_conflict") {
        controller.propose("close");
        expect(controller.getSnapshot().review).toBeNull();
        controller.forgetRecovery();
        await flush();
        expect(controller.getSnapshot().operation.kind).toBe("idle");
      }
    }
  });
  it("associates validation only with its submitted revision", async () => {
    const { controller, mutate } = setup();
    await flush();
    const response = deferred<LifecycleResult>();
    mutate.mockReturnValue(response.promise);
    close(controller);
    await flush();
    controller.changeReason("corrected");
    response.resolve({
      ok: false,
      status: 400,
      problem: {
        code: "invalid_incident_lifecycle_request",
        field: "reason",
        reason: "reason_too_long",
      },
    });
    await flush();
    expect(controller.getSnapshot().fieldError).toBeNull();
    expect(controller.getSnapshot().draft.reason).toBe("corrected");
  });
  it("does not treat rejected replay as proof of the original action failing", async () => {
    const { controller, mutate } = setup();
    await flush();
    mutate.mockRejectedValueOnce(new Error("lost"));
    close(controller);
    await flush();
    mutate.mockResolvedValue({
      ok: false,
      status: 403,
      problem: { code: "authorization_denied" },
    });
    controller.replay();
    await flush();
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "uncertain",
      problem: { code: "authorization_denied" },
    });
  });
  it("invalidates preflight on a current role change despite obsolete administrator recovery", async () => {
    const { controller, recover, mutate } = setup();
    await flush();
    const stale =
      deferred<Awaited<ReturnType<IncidentLifecyclePorts["recover"]>>>();
    recover.mockReturnValueOnce(stale.promise);
    close(controller);
    controller.setAuthority({ ...authority, role: "viewer" });
    stale.resolve({
      kind: "authorized",
      userId: metadataActorId,
      role: "admin",
    });
    await flush();
    expect(mutate).not.toHaveBeenCalled();
    expect(controller.getSnapshot().authority?.role).toBe("viewer");
    expect(controller.getSnapshot().draft.reason).not.toBe("");
    expect(controller.canConfirm()).toBe(false);
  });
  it("invalidates reviewed action after intent changes and after document hiding", async () => {
    const { controller, mutate, recover } = setup();
    await flush();
    controller.changeReason("reason");
    controller.propose("close");
    controller.cancelReview();
    controller.confirm();
    expect(mutate).not.toHaveBeenCalled();
    controller.propose("close");
    controller.setActive(false);
    controller.setActive(true);
    await flush();
    controller.confirm();
    expect(mutate).not.toHaveBeenCalled();
    expect(controller.getSnapshot().draft.reason).toBe("reason");
    for (const replay of [false, true]) {
      if (replay) {
        mutate.mockRejectedValueOnce(new Error("lost response"));
        close(controller, "reason");
        await flush();
      }
      const calls = mutate.mock.calls.length;
      const stale =
        deferred<Awaited<ReturnType<IncidentLifecyclePorts["recover"]>>>();
      recover.mockReturnValueOnce(stale.promise);
      if (replay) controller.replay();
      else close(controller, "reason");
      controller.setActive(false);
      controller.setActive(true);
      stale.resolve({
        kind: "authorized",
        userId: metadataActorId,
        role: "admin",
      });
      await flush();
      expect(mutate).toHaveBeenCalledTimes(calls);
      expect(controller.getSnapshot().operation.kind).toBe(
        replay ? "uncertain" : "idle",
      );
      expect(controller.getSnapshot().draft.reason).toBe("reason");
    }
  });
  it("keeps local recovery on demotion but rejects fresh and replay dispatch", async () => {
    const { controller, mutate, recover } = setup();
    await flush();
    mutate.mockRejectedValueOnce(new Error("lost"));
    close(controller);
    await flush();
    controller.setAuthority({ ...authority, role: "viewer" });
    recover.mockResolvedValue({
      kind: "authorized",
      role: "viewer",
      userId: metadataActorId,
    });
    controller.refresh();
    await flush();
    controller.replay();
    controller.propose("close");
    expect(mutate).toHaveBeenCalledTimes(1);
    expect(controller.getSnapshot().draft.reason).not.toBe("");
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
  });
  it("rechecks current admin authority before dispatch without treating preflight failure as a sent action", async () => {
    const { controller, mutate, recover } = setup();
    await flush();
    recover.mockResolvedValue({
      kind: "authorized",
      role: "viewer",
      userId: metadataActorId,
    });
    close(controller);
    await flush();
    expect(mutate).not.toHaveBeenCalled();
    expect(controller.getSnapshot().operation.kind).toBe("idle");
    expect(controller.getSnapshot().draft.reason).not.toBe("");
  });
  it("retains work across drawer hiding and revalidates on return without hidden dispatch", async () => {
    const { controller, mutate, read } = setup();
    await flush();
    controller.changeReason("retained");
    controller.propose("close");
    controller.setActive(false);
    controller.confirm();
    expect(mutate).not.toHaveBeenCalled();
    expect(controller.getSnapshot().draft.reason).toBe("retained");
    controller.setActive(true);
    await flush();
    expect(read).toHaveBeenCalledTimes(2);
    expect(controller.getSnapshot().review).toBeNull();
  });
  it("settles a dispatched action while hidden without starting follow-up reads", async () => {
    const { controller, mutate, read } = setup();
    await flush();
    const response = deferred<LifecycleResult>();
    mutate.mockReturnValue(response.promise);
    close(controller);
    await flush();
    controller.setActive(false);
    response.resolve({ ok: true, resource: closed() });
    await flush();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(read).toHaveBeenCalledTimes(1);
  });
  it("fences ignored-abort completion after incident actor and lifetime changes", async () => {
    for (const change of [
      { incidentId: metadataActorId },
      { actorId: metadataIncidentId },
      { lifetime: "new-session" },
    ]) {
      const { controller, mutate, publishResource } = setup();
      await flush();
      const response = deferred<LifecycleResult>();
      mutate.mockReturnValue(response.promise);
      close(controller);
      await flush();
      controller.setAuthority({ ...authority, ...change });
      await flush();
      publishResource.mockClear();
      response.resolve({ ok: true, resource: closed() });
      await flush();
      expect(controller.getSnapshot().operation.kind).toBe("idle");
      expect(controller.getSnapshot().draft.reason).toBe("");
      expect(publishResource).not.toHaveBeenCalled();
    }
  });
  it("clears protected work on explicit session or incident access loss", async () => {
    for (const kind of ["session_lost", "access_lost"] as const) {
      const { controller, recover, lost } = setup();
      await flush();
      controller.changeReason("protected");
      recover.mockResolvedValue({ kind });
      controller.refresh();
      await flush();
      expect(controller.getSnapshot().authority).toBeNull();
      expect(controller.getSnapshot().draft.reason).toBe("");
      expect(lost).toHaveBeenCalledTimes(1);
    }
  });
  it("rejects invalid resource state and conflicting same-version publication", async () => {
    const { controller } = setup();
    await flush();
    expect(
      controller.acceptResource(incident({ incident_id: metadataActorId })),
    ).toBe(false);
    expect(
      controller.acceptResource(
        incident({ status: "closed", closed_at: null }),
      ),
    ).toBe(false);
    expect(
      controller.acceptResource(
        incident({ title: "conflicting same version" }),
      ),
    ).toBe(false);
    expect(controller.getSnapshot().resource).toEqual(incident());
  });
  it("separates reason discard and departure forgetting from server cancellation", async () => {
    const { controller, mutate } = setup();
    await flush();
    const response = deferred<LifecycleResult>();
    mutate.mockReturnValue(response.promise);
    close(controller);
    await flush();
    controller.discardReason();
    expect(controller.getSnapshot().operation.kind).toBe("pending");
    expect(controller.hasDepartureWork()).toBe(true);
    const stay = controller.requestLeave();
    controller.resolveDeparture("stay");
    expect(await stay).toBe(false);
    const leave = controller.requestLeave();
    controller.resolveDeparture("discard");
    expect(await leave).toBe(true);
    expect(controller.getSnapshot().transportPending).toBe(true);
    response.resolve({ ok: true, resource: closed() });
    await flush();
    expect(controller.getSnapshot().operation.kind).toBe("idle");
  });
});
