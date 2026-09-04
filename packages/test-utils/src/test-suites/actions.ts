import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridScrollportClassName,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  workbookFilterPopoverTriggerTestId,
  workbookQueryEntryTestId,
} from "@cartulary/ui-contracts";
import { describe, expect, it, vi } from "vitest";

import * as gridApi from "../grid";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  changeGrouping,
  pasteGridMatrix,
  removeFilterChip,
  sortByHeader,
} from "../grid";
import { testTimelineViewSchemaId } from "./browser-fixtures";
import { installGridTargetFixture } from "./targeting-fixtures";

export function registerActionSuite() {
  describe("@cartulary/test-utils selector choreography", () => {
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
        workbookQueryEntryTestId(surface, "filter", "timeline.capture_state"),
        workbookQueryEntryTestId(surface, "filter", "timeline.capture_state"),
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

    it("formats paste matrix text", async () => {
      class TestDataTransfer {
        readonly #values = new Map<string, string>();

        getData(type: string) {
          return this.#values.get(type) ?? "";
        }

        setData(type: string, value: string) {
          this.#values.set(type, value);
        }
      }
      class TestClipboardEvent extends Event {
        readonly clipboardData: TestDataTransfer;

        constructor(
          type: string,
          options: EventInit & { clipboardData: TestDataTransfer },
        ) {
          super(type, options);
          this.clipboardData = options.clipboardData;
        }
      }
      vi.stubGlobal("DataTransfer", TestDataTransfer);
      vi.stubGlobal("ClipboardEvent", TestClipboardEvent);

      const fieldKey = "timeline.activity_synopsis_text";
      const recordId = "record-1";
      const targetTestId = rowCellTestId(recordId, fieldKey);
      const { page, scrollIntoViewCalls, target } = installGridTargetFixture({
        targetTestId,
      });
      let clipboardText = "";
      target?.addEventListener("paste", (event) => {
        clipboardText =
          (event as ClipboardEvent).clipboardData?.getData("text/plain") ?? "";
      });

      await pasteGridMatrix({
        fieldKey,
        matrix: [
          ["a", "b"],
          ["c", "d"],
        ],
        page,
        recordId,
        surface: testTimelineViewSchemaId,
      });

      expect(clipboardText).toBe("a\tb\nc\td");
      expect(document.activeElement).toBe(target);
      expect(scrollIntoViewCalls).toEqual([targetTestId]);
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
