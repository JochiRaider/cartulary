import {
  dataTestIdSelector,
  gridScrollportClassName,
  gridShellTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import { vi } from "vitest";

import {
  createBrowserPage,
  rectFromBox,
  testTimelineViewSchemaId,
} from "./browser-fixtures";

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
