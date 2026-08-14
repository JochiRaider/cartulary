import {
  rowInspectButtonTestId,
  saveStateTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { renderTimelineWorkbook } from "../testing/timelineWorkbookRenderTestSupport";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  errorEnvelope,
  extractTimelinePatchBody,
  findWorkbookCell,
  flushWorkbookAsync,
  installTimelineWorkbookTestGlobals,
  setInputValueWithoutEvent,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
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

describe("Timeline workbook autosave coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  function mockInitialTimelineRow(summary = "Alpha") {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary,
            captureState: "rough",
          }),
        ],
      }),
    );
  }

  function mockSummaryPatchResponse(summary: string, rowVersion = 2) {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion,
          summary,
          captureState: "enriched",
        }),
      }),
    );
  }

  function mockSourceTextPatchResponse(sourceText: string, rowVersion = 2) {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion,
          summary: "Alpha",
          sourceText,
          captureState: "enriched",
        }),
      }),
    );
  }

  async function renderSingleTimelineRow() {
    renderTimelineWorkbook();
    expect((await screen.findByTestId(saveStateTestId())).textContent).toBe(
      "Saved",
    );
    expect(screen.queryByRole("button", { name: /^save$/i })).toBeNull();
    return (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
  }

  async function openTimelineInspectorFromContext(recordId: string) {
    const summaryCell = await screen.findByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId,
        surface: "grid",
      }),
    );
    fireEvent.contextMenu(summaryCell, { clientX: 32, clientY: 48 });
    fireEvent.click(
      await screen.findByTestId(rowInspectButtonTestId(recordId)),
    );
  }

  async function expectSavedRowVersion(rowVersion: number) {
    await waitFor(() => {
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000001"),
        ).textContent,
      ).toBe(String(rowVersion));
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
  }

  function expectTimelinePatch(
    callIndex: number,
    options: { baseRowVersion: number; fieldKey: string; value: string },
  ) {
    const body = extractTimelinePatchBody(fetchMock, callIndex);
    expect(body.base_row_version).toBe(options.baseRowVersion);
    expect(body.changes[0]).toEqual({
      field_key: options.fieldKey,
      value: options.value,
    });
  }

  it("autosaves Enter without a Save button and keeps exact save-state labels", async () => {
    const draftRow = createDraftRow(1);
    draftRow.values.activitySynopsisText = "First timeline fact";

    expect(buildCreatePayload(draftRow, "timeline-client-1")).toEqual({
      client_txn_id: "timeline-client-1",
      "timeline.activity_synopsis_text": "First timeline fact",
    });

    mockInitialTimelineRow();
    mockSummaryPatchResponse("Updated via enter");
    const summaryInput = await renderSingleTimelineRow();
    await changeInputValue(summaryInput, "Updated via enter");
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expectTimelinePatch(1, {
      baseRowVersion: 1,
      fieldKey: "timeline.activity_synopsis_text",
      value: "Updated via enter",
    });
    await expectSavedRowVersion(2);
  });

  it("autosaves Tab without a Save button and keeps exact save-state labels", async () => {
    mockInitialTimelineRow();
    mockSummaryPatchResponse("Updated via tab");
    const tabInput = await renderSingleTimelineRow();
    await changeInputValue(tabInput, "Updated via tab");
    fireEvent.keyDown(tabInput, { key: "Tab" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expectTimelinePatch(1, {
      baseRowVersion: 1,
      fieldKey: "timeline.activity_synopsis_text",
      value: "Updated via tab",
    });
    await expectSavedRowVersion(2);
  });

  it("autosaves blur without a Save button and keeps exact save-state labels", async () => {
    const pendingBlurPatch = deferred<Response>();
    mockInitialTimelineRow();
    fetchMock.mockReturnValueOnce(pendingBlurPatch.promise);
    const blurInput = await renderSingleTimelineRow();
    await changeInputValue(blurInput, "Updated via blur");
    fireEvent.blur(blurInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expectTimelinePatch(1, {
      baseRowVersion: 1,
      fieldKey: "timeline.activity_synopsis_text",
      value: "Updated via blur",
    });
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Syncing");
    });

    pendingBlurPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "Updated via blur",
          captureState: "enriched",
        }),
      }),
    );
    await expectSavedRowVersion(2);
  });

  it("autosaves paste completion without a Save button and keeps exact save-state labels", async () => {
    mockInitialTimelineRow();
    mockSourceTextPatchResponse("Pasted transcript");
    await renderSingleTimelineRow();
    await openTimelineInspectorFromContext(
      "20000000-0000-4000-8000-000000000001",
    );
    const sourceText = (await screen.findByLabelText(
      "RAW Activity",
    )) as HTMLTextAreaElement;
    await changeInputValue(sourceText, "Pasted transcript");
    fireEvent.paste(sourceText);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expectTimelinePatch(1, {
      baseRowVersion: 1,
      fieldKey: "timeline.raw_activity_text",
      value: "Pasted transcript",
    });
    await expectSavedRowVersion(2);
  });

  it("reports Conflict after autosave failure and preserves local editor value", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      errorEnvelope("same_field_conflict", 409, {
        field_key: "timeline.activity_synopsis_text",
        base_row_version: 1,
        current_row_version: 2,
        base_value: "Alpha",
        server_value: "Server value",
        client_value: "Conflict value",
      }),
    );

    renderTimelineWorkbook();

    const conflictInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(conflictInput);
    await changeInputValue(conflictInput, "Conflict value");
    fireEvent.blur(conflictInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(conflictInput.value).toBe("Conflict value");
    });
  });

  it.each([
    "Enter",
    "Tab",
  ] as const)("commits %s from the current editor value when row state is stale", async (key) => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: `Stale-proof ${key}`,
          captureState: "enriched",
        }),
      }),
    );

    renderTimelineWorkbook();

    await screen.findByTestId(saveStateTestId());
    const summaryInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;

    setInputValueWithoutEvent(summaryInput, `Stale-proof ${key}`);
    fireEvent.keyDown(summaryInput, { key });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain(
      "/api/v1/records/20000000-0000-4000-8000-000000000001",
    );
    expect(extractTimelinePatchBody(fetchMock, 1)).toMatchObject({
      base_row_version: 1,
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: `Stale-proof ${key}`,
        },
      ],
    });
  });

  it("commits paste completion from the current editor value when row state is stale", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "Pasted stale-proof summary",
          captureState: "enriched",
        }),
      }),
    );

    renderTimelineWorkbook();

    await screen.findByTestId(saveStateTestId());
    const summaryInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;

    setInputValueWithoutEvent(summaryInput, "Pasted stale-proof summary");
    fireEvent.paste(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1)).toMatchObject({
      base_row_version: 1,
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Pasted stale-proof summary",
        },
      ],
    });
  });

  it("suppresses duplicate pending scalar saves that only differ by client_txn_id", async () => {
    const pendingPatch = deferred<Response>();

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(pendingPatch.promise);

    renderTimelineWorkbook();

    await screen.findByTestId(saveStateTestId());

    const summaryInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    await changeInputValue(summaryInput, "Pending deduplicated summary");
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1)).toMatchObject({
      base_row_version: 1,
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Pending deduplicated summary",
        },
      ],
    });

    fireEvent.blur(summaryInput);
    await flushWorkbookAsync();
    expect(fetchMock).toHaveBeenCalledTimes(2);

    pendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "Pending deduplicated summary",
          captureState: "enriched",
        }),
      }),
    );

    await waitFor(() => {
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000001"),
        ).textContent,
      ).toBe("2");
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
    await flushWorkbookAsync();
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("queues a follow-up scalar save when the logical payload changes while a prior save is pending", async () => {
    const firstPendingPatch = deferred<Response>();

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Alpha",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(firstPendingPatch.promise);
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 3,
          summary: "Second pending summary",
          captureState: "enriched",
        }),
      }),
    );

    renderTimelineWorkbook();

    await screen.findByTestId(saveStateTestId());

    const summaryInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    await changeInputValue(summaryInput, "First pending summary");
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1)).toMatchObject({
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "First pending summary",
        },
      ],
    });

    await changeInputValue(summaryInput, "Second pending summary");
    fireEvent.blur(summaryInput);

    await flushWorkbookAsync();
    expect(fetchMock).toHaveBeenCalledTimes(2);

    firstPendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "First pending summary",
          captureState: "enriched",
        }),
      }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(extractTimelinePatchBody(fetchMock, 2)).toMatchObject({
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Second pending summary",
        },
      ],
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000001"),
        ).textContent,
      ).toBe("3");
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
  });
});
