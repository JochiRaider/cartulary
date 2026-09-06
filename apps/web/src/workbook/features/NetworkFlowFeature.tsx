import type { ComponentProps } from "react";
import { NetworkAnalysisWorkspace } from "../../networkFlow/NetworkAnalysisWorkspace";

export function NetworkFlowFeature(
  props: ComponentProps<typeof NetworkAnalysisWorkspace>,
) {
  return <NetworkAnalysisWorkspace {...props} />;
}
