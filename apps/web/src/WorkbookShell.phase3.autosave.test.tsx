import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  deferred,
  errorEnvelope,
  extractTimelinePatchBody,
  installTimelineWorkbookTestGlobals,
  setInputValueWithoutEvent,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  timelineViewSchemaId,
} from "./timelineWorkbookTestSupport";
import {
  buildCreatePayload,
  createDraftRow,
  ensureDraftRow,
  TimelineWorkbook,
} from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

describe("Phase 3 Timeline workbook autosave coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
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
          collectionValues: {
            hostRefs: [],
            identityRefs: [],
            tags: [],
          },
          collectionDrafts: {
            hostRefs: "",
            identityRefs: "",
            tags: "",
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
          captureState: "enriched",
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-5",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 4,
          summary: "Updated via blur",
          sourceText: "Pasted transcript",
          captureState: "enriched",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);

    expect((await screen.findByTestId("save-state")).textContent).toBe("Saved");
    expect(screen.queryByRole("button", { name: /^save$/i })).toBeNull();

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    await changeInputValue(summaryInput, "Updated via enter");
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1).base_row_version).toBe(1);
    expect(extractTimelinePatchBody(fetchMock, 1).changes[0]).toEqual({
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
    await changeInputValue(tabInput, "Updated via tab");
    fireEvent.keyDown(tabInput, { key: "Tab" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(extractTimelinePatchBody(fetchMock, 2).base_row_version).toBe(2);
    expect(extractTimelinePatchBody(fetchMock, 2).changes[0]).toEqual({
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
    await changeInputValue(blurInput, "Updated via blur");
    fireEvent.blur(blurInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    expect(extractTimelinePatchBody(fetchMock, 3).base_row_version).toBe(3);
    expect(extractTimelinePatchBody(fetchMock, 3).changes[0]).toEqual({
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

    fireEvent.click(await screen.findByTestId("row-record-1-inspect"));
    const sourceText = (await screen.findByLabelText(
      "Source Text",
    )) as HTMLTextAreaElement;
    await changeInputValue(sourceText, "Pasted transcript");
    fireEvent.paste(sourceText);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(5);
    });
    expect(extractTimelinePatchBody(fetchMock, 4).base_row_version).toBe(4);
    expect(extractTimelinePatchBody(fetchMock, 4).changes[0]).toEqual({
      field_key: "timeline.source_text",
      value: "Pasted transcript",
    });
    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "4",
      );
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });

    cleanup();
    fetchMock.mockReset();
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
      errorEnvelope("same_field_conflict", 409, {
        field_key: "timeline.summary",
        base_row_version: 1,
        current_row_version: 2,
        base_value: "Alpha",
        server_value: "Server value",
        client_value: "Conflict value",
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);

    const conflictInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.focus(conflictInput);
    await changeInputValue(conflictInput, "Conflict value");
    fireEvent.blur(conflictInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
      expect(conflictInput.value).toBe("Conflict value");
    });
  });

  it.each([
    "Enter",
    "Tab",
  ] as const)("Phase 3 U-3-05 commits %s from the current editor value when row state is stale", async (key) => {
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
        change_set_id: `change-set-stale-${key.toLowerCase()}`,
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: `Stale-proof ${key}`,
          captureState: "enriched",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId("save-state");
    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;

    setInputValueWithoutEvent(summaryInput, `Stale-proof ${key}`);
    fireEvent.keyDown(summaryInput, { key });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain(
      "/api/v1/records/record-1",
    );
    expect(extractTimelinePatchBody(fetchMock, 1)).toMatchObject({
      base_row_version: 1,
      changes: [
        {
          field_key: "timeline.summary",
          value: `Stale-proof ${key}`,
        },
      ],
    });
  });

  it("Phase 3 U-3-05 commits paste completion from the current editor value when row state is stale", async () => {
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
        change_set_id: "change-set-stale-paste",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Pasted stale-proof summary",
          captureState: "enriched",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId("save-state");
    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
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
          field_key: "timeline.summary",
          value: "Pasted stale-proof summary",
        },
      ],
    });
  });

  it("Phase 3 U-3-05 suppresses duplicate pending scalar saves that only differ by client_txn_id", async () => {
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

    await screen.findByTestId("save-state");

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
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
          field_key: "timeline.summary",
          value: "Pending deduplicated summary",
        },
      ],
    });

    fireEvent.blur(summaryInput);
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    expect(fetchMock).toHaveBeenCalledTimes(2);

    pendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-dedup-1",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Pending deduplicated summary",
          captureState: "enriched",
        }),
      }),
    );

    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "2",
      );
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("Phase 3 U-3-05 queues a follow-up scalar save when the logical payload changes while a prior save is pending", async () => {
    const firstPendingPatch = deferred<Response>();

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
    fetchMock.mockReturnValueOnce(firstPendingPatch.promise);
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-dedup-2",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 3,
          summary: "Second pending summary",
          captureState: "enriched",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId("save-state");

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    await changeInputValue(summaryInput, "First pending summary");
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1)).toMatchObject({
      changes: [
        {
          field_key: "timeline.summary",
          value: "First pending summary",
        },
      ],
    });

    await changeInputValue(summaryInput, "Second pending summary");
    fireEvent.blur(summaryInput);

    await new Promise((resolve) => window.setTimeout(resolve, 0));
    expect(fetchMock).toHaveBeenCalledTimes(2);

    firstPendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-dedup-1",
        row: timelineRow({
          recordId: "record-1",
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
          field_key: "timeline.summary",
          value: "Second pending summary",
        },
      ],
    });
    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "3",
      );
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });
  });
});
