import type {
  WorkbookRelationshipChipPresentation,
  WorkbookRelationshipChipState,
} from "../../models/workbookRelationshipChip";
import type { MentionSubject } from "../actions/timelineMentionOperationModel";

export type RelationshipFieldKey =
  | "timeline.host_refs"
  | "timeline.identity_refs";

type MentionChipAnchor = {
  recordId: string;
  fieldKey: RelationshipFieldKey;
  itemRef: string;
  entityMentionId: string | null;
  targetEntityRecordId: string | null;
};

export type CollectionItem = {
  entityMentionId: string | null;
  itemRef: string;
  entityType: "host" | "identity";
  itemKind: "resolved_ref" | "unresolved_mention";
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
  entityMentionId: string | null;
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
  chipState: WorkbookRelationshipChipState;
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
  displayText?: string;
  entityMentionId: string | null;
  mentionRowVersion: number | null;
  itemRef: string;
  rowRecordId: string;
  fieldKey: RelationshipFieldKey;
  entityType: "host" | "identity";
  rawText: string;
  resolvedRecordId: string;
  matchedAliasText: string | null;
};

function relationshipItemLabel(
  item: CollectionItem | InspectorMention,
  entityIndex: Record<string, { label: string }>,
) {
  if ("status" in item && item.status === "dismissed") {
    return item.displayText || item.rawText;
  }
  if (item.resolvedRecordId) {
    const resolvedEntity = entityIndex[item.resolvedRecordId];
    if (resolvedEntity) return resolvedEntity.label;
  }
  return item.displayText || item.rawText;
}

function mentionChipStateForItem(
  item: CollectionItem | InspectorMention,
): WorkbookRelationshipChipState {
  if ("status" in item && item.status === "dismissed") return "dismissed";
  if (item.resolvedRecordId === null) return "unresolved";
  if (item.autoResolved || item.resolutionMethod === "auto_match")
    return "auto_resolved";
  return "resolved";
}

export function timelineRelationshipChipPresentation({
  entityIndex,
  item,
  selected = false,
  sourceRecordId = null,
}: {
  readonly entityIndex: Record<string, { label: string }>;
  readonly item: CollectionItem | InspectorMention;
  readonly selected?: boolean;
  readonly sourceRecordId?: string | null;
}): WorkbookRelationshipChipPresentation {
  const state = mentionChipStateForItem(item);
  const previousTargetId =
    "priorTargetEntityRecordId" in item ? item.priorTargetEntityRecordId : null;
  return {
    entityType: item.entityType,
    source: {
      kind: "entity_mention",
      recordId: "rowRecordId" in item ? item.rowRecordId : sourceRecordId,
      fieldKey:
        "fieldKey" in item
          ? item.fieldKey
          : item.entityType === "host"
            ? "timeline.host_refs"
            : "timeline.identity_refs",
      itemRef: item.itemRef,
    },
    rawText: item.rawText,
    targetRecordId: state === "dismissed" ? null : item.resolvedRecordId,
    previousTarget: previousTargetId
      ? {
          recordId: previousTargetId,
          label: entityIndex[previousTargetId]?.label ?? item.displayText,
        }
      : null,
    resolution: {
      method:
        state === "unresolved"
          ? null
          : state === "auto_resolved" || item.resolutionMethod === "auto_match"
            ? "auto"
            : item.resolutionMethod === "explicit_resolve_route" ||
                item.resolutionMethod === "resolve_item" ||
                item.resolutionMethod === "add_resolved_ref"
              ? "manual"
              : item.resolutionMethod === "legacy_import"
                ? "import"
                : item.resolutionMethod === "system"
                  ? "system"
                  : null,
      sourceMethod: item.resolutionMethod,
      provenance: item.provenance,
      confidence: item.confidence,
      matchedAliasText: item.matchedAliasText,
    },
    label:
      state === "unresolved" || state === "dismissed"
        ? item.rawText
        : relationshipItemLabel(item, entityIndex),
    selected,
    selectorIdentity: item.itemRef,
    state,
  };
}

type TimelineApiRowLike = {
  cells: Record<string, { value: unknown }>;
};

type MentionCollectionRowLike = {
  recordId: string | null;
  rowVersion?: number | null;
  collectionValues: {
    hostRefs: CollectionItem[];
    identityRefs: CollectionItem[];
  };
};

function parseEntityType(value: unknown): "host" | "identity" | null {
  return value === "host" || value === "identity" ? value : null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

type CollectionItemIdentity = Pick<
  CollectionItem,
  "entityMentionId" | "entityType" | "itemKind" | "itemRef"
>;

function decodeCollectionItemIdentity(
  value: Readonly<Record<string, unknown>>,
): CollectionItemIdentity | null {
  const itemRef = value.item_ref;
  const entityType = parseEntityType(value.entity_type);
  const itemKind = value.item_kind;
  if (
    typeof itemRef !== "string" ||
    itemRef.trim() === "" ||
    entityType === null ||
    (itemKind !== "resolved_ref" && itemKind !== "unresolved_mention")
  ) {
    return null;
  }
  return {
    entityMentionId:
      typeof value.entity_mention_id === "string" &&
      value.entity_mention_id !== ""
        ? value.entity_mention_id
        : null,
    entityType,
    itemKind,
    itemRef,
  };
}

function collectionItemText(value: Readonly<Record<string, unknown>>): {
  readonly displayText: string;
  readonly rawText: string;
} {
  const rawText = typeof value.raw_text === "string" ? value.raw_text : "";
  return {
    displayText:
      typeof value.display_text === "string" ? value.display_text : rawText,
    rawText,
  };
}

function collectionItemResolvedRecordId(
  value: Readonly<Record<string, unknown>>,
): string | null {
  const resolvedRecordId =
    typeof value.resolved_record_id === "string" &&
    value.resolved_record_id.trim() !== ""
      ? value.resolved_record_id
      : null;
  return resolvedRecordId;
}

function collectionItemTargetIsValid(
  identity: CollectionItemIdentity,
  rawText: string,
  resolvedRecordId: string | null,
): boolean {
  if (identity.itemKind === "resolved_ref") {
    return resolvedRecordId !== null;
  }
  return rawText.trim() !== "" && resolvedRecordId === null;
}

function collectionItemMetadata(
  value: Readonly<Record<string, unknown>>,
): Pick<
  CollectionItem,
  | "autoResolved"
  | "confidence"
  | "matchedAliasText"
  | "mentionRowVersion"
  | "provenance"
  | "resolutionMethod"
> {
  const confidence = value.confidence;
  const mentionRowVersion = value.mention_row_version;
  return {
    autoResolved: value.auto_resolved === true,
    confidence:
      typeof confidence === "number" && Number.isFinite(confidence)
        ? confidence
        : null,
    matchedAliasText:
      typeof value.matched_alias_text === "string"
        ? value.matched_alias_text
        : null,
    mentionRowVersion:
      typeof mentionRowVersion === "number" &&
      Number.isSafeInteger(mentionRowVersion) &&
      mentionRowVersion > 0
        ? mentionRowVersion
        : null,
    provenance: typeof value.provenance === "string" ? value.provenance : null,
    resolutionMethod:
      typeof value.resolution_method === "string"
        ? value.resolution_method
        : null,
  };
}

function decodeCollectionItem(value: unknown): CollectionItem | null {
  if (!isRecord(value)) return null;
  const identity = decodeCollectionItemIdentity(value);
  if (identity === null) return null;
  const text = collectionItemText(value);
  const resolvedRecordId = collectionItemResolvedRecordId(value);
  if (!collectionItemTargetIsValid(identity, text.rawText, resolvedRecordId)) {
    return null;
  }
  return {
    ...collectionItemMetadata(value),
    ...identity,
    ...text,
    resolvedRecordId,
  };
}

function mentionChipAnchor({
  entityMentionId,
  fieldKey,
  itemRef,
  recordId,
  targetEntityRecordId,
}: {
  fieldKey: RelationshipFieldKey;
  itemRef: string;
  recordId: string;
  targetEntityRecordId: string | null;
  entityMentionId: string | null;
}): MentionChipAnchor {
  return {
    recordId,
    fieldKey,
    itemRef,
    entityMentionId,
    targetEntityRecordId,
  };
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
    entityMentionId: item.entityMentionId,
    entityType: item.entityType,
    itemRef: item.itemRef,
    rawText: item.rawText,
    resolvedRecordId: targetEntityRecordId,
    mentionRowVersion: item.mentionRowVersion,
    resolutionMethod: item.resolutionMethod,
    autoResolved: item.autoResolved,
    status,
    chipState: mentionChipStateForItem(item),
    anchor: mentionChipAnchor({
      entityMentionId: item.entityMentionId,
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
    .map(decodeCollectionItem)
    .filter((item): item is CollectionItem => item !== null);
}

export function buildInspectorMentions(
  row: MentionCollectionRowLike | undefined,
  dismissedMentions: DismissedMention[],
  observedMentions: readonly MentionSubject[] = [],
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
  for (const observed of observedMentions) {
    if (observed.sourceRecordId !== row.recordId) continue;
    const index = activeMentions.findIndex(
      (item) => item.entityMentionId === observed.mentionId,
    );
    if (index < 0) {
      // A restore receipt can precede the projection that puts the active item back.
      // A newer source projection is decisive about removal from its active collection.
      if (
        observed.state !== "dismissed" &&
        row.rowVersion !== undefined &&
        row.rowVersion !== null &&
        row.rowVersion < observed.sourceRowVersion
      )
        activeMentions.push(
          activeInspectorMention(row.recordId, observed.sourceFieldKey, {
            entityMentionId: observed.mentionId,
            itemRef: observed.itemRef,
            entityType: observed.entityType,
            itemKind:
              observed.state === "resolved"
                ? "resolved_ref"
                : "unresolved_mention",
            rawText: observed.rawText,
            displayText: observed.rawText,
            resolvedRecordId: observed.resolvedRecordId,
            resolutionMethod: observed.resolutionMethod,
            mentionRowVersion: observed.mentionRowVersion,
            autoResolved: observed.resolutionMethod === "auto_match",
            provenance: null,
            confidence: null,
            matchedAliasText: null,
          }),
        );
      continue;
    }
    const current = activeMentions[index];
    if (
      !current ||
      (current.mentionRowVersion ?? 0) >= observed.mentionRowVersion
    )
      continue;
    if (observed.state === "dismissed") {
      activeMentions.splice(index, 1);
      continue;
    }
    activeMentions[index] = {
      ...current,
      rawText: observed.rawText,
      displayText: observed.rawText,
      status: observed.state,
      resolvedRecordId: observed.resolvedRecordId,
      resolutionMethod: observed.resolutionMethod,
      mentionRowVersion: observed.mentionRowVersion,
      autoResolved: observed.resolutionMethod === "auto_match",
      provenance: null,
      confidence: null,
      matchedAliasText: null,
      anchor: {
        ...current.anchor,
        targetEntityRecordId: observed.resolvedRecordId,
      },
      chipState: observed.state === "unresolved" ? "unresolved" : "resolved",
    };
  }
  const dismissed: InspectorMention[] = dismissedMentions
    .filter(
      (item) =>
        !activeMentions.some(
          (active) =>
            active.entityMentionId === item.entityMentionId &&
            (active.mentionRowVersion ?? 0) > (item.mentionRowVersion ?? 0),
        ),
    )
    .map((item) => ({
      ...item,
      status: "dismissed",
      chipState: "dismissed",
      anchor: mentionChipAnchor({
        entityMentionId: item.entityMentionId,
        recordId: item.rowRecordId,
        fieldKey: item.fieldKey,
        itemRef: item.itemRef,
        targetEntityRecordId: null,
      }),
      sourceKind: "entity_mention",
      isActiveRelationshipValue: false,
      priorTargetEntityRecordId: null,
      resolutionMethod: null,
      autoResolved: false,
      resolvedRecordId: null,
      mentionRowVersion: item.mentionRowVersion,
      displayText: item.rawText,
      provenance: null,
      confidence: null,
      matchedAliasText: null,
    }));

  return [
    ...activeMentions.filter(
      (active) =>
        !dismissed.some(
          (item) => item.entityMentionId === active.entityMentionId,
        ),
    ),
    ...dismissed,
  ];
}

export function buildAutoResolutionNotices(
  beforeRow: MentionCollectionRowLike | undefined,
  afterRow: MentionCollectionRowLike,
): AutoResolutionNotice[] {
  if (afterRow.recordId === null || (!beforeRow && afterRow.rowVersion !== 1)) {
    return [];
  }
  const beforeRefs = new Set(
    [
      ...(beforeRow?.collectionValues.hostRefs ?? []),
      ...(beforeRow?.collectionValues.identityRefs ?? []),
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
      entityMentionId: item.entityMentionId,
      mentionRowVersion: item.mentionRowVersion,
      itemRef: item.itemRef,
      rowRecordId: afterRow.recordId ?? "",
      fieldKey,
      entityType: item.entityType,
      rawText: item.rawText,
      resolvedRecordId: item.resolvedRecordId ?? "",
      matchedAliasText: item.matchedAliasText,
      displayText: item.displayText,
    }));
}
