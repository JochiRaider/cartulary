import { observeAsyncOperation } from "../../../services/asyncObservation";
import type { WorkbookProtocolBlobSlotReceipt } from "../../adapters/workbookProtocolTypes";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type {
  EvidenceFileAttempt,
  EvidenceFileTransport,
  EvidenceTransferAttempt,
} from "./evidenceFileOperation";

type Slot = Omit<WorkbookProtocolBlobSlotReceipt["data"], "upload_target">;
type EvidenceUploadPhase =
  | "selected"
  | "slot_pending"
  | "slot_uncertain"
  | "slot_rejected"
  | "ready"
  | "transferring"
  | "transfer_uncertain"
  | "transferred"
  | "fresh_required"
  | "finalized"
  | "discarded";

/** One file and one single-upload lease. Finalization policy belongs to its caller. */
export class EvidenceUploadSession {
  private file: File | null;
  private capability:
    | WorkbookProtocolBlobSlotReceipt["data"]["upload_target"]
    | null = null;
  private slotReceipt: {
    meta: WorkbookProtocolBlobSlotReceipt["meta"];
    data: Slot;
  } | null = null;
  private slotAttempt: EvidenceFileAttempt | null = null;
  private transferAttempt: EvidenceTransferAttempt | null = null;
  private transferAbort: AbortController | null = null;
  private lifetime = 0;
  private dispatch = 0;
  private retired = false;
  private issuedSession: string | null = null;
  private authority: WorkbookMutationAuthority | null = null;
  private phase: EvidenceUploadPhase = "selected";
  private failure: WorkbookOperationFailure | null = null;
  private pending = false;
  constructor(
    file: File,
    private readonly transport: EvidenceFileTransport,
    private readonly ids: SecureTransactionIdPort,
    private readonly changed: () => void,
    private readonly recoverAuthority: () => void,
    private readonly now = Date.now,
  ) {
    this.file = file;
  }
  get status() {
    return { phase: this.phase, pending: this.pending, failure: this.failure };
  }
  get filename() {
    return this.file?.name ?? "";
  }
  get blob() {
    return this.slotReceipt?.data ?? null;
  }
  get hasFile() {
    return this.file !== null;
  }
  get isFinalizable() {
    return this.phase === "transferred" || this.phase === "transfer_uncertain";
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    this.authority = authority;
    if (
      !authority ||
      authority.closed ||
      authority.role === "viewer" ||
      (this.issuedSession && this.issuedSession !== authority.sessionIdentity)
    ) {
      this.capability = null;
      if (this.phase === "ready") this.phase = "fresh_required";
      if (this.phase === "transferring") {
        this.phase = "transfer_uncertain";
        this.transferAbort?.abort();
      }
      this.changed();
    }
  }
  async prepare(authority: WorkbookMutationAuthority): Promise<boolean> {
    if (this.pending || this.retired || !this.file) return false;
    this.setAuthority(authority);
    if (this.phase === "transferred") return true;
    if (!["selected", "slot_uncertain", "ready"].includes(this.phase))
      return false;
    const lifetime = this.lifetime;
    if (!this.slotReceipt) {
      try {
        this.slotAttempt ??= this.transport.capture({
          stage: "slot",
          authority,
          clientTxnId: this.ids.create("evidence-slot"),
          file: this.file,
        });
      } catch {
        this.phase = "slot_rejected";
        this.changed();
        return false;
      }
      const attempt = this.slotAttempt,
        dispatch = ++this.dispatch,
        recovering = this.phase === "slot_uncertain";
      this.pending = true;
      this.phase = "slot_pending";
      this.changed();
      const observation = observeAsyncOperation(async (signal) => {
        try {
          const outcome = await this.transport.slot(attempt, signal);
          if (this.retired || lifetime !== this.lifetime || this.slotReceipt)
            return;
          if (outcome.kind === "accepted") {
            const { upload_target, ...data } = outcome.receipt.data;
            this.slotReceipt = { meta: outcome.receipt.meta, data };
            this.issuedSession = attempt.authority.sessionIdentity;
            this.capability =
              this.authority &&
              !this.authority.closed &&
              this.authority.role !== "viewer" &&
              this.authority.sessionIdentity === this.issuedSession
                ? upload_target
                : null;
            this.phase =
              this.capability && this.now() < Date.parse(data.target_expires_at)
                ? "ready"
                : "fresh_required";
            this.pending = false;
            this.changed();
          } else if (dispatch !== this.dispatch) return;
          else if (outcome.kind === "rejected") {
            this.failure = outcome.failure;
            this.phase = recovering ? "slot_uncertain" : "slot_rejected";
            if (
              outcome.failure.kind === "authentication_required" ||
              outcome.failure.kind === "authorization_lost"
            )
              this.recoverAuthority();
          } else this.phase = "slot_uncertain";
        } finally {
          if (
            !this.retired &&
            lifetime === this.lifetime &&
            dispatch === this.dispatch
          ) {
            this.pending = false;
            this.changed();
          }
        }
      });
      const result = await observation.result;
      if (
        result.kind !== "completed" &&
        dispatch === this.dispatch &&
        this.phase === "slot_pending"
      ) {
        this.pending = false;
        this.phase = "slot_uncertain";
        this.changed();
      }
      if (
        result.kind !== "completed" ||
        this.retired ||
        lifetime !== this.lifetime
      )
        return false;
    }
    if (
      this.phase !== "ready" ||
      !this.capability ||
      !this.slotReceipt ||
      !this.file ||
      this.authority?.sessionIdentity !== authority.sessionIdentity ||
      this.authority.closed ||
      this.authority.role === "viewer"
    )
      return false;
    if (this.now() >= Date.parse(this.slotReceipt.data.target_expires_at)) {
      this.capability = null;
      this.phase = "fresh_required";
      this.changed();
      return false;
    }
    this.transferAttempt = Object.freeze({
      authority: { ...authority },
      objectBlobId: this.slotReceipt.data.object_blob_id,
      target: this.capability,
    });
    const transfer = this.transferAttempt;
    this.capability = null; // The capability cannot be dispatched a second time.
    this.pending = true;
    this.phase = "transferring";
    const controller = new AbortController();
    this.transferAbort = controller;
    this.changed();
    try {
      const outcome = await this.transport.transfer(
        transfer,
        this.file,
        controller.signal,
      );
      if (this.retired || lifetime !== this.lifetime) return false;
      this.phase =
        outcome.kind === "accepted" ? "transferred" : "transfer_uncertain";
      if (outcome.kind === "not_dispatched") {
        // No request left the browser; keep this unused capability within its issuing session.
        this.capability =
          this.authority?.sessionIdentity === this.issuedSession
            ? transfer.target
            : null;
        this.phase = this.capability ? "ready" : "fresh_required";
        this.recoverAuthority();
      }
      if (outcome.kind === "authentication_required") this.recoverAuthority();
      return this.phase === "transferred";
    } catch {
      if (!this.retired && lifetime === this.lifetime)
        this.phase = "transfer_uncertain";
      return false;
    } finally {
      if (!this.retired && lifetime === this.lifetime) {
        this.pending = false;
        this.transferAttempt = null;
        this.transferAbort = null;
        this.changed();
      }
    }
  }
  /** Caller first proves no finalization remains uncertain or accepted. */
  freshSlot() {
    if (this.pending || !this.file || this.retired) return false;
    this.lifetime++;
    this.slotAttempt = null;
    this.slotReceipt = null;
    this.capability = null;
    this.failure = null;
    this.issuedSession = null;
    this.phase = "selected";
    this.changed();
    return true;
  }
  reviewAuthorityRejection() {
    if (
      this.phase !== "slot_rejected" ||
      !this.failure ||
      !["authentication_required", "authorization_lost"].includes(
        this.failure.kind,
      )
    )
      return false;
    return this.freshSlot();
  }
  retrySlotWithNewId() {
    if (
      this.phase !== "slot_rejected" ||
      this.failure?.kind !== "client_txn_conflict"
    )
      return false;
    return this.freshSlot();
  }
  finalized() {
    this.file = null;
    this.capability = null;
    this.transferAttempt = null;
    this.phase = "finalized";
    this.changed();
  }
  retire() {
    this.retired = true;
    this.lifetime++;
    this.transferAbort?.abort();
    this.transferAbort = null;
    this.file = null;
    this.capability = null;
    this.slotReceipt = null;
    this.slotAttempt = null;
    this.transferAttempt = null;
    this.authority = null;
    this.phase = "discarded";
    this.pending = false;
    this.changed();
  }
}
