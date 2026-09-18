import {
  gridRevealDelta as editorRevealDelta,
  elementCssScale,
  gridContentViewport,
  isFrozenDataCell,
} from "./viewportGeometry";

export { gridRevealDelta as editorRevealDelta } from "./viewportGeometry";

const editorViewport = (root: HTMLElement, target: HTMLElement) =>
  gridContentViewport(root, target, true);

function positionToolbar(
  root: HTMLElement,
  editor: HTMLElement,
  target: HTMLElement,
) {
  const toolbar = editor.querySelector<HTMLElement>(
    '[data-grid-editor-toolbar="true"]',
  );
  if (!toolbar) return;
  const bounds = editorViewport(root, target);
  const scale = elementCssScale(editor);
  const width = Math.max(0, bounds.right - bounds.left) / scale;
  const height = Math.max(0, bounds.bottom - bounds.top) / scale;
  toolbar.style.maxInlineSize = `${width}px`;
  toolbar.style.maxBlockSize = `${height}px`;
  const cell = editor.getBoundingClientRect();
  const box = toolbar.getBoundingClientRect();
  const left = Math.max(
    bounds.left,
    Math.min(cell.left, bounds.right - box.width),
  );
  const above = cell.top - box.height;
  const preferredTop =
    cell.bottom + box.height > bounds.bottom && above >= bounds.top
      ? above
      : cell.bottom;
  const top = Math.max(
    bounds.top,
    Math.min(preferredTop, bounds.bottom - box.height),
  );
  toolbar.style.setProperty(
    "--cartulary-editor-toolbar-left",
    `${(left - cell.left) / scale}px`,
  );
  toolbar.style.setProperty(
    "--cartulary-editor-toolbar-top",
    `${(top - cell.top) / scale}px`,
  );
}

/** Scroll only the owned grid; never focus, resize a column, or change a draft. */
export function revealGridEditorTarget(
  root: HTMLElement,
  editor: HTMLElement,
  target: HTMLElement,
) {
  if (!root.isConnected || !editor.contains(target) || !root.contains(editor))
    return;
  let bounds = editorViewport(root, target);
  if (bounds.right <= bounds.left || bounds.bottom <= bounds.top) return;
  const scale = elementCssScale(root);
  const toolbar = target.closest('[data-grid-editor-toolbar="true"]');
  // Correction placement must not scroll its original cell out of view.
  if (!toolbar) {
    const rect = target.getBoundingClientRect();
    const style = getComputedStyle(target);
    const ring =
      Math.max(
        0,
        (Number.parseFloat(style.outlineWidth) || 0) +
          (Number.parseFloat(style.outlineOffset) || 0),
      ) * elementCssScale(target);
    const horizontalRing =
      rect.width + ring * 2 <= bounds.right - bounds.left ? ring : 0;
    const verticalRing =
      rect.height + ring * 2 <= bounds.bottom - bounds.top ? ring : 0;
    const dx = isFrozenDataCell(target)
      ? 0
      : editorRevealDelta(
          rect.left - horizontalRing,
          rect.right + horizontalRing,
          bounds.left,
          bounds.right,
        ) / scale;
    const dy =
      editorRevealDelta(
        rect.top - verticalRing,
        rect.bottom + verticalRing,
        bounds.top,
        bounds.bottom,
      ) / scale;
    const rtl = getComputedStyle(root).direction === "rtl";
    const maxLeft = Math.max(0, root.scrollWidth - root.clientWidth);
    const left = Math.max(
      rtl ? -maxLeft : 0,
      Math.min(rtl ? 0 : maxLeft, root.scrollLeft + dx),
    );
    const top = Math.max(
      0,
      Math.min(
        Math.max(0, root.scrollHeight - root.clientHeight),
        root.scrollTop + dy,
      ),
    );
    if (left !== root.scrollLeft || top !== root.scrollTop)
      root.scrollTo({ left, top, behavior: "instant" });
  }
  positionToolbar(root, editor, target);
  if (toolbar instanceof HTMLElement) {
    bounds = editorViewport(root, target);
    const box = target.getBoundingClientRect();
    const clip = toolbar.getBoundingClientRect();
    const toolbarScale = elementCssScale(toolbar);
    toolbar.scrollTop +=
      editorRevealDelta(
        box.top,
        box.bottom,
        Math.max(bounds.top, clip.top + toolbar.clientTop * toolbarScale),
        Math.min(bounds.bottom, clip.bottom - toolbar.clientTop * toolbarScale),
      ) / toolbarScale;
  }
}

/** One editor lifetime. Resize and feedback never reclaim focus from newer work. */
export function bindGridEditorReveal(
  root: HTMLElement,
  editor: HTMLElement,
  isCurrent: () => boolean,
) {
  let disposed = false;
  let frame: number | null = null;
  const refresh = () => {
    if (disposed || frame !== null) return;
    frame = requestAnimationFrame(() => {
      frame = null;
      const active = editor.ownerDocument.activeElement;
      if (
        disposed ||
        !isCurrent() ||
        !(active instanceof HTMLElement) ||
        !editor.contains(active)
      )
        return;
      revealGridEditorTarget(root, editor, active);
    });
  };
  const resize =
    typeof ResizeObserver === "undefined" ? null : new ResizeObserver(refresh);
  resize?.observe(root);
  resize?.observe(editor);
  for (let parent = root.parentElement; parent; parent = parent.parentElement) {
    const style = getComputedStyle(parent);
    if (
      [style.overflowX, style.overflowY].some((value) =>
        ["auto", "scroll", "hidden", "clip"].includes(value),
      )
    )
      resize?.observe(parent);
  }
  for (const control of editor.querySelectorAll<HTMLElement>(
    "input, textarea, select, [data-grid-editor-toolbar]",
  ))
    resize?.observe(control);
  const view = root.ownerDocument.defaultView;
  editor.addEventListener("focusin", refresh);
  view?.addEventListener("resize", refresh);
  view?.visualViewport?.addEventListener("resize", refresh);
  refresh();
  return {
    refresh,
    dispose: () => {
      disposed = true;
      if (frame !== null) cancelAnimationFrame(frame);
      resize?.disconnect();
      editor.removeEventListener("focusin", refresh);
      view?.removeEventListener("resize", refresh);
      view?.visualViewport?.removeEventListener("resize", refresh);
    },
  };
}
