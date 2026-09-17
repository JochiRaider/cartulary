import { expect, type Page } from "@playwright/test";

export function recoveryEntry(page: Page) {
  return page.getByRole("button", { name: /^Recovery \(\d+\)$/ });
}
/** Exercises the production entry and identity-labelled list, including settled notices. */
export async function openRecoveryItem(page: Page, name: RegExp) {
  const panel = page.getByRole("region", {
    name: "Recovery navigation",
    exact: true,
  });
  if (!(await panel.isVisible())) await recoveryEntry(page).click();
  const all = panel.getByRole("button", { name: "All recovery", exact: true });
  if (await all.isVisible()) await all.click();
  const item = panel.getByRole("button", { name, includeHidden: true });
  await expect(item).toHaveCount(1);
  // Settlement can move the same identity into the collapsed Completed group
  // between observation and click. Retry through the real disclosure control.
  await expect(async () => {
    if (!(await item.isVisible())) {
      const completed = panel.locator("details:not([open]) > summary");
      if (await completed.isVisible()) await completed.click();
    }
    await item.click({ timeout: 1_000 });
  }).toPass({ timeout: 10_000 });
  return panel;
}

/** The common panel heading owns focus after explicit recovery activation. */
export async function expectRecoveryFocus(page: Page, label: string) {
  const heading = page
    .getByRole("region", { name: "Recovery navigation", exact: true })
    .locator(":scope > h2");
  await expect(heading).toHaveText(label);
  await expect(heading).toBeFocused();
}
