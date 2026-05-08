import { describe, expect, it } from "vitest";

import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  gridDraftRowSelector,
  gridSavedRowsSelector,
  gridScrollportClassName,
  gridScrollportSelector,
  gridShellTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  rowInspectorFieldTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
} from "./index";

describe("@cartulary/ui-contracts workbook row selectors", () => {
  it("derives inspector field ids from the stable row cell id", () => {
    expect(rowInspectorFieldTestId("record-1", "details")).toBe(
      "row-record-1-details-inspector",
    );
  });

  it("targets saved and draft workbook rows when scoped through the grid shell", () => {
    expect(gridShellTestId("timeline")).toBe("timeline-grid-shell");
    expect(gridScrollportClassName()).toBe("cartulary-grid-scrollport");
    expect(gridScrollportSelector()).toBe(".cartulary-grid-scrollport");
    expect(gridSavedRowsSelector()).toBe(
      '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])',
    );
    expect(gridDraftRowSelector()).toBe('[role="row"][data-grid-record-id=""]');
  });

  it("derives stable Phase 6 collaboration and status selectors", () => {
    expect(conflictMarkerTestId("record-1", "timeline.summary")).toBe(
      "conflict-marker-record-1-timeline-summary",
    );
    expect(rowPresenceMarkerTestId("record-1")).toBe("presence-row-record-1");
    expect(cellPresenceMarkerTestId("record-1", "timeline.summary")).toBe(
      "presence-cell-record-1-timeline-summary",
    );
    expect(saveStateTestId()).toBe("save-state");
    expect(pendingQueueNoticeTestId()).toBe("pending-queue-notice");
    expect(pendingQueueCountTestId()).toBe("pending-queue-count");
  });
});
