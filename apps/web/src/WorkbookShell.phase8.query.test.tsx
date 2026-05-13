import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";

import {
  applyFilterDraft,
  buildQueryRequest,
  emptyWorkbookQueryState,
  toggleSortField,
  updateGroupBy,
} from "./workbookQuery";

describe("Phase 8 workbook query controls", () => {
  it("Phase 8 U-8-GRID-01 emits stable schema keys for sort, filter, and group query controls", () => {
    const contract = requireViewContract("cartulary.view.timeline.v1");

    const sortedByHeader = toggleSortField(
      contract,
      emptyWorkbookQueryState(),
      "timeline.occurred_at",
    );
    expect(buildQueryRequest(contract, sortedByHeader)).toEqual({
      sort: [{ direction: "asc", field_key: "timeline.sort_ts" }],
    });

    const nonSortableCollectionHeader = toggleSortField(
      contract,
      sortedByHeader,
      "timeline.tags",
    );
    expect(nonSortableCollectionHeader).toBe(sortedByHeader);

    const grouped = updateGroupBy(
      contract,
      nonSortableCollectionHeader,
      "timeline.capture_state",
    );
    const filtered = applyFilterDraft(grouped, {
      booleanValue: "",
      fieldKey: "timeline.capture_state",
      value: "reviewed",
    });

    const request = buildQueryRequest(contract, filtered);
    expect(request).toEqual({
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [
        { direction: "asc", field_key: "timeline.capture_state" },
        { direction: "asc", field_key: "timeline.sort_ts" },
      ],
    });

    const encoded = JSON.stringify(request);
    expect(encoded).not.toContain("Occurred");
    expect(encoded).not.toContain("Capture State");
    expect(encoded).not.toContain("timeline_items");
    expect(encoded).not.toContain("record_id");
  });

  it("Phase 8 E-8-04 Notes full_text controls submit the exact-token operator", () => {
    const contract = requireViewContract("cartulary.view.notes.v1");

    const filtered = applyFilterDraft(emptyWorkbookQueryState(), {
      booleanValue: "",
      fieldKey: "note.full_text",
      value: "shell alpha shell",
    });

    expect(buildQueryRequest(contract, filtered)).toEqual({
      filters: [
        {
          arg: { query: "shell alpha shell" },
          field_key: "note.full_text",
          op: "full_text",
        },
      ],
    });
  });
});
