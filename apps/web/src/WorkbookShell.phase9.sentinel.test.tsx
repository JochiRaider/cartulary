import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  extractTimelinePatchBody,
  gridScalarInput,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  timelineViewSchemaId,
  waitForTimelineWorkbookReady,
} from "./timelineWorkbookTestSupport";
import { TimelineWorkbook } from "./WorkbookShell";

const phase9Sprint0SentinelMessage =
  "Phase 9 Sprint 0 blocker sentinel: this is not behavior completion evidence; replace this sentinel with real Phase 9 implementation evidence before claiming the row complete.";

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

  it("Phase 9 I-9-GRID-01 Sprint 0 blocker sentinel", () => {
    expect.fail(phase9Sprint0SentinelMessage);
  });
});
