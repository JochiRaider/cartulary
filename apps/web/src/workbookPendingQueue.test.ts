import { describe, expect, it } from "vitest";
import {
  type PendingQueueAdmissionResult,
  type PendingReplayOperationClass,
  type PendingReplayPresentationHint,
  type PendingReplaySource,
  type PendingReplayUnitInput,
  type PendingReplayUnitState,
  type PendingReplayVisibleEdit,
  pendingReplayCapacity,
  WorkbookPendingQueueModel,
} from "./workbookPendingQueue";

const incidentId = "incident-fe-p4";
const clientInstanceId = "client-instance-fe-p4";
const viewSchemaId = "cartulary.view.timeline.v1";

function createQueue() {
  return new WorkbookPendingQueueModel({ incidentId, clientInstanceId });
}

function expectAccepted(
  result: PendingQueueAdmissionResult,
): PendingReplayUnitState {
  expect(result.accepted).toBe(true);
  if (!result.accepted) {
    throw new Error(`expected accepted admission, got ${result.status}`);
  }
  return result.unit;
}

function createUnit(options: {
  readonly clientTxnId: string;
  readonly rowKey: string;
  readonly order: number;
  readonly source?: PendingReplaySource;
  readonly payloadIntent?: Record<string, unknown>;
  readonly operationClass?: PendingReplayOperationClass;
  readonly presentationHint?: PendingReplayPresentationHint;
  readonly visibleEdit?: PendingReplayVisibleEdit;
}): PendingReplayUnitInput {
  const payloadIntent = options.payloadIntent ?? {
    client_txn_id: options.clientTxnId,
    "timeline.summary": `${options.rowKey} summary`,
  };
  const unit: PendingReplayUnitInput = {
    id: `unit-${options.clientTxnId}`,
    kind: "create",
    source: options.source ?? "autosave",
    incidentId,
    clientInstanceId,
    viewSchemaId,
    rowKey: options.rowKey,
    recordId: null,
    method: "POST",
    path: `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/rows`,
    payloadIntent,
    clientTxnId: options.clientTxnId,
    coalesceKey: `draft:${options.rowKey}`,
    enqueueOrder: options.order,
  };
  if (options.operationClass !== undefined) {
    unit.operationClass = options.operationClass;
  }
  if (options.presentationHint !== undefined) {
    unit.presentationHint = options.presentationHint;
  }
  if (options.visibleEdit !== undefined) {
    unit.visibleEdit = options.visibleEdit;
  }
  return unit;
}

function patchUnit(options: {
  readonly clientTxnId: string;
  readonly recordId: string;
  readonly rowKey?: string;
  readonly order: number;
  readonly source?: PendingReplaySource;
  readonly baseRowVersion?: number;
  readonly fieldKey?: string;
  readonly value?: unknown;
  readonly actionPayload?: unknown;
  readonly operationClass?: PendingReplayOperationClass;
  readonly presentationHint?: PendingReplayPresentationHint;
  readonly visibleEdit?: PendingReplayVisibleEdit;
}): PendingReplayUnitInput {
  const fieldKey = options.fieldKey ?? "timeline.summary";
  const change =
    options.actionPayload === undefined
      ? { field_key: fieldKey, value: options.value ?? options.clientTxnId }
      : { field_key: fieldKey, action_payload: options.actionPayload };
  const unit: PendingReplayUnitInput = {
    id: `unit-${options.clientTxnId}`,
    kind: "patch",
    source: options.source ?? "autosave",
    incidentId,
    clientInstanceId,
    viewSchemaId,
    rowKey: options.rowKey ?? `row-${options.recordId}`,
    recordId: options.recordId,
    method: "PATCH",
    path: `/api/v1/records/${options.recordId}`,
    payloadIntent: {
      view_schema_id: viewSchemaId,
      base_row_version: options.baseRowVersion ?? 1,
      client_txn_id: options.clientTxnId,
      changes: [change],
    },
    clientTxnId: options.clientTxnId,
    coalesceKey: `record:${options.recordId}`,
    enqueueOrder: options.order,
  };
  if (options.operationClass !== undefined) {
    unit.operationClass = options.operationClass;
  }
  if (options.presentationHint !== undefined) {
    unit.presentationHint = options.presentationHint;
  }
  if (options.visibleEdit !== undefined) {
    unit.visibleEdit = options.visibleEdit;
  }
  return unit;
}

describe("FE-U-P4-01 pending queue unit model", () => {
  it("FE-U-P4-01 admits row-create, row-patch, and paste-derived replay units with route-scoped mutation identity", () => {
    const queue = createQueue();

    const createAdmission = expectAccepted(
      queue.admit(
        createUnit({
          clientTxnId: "txn-create",
          rowKey: "draft-1",
          order: 1,
        }),
      ),
    );
    const patchAdmission = expectAccepted(
      queue.admit({
        ...patchUnit({
          clientTxnId: "txn-patch",
          recordId: "record-1",
          order: 2,
          baseRowVersion: 7,
          fieldKey: "timeline.summary",
          value: "patched summary",
        }),
        payloadIntent: {
          view_schema_id: viewSchemaId,
          base_row_version: 7,
          client_txn_id: "txn-patch",
          changes: [
            { field_key: "timeline.summary", value: "patched summary" },
            { field_key: "timeline.capture_state", value: "rough" },
          ],
        },
      }),
    );
    const pasteCreateAdmission = expectAccepted(
      queue.admit(
        createUnit({
          clientTxnId: "txn-paste-create",
          rowKey: "paste-draft-1",
          order: 3,
          source: "paste",
        }),
      ),
    );
    const pastePatchAdmission = expectAccepted(
      queue.admit(
        patchUnit({
          clientTxnId: "txn-paste-patch",
          recordId: "record-2",
          order: 4,
          source: "paste",
          baseRowVersion: 4,
        }),
      ),
    );

    expect(createAdmission.identity).toEqual({
      kind: "create",
      method: "POST",
      route_scope: {
        incident_id: incidentId,
        path: `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/rows`,
        view_schema_id: viewSchemaId,
      },
      client_txn_id: "txn-create",
    });
    expect(patchAdmission.identity).toMatchObject({
      kind: "patch",
      method: "PATCH",
      route_scope: {
        path: "/api/v1/records/record-1",
        record_id: "record-1",
      },
      record_id: "record-1",
      client_txn_id: "txn-patch",
      view_schema_id: viewSchemaId,
      base_row_version: 7,
    });
    expect(
      patchAdmission.identity.kind === "patch"
        ? patchAdmission.identity.changes.map((change) => change.field_key)
        : [],
    ).toEqual(["timeline.capture_state", "timeline.summary"]);
    expect(pasteCreateAdmission.source).toBe("paste");
    expect(pasteCreateAdmission.identity.kind).toBe("create");
    expect(pastePatchAdmission.source).toBe("paste");
    expect(pastePatchAdmission.identity.kind).toBe("patch");
    expect(queue.snapshot().units.map((unit) => unit.clientTxnId)).toEqual([
      "txn-create",
      "txn-patch",
      "txn-paste-create",
      "txn-paste-patch",
    ]);
  });

  it("FE-U-P4-01 preserves FIFO replay order and refuses the 65th non-coalescible unit without eviction", () => {
    const queue = createQueue();
    const admittedTxnIds: string[] = [];

    for (let index = 1; index <= pendingReplayCapacity; index += 1) {
      const txnId = `txn-${String(index).padStart(2, "0")}`;
      const admission = queue.admit(
        patchUnit({
          clientTxnId: txnId,
          recordId: `record-${index}`,
          order: index,
          visibleEdit: {
            rowKey: `row-record-${index}`,
            fieldKey: "timeline.summary",
            value: `visible ${index}`,
          },
        }),
      );
      expectAccepted(admission);
      admittedTxnIds.push(txnId);
    }

    const overflow = queue.admit(
      patchUnit({
        clientTxnId: "txn-65",
        recordId: "record-65",
        order: 65,
        visibleEdit: {
          rowKey: "row-record-65",
          fieldKey: "timeline.summary",
          value: "visible overflow value",
        },
      }),
    );
    expect(overflow.accepted).toBe(false);
    expect(overflow.status).toBe("refused");
    if (overflow.status !== "refused") {
      throw new Error("expected overflow refusal");
    }
    expect(overflow.refusedReason).toBe("capacity");
    expect(overflow.preserveVisibleEditAsUnsaved).toBe(true);
    expect(overflow.primarySaveStateInput).toBe("Conflict");
    expect(overflow.snapshot.overflow).toEqual({
      message:
        "Local pending queue is full. The current edit remains unsaved local work.",
      refused_unit_id: "unit-txn-65",
      preserve_visible_edit_as_unsaved: true,
      visible_edit: {
        rowKey: "row-record-65",
        fieldKey: "timeline.summary",
        value: "visible overflow value",
      },
    });
    expect(queue.snapshot().units.map((unit) => unit.clientTxnId)).toEqual(
      admittedTxnIds,
    );

    const replayedTxnIds: string[] = [];
    for (let index = 1; index <= pendingReplayCapacity; index += 1) {
      const dispatch = queue.dispatchNext();
      expect(dispatch).not.toBeNull();
      if (dispatch === null) {
        throw new Error("expected replay dispatch");
      }
      replayedTxnIds.push(dispatch.unit.clientTxnId);
      queue.settleDispatched({
        ok: true,
        row: {
          record_id: dispatch.unit.recordId ?? "created-record",
          row_version: index + 1,
        },
      });
    }
    expect(replayedTxnIds).toEqual(admittedTxnIds);
    expect(queue.snapshot().units).toHaveLength(0);
  });

  it("FE-U-P4-01 coalesces only same-draft creates and contiguous same-record patches", () => {
    const createQueueModel = createQueue();
    const createFirst = expectAccepted(
      createQueueModel.admit(
        createUnit({
          clientTxnId: "txn-create-1",
          rowKey: "draft-a",
          order: 1,
          payloadIntent: {
            client_txn_id: "txn-create-1",
            "timeline.summary": "first draft",
            "timeline.evidence": {
              kind: "collection_actions_v1",
              actions: [{ op: "add_record_ref", linked_record_id: "e1" }],
            },
          },
        }),
      ),
    );
    expect(createFirst.clientTxnId).toBe("txn-create-1");
    const createSecond = createQueueModel.admit(
      createUnit({
        clientTxnId: "txn-create-2",
        rowKey: "draft-a",
        order: 2,
        payloadIntent: {
          client_txn_id: "txn-create-2",
          "timeline.summary": "second draft",
          "timeline.evidence": {
            kind: "collection_actions_v1",
            actions: [{ op: "add_record_ref", linked_record_id: "e2" }],
          },
        },
      }),
    );
    expect(createSecond.status).toBe("coalesced");
    const coalescedCreate = createQueueModel.snapshot().units[0];
    expect(coalescedCreate?.payloadIntent).toEqual({
      client_txn_id: "txn-create-1",
      "timeline.evidence": {
        kind: "collection_actions_v1",
        actions: [
          { op: "add_record_ref", linked_record_id: "e1" },
          { op: "add_record_ref", linked_record_id: "e2" },
        ],
      },
      "timeline.summary": "second draft",
    });

    const patchQueueModel = createQueue();
    expectAccepted(
      patchQueueModel.admit(
        patchUnit({
          clientTxnId: "txn-action-1",
          recordId: "record-a",
          order: 1,
          fieldKey: "timeline.evidence",
          actionPayload: {
            kind: "collection_actions_v1",
            actions: [{ op: "add_record_ref", linked_record_id: "e1" }],
          },
        }),
      ),
    );
    const patchSecond = patchQueueModel.admit(
      patchUnit({
        clientTxnId: "txn-action-2",
        recordId: "record-a",
        order: 2,
        fieldKey: "timeline.evidence",
        actionPayload: {
          kind: "collection_actions_v1",
          actions: [{ op: "add_record_ref", linked_record_id: "e2" }],
        },
      }),
    );
    expect(patchSecond.status).toBe("coalesced");
    const coalescedPatch = patchQueueModel.snapshot().units[0];
    expect(coalescedPatch?.payloadIntent.changes).toEqual([
      {
        field_key: "timeline.evidence",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            { op: "add_record_ref", linked_record_id: "e1" },
            { op: "add_record_ref", linked_record_id: "e2" },
          ],
        },
      },
    ]);

    const interleavedQueue = createQueue();
    expectAccepted(
      interleavedQueue.admit(
        patchUnit({
          clientTxnId: "txn-a1",
          recordId: "record-a",
          order: 1,
          value: "A1",
        }),
      ),
    );
    expectAccepted(
      interleavedQueue.admit(
        patchUnit({
          clientTxnId: "txn-b1",
          recordId: "record-b",
          order: 2,
          value: "B1",
        }),
      ),
    );
    expectAccepted(
      interleavedQueue.admit(
        patchUnit({
          clientTxnId: "txn-a2",
          recordId: "record-a",
          order: 3,
          value: "A2",
        }),
      ),
    );
    expect(
      interleavedQueue.snapshot().units.map((unit) => unit.clientTxnId),
    ).toEqual(["txn-a1", "txn-b1", "txn-a2"]);
  });

  it("FE-U-P4-01 preserves same-runtime retryable failures and does not survive a new page instance", () => {
    const retryQueue = createQueue();
    expectAccepted(
      retryQueue.admit(
        patchUnit({
          clientTxnId: "txn-retry",
          recordId: "record-retry",
          order: 1,
        }),
      ),
    );

    const firstDispatch = retryQueue.dispatchNext();
    expect(firstDispatch?.unit.clientTxnId).toBe("txn-retry");
    const retryResult = retryQueue.settleDispatched({
      ok: false,
      status: 503,
      error: {
        code: "temporary_unavailable",
        message: "Temporary public failure",
      },
    });
    expect(retryResult.outcome).toBe("retryable_failure");
    expect(retryQueue.snapshot().queuedCount).toBe(1);
    expect(retryQueue.snapshot().primarySaveStateInput).toBe("Syncing");
    expect(retryQueue.dispatchNext()?.unit.clientTxnId).toBe("txn-retry");

    const authQueue = createQueue();
    expectAccepted(
      authQueue.admit(
        patchUnit({
          clientTxnId: "txn-session",
          recordId: "record-session",
          order: 1,
        }),
      ),
    );
    expect(authQueue.dispatchNext()?.unit.clientTxnId).toBe("txn-session");
    const authResult = authQueue.settleDispatched({
      ok: false,
      status: 401,
      error: {
        code: "session_revoked",
        message: "Session revoked",
      },
    });
    expect(authResult.outcome).toBe("auth_paused");
    expect(authQueue.snapshot().authPaused).toBe(true);
    expect(authQueue.snapshot().queuedCount).toBe(1);
    expect(authQueue.dispatchNext()).toBeNull();
    authQueue.resumeAfterAuthRecovery();
    expect(authQueue.dispatchNext()?.unit.clientTxnId).toBe("txn-session");

    const recreatedPageInstance = createQueue();
    expect(recreatedPageInstance.snapshot().units).toHaveLength(0);
    expect(recreatedPageInstance.snapshot().primarySaveStateInput).toBe(
      "Saved",
    );
  });

  it("FE-U-P4-01 halts non-retryable public failures with Table G anchors and queued units behind the blocker", () => {
    const validationQueue = createQueue();
    expectAccepted(
      validationQueue.admit(
        patchUnit({
          clientTxnId: "txn-invalid",
          recordId: "record-invalid",
          order: 1,
        }),
      ),
    );
    expectAccepted(
      validationQueue.admit(
        patchUnit({
          clientTxnId: "txn-behind-invalid",
          recordId: "record-behind",
          order: 2,
        }),
      ),
    );
    expect(validationQueue.dispatchNext()?.unit.clientTxnId).toBe(
      "txn-invalid",
    );
    const validationResult = validationQueue.settleDispatched({
      ok: false,
      status: 400,
      error: {
        code: "invalid_mutation_payload",
        message: "Summary is required",
        details: {
          record_id: "record-invalid",
          field_key: "timeline.summary",
        },
      },
    });
    expect(validationResult.outcome).toBe("halted");
    if (validationResult.outcome !== "halted") {
      throw new Error("expected validation halt");
    }
    expect(validationResult.halt).toMatchObject({
      unit_id: "unit-txn-invalid",
      error_code: "invalid_mutation_payload",
      message: "Summary is required",
      anchor: {
        kind: "cell",
        record_id: "record-invalid",
        field_key: "timeline.summary",
      },
    });
    expect(
      validationResult.snapshot.units.map((unit) => unit.clientTxnId),
    ).toEqual(["txn-invalid", "txn-behind-invalid"]);
    expect(validationQueue.dispatchNext()).toBeNull();
    expect(validationQueue.snapshot().primarySaveStateInput).toBe("Conflict");

    const clientTxnConflictQueue = createQueue();
    expectAccepted(
      clientTxnConflictQueue.admit(
        patchUnit({
          clientTxnId: "txn-client-conflict",
          recordId: "record-client-conflict",
          order: 1,
        }),
      ),
    );
    clientTxnConflictQueue.dispatchNext();
    const clientTxnConflict = clientTxnConflictQueue.settleDispatched({
      ok: false,
      status: 409,
      error: {
        code: "client_txn_conflict",
        details: { client_txn_id: "txn-client-conflict" },
      },
    });
    expect(clientTxnConflict.outcome).toBe("halted");
    if (clientTxnConflict.outcome !== "halted") {
      throw new Error("expected client transaction conflict halt");
    }
    expect(clientTxnConflict.halt.anchor.kind).toBe("mutation");

    const rowVersionConflictQueue = createQueue();
    expectAccepted(
      rowVersionConflictQueue.admit(
        patchUnit({
          clientTxnId: "txn-row-version",
          recordId: "record-row-version",
          order: 1,
        }),
      ),
    );
    rowVersionConflictQueue.dispatchNext();
    const rowVersionConflict = rowVersionConflictQueue.settleDispatched({
      ok: false,
      status: 409,
      error: {
        code: "row_version_conflict",
        details: { record_id: "record-row-version" },
      },
    });
    expect(rowVersionConflict.outcome).toBe("halted");
    if (rowVersionConflict.outcome !== "halted") {
      throw new Error("expected row-version conflict halt");
    }
    expect(rowVersionConflict.halt.anchor).toEqual({
      kind: "record",
      record_id: "record-row-version",
    });

    const sameFieldQueue = createQueue();
    expectAccepted(
      sameFieldQueue.admit(
        patchUnit({
          clientTxnId: "txn-same-field",
          recordId: "record-same-field",
          order: 1,
        }),
      ),
    );
    expectAccepted(
      sameFieldQueue.admit(
        patchUnit({
          clientTxnId: "txn-behind-same-field",
          recordId: "record-behind-same-field",
          order: 2,
        }),
      ),
    );
    sameFieldQueue.dispatchNext();
    const sameFieldConflict = sameFieldQueue.settleDispatched({
      ok: false,
      status: 409,
      error: {
        code: "same_field_conflict",
        details: {
          record_id: "record-same-field",
          field_key: "timeline.summary",
        },
      },
    });
    expect(sameFieldConflict.outcome).toBe("same_field_conflict");
    if (sameFieldConflict.outcome !== "same_field_conflict") {
      throw new Error("expected same-field conflict");
    }
    expect(sameFieldConflict.conflict).toEqual({
      key: "record-same-field:timeline.summary",
      record_id: "record-same-field",
      field_key: "timeline.summary",
    });
    expect(
      sameFieldConflict.snapshot.units.map((unit) => unit.clientTxnId),
    ).toEqual(["txn-behind-same-field"]);
    expect(sameFieldQueue.dispatchNext()).toBeNull();
    expect(sameFieldQueue.snapshot().primarySaveStateInput).toBe("Conflict");
  });

  it("FE-U-P4-01 applies successful replay without retargeting by visible row order or labels", () => {
    const queue = createQueue();
    expectAccepted(
      queue.admit(
        patchUnit({
          clientTxnId: "txn-record-a",
          recordId: "record-a",
          order: 1,
          value: "A mutation",
          presentationHint: {
            label: "Z visible label",
            recordType: "later-visible-type",
            sortRank: 99,
            visibleRowIndex: 20,
          },
        }),
      ),
    );
    expectAccepted(
      queue.admit(
        patchUnit({
          clientTxnId: "txn-record-b",
          recordId: "record-b",
          order: 2,
          value: "B mutation",
          presentationHint: {
            label: "A visible label",
            recordType: "earlier-visible-type",
            sortRank: 1,
            visibleRowIndex: 0,
          },
        }),
      ),
    );

    const firstDispatch = queue.dispatchNext();
    expect(firstDispatch?.unit.recordId).toBe("record-a");
    expect(firstDispatch?.identity.kind).toBe("patch");
    expect(JSON.stringify(firstDispatch?.identity)).not.toContain(
      "visible label",
    );
    const firstSuccess = queue.settleDispatched({
      ok: true,
      row: {
        record_id: "record-a",
        row_version: 12,
      },
    });
    expect(firstSuccess.outcome).toBe("success");
    if (firstSuccess.outcome !== "success") {
      throw new Error("expected successful replay");
    }
    expect(firstSuccess.unit.recordId).toBe("record-a");
    expect(firstSuccess.row).toEqual({
      record_id: "record-a",
      row_version: 12,
    });
    expect(queue.dispatchNext()?.unit.recordId).toBe("record-b");
  });
});
