import {
  type LandingAdminPanelToken,
  phase1AccountTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

type AccountSettingsPanel = Extract<
  LandingAdminPanelToken,
  "account-profile" | "account-appearance" | "account-security"
>;

export class AccountSettings {
  constructor(private readonly page: Page) {}

  async open(panel: AccountSettingsPanel = "account-security") {
    const panelLabel =
      panel === "account-profile"
        ? "Profile"
        : panel === "account-appearance"
          ? "Appearance"
          : "Security";
    const expectedControl =
      panel === "account-profile"
        ? phase1AccountTestId("profile-email")
        : panel === "account-appearance"
          ? phase1AccountTestId("appearance-density-mode")
          : phase1AccountTestId("refresh-state");
    const expectedControlLocator = this.page.getByTestId(expectedControl);
    try {
      await expect(expectedControlLocator).toBeVisible({ timeout: 500 });
      return;
    } catch {
      // The modal can still be opening after the preceding auth transition.
    }
    const closeButton = this.page.getByRole("button", { name: "Close" });
    try {
      await expect(closeButton).toBeVisible({ timeout: 500 });
      await this.page.getByRole("tab", { name: panelLabel }).click();
      await expect(expectedControlLocator).toBeVisible();
      return;
    } catch {
      // Open the absent modal through account navigation.
    }
    const trigger = this.page.getByLabel("Account and application navigation");
    await expect(trigger).toBeVisible();
    await trigger.click();
    await this.page.getByRole("menuitem", { name: "Account settings" }).click();
    if (panel !== "account-profile") {
      await this.page.getByRole("tab", { name: panelLabel }).click();
    }
    await expect(expectedControlLocator).toBeVisible();
  }

  async refresh() {
    await this.open("account-security");
    await this.page.getByTestId(phase1AccountTestId("refresh-state")).click();
  }

  async changePassword(
    currentPassword: string,
    nextPassword: string,
    factorCode: string,
  ) {
    await this.open("account-security");
    await this.page
      .getByTestId(phase1AccountTestId("password-current"))
      .fill(currentPassword);
    await this.page
      .getByTestId(phase1AccountTestId("password-next"))
      .fill(nextPassword);
    await this.page
      .getByTestId(phase1AccountTestId("password-factor-code"))
      .fill(factorCode);
    await this.page.getByTestId(phase1AccountTestId("password-change")).click();
  }
}
