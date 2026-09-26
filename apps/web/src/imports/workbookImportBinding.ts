import type { ExtensionRouteReadiness } from "../extensions/extensionAvailability";
import type { WorkbookImportBinding } from "./WorkbookImportController";

export type WorkbookImportSurfaceBinding = Omit<
  WorkbookImportBinding,
  "scope" | "current" | "available"
> & {
  readonly incidentId: string;
  readonly readiness: ExtensionRouteReadiness;
};
