import { afterEach, describe, expect, it, vi } from "vitest";
import { createWorkbookMutationCommandPorts } from "./createWorkbookMutationCommandPorts";

function successResponse(): Response {
  return new Response(JSON.stringify({ data: {} }), {
    status: 200,
    headers: { "content-type": "application/json" },
  });
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

    await commands.timeline.assignTag({
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
      ok: false,
      status: 0,
      payload: {
        error: { message: "A secure transaction ID could not be created." },
      },
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
