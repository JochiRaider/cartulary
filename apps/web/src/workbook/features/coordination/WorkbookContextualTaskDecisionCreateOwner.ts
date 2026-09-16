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
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import {
  type ContextualCreateDraft,
  contextualCreateDraft,
  contextualCreateErrors,
  contextualSelectedIds,
  freezeContextualCreate,
} from "./contextualCreateModel";
import type {
  ContextualAuthorityReader,
  ContextualCreateEntry,
  ContextualCreateOutcome,
  ContextualCreateReader,
  ContextualCreateReceipt,
  ContextualCreateTransport,
} from "./contextualCreateOperation";

type Snapshot = Readonly<{
  authority: WorkbookMutationAuthority | null;
  draft: ContextualCreateDraft | null;
  attachment: symbol | null;
  errors: Readonly<Record<string, string>>;
  message: string | null;
  candidateRevision: number;
  reviewRevision: number;
  needsReview: boolean;
  preparing: boolean;
  entries: readonly ContextualCreateEntry[];
}>;

/** Owns contextual intent independently of sheet selection and inspector lifetime. */
export class WorkbookContextualTaskDecisionCreateOwner {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private lifetime = 0;
  private preparing = false;
  private transport: ContextualCreateTransport | null = null;
  private readonly entries = new Map<string, ContextualCreateEntry>();
  private readonly dispatches = new Map<string, number>();
  private readonly uncertainAttempts = new Set<string>();
  private readonly refreshes = new Set<string>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  private readonly removedViews = new Map<string, number>();
  private nextDraftId = 0;
  private draft: ContextualCreateDraft | null = null;
  private attachment: symbol | null = null;
  private errors: Readonly<Record<string, string>> = {};
  private message: string | null = null;
  private candidateRevision = 0;
  private reviewRevision = 0;
  private reviewedRevision = 0;
  private readonly listeners = new Set<() => void>();
  private readonly versions = new Map<string, number>();
  private reader: ContextualCreateReader | null = null;
  private authorityReader: ContextualAuthorityReader | null = null;
  private snapshot: Snapshot = {
    authority: null,
    draft: null,
    attachment: null,
    errors: {},
    message: null,
    candidateRevision: 0,
    reviewRevision: 0,
    needsReview: false,
    preparing: false,
    entries: [],
  };
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly effects: {
      readonly coordinate: (
        draft: ContextualCreateDraft,
        signal: AbortSignal,
      ) => Promise<WorkbookSourceWriteSettlement>;
      readonly accepted: (
        receipt: ContextualCreateReceipt,
        clientTxnId: string,
      ) => void;
      readonly refresh: (
        draft: ContextualCreateDraft,
        viewSchemaIds: readonly string[],
      ) => Promise<void>;
      readonly observed: (viewSchemaId: string, row: WorkbookQueryRow) => void;
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
    reader: ContextualCreateReader,
    authorityReader: ContextualAuthorityReader,
    transport?: ContextualCreateTransport,
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
      reviewRevision: this.reviewRevision,
      needsReview: this.reviewedRevision !== this.reviewRevision,
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
        ? freezeContextualCreate({ ...authority })
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.generation++;
    this.reviewRevision++;
    this.candidateRevision++;
    // Display labels are authorization observations, not durable reference identity.
    if (this.draft)
      this.draft = freezeContextualCreate({ ...this.draft, labels: {} });
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
    this.versions.clear();
    this.rows.clear();
    this.removedViews.clear();
    this.entries.clear();
    this.dispatches.clear();
    this.uncertainAttempts.clear();
    this.refreshes.clear();
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
  canReplay() {
    return (
      !!this.authority?.sessionIdentity &&
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
    return [...this.entries.values()].filter(
      (entry) => entry.phase === "uncertain" || entry.refresh === "required",
    ).length;
  }
  get uncertainCount() {
    return [...this.entries.values()].filter(
      (entry) => entry.phase === "uncertain",
    ).length;
  }
  begin(
    subject: WorkbookInspectorLiveRowBinding,
    feature: InspectorFeatureGroup,
    sheetRef: SheetRef,
    attachment: symbol,
  ) {
    if (!this.canSubmit() || !this.authority) return false;
    if (this.draft || this.busy) {
      this.message =
        "Your contextual draft is retained. Resume it, or explicitly discard it before starting another.";
      this.publish();
      return false;
    }
    const draft = contextualCreateDraft(
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
    this.reviewedRevision = this.reviewRevision;
    this.versions.set(
      draft.source.recordId,
      Math.max(
        draft.source.rowVersion,
        this.versions.get(draft.source.recordId) ?? 0,
      ),
    );
    this.publish();
    return true;
  }
  update(
    fieldKey: string,
    value: string,
    labels: Readonly<Record<string, string>> = {},
  ) {
    if (!this.draft || this.busy || !this.canSubmit()) return;
    if (
      !this.draft.target.fieldMap[fieldKey]?.createWritable &&
      !this.draft.target.createInputs.some(
        (input) => input.inputKey === fieldKey,
      )
    )
      return;
    this.draft = freezeContextualCreate({
      ...this.draft,
      revision: this.draft.revision + 1,
      values: { ...this.draft.values, [fieldKey]: value },
      labels: { ...this.draft.labels, ...labels },
    });
    this.errors = {};
    this.message = null;
    this.publish();
  }
  resume(attachment: symbol) {
    if (this.draft && this.authority) {
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
        contextualSelectedIds(this.draft).includes(recordId))
    ) {
      this.reviewRevision++;
      const labels = { ...this.draft.labels };
      if (recordId) delete labels[recordId];
      this.draft = freezeContextualCreate({
        ...this.draft,
        labels: recordId ? labels : {},
      });
    }
    this.publish();
  }
  observeSocket(message: RecordChangedMessage) {
    if (!this.authority || message.incident_id !== this.incidentId) return;
    const payload = message.payload;
    if (payload.row_version >= (this.versions.get(payload.record_id) ?? 0)) {
      for (const view of payload.affected_views) {
        const key = `${view.view_schema_id}:${payload.record_id}`;
        if (view.change_kind === "remove")
          this.removedViews.set(key, payload.row_version);
        else if (payload.row_version > (this.removedViews.get(key) ?? 0))
          this.removedViews.delete(key);
      }
    }
    this.observe(payload.record_id, payload.row_version);
    const entry = this.entries.get(payload.client_txn_id);
    if (
      !entry ||
      payload.actor_user_id !== entry.attempt.review.authority.actorId ||
      entry.observations.some((prior) => prior.event_id === message.event_id)
    )
      return;
    if (
      entry.receipt &&
      payload.change_set_id !== entry.receipt.data.change_set_id
    )
      return;
    const knownViews = new Set([
      entry.attempt.review.draft.source.viewSchemaId,
      entry.attempt.review.draft.target.viewSchemaId,
      ...entry.observations.flatMap((event) =>
        event.payload.affected_views.map((view) => view.view_schema_id),
      ),
    ]);
    const additionalViews = payload.affected_views.some(
      (view) =>
        getViewContract(view.view_schema_id) &&
        !knownViews.has(view.view_schema_id),
    );
    const extendCompletedRefresh =
      entry.receipt && entry.refresh === "complete" && additionalViews;
    this.replace(payload.client_txn_id, {
      observations: [
        ...entry.observations,
        freezeContextualCreate(structuredClone(message)),
      ],
      ...(extendCompletedRefresh ? { refresh: "required" as const } : {}),
    });
    // An echo cannot restart a failed refresh or race the explicit recovery action.
    // An in-flight refresh incorporates additional views before it completes.
    if (extendCompletedRefresh) void this.retryRefresh(payload.client_txn_id);
  }
  async recheckAuthority() {
    const baseline = this.authority,
      generation = this.generation,
      reader = this.authorityReader;
    if (!baseline || !reader) return;
    try {
      const current = await boundedRead(
        (signal) => reader(baseline, signal),
        new AbortController().signal,
      );
      if (generation === this.generation) this.setAuthority(current);
    } catch {
      /* Only the authority reader can establish incident/session access loss. */
    }
  }
  async review(acceptSourceVersion = true) {
    const reader = this.reader,
      authorityReader = this.authorityReader;
    const draft = this.draft,
      generation = this.generation,
      revision = this.reviewRevision;
    if (!draft || !this.authority || !reader || !authorityReader) return false;
    try {
      const baseline = this.authority;
      const current = await boundedRead(
        (signal) => authorityReader(baseline, signal),
        new AbortController().signal,
      );
      if (generation !== this.generation) return false;
      this.setAuthority(current);
      if (
        !this.canSubmit() ||
        current.actorId !== draft.actorId ||
        revision !== this.reviewRevision
      )
        return false;
      await boundedRead(
        (signal) => reader.verify(draft, signal),
        new AbortController().signal,
      );
      const row = await readWorkbookAuthoringRecord(
        reader,
        draft.source.viewSchemaId,
        draft.source.recordId,
        new AbortController().signal,
      );
      if (
        !row ||
        row.record_id !== draft.source.recordId ||
        row.row_version <
          Math.max(
            draft.source.rowVersion,
            this.versions.get(draft.source.recordId) ?? 0,
          ) ||
        row.row_version <=
          (this.removedViews.get(
            `${draft.source.viewSchemaId}:${draft.source.recordId}`,
          ) ?? 0)
      )
        throw new Error("The retained source could not be verified.");
      if (!acceptSourceVersion && row.row_version !== draft.source.rowVersion)
        throw new Error("Source changed. Review before creating.");
      if (
        this.draft !== draft ||
        generation !== this.generation ||
        revision !== this.reviewRevision
      )
        return false;
      if (row.row_version !== draft.source.rowVersion) {
        this.draft = freezeContextualCreate({
          ...draft,
          source: { ...draft.source, rowVersion: row.row_version },
        });
      }
      this.versions.set(row.record_id, row.row_version);
      this.reviewedRevision = this.reviewRevision;
      this.errors = contextualCreateErrors(draft);
      this.message = Object.keys(this.errors).length
        ? "Complete the highlighted fields."
        : null;
      this.publish();
      return Object.keys(this.errors).length === 0;
    } catch {
      if (generation === this.generation) {
        this.message =
          "Context could not be verified. Your values are retained; retry review.";
        this.reviewRevision++;
        this.publish();
      }
      return false;
    }
  }
  latestVersion(recordId: string) {
    return this.authority ? (this.versions.get(recordId) ?? null) : null;
  }
  latestRow(recordId: string, viewSchemaId?: string) {
    const row = this.rows.get(recordId);
    return this.authority &&
      row &&
      (!viewSchemaId ||
        row.row_version >
          (this.removedViews.get(`${viewSchemaId}:${recordId}`) ?? 0)) &&
      row.row_version >= (this.versions.get(recordId) ?? 0)
      ? row
      : null;
  }
  acceptRow(row: WorkbookQueryRow, viewSchemaId?: string) {
    if (
      !this.authority ||
      (viewSchemaId !== undefined &&
        row.row_version <=
          (this.removedViews.get(`${viewSchemaId}:${row.record_id}`) ?? 0)) ||
      row.row_version < (this.versions.get(row.record_id) ?? 0)
    )
      return this.latestRow(row.record_id, viewSchemaId);
    const prior = this.rows.get(row.record_id);
    if (prior && prior.row_version >= row.row_version) return prior;
    this.observe(row.record_id, row.row_version);
    const accepted = freezeContextualCreate(structuredClone(row));
    this.rows.set(row.record_id, accepted);
    return accepted;
  }
  async submit(attachment: symbol) {
    if (
      this.busy ||
      !this.draft ||
      this.attachment !== attachment ||
      !this.canSubmit() ||
      !this.transport ||
      this.reviewedRevision !== this.reviewRevision
    )
      return;
    const draft = this.draft,
      lifetime = this.lifetime,
      reviewedRevision = this.reviewRevision;
    this.preparing = true;
    this.message = null;
    this.publish();
    try {
      const ready = await boundedRead(
        (signal) => this.effects.coordinate(draft, signal),
        new AbortController().signal,
      );
      if (ready.kind !== "settled")
        throw new Error(
          "Resolve earlier source saves before creating. Your draft is retained.",
        );
      this.versions.set(
        draft.source.recordId,
        Math.max(
          this.versions.get(draft.source.recordId) ?? 0,
          ready.minimumRowVersion,
        ),
      );
      if (
        this.lifetime !== lifetime ||
        this.draft !== draft ||
        this.attachment !== attachment ||
        reviewedRevision !== this.reviewRevision
      )
        return;
      if (
        !(await this.review(false)) ||
        this.draft !== draft ||
        this.attachment !== attachment ||
        !this.authority ||
        lifetime !== this.lifetime
      )
        return;
      const finalSettlement = await boundedRead(
        (signal) => this.effects.coordinate(draft, signal),
        new AbortController().signal,
      );
      if (
        this.lifetime !== lifetime ||
        this.draft !== draft ||
        this.attachment !== attachment ||
        !this.authority ||
        reviewedRevision !== this.reviewRevision
      )
        return;
      if (
        finalSettlement.kind !== "settled" ||
        finalSettlement.minimumRowVersion > draft.source.rowVersion ||
        (this.versions.get(draft.source.recordId) ?? 0) >
          draft.source.rowVersion
      ) {
        this.reviewRevision++;
        throw new Error(
          "The source changed during preparation. Your values are retained; review again.",
        );
      }
      const attempt = this.transport.capture(
        { authority: this.authority, draft },
        this.ids.create("contextual-create"),
      );
      this.entries.set(
        attempt.clientTxnId,
        freezeContextualCreate({
          observations: [],
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
        this.message =
          error instanceof Error
            ? error.message
            : "Creation could not be prepared. Your draft is retained.";
    } finally {
      if (lifetime === this.lifetime) {
        this.preparing = false;
        this.publish();
      }
    }
  }
  private replace(id: string, patch: Partial<ContextualCreateEntry>) {
    const entry = this.entries.get(id);
    if (!entry) return;
    this.entries.set(id, freezeContextualCreate({ ...entry, ...patch }));
    this.publish();
  }
  private receive(
    id: string,
    outcome: ContextualCreateOutcome,
    replay: boolean,
  ) {
    const entry = this.entries.get(id);
    if (!entry || entry.receipt) return;
    const candidateReceipt =
      outcome.kind === "accepted" ? outcome.receipt : null;
    if (
      outcome.kind === "accepted" &&
      (outcome.receipt.data.view_schema_id !==
        entry.attempt.review.draft.target.viewSchemaId ||
        !outcome.receipt.data.change_set_id ||
        !outcome.receipt.meta.request_id ||
        !outcome.receipt.data.row.record_id ||
        outcome.receipt.data.row.row_version < 1 ||
        entry.observations.some(
          (event) =>
            event.payload.change_set_id !==
            candidateReceipt?.data.change_set_id,
        ))
    )
      outcome = { kind: "uncertain" };
    if (outcome.kind === "accepted") {
      const receipt = freezeContextualCreate(structuredClone(outcome.receipt));
      // Commit the receipt before any presentation, query or history effects.
      this.replace(id, {
        phase: "accepted",
        receipt,
        refresh: "required",
        message: null,
      });
      if (this.draft?.id === entry.attempt.review.draft.id) {
        this.draft = null;
        this.attachment = null;
        this.errors = {};
      }
      this.message = `${entry.attempt.review.draft.target.title} record created.`;
      this.acceptRow(receipt.data.row, receipt.data.view_schema_id);
      this.publish();
      try {
        this.effects.accepted(receipt, id);
      } catch {
        /* Accepted receipt remains recoverable. */
      }
      if (this.authority) void this.retryRefresh(id);
    } else if (
      outcome.kind === "rejected" &&
      !replay &&
      !this.uncertainAttempts.has(id) &&
      entry.phase !== "uncertain"
    ) {
      this.replace(id, { phase: "rejected", message: outcome.failure.message });
      this.message = outcome.failure.message;
      this.errors =
        outcome.failure.kind === "validation"
          ? Object.fromEntries(
              (outcome.failure.fields ?? []).map((field) => [
                field.field,
                field.message,
              ]),
            )
          : {};
      this.reviewRevision++;
      this.publish();
      if (
        workbookFailureLifecycle(outcome.failure).kind ===
        "authority_unavailable"
      )
        void this.recheckAuthority();
    } else {
      this.uncertainAttempts.add(id);
      this.replace(id, {
        phase: "uncertain",
        message:
          outcome.kind === "rejected"
            ? outcome.failure.message
            : "Creation outcome is unconfirmed. Recover this operation before starting another.",
      });
    }
  }
  private async dispatch(id: string, replay: boolean) {
    const entry = this.entries.get(id),
      transport = this.transport,
      lifetime = this.lifetime;
    if (!entry || !transport) return;
    const dispatch = (this.dispatches.get(id) ?? 0) + 1;
    this.dispatches.set(id, dispatch);
    const observation = this.observeTransport(async (signal) => {
      const outcome = await transport.send(entry.attempt, signal);
      if (
        lifetime === this.lifetime &&
        (outcome.kind === "accepted" || this.dispatches.get(id) === dispatch)
      )
        this.receive(id, outcome, replay);
      return outcome;
    });
    void observation.settled.then(() => {
      if (lifetime === this.lifetime && this.dispatches.get(id) === dispatch)
        this.replace(id, { transportPending: false });
    });
    const result = await observation.result;
    if (
      lifetime === this.lifetime &&
      this.dispatches.get(id) === dispatch &&
      result.kind !== "completed" &&
      !this.entries.get(id)?.receipt
    ) {
      this.uncertainAttempts.add(id);
      this.replace(id, {
        phase: "uncertain",
        transportPending: false,
        message:
          "Creation outcome is unconfirmed. Recover the original operation.",
      });
    }
  }
  async replay(id: string) {
    const entry = this.entries.get(id),
      baseline = this.authority,
      reader = this.authorityReader,
      lifetime = this.lifetime,
      generation = this.generation;
    if (
      !entry ||
      entry.phase !== "uncertain" ||
      entry.transportPending ||
      this.preparing ||
      !this.canReplay() ||
      !baseline ||
      !reader
    )
      return;
    this.preparing = true;
    this.publish();
    try {
      const current = await boundedRead(
        (signal) => reader(baseline, signal),
        new AbortController().signal,
      );
      if (lifetime !== this.lifetime || generation !== this.generation) return;
      this.setAuthority(current);
      if (
        !this.canReplay() ||
        current.actorId !== entry.attempt.review.authority.actorId ||
        current.incidentId !== entry.attempt.review.authority.incidentId ||
        this.entries.get(id)?.receipt
      )
        return;
      const currentGeneration = this.generation;
      const capabilityReader = this.reader;
      if (!capabilityReader) return;
      await boundedRead(
        (signal) => capabilityReader.verify(entry.attempt.review.draft, signal),
        new AbortController().signal,
      );
      if (
        lifetime !== this.lifetime ||
        currentGeneration !== this.generation ||
        !this.canReplay() ||
        this.entries.get(id)?.receipt
      )
        return;
      this.replace(id, {
        phase: "submitting",
        transportPending: true,
        message: null,
      });
      this.preparing = false;
      await this.dispatch(id, true);
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
      authorityReader = this.authorityReader,
      baseline = this.authority,
      lifetime = this.lifetime;
    let generation = this.generation;
    let additionalViews = false;
    if (
      !entry?.receipt ||
      entry.refresh === "complete" ||
      this.refreshes.has(id) ||
      !baseline ||
      !reader ||
      !authorityReader
    )
      return;
    this.refreshes.add(id);
    this.replace(id, { refresh: "refreshing" });
    try {
      const current = await boundedRead(
        (signal) => authorityReader(baseline, signal),
        new AbortController().signal,
      );
      if (generation !== this.generation || lifetime !== this.lifetime)
        throw new Error("Authority changed");
      this.setAuthority(current);
      if (
        !this.authority ||
        current.actorId !== entry.attempt.review.authority.actorId
      )
        throw new Error("Authority changed");
      generation = this.generation;
      const draft = entry.attempt.review.draft;
      const views = [
        ...new Set([
          draft.source.viewSchemaId,
          draft.target.viewSchemaId,
          ...(this.entries.get(id)?.observations ?? [])
            .flatMap((event) =>
              event.payload.affected_views.map((view) => view.view_schema_id),
            )
            .filter((view) => getViewContract(view)),
        ]),
      ];
      await boundedRead(async (signal) => {
        const results = await Promise.all(
          views.map(async (viewSchemaId) => ({
            viewSchemaId,
            result: await reader.page({
              viewSchemaId,
              cursor: null,
              queryState: emptyWorkbookQueryState(),
              signal,
            }),
          })),
        );
        if (results.some(({ result }) => result.kind !== "accepted"))
          throw new Error("Refresh failed");
        if (generation !== this.generation || lifetime !== this.lifetime)
          throw new Error("Authority changed");
        for (const { viewSchemaId, result } of results)
          if (result.kind === "accepted")
            for (const candidate of result.value.candidates)
              if (candidate.row) {
                const row = this.acceptRow(candidate.row, viewSchemaId);
                if (row) this.effects.observed(viewSchemaId, row);
              }
        await this.effects.refresh(draft, views);
      }, new AbortController().signal);
      if (lifetime !== this.lifetime || generation !== this.generation)
        throw new Error("Authority changed");
      additionalViews = (this.entries.get(id)?.observations ?? []).some(
        (event) =>
          event.payload.affected_views.some(
            (view) =>
              getViewContract(view.view_schema_id) &&
              !views.includes(view.view_schema_id),
          ),
      );
      if (additionalViews) return;
      this.candidateRevision++;
      this.replace(id, { refresh: "complete", message: null });
    } catch {
      if (lifetime === this.lifetime)
        this.replace(id, {
          refresh: "required",
          message:
            "Created, but views need refresh. Retry refresh sends reads only.",
        });
    } finally {
      if (lifetime === this.lifetime) {
        this.refreshes.delete(id);
        if (additionalViews) void this.retryRefresh(id);
      }
    }
  }
}
