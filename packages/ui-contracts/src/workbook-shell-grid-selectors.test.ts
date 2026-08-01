import { describe, expect, it } from "vitest";

import {
  cartularyDesignTokenVars,
  gridActionsHeaderTestId,
  gridDataCellsSelector,
  gridDataRowsSelector,
  gridDraftRowSelector,
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowsSelector,
  gridGroupRowTestId,
  gridRowGutterTestId,
  gridSavedRowsSelector,
  gridScrollportClassName,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  systemViewSwitcherGroupTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  workbookAddRowButtonTestId,
  workbookGridRowHeightPx,
  workbookInlineDraftRowTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorFeatureGroupTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
  workbookRowActionMenuButtonTestId,
  workbookRowContextMenuTestId,
  workbookShellSlotLabel,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "./index";

describe("@cartulary/ui-contracts workbook shell and grid selectors", () => {
  it("builds stable System views switcher selectors from closed groups and view_schema_id", () => {
    const originalSurface = {
      title: "Indicators",
      viewSchemaId: "cartulary.view.indicators.v1",
    };
    const renamedSurface = {
      title: "Observable Signals",
      viewSchemaId: "cartulary.view.indicators.v1",
    };

    expect(systemViewSwitcherTriggerTestId()).toBe("system-view-selector");
    expect(systemViewSwitcherMenuTestId()).toBe("system-view-switcher-menu");
    expect(systemViewSwitcherGroupTestId("scope-indicators")).toBe(
      "system-view-switcher-group-scope-indicators",
    );
    expect(
      systemViewSwitcherOptionTestId(
        "scope-indicators",
        originalSurface.viewSchemaId,
      ),
    ).toBe(
      "system-view-switcher-option-scope-indicators-cartulary.view.indicators.v1",
    );
    expect(
      systemViewSwitcherOptionTestId(
        "scope-indicators",
        renamedSurface.viewSchemaId,
      ),
    ).toBe(
      systemViewSwitcherOptionTestId(
        "scope-indicators",
        originalSurface.viewSchemaId,
      ),
    );
    expect(systemViewSwitcherGroupTestId("coordination")).toBe(
      "system-view-switcher-group-coordination",
    );
    expect(systemViewSwitcherGroupTestId("review-learning")).toBe(
      "system-view-switcher-group-review-learning",
    );
    expect(systemViewSwitcherGroupTestId("optional-artifact-surfaces")).toBe(
      "system-view-switcher-group-optional-artifact-surfaces",
    );
  });

  it("fails closed for System views switcher group and option selector inputs", () => {
    expect(() => systemViewSwitcherGroupTestId("future" as never)).toThrow(
      "Invalid system view switcher group token: future",
    );
    expect(() =>
      systemViewSwitcherOptionTestId("scope-indicators", "timeline"),
    ).toThrow("Invalid view_schema_id selector token: timeline");
    expect(() =>
      systemViewSwitcherOptionTestId(
        "scope-indicators",
        "cartulary.view.future.v1",
      ),
    ).toThrow(
      "Unknown view_schema_id selector token: cartulary.view.future.v1",
    );
  });

  it("keeps visible field labels and display names out of cell selectors", () => {
    const originalField = {
      fieldKey: "timeline.activity_synopsis_text",
      label: "Summary",
    };
    const renamedField = {
      fieldKey: "timeline.activity_synopsis_text",
      label: "Executive summary",
    };
    const originalRow = {
      displayName: "Alpha workstation",
      recordId: "record-alpha",
    };
    const renamedRow = {
      displayName: "Renamed workstation",
      recordId: "record-alpha",
    };

    expect(rowCellTestId(originalRow.recordId, originalField.fieldKey)).toBe(
      rowCellTestId(renamedRow.recordId, renamedField.fieldKey),
    );
    expect(gridFilterChipTestId("cartulary.view.timeline.v2", "status")).toBe(
      gridFilterChipTestId("cartulary.view.timeline.v2", "status"),
    );
    expect(gridFilterFieldTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-filter-field",
    );
    expect(gridGroupingSelectTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-group-by",
    );
    expect(
      gridGroupRowTestId(
        "cartulary.view.timeline.v2",
        "timeline.capture_state",
        "rough",
      ),
    ).toBe("cartulary.view.timeline.v2-group-timeline.capture_state-rough");
    expect(
      gridGroupRowsSelector(
        "cartulary.view.timeline.v2",
        "timeline.capture_state",
      ),
    ).toBe(
      '[data-testid^="cartulary.view.timeline.v2-group-timeline.capture_state-"]',
    );
  });

  it("derives row and row-action selectors from record_id", () => {
    expect(rowInspectButtonTestId("record-alpha")).toBe(
      "row-record-alpha-inspect",
    );
  });

  it("derives inspector field ids from the stable row cell id", () => {
    expect(
      rowInspectorFieldTestId("record-1", "timeline.raw_activity_text"),
    ).toBe("row-record-1-timeline.raw_activity_text-inspector");
  });

  it("targets saved and draft workbook rows when scoped through the grid shell", () => {
    expect(gridShellTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-grid-shell",
    );
    expect(gridScrollportClassName()).toBe("cartulary-grid-scrollport");
    expect(gridScrollportSelector()).toBe(".cartulary-grid-scrollport");
    expect(gridActionsHeaderTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-actions-header",
    );
    expect(gridRowGutterTestId("cartulary.view.timeline.v2", "record-1")).toBe(
      "cartulary.view.timeline.v2-row-gutter-record-1",
    );
    expect(gridSavedRowsSelector()).toBe(
      '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])',
    );
    expect(gridDataRowsSelector()).toBe(
      '[role="row"][data-cartulary-grid-row-kind="data"]',
    );
    expect(gridDataCellsSelector()).toBe(
      '[role="row"][data-cartulary-grid-row-kind="data"] [role="gridcell"]',
    );
    expect(gridDraftRowSelector()).toBe(
      '[role="row"][data-cartulary-grid-draft-row="true"]',
    );
  });

  it("derives sheet toolbar, inspector, draft row, and row menu selectors from stable workbook ids", () => {
    expect(workbookAddRowButtonTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-add-row",
    );
    expect(workbookInspectorToggleTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-inspector-toggle",
    );
    expect(
      workbookInspectorCloseButtonTestId("cartulary.view.timeline.v2"),
    ).toBe("cartulary.view.timeline.v2-inspector-close");
    expect(
      workbookInspectorPanelTestId("cartulary.view.timeline.v2", "workflow"),
    ).toBe("cartulary.view.timeline.v2-inspector-panel-workflow");
    expect(
      workbookInspectorFeatureGroupTestId(
        "cartulary.view.timeline.v2",
        "create_related.task_request",
      ),
    ).toBe(
      "cartulary.view.timeline.v2-inspector-feature-create_related.task_request",
    );
    expect(
      workbookInspectorFeatureActionTestId(
        "cartulary.view.timeline.v2",
        "history.rollback",
      ),
    ).toBe(
      "cartulary.view.timeline.v2-inspector-feature-action-history.rollback",
    );
    expect(workbookInlineDraftRowTestId("cartulary.view.timeline.v2")).toBe(
      "cartulary.view.timeline.v2-inline-draft-row",
    );
    expect(
      workbookRowActionMenuButtonTestId(
        "cartulary.view.timeline.v2",
        "record-1",
      ),
    ).toBe("cartulary.view.timeline.v2-row-action-menu-record-1");
    expect(
      workbookRowContextMenuTestId("cartulary.view.timeline.v2", "record-1"),
    ).toBe("cartulary.view.timeline.v2-row-context-menu-record-1");
    expect(() =>
      workbookInspectorPanelTestId(
        "cartulary.view.timeline.v2",
        "details-title" as never,
      ),
    ).toThrow("Invalid workbook inspector panel token: details-title");
    expect(() =>
      workbookInspectorFeatureGroupTestId(
        "cartulary.view.timeline.v2",
        "Create task",
      ),
    ).toThrow("Invalid feature_group_key selector token: Create task");
  });

  it("preserves closed shell and system-view order with exact accessibility labels", () => {
    expect(workbookShellSlots).toEqual([
      "top-bar",
      "view-bar",
      "primary-grid",
      "inspector",
      "status-strip",
    ]);
    const systemViewSwitcherGroupTokens = [
      "scope-indicators",
      "coordination",
      "review-learning",
      "optional-artifact-surfaces",
    ] as const;
    expect(systemViewSwitcherGroupTokens).toEqual([
      "scope-indicators",
      "coordination",
      "review-learning",
      "optional-artifact-surfaces",
    ]);
    expect(
      workbookShellSlots.map((slot) => [
        workbookShellSlotTestId(slot),
        workbookShellSlotLabel(slot),
      ]),
    ).toEqual([
      ["workbook-shell-slot-top-bar", "Workbook top bar"],
      ["workbook-shell-slot-view-bar", "View controls"],
      ["workbook-shell-slot-primary-grid", "Primary grid"],
      ["workbook-shell-slot-inspector", "Inspector"],
      ["workbook-shell-slot-status-strip", "Status strip"],
    ]);
    expect(() => workbookShellSlotLabel("toolbar" as never)).toThrow(
      "Invalid workbook shell slot token: toolbar",
    );
  });

  it("builds exact filter selectors and rejects invalid filter identity", () => {
    const viewSchemaId = "cartulary.view.timeline.v2";
    expect([
      gridFilterFieldTestId(viewSchemaId),
      gridFilterValueTestId(viewSchemaId),
      gridFilterApplyTestId(viewSchemaId),
      gridFilterChipTestId(viewSchemaId, "timeline.capture_state"),
    ]).toEqual([
      "cartulary.view.timeline.v2-filter-field",
      "cartulary.view.timeline.v2-filter-value",
      "cartulary.view.timeline.v2-filter-apply",
      "cartulary.view.timeline.v2-filter-chip-timeline.capture_state",
    ]);
    expect(() => gridFilterChipTestId(viewSchemaId, "Capture state")).toThrow(
      "Invalid field_key selector token: Capture state",
    );
    expect(() => gridFilterApplyTestId("timeline")).toThrow(
      "Invalid view_schema_id selector token: timeline",
    );
  });

  it("parses every density row height and fails closed for a malformed token", () => {
    expect([
      workbookGridRowHeightPx("compact"),
      workbookGridRowHeightPx("default"),
      workbookGridRowHeightPx("comfortable"),
    ]).toEqual([24, 32, 40]);

    const mutableTokenVars = cartularyDesignTokenVars as unknown as Record<
      string,
      string
    >;
    const tokenName = "--ct-density-compact-rowHeight";
    const originalValue = mutableTokenVars[tokenName];
    expect(originalValue).toBe("24px");
    try {
      mutableTokenVars[tokenName] = "calc(28px)";
      expect(() => workbookGridRowHeightPx("compact")).toThrow(
        "Invalid fixed grid row-height token --ct-density-compact-rowHeight: calc(28px)",
      );
    } finally {
      if (originalValue !== undefined)
        mutableTokenVars[tokenName] = originalValue;
    }
  });
});
