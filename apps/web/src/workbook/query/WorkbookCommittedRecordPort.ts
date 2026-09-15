import type { WorkbookPendingMutationAccepted } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
/** First-class record evidence only. Child-resource versions never enter here. */
export interface WorkbookCommittedRecordPort {
  subscribe(listener: () => void): () => void;
  getSnapshot(): { readonly authority: object | null };
  latestVersion(recordId: string): number | null;
  latestRow(recordId: string): WorkbookQueryRow | null;
  acceptRow(row: WorkbookQueryRow): WorkbookQueryRow | null;
}

/** Complete mutation evidence retained independently of subsequent query reads. */
export interface WorkbookAcceptedRecordPort
  extends WorkbookCommittedRecordPort {
  latestReceipt(recordId: string): WorkbookPendingMutationAccepted | null;
  observeReceipt(receipt: WorkbookPendingMutationAccepted): void;
}
