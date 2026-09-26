import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  historyDiffFixture,
  historyPresentationFixture,
} from "../../testing/workbookHistoryTestSupport";
import { createWorkbookRecordHistoryAdapter } from "../adapters/createWorkbookRecordHistoryAdapter";
import { useWorkbookRecordHistoryController } from "../inspector/useWorkbookRecordHistoryController";
import { useWorkbookRecordHistoryState } from "../inspector/useWorkbookRecordHistoryState";
import {
  workbookRecordHistoryLoadedData,
  workbookRecordHistoryReducer,
} from "../inspector/workbookRecordHistoryModel";
import { WorkbookRecordHistoryOwner } from "./WorkbookRecordHistoryOwner";
import { beginHistoryRead, rejectHistoryRead } from "./workbookHistoryBrowsing";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";
const subject = {
  kind: "live" as const,
  recordId,
  rowVersion: 4,
  label: "Row",
  surfaceLabel: "Timeline",
  viewSchemaId: "cartulary.view.timeline.v2",
};
const item = {
  actor_user_id: "40000000-0000-4000-8000-000000000001",
  committed_at: "2026-09-10T00:00:00Z",
  history_item_ref: "retained-item",
  history_entry_ref: "opaque-selector",
  operation: "patch",
  change_set_id: "30000000-0000-4000-8000-000000000001",
  reversible: true,
  available_rollback_actions: ["history_entry" as const],
  diff_summary: historyDiffFixture("A retained change"),
};
const data = {
  incident_id: incidentId,
  record_id: recordId,
  row_version: 4,
  deleted: false,
  representation_generation: "cartulary.history.1",
  items: [item],
};
const terminal = { limit: 100, has_more: false, next_cursor: null } as const;
afterEach(() => vi.unstubAllGlobals());

function setup() {
  const owner = new WorkbookRecordHistoryOwner(incidentId, {
    create: () => crypto.randomUUID(),
  });
  owner.setAuthority({
    actorId: item.actor_user_id,
    sessionIdentity: "session",
    incidentId,
    role: "reviewer",
    closed: false,
  });
  const send = vi.fn(async () => ({ kind: "uncertain" as const }));
  const load = vi.fn(
    async (
      _recordId: string,
      _signal: AbortSignal,
      request?: { cursorToken?: string },
    ) => ({
      kind: "accepted" as const,
      value: request?.cursorToken
        ? { ...data, paging: terminal }
        : {
            ...data,
            representation_generation: "cartulary.history.1",
            items: [],
            paging: {
              limit: 100,
              has_more: true as const,
              next_cursor: "older",
            },
          },
    }),
  );
  owner.configure({ load, send });
  const pending = {
    kind: "rollback" as const,
    action: "history_entry" as const,
    historyItemRef: item.history_item_ref,
    recordId,
    rowVersion: 4,
    target: {
      kind: "history_entry" as const,
      history_entry_ref: item.history_entry_ref,
    },
  };
  const binding = {
    isCurrent: () => true,
    coordinate: async () => 4,
    acknowledged: vi.fn(),
    reconcile: async () => {},
  };
  return { owner, send, load, pending, binding };
}

describe("History browsing characterization", () => {
  it("preserves the server continuation contract on a short page", async () => {
    const paging = { limit: 100, has_more: true, next_cursor: "opaque-next" };
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({ data, meta: { request_id: "request", paging } }),
            { status: 200, headers: { "content-type": "application/json" } },
          ),
      ),
    );
    const port = createWorkbookRecordHistoryAdapter({
      apiBase: undefined,
      incidentId,
    });
    expect(
      await port.load(recordId, new AbortController().signal),
    ).toMatchObject({ kind: "accepted", value: { ...data, paging } });
  });
  it("keeps accepted history when a refresh read fails", () => {
    const ready = historyPresentationFixture(subject, {
      ...data,
      paging: terminal,
    });
    const requested = beginHistoryRead(required(ready.browsing), "refresh");
    const failed = rejectHistoryRead(requested, required(requested.pending), {
      kind: "retryable",
      message: "Read failed",
    });
    const state = workbookRecordHistoryReducer(ready, {
      type: "browsing_changed",
      browsing: failed,
    });
    expect(workbookRecordHistoryLoadedData(state)).toEqual({
      ...data,
      paging: terminal,
    });
  });
  it("previews a selected retained item beyond the first page", async () => {
    const t = setup();
    const { result } = renderHook(() =>
      useWorkbookRecordHistoryController({
        presentation: useWorkbookRecordHistoryState(),
        coordinate: async () => subject.rowVersion,
        owner: t.owner,
        subject,
        canMutate: true,
        ownerEffects: {
          refresh: async () => {},
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted: vi.fn(),
        },
      }),
    );
    act(() => result.current.commands.open());
    await waitFor(() => expect(result.current.snapshot.phase).toBe("ready"));
    act(() => result.current.commands.previewRollback(item, "history_entry"));
    await waitFor(() =>
      expect(result.current.snapshot).toMatchObject({
        phase: "ready",
        pendingAction: t.pending,
      }),
    );
    expect(t.send).not.toHaveBeenCalled();
  });
  it("admits a confirmed retained item beyond the first page", async () => {
    const t = setup();
    const attempt = t.owner.admit({ subject, pending: t.pending }, t.binding);
    expect(attempt).not.toBeNull();
    if (attempt) await t.owner.execute(attempt);
    expect(t.send).toHaveBeenCalledOnce();
  });
  it("reviews a retained later-page target before replacement admission", async () => {
    const t = setup();
    const attempt = t.owner.admit({ subject, pending: t.pending }, t.binding);
    expect(attempt).not.toBeNull();
    if (attempt) await t.owner.review(attempt.id);
    expect(t.owner.getSnapshot()[0]?.currentHistory?.items).toContainEqual(
      item,
    );
    expect(t.send).not.toHaveBeenCalled();
  });
  it("fences duplicate continuation and retains the exact failed-page retry", async () => {
    const t = setup();
    const { result } = renderHook(() =>
      useWorkbookRecordHistoryController({
        presentation: useWorkbookRecordHistoryState(),
        coordinate: async () => subject.rowVersion,
        owner: t.owner,
        subject,
        canMutate: true,
        ownerEffects: {
          refresh: async () => {},
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted: vi.fn(),
        },
      }),
    );
    act(() => result.current.commands.open());
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.accepted).not.toBeNull(),
    );
    const read = required(t.load.getMockImplementation());
    let fail!: () => void;
    t.load.mockImplementationOnce(
      () =>
        new Promise((_resolve, reject) => {
          fail = () => reject(new Error("read failed"));
        }),
    );
    act(() => {
      result.current.commands.loadOlder();
      result.current.commands.loadOlder();
    });
    expect(t.load).toHaveBeenCalledTimes(2);
    act(() => fail());
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.failure).not.toBeNull(),
    );
    expect(result.current.snapshot.browsing?.accepted).not.toBeNull();
    t.load.mockImplementation(read);
    act(() => result.current.commands.retryRead());
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.accepted?.data.items).toEqual([
        item,
      ]),
    );
    expect(t.load.mock.calls.map((call) => call[2])).toEqual([
      {},
      { cursorToken: "older" },
      { cursorToken: "older" },
    ]);
    expect(t.send).not.toHaveBeenCalled();
  });
  it("clears old record before paint and rejects an obsolete continuation", async () => {
    const t = setup();
    const { result, rerender } = renderHook(
      ({ recordId }) =>
        useWorkbookRecordHistoryController({
          presentation: useWorkbookRecordHistoryState(),
          coordinate: async () => subject.rowVersion,
          owner: t.owner,
          subject: { ...subject, recordId },
          canMutate: true,
          ownerEffects: {
            refresh: async () => {},
            deleteAccepted: vi.fn(),
            restoreAccepted: vi.fn(),
            rollbackAccepted: vi.fn(),
          },
        }),
      { initialProps: { recordId } },
    );
    act(() => result.current.commands.open());
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.accepted).not.toBeNull(),
    );
    let finish!: (value: Awaited<ReturnType<typeof t.load>>) => void;
    t.load.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    act(() => result.current.commands.loadOlder());
    const signal = required(t.load.mock.calls[1])[1];
    rerender({ recordId: "20000000-0000-4000-8000-000000000002" });
    expect(workbookRecordHistoryLoadedData(result.current.snapshot)).toBeNull();
    expect(signal.aborted).toBe(true);
    await act(async () =>
      finish({ kind: "accepted", value: { ...data, paging: terminal } }),
    );
    expect(
      workbookRecordHistoryLoadedData(result.current.snapshot)?.record_id,
    ).not.toBe(recordId);
    expect(t.send).not.toHaveBeenCalled();
  });
  it("replaces session scope conceals lost access and preserves closed incident reads", async () => {
    const t = setup();
    const { result } = renderHook(() =>
      useWorkbookRecordHistoryController({
        presentation: useWorkbookRecordHistoryState(),
        coordinate: async () => subject.rowVersion,
        owner: t.owner,
        subject,
        canMutate: true,
        ownerEffects: {
          refresh: async () => {},
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted: vi.fn(),
        },
      }),
    );
    act(() => result.current.commands.open());
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.accepted).not.toBeNull(),
    );
    const old = required(result.current.snapshot.browsing).scope;
    act(() =>
      t.owner.setAuthority({
        actorId: item.actor_user_id,
        sessionIdentity: "replacement",
        incidentId,
        role: "reviewer",
        closed: false,
      }),
    );
    expect(workbookRecordHistoryLoadedData(result.current.snapshot)).toBeNull();
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.accepted).not.toBeNull(),
    );
    expect(result.current.snapshot.browsing?.scope).not.toEqual(old);
    act(() => t.owner.closeIncident());
    expect(
      workbookRecordHistoryLoadedData(result.current.snapshot),
    ).not.toBeNull();
    act(() => result.current.commands.loadOlder());
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.accepted?.data.items).toEqual([
        item,
      ]),
    );
    expect(t.owner.permitted("rollback")).toBe(false);
    act(() => t.owner.suspend());
    expect(workbookRecordHistoryLoadedData(result.current.snapshot)).toBeNull();
    expect(t.send).not.toHaveBeenCalled();
  });
  it("cancels hidden presentation reads without dispatching an action", async () => {
    const t = setup();
    const { result, rerender } = renderHook(
      ({ active }) =>
        useWorkbookRecordHistoryController({
          presentation: useWorkbookRecordHistoryState(),
          coordinate: async () => subject.rowVersion,
          owner: t.owner,
          subject,
          presentationActive: active,
          canMutate: true,
          ownerEffects: {
            refresh: async () => {},
            deleteAccepted: vi.fn(),
            restoreAccepted: vi.fn(),
            rollbackAccepted: vi.fn(),
          },
        }),
      { initialProps: { active: true } },
    );
    act(() => result.current.commands.open());
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.accepted).not.toBeNull(),
    );
    t.load.mockImplementationOnce(() => new Promise(() => {}));
    act(() => result.current.commands.loadOlder());
    rerender({ active: false });
    expect(t.load.mock.calls[1]?.[1].aborted).toBe(true);
    expect(result.current.snapshot.browsing?.pending).toBeNull();
    expect(result.current.snapshot.browsing?.accepted).not.toBeNull();
    expect(result.current.snapshot.browsing?.failure).not.toBeNull();
    rerender({ active: true });
    act(() => result.current.commands.retryRead());
    await waitFor(() =>
      expect(result.current.snapshot.browsing?.failure).toBeNull(),
    );
    expect(result.current.snapshot.browsing?.accepted?.data.items).toEqual([
      item,
    ]);
    expect(t.send).not.toHaveBeenCalled();
  });
});

function required<T>(value: T | undefined | null): T {
  if (value === undefined || value === null)
    throw new Error("Missing history fixture value");
  return value;
}
