import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridShellTestId,
  gridSortHeaderTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { requireFetchCall } from "./fetchMockTestSupport";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  extractTimelineJSONBody,
  extractTimelinePatchBody,
  gridScalarInput,
  installTimelineWorkbookTestGlobals,
  requiredGridRow,
  setInputValueWithoutEvent,
  successEnvelope,
  timelineRow,
  timelineViewSchemaId,
  visibleGridRows,
} from "./timelineWorkbookTestSupport";
import { TimelineWorkbook } from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

const timelineContract = requireViewContract(timelineViewSchemaId);

describe("Phase 3 Timeline workbook grid coverage", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("Phase 3 U-3-GRID-01 binds Timeline grid columns from the active view_schema and commits writable cells by field_key", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 4,
            summary: "Alpha",
            captureState: "rough",
            tags: ["critical-host"],
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-grid-1",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 5,
          summary: "Bound summary",
          captureState: "enriched",
          tags: ["critical-host"],
        }),
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId("save-state");

    const grid = await screen.findByTestId(gridShellTestId("timeline"));
    const headerFields = Array.from(
      grid.querySelectorAll('[role="columnheader"] [data-grid-field-key]'),
    ).map((node) => node.getAttribute("data-grid-field-key"));
    expect(headerFields.slice(0, 2)).toEqual([
      "timeline.capture_state",
      "row_version",
    ]);
    expect(headerFields.slice(2)).toEqual(
      timelineContract.defaultVisibleFields,
    );

    for (const fieldKey of timelineContract.defaultVisibleFields) {
      expect(
        within(grid).getByText(
          timelineContract.fieldMap[fieldKey]?.label ?? fieldKey,
        ),
      ).toBeTruthy();
    }
    expect(within(grid).queryByText("Details")).toBeNull();
    expect(within(grid).queryByText("Source Text")).toBeNull();

    await waitFor(() => {
      expect(visibleGridRows(container)).toHaveLength(1);
    });
    const summaryInput = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    await changeInputValue(summaryInput, "Bound summary");
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    const requestBody = extractTimelinePatchBody(fetchMock, 1);
    expect(requestBody.changes[0]).toEqual({
      field_key: "timeline.summary",
      value: "Bound summary",
    });
    expect(JSON.stringify(requestBody)).not.toContain("Summary");

    await waitFor(() => {
      expect(visibleGridRows(container)).toHaveLength(1);
    });
    expect(
      requiredGridRow(container, 0).getAttribute("data-grid-record-id"),
    ).toBe("record-1");
  });

  it("Phase 3 U-3-GRID-02 binds saved rows by record_id and row_version instead of visible row index", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-2",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-1",
            rowVersion: 7,
            summary: "Zulu",
            captureState: "reviewed",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-grid-2",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 8,
          summary: "Zulu rebound",
          captureState: "enriched",
        }),
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId("save-state");

    await waitFor(() => {
      expect(visibleGridRows(container)).toHaveLength(2);
    });
    const firstVisibleRow = requiredGridRow(container, 0);
    const secondVisibleRow = requiredGridRow(container, 1);
    expect(firstVisibleRow.getAttribute("data-grid-record-id")).toBe(
      "record-2",
    );
    expect(secondVisibleRow.getAttribute("data-grid-record-id")).toBe(
      "record-1",
    );
    expect(within(secondVisibleRow).getByDisplayValue("Zulu")).toBeTruthy();
    expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
      "7",
    );

    const summaryInput = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    await changeInputValue(summaryInput, "Zulu rebound");
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(
      String(requireFetchCall(fetchMock, 1, "timeline patch request #1")[0]),
    ).toContain("/api/v1/records/record-1");
    expect(extractTimelinePatchBody(fetchMock, 1).base_row_version).toBe(7);
  });

  it("Phase 3 U-3-GRID-03 keeps sorted and filtered local edits bound to the original record_id, base_row_version, and field_key", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 7,
            summary: "Zulu",
            captureState: "rough",
            hasEvidence: false,
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "rough",
            hasEvidence: false,
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-2",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "rough",
            hasEvidence: false,
          }),
          timelineRow({
            recordId: "record-1",
            rowVersion: 7,
            summary: "Zulu",
            captureState: "rough",
            hasEvidence: false,
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-2",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "rough",
            hasEvidence: false,
          }),
          timelineRow({
            recordId: "record-1",
            rowVersion: 7,
            summary: "Zulu",
            captureState: "rough",
            hasEvidence: false,
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-grid-3",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 8,
          summary: "Filtered anchor",
          captureState: "enriched",
          hasEvidence: false,
        }),
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId("save-state");

    fireEvent.click(
      await screen.findByTestId(
        gridSortHeaderTestId("timeline", "timeline.summary"),
      ),
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelineJSONBody(fetchMock, 1)).toEqual({
      sort: [{ direction: "asc", field_key: "timeline.summary" }],
    });

    fireEvent.change(screen.getByTestId(gridFilterFieldTestId("timeline")), {
      target: { value: "timeline.has_evidence" },
    });
    fireEvent.change(screen.getByTestId(gridFilterValueTestId("timeline")), {
      target: { value: "false" },
    });
    fireEvent.click(screen.getByTestId(gridFilterApplyTestId("timeline")));
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(extractTimelineJSONBody(fetchMock, 2)).toEqual({
      filters: [
        {
          arg: { value: false },
          field_key: "timeline.has_evidence",
          op: "eq",
        },
      ],
      sort: [{ direction: "asc", field_key: "timeline.summary" }],
    });

    await waitFor(() => {
      expect(visibleGridRows(container)).toHaveLength(2);
    });
    const firstVisibleRow = requiredGridRow(container, 0);
    const secondVisibleRow = requiredGridRow(container, 1);
    expect(firstVisibleRow.getAttribute("data-grid-record-id")).toBe(
      "record-2",
    );
    expect(secondVisibleRow.getAttribute("data-grid-record-id")).toBe(
      "record-1",
    );

    const summaryInput = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    setInputValueWithoutEvent(summaryInput, "Filtered anchor");
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    expect(
      String(requireFetchCall(fetchMock, 3, "timeline patch request #3")[0]),
    ).toContain("/api/v1/records/record-1");
    expect(extractTimelinePatchBody(fetchMock, 3)).toMatchObject({
      base_row_version: 7,
      changes: [
        {
          field_key: "timeline.summary",
          value: "Filtered anchor",
        },
      ],
    });
  });
});
