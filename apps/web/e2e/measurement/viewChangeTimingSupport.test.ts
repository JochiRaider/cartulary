// @vitest-environment jsdom
import { afterEach, expect, it, vi } from "vitest";
import {
  observeViewChangePaintInBrowser,
  type ViewChangePaintInput,
} from "./viewChangeTimingSupport";

afterEach(() => {
  document.body.replaceChildren();
  vi.unstubAllGlobals();
});

const rect = (top: number, bottom: number) => ({
  top,
  bottom,
  left: 0,
  right: 100,
  x: 0,
  y: top,
  width: 100,
  height: bottom - top,
  toJSON: () => ({}),
});
const input: ViewChangePaintInput = {
  operation: "sort",
  viewSchemaId: "cartulary.view.timeline.v2",
  startMark: "accepted",
  stopMark: "paint",
  completion: "stable_viewport",
  expectedRows: [
    { recordId: "a", minimumVersion: 2 },
    { recordId: "b", minimumVersion: 3 },
  ],
  groupId: null,
  timeoutMs: 90,
};

function fixture() {
  document.body.innerHTML =
    '<section data-workbook-view-schema-id="cartulary.view.timeline.v2" data-workbook-query-revision="2" data-workbook-query-pending="false" data-workbook-query-filter-count="0" data-workbook-query-sort-count="1" data-workbook-query-group-field=""><div role="grid"><div role="row" data-grid-record-id="a" data-grid-row-version="2"><div role="gridcell" tabindex="0"></div></div><div role="row" data-grid-record-id="b" data-grid-row-version="3"><div role="gridcell"></div></div></div></section>';
  const frame = document.querySelector<HTMLElement>("section");
  if (frame === null) throw new Error("View-change DOM fixture has no frame");
  const grid = frame.querySelector<HTMLElement>('[role="grid"]');
  if (grid === null) throw new Error("View-change DOM fixture has no grid");
  const [firstRow, secondRow] =
    grid.querySelectorAll<HTMLElement>('[role="row"]');
  const [firstCell, secondCell] =
    grid.querySelectorAll<HTMLElement>('[role="gridcell"]');
  if (!firstRow || !secondRow || !firstCell || !secondCell)
    throw new Error("View-change DOM fixture requires two rows and cells");
  const rows = [firstRow, secondRow] as const;
  const cells = [firstCell, secondCell] as const;
  grid.getBoundingClientRect = () => rect(0, 100);
  for (const row of rows) row.getBoundingClientRect = () => rect(0, 0);
  cells[0].getBoundingClientRect = () => rect(20, 60);
  cells[1].getBoundingClientRect = () => rect(60, 100);
  const entry = firstCell;
  entry.getBoundingClientRect = () => rect(20, 60);
  return { frame, grid, rows, cells, entry };
}

function clock(onFrame: (frame: number) => void, accepted = true) {
  let frame = 0;
  let now = 5;
  const marks = accepted
    ? [
        {
          name: "accepted",
          startTime: 2,
          detail: { viewSchemaId: input.viewSchemaId, afterRevision: 1 },
        },
      ]
    : [];
  vi.stubGlobal("performance", {
    now: () => now,
    getEntriesByName: (name: string) =>
      marks.filter((mark) => mark.name === name),
    mark: vi.fn(),
  });
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    queueMicrotask(() => {
      frame++;
      now += 16;
      onFrame(frame);
      callback(now);
    });
    return frame;
  });
  vi.stubGlobal("innerHeight", 100);
  vi.stubGlobal("innerWidth", 100);
  return { marks, frames: () => frame };
}

it("rejects stale query bindings, invalid windows, missing acceptance and unstable paint", async () => {
  for (const invalidate of [
    (f: ReturnType<typeof fixture>) => {
      f.frame.dataset.workbookQuerySortCount = "0";
    },
    (f: ReturnType<typeof fixture>) => {
      f.frame.dataset.workbookQueryFilterCount = "1";
    },
    (f: ReturnType<typeof fixture>) => {
      f.frame.dataset.workbookQueryRevision = "1";
    },
    (f: ReturnType<typeof fixture>) => {
      f.frame.dataset.workbookQueryPending = "true";
    },
    (f: ReturnType<typeof fixture>) => {
      f.rows[0].dataset.gridRecordId = "wrong";
    },
    (f: ReturnType<typeof fixture>) => {
      f.rows[0].dataset.gridRecordId = "b";
      f.rows[1].dataset.gridRecordId = "a";
    },
    (f: ReturnType<typeof fixture>) => {
      f.rows[1].dataset.gridRowVersion = "2";
    },
    (f: ReturnType<typeof fixture>) => {
      f.rows[1].dataset.gridRowVersion = "NaN";
    },
    (f: ReturnType<typeof fixture>) => {
      f.cells[1].getBoundingClientRect = () => rect(110, 150);
    },
    (f: ReturnType<typeof fixture>) => {
      f.cells[1].getBoundingClientRect = () => rect(60, 90);
    },
    (f: ReturnType<typeof fixture>) => {
      f.cells[1].getBoundingClientRect = () => rect(70, 100);
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.tabIndex = -1;
    },
    (f: ReturnType<typeof fixture>) => {
      f.frame.hidden = true;
    },
  ]) {
    const f = fixture();
    invalidate(f);
    clock(() => {});
    await expect(observeViewChangePaintInBrowser(input)).rejects.toThrow(
      "did not qualify",
    );
  }
  fixture();
  clock(() => {}, false);
  await expect(observeViewChangePaintInBrowser(input)).rejects.toThrow(
    "did not qualify",
  );
  fixture();
  const duplicate = clock(() => {});
  const original = duplicate.marks[0];
  if (original === undefined)
    throw new Error("View-change fixture has no acceptance mark");
  duplicate.marks.push(original);
  await expect(observeViewChangePaintInBrowser(input)).rejects.toThrow(
    "more than once",
  );
  const f = fixture();
  clock((frame) => {
    f.rows[1].dataset.gridRowVersion = frame % 2 === 0 ? "4" : "3";
  });
  await expect(observeViewChangePaintInBrowser(input)).rejects.toThrow(
    "did not qualify",
  );
  fixture();
  const passing = clock(() => {});
  await expect(observeViewChangePaintInBrowser(input)).resolves.toEqual({
    acceptedAt: 2,
    stoppedAt: 37,
  });
  expect(passing.frames()).toBe(2);
  fixture();
  clock(() => {});
  await expect(
    observeViewChangePaintInBrowser({ ...input, groupId: "s:disabled" }),
  ).rejects.toThrow("did not qualify");
});
