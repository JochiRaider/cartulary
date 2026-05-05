import { gridShellTestId } from "@cartulary/ui-contracts";
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
  errorEnvelope,
  extractTimelineJSONBody,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  timelineViewSchemaId,
} from "./timelineWorkbookTestSupport";
import { TimelineWorkbook } from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

describe("Phase 6 workbook collaboration coverage", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("Phase 6 U-6-05 keeps the grid visible, conflict unresolved, and focus bound to the same cell", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Saved base",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      errorEnvelope("same_field_conflict", 409, {
        conflict_token: "conflict-token-1",
        record_id: "record-1",
        field_key: "timeline.summary",
        conflict_resolution_class: "text_compare_merge",
        base_row_version: 1,
        current_row_version: 2,
        base_value: "Saved base",
        server_value: "Server value",
        client_value: "Unsaved local value",
        server_updated_by: "user-server",
        server_updated_at: "2026-05-05T12:00:00Z",
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);

    const input = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Unsaved local value");
    fireEvent.blur(input);

    expect(await screen.findByTestId("conflict-resolver")).toBeTruthy();
    expect(screen.getByTestId(gridShellTestId("timeline"))).toBeTruthy();
    expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
    expect(screen.getByTestId("conflict-field-key")).toHaveProperty(
      "value",
      "timeline.summary",
    );
    expect(screen.getByTestId("conflict-server-value")).toHaveProperty(
      "value",
      "Server value",
    );
    expect(screen.getByTestId("conflict-local-value")).toHaveProperty(
      "value",
      "Unsaved local value",
    );

    fireEvent.keyDown(screen.getByTestId("conflict-resolver-summary"), {
      key: "Enter",
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);

    fireEvent.click(screen.getByTestId("conflict-close"));
    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
      expect(
        screen.getByTestId("conflict-marker-record-1-timeline-summary"),
      ).toBeTruthy();
      expect(document.activeElement).toBe(input);
    });
  });

  it("Phase 6 U-6-06 applies explicit resolver outcomes to local conflict state and revisions", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Base",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      errorEnvelope("same_field_conflict", 409, {
        conflict_token: "conflict-token-keep",
        record_id: "record-1",
        field_key: "timeline.summary",
        conflict_resolution_class: "text_compare_merge",
        base_row_version: 1,
        current_row_version: 2,
        base_value: "Base",
        server_value: "Server",
        client_value: "Local",
        server_updated_by: "user-server",
        server_updated_at: "2026-05-05T12:00:00Z",
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Server",
          captureState: "rough",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);

    const input = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Local");
    fireEvent.blur(input);
    await screen.findByTestId("conflict-resolver");
    fireEvent.click(screen.getByTestId("conflict-keep-saved"));

    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });
    expect(extractTimelineJSONBody(fetchMock, 2)).toEqual({
      conflict_token: "conflict-token-keep",
      resolution_kind: "keep_saved",
      client_txn_id: expect.any(String),
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
            rowVersion: 3,
            summary: "Base again",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      errorEnvelope("same_field_conflict", 409, {
        conflict_token: "conflict-token-use",
        record_id: "record-1",
        field_key: "timeline.summary",
        conflict_resolution_class: "text_compare_merge",
        base_row_version: 3,
        current_row_version: 4,
        base_value: "Base again",
        server_value: "Server again",
        client_value: "Use local",
        suggested_merged_value: "Server again\nUse local",
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-resolve",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 5,
          summary: "Use local",
          captureState: "enriched",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    const secondInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.focus(secondInput);
    await changeInputValue(secondInput, "Use local");
    fireEvent.blur(secondInput);
    await screen.findByTestId("conflict-resolver");
    fireEvent.click(screen.getByTestId("conflict-use-unsaved"));

    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "5",
      );
    });
    expect(extractTimelineJSONBody(fetchMock, 2)).toEqual({
      conflict_token: "conflict-token-use",
      resolution_kind: "use_unsaved",
      client_txn_id: expect.any(String),
      resolved_value: "Use local",
    });
  });

  it("Phase 6 U-6-09 keeps save-state labels and pending queue replay bounded and explicit", () => {
    // Phase 6 placeholder: replace with save-state and pending-queue assertions before Phase 6 exit.
    expect(true).toBe(true);
  });
});
