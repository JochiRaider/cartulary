import { act, renderHook } from "@testing-library/react";
import { expect, it } from "vitest";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { useTimelineRows } from "./hooks/useTimelineRows";
import { timelineMutationOwnerFor } from "./mutations/WorkbookTimelineMutationOwner";

it("useTimelineRows owns the initial draft row ref and monotonic draft index", () => {
  const runtime = new WorkbookMutationRuntime(
    { incidentId: "incident", clientInstanceId: "client" },
    { create: () => "test-id" },
    {
      execute: async () => {
        throw new Error("Unexpected dispatch");
      },
    },
  );
  const owner = timelineMutationOwnerFor(runtime);
  const { result } = renderHook(() => useTimelineRows(owner));

  expect(result.current.rows).toHaveLength(1);
  expect(result.current.rows[0]?.recordId).toBeNull();
  expect(result.current.rowsRef.current).toBe(result.current.rows);
  expect(result.current.nextDraftIndex()).toBe(2);
  expect(result.current.nextDraftIndex()).toBe(3);

  act(() => {
    result.current.commands.replaceRows([]);
  });
  expect(result.current.rows).toEqual([]);
  expect(result.current.rowsRef.current).toBe(result.current.rows);
  expect(result.current.nextDraftIndex()).toBe(4);
});
