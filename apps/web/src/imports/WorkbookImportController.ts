import { clientTxnID } from "../services/browserApi";
import {
  commonJobDoesNotRegress,
  terminalCommonJob,
} from "../services/commonJobContract";
import {
  type ImportClient,
  type ImportFailure,
  type ImportWriteReceipt,
  type ImportWriteResult,
  importContractFailure,
  importFailureMessage,
  importInterruptedFailure,
  sameImportSessionSource,
  sameImportUnitSource,
} from "../services/importClient";
import { importSessionIdFromReceipt } from "../services/importJobContract";
import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import {
  boundedImportRead,
  browserImportClock,
  type ImportClock,
  importTiming,
  loadImportResources,
  observeImportJob,
} from "./importCoordinator";
import {
  captureImportWrite,
  captureWorkbookUpload,
  type ImportScope,
  type ImportWriteAttempt,
  immutableImportValue,
} from "./importRequests";
import {
  createWorkbookMappingDraft,
  type ImportRectangle,
  suggestWorkbookMapping,
  validImportRegion,
  workbookMappingRequest,
} from "./workbookImportMapping";
import {
  type ImportUnitState,
  importApplyBlocker,
  importOutcomeViews,
  initialWorkbookImportState,
  terminalImportSession,
  type WorkbookImportState,
} from "./workbookImportState";

export type WorkbookImportPort = Pick<
  ImportClient,
  "send" | "readJob" | "readSession" | "listUnits" | "readUnit" | "preview"
>;
export type WorkbookImportBinding = {
  readonly scope: ImportScope;
  readonly role: WorkbookIncidentRole | null;
  readonly closed: boolean;
  readonly available: boolean;
  readonly client: WorkbookImportPort;
  readonly current: () => boolean;
  readonly accessFailure: (status: number) => void;
};

/** One volatile import workflow, independent of the drawer and row-edit queue. */
export class WorkbookImportController {
  private state = initialWorkbookImportState();
  private visible = this.state;
  private binding: WorkbookImportBinding | null = null;
  private scope: ImportScope | null = null;
  private epoch = 0;
  private writeGeneration = 0;
  private listeners = new Set<() => void>();
  private stops = new Set<AbortController>();
  private observationStop: AbortController | null = null;
  private loadStop: AbortController | null = null;
  private previewStop: AbortController | null = null;
  private navigationStop: AbortController | null = null;
  private previewQueue: string[] = [];
  private presented = false;
  private disposed = false;
  private readonly clock: ImportClock;
  private readonly transactionId: () => string;

  constructor(
    options: {
      readonly clock?: ImportClock;
      readonly transactionId?: () => string;
    } = {},
  ) {
    this.clock = options.clock ?? browserImportClock;
    this.transactionId =
      options.transactionId ?? (() => clientTxnID("workbook-import"));
  }
  getSnapshot = () => this.visible;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(change: Partial<WorkbookImportState>) {
    this.state = immutableImportValue({
      ...this.state,
      ...change,
      revision: this.state.revision + 1,
    });
    this.visible =
      this.state.access === "active"
        ? this.state
        : immutableImportValue({
            ...initialWorkbookImportState(),
            revision: this.state.revision,
            access: this.state.access,
            message:
              this.state.access === "paused"
                ? "Confirm current access to resume import review."
                : "The Import profile is unavailable.",
          });
    for (const listener of this.listeners) listener();
  }
  bind(binding: WorkbookImportBinding | null) {
    if (this.disposed) return;
    if (binding === null) {
      this.pause();
      return;
    }
    const changedScope =
      this.scope !== null &&
      (this.scope.incidentId !== binding.scope.incidentId ||
        this.scope.actorId !== binding.scope.actorId ||
        this.scope.lifetime !== binding.scope.lifetime);
    if (changedScope) this.retire();
    const previous = this.binding;
    this.binding = binding;
    this.scope = { ...binding.scope };
    if (!binding.available) {
      this.retire();
      this.binding = binding;
      this.scope = { ...binding.scope };
      return;
    }
    if (!binding.current() || binding.role === null) {
      this.pause();
      return;
    }
    if (binding.role === "") {
      this.retire();
      return;
    }
    const wasActive = this.state.access === "active";
    const changed =
      previous !== null &&
      (previous.closed !== binding.closed ||
        previous.role !== binding.role ||
        previous.client !== binding.client);
    if (changed) this.fence();
    this.publish({
      access: "active",
      canWrite: this.canWrite(),
      ...(changed
        ? {
            observing: false,
            loading: false,
            actionPending: false,
            units: this.state.units.map((u) => ({
              ...u,
              previewLoading: false,
            })),
            operation: this.interruptedOperation(this.state.operation),
            cancellation: this.interruptedOperation(this.state.cancellation),
            message: binding.closed
              ? "Closed, read-only. Local mappings are copyable drafts."
              : "Access changed. Review current state before a fresh action.",
          }
        : {}),
    });
    if ((!wasActive || changed) && this.state.job) {
      this.publish({ jobCurrent: false });
      void this.resumeObservation();
    }
  }
  setPresented(presented: boolean) {
    this.presented = presented;
    if (!presented) {
      this.navigationStop?.abort();
      this.navigationStop = null;
    }
  }
  private canRead() {
    return (
      !this.disposed &&
      this.binding !== null &&
      this.binding.available &&
      this.binding.role !== null &&
      this.binding.role !== "" &&
      this.binding.current()
    );
  }
  private canWrite() {
    return (
      this.canRead() &&
      this.binding !== null &&
      !this.binding.closed &&
      ["editor", "reviewer", "admin"].includes(this.binding.role ?? "")
    );
  }
  private canMutateSession() {
    return (
      this.canWrite() &&
      (!this.state.session ||
        (!terminalImportSession(this.state.session) &&
          this.state.session.session_status !== "applying")) &&
      !(this.state.job && !terminalCommonJob(this.state.job))
    );
  }
  private canMutateUnit(unitId: string) {
    const unit = this.unit(unitId)?.unit;
    return (
      this.canMutateSession() &&
      unit !== undefined &&
      !["applying", "applied", "failed", "rejected"].includes(unit.unit_status)
    );
  }
  private fence() {
    this.epoch++;
    for (const controller of this.stops) controller.abort();
    this.stops.clear();
    this.observationStop = null;
    this.loadStop = null;
    this.previewStop = null;
    this.navigationStop = null;
    this.previewQueue = [];
  }
  private interruptedOperation(
    operation: WorkbookImportState["operation"],
  ): WorkbookImportState["operation"] {
    return operation?.phase === "pending"
      ? {
          ...operation,
          phase: "uncertain",
          failure: importInterruptedFailure(),
        }
      : operation;
  }
  pause() {
    this.fence();
    this.binding = null;
    this.publish({
      access: "paused",
      canWrite: false,
      observing: false,
      loading: false,
      actionPending: false,
      operation: this.interruptedOperation(this.state.operation),
      cancellation: this.interruptedOperation(this.state.cancellation),
    });
  }
  retire() {
    this.fence();
    this.binding = null;
    this.scope = null;
    this.state = initialWorkbookImportState();
    this.publish({});
  }
  dispose() {
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  }
  private current(epoch: number) {
    return epoch === this.epoch && this.canRead();
  }
  private controller() {
    const controller = new AbortController();
    this.stops.add(controller);
    return controller;
  }
  private release(controller: AbortController) {
    this.stops.delete(controller);
  }
  private handleFailure(failure: ImportFailure, unitId?: string) {
    if (failure.status === 401) {
      const binding = this.binding;
      this.pause();
      binding?.accessFailure(401);
      return;
    }
    if (
      failure.kind === "authority" ||
      failure.code === "extension_profile_not_claimed"
    ) {
      this.retire();
      return;
    }
    if (failure.status === 403 || failure.code === "incident_closed") {
      this.publish({ canWrite: false });
      this.binding?.accessFailure(failure.status);
    }
    if (failure.status === 404) {
      if (unitId)
        this.publish({
          units: this.state.units.filter(
            (u) => u.unit.import_unit_id !== unitId,
          ),
          operation:
            this.state.operation &&
            "unitId" in this.state.operation.attempt &&
            this.state.operation.attempt.unitId === unitId
              ? null
              : this.state.operation,
        });
      this.binding?.accessFailure(404);
    }
  }
  private unit(unitId: string) {
    return this.state.units.find((u) => u.unit.import_unit_id === unitId);
  }
  private changeUnit(
    unitId: string,
    update: (unit: ImportUnitState) => ImportUnitState,
  ) {
    this.publish({
      units: this.state.units.map((u) =>
        u.unit.import_unit_id === unitId ? update(u) : u,
      ),
    });
  }
  private writeBusy() {
    return (
      this.state.operation?.phase === "pending" ||
      this.state.operation?.phase === "uncertain" ||
      this.state.actionPending
    );
  }

  chooseFile(file: File | null) {
    if (
      !this.canWrite() ||
      this.writeBusy() ||
      this.state.session ||
      (this.state.job && !terminalCommonJob(this.state.job))
    )
      return;
    this.publish({ file, operation: null });
  }
  startNew() {
    if (
      !this.canWrite() ||
      this.writeBusy() ||
      (this.state.job && !terminalCommonJob(this.state.job))
    )
      return;
    const binding = this.binding;
    this.retire();
    if (binding) this.bind(binding);
  }
  async upload() {
    if (
      !this.canWrite() ||
      this.writeBusy() ||
      !this.state.file ||
      !this.scope ||
      this.state.session ||
      this.state.job
    )
      return;
    try {
      await this.submit(
        captureWorkbookUpload(
          this.scope,
          this.state.file,
          this.transactionId(),
        ),
      );
    } catch {
      this.publish({
        message:
          "A secure request identity could not be created. No request was sent.",
      });
    }
  }
  updateMapping(
    unitId: string,
    change: {
      readonly targetViewSchemaId?: string;
      readonly ordinal?: number;
      readonly fieldKey?: string;
    },
  ) {
    const item = this.unit(unitId);
    if (
      !this.canMutateUnit(unitId) ||
      !this.state.canWrite ||
      !item?.draft ||
      !item.preview ||
      this.writeBusy() ||
      this.state.session?.selected_unit_ids.includes(unitId)
    )
      return;
    const draft =
      change.targetViewSchemaId !== undefined
        ? suggestWorkbookMapping(item.preview, change.targetViewSchemaId)
        : {
            ...item.draft,
            dirty: true,
            fields: {
              ...item.draft.fields,
              ...(change.ordinal === undefined
                ? {}
                : { [change.ordinal]: change.fieldKey ?? "" }),
            },
          };
    this.changeUnit(unitId, (u) => ({ ...u, draft, failure: null }));
    this.publish({ operation: null });
  }
  async approve(unitId: string, selectAfter = true) {
    const item = this.unit(unitId),
      binding = this.binding;
    if (
      !binding ||
      !this.canMutateUnit(unitId) ||
      !this.state.canWrite ||
      this.writeBusy() ||
      !item?.draft ||
      !item.preview ||
      !this.state.session
    )
      return;
    if (item.unit.approved_mapping && !item.draft.dirty) {
      if (selectAfter) await this.select(unitId, true);
      return;
    }
    try {
      const body = workbookMappingRequest(
        item.unit,
        item.preview,
        item.draft,
        this.transactionId(),
      );
      if (!body) {
        this.publish({
          message: "Correct the highlighted mapping columns before approval.",
        });
        return;
      }
      const epoch = this.epoch;
      const accepted = await this.submit(
        captureImportWrite({
          kind: "mapping",
          scope: binding.scope,
          sessionId: item.unit.import_session_id,
          unitId,
          body,
        }),
      );
      if (accepted && selectAfter && this.current(epoch))
        await this.select(unitId, true);
    } catch {
      this.publish({
        message:
          "A secure mapping request could not be created. No request was sent.",
      });
    }
  }
  async select(unitId: string, selected: boolean) {
    const item = this.unit(unitId),
      binding = this.binding;
    if (
      !binding ||
      !item ||
      !this.canMutateUnit(unitId) ||
      !this.state.canWrite ||
      this.writeBusy() ||
      !this.state.session ||
      (selected && (!item.unit.approved_mapping || item.draft?.dirty))
    )
      return;
    try {
      await this.submit(
        captureImportWrite({
          kind: selected ? "select" : "skip",
          scope: binding.scope,
          sessionId: item.unit.import_session_id,
          unitId,
          body: { client_txn_id: this.transactionId() },
        }),
      );
    } catch {
      this.publish({
        message:
          "A secure selection request could not be created. No request was sent.",
      });
    }
  }
  async createRegion(unitId: string, rect: ImportRectangle) {
    const item = this.unit(unitId),
      binding = this.binding;
    if (
      !binding ||
      !this.canMutateUnit(unitId) ||
      !this.state.canWrite ||
      this.writeBusy() ||
      !item?.preview ||
      !validImportRegion(rect, item.unit)
    )
      return;
    try {
      await this.submit(
        captureImportWrite({
          kind: "region",
          scope: binding.scope,
          sessionId: item.unit.import_session_id,
          unitId,
          body: {
            client_txn_id: this.transactionId(),
            source_rect: {
              start_row: rect.startRow,
              start_column: rect.startColumn,
              end_row: rect.endRow,
              end_column: rect.endColumn,
            },
          },
        }),
      );
    } catch {
      this.publish({
        message:
          "A secure region request could not be created. No request was sent.",
      });
    }
  }
  async retryWrite() {
    const operation = this.state.operation;
    if (
      !operation ||
      operation.phase === "pending" ||
      !this.canWrite() ||
      !this.state.canWrite
    )
      return;
    if (operation.phase === "uncertain")
      await this.submit(operation.attempt, true);
    else if (
      operation.attempt.kind === "select" ||
      operation.attempt.kind === "skip"
    )
      await this.select(
        operation.attempt.unitId,
        operation.attempt.kind === "select",
      );
    else if (operation.attempt.kind === "mapping")
      await this.approve(operation.attempt.unitId);
    else if (operation.attempt.kind === "upload") {
      this.publish({ operation: null });
      await this.upload();
    } else if (operation.attempt.kind === "apply") {
      this.publish({ operation: null });
      await this.apply();
    } else if (operation.attempt.kind === "region") {
      const rect = operation.attempt.body.source_rect;
      await this.createRegion(operation.attempt.unitId, {
        startRow: rect.start_row,
        startColumn: rect.start_column,
        endRow: rect.end_row,
        endColumn: rect.end_column,
      });
    }
  }
  private async submit(
    attempt: ImportWriteAttempt,
    replay = false,
  ): Promise<boolean> {
    const binding = this.binding;
    if (
      !binding ||
      !this.canWrite() ||
      !this.state.canWrite ||
      attempt.scope.incidentId !== binding.scope.incidentId ||
      attempt.scope.actorId !== binding.scope.actorId ||
      attempt.scope.lifetime !== binding.scope.lifetime ||
      this.state.operation?.phase === "pending" ||
      (this.state.operation?.phase === "uncertain" &&
        (!replay || attempt !== this.state.operation.attempt))
    )
      return false;
    const epoch = this.epoch,
      generation = ++this.writeGeneration,
      stop = this.controller();
    this.loadStop?.abort();
    this.previewStop?.abort();
    this.publish({
      operation: { attempt, phase: "pending", failure: null },
      loading: false,
      message:
        attempt.kind === "upload"
          ? "Uploading workbook; awaiting acknowledgement."
          : "Submitting import changes.",
    });
    const result = await boundedImportRead<ImportWriteResult>(
      async (signal) => {
        if (!this.current(epoch) || !this.canWrite())
          return { kind: "failed", failure: importInterruptedFailure() };
        const pending = binding.client.send(attempt, signal, replay);
        void pending.then(
          (outcome) => {
            // A current late acknowledgement may settle the original uncertain intent.
            if (
              outcome.kind === "accepted" &&
              this.current(epoch) &&
              this.writeGeneration === generation &&
              this.state.operation?.attempt === attempt &&
              this.state.operation.phase === "uncertain"
            )
              this.acceptReceipt(attempt, outcome.receipt);
          },
          () => {},
        );
        return { kind: "received", value: await pending };
      },
      stop.signal,
      attempt.kind === "upload" ? importTiming.upload : importTiming.request,
      this.clock,
    );
    this.release(stop);
    if (!this.current(epoch)) return false;
    const outcome: ImportWriteResult =
      result.kind === "received"
        ? result.value
        : { kind: "uncertain", failure: result.failure };
    if (outcome.kind !== "accepted") {
      this.publish({
        operation: { attempt, phase: outcome.kind, failure: outcome.failure },
        message: importFailureMessage(outcome.failure),
      });
      if ("unitId" in attempt)
        this.changeUnit(attempt.unitId, (u) => ({
          ...u,
          failure: outcome.failure,
        }));
      this.handleFailure(
        outcome.failure,
        "unitId" in attempt ? attempt.unitId : undefined,
      );
      return false;
    }
    this.acceptReceipt(attempt, outcome.receipt);
    void this.pumpPreviews();
    return true;
  }
  private acceptReceipt(
    attempt: ImportWriteAttempt,
    receipt: ImportWriteReceipt,
  ) {
    this.publish({ operation: null });
    if (receipt.kind === "job") {
      this.publish({
        job: receipt.job,
        jobPurpose: attempt.kind === "upload" ? "discovery" : "apply",
        jobCurrent: false,
        observationFailure: null,
        cancellation: null,
        message:
          "Import job accepted. Server work continues independently of this drawer.",
      });
      void this.resumeObservation();
      return;
    }
    if (receipt.kind === "selection") {
      const value = receipt.selection;
      if (!this.state.session) {
        this.publish({ loadFailure: importContractFailure() });
        return;
      }
      this.changeUnit(value.unit.import_unit_id, (u) => ({
        ...u,
        unit: value.unit,
        failure: null,
      }));
      this.publish({
        session: {
          ...this.state.session,
          session_status: value.session_status,
          selected_unit_ids: value.selected_unit_ids,
        },
        message:
          attempt.kind === "skip"
            ? "Unit skipped. Its approved mapping is retained."
            : "Mapping approved. The selected unit is ready to apply.",
      });
      return;
    }
    const unit = receipt.unit;
    if (attempt.kind === "region") {
      const exists = this.unit(unit.import_unit_id);
      if (!exists)
        this.publish({
          units: [
            ...this.state.units,
            {
              unit,
              preview: null,
              previewLoading: false,
              previewFailure: null,
              draft: null,
              failure: null,
            },
          ],
        });
      this.publish({
        message: "Operator region created. Review its mapping separately.",
      });
      this.loadPreview(unit.import_unit_id);
      return;
    }
    this.changeUnit(unit.import_unit_id, (u) => ({
      ...u,
      unit,
      draft: u.preview ? createWorkbookMappingDraft(u.preview, unit) : u.draft,
      failure: null,
    }));
    this.publish({
      message: "Mapping approved. Selection is a separate durable operation.",
    });
  }

  async refresh(): Promise<boolean> {
    const binding = this.binding,
      sessionId =
        this.state.session?.import_session_id ??
        (this.state.job ? importSessionIdFromReceipt(this.state.job) : null);
    if (
      !binding ||
      !this.canRead() ||
      !sessionId ||
      this.state.operation?.phase === "pending" ||
      this.state.cancellation?.phase === "pending"
    )
      return false;
    this.loadStop?.abort();
    const stop = this.controller();
    this.loadStop = stop;
    const epoch = this.epoch;
    this.publish({ loading: true, loadFailure: null });
    const result = await loadImportResources(
      binding.client,
      sessionId,
      stop.signal,
      this.clock,
    );
    this.release(stop);
    if (!this.current(epoch) || this.loadStop !== stop || stop.signal.aborted)
      return false;
    this.loadStop = null;
    if (result.kind === "failed") {
      this.publish({ loading: false, loadFailure: result.failure });
      if (result.failure.status === 404)
        this.publish({ session: null, units: [], file: null, operation: null });
      this.handleFailure(result.failure);
      return false;
    }
    if (
      (this.state.session &&
        !sameImportSessionSource(this.state.session, result.value.session)) ||
      result.value.units.some((unit) => {
        const previous = this.unit(unit.import_unit_id)?.unit;
        return previous && !sameImportUnitSource(previous, unit);
      })
    ) {
      this.publish({ loading: false, loadFailure: importContractFailure() });
      return false;
    }
    this.publish({
      loading: false,
      session: result.value.session,
      units: result.value.units.map((unit) => {
        const previous = this.unit(unit.import_unit_id);
        return previous
          ? {
              ...previous,
              unit,
              draft: previous.draft?.dirty
                ? previous.draft
                : previous.preview
                  ? createWorkbookMappingDraft(previous.preview, unit)
                  : null,
            }
          : {
              unit,
              preview: null,
              previewLoading: false,
              previewFailure: null,
              draft: null,
              failure: null,
            };
      }),
      message: terminalImportSession(result.value.session)
        ? "Import completed. Review session and per-unit outcomes below."
        : `Discovered ${result.value.units.length} import unit${result.value.units.length === 1 ? "" : "s"}. Review mappings and select or skip each unit.`,
    });
    const first = this.state.units[0];
    if (first && !first.preview && !terminalImportSession(result.value.session))
      this.loadPreview(first.unit.import_unit_id);
    return true;
  }
  loadPreview(unitId: string) {
    const unit = this.unit(unitId);
    if (
      !this.canRead() ||
      !unit ||
      unit.previewLoading ||
      this.previewQueue.includes(unitId)
    )
      return;
    this.previewQueue.push(unitId);
    void this.pumpPreviews();
  }
  private async pumpPreviews() {
    if (
      this.previewStop ||
      this.state.operation?.phase === "pending" ||
      this.state.cancellation?.phase === "pending" ||
      this.state.actionPending
    )
      return;
    const id = this.previewQueue.shift(),
      binding = this.binding;
    if (!id || !binding || !this.canRead()) return;
    const item = this.unit(id);
    if (!item) return;
    const stop = this.controller();
    this.previewStop = stop;
    const epoch = this.epoch;
    this.changeUnit(id, (u) => ({
      ...u,
      previewLoading: true,
      previewFailure: null,
    }));
    const result = await boundedImportRead(
      (signal) => binding.client.preview(item.unit, signal),
      stop.signal,
      importTiming.request,
      this.clock,
    );
    this.release(stop);
    if (!this.current(epoch) || this.previewStop !== stop) return;
    this.previewStop = null;
    if (result.kind === "received")
      this.changeUnit(id, (u) => ({
        ...u,
        preview: result.value,
        previewLoading: false,
        previewFailure: null,
        draft: u.draft?.dirty
          ? u.draft
          : createWorkbookMappingDraft(result.value, u.unit),
      }));
    else {
      this.changeUnit(id, (u) => ({
        ...u,
        previewLoading: false,
        previewFailure: result.failure,
      }));
      if (!stop.signal.aborted) this.handleFailure(result.failure, id);
    }
    void this.pumpPreviews();
  }
  stopObservation() {
    this.observationStop?.abort();
    this.observationStop = null;
    this.publish({
      observing: false,
      message: "Observation paused. The server job may continue.",
    });
  }
  async resumeObservation() {
    const binding = this.binding,
      initial = this.state.job;
    if (
      !binding ||
      !initial ||
      !this.canRead() ||
      this.state.cancellation?.phase === "pending"
    )
      return;
    this.observationStop?.abort();
    const stop = this.controller();
    this.observationStop = stop;
    const epoch = this.epoch;
    this.publish({ observing: true, observationFailure: null });
    const result = await observeImportJob({
      initial,
      signal: stop.signal,
      clock: this.clock,
      read: (id, signal) =>
        binding.client.readJob(
          id,
          signal,
          this.state.jobPurpose === "apply"
            ? this.state.session?.import_session_id
            : undefined,
          this.state.jobPurpose ?? undefined,
        ),
      onJob: (job) => {
        if (this.current(epoch) && this.observationStop === stop)
          this.publish({
            job,
            jobCurrent: true,
            ...(job.status === "cancel_requested" || terminalCommonJob(job)
              ? { cancellation: null }
              : {}),
          });
      },
    });
    this.release(stop);
    if (!this.current(epoch) || this.observationStop !== stop) return;
    this.observationStop = null;
    this.publish({
      observing: false,
      observationFailure: result.kind === "paused" ? result.failure : null,
    });
    if (result.kind === "paused") {
      if (result.failure.status === 404) this.publish({ jobCurrent: false });
      this.handleFailure(result.failure);
    } else {
      await this.refresh();
    }
  }
  async apply() {
    const binding = this.binding,
      session = this.state.session;
    const blocker = importApplyBlocker(this.state);
    if (!binding || !session || !this.canWrite() || blocker) {
      if (blocker) this.publish({ message: blocker });
      return;
    }
    const selected = [...session.selected_unit_ids],
      epoch = this.epoch;
    this.publish({ actionPending: true });
    const refreshed = await this.refresh();
    if (!this.current(epoch)) return;
    this.publish({ actionPending: false });
    if (!refreshed || !this.canWrite()) return;
    if (
      JSON.stringify(selected) !==
      JSON.stringify(this.state.session?.selected_unit_ids)
    ) {
      this.publish({
        message: "Persisted selection changed. Review it before applying.",
      });
      return;
    }
    const currentBlocker = importApplyBlocker(this.state);
    if (currentBlocker) {
      this.publish({ message: currentBlocker });
      return;
    }
    try {
      await this.submit(
        captureImportWrite({
          kind: "apply",
          scope: binding.scope,
          sessionId: session.import_session_id,
          body: {
            client_txn_id: this.transactionId(),
            selected_unit_ids: selected,
          },
        }),
      );
    } catch {
      this.publish({
        message:
          "A secure apply request could not be created. No request was sent.",
      });
    }
  }
  canCancel() {
    const binding = this.binding,
      job = this.state.job;
    return (
      this.canRead() &&
      binding !== null &&
      job !== null &&
      this.state.jobCurrent &&
      (job.submitted_by_user_id === binding.scope.actorId ||
        binding.role === "admin") &&
      job.cancelable &&
      !terminalCommonJob(job) &&
      job.status !== "cancel_requested" &&
      this.state.cancellation?.phase !== "pending"
    );
  }
  async cancel() {
    const binding = this.binding,
      job = this.state.job;
    if (!binding || !job || !this.canCancel()) return;
    const previous = this.state.cancellation;
    let attempt: ImportWriteAttempt;
    try {
      attempt =
        previous?.phase === "uncertain"
          ? previous.attempt
          : captureImportWrite({
              kind: "cancel",
              scope: binding.scope,
              jobId: job.job_id,
              body: { client_txn_id: this.transactionId() },
            });
    } catch {
      this.publish({
        message:
          "A secure cancellation request could not be created. No request was sent.",
      });
      return;
    }
    const stop = this.controller(),
      epoch = this.epoch;
    this.stopObservation();
    this.loadStop?.abort();
    this.loadStop = null;
    this.previewStop?.abort();
    this.publish({
      loading: false,
      cancellation: { attempt, phase: "pending", failure: null },
    });
    const current = await boundedImportRead(
      (signal) =>
        binding.client.readJob(
          job.job_id,
          signal,
          this.state.jobPurpose === "apply"
            ? this.state.session?.import_session_id
            : undefined,
          this.state.jobPurpose ?? undefined,
        ),
      stop.signal,
      importTiming.request,
      this.clock,
    );
    if (!this.current(epoch)) {
      this.release(stop);
      return;
    }
    if (
      current.kind === "failed" ||
      !commonJobDoesNotRegress(job, current.value)
    ) {
      const failure =
        current.kind === "failed" ? current.failure : importContractFailure();
      this.publish({
        cancellation: {
          attempt,
          phase: previous?.phase === "uncertain" ? "uncertain" : "rejected",
          failure,
        },
        jobCurrent: false,
      });
      this.handleFailure(failure);
      this.release(stop);
      return;
    }
    this.publish({ job: current.value, jobCurrent: true });
    if (
      !current.value.cancelable ||
      terminalCommonJob(current.value) ||
      current.value.status === "cancel_requested"
    ) {
      this.publish({
        cancellation: null,
        message:
          current.value.status === "cancel_requested"
            ? "Cancellation requested. Committed units remain committed."
            : "The job no longer accepts cancellation. Review its current outcome.",
      });
      this.release(stop);
      void this.resumeObservation();
      return;
    }
    const result = await boundedImportRead<ImportWriteResult>(
      async (signal) => {
        if (!this.current(epoch))
          return { kind: "failed", failure: importInterruptedFailure() };
        return {
          kind: "received",
          value: await binding.client.send(
            attempt,
            signal,
            previous?.phase === "uncertain",
          ),
        };
      },
      stop.signal,
      importTiming.request,
      this.clock,
    );
    this.release(stop);
    if (!this.current(epoch)) return;
    const outcome =
      result.kind === "received"
        ? result.value
        : { kind: "uncertain" as const, failure: result.failure };
    if (
      outcome.kind === "accepted" &&
      outcome.receipt.kind === "job" &&
      commonJobDoesNotRegress(current.value, outcome.receipt.job)
    ) {
      const next = outcome.receipt.job;
      this.publish({
        cancellation: null,
        job: next,
        jobCurrent: true,
        message:
          next.status === "cancel_requested"
            ? "Cancellation requested. Committed units remain committed."
            : "Cancellation response received. Review the authoritative job and unit outcomes.",
      });
    } else {
      const failure =
        outcome.kind === "accepted" ? importContractFailure() : outcome.failure;
      this.publish({
        cancellation: {
          attempt,
          phase: outcome.kind === "accepted" ? "uncertain" : outcome.kind,
          failure,
        },
      });
      this.handleFailure(failure);
    }
    void this.resumeObservation();
  }
  async openResult(viewId: string, navigate: (viewId: string) => void) {
    const binding = this.binding,
      job = this.state.job;
    if (
      !binding ||
      !job ||
      !this.presented ||
      !this.canRead() ||
      this.state.actionPending ||
      !importOutcomeViews(this.state).some((v) => v.id === viewId)
    )
      return;
    const stop = this.controller(),
      epoch = this.epoch;
    this.navigationStop?.abort();
    this.navigationStop = stop;
    this.publish({ actionPending: true });
    const result = await boundedImportRead(
      (signal) =>
        binding.client.readJob(
          job.job_id,
          signal,
          this.state.session?.import_session_id,
          this.state.jobPurpose ?? undefined,
        ),
      stop.signal,
      importTiming.request,
      this.clock,
    );
    if (
      this.current(epoch) &&
      !stop.signal.aborted &&
      result.kind === "received" &&
      commonJobDoesNotRegress(job, result.value)
    ) {
      this.publish({ job: result.value, jobCurrent: true });
      if (
        (await this.refresh()) &&
        this.current(epoch) &&
        this.presented &&
        !stop.signal.aborted &&
        importOutcomeViews(this.state).some((v) => v.id === viewId)
      )
        navigate(viewId);
    } else if (this.current(epoch) && result.kind === "failed") {
      this.publish({ jobCurrent: false, observationFailure: result.failure });
      this.handleFailure(result.failure);
    }
    this.release(stop);
    if (this.current(epoch)) this.publish({ actionPending: false });
    if (this.navigationStop === stop) this.navigationStop = null;
  }
}
