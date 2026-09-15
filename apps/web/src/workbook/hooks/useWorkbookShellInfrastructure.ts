import { useEffect, useMemo, useRef } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import type { AuthorizationRecoveryPort } from "../../shared/authorizationRecovery";
import { createContextualCreateReader } from "../adapters/createContextualCreateReader";
import { createContextualCreateTransport } from "../adapters/createContextualCreateTransport";
import { createCoordinationCreateTransport } from "../adapters/createCoordinationCreateTransport";
import { createEvidenceFileTransport } from "../adapters/createEvidenceFileTransport";
import { createIndicatorCreateTransport } from "../adapters/createIndicatorCreateTransport";
import { createIndicatorLifecycleAdapter } from "../adapters/createIndicatorLifecycleAdapter";
import { createNoteCreateReader } from "../adapters/createNoteCreateReader";
import { createNoteCreateTransport } from "../adapters/createNoteCreateTransport";
import { createObservationReader } from "../adapters/createObservationReader";
import { createObservationTransport } from "../adapters/createObservationTransport";
import { createOrdinaryCreateTransport } from "../adapters/createOrdinaryCreateTransport";
import { createPartyCreationTransport } from "../adapters/createPartyCreationTransport";
import { createPartyLinkReader } from "../adapters/createPartyLinkReader";
import { createTimelineFileLinkTransport } from "../adapters/createTimelineFileLinkTransport";
import { createTimelineRelatedEvidenceTransport } from "../adapters/createTimelineRelatedEvidenceTransport";
import { createWorkbookAuthoringReader } from "../adapters/createWorkbookAuthoringReader";
import { createWorkbookBatchTransport } from "../adapters/createWorkbookBatchTransport";
import { createWorkbookClipboardPasteAdapter } from "../adapters/createWorkbookClipboardPasteAdapter";
import { createWorkbookDecisionSupersessionAdapter } from "../adapters/createWorkbookDecisionSupersessionAdapter";
import { createWorkbookEntityMergeAdapter } from "../adapters/createWorkbookEntityMergeAdapter";
import { createWorkbookIncidentAdapter } from "../adapters/createWorkbookIncidentAdapter";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { createWorkbookReferenceMemberReader } from "../adapters/createWorkbookReferenceMemberReader";
import { createWorkbookStartupAdapter } from "../adapters/createWorkbookStartupAdapter";
import { createWorkbookViewQueryAdapter } from "../adapters/createWorkbookViewQueryAdapter";
import { readWorkbookAuthoringRecord } from "../adapters/readWorkbookAuthoringRecord";
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
import { createWorkbookReferenceReader } from "../services/workbookReferenceReader";
import { timelineCaptureOwnerFor } from "../timeline/actions/timelineCaptureOwnerFor";
import { timelineMentionOwnerFor } from "../timeline/actions/timelineMentionOwnerFor";
import { createTimelineCandidateReader } from "../timeline/adapters/createTimelineCandidateReader";
import { createTimelineMentionEntityCreationAdapter } from "../timeline/adapters/createTimelineMentionEntityCreationAdapter";
import { createTimelineMentionResolutionAdapter } from "../timeline/adapters/createTimelineMentionResolutionAdapter";
import { createTimelineRecordActionAdapter } from "../timeline/adapters/createTimelineRecordActionAdapter";
import { timelineMutationOwnerFor } from "../timeline/mutations/WorkbookTimelineMutationOwner";
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
  readonly onAuthorityUncertain: (() => void) | undefined;
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
  onAuthorityUncertain,
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
        onAuthorityUncertain,
      ),
    [mutationRuntime, apiBase, incidentId, onAuthorityUncertain],
  );
  useMemo(
    () =>
      mutationRuntime.indicatorObservations.configure(
        createObservationReader({ apiBase, incidentId }),
        createObservationTransport({ apiBase, incidentId }),
        onAuthorityUncertain,
      ),
    [mutationRuntime, apiBase, incidentId, onAuthorityUncertain],
  );
  useMemo(
    () =>
      mutationRuntime.indicatorCreate.configure(
        createObservationReader({ apiBase, incidentId }),
        createIndicatorCreateTransport({ apiBase, incidentId }),
        onAuthorityUncertain,
      ),
    [mutationRuntime, apiBase, incidentId, onAuthorityUncertain],
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
        onAuthorityUncertain,
      ),
    [timelineCapture, apiBase, incidentId, onAuthorityUncertain],
  );
  useMemo(
    () =>
      mutationRuntime.decisionSupersession.configure(
        createWorkbookDecisionSupersessionAdapter({ apiBase, incidentId }),
        onAuthorityUncertain,
      ),
    [apiBase, incidentId, mutationRuntime, onAuthorityUncertain],
  );
  useMemo(
    () =>
      mutationRuntime.explicitPatches.configure(
        createRecordPatchTransport(
          createWorkbookOperationExecutor({ apiBase }),
        ),
        onAuthorityUncertain,
        (view, recordId, signal) =>
          readWorkbookAuthoringRecord(
            createWorkbookAuthoringReader({
              apiBase,
              incidentId,
              recheckAuthority: () => {
                void recheckMentionAuthority();
              },
            }),
            view,
            recordId,
            signal,
          ),
      ),
    [
      apiBase,
      incidentId,
      mutationRuntime,
      onAuthorityUncertain,
      recheckMentionAuthority,
    ],
  );
  const clipboardPastePort = useMemo(
    () => createWorkbookClipboardPasteAdapter(mutationRuntime.batches),
    [mutationRuntime],
  );
  useMemo(
    () =>
      mutationRuntime.batches.configure(
        createWorkbookBatchTransport({ apiBase, incidentId }),
      ),
    [apiBase, incidentId, mutationRuntime],
  );
  const mutationCommands = useMemo(
    () =>
      createWorkbookMutationCommandPorts({
        batches: mutationRuntime.batches,
        apiBase,
        incidentId,
        transactionIds,
      }),
    [apiBase, incidentId, transactionIds, mutationRuntime],
  );
  useMemo(
    () =>
      mutationRuntime.history.configure(
        mutationCommands.records,
        onAuthorityUncertain,
      ),
    [mutationRuntime, mutationCommands, onAuthorityUncertain],
  );
  useMemo(
    () =>
      mutationRuntime.entityMerge.configure(
        createWorkbookEntityMergeAdapter({ apiBase, incidentId }),
        onAuthorityUncertain,
      ),
    [apiBase, incidentId, mutationRuntime, onAuthorityUncertain],
  );
  useMemo(() => {
    const reader = createWorkbookAuthoringReader({
      apiBase,
      incidentId,
      recheckAuthority: () => {
        void recheckMentionAuthority();
      },
    });
    timelineMutationOwnerFor(mutationRuntime).configureReader(
      (recordId, signal) =>
        readWorkbookAuthoringRecord(
          reader,
          "cartulary.view.timeline.v2",
          recordId,
          signal,
        ),
    );
  }, [apiBase, incidentId, mutationRuntime, recheckMentionAuthority]);
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
          mutationRuntime.ordinaryCreate.suspend();
          mutationRuntime.coordinationCreate.suspend();
          mutationRuntime.contextualCreate.suspend();
          mutationRuntime.timelineRelatedEvidence.suspend();
          mutationRuntime.evidenceAttachments.suspend();
          mutationRuntime.timelineFiles.suspend();
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
      mutationRuntime.ordinaryCreate.configure(
        currentAuthorityReader,
        createOrdinaryCreateTransport(apiBase),
        () => {
          void recheckMentionAuthority();
        },
      ),
    [mutationRuntime, currentAuthorityReader, apiBase, recheckMentionAuthority],
  );
  useMemo(
    () =>
      mutationRuntime.ordinaryCreate.configureReader(
        createWorkbookAuthoringReader({
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void recheckMentionAuthority();
          },
        }),
      ),
    [mutationRuntime, apiBase, incidentId, recheckMentionAuthority],
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
  useMemo(() => {
    const reader = createWorkbookAuthoringReader({
      apiBase,
      incidentId,
      recheckAuthority: () => {
        void mutationRuntime.evidenceAttachments.recheckAuthority();
        void mutationRuntime.timelineFiles.recheckAuthority();
      },
    });
    const transport = createEvidenceFileTransport(apiBase);
    mutationRuntime.evidenceAttachments.configure(
      reader,
      currentAuthorityReader,
      transport,
    );
    mutationRuntime.timelineFiles.configure(
      reader,
      currentAuthorityReader,
      transport,
      createTimelineFileLinkTransport(apiBase),
    );
  }, [apiBase, incidentId, mutationRuntime, currentAuthorityReader]);
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
    onAuthorityUncertain,
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

export function useWorkbookReferenceReader(
  authorityScope: string,
  viewQuery: ReturnType<typeof createWorkbookViewQueryAdapter>,
  apiBase: string | undefined,
  incidentId: string,
) {
  const reader = useMemo(
    () =>
      createWorkbookReferenceReader({
        authorityScope,
        viewQuery,
        readMembers: createWorkbookReferenceMemberReader({
          apiBase,
          incidentId,
        }),
      }),
    [authorityScope, viewQuery, apiBase, incidentId],
  );
  useEffect(() => () => reader.dispose(), [reader]);
  return reader;
}
