import { describe, expect, it, vi } from "vitest";

import {
  assertRecordFieldMutationAnchor,
  fillDownGridCells,
} from "./mutationAnchors";

const surface = "cartulary.view.timeline.v2";

describe("workbook row mutation support", () => {
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
