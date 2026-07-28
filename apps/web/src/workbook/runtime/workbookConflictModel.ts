export type WorkbookSameFieldConflictPayload = {
  conflict_token: string;
  record_id: string;
  field_key: string;
  conflict_resolution_class: string;
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

export type WorkbookCollectionValue = {
  readonly kind: "collection_value_v1";
  readonly ordered: boolean;
  readonly items: readonly Record<string, unknown>[];
};

export type WorkbookCollectionActions = {
  readonly kind: "collection_actions_v1";
  readonly actions: readonly Record<string, unknown>[];
};

export type WorkbookConflictEntry = {
  readonly key: string;
  readonly conflict: WorkbookSameFieldConflictPayload;
  readonly resolutionClass: WorkbookConflictResolutionClass;
  readonly origin: {
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
  value: string,
): WorkbookConflictResolutionClass {
  switch (value) {
    case "collection_review":
    case "text_compare_merge":
      return value;
    default:
      return "atomic_replace";
  }
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
}: {
  readonly conflict: WorkbookSameFieldConflictPayload;
  readonly focusKey?: string | null | undefined;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
}): WorkbookConflictEntry {
  return {
    key: workbookConflictQueueKey(conflict),
    conflict,
    resolutionClass: workbookConflictResolutionClass(
      conflict.conflict_resolution_class,
    ),
    origin: { viewSchemaId, surfaceLabel, rowLabel },
    focusKey,
    localValue: conflict.client_value,
    mergedDraft:
      workbookConflictResolutionClass(conflict.conflict_resolution_class) ===
        "text_compare_merge" && typeof conflict.server_value === "string"
        ? conflict.server_value
        : "",
  };
}

export function isWorkbookCollectionValue(
  value: unknown,
): value is WorkbookCollectionValue {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const candidate = value as Record<string, unknown>;
  return (
    candidate.kind === "collection_value_v1" &&
    typeof candidate.ordered === "boolean" &&
    Array.isArray(candidate.items) &&
    candidate.items.every(
      (item) =>
        item !== null && typeof item === "object" && !Array.isArray(item),
    )
  );
}

function collectionItemIdentity(item: Record<string, unknown>): string {
  if (typeof item.item_ref === "string" && item.item_ref !== "") {
    return item.item_ref;
  }
  return JSON.stringify(item);
}

function removeCollectionItemAction(
  item: Record<string, unknown>,
): Record<string, unknown> | null {
  if (typeof item.item_ref !== "string" || item.item_ref === "") {
    return null;
  }
  const op = (() => {
    switch (item.item_kind) {
      case "alias":
        return "remove_alias";
      case "party_ref":
        return "remove_party_ref";
      case "record_ref":
      case "resolved_ref":
        return "remove_record_ref";
      case "risk_ref":
        return "remove_risk_ref";
      case "tag":
        return "remove_tag";
      default:
        return null;
    }
  })();
  return op === null ? null : { op, item_ref: item.item_ref };
}

function addCollectionItemAction(
  item: Record<string, unknown>,
): Record<string, unknown> | null {
  switch (item.item_kind) {
    case "alias":
      return typeof item.alias_text === "string"
        ? { op: "add_alias", alias_text: item.alias_text }
        : null;
    case "party_ref":
      return typeof item.party_id === "string"
        ? { op: "add_party_ref", party_id: item.party_id }
        : null;
    case "record_ref":
      return typeof item.linked_record_id === "string"
        ? { op: "add_record_ref", linked_record_id: item.linked_record_id }
        : null;
    case "resolved_ref": {
      const recordId =
        typeof item.resolved_record_id === "string"
          ? item.resolved_record_id
          : item.linked_record_id;
      return typeof recordId === "string"
        ? { op: "add_record_ref", linked_record_id: recordId }
        : null;
    }
    case "risk_ref":
      return typeof item.risk_ref_text === "string"
        ? { op: "add_risk_ref", risk_ref_text: item.risk_ref_text }
        : null;
    case "tag":
      return typeof item.display_text === "string"
        ? { op: "add_tag", tag_name: item.display_text }
        : null;
    case "unresolved_mention":
      return typeof item.raw_text === "string"
        ? { op: "add_token", raw_text: item.raw_text }
        : null;
    default:
      return null;
  }
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
  const actions = [
    ...savedValue.items
      .filter((item) => !finalIdentities.has(collectionItemIdentity(item)))
      .map(removeCollectionItemAction),
    ...finalValue.items
      .filter((item) => !savedIdentities.has(collectionItemIdentity(item)))
      .map(addCollectionItemAction),
  ].filter((action): action is Record<string, unknown> => action !== null);
  return actions.length === 0
    ? null
    : { kind: "collection_actions_v1", actions };
}

export function buildWorkbookConflictResolutionPayload({
  clientTxnId,
  entry,
  resolutionKind,
}: {
  readonly clientTxnId: string;
  readonly entry: WorkbookConflictEntry;
  readonly resolutionKind: WorkbookConflictResolutionKind;
}): Record<string, unknown> | null {
  const body: Record<string, unknown> = {
    conflict_token: entry.conflict.conflict_token,
    resolution_kind: resolutionKind,
    client_txn_id: clientTxnId,
  };
  if (resolutionKind === "keep_saved") {
    return body;
  }
  if (resolutionKind === "use_unsaved") {
    body.resolved_value = entry.localValue;
    return body;
  }
  if (entry.resolutionClass === "collection_review") {
    const actions = collectionActionsAgainstSaved(
      entry.conflict.server_value,
      entry.localValue,
    );
    if (actions === null) {
      return null;
    }
    body.resolved_value = actions;
    return body;
  }
  body.resolved_value = entry.mergedDraft;
  return body;
}

export type SameFieldConflictFields = Record<string, unknown> & {
  base_row_version: number;
  conflict_resolution_class: string;
  conflict_token: string;
  current_row_version: number;
  field_key: string;
  record_id: string;
};

export function parseSameFieldConflictFields(
  conflict: unknown,
): SameFieldConflictFields | null {
  if (!conflict || typeof conflict !== "object") {
    return null;
  }
  const object = conflict as Record<string, unknown>;
  if (
    typeof object.conflict_token !== "string" ||
    object.conflict_token.trim() === "" ||
    typeof object.record_id !== "string" ||
    object.record_id.trim() === "" ||
    typeof object.field_key !== "string" ||
    object.field_key.trim() === "" ||
    typeof object.conflict_resolution_class !== "string" ||
    object.conflict_resolution_class.trim() === "" ||
    typeof object.base_row_version !== "number" ||
    typeof object.current_row_version !== "number"
  ) {
    return null;
  }
  return object as SameFieldConflictFields;
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
  const object = parseSameFieldConflictFields(error.conflict);
  if (object === null) {
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
