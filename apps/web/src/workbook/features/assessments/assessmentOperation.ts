import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookProtocolCreateViewRowReceipt } from "../../adapters/workbookProtocolTypes";
import type {
  AssessmentCreateDraft,
  AssessmentCreateRequest,
} from "../../models/assessmentWorkbookModel";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookCommittedRecordPort } from "../../query/WorkbookCommittedRecordPort";

export type AssessmentDraft = Readonly<{
  values: AssessmentCreateDraft;
  mode: "standalone" | "follow_on";
  origin: { readonly recordId: string; readonly rowVersion: number } | null;
  revision: number;
}>;
export type AssessmentReview = Readonly<{
  authority: WorkbookMutationAuthority;
  draft: AssessmentDraft;
  sheetRef: SheetRef;
}>;
export type AssessmentAppendAttempt = Readonly<{
  clientTxnId: string;
  review: AssessmentReview;
  request: AssessmentCreateRequest;
  body: string;
  path: string;
  apiBase: string | undefined;
}>;
export type AssessmentAppendReceipt =
  Readonly<WorkbookProtocolCreateViewRowReceipt>;
export type AssessmentAppendOutcome =
  | { readonly kind: "accepted"; readonly receipt: AssessmentAppendReceipt }
  | { readonly kind: "uncertain" }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure };
export interface AssessmentAppendTransport {
  capture(
    review: AssessmentReview,
    clientTxnId: string,
  ): AssessmentAppendAttempt;
  send(
    attempt: AssessmentAppendAttempt,
    signal: AbortSignal,
  ): Promise<AssessmentAppendOutcome>;
}
export type AssessmentAppendEntry = Readonly<{
  attempt: AssessmentAppendAttempt;
  phase: "submitting" | "uncertain" | "rejected" | "accepted";
  transportPending: boolean;
  receipt: AssessmentAppendReceipt | null;
  refresh: "none" | "required" | "refreshing" | "complete";
  message: string | null;
}>;
export type AssessmentPresentationBinding = Readonly<{
  sheetRef: SheetRef;
  isCurrent: () => boolean;
}>;
export type AssessmentAuthorityReader = (
  baseline: WorkbookMutationAuthority,
  signal: AbortSignal,
) => Promise<WorkbookMutationAuthority>;

/** Request and receipt trees must not be editable through presentation aliases. */
export function freezeAssessment<T>(value: T): T {
  if (value && typeof value === "object" && !Object.isFrozen(value)) {
    Object.freeze(value);
    for (const child of Object.values(value)) freezeAssessment(child);
  }
  return value;
}

export interface AssessmentCommittedRecordPort
  extends WorkbookCommittedRecordPort {
  wasRemoved(recordId: string, version: number): boolean;
}
