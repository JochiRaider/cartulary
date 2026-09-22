import { expect, type Page } from "@playwright/test";

/** Query details have their own row; toolbar order assertions exclude chips. */
export async function expectQuerySummaryGeometry(page: Page, capacity: number) {
  const chips = page.getByRole("toolbar", { name: "Active query chips" });
  if (capacity === 0) {
    await expect(chips).toHaveCount(0);
    return;
  }
  const buttons = chips.getByRole("button");
  await expect(buttons).toHaveCount(capacity);
  const toolbar = await page
    .getByRole("region", {
      name: "Workbook query and action controls",
      exact: true,
    })
    .boundingBox();
  const summary = await chips.boundingBox();
  if (!toolbar || !summary) throw new Error("Missing query chrome geometry");
  expect(summary.y).toBeGreaterThanOrEqual(toolbar.y + toolbar.height - 1);
  expect(summary.height).toBeGreaterThanOrEqual(32);
  let priorRight = summary.x;
  for (const button of await buttons.all()) {
    const box = await button.boundingBox();
    if (!box) throw new Error("Missing query chip geometry");
    expect(box.height).toBeGreaterThanOrEqual(28);
    expect(box.x).toBeGreaterThanOrEqual(priorRight - 1);
    expect(box.x + box.width).toBeLessThanOrEqual(
      summary.x + summary.width + 1,
    );
    priorRight = box.x + box.width;
  }
}
