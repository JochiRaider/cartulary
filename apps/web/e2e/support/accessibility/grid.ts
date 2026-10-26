import { timelineScalarEditorTestId } from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect } from "../../fixtures";
import { mountedGridCell } from "../workbook/keyboardGrid";
import { expectVisibleFocus } from "./keyboard";

export async function expectVisibleSemanticGridCellFocus(
  content: Locator,
): Promise<Locator> {
  const identity = await content.evaluate((element) => ({
    fieldKey: element
      .closest<HTMLElement>("[data-grid-field-key]")
      ?.getAttribute("data-grid-field-key"),
    rowTestId: element
      .closest<HTMLElement>('[role="row"][data-testid]')
      ?.getAttribute("data-testid"),
  }));
  if (
    typeof identity.fieldKey !== "string" ||
    typeof identity.rowTestId !== "string"
  ) {
    throw new Error("Semantic grid cell is missing its stable test identity");
  }
  const cell = content
    .page()
    .getByTestId(identity.rowTestId)
    .getByRole("gridcell")
    .filter({
      has: content
        .page()
        .locator(`[data-grid-field-key="${identity.fieldKey}"]`),
    });
  await content.click();
  const editor = cell.locator(
    "input:not([type='checkbox']), textarea, select, [contenteditable='true']",
  );
  await expect
    .poll(async () => {
      if ((await editor.count()) > 0) return "editor";
      return cell.evaluate((element) =>
        element === document.activeElement ? "cell" : "pending",
      );
    })
    .not.toBe("pending");
  if ((await editor.count()) > 0) {
    await editor.first().press("Escape");
    await expect(editor).toHaveCount(0);
  }
  await expectVisibleFocus(cell);
  return cell;
}

export async function activateTimelineGridEditor(
  page: Page,
  recordId: string,
  fieldKey: string,
): Promise<Locator> {
  const display = await mountedGridCell(
    page,
    timelineViewSchemaId,
    recordId,
    fieldKey,
  );
  await display.click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey,
      recordId,
      surface: "grid",
    }),
  );
  await expectVisibleFocus(editor);
  return editor;
}
