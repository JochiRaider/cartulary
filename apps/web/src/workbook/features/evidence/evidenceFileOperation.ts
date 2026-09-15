import type {
  WorkbookProtocolAttachBlobReceipt,
  WorkbookProtocolBlobSlotReceipt,
  WorkbookProtocolCreateViewRowReceipt,
} from "../../adapters/workbookProtocolTypes";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";

export type EvidenceFileStage = "slot" | "attach" | "create";
export type EvidenceFileReceipt =
  | WorkbookProtocolAttachBlobReceipt
  | WorkbookProtocolCreateViewRowReceipt;
export type EvidenceFileAttempt = Readonly<{
  stage: EvidenceFileStage;
  authority: WorkbookMutationAuthority;
  clientTxnId: string;
  path: string;
  body: string;
  apiBase: string | undefined;
  recordId: string | null;
  baseRowVersion: number | null;
  objectBlobId: string | null;
}>;
export type EvidenceFileOutcome<T> =
  | { readonly kind: "accepted"; readonly receipt: T }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export type EvidenceTransferTarget =
  WorkbookProtocolBlobSlotReceipt["data"]["upload_target"];
export type EvidenceTransferAttempt = Readonly<{
  authority: WorkbookMutationAuthority;
  objectBlobId: string;
  target: EvidenceTransferTarget;
}>;
export interface EvidenceFileTransport {
  capture(input: {
    stage: EvidenceFileStage;
    authority: WorkbookMutationAuthority;
    clientTxnId: string;
    recordId?: string;
    baseRowVersion?: number;
    objectBlobId?: string;
    fields?: Readonly<Record<string, unknown>>;
    file?: File;
  }): EvidenceFileAttempt;
  slot(
    attempt: EvidenceFileAttempt,
    signal: AbortSignal,
  ): Promise<EvidenceFileOutcome<WorkbookProtocolBlobSlotReceipt>>;
  finalize(
    attempt: EvidenceFileAttempt,
    signal: AbortSignal,
  ): Promise<EvidenceFileOutcome<EvidenceFileReceipt>>;
  transfer(
    attempt: EvidenceTransferAttempt,
    file: File,
    signal: AbortSignal,
  ): Promise<
    | { readonly kind: "accepted" }
    | { readonly kind: "uncertain" }
    | { readonly kind: "not_dispatched" }
    | { readonly kind: "authentication_required" }
  >;
}

/** A file list is one gesture, never an implicit multi-record command. */
export function admitEvidenceFile(
  files: FileList | readonly File[],
):
  | { readonly kind: "empty" }
  | { readonly kind: "accepted"; readonly file: File }
  | { readonly kind: "rejected"; readonly message: string } {
  if (files.length > 1)
    return { kind: "rejected", message: "Choose one file at a time." };
  const file = files[0];
  return file ? { kind: "accepted", file } : { kind: "empty" };
}
