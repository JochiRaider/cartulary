import { useEffect, useMemo, useRef } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import type { AuthorizationRecoveryPort } from "../../shared/authorizationRecovery";
import { createContextualCreateReader } from "../adapters/createContextualCreateReader";
import { createContextualCreateTransport } from "../adapters/createContextualCreateTransport";
import { createCoordinationCreateTransport } from "../adapters/createCoordinationCreateTransport";
import { createIndicatorCreateTransport } from "../adapters/createIndicatorCreateTransport";
import { createIndicatorLifecycleAdapter } from "../adapters/createIndicatorLifecycleAdapter";
import { createNoteCreateReader } from "../adapters/createNoteCreateReader";
import { createNoteCreateTransport } from "../adapters/createNoteCreateTransport";
import { createObservationReader } from "../adapters/createObservationReader";
import { createObservationTransport } from "../adapters/createObservationTransport";
import { createPartyCreationTransport } from "../adapters/createPartyCreationTransport";
import { createPartyLinkReader } from "../adapters/createPartyLinkReader";
import { createTimelineRelatedEvidenceTransport } from "../adapters/createTimelineRelatedEvidenceTransport";
import { createWorkbookAuthoringReader } from "../adapters/createWorkbookAuthoringReader";
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
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
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
import { timelineMentionOwnerFor } from "../timeline/actions/timelineMentionOwnerFor";
import { createTimelineCandidateReader } from "../timeline/adapters/createTimelineCandidateReader";
import { createTimelineMentionEntityCreationAdapter } from "../timeline/adapters/createTimelineMentionEntityCreationAdapter";
import { createTimelineMentionResolutionAdapter } from "../timeline/adapters/createTimelineMentionResolutionAdapter";
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
  readonly partyAuthorization: AuthorizationRecoveryPort;
  readonly recheckMentionAuthority: () => Promise<void>;
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
  recheckMentionAuthority,
  partyAuthorization,
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
  useMemo(
    () =>
      mutationRuntime.indicatorObservations.configure(
        createObservationReader({ apiBase, incidentId }),
        createObservationTransport({ apiBase, incidentId }),
        onIncidentAccessLost,
      ),
    [mutationRuntime, apiBase, incidentId, onIncidentAccessLost],
  );
  useMemo(
    () =>
      mutationRuntime.indicatorCreate.configure(
        createObservationReader({ apiBase, incidentId }),
        createIndicatorCreateTransport({ apiBase, incidentId }),
        onIncidentAccessLost,
      ),
    [mutationRuntime, apiBase, incidentId, onIncidentAccessLost],
  );
  const timelineMentions = useMemo(
    () => timelineMentionOwnerFor(mutationRuntime),
    [mutationRuntime],
  );
  useMemo(() => {
    timelineMentions.configure(
      createTimelineMentionResolutionAdapter({ apiBase }),
      () => {
        void recheckMentionAuthority();
      },
    );
    timelineMentions.configureCreation(
      createTimelineMentionEntityCreationAdapter({ apiBase }),
    );
  }, [timelineMentions, apiBase, recheckMentionAuthority]);
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
  const currentAuthorityReader = useMemo(
    () => async (baseline: WorkbookMutationAuthority, signal: AbortSignal) => {
      const result = await partyAuthorization.recover({
        incidentId,
        signal,
      });
      if (signal.aborted) throw new Error("Authority read interrupted.");
      if (result.kind !== "authorized") {
        if (result.kind === "access_lost" || result.kind === "session_lost") {
          mutationRuntime.assessmentAuthoring.suspend();
          mutationRuntime.noteCreate.suspend();
          mutationRuntime.coordinationCreate.suspend();
          mutationRuntime.contextualCreate.suspend();
          mutationRuntime.timelineRelatedEvidence.suspend();
          mutationRuntime.partyLinks.suspend();
          mutationRuntime.explicitPatches.suspend();
          void recheckMentionAuthority();
        }
        throw new Error("Current authority is unavailable. Retry the read.");
      }
      const incident = await incidentPort.getIdentity({ signal });
      if (incident.kind !== "accepted")
        throw new Error("Current incident state is unavailable.");
      authorizationRecovered(result);
      return {
        ...baseline,
        actorId: result.userId,
        role: result.role,
        closed: incident.value.status !== "active",
      };
    },
    [
      partyAuthorization,
      incidentId,
      mutationRuntime,
      incidentPort,
      recheckMentionAuthority,
      authorizationRecovered,
    ],
  );
  useMemo(
    () =>
      mutationRuntime.noteCreate.configure(
        createNoteCreateReader({
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void recheckMentionAuthority();
          },
        }),
        currentAuthorityReader,
        createNoteCreateTransport(apiBase),
      ),
    [
      apiBase,
      incidentId,
      mutationRuntime,
      currentAuthorityReader,
      recheckMentionAuthority,
    ],
  );
  useMemo(
    () =>
      mutationRuntime.coordinationCreate.configure(
        createWorkbookAuthoringReader({
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void recheckMentionAuthority();
          },
        }),
        currentAuthorityReader,
        createCoordinationCreateTransport(apiBase),
      ),
    [
      apiBase,
      incidentId,
      mutationRuntime,
      currentAuthorityReader,
      recheckMentionAuthority,
    ],
  );
  useMemo(
    () =>
      mutationRuntime.contextualCreate.configure(
        createContextualCreateReader({
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void mutationRuntime.contextualCreate.recheckAuthority();
          },
        }),
        currentAuthorityReader,
        createContextualCreateTransport(apiBase),
      ),
    [apiBase, incidentId, mutationRuntime, currentAuthorityReader],
  );
  useMemo(
    () =>
      mutationRuntime.timelineRelatedEvidence.configure(
        createWorkbookAuthoringReader({
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void mutationRuntime.timelineRelatedEvidence.recheckAuthority();
          },
        }),
        currentAuthorityReader,
        createTimelineRelatedEvidenceTransport(apiBase),
      ),
    [apiBase, incidentId, mutationRuntime, currentAuthorityReader],
  );
  useMemo(
    () =>
      mutationRuntime.assessmentAuthoring.configure(
        mutationCommands.assessment,
        currentAuthorityReader,
      ),
    [mutationRuntime, mutationCommands, currentAuthorityReader],
  );
  useMemo(
    () =>
      mutationRuntime.partyLinks.configure(
        createPartyCreationTransport(apiBase),
        createPartyLinkReader({
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void mutationRuntime.partyLinks.recheckAuthority();
          },
        }),
        currentAuthorityReader,
      ),
    [apiBase, incidentId, mutationRuntime, currentAuthorityReader],
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
    timelineMentions,
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
