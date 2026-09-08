import { afterEach, describe, expect, it, vi } from "vitest";
import type { AuthorizationRecoveryResult } from "../shared/authorizationRecovery";
import {
  membershipAuthority as authority,
  membershipDeferred as deferred,
  membershipFixture as member,
  membershipPage as page,
} from "../testing/incidentMembershipManagementTestSupport";
import type {
  MembershipListResult,
  MembershipMutationResult,
} from "./api/incidentMembershipManagementClient";
import {
  IncidentMembershipManagementController,
  type MembershipManagementPorts,
} from "./incidentMembershipManagementController";

const controllers: IncidentMembershipManagementController[] = [];
const settle = async () => {
  for (let n = 0; n < 16; ++n) await Promise.resolve();
};
function setup() {
  const list = vi.fn<MembershipManagementPorts["list"]>(async () => page());
  const mutate = vi.fn<MembershipManagementPorts["mutate"]>(async () => ({
    ok: true,
    status: 200,
    resource: member({ role: "reviewer", membership_version: 2 }),
  }));
  const recover = vi.fn<MembershipManagementPorts["recover"]>(async (a) => ({
    kind: "authorized",
    role: a.role,
    userId: a.actorId,
  }));
  const lost = vi.fn();
  const isCurrent = vi.fn(() => true);
  const controller = new IncidentMembershipManagementController({
    list,
    mutate,
    recover,
    lost,
    isCurrent,
  });
  controllers.push(controller);
  controller.setAuthority(authority);
  controller.setActive(true);
  return { controller, list, mutate, recover, lost, isCurrent };
}
async function roleEditor(controller: IncidentMembershipManagementController) {
  await settle();
  controller.openMember(member(), "role");
  controller.editRole("reviewer");
}
afterEach(() => {
  for (const c of controllers.splice(0)) c.dispose();
  vi.useRealTimers();
});
describe("Incident membership management ownership", () => {
  it("bounds live browsing to one page and twenty cursors with explicit first and previous", async () => {
    const { controller, list } = setup();
    await settle();
    list.mockImplementation(async ({ cursor }) =>
      page(
        [member({ display_name: cursor ?? "first" })],
        `cursor-${list.mock.calls.length}`,
      ),
    );
    controller.refresh();
    await settle();
    for (let i = 0; i < 25; ++i) {
      controller.next();
      await settle();
    }
    expect(controller.getSnapshot().page?.rows).toHaveLength(1);
    expect(controller.getSnapshot().history).toHaveLength(20);
    controller.previous();
    await settle();
    expect(controller.getSnapshot().page?.position.number).toBe(25);
    controller.refresh();
    await settle();
    expect(controller.getSnapshot().history).toHaveLength(0);
    expect(controller.getSnapshot().page?.position.number).toBe(1);
    controller.next();
    await settle();
    const secondPosition = controller.getSnapshot().page?.position;
    list.mockResolvedValueOnce({
      ok: false,
      status: 503,
      problem: { code: "service_unavailable" },
    });
    controller.next();
    await settle();
    expect(controller.getSnapshot().read.kind).toBe("failed");
    controller.retryRead();
    await settle();
    expect(controller.getSnapshot().page?.position.number).toBe(3);
    controller.previous();
    await settle();
    expect(controller.getSnapshot().page?.position).toEqual(secondPosition);
  });
  it("coalesces reads and fences delayed responses across close incident and session", async () => {
    const { controller, list } = setup();
    await settle();
    const read = deferred<MembershipListResult>();
    list.mockReturnValue(read.promise);
    controller.refresh();
    controller.refresh();
    await settle();
    expect(list).toHaveBeenCalledTimes(2);
    controller.setActive(false);
    read.resolve(page([member({ display_name: "Obsolete" })]));
    await settle();
    expect(controller.getSnapshot().page).toBeNull();
    controller.setAuthority({ ...authority, lifetime: "replacement" });
    controller.setActive(true);
    await settle();
    controller.setAuthority({ ...authority, incidentId: "other" });
    await settle();
    controller.retire();
    await settle();
    expect(controller.getSnapshot().page).toBeNull();
  });
  it("rejects cursor cycles and retains prior rows after ordinary refresh failure", async () => {
    const { controller, list } = setup();
    await settle();
    list.mockResolvedValue(page([member()], "repeat"));
    controller.refresh();
    await settle();
    controller.next();
    await settle();
    expect(controller.getSnapshot().read.kind).toBe("cursor_rejected");
    list.mockResolvedValue({
      ok: false,
      status: 503,
      problem: { code: "service_unavailable" },
    });
    controller.refresh();
    await settle();
    expect(controller.getSnapshot().read.kind).toBe("failed");
    expect(controller.getSnapshot().page?.rows).toHaveLength(1);
    list.mockResolvedValue(page());
    controller.retryRead();
    await settle();
    expect(controller.getSnapshot().read.kind).toBe("idle");
  });
  it("keeps one draft through reads and close and requires stay or discard to switch", async () => {
    const { controller } = setup();
    await roleEditor(controller);
    controller.refresh();
    await settle();
    expect(controller.getSnapshot().editor?.role).toBe("reviewer");
    controller.openAdd();
    expect(controller.getSnapshot().departure?.kind).toBe("editor");
    controller.resolveDeparture("stay");
    expect(controller.getSnapshot().editor?.kind).toBe("role");
    controller.setActive(false);
    controller.setActive(true);
    await settle();
    expect(controller.getSnapshot().editor?.role).toBe("reviewer");
    controller.openAdd();
    controller.resolveDeparture("discard");
    expect(controller.getSnapshot().editor?.kind).toBe("add");
  });
  it("admits one immutable write before authorization and preserves newer draft revisions", async () => {
    const { controller, recover, mutate } = setup();
    await roleEditor(controller);
    const access = deferred<AuthorizationRecoveryResult>();
    recover.mockReturnValueOnce(access.promise);
    const write = deferred<MembershipMutationResult>();
    mutate.mockReturnValueOnce(write.promise);
    controller.submit();
    controller.submit();
    expect(controller.getSnapshot().operation.kind).toBe("pending");
    expect(mutate).not.toHaveBeenCalled();
    access.resolve({
      kind: "authorized",
      role: "admin",
      userId: authority.actorId,
    });
    await settle();
    controller.editRole("admin");
    controller.submit();
    expect(mutate).toHaveBeenCalledTimes(1);
    expect(mutate.mock.calls[0]?.[0].input.payload).toEqual({
      base_membership_version: 1,
      role: "reviewer",
    });
    write.resolve({
      ok: true,
      status: 200,
      resource: member({ role: "reviewer", membership_version: 2 }),
    });
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(controller.getSnapshot().editor).toMatchObject({
      role: "admin",
      reviewRequired: true,
    });
  });
  it("uses a fresh crypto identity for creates and exact immutable uncertain replay", async () => {
    const { controller, mutate } = setup();
    await settle();
    controller.openAdd();
    controller.editEmail(" analyst@example.test ");
    mutate.mockResolvedValueOnce({
      ok: false,
      status: 503,
      problem: { code: "service_unavailable" },
    });
    controller.submit();
    await settle();
    const first = mutate.mock.calls[0]?.[0].input;
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    controller.editEmail("changed@example.test");
    mutate.mockResolvedValueOnce({ ok: true, status: 200, resource: member() });
    controller.replay();
    await settle();
    expect(mutate.mock.calls[1]?.[0].input).toEqual(first);
    expect(controller.getSnapshot().editor?.email).toBe("changed@example.test");
    mutate.mockResolvedValueOnce({ ok: true, status: 201, resource: member() });
    controller.submit();
    await settle();
    const third = mutate.mock.calls[2]?.[0].input;
    expect(third?.kind).toBe("add");
    if (first?.kind !== "add" || third?.kind !== "add")
      throw new Error("Expected creates");
    expect(third.payload.client_txn_id).not.toBe(first.payload.client_txn_id);
    expect(third.payload.email).toBe("changed@example.test");
  });
  it("retains acknowledgement when follow-up list and authorization fail", async () => {
    const { controller, list, recover, mutate } = setup();
    await roleEditor(controller);
    recover
      .mockResolvedValueOnce({
        kind: "authorized",
        role: "admin",
        userId: authority.actorId,
      })
      .mockResolvedValue({ kind: "unavailable", failure: "transient" });
    list.mockResolvedValue({
      ok: false,
      status: 503,
      problem: { code: "service_unavailable" },
    });
    controller.submit();
    await settle();
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      accessRefresh: "failed",
      listRefresh: "idle",
    });
    recover.mockResolvedValue({
      kind: "authorized",
      role: "admin",
      userId: authority.actorId,
    });
    await controller.recoverAccess();
    await settle();
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      accessRefresh: "done",
      listRefresh: "failed",
    });
    expect(mutate).toHaveBeenCalledTimes(1);
  });
  it("distinguishes version conflicts from uncertain writes and requires observed review before another attempt", async () => {
    const { controller, mutate, list } = setup();
    await roleEditor(controller);
    mutate.mockResolvedValueOnce({
      ok: false,
      status: 409,
      problem: { code: "membership_version_conflict" },
    });
    controller.submit();
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("conflicted");
    controller.submit();
    expect(mutate).toHaveBeenCalledTimes(1);
    list.mockResolvedValue(
      page([member({ role: "admin", membership_version: 3 })]),
    );
    controller.observeCurrent();
    await settle();
    expect(controller.getSnapshot().editor?.base?.membership_version).toBe(1);
    controller.reviewObserved();
    expect(controller.getSnapshot().editor?.base?.membership_version).toBe(3);
    expect(mutate).toHaveBeenCalledTimes(1);
    mutate.mockResolvedValueOnce({
      ok: false,
      status: 500,
      problem: { code: "internal_error" },
    });
    controller.submit();
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    list.mockResolvedValue(
      page([member({ role: "reviewer", membership_version: 4 })]),
    );
    controller.observeCurrent();
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    controller.setActive(false);
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "uncertain",
      observed: null,
    });
    controller.reviewObserved();
    expect(controller.getSnapshot().editor?.base?.membership_version).toBe(3);
    controller.setActive(true);
    await settle();
    controller.reviewObserved();
    expect(mutate).toHaveBeenCalledTimes(2);
  });
  it("does not treat absence as a DELETE receipt and handles a real 204 without content", async () => {
    const { controller, mutate, list } = setup();
    await settle();
    controller.openMember(member(), "remove");
    mutate.mockResolvedValueOnce({
      ok: false,
      status: 502,
      problem: { code: "invalid_public_contract_response" },
    });
    controller.submit();
    await settle();
    list.mockResolvedValue(page([]));
    controller.observeCurrent();
    await settle();
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "uncertain",
      observed: null,
    });
    controller.reviewObserved();
    expect(controller.canSubmit()).toBe(false);
    const freshRead = deferred<MembershipListResult>();
    list.mockReturnValueOnce(freshRead.promise);
    controller.forgetRecovery();
    expect(controller.getSnapshot().page).toBeNull();
    controller.openMember(member(), "remove");
    expect(controller.getSnapshot().editor).toBeNull();
    await settle();
    const freshMember = member({ membership_version: 2 });
    freshRead.resolve(page([freshMember]));
    await settle();
    controller.openMember(freshMember, "remove");
    expect(controller.getSnapshot().editor?.base?.membership_version).toBe(2);
    mutate.mockResolvedValueOnce({ ok: true, status: 204, resource: null });
    controller.submit();
    await settle();
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      status: 204,
      receipt: null,
    });
  });
  it("invalidates removal confirmation when the observed subject version changes", async () => {
    const { controller, list, mutate } = setup();
    await settle();
    controller.openMember(member(), "remove");
    list.mockResolvedValue(page([member({ membership_version: 2 })]));
    controller.refresh();
    await settle();
    expect(controller.canSubmit()).toBe(false);
    controller.submit();
    expect(mutate).not.toHaveBeenCalled();
    controller.reviewObserved();
    expect(controller.canSubmit()).toBe(true);
  });
  it("does not add idempotency to no-op PATCH or reinterpret domain rejections", async () => {
    const { controller, mutate } = setup();
    await settle();
    controller.openMember(member(), "role");
    mutate.mockResolvedValueOnce({ ok: true, status: 200, resource: member() });
    controller.submit();
    await settle();
    expect(mutate.mock.calls[0]?.[0].input.payload).toEqual({
      base_membership_version: 1,
      role: "viewer",
    });
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      receipt: { membership_version: 1 },
    });
    for (const code of [
      "last_incident_admin",
      "membership_not_found",
      "user_inactive",
      "user_not_found",
      "membership_exists_use_patch",
    ] as const) {
      controller.openMember(member(), "role");
      mutate.mockResolvedValueOnce({
        ok: false,
        status: code.endsWith("not_found") ? 404 : 409,
        problem: { code },
      });
      controller.submit();
      await settle();
      expect(controller.getSnapshot().operation).toMatchObject({
        kind: "rejected",
        problem: { code },
      });
    }
  });
  it("bounds observation while an ignored abort keeps transport admission locked", async () => {
    vi.useFakeTimers();
    const { controller, mutate } = setup();
    await roleEditor(controller);
    const pending = deferred<MembershipMutationResult>();
    mutate.mockReturnValueOnce(pending.promise);
    controller.submit();
    await settle();
    await vi.advanceTimersByTimeAsync(30_001);
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    expect(controller.getSnapshot().transportPending).toBe(true);
    controller.forgetRecovery();
    controller.submit();
    expect(mutate).toHaveBeenCalledTimes(1);
    pending.resolve({
      ok: true,
      status: 200,
      resource: member({ role: "reviewer" }),
    });
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    expect(controller.getSnapshot().transportPending).toBe(false);
  });
  it("allows explicit route departure without claiming cancellation and fences late success", async () => {
    const { controller, mutate } = setup();
    await roleEditor(controller);
    const pending = deferred<MembershipMutationResult>();
    mutate.mockReturnValueOnce(pending.promise);
    controller.submit();
    await settle();
    const stay = controller.requestLeave();
    controller.resolveDeparture("stay");
    expect(await stay).toBe(false);
    const leave = controller.requestLeave();
    controller.resolveDeparture("discard");
    expect(await leave).toBe(true);
    controller.setAuthority({ ...authority, lifetime: "new-session" });
    controller.setActive(true);
    await settle();
    pending.resolve({
      ok: true,
      status: 200,
      resource: member({ role: "reviewer" }),
    });
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("idle");
    expect(controller.getSnapshot().editor).toBeNull();
  });
  it("retains pending and uncertain writes across ordinary close without background list publication", async () => {
    const { controller, mutate, list } = setup();
    await roleEditor(controller);
    const pending = deferred<MembershipMutationResult>();
    mutate.mockReturnValueOnce(pending.promise);
    controller.submit();
    await settle();
    const count = list.mock.calls.length;
    controller.setActive(false);
    pending.resolve({
      ok: true,
      status: 200,
      resource: member({ role: "reviewer", membership_version: 2 }),
    });
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(controller.getSnapshot().page).toBeNull();
    expect(list).toHaveBeenCalledTimes(count);
  });
  it("retires privileged drafts on downgrade while nonadmin membership reads remain available", async () => {
    const { controller, mutate, recover } = setup();
    await roleEditor(controller);
    controller.setAuthority({ ...authority, role: "viewer" });
    recover.mockResolvedValue({
      kind: "authorized",
      role: "viewer",
      userId: authority.actorId,
    });
    controller.refresh();
    await settle();
    expect(controller.getSnapshot().editor).toBeNull();
    expect(controller.getSnapshot().page?.rows).toHaveLength(1);
    controller.openAdd();
    controller.submit();
    expect(mutate).not.toHaveBeenCalled();
  });
  it("distinguishes temporary authority failure incident loss and session loss", async () => {
    for (const kind of [
      "unavailable",
      "access_lost",
      "session_lost",
    ] as const) {
      const { controller, recover, lost } = setup();
      await roleEditor(controller);
      recover.mockResolvedValue(
        kind === "unavailable" ? { kind, failure: "transient" } : { kind },
      );
      controller.refresh();
      await settle();
      expect(controller.getSnapshot().page).toBeNull();
      if (kind === "unavailable") {
        expect(lost).not.toHaveBeenCalled();
        expect(controller.getSnapshot().editor?.role).toBe("reviewer");
        expect(controller.canSubmit()).toBe(false);
      } else {
        expect(lost).toHaveBeenCalledWith(
          kind === "access_lost" ? "incident" : "session",
          authority,
        );
        expect(controller.getSnapshot().editor).toBeNull();
      }
    }
  });
});
