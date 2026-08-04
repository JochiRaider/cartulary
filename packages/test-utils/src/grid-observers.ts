import {
  dataTestIdSelector,
  gridGroupRowSelector,
  gridSavedRowsSelector,
  gridScrollportSelector,
  gridShellTestId,
} from "@cartulary/ui-contracts";

import {
  type BrowserPageLike,
  isLocatorVisible,
  requireEvaluate,
} from "./browser";

export const viewportVisibilityTolerancePx = 1;

export async function assertMountedGridRowCountAtMost(options: {
  maxRows: number;
  page: BrowserPageLike;
  surface: string;
}) {
  const { maxRows, page, surface } = options;
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `assertMountedGridRowCountAtMost(${surface}) requires locator.evaluate() support`,
  );
  const mountedRows = (await evaluate(
    (element, selector) =>
      element.querySelectorAll(typeof selector === "string" ? selector : "")
        .length,
    gridSavedRowsSelector(),
  )) as number;
  if (mountedRows > maxRows) {
    throw new Error(
      `Expected ${surface} to mount at most ${maxRows} saved rows, received ${mountedRows}`,
    );
  }
}

export async function isTestIdVisibleWithinGridViewport(
  page: BrowserPageLike,
  surface: string,
  testId: string,
) {
  const target = page.getByTestId(testId);
  if (!(await isLocatorVisible(target))) {
    return false;
  }

  let state: Awaited<ReturnType<typeof readTestIdGridViewportState>>;
  try {
    state = await readTestIdGridViewportState(page, surface, testId);
  } catch (error) {
    if (!(await isLocatorVisible(target))) {
      return false;
    }
    throw error;
  }
  return (
    state.top >= -viewportVisibilityTolerancePx &&
    state.left >= -viewportVisibilityTolerancePx &&
    state.bottom <= state.containerHeight + viewportVisibilityTolerancePx &&
    state.right <= state.containerWidth + viewportVisibilityTolerancePx
  );
}

export async function readTestIdGridViewportState(
  page: BrowserPageLike,
  surface: string,
  testId: string,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluateGrid = requireEvaluate(
    grid,
    `isTestIdVisibleWithinGridViewport(${surface}, ${testId}) requires locator.evaluate() support`,
  );
  return (await evaluateGrid(
    (element, options) => {
      const { scrollportSelector, surface, targetSelector, testId } =
        typeof options === "object" && options !== null
          ? (options as {
              scrollportSelector?: unknown;
              surface?: unknown;
              targetSelector?: unknown;
              testId?: unknown;
            })
          : {};
      const selector =
        typeof scrollportSelector === "string" ? scrollportSelector : "";
      const scrollports = Array.from(
        element.querySelectorAll<HTMLElement>(selector),
      );
      if (scrollports.length !== 1) {
        throw new Error(
          `Expected ${typeof surface === "string" ? surface : "workbook"} grid shell to contain exactly one ${selector} scrollport, received ${scrollports.length}`,
        );
      }
      const gridScrollport = scrollports[0];
      if (gridScrollport === undefined) {
        throw new Error(
          `Expected ${typeof surface === "string" ? surface : "workbook"} grid shell to contain exactly one ${selector} scrollport, received 0`,
        );
      }
      const normalizedTargetSelector =
        typeof targetSelector === "string" ? targetSelector : "";
      const target = element.querySelector<HTMLElement>(
        normalizedTargetSelector,
      );
      if (target === null) {
        throw new Error(
          `Expected ${typeof surface === "string" ? surface : "workbook"} grid shell to contain target ${typeof testId === "string" ? testId : normalizedTargetSelector}`,
        );
      }
      const containerRect = gridScrollport.getBoundingClientRect();
      const targetRect = target.getBoundingClientRect();
      return {
        bottom: targetRect.bottom - containerRect.top,
        containerHeight: containerRect.height,
        containerWidth: containerRect.width,
        left: targetRect.left - containerRect.left,
        right: targetRect.right - containerRect.left,
        top: targetRect.top - containerRect.top,
      };
    },
    {
      scrollportSelector: gridScrollportSelector(),
      surface,
      targetSelector: dataTestIdSelector(testId),
      testId,
    },
  )) as {
    bottom: number;
    containerHeight: number;
    containerWidth: number;
    left: number;
    right: number;
    top: number;
  };
}

export async function assertGroupRowPresentationOnly(options: {
  groupTestId: string;
  page: BrowserPageLike;
  surface: string;
}) {
  const group = options.page.getByTestId(options.groupTestId);
  const evaluate = requireEvaluate(
    group,
    `assertGroupRowPresentationOnly(${options.surface}) requires locator.evaluate() support`,
  );
  const state = (await evaluate((element, rawGroupRowSelector) => {
    const row = element.closest('[role="row"]');
    if (row === null) {
      return { hasRow: false };
    }
    return {
      buttonCount: row.querySelectorAll("button").length,
      editableControlCount: row.querySelectorAll(
        'input, textarea, select, [contenteditable="true"]',
      ).length,
      hasRow: true,
      interactiveCount: row.querySelectorAll(
        'a[href], button, input, textarea, select, [role="button"], [role="textbox"], [contenteditable="true"]',
      ).length,
      matchesGroupRow: row.matches(
        typeof rawGroupRowSelector === "string" ? rawGroupRowSelector : "",
      ),
      recordId: row.getAttribute("data-grid-record-id"),
    };
  }, gridGroupRowSelector())) as
    | { readonly hasRow: false }
    | {
        readonly buttonCount: number;
        readonly editableControlCount: number;
        readonly hasRow: true;
        readonly interactiveCount: number;
        readonly matchesGroupRow: boolean;
        readonly recordId: string | null;
      };
  if (!state.hasRow) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to have an ARIA row ancestor`,
    );
  }
  if (!state.matchesGroupRow) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to match the owner-backed group-row semantic contract`,
    );
  }
  if (state.recordId !== null && state.recordId !== "") {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to omit data-grid-record-id, received ${state.recordId}`,
    );
  }
  if (state.editableControlCount !== 0) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to expose no editable controls, received ${state.editableControlCount}`,
    );
  }
  if (state.buttonCount !== 1 || state.interactiveCount !== 1) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to expose exactly one expand/collapse control, received ${state.buttonCount} buttons and ${state.interactiveCount} interactive elements`,
    );
  }
}
