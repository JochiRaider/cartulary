import { appRouteTestId } from "@cartulary/ui-contracts";
import { NetworkFlowGridLoadFeature } from "../../workbook/features/NetworkFlowFeature";
import { AuthenticationDebugHarness } from "./AuthenticationDebugHarness";
import { IncidentDirectoryDebugHarness } from "./IncidentDirectoryDebugHarness";

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
      data-testid={appRouteTestId("debug-harness-shell")}
    >
      <AuthenticationDebugHarness />
      <IncidentDirectoryDebugHarness />
    </section>
  );
}
