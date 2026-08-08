import {
  type CartularyErrorPresentation,
  cartularyErrorPresentation,
} from "@cartulary/ui-contracts";

export type PublicErrorOperationFamily =
  | "authentication"
  | "evidence_preview"
  | "extension"
  | "field_mutation"
  | "surface_load"
  | "surface_refresh";

export function resolvePublicErrorPresentation({
  code,
  hasAuthorizedMaterialization,
  operationFamily,
  status,
}: {
  readonly code: string;
  readonly hasAuthorizedMaterialization: boolean;
  readonly operationFamily: PublicErrorOperationFamily;
  readonly status: number;
}): CartularyErrorPresentation {
  if (code === "same_field_conflict") {
    return cartularyErrorPresentation("same_field_conflict");
  }
  if (code === "client_txn_conflict") {
    return cartularyErrorPresentation("client_txn_conflict");
  }
  if (code === "pending_queue_capacity_exceeded") {
    return cartularyErrorPresentation("queue_overflow");
  }
  if (status === 401 || code === "session_required") {
    return cartularyErrorPresentation("authentication_required");
  }
  if (
    code === "authorization_denied" ||
    code === "incident_not_found" ||
    code === "incident_access_lost"
  ) {
    return cartularyErrorPresentation("permission_or_incident_access_loss");
  }
  if (operationFamily === "extension") {
    return cartularyErrorPresentation("extension_unavailable");
  }
  if (operationFamily === "evidence_preview") {
    return cartularyErrorPresentation("evidence_preview_blocked");
  }
  if (status === 400 || status === 422) {
    return cartularyErrorPresentation("local_validation");
  }
  if (
    operationFamily === "surface_refresh" ||
    (operationFamily === "surface_load" && hasAuthorizedMaterialization)
  ) {
    return cartularyErrorPresentation("stale_refresh");
  }
  if (operationFamily === "surface_load") {
    return cartularyErrorPresentation("initial_load_failure");
  }
  return cartularyErrorPresentation("unknown_future_error");
}
