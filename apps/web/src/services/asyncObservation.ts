/** Browser observation mechanics. These deadlines never cancel server work. */
export type ObservationClock = {
  readonly now: () => number;
  readonly schedule: (callback: () => void, milliseconds: number) => () => void;
};
export const browserObservationClock: ObservationClock = {
  now: () => performance.now(),
  schedule: (callback, milliseconds) => {
    const timer = setTimeout(callback, milliseconds);
    return () => clearTimeout(timer);
  },
};
export class ObservationStopped extends Error {
  constructor(readonly reason: "aborted" | "deadline") {
    super("Observation stopped; server work may continue.");
    this.name = "ObservationStopped";
  }
}
export async function boundedRead<T>(
  run: (signal: AbortSignal) => Promise<T>,
  signal: AbortSignal,
  timeout = 30_000,
  clock: ObservationClock = browserObservationClock,
): Promise<T> {
  const request = new AbortController();
  let cancelTimer = () => {};
  let stop = () => {};
  try {
    return await new Promise<T>((resolve, reject) => {
      stop = () => {
        request.abort();
        reject(new ObservationStopped("aborted"));
      };
      if (signal.aborted) {
        stop();
        return;
      }
      signal.addEventListener("abort", stop, { once: true });
      cancelTimer = clock.schedule(() => {
        request.abort();
        reject(new ObservationStopped("deadline"));
      }, timeout);
      // Catch synchronous transport rejection as well as asynchronous failure.
      void run(request.signal).then(resolve, reject);
    });
  } finally {
    cancelTimer();
    signal.removeEventListener("abort", stop);
  }
}
export async function observationDelay(
  milliseconds: number,
  signal: AbortSignal,
  clock: ObservationClock,
): Promise<void> {
  await new Promise<void>((resolve) => {
    let cancel = () => {};
    const finish = () => {
      cancel();
      signal.removeEventListener("abort", finish);
      resolve();
    };
    cancel = clock.schedule(finish, milliseconds);
    signal.addEventListener("abort", finish, { once: true });
    if (signal.aborted) finish();
  });
}
