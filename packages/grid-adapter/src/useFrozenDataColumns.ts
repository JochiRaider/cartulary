import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import { useLayoutEffect, useMemo, useRef, useSyncExternalStore } from "react";
import type { GridColumn, GridFrozenColumnStatus } from "./core";
import { elementCssScale, visibleGridViewport } from "./viewportGeometry";

const emptyPrefix: readonly string[] = [];

/** Placement is derived from measured tracks, independently of current sticky flags. */
export function frozenColumnPlacement(
  count: number,
  geometry: { viewport: number; structural: number; prefix: number } | null,
): GridFrozenColumnStatus {
  if (count === 0) return { kind: "none", visibleCount: 0 };
  if (
    !geometry ||
    !Object.values(geometry).every(Number.isFinite) ||
    geometry.viewport <= 0 ||
    geometry.structural < 0 ||
    geometry.prefix <= 0
  )
    return {
      kind: "suspended",
      visibleCount: count,
      reason: "geometry_unavailable",
    };
  return Math.floor(geometry.viewport) -
    Math.ceil(geometry.structural + geometry.prefix) >=
    cartularyDesignPresentation.workbookFrozenColumns.minimumScrollableWidthPx
    ? { kind: "active", visibleCount: count }
    : { kind: "suspended", visibleCount: count, reason: "insufficient_space" };
}

export function useFrozenDataColumns<Row>(input: {
  readonly getRoot: () => HTMLElement | null;
  readonly columns: readonly GridColumn<Row>[];
  readonly prefix: readonly string[] | undefined;
  readonly structuralCount: number;
}) {
  const prefix = input.prefix ?? emptyPrefix;
  if (
    prefix.some((key, index) => input.columns[index]?.fieldKey !== key) ||
    new Set(prefix).size !== prefix.length
  )
    throw new Error(
      "Frozen data columns must be a contiguous leading semantic prefix",
    );
  const latest = useRef(input);
  latest.current = input;
  const state = useMemo(() => {
    let snapshot: GridFrozenColumnStatus = { kind: "none", visibleCount: 0 };
    const listeners = new Set<() => void>();
    return {
      port: {
        getSnapshot: () => snapshot,
        subscribe: (listener: () => void) => {
          listeners.add(listener);
          return () => {
            listeners.delete(listener);
          };
        },
      },
      publish: (next: GridFrozenColumnStatus) => {
        if (JSON.stringify(next) === JSON.stringify(snapshot)) return;
        snapshot = next;
        for (const listener of listeners) listener();
      },
    };
  }, []);
  const snapshot = useSyncExternalStore(
    state.port.subscribe,
    state.port.getSnapshot,
  );
  const observation = useRef<{
    root: HTMLElement | null;
    dispose: () => void;
  } | null>(null);
  useLayoutEffect(() => {
    const root = input.getRoot();
    const measure = () => {
      const count = latest.current.prefix?.length ?? 0;
      let geometry: {
        viewport: number;
        structural: number;
        prefix: number;
      } | null = null;
      if (root?.isConnected && count > 0) {
        // RDG's resolved CSS grid tracks include virtualized columns. Reading them
        // avoids mirroring vendor widths or depending on which headers are mounted.
        const tracks = getComputedStyle(root)
          .gridTemplateColumns.split(/\s+/)
          .map((value) =>
            /^\d+(?:\.\d+)?px$/.test(value)
              ? Number.parseFloat(value)
              : Number.NaN,
          );
        const structuralCount = latest.current.structuralCount;
        if (
          tracks.length >= structuralCount + count &&
          tracks.every((width) => Number.isFinite(width) && width > 0)
        ) {
          const bounds = visibleGridViewport(root);
          geometry = {
            viewport: (bounds.right - bounds.left) / elementCssScale(root),
            structural: tracks
              .slice(0, structuralCount)
              .reduce((a, b) => a + b, 0),
            prefix: tracks
              .slice(structuralCount, structuralCount + count)
              .reduce((a, b) => a + b, 0),
          };
        }
      }
      const next = frozenColumnPlacement(count, geometry);
      root?.setAttribute("data-grid-freeze-state", next.kind);
      state.publish(next);
    };
    measure();
    if (observation.current?.root === root) return;
    observation.current?.dispose();
    const resize =
      typeof ResizeObserver === "undefined"
        ? null
        : new ResizeObserver(measure);
    const ancestors: HTMLElement[] = [];
    for (let node = root; node; node = node.parentElement) {
      resize?.observe(node);
      ancestors.push(node);
      node.addEventListener("scroll", measure, { passive: true });
    }
    const mutation = new MutationObserver(measure);
    for (const node of ancestors)
      mutation.observe(node, {
        attributes: true,
        attributeFilter: ["style", "class"],
      });
    window.addEventListener("resize", measure);
    window.visualViewport?.addEventListener("resize", measure);
    window.visualViewport?.addEventListener("scroll", measure);
    observation.current = {
      root,
      dispose: () => {
        resize?.disconnect();
        mutation.disconnect();
        for (const node of ancestors)
          node.removeEventListener("scroll", measure);
        window.removeEventListener("resize", measure);
        window.visualViewport?.removeEventListener("resize", measure);
        window.visualViewport?.removeEventListener("scroll", measure);
      },
    };
  });
  useLayoutEffect(
    () => () => {
      observation.current?.dispose();
      observation.current = null;
    },
    [],
  );
  return {
    port: state.port,
    activePrefix: snapshot.kind === "active" ? prefix : emptyPrefix,
  };
}
