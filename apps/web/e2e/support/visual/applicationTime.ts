import type { Page } from "@playwright/test";

type VisualFrameScope = typeof globalThis & {
  __cartularyVisualFrame?: (callback: FrameRequestCallback) => number;
  __cartularyVisualFrameNative?: boolean;
};

// Holding application time jumps the virtual clock this far ahead, firing any
// timers already due in that window, so a slow hand-off cannot target the past.
const holdLeadMs = 1_000;

// Visual captures wait for rendered frames. Playwright's virtual clock replaces
// requestAnimationFrame along with application timers, so every visual context
// records the native frame source before any clock can be installed.
export async function installVisualFrameSource(page: Page) {
  await page.context().addInitScript(() => {
    const scope = globalThis as VisualFrameScope;
    if (scope.__cartularyVisualFrame) return;
    const frame = scope.requestAnimationFrame;
    Object.defineProperty(scope, "__cartularyVisualFrame", {
      value: frame.bind(scope),
    });
    Object.defineProperty(scope, "__cartularyVisualFrameNative", {
      value: /\[native code\]/u.test(Function.prototype.toString.call(frame)),
    });
  });
}

// Installs Playwright's virtual clock for the page's context. Time keeps flowing
// until a scenario holds it; install before the navigation whose timers matter.
export async function installApplicationClock(page: Page) {
  await page.clock.install();
}

// Freezes application timers, intervals, Date and performance.now. Harness frame
// waits keep using native rendered frames, so capture settling still advances.
export async function holdApplicationTime(page: Page) {
  const native = await page.evaluate(
    () => (globalThis as VisualFrameScope).__cartularyVisualFrameNative,
  );
  if (native !== true)
    throw new Error(
      "Application time can be held only after the visual frame source captured native requestAnimationFrame",
    );
  const now = await page.evaluate(() => Date.now());
  await page.clock.pauseAt(now + holdLeadMs);
}

// Advances held application time exactly; due timers fire in order.
export async function advanceApplicationTime(page: Page, milliseconds: number) {
  await page.clock.runFor(milliseconds);
}

// Steps held application time one frame at a time until a scenario signal is
// observed. Virtual time per step is fixed, so host load changes only how long
// the loop takes, not how far application time moves past the signal.
export async function advanceApplicationTimeUntil(
  page: Page,
  observed: () => boolean,
  { stepMs = 16, limitMs = 10_000 }: { stepMs?: number; limitMs?: number } = {},
) {
  for (let elapsed = 0; !observed(); elapsed += stepMs) {
    if (elapsed >= limitMs)
      throw new Error(
        `Application signal was not observed within ${limitMs} ms of held application time`,
      );
    await page.clock.runFor(stepMs);
  }
}

export async function resumeApplicationTime(page: Page) {
  await page.clock.resume();
}
