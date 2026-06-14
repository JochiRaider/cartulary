import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "./workbookSurfaceRegistry";

export type GenericCollectionMode = "add" | "remove";

const invalidGenericPayloadValue = Symbol("invalid generic payload value");

function normalizeValue(value: string): string {
  return value.trim();
}

export function buildGenericCreatePayload(
  contract: ViewContract,
  draft: Record<string, string>,
  clientTxnId: string,
): Record<string, unknown> | null {
  if (!workbookCreateMinimumSatisfied(contract.viewSchemaId, draft)) {
    return null;
  }
  const payload: Record<string, unknown> = { client_txn_id: clientTxnId };
  const fields = contract.fields.filter(
    (field) => field.writeKind !== "read_only",
  );
  for (const field of fields) {
    const value = normalizeValue(draft[field.fieldKey] ?? "");
    if (field.writeKind === "action_payload") {
      if (value === "") {
        continue;
      }
      const actionPayload = buildGenericCollectionActions(field, value, "add");
      if (actionPayload !== null) {
        payload[field.fieldKey] = actionPayload;
      }
      continue;
    }

    if (value === "") {
      if (field.clearable) {
        payload[field.fieldKey] = null;
      }
      continue;
    }
    const payloadValue = genericDirectPayloadValue(field, value);
    if (payloadValue === invalidGenericPayloadValue) {
      return null;
    }
    payload[field.fieldKey] = payloadValue;
  }
  return Object.keys(payload).length > 1 ? payload : null;
}

export function buildGenericPatchChange(
  field: ViewFieldContract,
  rawValue: string,
  collectionMode: GenericCollectionMode = "add",
): Record<string, unknown> | null {
  const value = normalizeValue(rawValue);
  if (field.writeKind === "action_payload") {
    const actionPayload = buildGenericCollectionActions(
      field,
      value,
      collectionMode,
    );
    return actionPayload === null
      ? null
      : { field_key: field.fieldKey, action_payload: actionPayload };
  }
  if (value === "" && !field.clearable) {
    return null;
  }
  const payloadValue =
    value === "" && field.clearable
      ? null
      : genericDirectPayloadValue(field, value);
  if (payloadValue === invalidGenericPayloadValue) {
    return null;
  }
  return {
    field_key: field.fieldKey,
    value: payloadValue,
  };
}

function genericDirectPayloadValue(
  field: ViewFieldContract,
  value: string,
): string | number | boolean | typeof invalidGenericPayloadValue {
  if (field.readKind === "number") {
    if (!/^-?\d+$/u.test(value)) {
      return invalidGenericPayloadValue;
    }
    const parsed = Number.parseInt(value, 10);
    return Number.isSafeInteger(parsed) ? parsed : invalidGenericPayloadValue;
  }
  if (field.readKind === "boolean") {
    if (value === "true") {
      return true;
    }
    if (value === "false") {
      return false;
    }
    return invalidGenericPayloadValue;
  }
  return value;
}

function buildGenericCollectionActions(
  field: ViewFieldContract,
  rawValue: string,
  mode: GenericCollectionMode,
): Record<string, unknown> | null {
  const tokens = splitDraftValues(rawValue);
  if (tokens.length === 0) {
    return null;
  }
  const actions = tokens.map((value) => {
    if (field.fieldKey === "note.tags") {
      return mode === "remove"
        ? { op: "remove_tag", item_ref: value }
        : { op: "add_tag", tag_name: value };
    }
    if (isPartyRefCollection(field.fieldKey)) {
      return mode === "remove"
        ? { op: "remove_party_ref", item_ref: value }
        : { op: "add_party_ref", party_id: value };
    }
    if (field.fieldKey === "handoff.open_risk_refs") {
      return mode === "remove"
        ? { op: "remove_risk_ref", item_ref: value }
        : { op: "add_risk_ref", risk_ref_text: value };
    }
    return mode === "remove"
      ? { op: "remove_record_ref", item_ref: value }
      : { op: "add_record_ref", linked_record_id: value };
  });
  return { kind: "collection_actions_v1", actions };
}

export function splitDraftValues(rawValue: string): string[] {
  return rawValue
    .split(/\r?\n/u)
    .map((value) => normalizeValue(value))
    .filter((value) => value !== "");
}

export function workbookCreateMinimumSatisfied(
  viewSchemaId: string,
  draft: Record<string, string>,
): boolean {
  const has = (fieldKey: string) =>
    normalizeValue(draft[fieldKey] ?? "") !== "";
  switch (viewSchemaId) {
    case partiesViewSchemaId:
      return has("party.display_name") && has("party.party_kind");
    case notesViewSchemaId:
      return has("note.title") || has("note.body");
    case taskRequestsViewSchemaId:
      return has("task.title") && has("task.task_kind");
    case decisionsViewSchemaId:
      return (
        has("decision.summary") &&
        has("decision.decision_type") &&
        has("decision.rationale")
      );
    case evidenceViewSchemaId:
      return (
        has("evidence.title") ||
        has("evidence.storage_ref") ||
        has("evidence.collector_party_text") ||
        has("evidence.source_party_text")
      );
    case commLogViewSchemaId:
      return (
        has("comm_log.comm_type") &&
        has("comm_log.audience") &&
        has("comm_log.channel_or_meeting") &&
        has("comm_log.summary")
      );
    case handoffViewSchemaId:
      return (
        has("handoff.incoming_owner_user_id") &&
        has("handoff.current_state_summary")
      );
    case statusReviewViewSchemaId:
      return has("status_review.current_state_summary");
    case lessonViewSchemaId:
      return has("lesson.summary");
    case findingsViewSchemaId:
      return has("finding.statement");
    case investigativeQueriesViewSchemaId:
      return (
        has("investigative_query.platform") &&
        has("investigative_query.purpose") &&
        has("investigative_query.query_text")
      );
    case forensicKeywordsViewSchemaId:
      return has("forensic_keyword.pattern") && has("forensic_keyword.reason");
    default:
      return Object.values(draft).some((value) => normalizeValue(value) !== "");
  }
}

export function initialGenericCreateDraft(
  contract: ViewContract,
  currentUserId: string | null,
): Record<string, string> {
  const draft: Record<string, string> = {};
  for (const field of contract.fields) {
    if (field.writeKind === "read_only") {
      continue;
    }
    if (
      currentUserId &&
      (field.fieldKey === "task.owner_user_id" ||
        field.fieldKey === "decision.owner_user_id" ||
        field.fieldKey === "finding.owner_user_id" ||
        field.fieldKey === "status_review.review_owner_user_id" ||
        field.fieldKey === "lesson.owner_user_id")
    ) {
      draft[field.fieldKey] = currentUserId;
    }
    if (field.fieldKey === "finding.kind") {
      draft[field.fieldKey] = "finding";
    }
    if (field.fieldKey === "finding.state") {
      draft[field.fieldKey] = "open";
    }
    if (field.fieldKey === "forensic_keyword.match_mode") {
      draft[field.fieldKey] = "literal";
    }
    if (field.fieldKey === "forensic_keyword.case_sensitive") {
      draft[field.fieldKey] = "false";
    }
  }
  return draft;
}

export function isPartyRefCollection(fieldKey: string): boolean {
  return (
    fieldKey === "comm_log.audience_party_ids" ||
    fieldKey === "comm_log.attendee_party_ids"
  );
}
