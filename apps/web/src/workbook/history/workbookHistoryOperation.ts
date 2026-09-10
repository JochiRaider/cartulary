import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { WorkbookInspectorSubject } from "../inspector/workbookInspectorSubject";
import type {
  RecordHistoryData,
  RecordHistoryRollbackTarget,
  WorkbookRecordHistoryPendingAction,
} from "../inspector/workbookRecordHistoryModel";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";
import type { HistoryLookupState } from "./HistoryActionLookup";
import type {
  HistoryPage,
  HistoryPageProvenance,
  HistoryPageRequest,
} from "./workbookHistoryPage";

export type HistoryAuthority = {
  readonly sessionIdentity?: string;
  readonly actorId: string;
  readonly incidentId: string;
  readonly role: WorkbookIncidentRole;
  readonly closed: boolean;
};
export type HistoryIntent = {
  readonly subject: WorkbookInspectorSubject;
  readonly pending: WorkbookRecordHistoryPendingAction;
  readonly provenance?: HistoryPageProvenance;
};
export type HistoryAttempt = HistoryIntent & {
  readonly id: string;
  readonly actorId: string;
  readonly incidentId: string;
  readonly operation: "delete" | "restore" | "rollback";
  readonly body: string;
};
export type HistoryReceipt = {
  readonly recordId: string;
  readonly incidentId: string;
  readonly rowVersion: number;
  readonly changeSetId: string;
} & (
  | {
      readonly kind: "delete" | "restore";
      readonly deleted: boolean;
      readonly deletedAt: string | null;
      readonly deletedByUserId: string | null;
    }
  | {
      readonly kind: "rollback";
      readonly target: RecordHistoryRollbackTarget;
      readonly targetChangeSetId?: string;
      readonly affectedRecordIds: readonly string[];
    }
);
export type HistoryTransportOutcome =
  | { readonly kind: "acknowledged"; readonly receipt: HistoryReceipt }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export type WorkbookRecordHistoryPort = {
  readonly load: (
    recordId: string,
    signal: AbortSignal,
    request?: HistoryPageRequest,
  ) => Promise<WorkbookOperationOutcome<HistoryPage>>;
  readonly send: (
    attempt: HistoryAttempt,
    signal: AbortSignal,
  ) => Promise<HistoryTransportOutcome>;
};
export type HistoryBinding = {
  readonly isCurrent: () => boolean;
  readonly coordinate: (
    recordId: string,
    signal: AbortSignal,
  ) => Promise<number | null>;
  readonly acknowledged: (receipt: HistoryReceipt) => void;
  readonly reconcile: (
    receipt: HistoryReceipt,
    current: () => boolean,
    history: RecordHistoryData,
  ) => Promise<void>;
};
export type HistoryOperation = {
  readonly attempt: HistoryAttempt;
  readonly phase:
    | "preparing"
    | "submitting"
    | "uncertain"
    | "rejected"
    | "acknowledged";
  readonly dispatched: boolean;
  readonly transportPending: boolean;
  readonly receipt: HistoryReceipt | null;
  readonly failure: WorkbookOperationFailure | null;
  readonly reconciliation: "pending" | "refreshing" | "required" | "complete";
  readonly reviewFailure: boolean;
  readonly currentHistory: RecordHistoryData | null;
  readonly checking?: HistoryLookupState;
  readonly reviewState?: HistoryLookupState | undefined;
};
export function historyActionPermitted(
  authority: HistoryAuthority | null,
  operation: HistoryAttempt["operation"],
): boolean {
  if (authority === null || authority.closed || !authority.actorId)
    return false;
  return (
    authority.role === "admin" ||
    authority.role === "reviewer" ||
    (operation === "delete" && authority.role === "editor")
  );
}
export { historyTargetEqual } from "./workbookHistoryItem";
export function historyOperationLabel(
  attempt: Pick<HistoryIntent, "pending">,
): string {
  const pending = attempt.pending;
  if (pending.kind === "destructive")
    return pending.operation === "delete"
      ? "Soft-delete row"
      : "Restore deleted row";
  if (pending.target.kind === "row_restore")
    return `Restore row fields to revision ${pending.target.restore_to_revision_no}`;
  return pending.target.kind === "history_entry"
    ? "Reverse history entry"
    : "Reverse change set";
}
