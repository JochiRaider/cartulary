/** getBoundingClientRect includes CSS zoom; computed used width is in CSS pixels. */
export function elementCssScale(element: HTMLElement): number {
  const computed = getComputedStyle(element);
  const content = Number.parseFloat(computed.width);
  const borderBox =
    computed.boxSizing === "border-box"
      ? content
      : content +
        numeric(computed.paddingLeft) +
        numeric(computed.paddingRight) +
        numeric(computed.borderLeftWidth) +
        numeric(computed.borderRightWidth);
  const scale = element.getBoundingClientRect().width / borderBox;
  return Number.isFinite(scale) && scale > 0 ? scale : 1;
}

function numeric(value: string): number {
  return Number.parseFloat(value) || 0;
}

/** Fractional client bounds in viewport coordinates, excluding borders and scrollbars. */
export function gridClientRect(element: HTMLElement) {
  const box = element.getBoundingClientRect();
  const style = getComputedStyle(element);
  const scale = elementCssScale(element);
  const left = numeric(style.borderLeftWidth);
  const right = numeric(style.borderRightWidth);
  const top = numeric(style.borderTopWidth);
  const bottom = numeric(style.borderBottomWidth);
  const gutterX = Math.max(
    0,
    element.offsetWidth - element.clientWidth - Math.round(left + right),
  );
  const gutterY = Math.max(
    0,
    element.offsetHeight - element.clientHeight - Math.round(top + bottom),
  );
  const leftGutter = style.direction === "rtl" ? gutterX : 0;
  return {
    left: box.left + (left + leftGutter) * scale,
    right: box.right - (right + gutterX - leftGutter) * scale,
    top: box.top + top * scale,
    bottom: box.bottom - (bottom + gutterY) * scale,
  };
}

/** Shared by column measurement, pointer scrolling, and mounted-editor reveal. */
export function visibleGridViewport(root: HTMLElement) {
  const bounds = gridClientRect(root);
  const view = root.ownerDocument.defaultView;
  const visual = view?.visualViewport;
  bounds.left = Math.max(bounds.left, visual?.offsetLeft ?? 0);
  bounds.top = Math.max(bounds.top, visual?.offsetTop ?? 0);
  bounds.right = Math.min(
    bounds.right,
    (visual?.offsetLeft ?? 0) + (visual?.width ?? view?.innerWidth ?? 0),
  );
  bounds.bottom = Math.min(
    bounds.bottom,
    (visual?.offsetTop ?? 0) + (visual?.height ?? view?.innerHeight ?? 0),
  );
  for (let parent = root.parentElement; parent; parent = parent.parentElement) {
    const style = getComputedStyle(parent);
    const clip = gridClientRect(parent);
    if (["auto", "scroll", "hidden", "clip"].includes(style.overflowX)) {
      bounds.left = Math.max(bounds.left, clip.left);
      bounds.right = Math.min(bounds.right, clip.right);
    }
    if (["auto", "scroll", "hidden", "clip"].includes(style.overflowY)) {
      bounds.top = Math.max(bounds.top, clip.top);
      bounds.bottom = Math.min(bounds.bottom, clip.bottom);
    }
  }
  return bounds;
}
