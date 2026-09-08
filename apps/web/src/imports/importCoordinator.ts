import {
  type ImportFailure,
  type ImportReadResult,
  importContractFailure,
  importInterruptedFailure,
} from "../services/importClient";
import type { ImportJobResource } from "../services/importContractAdapter";
import {
  importJobDoesNotRegress,
  terminalImportJob,
} from "../services/importJobContract";

export type ImportClock = {
  readonly now: () => number;
  readonly schedule: (callback: () => void, milliseconds: number) => () => void;
};
export const browserImportClock: ImportClock = {
  now: () => performance.now(),
  schedule: (callback, milliseconds) => {
    const id = setTimeout(callback, milliseconds);
    return () => clearTimeout(id);
  },
};
export const importTiming = {
  upload: 120_000,
  request: 30_000,
  observation: 120_000,
  interval: 750,
} as const;

export async function boundedImportRead<T>(
  run: (signal: AbortSignal) => Promise<ImportReadResult<T>>,
  signal: AbortSignal,
  timeout: number = importTiming.request,
  clock: ImportClock = browserImportClock,
): Promise<ImportReadResult<T>> {
  const controller = new AbortController();
  let cancelTimer = () => {};
  let stop = () => {};
  try {
    return await new Promise<ImportReadResult<T>>((resolve) => {
      stop = () => {
        controller.abort();
        resolve({ kind: "failed", failure: importInterruptedFailure() });
      };
      if (signal.aborted) {
        stop();
        return;
      }
      signal.addEventListener("abort", stop, { once: true });
      cancelTimer = clock.schedule(stop, timeout);
      void run(controller.signal).then(resolve, () =>
        resolve({ kind: "failed", failure: importInterruptedFailure() }),
      );
    });
  } finally {
    cancelTimer();
    signal.removeEventListener("abort", stop);
  }
}

export type ImportObservationResult =
  | { readonly kind: "terminal"; readonly job: ImportJobResource }
  | {
      readonly kind: "paused";
      readonly job: ImportJobResource;
      readonly failure: ImportFailure;
    };

/** One serial observation window. Stopping it never changes server job status. */
export async function observeImportJob(options: {
  readonly initial: ImportJobResource;
  readonly read: (
    jobId: string,
    signal: AbortSignal,
  ) => Promise<ImportReadResult<ImportJobResource>>;
  readonly signal: AbortSignal;
  readonly onJob: (job: ImportJobResource) => void;
  readonly clock?: ImportClock;
  readonly windowMs?: number;
  readonly requestMs?: number;
  readonly intervalMs?: number;
}): Promise<ImportObservationResult> {
  const clock = options.clock ?? browserImportClock;
  const deadline = clock.now() + (options.windowMs ?? importTiming.observation);
  let job = options.initial;
  // Even terminal replay receipts need a current authorized observation.
  for (;;) {
    const remaining = deadline - clock.now();
    if (options.signal.aborted || remaining <= 0)
      return { kind: "paused", job, failure: importInterruptedFailure() };
    const next = await boundedImportRead(
      (signal) => options.read(job.job_id, signal),
      options.signal,
      Math.min(remaining, options.requestMs ?? importTiming.request),
      clock,
    );
    if (options.signal.aborted)
      return { kind: "paused", job, failure: importInterruptedFailure() };
    if (next.kind === "failed")
      return { kind: "paused", job, failure: next.failure };
    if (!importJobDoesNotRegress(job, next.value))
      return { kind: "paused", job, failure: importContractFailure() };
    job = next.value;
    options.onJob(job);
    if (terminalImportJob(job)) return { kind: "terminal", job };
    await new Promise<void>((resolve) => {
      let cancel = () => {};
      const finish = () => {
        cancel();
        options.signal.removeEventListener("abort", finish);
        resolve();
      };
      cancel = clock.schedule(
        finish,
        Math.min(
          options.intervalMs ?? importTiming.interval,
          Math.max(0, deadline - clock.now()),
        ),
      );
      options.signal.addEventListener("abort", finish, { once: true });
      if (options.signal.aborted) finish();
    });
  }
}

/** Loading is independent of acknowledged submission and never rewrites its receipt. */
export async function loadImportResources(
  client: Pick<
    import("../services/importClient").ImportClient,
    "readSession" | "listUnits"
  >,
  sessionId: string,
  signal: AbortSignal,
  clock?: ImportClock,
) {
  return boundedImportRead(
    async (requestSignal) => {
      const session = await client.readSession(sessionId, requestSignal);
      if (session.kind === "failed") return session;
      const units = await client.listUnits(sessionId, requestSignal);
      if (units.kind === "failed") return units;
      if (
        session.value.selected_unit_ids.some(
          (id) => !units.value.some((unit) => unit.import_unit_id === id),
        )
      )
        return { kind: "failed", failure: importContractFailure() } as const;
      return {
        kind: "received",
        value: { session: session.value, units: units.value },
      } as const;
    },
    signal,
    importTiming.request,
    clock,
  );
}

export async function submitImportAttempt(options: {
  readonly send: (
    signal: AbortSignal,
  ) => Promise<import("../services/importClient").ImportWriteResult>;
  readonly signal: AbortSignal;
  readonly upload?: boolean;
  readonly clock?: ImportClock;
}): Promise<import("../services/importClient").ImportWriteResult> {
  const result = await boundedImportRead(
    async (signal) => ({ kind: "received", value: await options.send(signal) }),
    options.signal,
    options.upload ? importTiming.upload : importTiming.request,
    options.clock,
  );
  return result.kind === "received"
    ? result.value
    : { kind: "uncertain", failure: result.failure };
}
