import {
  act,
  cleanup,
  fireEvent,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import { useState } from "react";
import { afterEach, expect, it, vi } from "vitest";
import {
  mentionCreationReceipt,
  mentionInspector,
  mentionReview,
  mentionWorkbookRow,
} from "../../../testing/timelineMentionTestSupport";
import { createTimelineMentionCandidateReader } from "../adapters/createTimelineMentionCandidateReader";
import { TimelineMentionActionControls } from "../components/TimelineMentionActionControls";
import { useTimelineMentionActions } from "../hooks/useTimelineMentionActions";
import type { TimelineMentionCandidatePort } from "./TimelineMentionCandidatePort";
import {
  initialMentionCreateDraft,
  mentionCreateRequest,
} from "./timelineMentionCreationModel";
import { useTimelineMentionCandidates } from "./useTimelineMentionCandidates";
import { WorkbookTimelineMentionOperationOwner } from "./WorkbookTimelineMentionOperationOwner";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
it("Mention loaded-target filtering preserves selected identity and exposes keyboard reachable unresolved dismissal", async () => {
  const review = mentionReview(),
    row = mentionWorkbookRow(review);
  const owner = new WorkbookTimelineMentionOperationOwner(
    review.subject.incidentId,
    { create: () => "unused" },
    { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() },
  );
  owner.setAuthority(review.authority);
  const port: TimelineMentionCandidatePort = {
    page: async () => ({
      kind: "accepted",
      value: {
        candidates: [
          {
            recordId: "first",
            rowVersion: 1,
            displayText: "First",
            entityType: "host",
          },
          {
            recordId: "second",
            rowVersion: 2,
            displayText: "Second",
            entityType: "host",
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    }),
  };
  function Panel() {
    const [selectedTargetId, setSelectedTargetId] = useState("");
    const actions = useTimelineMentionActions({
      owner,
      candidatePort: port,
      rowsRef: { current: [row] },
      earlierSaves: { current: Promise.resolve() },
      selectedMention: mentionInspector(review),
      selectedTargetId,
      setSelectedTargetId,
      presentationKey: "picker",
      presentationActive: true,
      waitForCommittedRecordIdle: async () => ({ row, rowVersion: 4 }),
      setInspectorMessage: vi.fn(),
    });
    return <TimelineMentionActionControls actions={actions} />;
  }
  render(<Panel />);
  await screen.findByRole("option", { name: "Second" });
  const select = screen.getByRole("combobox", {
    name: "Resolve to existing",
  }) as HTMLSelectElement;
  select.focus();
  expect(document.activeElement).toBe(select);
  fireEvent.change(select, { target: { value: "first" } });
  fireEvent.change(
    screen.getByRole("textbox", { name: "Filter loaded targets" }),
    { target: { value: "no match" } },
  );
  expect(select.value).toBe("first");
  expect(screen.queryByRole("option", { name: "Second" })).toBeNull();
  expect(screen.getByText("No loaded targets match this filter.")).toBeTruthy();
  fireEvent.keyDown(select, { key: "Escape" });
  const disclosure = screen.getByText("Correction and resolution");
  expect(document.activeElement).toBe(disclosure);
  expect((disclosure.parentElement as HTMLDetailsElement).open).toBe(false);
  expect(select.isConnected).toBe(true);
  expect(select.value).toBe("first");
  fireEvent.click(disclosure);
  await waitFor(() =>
    expect((disclosure.parentElement as HTMLDetailsElement).open).toBe(true),
  );
  expect(screen.getByRole("combobox", { name: "Resolve to existing" })).toBe(
    select,
  );
  expect(
    (
      screen.getByRole("textbox", {
        name: "Filter loaded targets",
      }) as HTMLInputElement
    ).value,
  ).toBe("no match");
  expect(
    (screen.getByRole("button", { name: "Dismiss" }) as HTMLButtonElement)
      .disabled,
  ).toBe(false);
});
it("Mention candidates query eligible targets independently with contract sorting and explicit cursor pages", async () => {
  const review = mentionReview();
  const receipt = mentionCreationReceipt({
    ...review,
    draft: initialMentionCreateDraft(review.subject),
  });
  const fetch = vi.fn().mockImplementation(
    async () =>
      new Response(
        JSON.stringify({
          data: {
            incident_id: review.subject.incidentId,
            view_schema_id: receipt.data.view_schema_id,
            rows: [receipt.data.row],
          },
          meta: {
            request_id: "targets",
            query: { filters: [], sort: [] },
            paging: { limit: 100, has_more: true, next_cursor: "next-page" },
          },
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
  );
  vi.stubGlobal("fetch", fetch);
  const reader = createTimelineMentionCandidateReader({
    apiBase: "/service",
    incidentId: review.subject.incidentId,
  });
  const result = await reader.page("host", null, new AbortController().signal);
  expect(result.kind).toBe("accepted");
  const request = JSON.parse(fetch.mock.calls[0]?.[1].body);
  expect(request).toMatchObject({
    limit: 100,
    filters: [
      {
        field_key: "host.host_state",
        op: "eq",
        arg: { values: ["stub", "canonical"] },
      },
    ],
  });
  expect(request.sort).toBeUndefined(); // Omission requests the contract default sort.
  expect(request.cursor_token).toBeUndefined();
  await reader.page("host", "requested-cursor", new AbortController().signal);
  expect(JSON.parse(fetch.mock.calls[1]?.[1].body).cursor_token).toBe(
    "requested-cursor",
  );
});
it("Mention candidate paging retains loaded targets through read failure and retries the missing page", async () => {
  const first = {
    recordId: "first",
    rowVersion: 1,
    displayText: "First",
    entityType: "host" as const,
  };
  const second = { ...first, recordId: "second", displayText: "Second" };
  const page = vi
    .fn<TimelineMentionCandidatePort["page"]>()
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [first], hasMore: true, nextCursor: "page-2" },
    })
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Offline" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [second], hasMore: false, nextCursor: null },
    });
  const port = { page };
  const hook = renderHook(() =>
    useTimelineMentionCandidates(port, "host", "mention", true),
  );
  await waitFor(() => expect(hook.result.current.phase).toBe("ready"));
  await act(() => hook.result.current.loadMore());
  expect(hook.result.current).toMatchObject({
    phase: "failed",
    candidates: [first],
    error: "Offline",
  });
  await act(() => hook.result.current.retry());
  expect(hook.result.current.candidates).toEqual([first, second]);
  expect(page.mock.calls.map((call) => call[1])).toEqual([
    null,
    "page-2",
    "page-2",
  ]);
});
it("Mention contextual authoring seeds only display name and requires reviewed identity fields", () => {
  for (const entityType of ["host", "identity"] as const) {
    const base = mentionReview(),
      subject = {
        ...base.subject,
        entityType,
        rawText: "user@machine.example",
      };
    const draft = initialMentionCreateDraft(subject);
    expect(Object.entries(draft).filter(([, value]) => value)).toEqual([
      [`${entityType}.display_name`, subject.rawText],
    ]);
    const request = mentionCreateRequest(
      { subject, authority: base.authority, draft },
      "reviewed-create",
    );
    expect(request).not.toBeNull();
    expect(JSON.stringify(request)).not.toContain('"' + entityType + '.fqdn"');
  }
});
