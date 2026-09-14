import type { GetIncidentResponse } from "@cartulary/protocol-ts/http";
import { expect, type Locator, type Page } from "@playwright/test";
import { openIncidentControls } from "../pages/deploymentAdministration";
import { openIncidentFromLanding } from "../pages/incidentDirectory";
import { rewriteSessionPresentation } from "./auth/sessionPresentation";
import { createIncident } from "./incidents/fixtures";
import { apiBase } from "./runtime/configuration";
import { uniqueIncidentKey } from "./runtime/fixtureIdentity";
import { publicHttpOperation } from "./transport/publicHttpOperationClient";
import { atJsonOrigin } from "./transport/publicJsonClient";
export async function openMetadata(page: Page) {
  await openIncidentControls(page, "incident-fields");
  return page.getByRole("region", {
    name: "Promoted incident fields",
    exact: true,
  });
}
/** Service-backed workbook with deterministic transport faults restricted to metadata/current access. */
export async function installMetadataPresentation(
  page: Page,
  existingIncidentId?: string,
) {
  const incidentId =
    existingIncidentId ??
    (await createIncident(
      page,
      uniqueIncidentKey("ME"),
      "Incident metadata review",
    ));
  const result = await publicHttpOperation({
    request: atJsonOrigin(page.request, apiBase),
    operationID: "getIncident",
    pathParameters: { incident_id: incidentId },
  });
  if (!result.ok) throw new Error("Metadata fixture read failed");
  let resource: GetIncidentResponse["data"] = {
    ...result.payload.data,
    incident_key: "IR-METADATA",
    title: "Incident metadata review",
    description: "Initial incident description",
    severity: "high",
    tlp: "TLP:AMBER",
    current_phase: "triage",
    primary_external_case_ref: "CASE-PRIMARY",
  };
  let readStatus = 200;
  let writeStatus = 200;
  let writeCode = "service_unavailable";
  let writeResource: GetIncidentResponse["data"] = {
    ...resource,
    severity: "critical",
    incident_version: 2,
  };
  let readGate: Promise<void> | null = null;
  let writeGate: Promise<void> | null = null;
  let access:
    | "admin"
    | "reviewer"
    | "editor"
    | "viewer"
    | "hidden"
    | "session"
    | "unavailable" = "admin";
  const writes: Record<string, unknown>[] = [];
  const reads: string[] = [];
  await page.route("**/api/v1/auth/session", async (route) => {
    if (access === "session" || access === "unavailable") {
      await route.fulfill({
        status: access === "session" ? 401 : 503,
        json: {
          error: {
            code:
              access === "session" ? "session_required" : "service_unavailable",
          },
          meta: { request_id: "metadata-access" },
        },
      });
      return;
    }
    const requestAccess = access;
    await rewriteSessionPresentation(route, (session) => ({
      ...session,
      display_name: "Metadata operator",
      memberships: session.memberships.flatMap((m) =>
        m.incident_id !== incidentId
          ? [m]
          : requestAccess === "hidden"
            ? []
            : [{ ...m, role: requestAccess }],
      ),
    }));
  });
  await page.route(`**/api/v1/incidents/${incidentId}`, async (route) => {
    if (route.request().method() === "GET") {
      reads.push(route.request().url());
      const captured = { resource, status: readStatus, gate: readGate };
      readGate = null;
      if (captured.gate !== null) await captured.gate;
      await route
        .fulfill({
          status: captured.status,
          json:
            captured.status === 200
              ? {
                  data: captured.resource,
                  meta: { request_id: "metadata-read" },
                }
              : {
                  error: { code: "service_unavailable" },
                  meta: { request_id: "metadata-read-failed" },
                },
        })
        .catch(() => {});
      return;
    }
    if (route.request().method() !== "PATCH") {
      await route.continue();
      return;
    }
    writes.push(route.request().postDataJSON());
    const captured = {
      resource: writeResource,
      status: writeStatus,
      code: writeCode,
      gate: writeGate,
    };
    writeGate = null;
    if (captured.gate !== null) await captured.gate;
    await route
      .fulfill({
        status: captured.status,
        json:
          captured.status === 200
            ? {
                data: captured.resource,
                meta: { request_id: "metadata-write" },
              }
            : {
                error: {
                  code: captured.code,
                  message: "Private diagnostic must not appear",
                },
                meta: { request_id: "metadata-rejection" },
              },
      })
      .catch(() => {});
  });
  if (!existingIncidentId) await openIncidentFromLanding(page, incidentId);
  return {
    incidentId,
    reads,
    writes,
    resource: () => resource,
    observe: (next: Partial<typeof resource>) => {
      resource = { ...resource, ...next };
      readStatus = 200;
    },
    failRead: () => {
      readStatus = 503;
    },
    mutation: (
      status: number,
      next: Partial<typeof resource> = {},
      code = "service_unavailable",
    ) => {
      writeStatus = status;
      writeResource = { ...resource, ...next };
      writeCode = code;
    },
    gateRead: (gate: Promise<void>) => {
      readGate = gate;
    },
    gateWrite: (gate: Promise<void>) => {
      writeGate = gate;
    },
    access: (next: typeof access) => {
      access = next;
    },
  };
}
export async function expectMetadataControlReachable(
  page: Page,
  control: Locator,
) {
  await control.focus();
  await control.scrollIntoViewIfNeeded();
  await assertMetadataControlReachable(page, control);
}

export async function assertMetadataControlReachable(
  page: Page,
  control: Locator,
) {
  const panel = page.getByRole("region", {
    name: "Promoted incident fields",
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
  if (!box || !viewport) throw new Error("Metadata control viewport missing");
  expect(box.x).toBeGreaterThanOrEqual(-1);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
  expect(box.y).toBeGreaterThanOrEqual(-1);
  expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
}
