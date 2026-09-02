import type { IncidentCollaborationMessage } from "../../collaboration/IncidentCollaborationSession";
import type { SheetRef } from "../../shared/sheetRef";
import type {
  WorkbookPresenceInput,
  WorkbookPresenceMode,
} from "../utils/workbookPresence";

type RecordChangedMessage = Extract<
  IncidentCollaborationMessage,
  { type: "record_changed" }
>;

export type RecordChangedPayload = RecordChangedMessage["payload"];

export type WorkbookPresenceDraft = {
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

type MentionActionMentionLike = {
  mentionRowVersion: number | null;
};

export type MentionResolutionAction =
  | "resolve_item"
  | "dismiss_item"
  | "revert_to_unresolved";

type MentionActionPayload = {
  action: MentionResolutionAction;
  base_mention_row_version: number;
  client_txn_id: string;
  resolved_record_id?: string;
};

export function buildWorkbookPresenceInput(
  presence: WorkbookPresenceDraft,
  sheetRef: SheetRef,
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
  presence: WorkbookPresenceDraft,
  sheetRef: SheetRef,
): WorkbookPresenceUpdateMessage {
  return {
    type: "presence_update",
    payload: {
      presence: buildWorkbookPresenceInput(presence, sheetRef),
    },
  };
}

export function isRecordChangedMessage(
  message: IncidentCollaborationMessage,
): message is RecordChangedMessage {
  return message.type === "record_changed";
}

export function buildMentionActionPayload(
  mention: MentionActionMentionLike,
  action: MentionResolutionAction,
  clientTxnId: string,
  resolvedRecordId?: string,
): MentionActionPayload | null {
  if (mention.mentionRowVersion === null) {
    return null;
  }
  const body: MentionActionPayload = {
    base_mention_row_version: mention.mentionRowVersion,
    client_txn_id: clientTxnId,
    action,
  };
  if (action === "resolve_item" && resolvedRecordId !== undefined) {
    body.resolved_record_id = resolvedRecordId;
  }
  return body;
}
