import type {
  CreateViewRowRequest,
  PatchRecordRequest,
} from "@cartulary/protocol-ts/http";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { describe, expect, expectTypeOf, it, vi } from "vitest";
import { fetchRecordHistory, fetchRecordHistoryCount } from "./history";
import {
  assertRecordFieldMutationAnchor,
  fillDownGridCells,
} from "./mutationAnchors";
import {
  createViewRow,
  patchRecord,
  readWorkbookMutation,
  waitForViewRow,
  waitForViewRowByCell,
} from "./query";

const surface = timelineViewSchemaId;

describe("workbook row mutation support", () => {
  it("reports the view and record identity when row polling times out", async () => {
    const fetch = vi.fn(async () => ({
      headers: () => ({}),
      json: async () => ({ data: { rows: [] } }),
      ok: () => true,
      status: () => 200,
    }));
    const page = { request: { fetch } } as unknown as Parameters<
      typeof waitForViewRow
    >[0];

    await expect(
      waitForViewRow(page, "incident-1", surface, "record-missing", {
        mode: "single_attempt",
      }),
    ).rejects.toThrow(
      /cartulary\.view\.timeline\.v2 default query should include created row record-missing[\s\S]*incident_id=incident-1[\s\S]*view_schema_id=cartulary\.view\.timeline\.v2[\s\S]*last_rows=\[\]/u,
    );
    expect(fetch).toHaveBeenCalledOnce();
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(
        /\/api\/v1\/incidents\/incident-1\/views\/cartulary\.view\.timeline\.v2\/query$/u,
      ),
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("reports the view field and context when cell polling times out", async () => {
    const fetch = vi.fn(async () => ({
      headers: () => ({}),
      json: async () => ({ data: { rows: [] } }),
      ok: () => true,
      status: () => 200,
    }));
    const page = { request: { fetch } } as unknown as Parameters<
      typeof waitForViewRowByCell
    >[0];

    await expect(
      waitForViewRowByCell(
        page,
        "incident-1",
        surface,
        "timeline.activity_synopsis_text",
        "missing projection",
        {
          diagnosticContext: "assessment projection",
          mode: "single_attempt",
        },
      ),
    ).rejects.toThrow(
      /timeline\.activity_synopsis_text="missing projection"[\s\S]*incident_id=incident-1[\s\S]*view_schema_id=cartulary\.view\.timeline\.v2[\s\S]*context=assessment projection[\s\S]*last_rows=\[\]/u,
    );
    expect(fetch).toHaveBeenCalledOnce();
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(
        /\/api\/v1\/incidents\/incident-1\/views\/cartulary\.view\.timeline\.v2\/query$/u,
      ),
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("validates mutation responses with operation-aware diagnostics", async () => {
    expectTypeOf(createViewRow)
      .parameter(3)
      .toEqualTypeOf<CreateViewRowRequest>();
    expectTypeOf(patchRecord).parameter(2).toEqualTypeOf<PatchRecordRequest>();
    expectTypeOf<{
      "timeline.activity_synopsis_text": string;
    }>().not.toExtend<CreateViewRowRequest>();
    expectTypeOf<{
      base_row_version: number;
      client_txn_id: string;
      view_schema_id: string;
    }>().not.toExtend<PatchRecordRequest>();

    const validPayload = {
      data: {
        change_set_id: "00000000-0000-4000-8000-000000000002",
        row: {
          cells: {},
          record_id: "00000000-0000-4000-8000-000000000001",
          row_version: 2,
        },
        view_schema_id: surface,
      },
      meta: { request_id: "request-workbook-mutation" },
    };

    await expect(
      readWorkbookMutation(
        workbookMutationResponse(validPayload),
        "patchRecord",
      ),
    ).resolves.toEqual(validPayload);
    await expect(
      readWorkbookMutation(
        workbookMutationResponse({ data: {} }),
        "patchRecord",
      ),
    ).rejects.toThrow(
      "public HTTP operation patchRecord failed: invalid_public_contract_response",
    );
  });

  it("validates record-history pagination and rejects malformed success payloads", async () => {
    const validPayload = {
      data: {
        deleted: false,
        incident_id: "00000000-0000-4000-8000-000000000002",
        items: [],
        record_id: "00000000-0000-4000-8000-000000000001",
        row_version: 3,
      },
      meta: {
        paging: { has_more: false, limit: 25, next_cursor: null },
        request_id: "request-record-history",
      },
    };
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(validPayload))
      .mockResolvedValueOnce(jsonResponse(validPayload))
      .mockResolvedValueOnce(jsonResponse({ data: {} }));
    const page = { request: { fetch } } as unknown as Parameters<
      typeof fetchRecordHistory
    >[0];

    await expect(
      fetchRecordHistory(page, validPayload.data.record_id, {
        cursorToken: "cursor-next",
        limit: 25,
      }),
    ).resolves.toEqual(validPayload.data);
    expect(fetch).toHaveBeenNthCalledWith(
      1,
      expect.stringMatching(
        /\/api\/v1\/records\/00000000-0000-4000-8000-000000000001\/history\?cursor_token=cursor-next&limit=25$/u,
      ),
      { method: "GET" },
    );

    await expect(
      fetchRecordHistoryCount(page, validPayload.data.record_id),
    ).resolves.toBe(0);
    expect(fetch).toHaveBeenNthCalledWith(
      2,
      expect.stringMatching(/\/history\?limit=100$/u),
      { method: "GET" },
    );

    await expect(
      fetchRecordHistory(page, validPayload.data.record_id),
    ).rejects.toThrow(/getRecordHistory failed with HTTP 502/u);
  });

  it("posts the stable fill-down envelope through the same-origin page context", async () => {
    const evaluate = vi.fn(async (_callback, argument) => {
      expect(argument).toEqual({
        data: {
          client_txn_id: "txn-fixed",
          field_key: "timeline.activity_synopsis_text",
          kind: "fill_down_v1",
          targets: [
            { base_row_version: 2, record_id: "record-1" },
            { base_row_version: 4, record_id: "record-2" },
          ],
          value: "triaged",
          view_schema_id: surface,
        },
        headers: {
          "content-type": "application/json",
          "x-csrf-token": "csrf",
        },
        method: "POST",
        url: `/api/v1/incidents/incident-1/views/${surface}/bulk-mutations`,
      });
      return { ok: true, status: 202 };
    });

    const response = await fillDownGridCells({
      apiBase: "https://ignored.example",
      clientTxnId: "txn-fixed",
      csrfHeaders: { "x-csrf-token": "csrf" },
      fieldKey: "timeline.activity_synopsis_text",
      incidentId: "incident-1",
      page: { evaluate },
      surface,
      targetRecords: [
        { baseRowVersion: 2, recordId: "record-1" },
        { baseRowVersion: 4, recordId: "record-2" },
      ],
      value: "triaged",
    });

    expect(response.ok()).toBe(true);
    expect(response.status?.()).toBe(202);
    expect(evaluate).toHaveBeenCalledOnce();
  });

  it("uses the request context as the explicit fallback and preserves omitted CSRF headers", async () => {
    const post = vi.fn(async () => ({
      ok: () => false,
      status: () => 409,
    }));

    const response = await fillDownGridCells({
      apiBase: "https://cartulary.test",
      clientTxnId: "txn-fixed",
      fieldKey: "host.hostname",
      incidentId: "incident-1",
      page: { request: { post } },
      surface: "cartulary.view.hosts.v1",
      targetRecords: [{ baseRowVersion: 1, recordId: "host-1" }],
      value: "host.example",
    });

    expect(response.status?.()).toBe(409);
    expect(post).toHaveBeenCalledWith(
      "https://cartulary.test/api/v1/incidents/incident-1/views/cartulary.view.hosts.v1/bulk-mutations",
      {
        data: {
          client_txn_id: "txn-fixed",
          field_key: "host.hostname",
          kind: "fill_down_v1",
          targets: [{ base_row_version: 1, record_id: "host-1" }],
          value: "host.example",
          view_schema_id: "cartulary.view.hosts.v1",
        },
      },
    );
  });

  it("requires an exact record, field, and optionally supplied value anchor", () => {
    const body = {
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "triaged",
        },
      ],
    };

    expect(() =>
      assertRecordFieldMutationAnchor({
        actualRecordId: "record-1",
        body,
        expectedRecordId: "record-1",
        fieldKey: "timeline.activity_synopsis_text",
      }),
    ).not.toThrow();
    expect(() =>
      assertRecordFieldMutationAnchor({
        actualRecordId: "record-1",
        body,
        expectedRecordId: "record-1",
        expectedValue: "different",
        fieldKey: "timeline.activity_synopsis_text",
      }),
    ).toThrow(/mutation value different/);
    expect(() =>
      assertRecordFieldMutationAnchor({
        actualRecordId: "record-2",
        body,
        expectedRecordId: "record-1",
        fieldKey: "timeline.activity_synopsis_text",
      }),
    ).toThrow(/Expected mutation for record_id record-1/);
  });
});

function workbookMutationResponse(payload: unknown) {
  return {
    ok: () => true,
    request: () => ({
      method: () => "PATCH",
      postData: () => JSON.stringify({ base_row_version: 1 }),
    }),
    status: () => 200,
    text: async () => JSON.stringify(payload),
    url: () => "https://cartulary.test/api/v1/records/record-1",
  } as unknown as Parameters<typeof readWorkbookMutation>[0];
}

function jsonResponse(payload: unknown) {
  return {
    headers: () => ({}),
    json: async () => payload,
    ok: () => true,
    status: () => 200,
  };
}
