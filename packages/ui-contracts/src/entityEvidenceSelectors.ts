import type { StableTestId } from "./selectorCore";
import {
  recordTestId,
  requireFieldKey,
  requireItemRef,
  requireRecordId,
  semanticSelectorTestId,
  tokenScopedTestId,
} from "./selectorCore";

import { viewScopedTestId } from "./viewSchemaSelectors";

export type EntityType = "host" | "identity";

export type EntityMergeControl =
  | "confirm"
  | "loser-record"
  | "message"
  | "plan"
  | "reason"
  | "start";

export type AssessmentCreateControl =
  | "assessed-at"
  | "confidence-band"
  | "message"
  | "rationale"
  | "state"
  | "subject"
  | "subject-type"
  | "submit"
  | "support-refs";

export type GenericWorkbookSelector =
  | "mutation-error"
  | "note-source-record"
  | "reference-load-error";

export type CoordinationWorkflowSelector =
  | "decision-reason"
  | "decision-replacement"
  | "decision-submit"
  | "decision-target"
  | "party-clear-both"
  | "party-clear-link"
  | "party-clear-text"
  | "party-create-from-text"
  | "party-existing"
  | "party-link-existing"
  | "party-pair"
  | "party-partial-completion"
  | "party-retry-created-link"
  | "task-blocked-reason"
  | "task-status"
  | "task-submit"
  | "task-target";

export const entityTypes = [
  "host",
  "identity",
] as const satisfies readonly EntityType[];

const entityMergeControlTestIds = Object.freeze({
  confirm: "merge-confirm",
  "loser-record": "merge-loser-record",
  message: "merge-message",
  plan: "merge-plan",
  reason: "merge-reason",
  start: "merge-start",
} satisfies Record<EntityMergeControl, string>);

const assessmentCreateControlTestIds = Object.freeze({
  "assessed-at": "assessment-create-assessed-at",
  "confidence-band": "assessment-create-confidence-band",
  message: "assessment-create-message",
  rationale: "assessment-create-rationale",
  state: "assessment-create-state",
  subject: "assessment-create-subject",
  "subject-type": "assessment-create-subject-type",
  submit: "assessment-create-submit",
  "support-refs": "assessment-create-support-refs",
} satisfies Record<AssessmentCreateControl, string>);

const genericWorkbookTestIds = Object.freeze({
  "mutation-error": "generic-mutation-error",
  "note-source-record": "generic-create-note-source-record",
  "reference-load-error": "generic-reference-load-error",
} satisfies Record<GenericWorkbookSelector, string>);

const coordinationWorkflowTestIds = Object.freeze({
  "decision-reason": "decision-supersede-reason",
  "decision-replacement": "decision-supersede-replacement",
  "decision-submit": "decision-supersede-submit",
  "decision-target": "decision-supersede-target",
  "party-clear-both": "party-link-clear-both",
  "party-clear-link": "party-link-clear-link",
  "party-clear-text": "party-link-clear-text",
  "party-create-from-text": "party-link-create-from-text",
  "party-existing": "party-link-existing-party",
  "party-link-existing": "party-link-link-existing",
  "party-pair": "party-link-pair",
  "party-partial-completion": "party-link-partial-completion",
  "party-retry-created-link": "party-link-retry-created",
  "task-blocked-reason": "task-lifecycle-blocked-reason",
  "task-status": "task-lifecycle-status",
  "task-submit": "task-lifecycle-submit",
  "task-target": "task-lifecycle-target",
} satisfies Record<CoordinationWorkflowSelector, string>);

export function entityMergeControlTestId(
  control: EntityMergeControl,
): StableTestId {
  return semanticSelectorTestId(
    entityMergeControlTestIds,
    control,
    "entity merge control",
  );
}

export function assessmentCreateControlTestId(
  control: AssessmentCreateControl,
): StableTestId {
  return semanticSelectorTestId(
    assessmentCreateControlTestIds,
    control,
    "assessment create control",
  );
}

export function genericWorkbookTestId(
  selector: GenericWorkbookSelector,
): StableTestId {
  return semanticSelectorTestId(
    genericWorkbookTestIds,
    selector,
    "generic workbook selector",
  );
}

export function coordinationWorkflowTestId(
  selector: CoordinationWorkflowSelector,
): StableTestId {
  return semanticSelectorTestId(
    coordinationWorkflowTestIds,
    selector,
    "coordination workflow selector",
  );
}

export function entityInspectButtonTestId(
  entityType: EntityType,
  recordId: string,
): string {
  return `inspect-${requireEntityType(entityType)}-${requireRecordId(recordId)}`;
}

export function entityInspectorTestId(entityType: EntityType): string {
  return `${requireEntityType(entityType)}-inspector`;
}

export function entityReusableIdentifiersSectionTestId(
  entityType: EntityType,
  recordId: string,
): string {
  return recordTestId(
    `${requireEntityType(entityType)}-reusable-identifiers`,
    recordId,
  );
}

export function entityReusableIdentifierItemTestId(
  entityType: EntityType,
  recordId: string,
  itemRef: string,
): string {
  return tokenScopedTestId(
    entityReusableIdentifiersSectionTestId(entityType, recordId),
    requireItemRef(itemRef),
  );
}

export function entityMergePreconditionDetailsTestId(
  entityType: EntityType,
  recordId: string,
): string {
  return recordTestId(
    `${requireEntityType(entityType)}-merge-precondition-details`,
    recordId,
  );
}

export function assessmentCreatePanelTestId(): string {
  return "assessment-create-panel";
}

export function evidencePreviewButtonTestId(recordId: string): string {
  return evidenceRecordControlTestId("preview", recordId);
}

export function evidenceDownloadButtonTestId(recordId: string): string {
  return evidenceRecordControlTestId("download", recordId);
}

export function evidenceAttachFileInputTestId(recordId: string): string {
  return evidenceRecordControlTestId("attach-file", recordId);
}

export function evidenceAccessMessageTestId(recordId: string): string {
  return evidenceRecordControlTestId("access-message", recordId);
}

export function evidencePreviewFrameTestId(recordId: string): string {
  return evidenceRecordControlTestId("preview-frame", recordId);
}

export function evidencePreviewPanelTestId(): string {
  return "evidence-preview-panel";
}

export function genericCreateFieldTestId(fieldKey: string): string {
  return `generic-create-field-${requireFieldKey(fieldKey)}`;
}

export function genericCreateSubmitTestId(viewSchemaId: string): string {
  return viewScopedTestId("generic-create-submit", viewSchemaId);
}

export function genericEditRecordSelectTestId(viewSchemaId: string): string {
  return genericEditControlTestId("record", viewSchemaId);
}

export function genericEditFieldSelectTestId(viewSchemaId: string): string {
  return genericEditControlTestId("field", viewSchemaId);
}

export function genericEditActionSelectTestId(viewSchemaId: string): string {
  return genericEditControlTestId("action", viewSchemaId);
}

export function genericEditValueTestId(viewSchemaId: string): string {
  return genericEditControlTestId("value", viewSchemaId);
}

export function genericEditSubmitTestId(viewSchemaId: string): string {
  return genericEditControlTestId("submit", viewSchemaId);
}

export function mentionResolveTargetSelectTestId(): string {
  return "inspector-resolve-target";
}

export function mentionResolveExistingButtonTestId(): string {
  return "inspector-resolve-existing";
}

export function mentionCreateEntityButtonTestId(
  entityType: EntityType,
): string {
  return `inspector-create-${requireEntityType(entityType)}`;
}

export function mentionDismissButtonTestId(): string {
  return "inspector-dismiss-mention";
}

export function mentionRestoreUnresolvedButtonTestId(): string {
  return "inspector-restore-unresolved";
}

function genericEditControlTestId(
  control: string,
  viewSchemaId: string,
): string {
  return viewScopedTestId(`generic-edit-${control}`, viewSchemaId);
}

function evidenceRecordControlTestId(
  control: string,
  recordId: string,
): string {
  return recordTestId(`evidence-${control}`, recordId);
}

function requireEntityType(value: EntityType): EntityType {
  if (value === "host" || value === "identity") {
    return value;
  }
  throw new Error(`Invalid entity type selector token: ${value}`);
}
