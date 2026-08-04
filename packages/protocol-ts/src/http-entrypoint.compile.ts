import {
  buildHTTPOperationPath,
  type ErrorEnvelope,
  encodeHTTPOperationQuery,
  errorEnvelopeDecoder,
  type HTTPOperationRequest,
  type HTTPOperationResponse,
  httpOperationBindings,
  type SheetRef,
  validateHTTPOperationResponse,
} from "@cartulary/protocol-ts/http";

declare const errorEnvelope: ErrorEnvelope;
declare const sheetRef: SheetRef;
declare const request: HTTPOperationRequest<"getIncident">;
declare const response: HTTPOperationResponse<"getIncident">;

export const httpCompileSurface = {
  binding: httpOperationBindings.getIncident,
  decodedError: errorEnvelopeDecoder.decode(errorEnvelope),
  path: buildHTTPOperationPath("getIncident", { incident_id: "incident-1" }),
  query: encodeHTTPOperationQuery("getIncident"),
  request,
  responseValidation: validateHTTPOperationResponse("getIncident", response),
  sheetRef,
};
