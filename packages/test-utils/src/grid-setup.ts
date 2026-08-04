import {
  dataTestIdSelector,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";

import {
  type BrowserLocator,
  type BrowserPageLike,
  delay,
  isLocatorVisible,
  requireEvaluate,
} from "./browser";
import {
  buildGridAxisScanOffsets,
  createGridTargetScanObservation,
  formatGridTargetScanFailure,
  observeGridTargetScanState,
  readGridScroll,
  readGridScrollDiagnostics,
} from "./grid-diagnostics";

type GridScrollAction =
  | { kind: "bottom" }
  | { kind: "offset"; left?: number; top: number };

export function scrollGridToBottom(page: BrowserPageLike, surface: string) {
  return applyGridScrollAction(page, surface, { kind: "bottom" });
}

export function scrollGridToOffset(
  page: BrowserPageLike,
  surface: string,
  top: number,
) {
  return applyGridScrollAction(page, surface, { kind: "offset", top });
}

function scrollGridToPosition(
  page: BrowserPageLike,
  surface: string,
  top: number,
  left: number,
) {
  return applyGridScrollAction(page, surface, {
    kind: "offset",
    left,
    top,
  });
}

async function applyGridScrollAction(
  page: BrowserPageLike,
  surface: string,
  action: GridScrollAction,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const operation =
    action.kind === "bottom"
      ? "scrollGridToBottom"
      : action.left === undefined
        ? "scrollGridToOffset"
        : "scrollGridToPosition";
  const evaluate = requireEvaluate(
    grid,
    `${operation}(${surface}) requires locator.evaluate() support`,
  );
  return (await evaluate(
    (element, rawOptions) => {
      const options =
        typeof rawOptions === "object" && rawOptions !== null
          ? (rawOptions as {
              kind?: unknown;
              left?: unknown;
              scrollportSelector?: unknown;
              surface?: unknown;
              top?: unknown;
            })
          : {};
      const selector =
        typeof options.scrollportSelector === "string"
          ? options.scrollportSelector
          : "";
      const normalizedSurface =
        typeof options.surface === "string" ? options.surface : "workbook";
      const scrollports = Array.from(
        element.querySelectorAll<HTMLElement>(selector),
      );
      if (scrollports.length !== 1) {
        throw new Error(
          `Expected ${normalizedSurface} grid shell to contain exactly one ${selector} scrollport, received ${scrollports.length}`,
        );
      }
      const scrollport = scrollports[0];
      if (scrollport === undefined) {
        throw new Error(
          `Expected ${normalizedSurface} grid shell to contain exactly one ${selector} scrollport, received 0`,
        );
      }
      if (options.kind === "bottom") {
        scrollport.scrollTop = scrollport.scrollHeight;
      } else {
        scrollport.scrollTop =
          typeof options.top === "number" && Number.isFinite(options.top)
            ? options.top
            : 0;
        if (typeof options.left === "number" && Number.isFinite(options.left)) {
          scrollport.scrollLeft = options.left;
        }
      }
      return { left: scrollport.scrollLeft, top: scrollport.scrollTop };
    },
    {
      ...action,
      scrollportSelector: gridScrollportSelector(),
      surface,
    },
  )) as { left: number; top: number };
}

export async function scrollGridCellIntoView(options: {
  cellKey: string;
  intervalMs?: number;
  page: BrowserPageLike;
  recordId: string;
  surface: string;
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
  surface: string;
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
  if (await alignVisibleGridTarget({ page, surface, target, targetTestId })) {
    return readGridScroll(page, surface);
  }

  const retryIntervalMs = Math.max(intervalMs, 0);
  const deadline = Date.now() + Math.max(timeoutMs, 0);
  const observation = createGridTargetScanObservation();

  for (;;) {
    if (await alignVisibleGridTarget({ page, surface, target, targetTestId })) {
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
        if (
          await alignVisibleGridTarget({ page, surface, target, targetTestId })
        ) {
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
      if (scanRangeGrew) break;
    }

    if (scanRangeGrew) continue;

    observation.completedScanCycles += 1;
    observation.completedScanMaxTop = Math.max(
      observation.completedScanMaxTop,
      scanMaxTop,
    );
    observation.completedScanMaxLeft = Math.max(
      observation.completedScanMaxLeft,
      scanMaxLeft,
    );
    if (Date.now() > deadline) break;
    await waitForGridTargetRetry(retryIntervalMs);
  }

  const finalState = await readGridScrollDiagnostics(page, surface);
  observeGridTargetScanState(observation, finalState);
  throw new Error(
    formatGridTargetScanFailure({
      observation,
      state: finalState,
      surface,
      targetTestId,
    }),
  );
}

async function alignVisibleGridTarget(options: {
  page: BrowserPageLike;
  surface: string;
  target: BrowserLocator;
  targetTestId: string;
}) {
  const { page, surface, target, targetTestId } = options;
  if (!(await isLocatorVisible(target))) return false;

  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluateGrid = requireEvaluate(
    grid,
    `scrollGridTargetIntoView(${surface}, ${targetTestId}) requires locator.evaluate() support`,
  );
  const aligned = (await evaluateGrid((element, rawTargetSelector) => {
    const targetSelector =
      typeof rawTargetSelector === "string" ? rawTargetSelector : "";
    const mountedTarget =
      element.ownerDocument.querySelector<HTMLElement>(targetSelector);
    if (mountedTarget === null || !mountedTarget.isConnected) return false;
    mountedTarget.scrollIntoView({ block: "nearest", inline: "nearest" });
    return true;
  }, dataTestIdSelector(targetTestId))) as boolean;
  if (!aligned) return false;
  return isLocatorVisible(target);
}

function waitForGridTargetRetry(intervalMs: number) {
  if (intervalMs <= 0) return Promise.resolve();
  return delay(intervalMs);
}
