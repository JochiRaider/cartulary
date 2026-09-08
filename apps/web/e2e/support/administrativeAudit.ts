import type {
  GetCurrentSessionResponse,
  ListAdministrativeAuditEventsResponse,
} from "@cartulary/protocol-ts/http";
import { expect, type Locator, type Page } from "@playwright/test";
import { DeploymentAdministration } from "../pages/deploymentAdministration";

export const auditBrowserPath = "/api/v1/administrative-audit-events";
export const auditBrowserEventId = "00000000-0000-4000-8000-000000002001";
type Event =
  ListAdministrativeAuditEventsResponse["data"]["audit_events"][number];
export function auditBrowserEvent(overrides: Partial<Event> = {}): Event {
  return {
    audit_event_id: auditBrowserEventId,
    scope_kind: "deployment",
    scope_id: null,
    occurred_at: "2026-05-24T12:00:00.123456789Z",
    actor_kind: "user",
    actor_user_id: "00000000-0000-4000-8000-000000000001",
    source: "ui",
    action_code: "user_profile_updated",
    target_kind: "user",
    target_id: "00000000-0000-4000-8000-000000000002",
    reason_code: null,
    changes: [
      {
        field_path: "display_name",
        value_state: "visible",
        before: "",
        after: "Audit reader",
      },
      {
        field_path: "future.counter",
        value_state: "visible",
        before: false,
        after: 0,
      },
      {
        field_path: `future.${"long_field_path_".repeat(8)}.value`,
        value_state: "visible",
        before: null,
        after: {
          label: "<script>inert text</script>",
          nested: [false, 0, "", null],
        },
      },
      {
        field_path: "password",
        value_state: "redacted",
        before: null,
        after: null,
      },
    ],
    ...overrides,
  };
}
export function auditBrowserEnvelope(
  rows: Event[] = [auditBrowserEvent()],
  cursor: string | null = null,
): ListAdministrativeAuditEventsResponse {
  return {
    data: { audit_events: rows },
    meta: {
      request_id: "audit-browser-fixture",
      paging:
        cursor === null
          ? { limit: 100, has_more: false, next_cursor: null }
          : { limit: 100, has_more: true, next_cursor: cursor },
    },
  };
}
export function auditBrowserBarrier() {
  let release = () => {};
  const promise = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { promise, release };
}
export async function openAdministrativeAudit(page: Page) {
  await page.goto("/deployment-administration");
  await new DeploymentAdministration(page).selectPanel("administrative-audit");
  return page.getByRole("region", {
    name: "Deployment audit browser",
    exact: true,
  });
}

/** Deterministic published HTTP projections for presentation and read races. */
export async function installAuditPresentation(page: Page) {
  let accessReads = 0;
  await page.route("**/api/v1/auth/session", async (route) => {
    ++accessReads;
    const response = await route.fetch();
    const envelope: GetCurrentSessionResponse = await response.json();
    await route.fulfill({
      response,
      json: {
        ...envelope,
        data: { ...envelope.data, display_name: "Audit operator" },
      },
    });
  });
  let body = auditBrowserEnvelope();
  let status = 200;
  let gate: Promise<void> | null = null;
  let error: unknown = null;
  const requests: URL[] = [];
  await page.route(`**${auditBrowserPath}*`, async (route) => {
    requests.push(new URL(route.request().url()));
    const acceptedBody = error ?? body;
    const acceptedStatus = status;
    const acceptedGate = gate;
    gate = null;
    if (acceptedGate !== null) await acceptedGate;
    await route
      .fulfill({ status: acceptedStatus, json: acceptedBody })
      .catch(() => {});
  });
  return {
    get accessReads() {
      return accessReads;
    },
    requests,
    setPage(rows: Event[], cursor: string | null = null) {
      body = auditBrowserEnvelope(rows, cursor);
      status = 200;
      error = null;
    },
    gateRead(promise: Promise<void>) {
      gate = promise;
    },
    fail(code = "service_unavailable", reason?: string) {
      status =
        code === "invalid_pagination_request" || code === "invalid_list_query"
          ? 400
          : 503;
      error = {
        error: {
          code,
          message: "Server diagnostic must not be shown",
          retryable: status === 503,
          ...(reason ? { details: { reason_code: reason } } : {}),
        },
        meta: { request_id: "audit-failure" },
      };
    },
  };
}
export async function expectAuditControlReachable(
  page: Page,
  control: Locator,
) {
  const layout = await page
    .getByRole("region", { name: "Deployment audit browser", exact: true })
    .evaluate((element) => {
      const box = element.getBoundingClientRect();
      return {
        fits:
          box.left >= -1 &&
          box.right <= window.innerWidth + 1 &&
          element.scrollWidth <= element.clientWidth + 1 &&
          document.documentElement.scrollWidth <=
            document.documentElement.clientWidth + 1,
        left: box.left,
        right: box.right,
        width: window.innerWidth,
        scroll: element.scrollWidth,
        client: element.clientWidth,
      };
    });
  expect(layout).toMatchObject({ fits: true });
  await control.focus();
  await control.scrollIntoViewIfNeeded();
  await expect(control).toBeFocused();
  const box = await control.boundingBox();
  const viewport = page.viewportSize();
  expect(box).not.toBeNull();
  expect(viewport).not.toBeNull();
  if (!box || !viewport) return;
  expect(box.x).toBeGreaterThanOrEqual(-1);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
  expect(box.y).toBeGreaterThanOrEqual(-1);
  expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
}
