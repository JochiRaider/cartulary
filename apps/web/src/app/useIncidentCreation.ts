import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { createIncident } from "./api/appShellClient";
import {
  IncidentCreationController,
  type IncidentCreationPorts,
} from "./incidentCreationModel";

export function useIncidentCreation(options: {
  sessionIdentity: string | null;
  sessionLost: IncidentCreationPorts["sessionLost"];
  openIncident: IncidentCreationPorts["openIncident"];
}) {
  const current = useRef(options);
  current.current = options;
  const [controller] = useState(
    () =>
      new IncidentCreationController({
        create: (request, signal) => createIncident({ request, signal }),
        isCurrentSession: (identity) =>
          current.current.sessionIdentity === identity,
        sessionLost: () => current.current.sessionLost(),
        openIncident: (incident, signal, canNavigate) =>
          current.current.openIncident(incident, signal, canNavigate),
      }),
  );
  useLayoutEffect(() => {
    controller.setSession(options.sessionIdentity);
  }, [controller, options.sessionIdentity]);
  useLayoutEffect(() => () => controller.dispose(), [controller]);
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  return { controller, state };
}
