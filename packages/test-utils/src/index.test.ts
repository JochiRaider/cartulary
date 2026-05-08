// @vitest-environment jsdom

import {
  gridScrollportClassName,
  gridShellTestId,
  rowInspectButtonTestId,
} from "@cartulary/ui-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";

import { assertGridFocusContinuity, scrollGridTargetIntoView } from "./index";

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  document.body.innerHTML = "";
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
          const scrollport = element.querySelector(
            `.${gridScrollportClassName()}`,
          );
          if (gridEvaluateCount >= 5 && scrollport instanceof HTMLDivElement) {
            scrollport.scrollTop = 240;
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
        timeoutMs: 250,
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

describe("@cartulary/test-utils virtualized grid targeting", () => {
  it("returns the existing scroll position when the target is already visible", async () => {
    const { grid, page, scrollIntoViewCalls, targetTestId } =
      installGridTargetFixture({
        currentScroll: { left: 8, top: 120 },
        isTargetVisible: () => true,
      });

    await expect(
      scrollGridTargetIntoView({
        intervalMs: 0,
        page,
        surface: "timeline",
        targetTestId,
        timeoutMs: 50,
      }),
    ).resolves.toEqual({ left: 8, top: 120 });

    expect(grid.scrollTop).toBe(120);
    expect(scrollIntoViewCalls).toEqual([]);
  });

  it("scans grid offsets until a virtualized target is mounted", async () => {
    const { grid, page, scrollIntoViewCalls, targetTestId } =
      installGridTargetFixture({
        clientHeight: 200,
        currentScroll: { left: 0, top: 0 },
        isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 400,
        onTargetScrollIntoView: (candidateGrid) => {
          candidateGrid.scrollTop = 520;
        },
        scrollHeight: 900,
      });

    await expect(
      scrollGridTargetIntoView({
        intervalMs: 0,
        page,
        surface: "timeline",
        targetTestId,
        timeoutMs: 1_000,
      }),
    ).resolves.toEqual({ left: 0, top: 520 });

    expect(grid.scrollTop).toBe(520);
    expect(scrollIntoViewCalls).toEqual([targetTestId]);
  });

  it("scrolls the explicit grid scrollport when the outer shell is also scrollable", async () => {
    const { grid, page, shell, targetTestId } = installGridTargetFixture({
      clientHeight: 200,
      currentScroll: { left: 0, top: 0 },
      isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 400,
      outerScrollable: true,
      scrollHeight: 900,
    });

    await expect(
      scrollGridTargetIntoView({
        intervalMs: 0,
        page,
        surface: "timeline",
        targetTestId,
        timeoutMs: 1_000,
      }),
    ).resolves.toEqual({ left: 0, top: 400 });

    expect(grid.scrollTop).toBe(400);
    expect(shell.scrollTop).toBe(250);
  });

  it("fails explicitly when the grid shell has no owned scrollport", async () => {
    const gridTestId = gridShellTestId("timeline");
    document.body.innerHTML = `<div data-testid="${gridTestId}"></div>`;
    const shell = document.querySelector(`[data-testid="${gridTestId}"]`);
    if (!(shell instanceof HTMLDivElement)) {
      throw new Error("Expected missing-scrollport fixture shell to exist");
    }

    const page = createBrowserPage({ [gridTestId]: shell });

    await expect(
      scrollGridTargetIntoView({
        intervalMs: 0,
        page,
        surface: "timeline",
        targetTestId: "missing-target",
        timeoutMs: 50,
      }),
    ).rejects.toThrow(
      "Expected timeline grid shell to contain exactly one .cartulary-grid-scrollport scrollport, received 0",
    );
  });

  it("throws diagnostics when the target never becomes visible", async () => {
    const { page } = installGridTargetFixture({
      includeTarget: false,
      mountedRowIds: ["record-a", "record-b"],
    });

    await expect(
      scrollGridTargetIntoView({
        intervalMs: 0,
        page,
        surface: "timeline",
        targetTestId: "missing-target",
        timeoutMs: 50,
      }),
    ).rejects.toThrow(
      /missing-target.*timeline.*scrollHeight=900.*mountedRowIds=record-a,record-b/,
    );
  });

  it("returns the final scroll snapshot after target alignment", async () => {
    const { page, targetTestId } = installGridTargetFixture({
      clientHeight: 200,
      currentScroll: { left: 3, top: 0 },
      isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 200,
      onTargetScrollIntoView: (candidateGrid) => {
        candidateGrid.scrollLeft = 21;
        candidateGrid.scrollTop = 333;
      },
      scrollHeight: 900,
    });

    await expect(
      scrollGridTargetIntoView({
        intervalMs: 0,
        page,
        surface: "timeline",
        targetTestId,
        timeoutMs: 1_000,
      }),
    ).resolves.toEqual({ left: 21, top: 333 });
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
      <div class="${gridScrollportClassName()}">
        <button data-testid="${focusTestId}">Inspect</button>
      </div>
    </div>
  `;

  const grid = document.querySelector(`[data-testid="${gridTestId}"]`);
  const scrollport = grid?.querySelector(`.${gridScrollportClassName()}`);
  const focusTarget = document.querySelector(`[data-testid="${focusTestId}"]`);
  if (
    !(grid instanceof HTMLDivElement) ||
    !(scrollport instanceof HTMLDivElement) ||
    !(focusTarget instanceof HTMLButtonElement)
  ) {
    throw new Error("Expected grid continuity fixture elements to exist");
  }

  scrollport.scrollTop = options.currentScroll.top;
  scrollport.scrollLeft = options.currentScroll.left;
  focusTarget.focus();

  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockImplementation(
    function mockRect(this: HTMLElement) {
      const testId = this.getAttribute("data-testid");
      if (
        testId === null &&
        this.classList.contains(gridScrollportClassName())
      ) {
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
    page: createBrowserPage(
      {
        [focusTestId]: focusTarget,
        [gridTestId]: grid,
      },
      pageOptions,
    ),
  };
}

function installGridTargetFixture(
  options: {
    clientHeight?: number;
    currentScroll?: { left: number; top: number };
    includeTarget?: boolean;
    isTargetVisible?: (grid: HTMLDivElement) => boolean;
    mountedRowIds?: string[];
    onTargetScrollIntoView?: (grid: HTMLDivElement) => void;
    outerScrollable?: boolean;
    scrollHeight?: number;
    targetTestId?: string;
  } = {},
) {
  const gridTestId = gridShellTestId("timeline");
  const targetTestId = options.targetTestId ?? "target-control";
  const mountedRows = (options.mountedRowIds ?? ["record-1"])
    .map(
      (recordId) => `<div role="row" data-grid-record-id="${recordId}"></div>`,
    )
    .join("");
  document.body.innerHTML = `
    <div data-testid="${gridTestId}">
      <div class="${gridScrollportClassName()}">
        ${mountedRows}
        ${
          options.includeTarget === false
            ? ""
            : `<input data-testid="${targetTestId}" />`
        }
      </div>
    </div>
  `;

  const shell = document.querySelector(`[data-testid="${gridTestId}"]`);
  const grid = shell?.querySelector(`.${gridScrollportClassName()}`);
  const target = document.querySelector(`[data-testid="${targetTestId}"]`);
  if (!(shell instanceof HTMLDivElement) || !(grid instanceof HTMLDivElement)) {
    throw new Error("Expected grid targeting fixture grid to exist");
  }
  if (
    options.includeTarget !== false &&
    !(target instanceof HTMLInputElement)
  ) {
    throw new Error("Expected grid targeting fixture target to exist");
  }

  Object.defineProperties(grid, {
    clientHeight: {
      configurable: true,
      value: options.clientHeight ?? 200,
    },
    clientWidth: {
      configurable: true,
      value: 400,
    },
    scrollHeight: {
      configurable: true,
      value: options.scrollHeight ?? 900,
    },
    scrollWidth: {
      configurable: true,
      value: 400,
    },
  });
  if (options.outerScrollable) {
    Object.defineProperties(shell, {
      clientHeight: {
        configurable: true,
        value: 100,
      },
      clientWidth: {
        configurable: true,
        value: 300,
      },
      scrollHeight: {
        configurable: true,
        value: 3_000,
      },
      scrollWidth: {
        configurable: true,
        value: 300,
      },
    });
    shell.scrollTop = 250;
  }
  grid.scrollLeft = options.currentScroll?.left ?? 0;
  grid.scrollTop = options.currentScroll?.top ?? 0;

  const scrollIntoViewCalls: string[] = [];
  return {
    grid,
    page: createBrowserPage(
      () => {
        const elements: Record<string, Element | undefined> = {
          [gridTestId]: shell,
        };
        if (target instanceof HTMLInputElement) {
          elements[targetTestId] = target;
        }
        return elements;
      },
      {
        isVisible(testId, element) {
          if (testId === targetTestId && element === target) {
            return options.isTargetVisible?.(grid) ?? true;
          }
          return element.isConnected;
        },
        onScrollIntoViewIfNeeded(testId) {
          scrollIntoViewCalls.push(testId);
          if (testId === targetTestId) {
            options.onTargetScrollIntoView?.(grid);
          }
        },
      },
    ),
    scrollIntoViewCalls,
    shell,
    targetTestId,
  };
}

function createBrowserPage(
  elements:
    | Record<string, Element | undefined>
    | (() => Record<string, Element | undefined>),
  options: {
    isVisible?: (
      testId: string,
      element: Element,
    ) => boolean | Promise<boolean>;
    onEvaluate?: (testId: string, element: Element) => void;
    onScrollIntoViewIfNeeded?: (testId: string, element: Element) => void;
  } = {},
) {
  const resolveElement = (value: string) => {
    const resolvedElements =
      typeof elements === "function" ? elements() : elements;
    return resolvedElements[value];
  };
  return {
    getByTestId(value: string) {
      return {
        click: async () => {
          const element = resolveElement(value);
          if (element === undefined) {
            throw new Error(`Unknown test id ${value}`);
          }
          if (element instanceof HTMLElement) {
            element.click();
          }
        },
        evaluate: async (
          pageFunction: (element: Element, arg?: unknown) => unknown,
          arg?: unknown,
        ) => {
          const element = resolveElement(value);
          if (element === undefined) {
            throw new Error(`Unknown test id ${value}`);
          }
          options.onEvaluate?.(value, element);
          return pageFunction(element, arg);
        },
        fill: async () => undefined,
        isVisible: async () => {
          const element = resolveElement(value);
          if (element === undefined) {
            return false;
          }
          return options.isVisible?.(value, element) ?? element.isConnected;
        },
        scrollIntoViewIfNeeded: async () => {
          const element = resolveElement(value);
          if (element === undefined) {
            throw new Error(`Unknown test id ${value}`);
          }
          options.onScrollIntoViewIfNeeded?.(value, element);
        },
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
