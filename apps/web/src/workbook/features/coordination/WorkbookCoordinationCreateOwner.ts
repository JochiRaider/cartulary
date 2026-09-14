import type { InspectorFeatureGroup } from "@cartulary/view-contracts";
import { getViewContract } from "@cartulary/view-contracts";
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
import type {
  WorkbookAuthoringAuthorityReader,
  WorkbookAuthoringReadPort,
} from "../../ports/WorkbookAuthoringReadPort";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import { freezeWorkbookValue } from "../../utils/freezeWorkbookValue";
import {
  type CoordinationDraft,
  type CoordinationSource,
  coordinationFeature,
  coordinationIds,
  coordinationReferenceView,
  coordinationSource,
  coordinationSourceInput,
  coordinationTarget,
  coordinationVariant,
  prepareCoordination,
} from "./coordinationCreateModel";
import type {
  CoordinationEntry,
  CoordinationOutcome,
  CoordinationReceipt,
  CoordinationTransport,
} from "./coordinationCreateOperation";

type Snapshot = Readonly<{
  authority: WorkbookMutationAuthority | null;
  draft: CoordinationDraft | null;
  attachment: symbol | null;
  errors: Readonly<Record<string, string>>;
  message: string | null;
  candidateRevision: number;
  needsReview: boolean;
  preparing: boolean;
  entries: readonly CoordinationEntry[];
}>;
/** One coordination author per account/incident; presentation never owns its draft. */
export class WorkbookCoordinationCreateOwner {
  private transport: CoordinationTransport | null = null;
  private readonly entries = new Map<string, CoordinationEntry>();
  private readonly dispatches = new Map<string, number>();
  private readonly refreshes = new Set<string>();
  private readonly notified = new Set<string>();
  private sourceCoordinator:
    | ((recordId: string, signal: AbortSignal) => Promise<boolean>)
    | null = null;
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private lifetime = 0;
  private generation = 0;
  private reader: WorkbookAuthoringReadPort | null = null;
  private authorityReader: WorkbookAuthoringAuthorityReader | null = null;
  private draft: CoordinationDraft | null = null;
  private attachment: symbol | null = null;
  private nextDraft = 0;
  private preparing = false;
  private candidateRevision = 0;
  private reviewRevision = 0;
  private reviewedRevision = 0;
  private errors: Readonly<Record<string, string>> = {};
  private message: string | null = null;
  private readonly versions = new Map<string, number>();
  private readonly removed = new Map<string, number>();
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
    private readonly ids: SecureTransactionIdPort = {
      create: () => {
        throw new Error("Transaction identity is unavailable.");
      },
    },
    private readonly effects?: {
      coordinate(
        source: CoordinationSource,
        signal: AbortSignal,
      ): Promise<WorkbookSourceWriteSettlement>;
      accepted(receipt: CoordinationReceipt, id: string): void;
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
    reader: WorkbookAuthoringReadPort,
    authorityReader: WorkbookAuthoringAuthorityReader,
    transport?: CoordinationTransport,
  ) {
    if (transport) this.transport = transport;
    this.reader = reader;
    this.authorityReader = authorityReader;
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
    else this.attachment = null;
    this.generation++;
    this.reviewRevision++;
    this.candidateRevision++;
    if (this.draft)
      this.draft = freezeWorkbookValue({
        ...this.draft,
        labels: {},
        source: this.draft.source ? { ...this.draft.source, label: "" } : null,
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
    this.entries.clear();
    this.dispatches.clear();
    this.refreshes.clear();
    this.notified.clear();
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
    this.removed.clear();
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
  registerSourceCoordinator(
    coordinate: (recordId: string, signal: AbortSignal) => Promise<boolean>,
  ) {
    this.sourceCoordinator = coordinate;
    return () => {
      if (this.sourceCoordinator === coordinate) this.sourceCoordinator = null;
    };
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
        "Your coordination draft is retained. Resume it, or explicitly discard it before starting another.";
      this.publish();
      return false;
    }
    const variant = coordinationVariant(feature.featureGroupKey);
    if (
      !variant ||
      coordinationFeature(subject.subject.viewSchemaId, variant) !== feature ||
      subject.subject.rowVersion < 1
    )
      return false;
    this.draft = freezeWorkbookValue({
      id: ++this.nextDraft,
      revision: 0,
      actorId: this.authority.actorId,
      incidentId: this.incidentId,
      variant,
      target: coordinationTarget(variant),
      values: {},
      labels: {},
      source: coordinationSource(subject),
      origin: {
        sheetRef: structuredClone(sheetRef),
        subject: structuredClone(subject.subject),
        feature,
      },
    });
    this.attachment = attachment;
    this.errors = {};
    this.message = null;
    this.reviewedRevision = this.reviewRevision;
    this.versions.set(
      subject.subject.recordId,
      Math.max(
        subject.subject.rowVersion,
        this.versions.get(subject.subject.recordId) ?? 0,
      ),
    );
    this.publish();
    return true;
  }
  /** Bind presentation callbacks to the reviewed account lifetime and draft identity. */
  captureDraftActions(attachment: symbol | null) {
    const id = this.draft?.id,
      lifetime = this.lifetime,
      generation = this.generation;
    const current = () =>
      !!this.authority &&
      id !== undefined &&
      this.draft?.id === id &&
      this.lifetime === lifetime &&
      this.generation === generation &&
      this.attachment === attachment;
    return {
      update: (
        field: string,
        value: string | null,
        labels: Readonly<Record<string, string>> = {},
      ) => {
        if (current()) this.update(field, value, labels);
      },
      omit: (field: string) => {
        if (current()) this.omit(field);
      },
      changeSource: (source: CoordinationSource | null) => {
        if (current()) this.changeSource(source);
      },
      validate: () => current() && this.validate(),
      review: async () => {
        if (current()) await this.review();
      },
      resume: (target: symbol) => {
        if (current()) this.resume(target);
      },
      detach: () => {
        if (current() && attachment) this.detach(attachment);
      },
      discard: () => {
        if (current()) this.discard();
      },
    };
  }
  update(
    field: string,
    value: string | null,
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
  omit(field: string) {
    if (!this.draft || this.busy || !this.canSubmit()) return;
    const values = { ...this.draft.values };
    delete values[field];
    this.draft = freezeWorkbookValue({
      ...this.draft,
      revision: this.draft.revision + 1,
      values,
    });
    this.errors = {};
    this.publish();
  }
  changeSource(source: CoordinationSource | null) {
    if (
      !this.draft ||
      this.busy ||
      !this.canSubmit() ||
      (source &&
        (!coordinationFeature(source.viewSchemaId, this.draft.variant) ||
          source.rowVersion < 1))
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
    if (
      this.draft &&
      (!recordId ||
        recordId === this.draft.source?.recordId ||
        this.draft.target.fields.some(
          (f) =>
            coordinationReferenceView(f) &&
            coordinationIds(this.draft?.values[f.fieldKey]).includes(recordId),
        ))
    ) {
      this.reviewRevision++;
      this.draft = freezeWorkbookValue({
        ...this.draft,
        labels: {},
        source: this.draft.source ? { ...this.draft.source, label: "" } : null,
      });
    }
    this.publish();
  }
  validate() {
    if (!this.draft) return false;
    this.errors = prepareCoordination(this.draft, "validation-only").errors;
    this.publish();
    return !Object.keys(this.errors).length;
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
    draft: CoordinationDraft,
    acceptVersion: boolean,
  ) {
    const reader = this.reader,
      lifetime = this.lifetime;
    if (!reader) throw new Error("Creation discovery is unavailable.");
    const signal = new AbortController().signal;
    const authority = await this.currentAuthority(draft.actorId, true, signal);
    if (!this.canSubmit())
      throw new Error(
        "Creation requires an active incident and current edit access.",
      );
    const generation = this.generation,
      revision = this.reviewRevision;
    const view =
      draft.source?.viewSchemaId ?? draft.origin.subject.viewSchemaId;
    const feature = coordinationFeature(view, draft.variant);
    if (!feature) throw new Error("Source capability is unavailable.");
    await boundedRead(
      (observed) =>
        reader.verify(
          { source: { viewSchemaId: view }, target: draft.target, feature },
          observed,
        ),
      signal,
    );
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
          "Source is unavailable. Choose another source or explicitly clear it; your draft is retained.",
        );
      if (row.row_version < (this.versions.get(source.recordId) ?? 0))
        throw new Error(
          "Source projection is behind a known change. Retry review.",
        );
      if (!acceptVersion && row.row_version !== source.rowVersion) {
        this.reviewRevision++;
        throw new Error("Source changed. Review before creating.");
      }
      source = { ...source, rowVersion: row.row_version };
    }
    const errors: Record<string, string> = {};
    const requirements = new Map<string, Set<string>>();
    for (const field of draft.target.fields) {
      const view = coordinationReferenceView(field);
      if (!view) continue;
      const ids = coordinationIds(draft.values[field.fieldKey]);
      if (!ids.length) continue;
      const set = requirements.get(view) ?? new Set<string>();
      for (const id of ids) set.add(id);
      requirements.set(view, set);
    }
    for (const [view, ids] of requirements) {
      await boundedRead(async (signal) => {
        let cursor: string | null = null;
        const visited = new Set<string>();
        do {
          const result = await boundedRead(
            (observed) =>
              reader.page({
                viewSchemaId: view,
                cursor,
                signal: observed,
                queryState: emptyWorkbookQueryState(),
              }),
            signal,
          );
          if (result.kind !== "accepted")
            throw new Error("References could not be verified. Retry review.");
          for (const candidate of result.value.candidates)
            if (
              !candidate.row ||
              (candidate.row.row_version >
                (this.removed.get(candidate.recordId) ?? 0) &&
                candidate.row.row_version >=
                  (this.versions.get(candidate.recordId) ?? 0))
            )
              ids.delete(candidate.recordId);
          if (!ids.size || !result.value.hasMore) break;
          cursor = result.value.nextCursor;
          if (!cursor || visited.has(cursor))
            throw new Error("Reference paging changed. Retry review.");
          visited.add(cursor);
        } while (!signal.aborted);
      }, signal);
      for (const field of draft.target.fields)
        if (
          coordinationReferenceView(field) === view &&
          coordinationIds(draft.values[field.fieldKey]).some((id) =>
            ids.has(id),
          )
        )
          errors[field.fieldKey] =
            "A selected reference is unavailable. Remove it or retry review.";
    }
    if (
      lifetime !== this.lifetime ||
      generation !== this.generation ||
      revision !== this.reviewRevision ||
      this.draft?.id !== draft.id ||
      this.draft.revision !== draft.revision
    )
      throw new Error("Context changed. Review the retained draft.");
    if (Object.keys(errors).length) {
      this.errors = errors;
      throw new Error(
        "Review unavailable references. Your selections are retained.",
      );
    }
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
      this.message = "Source, references and creation access reviewed.";
    } catch (error) {
      if (lifetime === this.lifetime) {
        this.message =
          error instanceof Error
            ? error.message
            : "Review could not be completed.";
        this.reviewRevision++;
      }
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
        "Review current source and access before creating this retained Coordination.";
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
            "Resolve earlier source saves or conflicts before creating. Your Coordination is retained.",
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
      const attempt = transport.capture(
        review,
        this.ids.create("coordination-create"),
      );
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
      : "The Coordination could not be prepared. Your draft is retained.";
  }
  private replace(id: string, patch: Partial<CoordinationEntry>) {
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
  private receive(id: string, outcome: CoordinationOutcome, dispatch: number) {
    const entry = this.entries.get(id);
    if (!entry || entry.receipt) return;
    if (outcome.kind === "accepted") {
      const receipt = outcome.receipt;
      if (
        typeof receipt.meta?.request_id !== "string" ||
        !receipt.meta.request_id.trim() ||
        !acceptedRecordMutation(
          receipt.data,
          entry.attempt.review.draft.target.viewSchemaId,
        ) ||
        (entry.attempt.review.draft.source
          ? !("source_record_id" in receipt.data) ||
            receipt.data.source_record_id !==
              entry.attempt.review.draft.source.recordId ||
            !("link_type" in receipt.data) ||
            receipt.data.link_type !== "references_artifact"
          : "source_record_id" in receipt.data || "link_type" in receipt.data)
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
          message: "Coordination created. Refreshing views…",
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
      this.message = "Coordination created.";
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
      if (
        outcome.failure.kind === "validation" &&
        this.draft?.id === entry.attempt.review.draft.id
      ) {
        this.errors = Object.fromEntries(
          (outcome.failure.fields ?? []).flatMap(({ field }) =>
            field === coordinationSourceInput
              ? [
                  [
                    field,
                    "Source is unavailable. Review it, choose another source or clear it.",
                  ],
                ]
              : this.draft?.target.fieldMap[field]?.createWritable
                ? [
                    [
                      field,
                      "This value could not be accepted. Review it and try again.",
                    ],
                  ]
                : [],
          ),
        );
      }
      if (outcome.failure.kind !== "validation") this.reviewRevision++;
    } else
      this.replace(id, {
        phase: "uncertain",
        uncertain: true,
        transportPending: false,
        message:
          "The result is unconfirmed. Recover this exact submission before editing or creating another Coordination.",
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
      let outcome: CoordinationOutcome;
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
          entry.attempt.review.draft.target.viewSchemaId,
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
      const current = this.entries.get(id);
      const complete =
        current?.observations.length === entry.observations.length;
      this.replace(id, {
        refresh: complete ? "complete" : "required",
        message: complete
          ? "Coordination created. Views refreshed."
          : "Refreshing additional affected views…",
      });
    } catch {
      if (lifetime === this.lifetime)
        this.replace(id, {
          refresh: "required",
          message:
            "Coordination created, but views need refresh. Retry refresh sends reads only.",
        });
    } finally {
      if (lifetime === this.lifetime) {
        this.refreshes.delete(id);
        const current = this.entries.get(id);
        if (
          current?.refresh === "required" &&
          current.observations.length !== entry.observations.length
        )
          void this.retryRefresh(id);
      }
    }
  }
  observeSocket(message: RecordChangedMessage) {
    if (!this.authority || message.incident_id !== this.incidentId) return;
    const payload = message.payload;
    const previousRemoval = this.removed.get(payload.record_id);
    if (payload.row_version >= (this.versions.get(payload.record_id) ?? 0)) {
      if (
        payload.affected_views.some(
          (view) =>
            view.change_kind === "remove" &&
            getViewContract(view.view_schema_id) !== undefined,
        )
      )
        this.removed.set(payload.record_id, payload.row_version);
      else if (payload.row_version > (this.removed.get(payload.record_id) ?? 0))
        this.removed.delete(payload.record_id);
    }
    this.observe(payload.record_id, payload.row_version);
    if (previousRemoval !== this.removed.get(payload.record_id)) this.observe();
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
      ...(entry.receipt ? { refresh: "required" as const } : {}),
    });
    if (entry.receipt) void this.retryRefresh(payload.client_txn_id);
  }
}
