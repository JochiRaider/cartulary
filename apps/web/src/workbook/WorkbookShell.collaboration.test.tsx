import {
  cellPresenceMarkerTestId,
  gridShellTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  saveStateActionButtonTestId,
  saveStateTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
  workbookConflictControlTestId,
  workbookConflictLocalValueTestId,
  workbookConflictResolverTestId,
  workbookConflictSavedValueTestId,
  workbookConflictSummaryTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryRetryButtonTestId,
  workbookEditRecoveryTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  buildRecordChangedPayload,
  changeInputValue,
  changeQueuedCellValue,
  cleanupTimelineWorkbookTestGlobals,
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
  waitForPendingQueueState,
  waitForTimelineConflictResolutionCalls,
  waitForTimelineRecordPatchCalls,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";
import { createBrowserSecureTransactionIdPort } from "./mutations/secureTransactionId";
import { WorkbookMutationRuntime } from "./runtime/WorkbookMutationRuntime";
import { pendingReplayCapacity } from "./utils/workbookPendingQueue";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

describe("workbook collaboration coverage", () => {
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

  it("presence indicators render from keyed socket state without changing save-state", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Presence base",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const activatedEditor = await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    );
    fireEvent.keyDown(activatedEditor, { key: "Escape" });
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
        server_time: "2026-07-13T12:00:00Z",
        heartbeat_interval_ms: 15_000,
        presence_ttl_ms: 45_000,
        resume_window_ms: 60_000,
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
            record_id: "20000000-0000-4000-8000-000000000001",
            mode: "editing",
            field_key: "timeline.activity_synopsis_text",
            observed_at: "2026-05-05T12:00:00Z",
            expires_at: "2026-05-05T12:01:00Z",
          },
          {
            connection_id: "other-connection",
            user_id: "other-user",
            display_name: "Other Analyst",
            sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
            record_id: "20000000-0000-4000-8000-000000000001",
            mode: "editing",
            field_key: "timeline.activity_synopsis_text",
            observed_at: "2026-05-05T12:00:00Z",
            expires_at: "2026-05-05T12:01:00Z",
          },
          {
            connection_id: "saved-view-connection",
            user_id: "saved-view-user",
            display_name: "Saved View Analyst",
            sheet_ref: { kind: "saved_view", id: timelineViewSchemaId },
            record_id: "20000000-0000-4000-8000-000000000001",
            mode: "editing",
            field_key: "timeline.activity_synopsis_text",
            observed_at: "2026-05-05T12:00:00Z",
            expires_at: "2026-05-05T12:01:00Z",
          },
        ],
      },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId(workbookPresenceSummaryTestId()).textContent,
      ).toContain("OA");
    });
    expect(
      screen.getByTestId(workbookPresenceSummaryTestId()).textContent,
    ).not.toContain("SA");
    await waitFor(() => {
      expect(
        screen.getByTestId(
          cellPresenceMarkerTestId(
            "20000000-0000-4000-8000-000000000001",
            "timeline.activity_synopsis_text",
          ),
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
      expect(
        screen.getByTestId(workbookPresenceSummaryTestId()).textContent,
      ).not.toContain("OA");
    });
  });

  it("applies sparse live patches by record_id without moving an active cell draft", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Local base",
            details: "Old details",
            captureState: "rough",
          }),
        ],
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    input.focus();
    await changeInputValue(input, "Unsaved local");
    latestTimelineWebSocket()?.emit({
      type: "record_changed",
      stream_seq: 1,
      payload: buildRecordChangedPayload({
        recordId: "20000000-0000-4000-8000-000000000001",
        rowVersion: 2,
        clientTxnId: "remote-patch",
        changedFieldKeys: ["timeline.raw_activity_text"],
        affectedViews: [
          {
            view_schema_id: timelineViewSchemaId,
            change_kind: "patch",
            patch_cells: {
              record_id: "20000000-0000-4000-8000-000000000001",
              row_version: 2,
              cells: {
                "timeline.raw_activity_text": { value: "Remote details" },
              },
            },
          },
        ],
      }),
    });

    await waitFor(() => {
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000001"),
        ).textContent,
      ).toBe("2");
    });
    expect(input.value).toBe("Unsaved local");
    expect(document.activeElement).toBe(input);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("keeps the grid visible, conflict unresolved, and focus bound to the same cell", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
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
        record_id: "20000000-0000-4000-8000-000000000001",
        field_key: "timeline.activity_synopsis_text",
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

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Unsaved local value");
    fireEvent.blur(input);

    expect(
      await screen.findByTestId(workbookConflictResolverTestId()),
    ).toBeTruthy();
    expect(
      screen.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeTruthy();
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Conflict");
    expect(
      screen
        .getByTestId(workbookConflictResolverTestId())
        .getAttribute("data-conflict-field-key"),
    ).toBe("timeline.activity_synopsis_text");
    expect(
      screen.getByTestId(workbookConflictSavedValueTestId()),
    ).toHaveProperty("value", "Server value");
    expect(
      screen.getByTestId(workbookConflictLocalValueTestId()),
    ).toHaveProperty("value", "Unsaved local value");

    fireEvent.keyDown(screen.getByTestId(workbookConflictSummaryTestId()), {
      key: "Enter",
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);

    fireEvent.click(screen.getByTestId(workbookConflictControlTestId("close")));
    await waitFor(() => {
      expect(screen.queryByTestId(workbookConflictResolverTestId())).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(document.activeElement).toBe(input);
      expect(input).toHaveProperty("value", "Unsaved local value");
    });
  });

  it("resolver keep-saved action submits an explicit outcome and applies returned row state", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
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
        record_id: "20000000-0000-4000-8000-000000000001",
        field_key: "timeline.activity_synopsis_text",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "Server",
          captureState: "rough",
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Local");
    fireEvent.blur(input);
    await screen.findByTestId(workbookConflictResolverTestId());
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("keep-saved")),
    );
    await waitForTimelineConflictResolutionCalls(fetchMock, 1);

    await waitFor(() => {
      expect(screen.queryByTestId(workbookConflictResolverTestId())).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000001"),
        ).textContent,
      ).toBe("2");
    });
    expect(extractTimelineConflictResolutionBody(fetchMock, 0)).toEqual({
      conflict_token: "conflict-token-keep",
      resolution_kind: "keep_saved",
      client_txn_id: expect.any(String),
    });
  });

  it("resolver use-unsaved action submits an explicit outcome and applies returned row state", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
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
        record_id: "20000000-0000-4000-8000-000000000001",
        field_key: "timeline.activity_synopsis_text",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 5,
          summary: "Use local",
          captureState: "enriched",
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Use local");
    fireEvent.blur(input);
    await screen.findByTestId(workbookConflictResolverTestId());
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("use-unsaved")),
    );
    await waitForTimelineConflictResolutionCalls(fetchMock, 1);

    await waitFor(() => {
      expect(screen.queryByTestId(workbookConflictResolverTestId())).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000001"),
        ).textContent,
      ).toBe("5");
    });
    expect(extractTimelineConflictResolutionBody(fetchMock, 0)).toEqual({
      conflict_token: "conflict-token-use",
      resolution_kind: "use_unsaved",
      client_txn_id: expect.any(String),
      resolved_value: "Use local",
    });
  });

  it("resolver merged-value action submits an explicit outcome and applies returned row state", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
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
        record_id: "20000000-0000-4000-8000-000000000001",
        field_key: "timeline.activity_synopsis_text",
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
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 8,
          summary: "Merge final",
          captureState: "enriched",
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.focus(input);
    await changeInputValue(input, "Merge local");
    fireEvent.blur(input);
    await screen.findByTestId(workbookConflictResolverTestId());
    expect(
      screen.getByTestId(workbookConflictControlTestId("merged-value")),
    ).toHaveProperty("value", "Merge server");
    fireEvent.change(
      screen.getByTestId(workbookConflictControlTestId("merged-value")),
      {
        target: { value: "Merge final" },
      },
    );
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("use-merged")),
    );
    await waitForTimelineConflictResolutionCalls(fetchMock, 1);

    await waitFor(() => {
      expect(screen.queryByTestId(workbookConflictResolverTestId())).toBeNull();
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000001"),
        ).textContent,
      ).toBe("8");
    });
    expect(extractTimelineConflictResolutionBody(fetchMock, 0)).toEqual({
      conflict_token: "conflict-token-merged",
      resolution_kind: "merged_value",
      client_txn_id: expect.any(String),
      resolved_value: "Merge final",
    });
  });

  it("keeps save-state labels and pending queue replay bounded and explicit", async () => {
    const firstPendingPatch = deferred<Response>();
    const secondPendingPatch = deferred<Response>();
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "One",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 1,
            summary: "Two",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(firstPendingPatch.promise);
    routedFetch.mockRecordPatchOnce(secondPendingPatch.promise);

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    const firstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.keyDown(
      await changeQueuedCellValue(firstInput, "One in flight"),
      { key: "Enter" },
    );
    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Syncing");
    });

    fireEvent.blur(await changeQueuedCellValue(firstInput, "One queued final"));
    await waitForPendingQueueState({
      expectedPendingUnits: 2,
      expectedSaveState: "Syncing",
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toHaveLength(1);

    firstPendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "One in flight",
          captureState: "rough",
        }),
      }),
    );

    await waitForTimelineRecordPatchCalls(fetchMock, 2);
    expect(extractTimelineRecordPatchBody(fetchMock, 1).changes).toEqual([
      {
        field_key: "timeline.activity_synopsis_text",
        value: "One queued final",
      },
    ]);

    secondPendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 3,
          summary: "One queued final",
          captureState: "rough",
        }),
      }),
    );

    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
  });

  it("does not coalesce non-contiguous same-record pending patches", () => {
    const runtime = new WorkbookMutationRuntime(
      {
        clientInstanceId: "client-non-contiguous",
        incidentId: "10000000-0000-4000-8000-000000000001",
      },
      createBrowserSecureTransactionIdPort(),
    );
    runtime.pauseForAuthRecovery();
    for (const [recordId, value, version] of [
      ["20000000-0000-4000-8000-000000000001", "A1 queued", 1],
      ["20000000-0000-4000-8000-000000000002", "B1 queued", 1],
      ["20000000-0000-4000-8000-000000000001", "A2 queued", 2],
    ] as const) {
      expect(
        runtime.enqueuePatch({
          baseRowVersion: version,
          changes: [
            {
              field_key: "timeline.activity_synopsis_text",
              value,
            },
          ],
          fieldKey: "timeline.activity_synopsis_text",
          localValue: value,
          recordId,
          rowLabel: recordId,
          surfaceLabel: "Timeline",
          viewSchemaId: timelineViewSchemaId,
        }),
      ).toEqual({ kind: "accepted" });
    }
    expect(
      runtime
        .pending()
        .model.snapshot()
        .units.map((unit) => unit.recordId),
    ).toEqual([
      "20000000-0000-4000-8000-000000000001",
      "20000000-0000-4000-8000-000000000002",
      "20000000-0000-4000-8000-000000000001",
    ]);
  });

  it("exposes the browser-runtime pending queue capacity as exactly 64 replay units", () => {
    expect(pendingReplayCapacity).toBe(64);
  });

  it("clears protected rows and blocks new edits after session revocation", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Auth base",
            captureState: "rough",
          }),
        ],
      }),
    );
    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    await waitFor(() => {
      expect(latestTimelineWebSocket()).not.toBeNull();
    });
    latestTimelineWebSocket()?.emit({
      type: "session_revoked",
      payload: { reason_code: "session_revoked" },
    });

    await waitFor(() => {
      expect(input.isConnected).toBe(false);
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toEqual([]);
    expect(
      screen.queryByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: "20000000-0000-4000-8000-000000000001",
          surface: "grid",
        }),
      ),
    ).toBeNull();
  });

  it("moves a blocking same-field conflict out of the pending queue and retains its editor", async () => {
    const firstPendingPatch = deferred<Response>();
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "One",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 1,
            summary: "Two",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(firstPendingPatch.promise);

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const firstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.blur(await changeQueuedCellValue(firstInput, "Conflict local"));
    await waitForTimelineRecordPatchCalls(fetchMock, 1);

    firstPendingPatch.resolve(
      errorEnvelope("same_field_conflict", 409, {
        conflict_token: "conflict-token-queued",
        record_id: "20000000-0000-4000-8000-000000000001",
        field_key: "timeline.activity_synopsis_text",
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
      expect(screen.getByTestId(workbookConflictResolverTestId())).toBeTruthy();
      expect(firstInput.value).toBe("Conflict local");
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toHaveLength(1);
  });

  it("drives WorkbookShell admission, coalescing, and save-state from the shared pending queue model", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Queue 1",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
            rowVersion: 1,
            summary: "Queue 2",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(errorEnvelope("session_required", 401));

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const firstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.blur(await changeQueuedCellValue(firstInput, "Queue 1 local"));
    await waitForTimelineRecordPatchCalls(fetchMock, 1);
    await waitForPendingQueueState({
      expectedPendingUnits: 1,
      expectedSaveState: "Syncing",
      noticeIncludes: "Authentication is required",
    });

    expect(timelineRecordPatchCallURLs(fetchMock)).toHaveLength(1);
    expect(firstInput.value).toBe("Queue 1 local");
  });

  it("drives WorkbookShell retry and success settlement from the shared pending queue model", async () => {
    const retrySuccess = deferred<Response>();
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000001",
            rowVersion: 1,
            summary: "Retry one",
            captureState: "rough",
          }),
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000002",
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
    routedFetch.mockRecordPatchOnce(retrySuccess.promise);
    routedFetch.mockRecordPatchOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000002",
          rowVersion: 2,
          summary: "Retry behind",
          captureState: "rough",
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const retryFirstInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000001",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.blur(await changeQueuedCellValue(retryFirstInput, "Retry head"));
    await waitForTimelineRecordPatchCalls(fetchMock, 2);
    await waitFor(() => {
      expect(
        screen.queryByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.activity_synopsis_text",
            recordId: "20000000-0000-4000-8000-000000000001",
            surface: "grid",
          }),
        ),
      ).toBeNull();
    });
    const retrySecondInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000002",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.blur(
      await changeQueuedCellValue(retrySecondInput, "Retry behind"),
    );
    await waitForPendingQueueState({
      expectedPendingUnits: 2,
      expectedSaveState: "Syncing",
    });
    retrySuccess.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000001",
          rowVersion: 2,
          summary: "Retry head",
          captureState: "rough",
        }),
      }),
    );
    await waitForTimelineRecordPatchCalls(fetchMock, 3);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toEqual([
      "/api/v1/records/20000000-0000-4000-8000-000000000001",
      "/api/v1/records/20000000-0000-4000-8000-000000000001",
      "/api/v1/records/20000000-0000-4000-8000-000000000002",
    ]);
  });

  it("drives WorkbookShell terminal halt presentation from the shared pending queue model", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000102",
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
    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const haltInput = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000102",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.blur(await changeQueuedCellValue(haltInput, "Halt local"));
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
      expect(
        screen.getByTestId(pendingQueueCountTestId()).textContent,
      ).toContain("1");
      expect(
        screen.getByTestId(pendingQueueNoticeTestId()).textContent,
      ).toContain("future_terminal_public_error");
    });
  });

  it("recovers a Timeline client transaction conflict with a fresh opaque request ID", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000103",
            rowVersion: 1,
            summary: "Retry base",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(errorEnvelope("client_txn_conflict", 409));
    routedFetch.mockRecordPatchOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000103",
          rowVersion: 2,
          summary: "Retry local",
          captureState: "rough",
        }),
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000103",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.blur(await changeQueuedCellValue(input, "Retry local"));

    const recoveryPanel = await screen.findByTestId(
      workbookEditRecoveryTestId(),
    );
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Conflict");
    expect(
      screen.getByRole("button", { name: "Retry with a new request ID" }),
    ).toBe(screen.getByTestId(workbookEditRecoveryRetryButtonTestId()));
    expect(screen.getByRole("button", { name: "Discard blocked edit" })).toBe(
      screen.getByTestId(workbookEditRecoveryDiscardButtonTestId()),
    );
    expect(recoveryPanel.textContent).not.toContain("timeline-client-");

    fireEvent.click(screen.getByTestId(saveStateActionButtonTestId()));
    expect(document.activeElement).toBe(
      screen.getByTestId(pendingQueueNoticeTestId()),
    );
    fireEvent.click(
      screen.getByTestId(workbookEditRecoveryRetryButtonTestId()),
    );
    await waitForTimelineRecordPatchCalls(fetchMock, 2);
    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });

    const first = extractTimelineRecordPatchBody(fetchMock, 0) as ReturnType<
      typeof extractTimelineRecordPatchBody
    > & { client_txn_id: string };
    const retried = extractTimelineRecordPatchBody(fetchMock, 1) as ReturnType<
      typeof extractTimelineRecordPatchBody
    > & { client_txn_id: string };
    expect(first.client_txn_id).toMatch(/^timeline-client-[0-9a-f-]{36}$/u);
    expect(retried.client_txn_id).toMatch(/^workbook-recovery-[0-9a-f-]{36}$/u);
    expect(retried.client_txn_id).not.toBe(first.client_txn_id);
    expect(retried.base_row_version).toBe(1);
    expect(retried.changes).toEqual(first.changes);
  });

  it("routes a same-field conflict from transaction recovery into the normal resolver", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000104",
            rowVersion: 3,
            summary: "Resolver base",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(errorEnvelope("client_txn_conflict", 409));
    routedFetch.mockRecordPatchOnce(
      errorEnvelope("same_field_conflict", 409, {
        conflict_token: "conflict-token-retry",
        record_id: "20000000-0000-4000-8000-000000000104",
        field_key: "timeline.activity_synopsis_text",
        conflict_resolution_class: "text_compare_merge",
        base_row_version: 3,
        current_row_version: 4,
        client_value: "Resolver local",
        server_value: "Resolver saved",
        base_value: "Resolver base",
      }),
    );

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000104",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.blur(await changeQueuedCellValue(input, "Resolver local"));
    fireEvent.click(
      await screen.findByTestId(workbookEditRecoveryRetryButtonTestId()),
    );

    expect(
      await screen.findByTestId(workbookConflictResolverTestId()),
    ).toBeTruthy();
    expect(screen.queryByTestId(workbookEditRecoveryTestId())).toBeNull();
    expect(
      (
        screen.getByTestId(
          workbookConflictLocalValueTestId(),
        ) as HTMLInputElement
      ).value,
    ).toBe("Resolver local");
    expect(
      (
        screen.getByTestId(
          workbookConflictSavedValueTestId(),
        ) as HTMLInputElement
      ).value,
    ).toBe("Resolver saved");
  });

  it("discards a blocked edit locally without issuing another mutation", async () => {
    const routedFetch = routeTimelineWorkbookFetchMock(fetchMock);
    routedFetch.mockRowQueryOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000101",
            rowVersion: 1,
            summary: "Discard base",
            captureState: "rough",
          }),
        ],
      }),
    );
    routedFetch.mockRecordPatchOnce(errorEnvelope("client_txn_conflict", 409));

    render(
      <TimelineWorkbookRuntimeFixture incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000101",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    fireEvent.blur(await changeQueuedCellValue(input, "Discard local"));
    fireEvent.click(
      await screen.findByTestId(workbookEditRecoveryDiscardButtonTestId()),
    );

    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
    expect(timelineRecordPatchCallURLs(fetchMock)).toEqual([
      "/api/v1/records/20000000-0000-4000-8000-000000000101",
    ]);
    const restored = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000101",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    expect(restored.value).toBe("Discard base");
  });
});
