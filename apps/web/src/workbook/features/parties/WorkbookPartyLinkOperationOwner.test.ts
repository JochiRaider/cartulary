import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { taskAuthority } from "../../../testing/taskWorkbookTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import {
  createPartyCreationTransport,
  type PartyCreationReceipt,
  type PartyCreationTransport,
} from "../../adapters/createPartyCreationTransport";
import type { RecordPatchTransport } from "../../adapters/workbookRecordPatchTransport";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { WorkbookExplicitPatchOwner } from "../../runtime/WorkbookExplicitPatchOwner";
import {
  type PartyPair,
  type PartyReview,
  partyCreateDraft,
  partyPairs,
  partyViewId,
} from "./partyLinkModel";
import { useGenericPartyLinkWorkflow } from "./useGenericPartyLinkWorkflow";
import { WorkbookPartyLinkOperationOwner } from "./WorkbookPartyLinkOperationOwner";

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
function fixture(pair: PartyPair = partyPairs[0]) {
  const sourceId = "00000000-0000-4000-8000-000000000410";
  let source: WorkbookQueryRow = fullWorkbookViewRow(
    requireViewContract(pair.viewSchemaId),
    sourceId,
    7,
    {
      [pair.textFieldKey]: "  Original source wording  ",
      [pair.refFieldKey]: null,
    },
  );
  const review: PartyReview = {
    authority: taskAuthority,
    pair,
    source: structuredClone(source),
    sheetRef: { kind: "view_schema", id: pair.viewSchemaId },
    sourceLabel: "Original source",
    presentation: "original-row-field-sheet",
  };
  const receipt: PartyCreationReceipt = {
    data: {
      view_schema_id: partyViewId,
      change_set_id: "00000000-0000-4000-8000-000000000501",
      row: fullWorkbookViewRow(
        requireViewContract(partyViewId),
        "00000000-0000-4000-8000-000000000411",
        1,
        { "party.display_name": "Reviewed name", "party.party_kind": "team" },
      ),
    },
    meta: { request_id: "create-receipt" },
  };
  const create = vi.fn(
    (prefix: string) => `${prefix}-${create.mock.calls.length}`,
  );
  const coordinate = vi.fn(async () => true),
    refresh = vi.fn(async () => {}),
    remember = vi.fn(),
    settle = vi.fn(),
    registerConflict = vi.fn(),
    accessLost = vi.fn();
  const patches = new WorkbookExplicitPatchOwner(
    taskAuthority.incidentId,
    { create },
    { coordinate, remember, settle, accepted: vi.fn(), registerConflict },
    100,
  );
  const patchSend = vi.fn<RecordPatchTransport["send"]>(async (request) => {
    const body = JSON.parse(request.body);
    source = {
      ...source,
      row_version: 8,
      cells: {
        ...source.cells,
        ...Object.fromEntries(
          body.changes.map((change: { field_key: string; value: unknown }) => [
            change.field_key,
            { value: change.value },
          ]),
        ),
      },
    };
    return {
      kind: "acknowledged",
      receipt: {
        viewSchemaId: pair.viewSchemaId,
        row: { ...structuredClone(source), view_schema_id: pair.viewSchemaId },
        changeSetId: "patch-receipt",
      },
    };
  });
  patches.configure({ send: patchSend }, accessLost);
  patches.setAuthority(taskAuthority);
  patches.observeQuery(source);
  const owner = new WorkbookPartyLinkOperationOwner(
    taskAuthority.incidentId,
    { create },
    patches,
    { coordinate, refresh, remember, settle },
  );
  const send = vi.fn<PartyCreationTransport["send"]>(async () => ({
    kind: "accepted",
    receipt,
  }));
  const read = vi.fn(async (view: string) =>
    view === partyViewId ? receipt.data.row : source,
  );
  const recheck = vi.fn(async () => taskAuthority);
  owner.configure(
    { ...createPartyCreationTransport(undefined), send },
    {
      source: read,
      page: async () => ({ rows: [], hasMore: false, nextCursor: null }),
    },
    recheck,
  );
  owner.setAuthority(taskAuthority);
  owner.setPresentation(review.presentation);
  const draft = {
    ...partyCreateDraft("Original source wording"),
    "party.display_name": "Reviewed name",
    "party.party_kind": "team",
  };
  return {
    owner,
    patches,
    review,
    receipt,
    draft,
    send,
    patchSend,
    create,
    coordinate,
    refresh,
    read,
    recheck,
    settle,
    accessLost,
    source: () => source,
    setSource: (row: WorkbookQueryRow) => {
      source = row;
    },
  };
}
afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

it("retains both Party receipts and exact independent source changes for all three pairs", async () => {
  for (const pair of partyPairs) {
    const f = fixture(pair);
    await f.owner.create(f.review, f.draft);
    expect(f.send).toHaveBeenCalledTimes(1);
    expect(f.patchSend).toHaveBeenCalledTimes(1);
    expect(f.source().cells[pair.textFieldKey]?.value).toBe(
      "  Original source wording  ",
    );
    expect(
      JSON.parse(f.patchSend.mock.calls[0]?.[0].body ?? "{}"),
    ).toMatchObject({
      base_row_version: 7,
      changes: [
        { field_key: pair.refFieldKey, value: f.receipt.data.row.record_id },
      ],
    });
    expect(f.owner.getSnapshot().creations[0]?.receipt).toEqual(f.receipt);
    expect(f.owner.getSnapshot().patches[0]?.receipt?.changeSetId).toBe(
      "patch-receipt",
    );
    expect(f.owner.getSnapshot().patches[0]?.reconciliation).toBe("complete");
    expect(f.send.mock.calls[0]?.[0].id).not.toBe(
      f.patchSend.mock.calls[0]?.[0].clientTxnId,
    );
    for (const [action, changes] of [
      ["clear_link", [{ field_key: pair.refFieldKey, value: null }]],
      ["clear_text", [{ field_key: pair.textFieldKey, value: null }]],
      [
        "clear_both",
        [
          { field_key: pair.textFieldKey, value: null },
          { field_key: pair.refFieldKey, value: null },
        ],
      ],
    ] as const) {
      const cleared = fixture(pair);
      await cleared.owner.patch(cleared.review, action);
      expect(
        JSON.parse(cleared.patchSend.mock.calls[0]?.[0].body ?? "{}").changes,
      ).toEqual(changes);
      expect(cleared.send).not.toHaveBeenCalled();
    }
  }
});

it("reserves Party creation and source patch before coordination and fences stale reviews", async () => {
  const f = fixture(),
    gate = deferred<boolean>();
  f.coordinate.mockReturnValue(gate.promise);
  const first = f.owner.create(f.review, f.draft);
  await f.owner.create(f.review, f.draft);
  expect(f.create).toHaveBeenCalledTimes(1);
  f.owner.setPresentation("new-row-field-sheet");
  gate.resolve(true);
  await first;
  expect(f.send).not.toHaveBeenCalled();
  expect(f.owner.getSnapshot().creations[0]?.phase).toBe("preparation_failed");
  const patch = fixture(),
    patchGate = deferred<boolean>();
  patch.coordinate.mockReturnValue(patchGate.promise);
  const pending = patch.owner.patch(patch.review, "clear_both");
  await patch.owner.patch(patch.review, "clear_both");
  expect(patch.create).toHaveBeenCalledTimes(1);
  patch.patches.acceptVersion(patch.review.source.record_id, 9);
  patchGate.resolve(true);
  await pending;
  expect(patch.patchSend).not.toHaveBeenCalled();
  expect(patch.patches.getSnapshot().entries[0]?.request?.baseRowVersion).toBe(
    7,
  );
});

it("retains accepted Party creation after navigation and rejected linking without repeating creation", async () => {
  const f = fixture(),
    gate = deferred<Awaited<ReturnType<PartyCreationTransport["send"]>>>();
  f.send.mockReturnValue(gate.promise);
  const pending = f.owner.create(f.review, f.draft);
  await vi.waitFor(() => expect(f.send).toHaveBeenCalledTimes(1));
  f.owner.setPresentation("another-sheet");
  gate.resolve({ kind: "accepted", receipt: f.receipt });
  await pending;
  expect(f.patchSend).not.toHaveBeenCalled();
  const saved = f.owner.getSnapshot().creations[0];
  expect(saved?.receipt).toEqual(f.receipt);
  const current = { ...f.review, presentation: "review-returned-source" };
  f.owner.setPresentation(current.presentation);
  f.patchSend.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "stale_target", message: "Source changed" },
  });
  await f.owner.linkCreated(saved?.id ?? "", current);
  expect(f.owner.getSnapshot().patches[0]?.phase).toBe("rejected");
  await f.owner.linkCreated(saved?.id ?? "", current);
  expect(f.send).toHaveBeenCalledTimes(1);
  expect(f.patchSend).toHaveBeenCalledTimes(2);
  expect(f.owner.getSnapshot().patches[1]?.receipt).not.toBeNull();
});

it("replays uncertain creation and linking with their captured identities after selection changes", async () => {
  const f = fixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  await f.owner.create(f.review, f.draft);
  const id = f.owner.getSnapshot().creations[0]?.id ?? "";
  f.draft["party.display_name"] = "Unreviewed new input";
  f.owner.setPresentation("different-pair");
  await f.owner.replayCreation(id);
  expect(f.send.mock.calls[1]?.[0]).toBe(f.send.mock.calls[0]?.[0]);
  expect(f.send.mock.calls[1]?.[0].draft["party.display_name"]).toBe(
    "Reviewed name",
  );
  expect(f.patchSend).not.toHaveBeenCalled();
  f.owner.setPresentation(f.review.presentation);
  f.patchSend.mockResolvedValueOnce({ kind: "uncertain" });
  await f.owner.linkCreated(id, f.review);
  const linkId = f.owner.getSnapshot().creations[0]?.linkId ?? "";
  f.owner.setPresentation("another-row");
  await f.patches.replay(linkId);
  expect(f.patchSend.mock.calls[1]?.[0]).toBe(f.patchSend.mock.calls[0]?.[0]);
  expect(f.patches.getSnapshot().entries[0]?.receipt?.changeSetId).toBe(
    "patch-receipt",
  );
  expect(f.create).toHaveBeenCalledTimes(2);
});

it("recovers accepted refresh failures through reads and preserves newer remote source versions", async () => {
  const f = fixture();
  f.refresh.mockRejectedValue(new Error("projection unavailable"));
  await f.owner.create(f.review, f.draft);
  const creation = f.owner.getSnapshot().creations[0],
    patch = f.owner.getSnapshot().patches[0];
  expect(creation?.refresh).toBe("required");
  expect(patch?.reconciliation).toBe("required");
  const remote = {
    ...f.source(),
    row_version: 12,
    cells: {
      ...f.source().cells,
      [f.review.pair.textFieldKey]: { value: "Newer remote text" },
    },
  };
  f.patches.acceptRow(remote);
  f.refresh.mockResolvedValue();
  await f.patches.refresh(patch?.id ?? "");
  expect(f.patches.getSnapshot().entries[0]?.reconciliation).toBe("required");
  f.setSource(remote);
  await f.patches.refresh(patch?.id ?? "");
  await f.owner.refreshCreation(creation?.id ?? "");
  expect(f.patches.getSnapshot().entries[0]?.reconciliation).toBe("complete");
  expect(f.patches.latestRow(remote.record_id)).toEqual(remote);
  expect(f.patches.getSnapshot().entries[0]?.receipt?.row.row_version).toBe(8);
  expect(f.patchSend).toHaveBeenCalledTimes(1);
  expect(f.send).toHaveBeenCalledTimes(1);
});

it("retains late accepted source receipts while hiding suspended and retired Party recovery", async () => {
  vi.useFakeTimers();
  const f = fixture(),
    gate = deferred<Awaited<ReturnType<RecordPatchTransport["send"]>>>();
  f.patchSend.mockReturnValue(gate.promise);
  const pending = f.owner.patch(f.review, "clear_both");
  await vi.advanceTimersByTimeAsync(101);
  await pending;
  expect(f.owner.getSnapshot().patches[0]?.phase).toBe("uncertain");
  const row = { ...f.source(), row_version: 8 };
  f.setSource(row);
  f.owner.setPresentation("different-row");
  gate.resolve({
    kind: "acknowledged",
    receipt: {
      row: { ...row, view_schema_id: f.review.pair.viewSchemaId },
      changeSetId: "late-receipt",
      viewSchemaId: f.review.pair.viewSchemaId,
    },
  });
  await vi.advanceTimersByTimeAsync(1);
  expect(f.owner.getSnapshot().patches[0]?.receipt?.changeSetId).toBe(
    "late-receipt",
  );
  f.patches.suspend();
  f.owner.suspend();
  expect(f.owner.getSnapshot().patches).toEqual([]);
  f.patches.setAuthority({ ...taskAuthority, role: "viewer" });
  f.owner.setAuthority({ ...taskAuthority, role: "viewer" });
  expect(f.owner.canSubmit()).toBe(false);
  expect(f.owner.getSnapshot().patches[0]?.receipt).not.toBeNull();
  f.patches.setAuthority({ ...taskAuthority, actorId: "another-account" });
  f.owner.setAuthority({ ...taskAuthority, actorId: "another-account" });
  expect(f.accessLost).not.toHaveBeenCalled();
});

it("rechecks Party authority without reporting target rejection as incident access loss", async () => {
  const f = fixture();
  f.patchSend.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "authorization_lost", message: "Recheck permission" },
  });
  await f.owner.patch(f.review, "clear_link");
  expect(f.recheck).toHaveBeenCalledTimes(2);
  expect(f.accessLost).not.toHaveBeenCalled();
  expect(f.accessLost).not.toHaveBeenCalled();
});

it("keeps a selected pair review actionable and detaches callbacks when another pair is selected", async () => {
  const f = fixture();
  const { result } = renderHook(() =>
    useGenericPartyLinkWorkflow({
      owner: f.owner,
      contract: requireViewContract(f.review.pair.viewSchemaId),
      row: f.review.source,
      sourceLabel: "Original source",
      sheetRef: f.review.sheetRef,
      resetKey: "inspector",
      fieldKey: "evidence.title",
      visible: true,
    }),
  );
  const original = result.current.review;
  if (!original) throw new Error("Review required");
  act(() => result.current.selectPair(original.pair.key));
  expect(f.owner.isCurrent(result.current.review ?? original)).toBe(true);
  act(() => result.current.selectPair(partyPairs[1].key));
  expect(f.owner.isCurrent(original)).toBe(false);
  await f.owner.create(original, f.draft);
  expect(f.send).not.toHaveBeenCalled();
  const current = result.current.review;
  if (!current) throw new Error("Review required");
  await act(() => f.owner.create(current, f.draft));
  expect(f.send).toHaveBeenCalledTimes(1);
});

it("retains late accepted creation while authority changes and never links under obsolete review", async () => {
  vi.useFakeTimers();
  const f = fixture(),
    gate = deferred<Awaited<ReturnType<PartyCreationTransport["send"]>>>();
  f.send.mockReturnValue(gate.promise);
  const pending = f.owner.create(f.review, f.draft);
  await vi.advanceTimersByTimeAsync(30_001);
  await pending;
  expect(f.owner.getSnapshot().creations[0]?.phase).toBe("uncertain");
  f.owner.suspend();
  f.patches.suspend();
  gate.resolve({ kind: "accepted", receipt: f.receipt });
  await vi.advanceTimersByTimeAsync(1);
  expect(f.owner.getSnapshot().creations).toEqual([]);
  f.patches.setAuthority(taskAuthority);
  f.owner.setAuthority(taskAuthority);
  expect(f.owner.getSnapshot().creations[0]?.receipt).toEqual(f.receipt);
  expect(f.owner.getSnapshot().creations[0]?.refresh).toBe("required");
  expect(f.patchSend).not.toHaveBeenCalled();
  await f.owner.refreshCreation(f.owner.getSnapshot().creations[0]?.id ?? "");
  expect(f.owner.getSnapshot().creations[0]?.refresh).toBe("complete");
});

it("rejects current viewer and closed incident authority before fresh Party dispatch or replay", async () => {
  for (const authority of [
    { ...taskAuthority, role: "viewer" as const },
    { ...taskAuthority, closed: true },
  ]) {
    const fresh = fixture();
    fresh.recheck.mockResolvedValue(authority);
    await fresh.owner.create(fresh.review, fresh.draft);
    expect(fresh.send).not.toHaveBeenCalled();
    expect(fresh.patchSend).not.toHaveBeenCalled();
    const replay = fixture();
    replay.send.mockResolvedValueOnce({ kind: "uncertain" });
    await replay.owner.create(replay.review, replay.draft);
    replay.recheck.mockResolvedValue(authority);
    await replay.owner.replayCreation(
      replay.owner.getSnapshot().creations[0]?.id ?? "",
    );
    expect(replay.send).toHaveBeenCalledTimes(1);
    expect(replay.patchSend).not.toHaveBeenCalled();
  }
});

it("unavailable Party targets fail preparation without source writes or incident access loss", async () => {
  const f = fixture();
  f.read.mockImplementation(async (view) => {
    if (view === partyViewId) throw new Error("Unavailable Party");
    return f.source();
  });
  await f.owner.patch(f.review, "link", "unavailable-party");
  expect(f.patchSend).not.toHaveBeenCalled();
  expect(f.accessLost).not.toHaveBeenCalled();
  expect(f.owner.getSnapshot().patches[0]?.phase).toBe("preparation_failed");
  expect(f.owner.canSubmit()).toBe(true);
});
