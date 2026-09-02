import { resolvePublicErrorPresentation } from "../../shared/publicErrorPresentation";

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

export function workbookEditRecoveryPresentation({
  errorCode,
  status = 409,
}: {
  readonly errorCode: string;
  readonly status?: number | undefined;
}): WorkbookEditRecoveryPresentation {
  const family = resolvePublicErrorPresentation({
    code: errorCode,
    hasAuthorizedMaterialization: true,
    operationFamily: "field_mutation",
    status,
  }).family;
  if (family === "client_txn_conflict") {
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
