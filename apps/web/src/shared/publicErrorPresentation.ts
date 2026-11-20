import {
  type CartularyErrorPresentation,
  cartularyErrorPresentation,
} from "@cartulary/ui-contracts";
import type { HTTPOperationFailureCode } from "../services/browserApi";
import {
  type PublicErrorReasonCode,
  validatedPublicErrorReason,
} from "../services/publicErrorIdentity";
import { type ClientErrorCode, isTemporaryUnavailability } from "./publicError";

/**
 * Codes the shared presentation may branch on: a code some generated
 * operation declares or a client-originated code.
 */
export type PresentedPublicErrorCode =
  | HTTPOperationFailureCode
  | ClientErrorCode;

export type PublicErrorOperationFamily =
  | "authentication"
  | "evidence_preview"
  | "evidence_download"
  | "extension"
  | "field_mutation"
  | "surface_load"
  | "surface_refresh";

export function resolvePublicErrorPresentation({
  code,
  hasAuthorizedMaterialization,
  operationFamily,
  reasonCode,
  status,
}: {
  readonly code: PresentedPublicErrorCode;
  readonly hasAuthorizedMaterialization: boolean;
  readonly operationFamily: PublicErrorOperationFamily;
  readonly reasonCode?: PublicErrorReasonCode | undefined;
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
  if (code === "authorization_denied" || code === "incident_not_found") {
    return cartularyErrorPresentation("permission_or_incident_access_loss");
  }
  // Temporary unavailability always keeps a retry path (Core 03 REQ-03-086a):
  // it never runs an extension's Base fallback or the actionless unknown family.
  if (isTemporaryUnavailability({ code, status })) {
    return cartularyErrorPresentation(
      operationFamily === "surface_load" && !hasAuthorizedMaterialization
        ? "initial_load_failure"
        : "stale_refresh",
    );
  }
  if (operationFamily === "extension") {
    return cartularyErrorPresentation("extension_unavailable");
  }
  if (
    operationFamily === "evidence_preview" &&
    code === "evidence_access_unavailable" &&
    validatedPublicErrorReason(code, reasonCode) !== undefined
  ) {
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
