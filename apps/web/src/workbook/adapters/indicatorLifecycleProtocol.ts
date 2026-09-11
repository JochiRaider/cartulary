import type {
  AppendIndicatorStateIntervalRequest,
  AppendIndicatorStateIntervalResponse,
} from "@cartulary/protocol-ts/http";
import { coreIndicatorLifecycleConstraints } from "@cartulary/protocol-ts/http";

export const indicatorLifecycleConstraints = coreIndicatorLifecycleConstraints;
export type IndicatorLifecycleValues = Omit<
  AppendIndicatorStateIntervalRequest,
  "base_row_version" | "client_txn_id"
>;
export type IndicatorLifecycleReceipt =
  AppendIndicatorStateIntervalResponse["data"];
export type IndicatorLifecycleInterval = IndicatorLifecycleReceipt["interval"];
