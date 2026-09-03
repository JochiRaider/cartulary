import type { WorkbookSchedulerPort } from "./workbookRuntimePorts";

type WorkbookMutationListener = () => void;

export class WorkbookRuntimeLifecycle {
  readonly #scheduler: WorkbookSchedulerPort;
  readonly #listeners = new Set<WorkbookMutationListener>();
  #drainScheduled = false;
  #disposed = false;

  constructor(scheduler: WorkbookSchedulerPort) {
    this.#scheduler = scheduler;
  }

  get disposed(): boolean {
    return this.#disposed;
  }

  subscribe(listener: WorkbookMutationListener): () => void {
    this.#listeners.add(listener);
    return () => this.#listeners.delete(listener);
  }

  emit(): void {
    for (const listener of this.#listeners) listener();
  }

  requestDrain(drainNext: () => Promise<void>): void {
    if (this.#disposed) return;
    if (this.#drainScheduled) return;
    this.#drainScheduled = true;
    this.#scheduler.enqueueMicrotask(() => {
      this.#drainScheduled = false;
      void drainNext();
    });
  }

  dispose(): boolean {
    if (this.#disposed) return false;
    this.#disposed = true;
    this.#listeners.clear();
    return true;
  }
}
