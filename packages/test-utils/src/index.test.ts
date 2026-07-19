// @vitest-environment jsdom

import {
  dataTestIdSelector,
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridRowGutterTestId,
  gridScrollportClassName,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  workbookFilterPopoverTriggerTestId,
} from "@cartulary/ui-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  assertGridFocusContinuity,
  assertGroupRowPresentationOnly,
  assertMarkerAnchoredToGridTarget,
  changeGrouping,
  collapseGridGroup,
  expandGridGroup,
  isTestIdVisibleWithinGridViewport,
  removeFilterChip,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
  sortByHeader,
} from "./grid";
import { assertAnchorTestId } from "./grid-editing";
import { pasteMatrixText } from "./matrix";

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  document.body.innerHTML = "";
});

const testTimelineViewSchemaId = "cartulary.view.timeline.v2";

describe("@cartulary/test-utils selector choreography", () => {
  it("returns row-cell anchors from the shared selector builder", () => {
    expect(
      assertAnchorTestId("record-1", "timeline.activity_synopsis_text"),
    ).toBe(rowCellTestId("record-1", "timeline.activity_synopsis_text"));
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
      workbookFilterPopoverTriggerTestId(surface),
      gridFilterFieldTestId(surface),
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
      <div role="row" data-grid-row-kind="group">
        <div role="gridcell" data-grid-row-kind="group">
          <button data-testid="group-row" type="button">reviewed</button>
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
      <div role="row" data-grid-row-kind="group" data-grid-record-id="record-1">
        <div role="gridcell" data-grid-row-kind="group">
          <button data-testid="group-row" type="button">reviewed</button>
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

describe("@cartulary/test-utils grid continuity", () => {
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
          if (gridEvaluateCount >= 5 && scrollport instanceof HTMLDivElement) {
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

describe("@cartulary/test-utils marker anchoring", () => {
  it("fails when visible marker geometry is detached from the target cell", async () => {
    const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
      markerRect: { height: 18, left: 310, top: 54, width: 28 },
    });

    await expect(
      assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId,
        page,
        surface: testTimelineViewSchemaId,
        targetTestId,
      }),
    ).rejects.toThrow("to be geometrically inside target cell");
  });

  it("fails when marker is visible in a different row", async () => {
    const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
      markerRecordId: "record-2",
      markerRect: { height: 18, left: 110, top: 174, width: 28 },
      markerRowRect: { height: 60, left: 0, top: 160, width: 450 },
    });

    await expect(
      assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId,
        page,
        surface: testTimelineViewSchemaId,
        targetTestId,
      }),
    ).rejects.toThrow(
      "Expected marker marker-record-1 to share row record_id record-1 with target row-record-1-timeline.activity_synopsis_text, received record-2",
    );
  });

  it("fails when marker is visible in a different field cell", async () => {
    const { markerTestId, page, targetTestId } = installMarkerAnchorFixture({
      markerCellFieldKey: "timeline.raw_activity_text",
      markerCellRect: { height: 60, left: 250, top: 40, width: 120 },
      markerRect: { height: 18, left: 260, top: 54, width: 28 },
    });

    await expect(
      assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId,
        page,
        surface: testTimelineViewSchemaId,
        targetTestId,
      }),
    ).rejects.toThrow(
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
    const detached = installMarkerAnchorFixture({
      markerCellFieldKey: "__cartulary_row_gutter__",
      markerRect: { height: 18, left: 16, top: 154, width: 34 },
      targetCellFieldKey: "__cartulary_row_gutter__",
      targetCellRect: { height: 60, left: 0, top: 40, width: 90 },
      targetTestId: gridRowGutterTestId(testTimelineViewSchemaId, "record-1"),
    });

    await expect(
      assertMarkerAnchoredToGridTarget({
        anchorKind: "row-gutter",
        markerTestId: detached.markerTestId,
        page: detached.page,
        surface: testTimelineViewSchemaId,
        targetTestId: detached.targetTestId,
      }),
    ).rejects.toThrow("to be vertically anchored to row record-1");
  });
});

function installGridContinuityFixture(
  options: {
    currentScroll: { left: number; top: number };
    focusRect: { height: number; left: number; top: number; width: number };
    gridRect: { height: number; left: number; top: number; width: number };
  },
  pageOptions: {
    onEvaluate?: (testId: string, element: Element) => void;
  } = {},
) {
  const focusTestId = rowInspectButtonTestId("record-1");
  const gridTestId = gridShellTestId(testTimelineViewSchemaId);
  document.body.innerHTML = `
    <div data-testid="${gridTestId}">
      <div class="${gridScrollportClassName()}">
        <button data-testid="${focusTestId}">Inspect</button>
      </div>
    </div>
  `;

  const grid = document.querySelector(dataTestIdSelector(gridTestId));
  const scrollport = grid?.querySelector(`.${gridScrollportClassName()}`);
  const focusTarget = document.querySelector(dataTestIdSelector(focusTestId));
  if (
    !(grid instanceof HTMLDivElement) ||
    !(scrollport instanceof HTMLDivElement) ||
    !(focusTarget instanceof HTMLButtonElement)
  ) {
    throw new Error("Expected grid continuity fixture elements to exist");
  }

  scrollport.scrollTop = options.currentScroll.top;
  scrollport.scrollLeft = options.currentScroll.left;
  focusTarget.focus();

  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockImplementation(
    function mockRect(this: HTMLElement) {
      const testId = this.getAttribute("data-testid");
      if (
        testId === null &&
        this.classList.contains(gridScrollportClassName())
      ) {
        return rectFromBox(options.gridRect);
      }
      if (testId === focusTestId) {
        return rectFromBox(options.focusRect);
      }
      return rectFromBox({ height: 0, left: 0, top: 0, width: 0 });
    },
  );

  return {
    focusTestId,
    page: createBrowserPage(
      {
        [focusTestId]: focusTarget,
        [gridTestId]: grid,
      },
      pageOptions,
    ),
  };
}

function installGridTargetFixture(
  options: {
    clientHeight?: number;
    clientWidth?: number;
    currentScroll?: { left: number; top: number };
    includeTarget?: boolean;
    isTargetVisible?: (grid: HTMLDivElement) => boolean;
    mountedRowIds?: string[];
    onGridEvaluate?: (grid: HTMLDivElement, evaluateCount: number) => void;
    onTargetScrollIntoView?: (grid: HTMLDivElement) => void;
    outerScrollable?: boolean;
    scrollHeight?: number;
    scrollWidth?: number;
    targetOutsideGrid?: boolean;
    targetTestId?: string;
  } = {},
) {
  const gridTestId = gridShellTestId(testTimelineViewSchemaId);
  const targetTestId = options.targetTestId ?? "target-control";
  const mountedRows = (options.mountedRowIds ?? ["record-1"])
    .map(
      (recordId) => `<div role="row" data-grid-record-id="${recordId}"></div>`,
    )
    .join("");
  document.body.innerHTML = `
    <div data-testid="${gridTestId}">
      <div class="${gridScrollportClassName()}">
        ${mountedRows}
        ${
          options.includeTarget === false || options.targetOutsideGrid === true
            ? ""
            : `<input data-testid="${targetTestId}" />`
        }
      </div>
    </div>
    ${
      options.includeTarget !== false && options.targetOutsideGrid === true
        ? `<input data-testid="${targetTestId}" />`
        : ""
    }
  `;

  const shell = document.querySelector(dataTestIdSelector(gridTestId));
  const grid = shell?.querySelector(`.${gridScrollportClassName()}`);
  const target = document.querySelector(dataTestIdSelector(targetTestId));
  if (!(shell instanceof HTMLDivElement) || !(grid instanceof HTMLDivElement)) {
    throw new Error("Expected grid targeting fixture grid to exist");
  }
  if (
    options.includeTarget !== false &&
    !(target instanceof HTMLInputElement)
  ) {
    throw new Error("Expected grid targeting fixture target to exist");
  }

  Object.defineProperties(grid, {
    clientHeight: {
      configurable: true,
      value: options.clientHeight ?? 200,
    },
    clientWidth: {
      configurable: true,
      value: options.clientWidth ?? 400,
    },
    scrollHeight: {
      configurable: true,
      value: options.scrollHeight ?? 900,
    },
    scrollWidth: {
      configurable: true,
      value: options.scrollWidth ?? 400,
    },
  });
  if (options.outerScrollable) {
    Object.defineProperties(shell, {
      clientHeight: {
        configurable: true,
        value: 100,
      },
      clientWidth: {
        configurable: true,
        value: 300,
      },
      scrollHeight: {
        configurable: true,
        value: 3_000,
      },
      scrollWidth: {
        configurable: true,
        value: 300,
      },
    });
    shell.scrollTop = 250;
  }
  grid.scrollLeft = options.currentScroll?.left ?? 0;
  grid.scrollTop = options.currentScroll?.top ?? 0;

  const scrollIntoViewCalls: string[] = [];
  if (target instanceof HTMLInputElement) {
    Object.defineProperty(target, "scrollIntoView", {
      configurable: true,
      value: () => {
        scrollIntoViewCalls.push(targetTestId);
        options.onTargetScrollIntoView?.(grid);
      },
    });
  }
  let gridEvaluateCount = 0;
  return {
    grid,
    page: createBrowserPage(
      () => {
        const elements: Record<string, Element | undefined> = {
          [gridTestId]: shell,
        };
        if (target instanceof HTMLInputElement) {
          elements[targetTestId] = target;
        }
        return elements;
      },
      {
        isVisible(testId, element) {
          if (testId === targetTestId && element === target) {
            return options.isTargetVisible?.(grid) ?? true;
          }
          return element.isConnected;
        },
        onEvaluate(testId, element) {
          if (testId === gridTestId && element === shell) {
            gridEvaluateCount += 1;
            options.onGridEvaluate?.(grid, gridEvaluateCount);
          }
        },
        onScrollIntoViewIfNeeded(testId) {
          scrollIntoViewCalls.push(testId);
          if (testId === targetTestId) {
            options.onTargetScrollIntoView?.(grid);
          }
        },
      },
    ),
    scrollIntoViewCalls,
    shell,
    target: target instanceof HTMLInputElement ? target : null,
    targetTestId,
  };
}

function installMarkerAnchorFixture(options: {
  markerCellFieldKey?: string;
  markerCellRect?: { height: number; left: number; top: number; width: number };
  markerRecordId?: string;
  markerRect: { height: number; left: number; top: number; width: number };
  markerRowRect?: { height: number; left: number; top: number; width: number };
  targetCellFieldKey?: string;
  targetCellRect?: { height: number; left: number; top: number; width: number };
  targetRecordId?: string;
  targetRowRect?: { height: number; left: number; top: number; width: number };
  targetTestId?: string;
}) {
  const gridTestId = gridShellTestId(testTimelineViewSchemaId);
  const markerTestId = "marker-record-1";
  const targetTestId =
    options.targetTestId ??
    rowCellTestId("record-1", "timeline.activity_synopsis_text");
  const targetRecordId = options.targetRecordId ?? "record-1";
  const markerRecordId = options.markerRecordId ?? targetRecordId;
  const targetCellFieldKey =
    options.targetCellFieldKey ?? "timeline.activity_synopsis_text";
  const markerCellFieldKey = options.markerCellFieldKey ?? targetCellFieldKey;
  const markerSameRow = markerRecordId === targetRecordId;
  const markerSameCell =
    markerSameRow && markerCellFieldKey === targetCellFieldKey;
  const targetRowRect = options.targetRowRect ?? {
    height: 60,
    left: 0,
    top: 40,
    width: 450,
  };
  const markerRowRect =
    options.markerRowRect ??
    (markerSameRow
      ? targetRowRect
      : { height: 60, left: 0, top: 160, width: 450 });
  const targetCellRect = options.targetCellRect ?? {
    height: 60,
    left: 100,
    top: 40,
    width: 120,
  };
  const markerCellRect =
    options.markerCellRect ??
    (markerSameCell
      ? targetCellRect
      : {
          height: markerRowRect.height,
          left: 250,
          top: markerRowRect.top,
          width: 120,
        });
  const targetInputRect = {
    height: 24,
    left: targetCellRect.left + 10,
    top: targetCellRect.top + 10,
    width: 80,
  };
  const rects = new Map<
    string,
    { height: number; left: number; top: number; width: number }
  >([
    ["scrollport", { height: 400, left: 0, top: 0, width: 500 }],
    ["target-row", targetRowRect],
    ["target-cell", targetCellRect],
    ["target-input", targetInputRect],
    ["marker-row", markerRowRect],
    ["marker-cell", markerCellRect],
    ["marker", options.markerRect],
  ]);

  const targetCellMarkup = `
    <div data-grid-field-key="${targetCellFieldKey}" data-rect-id="target-cell">
      <input data-testid="${targetTestId}" data-rect-id="target-input" />
      ${markerSameCell ? `<span data-testid="${markerTestId}" data-rect-id="marker">M</span>` : ""}
    </div>
  `;
  const markerCellMarkup = markerSameCell
    ? ""
    : `
      <div data-grid-field-key="${markerCellFieldKey}" data-rect-id="marker-cell">
        <span data-testid="${markerTestId}" data-rect-id="marker">M</span>
      </div>
    `;
  const markerRowMarkup = markerSameRow
    ? ""
    : `
      <div role="row" data-grid-record-id="${markerRecordId}" data-rect-id="marker-row">
        ${markerCellMarkup}
      </div>
    `;

  document.body.innerHTML = `
    <div data-testid="${gridTestId}">
      <div class="${gridScrollportClassName()}" data-rect-id="scrollport">
        <div role="row" data-grid-record-id="${targetRecordId}" data-rect-id="target-row">
          ${targetCellMarkup}
          ${markerSameRow ? markerCellMarkup : ""}
        </div>
        ${markerRowMarkup}
      </div>
    </div>
  `;
  const shell = document.querySelector(dataTestIdSelector(gridTestId));
  const marker = document.querySelector(dataTestIdSelector(markerTestId));
  const target = document.querySelector(dataTestIdSelector(targetTestId));
  if (
    !(shell instanceof HTMLDivElement) ||
    !(marker instanceof HTMLElement) ||
    !(target instanceof HTMLElement)
  ) {
    throw new Error("Expected marker anchor fixture elements to exist");
  }

  vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockImplementation(
    function mockRect(this: HTMLElement) {
      const rectId = this.getAttribute("data-rect-id");
      const box = rectId === null ? undefined : rects.get(rectId);
      return rectFromBox(box ?? { height: 0, left: 0, top: 0, width: 0 });
    },
  );

  return {
    markerTestId,
    page: createBrowserPage({
      [gridTestId]: shell,
      [markerTestId]: marker,
      [targetTestId]: target,
    }),
    targetTestId,
  };
}

function createBrowserPage(
  elements:
    | Record<string, Element | undefined>
    | (() => Record<string, Element | undefined>),
  options: {
    isVisible?: (
      testId: string,
      element: Element,
    ) => boolean | Promise<boolean>;
    onEvaluate?: (testId: string, element: Element) => void;
    onScrollIntoViewIfNeeded?: (testId: string, element: Element) => void;
  } = {},
) {
  const resolveElement = (value: string) => {
    const resolvedElements =
      typeof elements === "function" ? elements() : elements;
    return resolvedElements[value];
  };
  return {
    getByTestId(value: string) {
      return {
        click: async () => {
          const element = resolveElement(value);
          if (element === undefined) {
            throw new Error(`Unknown test id ${value}`);
          }
          if (element instanceof HTMLElement) {
            element.click();
          }
        },
        evaluate: async (
          pageFunction: (element: Element, arg?: unknown) => unknown,
          arg?: unknown,
        ) => {
          const element = resolveElement(value);
          if (element === undefined) {
            throw new Error(`Unknown test id ${value}`);
          }
          options.onEvaluate?.(value, element);
          return pageFunction(element, arg);
        },
        fill: async () => undefined,
        isVisible: async () => {
          const element = resolveElement(value);
          if (element === undefined) {
            return false;
          }
          return options.isVisible?.(value, element) ?? element.isConnected;
        },
        scrollIntoViewIfNeeded: async () => {
          const element = resolveElement(value);
          if (element === undefined) {
            throw new Error(`Unknown test id ${value}`);
          }
          options.onScrollIntoViewIfNeeded?.(value, element);
        },
      };
    },
  };
}

function rectFromBox(options: {
  height: number;
  left: number;
  top: number;
  width: number;
}) {
  return {
    bottom: options.top + options.height,
    height: options.height,
    left: options.left,
    right: options.left + options.width,
    top: options.top,
    width: options.width,
    x: options.left,
    y: options.top,
    toJSON: () => ({}),
  } as DOMRect;
}
