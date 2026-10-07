import { type RefObject, useLayoutEffect } from "react";

/** Viewport placement only: the command owner retains drafts, focus and dismissal. */
export function useWorkbookMenuPlacement(
  open: boolean,
  ref: RefObject<HTMLElement | null>,
) {
  useLayoutEffect(() => {
    const panel = ref.current;
    if (!open || !panel) return;
    const original = {
      maxWidth: panel.style.maxWidth,
      maxHeight: panel.style.maxHeight,
      translate: panel.style.translate,
    };
    const authoredMaxHeight = panel.style.maxBlockSize || panel.style.maxHeight;
    let shiftX = 0;
    const place = () => {
      if (!panel.isConnected || panel.offsetWidth === 0) return;
      const style = getComputedStyle(panel);
      const initial = panel.getBoundingClientRect();
      const cssWidth =
        Number.parseFloat(style.width) +
        (style.boxSizing === "border-box"
          ? 0
          : [
              style.paddingLeft,
              style.paddingRight,
              style.borderLeftWidth,
              style.borderRightWidth,
            ].reduce((sum, value) => sum + (Number.parseFloat(value) || 0), 0));
      const scale = initial.width / cssWidth;
      if (!Number.isFinite(scale) || scale <= 0) return;
      const viewport = panel.ownerDocument.documentElement;
      const spacing = Number.parseFloat(style.paddingLeft) || 0;
      const availableWidth = Math.max(
        0,
        viewport.clientWidth / scale - spacing * 2,
      );
      const availableHeight = Math.max(
        0,
        (viewport.clientHeight - initial.top) / scale - spacing,
      );
      panel.style.maxWidth = `${availableWidth}px`;
      panel.style.maxHeight = authoredMaxHeight
        ? `min(${authoredMaxHeight}, ${availableHeight}px)`
        : `${availableHeight}px`;
      const bounds = panel.getBoundingClientRect();
      const naturalLeft = bounds.left - shiftX * scale;
      const left = Math.max(
        spacing * scale,
        Math.min(
          naturalLeft,
          viewport.clientWidth - bounds.width - spacing * scale,
        ),
      );
      shiftX = (left - naturalLeft) / scale;
      panel.style.translate = `${shiftX}px 0`;
    };
    place();
    const observer =
      typeof ResizeObserver === "undefined" ? null : new ResizeObserver(place);
    observer?.observe(panel);
    window.addEventListener("resize", place);
    window.visualViewport?.addEventListener("resize", place);
    return () => {
      observer?.disconnect();
      window.removeEventListener("resize", place);
      window.visualViewport?.removeEventListener("resize", place);
      Object.assign(panel.style, original);
    };
  }, [open, ref]);
}
