export type WorkbookCollaborationClock = {
  readonly nowMs: () => number;
};

export type WorkbookCollaborationScheduledTask = {
  readonly cancel: () => void;
};

export type WorkbookCollaborationScheduler = {
  readonly schedule: (
    delayMs: number,
    task: () => void,
  ) => WorkbookCollaborationScheduledTask;
};

export const systemWorkbookCollaborationClock: WorkbookCollaborationClock = {
  nowMs: () => Date.now(),
};

export const systemWorkbookCollaborationScheduler: WorkbookCollaborationScheduler =
  {
    schedule(delayMs, task) {
      const handle = setTimeout(task, Math.max(0, delayMs));
      return { cancel: () => clearTimeout(handle) };
    },
  };
