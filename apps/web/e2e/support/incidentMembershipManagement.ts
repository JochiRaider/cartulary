import type {
  HTTPOperationErrorCode,
  ListIncidentMembershipsResponse,
} from "@cartulary/protocol-ts/http";
import { expect, type Locator, type Page } from "@playwright/test";
import { openIncidentControls } from "../pages/deploymentAdministration";
import { openIncidentFromLanding } from "../pages/incidentDirectory";
import { rewriteSessionPresentation } from "./auth/sessionPresentation";
import { createIncident } from "./incidents/fixtures";
import type { BarrierOutcome } from "./runtime/cleanup";
import { uniqueIncidentKey } from "./runtime/fixtureIdentity";
import {
  awaitResponseBarrier,
  checkedOperationErrorEnvelope,
  fulfillOperationError,
  installOwnedRoute,
  operationErrorEnvelope,
} from "./transport/requestInterception";

type Member = ListIncidentMembershipsResponse["data"]["memberships"][number];
const membershipRegion = "Incident memberships";
type MembershipWriteOperation =
  | "createIncidentMembership"
  | "deleteIncidentMembership"
  | "patchIncidentMembership";
type MembershipWriteErrorCode =
  | HTTPOperationErrorCode<MembershipWriteOperation>
  | "internal_error";

function membershipWriteOperation(method: string): MembershipWriteOperation {
  if (method === "POST") return "createIncidentMembership";
  if (method === "PATCH") return "patchIncidentMembership";
  if (method === "DELETE") return "deleteIncidentMembership";
  throw new Error(`Unexpected membership write method ${method}`);
}

export function membershipBrowserMember(
  incidentId: string,
  overrides: Partial<Member> = {},
): Member {
  return {
    incident_id: incidentId,
    user_id: "00000000-0000-4000-8000-000000000002",
    display_name: "Response analyst",
    role: "viewer",
    membership_version: 1,
    joined_at: "2026-08-01T00:00:00Z",
    updated_at: "2026-08-01T00:00:00Z",
    added_by_user_id: "00000000-0000-4000-8000-000000000001",
    updated_by_user_id: "00000000-0000-4000-8000-000000000001",
    ...overrides,
  };
}
export async function openMembershipManagement(page: Page) {
  await openIncidentControls(page, "memberships");
  return page.getByRole("region", { name: membershipRegion, exact: true });
}
/** Real workbook service fixture; deterministic responses are limited to this seam. */
export async function installMembershipManagementPresentation(
  page: Page,
  existingIncidentId?: string,
) {
  const incidentId =
    existingIncidentId ??
    (await createIncident(
      page,
      uniqueIncidentKey("MM"),
      "Incident access review",
    ));
  let access: "admin" | "viewer" | "hidden" | "session" | "unavailable" =
    "admin";
  await installOwnedRoute(page, "**/api/v1/auth/session", async (route) => {
    if (access === "session") {
      await fulfillOperationError(
        route,
        "getCurrentSession",
        401,
        "session_required",
      );
      return;
    }
    if (access === "unavailable") {
      await fulfillOperationError(
        route,
        "getCurrentSession",
        500,
        "internal_error",
      );
      return;
    }
    const requestAccess = access;
    await rewriteSessionPresentation(route, (session) => ({
      ...session,
      display_name: "Membership operator",
      memberships: session.memberships.flatMap((m) =>
        m.incident_id !== incidentId
          ? [m]
          : requestAccess === "hidden"
            ? []
            : [{ ...m, role: requestAccess === "viewer" ? "viewer" : "admin" }],
      ),
    }));
  });
  let body: unknown;
  let status = 200;
  let readGate: Promise<BarrierOutcome> | null = null;
  let writeGate: Promise<BarrierOutcome> | null = null;
  let writeStatus = 200;
  let writeBody: unknown;
  let writeError: MembershipWriteErrorCode | null = null;
  const reads: URL[] = [];
  const writes: { method: string; body: Record<string, unknown> }[] = [];
  const setPage = (rows: Member[], next: string | null = null) => {
    status = 200;
    body = {
      data: { memberships: rows },
      meta: {
        request_id: "membership-list",
        paging:
          next === null
            ? { limit: 100, has_more: false, next_cursor: null }
            : { limit: 100, has_more: true, next_cursor: next },
      },
    } satisfies ListIncidentMembershipsResponse;
  };
  setPage([membershipBrowserMember(incidentId)]);
  writeBody = {
    data: membershipBrowserMember(incidentId, {
      role: "reviewer",
      membership_version: 2,
    }),
    meta: { request_id: "membership-write" },
  };
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/memberships**`,
    async (route) => {
      const method = route.request().method();
      if (method === "GET") {
        reads.push(new URL(route.request().url()));
        const captured = { body, status, gate: readGate };
        readGate = null;
        if (!(await awaitResponseBarrier(route, captured.gate))) return;
        await route.fulfill({ status: captured.status, json: captured.body });
        return;
      }
      writes.push({ method, body: route.request().postDataJSON() });
      const captured = {
        body: writeBody,
        status: writeStatus,
        error: writeError,
        gate: writeGate,
      };
      writeGate = null;
      if (!(await awaitResponseBarrier(route, captured.gate))) return;
      await route.fulfill(
        captured.status === 204
          ? { status: 204, body: "" }
          : {
              status: captured.status,
              json:
                captured.error === null
                  ? captured.body
                  : checkedOperationErrorEnvelope(
                      membershipWriteOperation(method),
                      captured.status,
                      captured.error,
                      {
                        message: "Private diagnostic",
                        requestId: "membership-error",
                      },
                    ),
            },
      );
    },
  );
  await openIncidentFromLanding(page, incidentId);
  return {
    incidentId,
    reads,
    writes,
    setPage,
    setAccess(value: typeof access) {
      access = value;
    },
    gateRead(value: Promise<BarrierOutcome>) {
      readGate = value;
    },
    gateWrite(value: Promise<BarrierOutcome>) {
      writeGate = value;
    },
    failRead(
      code:
        | "internal_error"
        | "invalid_list_query"
        | "invalid_pagination_request" = "internal_error",
    ) {
      status = code === "internal_error" ? 500 : 400;
      body =
        code === "internal_error"
          ? operationErrorEnvelope("listIncidentMemberships", 500, code)
          : operationErrorEnvelope("listIncidentMemberships", 400, code, {
              message: "Private diagnostic",
              requestId: "failed-list",
            });
    },
    /** A success resource, or an error code the write operation declares. */
    mutation(
      nextStatus: number,
      resource: Member | null = null,
      code: MembershipWriteErrorCode = "internal_error",
    ) {
      writeStatus = nextStatus;
      writeError = resource || nextStatus < 400 ? null : code;
      writeBody = resource
        ? { data: resource, meta: { request_id: "membership-write" } }
        : null;
    },
  };
}
export async function expectMembershipManagementControlReachable(
  page: Page,
  control: Locator,
) {
  const panel = page.getByRole("region", {
    name: membershipRegion,
    exact: true,
  });
  expect(
    await panel.evaluate(
      (e) =>
        e.scrollWidth <= e.clientWidth + 1 &&
        document.documentElement.scrollWidth <=
          document.documentElement.clientWidth + 1,
    ),
  ).toBe(true);
  await control.focus();
  await control.scrollIntoViewIfNeeded();
  await expect(control).toBeFocused();
  const box = await control.boundingBox();
  const viewport = page.viewportSize();
  if (!box || !viewport) throw new Error("Missing membership control viewport");
  expect(box.x).toBeGreaterThanOrEqual(-1);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
  expect(box.y).toBeGreaterThanOrEqual(-1);
  expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
}
