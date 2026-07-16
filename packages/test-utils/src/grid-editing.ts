import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  type WorkbookSurface,
  workbookFilterPopoverTriggerTestId,
} from "@cartulary/ui-contracts";

import {
  type BrowserLocator,
  type BrowserPageLike,
  isLocatorVisible,
  requireBlur,
  requireDispatchEvent,
  requireEvaluate,
  requirePress,
} from "./browser";
import { pasteMatrixText } from "./matrix";
import { scrollGridCellIntoView, scrollGridTargetIntoView } from "./scrolling";

export type GridAnchorCommandScenario = {
  commit: (context: {
    input: BrowserLocator;
    page: BrowserPageLike;
    surface: WorkbookSurface;
  }) => Promise<void>;
  name: string;
};

export function gridAnchorCommandScenarios(
  surface: WorkbookSurface,
): readonly GridAnchorCommandScenario[] {
  void surface;
  return [
    {
      commit: async ({ input }) => {
        await requirePress(input, "Enter");
      },
      name: "enter",
    },
    {
      commit: async ({ input }) => {
        await requirePress(input, "Tab");
      },
      name: "tab",
    },
    {
      commit: async ({ input, page, surface }) => {
        await requireBlur(input);
        await page.getByTestId(gridShellTestId(surface)).click();
      },
      name: "blur",
    },
    {
      commit: async ({ input }) => {
        await requireDispatchEvent(input, "paste");
      },
      name: "single-cell-paste",
    },
  ];
}

export async function resizeGridColumn(options: {
  deltaPx: number;
  fieldKey: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
}) {
  void options.deltaPx;
  const headerTestId = gridSortHeaderTestId(options.surface, options.fieldKey);
  await scrollGridTargetIntoView({
    page: options.page,
    surface: options.surface,
    targetTestId: headerTestId,
  });
  const header = options.page.getByTestId(headerTestId);
  const evaluate = requireEvaluate(
    header,
    `resizeGridColumn(${options.surface}) requires locator.evaluate() support`,
  );
  return evaluate((element) => element.getBoundingClientRect().width);
}

export async function assertActiveFilterChipVisible(
  page: BrowserPageLike,
  surface: WorkbookSurface,
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
  surface: WorkbookSurface;
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
  await cell.scrollIntoViewIfNeeded?.();
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
  }, pasteMatrixText(options.matrix));
}

export async function sortByHeader(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
) {
  const headerTestId = gridSortHeaderTestId(surface, fieldKey);
  await scrollGridTargetIntoView({ page, surface, targetTestId: headerTestId });
  await page.getByTestId(headerTestId).click();
}

export async function applyFilterChip(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
  value: string,
) {
  await page.getByTestId(workbookFilterPopoverTriggerTestId(surface)).click();
  await page
    .getByTestId(gridFilterFieldTestId(surface))
    .selectOption?.(fieldKey);
  const valueControl = page.getByTestId(gridFilterValueTestId(surface));
  try {
    await valueControl.selectOption?.(value);
  } catch {
    await valueControl.fill(value);
  }
  await page.getByTestId(gridFilterApplyTestId(surface)).click();
}

export async function removeFilterChip(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
) {
  await page.getByTestId(gridFilterChipTestId(surface, fieldKey)).click();
}

export async function changeGrouping(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
) {
  await page
    .getByTestId(gridGroupingSelectTestId(surface))
    .selectOption?.(fieldKey);
}

export function assertAnchorTestId(recordId: string, fieldKey: string): string {
  return rowCellTestId(recordId, fieldKey);
}
