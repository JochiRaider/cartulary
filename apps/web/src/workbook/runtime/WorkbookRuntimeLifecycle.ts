import type { WorkbookSchedulerPort } from "./workbookRuntimePorts";

type WorkbookMutationListener = () => void;

export class WorkbookRuntimeLifecycle {
  readonly #scheduler: WorkbookSchedulerPort;
  readonly #listeners = new Set<WorkbookMutationListener>();
  #drainScheduled = false;
  #disposed = false;
  readonly #cleanups = new Set<() => void>();

  constructor(scheduler: WorkbookSchedulerPort) {
    this.#scheduler = scheduler;
  }

  get disposed(): boolean {
    return this.#disposed;
  }

  subscribe(listener: WorkbookMutationListener): () => void {
    if (this.#disposed) return () => {};
    this.#listeners.add(listener);
    return () => this.#listeners.delete(listener);
  }

  emit(): void {
    if (this.#disposed) return;
    for (const listener of this.#listeners) listener();
  }

  requestDrain(drainNext: () => Promise<void>): void {
    if (this.#disposed) return;
    if (this.#drainScheduled) return;
    this.#drainScheduled = true;
    this.#scheduler.enqueueMicrotask(() => {
      this.#drainScheduled = false;
      if (this.#disposed) return;
      void drainNext();
    });
  }

  retainCleanup(cleanup: () => void): void {
    if (this.#disposed) cleanup();
    else this.#cleanups.add(cleanup);
  }

  wait(signal: AbortSignal, milliseconds: number): Promise<void> {
    if (signal.aborted || this.#disposed) return Promise.resolve();
    return new Promise((resolve) => {
      let cancel: () => void = () => {};
      const finish = () => {
        cancel();
        signal.removeEventListener("abort", finish);
        this.#cleanups.delete(finish);
        resolve();
      };
      this.#cleanups.add(finish);
      signal.addEventListener("abort", finish, { once: true });
      cancel = this.#scheduler.scheduleDelay(milliseconds, finish);
    });
  }

  dispose(): boolean {
    if (this.#disposed) return false;
    this.#disposed = true;
    for (const cleanup of this.#cleanups) cleanup();
    this.#cleanups.clear();
    this.#listeners.clear();
    return true;
  }
}
