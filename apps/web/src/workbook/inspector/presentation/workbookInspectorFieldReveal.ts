type Bounds = { left: number; right: number; top: number; bottom: number };

const clippingOverflow = new Set(["auto", "scroll", "hidden", "clip"]);

function cssScale(element: HTMLElement): number {
  const rect = element.getBoundingClientRect();
  const scale = element.offsetWidth > 0 ? rect.width / element.offsetWidth : 1;
  return Number.isFinite(scale) && scale > 0 ? scale : 1;
}

function clientBounds(element: HTMLElement): Bounds {
  const rect = element.getBoundingClientRect();
  const scale = cssScale(element);
  return {
    left: rect.left + element.clientLeft * scale,
    top: rect.top + element.clientTop * scale,
    right: rect.left + (element.clientLeft + element.clientWidth) * scale,
    bottom: rect.top + (element.clientTop + element.clientHeight) * scale,
  };
}

/** The persistent inspector header is outside this body-owned clip rectangle. */
function usableBodyBounds(body: HTMLElement): Bounds {
  const bounds = clientBounds(body);
  const view = body.ownerDocument.defaultView;
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
  for (
    let ancestor = body.parentElement;
    ancestor;
    ancestor = ancestor.parentElement
  ) {
    const style = getComputedStyle(ancestor);
    const clip = clientBounds(ancestor);
    if (clippingOverflow.has(style.overflowX)) {
      bounds.left = Math.max(bounds.left, clip.left);
      bounds.right = Math.min(bounds.right, clip.right);
    }
    if (clippingOverflow.has(style.overflowY)) {
      bounds.top = Math.max(bounds.top, clip.top);
      bounds.bottom = Math.min(bounds.bottom, clip.bottom);
    }
  }
  return bounds;
}

function revealDelta(start: number, end: number, low: number, high: number) {
  if (
    high <= low ||
    (start >= low && end <= high) ||
    (start <= low && end >= high)
  )
    return 0;
  if (end - start > high - low) return start > low ? start - low : end - high;
  return start < low ? start - low : end - high;
}

function clamp(value: number, low: number, high: number) {
  return Math.max(low, Math.min(high, value));
}

/** Scrolls only the inspector body; authoring, caret and focus stay with the caller. */
export function revealWorkbookInspectorField(
  body: HTMLElement,
  field: HTMLElement,
  target: HTMLElement,
) {
  if (
    !body.isConnected ||
    !field.isConnected ||
    !field.contains(target) ||
    !body.contains(field)
  )
    return;
  revealTarget(body, target, field);
}

/** Reveals only the control and its focus ring, without moving focus or authoring. */
export function revealWorkbookInspectorTarget(
  body: HTMLElement,
  target: HTMLElement,
) {
  revealTarget(body, target);
}

function revealTarget(
  body: HTMLElement,
  target: HTMLElement,
  field?: HTMLElement,
) {
  if (!body.isConnected || !target.isConnected || !body.contains(target))
    return;
  const bounds = usableBodyBounds(body);
  if (bounds.right <= bounds.left || bounds.bottom <= bounds.top) return;
  const rect = target.getBoundingClientRect();
  const style = getComputedStyle(target);
  const ring =
    Math.max(
      0,
      (Number.parseFloat(style.outlineWidth) || 0) +
        (Number.parseFloat(style.outlineOffset) || 0),
    ) * cssScale(target);
  const targetBounds: Bounds = {
    left: rect.left - ring,
    right: rect.right + ring,
    top: rect.top - ring,
    bottom: rect.bottom + ring,
  };
  const label = field
    ?.querySelector<HTMLElement>(":scope > dt")
    ?.getBoundingClientRect();
  const value = field
    ?.querySelector<HTMLElement>(":scope > [data-inspector-field-value]")
    ?.getBoundingClientRect();
  const contextTop = Math.min(
    targetBounds.top,
    label?.top ?? targetBounds.top,
    value?.top ?? targetBounds.top,
  );
  const contextBottom = Math.max(
    targetBounds.bottom,
    label?.bottom ?? targetBounds.bottom,
    value?.bottom ?? targetBounds.bottom,
  );
  const editorLabel = field
    ?.querySelector("[data-inspector-editor-field] > legend")
    ?.getBoundingClientRect();
  const identificationTop = Math.min(
    editorLabel?.top ?? label?.top ?? targetBounds.top,
    targetBounds.top,
  );
  const identificationBottom = Math.max(
    editorLabel?.bottom ?? label?.bottom ?? targetBounds.bottom,
    targetBounds.bottom,
  );
  const visibleHeight = bounds.bottom - bounds.top;
  const preferredTop =
    contextBottom - contextTop <= visibleHeight
      ? revealDelta(contextTop, contextBottom, bounds.top, bounds.bottom)
      : identificationBottom - identificationTop <= visibleHeight
        ? revealDelta(
            identificationTop,
            identificationBottom,
            bounds.top,
            bounds.bottom,
          )
        : revealDelta(
            targetBounds.top,
            targetBounds.bottom,
            bounds.top,
            bounds.bottom,
          );
  const scale = cssScale(body);
  const maxTop = Math.max(0, body.scrollHeight - body.clientHeight);
  const maxLeft = Math.max(0, body.scrollWidth - body.clientWidth);
  const rtl = getComputedStyle(body).direction === "rtl";
  const top = clamp(body.scrollTop + preferredTop / scale, 0, maxTop);
  const left = clamp(
    body.scrollLeft +
      revealDelta(
        targetBounds.left,
        targetBounds.right,
        bounds.left,
        bounds.right,
      ) /
        scale,
    rtl ? -maxLeft : 0,
    rtl ? 0 : maxLeft,
  );
  if (top !== body.scrollTop || left !== body.scrollLeft)
    body.scrollTo({ top, left, behavior: "instant" });
  // At a scroll boundary the complete context may not fit; the control wins.
  const adjusted = usableBodyBounds(body);
  const actual = target.getBoundingClientRect();
  const correction = revealDelta(
    actual.top - ring,
    actual.bottom + ring,
    adjusted.top,
    adjusted.bottom,
  );
  if (correction !== 0)
    body.scrollTop = clamp(body.scrollTop + correction / scale, 0, maxTop);
}
