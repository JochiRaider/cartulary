import { describe, expect, it } from "vitest";
import {
  initialTimelineWorkbookLifecycleState,
  reduceTimelineWorkbookLifecycle,
  type TimelineWorkbookLifecycleAction,
} from "../hooks/useTimelineWorkbookRuntime";

describe("Timeline workbook runtime", () => {
  it("reduces load, refresh, save, conflict, and recovery transitions deterministically", () => {
    const actions: readonly TimelineWorkbookLifecycleAction[] = [
      { type: "load_error", value: "load failed" },
      { type: "initial_loading", value: false },
      { type: "refreshing", value: true },
      { type: "refresh_error", value: "refresh failed" },
      { type: "save_state", value: "Syncing" },
      {
        type: "save_state_secondary_message",
        value: "Pending replay",
      },
      { type: "save_state", value: "Conflict" },
      { type: "load_error", value: null },
      { type: "refresh_error", value: null },
      { type: "refreshing", value: false },
      { type: "save_state", value: "Saved" },
      { type: "save_state_secondary_message", value: null },
    ];

    const finalState = actions.reduce(
      reduceTimelineWorkbookLifecycle,
      initialTimelineWorkbookLifecycleState,
    );

    expect(finalState).toEqual({
      isInitialLoading: false,
      isRefreshing: false,
      loadError: null,
      refreshError: null,
      saveState: "Saved",
      saveStateSecondaryMessage: null,
    });
    expect(
      actions
        .slice(0, 7)
        .reduce(
          reduceTimelineWorkbookLifecycle,
          initialTimelineWorkbookLifecycleState,
        ),
    ).toEqual({
      isInitialLoading: false,
      isRefreshing: true,
      loadError: "load failed",
      refreshError: "refresh failed",
      saveState: "Conflict",
      saveStateSecondaryMessage: "Pending replay",
    });
  });
});
