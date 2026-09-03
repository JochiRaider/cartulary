import type { ViewContract } from "@cartulary/view-contracts";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import {
  type WorkbookInspectorLiveSubject,
  workbookInspectorSubjectsEqual,
} from "../../inspector/workbookInspectorSubject";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "./timelineRowModel";

export type TimelineRelatedActionContext = {
  readonly authorized: boolean;
  readonly surfaceKey: string;
};

export type TimelineRelatedWorkflowIdentity = {
  readonly subject: WorkbookInspectorLiveSubject;
  readonly surfaceKey: string;
  readonly workflowId: symbol;
};

export type TimelineRelatedSubmissionPlan =
  | {
      readonly kind: "dispatch";
      readonly contract: ViewContract;
      readonly draft: Readonly<Record<string, string>>;
      readonly evidenceLinkRequired: boolean;
      readonly featureGroupKey: string;
      readonly identity: TimelineRelatedWorkflowIdentity;
      readonly sourceRow: WorkbookRow;
    }
  | {
      readonly kind: "reject";
      readonly reason:
        | "authorization_lost"
        | "surface_changed"
        | "workflow_unavailable"
        | "capability_unavailable"
        | "selection_changed"
        | "source_invalid";
    };

export function timelineRelatedWorkflowIdentity(
  workflow: InspectorRelatedRecordWorkflowState,
  surfaceKey: string,
): TimelineRelatedWorkflowIdentity {
  return {
    subject: workflow.subject,
    surfaceKey,
    workflowId: workflow.workflowId,
  };
}

export function planTimelineRelatedSubmission(options: {
  readonly context: TimelineRelatedActionContext;
  readonly evidenceViewSchemaId: string;
  readonly identity: TimelineRelatedWorkflowIdentity;
  readonly selectedRow: WorkbookRow | null;
  readonly selectedSubject: WorkbookInspectorLiveSubject | null;
  readonly targetContracts: ReadonlyMap<string, ViewContract>;
  readonly workflow: InspectorRelatedRecordWorkflowState | null;
}): TimelineRelatedSubmissionPlan {
  if (!options.context.authorized) {
    return { kind: "reject", reason: "authorization_lost" };
  }
  if (options.context.surfaceKey !== options.identity.surfaceKey) {
    return { kind: "reject", reason: "surface_changed" };
  }
  const workflow = options.workflow;
  if (
    workflow === null ||
    workflow.phase !== "editing" ||
    workflow.workflowId !== options.identity.workflowId
  ) {
    return { kind: "reject", reason: "workflow_unavailable" };
  }
  const route = workflow.featureGroup.routeBinding;
  const contract = options.targetContracts.get(
    workflow.targetContract.viewSchemaId,
  );
  if (
    route.kind !== "view_row_create" ||
    route.owner !== "view_row_create_route" ||
    route.targetViewSchemaId !== workflow.targetContract.viewSchemaId ||
    contract === undefined ||
    !contract.createCapable
  ) {
    return { kind: "reject", reason: "capability_unavailable" };
  }
  if (
    !workbookInspectorSubjectsEqual(
      options.identity.subject,
      options.selectedSubject,
    ) ||
    !workbookInspectorSubjectsEqual(workflow.subject, options.selectedSubject)
  ) {
    return { kind: "reject", reason: "selection_changed" };
  }
  const row = options.selectedRow;
  if (
    row === null ||
    row.viewSchemaId !== timelineViewSchemaId ||
    row.recordId !== workflow.subject.recordId ||
    row.rowVersion !== workflow.subject.rowVersion ||
    row.pendingSignature !== null
  ) {
    return { kind: "reject", reason: "source_invalid" };
  }
  return {
    contract,
    draft: workflow.draft,
    evidenceLinkRequired:
      contract.viewSchemaId === options.evidenceViewSchemaId,
    featureGroupKey: workflow.featureGroup.featureGroupKey,
    identity: options.identity,
    kind: "dispatch",
    sourceRow: row,
  };
}

export function timelineRelatedWorkflowIsCurrent(options: {
  readonly context: TimelineRelatedActionContext;
  readonly identity: TimelineRelatedWorkflowIdentity;
  readonly selectedRow: WorkbookRow | null;
  readonly selectedSubject: WorkbookInspectorLiveSubject | null;
  readonly workflow: InspectorRelatedRecordWorkflowState | null;
}): boolean {
  const row = options.selectedRow;
  return (
    options.context.authorized &&
    options.context.surfaceKey === options.identity.surfaceKey &&
    options.workflow?.workflowId === options.identity.workflowId &&
    workbookInspectorSubjectsEqual(
      options.identity.subject,
      options.selectedSubject,
    ) &&
    row !== null &&
    row.viewSchemaId === timelineViewSchemaId &&
    row.recordId === options.identity.subject.recordId &&
    row.rowVersion === options.identity.subject.rowVersion &&
    row.pendingSignature === null
  );
}
