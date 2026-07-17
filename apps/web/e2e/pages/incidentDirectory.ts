import {
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  phase1LandingTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { Browser, Page } from "@playwright/test";

import { expect } from "@playwright/test";

export async function openIncidentFromLanding(page: Page, incidentId: string) {
  await page.goto("/");
  const incidentCard = page.getByTestId(landingIncidentCardTestId(incidentId));
  const routed = await Promise.race([
    page
      .waitForURL(new RegExp(`incident_id=${incidentId}`), { timeout: 5000 })
      .then(() => "workbook" as const),
    incidentCard
      .waitFor({ state: "visible", timeout: 5000 })
      .then(() => "landing" as const),
  ]).catch(() => "unknown" as const);
  if (routed === "landing") {
    await page.getByTestId(landingIncidentOpenButtonTestId(incidentId)).click();
  }
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
}

export async function openIncidentAsTrackedUser(
  browser: Browser,
  sessionTracker: {
    loginTrackedUser: (
      page: Page,
      details: {
        createdBy: string;
        email: string;
        password: string;
        purpose: string;
        userId: string;
      },
    ) => Promise<void>;
  },
  options: {
    createdBy: string;
    email: string;
    incidentId: string;
    password: string;
    purpose: string;
    userId: string;
  },
) {
  const context = await browser.newContext();
  const page = await context.newPage();
  await sessionTracker.loginTrackedUser(page, {
    createdBy: options.createdBy,
    email: options.email,
    password: options.password,
    purpose: options.purpose,
    userId: options.userId,
  });
  await openIncidentFromLanding(page, options.incidentId);
  return page;
}

export class IncidentDirectory {
  constructor(private readonly page: Page) {}

  async goto() {
    await this.page.goto("/");
    await this.open();
  }

  async open() {
    await this.closeAccountSettingsIfOpen();
    const landingShell = this.page.getByTestId(phase1LandingTestId("shell"));
    if (await this.isOpen()) {
      return;
    }
    let lastError: unknown = null;
    for (let attempt = 0; attempt < 3; attempt += 1) {
      try {
        const menuItem = this.page.getByRole("menuitem", {
          exact: true,
          name: "Incidents",
        });
        if (!(await menuItem.isVisible().catch(() => false))) {
          const trigger = this.page.getByLabel(
            "Account and application navigation",
          );
          await expect(trigger).toBeVisible();
          await trigger.click();
        }
        await expect(menuItem).toBeVisible({ timeout: 1_000 });
        await menuItem.click();
        await expect(landingShell).toBeVisible();
        await expect.poll(async () => this.isOpen()).toBe(true);
        lastError = null;
        break;
      } catch (error) {
        lastError = error;
      }
    }
    if (lastError !== null) {
      throw lastError;
    }
    await expect(landingShell).toBeVisible();
  }

  async refresh() {
    await this.closeAccountSettingsIfOpen();
    await this.page.getByTestId(phase1LandingTestId("refresh")).click();
  }

  async createAndOpenIncident(incidentKey: string, title: string) {
    const createOpenButton = this.page.getByTestId(
      phase1LandingTestId("create-open-button"),
    );
    const incidentKeyInput = this.page.getByTestId(
      phase1LandingTestId("incident-key"),
    );
    let lastOpenError: unknown = null;
    for (let attempt = 0; attempt < 3; attempt += 1) {
      try {
        await this.open();
        await expect(createOpenButton).toBeVisible({ timeout: 3_000 });
        await createOpenButton.click();
        await expect(incidentKeyInput).toBeVisible({ timeout: 3_000 });
        lastOpenError = null;
        break;
      } catch (error) {
        lastOpenError = error;
      }
    }
    if (lastOpenError !== null) {
      throw lastOpenError;
    }
    await incidentKeyInput.fill(incidentKey);
    await this.page
      .getByTestId(phase1LandingTestId("incident-title"))
      .fill(title);
    await this.page
      .getByTestId(phase1LandingTestId("create-submit-button"))
      .click();
    await expect(this.page).toHaveURL(/incident_id=/);
    const openedIncidentId = new URL(this.page.url()).searchParams.get(
      "incident_id",
    );
    expect(openedIncidentId).not.toBeNull();
    await expect(
      this.page.getByTestId(workbookShellReadyTestId()),
    ).toBeVisible();
    return openedIncidentId ?? "";
  }

  async openIncident(incidentId: string) {
    await this.closeAccountSettingsIfOpen();
    if (
      new URL(this.page.url()).searchParams.get("incident_id") === incidentId
    ) {
      await expect(
        this.page.getByTestId(workbookShellReadyTestId()),
      ).toBeVisible();
      return;
    }
    await this.page
      .getByTestId(landingIncidentOpenButtonTestId(incidentId))
      .click();
    await expect(this.page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
    await expect(
      this.page.getByTestId(workbookShellReadyTestId()),
    ).toBeVisible();
  }

  async patchIncidentFields(options: {
    currentPhase?: string;
    externalCase?: string;
    tlp?: string;
  }) {
    if (options.tlp !== undefined) {
      await this.page
        .getByTestId("incident-patch-tlp")
        .selectOption(options.tlp);
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

  private async isOpen() {
    const landingShell = this.page.getByTestId(phase1LandingTestId("shell"));
    if (!(await landingShell.isVisible().catch(() => false))) {
      return false;
    }
    return this.page.evaluate(() => {
      const search = new URLSearchParams(window.location.search);
      return (
        window.location.pathname === "/" &&
        !search.has("incident_id") &&
        window.history.state?.cartularyIncidentDirectory === true
      );
    });
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
