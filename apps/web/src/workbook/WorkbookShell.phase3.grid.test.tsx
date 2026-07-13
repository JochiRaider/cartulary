import { gridAdapterVendor } from "@cartulary/grid-adapter";
import {
  draftCellTestId,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  saveStateTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { deferred, requireFetchCall } from "../testing/fetchMockTestSupport";
import {
  buildRecordChangedPayload,
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  emitRecordChanged,
  errorEnvelope,
  extractTimelineJSONBody,
  extractTimelinePatchBody,
  gridScalarInput,
  installTimelineWorkbookTestGlobals,
  latestTimelineWebSocket,
  requiredGridRow,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  visibleGridRows,
  waitForTimelineWorkbookReady,
  waitForVisibleGridRowRecordIds,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import { TimelineWorkbook } from "./timeline/components/TimelineWorkbook";
import { decideWorkbookRecordFreshness } from "./timeline/models/workbookTimelineModel";

const timelineContract = requireViewContract(timelineViewSchemaId);

describe("Phase 3 Timeline workbook grid coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("classifies committed Timeline row freshness by record_id and row_version", () => {
    expect(
      decideWorkbookRecordFreshness({ recordId: "record-1", rowVersion: 1 }, 2),
    ).toEqual({ comparable: true, stale: true });
    expect(
      decideWorkbookRecordFreshness({ recordId: "record-1", rowVersion: 2 }, 2),
    ).toEqual({ comparable: true, stale: false });
    expect(
      decideWorkbookRecordFreshness({ recordId: "record-1", rowVersion: 3 }, 2),
    ).toEqual({ comparable: true, stale: false });
    expect(
      decideWorkbookRecordFreshness({ recordId: null, rowVersion: null }, 2),
    ).toEqual({ comparable: false, stale: false });
  });

  it("Phase 3 U-3-GRID-01 binds Timeline grid columns from the active view_schema and commits writable cells by field_key", async () => {
    expect(gridAdapterVendor).toBe("react-data-grid");

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

    await screen.findByTestId(saveStateTestId());
    const grid = await waitForTimelineWorkbookReady(container, 1);
    const headerFields = Array.from(
      grid.querySelectorAll('[role="columnheader"] [data-grid-field-key]'),
    ).map((node) => node.getAttribute("data-grid-field-key"));
    expect(headerFields).toEqual(timelineContract.defaultVisibleFields);

    for (const fieldKey of timelineContract.defaultVisibleFields) {
      expect(
        within(grid).getByText(
          timelineContract.fieldMap[fieldKey]?.label ?? fieldKey,
        ),
      ).toBeTruthy();
    }
    expect(within(grid).queryByText("Details")).toBeNull();
    expect(within(grid).queryByText("Source Text")).toBeNull();

    const summaryInput = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    await changeInputValue(summaryInput, "Bound summary");
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    const requestBody = extractTimelinePatchBody(fetchMock, 1);
    expect(requestBody.changes[0]).toEqual({
      field_key: "timeline.activity_synopsis_text",
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
    expect(gridAdapterVendor).toBe("react-data-grid");

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

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);
    const firstVisibleRow = requiredGridRow(container, 0);
    const secondVisibleRow = requiredGridRow(container, 1);
    expect(firstVisibleRow.getAttribute("data-grid-record-id")).toBe(
      "record-2",
    );
    expect(secondVisibleRow.getAttribute("data-grid-record-id")).toBe(
      "record-1",
    );
    expect(within(secondVisibleRow).getByDisplayValue("Zulu")).toBeTruthy();
    expect(
      screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
    ).toBe("7");

    const summaryInput = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
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

  it("preserves draft row edits across projection refresh before create commit", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-draft-create",
        row: timelineRow({
          recordId: "record-created",
          rowVersion: 1,
          summary: "First browser fact",
          captureState: "rough",
        }),
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 0);

    const initialDraftSummary = screen.getByTestId(
      draftCellTestId("timeline.activity_synopsis_text"),
    ) as HTMLInputElement;
    await changeInputValue(initialDraftSummary, "First browser fact");

    emitRecordChanged(
      latestTimelineWebSocket(),
      buildRecordChangedPayload({
        recordId: "external-record",
        rowVersion: 1,
        clientTxnId: "external-client-txn",
        changedFieldKeys: ["timeline.activity_synopsis_text"],
      }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    const refreshedDraftSummary = screen.getByTestId(
      draftCellTestId("timeline.activity_synopsis_text"),
    ) as HTMLInputElement;
    expect(refreshedDraftSummary.value).toBe("First browser fact");

    fireEvent.keyDown(refreshedDraftSummary, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(extractTimelineJSONBody(fetchMock, 2)).toMatchObject({
      "timeline.activity_synopsis_text": "First browser fact",
    });

    await waitFor(() => {
      expect(visibleGridRows(container)).toHaveLength(1);
    });
    expect(
      (
        screen.getByTestId(
          rowCellTestId("record-created", "timeline.activity_synopsis_text"),
        ) as HTMLInputElement
      ).value,
    ).toBe("First browser fact");
    expect(
      (
        screen.getByTestId(
          draftCellTestId("timeline.activity_synopsis_text"),
        ) as HTMLInputElement
      ).value,
    ).toBe("");
  });

  it("preserves an in-flight draft create across unrelated projection refresh", async () => {
    const pendingCreate = deferred<Response>();
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      }),
    );
    fetchMock.mockReturnValueOnce(pendingCreate.promise);
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 0);

    const draftSummary = screen.getByTestId(
      draftCellTestId("timeline.activity_synopsis_text"),
    ) as HTMLInputElement;
    await changeInputValue(draftSummary, "Pending browser fact");
    fireEvent.keyDown(draftSummary, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelineJSONBody(fetchMock, 1)).toMatchObject({
      "timeline.activity_synopsis_text": "Pending browser fact",
    });

    emitRecordChanged(
      latestTimelineWebSocket(),
      buildRecordChangedPayload({
        recordId: "external-record",
        rowVersion: 1,
        clientTxnId: "external-client-txn",
        changedFieldKeys: ["timeline.raw_activity_text"],
      }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(
      (
        screen.getByTestId(
          draftCellTestId("timeline.activity_synopsis_text"),
        ) as HTMLInputElement
      ).value,
    ).toBe("Pending browser fact");

    pendingCreate.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-pending-draft-create",
        row: timelineRow({
          recordId: "record-pending-created",
          rowVersion: 1,
          summary: "Pending browser fact",
          captureState: "rough",
        }),
      }),
    );

    await waitFor(() => {
      expect(visibleGridRows(container)).toHaveLength(1);
    });
    expect(
      (
        screen.getByTestId(
          rowCellTestId(
            "record-pending-created",
            "timeline.activity_synopsis_text",
          ),
        ) as HTMLInputElement
      ).value,
    ).toBe("Pending browser fact");
    expect(
      (
        screen.getByTestId(
          draftCellTestId("timeline.activity_synopsis_text"),
        ) as HTMLInputElement
      ).value,
    ).toBe("");
  });

  it("keeps Timeline controls mounted while a sorted query refresh is pending", async () => {
    const pendingSortedRows = deferred<Response>();
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
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(pendingSortedRows.promise);

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    fireEvent.click(
      await screen.findByTestId(
        gridSortHeaderTestId(
          timelineViewSchemaId,
          "timeline.activity_synopsis_text",
        ),
      ),
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    expect(
      screen.getByTestId(gridFilterFieldTestId(timelineViewSchemaId)),
    ).toBeTruthy();
    expect(
      screen.getByTestId(gridFilterValueTestId(timelineViewSchemaId)),
    ).toBeTruthy();
    expect(visibleGridRows(container)).toHaveLength(2);

    pendingSortedRows.resolve(
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
            captureState: "rough",
          }),
        ],
      }),
    );
    await waitForVisibleGridRowRecordIds(container, ["record-2", "record-1"]);
  });

  it("surfaces a sorted query refresh failure without unmounting Timeline controls", async () => {
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
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(errorEnvelope("projection_failed", 500));

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    fireEvent.click(
      await screen.findByTestId(
        gridSortHeaderTestId(
          timelineViewSchemaId,
          "timeline.activity_synopsis_text",
        ),
      ),
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    expect(await screen.findByTestId("timeline-refresh-error")).toBeTruthy();
    expect(
      screen.getByTestId(gridFilterFieldTestId(timelineViewSchemaId)),
    ).toBeTruthy();
    expect(visibleGridRows(container)).toHaveLength(2);
  });

  it("keeps accepted committed row state across stale invalidation queries and replay responses", async () => {
    const staleInvalidationQuery = deferred<Response>();
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha summary",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(staleInvalidationQuery.promise);
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-fresh-patch",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Zulu anchored",
          captureState: "enriched",
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 2,
            summary: "Zulu anchored",
            captureState: "enriched",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-stale-replay",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 1,
          summary: "Alpha summary",
          captureState: "rough",
        }),
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 1);

    emitRecordChanged(
      latestTimelineWebSocket(),
      buildRecordChangedPayload({
        recordId: "external-record",
        rowVersion: 1,
        clientTxnId: "external-stale-query",
        changedFieldKeys: ["timeline.raw_activity_text"],
      }),
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    const summaryInput = screen.getByTestId(
      rowCellTestId("record-1", "timeline.activity_synopsis_text"),
    ) as HTMLInputElement;
    await changeInputValue(summaryInput, "Zulu anchored");
    fireEvent.blur(summaryInput);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("2");
      expect(
        (
          screen.getByTestId(
            rowCellTestId("record-1", "timeline.activity_synopsis_text"),
          ) as HTMLInputElement
        ).value,
      ).toBe("Zulu anchored");
    });

    staleInvalidationQuery.resolve(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha summary",
            captureState: "rough",
          }),
        ],
      }),
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("2");
      expect(
        (
          screen.getByTestId(
            rowCellTestId("record-1", "timeline.activity_synopsis_text"),
          ) as HTMLInputElement
        ).value,
      ).toBe("Zulu anchored");
    });

    const replayInput = screen.getByTestId(
      rowCellTestId("record-1", "timeline.activity_synopsis_text"),
    ) as HTMLInputElement;
    await changeInputValue(replayInput, "Replay should not regress");
    fireEvent.blur(replayInput);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(5);
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("2");
      expect(
        (
          screen.getByTestId(
            rowCellTestId("record-1", "timeline.activity_synopsis_text"),
          ) as HTMLInputElement
        ).value,
      ).toBe("Zulu anchored");
    });
  });

  it("ignores stale query errors after a newer committed row response", async () => {
    const staleInvalidationQuery = deferred<Response>();
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha summary",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(staleInvalidationQuery.promise);
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-fresh-patch",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Zulu anchored",
          captureState: "enriched",
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 2,
            summary: "Zulu anchored",
            captureState: "enriched",
          }),
        ],
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 1);

    emitRecordChanged(
      latestTimelineWebSocket(),
      buildRecordChangedPayload({
        recordId: "external-record",
        rowVersion: 1,
        clientTxnId: "external-stale-error",
        changedFieldKeys: ["timeline.raw_activity_text"],
      }),
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    const summaryInput = screen.getByTestId(
      rowCellTestId("record-1", "timeline.activity_synopsis_text"),
    ) as HTMLInputElement;
    await changeInputValue(summaryInput, "Zulu anchored");
    fireEvent.blur(summaryInput);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });

    staleInvalidationQuery.resolve(errorEnvelope("projection_failed", 500));
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    await waitFor(() => {
      expect(screen.queryByTestId("timeline-refresh-error")).toBeNull();
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("2");
      expect(
        (
          screen.getByTestId(
            rowCellTestId("record-1", "timeline.activity_synopsis_text"),
          ) as HTMLInputElement
        ).value,
      ).toBe("Zulu anchored");
    });
  });

  it("ignores stale sparse live row patches after a newer committed row is accepted", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 2,
            summary: "Zulu anchored",
            captureState: "enriched",
          }),
        ],
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 1);

    emitRecordChanged(latestTimelineWebSocket(), {
      record_id: "record-1",
      row_version: 1,
      change_set_id: "change-set-stale-live",
      client_txn_id: "external-stale-live",
      actor_user_id: "user-2",
      changed_field_keys: ["timeline.activity_synopsis_text"],
      affected_views: [
        {
          view_schema_id: timelineViewSchemaId,
          change_kind: "patch",
          patch_cells: {
            record_id: "record-1",
            row_version: 1,
            cells: {
              "timeline.activity_synopsis_text": { value: "Alpha summary" },
            },
          },
        },
      ],
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(1);
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("2");
      expect(
        (
          screen.getByTestId(
            rowCellTestId("record-1", "timeline.activity_synopsis_text"),
          ) as HTMLInputElement
        ).value,
      ).toBe("Zulu anchored");
    });
  });

  it("Phase 3 U-3-GRID-03 keeps sorted and filtered local edits bound to the original record_id, base_row_version, and field_key", async () => {
    expect(gridAdapterVendor).toBe("react-data-grid");

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

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    fireEvent.click(
      await screen.findByTestId(
        gridSortHeaderTestId(
          timelineViewSchemaId,
          "timeline.activity_synopsis_text",
        ),
      ),
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelineJSONBody(fetchMock, 1)).toEqual({
      sort: [
        { direction: "asc", field_key: "timeline.activity_synopsis_text" },
      ],
    });
    await waitForVisibleGridRowRecordIds(container, ["record-2", "record-1"]);

    fireEvent.change(
      screen.getByTestId(gridFilterFieldTestId(timelineViewSchemaId)),
      {
        target: { value: "timeline.has_evidence" },
      },
    );
    fireEvent.change(
      screen.getByTestId(gridFilterValueTestId(timelineViewSchemaId)),
      {
        target: { value: "false" },
      },
    );
    fireEvent.click(
      screen.getByTestId(gridFilterApplyTestId(timelineViewSchemaId)),
    );
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
      sort: [
        { direction: "asc", field_key: "timeline.activity_synopsis_text" },
      ],
    });

    await waitForVisibleGridRowRecordIds(container, ["record-2", "record-1"]);

    fireEvent.change(
      screen.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
      {
        target: { value: "timeline.capture_state" },
      },
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    expect(extractTimelineJSONBody(fetchMock, 3)).toEqual({
      filters: [
        {
          arg: { value: false },
          field_key: "timeline.has_evidence",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [
        { direction: "asc", field_key: "timeline.capture_state" },
        { direction: "asc", field_key: "timeline.activity_synopsis_text" },
      ],
    });
    expect(
      screen.getByTestId(
        gridGroupRowTestId(
          timelineViewSchemaId,
          "timeline.capture_state",
          "rough",
        ),
      ),
    ).toBeTruthy();
    await waitForVisibleGridRowRecordIds(container, ["record-2", "record-1"]);

    const summaryInput = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    await changeInputValue(summaryInput, "Filtered anchor");
    expect(fetchMock).toHaveBeenCalledTimes(4);
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(5);
    });
    expect(
      String(requireFetchCall(fetchMock, 4, "timeline patch request #4")[0]),
    ).toContain("/api/v1/records/record-1");
    expect(extractTimelinePatchBody(fetchMock, 4)).toMatchObject({
      base_row_version: 7,
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Filtered anchor",
        },
      ],
    });
  });
});
