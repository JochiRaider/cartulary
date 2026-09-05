import type { SheetRef } from "../../shared/sheetRef";
import type {
  WorkbookProtocolCollectionActions,
  WorkbookProtocolResolveConflictRequest,
} from "../adapters/workbookProtocolTypes";

export type WorkbookSameFieldConflictPayload = {
  conflict_token: string;
  record_id: string;
  field_key: string;
  conflict_resolution_class: WorkbookConflictResolutionClass;
  base_row_version: number;
  current_row_version: number;
  client_value: unknown;
  server_value: unknown;
  base_value?: unknown;
  server_updated_by?: string;
  server_updated_at?: string;
  suggested_merged_value?: unknown;
};

export type WorkbookConflictResolutionClass =
  | "atomic_replace"
  | "collection_review"
  | "text_compare_merge";

type WorkbookCollectionValue = {
  readonly kind: "collection_value_v1";
  readonly ordered: boolean;
  readonly items: readonly Record<string, unknown>[];
};

export type WorkbookCollectionActions = WorkbookProtocolCollectionActions;

type WorkbookCollectionAction =
  WorkbookProtocolCollectionActions["actions"][number];

export type WorkbookConflictEntry = {
  readonly key: string;
  readonly conflict: WorkbookSameFieldConflictPayload;
  readonly resolutionClass: WorkbookConflictResolutionClass;
  readonly origin: {
    readonly sheetRef: SheetRef;
    readonly viewSchemaId: string;
    readonly surfaceLabel: string;
    readonly rowLabel: string;
  };
  readonly focusKey: string | null;
  readonly localValue: unknown;
  readonly mergedDraft: string;
};

export type WorkbookConflictResolutionKind =
  | "keep_saved"
  | "merged_value"
  | "use_unsaved";

export function workbookConflictResolutionClass(
  value: unknown,
): WorkbookConflictResolutionClass | null {
  switch (value) {
    case "atomic_replace":
    case "collection_review":
    case "text_compare_merge":
      return value;
  }
  return null;
}

export function workbookConflictQueueKey(
  conflict: Pick<WorkbookSameFieldConflictPayload, "field_key" | "record_id">,
): string {
  return `${conflict.record_id}:${conflict.field_key}`;
}

export function workbookConflictEntry({
  conflict,
  focusKey = null,
  rowLabel,
  surfaceLabel,
  viewSchemaId,
  sheetRef,
}: {
  readonly conflict: WorkbookSameFieldConflictPayload;
  readonly focusKey?: string | null | undefined;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
  readonly sheetRef?: SheetRef | undefined;
}): WorkbookConflictEntry {
  return {
    key: workbookConflictQueueKey(conflict),
    conflict,
    resolutionClass: conflict.conflict_resolution_class,
    origin: {
      viewSchemaId,
      surfaceLabel,
      rowLabel,
      sheetRef: sheetRef ?? { kind: "view_schema", id: viewSchemaId },
    },
    focusKey,
    localValue: conflict.client_value,
    mergedDraft:
      conflict.conflict_resolution_class === "text_compare_merge" &&
      typeof conflict.server_value === "string"
        ? conflict.server_value
        : "",
  };
}

function isWorkbookCollectionValue(
  value: unknown,
): value is WorkbookCollectionValue {
  if (!isRecord(value)) {
    return false;
  }
  return (
    value.kind === "collection_value_v1" &&
    typeof value.ordered === "boolean" &&
    Array.isArray(value.items) &&
    value.items.every(
      (item) =>
        item !== null && typeof item === "object" && !Array.isArray(item),
    )
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function collectionItemIdentity(item: Record<string, unknown>): string {
  if (typeof item.item_ref === "string" && item.item_ref !== "") {
    return item.item_ref;
  }
  return JSON.stringify(item);
}

function removeCollectionItemAction(
  item: Record<string, unknown>,
): WorkbookCollectionAction | null {
  if (typeof item.item_ref !== "string" || item.item_ref === "") {
    return null;
  }
  switch (item.item_kind) {
    case "alias":
      return { item_ref: item.item_ref, op: "remove_alias" };
    case "party_ref":
      return { item_ref: item.item_ref, op: "remove_party_ref" };
    case "record_ref":
    case "resolved_ref":
      return { item_ref: item.item_ref, op: "remove_record_ref" };
    case "risk_ref":
      return { item_ref: item.item_ref, op: "remove_risk_ref" };
    case "tag":
      return { item_ref: item.item_ref, op: "remove_tag" };
  }
  return null;
}

type WorkbookCollectionItemKind =
  | "alias"
  | "party_ref"
  | "record_ref"
  | "resolved_ref"
  | "risk_ref"
  | "tag"
  | "unresolved_mention";

function actionFromStringProperty(
  item: Record<string, unknown>,
  property: string,
  build: (value: string) => WorkbookCollectionAction,
): WorkbookCollectionAction | null {
  const value = item[property];
  return typeof value === "string" ? build(value) : null;
}

function resolvedReferenceAction(
  item: Record<string, unknown>,
): WorkbookCollectionAction | null {
  const recordId =
    typeof item.resolved_record_id === "string"
      ? item.resolved_record_id
      : item.linked_record_id;
  return typeof recordId === "string"
    ? { op: "add_record_ref", linked_record_id: recordId }
    : null;
}

const collectionItemActionBuilders = {
  alias: (item: Record<string, unknown>) =>
    actionFromStringProperty(item, "alias_text", (aliasText) => ({
      alias_text: aliasText,
      op: "add_alias",
    })),
  party_ref: (item: Record<string, unknown>) =>
    actionFromStringProperty(item, "party_id", (partyId) => ({
      op: "add_party_ref",
      party_id: partyId,
    })),
  record_ref: (item: Record<string, unknown>) =>
    actionFromStringProperty(item, "linked_record_id", (recordId) => ({
      linked_record_id: recordId,
      op: "add_record_ref",
    })),
  resolved_ref: resolvedReferenceAction,
  risk_ref: (item: Record<string, unknown>) =>
    actionFromStringProperty(item, "risk_ref_text", (riskRefText) => ({
      op: "add_risk_ref",
      risk_ref_text: riskRefText,
    })),
  tag: (item: Record<string, unknown>) =>
    actionFromStringProperty(item, "display_text", (tagName) => ({
      op: "add_tag",
      tag_name: tagName,
    })),
  unresolved_mention: (item: Record<string, unknown>) =>
    actionFromStringProperty(item, "raw_text", (rawText) => ({
      op: "add_token",
      raw_text: rawText,
    })),
} satisfies Readonly<
  Record<
    WorkbookCollectionItemKind,
    (item: Record<string, unknown>) => WorkbookCollectionAction | null
  >
>;

function isWorkbookCollectionItemKind(
  value: string,
): value is WorkbookCollectionItemKind {
  return Object.hasOwn(collectionItemActionBuilders, value);
}

function addCollectionItemAction(
  item: Record<string, unknown>,
): WorkbookCollectionAction | null {
  const itemKind = item.item_kind;
  return typeof itemKind === "string" && isWorkbookCollectionItemKind(itemKind)
    ? collectionItemActionBuilders[itemKind](item)
    : null;
}

export function collectionActionsAgainstSaved(
  savedValue: unknown,
  finalValue: unknown,
): WorkbookCollectionActions | null {
  if (
    !isWorkbookCollectionValue(savedValue) ||
    !isWorkbookCollectionValue(finalValue)
  ) {
    return null;
  }
  const finalIdentities = new Set(
    finalValue.items.map((item) => collectionItemIdentity(item)),
  );
  const savedIdentities = new Set(
    savedValue.items.map((item) => collectionItemIdentity(item)),
  );
  const actions: WorkbookCollectionAction[] = [
    ...savedValue.items
      .filter((item) => !finalIdentities.has(collectionItemIdentity(item)))
      .map(removeCollectionItemAction),
    ...finalValue.items
      .filter((item) => !savedIdentities.has(collectionItemIdentity(item)))
      .map(addCollectionItemAction),
  ].filter((action): action is WorkbookCollectionAction => action !== null);
  const [firstAction, ...remainingActions] = actions;
  return firstAction === undefined
    ? null
    : {
        actions: [firstAction, ...remainingActions],
        kind: "collection_actions_v1",
      };
}

export function buildWorkbookConflictResolutionPayload({
  clientTxnId,
  entry,
  resolutionKind,
}: {
  readonly clientTxnId: string;
  readonly entry: WorkbookConflictEntry;
  readonly resolutionKind: WorkbookConflictResolutionKind;
}): WorkbookProtocolResolveConflictRequest | null {
  const base = {
    conflict_token: entry.conflict.conflict_token,
    resolution_kind: resolutionKind,
    client_txn_id: clientTxnId,
  } satisfies Omit<WorkbookProtocolResolveConflictRequest, "resolved_value">;
  if (resolutionKind === "keep_saved") {
    return base;
  }
  if (resolutionKind === "use_unsaved") {
    return typeof entry.localValue === "string" || entry.localValue === null
      ? { ...base, resolved_value: entry.localValue }
      : null;
  }
  if (entry.resolutionClass === "collection_review") {
    const actions = collectionActionsAgainstSaved(
      entry.conflict.server_value,
      entry.localValue,
    );
    if (actions === null) {
      return null;
    }
    return { ...base, resolved_value: actions };
  }
  return { ...base, resolved_value: entry.mergedDraft };
}

export type SameFieldConflictFields = Record<string, unknown> & {
  base_row_version: number;
  conflict_resolution_class: WorkbookConflictResolutionClass;
  conflict_token: string;
  current_row_version: number;
  field_key: string;
  record_id: string;
};

export function parseSameFieldConflictFields(
  conflict: unknown,
): SameFieldConflictFields | null {
  if (!isRecord(conflict)) {
    return null;
  }
  const object = conflict;
  const resolutionClass = workbookConflictResolutionClass(
    object.conflict_resolution_class,
  );
  if (
    typeof object.conflict_token !== "string" ||
    object.conflict_token.trim() === "" ||
    typeof object.record_id !== "string" ||
    object.record_id.trim() === "" ||
    typeof object.field_key !== "string" ||
    object.field_key.trim() === "" ||
    resolutionClass === null ||
    typeof object.base_row_version !== "number" ||
    typeof object.current_row_version !== "number"
  ) {
    return null;
  }
  return {
    ...object,
    base_row_version: object.base_row_version,
    conflict_resolution_class: resolutionClass,
    conflict_token: object.conflict_token,
    current_row_version: object.current_row_version,
    field_key: object.field_key,
    record_id: object.record_id,
  };
}

export function parseSameFieldConflictPayload(
  conflict: unknown,
): WorkbookSameFieldConflictPayload | null {
  const object = parseSameFieldConflictFields(conflict);
  if (
    object === null ||
    !("client_value" in object) ||
    !("server_value" in object)
  ) {
    return null;
  }
  const parsed: WorkbookSameFieldConflictPayload = {
    conflict_token: object.conflict_token,
    record_id: object.record_id,
    field_key: object.field_key,
    conflict_resolution_class: object.conflict_resolution_class,
    base_row_version: object.base_row_version,
    current_row_version: object.current_row_version,
    client_value: object.client_value,
    server_value: object.server_value,
    base_value: object.base_value,
    suggested_merged_value: object.suggested_merged_value,
  };
  if (typeof object.server_updated_by === "string") {
    parsed.server_updated_by = object.server_updated_by;
  }
  if (typeof object.server_updated_at === "string") {
    parsed.server_updated_at = object.server_updated_at;
  }
  return parsed;
}

export function parseSameFieldConflict(
  payload: unknown,
): WorkbookSameFieldConflictPayload | null {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return null;
  }
  const error = payload.error;
  if (
    !error ||
    typeof error !== "object" ||
    !("code" in error) ||
    error.code !== "same_field_conflict" ||
    !("conflict" in error)
  ) {
    return null;
  }
  return parseSameFieldConflictPayload(error.conflict);
}
