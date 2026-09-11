import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type {
  IndicatorLifecycleInterval,
  IndicatorLifecycleReceipt,
  IndicatorLifecycleValues,
} from "../../adapters/indicatorLifecycleProtocol";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { IndicatorLifecycleDraftStore } from "./IndicatorLifecycleDraftStore";
import type { LifecycleDraft } from "./indicatorLifecycleModel";

export type LifecycleAuthority = Readonly<{
  actorId: string;
  incidentId: string;
  sessionIdentity: string;
  role: WorkbookIncidentRole;
  closed: boolean;
}>;
export type LifecyclePage<T> = Readonly<{
  items: readonly T[];
  nextCursor: string | null;
  hasMore: boolean;
}>;
export type LifecycleAttempt = Readonly<{
  id: string;
  authority: LifecycleAuthority;
  generation: number;
  draft: LifecycleDraft;
  values: IndicatorLifecycleValues;
  apiBase: string | undefined;
  path: string;
  body: string;
}>;
export type LifecycleOutcome =
  | {
      readonly kind: "acknowledged";
      readonly receipt: IndicatorLifecycleReceipt;
    }
  | { readonly kind: "uncertain" }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure };
export type LifecycleOperation = Readonly<{
  attempt: LifecycleAttempt;
  phase: "preparing" | "submitting" | "uncertain" | "rejected" | "acknowledged";
  transportPending: boolean;
  receipt: IndicatorLifecycleReceipt | null;
  failure: WorkbookOperationFailure | null;
  reconciliation: "pending" | "refreshing" | "required" | "complete";
}>;
export type LifecycleScope = Readonly<{
  signal: AbortSignal;
  isCurrent: () => boolean;
}>;
export type LifecycleBinding = Readonly<{
  isCurrent: () => boolean;
  matchesDraft?: () => boolean;
  reconcile: () => Promise<void>;
}>;
export interface IndicatorLifecycleReadPort {
  intervals(
    recordId: string,
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<WorkbookPortResult<LifecyclePage<IndicatorLifecycleInterval>>>;
  records(
    viewSchemaId: string,
    query: WorkbookQueryState,
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<WorkbookPortResult<LifecyclePage<WorkbookQueryRow>>>;
}
export interface IndicatorLifecycleTransportPort
  extends IndicatorLifecycleReadPort {
  capture(
    authority: LifecycleAuthority,
    generation: number,
    draft: LifecycleDraft,
    values: IndicatorLifecycleValues,
    id: string,
  ): LifecycleAttempt;
  send(
    attempt: LifecycleAttempt,
    signal: AbortSignal,
  ): Promise<LifecycleOutcome>;
}
export type LifecycleSnapshot = Readonly<{
  authority: LifecycleAuthority | null;
  generation: number;
  revision: number;
  entries: readonly LifecycleOperation[];
}>;
export interface IndicatorLifecycleOwnerPort
  extends IndicatorLifecycleReadPort {
  readonly drafts: IndicatorLifecycleDraftStore;
  subscribe(listener: () => void): () => void;
  getSnapshot(): LifecycleSnapshot;
  canSubmit(): boolean;
  canReplay(): boolean;
  blocksRecord(recordId: string): boolean;
  latestVersion(recordId: string): number | null;
  latestRow(recordId: string): WorkbookQueryRow | null;
  acceptRow(row: WorkbookQueryRow): WorkbookQueryRow | null;
  admit(
    draft: LifecycleDraft,
    binding: LifecycleBinding,
  ): LifecycleAttempt | null;
  execute(attempt: LifecycleAttempt): Promise<void>;
  replay(id: string): Promise<void>;
  refresh(id: string): Promise<void>;
  review(recordId: string): Promise<boolean>;
  dismiss(id: string): void;
}
