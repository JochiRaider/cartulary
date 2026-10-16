import type { SheetRef } from "../shared/sheetRef";

export const networkFlowActivityProfileId = "network_flow_activity";
export const networkFlowRouteFamily =
  "/api/v1/incidents/{incident_id}/network-flow";
export const networkAnalysisWorkspaceKey = "network_analysis";
export const reportingProfileId = "snapshot_reporting";
export const reportingSnapshotsRoute = "/api/v1/snapshots";
export const reportingReleasesRoute = "/api/v1/releases";
export const importProfileId = "import";
export const importRouteFamily = "/api/v1/import-sessions";

export function networkAnalysisSheetRef(): Extract<
  SheetRef,
  { kind: "extension_workspace" }
> {
  return {
    kind: "extension_workspace",
    extension_profile_id: networkFlowActivityProfileId,
    workspace_key: networkAnalysisWorkspaceKey,
  };
}
