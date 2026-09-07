import { useLayoutEffect, useSyncExternalStore } from "react";
import type { DeploymentUsersController } from "./deploymentUsersModel";

export function useDeploymentUsers(
  controller: DeploymentUsersController,
  {
    active = true,
    enterpriseAuthClaimed = false,
  }: { active?: boolean; enterpriseAuthClaimed?: boolean },
) {
  const snapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useLayoutEffect(() => {
    if (active) controller.open();
    else controller.close();
    controller.configure(enterpriseAuthClaimed);
  }, [controller, active, enterpriseAuthClaimed]);
  useLayoutEffect(() => controller.close, [controller]);
  return { snapshot, commands: controller };
}
