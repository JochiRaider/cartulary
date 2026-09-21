import {
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { WorkbookRecoveryFixture } from "../../testing/WorkbookRecoveryFixture";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { WorkbookInspectorRecordHistory } from "../inspector/WorkbookInspectorRecordHistory";
import type { RecordHistoryData } from "../inspector/workbookRecordHistoryModel";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { WorkbookHistoryContext } from "./WorkbookHistoryContext";
import { WorkbookHistoryRecovery } from "./WorkbookHistoryRecovery";
import type {
  HistoryAttempt,
  HistoryReceipt,
  WorkbookRecordHistoryPort,
} from "./workbookHistoryOperation";

const operations = [
  "delete",
  "restore",
  "history_entry",
  "change_set",
  "row_restore",
] as const;
const surfaces = ["Timeline", "Generic", "Entity", "Assessment"] as const;

const item = {
  actor_user_id: "reviewer",
  committed_at: "2026-09-10T00:00:00Z",
  history_item_ref: "opaque-item",
  history_entry_ref: "opaque-entry",
  change_set_id: "original-change",
  revision_no: 2,
  reversible: true,
  operation: "patch",
  diff_summary: { summary: "Changed fields", units: [] },
  available_rollback_actions: [
    "history_entry",
    "change_set",
    "row_restore",
  ] as const,
};
const ids = {
  Timeline: "cartulary.view.timeline.v2",
  Generic: "cartulary.view.evidence.v2",
  Entity: "cartulary.view.entity.host.v2",
  Assessment: "cartulary.view.assessments.v2",
};
afterEach(cleanup);

function setup(
  surface: (typeof surfaces)[number],
  operation: (typeof operations)[number] = "delete",
) {
  const runtime = new WorkbookMutationRuntime(
    { incidentId: "incident", clientInstanceId: "client" },
    { create: () => crypto.randomUUID() },
    createWorkbookPendingMutationAdapter({
      apiBase: undefined,
      incidentId: "incident",
    }),
  );
  const owner = runtime.history;
  owner.setAuthority({
    actorId: "reviewer",
    incidentId: "incident",
    role: "reviewer",
    closed: false,
  });
  const subject = {
    kind: operation === "restore" ? ("deleted" as const) : ("live" as const),
    stateLabel: "Deleted",
    recordId: "record",
    rowVersion: 4,
    label: "Reviewed row",
    surfaceLabel: surface,
    viewSchemaId: ids[surface],
  };
  let history: RecordHistoryData = {
    record_id: "record",
    incident_id: "incident",
    row_version: 4,
    deleted: operation === "restore",
    items: [item],
  };
  const receipts = new Map<string, HistoryReceipt>();
  const counts = { mutations: 0, changeSets: 0, revisions: 0 };
  const commit = (attempt: HistoryAttempt): HistoryReceipt => {
    const existing = receipts.get(attempt.id);
    if (existing) return existing;
    counts.mutations++;
    counts.changeSets++;
    counts.revisions++;
    const receipt: HistoryReceipt = {
      recordId: "record",
      incidentId: "incident",
      rowVersion: 5,
      changeSetId: "accepted-change",
      ...(attempt.pending.kind === "rollback"
        ? {
            kind: "rollback" as const,
            target: attempt.pending.target,
            affectedRecordIds: ["other", "record"],
          }
        : {
            kind: attempt.pending.operation,
            deleted: attempt.operation === "delete",
            deletedAt:
              attempt.operation === "delete" ? "2026-09-10T00:00:00Z" : null,
            deletedByUserId: attempt.operation === "delete" ? "reviewer" : null,
          }),
    };
    receipts.set(attempt.id, receipt);
    history = {
      ...history,
      row_version: 5,
      deleted: attempt.operation === "delete",
    };
    return receipt;
  };
  let sends = 0;
  const send = vi.fn<WorkbookRecordHistoryPort["send"]>(async (attempt) => {
    const receipt = commit(attempt);
    return ++sends === 1
      ? { kind: "uncertain" }
      : { kind: "acknowledged", receipt };
  });
  const port: WorkbookRecordHistoryPort = {
    send,
    load: vi.fn(async () => ({
      kind: "accepted" as const,
      value: {
        ...history,
        paging: { limit: 100, has_more: false as const, next_cursor: null },
      },
    })),
  };
  owner.configure(port);
  const refresh = vi.fn(async () => {});
  owner.registerSurface(subject.viewSchemaId, refresh);
  const beginMutation = vi.fn(() => () => {});
  const effects = {
    deleteAccepted: vi.fn(),
    restoreAccepted: vi.fn(),
    rollbackAccepted: vi.fn(),
    refresh,
  };
  const view = (inspector: boolean) => (
    <WorkbookRecoveryFixture>
      <WorkbookHistoryContext.Provider value={runtime}>
        <input aria-label="Newer interaction" />
        <WorkbookHistoryRecovery />
        {inspector ? (
          <WorkbookInspectorRecordHistory
            beginMutation={beginMutation}
            commands={port}
            subject={subject}
            actions={new Set(["delete", "restore", "rollback"])}
            canMutate
            ownerEffects={effects}
          />
        ) : null}
      </WorkbookHistoryContext.Provider>
    </WorkbookRecoveryFixture>
  );
  return {
    runtime,
    owner,
    subject,
    send,
    port,
    commit,
    counts,
    refresh,
    effects,
    view,
    beginMutation,
  };
}
async function confirm(operation: (typeof operations)[number]) {
  fireEvent.click(screen.getByRole("button", { name: "Open history" }));
  if (operation === "delete" || operation === "restore") {
    fireEvent.click(
      await screen.findByTestId(
        operation === "delete"
          ? rowHistoryDeleteButtonTestId()
          : rowHistoryRestoreButtonTestId(),
      ),
    );
    fireEvent.click(
      await screen.findByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation }),
      ),
    );
  } else {
    fireEvent.click(
      await screen.findByTestId(
        rowHistoryActionTestId({
          action: operation,
          historyItemRef: item.history_item_ref,
        }),
      ),
    );
    fireEvent.click(
      await screen.findByTestId(
        rowHistoryRollbackConfirmButtonTestId({
          action: operation,
          historyItemRef: item.history_item_ref,
        }),
      ),
    );
  }
}

describe("History recovery surfaces", () => {
  it.each([
    ["Timeline", "delete"],
    ["Timeline", "restore"],
    ["Timeline", "history_entry"],
    ["Timeline", "change_set"],
    ["Timeline", "row_restore"],
    ["Generic", "delete"],
    ["Generic", "restore"],
    ["Generic", "history_entry"],
    ["Generic", "change_set"],
    ["Generic", "row_restore"],
    ["Entity", "delete"],
    ["Entity", "restore"],
    ["Entity", "history_entry"],
    ["Entity", "change_set"],
    ["Entity", "row_restore"],
    ["Assessment", "delete"],
    ["Assessment", "restore"],
    ["Assessment", "history_entry"],
    ["Assessment", "change_set"],
    ["Assessment", "row_restore"],
  ] as const)("retains exact %s %s recovery after inspector closure", async (surface, operation) => {
    const t = setup(surface, operation);
    const { rerender } = render(t.view(true));
    await confirm(operation);
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.phase).toBe("uncertain"),
    );
    expect(t.runtime.getSnapshot().primaryLabel).toBe("Syncing");
    expect(t.runtime.pendingQueue().model.snapshot().units).toEqual([]);
    expect(t.beginMutation).not.toHaveBeenCalled();
    const newer = screen.getByRole("textbox", { name: "Newer interaction" });
    newer.focus();
    rerender(t.view(false));
    expect(document.activeElement).toBe(newer);
    openHistoryRecovery();
    expect(document.activeElement).toBe(
      screen.getByRole("heading", {
        level: 2,
        name: /Soft-delete row|Restore deleted row|Restore row fields|Reverse history entry|Reverse change set/,
      }),
    );
    if (surface === "Timeline" && operation === "history_entry") {
      const read = vi.spyOn(t.port, "load");
      for (let page = 1; page <= 3; page++)
        read.mockImplementationOnce(async () => ({
          kind: "accepted",
          value: {
            record_id: "record",
            incident_id: "incident",
            row_version: 5,
            deleted: false,
            items: [],
            paging: {
              limit: 100,
              has_more: true,
              next_cursor: `review-${page}`,
            },
          },
        }));
      fireEvent.click(
        screen.getByRole("button", { name: "Review current history" }),
      );
      await screen.findByRole("button", { name: "Continue checking" });
      expect(
        screen.getByRole("region", { name: "History action recovery" })
          .textContent,
      ).toContain("Outcome unknown");
      expect(screen.queryByText(/has not been sent/)).toBeNull();
      expect(t.owner.getSnapshot()[0]?.phase).toBe("uncertain");
      expect(t.send).toHaveBeenCalledOnce();
      fireEvent.click(screen.getByRole("button", { name: "Cancel checking" }));
      read.mockRestore();
    }
    fireEvent.click(
      screen.getByRole("button", { name: "Replay exact action" }),
    );
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("complete"),
    );
    expect(t.send).toHaveBeenCalledTimes(2);
    expect(t.send.mock.calls[1]?.[0]).toBe(t.send.mock.calls[0]?.[0]);
    expect(t.counts).toEqual({ mutations: 1, changeSets: 1, revisions: 1 });
    expect(t.runtime.getSnapshot().primaryLabel).toBe("Saved");
    expect(t.effects.deleteAccepted).not.toHaveBeenCalled();
    expect(t.effects.restoreAccepted).not.toHaveBeenCalled();
    expect(t.effects.rollbackAccepted).not.toHaveBeenCalled();
    expect(
      screen.getByRole("region", { name: "History action recovery" })
        .textContent,
    ).not.toContain(t.send.mock.calls[0]?.[0].id);
  });

  it("keeps acknowledgement through refresh failure and retries reads only", async () => {
    const t = setup("Generic");
    t.send.mockImplementation(async (attempt) => ({
      kind: "acknowledged",
      receipt: t.commit(attempt),
    }));
    t.refresh.mockRejectedValueOnce(new Error("projection unavailable"));
    render(t.view(true));
    await confirm("delete");
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("required"),
    );
    expect(t.owner.getSnapshot()[0]?.receipt?.changeSetId).toBe(
      "accepted-change",
    );
    expect(t.runtime.getSnapshot().primaryLabel).toBe("Saved");
    openHistoryRecovery();
    fireEvent.click(
      screen.getByRole("button", { name: "Refresh completed action" }),
    );
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.reconciliation).toBe("complete"),
    );
    expect(t.send).toHaveBeenCalledTimes(1);
    expect(t.effects.deleteAccepted).toHaveBeenCalledTimes(1);
    expect(t.refresh).toHaveBeenCalledTimes(2);
  });

  it("hides retained content during session suspension and reauthorizes replay after closure", async () => {
    const t = setup("Assessment");
    const { rerender } = render(t.view(true));
    await confirm("delete");
    await waitFor(() =>
      expect(t.owner.getSnapshot()[0]?.phase).toBe("uncertain"),
    );
    act(() => t.runtime.invalidate({ kind: "session_unavailable" }));
    expect(screen.queryByRole("button", { name: "Recovery (1)" })).toBeNull();
    expect(screen.queryByText("History access is unavailable")).toBeNull();
    expect(screen.queryByTestId(rowHistoryPanelTestId())).toBeNull();
    expect(screen.queryByText("Changed fields")).toBeNull();
    expect(screen.queryByText("Current row version")).toBeNull();
    expect(screen.queryByText("Record ID")).toBeNull();
    rerender(t.view(false));
    act(() =>
      t.owner.setAuthority({
        actorId: "reviewer",
        incidentId: "incident",
        role: "reviewer",
        closed: true,
      }),
    );
    openHistoryRecovery();
    expect(
      (
        screen.getByRole("button", {
          name: "Replay exact action",
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    act(() =>
      t.owner.setAuthority({
        actorId: "reviewer",
        incidentId: "incident",
        role: "viewer",
        closed: false,
      }),
    );
    expect(
      (
        screen.getByRole("button", {
          name: "Replay exact action",
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(t.send).toHaveBeenCalledTimes(1);
    act(() =>
      t.owner.setAuthority({
        actorId: "other-actor",
        incidentId: "incident",
        role: "reviewer",
        closed: false,
      }),
    );
    expect(screen.queryByText("Reviewed row")).toBeNull();
    expect(t.owner.getSnapshot()).toEqual([]);
  });
});

function openHistoryRecovery() {
  fireEvent.click(screen.getByRole("button", { name: /^Recovery \(\d+\)$/ }));
  fireEvent.click(
    screen.getByRole("button", {
      name: /^(Soft-delete row|Restore deleted row|Restore row fields|Reverse history entry|Reverse change set).* ·/,
    }),
  );
}
