import type { WorkbookSurface } from "@cartulary/ui-contracts";

import { type BrowserPageLike, delay, requireEvaluate } from "./browser";
import {
  readGridScroll,
  readTestIdGridViewportState,
  viewportVisibilityTolerancePx,
} from "./scrolling";

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

function waitForGridContinuityRetry(intervalMs: number) {
  if (intervalMs <= 0) {
    return Promise.resolve();
  }
  return delay(intervalMs);
}
