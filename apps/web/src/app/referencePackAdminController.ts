import { clientTxnID } from "../services/browserApi";
import {
  cancelReferencePackJob,
  captureReferencePackAttempt,
  listReferencePacks,
  loadReferencePackJob,
  type ReferencePackCommand,
  type ReferencePackJobResource,
  type ReferencePackQuery,
  readReferencePackVersion,
  referencePackContractProblem,
  referencePackTransportProblem,
  submitReferencePackAttempt,
} from "../services/referencePacks";
import {
  acceptsReferencePackJob,
  catalogAccepted,
  catalogFailed,
  catalogInvalidated,
  catalogStarted,
  defaultReferencePackPaging,
  initialReferencePackState,
  jobObserved,
  normalizeReferencePackQuery,
  type ReferencePackAdminState,
  type ReferencePackAuthority,
  type ReferencePackKnownJob,
  referencePackBusy,
  referencePackCommandLabel,
  referencePackEligible,
  sameReferencePackQuery,
  selectionChanged,
  terminalReferencePackJobStates,
} from "./referencePackAdminModel";

export type ReferencePackAccess = {
  readonly kind:
    | "authorized"
    | "access_lost"
    | "session_lost"
    | "cancelled"
    | "unavailable";
};
export type ReferencePackAdminPorts = {
  readonly list: typeof listReferencePacks;
  readonly version: typeof readReferencePackVersion;
  readonly submit: typeof submitReferencePackAttempt;
  readonly job: typeof loadReferencePackJob;
  readonly cancel: typeof cancelReferencePackJob;
  readonly confirmAccess: (
    authority: ReferencePackAuthority,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<ReferencePackAccess>;
  readonly isCurrent: (authority: ReferencePackAuthority) => boolean;
  readonly authorizationFailed: (status: 401 | 403) => void;
  readonly transactionId: (prefix: string) => string;
};
export const referencePackTiming = {
  read: 30_000,
  admission: 120_000,
  poll: 1_000,
} as const;
export const referencePackTransportPorts = {
  list: listReferencePacks,
  version: readReferencePackVersion,
  submit: submitReferencePackAttempt,
  job: loadReferencePackJob,
  cancel: cancelReferencePackJob,
  transactionId: clientTxnID,
};

/** Reference Pack execution and continuation fencing; no React or application-session ownership. */
export class ReferencePackAdminController {
  private state = initialReferencePackState();
  private readonly listeners = new Set<() => void>();
  private disposed = false;
  private epoch = 0;
  private sequence = 0;
  private activation = 0;
  private catalogAbort: AbortController | null = null;
  private jobAbort: AbortController | null = null;
  private jobRead: Promise<boolean> | null = null;
  private pollTimer: ReturnType<typeof setTimeout> | null = null;
  private readonly requests = new Set<AbortController>();
  private authorityInput: ReferencePackAuthority | null = null;
  constructor(private readonly ports: ReferencePackAdminPorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
  private publish(next: ReferencePackAdminState) {
    if (this.disposed || next === this.state) return;
    this.state = next;
    for (const listener of this.listeners) listener();
  }
  private announce(text: string) {
    this.publish({
      ...this.state,
      announcement: { serial: this.state.announcement.serial + 1, text },
    });
  }
  private current(authority: ReferencePackAuthority, epoch = this.epoch) {
    return (
      !this.disposed &&
      epoch === this.epoch &&
      this.state.authority?.lifetime === authority.lifetime &&
      this.state.authority.actorId === authority.actorId &&
      this.ports.isCurrent(authority)
    );
  }
  setAuthority(authority: ReferencePackAuthority | null) {
    if (this.disposed) return;
    this.authorityInput = authority;
    if (
      authority?.lifetime === this.state.authority?.lifetime &&
      authority?.actorId === this.state.authority?.actorId
    )
      return;
    this.retire();
    if (authority) this.publish({ ...this.state, authority, access: "ready" });
    if (authority && this.state.active) void this.reconcile();
  }
  setActive(active: boolean) {
    if (this.disposed || active === this.state.active) return;
    ++this.activation;
    ++this.sequence;
    this.stopPolling();
    this.catalogAbort?.abort();
    this.jobAbort?.abort();
    this.publish({
      ...this.state,
      active,
      operation:
        this.state.operation?.phase === "checking"
          ? {
              ...this.state.operation,
              phase: "rejected",
              attempt: null,
              problem: referencePackTransportProblem,
            }
          : this.state.operation,
      jobs: Object.fromEntries(
        Object.entries(this.state.jobs).map(([id, job]) => [
          id,
          {
            ...job,
            observation:
              job.observation === "reading" ? "paused" : job.observation,
            cancellation:
              job.cancellation?.phase === "checking"
                ? {
                    ...job.cancellation,
                    phase: "rejected",
                    problem: referencePackTransportProblem,
                  }
                : job.cancellation,
          },
        ]),
      ),
      catalog: {
        ...this.state.catalog,
        pending: null,
        dirty: true,
        paging: defaultReferencePackPaging,
      },
    });
    if (active) void this.reconcile();
  }
  retire() {
    ++this.epoch;
    ++this.sequence;
    ++this.activation;
    this.stopPolling();
    for (const request of this.requests) request.abort();
    this.requests.clear();
    this.catalogAbort = null;
    this.jobAbort = null;
    this.jobRead = null;
    this.publish({ ...initialReferencePackState(), active: this.state.active });
  }
  dispose() {
    this.retire();
    this.disposed = true;
    this.authorityInput = null;
    this.listeners.clear();
  }
  private async bounded<T>(
    work: (signal: AbortSignal) => Promise<T>,
    timeout: number,
    controller = new AbortController(),
  ): Promise<{ kind: "value"; value: T } | { kind: "interrupted" }> {
    this.requests.add(controller);
    let timer: ReturnType<typeof setTimeout> | undefined;
    let abort = () => {};
    const stopped = new Promise<{ kind: "interrupted" }>((resolve) => {
      abort = () => resolve({ kind: "interrupted" });
      controller.signal.addEventListener("abort", abort, { once: true });
      timer = setTimeout(() => controller.abort(), Math.max(0, timeout));
    });
    try {
      if (controller.signal.aborted) return { kind: "interrupted" };
      return await Promise.race([
        Promise.resolve()
          .then(() =>
            controller.signal.aborted
              ? Promise.reject(new Error("Retired request"))
              : work(controller.signal),
          )
          .then(
            (value) => ({ kind: "value", value }) as const,
            () => ({ kind: "interrupted" }) as const,
          ),
        stopped,
      ]);
    } finally {
      clearTimeout(timer);
      controller.signal.removeEventListener("abort", abort);
      this.requests.delete(controller);
    }
  }
  retryAccess = () => this.reconcile();
  private async reconcile(refresh = true) {
    const authority = this.state.authority ?? this.authorityInput;
    if (
      this.disposed ||
      !authority ||
      !this.state.active ||
      !this.ports.isCurrent(authority)
    )
      return;
    if (!this.state.authority) this.publish({ ...this.state, authority });
    const epoch = this.epoch;
    const activation = ++this.activation;
    this.stopPolling();
    this.publish({ ...this.state, access: "checking", reconciling: true });
    const current = () =>
      this.current(authority, epoch) &&
      activation === this.activation &&
      this.state.active;
    const observed = await this.bounded(
      (signal) => this.ports.confirmAccess(authority, signal, current),
      referencePackTiming.read,
    );
    if (!current()) return;
    if (
      observed.kind !== "value" ||
      observed.value.kind === "unavailable" ||
      observed.value.kind === "cancelled"
    ) {
      this.publish({
        ...this.state,
        access: "unavailable",
        reconciling: false,
      });
      return;
    }
    if (observed.value.kind !== "authorized") {
      this.authorizationLost(
        observed.value.kind === "session_lost" ? 401 : 403,
      );
      return;
    }
    this.publish({ ...this.state, access: "ready" });
    if (refresh) {
      await this.reload();
      for (const id of Object.keys(this.state.jobs)) {
        if (!current()) return;
        await this.observeJob(id);
      }
    }
    if (current()) {
      this.publish({ ...this.state, reconciling: false });
      this.schedulePoll();
    }
  }
  private authorizationLost(status: 401 | 403) {
    this.retire();
    this.ports.authorizationFailed(status);
  }
  setQuery(input: ReferencePackQuery) {
    if (!this.state.authority) return;
    const query = normalizeReferencePackQuery(input);
    const changed = !sameReferencePackQuery(query, this.state.query);
    this.publish({ ...this.state, input: { ...input }, query });
    if (changed) void this.reload();
  }
  setSelected(key: string, selected: boolean) {
    if (this.state.authority)
      this.publish(selectionChanged(this.state, key, selected));
  }
  clearSelection() {
    this.publish({ ...this.state, selectedKeys: [] });
  }
  setFile(file: File | null) {
    if (this.state.authority && !referencePackBusy(this.state))
      this.publish({ ...this.state, file });
  }
  setScrollTop(scrollTop: number) {
    if (scrollTop !== this.state.scrollTop)
      this.publish({ ...this.state, scrollTop });
  }
  reload = () => this.readCatalog(false);
  loadMore = () => this.readCatalog(true);
  invalidateCatalog() {
    this.publish(catalogInvalidated(this.state));
    if (
      !this.state.catalog.pending &&
      !this.state.catalog.problem &&
      this.state.active &&
      this.state.access === "ready"
    )
      void this.reload();
  }
  private async readCatalog(append: boolean) {
    const authority = this.state.authority;
    if (
      !authority ||
      !this.current(authority) ||
      !this.state.active ||
      this.state.access !== "ready"
    )
      return;
    const catalog = this.state.catalog;
    if (
      append &&
      (catalog.pending ||
        catalog.dirty ||
        !catalog.paging.has_more ||
        !catalog.paging.next_cursor ||
        !catalog.acceptedQuery ||
        !sameReferencePackQuery(catalog.acceptedQuery, this.state.query))
    )
      return;
    this.catalogAbort?.abort();
    const controller = new AbortController();
    this.catalogAbort = controller;
    const epoch = this.epoch;
    const sequence = ++this.sequence;
    const query = { ...this.state.query };
    const cursorToken = append ? catalog.paging.next_cursor : null;
    this.publish(catalogStarted(this.state, sequence, append));
    const outcome = await this.bounded(
      (signal) => this.ports.list({ query, cursorToken, signal }),
      referencePackTiming.read,
      controller,
    );
    if (
      !this.current(authority, epoch) ||
      sequence !== this.sequence ||
      !this.state.active
    )
      return;
    this.catalogAbort = null;
    if (outcome.kind !== "value") {
      this.publish(
        catalogFailed(this.state, sequence, referencePackTransportProblem),
      );
      return;
    }
    const result = outcome.value;
    if (result.kind === "access_failed") {
      this.authorizationLost(result.status);
      return;
    }
    if (result.kind === "failed") {
      this.publish(catalogFailed(this.state, sequence, result.problem));
      return;
    }
    if (!result.value.meta.paging) {
      this.publish(
        catalogFailed(this.state, sequence, referencePackContractProblem),
      );
      return;
    }
    this.publish(
      catalogAccepted(
        this.state,
        sequence,
        result.value.data.pack_versions,
        result.value.meta.paging,
      ),
    );
    if (this.state.catalog.dirty) void this.reload();
  }
  private canCommand() {
    return (
      this.state.authority !== null &&
      this.current(this.state.authority) &&
      this.state.active &&
      this.state.access === "ready" &&
      !this.state.reconciling
    );
  }
  async run(command: ReferencePackCommand) {
    if (!this.canCommand() || referencePackBusy(this.state)) return;
    if (command.kind === "import" && !this.state.file) {
      this.announce("Select a reference pack bundle first.");
      return;
    }
    if (command.kind === "refresh_selected" && !command.packKeys.length) return;
    const attempt = captureReferencePackAttempt(
      command,
      this.ports.transactionId(`reference-pack-${command.kind}`),
      command.kind === "import" ? (this.state.file ?? undefined) : undefined,
    );
    this.publish({
      ...this.state,
      operation: {
        id: attempt.request.client_txn_id,
        command: attempt.command,
        attempt,
        phase: "checking",
        jobId: null,
        problem: null,
      },
    });
    await this.executeAttempt(false);
  }
  retryAttempt = async () => {
    if (this.canCommand() && this.state.operation?.phase === "uncertain")
      await this.executeAttempt(true);
  };
  private async executeAttempt(replay: boolean) {
    const operation = this.state.operation;
    const authority = this.state.authority;
    if (!operation?.attempt || !authority || !this.current(authority)) return;
    const epoch = this.epoch;
    const activation = this.activation;
    const attempt = operation.attempt;
    const current = () =>
      this.current(authority, epoch) &&
      this.state.operation?.id === operation.id;
    this.publish({
      ...this.state,
      operation: {
        ...operation,
        phase: replay ? "replaying" : "checking",
        problem: null,
      },
    });
    if (!replay && "target" in operation.command) {
      const command = operation.command;
      const observed = await this.bounded(
        (signal) => this.ports.version(command.target, signal),
        referencePackTiming.read,
      );
      if (!current() || activation !== this.activation || !this.state.active)
        return;
      if (
        observed.kind === "value" &&
        observed.value.kind === "access_failed"
      ) {
        this.authorizationLost(observed.value.status);
        return;
      }
      const problem =
        observed.kind !== "value"
          ? referencePackTransportProblem
          : observed.value.kind === "failed"
            ? observed.value.problem
            : observed.value.kind === "read" &&
                referencePackEligible(observed.value.value, command.kind)
              ? null
              : ({
                  kind: "state_conflict",
                  status: 409,
                  code: "reference_pack_state_conflict",
                } as const);
      if (problem) {
        this.publish({
          ...this.state,
          operation: {
            ...operation,
            phase: "rejected",
            problem,
            attempt: null,
          },
        });
        this.announce(
          `${referencePackCommandLabel(command)} was not submitted. Review current state.`,
        );
        return;
      }
    }
    if (
      !current() ||
      (!replay && (activation !== this.activation || !this.state.active))
    )
      return;
    this.publish({
      ...this.state,
      operation: {
        ...operation,
        phase: replay ? "replaying" : "submitting",
        problem: null,
      },
    });
    this.announce(
      `${referencePackCommandLabel(operation.command)} submitted. Awaiting server acknowledgment.`,
    );
    const outcome = await this.bounded(
      (signal) => this.ports.submit(attempt, signal, replay),
      referencePackTiming.admission,
    );
    if (!current()) return;
    const result =
      outcome.kind === "value"
        ? outcome.value
        : ({
            kind: "uncertain",
            problem: referencePackTransportProblem,
          } as const);
    if (result.kind === "access_failed") {
      this.authorizationLost(result.status);
      return;
    }
    if (result.kind === "committed") {
      this.publish({
        ...this.state,
        operation: {
          ...operation,
          phase: "committed",
          attempt: null,
          problem: null,
        },
      });
      this.announce(
        `${referencePackCommandLabel(operation.command)} committed.`,
      );
      this.invalidateCatalog();
      return;
    }
    if (result.kind === "accepted") {
      const existing = this.state.jobs[result.job.job_id];
      if (
        result.job.submitted_by_user_id !== authority.actorId ||
        (existing &&
          (existing.operationId !== operation.id ||
            !acceptsReferencePackJob(existing.snapshot, result.job)))
      ) {
        this.publish({
          ...this.state,
          operation: {
            ...operation,
            phase: "uncertain",
            problem: referencePackContractProblem,
          },
        });
        return;
      }
      this.publish({
        ...this.state,
        file: operation.command.kind === "import" ? null : this.state.file,
        operation: {
          ...operation,
          phase: "accepted",
          jobId: result.job.job_id,
          attempt: null,
          problem: null,
        },
        jobs: {
          ...this.state.jobs,
          [result.job.job_id]: {
            operationId: operation.id,
            command: operation.command,
            snapshot: result.job,
            observation: "idle",
            problem: null,
            cancellation: null,
          },
        },
      });
      this.announce(
        `${referencePackCommandLabel(operation.command)} accepted. Operation ${result.job.status}.`,
      );
      if (this.state.active && this.state.access === "ready")
        void this.observeJob(result.job.job_id);
      return;
    }
    this.publish({
      ...this.state,
      operation: {
        ...operation,
        phase: result.kind,
        problem: result.problem,
        attempt: result.kind === "uncertain" ? attempt : null,
      },
    });
    this.announce(
      `${referencePackCommandLabel(operation.command)} ${result.kind === "uncertain" ? "has an uncertain outcome. Exact request recovery is available." : "was rejected."}`,
    );
  }
  private updateJob(
    id: string,
    update: (job: ReferencePackKnownJob) => ReferencePackKnownJob,
  ) {
    const job = this.state.jobs[id];
    if (job)
      this.publish({
        ...this.state,
        jobs: { ...this.state.jobs, [id]: update(job) },
      });
  }
  private stopPolling() {
    if (this.pollTimer !== null) clearTimeout(this.pollTimer);
    this.pollTimer = null;
  }
  private schedulePoll() {
    if (this.pollTimer !== null || this.jobRead !== null || !this.canCommand())
      return;
    const next = Object.entries(this.state.jobs).find(
      ([, job]) =>
        (!terminalReferencePackJobStates.has(job.snapshot.status) ||
          (this.state.operation?.phase === "accepted" &&
            this.state.operation.jobId === job.snapshot.job_id)) &&
        (job.observation === "idle" || job.observation === "paused") &&
        !["checking", "submitting"].includes(job.cancellation?.phase ?? ""),
    );
    if (!next) return;
    this.pollTimer = setTimeout(() => {
      this.pollTimer = null;
      void this.observeJob(next[0]);
    }, referencePackTiming.poll);
  }
  retryObservation = async (id: string) => {
    if (!this.canCommand() || this.state.jobs[id]?.observation === "reading")
      return;
    const epoch = this.epoch;
    if (this.jobRead !== null) await this.jobRead;
    if (epoch !== this.epoch || !this.canCommand()) return;
    await this.observeJob(id);
  };
  private async observeJob(
    id: string,
    timeout: number = referencePackTiming.read,
  ): Promise<boolean> {
    if (this.jobRead !== null) return this.jobRead;
    const authority = this.state.authority;
    const job = this.state.jobs[id];
    if (
      !authority ||
      !job ||
      !this.current(authority) ||
      !this.state.active ||
      this.state.access !== "ready"
    )
      return false;
    const epoch = this.epoch;
    const activation = this.activation;
    const controller = new AbortController();
    this.jobAbort = controller;
    this.stopPolling();
    this.updateJob(id, (value) => ({ ...value, observation: "reading" }));
    const read = (async () => {
      const outcome = await this.bounded(
        (signal) => this.ports.job(id, signal),
        timeout,
        controller,
      );
      if (
        !this.current(authority, epoch) ||
        activation !== this.activation ||
        !this.state.active ||
        this.state.jobs[id]?.operationId !== job.operationId
      )
        return false;
      const result =
        outcome.kind === "value"
          ? outcome.value
          : ({
              kind: "failed",
              problem: referencePackTransportProblem,
            } as const);
      if (result.kind === "access_failed") {
        this.authorizationLost(result.status);
        return false;
      }
      if (result.kind === "failed") {
        this.updateJob(id, (value) => ({
          ...value,
          observation:
            result.problem.kind === "unavailable" ? "unavailable" : "failed",
          problem: result.problem,
        }));
        this.announce(
          `${referencePackCommandLabel(job.command)} observation unavailable. Retry observation is available.`,
        );
        if (result.problem.kind === "unavailable") void this.reconcile(false);
        return false;
      }
      if (!this.acceptJob(id, result.value)) {
        this.updateJob(id, (value) => ({
          ...value,
          observation: "failed",
          problem: referencePackContractProblem,
        }));
        this.announce(
          `${referencePackCommandLabel(job.command)} returned an inconsistent observation. Retry observation is available.`,
        );
        return false;
      }
      return true;
    })();
    this.jobRead = read;
    try {
      return await read;
    } finally {
      if (this.jobRead === read) {
        this.jobRead = null;
        this.jobAbort = null;
        this.schedulePoll();
      }
    }
  }
  private acceptJob(id: string, snapshot: ReferencePackJobResource) {
    const job = this.state.jobs[id];
    if (!job || !acceptsReferencePackJob(job.snapshot, snapshot)) return false;
    const terminal = terminalReferencePackJobStates.has(snapshot.status);
    const newlyTerminal =
      terminal &&
      (!terminalReferencePackJobStates.has(job.snapshot.status) ||
        (this.state.operation?.phase === "accepted" &&
          this.state.operation.jobId === id));
    this.publish(jobObserved(this.state, id, snapshot));
    if (snapshot.status !== job.snapshot.status || newlyTerminal)
      this.announce(
        `${referencePackCommandLabel(job.command)} ${snapshot.status.replaceAll("_", " ")}.`,
      );
    if (newlyTerminal) this.invalidateCatalog();
    return true;
  }
  async cancelJob(id: string, replay = false) {
    const authority = this.state.authority;
    const job = this.state.jobs[id];
    if (
      !authority ||
      !job ||
      !this.canCommand() ||
      ["checking", "submitting"].includes(job.cancellation?.phase ?? "")
    )
      return;
    if (replay && job.cancellation?.phase !== "uncertain") return;
    if (
      !replay &&
      (job.cancellation?.phase === "uncertain" ||
        terminalReferencePackJobStates.has(job.snapshot.status))
    )
      return;
    const request =
      replay && job.cancellation
        ? job.cancellation.request
        : Object.freeze({
            client_txn_id: this.ports.transactionId("reference-pack-cancel"),
          });
    const epoch = this.epoch;
    const activation = this.activation;
    const current = () =>
      this.current(authority, epoch) &&
      this.state.jobs[id]?.cancellation?.request === request;
    this.stopPolling();
    this.updateJob(id, (value) => ({
      ...value,
      cancellation: { request, phase: "checking", problem: null },
    }));
    const deadline = performance.now() + referencePackTiming.read;
    if (this.jobRead !== null) await this.jobRead;
    if (!current() || activation !== this.activation || !this.state.active)
      return;
    if (!replay) {
      const observed =
        performance.now() < deadline &&
        (await this.observeJob(id, deadline - performance.now()));
      if (!current() || activation !== this.activation || !this.state.active)
        return;
      const latest = this.state.jobs[id]?.snapshot;
      if (
        !observed ||
        !latest?.cancelable ||
        terminalReferencePackJobStates.has(latest.status)
      ) {
        this.updateJob(id, (value) => ({
          ...value,
          cancellation: {
            request,
            phase: "rejected",
            problem: observed
              ? { kind: "rejected", status: 409, code: "job_cancel_rejected" }
              : referencePackTransportProblem,
          },
        }));
        this.announce(
          `${referencePackCommandLabel(job.command)} cancellation was not submitted. Review its current state.`,
        );
        this.schedulePoll();
        return;
      }
    }
    this.updateJob(id, (value) => ({
      ...value,
      cancellation: { request, phase: "submitting", problem: null },
    }));
    const outcome = await this.bounded(
      (signal) => this.ports.cancel(id, request, signal),
      referencePackTiming.read,
    );
    if (!current()) return;
    const result =
      outcome.kind === "value"
        ? outcome.value
        : ({
            kind: "uncertain",
            problem: referencePackTransportProblem,
          } as const);
    if (result.kind === "access_failed") {
      this.authorizationLost(result.status);
      return;
    }
    if (result.kind === "acknowledged") {
      const accepted = this.acceptJob(id, result.job);
      this.updateJob(id, (value) => ({
        ...value,
        cancellation: {
          request,
          phase: accepted ? "acknowledged" : "uncertain",
          problem: accepted ? null : referencePackContractProblem,
        },
      }));
      this.announce(
        accepted
          ? `${referencePackCommandLabel(job.command)} cancellation acknowledged. Current state: ${result.job.status.replaceAll("_", " ")}.`
          : `${referencePackCommandLabel(job.command)} cancellation outcome uncertain. Retry cancellation request is available.`,
      );
    } else {
      this.updateJob(id, (value) => ({
        ...value,
        cancellation: { request, phase: result.kind, problem: result.problem },
      }));
      this.announce(
        `${referencePackCommandLabel(job.command)} cancellation ${result.kind}.`,
      );
    }
    if (this.state.active && this.state.access === "ready")
      void this.observeJob(id);
  }
  dismissJob(id: string) {
    const job = this.state.jobs[id];
    if (
      !job ||
      !terminalReferencePackJobStates.has(job.snapshot.status) ||
      (this.state.operation?.jobId === id &&
        this.state.operation.phase === "accepted")
    )
      return;
    const jobs = { ...this.state.jobs };
    delete jobs[id];
    this.publish({ ...this.state, jobs });
  }
}
