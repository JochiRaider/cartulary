import type {
  WorkbookProtocolCreateViewRowReceipt,
  WorkbookProtocolCreateViewRowRequest,
} from "../../adapters/workbookProtocolTypes";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { ContextualCreateDraft } from "./contextualCreateModel";

export type ContextualCreateReview = Readonly<{
  authority: WorkbookMutationAuthority;
  draft: ContextualCreateDraft;
}>;
export type ContextualCreateAttempt = Readonly<{
  clientTxnId: string;
  review: ContextualCreateReview;
  request: WorkbookProtocolCreateViewRowRequest;
  body: string;
  path: string;
  apiBase: string | undefined;
}>;
export type ContextualCreateReceipt =
  Readonly<WorkbookProtocolCreateViewRowReceipt>;
export type ContextualCreateOutcome =
  | { readonly kind: "accepted"; readonly receipt: ContextualCreateReceipt }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export interface ContextualCreateTransport {
  capture(
    review: ContextualCreateReview,
    clientTxnId: string,
  ): ContextualCreateAttempt;
  send(
    attempt: ContextualCreateAttempt,
    signal: AbortSignal,
  ): Promise<ContextualCreateOutcome>;
}
export type ContextualCreateEntry = Readonly<{
  observations: readonly RecordChangedMessage[];
  attempt: ContextualCreateAttempt;
  phase: "submitting" | "uncertain" | "rejected" | "accepted";
  transportPending: boolean;
  receipt: ContextualCreateReceipt | null;
  refresh: "none" | "required" | "refreshing" | "complete";
  message: string | null;
}>;
export type ContextualAuthorityReader = (
  baseline: WorkbookMutationAuthority,
  signal: AbortSignal,
) => Promise<WorkbookMutationAuthority>;
export type ContextualCandidate = Readonly<{
  recordId: string;
  displayText: string;
  viewSchemaId: string;
  row?: WorkbookQueryRow;
}>;
export type ContextualCandidatePage = Readonly<{
  candidates: readonly ContextualCandidate[];
  hasMore: boolean;
  nextCursor: string | null;
}>;
export type ContextualCandidateQuery = Readonly<{
  viewSchemaId: string;
  cursor: string | null;
  queryState: WorkbookQueryState;
  signal: AbortSignal;
}>;
export interface ContextualCreateReader {
  page(
    input: ContextualCandidateQuery,
  ): Promise<WorkbookPortResult<ContextualCandidatePage>>;
  availableViews(signal: AbortSignal): Promise<readonly string[]>;
  verify(draft: ContextualCreateDraft, signal: AbortSignal): Promise<void>;
}
