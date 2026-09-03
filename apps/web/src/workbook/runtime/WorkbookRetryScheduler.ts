import type { WorkbookSchedulerPort } from "./workbookRuntimePorts";

export class WorkbookRetryScheduler {
  readonly #scheduler: WorkbookSchedulerPort;
  #cancelPending: (() => void) | null = null;

  constructor(scheduler: WorkbookSchedulerPort) {
    this.#scheduler = scheduler;
  }

  get pending(): boolean {
    return this.#cancelPending !== null;
  }

  schedule(delayMilliseconds: number, task: () => void): boolean {
    if (this.#cancelPending !== null) return false;
    this.#cancelPending = this.#scheduler.scheduleDelay(
      delayMilliseconds,
      () => {
        this.#cancelPending = null;
        task();
      },
    );
    return true;
  }

  cancel(): void {
    this.#cancelPending?.();
    this.#cancelPending = null;
  }
}
