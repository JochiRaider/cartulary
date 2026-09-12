import { boundedRead } from "../../services/asyncObservation";
import type { SheetRef } from "../../shared/sheetRef";
import {
  type CapturedRecordPatch,
  captureRecordPatch,
  type RecordPatchOutcome,
  type RecordPatchTransport,
} from "../adapters/workbookRecordPatchTransport";
import {
  type RecordPatchChange,
  TaskLifecycleDraftStore,
  taskFieldEqual,
  taskGuardFields,
  taskPatchErrors,
  taskViewId,
} from "../features/coordination/taskLifecycleModel";
import type { PartyReview } from "../features/parties/partyLinkModel";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookPendingMutationAccepted } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookConflictRegistration } from "./WorkbookConflictStore";
import type { WorkbookConflictResolutionKind } from "./workbookConflictModel";

export type ExplicitPatchAuthority = WorkbookMutationAuthority;
export type ExplicitPatchIntent = Readonly<{
  baseline: WorkbookQueryRow;
  partyReview?: PartyReview;
  changes: readonly RecordPatchChange[];
  purpose: string;
  sheetRef: SheetRef;
  surfaceLabel: string;
}>;
export type ExplicitPatchOperation = Readonly<{
  id: string;
  intent: ExplicitPatchIntent;
  request: CapturedRecordPatch | null;
  phase:
    | "preparation_failed"
    | "coordinating"
    | "submitting"
    | "uncertain"
    | "rejected"
    | "conflict"
    | "acknowledged";
  receipt: WorkbookPendingMutationAccepted | null;
  failure: WorkbookOperationFailure | null;
  reconciliation: "pending" | "refreshing" | "required" | "complete";
}>;
type Snapshot = Readonly<{
  revision: number;
  authority: ExplicitPatchAuthority | null;
  entries: readonly ExplicitPatchOperation[];
}>;

/** Ordinary explicit PATCH lifetime. The autosave FIFO never owns these operations. */
export class WorkbookExplicitPatchOwner {
  readonly drafts = new TaskLifecycleDraftStore();
  readonly inspectorDrafts = new TaskLifecycleDraftStore();
  private authority: ExplicitPatchAuthority | null = null;
  private actorId: string | null = null;
  private retired = false;
  private generation = 0;
  private transport: RecordPatchTransport | null = null;
  private accessLost: (() => void) | undefined;
  private partyRecovery: {
    prepare(
      review: PartyReview,
      signal: AbortSignal,
      target?: string,
    ): Promise<boolean>;
    refresh(entry: ExplicitPatchOperation): Promise<void>;
    recheck(review: PartyReview): Promise<void>;
  } | null = null;
  configurePartyRecovery(recovery: NonNullable<typeof this.partyRecovery>) {
    this.partyRecovery = recovery;
  }
  private refreshSurface: (() => Promise<void>) | null = null;
  private readonly entries = new Map<string, ExplicitPatchOperation>();
  private readonly attempts = new Map<string, RecordPatchTransport>();
  private readonly running = new Set<string>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  private readonly versions = new Map<string, number>();
  private readonly listeners = new Set<() => void>();
  private snapshot: Snapshot = { revision: 0, authority: null, entries: [] };
  constructor(
    private readonly incidentId: string,
    private readonly transactionIds: SecureTransactionIdPort,
    private readonly boundaries: {
      coordinate(
        recordId: string,
        signal: AbortSignal,
        viewSchemaId?: string,
      ): Promise<boolean>;
      remember?(id: string): void;
      settle?(id: string): void;
      registerConflict(input: WorkbookConflictRegistration): void;
      accepted(row: WorkbookQueryRow): void;
    },
    private readonly timeoutMs = 30_000,
  ) {}
  configure(transport: RecordPatchTransport, accessLost?: () => void) {
    this.transport = transport;
    this.accessLost = accessLost;
  }
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  private emit() {
    this.snapshot = {
      revision: this.snapshot.revision + 1,
      authority: this.authority,
      entries: this.authority ? [...this.entries.values()] : [],
    };
    for (const listener of this.listeners) listener();
  }
  setAuthority(authority: ExplicitPatchAuthority | null) {
    if (this.retired) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId && this.actorId !== authority.actorId))
    ) {
      this.retire();
      return;
    }
    if (JSON.stringify(this.authority) === JSON.stringify(authority)) return;
    this.generation++;
    this.authority = authority;
    if (authority) this.actorId = authority.actorId;
    this.emit();
  }
  suspend() {
    this.setAuthority(null);
  }
  retire() {
    this.retired = true;
    this.authority = null;
    this.generation++;
    this.entries.clear();
    this.attempts.clear();
    this.rows.clear();
    this.versions.clear();
    this.drafts.clear();
    this.inspectorDrafts.clear();
    this.emit();
  }
  canSubmit() {
    return (
      !this.retired &&
      !!this.authority &&
      this.authority.role !== "viewer" &&
      this.authority.role !== "" &&
      !this.authority.closed &&
      this.transport !== null
    );
  }
  latestVersion(id: string) {
    return this.authority ? (this.versions.get(id) ?? null) : null;
  }
  latestRow(id: string) {
    return this.authority ? (this.rows.get(id) ?? null) : null;
  }
  acceptVersion(id: string, version: number) {
    if (!this.retired && version > (this.versions.get(id) ?? 0)) {
      this.versions.set(id, version);
      this.emit();
    }
  }
  acceptRow(row: WorkbookQueryRow): WorkbookQueryRow | null {
    if (
      this.retired ||
      row.row_version < (this.versions.get(row.record_id) ?? 0)
    )
      return this.rows.get(row.record_id) ?? null;
    const previous = this.rows.get(row.record_id);
    if (previous && previous.row_version >= row.row_version) return previous;
    this.rows.set(row.record_id, structuredClone(row));
    this.versions.set(row.record_id, row.row_version);
    this.boundaries.accepted(row);
    this.emit();
    return this.rows.get(row.record_id) ?? null;
  }
  observeQuery(row: WorkbookQueryRow) {
    return this.acceptRow(row);
  }
  blocksRecord(id: string) {
    return [...this.entries.values()].some(
      (entry) =>
        entry.intent.baseline.record_id === id &&
        (entry.phase === "coordinating" ||
          entry.phase === "submitting" ||
          entry.phase === "uncertain" ||
          entry.phase === "conflict" ||
          (entry.phase === "acknowledged" &&
            entry.reconciliation !== "complete")),
    );
  }
  get pendingCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        entry.phase === "coordinating" ||
        entry.phase === "submitting" ||
        entry.reconciliation === "refreshing",
    ).length;
  }
  get blockedCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        entry.phase === "uncertain" || entry.reconciliation === "required",
    ).length;
  }
  registerRefresh(refresh: () => Promise<void>) {
    this.refreshSurface = refresh;
    return () => {
      if (this.refreshSurface === refresh) this.refreshSurface = null;
    };
  }
  private update(id: string, update: Partial<ExplicitPatchOperation>) {
    const current = this.entries.get(id);
    if (current && !this.retired) {
      this.entries.set(id, { ...current, ...update });
      this.emit();
    }
  }
  private reject(id: string, message: string) {
    this.update(id, {
      phase: this.entries.get(id)?.intent.partyReview
        ? "preparation_failed"
        : "rejected",
      failure: { kind: "stale_target", message },
    });
  }
  async submit(
    intent: ExplicitPatchIntent,
  ): Promise<ExplicitPatchOperation | null> {
    intent = structuredClone(intent);
    const recordId = intent.baseline.record_id;
    if (!this.canSubmit() || this.blocksRecord(recordId) || !this.transport)
      return null;
    let id: string;
    try {
      id = this.transactionIds.create("workbook-explicit-patch");
      if (!id.trim()) return null;
    } catch {
      return null;
    }
    const transport = this.transport,
      generation = this.generation;
    this.entries.set(id, {
      id,
      intent: structuredClone(intent),
      request: intent.partyReview
        ? captureRecordPatch({
            recordId,
            viewSchemaId: intent.partyReview.pair.viewSchemaId,
            baseRowVersion: intent.baseline.row_version,
            changes: intent.changes,
            clientTxnId: id,
          })
        : null,
      phase: "coordinating",
      receipt: null,
      failure: null,
      reconciliation: "pending",
    });
    this.attempts.set(id, transport);
    this.emit(); // synchronous reservation before any await
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.timeoutMs);
    try {
      const coordinated = await Promise.race([
        this.boundaries.coordinate(
          recordId,
          controller.signal,
          intent.partyReview?.pair.viewSchemaId ?? taskViewId,
        ),
        new Promise<false>((resolve) =>
          controller.signal.addEventListener("abort", () => resolve(false), {
            once: true,
          }),
        ),
      ]);
      if (!coordinated) {
        this.reject(
          id,
          "Earlier writes need recovery before this Task can be changed. Your draft is retained.",
        );
        return this.entries.get(id) ?? null;
      }
      if (generation !== this.generation || !this.canSubmit()) {
        this.reject(
          id,
          "Access changed before dispatch. Re-query this Task before submitting your retained draft.",
        );
        return this.entries.get(id) ?? null;
      }
      if (intent.partyReview) {
        const review = intent.partyReview,
          recovery = this.partyRecovery;
        const targetValue = intent.changes.find(
          (change) => change.field_key === review.pair.refFieldKey,
        )?.value;
        const target =
          typeof targetValue === "string" ? targetValue : undefined;
        if (
          !this.entries.get(id)?.request ||
          !recovery ||
          !(await boundedRead(
            (signal) => recovery.prepare(review, signal, target),
            controller.signal,
            this.timeoutMs,
          )) ||
          generation !== this.generation ||
          !this.canSubmit()
        ) {
          this.reject(
            id,
            "The source or access changed. Review this Party action again.",
          );
          return this.entries.get(id) ?? null;
        }
      } else {
        const current = this.latestRow(recordId) ?? intent.baseline;
        const dependencies = new Set(
          intent.changes.map((change) => change.field_key),
        );
        if (
          intent.changes.some((change) =>
            taskGuardFields.some((field) => field === change.field_key),
          )
        )
          for (const field of taskGuardFields) dependencies.add(field);
        if (
          [...dependencies].some(
            (field) => !taskFieldEqual(current, intent.baseline, field),
          ) ||
          current.row_version < (this.latestVersion(recordId) ?? 0)
        ) {
          this.reject(
            id,
            "Saved fields changed while earlier writes finished. Review current values; your complete draft is retained.",
          );
          return this.entries.get(id) ?? null;
        }
        const errors = taskPatchErrors(current, intent.changes);
        if (errors.length) {
          this.update(id, {
            phase: "rejected",
            failure: {
              kind: "validation",
              message: errors[0]?.message ?? "Invalid Task state",
              fields: errors,
            },
          });
          return this.entries.get(id) ?? null;
        }
        const request = captureRecordPatch({
          recordId,
          viewSchemaId: taskViewId,
          baseRowVersion: current.row_version,
          changes: intent.changes,
          clientTxnId: id,
        });
        if (!request) {
          this.reject(
            id,
            "The patch does not match the writable field contract.",
          );
          return this.entries.get(id) ?? null;
        }
        this.update(id, { request });
      }
    } catch {
      this.reject(
        id,
        intent.partyReview
          ? "The source or Party could not be refreshed. Review current records before trying again."
          : "Earlier writes could not be coordinated. Your draft is retained.",
      );
      return this.entries.get(id) ?? null;
    } finally {
      clearTimeout(timeout);
    }
    await this.execute(id);
    return this.entries.get(id) ?? null;
  }
  private async execute(id: string) {
    const entry = this.entries.get(id),
      transport = this.attempts.get(id);
    if (
      !entry?.request ||
      !transport ||
      this.running.has(id) ||
      !this.canSubmit()
    )
      return;
    this.running.add(id);
    this.boundaries.remember?.(id);
    this.update(id, { phase: "submitting", failure: null });
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout> | undefined;
    let outcome: RecordPatchOutcome;
    let observationEnded = false;
    try {
      outcome = await Promise.race([
        transport.send(entry.request, controller.signal).then(async (value) => {
          if (observationEnded && value.kind === "acknowledged")
            await this.acceptReceipt(id, value.receipt);
          return value;
        }),
        new Promise<RecordPatchOutcome>((resolve) => {
          timer = setTimeout(() => {
            controller.abort();
            resolve({ kind: "uncertain" });
          }, this.timeoutMs);
        }),
      ]);
    } catch {
      outcome = { kind: "uncertain" };
    } finally {
      observationEnded = true;
      clearTimeout(timer);
      this.running.delete(id);
    }
    if (this.retired || !this.entries.has(id) || this.entries.get(id)?.receipt)
      return;
    if (outcome.kind === "uncertain") {
      this.update(id, { phase: "uncertain" });
      return;
    }
    if (outcome.kind === "rejected") {
      const failure = outcome.failure;
      if (
        failure.kind === "authentication_required" ||
        failure.kind === "authorization_lost"
      ) {
        this.update(id, {
          phase: entry.phase === "uncertain" ? "uncertain" : "rejected",
          failure,
        });
        this.suspend();
        if (entry.intent.partyReview)
          void this.partyRecovery
            ?.recheck(entry.intent.partyReview)
            .catch(() => {});
        else this.accessLost?.();
        return;
      }
      if (
        entry.phase === "uncertain" ||
        failure.kind === "client_txn_conflict"
      ) {
        this.update(id, { phase: "uncertain", failure });
        return;
      }
      this.boundaries.settle?.(id);
      if (failure.kind === "same_field_conflict") {
        this.boundaries.registerConflict({
          conflict: failure.conflict,
          viewSchemaId:
            entry.intent.partyReview?.pair.viewSchemaId ?? taskViewId,
          sheetRef: entry.intent.sheetRef,
          surfaceLabel: entry.intent.surfaceLabel,
          rowLabel: entry.intent.baseline.record_id,
          focusKey: `${entry.intent.baseline.record_id}:${failure.conflict.field_key}`,
          compoundOperationId:
            entry.intent.partyReview !== undefined ||
            entry.intent.purpose === "task-lifecycle" ||
            entry.intent.changes.length > 1
              ? id
              : undefined,
          focusOrigin: "inspector",
        });
        this.update(id, { phase: "conflict", failure });
      } else
        this.update(id, {
          phase: "rejected",
          failure,
        });
      return;
    }
    await this.acceptReceipt(id, outcome.receipt);
  }
  private async acceptReceipt(
    id: string,
    receipt: WorkbookPendingMutationAccepted,
  ) {
    const entry = this.entries.get(id);
    if (!entry || this.retired || entry.receipt) return;
    this.boundaries.settle?.(id);
    this.acceptRow(receipt.row);
    this.update(id, {
      phase: "acknowledged",
      receipt: structuredClone(receipt),
      failure: null,
      reconciliation: "required",
    });
    if (entry.intent.purpose === "task-lifecycle")
      this.drafts.clear(entry.intent.baseline.record_id);
    await this.refresh(id);
  }
  async replay(id: string) {
    const entry = this.entries.get(id);
    if (
      !entry ||
      entry.phase !== "uncertain" ||
      !this.canSubmit() ||
      this.running.has(id)
    )
      return;
    this.running.add(id);
    try {
      const generation = this.generation;
      if (entry.intent.partyReview) {
        if (!this.partyRecovery) throw new Error("Party recovery unavailable");
        await this.partyRecovery.recheck(entry.intent.partyReview);
      } else {
        if (!this.refreshSurface) throw new Error("Task view is unavailable");
        await this.refreshSurface();
      }
      if (generation !== this.generation || !this.canSubmit()) return;
    } catch {
      this.update(id, {
        failure: {
          kind: "stale_target",
          message:
            "Refresh current Tasks before replaying. The original request is retained.",
        },
      });
      return;
    } finally {
      this.running.delete(id);
    }
    await this.execute(id);
  }
  async refresh(id: string) {
    const entry = this.entries.get(id);
    if (
      !entry?.receipt ||
      !this.authority ||
      entry.reconciliation === "refreshing"
    )
      return;
    const generation = this.generation;
    this.update(id, { reconciliation: "refreshing" });
    try {
      if (entry.intent.partyReview) {
        if (!this.partyRecovery) throw new Error("Party recovery unavailable");
        await this.partyRecovery.refresh(entry);
      } else {
        if (!this.refreshSurface) throw new Error("Task view is unavailable");
        await this.refreshSurface();
      }
      if (generation !== this.generation)
        throw new Error("Task access changed");
      this.update(id, { reconciliation: "complete" });
    } catch {
      this.update(id, { reconciliation: "required" });
    }
  }
  conflictResolved(
    recordId: string,
    resolutionKind: WorkbookConflictResolutionKind = "keep_saved",
    row?: WorkbookQueryRow,
  ) {
    for (const entry of this.entries.values()) {
      if (
        entry.phase !== "conflict" ||
        entry.intent.baseline.record_id !== recordId
      )
        continue;
      if (
        !entry.intent.partyReview &&
        entry.intent.purpose !== "task-lifecycle" &&
        entry.intent.changes.length === 1 &&
        row
      ) {
        this.inspectorDrafts.review(
          row,
          entry.intent.changes[0]?.field_key ?? "",
          false,
        );
        this.entries.delete(entry.id);
        this.attempts.delete(entry.id);
        this.emit();
      } else
        this.update(entry.id, {
          phase: "rejected",
          failure: {
            kind: "stale_target",
            message:
              resolutionKind === "keep_saved"
                ? "The saved field was kept. Review changed saved fields and submit your retained draft together."
                : "The field conflict was resolved.",
          },
        });
    }
  }
  dismiss(id: string) {
    const entry = this.entries.get(id);
    if (
      entry &&
      (entry.phase === "rejected" ||
        (entry.phase === "acknowledged" && entry.reconciliation === "complete"))
    ) {
      this.entries.delete(id);
      this.attempts.delete(id);
      this.emit();
    }
  }
}
