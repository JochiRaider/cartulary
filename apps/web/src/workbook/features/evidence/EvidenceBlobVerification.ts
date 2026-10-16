import { cartularyEvidenceVerificationContract as timing } from "@cartulary/ui-contracts";
import {
  boundedRead,
  browserObservationClock,
  observationDelay,
} from "../../../services/asyncObservation";
import {
  type CommonJobResource,
  commonJobDoesNotRegress,
  terminalCommonJob,
} from "../../../services/commonJobContract";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type {
  EvidenceFileAttempt,
  EvidenceFileOutcome,
  EvidenceFileTransport,
} from "./evidenceFileOperation";

/** Retains one verification admission and its Job independently of presentation. */
export class EvidenceBlobVerification {
  private attempt: EvidenceFileAttempt | null = null;
  private job: CommonJobResource | null = null;
  private controller: AbortController | null = null;
  private cancelController: AbortController | null = null;
  private cancelTxnId: string | null = null;
  private canceledLocally = false;
  private retired = false;
  private generation = 0;
  private busy = false;
  private failure: WorkbookOperationFailure | null = null;
  private phase:
    | "idle"
    | "admitting"
    | "uncertain"
    | "observing"
    | "interrupted"
    | "rejected"
    | "failed"
    | "canceled"
    | "verified" = "idle";
  constructor(
    private readonly transport: EvidenceFileTransport,
    private readonly ids: SecureTransactionIdPort,
    private readonly changed: () => void,
    private readonly recoverAuthority: () => void,
  ) {}
  get state() {
    return {
      phase: this.phase,
      job: this.job,
      blobId: this.attempt?.objectBlobId ?? null,
      pending: this.busy,
      failure: this.failure,
      canCancel: !!this.job?.cancelable && !this.cancelController,
      cancelUncertain:
        !!this.cancelTxnId &&
        !this.cancelController &&
        this.job?.status !== "cancel_requested" &&
        !this.terminal,
    };
  }
  get terminal() {
    return this.job !== null && terminalCommonJob(this.job);
  }
  private accept(outcome: EvidenceFileOutcome<CommonJobResource>) {
    if (outcome.kind === "rejected") {
      this.failure = outcome.failure;
      if (
        ["authentication_required", "authorization_lost"].includes(
          outcome.failure.kind,
        )
      )
        this.recoverAuthority();
      return false;
    }
    if (outcome.kind !== "accepted") return false;
    // A read started before cancellation may arrive after its acknowledgement.
    // Keep the newer accepted state, never regress it to running.
    if (this.job && !commonJobDoesNotRegress(this.job, outcome.receipt))
      return false;
    this.job = outcome.receipt;
    this.failure = null;
    this.phase =
      this.job.status === "succeeded"
        ? "verified"
        : this.job.status === "failed"
          ? "failed"
          : this.job.status === "canceled"
            ? "canceled"
            : "observing";
    this.changed();
    return true;
  }
  async verify(
    authority: WorkbookMutationAuthority,
    blobId: string,
    explicit: boolean,
  ) {
    if (
      this.busy ||
      this.retired ||
      authority.closed ||
      authority.role === "viewer"
    )
      return false;
    if (
      this.attempt &&
      (this.attempt.authority.actorId !== authority.actorId ||
        this.attempt.authority.incidentId !== authority.incidentId ||
        this.attempt.objectBlobId !== blobId)
    )
      return false;
    if (this.phase === "verified") return true;
    if (["failed", "canceled", "rejected"].includes(this.phase)) {
      if (!explicit) return false;
      this.attempt = null;
      this.job = null;
      this.cancelTxnId = null;
    }
    if (!this.attempt) {
      const clientTxnId = this.ids.create("evidence-verify");
      performance.mark?.(timing.start_mark, {
        detail: { transactionId: clientTxnId, blobId },
      });
      this.attempt = this.transport.capture({
        stage: "verify",
        authority,
        objectBlobId: blobId,
        clientTxnId,
      });
    }
    const attempt = this.attempt,
      generation = this.generation;
    const controller = new AbortController();
    this.controller = controller;
    this.busy = true;
    this.canceledLocally = false;
    try {
      if (!this.job) {
        const recovering = this.phase === "uncertain";
        this.phase = "admitting";
        this.changed();
        performance.mark?.(timing.dispatch_mark, {
          detail: {
            transactionId: attempt.clientTxnId,
            blobId,
            replay: recovering,
          },
        });
        const outcome = await boundedRead(
          (signal) => this.transport.verify(attempt, signal),
          controller.signal,
        );
        if (generation !== this.generation || controller.signal.aborted)
          return false;
        if (!this.accept(outcome)) {
          this.phase =
            outcome.kind === "rejected" && !recovering
              ? "rejected"
              : "uncertain";
          return false;
        }
        if (outcome.kind === "accepted")
          performance.mark?.(timing.admission_mark, {
            detail: {
              transactionId: attempt.clientTxnId,
              blobId,
              jobId: outcome.receipt.job_id,
            },
          });
      }
      while (this.job && !terminalCommonJob(this.job)) {
        this.phase = "observing";
        this.changed();
        await observationDelay(100, controller.signal, browserObservationClock);
        if (controller.signal.aborted) return false;
        const observedRevision = this.job;
        const outcome = await boundedRead(
          (signal) =>
            this.transport.readVerification(
              attempt,
              observedRevision.job_id,
              signal,
            ),
          controller.signal,
        );
        if (generation !== this.generation || controller.signal.aborted)
          return false;
        if (!this.accept(outcome)) {
          if (outcome.kind === "accepted" && this.job !== observedRevision)
            continue;
          this.phase = "interrupted";
          return false;
        }
      }
      return this.job?.status === "succeeded" && !this.canceledLocally;
    } catch {
      if (generation === this.generation)
        this.phase = this.job ? "interrupted" : "uncertain";
      return false;
    } finally {
      if (generation === this.generation) {
        this.controller = null;
        this.busy = false;
        this.changed();
      }
    }
  }
  async cancel() {
    if (
      !this.attempt ||
      !this.job ||
      this.terminal ||
      this.cancelController ||
      this.retired
    )
      return;
    this.cancelTxnId ??= this.ids.create("evidence-verify-cancel");
    const attempt = this.attempt,
      jobId = this.job.job_id,
      cancelTxnId = this.cancelTxnId;
    this.canceledLocally = true;
    const generation = this.generation,
      controller = new AbortController();
    this.cancelController = controller;
    this.changed();
    try {
      const outcome = await boundedRead(
        (signal) =>
          this.transport.cancelVerification(
            attempt,
            jobId,
            cancelTxnId,
            signal,
          ),
        controller.signal,
      );
      if (generation === this.generation && !controller.signal.aborted)
        this.accept(outcome);
    } catch {
      /* The exact cancellation identity is retained for explicit recovery. */
    } finally {
      if (generation === this.generation) {
        this.cancelController = null;
        this.changed();
      }
    }
  }
  suspend() {
    this.generation++;
    this.controller?.abort();
    this.cancelController?.abort();
    this.controller = null;
    this.cancelController = null;
    this.busy = false;
    if (!this.terminal && this.attempt)
      this.phase = this.job ? "interrupted" : "uncertain";
    this.changed();
  }
  retire() {
    this.retired = true;
    this.suspend();
  }
}
