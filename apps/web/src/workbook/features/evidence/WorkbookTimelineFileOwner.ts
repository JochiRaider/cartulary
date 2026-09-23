import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import { readWorkbookAuthoringRecord } from "../../adapters/readWorkbookAuthoringRecord";
import type { WorkbookProtocolCreateViewRowReceipt } from "../../adapters/workbookProtocolTypes";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type {
  WorkbookAuthoringAuthorityReader,
  WorkbookAuthoringReadPort,
} from "../../ports/WorkbookAuthoringReadPort";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { EvidenceFileFinalization } from "./EvidenceFileFinalization";
import { EvidenceUploadSession } from "./EvidenceUploadSession";
import {
  admitEvidenceFile,
  type EvidenceFileReceipt,
  type EvidenceFileTransport,
} from "./evidenceFileOperation";
import type { EvidenceWorkAttention } from "./evidenceWorkAttention";
import type {
  TimelineFileDraftPort,
  TimelineFileLinkAttempt,
  TimelineFileLinkTransport,
  TimelineFileSource,
} from "./timelineFileOperation";

type Entry = {
  workId: string;
  source: TimelineFileSource;
  filename: string;
  upload: EvidenceUploadSession;
  evidence: EvidenceFileFinalization;
  reviewedVersion: number | null;
  reviewCandidate: WorkbookQueryRow | "draft" | null;
  recordId: string | null;
  needsReview: boolean;
  preparing: boolean;
  reserving: boolean;
  link: TimelineFileLinkAttempt | null;
  linkDispatch: number;
  associationPresent: boolean;
  linkPending: boolean;
  linkUncertain: boolean;
  linkFailure: WorkbookOperationFailure | null;
  receipt:
    | WorkbookProtocolCreateViewRowReceipt
    | WorkbookPendingMutationAccepted
    | null;
  draftSubmitted: boolean;
  stopped: boolean;
  refresh: "none" | "required" | "refreshing" | "complete";
  message: string;
};
export type TimelineFileSnapshot = Readonly<{
  attention: EvidenceWorkAttention | null;
  key: string;
  recordId: string | null;
  reviewText: string | null;
  filename: string;
  sourceLabel: string;
  message: string;
  busy: boolean;
  needsReview: boolean;
  evidenceAccepted: boolean;
  attached: boolean;
  canResume: boolean;
  canFreshSlot: boolean;
  canNewId: boolean;
  canDiscard: boolean;
  refreshRequired: boolean;
}>;

/** File-backed Evidence and its original Timeline association have independent receipts. */
export class WorkbookTimelineFileOwner {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private workSequence = 0;
  private readonly entries = new Map<string, Entry>();
  private readonly versions = new Map<string, number>();
  private readonly listeners = new Set<() => void>();
  private reader: WorkbookAuthoringReadPort | null = null;
  private authorityReader: WorkbookAuthoringAuthorityReader | null = null;
  private transport: EvidenceFileTransport | null = null;
  private links: TimelineFileLinkTransport | null = null;
  private drafts: TimelineFileDraftPort | null = null;
  private unsubscribeDrafts: (() => void) | null = null;
  private presentation: string | null = null;
  private sourceCoordinator:
    | ((id: string, signal: AbortSignal) => Promise<boolean>)
    | null = null;
  private snapshot: readonly TimelineFileSnapshot[] = [];
  private admissionNotice: string | null = null;
  getAdmissionNotice = () => (this.authority ? this.admissionNotice : null);
  reportAdmission(message: string | null) {
    this.admissionNotice = message;
    this.publish();
    return message;
  }
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly effects: {
      coordinate(
        id: string,
        signal: AbortSignal,
      ): Promise<WorkbookSourceWriteSettlement>;
      accepted(
        receipt: EvidenceFileReceipt | WorkbookProtocolCreateViewRowReceipt,
        id: string,
      ): void;
      refresh(views: readonly string[]): Promise<void>;
    },
  ) {}
  configure(
    reader: WorkbookAuthoringReadPort,
    authorityReader: WorkbookAuthoringAuthorityReader,
    transport: EvidenceFileTransport,
    links: TimelineFileLinkTransport,
  ) {
    this.reader = reader;
    this.authorityReader = authorityReader;
    this.transport = transport;
    this.links = links;
  }
  configureDrafts(drafts: TimelineFileDraftPort) {
    if (this.drafts === drafts) return;
    this.unsubscribeDrafts?.();
    this.drafts = drafts;
    this.unsubscribeDrafts = drafts.subscribe(() => this.observeDrafts());
    this.observeDrafts();
  }
  registerSourceCoordinator(
    coordinate: (id: string, signal: AbortSignal) => Promise<boolean>,
  ) {
    this.sourceCoordinator = coordinate;
    return () => {
      if (this.sourceCoordinator === coordinate) this.sourceCoordinator = null;
    };
  }
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  attachmentDisabledReason() {
    return !this.authority
      ? "Current incident access is unavailable."
      : this.authority.closed
        ? "This incident is closed. Attachments are read-only."
        : this.authority.role === "viewer"
          ? "An editor role is required to attach files."
          : null;
  }
  private canWrite() {
    return (
      !!this.authority &&
      !this.authority.closed &&
      this.authority.role !== "viewer"
    );
  }
  private busy(entry: Entry) {
    return (
      entry.preparing ||
      entry.upload.status.pending ||
      entry.evidence.state.pending ||
      entry.linkPending
    );
  }
  private publish = () => {
    this.snapshot = this.authority
      ? [...this.entries.values()].map((e) => ({
          attention: this.attention(e),
          key: e.source.key,
          recordId: e.recordId,
          reviewText: !e.reviewCandidate
            ? null
            : e.reviewCandidate === "draft"
              ? "Original unsaved Timeline draft"
              : `${String(e.reviewCandidate.cells["timeline.activity_synopsis_text"]?.value ?? e.reviewCandidate.cells["timeline.raw_activity_text"]?.value ?? "Timeline row")} · version ${e.reviewCandidate.row_version}`,
          filename: e.filename,
          sourceLabel:
            e.source.label || e.recordId || "Original Timeline draft",
          message: this.message(e),
          busy: this.busy(e),
          needsReview:
            e.needsReview ||
            e.evidence.state.phase === "rejected" ||
            e.upload.status.phase === "slot_rejected",
          evidenceAccepted: !!e.evidence.state.receipt,
          attached: !!e.receipt || e.associationPresent,
          canResume:
            this.canWrite() &&
            !this.busy(e) &&
            !e.receipt &&
            !e.associationPresent,
          canFreshSlot:
            this.canWrite() && !this.busy(e) && this.freshAllowed(e),
          canNewId:
            this.canWrite() &&
            !this.busy(e) &&
            (e.evidence.state.failure?.kind === "client_txn_conflict" ||
              e.upload.status.failure?.kind === "client_txn_conflict" ||
              e.linkFailure?.kind === "client_txn_conflict"),
          canDiscard:
            !e.stopped ||
            (!this.busy(e) &&
              !e.linkUncertain &&
              e.evidence.state.phase !== "uncertain" &&
              !e.draftSubmitted),
          refreshRequired: e.refresh === "required",
        }))
      : [];
    for (const listener of this.listeners) listener();
  };
  private attention(e: Entry): EvidenceWorkAttention | null {
    const stage = e.evidence.state;
    const upload = e.upload.status;
    const uncertain =
      e.linkUncertain ||
      stage.phase === "uncertain" ||
      upload.phase === "slot_uncertain" ||
      upload.phase === "transfer_uncertain";
    if (
      e.refresh === "complete" ||
      (e.stopped &&
        !uncertain &&
        !this.busy(e) &&
        !stage.receipt &&
        !e.draftSubmitted)
    )
      return null;
    const category =
      e.receipt || e.associationPresent
        ? "refresh"
        : uncertain
          ? "uncertain"
          : this.busy(e) || e.draftSubmitted
            ? "in_progress"
            : e.needsReview || stage.phase === "rejected"
              ? "review"
              : e.linkFailure || upload.failure
                ? "failure"
                : "draft";
    const workId = e.workId;
    return {
      workId,
      category,
      label: this.message(e),
      outcomeIdentity: `${workId}:${upload.phase}:${stage.phase}:${e.linkUncertain}:${e.refresh}`,
    };
  }
  private message(e: Entry) {
    if (e.stopped && !e.receipt)
      return "Stopped. Dispatched work remains retained until its outcome is known; saved Evidence is preserved.";
    if (e.message) return e.message;
    if (e.receipt || e.associationPresent)
      return e.refresh === "complete"
        ? "Evidence attached."
        : "Evidence attached. Refresh pending.";
    if (e.linkUncertain)
      return "Timeline attachment is uncertain. Recover the same link.";
    if (e.linkPending) return "Attaching Evidence to Timeline.";
    if (e.linkFailure)
      return "Evidence saved. Timeline attachment needs recovery.";
    if (e.draftSubmitted)
      return "Evidence saved. Timeline creation is pending; resolve its retained save if needed.";
    if (e.evidence.state.receipt)
      return e.needsReview
        ? "Evidence saved. Review the original Timeline row before linking."
        : "Evidence saved. Resume Timeline attachment.";
    if (e.evidence.state.phase === "uncertain")
      return "Evidence creation is uncertain. Recover the same creation.";
    if (e.evidence.state.phase === "rejected")
      return "Evidence finalization was rejected. The file is retained.";
    if (e.evidence.state.pending) return "Finalizing Evidence.";
    switch (e.upload.status.phase) {
      case "slot_pending":
        return "Preparing upload.";
      case "slot_uncertain":
        return "Upload preparation is uncertain. Recover the same request.";
      case "slot_rejected":
        return "Upload preparation was rejected.";
      case "transferring":
        return "Uploading file. You can keep editing.";
      case "transfer_uncertain":
        return "Upload acknowledgement is uncertain. Recover by finalizing the file.";
      case "fresh_required":
        return "This upload target cannot be used. Start a fresh upload.";
      default:
        return "File retained for the original Timeline row.";
    }
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId && authority.actorId !== this.actorId))
    )
      this.retire();
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    this.generation++;
    this.authority =
      authority?.incidentId === this.incidentId ? authority : null;
    if (this.authority) this.actorId = this.authority.actorId;
    for (const entry of this.entries.values()) {
      entry.needsReview = true;
      entry.reviewCandidate = null;
      entry.upload.setAuthority(this.authority);
    }
    this.publish();
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  retire() {
    this.generation++;
    this.authority = null;
    this.actorId = null;
    this.presentation = null;
    this.admissionNotice = null;
    const entries = [...this.entries.values()];
    this.entries.clear();
    this.versions.clear();
    for (const entry of entries) {
      entry.upload.retire();
      entry.evidence.retire();
    }
    this.publish();
  }
  setPresentation(identity: string | null) {
    if (this.presentation === identity) return;
    this.presentation = identity;
    for (const entry of this.entries.values()) {
      if (!entry.receipt) entry.needsReview = true;
      entry.reviewCandidate = null;
    }
    this.publish();
  }
  acceptVersion(recordId: string, version: number) {
    this.versions.set(
      recordId,
      Math.max(version, this.versions.get(recordId) ?? 0),
    );
    for (const entry of this.entries.values()) {
      if (
        entry.recordId === recordId &&
        version > (entry.reviewedVersion ?? 0) &&
        !entry.receipt
      ) {
        entry.needsReview = true;
        entry.reviewCandidate = null;
      }
    }
    this.publish();
  }
  blocksRecord(recordId: string) {
    return [...this.entries.values()].some(
      (e) =>
        e.recordId === recordId &&
        (e.reserving || e.linkPending || e.linkUncertain),
    );
  }
  /** Save-state facts are separate from operation admission and refresh. */
  get unsettledMutationCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        !entry.receipt &&
        !entry.associationPresent &&
        (this.busy(entry) ||
          entry.linkUncertain ||
          entry.evidence.state.phase === "uncertain" ||
          entry.upload.status.phase === "slot_uncertain" ||
          entry.upload.status.phase === "transfer_uncertain"),
    ).length;
  }
  begin(source: TimelineFileSource, files: FileList | readonly File[]) {
    const admitted = admitEvidenceFile(files);
    if (admitted.kind !== "accepted")
      return this.reportAdmission(
        admitted.kind === "rejected" ? admitted.message : null,
      );
    if (!this.canWrite() || !this.transport)
      return this.reportAdmission("File attachment is currently unavailable.");
    const existing = this.entries.get(source.key);
    if (existing && (!existing.receipt || existing.refresh !== "complete"))
      return this.reportAdmission(
        "Resume or discard the retained file before choosing another.",
      );
    if (existing) this.discard(source.key);
    this.admissionNotice = null;
    const transport = this.transport;
    const upload = new EvidenceUploadSession(
      admitted.file,
      transport,
      this.ids,
      this.publish,
      () => {
        void this.recheckAuthority();
      },
    );
    const entry: Entry = {
      workId: `timeline-file:${this.incidentId}:${++this.workSequence}`,
      source: { ...source },
      recordId: source.recordId,
      reviewCandidate: null,
      reviewedVersion: source.rowVersion,
      filename: admitted.file.name.trim() || "Workbook attachment",
      upload,
      evidence: new EvidenceFileFinalization(
        transport,
        this.publish,
        (receipt, attempt) => {
          upload.finalized();
          this.effects.accepted(receipt, attempt.clientTxnId);
          this.publish();
        },
      ),
      needsReview: false,
      preparing: false,
      reserving: false,
      link: null,
      linkDispatch: 0,
      associationPresent: false,
      linkPending: false,
      linkUncertain: false,
      linkFailure: null,
      receipt: null,
      draftSubmitted: false,
      stopped: false,
      refresh: "none",
      message: "",
    };
    this.entries.set(source.key, entry);
    this.publish();
    void this.resume(source.key, false);
    return null;
  }
  async recheckAuthority() {
    const baseline = this.authority,
      reader = this.authorityReader,
      generation = this.generation;
    if (!baseline || !reader) return false;
    try {
      const current = await boundedRead(
        (signal) => reader(baseline, signal),
        new AbortController().signal,
      );
      if (generation !== this.generation) return false;
      this.setAuthority(current);
      return this.canWrite();
    } catch {
      if (generation === this.generation) this.suspend();
      return false;
    }
  }
  async resume(key: string, explicit = true) {
    const e = this.entries.get(key);
    if (!e || this.busy(e)) return;
    if (e.receipt || e.associationPresent) {
      await this.refresh(key);
      return;
    }
    e.preparing = true;
    e.message = "";
    this.publish();
    try {
      if (
        !(await this.recheckAuthority()) ||
        !this.authority ||
        this.entries.get(key) !== e ||
        !this.transport
      )
        return;
      const prior = e.evidence.state;
      if (!prior.receipt) {
        if (prior.phase === "uncertain" && prior.attempt)
          await e.evidence.send(prior.attempt);
        else {
          if (prior.phase === "rejected") return;
          if (
            !e.upload.isFinalizable &&
            !(await e.upload.prepare(this.authority))
          )
            return;
          if (!explicit && e.upload.status.phase === "transfer_uncertain")
            return;
          if (!this.canWrite() || !this.authority || !e.upload.blob) return;
          if (Date.now() >= Date.parse(e.upload.blob.pending_expires_at)) {
            e.message = "The pending upload expired. Start a fresh upload.";
            return;
          }
          await e.evidence.send(
            this.transport.capture({
              stage: "create",
              authority: this.authority,
              clientTxnId: this.ids.create("evidence-create"),
              objectBlobId: e.upload.blob.object_blob_id,
              fields: {
                "evidence.title": e.filename,
                "evidence.collector_party_text": "Workbook upload",
              },
            }),
          );
        }
      }
      if (!e.evidence.state.receipt) {
        const failure = e.evidence.state.failure;
        if (
          failure?.kind === "authentication_required" ||
          failure?.kind === "authorization_lost"
        )
          await this.recheckAuthority();
        return;
      }
      if (e.linkUncertain && e.link) {
        await this.sendLink(e, e.link);
      } else {
        if (e.stopped || e.linkFailure || e.needsReview) return;
        await this.link(e);
      }
    } catch {
      e.message =
        "The file and accepted progress are retained. Review the original Timeline row.";
    } finally {
      e.preparing = false;
      e.reserving = false;
      this.publish();
    }
    if (e.receipt && this.authority) await this.refresh(key);
  }
  private observeDrafts() {
    for (const e of this.entries.values()) {
      if (e.source.recordId || e.receipt) continue;
      const state = this.drafts?.resolve(e.source.key);
      if (state?.kind !== "promoted") {
        if (e.draftSubmitted && state?.kind === "draft") {
          e.draftSubmitted = false;
          e.needsReview = true;
        }
        continue;
      }
      e.recordId = state.receipt.row.record_id;
      e.reviewedVersion ??= state.receipt.row.row_version;
      if (
        e.draftSubmitted &&
        this.hasEvidence(
          state.receipt.row,
          e.evidence.state.receipt?.data.row.record_id ?? "",
        )
      ) {
        e.receipt = state.receipt;
        e.draftSubmitted = false;
        e.refresh = "required";
        if (this.authority) void this.refresh(e.source.key);
      }
    }
    this.publish();
  }
  private hasEvidence(row: WorkbookQueryRow, evidenceId: string) {
    const value = row.cells["timeline.attached_evidence_ids"]?.value;
    return (
      !!value &&
      typeof value === "object" &&
      "items" in value &&
      Array.isArray(value.items) &&
      value.items.some(
        (item) =>
          item &&
          typeof item === "object" &&
          (item.linked_record_id === evidenceId ||
            item.record_id === evidenceId),
      )
    );
  }
  private async source(e: Entry, review = false) {
    this.observeDrafts();
    if (!e.recordId || !this.reader) return null;
    const recordId = e.recordId;
    const coordinator = this.sourceCoordinator;
    const generation = this.generation,
      signal = new AbortController().signal;
    if (
      coordinator &&
      !(await boundedRead(
        (observed) => coordinator(recordId, observed),
        signal,
      ))
    ) {
      e.message = "Resolve earlier Timeline edits before attaching Evidence.";
      return null;
    }
    const settled = await boundedRead(
      (observed) => this.effects.coordinate(recordId, observed),
      signal,
    );
    if (settled.kind !== "settled") {
      e.message = "Resolve earlier Timeline saves before attaching Evidence.";
      return null;
    }
    if (!review) {
      e.reserving = true;
      this.publish();
    }
    const row = await readWorkbookAuthoringRecord(
      this.reader,
      timelineViewSchemaId,
      e.recordId,
      signal,
    );
    if (
      generation !== this.generation ||
      !this.canWrite() ||
      this.entries.get(e.source.key) !== e
    )
      return null;
    if (
      !row ||
      row.row_version <
        Math.max(
          settled.minimumRowVersion,
          this.versions.get(e.recordId) ?? 0,
        ) ||
      row.cells["timeline.capture_state"]?.value === "superseded"
    ) {
      e.message =
        "The original Timeline record could not be verified. Evidence remains saved.";
      return null;
    }
    if (!review && (e.needsReview || row.row_version !== e.reviewedVersion)) {
      e.needsReview = true;
      return null;
    }
    return row;
  }
  private async link(e: Entry) {
    const evidenceId = e.evidence.state.receipt?.data.row.record_id;
    if (!evidenceId || !this.reader || !this.authority || !this.links) return;
    const generation = this.generation;
    const target = await readWorkbookAuthoringRecord(
      this.reader,
      evidenceViewSchemaId,
      evidenceId,
      new AbortController().signal,
    );
    if (
      !target ||
      generation !== this.generation ||
      !this.canWrite() ||
      this.entries.get(e.source.key) !== e ||
      e.needsReview
    ) {
      e.message = "Saved Evidence could not be verified. Retry review.";
      return;
    }
    this.observeDrafts();
    if (!e.recordId) {
      const state = this.drafts?.resolve(e.source.key);
      if (state?.kind === "draft" && !e.draftSubmitted) {
        e.draftSubmitted = true;
        this.drafts?.attachEvidence(e.source.key, evidenceId);
        this.observeDrafts();
      } else if (state?.kind === "unavailable" || !state)
        e.message =
          "The original Timeline draft is unavailable. Evidence remains saved.";
      return;
    }
    const row = await this.source(e);
    if (!row || !this.authority || e.needsReview) return;
    // An existing association is a source fact, not an invented mutation receipt.
    if (this.hasEvidence(row, evidenceId)) {
      e.associationPresent = true;
      e.refresh = "required";
      e.message = "Evidence is already linked to the original Timeline row.";
      return;
    }
    const attempt = this.links.capture(
      this.authority,
      row,
      evidenceId,
      this.ids.create("timeline-evidence-link"),
    );
    await this.sendLink(e, attempt);
  }
  private async sendLink(e: Entry, attempt: TimelineFileLinkAttempt) {
    if (
      !this.links ||
      e.linkPending ||
      e.receipt ||
      (e.linkUncertain && e.link !== attempt)
    )
      return;
    const recovering = e.linkUncertain;
    e.link = attempt;
    e.linkPending = true;
    e.linkFailure = null;
    e.linkUncertain = false;
    this.publish();
    const links = this.links,
      dispatch = ++e.linkDispatch;
    const observation = observeAsyncOperation(async (signal) => {
      try {
        const outcome = await links.send(attempt, signal);
        if (this.entries.get(e.source.key) !== e || e.receipt) return;
        if (outcome.kind === "accepted") {
          e.receipt = outcome.receipt;
          e.linkPending = false;
          e.refresh = "required";
          e.linkUncertain = false;
          this.effects.accepted(outcome.receipt, attempt.clientTxnId);
        } else if (dispatch !== e.linkDispatch) return;
        else if (outcome.kind === "rejected") {
          e.linkFailure = outcome.failure;
          e.linkUncertain = recovering;
          e.needsReview = true;
        } else e.linkUncertain = true;
      } catch {
        if (dispatch === e.linkDispatch && !e.receipt) e.linkUncertain = true;
      } finally {
        if (dispatch === e.linkDispatch) e.linkPending = false;
        this.publish();
      }
    });
    const result = await observation.result;
    if (
      result.kind !== "completed" &&
      dispatch === e.linkDispatch &&
      !e.receipt
    ) {
      e.linkPending = false;
      e.linkUncertain = true;
    }
    this.publish();
  }
  async review(key: string) {
    const e = this.entries.get(key);
    if (
      !e ||
      this.busy(e) ||
      e.linkUncertain ||
      e.evidence.state.phase === "uncertain"
    )
      return;
    e.preparing = true;
    e.reviewCandidate = null;
    e.message = "";
    this.publish();
    try {
      if (!(await this.recheckAuthority())) return;
      if (
        e.linkFailure &&
        ![
          "stale_target",
          "same_field_conflict",
          "authentication_required",
          "authorization_lost",
        ].includes(e.linkFailure.kind) &&
        e.linkFailure.publicCode !== "row_version_conflict"
      )
        return;
      this.observeDrafts();
      if (!e.recordId) {
        if (this.drafts?.resolve(e.source.key).kind !== "draft") return;
        e.reviewCandidate = "draft";
      } else {
        const row = await this.source(e, true);
        if (!row) return;
        e.reviewCandidate = row;
      }
      e.message =
        "Check the original Timeline source, then use the reviewed source.";
    } catch {
      e.message = "The original Timeline source could not be verified.";
    } finally {
      e.preparing = false;
      this.publish();
    }
  }
  confirmReview(key: string) {
    const e = this.entries.get(key);
    if (!e?.reviewCandidate || !this.canWrite() || this.busy(e)) return;
    if (e.reviewCandidate !== "draft") {
      if (
        (this.versions.get(e.reviewCandidate.record_id) ?? 0) >
        e.reviewCandidate.row_version
      ) {
        e.reviewCandidate = null;
        this.publish();
        return;
      }
      e.reviewedVersion = e.reviewCandidate.row_version;
    }
    e.reviewCandidate = null;
    e.upload.reviewAuthorityRejection();
    if (
      e.evidence.state.failure &&
      ["authentication_required", "authorization_lost"].includes(
        e.evidence.state.failure.kind,
      )
    )
      e.evidence.resetRejected();
    e.link = null;
    e.linkFailure = null;
    e.needsReview = false;
    e.stopped = false;
    e.message = "Original Timeline source reviewed. Resume when ready.";
    this.publish();
  }
  private freshAllowed(e: Entry) {
    if (e.evidence.state.receipt || e.evidence.state.phase === "uncertain")
      return false;
    return (
      e.upload.status.phase === "fresh_required" ||
      (!!e.upload.blob &&
        Date.now() >= Date.parse(e.upload.blob.pending_expires_at)) ||
      ["blob_pending", "blob_failed", "accepted_contract_mismatch"].includes(
        e.evidence.state.failure?.publicReason ?? "",
      )
    );
  }
  freshSlot(key: string) {
    const e = this.entries.get(key);
    if (!e || this.busy(e) || !this.canWrite() || !this.freshAllowed(e)) return;
    e.evidence.resetRejected();
    e.upload.freshSlot();
    e.message = "";
    this.publish();
  }
  newRequestId(key: string) {
    const e = this.entries.get(key);
    if (!e || this.busy(e) || !this.canWrite()) return;
    if (e.upload.retrySlotWithNewId()) return;
    if (e.evidence.state.failure?.kind === "client_txn_conflict")
      e.evidence.resetRejected();
    if (e.linkFailure?.kind === "client_txn_conflict") {
      e.link = null;
      e.linkFailure = null;
      e.needsReview = true;
    }
    this.publish();
  }
  discard(key: string) {
    const e = this.entries.get(key);
    if (!e) return;
    if (
      e.linkPending ||
      e.evidence.state.pending ||
      e.linkUncertain ||
      e.evidence.state.phase === "uncertain" ||
      e.draftSubmitted
    ) {
      e.stopped = true;
      e.needsReview = true;
      e.reviewCandidate = null;
      this.publish();
      return;
    }
    this.entries.delete(key);
    e.upload.retire();
    e.evidence.retire();
    this.publish();
  }
  async refresh(key: string) {
    const e = this.entries.get(key);
    if (
      (!e?.receipt && !e?.associationPresent) ||
      !this.authority ||
      e.refresh === "refreshing"
    )
      return;
    const generation = this.generation;
    e.refresh = "refreshing";
    e.message = "";
    this.publish();
    try {
      await this.effects.refresh([evidenceViewSchemaId, timelineViewSchemaId]);
      e.refresh = generation === this.generation ? "complete" : "required";
    } catch {
      e.refresh = "required";
    }
    this.publish();
  }
}
