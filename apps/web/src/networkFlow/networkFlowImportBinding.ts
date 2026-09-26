import type { ExtensionRouteReadiness } from "../extensions/extensionAvailability";
import type { NetworkFlowImportBinding } from "./NetworkFlowImportController";

export type NetworkFlowImportSurfaceBinding = Omit<
  NetworkFlowImportBinding,
  "scope" | "current" | "available"
> & {
  readonly incidentId: string;
  readonly readiness: ExtensionRouteReadiness;
};
