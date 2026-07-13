import { gridShellTestId, rowCellTestId } from "@cartulary/ui-contracts";
import {
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  buildRecordChangedPayload,
  emitRecordChanged,
  successEnvelope,
  timelineRow,
  visibleGridRows,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import { TimelineWorkbook } from "./timeline/components/TimelineWorkbook";

describe("Phase 5 workbook evidence coverage", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let webSocketInstance: {
    onmessage: ((event: MessageEvent) => void) | null;
    close: () => void;
  } | null;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    webSocketInstance = null;
    vi.stubGlobal(
      "WebSocket",
      class {
        onmessage: ((event: MessageEvent) => void) | null = null;

        constructor() {
          webSocketInstance = this;
        }

        close() {}
      } as unknown as typeof WebSocket,
    );
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("Phase 5 U-5-08 reflects attached evidence counts on the workbook surface without forcing navigation", async () => {
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "timeline-1",
              rowVersion: 5,
              summary: "Endpoint screenshot",
              captureState: "rough",
              evidenceCount: 0,
              hasEvidence: false,
            }),
          ],
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [
            timelineRow({
              recordId: "timeline-1",
              rowVersion: 5,
              summary: "Endpoint screenshot",
              captureState: "rough",
              evidenceCount: 1,
              hasEvidence: true,
            }),
          ],
        }),
      );

    const { container } = render(
      <TimelineWorkbook incidentId="incident-1" currentIncidentRole="admin" />,
    );

    await screen.findByDisplayValue("Endpoint screenshot");
    expect(visibleGridRows(container).length).toBeGreaterThanOrEqual(1);
    const gridShellBeforeRefresh = screen.getByTestId(
      gridShellTestId(timelineViewSchemaId),
    );
    expect(webSocketInstance).not.toBeNull();

    emitRecordChanged(
      webSocketInstance,
      buildRecordChangedPayload({
        recordId: "timeline-1",
        rowVersion: 5,
        clientTxnId: "txn-evidence-attach",
        changedFieldKeys: ["timeline.evidence_count", "timeline.has_evidence"],
        affectedViews: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "invalidate",
          },
        ],
      }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(screen.getByTestId(gridShellTestId(timelineViewSchemaId))).toBe(
      gridShellBeforeRefresh,
    );
    await waitFor(() => {
      const committedRow = visibleGridRows(container).find(
        (row) => row.getAttribute("data-grid-record-id") === "timeline-1",
      );
      expect(committedRow).toBeTruthy();
      expect(
        within(committedRow as HTMLElement).getByDisplayValue(
          "Endpoint screenshot",
        ),
      ).toBeTruthy();
      expect(
        within(committedRow as HTMLElement).queryByTestId(
          rowCellTestId("timeline-1", "timeline.evidence_count"),
        ),
      ).toBeNull();
      expect(
        within(committedRow as HTMLElement).queryByTestId(
          rowCellTestId("timeline-1", "timeline.has_evidence"),
        ),
      ).toBeNull();
    });
  });
});
