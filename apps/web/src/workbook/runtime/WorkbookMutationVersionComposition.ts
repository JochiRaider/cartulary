import { assessmentsViewSchemaId } from "@cartulary/view-contracts";
import { decisionViewId } from "../features/coordination/decisionSupersessionModel";
import { indicatorLifecycleViewId } from "../features/indicators/indicatorLifecycleModel";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationAccepted } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookTimelineActionRuntimePort } from "../ports/WorkbookTimelineActionRuntimePort";
import type { WorkbookMutationFeatures } from "./WorkbookMutationFeatureAssembly";

export const entityViewSchemas: ReadonlySet<string> = new Set([
  hostsViewSchemaId,
  identitiesViewSchemaId,
]);

/** Fixed cross-owner version wiring; source owners still decide admission. */
export function createWorkbookMutationVersionComposition(
  features: WorkbookMutationFeatures,
  timeline: {
    readonly actions: () => WorkbookTimelineActionRuntimePort | null;
    readonly mentions: () => WorkbookTimelineActionRuntimePort | null;
  },
) {
  const {
    assessmentAuthoring,
    batches,
    contextualCreate,
    coordinationCreate,
    decisionSupersession,
    entityMerge,
    evidenceAttachments,
    explicitPatches,
    history,
    indicatorLifecycle,
    noteAssociations,
    noteCreate,
    ordinaryCreate,
    timelineFiles,
    timelineRelatedEvidence,
  } = features;

  const acceptEntityVersion = (recordId: string, version: number) => {
    noteCreate.observe(recordId, version);
    noteAssociations.observe(recordId, version);
    ordinaryCreate.observe(recordId, version);
    coordinationCreate.observe(recordId, version);
    contextualCreate.observe(recordId, version);
    entityMerge.acceptVersion(recordId, version);
    history.acceptVersion(recordId, version);
  };

  return {
    acceptEntityVersion,
    historyChanged: () => {
      for (const entry of history.getSnapshot()) {
        const receipt = entry.receipt;
        if (receipt) {
          timelineFiles.acceptVersion(receipt.recordId, receipt.rowVersion);
          evidenceAttachments.observe(receipt.recordId, receipt.rowVersion);
        }
        if (receipt) noteCreate.observe(receipt.recordId, receipt.rowVersion);
        if (receipt)
          ordinaryCreate.observe(receipt.recordId, receipt.rowVersion);
        if (receipt)
          coordinationCreate.observe(receipt.recordId, receipt.rowVersion);
        if (receipt)
          contextualCreate.observe(receipt.recordId, receipt.rowVersion);
        if (receipt)
          timelineRelatedEvidence.observe(receipt.recordId, receipt.rowVersion);
        if (
          receipt &&
          entityViewSchemas.has(entry.attempt.subject.viewSchemaId)
        )
          entityMerge.acceptVersion(receipt.recordId, receipt.rowVersion);
        if (receipt)
          explicitPatches.acceptVersion(receipt.recordId, receipt.rowVersion);
        if (
          receipt &&
          entry.attempt.subject.viewSchemaId === indicatorLifecycleViewId
        )
          indicatorLifecycle.acceptVersion(
            receipt.recordId,
            receipt.rowVersion,
          );
        if (
          receipt &&
          entry.attempt.subject.viewSchemaId === assessmentsViewSchemaId
        )
          assessmentAuthoring.acceptVersion(
            receipt.recordId,
            receipt.rowVersion,
            receipt.kind === "delete" && receipt.deleted,
          );
        if (receipt && entry.attempt.subject.viewSchemaId === decisionViewId)
          decisionSupersession.acceptVersion(
            receipt.recordId,
            receipt.rowVersion,
          );
      }
    },
    decisionChanged: () => {
      for (const entry of decisionSupersession.getSnapshot().entries)
        if (entry.receipt) {
          history.acceptVersion(
            entry.receipt.target_record_id,
            entry.receipt.target_row_version,
          );
          history.acceptVersion(
            entry.receipt.superseding_record_id,
            entry.receipt.superseding_row_version,
          );
        }
    },
    observeTimelineVersion: (recordId: string, version: number) => {
      noteCreate.observe(recordId, version);
      noteAssociations.observe(recordId, version);
      ordinaryCreate.observe(recordId, version);
      coordinationCreate.observe(recordId, version);
      contextualCreate.observe(recordId, version);
      timelineRelatedEvidence.observe(recordId, version);
      timelineFiles.acceptVersion(recordId, version);
      history.acceptVersion(recordId, version);
      timeline.actions()?.acceptVersion(recordId, version);
      timeline.mentions()?.acceptVersion(recordId, version);
    },
    acceptedPendingMutation: (
      unitId: string,
      accepted: WorkbookPendingMutationAccepted,
    ) => {
      const { row, viewSchemaId } = accepted;
      timelineFiles.acceptVersion(row.record_id, row.row_version);
      evidenceAttachments.observe(row.record_id, row.row_version);
      batches.acceptPrerequisiteRow(unitId, row.record_id, row.row_version);
      timelineRelatedEvidence.observe(row.record_id, row.row_version);
      noteAssociations.observe(row.record_id, row.row_version);
      noteCreate.observe(row.record_id, row.row_version);
      ordinaryCreate.observe(row.record_id, row.row_version);
      coordinationCreate.observe(row.record_id, row.row_version);
      contextualCreate.observe(row.record_id, row.row_version);
      explicitPatches.observeReceipt(accepted);
      if (entityViewSchemas.has(viewSchemaId))
        acceptEntityVersion(row.record_id, row.row_version);
      if (viewSchemaId === decisionViewId) decisionSupersession.acceptRow(row);
      if (viewSchemaId === indicatorLifecycleViewId)
        indicatorLifecycle.acceptRow(row);
    },
  };
}

export type WorkbookMutationVersionComposition = ReturnType<
  typeof createWorkbookMutationVersionComposition
>;
