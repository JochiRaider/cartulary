import { act, renderHook } from "@testing-library/react";
import { expect, it } from "vitest";
import { useTimelineRows } from "./hooks/useTimelineRows";

it("useTimelineRows owns the initial draft row ref and monotonic draft index", () => {
  const { result } = renderHook(() => useTimelineRows());

  expect(result.current.rows).toHaveLength(1);
  expect(result.current.rows[0]?.recordId).toBeNull();
  expect(result.current.rowsRef.current).toBe(result.current.rows);
  expect(result.current.nextDraftIndex()).toBe(2);
  expect(result.current.nextDraftIndex()).toBe(3);

  act(() => {
    result.current.setRows([]);
  });
  expect(result.current.rows).toEqual([]);
  expect(result.current.rowsRef.current).toBe(result.current.rows);
  expect(result.current.nextDraftIndex()).toBe(4);
});
