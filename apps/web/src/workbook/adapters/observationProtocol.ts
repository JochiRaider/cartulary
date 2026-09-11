import type { CreateManualIndicatorObservationResponse } from "@cartulary/protocol-ts/http";

export { coreIndicatorTypes } from "@cartulary/protocol-ts/http";
export type ObservationReceipt =
  CreateManualIndicatorObservationResponse["data"];
