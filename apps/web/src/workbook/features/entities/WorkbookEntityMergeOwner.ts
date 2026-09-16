import { observeAsyncOperation } from "../../../services/asyncObservation";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type {
  EntityMergeAttempt,
  EntityMergeBinding,
  EntityMergeCoordination,
  EntityMergeOperation,
  EntityMergeReceipt,
  WorkbookEntityMergePort,
} from "./entityMergeOperation";
import type {
  EntityMergeAuthority,
  EntityMergeReview,
} from "./entityMergeReview";

export type EntityMergeOwnerSnapshot = {
  readonly authority: EntityMergeAuthority | null;
  readonly generation: number;
  readonly entries: readonly EntityMergeOperation[];
};
const reviewRequired: WorkbookOperationFailure = {
  kind: "stale_target",
  message:
    "The reviewed records changed or earlier writes could not finish. Refresh both records and review the merge again. No merge was sent.",
};

// Explicit merges have workbook lifetime, separate from history and autosave.
export class WorkbookEntityMergeOwner {
  private authority: EntityMergeAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private entries = new Map<string, EntityMergeOperation>();
  private versions = new Map<string, number>();
  private bindings = new Map<string, EntityMergeBinding>();
  private observations = new Map<string, { cancel: () => void }>();
  private executing = new Set<string>();
  private port: WorkbookEntityMergePort | null = null;
  private authorityUncertain: (() => void) | undefined;
  private timelineRefresh: (() => Promise<void>) | null = null;
  private projectionRefresh:
    | ((receipt: EntityMergeReceipt) => Promise<void>)
    | null = null;
  private snapshot: EntityMergeOwnerSnapshot = {
    authority: null,
    generation: 0,
    entries: [],
  };
  private readonly listeners = new Set<() => void>();

  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly coordination: EntityMergeCoordination,
    private readonly observe: typeof observeAsyncOperation = observeAsyncOperation,
  ) {}

  getSnapshot = (): EntityMergeOwnerSnapshot => this.snapshot;
  subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
  configure(
    port: WorkbookEntityMergePort,
    authorityUncertain?: () => void,
  ): void {
    this.port = port;
    this.authorityUncertain = authorityUncertain;
  }
  registerTimelineRefresh(refresh: () => Promise<void>) {
    this.timelineRefresh = refresh;
    return () => {
      if (this.timelineRefresh === refresh) this.timelineRefresh = null;
    };
  }
  async refreshTimeline(): Promise<void> {
    if (!this.authority?.role || !this.timelineRefresh)
      throw new Error("Timeline projection unavailable");
    await this.timelineRefresh();
  }
  registerProjectionRefresh(
    refresh: (receipt: EntityMergeReceipt) => Promise<void>,
  ) {
    this.projectionRefresh = refresh;
    return () => {
      if (this.projectionRefresh === refresh) this.projectionRefresh = null;
    };
  }
  setAuthority(authority: EntityMergeAuthority | null): void {
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    if (
      authority !== null &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && authority.actorId !== this.actorId) ||
        !authority.role)
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
    if (this.authority !== null) this.actorId = this.authority.actorId;
    this.publish();
  }
  canSubmit(): boolean {
    const authority = this.authority;
    return (
      authority !== null &&
      authority.actorId !== "" &&
      authority.sessionIdentity !== "" &&
      !authority.closed &&
      (authority.role === "reviewer" || authority.role === "admin")
    );
  }
  suspend(): void {
    this.setAuthority(null);
  }
  retire(): void {
    this.generation++;
    for (const observation of this.observations.values()) observation.cancel();
    this.observations.clear();
    this.executing.clear();
    this.entries.clear();
    this.bindings.clear();
    this.versions.clear();
    this.projectionRefresh = null;
    this.timelineRefresh = null;
    this.actorId = null;
    this.authority = null;
    this.publish();
  }
  closeIncident(): void {
    if (this.authority !== null)
      this.setAuthority({ ...this.authority, closed: true });
  }
  acceptVersion(recordId: string, version: number): void {
    if (
      Number.isSafeInteger(version) &&
      version > (this.versions.get(recordId) ?? 0)
    )
      this.versions.set(recordId, version);
  }
  latestVersion(recordId: string): number | null {
    return this.versions.get(recordId) ?? null;
  }
  blocksRecord(recordId: string): boolean {
    return [...this.entries.values()].some(
      (entry) =>
        this.blocks(entry) &&
        [
          entry.attempt.review.survivor.recordId,
          entry.attempt.review.loser.recordId,
        ].includes(recordId),
    );
  }
  blocksEntityType(entityType: "host" | "identity"): boolean {
    return [...this.entries.values()].some(
      (entry) =>
        this.blocks(entry) && entry.attempt.review.entityType === entityType,
    );
  }
  private blocks(entry: EntityMergeOperation): boolean {
    return (
      entry.transportPending ||
      ["preparing", "submitting", "uncertain"].includes(entry.phase)
    );
  }
  /** Admitted writes awaiting settlement; excludes acknowledged refresh reads. */
  get unsettledMutationCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        !entry.receipt &&
        (entry.phase === "preparing" ||
          entry.phase === "submitting" ||
          entry.phase === "uncertain"),
    ).length;
  }
  get pendingCount(): number {
    return [...this.entries.values()].filter(
      (entry) => this.blocks(entry) || entry.reconciliation === "refreshing",
    ).length;
  }
  get blockedCount(): number {
    return [...this.entries.values()].filter(
      (entry) =>
        entry.phase === "uncertain" || entry.reconciliation === "required",
    ).length;
  }
  admit(
    review: EntityMergeReview,
    binding: EntityMergeBinding,
  ): EntityMergeAttempt | null {
    if (
      !this.port ||
      !this.canSubmit() ||
      !review.plan.valid ||
      !review.reason.trim() ||
      review.authorityGeneration !== this.generation ||
      review.authority.actorId !== this.actorId ||
      !binding.isCurrent() ||
      !binding.matchesReview() ||
      review.survivor.recordId === review.loser.recordId ||
      [review.survivor, review.loser].some(
        (record) =>
          this.blocksRecord(record.recordId) ||
          (this.latestVersion(record.recordId) ?? 0) > record.baseRowVersion,
      ) ||
      !this.coordination.canReserve(review)
    )
      return null;
    let attempt: EntityMergeAttempt;
    try {
      attempt = freeze(
        this.port.capture(review, this.ids.create("entity-merge")),
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
  private owns(attempt: EntityMergeAttempt): boolean {
    return (
      this.entries.get(attempt.id)?.attempt === attempt &&
      this.actorId === attempt.review.authority.actorId
    );
  }
  private authorized(attempt: EntityMergeAttempt): boolean {
    return this.owns(attempt) && this.canSubmit();
  }
  async execute(attempt: EntityMergeAttempt, replay = false): Promise<void> {
    const entry = this.entries.get(attempt.id);
    const port = this.port;
    if (
      !entry ||
      !port ||
      this.executing.has(attempt.id) ||
      entry.transportPending ||
      !this.authorized(attempt) ||
      (replay ? entry.phase !== "uncertain" : entry.phase !== "preparing")
    )
      return;
    const generation = this.generation;
    const binding = this.bindings.get(attempt.id);
    this.executing.add(attempt.id);
    this.update(attempt.id, { phase: "preparing", failure: null });
    const preparation = this.observe(async (signal) => {
      const ready = await this.coordination.coordinate(attempt.review, signal);
      return (
        ready &&
        !signal.aborted &&
        (replay ||
          (binding?.isCurrent() &&
            binding.matchesReview() &&
            [attempt.review.survivor, attempt.review.loser].every(
              (record) =>
                (this.latestVersion(record.recordId) ??
                  record.baseRowVersion) === record.baseRowVersion,
            )))
      );
    });
    this.observations.set(attempt.id, preparation);
    const prepared = await preparation.result;
    if (!this.owns(attempt)) return;
    if (
      prepared.kind !== "completed" ||
      !prepared.value ||
      !this.authorized(attempt) ||
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
                "Recovery could not be sent. The earlier merge outcome is still unknown; retain this exact attempt.",
            }
          : reviewRequired,
      });
      return;
    }
    this.update(attempt.id, { phase: "submitting", transportPending: true });
    const dispatchGeneration = this.generation;
    const observation = this.observe(async (signal) => {
      const outcome = await port.send(attempt, signal);
      if (!this.owns(attempt)) return;
      if (outcome.kind === "acknowledged") {
        const receipt = freeze(outcome.receipt);
        // Receipt ownership precedes any disposable callback or projection refresh.
        this.update(attempt.id, {
          phase: "acknowledged",
          receipt,
          failure: null,
          reconciliation: "pending",
        });
        const newer =
          (this.latestVersion(receipt.survivor_record_id) ?? 0) >
            receipt.survivor_row_version ||
          (this.latestVersion(receipt.loser_record_id) ?? 0) >
            receipt.loser_row_version;
        this.acceptVersion(
          receipt.survivor_record_id,
          receipt.survivor_row_version,
        );
        this.acceptVersion(receipt.loser_record_id, receipt.loser_row_version);
        if (
          !newer &&
          this.generation === generation &&
          this.authorized(attempt) &&
          binding?.isCurrent()
        ) {
          try {
            binding.acknowledged(receipt);
          } catch {
            /* Receipt remains authoritative. */
          }
        }
      } else if (outcome.kind === "rejected") {
        // A later rejection, including transaction conflict, cannot disprove an earlier commit.
        this.update(attempt.id, {
          phase: replay ? "uncertain" : "rejected",
          failure: outcome.failure,
        });
        if (outcome.failure.publicCode === "incident_closed")
          this.closeIncident();
        if (
          dispatchGeneration === this.generation &&
          workbookFailureLifecycle(outcome.failure).kind ===
            "authority_unavailable"
        ) {
          this.suspend();
          this.authorityUncertain?.();
        }
      } else this.update(attempt.id, { phase: "uncertain" });
    });
    this.observations.set(attempt.id, observation);
    const finishTransport = observation.settled.then(async () => {
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
    if (result.kind === "completed") await finishTransport;
  }
  async replay(id: string): Promise<void> {
    const entry = this.entries.get(id);
    if (entry?.phase === "uncertain") await this.execute(entry.attempt, true);
  }
  async refresh(
    id: string,
    options?: { readonly requireAcceptance?: boolean },
  ): Promise<void> {
    const entry = this.entries.get(id);
    if (
      !entry?.receipt ||
      entry.phase !== "acknowledged" ||
      entry.transportPending ||
      entry.reconciliation === "refreshing" ||
      !this.authority?.role
    ) {
      if (options?.requireAcceptance)
        throw new Error("Merge reconciliation unavailable");
      return;
    }
    const generation = this.generation;
    const receipt = entry.receipt;
    const binding = this.bindings.get(id);
    this.update(id, { reconciliation: "refreshing" });
    const observation = this.observe(async () => {
      const refresh = this.projectionRefresh;
      if (!refresh) throw new Error("Entity projections unavailable");
      await refresh(receipt);
      if (this.projectionRefresh !== refresh)
        throw new Error("Projection scope changed");
      if (this.generation !== generation || !this.authority?.role)
        throw new Error("Access changed");
      if (binding?.isCurrent()) await binding.reconcile(receipt);
    });
    const result = await observation.result;
    if (!this.owns(entry.attempt)) {
      if (options?.requireAcceptance) throw new Error("Merge lifetime changed");
      return;
    }
    this.update(id, {
      reconciliation:
        result.kind === "completed" && this.generation === generation
          ? "complete"
          : "required",
    });
    if (
      options?.requireAcceptance &&
      (result.kind !== "completed" || this.generation !== generation)
    )
      throw new Error(
        "Merge reconciliation did not establish current projections",
      );
  }
  dismiss(id: string): void {
    const entry = this.entries.get(id);
    if (!entry || this.blocks(entry) || entry.reconciliation === "refreshing")
      return;
    // Acknowledged-but-unrefreshed actions retain their recovery path.
    if (entry.phase === "acknowledged" && entry.reconciliation !== "complete")
      return;
    this.entries.delete(id);
    this.bindings.delete(id);
    this.publish();
  }
  private update(id: string, update: Partial<EntityMergeOperation>): void {
    const entry = this.entries.get(id);
    if (entry) {
      this.entries.set(id, Object.freeze({ ...entry, ...update }));
      this.publish();
    }
  }
  private publish(): void {
    this.snapshot = Object.freeze({
      authority: this.authority,
      generation: this.generation,
      entries: Object.freeze(
        this.authority?.role ? [...this.entries.values()] : [],
      ),
    });
    for (const listener of this.listeners) listener();
  }
}
function freeze<T>(value: T): T {
  if (value !== null && typeof value === "object") {
    for (const child of Object.values(value)) freeze(child);
    Object.freeze(value);
  }
  return value;
}
