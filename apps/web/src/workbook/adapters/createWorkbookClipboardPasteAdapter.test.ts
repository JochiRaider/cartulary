import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  fullWorkbookViewRow,
  successEnvelope,
} from "../../testing/timelineWorkbookTestSupport";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type { WorkbookBatchPlan } from "../runtime/workbookBatchOperation";
import {
  createWorkbookBatchTransport,
  validateWorkbookBatchReceipt,
} from "./createWorkbookBatchTransport";
import { createWorkbookClipboardPasteAdapter } from "./createWorkbookClipboardPasteAdapter";
import type { WorkbookClipboardPasteInput } from "./WorkbookClipboardPastePort";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";
const createdId = "20000000-0000-4000-8000-000000000002";
const changeSetId = "30000000-0000-4000-8000-000000000001";
const field = "timeline.activity_synopsis_text";
const authority = {
  actorId: "actor",
  sessionIdentity: "session",
  incidentId,
  role: "editor" as const,
  closed: false,
};
const input: WorkbookClipboardPasteInput = {
  clipboard_text: "first\nsecond",
  columns: [field],
  format: "tsv",
  start_field_key: field,
  targets: [
    { kind: "record", record_id: recordId, base_row_version: 2 },
    { kind: "create" },
  ],
  view_schema_id: timelineViewSchemaId,
};
const plan: WorkbookBatchPlan = {
  operation: "pasteWorkbookClipboard",
  request: input,
  recordIds: [recordId],
};
const conflict = {
  base_row_version: 2,
  current_row_version: 3,
  record_id: recordId,
  field_key: field,
  conflict_resolution_class: "text_compare_merge",
  conflict_token: "token",
  client_value: "first",
  server_value: "server",
  base_value: "base",
  server_updated_by: recordId,
  server_updated_at: "2026-09-13T00:00:00Z",
};
const row = (
  id: string,
  version: number,
  view: string = timelineViewSchemaId,
) => fullWorkbookViewRow(requireViewContract(view), id, version, {});
const transport = () =>
  createWorkbookBatchTransport({ apiBase: "/base", incidentId });
afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
describe("createWorkbookClipboardPasteAdapter", () => {
  it("sends one exact generated request and preserves typed rows and conflicts", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>(async () =>
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        rows: [row(createdId, 1)],
        conflicts: [conflict],
        change_set_id: changeSetId,
      }),
    );
    vi.stubGlobal("fetch", fetch);
    const adapter = transport();
    const attempt = adapter.capture(plan, authority, "txn-paste");
    const result = await adapter.send(attempt, new AbortController().signal);
    expect(result).toMatchObject({
      kind: "acknowledged",
      receipt: {
        changeSetId,
        conflicts: [conflict],
        rows: [{ record_id: createdId, row_version: 1 }],
      },
    });
    expect(fetch).toHaveBeenCalledOnce();
    expect(String(fetch.mock.calls[0]?.[0])).toBe(
      `/base/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
    );
    expect(JSON.parse(attempt.body)).toEqual({
      ...input,
      client_txn_id: "txn-paste",
    });
  });
  it("rejects invalid capture and delegates action identity to retained admission", () => {
    const adapter = transport();
    expect(() =>
      adapter.capture(
        { ...plan, request: { ...input, columns: [" "] } },
        authority,
        "txn",
      ),
    ).toThrow();
    const admit = vi.fn(() => "retained");
    const paste = createWorkbookClipboardPasteAdapter({ admit });
    const admission = { delivery: {} };
    expect(paste.paste(input, admission)).toBe("retained");
    expect(admit).toHaveBeenCalledWith(plan, admission);
  });
  it("retains uncertainty for malformed success and replays captured create bytes", async () => {
    const fetch = vi
      .fn()
      .mockRejectedValueOnce(new TypeError("response lost"))
      .mockResolvedValueOnce(
        successEnvelope({
          view_schema_id: timelineViewSchemaId,
          rows: [row(recordId, 3)],
          conflicts: [],
          change_set_id: changeSetId,
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          view_schema_id: timelineViewSchemaId,
          rows: [row(recordId, 3), row(createdId, 1)],
          conflicts: [],
          change_set_id: changeSetId,
        }),
      );
    vi.stubGlobal("fetch", fetch);
    const adapter = transport();
    const attempt = adapter.capture(plan, authority, "txn-paste");
    const signal = new AbortController().signal;
    expect(await adapter.send(attempt, signal)).toEqual({ kind: "uncertain" });
    expect(await adapter.send(attempt, signal)).toEqual({ kind: "uncertain" });
    expect(await adapter.send(attempt, signal)).toMatchObject({
      kind: "acknowledged",
    });
    expect(fetch.mock.calls.map((call) => call[1].body)).toEqual([
      attempt.body,
      attempt.body,
      attempt.body,
    ]);
  });
  it("validates complete fill and tag receipts including conflicts-only and no-op outcomes", () => {
    for (const kind of [
      "fill_down_v1",
      "multi_row_tag_assignment_v1",
    ] as const) {
      const targetField = kind === "fill_down_v1" ? field : "timeline.tags";
      const bulk: WorkbookBatchPlan = {
        operation: "applyWorkbookBulkMutation",
        recordIds: [recordId],
        request: {
          kind,
          view_schema_id: timelineViewSchemaId,
          field_key: field,
          value: "value",
          tag_name: "triaged",
          targets: [{ record_id: recordId, base_row_version: 2 }],
        },
      };
      const attempt = transport().capture(bulk, authority, "txn-bulk");
      const c = {
        ...conflict,
        field_key: targetField,
        conflict_resolution_class:
          kind === "fill_down_v1" ? "text_compare_merge" : "collection_review",
      };
      const data = {
        view_schema_id: timelineViewSchemaId,
        rows: [],
        conflicts: [c],
      };
      expect(validateWorkbookBatchReceipt(attempt, data)).toMatchObject({
        changeSetId: null,
        rows: [],
        conflicts: [c],
      });
      expect(
        validateWorkbookBatchReceipt(attempt, {
          ...data,
          change_set_id: changeSetId,
        }),
      ).toBeNull();
      expect(
        validateWorkbookBatchReceipt(attempt, {
          ...data,
          conflicts: [{ ...c, record_id: createdId }],
        }),
      ).toBeNull();
      expect(
        validateWorkbookBatchReceipt(attempt, { ...data, conflicts: [c, c] }),
      ).toBeNull();
      expect(
        validateWorkbookBatchReceipt(attempt, { ...data, conflicts: [] }),
      ).toMatchObject({ rows: [], conflicts: [] });
      expect(
        validateWorkbookBatchReceipt(attempt, {
          ...data,
          rows: [row(recordId, 3)],
          change_set_id: changeSetId,
        }),
      ).not.toBeNull();
    }
  });
  it("accepts ordered repeated Entity reuse and historical empty-conflict receipts", () => {
    for (const [view, entityType] of [
      [hostsViewSchemaId, "host"],
      [identitiesViewSchemaId, "identity"],
    ] as const) {
      const entity: WorkbookBatchPlan = {
        operation: "pasteWorkbookClipboard",
        entityType,
        recordIds: [],
        request: {
          ...input,
          view_schema_id: view,
          columns: [`${entityType}.display_name`],
          start_field_key: `${entityType}.display_name`,
          targets: [{ kind: "create" }, { kind: "create" }],
        },
      };
      const attempt = transport().capture(entity, authority, "txn-entity");
      const data = {
        view_schema_id: view,
        rows: [row(recordId, 5, view), row(recordId, 5, view)],
        change_set_id: changeSetId,
      };
      expect(validateWorkbookBatchReceipt(attempt, data)).toMatchObject({
        rows: [{ record_id: recordId }, { record_id: recordId }],
        conflicts: [],
      });
      expect(
        validateWorkbookBatchReceipt(attempt, { ...data, conflicts: null }),
      ).toBeNull();
      expect(
        validateWorkbookBatchReceipt(attempt, {
          ...data,
          rows: [row(recordId, 5, view)],
        }),
      ).toBeNull();
    }
  });
});
