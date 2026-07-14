import type { ComponentProps } from "react";
import { NetworkAnalysisWorkspace } from "../../networkFlow/NetworkAnalysisWorkspace";
import {
  networkAnalysisSheetRef,
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
} from "../../networkFlow/networkFlowClient";

export function NetworkFlowFeature(
  props: ComponentProps<typeof NetworkAnalysisWorkspace>,
) {
  return <NetworkAnalysisWorkspace {...props} />;
}

export {
  networkAnalysisSheetRef,
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
};
