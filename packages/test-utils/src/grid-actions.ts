import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  workbookFilterPopoverTriggerTestId,
} from "@cartulary/ui-contracts";

import {
  type BrowserPageLike,
  isLocatorVisible,
  requireEvaluate,
  requireSelectOption,
} from "./browser";
import { scrollGridCellIntoView, scrollGridTargetIntoView } from "./grid-setup";

function formatGridPasteMatrix(matrix: readonly (readonly string[])[]) {
  return matrix.map((row) => row.join("\t")).join("\n");
}

export async function assertActiveFilterChipVisible(
  page: BrowserPageLike,
  surface: string,
  fieldKey: string,
) {
  const chip = page.getByTestId(gridFilterChipTestId(surface, fieldKey));
  if (!(await isLocatorVisible(chip))) {
    throw new Error(
      `Expected active filter chip for ${fieldKey} on ${surface} to be visible`,
    );
  }
}

export async function pasteGridMatrix(options: {
  fieldKey: string;
  matrix: readonly (readonly string[])[];
  page: BrowserPageLike;
  recordId: string;
  surface: string;
}) {
  await scrollGridCellIntoView({
    cellKey: options.fieldKey,
    page: options.page,
    recordId: options.recordId,
    surface: options.surface,
  });
  const cell = options.page.getByTestId(
    rowCellTestId(options.recordId, options.fieldKey),
  );
  const evaluate = requireEvaluate(
    cell,
    `pasteGridMatrix(${options.surface}) requires locator.evaluate() support`,
  );
  await evaluate((element) => {
    if (element instanceof HTMLElement) {
      element.focus({ preventScroll: true });
    }
  });
  await evaluate((element, clipboardText) => {
    const data = new DataTransfer();
    data.setData("text/plain", String(clipboardText));
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  }, formatGridPasteMatrix(options.matrix));
}

export async function sortByHeader(
  page: BrowserPageLike,
  surface: string,
  fieldKey: string,
) {
  const headerTestId = gridSortHeaderTestId(surface, fieldKey);
  await scrollGridTargetIntoView({ page, surface, targetTestId: headerTestId });
  await page.getByTestId(headerTestId).click();
}

export async function applyFilterChip(
  page: BrowserPageLike,
  surface: string,
  fieldKey: string,
  value: string,
) {
  const fieldControl = page.getByTestId(gridFilterFieldTestId(surface));
  const selectField = requireSelectOption(
    fieldControl,
    `applyFilterChip(${surface}) requires filter field locator.selectOption() support`,
  );
  await page.getByTestId(workbookFilterPopoverTriggerTestId(surface)).click();
  await selectField(fieldKey);
  const valueControl = page.getByTestId(gridFilterValueTestId(surface));
  if (valueControl.selectOption === undefined) {
    await valueControl.fill(value);
  } else {
    try {
      await valueControl.selectOption(value);
    } catch {
      await valueControl.fill(value);
    }
  }
  await page.getByTestId(gridFilterApplyTestId(surface)).click();
}

export async function removeFilterChip(
  page: BrowserPageLike,
  surface: string,
  fieldKey: string,
) {
  await page.getByTestId(gridFilterChipTestId(surface, fieldKey)).click();
}

export async function changeGrouping(
  page: BrowserPageLike,
  surface: string,
  fieldKey: string,
) {
  const groupingControl = page.getByTestId(gridGroupingSelectTestId(surface));
  const selectGrouping = requireSelectOption(
    groupingControl,
    `changeGrouping(${surface}) requires locator.selectOption() support`,
  );
  await selectGrouping(fieldKey);
}

export async function setGridGroupExpanded(options: {
  expanded: boolean;
  groupTestId: string;
  page: BrowserPageLike;
  surface: string;
}) {
  const group = options.page.getByTestId(options.groupTestId);
  const evaluate = requireEvaluate(
    group,
    `setGridGroupExpanded(${options.surface}) requires locator.evaluate() support`,
  );
  const current = await evaluate((element) =>
    element.getAttribute("aria-expanded"),
  );
  if (current !== String(options.expanded)) {
    await group.click();
  }
  const next = await evaluate((element) =>
    element.getAttribute("aria-expanded"),
  );
  if (next !== String(options.expanded)) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to have aria-expanded=${String(options.expanded)}, received ${String(next)}`,
    );
  }
}

export function collapseGridGroup(
  options: Omit<Parameters<typeof setGridGroupExpanded>[0], "expanded">,
) {
  return setGridGroupExpanded({ ...options, expanded: false });
}

export function expandGridGroup(
  options: Omit<Parameters<typeof setGridGroupExpanded>[0], "expanded">,
) {
  return setGridGroupExpanded({ ...options, expanded: true });
}
