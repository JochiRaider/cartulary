import { useSyncExternalStore } from "react";
import { networkAnalysisSheetRef } from "../extensions/extensionWorkspaceIdentities";
import {
  useWorkbookRecoveryPresentation,
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../shared/workbookRecoveryNavigation";
import {
  NetworkFlowChromeStyles,
  networkFlowChromeRootClassName,
} from "./NetworkFlowControls";
import type { NetworkFlowImportController } from "./NetworkFlowImportController";
import { NetworkFlowMappingPanel } from "./NetworkFlowMappingPanel";
import { networkFlowImportHasWork } from "./networkFlowImportState";

export function NetworkFlowImportSurface({
  controller,
}: {
  readonly controller: NetworkFlowImportController;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const visible = state.access === "active" && networkFlowImportHasWork(state);
  const completed = state.stage === "finished";
  const attention = !!(
    state.sourceFailure ||
    state.previewFailure ||
    state.write?.disposition === "uncertain" ||
    state.write?.disposition === "rejected" ||
    state.handoff?.status === "failed"
  );
  const items: readonly WorkbookRecoveryItem[] = visible
    ? [
        {
          id: String(state.recoveryId),
          label: "Network Flow import",
          origin: "Network Analysis",
          sheetRef: networkAnalysisSheetRef(),
          summary: completed
            ? "Import completed"
            : attention
              ? "Review retained import"
              : state.stage === "mapping"
                ? "Mapping retained"
                : "Import in progress",
          attention: completed
            ? "completed"
            : attention
              ? "attention"
              : state.stage === "mapping"
                ? "draft"
                : "progress",
          order: state.recoveryId,
        },
      ]
    : [];
  const selected = useWorkbookRecoverySource("network-import", items, {
    activate: () => {
      controller.setPresented(true);
      return true;
    },
    detach: () => controller.setPresented(false),
  });
  useWorkbookRecoveryPresentation(
    "network-import",
    visible && state.presented ? String(state.recoveryId) : null,
    selected,
  );
  return (
    <WorkbookRecoveryDetail source="network-import" item={selected}>
      <div className={networkFlowChromeRootClassName}>
        <NetworkFlowChromeStyles />
        {visible && state.presented ? (
          <NetworkFlowMappingPanel controller={controller} state={state} />
        ) : null}
      </div>
    </WorkbookRecoveryDetail>
  );
}
