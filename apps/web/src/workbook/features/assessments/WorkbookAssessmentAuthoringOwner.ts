import {
  assessmentsViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorLocalErrorFeedback,
  workbookInspectorMessageFeedback,
  workbookInspectorOperationFailureFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  type AssessmentCreateDraft,
  assessmentCreateErrors,
  followOnAssessmentDraft,
  initialAssessmentDraft,
} from "../../models/assessmentWorkbookModel";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import {
  type AssessmentAppendEntry,
  type AssessmentAppendOutcome,
  type AssessmentAppendReceipt,
  type AssessmentAppendTransport,
  type AssessmentAuthorityReader,
  type AssessmentDraft,
  type AssessmentPresentationBinding,
  freezeAssessment,
} from "./assessmentOperation";

const contract = requireViewContract(assessmentsViewSchemaId);
type Snapshot = Readonly<{
  authority: WorkbookMutationAuthority | null;
  draft: AssessmentDraft | null;
  entries: readonly AssessmentAppendEntry[];
  preparing: boolean;
  errors: Readonly<Record<string, string>>;
  feedback: WorkbookInspectorFeedback | null;
  candidateRevision: number;
  reviewRevision: number;
}>;

/** Assessment intent and results outlive every inspector and sheet attachment. */
export class WorkbookAssessmentAuthoringOwner {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private lifetime = 0;
  private generation = 0;
  private draft: AssessmentDraft | null = null;
  private preparing = false;
  private feedback: WorkbookInspectorFeedback | null = null;
  private errors: Readonly<Record<string, string>> = {};
  private candidateRevision = 0;
  private reviewRevision = 0;
  private entries = new Map<string, AssessmentAppendEntry>();
  private readonly versions = new Map<string, number>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  private readonly removals = new Map<string, number>();
  private readonly listeners = new Set<() => void>();
  private readonly refreshes = new Set<string>();
  private transport: AssessmentAppendTransport | null = null;
  private authorityReader: AssessmentAuthorityReader | null = null;
  private snapshot: Snapshot = {
    authority: null,
    draft: null,
    entries: [],
    preparing: false,
    errors: {},
    feedback: null,
    candidateRevision: 0,
    reviewRevision: 0,
  };
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly effects: {
      readonly accepted: (
        receipt: AssessmentAppendReceipt,
        clientTxnId: string,
      ) => void;
      readonly refresh: () => Promise<void>;
    },
    private readonly observe = observeAsyncOperation,
  ) {}
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  configure(
    transport: AssessmentAppendTransport,
    authorityReader: AssessmentAuthorityReader,
  ) {
    this.transport = transport;
    this.authorityReader = authorityReader;
  }
  private publish() {
    this.snapshot = {
      authority: this.authority,
      draft: this.authority ? this.draft : null,
      entries: this.authority ? [...this.entries.values()] : [],
      preparing: !!this.authority && this.preparing,
      errors: this.authority ? this.errors : {},
      feedback: this.authority ? this.feedback : null,
      candidateRevision: this.candidateRevision,
      reviewRevision: this.reviewRevision,
    };
    for (const listener of this.listeners) listener();
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && this.actorId !== authority.actorId))
    )
      this.retire();
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    this.authority =
      authority?.incidentId === this.incidentId
        ? freezeAssessment({ ...authority })
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.generation++;
    this.reviewRevision++;
    this.candidateRevision++;
    this.publish();
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  retire() {
    this.lifetime++;
    this.generation++;
    this.reviewRevision++;
    this.candidateRevision++;
    this.authority = null;
    this.actorId = null;
    this.draft = null;
    this.preparing = false;
    this.feedback = null;
    this.errors = {};
    this.entries.clear();
    this.refreshes.clear();
    this.versions.clear();
    this.rows.clear();
    this.removals.clear();
    this.publish();
  }
  canReplay() {
    return (
      !!this.authority?.sessionIdentity &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  canSubmit() {
    return this.canReplay() && this.authority?.closed === false;
  }
  get busy() {
    return (
      this.preparing ||
      [...this.entries.values()].some(
        (entry) =>
          entry.transportPending ||
          entry.phase === "uncertain" ||
          entry.phase === "submitting",
      )
    );
  }
  /** Admitted writes awaiting settlement; excludes acknowledged refresh reads. */
  get unsettledMutationCount() {
    return (
      Number(this.preparing) +
      [...this.entries.values()].filter(
        (entry) =>
          !entry.receipt &&
          (entry.phase === "submitting" || entry.phase === "uncertain"),
      ).length
    );
  }
  get pendingCount() {
    return (
      Number(this.preparing) +
      [...this.entries.values()].filter(
        (entry) => entry.transportPending || entry.refresh === "refreshing",
      ).length
    );
  }
  get blockedCount() {
    return [...this.entries.values()].filter(
      (entry) => entry.phase === "uncertain" || entry.refresh === "required",
    ).length;
  }
  openStandalone() {
    return this.open(null);
  }
  openFollowOn(row: WorkbookQueryRow | null) {
    if (
      !row ||
      row.row_version < (this.latestVersion(row.record_id) ?? 0) ||
      this.wasRemoved(row.record_id, row.row_version)
    ) {
      this.reject("Select a current assessment before creating a follow-on.");
      return false;
    }
    return this.open(row);
  }
  private open(row: WorkbookQueryRow | null) {
    if (!this.canSubmit()) {
      this.reject("Assessment creation requires an active editor role.");
      return false;
    }
    if (this.draft || this.busy) {
      this.feedback = workbookInspectorMessageFeedback(
        "Your unfinished assessment is retained. Resume it, or discard the editable draft before starting another.",
        "polite",
      );
      this.publish();
      return false;
    }
    const values = row
      ? followOnAssessmentDraft(contract, row)
      : initialAssessmentDraft(contract);
    if (!values) {
      this.reject("The selected assessment has no valid subject.");
      return false;
    }
    this.draft = freezeAssessment({
      values,
      mode: row ? "follow_on" : "standalone",
      origin: row
        ? { recordId: row.record_id, rowVersion: row.row_version }
        : null,
      revision: 0,
    });
    this.errors = {};
    this.feedback = null;
    this.publish();
    return true;
  }
  updateDraft(update: (draft: AssessmentCreateDraft) => AssessmentCreateDraft) {
    if (!this.draft || this.busy || !this.canSubmit()) return;
    this.draft = freezeAssessment({
      ...this.draft,
      revision: this.draft.revision + 1,
      values: structuredClone(update(structuredClone(this.draft.values))),
    });
    this.errors = {};
    this.feedback = null;
    this.publish();
  }
  discardDraft() {
    if (this.busy) return false;
    this.draft = null;
    this.errors = {};
    this.feedback = null;
    this.reviewRevision++;
    this.publish();
    return true;
  }
  reject(message: string) {
    this.feedback = workbookInspectorLocalErrorFeedback(message);
    this.publish();
  }
  invalidateReview() {
    this.reviewRevision++;
    this.publish();
  }
  observeCandidates(recordId?: string) {
    this.candidateRevision++;
    if (
      !recordId ||
      this.draft?.values.subjectRecordId === recordId ||
      this.draft?.values.supportRecordIds.includes(recordId)
    )
      this.reviewRevision++;
    this.publish();
  }
  async recheckAuthority() {
    const baseline = this.authority,
      lifetime = this.lifetime,
      generation = this.generation;
    const reader = this.authorityReader;
    if (!baseline || !reader) return;
    try {
      const current = await boundedRead(
        (signal) => reader(baseline, signal),
        new AbortController().signal,
      );
      if (lifetime === this.lifetime && generation === this.generation)
        this.setAuthority(current);
    } catch {
      /* Local failures do not prove incident access loss. The authority reader applies confirmed loss. */
    }
  }
  surfaceAttached() {
    for (const entry of this.entries.values())
      if (entry.receipt && entry.refresh === "required")
        void this.retryRefresh(entry.attempt.clientTxnId);
  }
  latestVersion(id: string) {
    return this.authority ? (this.versions.get(id) ?? null) : null;
  }
  latestRow(id: string) {
    const row = this.rows.get(id);
    return this.authority &&
      row &&
      row.row_version >= (this.versions.get(id) ?? 0)
      ? row
      : null;
  }
  wasRemoved(id: string, version: number) {
    return !!this.authority && (this.removals.get(id) ?? 0) >= version;
  }
  acceptVersion(id: string, version: number, removed = false) {
    if (!this.authority || version <= (this.versions.get(id) ?? 0)) return;
    this.versions.set(id, version);
    if (removed) {
      this.removals.set(id, version);
      this.rows.delete(id);
    }
    if (this.draft?.origin?.recordId === id) this.reviewRevision++;
    this.publish();
  }
  acceptRow(row: WorkbookQueryRow) {
    if (
      !this.authority ||
      row.row_version < (this.versions.get(row.record_id) ?? 0) ||
      this.wasRemoved(row.record_id, row.row_version)
    )
      return this.latestRow(row.record_id);
    const previous = this.rows.get(row.record_id);
    if (previous && previous.row_version >= row.row_version) return previous;
    if (
      this.draft?.origin?.recordId === row.record_id &&
      row.row_version >
        (this.versions.get(row.record_id) ?? this.draft.origin.rowVersion)
    )
      this.reviewRevision++;
    this.versions.set(row.record_id, row.row_version);
    this.removals.delete(row.record_id);
    const accepted = freezeAssessment(structuredClone(row));
    this.rows.set(row.record_id, accepted);
    this.publish();
    return accepted;
  }
  private async recheck(
    signal: AbortSignal,
    mode: "create" | "replay" | "read" = "create",
  ) {
    const reader = this.authorityReader;
    if (!this.authority || !reader)
      throw new Error(
        "Current authority is unavailable. Retry after recovery.",
      );
    const before = this.authority;
    const generation = this.generation;
    const current = await boundedRead(
      (currentSignal) => reader(before, currentSignal),
      signal,
    );
    if (signal.aborted || generation !== this.generation)
      throw new Error("Access changed. Review the retained assessment.");
    this.setAuthority(current);
    if (
      current.actorId !== before.actorId ||
      current.incidentId !== this.incidentId ||
      !(mode === "read"
        ? !!this.authority
        : mode === "replay"
          ? this.canReplay()
          : this.canSubmit())
    )
      throw new Error("Assessment creation requires current editor authority.");
  }
  async submit(binding: AssessmentPresentationBinding) {
    if (this.busy || !this.draft || !binding.isCurrent()) return;
    if (!this.canSubmit() || !this.transport) {
      this.reject("Assessment creation requires an active editor role.");
      return;
    }
    this.errors = assessmentCreateErrors(this.draft.values);
    if (Object.keys(this.errors).length) {
      this.reject(
        Object.values(this.errors)[0] ??
          "Complete the required assessment fields.",
      );
      return;
    }
    const draft = this.draft;
    const lifetime = this.lifetime;
    const reviewRevision = this.reviewRevision;
    this.preparing = true;
    this.feedback = null;
    this.publish();
    try {
      await this.recheck(new AbortController().signal);
      if (
        lifetime !== this.lifetime ||
        !binding.isCurrent() ||
        this.draft !== draft ||
        this.reviewRevision !== reviewRevision ||
        !this.authority
      )
        return;
      const attempt = this.transport.capture(
        { authority: this.authority, draft, sheetRef: binding.sheetRef },
        this.ids.create("assessment"),
      );
      this.entries.set(
        attempt.clientTxnId,
        freezeAssessment({
          attempt,
          phase: "submitting",
          transportPending: true,
          receipt: null,
          refresh: "none",
          message: null,
        }),
      );
      this.preparing = false;
      this.publish();
      await this.dispatch(attempt.clientTxnId, false);
    } catch (error) {
      if (lifetime === this.lifetime)
        this.reject(
          error instanceof Error
            ? error.message
            : "The assessment could not be prepared.",
        );
    } finally {
      if (lifetime === this.lifetime) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  private replace(id: string, patch: Partial<AssessmentAppendEntry>) {
    const entry = this.entries.get(id);
    if (!entry) return;
    this.entries.set(id, freezeAssessment({ ...entry, ...patch }));
    this.publish();
  }
  private receive(
    id: string,
    outcome: AssessmentAppendOutcome,
    replay: boolean,
  ) {
    const entry = this.entries.get(id);
    if (!entry || entry.receipt) return;
    if (outcome.kind === "accepted") {
      const receipt = freezeAssessment(structuredClone(outcome.receipt));
      this.replace(id, {
        phase: "accepted",
        receipt,
        refresh: "required",
        message: null,
      });
      if (
        this.draft === entry.attempt.review.draft ||
        this.draft?.revision === entry.attempt.review.draft.revision
      )
        this.draft = null;
      this.feedback = workbookInspectorMessageFeedback(
        "Assessment created.",
        "polite",
      );
      this.acceptRow(receipt.data.row);
      this.effects.accepted(receipt, id);
      this.publish();
      if (this.authority) void this.retryRefresh(id);
    } else if (
      outcome.kind === "rejected" &&
      !replay &&
      entry.phase !== "uncertain"
    ) {
      this.replace(id, { phase: "rejected", message: outcome.failure.message });
      this.feedback = workbookInspectorOperationFailureFeedback(
        outcome.failure,
      );
      this.publish();
      if (
        workbookFailureLifecycle(outcome.failure).kind ===
        "authority_unavailable"
      )
        void this.recheckAuthority();
    } else {
      this.replace(id, {
        phase: "uncertain",
        message:
          outcome.kind === "rejected"
            ? outcome.failure.message
            : "Recover this append before starting another judgment.",
      });
    }
  }
  private async dispatch(id: string, replay: boolean) {
    const entry = this.entries.get(id);
    const transport = this.transport;
    if (!entry || !transport) return;
    const lifetime = this.lifetime;
    const observation = this.observe(async (signal) => {
      const outcome = await transport.send(entry.attempt, signal);
      if (lifetime === this.lifetime) this.receive(id, outcome, replay);
      return outcome;
    });
    void observation.settled.then(() => {
      if (lifetime === this.lifetime)
        this.replace(id, { transportPending: false });
    });
    const result = await observation.result;
    if (
      lifetime === this.lifetime &&
      result.kind !== "completed" &&
      !this.entries.get(id)?.receipt
    )
      this.replace(id, {
        phase: "uncertain",
        message: "Recover this append before starting another judgment.",
      });
  }
  async replay(id: string) {
    const entry = this.entries.get(id);
    if (
      !entry ||
      entry.phase !== "uncertain" ||
      entry.transportPending ||
      this.preparing ||
      !this.canReplay()
    )
      return;
    const lifetime = this.lifetime;
    this.preparing = true;
    this.publish();
    try {
      await this.recheck(new AbortController().signal, "replay");
      if (
        lifetime !== this.lifetime ||
        this.entries.get(id)?.receipt ||
        this.authority?.actorId !== entry.attempt.review.authority.actorId
      )
        return;
      this.replace(id, {
        phase: "submitting",
        transportPending: true,
        message: null,
      });
      this.preparing = false;
      this.publish();
      await this.dispatch(id, true);
    } catch (error) {
      if (lifetime === this.lifetime)
        this.replace(id, {
          message:
            error instanceof Error
              ? error.message
              : "Authority could not be verified.",
        });
    } finally {
      if (lifetime === this.lifetime) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  async retryRefresh(id: string) {
    const entry = this.entries.get(id);
    if (
      !this.authority ||
      !entry?.receipt ||
      entry.refresh === "complete" ||
      this.refreshes.has(id)
    )
      return;
    const lifetime = this.lifetime;
    let generation = this.generation;
    this.refreshes.add(id);
    this.replace(id, { refresh: "refreshing" });
    try {
      await this.recheck(new AbortController().signal, "read");
      if (lifetime !== this.lifetime || !this.authority) return;
      generation = this.generation;
      await boundedRead(
        () => this.effects.refresh(),
        new AbortController().signal,
      );
      if (lifetime === this.lifetime && generation === this.generation)
        this.replace(id, { refresh: "complete", message: null });
      else if (lifetime === this.lifetime)
        this.replace(id, { refresh: "required" });
    } catch {
      if (lifetime === this.lifetime)
        this.replace(id, {
          refresh: "required",
          message:
            "The view could not be refreshed. Retry refresh; this sends no new assessment.",
        });
    } finally {
      if (lifetime === this.lifetime) this.refreshes.delete(id);
    }
  }
}
