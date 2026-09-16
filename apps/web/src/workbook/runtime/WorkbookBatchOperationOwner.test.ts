import { describe, expect, it, vi } from "vitest";
import { WorkbookBatchOperationOwner } from "./WorkbookBatchOperationOwner";
import type {
  WorkbookBatchPlan,
  WorkbookBatchReceipt,
  WorkbookBatchTransport,
  WorkbookBatchTransportOutcome,
} from "./workbookBatchOperation";
import { workbookBatchRecoveryItems } from "./workbookBatchRecoveryItems";

function required<T>(value: T | null | undefined): T {
  if (value === undefined || value === null)
    throw new Error("Expected fixture value");
  return value;
}
const authority = {
  actorId: "actor",
  incidentId: "incident",
  sessionIdentity: "session",
  role: "editor" as const,
  closed: false,
};
function plan(recordId = "row"): WorkbookBatchPlan {
  return {
    operation: "applyWorkbookBulkMutation",
    recordIds: [recordId],
    request: {
      kind: "fill_down_v1",
      view_schema_id: "cartulary.view.timeline.v2",
      field_key: "timeline.activity_synopsis_text",
      value: "Captured",
      targets: [{ record_id: recordId, base_row_version: 2 }],
    },
  };
}
const receipt: WorkbookBatchReceipt = {
  viewSchemaId: "cartulary.view.timeline.v2",
  changeSetId: "change",
  rows: [{ record_id: "row", row_version: 3, cells: {} }],
  conflicts: [],
};
const accepted = { kind: "acknowledged" as const, receipt };
async function settle() {
  for (let i = 0; i < 12; i++) await Promise.resolve();
}
function setup() {
  let id = 0;
  const transport: WorkbookBatchTransport = {
    capture: vi.fn((plan, authority, id) => ({
      id,
      plan,
      authority,
      apiBase: undefined,
      path: "/captured",
      body: JSON.stringify({ ...plan.request, client_txn_id: id }),
    })),
    send: vi.fn(async () => accepted),
  };
  const coordination = {
    pending: vi.fn<
      () => readonly {
        id: string;
        recordId: string | null;
        viewSchemaId: string;
      }[]
    >(() => []),
    sealPending: vi.fn(),
    available: () => true,
    reserve: () => () => {},
    conflicts: vi.fn(() => false),
    accepted: vi.fn(),
    refresh: vi.fn(async () => {}),
  };
  const owner = new WorkbookBatchOperationOwner(
    "incident",
    { create: () => `batch-${++id}` },
    coordination,
  );
  owner.configure(transport);
  owner.setAuthority(authority);
  return { owner, transport, coordination };
}
describe("Workbook batch operation ownership", () => {
  it("projects one logical batch through uncertainty conflicts refresh and completion", async () => {
    const { owner, transport } = setup();
    vi.mocked(transport.send).mockResolvedValueOnce({ kind: "uncertain" });
    const id = required(owner.admit(plan(), { delivery: {} }));
    await settle();
    expect(
      workbookBatchRecoveryItems(owner.getSnapshot(), new Set()),
    ).toMatchObject([{ id, attention: "attention" }]);
    await owner.retry(id);
    await settle();
    expect(
      workbookBatchRecoveryItems(owner.getSnapshot(), new Set([id])),
    ).toMatchObject([{ id, attention: "attention" }]);
    expect(
      workbookBatchRecoveryItems(owner.getSnapshot(), new Set()),
    ).toMatchObject([{ id, attention: "completed" }]);
    owner.suspend();
    expect(workbookBatchRecoveryItems(owner.getSnapshot(), new Set())).toEqual(
      [],
    );
  });

  it("captures once, distinguishes delivery from repetition, and preserves original targets", async () => {
    const { owner, transport, coordination } = setup();
    vi.mocked(transport.send).mockResolvedValue({
      kind: "acknowledged",
      receipt: {
        ...receipt,
        rows: [{ record_id: "row", row_version: 5, cells: {} }],
      },
    });
    const original = plan();
    const earlierEdit = {
      id: "earlier-edit",
      recordId: "row",
      viewSchemaId: "cartulary.view.timeline.v2",
    };
    coordination.pending.mockReturnValue([earlierEdit]);
    let ready!: () => void;
    const admission = {
      delivery: {},
      ready: new Promise<void>((resolve) => {
        ready = resolve;
      }),
    };
    const id = owner.admit(original, admission);
    expect(owner.admit(original, admission)).toBe(id);
    expect(owner.allowsPending(earlierEdit)).toBe(true);
    owner.acceptPrerequisiteRow(earlierEdit.id, "row", 4);
    coordination.pending.mockReturnValue([]);
    const laterEdit = {
      id: "later-edit",
      recordId: "row",
      viewSchemaId: "cartulary.view.timeline.v2",
    };
    expect(owner.allowsPending(laterEdit)).toBe(false);
    expect(owner.allowsPending({ ...laterEdit, recordId: "unrelated" })).toBe(
      true,
    );
    owner.acceptPrerequisiteRow(laterEdit.id, "row", 50);
    Object.assign(original, { recordIds: ["new-selection"] });
    Object.assign(required(original.request.targets[0]), {
      base_row_version: 20,
    });
    ready();
    await settle();
    expect(transport.send).toHaveBeenCalledTimes(1);
    const attempt = required(required(owner.getSnapshot().entries[0]).attempt);
    expect(attempt.plan.recordIds).toEqual(["row"]);
    expect(JSON.parse(attempt.body).targets[0].base_row_version).toBe(4);
    expect(Object.isFrozen(attempt.plan.request.targets)).toBe(true);
    expect(owner.allowsPending(laterEdit)).toBe(true);
    owner.admit(plan(), { delivery: {} });
    await settle();
    expect(transport.send).toHaveBeenCalledTimes(2);
  });
  it("retains overlapping actions and explicitly replays the identical uncertain attempt", async () => {
    const { owner, transport } = setup();
    vi.mocked(transport.send).mockResolvedValueOnce({ kind: "uncertain" });
    const first = required(owner.admit(plan(), { delivery: {} }));
    owner.admit(plan(), { delivery: {} });
    owner.admit(plan("unrelated"), { delivery: {} });
    await settle();
    expect(transport.send).toHaveBeenCalledTimes(2);
    expect(
      owner.getSnapshot().entries.find((entry) => entry.id === first)?.phase,
    ).toBe("uncertain");
    const captured = required(vi.mocked(transport.send).mock.calls[0])[0];
    owner.retry(first);
    owner.retry(first);
    await settle();
    expect(required(vi.mocked(transport.send).mock.calls[2])[0]).toBe(captured);
    expect(transport.send).toHaveBeenCalledTimes(4);
    expect(transport.capture).toHaveBeenCalledTimes(3);
  });
  it("retains complete acceptance across failed reads and recovers with reads only", async () => {
    const { owner, transport, coordination } = setup();
    coordination.refresh.mockRejectedValueOnce(new Error("read failed"));
    const id = required(owner.admit(plan(), { delivery: {} }));
    await settle();
    expect(owner.getSnapshot().entries[0]).toMatchObject({
      receipt,
      phase: "acknowledged",
      reconciliation: "required",
    });
    owner.suspend();
    expect(owner.getSnapshot().entries).toEqual([]);
    owner.retry(id);
    expect(transport.send).toHaveBeenCalledTimes(1);
    owner.setAuthority({ ...authority, sessionIdentity: "reauthenticated" });
    owner.retry(id);
    await settle();
    expect(coordination.refresh).toHaveBeenCalledTimes(2);
    expect(transport.send).toHaveBeenCalledTimes(1);
    expect(required(owner.getSnapshot().entries[0]).reconciliation).toBe(
      "complete",
    );
  });
  it("fences late receipts after replacement and never treats replay denial as proof of no commit", async () => {
    const { owner, transport } = setup();
    vi.mocked(transport.send)
      .mockResolvedValueOnce({ kind: "uncertain" })
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "authorization_lost", message: "Access unavailable" },
      });
    const id = required(owner.admit(plan(), { delivery: {} }));
    await settle();
    owner.retry(id);
    await settle();
    owner.discard(id);
    expect(required(owner.getSnapshot().entries[0]).phase).toBe("uncertain");
    let finish!: (outcome: WorkbookBatchTransportOutcome) => void;
    vi.mocked(transport.send).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    owner.retry(id);
    owner.setAuthority({ ...authority, actorId: "replacement" });
    finish(accepted);
    await settle();
    expect(owner.getSnapshot().entries).toEqual([]);
  });
  it("retains conflicts-only acceptance and blocks only overlapping batches until resolution", async () => {
    const { owner, transport, coordination } = setup();
    coordination.conflicts.mockImplementation(
      (id?: string) => id === "batch-1",
    );
    vi.mocked(transport.send).mockResolvedValueOnce({
      kind: "acknowledged",
      receipt: { ...receipt, rows: [], changeSetId: null, conflicts: [] },
    });
    owner.admit(plan(), { delivery: {} });
    owner.admit(plan(), { delivery: {} });
    await settle();
    expect(
      required(owner.getSnapshot().entries[0]).receipt?.changeSetId,
    ).toBeNull();
    expect(transport.send).toHaveBeenCalledTimes(1);
    coordination.conflicts.mockReturnValue(false);
    owner.wake();
    await settle();
    expect(transport.send).toHaveBeenCalledTimes(2);
  });
  it("suspends queued work through role loss and closure without changing its targets", async () => {
    const { owner, transport } = setup();
    let ready!: () => void;
    owner.admit(plan(), {
      delivery: {},
      ready: new Promise<void>((resolve) => {
        ready = resolve;
      }),
    });
    owner.setAuthority({ ...authority, role: "viewer" });
    ready();
    await settle();
    expect(transport.send).not.toHaveBeenCalled();
    owner.setAuthority({ ...authority, closed: true });
    await settle();
    expect(transport.send).not.toHaveBeenCalled();
    owner.suspend();
    owner.retry("batch-1");
    await settle();
    expect(transport.send).not.toHaveBeenCalled();
    owner.setAuthority({ ...authority, sessionIdentity: "new-session" });
    await settle();
    expect(transport.send).toHaveBeenCalledOnce();
    expect(
      vi.mocked(transport.send).mock.calls[0]?.[0].plan.request.targets,
    ).toEqual([{ record_id: "row", base_row_version: 2 }]);
  });
  it("lets conflict correction pass later waiting batches and advances only undispatched versions", async () => {
    const { owner, transport, coordination } = setup();
    coordination.conflicts.mockReturnValue(true);
    const first = required(owner.admit(plan(), { delivery: {} }));
    const second = required(owner.admit(plan(), { delivery: {} }));
    await settle();
    expect(transport.send).toHaveBeenCalledOnce();
    expect(owner.blocksRecord("row", first)).toBe(false);
    owner.acceptBatchRow(first, "row", 4);
    owner.acceptBatchRow(first, "row", 3);
    expect(
      owner.getSnapshot().entries.find((entry) => entry.id === first)?.attempt
        ?.plan.request.targets[0],
    ).toMatchObject({ base_row_version: 2 });
    expect(
      owner.getSnapshot().entries.find((entry) => entry.id === second)?.plan
        .request.targets[0],
    ).toMatchObject({ base_row_version: 4 });
    coordination.conflicts.mockReturnValue(false);
    owner.wake();
    await settle();
    expect(transport.send).toHaveBeenCalledTimes(2);
  });
  it("keeps late accepted work hidden during session uncertainty and reconciles after recovery", async () => {
    const { owner, transport, coordination } = setup();
    let finish!: (value: WorkbookBatchTransportOutcome) => void;
    vi.mocked(transport.send).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    const id = required(owner.admit(plan(), { delivery: {} }));
    await settle();
    owner.suspend();
    finish(accepted);
    await settle();
    expect(owner.getSnapshot().entries).toEqual([]);
    expect(coordination.refresh).not.toHaveBeenCalled();
    owner.setAuthority({ ...authority, sessionIdentity: "recovered" });
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    owner.retry(id);
    await settle();
    expect(transport.send).toHaveBeenCalledOnce();
    expect(coordination.refresh).toHaveBeenCalledOnce();
  });
  it("releases older completed payloads after remount reads while retaining conflicts", async () => {
    const { owner, coordination } = setup();
    coordination.refresh.mockRejectedValue(new Error("detached"));
    const first = required(owner.admit(plan("a"), { delivery: {} }));
    const second = required(owner.admit(plan("b"), { delivery: {} }));
    const third = required(owner.admit(plan("c"), { delivery: {} }));
    await settle();
    coordination.conflicts.mockImplementation((id?: string) => id === first);
    owner.surfaceRefreshed(receipt.viewSchemaId);
    expect(owner.getSnapshot().entries.map((entry) => entry.id)).toEqual([
      first,
      third,
    ]);
    expect(
      owner.getSnapshot().entries.some((entry) => entry.id === second),
    ).toBe(false);
    coordination.conflicts.mockReturnValue(false);
    owner.wake();
    await settle();
    expect(owner.getSnapshot().entries.map((entry) => entry.id)).toEqual([
      third,
    ]);
  });
});
