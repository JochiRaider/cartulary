import type { WorkbookSurface } from "@cartulary/ui-contracts";
import {
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import { stringifyGridValue } from "../utils/workbookValueFormat";
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

export type PartyLinkPair = {
  key: string;
  label: string;
  textFieldKey: string;
  refFieldKey: string;
};

export function selectWorkbookEditTarget<
  Row,
  Field extends { readonly fieldKey: string },
>({
  fieldKey,
  fields,
  getRecordId,
  recordId,
  rows,
}: {
  readonly fieldKey: string;
  readonly fields: readonly Field[];
  readonly getRecordId: (row: Row) => string;
  readonly recordId: string;
  readonly rows: readonly Row[];
}): { readonly field: Field | null; readonly row: Row | null } {
  return {
    row: rows.find((row) => getRecordId(row) === recordId) ?? null,
    field:
      fields.find((field) => field.fieldKey === fieldKey) ?? fields[0] ?? null,
  };
}

export function normalizeGenericTextValue(value: string): string {
  return value.trim();
}

function normalizeValue(value: string): string {
  return normalizeGenericTextValue(value);
}

export function buildGenericCreatePayload(
  contract: ViewContract,
  draft: Record<string, string>,
  clientTxnId: string,
): Record<string, unknown> | null {
  if (!workbookCreateMinimumSatisfied(contract, draft)) {
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
    if (
      field.fieldKey === "host.aliases" ||
      field.fieldKey === "identity.aliases"
    ) {
      return mode === "remove"
        ? { op: "remove_alias", item_ref: value }
        : { op: "add_alias", alias_text: value };
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
  contractOrViewSchemaId: ViewContract | string,
  draft: Record<string, string>,
): boolean {
  const has = (fieldKey: string) =>
    normalizeValue(draft[fieldKey] ?? "") !== "";
  const contract =
    typeof contractOrViewSchemaId === "string" ? null : contractOrViewSchemaId;
  if (contract && contract.minimumCreateFieldSets.length > 0) {
    return contract.minimumCreateFieldSets.some((fieldSet) =>
      fieldSet.every((fieldKey) => has(fieldKey)),
    );
  }
  const viewSchemaId =
    typeof contractOrViewSchemaId === "string"
      ? contractOrViewSchemaId
      : contractOrViewSchemaId.viewSchemaId;
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

export function genericContractColumnWidth(field: ViewFieldContract): number {
  if (field.fieldKey.endsWith(".body") || field.fieldKey.endsWith(".summary")) {
    return 320;
  }
  if (
    field.fieldKey.endsWith(".edited_at") ||
    field.fieldKey.endsWith(".updated_at") ||
    field.fieldKey.endsWith(".timestamp_utc")
  ) {
    return 180;
  }
  if (field.readKind === "collection") {
    return 260;
  }
  return field.defaultHidden ? 160 : 220;
}

export function genericCellLabel(value: unknown): string {
  if (value === null || value === undefined || value === "") {
    return "None";
  }
  if (typeof value === "string" || typeof value === "number") {
    return String(value);
  }
  if (typeof value === "boolean") {
    return value ? "Yes" : "No";
  }
  if (typeof value === "object" && value !== null && "items" in value) {
    const items = (value as { items?: unknown }).items;
    if (Array.isArray(items)) {
      const labels = collectionItemLabels(items);
      if (labels.length > 0) {
        return labels.join(", ");
      }
      return "None";
    }
  }
  return JSON.stringify(value);
}

export function genericCellLabelForField(
  surface: WorkbookSurface,
  fieldKey: string,
  value: unknown,
): string {
  if (
    surface === evidenceViewSchemaId &&
    fieldKey === "evidence.storage_ref" &&
    typeof value === "string" &&
    /^object:\/\/[0-9a-f-]+$/iu.test(value.trim())
  ) {
    return "Managed object";
  }
  return genericCellLabel(value);
}

export function collectionItemLabels(items: readonly unknown[]): string[] {
  return items.flatMap((item) => {
    if (!item || typeof item !== "object" || Array.isArray(item)) {
      return [];
    }
    const raw = item as Record<string, unknown>;
    const candidates = [
      raw.display_text,
      raw.alias_text,
      raw.tag_name,
      raw.raw_text,
      raw.linked_record_id,
      raw.record_id,
      raw.item_ref,
    ];
    const label = candidates.find(
      (value): value is string =>
        typeof value === "string" && value.trim() !== "",
    );
    return label === undefined ? [] : [label];
  });
}

export function genericCreateMinimumMessage(viewSchemaId: string): string {
  switch (viewSchemaId) {
    case partiesViewSchemaId:
      return "Display name and kind are required.";
    case notesViewSchemaId:
      return "Title or body is required.";
    case taskRequestsViewSchemaId:
      return "Title and task kind are required.";
    case decisionsViewSchemaId:
      return "Summary, decision type, and rationale are required.";
    case evidenceViewSchemaId:
      return "Evidence needs a title, storage ref, collector, or source.";
    case commLogViewSchemaId:
      return "Type, audience, channel or meeting, and summary are required.";
    case handoffViewSchemaId:
      return "Incoming owner and current state summary are required.";
    case statusReviewViewSchemaId:
      return "Current state summary is required.";
    case lessonViewSchemaId:
      return "Summary is required.";
    case findingsViewSchemaId:
      return "Statement is required.";
    case investigativeQueriesViewSchemaId:
      return "Platform, purpose, and query text are required.";
    case forensicKeywordsViewSchemaId:
      return "Pattern and reason are required.";
    default:
      return "At least one value is required.";
  }
}

export function genericReferenceOptionsFromRows(
  viewSchemaId: string,
  rows: EntityApiRow[],
) {
  return rows.map((row) => ({
    recordId: row.record_id,
    label: genericRowLabel(requireViewContract(viewSchemaId), row),
    viewSchemaId,
  }));
}

export function partyLinkPairsForContract(
  contract: ViewContract,
): PartyLinkPair[] {
  const hasField = (fieldKey: string) => Boolean(contract.fieldMap[fieldKey]);
  const pairs: PartyLinkPair[] = [];
  if (
    hasField("evidence.collector_party_text") &&
    hasField("evidence.collector_party_id")
  ) {
    pairs.push({
      key: "evidence.collector_party_text:evidence.collector_party_id",
      label: "Collector",
      textFieldKey: "evidence.collector_party_text",
      refFieldKey: "evidence.collector_party_id",
    });
  }
  if (
    hasField("evidence.source_party_text") &&
    hasField("evidence.source_party_id")
  ) {
    pairs.push({
      key: "evidence.source_party_text:evidence.source_party_id",
      label: "Source",
      textFieldKey: "evidence.source_party_text",
      refFieldKey: "evidence.source_party_id",
    });
  }
  if (
    hasField("task.requester_party_text") &&
    hasField("task.requester_party_id")
  ) {
    pairs.push({
      key: "task.requester_party_text:task.requester_party_id",
      label: "Requester",
      textFieldKey: "task.requester_party_text",
      refFieldKey: "task.requester_party_id",
    });
  }
  return pairs;
}

export function extractEmailFromPartyText(value: string): string | null {
  const match = value.match(/[^\s<>@]+@[^\s<>@]+/u);
  return match?.[0] ?? null;
}

export function genericRowLabel(
  contract: ViewContract,
  row: EntityApiRow,
): string {
  const preferredFieldKeys = [
    "timeline.activity_synopsis_text",
    "host.display_name",
    "host.hostname",
    "identity.display_name",
    "identity.upn",
    "party.display_name",
    "task.title",
    "decision.summary",
    "evidence.title",
    "evidence.storage_ref",
    "note.title",
    "note.body",
    "comm_log.summary",
    "handoff.current_state_summary",
    "status_review.current_state_summary",
    "lesson.summary",
    "finding.statement",
    "investigative_query.purpose",
    "investigative_query.query_text",
    "forensic_keyword.pattern",
  ];
  for (const fieldKey of preferredFieldKeys) {
    if (!contract.fieldMap[fieldKey]) {
      continue;
    }
    const label = stringifyGridValue(row.cells[fieldKey]?.value).trim();
    if (label !== "") {
      return `${label} (${row.record_id})`;
    }
  }
  return row.record_id;
}

export function genericCollectionSupportsRemove(_fieldKey: string): boolean {
  return true;
}

export function genericCollectionItems(
  row: EntityApiRow,
  fieldKey: string,
): Array<{ itemRef: string; displayText: string }> {
  const value = row.cells[fieldKey]?.value;
  if (!value || typeof value !== "object" || !("items" in value)) {
    return [];
  }
  const rawItems = (value as { items?: unknown }).items;
  if (!Array.isArray(rawItems)) {
    return [];
  }
  return rawItems.flatMap((item) => {
    if (!item || typeof item !== "object") {
      return [];
    }
    const raw = item as Record<string, unknown>;
    const itemRef = typeof raw.item_ref === "string" ? raw.item_ref : "";
    if (itemRef === "") {
      return [];
    }
    const displayText =
      typeof raw.display_text === "string" && raw.display_text.trim() !== ""
        ? raw.display_text
        : itemRef;
    return [{ itemRef, displayText }];
  });
}

export function isMultilineGenericField(field: ViewFieldContract): boolean {
  return (
    field.stringContractId === "multiline_body_v1" ||
    field.fieldKey.endsWith(".body") ||
    field.fieldKey.endsWith(".notes") ||
    field.fieldKey.endsWith(".rationale") ||
    field.fieldKey.endsWith("_summary") ||
    field.fieldKey.endsWith(".details")
  );
}

function parseGenericErrorBase(payload: unknown): string {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return "Request failed.";
  }
  const error = payload.error;
  if (!error || typeof error !== "object") {
    return "Request failed.";
  }
  if ("code" in error && typeof error.code === "string") {
    if (
      "details" in error &&
      error.details &&
      typeof error.details === "object" &&
      "reason_code" in error.details &&
      typeof error.details.reason_code === "string"
    ) {
      return `${error.code}: ${error.details.reason_code}`;
    }
    return error.code;
  }
  if ("message" in error && typeof error.message === "string") {
    return error.message;
  }
  return "Request failed.";
}

export function parseMutationError(payload: unknown): string {
  const base = parseGenericErrorBase(payload);
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return base;
  }
  const error = payload.error;
  if (!error || typeof error !== "object" || !("conflict" in error)) {
    return base;
  }
  const conflict = error.conflict;
  if (!conflict || typeof conflict !== "object") {
    return base;
  }
  const fieldKey =
    "field_key" in conflict && typeof conflict.field_key === "string"
      ? conflict.field_key
      : null;
  return fieldKey ? `${base}: ${fieldKey}` : base;
}

export function enumValuesFor(
  contract: ViewContract,
  fieldKey: string,
  fallback: readonly string[],
): readonly string[] {
  return contract.fieldMap[fieldKey]?.enumValues ?? fallback;
}
