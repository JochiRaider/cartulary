import { useLayoutEffect, useMemo, useRef } from "react";
import { createColumnSizingPort } from "./columnSizing";
import type { GridColumn, GridDataRow, GridDensity } from "./core";

export function useGridColumnSizing<Row>(input: {
  readonly getRoot: () => HTMLElement | null;
  readonly columns: readonly GridColumn<Row>[];
  readonly rows: readonly GridDataRow<Row>[];
  readonly density: GridDensity;
  readonly dataState: string;
  readonly interactionMode: string;
  readonly surfaceKey: string;
  readonly groupingFieldKey: string | null;
}) {
  const latest = useRef(input);
  latest.current = input;
  const key = JSON.stringify([
    input.surfaceKey,
    input.groupingFieldKey,
    input.density,
    input.dataState,
    input.interactionMode,
    input.columns.map((c) => [c.fieldKey, c.width, c.minWidth, c.maxWidth]),
    input.rows.map((r) => [r.rowIdentity, r.mutationIdentity]),
  ]);
  const currentKey = useRef(key);
  currentKey.current = key;
  const sizing = useMemo(
    () =>
      createColumnSizingPort(() => ({
        root: latest.current.getRoot(),
        columns: latest.current.columns,
        presentationKey: currentKey.current,
      })),
    [],
  );
  useLayoutEffect(() => {
    currentKey.current = key;
    sizing.invalidate();
  }, [key, sizing]);
  const observation = useRef<{
    root: HTMLElement | null;
    dispose: () => void;
  } | null>(null);
  useLayoutEffect(() => {
    const root = latest.current.getRoot();
    if (observation.current?.root === root) return;
    observation.current?.dispose();
    sizing.invalidate();
    const notify = () => sizing.invalidate();
    const ancestors: HTMLElement[] = [];
    for (let node = root; node; node = node.parentElement) ancestors.push(node);
    for (const node of ancestors)
      node.addEventListener("scroll", notify, { passive: true });
    window.addEventListener("resize", notify);
    document.fonts?.addEventListener("loading", notify);
    document.fonts?.addEventListener("loadingdone", notify);
    document.addEventListener("visibilitychange", notify);
    const observer =
      typeof ResizeObserver === "undefined" ? null : new ResizeObserver(notify);
    for (const node of ancestors) observer?.observe(node);
    observation.current = {
      root,
      dispose: () => {
        for (const node of ancestors)
          node.removeEventListener("scroll", notify);
        window.removeEventListener("resize", notify);
        document.fonts?.removeEventListener("loading", notify);
        document.fonts?.removeEventListener("loadingdone", notify);
        document.removeEventListener("visibilitychange", notify);
        observer?.disconnect();
      },
    };
  });
  useLayoutEffect(
    () => () => {
      observation.current?.dispose();
      observation.current = null;
      sizing.dispose();
    },
    [sizing],
  );
  return sizing;
}
