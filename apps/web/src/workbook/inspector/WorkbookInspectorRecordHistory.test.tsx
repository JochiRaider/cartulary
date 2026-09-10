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
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { type ComponentProps, useMemo } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { WorkbookHistoryContext } from "../history/WorkbookHistoryContext";
import { WorkbookRecordHistoryOwner } from "../history/WorkbookRecordHistoryOwner";
import type { HistoryReceipt } from "../history/workbookHistoryOperation";

type RecordLifecycleAccepted = {
  readonly recordId: string;
  readonly rowVersion: number;
};

import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type {
  RecordHistoryData,
  RecordHistoryRollbackTarget,
} from "./workbookRecordHistoryModel";

interface RecordRouteCommandPort {
  execute(input: {
    readonly action: "delete" | "restore";
    readonly baseRowVersion: number;
    readonly reason: string;
    readonly recordId: string;
  }): Promise<WorkbookOperationOutcome<RecordLifecycleAccepted>>;
  loadHistory(input: {
    readonly recordId: string;
  }): Promise<WorkbookOperationOutcome<RecordHistoryData>>;
  rollback(input: {
    readonly baseRowVersion: number;
    readonly reason: string;
    readonly recordId: string;
    readonly target: RecordHistoryRollbackTarget;
  }): Promise<WorkbookOperationOutcome<RecordLifecycleAccepted>>;
}

import { WorkbookInspectorRecordHistory } from "./WorkbookInspectorRecordHistory";

afterEach(cleanup);

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
      <HistoryTestSubject
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          refresh: vi.fn(),
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
      (
        await screen.findByTestId(
          rowHistoryRollbackPreviewTestId({
            action: "history_entry",
            historyItemRef,
          }),
        )
      ).textContent,
    ).toContain(recordId);
    fireEvent.click(
      await screen.findByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );

    await waitFor(() => {
      expect(commands.rollback).toHaveBeenCalledWith({
        baseRowVersion: 5,
        reason: "Rollback from workbook history",
        recordId,
        target: {
          history_entry_ref: "server-history-selector",
          kind: "history_entry",
        },
      });
    });
    await waitFor(() =>
      expect(rollbackAccepted).toHaveBeenCalledWith(
        expect.objectContaining({ recordId, rowVersion: 6 }),
      ),
    );
    expect(screen.getByText("Reverse history entry completed.")).not.toBeNull();
    expect(commands.loadHistory).toHaveBeenCalledTimes(4);
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
      <HistoryTestSubject
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          refresh: vi.fn(),
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
      await screen.findByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      ),
    );
    await waitFor(() => expect(deleteAccepted).toHaveBeenCalledOnce());
    expect(commands.execute).toHaveBeenLastCalledWith({
      action: "delete",
      baseRowVersion: 5,
      reason: "Deleted from workbook history",
      recordId,
    });

    fireEvent.click(await screen.findByTestId(rowHistoryRestoreButtonTestId()));
    fireEvent.click(
      await screen.findByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "restore" }),
      ),
    );
    await waitFor(() => expect(restoreAccepted).toHaveBeenCalledOnce());
    expect(commands.execute).toHaveBeenLastCalledWith({
      action: "restore",
      baseRowVersion: 6,
      reason: "Restored from workbook history",
      recordId,
    });
    expect(screen.getByText("Restore deleted row completed.")).not.toBeNull();
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
      <HistoryTestSubject
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          refresh: vi.fn(),
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
      await screen.findByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    await waitFor(() => expect(commands.rollback).toHaveBeenCalledOnce());
    rerender(
      <HistoryTestSubject
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          refresh: vi.fn(),
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

    await waitFor(() => expect(commands.rollback).toHaveBeenCalledOnce());
    expect(rollbackAccepted).not.toHaveBeenCalled();
    expect(newerRollbackAccepted).not.toHaveBeenCalled();
    expect(screen.queryByText("Reverse history entry completed.")).toBeNull();
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
      <HistoryTestSubject
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          refresh: vi.fn(),
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
      await screen.findByTestId(
        rowHistoryRollbackCancelButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    await waitFor(() => expect(document.activeElement).toBe(action));

    fireEvent.click(action);
    fireEvent.keyDown(await screen.findByRole("alertdialog"), {
      key: "Escape",
    });
    await waitFor(() => expect(document.activeElement).toBe(action));

    fireEvent.click(action);
    fireEvent.click(
      await screen.findByTestId(
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
      .mockResolvedValue({ kind: "accepted", value: historyData() });
    const commands: RecordRouteCommandPort = {
      execute: vi.fn(),
      loadHistory,
      rollback: vi.fn(async () => ({
        kind: "accepted" as const,
        value: { recordId, rowVersion: 5 },
      })),
    };
    render(
      <HistoryTestSubject
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          refresh: vi.fn(),
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
      await screen.findByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: "history_entry",
          historyItemRef,
        }),
      ),
    );
    await waitFor(() => expect(loadHistory).toHaveBeenCalledTimes(4));
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
      refresh: vi.fn(),
      deleteAccepted: vi.fn(),
      restoreAccepted: vi.fn(),
      rollbackAccepted: vi.fn(),
    };
    const { rerender } = render(
      <HistoryTestSubject
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
      await screen.findByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      ),
    );
    await waitFor(() =>
      expect(document.activeElement).toBe(
        screen.getByTestId(rowHistoryPanelTestId()),
      ),
    );

    fireEvent.click(await screen.findByTestId(rowHistoryRestoreButtonTestId()));
    await screen.findByRole("alertdialog");
    rerender(
      <HistoryTestSubject
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
      <HistoryTestSubject
        beginMutation={() => vi.fn()}
        actions={new Set(["delete", "restore", "rollback"])}
        canMutate
        commands={commands}
        ownerEffects={{
          refresh: vi.fn(),
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
      await screen.findByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      ),
    );

    expect(
      await screen.findByText(
        "The outcome is unknown. Open History actions to recover this action.",
      ),
    ).not.toBeNull();
    expect(deleteAccepted).not.toHaveBeenCalled();
    expect(commands.loadHistory).toHaveBeenCalledTimes(3);
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

function HistoryTestSubject(
  props: Omit<
    ComponentProps<typeof WorkbookInspectorRecordHistory>,
    "commands"
  > & { commands: RecordRouteCommandPort },
) {
  const runtime = useMemo(() => {
    const history = new WorkbookRecordHistoryOwner(
      "10000000-0000-4000-8000-000000000001",
      { create: () => crypto.randomUUID() },
    );
    history.setAuthority({
      actorId: "reviewer",
      incidentId: history.incidentId,
      role: "reviewer",
      closed: false,
    });
    let accepted: HistoryReceipt | null = null;
    const port: import("../history/workbookHistoryOperation").WorkbookRecordHistoryPort =
      {
        load: async (recordId) => {
          const outcome = await props.commands.loadHistory({ recordId });
          if (
            outcome.kind === "accepted" &&
            accepted?.recordId === recordId &&
            accepted.rowVersion > outcome.value.row_version
          )
            return {
              kind: "accepted",
              value: {
                ...outcome.value,
                row_version: accepted.rowVersion,
                deleted: accepted.kind === "delete",
              },
            };
          return outcome;
        },
        send: async (attempt) => {
          const request = JSON.parse(attempt.body);
          const outcome =
            attempt.pending.kind === "rollback"
              ? await props.commands.rollback({
                  recordId: attempt.subject.recordId,
                  baseRowVersion: request.base_row_version,
                  reason: request.reason,
                  target: request.target,
                })
              : await props.commands.execute({
                  action: attempt.pending.operation,
                  recordId: attempt.subject.recordId,
                  baseRowVersion: request.base_row_version,
                  reason: request.reason,
                });
          if (outcome.kind === "rejected") return outcome;
          accepted = {
            ...outcome.value,
            incidentId: history.incidentId,
            changeSetId: "accepted-change-set",
            ...(attempt.pending.kind === "rollback"
              ? {
                  kind: "rollback" as const,
                  target: attempt.pending.target,
                  affectedRecordIds: [attempt.subject.recordId],
                }
              : {
                  kind: attempt.pending.operation,
                  deleted: attempt.operation === "delete",
                  deletedAt:
                    attempt.operation === "delete"
                      ? "2026-09-10T00:00:00Z"
                      : null,
                  deletedByUserId:
                    attempt.operation === "delete" ? "reviewer" : null,
                }),
          };
          return { kind: "acknowledged", receipt: accepted };
        },
      };
    history.configure(port);
    return {
      history,
      port,
      coordinateHistory: async (recordId: string) =>
        history.latestVersion(recordId),
    };
  }, [props.commands]);
  return (
    <WorkbookHistoryContext.Provider value={runtime}>
      <WorkbookInspectorRecordHistory {...props} commands={runtime.port} />
    </WorkbookHistoryContext.Provider>
  );
}
