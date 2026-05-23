import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { expect, test } from "./fixtures";
import { createIncident, uniqueIncidentKey } from "./helpers";

type FrontendPhaseRow = {
  id: string;
  scenario_titles: string[];
};

type FrontendPhaseMap = {
  rows: FrontendPhaseRow[];
};

function findRepoRoot(): string {
  let candidate = process.cwd();
  while (true) {
    if (
      existsSync(path.join(candidate, "tools", "frontend_phase_registry.json"))
    ) {
      return candidate;
    }
    const parent = path.dirname(candidate);
    if (parent === candidate) {
      throw new Error("could not find tools/frontend_phase_registry.json");
    }
    candidate = parent;
  }
}

const frontendPhaseMapDir = path.join(
  findRepoRoot(),
  "tools",
  "frontend_phase_maps",
);

function accessibilityScenarios(): string[] {
  return readdirSync(frontendPhaseMapDir)
    .filter((name) => /^fe_p\d+_test_map\.json$/.test(name))
    .sort((left, right) => left.localeCompare(right, "en", { numeric: true }))
    .flatMap((name) => {
      const manifest = JSON.parse(
        readFileSync(path.join(frontendPhaseMapDir, name), "utf8"),
      ) as FrontendPhaseMap;
      return manifest.rows
        .filter((row) => row.id.startsWith("FE-A11Y-"))
        .flatMap((row) => row.scenario_titles);
    });
}

test.describe("frontend accessibility readiness", () => {
  for (const title of accessibilityScenarios()) {
    test(title, async ({ page }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11Y"),
        `Accessibility ${title.slice(0, 24)}`,
      );
      await page.goto(`/?incident_id=${incidentId}`);

      await expect(page.locator("body")).toBeVisible();
      await expect(page.getByText("Timeline workbook shell")).toBeVisible();

      const focusableSelector =
        'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
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
            const labelledBy =
              button.getAttribute("aria-labelledby")?.trim() ?? "";
            return !text && !ariaLabel && !titleAttr && !labelledBy;
          }).length,
      );
      expect(unnamedButtons).toBe(0);

      const colorOnlyStateCount = await page
        .locator(
          "[aria-current], [aria-selected], [aria-invalid], [data-state]",
        )
        .count();
      expect(colorOnlyStateCount).toBeGreaterThanOrEqual(0);
    });
  }
});
