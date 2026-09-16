import { afterEach, describe, expect, it, vi } from "vitest";
import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";
import {
  NetworkFlowTableController,
  type TableTransport,
} from "./NetworkFlowTableController";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  captureTableAttempt,
  normalizeTableDisplayName,
  type TableAttempt,
  TableWriteError,
  validateTableReceipt,
} from "./networkFlowTableOperation";
import { networkFlowTableRecoveryItems } from "./networkFlowTableRecoveryItems";
import { tableAuthority, tableFixture } from "./tableLifecycleTestFixtures";

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((yes) => {
    resolve = yes;
  });
  return { promise, resolve };
}
async function setup(submit?: TableTransport["submit"]) {
  let authority = tableAuthority();
  let catalog = [tableFixture(), tableFixture("b", 1, "Other")];
  const pending = deferred<NetworkFlowTable>();
  const send = vi.fn(
    submit ??
      ((_attempt, _signal, dispatch) => {
        dispatch();
        return pending.promise;
      }),
  );
  const list = vi.fn(async () => catalog);
  const controller = new NetworkFlowTableController();
  const bind = () => controller.bind({ list, submit: send }, () => authority);
  bind();
  await controller.loadTables();
  return {
    controller,
    send,
    list,
    pending,
    setCatalog(value: NetworkFlowTable[]) {
      catalog = value;
    },
    setAuthority(patch: Partial<typeof authority>) {
      authority = { ...authority, ...patch };
      bind();
    },
  };
}
const id = tableFixture().network_flow_table_id;
function rename(controller: NetworkFlowTableController, name = "Renamed") {
  controller.openAction("rename", id);
  controller.setName(name);
}
const tick = async () => {
  for (let i = 0; i < 12; i++) await Promise.resolve();
};
afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});
describe("Table lifecycle operations", () => {
  it("projects stable extension recovery identity across draft dispatch and detachment", async () => {
    const s = await setup();
    rename(s.controller);
    const draft = networkFlowTableRecoveryItems(s.controller.getSnapshot());
    expect(draft).toHaveLength(1);
    expect(draft[0]?.sheetRef?.kind).toBe("extension_workspace");
    const pending = s.controller.submit();
    expect(
      networkFlowTableRecoveryItems(s.controller.getSnapshot())[0]?.attention,
    ).toBe("progress");
    s.controller.closeDialog();
    const operation = networkFlowTableRecoveryItems(s.controller.getSnapshot());
    expect(operation).toHaveLength(1);
    expect(operation[0]?.id).toBe(draft[0]?.id);
    expect(operation[0]?.attention).toBe("attention");
    s.pending.resolve(tableFixture("a", 2, "Renamed"));
    await pending;
    expect(
      networkFlowTableRecoveryItems(s.controller.getSnapshot())[0]?.id,
    ).toBe(draft[0]?.id);
    s.controller.dispose();
    expect(networkFlowTableRecoveryItems(s.controller.getSnapshot())).toEqual(
      [],
    );
  });

  it("admits exact operation roles and synchronously prevents duplicate submission", async () => {
    for (const [role, canRename, canDelete] of [
      ["viewer", false, false],
      ["editor", true, false],
      ["reviewer", false, true],
      ["admin", true, true],
    ] as const) {
      const s = await setup();
      s.setAuthority({ role });
      expect(s.controller.getSnapshot()).toMatchObject({
        canRename,
        canDelete,
      });
      rename(s.controller);
      const first = s.controller.submit();
      void s.controller.submit();
      expect(s.send).toHaveBeenCalledTimes(canRename ? 1 : 0);
      s.controller.dispose();
      await first;
    }
  });
  it("captures immutable bytes and fences queued dispatch after role loss", async () => {
    let dispatch!: () => void;
    const s = await setup((_attempt, _signal, admit) => {
      dispatch = admit;
      return new Promise(() => {});
    });
    rename(s.controller, "\u00a0Cafe\u0301\u3000");
    const request = s.controller.submit();
    const attempt = s.controller.getSnapshot().operation
      ?.attempt as TableAttempt;
    expect(JSON.parse(attempt.body)).toMatchObject({
      display_name: "Café",
      base_table_version: 1,
    });
    expect(Object.isFrozen(attempt)).toBe(true);
    s.setAuthority({ role: "viewer" });
    expect(() => dispatch()).toThrow();
    await request;
    expect(s.controller.getSnapshot().operation?.status).toBe("rejected");
    expect(s.controller.getSnapshot().operation?.failure).toMatchObject({
      certainty: "rejected",
      message: expect.stringContaining("was not submitted"),
    });
    expect(s.controller.getSnapshot().draft?.name).toBe(
      "\u00a0Cafe\u0301\u3000",
    );
  });
  it("bounds uncertainty and replays the identical attempt while preserving later metadata", async () => {
    vi.useFakeTimers();
    const queued = await setup(() => new Promise(() => {}));
    rename(queued.controller);
    const undispatched = queued.controller.submit();
    await vi.advanceTimersByTimeAsync(30_001);
    await undispatched;
    expect(queued.controller.getSnapshot().operation).toMatchObject({
      status: "rejected",
      failure: {
        certainty: "rejected",
        message: expect.stringContaining("was not submitted"),
      },
    });
    queued.controller.dispose();
    const s = await setup();
    rename(s.controller);
    const first = s.controller.submit();
    await vi.advanceTimersByTimeAsync(30_001);
    await first;
    const attempt = s.controller.getSnapshot().operation?.attempt;
    expect(s.controller.getSnapshot().operation?.status).toBe("uncertain");
    s.setCatalog([
      tableFixture("a", 3, "Later"),
      tableFixture("b", 1, "Other"),
    ]);
    await s.controller.loadTables();
    const second = s.controller.replay();
    expect(s.send.mock.calls[1]?.[0]).toBe(attempt);
    s.pending.resolve(tableFixture("a", 2, "Renamed"));
    await second;
    expect(s.controller.getSnapshot().operation?.status).toBe("acknowledged");
    expect(s.controller.getSnapshot().tables[0]?.display_name).toBe("Later");
  });
  it("requires explicit review on peer changes and never reselects a conflict target", async () => {
    const detail = new NetworkFlowRequestError({
      code: "network_flow_table_version_conflict",
      status: 409,
      reasonCode: "stale_version",
      retryAction: "refresh_resource",
      retryable: false,
      safeMessage: "The table changed.",
    });
    const s = await setup(async (_attempt, _signal, dispatch) => {
      dispatch();
      throw new TableWriteError("rejected", "The table changed.", detail);
    });
    rename(s.controller);
    s.setCatalog([tableFixture("a", 2, "Peer"), tableFixture("b", 1, "Other")]);
    s.controller.selectTable(tableFixture("b").network_flow_table_id);
    await s.controller.submit();
    expect(s.controller.getSnapshot().activeTableId).toBe(
      tableFixture("b").network_flow_table_id,
    );
    expect(s.controller.getSnapshot().draft).toMatchObject({
      reviewRequired: true,
      target: { table_version: 1 },
      name: "Renamed",
    });
    s.controller.reviewCurrent();
    expect(s.controller.getSnapshot().draft?.target.table_version).toBe(2);
    s.controller.openAction("delete", id);
    s.controller.setConfirmation("Peer");
    await s.controller.onResourceChange({
      resourceKind: "network_flow_table",
      resourceId: id,
      changeKind: "invalidate",
      reasonCode: "renamed",
    });
    expect(s.controller.getSnapshot().draft).toMatchObject({
      reviewRequired: true,
      confirmation: "",
    });
  });
  it("keeps acknowledgement independent of refresh failure and newer dialogs", async () => {
    const s = await setup();
    rename(s.controller);
    const request = s.controller.submit();
    s.controller.selectTable(tableFixture("b").network_flow_table_id);
    s.controller.openAction("delete", tableFixture("b").network_flow_table_id);
    const dialog = s.controller.getSnapshot().draft?.id;
    s.list.mockRejectedValueOnce(new Error("offline"));
    s.pending.resolve(tableFixture("a", 2, "Renamed"));
    await request;
    expect(s.controller.getSnapshot()).toMatchObject({
      activeTableId: tableFixture("b").network_flow_table_id,
      presentation: "draft",
      draft: { id: dialog },
      loadState: "error",
      operation: { status: "acknowledged" },
    });
  });
  it("fences late completions across navigation session and incident changes", async () => {
    for (const transition of [
      "departure",
      "session",
      "incident",
      "role",
    ] as const) {
      const s = await setup();
      rename(s.controller);
      const request = s.controller.submit();
      if (transition === "departure") s.controller.closeDialog();
      else
        s.setAuthority(
          transition === "session"
            ? { sessionIdentity: null }
            : transition === "incident"
              ? { incidentId: "99999999-9999-4999-8999-999999999999" }
              : { role: "viewer" },
        );
      s.pending.resolve(tableFixture("a", 2, "Renamed"));
      await request;
      expect(
        s.controller
          .getSnapshot()
          .tables.some((table) => table.display_name === "Renamed"),
      ).toBe(false);
      expect(s.controller.getSnapshot().operation?.status).not.toBe(
        "acknowledged",
      );
    }
  });
  it("retains a meaningful delete receipt while removing its active resource", async () => {
    const s = await setup();
    s.controller.openAction("delete", id);
    s.controller.setConfirmation("Flows");
    const request = s.controller.submit();
    s.setCatalog([tableFixture("b", 1, "Other")]);
    s.pending.resolve({
      ...tableFixture("a", 2),
      table_status: "soft_deleted",
      deleted_at: "2026-09-09T20:01:00Z",
    });
    await request;
    expect(s.controller.getSnapshot().operation).toMatchObject({
      status: "acknowledged",
      receipt: { table_status: "soft_deleted", table_version: 2 },
    });
    expect(s.controller.getSnapshot().activeTableId).toBe(
      tableFixture("b").network_flow_table_id,
    );
  });
  it("normalizes names with exact scalar whitespace and control rules", () => {
    expect(normalizeTableDisplayName(" \u0065\u0301 ")).toMatchObject({
      ok: true,
      name: "é",
      scalarCount: 1,
    });
    expect(normalizeTableDisplayName("😀".repeat(64))).toMatchObject({
      ok: true,
      scalarCount: 64,
    });
    expect(normalizeTableDisplayName("😀".repeat(65))).toMatchObject({
      ok: false,
      reason: "display_name_too_long",
    });
    expect(normalizeTableDisplayName("\ufeff")).toMatchObject({
      ok: true,
      name: "\ufeff",
    });
    for (const name of ["\t", "\u0085", "\n", "a\u0000b", "\ud800"])
      expect(normalizeTableDisplayName(name)).toMatchObject({
        ok: false,
        reason: "forbidden_control",
      });
    expect(normalizeTableDisplayName("\u00a0\u3000")).toMatchObject({
      ok: false,
      reason: "empty_display_name",
    });
  });
  it("validates receipt identity lifecycle immutable fields and normalized no-op", () => {
    const target = tableFixture();
    const attempt = captureTableAttempt(
      "rename",
      tableAuthority(),
      target,
      " Flows ",
    );
    expect(validateTableReceipt(target, 200, attempt)).toEqual(target);
    for (const patch of [
      { table_version: 2 },
      { incident_id: "foreign" },
      { network_flow_table_id: "foreign" },
      { mapping_fingerprint: "f".repeat(64) },
      { table_status: "soft_deleted" as const },
      { updated_at: "2026-09-09T20:01:00Z" },
    ])
      expect(() =>
        validateTableReceipt({ ...target, ...patch }, 200, attempt),
      ).toThrow(TableWriteError);
  });
  it("keeps uncertain outcomes after rejected replay and rejects duplicate active names locally", async () => {
    const s = await setup();
    rename(s.controller, "Other");
    await s.controller.submit();
    expect(s.send).not.toHaveBeenCalled();
    expect(s.controller.getSnapshot().draft?.error).toContain("unique");
    s.controller.setName("Renamed");
    void s.controller.submit();
    s.controller.closeDialog();
    await tick();
    s.send.mockImplementationOnce(async (_attempt, _signal, dispatch) => {
      dispatch();
      throw new TableWriteError("rejected", "Transaction conflict");
    });
    await s.controller.replay();
    expect(s.controller.getSnapshot().operation?.status).toBe("uncertain");
    s.controller.dispose();
  });
  it("pauses session state preserves closed drafts and cancels removed-resource work", async () => {
    const s = await setup();
    rename(s.controller);
    const request = s.controller.submit();
    await s.controller.onResourceChange({
      resourceKind: "*",
      resourceId: "*",
      changeKind: "remove",
      reasonCode: "session_revoked",
    });
    expect(s.controller.getSnapshot().hidden).toBe(true);
    s.setAuthority({ sessionIdentity: "reauthorized" });
    expect(s.controller.getSnapshot().operation?.status).toBe("uncertain");
    expect(s.controller.getSnapshot().draft?.name).toBe("Renamed");
    s.controller.onProtectedFailure(
      new NetworkFlowRequestError({
        code: "incident_closed",
        status: 409,
        retryable: false,
        retryAction: "do_not_retry",
        safeMessage: "Closed",
      }),
    );
    expect(s.controller.getSnapshot().draft?.name).toBe("Renamed");
    expect(s.controller.getSnapshot().tables).toEqual([]);
    await s.controller.onResourceChange({
      resourceKind: "network_flow_table",
      resourceId: id,
      changeKind: "remove",
      reasonCode: "soft_deleted",
    });
    expect(s.controller.getSnapshot().draft).toBeNull();
    expect(s.controller.getSnapshot().operation).toBeNull();
    s.pending.resolve(tableFixture("a", 2, "Renamed"));
    await request;
    expect(s.controller.getSnapshot().tables).toEqual([]);
  });
  it("refreshes current reads after write rejection while blocking further cached-role dispatch", async () => {
    const denial = new NetworkFlowRequestError({
      code: "authorization_denied",
      status: 403,
      retryable: false,
      retryAction: "do_not_retry",
      safeMessage: "The current role does not permit rename.",
    });
    const s = await setup(async (_attempt, _signal, dispatch) => {
      dispatch();
      throw new TableWriteError("rejected", denial.message, denial);
    });
    rename(s.controller);
    s.setCatalog([tableFixture("a", 2, "Peer"), tableFixture("b", 1, "Other")]);
    await s.controller.submit();
    expect(s.controller.getSnapshot()).toMatchObject({
      hidden: false,
      loadState: "ready",
      canRename: false,
      canDelete: false,
    });
    expect(s.controller.getSnapshot().tables[0]?.display_name).toBe("Peer");
    expect(await s.controller.loadTables()).toBe(true);
    await s.controller.submit();
    expect(s.send).toHaveBeenCalledTimes(1);
    s.setAuthority({ role: "reviewer" });
    expect(s.controller.getSnapshot()).toMatchObject({
      canRename: false,
      canDelete: true,
    });
  });
});

describe("Table catalog continuity", () => {
  it("fences pending lists crossing rename delete and collaboration", async () => {
    const s = await setup();
    const old = deferred<NetworkFlowTable[]>();
    s.list.mockImplementationOnce(() => old.promise);
    const read = s.controller.loadTables();
    rename(s.controller);
    const mutation = s.controller.submit();
    s.setCatalog([
      tableFixture("a", 2, "Renamed"),
      tableFixture("b", 1, "Other"),
    ]);
    s.pending.resolve(tableFixture("a", 2, "Renamed"));
    await mutation;
    old.resolve([tableFixture()]);
    await read;
    expect(s.controller.getSnapshot().tables[0]?.display_name).toBe("Renamed");
    const stale = deferred<NetworkFlowTable[]>();
    s.list.mockImplementationOnce(() => stale.promise);
    const staleRead = s.controller.loadTables();
    await s.controller.onResourceChange({
      resourceKind: "network_flow_table",
      resourceId: id,
      changeKind: "remove",
      reasonCode: "soft_deleted",
    });
    stale.resolve([tableFixture("a", 9, "Resurrected")]);
    await staleRead;
    await s.controller.loadTables();
    expect(
      s.controller.getSnapshot().tables.map((t) => t.network_flow_table_id),
    ).toEqual([tableFixture("b").network_flow_table_id]);
  });
  it("purges resource drafts and context knowledge only at the authorized lifetime boundary", async () => {
    const s = await setup();
    rename(s.controller);
    const changes = vi.fn();
    s.controller.subscribeChanges(changes);
    s.setCatalog([tableFixture("b", 1, "Other")]);
    await s.controller.loadTables();
    expect(s.controller.getSnapshot().draft).toBeNull();
    expect(changes).toHaveBeenCalledWith(
      expect.objectContaining({ resourceId: id, changeKind: "remove" }),
    );
    s.setAuthority({ sessionIdentity: null, actorId: null });
    expect(s.controller.getSnapshot().removedTableIds).toEqual([]);
    s.setAuthority({
      sessionIdentity: "new",
      actorId: "55555555-5555-4555-8555-555555555555",
    });
    s.setCatalog([tableFixture()]);
    await s.controller.loadTables();
    expect(s.controller.getSnapshot().tables).toHaveLength(1);
  });
  it("preserves navigation across a delayed validated import handoff", async () => {
    const s = await setup();
    const wait = deferred<NetworkFlowTable[]>();
    s.list.mockImplementationOnce(() => wait.promise);
    const table = tableFixture();
    const request = s.controller.handoffImportedTable({
      incidentId: table.incident_id,
      tableId: id,
      sessionId: table.source_import_session_id,
      unitId: table.source_import_unit_id,
      sourceHash: table.source_content_sha256,
      fingerprint: table.mapping_fingerprint,
      current: () => true,
      signal: new AbortController().signal,
    });
    s.controller.selectTable(tableFixture("b").network_flow_table_id);
    wait.resolve([table, tableFixture("b", 1, "Other")]);
    expect(await request).toEqual({ kind: "superseded" });
    expect(s.controller.getSnapshot().activeTableId).toBe(
      tableFixture("b").network_flow_table_id,
    );
  });
});
