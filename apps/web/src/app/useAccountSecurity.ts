import { useLayoutEffect, useSyncExternalStore } from "react";
import type { AccountSecurityController } from "./accountSecurityModel";
export function useAccountSecurity(controller: AccountSecurityController) {
  const snapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useLayoutEffect(() => {
    controller.open();
    return controller.close;
  }, [controller]);
  return { snapshot, commands: controller };
}
