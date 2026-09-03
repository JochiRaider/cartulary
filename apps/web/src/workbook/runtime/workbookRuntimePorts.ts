export type WorkbookClockPort = {
  readonly now: () => number;
};

export type WorkbookScheduledCancellation = () => void;

export type WorkbookSchedulerPort = {
  readonly enqueueMicrotask: (task: () => void) => void;
  readonly scheduleDelay: (
    delayMilliseconds: number,
    task: () => void,
  ) => WorkbookScheduledCancellation;
};

export type WorkbookRuntimeDependencies = {
  readonly clock: WorkbookClockPort;
  readonly scheduler: WorkbookSchedulerPort;
};

export const browserWorkbookRuntimeDependencies: WorkbookRuntimeDependencies = {
  clock: { now: () => Date.now() },
  scheduler: {
    enqueueMicrotask: (task) => queueMicrotask(task),
    scheduleDelay: (delayMilliseconds, task) => {
      const timer = setTimeout(task, delayMilliseconds);
      return () => clearTimeout(timer);
    },
  },
};
