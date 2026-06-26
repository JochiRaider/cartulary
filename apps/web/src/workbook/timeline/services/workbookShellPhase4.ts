import type { RelationshipFieldKey } from "../models/workbookMentionChips";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

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

type MentionPatchRowLike = {
  rowVersion: number | null;
};

type MentionPatchMentionLike = {
  itemRef: string;
  fieldKey: RelationshipFieldKey;
};

type MentionActionMentionLike = {
  mentionRowVersion: number | null;
};

export type MentionResolutionAction =
  | "resolve_item"
  | "dismiss_item"
  | "revert_to_unresolved";

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

export function buildMentionPatchPayload(
  row: MentionPatchRowLike,
  mention: MentionPatchMentionLike,
  action: MentionResolutionAction,
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

export function buildMentionActionPayload(
  mention: MentionActionMentionLike,
  action: MentionResolutionAction,
  clientTxnId: string,
  resolvedRecordId?: string,
) {
  if (mention.mentionRowVersion === null) {
    return null;
  }
  const body: Record<string, string | number> = {
    base_mention_row_version: mention.mentionRowVersion,
    client_txn_id: clientTxnId,
    action,
  };
  if (action === "resolve_item" && resolvedRecordId !== undefined) {
    body.resolved_record_id = resolvedRecordId;
  }
  return body;
}
