import {
  accountTestId,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsTriggerTestId,
  landingAdminShellTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { installAccountEditingFixture } from "./support/auth/accountEditingFixture";
import { createIncident } from "./support/incidents/fixtures";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";

test("a11y.account-menu keyboard focus names current state and nested navigation", async ({
  testActorPage: page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YMENU"),
    "Account menu accessibility",
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const trigger = page.getByRole("button", {
    name: "Account and application navigation",
  });
  await trigger.focus();
  await page.keyboard.press("Enter");
  const root = page.getByRole("menu", {
    name: "Account and application navigation",
    exact: true,
  });
  await expect(trigger).toHaveAttribute(
    "aria-controls",
    (await root.getAttribute("id")) ?? "",
  );
  const incidents = page.getByRole("menuitem", {
    name: "Incidents",
    exact: true,
  });
  await expect(incidents).toBeFocused();
  await expectVisibleFocus(incidents);
  await page.keyboard.press("End");
  await page.keyboard.press("ArrowUp");
  const controls = page.getByTestId(incidentControlsTriggerTestId());
  await expect(controls).toBeFocused();
  await expect(controls).toHaveAccessibleDescription(
    "Opens incident Controls menu",
  );
  await page.keyboard.press("ArrowRight");
  const summary = page.getByTestId(incidentControlsMenuItemTestId("summary"));
  await expect(summary).toBeFocused();
  await expect(summary).toHaveAttribute("aria-current", "true");
  await expectVisibleFocus(summary);
  const nested = page.getByTestId(incidentControlsMenuTestId());
  await expect(controls).toHaveAttribute(
    "aria-controls",
    (await nested.getAttribute("id")) ?? "",
  );
  await expect(
    page
      .getByTestId(accountTestId("application-menu"))
      .locator('[role="menuitem"][tabindex="0"]'),
  ).toHaveCount(1);
  await expect(
    page
      .getByTestId(accountTestId("application-menu"))
      .locator('[aria-live], [role="status"]'),
  ).toHaveCount(0);
  await expectAllInteractiveControlsNamed(page);
  await expectAndRecordContrast(page, [
    incidentControlsTriggerTestId(),
    incidentControlsMenuItemTestId("summary"),
    incidentControlsMenuItemTestId("memberships"),
  ]);
  await testInfo.attach("account-menu-accessibility-tree", {
    body: await page
      .getByTestId(accountTestId("application-menu"))
      .ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.keyboard.press("Escape");
  await expect(controls).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
  await page.keyboard.press("Enter");
  await page.keyboard.press("Home");
  await page.keyboard.press("Enter");
  await expect(page.getByTestId(landingAdminShellTestId("heading"))).toHaveText(
    "Incident directory",
  );
  await expect(
    page.getByTestId(landingAdminShellTestId("heading")),
  ).toBeFocused();
  await trigger.focus();
  await page.keyboard.press("Enter");
  await expect(incidents).toBeFocused();
  await expect(incidents).toHaveAttribute("aria-current", "page");
  await expect(page.getByTestId(incidentControlsTriggerTestId())).toHaveCount(
    0,
  );
  await page.keyboard.press("Shift+Tab");
  await expect(trigger).toBeFocused();
  await expect(root).toHaveCount(0);
});

test("a11y.account-settings native forms radios validation and recovery", async ({
  testActorPage: page,
}, testInfo) => {
  const fixture = await installAccountEditingFixture(page);
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  const settings = new AccountSettings(page);
  await settings.openProfile();
  const dialog = page.getByRole("dialog", { name: "Account settings" });
  await dialog.getByRole("button", { name: "Close", exact: true }).focus();
  await page.keyboard.press("Shift+Tab");
  await expect
    .poll(() =>
      dialog.evaluate((node) => node.contains(document.activeElement)),
    )
    .toBe(true);
  await dialog.getByRole("tab", { name: "Profile", exact: true }).focus();
  await page.keyboard.press("ArrowRight");
  await expect(
    dialog.getByRole("tab", { name: "Appearance", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("ArrowLeft");
  const field = dialog.getByRole("textbox", { name: "Display name" });
  await field.focus();
  await page.keyboard.press("ControlOrMeta+A");
  await page.keyboard.insertText("   ");
  await expect(
    dialog.getByRole("button", { name: "Save profile", exact: true }),
  ).toBeEnabled();
  await page.keyboard.press("Enter");
  await expect(field).toHaveAttribute("aria-invalid", "true");
  await expect(field).toHaveAccessibleDescription("Enter a display name.");
  await expect(field).toBeFocused();
  await expectVisibleFocus(field);
  await expect(dialog.getByRole("alert")).toHaveCount(1);
  await expectAllInteractiveControlsNamed(page);
  await field.fill("Accessible draft");
  await expect(
    dialog.getByRole("button", { name: "Save profile", exact: true }),
  ).toBeEnabled();
  fixture.fault("profile", "lost");
  await field.press("Enter");
  const retry = dialog.getByRole("button", { name: "Retry save" });
  await retry.focus();
  await expectVisibleFocus(retry);
  await page.keyboard.press("Enter");
  await expect(dialog.getByRole("status")).toContainText("Profile saved.");
  await expectAndRecordContrast(page, [
    accountTestId("profile-display-name"),
    accountTestId("profile-save"),
  ]);
  await dialog.getByRole("tab", { name: "Appearance" }).focus();
  await page.keyboard.press("Enter");
  const group = dialog.getByRole("radiogroup", {
    name: "Density",
    exact: true,
  });
  const radios = group.getByRole("radio");
  await expect(radios).toHaveCount(4);
  for (const [index, label] of [
    "Use surface default",
    "Compact",
    "Default",
    "Comfortable",
  ].entries())
    await expect(radios.nth(index)).toHaveAccessibleName(label);
  await radios.first().focus();
  await page.keyboard.press("ArrowLeft");
  await expect(radios.last()).toBeChecked();
  await expect(radios.last()).toBeFocused();
  await expectVisibleFocus(radios.last());
  await page.keyboard.press("ArrowRight");
  await expect(radios.first()).toBeChecked();
  await page.keyboard.press("ArrowRight");
  await expect(radios.nth(1)).toBeChecked();
  await page.keyboard.press("Tab");
  await expect(
    dialog.getByRole("button", { name: "Save appearance" }),
  ).toBeFocused();
  await expectVisibleFocus(
    dialog.getByRole("button", { name: "Save appearance" }),
  );
  await expectAndRecordContrast(page, [accountTestId("appearance-save")]);
  await expectAllInteractiveControlsNamed(page);
  await testInfo.attach("account-settings-accessibility-tree", {
    body: await dialog.ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.keyboard.press("Escape");
  await expect(
    page.getByRole("button", { name: "Account and application navigation" }),
  ).toBeFocused();
});
