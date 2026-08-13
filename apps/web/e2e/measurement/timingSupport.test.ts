// @vitest-environment jsdom

import {
  gridShellTestId,
  rowCellTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";

import {
  observeBlankRowPaintInBrowser,
  observeCellPaintInBrowser,
} from "./timingSupport";

const inViewport = {
  bottom: 40,
  height: 20,
  left: 10,
  right: 90,
  top: 20,
  width: 80,
  x: 10,
  y: 20,
  toJSON: () => ({}),
};
const viewport = {
  bottom: 100,
  height: 100,
  left: 0,
  right: 100,
  top: 0,
  width: 100,
  x: 0,
  y: 0,
  toJSON: () => ({}),
};
const offscreen = {
  ...inViewport,
  bottom: 220,
  top: 200,
  y: 200,
};

afterEach(() => {
  document.body.replaceChildren();
  vi.unstubAllGlobals();
});

function installFrameClock(onFrame: (frame: number) => void, startAtFrame = 0) {
  let frame = 0;
  let now = 10;
  const marks = startAtFrame === 0 ? [{ name: "accepted", startTime: 5 }] : [];
  vi.stubGlobal("performance", {
    clearMarks: vi.fn(),
    getEntriesByName: (name: string) =>
      marks.filter((mark) => mark.name === name),
    getEntriesByType: () => marks,
    mark: (name: string) => marks.push({ name, startTime: now }),
    now: () => now,
  });
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    queueMicrotask(() => {
      frame += 1;
      now += 16;
      if (frame === startAtFrame)
        marks.push({ name: "accepted", startTime: now });
      onFrame(frame);
      callback(now);
    });
    return frame;
  });
  return () => frame;
}

function cellFixture() {
  document.body.innerHTML = `
    <div role="grid">
      <div role="gridcell" tabindex="0">
        <input data-testid="row-a-timeline.activity_synopsis_text" value="priorx" />
      </div>
    </div>
  `;
  const grid = document.querySelector<HTMLElement>(
    '[role="grid"]',
  ) as HTMLElement;
  const cell = document.querySelector<HTMLElement>(
    '[role="gridcell"]',
  ) as HTMLElement;
  const input = document.querySelector<HTMLInputElement>(
    "input",
  ) as HTMLInputElement;
  grid.getBoundingClientRect = () => viewport;
  cell.getBoundingClientRect = () => inViewport;
  input.getBoundingClientRect = () => inViewport;
  input.focus();
  input.setSelectionRange(input.value.length, input.value.length);
  return { cell, input };
}

const cellCases: Array<{
  name: string;
  arrange: (fixture: ReturnType<typeof cellFixture>) => () => void;
}> = [
  {
    name: "wrong record or field",
    arrange: ({ input }) => {
      input.dataset.testid = "row-b-timeline.raw_activity_text";
      return () => {
        input.dataset.testid = "row-a-timeline.activity_synopsis_text";
      };
    },
  },
  {
    name: "hidden editor",
    arrange: ({ input }) => {
      input.style.visibility = "hidden";
      return () => {
        input.style.visibility = "visible";
      };
    },
  },
  {
    name: "offscreen editor",
    arrange: ({ input }) => {
      input.getBoundingClientRect = () => offscreen;
      return () => {
        input.getBoundingClientRect = () => inViewport;
      };
    },
  },
  {
    name: "wrong focus and caret",
    arrange: ({ cell, input }) => {
      cell.focus();
      input.setSelectionRange(0, 1);
      return () => {
        input.focus();
        input.setSelectionRange(input.value.length, input.value.length);
      };
    },
  },
  {
    name: "wrong typed value",
    arrange: ({ input }) => {
      input.value = "prior";
      return () => {
        input.value = "priorx";
        input.setSelectionRange(input.value.length, input.value.length);
      };
    },
  },
];

function blankRowFixture() {
  const summaryTestId = rowCellTestId(
    "record-a",
    "timeline.activity_synopsis_text",
  );
  const versionTestId = timelineRowVersionTestId("record-a");
  document.body.innerHTML = `
    <div role="grid" data-testid="${gridShellTestId("cartulary.view.timeline.v2")}">
      <div data-grid-record-id="record-a">
        <span data-testid="${summaryTestId}">created</span>
        <span data-testid="${versionTestId}">1</span>
      </div>
    </div>
  `;
  const grid = document.querySelector<HTMLElement>(
    '[role="grid"]',
  ) as HTMLElement;
  const row = document.querySelector<HTMLElement>(
    "[data-grid-record-id]",
  ) as HTMLElement;
  const testIdElements = Array.from(
    document.querySelectorAll<HTMLElement>("[data-testid]"),
  );
  const summary = testIdElements.find(
    (element) => element.dataset.testid === summaryTestId,
  ) as HTMLElement;
  const version = testIdElements.find(
    (element) => element.dataset.testid === versionTestId,
  ) as HTMLElement;
  grid.getBoundingClientRect = () => viewport;
  summary.getBoundingClientRect = () => inViewport;
  return { row, summary, version };
}

describe("AC-043 paint qualification", () => {
  it("rejects wrong, hidden, offscreen, unfocused, unstable, and pre-acceptance states", async () => {
    for (const testCase of cellCases) {
      document.body.replaceChildren();
      const fixture = cellFixture();
      const repair = testCase.arrange(fixture);
      const frameCount = installFrameClock((frame) => {
        if (frame === 2) repair();
      });

      await observeCellPaintInBrowser({
        expectedValue: "priorx",
        mode: "editor",
        startMark: "accepted",
        stopMark: "visible",
        testId: "row-a-timeline.activity_synopsis_text",
        timeoutMs: 1_000,
      });

      expect(frameCount(), testCase.name).toBeGreaterThanOrEqual(3);
    }

    document.body.replaceChildren();
    const { cell, input } = cellFixture();
    cell.focus();
    const selectedFrameCount = installFrameClock((frame) => {
      if (frame !== 2) return;
      const replacement = input.cloneNode(true) as HTMLInputElement;
      replacement.getBoundingClientRect = () => inViewport;
      input.replaceWith(replacement);
      cell.focus();
    });

    await observeCellPaintInBrowser({
      mode: "selected",
      startMark: "accepted",
      stopMark: "visible",
      testId: "row-a-timeline.activity_synopsis_text",
      timeoutMs: 1_000,
    });
    expect(selectedFrameCount()).toBeGreaterThanOrEqual(3);

    document.body.replaceChildren();
    cellFixture();
    const preAcceptanceFrameCount = installFrameClock(() => {}, 2);
    await observeCellPaintInBrowser({
      expectedValue: "priorx",
      mode: "editor",
      startMark: "accepted",
      stopMark: "visible",
      testId: "row-a-timeline.activity_synopsis_text",
      timeoutMs: 1_000,
    });
    expect(preAcceptanceFrameCount()).toBeGreaterThanOrEqual(3);

    document.body.replaceChildren();
    const invalidBlankRow = blankRowFixture();
    invalidBlankRow.summary.textContent = "wrong";
    invalidBlankRow.summary.style.visibility = "hidden";
    invalidBlankRow.summary.getBoundingClientRect = () => offscreen;
    invalidBlankRow.version.textContent = "0";
    const invalidBlankRowFrameCount = installFrameClock((frame) => {
      if (frame !== 2) return;
      invalidBlankRow.summary.textContent = "created";
      invalidBlankRow.summary.style.visibility = "visible";
      invalidBlankRow.summary.getBoundingClientRect = () => inViewport;
      invalidBlankRow.version.textContent = "1";
    });

    await observeBlankRowPaintInBrowser({
      expectedSummary: "created",
      startMark: "accepted",
      stopMark: "visible",
      timeoutMs: 1_000,
    });

    expect(invalidBlankRowFrameCount()).toBeGreaterThanOrEqual(3);
    document.body.replaceChildren();
    const unstableBlankRow = blankRowFixture();
    const unstableBlankRowFrameCount = installFrameClock((frame) => {
      if (frame !== 2) return;
      unstableBlankRow.row.dataset.gridRecordId = "record-b";
      unstableBlankRow.summary.dataset.testid =
        "row-record-b-timeline.activity_synopsis_text";
      unstableBlankRow.version.dataset.testid = "row-record-b-row_version";
      unstableBlankRow.version.textContent = "2";
    });

    await observeBlankRowPaintInBrowser({
      expectedSummary: "created",
      startMark: "accepted",
      stopMark: "visible",
      timeoutMs: 1_000,
    });

    expect(unstableBlankRowFrameCount()).toBeGreaterThanOrEqual(3);
  });
});
