import {
  dataTestIdSelector,
  gridScrollportClassName,
  gridShellTestId,
} from "@cartulary/ui-contracts";

import {
  createBrowserPage,
  testTimelineViewSchemaId,
} from "./browser-fixtures";

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
      },
    ),
    scrollIntoViewCalls,
    shell,
    target: target instanceof HTMLInputElement ? target : null,
    targetTestId,
  };
}
