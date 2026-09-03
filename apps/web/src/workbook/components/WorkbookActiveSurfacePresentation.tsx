import { lazy, type ReactNode, Suspense } from "react";
import { ExtensionAvailabilityProvider } from "../../extensions/ExtensionAvailabilityContext";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import {
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
} from "../../extensions/extensionWorkspaceIdentities";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import { shellContentNoticeStyle } from "../layout/workbookShellStyles";
import {
  WorkbookSurfacesFacade,
  type WorkbookSurfacesFacadeProps,
} from "../surfaces/WorkbookSurfacesFacade";

type ExtensionWorkspaceRendererProps = {
  readonly apiBase: string | undefined;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
};

function extensionWorkspaceRegistryKey(
  extensionProfileId: string,
  workspaceKey: string,
): string {
  return `${extensionProfileId}:${workspaceKey}`;
}

const LazyNetworkFlowFeature = lazy(async () => {
  const feature = await import("../features/NetworkFlowFeature");
  return { default: feature.NetworkFlowFeature };
});

const extensionWorkspaceRenderers: Readonly<
  Record<string, (props: ExtensionWorkspaceRendererProps) => ReactNode>
> = {
  [extensionWorkspaceRegistryKey(
    networkFlowActivityProfileId,
    networkAnalysisWorkspaceKey,
  )]: (props) => (
    <Suspense fallback={null}>
      <LazyNetworkFlowFeature {...props} />
    </Suspense>
  ),
};

type WorkbookActiveSurfacePresentationProps = {
  readonly extension: {
    readonly availability: ExtensionAvailabilityController;
    readonly revision: number;
  };
  readonly extensionRenderer: ExtensionWorkspaceRendererProps;
  readonly sheetRef: SheetRef;
  readonly surface: WorkbookSurfacesFacadeProps;
};

/** Selects one active built-in or extension renderer for the current sheet. */
export function WorkbookActiveSurfacePresentation({
  extension,
  extensionRenderer,
  sheetRef,
  surface,
}: WorkbookActiveSurfacePresentationProps) {
  if (sheetRef.kind !== "extension_workspace") {
    return <WorkbookSurfacesFacade {...surface} />;
  }
  if (
    !extension.availability.isRenderable({
      extensionProfileId: sheetRef.extension_profile_id,
      workspaceKey: sheetRef.workspace_key,
    })
  ) {
    return (
      <p style={shellContentNoticeStyle}>
        This extension workspace is not currently available.
      </p>
    );
  }
  const renderer =
    extensionWorkspaceRenderers[
      extensionWorkspaceRegistryKey(
        sheetRef.extension_profile_id,
        sheetRef.workspace_key,
      )
    ];
  if (!renderer) {
    return (
      <p style={shellContentNoticeStyle}>
        This extension workspace is not available in this client.
      </p>
    );
  }
  const lifecycleKey = `${extension.availability.currentTag()?.epochId ?? "disabled"}:${extension.revision}`;
  return (
    <ExtensionAvailabilityProvider
      controller={extension.availability}
      key={lifecycleKey}
    >
      {renderer(extensionRenderer)}
    </ExtensionAvailabilityProvider>
  );
}
