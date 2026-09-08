import type {
  GetCurrentSessionResponse,
  ListIncidentMembershipAuditEventsResponse,
} from "@cartulary/protocol-ts/http";
import { expect, type Locator, type Page } from "@playwright/test";
import { openIncidentControls } from "../pages/deploymentAdministration";
import { openIncidentFromLanding } from "../pages/incidentDirectory";
import { auditBrowserEvent } from "./administrativeAudit";
import { createIncident } from "./incidents/fixtures";
import { uniqueIncidentKey } from "./runtime/fixtureIdentity";

export const membershipAuditRegion = "Incident membership audit browser";
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
export async function installMembershipAuditPresentation(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("MA"),
    "Membership audit review",
  );
  let access: "admin" | "viewer" | "hidden" | "session" | "unavailable" =
    "admin";
  let accessReads = 0;
  await page.route("**/api/v1/auth/session", async (route) => {
    ++accessReads;
    if (access === "session" || access === "unavailable") {
      await route.fulfill({
        status: access === "session" ? 401 : 503,
        json: {
          error: {
            code:
              access === "session" ? "session_required" : "service_unavailable",
            message: "Not displayed",
            retryable: access === "unavailable",
          },
          meta: { request_id: "membership-access" },
        },
      });
      return;
    }
    const response = await route.fetch();
    const envelope: GetCurrentSessionResponse = await response.json();
    await route.fulfill({
      response,
      json: {
        ...envelope,
        data: {
          ...envelope.data,
          display_name: "Audit operator",
          memberships: envelope.data.memberships.flatMap((member) =>
            member.incident_id !== incidentId
              ? [member]
              : access === "hidden"
                ? []
                : [
                    {
                      ...member,
                      role: access === "viewer" ? "viewer" : "admin",
                    },
                  ],
          ),
        },
      },
    });
  });
  let body: unknown;
  let status = 200;
  let gate: Promise<void> | null = null;
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
  await page.route(
    `**/api/v1/incidents/${incidentId}/membership-audit-events*`,
    async (route) => {
      requests.push(new URL(route.request().url()));
      const captured = { body, status, gate };
      gate = null;
      if (captured.gate) await captured.gate;
      await route
        .fulfill({ status: captured.status, json: captured.body })
        .catch(() => {});
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
    gateRead(value: Promise<void>) {
      gate = value;
    },
    malformed(value: unknown) {
      status = 200;
      body = value;
    },
    fail(code = "service_unavailable", reason?: string) {
      status =
        code === "invalid_pagination_request" || code === "invalid_list_query"
          ? 400
          : code === "authorization_denied"
            ? 403
            : code === "incident_not_found"
              ? 404
              : code === "session_required"
                ? 401
                : 503;
      body = {
        error: {
          code,
          message: "Private diagnostic must remain hidden",
          retryable: status === 503,
          ...(reason ? { details: { reason_code: reason } } : {}),
        },
        meta: { request_id: "membership-failure" },
      };
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
