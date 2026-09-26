import type { GridFocusResult, GridHandle } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
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
import { useWorkbookInspectorCoordinator } from "../inspector/useWorkbookInspectorCoordinator";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { MentionReview } from "./actions/timelineMentionOperationModel";
import { WorkbookTimelineMentionOperationOwner } from "./actions/WorkbookTimelineMentionOperationOwner";
import { createTimelineMentionEntityCreationAdapter } from "./adapters/createTimelineMentionEntityCreationAdapter";
import { createTimelineMentionResolutionAdapter } from "./adapters/createTimelineMentionResolutionAdapter";
import { commitTimelineProjection } from "./adapters/timelineProjectionCommitAdapter";
import { createTimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import { useTimelineMentionActions } from "./hooks/useTimelineMentionActions";
import {
  type TimelineViewportContinuityRequest,
  useTimelineViewportContinuityController,
} from "./hooks/useTimelineViewportContinuityController";
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
  const rowsRef = { current: [mentionWorkbookRow(review)] };
  owner.registerReconciliation(async (receipt) => {
    rowsRef.current = [
      {
        ...mentionWorkbookRow(review),
        rowVersion: receipt.source_record.row_version,
      },
    ];
    await owner.refreshPresentation(receipt, receipt.source_record.row_version);
  });
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
    inspectorReviewGeneration: 0,
    inspectorAttachmentGeneration: 0,
    refreshProjection: async () => {},
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
      requireViewportContinuitySourceRecord: vi.fn(),
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

function completionFixture(
  review = mentionReview({ intent: { action: "dismiss_item" } }),
) {
  const owner = new WorkbookTimelineMentionOperationOwner(
    review.subject.incidentId,
    { create: () => "completion-attempt" },
    { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() },
  );
  const receipt = mentionReceipt(review);
  const send = vi.fn<TimelineMentionResolutionPort["send"]>(async () => ({
    kind: "accepted",
    receipt,
  }));
  owner.configure({
    ...createTimelineMentionResolutionAdapter({ apiBase: undefined }),
    send,
  });
  owner.setAuthority(review.authority);
  owner.registerReconciliation(async (accepted) => {
    await owner.refreshPresentation(
      accepted,
      accepted.source_record.row_version,
    );
  });
  let release!: () => void;
  let reject!: (error: Error) => void;
  const refreshProjection = vi.fn(
    () =>
      new Promise<void>((resolve, fail) => {
        release = resolve;
        reject = fail;
      }),
  );
  const grid = document.createElement("div");
  grid.tabIndex = -1;
  grid.scrollTop = 120;
  grid.scrollLeft = 40;
  Object.defineProperty(grid, "getBoundingClientRect", {
    value: () => new DOMRect(0, 0, 600, 400),
  });
  const cell = document.createElement("div");
  cell.tabIndex = -1;
  cell.role = "gridcell";
  cell.dataset.gridRecordId = review.subject.sourceRecordId;
  cell.dataset.gridFieldKey = "timeline.activity_synopsis_text";
  Object.defineProperty(cell, "getBoundingClientRect", {
    value: () => new DOMRect(80, 80, 200, 30),
  });
  grid.append(cell);
  const invoker = document.createElement("button");
  document.body.append(grid, invoker);
  invoker.focus();
  const waiting = new Set<() => void>();
  const requestFocus = vi.fn<GridHandle["requestFocus"]>(
    (target, options) =>
      new Promise<GridFocusResult>((resolve) => {
        const finish = (result: GridFocusResult) => {
          waiting.delete(ready);
          options?.signal?.removeEventListener("abort", abort);
          resolve(result);
        };
        const abort = () => finish("cancelled");
        const ready = () => {
          if (options?.signal?.aborted) return abort();
          const element = target.kind === "root" ? grid : cell;
          if (!element.isConnected) return;
          element.focus({ preventScroll: true });
          finish("focused");
        };
        options?.signal?.addEventListener("abort", abort, { once: true });
        waiting.add(ready);
        ready();
      }),
  );
  const gridHandleRef = {
    current: {
      requestFocus,
      getScrollElement: () => grid,
      getAnchorRect: () => cell.getBoundingClientRect(),
    } as unknown as GridHandle,
  };
  const gridShellRef = { current: grid };
  const viewportContinuityTokenRef = { current: 1 };
  const registry = createTimelineEditorDraftRegistry();
  const scope = {
    getSnapshot: () => ({ key: "source", readable: true }),
    subscribe: () => () => {},
  };
  const rowsRef = { current: [mentionWorkbookRow(review)] };
  const candidatePort = {
    page: vi.fn(async () => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          {
            recordId: review.intent.resolvedRecordId ?? "target",
            rowVersion: 1,
            displayText: "Target",
            entityType: review.subject.entityType,
          },
        ],
        nextCursor: null,
        hasMore: false,
      },
    })),
  };
  let props = {
    rowVersion: review.subject.sourceRowVersion,
    lifecycleKey: "source",
    recordId: review.subject.sourceRecordId,
    mentionRef: review.subject.itemRef as string | null,
  };
  const hook = renderHook(
    ({ rowVersion, lifecycleKey, recordId, mentionRef }) => {
      rowsRef.current = [{ ...mentionWorkbookRow(review), rowVersion }];
      const inspector = useWorkbookInspectorCoordinator({
        config: requireViewContract(timelineViewSchemaId).inspectorConfig,
        lifecycleKey,
        subject: {
          kind: "live",
          recordId,
          rowVersion,
          viewSchemaId: timelineViewSchemaId,
          label: "Source",
          surfaceLabel: "Timeline",
        },
        actionPorts: { resetOwnerState: () => {}, restoreFocus: () => {} },
      });
      const [request, setRequest] =
        useState<TimelineViewportContinuityRequest | null>(null);
      const continuity = useTimelineViewportContinuityController({
        gridHandleRef,
        gridShellRef,
        scope,
        editorDraftRegistry: registry,
        viewportContinuityTokenRef,
        viewportContinuityRequest: request,
        setViewportContinuityRequest: setRequest,
      });
      const mentions = useTimelineMentionActions({
        owner,
        rowsRef,
        candidatePort,
        earlierSaves: { current: Promise.resolve() },
        selectedMention: mentionInspector(review),
        selectedMentionRef: mentionRef,
        selectedRowId: recordId,
        inspectorReviewGeneration: inspector.snapshot.reviewGeneration,
        inspectorAttachmentGeneration: inspector.snapshot.attachmentGeneration,
        reviewSurfaceKey: lifecycleKey,
        presentationKey: lifecycleKey,
        presentationActive: inspector.snapshot.phase !== "closed",
        selectedTargetId: review.intent.resolvedRecordId ?? "",
        setSelectedTargetId: () => {},
        waitForCommittedRecordIdle: async () => ({
          row: rowsRef.current[0] ?? null,
          rowVersion,
        }),
        setInspectorMessage: () => {},
        refreshProjection,
        focusContinuity: continuity.commands,
      });
      return { mentions, inspector, request };
    },
    { initialProps: props },
  );
  act(() => hook.result.current.inspector.commands.open());
  const update = (next: Partial<typeof props>) => {
    props = { ...props, ...next };
    act(() => hook.rerender(props));
  };
  const project = (rowVersion: number) =>
    act(() => {
      invoker.remove();
      cell.dataset.gridRowVersion = String(rowVersion);
      commitTimelineProjection(() => update({ rowVersion }), true);
    });
  return {
    ...hook,
    review,
    receipt,
    owner,
    send,
    grid,
    cell,
    invoker,
    requestFocus,
    refreshProjection,
    project,
    update,
    mountTarget: () =>
      act(() => {
        grid.append(cell);
        for (const ready of waiting) ready();
      }),
    release: () => release(),
    reject: () => reject(new Error("Refresh unavailable")),
    dispose: () => {
      hook.unmount();
      grid.remove();
      invoker.remove();
    },
  };
}

it("preserves mention completion through its own same-row inspector refresh", async () => {
  const f = completionFixture();
  try {
    act(() => f.result.current.mentions.act(f.review.intent));
    await flush();
    expect(f.owner.getSnapshot().entries[0]?.receipt).toEqual(f.receipt);
    expect(f.refreshProjection).toHaveBeenCalledOnce();
    f.project(f.receipt.source_record.row_version);
    expect(document.activeElement).toBe(document.body);
    f.release();
    await flush();
    expect(document.activeElement).toBe(f.cell);
    expect([f.grid.scrollTop, f.grid.scrollLeft]).toEqual([120, 40]);
    expect(f.send).toHaveBeenCalledOnce();
  } finally {
    f.dispose();
  }
});

it("waits for the response source version across resolution dismissal and reversion", async () => {
  const base = mentionReview();
  for (const review of [
    base,
    mentionReview({ intent: { action: "dismiss_item" } }),
    mentionReview({
      subject: {
        ...base.subject,
        state: "resolved",
        resolvedRecordId: base.intent.resolvedRecordId ?? null,
        resolutionMethod: "explicit_resolve_route",
      },
      intent: { action: "revert_to_unresolved" },
    }),
  ]) {
    for (const extraVersion of [0, 1]) {
      const f = completionFixture(review);
      try {
        await flush();
        act(() => f.result.current.mentions.act(review.intent));
        await flush();
        expect(f.refreshProjection).toHaveBeenCalledWith(
          {
            recordId: review.subject.sourceRecordId,
            minimumRowVersion: f.receipt.source_record.row_version,
          },
          1,
        );
        f.project(f.receipt.source_record.row_version - 1);
        await flush();
        expect(f.requestFocus).not.toHaveBeenCalled();
        expect(f.result.current.request).not.toBeNull();
        f.project(f.receipt.source_record.row_version + extraVersion);
        f.release();
        await flush();
        expect(document.activeElement).toBe(f.cell);
        expect(f.send).toHaveBeenCalledOnce();
      } finally {
        f.dispose();
      }
    }
  }
});

it("accepts collaboration projection before the receipt without restoring early", async () => {
  const f = completionFixture();
  try {
    let accept!: (
      result: Awaited<ReturnType<TimelineMentionResolutionPort["send"]>>,
    ) => void;
    f.send.mockReturnValueOnce(
      new Promise((resolve) => {
        accept = resolve;
      }),
    );
    act(() => f.result.current.mentions.act(f.review.intent));
    await flush();
    f.project(f.receipt.source_record.row_version);
    expect(f.requestFocus).not.toHaveBeenCalled();
    accept({ kind: "accepted", receipt: f.receipt });
    await flush();
    f.release();
    await flush();
    expect(document.activeElement).toBe(f.cell);
    expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("complete");
  } finally {
    f.dispose();
  }
});

it("restores a terminal refresh fallback without replaying an accepted mention", async () => {
  const f = completionFixture();
  try {
    act(() => f.result.current.mentions.act(f.review.intent));
    await flush();
    f.invoker.remove();
    f.reject();
    await flush();
    expect(document.activeElement).toBe(f.cell);
    expect(f.owner.getSnapshot().entries[0]).toMatchObject({
      receipt: f.receipt,
      refresh: "required",
    });
    const key = f.owner.getSnapshot().entries[0]?.key ?? -1;
    void f.owner.refresh(key);
    await flush();
    const newer = document.createElement("input");
    document.body.append(newer);
    act(() => newer.focus());
    f.project(f.receipt.source_record.row_version);
    f.release();
    await flush();
    expect(document.activeElement).toBe(newer);
    expect(f.send).toHaveBeenCalledOnce();
    newer.remove();
  } finally {
    f.dispose();
  }
});

it("never revives mention completion after interaction attachment or authority cancellation", async () => {
  for (const cause of [
    "pointerdown",
    "keydown",
    "wheel",
    "input",
    "compositionstart",
    "focus",
    "close",
    "surface",
    "row",
    "mention",
    "authority",
    "account",
    "unmount",
  ]) {
    for (const failed of [false, true]) {
      const f = completionFixture();
      const newer = document.createElement("input");
      document.body.append(newer);
      try {
        act(() => f.result.current.mentions.act(f.review.intent));
        await flush();
        act(() => {
          if (cause === "close") {
            f.result.current.inspector.commands.close();
          } else if (cause === "surface") f.update({ lifecycleKey: "other" });
          else if (cause === "row") f.update({ recordId: "other" });
          else if (cause === "mention") f.update({ mentionRef: "other" });
          else if (cause === "authority") f.owner.setAuthority(null);
          else if (cause === "account") f.owner.retire();
          else if (cause === "unmount") f.unmount();
          else if (cause === "focus") newer.focus();
          else newer.dispatchEvent(new Event(cause, { bubbles: true }));
        });
        // Returning to the old destination cannot revive its old request.
        if (cause === "close")
          act(() => f.result.current.inspector.commands.open());
        if (cause === "surface") f.update({ lifecycleKey: "source" });
        if (cause === "row")
          f.update({ recordId: f.review.subject.sourceRecordId });
        if (cause === "mention")
          f.update({ mentionRef: f.review.subject.itemRef });
        if (cause !== "unmount") f.project(f.receipt.source_record.row_version);
        f.grid.scrollTop = 250;
        if (failed) f.reject();
        else f.release();
        await flush();
        expect(
          f.requestFocus,
          `${cause}, failed=${failed}`,
        ).not.toHaveBeenCalled();
        expect(f.grid.scrollTop).toBe(250);
        expect(f.send).toHaveBeenCalledOnce();
      } finally {
        newer.remove();
        f.dispose();
      }
    }
  }
});

it("keeps cancellation ownership while an accepted mention awaits virtualized focus", async () => {
  for (const close of [false, true]) {
    const f = completionFixture();
    try {
      act(() => f.result.current.mentions.act(f.review.intent));
      await flush();
      f.cell.remove();
      f.project(f.receipt.source_record.row_version);
      f.release();
      await flush();
      expect(f.requestFocus).toHaveBeenCalledOnce();
      expect(document.activeElement).toBe(document.body);
      if (close) act(() => f.result.current.inspector.commands.close());
      f.mountTarget();
      await flush();
      expect(document.activeElement).toBe(close ? document.body : f.cell);
    } finally {
      f.dispose();
    }
  }
});

it("does not lend the current mention focus token to another receipt on the same source", async () => {
  const f = completionFixture();
  try {
    let accept!: (
      result: Awaited<ReturnType<TimelineMentionResolutionPort["send"]>>,
    ) => void;
    f.send.mockReturnValueOnce(
      new Promise((resolve) => {
        accept = resolve;
      }),
    );
    act(() => f.result.current.mentions.act(f.review.intent));
    await flush();
    const oldReceipt = {
      ...f.receipt,
      source_record: {
        ...f.receipt.source_record,
        row_version: f.review.subject.sourceRowVersion,
      },
    };
    const olderRefresh = f.owner.refreshPresentation(
      oldReceipt,
      oldReceipt.source_record.row_version,
    );
    expect(f.refreshProjection).toHaveBeenLastCalledWith(
      {
        recordId: f.review.subject.sourceRecordId,
        minimumRowVersion: f.review.subject.sourceRowVersion,
      },
      undefined,
    );
    f.release();
    await olderRefresh;
    expect(f.requestFocus).not.toHaveBeenCalled();
    accept({ kind: "accepted", receipt: f.receipt });
    await flush();
    expect(f.refreshProjection).toHaveBeenLastCalledWith(
      {
        recordId: f.review.subject.sourceRecordId,
        minimumRowVersion: f.receipt.source_record.row_version,
      },
      1,
    );
    f.project(f.receipt.source_record.row_version);
    f.release();
    await flush();
    expect(document.activeElement).toBe(f.cell);
  } finally {
    f.dispose();
  }
});

it("waits for source rendering before settling a failed Entity creation refresh", async () => {
  const f = setup();
  try {
    let rejectCreation!: (error: Error) => void;
    let releaseSource!: () => void;
    f.owner.registerCreationReconciliation(
      () =>
        new Promise<void>((_resolve, reject) => {
          rejectCreation = reject;
        }),
    );
    f.owner.registerReconciliation(async (receipt) => {
      await new Promise<void>((resolve) => {
        releaseSource = resolve;
      });
      f.rowsRef.current = [
        {
          ...mentionWorkbookRow(f.review),
          rowVersion: receipt.source_record.row_version,
        },
      ];
      await f.owner.refreshPresentation(
        receipt,
        receipt.source_record.row_version,
      );
    });
    act(() => f.result.current.startCreate());
    const review = f.result.current.createReview;
    if (!review) throw new Error("Creation review required");
    f.owner.configureCreation({
      ...createTimelineMentionEntityCreationAdapter({ apiBase: undefined }),
      send: async () => ({
        kind: "accepted",
        receipt: mentionCreationReceipt(review),
      }),
    });
    f.send.mockImplementation(async (attempt) => ({
      kind: "accepted",
      receipt: mentionReceipt(attempt.review),
    }));
    act(() => f.result.current.submitCreate());
    await flush();
    rejectCreation(new Error("Entity refresh unavailable"));
    await flush();
    expect(f.owner.getSnapshot().creations[0]?.refresh).toBe("required");
    expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("refreshing");
    expect(
      f.input.focusContinuity.advanceViewportContinuity,
    ).not.toHaveBeenCalled();
    releaseSource();
    await flush();
    expect(
      f.input.focusContinuity.settleViewportContinuityFollowUp,
    ).toHaveBeenCalledExactlyOnceWith(9, "row-projection", "terminal");
    expect(
      f.input.focusContinuity.advanceViewportContinuity,
    ).toHaveBeenCalledExactlyOnceWith(9);
    expect(f.send).toHaveBeenCalledOnce();
  } finally {
    f.unmount();
  }
});
