import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { taskAuthority } from "../../testing/taskWorkbookTestSupport";
import type { RecordPatchTransport } from "../adapters/workbookRecordPatchTransport";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { createWorkbookMutationRuntime } from "../runtime/createWorkbookMutationRuntime";
import { useGenericSurfaceMutationController } from "./useGenericSurfaceMutationController";

const viewSchemaId = "cartulary.view.notes.v1",
  recordId = "00000000-0000-4000-8000-000000000603";
const baseline: WorkbookQueryRow & { readonly view_schema_id: string } = {
  record_id: recordId,
  row_version: 1,
  view_schema_id: viewSchemaId,
  cells: Object.fromEntries(
    requireViewContract(viewSchemaId).fields.map((field) => [
      field.fieldKey,
      { value: null },
    ]),
  ),
};
const request = {
  baseline,
  viewSchemaId,
  recordId,
  baseRowVersion: 1,
  purpose: "generic-patch",
  changes: [{ field_key: "note.body", value: "Authored note" }],
};
const receipt = {
  viewSchemaId,
  changeSetId: "change",
  row: { ...baseline, row_version: 2 },
};
function fixture() {
  let sequence = 0;
  const runtime = createWorkbookMutationRuntime(
    {
      incidentId: taskAuthority.incidentId,
      clientInstanceId: "ordinary-inspector",
    },
    { create: () => `patch-${++sequence}` },
    { execute: vi.fn() },
  );
  const send = vi.fn<RecordPatchTransport["send"]>(async () => ({
    kind: "acknowledged",
    receipt,
  }));
  const readSource = vi.fn(async () => baseline);
  runtime.explicitPatches.configure({ send }, undefined, readSource);
  runtime.setAuthority(taskAuthority);
  const refresh = vi.fn(async () => {});
  runtime.registerSurface(viewSchemaId, refresh);
  const hook = renderHook(() =>
    useGenericSurfaceMutationController({
      mutationRuntime: runtime,
      surfaceLabel: "Notes",
      sheetRef: { kind: "view_schema", id: viewSchemaId },
      selectedRecordId: recordId,
    }),
  );
  return { runtime, send, refresh, readSource, hook };
}
afterEach(cleanup);
it("verifies committed versions through the source owner before capture and excludes later grid authoring from admission", async () => {
  for (const changedEditedField of [false, true]) {
    const f = fixture();
    f.runtime.history.acceptVersion(recordId, 2);
    f.send.mockResolvedValue({
      kind: "acknowledged",
      receipt: { ...receipt, row: { ...receipt.row, row_version: 3 } },
    });
    f.readSource.mockImplementation(async () => ({
      ...baseline,
      row_version: 2,
      cells: {
        ...baseline.cells,
        [changedEditedField ? "note.body" : "note.title"]: {
          value: "Concurrent",
        },
      },
    }));
    const pending = f.runtime.explicitPatches.submit({
      ...request,
      sheetRef: { kind: "view_schema", id: viewSchemaId },
      surfaceLabel: "Notes",
    });
    expect(
      f.runtime.enqueuePatch({
        ...request,
        fieldKey: "note.body",
        localValue: "Later grid text",
        rowLabel: "Note",
        surfaceLabel: "Notes",
      }).kind,
    ).toBe("rejected_mutation");
    const entry = await pending;
    expect(f.readSource).toHaveBeenCalled();
    if (changedEditedField) {
      expect(entry?.phase).toBe("preparation_failed");
      expect(f.send).not.toHaveBeenCalled();
    } else {
      expect(entry?.request?.baseRowVersion).toBe(2);
      expect(JSON.parse(entry?.request?.body ?? "{}").changes).toEqual(
        request.changes,
      );
      expect(f.send).toHaveBeenCalledTimes(1);
    }
  }
});
it("keeps the original ordinary inspector attempt after response loss", async () => {
  const f = fixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  await act(async () => {
    await f.hook.result.current.submitPatchMutation(request);
  });
  const entry = f.runtime.explicitPatches.getSnapshot().entries[0];
  if (!entry) throw new Error("Missing retained attempt");
  expect(entry.phase).toBe("uncertain");
  await act(async () => {
    await f.hook.result.current.submitPatchMutation({
      ...request,
      changes: [{ field_key: "note.body", value: "Later authoring" }],
    });
  });
  expect(f.send).toHaveBeenCalledTimes(1);
  await act(() => f.runtime.explicitPatches.replay(entry.id));
  expect(f.send).toHaveBeenCalledTimes(2);
  expect(f.send.mock.calls[1]?.[0]).toBe(f.send.mock.calls[0]?.[0]);
  expect(f.runtime.explicitPatches.getSnapshot().entries[0]?.receipt).toEqual(
    receipt,
  );
});
it("retains ordinary acknowledgement independently of refresh failure", async () => {
  const f = fixture();
  f.refresh.mockRejectedValueOnce(new Error("refresh lost"));
  await act(async () => {
    await f.hook.result.current.submitPatchMutation(request);
  });
  const entry = f.runtime.explicitPatches.getSnapshot().entries[0];
  if (!entry) throw new Error("Missing retained receipt");
  expect(entry.receipt).toEqual(receipt);
  expect(entry.reconciliation).toBe("required");
  await act(() => f.runtime.explicitPatches.refresh(entry.id));
  expect(f.send).toHaveBeenCalledTimes(1);
  expect(
    f.runtime.explicitPatches.getSnapshot().entries[0]?.reconciliation,
  ).toBe("complete");
});

it("captures authoring and scope before transport and retains the receipt before effects", async () => {
  const f = fixture();
  const changes = [{ field_key: "note.body", value: "captured" }];
  const acknowledged = vi.fn(() =>
    expect(f.runtime.explicitPatches.getSnapshot().entries[0]?.receipt).toEqual(
      receipt,
    ),
  );
  const submitting = f.runtime.explicitPatches.submit(
    {
      ...request,
      changes,
      sheetRef: { kind: "view_schema", id: viewSchemaId },
      surfaceLabel: "Notes",
      authoringRevision: 4,
      presentationIdentity: "original",
    },
    [{ acknowledged }],
  );
  const changed = changes[0];
  if (changed) changed.value = "newer";
  await submitting;
  const entry = f.runtime.explicitPatches.getSnapshot().entries[0];
  expect(entry?.authority).toEqual(taskAuthority);
  expect(entry?.intent.authoringRevision).toBe(4);
  expect(JSON.parse(entry?.request?.body ?? "{}").changes).toEqual([
    { field_key: "note.body", value: "captured" },
  ]);
  expect(acknowledged).toHaveBeenCalledTimes(1);
});
it("separates definitive rejection from uncertainty without automatically rekeying a request", async () => {
  for (const uncertain of [false, true]) {
    const f = fixture();
    f.send.mockResolvedValueOnce(
      uncertain
        ? { kind: "uncertain" }
        : {
            kind: "rejected",
            failure: {
              kind: "client_txn_conflict",
              message: "Identity already used",
            },
          },
    );
    await act(async () => {
      await f.hook.result.current.submitPatchMutation(request);
    });
    const entry = f.runtime.explicitPatches.getSnapshot().entries[0];
    if (!entry) throw new Error("Missing attempt");
    if (uncertain) {
      f.send.mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "client_txn_conflict",
          message: "Identity already used",
        },
      });
      await act(() => f.runtime.explicitPatches.replay(entry.id));
    }
    expect(f.runtime.explicitPatches.getSnapshot().entries[0]?.phase).toBe(
      uncertain ? "uncertain" : "rejected",
    );
    expect(f.runtime.explicitPatches.getSnapshot().entries).toHaveLength(1);
    if (uncertain)
      expect(f.send.mock.calls[1]?.[0]).toBe(f.send.mock.calls[0]?.[0]);
  }
});
it("keeps collection conflicts on their review class and clears only captured authoring on resolution", async () => {
  const f = fixture(),
    resolved = vi.fn();
  f.send.mockResolvedValueOnce({
    kind: "rejected",
    failure: {
      kind: "same_field_conflict",
      message: "Review saved tags",
      conflict: {
        record_id: recordId,
        field_key: "note.tags",
        conflict_token: "original-token",
        conflict_resolution_class: "collection_review",
        base_row_version: 1,
        current_row_version: 2,
        client_value: null,
        server_value: null,
      },
    },
  });
  const entry = await f.runtime.explicitPatches.submit(
    {
      ...request,
      changes: [
        {
          field_key: "note.tags",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [{ op: "add_tag", tag_name: "reviewed" }],
          },
        },
      ],
      sheetRef: { kind: "view_schema", id: viewSchemaId },
      surfaceLabel: "Notes",
    },
    [{ conflictResolved: resolved }],
  );
  expect(entry?.phase).toBe("conflict");
  expect(
    f.runtime.getSnapshot().conflicts[0]?.conflict.conflict_resolution_class,
  ).toBe("collection_review");
  f.runtime.explicitPatches.conflictResolved(
    recordId,
    "keep_saved",
    receipt.row,
  );
  expect(resolved).toHaveBeenCalledTimes(1);
  expect(f.send).toHaveBeenCalledTimes(1);
});
it("rejects obsolete presentation before dispatch and conceals later completion during authority suspension", async () => {
  const f = fixture();
  const blocked = await f.runtime.explicitPatches.submit(
    {
      ...request,
      sheetRef: { kind: "view_schema", id: viewSchemaId },
      surfaceLabel: "Notes",
    },
    [
      {
        prepare: async () => {
          throw new Error("obsolete attachment");
        },
      },
    ],
  );
  expect(blocked?.phase).toBe("preparation_failed");
  expect(f.send).not.toHaveBeenCalled();
  let complete: (
    outcome: Awaited<ReturnType<RecordPatchTransport["send"]>>,
  ) => void = () => {};
  f.send.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        complete = resolve;
      }),
  );
  const submitted = f.runtime.explicitPatches.submit({
    ...request,
    sheetRef: { kind: "view_schema", id: viewSchemaId },
    surfaceLabel: "Notes",
  });
  await vi.waitFor(() => expect(f.send).toHaveBeenCalledTimes(1));
  f.runtime.setAuthority({ ...taskAuthority, role: "" });
  expect(f.runtime.explicitPatches.getSnapshot().entries).toEqual([]);
  expect(f.runtime.explicitPatches.latestRow(recordId)).toBeNull();
  f.runtime.explicitPatches.suspend();
  complete({ kind: "acknowledged", receipt });
  await submitted;
  expect(f.runtime.explicitPatches.getSnapshot().entries).toEqual([]);
  f.runtime.setAuthority({
    ...taskAuthority,
    sessionIdentity: "recovered-session",
  });
  expect(f.runtime.explicitPatches.getSnapshot().entries[1]?.receipt).toEqual(
    receipt,
  );
  f.runtime.setAuthority({
    ...taskAuthority,
    actorId: "replacement",
  });
  expect(f.runtime.explicitPatches.getSnapshot().entries).toEqual([]);
});
