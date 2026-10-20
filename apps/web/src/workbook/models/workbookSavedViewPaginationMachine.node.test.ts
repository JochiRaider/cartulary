import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { buildSavedViewLayoutJson } from "./workbookQuery";
import { normalizeWorkbookSavedViewPage } from "./workbookSavedViewPaginationMachine";

function savedView(id: string) {
  return {
    display_name: id,
    layout_json: buildSavedViewLayoutJson(
      requireViewContract("cartulary.view.timeline.v2"),
    ),
    owner_user_id: "user-1",
    query_json: { sort: [], filters: [] },
    saved_view_id: id,
    incident_id: "incident-1",
    created_at: "2026-07-31T20:00:00Z",
    updated_at: "2026-07-31T20:00:00Z",
    saved_view_version: 1,
    scope: "private" as const,
    view_schema_id: "cartulary.view.timeline.v2",
  };
}

describe("Workbook saved-view pagination machine", () => {
  it("validates adapter pages against incident, limit, cursor, and resources", () => {
    expect(
      normalizeWorkbookSavedViewPage({
        incidentId: "incident-1",
        limit: 100,
        paging: { has_more: true, limit: 100, next_cursor: "cursor-2" },
        savedViews: [{ ...savedView("saved-1"), incident_id: "incident-1" }],
      }),
    ).toEqual({ nextCursor: "cursor-2", savedViews: [savedView("saved-1")] });
    expect(
      normalizeWorkbookSavedViewPage({
        incidentId: "incident-1",
        limit: 100,
        paging: { has_more: false, limit: 99, next_cursor: null },
        savedViews: [],
      }),
    ).toBeNull();
    expect(
      normalizeWorkbookSavedViewPage({
        incidentId: "incident-1",
        limit: 100,
        paging: { has_more: false, limit: 100, next_cursor: null },
        savedViews: [{ ...savedView("saved-1"), incident_id: "incident-2" }],
      }),
    ).toBeNull();
  });

  it("rejects duplicate resources schema mismatch and inconsistent continuation without sorting", () => {
    const input = {
      incidentId: "incident-1",
      viewSchemaId: "cartulary.view.timeline.v2",
      limit: 50,
      paging: { has_more: false, limit: 50, next_cursor: null },
      savedViews: [savedView("z"), savedView("a")],
    };
    expect(
      normalizeWorkbookSavedViewPage(input)?.savedViews.map(
        (r) => r.saved_view_id,
      ),
    ).toEqual(["z", "a"]);
    expect(
      normalizeWorkbookSavedViewPage({
        ...input,
        savedViews: [savedView("same"), savedView("same")],
      }),
    ).toBeNull();
    expect(
      normalizeWorkbookSavedViewPage({
        ...input,
        viewSchemaId: "cartulary.view.notes.v1",
      }),
    ).toBeNull();
    expect(
      normalizeWorkbookSavedViewPage({
        ...input,
        savedViews: [],
        paging: { has_more: true, limit: 50, next_cursor: "next" },
      }),
    ).toBeNull();
    expect(
      normalizeWorkbookSavedViewPage({
        ...input,
        paging: { has_more: false, limit: 50, next_cursor: "next" },
      }),
    ).toBeNull();
    expect(
      normalizeWorkbookSavedViewPage({
        ...input,
        savedViews: [{ ...savedView("x"), saved_view_version: 0 }],
      }),
    ).toBeNull();
  });
});
