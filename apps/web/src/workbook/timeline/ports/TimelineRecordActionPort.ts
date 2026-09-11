import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type {
  TimelineCaptureReview,
  TimelineCaptureSubject,
} from "../actions/timelineCaptureActionModel";
import type { TimelineCaptureReceipt } from "../adapters/timelineCaptureProtocol";

export type TimelineRecordActionAccepted = {
  readonly captureState: string;
  readonly changeSetId: string;
  readonly incidentId: string;
  readonly reason: string | null;
  readonly recordId: string;
  readonly replacementRecordId: string | null;
  readonly rowVersion: number;
};
export type TimelineCaptureAttempt = Readonly<{
  id: string;
  review: TimelineCaptureReview;
  operationID: "markTimelineRecordReviewed" | "supersedeRecord";
  method: "POST";
  path: string;
  apiBase: string | undefined;
  body: string;
}>;
export type TimelineCaptureOutcome =
  | { readonly kind: "acknowledged"; readonly receipt: TimelineCaptureReceipt }
  | { readonly kind: "uncertain" }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure };
export interface TimelineRecordActionPort {
  capture(review: TimelineCaptureReview, id: string): TimelineCaptureAttempt;
  send(
    attempt: TimelineCaptureAttempt,
    signal: AbortSignal,
  ): Promise<TimelineCaptureOutcome>;
}
export type TimelineCaptureBinding = Readonly<{
  isCurrent: () => boolean;
  matchesReview: () => boolean;
  prepare: (signal: AbortSignal) => Promise<TimelineCaptureSubject | null>;
}>;
export type TimelineCaptureOperation = Readonly<{
  key: number;
  review: TimelineCaptureReview;
  attempt: TimelineCaptureAttempt | null;
  phase:
    | "preparing"
    | "preparation_failed"
    | "submitting"
    | "uncertain"
    | "rejected"
    | "acknowledged";
  transportPending: boolean;
  receipt: TimelineCaptureReceipt | null;
  failure: WorkbookOperationFailure | null;
  reconciliation: "pending" | "refreshing" | "required" | "complete";
}>;
