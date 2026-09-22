import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { TimelineFileDraftPort } from "../../features/evidence/timelineFileOperation";
import type { WorkbookViewApiRow } from "../../models/workbookContractRows";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { buildStableMutationSignature } from "../../utils/workbookPendingQueue";
import { timelineMentionOwnerFor } from "../actions/timelineMentionOwnerFor";
import { buildAttachedEvidenceCreateRequest } from "../adapters/timelineEvidenceRequestBuilders";
import { createTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { TimelineCaptureLifecycle } from "../models/TimelineCaptureLifecycle";
import { projectAcceptedTimelineRow } from "../models/timelineAcceptedProjection";
import { inputFocusKey } from "../models/timelineFieldRegistry";
import { buildCreatePayload } from "../models/timelineMutationIntents";
import { timelinePendingSavesRefsFor } from "../models/timelinePendingSaves";
import {
  createDraftRow,
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/timelineRowModel";
import {
  createTimelineMutationDriver,
  type TimelineMutationDriverPorts,
} from "./createTimelineMutationDriver";

export type TimelinePresentationPorts = Omit<
  TimelineMutationDriverPorts,
  "publishPendingQueueState"
>;

type ReadSource = (
  recordId: string,
  signal: AbortSignal,
) => Promise<WorkbookViewApiRow | null>;

/** Retains Timeline dispatch and settlement. Attachments lend presentation effects only. */
export class WorkbookTimelineMutationOwner {
  private attachment: {
    readonly read: () => TimelinePresentationPorts;
    readonly identity: object;
  } | null = null;
  private dispatchAttachment: object | null | undefined;
  private readSource: ReadSource | null = null;
  private recovery: (() => void) | null = null;
  private readonly reads = new Set<AbortController>();
  private readonly rows = { current: [] as WorkbookRow[] };
  private readonly batchSources = new Map<string, readonly WorkbookRow[]>();
  private readonly receipts = new Map<
    string,
    Parameters<TimelineMutationDriverPorts["applyAcceptedRowMutation"]>[1]
  >();
  private readonly promotions = new Map<
    string,
    WorkbookPendingMutationAccepted
  >();
  private readonly fileListeners = new Set<() => void>();
  readonly fileDrafts: TimelineFileDraftPort;
  readonly capture: TimelineCaptureLifecycle;
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
    const store = runtime.localDraftsForSurface(timelineViewSchemaId);
    this.capture = new TimelineCaptureLifecycle(store);
    const drafts = createTimelineEditorDraftRegistry(store, this.capture);
    const presentation = () =>
      this.retired ||
      runtime.pendingQueue().model.snapshot().authPaused ||
      (this.dispatchAttachment !== undefined &&
        this.dispatchAttachment !== this.attachment?.identity)
        ? null
        : (this.attachment?.read() ?? null);
    const publish = () => {
      runtime.notifyPendingChanged();
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
        if (mounted?.rawRow) runtime.explicitPatches.acceptRow(mounted.rawRow);
        const cached = runtime.explicitPatches.latestRow(unit.recordId);
        const floor = Math.max(
          runtime.history.latestVersion(unit.recordId) ?? 0,
          runtime.explicitPatches.latestVersion(unit.recordId) ?? 0,
        );
        if (cached && cached.row_version >= floor)
          return rowFromApi(
            normalizeTimelineFullRow(cached, "current Timeline source"),
          );
        if (mounted && (mounted.rowVersion ?? 0) >= floor) return mounted;
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
        if (key.startsWith("draft-")) this.promotions.set(key, accepted);
        for (const listener of this.fileListeners) listener();
        if (!presentation())
          runtime.retainSurfaceRefreshDebt(timelineViewSchemaId);
        const row = rowFromApi(accepted.row);
        if (key.startsWith("draft-")) this.capture.promote(key, row);
        const mentions = this.retired ? null : timelineMentionOwnerFor(runtime);
        if (
          options?.detectAutoResolution !== false &&
          options?.operationId &&
          options.previousRow
        )
          mentions?.acceptAutoResolutions([options.previousRow], [row], {
            kind: "entry",
            operationId: options.operationId,
            changeSetId: accepted.changeSetId,
          });
        else mentions?.observeSource(row);
        if (!presentation())
          this.rows.current = projectAcceptedTimelineRow({
            committed: row,
            currentRows: this.rows.current,
            rowKey: key,
            nextDraftIndex: this.capture.allocateDraftIndex,
          }).rows;
        return (
          presentation()?.applyAcceptedRowMutation(key, accepted, options) ??
          row
        );
      },
      settleEditorRevisions: drafts.settleRevisions,
      batchAuthoring: drafts.batch,
      captureEditorDrafts: drafts.captureRow,
      acceptEditorPredecessor: (row, fields, previousValues) => {
        drafts.acceptPredecessor(row, fields, previousValues);
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
        draftRevisions,
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
            draftRevisions,
          );
        runtime.registerConflict({
          draftRevisions,
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
      recordWorkbookTiming: (name, details) => {
        if (typeof performance !== "undefined")
          performance.mark(`cartulary.workbook.${name}`, { detail: details });
      },
      requestAuthorizationRecovery: () => {
        if (this.recovery) this.recovery();
        else presentation()?.requestAuthorizationRecovery();
      },
      setRefreshError: (message) => presentation()?.setRefreshError(message),
      setMutationError: (message) => presentation()?.setMutationError(message),
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
    this.fileDrafts = {
      subscribe: (listener) => {
        this.fileListeners.add(listener);
        return () => {
          this.fileListeners.delete(listener);
        };
      },
      resolve: (key) => {
        const receipt = this.promotions.get(key);
        if (receipt) return { kind: "promoted", receipt };
        if (
          runtime
            .pendingQueue()
            .model.snapshot()
            .units.some((unit) => unit.rowKey === key && unit.kind === "create")
        )
          return { kind: "pending" };
        const row = (
          this.attachment?.read().rowsRef.current ?? this.rows.current
        ).find((row) => row.key === key);
        return row && !row.recordId
          ? { kind: "draft" }
          : { kind: "unavailable" };
      },
      attachEvidence: (key, evidenceRecordId) => {
        if (this.retired || this.fileDrafts.resolve(key).kind !== "draft")
          return;
        const original = (
          this.attachment?.read().rowsRef.current ?? this.rows.current
        ).find((row) => row.key === key);
        if (!original || original.recordId) return;
        const row = drafts.materializeRow(original);
        const clientTxnId = ids.create("timeline-file-create");
        const payloadIntent = {
          ...buildCreatePayload(row, clientTxnId),
          ...buildAttachedEvidenceCreateRequest(evidenceRecordId, clientTxnId),
        };
        const mutationSignature = buildStableMutationSignature(payloadIntent);
        this.driver.enqueuePendingReplayUnit(
          {
            id: `pending-${clientTxnId}`,
            kind: "create",
            source: "autosave",
            incidentId: runtime.scope.incidentId,
            clientInstanceId: runtime.scope.clientInstanceId,
            viewSchemaId: timelineViewSchemaId,
            rowKey: key,
            recordId: null,
            focusField: "rawActivityText",
            focusKey: inputFocusKey(key, "rawActivityText", "grid"),
            surface: "grid",
            payloadIntent,
            clientTxnId,
            mutationSignature,
            coalesceKey: `draft:${key}`,
            enqueueOrder: pending.pendingReplayOrderRef.current++,
            operationClass: "hot_path",
            status: "queued",
            rowSnapshot: row,
            continueOnFreshDraft: true,
            detectAutoResolution: false,
            promoteToCommittedRowInspect: false,
            viewportContinuityToken: undefined,
          },
          () => {
            for (const listener of this.fileListeners) listener();
          },
        );
      },
    };
    runtime.timelineFiles.configureDrafts(this.fileDrafts);
    const registration = runtime.registerDriver({
      kind: "timeline_row",
      drain: this.driver.drain,
      captureBatchSources: (attempt) => {
        this.batchSources.set(
          attempt.id,
          structuredClone(
            (presentation()?.rowsRef.current ?? this.rows.current).filter(
              (row) =>
                row.recordId && attempt.plan.recordIds.includes(row.recordId),
            ),
          ),
        );
      },
      acceptBatchPredecessor: (receipt, attempt) => {
        if (receipt.viewSchemaId === timelineViewSchemaId) {
          const before = this.batchSources.get(attempt.id) ?? [];
          timelineMentionOwnerFor(runtime).acceptAutoResolutions(
            before,
            receipt.rows.map((row) =>
              rowFromApi(normalizeTimelineFullRow(row, "accepted batch")),
            ),
            {
              kind: "batch",
              operationId: attempt.id,
              changeSetId: receipt.changeSetId,
            },
          );
        }
        this.batchSources.delete(attempt.id);
        this.driver.acceptBatchPredecessor(receipt, attempt);
      },
    });
    if (!registration.accepted)
      throw new Error("Timeline mutation owner is already registered.");
    this.unregister = registration.unregister;
  }

  initialRows = () => {
    if (!this.rows.current.some((row) => row.recordId === null))
      this.rows.current = [
        ...this.rows.current,
        createDraftRow(this.capture.allocateDraftIndex()),
      ];
    return this.rows.current;
  };

  configureReader(read: ReadSource) {
    this.readSource = read;
  }
  bindAuthorizationRecovery(recover: () => void) {
    this.recovery = recover;
    return () => {
      if (this.recovery === recover) this.recovery = null;
    };
  }
  attach(read: () => TimelinePresentationPorts) {
    if (this.retired) throw new Error("Timeline mutation owner is retired.");
    if (this.attachment)
      throw new Error("Timeline mutation presentation is already attached.");
    const attachment = { read, identity: {} };
    this.attachment = attachment;
    this.runtime.requestDrain();
    return () => {
      if (this.attachment !== attachment) return;
      this.rows.current = read().rowsRef.current;
      this.attachment = null;
      const referenced = new Set([
        ...this.runtime
          .pendingQueue()
          .model.snapshot()
          .units.map((unit) => unit.rowKey),
        ...this.runtime.timelineFiles.getSnapshot().map((entry) => entry.key),
      ]);
      for (const key of this.capture.retireUnreferenced(referenced))
        this.promotions.delete(key);
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
    timelinePendingSavesRefsFor(
      this.runtime,
      this.runtime.pendingQueue(),
    ).scalarCommits.clear();
    this.batchSources.clear();
    this.promotions.clear();
    this.capture.clear();
    this.fileListeners.clear();
    this.rows.current = [];
    this.driver.retire();
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
