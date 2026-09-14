import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { WorkbookViewApiRow } from "../../models/workbookContractRows";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { createTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { inputFocusKey } from "../models/timelineFieldRegistry";
import { timelinePendingSavesRefsFor } from "../models/timelinePendingSaves";
import {
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/timelineRowModel";
import {
  createTimelineMutationDriver,
  type TimelineMutationDriverPorts,
} from "./createTimelineMutationDriver";

type ReadSource = (
  recordId: string,
  signal: AbortSignal,
) => Promise<WorkbookViewApiRow | null>;

/** Retains Timeline dispatch and settlement. Attachments lend presentation effects only. */
export class WorkbookTimelineMutationOwner {
  private attachment: {
    readonly read: () => TimelineMutationDriverPorts;
    readonly identity: object;
  } | null = null;
  private dispatchAttachment: object | null | undefined;
  private readSource: ReadSource | null = null;
  private recovery: (() => void) | null = null;
  private readonly reads = new Set<AbortController>();
  private readonly rows = { current: [] as WorkbookRow[] };
  private readonly receipts = new Map<
    string,
    Parameters<TimelineMutationDriverPorts["applyAcceptedRowMutation"]>[1]
  >();
  private readonly driver;
  private readonly unregister: () => void;
  private retired = false;

  constructor(
    private readonly runtime: WorkbookMutationRuntime,
    ids: SecureTransactionIdPort,
  ) {
    const pending = timelinePendingSavesRefsFor(
      runtime,
      runtime.pendingQueue(),
    );
    const drafts = createTimelineEditorDraftRegistry(
      runtime.localDraftsForSurface(timelineViewSchemaId),
    );
    const presentation = () =>
      this.retired ||
      runtime.pendingQueue().model.snapshot().authPaused ||
      (this.dispatchAttachment !== undefined &&
        this.dispatchAttachment !== this.attachment?.identity)
        ? null
        : (this.attachment?.read() ?? null);
    const publish = () => {
      runtime.notifyPendingChanged();
      presentation()?.publishPendingQueueState();
    };
    const retainedRows = this.rows;
    this.driver = createTimelineMutationDriver({
      mutationRuntime: runtime,
      mutationCommands: {
        createConflictRecoveryId: () =>
          ids.create("timeline-conflict-recovery"),
        createLogicalActionId: () => ids.create("timeline-action"),
      },
      pendingSavesRefs: pending,
      rowsRef: {
        get current() {
          return presentation()?.rowsRef.current ?? retainedRows.current;
        },
        set current(rows) {
          retainedRows.current = rows;
          const mounted = presentation();
          if (mounted) mounted.rowsRef.current = rows;
        },
      },
      conflictQueueRef: {
        get current() {
          return Object.fromEntries(
            runtime.getSnapshot().conflicts.map((entry) => [entry.key, entry]),
          );
        },
      },
      get sheetRef() {
        return (
          presentation()?.sheetRef ?? {
            kind: "view_schema",
            id: timelineViewSchemaId,
          }
        );
      },
      get postMutationQueryRefreshRequired() {
        return presentation()?.postMutationQueryRefreshRequired ?? false;
      },
      beginDispatch: () => {
        this.dispatchAttachment = this.attachment?.identity ?? null;
        return () => {
          this.dispatchAttachment = undefined;
        };
      },
      readCurrentRow: async (unit) => {
        if (unit.recordId === null)
          return (
            pending.replayContextByUnitId.get(unit.id)?.rowSnapshot ?? null
          );
        const mounted = presentation()?.latestCommittedTimelineRow(
          unit.recordId,
        );
        if (mounted) return mounted;
        if (!this.readSource || this.retired)
          throw new Error("Timeline source reader is unavailable");
        const controller = new AbortController();
        this.reads.add(controller);
        const epoch = runtime.authorizationEpoch;
        try {
          const row = await this.readSource(unit.recordId, controller.signal);
          if (controller.signal.aborted || epoch !== runtime.authorizationEpoch)
            throw new Error("Timeline source lifetime changed");
          if (!row) return null;
          if (
            row.record_id !== unit.recordId ||
            row.row_version <
              (runtime.history.latestVersion(unit.recordId) ?? 0)
          )
            throw new Error(
              "Timeline source identity or version is not current",
            );
          return rowFromApi(
            normalizeTimelineFullRow(row, "retained Timeline source"),
          );
        } finally {
          this.reads.delete(controller);
        }
      },
      latestCommittedTimelineRow: (id) => {
        const row = runtime.explicitPatches.latestRow(id);
        return (
          presentation()?.latestCommittedTimelineRow(id) ??
          (row
            ? rowFromApi(
                normalizeTimelineFullRow(row, "accepted Timeline source"),
              )
            : null)
        );
      },
      applyAcceptedRowMutation: (key, accepted, options) => {
        this.receipts.set(accepted.changeSetId, accepted);
        runtime.retainSurfaceRefreshDebt(timelineViewSchemaId);
        const row = rowFromApi(accepted.row);
        this.rows.current = this.rows.current
          .filter((item) => item.key !== key && item.recordId !== row.recordId)
          .concat(row);
        return (
          presentation()?.applyAcceptedRowMutation(key, accepted, options) ??
          row
        );
      },
      clearSubmittedScalarEditorDraftValuesForRow: (...args) => {
        drafts.clearSubmittedRow(...args);
        presentation()?.clearSubmittedScalarEditorDraftValuesForRow(...args);
      },
      clearViewportContinuity: (token) =>
        presentation()?.clearViewportContinuity(token),
      registerMutationConflict: (
        conflict,
        rowKey,
        field,
        surface,
        refresh,
        sheetRef,
      ) => {
        const mounted = presentation();
        if (mounted)
          return mounted.registerMutationConflict(
            conflict,
            rowKey,
            field,
            surface,
            refresh,
            sheetRef,
          );
        runtime.registerConflict({
          conflict,
          sheetRef,
          refresh,
          focusKey: inputFocusKey(rowKey, field, surface),
          rowLabel: conflict.record_id,
          surfaceLabel: "Timeline",
          viewSchemaId: timelineViewSchemaId,
        });
        return true;
      },
      loadRows: async (options) => {
        await presentation()?.loadRows(options);
      },
      publishPendingQueueState: publish,
      reconcileDiscardedPendingUnit: (...args) =>
        presentation()?.reconcileDiscardedPendingUnit(...args),
      recordWorkbookTiming: (name, details) => {
        if (typeof performance !== "undefined")
          performance.mark(`cartulary.workbook.${name}`, { detail: details });
      },
      requestAuthorizationRecovery: () => {
        if (this.recovery) this.recovery();
        else presentation()?.requestAuthorizationRecovery();
      },
      setRefreshError: (message) => presentation()?.setRefreshError(message),
      rowStoreCommands: {
        replaceRows: (rows) => {
          this.rows.current = rows;
          presentation()?.rowStoreCommands.replaceRows(rows);
        },
        updateRows: (update) => {
          this.rows.current = update(this.rows.current);
        },
      },
    });
    const registration = runtime.registerDriver({
      kind: "timeline_row",
      drain: this.driver.drain,
    });
    if (!registration.accepted)
      throw new Error("Timeline mutation owner is already registered.");
    this.unregister = registration.unregister;
  }

  configureReader(read: ReadSource) {
    this.readSource = read;
  }
  bindAuthorizationRecovery(recover: () => void) {
    this.recovery = recover;
    return () => {
      if (this.recovery === recover) this.recovery = null;
    };
  }
  attach(read: () => TimelineMutationDriverPorts) {
    if (this.retired) throw new Error("Timeline mutation owner is retired.");
    if (this.attachment)
      throw new Error("Timeline mutation presentation is already attached.");
    const attachment = { read, identity: {} };
    this.attachment = attachment;
    this.runtime.requestDrain();
    return () => {
      if (this.attachment !== attachment) return;
      this.attachment = null;
      this.driver.detachPresentation();
    };
  }
  readonly enqueuePendingReplayUnit: ReturnType<
    typeof createTimelineMutationDriver
  >["enqueuePendingReplayUnit"] = (...args) =>
    this.driver.enqueuePendingReplayUnit(...args);
  readonly retryBlockedEdit = (id: string) => this.driver.retryBlockedEdit(id);
  readonly discardBlockedEdit = (id: string) =>
    this.driver.discardBlockedEdit(id);
  retire() {
    this.retired = true;
    this.attachment = null;
    this.recovery = null;
    this.readSource = null;
    for (const read of this.reads) read.abort();
    this.reads.clear();
    this.receipts.clear();
    this.rows.current = [];
    this.driver.detachPresentation();
    this.unregister();
  }
}

export function timelineMutationOwnerFor(
  runtime: WorkbookMutationRuntime,
): WorkbookTimelineMutationOwner {
  return runtime.retainTimelineMutationOwner(
    (ids) => new WorkbookTimelineMutationOwner(runtime, ids),
  );
}
