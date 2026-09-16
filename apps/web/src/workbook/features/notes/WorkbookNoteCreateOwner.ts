import {
  getViewContract,
  type InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import type { SheetRef } from "../../../shared/sheetRef";
import { readWorkbookAuthoringRecord } from "../../adapters/readWorkbookAuthoringRecord";
import { acceptedRecordMutation } from "../../adapters/workbookRecordPatchTransport";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookAuthoringAuthorityReader } from "../../ports/WorkbookAuthoringReadPort";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import { freezeWorkbookValue } from "../../utils/freezeWorkbookValue";
import {
  type NoteCreateReader,
  type NoteDraft,
  type NoteSource,
  noteCreateView,
  noteFeature,
  noteSourceFromSubject,
  noteSourceViews,
  prepareNote,
} from "./noteCreateModel";
import type {
  NoteEntry,
  NoteOutcome,
  NoteReceipt,
  NoteReview,
  NoteTransport,
} from "./noteCreateOperation";

type Snapshot = Readonly<{
  authority: WorkbookMutationAuthority | null;
  draft: NoteDraft | null;
  attachment: symbol | null;
  errors: Readonly<Record<string, string>>;
  message: string | null;
  candidateRevision: number;
  needsReview: boolean;
  preparing: boolean;
  entries: readonly NoteEntry[];
}>;
/** One incident-lifetime Note author. Attachments never own its source or text. */
export class WorkbookNoteCreateOwner {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private lifetime = 0;
  private generation = 0;
  private authorityReader: WorkbookAuthoringAuthorityReader | null = null;
  private transport: NoteTransport | null = null;
  private readonly entries = new Map<string, NoteEntry>();
  private readonly dispatches = new Map<string, number>();
  private readonly refreshes = new Set<string>();
  private readonly notified = new Set<string>();
  private readonly removed = new Map<string, number>();
  private sourceCoordinator:
    | ((recordId: string, signal: AbortSignal) => Promise<boolean>)
    | null = null;
  private draft: NoteDraft | null = null;
  private attachment: symbol | null = null;
  private nextDraft = 0;
  private preparing = false;
  private candidateRevision = 0;
  private reviewRevision = 0;
  private reviewedRevision = 0;
  private errors: Readonly<Record<string, string>> = {};
  private message: string | null = null;
  private reader: NoteCreateReader | null = null;
  private readonly versions = new Map<string, number>();
  private readonly listeners = new Set<() => void>();
  private snapshot: Snapshot = {
    authority: null,
    draft: null,
    attachment: null,
    errors: {},
    message: null,
    candidateRevision: 0,
    needsReview: false,
    preparing: false,
    entries: [],
  };
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly effects?: {
      coordinate(
        source: NoteSource,
        signal: AbortSignal,
      ): Promise<WorkbookSourceWriteSettlement>;
      accepted(receipt: NoteReceipt, id: string): void;
      refresh(
        views: readonly string[],
        records: readonly string[],
      ): Promise<void>;
    },
    private readonly observeTransport = observeAsyncOperation,
  ) {}
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  configure(
    reader: NoteCreateReader,
    authorityReader: WorkbookAuthoringAuthorityReader,
    transport?: NoteTransport,
  ) {
    this.reader = reader;
    this.authorityReader = authorityReader;
    if (transport) this.transport = transport;
  }
  getReader() {
    return this.reader;
  }
  private publish() {
    this.snapshot = {
      authority: this.authority,
      draft: this.authority ? this.draft : null,
      attachment: this.authority ? this.attachment : null,
      errors: this.authority ? this.errors : {},
      message: this.authority ? this.message : null,
      candidateRevision: this.candidateRevision,
      needsReview: this.reviewRevision !== this.reviewedRevision,
      preparing: !!this.authority && this.preparing,
      entries: this.authority ? [...this.entries.values()] : [],
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
        ? freezeWorkbookValue({ ...authority })
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.generation++;
    this.reviewRevision++;
    this.candidateRevision++;
    if (this.draft?.source)
      this.draft = freezeWorkbookValue({
        ...this.draft,
        source: { ...this.draft.source, label: "" },
      });
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
    this.entries.clear();
    this.dispatches.clear();
    this.refreshes.clear();
    this.notified.clear();
    this.removed.clear();
    this.authority = null;
    this.actorId = null;
    this.draft = null;
    this.attachment = null;
    this.errors = {};
    this.message = null;
    this.versions.clear();
    this.preparing = false;
    this.reviewRevision++;
    this.candidateRevision++;
    this.publish();
  }
  canSubmit() {
    return (
      !!this.authority?.sessionIdentity &&
      !this.authority.closed &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  get busy() {
    return (
      this.preparing ||
      [...this.entries.values()].some(
        (entry) => entry.phase === "submitting" || entry.phase === "uncertain",
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
    return (
      this.uncertainCount +
      [...this.entries.values()].filter((entry) => entry.refresh === "required")
        .length
    );
  }
  get uncertainCount() {
    return [...this.entries.values()].filter(
      (entry) => entry.phase === "uncertain",
    ).length;
  }
  canReplay() {
    return (
      !!this.authority?.sessionIdentity &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  begin(
    subject: WorkbookInspectorLiveRowBinding,
    feature: InspectorFeatureGroup,
    sheetRef: SheetRef,
    attachment: symbol,
  ) {
    const canonical = noteFeature(subject.subject.viewSchemaId);
    if (
      !canonical ||
      canonical !== feature ||
      !subject.subject.recordId ||
      subject.subject.rowVersion < 1
    )
      return false;
    return this.start(
      { sheetRef, subject: subject.subject, feature: canonical },
      noteSourceFromSubject(subject),
      attachment,
    );
  }
  beginSheet(sheetRef: SheetRef, attachment: symbol) {
    return this.start(
      { sheetRef, subject: null, feature: null },
      null,
      attachment,
    );
  }
  private start(
    origin: NoteDraft["origin"],
    source: NoteSource | null,
    attachment: symbol,
  ) {
    if (!this.canSubmit() || !this.authority) return false;
    if (this.draft || this.busy) {
      this.message =
        "Your Note draft is retained. Resume it, or explicitly discard it before starting another.";
      this.publish();
      return false;
    }
    this.draft = freezeWorkbookValue({
      id: ++this.nextDraft,
      revision: 0,
      actorId: this.authority.actorId,
      incidentId: this.incidentId,
      values: {},
      source,
      origin: structuredClone(origin),
    });
    this.attachment = attachment;
    this.errors = {};
    this.message = null;
    this.reviewedRevision = this.reviewRevision;
    if (source)
      this.versions.set(
        source.recordId,
        Math.max(source.rowVersion, this.versions.get(source.recordId) ?? 0),
      );
    this.publish();
    return true;
  }
  update(field: string, value: string) {
    if (
      !this.draft ||
      this.busy ||
      !this.canSubmit() ||
      !["note.title", "note.body", "note.tags"].includes(field)
    )
      return;
    this.draft = freezeWorkbookValue({
      ...this.draft,
      revision: this.draft.revision + 1,
      values: { ...this.draft.values, [field]: value },
    });
    this.errors = {};
    this.message = null;
    this.publish();
  }
  changeSource(source: NoteSource | null) {
    if (!this.draft || this.busy || !this.canSubmit()) return;
    if (
      source &&
      (!noteSourceViews.includes(source.viewSchemaId) ||
        !source.recordId ||
        source.rowVersion < 1)
    )
      return;
    this.draft = freezeWorkbookValue({
      ...this.draft,
      revision: this.draft.revision + 1,
      source: source ? { ...source } : null,
    });
    this.reviewRevision++;
    this.reviewedRevision = this.reviewRevision;
    this.errors = {};
    this.message = null;
    this.publish();
  }
  resume(attachment: symbol) {
    if (this.draft && this.authority && this.attachment !== attachment) {
      this.attachment = attachment;
      this.publish();
    }
  }
  detach(attachment: symbol) {
    if (this.attachment === attachment) {
      this.attachment = null;
      this.publish();
    }
  }
  discard() {
    if (!this.authority || this.busy) return;
    this.draft = null;
    this.attachment = null;
    this.errors = {};
    this.message = null;
    this.reviewRevision++;
    this.publish();
  }
  observe(recordId?: string, version?: number) {
    if (!this.authority) return;
    if (recordId && version !== undefined) {
      if (version <= (this.versions.get(recordId) ?? 0)) return;
      this.versions.set(recordId, version);
    }
    this.candidateRevision++;
    if (this.draft && (!recordId || recordId === this.draft.source?.recordId)) {
      this.reviewRevision++;
      if (this.draft.source)
        this.draft = freezeWorkbookValue({
          ...this.draft,
          source: { ...this.draft.source, label: "" },
        });
    }
    this.publish();
  }
  validate() {
    if (!this.draft) return false;
    this.errors = prepareNote(this.draft, "validation-only").errors;
    this.publish();
    return !Object.keys(this.errors).length;
  }
  registerSourceCoordinator(
    coordinate: (recordId: string, signal: AbortSignal) => Promise<boolean>,
  ) {
    this.sourceCoordinator = coordinate;
    return () => {
      if (this.sourceCoordinator === coordinate) this.sourceCoordinator = null;
    };
  }
  private async currentAuthority(
    actorId: string,
    editing: boolean,
    signal: AbortSignal,
  ) {
    const baseline = this.authority,
      reader = this.authorityReader,
      lifetime = this.lifetime,
      generation = this.generation;
    if (!baseline || !reader)
      throw new Error(
        "Current authority is unavailable. Retry the authorization read.",
      );
    const current = await boundedRead(
      (observed) => reader(baseline, observed),
      signal,
    );
    if (
      lifetime !== this.lifetime ||
      generation !== this.generation ||
      signal.aborted
    )
      throw new Error("Authority changed. Review again.");
    this.setAuthority(current);
    if (
      !this.authority ||
      current.actorId !== actorId ||
      current.incidentId !== this.incidentId ||
      (editing && !this.canReplay())
    )
      throw new Error("Current edit access could not be verified.");
    return current;
  }
  async recheckAuthority() {
    const lifetime = this.lifetime;
    if (!this.authority) return;
    try {
      await this.currentAuthority(
        this.authority.actorId,
        false,
        new AbortController().signal,
      );
    } catch {
      if (lifetime === this.lifetime) this.suspend();
    }
  }
  private async currentReview(
    draft: NoteDraft,
    acceptVersion: boolean,
  ): Promise<NoteReview> {
    const reader = this.reader,
      lifetime = this.lifetime;
    if (!reader) throw new Error("Note discovery is unavailable.");
    const signal = new AbortController().signal;
    const authority = await this.currentAuthority(draft.actorId, true, signal);
    if (!this.canSubmit())
      throw new Error(
        "Note creation requires an active incident and current edit access.",
      );
    const generation = this.generation,
      revision = this.reviewRevision;
    await boundedRead((observed) => reader.verifyNote(draft, observed), signal);
    let source = draft.source;
    if (source) {
      const row = await readWorkbookAuthoringRecord(
        reader,
        source.viewSchemaId,
        source.recordId,
        signal,
      );
      if (!row || row.row_version <= (this.removed.get(source.recordId) ?? 0))
        throw new Error(
          "The Note source is unavailable. Choose another source or explicitly clear it; your text is retained.",
        );
      if (row.row_version < (this.versions.get(source.recordId) ?? 0))
        throw new Error(
          "The source projection is behind a known change. Retry review.",
        );
      if (!acceptVersion && row.row_version !== source.rowVersion) {
        this.reviewRevision++;
        throw new Error(
          "The Note source changed. Review the source before creating.",
        );
      }
      source = { ...source, rowVersion: row.row_version };
    }
    if (
      lifetime !== this.lifetime ||
      generation !== this.generation ||
      revision !== this.reviewRevision ||
      this.draft?.id !== draft.id ||
      this.draft.revision !== draft.revision
    )
      throw new Error("Context changed. Review the retained Note.");
    return freezeWorkbookValue({ authority, draft: { ...draft, source } });
  }
  async review() {
    const draft = this.draft,
      lifetime = this.lifetime;
    if (!draft || this.busy || !this.canSubmit()) return;
    this.preparing = true;
    this.publish();
    try {
      const review = await this.currentReview(draft, true);
      if (lifetime !== this.lifetime) return;
      this.draft = freezeWorkbookValue({
        ...review.draft,
        revision: draft.revision + 1,
      });
      this.reviewedRevision = this.reviewRevision;
      this.message =
        "Source and creation access reviewed. Your Note is ready for a fresh submission.";
    } catch (error) {
      if (lifetime === this.lifetime) this.message = this.safeMessage(error);
    } finally {
      if (lifetime === this.lifetime) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  async submit(attachment: symbol) {
    const draft = this.draft,
      lifetime = this.lifetime,
      transport = this.transport;
    if (
      !draft ||
      this.busy ||
      !this.canSubmit() ||
      this.attachment !== attachment ||
      !transport
    )
      return;
    if (this.reviewRevision !== this.reviewedRevision) {
      this.message =
        "Review current source and access before creating this retained Note.";
      this.publish();
      return;
    }
    if (!this.validate()) return;
    // Reserve before any await: two activation handlers in one frame share one attempt.
    this.preparing = true;
    this.message = null;
    this.publish();
    try {
      if (draft.source) {
        const source = draft.source;
        const coordinated = await boundedRead(async (signal) => {
          if (
            source.viewSchemaId === "cartulary.view.timeline.v2" &&
            this.sourceCoordinator &&
            !(await this.sourceCoordinator(source.recordId, signal))
          )
            return false;
          if (!this.effects) return true;
          const settlement = await this.effects.coordinate(source, signal);
          if (settlement.kind !== "settled") return false;
          this.versions.set(
            source.recordId,
            Math.max(
              this.versions.get(source.recordId) ?? 0,
              settlement.minimumRowVersion,
            ),
          );
          return true;
        }, new AbortController().signal);
        if (!coordinated)
          throw new Error(
            "Resolve earlier source saves or conflicts before creating. Your Note is retained.",
          );
      }
      const review = await this.currentReview(draft, false);
      if (
        lifetime !== this.lifetime ||
        this.attachment !== attachment ||
        this.draft?.revision !== draft.revision ||
        this.reviewRevision !== this.reviewedRevision
      )
        return;
      const attempt = transport.capture(review, this.ids.create("note-create"));
      this.entries.set(
        attempt.clientTxnId,
        freezeWorkbookValue({
          attempt,
          phase: "submitting",
          uncertain: false,
          transportPending: true,
          receipt: null,
          failure: null,
          refresh: "none",
          observations: [],
          message: null,
        }),
      );
      this.preparing = false;
      this.publish();
      await this.dispatch(attempt.clientTxnId);
    } catch (error) {
      if (lifetime === this.lifetime) this.message = this.safeMessage(error);
    } finally {
      if (lifetime === this.lifetime) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  private safeMessage(error: unknown) {
    return error instanceof Error
      ? error.message
      : "The Note could not be prepared. Your draft is retained.";
  }
  private replace(id: string, patch: Partial<NoteEntry>) {
    const entry = this.entries.get(id);
    if (!entry) return;
    this.entries.set(id, freezeWorkbookValue({ ...entry, ...patch }));
    this.publish();
  }
  private acceptEffects(id: string) {
    const entry = this.entries.get(id);
    if (!entry?.receipt || !this.authority || this.notified.has(id)) return;
    this.effects?.accepted(entry.receipt, id);
    this.notified.add(id);
  }
  private receive(id: string, outcome: NoteOutcome, dispatch: number) {
    const entry = this.entries.get(id);
    if (!entry || entry.receipt) return;
    if (outcome.kind === "accepted") {
      const receipt = outcome.receipt;
      if (
        typeof receipt.meta?.request_id !== "string" ||
        !receipt.meta.request_id.trim() ||
        !acceptedRecordMutation(receipt.data, noteCreateView) ||
        (entry.attempt.operationID === "createRecordLinkedNote" &&
          (!("source_record_id" in receipt.data) ||
            receipt.data.source_record_id !==
              entry.attempt.review.draft.source?.recordId ||
            !("link_type" in receipt.data) ||
            receipt.data.link_type !== "references_artifact"))
      ) {
        this.receive(id, { kind: "uncertain" }, dispatch);
        return;
      }
      // The complete receipt is monotonic and retained before any effects or clearing.
      this.entries.set(
        id,
        freezeWorkbookValue({
          ...entry,
          phase: "accepted",
          receipt: structuredClone(receipt),
          transportPending: false,
          failure: null,
          refresh: "required",
          message: "Note created. Refreshing views…",
        }),
      );
      if (
        this.draft?.id === entry.attempt.review.draft.id &&
        this.draft.revision === entry.attempt.review.draft.revision
      ) {
        this.draft = null;
        this.attachment = null;
        this.errors = {};
      }
      this.message = "Note created.";
      this.publish();
      try {
        this.acceptEffects(id);
      } catch {
        /* Receipt remains accepted; reconciliation retries effects. */
      }
      void this.retryRefresh(id);
      return;
    }
    if (this.dispatches.get(id) !== dispatch) return;
    if (outcome.kind === "rejected" && !entry.uncertain) {
      this.replace(id, {
        phase: "rejected",
        transportPending: false,
        failure: outcome.failure,
        message: outcome.failure.message,
      });
      this.message = outcome.failure.message;
      if (outcome.failure.kind !== "validation") this.reviewRevision++;
    } else
      this.replace(id, {
        phase: "uncertain",
        uncertain: true,
        transportPending: false,
        message:
          "The result is unconfirmed. Recover this exact submission before editing or creating another Note.",
      });
    if (
      outcome.kind === "rejected" &&
      workbookFailureLifecycle(outcome.failure).kind === "authority_unavailable"
    )
      void this.recheckAuthority();
    this.publish();
  }
  private async dispatch(id: string) {
    const entry = this.entries.get(id),
      transport = this.transport,
      lifetime = this.lifetime;
    if (!entry || entry.receipt || !transport) return;
    const sequence = (this.dispatches.get(id) ?? 0) + 1;
    this.dispatches.set(id, sequence);
    const observed = this.observeTransport(async (signal) => {
      let outcome: NoteOutcome;
      try {
        outcome = await transport.send(entry.attempt, signal);
      } catch {
        outcome = { kind: "uncertain" };
      }
      if (lifetime === this.lifetime) this.receive(id, outcome, sequence);
      return outcome;
    });
    const result = await observed.result;
    if (
      lifetime !== this.lifetime ||
      this.entries.get(id)?.receipt ||
      this.dispatches.get(id) !== sequence
    )
      return;
    if (result.kind !== "completed")
      this.receive(id, { kind: "uncertain" }, sequence);
  }
  async replay(id: string) {
    const entry = this.entries.get(id),
      lifetime = this.lifetime;
    if (
      !entry ||
      entry.phase !== "uncertain" ||
      entry.transportPending ||
      this.preparing ||
      !this.canReplay()
    )
      return;
    this.preparing = true;
    this.publish();
    try {
      await this.currentAuthority(
        entry.attempt.review.authority.actorId,
        true,
        new AbortController().signal,
      );
      // Exact replay must not fresh-read a deleted source or reject a closed incident.
      if (lifetime !== this.lifetime || this.entries.get(id)?.receipt) return;
      this.replace(id, {
        phase: "submitting",
        transportPending: true,
        message: null,
      });
      this.preparing = false;
      await this.dispatch(id);
    } catch {
      if (lifetime === this.lifetime)
        this.replace(id, {
          message: "Current authority could not be verified. Retry recovery.",
        });
    } finally {
      if (lifetime === this.lifetime) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  async retryRefresh(id: string) {
    const entry = this.entries.get(id),
      reader = this.reader,
      lifetime = this.lifetime;
    if (
      !entry?.receipt ||
      entry.refresh === "complete" ||
      this.refreshes.has(id) ||
      !reader ||
      !this.authority
    )
      return;
    this.refreshes.add(id);
    this.replace(id, { refresh: "refreshing" });
    try {
      await this.currentAuthority(
        entry.attempt.review.authority.actorId,
        false,
        new AbortController().signal,
      );
      const generation = this.generation;
      const views = [
        ...new Set([
          noteCreateView,
          ...(entry.attempt.review.draft.source
            ? [entry.attempt.review.draft.source.viewSchemaId]
            : []),
          ...entry.observations.flatMap((event) =>
            event.payload.affected_views
              .map((view) => view.view_schema_id)
              .filter((view) => getViewContract(view)),
          ),
        ]),
      ];
      const records = [
        entry.receipt.data.row.record_id,
        ...(entry.attempt.review.draft.source
          ? [entry.attempt.review.draft.source.recordId]
          : []),
      ];
      await boundedRead(async (signal) => {
        for (const viewSchemaId of views) {
          const result = await reader.page({
            viewSchemaId,
            cursor: null,
            queryState: emptyWorkbookQueryState(),
            signal,
          });
          if (result.kind !== "accepted")
            throw new Error("Projection refresh failed.");
          // Reads provide evidence only. Never insert a receipt or older query row into a live view.
        }
        if (lifetime !== this.lifetime || generation !== this.generation)
          throw new Error("Authority changed.");
        this.acceptEffects(id);
        await this.effects?.refresh(views, records);
      }, new AbortController().signal);
      if (lifetime !== this.lifetime || generation !== this.generation)
        throw new Error("Authority changed.");
      this.candidateRevision++;
      this.replace(id, {
        refresh: "complete",
        message: "Note created. Views refreshed.",
      });
    } catch {
      if (lifetime === this.lifetime)
        this.replace(id, {
          refresh: "required",
          message:
            "Note created, but views need refresh. Retry refresh sends reads only.",
        });
    } finally {
      if (lifetime === this.lifetime) this.refreshes.delete(id);
    }
  }
  observeSocket(message: RecordChangedMessage) {
    if (!this.authority || message.incident_id !== this.incidentId) return;
    const payload = message.payload;
    if (payload.row_version >= (this.versions.get(payload.record_id) ?? 0)) {
      if (
        payload.affected_views.some(
          (view) =>
            view.change_kind === "remove" &&
            [noteCreateView, ...noteSourceViews].includes(view.view_schema_id),
        )
      )
        this.removed.set(payload.record_id, payload.row_version);
      else if (payload.row_version > (this.removed.get(payload.record_id) ?? 0))
        this.removed.delete(payload.record_id);
    }
    this.observe(payload.record_id, payload.row_version);
    const entry = this.entries.get(payload.client_txn_id);
    if (
      !entry ||
      entry.attempt.review.authority.actorId !== payload.actor_user_id ||
      entry.observations.some((event) => event.event_id === message.event_id) ||
      (entry.receipt &&
        entry.receipt.data.change_set_id !== payload.change_set_id)
    )
      return;
    this.replace(payload.client_txn_id, {
      observations: [...entry.observations, structuredClone(message)],
    });
  }
}
