// @vitest-environment jsdom
import { afterEach, expect, it, vi } from "vitest";
import { parseFixtureRuntimeBundle } from "../support/performance/runtimeBundle";
import { observeEvidenceInspectorPaintInBrowser } from "./evidenceInspectorTimingSupport";

function required<T>(value: T | null | undefined): T {
  if (value === null || value === undefined)
    throw new Error("missing test fixture value");
  return value;
}

afterEach(() => {
  document.body.replaceChildren();
  vi.unstubAllGlobals();
});
it("requires fresh metadata and five exact visible Evidence bindings across two frames", async () => {
  const create = (onFrame: (frame: number) => void = () => {}) => {
    document.body.replaceChildren();
    vi.stubGlobal("innerWidth", 1920);
    vi.stubGlobal("innerHeight", 1080);
    vi.stubGlobal("devicePixelRatio", 1);
    let now = 10,
      frame = 0;
    const marks = [{ name: "accepted", startTime: 10 }];
    const resources = [
      {
        name: "http://localhost/api/v1/incidents/incident/evidence-metadata",
        startTime: 11,
        responseEnd: 15,
      },
    ];
    vi.stubGlobal("performance", {
      now: () => now,
      getEntriesByName: (name: string) =>
        marks.filter((mark) => mark.name === name),
      getEntriesByType: () => resources,
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
    const target = {
      record_id: "00000000-0000-0000-0000-000000000100",
      row_version: 2,
      summary: "Evidence target",
      items: Array.from({ length: 5 }, (_, i) => ({
        record_id: `00000000-0000-0000-0000-${String(i + 1).padStart(12, "0")}`,
        row_version: 1,
        title: `Evidence ${i + 1}`,
      })),
    };
    const inspector = document.createElement("aside");
    inspector.id = "inspector";
    inspector.dataset.recordId = target.record_id;
    inspector.dataset.rowVersion = "2";
    inspector.innerHTML =
      '<h2>Evidence target</h2><section id="evidence"><p>Attached evidence count: 100</p><p>Linked evidence: 100</p><ul></ul></section>';
    const list = required(inspector.querySelector("ul"));
    const entries = target.items.map((item) => {
      const entry = document.createElement("li");
      entry.dataset.evidenceRecordId = item.record_id;
      entry.dataset.evidenceRowVersion = "1";
      entry.dataset.evidenceMetadataState = "ready";
      entry.innerHTML = `<p>${item.title}</p><p>Available · File available</p><button aria-label="Preview ${item.title}">Preview</button>`;
      list.append(entry);
      return entry;
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
    inspector.getBoundingClientRect = () => rect(1400, 0, 520, 1000);
    inspector.querySelectorAll<HTMLElement>("*").forEach((el, index) => {
      el.getBoundingClientRect = () => rect(1500, 20 + index * 20, 300, 20);
    });
    document.body.append(inspector);
    return {
      inspector,
      entries,
      marks,
      resources,
      frames: () => frame,
      options: {
        inspectorSelector: "#inspector",
        evidenceSelector: "#evidence",
        metadataPath: "/api/v1/incidents/incident/evidence-metadata",
        target,
        startMark: "accepted",
        stopMark: "paint",
        timeoutMs: 100,
      },
    };
  };
  const valid = create();
  const result = await observeEvidenceInspectorPaintInBrowser(valid.options);
  expect(valid.frames()).toBe(2);
  expect(result).toEqual({
    accepted_at: 10,
    first_frame_at: 26,
    second_frame_at: 42,
    metadata_started_at: 11,
    metadata_completed_at: 15,
  });
  const intersecting = create();
  required(intersecting.entries[0]).getBoundingClientRect = () => ({
    x: 1390,
    y: 20,
    width: 30,
    height: 20,
    left: 1390,
    right: 1420,
    top: 20,
    bottom: 40,
    toJSON: () => ({}),
  });
  await expect(
    observeEvidenceInspectorPaintInBrowser(intersecting.options),
  ).resolves.toEqual(result);
  for (const mutate of [
    (f: ReturnType<typeof create>) => {
      f.options.target.items.pop();
    },
    (f: ReturnType<typeof create>) => {
      required(f.options.target.items[1]).record_id = required(
        f.options.target.items[0],
      ).record_id;
    },
    (f: ReturnType<typeof create>) => {
      f.inspector.dataset.rowVersion = "3";
    },
    (f: ReturnType<typeof create>) => {
      required(f.entries[0]).dataset.evidenceRowVersion = "2";
    },
    (f: ReturnType<typeof create>) => {
      required(f.entries[0]).dataset.evidenceMetadataState = "loading";
    },
    (f: ReturnType<typeof create>) => {
      required(required(f.entries[0]).querySelector("p")).textContent = "stale";
    },
    (f: ReturnType<typeof create>) => {
      required(required(f.entries[0]).querySelectorAll("p")[1]).textContent =
        "Requested · No file attached";
    },
    (f: ReturnType<typeof create>) => {
      required(required(f.entries[0]).querySelector("button")).setAttribute(
        "aria-disabled",
        "true",
      );
    },
    (f: ReturnType<typeof create>) => {
      required(f.entries[0]).style.visibility = "hidden";
    },
    (f: ReturnType<typeof create>) => {
      required(f.entries[0]).getBoundingClientRect = () => ({
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
    (f: ReturnType<typeof create>) => {
      f.marks.length = 0;
    },
    (f: ReturnType<typeof create>) => {
      f.marks.push(required(f.marks[0]));
    },
    (f: ReturnType<typeof create>) => {
      f.resources.length = 0;
    },
    (f: ReturnType<typeof create>) => {
      f.resources.push(required(f.resources[0]));
    },
    (f: ReturnType<typeof create>) => {
      required(f.resources[0]).startTime = 9;
    },
    (f: ReturnType<typeof create>) => {
      f.resources.push({
        ...required(f.resources[0]),
        name: "http://localhost/api/v1/evidence-records/id/preview-handle",
      });
    },
    () => {
      vi.stubGlobal("innerWidth", 1440);
    },
  ]) {
    const fixture = create();
    mutate(fixture);
    await expect(
      observeEvidenceInspectorPaintInBrowser(fixture.options),
    ).rejects.toThrow();
  }
  const oneFrame = create((frame) => {
    if (frame === 2) required(oneFrame.entries[0]).remove();
  });
  await expect(
    observeEvidenceInspectorPaintInBrowser(oneFrame.options),
  ).rejects.toThrow(/visible completion/u);
  const replaced = create((frame) => {
    if (frame === 2) {
      const element = required(replaced.entries[0]);
      element.replaceWith(element.cloneNode(true));
    }
  });
  await expect(
    observeEvidenceInspectorPaintInBrowser(replaced.options),
  ).rejects.toThrow(/visible completion/u);
});

it("binds private fixture credentials to their exact profile and account cardinality", () => {
  const expected = {
    fixtureProfileId: "evidence_heavy_snapshot_v2",
    snapshotKey: "a".repeat(64),
    credentialSetId: "foreground_analysts",
    accountCount: 1,
  };
  const bundle = {
    schema_id: "cartulary.performance_fixture_runtime.v2",
    fixture_profile_id: expected.fixtureProfileId,
    snapshot_key: expected.snapshotKey,
    credential_sets: [
      {
        set_id: expected.credentialSetId,
        credential_kind: "email_password",
        credentials: [
          {
            principal: "private@example.test",
            secret: "private-password-material-01",
          },
        ],
      },
    ],
  };
  expect(
    parseFixtureRuntimeBundle(JSON.stringify(bundle), expected).accounts,
  ).toHaveLength(1);
  for (const changed of [
    { ...expected, accountCount: 2 },
    { ...expected, credentialSetId: "background_analysts" },
    { ...expected, fixtureProfileId: "timeline_large_grid_snapshot_v2" },
  ])
    expect(() =>
      parseFixtureRuntimeBundle(JSON.stringify(bundle), changed),
    ).toThrow();
});
