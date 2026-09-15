import { observeAsyncOperation } from "../../../services/asyncObservation";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type {
  EvidenceFileAttempt,
  EvidenceFileReceipt,
  EvidenceFileTransport,
} from "./evidenceFileOperation";

/** Captured finalization and its complete receipt; callers own admission and review. */
export class EvidenceFileFinalization {
  private attempt: EvidenceFileAttempt | null = null;
  private receipt: EvidenceFileReceipt | null = null;
  private failure: WorkbookOperationFailure | null = null;
  private phase: "idle" | "submitting" | "uncertain" | "rejected" | "accepted" =
    "idle";
  private pending = false;
  private lifetime = 0;
  private dispatch = 0;
  constructor(
    private readonly transport: EvidenceFileTransport,
    private readonly changed: () => void,
    private readonly accepted: (
      receipt: EvidenceFileReceipt,
      attempt: EvidenceFileAttempt,
    ) => void,
  ) {}
  get state() {
    return {
      attempt: this.attempt,
      receipt: this.receipt,
      phase: this.phase,
      failure: this.failure,
      pending: this.pending,
    };
  }
  async send(attempt: EvidenceFileAttempt) {
    if (
      this.pending ||
      this.receipt ||
      (this.phase === "uncertain" && attempt !== this.attempt)
    )
      return false;
    const recovering = this.phase === "uncertain";
    this.attempt = attempt;
    this.phase = "submitting";
    this.failure = null;
    this.pending = true;
    const lifetime = this.lifetime,
      dispatch = ++this.dispatch;
    this.changed();
    const observation = observeAsyncOperation(async (signal) => {
      try {
        const outcome = await this.transport.finalize(attempt, signal);
        if (lifetime !== this.lifetime || this.receipt) return;
        if (outcome.kind === "accepted") {
          this.receipt = outcome.receipt;
          this.phase = "accepted";
          this.pending = false;
          // This checkpoint is durable in the browser owner before any presentation work.
          this.accepted(outcome.receipt, attempt);
        } else if (dispatch !== this.dispatch) return;
        else if (outcome.kind === "rejected") {
          this.failure = outcome.failure;
          this.phase = recovering ? "uncertain" : "rejected";
        } else this.phase = "uncertain";
      } catch {
        if (
          lifetime === this.lifetime &&
          dispatch === this.dispatch &&
          !this.receipt
        )
          this.phase = "uncertain";
      } finally {
        if (lifetime === this.lifetime && dispatch === this.dispatch) {
          this.pending = false;
          this.changed();
        }
      }
    });
    const result = await observation.result;
    if (
      lifetime === this.lifetime &&
      dispatch === this.dispatch &&
      result.kind !== "completed" &&
      !this.receipt
    ) {
      this.pending = false;
      this.phase = "uncertain";
      this.changed();
    }
    return this.receipt !== null;
  }
  resetRejected() {
    if (this.phase !== "rejected" || this.pending) return false;
    this.attempt = null;
    this.failure = null;
    this.phase = "idle";
    this.changed();
    return true;
  }
  retire() {
    this.lifetime++;
    this.attempt = null;
    this.receipt = null;
    this.failure = null;
    this.phase = "idle";
    this.pending = false;
  }
}
