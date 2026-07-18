import { authTestId, type StableTestId } from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";

import { expect } from "@playwright/test";

export class AuthGateway {
  constructor(private readonly page: Page) {}

  get loginUsername(): Locator {
    return this.page.getByTestId(authTestId("login-username"));
  }

  get loginPassword(): Locator {
    return this.page.getByTestId(authTestId("login-password"));
  }

  get loginTotpCode(): Locator {
    return this.page.getByTestId(authTestId("login-totp-code"));
  }

  async goto() {
    await this.page.goto("/");
  }

  async login(email: string, password: string, totpCode = "") {
    await this.loginUsername.fill(email);
    await this.loginPassword.fill(password);
    if (totpCode !== "") {
      if (!(await this.loginTotpCode.isVisible())) {
        await this.page.getByTestId(authTestId("login-submit")).click();
        await expect(this.loginTotpCode).toBeVisible();
      }
      await this.loginTotpCode.fill(totpCode);
    }
    await this.page.getByTestId(authTestId("login-submit")).click();
  }

  async requireText(testId: StableTestId) {
    const locator = this.page.getByTestId(testId);
    await expect
      .poll(async () => (await locator.textContent())?.trim() ?? "")
      .not.toBe("");
    const value = (await locator.textContent())?.trim() ?? "";
    if (value === "") {
      throw new Error(`missing text for ${testId}`);
    }
    return value;
  }

  async beginBootstrapEnrollment() {
    await this.page.getByTestId(authTestId("bootstrap-begin")).click();
  }

  async completeBootstrapEnrollment(code: string) {
    await this.page
      .getByTestId(authTestId("bootstrap-complete-code"))
      .fill(code);
    await this.page.getByTestId(authTestId("bootstrap-complete")).click();
  }
}
