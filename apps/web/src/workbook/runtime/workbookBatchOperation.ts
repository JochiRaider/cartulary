import type {
  WorkbookProtocolBulkRequest,
  WorkbookProtocolPasteRequest,
} from "../adapters/workbookProtocolTypes";
import type { WorkbookViewApiRow } from "../models/workbookContractRows";
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookSameFieldConflictPayload } from "./workbookConflictModel";

export type WorkbookBatchRequest =
  | {
      readonly operation: "pasteWorkbookClipboard";
      readonly request: Omit<WorkbookProtocolPasteRequest, "client_txn_id">;
    }
  | {
      readonly operation: "applyWorkbookBulkMutation";
      readonly request: Omit<WorkbookProtocolBulkRequest, "client_txn_id">;
    };

export type WorkbookBatchPlan = WorkbookBatchRequest & {
  readonly recordIds: readonly string[];
  readonly entityType?: "host" | "identity";
};
export type WorkbookBatchReceipt = {
  readonly viewSchemaId: string;
  readonly changeSetId: string | null;
  readonly rows: readonly WorkbookViewApiRow[];
  readonly conflicts: readonly WorkbookSameFieldConflictPayload[];
};
export type WorkbookBatchAttempt = {
  readonly id: string;
  readonly authority: WorkbookMutationAuthority;
  readonly plan: WorkbookBatchPlan;
  readonly apiBase: string | undefined;
  readonly path: string;
  readonly body: string;
};
export type WorkbookBatchTransportOutcome =
  | { readonly kind: "acknowledged"; readonly receipt: WorkbookBatchReceipt }
  | { readonly kind: "uncertain" }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure };
export interface WorkbookBatchTransport {
  capture(
    plan: WorkbookBatchPlan,
    authority: WorkbookMutationAuthority,
    id: string,
  ): WorkbookBatchAttempt;
  send(
    attempt: WorkbookBatchAttempt,
    signal: AbortSignal,
  ): Promise<WorkbookBatchTransportOutcome>;
}
export type WorkbookBatchEntry = {
  readonly id: string;
  readonly plan: WorkbookBatchPlan;
  readonly phase:
    | "waiting"
    | "preparing"
    | "submitting"
    | "uncertain"
    | "rejected"
    | "acknowledged";
  readonly attempt: WorkbookBatchAttempt | null;
  readonly receipt: WorkbookBatchReceipt | null;
  readonly failure: WorkbookOperationFailure | null;
  readonly transportPending: boolean;
  readonly reconciliation: "pending" | "refreshing" | "required" | "complete";
};
export type WorkbookBatchAdmission = {
  /** Same event delivered twice uses the same object; another gesture gets another. */
  readonly delivery: object;
  /** Captured prerequisite save work only; never waits for later edits. */
  readonly ready?: Promise<void>;
};
export type WorkbookBatchSnapshot = {
  readonly authority: WorkbookMutationAuthority | null;
  readonly entries: readonly WorkbookBatchEntry[];
  readonly admissionError: string | null;
};
