import type {
  WorkbookProtocolCreateViewRowReceipt,
  WorkbookProtocolCreateViewRowRequest,
} from "../../adapters/workbookProtocolTypes";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { CoordinationDraft } from "./coordinationCreateModel";

export type CoordinationReview = Readonly<{
  authority: WorkbookMutationAuthority;
  draft: CoordinationDraft;
}>;
export type CoordinationAttempt = Readonly<{
  operationID: "createViewRow";
  apiBase: string | undefined;
  path: string;
  pathParameters: Readonly<Record<string, string>>;
  request: WorkbookProtocolCreateViewRowRequest;
  body: string;
  clientTxnId: string;
  review: CoordinationReview;
}>;
export type CoordinationReceipt =
  Readonly<WorkbookProtocolCreateViewRowReceipt>;
export type CoordinationOutcome =
  | { readonly kind: "accepted"; readonly receipt: CoordinationReceipt }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export interface CoordinationTransport {
  capture(review: CoordinationReview, clientTxnId: string): CoordinationAttempt;
  send(
    attempt: CoordinationAttempt,
    signal: AbortSignal,
  ): Promise<CoordinationOutcome>;
}
export type CoordinationEntry = Readonly<{
  attempt: CoordinationAttempt;
  phase: "submitting" | "uncertain" | "rejected" | "accepted";
  uncertain: boolean;
  transportPending: boolean;
  receipt: CoordinationReceipt | null;
  failure: WorkbookOperationFailure | null;
  refresh: "none" | "required" | "refreshing" | "complete";
  observations: readonly RecordChangedMessage[];
  message: string | null;
}>;
