import { describe, expect, it } from "vitest";

import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  dataTestIdPrefixSelector,
  dataTestIdSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  mentionCreateEntityButtonTestId,
  relationshipChipTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryItemTestId,
  rowHistoryRollbackPreviewTestId,
  surfaceTabTestId,
  systemViewSwitcherTriggerTestId,
  timelineScalarEditorTestId,
} from "./index";

const requireFixtureValue = <T>(value: T | undefined, label: string): T => {
  if (value === undefined) {
    throw new Error(`Missing fixture value: ${label}`);
  }
  return value;
};

describe("@cartulary/ui-contracts selector core", () => {
  it("returns deterministic selectors for identical stable identifiers", () => {
    const first = {
      cell: rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      shell: gridShellTestId("cartulary.view.timeline.v2"),
      tab: surfaceTabTestId("cartulary.view.timeline.v2"),
    };
    const second = {
      cell: rowCellTestId("record-1", "timeline.activity_synopsis_text"),
      shell: gridShellTestId("cartulary.view.timeline.v2"),
      tab: surfaceTabTestId("cartulary.view.timeline.v2"),
    };

    expect(second).toEqual(first);
  });

  it("derives surface selectors from view_schema_id instead of visible titles", () => {
    const originalSurface = {
      title: "Timeline",
      viewSchemaId: "cartulary.view.timeline.v2",
    };
    const renamedSurface = {
      title: "Activity",
      viewSchemaId: "cartulary.view.timeline.v2",
    };

    expect(surfaceTabTestId(originalSurface.viewSchemaId)).toBe(
      "surface-tab-cartulary.view.timeline.v2",
    );
    expect(surfaceTabTestId(renamedSurface.viewSchemaId)).toBe(
      surfaceTabTestId(originalSurface.viewSchemaId),
    );
    expect(gridShellTestId(renamedSurface.viewSchemaId)).toBe(
      gridShellTestId(originalSurface.viewSchemaId),
    );
    expect(systemViewSwitcherTriggerTestId()).toBe("system-view-selector");
  });

  it("keeps tab, field, column, row, and fixture reordering out of selector identity", () => {
    const surfaces = [
      { title: "Timeline", viewSchemaId: "cartulary.view.timeline.v2" },
      { title: "Evidence", viewSchemaId: "cartulary.view.evidence.v1" },
    ];
    const fields = [
      { fieldKey: "timeline.activity_synopsis_text", label: "Summary" },
      { fieldKey: "timeline.raw_activity_text", label: "Details" },
    ];
    const rows = [
      { displayName: "Alpha", recordId: "record-alpha" },
      { displayName: "Beta", recordId: "record-beta" },
    ];
    const timelineSurface = requireFixtureValue(surfaces[0], "timeline");
    const evidenceSurface = requireFixtureValue(surfaces[1], "evidence");
    const summaryField = requireFixtureValue(fields[0], "summary");
    const detailsField = requireFixtureValue(fields[1], "details");
    const betaRow = requireFixtureValue(rows[1], "beta row");

    const selectors = {
      betaCell: rowCellTestId(betaRow.recordId, summaryField.fieldKey),
      detailsHeader: gridSortHeaderTestId(
        timelineSurface.viewSchemaId,
        detailsField.fieldKey,
      ),
      evidenceTab: surfaceTabTestId(evidenceSurface.viewSchemaId),
    };

    const reorderedSurfaces = [...surfaces].reverse();
    const reorderedFields = [...fields].reverse();
    const reorderedRows = [...rows].reverse();

    expect(
      surfaceTabTestId(
        reorderedSurfaces.find((surface) => surface.title === "Evidence")
          ?.viewSchemaId ?? "",
      ),
    ).toBe(selectors.evidenceTab);
    expect(
      gridSortHeaderTestId(
        timelineSurface.viewSchemaId,
        reorderedFields.find((field) => field.label === "Details")?.fieldKey ??
          "",
      ),
    ).toBe(selectors.detailsHeader);
    expect(
      rowCellTestId(
        reorderedRows.find((row) => row.displayName === "Beta")?.recordId ?? "",
        summaryField.fieldKey,
      ),
    ).toBe(selectors.betaCell);
  });

  it("rejects malformed stable IDs and keeps selector segments collision-resistant", () => {
    expect(() => gridShellTestId("timeline")).toThrow(
      "Invalid view_schema_id selector token: timeline",
    );
    expect(() => gridShellTestId("")).toThrow(
      "Invalid view_schema_id selector token: ",
    );
    expect(() => rowCellTestId("record-1", "timeline summary")).toThrow(
      "Invalid field_key selector token: timeline summary",
    );
    expect(() => rowCellTestId("record-1", "timeline..summary")).toThrow(
      "Invalid field_key selector token: timeline..summary",
    );
    expect(() => rowCellTestId("", "timeline.activity_synopsis_text")).toThrow(
      "Invalid record_id selector token: ",
    );
    expect(() =>
      rowCellTestId("   ", "timeline.activity_synopsis_text"),
    ).toThrow("Invalid record_id selector token:    ");
    expect(() =>
      rowHistoryActionTestId({
        action: "restore" as never,
        historyItemRef: "hitem_change_set_1",
      }),
    ).toThrow("Invalid row history rollback action token: restore");
    expect(() => mentionCreateEntityButtonTestId("account" as never)).toThrow(
      "Invalid entity type selector token: account",
    );
    expect(() =>
      rowHistoryActionTestId({
        action: "rollback" as never,
        historyItemRef: "hitem_change_set_1",
      }),
    ).toThrow("Invalid row history rollback action token: rollback");
    expect(() =>
      rowHistoryDestructiveConfirmPanelTestId({
        operation: "remove" as never,
      }),
    ).toThrow("Invalid row history destructive operation token: remove");
    expect(() => rowHistoryItemTestId({ historyItemRef: "" })).toThrow(
      "Invalid history_item_ref selector token: ",
    );
    expect(() => rowHistoryItemTestId({ historyItemRef: "   " })).toThrow(
      "Invalid history_item_ref selector token:    ",
    );

    expect(rowHistoryItemTestId({ historyItemRef: "h:item/1?x=y#z" })).toBe(
      "row-history-item-h%3Aitem%2F1%3Fx%3Dy%23z",
    );
    expect(
      rowHistoryActionTestId({
        action: "history_entry",
        historyItemRef: "h:item/1?x=y#z",
      }),
    ).toBe("row-history-action-h%3Aitem%2F1%3Fx%3Dy%23z-history_entry");
    expect(
      rowHistoryRollbackPreviewTestId({
        action: "history_entry",
        historyItemRef: "h:item/1?x=y#z",
      }),
    ).toBe(
      "row-history-rollback-preview-h%3Aitem%2F1%3Fx%3Dy%23z-history_entry",
    );
    expect(rowHistoryItemTestId({ historyItemRef: "a:b" })).not.toBe(
      rowHistoryItemTestId({ historyItemRef: "a-b" }),
    );
    expect(
      rowCellTestId("record/1?x=y#z", "timeline.activity_synopsis_text"),
    ).toBe("row-record%2F1%3Fx%3Dy%23z-timeline.activity_synopsis_text");
    expect(
      conflictMarkerTestId("record/1?x=y#z", "timeline.activity_synopsis_text"),
    ).toBe(
      "conflict-marker-record%2F1%3Fx%3Dy%23z-timeline.activity_synopsis_text",
    );
    expect(
      cellPresenceMarkerTestId(
        "record/1?x=y#z",
        "timeline.activity_synopsis_text",
      ),
    ).toBe(
      "presence-cell-record%2F1%3Fx%3Dy%23z-timeline.activity_synopsis_text",
    );

    expect(relationshipChipTestId("a:b")).not.toBe(
      relationshipChipTestId("a-b"),
    );
    expect(relationshipChipTestId("a b")).not.toBe(
      relationshipChipTestId("a-b"),
    );
  });

  it("validates view_schema_id selectors against the generated registry", () => {
    for (const viewSchemaId of [
      "cartulary.view.timeline.v2",
      "cartulary.view.comm_log.v1",
      "cartulary.view.findings.v1",
      "cartulary.view.forensic_keywords.v1",
    ]) {
      expect(gridShellTestId(viewSchemaId)).toBe(`${viewSchemaId}-grid-shell`);
    }

    expect(() => gridShellTestId("cartulary.view.future.v1")).toThrow(
      "Unknown view_schema_id selector token: cartulary.view.future.v1",
    );
  });

  it("builds escaped data-testid selectors and timeline editor test ids", () => {
    expect(dataTestIdSelector('row-record"1')).toBe(
      '[data-testid="row-record\\"1"]',
    );
    expect(dataTestIdSelector("row\\record")).toBe(
      '[data-testid="row\\\\record"]',
    );
    expect(
      dataTestIdSelector(rowHistoryItemTestId({ historyItemRef: 'h"1\\2' })),
    ).toBe('[data-testid="row-history-item-h%221%5C2"]');
    expect(dataTestIdPrefixSelector("saved-view-")).toBe(
      '[data-testid^="saved-view-"]',
    );
    expect(() => dataTestIdSelector("")).toThrow(
      "Invalid data-testid selector token: ",
    );

    for (const surface of ["grid", "inspector"] as const) {
      expect(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: "record-1",
          surface,
        }),
      ).toBe(
        surface === "inspector"
          ? "row-record-1-timeline.activity_synopsis_text-inspector"
          : "row-record-1-timeline.activity_synopsis_text-grid-editor",
      );
    }
    expect(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId: null,
      }),
    ).toBe("draft-row-timeline.activity_synopsis_text");
  });
});
