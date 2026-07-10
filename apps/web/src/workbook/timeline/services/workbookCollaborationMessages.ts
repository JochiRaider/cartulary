import type { WorkbookSheetRef } from "../../models/workbookStartup";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type {
  WorkbookPresenceInput,
  WorkbookPresenceMode,
} from "../../utils/workbookPresence";

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

export type TimelinePresenceDraft = {
  readonly fieldKey: string | null;
  readonly mode: WorkbookPresenceMode;
  readonly recordId: string | null;
};

export type WorkbookPresenceUpdateMessage = {
  readonly type: "presence_update";
  readonly payload: {
    readonly presence: WorkbookPresenceInput;
  };
};

export type WorkbookSocketSessionMessage =
  | {
      readonly type: "hello";
      readonly payload: {
        readonly client_instance_id: string;
        readonly presence: WorkbookPresenceInput;
      };
    }
  | {
      readonly type: "resume";
      readonly payload: {
        readonly client_instance_id: string;
        readonly last_seen_stream_seq: number;
        readonly presence: WorkbookPresenceInput;
        readonly resume_token: string;
      };
    };

type MentionActionMentionLike = {
  mentionRowVersion: number | null;
};

export type MentionResolutionAction =
  | "resolve_item"
  | "dismiss_item"
  | "revert_to_unresolved";

export function buildWorkbookPresenceInput(
  presence: TimelinePresenceDraft = {
    fieldKey: null,
    mode: "viewing",
    recordId: null,
  },
  sheetRef: WorkbookSheetRef = {
    kind: "view_schema",
    id: timelineViewSchemaId,
  },
): WorkbookPresenceInput {
  const input: WorkbookPresenceInput = {
    sheet_ref: { ...sheetRef },
    mode: presence.mode,
  };
  if (sheetRef.kind !== "extension_workspace" && presence.recordId !== null) {
    input.record_id = presence.recordId;
  }
  if (
    sheetRef.kind !== "extension_workspace" &&
    presence.mode === "editing" &&
    presence.fieldKey !== null
  ) {
    input.field_key = presence.fieldKey;
  }
  return input;
}

export function buildWorkbookPresenceUpdateMessage(
  presence: TimelinePresenceDraft,
  sheetRef: WorkbookSheetRef,
): WorkbookPresenceUpdateMessage {
  return {
    type: "presence_update",
    payload: {
      presence: buildWorkbookPresenceInput(presence, sheetRef),
    },
  };
}

export function buildWorkbookSocketSessionMessage({
  clientInstanceId,
  lastSeenStreamSeq,
  presence,
  resumeToken,
  sheetRef,
}: {
  readonly clientInstanceId: string;
  readonly lastSeenStreamSeq: number;
  readonly presence: TimelinePresenceDraft;
  readonly resumeToken: string | null;
  readonly sheetRef: WorkbookSheetRef;
}): WorkbookSocketSessionMessage {
  if (resumeToken) {
    return {
      type: "resume",
      payload: {
        client_instance_id: clientInstanceId,
        resume_token: resumeToken,
        last_seen_stream_seq: lastSeenStreamSeq,
        presence: buildWorkbookPresenceInput(presence, sheetRef),
      },
    };
  }
  return {
    type: "hello",
    payload: {
      client_instance_id: clientInstanceId,
      presence: buildWorkbookPresenceInput(presence, sheetRef),
    },
  };
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
