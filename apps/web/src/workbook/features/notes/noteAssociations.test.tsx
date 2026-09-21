import { requireViewContract } from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createNoteAssociationTransport } from "../../adapters/createNoteAssociationTransport";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookSourceWriteSettlement } from "../../ports/WorkbookSourceWriteCoordination";
import { NoteAssociationPanel } from "./NoteAssociationPanel";
import {
  type NoteAssociationOutcome,
  type NoteAssociationReader,
  type NoteAssociationReceipt,
  type NoteAssociationTransport,
  noteAssociationListKey,
  noteAssociationView,
} from "./noteAssociationOperation";
import { WorkbookNoteAssociationOwner } from "./WorkbookNoteAssociationOwner";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
const authority: WorkbookMutationAuthority = {
  actorId: "10000000-0000-4000-8000-000000000001",
  incidentId: "40000000-0000-4000-8000-000000000004",
  sessionIdentity: "session",
  role: "editor",
  closed: false,
};
const noteId = "30000000-0000-4000-8000-000000000003",
  otherId = "20000000-0000-4000-8000-000000000002";
function fixture() {
  const row = fullWorkbookViewRow(
    requireViewContract(noteAssociationView),
    noteId,
    1,
    { "note.title": "Current Note" },
  );
  const receipt: NoteAssociationReceipt = {
    data: {
      view_schema_id: noteAssociationView,
      row: { ...row, row_version: 2 },
      change_set_id: "50000000-0000-4000-8000-000000000005",
    },
    meta: { request_id: "association-request" },
  };
  const effects = {
    coordinate: vi.fn<() => Promise<WorkbookSourceWriteSettlement>>(
      async () => ({ kind: "settled", minimumRowVersion: 0 }),
    ),
    accepted: vi.fn(),
    refresh: vi.fn(async () => {}),
  };
  let sequence = 0;
  const ids = { create: vi.fn(() => `association-${++sequence}`) },
    owner = new WorkbookNoteAssociationOwner(
      authority.incidentId,
      ids,
      effects,
    );
  const reader: NoteAssociationReader = {
    verify: vi.fn(async () => {}),
    availableViews: vi.fn(async () => ({
      kind: "accepted" as const,
      value: [noteAssociationView],
    })),
    page: vi.fn(async () => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          {
            recordId: noteId,
            displayText: "Current Note",
            viewSchemaId: noteAssociationView,
            row,
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    })),
  };
  const transport = {
    ...createNoteAssociationTransport("/original-api"),
    send: vi.fn<NoteAssociationTransport["send"]>(async () => ({
      kind: "uncertain",
    })),
    list: vi.fn<NoteAssociationTransport["list"]>(async (_id, kind) => ({
      kind: "accepted",
      page: {
        note_record_id: noteId,
        row_version: 2,
        kind,
        items: [],
        next_cursor_token: null,
      },
    })),
  };
  const authorityReader = vi.fn(
    async (): Promise<WorkbookMutationAuthority> => authority,
  );
  owner.configure(reader, authorityReader, transport);
  owner.setAuthority(authority);
  const input = {
    row,
    kind: "related_note" as const,
    actions: [{ op: "add" as const, counterpart_record_id: otherId }],
    sheetRef: { kind: "view_schema" as const, id: noteAssociationView },
    label: "Current Note",
  };
  return {
    owner,
    ids,
    effects,
    reader,
    transport,
    authorityReader,
    row,
    receipt,
    input,
  };
}

describe("Note association lifetime", () => {
  it("replays captured bytes and retains accepted receipts through read-only refresh recovery", async () => {
    const f = fixture();
    await f.owner.submit(f.input);
    const first = f.owner.getSnapshot().entries[0];
    if (!first) throw new Error("Missing attempt");
    expect(first.phase).toBe("uncertain");
    expect(f.owner.blocksRecord(noteId)).toBe(true);
    await f.owner.submit(f.input);
    expect(f.transport.send).toHaveBeenCalledOnce();
    f.effects.refresh.mockRejectedValueOnce(new Error("read failed"));
    f.transport.send.mockResolvedValue({
      kind: "accepted",
      receipt: f.receipt,
    });
    await f.owner.replay(first.attempt.clientTxnId);
    await waitFor(() =>
      expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("required"),
    );
    expect(f.transport.send.mock.calls[1]?.[0]).toBe(first.attempt);
    expect(f.owner.unsettledMutationCount).toBe(0);
    expect(f.owner.latestRow(noteId)?.row_version).toBe(2);
    await f.owner.retryRefresh(first.attempt.clientTxnId);
    expect(f.transport.send).toHaveBeenCalledTimes(2);
    expect(f.owner.getSnapshot().entries[0]?.refresh).toBe("complete");
    expect(f.effects.accepted).toHaveBeenCalledOnce();
    expect(f.ids.create).toHaveBeenCalledOnce();

    const later = fixture();
    later.effects.refresh.mockRejectedValueOnce(
      new Error("first refresh failed"),
    );
    later.transport.send.mockResolvedValue({
      kind: "accepted",
      receipt: later.receipt,
    });
    await later.owner.submit(later.input);
    await waitFor(() =>
      expect(later.owner.getSnapshot().entries[0]?.refresh).toBe("required"),
    );
    later.transport.send.mockImplementation(async () => {
      later.transport.list.mockImplementation(async (_id, kind) => ({
        kind: "accepted",
        page: {
          note_record_id: noteId,
          row_version: 3,
          kind,
          items: [],
          next_cursor_token: null,
        },
      }));
      return {
        kind: "accepted",
        receipt: {
          ...later.receipt,
          data: {
            ...later.receipt.data,
            row: { ...later.row, row_version: 3 },
          },
        },
      };
    });
    vi.mocked(later.reader.page).mockResolvedValue({
      kind: "accepted",
      value: {
        candidates: [
          {
            recordId: noteId,
            viewSchemaId: noteAssociationView,
            displayText: "Current Note",
            row: { ...later.row, row_version: 2 },
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    });
    await later.owner.submit({
      ...later.input,
      row: { ...later.row, row_version: 2 },
    });
    await waitFor(() =>
      expect(
        later.owner.getSnapshot().entries.map((entry) => entry.refresh),
      ).toEqual(["complete", "complete"]),
    );
    expect(later.transport.send).toHaveBeenCalledTimes(2);
  });
  it("conceals late acceptance while suspended and retires account-replaced attempts", async () => {
    for (const replace of [false, true]) {
      const f = fixture(),
        pending = deferred<NoteAssociationOutcome>();
      f.transport.send.mockReturnValue(pending.promise);
      const submitting = f.owner.submit(f.input);
      await waitFor(() => expect(f.transport.send).toHaveBeenCalledOnce());
      if (replace)
        f.owner.setAuthority({ ...authority, actorId: "replacement" });
      else f.owner.suspend();
      pending.resolve({ kind: "accepted", receipt: f.receipt });
      await submitting;
      expect(f.owner.getSnapshot().entries).toEqual([]);
      expect(f.effects.accepted).not.toHaveBeenCalled();
      f.owner.setAuthority(authority);
      expect(f.owner.getSnapshot().entries).toHaveLength(replace ? 0 : 1);
      if (!replace) {
        await f.owner.retryRefresh("association-1");
        expect(f.effects.accepted).toHaveBeenCalledOnce();
      }
    }
  });
  it("rejects stale preparation and fences late list responses after authority loss", async () => {
    const f = fixture();
    f.effects.coordinate.mockResolvedValue({
      kind: "settled",
      minimumRowVersion: 3,
    });
    await f.owner.submit(f.input);
    expect(f.transport.send).not.toHaveBeenCalled();
    expect(
      f.owner.getSnapshot().errors[
        noteAssociationListKey(noteId, "related_note")
      ],
    ).toContain("changed");
    const pending =
      deferred<Awaited<ReturnType<NoteAssociationTransport["list"]>>>();
    f.transport.list.mockReturnValue(pending.promise);
    const read = f.owner.read(noteId, "source");
    f.owner.suspend();
    pending.resolve({
      kind: "accepted",
      page: {
        note_record_id: noteId,
        row_version: 2,
        kind: "source",
        items: [],
        next_cursor_token: null,
      },
    });
    await read;
    expect(f.owner.getSnapshot().lists).toEqual({});
  });
  it("preserves readable data on refresh failure and conceals it on incident revocation", async () => {
    const f = fixture(),
      key = noteAssociationListKey(noteId, "source");
    await f.owner.read(noteId, "source");
    f.transport.list.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Retry read" },
    });
    await f.owner.read(noteId, "source");
    expect(f.owner.getSnapshot().lists[key]).toMatchObject({
      state: "stale_failure",
      page: { row_version: 2 },
    });
    f.transport.list.mockResolvedValueOnce({
      kind: "accepted",
      page: {
        note_record_id: noteId,
        row_version: 1,
        kind: "source",
        items: [],
        next_cursor_token: null,
      },
    });
    await f.owner.read(noteId, "source");
    expect(f.owner.getSnapshot().lists[key]).toMatchObject({
      state: "stale_failure",
      page: { row_version: 2 },
    });
    f.transport.list.mockResolvedValueOnce({
      kind: "rejected",
      failure: {
        kind: "stale_target",
        publicCode: "incident_not_found",
        message: "Unavailable",
      },
    });
    await f.owner.read(noteId, "source");
    expect(f.owner.getSnapshot().authority).toBeNull();
    expect(f.owner.getSnapshot().lists).toEqual({});
  });
  it("navigates incoming related Notes without exposing a removal action", async () => {
    const f = fixture(),
      navigate = vi.fn(async () => {});
    f.transport.list.mockResolvedValue({
      kind: "accepted",
      page: {
        note_record_id: noteId,
        row_version: 1,
        kind: "related_note",
        next_cursor_token: null,
        items: [
          {
            item_ref: "opaque",
            counterpart_record_id: otherId,
            view_schema_id: noteAssociationView,
            display_label: "Referring Note",
            direction: "incoming",
          },
        ],
      },
    });
    render(
      <NoteAssociationPanel
        owner={f.owner}
        row={f.row}
        kind="related_note"
        sheetRef={f.input.sheetRef}
        label="Current Note"
        onNavigateNote={navigate}
      />,
    );
    await screen.findByRole("button", { name: "Referring Note" });
    expect(
      screen.getByRole("region", { name: "Referenced by" }),
    ).not.toBeNull();
    expect(screen.queryByRole("button", { name: /Remove/ })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Referring Note" }));
    expect(navigate).toHaveBeenCalledWith(otherId);
  });
  it("accepts a public no-op receipt and rejects contradictory mutation versions", async () => {
    const f = fixture(),
      transport = createNoteAssociationTransport("/api-root");
    const attempt = transport.capture({ ...f.input, authority }, "noop");
    const noop: NoteAssociationReceipt = {
      data: { view_schema_id: noteAssociationView, row: f.row },
      meta: { request_id: "noop-request" },
    };
    const fetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(noop), {
        status: 200,
        headers: {
          "Content-Type": "application/json",
          "X-Request-ID": "noop-request",
        },
      }),
    );
    vi.stubGlobal("fetch", fetch);
    expect(
      (await transport.send(attempt, new AbortController().signal)).kind,
    ).toBe("accepted");
    fetch.mockResolvedValue(
      new Response(
        JSON.stringify({
          ...noop,
          data: { ...noop.data, row: { ...f.row, row_version: 2 } },
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );
    expect(
      (await transport.send(attempt, new AbortController().signal)).kind,
    ).toBe("uncertain");
    expect(fetch.mock.calls[0]?.[0]).toBe(
      `/api-root/api/v1/records/${noteId}/note-associations`,
    );
    expect(fetch.mock.calls[0]?.[1].body).toBe(attempt.body);
  });
});
