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
import { WorkbookCandidateAuthorityContext } from "../../hooks/useWorkbookCandidateDiscovery";
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
      selectedMentionRef: review.subject.itemRef,
      selectedRowId: review.subject.sourceRecordId,
      inspectorReviewGeneration: 0,
      inspectorAttachmentGeneration: 0,
      refreshProjection: async () => {},
      reviewSurfaceKey: "picker",
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
it("Mention failed cursor-free restart retries the first page rather than the old continuation", async () => {
  const first = {
    recordId: "first",
    rowVersion: 1,
    displayText: "First",
    entityType: "host" as const,
  };
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
      kind: "rejected",
      failure: { kind: "retryable", message: "Restart failed" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [first], hasMore: false, nextCursor: null },
    });
  const port = { page };
  const hook = renderHook(() =>
    useTimelineMentionCandidates(port, "host", "mention", true),
  );
  await waitFor(() => expect(hook.result.current.phase).toBe("ready"));
  await act(() => hook.result.current.loadMore());
  await act(() => hook.result.current.restart());
  await act(() => hook.result.current.retry());
  expect(page.mock.calls.map((call) => call[1])).toEqual([
    null,
    "page-2",
    null,
    null,
  ]);
});
it("Mention retry remains connected and focused while its response is held", async () => {
  const review = mentionReview();
  const row = mentionWorkbookRow(review);
  const owner = new WorkbookTimelineMentionOperationOwner(
    review.subject.incidentId,
    { create: () => "unused" },
    { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() },
  );
  owner.setAuthority(review.authority);
  const first = {
    recordId: "first",
    rowVersion: 1,
    displayText: "First",
    entityType: "host" as const,
  };
  let finish:
    | ((
        value: Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>,
      ) => void)
    | undefined;
  const held = new Promise<
    Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>
  >((resolve) => {
    finish = resolve;
  });
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
    .mockImplementationOnce(() => held);
  const port = { page };
  function Panel() {
    const [selectedTargetId, setSelectedTargetId] = useState("");
    const actions = useTimelineMentionActions({
      owner,
      candidatePort: port,
      rowsRef: { current: [row] },
      earlierSaves: { current: Promise.resolve() },
      selectedMention: mentionInspector(review),
      selectedMentionRef: review.subject.itemRef,
      selectedRowId: review.subject.sourceRecordId,
      inspectorReviewGeneration: 0,
      inspectorAttachmentGeneration: 0,
      refreshProjection: async () => {},
      reviewSurfaceKey: "retry-focus",
      selectedTargetId,
      setSelectedTargetId,
      presentationKey: "retry-focus",
      presentationActive: true,
      waitForCommittedRecordIdle: async () => ({ row, rowVersion: 4 }),
      setInspectorMessage: vi.fn(),
    });
    return <TimelineMentionActionControls actions={actions} />;
  }
  render(<Panel />);
  await screen.findByRole("button", { name: "Load more targets" });
  fireEvent.click(screen.getByRole("button", { name: "Load more targets" }));
  const retry = await screen.findByRole("button", {
    name: "Retry target read",
  });
  retry.focus();
  fireEvent.click(retry);
  await waitFor(() => expect(page).toHaveBeenCalledTimes(3));
  expect(retry.isConnected).toBe(true);
  expect(document.activeElement).toBe(retry);
  const filter = screen.getByRole("textbox", { name: "Filter loaded targets" });
  filter.focus();
  fireEvent.scroll(document);
  await act(async () => {
    finish?.({
      kind: "accepted",
      value: { candidates: [first], hasMore: false, nextCursor: null },
    });
    await held;
  });
  expect(document.activeElement).toBe(filter);
});
it("Mention scope replacement retires pending read focus intent", async () => {
  const review = mentionReview();
  const row = mentionWorkbookRow(review);
  const owner = new WorkbookTimelineMentionOperationOwner(
    review.subject.incidentId,
    { create: () => "unused" },
    { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() },
  );
  owner.setAuthority(review.authority);
  const first = {
    recordId: "first",
    rowVersion: 1,
    displayText: "First",
    entityType: "host" as const,
  };
  let finish:
    | ((
        value: Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>,
      ) => void)
    | undefined;
  const held = new Promise<
    Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>
  >((resolve) => {
    finish = resolve;
  });
  const page = vi
    .fn<TimelineMentionCandidatePort["page"]>()
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [first], hasMore: true, nextCursor: "page-2" },
    })
    .mockImplementationOnce(() => held)
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [first], hasMore: false, nextCursor: null },
    });
  const port = { page };
  function Panel({ scope }: { scope: string }) {
    const [selectedTargetId, setSelectedTargetId] = useState("");
    const actions = useTimelineMentionActions({
      owner,
      candidatePort: port,
      rowsRef: { current: [row] },
      earlierSaves: { current: Promise.resolve() },
      selectedMention: mentionInspector(review),
      selectedMentionRef: review.subject.itemRef,
      selectedRowId: review.subject.sourceRecordId,
      inspectorReviewGeneration: 0,
      inspectorAttachmentGeneration: 0,
      refreshProjection: async () => {},
      reviewSurfaceKey: scope,
      selectedTargetId,
      setSelectedTargetId,
      presentationKey: scope,
      presentationActive: true,
      waitForCommittedRecordIdle: async () => ({ row, rowVersion: 4 }),
      setInspectorMessage: vi.fn(),
    });
    return <TimelineMentionActionControls actions={actions} />;
  }
  const view = render(<Panel scope="old" />);
  const loadMore = await screen.findByRole("button", {
    name: "Load more targets",
  });
  loadMore.focus();
  fireEvent.click(loadMore);
  await waitFor(() => expect(page).toHaveBeenCalledTimes(2));
  const select = screen.getByRole("combobox", {
    name: "Resolve to existing",
  }) as HTMLSelectElement;
  const focusSelect = vi.spyOn(select, "focus");
  view.rerender(<Panel scope="new" />);
  await waitFor(() => expect(page).toHaveBeenCalledTimes(3));
  expect(focusSelect).not.toHaveBeenCalled();
  await act(async () => {
    finish?.({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Old access" },
    });
    await held;
  });
  expect(focusSelect).not.toHaveBeenCalled();
});
it("Mention initial retry and terminal paging admit one read per activation", async () => {
  const candidate = {
    recordId: "first",
    rowVersion: 1,
    displayText: "First",
    entityType: "identity" as const,
  };
  let finish:
    | ((
        value: Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>,
      ) => void)
    | undefined;
  const held = new Promise<
    Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>
  >((resolve) => {
    finish = resolve;
  });
  const page = vi
    .fn<TimelineMentionCandidatePort["page"]>()
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Offline" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [candidate], hasMore: true, nextCursor: "next" },
    })
    .mockImplementationOnce(() => held);
  const port = { page };
  const hook = renderHook(() =>
    useTimelineMentionCandidates(port, "identity", "initial", true),
  );
  await waitFor(() => expect(hook.result.current.phase).toBe("failed"));
  expect(hook.result.current.failedRead).toMatchObject({
    kind: "initial",
    cursor: null,
  });
  await act(() => hook.result.current.retry());
  expect(hook.result.current.candidates).toEqual([candidate]);
  await act(async () => {
    void hook.result.current.loadMore();
    void hook.result.current.loadMore();
  });
  expect(page).toHaveBeenCalledTimes(3);
  expect(hook.result.current.candidates).toEqual([candidate]);
  await act(async () => {
    finish?.({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    await held;
  });
  expect(hook.result.current.hasMore).toBe(false);
  await act(() => hook.result.current.loadMore());
  expect(page).toHaveBeenCalledTimes(3);
  expect(page.mock.calls.map((call) => call[1])).toEqual([null, null, "next"]);
});
it("Mention malformed and repeated continuations require an explicit fresh chain", async () => {
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
      kind: "accepted",
      value: { candidates: [second], hasMore: true, nextCursor: "page-2" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [second], hasMore: false, nextCursor: null },
    });
  const port = { page };
  const hook = renderHook(() =>
    useTimelineMentionCandidates(port, "host", "chain", true, "first"),
  );
  await waitFor(() => expect(hook.result.current.phase).toBe("ready"));
  await act(() => hook.result.current.loadMore());
  expect(hook.result.current.failure).toBe("unusable_continuation");
  expect(hook.result.current.candidates).toEqual([first]);
  await act(() => hook.result.current.restart());
  expect(hook.result.current.candidates).toEqual([second]);
  expect(hook.result.current.staleCandidates).toEqual([first]);
  expect(page.mock.calls.map((call) => call[1])).toEqual([
    null,
    "page-2",
    null,
  ]);
});
it("Mention server-invalid continuation keeps its typed failed identity for exact retry", async () => {
  const first = {
    recordId: "first",
    rowVersion: 1,
    displayText: "First",
    entityType: "host" as const,
  };
  const invalid = {
    kind: "rejected" as const,
    failure: {
      kind: "validation" as const,
      message: "Invalid cursor",
      publicCode: "invalid_view_query",
      publicReason: "invalid_cursor_token" as const,
    },
  };
  const page = vi
    .fn<TimelineMentionCandidatePort["page"]>()
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [first], hasMore: true, nextCursor: "page-2" },
    })
    .mockResolvedValueOnce(invalid)
    .mockResolvedValueOnce(invalid);
  const port = { page };
  const hook = renderHook(() =>
    useTimelineMentionCandidates(port, "host", "invalid", true),
  );
  await waitFor(() => expect(hook.result.current.phase).toBe("ready"));
  await act(() => hook.result.current.loadMore());
  expect(hook.result.current.failure).toBe("unusable_continuation");
  expect(hook.result.current.failedRead).toMatchObject({
    kind: "continuation",
    cursor: "page-2",
  });
  await act(() => hook.result.current.retry());
  expect(hook.result.current.failure).toBe("unusable_continuation");
  expect(page.mock.calls.map((call) => call[1])).toEqual([
    null,
    "page-2",
    "page-2",
  ]);
});
it("Mention late scope results are fenced and current authority loss conceals candidates", async () => {
  const old = {
    recordId: "old",
    rowVersion: 1,
    displayText: "Old",
    entityType: "host" as const,
  };
  const fresh = { ...old, recordId: "fresh", displayText: "Fresh" };
  let releaseOld:
    | ((
        value: Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>,
      ) => void)
    | undefined;
  const held = new Promise<
    Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>
  >((resolve) => {
    releaseOld = resolve;
  });
  const page = vi
    .fn<TimelineMentionCandidatePort["page"]>()
    .mockImplementationOnce(() => held)
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [fresh], hasMore: true, nextCursor: "next" },
    })
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "No access" },
    });
  const port = { page };
  const onAuthorityFailure = vi.fn();
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <WorkbookCandidateAuthorityContext.Provider
      value={{
        identity: "account-incident",
        canRead: true,
        onAuthorityFailure,
      }}
    >
      {children}
    </WorkbookCandidateAuthorityContext.Provider>
  );
  const hook = renderHook(
    ({ scope }) => useTimelineMentionCandidates(port, "host", scope, true),
    { initialProps: { scope: "old" }, wrapper },
  );
  hook.rerender({ scope: "new" });
  await waitFor(() => expect(hook.result.current.candidates).toEqual([fresh]));
  await act(async () => {
    releaseOld?.({
      kind: "accepted",
      value: { candidates: [old], hasMore: false, nextCursor: null },
    });
    await held;
  });
  expect(hook.result.current.candidates).toEqual([fresh]);
  await act(() => hook.result.current.loadMore());
  expect(hook.result.current.failure).toBe("authority");
  expect(hook.result.current.candidates).toEqual([]);
  expect(hook.result.current.failedRead).toBeNull();
  expect(onAuthorityFailure).toHaveBeenCalledTimes(1);
});
it("Mention obsolete authority failure cannot affect a replacement entity type", async () => {
  const identity = {
    recordId: "identity",
    rowVersion: 1,
    displayText: "Identity",
    entityType: "identity" as const,
  };
  let releaseOld:
    | ((
        value: Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>,
      ) => void)
    | undefined;
  const held = new Promise<
    Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>
  >((resolve) => {
    releaseOld = resolve;
  });
  const page = vi
    .fn<TimelineMentionCandidatePort["page"]>()
    .mockImplementationOnce(() => held)
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [identity], hasMore: false, nextCursor: null },
    });
  const port = { page };
  const onAuthorityFailure = vi.fn();
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <WorkbookCandidateAuthorityContext.Provider
      value={{ identity: "same-authority", canRead: true, onAuthorityFailure }}
    >
      {children}
    </WorkbookCandidateAuthorityContext.Provider>
  );
  const hook = renderHook(
    ({ entityType }) =>
      useTimelineMentionCandidates(port, entityType, "mention", true),
    { initialProps: { entityType: "host" as "host" | "identity" }, wrapper },
  );
  hook.rerender({ entityType: "identity" });
  await waitFor(() =>
    expect(hook.result.current.candidates).toEqual([identity]),
  );
  await act(async () => {
    releaseOld?.({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Old access" },
    });
    await held;
  });
  expect(hook.result.current.candidates).toEqual([identity]);
  expect(onAuthorityFailure).not.toHaveBeenCalled();
});
it("Mention restart retains selected label and filter without admitting an unvalidated target", async () => {
  const review = mentionReview();
  const row = mentionWorkbookRow(review);
  const create = vi.fn(() => "unused");
  const owner = new WorkbookTimelineMentionOperationOwner(
    review.subject.incidentId,
    { create },
    { remember: vi.fn(), settle: vi.fn(), accepted: vi.fn() },
  );
  owner.setAuthority(review.authority);
  const first = {
    recordId: "first",
    rowVersion: 1,
    displayText: "First",
    entityType: "host" as const,
  };
  const second = { ...first, recordId: "second", displayText: "Second" };
  let finish:
    | ((
        value: Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>,
      ) => void)
    | undefined;
  const held = new Promise<
    Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>
  >((resolve) => {
    finish = resolve;
  });
  const page = vi
    .fn<TimelineMentionCandidatePort["page"]>()
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { candidates: [first], hasMore: false, nextCursor: null },
    })
    .mockImplementationOnce(() => held);
  const port = { page };
  function Panel() {
    const [selectedTargetId, setSelectedTargetId] = useState("");
    const actions = useTimelineMentionActions({
      owner,
      candidatePort: port,
      rowsRef: { current: [row] },
      earlierSaves: { current: Promise.resolve() },
      selectedMention: mentionInspector(review),
      selectedMentionRef: review.subject.itemRef,
      selectedRowId: review.subject.sourceRecordId,
      inspectorReviewGeneration: 0,
      inspectorAttachmentGeneration: 0,
      refreshProjection: async () => {},
      reviewSurfaceKey: "stale-selection",
      selectedTargetId,
      setSelectedTargetId,
      presentationKey: "stale-selection",
      presentationActive: true,
      waitForCommittedRecordIdle: async () => ({ row, rowVersion: 4 }),
      setInspectorMessage: vi.fn(),
    });
    return <TimelineMentionActionControls actions={actions} />;
  }
  render(<Panel />);
  await screen.findByRole("option", { name: "First" });
  const select = screen.getByRole("combobox", {
    name: "Resolve to existing",
  }) as HTMLSelectElement;
  fireEvent.change(select, { target: { value: "first" } });
  const filter = screen.getByRole("textbox", {
    name: "Filter loaded targets",
  }) as HTMLInputElement;
  fireEvent.change(filter, { target: { value: "Fir" } });
  fireEvent.click(
    screen.getByRole("button", { name: "Restart target discovery" }),
  );
  expect(select.value).toBe("first");
  expect(filter.value).toBe("Fir");
  expect(
    (
      screen.getByRole("option", {
        name: /First.*awaiting revalidation/u,
      }) as HTMLOptionElement
    ).disabled,
  ).toBe(true);
  expect(
    (
      screen.getByRole("button", {
        name: "Resolve to existing",
      }) as HTMLButtonElement
    ).disabled,
  ).toBe(true);
  expect(create).not.toHaveBeenCalled();
  await act(async () => {
    finish?.({
      kind: "accepted",
      value: { candidates: [second], hasMore: false, nextCursor: null },
    });
    await held;
  });
  expect(select.value).toBe("first");
  expect(
    (
      screen.getByRole("button", {
        name: "Resolve to existing",
      }) as HTMLButtonElement
    ).disabled,
  ).toBe(true);
  fireEvent.change(filter, { target: { value: "" } });
  fireEvent.change(select, { target: { value: "second" } });
  expect(
    (
      screen.getByRole("button", {
        name: "Resolve to existing",
      }) as HTMLButtonElement
    ).disabled,
  ).toBe(false);
  expect(create).not.toHaveBeenCalled();
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
