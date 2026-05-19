import {
  buildGridPresentationRows,
  type GridColumn,
  type GridRow,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import { conflictMarkerTestId } from "@cartulary/ui-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

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
  timelineViewSchemaId,
  waitForTimelineWorkbookReady,
} from "./timelineWorkbookTestSupport";
import { clipboardTextLooksTabular, TimelineWorkbook } from "./WorkbookShell";

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

    await screen.findByTestId("save-state");
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    summary.focus();
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-1:timeline.summary",
      );
    });

    fireEvent.keyDown(summary, { key: "ArrowDown" });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-2:timeline.summary",
      );
      expect(document.activeElement).toBe(
        screen.getByTestId("row-record-2-summary"),
      );
    });

    fireEvent.keyDown(screen.getByTestId("row-record-2-summary"), {
      key: "ArrowRight",
    });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-2:timeline.host_refs",
      );
      expect(document.activeElement).toBe(
        screen.getByTestId("row-record-2-hostRefs-input"),
      );
    });

    fireEvent.keyDown(screen.getByTestId("row-record-2-hostRefs-input"), {
      key: "ArrowLeft",
    });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-2:timeline.summary",
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

    await screen.findByTestId("save-state");
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    await changeInputValue(summary, "Draft before movement");
    fireEvent.keyDown(summary, { key: "ArrowDown" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1).changes[0]).toEqual({
      field_key: "timeline.summary",
      value: "Draft before movement",
    });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-2:timeline.summary",
      );
      expect(
        (screen.getByTestId("row-record-1-summary") as HTMLInputElement).value,
      ).toBe("Draft before movement");
    });

    fireEvent.keyDown(screen.getByTestId("row-record-1-summary"), {
      key: "ArrowUp",
    });
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

    await screen.findByTestId("save-state");
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    summary.focus();
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-1:timeline.summary",
      );
    });

    fireEvent.keyDown(summary, { key: "Enter" });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-2:timeline.summary",
      );
      expect(document.activeElement).toBe(
        screen.getByTestId("row-record-2-summary"),
      );
    });

    fireEvent.keyDown(screen.getByTestId("row-record-2-summary"), {
      key: "Enter",
      shiftKey: true,
    });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-1:timeline.summary",
      );
      expect(document.activeElement).toBe(
        screen.getByTestId("row-record-1-summary"),
      );
    });

    fireEvent.keyDown(screen.getByTestId("row-record-1-summary"), {
      key: "Tab",
    });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-1:timeline.host_refs",
      );
      expect(document.activeElement).toBe(
        screen.getByTestId("row-record-1-hostRefs-input"),
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

    await screen.findByTestId("save-state");
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    await changeInputValue(summary, "Enter draft before movement");
    fireEvent.keyDown(summary, { key: "Enter" });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1).changes[0]).toEqual({
      field_key: "timeline.summary",
      value: "Enter draft before movement",
    });
    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-2:timeline.summary",
      );
    });

    fireEvent.keyDown(screen.getByTestId("row-record-1-summary"), {
      key: "Enter",
      shiftKey: true,
    });
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

    await screen.findByTestId("save-state");
    await waitForTimelineWorkbookReady(container, 1);

    const summary = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    summary.focus();

    fireEvent.keyDown(summary, { key: " ", code: "Space" });
    fireEvent.keyDown(summary, { key: "k", ctrlKey: true });
    fireEvent.keyDown(summary, { key: "Escape" });
    fireEvent.keyDown(summary, { key: "v", ctrlKey: true });

    await waitFor(() => {
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        "timeline:record-1:timeline.summary",
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
    await screen.findByTestId("save-state");
    await waitForTimelineWorkbookReady(container, 3);

    const summary = gridScalarInput(
      container,
      "record-1",
      "summary",
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
      start_field_key: "timeline.summary",
      columns: ["timeline.summary", "timeline.host_refs"],
      targets: [
        { kind: "record", record_id: "record-1", base_row_version: 2 },
        { kind: "record", record_id: "record-2", base_row_version: 7 },
        { kind: "create" },
      ],
    });
    expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
      "timeline:record-1:timeline.summary",
    );
  });

  it("Phase 9 E-9-02 registers grouped paste conflicts without losing selection continuity", async () => {
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
            field_key: "timeline.summary",
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
            field_key: "timeline.summary",
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
    await screen.findByTestId("save-state");
    await waitForTimelineWorkbookReady(container, 2);

    const summary = gridScalarInput(
      container,
      "record-1",
      "summary",
    ) as HTMLInputElement;
    summary.focus();
    fireEvent.paste(summary, {
      clipboardData: {
        getData: () => "Client first\nClient second\nCreated after conflicts",
      },
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
    expect(screen.getByTestId("timeline-grid-shell")).toBeTruthy();
    expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
      "timeline:record-1:timeline.summary",
    );
    expect(
      screen.getByTestId(conflictMarkerTestId("record-1", "timeline.summary")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(conflictMarkerTestId("record-2", "timeline.summary")),
    ).toBeTruthy();
    expect(screen.getByTestId("paste-conflict-navigator")).toBeTruthy();
    expect(screen.getByTestId("paste-conflict-position").textContent).toBe(
      "1 of 2",
    );
    expect(screen.getByTestId("conflict-local-value")).toHaveProperty(
      "value",
      "Client first",
    );

    fireEvent.click(screen.getByTestId("paste-conflict-next"));
    expect(screen.getByTestId("paste-conflict-position").textContent).toBe(
      "2 of 2",
    );
    expect(screen.getByTestId("conflict-local-value")).toHaveProperty(
      "value",
      "Client second",
    );

    fireEvent.click(screen.getByTestId("conflict-close"));
    expect(screen.queryByTestId("conflict-resolver")).toBeNull();
    expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
    expect(
      screen.getByTestId(conflictMarkerTestId("record-1", "timeline.summary")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(conflictMarkerTestId("record-2", "timeline.summary")),
    ).toBeTruthy();
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
