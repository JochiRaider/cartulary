// @vitest-environment jsdom
import { afterEach, expect, it, vi } from "vitest";
import {
  type JobProgressPaintInput,
  observeJobProgressPaintInBrowser,
} from "./jobProgressPaint";

const elementFromPoint = Object.getOwnPropertyDescriptor(
  document,
  "elementFromPoint",
);
const baseInput: JobProgressPaintInput = {
  jobPurpose: "discovery",
  recordId: "record",
  rowVersion: 3,
  fieldKey: "timeline.analyst_text",
  draftValue: "priorx",
  gridSelector: "#grid",
  editorSelector: "#editor",
  progressSurfaceSelector: "#assistant",
  startMark: "start",
  admissionMark: "admission",
  dispatchMark: "dispatch",
  timeoutMs: 35,
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
function fixture(purpose: JobProgressPaintInput["jobPurpose"] = "discovery") {
  document.body.innerHTML =
    '<div id="grid"><div data-grid-record-id="record" data-grid-row-version="3" style="display:contents;overflow:hidden"><div data-grid-field-key="timeline.analyst_text"><input id="editor" value="priorx"/></div></div></div><div id="assistant"><section data-job-id="job" data-job-status="running" data-job-purpose="discovery" data-job-cancelable="true"><p data-job-progress="true">Progress: 2/unknown</p><button data-job-cancel="true">Cancel import</button></section></div>';
  for (const node of document.querySelectorAll<HTMLElement>("*"))
    node.getBoundingClientRect = () => box();
  const editor = document.querySelector<HTMLInputElement>("#editor");
  const job = document.querySelector<HTMLElement>("[data-job-id]");
  const row = document.querySelector<HTMLElement>("[data-grid-record-id]");
  const cancel = document.querySelector<HTMLButtonElement>("button");
  const progress = document.querySelector<HTMLElement>("p");
  if (!editor || !job || !row || !cancel || !progress)
    throw new Error("Import DOM fixture is missing");
  editor.getBoundingClientRect = () => box(20);
  progress.getBoundingClientRect = () => box(100);
  cancel.getBoundingClientRect = () => box(200);
  Object.defineProperty(document, "elementFromPoint", {
    configurable: true,
    value: (x: number) => {
      for (const node of [editor, progress, cancel]) {
        const rect = node.getBoundingClientRect();
        if (x >= rect.left && x <= rect.right) return node;
      }
      return document.body;
    },
  });
  job.dataset.jobPurpose = purpose;
  job.dataset.jobBlobId = "blob";
  editor.focus();
  return { editor, job, row, cancel, progress };
}
function clock(onFrame: (frame: number) => void = () => {}) {
  let frame = 0;
  const purpose =
    document.querySelector<HTMLElement>("[data-job-id]")?.dataset.jobPurpose;
  vi.stubGlobal("innerWidth", 1920);
  vi.stubGlobal("innerHeight", 1080);
  vi.stubGlobal("devicePixelRatio", 1);
  const marks: Record<
    string,
    { startTime: number; detail: Record<string, unknown> }[]
  > = {
    start: [
      {
        startTime: 1,
        detail: { transactionId: "tx", blobId: "blob", purpose },
      },
    ],
    dispatch: [
      {
        startTime: 2,
        detail: { transactionId: "tx", replay: false, blobId: "blob", purpose },
      },
    ],
    admission: [
      {
        startTime: 3,
        detail: { transactionId: "tx", jobId: "job", blobId: "blob", purpose },
      },
    ],
  };
  vi.stubGlobal("performance", {
    timeOrigin: 1000,
    now: () => frame * 5,
    getEntriesByName: (name: string) => marks[name] ?? [],
  });
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    queueMicrotask(() => {
      frame++;
      onFrame(frame);
      callback(frame * 5);
    });
    return frame;
  });
  return marks;
}
afterEach(() => {
  vi.unstubAllGlobals();
  if (elementFromPoint)
    Object.defineProperty(document, "elementFromPoint", elementFromPoint);
  else Reflect.deleteProperty(document, "elementFromPoint");
  document.body.innerHTML = "";
});

async function assertProgress(purpose: JobProgressPaintInput["jobPurpose"]) {
  const input = { ...baseInput, jobPurpose: purpose };
  fixture(purpose);
  clock();
  await expect(observeJobProgressPaintInBrowser(input)).resolves.toEqual({
    elapsed: 9,
    jobId: "job",
    jobStatus: "running",
    ...(purpose === "evidence_verification" ? { blobId: "blob" } : {}),
  });
  for (const mutate of [
    () => {
      Object.defineProperty(document, "elementFromPoint", {
        configurable: true,
        value: () => document.body,
      });
    },
    (f: ReturnType<typeof fixture>) => {
      f.job.dataset.jobStatus = "succeeded";
    },
    (f: ReturnType<typeof fixture>) => {
      f.job.dataset.jobId = "other";
    },
    (f: ReturnType<typeof fixture>) => {
      f.job.dataset.jobPurpose =
        purpose === "discovery" ? "apply" : "discovery";
    },
    (f: ReturnType<typeof fixture>) => {
      f.job.dataset.jobCancelable = "false";
    },
    (f: ReturnType<typeof fixture>) => {
      f.cancel.disabled = true;
    },
    (f: ReturnType<typeof fixture>) => {
      f.cancel.hidden = true;
    },
    (f: ReturnType<typeof fixture>) => {
      f.progress.textContent = "";
    },
    (f: ReturnType<typeof fixture>) => {
      f.progress.style.opacity = "0";
    },
    (f: ReturnType<typeof fixture>) => {
      f.editor.getBoundingClientRect = () => box(1910);
    },
    (f: ReturnType<typeof fixture>) => {
      f.editor.value = "prior";
    },
    (f: ReturnType<typeof fixture>) => {
      f.editor.readOnly = true;
    },
    (f: ReturnType<typeof fixture>) => {
      f.cancel.setAttribute("aria-disabled", "true");
    },
    (f: ReturnType<typeof fixture>) => {
      f.editor.blur();
    },
    (f: ReturnType<typeof fixture>) => {
      f.row.dataset.gridRowVersion = "4";
    },
    (f: ReturnType<typeof fixture>) => {
      f.row.dataset.gridRecordId = "other";
    },
    (f: ReturnType<typeof fixture>) => {
      f.editor.parentElement?.setAttribute("data-grid-field-key", "other");
    },
    (f: ReturnType<typeof fixture>) => {
      f.row.setAttribute("inert", "");
    },
    (f: ReturnType<typeof fixture>) => {
      f.job.parentElement?.append(f.job.cloneNode(true));
    },
  ]) {
    const f = fixture(purpose);
    clock();
    mutate(f);
    await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow();
  }
  for (const name of ["start", "dispatch", "admission"]) {
    fixture(purpose);
    const missing = clock();
    missing[name] = [];
    await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow();
    fixture(purpose);
    const repeated = clock();
    repeated[name] = [...(repeated[name] ?? []), ...(repeated[name] ?? [])];
    await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
      /repeated/u,
    );
  }
  for (const mutate of [
    (m: ReturnType<typeof clock>) => {
      if (m.admission?.[0]) m.admission[0].detail.transactionId = "other";
    },
    (m: ReturnType<typeof clock>) => {
      if (m.dispatch?.[0]) m.dispatch[0].detail.replay = true;
    },
    (m: ReturnType<typeof clock>) => {
      if (m.admission?.[0]) m.admission[0].startTime = 0;
    },
    (m: ReturnType<typeof clock>) => {
      if (m.dispatch?.[0]) m.dispatch[0].startTime = 4;
    },
  ]) {
    fixture(purpose);
    const marks = clock();
    mutate(marks);
    await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow();
  }
  const f = fixture(purpose);
  clock((frame) => {
    if (frame === 2) f.job.dataset.jobStatus = "succeeded";
  });
  await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
    /two visible frames/u,
  );
  fixture(purpose);
  clock((frame) => {
    if (frame === 2) vi.stubGlobal("innerWidth", 1919);
  });
  await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
    /geometry/u,
  );
}

it("requires one fresh discovery job with visible cancel and a stable editable Timeline binding", async () => {
  await assertProgress("discovery");
});
it("requires one fresh apply job with visible cancel and a stable editable Timeline binding", async () => {
  await assertProgress("apply");
});

it("requires fresh snapshot and report jobs with visible cancellation and stable grid editing", async () => {
  await assertProgress("snapshot");
  await assertProgress("report");
});

it("requires one fresh Evidence verification job with exact blob binding and editable grid", async () => {
  await assertProgress("evidence_verification");
  const input = { ...baseInput, jobPurpose: "evidence_verification" as const };
  for (const mark of ["start", "dispatch", "admission"]) {
    fixture("evidence_verification");
    const marks = clock();
    const entry = marks[mark]?.[0];
    if (!entry) throw new Error("Missing fixture mark");
    entry.detail.blobId = "other";
    await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
      /blob binding/u,
    );
  }
  const f = fixture("evidence_verification");
  clock();
  f.job.dataset.jobBlobId = "other";
  await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
    /two visible frames/u,
  );
});

it("requires distinct Reference Pack purposes and rejects altered action bindings", async () => {
  for (const purpose of [
    "reference_pack_import",
    "reference_pack_reverify",
    "reference_pack_refresh",
  ] as const) {
    await assertProgress(purpose);
    for (const name of ["start", "dispatch", "admission"]) {
      fixture(purpose);
      const marks = clock();
      const mark = marks[name]?.[0];
      if (!mark) throw new Error("Missing fixture mark");
      mark.detail.purpose = "other";
      await expect(
        observeJobProgressPaintInBrowser({ ...baseInput, jobPurpose: purpose }),
      ).rejects.toThrow(/purpose/u);
    }
  }
});

it("requires two companion frames and rejects missing changed or stalled Timeline documents", async () => {
  const input = {
    ...baseInput,
    jobPurpose: "reference_pack_import" as const,
    gridLocation: "opener" as const,
  };
  const prepare = (onFrame: (frame: number) => void = () => {}) => {
    const f = fixture("reference_pack_import");
    clock();
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const companion = frame.contentWindow;
    if (!companion) throw new Error("Missing companion fixture");
    for (const [key, value] of Object.entries({
      innerWidth: 1920,
      innerHeight: 1080,
      devicePixelRatio: 1,
    }))
      Object.defineProperty(companion, key, { configurable: true, value });
    const grid = document.querySelector("#grid");
    if (!grid) throw new Error("Missing grid");
    companion.document.body.append(grid);
    for (const node of [
      companion.document.body,
      companion.document.documentElement,
    ])
      node.getBoundingClientRect = () => box();
    Object.defineProperty(companion.document, "elementFromPoint", {
      value: () => f.editor,
      configurable: true,
    });
    let frames = 0;
    companion.requestAnimationFrame = (callback) => {
      queueMicrotask(() => {
        frames++;
        onFrame(frames);
        callback(performance.now());
      });
      return frames + 1;
    };
    vi.stubGlobal("opener", companion);
    f.editor.focus();
    return { ...f, companion, frames: () => frames };
  };
  const valid = prepare();
  const historical = valid.job.cloneNode(true) as HTMLElement;
  historical.dataset.jobId = "preparation-job";
  historical.dataset.jobStatus = "succeeded";
  valid.job.parentElement?.append(historical);
  await expect(observeJobProgressPaintInBrowser(input)).resolves.toEqual({
    elapsed: 9,
    jobId: "job",
    jobStatus: "running",
  });
  expect(valid.frames()).toBe(2);
  prepare();
  vi.stubGlobal("opener", null);
  await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
    /companion is missing/u,
  );
  prepare();
  vi.stubGlobal("opener", {
    get document() {
      throw new Error("cross origin");
    },
  });
  await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
    /same-origin/u,
  );
  prepare((frame) => {
    if (frame === 2) vi.stubGlobal("opener", null);
  });
  await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
    /geometry/u,
  );
  const changed = prepare((frame) => {
    if (frame === 2) changed.editor.value = "newer";
  });
  await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
    /two visible frames/u,
  );
  const stalled = prepare();
  stalled.companion.requestAnimationFrame = () => 0;
  await expect(observeJobProgressPaintInBrowser(input)).rejects.toThrow(
    /Companion animation frames/u,
  );
});
