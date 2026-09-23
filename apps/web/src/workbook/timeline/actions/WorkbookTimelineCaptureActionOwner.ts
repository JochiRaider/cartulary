import { observeAsyncOperation } from "../../../services/asyncObservation";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookTimelineActionRuntimePort } from "../../ports/WorkbookTimelineActionRuntimePort";
import type { TimelineCaptureReceipt } from "../adapters/timelineCaptureProtocol";
import type {
  TimelineCaptureBinding,
  TimelineCaptureOperation,
  TimelineRecordActionPort,
} from "../ports/TimelineRecordActionPort";
import type { TimelineCandidatePort } from "./TimelineCandidatePort";
import {
  type TimelineCaptureAuthority,
  type TimelineCaptureReview,
  timelineCaptureReviewValid,
} from "./timelineCaptureActionModel";

export type TimelineReconciliationScope = Readonly<{
  signal: AbortSignal;
  isCurrent: () => boolean;
}>;
const preparationFailure: WorkbookOperationFailure = {
  kind: "stale_target",
  message:
    "The Timeline row changed or earlier edits could not finish. Check the updated row and act again. No action was sent.",
};
type Snapshot = Readonly<{
  authority: TimelineCaptureAuthority | null;
  generation: number;
  revision: number;
  entries: readonly TimelineCaptureOperation[];
}>;

/** Timeline-only lifecycle attempts, retained by the incident/account workbook runtime. */
export class WorkbookTimelineCaptureActionOwner
  implements WorkbookTimelineActionRuntimePort, TimelineCandidatePort
{
  private authority: TimelineCaptureAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private sequence = 0;
  private revision = 0;
  private readonly entries = new Map<number, TimelineCaptureOperation>();
  private readonly versions = new Map<string, number>();
  private readonly listeners = new Set<() => void>();
  private readonly preparations = new Map<number, { cancel: () => void }>();
  private readonly sending = new Set<number>();
  private readonly remembered = new Set<string>();
  private readonly transports = new Map<number, number>();
  private port: TimelineRecordActionPort | null = null;
  private candidates: TimelineCandidatePort | null = null;
  private authorityUncertain: (() => void) | undefined;
  private reconcile:
    | ((
        receipt: TimelineCaptureReceipt,
        scope: TimelineReconciliationScope,
      ) => Promise<void>)
    | null = null;
  private snapshot: Snapshot = {
    authority: null,
    generation: 0,
    revision: 0,
    entries: [],
  };
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly accounting: {
      remember: (id: string) => void;
      settle: (id: string) => void;
      accepted: (recordId: string, version: number) => void;
    },
    private readonly observe: typeof observeAsyncOperation = observeAsyncOperation,
  ) {}
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  configure(
    port: TimelineRecordActionPort,
    candidates: TimelineCandidatePort,
    authorityUncertain?: () => void,
  ) {
    this.port = port;
    this.candidates = candidates;
    this.authorityUncertain = authorityUncertain;
  }
  registerReconciliation(reconcile: NonNullable<typeof this.reconcile>) {
    this.reconcile = reconcile;
    return () => {
      if (this.reconcile === reconcile) this.reconcile = null;
    };
  }
  setAuthority(authority: TimelineCaptureAuthority | null) {
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && authority.actorId !== this.actorId))
    )
      this.retire();
    this.generation++;
    for (const preparation of this.preparations.values()) preparation.cancel();
    this.authority =
      authority?.incidentId === this.incidentId && authority.role
        ? freeze({ ...authority })
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.publish();
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  suspendForAuthorityRecovery() {
    this.suspend();
    this.authorityUncertain?.();
  }
  retire() {
    this.generation++;
    for (const preparation of this.preparations.values()) preparation.cancel();
    for (const entry of this.entries.values())
      if (entry.attempt) this.settle(entry.attempt.id);
    this.preparations.clear();
    this.entries.clear();
    this.sending.clear();
    this.transports.clear();
    this.versions.clear();
    this.authority = null;
    this.actorId = null;
    this.reconcile = null;
    this.publish();
  }
  unavailableCause():
    | "session_missing"
    | "role_required"
    | "incident_closed"
    | null {
    if (!this.authority?.actorId || !this.authority.sessionIdentity)
      return "session_missing";
    if (!["reviewer", "admin"].includes(this.authority.role ?? ""))
      return "role_required";
    if (this.authority.closed) return "incident_closed";
    return null;
  }
  unavailableReason() {
    const cause = this.unavailableCause();
    return cause
      ? {
          session_missing: "Sign in to review Timeline rows.",
          role_required: "Reviewer or admin access is required.",
          incident_closed: "This incident is closed.",
        }[cause]
      : null;
  }
  canSubmit() {
    return this.unavailableReason() === null;
  }
  acceptVersion(id: string, version: number) {
    if (
      Number.isSafeInteger(version) &&
      version > (this.versions.get(id) ?? 0)
    ) {
      this.versions.set(id, version);
      this.publish();
    }
  }
  latestVersion = (id: string) => this.versions.get(id) ?? null;
  private blocks(entry: TimelineCaptureOperation) {
    return (
      ["preparing", "submitting", "uncertain"].includes(entry.phase) ||
      (entry.transportPending && entry.receipt === null)
    );
  }
  blocksRecord(id: string) {
    return [...this.entries.values()].some(
      (entry) => entry.review.target.recordId === id && this.blocks(entry),
    );
  }
  /** Admitted writes awaiting settlement; excludes acknowledged refresh reads. */
  get unsettledMutationCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        !entry.receipt &&
        (entry.phase === "preparing" ||
          entry.phase === "submitting" ||
          entry.phase === "uncertain"),
    ).length;
  }
  async page(cursor: string | null, signal: AbortSignal) {
    const generation = this.generation;
    if (!this.authority || !this.candidates)
      return { kind: "aborted" as const };
    const result = await this.candidates.page(cursor, signal);
    if (signal.aborted || generation !== this.generation)
      return { kind: "aborted" as const };
    if (
      result.kind === "rejected" &&
      workbookFailureLifecycle(result.failure).kind === "authority_unavailable"
    )
      this.suspendForAuthorityRecovery();
    if (result.kind === "accepted")
      for (const row of result.value.rows)
        this.acceptVersion(row.recordId, row.rowVersion);
    return result;
  }
  private current(
    review: TimelineCaptureReview,
    binding: TimelineCaptureBinding,
  ) {
    return (
      this.canSubmit() &&
      timelineCaptureReviewValid(review) &&
      review.authorityGeneration === this.generation &&
      JSON.stringify(review.authority) === JSON.stringify(this.authority) &&
      binding.isCurrent() &&
      binding.matchesReview() &&
      [
        review.target,
        ...(review.replacement ? [review.replacement] : []),
      ].every(
        (row) =>
          (this.latestVersion(row.recordId) ?? row.rowVersion) <=
          row.rowVersion,
      )
    );
  }
  private reserve(
    review: TimelineCaptureReview,
    binding: TimelineCaptureBinding,
  ) {
    if (
      !this.port ||
      !this.current(review, binding) ||
      this.blocksRecord(review.target.recordId)
    )
      return null;
    const key = ++this.sequence;
    this.entries.set(
      key,
      freeze({
        key,
        review: structuredClone(review),
        attempt: null,
        phase: "preparing",
        transportPending: false,
        receipt: null,
        failure: null,
        reconciliation: "pending",
      }),
    );
    this.publish();
    return key;
  }
  private async prepare(
    key: number,
    binding: TimelineCaptureBinding,
    signal?: AbortSignal,
  ) {
    const entry = this.entries.get(key);
    if (!entry) return false;
    const observation = this.observe(async (observedSignal) => {
      const row = await binding.prepare(observedSignal);
      if (entry.review.replacement) {
        const expected = entry.review.replacement;
        const cursors = new Set<string>();
        let cursor: string | null = null;
        let found = false;
        do {
          if (observedSignal.aborted) return false;
          const page = await this.page(cursor, observedSignal);
          if (page.kind !== "accepted") return false;
          const candidate = page.value.rows.find(
            (value) => value.recordId === expected.recordId,
          );
          if (candidate) {
            found =
              candidate.incidentId === expected.incidentId &&
              candidate.rowVersion === expected.rowVersion &&
              candidate.captureState === expected.captureState;
            break;
          }
          cursor = page.value.nextCursor;
          if (!page.value.hasMore || !cursor || cursors.has(cursor)) break;
          cursors.add(cursor);
        } while (!observedSignal.aborted);
        if (!found) return false;
      }
      if (row) this.acceptVersion(row.recordId, row.rowVersion);
      return (
        !observedSignal.aborted &&
        !signal?.aborted &&
        row?.recordId === entry.review.target.recordId &&
        row.rowVersion === entry.review.target.rowVersion &&
        row.captureState === entry.review.target.captureState &&
        this.current(entry.review, binding)
      );
    });
    const cancel = () => observation.cancel();
    this.preparations.set(key, observation);
    signal?.addEventListener("abort", cancel, { once: true });
    if (signal?.aborted) observation.cancel();
    const result = await observation.result;
    signal?.removeEventListener("abort", cancel);
    this.preparations.delete(key);
    if (this.entries.get(key) !== entry) return false;
    if (
      result.kind !== "completed" ||
      !result.value ||
      !this.current(entry.review, binding)
    ) {
      this.update(key, {
        phase: "preparation_failed",
        failure: preparationFailure,
      });
      return false;
    }
    return true;
  }
  /** Authoring preparation reserves the subject, but creates no transaction or request. */
  async prepareReview(
    review: TimelineCaptureReview,
    binding: TimelineCaptureBinding,
    signal: AbortSignal,
  ) {
    const key = this.reserve(review, binding);
    if (key === null) return false;
    const ready = await this.prepare(key, binding, signal);
    if (ready) {
      this.entries.delete(key);
      this.publish();
    }
    return ready;
  }
  /** Synchronous reservation closes the React double-activation window. */
  submit(review: TimelineCaptureReview, binding: TimelineCaptureBinding) {
    const key = this.reserve(review, binding);
    if (key === null) return false;
    void this.dispatchPrepared(key, binding);
    return true;
  }
  private async dispatchPrepared(key: number, binding: TimelineCaptureBinding) {
    if (!(await this.prepare(key, binding))) return;
    const entry = this.entries.get(key);
    if (!entry || !this.port || !this.current(entry.review, binding)) return;
    try {
      const attempt = freeze(
        this.port.capture(entry.review, this.ids.create("timeline-capture")),
      );
      if (
        [...this.entries.values()].some(
          (value) => value.attempt?.id === attempt.id,
        )
      )
        throw new Error("Duplicate transaction identity");
      this.update(key, { attempt, phase: "submitting" });
      await this.send(key, false);
    } catch {
      if (!this.entries.get(key)?.attempt)
        this.update(key, {
          phase: "preparation_failed",
          failure: {
            kind: "terminal",
            message:
              "A secure action identity could not be created. No action was sent.",
          },
        });
    }
  }
  async replay(key: number) {
    if (
      this.entries.get(key)?.phase === "uncertain" &&
      !this.sending.has(key) &&
      this.canSubmit()
    )
      await this.send(key, true);
  }
  private async send(key: number, replay: boolean) {
    const entry = this.entries.get(key),
      port = this.port;
    if (!entry?.attempt || !port || !this.canSubmit() || this.sending.has(key))
      return;
    const attempt = entry.attempt;
    const authorityGeneration = this.generation;
    const owns = () =>
      this.entries.get(key)?.attempt === attempt &&
      this.actorId === attempt.review.authority.actorId;
    this.sending.add(key);
    this.transports.set(key, (this.transports.get(key) ?? 0) + 1);
    this.remembered.add(attempt.id);
    this.accounting.remember(attempt.id);
    this.update(key, {
      phase: "submitting",
      transportPending: true,
      failure: null,
    });
    let uncertain = replay;
    const observation = this.observe(async (signal) => {
      const outcome = await port.send(attempt, signal);
      if (!owns() || this.entries.get(key)?.receipt) return;
      if (outcome.kind === "acknowledged") {
        const receipt = freeze(structuredClone(outcome.receipt));
        // Retain the complete operation-specific receipt before any projection work.
        this.update(key, {
          phase: "acknowledged",
          receipt,
          failure: null,
          reconciliation: "pending",
        });
        this.acceptVersion(receipt.data.record_id, receipt.data.row_version);
        this.accounting.accepted(
          receipt.data.record_id,
          receipt.data.row_version,
        );
        this.settle(attempt.id);
      } else if (outcome.kind === "rejected") {
        this.update(key, {
          phase: uncertain ? "uncertain" : "rejected",
          failure: outcome.failure,
        });
        if (authorityGeneration === this.generation) {
          if (outcome.failure.publicCode === "incident_closed")
            this.closeIncident();
          if (
            workbookFailureLifecycle(outcome.failure).kind ===
            "authority_unavailable"
          )
            this.suspendForAuthorityRecovery();
        }
      } else this.update(key, { phase: "uncertain" });
    });
    const settlement = observation.settled.then(async () => {
      if (!owns()) return;
      const remaining = Math.max(0, (this.transports.get(key) ?? 1) - 1);
      this.transports.set(key, remaining);
      this.update(key, { transportPending: remaining > 0 });
      if (remaining === 0) this.settle(attempt.id);
      if (this.entries.get(key)?.receipt) await this.refresh(key);
    });
    const result = await observation.result;
    if (!owns()) return;
    this.sending.delete(key);
    if (result.kind !== "completed") {
      uncertain = true;
      this.settle(attempt.id);
    }
    if (
      result.kind !== "completed" &&
      this.entries.get(key)?.phase === "submitting"
    )
      this.update(key, { phase: "uncertain" });
    if (result.kind === "completed") await settlement;
  }
  async refresh(key: number) {
    const entry = this.entries.get(key);
    if (
      !entry?.receipt ||
      !this.authority ||
      entry.reconciliation === "refreshing"
    )
      return;
    const generation = this.generation,
      receipt = entry.receipt;
    this.update(key, { reconciliation: "refreshing" });
    const observation = this.observe(async (signal) => {
      const isCurrent = () =>
        !signal.aborted &&
        this.generation === generation &&
        this.entries.get(key)?.receipt === receipt;
      const reconcile = this.reconcile;
      if (!reconcile || !isCurrent())
        throw new Error("Timeline reconciliation unavailable");
      await reconcile(receipt, { signal, isCurrent });
      if (!isCurrent() || reconcile !== this.reconcile)
        throw new Error("Timeline reconciliation detached");
    });
    const result = await observation.result;
    if (this.entries.get(key)?.receipt === receipt)
      this.update(key, {
        reconciliation:
          result.kind === "completed" && generation === this.generation
            ? "complete"
            : "required",
      });
  }
  dismiss(key: number) {
    const entry = this.entries.get(key);
    if (
      !entry ||
      this.blocks(entry) ||
      entry.reconciliation === "refreshing" ||
      (entry.receipt && entry.reconciliation !== "complete")
    )
      return;
    this.entries.delete(key);
    this.publish();
  }
  private settle(id: string) {
    if (this.remembered.delete(id)) this.accounting.settle(id);
  }
  private update(key: number, change: Partial<TimelineCaptureOperation>) {
    const entry = this.entries.get(key);
    if (entry) {
      this.entries.set(key, freeze({ ...entry, ...change }));
      this.publish();
    }
  }
  private publish() {
    this.snapshot = freeze({
      authority: this.authority,
      generation: this.generation,
      revision: ++this.revision,
      entries: this.authority ? [...this.entries.values()] : [],
    });
    for (const listener of this.listeners) listener();
  }
}
function freeze<T>(value: T): T {
  if (value && typeof value === "object") {
    for (const child of Object.values(value)) freeze(child);
    Object.freeze(value);
  }
  return value;
}
