import { expect, it, vi } from "vitest";
import { retainAutosaveFeedbackCase } from "./autosaveFeedback";

it("autosave diagnostics retain each failing phase before cleanup and preserve the original error", async () => {
  for (const phase of ["setup", "editor-ready", "response"]) {
    const failure = new Error(phase);
    const events: string[] = [];
    const observation = { loaded: 200, inspectorOpen: true, phase };
    const attach = vi.fn(async (value: typeof observation) => {
      events.push(value.phase);
    });
    await expect(
      retainAutosaveFeedbackCase(
        observation,
        async () => {
          throw failure;
        },
        attach,
        async () => {
          events.push("cleanup");
          throw new Error("secondary cleanup failure");
        },
      ),
    ).rejects.toBe(failure);
    expect(events).toEqual([phase, "cleanup", "cleanup-failed"]);
    expect(attach.mock.calls[0]?.[0]).toBe(observation);
  }
});

it("autosave diagnostics clean up when attachment fails and report cleanup-only failures", async () => {
  const cleanup = vi.fn(async () => {});
  const failure = new Error("attachment");
  await expect(
    retainAutosaveFeedbackCase(
      { phase: "complete" },
      async () => {},
      async () => {
        throw failure;
      },
      cleanup,
    ),
  ).rejects.toBe(failure);
  expect(cleanup).toHaveBeenCalledOnce();
  const phases: string[] = [];
  const cleanupFailure = new Error("cleanup");
  await expect(
    retainAutosaveFeedbackCase(
      { phase: "complete" },
      async () => {},
      async (o) => {
        phases.push(o.phase);
      },
      async () => {
        throw cleanupFailure;
      },
    ),
  ).rejects.toBe(cleanupFailure);
  expect(phases).toEqual(["complete", "cleanup-failed"]);
});
