import {
  boundedImportRead,
  browserImportClock,
  type ImportClock,
  importTiming,
  loadImportResources,
  observeImportJob,
  submitImportAttempt,
} from "../imports/importCoordinator";
import {
  captureImportWrite,
  captureWorkbookUpload,
  type ImportScope,
  type ImportWriteAttempt,
  immutableImportValue,
} from "../imports/importRequests";
import { clientTxnID } from "../services/browserApi";
import {
  commonJobDoesNotRegress,
  terminalCommonJob,
} from "../services/commonJobContract";
import {
  type ImportClient,
  type ImportFailure,
  type ImportWriteReceipt,
  importContractFailure,
  importFailureMessage,
  importInterruptedFailure,
  sameImportSessionSource,
  sameImportUnitSource,
} from "../services/importClient";
import { importSessionIdFromReceipt } from "../services/importJobContract";
import { requireClaimGatedAnalyticalImportTarget } from "../services/importTargetContractAdapter";
import {
  decodeNetworkFlowImportPreviewResult,
  networkFlowMappingCandidateSchemaId,
  networkFlowMappingMetadata,
} from "../services/networkFlowContractAdapter";
import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import {
  buildNetworkFlowMappingCandidate,
  createNetworkFlowMappingDraft,
  type NetworkFlowMappingDraft,
  networkFlowApprovalRequest,
  networkFlowApprovedPreviewMatches,
  networkFlowMappingDraftReadyForPreview,
  networkFlowMappingIssues,
} from "./networkFlowImportModel";
import {
  initialNetworkFlowImportState,
  type NetworkFlowImportAttempt,
  type NetworkFlowImportHandoffRequest,
  type NetworkFlowImportHandoffResult,
  type NetworkFlowImportState,
  unresolvedNetworkFlowWrite,
} from "./networkFlowImportState";

const target = requireClaimGatedAnalyticalImportTarget(
  networkFlowMappingMetadata.target_kind,
  networkFlowMappingMetadata.profile_id,
);
export type NetworkFlowImportPort = Pick<
  ImportClient,
  | "send"
  | "readJob"
  | "readSession"
  | "listUnits"
  | "readUnit"
  | "preview"
  | "previewMapping"
>;
export type NetworkFlowImportBinding = {
  readonly scope: ImportScope;
  readonly role: WorkbookIncidentRole | null;
  readonly closed: boolean;
  readonly available: boolean;
  readonly client: NetworkFlowImportPort;
  readonly current: () => boolean;
  readonly accessFailure: (failure: ImportFailure) => void;
};

/** One analytical workflow. Drafts, acknowledgements and publication never share a disposition. */
export class NetworkFlowImportController {
  private state = initialNetworkFlowImportState();
  private recoverySequence = 0;
  private visible = this.state;
  private binding: NetworkFlowImportBinding | null = null;
  private scope: ImportScope | null = null;
  private authority = 0;
  private candidateRevision = 0;
  private writeGeneration = 0;
  private handoffGeneration = 0;
  private busy = false;
  private workspaceActive = false;
  private listeners = new Set<() => void>();
  private stops = new Set<AbortController>();
  private previewStop: AbortController | null = null;
  private observationStop: AbortController | null = null;
  private handoffCallback:
    | ((
        request: NetworkFlowImportHandoffRequest,
      ) => Promise<NetworkFlowImportHandoffResult>)
    | null = null;
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
      options.transactionId ?? (() => clientTxnID("network-flow-import"));
  }
  getSnapshot = () => this.visible;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(change: Partial<NetworkFlowImportState>) {
    this.state = immutableImportValue({
      ...this.state,
      ...change,
      revision: this.state.revision + 1,
    });
    this.visible =
      this.state.access === "active"
        ? this.state
        : immutableImportValue({
            ...initialNetworkFlowImportState(),
            revision: this.state.revision,
            access: this.state.access,
            message:
              this.state.access === "paused"
                ? "Confirm current access to resume import review."
                : "Analytical import is unavailable.",
          });
    for (const listener of this.listeners) listener();
  }
  bind(binding: NetworkFlowImportBinding) {
    const changedScope =
      this.scope &&
      (this.scope.incidentId !== binding.scope.incidentId ||
        this.scope.actorId !== binding.scope.actorId ||
        this.scope.lifetime !== binding.scope.lifetime);
    if (changedScope) this.retire();
    const previous = this.binding;
    this.scope = { ...binding.scope };
    if (!binding.available || binding.role === "") {
      this.retire();
      return;
    }
    if (!binding.current() || binding.role === null) {
      this.pause();
      return;
    }
    const changed =
      !previous ||
      previous.closed !== binding.closed ||
      previous.role !== binding.role ||
      previous.client !== binding.client;
    if (changed) this.fence();
    this.binding = binding;
    this.publish({
      access: "active",
      canWrite: this.canWrite(),
      closed: binding.closed,
      workspaceActive: this.workspaceActive,
      ...(changed && this.state.draft
        ? {
            previewStale: true,
            message: binding.closed
              ? "Closed, read-only. Your mapping draft is retained for copying."
              : "Access changed. Review and preview the retained mapping before a fresh action.",
          }
        : {}),
    });
  }
  private canRead() {
    return (
      this.binding?.available === true &&
      this.binding.role !== null &&
      this.binding.role !== "" &&
      this.binding.current()
    );
  }
  private canWrite() {
    return (
      this.canRead() &&
      !this.binding?.closed &&
      ["editor", "reviewer", "admin"].includes(this.binding?.role ?? "")
    );
  }
  private current(authority: number) {
    return this.authority === authority && this.canRead();
  }
  private stop() {
    const stop = new AbortController();
    this.stops.add(stop);
    return stop;
  }
  private release(stop: AbortController) {
    this.stops.delete(stop);
  }
  private fence() {
    this.authority++;
    this.handoffGeneration++;
    this.candidateRevision++;
    for (const stop of this.stops) stop.abort();
    this.stops.clear();
    this.previewStop = null;
    this.observationStop = null;
    this.busy = false;
    const interrupted = (
      attempt: NetworkFlowImportAttempt | null,
    ): NetworkFlowImportAttempt | null =>
      attempt?.disposition === "pending"
        ? {
            ...attempt,
            disposition: "uncertain",
            failure: importInterruptedFailure(),
          }
        : attempt;
    this.state = immutableImportValue({
      ...this.state,
      write: interrupted(this.state.write),
      cancellation: interrupted(this.state.cancellation),
      cancellationChecking: false,
      previewStale: this.state.preview !== null || this.state.previewStale,
      discoveryJob: this.state.discoveryJob
        ? { ...this.state.discoveryJob, observing: false, current: false }
        : null,
      applyJob: this.state.applyJob
        ? { ...this.state.applyJob, observing: false, current: false }
        : null,
      ...(this.state.stage === "previewing" ? { stage: "mapping" } : {}),
      ...(this.state.handoff?.status === "loading"
        ? { handoff: { ...this.state.handoff, status: "pending" } }
        : {}),
    });
  }
  pause() {
    this.fence();
    this.binding = null;
    this.publish({ access: "paused", canWrite: false, presented: false });
  }
  observeClosure() {
    if (this.binding) this.bind({ ...this.binding, closed: true });
  }
  retire() {
    this.fence();
    this.binding = null;
    this.scope = null;
    this.state = initialNetworkFlowImportState();
    this.publish({});
  }
  setPresented(presented: boolean) {
    this.publish({ presented });
  }
  setWorkspaceActive(active: boolean) {
    this.workspaceActive = active;
    this.publish({ workspaceActive: active });
    if (!active) {
      this.handoffGeneration++;
      if (this.previewStop) {
        this.previewStop.abort();
        this.previewStop = null;
        this.candidateRevision++;
        this.busy = false;
        this.publish({ stage: "mapping", previewStale: true });
      }
      this.pauseObservation();
      this.publish({
        presented: false,
        ...(this.state.handoff?.status === "loading"
          ? { handoff: { ...this.state.handoff, status: "pending" } }
          : {}),
      });
    }
  }
  setHandoff(callback: typeof this.handoffCallback) {
    this.handoffGeneration++;
    this.handoffCallback = callback;
    if (this.state.handoff?.status === "loading")
      this.publish({ handoff: { ...this.state.handoff, status: "pending" } });
  }
  pauseObservation() {
    this.observationStop?.abort();
    this.observationStop = null;
    for (const key of ["discoveryJob", "applyJob"] as const) {
      const job = this.state[key];
      if (job?.observing)
        this.publish({
          [key]: {
            ...job,
            observing: false,
            failure: importInterruptedFailure(),
          },
        });
    }
  }
  canDiscardDraft() {
    if (
      this.busy ||
      unresolvedNetworkFlowWrite(this.state) ||
      this.state.cancellation?.disposition === "uncertain" ||
      this.state.cancellation?.disposition === "pending" ||
      this.state.applyJob ||
      this.state.approval ||
      (this.state.discoveryJob &&
        !terminalCommonJob(this.state.discoveryJob.resource))
    )
      return false;
    return true;
  }
  discardDraft() {
    if (!this.canDiscardDraft()) return;
    const binding = this.binding;
    this.retire();
    if (binding) this.bind(binding);
  }
  canStartNew() {
    const job = this.state.applyJob ?? this.state.discoveryJob;
    if (
      this.busy ||
      unresolvedNetworkFlowWrite(this.state) ||
      this.state.cancellation?.disposition === "uncertain" ||
      this.state.cancellation?.disposition === "pending" ||
      (job && !terminalCommonJob(job.resource)) ||
      (this.state.handoff && this.state.handoff.status !== "selected")
    )
      return false;
    return true;
  }
  startNew() {
    if (!this.canStartNew()) return;
    const binding = this.binding;
    this.retire();
    if (binding) this.bind(binding);
  }
  updateDraft(draft: NetworkFlowMappingDraft) {
    if (
      !this.state.canWrite ||
      !this.canWrite() ||
      !this.state.draft ||
      unresolvedNetworkFlowWrite(this.state) ||
      this.state.applyJob ||
      (this.busy && this.state.stage !== "previewing")
    )
      return;
    this.candidateRevision++;
    this.previewStop?.abort();
    this.previewStop = null;
    this.busy = false;
    this.publish({
      draft: immutableImportValue(JSON.parse(JSON.stringify(draft))),
      write:
        this.state.write?.disposition === "rejected" ? null : this.state.write,
      previewStale: true,
      previewFailure: null,
      stage: "mapping",
      message: "Mapping changed. Preview it before approval.",
    });
  }
  private candidateKey() {
    return this.state.draft && this.state.discovery
      ? JSON.stringify(
          buildNetworkFlowMappingCandidate(
            this.state.draft,
            this.state.discovery.preview.columns,
          ),
        )
      : null;
  }
  private applicablePreview() {
    try {
      return (
        this.state.preview !== null &&
        !this.state.previewStale &&
        this.state.preview.authority === this.authority &&
        this.state.preview.candidateKey === this.candidateKey()
      );
    } catch {
      return false;
    }
  }
  canContinue() {
    return (
      this.workspaceActive &&
      this.canWrite() &&
      this.state.canWrite &&
      !this.busy &&
      !this.state.sourceFailure &&
      !unresolvedNetworkFlowWrite(this.state) &&
      !this.state.applyJob &&
      this.state.stage === "mapping" &&
      this.state.draft !== null &&
      networkFlowMappingIssues(this.state.draft).length === 0 &&
      this.applicablePreview()
    );
  }
  canPreview() {
    return (
      this.workspaceActive &&
      this.canWrite() &&
      this.state.canWrite &&
      !this.busy &&
      !this.state.sourceFailure &&
      this.state.discovery !== null &&
      this.state.draft !== null &&
      !this.state.applyJob &&
      this.state.write?.disposition !== "pending" &&
      (this.state.write?.disposition !== "uncertain" ||
        this.state.previewStale) &&
      networkFlowMappingDraftReadyForPreview(this.state.draft)
    );
  }
  private failure(
    failure: ImportFailure,
    resource: "session" | "unit" | "job" | "preview" | "table",
  ) {
    if (failure.status === 401) {
      const binding = this.binding;
      this.pause();
      binding?.accessFailure(failure);
      return;
    }
    if (failure.code === "extension_profile_not_claimed") {
      this.retire();
      return;
    }
    if (failure.code === "incident_closed") {
      const binding = this.binding;
      this.observeClosure();
      binding?.accessFailure(failure);
      return;
    }
    if (failure.status === 403 || failure.kind === "authority") {
      this.fence();
      this.publish({ canWrite: false, message: importFailureMessage(failure) });
      this.binding?.accessFailure(failure);
      return;
    }
    if (failure.status === 404 && resource !== "job" && resource !== "table") {
      this.fence();
      this.publish({
        discovery: null,
        draft: null,
        preview: null,
        approval: null,
        selection: null,
        write: null,
        sourceFailure: failure,
        message: importFailureMessage(failure),
      });
      this.binding?.accessFailure(failure);
    }
  }
  async upload(file: File) {
    if (
      !this.binding ||
      !this.canWrite() ||
      !this.state.canWrite ||
      this.busy ||
      this.state.stage !== "idle"
    )
      return;
    this.publish({
      presented: true,
      stage: "uploading",
      recoveryId: ++this.recoverySequence,
      sourceFailure: null,
      message: "Uploading CSV; awaiting acknowledgement.",
    });
    const request = captureWorkbookUpload(
      this.binding.scope,
      file,
      this.transactionId(),
    );
    if (await this.submit(request)) await this.resumeObservation();
  }
  private async submit(request: ImportWriteAttempt, replay = false) {
    const binding = this.binding;
    if (
      !binding ||
      !this.canWrite() ||
      !this.state.canWrite ||
      this.busy ||
      (unresolvedNetworkFlowWrite(this.state) &&
        (!replay || this.state.write?.request !== request))
    )
      return false;
    this.busy = true;
    const authority = this.authority,
      generation = ++this.writeGeneration,
      stop = this.stop();
    this.publish({
      write: { request, disposition: "pending", failure: null, receipt: null },
    });
    const result = await submitImportAttempt({
      signal: stop.signal,
      clock: this.clock,
      upload: request.kind === "upload",
      send: (signal) => {
        if (!this.current(authority) || !this.canWrite())
          return Promise.resolve({
            kind: "uncertain",
            failure: importInterruptedFailure(),
          });
        const pending = binding.client.send(request, signal, replay);
        void pending.then(
          (outcome) => {
            if (
              outcome.kind === "accepted" &&
              this.current(authority) &&
              generation === this.writeGeneration &&
              this.state.write?.request === request &&
              this.state.write.disposition === "uncertain"
            ) {
              this.accept(request, outcome.receipt);
              if (outcome.receipt.kind === "job") void this.resumeObservation();
            }
          },
          () => {},
        );
        return pending;
      },
    });
    this.release(stop);
    if (!this.current(authority)) return false;
    this.busy = false;
    if (result.kind !== "accepted") {
      this.publish({
        write: {
          request,
          disposition: result.kind,
          failure: result.failure,
          receipt: null,
        },
        message: importFailureMessage(result.failure),
      });
      this.failure(result.failure, "unitId" in request ? "unit" : "session");
      return false;
    }
    return this.accept(request, result.receipt);
  }
  private accept(
    request: ImportWriteAttempt,
    receipt: ImportWriteReceipt,
  ): boolean {
    this.publish({
      write: { request, disposition: "accepted", receipt, failure: null },
    });
    if (
      receipt.kind === "job" &&
      (request.kind === "upload" || request.kind === "apply")
    ) {
      const key = request.kind === "upload" ? "discoveryJob" : "applyJob";
      this.publish({
        [key]: {
          attempt: request,
          resource: receipt.job,
          observing: false,
          current: false,
          failure: null,
        },
        stage: request.kind === "upload" ? "discovering" : "observing_apply",
        message:
          "Import job accepted. Server work continues independently of this dialog.",
      });
      return true;
    }
    if (
      receipt.kind === "unit" &&
      request.kind === "mapping" &&
      this.state.discovery &&
      this.state.preview
    ) {
      if (!sameImportUnitSource(this.state.discovery.unit, receipt.unit)) {
        this.publish({ sourceFailure: importContractFailure() });
        return false;
      }
      const fingerprint = this.state.preview.value.mapping_fingerprint;
      this.publish({
        approval: {
          attempt: request,
          unit: receipt.unit,
          candidateKey: this.state.preview.candidateKey,
          fingerprint: receipt.unit.mapping_fingerprint ?? "",
        },
        discovery: { ...this.state.discovery, unit: receipt.unit },
        stage: "mapping",
      });
      if (
        !this.applicablePreview() ||
        !networkFlowApprovedPreviewMatches(receipt.unit, fingerprint)
      ) {
        this.publish({
          previewStale: true,
          message:
            "The approved mapping no longer matches this preview. Review and preview it again.",
        });
        return false;
      }
      this.publish({
        message:
          "Mapping approval acknowledged. Selection has not yet been confirmed.",
      });
      return true;
    }
    if (
      receipt.kind === "selection" &&
      request.kind === "select" &&
      this.state.discovery
    ) {
      if (
        !sameImportUnitSource(
          this.state.discovery.unit,
          receipt.selection.unit,
        ) ||
        receipt.selection.import_session_id !== this.state.discovery.sessionId
      ) {
        this.publish({ sourceFailure: importContractFailure() });
        return false;
      }
      const matches =
        this.applicablePreview() &&
        receipt.selection.unit.mapping_fingerprint ===
          this.state.preview?.value.mapping_fingerprint;
      this.publish({
        selection: receipt.selection,
        discovery: { ...this.state.discovery, unit: receipt.selection.unit },
        stage: "mapping",
        previewStale: !matches,
        message: matches
          ? "Selection acknowledged. Apply has not yet been submitted."
          : "Selected approval differs from the reviewed preview. Review and preview it again.",
      });
      return matches;
    }
    this.publish({
      sourceFailure: importContractFailure(),
      message: "The import acknowledgement could not be correlated.",
    });
    return false;
  }
  async resumeObservation() {
    const binding = this.binding,
      key = this.state.applyJob ? "applyJob" : "discoveryJob",
      known = this.state[key];
    if (
      !binding ||
      !known ||
      !this.canRead() ||
      binding.closed ||
      !this.workspaceActive ||
      known.observing
    )
      return;
    const authority = this.authority,
      stop = this.stop();
    this.observationStop = stop;
    this.publish({ [key]: { ...known, observing: true, failure: null } });
    const purpose = key === "applyJob" ? "apply" : "discovery";
    const outcome = await observeImportJob({
      initial: known.resource,
      signal: stop.signal,
      clock: this.clock,
      read: (id, signal) =>
        binding.client.readJob(
          id,
          signal,
          this.state.discovery?.sessionId,
          purpose,
        ),
      onJob: (resource) => {
        if (this.current(authority) && this.observationStop === stop)
          this.publish({
            [key]: {
              ...known,
              resource,
              current: true,
              observing: true,
              failure: null,
            },
          });
      },
    });
    this.release(stop);
    if (!this.current(authority) || this.observationStop !== stop) return;
    this.observationStop = null;
    this.publish({
      [key]: {
        ...known,
        resource: outcome.job,
        current: outcome.kind === "terminal",
        observing: false,
        failure: outcome.kind === "paused" ? outcome.failure : null,
      },
    });
    if (outcome.kind === "paused") {
      this.publish({
        message:
          "Observation paused. Server work may continue; resume the known job.",
      });
      this.failure(outcome.failure, "job");
      return;
    }
    if (purpose === "discovery" && outcome.job.status === "succeeded")
      await this.refreshSource();
    else if (purpose === "discovery")
      this.publish({
        stage: "finished",
        message: `Discovery ${outcome.job.status}.`,
      });
    else if (outcome.job.status === "succeeded") {
      const refs =
        outcome.job.result_summary?.resource_refs?.filter(
          (ref) => ref.kind === target.target_kind,
        ) ?? [];
      const tableId = refs.length === 1 ? (refs[0]?.id ?? null) : null;
      if (this.state.handoff) {
        // Observation refresh does not repeat a completed selection or erase a handoff failure.
        if (this.state.handoff.status !== "selected")
          await this.recoverHandoff();
        return;
      }
      this.publish({
        stage: "handoff",
        handoff: {
          status: tableId ? "pending" : "unavailable",
          tableId,
          failure: tableId ? null : importContractFailure(),
        },
        message:
          "Import succeeded. Confirming the created table is a separate step.",
      });
      await this.recoverHandoff();
    } else
      this.publish({
        stage: "finished",
        message:
          outcome.job.status === "canceled"
            ? "Import canceled. Cancellation does not roll back committed work."
            : "Import failed. Review the job outcome before starting another import.",
      });
  }
  async refreshSource() {
    const binding = this.binding,
      job = this.state.discoveryJob;
    if (
      !binding ||
      !job ||
      !this.canRead() ||
      binding.closed ||
      this.busy ||
      unresolvedNetworkFlowWrite(this.state)
    )
      return;
    const sessionId = importSessionIdFromReceipt(job.resource);
    if (!sessionId) return;
    this.busy = true;
    const authority = this.authority,
      stop = this.stop();
    this.publish({ stage: "loading_source", sourceFailure: null });
    const resources = await loadImportResources(
      binding.client,
      sessionId,
      stop.signal,
      this.clock,
    );
    if (!this.current(authority)) {
      this.release(stop);
      return;
    }
    if (resources.kind === "failed") {
      this.busy = false;
      this.release(stop);
      this.publish({
        sourceFailure: resources.failure,
        message: importFailureMessage(resources.failure),
      });
      this.failure(resources.failure, "session");
      return;
    }
    const unit = resources.value.units[0],
      previous = this.state.discovery;
    if (
      resources.value.session.source_file_kind !== "csv" ||
      resources.value.units.length !== 1 ||
      !unit ||
      (previous &&
        (!sameImportSessionSource(previous.session, resources.value.session) ||
          !sameImportUnitSource(previous.unit, unit)))
    ) {
      this.busy = false;
      this.release(stop);
      this.publish({
        sourceFailure: importContractFailure(),
        message: "Discovery must identify one unchanged CSV source unit.",
      });
      return;
    }
    const preview = await boundedImportRead(
      (signal) => binding.client.preview(unit, signal),
      stop.signal,
      importTiming.request,
      this.clock,
    );
    this.release(stop);
    if (!this.current(authority)) return;
    this.busy = false;
    if (preview.kind === "failed") {
      this.publish({
        sourceFailure: preview.failure,
        message: importFailureMessage(preview.failure),
      });
      this.failure(preview.failure, "unit");
      return;
    }
    this.publish({
      discovery: {
        sessionId,
        session: resources.value.session,
        unit,
        preview: preview.value,
      },
      draft:
        this.state.draft ??
        createNetworkFlowMappingDraft(preview.value.columns),
      stage: "mapping",
      sourceFailure: null,
      message: "Review the discovered mapping before approval.",
    });
  }
  async requestPreview() {
    const binding = this.binding,
      discovery = this.state.discovery,
      draft = this.state.draft;
    if (
      !this.canPreview() ||
      !binding ||
      !this.canWrite() ||
      !this.state.canWrite ||
      !discovery ||
      !draft ||
      this.busy ||
      this.state.applyJob ||
      !networkFlowMappingDraftReadyForPreview(draft)
    )
      return;
    const candidate = extensionCandidate(
        buildNetworkFlowMappingCandidate(draft, discovery.preview.columns),
      ),
      key = this.candidateKey();
    if (!key) return;
    this.busy = true;
    const revision = ++this.candidateRevision,
      authority = this.authority,
      stop = this.stop();
    this.previewStop = stop;
    this.publish({
      stage: "previewing",
      previewStale: true,
      previewFailure: null,
      message: "Validating the current mapping preview.",
    });
    const result = await boundedImportRead(
      (signal) =>
        binding.client.previewMapping(discovery.unit, candidate, signal),
      stop.signal,
      importTiming.request,
      this.clock,
    );
    this.release(stop);
    if (!this.current(authority) || revision !== this.candidateRevision) return;
    this.busy = false;
    this.previewStop = null;
    let failure: ImportFailure | null =
      result.kind === "failed" ? result.failure : null;
    if (result.kind === "received") {
      try {
        const wrapper = result.value;
        if (
          wrapper.schema_id !==
            "cartulary.imports.extension_mapping_preview_result.v1" ||
          wrapper.import_session_id !== discovery.sessionId ||
          wrapper.import_unit_id !== discovery.unit.import_unit_id ||
          wrapper.target_kind !== target.target_kind ||
          wrapper.extension_profile_id !== target.extension_profile_id ||
          wrapper.owner_result_schema_id !==
            "cartulary.network_flow.import_preview_result.v1"
        )
          throw new Error("Invalid preview identity");
        const value = decodeNetworkFlowImportPreviewResult(
          wrapper.owner_result,
        );
        if (
          value.source_content_sha256 !==
            discovery.session.source_content_sha256 ||
          value.source_columns.length !== discovery.preview.columns.length ||
          value.source_columns.some(
            (column, i) =>
              column.source_column_ordinal !==
                discovery.preview.columns[i]?.source_column_ordinal ||
              column.raw_header_text !==
                String(discovery.preview.columns[i]?.source_header_text ?? ""),
          )
        )
          throw new Error("Changed preview source");
        this.publish({
          preview: { candidateKey: key, authority, value },
          previewStale: false,
          stage: "mapping",
          message: "Preview slice ready. Review it before explicit approval.",
        });
        return;
      } catch {
        failure = importContractFailure();
      }
    }
    if (failure) {
      this.publish({
        previewFailure: failure,
        stage: "mapping",
        message: importFailureMessage(failure),
      });
      this.failure(failure, "preview");
    }
  }
  async approve() {
    if (
      !this.canContinue() ||
      !this.binding ||
      !this.state.discovery ||
      !this.state.draft
    )
      return false;
    const discovery = this.state.discovery;
    const candidate = extensionCandidate(
      buildNetworkFlowMappingCandidate(
        this.state.draft,
        discovery.preview.columns,
      ),
    );
    this.publish({ stage: "approving" });
    return this.submit(
      captureImportWrite({
        kind: "mapping",
        scope: this.binding.scope,
        sessionId: discovery.sessionId,
        unitId: discovery.unit.import_unit_id,
        body: networkFlowApprovalRequest(
          discovery,
          candidate,
          this.transactionId(),
        ),
      }),
    );
  }
  async continueApply() {
    if (!this.canContinue()) return;
    const approval = this.state.approval;
    if (
      !approval ||
      approval.candidateKey !== this.candidateKey() ||
      approval.fingerprint !== this.state.preview?.value.mapping_fingerprint
    ) {
      if (!(await this.approve())) return;
    }
    if (
      !this.canContinue() ||
      !this.binding ||
      !this.state.discovery ||
      !this.state.approval
    )
      return;
    const discovery = this.state.discovery,
      preview = this.state.preview;
    if (
      !preview ||
      !networkFlowApprovedPreviewMatches(
        this.state.approval.unit,
        preview.value.mapping_fingerprint,
      )
    ) {
      this.publish({
        previewStale: true,
        message: "Review and preview the approved mapping again.",
      });
      return;
    }
    const selection = this.state.selection;
    if (
      !selection?.selected_unit_ids.includes(discovery.unit.import_unit_id) ||
      selection.unit.mapping_fingerprint !== preview.value.mapping_fingerprint
    ) {
      this.publish({ stage: "selecting" });
      if (
        !(await this.submit(
          captureImportWrite({
            kind: "select",
            scope: this.binding.scope,
            sessionId: discovery.sessionId,
            unitId: discovery.unit.import_unit_id,
            body: { client_txn_id: this.transactionId() },
          }),
        ))
      )
        return;
    }
    if (!this.canContinue() || !this.binding) return;
    const persisted = this.state.selection;
    if (
      !persisted ||
      persisted.session_status !== "ready_to_apply" ||
      persisted.selected_unit_ids.length !== 1 ||
      persisted.selected_unit_ids[0] !== discovery.unit.import_unit_id ||
      persisted.unit.unit_status !== "ready" ||
      !networkFlowApprovedPreviewMatches(
        persisted.unit,
        preview.value.mapping_fingerprint,
      )
    ) {
      this.publish({
        previewStale: true,
        message:
          "Persisted selection or approval changed. Refresh the source and review its mapping.",
      });
      return;
    }
    this.publish({ stage: "submitting_apply" });
    if (
      await this.submit(
        captureImportWrite({
          kind: "apply",
          scope: this.binding.scope,
          sessionId: discovery.sessionId,
          body: {
            client_txn_id: this.transactionId(),
            selected_unit_ids: [...persisted.selected_unit_ids],
          },
        }),
      )
    )
      await this.resumeObservation();
  }
  async retryWrite() {
    const operation = this.state.write;
    if (
      !this.workspaceActive ||
      !operation ||
      !["uncertain", "rejected"].includes(operation.disposition) ||
      !this.canWrite() ||
      !this.state.canWrite ||
      this.busy
    )
      return;
    const replay = operation.disposition === "uncertain";
    const prior = operation.request;
    if (prior.kind !== "upload" && !this.applicablePreview()) {
      this.publish({
        message:
          "Preview the retained submitted mapping under current access before recovering this request.",
      });
      return;
    }
    if (
      (prior.kind === "select" || prior.kind === "apply") &&
      this.state.approval?.fingerprint !==
        this.state.preview?.value.mapping_fingerprint
    ) {
      this.publish({
        previewStale: true,
        message:
          "The acknowledged approval differs from the current preview. Review the retained request; selection and apply remain blocked.",
      });
      return;
    }
    const request = replay ? prior : renewAttempt(prior, this.transactionId());
    if (await this.submit(request, replay)) {
      if (request.kind === "upload" || request.kind === "apply")
        await this.resumeObservation();
      else await this.continueApply();
    }
  }
  async recoverHandoff() {
    const handoff = this.state.handoff,
      callback = this.handoffCallback,
      discovery = this.state.discovery,
      approval = this.state.approval,
      job = this.state.applyJob;
    if (
      !handoff?.tableId ||
      !callback ||
      !discovery ||
      !approval ||
      !job?.current ||
      job.resource.status !== "succeeded" ||
      !this.workspaceActive ||
      !this.canRead() ||
      this.binding?.closed ||
      handoff.status === "selected" ||
      handoff.status === "loading"
    )
      return;
    const authority = this.authority,
      generation = ++this.handoffGeneration,
      stop = this.stop();
    const current = () =>
      this.current(authority) &&
      this.workspaceActive &&
      this.handoffGeneration === generation &&
      this.state.handoff?.status === "loading" &&
      this.state.applyJob?.resource.job_id === job.resource.job_id;
    this.publish({ handoff: { ...handoff, status: "loading", failure: null } });
    const result = await boundedImportRead(
      async (signal) =>
        receivedHandoff(
          await callback({
            incidentId: discovery.session.incident_id,
            sessionId: discovery.sessionId,
            unitId: discovery.unit.import_unit_id,
            sourceHash: discovery.session.source_content_sha256,
            fingerprint: approval.fingerprint,
            tableId: handoff.tableId as string,
            current,
            signal,
          }),
        ),
      stop.signal,
      importTiming.request,
      this.clock,
    );
    this.release(stop);
    if (!current()) return;
    const outcome =
      result.kind === "received"
        ? result.value
        : { kind: "failed" as const, failure: result.failure };
    if (outcome.kind === "superseded") {
      this.publish({ handoff: { ...handoff, status: "pending" } });
      return;
    }
    this.publish({
      handoff: {
        ...handoff,
        status: outcome.kind,
        failure: outcome.kind === "failed" ? outcome.failure : null,
      },
      stage: outcome.kind === "selected" ? "finished" : "handoff",
      presented: outcome.kind === "selected" ? false : this.state.presented,
      message:
        outcome.kind === "selected"
          ? "Import succeeded. The created table is selected."
          : "Import succeeded, but the created table could not be opened. Recover table handoff without importing again.",
    });
    // Table denial is not proof of incident access loss. Authentication and
    // lifecycle failures still require the same authoritative recovery boundary.
    if (outcome.kind === "failed") this.failure(outcome.failure, "table");
  }
  canCancel() {
    const job = (this.state.applyJob ?? this.state.discoveryJob)?.resource;
    return Boolean(
      this.canRead() &&
        !this.binding?.closed &&
        job &&
        !terminalCommonJob(job) &&
        job.cancelable &&
        (job.submitted_by_user_id === this.scope?.actorId ||
          this.binding?.role === "admin") &&
        !this.state.cancellationChecking &&
        this.state.cancellation?.disposition !== "pending",
    );
  }
  async cancel() {
    const binding = this.binding,
      key = this.state.applyJob ? "applyJob" : "discoveryJob",
      known = this.state[key];
    const replay = this.state.cancellation?.disposition === "uncertain";
    if (
      !binding ||
      !known ||
      !this.canRead() ||
      binding.closed ||
      this.state.cancellationChecking ||
      this.state.cancellation?.disposition === "pending" ||
      (!replay && !this.canCancel())
    )
      return;
    const authority = this.authority,
      stop = this.stop();
    this.pauseObservation();
    this.publish({ cancellationChecking: true });
    const read = await boundedImportRead(
      (signal) =>
        binding.client.readJob(
          known.resource.job_id,
          signal,
          this.state.discovery?.sessionId,
          key === "applyJob" ? "apply" : "discovery",
        ),
      stop.signal,
      importTiming.request,
      this.clock,
    );
    if (!this.current(authority)) {
      this.release(stop);
      return;
    }
    this.publish({ cancellationChecking: false });
    if (read.kind === "failed") {
      this.release(stop);
      this.publish({
        [key]: {
          ...known,
          failure: read.failure,
          current: false,
          observing: false,
        },
        message: importFailureMessage(read.failure),
      });
      this.failure(read.failure, "job");
      return;
    }
    if (!commonJobDoesNotRegress(known.resource, read.value)) {
      this.release(stop);
      this.publish({
        [key]: { ...known, observing: false, failure: importContractFailure() },
      });
      return;
    }
    this.publish({
      [key]: {
        ...known,
        resource: read.value,
        current: true,
        observing: false,
        failure: null,
      },
    });
    if (!replay && (!read.value.cancelable || terminalCommonJob(read.value))) {
      this.release(stop);
      await this.resumeObservation();
      return;
    }
    if (
      read.value.submitted_by_user_id !== binding.scope.actorId &&
      binding.role !== "admin"
    ) {
      this.release(stop);
      return;
    }
    const request =
      replay && this.state.cancellation
        ? this.state.cancellation.request
        : captureImportWrite({
            kind: "cancel",
            scope: binding.scope,
            jobId: read.value.job_id,
            body: { client_txn_id: this.transactionId() },
          });
    this.publish({
      cancellation: {
        request,
        disposition: "pending",
        receipt: null,
        failure: null,
      },
    });
    const acceptCancellation = (receipt: ImportWriteReceipt) => {
      if (
        receipt.kind !== "job" ||
        !commonJobDoesNotRegress(read.value, receipt.job)
      ) {
        this.publish({
          cancellation: {
            request,
            disposition: "uncertain",
            receipt: null,
            failure: importContractFailure(),
          },
        });
        return false;
      }
      const latest = this.state[key]?.resource ?? read.value;
      const resource = commonJobDoesNotRegress(latest, receipt.job)
        ? receipt.job
        : latest;
      if (!commonJobDoesNotRegress(receipt.job, resource)) {
        this.publish({
          cancellation: {
            request,
            disposition: "uncertain",
            receipt: null,
            failure: importContractFailure(),
          },
        });
        return false;
      }
      this.publish({
        cancellation: {
          request,
          disposition: "accepted",
          receipt,
          failure: null,
        },
        [key]: {
          ...known,
          resource,
          current: true,
          observing: false,
          failure: null,
        },
        message:
          "Cancellation acknowledged. This does not establish rollback or a terminal outcome.",
      });
      return true;
    };
    const result = await submitImportAttempt({
      signal: stop.signal,
      clock: this.clock,
      send: (signal) => {
        if (!this.current(authority))
          return Promise.resolve({
            kind: "uncertain",
            failure: importInterruptedFailure(),
          });
        const pending = binding.client.send(request, signal, replay);
        void pending.then(
          (outcome) => {
            if (
              outcome.kind === "accepted" &&
              this.current(authority) &&
              this.state.cancellation?.request === request &&
              this.state.cancellation.disposition === "uncertain" &&
              acceptCancellation(outcome.receipt)
            )
              void this.resumeObservation();
          },
          () => {},
        );
        return pending;
      },
    });
    this.release(stop);
    if (!this.current(authority)) return;
    if (result.kind !== "accepted") {
      this.publish({
        cancellation: {
          request,
          disposition: result.kind,
          receipt: null,
          failure: result.failure,
        },
        message: importFailureMessage(result.failure),
      });
      this.failure(result.failure, "job");
      if (result.failure.code === "job_cancel_rejected")
        await this.resumeObservation();
      return;
    }
    if (acceptCancellation(result.receipt)) await this.resumeObservation();
  }
}
function renewAttempt(
  attempt: ImportWriteAttempt,
  transactionId: string,
): ImportWriteAttempt {
  // The discriminant and captured intent are unchanged; only a definitely rejected ID is retired.
  return immutableImportValue(
    attempt.kind === "upload"
      ? {
          ...attempt,
          metadata: { ...attempt.metadata, client_txn_id: transactionId },
        }
      : ({
          ...attempt,
          body: { ...attempt.body, client_txn_id: transactionId },
        } as ImportWriteAttempt),
  );
}
function receivedHandoff(value: NetworkFlowImportHandoffResult) {
  return { kind: "received" as const, value };
}
function extensionCandidate(
  candidate: ReturnType<typeof buildNetworkFlowMappingCandidate>,
) {
  return {
    target_kind: target.target_kind,
    extension_profile_id: target.extension_profile_id,
    owner_mapping_schema_id: networkFlowMappingCandidateSchemaId,
    owner_mapping: { ...candidate },
  } as const;
}
