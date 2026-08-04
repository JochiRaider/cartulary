import {
  dataTestIdSelector,
  gridScrollportClassName,
  gridShellTestId,
  rowInspectButtonTestId,
} from "@cartulary/ui-contracts";
import { vi } from "vitest";

import {
  createBrowserPage,
  rectFromBox,
  testTimelineViewSchemaId,
} from "./browser-fixtures";

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
