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
    });
  });
});
