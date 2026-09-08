import { incidentLandingTestId } from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { openIncidentControls } from "./pages/deploymentAdministration";
import { openIncidentFromLanding } from "./pages/incidentDirectory";
import { auditBrowserBarrier } from "./support/administrativeAudit";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  installMembershipManagementPresentation,
  openMembershipManagement,
} from "./support/incidentMembershipManagement";
import {
  installMetadataPresentation,
  openMetadata,
} from "./support/incidentMetadata";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";

test("metadata edits each live field sparsely clears normalizes and reviews a real version conflict", async ({
  workerAdminPage: page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ME-LIVE"),
    "Live promoted metadata",
  );
  await openIncidentFromLanding(page, incidentId);
  const panel = await openMetadata(page);
  await expect(panel.getByLabel("Description", { exact: true })).toBeVisible();
  const writes: Record<string, unknown>[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/incidents/${incidentId}`)
    )
      writes.push(request.postDataJSON());
  });
  const fields = [
    [
      "description",
      "Description",
      "  e\u0301\nSecond line  ",
      "é\nSecond line",
    ],
    ["severity", "Severity", "urgent text", "urgent text"],
    ["tlp", "TLP", "TLP:AMBER+STRICT", "TLP:AMBER+STRICT"],
    [
      "current_phase",
      "Current phase",
      "containment custom",
      "containment custom",
    ],
    [
      "primary_external_case_ref",
      "Primary external case",
      "CASE-😀",
      "CASE-😀",
    ],
  ] as const;
  let version = 1;
  for (const [field, label, raw, saved] of fields) {
    const control = panel.getByLabel(label, { exact: true });
    if (field === "tlp") await control.selectOption(raw);
    else await control.fill(raw);
    const count = writes.length;
    await panel
      .getByRole("button", { name: "Save promoted fields", exact: true })
      .click();
    await expect(panel).toHaveAttribute("data-metadata-operation", "confirmed");
    await expect(control).toHaveValue(saved);
    await expect.poll(() => writes.length).toBe(count + 1);
    expect(writes.at(-1)).toEqual({
      base_incident_version: version++,
      [field]: raw,
    });
  }
  for (const [field, label] of fields) {
    const control = panel.getByLabel(label, { exact: true });
    if (field === "tlp") await control.selectOption("");
    else await control.fill("");
    await panel
      .getByRole("button", { name: "Save promoted fields", exact: true })
      .click();
    await expect(panel).toHaveAttribute("data-metadata-operation", "confirmed");
    await expect(control).toHaveValue("");
    expect(writes.at(-1)).toEqual({
      base_incident_version: version++,
      [field]: null,
    });
  }
  await panel.getByLabel("Severity", { exact: true }).fill(" \u2003 ");
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue("");
  // A normalized no-op acknowledges the same version.
  await expect(
    panel.getByText(
      `Version ${version}. Only changed promoted fields are submitted.`,
    ),
  ).toBeVisible();
  await panel.getByLabel("Severity", { exact: true }).fill("  intended  ");
  const competing = await publicHttpOperation({
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    operationID: "patchIncident",
    pathParameters: { incident_id: incidentId },
    body: { base_incident_version: version, severity: "concurrent" },
  });
  expect(competing.ok).toBe(true);
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect(panel).toHaveAttribute("data-metadata-operation", "conflicted");
  const review = panel.getByRole("region", {
    name: "Review promoted field changes",
  });
  await expect(review.getByText("concurrent", { exact: true })).toBeVisible();
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
    "  intended  ",
  );
  const count = writes.length;
  await panel
    .getByRole("button", { name: "Use this version", exact: true })
    .click();
  expect(writes).toHaveLength(count);
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
    "intended",
  );
  await openIncidentControls(page);
  await expect(page.getByText("intended", { exact: true })).toBeVisible();
  await page
    .getByLabel("Reason", { exact: true })
    .fill("Metadata journey complete");
  await page
    .getByRole("button", { name: "Close incident", exact: true })
    .click();
  await page
    .getByRole("button", { name: "Confirm Close incident", exact: true })
    .click();
  await expect(
    page.getByText("Close confirmed.", { exact: true }),
  ).toBeVisible();
  await openMetadata(page);
  await expect(panel.getByText(/This incident is closed/u)).toBeVisible();
  await expect(panel.getByText("intended", { exact: true })).toBeVisible();
});

test("metadata retains newer input and distinguishes confirmed failed reads from uncertain observation", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installMetadataPresentation(page);
  const panel = await openMetadata(page);
  const severity = panel.getByLabel("Severity", { exact: true });
  await severity.fill("critical");
  const gate = auditBrowserBarrier();
  fixture.gateWrite(gate.promise);
  fixture.mutation(200, { severity: "critical", incident_version: 2 });
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect.poll(() => fixture.writes.length).toBe(1);
  await severity.fill("newer exact  ");
  fixture.failRead();
  gate.release();
  await expect(
    panel.getByText("Saved promoted incident fields.", { exact: true }),
  ).toBeVisible();
  await expect(panel.getByText(/Metadata refresh failed/u)).toBeVisible();
  await expect(severity).toHaveValue("newer exact  ");
  fixture.observe({ severity: "critical", incident_version: 2 });
  await panel
    .getByRole("button", { name: "Check access and refresh", exact: true })
    .click();
  await expect(
    panel.getByText("Incident controls synced.", { exact: true }),
  ).toBeVisible();
  expect(fixture.writes).toHaveLength(1);
  fixture.mutation(503);
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect(panel).toHaveAttribute("data-metadata-operation", "uncertain");
  fixture.observe({ severity: "newer exact  ", incident_version: 3 });
  await panel
    .getByRole("button", { name: "Observe current values", exact: true })
    .click();
  await expect(panel.getByText(/The save result is uncertain/u)).toBeVisible();
  expect(fixture.writes).toHaveLength(2);
  await expect(
    page.getByText("Private diagnostic must not appear"),
  ).toHaveCount(0);
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await openMetadata(page);
  await expect(panel).toHaveAttribute("data-metadata-operation", "uncertain");
  await panel
    .getByRole("button", { name: "Use this version", exact: true })
    .click();
  await expect(panel.getByText(/remains unconfirmed/u)).toBeVisible();
  expect(fixture.writes).toHaveLength(2);
});

test("metadata and membership retained work receive sequential departure review including Stay and history", async ({
  workerAdminPage: page,
}) => {
  const membership = await installMembershipManagementPresentation(page);
  const fixture = await installMetadataPresentation(
    page,
    membership.incidentId,
  );
  const members = await openMembershipManagement(page);
  await members
    .getByRole("button", { name: /^Change role for Response analyst/u })
    .click();
  await members
    .getByRole("combobox", { name: /^Role for Response analyst/u })
    .selectOption("reviewer");
  const panel = await openMetadata(page);
  await panel.getByLabel("Severity", { exact: true }).fill("retained metadata");
  const leave = async () => {
    await page.getByLabel("Account and application navigation").click();
    await page
      .getByRole("menuitem", { name: "Incidents", exact: true })
      .click();
  };
  await leave();
  await expect(
    page.getByRole("dialog", { name: "Leave membership work?" }),
  ).toBeVisible();
  await expect(
    page.getByRole("dialog", { name: /^Leave .* work\?$/u }),
  ).toHaveCount(1);
  await page
    .getByRole("button", { name: "Leave and forget recovery", exact: true })
    .click();
  await expect(
    page.getByRole("dialog", { name: "Leave metadata work?" }),
  ).toBeVisible();
  await expect(
    page.getByRole("dialog", { name: /^Leave .* work\?$/u }),
  ).toHaveCount(1);
  await page.getByRole("button", { name: "Stay", exact: true }).click();
  await openMetadata(page);
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
    "retained metadata",
  );
  await page.evaluate(() => window.history.back());
  await expect(
    page.getByRole("dialog", { name: "Leave metadata work?" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Stay", exact: true }).click();
  await leave();
  await expect(
    page.getByRole("dialog", { name: "Leave membership work?" }),
  ).toHaveCount(0);
  await page
    .getByRole("button", { name: "Leave and forget recovery", exact: true })
    .click();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  expect(fixture.writes).toHaveLength(0);
  expect(membership.writes).toHaveLength(0);
});

test("metadata uses current roles distinguishes closure and clears protected state after session or incident loss", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installMetadataPresentation(page);
  const panel = await openMetadata(page);
  await panel.getByLabel("Severity", { exact: true }).fill("retained");
  for (const role of ["viewer", "editor", "reviewer", "admin"] as const) {
    fixture.access(role);
    await panel.getByRole("button", { name: "Refresh", exact: true }).click();
    if (role === "viewer" || role === "editor") {
      await expect(panel.getByText(/reviewer or admin role/u)).toBeVisible();
      await expect(
        panel.getByText("Initial incident description", { exact: true }),
      ).toBeVisible();
    } else {
      await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
        "retained",
      );
      await expect(
        panel.getByRole("button", {
          name: "Save promoted fields",
          exact: true,
        }),
      ).toHaveAttribute("aria-disabled", "true");
    }
  }
  fixture.observe({
    status: "closed",
    closed_at: "2026-08-01T00:00:00Z",
    incident_version: 2,
  });
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(panel.getByText(/This incident is closed/u)).toBeVisible();
  fixture.observe({ status: "active", closed_at: null, incident_version: 3 });
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
    "retained",
  );
  fixture.access("unavailable");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(
    panel.getByText(/Current incident access could not be checked/u),
  ).toBeVisible();
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveCount(0);
  fixture.access("admin");
  await panel
    .getByRole("button", { name: "Check access and refresh", exact: true })
    .click();
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
    "retained",
  );
  fixture.access("hidden");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(panel).toHaveCount(0);
  fixture.access("admin");
  await openIncidentFromLanding(page, fixture.incidentId);
  await openMetadata(page);
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
    "high",
  );
  fixture.access("session");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(panel).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "Sign in", exact: true }),
  ).toBeVisible();
});

test("metadata pending transport survives drawer return and cannot publish after deliberate incident departure", async ({
  workerAdminPage: page,
}) => {
  const fixture = await installMetadataPresentation(page);
  const panel = await openMetadata(page);
  const gate = auditBrowserBarrier();
  fixture.gateWrite(gate.promise);
  fixture.mutation(200, { severity: "critical", incident_version: 2 });
  await panel.getByLabel("Severity", { exact: true }).fill("critical");
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect.poll(() => fixture.writes.length).toBe(1);
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await openMetadata(page);
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
    "critical",
  );
  await panel
    .getByLabel("Severity", { exact: true })
    .fill("newer retained input");
  await page.getByLabel("Account and application navigation").click();
  await page.getByRole("menuitem", { name: "Incidents", exact: true }).click();
  const departure = page.getByRole("dialog", { name: "Leave metadata work?" });
  await expect(
    departure.getByText(/cannot cancel a pending server operation/u),
  ).toBeVisible();
  await departure
    .getByRole("button", { name: "Leave and forget recovery", exact: true })
    .click();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  const nextId = await createIncident(
    page,
    uniqueIncidentKey("ME-NEXT"),
    "Next metadata incident",
  );
  await openIncidentFromLanding(page, nextId);
  await openMetadata(page);
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue("");
  gate.release();
  await expect(panel).toHaveAttribute("data-metadata-operation", "idle");
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue("");
  expect(fixture.writes).toHaveLength(1);
});
