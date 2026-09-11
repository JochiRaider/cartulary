import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  fireEvent,
  render,
  renderHook,
  screen,
} from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { CoordinationWorkflowBindings } from "./CoordinationWorkflowBindings";
import {
  TaskLifecycleDraftStore,
  taskLifecycleChanges,
  taskPatchErrors,
  taskStatuses,
  taskViewId,
} from "./taskLifecycleModel";
import { useCoordinationWorkflowController } from "./useCoordinationWorkflowController";

const taskRow: WorkbookQueryRow = {
  record_id: "task-1",
  row_version: 7,
  cells: {
    "task.status": { value: "open" },
    "task.owner_user_id": { value: "member-1" },
  },
};
function mutationPorts() {
  return {
    beginMutation: vi.fn(() => vi.fn()),
    completeGenericMutation: vi.fn(async () => {}),
    submitPatchMutation: vi.fn(async () => ({
      changeSetId: "receipt",
      row: taskRow,
      viewSchemaId: taskViewId,
    })),
  };
}
function setup(row = taskRow) {
  const mutation = mutationPorts(),
    drafts = new TaskLifecycleDraftStore();
  return {
    mutation,
    drafts,
    ...renderHook(
      (props) =>
        useCoordinationWorkflowController({ mutation, drafts, ...props }),
      { initialProps: { row, disabled: false } },
    ),
  };
}
it("initializes Task lifecycle from its saved state", () => {
  const { result } = setup();
  expect(result.current.value("task.status")).toBe("open");
  expect(result.current.value("task.owner_user_id")).toBe("member-1");
});
it("preserves Task lifecycle coordination independently of Decision supersession", async () => {
  const { result, mutation } = setup();
  act(() => {
    result.current.update("task.status", "done");
    result.current.update("task.blocked_reason", "obsolete");
  });
  await act(async () => result.current.submit());
  expect(mutation.submitPatchMutation).toHaveBeenCalledWith(
    expect.objectContaining({
      recordId: "task-1",
      baseRowVersion: 7,
      changes: [{ field_key: "task.status", value: "done" }],
    }),
  );
});
it("makes late coordination outcomes inert after lifecycle invalidation", async () => {
  const { result, rerender, drafts } = setup();
  act(() => result.current.update("task.status", "blocked"));
  const second = { ...taskRow, record_id: "task-2" };
  rerender({ row: second, disabled: false });
  expect(result.current.value("task.status")).toBe("open");
  expect(drafts.read(taskRow).values["task.status"]).toBe("blocked");
  rerender({ row: taskRow, disabled: true });
  await act(async () => result.current.submit());
  expect(result.current.value("task.status")).toBe("blocked");
});
it("validates all Task status pairs and resulting lifecycle guards", () => {
  for (const from of taskStatuses)
    for (const to of taskStatuses) {
      const row = {
        ...taskRow,
        cells: {
          ...taskRow.cells,
          "task.status": { value: from },
          "task.blocked_reason": {
            value: from === "blocked" ? "Waiting" : null,
          },
          "task.completed_at": {
            value: from === "done" ? "2026-09-11T12:00:00Z" : null,
          },
        },
      };
      const changes = taskLifecycleChanges(
        {
          baseline: row,
          values: {
            "task.status": to,
            ...(to === "blocked" ? { "task.blocked_reason": "Waiting" } : {}),
          },
        },
        row,
      );
      expect(taskPatchErrors(row, changes).length, `${from} to ${to}`).toBe(
        (from === "done" && to === "canceled") ||
          (from === "canceled" && to === "done")
          ? 1
          : 0,
      );
    }
  const row = { ...taskRow, cells: { "task.status": { value: "canceled" } } };
  expect(
    taskPatchErrors(row, [{ field_key: "task.status", value: "blocked" }]).map(
      (error) => error.field,
    ),
  ).toEqual(["task.blocked_reason", "task.owner_user_id"]);
  expect(
    taskPatchErrors(row, [
      { field_key: "task.status", value: "blocked" },
      { field_key: "task.blocked_reason", value: "Waiting" },
      { field_key: "task.owner_user_id", value: "member-1" },
    ]),
  ).toEqual([]);
  expect(
    taskPatchErrors(taskRow, [
      { field_key: "task.status", value: "done" },
      { field_key: "task.completed_at", value: "bad time" },
    ])[0]?.field,
  ).toBe("task.completed_at");
});
it("preserves drafts until changed guard siblings are explicitly reviewed", async () => {
  const { result, mutation, rerender } = setup();
  act(() => result.current.update("task.status", "blocked"));
  act(() => result.current.update("task.blocked_reason", "Waiting"));
  const newer = {
    ...taskRow,
    row_version: 8,
    cells: { ...taskRow.cells, "task.owner_user_id": { value: "member-2" } },
  };
  rerender({ row: newer, disabled: false });
  await act(async () => result.current.submit());
  expect(mutation.submitPatchMutation).not.toHaveBeenCalled();
  expect(result.current.staleFields).toEqual(["task.owner_user_id"]);
  act(() => result.current.review("task.owner_user_id", false));
  await act(async () => result.current.submit());
  expect(mutation.submitPatchMutation).toHaveBeenCalledWith(
    expect.objectContaining({
      baseRowVersion: 8,
      changes: [
        { field_key: "task.status", value: "blocked" },
        { field_key: "task.blocked_reason", value: "Waiting" },
      ],
    }),
  );
});
it("admits one selected Task editor with role gates and no confirmation", async () => {
  const mutation = mutationPorts();
  const props = {
    contract: requireViewContract(taskViewId),
    disabled: false,
    currentIncidentRole: "editor" as const,
    disabledTokens: new Set<never>(),
    mutation,
    drafts: new TaskLifecycleDraftStore(),
    row: taskRow,
    referenceOptions: emptyGenericReferenceOptions(),
  };
  const { rerender } = render(<CoordinationWorkflowBindings {...props} />);
  expect(
    screen.queryByRole("combobox", { name: "Task lifecycle row" }),
  ).toBeNull();
  expect(
    screen.getAllByRole("button", { name: "Apply task status" }),
  ).toHaveLength(1);
  fireEvent.change(screen.getByLabelText("Task lifecycle status"), {
    target: { value: "done" },
  });
  await act(async () =>
    fireEvent.click(screen.getByRole("button", { name: "Apply task status" })),
  );
  expect(mutation.submitPatchMutation).toHaveBeenCalledTimes(1);
  expect(screen.queryByRole("dialog")).toBeNull();
  rerender(
    <CoordinationWorkflowBindings {...props} currentIncidentRole="viewer" />,
  );
  expect(
    screen
      .getByRole("button", { name: "Apply task status" })
      .matches(":disabled"),
  ).toBe(true);
  rerender(
    <CoordinationWorkflowBindings
      {...props}
      disabledTokens={new Set(["incident_closed"])}
    />,
  );
  expect(
    screen
      .getByRole("button", { name: "Apply task status" })
      .matches(":disabled"),
  ).toBe(true);
});
