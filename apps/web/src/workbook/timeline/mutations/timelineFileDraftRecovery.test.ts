import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import {
  mentionReview,
  mentionWorkbookRow,
} from "../../../testing/timelineMentionTestSupport";
import { timelineRow } from "../../../testing/timelineWorkbookTestSupport";
import type { WorkbookPendingMutationPort } from "../../ports/WorkbookPendingMutationPort";
import { createWorkbookMutationRuntime } from "../../runtime/createWorkbookMutationRuntime";
import { buildStableMutationSignature } from "../../runtime/pending/workbookPendingQueue";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type {
  WorkbookBatchTransport,
  WorkbookBatchTransportOutcome,
} from "../../runtime/workbookBatchOperation";
import { timelineMentionOwnerFor } from "../actions/timelineMentionOwnerFor";
import { createTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { buildCreatePayload } from "../models/timelineMutationIntents";
import { timelinePendingSavesRefsFor } from "../models/timelinePendingSaves";
import { rowFromApi } from "../models/timelineRowModel";
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
  const runtime = createWorkbookMutationRuntime(
    { incidentId: "incident", clientInstanceId: "client" },
    ids,
    { execute },
  );
  runtimes.push(runtime);
  const owner = timelineMutationOwnerFor(runtime),
    row = owner.initialRows()[0],
    rowsRef = { current: row ? [row] : [] };
  if (!row) throw new Error("Missing initial capture draft");
  const pending = timelinePendingSavesRefsFor(runtime, runtime.pendingQueue());
  const drafts = createTimelineEditorDraftRegistry(
    runtime.localDraftsForSurface(timelineViewSchemaId),
    owner.capture,
  );
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
    settleEditorRevisions: drafts.settleRevisions,
    batchAuthoring: drafts.batch,
    captureEditorDrafts: drafts.captureRow,
    acceptEditorPredecessor: drafts.acceptPredecessor,
    clearViewportContinuity: vi.fn(),
    conflictQueueRef: { current: {} },
    registerMutationConflict: () => true,
    latestCommittedTimelineRow: (id) =>
      rowsRef.current.find((row) => row.recordId === id) ?? null,
    loadRows: async () => {},
    postMutationQueryRefreshRequired: false,
    publishPendingQueueState: vi.fn(),
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
  const type = (
    onSettled?: Parameters<typeof owner.enqueuePendingReplayUnit>[1],
    value = "Typed while upload continues",
    surface: "grid" | "inspector" = "grid",
  ) => {
    let authored = {
      ...row,
      values: {
        ...row.values,
        activitySynopsisText: value,
      },
    };
    drafts.setDraft(
      { rowKey: row.key, field: "activitySynopsisText", surface },
      authored.values.activitySynopsisText,
      row,
      true,
    );
    authored = drafts.materializeRow(authored, { surface });
    rowsRef.current = [authored];
    const clientTxnId = ids.create("ordinary"),
      payloadIntent = buildCreatePayload(authored, clientTxnId, {
        allowZeroFieldCreate: true,
      });
    if (!payloadIntent) throw new Error("Expected qualifying draft input");
    owner.enqueuePendingReplayUnit(
      {
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
        mutationSignature: JSON.stringify([
          buildStableMutationSignature(payloadIntent),
          clientTxnId,
        ]),
        focusField: "activitySynopsisText",
        focusKey: "draft",
        surface,
        rowSnapshot: authored,
        continueOnFreshDraft: true,
        detectAutoResolution: false,
        promoteToCommittedRowInspect: false,
        viewportContinuityToken: undefined,
      },
      onSettled,
    );
  };
  return {
    runtime,
    owner,
    drafts,
    row,
    execute,
    acknowledgement,
    receipt,
    type,
    detach,
    ports,
  };
}
it("settles a valid empty capture without clearing authoring added after dispatch", async () => {
  const f = fixture();
  f.owner.fileDrafts.attachEvidence(f.row.key, evidenceId);
  await waitFor(() => expect(f.execute).toHaveBeenCalledOnce());
  const unit = f.execute.mock.calls[0]?.[0].unit;
  if (!unit) throw new Error("Missing dispatched create");
  const context = f.ports.pendingSavesRefs.replayContextByUnitId.get(unit.id);
  expect(context?.draftRevisions).toEqual(new Map());
  expect(context?.capturePredecessor).toBeUndefined();
  expect(unit.recordId).toBeNull();
  f.drafts.setDraft(
    { rowKey: f.row.key, field: "activitySynopsisText", surface: "grid" },
    "Typed after the empty capture",
    f.row,
    true,
  );
  f.acknowledgement.resolve({ kind: "accepted", value: f.receipt });
  await waitFor(() =>
    expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(0),
  );
  expect(
    f.drafts.draftValue({
      rowKey: f.row.key,
      field: "activitySynopsisText",
      surface: "grid",
    }),
  ).toBe("Typed after the empty capture");
  expect(f.execute).toHaveBeenCalledOnce();
  f.detach();
});

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

it("retains matching batch disclosure from captured sources across query races and presentation detachment", async () => {
  const f = fixture();
  const base = mentionReview();
  const review = {
    ...base,
    subject: {
      ...base.subject,
      incidentId: "incident",
      sourceRecordId: recordId,
      state: "resolved" as const,
      resolutionMethod: "auto_match",
      resolvedRecordId: "70000000-0000-4000-8000-000000000001",
    },
  };
  const after = mentionWorkbookRow(review);
  const before = {
    ...after,
    rowVersion: 1,
    collectionValues: {
      ...after.collectionValues,
      hostRefs: [],
      identityRefs: [],
    },
  };
  f.ports.rowStoreCommands.replaceRows([before]);
  const mentions = timelineMentionOwnerFor(f.runtime);
  mentions.setAuthority({ ...review.authority, incidentId: "incident" });
  f.runtime.batches.setAuthority({
    ...review.authority,
    incidentId: "incident",
    closed: false,
  });
  const response = deferred<WorkbookBatchTransportOutcome>();
  const send = vi
    .fn<WorkbookBatchTransport["send"]>()
    .mockReturnValue(response.promise);
  f.runtime.batches.configure({
    capture: (plan, authority, id) => ({
      plan,
      authority,
      id,
      apiBase: undefined,
      path: "/captured-fixture",
      body: JSON.stringify({ ...plan.request, client_txn_id: id }),
    }),
    send,
  });
  const id = f.runtime.batches.admit(
    {
      operation: "pasteWorkbookClipboard",
      recordIds: [recordId],
      request: {
        view_schema_id: timelineViewSchemaId,
        columns: ["timeline.host_refs"],
        start_field_key: "timeline.host_refs",
        clipboard_text: review.subject.rawText,
        targets: [{ kind: "record", record_id: recordId, base_row_version: 1 }],
      },
    },
    { delivery: {} },
  );
  expect(id).not.toBeNull();
  await vi.waitFor(() => expect(send).toHaveBeenCalledTimes(1));
  // The source query sees acceptance before the HTTP receipt; then Timeline leaves.
  f.ports.rowStoreCommands.replaceRows([after]);
  mentions.observeSource(after);
  f.detach();
  if (!after.rawRow) throw new Error("Complete accepted fixture required");
  response.resolve({
    kind: "acknowledged",
    receipt: {
      viewSchemaId: timelineViewSchemaId,
      changeSetId: "matching-batch-change",
      rows: [after.rawRow],
      conflicts: [],
    },
  });
  await vi.waitFor(() =>
    expect(mentions.getDisclosureSnapshot()).toHaveLength(1),
  );
  expect(mentions.getDisclosureSnapshot()[0]).toMatchObject({
    entityMentionId: review.subject.mentionId,
    acceptedCount: 1,
    operation: {
      kind: "batch",
      operationId: id,
      changeSetId: "matching-batch-change",
    },
  });
  expect(send).toHaveBeenCalledTimes(1);
});

it("settles a halted Timeline editor only when Retry or Discard resolves its logical operation", async () => {
  for (const action of ["retry", "discard"] as const) {
    const f = fixture();
    const settled = vi.fn();
    f.type(settled);
    await waitFor(() => expect(f.execute).toHaveBeenCalledOnce());
    f.acknowledgement.resolve({
      kind: "rejected",
      failure: {
        kind: "client_txn_conflict",
        message: "Request identity needs recovery",
      },
    });
    await waitFor(() =>
      expect(f.runtime.getSnapshot().blockedEdit).not.toBeNull(),
    );
    expect(settled).not.toHaveBeenCalled();
    const id = f.runtime.getSnapshot().blockedEdit?.unitId;
    if (!id) throw new Error("Missing halted operation");
    if (action === "retry") f.owner.retryBlockedEdit(id);
    else f.owner.discardBlockedEdit(id);
    await waitFor(() => expect(settled).toHaveBeenCalledOnce());
    expect(settled.mock.calls[0]?.[0].kind).toBe(
      action === "retry" ? "accepted" : "rejected_mutation",
    );
    expect(f.execute).toHaveBeenCalledTimes(action === "retry" ? 2 : 1);
    f.detach();
  }
});

it("settles terminal validation immediately while preserving retryable transaction recovery", async () => {
  const f = fixture();
  const settled = vi.fn();
  f.type(settled);
  await waitFor(() => expect(f.execute).toHaveBeenCalledOnce());
  f.acknowledgement.resolve({
    kind: "rejected",
    failure: {
      kind: "validation",
      publicCode: "invalid_request",
      message: "Correct the original edit",
    },
  });
  await waitFor(() => expect(settled).toHaveBeenCalledOnce());
  expect(settled.mock.calls[0]?.[0].kind).toBe("validation_error");
  const id = f.runtime.getSnapshot().blockedEdit?.unitId;
  if (!id) throw new Error("Missing rejected queue unit");
  f.owner.discardBlockedEdit(id);
  expect(settled).toHaveBeenCalledOnce();
  expect(f.execute).toHaveBeenCalledOnce();
  f.detach();
});

it("promotes newer retained capture authoring after detached acknowledgement", async () => {
  const f = fixture();
  const identity = {
    rowKey: f.row.key,
    field: "activitySynopsisText" as const,
    surface: "grid" as const,
  };
  f.drafts.setDraft(identity, "A", f.row, true);
  f.drafts.beginCapture(f.row.key);
  f.type();
  await vi.waitFor(() => expect(f.execute).toHaveBeenCalledTimes(1));
  f.drafts.setDraft(identity, "B", f.row, true);
  const revision = [...f.drafts.captureRow(f.row.key, "grid").values()][0];
  f.detach();
  f.acknowledgement.resolve({ kind: "accepted", value: f.receipt });
  await vi.waitFor(() =>
    expect(f.owner.fileDrafts.resolve(f.row.key).kind).toBe("promoted"),
  );
  expect(f.drafts.resolveRowKey(f.row.key)).toBe(recordId);
  expect(f.drafts.draftValue({ ...identity, rowKey: recordId })).toBe("B");
  expect([...f.drafts.captureRow(recordId, "grid").values()]).toEqual([
    revision,
  ]);
  expect(f.execute).toHaveBeenCalledTimes(1);
  const reattached = f.owner.initialRows();
  expect(reattached.filter((row) => row.recordId === null)).toHaveLength(1);
  expect(reattached.find((row) => row.recordId === null)?.key).not.toBe(
    f.row.key,
  );
  expect(f.drafts.draftValue({ ...identity, rowKey: recordId })).toBe("B");
});

it("settles capture successors by revision including clears and equal text without overwriting unrelated fields", async () => {
  for (const later of ["", "B", "A", "  Ω 東京  "]) {
    const f = fixture();
    f.type(undefined, "A");
    await waitFor(() => expect(f.execute).toHaveBeenCalledOnce());
    const captured = structuredClone(f.execute.mock.calls[0]?.[0].unit);
    if (later === "B")
      f.drafts.setDraft(
        { rowKey: f.row.key, field: "analystText", surface: "grid" },
        "Second owned field",
        f.row,
        true,
      );
    f.type(undefined, later);
    const identity = {
      rowKey: f.row.key,
      field: "activitySynopsisText" as const,
      surface: "grid" as const,
    };
    f.drafts.setDraft(
      { ...identity, surface: "inspector" },
      "Independent inspector",
      f.row,
      true,
    );
    const receipt = {
      ...f.receipt,
      row: {
        ...f.receipt.row,
        cells: {
          ...f.receipt.row.cells,
          "timeline.activity_synopsis_text": { value: "A" },
          "timeline.data_source_text": {
            value: "Independent authoritative source",
          },
        },
      },
    };
    f.acknowledgement.resolve({ kind: "accepted", value: receipt });
    await waitFor(() =>
      expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(0),
    );
    expect(f.execute.mock.calls[0]?.[0].unit).toEqual(captured);
    expect(f.execute).toHaveBeenCalledTimes(later === "A" ? 1 : 2);
    if (later !== "A")
      expect(f.execute.mock.calls[1]?.[0].unit).toMatchObject({
        kind: "patch",
        recordId,
        payloadIntent: {
          base_row_version: 1,
          changes: [
            { field_key: "timeline.activity_synopsis_text", value: later },
            ...(later === "B"
              ? [
                  {
                    field_key: "timeline.analyst_text",
                    value: "Second owned field",
                  },
                ]
              : []),
          ],
        },
      });
    expect(
      f.drafts.draftValue({ ...identity, rowKey: recordId }),
    ).toBeUndefined();
    expect(
      f.drafts.draftValue({
        ...identity,
        rowKey: recordId,
        surface: "inspector",
      }),
    ).toBe("Independent inspector");
    f.detach();
  }
});

it("coalesces only undispatched capture revisions and preserves an explicit clear before capture", async () => {
  const f = fixture();
  f.runtime.pendingQueue().model.pauseForAuthRecovery();
  f.type(undefined, "A");
  f.type(undefined, "");
  expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(1);
  f.runtime.pendingQueue().model.resumeAfterAuthRecovery();
  f.runtime.requestDrain();
  await waitFor(() => expect(f.execute).toHaveBeenCalledOnce());
  expect(f.execute.mock.calls[0]?.[0].unit.payloadIntent).toEqual({
    client_txn_id: expect.any(String),
  });
  f.acknowledgement.resolve({ kind: "accepted", value: f.receipt });
  await waitFor(() =>
    expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(0),
  );
  expect(f.execute).toHaveBeenCalledOnce();
  f.detach();
});

it("retains newer authoring and blocks orphan successors after creation discard", async () => {
  const f = fixture();
  f.type(undefined, "A");
  await waitFor(() => expect(f.execute).toHaveBeenCalledOnce());
  f.type(undefined, "B");
  f.drafts.setDraft(
    { rowKey: f.row.key, field: "activitySynopsisText", surface: "grid" },
    "C refused or unsubmitted",
    f.row,
    true,
  );
  f.acknowledgement.resolve({
    kind: "rejected",
    failure: {
      kind: "validation",
      publicCode: "invalid_request",
      message: "Invalid creation",
    },
  });
  await waitFor(() =>
    expect(f.runtime.getSnapshot().blockedEdit).not.toBeNull(),
  );
  const id = f.runtime.getSnapshot().blockedEdit?.unitId;
  if (!id) throw new Error("Missing blocked create");
  expect(f.owner.discardBlockedEdit(id)).toBe(true);
  await waitFor(() =>
    expect(f.runtime.getSnapshot().blockedEdit).not.toBeNull(),
  );
  expect(f.runtime.getSnapshot().blockedEdit?.unitId).not.toBe(id);
  expect(f.execute).toHaveBeenCalledOnce();
  expect(
    f.drafts.draftValue({
      rowKey: f.row.key,
      field: "activitySynopsisText",
      surface: "grid",
    }),
  ).toBe("C refused or unsubmitted");
  f.detach();
});

it("retains admission-refused capture authoring through acceptance at queue capacity", async () => {
  const f = fixture();
  f.type(undefined, "A");
  await waitFor(() => expect(f.execute).toHaveBeenCalledOnce());
  const first = f.runtime.pendingQueue().model.snapshot().units[0];
  if (!first) throw new Error("Missing initial create");
  for (let index = 1; index < 64; index++)
    f.runtime.pendingQueue().model.admit({
      ...first,
      id: `capacity-${index}`,
      rowKey: `other-${index}`,
      clientTxnId: `capacity-${index}`,
      mutationSignature: `capacity-${index}`,
      coalesceKey: `other-${index}`,
      enqueueOrder: index + 1,
      status: "queued",
    });
  const refused = vi.fn();
  f.type(refused, "B refused Ω");
  expect(refused).toHaveBeenCalledWith(
    expect.objectContaining({ kind: "rejected_mutation" }),
  );
  expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(64);
  // Fixture ballast has no source driver; keep dispatch fenced to the real create.
  f.runtime
    .pendingQueue()
    .model.setDispatchGuard((unit) => unit.id === first.id);
  f.acknowledgement.resolve({ kind: "accepted", value: f.receipt });
  await waitFor(() => expect(f.drafts.resolveRowKey(f.row.key)).toBe(recordId));
  expect(
    f.drafts.draftValue({
      rowKey: recordId,
      field: "activitySynopsisText",
      surface: "grid",
    }),
  ).toBe("B refused Ω");
  expect(f.execute).toHaveBeenCalledOnce();
  f.detach();
});

it("retains capture authoring through suspension and retires it before late incident or account receipts", async () => {
  for (const retirement of ["incident_changed", "runtime_disposed"] as const) {
    const f = fixture();
    f.type(undefined, "A");
    await waitFor(() => expect(f.execute).toHaveBeenCalledOnce());
    const identity = {
      rowKey: f.row.key,
      field: "activitySynopsisText" as const,
      surface: "grid" as const,
    };
    f.drafts.setDraft(identity, "Private B", f.row, true);
    f.runtime.applyAuthorizationRecoveryState("paused");
    expect(f.drafts.draftValue(identity)).toBe("Private B");
    f.runtime.applyAuthorizationRecoveryState("resumed");
    expect(f.drafts.draftValue(identity)).toBe("Private B");
    f.runtime.invalidate({ kind: "incident_role_changed", role: "viewer" });
    expect(f.drafts.draftValue(identity)).toBe("Private B");
    f.runtime.invalidate({ kind: "incident_closed" });
    expect(f.drafts.draftValue(identity)).toBe("Private B");
    f.runtime.invalidate(
      retirement === "incident_changed"
        ? { kind: retirement, nextIncidentId: "other" }
        : { kind: retirement },
    );
    expect(f.drafts.draftValue(identity)).toBeUndefined();
    expect(f.runtime.pendingQueue().model.snapshot().units).toHaveLength(0);
    f.acknowledgement.resolve({ kind: "accepted", value: f.receipt });
    await f.acknowledgement.promise;
    await Promise.resolve();
    expect(f.drafts.resolveRowKey(f.row.key)).toBe(f.row.key);
    expect(
      f.drafts.draftValue({ ...identity, rowKey: recordId }),
    ).toBeUndefined();
    expect(f.owner.fileDrafts.resolve(f.row.key)).toEqual({
      kind: "unavailable",
    });
    expect(f.execute).toHaveBeenCalledOnce();
  }
});
