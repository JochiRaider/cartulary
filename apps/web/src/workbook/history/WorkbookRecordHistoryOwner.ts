import { observeAsyncOperation } from "../../services/asyncObservation";
import {
  buildRecordRollbackTargetFromHistoryAction,
  type RecordHistoryData,
} from "../inspector/workbookRecordHistoryModel";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";
import {
  type HistoryAttempt,
  type HistoryAuthority,
  type HistoryBinding,
  type HistoryIntent,
  type HistoryOperation,
  type HistoryReceipt,
  historyActionPermitted,
  historyTargetEqual,
  type WorkbookRecordHistoryPort,
} from "./workbookHistoryOperation";

const reviewRequired: WorkbookOperationFailure = {
  kind: "stale_target",
  message:
    "The record changed. Review current history and confirm the action again.",
};
const accessRequired: WorkbookOperationFailure = {
  kind: "authorization_lost",
  message: "Current access does not permit this history action.",
};
const empty: readonly HistoryOperation[] = [];

/** History actions have their own lifetime and never become autosave replay units. */
export class WorkbookRecordHistoryOwner {
  private authority: HistoryAuthority | null = null;
  private actorId: string | null = null;
  private entries = new Map<string, HistoryOperation>();
  private bindings = new Map<string, HistoryBinding>();
  private observations = new Map<string, { cancel: () => void }>();
  private versions = new Map<string, number>();
  private listeners = new Set<() => void>();
  private snapshot: readonly HistoryOperation[] = empty;
  private epoch = 0;
  private executing = new Set<string>();
  private surfaceRefreshes = new Map<string, () => Promise<void> | void>();
  registerSurface(viewSchemaId: string, refresh: () => Promise<void> | void) {
    this.surfaceRefreshes.set(viewSchemaId, refresh);
    return () => {
      if (this.surfaceRefreshes.get(viewSchemaId) === refresh)
        this.surfaceRefreshes.delete(viewSchemaId);
    };
  }
  private relatedProjectionRefresh:
    | ((receipt: HistoryReceipt) => Promise<void>)
    | null = null;
  registerRelatedProjectionRefresh(
    refresh: (receipt: HistoryReceipt) => Promise<void>,
  ) {
    this.relatedProjectionRefresh = refresh;
    return () => {
      if (this.relatedProjectionRefresh === refresh)
        this.relatedProjectionRefresh = null;
    };
  }
  private port: WorkbookRecordHistoryPort | null = null;

  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly observe: typeof observeAsyncOperation = observeAsyncOperation,
  ) {}

  get readable() {
    return this.authority !== null && Boolean(this.authority.role);
  }
  async refreshSurface(viewSchemaId: string) {
    const refresh = this.surfaceRefreshes.get(viewSchemaId);
    if (!this.readable || !refresh) throw new Error("surface unavailable");
    await refresh();
  }
  canReplace(id: string) {
    const entry = this.entry(id);
    return (
      entry?.phase === "rejected" &&
      entry.failure?.kind === "client_txn_conflict" &&
      entry.currentHistory !== null &&
      matchesReview(entry.attempt, entry.currentHistory) &&
      this.authorized(entry.attempt)
    );
  }
  configure(port: WorkbookRecordHistoryPort) {
    this.port = port;
  }
  setAuthority(authority: HistoryAuthority | null) {
    if (
      this.authority === authority ||
      (authority !== null &&
        this.authority !== null &&
        authority.actorId === this.authority.actorId &&
        authority.incidentId === this.authority.incidentId &&
        authority.role === this.authority.role &&
        authority.closed === this.authority.closed &&
        authority.sessionIdentity === this.authority.sessionIdentity)
    )
      return;
    if (
      authority !== null &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && this.actorId !== authority.actorId))
    )
      this.retire();
    this.epoch += 1;
    this.detachBindings();
    this.authority =
      authority?.incidentId === this.incidentId ? authority : null;
    if (this.authority !== null) this.actorId = this.authority.actorId;
    if (authority?.role === "") {
      this.retire();
      return;
    }
    this.publish();
  }
  private detachBindings() {
    for (const [id, binding] of this.bindings)
      this.bindings.set(id, { ...binding, isCurrent: () => false });
  }
  suspend() {
    this.epoch += 1;
    this.detachBindings();
    this.authority = null;
    this.publish();
  }
  closeIncident() {
    this.epoch += 1;
    this.detachBindings();
    if (this.authority !== null)
      this.authority = { ...this.authority, closed: true };
    for (const entry of this.entries.values()) {
      if (!entry.dispatched && entry.phase === "preparing")
        this.update(entry.attempt.id, {
          phase: "rejected",
          failure: {
            kind: "terminal",
            publicCode: "incident_closed",
            message: "Closed, read-only",
          },
        });
    }
    this.publish();
  }
  retire() {
    this.epoch += 1;
    for (const observation of this.observations.values()) observation.cancel();
    this.relatedProjectionRefresh = null;
    this.surfaceRefreshes.clear();
    this.observations.clear();
    this.executing.clear();
    this.entries.clear();
    this.bindings.clear();
    this.versions.clear();
    this.authority = null;
    this.actorId = null;
    this.publish();
  }
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  permitted(operation: HistoryAttempt["operation"]) {
    return historyActionPermitted(this.authority, operation);
  }
  get pendingCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        entry.phase === "preparing" ||
        entry.phase === "submitting" ||
        entry.phase === "uncertain" ||
        entry.reconciliation === "refreshing",
    ).length;
  }
  get blockedCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        entry.phase === "rejected" &&
        entry.failure?.kind === "client_txn_conflict",
    ).length;
  }
  acceptVersion(recordId: string, version: number) {
    if (
      Number.isSafeInteger(version) &&
      version > (this.versions.get(recordId) ?? 0)
    )
      this.versions.set(recordId, version);
  }
  latestVersion(recordId: string) {
    return this.versions.get(recordId) ?? null;
  }
  private publish() {
    this.snapshot = !this.authority?.role ? empty : [...this.entries.values()];
    for (const listener of this.listeners) listener();
  }
  private entry(id: string) {
    return [...this.entries.values()].find((entry) => entry.attempt.id === id);
  }
  private update(id: string, patch: Partial<HistoryOperation>) {
    const entry = this.entry(id);
    if (!entry) return;
    this.entries.set(entry.attempt.subject.recordId, { ...entry, ...patch });
    this.publish();
  }
  private owns(attempt: HistoryAttempt) {
    return (
      this.entry(attempt.id)?.attempt === attempt &&
      this.actorId === attempt.actorId
    );
  }
  private authorized(attempt: HistoryAttempt) {
    return (
      this.owns(attempt) &&
      this.authority?.actorId === attempt.actorId &&
      this.permitted(attempt.operation)
    );
  }

  async load(
    recordId: string,
    signal?: AbortSignal,
  ): Promise<WorkbookOperationOutcome<RecordHistoryData>> {
    const authority = this.authority;
    const epoch = this.epoch;
    const port = this.port;
    if (authority === null || !authority.role || port === null)
      return { kind: "rejected", failure: accessRequired };
    const observation = this.observe((requestSignal) =>
      port.load(recordId, requestSignal),
    );
    const cancel = () => observation.cancel();
    signal?.addEventListener("abort", cancel, { once: true });
    if (signal?.aborted) cancel();
    const result = await observation.result;
    signal?.removeEventListener("abort", cancel);
    if (epoch !== this.epoch || authority !== this.authority || signal?.aborted)
      return { kind: "rejected", failure: accessRequired };
    if (result.kind !== "completed")
      return {
        kind: "rejected",
        failure: {
          kind: "retryable",
          message: "History could not be read. Try refreshing history.",
        },
      };
    if (result.value.kind === "accepted") {
      const data = result.value.value;
      if (
        data.incident_id !== this.incidentId ||
        data.record_id !== recordId ||
        data.row_version < (this.latestVersion(recordId) ?? 0)
      )
        return { kind: "rejected", failure: reviewRequired };
      this.acceptVersion(recordId, data.row_version);
    }
    if (
      result.value.kind === "rejected" &&
      (result.value.failure.kind === "authentication_required" ||
        result.value.failure.kind === "authorization_lost" ||
        result.value.failure.publicCode === "record_not_found" ||
        result.value.failure.publicCode === "incident_not_found")
    )
      this.suspend();
    return result.value;
  }

  admit(intent: HistoryIntent, binding: HistoryBinding): HistoryAttempt | null {
    const authority = this.authority;
    const operation =
      intent.pending.kind === "rollback"
        ? "rollback"
        : intent.pending.operation;
    const previous = this.entries.get(intent.subject.recordId);
    if (
      !authority ||
      !this.permitted(operation) ||
      !binding.isCurrent() ||
      (previous &&
        (previous.transportPending ||
          ["preparing", "submitting", "uncertain"].includes(previous.phase)))
    )
      return null;
    if (
      intent.subject.recordId !== intent.pending.recordId ||
      intent.subject.rowVersion !== intent.pending.rowVersion
    )
      return null;
    let id: string;
    try {
      id = this.ids.create(`history-${operation}`);
    } catch {
      return null;
    }
    const reason =
      operation === "delete"
        ? "Deleted from workbook history"
        : operation === "restore"
          ? "Restored from workbook history"
          : "Rollback from workbook history";
    const body = JSON.stringify({
      base_row_version: intent.pending.rowVersion,
      client_txn_id: id,
      reason,
      ...(intent.pending.kind === "rollback"
        ? { target: intent.pending.target }
        : {}),
    });
    const attempt: HistoryAttempt = freeze({
      ...structuredClone(intent),
      actorId: authority.actorId,
      incidentId: this.incidentId,
      id,
      operation,
      body,
    });
    this.entries.set(intent.subject.recordId, {
      attempt,
      phase: "preparing",
      dispatched: false,
      transportPending: false,
      receipt: null,
      failure: null,
      reconciliation: "pending",
      currentHistory: null,
      reviewFailure: false,
    });
    this.bindings.set(id, binding);
    this.publish();
    return attempt;
  }

  async execute(attempt: HistoryAttempt, replay = false): Promise<void> {
    const entry = this.entry(attempt.id);
    const binding = this.bindings.get(attempt.id);
    const port = this.port;
    if (!entry || entry.transportPending || this.executing.has(attempt.id))
      return;
    if (!port || !this.authorized(attempt)) {
      this.update(attempt.id, {
        phase: replay ? "uncertain" : "rejected",
        failure: accessRequired,
      });
      return;
    }
    const dispatchAuthority = this.authority;
    const dispatchEpoch = this.epoch;
    this.executing.add(attempt.id);
    this.update(attempt.id, { phase: "preparing", failure: null });
    const preparation = this.observe(async (signal) => {
      const version = binding
        ? await binding.coordinate(attempt.subject.recordId, signal)
        : this.latestVersion(attempt.subject.recordId);
      if (signal.aborted || !this.authorized(attempt)) return false;
      if (!replay && (version === null || version > attempt.pending.rowVersion))
        return false;
      if (replay) return true;
      const history = await this.load(attempt.subject.recordId, signal);
      return (
        history.kind === "accepted" && matchesReview(attempt, history.value)
      );
    });
    this.observations.set(attempt.id, preparation);
    const prepared = await preparation.result;
    if (!this.owns(attempt)) {
      this.executing.delete(attempt.id);
      return;
    }
    if (
      prepared.kind !== "completed" ||
      !prepared.value ||
      !this.authorized(attempt) ||
      this.authority !== dispatchAuthority ||
      this.epoch !== dispatchEpoch
    ) {
      this.update(attempt.id, {
        phase: replay ? "uncertain" : "rejected",
        failure: this.authorized(attempt) ? reviewRequired : accessRequired,
      });
      this.executing.delete(attempt.id);
      return;
    }
    this.update(attempt.id, {
      phase: "submitting",
      dispatched: true,
      transportPending: true,
    });
    const observation = this.observe(async (signal) => {
      const outcome = await port.send(attempt, signal);
      if (!this.owns(attempt)) return;
      const current = this.entry(attempt.id);
      if (current?.phase === "acknowledged") return;
      if (outcome.kind === "acknowledged") {
        const receipt = outcome.receipt;
        if (
          receipt.incidentId !== attempt.incidentId ||
          receipt.recordId !== attempt.subject.recordId ||
          receipt.kind !== attempt.operation
        ) {
          this.update(attempt.id, { phase: "uncertain" });
          return;
        }
        // Record the receipt before invoking any disposable surface callback.
        this.update(attempt.id, {
          phase: "acknowledged",
          receipt,
          failure: null,
          reconciliation: "pending",
        });
        const newer =
          (this.latestVersion(receipt.recordId) ?? 0) > receipt.rowVersion;
        this.acceptVersion(receipt.recordId, receipt.rowVersion);
        if (
          !newer &&
          this.authority === dispatchAuthority &&
          this.epoch === dispatchEpoch &&
          this.authorized(attempt) &&
          binding?.isCurrent()
        )
          binding.acknowledged(receipt);
      } else if (outcome.kind === "rejected") {
        // A later rejection cannot disprove an earlier indeterminate dispatch.
        this.update(attempt.id, {
          phase:
            replay && outcome.failure.kind !== "client_txn_conflict"
              ? "uncertain"
              : "rejected",
          failure: outcome.failure,
        });
        if (outcome.failure.publicCode === "incident_closed")
          this.closeIncident();
        if (
          outcome.failure.kind === "authentication_required" ||
          outcome.failure.kind === "authorization_lost" ||
          outcome.failure.publicCode === "record_not_found" ||
          outcome.failure.publicCode === "incident_not_found"
        )
          this.suspend();
      } else this.update(attempt.id, { phase: "uncertain" });
    });
    this.observations.set(attempt.id, observation);
    const result = await observation.result;
    if (
      this.owns(attempt) &&
      result.kind !== "completed" &&
      this.entry(attempt.id)?.phase === "submitting"
    )
      this.update(attempt.id, { phase: "uncertain" });
    void observation.settled.then(() => {
      if (!this.owns(attempt)) return;
      this.observations.delete(attempt.id);
      this.executing.delete(attempt.id);
      this.update(attempt.id, { transportPending: false });
      if (this.entry(attempt.id)?.phase === "acknowledged")
        void this.refresh(attempt.id);
    });
  }

  async replay(id: string) {
    const entry = this.entry(id);
    if (
      !entry ||
      entry.phase !== "uncertain" ||
      entry.transportPending ||
      !this.authorized(entry.attempt)
    )
      return;
    // Reservation is synchronous; a second same-turn replay observes preparing.
    this.update(id, { phase: "preparing" });
    await this.execute(entry.attempt, true);
  }
  async review(id: string) {
    const entry = this.entry(id);
    if (!entry) return;
    this.update(id, { currentHistory: null, reviewFailure: false });
    const history = await this.load(entry.attempt.subject.recordId);
    if (history.kind === "accepted")
      this.update(id, { currentHistory: history.value });
    else this.update(id, { reviewFailure: true });
  }
  async retryWithNewId(id: string) {
    const entry = this.entry(id);
    const binding = this.bindings.get(id);
    if (
      !entry ||
      entry.phase !== "rejected" ||
      entry.failure?.kind !== "client_txn_conflict" ||
      entry.transportPending ||
      !entry.currentHistory ||
      !matchesReview(entry.attempt, entry.currentHistory) ||
      !binding ||
      !this.authorized(entry.attempt)
    )
      return;
    const attempt = this.admit(entry.attempt, binding);
    if (attempt) await this.execute(attempt);
  }
  async refresh(id: string) {
    const entry = this.entry(id);
    if (
      !entry?.receipt ||
      entry.phase !== "acknowledged" ||
      entry.reconciliation === "refreshing" ||
      entry.reconciliation === "complete"
    )
      return;
    const authority = this.authority;
    const epoch = this.epoch;
    if (!authority) {
      this.update(id, { reconciliation: "required" });
      return;
    }
    const originalBinding = this.bindings.get(id);
    const refreshSurface = this.surfaceRefreshes.get(
      entry.attempt.subject.viewSchemaId,
    );
    const binding = originalBinding?.isCurrent()
      ? originalBinding
      : refreshSurface
        ? {
            isCurrent: () =>
              this.surfaceRefreshes.get(entry.attempt.subject.viewSchemaId) ===
              refreshSurface,
            reconcile: async () => {
              await refreshSurface();
            },
          }
        : undefined;
    this.update(id, { reconciliation: "refreshing" });
    let observing = true;
    const current = () =>
      observing &&
      this.owns(entry.attempt) &&
      this.authority === authority &&
      this.epoch === epoch &&
      binding?.isCurrent() === true;
    try {
      const history = await this.load(entry.attempt.subject.recordId);
      if (history.kind !== "accepted") throw new Error("history unavailable");
      this.update(id, { currentHistory: history.value });
      if (!binding || !current()) throw new Error("surface detached");
      const observation = this.observe(async () => {
        if (
          entry.receipt?.kind === "rollback" &&
          entry.receipt.affectedRecordIds.length > 1
        )
          await this.relatedProjectionRefresh?.(entry.receipt);
        if (!current()) throw new Error("surface detached");
        await binding.reconcile(
          entry.receipt as HistoryReceipt,
          current,
          history.value,
        );
      });
      const result = await observation.result;
      if (result.kind !== "completed" || !current())
        throw new Error("refresh unavailable");
      this.update(id, { reconciliation: "complete" });
    } catch {
      this.update(id, { reconciliation: "required" });
    } finally {
      observing = false;
    }
  }
  bindRecovery(id: string, binding: HistoryBinding) {
    if (this.entry(id)) this.bindings.set(id, binding);
  }
  dismiss(id: string) {
    const entry = this.entry(id);
    if (
      !entry ||
      entry.transportPending ||
      entry.phase === "preparing" ||
      entry.phase === "submitting" ||
      entry.phase === "uncertain"
    )
      return;
    if (entry.phase === "acknowledged" && entry.reconciliation !== "complete")
      return;
    this.entries.delete(entry.attempt.subject.recordId);
    this.bindings.delete(id);
    this.publish();
  }
}

function matchesReview(
  attempt: HistoryAttempt,
  data: RecordHistoryData,
): boolean {
  if (
    data.row_version !== attempt.pending.rowVersion ||
    data.deleted !== (attempt.subject.kind === "deleted")
  )
    return false;
  if (attempt.operation === "delete") return !data.deleted;
  if (attempt.operation === "restore") return data.deleted;
  const pending = attempt.pending;
  if (data.deleted || pending.kind !== "rollback") return false;
  const item = data.items.find(
    (item) => item.history_item_ref === pending.historyItemRef,
  );
  const target =
    item && buildRecordRollbackTargetFromHistoryAction(item, pending.action);
  return target != null && historyTargetEqual(target, pending.target);
}
function freeze<T>(value: T): T {
  if (value !== null && typeof value === "object") {
    for (const child of Object.values(value)) freeze(child);
    Object.freeze(value);
  }
  return value;
}
