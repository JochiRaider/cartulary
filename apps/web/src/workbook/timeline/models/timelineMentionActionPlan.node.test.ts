import { describe, expect, it } from "vitest";
import {
  mentionInspector,
  mentionReview,
  mentionWorkbookRow,
} from "../../../testing/timelineMentionTestSupport";
import { timelineMentionSubject } from "./timelineMentionActionPlan";

describe("Timeline Mention action plan", () => {
  it("binds explicit mention identity separately from opaque selector source and target", () => {
    const review = mentionReview(),
      mention = mentionInspector(review),
      row = mentionWorkbookRow(review);
    expect(
      timelineMentionSubject(mention, row, review.subject.incidentId),
    ).toEqual(review.subject);
    expect(mention.anchor.entityMentionId).toBe(review.subject.mentionId);
    expect(mention.itemRef).not.toContain(review.subject.mentionId);
  });
  it("refuses missing public identity and source versions without parsing selectors", () => {
    const review = mentionReview(),
      mention = mentionInspector(review),
      row = mentionWorkbookRow(review);
    expect(
      timelineMentionSubject(
        {
          ...mention,
          entityMentionId: null,
          itemRef: `entity_mention:${review.subject.mentionId}`,
        },
        row,
        review.subject.incidentId,
      ),
    ).toBeNull();
    expect(
      timelineMentionSubject(
        mention,
        { ...row, rowVersion: null },
        review.subject.incidentId,
      ),
    ).toBeNull();
    expect(
      timelineMentionSubject(
        mention,
        { ...row, recordId: "another-source" },
        review.subject.incidentId,
      ),
    ).toBeNull();
  });
});
