import type {
  MarkTimelineRecordReviewedResponse,
  SupersedeRecordResponse,
} from "@cartulary/protocol-ts/http";
import type { WorkbookRowObservation } from "../../query/WorkbookQueryRow";

export type TimelineReviewReceiptData = Readonly<
  MarkTimelineRecordReviewedResponse["data"]
>;
export type TimelineSupersedeReceiptData = Readonly<
  Extract<SupersedeRecordResponse["data"], { record_id: string }>
>;
export type TimelineCaptureReceipt = {
  readonly observation?: WorkbookRowObservation;
} & (
  | Readonly<{ operation: "mark-reviewed"; data: TimelineReviewReceiptData }>
  | Readonly<{ operation: "supersede"; data: TimelineSupersedeReceiptData }>
);
