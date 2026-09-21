import { act, renderHook, waitFor } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { WorkbookRecordHistoryOwner } from "../history/WorkbookRecordHistoryOwner";
import { useWorkbookRecordHistoryController } from "./useWorkbookRecordHistoryController";
import { useWorkbookRecordHistoryState } from "./useWorkbookRecordHistoryState";

it("acknowledges a history mutation before its projection refresh settles", async () => {
  let finishRefresh!: () => void;
  const refresh = new Promise<void>((resolve) => {
    finishRefresh = resolve;
  });
  const subject = {
    kind: "live" as const,
    recordId: "record-1",
    rowVersion: 1,
    label: "A row",
    surfaceLabel: "Timeline",
    viewSchemaId: "cartulary.view.timeline.v2",
  };
  const owner = new WorkbookRecordHistoryOwner("incident-1", {
    create: () => crypto.randomUUID(),
  });
  owner.setAuthority({
    actorId: "reviewer",
    incidentId: "incident-1",
    role: "reviewer",
    closed: false,
  });
  let committed = false;
  owner.configure({
    load: async () => ({
      kind: "accepted",
      value: {
        incident_id: "incident-1",
        record_id: "record-1",
        row_version: committed ? 2 : 1,
        deleted: committed,
        representation_generation: "cartulary.history.1",
        items: [],
        paging: { limit: 100, has_more: false, next_cursor: null },
      },
    }),
    send: async () => {
      committed = true;
      return {
        kind: "acknowledged",
        receipt: {
          kind: "delete",
          incidentId: "incident-1",
          recordId: "record-1",
          rowVersion: 2,
          changeSetId: "change-1",
          deleted: true,
          deletedAt: "2026-09-10T00:00:00Z",
          deletedByUserId: "reviewer",
        },
      };
    },
  });
  const { result } = renderHook(() =>
    useWorkbookRecordHistoryController({
      presentation: useWorkbookRecordHistoryState(),
      coordinate: async () => subject.rowVersion,
      owner,
      canMutate: true,
      subject,
      ownerEffects: {
        refresh: () => refresh,
        deleteAccepted: vi.fn(),
        restoreAccepted: vi.fn(),
        rollbackAccepted: vi.fn(),
      },
    }),
  );
  act(() => result.current.commands.open());
  await waitFor(() => expect(result.current.snapshot.phase).toBe("ready"));
  act(() => result.current.commands.previewDeleteRestore("delete"));
  await waitFor(() =>
    expect(
      result.current.snapshot.phase === "ready" &&
        result.current.snapshot.pendingAction,
    ).toBeTruthy(),
  );
  let completion!: Promise<void>;
  act(() => {
    completion = result.current.commands.confirm();
  });
  try {
    await waitFor(() =>
      expect(owner.getSnapshot()[0]?.phase).toBe("acknowledged"),
    );
    expect(result.current.snapshot.phase).not.toBe("submitting");
  } finally {
    await act(async () => {
      finishRefresh();
      await completion;
    });
  }
});

it("fences admitted effects when selection leaves and returns to the same record", async () => {
  const subject = {
    kind: "live" as const,
    recordId: "record-1",
    rowVersion: 1,
    label: "Row",
    surfaceLabel: "Timeline",
    viewSchemaId: "cartulary.view.timeline.v2",
  };
  const owner = new WorkbookRecordHistoryOwner("incident-1", {
    create: () => crypto.randomUUID(),
  });
  owner.setAuthority({
    actorId: "reviewer",
    incidentId: "incident-1",
    role: "reviewer",
    closed: false,
  });
  let committed = false;
  let finish!: () => void;
  const pendingResponse = new Promise<void>((resolve) => {
    finish = resolve;
  });
  const send = vi.fn(async () => {
    await pendingResponse;
    committed = true;
    return {
      kind: "acknowledged" as const,
      receipt: {
        kind: "delete" as const,
        incidentId: "incident-1",
        recordId: "record-1",
        rowVersion: 2,
        changeSetId: "change-1",
        deleted: true,
        deletedAt: "2026-09-10T00:00:00Z",
        deletedByUserId: "reviewer",
      },
    };
  });
  owner.configure({
    load: async (recordId) => ({
      kind: "accepted",
      value: {
        incident_id: "incident-1",
        record_id: recordId,
        row_version: committed ? 2 : 1,
        deleted: committed,
        representation_generation: "cartulary.history.1",
        items: [],
        paging: { limit: 100, has_more: false, next_cursor: null },
      },
    }),
    send,
  });
  const effects = {
    refresh: vi.fn(async () => {}),
    deleteAccepted: vi.fn(),
    restoreAccepted: vi.fn(),
    rollbackAccepted: vi.fn(),
  };
  const { result, rerender } = renderHook(
    ({ recordId }) =>
      useWorkbookRecordHistoryController({
        presentation: useWorkbookRecordHistoryState(),
        coordinate: async () => subject.rowVersion,
        owner,
        canMutate: true,
        subject: { ...subject, recordId },
        ownerEffects: effects,
      }),
    { initialProps: { recordId: "record-1" } },
  );
  act(() => result.current.commands.previewDeleteRestore("delete"));
  await waitFor(() =>
    expect(
      result.current.snapshot.phase === "ready" &&
        result.current.snapshot.pendingAction,
    ).toBeTruthy(),
  );
  let completion!: Promise<void>;
  act(() => {
    completion = result.current.commands.confirm();
  });
  await waitFor(() => expect(send).toHaveBeenCalledOnce());
  rerender({ recordId: "record-2" });
  rerender({ recordId: "record-1" });
  await act(async () => {
    finish();
    await completion;
  });
  await waitFor(() =>
    expect(owner.getSnapshot()[0]?.phase).toBe("acknowledged"),
  );
  expect(effects.deleteAccepted).not.toHaveBeenCalled();
  expect(effects.refresh).not.toHaveBeenCalled();
});
