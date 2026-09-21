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
import { historyDiffFixture } from "../../testing/workbookHistoryTestSupport";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import type { RecordHistoryItem } from "../adapters/workbookHistoryResponse";
import { WorkbookBatchRecovery } from "../components/WorkbookBatchRecovery";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { WorkbookBatchTransport } from "../runtime/workbookBatchOperation";
import { WorkbookHistoryContext } from "./WorkbookHistoryContext";
import type { WorkbookRecordHistoryPort } from "./workbookHistoryOperation";
import type { HistoryPage } from "./workbookHistoryPage";

const view = "cartulary.view.timeline.v2";
const authority = {
  actorId: "analyst",
  incidentId: "incident",
  sessionIdentity: "session",
  role: "reviewer" as const,
  closed: false,
};
const event = (ref: string, change = "batch-change"): RecordHistoryItem => ({
  actor_user_id: "analyst",
  committed_at: "2026-09-18T10:00:00Z",
  history_item_ref: ref,
  history_entry_ref: `opaque-${ref}`,
  operation: "patch",
  diff_summary: historyDiffFixture(`Change ${ref}`),
  change_set_id: change,
  reversible: true,
  available_rollback_actions: ["history_entry", "change_set"],
});
const page = (
  items: readonly RecordHistoryItem[],
  next: string | null = null,
  version = 3,
): HistoryPage => ({
  incident_id: "incident",
  record_id: "record",
  row_version: version,
  deleted: false,
  representation_generation: "cartulary.history.1",
  items,
  paging:
    next === null
      ? { limit: 100, has_more: false, next_cursor: null }
      : { limit: 100, has_more: true, next_cursor: next },
});
afterEach(cleanup);
function fixture(
  role: "editor" | "reviewer" = "reviewer",
  withConflict = false,
) {
  let next = 0;
  const runtime = new WorkbookMutationRuntime(
    { incidentId: "incident", clientInstanceId: "client" },
    { create: () => `attempt-${++next}` },
    createWorkbookPendingMutationAdapter({
      apiBase: undefined,
      incidentId: "incident",
    }),
  );
  runtime.batches.setAuthority({ ...authority, role });
  runtime.history.setAuthority({ ...authority, role });
  const refresh = vi.fn(async () => {});
  runtime.registerSurface(view, refresh);
  runtime.history.registerSurface(view, refresh);
  const batchSend = vi.fn<WorkbookBatchTransport["send"]>(async (attempt) => ({
    kind: "acknowledged",
    receipt: {
      viewSchemaId: view,
      changeSetId:
        attempt.plan.recordIds[0] === "record"
          ? "batch-change"
          : "newer-change",
      conflicts: withConflict
        ? [
            {
              record_id: "conflicted",
              field_key: "timeline.activity_synopsis_text",
              base_row_version: 1,
              current_row_version: 2,
              conflict_token: "retained-conflict",
              conflict_resolution_class: "text_compare_merge",
              base_value: "before",
              client_value: null,
              server_value: "peer",
            },
          ]
        : [],
      rows: [
        {
          record_id: attempt.plan.recordIds[0] ?? "record",
          row_version: 2,
          cells: {
            "timeline.activity_synopsis_text": {
              value: `Historical ${attempt.plan.recordIds[0]}`,
              display: "",
            },
          },
        },
      ],
    },
  }));
  runtime.batches.configure({
    capture: (plan, authority, id) => ({
      id,
      authority,
      plan,
      apiBase: undefined,
      path: "/captured",
      body: JSON.stringify(plan.request),
    }),
    send: batchSend,
  });
  const load = vi.fn<WorkbookRecordHistoryPort["load"]>(async () => ({
    kind: "accepted",
    value: page([event("original")]),
  }));
  const send = vi.fn<WorkbookRecordHistoryPort["send"]>(async (attempt) => ({
    kind: "acknowledged",
    receipt: {
      kind: "rollback",
      recordId: "record",
      incidentId: "incident",
      rowVersion: 4,
      changeSetId: "reversal",
      affectedRecordIds: ["other", "record"],
      target:
        attempt.pending.kind === "rollback"
          ? attempt.pending.target
          : { kind: "change_set", change_set_id: "batch-change" },
    },
  }));
  runtime.history.configure({ load, send });
  const admit = async (recordId = "record") => {
    let id: string | null = null;
    await act(async () => {
      id = runtime.batches.admit(
        {
          operation: "applyWorkbookBulkMutation",
          recordIds: withConflict ? [recordId, "conflicted"] : [recordId],
          request: {
            kind: "clear_cells_v1",
            view_schema_id: view,
            field_keys: ["timeline.activity_synopsis_text"],
            targets: [
              { record_id: recordId, base_row_version: 1 },
              ...(withConflict
                ? [{ record_id: "conflicted", base_row_version: 1 }]
                : []),
            ],
          },
        },
        { delivery: {} },
      );
      for (let i = 0; i < 30; i++) await Promise.resolve();
    });
    await waitFor(() =>
      expect(
        runtime.batches.getSnapshot().entries.find((entry) => entry.id === id)
          ?.reconciliation,
      ).toBe("complete"),
    );
    return id;
  };
  render(
    <WorkbookHistoryContext.Provider value={runtime}>
      <WorkbookRecoveryFixture>
        <WorkbookBatchRecovery runtime={runtime} />
        <button type="button">Newer editing</button>
      </WorkbookRecoveryFixture>
    </WorkbookHistoryContext.Provider>,
  );
  const open = async () => {
    fireEvent.click(screen.getByRole("button", { name: /^Recovery \(\d+\)$/ }));
    const completed = screen.queryByText("Completed", { exact: true });
    if (completed) fireEvent.click(completed);
    fireEvent.click(
      screen.getByRole("button", { name: "Clear contents · Timeline" }),
    );
    expect(load).not.toHaveBeenCalled();
    const button = screen.getByRole("button", {
      name: "Review this change: Historical record",
    });
    button.focus();
    fireEvent.click(button);
    await waitFor(() => expect(load).toHaveBeenCalled());
  };
  return { runtime, load, send, batchSend, admit, open, refresh };
}

describe("Batch History review", () => {
  it("reads only the explicitly chosen off-window record and marks multiple later-page entries without writing", async () => {
    const t = fixture("editor");
    t.load.mockImplementation(async (_record, _signal, request) => ({
      kind: "accepted",
      value:
        request?.cursorToken === "older"
          ? page([event("part-a"), event("part-b")], "last")
          : request?.cursorToken === "last"
            ? page([event("part-c")])
            : page([event("newer", "another-change")], "older"),
    }));
    await t.admit();
    await t.open();
    await waitFor(() =>
      expect(screen.getAllByText("Requested change")).toHaveLength(2),
    );
    expect(t.load).toHaveBeenCalledTimes(2);
    expect(t.load.mock.calls.map(([record]) => record)).toEqual([
      "record",
      "record",
    ]);
    expect(
      (
        screen.getAllByRole("button", {
          name: "Reverse change set",
        })[0] as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: "Load older entries" }));
    await waitFor(() =>
      expect(screen.getAllByText("Requested change")).toHaveLength(3),
    );
    expect(t.send).not.toHaveBeenCalled();
    expect(t.batchSend).toHaveBeenCalledTimes(1);
  });
  it("pauses bounded lookup and retries expired and failed pages without claiming absence", async () => {
    const t = fixture();
    let calls = 0;
    t.load.mockImplementation(async () => {
      calls++;
      if (calls === 4)
        return {
          kind: "rejected",
          failure: { kind: "retryable", message: "Page failed" },
        };
      if (calls === 5)
        return {
          kind: "rejected",
          failure: {
            kind: "terminal",
            publicCode: "invalid_pagination_request",
            message: "Cursor expired",
          },
        };
      return {
        kind: "accepted",
        value:
          calls < 6
            ? page([event(`other-${calls}`, "other")], `cursor-${calls}`)
            : page([event("found")], null, 4),
      };
    });
    await t.admit();
    await t.open();
    await screen.findByText("More history remains to be checked.");
    expect(t.load).toHaveBeenCalledTimes(3);
    fireEvent.click(screen.getByRole("button", { name: "Continue checking" }));
    await screen.findByText("Page failed");
    fireEvent.click(screen.getByRole("button", { name: "Retry checking" }));
    await screen.findByText("Cursor expired");
    expect(t.load.mock.calls[3]?.[2]).toEqual(t.load.mock.calls[4]?.[2]);
    fireEvent.click(
      screen.getByRole("button", { name: "Start checking again" }),
    );
    await screen.findByText("Requested change");
    expect(t.load.mock.calls[5]?.[2]).toEqual({});
    expect(t.send).not.toHaveBeenCalled();
  });
  it("cancels delayed reads on close and never reclaims newer focus", async () => {
    const t = fixture();
    let finish!: (value: { kind: "accepted"; value: HistoryPage }) => void;
    t.load.mockImplementation(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    await t.admit();
    await t.open();
    fireEvent.click(
      screen.getByRole("button", { name: "Close change review" }),
    );
    screen.getByRole("button", { name: "Newer editing" }).focus();
    await act(async () =>
      finish({ kind: "accepted", value: page([event("late")]) }),
    );
    expect(screen.queryByText("Requested change")).toBeNull();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Newer editing" }),
    );
    expect(t.load).toHaveBeenCalledTimes(1);
  });
  it("invalidates confirmation on demotion or closure while keeping current authorized reading", async () => {
    for (const next of [
      { ...authority, role: "editor" as const },
      { ...authority, closed: true },
    ]) {
      const t = fixture();
      await t.admit();
      await t.open();
      await screen.findByText("Requested change");
      fireEvent.click(
        screen.getByRole("button", { name: "Reverse change set" }),
      );
      await screen.findByRole("button", { name: "Confirm rollback" });
      await act(async () => {
        t.runtime.history.setAuthority(next);
        t.runtime.batches.setAuthority(next);
      });
      await screen.findByText("Requested change");
      expect(
        screen.queryByRole("button", { name: "Confirm rollback" }),
      ).toBeNull();
      expect(
        (
          screen.getByRole("button", {
            name: "Reverse change set",
          }) as HTMLButtonElement
        ).disabled,
      ).toBe(true);
      expect(t.send).not.toHaveBeenCalled();
      cleanup();
    }
  });
  it("clears protected review on History suspension scoped revocation and account replacement", async () => {
    for (const change of ["suspend", "revoke", "replace"] as const) {
      const t = fixture();
      await t.admit();
      await t.open();
      await screen.findByText("Requested change");
      await act(async () => {
        if (change === "suspend") t.runtime.history.suspend();
        else
          t.runtime.history.setAuthority(
            change === "revoke"
              ? { ...authority, role: "" }
              : {
                  ...authority,
                  actorId: "replacement",
                  sessionIdentity: "new",
                },
          );
      });
      expect(
        screen.queryByRole("region", { name: "Review this change" }),
      ).toBeNull();
      expect(screen.queryByText("Requested change")).toBeNull();
      expect(t.send).not.toHaveBeenCalled();
      cleanup();
    }
  });
  it("uses current deletion metadata and keeps record absence separate from incident authority", async () => {
    const t = fixture();
    t.load.mockResolvedValue({
      kind: "accepted",
      value: { ...page([event("deleted")], null, 7), deleted: true },
    });
    await t.admit();
    await t.open();
    await screen.findByText("Requested change");
    expect(
      screen.getByRole("button", { name: "Restore deleted row" }),
    ).not.toBeNull();
    t.load.mockResolvedValue({
      kind: "rejected",
      failure: {
        kind: "terminal",
        publicCode: "record_not_found",
        message: "Record unavailable",
      },
    });
    fireEvent.click(screen.getByRole("button", { name: "Refresh history" }));
    await screen.findAllByText("Record unavailable");
    expect(screen.queryByText("Requested change")).toBeNull();
    expect(t.runtime.history.readable).toBe(true);
    expect(
      t.runtime.batches.getSnapshot().entries[0]?.receipt?.changeSetId,
    ).toBe("batch-change");
  });
  it("keeps the active locator after receipt pruning and releases it on explicit close", async () => {
    const t = fixture();
    const id = await t.admit();
    await t.open();
    await screen.findByText("Requested change");
    screen.getByRole("button", { name: "Newer editing" }).focus();
    await t.admit("other-record");
    expect(
      t.runtime.batches.getSnapshot().entries.some((entry) => entry.id === id),
    ).toBe(false);
    expect(
      screen.getByText("Record at completion: Historical record"),
    ).not.toBeNull();
    expect(
      screen.queryByRole("button", { name: /Review this change:/ }),
    ).toBeNull();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Newer editing" }),
    );
    expect(t.load).toHaveBeenCalledTimes(1);
    fireEvent.click(
      screen.getByRole("button", { name: "Close change review" }),
    );
    expect(
      screen.queryByRole("region", { name: "Review this change" }),
    ).toBeNull();
    expect(t.send).not.toHaveBeenCalled();
  });
  it("reviews partial success while preserving the existing unresolved-conflict action gate", async () => {
    const t = fixture("reviewer", true);
    await t.admit();
    await t.open();
    await screen.findByText("Requested change");
    const receipt = t.runtime.batches.getSnapshot().entries[0]?.receipt;
    const conflicts = t.runtime.getSnapshot().conflicts;
    expect(receipt?.conflicts).toHaveLength(1);
    expect(conflicts).toHaveLength(1);
    fireEvent.click(screen.getByRole("button", { name: "Reverse change set" }));
    await screen.findByText(
      "Earlier row changes must settle before reviewing this action.",
    );
    expect(
      screen.queryByRole("button", { name: "Confirm rollback" }),
    ).toBeNull();
    expect(t.send).not.toHaveBeenCalled();
    expect(t.runtime.getSnapshot().conflicts).toEqual(conflicts);
    expect(t.runtime.batches.getSnapshot().entries[0]?.receipt).toBe(receipt);
    expect(
      screen.getByRole("button", { name: "Review conflicts" }),
    ).not.toBeNull();
  });
  it("delegates confirmed whole-change reversal to the existing owner without changing the batch receipt", async () => {
    const t = fixture();
    t.send.mockImplementationOnce(async (attempt) => {
      t.load.mockResolvedValue({
        kind: "accepted",
        value: page(
          [
            event("reversal", "reversal"),
            {
              ...event("original"),
              reversible: false,
              available_rollback_actions: [],
            },
          ],
          null,
          4,
        ),
      });
      return {
        kind: "acknowledged",
        receipt: {
          kind: "rollback",
          recordId: "record",
          incidentId: "incident",
          rowVersion: 4,
          changeSetId: "reversal",
          target:
            attempt.pending.kind === "rollback"
              ? attempt.pending.target
              : { kind: "change_set", change_set_id: "batch-change" },
          affectedRecordIds: ["other", "record"],
        },
      };
    });
    await t.admit();
    await t.open();
    await screen.findByText("Requested change");
    const receipt = t.runtime.batches.getSnapshot().entries[0]?.receipt;
    fireEvent.click(screen.getByRole("button", { name: "Reverse change set" }));
    await screen.findByRole("button", { name: "Confirm rollback" });
    expect(
      screen.getByText(/all reversible changes in this change set/),
    ).not.toBeNull();
    expect(t.send).not.toHaveBeenCalled();
    const confirm = screen.getByRole("button", { name: "Confirm rollback" });
    fireEvent.click(confirm);
    fireEvent.click(confirm);
    await waitFor(() => expect(t.send).toHaveBeenCalledTimes(1));
    const attempt = t.send.mock.calls[0]?.[0];
    expect(attempt && JSON.parse(attempt.body)).toMatchObject({
      base_row_version: 3,
      target: { kind: "change_set", change_set_id: "batch-change" },
    });
    await waitFor(() =>
      expect(t.runtime.history.getSnapshot()[0]?.receipt?.changeSetId).toBe(
        "reversal",
      ),
    );
    expect(t.runtime.batches.getSnapshot().entries[0]?.receipt).toBe(receipt);
    expect(t.runtime.history.getSnapshot()[0]?.receipt).toMatchObject({
      affectedRecordIds: ["other", "record"],
    });
  });
});
