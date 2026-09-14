import { afterEach, expect, it, vi } from "vitest";
import { timelineCaptureReview } from "../../../testing/timelineCaptureActionTestSupport";
import {
  mentionCreationReceipt,
  mentionReceipt,
  mentionReview,
} from "../../../testing/timelineMentionTestSupport";
import {
  successEnvelope,
  timelineRow,
} from "../../../testing/timelineWorkbookTestSupport";
import { createWorkbookRecordHistoryAdapter } from "../../adapters/createWorkbookRecordHistoryAdapter";
import type { HistoryAttempt } from "../../history/workbookHistoryOperation";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import { initialMentionCreateDraft } from "../actions/timelineMentionCreationModel";
import { createDraftRowForKey, rowFromApi } from "../models/timelineRowModel";
import { createTimelineBulkTagCommandAdapter } from "./createTimelineBulkTagCommandAdapter";
import { createTimelineEvidenceAttachmentAdapter } from "./createTimelineEvidenceAttachmentAdapter";
import { createTimelineMentionEntityCreationAdapter } from "./createTimelineMentionEntityCreationAdapter";
import { createTimelineMentionResolutionAdapter } from "./createTimelineMentionResolutionAdapter";
import { createTimelineRecordActionAdapter } from "./createTimelineRecordActionAdapter";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";
const replacementRecordId = "20000000-0000-4000-8000-000000000002";
const evidenceRecordId = "20000000-0000-4000-8000-000000000003";
const changeSetId = "30000000-0000-4000-8000-000000000001";
const objectBlobId = "40000000-0000-4000-8000-000000000001";

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
  const history = createWorkbookRecordHistoryAdapter({
    apiBase: "/base",
    incidentId,
  });

  await expect(
    history.load(recordId, new AbortController().signal),
  ).resolves.toMatchObject({
    kind: "accepted",
    value: { record_id: recordId, row_version: 4 },
  });
  await expect(
    history.send(historyAttempt("delete"), new AbortController().signal),
  ).resolves.toMatchObject({
    kind: "acknowledged",
    receipt: { kind: "delete", recordId, rowVersion: 5 },
  });
  await expect(
    history.send(historyAttempt("rollback"), new AbortController().signal),
  ).resolves.toEqual({ kind: "uncertain" });
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
        reason: null,
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
        reason: "Correct duplicate observation",
        record_id: recordId,
        replacement_record_id: evidenceRecordId,
        row_version: 6,
      }),
    );
  vi.stubGlobal("fetch", fetchMock);
  const actions = createTimelineRecordActionAdapter({ apiBase: undefined });
  const review = timelineCaptureReview();
  await expect(
    actions.send(
      actions.capture(review, "txn-review"),
      new AbortController().signal,
    ),
  ).resolves.toMatchObject({
    kind: "acknowledged",
    receipt: {
      operation: "mark-reviewed",
      data: {
        capture_state: "reviewed",
        change_set_id: changeSetId,
        incident_id: incidentId,
        reason: null,
        record_id: recordId,
        replacement_record_id: null,
        row_version: 5,
      },
    },
  });
  await expect(
    actions.send(
      actions.capture(
        timelineCaptureReview({
          action: "supersede",
          target: { ...review.target, rowVersion: 5, captureState: "reviewed" },
          replacement: { ...review.target, recordId: replacementRecordId },
          reason: "Correct duplicate observation",
        }),
        "txn-supersede",
      ),
      new AbortController().signal,
    ),
  ).resolves.toEqual({ kind: "uncertain" });
  expect(requestBody(fetchMock, 0)).toEqual({
    base_row_version: 4,
    client_txn_id: "txn-review",
  });
  expect(requestBody(fetchMock, 1)).toEqual({
    base_row_version: 5,
    client_txn_id: "txn-supersede",
    reason: "Correct duplicate observation",
    replacement_record_id: replacementRecordId,
  });
});

it("owns entity mention creation and resolution transport behind semantic outcomes", async () => {
  const review = mentionReview();
  const createReview = {
    subject: review.subject,
    authority: review.authority,
    draft: initialMentionCreateDraft(review.subject),
  };
  const created = mentionCreationReceipt(createReview);
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(successEnvelope(created.data))
    .mockResolvedValueOnce(successEnvelope(mentionReceipt(review)));
  vi.stubGlobal("fetch", fetchMock);
  const creation = createTimelineMentionEntityCreationAdapter({
    apiBase: "/base",
  });
  const resolution = createTimelineMentionResolutionAdapter({
    apiBase: "/base",
  });
  await expect(
    creation.send(
      creation.capture(createReview, "create-key"),
      new AbortController().signal,
    ),
  ).resolves.toMatchObject({
    kind: "accepted",
    receipt: { data: created.data },
  });
  await expect(
    resolution.send(
      resolution.capture(review, "resolve-key"),
      new AbortController().signal,
    ),
  ).resolves.toEqual({ kind: "accepted", receipt: mentionReceipt(review) });
  expect(requestBody(fetchMock, 0)).toMatchObject({
    client_txn_id: "create-key",
    "host.display_name": "Raw.Host",
  });
  expect(requestBody(fetchMock, 1)).toEqual({
    action: "resolve_item",
    base_mention_row_version: 2,
    client_txn_id: "resolve-key",
    resolved_record_id:
      review.intent.action === "resolve_item"
        ? review.intent.resolvedRecordId
        : "",
  });
});

it("admits the exact Timeline bulk-tag plan to retained ownership", () => {
  const admit = vi.fn(() => "batch-tag");
  const port = createTimelineBulkTagCommandAdapter({ admit });
  const admission = { delivery: {} };
  expect(
    port.assignTag(
      { tagName: "triaged", targets: [{ baseRowVersion: 4, recordId }] },
      admission,
    ),
  ).toBe("batch-tag");
  expect(admit).toHaveBeenCalledWith(
    {
      operation: "applyWorkbookBulkMutation",
      recordIds: [recordId],
      request: {
        kind: "multi_row_tag_assignment_v1",
        tag_name: "triaged",
        targets: [{ base_row_version: 4, record_id: recordId }],
        view_schema_id: timelineViewSchemaId,
      },
    },
    admission,
  );
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

  const evidenceAttachment = createTimelineEvidenceAttachmentAdapter({
    apiBase: "/base",
    createClientTxnId,
    incidentId,
  });
  const file = new File(["abc"], "evidence.txt", { type: "text/plain" });
  const evidence = await evidenceAttachment.createEvidence({ file });
  expect(evidence).toEqual({
    kind: "accepted",
    value: { evidenceRecordId },
  });
  const result = await evidenceAttachment.attachEvidence({
    evidenceRecordId,
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

  const malformedAttachment = createTimelineEvidenceAttachmentAdapter({
    apiBase: "/base",
    createClientTxnId,
    incidentId,
  });
  const malformedEvidence = await malformedAttachment.createEvidence({ file });
  expect(malformedEvidence.kind).toBe("accepted");
  const malformedResult = await malformedAttachment.attachEvidence({
    evidenceRecordId,
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
  const createAttachment = createTimelineEvidenceAttachmentAdapter({
    apiBase: "/base",
    createClientTxnId,
    incidentId,
  });
  const createEvidence = await createAttachment.createEvidence({ file });
  expect(createEvidence.kind).toBe("accepted");
  const createResult = await createAttachment.attachEvidence({
    evidenceRecordId,
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

it("fails Evidence materialization closed before any Timeline mutation", async () => {
  vi.spyOn(document, "cookie", "get").mockReturnValue(
    "cartulary_csrf=evidence-timeline-csrf",
  );
  const fetchMock = vi.fn(async () => {
    throw new TypeError("upload unavailable");
  });
  vi.stubGlobal("fetch", fetchMock);
  const port = createTimelineEvidenceAttachmentAdapter({
    apiBase: "/base",
    createClientTxnId: () => "txn-file-failure",
    incidentId,
  });

  await expect(
    port.createEvidence({
      file: new File(["abc"], "failed.txt", { type: "text/plain" }),
    }),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { kind: "terminal" },
  });
  expect(fetchMock).toHaveBeenCalledOnce();
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

function historyAttempt(operation: "delete" | "rollback"): HistoryAttempt {
  const rowVersion = operation === "delete" ? 4 : 5;
  const id = `txn-${operation}`;
  const target = { kind: "change_set" as const, change_set_id: changeSetId };
  return {
    id,
    actorId: "actor",
    incidentId,
    operation,
    subject: {
      kind: "live",
      recordId,
      rowVersion,
      label: "A row",
      surfaceLabel: "Timeline",
      viewSchemaId: timelineViewSchemaId,
    },
    pending:
      operation === "delete"
        ? { kind: "destructive", operation, recordId, rowVersion }
        : {
            kind: "rollback",
            action: "change_set",
            historyItemRef: "hitem-timeline-1",
            recordId,
            rowVersion,
            target,
          },
    body: JSON.stringify({
      base_row_version: rowVersion,
      client_txn_id: id,
      reason:
        operation === "delete"
          ? "Deleted from workbook history"
          : "Rollback from workbook history",
      ...(operation === "rollback" ? { target } : {}),
    }),
  };
}
