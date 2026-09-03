import {
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
} from "../../extensions/extensionWorkspaceIdentities";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookAccountModel } from "../../shared/workbookShellContracts";
import { requiredBuiltInWorkbookSurfaceIds } from "./workbookSurfaceRegistry";

export function isNetworkAnalysisSheetRef(sheetRef: SheetRef): boolean {
  return (
    sheetRef.kind === "extension_workspace" &&
    sheetRef.extension_profile_id === networkFlowActivityProfileId &&
    sheetRef.workspace_key === networkAnalysisWorkspaceKey
  );
}

export function workbookAccountPresentation(
  account: WorkbookAccountModel | undefined,
  currentUserLabel: string | undefined,
): { readonly displayName: string; readonly title: string } {
  const displayName =
    account?.display_name ?? currentUserLabel ?? "Unknown user";
  return {
    displayName,
    title: account
      ? `${account.display_name}${account.is_deployment_admin ? " (deployment administrator)" : ""}`
      : displayName,
  };
}

export function workbookActiveSystemSurfaceTitle(
  surface: string,
  activeContractTitle: string,
  networkAnalysisActive: boolean,
): string | null {
  return requiredBuiltInWorkbookSurfaceIds.includes(surface) ||
    networkAnalysisActive
    ? null
    : activeContractTitle;
}
