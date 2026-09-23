import type {
  WorkbookProtocolCreateViewRowReceipt,
  WorkbookProtocolCreateViewRowRequest,
} from "../../adapters/workbookProtocolTypes";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";

export type {
  WorkbookAuthoringAuthorityReader as ContextualAuthorityReader,
  WorkbookAuthoringReadPort as ContextualCreateReader,
} from "../../ports/WorkbookAuthoringReadPort";

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
