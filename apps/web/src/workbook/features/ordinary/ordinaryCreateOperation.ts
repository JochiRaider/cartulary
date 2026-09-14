import type { ViewContract } from "@cartulary/view-contracts";
import type {
  WorkbookProtocolCreateViewRowReceipt,
  WorkbookProtocolCreateViewRowRequest,
} from "../../adapters/workbookProtocolTypes";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { OrdinaryCreateDraft } from "./ordinaryCreateContract";

export type OrdinaryCreateAttempt = Readonly<{
  operationID: "createViewRow";
  apiBase: string | undefined;
  path: string;
  pathParameters: Readonly<{ incident_id: string; view_schema_id: string }>;
  body: string;
  request: WorkbookProtocolCreateViewRowRequest;
  clientTxnId: string;
  authority: WorkbookMutationAuthority;
  target: ViewContract;
  draft: OrdinaryCreateDraft;
}>;
export type OrdinaryCreateReceipt = WorkbookProtocolCreateViewRowReceipt;
export type OrdinaryCreateOutcome =
  | {
      readonly kind: "accepted";
      readonly receipt: OrdinaryCreateReceipt;
      readonly status: number;
    }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export interface OrdinaryCreateTransport {
  capture(
    input: Readonly<{
      authority: WorkbookMutationAuthority;
      target: ViewContract;
      draft: OrdinaryCreateDraft;
      request: WorkbookProtocolCreateViewRowRequest;
      clientTxnId: string;
    }>,
  ): OrdinaryCreateAttempt;
  verify(target: ViewContract, signal: AbortSignal): Promise<void>;
  send(
    attempt: OrdinaryCreateAttempt,
    signal: AbortSignal,
  ): Promise<OrdinaryCreateOutcome>;
}
export type OrdinaryCreateEntry = Readonly<{
  attempt: OrdinaryCreateAttempt;
  phase: "preparing" | "submitting" | "uncertain" | "rejected" | "accepted";
  dispatched: boolean;
  transportPending: boolean;
  uncertain: boolean;
  receipt: OrdinaryCreateReceipt | null;
  status: number | null;
  failure: WorkbookOperationFailure | null;
  refresh: "none" | "required" | "refreshing" | "complete";
  message: string | null;
}>;
