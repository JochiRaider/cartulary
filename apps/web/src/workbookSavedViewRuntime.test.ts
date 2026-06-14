import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  fallbackIdentityAfterSavedViewDelete,
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
    saved_view_version: 1,
    scope: "private",
    ...overrides,
  };
}

describe("workbookSavedViewRuntime", () => {
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
        sort: [{ direction: "asc", field_key: "timeline.sort_ts" }],
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
      { direction: "asc", fieldKey: "timeline.sort_ts" },
    ]);
  });
});
