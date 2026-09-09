import type { CloseIncidentRequest } from "@cartulary/protocol-ts/http";
import { expect, type Page } from "@playwright/test";
import { openIncidentControls } from "../pages/deploymentAdministration";
import { csrfHeaders } from "./auth/browserSession";
import { apiBase } from "./runtime/configuration";
import { publicHttpOperation } from "./transport/publicHttpOperationClient";
import { atJsonOrigin } from "./transport/publicJsonClient";

export async function openLifecycle(page: Page) {
  await openIncidentControls(page, "summary");
  const panel = page.getByRole("region", {
    name: "Incident lifecycle",
    exact: true,
  });
  await expect(panel.getByText(/Current accepted state:/u)).toBeVisible();
  return panel;
}
export async function confirmLifecycle(
  page: Page,
  action: "Close" | "Reopen",
  reason: string,
) {
  const panel = page.getByRole("region", {
    name: "Incident lifecycle",
    exact: true,
  });
  await panel
    .getByRole("textbox", { name: "Reason", exact: true })
    .fill(reason);
  await panel
    .getByRole("button", { name: `${action} incident`, exact: true })
    .click();
  await panel
    .getByRole("button", { name: `Confirm ${action} incident`, exact: true })
    .click();
}
export async function lifecycleAction(
  page: Page,
  incidentId: string,
  action: "closeIncident" | "reopenIncident",
  body: CloseIncidentRequest,
) {
  return publicHttpOperation({
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    operationID: action,
    pathParameters: { incident_id: incidentId },
    body,
  });
}
export async function currentLifecycle(page: Page, incidentId: string) {
  const result = await publicHttpOperation({
    request: atJsonOrigin(page.request, apiBase),
    operationID: "getIncident",
    pathParameters: { incident_id: incidentId },
  });
  if (!result.ok)
    throw new Error(`Current lifecycle read failed: ${result.status}`);
  return result.payload.data;
}

export async function expectLifecycleControlReachable(
  page: Page,
  control: import("@playwright/test").Locator,
) {
  await control.focus();
  await control.scrollIntoViewIfNeeded();
  await assertLifecycleControlReachable(page, control);
}

export async function assertLifecycleControlReachable(
  page: Page,
  control: import("@playwright/test").Locator,
) {
  const panel = page.getByRole("region", {
    name: "Incident lifecycle",
    exact: true,
  });
  expect(
    await panel.evaluate(
      (node) =>
        node.scrollWidth <= node.clientWidth + 1 &&
        document.documentElement.scrollWidth <=
          document.documentElement.clientWidth + 1,
    ),
  ).toBe(true);
  await expect(control).toBeFocused();
  const box = await control.boundingBox();
  const viewport = page.viewportSize();
  if (!box || !viewport) throw new Error("Lifecycle control viewport missing");
  expect(box.x).toBeGreaterThanOrEqual(-1);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
  expect(box.y).toBeGreaterThanOrEqual(-1);
  expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
}
