import {
  type GridColumn,
  type GridDataRow,
  type GridHandle,
  SemanticDataGrid,
} from "@cartulary/grid-adapter";
import {
  conflictMarkerTestId,
  draftCellTestId,
  gridShellTestId,
  rowCellTestId,
  saveStateTestId,
  workbookConflictControlTestId,
  workbookConflictLocalValueTestId,
  workbookConflictResolverTestId,
  workbookFocusAnchorTestId,
} from "@cartulary/ui-contracts";
import {
  createEvent,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { createRef } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  extractTimelineJSONBody,
  extractTimelinePatchBody,
  gridScalarInput,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  waitForTimelineWorkbookReady,
  waitForVisibleGridRowRecordIds,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import { clipboardTextLooksTabular } from "./utils/workbookClipboard";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

type PasteHarnessRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

const pasteColumns: readonly GridColumn<PasteHarnessRow>[] = [
  {
    contractWritable: true,
    editor: {
      commit: async () => ({ kind: "accepted" }),
      initialDraftValue: (row) => row.label,
      renderEditor: () => null,
    },
    fieldKey: "summary",
    label: "Summary",
    renderCell: ({ row }) => row.label,
  },
  {
    contractWritable: true,
    editor: {
      commit: async () => ({ kind: "accepted" }),
      initialDraftValue: (row) => row.state,
      renderEditor: () => null,
    },
    fieldKey: "state",
    label: "State",
    renderCell: ({ row }) => row.state,
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
): GridDataRow<PasteHarnessRow> {
  return {
    kind: "data",
    data: {
      label: key,
      state,
    },
    mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
    rowIdentity: { kind: "core_record", recordId: key },
  };
}

const pasteSurface = {
  kind: "view_schema",
  viewSchemaId: timelineViewSchemaId,
} as const;

function pasteAnchor(recordId: string, fieldKey: string) {
  return {
    fieldKey,
    rowIdentity: { kind: "core_record" as const, recordId },
    surface: pasteSurface,
  };
}

function pasteRecordTarget(recordId: string) {
  return {
    kind: "record" as const,
    mutationIdentity: {
      kind: "core_row_version" as const,
      baseRowVersion: 1,
    },
    rowIdentity: { kind: "core_record" as const, recordId },
    surface: pasteSurface,
  };
}

function renderPastePlanner(
  rows: readonly GridDataRow<PasteHarnessRow>[],
  options?: {
    readonly allowCreateRows?: boolean | undefined;
    readonly grouped?: boolean | undefined;
  },
) {
  const handle = createRef<GridHandle>();
  render(
    <SemanticDataGrid
      ref={handle}
      allowPasteCreateRows={options?.allowCreateRows}
      columns={pasteColumns}
      dataRows={rows}
      grouping={
        options?.grouped === true
          ? {
              fieldKey: "state",
              formatLabel: (value) => (value === null ? null : String(value)),
              getValue: (row) => row.state ?? null,
            }
          : null
      }
      surface={pasteSurface}
    />,
  );
  if (handle.current === null) {
    throw new Error("Expected the semantic paste planner handle");
  }
  return handle.current;
}

describe("keyboard and grid anchor coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("grid anchor shell support updates Cartulary anchors by record_id and field_key during keyboard navigation", async () => {
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
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 1,
            summary: "Beta",
            captureState: "rough",
          }),
        ],
      }),
    );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    summary.focus();
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
      );
    });

    fireEvent.keyDown(summary, { key: "Escape" });
    const firstSummaryCell = screen
      .getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.activity_synopsis_text",
        ),
      )
      .closest('[role="gridcell"]');
    fireEvent.keyDown(firstSummaryCell as HTMLElement, { key: "ArrowDown" });
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000002:timeline.activity_synopsis_text`,
      );
      expect(document.activeElement).toBe(
        screen
          .getByTestId(
            rowCellTestId(
              "20000000-0000-4000-8000-000000000002",
              "timeline.activity_synopsis_text",
            ),
          )
          .closest('[role="gridcell"]'),
      );
    });

    fireEvent.keyDown(
      screen
        .getByTestId(
          rowCellTestId(
            "20000000-0000-4000-8000-000000000002",
            "timeline.activity_synopsis_text",
          ),
        )
        .closest('[role="gridcell"]') as HTMLElement,
      {
        key: "ArrowRight",
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000002:timeline.data_source_text`,
      );
      expect(document.activeElement).toBe(
        screen
          .getByTestId(
            rowCellTestId(
              "20000000-0000-4000-8000-000000000002",
              "timeline.data_source_text",
            ),
          )
          .closest('[role="gridcell"]'),
      );
    });

    fireEvent.keyDown(
      screen
        .getByTestId(
          rowCellTestId(
            "20000000-0000-4000-8000-000000000002",
            "timeline.data_source_text",
          ),
        )
        .closest('[role="gridcell"]') as HTMLElement,
      {
        key: "ArrowLeft",
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000002:timeline.activity_synopsis_text`,
      );
    });
  });

  it("grid anchor shell support retains edge anchors and preserves drafts across valid focus movement", async () => {
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
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "Draft before movement",
          captureState: "enriched",
        }),
      }),
    );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    await changeInputValue(summary, "Draft before movement");
    fireEvent.keyDown(summary, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1).changes[0]).toEqual({
      field_key: "timeline.activity_synopsis_text",
      value: "Draft before movement",
    });
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000002:timeline.activity_synopsis_text`,
      );
      expect(
        screen.getByTestId(
          rowCellTestId(
            "20000000-0000-4000-8000-000000000001",
            "timeline.activity_synopsis_text",
          ),
        ).textContent,
      ).toBe("Draft before movement");
    });

    const firstCell = screen
      .getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.activity_synopsis_text",
        ),
      )
      .closest('[role="gridcell"]') as HTMLElement;
    fireEvent.mouseDown(firstCell);
    fireEvent.keyDown(firstCell, { key: "ArrowUp" });
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
      );
    });
  });

  it("grid anchor shell support updates Cartulary anchors for Enter and Shift+Enter while Tab exits", async () => {
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
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 1,
            summary: "Beta",
            captureState: "rough",
          }),
        ],
      }),
    );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    summary.focus();
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
      );
    });

    fireEvent.keyDown(summary, { key: "Enter" });
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000002:timeline.activity_synopsis_text`,
      );
      expect(document.activeElement).toBe(
        screen
          .getByTestId(
            rowCellTestId(
              "20000000-0000-4000-8000-000000000002",
              "timeline.activity_synopsis_text",
            ),
          )
          .closest('[role="gridcell"]'),
      );
    });

    const record2Editor = gridScalarInput(
      container,
      "20000000-0000-4000-8000-000000000002",
      "timeline.activity_synopsis_text",
    );
    fireEvent.keyDown(record2Editor, { key: "Enter", shiftKey: true });
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
      );
      expect(document.activeElement).toBe(
        screen
          .getByTestId(
            rowCellTestId(
              "20000000-0000-4000-8000-000000000001",
              "timeline.activity_synopsis_text",
            ),
          )
          .closest('[role="gridcell"]'),
      );
    });

    fireEvent.keyDown(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000001",
          "timeline.activity_synopsis_text",
        ),
      ),
      {
        key: "Tab",
      },
    );
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
      );
      expect(
        container
          .querySelector('[role="grid"]')
          ?.contains(document.activeElement),
      ).toBe(false);
    });
  });

  it("grid anchor shell support commits drafts before Enter navigation and retains edge Shift+Enter targets", async () => {
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
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "Enter draft before movement",
          captureState: "enriched",
        }),
      }),
    );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "20000000-0000-4000-8000-000000000001",
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
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000002:timeline.activity_synopsis_text`,
      );
    });

    const record1Editor = gridScalarInput(
      container,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    fireEvent.keyDown(record1Editor, { key: "Enter", shiftKey: true });
    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
      );
    });
  });

  it("grid anchor shell support fails closed for unavailable shortcuts without row mutation", async () => {
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

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    await screen.findByTestId(saveStateTestId());
    await waitForTimelineWorkbookReady(container, 1);

    const summary = gridScalarInput(
      container,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    ) as HTMLInputElement;
    summary.focus();

    fireEvent.keyDown(summary, { key: " ", code: "Space" });
    fireEvent.keyDown(summary, { key: "k", ctrlKey: true });
    fireEvent.keyDown(summary, { key: "Escape" });
    fireEvent.keyDown(summary, { key: "v", ctrlKey: true });

    await waitFor(() => {
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
      );
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("targets sorted paste rows by stable visible record identities", () => {
    const handle = renderPastePlanner([
      pasteGridRow("20000000-0000-4000-8000-000000000003", "closed"),
      pasteGridRow("20000000-0000-4000-8000-000000000001", "open"),
      pasteGridRow("20000000-0000-4000-8000-000000000002", "reviewed"),
    ]);

    expect(
      handle.planPasteTargets(
        pasteAnchor("20000000-0000-4000-8000-000000000001", "summary"),
        {
          columnCount: 2,
          rowCount: 2,
        },
      ),
    ).toEqual({
      columns: ["summary", "state"],
      rowTargets: [
        pasteRecordTarget("20000000-0000-4000-8000-000000000001"),
        pasteRecordTarget("20000000-0000-4000-8000-000000000002"),
      ],
    });
  });

  it("treats single-line comma clipboard text as scalar paste input", () => {
    expect(clipboardTextLooksTabular("alpha,beta")).toBe(false);
    expect(clipboardTextLooksTabular("alpha\tbeta")).toBe(true);
    expect(clipboardTextLooksTabular("alpha,beta\ngamma,delta")).toBe(true);
  });

  it("keeps single-line comma text on the scalar draft-input path", async () => {
    const existingRows = Array.from({ length: 9 }, (_, index) => {
      const rowNumber = index + 1;
      return timelineRow({
        recordId: `20000000-0000-4000-8000-${String(rowNumber).padStart(12, "0")}`,
        rowVersion: 1,
        occurredAt: `2026-06-${String(rowNumber).padStart(2, "0")}`,
        summary: `Existing row ${rowNumber}`,
        captureState: "rough",
      });
    });
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: existingRows,
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000010",
          rowVersion: 1,
          occurredAt: "2026-06-14,test1,host2",
          captureState: "rough",
        }),
        view_schema_id: timelineViewSchemaId,
      }),
    );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
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
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(
      fetchMock.mock.calls.some(([request]) =>
        String(request).endsWith("/clipboard-paste"),
      ),
    ).toBe(false);
  });

  it("keeps multi-row text in the native draft editor path without grid clipboard transport", async () => {
    const clipboardText = "2026-06-14,test1,host1\n2026-06-15,test2,host2";
    const createdRows = [
      timelineRow({
        recordId: "20000000-0000-4000-8000-000000000001",
        rowVersion: 1,
        occurredAt: clipboardText,
        captureState: "rough",
      }),
    ];
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: createdRows[0],
        view_schema_id: timelineViewSchemaId,
      }),
    );
    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await screen.findByTestId(saveStateTestId());

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
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(
      fetchMock.mock.calls.some(([request]) =>
        String(request).endsWith("/clipboard-paste"),
      ),
    ).toBe(false);
    expect(extractTimelineJSONBody(fetchMock, 1)).toMatchObject({
      "timeline.activity_utc_text": clipboardText,
    });
  });

  it("rendered paste dispatch uses stable anchors and quote-aware dimensions", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000003",
            rowVersion: 4,
            summary: "Closed visual first",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 2,
            summary: "Open visual anchor",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 3,
            summary: "Alpha, one",
            captureState: "enriched",
            hostRefs: [{ item_ref: "entity_mention:host-1" }],
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 8,
            summary: "Beta",
            captureState: "enriched",
            hostRefs: [{ item_ref: "entity_mention:host-2" }],
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000004",
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
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000003",
            rowVersion: 4,
            summary: "Closed visual first",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 3,
            summary: "Alpha, one",
            captureState: "enriched",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 8,
            summary: "Beta",
            captureState: "enriched",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000004",
            rowVersion: 1,
            summary: "Gamma",
            captureState: "rough",
          }),
        ],
      }),
    );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await screen.findByTestId(saveStateTestId());
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000003",
      "20000000-0000-4000-8000-000000000001",
      "20000000-0000-4000-8000-000000000002",
    ]);

    const summaryCell = screen.getByTestId(
      rowCellTestId(
        "20000000-0000-4000-8000-000000000001",
        "timeline.activity_synopsis_text",
      ),
    );
    const summaryGridCell = summaryCell.closest('[role="gridcell"]');
    expect(summaryGridCell).toBeTruthy();
    fireEvent.mouseDown(summaryGridCell as HTMLElement);
    (summaryGridCell as HTMLElement).focus();
    fireEvent.paste(summaryGridCell as HTMLElement, {
      clipboardData: {
        getData: () => '"Alpha, one",host-one\nBeta,host-two\nGamma,host-three',
      },
    });

    await waitFor(() => {
      expect(
        fetchMock.mock.calls.filter(([request]) =>
          String(request).endsWith("/clipboard-paste"),
        ),
      ).toHaveLength(1);
    });
    const pasteCallIndex = fetchMock.mock.calls.findIndex(([request]) =>
      String(request).endsWith("/clipboard-paste"),
    );
    expect(pasteCallIndex).toBeGreaterThanOrEqual(0);
    const body = extractTimelineJSONBody(fetchMock, pasteCallIndex);
    expect(body).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      format: "csv",
      start_field_key: "timeline.activity_synopsis_text",
      columns: ["timeline.activity_synopsis_text", "timeline.data_source_text"],
      targets: [
        {
          kind: "record",
          record_id: "20000000-0000-4000-8000-000000000001",
          base_row_version: 2,
        },
        {
          kind: "record",
          record_id: "20000000-0000-4000-8000-000000000002",
          base_row_version: 7,
        },
        { kind: "create" },
      ],
    });
    expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
      `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
    );
  });

  it("registers grouped paste conflicts without losing selection continuity", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Stale first",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000003",
            rowVersion: 1,
            summary: "Created after conflicts",
            captureState: "rough",
          }),
        ],
        conflicts: [
          {
            conflict_token: "conflict-token-first",
            record_id: "20000000-0000-4000-8000-000000000001",
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
            record_id: "20000000-0000-4000-8000-000000000002",
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
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 2,
            summary: "Server first",
            captureState: "enriched",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 2,
            summary: "Server second",
            captureState: "enriched",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000003",
            rowVersion: 1,
            summary: "Created after conflicts",
            captureState: "rough",
          }),
        ],
      }),
    );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await screen.findByTestId(saveStateTestId());
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
      "20000000-0000-4000-8000-000000000002",
    ]);

    const summaryCell = screen.getByTestId(
      rowCellTestId(
        "20000000-0000-4000-8000-000000000001",
        "timeline.activity_synopsis_text",
      ),
    );
    const summaryGridCell = summaryCell.closest('[role="gridcell"]');
    expect(summaryGridCell).toBeTruthy();
    fireEvent.mouseDown(summaryGridCell as HTMLElement);
    (summaryGridCell as HTMLElement).focus();
    fireEvent.paste(summaryGridCell as HTMLElement, {
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
          `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
    );
    expect(extractTimelineJSONBody(fetchMock, pasteCallIndex)).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      format: "csv",
      start_field_key: "timeline.activity_synopsis_text",
      columns: ["timeline.activity_synopsis_text"],
      targets: [
        {
          kind: "record",
          record_id: "20000000-0000-4000-8000-000000000001",
          base_row_version: 1,
        },
        {
          kind: "record",
          record_id: "20000000-0000-4000-8000-000000000002",
          base_row_version: 1,
        },
        { kind: "create" },
      ],
    });
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000001",
      "20000000-0000-4000-8000-000000000002",
      "20000000-0000-4000-8000-000000000003",
    ]);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(
        screen.getByTestId(gridShellTestId(timelineViewSchemaId)),
      ).toBeTruthy();
      expect(screen.getByTestId(workbookFocusAnchorTestId()).textContent).toBe(
        `${timelineViewSchemaId}:20000000-0000-4000-8000-000000000001:timeline.activity_synopsis_text`,
      );
      expect(
        screen.getByTestId(
          conflictMarkerTestId(
            "20000000-0000-4000-8000-000000000002",
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toBeTruthy();
      expect(
        screen.getByTestId(workbookConflictControlTestId("paste-navigator")),
      ).toBeTruthy();
      expect(
        screen.getByTestId(workbookConflictControlTestId("paste-position"))
          .textContent,
      ).toBe("1 of 2");
      expect(
        screen.getByTestId(workbookConflictLocalValueTestId()),
      ).toHaveProperty("value", "Client first");
    }, pasteConflictWait);

    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("paste-next")),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(workbookConflictControlTestId("paste-position"))
          .textContent,
      ).toBe("2 of 2");
      expect(
        screen.getByTestId(workbookConflictLocalValueTestId()),
      ).toHaveProperty("value", "Client second");
    }, pasteConflictWait);

    fireEvent.click(screen.getByTestId(workbookConflictControlTestId("close")));
    await waitFor(() => {
      expect(screen.queryByTestId(workbookConflictResolverTestId())).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(
        screen.getByTestId(
          conflictMarkerTestId(
            "20000000-0000-4000-8000-000000000002",
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toBeTruthy();
    }, pasteConflictWait);
  });

  it("maps filtered overflow to explicit create-row targets", () => {
    const handle = renderPastePlanner(
      [pasteGridRow("20000000-0000-4000-8000-000000000002", "reviewed")],
      {
        allowCreateRows: true,
      },
    );

    expect(
      handle.planPasteTargets(
        pasteAnchor("20000000-0000-4000-8000-000000000002", "state"),
        {
          columnCount: 1,
          rowCount: 3,
        },
      ),
    ).toEqual({
      columns: ["state"],
      rowTargets: [
        pasteRecordTarget("20000000-0000-4000-8000-000000000002"),
        { createIndex: 0, kind: "create", surface: pasteSurface },
        { createIndex: 1, kind: "create", surface: pasteSurface },
      ],
    });
  });

  it("rejects invalid anchors and create-disabled grouped overflow", () => {
    const handle = renderPastePlanner(
      [
        pasteGridRow("20000000-0000-4000-8000-000000000001", "open"),
        pasteGridRow("20000000-0000-4000-8000-000000000002", "reviewed"),
      ],
      { grouped: true },
    );

    expect(
      handle.planPasteTargets(pasteAnchor("", "summary"), {
        columnCount: 1,
        rowCount: 1,
      }),
    ).toBeNull();
    expect(
      handle.planPasteTargets(
        pasteAnchor("20000000-0000-4000-8000-000000000002", "summary"),
        {
          columnCount: 1,
          rowCount: 2,
        },
      ),
    ).toBeNull();
  });

  it("dispatches one versioned multi-row tag command from Timeline selection", async () => {
    const selectedRows = [
      timelineRow({
        recordId: "11111111-1111-4111-8111-111111111111",
        rowVersion: 3,
        summary: "Alpha",
        captureState: "rough",
      }),
      timelineRow({
        recordId: "22222222-2222-4222-8222-222222222222",
        rowVersion: 4,
        summary: "Beta",
        captureState: "reviewed",
      }),
    ];
    fetchMock
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: selectedRows,
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          change_set_id: "30000000-0000-4000-8000-000000000001",
          conflicts: [],
          rows: selectedRows,
          view_schema_id: timelineViewSchemaId,
        }),
      )
      .mockResolvedValueOnce(
        successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: selectedRows,
        }),
      );

    const { container } = render(
      <TimelineWorkbookRuntimeFixture
        currentIncidentRole="editor"
        incidentId="10000000-0000-4000-8000-000000000001"
      />,
    );
    await waitForTimelineWorkbookReady(container, 2);
    fireEvent.click(
      screen.getByRole("checkbox", {
        name: "Select record 11111111-1111-4111-8111-111111111111",
      }),
    );
    fireEvent.click(
      screen.getByRole("checkbox", {
        name: "Select record 22222222-2222-4222-8222-222222222222",
      }),
    );
    fireEvent.change(
      screen.getByRole("textbox", {
        name: "Tag for selected Timeline records",
      }),
      { target: { value: "bulk-tag" } },
    );
    fireEvent.click(screen.getByRole("button", { name: "Assign tag" }));

    const bulkCallIndex = await waitForFetchCallIndex(
      fetchMock,
      "Timeline bulk tag assignment",
      (call) =>
        fetchCallMethod(call) === "POST" &&
        fetchCallURL(call).endsWith(
          `/views/${timelineViewSchemaId}/bulk-mutations`,
        ),
    );
    expect(extractTimelineJSONBody(fetchMock, bulkCallIndex)).toMatchObject({
      view_schema_id: timelineViewSchemaId,
      kind: "multi_row_tag_assignment_v1",
      tag_name: "bulk-tag",
      targets: [
        {
          record_id: "11111111-1111-4111-8111-111111111111",
          base_row_version: 3,
        },
        {
          record_id: "22222222-2222-4222-8222-222222222222",
          base_row_version: 4,
        },
      ],
    });
    await screen.findByText("Assigned tag to 2 selected records.");
    expect(screen.getByText("2 selected")).toBeTruthy();
  });

  it("requires a valid semantic anchor for paste planning", () => {
    const handle = renderPastePlanner([
      pasteGridRow("20000000-0000-4000-8000-000000000001", "open"),
      pasteGridRow("20000000-0000-4000-8000-000000000002", "closed"),
    ]);
    const vendorSelection = { fieldKey: "state", rowIndex: 1 };

    expect(
      handle.planPasteTargets(pasteAnchor("", vendorSelection.fieldKey), {
        columnCount: 1,
        rowCount: 1,
      }),
    ).toBeNull();
  });
});
