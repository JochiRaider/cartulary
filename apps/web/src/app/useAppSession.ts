import {
  useEffect,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type { AppSessionController } from "./appSessionController";

/** React binds presentation and lifetime only; the controller owns session decisions. */
export function useAppSession(
  controller: AppSessionController,
  disposeResources: () => void,
) {
  const mounted = useRef<AppSessionController | null>(null);
  const committed = useRef({ controller, disposeResources, disposed: false });
  useLayoutEffect(() => {
    if (committed.current.controller === controller)
      committed.current.disposeResources = disposeResources;
    else committed.current = { controller, disposeResources, disposed: false };
  });
  useEffect(() => {
    const lifetime = committed.current;
    mounted.current = controller;
    controller.start();
    return () => {
      mounted.current = null;
      controller.stop();
      queueMicrotask(() => {
        if (mounted.current === controller || lifetime.disposed) return;
        lifetime.disposed = true;
        controller.dispose();
        lifetime.disposeResources();
      });
    };
  }, [controller]);
  return useSyncExternalStore(controller.subscribe, controller.getSnapshot);
}
