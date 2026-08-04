import {
  dataTestIdSelector,
  gridScrollportClassName,
  gridShellTestId,
  rowCellTestId,
  rowInspectButtonTestId,
} from "@cartulary/ui-contracts";
import { vi } from "vitest";

export const testTimelineViewSchemaId = "cartulary.view.timeline.v2";

export function installGridContinuityFixture(
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
  const gridTestId = gridShellTestId(testTimelineViewSchemaId);
  document.body.innerHTML = `
    <div data-testid="${gridTestId}">
      <div class="${gridScrollportClassName()}">
        <button data-testid="${focusTestId}">Inspect</button>
      </div>
    </div>
  `;

  const grid = document.querySelector(dataTestIdSelector(gridTestId));
  const scrollport = grid?.querySelector(`.${gridScrollportClassName()}`);
  const focusTarget = document.querySelector(dataTestIdSelector(focusTestId));
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

export function installGridTargetFixture(
  options: {
    clientHeight?: number;
    clientWidth?: number;
    currentScroll?: { left: number; top: number };
    includeTarget?: boolean;
    isTargetVisible?: (grid: HTMLDivElement) => boolean;
    mountedRowIds?: string[];
    onGridEvaluate?: (grid: HTMLDivElement, evaluateCount: number) => void;
    onTargetScrollIntoView?: (grid: HTMLDivElement) => void;
    outerScrollable?: boolean;
    scrollHeight?: number;
    scrollWidth?: number;
    targetOutsideGrid?: boolean;
    targetTestId?: string;
  } = {},
) {
  const gridTestId = gridShellTestId(testTimelineViewSchemaId);
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
          options.includeTarget === false || options.targetOutsideGrid === true
            ? ""
            : `<input data-testid="${targetTestId}" />`
        }
      </div>
    </div>
    ${
      options.includeTarget !== false && options.targetOutsideGrid === true
        ? `<input data-testid="${targetTestId}" />`
        : ""
    }
  `;

  const shell = document.querySelector(dataTestIdSelector(gridTestId));
  const grid = shell?.querySelector(`.${gridScrollportClassName()}`);
  const target = document.querySelector(dataTestIdSelector(targetTestId));
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
      value: options.clientWidth ?? 400,
    },
    scrollHeight: {
      configurable: true,
      value: options.scrollHeight ?? 900,
    },
    scrollWidth: {
      configurable: true,
      value: options.scrollWidth ?? 400,
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
  if (target instanceof HTMLInputElement) {
    Object.defineProperty(target, "scrollIntoView", {
      configurable: true,
      value: () => {
        scrollIntoViewCalls.push(targetTestId);
        options.onTargetScrollIntoView?.(grid);
      },
    });
  }
  let gridEvaluateCount = 0;
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
        onEvaluate(testId, element) {
          if (testId === gridTestId && element === shell) {
            gridEvaluateCount += 1;
            options.onGridEvaluate?.(grid, gridEvaluateCount);
          }
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
    target: target instanceof HTMLInputElement ? target : null,
    targetTestId,
  };
}

export function installMarkerAnchorFixture(options: {
  markerCellFieldKey?: string;
  markerCellRect?: { height: number; left: number; top: number; width: number };
  markerRecordId?: string;
  markerRect: { height: number; left: number; top: number; width: number };
  markerRowRect?: { height: number; left: number; top: number; width: number };
  targetCellFieldKey?: string;
  targetCellRect?: { height: number; left: number; top: number; width: number };
  targetRecordId?: string;
  targetRowRect?: { height: number; left: number; top: number; width: number };
  targetTestId?: string;
}) {
  const gridTestId = gridShellTestId(testTimelineViewSchemaId);
  const markerTestId = "marker-record-1";
  const targetTestId =
    options.targetTestId ??
    rowCellTestId("record-1", "timeline.activity_synopsis_text");
  const targetRecordId = options.targetRecordId ?? "record-1";
  const markerRecordId = options.markerRecordId ?? targetRecordId;
  const targetCellFieldKey =
    options.targetCellFieldKey ?? "timeline.activity_synopsis_text";
  const markerCellFieldKey = options.markerCellFieldKey ?? targetCellFieldKey;
  const markerSameRow = markerRecordId === targetRecordId;
  const markerSameCell =
    markerSameRow && markerCellFieldKey === targetCellFieldKey;
  const targetRowRect = options.targetRowRect ?? {
    height: 60,
    left: 0,
    top: 40,
    width: 450,
  };
  const markerRowRect =
    options.markerRowRect ??
    (markerSameRow
      ? targetRowRect
      : { height: 60, left: 0, top: 160, width: 450 });
  const targetCellRect = options.targetCellRect ?? {
    height: 60,
    left: 100,
    top: 40,
    width: 120,
  };
  const markerCellRect =
    options.markerCellRect ??
    (markerSameCell
      ? targetCellRect
      : {
          height: markerRowRect.height,
          left: 250,
          top: markerRowRect.top,
          width: 120,
        });
  const targetInputRect = {
    height: 24,
    left: targetCellRect.left + 10,
    top: targetCellRect.top + 10,
    width: 80,
  };
  const rects = new Map<
    string,
    { height: number; left: number; top: number; width: number }
  >([
    ["scrollport", { height: 400, left: 0, top: 0, width: 500 }],
    ["target-row", targetRowRect],
    ["target-cell", targetCellRect],
    ["target-input", targetInputRect],
    ["marker-row", markerRowRect],
    ["marker-cell", markerCellRect],
    ["marker", options.markerRect],
  ]);

  const targetCellMarkup = `
    <div data-grid-field-key="${targetCellFieldKey}" data-rect-id="target-cell">
      <input data-testid="${targetTestId}" data-rect-id="target-input" />
      ${markerSameCell ? `<span data-testid="${markerTestId}" data-rect-id="marker">M</span>` : ""}
    </div>
  `;
  const markerCellMarkup = markerSameCell
    ? ""
    : `
      <div data-grid-field-key="${markerCellFieldKey}" data-rect-id="marker-cell">
        <span data-testid="${markerTestId}" data-rect-id="marker">M</span>
      </div>
    `;
  const markerRowMarkup = markerSameRow
    ? ""
    : `
      <div role="row" data-grid-record-id="${markerRecordId}" data-rect-id="marker-row">
        ${markerCellMarkup}
      </div>
    `;

  document.body.innerHTML = `
    <div data-testid="${gridTestId}">
      <div class="${gridScrollportClassName()}" data-rect-id="scrollport">
        <div role="row" data-grid-record-id="${targetRecordId}" data-rect-id="target-row">
          ${targetCellMarkup}
          ${markerSameRow ? markerCellMarkup : ""}
        </div>
        ${markerRowMarkup}
      </div>
    </div>
  `;
  const shell = document.querySelector(dataTestIdSelector(gridTestId));
  const marker = document.querySelector(dataTestIdSelector(markerTestId));
  const target = document.querySelector(dataTestIdSelector(targetTestId));
  if (
    !(shell instanceof HTMLDivElement) ||
    !(marker instanceof HTMLElement) ||
    !(target instanceof HTMLElement)
  ) {
    throw new Error("Expected marker anchor fixture elements to exist");
  }

  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockImplementation(
    function mockRect(this: HTMLElement) {
      const rectId = this.getAttribute("data-rect-id");
      const box = rectId === null ? undefined : rects.get(rectId);
      return rectFromBox(box ?? { height: 0, left: 0, top: 0, width: 0 });
    },
  );

  return {
    markerTestId,
    page: createBrowserPage({
      [gridTestId]: shell,
      [markerTestId]: marker,
      [targetTestId]: target,
    }),
    targetTestId,
  };
}

export function createBrowserPage(
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

export function rectFromBox(options: {
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
