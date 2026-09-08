import { useLayoutEffect, useMemo, useRef } from "react";
import type { WorkbookImportSurfaceBinding } from "../../app/useWorkbookImport";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import type { WorkbookImportController } from "../../imports/WorkbookImportController";
import { ImportClient } from "../../services/importClient";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

export function useWorkbookImportBinding(options: {
  readonly controller: WorkbookImportController;
  readonly bind: (binding: WorkbookImportSurfaceBinding | null) => void;
  readonly incidentId: string;
  readonly apiBase: string | undefined;
  readonly availability: ExtensionAvailabilityController;
  readonly available: boolean;
  readonly role: WorkbookIncidentRole | null;
  readonly closed: boolean;
  readonly recoverAccess: () => Promise<void>;
}) {
  const callbacks = useRef(options);
  callbacks.current = options;
  const client = useMemo(
    () =>
      new ImportClient({
        availability: options.availability,
        apiBase: options.apiBase,
        incidentId: options.incidentId,
      }),
    [options.availability, options.apiBase, options.incidentId],
  );
  const { incidentId, available, role, closed, controller } = options;
  useLayoutEffect(() => {
    callbacks.current.bind({
      incidentId,
      available,
      role,
      closed,
      client,
      accessFailure: () => {
        void callbacks.current.recoverAccess();
      },
    });
  }, [incidentId, available, role, closed, client]);
  useLayoutEffect(
    () => () => {
      controller.setPresented(false);
      callbacks.current.bind(null);
    },
    [controller],
  );
}
