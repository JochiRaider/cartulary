// biome-ignore lint/correctness/noUndeclaredDependencies: the workspace root pins the browser accessibility engine used by the Make-owned harness.
import AxeBuilder from "@axe-core/playwright";
import {
  gridScrollportSelector,
  gridShellTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { withPresentation } from "./support/accessibility/presentation";
import { createIncident } from "./support/incidents/fixtures";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";

test("a11y.workbench Commands remain visible and actionable across narrow width and zoom", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("COMMANDS"),
    "Commands responsive admission",
  );
  await page.goto(`/?incident_id=${incident}`);
  const trigger = page.getByRole("button", { name: "Commands", exact: true });
  await expect(trigger).toBeEnabled();
  await trigger.click();
  const commands = page.getByRole("region", { name: "Commands", exact: true });
  const search = commands.getByRole("textbox", {
    name: "Search commands",
    exact: true,
  });
  await expect(search).toBeFocused();
  for (const [width, height, zoom] of [
    [1440, 900, 1],
    [390, 480, 1],
    [1280, 900, 2],
  ] as const) {
    await page.setViewportSize({ width, height });
    await withPresentation(page, { zoom: String(zoom) }, async () => {
      await expect(search).toBeFocused();
      await expect
        .poll(() =>
          commands.evaluate((element) => {
            const bounds = element.getBoundingClientRect();
            const input = element.querySelector("input");
            if (!input) return false;
            const field = input.getBoundingClientRect();
            const hit = document.elementFromPoint(
              field.left + field.width / 2,
              field.top + field.height / 2,
            );
            return (
              bounds.left >= 0 &&
              bounds.top >= 0 &&
              bounds.right <= innerWidth &&
              bounds.bottom <= innerHeight &&
              hit === input
            );
          }),
        )
        .toBe(true);
      await search.fill("open work");
      await expect(
        commands.getByRole("button", {
          name: "Follow up · Open Work",
          exact: true,
        }),
      ).toBeVisible();
      await search.press("Enter");
      await expect(commands).toBeVisible();
      const axe = await new AxeBuilder({ page })
        .include('section[aria-label="Commands"]')
        .analyze();
      expect(axe.violations).toEqual([]);
      await info.attach(`workbench-commands-${width}-${zoom}`, {
        body: await page.screenshot({ animations: "disabled", caret: "hide" }),
        contentType: "image/png",
      });
    });
  }

  await page.setViewportSize({ width: 1440, height: 900 });
  await search.fill("sort");
  await commands
    .getByRole("button", { name: "View · Sort", exact: true })
    .click();
  const sort = page.getByTestId(workbookSortMenuTestId(timelineViewSchemaId));
  await expect(sort).toBeVisible();
  await sort.press("Escape");
  await expect(sort).toHaveCount(0);
  await page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector())
    .focus();
  for (const width of [768, 1440, 768]) {
    await page.setViewportSize({ width, height: 768 });
    if (width === 768) {
      await expect(
        page.getByRole("button", { name: "Query controls", exact: true }),
      ).toHaveAttribute("aria-expanded", "false");
      await expect(
        page.getByTestId(workbookSortMenuTriggerTestId(timelineViewSchemaId)),
      ).toBeHidden();
    }
  }
  await trigger.click();
  await expect(search).toBeFocused();
  await search.fill("open work");
  await commands
    .getByRole("button", { name: "Follow up · Open Work", exact: true })
    .click();
  await expect(commands).toBeHidden();
  await expect(
    page.getByRole("heading", { name: "Work", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByText("Session only · 0/20 pins", { exact: true }),
  ).toBeVisible();
});
