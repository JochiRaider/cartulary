import { fetchJSON } from "../../../services/workbookApi";
import type {
  PendingReplayPayloadIntent,
  PendingReplayUnitState,
} from "../../utils/workbookPendingQueue";

export type TimelineMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id?: string;
    row: unknown;
  };
};

export type TimelineMutationFetchResult = Awaited<
  ReturnType<typeof fetchJSON<TimelineMutationEnvelope>>
>;

export type TimelineWorkbookTimingRecorder = (
  name: string,
  fields?: Record<string, unknown>,
) => void;

export type TimelineRecordActionName = "mark-reviewed" | "supersede";
export type TimelineDeleteRestoreOperation = "delete" | "restore";
export type TimelineConflictResolutionKind =
  | "keep_saved"
  | "use_unsaved"
  | "merged_value";

export function buildTimelineRecordActionPayload({
  action,
  baseRowVersion,
  clientTxnId,
  replacementRecordId,
}: {
  readonly action: TimelineRecordActionName;
  readonly baseRowVersion: number;
  readonly clientTxnId: string;
  readonly replacementRecordId?: string | null | undefined;
}): Record<string, unknown> {
  if (action === "mark-reviewed") {
    return {
      base_row_version: baseRowVersion,
      client_txn_id: clientTxnId,
      reason: "Reviewed from workbook",
    };
  }
  return {
    base_row_version: baseRowVersion,
    client_txn_id: clientTxnId,
    reason: "Superseded from workbook",
    replacement_record_id: replacementRecordId ?? "",
  };
}

export function buildTimelineDeleteRestorePayload({
  baseRowVersion,
  clientTxnId,
  operation,
}: {
  readonly baseRowVersion: number;
  readonly clientTxnId: string;
  readonly operation: TimelineDeleteRestoreOperation;
}): Record<string, unknown> {
  return {
    base_row_version: baseRowVersion,
    client_txn_id: clientTxnId,
    reason:
      operation === "delete"
        ? "Deleted from workbook history"
        : "Restored from workbook history",
  };
}

export function buildTimelineRollbackPayload({
  baseRowVersion,
  clientTxnId,
  target,
}: {
  readonly baseRowVersion: number;
  readonly clientTxnId: string;
  readonly target: unknown;
}): Record<string, unknown> {
  return {
    base_row_version: baseRowVersion,
    client_txn_id: clientTxnId,
    reason: "Rollback from workbook history",
    target,
  };
}

export function buildTimelineConflictResolutionPayload({
  clientTxnId,
  conflictResolutionClass,
  conflictToken,
  localValue,
  mergedDraft,
  resolutionKind,
}: {
  readonly clientTxnId: string;
  readonly conflictResolutionClass: string;
  readonly conflictToken: string;
  readonly localValue: unknown;
  readonly mergedDraft: unknown;
  readonly resolutionKind: TimelineConflictResolutionKind;
}): Record<string, unknown> {
  const body: Record<string, unknown> = {
    conflict_token: conflictToken,
    resolution_kind: resolutionKind,
    client_txn_id: clientTxnId,
  };
  if (resolutionKind === "use_unsaved") {
    body.resolved_value = localValue;
  } else if (resolutionKind === "merged_value") {
    body.resolved_value =
      conflictResolutionClass === "collection_review"
        ? localValue
        : mergedDraft;
  }
  return body;
}

export async function dispatchTimelinePendingReplayMutation({
  payload,
  recordTiming,
  unit,
}: {
  readonly payload: PendingReplayPayloadIntent;
  readonly recordTiming: TimelineWorkbookTimingRecorder;
  readonly unit: PendingReplayUnitState;
}): Promise<TimelineMutationFetchResult> {
  recordTiming("pending_fetch_start", {
    clientTxnId: unit.clientTxnId,
    kind: unit.kind,
    rowKey: unit.rowKey,
  });
  return fetchJSON<TimelineMutationEnvelope>(
    unit.path,
    {
      method: unit.method,
      body: JSON.stringify(payload),
    },
    {
      onJSONParsed: () => {
        recordTiming("pending_fetch_json_parsed", {
          clientTxnId: unit.clientTxnId,
          kind: unit.kind,
          rowKey: unit.rowKey,
        });
      },
      onResponse: (response) => {
        recordTiming("pending_fetch_response", {
          clientTxnId: unit.clientTxnId,
          kind: unit.kind,
          rowKey: unit.rowKey,
          serverTiming: response.headers.get("server-timing") ?? "",
          status: response.status,
        });
      },
    },
  );
}
