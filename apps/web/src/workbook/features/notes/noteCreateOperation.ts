import type {
  WorkbookProtocolCreateLinkedNoteReceipt,
  WorkbookProtocolCreateLinkedNoteRequest,
  WorkbookProtocolCreateViewRowReceipt,
} from "../../adapters/workbookProtocolTypes";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { NoteDraft } from "./noteCreateModel";

export type NoteReview = Readonly<{
  authority: WorkbookMutationAuthority;
  draft: NoteDraft;
}>;
export type NoteAttempt = Readonly<{
  operationID: "createRecordLinkedNote" | "createViewRow";
  apiBase: string | undefined;
  path: string;
  pathParameters: Readonly<Record<string, string>>;
  request: WorkbookProtocolCreateLinkedNoteRequest;
  body: string;
  clientTxnId: string;
  review: NoteReview;
}>;
export type NoteReceipt = Readonly<
  WorkbookProtocolCreateViewRowReceipt | WorkbookProtocolCreateLinkedNoteReceipt
>;
export type NoteOutcome =
  | { readonly kind: "accepted"; readonly receipt: NoteReceipt }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export interface NoteTransport {
  capture(review: NoteReview, clientTxnId: string): NoteAttempt;
  send(attempt: NoteAttempt, signal: AbortSignal): Promise<NoteOutcome>;
}
export type NoteEntry = Readonly<{
  attempt: NoteAttempt;
  phase: "submitting" | "uncertain" | "rejected" | "accepted";
  uncertain: boolean;
  transportPending: boolean;
  receipt: NoteReceipt | null;
  failure: WorkbookOperationFailure | null;
  refresh: "none" | "required" | "refreshing" | "complete";
  observations: readonly RecordChangedMessage[];
  message: string | null;
}>;
