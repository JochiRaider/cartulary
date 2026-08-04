import {
  gridSavedRowsSelector,
  gridScrollportSelector,
  gridShellTestId,
} from "@cartulary/ui-contracts";

import { type BrowserPageLike, requireEvaluate } from "./browser";

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

export type GridTargetScanObservation = {
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

export async function readGridScroll(page: BrowserPageLike, surface: string) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `readGridScroll(${surface}) requires locator.evaluate() support`,
  );
  const state = (await evaluate(readGridScrollState, {
    savedRowsSelector: gridSavedRowsSelector(),
    scrollportSelector: gridScrollportSelector(),
    surface,
  })) as GridScrollDiagnostics;
  return { left: state.left, top: state.top };
}

export async function readGridScrollDiagnostics(
  page: BrowserPageLike,
  surface: string,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `readGridScrollDiagnostics(${surface}) requires locator.evaluate() support`,
  );
  return (await evaluate(readGridScrollState, {
    savedRowsSelector: gridSavedRowsSelector(),
    scrollportSelector: gridScrollportSelector(),
    surface,
  })) as GridScrollDiagnostics;
}

function readGridScrollState(
  element: Element,
  rawOptions?: unknown,
): GridScrollDiagnostics {
  const options =
    typeof rawOptions === "object" && rawOptions !== null
      ? (rawOptions as {
          savedRowsSelector?: unknown;
          scrollportSelector?: unknown;
          surface?: unknown;
        })
      : {};
  const scrollportSelector =
    typeof options.scrollportSelector === "string"
      ? options.scrollportSelector
      : "";
  const surface =
    typeof options.surface === "string" ? options.surface : "workbook";
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
  const scrollHeight = gridScrollport.scrollHeight;
  const clientHeight = gridScrollport.clientHeight;
  const savedRowsSelector =
    typeof options.savedRowsSelector === "string"
      ? options.savedRowsSelector
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

export function buildGridAxisScanOffsets(
  maximum: number,
  viewportSize: number,
) {
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

export function createGridTargetScanObservation(): GridTargetScanObservation {
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

export function observeGridTargetScanState(
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

export function formatGridTargetScanFailure(options: {
  observation: GridTargetScanObservation;
  state: GridScrollDiagnostics;
  surface: string;
  targetTestId: string;
}) {
  const { observation, state, surface, targetTestId } = options;
  return [
    `Expected ${targetTestId} to become visible in the ${surface} grid viewport after scanning virtualized rows.`,
    `scrollTop=${state.top}`,
    `scrollLeft=${state.left}`,
    `clientHeight=${state.clientHeight}`,
    `clientWidth=${state.clientWidth}`,
    `scrollHeight=${state.scrollHeight}`,
    `scrollWidth=${state.scrollWidth}`,
    `maxLeft=${state.maxLeft}`,
    `maxTop=${state.maxTop}`,
    `mountedRowIds=${state.mountedRowIds.join(",") || "(none)"}`,
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
  ].join(" ");
}
