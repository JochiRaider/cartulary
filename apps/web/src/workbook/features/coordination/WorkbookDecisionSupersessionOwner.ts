import { observeAsyncOperation } from "../../../services/asyncObservation";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import {
  type DecisionAuthority,
  type DecisionSupersessionReview,
  decisionReviewValid,
} from "./decisionSupersessionModel";
import type {
  DecisionSupersessionAttempt,
  DecisionSupersessionBinding,
  DecisionSupersessionOperation,
  DecisionSupersessionOwnerPort,
  DecisionSupersessionReceipt,
  DecisionSupersessionSnapshot,
  DecisionSupersessionTransportPort,
} from "./decisionSupersessionOperation";

export type DecisionReconciliationScope = Readonly<{
  signal: AbortSignal;
  isCurrent: () => boolean;
}>;
type Reconcile = (
  receipt: DecisionSupersessionReceipt,
  scope: DecisionReconciliationScope,
) => Promise<void>;
const reviewRequired: WorkbookOperationFailure = {
  kind: "stale_target",
  message:
    "The reviewed Decisions changed or earlier writes could not finish. Refresh and review again. No supersession was sent.",
};

/** One incident/account's explicit Decision actions, outside the autosave FIFO. */
export class WorkbookDecisionSupersessionOwner
  implements DecisionSupersessionOwnerPort
{
  private authority: DecisionAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private revision = 0;
  private readonly entries = new Map<string, DecisionSupersessionOperation>();
  private readonly versions = new Map<string, number>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  private readonly bindings = new Map<string, DecisionSupersessionBinding>();
  private readonly observations = new Map<string, { cancel: () => void }>();
  private readonly executing = new Set<string>();
  private readonly refreshes = new Map<string, number>();
  private readonly listeners = new Set<() => void>();
  private port: DecisionSupersessionTransportPort | null = null;
  private reconcile: Reconcile | null = null;
  private authorityUncertain: (() => void) | undefined;
  private snapshot: DecisionSupersessionSnapshot = {
    authority: null,
    generation: 0,
    revision: 0,
    entries: [],
  };
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly coordination: {
      canReserve: (review: DecisionSupersessionReview) => boolean;
      coordinate: (
        review: DecisionSupersessionReview,
        signal: AbortSignal,
      ) => Promise<boolean>;
    },
    private readonly observe: typeof observeAsyncOperation = observeAsyncOperation,
  ) {}
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  configure(
    port: DecisionSupersessionTransportPort,
    authorityUncertain?: () => void,
  ) {
    this.port = port;
    this.authorityUncertain = authorityUncertain;
  }
  registerReconciliation(reconcile: Reconcile) {
    this.reconcile = reconcile;
    return () => {
      if (this.reconcile === reconcile) this.reconcile = null;
    };
  }
  setAuthority(authority: DecisionAuthority | null) {
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && authority.actorId !== this.actorId))
    )
      this.retire();
    this.generation++;
    this.bindings.clear();
    for (const [id, observation] of this.observations)
      if (this.entries.get(id)?.phase === "preparing") observation.cancel();
    this.authority =
      authority?.incidentId === this.incidentId && authority.role
        ? Object.freeze({ ...authority })
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.publish();
  }
  canSubmit() {
    return (
      !!this.authority?.actorId &&
      !!this.authority.sessionIdentity &&
      !this.authority.closed &&
      (this.authority.role === "reviewer" || this.authority.role === "admin")
    );
  }
  suspendForAuthorityRecovery() {
    this.suspend();
    this.authorityUncertain?.();
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  retire() {
    this.generation++;
    for (const observation of this.observations.values()) observation.cancel();
    this.observations.clear();
    this.executing.clear();
    this.entries.clear();
    this.bindings.clear();
    this.versions.clear();
    this.rows.clear();
    this.refreshes.clear();
    this.reconcile = null;
    this.authority = null;
    this.actorId = null;
    this.publish();
  }
  acceptVersion(id: string, version: number) {
    if (
      Number.isSafeInteger(version) &&
      version > (this.versions.get(id) ?? 0)
    ) {
      this.versions.set(id, version);
      this.publish();
    }
  }
  latestVersion(id: string) {
    return this.versions.get(id) ?? null;
  }
  latestRow(id: string) {
    return this.authority ? (this.rows.get(id) ?? null) : null;
  }
  acceptRow(row: WorkbookQueryRow): WorkbookQueryRow | null {
    if (
      !this.authority ||
      !Number.isSafeInteger(row.row_version) ||
      row.row_version < 1
    )
      return null;
    const prior = this.rows.get(row.record_id);
    if (row.row_version < (this.latestVersion(row.record_id) ?? 0))
      return prior ?? null;
    if (prior && prior.row_version >= row.row_version) return prior;
    const accepted = freeze(structuredClone(row));
    this.rows.set(row.record_id, accepted);
    this.versions.set(row.record_id, row.row_version);
    this.publish();
    return accepted;
  }
  async page(cursor: string | null, signal: AbortSignal) {
    const generation = this.generation;
    if (!this.authority || !this.port) return { kind: "aborted" as const };
    const result = await this.port.page(cursor, signal);
    if (signal.aborted || generation !== this.generation)
      return { kind: "aborted" as const };
    if (
      result.kind === "rejected" &&
      workbookFailureLifecycle(result.failure).kind === "authority_unavailable"
    ) {
      this.suspendForAuthorityRecovery();
    }
    return result;
  }
  blocksRecord(id: string) {
    return [...this.entries.values()].some(
      (entry) =>
        this.blocks(entry) &&
        [
          entry.attempt.review.target.recordId,
          entry.attempt.review.replacement.recordId,
        ].includes(id),
    );
  }
  private blocks(entry: DecisionSupersessionOperation) {
    return (
      entry.transportPending ||
      ["preparing", "submitting", "uncertain"].includes(entry.phase)
    );
  }
  get pendingCount() {
    return [...this.entries.values()].filter(
      (entry) => this.blocks(entry) || entry.reconciliation === "refreshing",
    ).length;
  }
  get blockedCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        entry.phase === "uncertain" || entry.reconciliation === "required",
    ).length;
  }
  admit(
    review: DecisionSupersessionReview,
    binding: DecisionSupersessionBinding,
  ) {
    if (
      !this.port ||
      !this.canSubmit() ||
      !decisionReviewValid(review) ||
      review.authorityGeneration !== this.generation ||
      JSON.stringify(review.authority) !== JSON.stringify(this.authority) ||
      !binding.isCurrent() ||
      !binding.matchesReview() ||
      [review.target, review.replacement].some(
        (record) =>
          this.blocksRecord(record.recordId) ||
          (this.latestVersion(record.recordId) ?? 0) > record.baseRowVersion,
      ) ||
      !this.coordination.canReserve(review)
    )
      return null;
    let attempt: DecisionSupersessionAttempt;
    try {
      attempt = freeze(
        this.port.capture(review, this.ids.create("decision-supersede")),
      );
    } catch {
      return null;
    }
    if (this.entries.has(attempt.id)) return null;
    this.entries.set(
      attempt.id,
      Object.freeze({
        attempt,
        phase: "preparing",
        transportPending: false,
        receipt: null,
        failure: null,
        reconciliation: "pending",
      }),
    );
    this.bindings.set(attempt.id, binding);
    this.publish();
    return attempt;
  }
  private owns(attempt: DecisionSupersessionAttempt) {
    return (
      this.entries.get(attempt.id)?.attempt === attempt &&
      this.actorId === attempt.review.authority.actorId
    );
  }
  async execute(
    attempt: DecisionSupersessionAttempt,
    replay = false,
  ): Promise<void> {
    const entry = this.entries.get(attempt.id);
    const port = this.port;
    if (
      !entry ||
      !port ||
      !this.owns(attempt) ||
      !this.canSubmit() ||
      this.executing.has(attempt.id) ||
      entry.transportPending ||
      (replay ? entry.phase !== "uncertain" : entry.phase !== "preparing")
    )
      return;
    const generation = this.generation;
    const binding = this.bindings.get(attempt.id);
    this.executing.add(attempt.id);
    this.update(attempt.id, { phase: "preparing", failure: null });
    const preparation = this.observe(async (signal) => {
      if (replay) return true;
      return (
        (await this.coordination.coordinate(attempt.review, signal)) &&
        !signal.aborted &&
        binding?.isCurrent() &&
        binding.matchesReview() &&
        [attempt.review.target, attempt.review.replacement].every(
          (record) =>
            (this.latestVersion(record.recordId) ?? record.baseRowVersion) ===
            record.baseRowVersion,
        )
      );
    });
    this.observations.set(attempt.id, preparation);
    const ready = await preparation.result;
    if (!this.owns(attempt)) return;
    if (
      ready.kind !== "completed" ||
      !ready.value ||
      !this.canSubmit() ||
      this.generation !== generation
    ) {
      this.observations.delete(attempt.id);
      this.executing.delete(attempt.id);
      this.update(attempt.id, {
        phase: replay ? "uncertain" : "rejected",
        failure: replay
          ? {
              kind: "stale_target",
              message:
                "Exact recovery could not be sent. The earlier outcome remains unknown.",
            }
          : reviewRequired,
      });
      return;
    }
    this.update(attempt.id, { phase: "submitting", transportPending: true });
    const observation = this.observe(async (signal) => {
      const outcome = await port.send(attempt, signal);
      if (!this.owns(attempt)) return;
      if (outcome.kind === "acknowledged") {
        const receipt = freeze(structuredClone(outcome.receipt));
        // Durable-in-runtime acknowledgement always precedes projection or UI effects.
        this.versions.set(
          receipt.target_record_id,
          Math.max(
            receipt.target_row_version,
            this.latestVersion(receipt.target_record_id) ?? 0,
          ),
        );
        this.versions.set(
          receipt.superseding_record_id,
          Math.max(
            receipt.superseding_row_version,
            this.latestVersion(receipt.superseding_record_id) ?? 0,
          ),
        );
        this.update(attempt.id, {
          phase: "acknowledged",
          receipt,
          failure: null,
          reconciliation: "pending",
        });
      } else if (outcome.kind === "rejected") {
        this.update(attempt.id, {
          phase: replay ? "uncertain" : "rejected",
          failure: outcome.failure,
        });
        if (outcome.failure.publicCode === "incident_closed")
          this.closeIncident();
        if (
          workbookFailureLifecycle(outcome.failure).kind ===
          "authority_unavailable"
        ) {
          this.suspendForAuthorityRecovery();
        }
      } else this.update(attempt.id, { phase: "uncertain" });
    });
    this.observations.set(attempt.id, observation);
    const settlement = observation.settled.then(async () => {
      if (!this.owns(attempt)) return;
      this.observations.delete(attempt.id);
      this.executing.delete(attempt.id);
      this.update(attempt.id, { transportPending: false });
      if (this.entries.get(attempt.id)?.phase === "acknowledged")
        await this.refresh(attempt.id);
    });
    const result = await observation.result;
    if (
      this.owns(attempt) &&
      result.kind !== "completed" &&
      this.entries.get(attempt.id)?.phase === "submitting"
    )
      this.update(attempt.id, { phase: "uncertain" });
    if (result.kind === "completed") await settlement;
  }
  async replay(id: string) {
    const entry = this.entries.get(id);
    if (entry?.phase === "uncertain") await this.execute(entry.attempt, true);
  }
  async refresh(id: string): Promise<void> {
    const entry = this.entries.get(id);
    if (
      !entry?.receipt ||
      entry.transportPending ||
      entry.reconciliation === "refreshing" ||
      !this.authority
    )
      return;
    const receipt = entry.receipt,
      generation = this.generation;
    const sequence = (this.refreshes.get(id) ?? 0) + 1;
    this.refreshes.set(id, sequence);
    const binding = this.bindings.get(id);
    this.update(id, { reconciliation: "refreshing" });
    const observation = this.observe(async (signal) => {
      const isCurrent = () =>
        !signal.aborted &&
        this.owns(entry.attempt) &&
        this.generation === generation &&
        this.refreshes.get(id) === sequence;
      const reconcile = this.reconcile;
      if (!reconcile || !isCurrent())
        throw new Error("Decision reconciliation unavailable");
      await reconcile(receipt, { signal, isCurrent });
      if (!isCurrent() || reconcile !== this.reconcile)
        throw new Error("Decision reconciliation scope changed");
      if (binding?.isCurrent()) await binding.reconcile(receipt);
      if (!isCurrent()) throw new Error("Decision reconciliation retired");
    });
    const result = await observation.result;
    if (!this.owns(entry.attempt) || this.refreshes.get(id) !== sequence)
      return;
    this.update(id, {
      reconciliation:
        result.kind === "completed" && this.generation === generation
          ? "complete"
          : "required",
    });
  }
  dismiss(id: string) {
    const entry = this.entries.get(id);
    if (
      !entry ||
      this.blocks(entry) ||
      entry.reconciliation === "refreshing" ||
      (entry.receipt && entry.reconciliation !== "complete")
    )
      return;
    this.entries.delete(id);
    this.bindings.delete(id);
    this.refreshes.delete(id);
    this.publish();
  }
  private update(id: string, change: Partial<DecisionSupersessionOperation>) {
    const entry = this.entries.get(id);
    if (entry) {
      this.entries.set(id, Object.freeze({ ...entry, ...change }));
      this.publish();
    }
  }
  private publish() {
    this.snapshot = Object.freeze({
      authority: this.authority,
      generation: this.generation,
      revision: ++this.revision,
      entries: Object.freeze(this.authority ? [...this.entries.values()] : []),
    });
    for (const listener of this.listeners) listener();
  }
}
function freeze<T>(value: T): T {
  if (value && typeof value === "object") {
    for (const child of Object.values(value)) freeze(child);
    Object.freeze(value);
  }
  return value;
}
