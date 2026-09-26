import {
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import type {
  AutoResolutionDisclosure,
  WorkbookTimelineMentionOperationOwner,
} from "../actions/WorkbookTimelineMentionOperationOwner";
import { disclosureReviewKey } from "../models/workbookMentionChips";
import { TimelineWorkbookNotices } from "./TimelineWorkbookNotices";

afterEach(cleanup);

const notice: AutoResolutionDisclosure = {
  identity: "source-bound-disclosure",
  rowRecordId: "source-row",
  sourceRowVersion: 2,
  fieldKey: "timeline.host_refs",
  entityMentionId: "mention-one",
  itemRef: "item-one",
  mentionRowVersion: 2,
  entityType: "host",
  rawText: "raw alias",
  resolvedRecordId: "canonical-host",
  matchedAliasText: "raw alias",
  operation: { kind: "entry", operationId: "accepted", changeSetId: null },
  acceptedCount: 1,
};
const disclosures = [notice];
const actions = { entries: [] };
const owner = {
  subscribe: () => () => {},
  getDisclosureSnapshot: () => disclosures,
  getActionSnapshot: () => actions,
  updateDisclosureLabels: vi.fn(),
  canSubmit: () => false,
} as unknown as WorkbookTimelineMentionOperationOwner;

it("keeps Review keyboard accessible while source-local progress is busy", () => {
  const review = vi.fn();
  render(
    <TimelineWorkbookNotices
      owner={owner}
      density="default"
      entityIndex={{}}
      reviewFeedback={{ key: disclosureReviewKey(notice), phase: "pending" }}
      onReviewAutoResolution={review}
      onUndoAutoResolution={vi.fn()}
    />,
  );
  const region = screen.getByRole("complementary", {
    name: "Auto-resolution disclosures",
  });
  const button = within(region).getByRole("button", { name: "Review" });
  expect((button as HTMLButtonElement).disabled).toBe(false);
  expect(button.getAttribute("aria-busy")).toBe("true");
  expect(within(region).getByRole("status").textContent).toContain(
    "Opening source",
  );
  button.focus();
  expect(document.activeElement).toBe(button);
  fireEvent.click(button);
  expect(review).toHaveBeenCalledOnce();
  expect(review).toHaveBeenCalledWith(notice);
});

it("shows a retryable Review read failure beside its disclosure", () => {
  const review = vi.fn();
  render(
    <TimelineWorkbookNotices
      owner={owner}
      density="default"
      entityIndex={{}}
      reviewFeedback={{
        key: disclosureReviewKey(notice),
        phase: "failure",
        message: "Could not open the source. Review again to retry.",
      }}
      onReviewAutoResolution={review}
      onUndoAutoResolution={vi.fn()}
    />,
  );
  expect(screen.getByRole("alert").textContent).toContain(
    "Review again to retry",
  );
  fireEvent.click(screen.getByRole("button", { name: "Review" }));
  expect(review).toHaveBeenCalledWith(notice);
  expect(
    screen.getByRole("complementary", { name: "Auto-resolution disclosures" }),
  ).toBeTruthy();
});
