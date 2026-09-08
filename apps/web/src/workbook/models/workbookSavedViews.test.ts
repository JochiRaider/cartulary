import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";

import {
  canMutateSavedView,
  normalizeSavedViewResource,
  savedViewLayoutJsonForPersistence,
  savedViewQueryJsonForPersistence,
} from "./workbookSavedViews";

describe("workbookSavedViews", () => {
  it("normalizes saved views without collapsing saved-view identity into view_schema_id", () => {
    const resource = {
      incident_id: "incident-1",
      created_at: "2026-07-31T20:00:00Z",
      updated_at: "2026-07-31T20:00:00Z",
      saved_view_id: "sv-1",
      view_schema_id: "cartulary.view.timeline.v2",
      display_name: "Analyst timeline",
      scope: "private",
      query_json: { filters: [], sort: [] },
      layout_json: savedViewLayoutJsonForPersistence(
        requireViewContract("cartulary.view.timeline.v2"),
        {},
      ),
      owner_user_id: "user-1",
      saved_view_version: 7,
    };
    expect(normalizeSavedViewResource(resource)).toEqual(resource);
    for (const bad of [
      { view_schema_id: "cartulary.view.unknown.v1" },
      { query_json: {} },
      { layout_json: {} },
      { owner_user_id: null },
      { saved_view_version: 0 },
      { display_name: " unnormalized " },
    ]) {
      expect(normalizeSavedViewResource({ ...resource, ...bad })).toBeNull();
    }
  });

  it("keeps system saved views immutable while allowing admins and owners to mutate user views", () => {
    const base = normalizeSavedViewResource({
      incident_id: "incident-1",
      created_at: "2026-07-31T20:00:00Z",
      updated_at: "2026-07-31T20:00:00Z",
      saved_view_id: "sv-1",
      view_schema_id: "cartulary.view.timeline.v2",
      display_name: "Analyst timeline",
      scope: "private",
      query_json: { filters: [], sort: [] },
      layout_json: savedViewLayoutJsonForPersistence(
        requireViewContract("cartulary.view.timeline.v2"),
        {},
      ),
      owner_user_id: "user-1",
      saved_view_version: 1,
    });
    if (!base) throw new Error("Expected valid fixture");
    const system = { ...base, scope: "system" as const, owner_user_id: null };
    expect(canMutateSavedView(system, "user-1", "admin")).toBe(false);
    expect(canMutateSavedView(base, "user-1", "viewer")).toBe(true);
    expect(canMutateSavedView(base, "user-2", "viewer")).toBe(false);
    expect(canMutateSavedView(base, "user-2", "admin")).toBe(true);
  });

  it("canonicalizes saved-view query and layout JSON through the workbook contract", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");

    expect(
      savedViewQueryJsonForPersistence(contract, {
        group_by: "timeline.capture_state",
        sort: [
          { field_key: "timeline.activity_synopsis_text", direction: "desc" },
          { field_key: "timeline.unknown", direction: "asc" },
        ],
      }),
    ).toEqual({
      filters: [],
      group_by: "timeline.capture_state",
      sort: [
        { field_key: "timeline.activity_synopsis_text", direction: "desc" },
      ],
    });

    const layout = savedViewLayoutJsonForPersistence(contract, {
      column_widths: [
        { field_key: "timeline.activity_synopsis_text", width_px: 320 },
        { field_key: "timeline.unknown", width_px: 900 },
      ],
      hidden_field_keys: ["timeline.raw_activity_text", "timeline.unknown"],
    });
    expect(layout.layout_schema_id).toBe("cartulary.layout.v1");
    expect(layout.column_order).toContain("timeline.activity_synopsis_text");
    expect(layout.column_order).not.toContain("timeline.unknown");
    expect(layout.column_widths).toEqual([
      { field_key: "timeline.activity_synopsis_text", width_px: 320 },
    ]);
    expect(layout.hidden_field_keys).toEqual(["timeline.raw_activity_text"]);
    expect(
      JSON.stringify(
        savedViewLayoutJsonForPersistence(contract, {
          active_panel: "history",
          inspector_open: true,
          local_form_state: { dirty: true },
          merge_plans: ["row-1"],
          preview_state: { record_id: "row-1" },
          rollback_previews: ["row-1"],
          stale_confirmation_state: { delete: true },
        }),
      ),
    ).not.toMatch(
      /active_panel|inspector_open|local_form_state|merge_plans|preview_state|rollback_previews|stale_confirmation_state/,
    );
  });
});
