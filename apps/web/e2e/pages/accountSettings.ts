import { accountTestId } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";

type AccountSettingsPanel =
  | "account-profile"
  | "account-appearance"
  | "account-security";

export class AccountSettings {
  constructor(private readonly page: Page) {}

  async openProfile() {
    await this.openPanel("account-profile");
    await expect(
      this.page.getByRole("textbox", { name: "Display name" }),
    ).toBeEnabled();
  }

  async selectDensity(
    label: "Use surface default" | "Compact" | "Default" | "Comfortable",
  ) {
    await this.openAppearance();
    await this.page
      .getByRole("radiogroup", { name: "Density", exact: true })
      .getByRole("radio", { name: label, exact: true })
      .check();
    await expect(
      this.page.getByRole("button", { name: "Refresh appearance" }),
    ).toBeEnabled();
  }

  async openAppearance() {
    await this.openPanel("account-appearance");
  }

  async openSecurity() {
    await this.openPanel("account-security");
  }

  private async openPanel(panel: AccountSettingsPanel) {
    const panelLabel =
      panel === "account-profile"
        ? "Profile"
        : panel === "account-appearance"
          ? "Appearance"
          : "Security";
    const expectedControl =
      panel === "account-profile"
        ? accountTestId("profile-display-name")
        : panel === "account-appearance"
          ? accountTestId("appearance-density-mode")
          : accountTestId("refresh-state");
    const expectedControlLocator = this.page.getByTestId(expectedControl);
    try {
      await expect(expectedControlLocator).toBeVisible({ timeout: 500 });
      return;
    } catch {
      // The modal can still be opening after the preceding auth transition.
    }
    const closeButton = this.page.getByRole("button", { name: "Close" });
    let dialogIsOpen = false;
    try {
      await expect(closeButton).toBeVisible({ timeout: 500 });
      dialogIsOpen = true;
    } catch {
      // Open the absent modal through account navigation below.
    }
    if (dialogIsOpen) {
      const tab = this.page.getByRole("tab", { name: panelLabel });
      if ((await tab.getAttribute("aria-selected")) !== "true") {
        await tab.click();
      }
      await expect(expectedControlLocator).toBeVisible();
      return;
    }
    const trigger = this.page.getByRole("button", {
      name: "Account and application navigation",
    });
    await expect(trigger).toBeVisible();
    await trigger.click();
    await this.page.getByRole("menuitem", { name: "Account settings" }).click();
    await this.page.getByRole("tab", { name: panelLabel }).click();
    await expect(expectedControlLocator).toBeVisible();
  }

  async refresh() {
    await this.openSecurity();
    await this.page.getByTestId(accountTestId("refresh-state")).click();
  }

  async changePassword(
    currentPassword: string,
    nextPassword: string,
    factorCode: string,
  ) {
    await this.openSecurity();
    await this.page
      .getByTestId(accountTestId("password-current"))
      .fill(currentPassword);
    await this.page
      .getByTestId(accountTestId("password-next"))
      .fill(nextPassword);
    await this.page
      .getByTestId(accountTestId("password-factor-code"))
      .fill(factorCode);
    await this.page.getByTestId(accountTestId("password-change")).click();
  }
}
