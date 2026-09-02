import { act, renderHook } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { initialWorkbookRecordHistoryState } from "../inspector/workbookRecordHistoryModel";
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
      rowHistory: initialWorkbookRecordHistoryState(),
      rows: [],
      selectedMentionRef: null,
      selectedRowId: null,
      setInspectorMessage: vi.fn(),
      setIsInspectorOpen,
      dispatchRowHistory: vi.fn(),
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
