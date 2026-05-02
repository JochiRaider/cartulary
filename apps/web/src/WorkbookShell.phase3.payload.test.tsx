import { gridDraftRowSelector, gridShellTestId } from "@cartulary/ui-contracts";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanupTimelineWorkbookTestGlobals,
  deferred,
  extractTimelineJSONBody,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  timelineViewSchemaId,
  visibleGridRows,
} from "./timelineWorkbookTestSupport";
import {
  buildCreatePayload,
  createDraftRow,
  TimelineWorkbook,
} from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

describe("Phase 3 Timeline workbook payload coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("Phase 3 U-3-12 builds zero-field Timeline create payloads only for explicit blank-row creation", () => {
    const draftRow = createDraftRow(1);

    expect(buildCreatePayload(draftRow, "timeline-client-blank")).toBeNull();
    expect(
      buildCreatePayload(draftRow, "timeline-client-blank", {
        allowZeroFieldCreate: true,
      }),
    ).toEqual({
      client_txn_id: "timeline-client-blank",
    });

    draftRow.values.summary = "First timeline fact";
    expect(buildCreatePayload(draftRow, "timeline-client-1")).toEqual({
      client_txn_id: "timeline-client-1",
      "timeline.summary": "First timeline fact",
    });
  });

  it("Phase 3 U-3-13 creates an explicit blank Timeline row with only client_txn_id and suppresses duplicate pending submits", async () => {
    const pendingCreate = deferred<Response>();

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      }),
    );
    fetchMock.mockReturnValueOnce(pendingCreate.promise);

    render(<TimelineWorkbook incidentId="incident-1" />);

    expect((await screen.findByTestId("save-state")).textContent).toBe("Saved");

    const gridShell = screen.getByTestId(gridShellTestId("timeline"));
    const draftRows = Array.from(
      gridShell.querySelectorAll<HTMLDivElement>(gridDraftRowSelector()),
    );
    const draftGridRow = draftRows.at(-1);
    expect(draftGridRow).toBeTruthy();
    const createButton = within(draftGridRow as HTMLDivElement).getByTestId(
      "draft-row-create",
    );
    fireEvent.mouseDown(createButton);
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
            change_set_id: "change-set-zero-field",
            row: timelineRow({
              recordId: "record-zero",
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

    const committedSummary = (await screen.findByTestId(
      "row-record-zero-summary",
    )) as HTMLInputElement;
    expect(committedSummary.value).toBe("");
    expect(
      screen.getByTestId("row-record-zero-capture-state").textContent,
    ).toBe("rough");
    expect(screen.getByTestId("row-record-zero-row-version").textContent).toBe(
      "1",
    );
    expect(screen.getByTestId("draft-row-summary")).toBeTruthy();
    await waitFor(() => {
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
      expect(visibleGridRows(document.body)).toHaveLength(1);
    });
  });
});
