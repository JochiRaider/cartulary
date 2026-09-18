export function isInteractiveCellActionTarget(target: EventTarget): boolean {
  return (
    target instanceof Element &&
    target.closest(
      "button, a, input, select, textarea, [role='button'], [role='checkbox'], [role='link'], [contenteditable], [data-grid-editor-interaction], [data-grid-prevent-cell-edit='true']",
    ) !== null
  );
}

/** Browser text and option interaction takes priority over grid departure. */
export function nativeEditorOwnsKey(event: {
  readonly target: EventTarget;
  readonly key: string;
  readonly shiftKey: boolean;
}): boolean {
  if (event.target instanceof HTMLTextAreaElement)
    return event.key === "Enter" && event.shiftKey;
  if (!(event.target instanceof HTMLSelectElement)) return false;
  if (
    event.key === "Enter" ||
    event.key.startsWith("Arrow") ||
    event.key === " "
  )
    return true;
  return (
    event.key === "Escape" &&
    typeof CSS !== "undefined" &&
    CSS.supports?.("selector(select:open)") === true &&
    event.target.matches(":open")
  );
}

export function isGridFillHandleTarget(target: EventTarget): boolean {
  return (
    target instanceof Element &&
    target.closest(".rdg-cell-drag-handle") !== null
  );
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
