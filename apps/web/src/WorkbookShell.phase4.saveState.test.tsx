import {
  saveStateTestId,
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
} from "./timelineWorkbookTestSupport";
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

    const statusStrip = screen.getByTestId(
      workbookShellSlotTestId("status-strip"),
    );
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
      expect(screen.getByTestId("pending-queue-notice").textContent).toContain(
        "Authentication is required before queued edits can replay.",
      );
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
