import {
  boundedRead,
  observeAsyncOperation,
} from "../../../services/asyncObservation";
import type {
  PartyCreationAttempt,
  PartyCreationOutcome,
  PartyCreationReceipt,
  PartyCreationTransport,
} from "../../adapters/createPartyCreationTransport";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type {
  ExplicitPatchAuthority,
  ExplicitPatchOperation,
  WorkbookExplicitPatchOwner,
} from "../../runtime/WorkbookExplicitPatchOwner";
import {
  type PartyAction,
  type PartyLinkReadPort,
  type PartyReview,
  partyChanges,
  partyViewId,
} from "./partyLinkModel";

type PartyCreationOperation = Readonly<{
  id: string;
  attempt: PartyCreationAttempt;
  phase:
    | "preparing"
    | "preparation_failed"
    | "submitting"
    | "uncertain"
    | "rejected"
    | "accepted";
  receipt: PartyCreationReceipt | null;
  failure: WorkbookOperationFailure | null;
  linkId: string | null;
  refresh: "pending" | "refreshing" | "required" | "complete";
}>;
export type PartyAuthorityReader = (
  baseline: ExplicitPatchAuthority,
  signal: AbortSignal,
) => Promise<ExplicitPatchAuthority>;
type PartyPatchOperation = ExplicitPatchOperation & {
  intent: ExplicitPatchOperation["intent"] & {
    owner: "party_link";
    review: PartyReview;
  };
};
function isPartyPatch(
  entry: ExplicitPatchOperation,
): entry is PartyPatchOperation {
  return entry.intent.owner === "party_link";
}
type Snapshot = Readonly<{
  candidateRevision: number;
  preparationFailure: { presentation: string; message: string } | null;
  authority: ExplicitPatchAuthority | null;
  generation: number;
  creations: readonly PartyCreationOperation[];
  patches: readonly PartyPatchOperation[];
}>;

/** Owns the two-commit Party workflow; source PATCH execution stays in explicitPatches. */
export class WorkbookPartyLinkOperationOwner {
  private authority: ExplicitPatchAuthority | null = null;
  private actor: string | null = null;
  private generation = 0;
  private retired = false;
  private presentation: string | null = null;
  private readonly creations = new Map<string, PartyCreationOperation>();
  private readonly transports = new Map<string, PartyCreationTransport>();
  private readonly running = new Set<string>();
  private readonly listeners = new Set<() => void>();
  private transport: PartyCreationTransport | null = null;
  private reader: PartyLinkReadPort | null = null;
  private recheck: PartyAuthorityReader | null = null;
  private preparationFailure: Snapshot["preparationFailure"] = null;
  private candidateRevision = 0;
  private readonly partyVersions = new Map<string, number>();
  private snapshot: Snapshot = {
    candidateRevision: 0,
    preparationFailure: null,
    authority: null,
    generation: 0,
    creations: [],
    patches: [],
  };
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    readonly patches: WorkbookExplicitPatchOwner,
    private readonly effects: {
      coordinate(
        review: PartyReview,
        signal: AbortSignal,
        reservationId: string,
      ): Promise<WorkbookSourceWriteSettlement>;
      remember(id: string): void;
      settle(id: string): void;
      refresh(view: string): Promise<void>;
    },
    private readonly observe = observeAsyncOperation,
  ) {
    patches.subscribe(() => this.publish());
  }
  configure(
    transport: PartyCreationTransport,
    reader: PartyLinkReadPort,
    recheck: PartyAuthorityReader,
  ) {
    this.transport = transport;
    this.reader = reader;
    this.recheck = recheck;
  }
  private async refreshAuthority(
    baseline: ExplicitPatchAuthority,
    signal: AbortSignal,
  ) {
    const read = this.recheck;
    if (!read || this.retired)
      throw new Error("Current authority is unavailable.");
    const generation = this.generation;
    const authority = await boundedRead(
      (currentSignal) => read(baseline, currentSignal),
      signal,
    );
    if (generation !== this.generation || signal.aborted)
      throw new Error("Access changed during review.");
    this.patches.setAuthority(authority);
    this.setAuthority(authority);
  }
  async recheckAuthority() {
    const baseline =
      this.authority ??
      [...this.creations.values()][0]?.attempt.review.authority ??
      this.patches.getSnapshot().entries.find(isPartyPatch)?.intent.review
        .authority;
    if (!baseline) return;
    try {
      await this.refreshAuthority(baseline, new AbortController().signal);
    } catch {
      /* Recovery remains retained. */
    }
  }
  getReader() {
    return this.reader;
  }
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish() {
    const authority =
      this.authority && this.patches.getSnapshot().authority
        ? this.authority
        : null;
    this.snapshot = {
      candidateRevision: this.candidateRevision,
      preparationFailure: authority ? this.preparationFailure : null,
      authority,
      generation: this.generation,
      creations: authority ? [...this.creations.values()] : [],
      patches: authority
        ? this.patches.getSnapshot().entries.filter(isPartyPatch)
        : [],
    };
    for (const listener of this.listeners) listener();
  }
  private update(id: string, update: Partial<PartyCreationOperation>) {
    const old = this.creations.get(id);
    if (old && !this.retired) {
      this.creations.set(id, { ...old, ...update });
      this.publish();
    }
  }
  setAuthority(authority: ExplicitPatchAuthority | null) {
    if (this.retired) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actor && this.actor !== authority.actorId))
    ) {
      this.retire();
      return;
    }
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    this.authority = authority ? structuredClone(authority) : null;
    if (authority) this.actor = authority.actorId;
    this.generation++;
    this.presentation = null;
    this.publish();
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  retire() {
    this.retired = true;
    this.authority = null;
    this.presentation = null;
    this.generation++;
    for (const id of this.creations.keys()) this.effects.settle(id);
    this.creations.clear();
    this.transports.clear();
    this.running.clear();
    this.publish();
  }
  setPresentation(presentation: string | null) {
    this.presentation = presentation;
  }
  canSubmit() {
    return (
      !this.retired &&
      !!this.authority &&
      this.patches.canSubmit() &&
      !this.authority.closed &&
      ["editor", "reviewer", "admin"].includes(this.authority.role)
    );
  }
  isCurrent(review: PartyReview) {
    return (
      this.canSubmit() &&
      review.presentation === this.presentation &&
      JSON.stringify(review.authority) === JSON.stringify(this.authority) &&
      review.source.row_version >=
        (this.patches.latestVersion(review.source.record_id) ?? 0)
    );
  }
  observePartyVersion(id: string, version: number) {
    if (this.retired || version <= (this.partyVersions.get(id) ?? 0)) return;
    this.partyVersions.set(id, version);
    this.candidateRevision++;
    this.publish();
  }
  acceptVersion(id: string, version: number) {
    this.patches.acceptVersion(id, version);
  }
  blocksRecord(id: string, ownReservationId?: string) {
    return [...this.creations.values()].some(
      (entry) =>
        entry.id !== ownReservationId &&
        entry.attempt.review.source.record_id === id &&
        ["preparing", "submitting", "uncertain"].includes(entry.phase),
    );
  }
  /** Save-state facts are separate from operation admission and refresh. */
  get unsettledMutationCount() {
    return [...this.creations.values()].filter(
      (entry) =>
        !entry.receipt &&
        ["preparing", "submitting", "uncertain"].includes(entry.phase),
    ).length;
  }
  private async readSource(
    review: PartyReview,
    signal: AbortSignal,
    minimum = review.source.row_version,
  ): Promise<WorkbookQueryRow> {
    if (!this.reader || !this.authority)
      throw new Error("Current source read is unavailable.");
    const generation = this.generation;
    const row = await this.reader.source(
      review.pair.viewSchemaId,
      review.source.record_id,
      signal,
    );
    if (
      generation !== this.generation ||
      row.record_id !== review.source.record_id ||
      row.row_version <
        Math.max(minimum, this.patches.latestVersion(row.record_id) ?? 0)
    )
      throw new Error("The source needs a current read.");
    this.patches.acceptRow(row);
    return row;
  }
  async create(review: PartyReview, draft: Readonly<Record<string, string>>) {
    if (
      !this.isCurrent(review) ||
      !this.transport ||
      this.blocksRecord(review.source.record_id) ||
      this.patches.blocksRecord(review.source.record_id) ||
      [...this.creations.values()].some(
        (entry) =>
          entry.attempt.review.presentation === review.presentation &&
          entry.phase === "accepted",
      )
    )
      return;
    let attempt: PartyCreationAttempt;
    try {
      attempt = this.transport.capture(
        review,
        draft,
        this.ids.create("party-context-create"),
      );
    } catch {
      this.preparationFailure = {
        presentation: review.presentation,
        message:
          "Party creation was not sent. Check the reviewed fields and try again.",
      };
      this.publish();
      return;
    }
    this.preparationFailure = null;
    review = attempt.review;
    const id = attempt.id;
    this.transports.set(id, this.transport);
    this.creations.set(id, {
      id,
      attempt,
      phase: "preparing",
      receipt: null,
      failure: null,
      linkId: null,
      refresh: "pending",
    });
    this.publish();
    const preparation = this.observe(async (signal) => {
      await this.refreshAuthority(review.authority, signal);
      const settlement = await this.effects.coordinate(review, signal, id);
      if (settlement.kind !== "settled") return false;
      this.patches.acceptVersion(
        review.source.record_id,
        settlement.minimumRowVersion,
      );
      await this.readSource(review, signal);
      return this.isCurrent(review);
    });
    const prepared = await preparation.result;
    if (
      prepared.kind !== "completed" ||
      !prepared.value ||
      !this.isCurrent(review)
    ) {
      this.update(id, {
        phase: "preparation_failed",
        failure: {
          kind: "stale_target",
          message:
            "Source or access changed. Review the current source before saving a Party.",
        },
      });
      return;
    }
    await this.executeCreation(id, true);
  }
  private async executeCreation(id: string, autoLink: boolean) {
    const entry = this.creations.get(id),
      port = this.transports.get(id);
    if (
      !entry ||
      !port ||
      !this.canSubmit() ||
      this.running.has(id) ||
      entry.receipt
    )
      return;
    this.running.add(id);
    this.effects.remember(id);
    this.update(id, { phase: "submitting", failure: null });
    let observed = false;
    const observation = this.observe(async (signal) => {
      const outcome = await port.send(entry.attempt, signal);
      if (observed && outcome.kind === "accepted")
        await this.acceptCreation(id, outcome.receipt, false);
      return outcome;
    });
    const result = await observation.result;
    observed = true;
    this.running.delete(id);
    if (
      this.retired ||
      !this.creations.has(id) ||
      this.creations.get(id)?.receipt
    )
      return;
    const outcome: PartyCreationOutcome =
      result.kind === "completed" ? result.value : { kind: "uncertain" };
    if (outcome.kind === "accepted") {
      await this.acceptCreation(id, outcome.receipt, autoLink);
      return;
    }
    if (outcome.kind === "uncertain") {
      this.update(id, { phase: "uncertain" });
      return;
    }
    const wasUncertain = entry.phase === "uncertain";
    this.update(id, {
      phase:
        wasUncertain || outcome.failure.kind === "client_txn_conflict"
          ? "uncertain"
          : "rejected",
      failure: outcome.failure,
    });
    if (!wasUncertain) this.effects.settle(id);
    if (
      workbookFailureLifecycle(outcome.failure).kind === "authority_unavailable"
    ) {
      this.suspend();
      void this.refreshAuthority(
        entry.attempt.review.authority,
        new AbortController().signal,
      ).catch(() => {});
    }
  }
  private async acceptCreation(
    id: string,
    receipt: PartyCreationReceipt,
    autoLink: boolean,
  ) {
    const entry = this.creations.get(id);
    if (!entry || this.retired || entry.receipt) return;
    this.update(id, {
      phase: "accepted",
      receipt: structuredClone(receipt),
      failure: null,
      refresh: "required",
    });
    this.effects.settle(id);
    if (autoLink && this.isCurrent(entry.attempt.review))
      await this.linkCreated(id, entry.attempt.review);
    await this.refreshCreation(id);
  }
  async replayCreation(id: string) {
    const entry = this.creations.get(id);
    if (
      !entry ||
      entry.phase !== "uncertain" ||
      !this.canSubmit() ||
      this.running.has(id)
    )
      return;
    this.running.add(id);
    const generation = this.generation;
    try {
      await this.refreshAuthority(
        entry.attempt.review.authority,
        new AbortController().signal,
      );
    } catch {
      return;
    } finally {
      this.running.delete(id);
    }
    if (generation !== this.generation || !this.canSubmit()) return;
    await this.executeCreation(id, false);
  }
  async patch(review: PartyReview, action: PartyAction, target = "") {
    if (
      !this.isCurrent(review) ||
      this.blocksRecord(review.source.record_id) ||
      (action === "link" && !target)
    )
      return null;
    return this.patches.submit(
      {
        viewSchemaId: review.pair.viewSchemaId,
        owner: "party_link",
        compound: true,
        baseline: review.source,
        changes: partyChanges(review.pair, action, target),
        purpose: `party-${action}`,
        sheetRef: review.sheetRef,
        surfaceLabel: review.sourceLabel,
        review,
      },
      [
        {
          dependencies: [review.pair.textFieldKey, review.pair.refFieldKey],
          prepare: async (signal) => {
            await this.refreshAuthority(review.authority, signal);
            await this.readSource(review, signal);
            if (target) {
              if (!this.reader) throw new Error("Party lookup is unavailable.");
              await this.reader.source(partyViewId, target, signal);
            }
            if (!this.isCurrent(review))
              throw new Error(
                "The source or access changed. Review this Party action again.",
              );
          },
          accessRejected: () =>
            this.refreshAuthority(
              review.authority,
              new AbortController().signal,
            ),
          recheck: () =>
            this.refreshAuthority(
              review.authority,
              new AbortController().signal,
            ),
          reconcile: async (entry) => {
            if (!entry.receipt) throw new Error("Missing Party receipt");
            await this.readSource(
              review,
              new AbortController().signal,
              entry.receipt.row.row_version,
            );
            await this.effects.refresh(review.pair.viewSchemaId);
          },
        },
      ],
    );
  }
  async linkCreated(id: string, review: PartyReview) {
    const entry = this.creations.get(id);
    if (
      !entry?.receipt ||
      entry.attempt.review.source.record_id !== review.source.record_id ||
      entry.attempt.review.pair.key !== review.pair.key ||
      !this.isCurrent(review)
    )
      return;
    const previous = entry.linkId
      ? this.patches
          .getSnapshot()
          .entries.find((patch) => patch.id === entry.linkId)
      : null;
    if (
      previous &&
      previous.phase !== "rejected" &&
      previous.phase !== "preparation_failed"
    )
      return;
    const result = await this.patch(
      review,
      "link",
      entry.receipt.data.row.record_id,
    );
    if (result) this.update(id, { linkId: result.id });
  }
  async refreshCreation(id: string) {
    const entry = this.creations.get(id);
    if (
      !entry?.receipt ||
      !this.authority ||
      !this.reader ||
      entry.refresh === "refreshing"
    )
      return;
    const generation = this.generation;
    this.update(id, { refresh: "refreshing" });
    try {
      const row = await this.reader.source(
        partyViewId,
        entry.receipt.data.row.record_id,
        new AbortController().signal,
      );
      if (
        generation !== this.generation ||
        row.row_version < entry.receipt.data.row.row_version
      )
        throw new Error("Party read is stale.");
      await this.effects.refresh(partyViewId);
      if (generation !== this.generation) throw new Error("Access changed.");
      this.update(id, { refresh: "complete" });
    } catch {
      this.update(id, { refresh: "required" });
    }
  }
}
