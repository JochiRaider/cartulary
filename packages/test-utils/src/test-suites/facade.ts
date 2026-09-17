import { describe, expect, it } from "vitest";
import type { BrowserPageLike } from "../browser";
import * as gridApi from "../grid";

export function registerFacadeSuite() {
  describe("@cartulary/test-utils selector choreography", () => {
    it("captures correction geometry without field values or layout mutations", async () => {
      const container = document.createElement("div");
      const input = document.createElement("input");
      const button = document.createElement("button");
      input.value = "private unfinished authoring";
      button.textContent = "private record label";
      container.append(input, button);
      document.body.append(container);
      input.scrollLeft = 9;
      button.focus();
      const page: BrowserPageLike = {
        getByTestId: () => ({
          click: async () => {},
          fill: async () => {},
          evaluate: async (fn, arg) => fn(input, arg),
        }),
      };
      const geometry = await gridApi.readGridTargetGeometry(
        page,
        "synthetic-editor",
      );
      expect(geometry).toMatchObject({
        ancestors: expect.arrayContaining([
          expect.objectContaining({
            tag: "INPUT",
            scrollLeft: 9,
            focused: false,
          }),
        ]),
        focusedAncestors: expect.arrayContaining([
          expect.objectContaining({ tag: "BUTTON", focused: true }),
        ]),
      });
      expect(JSON.stringify(geometry)).not.toContain("private");
      expect(document.activeElement).toBe(button);
      expect(input.scrollLeft).toBe(9);
      expect(input.value).toBe("private unfinished authoring");
    });
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
        "readGridTargetGeometry",
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
