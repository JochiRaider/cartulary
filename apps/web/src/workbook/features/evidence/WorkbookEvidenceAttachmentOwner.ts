import { evidenceViewSchemaId } from "@cartulary/view-contracts";
import { boundedRead } from "../../../services/asyncObservation";
import { readWorkbookAuthoringRecord } from "../../adapters/readWorkbookAuthoringRecord";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type {
  WorkbookAuthoringAuthorityReader,
  WorkbookAuthoringReadPort,
} from "../../ports/WorkbookAuthoringReadPort";
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

type Entry = {
  workId: string;
  recordId: string;
  filename: string;
  upload: EvidenceUploadSession;
  finalization: EvidenceFileFinalization;
  reviewedVersion: number;
  reviewCandidate: WorkbookQueryRow | null;
  needsReview: boolean;
  preparing: boolean;
  reserving: boolean;
  discarded: boolean;
  refresh: "none" | "required" | "refreshing" | "complete";
  message: string;
};
export type EvidenceAttachmentSnapshot = Readonly<{
  attention: EvidenceWorkAttention | null;
  recordId: string;
  filename: string;
  message: string;
  busy: boolean;
  accepted: boolean;
  reviewText: string | null;
  needsReview: boolean;
  canResume: boolean;
  canFreshSlot: boolean;
  canNewId: boolean;
  canDiscard: boolean;
  refreshRequired: boolean;
}>;

/** Existing Evidence finalization survives grid and inspector attachment. */
export class WorkbookEvidenceAttachmentOwner {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private workSequence = 0;
  private readonly entries = new Map<string, Entry>();
  private readonly versions = new Map<string, number>();
  private readonly listeners = new Set<() => void>();
  private readonly attachments = new Map<symbol, string>();
  private reader: WorkbookAuthoringReadPort | null = null;
  private authorityReader: WorkbookAuthoringAuthorityReader | null = null;
  private transport: EvidenceFileTransport | null = null;
  private snapshot: readonly EvidenceAttachmentSnapshot[] = [];
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly effects: {
      coordinate(
        recordId: string,
        signal: AbortSignal,
      ): Promise<WorkbookSourceWriteSettlement>;
      accepted(receipt: EvidenceFileReceipt, clientTxnId: string): void;
      refresh(): Promise<void>;
    },
  ) {}
  configure(
    reader: WorkbookAuthoringReadPort,
    authorityReader: WorkbookAuthoringAuthorityReader,
    transport: EvidenceFileTransport,
  ) {
    this.reader = reader;
    this.authorityReader = authorityReader;
    this.transport = transport;
  }
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  private canWrite() {
    return (
      !!this.authority &&
      !this.authority.closed &&
      this.authority.role !== "viewer"
    );
  }
  private publish = () => {
    this.snapshot = this.authority
      ? [...this.entries.values()].map((entry) => {
          const stage = entry.finalization.state;
          const busy =
            entry.preparing || entry.upload.status.pending || stage.pending;
          return {
            attention: this.attention(entry),
            recordId: entry.recordId,
            filename: entry.filename,
            message: this.message(entry),
            busy,
            accepted: !!stage.receipt,
            reviewText: entry.reviewCandidate
              ? `${String(entry.reviewCandidate.cells["evidence.title"]?.value ?? "Evidence")} · version ${entry.reviewCandidate.row_version}`
              : null,
            needsReview:
              entry.needsReview ||
              stage.phase === "rejected" ||
              entry.upload.status.phase === "slot_rejected",
            refreshRequired: entry.refresh === "required",
            canResume:
              this.canWrite() &&
              !busy &&
              (!entry.discarded || stage.phase === "uncertain") &&
              !stage.receipt,
            canFreshSlot:
              this.canWrite() &&
              !busy &&
              !entry.discarded &&
              this.freshAllowed(entry),
            canNewId:
              this.canWrite() &&
              !busy &&
              (stage.failure?.kind === "client_txn_conflict" ||
                entry.upload.status.failure?.kind === "client_txn_conflict"),
            canDiscard:
              !entry.discarded ||
              (!stage.pending && stage.phase !== "uncertain"),
          };
        })
      : [];
    for (const listener of this.listeners) listener();
  };
  private attention(entry: Entry): EvidenceWorkAttention | null {
    const stage = entry.finalization.state;
    const upload = entry.upload.status;
    const uncertain =
      stage.phase === "uncertain" ||
      upload.phase === "slot_uncertain" ||
      upload.phase === "transfer_uncertain";
    if (
      entry.refresh === "complete" ||
      (entry.discarded && !uncertain && !stage.pending && !stage.receipt)
    )
      return null;
    const category = stage.receipt
      ? "refresh"
      : uncertain
        ? "uncertain"
        : entry.preparing || upload.pending || stage.pending
          ? "in_progress"
          : entry.needsReview || stage.phase === "rejected"
            ? "review"
            : upload.failure
              ? "failure"
              : "draft";
    return {
      workId: entry.workId,
      category,
      label: this.message(entry),
      outcomeIdentity: `${entry.workId}:${upload.phase}:${stage.phase}:${entry.refresh}`,
    };
  }
  private message(entry: Entry) {
    if (entry.discarded && !entry.finalization.state.receipt)
      return "Stopped. The dispatched attachment remains retained until its outcome is known.";
    if (entry.message) return entry.message;
    if (entry.finalization.state.receipt)
      return entry.refresh === "complete"
        ? "File attached. Custody unchanged."
        : "File attached. Refresh pending.";
    if (entry.finalization.state.phase === "uncertain")
      return "Attachment acknowledgement is uncertain. Recover the same attachment.";
    if (entry.finalization.state.phase === "submitting")
      return "Attaching file.";
    if (entry.finalization.state.phase === "rejected")
      return "Attachment was rejected. Review the record before continuing.";
    if (entry.needsReview)
      return "Review the original Evidence record before attaching.";
    switch (entry.upload.status.phase) {
      case "slot_pending":
        return "Preparing upload.";
      case "slot_uncertain":
        return "Upload preparation is uncertain. Recover the same request.";
      case "slot_rejected":
        return "Upload preparation was rejected.";
      case "transferring":
        return "Uploading file.";
      case "transfer_uncertain":
        return "Upload acknowledgement is uncertain. Recover by finalizing the file.";
      case "fresh_required":
        return "This upload target cannot be used. Start a fresh upload.";
      default:
        return "File retained for attachment.";
    }
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId && this.actorId !== authority.actorId))
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
    const entries = [...this.entries.values()];
    this.entries.clear();
    this.attachments.clear();
    this.versions.clear();
    for (const entry of entries) {
      entry.upload.retire();
      entry.finalization.retire();
    }
    this.publish();
  }
  attach(token: symbol, recordId: string) {
    if (this.attachments.get(token) !== recordId) this.detach(token);
    this.attachments.set(token, recordId);
  }
  detach(token: symbol) {
    const recordId = this.attachments.get(token);
    this.attachments.delete(token);
    const entry = recordId ? this.entries.get(recordId) : null;
    if (entry && !entry.finalization.state.receipt) {
      entry.needsReview = true;
      entry.reviewCandidate = null;
      this.publish();
    }
  }
  observe(recordId: string, version: number) {
    this.versions.set(
      recordId,
      Math.max(version, this.versions.get(recordId) ?? 0),
    );
    const entry = this.entries.get(recordId);
    if (
      entry &&
      version > entry.reviewedVersion &&
      !entry.finalization.state.receipt
    ) {
      entry.needsReview = true;
      entry.reviewCandidate = null;
      this.publish();
    }
  }
  blocksRecord(recordId: string) {
    const entry = this.entries.get(recordId),
      stage = entry?.finalization.state;
    return (
      !!entry &&
      (entry.reserving || !!stage?.pending || stage?.phase === "uncertain")
    );
  }
  /** Save-state facts are separate from operation admission and refresh. */
  get unsettledMutationCount() {
    return [...this.entries.values()].filter(
      (entry) =>
        !entry.finalization.state.receipt &&
        (entry.preparing ||
          entry.upload.status.pending ||
          entry.upload.status.phase === "slot_uncertain" ||
          entry.upload.status.phase === "transfer_uncertain" ||
          entry.finalization.state.pending ||
          entry.finalization.state.phase === "uncertain"),
    ).length;
  }
  get pendingCount() {
    return [...this.entries.values()].filter(
      (e) =>
        e.preparing || e.upload.status.pending || e.finalization.state.pending,
    ).length;
  }
  get blockedCount() {
    return [...this.entries.values()].filter(
      (e) =>
        !e.finalization.state.receipt &&
        !e.upload.status.pending &&
        !e.preparing,
    ).length;
  }
  begin(
    row: WorkbookQueryRow,
    files: FileList | readonly File[],
  ): string | null {
    const admitted = admitEvidenceFile(files);
    if (admitted.kind !== "accepted")
      return admitted.kind === "rejected" ? admitted.message : null;
    if (!this.canWrite() || !this.transport)
      return "File attachment is currently unavailable.";
    const existing = this.entries.get(row.record_id);
    if (
      existing &&
      (!existing.finalization.state.receipt || existing.refresh !== "complete")
    )
      return "Resume or discard the retained attachment before choosing another file.";
    if (existing) this.discard(row.record_id);
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
      workId: `evidence-file:${this.incidentId}:${row.record_id}:${++this.workSequence}`,
      recordId: row.record_id,
      filename: admitted.file.name || "Attachment",
      upload,
      finalization: new EvidenceFileFinalization(
        transport,
        this.publish,
        (receipt, attempt) => {
          upload.finalized();
          entry.refresh = "required";
          entry.message = "";
          this.effects.accepted(receipt, attempt.clientTxnId);
          this.publish();
        },
      ),
      reviewedVersion: row.row_version,
      reviewCandidate: null,
      needsReview: false,
      preparing: false,
      reserving: false,
      discarded: false,
      refresh: "none",
      message: "",
    };
    this.entries.set(row.record_id, entry);
    this.publish();
    void this.resume(row.record_id, false);
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
  async resume(recordId: string, explicit = true) {
    const entry = this.entries.get(recordId);
    if (
      !entry ||
      entry.preparing ||
      entry.upload.status.pending ||
      entry.finalization.state.pending ||
      (entry.discarded && entry.finalization.state.phase !== "uncertain")
    )
      return;
    if (entry.finalization.state.receipt) {
      await this.refresh(recordId);
      return;
    }
    entry.preparing = true;
    entry.message = "";
    this.publish();
    try {
      if (
        !(await this.recheckAuthority()) ||
        !this.authority ||
        this.entries.get(recordId) !== entry
      )
        return;
      const prior = entry.finalization.state;
      if (prior.phase === "uncertain" && prior.attempt) {
        await entry.finalization.send(prior.attempt);
      } else {
        if (prior.phase === "rejected") return;
        if (
          !entry.upload.isFinalizable &&
          !(await entry.upload.prepare(this.authority))
        )
          return;
        if (!explicit && entry.upload.status.phase === "transfer_uncertain")
          return;
        if (entry.needsReview) return;
        const row = await this.verify(entry);
        if (!row || !this.authority || !this.transport || !entry.upload.blob)
          return;
        if (Date.now() >= Date.parse(entry.upload.blob.pending_expires_at)) {
          entry.message = "The pending upload expired. Start a fresh upload.";
          return;
        }
        const attempt = this.transport.capture({
          stage: "attach",
          authority: this.authority,
          clientTxnId: this.ids.create("evidence-attach"),
          recordId,
          baseRowVersion: row.row_version,
          objectBlobId: entry.upload.blob.object_blob_id,
        });
        await entry.finalization.send(attempt);
      }
      const failure = entry.finalization.state.failure;
      if (
        failure?.kind === "authentication_required" ||
        failure?.kind === "authorization_lost"
      )
        await this.recheckAuthority();
    } catch {
      entry.message =
        "Attachment is retained. Verify the original record before continuing.";
    } finally {
      entry.preparing = false;
      entry.reserving = false;
      this.publish();
    }
    if (entry.finalization.state.receipt && this.authority)
      await this.refresh(recordId);
  }
  private async verify(entry: Entry, review = false) {
    if (!this.reader) return null;
    const generation = this.generation,
      signal = new AbortController().signal;
    const settled = await boundedRead(
      (observed) => this.effects.coordinate(entry.recordId, observed),
      signal,
    );
    if (settled.kind !== "settled") {
      entry.message = "Resolve earlier Evidence edits before attaching.";
      return null;
    }
    if (!review) {
      entry.reserving = true;
      this.publish();
    }
    const row = await readWorkbookAuthoringRecord(
      this.reader,
      evidenceViewSchemaId,
      entry.recordId,
      signal,
    );
    if (
      generation !== this.generation ||
      !this.canWrite() ||
      this.entries.get(entry.recordId) !== entry
    )
      return null;
    if (
      !row ||
      row.row_version <
        Math.max(
          settled.minimumRowVersion,
          this.versions.get(entry.recordId) ?? 0,
        )
    ) {
      entry.message = "The original Evidence record could not be verified.";
      return null;
    }
    if (
      !review &&
      (entry.needsReview || row.row_version !== entry.reviewedVersion)
    ) {
      entry.needsReview = true;
      entry.reviewCandidate = null;
      return null;
    }
    return row;
  }
  async review(recordId: string) {
    const entry = this.entries.get(recordId);
    if (
      !entry ||
      entry.preparing ||
      entry.upload.status.pending ||
      entry.finalization.state.pending ||
      entry.finalization.state.phase === "uncertain"
    )
      return;
    entry.preparing = true;
    entry.reviewCandidate = null;
    entry.message = "";
    this.publish();
    try {
      if (!(await this.recheckAuthority())) return;
      const row = await this.verify(entry, true);
      if (!row) return;
      const failure = entry.finalization.state.failure;
      if (
        failure &&
        ![
          "stale_target",
          "same_field_conflict",
          "authentication_required",
          "authorization_lost",
        ].includes(failure.kind) &&
        failure.publicCode !== "row_version_conflict"
      )
        return;
      entry.reviewCandidate = row;
      entry.message =
        "Check this original Evidence record, then use the reviewed source.";
    } catch {
      entry.message = "The original Evidence record could not be verified.";
    } finally {
      entry.preparing = false;
      entry.reserving = false;
      this.publish();
    }
  }
  confirmReview(recordId: string) {
    const entry = this.entries.get(recordId);
    if (!entry?.reviewCandidate || !this.canWrite() || entry.preparing) return;
    if (
      (this.versions.get(recordId) ?? 0) > entry.reviewCandidate.row_version
    ) {
      entry.reviewCandidate = null;
      this.publish();
      return;
    }
    entry.finalization.resetRejected();
    entry.upload.reviewAuthorityRejection();
    entry.reviewedVersion = entry.reviewCandidate.row_version;
    entry.reviewCandidate = null;
    entry.needsReview = false;
    entry.message = "Original Evidence reviewed. Resume when ready.";
    this.publish();
  }
  private freshAllowed(entry: Entry) {
    const stage = entry.finalization.state;
    if (stage.receipt || stage.phase === "uncertain" || stage.pending)
      return false;
    return (
      entry.upload.status.phase === "fresh_required" ||
      (!!entry.upload.blob &&
        Date.now() >= Date.parse(entry.upload.blob.pending_expires_at)) ||
      (stage.phase === "rejected" &&
        ["blob_pending", "blob_failed", "accepted_contract_mismatch"].includes(
          stage.failure?.publicReason ?? "",
        ))
    );
  }
  freshSlot(recordId: string) {
    const entry = this.entries.get(recordId);
    if (
      !entry ||
      !this.canWrite() ||
      entry.preparing ||
      !this.freshAllowed(entry)
    )
      return;
    entry.finalization.resetRejected();
    entry.upload.freshSlot();
    entry.message = "";
    entry.needsReview = true;
    entry.reviewCandidate = null;
    this.publish();
  }
  newRequestId(recordId: string) {
    const entry = this.entries.get(recordId);
    if (!entry || !this.canWrite() || entry.preparing) return;
    if (entry.upload.retrySlotWithNewId()) {
      this.publish();
      return;
    }
    if (entry.finalization.state.failure?.kind !== "client_txn_conflict")
      return;
    entry.finalization.resetRejected();
    entry.needsReview = true;
    entry.reviewCandidate = null;
    this.publish();
  }
  discard(recordId: string) {
    const entry = this.entries.get(recordId);
    if (!entry) return;
    if (
      entry.finalization.state.pending ||
      entry.finalization.state.phase === "uncertain"
    ) {
      entry.discarded = true;
      this.publish();
      return;
    }
    this.entries.delete(recordId);
    entry.upload.retire();
    entry.finalization.retire();
    this.publish();
  }
  async refresh(recordId: string) {
    const entry = this.entries.get(recordId);
    if (
      !entry?.finalization.state.receipt ||
      !this.authority ||
      entry.refresh === "refreshing"
    )
      return;
    const generation = this.generation;
    entry.refresh = "refreshing";
    entry.message = "";
    this.publish();
    try {
      await this.effects.refresh();
      entry.refresh = generation === this.generation ? "complete" : "required";
    } catch {
      entry.refresh = "required";
    }
    this.publish();
  }
}
