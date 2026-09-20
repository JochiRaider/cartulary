import { afterEach, expect, it, vi } from "vitest";
import {
  mentionCreationReceipt,
  mentionReceipt,
  mentionReview,
  mentionWorkbookRow,
} from "../../../testing/timelineMentionTestSupport";
import { createWorkbookPendingMutationAdapter } from "../../adapters/createWorkbookPendingMutationAdapter";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { createTimelineMentionEntityCreationAdapter } from "../adapters/createTimelineMentionEntityCreationAdapter";
import { createTimelineMentionResolutionAdapter } from "../adapters/createTimelineMentionResolutionAdapter";
import type {
  TimelineMentionEntityCreationPort,
  TimelineMentionResolutionPort,
} from "../ports/TimelineMentionPort";
import {
  initialMentionCreateDraft,
  type MentionCreationOutcome,
} from "./timelineMentionCreationModel";
import {
  type MentionBinding,
  type MentionOutcome,
  type MentionReview,
  type MentionSubject,
  mentionTransitionAllowed,
} from "./timelineMentionOperationModel";
import { timelineMentionOwnerFor } from "./timelineMentionOwnerFor";
import { WorkbookTimelineMentionOperationOwner } from "./WorkbookTimelineMentionOperationOwner";

afterEach(() => vi.useRealTimers());
function creationFixture() {
  const f = fixture();
  const review = {
    subject: f.review.subject,
    authority: f.review.authority,
    draft: initialMentionCreateDraft(f.review.subject),
  };
  const receipt = mentionCreationReceipt(review);
  const create = vi
    .fn<TimelineMentionEntityCreationPort["send"]>()
    .mockResolvedValue({ kind: "accepted", receipt });
  f.owner.configureCreation({
    capture: createTimelineMentionEntityCreationAdapter({ apiBase: "/service" })
      .capture,
    send: create,
  });
  const refreshEntity = vi.fn(async () => {});
  f.owner.registerCreationReconciliation(refreshEntity);
  return {
    ...f,
    createReview: review,
    create,
    createReceipt: receipt,
    refreshEntity,
  };
}
it("Mention creation retains its receipt before rejected linking and renewed resolution never creates twice", async () => {
  const f = creationFixture();
  f.refreshEntity.mockRejectedValueOnce(new Error("Entity sheet unavailable"));
  f.send.mockImplementationOnce(async () => {
    expect(f.owner.getSnapshot().creations[0]?.receipt).toEqual(
      f.createReceipt,
    );
    return {
      kind: "rejected",
      failure: { kind: "stale_target", message: "Mention changed" },
    };
  });
  expect(f.owner.createAndResolve(f.createReview, f.binding)).toBe(true);
  expect(f.owner.createAndResolve(f.createReview, f.binding)).toBe(false);
  await flush();
  expect(f.owner.getSnapshot().entries[0]?.phase).toBe("rejected");
  expect(f.owner.getSnapshot().creations[0]?.refresh).toBe("required");
  await f.owner.refreshCreation(1);
  expect(f.owner.getSnapshot().creations[0]?.refresh).toBe("complete");
  expect(f.refreshEntity).toHaveBeenCalledTimes(2);
  expect(f.send).toHaveBeenCalledTimes(1);
  expect(f.create).toHaveBeenCalledTimes(1);
  const subject = {
    ...f.review.subject,
    mentionRowVersion: 3,
    sourceRowVersion: 6,
  };
  const binding = { ...f.binding, prepare: async () => subject };
  expect(f.owner.linkCreated(1, subject, binding)).toBe(true);
  await flush();
  expect(f.send.mock.calls[1]?.[0].review.subject.mentionRowVersion).toBe(3);
  expect(f.create).toHaveBeenCalledTimes(1);
});
it("Mention creation followed by uncertain linking recovers only the original resolution request", async () => {
  const f = creationFixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  f.owner.createAndResolve(f.createReview, f.binding);
  await flush();
  const resolution = f.owner.getSnapshot().entries[0];
  expect(resolution?.phase).toBe("uncertain");
  expect(f.owner.linkCreated(1, f.review.subject, f.binding)).toBe(false);
  await f.owner.replay(resolution?.key ?? -1);
  expect(f.send.mock.calls[1]?.[0]).toBe(f.send.mock.calls[0]?.[0]);
  expect(f.create).toHaveBeenCalledTimes(1);
  expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("complete");
});
it("Mention creation accepted after navigation retains the entity until renewed link review", async () => {
  const f = creationFixture(),
    response = deferred<MentionCreationOutcome>();
  f.create.mockReturnValueOnce(response.promise);
  let current = true;
  f.owner.createAndResolve(f.createReview, {
    ...f.binding,
    isCurrent: () => current,
  });
  await flush();
  current = false;
  response.resolve({ kind: "accepted", receipt: f.createReceipt });
  await flush();
  expect(f.owner.getSnapshot().creations[0]).toMatchObject({
    receipt: f.createReceipt,
    linkKey: null,
  });
  expect(f.send).not.toHaveBeenCalled();
  expect(f.owner.blocksMention(f.review.subject.mentionId)).toBe(false);
  expect(f.owner.linkCreated(1, f.review.subject, f.binding)).toBe(true);
  await flush();
  expect(f.create).toHaveBeenCalledTimes(1);
  expect(f.send).toHaveBeenCalledTimes(1);
});
it("Mention uncertain creation replays its exact attempt and requires renewed review before linking", async () => {
  const f = creationFixture();
  f.create.mockResolvedValueOnce({ kind: "uncertain" });
  f.owner.createAndResolve(f.createReview, f.binding);
  await flush();
  await f.owner.replayCreation(1);
  expect(f.create.mock.calls[1]?.[0]).toBe(f.create.mock.calls[0]?.[0]);
  expect(f.owner.getSnapshot().creations[0]?.receipt).toEqual(f.createReceipt);
  expect(f.send).not.toHaveBeenCalled();
  expect(f.owner.createAndResolve(f.createReview, f.binding)).toBe(false);
});
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((r) => {
    resolve = r;
  });
  return { promise, resolve };
}
async function flush() {
  for (let index = 0; index < 20; index++) await Promise.resolve();
}
function fixture(review: MentionReview = mentionReview()) {
  let sequence = 0;
  const ids = { create: vi.fn(() => `secure-mention-${++sequence}`) };
  const accounting = { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() };
  const owner = new WorkbookTimelineMentionOperationOwner(
    review.subject.incidentId,
    ids,
    accounting,
  );
  const send = vi
    .fn<TimelineMentionResolutionPort["send"]>()
    .mockResolvedValue({ kind: "accepted", receipt: mentionReceipt(review) });
  const capture = vi.fn(
    createTimelineMentionResolutionAdapter({ apiBase: "/service" }).capture,
  );
  const recheck = vi.fn();
  owner.configure({ capture, send }, recheck);
  owner.setAuthority(review.authority);
  const binding: MentionBinding = {
    isCurrent: vi.fn(() => true),
    prepare: vi.fn(async () => review.subject),
  };
  const reconcile = vi.fn(async () => {});
  const detach = owner.registerReconciliation(reconcile);
  return {
    owner,
    ids,
    accounting,
    capture,
    send,
    recheck,
    binding,
    reconcile,
    detach,
    review,
  };
}
it("Mention operations reserve immutable intent and identity before preceding saves without duplicate activation", async () => {
  const f = fixture(),
    barrier = deferred<MentionSubject | null>();
  const binding = { ...f.binding, prepare: () => barrier.promise };
  expect(f.owner.submit(f.review, binding)).toBe(true);
  expect(f.owner.submit(f.review, binding)).toBe(false);
  const attempt = f.owner.getSnapshot().entries[0]?.attempt;
  expect(f.ids.create).toHaveBeenCalledTimes(1);
  expect(Object.isFrozen(attempt?.review.subject)).toBe(true);
  expect(f.send).not.toHaveBeenCalled();
  barrier.resolve({ ...f.review.subject, sourceRowVersion: 5 });
  await flush();
  expect(f.send).toHaveBeenCalledTimes(1);
  expect(f.send.mock.calls[0]?.[0]).toBe(attempt);
  expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("complete");
});
it("Mention operations require renewed review after changed intent selection role or mention version", async () => {
  for (const change of ["version", "selection", "role", "target"] as const) {
    const f = fixture(),
      barrier = deferred<MentionSubject | null>();
    let current = true;
    f.owner.submit(f.review, {
      prepare: () => barrier.promise,
      isCurrent: () => current,
    });
    if (change === "selection") current = false;
    if (change === "role")
      f.owner.setAuthority({ ...f.review.authority, role: "viewer" });
    barrier.resolve({
      ...f.review.subject,
      ...(change === "version" ? { mentionRowVersion: 3 } : {}),
      ...(change === "target" ? { resolvedRecordId: "different" } : {}),
    });
    await flush();
    expect(f.send).not.toHaveBeenCalled();
    expect(f.owner.getSnapshot().entries[0]?.phase).toBe("preparation_failed");
  }
});
it("Mention operations enforce the complete legal transition and current authorization matrix", () => {
  for (const state of ["unresolved", "resolved", "dismissed"] as const)
    for (const action of [
      "resolve_item",
      "dismiss_item",
      "revert_to_unresolved",
    ] as const)
      for (const role of ["viewer", "editor", "reviewer", "admin"] as const) {
        const base = mentionReview(),
          review = mentionReview({
            subject: { ...base.subject, state },
            intent: action === "resolve_item" ? base.intent : { action },
            authority: { ...base.authority, role },
          });
        const f = fixture(review);
        const legal =
          action === "revert_to_unresolved"
            ? state !== "unresolved"
            : state !== "dismissed";
        expect(mentionTransitionAllowed(state, action)).toBe(legal);
        expect(f.owner.submit(review, f.binding)).toBe(
          legal && role !== "viewer",
        );
        f.owner.retire();
      }
  const f = fixture();
  for (const authority of [
    { ...f.review.authority, actions: [] },
    { ...f.review.authority, mutationsAvailable: false },
  ]) {
    f.owner.setAuthority(authority);
    expect(f.owner.submit({ ...f.review, authority }, f.binding)).toBe(false);
  }
});
it("Mention exact replay keeps original body version route and key despite newer observations", async () => {
  const f = fixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  f.owner.submit(f.review, f.binding);
  await flush();
  const original = f.owner.getSnapshot().entries[0]?.attempt;
  f.owner.observeMention({ ...f.review.subject, mentionRowVersion: 20 });
  f.owner.acceptVersion(f.review.subject.sourceRecordId, 30);
  await f.owner.replay(1);
  expect(f.send.mock.calls[1]?.[0]).toBe(original);
  expect(f.ids.create).toHaveBeenCalledTimes(1);
  expect(f.owner.latestVersion(f.review.subject.sourceRecordId)).toBe(30);
  expect(
    f.owner.latestMention(f.review.subject.mentionId)?.mentionRowVersion,
  ).toBe(20);
  expect(f.owner.getSnapshot().entries[0]?.receipt).toEqual(mentionReceipt());
});
it("Mention replay rejection preserves original uncertainty and local target errors preserve recovery", async () => {
  const f = fixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" }).mockResolvedValueOnce({
    kind: "rejected",
    failure: {
      kind: "stale_target",
      publicCode: "record_not_found",
      message: "Target no longer available",
    },
  });
  f.owner.submit(f.review, f.binding);
  await flush();
  await f.owner.replay(1);
  expect(f.owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
  expect(f.owner.getSnapshot().authority).not.toBeNull();
  expect(f.recheck).not.toHaveBeenCalled();
  f.owner.setAuthority({ ...f.review.authority, role: "viewer" });
  await f.owner.replay(1);
  expect(f.send).toHaveBeenCalledTimes(2);
  f.owner.setAuthority(f.review.authority);
  await f.owner.replay(1);
  expect(f.owner.getSnapshot().entries[0]?.phase).toBe("accepted");
});
it("Mention accepted receipts survive detached presentation and refresh recovery never repeats mutations", async () => {
  const f = fixture(),
    response = deferred<MentionOutcome>();
  f.send.mockReturnValueOnce(response.promise);
  f.owner.submit(f.review, f.binding);
  await flush();
  f.detach();
  response.resolve({ kind: "accepted", receipt: mentionReceipt() });
  await flush();
  expect(f.owner.getSnapshot().entries[0]).toMatchObject({
    phase: "accepted",
    receipt: mentionReceipt(),
    refresh: "required",
  });
  f.owner.registerReconciliation(async () => {
    throw new Error("Refresh offline");
  });
  await f.owner.refresh(1);
  expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("required");
  f.owner.registerReconciliation(f.reconcile);
  await f.owner.refresh(1);
  expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("complete");
  expect(f.send).toHaveBeenCalledTimes(1);
  expect(f.accounting.accepted).toHaveBeenCalledWith(
    f.review.subject.sourceRecordId,
    5,
  );
});
it("Mention timeout permits exact replay and retains a late accepted receipt", async () => {
  vi.useFakeTimers();
  const f = fixture(),
    original = deferred<MentionOutcome>(),
    replay = deferred<MentionOutcome>();
  f.send
    .mockReturnValueOnce(original.promise)
    .mockReturnValueOnce(replay.promise);
  f.owner.submit(f.review, f.binding);
  await flush();
  await vi.advanceTimersByTimeAsync(30_000);
  expect(f.owner.getSnapshot().entries[0]).toMatchObject({
    phase: "uncertain",
    transportPending: true,
  });
  const replaying = f.owner.replay(1);
  await flush();
  expect(f.send.mock.calls[1]?.[0]).toBe(f.send.mock.calls[0]?.[0]);
  original.resolve({ kind: "accepted", receipt: mentionReceipt() });
  await flush();
  replay.resolve({
    kind: "rejected",
    failure: { kind: "stale_target", message: "Stale replay" },
  });
  await replaying;
  expect(f.owner.getSnapshot().entries[0]).toMatchObject({
    phase: "accepted",
    receipt: mentionReceipt(),
    refresh: "complete",
    transportPending: false,
  });
});
it("Mention source and mention versions remain monotonic in both HTTP socket orderings", async () => {
  for (const socketFirst of [true, false]) {
    const f = fixture(),
      pending = deferred<MentionOutcome>();
    f.send.mockReturnValueOnce(pending.promise);
    f.owner.submit(f.review, f.binding);
    await flush();
    const remote = {
      ...f.review.subject,
      sourceRowVersion: 7,
      mentionRowVersion: 4,
      state: "dismissed" as const,
    };
    if (socketFirst) f.owner.observeMention(remote);
    pending.resolve({ kind: "accepted", receipt: mentionReceipt() });
    await flush();
    if (!socketFirst) f.owner.observeMention(remote);
    f.owner.observeMention(f.review.subject);
    expect(f.owner.latestMention(remote.mentionId)).toEqual(remote);
    expect(f.owner.latestVersion(remote.sourceRecordId)).toBe(7);
    expect(f.owner.getSnapshot().entries[0]?.receipt).toEqual(mentionReceipt());
  }
});
it("Mention suspension conceals protected recovery while account retirement fences late acceptance", async () => {
  const f = fixture(),
    response = deferred<MentionOutcome>();
  f.send.mockReturnValueOnce(response.promise);
  f.owner.submit(f.review, f.binding);
  await flush();
  f.owner.suspend();
  response.resolve({ kind: "accepted", receipt: mentionReceipt() });
  await flush();
  expect(f.owner.getSnapshot()).toMatchObject({
    authority: null,
    entries: [],
    mentions: [],
  });
  f.owner.setAuthority(f.review.authority);
  expect(f.owner.getSnapshot().entries[0]?.receipt).toEqual(mentionReceipt());
  f.owner.setAuthority({ ...f.review.authority, actorId: "another-account" });
  expect(f.owner.getSnapshot().entries).toEqual([]);
  const g = fixture(),
    late = deferred<MentionOutcome>();
  g.send.mockReturnValueOnce(late.promise);
  g.owner.submit(g.review, g.binding);
  await flush();
  g.owner.retire();
  late.resolve({ kind: "accepted", receipt: mentionReceipt() });
  await flush();
  g.owner.setAuthority(g.review.authority);
  expect(g.owner.getSnapshot().entries).toEqual([]);
  expect(g.accounting.accepted).not.toHaveBeenCalled();
});
it("Mention runtime retention accounts explicit work without consuming autosave capacity", async () => {
  const review = mentionReview();
  const runtime = new WorkbookMutationRuntime(
    {
      incidentId: review.subject.incidentId,
      clientInstanceId: "mention-runtime",
    },
    { create: () => "runtime-mention-key" },
    createWorkbookPendingMutationAdapter({
      apiBase: undefined,
      incidentId: review.subject.incidentId,
    }),
  );
  const owner = timelineMentionOwnerFor(runtime),
    response = deferred<MentionOutcome>();
  expect(timelineMentionOwnerFor(runtime)).toBe(owner);
  const adapter = createTimelineMentionResolutionAdapter({
    apiBase: undefined,
  });
  owner.configure({ ...adapter, send: () => response.promise });
  owner.setAuthority(review.authority);
  owner.submit(review, {
    isCurrent: () => true,
    prepare: async () => review.subject,
  });
  await flush();
  expect(runtime.pendingQueue().model.snapshot().units).toHaveLength(0);
  expect(runtime.getSnapshot().explicitInFlightCount).toBe(1);
  expect(
    runtime.timelineActionBlocksRecord(review.subject.sourceRecordId),
  ).toBe(true);
  runtime.observeTimelineVersion(review.subject.sourceRecordId, 40);
  response.resolve({ kind: "accepted", receipt: mentionReceipt() });
  await flush();
  expect(owner.latestVersion(review.subject.sourceRecordId)).toBe(40);
  runtime.invalidate({ kind: "runtime_disposed" });
  expect(owner.getSnapshot().entries).toEqual([]);
});

function autoDisclosureFixture() {
  const base = mentionReview();
  const review = mentionReview({
    subject: {
      ...base.subject,
      state: "resolved",
      resolutionMethod: "auto_match",
      resolvedRecordId: "70000000-0000-4000-8000-000000000001",
    },
    intent: { action: "revert_to_unresolved" },
  });
  const f = fixture(review);
  const source = mentionWorkbookRow(review);
  const row = {
    ...source,
    collectionValues: {
      ...source.collectionValues,
      hostRefs: source.collectionValues.hostRefs.map((item) => ({
        ...item,
        displayText: "Canonical target",
        matchedAliasText: "VPN alias",
      })),
    },
  };
  const before = {
    ...row,
    rowVersion: (row.rowVersion ?? 1) - 1,
    collectionValues: {
      ...row.collectionValues,
      hostRefs: [],
      identityRefs: [],
    },
  };
  const operation = {
    kind: "entry" as const,
    operationId: "first-entry",
    changeSetId: "first-change",
  };
  f.owner.acceptAutoResolutions([before], [row], operation);
  return { ...f, row, before, operation };
}
it("Auto-resolution disclosures retain source identity and canonical facts through version changes and missing pages", () => {
  const f = autoDisclosureFixture();
  const initial = f.owner.getDisclosureSnapshot()[0];
  expect(initial).toMatchObject({
    displayText: "Canonical target",
    matchedAliasText: "VPN alias",
  });
  f.owner.acceptVersion("unrelated", 5);
  expect(f.owner.getDisclosureSnapshot()[0]).toBe(initial);
  const newer = {
    ...f.row,
    rowVersion: 10,
    collectionValues: {
      ...f.row.collectionValues,
      hostRefs: f.row.collectionValues.hostRefs.map((item) => ({
        ...item,
        mentionRowVersion: 3,
        resolvedRecordId: "new-target",
        displayText: "New canonical",
      })),
    },
  };
  f.owner.observeSource(newer);
  expect(f.owner.getDisclosureSnapshot()[0]).toMatchObject({
    identity: initial?.identity,
    mentionRowVersion: 3,
    resolvedRecordId: "new-target",
    displayText: "New canonical",
  });
  f.owner.observeSource(f.before);
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  f.owner.observeSource({
    ...newer,
    rowVersion: 11,
    collectionValues: { ...newer.collectionValues, hostRefs: [] },
  });
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(0);
  f.owner.acceptAutoResolutions([f.before], [f.row], {
    ...f.operation,
    operationId: "late-receipt",
  });
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(0);
});
it("Auto-resolution batch counts belong to captured change sets and replay cannot duplicate or resurrect disclosure", () => {
  const f = autoDisclosureFixture();
  const second = {
    ...f.row,
    recordId: "second-source",
    key: "second-source",
    collectionValues: {
      ...f.row.collectionValues,
      hostRefs: f.row.collectionValues.hostRefs.map((item) => ({
        ...item,
        entityMentionId: "second-mention",
        itemRef: "second-ref",
      })),
    },
  };
  const third = {
    ...second,
    recordId: "third-source",
    key: "third-source",
    collectionValues: {
      ...second.collectionValues,
      hostRefs: second.collectionValues.hostRefs.map((item) => ({
        ...item,
        entityMentionId: "third-mention",
        itemRef: "third-ref",
      })),
    },
  };
  const empty = [second, third].map((row) => ({
    ...row,
    collectionValues: { ...row.collectionValues, hostRefs: [] },
  }));
  const batch = {
    kind: "batch" as const,
    operationId: "paste-one",
    changeSetId: "batch-change",
  };
  // A query can arrive before the retained operation receipt.
  f.owner.observeSource(second);
  f.owner.observeSource(third);
  f.owner.acceptAutoResolutions(empty, [second, third], batch);
  expect(
    f.owner.getDisclosureSnapshot().map((item) => item.acceptedCount),
  ).toEqual([1, 2, 2]);
  expect(
    new Set(f.owner.getDisclosureSnapshot().map((item) => item.identity)).size,
  ).toBe(3);
  f.owner.observeSource({
    ...second,
    rowVersion: 20,
    collectionValues: { ...second.collectionValues, hostRefs: [] },
  });
  f.owner.acceptAutoResolutions(empty, [second, third], batch);
  expect(
    f.owner.getDisclosureSnapshot().map((item) => item.acceptedCount),
  ).toEqual([1, 2]);
  // Ordinary scalar paste contains no newly admitted mentions.
  f.owner.acceptAutoResolutions([third], [{ ...third, rowVersion: 21 }], {
    ...batch,
    operationId: "scalar-paste",
    changeSetId: "scalar-change",
  });
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(2);
  // Created batch rows have no previous source but an authoritative initial version.
  const created = {
    ...second,
    recordId: "created-source",
    key: "created-source",
    rowVersion: 1,
    collectionValues: {
      ...second.collectionValues,
      hostRefs: second.collectionValues.hostRefs.map((item) => ({
        ...item,
        entityMentionId: "created-mention",
        itemRef: "created-ref",
      })),
    },
  };
  f.owner.acceptAutoResolutions([], [created], {
    ...batch,
    operationId: "created-batch",
    changeSetId: "created-change",
  });
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(3);
  expect(f.owner.getDisclosureSnapshot().at(-1)).toMatchObject({
    rowRecordId: "created-source",
    acceptedCount: 1,
  });
});
it("Auto-resolution correction retains failed and uncertain disclosure but accepted failed refresh never resends", async () => {
  const f = autoDisclosureFixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  f.reconcile.mockRejectedValueOnce(new Error("Source refresh failed"));
  expect(f.owner.submit(f.review, f.binding)).toBe(true);
  await flush();
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  const key = f.owner.getSnapshot().entries[0]?.key ?? -1;
  await f.owner.replay(key);
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(0);
  expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("required");
  expect(f.send.mock.calls[0]?.[0]).toBe(f.send.mock.calls[1]?.[0]);
  await f.owner.refresh(key);
  await f.owner.replay(key);
  expect(f.send).toHaveBeenCalledTimes(2);
  f.owner.acceptAutoResolutions([f.before], [f.row], f.operation);
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(0);
});
it("Auto-resolution suspension conceals disclosure and same-account recovery preserves it while replacement retires it", () => {
  const f = autoDisclosureFixture();
  f.owner.setAuthority({ ...f.review.authority, role: "viewer" });
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  expect(f.owner.canSubmit("revert_to_unresolved")).toBe(false);
  f.owner.suspend();
  expect(f.owner.getDisclosureSnapshot()).toEqual([]);
  f.owner.setAuthority(f.review.authority);
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  f.owner.setAuthority({
    ...f.review.authority,
    actorId: "replacement-account",
  });
  expect(f.owner.getDisclosureSnapshot()).toEqual([]);
});
it("Auto-resolution narrow snapshots ignore version-only publication and preserve meaningful action changes", () => {
  const f = autoDisclosureFixture();
  const mentions = f.owner.getMentionsSnapshot(),
    actions = f.owner.getActionSnapshot(),
    disclosures = f.owner.getDisclosureSnapshot();
  const notified = vi.fn();
  f.owner.subscribe(notified);
  f.owner.acceptVersion("unrelated-source", 99);
  expect(notified).toHaveBeenCalledOnce();
  expect(f.owner.getMentionsSnapshot()).toBe(mentions);
  expect(f.owner.getActionSnapshot()).toBe(actions);
  expect(f.owner.getDisclosureSnapshot()).toBe(disclosures);
  f.owner.submit(f.review, f.binding);
  expect(f.owner.getActionSnapshot()).not.toBe(actions);
});
