import type {
  GridCellAnchor,
  GridNavigationKey,
  GridSurfaceIdentity,
} from "./core";

export function isSemanticNavigationKey(
  key: string,
): key is Exclude<GridNavigationKey, "Enter" | "Tab"> {
  return (
    key === "ArrowDown" ||
    key === "ArrowLeft" ||
    key === "ArrowRight" ||
    key === "ArrowUp" ||
    key === "End" ||
    key === "Home" ||
    key === "PageDown" ||
    key === "PageUp"
  );
}

export function isPrintableGridEntry(event: {
  readonly altKey: boolean;
  readonly ctrlKey: boolean;
  readonly key: string;
  readonly metaKey: boolean;
}): boolean {
  return (
    event.key.length === 1 && !event.altKey && !event.ctrlKey && !event.metaKey
  );
}

export function isInteractiveCellActionTarget(target: EventTarget): boolean {
  return (
    target instanceof Element &&
    target.closest(
      "button, a, input, select, textarea, [role='button'], [data-grid-prevent-cell-edit='true']",
    ) !== null
  );
}

export function isGridFillHandleTarget(target: EventTarget): boolean {
  return (
    target instanceof Element &&
    target.closest(".rdg-cell-drag-handle") !== null
  );
}

export function semanticAnchorFromDomTarget(
  target: EventTarget,
  surface: GridSurfaceIdentity,
): GridCellAnchor | null {
  if (!(target instanceof Element) || surface.kind !== "view_schema") {
    return null;
  }
  const cell = target.closest<HTMLElement>('[role="gridcell"]');
  const content =
    target.closest<HTMLElement>(
      ".cartulary-grid-cell-content[data-grid-field-key]",
    ) ??
    cell?.querySelector<HTMLElement>(
      ".cartulary-grid-cell-content[data-grid-field-key]",
    );
  const row = cell?.closest<HTMLElement>(
    '[role="row"][data-grid-row-identity-kind="core_record"]',
  );
  const fieldKey = content?.getAttribute("data-grid-field-key");
  const recordId = row?.getAttribute("data-grid-record-id");
  if (
    fieldKey === null ||
    fieldKey === undefined ||
    recordId === null ||
    recordId === undefined
  ) {
    return null;
  }
  return {
    fieldKey,
    rowIdentity: { kind: "core_record", recordId },
    surface,
  };
}

export function visibleGridPageSize(
  gridElement: HTMLDivElement | null,
  rowHeight: number,
): number {
  if (gridElement === null || rowHeight <= 0) return 1;
  return Math.max(1, Math.floor(gridElement.clientHeight / rowHeight) - 2);
}

export function focusAdjacentOutsideGrid(
  gridElement: HTMLDivElement | null,
  backwards: boolean,
): boolean {
  if (gridElement === null) return false;
  const focusable = Array.from(
    document.querySelectorAll<HTMLElement>(
      "a[href], button, input, select, textarea, [tabindex]",
    ),
  ).filter(
    (element) =>
      !element.hasAttribute("disabled") &&
      element.getAttribute("aria-hidden") !== "true" &&
      element.tabIndex >= 0,
  );
  const gridIndexes = focusable.flatMap((element, index) =>
    element === gridElement || gridElement.contains(element) ? [index] : [],
  );
  if (gridIndexes.length === 0) return false;
  const targetIndex = backwards
    ? Math.min(...gridIndexes) - 1
    : Math.max(...gridIndexes) + 1;
  const target = focusable[targetIndex];
  if (target === undefined) {
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
    return !gridElement.contains(document.activeElement);
  }
  target.focus();
  return document.activeElement === target;
}
