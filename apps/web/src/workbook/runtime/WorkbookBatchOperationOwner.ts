import { observeAsyncOperation } from "../../services/asyncObservation";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { PendingReplayUnitState } from "./pending/workbookPendingQueue";
import type {
  WorkbookBatchAdmission,
  WorkbookBatchAttempt,
  WorkbookBatchEntry,
  WorkbookBatchPlan,
  WorkbookBatchReceipt,
  WorkbookBatchSnapshot,
  WorkbookBatchTransport,
  WorkbookBatchTransportOutcome,
} from "./workbookBatchOperation";

type PendingBatchOverlap = Pick<
  PendingReplayUnitState,
  "id" | "recordId" | "viewSchemaId"
>;

type Coordination = {
  readonly authorityChanged?: () => void;
  readonly pending: () => readonly PendingBatchOverlap[];
  readonly sealPending: () => void;
  readonly available: (plan: WorkbookBatchPlan) => boolean;
  readonly reserve: (plan: WorkbookBatchPlan) => (() => void) | null;
  readonly conflicts: (id: string) => boolean;
  readonly captured?: (attempt: WorkbookBatchAttempt) => void;
  readonly accepted: (
    receipt: WorkbookBatchReceipt,
    attempt: WorkbookBatchAttempt,
  ) => void;
  readonly refresh: (receipt: WorkbookBatchReceipt) => Promise<void>;
};
type Retained = {
  entry: WorkbookBatchEntry;
  readonly preceding: Set<string>;
  ready: boolean;
  hadUncertainty: boolean;
  release: (() => void) | null;
};

/** Owns explicit batches, not autosave replay or source-specific mapping. */
export class WorkbookBatchOperationOwner {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private transport: WorkbookBatchTransport | null = null;
  private readonly entries = new Map<string, Retained>();
  private deliveries = new WeakMap<object, string>();
  private readonly observations = new Map<string, { cancel: () => void }>();
  private readonly listeners = new Set<() => void>();
  private lifetime = 0;
  private authorityGeneration = 0;
  private wakeQueued = false;
  private admissionError: string | null = null;
  private snapshot: WorkbookBatchSnapshot = {
    authority: null,
    entries: [],
    admissionError: null,
  };

  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly coordination: Coordination,
    private readonly observe = observeAsyncOperation,
  ) {}
  getSnapshot = (): WorkbookBatchSnapshot => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  configure(transport: WorkbookBatchTransport) {
    this.transport = transport;
    this.wake();
  }
  canWrite(): boolean {
    return (
      this.authority !== null &&
      !this.authority.closed &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && authority.actorId !== this.actorId))
    )
      this.retire();
    this.authorityGeneration++;
    this.coordination.authorityChanged?.();
    this.authority =
      authority?.incidentId === this.incidentId ? immutable(authority) : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.publish();
    this.wake();
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  retire() {
    this.lifetime++;
    this.authorityGeneration++;
    this.coordination.authorityChanged?.();
    for (const observation of this.observations.values()) observation.cancel();
    this.observations.clear();
    for (const retained of this.entries.values()) retained.release?.();
    this.entries.clear();
    this.deliveries = new WeakMap();
    this.actorId = null;
    this.authority = null;
    this.admissionError = null;
    this.publish();
  }
  surfaceRefreshed(viewSchemaId: string): void {
    if (!this.authority) return;
    for (const retained of this.entries.values())
      if (
        retained.entry.phase === "acknowledged" &&
        retained.entry.receipt?.viewSchemaId === viewSchemaId &&
        retained.entry.reconciliation === "required"
      )
        this.update(retained, { reconciliation: "complete" });
    this.pruneCompleted();
    this.publish();
  }
  /** Save-state facts are separate from operation admission and refresh. */
  get unsettledMutationCount() {
    return [...this.entries.values()].filter(
      ({ entry }) =>
        !entry.receipt &&
        ["waiting", "preparing", "submitting", "uncertain"].includes(
          entry.phase,
        ),
    ).length;
  }
  get pendingCount() {
    return [...this.entries.values()].filter(
      ({ entry }) =>
        ["waiting", "preparing", "submitting"].includes(entry.phase) ||
        entry.reconciliation === "refreshing",
    ).length;
  }
  get blockedCount() {
    return [...this.entries.values()].filter(
      ({ entry }) =>
        ["uncertain", "rejected"].includes(entry.phase) ||
        entry.reconciliation === "required",
    ).length;
  }

  admit(
    plan: WorkbookBatchPlan,
    admission: WorkbookBatchAdmission,
  ): string | null {
    const duplicate = this.deliveries.get(admission.delivery);
    if (duplicate !== undefined) return duplicate;
    if (!this.canWrite() || !this.transport)
      return this.rejectAdmission(
        "This batch is unavailable until editing access is restored.",
      );
    let id: string;
    try {
      id = this.ids.create("workbook-batch");
    } catch {
      return this.rejectAdmission(
        "A secure request could not be created. Nothing was sent.",
      );
    }
    const retained: Retained = {
      entry: immutable({
        id,
        plan,
        phase: "waiting",
        attempt: null,
        receipt: null,
        failure: null,
        transportPending: false,
        reconciliation: "pending",
      }),
      preceding: new Set(this.coordination.pending().map((unit) => unit.id)),
      ready: admission.ready === undefined,
      hadUncertainty: false,
      release: null,
    };
    this.coordination.sealPending();
    this.deliveries.set(admission.delivery, id);
    this.entries.set(id, retained);
    this.admissionError = null;
    if (admission.ready !== undefined) {
      void admission.ready.then(
        () => {
          if (this.entries.get(id) !== retained) return;
          retained.ready = true;
          this.wake();
        },
        () => {
          if (this.entries.get(id) !== retained) return;
          this.update(retained, {
            phase: "rejected",
            failure: {
              kind: "stale_target",
              message:
                "Earlier edits could not finish. Review the original targets before trying again.",
            },
          });
        },
      );
    }
    this.publish();
    this.wake();
    return id;
  }
  private rejectAdmission(message: string): null {
    this.admissionError = message;
    this.publish();
    return null;
  }
  private blocks(retained: Retained): boolean {
    return (
      retained.entry.phase !== "acknowledged" ||
      this.coordination.conflicts(retained.entry.id)
    );
  }
  blocksEntityType(entityType: "host" | "identity"): boolean {
    return [...this.entries.values()].some(
      (entry) =>
        this.blocks(entry) && entry.entry.plan.entityType === entityType,
    );
  }
  blocksRecord(recordId: string, recoveryBatchId?: string): boolean {
    for (const retained of this.entries.values()) {
      if (retained.entry.id === recoveryBatchId) return false;
      if (
        this.blocks(retained) &&
        retained.entry.plan.recordIds.includes(recordId)
      )
        return true;
    }
    return false;
  }
  acceptBatchRow(batchId: string, recordId: string, rowVersion: number): void {
    let later = false;
    for (const retained of this.entries.values()) {
      if (retained.entry.id === batchId) {
        later = true;
        continue;
      }
      if (!later || retained.entry.attempt) continue;
      this.acceptPreparedVersion(retained, recordId, rowVersion);
    }
  }
  private acceptPreparedVersion(
    retained: Retained,
    recordId: string,
    rowVersion: number,
  ) {
    const plan = retained.entry.plan;
    const targets = plan.request.targets.map((target) =>
      "record_id" in target && target.record_id === recordId
        ? {
            ...target,
            base_row_version: Math.max(target.base_row_version, rowVersion),
          }
        : target,
    );
    this.update(retained, {
      plan: immutable({
        ...plan,
        request: { ...plan.request, targets },
      } as WorkbookBatchPlan),
    });
  }
  allowsPending(unit: PendingBatchOverlap): boolean {
    return ![...this.entries.values()].some(
      (retained) =>
        this.blocks(retained) &&
        !retained.preceding.has(unit.id) &&
        affectsUnit(retained.entry.plan, unit),
    );
  }
  acceptPrerequisiteRow(
    unitId: string,
    recordId: string,
    rowVersion: number,
  ): void {
    for (const retained of this.entries.values()) {
      if (retained.entry.attempt || !retained.preceding.has(unitId)) continue;
      this.acceptPreparedVersion(retained, recordId, rowVersion);
    }
  }
  wake(): void {
    if (this.wakeQueued) return;
    this.wakeQueued = true;
    queueMicrotask(() => {
      this.wakeQueued = false;
      this.pump();
    });
  }
  private pump() {
    if (this.pruneCompleted()) this.publish();
    if (!this.canWrite() || !this.transport) return;
    const earlier: Retained[] = [];
    for (const retained of this.entries.values()) {
      const { entry } = retained;
      if (
        entry.phase === "waiting" &&
        retained.ready &&
        !earlier.some(
          (other) =>
            this.blocks(other) && overlaps(other.entry.plan, entry.plan),
        ) &&
        !this.coordination
          .pending()
          .some(
            (unit) =>
              retained.preceding.has(unit.id) && affectsUnit(entry.plan, unit),
          ) &&
        this.coordination.available(entry.plan)
      ) {
        const release = this.coordination.reserve(entry.plan);
        if (release !== null) {
          retained.release = release;
          this.dispatch(retained);
        }
      }
      earlier.push(retained);
    }
  }
  retry(id: string): void {
    const retained = this.entries.get(id);
    if (!retained || !this.authority || retained.entry.transportPending) return;
    if (retained.entry.phase === "acknowledged") {
      void this.reconcile(retained);
      return;
    }
    if (!this.canWrite()) return;
    if (retained.entry.phase !== "uncertain" || !retained.entry.attempt) return;
    this.dispatch(retained);
  }
  discard(id: string): void {
    const retained = this.entries.get(id);
    // Uncertainty is never silently discarded: it might include committed creates.
    if (
      !retained ||
      retained.entry.transportPending ||
      !["waiting", "rejected"].includes(retained.entry.phase) ||
      retained.hadUncertainty
    )
      return;
    retained.release?.();
    this.entries.delete(id);
    this.publish();
    this.wake();
  }
  private dispatch(retained: Retained) {
    const transport = this.transport;
    if (!transport || !this.authority || !this.canWrite()) return;
    let attempt = retained.entry.attempt;
    if (!attempt) {
      try {
        attempt = immutable(
          transport.capture(
            retained.entry.plan,
            this.authority,
            retained.entry.id,
          ),
        );
        this.coordination.captured?.(attempt);
      } catch {
        retained.release?.();
        retained.release = null;
        this.update(retained, {
          phase: "rejected",
          failure: {
            kind: "validation",
            message:
              "The batch targets are invalid. Review the original range.",
          },
        });
        return;
      }
    }
    this.update(retained, {
      attempt,
      phase: "submitting",
      transportPending: true,
      failure: null,
    });
    const dispatched = attempt;
    const lifetime = this.lifetime;
    let settledOutcome = false;
    const observation = this.observe(async (signal) => {
      const outcome = await transport.send(dispatched, signal);
      if (
        lifetime === this.lifetime &&
        this.entries.get(dispatched.id) === retained
      ) {
        settledOutcome = true;
        this.settle(retained, outcome);
      }
      return outcome;
    });
    this.observations.set(dispatched.id, observation);
    void observation.result.then((result) => {
      if (
        lifetime !== this.lifetime ||
        this.entries.get(dispatched.id) !== retained ||
        settledOutcome
      )
        return;
      if (result.kind !== "completed") {
        retained.hadUncertainty = true;
        this.update(retained, { phase: "uncertain" });
      }
    });
    void observation.settled.then(() => {
      if (
        lifetime !== this.lifetime ||
        this.entries.get(dispatched.id) !== retained
      )
        return;
      this.observations.delete(dispatched.id);
      this.update(retained, { transportPending: false });
      this.pruneCompleted();
      this.publish();
    });
  }
  private settle(retained: Retained, outcome: WorkbookBatchTransportOutcome) {
    if (outcome.kind === "acknowledged") {
      // Retain before calling any source, listener, or mounted presentation effect.
      const receipt = immutable(outcome.receipt);
      this.update(retained, {
        receipt,
        phase: "acknowledged",
        failure: null,
      });
      for (const row of outcome.receipt.rows)
        this.acceptBatchRow(retained.entry.id, row.record_id, row.row_version);
      retained.release?.();
      retained.release = null;
      const attempt = retained.entry.attempt;
      try {
        if (attempt) this.coordination.accepted(receipt, attempt);
      } catch {
        this.update(retained, { reconciliation: "required" });
        return;
      }
      void this.reconcile(retained);
      this.wake();
    } else if (outcome.kind === "uncertain" || retained.hadUncertainty) {
      retained.hadUncertainty = true;
      this.update(retained, {
        phase: "uncertain",
        failure: outcome.kind === "rejected" ? outcome.failure : null,
      });
    } else {
      retained.release?.();
      retained.release = null;
      this.update(retained, { phase: "rejected", failure: outcome.failure });
    }
  }
  private async reconcile(retained: Retained) {
    const { receipt, id } = retained.entry;
    if (!receipt || retained.entry.reconciliation === "refreshing") return;
    if (!this.authority) {
      this.update(retained, { reconciliation: "required" });
      return;
    }
    const generation = this.authorityGeneration;
    this.update(retained, { reconciliation: "refreshing" });
    try {
      await this.coordination.refresh(receipt);
      if (this.entries.get(id) !== retained) return;
      this.update(retained, {
        reconciliation:
          generation === this.authorityGeneration ? "complete" : "required",
      });
    } catch {
      if (this.entries.get(id) === retained)
        this.update(retained, { reconciliation: "required" });
    }
    this.pruneCompleted();
    this.publish();
  }
  private pruneCompleted(): boolean {
    let changed = false;
    const latest = new Set<string>();
    for (const [id, retained] of [...this.entries].reverse()) {
      const { entry } = retained;
      if (
        entry.phase !== "acknowledged" ||
        entry.reconciliation !== "complete" ||
        entry.transportPending ||
        this.coordination.conflicts(id)
      )
        continue;
      const view = entry.plan.request.view_schema_id;
      if (latest.has(view)) {
        this.entries.delete(id);
        changed = true;
      } else latest.add(view);
    }
    return changed;
  }
  private update(retained: Retained, change: Partial<WorkbookBatchEntry>) {
    retained.entry = Object.freeze({ ...retained.entry, ...change });
    this.publish();
  }
  private publish() {
    this.snapshot = {
      authority: this.authority,
      entries: this.authority
        ? [...this.entries.values()].map(({ entry }) => entry)
        : [],
      admissionError: this.authority ? this.admissionError : null,
    };
    for (const listener of this.listeners) listener();
  }
}
function affectsUnit(plan: WorkbookBatchPlan, unit: PendingBatchOverlap) {
  return (
    plan.recordIds.includes(unit.recordId ?? "") ||
    (!!plan.entityType && plan.request.view_schema_id === unit.viewSchemaId)
  );
}
function overlaps(a: WorkbookBatchPlan, b: WorkbookBatchPlan) {
  return (
    a.recordIds.some((id) => b.recordIds.includes(id)) ||
    (!!a.entityType && a.entityType === b.entityType)
  );
}
function immutable<T>(value: T): T {
  const copy = structuredClone(value);
  const freeze = (item: unknown): void => {
    if (item && typeof item === "object") {
      for (const child of Object.values(item)) freeze(child);
      Object.freeze(item);
    }
  };
  freeze(copy);
  return copy;
}
