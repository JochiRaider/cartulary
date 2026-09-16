import { requireViewContract } from "@cartulary/view-contracts";
import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import { acceptedRecordMutation } from "../../adapters/workbookRecordPatchTransport";
import { initialGenericCreateDraft } from "../../models/genericWorkbookModel";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type {
  WorkbookAuthoringAuthorityReader,
  WorkbookAuthoringCandidate,
  WorkbookAuthoringReadPort,
} from "../../ports/WorkbookAuthoringReadPort";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { freezeWorkbookValue } from "../../utils/freezeWorkbookValue";
import type {
  OrdinaryCreateContribution,
  OrdinaryCreateDraft,
} from "./ordinaryCreateContract";
import type {
  OrdinaryCreateEntry,
  OrdinaryCreateOutcome,
  OrdinaryCreateReceipt,
  OrdinaryCreateTransport,
} from "./ordinaryCreateOperation";

export type OrdinaryCreateSchemaSnapshot = Readonly<{
  draft: OrdinaryCreateDraft;
  values: Readonly<Record<string, string>>;
  errors: Readonly<Record<string, string>>;
  ready: boolean;
  busy: boolean;
  referenceRevision: number;
  showErrors: boolean;
  message: string | null;
  entries: readonly OrdinaryCreateEntry[];
}>;
type Snapshot = Readonly<{
  authority: WorkbookMutationAuthority | null;
  schemas: Readonly<Record<string, OrdinaryCreateSchemaSnapshot>>;
}>;
/** Incident/account-lifetime ordinary authoring. Presentations borrow schema snapshots. */
export class WorkbookOrdinaryCreateOwner {
  private lifetime = 0;
  private generation = 0;
  private retired = false;
  private referenceRevision = 0;
  private authorityReader: WorkbookAuthoringAuthorityReader | null = null;
  private transport: OrdinaryCreateTransport | null = null;
  private recheckAuthority: () => void = () => {};
  private readonly entries = new Map<string, OrdinaryCreateEntry>();
  private readonly activeAttempt = new Map<string, string>();
  private readonly admission = new Set<string>();
  private readonly releases = new Map<string, () => void>();
  private readonly reads = new Map<symbol, AbortController>();
  private readonly dispatches = new Map<string, number>();
  private readonly refreshes = new Set<string>();
  private readonly notified = new Set<string>();
  private readonly messages = new Map<string, string>();
  private readonly shownErrors = new Set<string>();
  private readonly versions = new Map<string, number>();
  private readonly rows = new Map<string, WorkbookQueryRow>();
  configure(
    authorityReader: WorkbookAuthoringAuthorityReader,
    transport: OrdinaryCreateTransport,
    recheckAuthority: () => void = () => {},
  ) {
    this.authorityReader = authorityReader;
    this.transport = transport;
    this.recheckAuthority = recheckAuthority;
  }
  private reader: WorkbookAuthoringReadPort | null = null;
  configureReader(reader: WorkbookAuthoringReadPort) {
    this.reader = reader;
  }
  getReader() {
    return this.reader;
  }
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private readonly contributions = new Map<
    string,
    OrdinaryCreateContribution
  >();
  private readonly drafts = new Map<string, OrdinaryCreateDraft>();
  private readonly attachments = new Map<symbol, string>();
  private readonly listeners = new Set<() => void>();
  private nextDraft = 0;
  private snapshot: Snapshot = { authority: null, schemas: {} };
  constructor(
    readonly incidentId: string,
    contributions: readonly OrdinaryCreateContribution[],
    private readonly options: {
      ids?: SecureTransactionIdPort;
      effects?: {
        accepted(receipt: OrdinaryCreateReceipt, clientTxnId: string): void;
        refresh(receipt: OrdinaryCreateReceipt): Promise<void>;
      };
      observe?: typeof observeAsyncOperation;
    } = {},
  ) {
    for (const contribution of contributions)
      for (const view of contribution.views) {
        if (this.contributions.has(view))
          throw new Error("Duplicate ordinary creation contribution.");
        this.contributions.set(view, contribution);
        this.drafts.set(view, this.emptyDraft());
      }
  }
  private emptyDraft(): OrdinaryCreateDraft {
    return freezeWorkbookValue({
      id: ++this.nextDraft,
      revision: 0,
      values: {},
      references: {},
    });
  }
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  supports(view: string) {
    return this.contributions.has(view);
  }
  contribution(view: string) {
    return this.contributions.get(view) ?? null;
  }
  private publish() {
    const schemas: Record<string, OrdinaryCreateSchemaSnapshot> = {};
    if (this.authority)
      for (const [view, draft] of this.drafts) {
        const contract = requireViewContract(view),
          contribution = this.contributions.get(view);
        if (!contribution) continue;
        const defaults = {
          ...initialGenericCreateDraft(contract, this.authority.actorId),
          ...contribution.defaults?.(contract, this.authority.actorId),
        };
        for (const input of contract.createInputs)
          delete defaults[input.inputKey];
        const prepared = contribution.prepare(
          contract,
          draft.values,
          "ordinary-readiness",
        );
        schemas[view] = {
          draft,
          values: Object.fromEntries(
            Object.entries({ ...defaults, ...draft.values }).map(
              ([key, value]) => [key, value ?? ""],
            ),
          ),
          errors: prepared.errors,
          ready: !!prepared.request && this.canAuthor() && !this.busy(view),
          busy: this.busy(view),
          referenceRevision: this.referenceRevision,
          showErrors: this.shownErrors.has(view),
          message: this.messages.get(view) ?? null,
          entries: [...this.entries.values()].filter(
            (entry) => entry.attempt.target.viewSchemaId === view,
          ),
        };
      }
    this.snapshot = freezeWorkbookValue({ authority: this.authority, schemas });
    for (const listener of this.listeners) listener();
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (this.retired) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId && this.actorId !== authority.actorId))
    ) {
      this.retire();
      return;
    }
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    this.generation++;
    this.referenceRevision++;
    for (const controller of this.reads.values()) controller.abort();
    this.authority =
      authority?.incidentId === this.incidentId
        ? freezeWorkbookValue({ ...authority })
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.publish();
  }
  canAuthor() {
    return (
      !!this.authority?.sessionIdentity &&
      !this.authority.closed &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  retire() {
    this.retired = true;
    this.lifetime++;
    this.generation++;
    for (const controller of this.reads.values()) controller.abort();
    for (const release of this.releases.values()) release();
    this.reads.clear();
    this.releases.clear();
    this.entries.clear();
    this.dispatches.clear();
    this.refreshes.clear();
    this.notified.clear();
    this.activeAttempt.clear();
    this.admission.clear();
    this.messages.clear();
    this.shownErrors.clear();
    this.rows.clear();
    this.versions.clear();
    this.authority = null;
    this.actorId = null;
    this.attachments.clear();
    for (const view of this.drafts.keys())
      this.drafts.set(view, this.emptyDraft());
    this.publish();
  }
  attach(view: string, attachment: symbol) {
    if (this.supports(view)) this.attachments.set(attachment, view);
    return () => {
      this.attachments.delete(attachment);
    };
  }
  update(view: string, key: string, value: string | null | undefined) {
    const draft = this.drafts.get(view);
    if (!draft || !this.canAuthor()) return;
    const values = { ...draft.values };
    if (value === undefined) delete values[key];
    else values[key] = value;
    if (JSON.stringify(values) === JSON.stringify(draft.values)) return;
    this.drafts.set(
      view,
      freezeWorkbookValue({ ...draft, revision: draft.revision + 1, values }),
    );
    this.publish();
  }
  selectReferences(
    view: string,
    key: string,
    selected: readonly WorkbookAuthoringCandidate[],
  ) {
    const draft = this.drafts.get(view);
    if (!draft || !this.canAuthor()) return;
    const multiple =
      requireViewContract(view).fieldMap[key]?.writeKind === "action_payload";
    this.drafts.set(
      view,
      freezeWorkbookValue({
        ...draft,
        revision: draft.revision + 1,
        values: {
          ...draft.values,
          [key]: selected.length
            ? selected.map((item) => item.recordId).join("\n")
            : multiple
              ? ""
              : null,
        },
        references: { ...draft.references, [key]: structuredClone(selected) },
      }),
    );
    this.publish();
  }

  busy(view: string) {
    const active = this.entries.get(this.activeAttempt.get(view) ?? "");
    return (
      this.admission.has(view) ||
      active?.failure?.kind === "client_txn_conflict" ||
      [...this.entries.values()].some(
        (entry) =>
          entry.attempt.target.viewSchemaId === view &&
          (entry.phase === "preparing" ||
            entry.phase === "submitting" ||
            entry.phase === "uncertain"),
      )
    );
  }
  /** Admitted writes awaiting settlement; excludes acknowledged refresh reads. */
  get unsettledMutationCount() {
    return (
      this.admission.size +
      [...this.entries.values()].filter(
        (entry) =>
          !entry.receipt &&
          (entry.phase === "preparing" ||
            entry.phase === "submitting" ||
            entry.phase === "uncertain"),
      ).length
    );
  }
  get pendingCount() {
    return (
      this.admission.size +
      [...this.entries.values()].filter(
        (entry) => entry.transportPending || entry.refresh === "refreshing",
      ).length
    );
  }
  canReplay() {
    return (
      !!this.authority?.sessionIdentity &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  latestVersion(recordId: string) {
    return this.authority ? (this.versions.get(recordId) ?? null) : null;
  }
  latestRow(recordId: string) {
    if (!this.authority) return null;
    const row = this.rows.get(recordId);
    return row && row.row_version >= (this.versions.get(recordId) ?? 0)
      ? row
      : null;
  }
  observe(recordId: string, version: number) {
    if (this.retired || version <= (this.versions.get(recordId) ?? 0)) return;
    this.versions.set(recordId, version);
  }
  acceptRow(row: WorkbookQueryRow): WorkbookQueryRow | null {
    if (!this.authority || this.retired) return null;
    if (row.row_version < (this.versions.get(row.record_id) ?? 0))
      return this.latestRow(row.record_id);
    this.observe(row.record_id, row.row_version);
    const current = this.rows.get(row.record_id);
    if (!current || row.row_version > current.row_version)
      this.rows.set(row.record_id, freezeWorkbookValue(structuredClone(row)));
    return this.rows.get(row.record_id) ?? null;
  }
  private replace(id: string, changes: Partial<OrdinaryCreateEntry>) {
    const entry = this.entries.get(id);
    if (entry)
      this.entries.set(id, freezeWorkbookValue({ ...entry, ...changes }));
  }
  private release(id: string) {
    this.releases.get(id)?.();
    this.releases.delete(id);
  }
  private async readAuthority(
    baseline: WorkbookMutationAuthority,
    signal: AbortSignal,
  ) {
    const generation = this.generation;
    try {
      return await boundedRead(
        (child) => this.authorityReader?.(baseline, child) ?? Promise.reject(),
        signal,
      );
    } catch (error) {
      if (!signal.aborted && !this.retired && generation === this.generation) {
        this.suspend();
        this.recheckAuthority();
      }
      throw error;
    }
  }
  private async currentAuthority(
    replay: boolean,
    baseline: WorkbookMutationAuthority,
    signal: AbortSignal,
  ) {
    const generation = this.generation;
    if (!this.authorityReader)
      throw new Error("Current authority is unavailable.");
    const current = await this.readAuthority(baseline, signal);
    if (generation !== this.generation || this.retired || signal.aborted)
      throw new Error("Authorization changed.");
    if (
      current.incidentId !== this.incidentId ||
      current.actorId !== baseline.actorId
    ) {
      this.retire();
      throw new Error("Account or incident changed.");
    }
    this.setAuthority(current);
    if (!(replay ? this.canReplay() : this.canAuthor()))
      throw new Error("Creating rows is not currently permitted.");
    return current;
  }
  async submit(view: string) {
    const draft = this.drafts.get(view),
      contribution = this.contributions.get(view),
      transport = this.transport,
      authority = this.authority;
    if (
      !draft ||
      !contribution ||
      !transport ||
      !authority ||
      !this.canAuthor() ||
      this.busy(view)
    )
      return;
    // Reserve at the operation owner before any asynchronous authority or capability work.
    this.admission.add(view);
    this.messages.delete(view);
    this.shownErrors.add(view);
    const lifetime = this.lifetime,
      generation = this.generation;
    let id: string | null = null;
    const controller = new AbortController(),
      readKey = Symbol("create admission");
    this.reads.set(readKey, controller);
    try {
      const target = requireViewContract(view);
      if (
        !contribution.prepare(target, draft.values, "ordinary-readiness")
          .request
      )
        return;
      id = this.options.ids?.create("ordinary-create") ?? null;
      if (!id) throw new Error("A secure operation could not be prepared.");
      const request = contribution.prepare(target, draft.values, id).request;
      if (!request) return;
      const attempt = freezeWorkbookValue(
        structuredClone(
          transport.capture({
            authority,
            target,
            draft,
            request,
            clientTxnId: id,
          }),
        ),
      );
      const release = contribution.reserve?.(target);
      if (release === null) {
        this.messages.set(
          view,
          "Finish Entity recovery before creating this row.",
        );
        return;
      }
      if (release) this.releases.set(id, release);
      this.activeAttempt.set(view, id);
      this.entries.set(
        id,
        freezeWorkbookValue({
          attempt,
          phase: "preparing",
          dispatched: false,
          transportPending: false,
          uncertain: false,
          receipt: null,
          status: null,
          failure: null,
          refresh: "none",
          message: "Checking creation…",
        }),
      );
      this.publish();
      await this.currentAuthority(false, authority, controller.signal);
      if (lifetime !== this.lifetime || generation !== this.generation)
        throw new Error("Authorization changed.");
      await boundedRead(
        (signal) => transport.verify(target, signal),
        controller.signal,
      );
      if (
        lifetime !== this.lifetime ||
        generation !== this.generation ||
        !this.canAuthor()
      )
        throw new Error("Authorization changed.");
      this.admission.delete(view);
      await this.dispatch(id);
    } catch {
      if (lifetime === this.lifetime) {
        if (id && this.entries.has(id) && !this.entries.get(id)?.dispatched) {
          this.replace(id, {
            phase: "rejected",
            message: "Creation could not be verified. Your draft is retained.",
          });
          this.release(id);
        } else
          this.messages.set(
            view,
            "Creation could not be prepared. Your draft is retained.",
          );
      }
    } finally {
      this.reads.delete(readKey);
      if (lifetime === this.lifetime) {
        this.admission.delete(view);
        this.publish();
      }
    }
  }
  private applyAcceptance(id: string) {
    const entry = this.entries.get(id);
    if (!entry?.receipt || !this.authority || this.notified.has(id)) return;
    const accepted = acceptedRecordMutation(
      entry.receipt.data,
      entry.attempt.target.viewSchemaId,
    );
    if (!accepted) return;
    this.acceptRow(accepted.row);
    this.contributions
      .get(entry.attempt.target.viewSchemaId)
      ?.accepted?.(accepted.row);
    this.options.effects?.accepted(entry.receipt, id);
    this.notified.add(id);
  }
  private receive(
    id: string,
    outcome: OrdinaryCreateOutcome,
    sequence: number,
  ) {
    const entry = this.entries.get(id);
    if (!entry || entry.receipt || this.retired) return;
    if (outcome.kind === "accepted") {
      const receipt = outcome.receipt;
      if (
        (outcome.status !== 200 && outcome.status !== 201) ||
        this.contributions
          .get(entry.attempt.target.viewSchemaId)
          ?.validateReceipt?.(
            entry.attempt.target,
            entry.attempt.body,
            receipt,
            outcome.status,
          ) === false ||
        typeof receipt.meta?.request_id !== "string" ||
        !receipt.meta.request_id.trim() ||
        !acceptedRecordMutation(receipt.data, entry.attempt.target.viewSchemaId)
      ) {
        this.receive(id, { kind: "uncertain" }, sequence);
        return;
      }
      // The receipt is the monotonic write checkpoint, before resets, projections or reads.
      this.replace(id, {
        phase: "accepted",
        receipt: structuredClone(receipt),
        status: outcome.status,
        failure: null,
        transportPending: false,
        refresh: "required",
        message: "Row accepted. Refreshing workbook…",
      });
      this.release(id);
      this.referenceRevision++;
      const view = entry.attempt.target.viewSchemaId,
        draft = this.drafts.get(view);
      if (
        this.activeAttempt.get(view) === id &&
        draft?.id === entry.attempt.draft.id
      ) {
        this.drafts.set(
          view,
          draft.revision === entry.attempt.draft.revision
            ? this.emptyDraft()
            : freezeWorkbookValue({ ...draft, id: ++this.nextDraft }),
        );
        this.shownErrors.delete(view);
      }
      this.publish();
      try {
        this.applyAcceptance(id);
      } catch {
        /* Receipt remains available for read-only reconciliation. */
      }
      if (this.authority) void this.refreshAccepted(id, false);
      return;
    }
    if (this.dispatches.get(id) !== sequence) return;
    const terminalClosed =
      outcome.kind === "rejected" &&
      outcome.failure.publicCode === "incident_closed";
    if (outcome.kind === "rejected" && (!entry.uncertain || terminalClosed)) {
      this.replace(id, {
        phase: "rejected",
        transportPending: false,
        failure: outcome.failure,
        message: terminalClosed
          ? "The incident is closed. Your draft is retained."
          : outcome.failure.kind === "client_txn_conflict"
            ? "This submission cannot be recovered with its current request. Review your draft before submitting it as a new request."
            : outcome.failure.kind === "validation"
              ? "The row was rejected. Review the values and commit again; your exact draft is retained."
              : "Creation was rejected. Your draft is retained.",
      });
      this.release(id);
    } else
      this.replace(id, {
        phase: "uncertain",
        uncertain: true,
        transportPending: false,
        failure: outcome.kind === "rejected" ? outcome.failure : null,
        message:
          "This submission is unconfirmed. Recover it before committing another row on this sheet.",
      });
    if (terminalClosed) this.closeIncident();
    if (
      outcome.kind === "rejected" &&
      workbookFailureLifecycle(outcome.failure).kind === "authority_unavailable"
    ) {
      this.suspend();
      this.recheckAuthority();
    }
    this.publish();
  }
  private async dispatch(id: string) {
    const entry = this.entries.get(id),
      transport = this.transport,
      lifetime = this.lifetime;
    if (!entry || entry.receipt || !transport || this.retired) return;
    const sequence = (this.dispatches.get(id) ?? 0) + 1;
    this.dispatches.set(id, sequence);
    this.replace(id, {
      phase: "submitting",
      dispatched: true,
      transportPending: true,
      message: "Creating row…",
    });
    this.publish();
    const observed = (this.options.observe ?? observeAsyncOperation)(
      async (signal) => {
        let outcome: OrdinaryCreateOutcome;
        try {
          outcome = await transport.send(entry.attempt, signal);
        } catch {
          outcome = { kind: "uncertain" };
        }
        if (lifetime === this.lifetime) this.receive(id, outcome, sequence);
        return outcome;
      },
    );
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
      !this.canReplay()
    )
      return;
    const view = entry.attempt.target.viewSchemaId;
    if (this.admission.has(view)) return;
    this.admission.add(view);
    this.publish();
    const controller = new AbortController(),
      key = Symbol("replay authority");
    this.reads.set(key, controller);
    try {
      await this.currentAuthority(
        true,
        entry.attempt.authority,
        controller.signal,
      );
      // Committed replay precedes fresh lifecycle/capability checks, even while closed.
      if (lifetime !== this.lifetime || this.entries.get(id)?.receipt) return;
      this.admission.delete(view);
      await this.dispatch(id);
    } catch {
      if (lifetime === this.lifetime)
        this.replace(id, {
          message: "Current access could not be verified. Retry recovery.",
        });
    } finally {
      this.reads.delete(key);
      if (lifetime === this.lifetime) {
        this.admission.delete(view);
        this.publish();
      }
    }
  }
  async submitFreshAfterConflict(id: string) {
    const entry = this.entries.get(id);
    if (
      !entry ||
      entry.receipt ||
      entry.transportPending ||
      entry.failure?.kind !== "client_txn_conflict" ||
      !this.canAuthor()
    )
      return;
    const view = entry.attempt.target.viewSchemaId;
    if (this.admission.has(view)) return;
    this.replace(id, { phase: "rejected", uncertain: false });
    this.release(id);
    this.activeAttempt.delete(view);
    await this.submit(view);
  }
  async refreshAccepted(id: string, explicit = true) {
    const entry = this.entries.get(id),
      authority = this.authority,
      lifetime = this.lifetime;
    if (
      !entry?.receipt ||
      !authority ||
      this.refreshes.has(id) ||
      entry.refresh === "complete"
    )
      return;
    this.refreshes.add(id);
    this.replace(id, { refresh: "refreshing" });
    this.publish();
    const controller = new AbortController(),
      key = Symbol("accepted refresh");
    this.reads.set(key, controller);
    try {
      if (explicit) {
        // Read recovery admits viewers and closed incidents; it never needs write authority.
        const current = await this.readAuthority(authority, controller.signal);
        if (controller.signal.aborted || lifetime !== this.lifetime) return;
        if (
          current.actorId !== authority.actorId ||
          current.incidentId !== this.incidentId
        ) {
          this.retire();
          return;
        }
        this.setAuthority(current);
      }
      if (lifetime !== this.lifetime || !this.authority) return;
      this.applyAcceptance(id);
      await boundedRead(
        () =>
          this.options.effects?.refresh(
            entry.receipt as OrdinaryCreateReceipt,
          ) ?? Promise.resolve(),
        controller.signal,
      );
      if (lifetime !== this.lifetime || !this.authority) return;
      this.replace(id, { refresh: "complete", message: "Row accepted." });
    } catch {
      if (lifetime === this.lifetime)
        this.replace(id, {
          refresh: "required",
          message:
            "Row accepted. The workbook could not be refreshed; previously loaded rows may be stale.",
        });
    } finally {
      this.reads.delete(key);
      this.refreshes.delete(id);
      if (lifetime === this.lifetime) {
        if (this.entries.get(id)?.refresh === "refreshing")
          this.replace(id, { refresh: "required" });
        this.publish();
      }
    }
  }
}
