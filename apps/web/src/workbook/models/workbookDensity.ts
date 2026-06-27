import type { GridDensity } from "@cartulary/grid-adapter";
import type { DensityMode } from "@cartulary/protocol-ts";
import { timelineViewSchemaId } from "./workbookSurfaceRegistry";

export type AccountDensityMode = DensityMode | null;

export function resolveEffectiveWorkbookDensity(
  viewSchemaId: string,
  accountDensityMode: AccountDensityMode | undefined,
): GridDensity {
  if (accountDensityMode !== null && accountDensityMode !== undefined) {
    return accountDensityMode;
  }
  return viewSchemaId === timelineViewSchemaId ? "compact" : "default";
}
