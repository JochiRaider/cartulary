import type { GridFocusResult, GridFocusTarget } from "./core";

function isHidden(element: HTMLElement): boolean {
  if (element.closest("[hidden], [inert], [aria-hidden='true']") !== null)
    return true;
  for (
    let ancestor: HTMLElement | null = element;
    ancestor !== null;
    ancestor = ancestor.parentElement
  ) {
    const style = getComputedStyle(ancestor);
    if (style.display === "none" || style.visibility === "hidden") return true;
  }
  return false;
}

export type GridFocusResolution =
  | { readonly kind: "unavailable" }
  | { readonly kind: "pending" }
  | { readonly kind: "target"; readonly element: HTMLElement };

/** One cancellable focus lifetime, shared by the production and support adapters. */
export function createSemanticFocusRequests(driver: {
  readonly prepare: (target: GridFocusTarget) => void;
  readonly resolve: (target: GridFocusTarget) => GridFocusResolution;
}) {
  let pending: {
    readonly target: GridFocusTarget;
    readonly finish: (result: GridFocusResult) => void;
  } | null = null;
  let refreshScheduled = false;

  const cancel = () => pending?.finish("cancelled");
  const refresh = () => {
    if (refreshScheduled || pending === null) return;
    refreshScheduled = true;
    queueMicrotask(() => {
      refreshScheduled = false;
      const request = pending;
      if (request === null) return;
      const resolution = driver.resolve(request.target);
      if (resolution.kind === "unavailable") {
        request.finish("unavailable");
        return;
      }
      if (resolution.kind === "pending") return;
      const element = resolution.element;
      if (!element.isConnected) return;
      if (
        element.matches(":disabled, [aria-disabled='true']") ||
        isHidden(element)
      ) {
        request.finish("unavailable");
        return;
      }
      if (
        !element.hasAttribute("tabindex") &&
        !element.matches("input, select, textarea, button, a[href]")
      )
        element.tabIndex = -1;
      element.focus({ preventScroll: true });
      if (document.activeElement !== element) return;
      // Ref replacement during the commit must not acknowledge a detached target.
      queueMicrotask(() => {
        if (pending !== request) return;
        const current = driver.resolve(request.target);
        if (
          current.kind === "target" &&
          current.element === element &&
          element.isConnected &&
          document.activeElement === element
        )
          request.finish("focused");
      });
    });
  };

  return {
    cancel,
    refresh,
    requestFocus(
      target: GridFocusTarget,
      options: { readonly signal?: AbortSignal } = {},
    ): Promise<GridFocusResult> {
      cancel();
      if (options.signal?.aborted) return Promise.resolve("cancelled");
      return new Promise((resolve) => {
        const signal = options.signal;
        const request = {
          target,
          finish: (result: GridFocusResult) => {
            if (pending !== request) return;
            pending = null;
            signal?.removeEventListener("abort", cancel);
            document.removeEventListener("pointerdown", cancel, true);
            document.removeEventListener("keydown", cancel, true);
            resolve(result);
          },
        };
        pending = request;
        signal?.addEventListener("abort", cancel, { once: true });
        document.addEventListener("pointerdown", cancel, true);
        document.addEventListener("keydown", cancel, true);
        driver.prepare(target);
        refresh();
      });
    },
  };
}
