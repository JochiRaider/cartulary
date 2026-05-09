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
  buildRecordChangedPayload,
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  deferred,
  errorEnvelope,
  extractTimelineJSONBody,
  extractTimelinePatchBody,
  installTimelineWorkbookTestGlobals,
  latestTimelineWebSocket,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  timelineViewSchemaId,
} from "./timelineWorkbookTestSupport";
import { pendingReplayCapacity, TimelineWorkbook } from "./WorkbookShell";

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

  async function changeQueuedCellValue(
    input: HTMLInputElement | HTMLTextAreaElement,
    value: string,
  ) {
    const testId = input.getAttribute("data-testid");
    fireEvent.change(input, { target: { value } });
    await waitFor(() => {
      const currentInput =
        testId === null
          ? input
          : (screen.getByTestId(testId) as
              | HTMLInputElement
              | HTMLTextAreaElement);
      expect(currentInput).toHaveProperty("value", value);
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));
  }

  it("Phase 6 presence indicators render from keyed socket state without changing save-state", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Presence base",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId("row-record-1-summary");
    await waitFor(() => {
      expect(latestTimelineWebSocket()).not.toBeNull();
    });
    const socket = latestTimelineWebSocket();
    await waitFor(() => {
      expect(socket?.sentMessages.length).toBeGreaterThan(0);
    });
    socket?.emit({
      type: "hello_ack",
      payload: {
        connection_id: "self-connection",
        resume_token: "resume-presence",
      },
    });
    socket?.emit({
      type: "presence_snapshot",
      payload: {
        presences: [
          {
            connection_id: "self-connection",
            user_id: "self-user",
            display_name: "Self Analyst",
            sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
            record_id: "record-1",
            mode: "editing",
            field_key: "timeline.summary",
            observed_at: "2026-05-05T12:00:00Z",
            expires_at: "2026-05-05T12:01:00Z",
          },
          {
            connection_id: "other-connection",
            user_id: "other-user",
            display_name: "Other Analyst",
            sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
            record_id: "record-1",
            mode: "editing",
            field_key: "timeline.summary",
            observed_at: "2026-05-05T12:00:00Z",
            expires_at: "2026-05-05T12:01:00Z",
          },
          {
            connection_id: "saved-view-connection",
            user_id: "saved-view-user",
            display_name: "Saved View Analyst",
            sheet_ref: { kind: "saved_view", id: timelineViewSchemaId },
            record_id: "record-1",
            mode: "editing",
            field_key: "timeline.summary",
            observed_at: "2026-05-05T12:00:00Z",
            expires_at: "2026-05-05T12:01:00Z",
          },
        ],
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId("presence-header").textContent).toContain("OA");
    });
    expect(screen.getByTestId("presence-header").textContent).not.toContain(
      "SA",
    );
    expect(screen.getByTestId("presence-row-record-1").textContent).toContain(
      "OA",
    );
    expect(
      screen.getByTestId("presence-cell-record-1-timeline-summary").textContent,
    ).toContain("OA");
    expect(screen.getByTestId("save-state").textContent).toBe("Saved");

    socket?.emit({
      type: "presence_delta",
      payload: {
        delta_kind: "remove",
        presence: { connection_id: "other-connection" },
      },
    });
    await waitFor(() => {
      expect(screen.getByTestId("presence-header").textContent).not.toContain(
        "OA",
      );
    });
  });

  it("Phase 6 applies sparse live patches by record_id without moving an active cell draft", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Local base",
            details: "Old details",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    const input = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    input.focus();
    await changeInputValue(input, "Unsaved local");
    latestTimelineWebSocket()?.emit({
      type: "record_changed",
      stream_seq: 1,
      payload: buildRecordChangedPayload({
        recordId: "record-1",
        rowVersion: 2,
        clientTxnId: "remote-patch",
        changedFieldKeys: ["timeline.details"],
        affectedViews: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "patch",
            patch_cells: {
              record_id: "record-1",
              row_version: 2,
              cells: {
                "timeline.details": { value: "Remote details" },
              },
            },
          },
        ],
      }),
    });

    await waitFor(() => {
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "2",
      );
    });
    expect(input.value).toBe("Unsaved local");
    expect(document.activeElement).toBe(input);
    expect(fetchMock).toHaveBeenCalledTimes(1);
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

  it("Phase 6 resolver actions submit explicit outcomes and apply returned row state", async () => {
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

    cleanup();
    fetchMock.mockReset();
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 6,
            summary: "Merge base",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      errorEnvelope("same_field_conflict", 409, {
        conflict_token: "conflict-token-merged",
        record_id: "record-1",
        field_key: "timeline.summary",
        conflict_resolution_class: "text_compare_merge",
        base_row_version: 6,
        current_row_version: 7,
        base_value: "Merge base",
        server_value: "Merge server",
        client_value: "Merge local",
        suggested_merged_value: "Merge server\nMerge local",
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-merged",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 8,
          summary: "Merge final",
          captureState: "enriched",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    const mergedInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    fireEvent.focus(mergedInput);
    await changeInputValue(mergedInput, "Merge local");
    fireEvent.blur(mergedInput);
    await screen.findByTestId("conflict-resolver");
    expect(screen.getByTestId("conflict-merged-value")).toHaveProperty(
      "value",
      "Merge server\nMerge local",
    );
    fireEvent.change(screen.getByTestId("conflict-merged-value"), {
      target: { value: "Merge final" },
    });
    fireEvent.click(screen.getByTestId("conflict-use-merged"));

    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
      expect(screen.getByTestId("row-record-1-row-version").textContent).toBe(
        "8",
      );
    });
    expect(extractTimelineJSONBody(fetchMock, 2)).toEqual({
      conflict_token: "conflict-token-merged",
      resolution_kind: "merged_value",
      client_txn_id: expect.any(String),
      resolved_value: "Merge final",
    });
  });

  it("Phase 6 U-6-09 keeps save-state labels and pending queue replay bounded and explicit", async () => {
    const firstPendingPatch = deferred<Response>();
    const secondPendingPatch = deferred<Response>();
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "One",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Two",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(firstPendingPatch.promise);
    fetchMock.mockReturnValueOnce(secondPendingPatch.promise);

    render(<TimelineWorkbook incidentId="incident-1" />);

    const firstInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    const secondInput = (await screen.findByTestId(
      "row-record-2-summary",
    )) as HTMLInputElement;

    await changeQueuedCellValue(firstInput, "One in flight");
    fireEvent.blur(firstInput);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(screen.getByTestId("save-state").textContent).toBe("Syncing");
    });

    await changeQueuedCellValue(secondInput, "Two queued first");
    fireEvent.blur(secondInput);
    await changeQueuedCellValue(secondInput, "Two queued final");
    fireEvent.blur(secondInput);
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    expect(fetchMock).toHaveBeenCalledTimes(2);

    firstPendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-2",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "One in flight",
          captureState: "rough",
        }),
      }),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(extractTimelinePatchBody(fetchMock, 2).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "Two queued final",
      },
    ]);

    secondPendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-3",
        row: timelineRow({
          recordId: "record-2",
          rowVersion: 2,
          summary: "Two queued final",
          captureState: "rough",
        }),
      }),
    );

    await waitFor(() => {
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });
  });

  it("Phase 6 U-6-09 does not coalesce non-contiguous same-record pending patches", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "A base",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "B base",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-a1",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "A1 queued",
          captureState: "rough",
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-b1",
        row: timelineRow({
          recordId: "record-2",
          rowVersion: 2,
          summary: "B1 queued",
          captureState: "rough",
        }),
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-a2",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 3,
          summary: "A2 queued",
          captureState: "rough",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    const firstInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    const secondInput = (await screen.findByTestId(
      "row-record-2-summary",
    )) as HTMLInputElement;
    await waitFor(() => {
      expect(latestTimelineWebSocket()).not.toBeNull();
    });
    latestTimelineWebSocket()?.emit({
      type: "session_revoked",
      payload: { reason_code: "session_revoked" },
    });

    await changeQueuedCellValue(firstInput, "A1 queued");
    fireEvent.blur(firstInput);
    await changeQueuedCellValue(secondInput, "B1 queued");
    fireEvent.blur(secondInput);
    await changeQueuedCellValue(firstInput, "A2 queued");
    fireEvent.blur(firstInput);
    await new Promise((resolve) => window.setTimeout(resolve, 0));

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId("save-state").textContent).toBe("Syncing");
    expect(screen.getByTestId("pending-queue-count").textContent).toContain(
      "3",
    );

    latestTimelineWebSocket()?.emit({
      type: "hello_ack",
      payload: {
        resume_token: "resume-non-contiguous-coalescing",
      },
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain(
      "/api/v1/records/record-1",
    );
    expect(String(fetchMock.mock.calls[2]?.[0])).toContain(
      "/api/v1/records/record-2",
    );
    expect(String(fetchMock.mock.calls[3]?.[0])).toContain(
      "/api/v1/records/record-1",
    );
    expect(extractTimelinePatchBody(fetchMock, 1).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "A1 queued",
      },
    ]);
    expect(extractTimelinePatchBody(fetchMock, 2).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "B1 queued",
      },
    ]);
    expect(extractTimelinePatchBody(fetchMock, 3).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "A2 queued",
      },
    ]);
  });

  it("Phase 6 U-6-09 fixes the browser-runtime pending queue capacity at exactly 64 replay units", async () => {
    expect(pendingReplayCapacity).toBe(64);
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Queue 1",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Queue 2",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(errorEnvelope("session_required", 401));

    render(<TimelineWorkbook incidentId="incident-1" />);
    await screen.findByTestId("row-record-1-summary");

    for (let index = 1; index <= pendingReplayCapacity + 1; index += 1) {
      const recordID = index % 2 === 0 ? "record-2" : "record-1";
      const input = screen.getByTestId(
        `row-${recordID}-summary`,
      ) as HTMLInputElement;
      await changeQueuedCellValue(input, `Queue ${index} local`);
      fireEvent.blur(input);
      if (index === 1) {
        await waitFor(() => {
          expect(screen.getByTestId("pending-queue-notice")).toBeTruthy();
          expect(screen.getByTestId("save-state").textContent).toBe("Syncing");
        });
      }
    }

    await waitFor(() => {
      expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
      expect(screen.getByTestId("pending-queue-notice").textContent).toContain(
        "Local pending queue is full",
      );
      expect(screen.getByTestId("pending-queue-count").textContent).toContain(
        "64",
      );
    });
    expect(screen.getByTestId("row-record-1-summary")).toHaveProperty(
      "value",
      "Queue 65 local",
    );
    expect(
      fetchMock.mock.calls.filter(([url]) =>
        String(url).includes("/api/v1/records/"),
      ),
    ).toHaveLength(1);
  });

  it("Phase 6 U-6-09 preserves queued work through session revocation and resumes after re-authentication", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Auth base",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-auth",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Auth replay",
          captureState: "rough",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    const input = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    await waitFor(() => {
      expect(latestTimelineWebSocket()).not.toBeNull();
    });
    latestTimelineWebSocket()?.emit({
      type: "session_revoked",
      payload: { reason_code: "session_revoked" },
    });

    await changeQueuedCellValue(input, "Auth replay");
    fireEvent.blur(input);
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId("save-state").textContent).toBe("Syncing");
    expect(screen.getByTestId("pending-queue-notice")).toBeTruthy();
    expect(input.value).toBe("Auth replay");

    latestTimelineWebSocket()?.emit({
      type: "hello_ack",
      payload: {
        resume_token: "resume-after-auth",
      },
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "Auth replay",
      },
    ]);
    await waitFor(() => {
      expect(screen.getByTestId("save-state").textContent).toBe("Saved");
    });
  });

  it("Phase 6 U-6-09 moves the blocking same-field conflict out of the pending queue and keeps later writes queued", async () => {
    const firstPendingPatch = deferred<Response>();
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "One",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Two",
            captureState: "rough",
          }),
        ],
      }),
    );
    fetchMock.mockReturnValueOnce(firstPendingPatch.promise);

    render(<TimelineWorkbook incidentId="incident-1" />);
    const firstInput = (await screen.findByTestId(
      "row-record-1-summary",
    )) as HTMLInputElement;
    const secondInput = (await screen.findByTestId(
      "row-record-2-summary",
    )) as HTMLInputElement;

    await changeQueuedCellValue(firstInput, "Conflict local");
    fireEvent.blur(firstInput);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    await changeQueuedCellValue(secondInput, "Still queued");
    fireEvent.blur(secondInput);

    firstPendingPatch.resolve(
      errorEnvelope("same_field_conflict", 409, {
        conflict_token: "conflict-token-queued",
        record_id: "record-1",
        field_key: "timeline.summary",
        conflict_resolution_class: "text_compare_merge",
        base_row_version: 1,
        current_row_version: 2,
        base_value: "One",
        server_value: "Server",
        client_value: "Conflict local",
      }),
    );

    await waitFor(() => {
      expect(screen.getByTestId("save-state").textContent).toBe("Conflict");
      expect(screen.getByTestId("conflict-resolver")).toBeTruthy();
      expect(screen.getByTestId("pending-queue-count").textContent).toContain(
        "1",
      );
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
