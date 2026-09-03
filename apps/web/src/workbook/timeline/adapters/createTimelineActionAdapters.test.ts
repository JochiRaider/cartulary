import { afterEach, expect, it, vi } from "vitest";
import {
  successEnvelope,
  timelineRow,
} from "../../../testing/timelineWorkbookTestSupport";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import { createDraftRowForKey, rowFromApi } from "../models/timelineRowModel";
import { createTimelineEvidenceAttachmentAdapter } from "./createTimelineEvidenceAttachmentAdapter";
import { createTimelineHistoryAdapter } from "./createTimelineHistoryAdapter";
import { createTimelineMentionAdapter } from "./createTimelineMentionAdapter";
import { createTimelineRecordActionAdapter } from "./createTimelineRecordActionAdapter";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";
const replacementRecordId = "20000000-0000-4000-8000-000000000002";
const evidenceRecordId = "20000000-0000-4000-8000-000000000003";
const changeSetId = "30000000-0000-4000-8000-000000000001";
const objectBlobId = "40000000-0000-4000-8000-000000000001";
const mentionId = "60000000-0000-4000-8000-000000000001";

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

it("derives history routes and rejects target-inconsistent history responses", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(
      jsonResponse({
        data: {
          deleted: false,
          incident_id: incidentId,
          items: [
            {
              actor_user_id: "50000000-0000-4000-8000-000000000001",
              available_rollback_actions: ["history_entry", "change_set"],
              change_set_id: changeSetId,
              committed_at: "2026-07-31T12:00:00Z",
              diff_summary: { summary: "Updated Timeline row", units: [] },
              history_entry_ref: "href-timeline-1",
              history_item_ref: "hitem-timeline-1",
              operation: "field_update",
              reversible: true,
            },
          ],
          record_id: recordId,
          row_version: 4,
        },
        meta: {
          paging: { has_more: false, limit: 50, next_cursor: null },
          request_id: "req-history",
        },
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        change_set_id: changeSetId,
        deleted: true,
        deleted_at: "2026-07-31T12:01:00Z",
        deleted_by_user_id: "50000000-0000-4000-8000-000000000001",
        incident_id: incidentId,
        record_id: recordId,
        row_version: 5,
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        affected_record_ids: [replacementRecordId],
        incident_id: incidentId,
        record_id: replacementRecordId,
        rollback_change_set_id: changeSetId,
        row_version: 6,
        target: { change_set_id: changeSetId, kind: "change_set" },
        target_change_set_id: changeSetId,
      }),
    );
  vi.stubGlobal("fetch", fetchMock);
  const history = createTimelineHistoryAdapter({ apiBase: "/base" });

  await expect(history.load({ recordId })).resolves.toMatchObject({
    kind: "accepted",
    value: { record_id: recordId, row_version: 4 },
  });
  await expect(
    history.deleteOrRestore({
      baseRowVersion: 4,
      clientTxnId: "txn-delete",
      operation: "delete",
      recordId,
    }),
  ).resolves.toEqual({
    kind: "accepted",
    value: { recordId, rowVersion: 5 },
  });
  await expect(
    history.rollback({
      baseRowVersion: 5,
      clientTxnId: "txn-rollback",
      recordId,
      target: { change_set_id: changeSetId, kind: "change_set" },
    }),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { kind: "invalid_contract" },
  });
  expect(fetchMock.mock.calls.map(([input]) => String(input))).toEqual([
    `/base/api/v1/records/${recordId}/history`,
    `/base/api/v1/records/${recordId}`,
    `/base/api/v1/records/${recordId}/rollback`,
  ]);
  expect(requestBody(fetchMock, 1)).toEqual({
    base_row_version: 4,
    client_txn_id: "txn-delete",
    reason: "Deleted from workbook history",
  });
  expect(requestBody(fetchMock, 2)).toEqual({
    base_row_version: 5,
    client_txn_id: "txn-rollback",
    reason: "Rollback from workbook history",
    target: { change_set_id: changeSetId, kind: "change_set" },
  });
});

it("normalizes Timeline review outcomes and fails closed on supersede replacement drift", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(
      successEnvelope({
        capture_state: "reviewed",
        change_set_id: changeSetId,
        incident_id: incidentId,
        reason: "Reviewed from workbook",
        record_id: recordId,
        replacement_record_id: null,
        row_version: 5,
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        capture_state: "superseded",
        change_set_id: changeSetId,
        incident_id: incidentId,
        reason: "Superseded from workbook",
        record_id: recordId,
        replacement_record_id: evidenceRecordId,
        row_version: 6,
      }),
    );
  vi.stubGlobal("fetch", fetchMock);
  const actions = createTimelineRecordActionAdapter({ apiBase: undefined });

  await expect(
    actions.execute({
      action: "mark-reviewed",
      baseRowVersion: 4,
      clientTxnId: "txn-review",
      recordId,
      replacementRecordId: null,
    }),
  ).resolves.toEqual({
    kind: "accepted",
    value: {
      captureState: "reviewed",
      changeSetId,
      incidentId,
      reason: "Reviewed from workbook",
      recordId,
      replacementRecordId: null,
      rowVersion: 5,
    },
  });
  await expect(
    actions.execute({
      action: "supersede",
      baseRowVersion: 5,
      clientTxnId: "txn-supersede",
      recordId,
      replacementRecordId,
    }),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { kind: "invalid_contract" },
  });
  expect(requestBody(fetchMock, 1)).toEqual({
    base_row_version: 5,
    client_txn_id: "txn-supersede",
    reason: "Superseded from workbook",
    replacement_record_id: replacementRecordId,
  });
});

it("owns entity mention creation and resolution transport behind semantic outcomes", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(
      successEnvelope({
        change_set_id: changeSetId,
        row: { cells: {}, record_id: replacementRecordId, row_version: 1 },
        view_schema_id: hostsViewSchemaId,
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        active_link: null,
        change_set_id: changeSetId,
        entity_mention: {
          entity_mention_id: mentionId,
          entity_type: "host",
          raw_text: "server.example",
          resolution_method: "manual",
          resolution_status: "resolved",
          resolved_record_id: replacementRecordId,
          row_version: 3,
          source_field_key: "timeline.activity_synopsis_text",
        },
        incident_id: incidentId,
        source_record: { record_id: recordId, row_version: 7 },
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        active_link: null,
        change_set_id: changeSetId,
        entity_mention: {
          entity_mention_id: mentionId,
          entity_type: { malformed: true },
          raw_text: "server.example",
          resolution_method: "manual",
          resolution_status: "resolved",
          resolved_record_id: replacementRecordId,
          row_version: 4,
          source_field_key: "timeline.activity_synopsis_text",
        },
        incident_id: incidentId,
        source_record: { record_id: recordId, row_version: 8 },
      }),
    );
  vi.stubGlobal("fetch", fetchMock);
  const mentions = createTimelineMentionAdapter({
    apiBase: "/base",
    incidentId,
  });

  await expect(
    mentions.createEntity({
      clientTxnId: "txn-create-host",
      entityType: "host",
      rawText: "server.example",
    }),
  ).resolves.toEqual({
    kind: "accepted",
    value: { recordId: replacementRecordId },
  });
  await expect(
    mentions.resolve({
      action: "resolve_item",
      baseMentionRowVersion: 2,
      clientTxnId: "txn-resolve",
      expectedSourceRecordId: recordId,
      mentionId,
      resolvedRecordId: replacementRecordId,
    }),
  ).resolves.toMatchObject({
    kind: "accepted",
    value: {
      entityMention: { entityType: "host", rowVersion: 3 },
      sourceRecord: { recordId, rowVersion: 7 },
    },
  });
  await expect(
    mentions.resolve({
      action: "resolve_item",
      baseMentionRowVersion: 3,
      clientTxnId: "txn-resolve-malformed",
      expectedSourceRecordId: recordId,
      mentionId,
      resolvedRecordId: replacementRecordId,
    }),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { kind: "invalid_contract" },
  });
  expect(requestBody(fetchMock, 0)).toEqual({
    client_txn_id: "txn-create-host",
    "host.display_name": "server.example",
    "host.fqdn": "server.example",
  });
  expect(requestBody(fetchMock, 1)).toEqual({
    action: "resolve_item",
    base_mention_row_version: 2,
    client_txn_id: "txn-resolve",
    resolved_record_id: replacementRecordId,
  });
});

it("creates a blob-backed Evidence row atomically and reuses the row transaction ID after response uncertainty", async () => {
  vi.spyOn(document, "cookie", "get").mockReturnValue(
    "cartulary_csrf=evidence-timeline-csrf",
  );
  let evidenceRowAttempts = 0;
  let timelinePatchAttempts = 0;
  const fetchMock = vi.fn((input: RequestInfo | URL, _init?: RequestInit) => {
    const url = String(input);
    if (url === "/base/api/v1/object-blobs") {
      return Promise.resolve(
        successEnvelope({
          accepted_contract: {
            byte_size: 3,
            content_type_hint: "text/plain",
            filename_hint: "evidence.txt",
            incident_id: incidentId,
            sha256_hex: null,
          },
          incident_id: incidentId,
          object_blob_id: objectBlobId,
          pending_expires_at: "2026-07-31T12:10:00Z",
          target_expires_at: "2026-07-31T12:05:00Z",
          upload_state: "pending",
          upload_target: {
            expires_at: "2026-07-31T12:05:00Z",
            headers: { "Content-Type": "text/plain" },
            href: "/api/v1/object-uploads/upload-token",
            method: "PUT",
          },
        }),
      );
    }
    if (url === "/base/api/v1/object-uploads/upload-token") {
      return Promise.resolve(new Response("", { status: 200 }));
    }
    if (
      url ===
      `/base/api/v1/incidents/${incidentId}/views/${evidenceViewSchemaId}/rows`
    ) {
      evidenceRowAttempts += 1;
      if (evidenceRowAttempts === 1) {
        return Promise.reject(new TypeError("response lost"));
      }
      return Promise.resolve(
        successEnvelope({
          change_set_id: changeSetId,
          row: { cells: {}, record_id: evidenceRecordId, row_version: 1 },
          view_schema_id: evidenceViewSchemaId,
        }),
      );
    }
    if (url === `/base/api/v1/records/${recordId}`) {
      timelinePatchAttempts += 1;
      return Promise.resolve(
        successEnvelope({
          change_set_id: changeSetId,
          row: timelineRow({
            captureState: "enriched",
            evidenceCount: 1,
            recordId:
              timelinePatchAttempts === 1 ? recordId : replacementRecordId,
            rowVersion: timelinePatchAttempts === 1 ? 5 : 6,
          }),
          view_schema_id: timelineViewSchemaId,
        }),
      );
    }
    if (
      url ===
      `/base/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`
    ) {
      return Promise.resolve(
        successEnvelope({
          change_set_id: changeSetId,
          row: timelineRow({
            captureState: "enriched",
            evidenceCount: 1,
            recordId: replacementRecordId,
            rowVersion: 1,
          }),
          view_schema_id: timelineViewSchemaId,
        }),
      );
    }
    return Promise.resolve(
      jsonResponse({ error: { code: "unexpected" } }, 500),
    );
  });
  vi.stubGlobal("fetch", fetchMock);
  const createClientTxnId = vi
    .fn()
    .mockReturnValueOnce("txn-blob")
    .mockReturnValueOnce("txn-evidence-row")
    .mockReturnValueOnce("txn-timeline-link")
    .mockReturnValueOnce("txn-blob-malformed")
    .mockReturnValueOnce("txn-evidence-row-malformed")
    .mockReturnValueOnce("txn-timeline-link-malformed")
    .mockReturnValueOnce("txn-blob-create")
    .mockReturnValueOnce("txn-evidence-row-create")
    .mockReturnValueOnce("txn-timeline-create");
  const trackTimelineTxn = vi.fn();
  const target = rowFromApi({
    ...timelineRow({ captureState: "rough", recordId, rowVersion: 4 }),
    view_schema_id: timelineViewSchemaId,
  });

  const result = await createTimelineEvidenceAttachmentAdapter({
    apiBase: "/base",
    createClientTxnId,
    incidentId,
  }).attach({
    file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
    onTimelineClientTxnId: trackTimelineTxn,
    target,
  });

  expect(result).toMatchObject({
    clientTxnId: "txn-timeline-link",
    outcome: {
      kind: "accepted",
      value: {
        evidenceRecordId,
        row: { record_id: recordId, row_version: 5 },
        viewSchemaId: timelineViewSchemaId,
      },
    },
  });
  expect(createClientTxnId).toHaveBeenCalledTimes(3);
  expect(trackTimelineTxn).toHaveBeenCalledWith("txn-timeline-link");
  const rowBodies = fetchMock.mock.calls
    .filter(([input]) =>
      String(input).includes(`/views/${evidenceViewSchemaId}/rows`),
    )
    .map(([, init]) => String((init as RequestInit | undefined)?.body));
  expect(rowBodies).toHaveLength(2);
  expect(rowBodies[1]).toBe(rowBodies[0]);
  expect(JSON.parse(rowBodies[0] ?? "null")).toMatchObject({
    client_txn_id: "txn-evidence-row",
    "evidence.initial_object_blob_id": objectBlobId,
  });

  const malformedResult = await createTimelineEvidenceAttachmentAdapter({
    apiBase: "/base",
    createClientTxnId,
    incidentId,
  }).attach({
    file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
    onTimelineClientTxnId: trackTimelineTxn,
    target,
  });
  expect(malformedResult).toMatchObject({
    clientTxnId: "txn-timeline-link-malformed",
    outcome: {
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    },
  });

  const draftTarget = createDraftRowForKey("draft-evidence");
  if (draftTarget === null) throw new Error("expected draft Timeline target");
  const createResult = await createTimelineEvidenceAttachmentAdapter({
    apiBase: "/base",
    createClientTxnId,
    incidentId,
  }).attach({
    file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
    onTimelineClientTxnId: trackTimelineTxn,
    target: draftTarget,
  });
  expect(createResult).toMatchObject({
    clientTxnId: "txn-timeline-create",
    outcome: {
      kind: "accepted",
      value: {
        evidenceRecordId,
        row: { record_id: replacementRecordId, row_version: 1 },
      },
    },
  });
  const timelineCreateCall = fetchMock.mock.calls.find(([input]) =>
    String(input).endsWith(`/views/${timelineViewSchemaId}/rows`),
  );
  expect(JSON.parse(String(timelineCreateCall?.[1]?.body ?? "null"))).toEqual({
    client_txn_id: "txn-timeline-create",
    "timeline.attached_evidence_ids": {
      actions: [{ linked_record_id: evidenceRecordId, op: "add_record_ref" }],
      kind: "collection_actions_v1",
    },
  });
  expect(createClientTxnId).toHaveBeenCalledTimes(9);
});

function jsonResponse(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), {
    headers: { "content-type": "application/json" },
    status,
  });
}

function requestBody(fetchMock: ReturnType<typeof vi.fn>, index: number) {
  return JSON.parse(
    String((fetchMock.mock.calls[index]?.[1] as RequestInit).body),
  );
}
