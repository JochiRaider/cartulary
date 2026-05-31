import { workbookShellReadyTestId } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { blockedAccessibilityScenarioTitles } from "./a11yPhaseMap";
import { expect, test } from "./fixtures";
import { createIncident, uniqueIncidentKey } from "./helpers";

const focusableSelector =
  'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
const blockedScenarioTitles = blockedAccessibilityScenarioTitles();

async function expectLaterPhaseSmoke(page: Page) {
  const focusableCount = await page.locator(focusableSelector).count();
  expect(focusableCount).toBeGreaterThan(0);

  await page.locator(focusableSelector).first().focus();
  await page.keyboard.press("Tab");
  const activeElementIsFocusable = await page.evaluate(() => {
    const active = document.activeElement;
    return Boolean(active && active !== document.body);
  });
  expect(activeElementIsFocusable).toBeTruthy();

  const unnamedButtons = await page.locator("button").evaluateAll(
    (buttons) =>
      buttons.filter((button) => {
        const text = button.textContent?.trim() ?? "";
        const ariaLabel = button.getAttribute("aria-label")?.trim() ?? "";
        const titleAttr = button.getAttribute("title")?.trim() ?? "";
        const labelledBy = button.getAttribute("aria-labelledby")?.trim() ?? "";
        return !text && !ariaLabel && !titleAttr && !labelledBy;
      }).length,
  );
  expect(unnamedButtons).toBe(0);
}

test.describe("frontend accessibility preflight smoke for blocked later phases", () => {
  for (const title of blockedScenarioTitles) {
    test(title, async ({ page }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11Y"),
        `Accessibility ${title.slice(0, 24)}`,
      );
      await page.goto(`/?incident_id=${incidentId}`);

      await expect(page.locator("body")).toBeVisible();
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expectLaterPhaseSmoke(page);
    });
  }
});
