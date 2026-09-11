import type {
  MarkTimelineRecordReviewedResponse,
  SupersedeRecordResponse,
} from "@cartulary/protocol-ts/http";

export type TimelineReviewReceiptData = Readonly<
  MarkTimelineRecordReviewedResponse["data"]
>;
export type TimelineSupersedeReceiptData = Readonly<
  Extract<SupersedeRecordResponse["data"], { record_id: string }>
>;
export type TimelineCaptureReceipt =
  | Readonly<{ operation: "mark-reviewed"; data: TimelineReviewReceiptData }>
  | Readonly<{ operation: "supersede"; data: TimelineSupersedeReceiptData }>;
