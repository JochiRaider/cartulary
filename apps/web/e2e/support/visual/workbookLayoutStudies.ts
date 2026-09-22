import {
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Page, type TestInfo } from "@playwright/test";
import { settleVisualGeometry } from "./capture";

// Review attachments, not default layouts or automatically created saved views.
// Every arrangement uses the production Columns commands and the same records.
export async function captureWorkbookLayoutStudies(
  page: Page,
  info: TestInfo,
  normalizeGrid: () => Promise<void>,
) {
  const arrangements = [
    { name: "current", leading: [] },
    {
      name: "capture",
      leading: [
        "Activity Synopsis",
        "RAW Activity",
        "Device/Object",
        "Activity Date (UTC)",
        "Data Source",
      ],
    },
    {
      name: "review",
      leading: [
        "Activity Date (UTC)",
        "Activity Synopsis",
        "Device/Object",
        "Evidence",
        "Capture State",
      ],
    },
  ];
  for (const arrangement of arrangements) {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page
      .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
      .click();
    const panel = page.getByTestId(
      workbookColumnsMenuTestId(timelineViewSchemaId),
    );
    await panel
      .getByRole("button", { name: "Reset columns", exact: true })
      .click();
    for (const label of [...arrangement.leading].reverse()) {
      await panel.getByRole("checkbox", { name: label, exact: true }).check();
      const move = panel.getByRole("button", {
        name: `Move ${label} earlier`,
        exact: true,
      });
      // Move each leading field to the front, retaining the rest of the registry.
      for (let remaining = 100; await move.isEnabled(); remaining--) {
        if (remaining === 0) throw new Error(`Column failed to move: ${label}`);
        await move.click();
      }
    }
    await panel.press("Escape");
    for (const viewport of [
      { width: 1440, height: 900 },
      { width: 1280, height: 720 },
      { width: 1024, height: 720 },
      { width: 768, height: 640 },
    ]) {
      await page.setViewportSize(viewport);
      await page.evaluate(() => document.fonts.ready);
      await normalizeGrid();
      await page.mouse.move(0, 0);
      await settleVisualGeometry(page);
      await expect(
        page.locator('[data-grid-data-state="refreshing"]'),
      ).toHaveCount(0);
      await info.attach(
        `workbook-study-${arrangement.name}-${viewport.width}x${viewport.height}`,
        {
          body: await page.screenshot({ animations: "disabled" }),
          contentType: "image/png",
        },
      );
    }
  }
}
