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
    expect(
      normalizeSavedViewResource({
        saved_view_id: "sv-1",
        view_schema_id: "cartulary.view.timeline.v2",
        display_name: "Analyst timeline",
        scope: "private",
        query_json: { filters: [] },
        layout_json: { hidden_field_keys: ["timeline.raw_activity_text"] },
        owner_user_id: "user-1",
        saved_view_version: 7,
      }),
    ).toEqual({
      saved_view_id: "sv-1",
      view_schema_id: "cartulary.view.timeline.v2",
      display_name: "Analyst timeline",
      scope: "private",
      query_json: { filters: [] },
      layout_json: { hidden_field_keys: ["timeline.raw_activity_text"] },
      owner_user_id: "user-1",
      saved_view_version: 7,
    });

    expect(
      normalizeSavedViewResource({
        saved_view_id: "sv-2",
        view_schema_id: "cartulary.view.unknown.v1",
        display_name: "Unknown",
        scope: "private",
      }),
    ).toBeNull();
  });

  it("keeps system saved views immutable while allowing admins and owners to mutate user views", () => {
    const base = normalizeSavedViewResource({
      saved_view_id: "sv-1",
      view_schema_id: "cartulary.view.timeline.v2",
      display_name: "Analyst timeline",
      scope: "private",
      owner_user_id: "user-1",
    });
    const system = normalizeSavedViewResource({
      saved_view_id: "sv-system",
      view_schema_id: "cartulary.view.timeline.v2",
      display_name: "System timeline",
      scope: "system",
    });

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
