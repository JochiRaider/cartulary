import {
  accountTestId,
  deploymentAdminTestId,
  incidentControlsMenuItemTestId,
  incidentControlsTriggerTestId,
  incidentLandingTestId,
  landingAdminShellTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { expect } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { withPresentation } from "./support/accessibility/presentation";
import { installAccountEditingFixture } from "./support/auth/accountEditingFixture";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import { createRichTimelineIncident } from "./support/visual/timelineVisualFixture";

test.describe("Account navigation presentation", () => {
  // This email is visible fixture data in the search field and user row.
  // It must remain independent of source paths, titles and execution ordering.
  test.use({
    testActorEmail: "playwright-test-4a00afb3e20c32e802899533@example.test",
  });
  test("Capture account application menu root and nested Controls across contexts and constrained viewports.", async ({
    testActorPage: page,
    testActor,
  }) => {
    await installAccountEditingFixture(page, { profileOnly: true });
    const incidentId = await createRichTimelineIncident(
      page,
      uniqueIncidentKey("VISUALMENU"),
      "Account menu navigation",
    );
    await page.setViewportSize({ width: 1280, height: 720 });
    await navigateVisualApplication(page, `/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    const trigger = page.getByRole("button", {
      name: "Account and application navigation",
    });
    const controls = page.getByTestId(incidentControlsTriggerTestId());
    await trigger.focus();
    await page.keyboard.press("Enter");
    await expect(
      page.getByRole("menuitem", { name: "Incidents", exact: true }),
    ).toBeFocused();
    await assertViewportVisualRegression(page, "account-menu-workbook-root");
    await page.keyboard.press("Enter");
    await expect(
      page.getByTestId(landingAdminShellTestId("heading")),
    ).toHaveText("Incident directory");
    await expect(
      page.getByTestId(landingAdminShellTestId("heading")),
    ).toBeFocused();
    const filteredDirectory = page.waitForResponse(
      (response) =>
        response.request().method() === "GET" &&
        new URL(response.url()).searchParams.get("search") ===
          "Account menu navigation",
    );
    await page
      .getByTestId(incidentLandingTestId("search"))
      .fill("Account menu navigation");
    await filteredDirectory;
    await expect(
      page.getByTestId(incidentLandingTestId("incident-list")).getByRole("row"),
    ).toHaveCount(2);
    await trigger.click();
    await assertViewportVisualRegression(page, "account-menu-directory-root");
    await page.keyboard.press("ArrowDown");
    await page.keyboard.press("Enter");
    await expect(
      page.getByTestId(landingAdminShellTestId("heading")),
    ).toHaveText("Deployment administration");
    await expect(
      page.getByTestId(landingAdminShellTestId("heading")),
    ).toBeFocused();
    await page
      .getByRole("textbox", { name: "Search users", exact: true })
      .fill(testActor.email);
    await expect(
      page.getByTestId(deploymentAdminTestId("user-list")).getByRole("button"),
    ).toHaveCount(1);
    await trigger.click();
    await assertViewportVisualRegression(page, "account-menu-deployment-root");
    await navigateVisualApplication(page, `/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    for (const [name, width, height] of [
      ["account-menu-controls-narrow", 1024, 720],
      ["account-menu-controls-compact", 768, 640],
      ["account-menu-controls-short", 640, 480],
    ] as const) {
      await page.setViewportSize({ width, height });
      await trigger.click();
      await controls.click();
      await page.keyboard.press("End");
      await expect(
        page.getByTestId(incidentControlsMenuItemTestId("membership-audit")),
      ).toBeFocused();
      await assertViewportVisualRegression(page, name);
      await page.keyboard.press("Escape");
      await page.keyboard.press("Escape");
    }
    await page.setViewportSize({ width: 1280, height: 720 });
    await trigger.click();
    await page
      .getByRole("menuitem", { name: "Account settings", exact: true })
      .click();
    const label = "Analyst".repeat(36);
    await expect(
      page.getByTestId(accountTestId("profile-display-name")),
    ).toBeEnabled();
    await page.getByTestId(accountTestId("profile-display-name")).fill(label);
    await page.getByTestId(accountTestId("profile-save")).click();
    await expect(trigger).toHaveAttribute("title", label);
    await page
      .getByRole("dialog", { name: "Account settings" })
      .getByRole("button", { name: "Close", exact: true })
      .click();
    await withPresentation(page, { zoom: "200%" }, async () => {
      await trigger.click();
      await controls.click();
      await page.keyboard.press("End");
      await assertViewportVisualRegression(
        page,
        "account-menu-long-label-zoom",
      );
      await page.keyboard.press("Escape");
      await page.keyboard.press("Escape");
    });
    await page.setViewportSize({ width: 768, height: 640 });
    await withPresentation(
      page,
      {
        textSpacing:
          "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
      },
      async () => {
        await trigger.click();
        await controls.click();
        await page.keyboard.press("End");
        await assertViewportVisualRegression(
          page,
          "account-menu-long-label-text-spacing",
        );
      },
    );
  });
});

test("Capture account settings drafts pending conflict recovery and responsive states.", async ({
  testActorPage: page,
}) => {
  const fixture = await installAccountEditingFixture(page);
  await page.setViewportSize({ width: 1280, height: 720 });
  await navigateVisualApplication(page, "/");
  const settings = new AccountSettings(page);
  await settings.openProfile();
  const field = page.getByRole("textbox", { name: "Display name" });
  await field.fill("Account profile draft");
  await assertViewportVisualRegression(page, "account-settings-profile-dirty");
  const profilePending = fixture.hold("profile");
  await field.press("Enter");
  await profilePending.received;
  await expect(
    page.getByRole("button", { name: "Save profile" }),
  ).toHaveAttribute("aria-busy", "true");
  await assertViewportVisualRegression(
    page,
    "account-settings-profile-pending",
  );
  profilePending.release();
  await expect(
    page.getByRole("button", { name: "Account and application navigation" }),
  ).toHaveAttribute("title", "Account profile draft");
  await field.fill("My retained display name");
  fixture.fault("profile", "conflict");
  await field.press("Enter");
  await expect(
    page.getByRole("button", { name: "Review my edit" }),
  ).toBeEnabled();
  await assertViewportVisualRegression(
    page,
    "account-settings-profile-conflict",
  );
  await page.getByRole("button", { name: "Review my edit" }).click();
  fixture.fault("profile", "lost");
  await field.press("Enter");
  await expect(page.getByRole("button", { name: "Retry save" })).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "account-settings-profile-recovery",
  );
  await page.getByRole("button", { name: "Retry save" }).click();
  await expect(
    page.getByRole("button", { name: "Account and application navigation" }),
  ).toHaveAttribute("title", "My retained display name");
  await settings.selectDensity("Comfortable");
  await assertViewportVisualRegression(
    page,
    "account-settings-appearance-dirty",
  );
  const appearancePending = fixture.hold("appearance");
  await page.getByRole("button", { name: "Save appearance" }).click();
  await appearancePending.received;
  await assertViewportVisualRegression(
    page,
    "account-settings-appearance-pending",
  );
  appearancePending.release();
  await expect(
    page.getByRole("status").filter({ hasText: "Appearance saved." }),
  ).toBeVisible();
  await settings.selectDensity("Compact");
  fixture.fault("appearance", "conflict");
  await page.getByRole("button", { name: "Save appearance" }).click();
  await expect(
    page.getByRole("button", { name: "Review my edit" }),
  ).toBeEnabled();
  await assertViewportVisualRegression(
    page,
    "account-settings-appearance-conflict",
  );
  await page.getByRole("button", { name: "Review my edit" }).click();
  fixture.fault("appearance", "lost");
  await page.getByRole("button", { name: "Save appearance" }).click();
  await expect(page.getByRole("button", { name: "Retry save" })).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "account-settings-appearance-recovery",
  );
  await page.getByRole("button", { name: "Retry save" }).click();
  await expect(
    page.getByRole("status").filter({ hasText: "Appearance saved." }),
  ).toBeVisible();
  await settings.selectDensity("Use surface default");
  await page.setViewportSize({ width: 640, height: 480 });
  await assertViewportVisualRegression(
    page,
    "account-settings-appearance-short",
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await assertViewportVisualRegression(
      page,
      "account-settings-appearance-zoom",
    );
  });
  await page.setViewportSize({ width: 640, height: 640 });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await assertViewportVisualRegression(
        page,
        "account-settings-appearance-text-spacing",
      );
    },
  );
  await page.setViewportSize({ width: 768, height: 640 });
  await page.getByRole("tab", { name: "Profile" }).click();
  await field.fill("Analyst".repeat(36));
  await assertViewportVisualRegression(
    page,
    "account-settings-profile-long-name",
  );
});
