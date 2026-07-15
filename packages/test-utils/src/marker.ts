import { gridShellTestId, type WorkbookSurface } from "@cartulary/ui-contracts";

import { type BrowserPageLike, requireEvaluate } from "./browser";
import {
  isTestIdVisibleWithinGridViewport,
  scrollGridTargetIntoView,
} from "./scrolling";

const anchorTolerancePx = 2;

type GridRectSnapshot = {
  readonly bottom: number;
  readonly height: number;
  readonly left: number;
  readonly right: number;
  readonly top: number;
  readonly width: number;
};

type MarkerAnchorState = {
  readonly markerCellFieldKey: string;
  readonly markerRect: GridRectSnapshot;
  readonly markerRowRecordId: string;
  readonly targetCellFieldKey: string;
  readonly targetCellRect: GridRectSnapshot;
  readonly targetRowRect: GridRectSnapshot;
  readonly targetRowRecordId: string;
};

export async function assertMarkerAnchoredToGridTarget(options: {
  anchorKind: "cell" | "row-gutter";
  markerTestId: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
  targetTestId: string;
}) {
  const { anchorKind, markerTestId, page, surface, targetTestId } = options;
  await scrollGridTargetIntoView({ page, surface, targetTestId });
  const targetVisible = await isTestIdVisibleWithinGridViewport(
    page,
    surface,
    targetTestId,
  );
  if (!targetVisible) {
    throw new Error(
      `Expected target ${targetTestId} to be visible in the ${surface} grid viewport`,
    );
  }
  const state = await readMarkerAnchorState({
    markerTestId,
    page,
    surface,
    targetTestId,
  });
  if (state.markerRowRecordId !== state.targetRowRecordId) {
    throw new Error(
      `Expected marker ${markerTestId} to share row record_id ${state.targetRowRecordId} with target ${targetTestId}, received ${state.markerRowRecordId}`,
    );
  }
  if (anchorKind === "cell") {
    if (state.markerCellFieldKey !== state.targetCellFieldKey) {
      throw new Error(
        `Expected marker ${markerTestId} to share cell field_key ${state.targetCellFieldKey} with target ${targetTestId}, received ${state.markerCellFieldKey}`,
      );
    }
    if (
      !containsRect(state.targetCellRect, state.markerRect, anchorTolerancePx)
    ) {
      throw new Error(
        `Expected marker ${markerTestId} to be geometrically inside target cell ${targetTestId} (marker=${formatRect(state.markerRect)} targetCell=${formatRect(state.targetCellRect)})`,
      );
    }
    return;
  }

  if (state.markerCellFieldKey !== state.targetCellFieldKey) {
    throw new Error(
      `Expected row-gutter marker ${markerTestId} to be inside target gutter field_key ${state.targetCellFieldKey}, received ${state.markerCellFieldKey}`,
    );
  }
  const markerCenterY = (state.markerRect.top + state.markerRect.bottom) / 2;
  if (
    markerCenterY < state.targetRowRect.top - anchorTolerancePx ||
    markerCenterY > state.targetRowRect.bottom + anchorTolerancePx
  ) {
    throw new Error(
      `Expected row-gutter marker ${markerTestId} to be vertically anchored to row ${state.targetRowRecordId} (marker=${formatRect(state.markerRect)} targetRow=${formatRect(state.targetRowRect)})`,
    );
  }
  if (
    !containsRect(state.targetCellRect, state.markerRect, anchorTolerancePx)
  ) {
    throw new Error(
      `Expected row-gutter marker ${markerTestId} to be geometrically inside target gutter cell ${targetTestId} (marker=${formatRect(state.markerRect)} targetCell=${formatRect(state.targetCellRect)})`,
    );
  }
}

async function readMarkerAnchorState(options: {
  markerTestId: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
  targetTestId: string;
}) {
  const { markerTestId, page, surface, targetTestId } = options;
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `assertMarkerAnchoredToGridTarget(${surface}, ${markerTestId}, ${targetTestId}) requires locator.evaluate() support`,
  );
  return (await evaluate(readMarkerAnchorStateInGrid, {
    markerTestId,
    surface,
    targetTestId,
  })) as MarkerAnchorState;
}

function readMarkerAnchorStateInGrid(
  element: Element,
  rawOptions?: unknown,
): MarkerAnchorState {
  const options =
    typeof rawOptions === "object" && rawOptions !== null
      ? (rawOptions as {
          markerTestId?: unknown;
          surface?: unknown;
          targetTestId?: unknown;
        })
      : {};
  const markerTestId =
    typeof options.markerTestId === "string" ? options.markerTestId : "";
  const targetTestId =
    typeof options.targetTestId === "string" ? options.targetTestId : "";
  const surface =
    typeof options.surface === "string" ? options.surface : "workbook";
  function findElementByTestId(
    root: Element,
    testId: string,
    role: "marker" | "target",
  ) {
    const match = Array.from(
      root.querySelectorAll<HTMLElement>("[data-testid]"),
    )
      .filter((candidate) => candidate.getAttribute("data-testid") === testId)
      .at(0);
    if (match === undefined) {
      throw new Error(`Expected ${surface} grid ${role} ${testId} to exist`);
    }
    return match;
  }
  function requireClosestElement(
    candidate: Element,
    selector: string,
    description: string,
  ) {
    const match = candidate.closest<HTMLElement>(selector);
    if (match === null) {
      throw new Error(`Expected ${description} to have closest ${selector}`);
    }
    return match;
  }
  function rectSnapshot(rect: DOMRect) {
    return {
      bottom: rect.bottom,
      height: rect.height,
      left: rect.left,
      right: rect.right,
      top: rect.top,
      width: rect.width,
    };
  }
  function effectiveRowRect(row: HTMLElement) {
    const ownRect = row.getBoundingClientRect();
    if (ownRect.width > 0 && ownRect.height > 0) {
      return rectSnapshot(ownRect);
    }
    const childRects = Array.from(row.querySelectorAll<HTMLElement>("*"))
      .map((child) => child.getBoundingClientRect())
      .filter((rect) => rect.width > 0 && rect.height > 0);
    if (childRects.length === 0) {
      return rectSnapshot(ownRect);
    }
    return {
      bottom: Math.max(...childRects.map((rect) => rect.bottom)),
      height:
        Math.max(...childRects.map((rect) => rect.bottom)) -
        Math.min(...childRects.map((rect) => rect.top)),
      left: Math.min(...childRects.map((rect) => rect.left)),
      right: Math.max(...childRects.map((rect) => rect.right)),
      top: Math.min(...childRects.map((rect) => rect.top)),
      width:
        Math.max(...childRects.map((rect) => rect.right)) -
        Math.min(...childRects.map((rect) => rect.left)),
    };
  }

  const marker = findElementByTestId(element, markerTestId, "marker");
  const target = findElementByTestId(element, targetTestId, "target");
  const markerRow = requireClosestElement(
    marker,
    '[role="row"][data-grid-record-id]',
    `${surface} marker ${markerTestId} row`,
  );
  const targetRow = requireClosestElement(
    target,
    '[role="row"][data-grid-record-id]',
    `${surface} target ${targetTestId} row`,
  );
  const markerCell = requireClosestElement(
    marker,
    "[data-grid-field-key]",
    `${surface} marker ${markerTestId} cell`,
  );
  const targetCell = requireClosestElement(
    target,
    "[data-grid-field-key]",
    `${surface} target ${targetTestId} cell`,
  );
  return {
    markerCellFieldKey:
      markerCell.getAttribute("data-grid-field-key") ?? "(missing)",
    markerRect: rectSnapshot(marker.getBoundingClientRect()),
    markerRowRecordId:
      markerRow.getAttribute("data-grid-record-id") ?? "(missing)",
    targetCellFieldKey:
      targetCell.getAttribute("data-grid-field-key") ?? "(missing)",
    targetCellRect: rectSnapshot(targetCell.getBoundingClientRect()),
    targetRowRect: effectiveRowRect(targetRow),
    targetRowRecordId:
      targetRow.getAttribute("data-grid-record-id") ?? "(missing)",
  };
}

function containsRect(
  outer: GridRectSnapshot,
  inner: GridRectSnapshot,
  tolerancePx: number,
) {
  return (
    inner.top >= outer.top - tolerancePx &&
    inner.left >= outer.left - tolerancePx &&
    inner.bottom <= outer.bottom + tolerancePx &&
    inner.right <= outer.right + tolerancePx
  );
}

function formatRect(rect: GridRectSnapshot) {
  return `top=${rect.top},right=${rect.right},bottom=${rect.bottom},left=${rect.left},width=${rect.width},height=${rect.height}`;
}
