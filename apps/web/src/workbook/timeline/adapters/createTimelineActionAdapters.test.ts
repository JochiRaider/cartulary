import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, expect, it, vi } from "vitest";
import { timelineCaptureReview } from "../../../testing/timelineCaptureActionTestSupport";
import {
  mentionCreationReceipt,
  mentionReceipt,
  mentionReview,
} from "../../../testing/timelineMentionTestSupport";
import {
  fullWorkbookViewRow,
  successEnvelope,
  timelineRow,
} from "../../../testing/timelineWorkbookTestSupport";
import { createEvidenceFileTransport } from "../../adapters/createEvidenceFileTransport";
import { createTimelineFileLinkTransport } from "../../adapters/createTimelineFileLinkTransport";
import { createWorkbookRecordHistoryAdapter } from "../../adapters/createWorkbookRecordHistoryAdapter";
import type { HistoryAttempt } from "../../history/workbookHistoryOperation";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import { initialMentionCreateDraft } from "../actions/timelineMentionCreationModel";
import { createTimelineBulkTagCommandAdapter } from "./createTimelineBulkTagCommandAdapter";
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
  const port = createTimelineBulkTagCommandAdapter({
    admit,
    subscribe: () => () => {},
    getSnapshot: () => ({ authority: null, entries: [], admissionError: null }),
  });
  const admission = { delivery: {} };
  expect(
    port.assignTag(
      { tagName: "triaged", targets: [{ baseRowVersion: 4, recordId }] },
      admission,
    ),
  ).toEqual({ kind: "admitted", operationId: "batch-tag" });
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
  const authority = {
    actorId: "actor",
    sessionIdentity: "session",
    incidentId,
    role: "editor" as const,
    closed: false,
  };
  const transport = createEvidenceFileTransport("/base");
  const attempt = transport.capture({
    stage: "create",
    authority,
    clientTxnId: "txn-evidence-row",
    objectBlobId,
    fields: {
      "evidence.title": "evidence.txt",
      "evidence.collector_party_text": "Workbook upload",
    },
  });
  const row = fullWorkbookViewRow(
    requireViewContract(evidenceViewSchemaId),
    evidenceRecordId,
    1,
    {
      "evidence.title": "evidence.txt",
      "evidence.lifecycle_state": "requested",
      "evidence.storage_ref": `object://${objectBlobId}`,
    },
  );
  const fetchMock = vi
    .fn()
    .mockRejectedValueOnce(new TypeError("acknowledgement lost"))
    .mockImplementationOnce(async () =>
      successEnvelope({
        change_set_id: changeSetId,
        row,
        view_schema_id: evidenceViewSchemaId,
      }),
    );
  vi.stubGlobal("fetch", fetchMock);
  expect(
    await transport.finalize(attempt, new AbortController().signal),
  ).toEqual({ kind: "uncertain" });
  const accepted = await transport.finalize(
    attempt,
    new AbortController().signal,
  );
  expect(accepted).toMatchObject({
    kind: "accepted",
    receipt: { data: { row, change_set_id: changeSetId } },
  });
  expect(fetchMock.mock.calls[0]?.[1].body).toBe(
    fetchMock.mock.calls[1]?.[1].body,
  );
  expect(JSON.parse(attempt.body)).toEqual({
    client_txn_id: "txn-evidence-row",
    "evidence.initial_object_blob_id": objectBlobId,
    "evidence.title": "evidence.txt",
    "evidence.collector_party_text": "Workbook upload",
  });
  const links = createTimelineFileLinkTransport("/base");
  const source = timelineRow({
    captureState: "rough",
    recordId,
    rowVersion: 4,
  });
  const link = links.capture(
    authority,
    source,
    evidenceRecordId,
    "txn-timeline-link",
  );
  fetchMock.mockImplementationOnce(async () =>
    successEnvelope({
      change_set_id: changeSetId,
      row: timelineRow({
        captureState: "enriched",
        recordId,
        rowVersion: 5,
        evidenceCount: 1,
      }),
      view_schema_id: timelineViewSchemaId,
    }),
  );
  expect(await links.send(link, new AbortController().signal)).toMatchObject({
    kind: "accepted",
    receipt: { data: { row: { record_id: recordId, row_version: 5 } } },
  });
  fetchMock.mockImplementationOnce(async () =>
    successEnvelope({
      change_set_id: changeSetId,
      row: timelineRow({
        captureState: "enriched",
        recordId: replacementRecordId,
        rowVersion: 6,
      }),
      view_schema_id: timelineViewSchemaId,
    }),
  );
  expect(await links.send(link, new AbortController().signal)).toEqual({
    kind: "uncertain",
  });
});

it("fails Evidence materialization closed before any Timeline mutation", async () => {
  vi.spyOn(document, "cookie", "get").mockReturnValue(
    "cartulary_csrf=evidence-timeline-csrf",
  );
  const fetchMock = vi.fn(async () => {
    throw new TypeError("upload unavailable");
  });
  vi.stubGlobal("fetch", fetchMock);
  const transport = createEvidenceFileTransport("/base");
  const attempt = transport.capture({
    stage: "slot",
    authority: {
      actorId: "actor",
      sessionIdentity: "session",
      incidentId,
      role: "editor",
      closed: false,
    },
    clientTxnId: "txn-file-failure",
    file: new File(["abc"], "failed.txt", { type: "text/plain" }),
  });
  expect(await transport.slot(attempt, new AbortController().signal)).toEqual({
    kind: "uncertain",
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
