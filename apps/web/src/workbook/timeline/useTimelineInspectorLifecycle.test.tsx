import { act, renderHook } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { useTimelineInspectorLifecycle } from "./hooks/useTimelineInspectorSelection";

it("useTimelineInspectorLifecycle shares one close command across explicit and layout requests", () => {
  const clearRowHistory = vi.fn();
  const setIsInspectorOpen = vi.fn();
  const { result } = renderHook(() =>
    useTimelineInspectorLifecycle({
      cancelRowHistoryRequests: vi.fn(),
      clearRowHistory,
      gridShellRef: { current: null },
      inspectorInvalidationCause: null,
      inspectorInvalidationGeneration: 0,
      inspectorMentions: [],
      restoreTimelineFocusAnchor: () => false,
      rowHistory: {
        data: null,
        error: null,
        recordId: null,
        status: "idle",
      },
      rows: [],
      selectedMentionRef: null,
      selectedRowId: null,
      setInspectorMessage: vi.fn(),
      setIsInspectorOpen,
      setRowHistory: vi.fn(),
      setRowHistoryPendingAction: vi.fn(),
      setSelectedMentionRef: vi.fn(),
      setSelectedResolveTargetId: vi.fn(),
      setSelectedRowId: vi.fn(),
      workbookFocusAnchorRef: { current: null },
    }),
  );

  act(() => {
    result.current.closeInspector();
    result.current.closeInspector();
  });

  expect(setIsInspectorOpen).toHaveBeenNthCalledWith(1, false);
  expect(setIsInspectorOpen).toHaveBeenNthCalledWith(2, false);
  expect(clearRowHistory).not.toHaveBeenCalled();
});
