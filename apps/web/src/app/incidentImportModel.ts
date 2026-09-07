import { clientTxnID } from "../services/browserApi";
import {
  captureImport,
  type ImportAttempt,
  type ImportCancelAttempt,
  type ImportJobResponse,
  type IncidentImportJob,
  importedIncidentTarget,
  importJobAdvances,
  terminalImportJob,
  validImportJob,
} from "./api/incidentImportClient";

export type ImportProblem =
  | "transport"
  | "contract"
  | "rejected"
  | "conflict"
  | "unavailable";
type Admission =
  | { readonly kind: "idle" }
  | { readonly kind: "pending" | "uncertain"; readonly attempt: ImportAttempt }
  | { readonly kind: "rejected"; readonly problem: ImportProblem }
  | { readonly kind: "accepted"; readonly jobId: string };
type Observation =
  | { readonly kind: "stale" | "reading" | "ready" | "unavailable" }
  | { readonly kind: "failed"; readonly problem: ImportProblem };
type Cancellation =
  | { readonly kind: "idle" | "observed" }
  | {
      readonly kind: "pending" | "uncertain";
      readonly attempt: ImportCancelAttempt;
    }
  | { readonly kind: "rejected"; readonly problem: ImportProblem };
type Navigation =
  | { readonly kind: "idle" }
  | {
      readonly kind: "opening" | "opened" | "unavailable" | "access_lost";
      readonly jobId: string;
    };
export type KnownImport = {
  readonly job: IncidentImportJob;
  readonly filename: string;
  readonly observation: Observation;
  readonly cancellation: Cancellation;
};
export type IncidentImportState = {
  readonly access: "checking" | "ready";
  readonly selectedFile: File | null;
  readonly fieldError: "required" | null;
  readonly admission: Admission;
  readonly jobs: Readonly<Record<string, KnownImport>>;
  readonly order: readonly string[];
  readonly selectedJobId: string | null;
  readonly paused: boolean;
  readonly navigation: Navigation;
  readonly announcement: {
    readonly sequence: number;
    readonly text: string;
    readonly priority: "polite" | "assertive";
  };
};
export type ImportAuthority = {
  readonly lifetime: string;
  readonly actorId: string;
};
export type IncidentImportPorts = {
  readonly admit: (
    attempt: ImportAttempt,
    signal: AbortSignal,
    replay: boolean,
  ) => Promise<ImportJobResponse>;
  readonly read: (
    id: string,
    signal: AbortSignal,
  ) => Promise<ImportJobResponse>;
  readonly cancel: (
    id: string,
    attempt: ImportCancelAttempt,
    signal: AbortSignal,
  ) => Promise<ImportJobResponse>;
  readonly isCurrent: (authority: ImportAuthority) => boolean;
  readonly authorizationFailed: (status: number) => void;
  readonly openIncident: (
    id: string,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<"opened" | "cancelled" | "unavailable" | "access_lost">;
  readonly transactionId?: () => string;
};
export type IncidentImportBinding = {
  readonly controller: IncidentImportController;
  readonly state: IncidentImportState;
};
type WaitResult<T> =
  | { kind: "response"; value: T }
  | { kind: "lost" | "aborted" };
const initialState = (): IncidentImportState => ({
  access: "checking",
  selectedFile: null,
  fieldError: null,
  admission: { kind: "idle" },
  jobs: {},
  order: [],
  selectedJobId: null,
  paused: false,
  navigation: { kind: "idle" },
  announcement: { sequence: 0, text: "", priority: "polite" },
});
export const importStatusLabel: Record<IncidentImportJob["status"], string> = {
  queued: "Queued",
  running: "Processing",
  cancel_requested: "Cancellation requested",
  succeeded: "Import succeeded",
  failed: "Import failed",
  canceled: "Import canceled",
};
export const admissionUnresolved = (state: IncidentImportState) =>
  state.admission.kind === "pending" || state.admission.kind === "uncertain";
export const availableImport = (entry: KnownImport) =>
  entry.observation.kind === "ready" &&
  (entry.job.retained_until === null ||
    Date.parse(entry.job.retained_until) > Date.now());
export const cancelableImport = (entry: KnownImport) =>
  availableImport(entry) &&
  entry.job.cancelable &&
  (entry.job.status === "queued" || entry.job.status === "running") &&
  entry.cancellation.kind !== "pending";
export const openableImport = (entry: KnownImport) =>
  availableImport(entry) ? importedIncidentTarget(entry.job) : null;
function problem(result: ImportJobResponse): ImportProblem {
  if (result.ok) return "contract";
  if (result.payload.error?.code === "client_txn_conflict") return "conflict";
  if (result.payload.error?.code === "invalid_public_contract_response")
    return "contract";
  if (result.status === 404) return "unavailable";
  return result.status >= 500 ? "transport" : "rejected";
}

/** Owns only session-local Incident Bundle imports. No React or incident socket. */
export class IncidentImportController {
  private state = initialState();
  private authority: ImportAuthority | null = null;
  private epoch = 0;
  private active = false;
  private disposed = false;
  private listeners = new Set<() => void>();
  private stops = new Set<() => void>();
  private readStop: (() => void) | null = null;
  private readId: string | null = null;
  private readGeneration = 0;
  private navigationGeneration = 0;
  private navigationStop: (() => void) | null = null;
  private timer: ReturnType<typeof setTimeout> | undefined;
  private expiryTimer: ReturnType<typeof setTimeout> | undefined;
  private queue = new Set<string>();
  private cursor = 0;
  constructor(private readonly ports: IncidentImportPorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<IncidentImportState>) {
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  private announce(text: string, priority: "polite" | "assertive" = "polite") {
    if (this.active)
      this.publish({
        announcement: {
          sequence: this.state.announcement.sequence + 1,
          text,
          priority,
        },
      });
  }
  private current(epoch = this.epoch): boolean {
    return (
      !this.disposed &&
      epoch === this.epoch &&
      this.authority !== null &&
      this.ports.isCurrent(this.authority)
    );
  }
  setAuthority(authority: ImportAuthority | null) {
    if (
      this.authority?.lifetime === authority?.lifetime &&
      this.authority?.actorId === authority?.actorId
    )
      return;
    this.retire();
    if (!this.disposed) {
      this.authority = authority;
      if (authority !== null) this.publish({ access: "ready" });
    }
  }
  retire = () => {
    ++this.epoch;
    this.authority = null;
    this.stopReads();
    this.stopNavigation();
    for (const stop of this.stops) stop();
    this.stops.clear();
    clearTimeout(this.expiryTimer);
    this.cursor = 0;
    this.publish(initialState());
  };
  dispose = () => {
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  };
  setActive = (active: boolean) => {
    if (this.active === active) return;
    this.active = active;
    if (!active) {
      this.stopReads();
      this.stopNavigation();
      const jobs = Object.fromEntries(
        Object.entries(this.state.jobs).map(([id, entry]) => [
          id,
          {
            ...entry,
            observation:
              entry.observation.kind === "ready" ||
              entry.observation.kind === "reading"
                ? { kind: "stale" as const }
                : entry.observation,
          },
        ]),
      );
      this.publish({
        jobs,
        announcement: { ...this.state.announcement, text: "" },
      });
    } else if (this.current()) {
      for (const id of this.state.order) this.queue.add(id);
      this.schedule(0);
    }
  };
  selectFile = (file: File | null) => {
    if (!this.current() || !this.active || admissionUnresolved(this.state))
      return;
    this.publish({
      selectedFile: file,
      fieldError: null,
      admission: { kind: "idle" },
    });
  };
  submit = () => {
    if (!this.current() || !this.active || admissionUnresolved(this.state))
      return;
    if (this.state.selectedFile === null) {
      this.publish({ fieldError: "required" });
      this.announce("Select an incident bundle file.", "assertive");
      return;
    }
    const attempt = captureImport(
      this.state.selectedFile,
      this.transactionId(),
    );
    void this.admit(attempt);
  };
  retryAdmission = () => {
    if (
      this.current() &&
      this.active &&
      this.state.admission.kind === "uncertain"
    )
      void this.admit(this.state.admission.attempt, true);
  };
  private transactionId() {
    return this.ports.transactionId?.() ?? clientTxnID("incident-import");
  }
  private async admit(attempt: ImportAttempt, replay = false) {
    if (this.state.admission.kind === "pending") return;
    const epoch = this.epoch;
    this.publish({ admission: { kind: "pending", attempt }, fieldError: null });
    this.announce("Uploading bundle and awaiting admission.");
    const result = await this.wait(
      (signal) => this.ports.admit(attempt, signal, replay),
      120_000,
    );
    if (!this.current(epoch)) return;
    if (result.kind !== "response") {
      this.publish({ admission: { kind: "uncertain", attempt } });
      this.announce(
        "Import admission is unconfirmed. Retry the same upload to recover.",
        "assertive",
      );
      return;
    }
    const response = result.value;
    if (
      response.ok &&
      response.status === 202 &&
      validImportJob(response.payload.data) &&
      (replay ||
        ["queued", "running"].includes(response.payload.data.status)) &&
      response.payload.data.submitted_by_user_id === this.authority?.actorId
    ) {
      const job = response.payload.data;
      // Replayed admission is a receipt, not a newer observation of a known job.
      const existing = this.state.jobs[job.job_id];
      this.stopNavigation();
      this.publish({
        selectedFile: null,
        admission: { kind: "accepted", jobId: job.job_id },
        selectedJobId: job.job_id,
        order: existing ? this.state.order : [...this.state.order, job.job_id],
        jobs: {
          ...this.state.jobs,
          [job.job_id]: existing ?? {
            job,
            filename: attempt.filename,
            observation: { kind: "stale" },
            cancellation: { kind: "idle" },
          },
        },
      });
      this.announce(
        "Import accepted. Server processing can continue when you leave this panel.",
      );
      this.refresh(job.job_id);
      return;
    }
    if (!response.ok && (response.status === 401 || response.status === 403)) {
      this.authorizationFailure(response.status);
      return;
    }
    const uncertain =
      response.ok ||
      response.status >= 500 ||
      response.status === 408 ||
      response.status === 0;
    this.publish({
      admission: uncertain
        ? { kind: "uncertain", attempt }
        : { kind: "rejected", problem: problem(response) },
    });
    this.announce(
      uncertain
        ? "Import admission is unconfirmed. Retry the same upload to recover."
        : "Import admission was rejected. Review the bundle before submitting again.",
      "assertive",
    );
  }
  selectJob = (id: string) => {
    if (!this.current() || !this.active || !this.state.jobs[id]) return;
    if (id !== this.state.selectedJobId) {
      this.stopNavigation();
      this.publish({ selectedJobId: id });
    }
    if (this.state.jobs[id]?.observation.kind !== "ready") this.refresh(id);
  };
  refresh = (id = this.state.selectedJobId) => {
    if (!this.current() || !id || !this.state.jobs[id]) return;
    this.queue.add(id);
    this.schedule(0);
  };
  pause = () => {
    this.stopReads();
    this.publish({ paused: true });
    this.announce("Automatic job updates paused.");
  };
  resume = () => {
    if (!this.current()) return;
    this.publish({ paused: false });
    for (const id of this.state.order) this.queue.add(id);
    this.schedule(0);
    this.announce("Automatic job updates resumed.");
  };
  private stopReads() {
    clearTimeout(this.timer);
    this.timer = undefined;
    ++this.readGeneration;
    this.readStop?.();
    this.readStop = null;
    this.readId = null;
    this.queue.clear();
    const jobs = { ...this.state.jobs };
    for (const [id, entry] of Object.entries(jobs))
      if (entry.observation.kind === "reading")
        jobs[id] = { ...entry, observation: { kind: "stale" } };
    this.publish({ jobs });
  }
  private schedule(delay: number) {
    if (!this.current() || !this.active || this.readId !== null) return;
    clearTimeout(this.timer);
    this.timer = setTimeout(() => {
      this.timer = undefined;
      this.pump();
    }, delay);
  }
  private pump() {
    if (!this.current() || !this.active || this.readId !== null) return;
    let id: string | undefined;
    for (const queued of this.queue) {
      if (this.state.jobs[queued]?.cancellation.kind === "pending") continue;
      this.queue.delete(queued);
      id = queued;
      break;
    }
    if (id === undefined && !this.state.paused) {
      const count = this.state.order.length;
      for (let i = 0; i < count; i++) {
        const candidate = this.state.order[this.cursor++ % count];
        const entry = candidate ? this.state.jobs[candidate] : undefined;
        if (
          entry &&
          !terminalImportJob(entry.job) &&
          entry.observation.kind === "ready" &&
          entry.cancellation.kind !== "pending"
        ) {
          id = candidate;
          break;
        }
      }
    }
    if (id !== undefined) void this.observe(id);
  }
  private update(id: string, patch: Partial<KnownImport>) {
    const entry = this.state.jobs[id];
    if (entry)
      this.publish({
        jobs: { ...this.state.jobs, [id]: { ...entry, ...patch } },
      });
  }
  private acceptJob(id: string, job: IncidentImportJob): boolean {
    const previous = this.state.jobs[id];
    if (!previous || !importJobAdvances(previous.job, job)) return false;
    this.update(id, {
      job,
      observation: { kind: this.active ? "ready" : "stale" },
      cancellation:
        job.status === "cancel_requested" || terminalImportJob(job)
          ? { kind: "observed" }
          : previous.cancellation,
    });
    if (previous.job.status !== job.status)
      this.announce(importStatusLabel[job.status]);
    this.expireJobs();
    return true;
  }
  private async observe(id: string) {
    const epoch = this.epoch;
    const generation = ++this.readGeneration;
    this.readId = id;
    this.update(id, { observation: { kind: "reading" } });
    const result = await this.wait(
      (signal) => this.ports.read(id, signal),
      30_000,
      (stop) => {
        this.readStop = stop;
      },
    );
    if (
      !this.current(epoch) ||
      generation !== this.readGeneration ||
      !this.active
    )
      return;
    this.readStop = null;
    this.readId = null;
    if (
      result.kind === "response" &&
      result.value.ok &&
      this.acceptJob(id, result.value.payload.data)
    ) {
      // Status changes alone announce; successful polls preserve focus and silence.
    } else {
      const response = result.kind === "response" ? result.value : null;
      if (
        response &&
        !response.ok &&
        (response.status === 401 || response.status === 403)
      ) {
        this.authorizationFailure(response.status);
        return;
      }
      const issue = response ? problem(response) : "transport";
      this.update(id, {
        observation:
          issue === "unavailable"
            ? { kind: "unavailable" }
            : { kind: "failed", problem: issue },
      });
      this.announce(
        issue === "unavailable"
          ? "The import job is no longer available."
          : "Job observation failed. The last validated status is retained. Retry observation to recover.",
        "assertive",
      );
      if (issue === "unavailable") this.ports.authorizationFailed(404);
    }
    this.schedule(1000);
  }
  cancel = (id = this.state.selectedJobId) => {
    if (!this.current() || !this.active || !id) return;
    const entry = this.state.jobs[id];
    if (!entry || !cancelableImport(entry)) return;
    const attempt =
      entry.cancellation.kind === "uncertain"
        ? entry.cancellation.attempt
        : Object.freeze({ client_txn_id: this.transactionId() });
    void this.cancelAttempt(id, attempt);
  };
  private async cancelAttempt(id: string, attempt: ImportCancelAttempt) {
    const epoch = this.epoch;
    if (this.readId === id) this.stopReads();
    this.update(id, { cancellation: { kind: "pending", attempt } });
    this.announce("Requesting cancellation.");
    const result = await this.wait(
      (signal) => this.ports.cancel(id, attempt, signal),
      30_000,
    );
    if (!this.current(epoch)) return;
    const response = result.kind === "response" ? result.value : null;
    if (response?.ok && this.acceptJob(id, response.payload.data)) {
      this.update(id, { cancellation: { kind: "observed" } });
    } else {
      if (
        response &&
        !response.ok &&
        (response.status === 401 || response.status === 403)
      ) {
        this.authorizationFailure(response.status);
        return;
      }
      const issue = response ? problem(response) : "transport";
      const uncertain =
        !response ||
        response.ok ||
        (!response.ok && (response.status >= 500 || response.status === 408));
      this.update(id, {
        cancellation: uncertain
          ? { kind: "uncertain", attempt }
          : { kind: "rejected", problem: issue },
        observation: { kind: "stale" },
      });
      this.announce(
        uncertain
          ? "Cancellation is unconfirmed. Checking the job status."
          : "Cancellation was rejected. Checking the job status.",
        "assertive",
      );
    }
    this.refresh(id);
  }
  open = () => {
    if (
      !this.current() ||
      !this.active ||
      this.state.navigation.kind === "opening"
    )
      return;
    const id = this.state.selectedJobId;
    const entry = id ? this.state.jobs[id] : undefined;
    const target = entry ? openableImport(entry) : null;
    if (id && target) void this.openTarget(id, target);
  };
  private async openTarget(jobId: string, incidentId: string) {
    const epoch = this.epoch;
    const generation = ++this.navigationGeneration;
    const current = () =>
      this.current(epoch) &&
      this.active &&
      this.navigationGeneration === generation &&
      this.state.selectedJobId === jobId &&
      this.state.jobs[jobId] !== undefined &&
      openableImport(this.state.jobs[jobId]) === incidentId;
    this.publish({ navigation: { kind: "opening", jobId } });
    this.announce("Opening imported incident.");
    const result = await this.wait(
      (signal) => this.ports.openIncident(incidentId, signal, current),
      30_000,
      (stop) => {
        this.navigationStop = stop;
      },
    );
    if (!current()) return;
    this.navigationStop = null;
    const outcome = result.kind === "response" ? result.value : "unavailable";
    this.publish({
      navigation:
        outcome === "cancelled" ? { kind: "idle" } : { kind: outcome, jobId },
    });
    if (outcome === "unavailable" || outcome === "access_lost")
      this.announce(
        "The import succeeded, but the workbook could not be opened. You can retry opening it.",
        "assertive",
      );
  }
  private stopNavigation() {
    ++this.navigationGeneration;
    this.navigationStop?.();
    this.navigationStop = null;
    if (this.state.navigation.kind !== "idle")
      this.publish({ navigation: { kind: "idle" } });
  }
  private authorizationFailure(status: number) {
    this.retire();
    this.ports.authorizationFailed(status);
  }
  private expireJobs() {
    clearTimeout(this.expiryTimer);
    let next = Number.POSITIVE_INFINITY;
    for (const [id, entry] of Object.entries(this.state.jobs)) {
      if (
        entry.job.retained_until === null ||
        entry.observation.kind === "unavailable"
      )
        continue;
      const remaining = Date.parse(entry.job.retained_until) - Date.now();
      if (remaining <= 0) {
        this.update(id, { observation: { kind: "unavailable" } });
        if (this.state.selectedJobId === id) this.stopNavigation();
      } else next = Math.min(next, remaining);
    }
    if (Number.isFinite(next))
      this.expiryTimer = setTimeout(
        () => this.expireJobs(),
        Math.min(next, 2_147_483_647),
      );
  }
  /** Bounded local observations settle even when a transport ignores abort. */
  private wait<T>(
    run: (signal: AbortSignal) => Promise<T>,
    timeoutMs: number,
    register?: (stop: () => void) => void,
  ): Promise<WaitResult<T>> {
    return new Promise((resolve) => {
      const abort = new AbortController();
      let done = false;
      const finish = (result: WaitResult<T>) => {
        if (done) return;
        done = true;
        clearTimeout(timer);
        this.stops.delete(stop);
        resolve(result);
      };
      const stop = () => {
        finish({ kind: "aborted" });
        abort.abort();
      };
      const timer = setTimeout(() => {
        finish({ kind: "lost" });
        abort.abort();
      }, timeoutMs);
      this.stops.add(stop);
      register?.(stop);
      try {
        run(abort.signal).then(
          (value) => finish({ kind: "response", value }),
          () => finish({ kind: "lost" }),
        );
      } catch {
        finish({ kind: "lost" });
      }
    });
  }
}
