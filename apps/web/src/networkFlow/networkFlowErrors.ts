import { extractError, publicErrorView } from "../services/browserApi";
import { networkFlowErrorMetadata } from "../services/networkFlowContractAdapter";

export type NetworkFlowRetryAction =
  | "correct_request"
  | "refresh_resource"
  | "restart_query"
  | "reduce_scope_or_limits"
  | "retry_with_backoff"
  | "do_not_retry";

export class NetworkFlowRequestError extends Error {
  readonly code: string;
  readonly field: string | null;
  readonly reasonCode: string | null;
  readonly retryAction: NetworkFlowRetryAction;
  readonly retryable: boolean;
  readonly status: number;

  constructor(options: {
    readonly code: string;
    readonly field?: string | null;
    readonly reasonCode?: string | null;
    readonly retryAction: NetworkFlowRetryAction;
    readonly retryable: boolean;
    readonly safeMessage: string;
    readonly status: number;
  }) {
    super(options.safeMessage);
    this.name = "NetworkFlowRequestError";
    this.code = options.code;
    this.field = options.field ?? null;
    this.reasonCode = options.reasonCode ?? null;
    this.retryAction = options.retryAction;
    this.retryable = options.retryable;
    this.status = options.status;
  }
}

export type NetworkFlowWorkspaceError = NetworkFlowRequestError | string;

export function networkFlowErrorMessage(
  error: NetworkFlowWorkspaceError | null,
): string | null {
  return error instanceof NetworkFlowRequestError ? error.message : error;
}

export function networkFlowRequestError(
  status: number,
  payload: unknown,
  fallbackMessage = "Network Flow request failed.",
): NetworkFlowRequestError {
  const apiError = extractError(payload);
  const view = publicErrorView(apiError, status);
  const code = view?.code ?? "unknown_public_error";
  const contract = networkFlowErrorMetadata.errors.find(
    (candidate) => candidate.code === code,
  );
  const retryAction =
    contract?.retry_action ?? defaultRetryAction(status, code);
  return new NetworkFlowRequestError({
    code,
    field: safeDetail(apiError?.details?.field),
    reasonCode: safeDetail(apiError?.details?.reason_code),
    retryAction,
    retryable:
      typeof apiError?.retryable === "boolean"
        ? apiError.retryable
        : retryAction === "retry_with_backoff",
    safeMessage: view?.statusText ?? fallbackMessage,
    status,
  });
}

export function networkFlowErrorFromUnknown(
  caught: unknown,
  fallbackMessage: string,
): NetworkFlowRequestError {
  if (caught instanceof NetworkFlowRequestError) {
    return caught;
  }
  return new NetworkFlowRequestError({
    code: "network_flow_transport_failed",
    retryAction: "retry_with_backoff",
    retryable: true,
    safeMessage: fallbackMessage,
    status: 0,
  });
}

export function isNetworkFlowAuthorizationLoss(
  error: NetworkFlowRequestError,
): boolean {
  return (
    error.status === 401 ||
    error.code === "session_required" ||
    error.code === "incident_not_found" ||
    error.code === "authorization_denied" ||
    (error.code === "network_flow_cursor_invalid" &&
      error.reasonCode === "authorization_lost")
  );
}

export function isNetworkFlowLifecycleLoss(
  error: NetworkFlowRequestError,
): boolean {
  return (
    error.code === "incident_closed" ||
    error.code === "network_flow_table_not_active" ||
    error.code === "network_flow_table_not_found"
  );
}

export function isNetworkFlowProtectedStateLoss(
  error: NetworkFlowRequestError,
): boolean {
  return (
    isNetworkFlowAuthorizationLoss(error) || isNetworkFlowLifecycleLoss(error)
  );
}

export function isNetworkFlowCursorInvalid(
  error: NetworkFlowRequestError,
): boolean {
  return (
    error.code === "network_flow_cursor_invalid" &&
    error.retryAction === "restart_query"
  );
}

function safeDetail(value: unknown): string | null {
  return typeof value === "string" && value.length <= 128 ? value : null;
}

function defaultRetryAction(
  status: number,
  code: string,
): NetworkFlowRetryAction {
  if (status === 401 || status === 403 || code === "session_required") {
    return "do_not_retry";
  }
  if (status === 404 || status === 409) {
    return "refresh_resource";
  }
  if (status >= 500 || status === 0) {
    return "retry_with_backoff";
  }
  return "correct_request";
}
