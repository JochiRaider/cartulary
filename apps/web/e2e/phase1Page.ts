import {
  landingIncidentOpenButtonTestId,
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1LandingTestId,
  type StableTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";

import { expect } from "./fixtures";
import {
  apiBase,
  authHeadersForStorageState,
  testRouteHeaders,
} from "./helpers";

type CurrentSessionResponse = {
  data: {
    session_expires_at: string;
  };
};

export class Phase1Page {
  constructor(private readonly page: Page) {}

  get loginUsername(): Locator {
    return this.page.getByTestId(phase1AuthTestId("login-username"));
  }

  get loginPassword(): Locator {
    return this.page.getByTestId(phase1AuthTestId("login-password"));
  }

  get loginTotpCode(): Locator {
    return this.page.getByTestId(phase1AuthTestId("login-totp-code"));
  }

  async goto() {
    await this.page.goto("/");
  }

  async login(email: string, password: string, totpCode = "") {
    await this.loginUsername.fill(email);
    await this.loginPassword.fill(password);
    await this.loginTotpCode.fill(totpCode);
    await this.page.getByTestId(phase1AuthTestId("login-submit")).click();
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
    await this.page.getByTestId(phase1AuthTestId("bootstrap-begin")).click();
  }

  async completeBootstrapEnrollment(code: string) {
    await this.page
      .getByTestId(phase1AuthTestId("bootstrap-complete-code"))
      .fill(code);
    await this.page.getByTestId(phase1AuthTestId("bootstrap-complete")).click();
  }

  async refreshAccount() {
    await this.page.getByTestId(phase1AccountTestId("refresh-state")).click();
  }

  async refreshLanding() {
    await this.page.getByTestId(phase1LandingTestId("refresh")).click();
  }

  async createAndOpenIncident(incidentKey: string, title: string) {
    await this.page
      .getByTestId(phase1LandingTestId("incident-key"))
      .fill(incidentKey);
    await this.page
      .getByTestId(phase1LandingTestId("incident-title"))
      .fill(title);
    await this.page.getByTestId(phase1LandingTestId("create-button")).click();
  }

  async openIncident(incidentId: string) {
    await this.page
      .getByTestId(landingIncidentOpenButtonTestId(incidentId))
      .click();
  }

  async returnToLanding() {
    await this.page.getByTestId(phase1LandingTestId("return")).click();
  }

  async patchIncidentFields(options: {
    currentPhase?: string;
    externalCase?: string;
    tlp?: string;
  }) {
    if (options.tlp !== undefined) {
      await this.page.getByTestId("incident-patch-tlp").fill(options.tlp);
    }
    if (options.currentPhase !== undefined) {
      await this.page
        .getByTestId("incident-patch-current-phase")
        .fill(options.currentPhase);
    }
    if (options.externalCase !== undefined) {
      await this.page
        .getByTestId("incident-patch-external-case")
        .fill(options.externalCase);
    }
    await this.page.getByTestId("incident-patch-button").click();
  }

  async changePassword(
    currentPassword: string,
    nextPassword: string,
    factorCode: string,
  ) {
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

  async createUser(options: {
    displayName: string;
    email: string;
    isDeploymentAdmin?: boolean;
    mfaRequired?: boolean;
    password: string;
  }) {
    await this.page
      .getByTestId(phase1AdminTestId("create-email"))
      .fill(options.email);
    await this.page
      .getByTestId(phase1AdminTestId("create-display-name"))
      .fill(options.displayName);
    await this.page
      .getByTestId(phase1AdminTestId("create-password"))
      .fill(options.password);
    await this.setCheckbox(
      phase1AdminTestId("create-mfa-required"),
      options.mfaRequired ?? true,
    );
    await this.setCheckbox(
      phase1AdminTestId("create-is-deployment-admin"),
      options.isDeploymentAdmin ?? false,
    );
    await this.page.getByTestId(phase1AdminTestId("create-user")).click();
  }

  async loadTargetUser(userId: string) {
    const targetPath = `/api/v1/users/${userId}`;
    await this.page
      .getByTestId(phase1AdminTestId("target-user-id-input"))
      .fill(userId);
    const [response] = await Promise.all([
      this.page.waitForResponse((candidate) => {
        const method = candidate.request().method().toUpperCase();
        if (method !== "GET") {
          return false;
        }
        return new URL(candidate.url()).pathname === targetPath;
      }),
      this.page.getByTestId(phase1AdminTestId("load-user")).click(),
    ]);
    expect(response.ok()).toBeTruthy();
    await expect(
      this.page.getByTestId(phase1AdminTestId("target-user-id")),
    ).toHaveText(userId);
    await this.requireText(phase1AdminTestId("target-user-version"));
  }

  async patchTargetUser() {
    await this.page.getByTestId(phase1AdminTestId("patch-user")).click();
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

  async setClockOffset(offsetSeconds: number) {
    const response = await this.page.request.post(
      `${apiBase}/api/v1/test/clock/set`,
      {
        data: {
          offset_seconds: offsetSeconds,
        },
        headers: testRouteHeaders(),
      },
    );
    expect(response.ok()).toBeTruthy();
  }

  async setClockFixed(fixedNow: string | Date) {
    const response = await this.page.request.post(
      `${apiBase}/api/v1/test/clock/set`,
      {
        data: {
          fixed_now:
            fixedNow instanceof Date ? fixedNow.toISOString() : fixedNow,
        },
        headers: testRouteHeaders(),
      },
    );
    expect(response.ok()).toBeTruthy();
  }

  async setClockAfter(timestamp: string, deltaMs = 2000) {
    const baseline = new Date(timestamp);
    if (Number.isNaN(baseline.getTime())) {
      throw new Error(
        `cannot set test clock after invalid timestamp ${timestamp}`,
      );
    }
    await this.setClockFixed(new Date(baseline.getTime() + deltaMs));
  }

  async currentSession() {
    const storageState = await this.page.context().storageState();
    const response = await this.page.request.get(
      `${apiBase}/api/v1/auth/session`,
      {
        headers: authHeadersForStorageState(storageState),
      },
    );
    expect(response.ok()).toBeTruthy();
    return ((await response.json()) as CurrentSessionResponse).data;
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
