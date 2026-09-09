import { afterEach, describe, expect, it, vi } from "vitest";
import {
  observeSavedGraphJob,
  type SavedGraphJobTarget,
  validSavedGraphJob,
} from "./savedGraphObservation";
import {
  deferredSavedGraph,
  savedGraphFixture,
  savedGraphJobFixture,
} from "./savedGraphTestFixtures";

const graph = savedGraphFixture();
const target: SavedGraphJobTarget = {
  incidentId: graph.incident_id,
  graphId: graph.graph_view_id,
  jobId: graph.latest_job_id ?? "",
  statusRoute: `/api/v1/jobs/${graph.latest_job_id}`,
};
afterEach(() => vi.useRealTimers());
describe("Saved graph bounded observation", () => {
  it("observes serially and requires a correctly scoped terminal graph reference", async () => {
    vi.useFakeTimers();
    const read = vi
      .fn()
      .mockResolvedValueOnce(savedGraphJobFixture())
      .mockResolvedValueOnce(savedGraphJobFixture("succeeded"));
    const publish = vi.fn();
    const observation = observeSavedGraphJob({
      target,
      previous: null,
      read,
      publish,
      signal: new AbortController().signal,
    });
    await vi.advanceTimersByTimeAsync(1499);
    expect(read).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(1);
    expect((await observation).state).toBe("terminal");
    expect(publish).toHaveBeenCalledTimes(2);
    expect(
      validSavedGraphJob(savedGraphJobFixture("succeeded"), {
        ...target,
        graphId: "other",
      }),
    ).toBe(false);
    expect(
      validSavedGraphJob(savedGraphJobFixture(), {
        ...target,
        incidentId: "other",
      }),
    ).toBe(false);
    expect(
      validSavedGraphJob(savedGraphJobFixture(), {
        ...target,
        statusRoute: "https://foreign.example",
      }),
    ).toBe(false);
  });
  it("bounds hung reads without turning expiry or abort into server failure", async () => {
    vi.useFakeTimers();
    const pending =
      deferredSavedGraph<ReturnType<typeof savedGraphJobFixture>>();
    const publish = vi.fn();
    const observation = observeSavedGraphJob({
      target,
      previous: null,
      read: () => pending.promise,
      publish,
      signal: new AbortController().signal,
    });
    await vi.advanceTimersByTimeAsync(30_000);
    expect((await observation).state).toBe("paused");
    expect((await observation).job).toBeNull();
    pending.resolve(savedGraphJobFixture("succeeded"));
    await Promise.resolve();
    expect(publish).not.toHaveBeenCalled();
  });
  it("expires the observation window and permits explicit resume from retained progress", async () => {
    vi.useFakeTimers();
    const read = vi.fn(async () => savedGraphJobFixture("running"));
    const observation = observeSavedGraphJob({
      target,
      previous: null,
      read,
      publish: () => {},
      signal: new AbortController().signal,
    });
    await vi.advanceTimersByTimeAsync(120_000);
    const paused = await observation;
    expect(paused.state).toBe("paused");
    expect(paused.job?.status).toBe("running");
    expect(read).toHaveBeenCalledTimes(80);
    const resumed = await observeSavedGraphJob({
      target,
      previous: paused.job,
      read: async () => savedGraphJobFixture("succeeded"),
      publish: () => {},
      signal: new AbortController().signal,
    });
    expect(resumed.state).toBe("terminal");
  });
  it("rejects regressing terminal status and pauses on a missing job without inferring failure", async () => {
    const publish = vi.fn();
    const regression = await observeSavedGraphJob({
      target,
      previous: savedGraphJobFixture("succeeded"),
      read: async () => savedGraphJobFixture("running"),
      publish,
      signal: new AbortController().signal,
    });
    expect(regression.state).toBe("paused");
    expect(publish).not.toHaveBeenCalled();
    const missing = await observeSavedGraphJob({
      target,
      previous: null,
      read: async () => {
        throw new Error("job_not_found");
      },
      publish,
      signal: new AbortController().signal,
    });
    expect(missing.state).toBe("paused");
    expect(missing.job).toBeNull();
  });
});
