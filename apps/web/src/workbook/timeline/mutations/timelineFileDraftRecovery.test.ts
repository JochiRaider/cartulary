import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { timelineRow } from "../../../testing/timelineWorkbookTestSupport";
import type { WorkbookPendingMutationPort } from "../../ports/WorkbookPendingMutationPort";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { buildStableMutationSignature } from "../../utils/workbookPendingQueue";
import { buildCreatePayload } from "../models/timelineMutationIntents";
import { timelinePendingSavesRefsFor } from "../models/timelinePendingSaves";
import { createDraftRow, rowFromApi } from "../models/timelineRowModel";
import type { TimelineMutationDriverPorts } from "./createTimelineMutationDriver";
import { timelineMutationOwnerFor } from "./WorkbookTimelineMutationOwner";

const recordId = "20000000-0000-4000-8000-000000000002",
  evidenceId = "30000000-0000-4000-8000-000000000003";
const runtimes: WorkbookMutationRuntime[] = [];
afterEach(() => {
  for (const runtime of runtimes.splice(0))
    runtime.invalidate({ kind: "runtime_disposed" });
});
function fixture() {
  let sequence = 0;
  const ids = { create: (prefix: string) => `${prefix}-${++sequence}` };
  const acknowledgement =
    deferred<Awaited<ReturnType<WorkbookPendingMutationPort["execute"]>>>();
  const acceptedRow = timelineRow({
    recordId,
    rowVersion: 1,
    captureState: "rough",
    evidenceCount: 1,
  });
  const receipt = {
    row: { ...acceptedRow, view_schema_id: timelineViewSchemaId },
    viewSchemaId: timelineViewSchemaId,
    changeSetId: "40000000-0000-4000-8000-000000000004",
  };
  const execute = vi
    .fn<WorkbookPendingMutationPort["execute"]>()
    .mockReturnValueOnce(acknowledgement.promise)
    .mockResolvedValue({
      kind: "accepted",
      value: {
        ...receipt,
        row: {
          ...acceptedRow,
          row_version: 2,
          view_schema_id: timelineViewSchemaId,
        },
      },
    });
  const runtime = new WorkbookMutationRuntime(
    { incidentId: "incident", clientInstanceId: "client" },
    ids,
    { execute },
  );
  runtimes.push(runtime);
  const owner = timelineMutationOwnerFor(runtime),
    row = createDraftRow(1),
    rowsRef = { current: [row] };
  const pending = timelinePendingSavesRefsFor(runtime, runtime.pendingQueue());
  const ports: TimelineMutationDriverPorts = {
    mutationRuntime: runtime,
    pendingSavesRefs: pending,
    rowsRef,
    sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
    mutationCommands: {
      createLogicalActionId: () => ids.create("ordinary"),
      createConflictRecoveryId: () => ids.create("recovery"),
    },
    applyAcceptedRowMutation: (_key, value) => {
      const row = rowFromApi(value.row);
      rowsRef.current = [row];
      return row;
    },
    clearSubmittedScalarEditorDraftValuesForRow: vi.fn(),
    captureEditorDrafts: () => new Map(),
    acceptEditorPredecessor: () => {},
    clearViewportContinuity: vi.fn(),
    conflictQueueRef: { current: {} },
    registerMutationConflict: () => true,
    latestCommittedTimelineRow: (id) =>
      rowsRef.current.find((row) => row.recordId === id) ?? null,
    loadRows: async () => {},
    postMutationQueryRefreshRequired: false,
    publishPendingQueueState: vi.fn(),
    reconcileDiscardedPendingUnit: vi.fn(),
    recordWorkbookTiming: vi.fn(),
    requestAuthorizationRecovery: vi.fn(),
    setRefreshError: vi.fn(),
    setMutationError: vi.fn(),
    rowStoreCommands: {
      replaceRows: (rows) => {
        rowsRef.current = rows;
      },
      updateRows: (update) => {
        rowsRef.current = update(rowsRef.current);
      },
    },
  };
  owner.configureReader(async () => acceptedRow);
  const detach = owner.attach(() => ports);
  const type = () => {
    const authored = {
      ...row,
      values: {
        ...row.values,
        activitySynopsisText: "Typed while upload continues",
      },
    };
    rowsRef.current = [authored];
    const clientTxnId = ids.create("ordinary"),
      payloadIntent = buildCreatePayload(authored, clientTxnId);
    if (!payloadIntent) throw new Error("Expected qualifying draft input");
    owner.enqueuePendingReplayUnit({
      id: `pending-${clientTxnId}`,
      kind: "create",
      source: "autosave",
      incidentId: "incident",
      clientInstanceId: "client",
      viewSchemaId: timelineViewSchemaId,
      rowKey: row.key,
      recordId: null,
      clientTxnId,
      payloadIntent,
      coalesceKey: `draft:${row.key}`,
      enqueueOrder: pending.pendingReplayOrderRef.current++,
      operationClass: "hot_path",
      status: "queued",
      mutationSignature: buildStableMutationSignature(payloadIntent),
      focusField: "activitySynopsisText",
      focusKey: "draft",
      surface: "grid",
      rowSnapshot: authored,
      continueOnFreshDraft: true,
      detectAutoResolution: false,
      promoteToCommittedRowInspect: false,
      viewportContinuityToken: undefined,
    });
  };
  return {
    runtime,
    owner,
    row,
    execute,
    acknowledgement,
    receipt,
    type,
    detach,
    ports,
  };
}
it("shares screenshot creation with later ordinary input and retains its promotion after detachment", async () => {
  const f = fixture();
  f.owner.fileDrafts.attachEvidence(f.row.key, evidenceId);
  await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
  f.type();
  f.detach();
  f.acknowledgement.resolve({ kind: "accepted", value: f.receipt });
  await vi.waitFor(() =>
    expect(f.owner.fileDrafts.resolve(f.row.key).kind).toBe("promoted"),
  );
  f.runtime.registerSurface(timelineViewSchemaId, async () => {});
  f.runtime.requestDrain();
  await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(2));
  expect(f.execute.mock.calls.map(([input]) => input.unit.kind)).toEqual([
    "create",
    "patch",
  ]);
  expect(f.execute.mock.calls[1]?.[0].unit.recordId).toBe(recordId);
  expect(f.execute.mock.calls[0]?.[0].unit.payloadIntent).toMatchObject({
    "timeline.attached_evidence_ids": {
      actions: [{ op: "add_record_ref", linked_record_id: evidenceId }],
    },
  });
  expect(f.owner.fileDrafts.resolve(f.row.key)).toMatchObject({
    kind: "promoted",
    receipt: f.receipt,
  });
  expect(f.ports.clearViewportContinuity).not.toHaveBeenCalled();
});
it("waits for the original ordinary create instead of admitting a second screenshot create", async () => {
  const f = fixture();
  f.type();
  await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
  expect(f.owner.fileDrafts.resolve(f.row.key)).toEqual({ kind: "pending" });
  f.owner.fileDrafts.attachEvidence(f.row.key, evidenceId);
  expect(f.execute).toHaveBeenCalledTimes(1);
  f.acknowledgement.resolve({ kind: "accepted", value: f.receipt });
  await vi.waitFor(() =>
    expect(f.owner.fileDrafts.resolve(f.row.key).kind).toBe("promoted"),
  );
  expect(f.execute).toHaveBeenCalledTimes(1);
});
