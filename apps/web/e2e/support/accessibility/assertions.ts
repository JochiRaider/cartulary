// biome-ignore lint/correctness/noUndeclaredDependencies: the workspace root pins the browser accessibility engine used by the Make-owned harness.
import AxeBuilder from "@axe-core/playwright";
import { expect, type Locator, type Page } from "@playwright/test";

export async function expectNamedControl(
  locator: Locator,
  name: string | RegExp,
) {
  await expect(locator).toHaveCount(1);
  await expect(locator).toBeVisible();
  await expect(locator).toHaveAccessibleName(name);
}

export async function expectAllInteractiveControlsNamed(page: Page) {
  // Disabled controls still communicate a product action and require names.
  const controls = page.locator(
    'button:visible, a[href]:visible, input:not([type="hidden"]):visible, select:visible, textarea:visible, [role="button"]:visible',
  );
  await expect(controls.first()).toBeVisible();
  for (const control of await controls.all()) {
    // Pointer-only affordances (for example column resize grips) may be
    // explicitly absent from the accessibility tree, but must not receive
    // keyboard focus. Required named targets use expectNamedControl directly.
    if (
      await control.evaluate(
        (element) => element.closest('[aria-hidden="true"]') !== null,
      )
    ) {
      await expect(control).toHaveJSProperty("tabIndex", -1);
      await expect(control).not.toBeFocused();
    } else {
      await expect(control).toHaveAccessibleName(/\S/);
    }
  }
  const audit = await new AxeBuilder({ page })
    .withRules([
      "aria-hidden-focus",
      "button-name",
      "input-button-name",
      "label",
      "link-name",
      "select-name",
    ])
    .analyze();
  expect(audit.violations).toEqual([]);
}

export async function expectFocusContrast(locator: Locator) {
  await expect(locator).toHaveCount(1);
  await withContrastTarget(locator, async (selector) => {
    const checks = await collectContrastChecks(locator.page(), [
      { id: "focused-control", selector, kind: "focus" },
    ]);
    expect(checks.filter((check) => check.result !== "pass")).toEqual([]);
  });
}

import { collectContrastChecks, withContrastTarget } from "./contrast";
