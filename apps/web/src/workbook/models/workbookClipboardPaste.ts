import type {
  GridPasteTargetResolution,
  GridSurfaceIdentity,
} from "@cartulary/grid-adapter";
import type { WorkbookClipboardPasteInput } from "../adapters/WorkbookClipboardPastePort";

type PasteColumns = WorkbookClipboardPasteInput["columns"];
type PasteTargets = WorkbookClipboardPasteInput["targets"];
type PasteViewSchemaId = WorkbookClipboardPasteInput["view_schema_id"];

export function workbookPasteColumns(
  values: readonly string[],
): PasteColumns | null {
  const [first, ...rest] = values;
  if (
    first === undefined ||
    values.length > 64 ||
    values.some((value) => value.trim().length === 0)
  ) {
    return null;
  }
  return [first, ...rest];
}

export function workbookPasteTargets(
  values: readonly PasteTargets[number][],
): PasteTargets | null {
  const [first, ...rest] = values;
  if (first === undefined || values.length > 500) return null;
  for (const target of values) {
    if (target.kind === "create") continue;
    if (target.kind !== "record") return null;
    if (target.record_id.trim().length === 0 || target.base_row_version < 1) {
      return null;
    }
  }
  return [first, ...rest];
}

export function workbookPasteViewSchemaId(
  value: string,
): PasteViewSchemaId | null {
  switch (value) {
    case "cartulary.view.timeline.v2":
    case "cartulary.view.hosts.v1":
    case "cartulary.view.identities.v1":
      return value;
    default:
      return null;
  }
}

function viewSchemaSurfaceMatches(
  surface: GridSurfaceIdentity,
  viewSchemaId: string,
): boolean {
  return (
    surface.kind === "view_schema" && surface.viewSchemaId === viewSchemaId
  );
}

export function workbookPasteResolutionMatchesSurface(
  resolution: GridPasteTargetResolution,
  viewSchemaId: string,
): boolean {
  return resolution.rowTargets.every((target) =>
    viewSchemaSurfaceMatches(target.surface, viewSchemaId),
  );
}
