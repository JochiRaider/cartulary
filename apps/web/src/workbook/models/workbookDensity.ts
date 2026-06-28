import type { GridDensity } from "@cartulary/grid-adapter";
import type { WorkbookDensityMode } from "../../shared/workbookShellContracts";
import { timelineViewSchemaId } from "./workbookSurfaceRegistry";

export type AccountDensityMode = WorkbookDensityMode | null;

export function resolveEffectiveWorkbookDensity(
  viewSchemaId: string,
  accountDensityMode: AccountDensityMode | undefined,
): GridDensity {
  if (accountDensityMode !== null && accountDensityMode !== undefined) {
    return accountDensityMode;
  }
  return viewSchemaId === timelineViewSchemaId ? "compact" : "default";
}
