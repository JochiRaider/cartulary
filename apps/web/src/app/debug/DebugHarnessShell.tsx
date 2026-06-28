import { phase1RouteTestId } from "@cartulary/ui-contracts";
import { Phase1Harness } from "./Phase1Harness";
import { Phase2Harness } from "./Phase2Harness";

export function DebugHarnessShell() {
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
