import {
  type ErrorEnvelope,
  errorEnvelopeDecoder,
} from "@cartulary/protocol-ts/http";

export type WorkbookPublicErrorDecodeResult =
  | {
      readonly kind: "decoded";
      readonly envelope: ErrorEnvelope;
    }
  | {
      readonly kind: "invalid_contract";
      readonly message: string;
    };

function isInvalidSuccessEnvelope(value: unknown): boolean {
  if (!value || typeof value !== "object" || !("error" in value)) {
    return false;
  }
  const error = value.error;
  return (
    error !== null &&
    typeof error === "object" &&
    "code" in error &&
    error.code === "invalid_public_contract_response"
  );
}

export function decodeWorkbookPublicError(
  payload: unknown,
): WorkbookPublicErrorDecodeResult {
  if (isInvalidSuccessEnvelope(payload)) {
    return {
      kind: "invalid_contract",
      message: "The server returned an invalid public contract response.",
    };
  }
  const decoded = errorEnvelopeDecoder.decode(payload);
  return decoded.ok
    ? { kind: "decoded", envelope: decoded.value }
    : {
        kind: "invalid_contract",
        message: "The server returned an invalid public error response.",
      };
}
