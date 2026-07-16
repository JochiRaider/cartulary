import { phase1RouteTestId } from "@cartulary/ui-contracts";
import { NetworkFlowGridLoadFeature } from "../../workbook/features/NetworkFlowFeature";
import { Phase1Harness } from "./Phase1Harness";
import { Phase2Harness } from "./Phase2Harness";

export function DebugHarnessShell() {
  if (
    new URLSearchParams(window.location.search).get("fixture") ===
    "network-flow-grid-load"
  ) {
    return <NetworkFlowGridLoadFeature />;
  }
  return (
    <section
      aria-label="Debug harness controls"
      data-testid={phase1RouteTestId("debug-harness-shell")}
    >
      <Phase1Harness />
      <Phase2Harness />
    </section>
  );
}
