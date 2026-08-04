import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridScrollportClassName,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  workbookFilterPopoverTriggerTestId,
} from "@cartulary/ui-contracts";
import { describe, expect, it } from "vitest";

import * as accessibilityApi from "../accessibility";
import * as gridApi from "../grid";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  assertGroupRowPresentationOnly,
  changeGrouping,
  collapseGridGroup,
  expandGridGroup,
  removeFilterChip,
  sortByHeader,
} from "../grid";
import { pasteMatrixText } from "../matrix";
import * as visualApi from "../visual";
import { installGridTargetFixture, testTimelineViewSchemaId } from "./fixtures";

export function registerSelectorGroupingSuite() {
  describe("@cartulary/test-utils selector choreography", () => {
    it("exposes exact public facade runtime shapes", () => {
      expect(Object.keys(accessibilityApi).sort()).toEqual([
        "assertGridFocusContinuity",
        "assertMarkerAnchoredToGridTarget",
      ]);
      expect(Object.keys(gridApi).sort()).toEqual([
        "applyFilterChip",
        "assertActiveFilterChipVisible",
        "assertGridFocusContinuity",
        "assertGroupRowPresentationOnly",
        "assertMarkerAnchoredToGridTarget",
        "assertMountedGridRowCountAtMost",
        "changeGrouping",
        "collapseGridGroup",
        "delay",
        "expandGridGroup",
        "gridAnchorCommandScenarios",
        "isLocatorVisible",
        "isTestIdVisibleWithinGridViewport",
        "pasteGridMatrix",
        "removeFilterChip",
        "requireEvaluate",
        "requireSelectOption",
        "scrollGridCellIntoView",
        "scrollGridTargetIntoView",
        "scrollGridToBottom",
        "scrollGridToOffset",
        "sortByHeader",
        "supportsVisibilityCheck",
      ]);
      expect(Object.keys(visualApi).sort()).toEqual([
        "assertMarkerAnchoredToGridTarget",
        "scrollGridCellIntoView",
        "scrollGridTargetIntoView",
        "scrollGridToBottom",
        "scrollGridToOffset",
      ]);
    });

    it("returns row-cell anchors from the shared selector builder", async () => {
      const recordId = "record-1";
      const fieldKey = "timeline.activity_synopsis_text";
      const targetTestId = rowCellTestId(recordId, fieldKey);
      const { page, scrollIntoViewCalls } = installGridTargetFixture({
        targetTestId,
      });

      await gridApi.scrollGridCellIntoView({
        cellKey: fieldKey,
        page,
        recordId,
        surface: testTimelineViewSchemaId,
      });

      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("targets sort, filter, and grouping controls through shared builders", async () => {
      const observed: string[] = [];
      const selected: Record<string, string | readonly string[]> = {};
      const filled: Record<string, string> = {};
      const gridShell = document.createElement("div");
      const gridScrollport = document.createElement("div");
      gridScrollport.className = gridScrollportClassName();
      gridShell.append(gridScrollport);
      document.body.append(gridShell);
      const gridTestId = gridShellTestId(testTimelineViewSchemaId);
      const elements = new Map<string, HTMLElement>([[gridTestId, gridShell]]);
      const page = {
        getByTestId(value: string) {
          observed.push(value);
          let element = elements.get(value);
          if (element === undefined) {
            element = document.createElement("button");
            element.dataset.testid = value;
            Object.defineProperty(element, "scrollIntoView", {
              configurable: true,
              value: () => undefined,
            });
            gridScrollport.append(element);
            elements.set(value, element);
          }
          return {
            click: async () => undefined,
            evaluate: async (
              pageFunction: (element: Element, arg?: unknown) => unknown,
              arg?: unknown,
            ) => pageFunction(element, arg),
            fill: async (nextValue: string) => {
              filled[value] = nextValue;
            },
            isVisible: async () => true,
            selectOption: async (nextValue: string | readonly string[]) => {
              selected[value] = nextValue;
            },
          };
        },
      };
      const surface = testTimelineViewSchemaId;

      await sortByHeader(page, surface, "timeline.activity_synopsis_text");
      await applyFilterChip(page, surface, "timeline.capture_state", "rough");
      await assertActiveFilterChipVisible(
        page,
        surface,
        "timeline.capture_state",
      );
      await removeFilterChip(page, surface, "timeline.capture_state");
      await changeGrouping(page, surface, "timeline.capture_state");

      expect(observed).toEqual([
        gridSortHeaderTestId(surface, "timeline.activity_synopsis_text"),
        gridShellTestId(surface),
        gridShellTestId(surface),
        gridSortHeaderTestId(surface, "timeline.activity_synopsis_text"),
        gridFilterFieldTestId(surface),
        workbookFilterPopoverTriggerTestId(surface),
        gridFilterValueTestId(surface),
        gridFilterApplyTestId(surface),
        gridFilterChipTestId(surface, "timeline.capture_state"),
        gridFilterChipTestId(surface, "timeline.capture_state"),
        gridGroupingSelectTestId(surface),
      ]);
      expect(selected[gridFilterFieldTestId(surface)]).toBe(
        "timeline.capture_state",
      );
      expect(selected[gridFilterValueTestId(surface)]).toBe("rough");
      expect(selected[gridGroupingSelectTestId(surface)]).toBe(
        "timeline.capture_state",
      );
      expect(filled).toEqual({});
    });

    it("fails required grouping selection closed and propagates rejection", async () => {
      const missing = createSelectionActionPage({
        groupingSelect: "missing",
      });
      await expect(
        changeGrouping(
          missing.page,
          testTimelineViewSchemaId,
          "timeline.capture_state",
        ),
      ).rejects.toThrow(
        `changeGrouping(${testTimelineViewSchemaId}) requires locator.selectOption() support`,
      );
      expect(missing.calls).toEqual([
        `resolve:${gridGroupingSelectTestId(testTimelineViewSchemaId)}`,
      ]);

      const rejected = createSelectionActionPage({
        groupingSelect: "reject",
      });
      await expect(
        changeGrouping(
          rejected.page,
          testTimelineViewSchemaId,
          "timeline.capture_state",
        ),
      ).rejects.toThrow("select rejected");
      expect(rejected.calls).toEqual([
        `resolve:${gridGroupingSelectTestId(testTimelineViewSchemaId)}`,
        `select:${gridGroupingSelectTestId(testTimelineViewSchemaId)}:timeline.capture_state`,
      ]);
    });

    it("checks required filter field selection before opening and stops after rejection", async () => {
      const missing = createSelectionActionPage({ fieldSelect: "missing" });
      await expect(
        applyFilterChip(
          missing.page,
          testTimelineViewSchemaId,
          "timeline.capture_state",
          "rough",
        ),
      ).rejects.toThrow(
        `applyFilterChip(${testTimelineViewSchemaId}) requires filter field locator.selectOption() support`,
      );
      expect(missing.calls).toEqual([
        `resolve:${gridFilterFieldTestId(testTimelineViewSchemaId)}`,
      ]);

      const rejected = createSelectionActionPage({ fieldSelect: "reject" });
      await expect(
        applyFilterChip(
          rejected.page,
          testTimelineViewSchemaId,
          "timeline.capture_state",
          "rough",
        ),
      ).rejects.toThrow("select rejected");
      expect(rejected.calls).toEqual([
        `resolve:${gridFilterFieldTestId(testTimelineViewSchemaId)}`,
        `resolve:${workbookFilterPopoverTriggerTestId(testTimelineViewSchemaId)}`,
        `click:${workbookFilterPopoverTriggerTestId(testTimelineViewSchemaId)}`,
        `select:${gridFilterFieldTestId(testTimelineViewSchemaId)}:timeline.capture_state`,
      ]);
    });

    it("uses one successful filter value path and clicks apply exactly once", async () => {
      const fixture = createSelectionActionPage({ valueSelect: "resolve" });
      await applyFilterChip(
        fixture.page,
        testTimelineViewSchemaId,
        "timeline.capture_state",
        "rough",
      );

      expect(
        fixture.calls.filter((call) =>
          call.startsWith(
            `select:${gridFilterValueTestId(testTimelineViewSchemaId)}:`,
          ),
        ),
      ).toEqual([
        `select:${gridFilterValueTestId(testTimelineViewSchemaId)}:rough`,
      ]);
      expect(
        fixture.calls.filter((call) =>
          call.startsWith(
            `fill:${gridFilterValueTestId(testTimelineViewSchemaId)}:`,
          ),
        ),
      ).toEqual([]);
      expect(
        fixture.calls.filter(
          (call) =>
            call === `click:${gridFilterApplyTestId(testTimelineViewSchemaId)}`,
        ),
      ).toHaveLength(1);
    });

    it("falls back from absent or rejected filter value selection and stops when fill rejects", async () => {
      for (const valueSelect of ["missing", "reject"] as const) {
        const fixture = createSelectionActionPage({ valueSelect });
        await applyFilterChip(
          fixture.page,
          testTimelineViewSchemaId,
          "timeline.activity_synopsis_text",
          "needle",
        );
        expect(fixture.calls).toContain(
          `fill:${gridFilterValueTestId(testTimelineViewSchemaId)}:needle`,
        );
        expect(
          fixture.calls.filter(
            (call) =>
              call ===
              `click:${gridFilterApplyTestId(testTimelineViewSchemaId)}`,
          ),
        ).toHaveLength(1);
      }

      const dualFailure = createSelectionActionPage({
        fill: "reject",
        valueSelect: "reject",
      });
      await expect(
        applyFilterChip(
          dualFailure.page,
          testTimelineViewSchemaId,
          "timeline.activity_synopsis_text",
          "needle",
        ),
      ).rejects.toThrow("fill rejected");
      expect(dualFailure.calls).not.toContain(
        `resolve:${gridFilterApplyTestId(testTimelineViewSchemaId)}`,
      );
      expect(dualFailure.calls).not.toContain(
        `click:${gridFilterApplyTestId(testTimelineViewSchemaId)}`,
      );
    });

    it("formats paste matrix text", () => {
      expect(
        pasteMatrixText([
          ["a", "b"],
          ["c", "d"],
        ]),
      ).toBe("a\tb\nc\td");
    });

    it("toggles group outline expansion by aria state", async () => {
      let ariaExpanded = "true";
      let clickCount = 0;
      const element = {
        getAttribute(name: string) {
          return name === "aria-expanded" ? ariaExpanded : null;
        },
      } as Element;
      const page = {
        getByTestId(value: string) {
          expect(value).toBe("group-row");
          return {
            click: async () => {
              clickCount += 1;
              ariaExpanded = ariaExpanded === "true" ? "false" : "true";
            },
            evaluate: async (
              pageFunction: (element: Element, arg?: unknown) => unknown,
              arg?: unknown,
            ) => pageFunction(element, arg),
            fill: async () => undefined,
          };
        },
      };

      await collapseGridGroup({
        groupTestId: "group-row",
        page,
        surface: testTimelineViewSchemaId,
      });
      expect(ariaExpanded).toBe("false");
      expect(clickCount).toBe(1);

      await collapseGridGroup({
        groupTestId: "group-row",
        page,
        surface: testTimelineViewSchemaId,
      });
      expect(clickCount).toBe(1);

      await expandGridGroup({
        groupTestId: "group-row",
        page,
        surface: testTimelineViewSchemaId,
      });
      expect(ariaExpanded).toBe("true");
      expect(clickCount).toBe(2);
    });

    it("asserts group rows remain presentation-only", async () => {
      const page = {
        getByTestId(value: string) {
          const element = Array.from(
            document.querySelectorAll<HTMLElement>("[data-testid]"),
          ).find((candidate) => candidate.dataset.testid === value);
          if (element === undefined) {
            throw new Error(`Missing test id ${value}`);
          }
          return {
            click: async () => undefined,
            evaluate: async (
              pageFunction: (element: Element, arg?: unknown) => unknown,
              arg?: unknown,
            ) => pageFunction(element, arg),
            fill: async () => undefined,
          };
        },
      };

      document.body.innerHTML = `
        <div role="row" aria-level="1" aria-expanded="true">
          <div role="gridcell">
            <button aria-expanded="true" data-testid="group-row" type="button">reviewed</button>
          </div>
        </div>
      `;
      await expect(
        assertGroupRowPresentationOnly({
          groupTestId: "group-row",
          page,
          surface: testTimelineViewSchemaId,
        }),
      ).resolves.toBeUndefined();

      document.body.innerHTML = `
        <div role="row" aria-level="1" aria-expanded="true" data-grid-record-id="record-1">
          <div role="gridcell">
            <button aria-expanded="true" data-testid="group-row" type="button">reviewed</button>
          </div>
        </div>
      `;
      await expect(
        assertGroupRowPresentationOnly({
          groupTestId: "group-row",
          page,
          surface: testTimelineViewSchemaId,
        }),
      ).rejects.toThrow(/omit data-grid-record-id/);
    });
  });
}

function createSelectionActionPage(
  options: {
    fieldSelect?: "missing" | "reject" | "resolve";
    fill?: "reject" | "resolve";
    groupingSelect?: "missing" | "reject" | "resolve";
    valueSelect?: "missing" | "reject" | "resolve";
  } = {},
) {
  const calls: string[] = [];
  const fieldTestId = gridFilterFieldTestId(testTimelineViewSchemaId);
  const groupingTestId = gridGroupingSelectTestId(testTimelineViewSchemaId);
  const valueTestId = gridFilterValueTestId(testTimelineViewSchemaId);
  return {
    calls,
    page: {
      getByTestId(testId: string) {
        calls.push(`resolve:${testId}`);
        const selectBehavior =
          testId === fieldTestId
            ? (options.fieldSelect ?? "resolve")
            : testId === groupingTestId
              ? (options.groupingSelect ?? "resolve")
              : testId === valueTestId
                ? (options.valueSelect ?? "resolve")
                : "missing";
        return {
          click: async () => {
            calls.push(`click:${testId}`);
          },
          fill: async (value: string) => {
            calls.push(`fill:${testId}:${value}`);
            if (options.fill === "reject") {
              throw new Error("fill rejected");
            }
          },
          ...(selectBehavior === "missing"
            ? {}
            : {
                selectOption: async (value: string | readonly string[]) => {
                  calls.push(`select:${testId}:${String(value)}`);
                  if (selectBehavior === "reject") {
                    throw new Error("select rejected");
                  }
                },
              }),
        };
      },
    },
  };
}
