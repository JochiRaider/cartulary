import { describe, expect, it } from "vitest";
import {
  acceptWorkbookSavedViewPage,
  normalizeWorkbookSavedViewPage,
  startWorkbookSavedViewPagination,
  workbookSavedViewPaginationIsCurrent,
} from "./workbookSavedViewPaginationMachine";

function savedView(id: string) {
  return {
    display_name: id,
    layout_json: {},
    owner_user_id: "user-1",
    query_json: {},
    saved_view_id: id,
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

  it("accumulates ordered pages and publishes only on the terminal page", () => {
    const initial = startWorkbookSavedViewPagination(3);
    const first = acceptWorkbookSavedViewPage(initial, {
      nextCursor: "cursor-2",
      savedViews: [savedView("saved-1")],
    });
    expect(first.kind).toBe("continue");
    if (first.kind !== "continue") throw new Error("expected continuation");
    const complete = acceptWorkbookSavedViewPage(first.machine, {
      nextCursor: null,
      savedViews: [savedView("saved-2")],
    });
    expect(complete).toEqual({
      kind: "complete",
      savedViews: [savedView("saved-1"), savedView("saved-2")],
    });
    expect(workbookSavedViewPaginationIsCurrent(first.machine, 3)).toBe(true);
    expect(workbookSavedViewPaginationIsCurrent(first.machine, 4)).toBe(false);
  });

  it("rejects duplicate resources, cursor cycles, and malformed resources", () => {
    const first = acceptWorkbookSavedViewPage(
      startWorkbookSavedViewPagination(1),
      {
        nextCursor: "cursor-2",
        savedViews: [savedView("saved-1")],
      },
    );
    if (first.kind !== "continue") throw new Error("expected continuation");
    expect(
      acceptWorkbookSavedViewPage(first.machine, {
        nextCursor: null,
        savedViews: [savedView("saved-1")],
      }),
    ).toEqual({
      kind: "invalid",
      message: "Saved-view listing returned a duplicate resource.",
    });
    expect(
      acceptWorkbookSavedViewPage(first.machine, {
        nextCursor: "cursor-2",
        savedViews: [],
      }),
    ).toEqual({
      kind: "invalid",
      message: "Saved-view listing returned a cyclic cursor.",
    });
    expect(
      acceptWorkbookSavedViewPage(startWorkbookSavedViewPagination(1), {
        nextCursor: null,
        savedViews: [{ ...savedView("saved-1"), scope: "future" } as never],
      }),
    ).toEqual({
      kind: "invalid",
      message: "Saved-view listing returned an invalid resource.",
    });
  });
});
