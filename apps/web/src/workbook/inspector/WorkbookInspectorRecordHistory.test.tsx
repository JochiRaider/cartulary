import {
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryItemTestId,
  rowHistoryRestoreButtonTestId,
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
    const rollbackAccepted = vi.fn(async () => undefined);
    render(
      <WorkbookInspectorRecordHistory
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted,
        }}
        subject={{ kind: "live", recordId, rowVersion: 5 }}
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
    expect(rollbackAccepted).toHaveBeenCalledWith({
      recordId,
      rowVersion: 6,
    });
    expect(screen.getByText(`Rolled back record ${recordId}.`)).not.toBeNull();
    expect(commands.loadHistory).toHaveBeenCalledTimes(2);
  });

  it("retains a tombstone version for delete and restores from that exact version", async () => {
    let deleted = false;
    let rowVersion = 5;
    const commands: RecordRouteCommandPort = {
      execute: vi.fn(async ({ action }) => {
        deleted = action === "delete";
        rowVersion += 1;
        return {
          kind: "accepted" as const,
          value: { recordId, rowVersion },
        };
      }),
      loadHistory: vi.fn(async () => ({
        kind: "accepted" as const,
        value: historyData({ deleted, rowVersion }),
      })),
      rollback: vi.fn(),
    };
    const deleteAccepted = vi.fn();
    const restoreAccepted = vi.fn();
    render(
      <WorkbookInspectorRecordHistory
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted,
          restoreAccepted,
          rollbackAccepted: vi.fn(),
        }}
        subject={{ kind: "live", recordId, rowVersion: 5 }}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Open history" }));
    fireEvent.click(await screen.findByTestId(rowHistoryDeleteButtonTestId()));
    fireEvent.click(
      screen.getByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      ),
    );
    await waitFor(() => expect(deleteAccepted).toHaveBeenCalledOnce());
    expect(commands.execute).toHaveBeenLastCalledWith({
      action: "delete",
      baseRowVersion: 5,
      reason: "Deleted from the workbook inspector",
      recordId,
    });

    fireEvent.click(await screen.findByTestId(rowHistoryRestoreButtonTestId()));
    fireEvent.click(
      screen.getByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "restore" }),
      ),
    );
    await waitFor(() => expect(restoreAccepted).toHaveBeenCalledOnce());
    expect(commands.execute).toHaveBeenLastCalledWith({
      action: "restore",
      baseRowVersion: 6,
      reason: "Restored from the workbook inspector",
      recordId,
    });
    expect(screen.getByText(`Restored record ${recordId}.`)).not.toBeNull();
  });

  it("finishes captured owner effects without committing a stale mutation result", async () => {
    const pendingRollback = deferred<{
      readonly kind: "accepted";
      readonly value: {
        readonly recordId: string;
        readonly rowVersion: number;
      };
    }>();
    const commands: RecordRouteCommandPort = {
      execute: vi.fn(),
      loadHistory: vi.fn(async ({ recordId: activeRecordId }) => ({
        kind: "accepted" as const,
        value: historyData({ recordId: activeRecordId }),
      })),
      rollback: vi.fn(() => pendingRollback.promise),
    };
    const rollbackAccepted = vi.fn(async () => undefined);
    const { rerender } = render(
      <WorkbookInspectorRecordHistory
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted,
        }}
        subject={{ kind: "live", recordId, rowVersion: 5 }}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Open history" }));
    fireEvent.click(
      await screen.findByTestId(
        rowHistoryActionTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    fireEvent.click(
      screen.getByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    rerender(
      <WorkbookInspectorRecordHistory
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted,
        }}
        subject={{ kind: "live", recordId: "record-b", rowVersion: 1 }}
      />,
    );
    pendingRollback.resolve({
      kind: "accepted",
      value: { recordId, rowVersion: 6 },
    });

    await waitFor(() => expect(rollbackAccepted).toHaveBeenCalledOnce());
    expect(screen.queryByText(`Rolled back record ${recordId}.`)).toBeNull();
    expect(screen.getByText("record-b")).not.toBeNull();
  });
});

function historyData({
  deleted = false,
  recordId: historyRecordId = recordId,
  rowVersion = 5,
}: {
  readonly deleted?: boolean;
  readonly recordId?: string;
  readonly rowVersion?: number;
} = {}) {
  return {
    deleted,
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
    record_id: historyRecordId,
    row_version: rowVersion,
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((next) => {
    resolve = next;
  });
  return { promise, resolve };
}
