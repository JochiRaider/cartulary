import type { Locator, Page } from "@playwright/test";

import { expect } from "./fixtures";
import { apiBase } from "./helpers";

export class Phase1Page {
  constructor(private readonly page: Page) {}

  get loginUsername(): Locator {
    return this.page.getByTestId("auth-login-username");
  }

  get loginPassword(): Locator {
    return this.page.getByTestId("auth-login-password");
  }

  get loginTotpCode(): Locator {
    return this.page.getByTestId("auth-login-totp-code");
  }

  async goto() {
    await this.page.goto("/");
  }

  async login(email: string, password: string, totpCode = "") {
    await this.loginUsername.fill(email);
    await this.loginPassword.fill(password);
    await this.loginTotpCode.fill(totpCode);
    await this.page.getByTestId("auth-login-submit").click();
  }

  async requireText(testId: string) {
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
    await this.page.getByTestId("auth-bootstrap-begin").click();
  }

  async completeBootstrapEnrollment(code: string) {
    await this.page.getByTestId("auth-bootstrap-complete-code").fill(code);
    await this.page.getByTestId("auth-bootstrap-complete").click();
  }

  async refreshAccount() {
    await this.page.getByTestId("account-refresh-state").click();
  }

  async changePassword(
    currentPassword: string,
    nextPassword: string,
    factorCode: string,
  ) {
    await this.page
      .getByTestId("account-password-current")
      .fill(currentPassword);
    await this.page.getByTestId("account-password-next").fill(nextPassword);
    await this.page
      .getByTestId("account-password-factor-code")
      .fill(factorCode);
    await this.page.getByTestId("account-password-change").click();
  }

  async createUser(options: {
    displayName: string;
    email: string;
    isDeploymentAdmin?: boolean;
    mfaRequired?: boolean;
    password: string;
  }) {
    await this.page.getByTestId("admin-create-email").fill(options.email);
    await this.page
      .getByTestId("admin-create-display-name")
      .fill(options.displayName);
    await this.page.getByTestId("admin-create-password").fill(options.password);
    await this.setCheckbox(
      "admin-create-mfa-required",
      options.mfaRequired ?? true,
    );
    await this.setCheckbox(
      "admin-create-is-deployment-admin",
      options.isDeploymentAdmin ?? false,
    );
    await this.page.getByTestId("admin-create-user").click();
  }

  async loadTargetUser(userId: string) {
    const targetPath = `/api/v1/users/${userId}`;
    await this.page.getByTestId("admin-target-user-id-input").fill(userId);
    const [response] = await Promise.all([
      this.page.waitForResponse((candidate) => {
        const method = candidate.request().method().toUpperCase();
        if (method !== "GET") {
          return false;
        }
        return new URL(candidate.url()).pathname === targetPath;
      }),
      this.page.getByTestId("admin-load-user").click(),
    ]);
    expect(response.ok()).toBeTruthy();
    await expect(this.page.getByTestId("admin-target-user-id")).toHaveText(
      userId,
    );
    await this.requireText("admin-target-user-version");
  }

  async patchTargetUser() {
    await this.page.getByTestId("admin-patch-user").click();
  }

  async setCheckbox(testId: string, checked: boolean) {
    const checkbox = this.page.getByTestId(testId);
    await checkbox.setChecked(checked);
    if (checked) {
      await expect(checkbox).toBeChecked();
      return;
    }
    await expect(checkbox).not.toBeChecked();
  }

  async setClockOffset(offsetSeconds: number) {
    const response = await this.page.request.post(
      `${apiBase}/api/v1/test/clock/set`,
      {
        data: {
          offset_seconds: offsetSeconds,
        },
      },
    );
    expect(response.ok()).toBeTruthy();
  }

  async resetClockOffset() {
    await this.setClockOffset(0);
  }

  async withClockOffset<T>(
    offsetSeconds: number,
    work: () => Promise<T>,
  ): Promise<T> {
    await this.setClockOffset(offsetSeconds);
    try {
      return await work();
    } finally {
      await this.resetClockOffset();
    }
  }
}
