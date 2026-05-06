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

    await changeInputValue(firstInput, "One in flight");
    fireEvent.blur(firstInput);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(screen.getByTestId("save-state").textContent).toBe("Syncing");
    });

    await changeInputValue(secondInput, "Two queued first");
    fireEvent.blur(secondInput);
    await changeInputValue(secondInput, "Two queued final");
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

  it("Phase 6 U-6-09 fixes the browser-runtime pending queue capacity at exactly 64 replay units", () => {
    expect(pendingReplayCapacity).toBe(64);
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

    await changeInputValue(input, "Auth replay");
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

    await changeInputValue(firstInput, "Conflict local");
    fireEvent.blur(firstInput);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    await changeInputValue(secondInput, "Still queued");
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
