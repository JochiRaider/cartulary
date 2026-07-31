import {
  rowCellTestId,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  saveStateTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  extractTimelineRecordActionBody,
  extractTimelineRecordPatchBody,
  flushWorkbookAsync,
  installTimelineWorkbookTestGlobals,
  routeTimelineWorkbookFetchMock,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRecordActionCalls,
  timelineRow,
  waitForTimelineRecordActionCalls,
  waitForTimelineRecordPatchCalls,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

describe("Timeline workbook action sequencing", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  async function openTimelineInspectorFromContext(recordId: string) {
    const summaryCell = await screen.findByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    );
    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
    fireEvent.click(
      await screen.findByTestId(rowInspectButtonTestId(recordId)),
    );
  }

  async function openTimelineRowContextMenu(recordId: string) {
    const summaryCell = await screen.findByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    );
    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
  }

  function queueTwoRowInitialState(options: {
    recordOneCaptureState: string;
    recordOneDetails?: string;
    recordOneRowVersion: number;
  }) {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: options.recordOneRowVersion,
            summary: "Alpha",
            details: options.recordOneDetails ?? "",
            captureState: options.recordOneCaptureState,
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
            rowVersion: 1,
            summary: "Replacement",
            captureState: "rough",
          }),
        ],
      }),
    );
    return routedFetch;
  }

  it("sends Timeline actions with the current row version after earlier workbook mutations", async () => {
    const routedFetch = queueTwoRowInitialState({
      recordOneCaptureState: "rough",
      recordOneDetails: "Original details",
      recordOneRowVersion: 1,
    });
    routedFetch.mockRecordActionOnce(
      successEnvelope({
        record_id: "20000000-0000-4000-8000-000000000101",
        incident_id: "10000000-0000-4000-8000-000000000001",
        row_version: 2,
        capture_state: "reviewed",
        change_set_id: "30000000-0000-4000-8000-000000000101",
        reason: "Reviewed from workbook",
        replacement_record_id: null,
      }),
    );
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: 2,
            summary: "Alpha",
            details: "Original details",
            captureState: "reviewed",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
            rowVersion: 1,
            summary: "Replacement",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000102",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000101",
          rowVersion: 3,
          summary: "Alpha",
          details: "Material edit after review",
          captureState: "enriched",
        }),
      }),
    );
    routedFetch.mockRecordActionOnce(
      successEnvelope({
        record_id: "20000000-0000-4000-8000-000000000101",
        incident_id: "10000000-0000-4000-8000-000000000001",
        row_version: 4,
        capture_state: "superseded",
        change_set_id: "30000000-0000-4000-8000-000000000103",
        reason: "Superseded from workbook",
        replacement_record_id: "20000000-0000-4000-8000-000000000102",
      }),
    );
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: 4,
            summary: "Alpha",
            details: "Material edit after review",
            captureState: "superseded",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
            rowVersion: 1,
            summary: "Replacement",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="reviewer"
      />,
    );

    await screen.findByTestId(saveStateTestId());
    await openTimelineRowContextMenu("20000000-0000-4000-8000-000000000101");
    fireEvent.click(
      await screen.findByTestId(
        timelineRowMarkReviewedButtonTestId(
          "20000000-0000-4000-8000-000000000101",
        ),
      ),
    );

    await waitForTimelineRecordActionCalls(fetchMock, "mark-reviewed", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "mark-reviewed"),
    ).toMatchObject({
      base_row_version: 1,
      reason: "Reviewed from workbook",
    });

    await waitFor(() => {
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000101"),
        ).textContent,
      ).toBe("2");
    });

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000101",
    );
    const detailsInput = (await screen.findByTestId(
      rowInspectorFieldTestId(
        "20000000-0000-4000-8000-000000000101",
        "timeline.raw_activity_text",
      ),
    )) as HTMLTextAreaElement;
    await changeInputValue(detailsInput, "Material edit after review");
    fireEvent.blur(detailsInput);

    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    expect(extractTimelineRecordPatchBody(fetchMock, 0)).toMatchObject({
      base_row_version: 2,
      changes: [
        {
          field_key: "timeline.raw_activity_text",
          value: "Material edit after review",
        },
      ],
    });

    await waitFor(() => {
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000101"),
        ).textContent,
      ).toBe("3");
    });

    await openTimelineRowContextMenu("20000000-0000-4000-8000-000000000101");
    fireEvent.change(
      await screen.findByTestId(
        timelineRowReplacementInputTestId(
          "20000000-0000-4000-8000-000000000101",
        ),
      ),
      {
        target: { value: "20000000-0000-4000-8000-000000000102" },
      },
    );
    fireEvent.click(
      await screen.findByTestId(
        timelineRowSupersedeButtonTestId(
          "20000000-0000-4000-8000-000000000101",
        ),
      ),
    );

    await waitForTimelineRecordActionCalls(fetchMock, "supersede", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "supersede"),
    ).toMatchObject({
      base_row_version: 3,
      reason: "Superseded from workbook",
      replacement_record_id: "20000000-0000-4000-8000-000000000102",
    });
  });

  it("waits for a pending material autosave before supersede dispatch", async () => {
    const pendingDetailsPatch = deferred<Response>();
    const routedFetch = queueTwoRowInitialState({
      recordOneCaptureState: "reviewed",
      recordOneDetails: "Original details",
      recordOneRowVersion: 2,
    });
    routedFetch.mockRecordPatchOnce(pendingDetailsPatch.promise);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: 2,
            summary: "Alpha",
            details: "Original details",
            captureState: "reviewed",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
            rowVersion: 1,
            summary: "Replacement",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordActionOnce(
      successEnvelope({
        record_id: "20000000-0000-4000-8000-000000000101",
        incident_id: "10000000-0000-4000-8000-000000000001",
        row_version: 4,
        capture_state: "superseded",
        change_set_id: "30000000-0000-4000-8000-000000000103",
        reason: "Superseded from workbook",
        replacement_record_id: "20000000-0000-4000-8000-000000000102",
      }),
    );
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: 4,
            summary: "Alpha",
            details: "Material edit after review",
            captureState: "superseded",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
            rowVersion: 1,
            summary: "Replacement",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="reviewer"
      />,
    );

    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000101",
    );
    const detailsInput = (await screen.findByTestId(
      rowInspectorFieldTestId(
        "20000000-0000-4000-8000-000000000101",
        "timeline.raw_activity_text",
      ),
    )) as HTMLTextAreaElement;
    await changeInputValue(detailsInput, "Material edit after review");
    fireEvent.blur(detailsInput);

    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    expect(extractTimelineRecordPatchBody(fetchMock, 0)).toMatchObject({
      base_row_version: 2,
      changes: [
        {
          field_key: "timeline.raw_activity_text",
          value: "Material edit after review",
        },
      ],
    });

    await fetch(
      `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${timelineViewSchemaId}/query`,
      { method: "POST", body: JSON.stringify({}) },
    );

    await openTimelineRowContextMenu("20000000-0000-4000-8000-000000000101");
    fireEvent.change(
      await screen.findByTestId(
        timelineRowReplacementInputTestId(
          "20000000-0000-4000-8000-000000000101",
        ),
      ),
      {
        target: { value: "20000000-0000-4000-8000-000000000102" },
      },
    );
    fireEvent.click(
      await screen.findByTestId(
        timelineRowSupersedeButtonTestId(
          "20000000-0000-4000-8000-000000000101",
        ),
      ),
    );
    await flushWorkbookAsync();
    expect(timelineRecordActionCalls(fetchMock, "supersede")).toHaveLength(0);

    pendingDetailsPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000102",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000101",
          rowVersion: 3,
          summary: "Alpha",
          details: "Material edit after review",
          captureState: "enriched",
        }),
      }),
    );

    await waitForTimelineRecordActionCalls(fetchMock, "supersede", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "supersede"),
    ).toMatchObject({
      base_row_version: 3,
      reason: "Superseded from workbook",
      replacement_record_id: "20000000-0000-4000-8000-000000000102",
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(
          rowCellTestId(
            "20000000-0000-4000-8000-000000000101",
            "timeline.capture_state",
          ),
        ).textContent,
      ).toBe("superseded");
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000101"),
        ).textContent,
      ).toBe("4");
    });
  });

  it("keeps action result row versions ahead of stale projection reloads", async () => {
    const routedFetch = queueTwoRowInitialState({
      recordOneCaptureState: "rough",
      recordOneRowVersion: 1,
    });
    routedFetch.mockRecordActionOnce(
      successEnvelope({
        record_id: "20000000-0000-4000-8000-000000000101",
        incident_id: "10000000-0000-4000-8000-000000000001",
        row_version: 2,
        capture_state: "reviewed",
        change_set_id: "30000000-0000-4000-8000-000000000101",
        reason: "Reviewed from workbook",
        replacement_record_id: null,
      }),
    );
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
            rowVersion: 1,
            summary: "Replacement",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: 2,
            summary: "Alpha",
            captureState: "reviewed",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
            rowVersion: 1,
            summary: "Replacement",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordActionOnce(
      successEnvelope({
        record_id: "20000000-0000-4000-8000-000000000101",
        incident_id: "10000000-0000-4000-8000-000000000001",
        row_version: 3,
        capture_state: "superseded",
        change_set_id: "30000000-0000-4000-8000-000000000103",
        reason: "Superseded from workbook",
        replacement_record_id: "20000000-0000-4000-8000-000000000102",
      }),
    );
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: 3,
            summary: "Alpha",
            captureState: "superseded",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
            rowVersion: 1,
            summary: "Replacement",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        currentIncidentRole="reviewer"
      />,
    );

    await openTimelineRowContextMenu("20000000-0000-4000-8000-000000000101");
    fireEvent.click(
      await screen.findByTestId(
        timelineRowMarkReviewedButtonTestId(
          "20000000-0000-4000-8000-000000000101",
        ),
      ),
    );

    await waitForTimelineRecordActionCalls(fetchMock, "mark-reviewed", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "mark-reviewed"),
    ).toMatchObject({
      base_row_version: 1,
    });

    await waitFor(() => {
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000101"),
        ).textContent,
      ).toBe("2");
    });

    await openTimelineRowContextMenu("20000000-0000-4000-8000-000000000101");
    fireEvent.change(
      await screen.findByTestId(
        timelineRowReplacementInputTestId(
          "20000000-0000-4000-8000-000000000101",
        ),
      ),
      {
        target: { value: "20000000-0000-4000-8000-000000000102" },
      },
    );
    fireEvent.click(
      await screen.findByTestId(
        timelineRowSupersedeButtonTestId(
          "20000000-0000-4000-8000-000000000101",
        ),
      ),
    );

    await waitForTimelineRecordActionCalls(fetchMock, "supersede", 1);
    expect(
      extractTimelineRecordActionBody(fetchMock, "supersede"),
    ).toMatchObject({
      base_row_version: 2,
      replacement_record_id: "20000000-0000-4000-8000-000000000102",
    });
  });
});
