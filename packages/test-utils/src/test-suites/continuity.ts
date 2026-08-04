import {
  gridScrollportClassName,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import { describe, expect, it, vi } from "vitest";

import { assertGridFocusContinuity } from "../grid";
import { testTimelineViewSchemaId } from "./browser-fixtures";
import { installGridContinuityFixture } from "./continuity-fixtures";

export function registerContinuitySuite() {
  describe("@cartulary/test-utils grid continuity", () => {
    it("observes focus and viewport continuity without focusing or scrolling", async () => {
      const { focusTestId, page } = installGridContinuityFixture({
        currentScroll: { left: 10, top: 180 },
        focusRect: { height: 40, left: 85, top: 170, width: 80 },
        gridRect: { height: 300, left: 40, top: 100, width: 400 },
      });
      const focusSpy = vi.spyOn(HTMLElement.prototype, "focus");
      const scrollIntoViewSpy = vi
        .spyOn(HTMLElement.prototype, "scrollIntoView")
        .mockImplementation(() => undefined);

      await expect(
        assertGridFocusContinuity({
          focusTestId,
          intervalMs: 0,
          page,
          preservedScroll: { left: 10, top: 180 },
          surface: testTimelineViewSchemaId,
          timeoutMs: 10,
        }),
      ).resolves.toBeUndefined();

      expect(focusSpy).not.toHaveBeenCalled();
      expect(scrollIntoViewSpy).not.toHaveBeenCalled();
    });

    it("allows visibility-preserving continuity when scroll changes by default", async () => {
      const { focusTestId, page } = installGridContinuityFixture({
        currentScroll: { left: 10, top: 180 },
        focusRect: { height: 40, left: 85, top: 170, width: 80 },
        gridRect: { height: 300, left: 40, top: 100, width: 400 },
      });

      await expect(
        assertGridFocusContinuity({
          focusTestId,
          intervalMs: 0,
          page,
          preservedScroll: { left: 18, top: 240 },
          surface: testTimelineViewSchemaId,
          timeoutMs: 10,
        }),
      ).resolves.toBeUndefined();
    });

    it("fails when exact vertical scroll is required and scrollTop changed under fake timers", async () => {
      const { focusTestId, page } = installGridContinuityFixture({
        currentScroll: { left: 18, top: 180 },
        focusRect: { height: 40, left: 85, top: 170, width: 80 },
        gridRect: { height: 300, left: 40, top: 100, width: 400 },
      });
      vi.useFakeTimers();

      await expect(
        assertGridFocusContinuity({
          focusTestId,
          intervalMs: 0,
          page,
          preservedScroll: { left: 18, top: 240 },
          requireExactVerticalScroll: true,
          surface: testTimelineViewSchemaId,
          timeoutMs: 10,
        }),
      ).rejects.toThrow(
        `Expected ${testTimelineViewSchemaId} vertical scroll 240, received 180`,
      );
    });

    it("retries until the preserved vertical scroll converges", async () => {
      let gridEvaluateCount = 0;
      const { focusTestId, page } = installGridContinuityFixture(
        {
          currentScroll: { left: 18, top: 180 },
          focusRect: { height: 40, left: 85, top: 170, width: 80 },
          gridRect: { height: 300, left: 40, top: 100, width: 400 },
        },
        {
          onEvaluate(testId, element) {
            if (testId !== gridShellTestId(testTimelineViewSchemaId)) {
              return;
            }
            gridEvaluateCount += 1;
            const scrollport = element.querySelector(
              `.${gridScrollportClassName()}`,
            );
            if (
              gridEvaluateCount >= 5 &&
              scrollport instanceof HTMLDivElement
            ) {
              scrollport.scrollTop = 240;
            }
          },
        },
      );

      await expect(
        assertGridFocusContinuity({
          focusTestId,
          intervalMs: 0,
          page,
          preservedScroll: { left: 18, top: 240 },
          requireExactVerticalScroll: true,
          surface: testTimelineViewSchemaId,
          timeoutMs: 250,
        }),
      ).resolves.toBeUndefined();

      expect(gridEvaluateCount).toBeGreaterThan(2);
    });

    it("fails when exact horizontal scroll is required and scrollLeft changed", async () => {
      const { focusTestId, page } = installGridContinuityFixture({
        currentScroll: { left: 10, top: 240 },
        focusRect: { height: 40, left: 85, top: 170, width: 80 },
        gridRect: { height: 300, left: 40, top: 100, width: 400 },
      });

      await expect(
        assertGridFocusContinuity({
          focusTestId,
          intervalMs: 0,
          page,
          preservedScroll: { left: 18, top: 240 },
          requireExactHorizontalScroll: true,
          surface: testTimelineViewSchemaId,
          timeoutMs: 10,
        }),
      ).rejects.toThrow(
        `Expected ${testTimelineViewSchemaId} horizontal scroll 18, received 10`,
      );
    });
  });
}
