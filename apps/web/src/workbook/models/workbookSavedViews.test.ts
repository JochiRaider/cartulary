import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildSavedViewLayoutJson,
  workbookLayoutStateFromSavedViewLayoutJson,
} from "./workbookQuery";

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
      layout_json: buildSavedViewLayoutJson(
        requireViewContract("cartulary.view.timeline.v2"),
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
      layout_json: buildSavedViewLayoutJson(
        requireViewContract("cartulary.view.timeline.v2"),
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

    const current = buildSavedViewLayoutJson(contract, {
      columnWidths: [
        { fieldKey: "timeline.activity_synopsis_text", widthPx: 320 },
      ],
      hiddenFieldKeys: ["timeline.raw_activity_text"],
      frozenThroughFieldKey: "timeline.raw_activity_text",
    });
    expect(savedViewLayoutJsonForPersistence(contract, current)).toEqual(
      current,
    );
    const { frozen_through_field_key: _boundary, ...legacy } = current;
    const old = { ...legacy, layout_schema_id: "cartulary.layout.v1" };
    for (const invalid of [
      old,
      legacy,
      null,
      {},
      { ...current, layout_schema_id: "cartulary.layout.v99" },
      { ...current, frozen_through_field_key: "record_id" },
      { ...current, frozen_through_field_key: "unknown" },
      { ...old, frozen_through_field_key: null },
      { ...current, inspector_open: true },
      {
        ...current,
        column_widths: [
          { field_key: "timeline.activity_synopsis_text", width_px: 100.7 },
        ],
      },
      {
        ...current,
        column_order: [...current.column_order, current.column_order[0]],
      },
    ]) {
      expect(
        workbookLayoutStateFromSavedViewLayoutJson(contract, invalid),
      ).toBeNull();
      expect(() =>
        savedViewLayoutJsonForPersistence(contract, invalid),
      ).toThrow("Invalid saved-view layout");
    }
  });
});
