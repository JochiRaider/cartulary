import {
  evidenceViewSchemaId,
  type InspectorFeatureGroup,
  partiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import type { SheetRef } from "../../../shared/sheetRef";
import { readWorkbookAuthoringRecord } from "../../adapters/readWorkbookAuthoringRecord";
import type { WorkbookProtocolConflictResolutionReceipt } from "../../adapters/workbookProtocolTypes";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type {
  WorkbookAuthoringAuthorityReader,
  WorkbookAuthoringReadPort,
} from "../../ports/WorkbookAuthoringReadPort";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookSameFieldConflictPayload } from "../../runtime/workbookConflictModel";
import { freezeWorkbookValue } from "../../utils/freezeWorkbookValue";
import {
  buildTimelineRelatedEvidenceDraft,
  prepareTimelineRelatedEvidence,
  type TimelineRelatedEvidenceDraft,
} from "./timelineRelatedEvidenceModel";
import type {
  RelatedEvidenceAttempt,
  RelatedEvidenceCheckpoint,
  RelatedEvidenceOutcome,
  RelatedEvidenceReceipt,
  RelatedEvidenceReview,
  RelatedEvidenceStageEntry,
  RelatedEvidenceTransport,
} from "./timelineRelatedEvidenceOperation";

type Snapshot = Readonly<{
  authority: WorkbookMutationAuthority | null;
  draft: TimelineRelatedEvidenceDraft | null;
  attachment: symbol | null;
  errors: Readonly<Record<string, string>>;
  message: string | null;
  candidateRevision: number;
  needsReview: boolean;
  preparing: boolean;
  checkpoints: readonly RelatedEvidenceCheckpoint[];
}>;

/** Incident-lifetime intent. Presentation can detach without changing original identity. */
export class WorkbookTimelineRelatedEvidenceOwner {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private lifetime = 0;
  private generation = 0;
  private draft: TimelineRelatedEvidenceDraft | null = null;
  private attachment: symbol | null = null;
  private nextDraftId = 0;
  private reviewRevision = 0;
  private reviewedRevision = 0;
  private candidateRevision = 0;
  private preparing = false;
  private preparation = 0;
  private presentationRevision = 0;
  private reviewedSource: WorkbookQueryRow | null = null;
  private transport: RelatedEvidenceTransport | null = null;
  private readonly checkpoints = new Map<string, RelatedEvidenceCheckpoint>();
  private readonly dispatches = new Map<string, number>();
  private readonly refreshes = new Set<string>();
  private readonly newIdReviewed = new Set<string>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  private readonly removed = new Map<string, number>();
  private sourceCoordinator:
    | ((recordId: string, signal: AbortSignal) => Promise<boolean>)
    | null = null;
  private errors: Readonly<Record<string, string>> = {};
  private message: string | null = null;
  private readonly listeners = new Set<() => void>();
  private readonly versions = new Map<string, number>();
  private reader: WorkbookAuthoringReadPort | null = null;
  private authorityReader: WorkbookAuthoringAuthorityReader | null = null;
  private snapshot: Snapshot = {
    authority: null,
    draft: null,
    attachment: null,
    errors: {},
    message: null,
    candidateRevision: 0,
    needsReview: false,
    preparing: false,
    checkpoints: [],
  };
  constructor(
    readonly incidentId: string,
    private readonly ids?: SecureTransactionIdPort,
    private readonly effects?: {
      readonly coordinate: (
        recordId: string,
        signal: AbortSignal,
      ) => Promise<boolean>;
      readonly accepted: (
        receipt: RelatedEvidenceReceipt,
        clientTxnId: string,
      ) => void;
      readonly refresh: (
        views: readonly string[],
        recordIds: readonly string[],
      ) => Promise<void>;
      readonly conflict: (
        checkpoint: RelatedEvidenceCheckpoint,
        conflict: WorkbookSameFieldConflictPayload,
      ) => void;
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
    reader: WorkbookAuthoringReadPort,
    authorityReader: WorkbookAuthoringAuthorityReader,
    transport?: RelatedEvidenceTransport,
  ) {
    this.reader = reader;
    this.authorityReader = authorityReader;
    if (transport) this.transport = transport;
  }
  registerSourceCoordinator(
    coordinate: (recordId: string, signal: AbortSignal) => Promise<boolean>,
  ) {
    this.sourceCoordinator = coordinate;
    return () => {
      if (this.sourceCoordinator === coordinate) this.sourceCoordinator = null;
    };
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
      checkpoints: this.authority ? [...this.checkpoints.values()] : [],
    };
    for (const listener of this.listeners) listener();
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId && this.actorId !== authority.actorId))
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
    this.presentationRevision++;
    if (this.draft)
      this.draft = freezeWorkbookValue({ ...this.draft, labels: {} });
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
    this.authority = null;
    this.actorId = null;
    this.draft = null;
    this.attachment = null;
    this.errors = {};
    this.message = null;
    this.preparing = false;
    this.versions.clear();
    this.reviewRevision++;
    this.candidateRevision++;
    this.publish();
    this.checkpoints.clear();
    this.dispatches.clear();
    this.refreshes.clear();
    this.newIdReviewed.clear();
    this.rows.clear();
    this.removed.clear();
    this.reviewedSource = null;
    this.presentationRevision++;
    this.publish();
  }
  canSubmit() {
    return (
      !!this.authority?.sessionIdentity &&
      !this.authority.closed &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  canReplay() {
    return (
      !!this.authority?.sessionIdentity &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  private stages() {
    return [...this.checkpoints.values()].flatMap((checkpoint) => [
      checkpoint.create,
      ...checkpoint.links,
    ]);
  }
  get busy() {
    return (
      this.preparing ||
      this.stages().some(
        (entry) => entry.phase === "submitting" || entry.phase === "uncertain",
      )
    );
  }
  get pendingCount() {
    return (
      Number(this.preparing) +
      this.stages().filter(
        (entry) => entry.transportPending || entry.refresh === "refreshing",
      ).length
    );
  }
  get uncertainCount() {
    return this.stages().filter((entry) => entry.phase === "uncertain").length;
  }
  blocksRecord(recordId: string) {
    return this.stages().some(
      (entry) =>
        entry.attempt.stage === "link" &&
        entry.attempt.review.draft.source.recordId === recordId &&
        (entry.phase === "submitting" || entry.phase === "uncertain"),
    );
  }
  get blockedCount() {
    return [...this.checkpoints.values()].filter(
      (checkpoint) =>
        checkpoint.create.phase === "uncertain" ||
        (checkpoint.create.receipt && !this.linkComplete(checkpoint)) ||
        [checkpoint.create, ...checkpoint.links].some(
          (entry) => entry.refresh === "required",
        ),
    ).length;
  }
  linkComplete(checkpoint: RelatedEvidenceCheckpoint) {
    return (
      checkpoint.links.some((entry) => !!entry.receipt) ||
      (!!checkpoint.resolution &&
        !!checkpoint.create.receipt &&
        this.hasAssociation(
          checkpoint.resolution.receipt.data.row,
          checkpoint.create.receipt.data.row.record_id,
        )) ||
      checkpoint.associationPresent
    );
  }
  begin(
    subject: WorkbookInspectorLiveRowBinding,
    feature: InspectorFeatureGroup,
    sheetRef: SheetRef,
    attachment: symbol,
  ) {
    if (!this.canSubmit() || !this.authority) return false;
    if (
      [...this.checkpoints.values()].some(
        (checkpoint) =>
          checkpoint.create.receipt &&
          !this.linkComplete(checkpoint) &&
          checkpoint.create.attempt.review.draft.source.recordId ===
            subject.subject.recordId,
      )
    ) {
      this.message =
        "Evidence already created for this Timeline record needs link recovery. Review its retained result before creating more Evidence.";
      this.publish();
      return false;
    }
    if (this.draft || this.busy) {
      this.message =
        "Your Evidence draft is retained. Resume it or explicitly discard it before starting another.";
      this.publish();
      return false;
    }
    const draft = buildTimelineRelatedEvidenceDraft(
      ++this.nextDraftId,
      this.authority,
      subject,
      feature,
      sheetRef,
    );
    if (!draft) return false;
    this.draft = draft;
    this.attachment = attachment;
    this.errors = {};
    this.message = null;
    this.reviewedSource = freezeWorkbookValue({
      record_id: draft.source.recordId,
      row_version: draft.source.rowVersion,
      cells: structuredClone(subject.cells),
    });
    this.presentationRevision++;
    this.versions.set(draft.source.recordId, draft.source.rowVersion);
    this.reviewedRevision = this.reviewRevision;
    this.publish();
    return true;
  }
  update(
    field: string,
    value: string,
    labels: Readonly<Record<string, string>> = {},
  ) {
    if (
      !this.draft ||
      this.busy ||
      !this.canSubmit() ||
      !this.draft.target.fieldMap[field]?.createWritable
    )
      return;
    this.draft = freezeWorkbookValue({
      ...this.draft,
      revision: this.draft.revision + 1,
      values: { ...this.draft.values, [field]: value },
      labels: { ...this.draft.labels, ...labels },
    });
    this.errors = {};
    this.message = null;
    this.publish();
  }
  resume(attachment: symbol) {
    if (this.draft && this.authority) {
      this.attachment = attachment;
      this.reviewRevision++;
      this.presentationRevision++;
      this.publish();
    }
  }
  detach(attachment: symbol) {
    if (this.attachment === attachment) {
      this.attachment = null;
      this.reviewRevision++;
      this.presentationRevision++;
      this.publish();
    }
  }
  discard() {
    if (this.busy) return;
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
    if (
      this.draft &&
      (!recordId ||
        recordId === this.draft.source.recordId ||
        Object.values(this.draft.values).includes(recordId))
    ) {
      this.reviewRevision++;
      this.draft = freezeWorkbookValue({ ...this.draft, labels: {} });
    }
    for (const [id, checkpoint] of this.checkpoints) {
      const draft = checkpoint.create.attempt.review.draft;
      if (
        !recordId ||
        recordId === draft.source.recordId ||
        recordId === checkpoint.create.receipt?.data.row.record_id
      ) {
        this.checkpoints.set(
          id,
          freezeWorkbookValue({
            ...checkpoint,
            review: null,
            observationRevision: checkpoint.observationRevision + 1,
          }),
        );
        if (checkpoint.create.receipt) void this.retryRefresh(id);
      }
    }
    this.publish();
  }
  async review() {
    const draft = this.draft,
      lifetime = this.lifetime;
    if (!draft || !this.canSubmit() || this.busy) return false;
    const preparation = ++this.preparation;
    this.preparing = true;
    this.publish();
    try {
      if (this.effects && !(await this.coordinate(draft.source.recordId)))
        throw new Error("Resolve earlier Timeline saves before review.");
      const review = await this.currentReview(draft, true);
      if (
        lifetime !== this.lifetime ||
        this.draft?.revision !== draft.revision ||
        this.draft?.id !== draft.id
      )
        return false;
      this.versions.set(review.source.record_id, review.source.row_version);
      this.reviewedSource = review.source;
      this.reviewedRevision = this.reviewRevision;
      this.message = "Original Timeline record and metadata reviewed.";
      return true;
    } catch (error) {
      if (lifetime === this.lifetime) {
        this.message = this.safeMessage(error);
        this.reviewRevision++;
      }
      return false;
    } finally {
      if (lifetime === this.lifetime && preparation === this.preparation) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  async recheckAuthority() {
    const baseline = this.authority,
      reader = this.authorityReader,
      generation = this.generation;
    if (!baseline || !reader) return;
    try {
      const current = await boundedRead(
        (signal) => reader(baseline, signal),
        new AbortController().signal,
      );
      if (generation === this.generation) this.setAuthority(current);
    } catch {
      /* Only the authority reader establishes incident/session access loss. */
    }
  }
  private async coordinate(recordId: string) {
    return boundedRead(async (signal) => {
      if (!this.effects || !(await this.effects.coordinate(recordId, signal)))
        return false;
      return this.sourceCoordinator
        ? this.sourceCoordinator(recordId, signal)
        : true;
    }, new AbortController().signal);
  }
  private async currentReview(
    draft: TimelineRelatedEvidenceDraft,
    checkReferences: boolean,
  ): Promise<RelatedEvidenceReview> {
    const reader = this.reader,
      authorityReader = this.authorityReader,
      baseline = this.authority;
    const lifetime = this.lifetime,
      generation = this.generation,
      presentation = this.presentationRevision;
    if (!reader || !authorityReader || !baseline)
      throw new Error("Current authority is unavailable.");
    const current = await boundedRead(
      (signal) => authorityReader(baseline, signal),
      new AbortController().signal,
    );
    if (lifetime !== this.lifetime || generation !== this.generation)
      throw new Error("Authority changed. Review again.");
    this.setAuthority(current);
    if (
      !this.canSubmit() ||
      current.actorId !== draft.actorId ||
      current.incidentId !== draft.incidentId
    )
      throw new Error(
        "Creation and linking require current edit access to an active incident.",
      );
    const authorizedGeneration = this.generation,
      observedRevision = this.reviewRevision;
    await boundedRead(
      (signal) => reader.verify(draft, signal),
      new AbortController().signal,
    );
    const source = await readWorkbookAuthoringRecord(
      reader,
      draft.source.viewSchemaId,
      draft.source.recordId,
      new AbortController().signal,
    );
    if (
      !source ||
      source.cells["timeline.capture_state"]?.value === "superseded" ||
      source.row_version <= (this.removed.get(source.record_id) ?? 0)
    )
      throw new Error(
        "The original Timeline record is unavailable or superseded. Retained work is preserved.",
      );
    if (source.row_version < (this.versions.get(source.record_id) ?? 0))
      throw new Error(
        "The original Timeline projection is behind a known change. Retry review.",
      );
    if (checkReferences) {
      const errors = prepareTimelineRelatedEvidence(draft).errors;
      for (const field of draft.target.fields.filter(
        (field) =>
          field.directReferenceContractId === "same_incident_party_ref_v1",
      )) {
        const id = draft.values[field.fieldKey];
        if (
          id &&
          !errors[field.fieldKey] &&
          !(await readWorkbookAuthoringRecord(
            reader,
            partiesViewSchemaId,
            id,
            new AbortController().signal,
          ))
        )
          errors[field.fieldKey] =
            "This Party is unavailable. Choose another reference or remove it.";
      }
      this.errors = errors;
      if (Object.keys(errors).length)
        throw new Error(
          "Correct the highlighted fields. Your input is retained.",
        );
    }
    if (
      lifetime !== this.lifetime ||
      authorizedGeneration !== this.generation ||
      presentation !== this.presentationRevision ||
      observedRevision !== this.reviewRevision ||
      source.row_version < (this.versions.get(source.record_id) ?? 0)
    )
      throw new Error("Context changed. Review the retained operation.");
    return freezeWorkbookValue({
      authority: current,
      draft,
      source,
      presentationRevision: this.presentationRevision,
    });
  }
  async submit(attachment: symbol) {
    const draft = this.draft,
      expected = this.reviewedSource,
      lifetime = this.lifetime,
      revision = this.reviewRevision;
    if (
      this.busy ||
      !draft ||
      !expected ||
      !this.canSubmit() ||
      this.attachment !== attachment ||
      !this.transport ||
      !this.ids ||
      revision !== this.reviewedRevision ||
      [...this.checkpoints.values()].some(
        (checkpoint) =>
          checkpoint.create.attempt.review.draft.id === draft.id &&
          checkpoint.create.failure?.kind === "client_txn_conflict" &&
          !this.newIdReviewed.has(checkpoint.id),
      )
    )
      return;
    const preparation = ++this.preparation;
    this.preparing = true;
    this.message = null;
    this.publish();
    try {
      if (!(await this.coordinate(draft.source.recordId)))
        throw new Error(
          "Resolve earlier Timeline saves before creating. Your draft is retained.",
        );
      const review = await this.currentReview(draft, true);
      if (
        lifetime !== this.lifetime ||
        this.draft?.id !== draft.id ||
        this.draft.revision !== draft.revision ||
        this.attachment !== attachment ||
        this.reviewRevision !== revision
      )
        return;
      if (review.source.row_version !== expected.row_version)
        throw new Error(
          "The original Timeline record changed. Review it before creating Evidence.",
        );
      const attempt = this.transport.capture(
        "create",
        review,
        this.ids.create("timeline-evidence-create"),
        null,
      );
      const entry = this.initialStage(attempt);
      this.checkpoints.set(
        attempt.clientTxnId,
        freezeWorkbookValue({
          id: attempt.clientTxnId,
          observationRevision: 0,
          create: entry,
          links: [],
          review: null,
          message: null,
          sourceUnavailable: false,
          evidenceUnavailable: false,
          associationPresent: false,
        }),
      );
      this.preparing = false;
      this.publish();
      await this.dispatch(attempt.clientTxnId, attempt.clientTxnId);
    } catch (error) {
      if (lifetime === this.lifetime) {
        this.message = this.safeMessage(error);
        if (!Object.keys(this.errors).length) this.reviewRevision++;
      }
    } finally {
      if (lifetime === this.lifetime && preparation === this.preparation) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  private initialStage(
    attempt: RelatedEvidenceAttempt,
  ): RelatedEvidenceStageEntry {
    return freezeWorkbookValue({
      attempt,
      phase: "submitting",
      transportPending: true,
      receipt: null,
      failure: null,
      uncertain: false,
      refresh: "none",
      observations: [],
    });
  }
  private safeMessage(error: unknown) {
    return error instanceof Error
      ? error.message
      : "The operation could not be prepared. Retained work is preserved.";
  }
  private replace(id: string, patch: Partial<RelatedEvidenceCheckpoint>) {
    const checkpoint = this.checkpoints.get(id);
    if (checkpoint) {
      this.checkpoints.set(
        id,
        freezeWorkbookValue({ ...checkpoint, ...patch }),
      );
      this.publish();
    }
  }
  private stage(id: string, txn: string) {
    const checkpoint = this.checkpoints.get(id);
    return checkpoint?.create.attempt.clientTxnId === txn
      ? checkpoint.create
      : checkpoint?.links.find((entry) => entry.attempt.clientTxnId === txn);
  }
  private replaceStage(
    id: string,
    txn: string,
    patch: Partial<RelatedEvidenceStageEntry>,
  ) {
    const checkpoint = this.checkpoints.get(id),
      entry = this.stage(id, txn);
    if (!checkpoint || !entry) return;
    const next = freezeWorkbookValue({ ...entry, ...patch });
    this.replace(id, {
      observationRevision:
        checkpoint.observationRevision +
        Number(!!patch.receipt || !!patch.observations),
      ...(txn === checkpoint.create.attempt.clientTxnId
        ? { create: next }
        : {
            links: checkpoint.links.map((item) =>
              item === entry ? next : item,
            ),
          }),
    });
  }
  private receive(id: string, txn: string, outcome: RelatedEvidenceOutcome) {
    const entry = this.stage(id, txn);
    if (!entry || entry.receipt) return;
    if (outcome.kind === "accepted") {
      const { data, meta } = outcome.receipt,
        expectedView =
          entry.attempt.stage === "create"
            ? evidenceViewSchemaId
            : timelineViewSchemaId;
      if (
        typeof meta?.request_id !== "string" ||
        !meta.request_id.trim() ||
        !data?.change_set_id ||
        data.view_schema_id !== expectedView ||
        !data.row?.record_id ||
        !Number.isSafeInteger(data.row.row_version) ||
        data.row.row_version < 1 ||
        (entry.attempt.stage === "link" &&
          (data.row.record_id !== entry.attempt.review.draft.source.recordId ||
            data.row.row_version <= entry.attempt.review.source.row_version)) ||
        entry.observations.some(
          (event) =>
            event.payload.change_set_id !== data.change_set_id ||
            event.payload.record_id !== data.row.record_id,
        )
      )
        outcome = { kind: "uncertain" };
    }
    if (outcome.kind === "accepted") {
      const receipt = freezeWorkbookValue(structuredClone(outcome.receipt));
      this.versions.set(
        receipt.data.row.record_id,
        Math.max(
          receipt.data.row.row_version,
          this.versions.get(receipt.data.row.record_id) ?? 0,
        ),
      );
      this.replaceStage(id, txn, {
        phase: "accepted",
        receipt,
        failure: null,
        refresh: "required",
      });
      if (
        entry.attempt.stage === "create" &&
        this.draft?.id === entry.attempt.review.draft.id
      ) {
        this.draft = null;
        this.errors = {};
      }
      this.replace(id, {
        message:
          entry.attempt.stage === "create"
            ? "Evidence created; Timeline link incomplete."
            : "Evidence created and linked to the original Timeline record.",
        review: null,
      });
      // Receipt retention precedes effects, including effects that can synchronously detach presentation.
      try {
        this.effects?.accepted(receipt, txn);
      } catch {
        /* Reconciliation remains recoverable. */
      }
      if (
        entry.attempt.stage === "create" &&
        !entry.uncertain &&
        entry.attempt.review.presentationRevision === this.presentationRevision
      )
        void this.link(id, true);
      else if (this.authority) void this.retryRefresh(id);
    } else if (
      outcome.kind === "rejected" &&
      !entry.uncertain &&
      entry.phase !== "uncertain"
    ) {
      this.replaceStage(id, txn, {
        phase: "rejected",
        failure: outcome.failure,
      });
      this.replace(id, {
        message:
          entry.attempt.stage === "create"
            ? outcome.failure.message
            : `Evidence created; Timeline link incomplete. ${outcome.failure.message}`,
        review: null,
      });
      if (entry.attempt.stage === "create") {
        this.errors =
          outcome.failure.kind === "validation"
            ? Object.fromEntries(
                (outcome.failure.fields ?? []).map((field) => [
                  field.field,
                  field.message,
                ]),
              )
            : {};
        this.message = outcome.failure.message;
        this.reviewRevision++;
      } else if (outcome.failure.kind === "same_field_conflict") {
        const checkpoint = this.checkpoints.get(id);
        if (checkpoint)
          this.effects?.conflict(checkpoint, outcome.failure.conflict);
      }
      this.publish();
      if (
        ["authorization_lost", "authentication_required"].includes(
          outcome.failure.kind,
        )
      )
        void this.recheckAuthority();
    } else {
      this.replaceStage(id, txn, {
        phase: "uncertain",
        uncertain: true,
        failure: outcome.kind === "rejected" ? outcome.failure : null,
      });
      this.replace(id, {
        message: `${entry.attempt.stage === "create" ? "Creation" : "Timeline linking"} outcome is unconfirmed. Replay the retained operation to recover it.`,
      });
    }
  }
  private async dispatch(id: string, txn: string) {
    const entry = this.stage(id, txn),
      transport = this.transport,
      lifetime = this.lifetime;
    if (!entry || !transport) return;
    const dispatch = (this.dispatches.get(txn) ?? 0) + 1;
    this.dispatches.set(txn, dispatch);
    const observation = this.observeTransport(async (signal) => {
      const result = await transport.send(entry.attempt, signal);
      if (
        lifetime === this.lifetime &&
        (result.kind === "accepted" || dispatch === this.dispatches.get(txn))
      )
        this.receive(id, txn, result);
      return result;
    });
    void observation.settled.then(() => {
      if (lifetime === this.lifetime && dispatch === this.dispatches.get(txn))
        this.replaceStage(id, txn, { transportPending: false });
    });
    const result = await observation.result;
    if (
      lifetime === this.lifetime &&
      dispatch === this.dispatches.get(txn) &&
      result.kind !== "completed" &&
      !this.stage(id, txn)?.receipt
    ) {
      this.replaceStage(id, txn, {
        phase: "uncertain",
        uncertain: true,
        transportPending: false,
      });
      this.replace(id, {
        message:
          "Outcome unconfirmed. Recover the original operation; closing the panel does not cancel the server write.",
      });
    }
  }
  async replay(id: string, txn: string) {
    const entry = this.stage(id, txn),
      baseline = this.authority,
      authorityReader = this.authorityReader,
      reader = this.reader;
    const generation = this.generation,
      lifetime = this.lifetime;
    if (
      !entry ||
      entry.phase !== "uncertain" ||
      entry.transportPending ||
      this.preparing ||
      !this.canReplay() ||
      !baseline ||
      !authorityReader ||
      !reader
    )
      return;
    const preparation = ++this.preparation;
    this.preparing = true;
    this.publish();
    try {
      const authority = await boundedRead(
        (signal) => authorityReader(baseline, signal),
        new AbortController().signal,
      );
      if (lifetime !== this.lifetime || generation !== this.generation) return;
      this.setAuthority(authority);
      if (
        !this.canReplay() ||
        authority.actorId !== entry.attempt.review.authority.actorId ||
        authority.incidentId !== entry.attempt.review.authority.incidentId
      )
        throw new Error(
          "Current edit access is required to recover this operation.",
        );
      const currentGeneration = this.generation;
      await boundedRead(
        (signal) => reader.verify(entry.attempt.review.draft, signal),
        new AbortController().signal,
      );
      if (
        lifetime !== this.lifetime ||
        currentGeneration !== this.generation ||
        this.stage(id, txn)?.receipt
      )
        return;
      this.replaceStage(id, txn, {
        phase: "submitting",
        transportPending: true,
      });
      this.preparing = false;
      await this.dispatch(id, txn);
    } catch (error) {
      if (lifetime === this.lifetime)
        this.replace(id, { message: this.safeMessage(error) });
    } finally {
      if (lifetime === this.lifetime && preparation === this.preparation) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  reviewNewRequestId(id: string, txn: string) {
    const entry = this.stage(id, txn);
    if (
      !entry ||
      entry.phase !== "rejected" ||
      entry.failure?.kind !== "client_txn_conflict" ||
      this.busy
    )
      return;
    this.newIdReviewed.add(txn);
    this.reviewRevision++;
    this.replace(id, {
      review: null,
      message:
        "Review the retained values and original source before a new request ID is created.",
    });
  }
  async reviewLink(id: string) {
    const checkpoint = this.checkpoints.get(id),
      lifetime = this.lifetime;
    if (
      !checkpoint?.create.receipt ||
      this.busy ||
      !this.canSubmit() ||
      this.linkComplete(checkpoint)
    )
      return false;
    const preparation = ++this.preparation;
    this.preparing = true;
    this.publish();
    try {
      if (
        !(await this.coordinate(
          checkpoint.create.attempt.review.draft.source.recordId,
        ))
      )
        throw new Error(
          "Resolve earlier Timeline saves before reviewing the link.",
        );
      const review = await this.currentReview(
        checkpoint.create.attempt.review.draft,
        false,
      );
      const target = await this.readTarget(checkpoint);
      if (!target) {
        this.replace(id, { evidenceUnavailable: true });
        throw new Error(
          "Evidence was created but is currently unavailable. Linking remains incomplete.",
        );
      }
      if (
        lifetime !== this.lifetime ||
        review.presentationRevision !== this.presentationRevision ||
        !this.canSubmit() ||
        review.source.row_version <
          (this.versions.get(review.source.record_id) ?? 0)
      )
        return false;
      this.replace(id, {
        review,
        sourceUnavailable: false,
        evidenceUnavailable: false,
        associationPresent: this.hasAssociation(
          review.source,
          target.record_id,
        ),
        message:
          "Original Timeline record reviewed. Link the created Evidence when ready.",
      });
      return true;
    } catch (error) {
      if (lifetime === this.lifetime)
        this.replace(id, {
          review: null,
          message: `Evidence created; Timeline link incomplete. ${this.safeMessage(error)}`,
        });
      return false;
    } finally {
      if (lifetime === this.lifetime && preparation === this.preparation) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  async link(id: string, automatic = false) {
    const checkpoint = this.checkpoints.get(id),
      lifetime = this.lifetime;
    const expected = automatic
      ? checkpoint?.create.attempt.review
      : checkpoint?.review;
    if (
      !checkpoint?.create.receipt ||
      !expected ||
      this.preparing ||
      !this.canSubmit() ||
      !this.transport ||
      !this.ids ||
      this.linkComplete(checkpoint) ||
      checkpoint.links.some(
        (entry) =>
          entry.phase === "uncertain" ||
          entry.phase === "submitting" ||
          (entry.failure?.kind === "client_txn_conflict" &&
            !this.newIdReviewed.has(entry.attempt.clientTxnId)),
      )
    )
      return;
    const preparation = ++this.preparation;
    this.preparing = true;
    this.publish();
    try {
      if (!(await this.coordinate(expected.draft.source.recordId)))
        throw new Error(
          "Resolve earlier Timeline saves, then review the link.",
        );
      const review = await this.currentReview(expected.draft, false);
      if (
        review.presentationRevision !== expected.presentationRevision ||
        review.source.row_version !== expected.source.row_version
      )
        throw new Error(
          "The source or presentation changed. Review the original Timeline record before linking.",
        );
      const target = await this.readTarget(checkpoint);
      if (!target) {
        this.replace(id, { evidenceUnavailable: true });
        throw new Error(
          "Created Evidence is unavailable. Retry read-only refresh to check visibility.",
        );
      }
      if (
        lifetime !== this.lifetime ||
        review.presentationRevision !== this.presentationRevision ||
        review.source.row_version <
          (this.versions.get(review.source.record_id) ?? 0) ||
        !this.canSubmit()
      )
        return;
      if (this.hasAssociation(review.source, target.record_id)) {
        this.replace(id, { associationPresent: true, review: null });
        return;
      }
      const attempt = this.transport.capture(
        "link",
        review,
        this.ids.create("timeline-evidence-link"),
        checkpoint.create.receipt.data.row.record_id,
      );
      this.replace(id, {
        links: [...checkpoint.links, this.initialStage(attempt)],
        review: null,
        sourceUnavailable: false,
        evidenceUnavailable: false,
      });
      this.preparing = false;
      await this.dispatch(id, attempt.clientTxnId);
    } catch (error) {
      if (lifetime === this.lifetime)
        this.replace(id, {
          review: null,
          message: `Evidence created; Timeline link incomplete. ${this.safeMessage(error)}`,
        });
    } finally {
      if (lifetime === this.lifetime) {
        if (preparation === this.preparation) this.preparing = false;
        this.publish();
        void this.retryRefresh(id);
      }
    }
  }
  private hasAssociation(row: WorkbookQueryRow, recordId: string) {
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
          (item.linked_record_id === recordId || item.record_id === recordId),
      )
    );
  }
  conflictResolved(
    conflictToken: string,
    kind: string,
    receipt: WorkbookProtocolConflictResolutionReceipt,
  ) {
    const checkpoint = [...this.checkpoints.values()].find((checkpoint) =>
      checkpoint.links.some(
        (entry) =>
          entry.failure?.kind === "same_field_conflict" &&
          entry.failure.conflict.conflict_token === conflictToken,
      ),
    );
    if (
      !checkpoint?.create.receipt ||
      receipt.data.view_schema_id !== timelineViewSchemaId ||
      receipt.data.row.record_id !==
        checkpoint.create.attempt.review.draft.source.recordId
    )
      return;
    const id = checkpoint.id;
    this.versions.set(
      receipt.data.row.record_id,
      Math.max(
        receipt.data.row.row_version,
        this.versions.get(receipt.data.row.record_id) ?? 0,
      ),
    );
    const present = this.hasAssociation(
      receipt.data.row,
      checkpoint.create.receipt.data.row.record_id,
    );
    this.replace(id, {
      resolution: freezeWorkbookValue({
        kind,
        receipt: structuredClone(receipt),
      }),
      associationPresent: present,
      review: null,
      message: present
        ? "Evidence linked by the reviewed collection resolution."
        : "The saved collection was kept. Evidence creation succeeded; linking remains incomplete.",
    });
    void this.retryRefresh(id);
  }
  conflictChanged(
    conflictToken: string,
    conflict: WorkbookSameFieldConflictPayload,
  ) {
    for (const [id, checkpoint] of this.checkpoints) {
      for (const entry of checkpoint.links) {
        if (
          entry.failure?.kind !== "same_field_conflict" ||
          entry.failure.conflict.conflict_token !== conflictToken
        )
          continue;
        this.replaceStage(id, entry.attempt.clientTxnId, {
          failure: freezeWorkbookValue({
            ...entry.failure,
            conflict: structuredClone(conflict),
          }),
        });
        this.replace(id, { review: null });
      }
    }
  }
  private async readTarget(checkpoint: RelatedEvidenceCheckpoint) {
    const target = checkpoint.create.receipt?.data.row.record_id;
    if (!this.reader || !target) return null;
    const row = await readWorkbookAuthoringRecord(
      this.reader,
      evidenceViewSchemaId,
      target,
      new AbortController().signal,
    );
    if (row && row.row_version <= (this.removed.get(target) ?? 0)) return null;
    if (row && row.row_version < (this.versions.get(target) ?? 0))
      throw new Error(
        "The Evidence projection is behind a known change. Retry review.",
      );
    return row;
  }
  latestRow(recordId: string) {
    const row = this.rows.get(recordId);
    return this.authority &&
      row &&
      !this.removed.has(recordId) &&
      row.row_version >= (this.versions.get(recordId) ?? 0)
      ? row
      : null;
  }
  observeSocket(message: RecordChangedMessage) {
    if (!this.authority || message.incident_id !== this.incidentId) return;
    const payload = message.payload;
    if (payload.row_version >= (this.versions.get(payload.record_id) ?? 0)) {
      if (
        payload.affected_views.some(
          (view) =>
            view.change_kind === "remove" &&
            (view.view_schema_id === timelineViewSchemaId ||
              view.view_schema_id === evidenceViewSchemaId),
        )
      )
        this.removed.set(payload.record_id, payload.row_version);
      else if (payload.row_version > (this.removed.get(payload.record_id) ?? 0))
        this.removed.delete(payload.record_id);
    }
    this.observe(payload.record_id, payload.row_version);
    for (const [id, checkpoint] of this.checkpoints) {
      const entry = [checkpoint.create, ...checkpoint.links].find(
        (stage) =>
          stage.attempt.clientTxnId === payload.client_txn_id &&
          (stage.attempt.stage === "create" ||
            payload.record_id === stage.attempt.review.draft.source.recordId),
      );
      if (
        !entry &&
        checkpoint.create.receipt &&
        (payload.record_id ===
          checkpoint.create.attempt.review.draft.source.recordId ||
          payload.record_id === checkpoint.create.receipt.data.row.record_id)
      ) {
        this.replace(id, {
          observationRevision: checkpoint.observationRevision + 1,
        });
        for (const stage of [checkpoint.create, ...checkpoint.links])
          if (stage.receipt)
            this.replaceStage(id, stage.attempt.clientTxnId, {
              refresh: "required",
            });
        void this.retryRefresh(id);
      }
      if (
        !entry ||
        entry.attempt.review.authority.actorId !== payload.actor_user_id ||
        entry.observations.some(
          (event) => event.event_id === message.event_id,
        ) ||
        (entry.receipt &&
          entry.receipt.data.change_set_id !== payload.change_set_id)
      )
        continue;
      this.replaceStage(id, entry.attempt.clientTxnId, {
        observations: [
          ...entry.observations,
          freezeWorkbookValue(structuredClone(message)),
        ],
        ...(entry.receipt ? { refresh: "required" as const } : {}),
      });
      if (entry.receipt) void this.retryRefresh(id);
    }
  }
  async retryRefresh(id: string) {
    const checkpoint = this.checkpoints.get(id),
      reader = this.reader,
      authorityReader = this.authorityReader,
      baseline = this.authority,
      lifetime = this.lifetime,
      generation = this.generation;
    if (
      !checkpoint?.create.receipt ||
      !reader ||
      !authorityReader ||
      !baseline ||
      this.refreshes.has(id)
    )
      return;
    const accepted = [checkpoint.create, ...checkpoint.links].filter(
      (entry) => !!entry.receipt,
    );
    this.refreshes.add(id);
    for (const entry of accepted)
      this.replaceStage(id, entry.attempt.clientTxnId, {
        refresh: "refreshing",
      });
    try {
      const current = await boundedRead(
        (signal) => authorityReader(baseline, signal),
        new AbortController().signal,
      );
      if (lifetime !== this.lifetime || generation !== this.generation)
        throw new Error("Authority changed.");
      this.setAuthority(current);
      if (
        !this.authority ||
        current.actorId !== checkpoint.create.attempt.review.authority.actorId
      )
        throw new Error("Authority changed.");
      const currentGeneration = this.generation;
      const sourceId = checkpoint.create.attempt.review.draft.source.recordId;
      const targetId = checkpoint.create.receipt.data.row.record_id;
      const results = await Promise.allSettled([
        readWorkbookAuthoringRecord(
          reader,
          timelineViewSchemaId,
          sourceId,
          new AbortController().signal,
        ),
        this.readTarget(checkpoint),
      ]);
      if (lifetime !== this.lifetime || currentGeneration !== this.generation)
        throw new Error("Authority changed.");
      const source =
        results[0].status === "fulfilled" ? results[0].value : undefined;
      const target =
        results[1].status === "fulfilled" ? results[1].value : undefined;
      if (
        this.checkpoints.get(id)?.observationRevision !==
        checkpoint.observationRevision
      )
        throw new Error(
          "Newer acceptance or socket observation requires another read.",
        );
      if (source === null) this.rows.delete(sourceId);
      if (target === null)
        this.rows.delete(checkpoint.create.receipt.data.row.record_id);
      for (const row of [source, target]) {
        if (
          !row ||
          row.row_version < (this.versions.get(row.record_id) ?? 0) ||
          row.row_version <= (this.removed.get(row.record_id) ?? 0)
        )
          continue;
        this.versions.set(row.record_id, row.row_version);
        this.removed.delete(row.record_id);
        // Fresh reads may update derived counts without a source-version increment.
        this.rows.set(row.record_id, freezeWorkbookValue(structuredClone(row)));
      }
      this.replace(id, {
        ...(source !== undefined
          ? {
              sourceUnavailable:
                source === null ||
                source.row_version <= (this.removed.get(sourceId) ?? 0) ||
                source.cells["timeline.capture_state"]?.value === "superseded",
            }
          : {}),
        ...(target !== undefined
          ? { evidenceUnavailable: target === null }
          : {}),
        ...(source && target
          ? {
              associationPresent: this.hasAssociation(source, target.record_id),
            }
          : {}),
      });
      if (
        results.some((result) => result.status === "rejected") ||
        [source, target].some(
          (row) =>
            row && row.row_version < (this.versions.get(row.record_id) ?? 0),
        )
      )
        throw new Error("Projection reads failed.");
      await boundedRead(
        async () =>
          this.effects?.refresh(
            [timelineViewSchemaId, evidenceViewSchemaId, partiesViewSchemaId],
            [sourceId, targetId],
          ),
        new AbortController().signal,
      );
      if (lifetime !== this.lifetime || currentGeneration !== this.generation)
        throw new Error("Authority changed.");
      if (
        this.checkpoints.get(id)?.observationRevision !==
        checkpoint.observationRevision
      )
        throw new Error("A newer observation needs refresh.");
      for (const entry of accepted)
        this.replaceStage(id, entry.attempt.clientTxnId, {
          refresh: "complete",
        });
      const latest = this.checkpoints.get(id);
      if (latest?.message?.startsWith("Accepted writes are retained."))
        this.replace(id, {
          message: this.linkComplete(latest)
            ? "Evidence created and linked. Views refreshed."
            : "Evidence created; Timeline link incomplete.",
        });
    } catch {
      if (lifetime === this.lifetime) {
        for (const entry of accepted)
          this.replaceStage(id, entry.attempt.clientTxnId, {
            refresh: "required",
          });
        this.replace(id, {
          message:
            "Accepted writes are retained. Projection refresh is incomplete; retry refresh sends reads only.",
        });
      }
    } finally {
      if (lifetime === this.lifetime) {
        this.refreshes.delete(id);
        this.publish();
        if (
          this.authority &&
          this.checkpoints.get(id)?.observationRevision !==
            checkpoint.observationRevision
        )
          void this.retryRefresh(id);
      }
    }
  }
}
