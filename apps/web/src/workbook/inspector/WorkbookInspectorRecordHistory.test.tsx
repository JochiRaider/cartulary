import {
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryItemTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackCancelButtonTestId,
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
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted,
        }}
        subject={historySubject(recordId, 5)}
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
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted,
          restoreAccepted,
          rollbackAccepted: vi.fn(),
        }}
        subject={historySubject(recordId, 5)}
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
    const newerRollbackAccepted = vi.fn(async () => undefined);
    const { rerender } = render(
      <WorkbookInspectorRecordHistory
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted,
        }}
        subject={historySubject(recordId, 5)}
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
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted: newerRollbackAccepted,
        }}
        subject={historySubject("record-b", 1)}
      />,
    );
    pendingRollback.resolve({
      kind: "accepted",
      value: { recordId, rowVersion: 6 },
    });

    await waitFor(() => expect(rollbackAccepted).toHaveBeenCalledOnce());
    expect(newerRollbackAccepted).not.toHaveBeenCalled();
    expect(screen.queryByText(`Rolled back record ${recordId}.`)).toBeNull();
    expect(screen.getByText("record-b")).not.toBeNull();
  });

  it("returns cancel, Escape, and rejected submissions to the same semantic action", async () => {
    const commands: RecordRouteCommandPort = {
      execute: vi.fn(),
      loadHistory: vi.fn(async () => ({
        kind: "accepted" as const,
        value: historyData(),
      })),
      rollback: vi.fn(async () => ({
        failure: { kind: "retryable" as const, message: "Rollback rejected" },
        kind: "rejected" as const,
      })),
    };
    render(
      <WorkbookInspectorRecordHistory
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted: vi.fn(),
        }}
        subject={historySubject(recordId, 5)}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Open history" }));
    const actionTestId = rowHistoryActionTestId({
      action: "history_entry",
      historyItemRef,
    });
    const action = await screen.findByTestId(actionTestId);

    fireEvent.click(action);
    fireEvent.click(
      screen.getByTestId(
        rowHistoryRollbackCancelButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    await waitFor(() => expect(document.activeElement).toBe(action));

    fireEvent.click(action);
    fireEvent.keyDown(screen.getByRole("alertdialog"), { key: "Escape" });
    await waitFor(() => expect(document.activeElement).toBe(action));

    fireEvent.click(action);
    fireEvent.click(
      screen.getByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    await screen.findByText("Rollback rejected");
    await waitFor(() => expect(document.activeElement).toBe(action));
  });

  it("restores successful rollback focus only when the exact action identity survives", async () => {
    const loadHistory = vi
      .fn()
      .mockResolvedValueOnce({ kind: "accepted", value: historyData() })
      .mockResolvedValueOnce({ kind: "accepted", value: historyData() });
    const commands: RecordRouteCommandPort = {
      execute: vi.fn(),
      loadHistory,
      rollback: vi.fn(async () => ({
        kind: "accepted" as const,
        value: { recordId, rowVersion: 5 },
      })),
    };
    render(
      <WorkbookInspectorRecordHistory
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted: vi.fn(),
          restoreAccepted: vi.fn(),
          rollbackAccepted: vi.fn(),
        }}
        subject={historySubject(recordId, 5)}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Open history" }));
    const actionTestId = rowHistoryActionTestId({
      action: "history_entry",
      historyItemRef,
    });
    fireEvent.click(await screen.findByTestId(actionTestId));
    fireEvent.click(
      screen.getByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    await waitFor(() => expect(loadHistory).toHaveBeenCalledTimes(2));
    await waitFor(() =>
      expect(document.activeElement).toBe(screen.getByTestId(actionTestId)),
    );
  });

  it("uses History as the success fallback and discards invalidated focus requests", async () => {
    let deleted = false;
    let rowVersion = 5;
    const commands: RecordRouteCommandPort = {
      execute: vi.fn(async () => {
        deleted = true;
        rowVersion = 6;
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
    const ownerEffects = {
      deleteAccepted: vi.fn(),
      restoreAccepted: vi.fn(),
      rollbackAccepted: vi.fn(),
    };
    const { rerender } = render(
      <WorkbookInspectorRecordHistory
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={ownerEffects}
        subject={historySubject(recordId, 5)}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Open history" }));
    fireEvent.click(await screen.findByTestId(rowHistoryDeleteButtonTestId()));
    fireEvent.click(
      screen.getByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      ),
    );
    await waitFor(() =>
      expect(document.activeElement).toBe(
        screen.getByTestId(rowHistoryPanelTestId()),
      ),
    );

    fireEvent.click(await screen.findByTestId(rowHistoryRestoreButtonTestId()));
    rerender(
      <WorkbookInspectorRecordHistory
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate={false}
        commands={commands}
        ownerEffects={ownerEffects}
        subject={{
          ...historySubject(recordId, 6),
          kind: "deleted",
          stateLabel: "Deleted",
        }}
      />,
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId(
          rowHistoryDestructiveCancelButtonTestId({ operation: "restore" }),
        ),
      ).toBeNull(),
    );
    expect(document.activeElement).not.toBe(
      screen.getByTestId(rowHistoryPanelTestId()),
    );
  });

  it("rejects a server acceptance for the wrong record before owner effects", async () => {
    const deleteAccepted = vi.fn();
    const commands: RecordRouteCommandPort = {
      execute: vi.fn(async () => ({
        kind: "accepted" as const,
        value: { recordId: "record-b", rowVersion: 6 },
      })),
      loadHistory: vi.fn(async () => ({
        kind: "accepted" as const,
        value: historyData(),
      })),
      rollback: vi.fn(),
    };
    render(
      <WorkbookInspectorRecordHistory
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          deleteAccepted,
          restoreAccepted: vi.fn(),
          rollbackAccepted: vi.fn(),
        }}
        subject={historySubject(recordId, 5)}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Open history" }));
    fireEvent.click(await screen.findByTestId(rowHistoryDeleteButtonTestId()));
    fireEvent.click(
      screen.getByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      ),
    );

    expect(
      await screen.findByText(
        "The history operation returned an invalid record identity.",
      ),
    ).not.toBeNull();
    expect(deleteAccepted).not.toHaveBeenCalled();
    expect(commands.loadHistory).toHaveBeenCalledOnce();
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

function historySubject(subjectRecordId: string, rowVersion: number) {
  return {
    kind: "live" as const,
    label: "Timeline row",
    recordId: subjectRecordId,
    rowVersion,
    surfaceLabel: "Timeline",
    viewSchemaId: "cartulary.view.timeline.v2",
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((next) => {
    resolve = next;
  });
  return { promise, resolve };
}
