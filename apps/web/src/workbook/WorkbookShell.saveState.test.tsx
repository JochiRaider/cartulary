import {
  gridShellTestId,
  pendingQueueNoticeTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { renderTimelineWorkbook } from "../testing/timelineWorkbookRenderTestSupport";
import {
  cleanupTimelineWorkbookTestGlobals,
  findWorkbookCell,
  installTimelineWorkbookTestGlobals,
  latestTimelineWebSocket,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

describe("WorkbookShell save-state status strip", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("renders derived secondary save-state detail inside the workbook status strip", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "20000000-0000-4000-8000-000000000501",
            rowVersion: 1,
            summary: "Auth base",
            captureState: "rough",
          }),
        ],
      }),
    );
    const pendingPatch = deferred<Response>();
    fetchMock.mockReturnValueOnce(pendingPatch.promise);

    renderTimelineWorkbook();
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000501",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    const workbookShell = screen.getByTestId(
      timelineMutationSubstrateReadyTestId(),
    );
    expect(workbookShell.style.gridTemplateRows).toBe(
      "var(--ct-layout-viewBarHeight) minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
    );

    const statusStrip = screen.getByTestId(
      workbookShellSlotTestId("status-strip"),
    );
    expect(
      screen.getByTestId(workbookShellSlotTestId("view-bar")).style.gridRow,
    ).toBe("1");
    expect(
      screen.getByTestId(workbookShellSlotTestId("primary-grid")).parentElement
        ?.style.gridRow,
    ).toBe("2");
    expect(statusStrip.style.gridRow).toBe("3");
    expect(statusStrip.style.overflow).toBe("hidden");
    expect(statusStrip.style.minBlockSize).toBe(
      "var(--ct-layout-statusStripHeight)",
    );
    expect(
      screen.getByTestId(workbookShellSlotTestId("primary-grid")).style
        .overflow,
    ).toBe("hidden");
    expect(
      screen.getByTestId(gridShellTestId(timelineViewSchemaId)).style.blockSize,
    ).toBe("100%");
    expect(within(statusStrip).getByTestId(saveStateTestId()).textContent).toBe(
      "Saved",
    );
    expect(
      within(statusStrip).queryByText(
        "Authentication is required before queued edits can replay.",
      ),
    ).toBeNull();

    await waitFor(() => expect(latestTimelineWebSocket()).not.toBeNull());
    const socket = latestTimelineWebSocket();
    await waitFor(() => {
      expect(socket?.sentMessages.length).toBeGreaterThan(0);
    });
    socket?.emit({
      type: "hello_ack",
      payload: {
        connection_id: "self-connection",
        resume_token: "resume-save-state",
        server_time: "2026-07-13T12:00:00Z",
        heartbeat_interval_ms: 15_000,
        presence_ttl_ms: 45_000,
        resume_window_ms: 300_000,
      },
    });
    socket?.emit({
      type: "presence_snapshot",
      payload: {
        presences: [
          {
            connection_id: "other-connection",
            user_id: "other-user",
            display_name: "Other Analyst",
            sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
            record_id: "20000000-0000-4000-8000-000000000501",
            mode: "viewing",
            field_key: "timeline.activity_synopsis_text",
            observed_at: "2026-06-02T12:00:00Z",
            expires_at: "2026-06-02T12:01:00Z",
          },
        ],
      },
    });
    await waitFor(() => {
      expect(within(statusStrip).getByText("OA")).toBeTruthy();
    });
    expect(within(statusStrip).getByTestId(saveStateTestId()).textContent).toBe(
      "Saved",
    );

    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "Auth queued" } });
    fireEvent.blur(input);
    await waitFor(() => {
      expect(
        within(statusStrip).getByTestId(saveStateTestId()).textContent,
      ).toBe("Syncing");
    });

    socket?.emit({
      type: "session_revoked",
      payload: { reason_code: "session_revoked" },
    });

    await waitFor(() => {
      expect(
        within(statusStrip).getByTestId(saveStateTestId()).textContent,
      ).toBe("Syncing");
      expect(
        within(statusStrip).getByText(
          "Authentication is required before queued edits can replay.",
        ),
      ).toBeTruthy();
      expect(
        screen.getByTestId(pendingQueueNoticeTestId()).textContent,
      ).toContain("Authentication is required before queued edits can replay.");
    });
    const statusDetail = within(statusStrip).getByText(
      "Authentication is required before queued edits can replay.",
    );
    expect(statusDetail.style.display).toBe("block");
    expect(statusDetail.style.overflow).toBe("hidden");
    expect(statusDetail.style.textOverflow).toBe("ellipsis");
    expect(statusDetail.style.whiteSpace).toBe("nowrap");
    const pendingNotice = screen.getByTestId(pendingQueueNoticeTestId());
    expect(pendingNotice.style.display).toBe("flex");
    expect(pendingNotice.style.alignItems).toBe("center");
    expect(pendingNotice.style.overflow).toBe("hidden");
    expect(pendingNotice.parentElement?.style.position).toBe("absolute");
    expect(pendingNotice.parentElement?.style.maxBlockSize).toBe(
      "min(14rem, 32vh)",
    );
    pendingPatch.resolve(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000501",
        row: timelineRow({
          recordId: "20000000-0000-4000-8000-000000000501",
          rowVersion: 2,
          summary: "Auth queued",
          captureState: "rough",
        }),
      }),
    );
    expect(pendingNotice.parentElement?.style.overflowY).toBe("auto");
    expect(pendingNotice.parentElement?.style.pointerEvents).toBe("none");
    expect(pendingNotice.style.pointerEvents).toBe("none");
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
