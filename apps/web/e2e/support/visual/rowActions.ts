import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { mountedGridCell } from "../workbook/keyboardGrid";
export async function clickTimelineRowAction(
  page: Page,
  recordId: string,
  actionTestId: string,
) {
  // Keep the existing inspector and the fixture's left-column framing. Row
  // actions require a semantic cell anchor, not a click on the row container.
  const anchor = await mountedGridCell(
    page,
    timelineViewSchemaId,
    recordId,
    "timeline.date_entered_text",
  );
  await anchor.evaluate((element) => {
    const cell = element.closest<HTMLElement>('[role="gridcell"]');
    if (!cell) throw new Error("Expected Timeline action anchor cell");
    cell.focus();
  });
  await page.keyboard.press("Shift+F10");
  await page.getByTestId(actionTestId).click();
}
