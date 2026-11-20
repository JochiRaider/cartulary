import type { HTTPOperationErrorCode } from "../../services/browserApi";
import type { ClientErrorCode } from "../../shared/publicError";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";

/**
 * The pending queue's settlement classification of a replay failure: the
 * Workbook failure kind a mutation driver reports, the queue's own closure
 * halt, or a client-originated dispatch failure. It is closed, so comparing it
 * with any other literal is a type error (Core 03 REQ-03-086a).
 */
export type PendingReplaySettlementCode =
  | WorkbookOperationFailure["kind"]
  | Extract<HTTPOperationErrorCode, "incident_closed">
  | Extract<
      ClientErrorCode,
      "pending_runtime_metadata_missing" | "transport_failure"
    >;

const clientTransactionRecoveryMessage =
  "A queued edit could not be replayed safely. Retry it with a new request ID, or discard the blocked edit to continue.";

const terminalReplayRecoveryMessage =
  "A queued edit could not be completed safely. Discard the blocked edit to continue with later queued edits.";

export type WorkbookEditRecoveryPresentation =
  | {
      readonly kind: "client_txn_conflict";
      readonly message: typeof clientTransactionRecoveryMessage;
      readonly retryAllowed: true;
    }
  | {
      readonly kind: "terminal_replay_failure";
      readonly message: typeof terminalReplayRecoveryMessage;
      readonly retryAllowed: false;
    };

/**
 * Presents a pending replay halt. `errorCode` is the queue's settlement
 * classification, not a decoded server code; only a transaction conflict
 * admits retry with a new request identity.
 */
export function workbookEditRecoveryPresentation({
  errorCode,
}: {
  readonly errorCode: PendingReplaySettlementCode;
  readonly status?: number | undefined;
}): WorkbookEditRecoveryPresentation {
  if (errorCode === "client_txn_conflict") {
    return {
      kind: "client_txn_conflict",
      message: clientTransactionRecoveryMessage,
      retryAllowed: true,
    };
  }
  return {
    kind: "terminal_replay_failure",
    message: terminalReplayRecoveryMessage,
    retryAllowed: false,
  };
}
