import {
  deploymentAdminTestId,
  deploymentUserRowTestId,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsPanelTestId,
  incidentControlsSurfaceTestId,
  incidentControlsTriggerTestId,
  landingAdminMenuItemTestId,
  type StableTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";

type DeploymentAdministrationPanelSelectorId =
  | "administrative-audit"
  | "deployment-users"
  | "incident-import"
  | "reference-packs";

type IncidentControlsSectionSelectorId =
  | "import-assistant"
  | "incident-fields"
  | "membership-audit"
  | "memberships"
  | "summary";

const incidentControlsLoadedStatePattern = /^(partial|synced)$/;

export async function openIncidentControls(
  page: Page,
  section: IncidentControlsSectionSelectorId = "summary",
) {
  await page.getByLabel("Account and application navigation").click();
  const trigger = page.getByTestId(incidentControlsTriggerTestId());
  await expect(trigger).toHaveAttribute("aria-haspopup", "menu");
  await trigger.click();
  await expect(page.getByTestId(incidentControlsMenuTestId())).toBeVisible();
  const menuItem = page.getByTestId(incidentControlsMenuItemTestId(section));
  await expect(menuItem).toHaveAttribute("role", "menuitem");
  await menuItem.click();
  const panel = page.getByTestId(incidentControlsPanelTestId());
  await expect(panel).toBeVisible();
  const surface = page.getByTestId(incidentControlsSurfaceTestId());
  await expect(surface).toBeVisible();
  await expect(surface).toHaveAttribute(
    "data-incident-controls-section",
    section,
  );
  await expect(surface).toHaveAttribute(
    "data-incident-controls-load-state",
    incidentControlsLoadedStatePattern,
  );
}

export class DeploymentAdministration {
  constructor(private readonly page: Page) {}

  async open() {
    await this.closeAccountSettingsIfOpen();
    if (new URL(this.page.url()).pathname !== "/deployment-administration") {
      const trigger = this.page.getByLabel(
        "Account and application navigation",
      );
      await expect(trigger).toBeVisible();
      await trigger.click();
      await this.page
        .getByRole("menuitem", { name: "Deployment administration" })
        .click();
    }
    await expect(
      this.page.getByTestId(landingAdminMenuItemTestId("deployment-users")),
    ).toBeVisible();
  }

  async selectPanel(panel: DeploymentAdministrationPanelSelectorId) {
    await this.open();
    const menuItem = this.page.getByTestId(landingAdminMenuItemTestId(panel));
    await menuItem.click();
    await expect(menuItem).toHaveAttribute("aria-pressed", "true");
  }

  async createUser(options: {
    displayName: string;
    email: string;
    isDeploymentAdmin?: boolean;
    mfaRequired?: boolean;
    password: string;
  }) {
    await this.selectPanel("deployment-users");
    await this.page.getByTestId(deploymentAdminTestId("create-user")).click();
    await this.page
      .getByTestId(deploymentAdminTestId("create-email"))
      .fill(options.email);
    await this.page
      .getByTestId(deploymentAdminTestId("create-display-name"))
      .fill(options.displayName);
    await this.page
      .getByTestId(deploymentAdminTestId("create-password"))
      .fill(options.password);
    await this.setCheckbox(
      deploymentAdminTestId("create-mfa-required"),
      options.mfaRequired ?? true,
    );
    await this.setCheckbox(
      deploymentAdminTestId("create-is-deployment-admin"),
      options.isDeploymentAdmin ?? false,
    );
    await this.page.getByTestId(deploymentAdminTestId("create-user")).click();
  }

  async loadTargetUser(userId: string) {
    await this.selectPanel("deployment-users");
    const targetPath = `/api/v1/users/${userId}`;
    const userFilter = this.page.getByTestId(
      deploymentAdminTestId("user-filter"),
    );
    const previousFilter = await userFilter.inputValue();
    const listResponsePromise =
      previousFilter === userId
        ? null
        : this.page.waitForResponse((candidate) => {
            const url = new URL(candidate.url());
            return (
              candidate.request().method().toUpperCase() === "GET" &&
              url.pathname === "/api/v1/users" &&
              url.searchParams.get("search") === userId
            );
          });
    await userFilter.fill(userId);
    if (listResponsePromise !== null) {
      expect((await listResponsePromise).ok()).toBeTruthy();
    }
    const targetRow = this.page.getByTestId(deploymentUserRowTestId(userId));
    await expect(targetRow).toBeVisible();
    const [response] = await Promise.all([
      this.page.waitForResponse(
        (candidate) =>
          candidate.request().method().toUpperCase() === "GET" &&
          new URL(candidate.url()).pathname === targetPath,
      ),
      targetRow.click(),
    ]);
    expect(response.ok()).toBeTruthy();
    await expect(
      this.page.getByTestId(deploymentAdminTestId("target-user-id")),
    ).toHaveText(userId);
    await expect
      .poll(
        async () =>
          (
            await this.page
              .getByTestId(deploymentAdminTestId("target-user-version"))
              .textContent()
          )?.trim() ?? "",
      )
      .not.toBe("");
  }

  async patchTargetUser() {
    await this.page.getByTestId(deploymentAdminTestId("patch-user")).click();
  }

  async setCheckbox(testId: StableTestId, checked: boolean) {
    const checkbox = this.page.getByTestId(testId);
    await checkbox.setChecked(checked);
    if (checked) {
      await expect(checkbox).toBeChecked();
      return;
    }
    await expect(checkbox).not.toBeChecked();
  }

  private async closeAccountSettingsIfOpen() {
    const closeButton = this.page.getByRole("button", {
      exact: true,
      name: "Close",
    });
    if ((await closeButton.count()) > 0 && (await closeButton.isVisible())) {
      await closeButton.click();
    }
  }
}
