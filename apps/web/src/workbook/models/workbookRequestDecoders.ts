import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import type {
  WorkbookProtocolCollectionActions,
  WorkbookProtocolCreateLinkedNoteRequest,
  WorkbookProtocolCreateViewRowRequest,
  WorkbookProtocolPatchRecordRequest,
} from "../adapters/workbookProtocolTypes";

type CollectionAction = WorkbookProtocolCollectionActions["actions"][number];
type RecordPatchChange = WorkbookProtocolPatchRecordRequest["changes"][number];

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function hasOnlyKeys(
  value: Readonly<Record<string, unknown>>,
  keys: ReadonlySet<string>,
): boolean {
  return Object.keys(value).every((key) => keys.has(key));
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim() !== "";
}

const itemReferenceKeys = new Set(["item_ref", "op"]);

function validItemReferenceAction(
  value: Readonly<Record<string, unknown>>,
): boolean {
  return (
    hasOnlyKeys(value, itemReferenceKeys) && isNonEmptyString(value.item_ref)
  );
}

const collectionActionValidators = {
  add_alias: (value: Readonly<Record<string, unknown>>) =>
    hasOnlyKeys(value, new Set(["alias_text", "op"])) &&
    isNonEmptyString(value.alias_text),
  add_party_ref: (value: Readonly<Record<string, unknown>>) =>
    hasOnlyKeys(value, new Set(["op", "party_id"])) &&
    isNonEmptyString(value.party_id),
  add_record_ref: (value: Readonly<Record<string, unknown>>) =>
    hasOnlyKeys(value, new Set(["linked_record_id", "op"])) &&
    isNonEmptyString(value.linked_record_id),
  add_resolved_ref: (value: Readonly<Record<string, unknown>>) =>
    hasOnlyKeys(value, new Set(["op", "raw_text", "resolved_record_id"])) &&
    isNonEmptyString(value.raw_text) &&
    isNonEmptyString(value.resolved_record_id),
  add_risk_ref: (value: Readonly<Record<string, unknown>>) =>
    hasOnlyKeys(value, new Set(["op", "risk_ref_text"])) &&
    isNonEmptyString(value.risk_ref_text),
  add_tag: (value: Readonly<Record<string, unknown>>) =>
    hasOnlyKeys(value, new Set(["op", "tag_name"])) &&
    isNonEmptyString(value.tag_name),
  add_token: (value: Readonly<Record<string, unknown>>) =>
    hasOnlyKeys(value, new Set(["op", "raw_text"])) &&
    isNonEmptyString(value.raw_text),
  dismiss_item: validItemReferenceAction,
  remove_alias: validItemReferenceAction,
  remove_party_ref: validItemReferenceAction,
  remove_record_ref: validItemReferenceAction,
  remove_risk_ref: validItemReferenceAction,
  remove_tag: validItemReferenceAction,
  resolve_item: (value: Readonly<Record<string, unknown>>) =>
    hasOnlyKeys(value, new Set(["item_ref", "op", "resolved_record_id"])) &&
    isNonEmptyString(value.item_ref) &&
    isNonEmptyString(value.resolved_record_id),
  revert_to_unresolved: validItemReferenceAction,
} satisfies Readonly<
  Record<
    CollectionAction["op"],
    (value: Readonly<Record<string, unknown>>) => boolean
  >
>;

function isCollectionActionKind(
  value: string,
): value is CollectionAction["op"] {
  return Object.hasOwn(collectionActionValidators, value);
}

function isCollectionAction(value: unknown): value is CollectionAction {
  if (
    !isRecord(value) ||
    typeof value.op !== "string" ||
    !isCollectionActionKind(value.op)
  ) {
    return false;
  }
  return collectionActionValidators[value.op](value);
}

export function decodeCollectionActions(
  value: unknown,
): WorkbookProtocolCollectionActions | null {
  if (
    !isRecord(value) ||
    !hasOnlyKeys(value, new Set(["actions", "kind"])) ||
    value.kind !== "collection_actions_v1" ||
    !Array.isArray(value.actions) ||
    value.actions.length < 1 ||
    value.actions.length > 64 ||
    !value.actions.every(isCollectionAction)
  ) {
    return null;
  }
  const [firstAction, ...remainingActions] = value.actions;
  return firstAction === undefined
    ? null
    : { actions: [firstAction, ...remainingActions], kind: value.kind };
}

function isRFC3339Timestamp(value: unknown): boolean {
  return (
    typeof value === "string" &&
    /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/u.test(
      value,
    ) &&
    !Number.isNaN(Date.parse(value))
  );
}

function validNonNullDirectCreateValue(
  field: ViewFieldContract,
  value: unknown,
): boolean {
  if (field.enumValues !== null) {
    return typeof value === "string" && field.enumValues.includes(value);
  }
  if (field.directReferenceContractId !== null) {
    return isNonEmptyString(value);
  }
  if (field.directScalarContractId !== null) {
    return (
      field.directScalarContractId === "timestamp_instant_v1" &&
      isRFC3339Timestamp(value)
    );
  }
  if (field.stringContractId !== null) return typeof value === "string";
  if (field.readKind === "number") return Number.isSafeInteger(value);
  if (field.readKind === "boolean") return typeof value === "boolean";
  if (field.readKind === "text") return typeof value === "string";
  return false;
}

function validCreateFieldValue(
  field: ViewFieldContract,
  value: unknown,
): boolean {
  if (field.writeKind === "action_payload") {
    return decodeCollectionActions(value) !== null;
  }
  if (field.writeKind !== "direct_value") return false;
  return value === null
    ? field.clearable
    : validNonNullDirectCreateValue(field, value);
}

function isCreateViewRowRequest(
  contract: ViewContract,
  value: unknown,
): value is WorkbookProtocolCreateViewRowRequest {
  if (!isRecord(value) || !isNonEmptyString(value.client_txn_id)) return false;
  for (const [key, fieldValue] of Object.entries(value)) {
    if (key === "client_txn_id") continue;
    const field = contract.fieldMap[key];
    if (field !== undefined) {
      if (!field.createWritable || !validCreateFieldValue(field, fieldValue)) {
        return false;
      }
      continue;
    }
    const createInput = contract.createInputs.find(
      (input) => input.inputKey === key,
    );
    if (createInput === undefined || typeof fieldValue !== "string") {
      return false;
    }
  }
  return true;
}

export function decodeCreateViewRowRequest(
  contract: ViewContract,
  value: unknown,
): WorkbookProtocolCreateViewRowRequest | null {
  return isCreateViewRowRequest(contract, value) ? value : null;
}

export function decodeCreateRecordLinkedNoteRequest(
  value: unknown,
): WorkbookProtocolCreateLinkedNoteRequest | null {
  if (
    !isRecord(value) ||
    !hasOnlyKeys(
      value,
      new Set(["client_txn_id", "note.body", "note.tags", "note.title"]),
    ) ||
    !isNonEmptyString(value.client_txn_id) ||
    ("note.body" in value && typeof value["note.body"] !== "string") ||
    ("note.title" in value && typeof value["note.title"] !== "string")
  ) {
    return null;
  }
  const tags =
    "note.tags" in value ? decodeCollectionActions(value["note.tags"]) : null;
  if ("note.tags" in value && tags === null) return null;
  return {
    ...(typeof value["note.body"] === "string"
      ? { "note.body": value["note.body"] }
      : {}),
    ...(tags === null ? {} : { "note.tags": tags }),
    ...(typeof value["note.title"] === "string"
      ? { "note.title": value["note.title"] }
      : {}),
    client_txn_id: value.client_txn_id,
  };
}

function decodeRecordPatchChange(value: unknown): RecordPatchChange | null {
  if (
    !isRecord(value) ||
    !hasOnlyKeys(value, new Set(["action_payload", "field_key", "value"])) ||
    !isNonEmptyString(value.field_key)
  ) {
    return null;
  }
  const hasAction = "action_payload" in value;
  const hasValue = "value" in value;
  if (hasAction === hasValue) return null;
  if (hasAction) {
    const actionPayload = decodeCollectionActions(value.action_payload);
    return actionPayload === null
      ? null
      : { action_payload: actionPayload, field_key: value.field_key };
  }
  return { field_key: value.field_key, value: value.value };
}

export function decodeRecordPatchChanges(
  values: readonly unknown[],
): WorkbookProtocolPatchRecordRequest["changes"] | null {
  const decoded = values.map(decodeRecordPatchChange);
  if (decoded.some((change) => change === null)) return null;
  const firstChange = decoded[0];
  if (firstChange === null || firstChange === undefined) return null;
  const remainingChanges: RecordPatchChange[] = [];
  for (let index = 1; index < decoded.length; index += 1) {
    const change = decoded[index];
    if (change === null || change === undefined) return null;
    remainingChanges.push(change);
  }
  return [firstChange, ...remainingChanges];
}

export function buildPatchRecordRequest(input: {
  readonly baseRowVersion: number;
  readonly changes: readonly unknown[];
  readonly clientTxnId: string;
  readonly viewSchemaId: string;
}): WorkbookProtocolPatchRecordRequest | null {
  const changes = decodeRecordPatchChanges(input.changes);
  if (
    !Number.isInteger(input.baseRowVersion) ||
    input.baseRowVersion < 1 ||
    !isNonEmptyString(input.clientTxnId) ||
    !isNonEmptyString(input.viewSchemaId) ||
    changes === null
  ) {
    return null;
  }
  return {
    base_row_version: input.baseRowVersion,
    changes,
    client_txn_id: input.clientTxnId,
    view_schema_id: input.viewSchemaId,
  };
}
