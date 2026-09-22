import { assessmentsViewSchemaId } from "@cartulary/view-contracts";
import { WorkbookAssessmentAuthoringOwner } from "../features/assessments/WorkbookAssessmentAuthoringOwner";
import type { DecisionSupersessionReview } from "../features/coordination/decisionSupersessionModel";
import { taskExplicitPatchContribution } from "../features/coordination/taskExplicitPatchContribution";
import {
  type TaskLifecycleDraftStore,
  taskViewId,
} from "../features/coordination/taskLifecycleModel";
import { WorkbookContextualTaskDecisionCreateOwner } from "../features/coordination/WorkbookContextualTaskDecisionCreateOwner";
import { WorkbookCoordinationCreateOwner } from "../features/coordination/WorkbookCoordinationCreateOwner";
import { WorkbookDecisionSupersessionOwner } from "../features/coordination/WorkbookDecisionSupersessionOwner";
import type { EntityMergeReview } from "../features/entities/entityMergeReview";
import { WorkbookEntityMergeOwner } from "../features/entities/WorkbookEntityMergeOwner";
import { WorkbookEvidenceAttachmentOwner } from "../features/evidence/WorkbookEvidenceAttachmentOwner";
import { WorkbookTimelineFileOwner } from "../features/evidence/WorkbookTimelineFileOwner";
import { WorkbookTimelineRelatedEvidenceOwner } from "../features/evidence/WorkbookTimelineRelatedEvidenceOwner";
import type { LifecycleDraft } from "../features/indicators/indicatorLifecycleModel";
import { WorkbookIndicatorCreateOwner } from "../features/indicators/WorkbookIndicatorCreateOwner";
import { WorkbookIndicatorLifecycleOwner } from "../features/indicators/WorkbookIndicatorLifecycleOwner";
import { WorkbookObservationOwner } from "../features/indicators/WorkbookObservationOwner";
import { WorkbookNoteAssociationOwner } from "../features/notes/WorkbookNoteAssociationOwner";
import { WorkbookNoteCreateOwner } from "../features/notes/WorkbookNoteCreateOwner";
import { createOrdinaryCreateContributions } from "../features/ordinary/ordinaryCreateContributions";
import { WorkbookOrdinaryCreateOwner } from "../features/ordinary/WorkbookOrdinaryCreateOwner";
import { WorkbookPartyLinkOperationOwner } from "../features/parties/WorkbookPartyLinkOperationOwner";
import { WorkbookRecordHistoryOwner } from "../history/WorkbookRecordHistoryOwner";
import type { EntityRecordWriteTarget } from "../mutations/entityRecordWriteBoundary";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookSourceWriteSettlement } from "../ports/WorkbookSourceWriteCoordination";
import type { WorkbookTimelineActionRuntimePort } from "../ports/WorkbookTimelineActionRuntimePort";
import type { WorkbookReadScope } from "../query/WorkbookQueryRow";
import type { PendingReplayScope } from "./pending/workbookPendingQueue";
import { WorkbookBatchOperationOwner } from "./WorkbookBatchOperationOwner";
import type {
  WorkbookConflictRegistration,
  WorkbookConflictStore,
} from "./WorkbookConflictStore";
import { WorkbookExplicitPatchOwner } from "./WorkbookExplicitPatchOwner";
import type { WorkbookMutationDriverRegistry } from "./WorkbookMutationDriverRegistry";
import type { WorkbookSurfaceRegistry } from "./WorkbookSurfaceRegistry";
import type { WorkbookPendingQueueRuntime } from "./workbookPendingReplayRuntime";

export type WorkbookMutationFeatures = {
  readonly evidenceAttachments: WorkbookEvidenceAttachmentOwner;
  readonly timelineFiles: WorkbookTimelineFileOwner;
  readonly timelineRelatedEvidence: WorkbookTimelineRelatedEvidenceOwner;
  readonly ordinaryCreate: WorkbookOrdinaryCreateOwner;
  readonly noteAssociations: WorkbookNoteAssociationOwner;
  readonly noteCreate: WorkbookNoteCreateOwner;
  readonly coordinationCreate: WorkbookCoordinationCreateOwner;
  readonly contextualCreate: WorkbookContextualTaskDecisionCreateOwner;
  readonly assessmentAuthoring: WorkbookAssessmentAuthoringOwner;
  readonly explicitPatches: WorkbookExplicitPatchOwner;
  readonly partyLinks: WorkbookPartyLinkOperationOwner;
  readonly entityMerge: WorkbookEntityMergeOwner;
  readonly decisionSupersession: WorkbookDecisionSupersessionOwner;
  readonly indicatorCreate: WorkbookIndicatorCreateOwner;
  readonly indicatorObservations: WorkbookObservationOwner;
  readonly indicatorLifecycle: WorkbookIndicatorLifecycleOwner;
  readonly history: WorkbookRecordHistoryOwner;
  readonly batches: WorkbookBatchOperationOwner;
};

/** Only shared coordination capabilities cross the construction boundary. */
export type WorkbookMutationFeatureHost = {
  readonly scope: PendingReplayScope;
  readonly transactionIds: SecureTransactionIdPort;
  readonly retired: boolean;
  readonly recordReadScope: WorkbookReadScope | null;
  readonly pending: WorkbookPendingQueueRuntime;
  readonly conflicts: WorkbookConflictStore;
  readonly drivers: WorkbookMutationDriverRegistry;
  readonly surfaces: WorkbookSurfaceRegistry;
  readonly taskDrafts: TaskLifecycleDraftStore;
  readonly entityWrites: ReadonlyMap<symbol, EntityRecordWriteTarget>;
  readonly timelineActions: WorkbookTimelineActionRuntimePort | null;
  readonly timelineMentionOperations: WorkbookTimelineActionRuntimePort | null;
  coordinateSourceWrites(
    recordId: string,
    signal: AbortSignal,
    viewSchemaId: string,
    reservation?: {
      readonly noteAssociation?: boolean;
      readonly partyReservationId?: string;
      readonly explicitPatchId?: string;
      readonly fileOwner?: "evidence" | "timeline";
    },
  ): Promise<WorkbookSourceWriteSettlement>;
  beginEntityWrite(target: EntityRecordWriteTarget): (() => void) | null;
  reserveEntityWrite(target: EntityRecordWriteTarget): (() => void) | null;
  acceptEntityVersion(recordId: string, version: number): void;
  coordinateEntityMerge(
    review: EntityMergeReview,
    signal: AbortSignal,
  ): Promise<boolean>;
  coordinateDecisionSupersession(
    review: DecisionSupersessionReview,
    signal: AbortSignal,
  ): Promise<boolean>;
  coordinateIndicatorLifecycle(
    draft: LifecycleDraft,
    signal: AbortSignal,
  ): Promise<boolean>;
  timelineActionBlocksRecord(recordId: string): boolean;
  rememberClientTransaction(id: string): void;
  resolveSocketClientTxn(id: string): boolean;
  registerConflict(input: WorkbookConflictRegistration): void;
};

/** Fixed composition; no callbacks run until the completed feature set is returned. */
export function assembleWorkbookMutationFeatures(
  host: WorkbookMutationFeatureHost,
): WorkbookMutationFeatures {
  const evidenceAttachments: WorkbookEvidenceAttachmentOwner =
    new WorkbookEvidenceAttachmentOwner(
      host.scope.incidentId,
      host.transactionIds,
      {
        coordinate: (recordId, signal) =>
          host.coordinateSourceWrites(
            recordId,
            signal,
            "cartulary.view.evidence.v1",
            { fileOwner: "evidence" },
          ),
        accepted: (receipt, id) => {
          host.rememberClientTransaction(id);
          history.acceptVersion(
            receipt.data.row.record_id,
            receipt.data.row.row_version,
          );
          host.surfaces.invalidate("cartulary.view.evidence.v1");
        },
        refresh: () =>
          host.surfaces.refreshIfMounted("cartulary.view.evidence.v1"),
      },
    );
  const timelineFiles: WorkbookTimelineFileOwner =
    new WorkbookTimelineFileOwner(host.scope.incidentId, host.transactionIds, {
      coordinate: (recordId, signal) =>
        host.coordinateSourceWrites(
          recordId,
          signal,
          "cartulary.view.timeline.v2",
          { fileOwner: "timeline" },
        ),
      accepted: (receipt, id) => {
        host.rememberClientTransaction(id);
        history.acceptVersion(
          receipt.data.row.record_id,
          receipt.data.row.row_version,
        );
        host.surfaces.invalidate(receipt.data.view_schema_id);
      },
      refresh: async (views) => {
        await Promise.all(
          views.map((view) => host.surfaces.refreshIfMounted(view)),
        );
      },
    });
  const timelineRelatedEvidence: WorkbookTimelineRelatedEvidenceOwner =
    new WorkbookTimelineRelatedEvidenceOwner(
      host.scope.incidentId,
      host.transactionIds,
      {
        coordinate: (recordId, signal) =>
          host.coordinateSourceWrites(
            recordId,
            signal,
            "cartulary.view.timeline.v2",
          ),
        accepted: (receipt, id) => {
          host.rememberClientTransaction(id);
          history.acceptVersion(
            receipt.data.row.record_id,
            receipt.data.row.row_version,
          );
        },
        refresh: async (views, recordIds) => {
          await Promise.all([
            ...views.map((view) => host.surfaces.refreshIfMounted(view)),
            ...recordIds.map(async (recordId) => {
              const observation = await history.loadProjection(recordId);
              if (observation.kind !== "accepted")
                throw new Error("Record history refresh is incomplete.");
              timelineRelatedEvidence.observe(
                recordId,
                observation.value.row_version,
              );
            }),
          ]);
        },
        conflict: (checkpoint, conflict) => {
          const draft = checkpoint.create.attempt.review.draft;
          host.registerConflict({
            conflict,
            focusOrigin: "inspector",
            sheetRef: draft.presentation.sheetRef,
            rowLabel: "Original Timeline record",
            surfaceLabel: "Timeline",
            viewSchemaId: draft.source.viewSchemaId,
          });
        },
      },
    );
  const ordinaryCreate: WorkbookOrdinaryCreateOwner =
    new WorkbookOrdinaryCreateOwner(
      host.scope.incidentId,
      createOrdinaryCreateContributions({
        begin: (target) => host.beginEntityWrite(target),
        acceptVersion: (id, version) => host.acceptEntityVersion(id, version),
      }),
      {
        ids: host.transactionIds,
        effects: {
          accepted: (receipt, id) => {
            host.rememberClientTransaction(id);
            history.acceptVersion(
              receipt.data.row.record_id,
              receipt.data.row.row_version,
            );
          },
          refresh: async (receipt) => {
            await Promise.all([
              host.surfaces.refreshRequired(receipt.data.view_schema_id),
              history.refreshRecordPresentation(receipt.data.row.record_id),
            ]);
          },
        },
      },
    );
  const noteAssociations: WorkbookNoteAssociationOwner =
    new WorkbookNoteAssociationOwner(
      host.scope.incidentId,
      host.transactionIds,
      {
        coordinate: (recordId, signal) =>
          host.coordinateSourceWrites(
            recordId,
            signal,
            "cartulary.view.notes.v1",
            { noteAssociation: true },
          ),
        accepted: (receipt, id) => {
          host.rememberClientTransaction(id);
          history.acceptVersion(
            receipt.data.row.record_id,
            receipt.data.row.row_version,
          );
        },
        refresh: async (recordId) => {
          await Promise.all([
            host.surfaces.refreshIfMounted("cartulary.view.notes.v1"),
            host.surfaces.refreshIfMounted("cartulary.view.evidence.v1"),
            history.refreshRecordPresentation(recordId),
          ]);
        },
      },
    );
  const noteCreate: WorkbookNoteCreateOwner = new WorkbookNoteCreateOwner(
    host.scope.incidentId,
    host.transactionIds,
    {
      coordinate: (source, signal) =>
        host.coordinateSourceWrites(
          source.recordId,
          signal,
          source.viewSchemaId,
        ),
      accepted: (receipt, id) => {
        host.rememberClientTransaction(id);
        history.acceptVersion(
          receipt.data.row.record_id,
          receipt.data.row.row_version,
        );
      },
      refresh: async (views, records) => {
        await Promise.all([
          ...views.map((view) => host.surfaces.refreshIfMounted(view)),
          ...records.map(async (recordId) => {
            const result = await history.loadProjection(recordId);
            if (result.kind !== "accepted")
              throw new Error("Record history refresh is incomplete.");
            noteCreate.observe(recordId, result.value.row_version);
            ordinaryCreate.observe(recordId, result.value.row_version);
            await history.refreshRecordPresentation(recordId);
          }),
        ]);
      },
    },
  );
  const coordinationCreate: WorkbookCoordinationCreateOwner =
    new WorkbookCoordinationCreateOwner(
      host.scope.incidentId,
      host.transactionIds,
      {
        coordinate: (source, signal) =>
          host.coordinateSourceWrites(
            source.recordId,
            signal,
            source.viewSchemaId,
          ),
        accepted: (receipt, id) => {
          host.rememberClientTransaction(id);
          history.acceptVersion(
            receipt.data.row.record_id,
            receipt.data.row.row_version,
          );
        },
        refresh: async (views, records) => {
          await Promise.all([
            ...views.map((view) => host.surfaces.refreshIfMounted(view)),
            ...records.map(async (recordId) => {
              const result = await history.loadProjection(recordId);
              if (result.kind !== "accepted")
                throw new Error("Record history refresh is incomplete.");
              coordinationCreate.observe(recordId, result.value.row_version);
              await history.refreshRecordPresentation(recordId);
            }),
          ]);
        },
      },
    );
  const contextualCreate: WorkbookContextualTaskDecisionCreateOwner =
    new WorkbookContextualTaskDecisionCreateOwner(
      host.scope.incidentId,
      host.transactionIds,
      {
        coordinate: (draft, signal) =>
          host.coordinateSourceWrites(
            draft.source.recordId,
            signal,
            draft.source.viewSchemaId,
          ),
        accepted: (receipt, id) => {
          host.rememberClientTransaction(id);
          history.acceptVersion(
            receipt.data.row.record_id,
            receipt.data.row.row_version,
          );
          if (receipt.data.view_schema_id === taskViewId)
            explicitPatches.acceptRow(receipt.data.row);
          else decisionSupersession.acceptRow(receipt.data.row);
        },
        observed: (view, row) => {
          if (view === taskViewId) explicitPatches.acceptRow(row);
          else if (view === "cartulary.view.decisions.v1")
            decisionSupersession.acceptRow(row);
        },
        refresh: async (_draft, views) => {
          await Promise.all(
            views.map((view) => host.surfaces.refreshIfMounted(view)),
          );
        },
      },
    );
  const assessmentAuthoring: WorkbookAssessmentAuthoringOwner =
    new WorkbookAssessmentAuthoringOwner(
      host.scope.incidentId,
      host.transactionIds,
      {
        accepted: (receipt, id) => {
          host.rememberClientTransaction(id);
          history.acceptVersion(
            receipt.data.row.record_id,
            receipt.data.row.row_version,
          );
        },
        refresh: () => host.surfaces.refreshRequired(assessmentsViewSchemaId),
      },
    );
  const explicitPatches: WorkbookExplicitPatchOwner =
    new WorkbookExplicitPatchOwner(host.scope.incidentId, host.transactionIds, {
      readScope: () => host.recordReadScope,
      coordinate: (recordId, signal, viewSchemaId, reservationId) =>
        host.coordinateSourceWrites(recordId, signal, viewSchemaId, {
          explicitPatchId: reservationId,
        }),
      contribute: (intent) =>
        taskExplicitPatchContribution(intent, host.taskDrafts),
      refresh: (view) => host.surfaces.refreshRequired(view),
      remember: (id) => host.rememberClientTransaction(id),
      settle: (id) => {
        host.resolveSocketClientTxn(id);
      },
      registerConflict: (input) => {
        host.registerConflict(input);
      },
      accepted: (row) => history.acceptVersion(row.record_id, row.row_version),
    });
  const partyLinks: WorkbookPartyLinkOperationOwner =
    new WorkbookPartyLinkOperationOwner(
      host.scope.incidentId,
      host.transactionIds,
      explicitPatches,
      {
        coordinate: (review, signal, reservationId) =>
          host.coordinateSourceWrites(
            review.source.record_id,
            signal,
            review.pair.viewSchemaId,
            { partyReservationId: reservationId },
          ),
        remember: (id) => host.rememberClientTransaction(id),
        settle: (id) => {
          host.resolveSocketClientTxn(id);
        },
        refresh: (view) => host.surfaces.refreshIfMounted(view),
      },
    );
  const entityMerge: WorkbookEntityMergeOwner = new WorkbookEntityMergeOwner(
    host.scope.incidentId,
    host.transactionIds,
    {
      canReserve: (review) =>
        !host.retired &&
        !batches.blocksEntityType(review.entityType) &&
        ![review.survivor.recordId, review.loser.recordId].some((id) =>
          explicitPatches.blocksRecord(id),
        ),
      coordinate: (review, signal) =>
        host.coordinateEntityMerge(review, signal),
    },
  );
  const decisionSupersession: WorkbookDecisionSupersessionOwner =
    new WorkbookDecisionSupersessionOwner(
      host.scope.incidentId,
      host.transactionIds,
      {
        canReserve: (review) =>
          !host.retired &&
          [review.target, review.replacement].every(
            (record) =>
              !explicitPatches.blocksRecord(record.recordId) &&
              (history.latestVersion(record.recordId) ?? 0) <=
                record.baseRowVersion,
          ),
        coordinate: (review, signal) =>
          host.coordinateDecisionSupersession(review, signal),
      },
    );
  const indicatorCreate: WorkbookIndicatorCreateOwner =
    new WorkbookIndicatorCreateOwner(
      host.scope.incidentId,
      host.transactionIds,
      (receipt, id) => {
        host.rememberClientTransaction(id);
        history.acceptVersion(receipt.row.record_id, receipt.row.row_version);
      },
    );
  const indicatorObservations: WorkbookObservationOwner =
    new WorkbookObservationOwner(
      host.scope.incidentId,
      host.transactionIds,
      (receipt, id) => {
        host.rememberClientTransaction(id);
        for (const record of receipt.affected_records)
          history.acceptVersion(record.record_id, record.row_version);
      },
    );
  const indicatorLifecycle: WorkbookIndicatorLifecycleOwner =
    new WorkbookIndicatorLifecycleOwner(
      host.scope.incidentId,
      host.transactionIds,
      {
        canReserve: (draft) =>
          !host.retired &&
          (history.latestVersion(draft.recordId) ?? 0) <=
            draft.baseRowVersion &&
          !history
            .getSnapshot()
            .some(
              (entry) =>
                entry.attempt.subject.recordId === draft.recordId &&
                entry.phase === "uncertain",
            ),
        coordinate: (draft, signal) =>
          host.coordinateIndicatorLifecycle(draft, signal),
        accepted: (receipt, clientTxnId) => {
          host.rememberClientTransaction(clientTxnId);
          for (const record of receipt.affected_records)
            history.acceptVersion(record.record_id, record.row_version);
        },
      },
    );
  const history: WorkbookRecordHistoryOwner = new WorkbookRecordHistoryOwner(
    host.scope.incidentId,
    host.transactionIds,
    undefined,
    (recordId) =>
      !entityMerge.blocksRecord(recordId) &&
      !decisionSupersession.blocksRecord(recordId) &&
      !indicatorLifecycle.blocksRecord(recordId) &&
      !explicitPatches.blocksRecord(recordId) &&
      !partyLinks.blocksRecord(recordId) &&
      !evidenceAttachments.blocksRecord(recordId) &&
      !timelineFiles.blocksRecord(recordId) &&
      !timelineRelatedEvidence.blocksRecord(recordId) &&
      !host.timelineActions?.blocksRecord(recordId) &&
      !host.timelineMentionOperations?.blocksRecord(recordId),
  );
  const batches: WorkbookBatchOperationOwner = new WorkbookBatchOperationOwner(
    host.scope.incidentId,
    host.transactionIds,
    {
      pending: () => host.pending.model.snapshot().units,
      sealPending: () => host.pending.model.sealPending(),
      available: (plan) =>
        !host.retired &&
        !host.pending.model.snapshot().authPaused &&
        !host.conflicts
          .entries()
          .some((entry) => plan.recordIds.includes(entry.conflict.record_id)) &&
        !plan.recordIds.some(
          (id) =>
            entityMerge.blocksRecord(id) ||
            explicitPatches.blocksRecord(id) ||
            decisionSupersession.blocksRecord(id) ||
            partyLinks.blocksRecord(id) ||
            host.timelineActionBlocksRecord(id),
        ) &&
        !(
          plan.entityType &&
          (entityMerge.blocksEntityType(plan.entityType) ||
            [...host.entityWrites.values()].some(
              (write) =>
                write.unknownEntityType === plan.entityType ||
                write.entityType === plan.entityType ||
                (!write.entityType && write.recordIds.length > 0),
            ))
        ),
      reserve: (plan) =>
        plan.entityType
          ? host.reserveEntityWrite({
              recordIds: plan.recordIds,
              unknownEntityType: plan.entityType,
            })
          : () => {},
      conflicts: (id) =>
        host.conflicts.entries().some((entry) => entry.batchOperationId === id),
      captured: (attempt) => host.drivers.captureBatchSources(attempt),
      accepted: (receipt, attempt) => {
        host.rememberClientTransaction(attempt.id);
        host.drivers.acceptBatchPredecessor(receipt, attempt);
        host.surfaces.invalidate(receipt.viewSchemaId);
        for (const row of receipt.rows) {
          history.acceptVersion(row.record_id, row.row_version);
          if (attempt.plan.entityType)
            host.acceptEntityVersion(row.record_id, row.row_version);
        }
        for (const conflict of receipt.conflicts)
          host.registerConflict({
            batchOperationId: attempt.id,
            conflict,
            focusOrigin: "grid",
            rowLabel: "Affected row",
            surfaceLabel: "Timeline",
            viewSchemaId: receipt.viewSchemaId,
          });
        if (batches.getSnapshot().authority)
          host.surfaces.applyBatch(receipt.viewSchemaId, receipt.rows);
      },
      refresh: (receipt) => host.surfaces.refreshRequired(receipt.viewSchemaId),
    },
  );
  return {
    evidenceAttachments,
    timelineFiles,
    timelineRelatedEvidence,
    ordinaryCreate,
    noteAssociations,
    noteCreate,
    coordinationCreate,
    contextualCreate,
    assessmentAuthoring,
    explicitPatches,
    partyLinks,
    entityMerge,
    decisionSupersession,
    indicatorCreate,
    indicatorObservations,
    indicatorLifecycle,
    history,
    batches,
  };
}
