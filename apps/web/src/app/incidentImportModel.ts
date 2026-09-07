import { clientTxnID } from "../services/browserApi";
import {
  captureImport,
  type ImportAdmissionOutcome,
  type ImportAttempt,
  type ImportCancelAttempt,
  type ImportCancellationOutcome,
  type ImportObservationOutcome,
  importedIncidentTarget,
  terminalImportJob,
} from "./api/incidentImportClient";
import {
  admissionUnresolved,
  cancelableImport,
  type ImportAuthority,
  type ImportEvent,
  type IncidentImportState,
  importStatusLabel,
  initialImportState,
  openableImport,
  transitionImport,
} from "./incidentImportState";

export type {
  ImportAuthority,
  IncidentImportState,
  KnownImport,
} from "./incidentImportState";
export {
  admissionUnresolved,
  availableImport,
  cancelableImport,
  importStatusLabel,
  openableImport,
} from "./incidentImportState";
export type ImportAccessConfirmation =
  | { readonly kind: "authorized"; readonly authority: ImportAuthority }
  | {
      readonly kind:
        | "session_lost"
        | "access_lost"
        | "unavailable"
        | "cancelled";
    };
export type IncidentImportPorts = {
  readonly confirmAccess: (
    authority: ImportAuthority,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<ImportAccessConfirmation>;
  readonly admit: (
    attempt: ImportAttempt,
    signal: AbortSignal,
    replay: boolean,
  ) => Promise<ImportAdmissionOutcome>;
  readonly read: (
    id: string,
    signal: AbortSignal,
  ) => Promise<ImportObservationOutcome>;
  readonly cancel: (
    id: string,
    attempt: ImportCancelAttempt,
    signal: AbortSignal,
  ) => Promise<ImportCancellationOutcome>;
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

type ActionIntent = {
  readonly jobId: string;
  readonly action: "cancel" | "open";
  readonly epoch: number;
  readonly deadline: number;
};

/** Bounded asynchronous executor for the session-local Incident Bundle workflow. */
export class IncidentImportController {
  private state = initialImportState();
  private authority: ImportAuthority | null = null;
  private boundAuthority: ImportAuthority | null = null;
  private recoveryAuthority: ImportAuthority | null = null;
  private recoveryGeneration = 0;
  private recoveryStop: (() => void) | null = null;
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
  private pendingAction: ActionIntent | null = null;
  private readAction: ActionIntent | null = null;
  private actionTimer: ReturnType<typeof setTimeout> | undefined;
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
  private dispatch(event: ImportEvent) {
    const next = transitionImport(this.state, event);
    if (next === this.state) return;
    this.state = next;
    for (const listener of this.listeners) listener();
  }
  private announce(text: string, priority: "polite" | "assertive" = "polite") {
    if (this.active) this.dispatch({ type: "announce", text, priority });
  }
  private current(epoch = this.epoch): boolean {
    return (
      !this.disposed &&
      epoch === this.epoch &&
      this.authority !== null &&
      this.ports.isCurrent(this.authority)
    );
  }
  private authorized() {
    return this.current() && this.state.access === "ready";
  }
  setAuthority(authority: ImportAuthority | null) {
    if (
      this.boundAuthority?.lifetime === authority?.lifetime &&
      this.boundAuthority?.actorId === authority?.actorId
    )
      return;
    this.retire();
    this.boundAuthority = authority;
    if (!this.disposed) {
      this.authority = authority;
      if (authority !== null)
        this.dispatch({ type: "access", access: "ready" });
    }
  }
  retire = () => {
    ++this.epoch;
    ++this.recoveryGeneration;
    this.recoveryStop?.();
    this.recoveryStop = null;
    this.recoveryAuthority = null;
    this.authority = null;
    this.stopReads();
    this.stopNavigation();
    for (const stop of this.stops) stop();
    this.stops.clear();
    this.stopAction();
    this.cursor = 0;
    this.dispatch({ type: "reset" });
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
      this.stopAction();
      this.stopReads();
      this.stopNavigation();
      this.dispatch({ type: "inactive" });
    } else if (this.authorized()) {
      for (const id of this.state.order) this.queue.add(id);
      this.schedule(0);
    }
  };
  selectFile = (file: File | null) => {
    if (!this.authorized() || !this.active || admissionUnresolved(this.state))
      return;
    this.dispatch({ type: "file", file });
  };
  submit = () => {
    if (!this.authorized() || !this.active || admissionUnresolved(this.state))
      return;
    if (this.state.selectedFile === null) {
      this.dispatch({ type: "required" });
      this.announce("Select an incident bundle file.", "assertive");
      return;
    }
    void this.admit(
      captureImport(this.state.selectedFile, this.transactionId()),
    );
  };
  retryAdmission = () => {
    if (
      this.authorized() &&
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
    this.dispatch({ type: "admission_started", attempt });
    this.announce("Uploading bundle and awaiting admission.");
    const result = await this.wait(
      (signal) => this.ports.admit(attempt, signal, replay),
      120_000,
    );
    if (!this.current(epoch) || !this.authority) return;
    const outcome =
      result.kind === "response"
        ? result.value
        : { kind: "uncertain" as const, problem: "transport" as const };
    if (outcome.kind === "access_failed") {
      this.authorizationFailure(outcome.status);
      return;
    }
    if (outcome.kind === "accepted") {
      this.stopAction();
      this.stopNavigation();
    }
    this.dispatch({
      type: "admission_finished",
      attempt,
      outcome,
      actorId: this.authority.actorId,
    });
    const admission = this.state.admission;
    if (admission.kind === "accepted") {
      this.announce(
        "Import accepted. Server processing can continue when you leave this panel.",
      );
      this.refresh(admission.jobId);
    } else
      this.announce(
        admission.kind === "uncertain"
          ? "Import admission is unconfirmed. Retry the same upload to recover."
          : "Import admission was rejected. Review the bundle before submitting again.",
        "assertive",
      );
  }
  selectJob = (id: string) => {
    if (!this.authorized() || !this.active || !this.state.jobs[id]) return;
    if (id !== this.state.selectedJobId) {
      this.stopAction();
      this.stopNavigation();
      this.dispatch({ type: "selected", id });
    }
    if (this.state.jobs[id]?.observation.kind !== "ready") this.refresh(id);
  };
  refresh = (id = this.state.selectedJobId) => {
    if (!this.authorized() || !id || !this.state.jobs[id]) return;
    this.queue.add(id);
    this.schedule(0);
  };
  pause = () => {
    clearTimeout(this.timer);
    this.dispatch({ type: "paused", paused: true });
    this.schedule(0);
    this.announce("Automatic job updates paused.");
  };
  resume = () => {
    if (!this.authorized()) return;
    this.dispatch({ type: "paused", paused: false });
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
    this.dispatch({ type: "reads_stopped" });
  }
  private schedule(delay: number) {
    if (!this.authorized() || !this.active || this.readId !== null) return;
    clearTimeout(this.timer);
    this.timer = setTimeout(() => {
      this.timer = undefined;
      this.pump();
    }, delay);
  }
  private pump() {
    if (!this.authorized() || !this.active || this.readId !== null) return;
    if (this.pendingAction) {
      const action = this.pendingAction;
      if (this.actionCurrent(action)) void this.observe(action.jobId, action);
      else this.failAction(action);
      return;
    }
    let id: string | undefined;
    for (const queued of this.queue) {
      if (
        this.state.jobs[queued]?.cancellation.kind === "pending" ||
        (this.state.navigation.kind === "opening" &&
          this.state.navigation.jobId === queued)
      )
        continue;
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
  private async observe(id: string, action: ActionIntent | null = null) {
    const epoch = this.epoch;
    const generation = ++this.readGeneration;
    this.readId = id;
    this.readAction = action;
    this.dispatch({ type: "read_started", id });
    const result = await this.wait(
      (signal) => this.ports.read(id, signal),
      action ? Math.max(0, action.deadline - performance.now()) : 30_000,
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
    this.readAction = null;
    const outcome =
      result.kind === "response"
        ? result.value
        : { kind: "failed" as const, problem: "transport" as const };
    if (outcome.kind === "access_failed") {
      this.authorizationFailure(outcome.status);
      return;
    }
    const previous = this.state.jobs[id]?.job.status;
    this.dispatch({ type: "read_finished", id, outcome, active: this.active });
    const entry = this.state.jobs[id];
    if (entry?.observation.kind === "ready") {
      if (previous !== entry.job.status)
        this.announce(importStatusLabel[entry.job.status]);
    } else {
      this.announce(
        outcome.kind === "unavailable"
          ? "The import job is no longer available."
          : "Job observation failed. The last validated status is retained. Retry observation to recover.",
        "assertive",
      );
      if (outcome.kind === "unavailable") this.beginAccessConfirmation();
    }
    if (action && this.pendingAction === action) {
      if (entry?.observation.kind === "ready" && this.actionCurrent(action)) {
        const target = openableImport(entry);
        if (action.action === "cancel" && cancelableImport(entry)) {
          const attempt =
            entry.cancellation.kind === "uncertain"
              ? entry.cancellation.attempt
              : Object.freeze({ client_txn_id: this.transactionId() });
          this.stopAction();
          void this.cancelAttempt(id, attempt);
        } else if (action.action === "open" && target) {
          this.stopAction();
          void this.openTarget(id, target);
        } else this.failAction(action);
      } else this.failAction(action);
    }
    this.schedule(this.pendingAction ? 0 : 1000);
  }
  cancel = (id = this.state.selectedJobId) => {
    if (!id || id !== this.state.selectedJobId) return;
    const entry = this.state.jobs[id];
    if (entry && cancelableImport(entry)) this.requestAction("cancel", id);
  };
  private requestAction(action: "open" | "cancel", jobId: string) {
    if (
      !this.authorized() ||
      !this.active ||
      this.pendingAction ||
      this.state.navigation.kind === "opening"
    )
      return;
    const intent: ActionIntent = {
      action,
      jobId,
      epoch: this.epoch,
      deadline: performance.now() + 30_000,
    };
    this.pendingAction = intent;
    this.dispatch({
      type: "action",
      action: { kind: "checking", action, jobId },
    });
    this.announce("Checking current job status.");
    this.actionTimer = setTimeout(() => {
      if (this.pendingAction !== intent) return;
      this.failAction(intent);
      if (this.readAction === intent) this.readStop?.();
    }, 30_000);
    this.schedule(0);
  }
  private actionCurrent(action: ActionIntent) {
    return (
      this.pendingAction === action &&
      this.current(action.epoch) &&
      this.state.access === "ready" &&
      this.active &&
      this.state.selectedJobId === action.jobId &&
      performance.now() < action.deadline
    );
  }
  private stopAction() {
    clearTimeout(this.actionTimer);
    this.pendingAction = null;
    if (this.state.action.kind !== "idle")
      this.dispatch({ type: "action", action: { kind: "idle" } });
  }
  private failAction(action: ActionIntent) {
    if (this.pendingAction !== action) return;
    this.stopAction();
    this.dispatch({
      type: "action",
      action: { kind: "failed", action: action.action, jobId: action.jobId },
    });
    this.announce(
      "The action could not be confirmed. Review the current job status and retry.",
      "assertive",
    );
  }
  private async cancelAttempt(id: string, attempt: ImportCancelAttempt) {
    const epoch = this.epoch;
    if (this.readId === id) this.stopReads();
    this.dispatch({ type: "cancel_started", id, attempt });
    this.announce("Requesting cancellation.");
    const result = await this.wait(
      (signal) => this.ports.cancel(id, attempt, signal),
      30_000,
    );
    if (!this.current(epoch)) return;
    const outcome =
      result.kind === "response"
        ? result.value
        : { kind: "uncertain" as const, problem: "transport" as const };
    if (outcome.kind === "access_failed") {
      this.authorizationFailure(outcome.status);
      return;
    }
    const previous = this.state.jobs[id]?.job.status;
    this.dispatch({
      type: "cancel_finished",
      id,
      attempt,
      outcome,
      active: this.active,
    });
    const entry = this.state.jobs[id];
    if (entry?.cancellation.kind === "observed") {
      if (previous !== entry.job.status)
        this.announce(importStatusLabel[entry.job.status]);
    } else
      this.announce(
        entry?.cancellation.kind === "uncertain"
          ? "Cancellation is unconfirmed. Checking the job status."
          : "Cancellation was rejected. Checking the job status.",
        "assertive",
      );
    if (outcome.kind === "unavailable") this.beginAccessConfirmation();
    else this.refresh(id);
  }
  open = () => {
    const id = this.state.selectedJobId;
    const entry = id ? this.state.jobs[id] : undefined;
    if (id && entry && openableImport(entry)) this.requestAction("open", id);
  };
  private async openTarget(jobId: string, incidentId: string) {
    const epoch = this.epoch;
    const generation = ++this.navigationGeneration;
    const current = () => {
      const entry = this.state.jobs[jobId];
      return (
        this.current(epoch) &&
        this.active &&
        this.navigationGeneration === generation &&
        this.state.selectedJobId === jobId &&
        entry !== undefined &&
        entry.availability !== "unavailable" &&
        !(
          entry.observation.kind === "failed" &&
          entry.observation.problem === "contract"
        ) &&
        importedIncidentTarget(entry.job) === incidentId
      );
    };
    this.dispatch({
      type: "navigation",
      navigation: { kind: "opening", jobId },
    });
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
    this.dispatch({
      type: "navigation",
      navigation:
        outcome === "cancelled" ? { kind: "idle" } : { kind: outcome, jobId },
    });
    if (outcome === "unavailable" || outcome === "access_lost")
      this.announce(
        "The import succeeded, but the workbook could not be opened. You can retry opening it.",
        "assertive",
      );
    this.schedule(0);
  }
  private stopNavigation() {
    ++this.navigationGeneration;
    this.navigationStop?.();
    this.navigationStop = null;
    if (this.state.navigation.kind !== "idle")
      this.dispatch({ type: "navigation", navigation: { kind: "idle" } });
  }
  private authorizationFailure(status: number) {
    const authority = this.authority;
    this.retire();
    this.ports.authorizationFailed(status);
    if (status === 403 && authority && !this.disposed) {
      this.recoveryAuthority = authority;
      this.beginAccessConfirmation();
    }
  }
  retryAccess = () => {
    if (this.state.access === "unavailable") this.beginAccessConfirmation();
  };
  private beginAccessConfirmation() {
    const authority = this.recoveryAuthority ?? this.authority;
    if (!authority || this.disposed || this.recoveryStop) return;
    this.recoveryAuthority = authority;
    this.stopAction();
    this.stopReads();
    this.stopNavigation();
    this.dispatch({ type: "access", access: "checking" });
    void this.confirmAccess(authority);
  }
  private async confirmAccess(authority: ImportAuthority) {
    const epoch = this.epoch;
    const generation = ++this.recoveryGeneration;
    const current = () =>
      !this.disposed &&
      epoch === this.epoch &&
      generation === this.recoveryGeneration;
    const result = await this.wait(
      (signal) => this.ports.confirmAccess(authority, signal, current),
      30_000,
      (stop) => {
        this.recoveryStop = stop;
      },
    );
    if (!current()) return;
    this.recoveryStop = null;
    const outcome =
      result.kind === "response"
        ? result.value
        : { kind: "unavailable" as const };
    if (outcome.kind === "authorized") {
      if (
        outcome.authority.lifetime !== authority.lifetime ||
        outcome.authority.actorId !== authority.actorId ||
        !this.ports.isCurrent(outcome.authority)
      ) {
        this.retire();
        this.dispatch({ type: "access", access: "lost" });
        return;
      }
      this.authority = outcome.authority;
      this.recoveryAuthority = null;
      this.dispatch({ type: "access", access: "ready" });
      this.announce("Import access confirmed.");
      this.schedule(0);
    } else if (
      outcome.kind === "access_lost" ||
      outcome.kind === "session_lost"
    ) {
      this.retire();
      this.dispatch({ type: "access", access: "lost" });
      if (outcome.kind === "session_lost") this.ports.authorizationFailed(401);
    } else {
      this.dispatch({ type: "access", access: "unavailable" });
      this.announce(
        "Import access could not be confirmed. Retry access to continue.",
        "assertive",
      );
    }
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
