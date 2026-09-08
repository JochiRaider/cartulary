import type {
  Membership,
  MembershipAuthority,
  MembershipPaging,
} from "../app/incidentMembershipManagementModel";
export const membershipAuthority: MembershipAuthority = {
  incidentId: "00000000-0000-4000-8000-000000001001",
  actorId: "00000000-0000-4000-8000-000000000001",
  lifetime: "membership-session",
  role: "admin",
};
export const membershipFixture = (
  overrides: Partial<Membership> = {},
): Membership => ({
  incident_id: membershipAuthority.incidentId,
  user_id: "00000000-0000-4000-8000-000000000002",
  display_name: "Analyst",
  role: "viewer",
  membership_version: 1,
  joined_at: "2026-08-01T00:00:00Z",
  added_by_user_id: membershipAuthority.actorId,
  updated_at: "2026-08-01T00:00:00Z",
  updated_by_user_id: membershipAuthority.actorId,
  ...overrides,
});
export function membershipPage(
  rows: readonly Membership[] = [membershipFixture()],
  next: string | null = null,
) {
  return {
    ok: true as const,
    rows,
    paging: (next === null
      ? { limit: 100, has_more: false, next_cursor: null }
      : {
          limit: 100,
          has_more: true,
          next_cursor: next,
        }) satisfies MembershipPaging,
  };
}
export function membershipEnvelope(
  rows: readonly Membership[] = [membershipFixture()],
  next: string | null = null,
) {
  return {
    data: { memberships: rows },
    meta: {
      request_id: "membership-test",
      paging: membershipPage(rows, next).paging,
    },
  };
}
export const membershipJSON = (payload: unknown, status = 200) =>
  new Response(JSON.stringify(payload), {
    status,
    headers: { "Content-Type": "application/json" },
  });
export function membershipDeferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
