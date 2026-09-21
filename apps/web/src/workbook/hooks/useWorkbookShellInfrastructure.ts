import { useEffect, useMemo, useRef, useSyncExternalStore } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import type { AuthorizationRecoveryPort } from "../../shared/authorizationRecovery";
import { createContextualCreateTransport } from "../adapters/createContextualCreateTransport";
import { createCoordinationCreateTransport } from "../adapters/createCoordinationCreateTransport";
import { createEvidenceFileTransport } from "../adapters/createEvidenceFileTransport";
import { createIndicatorCreateTransport } from "../adapters/createIndicatorCreateTransport";
import { createIndicatorLifecycleAdapter } from "../adapters/createIndicatorLifecycleAdapter";
import { createNoteAssociationReader } from "../adapters/createNoteAssociationReader";
import { createNoteAssociationTransport } from "../adapters/createNoteAssociationTransport";
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
import { createTimelineMentionSourceReader } from "../timeline/adapters/createTimelineMentionSourceReader";
import { createTimelineRecordActionAdapter } from "../timeline/adapters/createTimelineRecordActionAdapter";
import {
  normalizeTimelineFullRow,
  rowFromApi,
} from "../timeline/models/timelineRowModel";
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
  readonly acceptedAuthority: Readonly<{
    userId: string | null;
    role: string | null;
    revision: number;
  }>;
  readonly sessionIdentity: string | null;
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
  acceptedAuthority,
  sessionIdentity,
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
  const observationRuntime = useRef<WorkbookMutationRuntime | null>(null);
  const transactionIds = useMemo(createBrowserSecureTransactionIdPort, []);
  const pendingMutationPort = useMemo(
    () =>
      createWorkbookPendingMutationAdapter({
        apiBase,
        incidentId,
        recordTiming: recordPendingMutationTiming,
        readScope: () => observationRuntime.current?.recordReadScope ?? null,
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
  observationRuntime.current = mutationRuntime;
  useMemo(
    () =>
      mutationRuntime.indicatorLifecycle.configure(
        createIndicatorLifecycleAdapter({
          apiBase,
          incidentId,
          readScope: () => mutationRuntime.recordReadScope,
        }),
        onAuthorityUncertain,
      ),
    [mutationRuntime, apiBase, incidentId, onAuthorityUncertain],
  );
  useMemo(
    () =>
      mutationRuntime.indicatorObservations.configure(
        createObservationReader({
          apiBase,
          incidentId,
          readScope: () => mutationRuntime.recordReadScope,
        }),
        createObservationTransport({ apiBase, incidentId }),
        onAuthorityUncertain,
      ),
    [mutationRuntime, apiBase, incidentId, onAuthorityUncertain],
  );
  useMemo(
    () =>
      mutationRuntime.indicatorCreate.configure(
        createObservationReader({
          apiBase,
          incidentId,
          readScope: () => mutationRuntime.recordReadScope,
        }),
        createIndicatorCreateTransport({
          apiBase,
          incidentId,
          readScope: () => mutationRuntime.recordReadScope,
        }),
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
    const readMentionSource = createTimelineMentionSourceReader({
      readScope: () => mutationRuntime.recordReadScope,
      apiBase,
      incidentId,
    });
    timelineMentions.configureSourceReader(async (recordId, signal) =>
      rowFromApi(
        normalizeTimelineFullRow(
          await readMentionSource(recordId, signal),
          "mention disclosure source",
        ),
      ),
    );
    timelineMentions.configureCreation(
      createTimelineMentionEntityCreationAdapter({
        apiBase,
        readScope: () => mutationRuntime.recordReadScope,
      }),
    );
  }, [
    timelineMentions,
    apiBase,
    incidentId,
    recheckMentionAuthority,
    mutationRuntime,
  ]);
  const timelineCapture = useMemo(
    () => timelineCaptureOwnerFor(mutationRuntime),
    [mutationRuntime],
  );
  useMemo(
    () =>
      timelineCapture.configure(
        createTimelineRecordActionAdapter({
          apiBase,
          readScope: () => mutationRuntime.recordReadScope,
        }),
        createTimelineCandidateReader({ apiBase, incidentId }),
        onAuthorityUncertain,
      ),
    [
      timelineCapture,
      apiBase,
      incidentId,
      onAuthorityUncertain,
      mutationRuntime,
    ],
  );
  useMemo(
    () =>
      mutationRuntime.decisionSupersession.configure(
        createWorkbookDecisionSupersessionAdapter({
          apiBase,
          incidentId,
          readScope: () => mutationRuntime.recordReadScope,
        }),
        onAuthorityUncertain,
      ),
    [apiBase, incidentId, mutationRuntime, onAuthorityUncertain],
  );
  useMemo(
    () =>
      mutationRuntime.explicitPatches.configure(
        createRecordPatchTransport(
          createWorkbookOperationExecutor({ apiBase }),
          () => mutationRuntime.recordReadScope,
        ),
        onAuthorityUncertain,
        (view, recordId, signal) =>
          readWorkbookAuthoringRecord(
            createWorkbookAuthoringReader({
              readScope: () => mutationRuntime.recordReadScope,
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
        createWorkbookBatchTransport({
          apiBase,
          incidentId,
          readScope: () => mutationRuntime.recordReadScope,
        }),
      ),
    [apiBase, incidentId, mutationRuntime],
  );
  const mutationCommands = useMemo(
    () =>
      createWorkbookMutationCommandPorts({
        readScope: () => mutationRuntime.recordReadScope,
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
      readScope: () => mutationRuntime.recordReadScope,
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
          mutationRuntime.noteAssociations.suspend();
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
        createOrdinaryCreateTransport(
          apiBase,
          () => mutationRuntime.recordReadScope,
        ),
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
          readScope: () => mutationRuntime.recordReadScope,
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
      mutationRuntime.noteAssociations.configure(
        createNoteAssociationReader({
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void recheckMentionAuthority();
          },
        }),
        currentAuthorityReader,
        createNoteAssociationTransport(
          apiBase,
          () => mutationRuntime.recordReadScope,
        ),
      ),
    [
      mutationRuntime,
      apiBase,
      incidentId,
      currentAuthorityReader,
      recheckMentionAuthority,
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
        createNoteCreateTransport(
          apiBase,
          () => mutationRuntime.recordReadScope,
        ),
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
          readScope: () => mutationRuntime.recordReadScope,
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void recheckMentionAuthority();
          },
        }),
        currentAuthorityReader,
        createCoordinationCreateTransport(
          apiBase,
          () => mutationRuntime.recordReadScope,
        ),
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
        createWorkbookAuthoringReader({
          readScope: () => mutationRuntime.recordReadScope,
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void mutationRuntime.contextualCreate.recheckAuthority();
          },
        }),
        currentAuthorityReader,
        createContextualCreateTransport(
          apiBase,
          () => mutationRuntime.recordReadScope,
        ),
      ),
    [apiBase, incidentId, mutationRuntime, currentAuthorityReader],
  );
  useMemo(() => {
    const reader = createWorkbookAuthoringReader({
      readScope: () => mutationRuntime.recordReadScope,
      apiBase,
      incidentId,
      recheckAuthority: () => {
        void mutationRuntime.evidenceAttachments.recheckAuthority();
        void mutationRuntime.timelineFiles.recheckAuthority();
      },
    });
    const transport = createEvidenceFileTransport(
      apiBase,
      () => mutationRuntime.recordReadScope,
    );
    mutationRuntime.evidenceAttachments.configure(
      reader,
      currentAuthorityReader,
      transport,
    );
    mutationRuntime.timelineFiles.configure(
      reader,
      currentAuthorityReader,
      transport,
      createTimelineFileLinkTransport(
        apiBase,
        () => mutationRuntime.recordReadScope,
      ),
    );
  }, [apiBase, incidentId, mutationRuntime, currentAuthorityReader]);
  useMemo(
    () =>
      mutationRuntime.timelineRelatedEvidence.configure(
        createWorkbookAuthoringReader({
          readScope: () => mutationRuntime.recordReadScope,
          apiBase,
          incidentId,
          recheckAuthority: () => {
            void mutationRuntime.timelineRelatedEvidence.recheckAuthority();
          },
        }),
        currentAuthorityReader,
        createTimelineRelatedEvidenceTransport(
          apiBase,
          () => mutationRuntime.recordReadScope,
        ),
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
        createPartyCreationTransport(
          apiBase,
          () => mutationRuntime.recordReadScope,
        ),
        createPartyLinkReader({
          readScope: () => mutationRuntime.recordReadScope,
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
  // Rebind reads only when accepted read authority changes. Rechecking the
  // same authority must not abort an acknowledged operation's refresh.
  const readScopeKey = useSyncExternalStore(mutationRuntime.subscribe, () =>
    JSON.stringify(mutationRuntime.recordReadScope),
  );
  const viewQuery = useMemo(
    () =>
      createWorkbookViewQueryAdapter({
        apiBase,
        incidentId,
        readScope: () => {
          const scope = mutationRuntime.recordReadScope;
          return acceptedAuthority.role &&
            scope?.actorId === acceptedAuthority.userId &&
            scope?.sessionIdentity === sessionIdentity &&
            JSON.stringify(scope) === readScopeKey
            ? scope
            : null;
        },
      }),
    [
      apiBase,
      incidentId,
      mutationRuntime,
      acceptedAuthority.role,
      acceptedAuthority.userId,
      sessionIdentity,
      readScopeKey,
    ],
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
