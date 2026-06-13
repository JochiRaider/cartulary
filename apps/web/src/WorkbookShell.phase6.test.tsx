import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  gridShellTestId,
  saveStateTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  buildRecordChangedPayload,
  changeInputValue,
  changeQueuedCellValue,
  cleanupTimelineWorkbookTestGlobals,
  deferred,
  errorEnvelope,
  extractTimelineConflictResolutionBody,
  extractTimelineRecordPatchBody,
  findWorkbookCell,
  installTimelineWorkbookTestGlobals,
  latestTimelineWebSocket,
  routeTimelineWorkbookFetchMock,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRecordPatchCallURLs,
  timelineRow,
  timelineViewSchemaId,
  waitForPendingQueueState,
  waitForTimelineConflictResolutionCalls,
  waitForTimelineRecordPatchCalls,
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

  function retryableErrorEnvelope(code: string, status: number) {
    return new Response(
      JSON.stringify({
        error: {
          status,
          code,
          message: code,
          request_id: "req-retryable-error",
          retryable: true,
          details: {},
        },
      }),
      {
        status,
        headers: { "Content-Type": "application/json" },
      },
    );
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
    await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    );
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
    await waitFor(() => {
      expect(
        screen.getByTestId(
          cellPresenceMarkerTestId("record-1", "timeline.summary"),
        ).textContent,
      ).toContain("OA");
    });
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");

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
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
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
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("2");
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

    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Unsaved local value");
    fireEvent.blur(input);

    expect(await screen.findByTestId("conflict-resolver")).toBeTruthy();
    expect(
      screen.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeTruthy();
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Conflict");
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
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(
        screen.getByTestId(
          conflictMarkerTestId("record-1", "timeline.summary"),
        ),
      ).toBeTruthy();
      expect(document.activeElement).toBe(input);
    });
  });

  it("Phase 6 resolver keep-saved action submits an explicit outcome and applies returned row state", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
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
    routedFetch.mockRecordPatchOnce(
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
    routedFetch.mockConflictResolutionOnce(
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

    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Local");
    fireEvent.blur(input);
    await screen.findByTestId("conflict-resolver");
    fireEvent.click(screen.getByTestId("conflict-keep-saved"));
    await waitForTimelineConflictResolutionCalls(fetchMock, 1);

    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("2");
    });
    expect(extractTimelineConflictResolutionBody(fetchMock, 0)).toEqual({
      conflict_token: "conflict-token-keep",
      resolution_kind: "keep_saved",
      client_txn_id: expect.any(String),
    });
  });

  it("Phase 6 resolver use-unsaved action submits an explicit outcome and applies returned row state", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
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
    routedFetch.mockRecordPatchOnce(
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
    routedFetch.mockConflictResolutionOnce(
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
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Use local");
    fireEvent.blur(input);
    await screen.findByTestId("conflict-resolver");
    fireEvent.click(screen.getByTestId("conflict-use-unsaved"));
    await waitForTimelineConflictResolutionCalls(fetchMock, 1);

    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("5");
    });
    expect(extractTimelineConflictResolutionBody(fetchMock, 0)).toEqual({
      conflict_token: "conflict-token-use",
      resolution_kind: "use_unsaved",
      client_txn_id: expect.any(String),
      resolved_value: "Use local",
    });
  });

  it("Phase 6 resolver merged-value action submits an explicit outcome and applies returned row state", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
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
    routedFetch.mockRecordPatchOnce(
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
    routedFetch.mockConflictResolutionOnce(
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
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Merge local");
    fireEvent.blur(input);
    await screen.findByTestId("conflict-resolver");
    expect(screen.getByTestId("conflict-merged-value")).toHaveProperty(
      "value",
      "Merge server\nMerge local",
    );
    fireEvent.change(screen.getByTestId("conflict-merged-value"), {
      target: { value: "Merge final" },
    });
    fireEvent.click(screen.getByTestId("conflict-use-merged"));
    await waitForTimelineConflictResolutionCalls(fetchMock, 1);

    await waitFor(() => {
      expect(screen.queryByTestId("conflict-resolver")).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen.getByTestId(timelineRowVersionTestId("record-1")).textContent,
      ).toBe("8");
    });
    expect(extractTimelineConflictResolutionBody(fetchMock, 0)).toEqual({
      conflict_token: "conflict-token-merged",
      resolution_kind: "merged_value",
      client_txn_id: expect.any(String),
      resolved_value: "Merge final",
    });
  });

  it("Phase 6 U-6-09 keeps save-state labels and pending queue replay bounded and explicit", async () => {
    const firstPendingPatch = deferred<Response>();
    const secondPendingPatch = deferred<Response>();
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
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
    routedFetch.mockRecordPatchOnce(firstPendingPatch.promise);
    routedFetch.mockRecordPatchOnce(secondPendingPatch.promise);

    render(<TimelineWorkbook incidentId="incident-1" />);

    const firstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    const secondInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-2",
      "timeline.summary",
    )) as HTMLInputElement;

    await changeQueuedCellValue(firstInput, "One in flight");
    fireEvent.blur(firstInput);
    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Syncing");
    });

    await changeQueuedCellValue(secondInput, "Two queued first");
    fireEvent.blur(secondInput);
    await changeQueuedCellValue(secondInput, "Two queued final");
    fireEvent.blur(secondInput);
    await waitForPendingQueueState({
      expectedPendingUnits: 2,
      expectedSaveState: "Syncing",
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toHaveLength(1);

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

    await waitForTimelineRecordPatchCalls(fetchMock, 2);
    expect(extractTimelineRecordPatchBody(fetchMock, 1).changes).toEqual([
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
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
  });

  it("Phase 6 U-6-09 does not coalesce non-contiguous same-record pending patches", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
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
    routedFetch.mockRecordPatchOnce(
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
    routedFetch.mockRecordPatchOnce(
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
    routedFetch.mockRecordPatchOnce(
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
    const firstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    const secondInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-2",
      "timeline.summary",
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

    await waitForPendingQueueState({
      expectedPendingUnits: 3,
      expectedSaveState: "Syncing",
      noticeIncludes: "Authentication is required",
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toEqual([]);

    latestTimelineWebSocket()?.emit({
      type: "hello_ack",
      payload: {
        resume_token: "resume-non-contiguous-coalescing",
      },
    });

    await waitForTimelineRecordPatchCalls(fetchMock, 3);
    expect(timelineRecordPatchCallURLs(fetchMock)).toEqual([
      "/api/v1/records/record-1",
      "/api/v1/records/record-2",
      "/api/v1/records/record-1",
    ]);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
    expect(extractTimelineRecordPatchBody(fetchMock, 0).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "A1 queued",
      },
    ]);
    expect(extractTimelineRecordPatchBody(fetchMock, 1).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "B1 queued",
      },
    ]);
    expect(extractTimelineRecordPatchBody(fetchMock, 2).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "A2 queued",
      },
    ]);
  });

  it("Phase 6 U-6-09 exposes the browser-runtime pending queue capacity as exactly 64 replay units", () => {
    expect(pendingReplayCapacity).toBe(64);
  });

  it("Phase 6 U-6-09 preserves queued work through session revocation and resumes after re-authentication", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
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
    routedFetch.mockRecordPatchOnce(
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
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
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
    await waitForPendingQueueState({
      expectedPendingUnits: 1,
      expectedSaveState: "Syncing",
      noticeIncludes: "Authentication is required",
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toEqual([]);
    expect(input.value).toBe("Auth replay");

    latestTimelineWebSocket()?.emit({
      type: "hello_ack",
      payload: {
        resume_token: "resume-after-auth",
      },
    });

    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    expect(extractTimelineRecordPatchBody(fetchMock, 0).changes).toEqual([
      {
        field_key: "timeline.summary",
        value: "Auth replay",
      },
    ]);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
  });

  it("Phase 6 U-6-09 moves the blocking same-field conflict out of the pending queue and keeps later writes queued", async () => {
    const firstPendingPatch = deferred<Response>();
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
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
    routedFetch.mockRecordPatchOnce(firstPendingPatch.promise);

    render(<TimelineWorkbook incidentId="incident-1" />);
    const firstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    const secondInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-2",
      "timeline.summary",
    )) as HTMLInputElement;

    await changeQueuedCellValue(firstInput, "Conflict local");
    fireEvent.blur(firstInput);
    await waitForTimelineRecordPatchCalls(fetchMock, 1);
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
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(screen.getByTestId("conflict-resolver")).toBeTruthy();
      expect(screen.getByTestId("pending-queue-count").textContent).toContain(
        "1",
      );
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toHaveLength(1);
  });

  it("FE-U-P4-01 drives WorkbookShell admission, coalescing, and save-state from the shared pending queue model", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
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
    routedFetch.mockRecordPatchOnce(errorEnvelope("session_required", 401));

    render(<TimelineWorkbook incidentId="incident-1" />);
    const firstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    const secondInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-2",
      "timeline.summary",
    )) as HTMLInputElement;

    await changeQueuedCellValue(firstInput, "Queue 1 local");
    fireEvent.blur(firstInput);
    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    await waitForPendingQueueState({
      expectedPendingUnits: 1,
      expectedSaveState: "Syncing",
      noticeIncludes: "Authentication is required",
    });

    await changeQueuedCellValue(secondInput, "Queue 2 first");
    fireEvent.blur(secondInput);
    await changeQueuedCellValue(secondInput, "Queue 2 final");
    fireEvent.blur(secondInput);
    await waitForPendingQueueState({
      expectedPendingUnits: 2,
      expectedSaveState: "Syncing",
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toHaveLength(1);
    expect(secondInput.value).toBe("Queue 2 final");
  });

  it("FE-U-P4-01 drives WorkbookShell retry and success settlement from the shared pending queue model", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-1",
            rowVersion: 1,
            summary: "Retry one",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "record-2",
            rowVersion: 1,
            summary: "Retry two",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(
      retryableErrorEnvelope("future_retryable_public_error", 409),
    );
    routedFetch.mockRecordPatchOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-retry-1",
        row: timelineRow({
          recordId: "record-1",
          rowVersion: 2,
          summary: "Retry head",
          captureState: "rough",
        }),
      }),
    );
    routedFetch.mockRecordPatchOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "change-set-retry-2",
        row: timelineRow({
          recordId: "record-2",
          rowVersion: 2,
          summary: "Retry behind",
          captureState: "rough",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    const retryFirstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-1",
      "timeline.summary",
    )) as HTMLInputElement;
    const retrySecondInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-2",
      "timeline.summary",
    )) as HTMLInputElement;
    await changeQueuedCellValue(retryFirstInput, "Retry head");
    fireEvent.blur(retryFirstInput);
    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    await changeQueuedCellValue(retrySecondInput, "Retry behind");
    fireEvent.blur(retrySecondInput);
    await waitForTimelineRecordPatchCalls(fetchMock, 3);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toEqual([
      "/api/v1/records/record-1",
      "/api/v1/records/record-1",
      "/api/v1/records/record-2",
    ]);
  });

  it("FE-U-P4-01 drives WorkbookShell terminal halt presentation from the shared pending queue model", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-halt",
            rowVersion: 1,
            summary: "Halt base",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(
      errorEnvelope("future_terminal_public_error", 409),
    );
    render(<TimelineWorkbook incidentId="incident-1" />);
    const haltInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-halt",
      "timeline.summary",
    )) as HTMLInputElement;
    await changeQueuedCellValue(haltInput, "Halt local");
    fireEvent.blur(haltInput);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(screen.getByTestId("pending-queue-count").textContent).toContain(
        "1",
      );
      expect(screen.getByTestId("pending-queue-notice").textContent).toContain(
        "future_terminal_public_error",
      );
    });
  });
});
