import type { WorkbookProtocolMergeReceipt } from "../../adapters/workbookProtocolTypes";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { EntityMergeReview } from "./entityMergeReview";

export type EntityMergeAttempt = {
  readonly id: string;
  readonly review: EntityMergeReview;
  readonly path: string;
  readonly apiBase: string | undefined;
  readonly body: string;
};
type Immutable<T> = T extends object
  ? { readonly [K in keyof T]: Immutable<T[K]> }
  : T;
export type EntityMergeReceipt = Immutable<WorkbookProtocolMergeReceipt>;
export type EntityMergeTransportOutcome =
  | { readonly kind: "acknowledged"; readonly receipt: EntityMergeReceipt }
  | { readonly kind: "uncertain" }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure };
export interface WorkbookEntityMergePort {
  capture(review: EntityMergeReview, transactionId: string): EntityMergeAttempt;
  send(
    attempt: EntityMergeAttempt,
    signal: AbortSignal,
  ): Promise<EntityMergeTransportOutcome>;
}
export type EntityMergeOperation = {
  readonly attempt: EntityMergeAttempt;
  readonly phase:
    | "preparing"
    | "submitting"
    | "uncertain"
    | "rejected"
    | "acknowledged";
  readonly transportPending: boolean;
  readonly receipt: EntityMergeReceipt | null;
  readonly failure: WorkbookOperationFailure | null;
  readonly reconciliation: "pending" | "refreshing" | "required" | "complete";
};
export type EntityMergeBinding = {
  readonly isCurrent: () => boolean;
  readonly matchesReview: () => boolean;
  readonly acknowledged: (receipt: EntityMergeReceipt) => void;
  readonly reconcile: (receipt: EntityMergeReceipt) => Promise<void>;
};

// Admission and coordination belong to the existing workbook runtime. This
// port neither creates a client lock API nor enqueues merges in autosave.
export type EntityMergeCoordination = {
  readonly canReserve: (review: EntityMergeReview) => boolean;
  readonly coordinate: (
    review: EntityMergeReview,
    signal: AbortSignal,
  ) => Promise<boolean>;
};
