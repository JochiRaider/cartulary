import {
  type GridCellAnchor,
  type GridCellNavigationOptions,
  type GridCellNavigationResult,
  type GridHandle,
  type GridPresentationSnapshot,
  gridRowIdentitiesEqual,
  gridSurfaceIdentitiesEqual,
} from "./core";
import type { ActiveEditorSession } from "./editorSessionPolicy";

export function createSemanticCellNavigation(driver: {
  readonly read: () => {
    readonly available: boolean;
    readonly authority: string;
    readonly presentation: GridPresentationSnapshot | null;
    readonly editor: ActiveEditorSession | null;
  };
  readonly focus: GridHandle["requestFocus"];
  readonly cancelInteraction: () => void;
}) {
  let pending: AbortController | null = null;
  const cancel = () => {
    pending?.abort();
    pending = null;
  };
  return {
    cancel,
    async navigate(
      anchor: GridCellAnchor,
      options: GridCellNavigationOptions = {},
    ): Promise<GridCellNavigationResult> {
      cancel();
      if (options.signal?.aborted) return "cancelled";
      const controller = new AbortController();
      pending = controller;
      const abort = () => controller.abort();
      options.signal?.addEventListener("abort", abort, { once: true });
      const captured = { ...driver.read() };
      const current = () => {
        const latest = driver.read();
        const model = latest.presentation;
        return (
          !controller.signal.aborted &&
          pending === controller &&
          latest.available &&
          captured.authority === latest.authority &&
          model !== null &&
          captured.presentation === model &&
          gridSurfaceIdentitiesEqual(model.surface, anchor.surface) &&
          model.fieldKeys.includes(anchor.fieldKey) &&
          model.rowIdentities.some((row) =>
            gridRowIdentitiesEqual(row, anchor.rowIdentity),
          ) &&
          options.isCurrent?.() !== false
        );
      };
      try {
        if (!current()) return "unavailable";
        driver.cancelInteraction();
        if (captured.editor) {
          captured.editor.focus();
          let stop: () => void = () => {};
          const interrupted = new Promise<false>((resolve) => {
            stop = () => resolve(false);
            controller.signal.addEventListener("abort", stop, { once: true });
          });
          const accepted = await Promise.race([
            captured.editor.requestCommit(),
            interrupted,
          ]);
          controller.signal.removeEventListener("abort", stop);
          if (controller.signal.aborted) return "cancelled";
          if (!current()) return "unavailable";
          if (!accepted) {
            if (driver.read().editor === captured.editor)
              captured.editor.focus();
            return "rejected";
          }
        }
        if (!current()) return "cancelled";
        options.beforeFocus?.();
        return await driver.focus(
          { kind: "cell", anchor },
          { signal: controller.signal },
        );
      } finally {
        options.signal?.removeEventListener("abort", abort);
        if (pending === controller) pending = null;
      }
    },
  };
}
