import { requireViewContract } from "@cartulary/view-contracts";
import { cleanup, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createNoteCreateTransport } from "../../adapters/createNoteCreateTransport";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import {
  type NoteCreateReader,
  noteCreateView,
  noteFeature,
  noteSourceViews,
} from "./noteCreateModel";
import type {
  NoteOutcome,
  NoteReceipt,
  NoteTransport,
} from "./noteCreateOperation";
import { noteRecoveryItems } from "./noteRecoveryItems";
import { WorkbookNoteCreateOwner } from "./WorkbookNoteCreateOwner";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});
const authority: WorkbookMutationAuthority = {
  actorId: "10000000-0000-4000-8000-000000000001",
  incidentId: "40000000-0000-4000-8000-000000000004",
  sessionIdentity: "session",
  role: "editor",
  closed: false,
};
const sourceId = "20000000-0000-4000-8000-000000000002",
  noteId = "30000000-0000-4000-8000-000000000003";
const token = Symbol("test");
function fixture(view = noteSourceViews[0]) {
  let sequence = 0;
  const ids = { create: vi.fn(() => `note-create-${++sequence}`) };
  const effects = {
    coordinate: vi.fn<() => Promise<WorkbookSourceWriteSettlement>>(
      async () => ({ kind: "settled", minimumRowVersion: 0 }),
    ),
    accepted: vi.fn(),
    refresh: vi.fn(async () => {}),
  };
  const owner = new WorkbookNoteCreateOwner(authority.incidentId, ids, effects);
  owner.setAuthority(authority);
  const reader: NoteCreateReader = {
    availableViews: vi.fn(async () => noteSourceViews),
    verifyNote: vi.fn(async () => {}),
    page: vi.fn(async (input) => ({
      kind: "accepted" as const,
      value: {
        candidates:
          input.viewSchemaId === view
            ? [
                {
                  recordId: sourceId,
                  displayText: "Source",
                  viewSchemaId: view,
                  row: { record_id: sourceId, row_version: 1, cells: {} },
                },
              ]
            : [],
        hasMore: false,
        nextCursor: null,
      },
    })),
  };
  const authorityReader = vi.fn(
    async (): Promise<WorkbookMutationAuthority> => authority,
  );
  const transport = {
    ...createNoteCreateTransport("/original-api"),
    send: vi.fn<NoteTransport["send"]>(async () => ({ kind: "uncertain" })),
  };
  owner.configure(reader, authorityReader, transport);
  owner.begin(
    {
      subject: {
        kind: "live",
        recordId: sourceId,
        viewSchemaId: view,
        rowVersion: 1,
        label: "Source",
        surfaceLabel: requireViewContract(view).title,
      },
      cells: {},
    },
    required(noteFeature(view)),
    { kind: "view_schema", id: view },
    token,
  );
  owner.update("note.title", "Retained Note");
  const receipt: NoteReceipt = {
    data: {
      view_schema_id: noteCreateView,
      source_record_id: sourceId,
      link_type: "references_artifact",
      change_set_id: "50000000-0000-4000-8000-000000000005",
      row: fullWorkbookViewRow(requireViewContract(noteCreateView), noteId, 1, {
        "note.title": "Retained Note",
      }),
    },
    meta: { request_id: "request-1" },
  };
  return { owner, reader, authorityReader, transport, ids, effects, receipt };
}
describe("Note atomic recovery", () => {
  it("separates retained authoring and acknowledged refresh from unsettled save work", async () => {
    const { owner, transport, receipt, effects } = fixture();
    expect(owner.unsettledMutationCount).toBe(0);
    expect(owner.getSnapshot().draft).not.toBeNull();
    const retained = noteRecoveryItems(owner.getSnapshot());
    expect(retained).toHaveLength(1);
    expect(retained[0]?.attention).toBe("draft");
    await owner.submit(token);
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
    expect(owner.unsettledMutationCount).toBe(1);
    expect(noteRecoveryItems(owner.getSnapshot())).toMatchObject([
      { id: retained[0]?.id, attention: "attention" },
    ]);
    const refresh = deferred<void>();
    effects.refresh.mockReturnValue(refresh.promise);
    transport.send.mockResolvedValueOnce({ kind: "accepted", receipt });
    const recovery = owner.replay(
      required(owner.getSnapshot().entries[0]).attempt.clientTxnId,
    );
    await waitFor(() => expect(effects.refresh).toHaveBeenCalled());
    expect(owner.unsettledMutationCount).toBe(0);
    expect(owner.pendingCount).toBeGreaterThan(0);
    refresh.reject(new Error("refresh unavailable"));
    await recovery;
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    expect(owner.unsettledMutationCount).toBe(0);
    await waitFor(() => expect(owner.blockedCount).toBe(1));
  });

  it("retains socket observations without fabricating receipts or regressing newer source removal evidence", async () => {
    const { owner, transport, receipt, effects, reader } = fixture();
    const pending = deferred<NoteOutcome>();
    transport.send.mockReturnValue(pending.promise);
    const submitted = owner.submit(token);
    await waitFor(() => expect(transport.send).toHaveBeenCalledOnce());
    const event: RecordChangedMessage = {
      type: "record_changed",
      incident_id: authority.incidentId,
      event_id: "note-event",
      emitted_at: "2026-09-13T05:00:00Z",
      stream_seq: 1,
      payload: {
        record_id: noteId,
        row_version: 1,
        client_txn_id: "note-create-1",
        actor_user_id: authority.actorId,
        change_set_id: receipt.data.change_set_id,
        changed_field_keys: [],
        affected_views: [
          { view_schema_id: noteCreateView, change_kind: "invalidate" },
          { view_schema_id: "future.additive.view", change_kind: "invalidate" },
        ],
      },
    };
    owner.observeSocket(event);
    owner.observeSocket(event);
    expect(owner.getSnapshot().entries[0]?.receipt).toBeNull();
    expect(owner.getSnapshot().entries[0]?.observations).toHaveLength(1);
    expect(effects.accepted).not.toHaveBeenCalled();
    pending.resolve({ kind: "accepted", receipt });
    await submitted;
    await waitFor(() =>
      expect(owner.getSnapshot().entries[0]?.refresh).toBe("complete"),
    );
    expect(
      vi.mocked(reader.page).mock.calls.map(([input]) => input.viewSchemaId),
    ).not.toContain("future.additive.view");
    const removed: RecordChangedMessage = {
      ...event,
      event_id: "source-removed",
      payload: {
        ...event.payload,
        record_id: sourceId,
        row_version: 4,
        client_txn_id: "other-write",
        affected_views: [
          { view_schema_id: noteSourceViews[0], change_kind: "remove" },
        ],
      },
    };
    owner.observeSocket(removed);
    owner.observeSocket({
      ...removed,
      event_id: "old-source",
      payload: {
        ...removed.payload,
        row_version: 2,
        affected_views: [
          { view_schema_id: noteSourceViews[0], change_kind: "invalidate" },
        ],
      },
    });
    owner.beginSheet({ kind: "view_schema", id: noteCreateView }, token);
    owner.update("note.body", "New draft");
    owner.changeSource({
      recordId: sourceId,
      viewSchemaId: noteSourceViews[0],
      rowVersion: 1,
      label: "Stale page",
    });
    await owner.submit(token);
    expect(transport.send).toHaveBeenCalledTimes(1);
    expect(owner.getSnapshot().message).toContain("unavailable");
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
  });
  it("requires deliberate renewed review after closure reopening and role restoration", async () => {
    const { owner, authorityReader, transport } = fixture();
    for (const changed of [
      { ...authority, closed: true },
      { ...authority, role: "viewer" as const },
    ]) {
      owner.setAuthority(changed);
      authorityReader.mockResolvedValue(changed);
      await owner.submit(token);
      expect(transport.send).not.toHaveBeenCalled();
      owner.setAuthority(authority);
      authorityReader.mockResolvedValue(authority);
      await owner.submit(token);
      expect(transport.send).not.toHaveBeenCalled();
      expect(owner.getSnapshot().draft?.values["note.title"]).toBe(
        "Retained Note",
      );
      await owner.review();
      expect(owner.getSnapshot().needsReview).toBe(false);
    }
    await owner.submit(token);
    expect(transport.send).toHaveBeenCalledTimes(1);
  });
  it("guards same-frame submissions before source preparation and freezes all four source routes", async () => {
    for (const view of noteSourceViews) {
      const { owner, effects, transport, ids } = fixture(view);
      const save = deferred<WorkbookSourceWriteSettlement>();
      effects.coordinate.mockReturnValue(save.promise);
      const first = owner.submit(token),
        second = owner.submit(token);
      expect(effects.coordinate).toHaveBeenCalledTimes(1);
      expect(transport.send).not.toHaveBeenCalled();
      owner.update("note.title", "Changed during preparation");
      owner.changeSource(null);
      save.resolve({ kind: "settled", minimumRowVersion: 0 });
      await Promise.all([first, second]);
      expect(transport.send).toHaveBeenCalledTimes(1);
      expect(ids.create).toHaveBeenCalledTimes(1);
      const attempt = required(owner.getSnapshot().entries[0]).attempt;
      expect(attempt.operationID).toBe("createRecordLinkedNote");
      expect(attempt.path).toBe(`/api/v1/records/${sourceId}/linked-notes`);
      expect(attempt.pathParameters).toEqual({ record_id: sourceId });
      expect(attempt.apiBase).toBe("/original-api");
      expect(attempt.review.draft.source?.viewSchemaId).toBe(view);
      expect(attempt.body).toBe(
        JSON.stringify({
          client_txn_id: "note-create-1",
          "note.title": "Retained Note",
        }),
      );
      expect(Object.isFrozen(attempt.request)).toBe(true);
      expect(Object.isFrozen(attempt.review.draft.source)).toBe(true);
    }
  });
  it("uses ordinary atomic creation only after explicit source clearing without losing authored fields", async () => {
    const { owner, transport } = fixture();
    owner.update("note.body", "Body");
    owner.changeSource(null);
    await owner.submit(token);
    const attempt = required(transport.send.mock.calls[0])[0];
    expect(attempt.operationID).toBe("createViewRow");
    expect(attempt.pathParameters).toEqual({
      incident_id: authority.incidentId,
      view_schema_id: noteCreateView,
    });
    expect(attempt.request).toMatchObject({
      "note.title": "Retained Note",
      "note.body": "Body",
    });
  });
  it("replays exact bytes after response loss without fresh source checks and blocks draft mutation", async () => {
    const { owner, transport, reader, authorityReader, receipt, ids } =
      fixture();
    await owner.submit(token);
    const attempt = required(owner.getSnapshot().entries[0]).attempt;
    owner.update("note.title", "Changed");
    owner.changeSource(null);
    owner.discard();
    expect(owner.getSnapshot().draft?.values["note.title"]).toBe(
      "Retained Note",
    );
    expect(owner.getSnapshot().draft?.source?.recordId).toBe(sourceId);
    const readCalls = vi.mocked(reader.verifyNote).mock.calls.length;
    owner.closeIncident();
    authorityReader.mockResolvedValue({ ...authority, closed: true });
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    transport.send.mockResolvedValue({ kind: "accepted", receipt });
    await owner.replay(attempt.clientTxnId);
    expect(transport.send.mock.calls[1]?.[0]).toBe(attempt);
    expect(ids.create).toHaveBeenCalledTimes(1);
    expect(reader.verifyNote).toHaveBeenCalledTimes(readCalls);
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
  });
  it("accepts a late original receipt monotonically while an exact replay is running", async () => {
    vi.useFakeTimers();
    const { owner, transport, receipt, effects } = fixture();
    const original = deferred<NoteOutcome>(),
      replay = deferred<NoteOutcome>();
    transport.send
      .mockReturnValueOnce(original.promise)
      .mockReturnValueOnce(replay.promise);
    const submitted = owner.submit(token);
    await vi.advanceTimersByTimeAsync(0);
    await vi.advanceTimersByTimeAsync(30_000);
    await submitted;
    const id = required(owner.getSnapshot().entries[0]).attempt.clientTxnId;
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
    owner.detach(token);
    const recovering = owner.replay(id);
    await vi.advanceTimersByTimeAsync(0);
    original.resolve({ kind: "accepted", receipt });
    await vi.advanceTimersByTimeAsync(0);
    replay.resolve({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "late rejection" },
    });
    await recovering;
    expect(owner.getSnapshot().entries[0]?.phase).toBe("accepted");
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    expect(effects.accepted).toHaveBeenCalledTimes(1);
    expect(owner.getSnapshot().draft).toBeNull();
  });
  it("retains acceptance before failed effects and recovers refresh using reads only", async () => {
    const { owner, transport, receipt, effects, reader } = fixture();
    effects.refresh.mockRejectedValueOnce(new Error("refresh failed"));
    transport.send.mockResolvedValue({ kind: "accepted", receipt });
    await owner.submit(token);
    await waitFor(() =>
      expect(owner.getSnapshot().entries[0]?.refresh).toBe("required"),
    );
    const entry = required(owner.getSnapshot().entries[0]);
    expect(entry.receipt).toEqual(receipt);
    expect(owner.getSnapshot().draft).toBeNull();
    const reads = vi.mocked(reader.page).mock.calls.length;
    await owner.retryRefresh(entry.attempt.clientTxnId);
    expect(owner.getSnapshot().entries[0]?.refresh).toBe("complete");
    expect(reader.page).toHaveBeenCalledTimes(reads + 2);
    expect(transport.send).toHaveBeenCalledTimes(1);
    expect(effects.accepted).toHaveBeenCalledTimes(1);
  });
  it("preserves editable rejection input and allocates a new ID for a deliberate corrected action", async () => {
    const { owner, transport, ids } = fixture();
    transport.send.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "validation", message: "Correct the title" },
    });
    await owner.submit(token);
    expect(owner.busy).toBe(false);
    owner.update("note.title", "Corrected");
    await owner.submit(token);
    expect(ids.create).toHaveBeenCalledTimes(2);
    expect(
      owner.getSnapshot().entries.map((entry) => entry.attempt.clientTxnId),
    ).toEqual(["note-create-1", "note-create-2"]);
    expect(transport.send.mock.calls[0]?.[0].body).toContain("Retained Note");
    expect(transport.send.mock.calls[1]?.[0].body).toContain("Corrected");
  });
  it("requires explicit review after source version changes and blocks deleted sources and prior conflicts", async () => {
    const { owner, transport, reader, effects } = fixture();
    effects.coordinate.mockResolvedValueOnce({
      kind: "blocked",
      reason: "pending_recovery",
    });
    await owner.submit(token);
    expect(transport.send).not.toHaveBeenCalled();
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: {
        candidates: [
          {
            recordId: sourceId,
            displayText: "Source",
            viewSchemaId: noteSourceViews[0],
            row: { record_id: sourceId, row_version: 2, cells: {} },
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    });
    await owner.submit(token);
    expect(transport.send).not.toHaveBeenCalled();
    expect(owner.getSnapshot().needsReview).toBe(true);
    await owner.review();
    expect(owner.getSnapshot().draft?.source?.rowVersion).toBe(2);
    expect(owner.getSnapshot().needsReview).toBe(false);
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    await owner.submit(token);
    expect(transport.send).not.toHaveBeenCalled();
    expect(owner.getSnapshot().message).toContain("unavailable");
    owner.changeSource(null);
    await owner.submit(token);
    expect(transport.send).toHaveBeenCalledTimes(1);
  });
  it("prevents stale authority callbacks from dispatching after account replacement", async () => {
    const { owner, transport, authorityReader } = fixture();
    const pending = deferred<WorkbookMutationAuthority>();
    authorityReader.mockReturnValue(pending.promise);
    const submitting = owner.submit(token);
    await waitFor(() => expect(authorityReader).toHaveBeenCalled());
    owner.setAuthority({ ...authority, actorId: "replacement" });
    pending.resolve(authority);
    await submitting;
    expect(owner.getSnapshot().draft).toBeNull();
    expect(transport.send).not.toHaveBeenCalled();
    expect(owner.getSnapshot().authority?.actorId).toBe("replacement");
  });
  it("retains concealed late acceptance during suspension and ignores retired lifetime callbacks", async () => {
    for (const retire of [false, true]) {
      const { owner, transport, receipt, effects } = fixture();
      const pending = deferred<NoteOutcome>();
      transport.send.mockReturnValue(pending.promise);
      const submitted = owner.submit(token);
      await waitFor(() => expect(transport.send).toHaveBeenCalled());
      if (retire) owner.retire();
      else owner.suspend();
      pending.resolve({ kind: "accepted", receipt });
      await submitted;
      expect(owner.getSnapshot().entries).toEqual([]);
      expect(effects.accepted).not.toHaveBeenCalled();
      owner.setAuthority(authority);
      expect(owner.getSnapshot().entries).toHaveLength(retire ? 0 : 1);
      if (!retire) {
        expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
        await owner.retryRefresh("note-create-1");
        expect(effects.accepted).toHaveBeenCalledTimes(1);
      }
    }
  });
  it("keeps malformed success receipts uncertain and refuses replay without verified edit authority", async () => {
    const { owner, transport, receipt, authorityReader } = fixture();
    transport.send.mockResolvedValue({
      kind: "accepted",
      receipt: {
        ...receipt,
        data: { ...receipt.data, row: { ...receipt.data.row, cells: {} } },
      },
    });
    await owner.submit(token);
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
    authorityReader.mockRejectedValueOnce(new Error("unknown"));
    await owner.replay("note-create-1");
    expect(transport.send).toHaveBeenCalledTimes(1);
    authorityReader.mockResolvedValue({ ...authority, role: "viewer" });
    await owner.replay("note-create-1");
    expect(transport.send).toHaveBeenCalledTimes(1);
    expect(owner.getSnapshot().draft?.values["note.title"]).toBe(
      "Retained Note",
    );
  });
  it("keeps replay rejections uncertain because they cannot disprove an earlier commit", async () => {
    const { owner, transport } = fixture();
    await owner.submit(token);
    transport.send.mockResolvedValue({
      kind: "rejected",
      failure: { kind: "stale_target", message: "Source removed" },
    });
    await owner.replay("note-create-1");
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
    expect(owner.busy).toBe(true);
    owner.changeSource(null);
    owner.update("note.title", "duplicate risk");
    await owner.submit(token);
    expect(transport.send).toHaveBeenCalledTimes(2);
    expect(owner.getSnapshot().draft?.source?.recordId).toBe(sourceId);
    expect(owner.getSnapshot().draft?.values["note.title"]).toBe(
      "Retained Note",
    );
  });
  it("sends linked replay as POST with exact body and validates a complete HTTP receipt", async () => {
    const { owner, receipt } = fixture();
    const draft = required(owner.getSnapshot().draft);
    const transport = createNoteCreateTransport("/api-root");
    const attempt = transport.capture({ authority, draft }, "txn-http");
    const fetch = vi
      .fn()
      .mockRejectedValueOnce(new Error("response lost"))
      .mockResolvedValue(
        new Response(JSON.stringify(receipt), {
          status: 200,
          headers: {
            "Content-Type": "application/json",
            "X-Request-ID": "request-1",
          },
        }),
      );
    vi.stubGlobal("fetch", fetch);
    expect(await transport.send(attempt, new AbortController().signal)).toEqual(
      { kind: "uncertain" },
    );
    expect(
      (await transport.send(attempt, new AbortController().signal)).kind,
    ).toBe("accepted");
    expect(fetch).toHaveBeenCalledTimes(2);
    for (const call of fetch.mock.calls) {
      expect(call[0]).toContain(`/records/${sourceId}/linked-notes`);
      expect(call[1]).toMatchObject({ method: "POST", body: attempt.body });
    }
    const wrongSource = {
      ...receipt,
      data: { ...receipt.data, source_record_id: noteId },
    };
    fetch.mockResolvedValue(
      new Response(JSON.stringify(wrongSource), {
        status: 200,
        headers: {
          "Content-Type": "application/json",
          "X-Request-ID": "request-1",
        },
      }),
    );
    expect(await transport.send(attempt, new AbortController().signal)).toEqual(
      { kind: "uncertain" },
    );
  });
});
function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Required fixture value");
  return value;
}
