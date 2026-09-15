import type { WorkbookProtocolCreateViewRowReceipt } from "../../adapters/workbookProtocolTypes";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { EvidenceFileOutcome } from "./evidenceFileOperation";

export type TimelineFileSource = Readonly<{
  label?: string;
  key: string;
  recordId: string | null;
  rowVersion: number | null;
}>;
export type TimelineFileLinkAttempt = Readonly<{
  authority: WorkbookMutationAuthority;
  source: WorkbookQueryRow;
  evidenceRecordId: string;
  clientTxnId: string;
  path: string;
  body: string;
  apiBase: string | undefined;
}>;
export interface TimelineFileLinkTransport {
  capture(
    authority: WorkbookMutationAuthority,
    source: WorkbookQueryRow,
    evidenceRecordId: string,
    clientTxnId: string,
  ): TimelineFileLinkAttempt;
  send(
    attempt: TimelineFileLinkAttempt,
    signal: AbortSignal,
  ): Promise<EvidenceFileOutcome<WorkbookProtocolCreateViewRowReceipt>>;
}
/** The existing Timeline creation owner is the only local-draft promotion authority. */
export interface TimelineFileDraftPort {
  subscribe(listener: () => void): () => void;
  resolve(key: string):
    | { readonly kind: "draft" }
    | { readonly kind: "pending" }
    | { readonly kind: "unavailable" }
    | {
        readonly kind: "promoted";
        readonly receipt: WorkbookPendingMutationAccepted;
      };
  attachEvidence(key: string, evidenceRecordId: string): void;
}
