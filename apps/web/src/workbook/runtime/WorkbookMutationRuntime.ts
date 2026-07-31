import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import { apiPath } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../services/workbookApi";
import type { WorkbookMutationInvalidationReason } from "../lifecycle/workbookInvalidation";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import {
  executeWorkbookConflictResolution,
  type WorkbookResolvedMutation,
} from "../mutations/workbookConflictResolutionAdapter";
import {
  type PendingReplayScope,
  type PendingReplayUnitState,
  parsePendingReplayPublicError,
} from "../utils/workbookPendingQueue";
import {
  buildWorkbookConflictResolutionPayload,
  parseSameFieldConflict,
  type WorkbookConflictEntry,
  type WorkbookConflictResolutionKind,
  workbookConflictEntry,
} from "./workbookConflictModel";
import {
  createWorkbookPendingQueueRuntime,
  type WorkbookPendingQueueRuntime,
} from "./workbookPendingReplayRuntime";

type GenericMutationEnvelope = {
  data: {
    row: {
      record_id: string;
      row_version: number;
    };
  };
};

type WorkbookManagedPatchMeta = {
  readonly apiBase: string | undefined;
  readonly fieldKey: string;
  readonly focusKey: string | null;
  readonly localValue: unknown;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
};

export type WorkbookQueuedPatchRequest = {
  readonly apiBase?: string | undefined;
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
    readonly canRetryWithNewClientTxnId: boolean;
    readonly message: string;
    readonly unitId: string;
  } | null;
  readonly conflictPanelOpen: boolean;
  readonly conflicts: readonly WorkbookConflictEntry[];
  readonly explicitInFlightCount: number;
  readonly primaryLabel: "Conflict" | "Saved" | "Syncing";
  readonly overflowMessage: string | null;
  readonly secondaryMessage: string | null;
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
  private readonly pendingRuntime: WorkbookPendingQueueRuntime<unknown>;
  private readonly managedPatchByUnitId = new Map<
    string,
    WorkbookManagedPatchMeta
  >();
  private readonly visibleEdits = new Map<string, unknown>();
  private readonly conflictsByKey = new Map<string, WorkbookConflictEntry>();
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
  ) {
    this.scope = { ...scope };
    this.transactionIds = transactionIds;
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
    return {
      authPaused: queue.authPaused,
      blockedEdit:
        queue.halted === null
          ? null
          : {
              canRetryWithNewClientTxnId:
                queue.halted.error_code === "client_txn_conflict",
              message: queue.halted.message,
              unitId: queue.halted.unit_id,
            },
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
    };
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
      apiBase: request.apiBase,
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
    rowLabel,
    surfaceLabel,
    viewSchemaId,
  }: {
    readonly conflict: Parameters<typeof workbookConflictEntry>[0]["conflict"];
    readonly focusKey?: string | null | undefined;
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

  retryBlockedEdit(): string | null {
    const halted = this.pendingRuntime.model.snapshot().halted;
    if (halted === null) return "There is no blocked edit to retry.";
    let transactionId: string;
    try {
      transactionId = this.transactionIds.create("workbook-recovery");
    } catch {
      return "A secure replacement transaction ID could not be created.";
    }
    const result = this.pendingRuntime.model.retryHaltedWithNewClientTxnId(
      halted.unit_id,
      transactionId,
    );
    if (!result.recovered) {
      return `The blocked edit could not be retried (${result.reason}).`;
    }
    this.emit();
    this.requestDrain();
    return null;
  }

  async discardBlockedEdit(): Promise<string | null> {
    const queue = this.pendingRuntime.model.snapshot();
    const halted = queue.halted;
    if (halted === null) return "There is no blocked edit to discard.";
    const haltedUnit = queue.units.find((unit) => unit.id === halted.unit_id);
    const surfaceDiscard =
      haltedUnit === undefined
        ? undefined
        : this.blockedEditDiscardBySurface.get(haltedUnit.viewSchemaId);
    if (surfaceDiscard !== undefined) {
      if (!(await surfaceDiscard(halted.unit_id))) {
        return "The blocked edit could not be discarded by its originating surface.";
      }
      this.emit();
      this.requestDrain();
      return null;
    }
    const result = this.pendingRuntime.model.discardHaltedUnit(halted.unit_id);
    if (!result.recovered) {
      return `The blocked edit could not be discarded (${result.reason}).`;
    }
    const meta = this.managedPatchByUnitId.get(result.unit.id);
    this.managedPatchByUnitId.delete(result.unit.id);
    this.clearVisibleEditsForUnit(result.unit, meta?.viewSchemaId);
    this.emit();
    if (meta !== undefined) await this.refreshSurface(meta.viewSchemaId);
    this.requestDrain();
    return null;
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

    let result: Awaited<ReturnType<typeof fetchWorkbookJSON>>;
    try {
      result = await fetchWorkbookJSON(
        apiPath(meta.apiBase, `/api/v1/records/${dispatch.unit.recordId}`),
        {
          method: "PATCH",
          body: JSON.stringify({
            view_schema_id: dispatch.unit.viewSchemaId,
            base_row_version:
              dispatch.identity.kind === "patch"
                ? dispatch.identity.base_row_version
                : null,
            client_txn_id: dispatch.unit.clientTxnId,
            changes: dispatch.payloadIntent.changes,
          }),
        },
      );
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

    if (!result.ok) {
      const publicError = parsePendingReplayPublicError(result.payload);
      const settlement = pending.model.settleDispatched({
        ok: false,
        status: result.status,
        error: publicError,
      });
      if (settlement.outcome === "same_field_conflict") {
        const conflict = parseSameFieldConflict(result.payload);
        if (conflict !== null) {
          const entry = workbookConflictEntry({
            conflict,
            focusKey: meta.focusKey,
            rowLabel: meta.rowLabel,
            surfaceLabel: meta.surfaceLabel,
            viewSchemaId: meta.viewSchemaId,
          });
          this.conflictsByKey.set(entry.key, entry);
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

    let envelope: GenericMutationEnvelope;
    try {
      envelope = readEnvelope<GenericMutationEnvelope>(result.payload);
    } catch {
      pending.model.settleDispatched({
        ok: false,
        status: 502,
        error: {
          code: "invalid_response",
          message: parseErrorMessage(result.payload),
          retryable: true,
        },
      });
      this.emit();
      this.scheduleRetry();
      return;
    }
    const settlement = pending.model.settleDispatched({
      ok: true,
      row: envelope.data.row,
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
