import {
  buildGridPresentationRows,
  type GridColumn,
  type GridRow,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import {
  conflictMarkerTestId,
  draftCellTestId,
  gridShellTestId,
  rowCellTestId,
  saveStateTestId,
} from "@cartulary/ui-contracts";
import {
  createEvent,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  extractTimelineJSONBody,
  extractTimelinePatchBody,
  focusReadyGridScalarInput,
  gridScalarInput,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  waitForTimelineWorkbookReady,
  waitForVisibleGridRowRecordIds,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import { TimelineWorkbook } from "./timeline/components/TimelineWorkbook";
import { clipboardTextLooksTabular } from "./utils/workbookClipboard";

type PasteHarnessRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

const pasteColumns: readonly GridColumn<PasteHarnessRow>[] = [
  {
    fieldKey: "summary",
    label: "Summary",
    renderCell: (row) => row.label,
  },
  {
    fieldKey: "state",
    label: "State",
    renderCell: (row) => row.state,
  },
];
const pasteConflictWait = { timeout: 3000 };

function fetchCallURL(call: readonly unknown[]) {
  const input = call[0];
  if (typeof input === "string") {
    return input;
  }
  if (input instanceof URL) {
    return input.toString();
  }
  if (typeof Request !== "undefined" && input instanceof Request) {
    return input.url;
  }
  return "";
}

function fetchCallMethod(call: readonly unknown[]) {
  const input = call[0];
  const init = call[1] as RequestInit | undefined;
  if (init?.method) {
    return init.method.toUpperCase();
  }
  if (typeof Request !== "undefined" && input instanceof Request) {
    return input.method.toUpperCase();
  }
  return "GET";
}

function fetchCallSummary(fetchMock: TimelineWorkbookFetchMock) {
  return fetchMock.mock.calls
    .map((call, index) => {
      const method = fetchCallMethod(call);
      const url = fetchCallURL(call);
      return `${index}: ${method} ${url || "(unknown url)"}`;
    })
    .join("\n");
}

async function waitForFetchCallIndex(
  fetchMock: TimelineWorkbookFetchMock,
  description: string,
  predicate: (call: readonly unknown[]) => boolean,
) {
  let matchIndex = -1;
  await waitFor(
    () => {
      matchIndex = fetchMock.mock.calls.findIndex((call) => predicate(call));
      if (matchIndex < 0) {
        throw new Error(`Expected fetch call for ${description}.`);
      }
    },
    {
      onTimeout: (error) =>
        new Error(
          `${error.message}\nObserved fetch calls:\n${fetchCallSummary(
            fetchMock,
          )}`,
        ),
      timeout: pasteConflictWait.timeout,
    },
  );
  return matchIndex;
}

function pasteGridRow(
  key: string,
  state: string | null | undefined,
): GridRow<PasteHarnessRow> {
  return {
    data: {
      label: key,
      state,
    },
    key,
    recordId: key,
  };
}

function pasteDraftRow(
  key: string,
  state: string | null | undefined,
): GridRow<PasteHarnessRow> {
  return {
    data: {
      label: key,
      state,
    },
    key,
    recordId: null,
  };
}

describe("Phase 9 Sprint 1 keyboard and grid anchor coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("Phase 9 grid anchor shell support updates Cartulary anchors by record_id and field_key during keyboard navigation", async () => {
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
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Beta",
            captureState: "rough",
          }),
        ],
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    summary.focus();
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-1:timeline.activity_synopsis_text`,
      );
    });

    fireEvent.keyDown(summary, { key: "ArrowDown" });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-2:timeline.activity_synopsis_text`,
      );
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-2", "timeline.activity_synopsis_text"),
        ),
      );
    });

    fireEvent.keyDown(
      screen.getByTestId(
        rowCellTestId("record-2", "timeline.activity_synopsis_text"),
      ),
      {
        key: "ArrowRight",
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-2:timeline.data_source_text`,
      );
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-2", "timeline.data_source_text"),
        ),
      );
    });

    fireEvent.keyDown(
      screen.getByTestId(
        rowCellTestId("record-2", "timeline.data_source_text"),
      ),
      {
        key: "ArrowLeft",
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-2:timeline.activity_synopsis_text`,
      );
    });
  });

  it("Phase 9 grid anchor shell support clears invalid anchors and preserves drafts across valid focus movement", async () => {
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
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Beta",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-phase9-anchor",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Draft before movement",
          captureState: "enriched",
        }),
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    await changeInputValue(summary, "Draft before movement");
    fireEvent.keyDown(summary, { key: "ArrowDown" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1).changes[0]).toEqual({
      field_key: "timeline.activity_synopsis_text",
      value: "Draft before movement",
    });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-2:timeline.activity_synopsis_text`,
      );
      expect(
        (
          screen.getByTestId(
            rowCellTestId("record-1", "timeline.activity_synopsis_text"),
          ) as HTMLInputElement
        ).value,
      ).toBe("Draft before movement");
    });

    fireEvent.keyDown(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      ),
      {
        key: "ArrowUp",
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "cleared",
      );
    });
  });

  it("Phase 9 grid anchor shell support updates Cartulary anchors for Enter, Shift+Enter, and Tab navigation", async () => {
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
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Beta",
            captureState: "rough",
          }),
        ],
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    summary.focus();
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-1:timeline.activity_synopsis_text`,
      );
    });

    fireEvent.keyDown(summary, { key: "Enter" });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-2:timeline.activity_synopsis_text`,
      );
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-2", "timeline.activity_synopsis_text"),
        ),
      );
    });

    fireEvent.keyDown(
      screen.getByTestId(
        rowCellTestId("record-2", "timeline.activity_synopsis_text"),
      ),
      {
        key: "Enter",
        shiftKey: true,
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-1:timeline.activity_synopsis_text`,
      );
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-1", "timeline.activity_synopsis_text"),
        ),
      );
    });

    fireEvent.keyDown(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      ),
      {
        key: "Tab",
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-1:timeline.data_source_text`,
      );
      expect(document.activeElement).toBe(
        screen.getByTestId(
          rowCellTestId("record-1", "timeline.data_source_text"),
        ),
      );
    });
  });

  it("Phase 9 grid anchor shell support commits drafts before Enter navigation and clears invalid Shift+Enter targets", async () => {
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
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Beta",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-phase9-enter-anchor",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Enter draft before movement",
          captureState: "enriched",
        }),
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    await changeInputValue(summary, "Enter draft before movement");
    fireEvent.keyDown(summary, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1).changes[0]).toEqual({
      field_key: "timeline.activity_synopsis_text",
      value: "Enter draft before movement",
    });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-2:timeline.activity_synopsis_text`,
      );
    });

    fireEvent.keyDown(
      screen.getByTestId(
        rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      ),
      {
        key: "Enter",
        shiftKey: true,
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "cleared",
      );
    });
  });

  it("Phase 9 grid anchor shell support fails closed for unavailable shortcuts without row mutation", async () => {
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

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 1);

    const summary = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    summary.focus();

    fireEvent.keyDown(summary, { key: " ", code: "Space" });
    fireEvent.keyDown(summary, { key: "k", ctrlKey: true });
    fireEvent.keyDown(summary, { key: "Escape" });
    fireEvent.keyDown(summary, { key: "v", ctrlKey: true });

    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-1:timeline.activity_synopsis_text`,
      );
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("Phase 9 I-9-GRID-01 targets sorted paste rows by stable visible record identities", () => {
    const presentationRows = buildGridPresentationRows({
      rows: [
        pasteGridRow("record-3", "closed"),
        pasteGridRow("record-1", "open"),
        pasteGridRow("record-2", "reviewed"),
      ],
    });

    expect(
      resolveGridPasteTargets({
        columns: pasteColumns,
        current: { fieldKey: "summary", recordId: "record-1" },
        pastedColumnCount: 2,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["summary", "state"],
      rowTargets: [
        { kind: "record", recordId: "record-1" },
        { kind: "record", recordId: "record-2" },
      ],
    });
  });

  it("Phase 9 I-9-GRID-01 treats single-line comma clipboard text as scalar paste input", () => {
    expect(clipboardTextLooksTabular("alpha,beta")).toBe(false);
    expect(clipboardTextLooksTabular("alpha\tbeta")).toBe(true);
    expect(clipboardTextLooksTabular("alpha,beta\ngamma,delta")).toBe(true);
  });

  it("dispatches single-line CSV pasted into the draft Time cell as a create-row paste", async () => {
    const existingRows = Array.from({ length: 9 }, (_, index) => {
      const rowNumber = index + 1;
      return timelineRow({
        recordId: `record-${rowNumber}`,
        rowVersion: 1,
        occurredAt: `2026-06-${String(rowNumber).padStart(2, "0")}`,
        summary: `Existing row ${rowNumber}`,
        captureState: "rough",
      });
    });
    const createdRow = timelineRow({
      recordId: "record-10",
      rowVersion: 1,
      occurredAt: "2026-06-14",
      summary: "test1",
      captureState: "rough",
    });
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: existingRows,
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-draft-csv-paste",
        rows: [createdRow],
        conflicts: [],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [...existingRows, createdRow],
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId(saveStateTestId());
    await waitForVisibleGridRowRecordIds(
      container,
      existingRows.map((row) => row.record_id),
    );

    const draftTime = screen.getByTestId(
      draftCellTestId("timeline.activity_utc_text"),
    );
    const pasteEvent = createEvent.paste(draftTime, {
      clipboardData: {
        getData: () => "2026-06-14,test1,host2",
      },
    });
    fireEvent(draftTime, pasteEvent);
    expect(pasteEvent.defaultPrevented).toBe(true);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(draftTime).not.toHaveProperty("value", "2026-06-14,test1,host2");
    expect(extractTimelineJSONBody(fetchMock, 1)).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      clipboard_text: "2026-06-14,test1,host2",
      format: "csv",
      start_field_key: "timeline.activity_utc_text",
      columns: [
        "timeline.activity_utc_text",
        "timeline.activity_local_text",
        "timeline.raw_activity_text",
      ],
      targets: [{ kind: "create" }],
    });
  });

  it("dispatches multi-row CSV pasted into the draft Time cell as create-row targets", async () => {
    const createdRows = [
      timelineRow({
        recordId: "record-1",
        rowVersion: 1,
        occurredAt: "2026-06-14",
        summary: "test1",
        captureState: "rough",
      }),
      timelineRow({
        recordId: "record-2",
        rowVersion: 1,
        occurredAt: "2026-06-15",
        summary: "test2",
        captureState: "rough",
      }),
    ];
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
        change_set_id: "change-set-draft-csv-paste-multi-row",
        rows: createdRows,
        conflicts: [],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: createdRows,
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId(saveStateTestId());

    const clipboardText = "2026-06-14,test1,host1\n2026-06-15,test2,host2";
    const draftTime = screen.getByTestId(
      draftCellTestId("timeline.activity_utc_text"),
    );
    const pasteEvent = createEvent.paste(draftTime, {
      clipboardData: {
        getData: () => clipboardText,
      },
    });
    fireEvent(draftTime, pasteEvent);
    expect(pasteEvent.defaultPrevented).toBe(true);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(draftTime).not.toHaveProperty("value", clipboardText);
    expect(extractTimelineJSONBody(fetchMock, 1)).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      clipboard_text: clipboardText,
      format: "csv",
      start_field_key: "timeline.activity_utc_text",
      columns: [
        "timeline.activity_utc_text",
        "timeline.activity_local_text",
        "timeline.raw_activity_text",
      ],
      targets: [{ kind: "create" }, { kind: "create" }],
    });
  });

  it("Phase 9 I-9-GRID-01 rendered paste dispatch uses stable anchors and quote-aware dimensions", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-3",
            rowVersion: 4,
            summary: "Closed visual first",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-1",
            rowVersion: 2,
            summary: "Open visual anchor",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 7,
            summary: "Reviewed visual next",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-phase9-rendered-paste",
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 3,
            summary: "Alpha, one",
            captureState: "enriched",
            hostRefs: [{ item_ref: "entity_mention:host-1" }],
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 8,
            summary: "Beta",
            captureState: "enriched",
            hostRefs: [{ item_ref: "entity_mention:host-2" }],
          }),
          timelineRow({
            recordId: "record-4",
            rowVersion: 1,
            summary: "Gamma",
            captureState: "rough",
            hostRefs: [{ item_ref: "entity_mention:host-3" }],
          }),
        ],
        conflicts: [],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-3",
            rowVersion: 4,
            summary: "Closed visual first",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-1",
            rowVersion: 3,
            summary: "Alpha, one",
            captureState: "enriched",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 8,
            summary: "Beta",
            captureState: "enriched",
          }),
          timelineRow({
            recordId: "record-4",
            rowVersion: 1,
            summary: "Gamma",
            captureState: "rough",
          }),
        ],
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId(saveStateTestId());
    await waitForVisibleGridRowRecordIds(container, [
      "record-3",
      "record-1",
      "record-2",
    ]);

    const summary = gridScalarInput(
      container,
      "record-1",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    summary.focus();
    fireEvent.paste(summary, {
      clipboardData: {
        getData: () => '"Alpha, one",host-one\nBeta,host-two\nGamma,host-three',
      },
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    const body = extractTimelineJSONBody(fetchMock, 1);
    expect(body).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      format: "csv",
      start_field_key: "timeline.activity_synopsis_text",
      columns: ["timeline.activity_synopsis_text", "timeline.data_source_text"],
      targets: [
        { kind: "record", record_id: "record-1", base_row_version: 2 },
        { kind: "record", record_id: "record-2", base_row_version: 7 },
        { kind: "create" },
      ],
    });
    expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
      `${timelineViewSchemaId}:record-1:timeline.activity_synopsis_text`,
    );
  });

  it("Phase 9 I-9-SUPPORT-01 registers grouped paste conflicts without losing selection continuity", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Stale first",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Stale second",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-phase9-rendered-paste-conflicts",
        rows: [
          timelineRow({
            recordId: "record-3",
            rowVersion: 1,
            summary: "Created after conflicts",
            captureState: "rough",
          }),
        ],
        conflicts: [
          {
            conflict_token: "conflict-token-first",
            record_id: "record-1",
            field_key: "timeline.activity_synopsis_text",
            conflict_resolution_class: "text_compare_merge",
            base_row_version: 1,
            current_row_version: 2,
            base_value: "Stale first",
            server_value: "Server first",
            client_value: "Client first",
            server_updated_by: "server-user",
            server_updated_at: "2026-05-17T20:00:00Z",
          },
          {
            conflict_token: "conflict-token-second",
            record_id: "record-2",
            field_key: "timeline.activity_synopsis_text",
            conflict_resolution_class: "text_compare_merge",
            base_row_version: 1,
            current_row_version: 2,
            base_value: "Stale second",
            server_value: "Server second",
            client_value: "Client second",
            server_updated_by: "server-user",
            server_updated_at: "2026-05-17T20:00:01Z",
          },
        ],
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
            summary: "Server first",
            captureState: "enriched",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 2,
            summary: "Server second",
            captureState: "enriched",
          }),
          timelineRow({
            recordId: "record-3",
            rowVersion: 1,
            summary: "Created after conflicts",
            captureState: "rough",
          }),
        ],
      }),
    );

    const { container } = render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId(saveStateTestId());
    await waitForVisibleGridRowRecordIds(container, ["record-1", "record-2"]);

    const summary = (await focusReadyGridScalarInput({
      container,
      fieldKey: "timeline.activity_synopsis_text",
      recordId: "record-1",
    })) as HTMLInputElement;
    fireEvent.paste(summary, {
      clipboardData: {
        getData: () => "Client first\nClient second\nCreated after conflicts",
      },
    });

    const pasteCallIndex = await waitForFetchCallIndex(
      fetchMock,
      "Timeline grouped paste conflict mutation",
      (call) =>
        fetchCallMethod(call) === "POST" &&
        fetchCallURL(call).includes(
          `/api/v1/incidents/incident-1/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
    );
    expect(extractTimelineJSONBody(fetchMock, pasteCallIndex)).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      format: "csv",
      start_field_key: "timeline.activity_synopsis_text",
      columns: ["timeline.activity_synopsis_text"],
      targets: [
        { kind: "record", record_id: "record-1", base_row_version: 1 },
        { kind: "record", record_id: "record-2", base_row_version: 1 },
        { kind: "create" },
      ],
    });
    await waitForVisibleGridRowRecordIds(container, [
      "record-1",
      "record-2",
      "record-3",
    ]);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(
        screen.getByTestId(gridShellTestId(timelineViewSchemaId)),
      ).toBeTruthy();
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        `${timelineViewSchemaId}:record-1:timeline.activity_synopsis_text`,
      );
      expect(
        screen.getByTestId(
          conflictMarkerTestId("record-1", "timeline.activity_synopsis_text"),
        ),
      ).toBeTruthy();
      expect(
        screen.getByTestId(
          conflictMarkerTestId("record-2", "timeline.activity_synopsis_text"),
        ),
      ).toBeTruthy();
      expect(screen.getByTestId("paste-conflict-navigator")).toBeTruthy();
      expect(screen.getByTestId("paste-conflict-position").textContent).toBe(
        "1 of 2",
      );
      expect(screen.getByTestId("conflict-local-value")).toHaveProperty(
        "value",
        "Client first",
      );
    }, pasteConflictWait);

    fireEvent.click(screen.getByTestId("paste-conflict-next"));
    await waitFor(() => {
      expect(screen.getByTestId("paste-conflict-position").textContent).toBe(
        "2 of 2",
      );
      expect(screen.getByTestId("conflict-local-value")).toHaveProperty(
        "value",
        "Client second",
      );
    }, pasteConflictWait);

    fireEvent.click(screen.getByTestId("conflict-close"));
    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(
        screen.getByTestId(
          conflictMarkerTestId("record-1", "timeline.activity_synopsis_text"),
        ),
      ).toBeTruthy();
      expect(
        screen.getByTestId(
          conflictMarkerTestId("record-2", "timeline.activity_synopsis_text"),
        ),
      ).toBeTruthy();
    }, pasteConflictWait);
  });

  it("Phase 9 I-9-GRID-01 maps filtered overflow to explicit create-row anchors", () => {
    const presentationRows = buildGridPresentationRows({
      rows: [pasteGridRow("record-2", "reviewed")],
    });

    expect(
      resolveGridPasteTargets({
        columns: pasteColumns,
        current: { fieldKey: "state", recordId: "record-2" },
        pastedColumnCount: 2,
        pastedRowCount: 3,
        presentationRows,
      }),
    ).toEqual({
      columns: ["state"],
      rowTargets: [
        { kind: "record", recordId: "record-2" },
        { createIndex: 0, kind: "create" },
        { createIndex: 1, kind: "create" },
      ],
    });
  });

  it("Phase 9 I-9-GRID-01 rejects group and presentation-only paste anchors", () => {
    const groupedRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      groupBy: "state",
      rows: [
        pasteGridRow("record-1", "open"),
        pasteDraftRow("draft-1", "rough"),
        pasteGridRow("record-2", "reviewed"),
      ],
    });

    expect(
      resolveGridPasteTargets({
        columns: pasteColumns,
        current: { fieldKey: "summary", recordId: "record-1" },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows: groupedRows,
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        allowCreateRows: false,
        columns: pasteColumns,
        current: { fieldKey: "summary", recordId: "record-2" },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows: groupedRows,
      }),
    ).toBeNull();
  });

  it("Phase 9 I-9-GRID-01 requires a Cartulary anchor instead of vendor coordinates alone", () => {
    const presentationRows = buildGridPresentationRows({
      rows: [
        pasteGridRow("record-1", "open"),
        pasteGridRow("record-2", "closed"),
      ],
    });
    const vendorSelection = { fieldKey: "state", rowIndex: 1 };

    expect(
      resolveGridPasteTargets({
        columns: pasteColumns,
        current: { fieldKey: vendorSelection.fieldKey, recordId: "" },
        pastedColumnCount: 1,
        pastedRowCount: 1,
        presentationRows,
      }),
    ).toBeNull();
  });
});
