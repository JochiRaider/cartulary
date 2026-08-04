import { gridRowGutterTestId, gridShellTestId } from "@cartulary/ui-contracts";
import { describe, expect, it, vi } from "vitest";

import { assertMarkerAnchoredToGridTarget } from "../grid";
import { testTimelineViewSchemaId } from "./browser-fixtures";
import { installMarkerAnchorFixture } from "./marker-fixtures";

export function registerMarkerSuite() {
  describe("@cartulary/test-utils marker anchoring", () => {
    it("fails when visible marker geometry is detached from the target cell", async () => {
      vi.useFakeTimers();
      const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
        markerRect: { height: 18, left: 310, top: 54, width: 28 },
      });

      await expectTimedMarkerFailure(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "cell",
          markerTestId,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
        }),
        "to be geometrically inside target cell",
      );
    });

    it("fails when marker is visible in a different row", async () => {
      vi.useFakeTimers();
      const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
        markerRecordId: "record-2",
        markerRect: { height: 18, left: 110, top: 174, width: 28 },
        markerRowRect: { height: 60, left: 0, top: 160, width: 450 },
      });

      await expectTimedMarkerFailure(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "cell",
          markerTestId,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
        }),

        "Expected marker marker-record-1 to share row record_id record-1 with target row-record-1-timeline.activity_synopsis_text, received record-2",
      );
    });

    it("fails when marker is visible in a different field cell", async () => {
      vi.useFakeTimers();
      const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
        markerCellFieldKey: "timeline.raw_activity_text",
        markerCellRect: { height: 60, left: 250, top: 40, width: 120 },
        markerRect: { height: 18, left: 260, top: 54, width: 28 },
      });

      await expectTimedMarkerFailure(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "cell",
          markerTestId,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
        }),

        "Expected marker marker-record-1 to share cell field_key timeline.activity_synopsis_text with target row-record-1-timeline.activity_synopsis_text, received timeline.raw_activity_text",
      );
    });

    it("passes when same-cell marker geometry is inside the target field cell", async () => {
      const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
        markerRect: { height: 18, left: 130, top: 72, width: 48 },
      });

      await expect(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "cell",
          markerTestId,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
        }),
      ).resolves.toBeUndefined();
    });

    it("does not invoke mutating browser capabilities after assertion entry", async () => {
      const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
        markerRect: { height: 18, left: 130, top: 72, width: 48 },
      });
      const mutations: string[] = [];
      const observerPage = {
        getByTestId(value: string) {
          const locator = page.getByTestId(value);
          const record = (capability: string) => async () => {
            mutations.push(`${capability}:${value}`);
          };
          return {
            ...locator,
            blur: record("blur"),
            click: record("click"),
            dispatchEvent: record("dispatchEvent"),
            fill: record("fill"),
            press: record("press"),
            scrollIntoViewIfNeeded: record("scrollIntoViewIfNeeded"),
            selectOption: record("selectOption"),
          };
        },
      };

      await assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId,
        page: observerPage,
        surface: testTimelineViewSchemaId,
        targetTestId,
      });

      expect(mutations).toEqual([]);
    });

    it("reacquires marker geometry after an ordinary grid remount", async () => {
      vi.useFakeTimers();
      const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
        markerRect: { height: 18, left: 130, top: 72, width: 48 },
      });
      const gridTestId = gridShellTestId(testTimelineViewSchemaId);
      let gridEvaluateCount = 0;
      const remountingPage = {
        getByTestId(value: string) {
          const locator = page.getByTestId(value);
          if (value !== gridTestId || locator.evaluate === undefined) {
            return locator;
          }
          return {
            ...locator,
            evaluate: async (
              pageFunction: (element: Element, arg?: unknown) => unknown,
              arg?: unknown,
            ) => {
              gridEvaluateCount += 1;
              if (gridEvaluateCount === 1) {
                throw new Error("ordinary virtualized row remount");
              }
              return locator.evaluate?.(pageFunction, arg);
            },
          };
        },
      };

      const assertion = assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId,
        page: remountingPage,
        surface: testTimelineViewSchemaId,
        targetTestId,
      });
      await vi.advanceTimersByTimeAsync(50);

      await expect(assertion).resolves.toBeUndefined();
      expect(gridEvaluateCount).toBeGreaterThanOrEqual(3);
    });

    it("times out with bounded diagnostics when the target never becomes observable", async () => {
      vi.useFakeTimers();
      const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
        markerRect: { height: 18, left: 130, top: 72, width: 48 },
      });
      const hiddenTargetPage = {
        getByTestId(value: string) {
          const locator = page.getByTestId(value);
          return value === targetTestId
            ? { ...locator, isVisible: async () => false }
            : locator;
        },
      };

      await expectTimedMarkerFailure(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "cell",
          markerTestId,
          page: hiddenTargetPage,
          surface: testTimelineViewSchemaId,
          targetTestId,
        }),
        `Expected target ${targetTestId} to be visible in the ${testTimelineViewSchemaId} grid viewport`,
      );
    });

    it("honors the exact two-pixel geometry tolerance", async () => {
      const tolerated = installMarkerAnchorFixture({
        markerRect: { height: 64, left: 98, top: 38, width: 124 },
      });
      await expect(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "cell",
          markerTestId: tolerated.markerTestId,
          page: tolerated.page,
          surface: testTimelineViewSchemaId,
          targetTestId: tolerated.targetTestId,
        }),
      ).resolves.toBeUndefined();

      vi.restoreAllMocks();
      vi.useFakeTimers();
      const outsideTolerance = installMarkerAnchorFixture({
        markerRect: { height: 66, left: 97, top: 37, width: 126 },
      });
      await expectTimedMarkerFailure(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "cell",
          markerTestId: outsideTolerance.markerTestId,
          page: outsideTolerance.page,
          surface: testTimelineViewSchemaId,
          targetTestId: outsideTolerance.targetTestId,
        }),
        "to be geometrically inside target cell",
      );
    });

    it("passes row-gutter markers only when vertically aligned to the intended row", async () => {
      const anchored = installMarkerAnchorFixture({
        markerCellFieldKey: "__cartulary_row_gutter__",
        markerRect: { height: 18, left: 16, top: 54, width: 34 },
        targetCellFieldKey: "__cartulary_row_gutter__",
        targetCellRect: { height: 60, left: 0, top: 40, width: 90 },
        targetTestId: gridRowGutterTestId(testTimelineViewSchemaId, "record-1"),
      });

      await expect(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "row-gutter",
          markerTestId: anchored.markerTestId,
          page: anchored.page,
          surface: testTimelineViewSchemaId,
          targetTestId: anchored.targetTestId,
        }),
      ).resolves.toBeUndefined();

      vi.restoreAllMocks();
      vi.useFakeTimers();
      const detached = installMarkerAnchorFixture({
        markerCellFieldKey: "__cartulary_row_gutter__",
        markerRect: { height: 18, left: 16, top: 154, width: 34 },
        targetCellFieldKey: "__cartulary_row_gutter__",
        targetCellRect: { height: 60, left: 0, top: 40, width: 90 },
        targetTestId: gridRowGutterTestId(testTimelineViewSchemaId, "record-1"),
      });

      await expectTimedMarkerFailure(
        assertMarkerAnchoredToGridTarget({
          anchorKind: "row-gutter",
          markerTestId: detached.markerTestId,
          page: detached.page,
          surface: testTimelineViewSchemaId,
          targetTestId: detached.targetTestId,
        }),
        "to be vertically anchored to row record-1",
      );
    });
  });
}

async function expectTimedMarkerFailure(
  assertion: Promise<void>,
  expectedMessage: string,
) {
  const rejection = expect(assertion).rejects.toThrow(expectedMessage);
  await vi.advanceTimersByTimeAsync(3_000);
  await rejection;
}
