import { useLayoutEffect, useMemo, useRef, useSyncExternalStore } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
} from "../../extensions/extensionWorkspaceIdentities";
import type { WorkbookImportController } from "../../imports/WorkbookImportController";
import type { WorkbookImportSurfaceBinding } from "../../imports/workbookImportBinding";
import { ImportClient } from "../../services/importClient";
import { useWorkflowAttachment } from "../../shared/useWorkflowAttachment";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { AttachWorkflow } from "../../shared/workflowAttachment";

export function useWorkbookImportBinding(options: {
  readonly controller: WorkbookImportController;
  readonly attach: AttachWorkflow<WorkbookImportSurfaceBinding>;
  readonly incidentId: string;
  readonly apiBase: string | undefined;
  readonly availability: ExtensionAvailabilityController;
  readonly role: WorkbookIncidentRole | null;
  readonly closed: boolean;
  readonly recoverAccess: () => Promise<void>;
}) {
  const readiness = useSyncExternalStore(
    (listener) => options.availability.subscribeAuthority(listener),
    () =>
      options.availability.routeReadiness(importProfileId, importRouteFamily),
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
      }),
    [options.availability, options.apiBase, options.incidentId],
  );
  const { incidentId, role, closed } = options;
  useWorkflowAttachment(options.attach, {
    incidentId,
    readiness,
    role,
    closed,
    client,
    accessFailure: () => {
      void callbacks.current.recoverAccess();
    },
  });
}
