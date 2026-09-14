import { requireViewContract } from "@cartulary/view-contracts";
import { boundedRead } from "../../services/asyncObservation";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookProtocolPatchRecordRequest } from "../adapters/workbookProtocolTypes";
import {
  type CapturedRecordPatch,
  captureRecordPatch,
  type RecordPatchOutcome,
  type RecordPatchTransport,
} from "../adapters/workbookRecordPatchTransport";
import { workbookSavedFieldEqual } from "../models/workbookSavedValues";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookPendingMutationAccepted } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookConflictRegistration } from "./WorkbookConflictStore";
import type { WorkbookConflictResolutionKind } from "./workbookConflictModel";

export type ExplicitPatchAuthority = WorkbookMutationAuthority;
export type ExplicitPatchIntent = Readonly<{
  viewSchemaId: string;
  baseline: WorkbookQueryRow;
  changes: readonly WorkbookProtocolPatchRecordRequest["changes"][number][];
  purpose: string;
  sheetRef: SheetRef;
  surfaceLabel: string;
  owner?: string;
  review?: unknown;
  authoringRevision?: number;
  presentationIdentity?: string;
  compound?: boolean;
}>;
export type ExplicitPatchOperation = Readonly<{
  id: string;
  authority: ExplicitPatchAuthority;
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
/** Domain preparation and effects are supplied by their source owners. */
export type ExplicitPatchContribution = Readonly<{
  dependencies?: readonly string[];
  prepare?(signal: AbortSignal): Promise<void>;
  validate?(
    current: WorkbookQueryRow,
    intent: ExplicitPatchIntent,
  ): WorkbookOperationFailure | null;
  recheck?(): Promise<void>;
  accessRejected?(): Promise<void>;
  acknowledged?(entry: ExplicitPatchOperation): void;
  /** Complete source reconciliation, including surface refresh/debt; replaces the default refresh. */
  reconcile?(entry: ExplicitPatchOperation): Promise<void>;
  conflictResolved?(
    entry: ExplicitPatchOperation,
    kind: WorkbookConflictResolutionKind,
    row?: WorkbookQueryRow,
  ): void;
}>;
type Snapshot = Readonly<{
  revision: number;
  authority: ExplicitPatchAuthority | null;
  entries: readonly ExplicitPatchOperation[];
}>;

/** Retained explicit PATCH lifetime, separate from the grid autosave FIFO. */
export class WorkbookExplicitPatchOwner {
  private authority: ExplicitPatchAuthority | null = null;
  private actorId: string | null = null;
  private retired = false;
  private generation = 0;
  private transport: RecordPatchTransport | null = null;
  private accessLost: (() => void) | undefined;
  private readonly entries = new Map<string, ExplicitPatchOperation>();
  private readonly attempts = new Map<string, RecordPatchTransport>();
  private readonly contributions = new Map<
    string,
    readonly ExplicitPatchContribution[]
  >();
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
        viewSchemaId: string,
      ): Promise<boolean>;
      contribute?(
        intent: ExplicitPatchIntent,
      ): readonly ExplicitPatchContribution[];
      refresh?(viewSchemaId: string): Promise<void>;
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
      entries: this.canRead() ? [...this.entries.values()] : [],
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
    this.contributions.clear();
    this.rows.clear();
    this.versions.clear();
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
  private canRead() {
    return (
      !this.retired && this.authority !== null && this.authority.role !== ""
    );
  }
  latestVersion(id: string) {
    return this.canRead() ? (this.versions.get(id) ?? null) : null;
  }
  latestRow(id: string) {
    return this.canRead() ? (this.rows.get(id) ?? null) : null;
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
        (["coordinating", "submitting", "uncertain", "conflict"].includes(
          entry.phase,
        ) ||
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
  private update(id: string, update: Partial<ExplicitPatchOperation>) {
    const current = this.entries.get(id);
    if (current && !this.retired) {
      this.entries.set(id, { ...current, ...update });
      this.emit();
    }
  }
  private reject(id: string, message: string) {
    this.update(id, {
      phase: "preparation_failed",
      failure: { kind: "stale_target", message },
    });
  }
  async submit(
    intent: ExplicitPatchIntent,
    extra: readonly ExplicitPatchContribution[] = [],
  ): Promise<ExplicitPatchOperation | null> {
    intent = immutableClone(intent);
    const recordId = intent.baseline.record_id;
    if (
      !this.canSubmit() ||
      !this.authority ||
      this.blocksRecord(recordId) ||
      !this.transport
    )
      return null;
    let id: string;
    try {
      id = this.transactionIds.create("workbook-explicit-patch");
      if (!id.trim() || this.entries.has(id)) return null;
    } catch {
      return null;
    }
    const generation = this.generation;
    this.entries.set(id, {
      id,
      authority: immutableClone(this.authority),
      intent,
      request: null,
      phase: "coordinating",
      receipt: null,
      failure: null,
      reconciliation: "pending",
    });
    this.attempts.set(id, this.transport);
    const contributions = [
      ...(this.boundaries.contribute?.(intent) ?? []),
      ...extra,
    ];
    this.contributions.set(id, contributions);
    this.emit(); // Reserve synchronously, before coordination or any callback can admit another action.
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.timeoutMs);
    try {
      const coordinated = await boundedRead(
        (signal) =>
          this.boundaries.coordinate(recordId, signal, intent.viewSchemaId),
        controller.signal,
        this.timeoutMs,
      );
      if (!coordinated)
        throw new Error(
          "Earlier writes need recovery. Your draft is retained.",
        );
      for (const contribution of contributions)
        if (contribution.prepare)
          await boundedRead(
            contribution.prepare,
            controller.signal,
            this.timeoutMs,
          );
      if (generation !== this.generation || !this.canSubmit())
        throw new Error(
          "Access changed before dispatch. Review current authority and values.",
        );
      const current = this.latestRow(recordId) ?? intent.baseline;
      const dependencies = new Set([
        ...intent.changes.map((change) => change.field_key),
        ...contributions.flatMap((item) => item.dependencies ?? []),
      ]);
      const contract = requireViewContract(intent.viewSchemaId);
      if (
        intent.changes.some(
          (change) =>
            !contract.fieldMap[change.field_key]?.patchWritable ||
            !Object.hasOwn(current.cells, change.field_key),
        )
      )
        throw new Error("The record or field is unavailable for editing.");
      if (
        current.row_version < (this.latestVersion(recordId) ?? 0) ||
        [...dependencies].some(
          (field) => !workbookSavedFieldEqual(current, intent.baseline, field),
        )
      )
        throw new Error(
          "Saved fields changed while earlier writes finished. Review current values; your complete draft is retained.",
        );
      for (const contribution of contributions) {
        const failure = contribution.validate?.(current, intent);
        if (failure) {
          this.update(id, { phase: "rejected", failure });
          return this.entries.get(id) ?? null;
        }
      }
      const request = captureRecordPatch({
        recordId,
        viewSchemaId: intent.viewSchemaId,
        baseRowVersion: current.row_version,
        changes: intent.changes,
        clientTxnId: id,
      });
      if (!request)
        throw new Error(
          "The patch does not match the writable field contract.",
        );
      this.update(id, { request });
    } catch {
      this.reject(
        id,
        "The record, access, or saved values could not be confirmed. Review current values; your draft is retained.",
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
    let ended = false;
    try {
      outcome = await Promise.race([
        transport.send(entry.request, controller.signal).then(async (value) => {
          if (ended && value.kind === "acknowledged")
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
      ended = true;
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
        const handlers = (this.contributions.get(id) ?? []).filter(
          (item) => item.accessRejected,
        );
        if (handlers.length)
          for (const handler of handlers)
            void handler.accessRejected?.().catch(() => {});
        else this.accessLost?.();
        return;
      }
      if (entry.phase === "uncertain") {
        this.update(id, { phase: "uncertain", failure });
        return;
      }
      this.boundaries.settle?.(id);
      if (failure.kind === "same_field_conflict") {
        this.boundaries.registerConflict({
          conflict: failure.conflict,
          viewSchemaId: entry.intent.viewSchemaId,
          sheetRef: entry.intent.sheetRef,
          surfaceLabel: entry.intent.surfaceLabel,
          rowLabel: entry.intent.baseline.record_id,
          focusKey: `${entry.intent.baseline.record_id}:${failure.conflict.field_key}`,
          compoundOperationId:
            entry.intent.compound || entry.intent.changes.length > 1
              ? id
              : undefined,
          focusOrigin: "inspector",
        });
        this.update(id, { phase: "conflict", failure });
      } else this.update(id, { phase: "rejected", failure });
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
    // Correlation also fences late transport callbacks and custom source adapters.
    if (
      receipt.viewSchemaId !== entry.intent.viewSchemaId ||
      receipt.row.record_id !== entry.intent.baseline.record_id ||
      receipt.row.row_version <= (entry.request?.baseRowVersion ?? 0) ||
      !receipt.changeSetId
    ) {
      this.update(id, { phase: "uncertain" });
      return;
    }
    this.update(id, {
      phase: "acknowledged",
      receipt: immutableClone(receipt),
      failure: null,
      reconciliation: "required",
    });
    this.boundaries.settle?.(id);
    this.acceptRow(receipt.row);
    const accepted = this.entries.get(id);
    if (accepted)
      for (const contribution of this.contributions.get(id) ?? [])
        contribution.acknowledged?.(accepted);
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
      const generation = this.generation,
        contributions = this.contributions.get(id) ?? [];
      const rechecks = contributions.filter((item) => item.recheck);
      if (rechecks.length)
        for (const contribution of rechecks) await contribution.recheck?.();
      else {
        if (!this.boundaries.refresh) throw new Error("Source unavailable");
        await this.boundaries.refresh(entry.intent.viewSchemaId);
      }
      if (generation !== this.generation || !this.canSubmit()) return;
    } catch {
      this.update(id, {
        failure: {
          kind: "stale_target",
          message:
            "Refresh current authority before replaying. The original request is retained.",
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
      !this.canRead() ||
      entry.reconciliation === "refreshing"
    )
      return;
    const generation = this.generation;
    this.update(id, { reconciliation: "refreshing" });
    try {
      const reconcilers = (this.contributions.get(id) ?? []).filter(
        (item) => item.reconcile,
      );
      if (reconcilers.length) {
        for (const contribution of reconcilers)
          await contribution.reconcile?.(entry);
      } else {
        if (!this.boundaries.refresh) throw new Error("Source unavailable");
        await this.boundaries.refresh(entry.intent.viewSchemaId);
      }
      if (generation !== this.generation) throw new Error("Access changed");
      this.update(id, { reconciliation: "complete" });
    } catch {
      this.update(id, { reconciliation: "required" });
    }
  }
  conflictResolved(
    recordId: string,
    kind: WorkbookConflictResolutionKind = "keep_saved",
    row?: WorkbookQueryRow,
  ) {
    for (const entry of this.entries.values())
      if (
        entry.phase === "conflict" &&
        entry.intent.baseline.record_id === recordId
      ) {
        for (const contribution of this.contributions.get(entry.id) ?? [])
          contribution.conflictResolved?.(entry, kind, row);
        this.update(entry.id, {
          phase: "rejected",
          failure: {
            kind: "stale_target",
            message:
              "The field conflict was resolved. Review saved values before submitting any retained draft.",
          },
        });
      }
  }
  dismiss(id: string) {
    const entry = this.entries.get(id);
    if (
      entry &&
      (["preparation_failed", "rejected"].includes(entry.phase) ||
        (entry.phase === "acknowledged" && entry.reconciliation === "complete"))
    ) {
      this.entries.delete(id);
      this.attempts.delete(id);
      this.contributions.delete(id);
      this.emit();
    }
  }
}
function immutableClone<T>(value: T): T {
  const clone = structuredClone(value);
  const freeze = (item: unknown) => {
    if (item && typeof item === "object") {
      for (const child of Object.values(item)) freeze(child);
      Object.freeze(item);
    }
  };
  freeze(clone);
  return clone;
}
