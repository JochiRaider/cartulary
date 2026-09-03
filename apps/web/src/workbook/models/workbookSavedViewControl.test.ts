import { describe, expect, it } from "vitest";
import {
  createSavedViewControlState,
  parseSavedViewEditableScope,
  projectActiveSurfaceSavedViews,
  reduceSavedViewControl,
  sameSavedViewActionIdentity,
  savedViewSurfaceControlState,
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

  it("keeps form state surface-keyed and ignores stale action completion", () => {
    let state = createSavedViewControlState(
      timelineView.view_schema_id,
      timelineView,
    );
    state = reduceSavedViewControl(state, {
      type: "change_name",
      surface: timelineView.view_schema_id,
      displayName: "Draft timeline name",
    });
    state = reduceSavedViewControl(state, {
      type: "activate",
      surface: evidenceView.view_schema_id,
      selectedSavedView: evidenceView,
    });
    state = reduceSavedViewControl(state, {
      type: "activate",
      surface: timelineView.view_schema_id,
      selectedSavedView: timelineView,
    });
    expect(
      savedViewSurfaceControlState(
        state,
        timelineView.view_schema_id,
        timelineView,
      ).displayName,
    ).toBe("Draft timeline name");

    const activeIdentity = {
      surface: timelineView.view_schema_id,
      savedViewId: timelineView.saved_view_id,
      savedViewVersion: timelineView.saved_view_version,
      actionKind: "update",
      generation: 2,
    } as const;
    state = reduceSavedViewControl(state, {
      type: "start_action",
      identity: activeIdentity,
    });
    const staleIdentity = { ...activeIdentity, generation: 1 };
    state = reduceSavedViewControl(state, {
      type: "complete_action",
      identity: staleIdentity,
      feedback: { kind: "success", message: "Stale completion" },
    });
    expect(
      savedViewSurfaceControlState(
        state,
        timelineView.view_schema_id,
        timelineView,
      ).activeAction,
    ).toEqual(activeIdentity);
    expect(sameSavedViewActionIdentity(activeIdentity, staleIdentity)).toBe(
      false,
    );
  });

  it("parses only the editable scope vocabulary", () => {
    expect(parseSavedViewEditableScope("private")).toBe("private");
    expect(parseSavedViewEditableScope("shared")).toBe("shared");
    expect(parseSavedViewEditableScope("system")).toBeNull();
    expect(parseSavedViewEditableScope("future")).toBeNull();
  });
});
