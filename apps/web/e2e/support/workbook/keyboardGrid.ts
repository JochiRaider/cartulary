import {
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import { rowCellTestId } from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect } from "../../fixtures";

export async function mountedGridCell(
  page: Page,
  surface: string,
  recordId: string,
  fieldKey: string,
): Promise<Locator> {
  await scrollGridCellIntoView({
    cellKey: fieldKey,
    page,
    recordId,
    surface,
  });
  return page.getByTestId(rowCellTestId(recordId, fieldKey));
}

export async function mountedGridTarget(
  page: Page,
  surface: string,
  targetTestId: string,
): Promise<Locator> {
  await scrollGridTargetIntoView({ page, surface, targetTestId });
  return page.getByTestId(targetTestId);
}

export function semanticGridCell(content: Locator): Locator {
  return content.locator("xpath=ancestor::*[@role='gridcell'][1]");
}

export async function expectCellTextOrValue(locator: Locator, value: string) {
  const mode = await locator.evaluate((element) => {
    if (
      element instanceof HTMLInputElement ||
      element instanceof HTMLTextAreaElement ||
      element instanceof HTMLSelectElement
    ) {
      return "value";
    }
    return "text";
  });
  if (mode === "value") {
    await expect(locator).toHaveValue(value);
    return;
  }
  await expect(locator).toContainText(value);
}
