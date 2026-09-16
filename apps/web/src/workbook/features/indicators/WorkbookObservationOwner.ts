import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { ObservationDraftStore } from "./ObservationDraftStore";
import {
  freezeObservation,
  observationCanTransition,
  observationIndicatorView,
  observationUUID,
  observationVersion,
} from "./observationModel";
import {
  type ObservationAttempt,
  type ObservationAuthority,
  type ObservationBinding,
  type ObservationIntent,
  type ObservationOperation,
  type ObservationOwnerPort,
  type ObservationReadPort,
  type ObservationReceipt,
  type ObservationScope,
  type ObservationSnapshot,
  type ObservationSubject,
  type ObservationTransportPort,
  observationIntentKey,
} from "./observationOperation";

type Reconcile = (
  attempt: ObservationAttempt,
  receipt: ObservationReceipt,
  scope: ObservationScope,
) => Promise<void>;
/** Owns admitted observation writes for the workbook runtime lifetime. */
export class WorkbookObservationOwner implements ObservationOwnerPort {
  private authority: ObservationAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private revision = 0;
  private snapshot: ObservationSnapshot = {
    authority: null,
    generation: 0,
    revision: 0,
    entries: [],
  };
  private readonly listeners = new Set<() => void>();
  private readonly entries = new Map<string, ObservationOperation>();
  private readonly bindings = new Map<string, ObservationBinding>();
  private readonly versions = new Map<string, number>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  private readonly pending = new Map<string, { cancel: () => void }>();
  private readonly executing = new Set<string>();
  private readonly refreshes = new Map<string, number>();
  private reader: ObservationReadPort | null = null;
  private transport: ObservationTransportPort | null = null;
  private reconcile: Reconcile | null = null;
  private authorityUncertain: (() => void) | undefined;
  readonly drafts = new ObservationDraftStore(() => this.publish());
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly accepted: (
      receipt: ObservationReceipt,
      id: string,
    ) => void,
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
    reader: ObservationReadPort,
    transport: ObservationTransportPort,
    authorityUncertain?: () => void,
  ) {
    this.reader = reader;
    this.transport = transport;
    this.authorityUncertain = authorityUncertain;
  }
  registerReconciliation(reconcile: Reconcile) {
    this.reconcile = reconcile;
    return () => {
      if (this.reconcile === reconcile) this.reconcile = null;
    };
  }
  setAuthority(authority: ObservationAuthority | null) {
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && authority.actorId !== this.actorId))
    )
      this.retire();
    this.generation++;
    this.bindings.clear();
    for (const [id, observation] of this.pending)
      if (this.entries.get(id)?.phase === "preparing") observation.cancel();
    this.authority =
      authority?.incidentId === this.incidentId
        ? freezeObservation({ ...authority })
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
    for (const observation of this.pending.values()) observation.cancel();
    this.pending.clear();
    this.executing.clear();
    this.entries.clear();
    this.bindings.clear();
    this.versions.clear();
    this.rows.clear();
    this.refreshes.clear();
    this.authority = null;
    this.actorId = null;
    this.reconcile = null;
    this.drafts.clear();
    this.publish();
  }
  acceptVersion(id: string, version: number) {
    if (observationVersion(version) && version > (this.versions.get(id) ?? 0)) {
      this.versions.set(id, version);
      this.publish();
    }
  }
  latestVersion(id: string) {
    return this.authority ? (this.versions.get(id) ?? null) : null;
  }
  latestRow(id: string) {
    return this.authority ? (this.rows.get(id) ?? null) : null;
  }
  acceptRow(row: WorkbookQueryRow) {
    if (!this.authority || !observationVersion(row.row_version)) return null;
    const previous = this.rows.get(row.record_id);
    if (
      row.row_version < (this.versions.get(row.record_id) ?? 0) ||
      (previous && previous.row_version >= row.row_version)
    )
      return previous ?? null;
    const accepted = freezeObservation(structuredClone(row));
    this.rows.set(row.record_id, accepted);
    this.versions.set(row.record_id, row.row_version);
    this.publish();
    return accepted;
  }
  async observations(
    subject: ObservationSubject,
    cursor: string | null,
    signal: AbortSignal,
  ) {
    const generation = this.generation;
    if (!this.authority || !this.reader) return { kind: "aborted" as const };
    const result = await this.reader.observations(subject, cursor, signal);
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
    view: string,
    query: WorkbookQueryState,
    cursor: string | null,
    signal: AbortSignal,
  ) {
    const generation = this.generation;
    if (!this.authority || !this.reader) return { kind: "aborted" as const };
    const result = await this.reader.records(view, query, cursor, signal);
    if (signal.aborted || generation !== this.generation)
      return { kind: "aborted" as const };
    if (
      result.kind === "rejected" &&
      workbookFailureLifecycle(result.failure).kind === "authority_unavailable"
    )
      this.suspendForAuthorityRecovery();
    return result;
  }
  private blocks(entry: ObservationOperation) {
    return (
      entry.transportPending ||
      ["preparing", "submitting", "uncertain"].includes(entry.phase)
    );
  }
  busy(intent: ObservationIntent) {
    return [...this.entries.values()].some(
      (entry) =>
        observationIntentKey(entry.attempt.intent) ===
          observationIntentKey(intent) && this.blocks(entry),
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
  admit(intent: ObservationIntent, binding: ObservationBinding) {
    if (
      !this.transport ||
      !this.authority ||
      !this.canSubmit() ||
      !binding.isCurrent() ||
      !binding.matchesDraft() ||
      this.busy(intent)
    )
      return null;
    if (intent.action === "create") {
      const bytes = new TextEncoder().encode(intent.source.text),
        { startByte, endByte, text } = intent.selection;
      if (
        intent.source.incidentId !== this.incidentId ||
        !observationUUID(intent.source.recordId) ||
        !observationVersion(intent.source.rowVersion) ||
        !Number.isInteger(startByte) ||
        !Number.isInteger(endByte) ||
        startByte < 0 ||
        endByte <= startByte ||
        endByte > bytes.length ||
        !text ||
        text.includes("\0") ||
        (intent.targetId !== undefined && !observationUUID(intent.targetId))
      )
        return null;
      try {
        if (
          new TextDecoder("utf-8", { fatal: true }).decode(
            bytes.slice(startByte, endByte),
          ) !== text
        )
          return null;
      } catch {
        return null;
      }
    } else if (
      intent.observation.incident_id !== this.incidentId ||
      !observationUUID(intent.observation.observation_id) ||
      !observationVersion(intent.observation.row_version) ||
      !observationCanTransition(
        intent.observation,
        intent.action,
        intent.action === "resolve" ? intent.targetId : undefined,
      )
    )
      return null;
    let attempt: ObservationAttempt;
    try {
      attempt = freezeObservation(
        this.transport.capture(
          this.authority,
          this.generation,
          intent,
          this.ids.create("indicator-observation"),
        ),
      );
    } catch {
      return null;
    }
    if (!attempt.id || this.entries.has(attempt.id)) return null;
    this.entries.set(
      attempt.id,
      freezeObservation({
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
  private owns(attempt: ObservationAttempt) {
    return (
      this.entries.get(attempt.id)?.attempt === attempt &&
      this.actorId === attempt.authority.actorId
    );
  }
  private async prepare(intent: ObservationIntent, signal: AbortSignal) {
    // Fresh target/child reads never run on replay: the original may already be committed.
    if (intent.action !== "create") {
      let cursor: string | null = null;
      const seen = new Set<string>();
      for (;;) {
        const result = await boundedRead(
          (signal) =>
            this.observations(
              { kind: "source", recordId: intent.observation.source_record_id },
              cursor,
              signal,
            ),
          signal,
        );
        if (result.kind !== "accepted") return false;
        const item = result.value.items.find(
          (item) => item.observation_id === intent.observation.observation_id,
        );
        if (item) {
          if (
            Object.entries(intent.observation).some(
              ([key, value]) => item[key as keyof typeof item] !== value,
            )
          )
            return false;
          break;
        }
        cursor = result.value.nextCursor;
        if (!result.value.hasMore || !cursor || seen.has(cursor)) return false;
        seen.add(cursor);
      }
    }
    const target =
      intent.action === "create" || intent.action === "resolve"
        ? intent.targetId
        : undefined;
    if (!target) return true;
    let cursor: string | null = null;
    const seen = new Set<string>();
    for (;;) {
      const result = await boundedRead(
        (signal) =>
          this.records(
            observationIndicatorView,
            emptyWorkbookQueryState(),
            cursor,
            signal,
          ),
        signal,
      );
      if (result.kind !== "accepted") return false;
      if (result.value.items.some((row) => row.record_id === target))
        return true;
      cursor = result.value.nextCursor;
      if (!result.value.hasMore || !cursor || seen.has(cursor)) return false;
      seen.add(cursor);
    }
  }
  async execute(attempt: ObservationAttempt, replay = false): Promise<void> {
    const entry = this.entries.get(attempt.id),
      transport = this.transport;
    if (
      !entry ||
      !transport ||
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
        ((await binding?.prepare(signal)) === true &&
          !signal.aborted &&
          (await this.prepare(attempt.intent, signal)) &&
          !signal.aborted &&
          binding?.isCurrent() === true &&
          binding.matchesDraft()),
    );
    this.pending.set(attempt.id, preparation);
    const ready = await preparation.result;
    if (!this.owns(attempt)) return;
    if (
      ready.kind !== "completed" ||
      !ready.value ||
      generation !== this.generation ||
      !(replay ? this.canReplay() : this.canSubmit()) ||
      (!replay && attempt.generation !== generation)
    ) {
      this.executing.delete(attempt.id);
      this.pending.delete(attempt.id);
      this.update(attempt.id, {
        phase: replay ? "uncertain" : "rejected",
        failure: {
          kind: "stale_target",
          message: replay
            ? "Recovery was not sent. The original outcome remains unknown."
            : "The source, observation, target or access changed. Review and reselect before trying again. No observation change was sent.",
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
      const outcome = await transport.send(attempt, signal);
      if (!this.owns(attempt)) return;
      if (outcome.kind === "acknowledged") {
        const receipt = freezeObservation(structuredClone(outcome.receipt));
        for (const record of receipt.affected_records)
          this.versions.set(
            record.record_id,
            Math.max(
              record.row_version,
              this.versions.get(record.record_id) ?? 0,
            ),
          );
        this.update(attempt.id, {
          phase: "acknowledged",
          receipt,
          failure: null,
          reconciliation: "pending",
        });
        this.accepted(receipt, attempt.id);
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
    this.pending.set(attempt.id, observation);
    const settlement = observation.settled.then(async () => {
      if (!this.owns(attempt)) return;
      this.pending.delete(attempt.id);
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
    const receipt = entry.receipt,
      generation = this.generation,
      sequence = (this.refreshes.get(id) ?? 0) + 1;
    this.refreshes.set(id, sequence);
    this.update(id, { reconciliation: "refreshing" });
    const observation = this.observe(async (signal) => {
      const isCurrent = () =>
        !signal.aborted &&
        this.owns(entry.attempt) &&
        generation === this.generation &&
        this.refreshes.get(id) === sequence;
      const reconcile = this.reconcile;
      if (!reconcile || !isCurrent())
        throw new Error("Observation reconciliation unavailable");
      await reconcile(entry.attempt, receipt, { signal, isCurrent });
      if (!isCurrent() || reconcile !== this.reconcile)
        throw new Error("Observation reconciliation detached");
      const binding = this.bindings.get(id);
      if (binding?.isCurrent()) await binding.reconcile();
      if (!isCurrent()) throw new Error("Observation reconciliation retired");
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
  private update(id: string, patch: Partial<ObservationOperation>) {
    const entry = this.entries.get(id);
    if (entry) {
      this.entries.set(id, freezeObservation({ ...entry, ...patch }));
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
