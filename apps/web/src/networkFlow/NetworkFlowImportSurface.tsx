import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import { useSyncExternalStore } from "react";
import {
  NetworkFlowButton,
  NetworkFlowChromeStyles,
  networkFlowChromeRootClassName,
} from "./NetworkFlowControls";
import type { NetworkFlowImportController } from "./NetworkFlowImportController";
import { NetworkFlowMappingModal } from "./NetworkFlowMappingModal";
import { networkFlowImportHasWork } from "./networkFlowImportState";

/** Retained presentation stays above workspace unmount, including copy-only closure recovery. */
export function NetworkFlowImportSurface({
  controller,
}: {
  readonly controller: NetworkFlowImportController;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (state.access !== "active" || !networkFlowImportHasWork(state))
    return null;
  return (
    <div
      className={networkFlowChromeRootClassName}
      style={{ display: "contents" }}
    >
      <NetworkFlowChromeStyles />
      {state.presented ? (
        <NetworkFlowMappingModal controller={controller} state={state} />
      ) : null}
    </div>
  );
}

export function NetworkFlowImportRecovery({
  controller,
}: {
  readonly controller: NetworkFlowImportController;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (state.access !== "active" || !networkFlowImportHasWork(state))
    return null;
  return (
    <span className={networkFlowChromeRootClassName}>
      <NetworkFlowButton
        data-testid={networkAnalysisTestId("import-recovery")}
        aria-label="Review retained Network Flow import"
        onClick={() => controller.setPresented(true)}
      >
        Import recovery
      </NetworkFlowButton>
    </span>
  );
}
