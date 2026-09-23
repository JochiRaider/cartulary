import { observeAsyncOperation } from "../../../services/asyncObservation";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type { WorkbookTimelineActionRuntimePort } from "../../ports/WorkbookTimelineActionRuntimePort";
import type { WorkbookRow } from "../models/timelineRowModel";
import {
  type AutoResolutionNotice,
  buildAutoResolutionNotices,
} from "../models/workbookMentionChips";
import type {
  TimelineMentionEntityCreationPort,
  TimelineMentionResolutionPort,
} from "../ports/TimelineMentionPort";
import {
  type MentionCreateReview,
  type MentionCreationAttempt,
  type MentionCreationOperation,
  mentionCreateRequest,
} from "./timelineMentionCreationModel";
import {
  freezeMention,
  type MentionAttempt,
  type MentionAuthority,
  type MentionBinding,
  type MentionOperation,
  type MentionReceipt,
  type MentionReview,
  type MentionSubject,
  mentionReviewValid,
  sameMentionIntent,
} from "./timelineMentionOperationModel";

export type MentionReconciliationScope = Readonly<{
  signal: AbortSignal;
  isCurrent: () => boolean;
}>;
export type AutoResolutionOperation = Readonly<{
  kind: "entry" | "batch";
  operationId: string;
  changeSetId: string | null;
}>;
export type AutoResolutionDisclosure = AutoResolutionNotice &
  Readonly<{
    identity: string;
    sourceRowVersion: number;
    operation: AutoResolutionOperation;
    acceptedCount: number;
  }>;
const sameItems = <T>(previous: readonly T[], next: readonly T[]) =>
  previous.length === next.length &&
  previous.every((item, i) => item === next[i])
    ? previous
    : next;
type Snapshot = Readonly<{
  authority: MentionAuthority | null;
  generation: number;
  entries: readonly MentionOperation[];
  mentions: readonly MentionSubject[];
  creations: readonly MentionCreationOperation[];
}>;

/** Concrete mention attempts belong to the incident/account runtime, never the inspector. */
export class WorkbookTimelineMentionOperationOwner
  implements WorkbookTimelineActionRuntimePort
{
  private authority: MentionAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private sequence = 0;
  private readonly entries = new Map<number, MentionOperation>();
  private readonly creations = new Map<number, MentionCreationOperation>();
  private creationPort: TimelineMentionEntityCreationPort | null = null;
  private readonly versions = new Map<string, number>();
  private readonly disclosures = new Map<string, AutoResolutionDisclosure>();
  private readonly disclosedOperations = new Set<string>();
  private disclosureSnapshot: readonly AutoResolutionDisclosure[] = [];
  private actionSnapshot: Omit<Snapshot, "mentions"> = {
    authority: null,
    generation: 0,
    entries: [],
    creations: [],
  };
  private readonly removedMentions = new Map<string, number>();
  private readonly mentions = new Map<string, MentionSubject>();
  private readonly listeners = new Set<() => void>();
  private readonly preparations = new Map<number, { cancel: () => void }>();
  private readonly sending = new Set<number>();
  private readonly transports = new Map<number, number>();
  private readonly remembered = new Set<string>();
  private sourceReader:
    | ((recordId: string, signal: AbortSignal) => Promise<WorkbookRow>)
    | null = null;
  private port: TimelineMentionResolutionPort | null = null;
  private creationReconcile:
    | ((scope: MentionReconciliationScope) => Promise<void>)
    | null = null;
  private recheckAuthority: (() => void) | undefined;
  private presentationRefresh:
    | ((recordId: string, version: number) => Promise<void>)
    | null = null;
  private reconcile:
    | ((
        receipt: MentionReceipt,
        scope: MentionReconciliationScope,
      ) => Promise<void>)
    | null = null;
  private snapshot: Snapshot = {
    authority: null,
    generation: 0,
    entries: [],
    mentions: [],
    creations: [],
  };
  constructor(
    readonly incidentId: string,
    private readonly ids: SecureTransactionIdPort,
    private readonly accounting: {
      remember: (id: string) => void;
      settle: (id: string) => void;
      accepted: (recordId: string, version: number) => void;
      entityAccepted?: (recordId: string, version: number) => void;
    },
    private readonly observe: typeof observeAsyncOperation = observeAsyncOperation,
  ) {}
  getSnapshot = () => this.snapshot;
  getMentionsSnapshot = () => this.snapshot.mentions;
  getActionSnapshot = () => this.actionSnapshot;
  getDisclosureSnapshot = () => this.disclosureSnapshot;

  canUndoDisclosure(notice: AutoResolutionNotice) {
    const subject = this.latestMention(notice.entityMentionId ?? "");
    return (
      this.canSubmit("revert_to_unresolved") &&
      !!subject &&
      !this.blocksMention(subject.mentionId) &&
      subject.sourceRecordId === notice.rowRecordId &&
      subject.sourceFieldKey === notice.fieldKey &&
      subject.itemRef === notice.itemRef &&
      subject.mentionRowVersion === notice.mentionRowVersion &&
      subject.resolvedRecordId === notice.resolvedRecordId &&
      subject.state === "resolved" &&
      subject.resolutionMethod === "auto_match"
    );
  }

  updateDisclosureLabels(index: Readonly<Record<string, { label: string }>>) {
    let changed = false;
    for (const [key, notice] of this.disclosures) {
      const label = index[notice.resolvedRecordId]?.label;
      if (label && label !== notice.displayText) {
        this.disclosures.set(
          key,
          freezeMention({ ...notice, displayText: label }),
        );
        changed = true;
      }
    }
    if (changed) this.publish();
  }

  /** Accepted operation identity, not the lifetime of a mounted sheet, admits disclosure. */
  acceptAutoResolutions(
    before: readonly WorkbookRow[],
    after: readonly WorkbookRow[],
    operation: AutoResolutionOperation,
  ) {
    const operationKey = JSON.stringify([
      operation.kind,
      operation.operationId,
      operation.changeSetId,
    ]);
    if (!this.disclosedOperations.has(operationKey)) {
      this.disclosedOperations.add(operationKey);
      const notices = after.flatMap((row) =>
        buildAutoResolutionNotices(
          before.find(
            (previous) =>
              previous.recordId === row.recordId || previous.key === row.key,
          ) ?? (operation.kind === "entry" ? before[0] : undefined),
          row,
        ),
      );
      for (const notice of notices) {
        if (!notice.entityMentionId || !notice.mentionRowVersion) continue;
        const row = after.find((row) => row.recordId === notice.rowRecordId);
        if (!row?.rowVersion) continue;
        const identity = JSON.stringify([
          this.incidentId,
          notice.rowRecordId,
          notice.fieldKey,
          notice.entityMentionId,
        ]);
        if (
          (this.removedMentions.get(notice.entityMentionId) ?? 0) >=
          row.rowVersion
        )
          continue;
        const known = this.mentions.get(notice.entityMentionId);
        if (
          known &&
          known.mentionRowVersion >= notice.mentionRowVersion &&
          (known.state !== "resolved" ||
            known.resolutionMethod !== "auto_match")
        )
          continue;
        if (!this.disclosures.has(identity))
          this.disclosures.set(
            identity,
            freezeMention({
              ...notice,
              identity,
              ...(known &&
              known.mentionRowVersion >= notice.mentionRowVersion &&
              known.resolvedRecordId
                ? {
                    rawText: known.rawText,
                    itemRef: known.itemRef,
                    resolvedRecordId: known.resolvedRecordId,
                    mentionRowVersion: known.mentionRowVersion,
                    displayText: known.displayText ?? notice.displayText,
                    matchedAliasText:
                      known.matchedAliasText !== undefined
                        ? known.matchedAliasText
                        : notice.matchedAliasText,
                  }
                : {}),
              sourceRowVersion: Math.max(
                row.rowVersion,
                known?.sourceRowVersion ?? 0,
              ),
              operation,
              acceptedCount: notices.length,
            }),
          );
      }
    }
    for (const row of after) this.observeSource(row);
    this.publish();
  }

  /** Only complete, current source rows can establish removal; query absence cannot. */
  observeSource(row: WorkbookRow) {
    if (
      !row.recordId ||
      !row.rowVersion ||
      row.rowVersion < (this.latestVersion(row.recordId) ?? 0)
    )
      return;
    const items = [
      ...row.collectionValues.hostRefs.map((item) => ({
        item,
        field: "timeline.host_refs" as const,
      })),
      ...row.collectionValues.identityRefs.map((item) => ({
        item,
        field: "timeline.identity_refs" as const,
      })),
    ];
    for (const [id, subject] of this.mentions) {
      if (
        subject.sourceRecordId === row.recordId &&
        subject.sourceRowVersion <= row.rowVersion &&
        !items.some(({ item }) => item.entityMentionId === id) &&
        subject.state !== "dismissed"
      ) {
        this.removedMentions.set(id, row.rowVersion);
        this.mentions.delete(id);
      }
    }
    for (const { item, field } of items) {
      if (!item.entityMentionId || !item.mentionRowVersion) continue;
      this.observeMention({
        incidentId: this.incidentId,
        sourceRecordId: row.recordId,
        sourceRowVersion: row.rowVersion,
        sourceFieldKey: field,
        mentionId: item.entityMentionId,
        itemRef: item.itemRef,
        entityType: item.entityType,
        rawText: item.rawText,
        mentionRowVersion: item.mentionRowVersion,
        state: item.itemKind === "resolved_ref" ? "resolved" : "unresolved",
        resolvedRecordId: item.resolvedRecordId,
        resolutionMethod: item.resolutionMethod,
        displayText: item.displayText,
        matchedAliasText: item.matchedAliasText,
      });
    }
    for (const [key, notice] of this.disclosures) {
      if (
        notice.rowRecordId !== row.recordId ||
        row.rowVersion < notice.sourceRowVersion
      )
        continue;
      const match = items.find(
        ({ item, field }) =>
          field === notice.fieldKey &&
          item.entityMentionId === notice.entityMentionId,
      );
      if (!match) {
        this.disclosures.delete(key);
        continue;
      }
      const { item } = match;
      if ((item.mentionRowVersion ?? 0) < (notice.mentionRowVersion ?? 0))
        continue;
      if (
        item.itemKind !== "resolved_ref" ||
        !item.autoResolved ||
        !item.resolvedRecordId
      ) {
        this.disclosures.delete(key);
        continue;
      }
      const next = {
        ...notice,
        sourceRowVersion: row.rowVersion,
        itemRef: item.itemRef,
        mentionRowVersion: item.mentionRowVersion,
        resolvedRecordId: item.resolvedRecordId,
        rawText: item.rawText,
        displayText:
          item.resolvedRecordId === notice.resolvedRecordId
            ? notice.displayText || item.displayText
            : item.displayText,
        matchedAliasText: item.matchedAliasText,
      };
      if (JSON.stringify(next) !== JSON.stringify(notice))
        this.disclosures.set(key, freezeMention(next));
    }
    this.acceptVersion(row.recordId, row.rowVersion);
    this.publish();
  }

  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  configure(
    port: TimelineMentionResolutionPort,
    recheckAuthority?: () => void,
  ) {
    this.port = port;
    this.recheckAuthority = recheckAuthority;
  }
  configureSourceReader(reader: NonNullable<typeof this.sourceReader>) {
    this.sourceReader = reader;
  }
  async readSource(recordId: string, signal: AbortSignal) {
    const generation = this.generation;
    if (!this.authority || !this.sourceReader)
      throw new Error("Source unavailable");
    const row = await this.sourceReader(recordId, signal);
    if (
      signal.aborted ||
      generation !== this.generation ||
      row.recordId !== recordId ||
      (row.rowVersion ?? 0) < (this.latestVersion(recordId) ?? 0)
    )
      throw new Error("Source changed or unavailable");
    this.observeSource(row);
    return row;
  }
  configureCreation(port: TimelineMentionEntityCreationPort) {
    this.creationPort = port;
  }
  registerReconciliation(reconcile: NonNullable<typeof this.reconcile>) {
    this.reconcile = reconcile;
    return () => {
      if (this.reconcile === reconcile) this.reconcile = null;
    };
  }
  registerCreationReconciliation(
    reconcile: NonNullable<typeof this.creationReconcile>,
  ) {
    this.creationReconcile = reconcile;
    return () => {
      if (this.creationReconcile === reconcile) this.creationReconcile = null;
    };
  }
  registerPresentationRefresh(
    refresh: NonNullable<typeof this.presentationRefresh>,
  ) {
    this.presentationRefresh = refresh;
    return () => {
      if (this.presentationRefresh === refresh) this.presentationRefresh = null;
    };
  }
  async refreshPresentation(recordId: string, version: number) {
    await this.presentationRefresh?.(recordId, version);
  }
  setAuthority(authority: MentionAuthority | null) {
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && authority.actorId !== this.actorId))
    )
      this.retire();
    this.generation++;
    for (const preparation of this.preparations.values()) preparation.cancel();
    this.authority =
      authority?.incidentId === this.incidentId
        ? freezeMention(structuredClone(authority))
        : null;
    if (this.authority) this.actorId = this.authority.actorId;
    this.publish();
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority)
      this.setAuthority({ ...this.authority, mutationsAvailable: false });
  }
  retire() {
    this.generation++;
    for (const preparation of this.preparations.values()) preparation.cancel();
    for (const id of this.remembered) this.settle(id);
    this.preparations.clear();
    this.entries.clear();
    this.creations.clear();
    this.sending.clear();
    this.transports.clear();
    this.versions.clear();
    this.mentions.clear();
    this.removedMentions.clear();
    this.disclosures.clear();
    this.disclosedOperations.clear();
    this.authority = null;
    this.actorId = null;
    this.reconcile = null;
    this.presentationRefresh = null;
    this.creationReconcile = null;
    this.publish();
  }
  canSubmit(action: MentionReview["intent"]["action"]) {
    const authority = this.authority;
    return (
      !!authority?.actorId &&
      !!authority.sessionIdentity &&
      authority.mutationsAvailable &&
      ["editor", "reviewer", "admin"].includes(authority.role ?? "") &&
      authority.actions.includes(action)
    );
  }
  acceptVersion(recordId: string, version: number) {
    if (
      Number.isSafeInteger(version) &&
      version > (this.versions.get(recordId) ?? 0)
    ) {
      this.versions.set(recordId, version);
      this.publish();
    }
  }
  latestVersion = (recordId: string) => this.versions.get(recordId) ?? null;
  observeMention(subject: MentionSubject) {
    if (
      subject.incidentId !== this.incidentId ||
      subject.sourceRowVersion <
        (this.removedMentions.get(subject.mentionId) ?? 0)
    )
      return;
    this.removedMentions.delete(subject.mentionId);
    this.acceptVersion(subject.sourceRecordId, subject.sourceRowVersion);
    const previous = this.mentions.get(subject.mentionId);
    for (const [key, notice] of this.disclosures) {
      if (
        notice.entityMentionId === subject.mentionId &&
        notice.rowRecordId === subject.sourceRecordId &&
        notice.fieldKey === subject.sourceFieldKey &&
        subject.mentionRowVersion >= (notice.mentionRowVersion ?? 0) &&
        (subject.state !== "resolved" ||
          subject.resolutionMethod !== "auto_match")
      )
        this.disclosures.delete(key);
    }
    const next = { ...previous, ...subject };
    if (
      !previous ||
      subject.mentionRowVersion > previous.mentionRowVersion ||
      (subject.mentionRowVersion === previous.mentionRowVersion &&
        (next.displayText !== previous.displayText ||
          next.matchedAliasText !== previous.matchedAliasText))
    ) {
      this.mentions.set(
        subject.mentionId,
        freezeMention(structuredClone(next)),
      );
      this.publish();
    }
  }
  latestMention = (id: string) => this.mentions.get(id) ?? null;
  private blocks(entry: MentionOperation) {
    return (
      ["preparing", "submitting", "uncertain"].includes(entry.phase) ||
      (entry.transportPending && !entry.receipt) ||
      (!!entry.receipt && entry.refresh !== "complete")
    );
  }
  blocksMention(mentionId: string, exceptCreation?: number) {
    return (
      [...this.entries.values()].some(
        (entry) =>
          entry.attempt.review.subject.mentionId === mentionId &&
          this.blocks(entry),
      ) ||
      [...this.creations.values()].some(
        (entry) =>
          entry.key !== exceptCreation &&
          entry.attempt.review.subject.mentionId === mentionId &&
          this.creationBlocks(entry),
      )
    );
  }
  private creationBlocks(entry: MentionCreationOperation) {
    return (
      ["preparing", "submitting", "uncertain"].includes(entry.phase) ||
      (entry.transportPending && !entry.receipt)
    );
  }
  blocksRecord(recordId: string) {
    return (
      [...this.entries.values()].some(
        (entry) =>
          entry.attempt.review.subject.sourceRecordId === recordId &&
          this.blocks(entry),
      ) ||
      [...this.creations.values()].some(
        (entry) =>
          entry.attempt.review.subject.sourceRecordId === recordId &&
          this.creationBlocks(entry),
      )
    );
  }
  /** Save-state facts are separate from operation admission and refresh. */
  get unsettledMutationCount() {
    return [...this.entries.values(), ...this.creations.values()].filter(
      (entry) =>
        !entry.receipt &&
        ["preparing", "submitting", "uncertain"].includes(entry.phase),
    ).length;
  }
  private current(review: MentionReview, binding: MentionBinding) {
    const latest = this.latestMention(review.subject.mentionId);
    return (
      this.canSubmit(review.intent.action) &&
      mentionReviewValid(review) &&
      JSON.stringify(review.authority) === JSON.stringify(this.authority) &&
      binding.isCurrent() &&
      (!latest || latest.mentionRowVersion <= review.subject.mentionRowVersion)
    );
  }
  /** Reserve and freeze route/body/key synchronously, before awaiting preceding saves. */
  submit(review: MentionReview, binding: MentionBinding): boolean {
    return this.reserveResolution(review, binding) !== null;
  }
  private reserveResolution(
    review: MentionReview,
    binding: MentionBinding,
    creationKey?: number,
  ): number | null {
    if (
      !this.port ||
      !this.current(review, binding) ||
      this.blocksMention(review.subject.mentionId, creationKey)
    )
      return null;
    let attempt: MentionAttempt;
    try {
      attempt = freezeMention(
        this.port.capture(review, this.ids.create("timeline-mention")),
      );
      if (
        [...this.entries.values()].some(
          (entry) => entry.attempt.id === attempt.id,
        )
      )
        return null;
    } catch {
      return null;
    }
    const key = ++this.sequence;
    this.entries.set(
      key,
      freezeMention({
        key,
        attempt,
        phase: "preparing",
        receipt: null,
        failure: null,
        refresh: "pending",
        transportPending: false,
      }),
    );
    this.publish();
    void this.prepare(key, binding);
    return key;
  }
  private async prepare(key: number, binding: MentionBinding) {
    const entry = this.entries.get(key);
    if (!entry) return;
    const observation = this.observe(async (signal) => {
      const subject = await binding.prepare(signal);
      return (
        !signal.aborted &&
        subject !== null &&
        sameMentionIntent(entry.attempt.review.subject, subject) &&
        this.current(entry.attempt.review, binding)
      );
    });
    this.preparations.set(key, observation);
    const result = await observation.result;
    this.preparations.delete(key);
    if (this.entries.get(key) !== entry) return;
    if (
      result.kind !== "completed" ||
      !result.value ||
      !this.current(entry.attempt.review, binding)
    ) {
      this.update(key, {
        phase: "preparation_failed",
        failure: {
          kind: "stale_target",
          message:
            "The mention or reviewed intent changed, or earlier edits could not finish. Review it again. This action was not submitted.",
        },
      });
      return;
    }
    await this.send(key, false);
  }
  async replay(key: number) {
    if (this.entries.get(key)?.phase === "uncertain" && !this.sending.has(key))
      await this.send(key, true);
  }
  private async send(key: number, replay: boolean) {
    const entry = this.entries.get(key),
      port = this.port;
    if (
      !entry ||
      !port ||
      !this.canSubmit(entry.attempt.review.intent.action) ||
      entry.attempt.review.authority.actorId !== this.authority?.actorId ||
      this.sending.has(key)
    )
      return;
    const attempt = entry.attempt,
      generation = this.generation;
    const owns = () =>
      this.entries.get(key)?.attempt === attempt &&
      this.actorId === attempt.review.authority.actorId;
    this.sending.add(key);
    this.transports.set(key, (this.transports.get(key) ?? 0) + 1);
    this.remembered.add(attempt.id);
    this.accounting.remember(attempt.id);
    this.update(key, {
      phase: "submitting",
      transportPending: true,
      failure: null,
    });
    let uncertain = replay;
    const observation = this.observe(async (signal) => {
      const outcome = await port.send(attempt, signal);
      if (!owns() || this.entries.get(key)?.receipt) return;
      if (outcome.kind === "accepted") {
        const receipt = freezeMention(structuredClone(outcome.receipt));
        // Acceptance belongs to the attempt even after navigation, role changes or HTTP/socket reordering.
        this.update(key, {
          phase: "accepted",
          receipt,
          failure: null,
          refresh: "pending",
        });
        const mention = receipt.entity_mention;
        this.observeMention({
          ...attempt.review.subject,
          sourceRowVersion: receipt.source_record.row_version,
          mentionRowVersion: mention.row_version,
          state: mention.resolution_status,
          resolvedRecordId: mention.resolved_record_id,
          resolutionMethod: mention.resolution_method,
        });
        this.accounting.accepted(
          receipt.source_record.record_id,
          receipt.source_record.row_version,
        );
        this.settle(attempt.id);
      } else if (outcome.kind === "rejected") {
        this.update(key, {
          phase: uncertain ? "uncertain" : "rejected",
          failure: outcome.failure,
        });
        if (generation === this.generation) {
          if (outcome.failure.publicCode === "incident_closed")
            this.closeIncident();
          // A mention or target 404 and a role rejection are local until current incident/session reads prove access loss.
          if (
            workbookFailureLifecycle(outcome.failure).kind ===
            "authority_unavailable"
          )
            this.recheckAuthority?.();
        }
      } else this.update(key, { phase: "uncertain" });
    });
    const settlement = observation.settled.then(async () => {
      if (!owns()) return;
      const remaining = Math.max(0, (this.transports.get(key) ?? 1) - 1);
      this.transports.set(key, remaining);
      this.update(key, { transportPending: remaining > 0 });
      if (!remaining) this.settle(attempt.id);
      if (this.entries.get(key)?.receipt) await this.refresh(key);
    });
    const result = await observation.result;
    if (!owns()) return;
    this.sending.delete(key);
    if (result.kind !== "completed") {
      uncertain = true;
      this.settle(attempt.id);
      if (this.entries.get(key)?.phase === "submitting")
        this.update(key, { phase: "uncertain" });
    } else await settlement;
  }
  /** Read-only reconciliation is deliberately independent of replay. */
  async refresh(key: number) {
    const entry = this.entries.get(key);
    if (
      !entry?.receipt ||
      !this.authority ||
      entry.refresh === "refreshing" ||
      entry.refresh === "complete"
    )
      return;
    const receipt = entry.receipt,
      generation = this.generation;
    this.update(key, { refresh: "refreshing" });
    let refreshMessage =
      "The action completed. Refresh the source to reconcile its current state.";
    const observation = this.observe(async (signal) => {
      const isCurrent = () =>
        !signal.aborted &&
        this.generation === generation &&
        this.entries.get(key)?.receipt === receipt;
      const reconcile = this.reconcile;
      if (!reconcile || !isCurrent())
        throw new Error("Mention refresh unavailable");
      try {
        await reconcile(receipt, { signal, isCurrent });
      } catch (error) {
        if (error instanceof Error) refreshMessage = error.message;
        throw error;
      }
      if (!isCurrent() || reconcile !== this.reconcile)
        throw new Error("Mention refresh detached");
    });
    const result = await observation.result;
    if (this.entries.get(key)?.receipt === receipt)
      this.update(key, {
        failure:
          result.kind === "completed"
            ? null
            : { kind: "retryable", message: refreshMessage },
        refresh:
          result.kind === "completed" && generation === this.generation
            ? "complete"
            : "required",
      });
  }
  creationForMention(id: string) {
    return (
      [...this.creations.values()].reverse().find((entry) => {
        if (entry.attempt.review.subject.mentionId !== id) return false;
        const linked =
          entry.linkKey === null ? null : this.entries.get(entry.linkKey);
        return (
          !linked?.receipt ||
          linked.refresh !== "complete" ||
          (this.latestMention(id)?.mentionRowVersion ?? 0) <=
            linked.receipt.entity_mention.row_version
        );
      }) ?? null
    );
  }
  canCreate(type: MentionSubject["entityType"]) {
    return (
      this.canSubmit("resolve_item") &&
      (this.authority?.createTypes.includes(type) ?? false)
    );
  }
  private currentCreate(review: MentionCreateReview, binding: MentionBinding) {
    const latest = this.latestMention(review.subject.mentionId);
    return (
      this.canCreate(review.subject.entityType) &&
      review.subject.state === "unresolved" &&
      JSON.stringify(review.authority) === JSON.stringify(this.authority) &&
      binding.isCurrent() &&
      (!latest || latest.mentionRowVersion <= review.subject.mentionRowVersion)
    );
  }
  createAndResolve(
    review: MentionCreateReview,
    binding: MentionBinding,
  ): boolean {
    if (
      !this.creationPort ||
      !this.currentCreate(review, binding) ||
      this.blocksMention(review.subject.mentionId) ||
      (this.creationForMention(review.subject.mentionId)?.receipt ?? false)
    )
      return false;
    let attempt: MentionCreationAttempt;
    try {
      const id = this.ids.create("timeline-mention-create");
      if (!mentionCreateRequest(review, id)) return false;
      attempt = freezeMention(this.creationPort.capture(review, id));
      if (
        [...this.creations.values()].some((entry) => entry.attempt.id === id) ||
        [...this.entries.values()].some((entry) => entry.attempt.id === id)
      )
        return false;
    } catch {
      return false;
    }
    const key = ++this.sequence;
    this.creations.set(
      key,
      freezeMention({
        key,
        attempt,
        phase: "preparing",
        receipt: null,
        failure: null,
        transportPending: false,
        linkKey: null,
        refresh: "pending",
      }),
    );
    this.publish();
    void this.prepareCreation(key, binding);
    return true;
  }
  private async prepareCreation(key: number, binding: MentionBinding) {
    const entry = this.creations.get(key);
    if (!entry) return;
    const observation = this.observe(async (signal) => {
      const subject = await binding.prepare(signal);
      return (
        !signal.aborted &&
        subject !== null &&
        sameMentionIntent(entry.attempt.review.subject, subject) &&
        this.currentCreate(entry.attempt.review, binding)
      );
    });
    this.preparations.set(key, observation);
    const result = await observation.result;
    this.preparations.delete(key);
    if (this.creations.get(key) !== entry) return;
    if (
      result.kind !== "completed" ||
      !result.value ||
      !this.currentCreate(entry.attempt.review, binding)
    ) {
      this.updateCreation(key, {
        phase: "preparation_failed",
        failure: {
          kind: "stale_target",
          message:
            "The mention or create review changed. Review again. Creation was not submitted.",
        },
      });
      return;
    }
    await this.sendCreation(key, false, binding);
  }
  async replayCreation(key: number) {
    if (
      this.creations.get(key)?.phase === "uncertain" &&
      !this.sending.has(key)
    )
      await this.sendCreation(key, true);
  }
  private async sendCreation(
    key: number,
    replay: boolean,
    binding?: MentionBinding,
  ) {
    const entry = this.creations.get(key),
      port = this.creationPort;
    if (
      !entry ||
      !port ||
      !this.canCreate(entry.attempt.review.subject.entityType) ||
      this.sending.has(key) ||
      entry.attempt.review.authority.actorId !== this.authority?.actorId
    )
      return;
    const attempt = entry.attempt,
      generation = this.generation;
    const owns = () =>
      this.creations.get(key)?.attempt === attempt &&
      this.actorId === attempt.review.authority.actorId;
    this.sending.add(key);
    this.transports.set(key, (this.transports.get(key) ?? 0) + 1);
    this.remembered.add(attempt.id);
    this.accounting.remember(attempt.id);
    this.updateCreation(key, {
      phase: "submitting",
      failure: null,
      transportPending: true,
    });
    let uncertain = replay;
    const observation = this.observe(async (signal) => {
      const outcome = await port.send(attempt, signal);
      if (!owns() || this.creations.get(key)?.receipt) return;
      if (outcome.kind === "accepted") {
        // Ordinary creation is committed independently. Retain entity and receipt before any linking decision.
        this.updateCreation(key, {
          phase: "accepted",
          receipt: freezeMention(structuredClone(outcome.receipt)),
          failure: null,
        });
        this.accounting.entityAccepted?.(
          outcome.receipt.data.row.record_id,
          outcome.receipt.data.row.row_version,
        );
        this.settle(attempt.id);
      } else if (outcome.kind === "rejected") {
        this.updateCreation(key, {
          phase: uncertain ? "uncertain" : "rejected",
          failure: outcome.failure,
        });
        if (generation === this.generation) {
          if (outcome.failure.publicCode === "incident_closed")
            this.closeIncident();
          if (
            workbookFailureLifecycle(outcome.failure).kind ===
            "authority_unavailable"
          )
            this.recheckAuthority?.();
        }
      } else this.updateCreation(key, { phase: "uncertain" });
    });
    const settlement = observation.settled.then(async () => {
      if (!owns()) return;
      const remaining = Math.max(0, (this.transports.get(key) ?? 1) - 1);
      this.transports.set(key, remaining);
      this.updateCreation(key, { transportPending: remaining > 0 });
      if (!remaining) this.settle(attempt.id);
      const accepted = this.creations.get(key);
      if (accepted?.receipt) void this.refreshCreation(key);
      if (
        binding &&
        accepted?.receipt &&
        accepted.linkKey === null &&
        this.currentCreate(attempt.review, binding)
      )
        this.linkCreated(key, attempt.review.subject, binding);
    });
    const result = await observation.result;
    if (!owns()) return;
    this.sending.delete(key);
    if (result.kind !== "completed") {
      uncertain = true;
      this.settle(attempt.id);
      if (this.creations.get(key)?.phase === "submitting")
        this.updateCreation(key, { phase: "uncertain" });
    } else await settlement;
  }
  /** Entity-sheet refresh is independent of the separately committed mention action. */
  async refreshCreation(key: number) {
    const entry = this.creations.get(key);
    if (
      !entry?.receipt ||
      !this.authority ||
      entry.refresh === "refreshing" ||
      entry.refresh === "complete"
    )
      return;
    const receipt = entry.receipt,
      generation = this.generation;
    this.updateCreation(key, { refresh: "refreshing" });
    const observation = this.observe(async (signal) => {
      const isCurrent = () =>
        !signal.aborted &&
        this.generation === generation &&
        this.creations.get(key)?.receipt === receipt;
      const reconcile = this.creationReconcile;
      if (!reconcile || !isCurrent())
        throw new Error("Entity refresh unavailable");
      await reconcile({ signal, isCurrent });
      if (!isCurrent() || reconcile !== this.creationReconcile)
        throw new Error("Entity refresh detached");
    });
    const result = await observation.result;
    if (this.creations.get(key)?.receipt === receipt)
      this.updateCreation(key, {
        refresh:
          result.kind === "completed" && generation === this.generation
            ? "complete"
            : "required",
      });
  }
  /** An explicit renewed review may use the current mention version, never a new creation. */
  linkCreated(
    key: number,
    subject: MentionSubject,
    binding: MentionBinding,
  ): boolean {
    const entry = this.creations.get(key),
      authority = this.authority;
    if (!entry?.receipt || !authority) return false;
    const original = entry.attempt.review.subject;
    if (
      subject.mentionId !== original.mentionId ||
      subject.sourceRecordId !== original.sourceRecordId ||
      subject.entityType !== original.entityType ||
      subject.incidentId !== original.incidentId
    )
      return false;
    const linkKey = this.reserveResolution(
      {
        subject,
        authority,
        intent: {
          action: "resolve_item",
          resolvedRecordId: entry.receipt.data.row.record_id,
        },
      },
      binding,
      key,
    );
    if (linkKey === null) return false;
    this.updateCreation(key, { linkKey });
    return true;
  }
  private updateCreation(
    key: number,
    change: Partial<MentionCreationOperation>,
  ) {
    const entry = this.creations.get(key);
    if (entry) {
      this.creations.set(key, freezeMention({ ...entry, ...change }));
      this.publish();
    }
  }
  private settle(id: string) {
    if (this.remembered.delete(id)) this.accounting.settle(id);
  }
  private update(key: number, change: Partial<MentionOperation>) {
    const entry = this.entries.get(key);
    if (entry) {
      this.entries.set(key, freezeMention({ ...entry, ...change }));
      this.publish();
    }
  }
  private publish() {
    this.snapshot = freezeMention({
      authority: this.authority,
      generation: this.generation,
      entries: sameItems(
        this.snapshot.entries,
        this.authority ? [...this.entries.values()] : [],
      ),
      mentions: sameItems(
        this.snapshot.mentions,
        this.authority ? [...this.mentions.values()] : [],
      ),
      creations: sameItems(
        this.snapshot.creations,
        this.authority ? [...this.creations.values()] : [],
      ),
    });
    const { authority, generation, entries, creations } = this.snapshot;
    if (
      authority !== this.actionSnapshot.authority ||
      generation !== this.actionSnapshot.generation ||
      entries !== this.actionSnapshot.entries ||
      creations !== this.actionSnapshot.creations
    )
      this.actionSnapshot = freezeMention({
        authority,
        generation,
        entries,
        creations,
      });
    this.disclosureSnapshot = sameItems(
      this.disclosureSnapshot,
      this.authority ? [...this.disclosures.values()] : [],
    );
    for (const listener of this.listeners) listener();
  }
}
