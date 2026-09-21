import type { SheetRef } from "../../../shared/sheetRef";
import type {
  WorkbookProtocolNoteAssociationsPage,
  WorkbookProtocolNoteAssociationsReceipt,
  WorkbookProtocolNoteAssociationsRequest,
} from "../../adapters/workbookProtocolTypes";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

export const noteAssociationView = "cartulary.view.notes.v1";
export type NoteAssociationKind =
  WorkbookProtocolNoteAssociationsRequest["kind"];
export type NoteAssociationAction =
  WorkbookProtocolNoteAssociationsRequest["actions"][number];
export type NoteAssociationPage = WorkbookProtocolNoteAssociationsPage["data"];
export type NoteAssociationReview = Readonly<{
  authority: WorkbookMutationAuthority;
  row: WorkbookQueryRow;
  kind: NoteAssociationKind;
  actions: readonly NoteAssociationAction[];
  sheetRef: SheetRef;
  label: string;
}>;
export type NoteAssociationAttempt = Readonly<{
  review: NoteAssociationReview;
  clientTxnId: string;
  apiBase: string | undefined;
  path: string;
  body: string;
  request: WorkbookProtocolNoteAssociationsRequest;
}>;
export type NoteAssociationReceipt = WorkbookProtocolNoteAssociationsReceipt;
export type NoteAssociationOutcome =
  | { readonly kind: "accepted"; readonly receipt: NoteAssociationReceipt }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export interface NoteAssociationTransport {
  capture(
    review: NoteAssociationReview,
    clientTxnId: string,
  ): NoteAssociationAttempt;
  send(
    attempt: NoteAssociationAttempt,
    signal: AbortSignal,
  ): Promise<NoteAssociationOutcome>;
  list(
    noteId: string,
    kind: NoteAssociationKind,
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<
    | { readonly kind: "accepted"; readonly page: NoteAssociationPage }
    | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  >;
}
export type NoteAssociationReader = Pick<
  WorkbookAuthoringReadPort,
  "page" | "availableViews"
> & {
  verify(kind: NoteAssociationKind, signal: AbortSignal): Promise<void>;
};
export type NoteAssociationEntry = Readonly<{
  attempt: NoteAssociationAttempt;
  order: number;
  phase: "submitting" | "uncertain" | "rejected" | "accepted";
  uncertain: boolean;
  transportPending: boolean;
  receipt: NoteAssociationReceipt | null;
  failure: WorkbookOperationFailure | null;
  refresh: "none" | "required" | "refreshing" | "complete";
}>;
export type NoteAssociationList = Readonly<{
  noteId: string;
  kind: NoteAssociationKind;
  state:
    | "initial_loading"
    | "ready"
    | "refreshing"
    | "stale_failure"
    | "unavailable";
  page: NoteAssociationPage | null;
  message: string | null;
}>;
export function noteAssociationListKey(
  noteId: string,
  kind: NoteAssociationKind,
) {
  return JSON.stringify([noteId, kind]);
}
