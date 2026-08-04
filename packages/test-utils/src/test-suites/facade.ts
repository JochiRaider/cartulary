import { describe, expect, it } from "vitest";

import * as gridApi from "../grid";

export function registerFacadeSuite() {
  describe("@cartulary/test-utils selector choreography", () => {
    it("exposes exact public facade runtime shapes", () => {
      expect(Object.keys(gridApi).sort()).toEqual([
        "applyFilterChip",
        "assertActiveFilterChipVisible",
        "assertGridFocusContinuity",
        "assertGroupRowPresentationOnly",
        "assertMarkerAnchoredToGridTarget",
        "assertMountedGridRowCountAtMost",
        "changeGrouping",
        "collapseGridGroup",
        "expandGridGroup",
        "isTestIdVisibleWithinGridViewport",
        "pasteGridMatrix",
        "removeFilterChip",
        "scrollGridCellIntoView",
        "scrollGridTargetIntoView",
        "scrollGridToBottom",
        "scrollGridToOffset",
        "sortByHeader",
      ]);
    });
  });
}
