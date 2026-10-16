// @vitest-environment jsdom
import { afterEach, expect, it, vi } from "vitest";
import {
  observePresencePaintInBrowser,
  type PresencePaintInput,
  presenceElapsed,
} from "./presenceTimingSupport";

const input: PresencePaintInput = {
  kind: "cell",
  actorUserId: "actor",
  markerSelector: "#marker",
  recordId: "record",
  fieldKey: "timeline.activity_synopsis_text",
  viewSchemaId: "cartulary.view.timeline.v2",
  timeoutMs: 30,
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
  toJSON: () => ({}),
});

function fixture() {
  document.body.innerHTML =
    '<section data-workbook-view-schema-id="cartulary.view.timeline.v2"><div role="grid"><div data-grid-record-id="record" data-grid-row-version="1"><div data-grid-field-key="timeline.activity_synopsis_text"><span id="marker"></span></div></div></div></section>';
  const grid = document.querySelector<HTMLElement>('[role="grid"]');
  if (grid === null) throw new Error("Presence DOM fixture has no grid");
  grid.getBoundingClientRect = () => ({
    ...box(0),
    right: 1920,
    bottom: 1080,
    width: 1920,
    height: 1080,
  });
  const row = document.querySelector<HTMLElement>("[data-grid-record-id]");
  const field = document.querySelector<HTMLElement>("[data-grid-field-key]");
  const marker = document.querySelector<HTMLElement>("#marker");
  if (!row || !field || !marker)
    throw new Error("Presence DOM fixture is incomplete");
  const add = () => {
    const node = document.createElement("span");
    node.dataset.presenceUserId = "actor";
    node.dataset.presenceMode = "editing";
    node.textContent = "A";
    node.getBoundingClientRect = () => box();
    marker.append(node);
    return node;
  };
  return { grid, row, field, marker, add };
}

function clock(
  onFrame: (frame: number) => void,
  origin: () => number = () => 1000,
) {
  let frame = 0;
  vi.stubGlobal("innerWidth", 1920);
  vi.stubGlobal("innerHeight", 1080);
  vi.stubGlobal("devicePixelRatio", 1);
  vi.stubGlobal("performance", {
    get timeOrigin() {
      return origin();
    },
    now: () => frame * 5,
  });
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    queueMicrotask(() => {
      frame++;
      onFrame(frame);
      callback(frame * 5);
    });
    return frame;
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
  document.body.innerHTML = "";
});

it("rejects stale, unbound and nonvisible presence and invalid cross-context timing", async () => {
  let f = fixture();
  clock((frame) => {
    if (frame === 1) f.add();
  });
  await expect(observePresencePaintInBrowser(input)).resolves.toEqual({
    origin: 1000,
    stoppedAt: 10,
    rowVersion: 1,
  });
  // RDG rows have no CSS box. Their overflow cannot clip a painted child.
  f = fixture();
  f.row.style.display = "contents";
  f.row.style.overflow = "hidden";
  clock((frame) => {
    if (frame === 1) f.add();
  });
  await expect(observePresencePaintInBrowser(input)).resolves.toMatchObject({
    rowVersion: 1,
  });

  for (const mutate of [
    (node: HTMLElement) => {
      node.dataset.presenceUserId = "different";
    },
    (node: HTMLElement) => {
      node.dataset.presenceMode = "viewing";
    },
    (node: HTMLElement) => {
      node.hidden = true;
    },
    (node: HTMLElement) => {
      node.style.opacity = "0";
    },
    (node: HTMLElement) => {
      node.style.visibility = "hidden";
    },
    (node: HTMLElement) => {
      node.getBoundingClientRect = () => box(2000);
    },
    (_node: HTMLElement) => {
      f.row.dataset.gridRecordId = "different";
    },
    (_node: HTMLElement) => {
      f.row.dataset.gridRowVersion = "0";
    },
    (_node: HTMLElement) => {
      f.field.dataset.gridFieldKey = "different";
    },
    (_node: HTMLElement) => {
      f.grid.setAttribute("inert", "");
    },
    (_node: HTMLElement) => {
      f.add();
    },
  ]) {
    f = fixture();
    clock((frame) => {
      if (frame === 1) mutate(f.add());
    });
    await expect(observePresencePaintInBrowser(input)).rejects.toThrow(
      "two visible frames",
    );
  }
  f = fixture();
  clock(() => {});
  f.add();
  await expect(observePresencePaintInBrowser(input)).rejects.toThrow(
    "already present",
  );
  f = fixture();
  clock((frame) => {
    if (frame === 1) f.add();
    else f.marker.replaceChildren();
  });
  await expect(observePresencePaintInBrowser(input)).rejects.toThrow(
    "two visible frames",
  );
  f = fixture();
  clock((frame) => {
    if (frame === 1) f.add();
    f.row.dataset.gridRowVersion = String(frame);
  });
  await expect(observePresencePaintInBrowser(input)).rejects.toThrow(
    "two visible frames",
  );
  let origin = 1000;
  f = fixture();
  clock(
    (frame) => {
      if (frame === 1) f.add();
      else origin++;
    },
    () => origin,
  );
  await expect(observePresencePaintInBrowser(input)).rejects.toThrow(
    "clock origin or geometry",
  );
  f = fixture();
  clock((frame) => {
    if (frame === 1) f.add();
    else vi.stubGlobal("innerWidth", 1280);
  });
  await expect(observePresencePaintInBrowser(input)).rejects.toThrow(
    "clock origin or geometry",
  );
  f = fixture();
  clock((frame) => {
    if (frame === 1) f.add().dataset.presenceMode = "focused";
  });
  await expect(
    observePresencePaintInBrowser({ ...input, kind: "row", fieldKey: null }),
  ).resolves.toMatchObject({ rowVersion: 1 });
  f = fixture();
  document.body.append(f.marker);
  clock((frame) => {
    if (frame === 1) f.add().dataset.presenceMode = "viewing";
  });
  await expect(
    observePresencePaintInBrowser({
      ...input,
      kind: "surface",
      recordId: null,
      fieldKey: null,
    }),
  ).resolves.toMatchObject({ rowVersion: null });

  expect(
    presenceElapsed(
      { origin: 1000, acceptedAt: 200 },
      { origin: 1100, stoppedAt: 150, rowVersion: null },
    ),
  ).toBe(50);
  for (const invalid of [NaN, Infinity, -1])
    expect(() =>
      presenceElapsed(
        { origin: 1000, acceptedAt: invalid },
        { origin: 1100, stoppedAt: 150, rowVersion: null },
      ),
    ).toThrow();
  expect(() =>
    presenceElapsed(
      { origin: 1000, acceptedAt: 300 },
      { origin: 1100, stoppedAt: 150, rowVersion: null },
    ),
  ).toThrow();
});
