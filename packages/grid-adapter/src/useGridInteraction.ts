import { useLayoutEffect, useMemo, useRef, useState } from "react";
import type { GridCellAnchor, GridCellRange } from "./core";
import {
  createGridInteractionController,
  type GridInteractionSnapshot,
} from "./gridInteractionController";
import {
  bindGridInteractionDom,
  type RegisteredGridCell,
} from "./gridInteractionDom";
import { sameGridCellRange } from "./semanticPresentation";

export function useGridInteraction(input: {
  readonly read: () => GridInteractionSnapshot;
  readonly root: () => HTMLElement | null;
  readonly cells: () => ReadonlyMap<string, RegisteredGridCell>;
  readonly accept: (
    range: GridCellRange,
    edit: boolean,
    inspect: boolean,
  ) => void;
  readonly changeRange: (range: GridCellRange | null) => void;
  readonly announce: (message: string) => void;
}) {
  const latest = useRef(input);
  latest.current = input;
  const binding = useRef<ReturnType<typeof bindGridInteractionDom> | null>(
    null,
  );
  const mountedRoot = useRef<HTMLElement | null>(null);
  const [preview, setPreview] = useState<GridCellRange | null>(null);
  const controller = useMemo(
    () =>
      createGridInteractionController({
        read: () => latest.current.read(),
        accept: (...args) => latest.current.accept(...args),
        changeRange: (range) => latest.current.changeRange(range),
        announce: (message) => latest.current.announce(message),
        preview: (range) =>
          setPreview((current) =>
            sameGridCellRange(current, range) ? current : range,
          ),
        stopped: () => binding.current?.stop(),
      }),
    [],
  );
  useLayoutEffect(() => {
    const root = input.root();
    if (root !== mountedRoot.current) {
      binding.current?.dispose();
      mountedRoot.current = root;
      binding.current = root
        ? bindGridInteractionDom(root, controller, () => latest.current.cells())
        : null;
    }
    controller.reconcile();
  });
  useLayoutEffect(
    () => () => {
      binding.current?.dispose();
      binding.current = null;
      mountedRoot.current = null;
    },
    [],
  );
  const click = (anchor: GridCellAnchor, shiftKey: boolean) => {
    // Accessibility activation and DOM-unit click events need no pointer coordinates.
    if (controller.begin(anchor, { pointerId: -1, x: 0, y: 0, shiftKey }))
      controller.release({ x: 0, y: 0 }, anchor);
  };
  return { controller, binding, click, preview };
}
