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

/** Adapter-owned sticky regions in viewport coordinates; semantic order is unchanged. */
export function gridViewportRegions(root: HTMLElement) {
  const viewport = visibleGridViewport(root);
  const rtl = getComputedStyle(root).direction === "rtl";
  let structuralEdge = rtl ? viewport.right : viewport.left;
  let frozenEdge = structuralEdge;
  for (const header of root.querySelectorAll<HTMLElement>(
    '[role="columnheader"]',
  )) {
    const style = getComputedStyle(header);
    if (
      style.position !== "sticky" ||
      !style.insetInlineStart ||
      style.insetInlineStart === "auto"
    )
      continue;
    const rect = header.getBoundingClientRect();
    const edge = rtl ? rect.left : rect.right;
    frozenEdge = rtl ? Math.min(frozenEdge, edge) : Math.max(frozenEdge, edge);
    if (!header.classList.contains("cartulary-grid-frozen-data"))
      structuralEdge = rtl
        ? Math.min(structuralEdge, edge)
        : Math.max(structuralEdge, edge);
  }
  const data = { ...viewport };
  const scrollable = { ...viewport };
  const frozen = { ...viewport };
  if (rtl) {
    data.right = Math.min(data.right, structuralEdge);
    scrollable.right = Math.min(scrollable.right, frozenEdge);
    frozen.right = data.right;
    frozen.left = Math.max(frozen.left, frozenEdge);
  } else {
    data.left = Math.max(data.left, structuralEdge);
    scrollable.left = Math.max(scrollable.left, frozenEdge);
    frozen.left = data.left;
    frozen.right = Math.min(frozen.right, frozenEdge);
  }
  return { viewport, data, scrollable, frozen };
}

export function isFrozenDataCell(target: HTMLElement) {
  return target.closest(".cartulary-grid-frozen-data") !== null;
}

export function gridColumnViewport(
  root: HTMLElement,
  target?: HTMLElement,
  correction = false,
) {
  const regions = gridViewportRegions(root);
  return target && isFrozenDataCell(target)
    ? correction
      ? regions.data
      : regions.frozen
    : regions.scrollable;
}

export function gridContentViewport(
  root: HTMLElement,
  target?: HTMLElement,
  correction = false,
) {
  const bounds = gridColumnViewport(root, target, correction);
  for (const header of root.querySelectorAll<HTMLElement>(
    '[role="columnheader"]',
  ))
    bounds.top = Math.max(
      bounds.top,
      Math.min(bounds.bottom, header.getBoundingClientRect().bottom),
    );
  // A trailing creation row does not obstruct its own editor.
  if (!target?.closest('[data-cartulary-grid-draft-row="true"]')) {
    for (const draft of root.querySelectorAll<HTMLElement>(
      '[data-cartulary-grid-draft-row="true"] [role="gridcell"]',
    ))
      if (getComputedStyle(draft).position === "sticky")
        bounds.bottom = Math.min(
          bounds.bottom,
          draft.getBoundingClientRect().top,
        );
  }
  return bounds;
}

/** Minimal translation; an oversized target already spanning its area stays put. */
export function gridRevealDelta(
  start: number,
  end: number,
  low: number,
  high: number,
) {
  if (
    high <= low ||
    (start >= low && end <= high) ||
    (start <= low && end >= high)
  )
    return 0;
  if (end - start > high - low) return start > low ? start - low : end - high;
  return start < low ? start - low : end - high;
}

/** Reveal a mounted semantic cell without moving focus or scrolling the document. */
export function revealGridCell(root: HTMLElement, target: HTMLElement) {
  const bounds = gridContentViewport(root, target);
  const rect = target.getBoundingClientRect();
  const scale = elementCssScale(root);
  const dx = isFrozenDataCell(target)
    ? 0
    : gridRevealDelta(rect.left, rect.right, bounds.left, bounds.right) / scale;
  const dy =
    gridRevealDelta(rect.top, rect.bottom, bounds.top, bounds.bottom) / scale;
  if (dx || dy)
    root.scrollTo({
      left: root.scrollLeft + dx,
      top: root.scrollTop + dy,
      behavior: "instant",
    });
}
