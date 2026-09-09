import { useEffect, useSyncExternalStore } from "react";
import type { SavedGraphController } from "./SavedGraphController";

/** React subscribes to the workbook-owned saved graph operation and navigation. */
export function useNetworkFlowSavedGraphController(options: {
  readonly controller: SavedGraphController;
  readonly enabled: boolean;
}) {
  const { controller } = options;
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useEffect(() => {
    controller.setActive(options.enabled);
    return () => controller.setActive(false);
  }, [controller, options.enabled]);
  return {
    ...state,
    ...state.navigation,
    mutationPending: controller.mutationPending,
    selectedGraph:
      state.graphs.find(
        (graph) => graph.graph_view_id === state.selectedGraphViewId,
      ) ?? null,
    loadGraphs: controller.loadGraphs,
    selectGraphView: controller.selectGraphView,
    openAction: controller.openAction,
    closeDialog: controller.closeDialog,
    reopenOperation: controller.reopenOperation,
    setDraft: controller.setDraft,
    submit: controller.submit,
    replay: controller.replay,
    reviewCurrent: controller.reviewCurrent,
    resumeObservation: controller.resumeObservation,
    loadResult: controller.recoverResult,
    selectObject: controller.navigation.selectObject,
    loadContributors: controller.navigation.loadContributors,
    setPage: controller.navigation.setPage,
  };
}
