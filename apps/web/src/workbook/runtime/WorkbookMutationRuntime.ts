import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import { type SheetRef, sheetRefKey } from "../../shared/sheetRef";
import type { EntityMergeReview } from "../features/entities/entityMergeReview";
import { WorkbookEntityMergeOwner } from "../features/entities/WorkbookEntityMergeOwner";
import { WorkbookRecordHistoryOwner } from "../history/WorkbookRecordHistoryOwner";
import type { WorkbookMutationInvalidationReason } from "../lifecycle/workbookInvalidation";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type { EntityRecordWriteTarget } from "../mutations/entityRecordWriteBoundary";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import { executeWorkbookConflictResolution } from "../mutations/workbookConflictResolutionAdapter";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import type {
  PendingReplayRecoveryRefusal,
  PendingReplayScope,
} from "../utils/workbookPendingQueue";
import { WorkbookClientTransactionLedger } from "./WorkbookClientTransactionLedger";
import {
  createWorkbookConflictStore,
  type WorkbookConflictRegistration,
  type WorkbookConflictStore,
} from "./WorkbookConflictStore";
import {
  createWorkbookManagedPatchDriver,
  type WorkbookManagedPatchDriver,
  type WorkbookQueuedPatchRequest,
} from "./WorkbookManagedPatchDriver";
import {
  createWorkbookMutationDriverRegistry,
  type WorkbookMutationDriver,
  type WorkbookMutationDriverRegistration,
  type WorkbookMutationDriverRegistry,
  type WorkbookMutationOwnerEnvelope,
} from "./WorkbookMutationDriverRegistry";
import { WorkbookRetryScheduler } from "./WorkbookRetryScheduler";
import { WorkbookRuntimeLifecycle } from "./WorkbookRuntimeLifecycle";
import {
  type WorkbookSurfaceBlockedEditDiscard,
  type WorkbookSurfaceConflictFocusRestore,
  type WorkbookSurfaceRefresh,
  WorkbookSurfaceRegistry,
  type WorkbookSurfaceResolvedMutationApply,
} from "./WorkbookSurfaceRegistry";
import {
  buildWorkbookConflictResolutionPayload,
  type WorkbookConflictEntry,
  type WorkbookConflictResolutionKind,
  workbookConflictEntry,
} from "./workbookConflictModel";
import {
  projectWorkbookMutationStatus,
  type WorkbookMutationSnapshot,
  type WorkbookRefreshStatusFact,
} from "./workbookMutationStatusProjector";
import {
  createWorkbookPendingQueueRuntime,
  type WorkbookPendingQueueRuntime,
} from "./workbookPendingReplayRuntime";
import {
  browserWorkbookRuntimeDependencies,
  type WorkbookRuntimeDependencies,
  type WorkbookSchedulerPort,
} from "./workbookRuntimePorts";

const entityViewSchemas: ReadonlySet<string> = new Set([
  hostsViewSchemaId,
  identitiesViewSchemaId,
]);

export type { WorkbookQueuedPatchRequest } from "./WorkbookManagedPatchDriver";
export type {
  WorkbookMutationSnapshot,
  WorkbookStatusPresentation,
} from "./workbookMutationStatusProjector";

export type WorkbookSaveAnnouncement = {
  readonly sequence: number;
  readonly priority: "polite" | "assertive";
  readonly message: string;
};

export type WorkbookEditRecoveryActionResult =
  | { readonly ok: true }
  | {
      readonly ok: false;
      readonly reason:
        | PendingReplayRecoveryRefusal
        | "origin_refused"
        | "secure_id_unavailable";
    };

/**
 * Shell-lifetime authority for Workbook mutation recovery and save state.
 *
 * The queue is scoped by incident and browser-tab client instance. Timeline
 * claims its exact queue envelopes through the registered Timeline driver,
 * while the managed-patch driver owns renderer-neutral Base-surface patches.
 * Both paths share the same FIFO, conflict gate, capacity, transport, retry
 * scheduling, transaction ledger, and save-state projection.
 */
export class WorkbookMutationRuntime {
  readonly scope: PendingReplayScope;
  readonly history: WorkbookRecordHistoryOwner;
  readonly entityMerge: WorkbookEntityMergeOwner;
  private readonly transactionIds: SecureTransactionIdPort;
  private readonly pendingRuntime: WorkbookPendingQueueRuntime;
  private readonly pendingMutationPort: WorkbookPendingMutationPort;
  private readonly scheduler: WorkbookSchedulerPort;
  private readonly conflicts: WorkbookConflictStore;
  private readonly drivers: WorkbookMutationDriverRegistry;
  private readonly ledger: WorkbookClientTransactionLedger;
  private readonly lifecycle: WorkbookRuntimeLifecycle;
  private readonly managedPatches: WorkbookManagedPatchDriver;
  private readonly retryScheduler: WorkbookRetryScheduler;
  private readonly surfaces: WorkbookSurfaceRegistry;
  private readonly refreshStatusBySheet = new Map<
    string,
    WorkbookRefreshStatusFact
  >();
  private explicitInFlightCount = 0;
  private entityLifetimeRetired = false;
  private readonly entityWrites = new Map<symbol, EntityRecordWriteTarget>();
  private snapshot: WorkbookMutationSnapshot;
  private saveAnnouncement: WorkbookSaveAnnouncement | null = null;
  private announcementSequence = 0;
  private announcedSequence = 0;

  constructor(
    scope: PendingReplayScope,
    transactionIds: SecureTransactionIdPort,
    pendingMutationPort: WorkbookPendingMutationPort,
    dependencies: WorkbookRuntimeDependencies = browserWorkbookRuntimeDependencies,
  ) {
    this.scope = { ...scope };
    this.entityMerge = new WorkbookEntityMergeOwner(
      scope.incidentId,
      transactionIds,
      {
        canReserve: () => !this.lifecycle.disposed,
        coordinate: (review, signal) =>
          this.coordinateEntityMerge(review, signal),
      },
    );
    this.history = new WorkbookRecordHistoryOwner(
      scope.incidentId,
      transactionIds,
      undefined,
      (recordId) => !this.entityMerge.blocksRecord(recordId),
    );
    this.transactionIds = transactionIds;
    this.pendingMutationPort = pendingMutationPort;
    this.pendingRuntime = createWorkbookPendingQueueRuntime(this.scope);
    this.scheduler = dependencies.scheduler;
    this.conflicts = createWorkbookConflictStore();
    this.drivers = createWorkbookMutationDriverRegistry();
    this.ledger = new WorkbookClientTransactionLedger();
    this.lifecycle = new WorkbookRuntimeLifecycle(dependencies.scheduler);
    this.retryScheduler = new WorkbookRetryScheduler(dependencies.scheduler);
    this.surfaces = new WorkbookSurfaceRegistry(() => this.emit());
    this.managedPatches = createWorkbookManagedPatchDriver({
      beginMutationReport: () => this.beginExplicitMutation(),
      clock: dependencies.clock,
      conflicts: this.conflicts,
      drivers: this.drivers,
      emit: () => this.emit(),
      executeMutation: (input) => this.dispatchPendingMutation(input),
      ledger: this.ledger,
      pendingRuntime: this.pendingRuntime,
      requestDrain: () => this.requestDrain(),
      retryScheduler: this.retryScheduler,
      scope: this.scope,
      surfaces: this.surfaces,
      transactionIds,
    });
    const managedDriverRegistration = this.drivers.register(
      this.managedPatches,
    );
    if (!managedDriverRegistration.accepted) {
      throw new Error("managed-patch mutation driver registration failed");
    }
    this.snapshot = this.calculateSnapshot();
    this.history.subscribe(() => {
      for (const entry of this.history.getSnapshot()) {
        const receipt = entry.receipt;
        if (
          receipt &&
          entityViewSchemas.has(entry.attempt.subject.viewSchemaId)
        )
          this.entityMerge.acceptVersion(receipt.recordId, receipt.rowVersion);
      }
      this.emit();
    });
    this.entityMerge.subscribe(() => this.emit());
  }

  async coordinateHistory(
    recordId: string,
    signal: AbortSignal,
  ): Promise<number | null> {
    while (!signal.aborted && !this.lifecycle.disposed) {
      const queue = this.pendingRuntime.model.snapshot();
      if (
        queue.authPaused ||
        queue.halted ||
        queue.overflow ||
        queue.sameFieldConflicts.length ||
        this.conflicts.entries().length
      )
        return null;
      if (!queue.units.some((unit) => unit.recordId === recordId))
        return this.history.latestVersion(recordId);
      await new Promise<void>((resolve) => window.setTimeout(resolve, 16));
    }
    return null;
  }

  beginEntityWrite(target: EntityRecordWriteTarget): (() => void) | null {
    if (
      this.entityLifetimeRetired ||
      this.lifecycle.disposed ||
      target.recordIds.some((id) => this.entityMerge.blocksRecord(id)) ||
      (target.unknownEntityType &&
        this.entityMerge.blocksEntityType(target.unknownEntityType))
    )
      return null;
    const token = Symbol("entity write");
    this.entityWrites.set(token, target);
    return () => {
      this.entityWrites.delete(token);
    };
  }

  acceptEntityVersion(recordId: string, version: number): void {
    if (this.entityLifetimeRetired) return;
    this.entityMerge.acceptVersion(recordId, version);
    this.history.acceptVersion(recordId, version);
  }

  private async coordinateEntityMerge(
    review: EntityMergeReview,
    signal: AbortSignal,
  ): Promise<boolean> {
    const ids = [review.survivor.recordId, review.loser.recordId];
    while (!signal.aborted && !this.lifecycle.disposed) {
      const queue = this.pendingRuntime.model.snapshot();
      const queued = queue.units.some((unit) =>
        ids.includes(unit.recordId ?? ""),
      );
      if (
        queued &&
        (queue.authPaused ||
          queue.halted ||
          queue.overflow ||
          queue.sameFieldConflicts.length ||
          this.conflicts.entries().length)
      )
        return false;
      const history = this.history
        .getSnapshot()
        .filter((entry) => ids.includes(entry.attempt.subject.recordId));
      if (history.some((entry) => entry.phase === "uncertain")) return false;
      const earlierHistory = history.some(
        (entry) =>
          entry.transportPending ||
          entry.phase === "preparing" ||
          entry.phase === "submitting",
      );
      const direct = [...this.entityWrites.values()].some(
        (target) =>
          target.recordIds.some((id) => ids.includes(id)) ||
          target.unknownEntityType === review.entityType,
      );
      if (!queued && !earlierHistory && !direct) return true;
      await new Promise<void>((resolve) => window.setTimeout(resolve, 16));
    }
    return false;
  }

  pendingQueue(): WorkbookPendingQueueRuntime {
    return this.pendingRuntime;
  }

  visibleEdit(
    viewSchemaId: string,
    recordId: string,
    fieldKey: string,
  ): unknown | undefined {
    return this.managedPatches.visibleEdit(viewSchemaId, recordId, fieldKey);
  }

  getSnapshot = (): WorkbookMutationSnapshot => this.snapshot;

  private calculateSnapshot(): WorkbookMutationSnapshot {
    return projectWorkbookMutationStatus({
      conflictPanelOpen: this.conflicts.panelOpen,
      conflicts: this.conflicts.entries(),
      explicitInFlightCount:
        this.explicitInFlightCount +
        this.history.pendingCount +
        this.entityMerge.pendingCount,
      explicitRecoveryBlocked:
        this.history.blockedCount > 0 || this.entityMerge.blockedCount > 0,
      queue: this.pendingRuntime.model.snapshot(),
      refreshes: Array.from(this.refreshStatusBySheet.values()),
    });
  }

  subscribe = (listener: () => void): (() => void) =>
    this.lifecycle.subscribe(listener);

  registerDriver(
    driver: WorkbookMutationDriver,
  ): WorkbookMutationDriverRegistration {
    return this.drivers.register(driver);
  }

  claimMutationUnit(
    unitId: string,
    envelope: WorkbookMutationOwnerEnvelope,
  ): void {
    this.drivers.claim(unitId, envelope);
  }

  releaseMutationUnit(unitId: string): void {
    this.drivers.release(unitId);
  }

  rememberClientTransaction(clientTxnId: string): void {
    this.ledger.remember(clientTxnId);
  }

  dispatchPendingMutation(
    input: Parameters<WorkbookPendingMutationPort["execute"]>[0],
  ): ReturnType<WorkbookPendingMutationPort["execute"]> {
    this.ledger.remember(input.unit.clientTxnId);
    return this.pendingMutationPort.execute(input).then((outcome) => {
      if (
        outcome.kind === "accepted" &&
        entityViewSchemas.has(outcome.value.viewSchemaId)
      )
        this.acceptEntityVersion(
          outcome.value.row.record_id,
          outcome.value.row.row_version,
        );
      return outcome;
    });
  }

  scheduleRetry(delayMilliseconds: number): boolean {
    return this.retryScheduler.schedule(delayMilliseconds, () =>
      this.requestDrain(),
    );
  }

  registerSurface(
    viewSchemaId: string,
    refresh: WorkbookSurfaceRefresh,
    applyResolvedMutation?: WorkbookSurfaceResolvedMutationApply,
    restoreConflictFocus?: WorkbookSurfaceConflictFocusRestore,
    discardBlockedEdit?: WorkbookSurfaceBlockedEditDiscard,
  ): () => void {
    return this.surfaces.register(
      viewSchemaId,
      refresh,
      applyResolvedMutation,
      restoreConflictFocus,
      discardBlockedEdit,
    );
  }

  beginExplicitMutation(): () => void {
    this.explicitInFlightCount += 1;
    this.emit();
    let finished = false;
    return () => {
      if (finished) return;
      finished = true;
      this.explicitInFlightCount = Math.max(0, this.explicitInFlightCount - 1);
      this.emit();
    };
  }

  notifyPendingChanged(): void {
    this.emit();
  }

  beginRefreshStatus(sheetRef: SheetRef): () => void {
    const key = sheetRefKey(sheetRef);
    const previous = this.refreshStatusBySheet.get(key);
    this.refreshStatusBySheet.set(key, {
      sheetRef,
      count: (previous?.count ?? 0) + 1,
    });
    this.emit();
    let finished = false;
    return () => {
      if (finished) return;
      finished = true;
      const count = (this.refreshStatusBySheet.get(key)?.count ?? 1) - 1;
      if (count === 0) this.refreshStatusBySheet.delete(key);
      else this.refreshStatusBySheet.set(key, { sheetRef, count });
      this.emit();
    };
  }

  enqueuePatch(request: WorkbookQueuedPatchRequest): GridEditCommitOutcome {
    if (this.entityMerge.blocksRecord(request.recordId))
      return {
        kind: "rejected_mutation",
        message:
          "This record has a pending merge. Recover it in Merge actions before editing.",
      };
    return this.managedPatches.enqueue(request);
  }

  updateConflictDraft(key: string, mergedDraft: string): void {
    if (this.conflicts.updateDraft(key, mergedDraft)) this.emit();
  }

  registerConflict({
    sheetRef,
    conflict,
    focusKey = null,
    refresh,
    rowLabel,
    surfaceLabel,
    viewSchemaId,
  }: WorkbookConflictRegistration): WorkbookConflictEntry {
    const entry = this.conflicts.register({
      sheetRef,
      conflict,
      focusKey,
      refresh,
      rowLabel,
      surfaceLabel,
      viewSchemaId,
    });
    this.emit();
    return entry;
  }

  clearConflict(key: string): void {
    const conflict = this.conflicts.clear(key);
    if (conflict !== undefined) {
      this.managedPatches.clearVisibleConflict(conflict);
    }
    this.pendingRuntime.model.clearSameFieldConflict(key);
    this.emit();
    this.requestDrain();
  }

  activateConflict(): void {
    this.conflicts.activate();
    this.emit();
  }

  dismissConflict(key: string): void {
    const conflict = this.conflicts.dismiss(key);
    if (conflict === undefined) return;
    this.emit();
    const restore = this.surfaces.restoreConflictFocus(
      conflict.origin.viewSchemaId,
    );
    if (restore !== null) {
      this.scheduler.enqueueMicrotask(() => restore(conflict));
    }
  }

  async retryBlockedEdit(): Promise<WorkbookEditRecoveryActionResult> {
    const halted = this.pendingRuntime.model.snapshot().halted;
    if (halted === null) return { ok: false, reason: "not_halted" };
    let transactionId: string;
    try {
      transactionId = this.transactionIds.create("workbook-recovery");
    } catch {
      return { ok: false, reason: "secure_id_unavailable" };
    }
    const result = this.pendingRuntime.model.retryHaltedWithNewClientTxnId(
      halted.unit_id,
      transactionId,
    );
    if (!result.recovered) {
      return { ok: false, reason: result.reason };
    }
    this.emit();
    this.requestDrain();
    return { ok: true };
  }

  async discardBlockedEdit(): Promise<WorkbookEditRecoveryActionResult> {
    const queue = this.pendingRuntime.model.snapshot();
    const halted = queue.halted;
    if (halted === null) return { ok: false, reason: "not_halted" };
    const haltedUnit = queue.units.find((unit) => unit.id === halted.unit_id);
    const surfaceDiscard =
      haltedUnit === undefined
        ? null
        : this.surfaces.discardBlockedEdit(haltedUnit.viewSchemaId);
    if (surfaceDiscard !== null) {
      if (!(await surfaceDiscard(halted.unit_id))) {
        return { ok: false, reason: "origin_refused" };
      }
      this.emit();
      this.requestDrain();
      return { ok: true };
    }
    const result = this.pendingRuntime.model.discardHaltedUnit(halted.unit_id);
    if (!result.recovered) {
      return { ok: false, reason: result.reason };
    }
    const meta = this.managedPatches.discard(result.unit);
    this.emit();
    if (meta !== undefined) await this.surfaces.refresh(meta.viewSchemaId);
    this.requestDrain();
    return { ok: true };
  }

  async resolveConflict({
    apiBase,
    key,
    resolutionKind,
  }: {
    readonly apiBase?: string | undefined;
    readonly key: string;
    readonly resolutionKind: WorkbookConflictResolutionKind;
  }): Promise<string | null> {
    const entry = this.conflicts.get(key);
    if (entry === undefined) return "The conflict is no longer available.";
    const releaseRecord = this.beginEntityWrite({
      recordIds: [entry.conflict.record_id],
    });
    if (releaseRecord === null)
      return "Recover this record's pending merge before resolving its edit.";
    let transactionId: string;
    try {
      transactionId = this.transactionIds.create(
        "workbook-conflict-resolution",
      );
    } catch {
      releaseRecord();
      return "A secure transaction ID could not be created. No resolution was sent.";
    }
    const body = buildWorkbookConflictResolutionPayload({
      clientTxnId: transactionId,
      entry,
      resolutionKind,
    });
    if (body === null) {
      releaseRecord();
      return "The reviewed collection contains a change that cannot be represented safely.";
    }
    const finishMutation = this.beginExplicitMutation();
    try {
      const outcome = await executeWorkbookConflictResolution({
        apiBase,
        conflictToken: entry.conflict.conflict_token,
        recordId: entry.conflict.record_id,
        request: body,
      });
      if (outcome.kind === "rejected") {
        if (outcome.failure.kind === "same_field_conflict") {
          const refreshedEntry = workbookConflictEntry({
            conflict: outcome.failure.conflict,
            focusKey: entry.focusKey,
            rowLabel: entry.origin.rowLabel,
            surfaceLabel: entry.origin.surfaceLabel,
            viewSchemaId: entry.origin.viewSchemaId,
            sheetRef: entry.origin.sheetRef,
          });
          this.conflicts.replace({
            ...refreshedEntry,
            mergedDraft: entry.mergedDraft,
          });
          this.emit();
          return "The saved value changed again. Review the refreshed conflict.";
        }
        if (
          outcome.failure.kind === "validation" &&
          outcome.failure.message === "invalid_mutation_payload"
        ) {
          return await this.refreshInvalidConflictToken(key, entry);
        }
        return outcome.failure.message;
      }
      const resolvedRow = outcome.value.row;
      if (
        entityViewSchemas.has(outcome.value.viewSchemaId) &&
        resolvedRow !== null &&
        typeof resolvedRow === "object" &&
        "record_id" in resolvedRow &&
        resolvedRow.record_id === entry.conflict.record_id &&
        "row_version" in resolvedRow &&
        typeof resolvedRow.row_version === "number"
      )
        this.acceptEntityVersion(
          entry.conflict.record_id,
          resolvedRow.row_version,
        );
      this.clearConflict(key);
      const applyResolvedMutation = this.surfaces.applyResolvedMutation(
        entry.origin.viewSchemaId,
      );
      if (applyResolvedMutation === null) {
        await this.surfaces.refresh(entry.origin.viewSchemaId);
      } else {
        await applyResolvedMutation(outcome.value, entry);
      }
      return null;
    } finally {
      releaseRecord();
      finishMutation();
    }
  }

  requestDrain(): void {
    this.lifecycle.requestDrain(async () => {
      const candidate = this.pendingRuntime.model.peekNextQueued();
      if (candidate === null) return;
      await this.drivers.drain(candidate.unit);
    });
  }

  applyAuthorizationRecoveryState(state: "paused" | "resumed"): void {
    if (state === "paused") {
      this.history.suspend();
      this.entityMerge.suspend();
      this.pendingRuntime.model.pauseForAuthRecovery();
      this.emit();
      return;
    }
    this.pendingRuntime.model.resumeAfterAuthRecovery();
    this.emit();
    this.requestDrain();
  }

  /** Called only after version-fenced current incident observation. No automatic replay. */
  observeIncidentReopened(): void {
    this.pendingRuntime.model.resumeAfterIncidentReopen();
    this.emit();
  }

  pauseForTerminalLifecycle(): void {
    this.pendingRuntime.model.pauseForTerminalLifecycle();
    this.emit();
  }

  invalidate(reason: WorkbookMutationInvalidationReason): void {
    if (reason.kind === "runtime_disposed") {
      if (this.lifecycle.disposed) return;
      this.entityLifetimeRetired = true;
      this.entityWrites.clear();
      this.history.retire();
      this.entityMerge.retire();
      this.retryScheduler.cancel();
      this.managedPatches.dispose();
      for (const unit of this.pendingRuntime.model.snapshot().units)
        this.drivers.release(unit.id);
      this.pendingRuntime.model.retire();
      for (const conflict of this.conflicts.entries())
        this.conflicts.clear(conflict.key);
      this.refreshStatusBySheet.clear();
      this.explicitInFlightCount = 0;
      this.emit();
      this.lifecycle.dispose();
      return;
    }
    if (reason.kind === "incident_closed") {
      this.history.closeIncident();
      this.entityMerge.closeIncident();
      this.pendingRuntime.model.pauseForIncidentClosure();
      this.emit();
      return;
    }
    if (reason.kind === "incident_changed") {
      this.entityLifetimeRetired = true;
      this.entityWrites.clear();
      this.history.retire();
      this.entityMerge.retire();
      this.pauseForTerminalLifecycle();
      return;
    }
    this.history.suspend();
    this.entityMerge.suspend();
    this.applyAuthorizationRecoveryState("paused");
  }

  resolveSocketClientTxn(clientTxnId: string | null | undefined): boolean {
    return this.ledger.settle(clientTxnId, this.pendingRuntime);
  }

  /** Acknowledgement lives with the runtime so shell remounts cannot replay an event. */
  takeSaveAnnouncement(): WorkbookSaveAnnouncement | null {
    if (
      this.saveAnnouncement === null ||
      this.announcedSequence === this.saveAnnouncement.sequence
    )
      return null;
    this.announcedSequence = this.saveAnnouncement.sequence;
    return this.saveAnnouncement;
  }

  private emit(): void {
    const previousLabel = this.snapshot.primaryLabel;
    this.snapshot = this.calculateSnapshot();
    if (this.snapshot.primaryLabel !== previousLabel) {
      const label = this.snapshot.primaryLabel;
      this.saveAnnouncement = {
        sequence: ++this.announcementSequence,
        priority: label === "Conflict" ? "assertive" : "polite",
        message:
          label === "Syncing"
            ? "Syncing changes"
            : label === "Conflict" && this.snapshot.unresolvedConflictCount > 0
              ? `Conflict. ${this.snapshot.unresolvedConflictCount} unresolved`
              : label,
      };
    }
    this.lifecycle.emit();
  }

  private async refreshInvalidConflictToken(
    key: string,
    entry: WorkbookConflictEntry,
  ): Promise<string | null> {
    const refresh = this.conflicts.refresh(key);
    if (refresh === undefined) return "invalid_mutation_payload";
    let outcome: WorkbookOperationOutcome<unknown>;
    try {
      outcome = await refresh();
    } catch {
      return "The conflict could not be refreshed. Your draft is still available.";
    }
    if (outcome.kind === "accepted") {
      this.clearConflict(key);
      await this.surfaces.refresh(entry.origin.viewSchemaId);
      return null;
    }
    if (outcome.failure.kind !== "same_field_conflict") {
      return `${outcome.failure.message} Your draft is still available.`;
    }
    const refreshedEntry = workbookConflictEntry({
      conflict: outcome.failure.conflict,
      focusKey: entry.focusKey,
      rowLabel: entry.origin.rowLabel,
      surfaceLabel: entry.origin.surfaceLabel,
      viewSchemaId: entry.origin.viewSchemaId,
      sheetRef: entry.origin.sheetRef,
    });
    this.conflicts.replace({
      ...refreshedEntry,
      mergedDraft: entry.mergedDraft,
    });
    this.emit();
    return "The conflict token expired. Review the refreshed conflict; your draft was preserved.";
  }
}
