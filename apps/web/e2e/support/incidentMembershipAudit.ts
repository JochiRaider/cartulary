import type {
  HTTPOperationErrorCode,
  ListIncidentMembershipAuditEventsResponse,
} from "@cartulary/protocol-ts/http";
import { expect, type Locator, type Page } from "@playwright/test";
import { openIncidentControls } from "../pages/deploymentAdministration";
import { openIncidentFromLanding } from "../pages/incidentDirectory";
import { auditBrowserEvent } from "./administrativeAudit";
import { rewriteSessionPresentation } from "./auth/sessionPresentation";
import { createIncident } from "./incidents/fixtures";
import type { BarrierOutcome } from "./runtime/cleanup";
import { uniqueIncidentKey } from "./runtime/fixtureIdentity";
import {
  awaitResponseBarrier,
  checkedOperationErrorEnvelope,
  fulfillOperationError,
  installOwnedRoute,
} from "./transport/requestInterception";

const membershipAuditFailureStatus = {
  internal_error: 500,
  invalid_list_query: 400,
  invalid_pagination_request: 400,
  session_required: 401,
  authorization_denied: 403,
  incident_not_found: 404,
} as const satisfies Partial<
  Record<
    | HTTPOperationErrorCode<"listIncidentMembershipAuditEvents">
    | "internal_error",
    number
  >
>;
type MembershipAuditFailureCode = keyof typeof membershipAuditFailureStatus;

const membershipAuditRegion = "Incident membership audit browser";
type Event =
  ListIncidentMembershipAuditEventsResponse["data"]["audit_events"][number];
export function membershipBrowserEvent(
  incidentId: string,
  overrides: Partial<Event> = {},
): Event {
  return auditBrowserEvent({
    scope_kind: "incident",
    scope_id: incidentId,
    action_code: "membership_role_changed",
    target_kind: "incident_membership",
    ...overrides,
  });
}
export async function openMembershipAudit(page: Page) {
  await openIncidentControls(page, "membership-audit");
  return page.getByRole("region", { name: membershipAuditRegion, exact: true });
}
/** Public service fixture for the workbook; deterministic audit/session responses only. */
export async function installMembershipAuditPresentation(
  page: Page,
  existingIncidentId?: string,
) {
  const incidentId =
    existingIncidentId ??
    (await createIncident(
      page,
      uniqueIncidentKey("MA"),
      "Membership audit review",
    ));
  let access: "admin" | "viewer" | "hidden" | "session" | "unavailable" =
    "admin";
  let accessReads = 0;
  await installOwnedRoute(page, "**/api/v1/auth/session", async (route) => {
    ++accessReads;
    if (access === "session") {
      await fulfillOperationError(
        route,
        "getCurrentSession",
        401,
        "session_required",
        { message: "Not displayed", requestId: "membership-access" },
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
      display_name: "Audit operator",
      memberships: session.memberships.flatMap((member) =>
        member.incident_id !== incidentId
          ? [member]
          : requestAccess === "hidden"
            ? []
            : [
                {
                  ...member,
                  role: requestAccess === "viewer" ? "viewer" : "admin",
                },
              ],
      ),
    }));
  });
  let body: unknown;
  let status = 200;
  let gate: Promise<BarrierOutcome> | null = null;
  const requests: URL[] = [];
  const setPage = (rows: Event[], cursor: string | null = null) => {
    status = 200;
    body = {
      data: { audit_events: rows },
      meta: {
        request_id: "membership-browser",
        paging:
          cursor === null
            ? { limit: 100, has_more: false, next_cursor: null }
            : { limit: 100, has_more: true, next_cursor: cursor },
      },
    } satisfies ListIncidentMembershipAuditEventsResponse;
  };
  setPage([membershipBrowserEvent(incidentId)]);
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/membership-audit-events*`,
    async (route) => {
      requests.push(new URL(route.request().url()));
      const captured = { body, status, gate };
      gate = null;
      if (!(await awaitResponseBarrier(route, captured.gate))) return;
      await route.fulfill({ status: captured.status, json: captured.body });
    },
  );
  await openIncidentFromLanding(page, incidentId);
  return {
    incidentId,
    requests,
    setPage,
    get accessReads() {
      return accessReads;
    },
    setAccess(value: typeof access) {
      access = value;
    },
    gateRead(value: Promise<BarrierOutcome>) {
      gate = value;
    },
    malformed(value: unknown) {
      status = 200;
      body = value;
    },
    /** A failure listIncidentMembershipAuditEvents declares; internal_error by default. */
    fail(code: MembershipAuditFailureCode = "internal_error", reason?: string) {
      status = membershipAuditFailureStatus[code];
      body = checkedOperationErrorEnvelope(
        "listIncidentMembershipAuditEvents",
        status,
        code,
        {
          message: "Private diagnostic must remain hidden",
          requestId: "membership-failure",
          ...(reason ? { details: { reason_code: reason } } : {}),
        },
      );
    },
  };
}
export async function expectMembershipControlReachable(
  page: Page,
  control: Locator,
) {
  const panel = page.getByRole("region", {
    name: membershipAuditRegion,
    exact: true,
  });
  expect(
    await panel.evaluate((element) => ({
      fits: element.scrollWidth <= element.clientWidth + 1,
      documentFits:
        document.documentElement.scrollWidth <=
        document.documentElement.clientWidth + 1,
    })),
  ).toEqual({ fits: true, documentFits: true });
  await control.focus();
  await control.scrollIntoViewIfNeeded();
  await expect(control).toBeFocused();
  const box = await control.boundingBox();
  const viewport = page.viewportSize();
  if (!box || !viewport) throw new Error("Missing control viewport");
  expect(box.x).toBeGreaterThanOrEqual(-1);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
  expect(box.y).toBeGreaterThanOrEqual(-1);
  expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
}
