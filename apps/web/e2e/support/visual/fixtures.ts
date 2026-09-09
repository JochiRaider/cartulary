import { test as base, expect } from "../../fixtures";
import { installVisualPreferences } from "../auth/visualPreferences";
import {
  assertApplicationAssetsReady,
  installApplicationAssetMonitor,
} from "../runtime/applicationReadiness";

export const test = base.extend({
  page: async ({ page, workerAdmin }, use, testInfo) => {
    const { preferences, assets } = await initializeVisualPage(
      page,
      workerAdmin.user_id,
    );
    await use(page);
    if (assets.length) {
      try {
        await testInfo.attach("application-asset-failures", {
          body: JSON.stringify({
            stage: "application_readiness",
            failures: assets,
          }),
          contentType: "application/json",
        });
      } catch {
        /* Preserve the resource failure if diagnostic attachment also fails. */
      }
      assertApplicationAssetsReady(page);
    }
    expect(
      preferences.unexpectedWrites,
      "visual preferences must not escape to the shared account",
    ).toEqual([]);
  },
});

import type { Page } from "@playwright/test";

export async function initializeVisualPage(page: Page, userId: string) {
  const preferences = await installVisualPreferences(page, userId);
  const assets = installApplicationAssetMonitor(page);
  return { preferences, assets };
}

export async function injectDesignFixture(
  page: Page,
  options: {
    ariaLabel: string;
    fixtureName: string;
    html: string;
    missingMainMessage: string;
    styleText: string;
  },
) {
  await page.evaluate((fixtureOptions) => {
    document
      .querySelector(
        `style[data-design-fixture-style='${fixtureOptions.fixtureName}']`,
      )
      ?.remove();
    document
      .querySelector(`[data-design-fixture='${fixtureOptions.fixtureName}']`)
      ?.remove();

    const main = document.querySelector("main.cartulary-shell");
    if (!(main instanceof HTMLElement)) {
      throw new Error(fixtureOptions.missingMainMessage);
    }

    const style = document.createElement("style");
    style.dataset.designFixtureStyle = fixtureOptions.fixtureName;
    style.textContent = fixtureOptions.styleText;
    document.head.append(style);

    const fixture = document.createElement("section");
    fixture.dataset.designFixture = fixtureOptions.fixtureName;
    fixture.setAttribute("aria-label", fixtureOptions.ariaLabel);
    fixture.innerHTML = fixtureOptions.html;
    main.append(fixture);
  }, options);
}
