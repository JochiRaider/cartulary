import type { ObservationReceipt } from "../../adapters/observationProtocol";

export type { ObservationReceipt } from "../../adapters/observationProtocol";

import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { ObservationDraftStore } from "./ObservationDraftStore";
import type {
  ObservationSelection,
  ObservationSource,
  ObservationType,
} from "./observationModel";

export type ObservationAuthority = Readonly<{
  actorId: string;
  incidentId: string;
  sessionIdentity: string;
  role: WorkbookIncidentRole;
  closed: boolean;
}>;
export type ObservationSubject = Readonly<{
  kind: "source" | "indicator";
  recordId: string;
}>;
export type ObservationIntent =
  | Readonly<{
      action: "create";
      source: ObservationSource;
      selection: ObservationSelection;
      parsedType?: ObservationType;
      targetId?: string;
    }>
  | Readonly<{
      action: "resolve";
      observation: IndicatorObservation;
      targetId: string;
    }>
  | Readonly<{
      action: "dismiss" | "restore";
      observation: IndicatorObservation;
    }>;
export type ObservationAttempt = Readonly<{
  id: string;
  authority: ObservationAuthority;
  generation: number;
  intent: ObservationIntent;
  path: string;
  body: string;
  apiBase: string | undefined;
  operation:
    | "createManualIndicatorObservation"
    | "resolveIndicatorObservation"
    | "dismissIndicatorObservation"
    | "restoreIndicatorObservation";
}>;
export type ObservationPage<T> = Readonly<{
  items: readonly T[];
  hasMore: boolean;
  nextCursor: string | null;
}>;
export type ObservationOutcome =
  | { kind: "acknowledged"; receipt: ObservationReceipt }
  | { kind: "uncertain" }
  | { kind: "rejected"; failure: WorkbookOperationFailure };
export type ObservationScope = Readonly<{
  signal: AbortSignal;
  isCurrent: () => boolean;
}>;
export type ObservationBinding = Readonly<{
  isCurrent: () => boolean;
  matchesDraft: () => boolean;
  prepare: (signal: AbortSignal) => Promise<boolean>;
  reconcile: () => Promise<void>;
}>;
export type ObservationSourcePort = Readonly<{
  subscribe: (listener: () => void) => () => void;
  fields: readonly { fieldKey: string; label: string; value: string }[];
  source: (fieldKey: string) => ObservationSource | null;
  ready: () => boolean;
  prepare: (source: ObservationSource, signal: AbortSignal) => Promise<boolean>;
}>;
export type ObservationOperation = Readonly<{
  attempt: ObservationAttempt;
  phase: "preparing" | "submitting" | "uncertain" | "rejected" | "acknowledged";
  transportPending: boolean;
  failure: WorkbookOperationFailure | null;
  receipt: ObservationReceipt | null;
  reconciliation: "pending" | "refreshing" | "required" | "complete";
}>;
export type ObservationSnapshot = Readonly<{
  authority: ObservationAuthority | null;
  generation: number;
  revision: number;
  entries: readonly ObservationOperation[];
}>;
export interface ObservationReadPort {
  observations(
    subject: ObservationSubject,
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<WorkbookPortResult<ObservationPage<IndicatorObservation>>>;
  records(
    viewSchemaId: string,
    query: WorkbookQueryState,
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<WorkbookPortResult<ObservationPage<WorkbookQueryRow>>>;
}
export interface ObservationTransportPort {
  capture(
    authority: ObservationAuthority,
    generation: number,
    intent: ObservationIntent,
    id: string,
  ): ObservationAttempt;
  send(
    attempt: ObservationAttempt,
    signal: AbortSignal,
  ): Promise<ObservationOutcome>;
}
export interface ObservationOwnerPort extends ObservationReadPort {
  readonly drafts: ObservationDraftStore;
  subscribe(listener: () => void): () => void;
  getSnapshot(): ObservationSnapshot;
  canSubmit(): boolean;
  canReplay(): boolean;
  busy(intent: ObservationIntent): boolean;
  admit(
    intent: ObservationIntent,
    binding: ObservationBinding,
  ): ObservationAttempt | null;
  execute(attempt: ObservationAttempt): Promise<void>;
  replay(id: string): Promise<void>;
  refresh(id: string): Promise<void>;
  dismiss(id: string): void;
  latestVersion(id: string): number | null;
  latestRow(id: string): WorkbookQueryRow | null;
}
export function observationIntentSource(intent: ObservationIntent): string {
  return intent.action === "create"
    ? intent.source.recordId
    : intent.observation.source_record_id;
}
export function observationIntentKey(intent: ObservationIntent): string {
  return intent.action === "create"
    ? `create:${intent.source.recordId}`
    : `observation:${intent.observation.observation_id}`;
}
