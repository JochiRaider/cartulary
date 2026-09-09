import { ExtensionAvailabilityUnavailableError } from "../extensions/extensionAvailability";
import { ObservationStopped } from "../services/asyncObservation";
import {
  NetworkFlowContractDecodeError,
  validNetworkFlowErrorEnvelope,
} from "../services/networkFlowContractAdapter";
import {
  isNetworkFlowAuthorizationLoss,
  NetworkFlowRequestError,
  networkFlowRequestError,
} from "./networkFlowErrors";

export type SavedGraphReadSurface =
  | "list"
  | "declaration"
  | "result"
  | "contributors"
  | "job";
export type SavedGraphReadDisposition =
  | "retain"
  | "scope"
  | "declaration"
  | "binding";

/** Only validated public errors can withdraw an authority that still permits reads. */
export function savedGraphResponseError(
  status: number,
  payload: unknown,
): NetworkFlowRequestError {
  if (!validNetworkFlowErrorEnvelope(status, payload))
    return savedGraphReadFailure(
      new SyntaxError(),
      "The server returned an invalid saved-graph response. Reload to recover.",
    );
  return networkFlowRequestError(status, payload);
}
export function savedGraphReadFailure(
  caught: unknown,
  message: string,
): NetworkFlowRequestError {
  if (caught instanceof NetworkFlowRequestError) return caught;
  const invalid =
    caught instanceof SyntaxError ||
    caught instanceof NetworkFlowContractDecodeError;
  const unavailable = caught instanceof ExtensionAvailabilityUnavailableError;
  const stopped = caught instanceof ObservationStopped;
  return new NetworkFlowRequestError({
    code: unavailable
      ? "extension_workspace_unavailable"
      : invalid
        ? "network_flow_invalid_response"
        : stopped
          ? "network_flow_observation_stopped"
          : "network_flow_transport_failed",
    status: 0,
    reasonCode: stopped ? caught.reason : null,
    retryAction: unavailable ? "refresh_resource" : "retry_with_backoff",
    retryable: !unavailable,
    safeMessage: message,
  });
}
export function savedGraphReadDisposition(
  error: NetworkFlowRequestError,
  surface: SavedGraphReadSurface,
): SavedGraphReadDisposition {
  if (
    isNetworkFlowAuthorizationLoss(error) ||
    error.code === "incident_closed" ||
    error.code === "extension_workspace_unavailable" ||
    error.code === "extension_profile_unclaimed"
  )
    return "scope";
  // A retained job can expire independently of its declaration or selected result.
  if (surface === "job") return "retain";
  if (error.code === "network_flow_graph_view_not_found") return "declaration";
  if (
    [
      "network_flow_graph_view_not_materialized",
      "network_flow_graph_query_stale",
      "network_flow_table_not_found",
      "network_flow_table_not_active",
      "network_flow_source_changed",
    ].includes(error.code) ||
    (error.code === "network_flow_cursor_invalid" &&
      error.reasonCode === "scope_stale")
  )
    return "binding";
  return "retain";
}
