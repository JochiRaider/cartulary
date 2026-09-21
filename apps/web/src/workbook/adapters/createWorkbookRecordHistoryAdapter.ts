import {
  type DeleteRecordResponse,
  type GetRecordHistoryResponse,
  httpOperationBindings,
  type RollbackRecordResponse,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import {
  historyTargetEqual,
  normalizeRecordHistoryData,
} from "../history/workbookHistoryItem";
import type {
  HistoryAttempt,
  HistoryReceipt,
  WorkbookRecordHistoryPort,
} from "../history/workbookHistoryOperation";
import { validHistoryPaging } from "../history/workbookHistoryPage";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

const invalid = {
  kind: "invalid_contract" as const,
  message: "Invalid row history response.",
};

export function createWorkbookRecordHistoryAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): WorkbookRecordHistoryPort {
  return {
    async load(recordId, signal, request = {}) {
      try {
        const result = await fetchHTTPOperation<GetRecordHistoryResponse>({
          apiBase: options.apiBase,
          operationID: "getRecordHistory",
          query: {
            ...(request.limit === undefined ? {} : { limit: request.limit }),
            ...(request.cursorToken === undefined
              ? {}
              : { cursor_token: request.cursorToken }),
          },
          pathParameters: { record_id: recordId },
          init: { signal },
        });
        if (!result.ok) {
          const failure = classifyWorkbookOperationFailure(
            result.status,
            result.payload,
            "getRecordHistory",
          );
          return {
            kind: "rejected",
            failure: failure.kind === "invalid_contract" ? invalid : failure,
          };
        }
        const data = normalizeRecordHistoryData(result.payload.data);
        if (
          data === null ||
          !validHistoryPaging(result.payload.meta.paging) ||
          data.record_id !== recordId ||
          data.incident_id !== options.incidentId
        )
          return { kind: "rejected", failure: invalid };
        return {
          kind: "accepted",
          value: { ...data, paging: result.payload.meta.paging },
        };
      } catch {
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "History could not be read. Try refreshing history.",
          },
        };
      }
    },
    async send(attempt, signal) {
      if (attempt.incidentId !== options.incidentId || signal.aborted)
        return {
          kind: "rejected",
          failure: {
            kind: "stale_target",
            message: "Review the current record before continuing.",
          },
        };
      const operationID =
        attempt.operation === "delete"
          ? "deleteRecord"
          : attempt.operation === "restore"
            ? "restoreRecord"
            : "rollbackRecord";
      let responseStatus: number | null = null;
      try {
        const result = await fetchHTTPOperation<
          DeleteRecordResponse | RollbackRecordResponse
        >({
          apiBase: options.apiBase,
          operationID,
          pathParameters: { record_id: attempt.subject.recordId },
          onResponse: (response) => {
            responseStatus = response.status;
          },
          init: {
            method: httpOperationBindings[operationID].method,
            body: attempt.body,
            signal,
          },
        });
        if (!result.ok) {
          if (
            responseStatus === null ||
            responseStatus < 400 ||
            responseStatus >= 500 ||
            !result.payload.error?.code
          )
            return { kind: "uncertain" };
          const failure = classifyWorkbookOperationFailure(
            result.status,
            result.payload,
            operationID,
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = historyReceipt(attempt, result.payload);
        return receipt === null
          ? { kind: "uncertain" }
          : { kind: "acknowledged", receipt };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}

function historyReceipt(
  attempt: HistoryAttempt,
  response: DeleteRecordResponse | RollbackRecordResponse,
): HistoryReceipt | null {
  const data = response.data;
  if (
    data.incident_id !== attempt.incidentId ||
    data.record_id !== attempt.subject.recordId ||
    !Number.isSafeInteger(data.row_version) ||
    data.row_version < 1
  )
    return null;
  const common = {
    incidentId: data.incident_id,
    recordId: data.record_id,
    rowVersion: data.row_version,
  };
  if (attempt.operation === "rollback") {
    if (
      !("target" in data) ||
      attempt.pending.kind !== "rollback" ||
      !historyTargetEqual(attempt.pending.target, data.target)
    )
      return null;
    if (
      !data.rollback_change_set_id ||
      data.affected_record_ids.length === 0 ||
      data.affected_record_ids.some(
        (id, index, ids) => !id || (index > 0 && (ids[index - 1] ?? "") >= id),
      )
    )
      return null;
    return {
      ...common,
      kind: "rollback",
      target: data.target,
      changeSetId: data.rollback_change_set_id,
      affectedRecordIds: [...data.affected_record_ids],
      ...(data.target_change_set_id == null
        ? {}
        : { targetChangeSetId: data.target_change_set_id }),
    };
  }
  if (
    data.row_version <= attempt.pending.rowVersion ||
    !("deleted" in data) ||
    data.deleted !== (attempt.operation === "delete") ||
    !data.change_set_id
  )
    return null;
  if (
    data.deleted
      ? !data.deleted_at || !data.deleted_by_user_id
      : data.deleted_at !== null || data.deleted_by_user_id !== null
  )
    return null;
  return {
    ...common,
    kind: attempt.operation,
    changeSetId: data.change_set_id,
    deleted: data.deleted,
    deletedAt: data.deleted_at,
    deletedByUserId: data.deleted_by_user_id,
  };
}
