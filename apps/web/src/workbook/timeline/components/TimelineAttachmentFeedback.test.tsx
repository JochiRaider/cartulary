import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import { createRef } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type {
  TimelineFileAdmissionNotice,
  TimelineFileSnapshot,
  WorkbookTimelineFileOwner,
} from "../../features/evidence/WorkbookTimelineFileOwner";
import {
  TimelineAttachmentDetails,
  TimelineAttachmentFeedback,
} from "./TimelineAttachmentFeedback";

const file = (
  workId: string,
  feedbackKind: TimelineFileSnapshot["feedbackKind"] = "needs_attention",
): TimelineFileSnapshot => ({
  workId,
  outcomeIdentity: `${workId}:1`,
  feedbackKind,
  attention:
    feedbackKind === "completed"
      ? null
      : {
          workId,
          outcomeIdentity: `${workId}:1`,
          category: "uncertain",
          label: "Recover the original upload.",
        },
  key: "original-source",
  recordId: "original-record",
  filename: "same-file.txt",
  sourceLabel: "Original source",
  message:
    feedbackKind === "completed"
      ? "Evidence attached."
      : "Recover the original upload.",
  reviewText: null,
  busy: false,
  needsReview: false,
  evidenceAccepted: false,
  attached: feedbackKind === "completed",
  canResume: feedbackKind !== "completed",
  canFreshSlot: false,
  canNewId: false,
  canDiscard: true,
  refreshRequired: false,
});
function fixture(initial: readonly TimelineFileSnapshot[] = []) {
  let files = initial;
  let admission: TimelineFileAdmissionNotice | null = null;
  const listeners = new Set<() => void>();
  const actions = {
    confirmReview: vi.fn(),
    review: vi.fn(),
    resume: vi.fn(),
    freshSlot: vi.fn(),
    newRequestId: vi.fn(),
    discard: vi.fn(),
    refresh: vi.fn(),
  };
  const owner = {
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    getSnapshot: () => files,
    getAdmissionNotice: () => admission,
    ...actions,
  } as unknown as WorkbookTimelineFileOwner;
  return {
    owner,
    actions,
    publish(next: readonly TimelineFileSnapshot[], notice = admission) {
      act(() => {
        files = next;
        admission = notice;
        for (const listener of listeners) listener();
      });
    },
  };
}
const trigger = () => screen.getByRole("button", { name: /^Attachments:/ });
const region = () =>
  screen.getByRole("region", { name: "Timeline attachments" });
afterEach(cleanup);

describe("Timeline attachment feedback", () => {
  it("keeps completed history collapsed and excludes it from unfinished counts without discarding receipts", () => {
    const f = fixture([
      file("one", "completed"),
      file("two", "completed"),
      file("stopped"),
    ]);
    render(
      <TimelineAttachmentFeedback owner={f.owner} fallbackRef={createRef()} />,
    );
    expect(trigger().textContent).toBe(
      "Attachments: 1 need attention, 0 in progress, 2 completed",
    );
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
    expect(screen.queryByRole("group")).toBeNull();
    fireEvent.click(trigger());
    expect(screen.getAllByRole("button", { name: "Resume" })).toHaveLength(1);
    const completed = screen
      .getByText("Completed attachments (2)")
      .closest("details");
    expect(completed?.open).toBe(false);
    if (!completed) throw new Error("Missing history");
    completed.open = true;
    expect(
      within(completed).queryByRole("button", {
        name: "Discard retained file work",
      }),
    ).toBeNull();
    fireEvent.click(trigger());
    for (const action of Object.values(f.actions))
      expect(action).not.toHaveBeenCalled();
    expect(f.owner.getSnapshot()).toHaveLength(3);
  });

  it("keeps stage-specific actions bound to the original source", () => {
    const entry = {
      ...file("review"),
      needsReview: true,
      reviewText: "Reviewed version 2",
      canFreshSlot: true,
      canNewId: true,
      refreshRequired: true,
    };
    const f = fixture([entry]);
    render(
      <TimelineAttachmentFeedback owner={f.owner} fallbackRef={createRef()} />,
    );
    fireEvent.click(trigger());
    for (const [label, action] of [
      ["Review original source", f.actions.review],
      ["Use reviewed source", f.actions.confirmReview],
      ["Resume", f.actions.resume],
      ["Start fresh upload", f.actions.freshSlot],
      ["Use new request", f.actions.newRequestId],
      ["Refresh", f.actions.refresh],
      ["Discard retained file work", f.actions.discard],
    ] as const) {
      fireEvent.click(within(region()).getByRole("button", { name: label }));
      expect(action).toHaveBeenCalledExactlyOnceWith("original-source");
    }
  });

  it("preserves focus and disclosure state through background outcomes and returns lost controls locally", () => {
    const f = fixture([file("work")]);
    const fallback = createRef<HTMLDivElement>();
    render(
      <div ref={fallback} tabIndex={-1}>
        <input aria-label="Grid editor" />
        <TimelineAttachmentFeedback owner={f.owner} fallbackRef={fallback} />
      </div>,
    );
    const editor = screen.getByRole("textbox");
    editor.focus();
    f.publish([
      {
        ...file("work"),
        outcomeIdentity: "work:2",
        message: "Stopped work remains retained.",
      },
    ]);
    expect(document.activeElement).toBe(editor);
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
    fireEvent.click(trigger());
    const resume = within(region()).getByRole("button", {
      name: "Resume",
    });
    resume.focus();
    f.publish([
      {
        ...file("work", "in_progress"),
        busy: true,
        canResume: false,
        outcomeIdentity: "work:3",
      },
    ]);
    expect(document.activeElement?.getAttribute("aria-label")).toBe(
      "Attachment detail: same-file.txt",
    );
    f.publish([{ ...file("work", "completed"), outcomeIdentity: "work:4" }]);
    expect(document.activeElement).toBe(trigger());
    expect(trigger().getAttribute("aria-expanded")).toBe("true");
    fireEvent.keyDown(trigger(), { key: "Escape" });
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
    expect(document.activeElement).toBe(trigger());
  });

  it("falls back to the work area only when the last focused work disappears", () => {
    const f = fixture([file("work")]);
    const fallback = createRef<HTMLDivElement>();
    render(
      <div ref={fallback} tabIndex={-1}>
        <TimelineAttachmentFeedback owner={f.owner} fallbackRef={fallback} />
      </div>,
    );
    fireEvent.click(trigger());
    within(region())
      .getByRole("button", { name: "Discard retained file work" })
      .focus();
    f.publish([]);
    expect(document.activeElement).toBe(fallback.current);
  });

  it("announces new outcomes once including identical admission rejections without replaying retained outcomes on mount", () => {
    const f = fixture([file("work")]);
    const mounted = render(
      <TimelineAttachmentFeedback owner={f.owner} fallbackRef={createRef()} />,
    );
    expect(screen.getByRole("alert").textContent).toBe("");
    const failed = { ...file("work"), outcomeIdentity: "work:2" };
    f.publish([failed]);
    const first = screen.getByRole("alert").firstElementChild;
    expect(first?.textContent).toContain("Recover the original upload.");
    f.publish([{ ...failed }]);
    expect(screen.getByRole("alert").firstElementChild).toBe(first);
    fireEvent.click(trigger());
    expect(screen.getByRole("alert").firstElementChild).toBe(first);
    f.publish([failed], { identity: 1, message: "Choose one file at a time." });
    const rejection = screen.getByRole("alert").firstElementChild;
    expect(
      within(region()).getByText("Choose one file at a time."),
    ).toBeTruthy();
    f.publish([failed], { identity: 2, message: "Choose one file at a time." });
    expect(screen.getByRole("alert").firstElementChild).not.toBe(rejection);
    f.publish([{ ...failed, outcomeIdentity: "work:3" }], {
      identity: 3,
      message: "Choose one file at a time.",
    });
    expect(screen.getByRole("alert").textContent).toContain(
      "Choose one file at a time. same-file.txt: Recover the original upload.",
    );
    mounted.unmount();
    render(
      <TimelineAttachmentFeedback owner={f.owner} fallbackRef={createRef()} />,
    );
    expect(screen.getByRole("alert").textContent).toBe("");
  });

  it("shows new work on the same row and updates the feedback leaf without rebuilding its parent", () => {
    const f = fixture([file("old", "completed")]);
    let renders = 0;
    function Parent({ accepted = "Original accepted row" }) {
      renders++;
      return (
        <>
          <p>{accepted}</p>
          <TimelineAttachmentFeedback
            owner={f.owner}
            fallbackRef={createRef()}
          />
        </>
      );
    }
    const mounted = render(<Parent />);
    const before = renders;
    f.publish([file("new", "in_progress")]);
    expect(renders).toBe(before);
    expect(trigger().textContent).toBe(
      "Attachments: 0 need attention, 1 in progress, 0 completed",
    );
    fireEvent.click(trigger());
    expect(
      screen
        .getByRole("group")
        .closest("[data-attachment-work-id]")
        ?.getAttribute("data-attachment-work-id"),
    ).toBe("new");
    mounted.rerender(<Parent accepted="Updated accepted row" />);
    expect(screen.getByText("Updated accepted row")).toBeTruthy();
    expect(renders).toBe(before + 1);
  });

  it("keeps Inspector recovery source-specific and completed details collapsed with no duplicate live outcome", () => {
    const f = fixture();
    render(
      <TimelineAttachmentDetails
        owner={f.owner}
        files={[
          file("unfinished"),
          { ...file("complete", "completed"), filename: "completed.txt" },
        ]}
        presentation="inspector"
        fallbackRef={createRef()}
      />,
    );
    const recovery = screen.getByRole("group", {
      name: "Inspector file recovery: same-file.txt",
    });
    expect(within(recovery).getByRole("status").getAttribute("aria-live")).toBe(
      "off",
    );
    expect(
      recovery
        .closest("[data-evidence-work-id]")
        ?.getAttribute("data-evidence-work-id"),
    ).toBe("unfinished");
    expect(
      screen.getByText("Completed attachments (1)").closest("details")?.open,
    ).toBe(false);
  });
});
