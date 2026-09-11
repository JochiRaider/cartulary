import { afterEach, expect, it, vi } from "vitest";
import { timelineCaptureReview } from "../../../testing/timelineCaptureActionTestSupport";
import { createTimelineRecordActionAdapter } from "../adapters/createTimelineRecordActionAdapter";
import type {
  TimelineCaptureBinding,
  TimelineCaptureOutcome,
  TimelineRecordActionPort,
} from "../ports/TimelineRecordActionPort";
import type { TimelineCaptureReview } from "./timelineCaptureActionModel";
import { WorkbookTimelineCaptureActionOwner } from "./WorkbookTimelineCaptureActionOwner";

afterEach(() => vi.useRealTimers());
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((r) => {
    resolve = r;
  });
  return { promise, resolve };
}
function acknowledged(
  review = timelineCaptureReview(),
): TimelineCaptureOutcome {
  return {
    kind: "acknowledged",
    receipt: {
      operation: review.action,
      data: {
        record_id: review.target.recordId,
        incident_id: review.authority.incidentId,
        row_version: review.target.rowVersion + 1,
        capture_state:
          review.action === "mark-reviewed" ? "reviewed" : "superseded",
        reason: review.reason,
        replacement_record_id: review.replacement?.recordId ?? null,
        change_set_id: "30000000-0000-4000-8000-000000000001",
      },
    },
  };
}
function fixture(review: TimelineCaptureReview = timelineCaptureReview()) {
  const ids = { create: vi.fn(() => "secure-timeline-id") };
  const accounting = { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() };
  const owner = new WorkbookTimelineCaptureActionOwner(
    review.authority.incidentId,
    ids,
    accounting,
  );
  const send = vi
    .fn<TimelineRecordActionPort["send"]>()
    .mockResolvedValue(acknowledged(review));
  const adapter = createTimelineRecordActionAdapter({ apiBase: "/service" });
  const capture = vi.fn(adapter.capture);
  owner.configure(
    { capture, send },
    {
      page: vi.fn(async () => ({
        kind: "accepted" as const,
        value: {
          rows: [
            review.target,
            ...(review.replacement ? [review.replacement] : []),
          ],
          nextCursor: null,
          hasMore: false,
        },
      })),
    },
  );
  owner.setAuthority(review.authority);
  const binding: TimelineCaptureBinding = {
    isCurrent: vi.fn(() => true),
    matchesReview: vi.fn(() => true),
    prepare: vi.fn(async () => review.target),
  };
  const reconcile = vi.fn(async () => {});
  owner.registerReconciliation(reconcile);
  return { owner, ids, accounting, send, capture, binding, reconcile, review };
}
it("Timeline reserves before React updates and captures identity only after preparation", async () => {
  const f = fixture(),
    pending = deferred<ReturnType<typeof timelineCaptureReview>["target"]>();
  const binding = { ...f.binding, prepare: () => pending.promise };
  expect(f.owner.submit(f.review, binding)).toBe(true);
  expect(f.owner.submit(f.review, binding)).toBe(false);
  expect(f.ids.create).not.toHaveBeenCalled();
  expect(f.owner.blocksRecord(f.review.target.recordId)).toBe(true);
  expect(f.owner.blocksRecord("unrelated")).toBe(false);
  pending.resolve(f.review.target);
  await vi.waitFor(() =>
    expect(f.owner.getSnapshot().entries[0]?.reconciliation).toBe("complete"),
  );
  expect(f.ids.create).toHaveBeenCalledTimes(1);
  expect(f.send).toHaveBeenCalledTimes(1);
  expect(f.owner.pendingCount).toBe(0);
  expect(f.owner.blocksRecord(f.review.target.recordId)).toBe(false);
});
it("Timeline authoring preparation has no transaction and a newer committed version requires another click", async () => {
  const f = fixture();
  expect(
    await f.owner.prepareReview(
      f.review,
      f.binding,
      new AbortController().signal,
    ),
  ).toBe(true);
  expect(f.ids.create).not.toHaveBeenCalled();
  expect(f.owner.getSnapshot().entries).toEqual([]);
  const binding = {
    ...f.binding,
    prepare: async () => ({ ...f.review.target, rowVersion: 5 }),
  };
  expect(f.owner.submit(f.review, binding)).toBe(true);
  await vi.waitFor(() =>
    expect(f.owner.getSnapshot().entries[0]?.phase).toBe("preparation_failed"),
  );
  expect(f.send).not.toHaveBeenCalled();
  expect(f.ids.create).not.toHaveBeenCalled();
  expect(f.owner.latestVersion(f.review.target.recordId)).toBe(5);
});
it("Timeline invalidates admission for authority closure state and stale reviewed inputs", () => {
  for (const action of ["mark-reviewed", "supersede"] as const)
    for (const captureState of [
      "rough",
      "enriched",
      "reviewed",
      "superseded",
      "missing",
    ])
      for (const role of ["viewer", "editor", "reviewer", "admin"] as const)
        for (const closed of [false, true]) {
          const base = timelineCaptureReview();
          const review = timelineCaptureReview({
            action,
            target: { ...base.target, captureState },
            reason: action === "supersede" ? "Analyst reason" : null,
            authority: { ...base.authority, role, closed },
          });
          const f = fixture(review);
          const allowed =
            !closed &&
            ["reviewer", "admin"].includes(role) &&
            [
              "rough",
              "enriched",
              ...(action === "supersede" ? ["reviewed"] : []),
            ].includes(captureState);
          expect(
            f.owner.submit(review, { ...f.binding, prepare: async () => null }),
          ).toBe(allowed);
          f.owner.retire();
        }
  const f = fixture();
  f.owner.acceptVersion(f.review.target.recordId, 5);
  expect(f.owner.submit(f.review, f.binding)).toBe(false);
});
it("Timeline preserves preparation drafts and cancels changed selection before dispatch", async () => {
  for (const change of ["selection", "session", "missing"]) {
    const f = fixture(),
      pending = deferred<
        ReturnType<typeof timelineCaptureReview>["target"] | null
      >();
    let current = true;
    f.owner.submit(f.review, {
      ...f.binding,
      isCurrent: () => current,
      prepare: () => pending.promise,
    });
    if (change === "selection") current = false;
    if (change === "session")
      f.owner.setAuthority({
        ...f.review.authority,
        sessionIdentity: "replacement",
      });
    pending.resolve(change === "missing" ? null : f.review.target);
    await vi.waitFor(() =>
      expect(f.owner.getSnapshot().entries[0]?.phase).toBe(
        "preparation_failed",
      ),
    );
    expect(f.send).not.toHaveBeenCalled();
    expect(f.ids.create).not.toHaveBeenCalled();
  }
});
it("Timeline replays exactly after lost transport without fresh state or version eligibility", async () => {
  const base = timelineCaptureReview();
  const review = timelineCaptureReview({
    action: "supersede",
    reason: "Duplicate source",
    replacement: {
      ...base.target,
      recordId: "20000000-0000-4000-8000-000000000002",
    },
  });
  const f = fixture(review);
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  f.owner.submit(review, f.binding);
  await vi.waitFor(() =>
    expect(f.owner.getSnapshot().entries[0]?.phase).toBe("uncertain"),
  );
  const original = f.send.mock.calls[0]?.[0];
  f.owner.acceptVersion(review.target.recordId, 9);
  await f.owner.replay(firstKey(f.owner));
  expect(f.send.mock.calls[1]?.[0]).toBe(original);
  expect(f.capture).toHaveBeenCalledTimes(1);
  expect(f.ids.create).toHaveBeenCalledTimes(1);
  expect(f.owner.latestVersion(review.target.recordId)).toBe(9);
  expect(f.owner.getSnapshot().entries[0]?.receipt?.data.change_set_id).toBe(
    "30000000-0000-4000-8000-000000000001",
  );
});
it("Timeline retains timeout attempts and late receipts independently of presentation", async () => {
  vi.useFakeTimers();
  const f = fixture(),
    response = deferred<TimelineCaptureOutcome>();
  f.send.mockReturnValueOnce(response.promise);
  f.owner.submit(f.review, f.binding);
  await vi.advanceTimersByTimeAsync(0);
  await vi.advanceTimersByTimeAsync(30_000);
  expect(f.owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
  f.owner.acceptVersion(f.review.target.recordId, 12);
  response.resolve(acknowledged());
  await vi.advanceTimersByTimeAsync(0);
  expect(f.owner.getSnapshot().entries[0]?.phase).toBe("acknowledged");
  expect(f.owner.latestVersion(f.review.target.recordId)).toBe(12);
  expect(f.owner.pendingCount).toBe(0);
  expect(f.accounting.settle).toHaveBeenCalled();
});
it("Timeline permits explicit replay while a timed out transport has not settled", async () => {
  vi.useFakeTimers();
  const f = fixture(),
    response = deferred<TimelineCaptureOutcome>();
  f.send.mockReturnValueOnce(response.promise);
  f.owner.submit(f.review, f.binding);
  await vi.advanceTimersByTimeAsync(30_000);
  const key = firstKey(f.owner);
  await f.owner.replay(key);
  expect(f.send).toHaveBeenCalledTimes(2);
  expect(f.owner.blocksRecord(f.review.target.recordId)).toBe(false);
  response.resolve({
    kind: "rejected",
    failure: { kind: "stale_target", message: "Late rejection" },
  });
  await vi.advanceTimersByTimeAsync(0);
  expect(f.owner.getSnapshot().entries[0]?.phase).toBe("acknowledged");
  expect(f.accounting.accepted).toHaveBeenCalledTimes(1);
});
it("Timeline records acknowledgement before refresh and refresh recovery sends no mutation", async () => {
  const f = fixture();
  f.reconcile.mockImplementationOnce(async () => {
    expect(f.owner.getSnapshot().entries[0]?.receipt).not.toBeNull();
    throw new Error("Projection missing");
  });
  f.owner.submit(f.review, f.binding);
  await vi.waitFor(() =>
    expect(f.owner.getSnapshot().entries[0]?.reconciliation).toBe("required"),
  );
  await f.owner.refresh(firstKey(f.owner));
  expect(f.owner.getSnapshot().entries[0]?.reconciliation).toBe("complete");
  expect(f.send).toHaveBeenCalledTimes(1);
});
it("Timeline conceals retained attempts during access loss and fences account retirement", async () => {
  const f = fixture(),
    response = deferred<TimelineCaptureOutcome>();
  f.send.mockReturnValueOnce(response.promise);
  f.owner.submit(f.review, f.binding);
  await vi.waitFor(() => expect(f.send).toHaveBeenCalledTimes(1));
  f.owner.suspend();
  expect(f.owner.getSnapshot().entries).toEqual([]);
  response.resolve(acknowledged());
  await vi.waitFor(() =>
    expect(f.accounting.accepted).toHaveBeenCalledTimes(1),
  );
  f.owner.setAuthority({ ...f.review.authority, sessionIdentity: "restored" });
  expect(f.owner.getSnapshot().entries[0]?.receipt).not.toBeNull();
  f.owner.setAuthority({ ...f.review.authority, actorId: "another-account" });
  expect(f.owner.getSnapshot().entries).toEqual([]);
  const recovered = fixture(),
    staleFailure = deferred<TimelineCaptureOutcome>();
  recovered.send.mockReturnValueOnce(staleFailure.promise);
  recovered.owner.submit(recovered.review, recovered.binding);
  await vi.waitFor(() => expect(recovered.send).toHaveBeenCalledTimes(1));
  recovered.owner.suspend();
  recovered.owner.setAuthority({
    ...recovered.review.authority,
    sessionIdentity: "restored",
  });
  staleFailure.resolve({
    kind: "rejected",
    failure: {
      kind: "authorization_lost",
      publicCode: "authorization_denied",
      message: "Old session expired",
    },
  });
  await vi.waitFor(() =>
    expect(recovered.owner.getSnapshot().entries[0]?.phase).toBe("rejected"),
  );
  expect(recovered.owner.getSnapshot().authority?.sessionIdentity).toBe(
    "restored",
  );
});
it("Timeline exact recovery rejection leaves the original outcome uncertain", async () => {
  const f = fixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" }).mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "client_txn_conflict", message: "Transaction conflict" },
  });
  f.owner.submit(f.review, f.binding);
  await vi.waitFor(() =>
    expect(f.owner.getSnapshot().entries[0]?.phase).toBe("uncertain"),
  );
  await f.owner.replay(firstKey(f.owner));
  expect(f.owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
  expect(f.owner.getSnapshot().entries[0]?.failure?.kind).toBe(
    "client_txn_conflict",
  );
});

function firstKey(owner: WorkbookTimelineCaptureActionOwner) {
  const entry = owner.getSnapshot().entries[0];
  if (!entry) throw new Error("Expected retained attempt");
  return entry.key;
}
