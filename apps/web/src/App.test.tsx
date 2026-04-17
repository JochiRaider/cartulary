import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  buildCreatePayload,
  createDraftRow,
  ensureDraftRow,
  TimelineWorkbook,
} from "./App";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

describe("Phase 3 Timeline workbook", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let webSocketMock: typeof WebSocket;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    webSocketMock = class {
      onmessage: ((event: MessageEvent) => void) | null = null;

      close() {}
    } as unknown as typeof WebSocket;
    vi.stubGlobal("WebSocket", webSocketMock);
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("Phase 3 U-3-05 keeps a continuation row after autosaved create", async () => {
    const draftRow = createDraftRow(1);
    draftRow.values.summary = "First timeline fact";

    const createPayload = buildCreatePayload(draftRow, "timeline-client-1");
    expect(createPayload).toEqual({
      client_txn_id: "timeline-client-1",
      "timeline.summary": "First timeline fact",
    });

    const persistedRow = {
      ...createDraftRow(99),
      key: "record-1",
      recordId: "record-1",
      rowVersion: 1,
      captureState: "rough",
      values: {
        occurredAt: "",
        summary: "First timeline fact",
        details: "",
        sourceText: "",
      },
      committedValues: {
        occurredAt: "",
        summary: "First timeline fact",
        details: "",
        sourceText: "",
      },
    };

    const continuedRows = ensureDraftRow([persistedRow], 2);
    expect(continuedRows).toHaveLength(2);
    expect(continuedRows[0]?.recordId).toBe("record-1");
    expect(continuedRows[1]?.recordId).toBeNull();
    expect(continuedRows[1]?.values.summary).toBe("");
  });

  it("Phase 3 component autosaves on Enter, Tab, and paste completion", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-2",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Alpha enter",
          captureState: "enriched",
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-3",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 3,
          summary: "Alpha tab",
          captureState: "enriched",
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-4",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 4,
          summary: "Alpha tab",
          sourceText: "Pasted transcript",
          captureState: "enriched",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(summaryInput, { target: { value: "Alpha enter" } });
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractBody(fetchMock, 1).base_row_version).toBe(1);
    expect(extractBody(fetchMock, 1).changes[0]).toEqual({
      field_key: "timeline.summary",
      value: "Alpha enter",
    });

    const tabInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(tabInput, { target: { value: "Alpha tab" } });
    fireEvent.keyDown(tabInput, { key: "Tab" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(extractBody(fetchMock, 2).base_row_version).toBe(2);
    expect(extractBody(fetchMock, 2).changes[0]).toEqual({
      field_key: "timeline.summary",
      value: "Alpha tab",
    });

    const sourceText = (await screen.findByTestId(
      "row-record-1-sourceText",
    )) as HTMLTextAreaElement;
    fireEvent.change(sourceText, { target: { value: "Pasted transcript" } });
    fireEvent.paste(sourceText);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    expect(extractBody(fetchMock, 3).base_row_version).toBe(3);
    expect(extractBody(fetchMock, 3).changes[0]).toEqual({
      field_key: "timeline.source_text",
      value: "Pasted transcript",
    });
  });

  it("Phase 3 component shows the exact save-state labels Syncing, Saved, and Conflict", async () => {
    const pendingPatch = deferred<Response>();

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(pendingPatch.promise);
    fetchMock.mockResolvedValueOnce(errorEnvelope("row_version_conflict", 409));

    render(<TimelineWorkbook incidentId="incident-1" />);

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(summaryInput, { target: { value: "Pending save" } });
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(screen.getByText("Syncing")).toBeTruthy();
    });

    pendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-5",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Pending save",
          captureState: "enriched",
        }),
      }),
    );

    await waitFor(() => {
      expect(screen.getByText("Saved")).toBeTruthy();
    });

    const conflictedInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(conflictedInput, { target: { value: "Conflict value" } });
    fireEvent.blur(conflictedInput);

    await waitFor(() => {
      expect(screen.getByText("Conflict")).toBeTruthy();
    });
  });
});

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void;
  let reject!: (reason?: unknown) => void;

  const promise = new Promise<T>((innerResolve, innerReject) => {
    resolve = innerResolve;
    reject = innerReject;
  });

  return { promise, resolve, reject };
}

function extractBody(fetchSpy: ReturnType<typeof vi.fn>, index: number) {
  return JSON.parse(String(fetchSpy.mock.calls[index]?.[1]?.body ?? "{}")) as {
    base_row_version: number;
    changes: Array<{ field_key: string; value: string | null }>;
  };
}

function successEnvelope(data: unknown, status = 200) {
  return new Response(
    JSON.stringify({
      data,
      meta: { request_id: `req-${status}` },
    }),
    {
      status,
      headers: { "Content-Type": "application/json" },
    },
  );
}

function errorEnvelope(code: string, status: number) {
  return new Response(
    JSON.stringify({
      error: {
        status,
        code,
        message: code,
        request_id: "req-error",
        retryable: false,
        details: {},
      },
    }),
    {
      status,
      headers: { "Content-Type": "application/json" },
    },
  );
}

function timelineRow({
  recordId,
  rowVersion,
  occurredAt = "",
  summary = "",
  details = "",
  sourceText = "",
  captureState,
}: {
  recordId: string;
  rowVersion: number;
  occurredAt?: string;
  summary?: string;
  details?: string;
  sourceText?: string;
  captureState: string;
}) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "timeline.occurred_at": { value: occurredAt },
      "timeline.summary": { value: summary },
      "timeline.details": { value: details },
      "timeline.source_text": { value: sourceText },
      "timeline.capture_state": { value: captureState },
    },
  };
}
