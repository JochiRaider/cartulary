import { type RefObject, useLayoutEffect } from "react";

/** Viewport placement only: the command owner retains drafts, focus and dismissal. */
export function useWorkbookMenuPlacement(
  open: boolean,
  ref: RefObject<HTMLElement | null>,
  anchorRef: RefObject<HTMLElement | null>,
) {
  useLayoutEffect(() => {
    const panel = ref.current;
    const anchor = anchorRef.current;
    if (!open || !panel || !anchor) return;
    const placementProperties = [
      "position",
      "top",
      "right",
      "bottom",
      "left",
      "margin-top",
      "margin-right",
      "margin-bottom",
      "margin-left",
      "max-width",
      "max-height",
    ];
    const originalPlacement = placementProperties.map((property) => ({
      property,
      value: panel.style.getPropertyValue(property),
      priority: panel.style.getPropertyPriority(property),
    }));
    const authoredMaxHeight = panel.style.maxBlockSize || panel.style.maxHeight;
    // A nested command panel must escape its parent menu's scroll clipping.
    // DOM containment remains intact for the owner's outside/focus handling.
    panel.showPopover?.();
    panel.style.position = "fixed";
    panel.style.inset = "auto";
    panel.style.margin = "0";
    const place = () => {
      if (!panel.isConnected || !anchor.isConnected || panel.offsetWidth === 0)
        return;
      const style = getComputedStyle(panel);
      const bounds = panel.getBoundingClientRect();
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
      const scale = bounds.width / cssWidth;
      if (!Number.isFinite(scale) || scale <= 0) return;
      const viewport = window.visualViewport;
      const left = viewport?.offsetLeft ?? 0;
      const top = viewport?.offsetTop ?? 0;
      const width = viewport?.width ?? window.innerWidth;
      const height = viewport?.height ?? window.innerHeight;
      const spacing = Number.parseFloat(style.paddingLeft) || 0;
      panel.style.maxWidth = `${Math.max(0, width / scale - spacing * 2)}px`;
      const maxHeight = `${Math.max(0, height / scale - spacing * 2)}px`;
      panel.style.maxHeight = authoredMaxHeight
        ? `min(${authoredMaxHeight}, ${maxHeight})`
        : maxHeight;
      const placed = panel.getBoundingClientRect();
      const target = anchor.getBoundingClientRect();
      panel.style.left = `${Math.max(left + spacing * scale, Math.min(target.left, left + width - placed.width - spacing * scale)) / scale}px`;
      panel.style.top = `${Math.max(top + spacing * scale, Math.min(target.bottom, top + height - placed.height - spacing * scale)) / scale}px`;
    };
    place();
    const observer =
      typeof ResizeObserver === "undefined" ? null : new ResizeObserver(place);
    observer?.observe(panel);
    observer?.observe(anchor);
    window.addEventListener("resize", place);
    window.addEventListener("scroll", place, true);
    window.visualViewport?.addEventListener("resize", place);
    return () => {
      observer?.disconnect();
      window.removeEventListener("resize", place);
      window.removeEventListener("scroll", place, true);
      window.visualViewport?.removeEventListener("resize", place);
      if (panel.popover === "manual") panel.hidePopover?.();
      for (const { property, value, priority } of originalPlacement) {
        if (value) panel.style.setProperty(property, value, priority);
        else panel.style.removeProperty(property);
      }
    };
  }, [open, ref, anchorRef]);
}
