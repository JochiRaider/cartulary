import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { listVisibleIncidents } from "./api/appShellClient";
import {
  IncidentDirectoryController,
  type IncidentDirectoryPorts,
} from "./incidentDirectoryModel";

export function useIncidentDirectory(options: {
  sessionIdentity: string | null;
  active: boolean;
  sessionLost: IncidentDirectoryPorts["sessionLost"];
}) {
  const current = useRef(options);
  current.current = options;
  const [controller] = useState(
    () =>
      new IncidentDirectoryController({
        list: ({ query, ...request }) =>
          listVisibleIncidents({
            ...request,
            search: query.search,
            status:
              query.statusFilter === "all" ? undefined : query.statusFilter,
          }),
        isCurrentSession: (identity) =>
          current.current.sessionIdentity === identity,
        sessionLost: () => current.current.sessionLost(),
      }),
  );
  useLayoutEffect(() => {
    controller.setSession(options.sessionIdentity);
    controller.setActive(options.active);
  }, [controller, options.sessionIdentity, options.active]);
  useLayoutEffect(() => () => controller.dispose(), [controller]);
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  return { controller, state };
}
