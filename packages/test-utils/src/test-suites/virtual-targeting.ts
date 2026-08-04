import {
  dataTestIdSelector,
  gridShellTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import { describe, expect, it, vi } from "vitest";

import {
  isTestIdVisibleWithinGridViewport,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "../grid";
import {
  createBrowserPage,
  testTimelineViewSchemaId,
} from "./browser-fixtures";
import { installMarkerAnchorFixture } from "./marker-fixtures";
import { installGridTargetFixture } from "./targeting-fixtures";

export function registerVirtualTargetingSuite() {
  describe("@cartulary/test-utils virtualized grid targeting", () => {
    it("distinguishes mounted targets outside the clipped grid viewport", async () => {
      const outside = installMarkerAnchorFixture({
        markerRect: { height: 18, left: 130, top: 472, width: 48 },
        targetCellRect: { height: 60, left: 100, top: 440, width: 120 },
      });

      await expect(
        isTestIdVisibleWithinGridViewport(
          outside.page,
          testTimelineViewSchemaId,
          outside.targetTestId,
        ),
      ).resolves.toBe(false);

      document.body.innerHTML = "";
      vi.restoreAllMocks();
      const inside = installMarkerAnchorFixture({
        markerRect: { height: 18, left: 130, top: 72, width: 48 },
      });
      await expect(
        isTestIdVisibleWithinGridViewport(
          inside.page,
          testTimelineViewSchemaId,
          inside.targetTestId,
        ),
      ).resolves.toBe(true);
    });

    it("treats an unmounted virtual target as outside the grid viewport", async () => {
      const { page, targetTestId } = installGridTargetFixture({
        includeTarget: false,
      });

      await expect(
        isTestIdVisibleWithinGridViewport(
          page,
          testTimelineViewSchemaId,
          targetTestId,
        ),
      ).resolves.toBe(false);
    });

    it("treats a target unmounted between visibility and geometry reads as outside the viewport", async () => {
      let visibilityChecks = 0;
      let targetToUnmount: HTMLInputElement | null = null;
      const { page, target, targetTestId } = installGridTargetFixture({
        isTargetVisible: () => {
          visibilityChecks += 1;
          if (visibilityChecks === 1) {
            targetToUnmount?.remove();
            return true;
          }
          return false;
        },
      });
      targetToUnmount = target;

      await expect(
        isTestIdVisibleWithinGridViewport(
          page,
          testTimelineViewSchemaId,
          targetTestId,
        ),
      ).resolves.toBe(false);
      expect(visibilityChecks).toBe(2);
    });

    it("retries through the stable grid shell when a target unmounts before alignment", async () => {
      let visibilityChecks = 0;
      let targetToUnmount: HTMLInputElement | null = null;
      const { page, scrollIntoViewCalls, target, targetTestId } =
        installGridTargetFixture({
          isTargetVisible: () => {
            visibilityChecks += 1;
            if (visibilityChecks === 1) {
              targetToUnmount?.remove();
              return true;
            }
            return false;
          },
        });
      targetToUnmount = target;

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 0,
        }),
      ).rejects.toThrow(/Expected target-control to become visible/);
      expect(scrollIntoViewCalls).toEqual([]);
    });

    it("aligns an already-visible target before returning the existing scroll position", async () => {
      const { grid, page, scrollIntoViewCalls, targetTestId } =
        installGridTargetFixture({
          currentScroll: { left: 8, top: 120 },
          isTargetVisible: () => true,
        });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 50,
        }),
      ).resolves.toEqual({ left: 8, top: 120 });

      expect(grid.scrollTop).toBe(120);
      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("aligns an editor portaled outside the grid shell through the stable document", async () => {
      const { page, scrollIntoViewCalls, targetTestId } =
        installGridTargetFixture({
          currentScroll: { left: 8, top: 120 },
          isTargetVisible: () => true,
          targetOutsideGrid: true,
        });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 50,
        }),
      ).resolves.toEqual({ left: 8, top: 120 });

      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("scans grid offsets until a virtualized target is mounted", async () => {
      const { grid, page, scrollIntoViewCalls, targetTestId } =
        installGridTargetFixture({
          clientHeight: 200,
          clientWidth: 400,
          currentScroll: { left: 0, top: 0 },
          isTargetVisible: (candidateGrid) =>
            candidateGrid.scrollTop >= 400 && candidateGrid.scrollLeft >= 600,
          onTargetScrollIntoView: (candidateGrid) => {
            candidateGrid.scrollLeft = 720;
            candidateGrid.scrollTop = 520;
          },
          scrollHeight: 900,
          scrollWidth: 1_200,
        });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 1_000,
        }),
      ).resolves.toEqual({ left: 720, top: 520 });

      expect(grid.scrollLeft).toBe(720);
      expect(grid.scrollTop).toBe(520);
      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("targets a row cell by stable record id and cell key", async () => {
      const { grid, page, scrollIntoViewCalls, targetTestId } =
        installGridTargetFixture({
          clientHeight: 200,
          currentScroll: { left: 0, top: 0 },
          isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 400,
          scrollHeight: 900,
          targetTestId: rowCellTestId(
            "record-1",
            "timeline.activity_synopsis_text",
          ),
        });

      await expect(
        scrollGridCellIntoView({
          cellKey: "timeline.activity_synopsis_text",
          intervalMs: 0,
          page,
          recordId: "record-1",
          surface: testTimelineViewSchemaId,
          timeoutMs: 1_000,
        }),
      ).resolves.toEqual({ left: 0, top: 400 });

      expect(grid.scrollTop).toBe(400);
      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("finishes a scan through the bottom offset after the deadline has expired", async () => {
      const { grid, page, scrollIntoViewCalls, targetTestId } =
        installGridTargetFixture({
          clientHeight: 502,
          currentScroll: { left: 0, top: 0 },
          isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 2_074,
          scrollHeight: 2_576,
        });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 1,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 0,
        }),
      ).resolves.toEqual({ left: 0, top: 2_074 });

      expect(grid.scrollTop).toBe(2_074);
      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("recomputes scan offsets after an initially non-scrollable grid hydrates", async () => {
      const { grid, page, scrollIntoViewCalls, targetTestId } =
        installGridTargetFixture({
          clientHeight: 200,
          currentScroll: { left: 0, top: 0 },
          isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 400,
          onGridEvaluate: (candidateGrid, evaluateCount) => {
            if (evaluateCount >= 3) {
              Object.defineProperty(candidateGrid, "scrollHeight", {
                configurable: true,
                value: 900,
              });
            }
          },
          scrollHeight: 200,
        });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 1_000,
        }),
      ).resolves.toEqual({ left: 0, top: 400 });

      expect(grid.scrollTop).toBe(400);
      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("extends the scan when the grid scroll range grows during scanning", async () => {
      const { grid, page, scrollIntoViewCalls, targetTestId } =
        installGridTargetFixture({
          clientHeight: 200,
          currentScroll: { left: 0, top: 0 },
          isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 700,
          onGridEvaluate: (candidateGrid, evaluateCount) => {
            if (evaluateCount >= 3) {
              Object.defineProperty(candidateGrid, "scrollHeight", {
                configurable: true,
                value: 1_000,
              });
            }
          },
          scrollHeight: 400,
        });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 1_000,
        }),
      ).resolves.toEqual({ left: 0, top: 700 });

      expect(grid.scrollTop).toBe(700);
      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("waits for target hydration when the grid never needs a scroll range", async () => {
      let targetHydrated = false;
      const { grid, page, scrollIntoViewCalls, targetTestId } =
        installGridTargetFixture({
          clientHeight: 200,
          currentScroll: { left: 0, top: 0 },
          isTargetVisible: () => targetHydrated,
          onGridEvaluate: (_candidateGrid, evaluateCount) => {
            if (evaluateCount >= 3) {
              targetHydrated = true;
            }
          },
          scrollHeight: 200,
        });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 1_000,
        }),
      ).resolves.toEqual({ left: 0, top: 0 });

      expect(grid.scrollTop).toBe(0);
      expect(scrollIntoViewCalls).toEqual([targetTestId]);
    });

    it("scrolls the explicit grid scrollport when the outer shell is also scrollable", async () => {
      const { grid, page, shell, targetTestId } = installGridTargetFixture({
        clientHeight: 200,
        currentScroll: { left: 0, top: 0 },
        isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 400,
        outerScrollable: true,
        scrollHeight: 900,
      });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 1_000,
        }),
      ).resolves.toEqual({ left: 0, top: 400 });

      expect(grid.scrollTop).toBe(400);
      expect(shell.scrollTop).toBe(250);
    });

    it("fails explicitly when the grid shell has no owned scrollport", async () => {
      const gridTestId = gridShellTestId(testTimelineViewSchemaId);
      document.body.innerHTML = `<div data-testid="${gridTestId}"></div>`;
      const shell = document.querySelector(dataTestIdSelector(gridTestId));
      if (!(shell instanceof HTMLDivElement)) {
        throw new Error("Expected missing-scrollport fixture shell to exist");
      }

      const page = createBrowserPage({ [gridTestId]: shell });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId: "missing-target",
          timeoutMs: 50,
        }),
      ).rejects.toThrow(
        `Expected ${testTimelineViewSchemaId} grid shell to contain exactly one .cartulary-grid-scrollport scrollport, received 0`,
      );
    });

    it("throws diagnostics when the target never becomes visible", async () => {
      const { page } = installGridTargetFixture({
        includeTarget: false,
        mountedRowIds: ["record-a", "record-b"],
      });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId: "missing-target",
          timeoutMs: 50,
        }),
      ).rejects.toThrow(
        /missing-target.*cartulary\.view\.timeline\.v2.*scrollHeight=900.*mountedRowIds=record-a,record-b.*completedScanCycles=.*scrollRangeGrowths=.*observedMaxTop=700.*completedScanMaxTop=700.*observedMountedRowIds=record-a,record-b/,
      );
    });

    it("returns the final scroll snapshot after target alignment", async () => {
      const { page, targetTestId } = installGridTargetFixture({
        clientHeight: 200,
        currentScroll: { left: 3, top: 0 },
        isTargetVisible: (candidateGrid) => candidateGrid.scrollTop >= 200,
        onTargetScrollIntoView: (candidateGrid) => {
          candidateGrid.scrollLeft = 21;
          candidateGrid.scrollTop = 333;
        },
        scrollHeight: 900,
      });

      await expect(
        scrollGridTargetIntoView({
          intervalMs: 0,
          page,
          surface: testTimelineViewSchemaId,
          targetTestId,
          timeoutMs: 1_000,
        }),
      ).resolves.toEqual({ left: 21, top: 333 });
    });
  });
}
