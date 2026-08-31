import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import type { WorkbookMutationInvalidationReason } from "../lifecycle/workbookInvalidation";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import {
  executeWorkbookConflictResolution,
  type WorkbookResolvedMutation,
} from "../mutations/workbookConflictResolutionAdapter";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import { workbookEditRecoveryPresentation } from "../utils/workbookEditRecoveryPresentation";
import type {
  PendingQueueSnapshot,
  PendingReplayRecoveryRefusal,
  PendingReplayScope,
  PendingReplayUnitState,
} from "../utils/workbookPendingQueue";
import type { WorkbookStatusSecondaryCandidate } from "../utils/workbookStatusSecondary";
import {
  buildWorkbookConflictResolutionPayload,
  type WorkbookConflictEntry,
  type WorkbookConflictResolutionKind,
  workbookConflictEntry,
} from "./workbookConflictModel";
import { workbookPendingMutationFailureResult } from "./workbookPendingMutationSettlement";
import {
  createWorkbookPendingQueueRuntime,
  type WorkbookPendingQueueRuntime,
} from "./workbookPendingReplayRuntime";

type WorkbookManagedPatchMeta = {
  readonly fieldKey: string;
  readonly focusKey: string | null;
  readonly localValue: unknown;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
};

export type WorkbookQueuedPatchRequest = {
  readonly baseRowVersion: number;
  readonly changes: readonly Record<string, unknown>[];
  readonly fieldKey: string;
  readonly focusKey?: string | null | undefined;
  readonly localValue: unknown;
  readonly recordId: string;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
};

export type WorkbookMutationSnapshot = {
  readonly authPaused: boolean;
  readonly blockedEdit: {
    readonly kind: "client_txn_conflict" | "terminal_replay_failure";
    readonly message: string;
    readonly unitId: string;
  } | null;
  readonly conflictPanelOpen: boolean;
  readonly conflicts: readonly WorkbookConflictEntry[];
  readonly explicitInFlightCount: number;
  readonly primaryLabel: "Conflict" | "Saved" | "Syncing";
  readonly overflowMessage: string | null;
  readonly secondaryMessage: string | null;
  readonly secondaryCandidates: readonly WorkbookStatusSecondaryCandidate[];
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

export type WorkbookSurfaceSaveStateProjection = {
  readonly primaryLabel: "Conflict" | "Saved" | "Syncing";
  readonly secondaryMessage: string | null;
};

type WorkbookMutationListener = () => void;
type WorkbookSurfaceRefresh = () => Promise<void> | void;
type WorkbookSurfaceResolvedMutationApply = (
  mutation: WorkbookResolvedMutation,
  conflict: WorkbookConflictEntry,
) => Promise<void> | void;
type WorkbookSurfaceConflictFocusRestore = (
  conflict: WorkbookConflictEntry,
) => void;
type WorkbookSurfaceBlockedEditDiscard = (
  unitId: string,
) => Promise<boolean> | boolean;
type WorkbookConflictRefresh = () => Promise<WorkbookOperationOutcome<unknown>>;

/**
 * Shell-lifetime authority for Workbook mutation recovery and save state.
 *
 * The queue is scoped by incident and browser-tab client instance. Timeline
 * attaches its richer row metadata through `pending()`, while the common
 * dispatcher owns renderer-neutral Base-surface patches. Both paths share the
 * same FIFO, conflict gate, capacity, and save-state projection.
 */
export class WorkbookMutationRuntime {
  readonly scope: PendingReplayScope;
  private readonly transactionIds: SecureTransactionIdPort;
  private readonly pendingMutationPort: WorkbookPendingMutationPort;
  private readonly pendingRuntime: WorkbookPendingQueueRuntime<unknown>;
  private readonly managedPatchByUnitId = new Map<
    string,
    WorkbookManagedPatchMeta
  >();
  private readonly visibleEdits = new Map<string, unknown>();
  private readonly conflictsByKey = new Map<string, WorkbookConflictEntry>();
  private readonly conflictRefreshByKey = new Map<
    string,
    WorkbookConflictRefresh
  >();
  private readonly listeners = new Set<WorkbookMutationListener>();
  private readonly drainers = new Set<() => void>();
  private readonly refreshBySurface = new Map<string, WorkbookSurfaceRefresh>();
  private readonly resolvedApplyBySurface = new Map<
    string,
    WorkbookSurfaceResolvedMutationApply
  >();
  private readonly conflictFocusRestoreBySurface = new Map<
    string,
    WorkbookSurfaceConflictFocusRestore
  >();
  private readonly blockedEditDiscardBySurface = new Map<
    string,
    WorkbookSurfaceBlockedEditDiscard
  >();
  private readonly saveStateBySurface = new Map<
    string,
    WorkbookSurfaceSaveStateProjection
  >();
  private readonly dirtySurfaces = new Set<string>();
  private readonly recentClientTxnIds = new Set<string>();
  private explicitInFlightCount = 0;
  private conflictPanelDismissed = false;
  private drainScheduled = false;
  private disposed = false;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;
  private snapshot: WorkbookMutationSnapshot;

  constructor(
    scope: PendingReplayScope,
    transactionIds: SecureTransactionIdPort,
    pendingMutationPort: WorkbookPendingMutationPort,
  ) {
    this.scope = { ...scope };
    this.transactionIds = transactionIds;
    this.pendingMutationPort = pendingMutationPort;
    this.pendingRuntime = createWorkbookPendingQueueRuntime(this.scope);
    this.snapshot = this.calculateSnapshot();
  }

  pending<TMeta>(): WorkbookPendingQueueRuntime<TMeta> {
    return this.pendingRuntime as WorkbookPendingQueueRuntime<TMeta>;
  }

  ownsManagedUnit(unitId: string): boolean {
    return this.managedPatchByUnitId.has(unitId);
  }

  visibleEdit(
    viewSchemaId: string,
    recordId: string,
    fieldKey: string,
  ): unknown | undefined {
    return this.visibleEdits.get(
      this.visibleEditKey(viewSchemaId, recordId, fieldKey),
    );
  }

  getSnapshot = (): WorkbookMutationSnapshot => this.snapshot;

  private calculateSnapshot(): WorkbookMutationSnapshot {
    const queue = this.pendingRuntime.model.snapshot();
    const surfaceSaveStates = Array.from(this.saveStateBySurface.values());
    const surfaceConflict = surfaceSaveStates.find(
      (state) => state.primaryLabel === "Conflict",
    );
    const surfaceSyncing = surfaceSaveStates.find(
      (state) => state.primaryLabel === "Syncing",
    );
    const hasPending =
      queue.queuedCount > 0 ||
      queue.inFlightCount > 0 ||
      this.explicitInFlightCount > 0;
    const queueHasConflict =
      this.conflictsByKey.size > 0 ||
      queue.halted !== null ||
      queue.overflow !== null ||
      queue.sameFieldConflicts.length > 0;
    const primaryLabel =
      queueHasConflict || surfaceConflict !== undefined
        ? "Conflict"
        : hasPending || surfaceSyncing !== undefined
          ? "Syncing"
          : "Saved";
    const projectedSecondary =
      primaryLabel === "Conflict"
        ? surfaceConflict?.secondaryMessage
        : primaryLabel === "Syncing"
          ? surfaceSyncing?.secondaryMessage
          : null;
    const secondaryCandidates = this.secondaryCandidates(queue);
    return {
      authPaused: queue.authPaused,
      blockedEdit:
        queue.halted === null
          ? null
          : (() => {
              const presentation = workbookEditRecoveryPresentation({
                errorCode: queue.halted.error_code,
              });
              return {
                kind: presentation.kind,
                message: presentation.message,
                unitId: queue.halted.unit_id,
              };
            })(),
      conflictPanelOpen:
        this.conflictsByKey.size > 0 && !this.conflictPanelDismissed,
      conflicts: Array.from(this.conflictsByKey.values()),
      explicitInFlightCount: this.explicitInFlightCount,
      primaryLabel,
      overflowMessage: queue.overflow?.message ?? null,
      secondaryMessage:
        (queue.saveStatePresentation.primaryLabel === primaryLabel
          ? queue.saveStatePresentation.secondaryMessage
          : null) ??
        projectedSecondary ??
        (this.explicitInFlightCount > 0
          ? `${this.explicitInFlightCount} explicit change${
              this.explicitInFlightCount === 1 ? "" : "s"
            } in flight`
          : null),
      secondaryCandidates,
    };
  }

  private secondaryCandidates(
    queue: PendingQueueSnapshot,
  ): WorkbookStatusSecondaryCandidate[] {
    const candidates: WorkbookStatusSecondaryCandidate[] = [];
    const haltedSurfaceId =
      queue.halted === null
        ? undefined
        : queue.units.find((unit) => unit.id === queue.halted?.unit_id)
            ?.viewSchemaId;
    if (queue.halted !== null && haltedSurfaceId !== undefined) {
      candidates.push({
        kind:
          queue.halted.error_code === "client_txn_conflict"
            ? "client_txn_conflict"
            : "terminal_replay_failure",
        message: queue.halted.message,
        surfaceId: haltedSurfaceId,
      });
    }
    if (
      queue.overflow !== null &&
      queue.overflow.view_schema_id !== undefined
    ) {
      candidates.push({
        kind: "queue_overflow",
        message: queue.overflow.message,
        surfaceId: queue.overflow.view_schema_id,
      });
    }
    for (const conflict of queue.sameFieldConflicts) {
      if (conflict.view_schema_id === undefined) continue;
      candidates.push({
        kind: "same_field_conflict",
        message:
          queue.saveStatePresentation.secondaryMessage ??
          "Same-field conflict requires review.",
        surfaceId: conflict.view_schema_id,
      });
    }
    for (const conflict of this.conflictsByKey.values()) {
      candidates.push({
        kind: "same_field_conflict",
        message: "Same-field conflict requires review.",
        surfaceId: conflict.origin.viewSchemaId,
      });
    }
    const pendingSurfaceIds = new Set(
      queue.units.map((unit) => unit.viewSchemaId),
    );
    for (const surfaceId of pendingSurfaceIds) {
      candidates.push({
        kind: queue.authPaused
          ? "authentication_required"
          : "queued_or_in_flight",
        message: queue.authPaused
          ? "Authentication is required before queued edits can replay."
          : "Queued edits are waiting to replay.",
        surfaceId,
      });
    }
    for (const [surfaceId, state] of this.saveStateBySurface) {
      if (state.secondaryMessage === null) continue;
      candidates.push({
        kind:
          state.primaryLabel === "Conflict"
            ? "terminal_replay_failure"
            : "refresh_paused",
        message: state.secondaryMessage,
        surfaceId,
      });
    }
    return candidates;
  }

  subscribe = (listener: WorkbookMutationListener): (() => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };

  registerDrainer(drainer: () => void): () => void {
    this.drainers.add(drainer);
    return () => {
      this.drainers.delete(drainer);
    };
  }

  registerSurface(
    viewSchemaId: string,
    refresh: WorkbookSurfaceRefresh,
    applyResolvedMutation?: WorkbookSurfaceResolvedMutationApply,
    restoreConflictFocus?: WorkbookSurfaceConflictFocusRestore,
    discardBlockedEdit?: WorkbookSurfaceBlockedEditDiscard,
  ): () => void {
    this.refreshBySurface.set(viewSchemaId, refresh);
    if (applyResolvedMutation !== undefined) {
      this.resolvedApplyBySurface.set(viewSchemaId, applyResolvedMutation);
    }
    if (restoreConflictFocus !== undefined) {
      this.conflictFocusRestoreBySurface.set(
        viewSchemaId,
        restoreConflictFocus,
      );
    }
    if (discardBlockedEdit !== undefined) {
      this.blockedEditDiscardBySurface.set(viewSchemaId, discardBlockedEdit);
    }
    if (this.dirtySurfaces.delete(viewSchemaId)) {
      void Promise.resolve(refresh()).catch(() => {
        this.dirtySurfaces.add(viewSchemaId);
        this.emit();
      });
    }
    return () => {
      if (this.refreshBySurface.get(viewSchemaId) === refresh) {
        this.refreshBySurface.delete(viewSchemaId);
        this.resolvedApplyBySurface.delete(viewSchemaId);
        this.conflictFocusRestoreBySurface.delete(viewSchemaId);
        this.blockedEditDiscardBySurface.delete(viewSchemaId);
      }
    };
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

  projectSurfaceSaveState(
    viewSchemaId: string,
    projection: WorkbookSurfaceSaveStateProjection,
  ): void {
    const current = this.saveStateBySurface.get(viewSchemaId);
    if (
      current?.primaryLabel === projection.primaryLabel &&
      current.secondaryMessage === projection.secondaryMessage
    ) {
      return;
    }
    this.saveStateBySurface.set(viewSchemaId, { ...projection });
    this.emit();
  }

  clearSurfaceSaveState(viewSchemaId: string): void {
    if (this.saveStateBySurface.delete(viewSchemaId)) this.emit();
  }

  enqueuePatch(request: WorkbookQueuedPatchRequest): GridEditCommitOutcome {
    const enqueueOrder = Date.now();
    let transactionId: string;
    try {
      transactionId = this.transactionIds.create(
        `workbook-autosave-${request.viewSchemaId}`,
      );
    } catch {
      return {
        kind: "rejected_mutation",
        message:
          "This edit remains local because a secure transaction ID could not be created.",
      };
    }
    const unitId = `${transactionId}:patch`;
    this.rememberClientTxnId(transactionId);
    const admission = this.pendingRuntime.model.admit({
      id: unitId,
      kind: "patch",
      source: "autosave",
      incidentId: this.scope.incidentId,
      clientInstanceId: this.scope.clientInstanceId,
      viewSchemaId: request.viewSchemaId,
      rowKey: request.recordId,
      recordId: request.recordId,
      payloadIntent: {
        view_schema_id: request.viewSchemaId,
        base_row_version: request.baseRowVersion,
        client_txn_id: transactionId,
        changes: request.changes,
      },
      clientTxnId: transactionId,
      coalesceKey: `${request.viewSchemaId}:${request.recordId}`,
      enqueueOrder,
      operationClass: "hot_path",
      visibleEdit: {
        rowKey: request.recordId,
        fieldKey: request.fieldKey,
        value: request.localValue,
      },
    });
    if (!admission.accepted) {
      if (
        admission.status === "refused" &&
        admission.preserveVisibleEditAsUnsaved
      ) {
        this.visibleEdits.set(
          this.visibleEditKey(
            request.viewSchemaId,
            request.recordId,
            request.fieldKey,
          ),
          request.localValue,
        );
      }
      this.emit();
      return {
        kind: "rejected_mutation",
        message:
          admission.status === "duplicate"
            ? "This edit is already queued."
            : (admission.overflowMessage ??
              "This edit could not be added to the local pending queue."),
      };
    }
    this.visibleEdits.set(
      this.visibleEditKey(
        request.viewSchemaId,
        request.recordId,
        request.fieldKey,
      ),
      request.localValue,
    );
    this.managedPatchByUnitId.set(admission.unit.id, {
      fieldKey: request.fieldKey,
      focusKey: request.focusKey ?? null,
      localValue: request.localValue,
      rowLabel: request.rowLabel,
      surfaceLabel: request.surfaceLabel,
      viewSchemaId: request.viewSchemaId,
    });
    this.emit();
    this.requestDrain();
    return { kind: "accepted" };
  }

  updateConflictDraft(key: string, mergedDraft: string): void {
    const conflict = this.conflictsByKey.get(key);
    if (conflict === undefined) return;
    this.conflictsByKey.set(key, { ...conflict, mergedDraft });
    this.emit();
  }

  registerConflict({
    conflict,
    focusKey = null,
    refresh,
    rowLabel,
    surfaceLabel,
    viewSchemaId,
  }: {
    readonly conflict: Parameters<typeof workbookConflictEntry>[0]["conflict"];
    readonly focusKey?: string | null | undefined;
    readonly refresh?: WorkbookConflictRefresh | undefined;
    readonly rowLabel: string;
    readonly surfaceLabel: string;
    readonly viewSchemaId: string;
  }): WorkbookConflictEntry {
    const entry = workbookConflictEntry({
      conflict,
      focusKey,
      rowLabel,
      surfaceLabel,
      viewSchemaId,
    });
    const current = this.conflictsByKey.get(entry.key);
    if (current === undefined) this.conflictPanelDismissed = false;
    this.conflictsByKey.set(
      entry.key,
      current === undefined
        ? entry
        : {
            ...entry,
            mergedDraft:
              current.resolutionClass === entry.resolutionClass
                ? current.mergedDraft
                : entry.mergedDraft,
          },
    );
    if (refresh !== undefined)
      this.conflictRefreshByKey.set(entry.key, refresh);
    this.emit();
    return entry;
  }

  clearConflict(key: string): void {
    const conflict = this.conflictsByKey.get(key);
    if (conflict !== undefined) {
      this.visibleEdits.delete(
        this.visibleEditKey(
          conflict.origin.viewSchemaId,
          conflict.conflict.record_id,
          conflict.conflict.field_key,
        ),
      );
    }
    this.conflictsByKey.delete(key);
    this.conflictRefreshByKey.delete(key);
    this.conflictPanelDismissed = false;
    this.pendingRuntime.model.clearSameFieldConflict(key);
    this.emit();
    this.requestDrain();
  }

  activateConflict(): void {
    this.conflictPanelDismissed = false;
    this.emit();
  }

  dismissConflict(key: string): void {
    const conflict = this.conflictsByKey.get(key);
    if (conflict === undefined) return;
    this.conflictPanelDismissed = true;
    this.emit();
    const restore = this.conflictFocusRestoreBySurface.get(
      conflict.origin.viewSchemaId,
    );
    if (restore !== undefined) queueMicrotask(() => restore(conflict));
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
        ? undefined
        : this.blockedEditDiscardBySurface.get(haltedUnit.viewSchemaId);
    if (surfaceDiscard !== undefined) {
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
    const meta = this.managedPatchByUnitId.get(result.unit.id);
    this.managedPatchByUnitId.delete(result.unit.id);
    this.clearVisibleEditsForUnit(result.unit, meta?.viewSchemaId);
    this.emit();
    if (meta !== undefined) await this.refreshSurface(meta.viewSchemaId);
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
    const entry = this.conflictsByKey.get(key);
    if (entry === undefined) return "The conflict is no longer available.";
    let transactionId: string;
    try {
      transactionId = this.transactionIds.create(
        "workbook-conflict-resolution",
      );
    } catch {
      return "A secure transaction ID could not be created. No resolution was sent.";
    }
    const body = buildWorkbookConflictResolutionPayload({
      clientTxnId: transactionId,
      entry,
      resolutionKind,
    });
    if (body === null) {
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
          });
          this.conflictsByKey.set(key, {
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
      this.clearConflict(key);
      const applyResolvedMutation = this.resolvedApplyBySurface.get(
        entry.origin.viewSchemaId,
      );
      if (applyResolvedMutation === undefined) {
        await this.refreshSurface(entry.origin.viewSchemaId);
      } else {
        await applyResolvedMutation(outcome.value, entry);
      }
      return null;
    } finally {
      finishMutation();
    }
  }

  requestDrain(): void {
    if (this.disposed) return;
    for (const drainer of this.drainers) drainer();
    if (this.drainScheduled) return;
    this.drainScheduled = true;
    queueMicrotask(() => {
      this.drainScheduled = false;
      void this.drainManagedPatches();
    });
  }

  pauseForAuthRecovery(): void {
    this.pendingRuntime.model.pauseForAuthRecovery();
    this.emit();
  }

  resumeAfterAuthRecovery(): void {
    this.pendingRuntime.model.resumeAfterAuthRecovery();
    this.emit();
    this.requestDrain();
  }

  pauseForTerminalLifecycle(): void {
    this.pendingRuntime.model.pauseForTerminalLifecycle();
    this.emit();
  }

  invalidate(reason: WorkbookMutationInvalidationReason): void {
    if (reason.kind === "runtime_disposed") {
      if (this.disposed) return;
      this.disposed = true;
      if (this.retryTimer !== null) clearTimeout(this.retryTimer);
      this.retryTimer = null;
      this.pauseForTerminalLifecycle();
      this.listeners.clear();
      this.drainers.clear();
      return;
    }
    if (
      reason.kind === "incident_closed" ||
      reason.kind === "incident_changed"
    ) {
      this.pauseForTerminalLifecycle();
      return;
    }
    this.pauseForAuthRecovery();
  }

  resolveSocketClientTxn(clientTxnId: string | null | undefined): boolean {
    if (!clientTxnId) return false;
    if (this.recentClientTxnIds.delete(clientTxnId)) return true;
    return this.pendingRuntime.model
      .snapshot()
      .units.some((unit) => unit.clientTxnId === clientTxnId);
  }

  private emit(): void {
    this.snapshot = this.calculateSnapshot();
    for (const listener of this.listeners) listener();
  }

  private rememberClientTxnId(clientTxnId: string): void {
    this.recentClientTxnIds.add(clientTxnId);
    if (this.recentClientTxnIds.size <= 128) return;
    const oldest = this.recentClientTxnIds.values().next().value;
    if (typeof oldest === "string") this.recentClientTxnIds.delete(oldest);
  }

  private visibleEditKey(
    viewSchemaId: string,
    recordId: string,
    fieldKey: string,
  ): string {
    return `${viewSchemaId}\u0000${recordId}\u0000${fieldKey}`;
  }

  private clearVisibleEditsForUnit(
    unit: PendingReplayUnitState,
    viewSchemaId = unit.viewSchemaId,
  ): void {
    const changes = Array.isArray(unit.payloadIntent.changes)
      ? unit.payloadIntent.changes
      : [];
    for (const change of changes) {
      if (
        change !== null &&
        typeof change === "object" &&
        "field_key" in change &&
        typeof change.field_key === "string"
      ) {
        this.visibleEdits.delete(
          this.visibleEditKey(
            viewSchemaId,
            unit.recordId ?? unit.rowKey,
            change.field_key,
          ),
        );
      }
    }
  }

  private scheduleRetry(): void {
    if (this.disposed || this.retryTimer !== null) return;
    this.retryTimer = setTimeout(() => {
      this.retryTimer = null;
      this.requestDrain();
    }, 750);
  }

  private async refreshSurface(viewSchemaId: string): Promise<void> {
    const refresh = this.refreshBySurface.get(viewSchemaId);
    if (refresh === undefined) {
      this.dirtySurfaces.add(viewSchemaId);
      return;
    }
    try {
      await refresh();
      this.dirtySurfaces.delete(viewSchemaId);
    } catch {
      this.dirtySurfaces.add(viewSchemaId);
    }
  }

  private async refreshInvalidConflictToken(
    key: string,
    entry: WorkbookConflictEntry,
  ): Promise<string | null> {
    const refresh = this.conflictRefreshByKey.get(key);
    if (refresh === undefined) return "invalid_mutation_payload";
    let outcome: WorkbookOperationOutcome<unknown>;
    try {
      outcome = await refresh();
    } catch {
      return "The conflict could not be refreshed. Your draft is still available.";
    }
    if (outcome.kind === "accepted") {
      this.clearConflict(key);
      await this.refreshSurface(entry.origin.viewSchemaId);
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
    });
    this.conflictsByKey.set(key, {
      ...refreshedEntry,
      mergedDraft: entry.mergedDraft,
    });
    this.emit();
    return "The conflict token expired. Review the refreshed conflict; your draft was preserved.";
  }

  private async drainManagedPatches(): Promise<void> {
    if (this.conflictsByKey.size > 0) return;
    const pending = this.pendingRuntime;
    const next = pending.model.peekNextQueued();
    if (next === null) return;
    const meta = this.managedPatchByUnitId.get(next.unit.id);
    if (meta === undefined) return;
    const dispatch = pending.model.markDispatched(next.unit.id);
    if (dispatch === null) return;
    this.emit();

    let result: Awaited<ReturnType<WorkbookPendingMutationPort["execute"]>>;
    try {
      result = await this.pendingMutationPort.execute({
        committedRowVersion:
          dispatch.identity.kind === "patch"
            ? dispatch.identity.base_row_version
            : null,
        unit: dispatch.unit,
      });
    } catch {
      pending.model.settleDispatched({
        ok: false,
        status: 0,
        error: {
          code: "transport_failure",
          message: "Transport failure",
          retryable: true,
        },
      });
      this.emit();
      this.scheduleRetry();
      return;
    }

    if (result.kind === "rejected") {
      const publicFailure = workbookPendingMutationFailureResult(
        result.failure,
      );
      const settlement = pending.model.settleDispatched({
        ok: false,
        status: publicFailure.status,
        error: publicFailure.error,
      });
      if (settlement.outcome === "same_field_conflict") {
        const conflict =
          result.failure.kind === "same_field_conflict"
            ? result.failure.conflict
            : null;
        if (conflict !== null) {
          const entry = workbookConflictEntry({
            conflict,
            focusKey: meta.focusKey,
            rowLabel: meta.rowLabel,
            surfaceLabel: meta.surfaceLabel,
            viewSchemaId: meta.viewSchemaId,
          });
          this.conflictsByKey.set(entry.key, entry);
          const conflictUnit = settlement.unit;
          this.conflictRefreshByKey.set(entry.key, async () => {
            let clientTxnId: string;
            try {
              clientTxnId = this.transactionIds.create(
                "workbook-conflict-refresh",
              );
            } catch {
              return {
                kind: "rejected",
                failure: {
                  kind: "validation",
                  message: "A secure transaction ID could not be created.",
                },
              };
            }
            this.rememberClientTxnId(clientTxnId);
            const identity = conflictUnit.identity;
            if (identity.kind !== "patch") {
              return {
                kind: "rejected",
                failure: {
                  kind: "validation",
                  message: "The original conflict mutation is unavailable.",
                },
              };
            }
            const refreshedUnit: PendingReplayUnitState = {
              ...conflictUnit,
              id: `${clientTxnId}:patch`,
              clientTxnId,
              status: "in_flight",
              identity: {
                ...identity,
                client_txn_id: clientTxnId,
              },
            };
            return this.pendingMutationPort.execute({
              committedRowVersion: conflict.base_row_version,
              unit: refreshedUnit,
            });
          });
        }
        this.managedPatchByUnitId.delete(settlement.unit.id);
      } else if (settlement.outcome === "retryable_failure") {
        this.scheduleRetry();
      }
      this.emit();
      if (
        settlement.outcome !== "auth_paused" &&
        settlement.outcome !== "halted" &&
        settlement.outcome !== "same_field_conflict"
      ) {
        this.requestDrain();
      }
      return;
    }

    const settlement = pending.model.settleDispatched({
      ok: true,
      row: result.value.row,
    });
    if (settlement.outcome === "success") {
      this.managedPatchByUnitId.delete(settlement.unit.id);
      this.clearVisibleEditsForUnit(settlement.unit, meta.viewSchemaId);
      await this.refreshSurface(meta.viewSchemaId);
    }
    this.emit();
    this.requestDrain();
  }
}
