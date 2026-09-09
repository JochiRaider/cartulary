import { lazy, Suspense } from "react";
import type { NetworkAnalysisWorkspaceProps } from "../../networkFlow/NetworkAnalysisWorkspace";

export {
  type NetworkFlowImportBinding,
  NetworkFlowImportController,
  type NetworkFlowImportPort,
} from "../../networkFlow/NetworkFlowImportController";
export {
  NetworkFlowImportRecovery,
  NetworkFlowImportSurface,
} from "../../networkFlow/NetworkFlowImportSurface";

const NetworkAnalysisWorkspace = lazy(async () => {
  const module = await import("../../networkFlow/NetworkAnalysisWorkspace");
  return { default: module.NetworkAnalysisWorkspace };
});

export function NetworkFlowFeature(props: NetworkAnalysisWorkspaceProps) {
  return (
    <Suspense fallback={null}>
      <NetworkAnalysisWorkspace {...props} />
    </Suspense>
  );
}
