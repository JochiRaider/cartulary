import {
  cartularyDesignTokenVars,
  workbookPreferenceTestId,
} from "@cartulary/ui-contracts";
import { expect } from "./fixtures";
import { openIncidentFromLanding } from "./pages/incidentDirectory";
import { withPresentation } from "./support/accessibility/presentation";
import { selectVisualDensity } from "./support/auth/visualPreferences";
import {
  assertLifecycleControlReachable,
  openLifecycle,
} from "./support/incidentLifecycle";
import {
  installMembershipAuditPresentation,
  membershipBrowserEvent,
  openMembershipAudit,
} from "./support/incidentMembershipAudit";
import {
  installMembershipManagementPresentation,
  membershipBrowserMember,
  openMembershipManagement,
} from "./support/incidentMembershipManagement";
import {
  assertMetadataControlReachable,
  installMetadataPresentation,
  openMetadata,
} from "./support/incidentMetadata";
import { reloadVisualApplication } from "./support/runtime/applicationReadiness";
import { responseBarrier } from "./support/runtime/cleanup";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";
import {
  awaitResponseBarrier,
  installOwnedRoute,
} from "./support/transport/requestInterception";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import { createRichTimelineIncident } from "./support/visual/timelineVisualFixture";

test("Capture Membership audit loading inspected stale empty cursor recovery and density.", async ({
  testActorPage: page,
}) => {
  const richIncidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUAL-INCIDENT-CONTROLS"),
    "Service-account investigation",
  );
  const fixture = await installMembershipAuditPresentation(
    page,
    richIncidentId,
  );

  const loading = responseBarrier(page);
  fixture.gateRead(loading.promise);
  await page.setViewportSize({ width: 1280, height: 720 });
  const panel = await openMembershipAudit(page);
  await expect(panel.getByRole("status")).toContainText("Loading");
  await assertViewportVisualRegression(page, "membership-audit-loading", {
    anchor: { locator: panel, align: "start" },
  });
  loading.release();
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  await panel.getByRole("button", { name: /^Inspect / }).click();
  const changes = panel.getByRole("region", {
    name: "Published field changes",
    exact: true,
  });
  const scrollChanges = async () => {
    await expect(changes).toBeVisible();
  };
  await scrollChanges();
  await assertViewportVisualRegression(page, "membership-audit-inspected", {
    anchor: { locator: changes, align: "start" },
  });
  fixture.fail();
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(
    panel.getByRole("button", { name: "Try the read again", exact: true }),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "membership-audit-stale", {
    anchor: { locator: panel, align: "start" },
  });
  fixture.setPage([]);
  await panel
    .getByRole("button", { name: "Try the read again", exact: true })
    .click();
  await expect(
    panel.getByText("No membership audit events yet."),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "membership-audit-empty", {
    anchor: { locator: panel, align: "start" },
  });
  fixture.setPage([membershipBrowserEvent(fixture.incidentId)], "next-page");
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(
    panel.getByRole("button", { name: "Next page", exact: true }),
  ).toBeEnabled();
  fixture.fail("invalid_pagination_request", "cursor_expired");
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await expect(
    panel.getByRole("button", { name: "Reload first page", exact: true }),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "membership-audit-cursor-recovery",
    { anchor: { locator: panel, align: "start" } },
  );
  fixture.setPage([membershipBrowserEvent(fixture.incidentId)]);
  await panel
    .getByRole("button", { name: "Reload first page", exact: true })
    .click();
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  await panel.getByRole("button", { name: /^Inspect / }).click();
  await page.setViewportSize({ width: 390, height: 480 });
  await scrollChanges();
  await assertViewportVisualRegression(
    page,
    "membership-audit-inspected-narrow",
    { anchor: { locator: changes, align: "start" } },
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await scrollChanges();
    await assertViewportVisualRegression(
      page,
      "membership-audit-inspected-zoom",
      {
        anchor: { locator: changes, align: "start", outerScroll: "drawer_end" },
      },
    );
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await scrollChanges();
      await assertViewportVisualRegression(
        page,
        "membership-audit-inspected-spacing",
        { anchor: { locator: changes, align: "start" } },
      );
    },
  );
  await page.setViewportSize({ width: 768, height: 640 });
  for (const density of ["compact", "comfortable"] as const) {
    await selectVisualDensity(page, density);
    await test.step("reload visual application", () =>
      reloadVisualApplication(page));

    await openMembershipAudit(page);
    await expect(panel.getByRole("status")).toContainText("Page 1:");
    await expect(panel).toHaveCSS(
      "font-size",
      cartularyDesignTokenVars[`--ct-density-${density}-fontSize`],
    );
    await expect(panel.getByRole("listitem")).toHaveCSS(
      "padding",
      cartularyDesignTokenVars[`--ct-density-${density}-cellPadding`],
    );
    await test.info().attach(`membership-audit-density-${density}`, {
      body: JSON.stringify(
        await panel.evaluate((element) => ({
          fontSize: getComputedStyle(element).fontSize,
          lineHeight: getComputedStyle(element).lineHeight,
        })),
      ),
      contentType: "application/json",
    });
    await panel.getByRole("button", { name: /^Inspect / }).click();
    await scrollChanges();
    await assertViewportVisualRegression(page, `membership-audit-${density}`, {
      anchor: { locator: changes, align: "start" },
    });
  }
});

test("Capture Membership management drafts pending uncertainty confirmed recovery and density.", async ({
  testActorPage: page,
}) => {
  const richIncidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUAL-INCIDENT-CONTROLS"),
    "Service-account investigation",
  );
  const fixture = await installMembershipManagementPresentation(
    page,
    richIncidentId,
  );

  const loading = responseBarrier(page);
  fixture.gateRead(loading.promise);
  await page.setViewportSize({ width: 1280, height: 720 });
  const panel = await openMembershipManagement(page);
  await expect(
    panel.getByText("Loading memberships…", { exact: true }),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "membership-management-loading", {
    anchor: { locator: panel, align: "start" },
  });
  loading.release();
  await expect(
    panel.getByText("Response analyst", { exact: true }),
  ).toBeVisible();
  await panel
    .getByRole("button", { name: /^Change role for Response analyst/u })
    .click();
  const role = panel.getByRole("combobox", {
    name: /^Role for Response analyst/u,
  });
  await role.selectOption("reviewer");
  await assertViewportVisualRegression(page, "membership-management-role", {
    anchor: { locator: panel, align: "start" },
  });
  const pending = responseBarrier(page);
  fixture.gateWrite(pending.promise);
  fixture.mutation(503);
  await panel.getByRole("button", { name: "Save role", exact: true }).click();
  await expect(
    panel.getByText(/Saving the reviewed membership action/u),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "membership-management-pending", {
    anchor: { locator: panel, align: "start" },
  });
  pending.release();
  await expect(panel.getByText(/has an uncertain outcome/u)).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "membership-management-uncertain",
    { anchor: { locator: panel, align: "start" } },
  );
  fixture.setPage([
    membershipBrowserMember(fixture.incidentId, {
      role: "reviewer",
      membership_version: 2,
    }),
  ]);
  await panel
    .getByRole("button", { name: "Observe current membership", exact: true })
    .click();
  await expect(
    panel.getByText(/Observed role: reviewer; version: 2/u),
  ).toBeVisible();
  await panel
    .getByRole("button", { name: "Review observed membership", exact: true })
    .first()
    .click();
  fixture.mutation(
    200,
    membershipBrowserMember(fixture.incidentId, {
      role: "reviewer",
      membership_version: 2,
    }),
  );
  fixture.failRead();
  await panel.getByRole("button", { name: "Save role", exact: true }).click();
  await expect(
    panel.getByRole("button", { name: "Retry member refresh", exact: true }),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "membership-management-confirmed-refresh-failure",
    { anchor: { locator: panel, align: "start" } },
  );
  fixture.setPage([membershipBrowserMember(fixture.incidentId)]);
  await panel
    .getByRole("button", { name: "Retry member refresh", exact: true })
    .click();
  await expect(panel.getByText(/Page 1:/u)).toBeVisible();
  await panel
    .getByRole("button", {
      name: /^Remove incident access for Response analyst/u,
    })
    .click();
  const review = panel.getByRole("form");
  await page.setViewportSize({ width: 390, height: 480 });
  await assertViewportVisualRegression(
    page,
    "membership-management-removal-narrow",
    { anchor: { locator: review, align: "start" } },
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await assertViewportVisualRegression(
      page,
      "membership-management-removal-zoom",
      {
        anchor: {
          locator: review.getByRole("button", {
            name: "Confirm removal",
            exact: true,
          }),
          align: "center",
          outerScroll: "drawer_end",
        },
      },
    );
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await assertViewportVisualRegression(
        page,
        "membership-management-removal-spacing",
        { anchor: { locator: review, align: "start" } },
      );
    },
  );
  await page.setViewportSize({ width: 768, height: 640 });
  for (const density of ["compact", "comfortable"] as const) {
    await selectVisualDensity(page, density);
    await test.step("reload visual application", () =>
      reloadVisualApplication(page));

    await openMembershipManagement(page);
    await expect(panel.getByText(/Page 1:/u)).toBeVisible();
    await expect(panel).toHaveCSS(
      "font-size",
      cartularyDesignTokenVars[`--ct-density-${density}-fontSize`],
    );
    await expect(panel.getByRole("article")).toHaveCSS(
      "padding",
      cartularyDesignTokenVars[`--ct-density-${density}-cellPadding`],
    );
    await panel
      .getByRole("button", { name: /^Change role for Response analyst/u })
      .click();
    await assertViewportVisualRegression(
      page,
      `membership-management-${density}`,
      { anchor: { locator: review, align: "start" } },
    );
  }
});

test("Capture Metadata editing loading dirty conflict uncertainty confirmation responsive and density.", async ({
  testActorPage: page,
}) => {
  const richIncidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUAL-INCIDENT-CONTROLS"),
    "Service-account investigation",
  );
  await openIncidentFromLanding(page, richIncidentId);
  const fixture = await installMetadataPresentation(page, richIncidentId);

  await page.setViewportSize({ width: 1280, height: 720 });
  const gate = responseBarrier(page);
  fixture.gateRead(gate.promise);
  const panel = await openMetadata(page);
  await expect(
    panel.getByText("Refreshing promoted fields…", { exact: true }),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "metadata-loading", {
    anchor: { locator: panel, align: "start" },
  });
  gate.release();
  const severity = panel.getByLabel("Severity", { exact: true });
  await severity.fill("critical");
  await assertViewportVisualRegression(page, "metadata-dirty", {
    anchor: { locator: panel, align: "start" },
  });
  const pending = responseBarrier(page);
  fixture.gateWrite(pending.promise);
  fixture.mutation(503);
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect(
    panel.getByText("Saving promoted incident fields…", { exact: true }),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "metadata-saving", {
    anchor: { locator: panel, align: "start" },
  });
  pending.release();
  await expect(panel).toHaveAttribute("data-metadata-operation", "uncertain");
  const review = panel.getByRole("region", {
    name: "Review promoted field changes",
  });
  await assertViewportVisualRegression(page, "metadata-uncertain", {
    anchor: { locator: review, align: "start" },
  });
  await panel
    .getByRole("button", { name: "Observe current values", exact: true })
    .click();
  await expect(
    panel.getByText("Incident controls synced.", { exact: true }),
  ).toBeVisible();
  await panel
    .getByRole("button", { name: "Use this version", exact: true })
    .click();
  fixture.observe({ severity: "concurrent", incident_version: 2 });
  fixture.mutation(409, {}, "incident_version_conflict");
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect(review.getByText("concurrent", { exact: true })).toBeVisible();
  await assertViewportVisualRegression(page, "metadata-conflict", {
    anchor: { locator: review, align: "start" },
  });
  await page.setViewportSize({ width: 390, height: 480 });
  const useVersion = panel.getByRole("button", {
    name: "Use this version",
    exact: true,
  });
  await assertViewportVisualRegression(page, "metadata-review-narrow", {
    anchor: { locator: useVersion, align: "center", focus: true },
  });
  await assertMetadataControlReachable(page, useVersion);
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await assertViewportVisualRegression(page, "metadata-review-zoom", {
      anchor: {
        locator: useVersion,
        align: "center",
        focus: true,
        outerScroll: "drawer_end",
      },
    });
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await assertViewportVisualRegression(page, "metadata-review-spacing", {
        anchor: { locator: review, align: "start" },
      });
    },
  );
  await page.setViewportSize({ width: 768, height: 640 });
  await panel
    .getByRole("button", { name: "Use this version", exact: true })
    .click();
  fixture.mutation(200, { severity: "critical", incident_version: 3 });
  fixture.failRead();
  await panel
    .getByRole("button", { name: "Save promoted fields", exact: true })
    .click();
  await expect(panel.getByText(/Metadata refresh failed/u)).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "metadata-confirmed-refresh-failure",
    { anchor: { locator: panel, align: "start" } },
  );
  fixture.observe({
    severity: "critical",
    incident_version: 4,
    status: "closed",
    closed_at: "2026-08-01T00:00:00Z",
  });
  await panel
    .getByRole("button", { name: "Check access and refresh", exact: true })
    .click();
  await expect(panel.getByText(/This incident is closed/u)).toBeVisible();
  await assertViewportVisualRegression(page, "metadata-closed", {
    anchor: { locator: panel, align: "start" },
  });
  fixture.observe({ status: "active", closed_at: null, incident_version: 5 });
  for (const density of ["compact", "comfortable"] as const) {
    await selectVisualDensity(page, density);
    await test.step("reload visual application", () =>
      reloadVisualApplication(page));

    await page.setViewportSize({ width: 768, height: 640 });
    await openMetadata(page);
    await expect(severity).toBeVisible();
    await expect(panel).toHaveCSS(
      "font-size",
      cartularyDesignTokenVars[`--ct-density-${density}-fontSize`],
    );
    await expect(panel.locator("[data-metadata-field]").first()).toHaveCSS(
      "padding",
      cartularyDesignTokenVars[`--ct-density-${density}-cellPadding`],
    );
    await severity.fill("critical exact input");
    await assertViewportVisualRegression(page, `metadata-${density}`, {
      anchor: { locator: panel, align: "start" },
    });
  }
});

test("Capture Lifecycle review pending exact recovery confirmation responsive and density.", async ({
  testActorPage: page,
}) => {
  const inspectionReady = async () => {
    // Independent observations can change the drawer height. Settle them before
    // focus and scroll establish this lifecycle fixture's capture position.
    await expect(
      page.getByText("Incident controls synced.", { exact: true }),
    ).toBeVisible();
    for (const kind of ["home", "default"] as const) {
      await expect(
        page.getByTestId(workbookPreferenceTestId(kind, "read")),
      ).toHaveText("");
    }
  };
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("LC-VISUAL"),
    "Incident lifecycle review",
  );
  await openIncidentFromLanding(page, incidentId);
  const panel = await openLifecycle(page);
  await inspectionReady();
  await page.setViewportSize({ width: 1280, height: 720 });
  const capture = async (name: string, anchor = panel) => {
    const focus = anchor !== panel && (await anchor.isEnabled());
    await assertViewportVisualRegression(page, name, {
      anchor: {
        locator: anchor,
        align: anchor === panel ? "start" : "center",
        focus,
      },
      ready: inspectionReady,
    });
    if (focus) await assertLifecycleControlReachable(page, anchor);
    await test.info().attach(`${name}-review`, {
      body: await page.screenshot({
        animations: "disabled",
        caret: "hide",
        fullPage: false,
      }),
      contentType: "image/png",
    });
  };
  const reason = panel.getByRole("textbox", { name: "Reason", exact: true });
  await reason.fill(
    "Containment reviewed.\nRetain local investigation drafts.",
  );
  await capture("lifecycle-reason");
  await panel
    .getByRole("button", { name: "Close incident", exact: true })
    .click();
  const confirm = panel.getByRole("button", {
    name: "Confirm Close incident",
    exact: true,
  });

  await capture("lifecycle-review", confirm);
  await page.setViewportSize({ width: 390, height: 480 });

  await capture("lifecycle-review-narrow", confirm);
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await capture("lifecycle-review-zoom", confirm);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });

      await capture("lifecycle-review-spacing", confirm);
    },
  );
  await page.setViewportSize({ width: 1024, height: 720 });
  const pending = responseBarrier(page);
  let writes = 0;
  let failedReads = false;
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/close`,
    async (route) => {
      ++writes;
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      if (writes === 1) {
        if (!(await awaitResponseBarrier(route, pending.promise))) return;
        await route.abort("failed");
      } else await route.fulfill({ response });
    },
  );
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}`,
    async (route) => {
      if (failedReads && route.request().method() === "GET")
        await route.fulfill({
          status: 503,
          json: {
            error: { code: "internal_error" },
            meta: { request_id: "lc-visual-read-failure" },
          },
        });
      else await route.continue();
    },
  );
  await confirm.click();
  await expect.poll(() => writes).toBe(1);
  await expect(panel.getByText(/Close request sent/u)).toBeVisible();
  await capture("lifecycle-pending", confirm);
  pending.release();
  await expect(panel.getByText(/Close result is uncertain/u)).toBeVisible();
  const replay = panel.getByRole("button", {
    name: "Replay original action",
    exact: true,
  });

  await capture("lifecycle-uncertain", replay);
  failedReads = true;
  await replay.click();
  await expect(
    panel.getByText(
      /The action is confirmed, but current state could not be refreshed/u,
    ),
  ).toBeVisible();
  const refresh = panel.getByRole("button", {
    name: "Refresh current incident",
    exact: true,
  });

  await capture("lifecycle-confirmed-refresh-failure", refresh);
  failedReads = false;
  await refresh.click();
  await expect(
    panel.getByText(/Current accepted state: Closed, read-only · Version 2/u),
  ).toBeVisible();
  await capture("lifecycle-closed", panel);
  for (const density of ["compact", "comfortable"] as const) {
    await selectVisualDensity(page, density);
    await test.step("reload visual application", () =>
      reloadVisualApplication(page));
    await openLifecycle(page);
    await inspectionReady();
    await page.setViewportSize({ width: 768, height: 640 });
    await expect(panel).toHaveCSS(
      "font-size",
      cartularyDesignTokenVars[`--ct-density-${density}-fontSize`],
    );
    await expect(panel).toHaveCSS(
      "padding",
      cartularyDesignTokenVars[`--ct-density-${density}-cellPadding`],
    );
    await reason.fill("Review new evidence before reopening.");
    await panel
      .getByRole("button", { name: "Reopen incident", exact: true })
      .click();
    const reopen = panel.getByRole("button", {
      name: "Confirm Reopen incident",
      exact: true,
    });

    await capture(`lifecycle-${density}`, reopen);
  }
});
