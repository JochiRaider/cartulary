import {
  rowHistoryActionTestId,
  rowHistoryItemTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
} from "@cartulary/ui-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { RecordRouteCommandPort } from "../mutations/workbookMutationCommandPorts";
import { WorkbookInspectorRecordHistory } from "./WorkbookInspectorRecordHistory";

const historyItemRef = "history-item-1";
const recordId = "20000000-0000-4000-8000-000000000001";

describe("WorkbookInspectorRecordHistory", () => {
  it("loads advertised history and rolls back only through its stable selector", async () => {
    const commands: RecordRouteCommandPort = {
      execute: vi.fn(),
      loadHistory: vi.fn(async () => ({
        kind: "accepted" as const,
        value: {
          deleted: false,
          incident_id: "10000000-0000-4000-8000-000000000001",
          items: [
            {
              actor_user_id: "40000000-0000-4000-8000-000000000001",
              available_rollback_actions: ["history_entry" as const],
              change_set_id: "30000000-0000-4000-8000-000000000001",
              committed_at: "2026-08-30T20:00:00Z",
              diff_summary: { summary: "Changed title", units: [] },
              history_entry_ref: "server-history-selector",
              history_item_ref: historyItemRef,
              operation: "patch",
              reversible: true,
            },
          ],
          record_id: recordId,
          row_version: 5,
        },
      })),
      rollback: vi.fn(async () => ({
        kind: "accepted" as const,
        value: { recordId, rowVersion: 6 },
      })),
    };
    const onMessage = vi.fn();
    const onRefresh = vi.fn(async () => undefined);
    render(
      <WorkbookInspectorRecordHistory
        canMutate
        commands={commands}
        subject={{ recordId, rowVersion: 5 }}
        onMessage={onMessage}
        onRefresh={onRefresh}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Open history" }));
    const historyItem = await screen.findByTestId(
      rowHistoryItemTestId({ historyItemRef }),
    );
    expect(historyItem.textContent).not.toContain(
      "Changed by 40000000-0000-4000-8000-000000000001",
    );
    expect(historyItem.textContent).toContain("Actor ID");
    fireEvent.click(
      screen.getByTestId(
        rowHistoryActionTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    expect(
      screen.getByTestId(
        rowHistoryRollbackPreviewTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ).textContent,
    ).toContain(recordId);
    fireEvent.click(
      screen.getByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );

    await waitFor(() => {
      expect(commands.rollback).toHaveBeenCalledWith({
        baseRowVersion: 5,
        reason: "Rollback history_entry from the workbook inspector",
        recordId,
        target: {
          history_entry_ref: "server-history-selector",
          kind: "history_entry",
        },
      });
    });
    expect(onRefresh).toHaveBeenCalledOnce();
    expect(onMessage).toHaveBeenCalledWith(`Rolled back record ${recordId}.`);
    expect(commands.loadHistory).toHaveBeenCalledTimes(2);
  });
});
