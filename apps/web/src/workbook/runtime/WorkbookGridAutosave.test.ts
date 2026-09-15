import { describe, expect, it, vi } from "vitest";
import { timelineRow } from "../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";
const fieldKey = "timeline.activity_synopsis_text";

function fixture() {
  let sequence = 0;
  const replies: Array<
    (value: Awaited<ReturnType<WorkbookPendingMutationPort["execute"]>>) => void
  > = [];
  const execute = vi.fn<WorkbookPendingMutationPort["execute"]>(
    () => new Promise((resolve) => replies.push(resolve)),
  );
  const runtime = new WorkbookMutationRuntime(
    { incidentId, clientInstanceId: "grid-autosave-client" },
    { create: () => `grid-autosave-${++sequence}` },
    { execute },
  );
  runtime.explicitPatches.setAuthority({
    actorId: "actor",
    sessionIdentity: "session",
    incidentId,
    role: "editor",
    closed: false,
  });
  let saved: WorkbookQueryRow = timelineRow({
    captureState: "rough",
    recordId,
    rowVersion: 1,
  });
  const initial = structuredClone(saved);
  runtime.explicitPatches.acceptRow(saved);
  const identity = (field = fieldKey) => ({
    viewSchemaId: timelineViewSchemaId,
    recordId,
    fieldKey: field,
  });
  const author = (value: string, field = fieldKey) =>
    runtime.gridDrafts.update(identity(field), saved, value);
  const edit = (
    value: string,
    field = fieldKey,
    dependencies: readonly string[] = [],
  ) => {
    author(value, field);
    return runtime.enqueuePatch({
      baseline: initial,
      dependencies,
      baseRowVersion: 1,
      changes: [{ field_key: field, value }],
      fieldKey: field,
      localValue: value,
      recordId,
      rowLabel: "Event",
      surfaceLabel: "Timeline",
      viewSchemaId: timelineViewSchemaId,
    });
  };
  const accept = (index: number, version: number) => {
    const unit = execute.mock.calls[index]?.[0].unit;
    const cells = { ...saved.cells };
    if (unit?.identity.kind === "patch")
      for (const change of unit.identity.changes)
        cells[change.field_key] = { value: change.value };
    const row = { ...saved, row_version: version, cells };
    saved = row;
    replies[index]?.({
      kind: "accepted",
      value: {
        changeSetId: `30000000-0000-4000-8000-${String(version).padStart(12, "0")}`,
        viewSchemaId: timelineViewSchemaId,
        row: { ...row, view_schema_id: timelineViewSchemaId },
      },
    });
  };
  const observe = (version: number, fields: Record<string, unknown>) => {
    saved = {
      ...saved,
      row_version: version,
      cells: {
        ...saved.cells,
        ...Object.fromEntries(
          Object.entries(fields).map(([field, value]) => [field, { value }]),
        ),
      },
    };
    runtime.explicitPatches.observeQuery(saved);
  };
  const pause = () => runtime.pendingQueue().model.pauseForAuthRecovery();
  const resume = () => {
    runtime.pendingQueue().model.resumeAfterAuthRecovery();
    runtime.requestDrain();
  };
  const reject = (
    index: number,
    kind: "retryable" | "validation" | "authorization_lost",
  ) =>
    replies[index]?.({
      kind: "rejected",
      failure: { kind, message: "Test rejection" },
    });
  return {
    runtime,
    execute,
    edit,
    accept,
    author,
    identity,
    observe,
    pause,
    resume,
    reject,
    initial,
  };
}

describe("Committed grid autosave", () => {
  it("does not report queue admission as authoritative editor acceptance", () => {
    const f = fixture();
    f.runtime.invalidate({ kind: "session_unavailable" });
    expect(f.edit("A").kind).not.toBe("accepted");
    expect(f.execute).not.toHaveBeenCalled();
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("keeps a newer same-field edit when its predecessor is acknowledged", async () => {
    const f = fixture();
    f.edit("A");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    f.edit("B");
    f.accept(0, 2);
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(2));
    expect(
      f.runtime.visibleEdit(timelineViewSchemaId, recordId, fieldKey),
    ).toBe("B");
    f.accept(1, 3);
    await vi.waitFor(() =>
      expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(0),
    );
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("does not regress a newer authoring baseline when an older predecessor receipt arrives", async () => {
    const f = fixture();
    f.edit("A");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    f.observe(3, { [fieldKey]: "Remote after A" });
    const reviewed = f.runtime.explicitPatches.latestRow(recordId);
    if (!reviewed) throw new Error("Expected current saved row");
    f.runtime.gridDrafts.review(f.identity(), reviewed);
    const b = f.edit("B authored against three");
    f.accept(0, 2);
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(2));
    expect(f.execute.mock.calls[1]?.[0].committedRowVersion).toBe(3);
    f.accept(1, 4);
    if (b.kind !== "admitted") throw new Error("Expected admission");
    expect((await b.completion).kind).toBe("accepted");
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("prepares an undispatched dependent edit from the accepted predecessor version", async () => {
    const f = fixture();
    f.edit("A");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    f.edit("B", "timeline.analyst_text");
    f.accept(0, 2);
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(2));
    expect(f.execute.mock.calls[1]?.[0].committedRowVersion).toBe(2);
    f.accept(1, 3);
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("settles coalesced contributors without claiming superseded text was saved", async () => {
    const f = fixture();
    f.pause();
    const a = f.edit("A"),
      b = f.edit("B"),
      other = f.edit("Analyst", "timeline.analyst_text");
    if (
      a.kind !== "admitted" ||
      b.kind !== "admitted" ||
      other.kind !== "admitted"
    )
      throw new Error("Expected admissions");
    expect(new Set([a.unitId, b.unitId, other.unitId]).size).toBe(1);
    const settled = vi.fn();
    void b.completion.then(settled);
    f.resume();
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    expect(settled).not.toHaveBeenCalled();
    expect(f.execute.mock.calls[0]?.[0].unit.identity).toMatchObject({
      changes: [
        { field_key: fieldKey, value: "B" },
        { field_key: "timeline.analyst_text", value: "Analyst" },
      ],
    });
    f.accept(0, 2);
    expect(await a.completion).toMatchObject({ kind: "superseded" });
    expect(await b.completion).toEqual({ kind: "accepted" });
    expect(await other.completion).toEqual({ kind: "accepted" });
    expect(f.runtime.gridDrafts.read(f.identity())).toBeNull();
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("preserves newer unsubmitted equal-valued authoring through acknowledgement", async () => {
    const f = fixture();
    const a = f.edit("A");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    f.author("B");
    f.author("A");
    f.accept(0, 2);
    if (a.kind !== "admitted") throw new Error("Expected admission");
    await a.completion;
    expect(f.runtime.gridDrafts.read(f.identity())?.value).toBe("A");
    expect(
      f.runtime.gridDrafts.staleFields(
        f.identity(),
        f.runtime.explicitPatches.latestRow(recordId) ?? f.initial,
      ),
    ).toEqual([]);
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("prepares unrelated collaboration changes but halts changed dependencies for review", async () => {
    const f = fixture();
    f.pause();
    f.edit("A");
    f.observe(2, { "timeline.analyst_text": "Remote" });
    f.resume();
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    expect(f.execute.mock.calls[0]?.[0].committedRowVersion).toBe(2);
    f.accept(0, 3);
    await vi.waitFor(() =>
      expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(0),
    );
    f.pause();
    const b = f.edit("B", fieldKey, ["timeline.analyst_text"]);
    f.observe(4, { "timeline.analyst_text": "Changed dependency" });
    f.resume();
    if (b.kind !== "admitted") throw new Error("Expected admission");
    expect(await b.completion).toMatchObject({ kind: "stale_target" });
    expect(f.execute).toHaveBeenCalledTimes(1);
    expect(f.runtime.gridDrafts.read(f.identity())?.value).toBe("B");
    expect((await f.runtime.discardBlockedEdit()).ok).toBe(true);
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("replays the captured attempt after collaboration advances the record", async () => {
    const f = fixture();
    f.edit("A");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    const captured = structuredClone(f.execute.mock.calls[0]?.[0]);
    f.observe(9, { [fieldKey]: "Remote" });
    f.reject(0, "retryable");
    await vi.waitFor(() =>
      expect(f.runtime.pendingQueue().model.snapshot().inFlightCount).toBe(0),
    );
    f.runtime.requestDrain();
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(2));
    expect(f.execute.mock.calls[1]?.[0]).toEqual(captured);
    f.accept(1, 2);
    await vi.waitFor(() =>
      expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(0),
    );
    expect(f.runtime.explicitPatches.latestRow(recordId)?.row_version).toBe(9);
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("retains contiguous collaboration rows and refuses sparse gaps without altering captured replay", async () => {
    const f = fixture();
    const change = (version: number, value: string) =>
      f.runtime.explicitPatches.observeRecordChanged({
        record_id: recordId,
        row_version: version,
        client_txn_id: `remote-${version}`,
        change_set_id: `change-${version}`,
        actor_user_id: "other",
        changed_field_keys: [fieldKey],
        affected_views: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "patch",
            patch_cells: {
              record_id: recordId,
              row_version: version,
              cells: { [fieldKey]: { value } },
            },
          },
        ],
      });
    change(2, "Remote two");
    expect(
      f.runtime.explicitPatches.latestRow(recordId)?.cells[fieldKey]?.value,
    ).toBe("Remote two");
    change(4, "Missing predecessor");
    expect(f.runtime.explicitPatches.latestVersion(recordId)).toBe(4);
    expect(f.runtime.explicitPatches.latestRow(recordId)?.row_version).toBe(2);
    f.pause();
    const ticket = f.edit("Retained local");
    f.resume();
    if (ticket.kind !== "admitted") throw new Error("Expected local admission");
    expect((await ticket.completion).kind).toBe("stale_target");
    expect(f.execute).not.toHaveBeenCalled();
    expect(f.runtime.gridDrafts.read(f.identity())?.value).toBe(
      "Retained local",
    );
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("keeps acceptance and newer text when refresh fails and recovers by reads after remount", async () => {
    const f = fixture();
    const refresh = vi.fn(async () => {
      throw new Error("Read unavailable");
    });
    const detach = f.runtime.registerSurface(timelineViewSchemaId, refresh);
    const a = f.edit("A");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    f.author("Unsubmitted B");
    f.accept(0, 2);
    if (a.kind !== "admitted") throw new Error("Expected admission");
    expect(await a.completion).toEqual({ kind: "accepted" });
    await vi.waitFor(() => expect(refresh).toHaveBeenCalledTimes(1));
    expect(f.runtime.explicitPatches.latestReceipt(recordId)).toMatchObject({
      row: { row_version: 2, cells: { [fieldKey]: { value: "A" } } },
      changeSetId: expect.any(String),
    });
    f.runtime.explicitPatches.observeQuery(f.initial);
    expect(f.runtime.explicitPatches.latestRow(recordId)?.row_version).toBe(2);
    expect(f.runtime.surfaceRefreshRequired(timelineViewSchemaId)).toBe(true);
    await f.runtime.refreshSurface(timelineViewSchemaId);
    expect(refresh).toHaveBeenCalledTimes(2);
    expect(f.execute).toHaveBeenCalledTimes(1);
    detach();
    const recover = vi.fn(async () => {});
    f.runtime.registerSurface(timelineViewSchemaId, recover);
    await vi.waitFor(() => expect(recover).toHaveBeenCalledTimes(1));
    expect(f.runtime.gridDrafts.read(f.identity())?.value).toBe(
      "Unsubmitted B",
    );
    expect(f.execute).toHaveBeenCalledTimes(1);
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("preserves newer refused text when a full queue predecessor completes", async () => {
    const f = fixture();
    f.edit("A");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    for (let index = 0; index < 63; index++) {
      f.runtime.pendingQueue().model.sealPending();
      expect(f.edit(`Queued ${index}`).kind).toBe("admitted");
    }
    f.runtime.pendingQueue().model.sealPending();
    expect(f.edit("Refused").kind).toBe("rejected_mutation");
    f.accept(0, 2);
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(2));
    expect(
      f.runtime.visibleEdit(timelineViewSchemaId, recordId, fieldKey),
    ).toBe("Refused");
    expect(f.runtime.gridDrafts.read(f.identity())?.value).toBe("Refused");
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("requests current authorization after denial without retiring readable raw work", async () => {
    const f = fixture(),
      recover = vi.fn();
    const unbind = f.runtime.bindAuthorizationRecovery(recover);
    f.edit("Denied but retained");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    f.reject(0, "authorization_lost");
    await vi.waitFor(() => expect(recover).toHaveBeenCalledTimes(1));
    expect(f.runtime.retired).toBe(false);
    expect(f.runtime.gridDrafts.read(f.identity())?.value).toBe(
      "Denied but retained",
    );
    expect(f.runtime.getSnapshot().authPaused).toBe(true);
    unbind();
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("discards only the rejected operation while retaining newer raw and queued work", async () => {
    const f = fixture();
    f.edit("A");
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
    const b = f.edit("B");
    f.reject(0, "validation");
    await vi.waitFor(() =>
      expect(f.runtime.pendingQueue().model.snapshot().halted).not.toBeNull(),
    );
    expect((await f.runtime.discardBlockedEdit()).ok).toBe(true);
    await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(2));
    expect(
      f.runtime.visibleEdit(timelineViewSchemaId, recordId, fieldKey),
    ).toBe("B");
    expect(f.runtime.gridDrafts.read(f.identity())?.value).toBe("B");
    f.accept(1, 2);
    if (b.kind !== "admitted") throw new Error("Expected admission");
    expect(await b.completion).toEqual({ kind: "accepted" });
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });
});
