import { useLayoutEffect, useMemo, useRef, useState } from "react";
import type { NetworkFlowImportSurfaceBinding } from "../../app/useNetworkFlowImport";
import { useIncidentCollaborationSession } from "../../collaboration/IncidentCollaborationSession";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import {
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../../extensions/extensionWorkspaceIdentities";
import { ImportClient } from "../../services/importClient";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { NetworkFlowImportController } from "../features/NetworkFlowOperations";

export function useNetworkFlowImportBinding(options: {
  readonly controller: NetworkFlowImportController;
  readonly bind: (binding: NetworkFlowImportSurfaceBinding | null) => void;
  readonly incidentId: string;
  readonly apiBase: string | undefined;
  readonly availability: ExtensionAvailabilityController;
  readonly available: boolean;
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
  const callbacks = useRef(options);
  callbacks.current = options;
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
  const { incidentId, available, role, closed, controller, lifecycleVersion } =
    options;
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
        if (event.kind === "session_revoked") controller.retire();
        if (event.kind === "authorization_lost") {
          controller.pause();
          void callbacks.current.recoverAccess();
        }
      }),
    [controller, session],
  );
  useLayoutEffect(() => {
    callbacks.current.bind({
      incidentId,
      available,
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
  }, [
    incidentId,
    available,
    role,
    closed,
    client,
    closureNotice,
    lifecycleVersion,
  ]);
  useLayoutEffect(
    () => () => {
      controller.setPresented(false);
      callbacks.current.bind(null);
    },
    [controller],
  );
}
