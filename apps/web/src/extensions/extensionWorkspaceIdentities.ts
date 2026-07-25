import type { WorkbookSheetRef } from "../shared/workbookSheetRef";

export const networkFlowActivityProfileId = "network_flow_activity";
export const networkAnalysisWorkspaceKey = "network_analysis";
export const importProfileId = "import";
export const importRouteFamily = "/api/v1/import-sessions";

export function networkAnalysisSheetRef(): Extract<
  WorkbookSheetRef,
  { kind: "extension_workspace" }
> {
  return {
    kind: "extension_workspace",
    extension_profile_id: networkFlowActivityProfileId,
    workspace_key: networkAnalysisWorkspaceKey,
  };
}
