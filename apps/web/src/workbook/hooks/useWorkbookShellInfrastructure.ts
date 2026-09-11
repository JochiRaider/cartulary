import { useEffect, useMemo, useRef } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import { createIndicatorLifecycleAdapter } from "../adapters/createIndicatorLifecycleAdapter";
import { createWorkbookClipboardPasteAdapter } from "../adapters/createWorkbookClipboardPasteAdapter";
import { createWorkbookDecisionSupersessionAdapter } from "../adapters/createWorkbookDecisionSupersessionAdapter";
import { createWorkbookEntityMergeAdapter } from "../adapters/createWorkbookEntityMergeAdapter";
import { createWorkbookIncidentAdapter } from "../adapters/createWorkbookIncidentAdapter";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { createWorkbookStartupAdapter } from "../adapters/createWorkbookStartupAdapter";
import { createWorkbookViewQueryAdapter } from "../adapters/createWorkbookViewQueryAdapter";
import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import { createRecordPatchTransport } from "../adapters/workbookRecordPatchTransport";
import { createWorkbookMutationCommandPorts } from "../mutations/createWorkbookMutationCommandPorts";
import { createBrowserSecureTransactionIdPort } from "../mutations/secureTransactionId";
import { useWorkbookMutationRuntime } from "../runtime/useWorkbookMutationRuntime";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { WorkbookMutationRuntimeRegistry } from "../runtime/WorkbookMutationRuntimeRegistry";
import type { SavedViewBinding } from "../savedviews/savedViewOperationModel";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";
import {
  createReferenceQueryBroker,
  type ReferenceQueryBrokerPort,
} from "../services/referenceQueryBroker";
import { timelineCaptureOwnerFor } from "../timeline/actions/timelineCaptureOwnerFor";
import { createTimelineCandidateReader } from "../timeline/adapters/createTimelineCandidateReader";
import { createTimelineRecordActionAdapter } from "../timeline/adapters/createTimelineRecordActionAdapter";
import { useWorkbookShellRuntime } from "./useWorkbookShellRuntime";

function recordPendingMutationTiming(
  name: string,
  details: Readonly<Record<string, unknown>> = {},
) {
  if (typeof performance === "undefined") return;
  performance.mark(`cartulary.workbook.${name}`, { detail: details });
}

type WorkbookShellInfrastructureOptions = {
  readonly savedViewOwner: WorkbookSavedViewController;
  readonly bindWorkbookSavedViews: (binding: SavedViewBinding | null) => void;
  readonly authorizationRecovered: SavedViewBinding["authorizationRecovered"];
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
  savedViewOwner,
  bindWorkbookSavedViews,
  authorizationRecovered,
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
  useMemo(
    () =>
      mutationRuntime.indicatorLifecycle.configure(
        createIndicatorLifecycleAdapter({ apiBase, incidentId }),
        onIncidentAccessLost,
      ),
    [mutationRuntime, apiBase, incidentId, onIncidentAccessLost],
  );
  const timelineCapture = useMemo(
    () => timelineCaptureOwnerFor(mutationRuntime),
    [mutationRuntime],
  );
  useMemo(
    () =>
      timelineCapture.configure(
        createTimelineRecordActionAdapter({ apiBase }),
        createTimelineCandidateReader({ apiBase, incidentId }),
        onIncidentAccessLost,
      ),
    [timelineCapture, apiBase, incidentId, onIncidentAccessLost],
  );
  const entityWrites = useMemo(
    () => ({
      begin: mutationRuntime.beginEntityWrite.bind(mutationRuntime),
      acceptVersion: mutationRuntime.acceptEntityVersion.bind(mutationRuntime),
    }),
    [mutationRuntime],
  );
  const decisionWrites = useMemo(
    () => ({
      begin: mutationRuntime.beginDecisionWrite.bind(mutationRuntime),
      acceptRow: mutationRuntime.decisionSupersession.acceptRow.bind(
        mutationRuntime.decisionSupersession,
      ),
    }),
    [mutationRuntime],
  );
  useMemo(
    () =>
      mutationRuntime.decisionSupersession.configure(
        createWorkbookDecisionSupersessionAdapter({ apiBase, incidentId }),
        onIncidentAccessLost,
      ),
    [apiBase, incidentId, mutationRuntime, onIncidentAccessLost],
  );
  useMemo(
    () =>
      mutationRuntime.explicitPatches.configure(
        createRecordPatchTransport(
          createWorkbookOperationExecutor({ apiBase }),
        ),
        onIncidentAccessLost,
      ),
    [apiBase, mutationRuntime, onIncidentAccessLost],
  );
  const clipboardPastePort = useMemo(
    () =>
      createWorkbookClipboardPasteAdapter({
        apiBase,
        incidentId,
        transactionIds,
        entityWrites,
      }),
    [apiBase, incidentId, transactionIds, entityWrites],
  );
  const mutationCommands = useMemo(
    () =>
      createWorkbookMutationCommandPorts({
        apiBase,
        incidentId,
        transactionIds,
        entityWrites,
        decisionWrites,
      }),
    [apiBase, incidentId, transactionIds, entityWrites, decisionWrites],
  );
  useMemo(
    () => mutationRuntime.history.configure(mutationCommands.records),
    [mutationRuntime, mutationCommands],
  );
  useMemo(
    () =>
      mutationRuntime.entityMerge.configure(
        createWorkbookEntityMergeAdapter({ apiBase, incidentId }),
      ),
    [apiBase, incidentId, mutationRuntime],
  );
  const mutationSnapshot = useWorkbookMutationRuntime(mutationRuntime);
  const surfaceSelectionVersionRef = useRef(0);
  const incidentPort = useMemo(
    () => createWorkbookIncidentAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const startupPort = useMemo(
    () => createWorkbookStartupAdapter({ apiBase, incidentId }),
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
    savedViewOwner,
    bindWorkbookSavedViews,
    authorizationRecovered,
    apiBase,
    startupPort,
  });

  return {
    clipboardPastePort,
    incidentPort,
    mutationCommands,
    mutationRuntime,
    timelineCapture,
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
