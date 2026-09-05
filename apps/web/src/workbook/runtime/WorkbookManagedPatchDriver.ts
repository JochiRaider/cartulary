import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookProtocolPatchRecordRequest } from "../adapters/workbookProtocolTypes";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import type {
  PendingReplayScope,
  PendingReplayUnitState,
} from "../utils/workbookPendingQueue";
import type { WorkbookClientTransactionLedger } from "./WorkbookClientTransactionLedger";
import type { WorkbookConflictStore } from "./WorkbookConflictStore";
import type {
  WorkbookManagedPatchMutationDriver,
  WorkbookMutationDriverRegistry,
  WorkbookMutationOwnerEnvelope,
} from "./WorkbookMutationDriverRegistry";
import type { WorkbookRetryScheduler } from "./WorkbookRetryScheduler";
import type { WorkbookSurfaceRegistry } from "./WorkbookSurfaceRegistry";
import type {
  WorkbookConflictEntry,
  workbookConflictEntry,
} from "./workbookConflictModel";
import { workbookPendingMutationFailureResult } from "./workbookPendingMutationSettlement";
import type { WorkbookPendingQueueRuntime } from "./workbookPendingReplayRuntime";
import type { WorkbookClockPort } from "./workbookRuntimePorts";

type WorkbookManagedPatchRequestContext = {
  readonly fieldKey: string;
  readonly focusKey: string | null;
  readonly localValue: unknown;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
  readonly sheetRef?: SheetRef | undefined;
};

type RecordPatchChange = WorkbookProtocolPatchRecordRequest["changes"][number];

export type WorkbookQueuedPatchRequest = {
  readonly baseRowVersion: number;
  readonly changes: readonly RecordPatchChange[];
  readonly fieldKey: string;
  readonly focusKey?: string | null | undefined;
  readonly localValue: unknown;
  readonly recordId: string;
  readonly rowLabel: string;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
  readonly sheetRef?: SheetRef | undefined;
};

type WorkbookManagedPatchDriverOptions = {
  readonly clock: WorkbookClockPort;
  readonly beginMutationReport: () => () => void;
  readonly conflicts: WorkbookConflictStore;
  readonly drivers: WorkbookMutationDriverRegistry;
  readonly emit: () => void;
  readonly executeMutation: WorkbookPendingMutationPort["execute"];
  readonly ledger: WorkbookClientTransactionLedger;
  readonly pendingRuntime: WorkbookPendingQueueRuntime;
  readonly requestDrain: () => void;
  readonly retryScheduler: WorkbookRetryScheduler;
  readonly scope: PendingReplayScope;
  readonly surfaces: WorkbookSurfaceRegistry;
  readonly transactionIds: SecureTransactionIdPort;
};

/** Owns managed-patch admission, dispatch, settlement, and local projection. */
class WorkbookManagedPatchDriverState
  implements WorkbookManagedPatchMutationDriver
{
  readonly kind = "managed_patch";
  readonly #requestContextByUnitId = new Map<
    string,
    WorkbookManagedPatchRequestContext
  >();
  readonly #visibleEdits = new Map<string, unknown>();
  readonly #options: WorkbookManagedPatchDriverOptions;

  constructor(options: WorkbookManagedPatchDriverOptions) {
    this.#options = options;
  }

  visibleEdit(
    viewSchemaId: string,
    recordId: string,
    fieldKey: string,
  ): unknown | undefined {
    return this.#visibleEdits.get(
      this.#visibleEditKey(viewSchemaId, recordId, fieldKey),
    );
  }

  enqueue(request: WorkbookQueuedPatchRequest): GridEditCommitOutcome {
    const transactionId = this.#createTransactionId(request.viewSchemaId);
    if (transactionId === null) {
      return {
        kind: "rejected_mutation",
        message:
          "This edit remains local because a secure transaction ID could not be created.",
      };
    }
    const admission = this.#options.pendingRuntime.model.admit({
      id: `${transactionId}:patch`,
      kind: "patch",
      source: "autosave",
      incidentId: this.#options.scope.incidentId,
      clientInstanceId: this.#options.scope.clientInstanceId,
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
      enqueueOrder: this.#options.clock.now(),
      operationClass: "hot_path",
      presentationHint: {
        ...(request.sheetRef === undefined
          ? {}
          : { sheetRef: request.sheetRef }),
      },
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
        this.#setVisibleEdit(request);
      }
      this.#options.emit();
      return {
        kind: "rejected_mutation",
        message:
          admission.status === "duplicate"
            ? "This edit is already queued."
            : (admission.overflowMessage ??
              "This edit could not be added to the local pending queue."),
      };
    }
    this.#setVisibleEdit(request);
    this.#options.drivers.claim(admission.unit.id, {
      kind: "managed_patch",
      viewSchemaId: request.viewSchemaId,
    });
    this.#requestContextByUnitId.set(admission.unit.id, {
      fieldKey: request.fieldKey,
      focusKey: request.focusKey ?? null,
      localValue: request.localValue,
      rowLabel: request.rowLabel,
      surfaceLabel: request.surfaceLabel,
      sheetRef: request.sheetRef,
      viewSchemaId: request.viewSchemaId,
    });
    this.#options.emit();
    this.#options.requestDrain();
    return { kind: "accepted" };
  }

  discard(
    unit: PendingReplayUnitState,
  ): WorkbookManagedPatchRequestContext | undefined {
    const meta = this.#requestContextByUnitId.get(unit.id);
    this.#requestContextByUnitId.delete(unit.id);
    this.#options.drivers.release(unit.id);
    this.#clearVisibleEditsForUnit(unit, meta?.viewSchemaId);
    return meta;
  }

  clearVisibleConflict(conflict: WorkbookConflictEntry): void {
    this.#visibleEdits.delete(
      this.#visibleEditKey(
        conflict.origin.viewSchemaId,
        conflict.conflict.record_id,
        conflict.conflict.field_key,
      ),
    );
  }

  async drain(
    expectedUnit: PendingReplayUnitState,
    envelope: Extract<
      WorkbookMutationOwnerEnvelope,
      { readonly kind: "managed_patch" }
    >,
  ): Promise<void> {
    if (this.#options.conflicts.size > 0) return;
    const next = this.#options.pendingRuntime.model.peekNextQueued();
    if (
      next === null ||
      next.unit.id !== expectedUnit.id ||
      next.unit.viewSchemaId !== envelope.viewSchemaId
    ) {
      return;
    }
    const meta = this.#requestContextByUnitId.get(next.unit.id);
    if (meta === undefined) return;
    const dispatch = this.#options.pendingRuntime.model.markDispatched(
      next.unit.id,
    );
    if (dispatch === null) return;
    this.#options.emit();
    const result = await this.#execute(dispatch.unit, dispatch.identity);
    if (result === null) return;
    if (result.kind === "rejected") {
      this.#settleRejected(result, meta);
      return;
    }
    const settlement = this.#options.pendingRuntime.model.settleDispatched({
      ok: true,
      row: result.value.row,
    });
    if (settlement.outcome === "success") {
      this.#requestContextByUnitId.delete(settlement.unit.id);
      this.#options.drivers.release(settlement.unit.id);
      this.#clearVisibleEditsForUnit(settlement.unit, meta.viewSchemaId);
      const finishReport = this.#options.beginMutationReport();
      try {
        await this.#options.surfaces.refresh(meta.viewSchemaId);
      } finally {
        finishReport();
      }
    }
    this.#options.emit();
    this.#options.requestDrain();
  }

  #createTransactionId(viewSchemaId: string): string | null {
    try {
      const transactionId = this.#options.transactionIds.create(
        `workbook-autosave-${viewSchemaId}`,
      );
      this.#options.ledger.remember(transactionId);
      return transactionId;
    } catch {
      return null;
    }
  }

  #setVisibleEdit(request: WorkbookQueuedPatchRequest): void {
    this.#visibleEdits.set(
      this.#visibleEditKey(
        request.viewSchemaId,
        request.recordId,
        request.fieldKey,
      ),
      request.localValue,
    );
  }

  async #execute(
    unit: PendingReplayUnitState,
    identity: PendingReplayUnitState["identity"],
  ): Promise<Awaited<
    ReturnType<WorkbookPendingMutationPort["execute"]>
  > | null> {
    try {
      return await this.#options.executeMutation({
        committedRowVersion:
          identity.kind === "patch" ? identity.base_row_version : null,
        unit,
      });
    } catch {
      this.#options.pendingRuntime.model.settleDispatched({
        ok: false,
        status: 0,
        error: {
          code: "transport_failure",
          message: "Transport failure",
          retryable: true,
        },
      });
      this.#options.emit();
      this.#options.retryScheduler.schedule(750, this.#options.requestDrain);
      return null;
    }
  }

  #settleRejected(
    result: Extract<
      Awaited<ReturnType<WorkbookPendingMutationPort["execute"]>>,
      { readonly kind: "rejected" }
    >,
    meta: WorkbookManagedPatchRequestContext,
  ): void {
    const publicFailure = workbookPendingMutationFailureResult(result.failure);
    const settlement = this.#options.pendingRuntime.model.settleDispatched({
      ok: false,
      status: publicFailure.status,
      error: publicFailure.error,
    });
    if (settlement.outcome === "same_field_conflict") {
      if (result.failure.kind === "same_field_conflict") {
        this.#registerSettledConflict(
          result.failure.conflict,
          settlement.unit,
          meta,
        );
      }
      this.#requestContextByUnitId.delete(settlement.unit.id);
      this.#options.drivers.release(settlement.unit.id);
    } else if (settlement.outcome === "retryable_failure") {
      this.#options.retryScheduler.schedule(750, this.#options.requestDrain);
    }
    this.#options.emit();
    if (
      settlement.outcome !== "auth_paused" &&
      settlement.outcome !== "halted" &&
      settlement.outcome !== "same_field_conflict"
    ) {
      this.#options.requestDrain();
    }
  }

  #registerSettledConflict(
    conflict: Parameters<typeof workbookConflictEntry>[0]["conflict"],
    conflictUnit: PendingReplayUnitState,
    meta: WorkbookManagedPatchRequestContext,
  ): void {
    const entry = this.#options.conflicts.register({
      conflict,
      focusKey: meta.focusKey,
      rowLabel: meta.rowLabel,
      surfaceLabel: meta.surfaceLabel,
      sheetRef: meta.sheetRef,
      viewSchemaId: meta.viewSchemaId,
    });
    this.#options.conflicts.setRefresh(entry.key, async () => {
      let clientTxnId: string;
      try {
        clientTxnId = this.#options.transactionIds.create(
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
      this.#options.ledger.remember(clientTxnId);
      if (conflictUnit.identity.kind !== "patch") {
        return {
          kind: "rejected",
          failure: {
            kind: "validation",
            message: "The original conflict mutation is unavailable.",
          },
        };
      }
      return this.#options.executeMutation({
        committedRowVersion: conflict.base_row_version,
        unit: {
          ...conflictUnit,
          id: `${clientTxnId}:patch`,
          clientTxnId,
          status: "in_flight",
          identity: {
            ...conflictUnit.identity,
            client_txn_id: clientTxnId,
          },
        },
      });
    });
  }

  #visibleEditKey(
    viewSchemaId: string,
    recordId: string,
    fieldKey: string,
  ): string {
    return `${viewSchemaId}\u0000${recordId}\u0000${fieldKey}`;
  }

  #clearVisibleEditsForUnit(
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
        this.#visibleEdits.delete(
          this.#visibleEditKey(
            viewSchemaId,
            unit.recordId ?? unit.rowKey,
            change.field_key,
          ),
        );
      }
    }
  }
}

export function createWorkbookManagedPatchDriver(
  options: WorkbookManagedPatchDriverOptions,
) {
  const state = new WorkbookManagedPatchDriverState(options);
  return {
    kind: state.kind,
    visibleEdit: (viewSchemaId: string, recordId: string, fieldKey: string) =>
      state.visibleEdit(viewSchemaId, recordId, fieldKey),
    enqueue: (request: WorkbookQueuedPatchRequest) => state.enqueue(request),
    discard: (unit: PendingReplayUnitState) => state.discard(unit),
    clearVisibleConflict: (conflict: WorkbookConflictEntry) =>
      state.clearVisibleConflict(conflict),
    drain: (
      unit: PendingReplayUnitState,
      envelope: Extract<
        WorkbookMutationOwnerEnvelope,
        { readonly kind: "managed_patch" }
      >,
    ) => state.drain(unit, envelope),
  } satisfies WorkbookManagedPatchMutationDriver & {
    readonly visibleEdit: typeof state.visibleEdit;
    readonly enqueue: typeof state.enqueue;
    readonly discard: typeof state.discard;
    readonly clearVisibleConflict: typeof state.clearVisibleConflict;
  };
}

export type WorkbookManagedPatchDriver = ReturnType<
  typeof createWorkbookManagedPatchDriver
>;
