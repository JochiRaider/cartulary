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
  let webSocketInstance: {
    onmessage: ((event: MessageEvent) => void) | null;
    close: () => void;
  } | null;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    webSocketInstance = null;
    webSocketMock = class {
      onmessage: ((event: MessageEvent) => void) | null = null;

      constructor() {
        webSocketInstance = this;
      }

      close() {}
    } as unknown as typeof WebSocket;
    vi.stubGlobal("WebSocket", webSocketMock);
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels", async () => {
    const draftRow = createDraftRow(1);
    draftRow.values.summary = "First timeline fact";

    expect(buildCreatePayload(draftRow, "timeline-client-1")).toEqual({
      client_txn_id: "timeline-client-1",
      "timeline.summary": "First timeline fact",
    });

    const continuedRows = ensureDraftRow(
      [
        {
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
        },
      ],
      2,
    );
    expect(continuedRows).toHaveLength(2);
    expect(continuedRows[1]?.recordId).toBeNull();
    expect(continuedRows[1]?.values.summary).toBe("");

    const pendingBlurPatch = deferred<Response>();

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
          summary: "Updated via enter",
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
          summary: "Updated via tab",
          captureState: "enriched",
        }),
      }),
    );
    fetchMock.mockReturnValueOnce(pendingBlurPatch.promise);
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-4",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 4,
          summary: "Updated via blur",
          sourceText: "Pasted transcript",
          captureState: "enriched",
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(errorEnvelope("row_version_conflict", 409));

    render(<TimelineWorkbook incidentId="incident-1" />);

    expect((await screen.findByTestId("save-state")).textContent).toBe("Saved");
    expect(screen.queryByRole("button", { name: /^save$/i })).toBeNull();

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(summaryInput, { target: { value: "Updated via enter" } });
    await waitFor(() => {
      expect(summaryInput.value).toBe("Updated via enter");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractBody(fetchMock, 1).base_row_version).toBe(1);
    expect(extractBody(fetchMock, 1).changes[0]).toEqual({
      field_key: "timeline.summary",
      value: "Updated via enter",
    });
    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "2",
      );
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });

    const tabInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(tabInput, { target: { value: "Updated via tab" } });
    await waitFor(() => {
      expect(tabInput.value).toBe("Updated via tab");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.keyDown(tabInput, { key: "Tab" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(extractBody(fetchMock, 2).base_row_version).toBe(2);
    expect(extractBody(fetchMock, 2).changes[0]).toEqual({
      field_key: "timeline.summary",
      value: "Updated via tab",
    });
    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "3",
      );
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });

    const blurInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(blurInput, { target: { value: "Updated via blur" } });
    await waitFor(() => {
      expect(blurInput.value).toBe("Updated via blur");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.blur(blurInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    expect(extractBody(fetchMock, 3).base_row_version).toBe(3);
    expect(extractBody(fetchMock, 3).changes[0]).toEqual({
      field_key: "timeline.summary",
      value: "Updated via blur",
    });
    await waitFor(() => {
      expect(screen.getByTestId("save-state").textContent).toBe("Syncing");
    });

    pendingBlurPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-4",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 4,
          summary: "Updated via blur",
          captureState: "enriched",
        }),
      }),
    );
    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "4",
      );
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });

    const sourceText = (await screen.findByTestId(
      "row-record-1-sourceText",
    )) as HTMLTextAreaElement;
    fireEvent.change(sourceText, { target: { value: "Pasted transcript" } });
    await waitFor(() => {
      expect(sourceText.value).toBe("Pasted transcript");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.paste(sourceText);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(5);
    });
    expect(extractBody(fetchMock, 4).base_row_version).toBe(4);
    expect(extractBody(fetchMock, 4).changes[0]).toEqual({
      field_key: "timeline.source_text",
      value: "Pasted transcript",
    });
    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "4",
      );
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });

    const conflictedInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(conflictedInput, { target: { value: "Conflict value" } });
    await waitFor(() => {
      expect(conflictedInput.value).toBe("Conflict value");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.blur(conflictedInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(6);
      expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
    });
  });

  it("Phase 3 support suppresses self-originated websocket invalidations without refocusing the draft row", async () => {
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

    render(<TimelineWorkbook incidentId="incident-1" />);

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    summaryInput.focus();
    fireEvent.change(summaryInput, { target: { value: "Alpha enter" } });
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    emitRecordChanged(webSocketInstance, {
      record_id: "record-1",
      row_version: 2,
      change_set_id: "change-set-socket",
      client_txn_id: "timeline-client-1",
      actor_user_id: "user-1",
      changed_field_keys: ["timeline.summary"],
      affected_views: [
        {
          view_schema_id: timelineViewSchemaId,
          change_kind: "invalidate",
        },
      ],
    });

    pendingPatch.resolve(
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

    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "2",
      );
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(document.activeElement).not.toBe(
      screen.getByTestId("draft-row-summary"),
    );
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

function emitRecordChanged(
  socket:
    | {
        onmessage: ((event: MessageEvent) => void) | null;
      }
    | null
    | undefined,
  payload: {
    record_id: string;
    row_version: number;
    change_set_id: string;
    client_txn_id: string;
    actor_user_id: string;
    changed_field_keys: string[];
    affected_views: Array<{
      view_schema_id: string;
      change_kind: string;
    }>;
  },
) {
  socket?.onmessage?.({
    data: JSON.stringify({
      type: "record_changed",
      payload,
    }),
  } as MessageEvent);
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
