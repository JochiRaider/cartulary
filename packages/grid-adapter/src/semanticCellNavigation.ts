import {
  type GridCellAnchor,
  type GridCellNavigationOptions,
  type GridCellNavigationResult,
  type GridCellRange,
  type GridFocusTarget,
  type GridHandle,
  type GridPresentationSnapshot,
  gridRowIdentitiesEqual,
  gridSurfaceIdentitiesEqual,
} from "./core";
import type { ActiveEditorSession } from "./editorSessionPolicy";
import { retainGridCellRange, sameGridCellRange } from "./semanticPresentation";

type DepartureTarget =
  | GridFocusTarget
  | { readonly kind: "exit"; readonly backwards: boolean };
type DepartureOptions = GridCellNavigationOptions & {
  /** Omission means replacement; a range preserves this exact geometry. */
  readonly range?: GridCellRange | undefined;
};

/** One destination lifetime; editor sessions alone own write settlement. */
export function createSemanticCellNavigation(driver: {
  readonly read: () => {
    readonly available: boolean;
    readonly authority: string;
    readonly presentation: GridPresentationSnapshot | null;
    readonly editor: ActiveEditorSession | null;
    readonly scope?: string;
    readonly range?: GridCellRange | null;
  };
  readonly focus: GridHandle["requestFocus"];
  readonly changeRange?: (range: GridCellRange | null) => void;
  readonly exit?: (backwards: boolean) => boolean;
  readonly cancelInteraction: () => void;
}) {
  let pending: AbortController | null = null;
  let validatePending: (() => boolean) | null = null;
  const cancel = () => {
    pending?.abort();
    pending = null;
    validatePending = null;
  };
  const depart = async (
    target: DepartureTarget,
    options: DepartureOptions = {},
  ): Promise<GridCellNavigationResult> => {
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
      const range = options.range;
      return (
        !controller.signal.aborted &&
        pending === controller &&
        latest.available &&
        captured.authority === latest.authority &&
        captured.scope === latest.scope &&
        model !== null &&
        (range === undefined
          ? captured.presentation === model
          : captured.presentation !== null &&
            sameGridCellRange(latest.range ?? null, range) &&
            retainGridCellRange(captured.presentation, model, range) !==
              null) &&
        (target.kind !== "cell" ||
          (gridSurfaceIdentitiesEqual(model.surface, target.anchor.surface) &&
            model.fieldKeys.includes(target.anchor.fieldKey) &&
            model.rowIdentities.some((row) =>
              gridRowIdentitiesEqual(row, target.anchor.rowIdentity),
            ))) &&
        options.isCurrent?.() !== false
      );
    };
    validatePending = current;
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
          if (driver.read().editor === captured.editor) captured.editor.focus();
          return "rejected";
        }
      }
      if (!current()) return "cancelled";
      options.beforeFocus?.();
      if (!current()) return "cancelled";
      // Selection and physical focus are separate effects of the captured plan.
      if (target.kind === "cell")
        driver.changeRange?.(
          options.range ?? { start: target.anchor, end: target.anchor },
        );
      else if (target.kind === "draft") driver.changeRange?.(null);
      if (target.kind === "exit")
        return driver.exit?.(target.backwards) ? "focused" : "unavailable";
      return await driver.focus(target, { signal: controller.signal });
    } finally {
      options.signal?.removeEventListener("abort", abort);
      if (pending === controller) {
        pending = null;
        validatePending = null;
      }
    }
  };
  return {
    cancel,
    reconcile() {
      if (validatePending && !validatePending()) cancel();
    },
    depart,
    navigate: (
      anchor: GridCellAnchor,
      options: GridCellNavigationOptions = {},
    ) => depart({ kind: "cell", anchor }, options),
  };
}
