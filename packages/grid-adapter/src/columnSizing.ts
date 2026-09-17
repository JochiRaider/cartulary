import type { GridColumnMeasurement, GridColumnSizingPort } from "./core";
import { elementCssScale, visibleGridViewport } from "./viewportGeometry";

type SizingColumn = {
  readonly fieldKey: string;
  readonly width?: number | undefined;
  readonly minWidth?: number | undefined;
  readonly maxWidth?: number | undefined;
};

export function normalizeMeasuredColumnWidth(
  width: number,
  column: SizingColumn,
  fit = false,
): number | null {
  if (!Number.isFinite(width) || (fit && width <= 0)) return null;
  return Math.max(
    column.minWidth ?? 50,
    Math.min(
      column.maxWidth ?? Number.POSITIVE_INFINITY,
      fit ? Math.ceil(width) : Math.round(width),
    ),
  );
}

function numeric(value: string): number {
  return Number.parseFloat(value) || 0;
}

export function createColumnSizingPort(
  read: () => {
    readonly root: HTMLElement | null;
    readonly columns: readonly SizingColumn[];
    readonly presentationKey: string;
  },
) {
  const listeners = new Set<() => void>();
  const pending = new Set<() => void>();
  let generation = 0;
  const invalidate = () => {
    generation += 1;
    for (const cancel of [...pending]) cancel();
    for (const listener of listeners) listener();
  };
  const headerFor = (root: HTMLElement, fieldKey: string) =>
    [
      ...root.querySelectorAll<HTMLElement>(
        '[role="columnheader"][data-grid-field-key]',
      ),
    ].find((node) => node.dataset.gridFieldKey === fieldKey);
  const unavailableReason = (fieldKey: string): string | null => {
    const { root, columns } = read();
    if (
      !root?.isConnected ||
      !columns.some((column) => column.fieldKey === fieldKey)
    )
      return "Show this column before fitting it.";
    const header = headerFor(root, fieldKey);
    if (
      !header ||
      !intersects(header.getBoundingClientRect(), visibleGridViewport(root))
    )
      return "Scroll this column into view to fit it.";
    if (document.fonts?.status === "loading")
      return "Fonts are still loading. Try Fit again when they are ready.";
    if (document.visibilityState === "hidden")
      return "Return to this window to fit the column.";
    return null;
  };
  const presentation = (fieldKey: string) => {
    const { root, presentationKey } = read();
    if (!root) return "";
    const header = headerFor(root, fieldKey);
    return JSON.stringify([
      presentationKey,
      root.scrollLeft,
      root.scrollTop,
      root.clientWidth,
      root.clientHeight,
      visibleGridViewport(root),
      [header, ...nodesFor(root, fieldKey)]
        .filter((node) => node !== undefined)
        .map((node) => {
          const style = getComputedStyle(node);
          return [
            node.innerHTML,
            style.font,
            style.lineHeight,
            style.letterSpacing,
            style.wordSpacing,
            style.padding,
            style.border,
            style.whiteSpace,
          ];
        }),
      elementCssScale(root),
    ]);
  };
  const port: GridColumnSizingPort = {
    unavailableReason,
    subscribe: (listener) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    measureVisibleContent: (fieldKey, { signal }) => {
      if (signal.aborted) return Promise.resolve({ kind: "cancelled" });
      const reason = unavailableReason(fieldKey);
      if (reason !== null)
        return Promise.resolve({ kind: "unavailable", reason });
      const requestGeneration = generation;
      const captured = presentation(fieldKey);
      return new Promise<GridColumnMeasurement>((resolve) => {
        let frame = 0;
        let deadline: ReturnType<typeof setTimeout>;
        const finish = (result: GridColumnMeasurement) => {
          cancelAnimationFrame(frame);
          clearTimeout(deadline);
          pending.delete(cancel);
          signal.removeEventListener("abort", cancel);
          resolve(result);
        };
        const cancel = () => finish({ kind: "cancelled" });
        pending.add(cancel);
        signal.addEventListener("abort", cancel, { once: true });
        deadline = setTimeout(
          () =>
            finish({
              kind: "unavailable",
              reason: "Column measurement is unavailable. Try Fit again.",
            }),
          1000,
        );
        frame = requestAnimationFrame(() => {
          if (
            signal.aborted ||
            requestGeneration !== generation ||
            captured !== presentation(fieldKey)
          ) {
            cancel();
            return;
          }
          const reasonNow = unavailableReason(fieldKey);
          if (reasonNow !== null) {
            finish({ kind: "unavailable", reason: reasonNow });
            return;
          }
          const { root, columns } = read();
          const column = columns.find((entry) => entry.fieldKey === fieldKey);
          const header = root && headerFor(root, fieldKey);
          if (!root || !column || !header) {
            cancel();
            return;
          }
          const cells = nodesFor(root, fieldKey);
          const bodyRect = visibleGridViewport(root);
          const viewport = {
            ...bodyRect,
            top: Math.max(bodyRect.top, header.getBoundingClientRect().bottom),
          };
          const visible = cells.filter((node) =>
            intersects(node.getBoundingClientRect(), viewport),
          );
          const extent = Math.max(
            ...[
              header,
              ...visible.map(
                (node) =>
                  node.closest<HTMLElement>('[role="gridcell"]') ?? node,
              ),
            ].map(measureProductionNode),
          );
          const widthPx = normalizeMeasuredColumnWidth(extent, column, true);
          if (widthPx === null) {
            finish({
              kind: "unavailable",
              reason: "Column measurement is unavailable. Try Fit again.",
            });
            return;
          }
          if (
            signal.aborted ||
            requestGeneration !== generation ||
            captured !== presentation(fieldKey)
          ) {
            cancel();
            return;
          }
          finish({
            kind: "measured",
            widthPx,
            capped: extent > widthPx,
            cellCount: visible.length,
          });
        });
      });
    },
  };
  return {
    port,
    invalidate,
    dispose: () => {
      invalidate();
      listeners.clear();
    },
  };
}

function nodesFor(root: HTMLElement, fieldKey: string): HTMLElement[] {
  return [
    ...root.querySelectorAll<HTMLElement>(
      '[data-grid-sizing-committed="true"]',
    ),
  ].filter(
    (node) =>
      node.dataset.gridFieldKey === fieldKey &&
      !node.closest('[data-cartulary-grid-draft-row="true"]') &&
      !node.querySelector(
        '[data-grid-editing="true"], [data-grid-sizing-draft="true"]',
      ),
  );
}
function intersects(
  a: { top: number; bottom: number; left: number; right: number },
  b: { top: number; bottom: number; left: number; right: number },
): boolean {
  return (
    b.right > b.left &&
    b.bottom > b.top &&
    a.right > a.left &&
    a.bottom > a.top &&
    a.right > b.left &&
    a.left < b.right &&
    a.bottom > b.top &&
    a.top < b.bottom
  );
}

/** Copies the shipped presentation, not row strings or a parallel React renderer. */
function measureProductionNode(source: HTMLElement): number {
  const clone = source.cloneNode(true) as HTMLElement;
  const originals = [source, ...source.querySelectorAll<Element>("*")];
  const copies = [clone, ...clone.querySelectorAll<Element>("*")];
  originals.forEach((original, index) => {
    const copy = copies[index];
    if (!(copy instanceof HTMLElement || copy instanceof SVGElement)) return;
    const style = getComputedStyle(original);
    for (const property of style)
      copy.style.setProperty(property, style.getPropertyValue(property));
    copy.removeAttribute("id");
    copy.removeAttribute("data-testid");
    if (
      copy instanceof HTMLElement &&
      !["BUTTON", "SVG", "IMG"].includes(copy.tagName)
    ) {
      copy.style.inlineSize = "max-content";
      copy.style.minInlineSize = "0";
      copy.style.maxInlineSize = "none";
      copy.style.flex = "0 0 auto";
      copy.style.textOverflow = "clip";
      copy.style.overflow = "visible";
    }
  });
  clone
    .querySelectorAll(
      "[data-grid-column-resize], [data-grid-sizing-exclude], input, textarea, select, [data-grid-editor-kind]",
    )
    .forEach((node) => {
      node.remove();
    });
  clone.inert = true;
  clone.setAttribute("aria-hidden", "true");
  Object.assign(clone.style, {
    position: "fixed",
    inset: "0 auto auto 0",
    visibility: "hidden",
    pointerEvents: "none",
    contain: "layout style",
    transform: "none",
    zoom: "1",
    inlineSize: "max-content",
    minInlineSize: "0",
    maxInlineSize: "none",
  });
  document.body.append(clone);
  try {
    const used = getComputedStyle(clone);
    const width = numeric(used.width);
    const border =
      numeric(used.borderLeftWidth) + numeric(used.borderRightWidth);
    const padding = numeric(used.paddingLeft) + numeric(used.paddingRight);
    return width + (used.boxSizing === "border-box" ? 0 : padding + border);
  } finally {
    clone.remove();
  }
}
