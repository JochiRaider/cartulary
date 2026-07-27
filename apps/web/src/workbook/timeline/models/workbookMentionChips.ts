export type RelationshipFieldKey =
  | "timeline.host_refs"
  | "timeline.identity_refs";

export type MentionChipState =
  | "unresolved"
  | "resolved"
  | "auto_resolved"
  | "dismissed";

export type MentionChipAnchor = {
  recordId: string;
  fieldKey: RelationshipFieldKey;
  itemRef: string;
  entityMentionId: string | null;
  targetEntityRecordId: string | null;
};

export type CollectionItem = {
  itemRef: string;
  entityType: "host" | "identity";
  itemKind: "resolved_ref" | "unresolved_mention" | string;
  displayText: string;
  rawText: string;
  resolvedRecordId: string | null;
  mentionRowVersion: number | null;
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
  mentionRowVersion: number | null;
  resolutionMethod: string | null;
  autoResolved: boolean;
  displayText?: string;
  priorTargetEntityRecordId?: string | null;
  provenance?: string | null;
  confidence?: number | null;
  matchedAliasText?: string | null;
};

export type InspectorMention = DismissedMention & {
  status: "unresolved" | "resolved" | "dismissed";
  chipState: MentionChipState;
  anchor: MentionChipAnchor;
  sourceKind: "entity_mention";
  isActiveRelationshipValue: boolean;
  priorTargetEntityRecordId: string | null;
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

type TimelineApiRowLike = {
  cells: Record<string, { value: unknown }>;
};

type MentionCollectionRowLike = {
  recordId: string | null;
  collectionValues: {
    hostRefs: CollectionItem[];
    identityRefs: CollectionItem[];
  };
};

export function reconcileDismissedMentionsForRow(
  dismissedMentionsByRow: Record<string, DismissedMention[]>,
  row: MentionCollectionRowLike,
) {
  if (row.recordId === null) {
    return dismissedMentionsByRow;
  }

  const activeItems = new Map(
    [
      ...row.collectionValues.hostRefs,
      ...row.collectionValues.identityRefs,
    ].map((item) => [item.itemRef, item]),
  );
  const current = dismissedMentionsByRow[row.recordId] ?? [];
  const remaining = current.filter((dismissed) => {
    const active = activeItems.get(dismissed.itemRef);
    if (active === undefined) {
      return true;
    }
    if (dismissed.mentionRowVersion === null) {
      return false;
    }
    return (
      active.mentionRowVersion === null ||
      active.mentionRowVersion <= dismissed.mentionRowVersion
    );
  });
  if (remaining.length === current.length) {
    return dismissedMentionsByRow;
  }

  const next = { ...dismissedMentionsByRow };
  if (remaining.length < 1) {
    delete next[row.recordId];
  } else {
    next[row.recordId] = remaining;
  }
  return next;
}

function safeEntityType(value: unknown): "host" | "identity" {
  return value === "identity" ? "identity" : "host";
}

function entityMentionIdFromItemRef(itemRef: string): string | null {
  const prefix = "entity_mention:";
  if (!itemRef.startsWith(prefix)) {
    return null;
  }
  const mentionId = itemRef.slice(prefix.length);
  return mentionId === "" ? null : mentionId;
}

function mentionChipAnchor({
  fieldKey,
  itemRef,
  recordId,
  targetEntityRecordId,
}: {
  fieldKey: RelationshipFieldKey;
  itemRef: string;
  recordId: string;
  targetEntityRecordId: string | null;
}): MentionChipAnchor {
  return {
    recordId,
    fieldKey,
    itemRef,
    entityMentionId: entityMentionIdFromItemRef(itemRef),
    targetEntityRecordId,
  };
}

function mentionChipState({
  autoResolved,
  itemKind,
  status,
}: {
  autoResolved: boolean;
  itemKind: string;
  status: "unresolved" | "resolved" | "dismissed";
}): MentionChipState {
  if (status === "dismissed") {
    return "dismissed";
  }
  if (status === "unresolved" || itemKind !== "resolved_ref") {
    return "unresolved";
  }
  if (autoResolved) {
    return "auto_resolved";
  }
  return "resolved";
}

function activeInspectorMention(
  rowRecordId: string,
  fieldKey: RelationshipFieldKey,
  item: CollectionItem,
): InspectorMention {
  const status = item.itemKind === "resolved_ref" ? "resolved" : "unresolved";
  const targetEntityRecordId =
    status === "resolved" ? item.resolvedRecordId : null;
  return {
    rowRecordId,
    fieldKey,
    entityType: item.entityType,
    itemRef: item.itemRef,
    rawText: item.rawText,
    resolvedRecordId: targetEntityRecordId,
    mentionRowVersion: item.mentionRowVersion,
    resolutionMethod: item.resolutionMethod,
    autoResolved: item.autoResolved,
    status,
    chipState: mentionChipState({
      autoResolved: item.autoResolved,
      itemKind: item.itemKind,
      status,
    }),
    anchor: mentionChipAnchor({
      recordId: rowRecordId,
      fieldKey,
      itemRef: item.itemRef,
      targetEntityRecordId,
    }),
    sourceKind: "entity_mention",
    isActiveRelationshipValue: true,
    priorTargetEntityRecordId: null,
    displayText: item.displayText,
    provenance: item.provenance,
    confidence: item.confidence,
    matchedAliasText: item.matchedAliasText,
  };
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
      const mentionRowVersion = object.mention_row_version;
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
        mentionRowVersion:
          typeof mentionRowVersion === "number" ? mentionRowVersion : null,
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

export function buildInspectorMentions(
  row: MentionCollectionRowLike | undefined,
  dismissedMentions: DismissedMention[],
): InspectorMention[] {
  if (!row || row.recordId === null) {
    return [];
  }

  const activeMentions: InspectorMention[] = [
    ...row.collectionValues.hostRefs.map<InspectorMention>((item) =>
      activeInspectorMention(row.recordId ?? "", "timeline.host_refs", item),
    ),
    ...row.collectionValues.identityRefs.map<InspectorMention>((item) =>
      activeInspectorMention(
        row.recordId ?? "",
        "timeline.identity_refs",
        item,
      ),
    ),
  ];
  const dismissed: InspectorMention[] = dismissedMentions.map((item) => ({
    ...item,
    status: "dismissed",
    chipState: "dismissed",
    anchor: mentionChipAnchor({
      recordId: item.rowRecordId,
      fieldKey: item.fieldKey,
      itemRef: item.itemRef,
      targetEntityRecordId: null,
    }),
    sourceKind: "entity_mention",
    isActiveRelationshipValue: false,
    priorTargetEntityRecordId:
      item.priorTargetEntityRecordId ?? item.resolvedRecordId,
    resolvedRecordId: null,
    mentionRowVersion: item.mentionRowVersion,
    displayText: item.displayText ?? item.rawText,
    provenance: item.provenance ?? null,
    confidence: item.confidence ?? null,
    matchedAliasText: item.matchedAliasText ?? null,
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
