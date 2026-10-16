import { cartularyReportingProgressContract as timing } from "@cartulary/ui-contracts";
import {
  browserObservationClock,
  type ObservationClock,
  observeAsyncOperation,
} from "../services/asyncObservation";
import { createClientTransactionId } from "../services/clientTransactionId";
import {
  commonJobDoesNotRegress,
  terminalCommonJob,
  validCommonJob,
} from "../services/commonJobContract";
import {
  defaultReportTemplate,
  type ReportingAttempt,
  type ReportingCancellation,
  type ReportingClient,
  type ReportingJob,
  type ReportingReceipt,
} from "../services/reportingClient";

export type ReportingBinding = {
  readonly incidentId: string;
  readonly actorId: string;
  readonly lifetime: string;
  readonly role: string | null;
  readonly client: ReportingClient;
  readonly current: () => boolean;
  readonly accessFailure: () => void;
};
type ReportingState = {
  readonly available: boolean;
  readonly writable: boolean;
  readonly presented: boolean;
  readonly busy: boolean;
  readonly paused: boolean;
  readonly attempt: ReportingAttempt | null;
  readonly cancellation: ReportingCancellation | null;
  readonly kind: "snapshot" | "report" | null;
  readonly job: ReportingJob | null;
  readonly snapshotId: string | null;
  readonly releaseId: string | null;
  readonly problem: string | null;
};
const initial = (): ReportingState => ({
  available: false,
  writable: false,
  presented: false,
  busy: false,
  paused: false,
  attempt: null,
  cancellation: null,
  kind: null,
  job: null,
  snapshotId: null,
  releaseId: null,
  problem: null,
});
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const canWrite = (role: string | null) =>
  ["editor", "reviewer", "admin"].includes(role ?? "");

/** Reporting owns captured commands and job observation, independently of the drawer. */
export class ReportingController {
  private state = initial();
  private binding: ReportingBinding | null = null;
  private listeners = new Set<() => void>();
  private epoch = 0;
  private cancelObservation = () => {};
  private cancelPoll = () => {};
  private observingUntil = 0;
  private reading = false;
  private disposed = false;
  constructor(
    private readonly clock: ObservationClock = browserObservationClock,
    private readonly transactionId: () => string = () =>
      createClientTransactionId("reporting"),
  ) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private update(next: Partial<ReportingState>) {
    this.state = { ...this.state, ...next };
    for (const listener of this.listeners) listener();
  }
  bind(binding: ReportingBinding) {
    if (this.disposed) return;
    const previous = this.binding;
    if (
      previous &&
      (previous.incidentId !== binding.incidentId ||
        previous.actorId !== binding.actorId ||
        previous.lifetime !== binding.lifetime)
    )
      this.retire();
    if (previous && previous.role !== binding.role) this.pause();
    this.binding = binding;
    const writable = canWrite(binding.role);
    if (!this.state.available || this.state.writable !== writable)
      this.update({ available: true, writable });
  }
  setPresented(presented: boolean) {
    if (this.disposed || presented === this.state.presented) return;
    if (!presented) this.pause();
    this.update({ presented });
  }
  pause() {
    ++this.epoch;
    this.cancelPoll();
    this.cancelObservation();
    this.observingUntil = 0;
    this.reading = false;
    if (this.state.available || this.state.busy)
      this.update({
        available: false,
        busy: false,
        paused: !!(
          this.state.job ||
          this.state.attempt ||
          this.state.cancellation
        ),
      });
  }
  retire() {
    this.pause();
    this.binding = null;
    this.state = initial();
    for (const listener of this.listeners) listener();
  }
  dispose() {
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  }
  private authority(write: boolean) {
    const b = this.binding;
    return !this.disposed &&
      this.state.available &&
      this.state.presented &&
      b?.current() &&
      (!write || canWrite(b.role))
      ? b
      : null;
  }
  private newCommandAllowed() {
    return (
      !this.state.busy &&
      !this.state.attempt &&
      !this.state.cancellation &&
      (!this.state.job || terminalCommonJob(this.state.job))
    );
  }
  private newIdentity() {
    try {
      return this.transactionId();
    } catch {
      return null;
    }
  }
  createSnapshot = async () => {
    const b = this.authority(true);
    if (!b || !this.newCommandAllowed()) return;
    const id = this.newIdentity();
    if (!id) {
      this.update({ problem: "Secure request identity is unavailable." });
      return;
    }
    performance.mark?.(timing.start_mark, { detail: { transactionId: id } });
    const attempt: ReportingAttempt = Object.freeze({
      kind: "snapshot",
      request: Object.freeze({ incident_id: b.incidentId, client_txn_id: id }),
    });
    this.update({ attempt, kind: "snapshot", job: null, problem: null });
    await this.submit(attempt, b);
  };
  createReport = async () => {
    const b = this.authority(true),
      snapshotId = this.state.snapshotId;
    if (!b || !snapshotId || !this.newCommandAllowed()) return;
    const id = this.newIdentity();
    if (!id) {
      this.update({ problem: "Secure request identity is unavailable." });
      return;
    }
    performance.mark?.(timing.start_mark, { detail: { transactionId: id } });
    const attempt: ReportingAttempt = Object.freeze({
      kind: "report",
      request: Object.freeze({
        snapshot_id: snapshotId,
        client_txn_id: id,
        template_id: defaultReportTemplate.id,
        template_version: defaultReportTemplate.version,
        output_kind: "slidev",
        release_scope: "internal_draft",
      }),
    });
    // Capture before the supporting read. A failed read does not dispatch the command.
    const outcome = await this.observe(b, (signal) =>
      b.client.snapshot(snapshotId, signal),
    );
    if (!outcome) return;
    if (
      outcome.kind !== "completed" ||
      !outcome.value.ok ||
      outcome.value.payload.data.snapshot_id !== snapshotId ||
      outcome.value.payload.data.incident_id !== b.incidentId
    ) {
      this.update({
        problem:
          "The snapshot could not be verified. Refresh authority and try again.",
      });
      if (
        outcome.kind === "completed" &&
        !outcome.value.ok &&
        [401, 403].includes(outcome.value.status)
      )
        b.accessFailure();
      return;
    }
    if (!this.authority(true)) return;
    this.update({ attempt, kind: "report", job: null, problem: null });
    await this.submit(attempt, b);
  };
  replay = async () => {
    const b = this.authority(true),
      attempt = this.state.attempt;
    if (!b || !attempt || this.state.busy) return;
    await this.submit(attempt, b, true);
  };
  private async submit(
    attempt: ReportingAttempt,
    b: ReportingBinding,
    replay = false,
  ) {
    performance.mark?.(timing.dispatch_mark, {
      detail: { transactionId: attempt.request.client_txn_id, replay },
    });
    const result = await this.observe(b, (signal) =>
      b.client.submit(attempt, signal),
    );
    if (!result) return;
    if (
      result.kind === "completed" &&
      result.value.ok &&
      result.value.status === 202 &&
      this.accept(result.value.payload.data, b, true)
    ) {
      performance.mark?.(timing.admission_mark, {
        detail: {
          transactionId: attempt.request.client_txn_id,
          jobId: result.value.payload.data.job_id,
        },
      });
      this.update({ attempt: null, paused: false, problem: null });
      this.beginPolling();
      return;
    }
    const receipt = result.kind === "completed" ? result.value : null;
    const rejected =
      receipt &&
      !receipt.ok &&
      receipt.status >= 400 &&
      receipt.status < 500 &&
      receipt.status !== 408;
    this.update({
      attempt: rejected ? null : attempt,
      paused: true,
      problem: rejected
        ? "The request was rejected. Review your access and request before trying again."
        : "The request outcome is uncertain. Recover this exact request before starting another.",
    });
    if (receipt && [401, 403].includes(receipt.status)) b.accessFailure();
  }
  refresh = async () => {
    const b = this.authority(false),
      job = this.state.job;
    if (!b || !job || this.state.busy || this.reading || this.state.attempt)
      return;
    const result = await this.observe(
      b,
      (signal) => b.client.read(job.job_id, signal),
      true,
    );
    if (!result) return;
    if (
      result.kind === "completed" &&
      result.value.ok &&
      result.value.status === 200 &&
      this.accept(result.value.payload.data, b, false)
    ) {
      this.update({ paused: false, problem: null });
      if (!this.observingUntil)
        this.observingUntil = this.clock.now() + 120_000;
      this.schedulePoll();
      return;
    }
    this.update({
      paused: true,
      problem: "Job status is unavailable. Refresh to resume observation.",
    });
    if (result.kind === "completed" && [401, 403].includes(result.value.status))
      b.accessFailure();
  };
  cancel = async () => {
    const b = this.authority(true),
      job = this.state.job;
    if (
      !b ||
      !job ||
      this.state.busy ||
      (!this.state.cancellation && !job.cancelable)
    )
      return;
    const id = this.state.cancellation?.client_txn_id ?? this.newIdentity();
    if (!id) {
      this.update({ problem: "Secure request identity is unavailable." });
      return;
    }
    const request =
      this.state.cancellation ?? Object.freeze({ client_txn_id: id });
    this.cancelPoll();
    this.cancelObservation();
    this.update({ cancellation: request });
    const result = await this.observe(b, (signal) =>
      b.client.cancel(job.job_id, request, signal),
    );
    if (!result) return;
    if (
      result.kind === "completed" &&
      result.value.ok &&
      result.value.status === 200 &&
      this.accept(result.value.payload.data, b, false)
    ) {
      this.update({ cancellation: null, problem: null, paused: false });
      this.beginPolling();
      return;
    }
    const receipt: ReportingReceipt | null =
      result.kind === "completed" ? result.value : null;
    const rejected =
      receipt &&
      !receipt.ok &&
      receipt.status >= 400 &&
      receipt.status < 500 &&
      receipt.status !== 408;
    this.update({
      cancellation: rejected ? null : request,
      paused: true,
      problem: rejected
        ? "Cancellation was rejected. Refresh the current job status."
        : "Cancellation is uncertain. Recover the same cancellation request.",
    });
    if (receipt && [401, 403].includes(receipt.status)) b.accessFailure();
  };
  private async observe<T>(
    b: ReportingBinding,
    request: (signal: AbortSignal) => Promise<T>,
    reading = false,
  ) {
    this.cancelPoll();
    const epoch = ++this.epoch;
    this.reading = reading;
    if (!reading) this.update({ busy: true });
    const observation = observeAsyncOperation(request, this.clock);
    this.cancelObservation = observation.cancel;
    const result = await observation.result;
    if (this.disposed || epoch !== this.epoch) return null;
    this.reading = false;
    if (!b.current()) {
      this.pause();
      return null;
    }
    this.update({ busy: false });
    return result;
  }
  private accept(job: ReportingJob, b: ReportingBinding, admission: boolean) {
    const previous = admission ? null : this.state.job;
    if (
      !validCommonJob(job, b.incidentId, previous?.job_id) ||
      job.submitted_by_user_id !== b.actorId ||
      (previous && !commonJobDoesNotRegress(previous, job))
    )
      return false;
    let snapshotId = this.state.snapshotId,
      releaseId = this.state.releaseId;
    if (job.status === "succeeded") {
      const kind = this.state.kind === "snapshot" ? "snapshot" : "release";
      const refs = job.result_summary?.resource_refs ?? [];
      if (
        job.result_summary?.code !== `${kind}_created` ||
        refs.length !== 1 ||
        refs[0]?.kind !== kind ||
        !uuid.test(refs[0].id) ||
        refs[0].route !== `/api/v1/${kind}s/${refs[0].id}`
      )
        return false;
      if (kind === "snapshot") {
        snapshotId = refs[0].id;
        releaseId = null;
      } else releaseId = refs[0].id;
    }
    this.update({
      job,
      snapshotId,
      releaseId,
      ...(terminalCommonJob(job) || job.status === "cancel_requested"
        ? { cancellation: null }
        : {}),
    });
    return true;
  }
  private beginPolling() {
    this.observingUntil = this.clock.now() + 120_000;
    this.schedulePoll();
  }
  private schedulePoll() {
    this.cancelPoll();
    if (
      !this.state.job ||
      terminalCommonJob(this.state.job) ||
      !this.authority(false)
    )
      return;
    if (this.clock.now() >= this.observingUntil) {
      this.update({ paused: true });
      return;
    }
    this.cancelPoll = this.clock.schedule(() => {
      void this.refresh();
    }, 500);
  }
}
