import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import type { TimelineCandidatePort } from "./TimelineCandidatePort";
import { TimelineSupersessionEditor } from "./TimelineSupersessionEditor";
import {
  normalizeTimelineReason,
  type TimelineCaptureSubject,
  timelineCaptureIneligibility,
  timelineReplacementIneligibility,
} from "./timelineCaptureActionModel";

afterEach(cleanup);
const target: TimelineCaptureSubject = {
  recordId: "record-1",
  incidentId: "incident",
  rowVersion: 3,
  label: "Repeated label",
  context: "Monday · host-a",
  captureState: "enriched",
  replacementRecordId: null,
};
const replacement: TimelineCaptureSubject = {
  ...target,
  recordId: "record-2",
  context: "Tuesday · host-b",
};
function props() {
  return {
    target,
    authority: {
      actorId: "actor",
      incidentId: "incident",
      sessionIdentity: "session",
      role: "reviewer" as const,
      closed: false,
    },
    generation: 1,
    originKey: "Timeline:record-1",
    blocked: false,
    latestVersion: () => 3,
    port: {
      page: vi.fn<TimelineCandidatePort["page"]>(async () => ({
        kind: "accepted" as const,
        value: {
          rows: [target, replacement],
          hasMore: false,
          nextCursor: null,
        },
      })),
    },
    onPrepare: vi.fn(async () => true),
    onConfirm: vi.fn(() => true),
    onCancel: vi.fn(),
    onAccessLost: vi.fn(),
  };
}

it("Timeline capture policy preserves distinct review and supersession eligibility and reason semantics", () => {
  for (const captureState of [
    "rough",
    "enriched",
    "reviewed",
    "superseded",
    "invalid",
  ]) {
    expect(
      timelineCaptureIneligibility(
        { ...target, captureState },
        "mark-reviewed",
      ) === null,
    ).toBe(["rough", "enriched"].includes(captureState));
    expect(
      timelineCaptureIneligibility({ ...target, captureState }, "supersede") ===
        null,
    ).toBe(["rough", "enriched", "reviewed"].includes(captureState));
  }
  expect(timelineCaptureIneligibility(null, "mark-reviewed")).not.toBeNull();
  expect(timelineReplacementIneligibility(target, target)).not.toBeNull();
  expect(
    timelineReplacementIneligibility(
      { ...replacement, incidentId: "foreign" },
      target,
    ),
  ).not.toBeNull();
  expect(normalizeTimelineReason(" \u0085e\u0301\r\nreason\u00a0")).toBe(
    "é\nreason",
  );
  for (const reason of ["", " \n\t", "x\u0000", "\ud800", "x".repeat(4097)])
    expect(normalizeTimelineReason(reason)).toBeNull();
  expect(normalizeTimelineReason("😀".repeat(4096))).not.toBeNull();
});

it("Timeline supersession requires authored reason and frozen confirmation with optional replacement", async () => {
  const input = props();
  render(<TimelineSupersessionEditor {...input} />);
  expect(
    (
      screen.getByRole("button", {
        name: "Review supersession",
      }) as HTMLButtonElement
    ).disabled,
  ).toBe(true);
  await screen.findByRole("option", { name: /host-b/ });
  expect(screen.queryByRole("option", { name: /host-a/ })).toBeNull();
  fireEvent.change(screen.getByLabelText("Reason for supersession"), {
    target: { value: "  Duplicate source entry  " },
  });
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  const confirm = await screen.findByRole("button", {
    name: "Confirm supersede",
  });
  expect(input.onConfirm).not.toHaveBeenCalled();
  expect(screen.getByText("No replacement will be linked.")).not.toBeNull();
  expect(document.activeElement).toBe(confirm);
  fireEvent.click(confirm);
  expect(input.onConfirm).toHaveBeenCalledWith(
    expect.objectContaining({
      reason: "Duplicate source entry",
      replacement: null,
      target,
    }),
    expect.any(Function),
  );
  fireEvent.click(screen.getByRole("button", { name: "Back" }));
  expect(
    (screen.getByLabelText("Reason for supersession") as HTMLTextAreaElement)
      .value,
  ).toBe("  Duplicate source entry  ");
  fireEvent.change(
    screen.getByLabelText("Replacement Timeline row (optional)"),
    { target: { value: replacement.recordId } },
  );
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  fireEvent.click(
    await screen.findByRole("button", { name: "Confirm supersede" }),
  );
  expect(input.onConfirm).toHaveBeenLastCalledWith(
    expect.objectContaining({ replacement }),
    expect.any(Function),
  );
});

it("Timeline supersession invalidates confirmation on version authority replacement or reason changes", async () => {
  const input = props();
  const view = render(<TimelineSupersessionEditor {...input} />);
  await screen.findByRole("option", { name: /host-b/ });
  fireEvent.change(screen.getByLabelText("Reason for supersession"), {
    target: { value: "Correcting duplicate" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  await screen.findByRole("button", { name: "Confirm supersede" });
  view.rerender(
    <TimelineSupersessionEditor
      {...input}
      target={{ ...target, rowVersion: 4 }}
    />,
  );
  expect(
    screen.queryByRole("button", { name: "Confirm supersede" }),
  ).toBeNull();
  expect(document.activeElement).toBe(
    screen.getByLabelText("Reason for supersession"),
  );
  expect(input.onConfirm).not.toHaveBeenCalled();
  view.rerender(
    <TimelineSupersessionEditor
      {...input}
      authority={{ ...input.authority, closed: true }}
      generation={2}
    />,
  );
  expect(
    (
      screen.getByRole("button", {
        name: "Review supersession",
      }) as HTMLButtonElement
    ).disabled,
  ).toBe(true);
});

it("Timeline supersession fences changed inputs and closure during preparation", async () => {
  const input = props();
  const preparing = deferred<boolean>();
  input.onPrepare.mockReturnValue(preparing.promise);
  const view = render(<TimelineSupersessionEditor {...input} />);
  fireEvent.change(screen.getByLabelText("Reason for supersession"), {
    target: { value: "First reason" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  expect(input.onPrepare).toHaveBeenCalledTimes(1);
  fireEvent.change(screen.getByLabelText("Reason for supersession"), {
    target: { value: "Second reason" },
  });
  preparing.resolve(true);
  await screen.findByText(/reviewed inputs changed/);
  expect(
    screen.queryByRole("button", { name: "Confirm supersede" }),
  ).toBeNull();
  view.unmount();
  expect(input.onConfirm).not.toHaveBeenCalled();
});

it("Timeline replacement read failure still permits explicitly superseding without replacement", async () => {
  const input = props();
  input.port.page.mockRejectedValue(new TypeError("offline"));
  render(<TimelineSupersessionEditor {...input} />);
  await screen.findByRole("button", { name: "Retry replacements" });
  fireEvent.change(screen.getByLabelText("Reason for supersession"), {
    target: { value: "Obsolete" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  expect(
    (
      (await screen.findByRole("button", {
        name: "Confirm supersede",
      })) as HTMLButtonElement
    ).disabled,
  ).toBe(false);
});

it("Timeline replacement pagination finds duplicate labels beyond the current page and fences stale reads", async () => {
  const input = props();
  const pageTwo = deferred<Awaited<ReturnType<typeof input.port.page>>>();
  input.port.page
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { rows: [target], hasMore: true, nextCursor: "page-2" },
    })
    .mockReturnValueOnce(pageTwo.promise);
  const view = render(<TimelineSupersessionEditor {...input} />);
  await waitFor(() =>
    expect(
      (
        screen.getByRole("button", {
          name: "Next replacements",
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(false),
  );
  fireEvent.click(screen.getByRole("button", { name: "Next replacements" }));
  expect(input.port.page).toHaveBeenLastCalledWith(
    "page-2",
    expect.any(AbortSignal),
  );
  pageTwo.resolve({
    kind: "accepted",
    value: { rows: [replacement], hasMore: false, nextCursor: null },
  });
  expect(
    (
      (await screen.findByRole("option", {
        name: /host-b/,
      })) as HTMLOptionElement
    ).value,
  ).toBe("record-2");
  view.rerender(<TimelineSupersessionEditor {...input} generation={2} />);
  await waitFor(() =>
    expect(input.port.page).toHaveBeenLastCalledWith(
      null,
      expect.any(AbortSignal),
    ),
  );
});

it("Timeline candidate responses from retired authority cannot repaint the replacement selector", async () => {
  const input = props();
  const stale = deferred<Awaited<ReturnType<TimelineCandidatePort["page"]>>>();
  input.port.page.mockReturnValueOnce(stale.promise).mockResolvedValueOnce({
    kind: "accepted",
    value: { rows: [], nextCursor: null, hasMore: false },
  });
  const view = render(<TimelineSupersessionEditor {...input} />);
  await waitFor(() => expect(input.port.page).toHaveBeenCalledTimes(1));
  const oldSignal = input.port.page.mock.calls[0]?.[1];
  view.rerender(<TimelineSupersessionEditor {...input} generation={2} />);
  await waitFor(() => expect(input.port.page).toHaveBeenCalledTimes(2));
  expect(oldSignal?.aborted).toBe(true);
  stale.resolve({
    kind: "accepted",
    value: {
      rows: [{ ...replacement, label: "Retired protected candidate" }],
      nextCursor: null,
      hasMore: false,
    },
  });
  await waitFor(() =>
    expect(
      screen.queryByRole("option", { name: /Retired protected/ }),
    ).toBeNull(),
  );
  expect(screen.getByRole("option", { name: "No replacement" })).toBeTruthy();
});
