export type WorkbookSurface = "timeline" | "hosts" | "identities";

export function gridShellTestId(surface: WorkbookSurface): string {
  return `${surface}-grid-shell`;
}

export function gridSortHeaderTestId(
  surface: WorkbookSurface,
  fieldKey: string,
): string {
  return `${surface}-sort-${sanitizeToken(fieldKey)}`;
}

export function gridFilterChipTestId(
  surface: WorkbookSurface,
  fieldKey: string,
): string {
  return `${surface}-filter-chip-${sanitizeToken(fieldKey)}`;
}

export function gridFilterFieldTestId(surface: WorkbookSurface): string {
  return `${surface}-filter-field`;
}

export function gridFilterValueTestId(surface: WorkbookSurface): string {
  return `${surface}-filter-value`;
}

export function gridFilterApplyTestId(surface: WorkbookSurface): string {
  return `${surface}-filter-apply`;
}

export function gridGroupingSelectTestId(surface: WorkbookSurface): string {
  return `${surface}-group-by`;
}

export function gridGroupRowTestId(
  surface: WorkbookSurface,
  fieldKey: string,
  value: string,
): string {
  return `${surface}-group-${sanitizeToken(fieldKey)}-${sanitizeToken(value)}`;
}

export function rowCellTestId(recordId: string, fieldKey: string): string {
  return `row-${recordId}-${fieldKey}`;
}

export function rowInspectButtonTestId(recordId: string): string {
  return `row-${recordId}-inspect`;
}

export function draftCellTestId(fieldKey: string): string {
  return `draft-row-${fieldKey}`;
}

export function relationshipItemsTestId(
  recordId: string,
  relationshipKey: string,
): string {
  return `row-${recordId}-${relationshipKey}-items`;
}

export function timelineRowVersionTestId(recordId: string): string {
  return `row-${recordId}-row-version`;
}

export function pasteMatrixText(matrix: readonly (readonly string[])[]): string {
  return matrix.map((row) => row.join("\t")).join("\n");
}

type BrowserLocator = {
  click: () => Promise<void>;
  evaluate?: (pageFunction: (element: Element) => unknown) => Promise<unknown>;
  fill: (value: string) => Promise<void>;
  press?: (value: string) => Promise<void>;
  scrollIntoViewIfNeeded?: () => Promise<void>;
  selectOption?: (value: string) => Promise<void>;
};

type BrowserPageLike = {
  getByTestId: (value: string) => BrowserLocator;
};

export async function sortByHeader(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
) {
  await page.getByTestId(gridSortHeaderTestId(surface, fieldKey)).click();
}

export async function applyFilterChip(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
  value: string,
) {
  await page.getByTestId(gridFilterFieldTestId(surface)).selectOption?.(
    fieldKey,
  );
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
  await page.getByTestId(gridGroupingSelectTestId(surface)).selectOption?.(
    fieldKey,
  );
}

export async function scrollToCell(
  page: BrowserPageLike,
  recordId: string,
  fieldKey: string,
) {
  await page
    .getByTestId(rowCellTestId(recordId, fieldKey))
    .scrollIntoViewIfNeeded?.();
}

export function assertAnchorTestId(recordId: string, fieldKey: string): string {
  return rowCellTestId(recordId, fieldKey);
}

export async function readGridScroll(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `readGridScroll(${surface}) requires locator.evaluate() support`,
  );
  return evaluate((element) => {
    const gridShell = element as HTMLDivElement;
    return {
      left: gridShell.scrollLeft,
      top: gridShell.scrollTop,
    };
  }) as Promise<{ left: number; top: number }>;
}

export async function scrollGridToBottom(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `scrollGridToBottom(${surface}) requires locator.evaluate() support`,
  );
  return evaluate((element) => {
    const gridShell = element as HTMLDivElement;
    gridShell.scrollTop = gridShell.scrollHeight;
    return {
      left: gridShell.scrollLeft,
      top: gridShell.scrollTop,
    };
  }) as Promise<{ left: number; top: number }>;
}

export async function isTestIdVisibleWithinGridViewport(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  testId: string,
) {
  const state = await readTestIdGridViewportState(page, surface, testId);
  return (
    state.top >= -viewportVisibilityTolerancePx &&
    state.left >= -viewportVisibilityTolerancePx &&
    state.bottom <= state.containerHeight + viewportVisibilityTolerancePx &&
    state.right <= state.containerWidth + viewportVisibilityTolerancePx
  );
}

async function readTestIdGridViewportState(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  testId: string,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const target = page.getByTestId(testId);
  const evaluateGrid = requireEvaluate(
    grid,
    `isTestIdVisibleWithinGridViewport(${surface}, ${testId}) requires locator.evaluate() support`,
  );
  const evaluateTarget = requireEvaluate(
    target,
    `isTestIdVisibleWithinGridViewport(${surface}, ${testId}) requires locator.evaluate() support`,
  );
  const containerRect = (await evaluateGrid((element) => {
    const rect = element.getBoundingClientRect();
    return {
      bottom: rect.bottom,
      height: rect.height,
      left: rect.left,
      right: rect.right,
      top: rect.top,
      width: rect.width,
    };
  })) as {
    bottom: number;
    height: number;
    left: number;
    right: number;
    top: number;
    width: number;
  };
  const elementRect = (await evaluateTarget((element) => {
    const rect = element.getBoundingClientRect();
    return {
      bottom: rect.bottom,
      height: rect.height,
      left: rect.left,
      right: rect.right,
      top: rect.top,
      width: rect.width,
    };
  })) as {
    bottom: number;
    height: number;
    left: number;
    right: number;
    top: number;
    width: number;
  };

  const top = elementRect.top - containerRect.top;
  const left = elementRect.left - containerRect.left;
  const bottom = elementRect.bottom - containerRect.top;
  const right = elementRect.right - containerRect.left;
  return {
    bottom,
    containerHeight: containerRect.height,
    containerWidth: containerRect.width,
    left,
    right,
    top,
  };
}

export async function assertGridFocusContinuity(options: {
  focusTestId: string;
  intervalMs?: number;
  page: BrowserPageLike;
  preservedScroll: { left: number; top: number };
  requireExactHorizontalScroll?: boolean;
  requireExactVerticalScroll?: boolean;
  surface: WorkbookSurface;
  timeoutMs?: number;
}) {
  const {
    focusTestId,
    intervalMs = 50,
    page,
    preservedScroll,
    requireExactHorizontalScroll = true,
    requireExactVerticalScroll = true,
    surface,
    timeoutMs = 3_000,
  } = options;
  const deadline = Date.now() + timeoutMs;
  let lastError: Error | null = null;

  while (Date.now() <= deadline) {
    try {
      const focusTarget = page.getByTestId(focusTestId);
      const evaluateFocusTarget = requireEvaluate(
        focusTarget,
        `assertGridFocusContinuity(${surface}, ${focusTestId}) requires locator.evaluate() support`,
      );
      const isFocused = (await evaluateFocusTarget(
        (element) => document.activeElement === element,
      )) as boolean;
      if (!isFocused) {
        throw new Error(
          `Expected ${focusTestId} to be focused within the ${surface} grid continuity restore`,
        );
      }
      const viewportState = await readTestIdGridViewportState(
        page,
        surface,
        focusTestId,
      );
      const isVisibleWithinViewport =
        viewportState.top >= -viewportVisibilityTolerancePx &&
        viewportState.left >= -viewportVisibilityTolerancePx &&
        viewportState.bottom <=
          viewportState.containerHeight + viewportVisibilityTolerancePx &&
        viewportState.right <=
          viewportState.containerWidth + viewportVisibilityTolerancePx;
      if (!isVisibleWithinViewport) {
        throw new Error(
          `Expected ${focusTestId} to remain fully visible within the ${surface} grid viewport (top=${viewportState.top}, bottom=${viewportState.bottom}, left=${viewportState.left}, right=${viewportState.right}, containerHeight=${viewportState.containerHeight}, containerWidth=${viewportState.containerWidth})`,
        );
      }
      const currentScroll = await readGridScroll(page, surface);
      if (
        requireExactVerticalScroll &&
        currentScroll.top !== preservedScroll.top
      ) {
        throw new Error(
          `Expected ${surface} vertical scroll ${preservedScroll.top}, received ${currentScroll.top}`,
        );
      }
      if (
        requireExactHorizontalScroll &&
        currentScroll.left !== preservedScroll.left
      ) {
        throw new Error(
          `Expected ${surface} horizontal scroll ${preservedScroll.left}, received ${currentScroll.left}`,
        );
      }
      return;
    } catch (error) {
      lastError =
        error instanceof Error ? error : new Error(String(error ?? "unknown"));
      if (Date.now() > deadline) {
        break;
      }
      await delay(intervalMs);
    }
  }
  throw lastError ?? new Error("Grid continuity assertion timed out");
}

function delay(durationMs: number) {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, durationMs);
  });
}

const viewportVisibilityTolerancePx = 1;

function requireEvaluate(
  locator: BrowserLocator,
  message: string,
): (pageFunction: (element: Element) => unknown) => Promise<unknown> {
  if (typeof locator.evaluate !== "function") {
    throw new Error(message);
  }
  return (pageFunction) => locator.evaluate?.(pageFunction) as Promise<unknown>;
}

function sanitizeToken(value: string): string {
  return value.replace(/[^a-zA-Z0-9_-]+/gu, "-");
}
