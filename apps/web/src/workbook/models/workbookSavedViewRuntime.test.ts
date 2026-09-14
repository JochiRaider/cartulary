import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
} from "./workbookQuery";
import {
  fallbackIdentityAfterSavedViewDelete,
  savedViewChanges,
  savedViewIdentityForSelection,
  savedViewQueryStateForRuntime,
  upsertSavedViewList,
} from "./workbookSavedViewRuntime";
import type { SavedViewResource } from "./workbookSavedViews";
import {
  notesViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

function savedView(
  overrides: Partial<SavedViewResource> &
    Pick<SavedViewResource, "saved_view_id" | "view_schema_id">,
): SavedViewResource {
  return {
    display_name: "Saved",
    layout_json: {},
    owner_user_id: "user-1",
    query_json: {},
    incident_id: "incident-1",
    created_at: "2026-07-31T20:00:00Z",
    updated_at: "2026-07-31T20:00:00Z",
    saved_view_version: 1,
    scope: "private",
    ...overrides,
  };
}

describe("workbookSavedViewRuntime", () => {
  it("omits structurally equal layout and includes a genuine portable layout change", () => {
    const contract = requireViewContract(timelineViewSchemaId);
    const layout = buildSavedViewLayoutJson(contract, {
      hiddenFieldKeys: ["timeline.analyst_text"],
    });
    const query = buildSavedViewQueryJson(contract, {
      filters: [],
      sort: [],
      groupBy: null,
    });
    const base = savedView({
      saved_view_id: "saved-1",
      view_schema_id: timelineViewSchemaId,
      layout_json: layout,
      query_json: query,
    });
    const definition = {
      displayName: base.display_name,
      scope: "private" as const,
      queryJson: query,
      layoutJson: {
        column_widths: layout.column_widths,
        hidden_field_keys: layout.hidden_field_keys,
        column_order: layout.column_order,
        layout_schema_id: layout.layout_schema_id,
      },
    };
    expect(savedViewChanges(base, definition)).toEqual({});
    expect(
      savedViewChanges(base, {
        ...definition,
        layoutJson: { ...layout, hidden_field_keys: [] },
      }),
    ).toEqual({ layoutJson: { ...layout, hidden_field_keys: [] } });
    expect(base.layout_json).toEqual(layout);
  });
  it("upserts saved views by identity and keeps display-name ordering", () => {
    const alpha = savedView({
      display_name: "Alpha",
      saved_view_id: "saved-alpha",
      view_schema_id: timelineViewSchemaId,
    });
    const beta = savedView({
      display_name: "Beta",
      saved_view_id: "saved-beta",
      view_schema_id: notesViewSchemaId,
    });
    const updatedBeta = { ...beta, display_name: "Aardvark" };

    expect(
      upsertSavedViewList([beta], alpha).map((item) => item.display_name),
    ).toEqual(["Alpha", "Beta"]);
    expect(
      upsertSavedViewList([alpha, beta], updatedBeta).map(
        (item) => item.display_name,
      ),
    ).toEqual(["Aardvark", "Alpha"]);
  });

  it("preserves saved-view sheet_ref and base view_schema identity separately", () => {
    const selected = savedView({
      saved_view_id: "saved-1",
      view_schema_id: notesViewSchemaId,
    });
    expect(savedViewIdentityForSelection(selected)).toEqual({
      sheetRef: { kind: "saved_view", id: "saved-1" },
      viewSchemaId: notesViewSchemaId,
    });
    expect(
      fallbackIdentityAfterSavedViewDelete(
        { kind: "saved_view", id: "saved-1" },
        selected,
      ),
    ).toEqual({
      sheetRef: { kind: "view_schema", id: notesViewSchemaId },
      viewSchemaId: notesViewSchemaId,
    });
    expect(
      fallbackIdentityAfterSavedViewDelete(
        { kind: "view_schema", id: notesViewSchemaId },
        selected,
      ),
    ).toBeNull();
  });

  it("derives runtime query state from saved-view query_json without mutating identity", () => {
    const contract = requireViewContract(timelineViewSchemaId);
    const queryState = savedViewQueryStateForRuntime(contract, {
      query_json: {
        filters: [
          {
            arg: { value: "reviewed" },
            field_key: "timeline.capture_state",
            op: "eq",
          },
        ],
        sort: [{ direction: "asc", field_key: "timeline.activity_sort_ts" }],
      },
    });
    expect(queryState.filters).toEqual([
      {
        arg: { value: "reviewed" },
        fieldKey: "timeline.capture_state",
        op: "eq",
      },
    ]);
    expect(queryState.sort).toEqual([
      { direction: "asc", fieldKey: "timeline.activity_sort_ts" },
    ]);
  });
});
