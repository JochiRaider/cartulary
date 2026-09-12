import { expect, it, vi } from "vitest";
import {
  mentionReceipt,
  mentionReview,
  mentionWorkbookRow,
} from "../../../testing/timelineMentionTestSupport";
import { reconcileTimelineMentionReceipt } from "./reconcileTimelineMentionReceipt";
import { WorkbookTimelineMentionOperationOwner } from "./WorkbookTimelineMentionOperationOwner";

function fixture() {
  const review = mentionReview(),
    receipt = mentionReceipt(review);
  const owner = new WorkbookTimelineMentionOperationOwner(
    review.subject.incidentId,
    { create: () => "unused" },
    { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() },
  );
  owner.setAuthority(review.authority);
  const subject = {
    ...review.subject,
    state: "resolved" as const,
    resolutionMethod: "explicit_resolve_route",
    resolvedRecordId: receipt.entity_mention.resolved_record_id,
    mentionRowVersion: 3,
    sourceRowVersion: 5,
  };
  const row = mentionWorkbookRow({ ...review, subject }).rawRow;
  if (!row) throw new Error("Row fixture required");
  return {
    review,
    receipt,
    owner,
    subject,
    row,
    scope: { signal: new AbortController().signal, isCurrent: () => true },
  };
}
it("Mention reconciliation accepts cleared dismissal and keeps newer remote correction over an older receipt", async () => {
  const f = fixture();
  const dismissed = {
    ...f.subject,
    state: "dismissed" as const,
    resolvedRecordId: null,
    resolutionMethod: null,
  };
  const receipt = mentionReceipt({
    ...f.review,
    intent: { action: "dismiss_item" },
  });
  const empty = mentionWorkbookRow({ ...f.review, subject: dismissed }).rawRow;
  if (!empty) throw new Error("Row required");
  await reconcileTimelineMentionReceipt(
    f.owner,
    async () => empty,
    receipt,
    f.scope,
  );
  const newer = {
    ...f.subject,
    mentionRowVersion: 8,
    sourceRowVersion: 12,
    resolvedRecordId: "70000000-0000-4000-8000-000000000008",
  };
  const row = mentionWorkbookRow({ ...f.review, subject: newer }).rawRow;
  if (!row) throw new Error("Row required");
  await reconcileTimelineMentionReceipt(
    f.owner,
    async () => row,
    f.receipt,
    f.scope,
  );
  expect(f.owner.latestMention(newer.mentionId)).toEqual(newer);
  await expect(
    reconcileTimelineMentionReceipt(
      f.owner,
      async () => f.row,
      f.receipt,
      f.scope,
    ),
  ).rejects.toThrow("newer source");
  expect(f.owner.latestMention(newer.mentionId)).toEqual(newer);
});
it("Mention reconciliation rejects stale sources malformed same-version projections and detached reads", async () => {
  for (const mismatch of [
    "source",
    "text",
    "target",
    "mention_version",
    "detached",
  ] as const) {
    const f = fixture();
    const subject = {
      ...f.subject,
      ...(mismatch === "source" ? { sourceRowVersion: 4 } : {}),
      ...(mismatch === "text" ? { rawText: "changed" } : {}),
      ...(mismatch === "target" ? { resolvedRecordId: "wrong" } : {}),
      ...(mismatch === "mention_version"
        ? { mentionRowVersion: 2, sourceRowVersion: 10 }
        : {}),
    };
    const row = mentionWorkbookRow({ ...f.review, subject }).rawRow;
    if (!row) throw new Error("Row required");
    await expect(
      reconcileTimelineMentionReceipt(f.owner, async () => row, f.receipt, {
        ...f.scope,
        isCurrent: () => mismatch !== "detached",
      }),
    ).rejects.toThrow();
    expect(f.owner.latestMention(f.subject.mentionId)).toBeNull();
  }
});
