import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";

import {
  applyFilterDraft,
  buildQueryRequest,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  filterChipLabel,
  filterInputMode,
  toggleSortField,
  updateGroupBy,
} from "./workbookQuery";

describe("workbookQuery", () => {
  it("builds tag and boolean filters from the client-local draft state", () => {
    const state = applyFilterDraft(emptyWorkbookQueryState(), {
      booleanValue: "",
      fieldKey: "timeline.tags",
      value: "phish, c2",
    });

    expect(state.filters).toEqual([
      {
        fieldKey: "timeline.tags",
        op: "contains_any",
        arg: {
          values: ["c2", "phish"],
        },
      },
    ]);

    expect(
      applyFilterDraft(emptyWorkbookQueryState(), {
        booleanValue: "true",
        fieldKey: "timeline.has_evidence",
        value: "",
      }).filters,
    ).toEqual([
      {
        fieldKey: "timeline.has_evidence",
        op: "eq",
        arg: {
          value: true,
        },
      },
    ]);
  });

  it("prepends group-by sorting when the user sort does not already cluster grouped rows", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const next = updateGroupBy(
      contract,
      toggleSortField(
        contract,
        emptyWorkbookQueryState(),
        "timeline.activity_synopsis_text",
      ),
      "timeline.capture_state",
    );

    expect(buildQueryRequest(contract, next)).toEqual({
      group_by: "timeline.capture_state",
      sort: [
        { field_key: "timeline.capture_state", direction: "asc" },
        { field_key: "timeline.activity_synopsis_text", direction: "asc" },
      ],
    });
  });

  it("describes filter chip labels from contract metadata", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    expect(
      filterChipLabel(contract, {
        fieldKey: "timeline.capture_state",
        op: "eq",
        arg: { value: "reviewed" },
      }),
    ).toBe("Capture State: reviewed");
  });

  it("initializes filter drafts from the contract and resolves input modes", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    expect(defaultFilterDraft(contract).fieldKey).toBe(
      "timeline.date_entered_sort_day",
    );
    expect(filterInputMode("timeline.date_entered_sort_day")).toBe("date");
    expect(filterInputMode("timeline.tags")).toBe("tagset");
  });
});
