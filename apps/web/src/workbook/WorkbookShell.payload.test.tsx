import {
  draftCellTestId,
  draftRowCreateButtonTestId,
  gridDraftRowSelector,
  gridShellTestId,
  rowCellTestId,
  saveStateTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  fireEvent,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { renderTimelineWorkbook } from "../testing/timelineWorkbookRenderTestSupport";
import {
  cleanupTimelineWorkbookTestGlobals,
  extractTimelineJSONBody,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  visibleGridRows,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import {
  buildCreatePayload,
  createDraftRow,
} from "./timeline/models/workbookTimelineModel";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

describe("Timeline workbook payload coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("builds zero-field Timeline create payloads only for explicit blank-row creation", () => {
    const draftRow = createDraftRow(1);

    expect(buildCreatePayload(draftRow, "timeline-client-blank")).toBeNull();
    expect(
      buildCreatePayload(draftRow, "timeline-client-blank", {
        allowZeroFieldCreate: true,
      }),
    ).toEqual({
      client_txn_id: "timeline-client-blank",
    });

    draftRow.values.activitySynopsisText = "First timeline fact";
    expect(buildCreatePayload(draftRow, "timeline-client-1")).toEqual({
      client_txn_id: "timeline-client-1",
      "timeline.activity_synopsis_text": "First timeline fact",
    });
  });

  it("creates an explicit blank Timeline row with only client_txn_id and suppresses duplicate pending submits", async () => {
    const pendingCreate = deferred<Response>();

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      }),
    );
    fetchMock.mockReturnValueOnce(pendingCreate.promise);

    renderTimelineWorkbook();

    expect((await screen.findByTestId(saveStateTestId())).textContent).toBe(
      "Saved",
    );

    const gridShell = screen.getByTestId(gridShellTestId(timelineViewSchemaId));
    const draftRows = Array.from(
      gridShell.querySelectorAll<HTMLDivElement>(gridDraftRowSelector()),
    );
    const draftGridRow = draftRows.at(-1);
    expect(draftGridRow).toBeTruthy();
    const createButton = within(draftGridRow as HTMLDivElement).getByTestId(
      draftRowCreateButtonTestId(),
    );
    fireEvent.mouseDown(createButton);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    fireEvent.mouseDown(createButton);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(extractTimelineJSONBody(fetchMock, 1)).toEqual({
      client_txn_id: expect.any(String),
    });

    await act(async () => {
      pendingCreate.resolve(
        successEnvelope(
          {
            view_schema_id: timelineViewSchemaId,
            change_set_id: "30000000-0000-4000-8000-000000000001",
            row: timelineRow({
              recordId: "20000000-0000-4000-8000-000000000010",
              rowVersion: 1,
              captureState: "rough",
            }),
          },
          201,
        ),
      );
      await pendingCreate.promise;
      await new Promise((resolve) => window.setTimeout(resolve, 0));
      await new Promise((resolve) => window.setTimeout(resolve, 0));
      await new Promise((resolve) => window.setTimeout(resolve, 0));
    });

    const committedSummary = await screen.findByTestId(
      rowCellTestId(
        "20000000-0000-4000-8000-000000000010",
        "timeline.activity_synopsis_text",
      ),
    );
    expect(committedSummary.textContent).toBe("—");
    expect(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000010",
          "timeline.capture_state",
        ),
      ).textContent,
    ).toBe("rough");
    expect(
      screen.getByTestId(
        timelineRowVersionTestId("20000000-0000-4000-8000-000000000010"),
      ).textContent,
    ).toBe("1");
    expect(
      screen.getByTestId(draftCellTestId("timeline.activity_synopsis_text")),
    ).toBeTruthy();
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(visibleGridRows(document.body)).toHaveLength(1);
    });
  });
});
