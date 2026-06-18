import {
  gridShellTestId,
  pendingQueueNoticeTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanupTimelineWorkbookTestGlobals,
  findWorkbookCell,
  installTimelineWorkbookTestGlobals,
  latestTimelineWebSocket,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  timelineViewSchemaId,
} from "../testing/timelineWorkbookTestSupport";
import { TimelineWorkbook } from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

describe("FE-U-P4-02 WorkbookShell save-state status strip", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it("FE-U-P4-02 renders derived secondary save-state detail inside the workbook status strip", async () => {
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            recordId: "record-auth",
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
          recordId: "record-auth",
          rowVersion: 2,
          summary: "Auth queued",
          captureState: "rough",
        }),
      }),
    );

    render(<TimelineWorkbook incidentId="incident-1" />);
    const input = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "record-auth",
      "timeline.summary",
    )) as HTMLInputElement;
    const workbookShell = screen.getByTestId(
      timelineMutationSubstrateReadyTestId(),
    );
    expect(workbookShell.style.gridTemplateRows).toBe(
      "auto var(--ct-layout-viewBarHeight) minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
    );

    const statusStrip = screen.getByTestId(
      workbookShellSlotTestId("status-strip"),
    );
    expect(
      screen.getByTestId(workbookShellSlotTestId("view-bar")).style.gridRow,
    ).toBe("2");
    expect(
      screen.getByTestId(workbookShellSlotTestId("primary-grid")).parentElement
        ?.style.gridRow,
    ).toBe("3");
    expect(statusStrip.style.gridRow).toBe("4");
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
            record_id: "record-auth",
            mode: "viewing",
            field_key: "timeline.summary",
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

    socket?.emit({
      type: "session_revoked",
      payload: { reason_code: "session_revoked" },
    });
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "Auth queued" } });
    fireEvent.blur(input);

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
    expect(pendingNotice.parentElement?.style.overflowY).toBe("auto");
    expect(pendingNotice.parentElement?.style.pointerEvents).toBe("none");
    expect(pendingNotice.style.pointerEvents).toBe("auto");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
