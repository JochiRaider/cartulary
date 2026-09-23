import type { ViewContract } from "@cartulary/view-contracts";
import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { IndicatorCreateDraftStore } from "./IndicatorCreateDraftStore";
import {
  type IndicatorCreateValues,
  indicatorCreateErrors,
} from "./indicatorCreateModel";
import type {
  IndicatorCreateAdmission,
  IndicatorCreateAttempt,
  IndicatorCreateBinding,
  IndicatorCreateOperation,
  IndicatorCreateOwnerPort,
  IndicatorCreateReceipt,
  IndicatorCreateReconcile,
  IndicatorCreateSnapshot,
  IndicatorCreateTransport,
} from "./indicatorCreateOperation";
import {
  freezeObservation,
  observationUUID,
  observationVersion,
} from "./observationModel";
import type {
  ObservationAttempt,
  ObservationAuthority,
  ObservationReadPort,
} from "./observationOperation";

/** Retains the canonical operation independently of both panel and observation mutation. */
export class WorkbookIndicatorCreateOwner implements IndicatorCreateOwnerPort {
  private authority: ObservationAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private revision = 0;
  private snapshot: IndicatorCreateSnapshot = {
    authority: null,
    generation: 0,
    revision: 0,
    entries: [],
  };
  private readonly listeners = new Set<() => void>();
  private readonly entries = new Map<string, IndicatorCreateOperation>();
  private readonly bindings = new Map<string, IndicatorCreateBinding>();
  private readonly pending = new Map<string, { cancel: () => void }>();
  private readonly executing = new Set<string>();
  private readonly refreshes = new Map<string, number>();
  private readonly versions = new Map<string, number>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  private reader: ObservationReadPort | null = null;
  private transport: IndicatorCreateTransport | null = null;
  private reconcile: IndicatorCreateReconcile | null = null;
  private authorityUncertain: (() => void) | undefined;
  readonly drafts = new IndicatorCreateDraftStore(() => this.publish());
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly accepted: (
      receipt: IndicatorCreateReceipt,
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
    transport: IndicatorCreateTransport,
    authorityUncertain?: () => void,
  ) {
    this.reader = reader;
    this.transport = transport;
    this.authorityUncertain = authorityUncertain;
  }
  registerReconciliation(reconcile: IndicatorCreateReconcile) {
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
    for (const [id, pending] of this.pending)
      if (this.entries.get(id)?.phase === "preparing") pending.cancel();
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
    for (const pending of this.pending.values()) pending.cancel();
    this.pending.clear();
    this.executing.clear();
    this.entries.clear();
    this.bindings.clear();
    this.refreshes.clear();
    this.versions.clear();
    this.rows.clear();
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
    const prior = this.rows.get(row.record_id);
    if (
      row.row_version < (this.versions.get(row.record_id) ?? 0) ||
      (prior && prior.row_version >= row.row_version)
    )
      return prior ?? null;
    const accepted = freezeObservation(structuredClone(row));
    this.rows.set(row.record_id, accepted);
    this.versions.set(row.record_id, row.row_version);
    this.publish();
    return accepted;
  }
  private blocks(entry: IndicatorCreateOperation) {
    return (
      entry.transportPending ||
      ["preparing", "submitting", "uncertain"].includes(entry.phase)
    );
  }
  busy(observationId: string) {
    return [...this.entries.values()].some(
      (entry) =>
        entry.attempt.observation.observation_id === observationId &&
        this.blocks(entry),
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
  admit(
    observation: IndicatorObservation,
    contract: ViewContract,
    values: IndicatorCreateValues,
    binding: IndicatorCreateBinding,
  ): IndicatorCreateAdmission {
    const unavailable = {
      kind: "unavailable" as const,
      message:
        "Canonical creation is unavailable or the observation changed. Review before trying again.",
    };
    if (
      !this.authority ||
      !this.transport ||
      !this.canSubmit() ||
      !binding.isCurrent() ||
      !binding.matchesDraft() ||
      this.busy(observation.observation_id) ||
      observation.incident_id !== this.incidentId ||
      !observationUUID(observation.observation_id) ||
      !observationUUID(observation.source_record_id) ||
      !observationVersion(observation.row_version) ||
      observation.resolution_status === "dismissed" ||
      [...this.entries.values()].some(
        (entry) =>
          entry.attempt.observation.observation_id ===
            observation.observation_id && entry.receipt,
      )
    )
      return unavailable;
    const errors = indicatorCreateErrors(contract, values);
    if (Object.keys(errors).length) return { kind: "invalid", errors };
    let attempt: IndicatorCreateAttempt;
    try {
      attempt = freezeObservation(
        this.transport.capture(
          this.authority,
          this.generation,
          observation,
          contract,
          values,
          this.ids.create("indicator-create"),
        ),
      );
    } catch {
      return unavailable;
    }
    if (!attempt.id || this.entries.has(attempt.id)) return unavailable;
    this.entries.set(
      attempt.id,
      freezeObservation({
        attempt,
        phase: "preparing",
        transportPending: false,
        receipt: null,
        failure: null,
        refresh: "pending",
        resolutionAttemptIds: [],
      }),
    );
    this.bindings.set(attempt.id, binding);
    this.publish();
    return { kind: "admitted", attempt };
  }
  private owns(attempt: IndicatorCreateAttempt) {
    return (
      this.entries.get(attempt.id)?.attempt === attempt &&
      this.actorId === attempt.authority.actorId
    );
  }
  private async prepare(attempt: IndicatorCreateAttempt, signal: AbortSignal) {
    const reader = this.reader;
    if (!reader) return false;
    let cursor: string | null = null;
    const seen = new Set<string>();
    for (;;) {
      const result = await boundedRead(
        (signal) =>
          reader.observations(
            { kind: "source", recordId: attempt.observation.source_record_id },
            cursor,
            signal,
          ),
        signal,
      );
      if (
        signal.aborted ||
        !this.owns(attempt) ||
        attempt.generation !== this.generation
      )
        return false;
      if (result.kind !== "accepted") {
        if (
          result.kind === "rejected" &&
          workbookFailureLifecycle(result.failure).kind ===
            "authority_unavailable"
        )
          this.suspendForAuthorityRecovery();
        return false;
      }
      const item = result.value.items.find(
        (item) => item.observation_id === attempt.observation.observation_id,
      );
      if (item)
        return Object.entries(attempt.observation).every(
          ([key, value]) => item[key as keyof typeof item] === value,
        );
      cursor = result.value.nextCursor;
      if (!result.value.hasMore || !cursor || seen.has(cursor)) return false;
      seen.add(cursor);
    }
  }
  async execute(
    attempt: IndicatorCreateAttempt,
    replay = false,
  ): Promise<void> {
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
        ((await this.prepare(attempt, signal)) &&
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
            ? "Recovery was not sent. The original create outcome remains unknown."
            : "The observation or access changed. Review the observation; no canonical create was sent.",
        },
      });
      return;
    }
    this.update(attempt.id, {
      phase: "submitting",
      transportPending: true,
      failure: null,
    });
    const operation = this.observe(async (signal) => {
      const outcome = await transport.send(attempt, signal);
      if (!this.owns(attempt)) return;
      if (outcome.kind === "accepted") {
        const receipt = freezeObservation(structuredClone(outcome.receipt));
        this.versions.set(
          receipt.row.record_id,
          Math.max(
            receipt.row.row_version,
            this.versions.get(receipt.row.record_id) ?? 0,
          ),
        );
        this.update(attempt.id, {
          phase: "accepted",
          receipt,
          failure: null,
          refresh: "pending",
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
    this.pending.set(attempt.id, operation);
    const settlement = operation.settled.then(async () => {
      if (!this.owns(attempt)) return;
      this.pending.delete(attempt.id);
      this.executing.delete(attempt.id);
      this.update(attempt.id, { transportPending: false });
      if (this.entries.get(attempt.id)?.receipt) await this.refresh(attempt.id);
    });
    const result = await operation.result;
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
      entry.refresh === "refreshing"
    )
      return;
    if (!this.authority) {
      this.update(id, { refresh: "required" });
      return;
    }
    const generation = this.generation,
      sequence = (this.refreshes.get(id) ?? 0) + 1;
    this.refreshes.set(id, sequence);
    this.update(id, { refresh: "refreshing" });
    const operation = this.observe(async (signal) => {
      const isCurrent = () =>
        !signal.aborted &&
        this.owns(entry.attempt) &&
        generation === this.generation &&
        this.refreshes.get(id) === sequence;
      const reconcile = this.reconcile;
      if (!reconcile || !isCurrent())
        throw new Error("Canonical result refresh unavailable");
      await reconcile(entry.attempt, entry.receipt as IndicatorCreateReceipt, {
        signal,
        isCurrent,
      });
      if (!isCurrent() || reconcile !== this.reconcile)
        throw new Error("Canonical result refresh detached");
    });
    const result = await operation.result;
    if (this.owns(entry.attempt) && this.refreshes.get(id) === sequence)
      this.update(id, {
        refresh:
          result.kind === "completed" && generation === this.generation
            ? "complete"
            : "required",
      });
  }
  associateResolution(id: string, resolution: ObservationAttempt) {
    const entry = this.entries.get(id);
    if (
      !entry?.receipt ||
      !this.authority ||
      resolution.authority.actorId !== entry.attempt.authority.actorId ||
      resolution.authority.incidentId !== this.incidentId ||
      resolution.intent.action !== "resolve" ||
      resolution.intent.observation.observation_id !==
        entry.attempt.observation.observation_id ||
      resolution.intent.targetId !== entry.receipt.row.record_id ||
      entry.resolutionAttemptIds.includes(resolution.id)
    )
      return;
    this.update(id, {
      resolutionAttemptIds: [...entry.resolutionAttemptIds, resolution.id],
    });
  }
  private update(id: string, patch: Partial<IndicatorCreateOperation>) {
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
