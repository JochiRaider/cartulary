const timelineViewSchemaId = "cartulary.view.timeline.v1";

export type RelationshipFieldKey =
  | "timeline.host_refs"
  | "timeline.identity_refs";

export type CollectionItem = {
  itemRef: string;
  entityType: "host" | "identity";
  itemKind: "resolved_ref" | "unresolved_mention" | string;
  displayText: string;
  rawText: string;
  resolvedRecordId: string | null;
  resolutionMethod: string | null;
  autoResolved: boolean;
  provenance: string | null;
  confidence: number | null;
  matchedAliasText: string | null;
};

export type DismissedMention = {
  rowRecordId: string;
  fieldKey: RelationshipFieldKey;
  entityType: "host" | "identity";
  itemRef: string;
  rawText: string;
  resolvedRecordId: string | null;
  resolutionMethod: string | null;
  autoResolved: boolean;
};

export type InspectorMention = DismissedMention & {
  status: "unresolved" | "resolved" | "dismissed";
  displayText: string;
  provenance: string | null;
  confidence: number | null;
  matchedAliasText: string | null;
};

export type AutoResolutionNotice = {
  itemRef: string;
  rowRecordId: string;
  fieldKey: RelationshipFieldKey;
  entityType: "host" | "identity";
  rawText: string;
  resolvedRecordId: string;
  matchedAliasText: string | null;
};

export type RecordChangedPayload = {
  record_id: string;
  row_version: number;
  change_set_id: string;
  client_txn_id: string;
  actor_user_id: string;
  changed_field_keys: string[];
  affected_views: Array<{
    patch_cells?: {
      record_id: string;
      row_version: number;
      cells: Record<string, { value: unknown }>;
      group_values?: Record<string, unknown>;
    };
    view_schema_id: string;
    change_kind: string;
  }>;
};

type CollaborationMessage = {
  type: string;
  payload?: unknown;
};

type TimelineApiRowLike = {
  cells: Record<string, { value: unknown }>;
};

type MentionPatchRowLike = {
  rowVersion: number | null;
};

type MentionPatchMentionLike = {
  itemRef: string;
  fieldKey: RelationshipFieldKey;
};

type MentionCollectionRowLike = {
  recordId: string | null;
  collectionValues: {
    hostRefs: CollectionItem[];
    identityRefs: CollectionItem[];
  };
};

function safeEntityType(value: unknown): "host" | "identity" {
  return value === "identity" ? "identity" : "host";
}

export function isRecordChangedMessage(
  message: unknown,
): message is { type: "record_changed"; payload: RecordChangedPayload } {
  if (!message || typeof message !== "object") {
    return false;
  }
  const candidate = message as CollaborationMessage;
  if (candidate.type !== "record_changed") {
    return false;
  }

  const payload = candidate.payload;
  if (!payload || typeof payload !== "object") {
    return false;
  }

  return (
    "client_txn_id" in payload && typeof payload.client_txn_id === "string"
  );
}

export function shouldIgnoreSelfOriginatedRecordChange(
  message: unknown,
  resolvePendingSocketTxn: (clientTxnId: string | null | undefined) => boolean,
): boolean {
  if (!isRecordChangedMessage(message)) {
    return false;
  }
  return resolvePendingSocketTxn(message.payload.client_txn_id);
}

export function readCollectionItems(
  row: TimelineApiRowLike,
  fieldKey: RelationshipFieldKey,
): CollectionItem[] {
  const raw = row.cells[fieldKey]?.value;
  const value =
    raw &&
    typeof raw === "object" &&
    !Array.isArray(raw) &&
    "items" in raw &&
    Array.isArray(raw.items)
      ? raw.items
      : [];
  return value
    .map((item) => {
      if (!item || typeof item !== "object") {
        return null;
      }
      const object = item as Record<string, unknown>;
      const confidenceValue = object.confidence;
      return {
        itemRef:
          typeof object.item_ref === "string" ? object.item_ref : "unknown",
        entityType: safeEntityType(object.entity_type),
        itemKind:
          typeof object.item_kind === "string"
            ? object.item_kind
            : "unresolved_mention",
        displayText:
          typeof object.display_text === "string"
            ? object.display_text
            : typeof object.raw_text === "string"
              ? object.raw_text
              : "",
        rawText: typeof object.raw_text === "string" ? object.raw_text : "",
        resolvedRecordId:
          typeof object.resolved_record_id === "string"
            ? object.resolved_record_id
            : null,
        resolutionMethod:
          typeof object.resolution_method === "string"
            ? object.resolution_method
            : null,
        autoResolved: object.auto_resolved === true,
        provenance:
          typeof object.provenance === "string" ? object.provenance : null,
        confidence:
          typeof confidenceValue === "number" ? confidenceValue : null,
        matchedAliasText:
          typeof object.matched_alias_text === "string"
            ? object.matched_alias_text
            : null,
      } satisfies CollectionItem;
    })
    .filter((item): item is CollectionItem => item !== null);
}

export function buildMentionPatchPayload(
  row: MentionPatchRowLike,
  mention: MentionPatchMentionLike,
  action: "resolve_item" | "dismiss_item" | "revert_to_unresolved",
  clientTxnId: string,
  resolvedRecordId?: string,
) {
  if (row.rowVersion === null) {
    return null;
  }

  const actionEntry: Record<string, string> = {
    op: action,
    item_ref: mention.itemRef,
  };
  if (resolvedRecordId) {
    actionEntry.resolved_record_id = resolvedRecordId;
  }
  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes: [
      {
        field_key: mention.fieldKey,
        action_payload: {
          kind: "collection_actions_v1",
          actions: [actionEntry],
        },
      },
    ],
  };
}

export function buildInspectorMentions(
  row: MentionCollectionRowLike | undefined,
  dismissedMentions: DismissedMention[],
): InspectorMention[] {
  if (!row || row.recordId === null) {
    return [];
  }

  const activeMentions: InspectorMention[] = [
    ...row.collectionValues.hostRefs.map<InspectorMention>((item) => ({
      rowRecordId: row.recordId ?? "",
      fieldKey: "timeline.host_refs",
      entityType: item.entityType,
      itemRef: item.itemRef,
      rawText: item.rawText,
      resolvedRecordId: item.resolvedRecordId,
      resolutionMethod: item.resolutionMethod,
      autoResolved: item.autoResolved,
      status: item.itemKind === "resolved_ref" ? "resolved" : "unresolved",
      displayText: item.displayText,
      provenance: item.provenance,
      confidence: item.confidence,
      matchedAliasText: item.matchedAliasText,
    })),
    ...row.collectionValues.identityRefs.map<InspectorMention>((item) => ({
      rowRecordId: row.recordId ?? "",
      fieldKey: "timeline.identity_refs",
      entityType: item.entityType,
      itemRef: item.itemRef,
      rawText: item.rawText,
      resolvedRecordId: item.resolvedRecordId,
      resolutionMethod: item.resolutionMethod,
      autoResolved: item.autoResolved,
      status: item.itemKind === "resolved_ref" ? "resolved" : "unresolved",
      displayText: item.displayText,
      provenance: item.provenance,
      confidence: item.confidence,
      matchedAliasText: item.matchedAliasText,
    })),
  ];
  const dismissed: InspectorMention[] = dismissedMentions.map((item) => ({
    ...item,
    status: "dismissed",
    displayText: item.rawText,
    provenance: null,
    confidence: null,
    matchedAliasText: null,
  }));

  return [...activeMentions, ...dismissed];
}

export function buildAutoResolutionNotices(
  beforeRow: MentionCollectionRowLike | undefined,
  afterRow: MentionCollectionRowLike,
): AutoResolutionNotice[] {
  if (!beforeRow || afterRow.recordId === null) {
    return [];
  }
  const beforeRefs = new Set(
    [
      ...beforeRow.collectionValues.hostRefs,
      ...beforeRow.collectionValues.identityRefs,
    ].map((item) => item.itemRef),
  );
  const newItems = [
    ...afterRow.collectionValues.hostRefs.map((item) => ({
      fieldKey: "timeline.host_refs" as const,
      item,
    })),
    ...afterRow.collectionValues.identityRefs.map((item) => ({
      fieldKey: "timeline.identity_refs" as const,
      item,
    })),
  ];
  return newItems
    .filter(
      ({ item }) =>
        !beforeRefs.has(item.itemRef) &&
        item.autoResolved &&
        item.resolvedRecordId !== null,
    )
    .map(({ fieldKey, item }) => ({
      itemRef: item.itemRef,
      rowRecordId: afterRow.recordId ?? "",
      fieldKey,
      entityType: item.entityType,
      rawText: item.rawText,
      resolvedRecordId: item.resolvedRecordId ?? "",
      matchedAliasText: item.matchedAliasText,
    }));
}
