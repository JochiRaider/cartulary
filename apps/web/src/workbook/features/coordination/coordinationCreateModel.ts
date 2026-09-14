import {
  getViewContract,
  type InspectorFeatureGroup,
  listViewContracts,
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookProtocolCreateViewRowRequest } from "../../adapters/workbookProtocolTypes";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import {
  normalizeWorkbookAuthoringText as normalizeCoordinationText,
  validWorkbookTimestamp as validCoordinationTimestamp,
} from "../../models/workbookAuthoringValues";
import { decodeCreateViewRowRequest } from "../../models/workbookRequestDecoders";
import { freezeWorkbookValue } from "../../utils/freezeWorkbookValue";

export const coordinationSourceInput = "coordination.source_record_id";
export const coordinationVariants = [
  "comm_log",
  "handoff",
  "status_review",
  "lesson",
] as const;
export type CoordinationVariant = (typeof coordinationVariants)[number];
export type CoordinationSource = Readonly<{
  recordId: string;
  viewSchemaId: string;
  rowVersion: number;
  label: string;
}>;
export type CoordinationDraft = Readonly<{
  id: number;
  revision: number;
  actorId: string;
  incidentId: string;
  variant: CoordinationVariant;
  target: ViewContract;
  values: Readonly<Record<string, string | null>>;
  labels: Readonly<Record<string, string>>;
  source: CoordinationSource | null;
  origin: Readonly<{
    sheetRef: SheetRef;
    subject: WorkbookInspectorLiveRowBinding["subject"];
    feature: InspectorFeatureGroup;
  }>;
}>;
export function coordinationVariant(
  feature: string,
): CoordinationVariant | null {
  return (
    coordinationVariants.find(
      (variant) => feature === `create_related.${variant}`,
    ) ?? null
  );
}
export function coordinationFeature(
  view: string,
  variant: CoordinationVariant,
) {
  const feature = getViewContract(view)?.inspectorConfig.featureGroups.find(
    (f) => f.featureGroupKey === `create_related.${variant}`,
  );
  return feature?.routeBinding.kind === "view_row_create" &&
    feature.routeBinding.owner === "view_row_create_route" &&
    feature.routeBinding.targetViewSchemaId ===
      `cartulary.view.${variant}.v1` &&
    feature.seedBindings.length === 1 &&
    feature.seedBindings[0]?.targetInputKey === coordinationSourceInput &&
    feature.seedBindings[0]?.source.kind === "selected_record_id"
    ? feature
    : null;
}
export function coordinationSourceViews(
  variant: CoordinationVariant,
): readonly string[] {
  return listViewContracts()
    .filter((view) => coordinationFeature(view.viewSchemaId, variant))
    .map((view) => view.viewSchemaId);
}
export function coordinationSource(
  subject: WorkbookInspectorLiveRowBinding,
): CoordinationSource {
  return freezeWorkbookValue({
    recordId: subject.subject.recordId,
    viewSchemaId: subject.subject.viewSchemaId,
    rowVersion: subject.subject.rowVersion,
    label: subject.subject.label,
  });
}
export const coordinationTarget = (variant: CoordinationVariant) =>
  requireViewContract(`cartulary.view.${variant}.v1`);
const referenceViews: Readonly<Record<string, string>> = {
  "comm_log.decision_ids": "decisions",
  "comm_log.action_task_ids": "task_requests",
  "comm_log.audience_party_ids": "parties",
  "comm_log.attendee_party_ids": "parties",
  "handoff.open_task_ids": "task_requests",
  "handoff.open_decision_ids": "decisions",
  "status_review.blocked_task_ids": "task_requests",
  "status_review.pending_evidence_ids": "evidence",
  "status_review.open_decision_ids": "decisions",
  "lesson.follow_up_task_ids": "task_requests",
  "lesson.evidence_refs": "evidence",
};
export function coordinationReferenceView(
  field: ViewFieldContract,
): string | null {
  if (field.directReferenceContractId === "incident_member_user_ref_v1")
    return "incident_members";
  const surface = referenceViews[field.fieldKey];
  return surface ? `cartulary.view.${surface}.v1` : null;
}
export function coordinationIds(
  raw: string | null | undefined,
): readonly string[] {
  return raw ? raw.split("\n") : [];
}
export const exactRecordId = (raw: string) =>
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u.test(raw);
export { normalizeWorkbookAuthoringText as normalizeCoordinationText } from "../../models/workbookAuthoringValues";
export function prepareCoordination(
  draft: CoordinationDraft,
  clientTxnId: string,
) {
  const errors: Record<string, string> = {};
  const request: Record<string, unknown> = {
    client_txn_id: clientTxnId,
    [coordinationSourceInput]: draft.source?.recordId ?? null,
  };
  const required = new Set(draft.target.minimumCreateFieldSets[0] ?? []);
  for (const field of draft.target.fields.filter((f) => f.createWritable)) {
    const key = field.fieldKey;
    if (!Object.hasOwn(draft.values, key)) {
      if (required.has(key)) errors[key] = `${field.label} is required.`;
      continue;
    }
    const raw = draft.values[key];
    if (raw == null) {
      if (
        field.clearable &&
        field.readKind !== "collection" &&
        !required.has(key)
      )
        request[key] = null;
      else errors[key] = `${field.label} cannot be cleared.`;
      continue;
    }
    const refView = coordinationReferenceView(field);
    if (refView) {
      const ids = coordinationIds(raw);
      if (
        ids.some((id) => !exactRecordId(id)) ||
        new Set(ids).size !== ids.length ||
        ids.length > 64 ||
        (field.readKind !== "collection" && ids.length !== 1)
      )
        errors[key] = `Choose valid ${field.label.toLowerCase()}.`;
      else if (field.readKind === "collection") {
        if (ids.length)
          request[key] = {
            kind: "collection_actions_v1",
            actions: ids.map((id) =>
              refView === "cartulary.view.parties.v1"
                ? { op: "add_party_ref", party_id: id }
                : { op: "add_record_ref", linked_record_id: id },
            ),
          };
      } else request[key] = ids[0];
    } else if (key === "handoff.open_risk_refs") {
      const risks = raw
        .split(/\r\n?|\n/u)
        .map((value) =>
          normalizeCoordinationText(value, "single_line_title_v1"),
        );
      const invalid = risks.find((risk) => risk.error);
      const values = [
        ...new Set(risks.map((risk) => risk.value).filter(Boolean)),
      ];
      if (invalid?.error) errors[key] = invalid.error;
      else if (risks.filter((risk) => risk.value).length > 64)
        errors[key] = "Use at most 64 risk references.";
      else if (values.length)
        request[key] = {
          kind: "collection_actions_v1",
          actions: values.map((risk_ref_text) => ({
            op: "add_risk_ref",
            risk_ref_text,
          })),
        };
    } else if (field.stringContractId) {
      const normalized = normalizeCoordinationText(raw, field.stringContractId);
      if (normalized.error) errors[key] = normalized.error;
      else if (!normalized.value && !field.clearable)
        errors[key] = `${field.label} is required.`;
      else request[key] = normalized.value || null;
    } else if (field.enumValues) {
      if (!field.enumValues.includes(raw))
        errors[key] = `Choose ${field.label.toLowerCase()}.`;
      else request[key] = raw;
    } else if (field.directScalarContractId === "timestamp_instant_v1") {
      // The protocol decoder performs the declared RFC 3339 validation; preserve the exact input.
      request[key] = raw;
    } else errors[key] = "This create field is unavailable.";
  }
  for (const key of Object.keys(draft.values))
    if (!draft.target.fieldMap[key]?.createWritable)
      errors[key] = "This field is not available on create.";
  if (
    draft.source &&
    (!exactRecordId(draft.source.recordId) ||
      !coordinationFeature(draft.source.viewSchemaId, draft.variant))
  )
    errors[coordinationSourceInput] =
      "Choose an available source for this artifact.";
  if (
    !draft.target.createCapable ||
    draft.target.createInputs.length !== 1 ||
    draft.target.createInputs[0]?.inputKey !== coordinationSourceInput
  )
    errors[coordinationSourceInput] =
      "Creation capability changed. Review the retained draft.";
  for (const field of draft.target.fields.filter(
    (f) =>
      f.directScalarContractId === "timestamp_instant_v1" &&
      Object.hasOwn(request, f.fieldKey),
  )) {
    const value = request[field.fieldKey];
    if (value === null && field.clearable) continue;
    if (!validCoordinationTimestamp(value))
      errors[field.fieldKey] = "Enter an RFC 3339 timestamp with a timezone.";
  }
  const decoded = Object.keys(errors).length
    ? null
    : decodeCreateViewRowRequest(draft.target, request);
  if (!decoded && !Object.keys(errors).length)
    errors[coordinationSourceInput] =
      "The create contract could not validate these values.";
  return {
    errors,
    request: decoded
      ? (freezeWorkbookValue(decoded) as WorkbookProtocolCreateViewRowRequest)
      : null,
  };
}
