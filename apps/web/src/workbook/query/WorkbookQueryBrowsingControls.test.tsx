import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {
  StrictMode,
  Suspense,
  startTransition,
  useEffect,
  useState,
} from "react";
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { acceptedQueryMetadata } from "../../testing/workbookQueryTestSupport";
import { useWorkbookQueryController } from "../hooks/useWorkbookQueryController";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { notesViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useGenericSurfaceQuery } from "./useGenericSurfaceQuery";
import {
  useWorkbookBrowsingRegistry,
  useWorkbookQueryBrowser,
  WorkbookQueryBrowsingProvider,
} from "./WorkbookQueryBrowsingContext";
import { WorkbookQueryBrowsingControls } from "./WorkbookQueryBrowsingControls";
import type {
  WorkbookViewQueryPort,
  WorkbookViewQueryResult,
} from "./WorkbookViewQueryPort";

function requireObserved<T>(value: T | undefined): T {
  if (value === undefined)
    throw new Error("Expected committed fixture observation");
  return value;
}
const contract = requireViewContract(notesViewSchemaId);
const recordId = (number: number) =>
  `00000000-0000-4000-8000-${String(number).padStart(12, "0")}`;
function Harness({
  port,
  suspended,
  controlsViewSchemaId = notesViewSchemaId,
}: {
  readonly port: WorkbookViewQueryPort;
  readonly suspended?: Promise<void>;
  readonly controlsViewSchemaId?: string;
}) {
  const intent = useWorkbookQueryController({ surface: notesViewSchemaId });
  const query = useGenericSurfaceQuery({
    active: true,
    contract,
    queryState: intent.snapshot.genericQueryState,
    viewQuery: port,
    viewSchemaId: notesViewSchemaId,
    onAuthorityUncertain: undefined,
  });
  useEffect(() => {
    void query.refresh();
  }, [query.refresh]);
  if (suspended !== undefined) throw suspended;
  return (
    <>
      <button
        type="button"
        onClick={() =>
          intent.snapshot.activeQueryControls.onSortChange([
            { fieldKey: "note.title", direction: "desc" },
          ])
        }
      >
        Sort titles descending
      </button>
      <WorkbookQueryBrowsingControls viewSchemaId={controlsViewSchemaId} />
      <output aria-label="Applied sort">
        {query.acceptedQueryState.sort[0]?.direction ?? "default"}
      </output>
      <output aria-label="Visible rows">
        {query.rows
          .map((row) => String(row.cells["note.title"]?.value))
          .join(",")}
      </output>
    </>
  );
}
function queryPort() {
  const cursors = new Map([
    [" token+/100= ", 100],
    [" token+/200= ", 200],
    [" token+/300= ", 300],
    [" token+/400= ", 400],
  ]);
  const query = vi.fn<WorkbookViewQueryPort["query"]>(async (input) => {
    const start =
      input.cursorToken === undefined ? 0 : cursors.get(input.cursorToken);
    if (start === undefined) throw new Error("Unknown fixture cursor");
    const end = Math.min(start + 100, 405);
    return {
      kind: "accepted",
      value: {
        ...acceptedQueryMetadata(notesViewSchemaId, input.queryState),
        incidentId: "incident",
        viewSchemaId: notesViewSchemaId,
        producingRequest: {
          queryState: input.queryState,
          limit: 100,
          ...(input.cursorToken === undefined
            ? {}
            : { cursorToken: input.cursorToken }),
        },
        rows: Array.from({ length: end - start }, (_, index) =>
          fullWorkbookViewRow(contract, recordId(start + index), 1, {
            "note.title": `Note ${start + index}`,
            "note.body": "body",
          }),
        ),
        paging: {
          limit: 100,
          hasMore: end < 405,
          nextCursor: end < 405 ? ` token+/${end}= ` : null,
        },
      },
    };
  });
  return query;
}
function fixture() {
  const query = queryPort();
  render(
    <WorkbookQueryBrowsingProvider>
      <Harness port={{ query }} />
    </WorkbookQueryBrowsingProvider>,
  );
  return query;
}
afterEach(cleanup);

it("requires the browsing provider and refuses inactive or uncommitted reads", async () => {
  const query = queryPort();
  expect(() =>
    renderHook(() => useWorkbookQueryBrowser({ query }, notesViewSchemaId)),
  ).toThrow("WorkbookQueryBrowsingProvider is required");
  const state = emptyWorkbookQueryState();
  const input = {
    active: false,
    contract,
    queryState: state,
    viewQuery: { query },
    viewSchemaId: notesViewSchemaId,
    onAuthorityUncertain: undefined,
  };
  const hook = renderHook(() => useGenericSurfaceQuery(input), {
    wrapper: WorkbookQueryBrowsingProvider,
  });
  expect(hook.result.current.browser).toBeUndefined();
  await expect(
    hook.result.current.refresh({ requireAcceptance: true }),
  ).rejects.toThrow();
  expect(query).not.toHaveBeenCalled();
  const gate = deferred<void>();
  let candidate: ReturnType<typeof useGenericSurfaceQuery> | undefined;
  function Uncommitted(): never {
    candidate = useGenericSurfaceQuery({ ...input, active: true });
    throw gate.promise;
  }
  render(
    <WorkbookQueryBrowsingProvider>
      <Suspense fallback={<p>Preparing</p>}>
        <Uncommitted />
      </Suspense>
    </WorkbookQueryBrowsingProvider>,
  );
  expect(candidate?.browser).toBeUndefined();
  await expect(
    requireObserved(candidate).refresh({ requireAcceptance: true }),
  ).rejects.toThrow();
  expect(query).not.toHaveBeenCalled();
});

it("reestablishes committed ownership in Strict Mode and rejects the retired response", async () => {
  const first = deferred<WorkbookViewQueryResult>();
  const second = deferred<WorkbookViewQueryResult>();
  const source = queryPort();
  const query = vi
    .fn<WorkbookViewQueryPort["query"]>()
    .mockReturnValueOnce(first.promise)
    .mockReturnValueOnce(second.promise);
  render(
    <StrictMode>
      <WorkbookQueryBrowsingProvider>
        <Harness port={{ query }} />
      </WorkbookQueryBrowsingProvider>
    </StrictMode>,
  );
  await waitFor(() => expect(query).toHaveBeenCalledTimes(2));
  expect(query.mock.calls[0]?.[0].signal?.aborted).toBe(true);
  await act(async () =>
    second.resolve(await source(requireObserved(query.mock.calls[1])[0])),
  );
  await screen.findByText("100 records loaded; more available.");
  await act(() =>
    first.resolve({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Retired authority" },
    }),
  );
  expect(screen.queryByText("Retired authority")).toBeNull();
  expect(screen.getByText("100 records loaded; more available.")).toBeTruthy();
});

it("retires a replaced query port before reading its successor and refuses its late result", async () => {
  const pending = deferred<WorkbookViewQueryResult>();
  const source = queryPort();
  const oldQuery = vi.fn<WorkbookViewQueryPort["query"]>(() => pending.promise);
  const state = emptyWorkbookQueryState();
  const hook = renderHook(
    ({ query }) =>
      useGenericSurfaceQuery({
        active: true,
        contract,
        queryState: state,
        viewQuery: { query },
        viewSchemaId: notesViewSchemaId,
        onAuthorityUncertain: undefined,
      }),
    {
      initialProps: { query: oldQuery },
      wrapper: WorkbookQueryBrowsingProvider,
    },
  );
  let read: Promise<void> | undefined;
  act(() => {
    read = hook.result.current.refresh({ requireAcceptance: true });
  });
  const rejected = expect(read).rejects.toThrow();
  const previous = requireObserved(hook.result.current.browser);
  hook.rerender({ query: source });
  expect(oldQuery.mock.calls[0]?.[0].signal?.aborted).toBe(true);
  await act(() => hook.result.current.refresh());
  await act(async () => {
    pending.resolve(await source(requireObserved(oldQuery.mock.calls[0])[0]));
    await rejected;
  });
  expect(previous.getSnapshot()).toMatchObject({
    accepted: null,
    hasEarlier: false,
  });
  expect(hook.result.current.rows).toHaveLength(100);
  expect(hook.result.current.browser).not.toBe(previous);
});

it("fences predecessor cleanup and readers after a successor commits", async () => {
  const query = queryPort();
  const state = emptyWorkbookQueryState();
  const observed = new Map<string, ReturnType<typeof useGenericSurfaceQuery>>();
  let registry: ReturnType<typeof useWorkbookBrowsingRegistry> | undefined;
  function Reader({ name }: { name: string }) {
    registry = useWorkbookBrowsingRegistry();
    observed.set(
      name,
      useGenericSurfaceQuery({
        active: true,
        contract,
        queryState: state,
        viewQuery: { query },
        viewSchemaId: notesViewSchemaId,
        onAuthorityUncertain: undefined,
      }),
    );
    return null;
  }
  const subject = (old: boolean, next: boolean) => (
    <WorkbookQueryBrowsingProvider>
      {old ? <Reader key="old" name="old" /> : null}
      {next ? <Reader key="next" name="next" /> : null}
    </WorkbookQueryBrowsingProvider>
  );
  const mounted = render(subject(true, false));
  const oldRefresh = requireObserved(observed.get("old")).refresh;
  await act(() => oldRefresh());
  mounted.rerender(subject(true, true));
  await act(() => requireObserved(observed.get("next")).refresh());
  const accepted = requireObserved(
    requireObserved(registry).find(notesViewSchemaId),
  ).getSnapshot().accepted;
  mounted.rerender(subject(false, true));
  expect(
    requireObserved(
      requireObserved(registry).find(notesViewSchemaId),
    ).getSnapshot().accepted,
  ).toBe(accepted);
  await expect(oldRefresh({ requireAcceptance: true })).rejects.toThrow();
  expect(query).toHaveBeenCalledTimes(2);
  await act(() =>
    requireObserved(registry).activate(notesViewSchemaId, "more"),
  );
  expect(requireObserved(observed.get("next")).rows).toHaveLength(200);
  expect(query).toHaveBeenCalledTimes(3);
});

it("retains bounded return state on detachment and clears it on authority retirement", async () => {
  const query = queryPort();
  let registry: ReturnType<typeof useWorkbookBrowsingRegistry> | undefined;
  function Owner() {
    registry = useWorkbookBrowsingRegistry();
    return null;
  }
  const subject = (visible: boolean, session = "one") => (
    <WorkbookQueryBrowsingProvider key={session}>
      <Owner />
      {visible ? <Harness port={{ query }} /> : null}
    </WorkbookQueryBrowsingProvider>
  );
  const mounted = render(subject(true));
  await screen.findByText("100 records loaded; more available.");
  for (let index = 0; index < 3; index++)
    await act(() =>
      requireObserved(registry).activate(notesViewSchemaId, "more"),
    );
  const detached = requireObserved(
    requireObserved(registry).find(notesViewSchemaId),
  );
  detached.rememberAnchor(recordId(250));
  mounted.rerender(subject(false));
  expect(requireObserved(registry).find(notesViewSchemaId)).toBeUndefined();
  expect(detached.getSnapshot()).toMatchObject({
    accepted: null,
    pageCount: 0,
    hasEarlier: true,
  });
  mounted.rerender(subject(true));
  await screen.findByText("100 records loaded; more available.");
  expect(query.mock.calls.at(-1)?.[0].cursorToken).toBe(" token+/200= ");
  const pending = deferred<WorkbookViewQueryResult>();
  query.mockReturnValueOnce(pending.promise);
  let late: Promise<void>;
  act(() => {
    late = requireObserved(registry).activate(notesViewSchemaId, "more");
  });
  mounted.rerender(subject(true, "two"));
  await screen.findByText("100 records loaded; more available.");
  expect(detached.getSnapshot()).toMatchObject({
    accepted: null,
    pageCount: 0,
    hasEarlier: false,
    canonicalQuery: null,
  });
  expect(query.mock.calls.at(-1)?.[0].cursorToken).toBeUndefined();
  await act(async () => {
    pending.resolve({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Old session" },
    });
    await late;
  });
  expect(
    requireObserved(
      requireObserved(registry).find(notesViewSchemaId),
    ).getSnapshot().accepted?.rows,
  ).toHaveLength(100);
});

it("preserves committed browsing when a replacement render is suspended and abandoned", async () => {
  const query = queryPort();
  const replacement = queryPort();
  const gate = deferred<void>();
  let registry: ReturnType<typeof useWorkbookBrowsingRegistry> | undefined;
  let replace: (value: {
    port: WorkbookViewQueryPort;
    suspended?: Promise<void>;
  }) => void;
  function Subject() {
    registry = useWorkbookBrowsingRegistry();
    const [input, setInput] = useState<{
      port: WorkbookViewQueryPort;
      suspended?: Promise<void>;
    }>({ port: { query } });
    replace = setInput;
    return (
      <Suspense fallback={<p>Waiting</p>}>
        <Harness {...input} />
      </Suspense>
    );
  }
  render(
    <WorkbookQueryBrowsingProvider>
      <Subject />
    </WorkbookQueryBrowsingProvider>,
  );
  await screen.findByText("100 records loaded; more available.");
  const original = requireObserved(
    requireObserved(registry).find(notesViewSchemaId),
  );
  const accepted = original.getSnapshot().accepted;
  act(() =>
    startTransition(() =>
      replace({ port: { query: replacement }, suspended: gate.promise }),
    ),
  );
  expect(requireObserved(registry).find(notesViewSchemaId)).toBe(original);
  expect(original.getSnapshot().accepted).toBe(accepted);
  expect(replacement).not.toHaveBeenCalled();
  act(() => replace({ port: { query } }));
  await userEvent
    .setup()
    .click(screen.getByRole("button", { name: "Load more" }));
  await screen.findByText("200 records loaded; more available.");
  expect(query.mock.calls.at(-1)?.[0].cursorToken).toBe(" token+/100= ");
});

it("reaches later records and returns by keyboard while preserving continuation control focus", async () => {
  const query = fixture();
  const user = userEvent.setup();
  await screen.findByText("100 records loaded; more available.");
  expect(screen.queryByRole("button", { name: "Earlier rows" })).toBeNull();
  const more = screen.getByRole("button", { name: "Load more" });
  more.focus();
  await user.keyboard("{Enter}");
  await screen.findByText("200 records loaded; more available.");
  expect(document.activeElement).toBe(more);
  await user.keyboard("{Enter}");
  await screen.findByText("300 records loaded; more available.");
  await user.keyboard("{Enter}");
  await waitFor(() =>
    expect(screen.getByLabelText("Visible rows").textContent).toContain(
      "Note 399",
    ),
  );
  expect(screen.getByLabelText("Visible rows").textContent?.split(",")[0]).toBe(
    "Note 100",
  );
  expect(query).toHaveBeenCalledTimes(4);
  const earlier = screen.getByRole("button", { name: "Earlier rows" });
  earlier.focus();
  await user.keyboard("{Enter}");
  await waitFor(() =>
    expect(
      screen.getByLabelText("Visible rows").textContent?.split(",")[0],
    ).toBe("Note 0"),
  );
  expect(query.mock.calls.at(-1)?.[0].cursorToken).toBeUndefined();
  expect(document.activeElement).toBe(earlier);
  await user.tab();
  expect(screen.queryByRole("button", { name: "Earlier rows" })).toBeNull();

  more.focus();
  for (let page = 0; page < 4; page++) {
    await user.keyboard("{Enter}");
    await waitFor(() =>
      expect(more.getAttribute("aria-disabled")).toBe(
        page === 3 ? "true" : "false",
      ),
    );
  }
  await screen.findByText("205 records loaded; end of current results.");
  expect(document.activeElement).toBe(more);
  const reads = query.mock.calls.length;
  await user.keyboard("{Enter}");
  expect(query).toHaveBeenCalledTimes(reads);
  await user.tab();
  expect(screen.queryByRole("button", { name: "Load more" })).toBeNull();
  expect(screen.getByRole("button", { name: "Refresh" })).toBeTruthy();
});

it("retains accepted presentation on a failed replacement and offers local retry and revert", async () => {
  const query = fixture();
  const user = userEvent.setup();
  await screen.findByText("100 records loaded; more available.");
  const pending = deferred<WorkbookViewQueryResult>();
  query.mockImplementationOnce(() => pending.promise);
  await user.click(
    screen.getByRole("button", { name: "Sort titles descending" }),
  );
  expect(screen.getByLabelText("Applied sort").textContent).toBe("default");
  await act(() =>
    pending.resolve({
      kind: "rejected",
      failure: { kind: "retryable", message: "Offline" },
    }),
  );
  expect(screen.getByText("Query changes are unapplied.")).toBeTruthy();
  expect(
    screen
      .getByRole("button", { name: "Load more" })
      .getAttribute("aria-disabled"),
  ).toBe("true");
  const revert = screen.getByRole("button", { name: "Revert" });
  await user.click(revert);
  await waitFor(() =>
    expect(revert.getAttribute("aria-disabled")).toBe("true"),
  );
  await user.tab();
  expect(screen.queryByRole("button", { name: "Revert" })).toBeNull();
  expect(screen.getByLabelText("Applied sort").textContent).toBe("default");
  query.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "retryable", message: "Offline" },
  });
  await user.click(
    screen.getByRole("button", { name: "Sort titles descending" }),
  );
  await screen.findByRole("button", { name: "Retry" });
  await user.click(screen.getByRole("button", { name: "Retry" }));
  await waitFor(() =>
    expect(screen.getByLabelText("Applied sort").textContent).toBe("desc"),
  );
  expect(query.mock.calls.at(-1)?.[0].queryState.sort).toEqual([
    { fieldKey: "note.title", direction: "desc" },
  ]);
});

it("keeps keyboard Retry focused through held reads, repeated failure and success", async () => {
  const query = fixture();
  const user = userEvent.setup();
  await screen.findByText("100 records loaded; more available.");
  query.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "retryable", message: "Offline" },
  });
  await user.click(
    screen.getByRole("button", { name: "Sort titles descending" }),
  );
  const retry = await screen.findByRole("button", { name: "Retry" });
  retry.focus();
  const repeatedFailure = deferred<WorkbookViewQueryResult>();
  query.mockImplementationOnce(() => repeatedFailure.promise);
  await user.keyboard("{Enter}");
  await waitFor(() => expect(query).toHaveBeenCalledTimes(3));
  expect(retry.isConnected).toBe(true);
  expect(document.activeElement).toBe(retry);
  expect(retry.getAttribute("aria-busy")).toBe("true");
  expect(
    screen
      .getByText("100 records loaded. Retrying records…")
      .getAttribute("aria-live"),
  ).toBe("off");
  expect(screen.queryByText("Query changes are unapplied.")).toBeNull();
  await user.keyboard("{Enter}");
  expect(query).toHaveBeenCalledTimes(3);
  await act(() =>
    repeatedFailure.resolve({
      kind: "rejected",
      failure: { kind: "retryable", message: "Still offline" },
    }),
  );
  expect(document.activeElement).toBe(retry);
  expect(retry.getAttribute("aria-disabled")).toBe("false");

  const success = deferred<WorkbookViewQueryResult>();
  const source = queryPort();
  query.mockImplementationOnce(() => success.promise);
  await user.keyboard("{Enter}");
  await waitFor(() => expect(query).toHaveBeenCalledTimes(4));
  expect(document.activeElement).toBe(retry);
  const request = requireObserved(query.mock.calls[3])[0];
  await act(async () => success.resolve(await source(request)));
  await screen.findByText("100 records loaded; more available.");
  expect(document.activeElement).toBe(retry);
  expect(retry.getAttribute("aria-disabled")).toBe("true");
  expect(
    screen
      .getByText("100 records loaded; more available.")
      .getAttribute("aria-live"),
  ).toBe("polite");
  await user.tab();
  expect(screen.queryByRole("button", { name: "Retry" })).toBeNull();
});

it("keeps keyboard Revert locally focused after restoring the accepted query", async () => {
  const query = fixture();
  const user = userEvent.setup();
  await screen.findByText("100 records loaded; more available.");
  query.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "retryable", message: "Offline" },
  });
  await user.click(
    screen.getByRole("button", { name: "Sort titles descending" }),
  );
  const revert = await screen.findByRole("button", { name: "Revert" });
  revert.focus();
  await user.keyboard("{Enter}");
  await waitFor(() => expect(query).toHaveBeenCalledTimes(3));
  expect(document.activeElement).toBe(revert);
  expect(revert.isConnected).toBe(true);
  expect(revert.getAttribute("aria-disabled")).toBe("true");
  await user.tab();
  expect(screen.queryByRole("button", { name: "Revert" })).toBeNull();
});

it("retires focused recovery after surface and authority changes", async () => {
  const query = queryPort();
  const port = { query };
  let registry: ReturnType<typeof useWorkbookBrowsingRegistry> | undefined;
  function Owner() {
    registry = useWorkbookBrowsingRegistry();
    return null;
  }
  const subject = (controlsViewSchemaId: string) => (
    <WorkbookQueryBrowsingProvider>
      <Owner />
      <Harness port={port} controlsViewSchemaId={controlsViewSchemaId} />
    </WorkbookQueryBrowsingProvider>
  );
  const mounted = render(subject(notesViewSchemaId));
  const user = userEvent.setup();
  await screen.findByText("100 records loaded; more available.");
  query.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "retryable", message: "Offline" },
  });
  await user.click(
    screen.getByRole("button", { name: "Sort titles descending" }),
  );
  const retry = await screen.findByRole("button", { name: "Retry" });
  retry.focus();
  mounted.rerender(subject("different-surface"));
  expect(retry.isConnected).toBe(false);
  await act(() =>
    requireObserved(registry).activate(notesViewSchemaId, "retry"),
  );
  mounted.rerender(subject(notesViewSchemaId));
  expect(screen.queryByRole("button", { name: "Retry" })).toBeNull();

  query.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "retryable", message: "Offline again" },
  });
  await act(() =>
    requireObserved(registry).activate(notesViewSchemaId, "restart"),
  );
  const current = await screen.findByRole("button", { name: "Retry" });
  current.focus();
  act(() => requireObserved(registry).invalidateAll());
  expect(screen.queryByRole("button", { name: "Retry" })).toBeNull();
});
