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
import { createWorkbookDecisionSupersessionAdapter } from "../adapters/createWorkbookDecisionSupersessionAdapter";
import { createWorkbookEntityMergeAdapter } from "../adapters/createWorkbookEntityMergeAdapter";
import { createWorkbookIncidentAdapter } from "../adapters/createWorkbookIncidentAdapter";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { readWorkbookAuthoringRecord } from "../adapters/readWorkbookAuthoringRecord";
import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import { createRecordPatchTransport } from "../adapters/workbookRecordPatchTransport";
import { createWorkbookMutationCommandPorts } from "../mutations/createWorkbookMutationCommandPorts";
import { createBrowserSecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import { createWorkbookMutationRuntime } from "../runtime/createWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
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

/** Build retained dispatch capabilities once, before exposing the runtime to presentation. */
export function createWorkbookMutationInfrastructure({
  apiBase,
  incidentId,
  clientInstanceId,
  partyAuthorization,
}: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly clientInstanceId: string;
  readonly partyAuthorization: AuthorizationRecoveryPort;
}): WorkbookMutationRuntime {
  const transactionIds = createBrowserSecureTransactionIdPort();
  const pendingMutationPort = createWorkbookPendingMutationAdapter({
    apiBase,
    incidentId,
    readScope: () => mutationRuntime.recordReadScope,
    recordTiming: (name, details = {}) => {
      if (typeof performance !== "undefined")
        performance.mark(`cartulary.workbook.${name}`, { detail: details });
    },
  });
  const mutationRuntime = createWorkbookMutationRuntime(
    { clientInstanceId, incidentId },
    transactionIds,
    pendingMutationPort,
  );
  const onAuthorityUncertain = () =>
    mutationRuntime.notifyPresentationAuthorityUncertain();
  const recheckMentionAuthority = async () => onAuthorityUncertain();
  const authorizationRecovered = (
    result: Parameters<
      WorkbookMutationRuntime["notifyPresentationAuthorizationRecovered"]
    >[0],
  ) => mutationRuntime.notifyPresentationAuthorizationRecovered(result);
  mutationRuntime.indicatorLifecycle.configure(
    createIndicatorLifecycleAdapter({
      apiBase,
      incidentId,
      readScope: () => mutationRuntime.recordReadScope,
    }),
    onAuthorityUncertain,
  );
  mutationRuntime.indicatorObservations.configure(
    createObservationReader({
      apiBase,
      incidentId,
      readScope: () => mutationRuntime.recordReadScope,
    }),
    createObservationTransport({ apiBase, incidentId }),
    onAuthorityUncertain,
  );
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
  );
  const timelineMentions = timelineMentionOwnerFor(mutationRuntime);
  {
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
  }
  const timelineCapture = timelineCaptureOwnerFor(mutationRuntime);
  timelineCapture.configure(
    createTimelineRecordActionAdapter({
      apiBase,
      readScope: () => mutationRuntime.recordReadScope,
    }),
    createTimelineCandidateReader({ apiBase, incidentId }),
    onAuthorityUncertain,
  );
  mutationRuntime.decisionSupersession.configure(
    createWorkbookDecisionSupersessionAdapter({
      apiBase,
      incidentId,
      readScope: () => mutationRuntime.recordReadScope,
    }),
    onAuthorityUncertain,
  );
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
  );
  mutationRuntime.batches.configure(
    createWorkbookBatchTransport({
      apiBase,
      incidentId,
      readScope: () => mutationRuntime.recordReadScope,
    }),
  );
  const mutationCommands = createWorkbookMutationCommandPorts({
    readScope: () => mutationRuntime.recordReadScope,
    batches: mutationRuntime.batches,
    apiBase,
    incidentId,
    transactionIds,
  });
  mutationRuntime.history.configure(
    mutationCommands.records,
    onAuthorityUncertain,
  );
  mutationRuntime.entityMerge.configure(
    createWorkbookEntityMergeAdapter({ apiBase, incidentId }),
    onAuthorityUncertain,
  );
  {
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
  }
  const incidentPort = createWorkbookIncidentAdapter({ apiBase, incidentId });
  const currentAuthorityReader = (
    () => async (baseline: WorkbookMutationAuthority, signal: AbortSignal) => {
      const result = await partyAuthorization.recover({
        incidentId,
        signal,
      });
      if (signal.aborted) throw new Error("Authority read interrupted.");
      if (result.kind !== "authorized") {
        if (result.kind === "access_lost" || result.kind === "session_lost") {
          mutationRuntime.applyAuthorizationRecoveryState("paused");
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
    }
  )();
  mutationRuntime.ordinaryCreate.configure(
    currentAuthorityReader,
    createOrdinaryCreateTransport(
      apiBase,
      () => mutationRuntime.recordReadScope,
    ),
    () => {
      void recheckMentionAuthority();
    },
  );
  mutationRuntime.ordinaryCreate.configureReader(
    createWorkbookAuthoringReader({
      readScope: () => mutationRuntime.recordReadScope,
      apiBase,
      incidentId,
      recheckAuthority: () => {
        void recheckMentionAuthority();
      },
    }),
  );
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
  );
  mutationRuntime.noteCreate.configure(
    createNoteCreateReader({
      apiBase,
      incidentId,
      recheckAuthority: () => {
        void recheckMentionAuthority();
      },
    }),
    currentAuthorityReader,
    createNoteCreateTransport(apiBase, () => mutationRuntime.recordReadScope),
  );
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
  );
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
  );
  {
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
  }
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
  );
  mutationRuntime.assessmentAuthoring.configure(
    mutationCommands.assessment,
    currentAuthorityReader,
  );
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
  );
  return mutationRuntime;
}
