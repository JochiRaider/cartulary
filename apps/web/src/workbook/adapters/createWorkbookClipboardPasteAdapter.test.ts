import { afterEach, describe, expect, it, vi } from "vitest";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { createWorkbookClipboardPasteAdapter } from "./createWorkbookClipboardPasteAdapter";
import type { WorkbookClipboardPasteInput } from "./WorkbookClipboardPastePort";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";

const input: WorkbookClipboardPasteInput = {
  clipboard_text: "2026-09-03\tvalidated paste",
  columns: ["timeline.date_entered_text", "timeline.activity_synopsis_text"],
  format: "tsv",
  start_field_key: "timeline.date_entered_text",
  targets: [
    {
      base_row_version: 2,
      kind: "record",
      record_id: recordId,
    },
  ],
  view_schema_id: timelineViewSchemaId,
};

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("createWorkbookClipboardPasteAdapter", () => {
  it("sends one exact generated request and preserves typed rows and conflicts", async () => {
    const conflict = {
      base_row_version: 2,
      client_value: "client",
      conflict_resolution_class: "text_compare_merge",
      conflict_token: "conflict-token",
      current_row_version: 3,
      field_key: "timeline.activity_synopsis_text",
      record_id: recordId,
      server_value: "server",
    };
    const row = {
      cells: {},
      record_id: recordId,
      row_version: 3,
    };
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({
        data: {
          change_set_id: "30000000-0000-4000-8000-000000000001",
          conflicts: [conflict],
          rows: [row],
          view_schema_id: timelineViewSchemaId,
        },
        meta: { request_id: "request-paste" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const onClientTxnId = vi.fn();
    const paste = createWorkbookClipboardPasteAdapter({
      apiBase: "/base",
      incidentId,
      transactionIds: { create: () => "txn-paste" },
    });

    await expect(
      paste.paste({ ...input, onClientTxnId }),
    ).resolves.toMatchObject({
      clientTxnId: "txn-paste",
      outcome: {
        kind: "accepted",
        value: {
          conflicts: [conflict],
          rows: [row],
          viewSchemaId: timelineViewSchemaId,
        },
      },
    });
    expect(onClientTxnId).toHaveBeenCalledWith("txn-paste");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(String(fetchMock.mock.calls[0]?.[0])).toBe(
      `/base/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
    );
    expect(requestBody(fetchMock)).toEqual({
      client_txn_id: "txn-paste",
      clipboard_text: input.clipboard_text,
      columns: input.columns,
      format: "tsv",
      start_field_key: input.start_field_key,
      targets: input.targets,
      view_schema_id: timelineViewSchemaId,
    });
  });

  it("fails closed before transport for invalid targets and secure-ID failure", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const create = vi.fn(() => {
      throw new Error("unavailable");
    });
    const onClientTxnId = vi.fn();
    const paste = createWorkbookClipboardPasteAdapter({
      apiBase: undefined,
      incidentId,
      transactionIds: { create },
    });

    await expect(
      paste.paste({ ...input, columns: [" "], onClientTxnId }),
    ).resolves.toMatchObject({
      clientTxnId: null,
      outcome: { kind: "rejected", failure: { kind: "invalid_contract" } },
    });
    expect(create).not.toHaveBeenCalled();
    await expect(
      paste.paste({ ...input, onClientTxnId }),
    ).resolves.toMatchObject({
      clientTxnId: null,
      outcome: { kind: "rejected", failure: { kind: "terminal" } },
    });
    expect(onClientTxnId).not.toHaveBeenCalled();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects a mismatched response surface and contains transport failure", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          data: {
            conflicts: [],
            rows: [],
            view_schema_id: hostsViewSchemaId,
          },
          meta: { request_id: "wrong-view" },
        }),
      )
      .mockRejectedValueOnce(new TypeError("response lost"));
    vi.stubGlobal("fetch", fetchMock);
    let sequence = 0;
    const paste = createWorkbookClipboardPasteAdapter({
      apiBase: undefined,
      incidentId,
      transactionIds: { create: () => `txn-${++sequence}` },
    });

    await expect(paste.paste(input)).resolves.toMatchObject({
      clientTxnId: "txn-1",
      outcome: { kind: "rejected", failure: { kind: "invalid_contract" } },
    });
    await expect(paste.paste(input)).resolves.toMatchObject({
      clientTxnId: "txn-2",
      outcome: { kind: "rejected", failure: { kind: "retryable" } },
    });
  });
});

function jsonResponse(payload: unknown): Response {
  return new Response(JSON.stringify(payload), {
    headers: { "content-type": "application/json" },
    status: 200,
  });
}

function requestBody(fetchMock: ReturnType<typeof vi.fn>): unknown {
  const init = fetchMock.mock.calls[0]?.[1];
  if (init === undefined || typeof init !== "object" || !("body" in init)) {
    throw new Error("expected a request body");
  }
  return JSON.parse(String(init.body));
}
