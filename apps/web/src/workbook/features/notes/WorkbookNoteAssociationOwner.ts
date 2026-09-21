import { requireViewContract } from "@cartulary/view-contracts";
import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import { readWorkbookAuthoringRecord } from "../../adapters/readWorkbookAuthoringRecord";
import { normalizeRecordMutationRow } from "../../adapters/workbookRecordPatchTransport";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookAuthoringAuthorityReader } from "../../ports/WorkbookAuthoringReadPort";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { freezeWorkbookValue } from "../../utils/freezeWorkbookValue";
import {
  type NoteAssociationEntry,
  type NoteAssociationKind,
  type NoteAssociationList,
  type NoteAssociationOutcome,
  type NoteAssociationReader,
  type NoteAssociationReceipt,
  type NoteAssociationReview,
  type NoteAssociationTransport,
  noteAssociationListKey,
  noteAssociationView,
} from "./noteAssociationOperation";

type Snapshot = Readonly<{
  authority: WorkbookMutationAuthority | null;
  generation: number;
  candidateRevision: number;
  preparing: readonly string[];
  entries: readonly NoteAssociationEntry[];
  lists: Readonly<Record<string, NoteAssociationList>>;
  errors: Readonly<Record<string, string>>;
}>;

/** Incident-scoped association attempts, independent of Note creation drafts and
 * inspector attachment. Complete receipts survive every read/presentation failure. */
export class WorkbookNoteAssociationOwner {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private lifetime = 0;
  private generation = 0;
  private candidateRevision = 0;
  private nextOrder = 0;
  private authorityReader: WorkbookAuthoringAuthorityReader | null = null;
  private reader: NoteAssociationReader | null = null;
  private transport: NoteAssociationTransport | null = null;
  private readonly entries = new Map<string, NoteAssociationEntry>();
  private readonly dispatches = new Map<string, number>();
  private readonly preparing = new Set<string>();
  private readonly refreshing = new Set<string>();
  private readonly notified = new Set<string>();
  private readonly lists = new Map<string, NoteAssociationList>();
  private readonly reads = new Map<string, number>();
  private readonly versions = new Map<string, number>();
  private readonly errors = new Map<string, string>();
  private readonly listeners = new Set<() => void>();
  private snapshot: Snapshot = {
    authority: null,
    generation: 0,
    candidateRevision: 0,
    preparing: [],
    entries: [],
    lists: {},
    errors: {},
  };
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly effects: {
      coordinate(
        recordId: string,
        signal: AbortSignal,
      ): Promise<WorkbookSourceWriteSettlement>;
      accepted(receipt: NoteAssociationReceipt, id: string): void;
      refresh(recordId: string): Promise<void>;
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
  getReader() {
    return this.reader;
  }
  configure(
    reader: NoteAssociationReader,
    authorityReader: WorkbookAuthoringAuthorityReader,
    transport: NoteAssociationTransport,
  ) {
    this.reader = reader;
    this.authorityReader = authorityReader;
    this.transport = transport;
  }
  private publish() {
    this.snapshot = {
      authority: this.authority,
      generation: this.generation,
      candidateRevision: this.candidateRevision,
      preparing: this.authority ? [...this.preparing] : [],
      entries: this.authority ? [...this.entries.values()] : [],
      lists: this.authority ? Object.fromEntries(this.lists) : {},
      errors: this.authority ? Object.fromEntries(this.errors) : {},
    };
    for (const listener of this.listeners) listener();
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && authority.actorId !== this.actorId))
    )
      this.retire();
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    this.authority =
      authority?.incidentId === this.incidentId
        ? freezeWorkbookValue({ ...authority })
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.generation++;
    this.candidateRevision++;
    this.lists.clear();
    this.reads.clear();
    this.errors.clear();
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
    this.candidateRevision++;
    this.authority = null;
    this.actorId = null;
    this.entries.clear();
    this.dispatches.clear();
    this.preparing.clear();
    this.refreshing.clear();
    this.notified.clear();
    this.lists.clear();
    this.reads.clear();
    this.versions.clear();
    this.errors.clear();
    this.publish();
  }
  canReplay() {
    return (
      !!this.authority?.sessionIdentity &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  canSubmit() {
    return this.canReplay() && !this.authority?.closed;
  }
  blocksRecord(recordId: string) {
    return (
      this.preparing.has(recordId) ||
      [...this.entries.values()].some(
        (entry) =>
          entry.attempt.review.row.record_id === recordId &&
          (entry.phase === "submitting" || entry.phase === "uncertain"),
      )
    );
  }
  get unsettledMutationCount() {
    return (
      this.preparing.size +
      [...this.entries.values()].filter(
        (entry) => entry.phase === "submitting" || entry.phase === "uncertain",
      ).length
    );
  }
  latestRow(recordId: string): WorkbookQueryRow | null {
    if (!this.authority) return null;
    return [...this.entries.values()].reduce<WorkbookQueryRow | null>(
      (latest, entry) => {
        const row =
          entry.receipt &&
          normalizeRecordMutationRow(
            entry.receipt.data.row,
            noteAssociationView,
            recordId,
          );
        return row && row.row_version > (latest?.row_version ?? 0)
          ? row
          : latest;
      },
      null,
    );
  }
  observe(recordId?: string, version?: number) {
    if (!this.authority) return;
    if (recordId && version !== undefined) {
      if (version <= (this.versions.get(recordId) ?? 0)) return;
      this.versions.set(recordId, version);
    }
    this.candidateRevision++;
    this.publish();
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
        "Current authorization is unavailable. Retry after access is restored.",
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
      throw new Error("Authorization changed. Review this Note again.");
    this.setAuthority(current);
    if (
      !this.authority ||
      current.actorId !== actorId ||
      (editing && !this.canReplay())
    )
      throw new Error("Current edit access could not be verified.");
    return current;
  }
  async read(
    noteId: string,
    kind: NoteAssociationKind,
    more = false,
  ): Promise<boolean> {
    const transport = this.transport,
      lifetime = this.lifetime,
      generation = this.generation;
    if (!this.authority || !transport) return false;
    const key = noteAssociationListKey(noteId, kind),
      previous = this.lists.get(key),
      page = previous?.page ?? null;
    if (more && !page?.next_cursor_token) return false;
    const sequence = (this.reads.get(key) ?? 0) + 1;
    this.reads.set(key, sequence);
    this.lists.set(key, {
      noteId,
      kind,
      state: page ? "refreshing" : "initial_loading",
      page,
      message: null,
    });
    this.publish();
    const current = () =>
      lifetime === this.lifetime &&
      generation === this.generation &&
      this.reads.get(key) === sequence &&
      !!this.authority;
    try {
      const result = await boundedRead(
        (signal) =>
          transport.list(
            noteId,
            kind,
            more ? (page?.next_cursor_token ?? null) : null,
            signal,
          ),
        new AbortController().signal,
      );
      if (!current()) return false;
      if (result.kind === "rejected") {
        if (
          workbookFailureLifecycle(result.failure).kind ===
          "authority_unavailable"
        )
          this.suspend();
        throw new Error(result.failure.message);
      }
      if (
        result.page.row_version < (this.versions.get(noteId) ?? 0) ||
        (more && page?.row_version !== result.page.row_version)
      )
        throw new Error(
          "This Note changed during paging. Refresh associations before continuing.",
        );
      const items =
        more && page
          ? [...page.items, ...result.page.items]
          : result.page.items;
      if (new Set(items.map((item) => item.item_ref)).size !== items.length)
        throw new Error(
          "The association page repeated an item. Refresh associations.",
        );
      this.versions.set(
        noteId,
        Math.max(this.versions.get(noteId) ?? 0, result.page.row_version),
      );
      this.lists.set(key, {
        noteId,
        kind,
        state: "ready",
        page: { ...result.page, items },
        message: null,
      });
      this.publish();
      return true;
    } catch (error) {
      if (current()) {
        this.lists.set(key, {
          noteId,
          kind,
          state: page ? "stale_failure" : "unavailable",
          page,
          message:
            error instanceof Error
              ? error.message
              : "Could not load associations. Retry the read.",
        });
        this.publish();
      }
      return false;
    }
  }
  async submit(input: Omit<NoteAssociationReview, "authority">) {
    const lifetime = this.lifetime,
      reader = this.reader,
      transport = this.transport,
      noteId = input.row.record_id,
      actor = this.authority?.actorId;
    if (
      !actor ||
      !reader ||
      !transport ||
      !this.canSubmit() ||
      this.blocksRecord(noteId) ||
      !input.actions.length ||
      input.actions.length > 64
    )
      return;
    this.preparing.add(noteId);
    this.errors.delete(noteAssociationListKey(noteId, input.kind));
    this.publish();
    try {
      const signal = new AbortController().signal;
      const settlement = await boundedRead(
        (observed) => this.effects.coordinate(noteId, observed),
        signal,
      );
      if (settlement.kind !== "settled")
        throw new Error(
          "Resolve earlier saves or conflicts before changing associations.",
        );
      const authority = await this.currentAuthority(actor, true, signal),
        generation = this.generation;
      if (!this.canSubmit())
        throw new Error("This incident is closed. Associations are read-only.");
      const target = requireViewContract(noteAssociationView);
      const feature = target.inspectorConfig.featureGroups.find(
        (feature) =>
          feature.mutates &&
          feature.routeBinding.owner === "note_associations_route" &&
          feature.routeBinding.actionKey === input.kind,
      );
      if (!feature)
        throw new Error("The association capability is unavailable.");
      await boundedRead(
        (observed) => reader.verify(input.kind, observed),
        signal,
      );
      const row = await readWorkbookAuthoringRecord(
        reader,
        noteAssociationView,
        noteId,
        signal,
      );
      if (
        !row ||
        row.row_version !== input.row.row_version ||
        row.row_version <
          Math.max(this.versions.get(noteId) ?? 0, settlement.minimumRowVersion)
      )
        throw new Error(
          "This Note changed. Refresh and review its associations before submitting.",
        );
      if (lifetime !== this.lifetime || generation !== this.generation) return;
      const attempt = transport.capture(
        { ...input, row, authority },
        this.ids.create("note-associations"),
      );
      this.entries.set(
        attempt.clientTxnId,
        freezeWorkbookValue({
          attempt,
          order: ++this.nextOrder,
          phase: "submitting",
          uncertain: false,
          transportPending: true,
          receipt: null,
          failure: null,
          refresh: "none",
        }),
      );
      this.preparing.delete(noteId);
      this.publish();
      await this.dispatch(attempt.clientTxnId);
    } catch (error) {
      if (lifetime === this.lifetime && this.authority)
        this.errors.set(
          noteAssociationListKey(noteId, input.kind),
          error instanceof Error
            ? error.message
            : "Could not prepare associations. Retry after reviewing this Note.",
        );
    } finally {
      if (lifetime === this.lifetime) {
        this.preparing.delete(noteId);
        this.publish();
      }
    }
  }
  private replace(id: string, patch: Partial<NoteAssociationEntry>) {
    const entry = this.entries.get(id);
    if (!entry) return;
    this.entries.set(id, freezeWorkbookValue({ ...entry, ...patch }));
    this.publish();
  }
  private acceptEffects(id: string) {
    const entry = this.entries.get(id);
    if (!entry?.receipt || !this.authority || this.notified.has(id)) return;
    this.effects.accepted(entry.receipt, id);
    this.notified.add(id);
  }
  private receive(
    id: string,
    outcome: NoteAssociationOutcome,
    dispatch: number,
  ) {
    const entry = this.entries.get(id);
    if (!entry || entry.receipt) return;
    if (outcome.kind === "accepted") {
      const receipt = outcome.receipt,
        row = normalizeRecordMutationRow(
          receipt.data.row,
          noteAssociationView,
          entry.attempt.review.row.record_id,
        );
      if (
        receipt.data.view_schema_id !== noteAssociationView ||
        !row ||
        !receipt.meta?.request_id?.trim() ||
        (receipt.data.change_set_id
          ? row.row_version <= entry.attempt.request.base_row_version
          : row.row_version !== entry.attempt.request.base_row_version)
      ) {
        this.receive(id, { kind: "uncertain" }, dispatch);
        return;
      }
      this.replace(id, {
        phase: "accepted",
        receipt: structuredClone(receipt),
        failure: null,
        transportPending: false,
        refresh: "required",
      });
      this.versions.set(
        row.record_id,
        Math.max(this.versions.get(row.record_id) ?? 0, row.row_version),
      );
      try {
        this.acceptEffects(id);
      } catch {
        /* Retained receipt permits read-only reconciliation. */
      }
      void this.retryRefresh(id);
      return;
    }
    if (this.dispatches.get(id) !== dispatch) return;
    if (outcome.kind === "rejected" && !entry.uncertain)
      this.replace(id, {
        phase: "rejected",
        transportPending: false,
        failure: outcome.failure,
      });
    else
      this.replace(id, {
        phase: "uncertain",
        uncertain: true,
        transportPending: false,
      });
    if (
      outcome.kind === "rejected" &&
      workbookFailureLifecycle(outcome.failure).kind === "authority_unavailable"
    )
      this.suspend();
  }
  private async dispatch(id: string) {
    const entry = this.entries.get(id),
      transport = this.transport,
      lifetime = this.lifetime;
    if (!entry || entry.receipt || !transport) return;
    const sequence = (this.dispatches.get(id) ?? 0) + 1;
    this.dispatches.set(id, sequence);
    const observed = this.observeTransport(async (signal) => {
      let outcome: NoteAssociationOutcome;
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
      lifetime === this.lifetime &&
      !this.entries.get(id)?.receipt &&
      this.dispatches.get(id) === sequence &&
      result.kind !== "completed"
    )
      this.receive(id, { kind: "uncertain" }, sequence);
  }
  async replay(id: string) {
    const entry = this.entries.get(id),
      lifetime = this.lifetime;
    if (
      !entry ||
      entry.phase !== "uncertain" ||
      entry.transportPending ||
      !this.canReplay() ||
      this.preparing.has(entry.attempt.review.row.record_id)
    )
      return;
    const noteId = entry.attempt.review.row.record_id;
    this.preparing.add(noteId);
    this.publish();
    try {
      await this.currentAuthority(
        entry.attempt.review.authority.actorId,
        true,
        new AbortController().signal,
      );
      if (lifetime !== this.lifetime || this.entries.get(id)?.receipt) return;
      this.replace(id, { phase: "submitting", transportPending: true });
      await this.dispatch(id);
    } catch {
      if (lifetime === this.lifetime)
        this.errors.set(
          noteAssociationListKey(noteId, entry.attempt.review.kind),
          "Current authorization could not be verified. Retry recovery.",
        );
    } finally {
      if (lifetime === this.lifetime) {
        this.preparing.delete(noteId);
        this.publish();
      }
    }
  }
  async retryRefresh(id: string) {
    const entry = this.entries.get(id),
      lifetime = this.lifetime;
    if (
      !entry?.receipt ||
      entry.refresh === "complete" ||
      this.refreshing.has(id) ||
      !this.authority
    )
      return;
    this.refreshing.add(id);
    this.replace(id, { refresh: "refreshing" });
    try {
      await this.currentAuthority(
        entry.attempt.review.authority.actorId,
        false,
        new AbortController().signal,
      );
      const generation = this.generation,
        noteId = entry.attempt.review.row.record_id;
      const results = await Promise.all(
        (["source", "evidence", "related_note"] as const).map((kind) =>
          this.read(noteId, kind),
        ),
      );
      if (results.some((ok) => !ok))
        throw new Error("Association refresh incomplete.");
      this.acceptEffects(id);
      await boundedRead(
        () => this.effects.refresh(noteId),
        new AbortController().signal,
      );
      if (lifetime !== this.lifetime || generation !== this.generation) return;
      // These reads reconcile every older accepted association attempt for the
      // same Note. Retaining a separate stale recovery item would misrepresent
      // data that has already been refreshed past that receipt's version.
      const refreshedVersion = Math.min(
        ...(["source", "evidence", "related_note"] as const).map(
          (kind) =>
            this.lists.get(noteAssociationListKey(noteId, kind))?.page
              ?.row_version ?? 0,
        ),
      );
      for (const [acceptedId, accepted] of this.entries) {
        if (
          accepted.receipt &&
          this.notified.has(acceptedId) &&
          accepted.attempt.review.row.record_id === noteId &&
          accepted.receipt.data.row.row_version <= refreshedVersion
        )
          this.replace(acceptedId, { refresh: "complete" });
      }
    } catch {
      if (
        lifetime === this.lifetime &&
        this.entries.get(id)?.refresh !== "complete"
      )
        this.replace(id, { refresh: "required" });
    } finally {
      if (lifetime === this.lifetime) this.refreshing.delete(id);
    }
  }
}
