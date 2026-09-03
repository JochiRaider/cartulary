import type { GridCellAnchor } from "./core";
import { sameGridCellAnchor } from "./semanticPresentation";

export type SemanticActiveCellTransition =
  | { readonly kind: "no_change" }
  | { readonly anchor: GridCellAnchor | null; readonly kind: "publish" };

export function decideSemanticActiveCellTransition(
  current: GridCellAnchor | null,
  next: GridCellAnchor | null,
): SemanticActiveCellTransition {
  return sameGridCellAnchor(current, next)
    ? { kind: "no_change" }
    : { anchor: next, kind: "publish" };
}
