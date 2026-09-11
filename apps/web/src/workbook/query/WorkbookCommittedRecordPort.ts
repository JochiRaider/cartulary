import type { WorkbookQueryRow } from "./WorkbookQueryRow";
/** First-class record evidence only. Child-resource versions never enter here. */
export interface WorkbookCommittedRecordPort {
  subscribe(listener: () => void): () => void;
  getSnapshot(): { readonly authority: object | null };
  latestVersion(recordId: string): number | null;
  latestRow(recordId: string): WorkbookQueryRow | null;
  acceptRow(row: WorkbookQueryRow): WorkbookQueryRow | null;
}
