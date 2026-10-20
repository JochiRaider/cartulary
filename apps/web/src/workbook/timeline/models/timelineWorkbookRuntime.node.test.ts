import { describe, expect, it } from "vitest";
import {
  initialWorkbookLifecycleState,
  reduceWorkbookLifecycle,
  type WorkbookLifecycleAction,
} from "../../runtime/workbookLifecycleModel";

describe("Timeline workbook runtime", () => {
  it("reduces load and refresh transitions without owning shell save state", () => {
    const actions: readonly WorkbookLifecycleAction[] = [
      { type: "load_error", value: "load failed" },
      { type: "initial_loading", value: false },
      { type: "refreshing", value: true },
      { type: "refresh_error", value: "refresh failed" },
      { type: "load_error", value: null },
      { type: "refresh_error", value: null },
      { type: "refreshing", value: false },
    ];

    const finalState = actions.reduce(
      reduceWorkbookLifecycle,
      initialWorkbookLifecycleState,
    );

    expect(finalState).toEqual({
      isInitialLoading: false,
      isRefreshing: false,
      loadError: null,
      refreshError: null,
      operationError: null,
    });
    expect(
      actions
        .slice(0, 4)
        .reduce(reduceWorkbookLifecycle, initialWorkbookLifecycleState),
    ).toEqual({
      isInitialLoading: false,
      isRefreshing: true,
      loadError: "load failed",
      refreshError: "refresh failed",
      operationError: null,
    });
  });

  it("keeps identical operation and query messages independent", () => {
    const failedRead = reduceWorkbookLifecycle(initialWorkbookLifecycleState, {
      type: "refresh_error",
      value: "Request failed",
    });
    const failedPaste = reduceWorkbookLifecycle(failedRead, {
      type: "operation_error",
      value: { family: "paste", message: "Request failed" },
    });
    const recoveredRead = reduceWorkbookLifecycle(failedPaste, {
      type: "refresh_error",
      value: null,
    });
    expect(recoveredRead.operationError).toEqual({
      family: "paste",
      message: "Request failed",
    });
    const recoveredPaste = reduceWorkbookLifecycle(failedPaste, {
      type: "operation_error",
      value: null,
    });
    expect(recoveredPaste.refreshError).toBe("Request failed");
    expect(recoveredPaste.operationError).toBeNull();
  });
});
