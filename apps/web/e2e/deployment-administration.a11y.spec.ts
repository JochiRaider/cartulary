import { deploymentAdminTestId } from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { DeploymentAdministration } from "./pages/deploymentAdministration";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import {
  auditBrowserEvent,
  expectAuditControlReachable,
  installAuditPresentation,
  openAdministrativeAudit,
} from "./support/administrativeAudit";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import { responseBarrier } from "./support/runtime/cleanup";
import { uniqueEmail } from "./support/runtime/fixtureIdentity";

test("a11y.deployment-users guarded drafts and credential dialogs remain keyboard reachable", async ({
  testActorPage: page,
  testActorRequest,
}, testInfo) => {
  const target = await createDeploymentUser(testActorRequest, {
    email: uniqueEmail("admin-accessibility"),
    display_name: "Accessible target",
    initial_password: "AccessibleTargetPass!",
    mfa_required: false,
    is_deployment_admin: false,
  });
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  await new DeploymentAdministration(page).loadTargetUser(target.user_id);
  const field = page.getByTestId(deploymentAdminTestId("patch-display-name"));
  await field.fill("Unsaved accessible draft");
  await page.getByRole("button", { name: "Clear", exact: true }).click();
  const guard = page.getByRole("dialog", { name: "Unsaved user changes" });
  await expect(guard).toBeVisible();
  expect(await guard.evaluate((node) => node.matches(":modal"))).toBe(true);
  await field.evaluate((node) => node.focus());
  await expect(
    guard.getByRole("button", { name: "Stay", exact: true }),
  ).toBeFocused();
  await expect(
    guard.getByRole("button", { name: "Stay", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(
    guard.getByRole("button", { name: "Save and leave", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    guard.getByRole("button", { name: "Stay", exact: true }),
  ).toBeFocused();
  await expectAllInteractiveControlsNamed(page);
  await page.keyboard.press("Escape");
  await expect(guard).toHaveCount(0);
  await expect(field).toHaveValue("Unsaved accessible draft");
  await page
    .getByRole("button", { name: "Discard user edits", exact: true })
    .click();
  const reset = page.getByTestId(deploymentAdminTestId("password-reset"));
  await reset.click();
  const dialog = page.getByRole("dialog", {
    name: "Confirm credential action",
  });
  expect(await dialog.evaluate((node) => node.matches(":modal"))).toBe(true);
  const password = dialog.getByLabel("New password", { exact: true });
  await password.fill("short");
  await password.press("Enter");
  await expect(password).toHaveAttribute("aria-invalid", "true");
  await expect(password).toHaveAccessibleDescription(/12/);
  await expect(dialog.getByRole("status")).toHaveCount(1);
  await page.mouse.click(1, 1);
  await expect(dialog).toBeVisible();
  await password.fill("Transient Password!");
  for (const viewport of [
    { width: 360, height: 480 },
    { width: 800, height: 320 },
  ]) {
    await page.setViewportSize(viewport);
    await dialog.evaluate((node) => {
      node.style.lineHeight = "1.5";
      node.style.letterSpacing = "0.12em";
      node.style.wordSpacing = "0.16em";
    });
    await dialog
      .getByRole("button", { name: "Close credential action" })
      .focus();
    await page.keyboard.press("Shift+Tab");
    await expect(
      dialog.getByRole("button", { name: "Confirm", exact: true }),
    ).toBeFocused();
    const box = await dialog.boundingBox();
    expect(box).not.toBeNull();
    if (box) {
      expect(box.x).toBeGreaterThanOrEqual(0);
      expect(box.width).toBeLessThanOrEqual(viewport.width);
      expect(box.height).toBeLessThanOrEqual(viewport.height);
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await dialog.getByRole("button", { name: "Confirm", exact: true }).focus();
    const zoomBounds = await dialog.boundingBox();
    expect(zoomBounds).not.toBeNull();
    if (zoomBounds) {
      expect(zoomBounds.x).toBeGreaterThanOrEqual(0);
      expect(zoomBounds.y).toBeGreaterThanOrEqual(0);
      expect(zoomBounds.x + zoomBounds.width).toBeLessThanOrEqual(1280);
      expect(zoomBounds.y + zoomBounds.height).toBeLessThanOrEqual(720);
    }
  });
  await testInfo.attach("deployment-user-dialog-accessibility-tree", {
    body: await dialog.ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(reset).toBeFocused();
  await reset.click();
  await expect(dialog.getByLabel("New password", { exact: true })).toHaveValue(
    "",
  );
  await page
    .getByRole("button", {
      name: "Reset password",
      exact: true,
      includeHidden: true,
    })
    .evaluate((node) => node.remove());
  await page.keyboard.press("Escape");
  await expect(
    page.getByRole("button", { name: "Account and application navigation" }),
  ).toBeFocused();
});

test("a11y.administrative-audit exact filters paging inspection and recovery stay reachable", async ({
  testActorPage: page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const fixture = await installAuditPresentation(page);
  fixture.setPage([auditBrowserEvent()], "next-page");
  const panel = await openAdministrativeAudit(page);
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  const target = panel.getByLabel("Target ID", { exact: true });
  await target.fill("exact-target");
  await target.press("Enter");
  const kind = panel.getByLabel("Target kind", { exact: true });
  await expect(kind).toBeFocused();
  await expect(kind).toHaveAccessibleDescription(/Choose a target kind/);
  await expectVisibleFocus(kind);
  await panel
    .getByRole("button", { name: "Clear filters", exact: true })
    .click();
  const inspect = panel.getByRole("button", { name: /^Inspect / });
  await inspect.focus();
  await page.keyboard.press("Enter");
  const hide = panel.getByRole("button", { name: /^Hide / });
  await expect(hide).toHaveAttribute("aria-expanded", "true");
  await expect(
    panel.getByRole("region", { name: /^Details for / }),
  ).toContainText("Redacted");
  await expect(panel.locator("script")).toHaveCount(0);
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 640, height: 480 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    await expectAuditControlReachable(page, hide);
    await expectVisibleFocus(hide);
    const identityCell = panel
      .getByRole("table", { name: "Deployment audit events", exact: true })
      .locator("tbody > tr > td")
      .first();
    expect((await identityCell.boundingBox())?.width).toBeGreaterThan(75);
    await expectAuditControlReachable(
      page,
      panel.getByRole("button", { name: "Next page", exact: true }),
    );
    await expectAuditControlReachable(
      page,
      panel.getByRole("button", { name: "Apply filters", exact: true }),
    );
    await expectAuditControlReachable(
      page,
      panel.getByRole("button", { name: "Refresh", exact: true }),
    );
    await expectAuditControlReachable(
      page,
      page.getByLabel("Account and application navigation"),
    );
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectAuditControlReachable(page, hide);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await expectAuditControlReachable(page, hide);
      await expectAllInteractiveControlsNamed(page);
      const gate = responseBarrier(page);
      fixture.fail();
      fixture.gateRead(gate.promise);
      await panel.getByRole("button", { name: "Refresh", exact: true }).click();
      await expect(panel.getByRole("status")).toContainText("Refreshing");
      gate.release();
      const retry = panel.getByRole("button", {
        name: "Retry refresh",
        exact: true,
      });
      await expect(retry).toBeVisible();
      await expectAuditControlReachable(page, retry);
      await expectVisibleFocus(retry);
      await testInfo.attach("administrative-audit-accessibility-tree", {
        body: await panel.ariaSnapshot(),
        contentType: "text/plain",
      });
      fixture.setPage([]);
      await page.keyboard.press("Enter");
      await expect(
        panel.getByText("No deployment audit events are available."),
      ).toBeVisible();
      await expect(
        panel.getByRole("button", { name: "Refresh", exact: true }),
      ).toBeFocused();
    },
  );
});
