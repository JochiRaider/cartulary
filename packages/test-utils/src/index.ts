import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridSavedRowsSelector,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";

export function pasteMatrixText(
  matrix: readonly (readonly string[])[],
): string {
  return matrix.map((row) => row.join("\t")).join("\n");
}

type BrowserLocator = {
  click: () => Promise<void>;
  evaluate?: (
    pageFunction: (element: Element, arg?: unknown) => unknown,
    arg?: unknown,
  ) => Promise<unknown>;
  fill: (value: string) => Promise<void>;
  isVisible?: () => Promise<boolean>;
  press?: (value: string) => Promise<void>;
  scrollIntoViewIfNeeded?: () => Promise<void>;
  selectOption?: (value: string | readonly string[]) => Promise<unknown>;
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
  return readScrollSnapshot(evaluate, surface);
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
  return readScrollSnapshot(evaluate, surface, { kind: "bottom" });
}

export async function scrollGridToOffset(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  top: number,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `scrollGridToOffset(${surface}) requires locator.evaluate() support`,
  );
  return readScrollSnapshot(evaluate, surface, { kind: "offset", top });
}

export async function scrollGridTargetIntoView(options: {
  intervalMs?: number;
  page: BrowserPageLike;
  surface: WorkbookSurface;
  targetTestId: string;
  timeoutMs?: number;
}) {
  const {
    intervalMs = 50,
    page,
    surface,
    targetTestId,
    timeoutMs = 3_000,
  } = options;
  const target = page.getByTestId(targetTestId);
  if (await isLocatorVisible(target)) {
    return readGridScroll(page, surface);
  }

  const retryIntervalMs = Math.max(intervalMs, 0);
  const deadline = Date.now() + Math.max(timeoutMs, 0);
  const observation = createGridTargetScanObservation();

  for (;;) {
    if (await isLocatorVisible(target)) {
      await target.scrollIntoViewIfNeeded?.();
      return readGridScroll(page, surface);
    }

    const state = await readGridScrollDiagnostics(page, surface);
    observeGridTargetScanState(observation, state);
    const scanOffsets = buildGridScanOffsets(state);

    for (const top of scanOffsets) {
      await scrollGridToOffset(page, surface, top);
      observation.scrollAttempts += 1;
      await waitForGridTargetRetry(retryIntervalMs);
      if (await isLocatorVisible(target)) {
        await target.scrollIntoViewIfNeeded?.();
        return readGridScroll(page, surface);
      }
      if (Date.now() > deadline) {
        break;
      }
    }

    if (Date.now() > deadline) {
      break;
    }
    await waitForGridTargetRetry(retryIntervalMs);
  }

  const finalState = await readGridScrollDiagnostics(page, surface);
  observeGridTargetScanState(observation, finalState);
  throw new Error(
    [
      `Expected ${targetTestId} to become visible in the ${surface} grid viewport after scanning virtualized rows.`,
      `scrollTop=${finalState.top}`,
      `scrollLeft=${finalState.left}`,
      `clientHeight=${finalState.clientHeight}`,
      `scrollHeight=${finalState.scrollHeight}`,
      `maxTop=${finalState.maxTop}`,
      `mountedRowIds=${finalState.mountedRowIds.join(",") || "(none)"}`,
      `scanCycles=${observation.scanCycles}`,
      `scrollAttempts=${observation.scrollAttempts}`,
      `observedScrollable=${observation.scrollableScanCycles > 0}`,
      `observedMaxTop=${observation.maxTop}`,
      `observedMountedRowIds=${
        Array.from(observation.mountedRowIds).join(",") || "(none)"
      }`,
    ].join(" "),
  );
}

export async function assertMountedGridRowCountAtMost(options: {
  maxRows: number;
  page: BrowserPageLike;
  surface: WorkbookSurface;
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

export async function assertMarkerAnchoredToGridTarget(options: {
  markerTestId: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
  targetTestId: string;
}) {
  const { markerTestId, page, surface, targetTestId } = options;
  const markerVisible = await isTestIdVisibleWithinGridViewport(
    page,
    surface,
    markerTestId,
  );
  if (!markerVisible) {
    throw new Error(
      `Expected marker ${markerTestId} to be visible in the ${surface} grid viewport`,
    );
  }
  const targetVisible = await isTestIdVisibleWithinGridViewport(
    page,
    surface,
    targetTestId,
  );
  if (!targetVisible) {
    throw new Error(
      `Expected target ${targetTestId} to be visible in the ${surface} grid viewport`,
    );
  }
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
  const containerRect = (await evaluateGrid(
    (element, options) => {
      const { scrollportSelector, surface } =
        typeof options === "object" && options !== null
          ? (options as { scrollportSelector?: unknown; surface?: unknown })
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
      const rect = gridScrollport.getBoundingClientRect();
      return {
        bottom: rect.bottom,
        height: rect.height,
        left: rect.left,
        right: rect.right,
        top: rect.top,
        width: rect.width,
      };
    },
    { scrollportSelector: gridScrollportSelector(), surface },
  )) as {
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

/**
 * Base continuity requires the focused control to remain focused and fully
 * visible. Exact scroll preservation is stricter and must be opted into.
 */
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
    requireExactHorizontalScroll = false,
    requireExactVerticalScroll = false,
    surface,
    timeoutMs = 3_000,
  } = options;
  const retryIntervalMs = Math.max(intervalMs, 0);
  const maxAttempts = Math.max(
    1,
    Math.ceil(timeoutMs / Math.max(retryIntervalMs, 1)) + 1,
  );
  const deadline = Date.now() + timeoutMs;
  let lastError: Error | null = null;

  for (let attempt = 0; attempt < maxAttempts; attempt += 1) {
    try {
      await assertGridFocusContinuityOnce({
        focusTestId,
        page,
        preservedScroll,
        requireExactHorizontalScroll,
        requireExactVerticalScroll,
        surface,
      });
      return;
    } catch (error) {
      lastError =
        error instanceof Error ? error : new Error(String(error ?? "unknown"));
      if (attempt === maxAttempts - 1 || Date.now() > deadline) {
        break;
      }
      await waitForGridContinuityRetry(retryIntervalMs);
    }
  }
  throw lastError ?? new Error("Grid continuity assertion timed out");
}

async function assertGridFocusContinuityOnce(options: {
  focusTestId: string;
  page: BrowserPageLike;
  preservedScroll: { left: number; top: number };
  requireExactHorizontalScroll: boolean;
  requireExactVerticalScroll: boolean;
  surface: WorkbookSurface;
}) {
  const {
    focusTestId,
    page,
    preservedScroll,
    requireExactHorizontalScroll,
    requireExactVerticalScroll,
    surface,
  } = options;
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
  if (requireExactVerticalScroll && currentScroll.top !== preservedScroll.top) {
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
}

function delay(durationMs: number) {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, durationMs);
  });
}

function waitForGridContinuityRetry(intervalMs: number) {
  if (intervalMs <= 0) {
    return Promise.resolve();
  }
  return delay(intervalMs);
}

type BrowserEvaluate = NonNullable<BrowserLocator["evaluate"]>;

type GridScrollAction =
  | { kind: "bottom" }
  | { kind: "none" }
  | { kind: "offset"; top: number };

type GridScrollDiagnostics = {
  readonly clientHeight: number;
  readonly clientWidth: number;
  readonly left: number;
  readonly maxTop: number;
  readonly mountedRowIds: readonly string[];
  readonly scrollHeight: number;
  readonly scrollWidth: number;
  readonly top: number;
};

async function readScrollSnapshot(
  evaluate: BrowserEvaluate,
  surface: WorkbookSurface,
  action: GridScrollAction = { kind: "none" },
) {
  const state = (await evaluate(readGridScrollState, {
    ...action,
    savedRowsSelector: gridSavedRowsSelector(),
    scrollportSelector: gridScrollportSelector(),
    surface,
  })) as GridScrollDiagnostics;
  return {
    left: state.left,
    top: state.top,
  };
}

async function readGridScrollDiagnostics(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `readGridScrollDiagnostics(${surface}) requires locator.evaluate() support`,
  );
  return (await evaluate(readGridScrollState, {
    kind: "none",
    savedRowsSelector: gridSavedRowsSelector(),
    scrollportSelector: gridScrollportSelector(),
    surface,
  })) as GridScrollDiagnostics;
}

function readGridScrollState(
  element: Element,
  rawAction?: unknown,
): GridScrollDiagnostics {
  const action =
    typeof rawAction === "object" && rawAction !== null
      ? (rawAction as {
          kind?: unknown;
          savedRowsSelector?: unknown;
          scrollportSelector?: unknown;
          surface?: unknown;
          top?: unknown;
        })
      : { kind: "none" };
  const scrollportSelector =
    typeof action.scrollportSelector === "string"
      ? action.scrollportSelector
      : "";
  const surface =
    typeof action.surface === "string" ? action.surface : "workbook";
  const scrollports = Array.from(
    element.querySelectorAll<HTMLElement>(scrollportSelector),
  );
  if (scrollports.length !== 1) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
    );
  }
  const gridScrollport = scrollports[0];
  if (gridScrollport === undefined) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received 0`,
    );
  }
  if (action.kind === "bottom") {
    gridScrollport.scrollTop = gridScrollport.scrollHeight;
  }
  if (action.kind === "offset") {
    const nextTop =
      typeof action.top === "number" && Number.isFinite(action.top)
        ? action.top
        : 0;
    gridScrollport.scrollTop = nextTop;
  }

  const scrollHeight = gridScrollport.scrollHeight;
  const clientHeight = gridScrollport.clientHeight;
  const savedRowsSelector =
    typeof action.savedRowsSelector === "string"
      ? action.savedRowsSelector
      : '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])';
  return {
    clientHeight,
    clientWidth: gridScrollport.clientWidth,
    left: gridScrollport.scrollLeft,
    maxTop: Math.max(0, scrollHeight - clientHeight),
    mountedRowIds: Array.from(
      element.querySelectorAll<HTMLElement>(savedRowsSelector),
    ).map((row) => row.getAttribute("data-grid-record-id") ?? ""),
    scrollHeight,
    scrollWidth: gridScrollport.scrollWidth,
    top: gridScrollport.scrollTop,
  };
}

function buildGridScanOffsets(state: GridScrollDiagnostics) {
  const maxTop = Math.max(0, state.maxTop);
  if (maxTop === 0) {
    return [0];
  }
  const step = Math.max(1, Math.floor(Math.max(state.clientHeight, 1) / 2));
  const offsets = [0];
  for (let top = step; top < maxTop; top += step) {
    offsets.push(top);
  }
  offsets.push(maxTop);
  return Array.from(new Set(offsets));
}

type GridTargetScanObservation = {
  maxTop: number;
  mountedRowIds: Set<string>;
  scanCycles: number;
  scrollableScanCycles: number;
  scrollAttempts: number;
};

function createGridTargetScanObservation(): GridTargetScanObservation {
  return {
    maxTop: 0,
    mountedRowIds: new Set(),
    scanCycles: 0,
    scrollableScanCycles: 0,
    scrollAttempts: 0,
  };
}

function observeGridTargetScanState(
  observation: GridTargetScanObservation,
  state: GridScrollDiagnostics,
) {
  observation.scanCycles += 1;
  observation.maxTop = Math.max(observation.maxTop, state.maxTop);
  if (state.maxTop > 0) {
    observation.scrollableScanCycles += 1;
  }
  for (const rowId of state.mountedRowIds) {
    if (rowId !== "") {
      observation.mountedRowIds.add(rowId);
    }
  }
}

function waitForGridTargetRetry(intervalMs: number) {
  if (intervalMs <= 0) {
    return Promise.resolve();
  }
  return delay(intervalMs);
}

async function isLocatorVisible(locator: BrowserLocator) {
  if (typeof locator.isVisible === "function") {
    return locator.isVisible();
  }
  try {
    const evaluate = requireEvaluate(
      locator,
      "isLocatorVisible requires locator.evaluate() support",
    );
    return Boolean(
      await evaluate((element) => {
        const rect = element.getBoundingClientRect();
        return (
          element.isConnected &&
          rect.width > 0 &&
          rect.height > 0 &&
          getComputedStyle(element).visibility !== "hidden"
        );
      }),
    );
  } catch {
    return false;
  }
}

const viewportVisibilityTolerancePx = 1;

function requireEvaluate(
  locator: BrowserLocator,
  message: string,
): BrowserEvaluate {
  if (typeof locator.evaluate !== "function") {
    throw new Error(message);
  }
  return (pageFunction, arg) =>
    locator.evaluate?.(pageFunction, arg) as Promise<unknown>;
}
