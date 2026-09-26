import { act, renderHook } from "@testing-library/react";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import {
  mentionCreationReceipt,
  mentionInspector,
  mentionReceipt,
  mentionReview,
  mentionWorkbookRow,
} from "../../testing/timelineMentionTestSupport";
import type { MentionReview } from "./actions/timelineMentionOperationModel";
import { WorkbookTimelineMentionOperationOwner } from "./actions/WorkbookTimelineMentionOperationOwner";
import { createTimelineMentionEntityCreationAdapter } from "./adapters/createTimelineMentionEntityCreationAdapter";
import { createTimelineMentionResolutionAdapter } from "./adapters/createTimelineMentionResolutionAdapter";
import { useTimelineMentionActions } from "./hooks/useTimelineMentionActions";
import type { DisclosureReviewNavigationScope } from "./models/timelineControllerPorts";
import type { WorkbookRow } from "./models/timelineRowModel";
import type { TimelineMentionResolutionPort } from "./ports/TimelineMentionPort";

function setup(review: MentionReview = mentionReview()) {
  let sequence = 0;
  let focusToken = 8;
  const owner = new WorkbookTimelineMentionOperationOwner(
    review.subject.incidentId,
    { create: () => (++sequence === 1 ? "hook-key" : `hook-key-${sequence}`) },
    { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() },
  );
  const send = vi
    .fn<TimelineMentionResolutionPort["send"]>()
    .mockResolvedValue({ kind: "accepted", receipt: mentionReceipt(review) });
  owner.configure({
    ...createTimelineMentionResolutionAdapter({ apiBase: undefined }),
    send,
  });
  owner.setAuthority(review.authority);
  owner.registerReconciliation(async () => {});
  const rowsRef = { current: [mentionWorkbookRow(review)] };
  const input = {
    owner,
    rowsRef,
    acceptDisclosureSource: (row: WorkbookRow) => ({
      accepted: true,
      row,
      stale: false,
    }),
    earlierSaves: { current: Promise.resolve() },
    selectedMention: mentionInspector(review),
    selectedMentionRef: review.subject.itemRef as string | null,
    selectedRowId: review.subject.sourceRecordId,
    inspectorInvalidationGeneration: 0,
    reviewSurfaceKey: "timeline:source",
    presentationKey: "timeline:source",
    presentationActive: true,
    candidatePort: {
      page: vi.fn(async () => ({
        kind: "accepted" as const,
        value: { candidates: [], nextCursor: null, hasMore: false },
      })),
    },
    waitForCommittedRecordIdle: vi.fn(async () => ({
      row: rowsRef.current[0] ?? null,
      rowVersion: rowsRef.current[0]?.rowVersion ?? 0,
    })),
    setInspectorMessage: vi.fn(),
    focusContinuity: {
      beginViewportContinuity: vi.fn(() => ++focusToken),
      advanceViewportContinuity: vi.fn(),
      settleViewportContinuityFollowUp: vi.fn(),
      clearViewportContinuity: vi.fn(),
    },
  };
  const hook = renderHook(
    (props: typeof input) => {
      const [selectedTargetId, setSelectedTargetId] = useState("");
      return useTimelineMentionActions({
        ...props,
        selectedTargetId,
        setSelectedTargetId,
      });
    },
    { initialProps: input },
  );
  return { ...hook, owner, send, input, rowsRef, review };
}
async function flush() {
  await act(async () => {
    for (let index = 0; index < 20; index++) await Promise.resolve();
  });
}
function disclosureFixture() {
  const base = mentionReview();
  const f = setup(
    mentionReview({
      subject: {
        ...base.subject,
        state: "resolved",
        resolutionMethod: "auto_match",
        resolvedRecordId: "70000000-0000-4000-8000-000000000001",
      },
      intent: { action: "revert_to_unresolved" },
    }),
  );
  const row = f.rowsRef.current[0];
  if (!row) throw new Error("Source fixture required");
  act(() =>
    f.owner.acceptAutoResolutions(
      [{ ...row, collectionValues: { ...row.collectionValues, hostRefs: [] } }],
      [row],
      {
        kind: "entry",
        operationId: "review-match",
        changeSetId: "review-change",
      },
    ),
  );
  const notice = f.owner.getDisclosureSnapshot()[0];
  if (!notice) throw new Error("Disclosure required");
  f.rowsRef.current = [];
  return { ...f, row, notice };
}
const navigateReview = vi.fn(
  (
    _recordId: string,
    _itemRef: string,
    scope: DisclosureReviewNavigationScope,
  ) => scope.settle(true),
);
describe("Timeline mention actions", () => {
  it("binds auto-resolution undo to the same current mention version and target", async () => {
    const base = mentionReview();
    const review = mentionReview({
      subject: {
        ...base.subject,
        state: "resolved",
        resolvedRecordId: "70000000-0000-4000-8000-000000000001",
        resolutionMethod: "auto_match",
      },
      intent: { action: "revert_to_unresolved" },
    });
    const f = setup(review);
    const notice = {
      entityMentionId: review.subject.mentionId,
      mentionRowVersion: review.subject.mentionRowVersion,
      itemRef: review.subject.itemRef,
      rowRecordId: review.subject.sourceRecordId,
      fieldKey: review.subject.sourceFieldKey,
      entityType: review.subject.entityType,
      rawText: review.subject.rawText,
      resolvedRecordId: review.subject.resolvedRecordId ?? "",
      matchedAliasText: "alias",
    };
    act(() =>
      f.result.current.handleUndoAutoResolutionNotice({
        ...notice,
        mentionRowVersion: 1,
      }),
    );
    await flush();
    expect(f.send).not.toHaveBeenCalled();
    act(() =>
      f.result.current.handleUndoAutoResolutionNotice({
        ...notice,
        resolvedRecordId: "another-target",
      }),
    );
    await flush();
    expect(f.send).not.toHaveBeenCalled();
    act(() => f.result.current.handleUndoAutoResolutionNotice(notice));
    await flush();
    expect(JSON.parse(f.send.mock.calls[0]?.[0].body ?? "{}")).toEqual({
      action: "revert_to_unresolved",
      base_mention_row_version: 2,
      client_txn_id: "hook-key",
    });
    expect(
      f.input.focusContinuity.advanceViewportContinuity,
    ).toHaveBeenCalledWith(9);
    f.unmount();
  });
  it("waits for creation and mention projections before restoring focus", async () => {
    const f = setup();
    let release!: () => void;
    f.owner.registerCreationReconciliation(
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        }),
    );
    act(() => f.result.current.startCreate());
    const review = f.result.current.createReview;
    if (!review) throw new Error("Create review missing");
    const receipt = mentionCreationReceipt(review);
    f.owner.configureCreation({
      ...createTimelineMentionEntityCreationAdapter({ apiBase: undefined }),
      send: async () => ({ kind: "accepted", receipt }),
    });
    f.send.mockImplementation(async (attempt) => ({
      kind: "accepted",
      receipt: mentionReceipt(attempt.review),
    }));
    act(() => f.result.current.submitCreate());
    await flush();
    expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("complete");
    expect(f.owner.getSnapshot().creations[0]?.refresh).toBe("refreshing");
    expect(
      f.input.focusContinuity.advanceViewportContinuity,
    ).not.toHaveBeenCalled();
    release();
    await flush();
    expect(
      f.input.focusContinuity.advanceViewportContinuity,
    ).toHaveBeenCalledExactlyOnceWith(9);
    f.unmount();
  });
  it("waits for preceding saves and refuses changed mentions instead of rebasing", async () => {
    const f = setup(mentionReview({ intent: { action: "dismiss_item" } }));
    let release!: () => void;
    f.input.earlierSaves.current = new Promise<void>((resolve) => {
      release = resolve;
    });
    act(() => f.result.current.act({ action: "dismiss_item" }));
    expect(f.send).not.toHaveBeenCalled();
    f.rowsRef.current = [
      mentionWorkbookRow({
        ...f.review,
        subject: { ...f.review.subject, mentionRowVersion: 3 },
      }),
    ];
    release();
    await flush();
    expect(f.send).not.toHaveBeenCalled();
    expect(f.owner.getSnapshot().entries[0]?.phase).toBe("preparation_failed");
    expect(f.owner.getSnapshot().entries[0]?.attempt.body).toContain(
      '"base_mention_row_version":2',
    );
    f.unmount();
  });
  it("retains accepted outcomes on unmount and invalidates a detached create review", async () => {
    const f = setup(mentionReview({ intent: { action: "dismiss_item" } }));
    let accept!: (
      value: Awaited<ReturnType<TimelineMentionResolutionPort["send"]>>,
    ) => void;
    f.send.mockReturnValueOnce(
      new Promise((resolve) => {
        accept = resolve;
      }),
    );
    act(() => f.result.current.startCreate());
    expect(f.result.current.createReview?.draft["host.display_name"]).toBe(
      f.review.subject.rawText,
    );
    expect(f.result.current.createReview?.draft["host.fqdn"]).toBe("");
    f.rerender({ ...f.input, presentationKey: "another-sheet" });
    expect(f.result.current.createReview).toBeNull();
    f.rerender(f.input);
    act(() => f.result.current.act({ action: "dismiss_item" }));
    await flush();
    f.unmount();
    accept({ kind: "accepted", receipt: mentionReceipt(f.review) });
    await flush();
    expect(f.owner.getSnapshot().entries[0]?.receipt).toEqual(
      mentionReceipt(f.review),
    );
    expect(
      f.input.focusContinuity.advanceViewportContinuity,
    ).not.toHaveBeenCalled();
  });
});

it("retries an uncertain Undo through a fresh presentation intent and the captured attempt", async () => {
  const f = disclosureFixture();
  f.owner.configureSourceReader(async () => f.row);
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  act(() => f.result.current.handleUndoAutoResolutionNotice(f.notice));
  await flush();
  const original = f.owner.getSnapshot().entries.at(-1);
  expect(original?.phase).toBe("uncertain");
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  expect(f.input.focusContinuity.clearViewportContinuity).toHaveBeenCalledWith(
    9,
  );
  let accept!: (
    value: Awaited<ReturnType<TimelineMentionResolutionPort["send"]>>,
  ) => void;
  f.send.mockReturnValueOnce(
    new Promise((resolve) => {
      accept = resolve;
    }),
  );
  const invoker = document.createElement("button");
  document.body.append(invoker);
  invoker.focus();
  act(() => {
    f.result.current.handleRetryUndoAutoResolutionNotice(
      f.notice,
      original?.key ?? -1,
    );
    f.result.current.handleRetryUndoAutoResolutionNotice(
      f.notice,
      original?.key ?? -1,
    );
  });
  expect(f.result.current.retryingUndoKey).toBe(original?.key);
  expect(f.owner.getSnapshot().entries.at(-1)?.phase).toBe("submitting");
  expect(f.send).toHaveBeenCalledTimes(2);
  expect(f.send.mock.calls[1]?.[0]).toBe(f.send.mock.calls[0]?.[0]);
  expect(
    f.input.focusContinuity.beginViewportContinuity,
  ).toHaveBeenLastCalledWith(
    { kind: "row-inspect", recordId: f.notice.rowRecordId },
    { requirements: ["row-projection"] },
  );
  accept({ kind: "accepted", receipt: mentionReceipt(f.review) });
  await flush();
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(0);
  expect(
    f.input.focusContinuity.advanceViewportContinuity,
  ).toHaveBeenCalledWith(10);
  invoker.remove();
  f.unmount();
});

it("retires presentation intent on same-row Inspector retarget without discarding an accepted Undo", async () => {
  const f = disclosureFixture();
  f.owner.configureSourceReader(async () => f.row);
  f.owner.registerReconciliation(async () => {
    throw new Error("Refresh unavailable");
  });
  let accept!: (
    value: Awaited<ReturnType<TimelineMentionResolutionPort["send"]>>,
  ) => void;
  f.send.mockReturnValueOnce(
    new Promise((resolve) => {
      accept = resolve;
    }),
  );
  act(() => f.result.current.handleUndoAutoResolutionNotice(f.notice));
  await flush();
  f.rerender({ ...f.input, selectedMentionRef: "another-item" });
  expect(f.input.focusContinuity.clearViewportContinuity).toHaveBeenCalledWith(
    9,
  );
  accept({ kind: "accepted", receipt: mentionReceipt(f.review) });
  await flush();
  expect(f.owner.getSnapshot().entries.at(-1)?.receipt).toBeTruthy();
  expect(f.owner.getSnapshot().entries.at(-1)?.refresh).toBe("required");
  expect(f.send).toHaveBeenCalledOnce();
  expect(
    f.input.focusContinuity.advanceViewportContinuity,
  ).not.toHaveBeenCalled();
  f.unmount();
});

it("coalesces a pending Review and fences its late completion after native editing", async () => {
  const f = disclosureFixture();
  navigateReview.mockClear();
  let resolveRead!: (row: typeof f.row) => void;
  const read = vi.fn((_id: string, _signal: AbortSignal) => {
    return new Promise<typeof f.row>((resolve) => {
      resolveRead = resolve;
    });
  });
  f.owner.configureSourceReader(read);
  act(() => {
    f.result.current.startDisclosureReview(f.notice, navigateReview);
    f.result.current.startDisclosureReview(f.notice, navigateReview);
  });
  expect(read).toHaveBeenCalledOnce();
  const editor = document.createElement("input");
  document.body.append(editor);
  act(() => editor.dispatchEvent(new Event("input", { bubbles: true })));
  expect(read.mock.calls[0]?.[1].aborted).toBe(true);
  resolveRead(f.row);
  await flush();
  expect(navigateReview).not.toHaveBeenCalled();
  expect(f.result.current.reviewFeedback).toBeNull();
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  editor.remove();
  f.unmount();
});

it("shows delayed Review progress and cancels on a same-row Inspector retarget", async () => {
  const f = disclosureFixture();
  let resolveRead!: (row: typeof f.row) => void;
  const read = vi.fn(
    () =>
      new Promise<typeof f.row>((resolve) => {
        resolveRead = resolve;
      }),
  );
  f.owner.configureSourceReader(read);
  act(() => f.result.current.startDisclosureReview(f.notice, navigateReview));
  expect(f.result.current.reviewFeedback).toBeNull();
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 150));
  });
  expect(f.result.current.reviewFeedback).toEqual({
    key: expect.any(String),
    phase: "pending",
  });
  f.rerender({ ...f.input, selectedMentionRef: "another-item" });
  expect(f.result.current.reviewFeedback).toBeNull();
  resolveRead(f.row);
  await flush();
  expect(navigateReview).not.toHaveBeenCalled();
  f.unmount();
});

it("rejects a same-row Inspector retarget while Review focus is pending", async () => {
  const f = disclosureFixture();
  f.owner.configureSourceReader(async () => f.row);
  f.rerender({ ...f.input, selectedMentionRef: null });
  let scope: DisclosureReviewNavigationScope | undefined;
  act(() =>
    f.result.current.startDisclosureReview(f.notice, (_row, _item, owned) => {
      scope = owned;
    }),
  );
  await flush();
  expect(scope).toBeDefined();
  f.rerender({ ...f.input, selectedMentionRef: f.notice.itemRef });
  expect(scope?.isCurrent()).toBe(true);
  f.rerender({ ...f.input, selectedMentionRef: "another-item" });
  expect(scope?.isCurrent()).toBe(false);
  act(() => scope?.settle(true));
  expect(f.result.current.reviewFeedback).toBeNull();
  f.unmount();
});

it("keeps a genuine Review failure local and retries without retaining the old error", async () => {
  const f = disclosureFixture();
  navigateReview.mockClear();
  let attempts = 0;
  f.owner.configureSourceReader(async () => {
    if (++attempts === 1) throw new Error("Transport unavailable");
    return f.row;
  });
  act(() => f.result.current.startDisclosureReview(f.notice, navigateReview));
  await flush();
  expect(f.result.current.reviewFeedback?.phase).toBe("failure");
  act(() => f.result.current.startDisclosureReview(f.notice, navigateReview));
  await flush();
  expect(attempts).toBe(2);
  expect(navigateReview).toHaveBeenCalledOnce();
  expect(f.result.current.reviewFeedback).toBeNull();
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  expect(f.send).not.toHaveBeenCalled();
  f.unmount();
});

it("refuses Inspector navigation when the committed source rejects a stale read", async () => {
  const f = disclosureFixture();
  const navigate = vi.fn();
  f.owner.configureSourceReader(async () => f.row);
  f.rerender({
    ...f.input,
    acceptDisclosureSource: () => ({
      accepted: false,
      row: f.row,
      stale: true,
    }),
  });
  act(() => f.result.current.startDisclosureReview(f.notice, navigate));
  await flush();
  expect(navigate).not.toHaveBeenCalled();
  expect(f.result.current.reviewFeedback?.phase).toBe("failure");
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  f.unmount();
});

it("lets only the latest Review navigate when two mentions share a source row", async () => {
  const f = disclosureFixture();
  navigateReview.mockClear();
  const firstItem = f.row.collectionValues.hostRefs[0];
  if (!firstItem) throw new Error("First mention required");
  const secondItem = {
    ...firstItem,
    entityMentionId: "second-mention",
    itemRef: "second-item",
    rawText: "second alias",
    displayText: "second alias",
  };
  const secondRow = {
    ...f.row,
    collectionValues: {
      ...f.row.collectionValues,
      hostRefs: [firstItem, secondItem],
    },
  };
  act(() =>
    f.owner.acceptAutoResolutions([f.row], [secondRow], {
      kind: "entry",
      operationId: "second-match",
      changeSetId: "second-change",
    }),
  );
  const secondNotice = f.owner
    .getDisclosureSnapshot()
    .find((notice) => notice.itemRef === secondItem.itemRef);
  if (!secondNotice) throw new Error("Second disclosure required");
  let rejectOld!: (error: Error) => void;
  let resolveNew!: (row: typeof secondRow) => void;
  const oldRead = new Promise<typeof secondRow>((_resolve, reject) => {
    rejectOld = reject;
  });
  const newRead = new Promise<typeof secondRow>((resolve) => {
    resolveNew = resolve;
  });
  const read = vi
    .fn()
    .mockReturnValueOnce(oldRead)
    .mockReturnValueOnce(newRead);
  f.owner.configureSourceReader(read);
  act(() => {
    f.result.current.startDisclosureReview(f.notice, navigateReview);
    f.result.current.startDisclosureReview(secondNotice, navigateReview);
  });
  expect(read).toHaveBeenCalledTimes(2);
  expect(read.mock.calls[0]?.[1].aborted).toBe(true);
  resolveNew(secondRow);
  await flush();
  expect(navigateReview).toHaveBeenCalledExactlyOnceWith(
    secondNotice.rowRecordId,
    secondNotice.itemRef,
    expect.any(Object),
  );
  rejectOld(new Error("Late failure"));
  await flush();
  expect(f.result.current.reviewFeedback).toBeNull();
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(2);
  f.unmount();
});

it("fences Review completion after sheet departure authority retirement and unmount", async () => {
  for (const departure of ["sheet", "authority", "unmount"] as const) {
    const f = disclosureFixture();
    navigateReview.mockClear();
    let resolveRead!: (row: typeof f.row) => void;
    const read = vi.fn((_id: string, _signal: AbortSignal) => {
      return new Promise<typeof f.row>((resolve) => {
        resolveRead = resolve;
      });
    });
    f.owner.configureSourceReader(read);
    act(() => f.result.current.startDisclosureReview(f.notice, navigateReview));
    if (departure === "sheet")
      f.rerender({ ...f.input, reviewSurfaceKey: "hosts:another-sheet" });
    else if (departure === "authority") act(() => f.owner.setAuthority(null));
    else f.unmount();
    await flush();
    expect(read.mock.calls[0]?.[1].aborted).toBe(true);
    resolveRead(f.row);
    await flush();
    expect(navigateReview).not.toHaveBeenCalled();
    if (departure !== "unmount") {
      expect(f.result.current.reviewFeedback).toBeNull();
      f.unmount();
    }
  }
});

it("Timeline disclosure actions read only their unavailable source and preserve accepted correction through failed refresh", async () => {
  const base = mentionReview();
  const f = setup(
    mentionReview({
      subject: {
        ...base.subject,
        state: "resolved",
        resolutionMethod: "auto_match",
        resolvedRecordId: "70000000-0000-4000-8000-000000000001",
      },
      intent: { action: "revert_to_unresolved" },
    }),
  );
  const row = f.rowsRef.current[0];
  if (!row) throw new Error("Source fixture required");
  act(() =>
    f.owner.acceptAutoResolutions(
      [{ ...row, collectionValues: { ...row.collectionValues, hostRefs: [] } }],
      [row],
      { kind: "entry", operationId: "match", changeSetId: "match-change" },
    ),
  );
  const notice = f.owner.getDisclosureSnapshot()[0];
  if (!notice) throw new Error("Disclosure required");
  f.rowsRef.current = [];
  const read = vi.fn(async (_recordId: string, _signal: AbortSignal) => row);
  f.owner.configureSourceReader(read);
  const navigate = vi.fn(
    (
      _recordId: string,
      _itemRef: string,
      scope: DisclosureReviewNavigationScope,
    ) => scope.settle(true),
  );
  act(() => f.result.current.startDisclosureReview(notice, navigate));
  await flush();
  expect(navigate).toHaveBeenCalledOnce();
  expect(read.mock.calls[0]?.[0]).toBe(notice.rowRecordId);
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  f.owner.configureSourceReader(async () => {
    throw new Error("Source unavailable");
  });
  act(() => f.result.current.handleUndoAutoResolutionNotice(notice));
  await flush();
  expect(f.send).not.toHaveBeenCalled();
  expect(f.owner.getSnapshot().entries.at(-1)?.phase).toBe(
    "preparation_failed",
  );
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(1);
  f.owner.configureSourceReader(read);
  f.owner.registerReconciliation(async () => {
    throw new Error("Refresh unavailable");
  });
  act(() => f.result.current.handleUndoAutoResolutionNotice(notice));
  await flush();
  expect(f.owner.getDisclosureSnapshot()).toHaveLength(0);
  expect(f.owner.getSnapshot().entries.at(-1)?.refresh).toBe("required");
  expect(
    f.input.focusContinuity.advanceViewportContinuity,
  ).toHaveBeenCalledTimes(1);
  expect(f.send).toHaveBeenCalledTimes(1);
  f.unmount();
});
