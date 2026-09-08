import {
  type ReactNode,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { IncidentResource } from "../shared/incidentResource";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import { metadataActorId } from "../testing/incidentMetadataTestSupport";
import {
  mutateIncidentLifecycle,
  readLifecycleIncident,
} from "./api/incidentLifecycleClient";
import { IncidentLifecycleFeature } from "./IncidentLifecyclePanel";
import { IncidentLifecycleController } from "./incidentLifecycleController";
import { IncidentResourceController } from "./incidentResourceController";

/** Same retained-controller/conditional-surface composition as App. */
export function LifecycleTestSurface(props: {
  preferenceControls?: ReactNode;
  activeSection: "summary" | "memberships" | null;
  currentIncidentRole: "admin" | "viewer";
  incidentId: string;
  acceptedIncident?: IncidentResource | undefined;
  onSessionRoleChange?: (() => Promise<void>) | undefined;
}) {
  const current = useRef(props);
  current.current = props;
  const [resources] = useState(
    () =>
      new IncidentResourceController({
        isCurrent: (a) => a.incidentId === current.current.incidentId,
        accepted: () => {},
      }),
  );
  const [controller] = useState(
    () =>
      new IncidentLifecycleController({
        read: readLifecycleIncident,
        mutate: mutateIncidentLifecycle,
        isCurrent: (a) => a.incidentId === current.current.incidentId,
        lost: () => {},
        recover: async (a) => {
          if (controller.getSnapshot().operation.kind === "confirmed")
            await current.current.onSessionRoleChange?.();
          return {
            kind: "authorized",
            userId: a.actorId,
            role: current.current.currentIncidentRole,
          };
        },
        publishResource: resources.accept,
      }),
  );
  const accepted = useSyncExternalStore(
    resources.subscribe,
    resources.getSnapshot,
  );
  const authority = () => ({
    incidentId: current.current.incidentId,
    actorId: metadataActorId,
    lifetime: "session",
    role: current.current.currentIncidentRole,
  });
  useLayoutEffect(() => {
    controller.setAuthority(authority());
    if (props.acceptedIncident)
      resources.accept(props.acceptedIncident, authority());
    if (accepted?.incident_id === props.incidentId)
      controller.acceptResource(accepted, false);
  });
  useLayoutEffect(() => () => controller.dispose(), [controller]);
  const bindSurface = (
    surface: WorkbookIncidentControlsRendererProps | null,
  ) => {
    controller.setAuthority(authority());
    controller.setActive(surface?.activeSection === "summary");
  };
  return props.activeSection === "summary" ? (
    <IncidentLifecycleFeature
      incidentId={props.incidentId}
      currentIncidentRole={props.currentIncidentRole}
      activeSection="summary"
      controller={controller}
      preferenceControls={props.preferenceControls}
      bindSurface={bindSurface}
      acceptedIncident={
        accepted?.incident_id === props.incidentId ? accepted : null
      }
      onIncidentObserved={(resource) => resources.accept(resource, authority())}
    />
  ) : null;
}
