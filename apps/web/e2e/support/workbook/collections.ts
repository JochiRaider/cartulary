import {
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Locator, type Page } from "@playwright/test";

export async function showTimelineCollectionColumns(
  page: Page,
  labels: readonly string[] = ["Hosts", "Identities", "Tags"],
) {
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  const menu = page.getByTestId(
    workbookColumnsMenuTestId(timelineViewSchemaId),
  );
  for (const label of [...labels].reverse()) {
    const visibility = menu.getByRole("menuitemcheckbox", {
      name: label,
      exact: true,
    });
    if ((await visibility.getAttribute("aria-checked")) === "false")
      await visibility.click();
    const earlier = menu.getByRole("menuitem", {
      name: `Move ${label} earlier`,
      exact: true,
    });
    for (
      let remaining = 40;
      remaining > 0 && (await earlier.isEnabled());
      remaining -= 1
    )
      await earlier.click();
    await expect(earlier).toBeDisabled();
  }
  await menu.getByRole("menuitemcheckbox").first().focus();
  await page.keyboard.press("Escape");
  await expect(menu).toHaveCount(0);
}

export async function expectCollectionControlPainted(control: Locator) {
  await expect(control).toBeVisible();
  await expect(control).toBeInViewport();
  const result = await control.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    if (rect.width <= 0 || rect.height <= 0) return "empty control";
    for (
      let parent: Element | null = element;
      parent;
      parent = parent.parentElement
    ) {
      const style = window.getComputedStyle(parent);
      if (
        style.opacity === "0" ||
        style.visibility === "hidden" ||
        style.clipPath !== "none"
      )
        return "unpainted control";
      const bounds = parent.getBoundingClientRect();
      if (
        parent !== element &&
        style.overflowX !== "visible" &&
        (rect.left < bounds.left - 1 || rect.right > bounds.right + 1)
      )
        return "horizontal clipping";
      if (
        parent !== element &&
        style.overflowY !== "visible" &&
        (rect.top < bounds.top - 1 || rect.bottom > bounds.bottom + 1)
      )
        return `vertical clipping: ${element.tagName} (${rect.height}px) inside ${parent.tagName} (${bounds.height}px)`;
    }
    return "painted";
  });
  expect(result).toBe("painted");
}
