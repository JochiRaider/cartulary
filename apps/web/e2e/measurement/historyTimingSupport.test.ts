// @vitest-environment jsdom
import { afterEach, expect, it, vi } from "vitest";
import {
  type HistoryPaintInput,
  observeHistoryPaintInBrowser,
} from "./historyTimingSupport";

const input: HistoryPaintInput = {
  kind: "history_entry",
  recordId: "record",
  actorId: "actor",
  priorVersion: 3,
  priorHistoryRefs: ["prior"],
  priorChangeSetIds: ["prior-change"],
  gridSelector: "#grid",
  historySelector: "#history",
  startMark: "accept",
  timeoutMs: 35,
  cells: [
    {
      selector: "#analyst",
      fieldKey: "timeline.analyst_text",
      value: "history-before",
    },
    {
      selector: "#device",
      fieldKey: "timeline.device_object_text",
      value: "device-before",
    },
  ],
};
const box = (left = 10) => ({
  left,
  top: 10,
  right: left + 20,
  bottom: 30,
  width: 20,
  height: 20,
  x: left,
  y: 10,
  toJSON() {},
});

function fixture() {
  document.body.innerHTML =
    '<div id="grid"><div data-grid-record-id="record" data-grid-row-version="4" style="display:contents;overflow:hidden"><div data-grid-field-key="timeline.analyst_text"><span id="analyst">history-before</span></div><div data-grid-field-key="timeline.device_object_text"><span id="device">device-before</span></div></div></div><section id="history" data-history-record-id="record"><ol><li data-history-item-ref="new" data-history-change-set-id="new-change" data-history-actor-id="actor" data-history-operation="rollback" data-history-revision="4"><strong>Reversed</strong><span data-generated-metadata="history-actor">Reviewer</span><time data-generated-metadata="history-time" datetime="2026-04-05T00:00:00Z">April 5</time></li></ol></section>';
  for (const node of document.querySelectorAll<HTMLElement>("*"))
    node.getBoundingClientRect = () => box();
  for (const node of document.querySelectorAll<HTMLElement>("#grid,#history"))
    node.getBoundingClientRect = () => ({
      ...box(0),
      right: 1920,
      bottom: 1080,
      width: 1920,
      height: 1080,
    });
  const grid = document.querySelector<HTMLElement>("#grid");
  const row = document.querySelector<HTMLElement>("[data-grid-record-id]");
  const entry = document.querySelector<HTMLElement>("[data-history-item-ref]");
  const cell = document.querySelector<HTMLElement>("#analyst");
  const time = document.querySelector<HTMLTimeElement>("time");
  const history = document.querySelector<HTMLElement>("#history");
  if (!grid || !row || !entry || !cell || !time || !history)
    throw new Error("History DOM fixture is incomplete");
  return { grid, row, entry, cell, time, history };
}
function clock(
  onFrame: (frame: number) => void = () => {},
  marks?: () => unknown[],
) {
  let frame = 0;
  vi.stubGlobal("innerWidth", 1920);
  vi.stubGlobal("innerHeight", 1080);
  vi.stubGlobal("devicePixelRatio", 1);
  const mark = {
    startTime: 1,
    detail: {
      recordId: "record",
      rowVersion: 3,
      action: input.kind,
      viewSchemaId: "cartulary.view.timeline.v2",
    },
  };
  vi.stubGlobal("performance", {
    timeOrigin: 1000,
    now: () => frame * 5,
    getEntriesByName: marks ?? (() => [mark]),
  });
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    queueMicrotask(() => {
      frame++;
      onFrame(frame);
      callback(frame * 5);
    });
    return frame;
  });
  return mark;
}
afterEach(() => {
  vi.unstubAllGlobals();
  document.body.innerHTML = "";
});

it("requires one accepted history action and stable visible restored cells with new attribution", async () => {
  fixture();
  clock();
  await expect(observeHistoryPaintInBrowser(input)).resolves.toEqual({
    elapsed: 9,
    rowVersion: 4,
    historyRef: "new",
    changeSetId: "new-change",
    committedAt: "2026-04-05T00:00:00Z",
  });
  for (const mutate of [
    (f: ReturnType<typeof fixture>) => {
      f.row.dataset.gridRowVersion = "3";
    },
    (f: ReturnType<typeof fixture>) => {
      f.row.dataset.gridRowVersion = "5";
    },
    (f: ReturnType<typeof fixture>) => {
      f.row.dataset.gridRecordId = "other";
    },
    (f: ReturnType<typeof fixture>) => {
      f.cell.textContent = "mistake";
    },
    (f: ReturnType<typeof fixture>) => {
      const parent = f.cell.parentElement;
      if (parent === null)
        throw new Error("History fixture cell has no field wrapper");
      parent.dataset.gridFieldKey = "other";
    },
    (f: ReturnType<typeof fixture>) => {
      f.cell.getBoundingClientRect = () => box(2000);
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.dataset.historyItemRef = "prior";
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.dataset.historyChangeSetId = "prior-change";
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.dataset.historyActorId = "other";
    },
    (f: ReturnType<typeof fixture>) => {
      f.history.dataset.historyRecordId = "other";
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.dataset.historyOperation = "patch";
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.dataset.historyRevision = "3";
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.hidden = true;
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.style.opacity = "0";
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.setAttribute("inert", "");
    },
    (f: ReturnType<typeof fixture>) => {
      f.time.setAttribute("datetime", "invalid");
    },
    (f: ReturnType<typeof fixture>) => {
      f.time.textContent = "";
    },
    (f: ReturnType<typeof fixture>) => {
      f.time.getBoundingClientRect = () => box(2000);
    },
    (f: ReturnType<typeof fixture>) => {
      f.entry.after(f.entry.cloneNode(true));
    },
  ]) {
    const f = fixture();
    mutate(f);
    clock();
    await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(
      /two visible frames/u,
    );
  }
  fixture();
  clock(
    () => {},
    () => [],
  );
  await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(
    /two visible frames/u,
  );
  fixture();
  clock(
    () => {},
    () => [{}, {}],
  );
  await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(
    /repeated/u,
  );
  fixture();
  const mark = clock();
  mark.detail.recordId = "other";
  await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(
    /target does not match/u,
  );
  let f = fixture();
  clock((frame) => {
    if (frame >= 2) f.cell.textContent = "mistake";
  });
  await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(
    /two visible frames/u,
  );
  f = fixture();
  clock((frame) => {
    f.entry.dataset.historyItemRef = `new-${frame}`;
  });
  await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(
    /two visible frames/u,
  );
  fixture();
  clock((frame) => {
    if (frame === 2) vi.stubGlobal("innerWidth", 1900);
  });
  await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(
    /geometry/u,
  );
  f = fixture();
  Object.defineProperty(f.grid, "scrollHeight", { value: 200 });
  clock((frame) => {
    if (frame === 2) f.grid.scrollTop = 1;
  });
  await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(/scroll/u);
  f = fixture();
  Object.defineProperty(f.history, "scrollHeight", { value: 200 });
  clock((frame) => {
    if (frame === 2) f.history.scrollTop = 1;
  });
  await expect(observeHistoryPaintInBrowser(input)).resolves.toMatchObject({
    rowVersion: 4,
  });
  f = fixture();
  Object.defineProperty(f.history, "scrollHeight", { value: 200 });
  clock((frame) => {
    if (frame === 2) {
      f.history.scrollTop = 1;
      f.time.getBoundingClientRect = () => box(2200);
    }
  });
  await expect(observeHistoryPaintInBrowser(input)).rejects.toThrow(
    /two visible frames/u,
  );
  for (const kind of ["change_set", "row_restore"] as const) {
    f = fixture();
    const mark = clock();
    mark.detail.action = kind;
    if (kind === "row_restore")
      f.entry.dataset.historyOperation = "row_restore";
    await expect(
      observeHistoryPaintInBrowser({ ...input, kind }),
    ).resolves.toMatchObject({ rowVersion: 4 });
  }
});
