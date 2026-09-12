import type { ViewContract } from "@cartulary/view-contracts";
import type { CreateViewRowResponse } from "../../adapters/indicatorCreateProtocol";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { IndicatorCreateDraftStore } from "./IndicatorCreateDraftStore";
import type {
  IndicatorCreateErrors,
  IndicatorCreateValues,
} from "./indicatorCreateModel";
import type {
  ObservationAttempt,
  ObservationAuthority,
  ObservationScope,
} from "./observationOperation";

export type IndicatorCreateAttempt = Readonly<{
  id: string;
  authority: ObservationAuthority;
  generation: number;
  observation: IndicatorObservation;
  contract: ViewContract;
  operation: "createViewRow";
  path: string;
  apiBase: string | undefined;
  body: string;
}>;
export type IndicatorCreateReceipt = Readonly<{
  status: 200 | 201;
  response: CreateViewRowResponse;
  row: WorkbookQueryRow;
}>;
export type IndicatorCreateOutcome =
  | { kind: "accepted"; receipt: IndicatorCreateReceipt }
  | { kind: "rejected"; failure: WorkbookOperationFailure }
  | { kind: "uncertain" };
export type IndicatorCreateOperation = Readonly<{
  attempt: IndicatorCreateAttempt;
  phase: "preparing" | "submitting" | "rejected" | "uncertain" | "accepted";
  transportPending: boolean;
  receipt: IndicatorCreateReceipt | null;
  failure: WorkbookOperationFailure | null;
  refresh: "pending" | "refreshing" | "required" | "complete";
  resolutionAttemptIds: readonly string[];
}>;
export type IndicatorCreateSnapshot = Readonly<{
  authority: ObservationAuthority | null;
  generation: number;
  revision: number;
  entries: readonly IndicatorCreateOperation[];
}>;
export type IndicatorCreateBinding = Readonly<{
  isCurrent: () => boolean;
  matchesDraft: () => boolean;
}>;
export type IndicatorCreateAdmission =
  | { kind: "admitted"; attempt: IndicatorCreateAttempt }
  | { kind: "invalid"; errors: IndicatorCreateErrors }
  | { kind: "unavailable"; message: string };
export type IndicatorCreateReconcile = (
  attempt: IndicatorCreateAttempt,
  receipt: IndicatorCreateReceipt,
  scope: ObservationScope,
) => Promise<void>;
export interface IndicatorCreateTransport {
  capture(
    authority: ObservationAuthority,
    generation: number,
    observation: IndicatorObservation,
    contract: ViewContract,
    values: IndicatorCreateValues,
    id: string,
  ): IndicatorCreateAttempt;
  send(
    attempt: IndicatorCreateAttempt,
    signal: AbortSignal,
  ): Promise<IndicatorCreateOutcome>;
}
export interface IndicatorCreateOwnerPort {
  readonly drafts: IndicatorCreateDraftStore;
  subscribe(listener: () => void): () => void;
  getSnapshot(): IndicatorCreateSnapshot;
  canSubmit(): boolean;
  canReplay(): boolean;
  busy(observationId: string): boolean;
  admit(
    observation: IndicatorObservation,
    contract: ViewContract,
    values: IndicatorCreateValues,
    binding: IndicatorCreateBinding,
  ): IndicatorCreateAdmission;
  execute(attempt: IndicatorCreateAttempt): Promise<void>;
  replay(id: string): Promise<void>;
  refresh(id: string): Promise<void>;
  associateResolution(id: string, attempt: ObservationAttempt): void;
}
