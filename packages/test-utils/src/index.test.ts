import { afterEach, describe, expect, it, vi } from "vitest";

import {
  assertGridFocusContinuity,
  gridDraftRowSelector,
  gridSavedRowsSelector,
  gridShellTestId,
  rowInspectButtonTestId,
} from "./index";

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  document.body.innerHTML = "";
});

describe("@cartulary/test-utils workbook row selectors", () => {
  it("targets saved and draft workbook rows when scoped through the grid shell", () => {
    document.body.innerHTML = `
      <div data-testid="${gridShellTestId("timeline")}">
        <div role="row">header row</div>
        <div role="row" data-grid-record-id="record-1">saved row</div>
        <div role="row" data-grid-record-id="">draft row</div>
      </div>
      <div data-testid="${gridShellTestId("hosts")}">
        <div role="row" data-grid-record-id="host-1">host row</div>
      </div>
    `;

    const timelineShell = document.querySelector(
      `[data-testid="${gridShellTestId("timeline")}"]`,
    );
    const hostsShell = document.querySelector(
      `[data-testid="${gridShellTestId("hosts")}"]`,
    );

    expect(timelineShell?.querySelectorAll(gridSavedRowsSelector())).toHaveLength(
      1,
    );
    expect(timelineShell?.querySelectorAll(gridDraftRowSelector())).toHaveLength(
      1,
    );
    expect(hostsShell?.querySelectorAll(gridSavedRowsSelector())).toHaveLength(
      1,
    );
    expect(hostsShell?.querySelectorAll(gridDraftRowSelector())).toHaveLength(0);
  });
});

describe("@cartulary/test-utils grid continuity", () => {
  it("allows visibility-preserving continuity when scroll changes by default", async () => {
    const { focusTestId, page } = installGridContinuityFixture({
      currentScroll: { left: 10, top: 180 },
      focusRect: { height: 40, left: 85, top: 170, width: 80 },
      gridRect: { height: 300, left: 40, top: 100, width: 400 },
    });

    await expect(
      assertGridFocusContinuity({
        focusTestId,
        intervalMs: 0,
        page,
        preservedScroll: { left: 18, top: 240 },
        surface: "timeline",
        timeoutMs: 10,
      }),
    ).resolves.toBeUndefined();
  });

  it("fails when exact vertical scroll is required and scrollTop changed under fake timers", async () => {
    const { focusTestId, page } = installGridContinuityFixture({
      currentScroll: { left: 18, top: 180 },
      focusRect: { height: 40, left: 85, top: 170, width: 80 },
      gridRect: { height: 300, left: 40, top: 100, width: 400 },
    });
    vi.useFakeTimers();

    await expect(
      assertGridFocusContinuity({
        focusTestId,
        intervalMs: 0,
        page,
        preservedScroll: { left: 18, top: 240 },
        requireExactVerticalScroll: true,
        surface: "timeline",
        timeoutMs: 10,
      }),
    ).rejects.toThrow("Expected timeline vertical scroll 240, received 180");
  });

  it("retries until the preserved vertical scroll converges", async () => {
    let gridEvaluateCount = 0;
    const { focusTestId, page } = installGridContinuityFixture(
      {
        currentScroll: { left: 18, top: 180 },
        focusRect: { height: 40, left: 85, top: 170, width: 80 },
        gridRect: { height: 300, left: 40, top: 100, width: 400 },
      },
      {
        onEvaluate(testId, element) {
          if (testId !== gridShellTestId("timeline")) {
            return;
          }
          gridEvaluateCount += 1;
          if (gridEvaluateCount >= 5 && element instanceof HTMLDivElement) {
            element.scrollTop = 240;
          }
        },
      },
    );

    await expect(
      assertGridFocusContinuity({
        focusTestId,
        intervalMs: 0,
        page,
        preservedScroll: { left: 18, top: 240 },
        requireExactVerticalScroll: true,
        surface: "timeline",
        timeoutMs: 10,
      }),
    ).resolves.toBeUndefined();

    expect(gridEvaluateCount).toBeGreaterThan(2);
  });

  it("fails when exact horizontal scroll is required and scrollLeft changed", async () => {
    const { focusTestId, page } = installGridContinuityFixture({
      currentScroll: { left: 10, top: 240 },
      focusRect: { height: 40, left: 85, top: 170, width: 80 },
      gridRect: { height: 300, left: 40, top: 100, width: 400 },
    });

    await expect(
      assertGridFocusContinuity({
        focusTestId,
        intervalMs: 0,
        page,
        preservedScroll: { left: 18, top: 240 },
        requireExactHorizontalScroll: true,
        surface: "timeline",
        timeoutMs: 10,
      }),
    ).rejects.toThrow("Expected timeline horizontal scroll 18, received 10");
  });
});

function installGridContinuityFixture(
  options: {
    currentScroll: { left: number; top: number };
    focusRect: { height: number; left: number; top: number; width: number };
    gridRect: { height: number; left: number; top: number; width: number };
  },
  pageOptions: {
    onEvaluate?: (testId: string, element: Element) => void;
  } = {},
) {
  const focusTestId = rowInspectButtonTestId("record-1");
  const gridTestId = gridShellTestId("timeline");
  document.body.innerHTML = `
    <div data-testid="${gridTestId}">
      <button data-testid="${focusTestId}">Inspect</button>
    </div>
  `;

  const grid = document.querySelector(`[data-testid="${gridTestId}"]`);
  const focusTarget = document.querySelector(`[data-testid="${focusTestId}"]`);
  if (!(grid instanceof HTMLDivElement) || !(focusTarget instanceof HTMLButtonElement)) {
    throw new Error("Expected grid continuity fixture elements to exist");
  }

  grid.scrollTop = options.currentScroll.top;
  grid.scrollLeft = options.currentScroll.left;
  focusTarget.focus();

  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockImplementation(
    function mockRect(this: HTMLElement) {
      const testId = this.getAttribute("data-testid");
      if (testId === gridTestId) {
        return rectFromBox(options.gridRect);
      }
      if (testId === focusTestId) {
        return rectFromBox(options.focusRect);
      }
      return rectFromBox({ height: 0, left: 0, top: 0, width: 0 });
    },
  );

  return {
    focusTestId,
    page: createBrowserPage({
      [focusTestId]: focusTarget,
      [gridTestId]: grid,
    }, pageOptions),
  };
}

function createBrowserPage(
  elements: Record<string, Element>,
  options: {
    onEvaluate?: (testId: string, element: Element) => void;
  } = {},
) {
  return {
    getByTestId(value: string) {
      const element = elements[value];
      if (element === undefined) {
        throw new Error(`Unknown test id ${value}`);
      }
      return {
        click: async () => {
          if (element instanceof HTMLElement) {
            element.click();
          }
        },
        evaluate: async (pageFunction: (element: Element) => unknown) => {
          options.onEvaluate?.(value, element);
          return pageFunction(element);
        },
        fill: async () => undefined,
      };
    },
  };
}

function rectFromBox(options: {
  height: number;
  left: number;
  top: number;
  width: number;
}) {
  return {
    bottom: options.top + options.height,
    height: options.height,
    left: options.left,
    right: options.left + options.width,
    top: options.top,
    width: options.width,
    x: options.left,
    y: options.top,
    toJSON: () => ({}),
  } as DOMRect;
}
