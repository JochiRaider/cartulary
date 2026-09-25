import type { EntityRecordWriteTarget } from "../mutations/entityRecordWriteBoundary";
import type { WorkbookSchedulerPort } from "./workbookRuntimePorts";

export type WorkbookWriteReadiness<T> =
  | { readonly kind: "waiting" }
  | {
      readonly kind: "completed";
      readonly value: T;
      readonly accept?: () => void;
    };

type Waiter = {
  readonly evaluate: () => void;
  readonly cancel: () => void;
};

/** Retained, presentation-independent lifetime for dependent write readiness. */
export class WorkbookWriteCoordinator {
  readonly #scheduler: Pick<WorkbookSchedulerPort, "enqueueMicrotask">;
  readonly #waiters = new Set<Waiter>();
  readonly #decisionWrites = new Map<symbol, readonly string[]>();
  readonly #entityWrites = new Map<symbol, EntityRecordWriteTarget>();
  #revision = 0;
  #scheduled = false;
  #disposed = false;

  constructor(scheduler: Pick<WorkbookSchedulerPort, "enqueueMicrotask">) {
    this.#scheduler = scheduler;
  }

  get decisionWrites(): ReadonlyMap<symbol, readonly string[]> {
    return this.#decisionWrites;
  }

  get entityWrites(): ReadonlyMap<symbol, EntityRecordWriteTarget> {
    return this.#entityWrites;
  }

  notifyChanged(): void {
    if (this.#disposed) return;
    this.#revision += 1;
    if (this.#scheduled || this.#waiters.size === 0) return;
    this.#scheduled = true;
    this.#scheduler.enqueueMicrotask(() => {
      this.#scheduled = false;
      if (this.#disposed) return;
      for (const waiter of [...this.#waiters]) waiter.evaluate();
    });
  }

  wait<T>(
    signal: AbortSignal,
    cancelled: T,
    read: () => WorkbookWriteReadiness<T>,
  ): Promise<T> {
    if (signal.aborted || this.#disposed) return Promise.resolve(cancelled);
    return new Promise<T>((resolve, reject) => {
      let settled = false;
      const fail = (error: unknown) => {
        if (settled) return;
        settled = true;
        this.#waiters.delete(waiter);
        signal.removeEventListener("abort", waiter.cancel);
        reject(error);
      };
      const finish = (value: T, accept?: () => void) => {
        if (settled) return;
        settled = true;
        this.#waiters.delete(waiter);
        signal.removeEventListener("abort", waiter.cancel);
        try {
          accept?.();
          resolve(value);
        } catch (error) {
          reject(error);
        }
      };
      const waiter: Waiter = {
        cancel: () => finish(cancelled),
        evaluate: () => {
          if (settled) return;
          if (signal.aborted || this.#disposed) {
            finish(cancelled);
            return;
          }
          const revision = this.#revision;
          let readiness: WorkbookWriteReadiness<T>;
          try {
            readiness = read();
          } catch (error) {
            fail(error);
            return;
          }
          if (revision !== this.#revision) {
            this.notifyChanged();
            return;
          }
          if (readiness.kind === "completed")
            finish(readiness.value, readiness.accept);
        },
      };
      this.#waiters.add(waiter);
      signal.addEventListener("abort", waiter.cancel, { once: true });
      if (signal.aborted) waiter.cancel();
      else waiter.evaluate();
    });
  }

  reserveDecision(records: readonly string[]): () => void {
    if (this.#disposed) return () => {};
    const token = Symbol("Decision write");
    this.#decisionWrites.set(token, [...records]);
    this.notifyChanged();
    return () => {
      if (!this.#decisionWrites.delete(token)) return;
      this.notifyChanged();
    };
  }

  reserveEntity(
    target: EntityRecordWriteTarget,
    wakeBatches: () => void,
  ): () => void {
    if (this.#disposed) return () => {};
    const token = Symbol("entity write");
    this.#entityWrites.set(token, target);
    this.notifyChanged();
    return () => {
      if (!this.#entityWrites.delete(token)) return;
      wakeBatches();
      this.notifyChanged();
    };
  }

  dispose(): void {
    if (this.#disposed) return;
    this.#disposed = true;
    for (const waiter of [...this.#waiters]) waiter.cancel();
    this.#decisionWrites.clear();
    this.#entityWrites.clear();
  }
}
