import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type {
  DecisionAuthority,
  DecisionSupersessionReview,
} from "./decisionSupersessionModel";

export interface DecisionRecordWriteBoundary {
  begin(recordIds: readonly string[]): (() => void) | null;
  acceptRow(row: WorkbookQueryRow): WorkbookQueryRow | null;
}

export type DecisionSupersessionAttempt = Readonly<{
  id: string;
  review: DecisionSupersessionReview;
  path: string;
  apiBase: string | undefined;
  body: string;
}>;
export type DecisionSupersessionReceipt = Readonly<{
  view_schema_id: "cartulary.view.decisions.v1";
  change_set_id: string;
  target_record_id: string;
  superseding_record_id: string;
  target_row_version: number;
  superseding_row_version: number;
  target_status: "superseded" | "executed";
  reason: string;
}>;
export type DecisionSupersessionOutcome =
  | {
      readonly kind: "acknowledged";
      readonly receipt: DecisionSupersessionReceipt;
    }
  | { readonly kind: "uncertain" }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure };
export type DecisionCandidatePage = Readonly<{
  rows: readonly WorkbookQueryRow[];
  nextCursor: string | null;
  hasMore: boolean;
}>;
export interface DecisionSupersessionReadPort {
  page(
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<WorkbookPortResult<DecisionCandidatePage>>;
}
export interface DecisionSupersessionTransportPort
  extends DecisionSupersessionReadPort {
  capture(
    review: DecisionSupersessionReview,
    id: string,
  ): DecisionSupersessionAttempt;
  send(
    attempt: DecisionSupersessionAttempt,
    signal: AbortSignal,
  ): Promise<DecisionSupersessionOutcome>;
}
export type DecisionSupersessionOperation = Readonly<{
  attempt: DecisionSupersessionAttempt;
  phase: "preparing" | "submitting" | "uncertain" | "rejected" | "acknowledged";
  transportPending: boolean;
  receipt: DecisionSupersessionReceipt | null;
  failure: WorkbookOperationFailure | null;
  reconciliation: "pending" | "refreshing" | "required" | "complete";
}>;
export type DecisionSupersessionBinding = Readonly<{
  isCurrent: () => boolean;
  matchesReview: () => boolean;
  reconcile: (receipt: DecisionSupersessionReceipt) => Promise<void>;
}>;
export type DecisionSupersessionSnapshot = Readonly<{
  authority: DecisionAuthority | null;
  generation: number;
  revision: number;
  entries: readonly DecisionSupersessionOperation[];
}>;
export interface DecisionSupersessionOwnerPort
  extends DecisionSupersessionReadPort {
  subscribe(listener: () => void): () => void;
  getSnapshot(): DecisionSupersessionSnapshot;
  canSubmit(): boolean;
  acceptRow(row: WorkbookQueryRow): WorkbookQueryRow | null;
  latestRow(id: string): WorkbookQueryRow | null;
  latestVersion(id: string): number | null;
  blocksRecord(id: string): boolean;
  admit(
    review: DecisionSupersessionReview,
    binding: DecisionSupersessionBinding,
  ): DecisionSupersessionAttempt | null;
  execute(attempt: DecisionSupersessionAttempt): Promise<void>;
  replay(id: string): Promise<void>;
  refresh(id: string): Promise<void>;
  dismiss(id: string): void;
}
