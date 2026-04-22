import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  buildRecordChangedPayload,
  deferred,
  emitRecordChanged,
  successEnvelope,
  timelineRow,
  timelineViewSchemaId,
} from "./timelineWorkbookTestSupport";
import { TimelineWorkbook } from "./WorkbookShell";

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

    emitRecordChanged(
      webSocketInstance,
      buildRecordChangedPayload({
        recordId: "record-1",
        rowVersion: 2,
        clientTxnId: "timeline-client-1",
        changeSetId: "change-set-socket",
        changedFieldKeys: ["timeline.summary"],
      }),
    );

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
