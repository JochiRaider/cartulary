import type { MentionResolutionAction } from "../../collaboration/workbookCollaborationMessages";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "./timelineRowModel";
import type { InspectorMention } from "./workbookMentionChips";

export type TimelineMentionActionContext = {
  readonly authorized: boolean;
  readonly capabilityAvailable: boolean;
  readonly surfaceKey: string;
};

export type TimelineMentionSubject = {
  readonly itemRef: string;
  readonly rowRecordId: string;
  readonly surfaceKey: string;
};

export type TimelineMentionResolutionPlan =
  | {
      readonly kind: "dispatch";
      readonly mention: InspectorMention;
      readonly request: {
        readonly action: MentionResolutionAction;
        readonly baseMentionRowVersion: number;
        readonly expectedSourceRecordId: string;
        readonly mentionId: string;
        readonly resolvedRecordId?: string | undefined;
      };
      readonly row: WorkbookRow;
    }
  | {
      readonly kind: "reject";
      readonly reason:
        | "authorization_lost"
        | "capability_unavailable"
        | "surface_changed"
        | "source_missing"
        | "source_invalid"
        | "mention_missing"
        | "mention_version_missing"
        | "target_missing";
    };

export function timelineMentionSubject(
  mention: InspectorMention,
  surfaceKey: string,
): TimelineMentionSubject {
  return {
    itemRef: mention.itemRef,
    rowRecordId: mention.rowRecordId,
    surfaceKey,
  };
}

export function planTimelineMentionEntityCreation(options: {
  readonly context: TimelineMentionActionContext;
  readonly mention: InspectorMention;
  readonly rows: readonly WorkbookRow[];
  readonly subject: TimelineMentionSubject;
}): TimelineMentionResolutionPlan {
  if (options.mention.rawText.trim() === "") {
    return { kind: "reject", reason: "mention_missing" };
  }
  return planTimelineMentionResolution({
    action: "resolve_item",
    allowUnknownResolvedTarget: true,
    context: options.context,
    knownEntityTypes: new Map<string, "host" | "identity">(),
    mention: options.mention,
    resolvedRecordId: "pending-created-entity",
    rows: options.rows,
    subject: options.subject,
  });
}

export function planTimelineMentionResolution(options: {
  readonly action: MentionResolutionAction;
  readonly allowUnknownResolvedTarget?: boolean;
  readonly context: TimelineMentionActionContext;
  readonly knownEntityTypes: ReadonlyMap<string, "host" | "identity">;
  readonly mention: InspectorMention;
  readonly resolvedRecordId?: string | undefined;
  readonly rows: readonly WorkbookRow[];
  readonly subject: TimelineMentionSubject;
}): TimelineMentionResolutionPlan {
  const unavailable = actionAvailability(options.context, options.subject);
  if (unavailable !== null) return unavailable;
  const row = options.rows.find(
    (candidate) => candidate.recordId === options.subject.rowRecordId,
  );
  if (row === undefined) return { kind: "reject", reason: "source_missing" };
  if (
    row.viewSchemaId !== timelineViewSchemaId ||
    row.recordId === null ||
    row.rowVersion === null ||
    row.pendingSignature !== null
  ) {
    return { kind: "reject", reason: "source_invalid" };
  }
  const mention = currentMention(row, options.mention, options.action);
  if (mention === null || mention.itemRef !== options.subject.itemRef) {
    return { kind: "reject", reason: "mention_missing" };
  }
  const mentionId = entityMentionId(mention);
  if (mentionId === null || mention.mentionRowVersion === null) {
    return { kind: "reject", reason: "mention_version_missing" };
  }
  if (
    options.action === "resolve_item" &&
    (options.resolvedRecordId === undefined ||
      (!options.allowUnknownResolvedTarget &&
        options.knownEntityTypes.get(options.resolvedRecordId) !==
          mention.entityType))
  ) {
    return { kind: "reject", reason: "target_missing" };
  }
  return {
    kind: "dispatch",
    mention,
    request: {
      action: options.action,
      baseMentionRowVersion: mention.mentionRowVersion,
      expectedSourceRecordId: row.recordId,
      mentionId,
      ...(options.resolvedRecordId === undefined
        ? {}
        : { resolvedRecordId: options.resolvedRecordId }),
    },
    row,
  };
}

function actionAvailability(
  context: TimelineMentionActionContext,
  subject: TimelineMentionSubject,
): Extract<TimelineMentionResolutionPlan, { kind: "reject" }> | null {
  if (!context.authorized) {
    return { kind: "reject", reason: "authorization_lost" };
  }
  if (!context.capabilityAvailable) {
    return { kind: "reject", reason: "capability_unavailable" };
  }
  return context.surfaceKey === subject.surfaceKey
    ? null
    : { kind: "reject", reason: "surface_changed" };
}

function currentMention(
  row: WorkbookRow,
  original: InspectorMention,
  action: MentionResolutionAction,
): InspectorMention | null {
  const activeItem = [
    ...row.collectionValues.hostRefs,
    ...row.collectionValues.identityRefs,
  ].find((item) => item.itemRef === original.itemRef);
  if (activeItem === undefined) {
    return original.status === "dismissed" && action === "revert_to_unresolved"
      ? original
      : null;
  }
  return {
    ...original,
    anchor: {
      ...original.anchor,
      targetEntityRecordId: activeItem.resolvedRecordId,
    },
    autoResolved: activeItem.autoResolved,
    confidence: activeItem.confidence,
    displayText: activeItem.displayText,
    entityType: activeItem.entityType,
    matchedAliasText: activeItem.matchedAliasText,
    mentionRowVersion: activeItem.mentionRowVersion,
    provenance: activeItem.provenance,
    rawText: activeItem.rawText,
    resolutionMethod: activeItem.resolutionMethod,
    resolvedRecordId: activeItem.resolvedRecordId,
  };
}

function entityMentionId(mention: InspectorMention): string | null {
  if (
    typeof mention.anchor.entityMentionId === "string" &&
    mention.anchor.entityMentionId !== ""
  ) {
    return mention.anchor.entityMentionId;
  }
  return mention.itemRef.startsWith("entity_mention:")
    ? mention.itemRef.slice("entity_mention:".length) || null
    : null;
}
