import { useEffect, useLayoutEffect, useSyncExternalStore } from "react";
import type { NetworkFlowIndicatorLinkController } from "./NetworkFlowIndicatorLinkController";

/** Presentation subscribes; the workbook owns admission, settlement and recovery. */
export function useNetworkFlowIndicatorLinkController(options: {
  readonly controller: NetworkFlowIndicatorLinkController;
  readonly selectionContext: string;
}) {
  const { controller } = options;
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useEffect(() => controller.activate(), [controller]);
  useLayoutEffect(
    () => controller.setSelectionContext(options.selectionContext),
    [controller, options.selectionContext],
  );
  return state;
}
