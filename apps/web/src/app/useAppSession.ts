import { useEffect, useRef, useSyncExternalStore } from "react";
import type { AppSessionController } from "./appSessionController";

/** React binds presentation and lifetime only; the controller owns session decisions. */
export function useAppSession(
  controller: AppSessionController,
  disposeResources: () => void,
) {
  const mounted = useRef(false);
  const dispose = useRef(disposeResources);
  dispose.current = disposeResources;
  useEffect(() => {
    mounted.current = true;
    controller.start();
    return () => {
      mounted.current = false;
      controller.stop();
      queueMicrotask(() => {
        if (mounted.current) return;
        controller.dispose();
        dispose.current();
      });
    };
  }, [controller]);
  return useSyncExternalStore(controller.subscribe, controller.getSnapshot);
}
