import { afterEach, describe, expect, it, vi } from "vitest";
import { historyDiffFixture } from "../../testing/workbookHistoryTestSupport";
import type { HistoryAttempt } from "../history/workbookHistoryOperation";
import { createWorkbookRecordHistoryAdapter } from "./createWorkbookRecordHistoryAdapter";

const recordId = "20000000-0000-4000-8000-000000000001";
const incidentId = "10000000-0000-4000-8000-000000000001";
const changeSetId = "30000000-0000-4000-8000-000000000001";
const actorId = "40000000-0000-4000-8000-000000000001";
const attempt: HistoryAttempt = {
  id: "history-delete-one",
  actorId,
  incidentId,
  operation: "delete",
  subject: {
    kind: "live",
    recordId,
    rowVersion: 4,
    label: "Row",
    surfaceLabel: "Timeline",
    viewSchemaId: "cartulary.view.timeline.v2",
  },
  pending: {
    kind: "destructive",
    operation: "delete",
    recordId,
    rowVersion: 4,
  },
  body: JSON.stringify({
    base_row_version: 4,
    client_txn_id: "history-delete-one",
    reason: "Deleted from workbook history",
  }),
};
const receipt = {
  record_id: recordId,
  incident_id: incidentId,
  row_version: 5,
  change_set_id: changeSetId,
  deleted: true,
  deleted_at: "2026-09-10T00:00:00Z",
  deleted_by_user_id: actorId,
};
const response = (data: unknown) =>
  new Response(JSON.stringify({ data, meta: { request_id: "request" } }), {
    status: 200,
    headers: { "content-type": "application/json" },
  });
afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
describe("History captured transport", () => {
  it("admits complete semantic history and rejects malformed mandatory detail locally", async () => {
    const data = {
      record_id: recordId,
      incident_id: incidentId,
      row_version: 4,
      deleted: false,
      representation_generation: "cartulary.history.1",
      items: [
        {
          actor_user_id: actorId,
          source_actor_id: "imported-author",
          committed_at: "2026-09-21T00:00:00Z",
          history_item_ref: "hitem_semantic",
          operation: "patch",
          change_set_id: changeSetId,
          reversible: false,
          available_rollback_actions: [],
          diff_summary: historyDiffFixture("Updated text"),
        },
      ],
    };
    const readResponse = (value: unknown) =>
      new Response(
        JSON.stringify({
          data: value,
          meta: {
            request_id: "read",
            paging: { limit: 100, has_more: false, next_cursor: null },
          },
        }),
        { status: 200, headers: { "content-type": "application/json" } },
      );
    const fetch = vi.fn().mockImplementation(async () => readResponse(data));
    vi.stubGlobal("fetch", fetch);
    const port = createWorkbookRecordHistoryAdapter({
      apiBase: undefined,
      incidentId,
    });
    expect(
      await port.load(recordId, new AbortController().signal),
    ).toMatchObject({
      kind: "accepted",
      value: { items: [{ source_actor_id: "imported-author" }] },
    });
    const malformed = [
      { ...data, representation_generation: undefined },
      {
        ...data,
        items: [
          {
            ...data.items[0],
            diff_summary: {
              ...data.items[0]?.diff_summary,
              schema_id: "cartulary.history_diff.v0",
            },
          },
        ],
      },
      {
        ...data,
        items: [
          {
            ...data.items[0],
            diff_summary: {
              schema_id: "cartulary.history_diff.v1",
              summary: "Incomplete",
              units: [],
            },
          },
        ],
      },
      {
        ...data,
        items: [
          {
            ...data.items[0],
            diff_summary: {
              ...data.items[0]?.diff_summary,
              units: [{ target_kind: "record", target_id: recordId }],
            },
          },
        ],
      },
      {
        ...data,
        items: [
          {
            ...data.items[0],
            diff_summary: {
              ...data.items[0]?.diff_summary,
              units: [
                {
                  ...data.items[0]?.diff_summary.units[0],
                  changes: [
                    {
                      field_key: "evidence.title",
                      before: { state: "null" },
                      after: {
                        state: "present",
                        value: { upload_token: "secret" },
                      },
                    },
                  ],
                },
              ],
            },
          },
        ],
      },
    ];
    for (const value of malformed) {
      fetch.mockImplementationOnce(async () => readResponse(value));
      expect(
        await port.load(recordId, new AbortController().signal),
      ).toMatchObject({
        kind: "rejected",
        failure: { kind: "invalid_contract" },
      });
    }
  });
  it("sends the retained body exactly with current CSRF and preserves the full receipt", async () => {
    const fetch = vi.fn().mockResolvedValue(response(receipt));
    vi.stubGlobal("fetch", fetch);
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=current-csrf",
    );
    const port = createWorkbookRecordHistoryAdapter({
      apiBase: undefined,
      incidentId,
    });
    expect(
      await port.send(attempt, new AbortController().signal),
    ).toMatchObject({
      kind: "acknowledged",
      receipt: {
        kind: "delete",
        recordId,
        incidentId,
        rowVersion: 5,
        changeSetId,
        deleted: true,
        deletedByUserId: actorId,
      },
    });
    const init = fetch.mock.calls[0]?.[1] as RequestInit;
    expect(init.body).toBe(attempt.body);
    expect((init.headers as Headers).get("X-CSRF-Token")).toBe("current-csrf");
  });
  it.each([
    "restore",
    "history_entry",
    "change_set",
    "row_restore",
  ] as const)("retains the complete receipt and exact selector for %s", async (kind) => {
    const target =
      kind === "history_entry"
        ? { kind, history_entry_ref: "opaque-selector" }
        : kind === "change_set"
          ? { kind, change_set_id: changeSetId }
          : { kind: "row_restore" as const, restore_to_revision_no: 2 };
    const pending: HistoryAttempt["pending"] =
      kind === "restore"
        ? { kind: "destructive", operation: "restore", recordId, rowVersion: 4 }
        : {
            kind: "rollback",
            action: kind,
            recordId,
            rowVersion: 4,
            historyItemRef: "exact-item",
            target,
          };
    const captured: HistoryAttempt = {
      ...attempt,
      operation: kind === "restore" ? "restore" : "rollback",
      pending,
      body: JSON.stringify({
        base_row_version: 4,
        client_txn_id: attempt.id,
        reason: "Reviewed action",
        ...(kind === "restore" ? {} : { target }),
      }),
    };
    const data =
      kind === "restore"
        ? {
            ...receipt,
            deleted: false,
            deleted_at: null,
            deleted_by_user_id: null,
          }
        : {
            incident_id: incidentId,
            record_id: recordId,
            row_version: 5,
            target,
            target_change_set_id: changeSetId,
            rollback_change_set_id: changeSetId,
            affected_record_ids: [recordId, actorId],
          };
    const fetch = vi.fn().mockResolvedValue(response(data));
    vi.stubGlobal("fetch", fetch);
    const port = createWorkbookRecordHistoryAdapter({
      apiBase: undefined,
      incidentId,
    });
    const result = await port.send(captured, new AbortController().signal);
    expect(result).toMatchObject({
      kind: "acknowledged",
      receipt:
        kind === "restore"
          ? {
              kind,
              deleted: false,
              deletedAt: null,
              deletedByUserId: null,
              changeSetId,
            }
          : {
              kind: "rollback",
              target,
              targetChangeSetId: changeSetId,
              changeSetId,
              affectedRecordIds: [recordId, actorId],
            },
    });
    expect(fetch.mock.calls[0]?.[1].body).toBe(captured.body);
    if (kind !== "restore") {
      fetch.mockResolvedValueOnce(
        response({ ...data, affected_record_ids: [actorId, recordId] }),
      );
      expect(await port.send(captured, new AbortController().signal)).toEqual({
        kind: "uncertain",
      });
      fetch.mockResolvedValueOnce(
        response({
          ...data,
          target: { kind: "row_restore", restore_to_revision_no: 99 },
        }),
      );
      expect(await port.send(captured, new AbortController().signal)).toEqual({
        kind: "uncertain",
      });
    }
  });
  it("keeps malformed success transport loss and server failure uncertain", async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response({ ...receipt, record_id: actorId }))
      .mockResolvedValueOnce(response({ ...receipt, deleted: false }))
      .mockResolvedValueOnce(response({ ...receipt, incident_id: actorId }))
      .mockResolvedValueOnce(response({ ...receipt, row_version: 4 }))
      .mockResolvedValueOnce(response({}))
      .mockResolvedValueOnce(new Response("failure", { status: 503 }))
      .mockRejectedValueOnce(new Error("response lost"));
    vi.stubGlobal("fetch", fetch);
    const port = createWorkbookRecordHistoryAdapter({
      apiBase: undefined,
      incidentId,
    });
    for (let index = 0; index < 7; index++)
      expect(await port.send(attempt, new AbortController().signal)).toEqual({
        kind: "uncertain",
      });
  });
  it("distinguishes authoritative rejection from uncertainty", async () => {
    const fetch = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          error: {
            code: "client_txn_conflict",
            status: 409,
            message: "conflict",
            retryable: false,
            details: {},
            request_id: "request",
          },
        }),
        { status: 409, headers: { "content-type": "application/json" } },
      ),
    );
    vi.stubGlobal("fetch", fetch);
    const port = createWorkbookRecordHistoryAdapter({
      apiBase: undefined,
      incidentId,
    });
    expect(
      await port.send(attempt, new AbortController().signal),
    ).toMatchObject({
      kind: "rejected",
      failure: { kind: "client_txn_conflict" },
    });
  });
});
