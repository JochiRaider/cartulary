import { act, renderHook } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import type { CoordinationMutationCommandPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { useCoordinationWorkflowController } from "./useCoordinationWorkflowController";

const taskRow: WorkbookQueryRow = {
  cells: {},
  record_id: "task-1",
  row_version: 7,
};
const decisionRows: readonly WorkbookQueryRow[] = [
  { cells: {}, record_id: "decision-1", row_version: 4 },
  { cells: {}, record_id: "decision-2", row_version: 6 },
];

function mutationPorts() {
  return {
    beginMutation: vi.fn(),
    completeGenericMutation: vi.fn(async () => undefined),
    rejectMutationFailure: vi.fn(),
    setValidationError: vi.fn(),
  };
}

it("owns task lifecycle and decision supersede coordination with distinct semantic outcomes", async () => {
  const updateTaskLifecycle = vi.fn(async () => ({
    kind: "accepted" as const,
    value: {
      changeSetId: "change-task",
      row: taskRow,
      status: "done" as const,
      viewSchemaId: "cartulary.view.task_requests.v1",
    },
  }));
  const supersedeDecision = vi.fn(async () => ({
    kind: "accepted" as const,
    value: {
      changeSetId: "change-decision",
      replacementRecordId: "decision-2",
      replacementRowVersion: 7,
      targetRecordId: "decision-1",
      targetRowVersion: 5,
      targetStatus: "superseded",
      viewSchemaId: "cartulary.view.decisions.v1",
    },
  }));
  const commands: CoordinationMutationCommandPort = {
    updateTaskLifecycle,
    supersedeDecision,
  };
  const mutation = mutationPorts();
  const { result, rerender } = renderHook(
    ({ resetKey, rows }) =>
      useCoordinationWorkflowController({
        mutation,
        mutationCommands: commands,
        resetKey,
        rows,
      }),
    {
      initialProps: {
        resetKey: "tasks",
        rows: [taskRow] as readonly WorkbookQueryRow[],
      },
    },
  );

  act(() => {
    result.current.lifecycle.setRecordId("task-1");
    result.current.lifecycle.setStatus("done");
    result.current.lifecycle.setBlockedReason("no longer relevant");
  });
  await act(async () => result.current.lifecycle.submit());

  expect(updateTaskLifecycle).toHaveBeenCalledWith({
    baseRowVersion: 7,
    blockedReason: undefined,
    recordId: "task-1",
    status: "done",
  });
  expect(result.current.lifecycle.blockedReason).toBe("");

  rerender({ resetKey: "decisions", rows: decisionRows });
  act(() => {
    result.current.supersede.setTargetId("decision-1");
    result.current.supersede.setReplacementId("decision-2");
    result.current.supersede.setReason("  replaced by later review  ");
  });
  await act(async () => result.current.supersede.submit());

  expect(supersedeDecision).toHaveBeenCalledWith({
    baseRowVersion: 4,
    reason: "replaced by later review",
    replacementRecordId: "decision-2",
    targetRecordId: "decision-1",
  });
  expect(result.current.supersede.reason).toBe("");
  expect(mutation.completeGenericMutation).toHaveBeenCalledTimes(2);
  expect(mutation.rejectMutationFailure).not.toHaveBeenCalled();
});

it("makes late coordination outcomes inert after lifecycle invalidation", async () => {
  let resolveTask:
    | ((
        value: Awaited<
          ReturnType<CoordinationMutationCommandPort["updateTaskLifecycle"]>
        >,
      ) => void)
    | undefined;
  const pendingTask = new Promise<
    Awaited<ReturnType<CoordinationMutationCommandPort["updateTaskLifecycle"]>>
  >((resolve) => {
    resolveTask = resolve;
  });
  const commands: CoordinationMutationCommandPort = {
    updateTaskLifecycle: vi.fn(() => pendingTask),
    supersedeDecision: vi.fn(),
  };
  const mutation = mutationPorts();
  const { result, rerender } = renderHook(
    ({ resetKey }) =>
      useCoordinationWorkflowController({
        mutation,
        mutationCommands: commands,
        resetKey,
        rows: [taskRow],
      }),
    { initialProps: { resetKey: "active" } },
  );

  act(() => {
    result.current.lifecycle.setRecordId("task-1");
    result.current.lifecycle.setStatus("done");
  });
  act(() => {
    void result.current.lifecycle.submit();
  });
  rerender({ resetKey: "authorization-lost" });
  await act(async () => {
    resolveTask?.({
      kind: "accepted",
      value: {
        changeSetId: "late-change",
        row: taskRow,
        status: "done",
        viewSchemaId: "cartulary.view.task_requests.v1",
      },
    });
    await pendingTask;
  });

  expect(mutation.beginMutation).toHaveBeenCalledTimes(1);
  expect(mutation.completeGenericMutation).not.toHaveBeenCalled();
  expect(mutation.rejectMutationFailure).not.toHaveBeenCalled();
  expect(result.current.lifecycle.recordId).toBe("");
});
