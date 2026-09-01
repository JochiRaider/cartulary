import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import { hostsViewSchemaId } from "../models/workbookSurfaceRegistry";
import { createWorkbookMutationCommandPorts } from "./createWorkbookMutationCommandPorts";

function successResponse(): Response {
  return new Response(
    JSON.stringify({
      data: {
        change_set_id: "change-set-1",
        row: {
          cells: {},
          record_id: "00000000-0000-4000-8000-000000000010",
          row_version: 1,
        },
        view_schema_id: "cartulary.view.timeline.v2",
      },
      meta: { request_id: "request-success" },
    }),
    {
      status: 200,
      headers: { "content-type": "application/json" },
    },
  );
}

function requestBodies(fetchMock: ReturnType<typeof vi.fn>) {
  return fetchMock.mock.calls.map((call) => {
    const init = call[1] as RequestInit | undefined;
    return JSON.parse(String(init?.body)) as Record<string, unknown>;
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("semantic mutation command ports", () => {
  it("validates generated Workbook operation responses and contains public failures", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              incident_id: "00000000-0000-4000-8000-000000000001",
              rows: [],
              view_schema_id: "cartulary.view.timeline.v2",
            },
            meta: {
              query: { filters: [], sort: [] },
              request_id: "request-query",
            },
          }),
          { status: 200, headers: { "content-type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ data: { rows: [] } }), {
          status: 200,
          headers: { "content-type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            error: {
              code: "authorization_denied",
              conflict: undefined,
              details: { reason_code: "incident_access_lost" },
              message: "Access denied.",
              request_id: "request-denied",
              retryable: false,
              status: 403,
            },
          }),
          { status: 403, headers: { "content-type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ error: { message: "raw detail" } }), {
          status: 500,
          headers: { "content-type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            error: {
              code: "evidence_access_unavailable",
              details: { reason_code: "unsupported_preview" },
              message: "object://must-not-cross-the-adapter",
              request_id: "request-evidence-denied",
              retryable: false,
              status: 409,
            },
          }),
          { status: 409, headers: { "content-type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            error: {
              code: "row_version_conflict",
              details: {},
              message: "Server conflict prose.",
              request_id: "request-row-version-conflict",
              retryable: false,
              status: 409,
            },
          }),
          { status: 409, headers: { "content-type": "application/json" } },
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const operations = createWorkbookOperationExecutor({ apiBase: undefined });
    const queryInput = {
      operationID: "queryWorkbookView" as const,
      pathParameters: {
        incident_id: "00000000-0000-4000-8000-000000000001",
        view_schema_id: "cartulary.view.timeline.v2",
      },
      request: { limit: 50 },
    };

    await expect(operations.execute(queryInput)).resolves.toMatchObject({
      kind: "accepted",
      value: { data: { rows: [] } },
    });
    await expect(operations.execute(queryInput)).resolves.toMatchObject({
      kind: "rejected",
      failure: {
        kind: "invalid_contract",
        message: "The server returned an invalid public contract response.",
        presentation: { family: "initial_load_failure" },
      },
    });
    await expect(operations.execute(queryInput)).resolves.toMatchObject({
      kind: "rejected",
      failure: {
        kind: "authorization_lost",
        message: "Access denied.",
        publicCode: "authorization_denied",
        presentation: { family: "permission_or_incident_access_loss" },
      },
    });
    await expect(operations.execute(queryInput)).resolves.toMatchObject({
      kind: "rejected",
      failure: {
        kind: "invalid_contract",
        message: "The server returned an invalid public error response.",
        presentation: { family: "initial_load_failure" },
      },
    });
    await expect(
      operations.execute({
        operationID: "issueEvidencePreviewHandle",
        pathParameters: {
          record_id: "00000000-0000-4000-8000-000000000010",
        },
        request: {},
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: {
        kind: "terminal",
        message: "evidence_access_unavailable: unsupported_preview",
        publicCode: "evidence_access_unavailable",
        presentation: { family: "evidence_preview_blocked" },
      },
    });
    await expect(operations.execute(queryInput)).resolves.toMatchObject({
      kind: "rejected",
      failure: {
        kind: "stale_target",
        publicCode: "row_version_conflict",
      },
    });
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/incidents/00000000-0000-4000-8000-000000000001/views/cartulary.view.timeline.v2/query",
      expect.objectContaining({ method: "POST", body: '{"limit":50}' }),
    );
  });

  it("own exact Timeline, generic, entity, assessment, and coordination request identities and payloads", async () => {
    const fetchMock = vi.fn(async () => successResponse());
    vi.stubGlobal("fetch", fetchMock);
    const commands = createWorkbookMutationCommandPorts({
      apiBase: undefined,
      incidentId: "incident-1",
      transactionIds: {
        create: (prefix) => `${prefix}-id`,
      },
    });

    await commands.timeline.bulk.assignTag({
      tagName: "triaged",
      targets: [{ recordId: "timeline-1", baseRowVersion: 3 }],
    });
    await commands.generic.patchRecord({
      baseRowVersion: 4,
      changes: [{ field_key: "task.title", value: "Updated" }],
      purpose: "generic-patch",
      recordId: "task-1",
      viewSchemaId: "cartulary.view.tasks.v1",
    });
    await commands.entity.merge({
      loserBaseRowVersion: 6,
      loserRecordId: "host-2",
      reason: "Duplicate",
      survivorBaseRowVersion: 5,
      survivorRecordId: "host-1",
    });
    await commands.assessment.create({
      draft: {
        assessedAt: "2026-07-31T05:00:00Z",
        assessmentState: "confirmed",
        confidenceBand: "high",
        rationale: "Correlated evidence",
        subjectRecordId: "host-1",
        subjectType: "host",
        supportRecordIds: ["evidence-1"],
      },
    });
    await commands.coordination.updateTaskLifecycle({
      baseRowVersion: 7,
      blockedReason: "Waiting for owner",
      recordId: "task-2",
      status: "blocked",
    });

    expect(requestBodies(fetchMock)).toEqual([
      {
        view_schema_id: "cartulary.view.timeline.v2",
        client_txn_id: "timeline-client-id",
        kind: "multi_row_tag_assignment_v1",
        tag_name: "triaged",
        targets: [{ record_id: "timeline-1", base_row_version: 3 }],
      },
      {
        view_schema_id: "cartulary.view.tasks.v1",
        base_row_version: 4,
        client_txn_id: "generic-patch-cartulary.view.tasks.v1-id",
        changes: [{ field_key: "task.title", value: "Updated" }],
      },
      {
        loser_record_id: "host-2",
        survivor_base_row_version: 5,
        loser_base_row_version: 6,
        client_txn_id: "merge-id",
        reason: "Duplicate",
      },
      {
        client_txn_id: "assessment-id",
        "assessment.subject_ref": "host-1",
        "assessment.subject_type": "host",
        "assessment.assessment_state": "confirmed",
        "assessment.confidence_score": 85,
        "assessment.rationale": "Correlated evidence",
        "assessment.assessed_at": "2026-07-31T05:00:00Z",
        "assessment.support_refs": {
          kind: "collection_actions_v1",
          actions: [{ op: "add_record_ref", linked_record_id: "evidence-1" }],
        },
      },
      {
        view_schema_id: "cartulary.view.task_requests.v1",
        base_row_version: 7,
        client_txn_id: "task-lifecycle-id",
        changes: [
          { field_key: "task.status", value: "blocked" },
          {
            field_key: "task.blocked_reason",
            value: "Waiting for owner",
          },
        ],
      },
    ]);
  });

  it("normalizes task lifecycle and decision supersede into owner-specific outcomes", async () => {
    const taskRecordId = "00000000-0000-4000-8000-000000000410";
    const decisionTargetId = "00000000-0000-4000-8000-000000000420";
    const decisionReplacementId = "00000000-0000-4000-8000-000000000421";
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              change_set_id: "00000000-0000-4000-8000-000000000510",
              row: {
                cells: {},
                record_id: taskRecordId,
                row_version: 8,
              },
              view_schema_id: "cartulary.view.task_requests.v1",
            },
            meta: { request_id: "request-task" },
          }),
          { status: 200, headers: { "content-type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              change_set_id: "00000000-0000-4000-8000-000000000511",
              reason: "Replaced after review",
              superseding_record_id: decisionReplacementId,
              superseding_row_version: 7,
              target_record_id: decisionTargetId,
              target_row_version: 5,
              target_status: "superseded",
              view_schema_id: "cartulary.view.decisions.v1",
            },
            meta: { request_id: "request-decision" },
          }),
          { status: 200, headers: { "content-type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              capture_state: "superseded",
              change_set_id: "00000000-0000-4000-8000-000000000512",
              incident_id: "00000000-0000-4000-8000-000000000001",
              reason: "Wrong owner response",
              record_id: decisionTargetId,
              replacement_record_id: decisionReplacementId,
              row_version: 6,
            },
            meta: { request_id: "request-wrong-owner" },
          }),
          { status: 200, headers: { "content-type": "application/json" } },
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const commands = createWorkbookMutationCommandPorts({
      apiBase: undefined,
      incidentId: "00000000-0000-4000-8000-000000000001",
      transactionIds: { create: (prefix) => `${prefix}-id` },
    });

    await expect(
      commands.coordination.updateTaskLifecycle({
        baseRowVersion: 7,
        recordId: taskRecordId,
        status: "done",
      }),
    ).resolves.toEqual({
      kind: "accepted",
      value: {
        changeSetId: "00000000-0000-4000-8000-000000000510",
        row: { cells: {}, record_id: taskRecordId, row_version: 8 },
        status: "done",
        viewSchemaId: "cartulary.view.task_requests.v1",
      },
    });
    await expect(
      commands.coordination.supersedeDecision({
        baseRowVersion: 4,
        reason: "Replaced after review",
        replacementRecordId: decisionReplacementId,
        targetRecordId: decisionTargetId,
      }),
    ).resolves.toEqual({
      kind: "accepted",
      value: {
        changeSetId: "00000000-0000-4000-8000-000000000511",
        replacementRecordId: decisionReplacementId,
        replacementRowVersion: 7,
        targetRecordId: decisionTargetId,
        targetRowVersion: 5,
        targetStatus: "superseded",
        viewSchemaId: "cartulary.view.decisions.v1",
      },
    });
    await expect(
      commands.coordination.supersedeDecision({
        baseRowVersion: 5,
        reason: "Reject wrong response branch",
        replacementRecordId: decisionReplacementId,
        targetRecordId: decisionTargetId,
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
  });

  it("keeps secure transaction identity failure local without transport", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const commands = createWorkbookMutationCommandPorts({
      apiBase: undefined,
      incidentId: "incident-1",
      transactionIds: {
        create: () => {
          throw new Error("randomness unavailable");
        },
      },
    });

    await expect(
      commands.generic.patchRecord({
        baseRowVersion: 1,
        changes: [{ field_key: "task.title", value: "Local draft" }],
        purpose: "generic-patch",
        recordId: "task-1",
        viewSchemaId: "cartulary.view.tasks.v1",
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: {
        kind: "terminal",
        message: "A secure transaction ID could not be created.",
      },
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("executes record lifecycle actions through the declared existing route", async () => {
    const recordId = "20000000-0000-4000-8000-000000000001";
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            change_set_id: "30000000-0000-4000-8000-000000000001",
            deleted: true,
            deleted_at: "2026-08-30T20:00:00Z",
            deleted_by_user_id: "40000000-0000-4000-8000-000000000001",
            incident_id: "10000000-0000-4000-8000-000000000001",
            record_id: recordId,
            row_version: 6,
          },
          meta: { request_id: "request-record-delete" },
        }),
        { status: 200, headers: { "content-type": "application/json" } },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    const commands = createWorkbookMutationCommandPorts({
      apiBase: undefined,
      incidentId: "10000000-0000-4000-8000-000000000001",
      transactionIds: { create: (prefix) => `${prefix}-id` },
    });

    await expect(
      commands.records.execute({
        action: "delete",
        baseRowVersion: 5,
        reason: "Deleted from the inspector",
        recordId,
      }),
    ).resolves.toEqual({
      kind: "accepted",
      value: { recordId, rowVersion: 6 },
    });
    expect(fetchMock).toHaveBeenCalledWith(
      `/api/v1/records/${recordId}`,
      expect.objectContaining({
        body: JSON.stringify({
          base_row_version: 5,
          client_txn_id: "record-delete-id",
          reason: "Deleted from the inspector",
        }),
        method: "DELETE",
      }),
    );
  });

  it("normalizes generic same-field conflicts before they reach controllers", async () => {
    const fetchMock = vi.fn(
      async () =>
        new Response(
          JSON.stringify({
            error: {
              code: "same_field_conflict",
              conflict: {
                base_row_version: 3,
                base_value: "before",
                client_value: "local",
                conflict_resolution_class: "text_compare_merge",
                conflict_token: "conflict-1",
                current_row_version: 4,
                field_key: "task.title",
                record_id: "task-1",
                server_value: "server",
              },
              details: {},
              message: "Conflict.",
              request_id: "request-conflict",
              retryable: false,
              status: 409,
            },
          }),
          { status: 409, headers: { "content-type": "application/json" } },
        ),
    );
    vi.stubGlobal("fetch", fetchMock);
    const commands = createWorkbookMutationCommandPorts({
      apiBase: undefined,
      incidentId: "incident-1",
      transactionIds: { create: (prefix) => `${prefix}-id` },
    });

    await expect(
      commands.generic.patchRecord({
        baseRowVersion: 3,
        changes: [{ field_key: "task.title", value: "local" }],
        purpose: "generic-patch",
        recordId: "task-1",
        viewSchemaId: "cartulary.view.task_requests.v1",
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: {
        kind: "same_field_conflict",
        conflict: {
          conflict_token: "conflict-1",
          field_key: "task.title",
          record_id: "task-1",
        },
      },
    });
  });

  it("normalizes entity create, patch, and paste results and rejects malformed success contracts", async () => {
    const recordId = "00000000-0000-4000-8000-000000000210";
    const viewMutation = {
      data: {
        change_set_id: "00000000-0000-4000-8000-000000000310",
        row: { cells: {}, record_id: recordId, row_version: 2 },
        view_schema_id: hostsViewSchemaId,
      },
      meta: { request_id: "request-entity" },
    };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify(viewMutation), {
          status: 200,
          headers: { "content-type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(viewMutation), {
          status: 200,
          headers: { "content-type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              change_set_id: "00000000-0000-4000-8000-000000000311",
              conflicts: [],
              rows: [viewMutation.data.row],
              view_schema_id: hostsViewSchemaId,
            },
            meta: { request_id: "request-paste" },
          }),
          { status: 200, headers: { "content-type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ data: { rows: [] } }), {
          status: 200,
          headers: { "content-type": "application/json" },
        }),
      );
    vi.stubGlobal("fetch", fetchMock);
    const commands = createWorkbookMutationCommandPorts({
      apiBase: undefined,
      incidentId: "00000000-0000-4000-8000-000000000001",
      transactionIds: { create: (prefix) => `${prefix}-id` },
    });
    const contract = requireViewContract(hostsViewSchemaId);

    await expect(
      commands.entity.createRecord({
        contract,
        draft: { "host.hostname": "edge-01.example.test" },
      }),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: { row: { record_id: recordId }, viewSchemaId: hostsViewSchemaId },
    });
    await expect(
      commands.entity.patchRecord({
        baseRowVersion: 1,
        changes: [{ field_key: "host.display_name", value: "Edge 01" }],
        purpose: "entity-patch",
        recordId,
        viewSchemaId: hostsViewSchemaId,
      }),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: { row: { record_id: recordId }, viewSchemaId: hostsViewSchemaId },
    });
    const pasteInput = {
      clipboardText: "Pasted host",
      columns: ["host.display_name"],
      format: "tsv",
      startFieldKey: "host.display_name",
      targetCount: 1,
      viewSchemaId: hostsViewSchemaId,
    } as const;
    await expect(
      commands.entity.pasteCreate(pasteInput),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: {
        rows: [{ record_id: recordId }],
        viewSchemaId: hostsViewSchemaId,
      },
    });
    await expect(
      commands.entity.pasteCreate(pasteInput),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: {
        kind: "invalid_contract",
        message: "The server returned an invalid public contract response.",
        presentation: { family: "unknown_future_error" },
      },
    });
  });
});
