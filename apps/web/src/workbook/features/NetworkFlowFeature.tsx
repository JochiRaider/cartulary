import {
  NetworkAnalysisWorkspace,
  type NetworkAnalysisWorkspaceProps,
} from "../../networkFlow/NetworkAnalysisWorkspace";

/** Loaded only when the authorized Network Analysis surface is selected. */
export function NetworkFlowFeature(props: NetworkAnalysisWorkspaceProps) {
  return <NetworkAnalysisWorkspace {...props} />;
}
