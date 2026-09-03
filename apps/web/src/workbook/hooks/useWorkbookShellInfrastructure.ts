import { useEffect, useMemo, useRef } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import { createWorkbookClipboardPasteAdapter } from "../adapters/createWorkbookClipboardPasteAdapter";
import { createWorkbookIncidentAdapter } from "../adapters/createWorkbookIncidentAdapter";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { createWorkbookPreferenceAdapter } from "../adapters/createWorkbookPreferenceAdapter";
import { createWorkbookSavedViewAdapter } from "../adapters/createWorkbookSavedViewAdapter";
import { createWorkbookStartupAdapter } from "../adapters/createWorkbookStartupAdapter";
import { createWorkbookViewQueryAdapter } from "../adapters/createWorkbookViewQueryAdapter";
import { createWorkbookMutationCommandPorts } from "../mutations/createWorkbookMutationCommandPorts";
import { createBrowserSecureTransactionIdPort } from "../mutations/secureTransactionId";
import { useWorkbookMutationRuntime } from "../runtime/useWorkbookMutationRuntime";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { WorkbookMutationRuntimeRegistry } from "../runtime/WorkbookMutationRuntimeRegistry";
import {
  createReferenceQueryBroker,
  type ReferenceQueryBrokerPort,
} from "../services/referenceQueryBroker";
import { useWorkbookShellRuntime } from "./useWorkbookShellRuntime";

function recordPendingMutationTiming(
  name: string,
  details: Readonly<Record<string, unknown>> = {},
) {
  if (typeof performance === "undefined") return;
  performance.mark(`cartulary.workbook.${name}`, { detail: details });
}

type WorkbookShellInfrastructureOptions = {
  readonly apiBase: string | undefined;
  readonly clientInstanceId: string;
  readonly extensionAvailability: ExtensionAvailabilityController;
  readonly incidentId: string;
  readonly mutationRuntimeRegistry: WorkbookMutationRuntimeRegistry;
  readonly onExtensionAvailabilityChange: () => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
};

/** Constructs incident-scoped adapters and exactly one registry-owned runtime. */
export function useWorkbookShellInfrastructure({
  apiBase,
  clientInstanceId,
  extensionAvailability,
  incidentId,
  mutationRuntimeRegistry,
  onExtensionAvailabilityChange,
  onIncidentAccessLost,
}: WorkbookShellInfrastructureOptions) {
  const transactionIds = useMemo(createBrowserSecureTransactionIdPort, []);
  const pendingMutationPort = useMemo(
    () =>
      createWorkbookPendingMutationAdapter({
        apiBase,
        incidentId,
        recordTiming: recordPendingMutationTiming,
      }),
    [apiBase, incidentId],
  );
  const clipboardPastePort = useMemo(
    () =>
      createWorkbookClipboardPasteAdapter({
        apiBase,
        incidentId,
        transactionIds,
      }),
    [apiBase, incidentId, transactionIds],
  );
  const mutationRuntime = useMemo(
    () =>
      mutationRuntimeRegistry.acquire(
        { clientInstanceId, incidentId },
        () =>
          new WorkbookMutationRuntime(
            { clientInstanceId, incidentId },
            transactionIds,
            pendingMutationPort,
          ),
      ),
    [
      clientInstanceId,
      incidentId,
      mutationRuntimeRegistry,
      pendingMutationPort,
      transactionIds,
    ],
  );
  const mutationCommands = useMemo(
    () =>
      createWorkbookMutationCommandPorts({
        apiBase,
        incidentId,
        transactionIds,
      }),
    [apiBase, incidentId, transactionIds],
  );
  const mutationSnapshot = useWorkbookMutationRuntime(mutationRuntime);
  const surfaceSelectionVersionRef = useRef(0);
  const incidentPort = useMemo(
    () => createWorkbookIncidentAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const preferencePort = useMemo(
    () => createWorkbookPreferenceAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const startupPort = useMemo(
    () => createWorkbookStartupAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const savedViewPort = useMemo(
    () => createWorkbookSavedViewAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const viewQuery = useMemo(
    () => createWorkbookViewQueryAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const workbookRuntime = useWorkbookShellRuntime({
    incidentId,
    onIncidentAccessLost,
    surfaceSelectionVersionRef,
    extensionAvailability,
    onExtensionAvailabilityChange,
    preferencePort,
    savedViewPort,
    startupPort,
  });

  return {
    clipboardPastePort,
    incidentPort,
    mutationCommands,
    mutationRuntime,
    mutationSnapshot,
    viewQuery,
    workbookRuntime,
  };
}

export function useWorkbookReferenceQueryBroker(
  authorizationGeneration: string,
  viewQuery: ReturnType<typeof createWorkbookViewQueryAdapter>,
): ReferenceQueryBrokerPort {
  const broker = useMemo(
    () =>
      createReferenceQueryBroker({
        authorizationGeneration,
        viewQuery,
      }),
    [authorizationGeneration, viewQuery],
  );
  useEffect(
    () => () => {
      broker.dispose();
    },
    [broker],
  );
  return broker;
}
