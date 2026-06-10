import { describe, expect, it } from "vitest";
import {
  deriveWorkbookSaveState,
  type PendingQueueAdmissionResult,
  type PendingReplayOperationClass,
  type PendingReplayPresentationHint,
  type PendingReplaySource,
  type PendingReplayUnitInput,
  type PendingReplayUnitState,
  type PendingReplayVisibleEdit,
  parsePendingReplayPublicError,
  pendingReplayCapacity,
  sameFieldConflictQueueKey,
  WorkbookPendingQueueModel,
  workbookSaveStateConflictAnchorIdentity,
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

function sameFieldConflictError(options: {
  readonly conflictToken: string;
  readonly recordId: string;
  readonly fieldKey: string;
  readonly baseRowVersion?: number;
  readonly currentRowVersion?: number;
}) {
  return {
    code: "same_field_conflict",
    message: "Public same-field conflict",
    conflict: {
      conflict_token: options.conflictToken,
      record_id: options.recordId,
      field_key: options.fieldKey,
      conflict_resolution_class: "text_compare_merge",
      base_row_version: options.baseRowVersion ?? 1,
      current_row_version: options.currentRowVersion ?? 2,
      client_value: "local draft",
      server_value: "server value",
      server_updated_by: "user-server",
      server_updated_at: "2026-06-02T00:00:00Z",
      base_value: "base value",
    },
  };
}

const allowedPrimarySaveStateLabels = ["Conflict", "Saved", "Syncing"];
const forbiddenPrimarySaveStateLabels = [
  "Failed",
  "Pending",
  "Queued",
  "Replay halted",
  "Retrying",
];

function expectCorePrimaryLabel(label: string) {
  expect(allowedPrimarySaveStateLabels).toContain(label);
  expect(forbiddenPrimarySaveStateLabels).not.toContain(label);
}

describe("FE-U-P4-01 pending queue unit model", () => {
  it("FE-U-P4-01 keeps pending queues isolated by incident and client instance scope", () => {
    const queue = createQueue();
    const otherClientQueue = new WorkbookPendingQueueModel({
      incidentId,
      clientInstanceId: "client-instance-other",
    });

    expectAccepted(
      queue.admit(
        patchUnit({
          clientTxnId: "txn-scope-a",
          recordId: "record-scope-a",
          order: 1,
        }),
      ),
    );
    expectAccepted(
      otherClientQueue.admit({
        ...patchUnit({
          clientTxnId: "txn-scope-b",
          recordId: "record-scope-b",
          order: 1,
        }),
        clientInstanceId: "client-instance-other",
      }),
    );

    expect(queue.snapshot().queuedCount).toBe(1);
    expect(otherClientQueue.snapshot().queuedCount).toBe(1);
    expect(queue.dispatchNext()?.unit.clientTxnId).toBe("txn-scope-a");
    expect(otherClientQueue.dispatchNext()?.unit.clientTxnId).toBe(
      "txn-scope-b",
    );

    const foreignIncidentAdmission = queue.admit({
      ...patchUnit({
        clientTxnId: "txn-foreign-incident",
        recordId: "record-foreign-incident",
        order: 2,
      }),
      incidentId: "incident-other",
    });
    expect(foreignIncidentAdmission.accepted).toBe(false);
    expect(foreignIncidentAdmission.status).toBe("refused");
    if (foreignIncidentAdmission.status !== "refused") {
      throw new Error("expected scope mismatch refusal");
    }
    expect(foreignIncidentAdmission.refusedReason).toBe("scope_mismatch");
    expect(foreignIncidentAdmission.preserveVisibleEditAsUnsaved).toBe(false);
  });

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
    const apiQueue = createQueue();
    expectAccepted(
      apiQueue.admit(
        patchUnit({
          clientTxnId: "txn-api-1",
          recordId: "record-api-1",
          order: 1,
        }),
      ),
    );
    expectAccepted(
      apiQueue.admit(
        patchUnit({
          clientTxnId: "txn-api-2",
          recordId: "record-api-2",
          order: 2,
        }),
      ),
    );
    const peeked = apiQueue.peekNextQueued();
    expect(peeked?.unit.clientTxnId).toBe("txn-api-1");
    expect(apiQueue.snapshot().inFlightCount).toBe(0);
    expect(apiQueue.markDispatched("unit-txn-api-2")).toBeNull();
    const marked = apiQueue.markDispatched(peeked?.unit.id ?? "");
    expect(marked?.unit.clientTxnId).toBe("txn-api-1");
    expect(
      apiQueue.snapshot().units.map((unit) => [unit.clientTxnId, unit.status]),
    ).toEqual([
      ["txn-api-1", "in_flight"],
      ["txn-api-2", "queued"],
    ]);
    expect(apiQueue.peekNextQueued()).toBeNull();
    apiQueue.settleDispatched({
      ok: true,
      row: { record_id: "record-api-1", row_version: 2 },
    });
    expect(apiQueue.peekNextQueued()?.unit.clientTxnId).toBe("txn-api-2");

    const materializationRetryQueue = createQueue();
    expectAccepted(
      materializationRetryQueue.admit(
        patchUnit({
          clientTxnId: "txn-materialize",
          recordId: "record-materialize",
          order: 1,
        }),
      ),
    );
    const materializationPeek = materializationRetryQueue.peekNextQueued();
    expect(materializationPeek?.unit.clientTxnId).toBe("txn-materialize");
    expect(materializationRetryQueue.snapshot().inFlightCount).toBe(0);
    expect(materializationRetryQueue.peekNextQueued()?.unit.clientTxnId).toBe(
      "txn-materialize",
    );

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

    for (const operationClass of [
      "destructive",
      "conflict_resolution",
      "non_hot_path",
    ] as const) {
      const forbiddenQueue = createQueue();
      expectAccepted(
        forbiddenQueue.admit(
          patchUnit({
            clientTxnId: `txn-${operationClass}-1`,
            recordId: `record-${operationClass}`,
            order: 1,
            operationClass,
          }),
        ),
      );
      const secondAdmission = forbiddenQueue.admit(
        patchUnit({
          clientTxnId: `txn-${operationClass}-2`,
          recordId: `record-${operationClass}`,
          order: 2,
          operationClass,
        }),
      );
      expect(secondAdmission.status).toBe("accepted");
      expect(
        forbiddenQueue.snapshot().units.map((unit) => unit.clientTxnId),
      ).toEqual([`txn-${operationClass}-1`, `txn-${operationClass}-2`]);
    }
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

    const unknownRetryableQueue = createQueue();
    const parsedUnknownRetryable = parsePendingReplayPublicError({
      error: {
        code: "future_public_error",
        message: "Future public error",
        retryable: true,
      },
    });
    expect(parsedUnknownRetryable).toEqual({
      code: "future_public_error",
      message: "Future public error",
      retryable: true,
    });
    const parsedUnsafeUnknownRetryable = parsePendingReplayPublicError({
      error: {
        code: "future_public_error",
        message: "stack trace at handler (/home/cartulary/internal.go:42)",
        retryable: true,
        status: 409,
        details: {
          private_path: "/home/cartulary/internal.go",
          reason_code: "future_public_error",
        },
      },
    });
    expect(parsedUnsafeUnknownRetryable).toMatchObject({
      code: "future_public_error",
      message: "Conflict.",
      retryable: true,
    });
    expectAccepted(
      unknownRetryableQueue.admit(
        patchUnit({
          clientTxnId: "txn-unknown-retryable",
          recordId: "record-unknown-retryable",
          order: 1,
        }),
      ),
    );
    expectAccepted(
      unknownRetryableQueue.admit(
        patchUnit({
          clientTxnId: "txn-behind-unknown-retryable",
          recordId: "record-behind-unknown-retryable",
          order: 2,
        }),
      ),
    );
    expect(unknownRetryableQueue.dispatchNext()?.unit.clientTxnId).toBe(
      "txn-unknown-retryable",
    );
    const unknownRetryableResult = unknownRetryableQueue.settleDispatched({
      ok: false,
      status: 409,
      error: {
        code: "future_public_error",
        message: "Future public error",
        retryable: true,
      },
    });
    expect(unknownRetryableResult.outcome).toBe("retryable_failure");
    expect(
      unknownRetryableResult.snapshot.units.map((unit) => unit.clientTxnId),
    ).toEqual(["txn-unknown-retryable", "txn-behind-unknown-retryable"]);
    expect(unknownRetryableQueue.dispatchNext()?.unit.clientTxnId).toBe(
      "txn-unknown-retryable",
    );

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

    const unknownTerminalQueue = createQueue();
    expectAccepted(
      unknownTerminalQueue.admit(
        patchUnit({
          clientTxnId: "txn-unknown-terminal",
          recordId: "record-unknown-terminal",
          order: 1,
        }),
      ),
    );
    expectAccepted(
      unknownTerminalQueue.admit(
        patchUnit({
          clientTxnId: "txn-behind-unknown-terminal",
          recordId: "record-behind-unknown-terminal",
          order: 2,
        }),
      ),
    );
    unknownTerminalQueue.dispatchNext();
    const unknownTerminalResult = unknownTerminalQueue.settleDispatched({
      ok: false,
      status: 409,
      error: {
        code: "future_terminal_public_error",
        message: "Future terminal public error",
      },
    });
    expect(unknownTerminalResult.outcome).toBe("halted");
    if (unknownTerminalResult.outcome !== "halted") {
      throw new Error("expected unknown terminal public error halt");
    }
    expect(
      unknownTerminalResult.snapshot.units.map((unit) => unit.clientTxnId),
    ).toEqual(["txn-unknown-terminal", "txn-behind-unknown-terminal"]);
    expect(unknownTerminalQueue.dispatchNext()).toBeNull();
    expect(unknownTerminalResult.halt.error_code).toBe(
      "future_terminal_public_error",
    );
    expect(unknownTerminalResult.halt.message).toBe(
      "Future terminal public error",
    );

    const unsafeUnknownTerminalQueue = createQueue();
    expectAccepted(
      unsafeUnknownTerminalQueue.admit(
        patchUnit({
          clientTxnId: "txn-unsafe-unknown-terminal",
          recordId: "record-unsafe-unknown-terminal",
          order: 1,
        }),
      ),
    );
    unsafeUnknownTerminalQueue.dispatchNext();
    const unsafeUnknownTerminalResult =
      unsafeUnknownTerminalQueue.settleDispatched({
        ok: false,
        status: 418,
        error: {
          code: "future_terminal_public_error",
          message:
            "stack trace at handler (/home/cartulary/internal/private.go:42)",
          details: {
            private_path: "/home/cartulary/internal/private.go",
            reason_code: "future_terminal_public_error",
          },
        },
      });
    expect(unsafeUnknownTerminalResult.outcome).toBe("halted");
    if (unsafeUnknownTerminalResult.outcome !== "halted") {
      throw new Error("expected unsafe unknown terminal public error halt");
    }
    expect(unsafeUnknownTerminalResult.halt.error_code).toBe(
      "future_terminal_public_error",
    );
    expect(unsafeUnknownTerminalResult.halt.message).toBe("Request failed.");
    expect(JSON.stringify(unsafeUnknownTerminalResult.halt)).not.toContain(
      "/home/cartulary",
    );

    const missingCodeMessage = parsePendingReplayPublicError({
      error: {
        status: 500,
        details: {
          private_path: "/home/cartulary/internal/private.go",
        },
      },
    });
    expect(missingCodeMessage.code).toBe("unknown_public_error");
    expect(missingCodeMessage.message).toBeUndefined();

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
      error: sameFieldConflictError({
        conflictToken: "conflict-token-same-field",
        recordId: "record-same-field",
        fieldKey: "timeline.summary",
        baseRowVersion: 7,
        currentRowVersion: 8,
      }),
    });
    expect(sameFieldConflict.outcome).toBe("same_field_conflict");
    if (sameFieldConflict.outcome !== "same_field_conflict") {
      throw new Error("expected same-field conflict");
    }
    expect(sameFieldConflict.conflict).toEqual({
      key: "record-same-field:timeline.summary",
      conflict_token: "conflict-token-same-field",
      record_id: "record-same-field",
      field_key: "timeline.summary",
      conflict_resolution_class: "text_compare_merge",
      base_row_version: 7,
      current_row_version: 8,
    });
    expect(
      sameFieldConflict.snapshot.units.map((unit) => unit.clientTxnId),
    ).toEqual(["txn-behind-same-field"]);
    expect(sameFieldQueue.dispatchNext()).toBeNull();
    expect(sameFieldQueue.snapshot().primarySaveStateInput).toBe("Conflict");
    sameFieldQueue.clearSameFieldConflict("record-same-field:timeline.summary");
    expect(sameFieldQueue.snapshot().sameFieldConflicts).toEqual([]);
    expect(sameFieldQueue.dispatchNext()?.unit.clientTxnId).toBe(
      "txn-behind-same-field",
    );

    const detailsOnlySameFieldQueue = createQueue();
    expectAccepted(
      detailsOnlySameFieldQueue.admit(
        patchUnit({
          clientTxnId: "txn-details-only-same-field",
          recordId: "record-details-only-same-field",
          order: 1,
        }),
      ),
    );
    detailsOnlySameFieldQueue.dispatchNext();
    const detailsOnlySameField = detailsOnlySameFieldQueue.settleDispatched({
      ok: false,
      status: 409,
      error: {
        code: "same_field_conflict",
        details: {
          record_id: "record-details-only-same-field",
          field_key: "timeline.summary",
        },
      },
    });
    expect(detailsOnlySameField.outcome).toBe("halted");
    if (detailsOnlySameField.outcome !== "halted") {
      throw new Error("expected details-only same-field conflict halt");
    }
    expect(detailsOnlySameField.snapshot.sameFieldConflicts).toEqual([]);
    expect(
      detailsOnlySameField.snapshot.units.map((unit) => unit.clientTxnId),
    ).toEqual(["txn-details-only-same-field"]);
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

describe("FE-U-P4-02 save-state unit model", () => {
  it("FE-U-P4-02 derives exactly one primary save-state label for every Table C input condition", () => {
    const sameFieldConflict = {
      key: "record-conflict:timeline.summary",
      conflict_token: "conflict-token",
      record_id: "record-conflict",
      field_key: "timeline.summary",
      conflict_resolution_class: "text_compare_merge",
      base_row_version: 7,
      current_row_version: 8,
    };
    const overflow = {
      message:
        "Local pending queue is full. The current edit remains unsaved local work.",
      refused_unit_id: "unit-overflow",
      preserve_visible_edit_as_unsaved: true,
      visible_edit: {
        rowKey: "row-overflow",
        fieldKey: "timeline.summary",
        value: "overflow draft",
      },
    } as const;
    const halted = {
      unit_id: "unit-halted",
      error_code: "invalid_mutation_payload",
      message: "Summary is required",
      anchor: {
        kind: "cell",
        record_id: "record-halted",
        field_key: "timeline.summary",
      },
    } as const;

    const cases = [
      {
        name: "unresolved same-field conflict",
        input: {
          queuedCount: 1,
          inFlightCount: 1,
          sameFieldConflicts: [sameFieldConflict],
          overflow,
          halted,
        },
        expected: "Conflict",
        secondaryKind: "same_field_conflict",
      },
      {
        name: "queue overflow",
        input: {
          queuedCount: pendingReplayCapacity,
          inFlightCount: 0,
          overflow,
        },
        expected: "Conflict",
        secondaryKind: "overflow",
      },
      {
        name: "replay halted",
        input: {
          queuedCount: 1,
          inFlightCount: 0,
          halted,
        },
        expected: "Conflict",
        secondaryKind: "replay_halted",
      },
      {
        name: "mutation in flight",
        input: { queuedCount: 0, inFlightCount: 1 },
        expected: "Syncing",
        secondaryKind: "queued",
      },
      {
        name: "non-empty pending queue",
        input: { queuedCount: 1, inFlightCount: 0 },
        expected: "Syncing",
        secondaryKind: "queued",
      },
      {
        name: "paused replay awaiting authentication",
        input: { queuedCount: 1, inFlightCount: 0, authPaused: true },
        expected: "Syncing",
        secondaryKind: "auth_paused",
      },
      {
        name: "paused replay awaiting refresh",
        input: { queuedCount: 1, inFlightCount: 0, refreshPaused: true },
        expected: "Syncing",
        secondaryKind: "queued",
      },
      {
        name: "legacy direct mutation in progress",
        input: { queuedCount: 0, inFlightCount: 0, pendingMutationCount: 1 },
        expected: "Syncing",
        secondaryKind: "queued",
      },
      {
        name: "fully saved",
        input: { queuedCount: 0, inFlightCount: 0 },
        expected: "Saved",
        secondaryKind: null,
      },
      {
        name: "ambient presence-only update",
        input: { queuedCount: 0, inFlightCount: 0 },
        expected: "Saved",
        secondaryKind: null,
      },
    ] as const;

    for (const entry of cases) {
      const presentation = deriveWorkbookSaveState(entry.input);
      expect(presentation.primaryLabel, entry.name).toBe(entry.expected);
      expectCorePrimaryLabel(presentation.primaryLabel);
      expect(presentation.secondaryKind, entry.name).toBe(entry.secondaryKind);
      expect(
        allowedPrimarySaveStateLabels.filter(
          (label) => label === presentation.primaryLabel,
        ),
        entry.name,
      ).toHaveLength(1);
    }
  });

  it("FE-U-P4-02 keeps failure, overflow, validation, and replay details as same-surface secondary status messages", () => {
    const overflowPresentation = deriveWorkbookSaveState({
      queuedCount: pendingReplayCapacity,
      inFlightCount: 0,
      overflow: {
        message:
          "Local pending queue is full. The current edit remains unsaved local work.",
        refused_unit_id: "unit-overflow",
        preserve_visible_edit_as_unsaved: true,
        visible_edit: null,
      },
    });
    expect(overflowPresentation.primaryLabel).toBe("Conflict");
    expect(overflowPresentation.secondaryKind).toBe("overflow");
    expect(overflowPresentation.secondaryMessage).toContain(
      "current edit remains unsaved local work",
    );

    const validationPresentation = deriveWorkbookSaveState({
      queuedCount: 1,
      inFlightCount: 0,
      halted: {
        unit_id: "unit-validation",
        error_code: "invalid_mutation_payload",
        message: "Summary is required",
        anchor: {
          kind: "cell",
          record_id: "record-validation",
          field_key: "timeline.summary",
        },
      },
    });
    expect(validationPresentation.primaryLabel).toBe("Conflict");
    expect(validationPresentation.secondaryKind).toBe("replay_halted");
    expect(validationPresentation.secondaryMessage).toBe("Summary is required");

    const replayFailurePresentation = deriveWorkbookSaveState({
      queuedCount: 2,
      inFlightCount: 0,
      halted: {
        unit_id: "unit-failure",
        error_code: "future_terminal_public_error",
        message: "Future terminal public error",
        anchor: {
          kind: "mutation",
          client_txn_id: "txn-failure",
          route_scope: {
            kind: "patch",
            method: "PATCH",
            route_scope: {
              path: "/api/v1/records/record-failure",
              record_id: "record-failure",
            },
            record_id: "record-failure",
            client_txn_id: "txn-failure",
            view_schema_id: viewSchemaId,
            base_row_version: 1,
            changes: [{ field_key: "timeline.summary", value: "failure" }],
          },
        },
      },
    });
    expect(replayFailurePresentation.primaryLabel).toBe("Conflict");
    expect(replayFailurePresentation.secondaryKind).toBe("replay_halted");
    expect(replayFailurePresentation.secondaryMessage).toBe(
      "Future terminal public error",
    );

    for (const presentation of [
      overflowPresentation,
      validationPresentation,
      replayFailurePresentation,
    ]) {
      expectCorePrimaryLabel(presentation.primaryLabel);
      expect(forbiddenPrimarySaveStateLabels).not.toContain(
        presentation.secondaryMessage,
      );
    }
  });

  it("FE-U-P4-02 preserves same-field conflict anchors by record_id, field_key, and base_row_version", () => {
    const sameFieldQueue = createQueue();
    expectAccepted(
      sameFieldQueue.admit(
        patchUnit({
          clientTxnId: "txn-save-state-conflict",
          recordId: "record-save-state-conflict",
          order: 1,
          baseRowVersion: 11,
          fieldKey: "timeline.summary",
        }),
      ),
    );
    sameFieldQueue.dispatchNext();
    const settlement = sameFieldQueue.settleDispatched({
      ok: false,
      status: 409,
      error: sameFieldConflictError({
        conflictToken: "conflict-token-save-state",
        recordId: "record-save-state-conflict",
        fieldKey: "timeline.summary",
        baseRowVersion: 11,
        currentRowVersion: 12,
      }),
    });
    expect(settlement.outcome).toBe("same_field_conflict");
    expect(sameFieldQueue.snapshot().saveStatePresentation).toEqual({
      primaryLabel: "Conflict",
      secondaryKind: "same_field_conflict",
      secondaryMessage: "1 same-field conflict needs review.",
      conflictAnchors: [
        {
          record_id: "record-save-state-conflict",
          field_key: "timeline.summary",
          base_row_version: 11,
          current_row_version: 12,
        },
      ],
    });
    expect(
      sameFieldQueue.snapshot().saveStatePresentation.secondaryMessage,
    ).not.toContain("record-save-state-conflict");
    expect(
      sameFieldQueue.snapshot().saveStatePresentation.secondaryMessage,
    ).not.toContain("timeline.summary");

    const localDraftPresentation = deriveWorkbookSaveState({
      queuedCount: 0,
      inFlightCount: 0,
      localDraftConflicts: [
        {
          record_id: "record-local-draft",
          field_key: "timeline.details",
          base_row_version: 4,
          current_row_version: 5,
        },
      ],
    });
    expect(localDraftPresentation).toEqual({
      primaryLabel: "Conflict",
      secondaryKind: "same_field_conflict",
      secondaryMessage: "1 same-field conflict needs review.",
      conflictAnchors: [
        {
          record_id: "record-local-draft",
          field_key: "timeline.details",
          base_row_version: 4,
          current_row_version: 5,
        },
      ],
    });

    const duplicateSourcePresentation = deriveWorkbookSaveState({
      queuedCount: 0,
      inFlightCount: 0,
      sameFieldConflicts: [
        {
          key: "record-duplicate:timeline.summary",
          conflict_token: "conflict-duplicate",
          record_id: "record-duplicate",
          field_key: "timeline.summary",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 9,
          current_row_version: 10,
        },
      ],
      localDraftConflicts: [
        {
          record_id: "record-duplicate",
          field_key: "timeline.summary",
          base_row_version: 9,
          current_row_version: 10,
        },
      ],
    });
    expect(duplicateSourcePresentation.secondaryMessage).toBe(
      "1 same-field conflict needs review.",
    );
    expect(duplicateSourcePresentation.conflictAnchors).toEqual([
      {
        record_id: "record-duplicate",
        field_key: "timeline.summary",
        base_row_version: 9,
        current_row_version: 10,
      },
    ]);

    const multiConflictPresentation = deriveWorkbookSaveState({
      queuedCount: 3,
      inFlightCount: 1,
      sameFieldConflicts: [
        {
          key: "record-a:timeline.summary",
          conflict_token: "conflict-a",
          record_id: "record-a",
          field_key: "timeline.summary",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 1,
          current_row_version: 2,
        },
        {
          key: "record-b:timeline.details",
          conflict_token: "conflict-b",
          record_id: "record-b",
          field_key: "timeline.details",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 3,
          current_row_version: 4,
        },
      ],
      localDraftConflicts: [
        {
          record_id: "record-c",
          field_key: "timeline.tags",
          base_row_version: 5,
        },
      ],
    });
    expect(multiConflictPresentation.primaryLabel).toBe("Conflict");
    expect(multiConflictPresentation.secondaryKind).toBe("same_field_conflict");
    expect(multiConflictPresentation.secondaryMessage).toBe(
      "3 same-field conflicts need review.",
    );
    expect(multiConflictPresentation.conflictAnchors).toEqual([
      {
        record_id: "record-a",
        field_key: "timeline.summary",
        base_row_version: 1,
        current_row_version: 2,
      },
      {
        record_id: "record-b",
        field_key: "timeline.details",
        base_row_version: 3,
        current_row_version: 4,
      },
      {
        record_id: "record-c",
        field_key: "timeline.tags",
        base_row_version: 5,
      },
    ]);
    expect(multiConflictPresentation.secondaryMessage).not.toContain("record-");
    expect(multiConflictPresentation.secondaryMessage).not.toContain(
      "timeline.",
    );
  });
});

describe("FE-U-P7-02 conflict anchoring and resolver state unit model", () => {
  it("FE-U-P7-02 keeps same-field conflict queue identity separate from pending replay and visible order", () => {
    const queue = createQueue();
    expectAccepted(
      queue.admit(
        patchUnit({
          clientTxnId: "txn-visible-late",
          recordId: "record-anchor",
          order: 1,
          baseRowVersion: 41,
          fieldKey: "timeline.summary",
          value: "Local anchored draft",
          presentationHint: {
            label: "Visible row 99",
            recordType: "timeline-visible-label",
            sortRank: 99,
            visibleRowIndex: 99,
          },
          visibleEdit: {
            rowKey: "visible-row-99",
            fieldKey: "timeline.summary",
            value: "Local anchored draft",
          },
        }),
      ),
    );
    expectAccepted(
      queue.admit(
        patchUnit({
          clientTxnId: "txn-behind-conflict",
          recordId: "record-behind",
          order: 2,
          baseRowVersion: 3,
          fieldKey: "timeline.details",
          value: "Behind conflict",
          presentationHint: {
            label: "Visible row 1",
            recordType: "timeline-visible-label",
            sortRank: 1,
            visibleRowIndex: 1,
          },
        }),
      ),
    );

    queue.dispatchNext();
    const settlement = queue.settleDispatched({
      ok: false,
      status: 409,
      error: sameFieldConflictError({
        conflictToken: "conflict-token-anchor",
        recordId: "record-anchor",
        fieldKey: "timeline.summary",
        baseRowVersion: 41,
        currentRowVersion: 42,
      }),
    });

    expect(settlement.outcome).toBe("same_field_conflict");
    if (settlement.outcome !== "same_field_conflict") {
      throw new Error("expected same-field conflict settlement");
    }
    expect(settlement.conflict.key).toBe(
      sameFieldConflictQueueKey({
        record_id: "record-anchor",
        field_key: "timeline.summary",
      }),
    );
    expect(settlement.conflict).toMatchObject({
      record_id: "record-anchor",
      field_key: "timeline.summary",
      base_row_version: 41,
      current_row_version: 42,
    });
    expect(settlement.conflict.key).not.toContain("99");
    expect(settlement.conflict.key).not.toContain("Visible row");
    const [conflictAnchor] =
      settlement.snapshot.saveStatePresentation.conflictAnchors;
    if (conflictAnchor === undefined) {
      throw new Error("expected same-field conflict save-state anchor");
    }
    expect(workbookSaveStateConflictAnchorIdentity(conflictAnchor)).toBe(
      "record-anchor\u0000timeline.summary\u000041",
    );
    expect(settlement.snapshot.units.map((unit) => unit.clientTxnId)).toEqual([
      "txn-behind-conflict",
    ]);
    expect(queue.dispatchNext()).toBeNull();

    queue.clearSameFieldConflict(settlement.conflict.key);
    expect(queue.snapshot().sameFieldConflicts).toEqual([]);
    expect(queue.dispatchNext()?.unit.clientTxnId).toBe("txn-behind-conflict");
  });

  it("FE-U-P7-02 derives save-state conflict anchors from record_id field_key and base_row_version", () => {
    const presentation = deriveWorkbookSaveState({
      queuedCount: 5,
      inFlightCount: 1,
      sameFieldConflicts: [
        {
          key: "record-a:timeline.summary",
          conflict_token: "conflict-a",
          record_id: "record-a",
          field_key: "timeline.summary",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 7,
          current_row_version: 8,
        },
        {
          key: "record-a:timeline.summary",
          conflict_token: "conflict-a-refresh",
          record_id: "record-a",
          field_key: "timeline.summary",
          conflict_resolution_class: "text_compare_merge",
          base_row_version: 9,
          current_row_version: 10,
        },
      ],
      localDraftConflicts: [
        {
          record_id: "record-a",
          field_key: "timeline.summary",
          base_row_version: 7,
          current_row_version: 8,
        },
        {
          record_id: "record-b",
          field_key: "timeline.details",
          base_row_version: 2,
          current_row_version: 5,
        },
      ],
    });

    expect(presentation.primaryLabel).toBe("Conflict");
    expect(presentation.secondaryKind).toBe("same_field_conflict");
    expect(presentation.secondaryMessage).toBe(
      "3 same-field conflicts need review.",
    );
    expect(
      presentation.conflictAnchors.map((anchor) =>
        workbookSaveStateConflictAnchorIdentity(anchor),
      ),
    ).toEqual([
      "record-a\u0000timeline.summary\u00007",
      "record-a\u0000timeline.summary\u00009",
      "record-b\u0000timeline.details\u00002",
    ]);
    expect(presentation.secondaryMessage).not.toContain("record-a");
    expect(presentation.secondaryMessage).not.toContain("timeline.summary");
    expect(presentation.secondaryMessage).not.toContain("conflict-a");
    expect(presentation.secondaryMessage).not.toContain("/api/v1");
  });
});
