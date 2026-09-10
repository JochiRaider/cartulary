import { extractError, publicErrorView } from "../services/browserApi";
import {
  NetworkFlowContractDecodeError,
  networkFlowErrorMetadata,
} from "../services/networkFlowContractAdapter";

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
  readonly filterIndex: number | null;
  readonly operator: string | null;
  readonly reasonCode: string | null;
  readonly retryAction: NetworkFlowRetryAction;
  readonly retryable: boolean;
  readonly status: number;

  constructor(options: {
    readonly code: string;
    readonly field?: string | null;
    readonly filterIndex?: number | null;
    readonly operator?: string | null;
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
    this.filterIndex = options.filterIndex ?? null;
    this.operator = options.operator ?? null;
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
  const suppliedAction = apiError?.details?.retry_action;
  const retryAction =
    suppliedAction === undefined
      ? (contract?.retry_action ?? defaultRetryAction(status, code))
      : isRetryAction(suppliedAction)
        ? suppliedAction
        : "do_not_retry";
  return new NetworkFlowRequestError({
    code,
    field:
      safeDetail(apiError?.details?.field_key) ??
      safeDetail(apiError?.details?.field),
    filterIndex:
      typeof apiError?.details?.filter_index === "number" &&
      Number.isSafeInteger(apiError.details.filter_index) &&
      apiError.details.filter_index >= 0
        ? apiError.details.filter_index
        : null,
    operator: safeDetail(apiError?.details?.op),
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
  if (
    caught instanceof NetworkFlowContractDecodeError ||
    caught instanceof SyntaxError
  ) {
    return new NetworkFlowRequestError({
      code: "network_flow_response_invalid",
      retryAction: "do_not_retry",
      retryable: false,
      safeMessage:
        "The server returned an invalid Network Flow page. No new results were accepted.",
      status: 0,
    });
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
      (error.reasonCode === "authorization_lost" ||
        error.reasonCode === "actor_mismatch"))
  );
}

function isRetryAction(value: unknown): value is NetworkFlowRetryAction {
  return (
    typeof value === "string" &&
    [
      "correct_request",
      "refresh_resource",
      "restart_query",
      "reduce_scope_or_limits",
      "retry_with_backoff",
      "do_not_retry",
    ].includes(value)
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
