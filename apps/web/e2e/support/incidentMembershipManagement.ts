import type { ListIncidentMembershipsResponse } from "@cartulary/protocol-ts/http";
import { expect, type Locator, type Page } from "@playwright/test";
import { openIncidentControls } from "../pages/deploymentAdministration";
import { openIncidentFromLanding } from "../pages/incidentDirectory";
import { rewriteSessionPresentation } from "./auth/sessionPresentation";
import { createIncident } from "./incidents/fixtures";
import { uniqueIncidentKey } from "./runtime/fixtureIdentity";

type Member = ListIncidentMembershipsResponse["data"]["memberships"][number];
const membershipRegion = "Incident memberships";
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
export async function installMembershipManagementPresentation(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("MM"),
    "Incident access review",
  );
  let access: "admin" | "viewer" | "hidden" | "session" | "unavailable" =
    "admin";
  await page.route("**/api/v1/auth/session", async (route) => {
    if (access === "session" || access === "unavailable") {
      await route.fulfill({
        status: access === "session" ? 401 : 503,
        json: {
          error: {
            code:
              access === "session" ? "session_required" : "service_unavailable",
          },
          meta: { request_id: "membership-access" },
        },
      });
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
  let readGate: Promise<void> | null = null;
  let writeGate: Promise<void> | null = null;
  let writeStatus = 200;
  let writeBody: unknown;
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
  await page.route(
    `**/api/v1/incidents/${incidentId}/memberships**`,
    async (route) => {
      const method = route.request().method();
      if (method === "GET") {
        reads.push(new URL(route.request().url()));
        const captured = { body, status, gate: readGate };
        readGate = null;
        if (captured.gate !== null) await captured.gate;
        await route
          .fulfill({ status: captured.status, json: captured.body })
          .catch(() => {});
        return;
      }
      writes.push({ method, body: route.request().postDataJSON() });
      const captured = {
        body: writeBody,
        status: writeStatus,
        gate: writeGate,
      };
      writeGate = null;
      if (captured.gate !== null) await captured.gate;
      await route
        .fulfill(
          captured.status === 204
            ? { status: 204, body: "" }
            : { status: captured.status, json: captured.body },
        )
        .catch(() => {});
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
    gateRead(value: Promise<void>) {
      readGate = value;
    },
    gateWrite(value: Promise<void>) {
      writeGate = value;
    },
    failRead(code = "service_unavailable") {
      status = code === "invalid_pagination_request" ? 400 : 503;
      body = {
        error: { code, message: "Private diagnostic" },
        meta: { request_id: "failed-list" },
      };
    },
    mutation(
      nextStatus: number,
      resource: Member | null = null,
      code = "service_unavailable",
    ) {
      writeStatus = nextStatus;
      writeBody = resource
        ? { data: resource, meta: { request_id: "membership-write" } }
        : {
            error: { code, message: "Private diagnostic" },
            meta: { request_id: "membership-error" },
          };
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
