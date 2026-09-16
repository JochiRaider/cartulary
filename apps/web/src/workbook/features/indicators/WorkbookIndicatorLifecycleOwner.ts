import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import type { IndicatorLifecycleReceipt } from "../../adapters/indicatorLifecycleProtocol";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { IndicatorLifecycleDraftStore } from "./IndicatorLifecycleDraftStore";
import {
  freezeLifecycle,
  indicatorLifecycleViewId,
  type LifecycleDraft,
  validateLifecycleDraft,
} from "./indicatorLifecycleModel";
import type {
  IndicatorLifecycleOwnerPort,
  IndicatorLifecycleTransportPort,
  LifecycleAttempt,
  LifecycleAuthority,
  LifecycleBinding,
  LifecycleOperation,
  LifecycleScope,
  LifecycleSnapshot,
} from "./indicatorLifecycleOperation";

type Reconcile = (
  receipt: IndicatorLifecycleReceipt,
  scope: LifecycleScope,
) => Promise<void>;
export class WorkbookIndicatorLifecycleOwner
  implements IndicatorLifecycleOwnerPort
{
  private authority: LifecycleAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private revision = 0;
  private snapshot: LifecycleSnapshot = {
    authority: null,
    generation: 0,
    revision: 0,
    entries: [],
  };
  private readonly listeners = new Set<() => void>();
  private readonly entries = new Map<string, LifecycleOperation>();
  private readonly bindings = new Map<string, LifecycleBinding>();
  private readonly versions = new Map<string, number>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  private readonly observations = new Map<string, { cancel: () => void }>();
  private readonly executing = new Set<string>();
  private readonly refreshes = new Map<string, number>();
  private readonly reviews = new Set<string>();
  private port: IndicatorLifecycleTransportPort | null = null;
  private reconcile: Reconcile | null = null;
  private authorityUncertain: (() => void) | undefined;
  readonly drafts = new IndicatorLifecycleDraftStore(() => this.publish());
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly coordination: {
      canReserve: (draft: LifecycleDraft) => boolean;
      coordinate: (
        draft: LifecycleDraft,
        signal: AbortSignal,
      ) => Promise<boolean>;
      accepted: (
        receipt: IndicatorLifecycleReceipt,
        clientTxnId: string,
      ) => void;
    },
    private readonly observe = observeAsyncOperation,
  ) {}
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  configure(
    port: IndicatorLifecycleTransportPort,
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
  setAuthority(authority: LifecycleAuthority | null) {
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
      authority?.incidentId === this.incidentId
        ? freezeLifecycle({ ...authority })
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.publish();
  }
  canReplay() {
    return (
      !!this.authority?.sessionIdentity &&
      !!this.authority.actorId &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  canSubmit() {
    return this.canReplay() && this.authority?.closed === false;
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  suspendForAuthorityRecovery() {
    this.suspend();
    this.authorityUncertain?.();
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
    this.reviews.clear();
    this.authority = null;
    this.actorId = null;
    this.reconcile = null;
    this.drafts.clear();
    this.publish();
  }
  acceptVersion(recordId: string, version: number) {
    if (
      Number.isSafeInteger(version) &&
      version > (this.versions.get(recordId) ?? 0)
    ) {
      this.versions.set(recordId, version);
      this.publish();
    }
  }
  latestVersion(recordId: string) {
    return this.versions.get(recordId) ?? null;
  }
  latestRow(recordId: string) {
    return this.authority ? (this.rows.get(recordId) ?? null) : null;
  }
  acceptRow(row: WorkbookQueryRow) {
    if (
      !this.authority ||
      !Number.isSafeInteger(row.row_version) ||
      row.row_version < 1
    )
      return null;
    const prior = this.rows.get(row.record_id);
    if (
      row.row_version < (this.latestVersion(row.record_id) ?? 0) ||
      (prior && prior.row_version >= row.row_version)
    )
      return prior ?? null;
    const accepted = freezeLifecycle(structuredClone(row));
    this.rows.set(row.record_id, accepted);
    this.versions.set(row.record_id, row.row_version);
    this.publish();
    return accepted;
  }
  async intervals(
    recordId: string,
    cursor: string | null,
    signal: AbortSignal,
  ) {
    const generation = this.generation;
    if (!this.authority || !this.port) return { kind: "aborted" as const };
    const result = await this.port.intervals(recordId, cursor, signal);
    if (signal.aborted || generation !== this.generation)
      return { kind: "aborted" as const };
    if (
      result.kind === "rejected" &&
      workbookFailureLifecycle(result.failure).kind === "authority_unavailable"
    )
      this.suspendForAuthorityRecovery();
    return result;
  }
  async records(
    viewId: string,
    query: WorkbookQueryState,
    cursor: string | null,
    signal: AbortSignal,
  ) {
    const generation = this.generation;
    if (!this.authority || !this.port) return { kind: "aborted" as const };
    const result = await this.port.records(viewId, query, cursor, signal);
    if (signal.aborted || generation !== this.generation)
      return { kind: "aborted" as const };
    if (
      result.kind === "rejected" &&
      workbookFailureLifecycle(result.failure).kind === "authority_unavailable"
    )
      this.suspendForAuthorityRecovery();
    return result;
  }
  private blocks(entry: LifecycleOperation) {
    return (
      entry.transportPending ||
      ["preparing", "submitting", "uncertain"].includes(entry.phase)
    );
  }
  blocksRecord(id: string) {
    return [...this.entries.values()].some(
      (entry) => entry.attempt.draft.recordId === id && this.blocks(entry),
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
  admit(draft: LifecycleDraft, binding: LifecycleBinding) {
    if (
      !this.port ||
      !this.authority ||
      !this.canSubmit() ||
      !binding.isCurrent() ||
      binding.matchesDraft?.() === false ||
      this.blocksRecord(draft.recordId) ||
      this.reviews.has(draft.recordId) ||
      this.drafts.get(draft.recordId) !== draft ||
      !Number.isSafeInteger(draft.baseRowVersion) ||
      draft.baseRowVersion < 1 ||
      this.latestVersion(draft.recordId) !== draft.baseRowVersion ||
      !this.coordination.canReserve(draft)
    )
      return null;
    const validated = validateLifecycleDraft(draft.values);
    if (!validated.values) return null;
    let attempt: LifecycleAttempt;
    try {
      attempt = freezeLifecycle(
        this.port.capture(
          this.authority,
          this.generation,
          draft,
          validated.values,
          this.ids.create("indicator-lifecycle"),
        ),
      );
    } catch {
      return null;
    }
    if (!attempt.id || this.entries.has(attempt.id)) return null;
    this.entries.set(
      attempt.id,
      freezeLifecycle({
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
  private owns(attempt: LifecycleAttempt) {
    return (
      this.entries.get(attempt.id)?.attempt === attempt &&
      this.actorId === attempt.authority.actorId
    );
  }
  async execute(attempt: LifecycleAttempt, replay = false): Promise<void> {
    const entry = this.entries.get(attempt.id),
      port = this.port;
    if (
      !entry ||
      !port ||
      !this.owns(attempt) ||
      !(replay ? this.canReplay() : this.canSubmit()) ||
      this.executing.has(attempt.id) ||
      entry.transportPending ||
      (replay ? entry.phase !== "uncertain" : entry.phase !== "preparing")
    )
      return;
    this.executing.add(attempt.id);
    const generation = this.generation,
      binding = this.bindings.get(attempt.id);
    const preparation = this.observe(
      async (signal) =>
        replay ||
        ((await this.coordination.coordinate(attempt.draft, signal)) &&
          !signal.aborted &&
          binding?.isCurrent() === true &&
          binding.matchesDraft?.() !== false &&
          this.drafts.get(attempt.draft.recordId)?.revision ===
            attempt.draft.revision &&
          this.latestVersion(attempt.draft.recordId) ===
            attempt.draft.baseRowVersion),
    );
    this.observations.set(attempt.id, preparation);
    const ready = await preparation.result;
    if (!this.owns(attempt)) return;
    if (
      ready.kind !== "completed" ||
      !ready.value ||
      this.generation !== generation ||
      !(replay ? this.canReplay() : this.canSubmit()) ||
      (!replay && attempt.generation !== this.generation)
    ) {
      this.executing.delete(attempt.id);
      this.observations.delete(attempt.id);
      this.update(attempt.id, {
        phase: replay ? "uncertain" : "rejected",
        failure: {
          kind: "stale_target",
          message: replay
            ? "Recovery could not be sent. The original outcome remains unknown."
            : "The Indicator or access changed. Refresh and review the retained draft. No interval was sent.",
        },
      });
      return;
    }
    this.update(attempt.id, {
      phase: "submitting",
      transportPending: true,
      failure: null,
    });
    const observation = this.observe(async (signal) => {
      const outcome = await port.send(attempt, signal);
      if (!this.owns(attempt)) return;
      if (outcome.kind === "acknowledged") {
        const receipt = freezeLifecycle(structuredClone(outcome.receipt));
        for (const record of receipt.affected_records)
          this.versions.set(
            record.record_id,
            Math.max(
              record.row_version,
              this.latestVersion(record.record_id) ?? 0,
            ),
          );
        this.update(attempt.id, {
          phase: "acknowledged",
          receipt,
          failure: null,
          reconciliation: "pending",
        });
        this.coordination.accepted(receipt, attempt.id);
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
        )
          this.suspendForAuthorityRecovery();
      } else this.update(attempt.id, { phase: "uncertain" });
    });
    this.observations.set(attempt.id, observation);
    const settlement = observation.settled.then(async () => {
      if (!this.owns(attempt)) return;
      this.observations.delete(attempt.id);
      this.executing.delete(attempt.id);
      this.update(attempt.id, { transportPending: false });
      if (this.entries.get(attempt.id)?.receipt) await this.refresh(attempt.id);
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
      entry.reconciliation === "refreshing"
    )
      return;
    if (!this.authority) {
      this.update(id, { reconciliation: "required" });
      return;
    }
    const receipt = entry.receipt;
    const generation = this.generation,
      sequence = (this.refreshes.get(id) ?? 0) + 1;
    this.refreshes.set(id, sequence);
    this.update(id, { reconciliation: "refreshing" });
    const observation = this.observe(async (signal) => {
      const isCurrent = () =>
        !signal.aborted &&
        this.owns(entry.attempt) &&
        this.generation === generation &&
        this.refreshes.get(id) === sequence;
      const reconcile = this.reconcile;
      if (!reconcile || !isCurrent())
        throw new Error("Lifecycle reconciliation unavailable");
      await reconcile(receipt, { signal, isCurrent });
      if (!isCurrent() || reconcile !== this.reconcile)
        throw new Error("Lifecycle reconciliation detached");
      const binding = this.bindings.get(id);
      if (binding?.isCurrent()) await binding.reconcile();
      if (!isCurrent()) throw new Error("Lifecycle reconciliation retired");
    });
    const result = await observation.result;
    if (this.owns(entry.attempt) && this.refreshes.get(id) === sequence)
      this.update(id, {
        reconciliation:
          result.kind === "completed" && generation === this.generation
            ? "complete"
            : "required",
      });
  }
  async review(recordId: string): Promise<boolean> {
    const draft = this.drafts.get(recordId),
      generation = this.generation;
    if (
      !draft ||
      !this.canSubmit() ||
      this.blocksRecord(recordId) ||
      this.reviews.has(recordId)
    )
      return false;
    this.reviews.add(recordId);
    const controller = new AbortController();
    try {
      let cursor: string | null = null;
      const seen = new Set<string>();
      do {
        const page = await boundedRead(
          (signal) =>
            this.records(
              indicatorLifecycleViewId,
              emptyWorkbookQueryState(),
              cursor,
              signal,
            ),
          controller.signal,
        );
        if (
          page.kind !== "accepted" ||
          generation !== this.generation ||
          !this.canSubmit() ||
          this.drafts.get(recordId) !== draft
        )
          return false;
        const row = page.value.items.find((row) => row.record_id === recordId);
        if (row) {
          if (row.row_version < (this.latestVersion(recordId) ?? 0))
            return false;
          this.acceptRow(row);
          this.drafts.review(recordId, row.row_version);
          return true;
        }
        cursor = page.value.nextCursor;
        if (!page.value.hasMore || !cursor || seen.has(cursor)) return false;
        seen.add(cursor);
      } while (generation === this.generation);
      return false;
    } catch {
      return false;
    } finally {
      controller.abort();
      this.reviews.delete(recordId);
    }
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
  private update(id: string, patch: Partial<LifecycleOperation>) {
    const entry = this.entries.get(id);
    if (entry) {
      this.entries.set(id, freezeLifecycle({ ...entry, ...patch }));
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
