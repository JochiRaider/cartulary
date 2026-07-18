import {
  gridSavedRowsSelector,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";

import {
  type BrowserEvaluate,
  type BrowserLocator,
  type BrowserPageLike,
  delay,
  isLocatorVisible,
  requireEvaluate,
} from "./browser";

export const viewportVisibilityTolerancePx = 1;

type GridScrollAction =
  | { kind: "bottom" }
  | { kind: "none" }
  | { kind: "offset"; left?: number; top: number };

export type GridScrollDiagnostics = {
  readonly clientHeight: number;
  readonly clientWidth: number;
  readonly left: number;
  readonly maxLeft: number;
  readonly maxTop: number;
  readonly mountedRowIds: readonly string[];
  readonly scrollHeight: number;
  readonly scrollWidth: number;
  readonly top: number;
};

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

async function scrollGridToPosition(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  top: number,
  left: number,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `scrollGridToPosition(${surface}) requires locator.evaluate() support`,
  );
  return readScrollSnapshot(evaluate, surface, { kind: "offset", left, top });
}

export async function scrollGridCellIntoView(options: {
  cellKey: string;
  intervalMs?: number;
  page: BrowserPageLike;
  recordId: string;
  surface: WorkbookSurface;
  timeoutMs?: number;
}) {
  const scanOptions: Parameters<typeof scrollGridTargetIntoView>[0] = {
    page: options.page,
    surface: options.surface,
    targetTestId: rowCellTestId(options.recordId, options.cellKey),
  };
  if (options.intervalMs !== undefined) {
    scanOptions.intervalMs = options.intervalMs;
  }
  if (options.timeoutMs !== undefined) {
    scanOptions.timeoutMs = options.timeoutMs;
  }
  return scrollGridTargetIntoView(scanOptions);
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
  if (await alignVisibleGridTarget(target)) {
    return readGridScroll(page, surface);
  }

  const retryIntervalMs = Math.max(intervalMs, 0);
  const deadline = Date.now() + Math.max(timeoutMs, 0);
  const observation = createGridTargetScanObservation();

  for (;;) {
    if (await alignVisibleGridTarget(target)) {
      return readGridScroll(page, surface);
    }

    let state = await readGridScrollDiagnostics(page, surface);
    observeGridTargetScanState(observation, state);
    let scanRangeGrew = false;
    const scanMaxLeft = state.maxLeft;
    const scanMaxTop = state.maxTop;
    const scanTopOffsets = buildGridAxisScanOffsets(
      state.maxTop,
      state.clientHeight,
    );
    const scanLeftOffsets = buildGridAxisScanOffsets(
      state.maxLeft,
      state.clientWidth,
    );

    for (const left of scanLeftOffsets) {
      for (const top of scanTopOffsets) {
        await scrollGridToPosition(page, surface, top, left);
        observation.scrollAttempts += 1;
        await waitForGridTargetRetry(retryIntervalMs);
        if (await alignVisibleGridTarget(target)) {
          return readGridScroll(page, surface);
        }
        state = await readGridScrollDiagnostics(page, surface);
        observeGridTargetScanState(observation, state);
        if (state.maxTop > scanMaxTop || state.maxLeft > scanMaxLeft) {
          observation.scrollRangeGrowths += 1;
          scanRangeGrew = true;
          break;
        }
      }
      if (scanRangeGrew) {
        break;
      }
    }

    if (scanRangeGrew) {
      continue;
    }

    observation.completedScanCycles += 1;
    observation.completedScanMaxTop = Math.max(
      observation.completedScanMaxTop,
      scanMaxTop,
    );
    observation.completedScanMaxLeft = Math.max(
      observation.completedScanMaxLeft,
      scanMaxLeft,
    );
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
      `clientWidth=${finalState.clientWidth}`,
      `scrollHeight=${finalState.scrollHeight}`,
      `scrollWidth=${finalState.scrollWidth}`,
      `maxLeft=${finalState.maxLeft}`,
      `maxTop=${finalState.maxTop}`,
      `mountedRowIds=${finalState.mountedRowIds.join(",") || "(none)"}`,
      `scanCycles=${observation.scanCycles}`,
      `completedScanCycles=${observation.completedScanCycles}`,
      `scrollAttempts=${observation.scrollAttempts}`,
      `scrollRangeGrowths=${observation.scrollRangeGrowths}`,
      `observedScrollable=${observation.scrollableScanCycles > 0}`,
      `observedMaxLeft=${observation.maxLeft}`,
      `observedMaxTop=${observation.maxTop}`,
      `completedScanMaxLeft=${observation.completedScanMaxLeft}`,
      `completedScanMaxTop=${observation.completedScanMaxTop}`,
      `observedMountedRowIds=${
        Array.from(observation.mountedRowIds).join(",") || "(none)"
      }`,
    ].join(" "),
  );
}

async function alignVisibleGridTarget(target: BrowserLocator) {
  if (!(await isLocatorVisible(target))) {
    return false;
  }
  try {
    await target.scrollIntoViewIfNeeded?.();
  } catch {
    // RDG may remount a virtualized cell while Playwright is waiting for that
    // element to settle. Treat the detached node as a scan retry; the stable
    // semantic locator will resolve the replacement on the next attempt.
    return false;
  }
  return isLocatorVisible(target);
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

export async function isTestIdVisibleWithinGridViewport(
  page: BrowserPageLike,
  surface: WorkbookSurface,
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
    // Virtualized rows can unmount between the visibility check and geometry
    // read. That target is outside the viewport by definition. Preserve real
    // grid-contract failures when the target remains mounted.
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

export async function readGridScrollDiagnostics(
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
          left?: unknown;
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
    if (typeof action.left === "number" && Number.isFinite(action.left)) {
      gridScrollport.scrollLeft = action.left;
    }
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
    maxLeft: Math.max(
      0,
      gridScrollport.scrollWidth - gridScrollport.clientWidth,
    ),
    maxTop: Math.max(0, scrollHeight - clientHeight),
    mountedRowIds: Array.from(
      element.querySelectorAll<HTMLElement>(savedRowsSelector),
    ).map((row) => row.getAttribute("data-grid-record-id") ?? ""),
    scrollHeight,
    scrollWidth: gridScrollport.scrollWidth,
    top: gridScrollport.scrollTop,
  };
}

function buildGridAxisScanOffsets(maximum: number, viewportSize: number) {
  const maxOffset = Math.max(0, maximum);
  if (maxOffset === 0) {
    return [0];
  }
  const step = Math.max(1, Math.floor(Math.max(viewportSize, 1) / 2));
  const offsets = [0];
  for (let offset = step; offset < maxOffset; offset += step) {
    offsets.push(offset);
  }
  offsets.push(maxOffset);
  return Array.from(new Set(offsets));
}

type GridTargetScanObservation = {
  completedScanCycles: number;
  completedScanMaxLeft: number;
  completedScanMaxTop: number;
  maxLeft: number;
  maxTop: number;
  mountedRowIds: Set<string>;
  scanCycles: number;
  scrollRangeGrowths: number;
  scrollableScanCycles: number;
  scrollAttempts: number;
};

function createGridTargetScanObservation(): GridTargetScanObservation {
  return {
    completedScanCycles: 0,
    completedScanMaxLeft: 0,
    completedScanMaxTop: 0,
    maxLeft: 0,
    maxTop: 0,
    mountedRowIds: new Set(),
    scanCycles: 0,
    scrollRangeGrowths: 0,
    scrollableScanCycles: 0,
    scrollAttempts: 0,
  };
}

function observeGridTargetScanState(
  observation: GridTargetScanObservation,
  state: GridScrollDiagnostics,
) {
  observation.scanCycles += 1;
  observation.maxLeft = Math.max(observation.maxLeft, state.maxLeft);
  observation.maxTop = Math.max(observation.maxTop, state.maxTop);
  if (state.maxTop > 0 || state.maxLeft > 0) {
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
