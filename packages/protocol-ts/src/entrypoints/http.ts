import type {
  CollectionActionsV1,
  ErrorEnvelope,
  EvidenceCreateRequest,
  RecordHistoryData,
  RecordHistoryItem,
  SafeUserResource,
  SessionResource,
  SheetRef,
  TimelineCreateRequest,
  ViewCell,
  ViewRow,
} from "../generated/core-http-types.js";
import { validateCartularyCoreHttpErrorEnvelopeV1 } from "../generated/core-http-validators.js";
import { createDecoder } from "../internal/decoder.js";

export * from "../generated/http-operation-bindings.js";
export type { Decoder } from "../internal/decoder.js";
export type {
  CollectionActionsV1,
  ErrorEnvelope,
  EvidenceCreateRequest,
  RecordHistoryData,
  RecordHistoryItem,
  SafeUserResource,
  SessionResource,
  SheetRef,
  TimelineCreateRequest,
  ViewCell,
  ViewRow,
};

export const errorEnvelopeDecoder = createDecoder<ErrorEnvelope>(
  "cartulary.core_http.ErrorEnvelope.v1",
  validateCartularyCoreHttpErrorEnvelopeV1,
);
