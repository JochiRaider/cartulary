import { afterEach, describe, expect, it, vi } from "vitest";
import { importTestJob } from "../testing/workbookImportTestSupport";
import { observeImportJob } from "./importCoordinator";

afterEach(() => vi.useRealTimers());
describe("Import job observation", () => {
  it("bounds hung reads without changing the acknowledged job and resumes that job", async () => {
    vi.useFakeTimers();
    const initial = importTestJob();
    const stop = new AbortController();
    const onJob = vi.fn();
    const observed = observeImportJob({
      initial,
      signal: stop.signal,
      onJob,
      read: () => new Promise(() => {}),
      requestMs: 10,
    });
    await vi.advanceTimersByTimeAsync(10);
    expect(await observed).toMatchObject({ kind: "paused", job: initial });
    const read = vi.fn().mockResolvedValue({
      kind: "received",
      value: importTestJob("succeeded"),
    });
    expect(
      await observeImportJob({ initial, signal: stop.signal, onJob, read }),
    ).toMatchObject({ kind: "terminal", job: { status: "succeeded" } });
    expect(read.mock.calls[0]?.[0]).toBe(initial.job_id);
  });
  it("ignores late reads after observation abort and rejects regressions", async () => {
    const stop = new AbortController();
    const onJob = vi.fn();
    let resolve = (_value: unknown) => {};
    const observed = observeImportJob({
      initial: importTestJob(),
      signal: stop.signal,
      onJob,
      read: () =>
        new Promise((done) => {
          resolve = done as typeof resolve;
        }),
    });
    stop.abort();
    await observed;
    resolve({ kind: "received", value: importTestJob("succeeded") });
    await Promise.resolve();
    expect(onJob).not.toHaveBeenCalled();
    const result = await observeImportJob({
      initial: importTestJob("cancel_requested"),
      signal: new AbortController().signal,
      onJob,
      read: async () => ({ kind: "received", value: importTestJob("running") }),
    });
    expect(result.kind).toBe("paused");
    expect(onJob).not.toHaveBeenCalled();
  });
  it("bounds slow jobs by elapsed time and preserves canceled or failed terminal facts", async () => {
    vi.useFakeTimers();
    const onJob = vi.fn();
    const signal = new AbortController().signal;
    const result = observeImportJob({
      initial: importTestJob(),
      signal,
      onJob,
      read: async () => ({ kind: "received", value: importTestJob("running") }),
      intervalMs: 10,
      windowMs: 25,
    });
    await vi.advanceTimersByTimeAsync(25);
    expect(await result).toMatchObject({
      kind: "paused",
      job: { status: "running" },
    });
    for (const status of ["failed", "canceled"] as const)
      expect(
        await observeImportJob({
          initial: importTestJob("running"),
          signal,
          onJob,
          read: async () => ({
            kind: "received",
            value: importTestJob(status),
          }),
        }),
      ).toMatchObject({ kind: "terminal", job: { status } });
  });
});
