import {
  type CSSProperties,
  type RefObject,
  useLayoutEffect,
  useState,
} from "react";

type MenuBounds = Pick<
  CSSProperties,
  "left" | "top" | "maxBlockSize" | "maxInlineSize"
>;

/** Fixed placement keeps the menu in DOM tab order while escaping shell clipping. */
export function useAccountMenuLayout(
  open: boolean,
  triggerRef: RefObject<HTMLButtonElement | null>,
  panelRef: RefObject<HTMLDivElement | null>,
): MenuBounds {
  const [bounds, setBounds] = useState<MenuBounds>({});
  useLayoutEffect(() => {
    if (!open) return;
    const trigger = triggerRef.current;
    const panel = panelRef.current;
    if (!trigger || !panel) return;
    const place = () => {
      const rect = panel.getBoundingClientRect();
      const anchor = trigger.getBoundingClientRect();
      // Rects include CSS zoom; layout dimensions do not. Browser zoom is
      // already reflected by the viewport. No workbook chrome policy belongs here.
      const scale =
        panel.offsetWidth > 0 && rect.width > 0
          ? rect.width / panel.offsetWidth
          : 1;
      const gap =
        (Number.parseFloat(getComputedStyle(panel).paddingTop) || 0) * scale;
      const viewport = window.visualViewport;
      const leftEdge = (viewport?.offsetLeft ?? 0) + gap;
      const topEdge = (viewport?.offsetTop ?? 0) + gap;
      const rightEdge =
        (viewport?.offsetLeft ?? 0) +
        (viewport?.width ?? window.innerWidth) -
        gap;
      const bottomEdge =
        (viewport?.offsetTop ?? 0) +
        (viewport?.height ?? window.innerHeight) -
        gap;
      const below = Math.max(
        0,
        bottomEdge - Math.max(topEdge, anchor.bottom + gap),
      );
      const above = Math.max(
        0,
        Math.min(bottomEdge, anchor.top - gap) - topEdge,
      );
      const desiredHeight = panel.scrollHeight * scale;
      const useAbove = desiredHeight > below && above > below;
      const available = useAbove ? above : below;
      const width = Math.min(rect.width, Math.max(0, rightEdge - leftEdge));
      const next: MenuBounds = {
        left:
          Math.max(
            leftEdge,
            Math.min(anchor.right - width, rightEdge - width),
          ) / scale,
        top:
          (useAbove
            ? Math.max(
                topEdge,
                anchor.top - gap - Math.min(desiredHeight, available),
              )
            : Math.max(topEdge, Math.min(anchor.bottom + gap, bottomEdge))) /
          scale,
        maxBlockSize: available / scale,
        maxInlineSize: Math.max(0, rightEdge - leftEdge) / scale,
      };
      setBounds((current) =>
        current.left === next.left &&
        current.top === next.top &&
        current.maxBlockSize === next.maxBlockSize &&
        current.maxInlineSize === next.maxInlineSize
          ? current
          : next,
      );
    };
    place();
    const observer = new ResizeObserver(place);
    observer.observe(trigger);
    observer.observe(panel);
    const rootStyle = new MutationObserver(place);
    rootStyle.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["style"],
    });
    window.addEventListener("resize", place);
    window.addEventListener("scroll", place, true);
    window.visualViewport?.addEventListener("resize", place);
    window.visualViewport?.addEventListener("scroll", place);
    return () => {
      observer.disconnect();
      rootStyle.disconnect();
      window.removeEventListener("resize", place);
      window.removeEventListener("scroll", place, true);
      window.visualViewport?.removeEventListener("resize", place);
      window.visualViewport?.removeEventListener("scroll", place);
    };
  }, [open, panelRef, triggerRef]);

  useLayoutEffect(() => {
    const panel = panelRef.current;
    const active = document.activeElement;
    if (
      !open ||
      !panel ||
      !(active instanceof HTMLElement) ||
      !panel.contains(active)
    )
      return;
    const viewport = panel.getBoundingClientRect();
    const item = active.getBoundingClientRect();
    const scale =
      panel.offsetWidth > 0 && viewport.width > 0
        ? viewport.width / panel.offsetWidth
        : 1;
    const padding =
      (Number.parseFloat(getComputedStyle(panel).paddingTop) || 0) * scale;
    if (item.top < viewport.top + padding)
      panel.scrollTop -= (viewport.top + padding - item.top) / scale;
    else if (item.bottom > viewport.bottom - padding)
      panel.scrollTop += (item.bottom - viewport.bottom + padding) / scale;
  });
  return bounds;
}
