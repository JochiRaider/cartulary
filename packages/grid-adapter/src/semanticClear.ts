import type {
  GridCellAnchor,
  GridCellRange,
  GridCellTarget,
  GridClearIntent,
} from "./core";
import { gridRowIdentitiesEqual } from "./core";
import {
  type GridSemanticPresentationModel,
  resolveSemanticCellRange,
} from "./semanticPresentation";

export function isClearNavigationKey(event: {
  readonly key: string;
  readonly altKey: boolean;
  readonly ctrlKey: boolean;
  readonly metaKey: boolean;
  readonly shiftKey: boolean;
}): boolean {
  return (
    event.key === "Delete" &&
    !event.altKey &&
    !event.ctrlKey &&
    !event.metaKey &&
    !event.shiftKey
  );
}

/** Source owners decide eligibility. Never silently omit an unavailable member. */
export function captureSemanticClear<Row>(
  model: GridSemanticPresentationModel<Row>,
  anchor: GridCellAnchor | null,
  range: GridCellRange | null,
  delivery: object = {},
): GridClearIntent | null {
  if (anchor === null) return null;
  const selected = range ?? { start: anchor, end: anchor };
  const expandedRange = resolveSemanticCellRange(model, selected);
  if (expandedRange === null) return null;
  const targets: GridCellTarget[] = [];
  for (const identity of expandedRange.rowIdentities) {
    const row = model.dataRows.find((candidate) =>
      gridRowIdentitiesEqual(candidate.rowIdentity, identity),
    );
    if (!row?.mutationIdentity) return null;
    for (const fieldKey of expandedRange.fieldKeys)
      targets.push({
        surface: model.surface,
        rowIdentity: identity,
        fieldKey,
        mutationIdentity: row.mutationIdentity,
      });
  }
  return { anchor, range: selected, expandedRange, targets, delivery };
}
