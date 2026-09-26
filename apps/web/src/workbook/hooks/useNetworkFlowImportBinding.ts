import {
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { useIncidentCollaborationSession } from "../../collaboration/IncidentCollaborationSession";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../../extensions/extensionWorkspaceIdentities";
import { ImportClient } from "../../services/importClient";
import { useWorkflowAttachment } from "../../shared/useWorkflowAttachment";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { AttachWorkflow } from "../../shared/workflowAttachment";
import type {
  NetworkFlowImportController,
  NetworkFlowImportSurfaceBinding,
} from "../features/NetworkFlowOperations";

export function useNetworkFlowImportBinding(options: {
  readonly controller: NetworkFlowImportController;
  readonly attach: AttachWorkflow<NetworkFlowImportSurfaceBinding>;
  readonly incidentId: string;
  readonly apiBase: string | undefined;
  readonly availability: ExtensionAvailabilityController;
  readonly role: WorkbookIncidentRole | null;
  readonly closed: boolean;
  readonly lifecycleVersion: number | undefined;
  readonly recoverIncident: () => Promise<void>;
  readonly recoverAccess: () => Promise<void>;
}) {
  const session = useIncidentCollaborationSession();
  const [closureNotice, setClosureNotice] = useState<{
    incidentId: string;
    lifecycleVersion: number | undefined;
  } | null>(null);
  const readiness = useSyncExternalStore(
    (listener) => options.availability.subscribeAuthority(listener),
    () =>
      (() => {
        const availability = options.availability;
        const importState = availability.routeReadiness(
          importProfileId,
          importRouteFamily,
        );
        const networkState = availability.routeReadiness(
          networkFlowActivityProfileId,
          networkFlowRouteFamily,
        );
        if (importState === "invalid" || networkState === "invalid")
          return "invalid";
        if (importState === "unavailable" || networkState === "unavailable")
          return "unavailable";
        if (importState === "pending" || networkState === "pending")
          return "pending";
        return "available";
      })(),
  );
  const callbacks = useRef(options);
  useLayoutEffect(() => {
    callbacks.current = options;
  });
  const client = useMemo(
    () =>
      new ImportClient({
        availability: options.availability,
        apiBase: options.apiBase,
        incidentId: options.incidentId,
        canDispatch: () => {
          const current = callbacks.current;
          return (
            current.controller.getSnapshot().access === "active" &&
            !current.controller.getSnapshot().closed &&
            !current.closed &&
            current.role !== null &&
            current.role !== "" &&
            current.availability.isRouteAvailable(
              networkFlowActivityProfileId,
              networkFlowRouteFamily,
            )
          );
        },
      }),
    [options.availability, options.apiBase, options.incidentId],
  );
  const { incidentId, role, closed, controller, lifecycleVersion } = options;
  useLayoutEffect(
    () =>
      session.subscribe((event) => {
        if (event.kind === "incident_closed") {
          const retained = controller.getSnapshot();
          if (retained.access !== "active" || retained.stage === "idle") return;
          controller.observeClosure();
          setClosureNotice({
            incidentId: callbacks.current.incidentId,
            lifecycleVersion: callbacks.current.lifecycleVersion,
          });
          void callbacks.current.recoverIncident();
        }
        if (event.kind === "authorization_revoked") {
          if (event.scope === "session") controller.retire();
          else controller.pause();
        }
        if (event.kind === "authorization_lost") {
          controller.pause();
          void callbacks.current.recoverAccess();
        }
      }),
    [controller, session],
  );
  useWorkflowAttachment(options.attach, {
    incidentId,
    readiness,
    role,
    closed:
      closed ||
      (closureNotice?.incidentId === incidentId &&
        closureNotice.lifecycleVersion === lifecycleVersion),
    client,
    accessFailure: (failure) => {
      if (failure.code === "incident_closed") {
        setClosureNotice({
          incidentId: callbacks.current.incidentId,
          lifecycleVersion: callbacks.current.lifecycleVersion,
        });
        void callbacks.current.recoverIncident();
      } else {
        void callbacks.current.recoverAccess();
      }
    },
  });
}
