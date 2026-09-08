import { describe, expect, it } from "vitest";
import {
  parseSavedViewEditableScope,
  projectActiveSurfaceSavedViews,
  workbookSavedViewsResource,
} from "./workbookSavedViewControl";

const timelineView = {
  saved_view_id: "saved-timeline",
  view_schema_id: "cartulary.view.timeline.v2",
  display_name: "Timeline view",
  scope: "private",
  query_json: {},
  layout_json: {},
  owner_user_id: "user-1",
  incident_id: "incident-1",
  created_at: "2026-07-31T20:00:00Z",
  updated_at: "2026-07-31T20:00:00Z",
  saved_view_version: 4,
} as const;

const evidenceView = {
  ...timelineView,
  saved_view_id: "saved-evidence",
  view_schema_id: "cartulary.view.evidence.v1",
  display_name: "Evidence view",
  scope: "shared",
} as const;

describe("workbookSavedViewControl", () => {
  it("projects loading, unavailable, invalid, and surface-filtered ready resources", () => {
    const loading = projectActiveSurfaceSavedViews(
      { kind: "loading" },
      timelineView.view_schema_id,
      { kind: "view_schema", id: timelineView.view_schema_id },
    );
    expect(loading).toMatchObject({
      resourceKind: "loading",
      resourceMessage: "Loading saved views…",
      savedViews: [],
    });

    const unavailable = projectActiveSurfaceSavedViews(
      { kind: "unavailable", message: "Listing unavailable." },
      timelineView.view_schema_id,
      { kind: "view_schema", id: timelineView.view_schema_id },
    );
    expect(unavailable.resourceMessage).toBe("Listing unavailable.");

    const invalid = workbookSavedViewsResource([timelineView, evidenceView], {
      kind: "saved_view",
      id: "removed-view",
    });
    expect(invalid).toEqual({
      kind: "invalid_selection",
      savedViews: [timelineView, evidenceView],
      selectedSavedViewId: "removed-view",
    });

    const ready = projectActiveSurfaceSavedViews(
      { kind: "ready", savedViews: [timelineView, evidenceView] },
      timelineView.view_schema_id,
      { kind: "saved_view", id: timelineView.saved_view_id },
    );
    expect(ready.savedViews).toEqual([timelineView]);
    expect(ready.privateSavedViews).toEqual([timelineView]);
    expect(ready.sharedSavedViews).toEqual([]);
    expect(ready.selectedSavedView).toEqual(timelineView);
  });

  it("parses only the editable scope vocabulary", () => {
    expect(parseSavedViewEditableScope("private")).toBe("private");
    expect(parseSavedViewEditableScope("shared")).toBe("shared");
    expect(parseSavedViewEditableScope("system")).toBeNull();
    expect(parseSavedViewEditableScope("future")).toBeNull();
  });
});
