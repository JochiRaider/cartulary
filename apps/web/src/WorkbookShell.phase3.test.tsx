import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridShellTestId,
  gridSortHeaderTestId,
} from "@cartulary/test-utils";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { requireFetchCall, requireJSONBodyAt } from "./fetchMockTestSupport";
import {
  deferred,
  errorEnvelope,
  requiredGridRow,
  successEnvelope,
  timelineRow,
  timelineViewSchemaId,
  visibleGridRows,
} from "./timelineWorkbookTestSupport";
import {
  buildCreatePayload,
  createDraftRow,
  ensureDraftRow,
  TimelineWorkbook,
} from "./WorkbookShell";

const timelineContract = requireViewContract(timelineViewSchemaId);

describe("Phase 3 Timeline workbook authoritative coverage", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal(
      "WebSocket",
      class {
        onmessage: ((event: MessageEvent) => void) | null = null;

        close() {}
      } as unknown as typeof WebSocket,
    );
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

    fireEvent.click(await screen.findByTestId("row-record-1-inspect"));
    const sourceText = (await screen.findByLabelText(
      "Source Text",
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
    fireEvent.change(conflictInput, { target: { value: "Conflict value" } });
    await waitFor(() => {
      expect(conflictInput.value).toBe("Conflict value");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.blur(conflictInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
      expect(conflictInput.value).toBe("Conflict value");
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
    fireEvent.change(summaryInput, {
      target: { value: "Pending deduplicated summary" },
    });
    await waitFor(() => {
      expect(summaryInput.value).toBe("Pending deduplicated summary");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractBody(fetchMock, 1)).toMatchObject({
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
    fireEvent.change(summaryInput, {
      target: { value: "First pending summary" },
    });
    await waitFor(() => {
      expect(summaryInput.value).toBe("First pending summary");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.keyDown(summaryInput, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractBody(fetchMock, 1)).toMatchObject({
      changes: [
        {
          field_key: "timeline.summary",
          value: "First pending summary",
        },
      ],
    });

    fireEvent.change(summaryInput, {
      target: { value: "Second pending summary" },
    });
    await waitFor(() => {
      expect(summaryInput.value).toBe("Second pending summary");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
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
    expect(extractBody(fetchMock, 2)).toMatchObject({
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

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(summaryInput, { target: { value: "Bound summary" } });
    await waitFor(() => {
      expect(summaryInput.value).toBe("Bound summary");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    const requestBody = extractBody(fetchMock, 1);
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

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(summaryInput, { target: { value: "Zulu rebound" } });
    await waitFor(() => {
      expect(summaryInput.value).toBe("Zulu rebound");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(
      String(requireFetchCall(fetchMock, 1, "timeline patch request #1")[0]),
    ).toContain("/api/v1/records/record-1");
    expect(extractBody(fetchMock, 1).base_row_version).toBe(7);
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
    expect(extractJSONBody(fetchMock, 1)).toEqual({
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
    expect(extractJSONBody(fetchMock, 2)).toEqual({
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

    const summaryInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.change(summaryInput, { target: { value: "Filtered anchor" } });
    await waitFor(() => {
      expect(summaryInput.value).toBe("Filtered anchor");
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    fireEvent.blur(summaryInput);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    expect(
      String(requireFetchCall(fetchMock, 3, "timeline patch request #3")[0]),
    ).toContain("/api/v1/records/record-1");
    expect(extractBody(fetchMock, 3)).toMatchObject({
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

function extractBody(fetchSpy: ReturnType<typeof vi.fn>, index: number) {
  return requireJSONBodyAt<{
    base_row_version: number;
    changes: Array<{ field_key: string; value: string | null }>;
  }>(fetchSpy, index, `timeline request body at index ${index}`);
}

function extractJSONBody(fetchSpy: ReturnType<typeof vi.fn>, index: number) {
  return requireJSONBodyAt<Record<string, unknown>>(
    fetchSpy,
    index,
    `timeline JSON request body at index ${index}`,
  );
}
