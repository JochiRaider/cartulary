import { expect, test } from "./fixtures";
import { openIncidentFromLanding } from "./pages/incidentDirectory";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import {
  expectLifecycleControlReachable,
  openLifecycle,
} from "./support/incidentLifecycle";
import {
  expectMembershipControlReachable,
  installMembershipAuditPresentation,
  membershipBrowserEvent,
  openMembershipAudit,
} from "./support/incidentMembershipAudit";
import {
  expectMembershipManagementControlReachable,
  installMembershipManagementPresentation,
  openMembershipManagement,
} from "./support/incidentMembershipManagement";
import {
  expectMetadataControlReachable,
  installMetadataPresentation,
  openMetadata,
} from "./support/incidentMetadata";
import { createIncident } from "./support/incidents/fixtures";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";
import { installOwnedRoute } from "./support/transport/requestInterception";

test("a11y.membership-audit exact filters inspection paging recovery and drawer focus remain reachable", async ({
  testActorPage: page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const fixture = await installMembershipAuditPresentation(page);
  fixture.setPage([membershipBrowserEvent(fixture.incidentId)], "next-page");
  const panel = await openMembershipAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  await panel.getByLabel("Target ID", { exact: true }).fill("exact-target");
  await panel.getByLabel("Target ID", { exact: true }).press("Enter");
  const kind = panel.getByLabel("Target kind", { exact: true });
  await expect(kind).toBeFocused();
  await expect(kind).toHaveAccessibleDescription(/Choose/);
  await expectVisibleFocus(kind);
  await panel
    .getByRole("button", { name: "Clear filters", exact: true })
    .click();
  await panel.getByRole("button", { name: /^Inspect / }).focus();
  await page.keyboard.press("Enter");
  const hide = panel.getByRole("button", { name: /^Hide / });
  await expect(hide).toHaveAttribute("aria-expanded", "true");
  await expect(
    panel.getByRole("region", { name: /^Details for / }),
  ).toContainText("Redacted");
  await expect(panel.locator("script")).toHaveCount(0);
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
    { width: 640, height: 480 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [
      hide,
      panel.getByRole("button", { name: "Next page", exact: true }),
      panel.getByRole("button", { name: "Apply filters", exact: true }),
      panel.getByRole("button", { name: "Refresh", exact: true }),
      page.getByRole("button", {
        name: "Close incident controls",
        exact: true,
      }),
    ]) {
      await expectMembershipControlReachable(page, control);
      await expectVisibleFocus(control);
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectMembershipControlReachable(page, hide);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await expectMembershipControlReachable(page, hide);
      await expectAllInteractiveControlsNamed(page);
      fixture.fail();
      await panel.getByRole("button", { name: "Refresh", exact: true }).click();
      const retry = panel.getByRole("button", {
        name: "Try the read again",
        exact: true,
      });
      await expect(retry).toBeVisible();
      await expectMembershipControlReachable(page, retry);
      await expectVisibleFocus(retry);
      await testInfo.attach("membership-audit-accessibility-tree", {
        body: await panel.ariaSnapshot(),
        contentType: "text/plain",
      });
      fixture.setPage([]);
      await page.keyboard.press("Enter");
      await expect(
        panel.getByText("No membership audit events yet."),
      ).toBeVisible();
      await expect(
        panel.getByRole("button", { name: "Refresh", exact: true }),
      ).toBeFocused();
    },
  );
  await page.keyboard.press("Escape");
  await expect(
    page.getByLabel("Account and application navigation"),
  ).toBeFocused();
});

test("a11y.membership-management draft review keyboard recovery and removal focus remain reachable", async ({
  testActorPage: page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const fixture = await installMembershipManagementPresentation(page);
  const panel = await openMembershipManagement(page);
  await expect(
    panel.getByText("Response analyst", { exact: true }),
  ).toBeVisible();
  const change = panel.getByRole("button", {
    name: /^Change role for Response analyst/u,
  });
  await change.focus();
  await page.keyboard.press("Enter");
  const role = panel.getByRole("combobox", {
    name: /^Role for Response analyst/u,
  });
  await expect(role).toBeFocused();
  await expectVisibleFocus(role);
  await role.selectOption("reviewer");
  await panel
    .getByRole("button", { name: "Add existing account", exact: true })
    .click();
  await expect(
    page.getByRole("dialog", { name: "Discard this membership draft?" }),
  ).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(
    page.getByRole("dialog", { name: "Discard this membership draft?" }),
  ).toHaveCount(0);
  await expect(panel).toBeVisible();
  await expect(role).toHaveValue("reviewer");
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [
      role,
      panel.getByRole("button", { name: "Save role", exact: true }),
      panel.getByRole("button", { name: "Cancel", exact: true }),
      page.getByRole("button", {
        name: "Close incident controls",
        exact: true,
      }),
    ]) {
      await expectMembershipManagementControlReachable(page, control);
      await expectVisibleFocus(control);
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectMembershipManagementControlReachable(page, role);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await expectMembershipManagementControlReachable(
        page,
        panel.getByRole("button", { name: "Save role", exact: true }),
      );
      await expectAllInteractiveControlsNamed(page);
      await testInfo.attach("membership-management-accessibility-tree", {
        body: await panel.ariaSnapshot(),
        contentType: "text/plain",
      });
    },
  );
  await panel.getByRole("button", { name: "Cancel", exact: true }).click();
  await page
    .getByRole("button", { name: "Discard draft", exact: true })
    .click();
  await panel
    .getByRole("button", {
      name: /^Remove incident access for Response analyst/u,
    })
    .focus();
  await page.keyboard.press("Enter");
  await expect(panel.getByRole("form").getByRole("heading")).toBeFocused();
  fixture.mutation(204);
  fixture.setPage([]);
  const confirm = panel.getByRole("button", {
    name: "Confirm removal",
    exact: true,
  });
  await confirm.focus();
  await page.keyboard.press("Enter");
  await expect(
    panel.getByText("Incident membership removed.", { exact: true }),
  ).toBeVisible();
  await expect(panel.getByRole("heading")).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(
    page.getByLabel("Account and application navigation"),
  ).toBeFocused();
});

test("a11y.metadata native inputs inline review discard and drawer focus remain reachable", async ({
  testActorPage: page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const fixture = await installMetadataPresentation(page);
  const panel = await openMetadata(page);
  const description = panel.getByLabel("Description", { exact: true });
  await description.fill(`Long description ${"incident context ".repeat(60)}`);
  await description.press("End");
  await page.keyboard.press("Enter");
  await page.keyboard.type("Second line");
  await expect(description).toHaveValue(/\nSecond line$/u);
  expect(fixture.writes).toHaveLength(0);
  const save = panel.getByRole("button", {
    name: "Save promoted fields",
    exact: true,
  });
  const discard = panel.getByRole("button", {
    name: "Discard changes",
    exact: true,
  });
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [
      description,
      panel.getByLabel("TLP", { exact: true }),
      save,
      discard,
      page.getByRole("button", {
        name: "Close incident controls",
        exact: true,
      }),
    ]) {
      await expectMetadataControlReachable(page, control);
      await expectVisibleFocus(control);
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectMetadataControlReachable(page, save);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 390, height: 480 });
      await expectMetadataControlReachable(page, discard);
      await page.keyboard.press("Enter");
      await expect(description).toHaveValue("Initial incident description");
    },
  );
  await page.setViewportSize({ width: 768, height: 640 });
  await panel.getByLabel("Severity", { exact: true }).fill("intended");
  fixture.observe({ severity: "current", incident_version: 2 });
  fixture.mutation(409, {}, "incident_version_conflict");
  await save.focus();
  await page.keyboard.press("Enter");
  const useVersion = panel.getByRole("button", {
    name: "Use this version",
    exact: true,
  });
  await expectMetadataControlReachable(page, useVersion);
  await expectVisibleFocus(useVersion);
  await page.keyboard.press("Enter");
  expect(fixture.writes).toHaveLength(1);
  await expect(save).toHaveAttribute("aria-disabled", "false");
  await expectAllInteractiveControlsNamed(page);
  await testInfo.attach("metadata-accessibility-tree", {
    body: await panel.ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.keyboard.press("Escape");
  await expect(panel).toHaveCount(0);
  await expect(
    page.getByLabel("Account and application navigation"),
  ).toBeFocused();
  await openMetadata(page);
  await expect(panel.getByLabel("Severity", { exact: true })).toHaveValue(
    "intended",
  );
});

test("a11y.lifecycle multiline reason inline review exact recovery and departure retain keyboard focus", async ({
  testActorPage: page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("LC-A11Y"),
    "Lifecycle accessibility review",
  );
  await openIncidentFromLanding(page, incidentId);
  const panel = await openLifecycle(page);
  const reason = panel.getByRole("textbox", { name: "Reason", exact: true });
  await reason.fill("First reason line");
  await reason.press("End");
  await page.keyboard.press("Enter");
  await page.keyboard.type("Second reason line");
  await expect(reason).toHaveValue("First reason line\nSecond reason line");
  const close = panel.getByRole("button", {
    name: "Close incident",
    exact: true,
  });
  await close.focus();
  await page.keyboard.press("Enter");
  const confirm = panel.getByRole("button", {
    name: "Confirm Close incident",
    exact: true,
  });
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [
      reason,
      confirm,
      panel.getByRole("button", { name: "Cancel review", exact: true }),
      panel.getByRole("button", {
        name: "Refresh current incident",
        exact: true,
      }),
    ]) {
      await expectLifecycleControlReachable(page, control);
      await expectVisibleFocus(control);
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectLifecycleControlReachable(page, confirm);
    await expectVisibleFocus(confirm);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await expectLifecycleControlReachable(page, confirm);
    },
  );
  let requests = 0;
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/close`,
    async (route) => {
      ++requests;
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      if (requests === 1) await route.abort("failed");
      else await route.fulfill({ response });
    },
  );
  await confirm.focus();
  await page.keyboard.press("Enter");
  await expect(panel.getByText(/Close result is uncertain/u)).toBeVisible();
  const refresh = panel.getByRole("button", {
    name: "Refresh current incident",
    exact: true,
  });
  await expect(refresh).toBeFocused();
  await expectVisibleFocus(refresh);
  const replay = panel.getByRole("button", {
    name: "Replay original action",
    exact: true,
  });
  await expectLifecycleControlReachable(page, replay);
  await expectVisibleFocus(replay);
  await page.keyboard.press("Enter");
  await expect(
    panel.getByText("Close confirmed.", { exact: true }),
  ).toBeVisible();
  await expect(refresh).toBeFocused();
  expect(requests).toBe(2);
  await expectAllInteractiveControlsNamed(page);
  await testInfo.attach("lifecycle-accessibility-tree", {
    body: await panel.ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.keyboard.press("Escape");
  await expect(panel).toHaveCount(0);
  await expect(
    page.getByLabel("Account and application navigation"),
  ).toBeFocused();
  await openLifecycle(page);
  await expect(
    panel.getByText("Close confirmed.", { exact: true }),
  ).toBeVisible();
});
