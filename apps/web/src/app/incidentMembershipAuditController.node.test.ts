import { afterEach, describe, expect, it, vi } from "vitest";
import { auditEvent } from "../testing/administrativeAuditTestSupport";
import type { MembershipAuditResult } from "./api/incidentMembershipAuditClient";
import {
  IncidentMembershipAuditController,
  type MembershipAuditPorts,
} from "./incidentMembershipAuditController";
import {
  emptyMembershipAuditQuery,
  type MembershipAuditAuthority,
  membershipAuditDeadline,
} from "./incidentMembershipAuditModel";

const authority: MembershipAuditAuthority = {
  lifetime: "session-1",
  actorId: "actor-1",
  incidentId: "incident-1",
};
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
function page(
  id = "event-1",
  cursor: string | null = "next",
): MembershipAuditResult {
  return {
    ok: true,
    rows: [
      auditEvent({
        audit_event_id: id,
        scope_kind: "incident",
        scope_id: authority.incidentId,
      }),
    ],
    paging: cursor
      ? { limit: 100, has_more: true, next_cursor: cursor }
      : { limit: 100, has_more: false, next_cursor: null },
  };
}
const failure = (status: number, code: string): MembershipAuditResult => ({
  ok: false,
  status,
  error: { code },
});
const controllers: IncidentMembershipAuditController[] = [];
function setup(overrides: Partial<MembershipAuditPorts> = {}) {
  const reads: ReturnType<typeof deferred<MembershipAuditResult>>[] = [];
  const list = vi.fn<MembershipAuditPorts["list"]>(() => {
    const pending = deferred<MembershipAuditResult>();
    reads.push(pending);
    return pending.promise;
  });
  const lost = vi.fn();
  const recover = vi.fn<MembershipAuditPorts["recover"]>(async (subject) => ({
    kind: "authorized",
    role: "admin",
    userId: subject.actorId,
  }));
  const controller = new IncidentMembershipAuditController({
    list,
    lost,
    recover,
    isCurrent: () => true,
    ...overrides,
  });
  controllers.push(controller);
  controller.setAuthority(authority);
  controller.setActive(true);
  return { controller, list, reads, lost, recover };
}
const settle = async () => {
  for (let index = 0; index < 6; index++) await Promise.resolve();
};
afterEach(() => {
  controllers.splice(0).forEach((controller) => {
    controller.dispose();
  });
  vi.useRealTimers();
});

describe("Incident membership audit read ownership", () => {
  it("separates drafts applied intent and accepted page through validation and replacement failure", async () => {
    const { controller: c, reads, list } = setup();
    await settle();
    required(reads[0]).resolve(page());
    await settle();
    c.edit("action_code", "membership_deleted");
    c.next();
    await settle();
    expect(list.mock.calls.at(-1)?.[0]).toMatchObject({
      query: emptyMembershipAuditQuery,
      cursor: "next",
      limit: 100,
      authority,
    });
    required(reads[1]).resolve(page("second", null));
    await settle();
    c.edit("occurred_at_lt", "2026-05-24T12:00:00");
    c.apply();
    expect(c.getSnapshot().page?.rows[0]?.audit_event_id).toBe("second");
    expect(reads).toHaveLength(2);
    c.edit("occurred_at_lt", "");
    c.apply();
    await settle();
    expect(c.getSnapshot().page?.query.action_code).toBe("");
    expect(c.getSnapshot().applied.action_code).toBe("membership_deleted");
    c.edit("action_code", "membership_created");
    required(reads[2]).resolve(failure(400, "invalid_list_query"));
    await settle();
    expect(c.getSnapshot().errors).toEqual({});
    expect(c.canContinue()).toBe(false);
    c.refresh();
    await settle();
    expect(list.mock.calls.at(-1)?.[0]).toMatchObject({
      cursor: null,
      query: { action_code: "membership_deleted" },
    });
  });
  it("coalesces duplicate first reads and pages while replacements fence delayed success and errors", async () => {
    const { controller: c, reads } = setup();
    await settle();
    c.apply();
    c.refresh();
    expect(reads).toHaveLength(1);
    c.edit("action_code", "membership_created");
    c.apply();
    await settle();
    c.apply();
    expect(reads).toHaveLength(2);
    required(reads[1]).resolve(page("new"));
    await settle();
    required(reads[0]).resolve(failure(401, "session_required"));
    await settle();
    expect(c.getSnapshot().page?.rows[0]?.audit_event_id).toBe("new");
    c.next();
    c.next();
    await settle();
    expect(reads).toHaveLength(3);
    c.refresh();
    await settle();
    required(reads[3]).resolve(page("fresh"));
    await settle();
    required(reads[2]).resolve(page("late"));
    await settle();
    expect(c.getSnapshot().page?.rows[0]?.audit_event_id).toBe("fresh");
  });
  it("retains one page and twenty prior cursors while permitting older traversal and bounded backward navigation", async () => {
    const { controller: c, reads, list } = setup();
    await settle();
    for (let n = 0; n < 25; n++) {
      required(reads[n]).resolve(page(`event-${n}`, `cursor-${n}`));
      await settle();
      if (n < 24) {
        c.next();
        await settle();
      }
    }
    expect(c.getSnapshot().page?.rows).toHaveLength(1);
    expect(c.getSnapshot().history).toHaveLength(20);
    expect(c.getSnapshot().history[0]?.number).toBe(5);
    c.previous();
    await settle();
    expect(list.mock.calls.at(-1)?.[0]).toMatchObject({ cursor: "cursor-22" });
    required(reads[25]).resolve(page("previous", "cursor-23"));
    await settle();
    for (let index = 0; index < 19; index++) {
      c.previous();
      await settle();
      required(reads[26 + index]).resolve(
        page(`back-${index}`, `forward-${index}`),
      );
      await settle();
    }
    expect(c.getSnapshot().page?.position.number).toBe(5);
    expect(c.getSnapshot().history).toEqual([]);
    c.previous();
    await settle();
    expect(reads).toHaveLength(45);
    c.refresh();
    await settle();
    expect(list.mock.calls.at(-1)?.[0]).toMatchObject({ cursor: null });
    required(reads[45]).resolve(page("first"));
    await settle();
    expect(c.getSnapshot().history).toEqual([]);
  });
  it("requires explicit first page after cursor rejection including close and reopen", async () => {
    const { controller: c, reads, recover } = setup();
    await settle();
    required(reads[0]).resolve(page());
    await settle();
    c.next();
    await settle();
    required(reads[1]).resolve(failure(400, "invalid_pagination_request"));
    await settle();
    expect(c.getSnapshot().read.kind).toBe("cursor_rejected");
    c.next();
    c.retry();
    expect(reads).toHaveLength(2);
    c.setActive(false);
    expect(c.getSnapshot().page).toBeNull();
    c.setActive(true);
    await settle();
    expect(recover).toHaveBeenCalledTimes(3);
    expect(reads).toHaveLength(2);
    expect(c.getSnapshot().read.kind).toBe("cursor_rejected");
    c.refresh();
    await settle();
    expect(reads).toHaveLength(3);
  });
  it("rejects cyclic cursor history without accumulating rows", async () => {
    const { controller: c, reads } = setup();
    await settle();
    required(reads[0]).resolve(page());
    await settle();
    c.next();
    await settle();
    required(reads[1]).resolve(page("second", "third"));
    await settle();
    c.next();
    await settle();
    required(reads[2]).resolve(page("third", "next"));
    await settle();
    expect(c.getSnapshot().read.kind).toBe("cursor_rejected");
    expect(c.getSnapshot().page?.rows[0]?.audit_event_id).toBe("second");
  });
  it("preserves only filters on section drawer and visibility exit and fences late results", async () => {
    for (const completion of [page("late"), failure(401, "session_required")]) {
      const { controller: c, reads, lost } = setup();
      await settle();
      required(reads[0]).resolve(page());
      await settle();
      c.edit("action_code", "membership_deleted");
      c.apply();
      await settle();
      c.edit("target_id", "draft");
      c.toggleExpanded("event-1");
      c.setActive(false);
      required(reads[1]).resolve(completion);
      await settle();
      expect(c.getSnapshot()).toMatchObject({
        active: false,
        page: null,
        history: [],
        expandedId: null,
        errors: {},
        inputs: { target_id: "draft" },
        applied: { action_code: "membership_deleted" },
      });
      expect(lost).not.toHaveBeenCalled();
      c.setActive(true);
      await settle();
      expect(reads).toHaveLength(3);
    }
  });
  it("clears queries details and admissions on incident actor session and role retirement without resurrection", async () => {
    for (const next of [
      null,
      { ...authority, incidentId: "incident-2" },
      { ...authority, actorId: "actor-2" },
      { ...authority, lifetime: "session-2" },
    ]) {
      for (const completion of [
        page("late"),
        failure(404, "incident_not_found"),
      ]) {
        const { controller: c, reads, lost } = setup();
        await settle();
        required(reads[0]).resolve(page());
        await settle();
        c.edit("target_id", "private draft");
        c.toggleExpanded("event-1");
        c.refresh();
        await settle();
        c.setAuthority(next);
        if (next === null) {
          expect(c.getSnapshot().page).toBeNull();
          expect(c.getSnapshot().inputs).toEqual(emptyMembershipAuditQuery);
          c.setAuthority(authority);
        }
        required(reads[1]).resolve(completion);
        await settle();
        expect(c.getSnapshot()).toMatchObject({
          page: null,
          inputs: emptyMembershipAuditQuery,
          applied: emptyMembershipAuditQuery,
          expandedId: null,
          history: [],
          errors: {},
        });
        expect(lost).not.toHaveBeenCalled();
      }
    }
  });
  it("distinguishes current session hidden incident and visible role rejection and clears protected state", async () => {
    for (const [status, code, reason] of [
      [401, "session_required", "session"],
      [404, "incident_not_found", "incident"],
      [403, "authorization_denied", "role"],
    ] as const) {
      const { controller: c, reads, lost } = setup();
      await settle();
      required(reads[0]).resolve(page());
      await settle();
      c.edit("target_id", "private");
      c.refresh();
      await settle();
      required(reads[1]).resolve(failure(status, code));
      await settle();
      expect(c.getSnapshot()).toMatchObject({
        page: null,
        inputs: emptyMembershipAuditQuery,
        read: { kind: "denied", reason },
      });
      expect(lost).toHaveBeenCalledWith(reason, authority);
      c.setAuthority(authority);
      expect(reads).toHaveLength(2);
    }
  });
  it("distinguishes nonadmin recovery from temporary and cancelled observation without assuming revocation", async () => {
    const { controller: c, recover, reads, lost } = setup();
    await settle();
    required(reads[0]).resolve(page());
    await settle();
    for (const result of [
      { kind: "unavailable", failure: "transient" },
      { kind: "unavailable", failure: "contract" },
      { kind: "cancelled" },
    ] as const) {
      recover.mockResolvedValueOnce(result);
      c.refresh();
      await settle();
      expect(c.getSnapshot().page).not.toBeNull();
      expect(c.getSnapshot().read.kind).toBe("failed");
      expect(lost).not.toHaveBeenCalled();
    }
    recover.mockResolvedValueOnce({
      kind: "authorized",
      role: "reviewer",
      userId: authority.actorId,
    });
    c.retry();
    await settle();
    expect(lost).toHaveBeenCalledWith("role", authority);
    expect(c.getSnapshot().page).toBeNull();
  });
  it("bounds the entire observation and ignores transport completion after timeout or disposal", async () => {
    vi.useFakeTimers();
    const { controller: c, reads } = setup();
    await settle();
    await vi.advanceTimersByTimeAsync(membershipAuditDeadline);
    expect(c.getSnapshot().read.kind).toBe("failed");
    required(reads[0]).resolve(page("late"));
    await settle();
    expect(c.getSnapshot().page).toBeNull();
    c.retry();
    await settle();
    c.dispose();
    required(reads[1]).resolve(page("disposed"));
    await settle();
    expect(c.getSnapshot()).toEqual(
      expect.objectContaining({ page: null, active: false }),
    );
    const disposed = setup();
    await settle();
    disposed.controller.dispose();
    required(disposed.reads[0]).resolve(failure(401, "session_required"));
    await settle();
    expect(disposed.lost).not.toHaveBeenCalled();
    expect(disposed.controller.getSnapshot()).toMatchObject({
      page: null,
      active: false,
    });
    const observation =
      deferred<Awaited<ReturnType<MembershipAuditPorts["recover"]>>>();
    const pending = setup({ recover: () => observation.promise });
    await vi.advanceTimersByTimeAsync(membershipAuditDeadline);
    observation.resolve({
      kind: "authorized",
      role: "admin",
      userId: authority.actorId,
    });
    await settle();
    expect(pending.list).not.toHaveBeenCalled();
    expect(pending.controller.getSnapshot().read.kind).toBe("failed");
  });
  it("fences late access observation success and losses across every retired admission", async () => {
    for (const boundary of [
      "suspend",
      "incident",
      "actor",
      "session",
      "role",
      "replace",
      "dispose",
    ] as const) {
      for (const outcome of [
        { kind: "authorized", role: "admin", userId: authority.actorId },
        { kind: "authorized", role: "viewer", userId: authority.actorId },
        { kind: "session_lost" },
        { kind: "access_lost" },
        { kind: "unavailable", failure: "transient" },
      ] as const) {
        const observation =
          deferred<Awaited<ReturnType<MembershipAuditPorts["recover"]>>>();
        const recover = vi
          .fn<MembershipAuditPorts["recover"]>(async (subject) => ({
            kind: "authorized",
            role: "admin",
            userId: subject.actorId,
          }))
          .mockImplementationOnce(() => observation.promise);
        const { controller: c, lost } = setup({ recover });
        c.edit("target_id", "private draft");
        if (boundary === "suspend") c.setActive(false);
        else if (boundary === "dispose") c.dispose();
        else if (boundary === "role") c.setAuthority(null);
        else if (boundary === "replace") {
          c.edit("target_id", "");
          c.edit("action_code", "membership_deleted");
          c.apply();
        } else
          c.setAuthority({
            ...authority,
            ...(boundary === "incident"
              ? { incidentId: "incident-2" }
              : boundary === "actor"
                ? { actorId: "actor-2" }
                : { lifetime: "session-2" }),
          });
        await settle();
        const accepted = c.getSnapshot();
        observation.resolve(outcome);
        await settle();
        expect(c.getSnapshot()).toBe(accepted);
        expect(lost).not.toHaveBeenCalled();
      }
    }
  });
  it("does not publish status details or callbacks when the external authority is no longer current", async () => {
    let current = true;
    const { controller: c, reads, lost } = setup({ isCurrent: () => current });
    await settle();
    current = false;
    const snapshot = c.getSnapshot();
    required(reads[0]).resolve(failure(401, "session_required"));
    await settle();
    c.toggleExpanded("event-1");
    c.edit("target_id", "blocked");
    c.refresh();
    expect(c.getSnapshot()).toBe(snapshot);
    expect(lost).not.toHaveBeenCalled();
  });
});

function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Expected audit fixture value");
  return value;
}
