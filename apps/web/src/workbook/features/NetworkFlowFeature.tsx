import type { ComponentProps } from "react";
import { NetworkAnalysisWorkspace } from "../../networkFlow/NetworkAnalysisWorkspace";
import { NetworkFlowGridLoadFixture } from "../../networkFlow/NetworkFlowGridLoadFixture";

export function NetworkFlowFeature(
  props: ComponentProps<typeof NetworkAnalysisWorkspace>,
) {
  return <NetworkAnalysisWorkspace {...props} />;
}

export function NetworkFlowGridLoadFeature() {
  return <NetworkFlowGridLoadFixture />;
}
