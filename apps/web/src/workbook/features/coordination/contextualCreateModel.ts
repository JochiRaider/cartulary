import {
  decisionsViewSchemaId,
  getViewContract,
  type InspectorFeatureGroup,
  taskRequestsViewSchemaId,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookProtocolCreateViewRowRequest } from "../../adapters/workbookProtocolTypes";
import { buildInspectorRelatedRecordDraft } from "../../inspector/inspectorRelatedRecordModel";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import {
  buildGenericPatchChange,
  workbookCreateMinimumSatisfied,
  workbookCreationAvailable,
} from "../../models/genericWorkbookModel";
import { decodeCreateViewRowRequest } from "../../models/workbookRequestDecoders";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import { freezeWorkbookValue as freezeContextualCreate } from "../../utils/freezeWorkbookValue";

export { freezeWorkbookValue as freezeContextualCreate } from "../../utils/freezeWorkbookValue";

export type ContextualCreateFeature =
  | "create_related.task_request"
  | "create_related.decision";
export function isContextualCreateFeature(
  key: string,
): key is ContextualCreateFeature {
  return (
    key === "create_related.task_request" || key === "create_related.decision"
  );
}
export function contextualTarget(key: ContextualCreateFeature) {
  return key === "create_related.task_request"
    ? taskRequestsViewSchemaId
    : decisionsViewSchemaId;
}
export type ContextualCreateDraft = Readonly<{
  id: number;
  revision: number;
  actorId: string;
  incidentId: string;
  source: {
    readonly recordId: string;
    readonly viewSchemaId: string;
    readonly rowVersion: number;
  };
  presentation: {
    readonly sheetRef: SheetRef;
    readonly label: string;
    readonly surfaceLabel: string;
  };
  feature: InspectorFeatureGroup;
  target: ViewContract;
  seeds: Readonly<Record<string, string>>;
  values: Readonly<Record<string, string>>;
  labels: Readonly<Record<string, string>>;
}>;

export function contextualCreateDraft(
  id: number,
  authority: WorkbookMutationAuthority,
  subject: WorkbookInspectorLiveRowBinding,
  feature: InspectorFeatureGroup,
  sheetRef: SheetRef,
): ContextualCreateDraft | null {
  if (!isContextualCreateFeature(feature.featureGroupKey)) return null;
  const source = getViewContract(subject.subject.viewSchemaId);
  const declared = source?.inspectorConfig.featureGroups.find(
    (group) => group.featureGroupKey === feature.featureGroupKey,
  );
  const target = getViewContract(contextualTarget(feature.featureGroupKey));
  if (
    !declared ||
    !target ||
    JSON.stringify(declared) !== JSON.stringify(feature) ||
    !workbookCreationAvailable(target) ||
    feature.seedBindings.some(
      (binding) => !target.fieldMap[binding.targetFieldKey]?.createWritable,
    )
  )
    return null;
  const result = buildInspectorRelatedRecordDraft({
    currentUserId: authority.actorId,
    featureGroup: declared,
    subject,
    targetContract: target,
  });
  if (result.kind !== "ready") return null;
  const values = { ...result.draft };
  if (target.viewSchemaId === taskRequestsViewSchemaId) {
    values["task.status"] ||= "open";
    values["task.priority"] ||= "normal";
  } else values["decision.status"] ||= "proposed";
  return freezeContextualCreate({
    id,
    revision: 0,
    actorId: authority.actorId,
    incidentId: authority.incidentId,
    source: {
      recordId: subject.subject.recordId,
      viewSchemaId: subject.subject.viewSchemaId,
      rowVersion: subject.subject.rowVersion,
    },
    presentation: {
      sheetRef: structuredClone(sheetRef),
      label: subject.subject.label,
      surfaceLabel: subject.subject.surfaceLabel,
    },
    feature: declared,
    target,
    seeds: Object.fromEntries(
      declared.seedBindings.map((binding) => [
        binding.targetFieldKey,
        result.draft[binding.targetFieldKey] ?? "",
      ]),
    ),
    values,
    labels: { [subject.subject.recordId]: subject.subject.label },
  });
}

export function contextualReferenceKind(
  field: ViewFieldContract,
): "members" | "parties" | "decisions" | "records" | null {
  if (field.directReferenceContractId === "incident_member_user_ref_v1")
    return "members";
  if (field.directReferenceContractId === "same_incident_party_ref_v1")
    return "parties";
  if (field.directReferenceContractId === "same_incident_decision_ref_v1")
    return "decisions";
  if (
    [
      "task.linked_record_ids",
      "decision.support_refs",
      "decision.affected_record_ids",
    ].includes(field.fieldKey)
  )
    return "records";
  return null;
}
export function contextualReferenceIds(value: string): string[] {
  return value === "" ? [] : value.split("\n");
}
export function contextualSelectedIds(
  draft: ContextualCreateDraft,
): readonly string[] {
  return draft.target.fields.flatMap((field) =>
    contextualReferenceKind(field) &&
    contextualReferenceKind(field) !== "members"
      ? contextualReferenceIds(draft.values[field.fieldKey] ?? "")
      : [],
  );
}
const stableId =
  /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/u;
export function contextualCreateErrors(
  draft: ContextualCreateDraft,
): Record<string, string> {
  const errors: Record<string, string> = {};
  const { values, target } = draft;
  if (!workbookCreateMinimumSatisfied(target, { ...values })) {
    for (const key of target.minimumCreateFieldSets[0] ?? [])
      if (!values[key]?.trim())
        errors[key] = `${target.fieldMap[key]?.label ?? key} is required.`;
  }
  for (const field of target.fields.filter((field) => field.createWritable)) {
    const raw = values[field.fieldKey] ?? "";
    if (!raw) continue;
    const kind = contextualReferenceKind(field);
    if (kind) {
      const ids = contextualReferenceIds(raw);
      if (
        ids.some((id) => !stableId.test(id)) ||
        new Set(ids).size !== ids.length ||
        (field.readKind === "collection" ? ids.length > 64 : ids.length !== 1)
      )
        errors[field.fieldKey] =
          "Choose valid references (at most 64 per collection).";
    } else if (
      (field.enumValues && !field.enumValues.includes(raw)) ||
      buildGenericPatchChange(field, raw, "add", target.viewSchemaId) === null
    ) {
      errors[field.fieldKey] = `Enter a valid ${field.label.toLowerCase()}.`;
    }
  }
  for (const input of target.createInputs)
    if (input.required && !values[input.inputKey]?.trim())
      errors[input.inputKey] = `${input.inputKey} is required.`;
  if (target.viewSchemaId === taskRequestsViewSchemaId) {
    const status = values["task.status"] || "open";
    if (status === "blocked" && !values["task.blocked_reason"]?.trim())
      errors["task.blocked_reason"] = "Blocked Tasks need a reason.";
    if (status !== "blocked" && values["task.blocked_reason"]?.trim())
      errors["task.blocked_reason"] =
        "A reason is only saved while the Task is blocked.";
    if (status !== "done" && values["task.completed_at"]?.trim())
      errors["task.completed_at"] =
        "Completion time is only saved for a done Task.";
  } else if (values["decision.status"] === "superseded")
    errors["decision.status"] =
      "Create a proposed, approved, rejected, or executed Decision.";
  return errors;
}
export function contextualCreateRequest(
  draft: ContextualCreateDraft,
  clientTxnId: string,
): WorkbookProtocolCreateViewRowRequest | null {
  if (Object.keys(contextualCreateErrors(draft)).length) return null;
  const request: Record<string, unknown> = { client_txn_id: clientTxnId };
  for (const field of draft.target.fields.filter(
    (field) => field.createWritable,
  )) {
    const raw = draft.values[field.fieldKey] ?? "";
    if (!raw.trim()) continue;
    const change = buildGenericPatchChange(
      field,
      raw,
      "add",
      draft.target.viewSchemaId,
    );
    if (!change) return null;
    request[field.fieldKey] =
      "value" in change ? change.value : change.action_payload;
  }
  for (const input of draft.target.createInputs) {
    const raw = draft.values[input.inputKey];
    if (raw?.trim()) request[input.inputKey] = raw.trim();
  }
  return decodeCreateViewRowRequest(draft.target, request);
}
