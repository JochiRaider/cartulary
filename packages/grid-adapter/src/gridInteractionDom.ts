import type { GridCellAnchor } from "./core";
import {
  isGridFillHandleTarget,
  isInteractiveCellActionTarget,
} from "./domInteraction";
import {
  type createGridInteractionController,
  gridEdgeScrollDelta,
} from "./gridInteractionController";
import { semanticPresentationContainsAnchor } from "./semanticPresentation";
import { elementCssScale, visibleGridViewport } from "./viewportGeometry";

export type RegisteredGridCell = {
  readonly anchor: GridCellAnchor;
  readonly cell: HTMLElement;
  readonly token: object;
};

type Controller = ReturnType<typeof createGridInteractionController>;

/** A mounted grid owns capture and animation; no document scrolling or query capability. */
export function bindGridInteractionDom(
  root: HTMLElement,
  controller: Controller,
  readCells: () => ReadonlyMap<string, RegisteredGridCell>,
) {
  let capturedId: number | null = null;
  let handledPointer = false;
  let compatibilityClick = false;
  let frame: number | null = null;
  let lastFrame = 0;
  let position = { x: 0, y: 0 };
  let disposed = false;

  const fromTarget = (target: EventTarget | null): GridCellAnchor | null => {
    if (!(target instanceof Element)) return null;
    const node = target.closest('[role="gridcell"]');
    if (!node || !root.contains(node)) return null;
    for (const entry of readCells().values())
      if (entry.cell === node || node.contains(entry.cell)) return entry.anchor;
    return null;
  };
  const nativeOwner = (target: EventTarget | null) =>
    target instanceof Element &&
    (isInteractiveCellActionTarget(target) ||
      isGridFillHandleTarget(target) ||
      target.closest(
        '[data-grid-editor-interaction], [data-grid-editor-external-action="true"], [contenteditable], [role="columnheader"]',
      ) !== null);
  const cssPoint = () => {
    const scale = elementCssScale(root);
    return { x: position.x / scale, y: position.y / scale };
  };
  const endpoint = (clamp: boolean): GridCellAnchor | null => {
    if (!clamp)
      return fromTarget(
        document.elementFromPoint?.(position.x, position.y) ?? null,
      );
    const bounds = visibleGridViewport(root);
    const bodyTop = Math.max(
      bounds.top,
      root
        .querySelector<HTMLElement>('[role="columnheader"]')
        ?.getBoundingClientRect().bottom ?? bounds.top,
    );
    let closest: { anchor: GridCellAnchor; distance: number } | null = null;
    const model = controller.capturedModel;
    if (model === null) return null;
    // Only mounted data cells in the current clipped viewport participate.
    for (const entry of readCells().values()) {
      if (
        !entry.cell.isConnected ||
        !semanticPresentationContainsAnchor(model, entry.anchor)
      )
        continue;
      const rect = entry.cell.getBoundingClientRect();
      const left = Math.max(bounds.left, rect.left);
      const right = Math.min(bounds.right, rect.right);
      const top = Math.max(bodyTop, rect.top);
      const bottom = Math.min(bounds.bottom, rect.bottom);
      if (right <= left || bottom <= top) continue;
      const dx = Math.max(left - position.x, 0, position.x - right);
      const dy = Math.max(top - position.y, 0, position.y - bottom);
      const distance = dx * dx + dy * dy;
      if (closest === null || distance < closest.distance)
        closest = { anchor: entry.anchor, distance };
    }
    return closest?.anchor ?? null;
  };
  const stop = () => {
    document.removeEventListener("pointermove", move, true);
    document.removeEventListener("pointerup", up, true);
    document.removeEventListener("pointercancel", lost, true);
    root.removeEventListener("lostpointercapture", lost);
    if (!controller.active) {
      document.removeEventListener("keydown", key, true);
      window.removeEventListener("blur", interrupt);
      document.removeEventListener("visibilitychange", visibility);
    }
    if (frame !== null) cancelAnimationFrame(frame);
    frame = null;
    lastFrame = 0;
    const id = capturedId;
    capturedId = null;
    if (id !== null && root.hasPointerCapture?.(id))
      root.releasePointerCapture(id);
    root.removeAttribute("data-grid-pointer-selecting");
  };
  const tick = (now: number) => {
    frame = null;
    if (disposed || !root.isConnected || controller.pointerId === null) {
      controller.cancel();
      stop();
      return;
    }
    controller.reconcile();
    if (!controller.active) {
      stop();
      return;
    }
    const scale = elementCssScale(root);
    if (controller.scrolling) {
      const bounds = visibleGridViewport(root);
      const header = root.querySelector<HTMLElement>('[role="columnheader"]');
      const top = Math.max(
        bounds.top,
        header?.getBoundingClientRect().bottom ?? bounds.top,
      );
      const elapsed = lastFrame === 0 ? 0 : now - lastFrame;
      const left = Math.max(
        0,
        Math.min(
          root.scrollWidth - root.clientWidth,
          root.scrollLeft +
            gridEdgeScrollDelta(
              position.x / scale,
              bounds.left / scale,
              bounds.right / scale,
              elapsed,
            ),
        ),
      );
      const scrollTop = Math.max(
        0,
        Math.min(
          root.scrollHeight - root.clientHeight,
          root.scrollTop +
            gridEdgeScrollDelta(
              position.y / scale,
              top / scale,
              bounds.bottom / scale,
              elapsed,
            ),
        ),
      );
      if (left !== root.scrollLeft || scrollTop !== root.scrollTop)
        root.scrollTo({ left, top: scrollTop, behavior: "instant" });
    }
    controller.move(cssPoint(), endpoint(true));
    lastFrame = now;
    if (capturedId !== null) frame = requestAnimationFrame(tick);
  };
  const invalidModifiers = (event: MouseEvent) =>
    event.altKey || event.ctrlKey || event.metaKey;
  const down = (event: PointerEvent) => {
    compatibilityClick = false;
    handledPointer = false;
    controller.cancel();
    stop();
    if (
      !root.contains(event.target as Node) ||
      event.button !== 0 ||
      event.isPrimary === false ||
      (event.pointerType && event.pointerType !== "mouse") ||
      invalidModifiers(event) ||
      nativeOwner(event.target)
    )
      return;
    const anchor = fromTarget(event.target);
    if (anchor === null) return;
    position = { x: event.clientX, y: event.clientY };
    if (
      !controller.begin(anchor, {
        ...cssPoint(),
        pointerId: event.pointerId,
        shiftKey: event.shiftKey,
      })
    )
      return;
    document.addEventListener("pointermove", move, true);
    document.addEventListener("pointerup", up, true);
    document.addEventListener("pointercancel", lost, true);
    root.addEventListener("lostpointercapture", lost);
    document.addEventListener("keydown", key, true);
    window.addEventListener("blur", interrupt);
    document.addEventListener("visibilitychange", visibility);
    handledPointer = true;
    compatibilityClick = true;
    capturedId = event.pointerId;
    root.setPointerCapture?.(event.pointerId);
    root.setAttribute("data-grid-pointer-selecting", "true");
    event.preventDefault();
    event.stopPropagation();
    frame = requestAnimationFrame(tick);
  };
  const move = (event: PointerEvent) => {
    if (event.pointerId !== capturedId) return;
    if (invalidModifiers(event) || event.buttons !== 1) {
      controller.cancel(true);
      stop();
      return;
    }
    position = { x: event.clientX, y: event.clientY };
    // Classification is synchronous; geometry and selection work are frame-bounded.
    controller.track(cssPoint());
    event.preventDefault();
  };
  const up = (event: PointerEvent) => {
    if (event.pointerId !== capturedId) return;
    position = { x: event.clientX, y: event.clientY };
    controller.release(cssPoint(), endpoint(controller.scrolling));
    stop();
    event.preventDefault();
    event.stopPropagation();
  };
  const lost = (event: PointerEvent) => {
    if (event.pointerId !== capturedId) return;
    controller.cancel(true);
    stop();
  };
  const interrupt = () => {
    controller.cancel(true);
    stop();
  };
  const visibility = () => {
    if (document.visibilityState === "hidden") interrupt();
  };
  const key = (event: KeyboardEvent) => {
    if (event.key === "Shift") return;
    if (!controller.active) return;
    const escaped = event.key === "Escape";
    controller.cancel(escaped);
    stop();
    if (escaped) {
      event.preventDefault();
      event.stopImmediatePropagation();
    }
  };
  const click = (event: MouseEvent) => {
    if (compatibilityClick && handledPointer && event.detail !== 0) {
      compatibilityClick = false;
      handledPointer = false;
      event.preventDefault();
      event.stopImmediatePropagation();
    }
  };
  const dragstart = (event: DragEvent) => {
    if (handledPointer) event.preventDefault();
  };
  const enterNativeEditor = (event: FocusEvent) => {
    if (
      event.target instanceof Element &&
      event.target.matches(
        "input, textarea, select, [contenteditable='true']",
      ) &&
      event.target.closest('[data-cartulary-grid-draft-row="true"]')
    )
      controller.enterNativeEditor();
  };
  const observer = new MutationObserver(() => {
    controller.reconcile();
    if (!root.isConnected || !controller.active) stop();
  });
  observer.observe(root, { childList: true, subtree: true });
  document.addEventListener("pointerdown", down, true);

  document.addEventListener("click", click, true);
  root.addEventListener("dragstart", dragstart);
  root.addEventListener("focusin", enterNativeEditor);

  return {
    stop,
    fromTarget,
    get handledPointer() {
      return handledPointer;
    },
    dispose() {
      disposed = true;
      controller.dispose();
      stop();
      observer.disconnect();
      document.removeEventListener("pointerdown", down, true);
      document.removeEventListener("pointermove", move, true);
      document.removeEventListener("pointerup", up, true);
      document.removeEventListener("pointercancel", lost, true);
      root.removeEventListener("lostpointercapture", lost);
      document.removeEventListener("keydown", key, true);
      document.removeEventListener("click", click, true);
      root.removeEventListener("dragstart", dragstart);
      root.removeEventListener("focusin", enterNativeEditor);
      window.removeEventListener("blur", interrupt);
      document.removeEventListener("visibilitychange", visibility);
    },
  };
}
