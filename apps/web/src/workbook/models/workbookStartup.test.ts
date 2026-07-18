import { describe, expect, it } from "vitest";
import {
  normalizeWorkbookStartupSelection,
  resolveWorkbookStartupFallback,
  type WorkbookSheetRef,
  workbookStartupQueryFromURLParams,
} from "./workbookStartup";
import {
  evidenceViewSchemaId,
  findingsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

const savedViewId = "11111111-1111-4111-8111-111111111111";

function viewSchemaRef(viewSchemaId: string): WorkbookSheetRef {
  return { kind: "view_schema", id: viewSchemaId };
}

function savedViewRef(id: string): WorkbookSheetRef {
  return { kind: "saved_view", id };
}

function extensionWorkspaceRef(): WorkbookSheetRef {
  return {
    kind: "extension_workspace",
    extension_profile_id: "network_flow_activity",
    workspace_key: "network_analysis",
  };
}

describe("workbook startup model", () => {
  it("resolves workbook startup fallback order by stable sheet_ref identity", () => {
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

  it("skips or clears invalid startup pointers without failing workbook open", () => {
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
    expect(
      timelineAfterInvalid.clearedPointers.map((entry) => entry.source),
    ).toEqual(["home", "default"]);
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

    expect(
      workbookStartupQueryFromURLParams(
        new URLSearchParams({
          extension_profile_id: "network_flow_activity",
          sheet_ref_id: "network_analysis",
          sheet_ref_kind: "extension_workspace",
          workspace_key: "legacy-alias-is-not-forwarded",
        }),
      ),
    ).toBe(
      "?sheet_ref_kind=extension_workspace&sheet_ref_id=network_analysis&extension_profile_id=network_flow_activity",
    );
  });

  it("rejects unsupported startup sheet_ref kinds at the frontend boundary", () => {
    expect(
      workbookStartupQueryFromURLParams(
        new URLSearchParams({
          sheet_ref_id: "legacy-surface",
          sheet_ref_kind: "legacy_workspace",
        }),
      ),
    ).toBe("?sheet_ref_kind=legacy_workspace&sheet_ref_id=legacy-surface");

    expect(
      normalizeWorkbookStartupSelection({
        cleared_pointers: [],
        default_sheet_ref: null,
        home_sheet_ref: null,
        incident_id: "incident-1",
        selected_saved_view: null,
        selected_sheet_ref: { kind: "legacy_workspace", id: "legacy-surface" },
        selected_view_schema_id: timelineViewSchemaId,
        source: "explicit",
      }),
    ).toBeNull();

    const normalized = normalizeWorkbookStartupSelection({
      cleared_pointers: [
        {
          reason_code: "unsupported_sheet_ref_kind",
          sheet_ref: { kind: "legacy_workspace", id: "legacy-surface" },
          source: "home",
        },
      ],
      default_sheet_ref: null,
      home_sheet_ref: null,
      incident_id: "incident-1",
      selected_saved_view: null,
      selected_sheet_ref: viewSchemaRef(timelineViewSchemaId),
      selected_view_schema_id: timelineViewSchemaId,
      source: "timeline",
    });
    expect(normalized?.clearedPointers).toEqual([
      {
        reasonCode: "unsupported_sheet_ref_kind",
        sheetRef: { kind: "legacy_workspace", id: "legacy-surface" },
        source: "home",
      },
    ]);
  });

  it("treats deleted saved-view pointers as unavailable saved views", () => {
    const deletedSavedView = resolveWorkbookStartupFallback({
      default: {
        invalidReasonCode: "saved_view_not_found",
        selectedViewSchemaId: evidenceViewSchemaId,
        sheetRef: savedViewRef("44444444-4444-4444-8444-444444444444"),
        valid: false,
      },
    });

    expect(deletedSavedView).toMatchObject({
      defaultSheetRef: null,
      selectedSheetRef: viewSchemaRef(timelineViewSchemaId),
      selectedViewSchemaId: timelineViewSchemaId,
      source: "timeline",
    });
    expect(deletedSavedView.clearedPointers).toEqual([
      {
        reasonCode: "saved_view_not_found",
        sheetRef: savedViewRef("44444444-4444-4444-8444-444444444444"),
        source: "default",
      },
    ]);
  });

  it("preserves required-reference-pack unavailable startup reasons", () => {
    const packUnavailable = resolveWorkbookStartupFallback({
      home: {
        invalidReasonCode: "required_reference_pack_unavailable",
        sheetRef: viewSchemaRef(findingsViewSchemaId),
        valid: false,
      },
    });

    expect(packUnavailable).toMatchObject({
      homeSheetRef: null,
      selectedSheetRef: viewSchemaRef(timelineViewSchemaId),
      selectedViewSchemaId: timelineViewSchemaId,
      source: "timeline",
    });
    expect(packUnavailable.clearedPointers).toEqual([
      {
        reasonCode: "required_reference_pack_unavailable",
        sheetRef: viewSchemaRef(findingsViewSchemaId),
        source: "home",
      },
    ]);
  });

  it("rejects backend startup selections with non-standardized view_schema_id", () => {
    expect(
      normalizeWorkbookStartupSelection({
        cleared_pointers: [],
        default_sheet_ref: null,
        home_sheet_ref: null,
        incident_id: "incident-1",
        selected_saved_view: null,
        selected_sheet_ref: {
          kind: "view_schema",
          id: "cartulary.view.experimental.v1",
        },
        selected_view_schema_id: "cartulary.view.experimental.v1",
        source: "explicit",
      }),
    ).toBeNull();
  });

  it("applies selected saved-view identity without collapsing base view_schema_id", () => {
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
    if (selected?.selectedSheetRef.kind !== "saved_view") {
      throw new Error("expected saved-view startup identity");
    }
    expect(selected.selectedSheetRef.id).not.toBe(
      selected.selectedViewSchemaId,
    );
    expect(selected?.selectedSavedView).toMatchObject({
      saved_view_id: savedViewId,
      view_schema_id: evidenceViewSchemaId,
    });
  });

  it("preserves extension workspace identity with no base schema or saved view", () => {
    const selected = normalizeWorkbookStartupSelection({
      cleared_pointers: [],
      default_sheet_ref: null,
      home_sheet_ref: extensionWorkspaceRef(),
      incident_id: "incident-1",
      selected_saved_view: null,
      selected_sheet_ref: extensionWorkspaceRef(),
      selected_view_schema_id: null,
      source: "home",
    });

    expect(selected).toMatchObject({
      selectedSavedView: null,
      selectedSheetRef: extensionWorkspaceRef(),
      selectedViewSchemaId: null,
      source: "home",
    });
    expect(
      normalizeWorkbookStartupSelection({
        cleared_pointers: [],
        default_sheet_ref: null,
        home_sheet_ref: null,
        incident_id: "incident-1",
        selected_saved_view: null,
        selected_sheet_ref: {
          ...extensionWorkspaceRef(),
          id: "network_analysis",
        },
        selected_view_schema_id: null,
        source: "explicit",
      }),
    ).toBeNull();
  });
});
