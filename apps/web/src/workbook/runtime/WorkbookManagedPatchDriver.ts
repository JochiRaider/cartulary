import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookProtocolPatchRecordRequest } from "../adapters/workbookProtocolTypes";
import type {
  WorkbookGridDraftCapture,
  WorkbookGridDraftStore,
} from "../models/WorkbookGridDraftStore";
import { workbookSavedFieldEqual } from "../models/workbookSavedValues";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookAcceptedRecordPort } from "../query/WorkbookCommittedRecordPort";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
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

export type WorkbookPatchAdmission =
  | {
      readonly kind: "admitted";
      readonly unitId: string;
      readonly completion: Promise<GridEditCommitOutcome>;
    }
  | Exclude<GridEditCommitOutcome, { readonly kind: "accepted" }>;

type PatchContributor = {
  readonly fieldKey: string;
  readonly overlayRevision: number;
  readonly draft: WorkbookGridDraftCapture | null;
  baseline: WorkbookQueryRow | null;
  readonly dependencies: readonly string[];
  readonly settle: (outcome: GridEditCommitOutcome) => void;
};

export type WorkbookQueuedPatchRequest = {
  readonly baseline?: WorkbookQueryRow;
  readonly dependencies?: readonly string[];
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
  readonly conflicts: WorkbookConflictStore;
  readonly drafts: WorkbookGridDraftStore;
  readonly records: WorkbookAcceptedRecordPort;
  readonly drivers: WorkbookMutationDriverRegistry;
  readonly emit: () => void;
  readonly executeMutation: WorkbookPendingMutationPort["execute"];
  readonly ledger: WorkbookClientTransactionLedger;
  readonly pendingRuntime: WorkbookPendingQueueRuntime;
  readonly requestDrain: () => void;
  readonly recoverAuthorization: () => void;
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
  readonly #visibleEdits = new Map<
    string,
    { value: unknown; revision: number }
  >();
  readonly #contributors = new Map<string, PatchContributor[]>();
  readonly #predecessorVersions = new Map<string, number>();
  readonly #conflictContributors = new Map<
    string,
    { unit: PendingReplayUnitState; contributors: PatchContributor[] }
  >();
  #revision = 0;
  #disposed = false;
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
    )?.value;
  }

  dispose(): void {
    this.#disposed = true;
    for (const contributors of this.#contributors.values())
      for (const contributor of contributors)
        contributor.settle({
          kind: "rejected_mutation",
          message: "This workbook session has ended.",
        });
    this.#contributors.clear();
    this.#conflictContributors.clear();
    this.#predecessorVersions.clear();
    this.#requestContextByUnitId.clear();
    this.#visibleEdits.clear();
  }

  enqueue(request: WorkbookQueuedPatchRequest): WorkbookPatchAdmission {
    if (this.#disposed)
      return {
        kind: "rejected_mutation",
        message: "This workbook session has ended.",
      };
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
    if (!admission.accepted && admission.status !== "duplicate") {
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
          admission.overflowMessage ??
          "This edit could not be added to the local pending queue.",
      };
    }
    const overlayRevision = this.#setVisibleEdit(request);
    const identity = {
      viewSchemaId: request.viewSchemaId,
      recordId: request.recordId,
      fieldKey: request.fieldKey,
    };
    const retained = this.#options.drafts.read(identity);
    const completion = new Promise<GridEditCommitOutcome>((settle) => {
      const contributors = this.#contributors.get(admission.unit.id) ?? [];
      contributors.push({
        fieldKey: request.fieldKey,
        overlayRevision,
        draft: this.#options.drafts.capture(identity),
        baseline: structuredClone(
          retained?.baseline ?? request.baseline ?? null,
        ),
        dependencies: request.dependencies ?? [],
        settle,
      });
      this.#contributors.set(admission.unit.id, contributors);
    });
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
    return { kind: "admitted", unitId: admission.unit.id, completion };
  }

  discard(
    unit: PendingReplayUnitState,
  ): WorkbookManagedPatchRequestContext | undefined {
    const meta = this.#requestContextByUnitId.get(unit.id);
    this.#requestContextByUnitId.delete(unit.id);
    this.#options.drivers.release(unit.id);
    this.#clearVisibleEditsForUnit(unit, meta?.viewSchemaId);
    this.#settleContributors(unit.id, {
      kind: "rejected_mutation",
      message:
        "The queued edit was discarded. Its raw draft remains available.",
    });
    this.#contributors.delete(unit.id);
    return meta;
  }

  clearVisibleConflict(conflict: WorkbookConflictEntry): void {
    const captured = this.#conflictContributors.get(conflict.key);
    if (!captured) return;
    this.#clearContributorEdits(captured.unit, captured.contributors);
    for (const contributor of captured.contributors)
      this.#options.drafts.acknowledge(contributor.draft);
    this.#conflictContributors.delete(conflict.key);
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
      (unit) => this.#prepareBase(unit),
    );
    if (dispatch === null) return;
    this.#options.emit();
    const result = await this.#execute(dispatch.unit, dispatch.identity);
    if (this.#disposed || result === null) return;
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
      const fields =
        result.value.row.record_id === settlement.unit.recordId &&
        settlement.unit.identity.kind === "patch"
          ? settlement.unit.identity.changes.map((change) => change.field_key)
          : [];
      this.#options.drafts.acceptPredecessor(
        meta.viewSchemaId,
        result.value.row,
        fields,
      );
      this.#predecessorVersions.set(
        result.value.row.record_id,
        result.value.row.row_version,
      );
      for (const [id, contributors] of this.#contributors) {
        if (id === settlement.unit.id) continue;
        for (const contributor of contributors) {
          if (
            contributor.baseline?.record_id !== result.value.row.record_id ||
            contributor.baseline.row_version > result.value.row.row_version
          )
            continue;
          const cells = { ...contributor.baseline.cells };
          for (const field of fields)
            if (result.value.row.cells[field])
              cells[field] = structuredClone(result.value.row.cells[field]);
          contributor.baseline = { ...contributor.baseline, cells };
        }
      }
      const contributors = this.#contributors.get(settlement.unit.id) ?? [];
      for (const contributor of contributors)
        this.#options.drafts.acknowledge(contributor.draft);
      this.#settleContributors(settlement.unit.id, { kind: "accepted" });
      this.#contributors.delete(settlement.unit.id);
      void this.#options.surfaces.refresh(meta.viewSchemaId);
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

  #setVisibleEdit(request: WorkbookQueuedPatchRequest): number {
    const revision = ++this.#revision;
    this.#visibleEdits.set(
      this.#visibleEditKey(
        request.viewSchemaId,
        request.recordId,
        request.fieldKey,
      ),
      { value: request.localValue, revision },
    );
    return revision;
  }

  #prepareBase(unit: PendingReplayUnitState): number | null {
    if (unit.identity.kind !== "patch") return null;
    const originalBase = unit.identity.base_row_version;
    if (originalBase === null) return null;
    const contributors = this.#contributors.get(unit.id) ?? [];
    const current = this.#options.records.latestRow(
      unit.recordId ?? unit.rowKey,
    );
    const floor =
      this.#options.records.latestVersion(unit.recordId ?? unit.rowKey) ??
      this.#predecessorVersions.get(unit.recordId ?? unit.rowKey) ??
      originalBase;
    const latest = [
      ...new Map(contributors.map((item) => [item.fieldKey, item])).values(),
    ];
    const stale =
      (current && current.row_version < floor) ||
      latest.some((item) =>
        item.baseline && current
          ? [item.fieldKey, ...item.dependencies].some(
              (field) =>
                !workbookSavedFieldEqual(
                  item.baseline ?? current,
                  current,
                  field,
                ),
            )
          : floor !== originalBase &&
            this.#predecessorVersions.get(unit.recordId ?? unit.rowKey) !==
              floor,
      );
    if (stale) {
      const message =
        "Review the changed saved value before submitting this retained draft. Discard the queued attempt to edit it locally.";
      this.#options.pendingRuntime.model.haltBeforeDispatch(unit.id, message);
      this.#settleContributors(unit.id, { kind: "stale_target", message });
      this.#options.emit();
      return null;
    }
    return Math.max(floor, originalBase);
  }

  #settleContributors(unitId: string, outcome: GridEditCommitOutcome) {
    const contributors = this.#contributors.get(unitId) ?? [];
    const latest = new Map(contributors.map((item) => [item.fieldKey, item]));
    for (const contributor of contributors) {
      if (outcome.kind !== "accepted")
        this.#options.drafts.setValidation(contributor.draft, outcome.message);
      contributor.settle(
        outcome.kind === "accepted" &&
          latest.get(contributor.fieldKey) !== contributor
          ? {
              kind: "superseded",
              message:
                "A newer admitted value replaced this edit before dispatch.",
            }
          : outcome,
      );
    }
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
      this.#settleContributors(settlement.unit.id, {
        kind: "conflict",
        message: "Review this edit in the conflict queue.",
      });
      if (result.failure.kind === "same_field_conflict") {
        this.#registerSettledConflict(
          result.failure.conflict,
          settlement.unit,
          meta,
        );
      }
      this.#requestContextByUnitId.delete(settlement.unit.id);
      this.#options.drivers.release(settlement.unit.id);
      this.#contributors.delete(settlement.unit.id);
    } else if (settlement.outcome === "retryable_failure") {
      this.#options.retryScheduler.schedule(750, this.#options.requestDrain);
    } else if (
      settlement.outcome === "halted" ||
      settlement.outcome === "auth_paused"
    ) {
      this.#settleContributors(settlement.unit.id, {
        kind: "rejected_mutation",
        message: result.failure.message,
      });
    }
    this.#options.emit();
    if (settlement.outcome === "auth_paused")
      this.#options.recoverAuthorization();
    if (
      settlement.outcome !== "auth_paused" &&
      settlement.outcome !== "halted" &&
      settlement.outcome !== "same_field_conflict" &&
      settlement.outcome !== "retryable_failure"
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
    this.#conflictContributors.set(entry.key, {
      unit: conflictUnit,
      contributors: this.#contributors.get(conflictUnit.id) ?? [],
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
    this.#clearContributorEdits(
      unit,
      this.#contributors.get(unit.id) ?? [],
      viewSchemaId,
    );
  }

  #clearContributorEdits(
    unit: PendingReplayUnitState,
    contributors: readonly PatchContributor[],
    viewSchemaId = unit.viewSchemaId,
  ) {
    for (const contributor of contributors) {
      const key = this.#visibleEditKey(
        viewSchemaId,
        unit.recordId ?? unit.rowKey,
        contributor.fieldKey,
      );
      if (this.#visibleEdits.get(key)?.revision === contributor.overlayRevision)
        this.#visibleEdits.delete(key);
    }
  }
}

export function createWorkbookManagedPatchDriver(
  options: WorkbookManagedPatchDriverOptions,
) {
  const state = new WorkbookManagedPatchDriverState(options);
  return {
    kind: state.kind,
    dispose: () => state.dispose(),
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
    readonly dispose: typeof state.dispose;
    readonly visibleEdit: typeof state.visibleEdit;
    readonly enqueue: typeof state.enqueue;
    readonly discard: typeof state.discard;
    readonly clearVisibleConflict: typeof state.clearVisibleConflict;
  };
}

export type WorkbookManagedPatchDriver = ReturnType<
  typeof createWorkbookManagedPatchDriver
>;
