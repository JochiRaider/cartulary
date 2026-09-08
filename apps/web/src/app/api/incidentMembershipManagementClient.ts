import type {
  CreateIncidentMembershipResponse,
  ListIncidentMembershipsResponse,
  PatchIncidentMembershipResponse,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import { normalizeAuditInstant } from "../../shared/auditReadValues";
import type {
  Membership,
  MembershipAuthority,
  MembershipInput,
  MembershipPaging,
  MembershipProblem,
} from "../incidentMembershipManagementModel";
import {
  membershipPageLimit,
  membershipProblemCodes,
} from "../incidentMembershipManagementModel";

export type MembershipFailure = {
  readonly ok: false;
  readonly status: number;
  readonly problem: MembershipProblem;
};
export type MembershipListResult =
  | {
      readonly ok: true;
      readonly rows: readonly Membership[];
      readonly paging: MembershipPaging;
    }
  | MembershipFailure;
export type MembershipMutationResult =
  | {
      readonly ok: true;
      readonly status: 200 | 201 | 204;
      readonly resource: Membership | null;
    }
  | MembershipFailure;
const invalid = (): MembershipFailure => ({
  ok: false,
  status: 502,
  problem: { code: "invalid_public_contract_response" },
});
function failure(
  status: number,
  error?: { code?: string; details?: Record<string, unknown> },
): MembershipFailure {
  return {
    ok: false,
    status,
    problem: {
      code:
        membershipProblemCodes.find((code) => code === error?.code) ??
        "unknown_public_error",
      field: ["email", "role", "base_membership_version"].find(
        (field) => field === error?.details?.field,
      ),
    },
  };
}
export async function listIncidentMembershipPage(request: {
  readonly authority: MembershipAuthority;
  readonly cursor: string | null;
  readonly signal: AbortSignal;
}): Promise<MembershipListResult> {
  const result = await fetchHTTPOperation<ListIncidentMembershipsResponse>({
    operationID: "listIncidentMemberships",
    apiBase: request.authority.apiBase,
    pathParameters: { incident_id: request.authority.incidentId },
    query: {
      limit: membershipPageLimit,
      ...(request.cursor === null ? {} : { cursor_token: request.cursor }),
    },
    init: { signal: request.signal },
  });
  if (!result.ok) return failure(result.status, result.payload.error);
  const { memberships: rows } = result.payload.data;
  const paging = result.payload.meta.paging;
  if (
    result.status !== 200 ||
    !paging ||
    paging.limit !== membershipPageLimit ||
    rows.length > membershipPageLimit ||
    (paging.has_more ? !paging.next_cursor : paging.next_cursor !== null)
  )
    return invalid();
  if (paging.has_more && paging.next_cursor === request.cursor)
    return {
      ok: false,
      status: 400,
      problem: { code: "invalid_pagination_request" },
    };
  let previous: { time: bigint; id: string } | null = null;
  const ids = new Set<string>();
  for (const row of rows) {
    const instant = normalizeAuditInstant(row.joined_at);
    if (
      row.incident_id !== request.authority.incidentId ||
      !instant ||
      ids.has(row.user_id) ||
      !Number.isSafeInteger(row.membership_version)
    )
      return invalid();
    if (
      previous &&
      (instant.nanoseconds < previous.time ||
        (instant.nanoseconds === previous.time && row.user_id <= previous.id))
    )
      return invalid();
    previous = { time: instant.nanoseconds, id: row.user_id };
    ids.add(row.user_id);
  }
  return { ok: true, rows, paging };
}
export async function mutateIncidentMembership(request: {
  readonly authority: MembershipAuthority;
  readonly input: MembershipInput;
  readonly signal: AbortSignal;
}): Promise<MembershipMutationResult> {
  const { input, authority } = request;
  const operationID =
    input.kind === "add"
      ? "createIncidentMembership"
      : input.kind === "role"
        ? "patchIncidentMembership"
        : "deleteIncidentMembership";
  const result = await fetchHTTPOperation<
    CreateIncidentMembershipResponse | PatchIncidentMembershipResponse
  >({
    operationID,
    apiBase: authority.apiBase,
    pathParameters: {
      incident_id: authority.incidentId,
      ...(input.kind === "add" ? {} : { user_id: input.userId }),
    },
    init: {
      method:
        input.kind === "add"
          ? "POST"
          : input.kind === "role"
            ? "PATCH"
            : "DELETE",
      body: JSON.stringify(input.payload),
      signal: request.signal,
    },
  });
  if (!result.ok) {
    const rejected = failure(result.status, result.payload.error);
    return rejected.problem.code === "unknown_public_error" ||
      rejected.problem.code === "invalid_public_contract_response"
      ? invalid()
      : rejected;
  }
  if (input.kind === "remove")
    return result.status === 204
      ? { ok: true, status: 204, resource: null }
      : invalid();
  if (
    input.kind === "add"
      ? result.status !== 200 && result.status !== 201
      : result.status !== 200
  )
    return invalid();
  const resource = result.payload.data;
  if (
    resource.incident_id !== authority.incidentId ||
    !Number.isSafeInteger(resource.membership_version) ||
    (input.kind === "role" && resource.user_id !== input.userId) ||
    resource.role !== input.payload.role
  )
    return invalid();
  return { ok: true, status: result.status as 200 | 201, resource };
}
