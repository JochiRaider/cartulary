// @vitest-environment jsdom
import { cartularyAc003PerformanceContract as contract } from "@cartulary/ui-contracts";
import { afterEach, expect, it, vi } from "vitest";
import { observePastePaintInBrowser } from "./pasteTimingSupport";

function at<T>(values: ArrayLike<T>, index: number): T {
  const value = values[index];
  if (value === undefined) throw new Error(`missing fixture item ${index}`);
  return value;
}

afterEach(() => {
  document.body.replaceChildren();
  vi.unstubAllGlobals();
});

it("requires 100 visible cells on 20 new stable records across two accepted frames", async () => {
  const create = (onFrame: (frame: number) => void = () => {}) => {
    document.body.replaceChildren();
    vi.stubGlobal("innerWidth", 1920);
    vi.stubGlobal("innerHeight", 1080);
    vi.stubGlobal("devicePixelRatio", 1);
    let now = 10,
      frame = 0;
    const marks = [{ name: "accepted", startTime: 5 }];
    vi.stubGlobal("performance", {
      now: () => now,
      getEntriesByName: (name: string) =>
        marks.filter((mark) => mark.name === name),
      mark: (name: string) => marks.push({ name, startTime: now }),
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
    const rect = (x: number, y: number, width: number, height: number) => ({
      x,
      y,
      width,
      height,
      left: x,
      top: y,
      right: x + width,
      bottom: y + height,
      toJSON: () => ({}),
    });
    const grid = document.createElement("div");
    grid.id = "grid";
    grid.getBoundingClientRect = () => rect(0, 0, 1900, 1000);
    const rows = contract.rows.map((values, index) => {
      const row = document.createElement("div");
      row.dataset.gridRecordId = `00000000-0000-0000-0000-${String(index + 1).padStart(12, "0")}`;
      row.dataset.version = "1";
      for (const [column, value] of values.entries()) {
        const cell = document.createElement("span");
        cell.dataset.field = String(column);
        cell.textContent = value;
        cell.getBoundingClientRect = () =>
          rect(column * 100, index * 30, 100, 30);
        row.append(cell);
      }
      grid.append(row);
      return row;
    });
    document.body.append(grid);
    return {
      rows,
      marks,
      frames: () => frame,
      options: {
        gridSelector: "#grid",
        rowSelector: "[data-grid-record-id]",
        fieldSelectors: [0, 1, 2, 3, 4].map(
          (column) => `[data-field='${column}']`,
        ),
        versionAttribute: "data-version",
        existingRecordIds: [] as string[],
        expectedRows: contract.rows.map((row) => [...row]),
        startMark: "accepted",
        stopMark: "visible",
        viewportWidth: 1920,
        viewportHeight: 1080,
        deviceScaleFactor: 1,
        timeoutMs: 100,
      },
    };
  };
  const valid = create();
  const result = await observePastePaintInBrowser(valid.options);
  expect(valid.frames()).toBe(2);
  expect(result.createdRecords).toHaveLength(20);
  expect(new Set(result.createdRecords.map((row) => row.record_id)).size).toBe(
    20,
  );
  expect(result.stoppedAt - result.acceptedAt).toBe(37);

  const mutations = [
    (fixture: ReturnType<typeof create>) => {
      fixture.options.expectedRows.pop();
    },
    (fixture: ReturnType<typeof create>) => {
      at(fixture.options.expectedRows, 0).pop();
    },
    (fixture: ReturnType<typeof create>) => {
      fixture.options.existingRecordIds.push(
        at(fixture.rows, 0).getAttribute("data-grid-record-id") ?? "",
      );
    },
    (fixture: ReturnType<typeof create>) => {
      at(fixture.rows, 1).dataset.gridRecordId = at(
        fixture.rows,
        0,
      ).dataset.gridRecordId;
    },
    (fixture: ReturnType<typeof create>) => {
      at(fixture.rows, 0).dataset.version = "1.5";
    },
    (fixture: ReturnType<typeof create>) => {
      at(at(fixture.rows, 0).children, 4).remove();
    },
    (fixture: ReturnType<typeof create>) => {
      at(at(fixture.rows, 0).children, 4).textContent = "stale";
    },
    (fixture: ReturnType<typeof create>) => {
      (at(fixture.rows, 0).children[4] as HTMLElement).style.visibility =
        "hidden";
    },
    (fixture: ReturnType<typeof create>) => {
      at(at(fixture.rows, 0).children, 4).getBoundingClientRect = () => ({
        x: 2000,
        y: 0,
        width: 10,
        height: 10,
        left: 2000,
        right: 2010,
        top: 0,
        bottom: 10,
        toJSON: () => ({}),
      });
    },
    (fixture: ReturnType<typeof create>) => {
      fixture.marks.length = 0;
    },
    (fixture: ReturnType<typeof create>) => {
      fixture.marks.push({ name: "accepted", startTime: 6 });
    },
    () => {
      vi.stubGlobal("innerWidth", 1919);
    },
  ];
  for (const mutate of mutations) {
    const fixture = create();
    mutate(fixture);
    await expect(observePastePaintInBrowser(fixture.options)).rejects.toThrow();
  }
  const oneFrame = create((frame) => {
    if (frame === 2) at(oneFrame.rows, 0).remove();
  });
  await expect(observePastePaintInBrowser(oneFrame.options)).rejects.toThrow(
    /visible completion/,
  );
  const changedBinding = create((frame) => {
    if (frame === 2) at(changedBinding.rows, 0).dataset.version = "2";
  });
  await observePastePaintInBrowser(changedBinding.options);
  expect(changedBinding.frames()).toBe(3);
});
