import { describe, expect, it } from "vitest";

import {
  cartularyDesignTokenVars,
  gridActionsHeaderTestId,
  gridDataCellsSelector,
  gridDataRowsSelector,
  gridDraftRowSelector,
  gridFillHandleSelector,
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowSelector,
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
  workbookGridDensityMetrics,
  workbookGridRowHeightPx,
  workbookInlineDraftRowTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
  workbookLayoutMetrics,
  workbookRowActionMenuButtonTestId,
  workbookRowContextMenuTestId,
  workbookShellSlotLabel,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "./index";

describe("@cartulary/ui-contracts workbook shell and grid selectors", () => {
  it("builds owner-backed group row selectors for both expansion states", () => {
    expect(gridGroupRowSelector()).toBe(
      '[role="row"][aria-level="1"][aria-expanded]',
    );
    expect(gridGroupRowSelector(undefined)).toBe(
      '[role="row"][aria-level="1"][aria-expanded]',
    );
    expect(gridGroupRowSelector(true)).toBe(
      '[role="row"][aria-level="1"][aria-expanded="true"]',
    );
    expect(gridGroupRowSelector(false)).toBe(
      '[role="row"][aria-level="1"][aria-expanded="false"]',
    );
  });

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
    expect(gridFillHandleSelector()).toBe(
      '[data-cartulary-fill-handle="true"]',
    );
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
    expect(workbookGridDensityMetrics("compact")).toEqual({
      cellPaddingBlockCssPx: 2,
      cellPaddingInlineCssPx: 5,
      fontSizeCssPx: 12,
      lineHeight: 1.2,
      rowHeightCssPx: 24,
    });
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
        "Invalid fixed grid density token --ct-density-compact-rowHeight: calc(28px)",
      );
    } finally {
      if (originalValue !== undefined)
        mutableTokenVars[tokenName] = originalValue;
    }

    const paddingTokenName = "--ct-density-compact-cellPadding";
    const originalPadding = mutableTokenVars[paddingTokenName];
    try {
      mutableTokenVars[paddingTokenName] = "2px calc(5px)";
      expect(() => workbookGridDensityMetrics("compact")).toThrow(
        "Invalid grid density cell-padding token --ct-density-compact-cellPadding: 2px calc(5px)",
      );
    } finally {
      if (originalPadding !== undefined)
        mutableTokenVars[paddingTokenName] = originalPadding;
    }
  });

  it("parses fixed and viewport-relative layout metrics from generated tokens", () => {
    expect(workbookLayoutMetrics(1280)).toEqual({
      baseMinHeightCssPx: 720,
      baseMinWidthCssPx: 1280,
      compactMinHeightCssPx: 640,
      compactMinWidthCssPx: 768,
      inspectorDefaultWidthCssPx: 420,
      inspectorEffectiveMaxWidthCssPx: 560,
      inspectorMinWidthCssPx: 360,
      narrowMinWidthCssPx: 1024,
    });
    expect(workbookLayoutMetrics(1024).inspectorEffectiveMaxWidthCssPx).toBe(
      460.8,
    );
    expect(workbookLayoutMetrics(768).inspectorEffectiveMaxWidthCssPx).toBe(
      360,
    );

    const mutableTokenVars = cartularyDesignTokenVars as unknown as Record<
      string,
      string
    >;
    const tokenName = "--ct-layout-inspectorMaxWidth";
    const originalValue = mutableTokenVars[tokenName];
    try {
      mutableTokenVars[tokenName] = "min(560px,45vw)";
      expect(() => workbookLayoutMetrics(1280)).toThrow(
        "Invalid css_min_px_vw_v1 token --ct-layout-inspectorMaxWidth: min(560px,45vw)",
      );
    } finally {
      if (originalValue !== undefined)
        mutableTokenVars[tokenName] = originalValue;
    }
  });
});
