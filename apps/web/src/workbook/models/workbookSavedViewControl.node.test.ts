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

describe("workbookSavedViewControl", () => {
  it("projects selected identity independently through pending failed and unavailable observations", () => {
    const ref = { kind: "saved_view" as const, id: timelineView.saved_view_id };
    expect(workbookSavedViewsResource(undefined, ref)).toMatchObject({
      kind: "loading",
      selectedSavedViewId: ref.id,
    });
    const stale = workbookSavedViewsResource(
      {
        resource: timelineView,
        status: "ready",
        pending: false,
        problem: { kind: "transport", message: "Read failed" },
        revision: 1,
      },
      ref,
    );
    expect(
      projectActiveSurfaceSavedViews(stale, timelineView.view_schema_id, ref),
    ).toMatchObject({
      selectedSavedView: timelineView,
      resourceMessage: "Read failed",
    });
    expect(
      workbookSavedViewsResource(
        {
          resource: null,
          status: "unavailable",
          pending: false,
          problem: null,
          revision: 2,
        },
        ref,
      ),
    ).toMatchObject({ kind: "unavailable", selectedSavedViewId: ref.id });
    expect(
      workbookSavedViewsResource(undefined, {
        kind: "view_schema",
        id: timelineView.view_schema_id,
      }),
    ).toMatchObject({
      kind: "ready",
      selectedSavedView: null,
      selectedSavedViewId: "",
    });
  });

  it("parses only the editable scope vocabulary", () => {
    expect(parseSavedViewEditableScope("private")).toBe("private");
    expect(parseSavedViewEditableScope("shared")).toBe("shared");
    expect(parseSavedViewEditableScope("system")).toBeNull();
    expect(parseSavedViewEditableScope("future")).toBeNull();
  });
});
