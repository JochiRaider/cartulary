import { describe, expect, it } from "vitest";

import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";
import {
  normalizeWorkbookStartupSelection,
  resolveWorkbookStartupFallback,
  workbookStartupQueryFromURLParams,
  type WorkbookSheetRef,
} from "./workbookStartup";

const savedViewId = "11111111-1111-4111-8111-111111111111";

function viewSchemaRef(viewSchemaId: string): WorkbookSheetRef {
  return { kind: "view_schema", id: viewSchemaId };
}

function savedViewRef(id: string): WorkbookSheetRef {
  return { kind: "saved_view", id };
}

describe("workbook startup model", () => {
  it("FE-U-P2-01 resolves workbook startup fallback order by stable sheet_ref identity", () => {
    const explicit = resolveWorkbookStartupFallback({
      default: {
        sheetRef: viewSchemaRef(evidenceViewSchemaId),
        valid: true,
      },
      explicit: {
        sheetRef: viewSchemaRef(hostsViewSchemaId),
        valid: true,
      },
      home: {
        selectedViewSchemaId: identitiesViewSchemaId,
        sheetRef: savedViewRef(savedViewId),
        valid: true,
      },
    });
    expect(explicit).toMatchObject({
      selectedSheetRef: viewSchemaRef(hostsViewSchemaId),
      selectedViewSchemaId: hostsViewSchemaId,
      source: "explicit",
    });
    expect(explicit.clearedPointers).toEqual([]);

    const home = resolveWorkbookStartupFallback({
      default: {
        sheetRef: viewSchemaRef(evidenceViewSchemaId),
        valid: true,
      },
      home: {
        selectedViewSchemaId: identitiesViewSchemaId,
        sheetRef: savedViewRef(savedViewId),
        valid: true,
      },
    });
    expect(home).toMatchObject({
      selectedSheetRef: savedViewRef(savedViewId),
      selectedViewSchemaId: identitiesViewSchemaId,
      source: "home",
    });

    const defaultSelection = resolveWorkbookStartupFallback({
      default: {
        sheetRef: viewSchemaRef(evidenceViewSchemaId),
        valid: true,
      },
      home: null,
    });
    expect(defaultSelection).toMatchObject({
      selectedSheetRef: viewSchemaRef(evidenceViewSchemaId),
      selectedViewSchemaId: evidenceViewSchemaId,
      source: "default",
    });

    const timeline = resolveWorkbookStartupFallback({});
    expect(timeline).toMatchObject({
      selectedSheetRef: viewSchemaRef(timelineViewSchemaId),
      selectedViewSchemaId: timelineViewSchemaId,
      source: "timeline",
    });
  });

  it("FE-U-P2-01 skips or clears invalid startup pointers without failing workbook open", () => {
    const defaultAfterInvalid = resolveWorkbookStartupFallback({
      default: {
        sheetRef: viewSchemaRef(taskRequestsViewSchemaId),
        valid: true,
      },
      explicit: {
        invalidReasonCode: "unknown_view_schema",
        sheetRef: viewSchemaRef("cartulary.view.unknown.v1"),
        valid: false,
      },
      home: {
        invalidReasonCode: "saved_view_not_visible",
        selectedViewSchemaId: hostsViewSchemaId,
        sheetRef: savedViewRef("22222222-2222-4222-8222-222222222222"),
        valid: false,
      },
    });
    expect(defaultAfterInvalid).toMatchObject({
      homeSheetRef: null,
      selectedSheetRef: viewSchemaRef(taskRequestsViewSchemaId),
      selectedViewSchemaId: taskRequestsViewSchemaId,
      source: "default",
    });
    expect(defaultAfterInvalid.clearedPointers).toEqual([
      {
        reasonCode: "saved_view_not_visible",
        sheetRef: savedViewRef("22222222-2222-4222-8222-222222222222"),
        source: "home",
      },
    ]);

    const timelineAfterInvalid = resolveWorkbookStartupFallback({
      default: {
        invalidReasonCode: "saved_view_not_found",
        selectedViewSchemaId: evidenceViewSchemaId,
        sheetRef: savedViewRef("33333333-3333-4333-8333-333333333333"),
        valid: false,
      },
      home: {
        invalidReasonCode: "invalid_saved_view_id",
        sheetRef: savedViewRef("not-a-uuid"),
        valid: false,
      },
    });
    expect(timelineAfterInvalid).toMatchObject({
      defaultSheetRef: null,
      homeSheetRef: null,
      selectedSheetRef: viewSchemaRef(timelineViewSchemaId),
      selectedViewSchemaId: timelineViewSchemaId,
      source: "timeline",
    });
    expect(timelineAfterInvalid.clearedPointers.map((entry) => entry.source)).toEqual([
      "home",
      "default",
    ]);
    expect(
      timelineAfterInvalid.clearedPointers.map((entry) => entry.reasonCode),
    ).toEqual(["invalid_saved_view_id", "saved_view_not_found"]);

    expect(
      workbookStartupQueryFromURLParams(
        new URLSearchParams({
          ignored: "not-forwarded",
          sheet_ref_id: savedViewId,
          sheet_ref_kind: "saved_view",
        }),
      ),
    ).toBe(`?sheet_ref_kind=saved_view&sheet_ref_id=${savedViewId}`);
  });

  it("FE-U-P2-01 applies selected saved-view identity without collapsing base view_schema_id", () => {
    const selected = normalizeWorkbookStartupSelection({
      cleared_pointers: [],
      default_sheet_ref: null,
      home_sheet_ref: savedViewRef(savedViewId),
      incident_id: "incident-1",
      selected_saved_view: {
        saved_view_id: savedViewId,
        view_schema_id: evidenceViewSchemaId,
      },
      selected_sheet_ref: savedViewRef(savedViewId),
      selected_view_schema_id: evidenceViewSchemaId,
      source: "home",
    });

    expect(selected).toMatchObject({
      selectedSheetRef: savedViewRef(savedViewId),
      selectedViewSchemaId: evidenceViewSchemaId,
      source: "home",
    });
    expect(selected?.selectedSheetRef.id).not.toBe(
      selected?.selectedViewSchemaId,
    );
    expect(selected?.selectedSavedView).toMatchObject({
      saved_view_id: savedViewId,
      view_schema_id: evidenceViewSchemaId,
    });
  });
});
