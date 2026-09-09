import {
  boundedRead,
  browserObservationClock,
  type ObservationClock,
  observationDelay,
} from "../services/asyncObservation";
import { fetchHTTPOperation } from "../services/browserApi";
import {
  type CommonJobResource,
  commonJobDoesNotRegress,
  terminalCommonJob,
  validCommonJob,
} from "../services/commonJobContract";
import { networkFlowRequestError } from "./networkFlowErrors";

export type SavedGraphJobTarget = {
  readonly incidentId: string;
  readonly graphId: string;
  readonly jobId: string;
  readonly statusRoute: string;
};
export type SavedGraphObservation = {
  readonly target: SavedGraphJobTarget;
  readonly state: "observing" | "paused" | "terminal";
  readonly job: CommonJobResource | null;
  readonly message: string | null;
};
export const savedGraphObservationTiming = {
  interval: 1_500,
  request: 30_000,
  window: 120_000,
  maximumActive: 4,
} as const;
export function validSavedGraphJob(
  job: CommonJobResource,
  target: SavedGraphJobTarget,
): boolean {
  if (
    target.statusRoute !== `/api/v1/jobs/${target.jobId}` ||
    !validCommonJob(job, target.incidentId, target.jobId)
  )
    return false;
  if (job.status !== "succeeded") return true;
  const refs = job.result_summary?.resource_refs ?? [];
  return (
    job.result_summary?.code === "network_flow_graph_view_materialized" &&
    refs.length === 1 &&
    refs[0]?.kind === "network_flow_graph_view" &&
    refs[0].id === target.graphId &&
    refs[0].route ===
      `/api/v1/incidents/${target.incidentId}/network-flow/graph-views/${target.graphId}`
  );
}
export async function readSavedGraphJob(
  apiBase: string | undefined,
  target: SavedGraphJobTarget,
  signal: AbortSignal,
): Promise<CommonJobResource> {
  if (target.statusRoute !== `/api/v1/jobs/${target.jobId}`)
    throw new Error("Invalid job reference.");
  const response = await fetchHTTPOperation<{ data: CommonJobResource }>({
    apiBase,
    operationID: "getJob",
    pathParameters: { job_id: target.jobId },
    init: { method: "GET", signal },
  });
  if (!response.ok)
    throw networkFlowRequestError(
      response.status,
      response.payload,
      "Job status is unavailable. Server work may continue.",
    );
  if (
    response.status !== 200 ||
    !validSavedGraphJob(response.payload.data, target)
  )
    throw new Error("Job status does not match the saved graph.");
  return response.payload.data;
}
/** A serial window with bounded reads; callers own explicit resume and scope fencing. */
export async function observeSavedGraphJob(options: {
  readonly target: SavedGraphJobTarget;
  readonly previous: CommonJobResource | null;
  readonly read: (
    target: SavedGraphJobTarget,
    signal: AbortSignal,
  ) => Promise<CommonJobResource>;
  readonly signal: AbortSignal;
  readonly publish: (job: CommonJobResource) => void;
  readonly clock?: ObservationClock | undefined;
}): Promise<SavedGraphObservation> {
  const clock = options.clock ?? browserObservationClock;
  const end = clock.now() + savedGraphObservationTiming.window;
  let job = options.previous;
  try {
    while (!options.signal.aborted && clock.now() < end) {
      const next = await boundedRead(
        (signal) => options.read(options.target, signal),
        options.signal,
        Math.min(savedGraphObservationTiming.request, end - clock.now()),
        clock,
      );
      if (options.signal.aborted) break;
      if (
        !validSavedGraphJob(next, options.target) ||
        (job !== null && !commonJobDoesNotRegress(job, next))
      )
        throw new Error(
          "The observed job identity or status progression is invalid.",
        );
      job = next;
      options.publish(job);
      if (terminalCommonJob(job))
        return {
          target: options.target,
          state: "terminal",
          job,
          message: null,
        };
      await observationDelay(
        Math.min(savedGraphObservationTiming.interval, end - clock.now()),
        options.signal,
        clock,
      );
    }
    return {
      target: options.target,
      state: "paused",
      job,
      message:
        "Observation stopped. Server work may continue. Resume observation or reload the declaration.",
    };
  } catch {
    return {
      target: options.target,
      state: "paused",
      job,
      message:
        "Job status could not be observed. This does not prove failure or cancellation. Resume observation or reload the declaration.",
    };
  }
}
