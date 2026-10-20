import { afterEach, describe, expect, it, vi } from "vitest";
import {
  auditActorId,
  auditEvent,
  auditPageResult,
} from "../testing/administrativeAuditTestSupport";
import { deferred } from "../testing/fetchMockTestSupport";
import {
  AdministrativeAuditController,
  type AuditPorts,
} from "./administrativeAuditController";
import type { AuditReadResult } from "./api/administrativeAuditClient";

const authority = { actorId: auditActorId, lifetime: "audit-session" };
const controllers: AdministrativeAuditController[] = [];
function setup() {
  const ports = {
    list: vi.fn<AuditPorts["list"]>().mockResolvedValue(auditPageResult()),
    isCurrent: vi.fn<AuditPorts["isCurrent"]>().mockReturnValue(true),
    confirmAccess: vi
      .fn<AuditPorts["confirmAccess"]>()
      .mockResolvedValue({ kind: "authorized" }),
    authorizationFailed: vi.fn<AuditPorts["authorizationFailed"]>(),
  };
  const controller = new AdministrativeAuditController(ports);
  controllers.push(controller);
  controller.setAuthority(authority);
  return { controller, ports };
}
async function settle() {
  for (let i = 0; i < 8; ++i) await Promise.resolve();
}
const unavailable: AuditReadResult = {
  ok: false,
  status: 503,
  error: { code: "internal_error", message: "private server details" },
};

describe("Administrative audit read ownership", () => {
  afterEach(() => {
    for (const controller of controllers.splice(0)) controller.dispose();
    vi.useRealTimers();
  });
  it("separates editable applied and accepted state and retains invalid input", async () => {
    const { controller: c, ports } = setup();
    c.setActive(true);
    await settle();
    c.edit("actor_user_id", "unfinished");
    c.refresh();
    await settle();
    expect(ports.list.mock.calls.at(-1)?.[0].query.actor_user_id).toBe("");
    c.apply();
    expect(c.getSnapshot().fieldErrors.actor_user_id).toBeTruthy();
    expect(c.getSnapshot().announcementRole).toBe("alert");
    expect(c.getSnapshot().page?.rows).toHaveLength(1);
    c.edit("actor_user_id", auditActorId);
    c.apply();
    expect(c.getSnapshot().page).toBeNull();
    await settle();
    expect(c.getSnapshot().applied.actor_user_id).toBe(auditActorId);
    expect(c.getSnapshot().announcementRole).toBe("status");
  });
  it("suppresses stale success error and page publication after replacement", async () => {
    for (const result of [
      auditPageResult([auditEvent({ action_code: "old" })]),
      unavailable,
    ]) {
      const { controller: c, ports } = setup();
      const delayed = deferred<AuditReadResult>();
      ports.list.mockReturnValueOnce(delayed.promise);
      c.setActive(true);
      c.edit("action_code", "password_reset");
      c.apply();
      await settle();
      const state = c.getSnapshot();
      delayed.resolve(result);
      await settle();
      expect(c.getSnapshot()).toBe(state);
    }
    const { controller: c, ports } = setup();
    ports.list.mockResolvedValueOnce(auditPageResult(undefined, "older"));
    c.setActive(true);
    await settle();
    const page = deferred<AuditReadResult>();
    ports.list.mockReturnValueOnce(page.promise);
    c.next();
    c.next();
    expect(ports.list).toHaveBeenCalledTimes(2);
    c.edit("action_code", "password_reset");
    c.apply();
    await settle();
    const accepted = c.getSnapshot();
    page.resolve(auditPageResult([auditEvent({ action_code: "old-page" })]));
    await settle();
    expect(c.getSnapshot()).toBe(accepted);
  });
  it("bounds navigation memory while permitting arbitrarily old pages", async () => {
    const { controller: c, ports } = setup();
    let index = 0;
    ports.list.mockImplementation(async () =>
      auditPageResult(undefined, `cursor-${++index}`),
    );
    c.setActive(true);
    await settle();
    for (let i = 0; i < 25; ++i) {
      c.next();
      await settle();
    }
    expect(c.getSnapshot().page?.descriptor.number).toBe(26);
    expect(c.getSnapshot().page?.rows).toHaveLength(1);
    expect(c.getSnapshot().history).toHaveLength(20);
    c.previous();
    await settle();
    expect(c.getSnapshot().page?.descriptor.number).toBe(25);
    expect(ports.list.mock.calls.at(-1)?.[0].cursor).toBe("cursor-24");
    c.refresh();
    expect(c.getSnapshot().history).toHaveLength(0);
    expect(c.getSnapshot().continuationValid).toBe(false);
    await settle();
    expect(c.getSnapshot().page?.descriptor.number).toBe(1);
    expect(ports.list.mock.calls.at(-1)?.[0].cursor).toBeNull();
  });
  it("requires explicit fresh recovery for rejected or cycling cursors", async () => {
    const { controller: c, ports } = setup();
    ports.list.mockResolvedValueOnce(auditPageResult(undefined, "older"));
    c.setActive(true);
    await settle();
    ports.list.mockResolvedValueOnce({
      ok: false,
      status: 400,
      error: {
        code: "invalid_pagination_request",
        details: { reason_code: "cursor_query_mismatch" },
      },
    });
    c.next();
    await settle();
    expect(c.getSnapshot().problem?.kind).toBe("cursor");
    expect(c.getSnapshot().page?.rows).toHaveLength(1);
    c.retry();
    c.next();
    c.setActive(false);
    c.setActive(true);
    await settle();
    expect(ports.list).toHaveBeenCalledTimes(2);
    expect(ports.confirmAccess).toHaveBeenCalledTimes(1);
    c.refresh();
    await settle();
    expect(ports.list).toHaveBeenCalledTimes(3);
    expect(c.getSnapshot().problem).toBeNull();
    ports.list.mockResolvedValueOnce({
      ok: false,
      status: 400,
      error: { code: "invalid_pagination_request" },
    });
    c.refresh();
    await settle();
    ports.confirmAccess.mockResolvedValueOnce({ kind: "unavailable" });
    const readsBeforeAccess = ports.list.mock.calls.length;
    c.setActive(false);
    c.setActive(true);
    await settle();
    expect(c.getSnapshot().problem?.kind).toBe("access");
    c.setActive(false);
    c.setActive(true);
    await settle();
    expect(c.getSnapshot().problem?.kind).toBe("cursor");
    expect(ports.list).toHaveBeenCalledTimes(readsBeforeAccess);
    ports.confirmAccess.mockResolvedValueOnce({ kind: "access_lost" });
    c.setActive(false);
    c.setActive(true);
    await settle();
    expect(c.getSnapshot().page).toBeNull();
    expect(c.getSnapshot().authority).toBeNull();
    expect(ports.authorizationFailed).toHaveBeenCalledWith(403, authority);
  });
  it("distinguishes empty initial failure and stale refresh with explicit retry", async () => {
    const { controller: c, ports } = setup();
    ports.list.mockResolvedValueOnce(unavailable);
    c.setActive(true);
    expect(c.getSnapshot().page).toBeNull();
    expect(c.getSnapshot().activity).toBe("initial");
    await settle();
    expect(c.getSnapshot().problem?.kind).toBe("initial");
    c.retry();
    await settle();
    c.toggleExpanded(auditEvent().audit_event_id);
    ports.list.mockResolvedValueOnce(unavailable);
    c.refresh();
    await settle();
    expect(c.getSnapshot().problem?.kind).toBe("refresh");
    expect(c.getSnapshot().expandedId).toBe(auditEvent().audit_event_id);
    expect(JSON.stringify(c.getSnapshot())).not.toContain(
      "private server details",
    );
    ports.list.mockResolvedValueOnce(auditPageResult([]));
    c.retry();
    await settle();
    expect(c.getSnapshot().page?.rows).toEqual([]);
    expect(c.getSnapshot().problem).toBeNull();
    expect(c.getSnapshot().expandedId).toBeNull();
  });
  it("fences ignored aborts at the observation deadline", async () => {
    vi.useFakeTimers();
    const { controller: c, ports } = setup();
    const late = deferred<AuditReadResult>();
    ports.list.mockReturnValueOnce(late.promise);
    c.setActive(true);
    await vi.advanceTimersByTimeAsync(30_000);
    expect(c.getSnapshot().activity).toBeNull();
    expect(c.getSnapshot().problem?.message).toContain("timed out");
    c.retry();
    await settle();
    const state = c.getSnapshot();
    late.resolve(unavailable);
    await settle();
    expect(c.getSnapshot()).toBe(state);
  });
  it("identifies a failed page separately and retries its exact continuation", async () => {
    const { controller: c, ports } = setup();
    ports.list.mockResolvedValueOnce(auditPageResult([auditEvent()], "older"));
    c.setActive(true);
    await settle();
    const page = c.getSnapshot().page;
    ports.list.mockResolvedValueOnce(unavailable);
    c.next();
    await settle();
    expect(c.getSnapshot().page).toBe(page);
    expect(c.getSnapshot().problem).toMatchObject({
      kind: "page",
      message: expect.stringContaining("requested audit page"),
    });
    c.retry();
    await settle();
    expect(ports.list.mock.calls.at(-1)?.[0].cursor).toBe("older");
    expect(c.getSnapshot().page?.descriptor.number).toBe(2);
  });
  it("keeps public query rejection on the applied form without marking newer edits invalid", async () => {
    for (const reason of ["invalid_filter_value", "invalid_filter_range"]) {
      const { controller: c, ports } = setup();
      c.setActive(true);
      await settle();
      const rejected = deferred<AuditReadResult>();
      ports.list.mockReturnValueOnce(rejected.promise);
      c.edit("target_kind", "user");
      c.edit("target_id", "applied-target");
      c.apply();
      c.edit("target_id", "newer-draft");
      rejected.resolve({
        ok: false,
        status: 400,
        error: {
          code: "invalid_list_query",
          message: "private diagnostic",
          details: { reason_code: reason },
        },
      });
      await settle();
      expect(c.getSnapshot()).toMatchObject({
        applied: { target_id: "applied-target" },
        inputs: { target_id: "newer-draft" },
        fieldErrors: {},
        problem: { kind: "query" },
      });
      expect(c.getSnapshot().problem?.message).toContain(
        reason === "invalid_filter_range" ? "time range" : "filter",
      );
      expect(JSON.stringify(c.getSnapshot())).not.toContain(
        "private diagnostic",
      );
      c.apply();
      await settle();
      expect(c.getSnapshot().problem).toBeNull();
      expect(c.getSnapshot().applied.target_id).toBe("newer-draft");
    }
  });
  it("pauses hidden reads and confirms access before rereading the retained page", async () => {
    const { controller: c, ports } = setup();
    expect(ports.list).not.toHaveBeenCalled();
    ports.list.mockResolvedValueOnce(auditPageResult(undefined, "older"));
    c.setActive(true);
    await settle();
    c.next();
    await settle();
    c.toggleExpanded(auditEvent().audit_event_id);
    c.edit("target_id", "draft");
    c.rememberScroll(120, 25);
    c.setActive(false);
    c.refresh();
    expect(ports.list).toHaveBeenCalledTimes(2);
    const access = deferred<Awaited<ReturnType<AuditPorts["confirmAccess"]>>>();
    ports.confirmAccess.mockReturnValueOnce(access.promise);
    c.setActive(true);
    expect(ports.list).toHaveBeenCalledTimes(2);
    expect(c.getSnapshot().inputs.target_id).toBe("draft");
    access.resolve({ kind: "authorized" });
    await settle();
    expect(ports.list.mock.calls.at(-1)?.[0].cursor).toBe("older");
    expect(c.getSnapshot().expandedId).toBe(auditEvent().audit_event_id);
    expect(c.getSnapshot().scrollTop).toBe(120);
  });
  it("does not treat unavailable or cancelled access observation as success or denial", async () => {
    const { controller: c, ports } = setup();
    c.setActive(true);
    await settle();
    for (const kind of ["unavailable", "cancelled"] as const) {
      ports.confirmAccess.mockResolvedValueOnce({ kind });
      c.setActive(false);
      c.setActive(true);
      await settle();
      expect(c.getSnapshot().problem?.kind).toBe("access");
      expect(c.getSnapshot().page?.rows).toHaveLength(1);
    }
    expect(ports.authorizationFailed).not.toHaveBeenCalled();
    expect(ports.list).toHaveBeenCalledTimes(1);
  });
  it("clears protected state on retirement and ignores late success and failure", async () => {
    for (const result of [auditPageResult(), unavailable]) {
      const { controller: c, ports } = setup();
      c.setActive(true);
      await settle();
      c.toggleExpanded(auditEvent().audit_event_id);
      c.edit("target_id", "protected draft");
      const late = deferred<AuditReadResult>();
      ports.list.mockReturnValueOnce(late.promise);
      c.refresh();
      c.retire();
      const state = c.getSnapshot();
      late.resolve(result);
      await settle();
      expect(c.getSnapshot()).toBe(state);
      expect(state.page).toBeNull();
      expect(state.expandedId).toBeNull();
      expect(state.inputs.target_id).toBe("");
    }
  });
  it("clears immediately on current session or administrator rejection", async () => {
    for (const [status, code] of [
      [401, "session_required"],
      [403, "authorization_denied"],
    ] as const) {
      const { controller: c, ports } = setup();
      c.setActive(true);
      await settle();
      ports.list.mockResolvedValueOnce({ ok: false, status, error: { code } });
      c.refresh();
      await settle();
      expect(c.getSnapshot().page).toBeNull();
      expect(c.getSnapshot().authority).toBeNull();
      expect(ports.authorizationFailed).toHaveBeenCalledWith(status, authority);
    }
  });
});
