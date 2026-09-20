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
import type { TimelineMentionResolutionPort } from "./ports/TimelineMentionPort";

function setup(review: MentionReview = mentionReview()) {
  let sequence = 0;
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
    earlierSaves: { current: Promise.resolve() },
    selectedMention: mentionInspector(review),
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
    restoreActionFocus: vi.fn(),
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
    expect(f.input.restoreActionFocus).toHaveBeenCalledWith(
      review.subject.sourceRecordId,
    );
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
    expect(f.input.restoreActionFocus).not.toHaveBeenCalled();
    release();
    await flush();
    expect(f.input.restoreActionFocus).toHaveBeenCalledExactlyOnceWith(
      f.review.subject.sourceRecordId,
    );
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
    expect(f.input.restoreActionFocus).not.toHaveBeenCalled();
  });
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
  await act(async () => {
    expect(await f.result.current.prepareDisclosureReview(notice)).toBe(row);
  });
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
  expect(f.input.restoreActionFocus).toHaveBeenCalledTimes(1);
  expect(f.send).toHaveBeenCalledTimes(1);
  f.unmount();
});
