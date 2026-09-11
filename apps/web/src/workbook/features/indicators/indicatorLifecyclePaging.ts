import { boundedRead } from "../../../services/asyncObservation";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";
import type { LifecyclePage } from "./indicatorLifecycleOperation";

export type LifecyclePagingState<T> = Readonly<{
  items: readonly T[];
  phase: "initial_loading" | "ready" | "loading_more" | "refreshing" | "failed";
  nextCursor: string | null;
  hasMore: boolean;
  failure: WorkbookOperationFailure | null;
  restartRequired: boolean;
}>;

function sameCollectionValue(left: unknown, right: unknown): boolean {
  if (Object.is(left, right)) return true;
  if (Array.isArray(left) || Array.isArray(right))
    return (
      Array.isArray(left) &&
      Array.isArray(right) &&
      left.length === right.length &&
      left.every((value, index) => sameCollectionValue(value, right[index]))
    );
  if (
    left === null ||
    right === null ||
    typeof left !== "object" ||
    typeof right !== "object"
  )
    return false;
  const a = left as Record<string, unknown>,
    b = right as Record<string, unknown>;
  const keys = Object.keys(a);
  return (
    keys.length === Object.keys(b).length &&
    keys.every(
      (key) => Object.hasOwn(b, key) && sameCollectionValue(a[key], b[key]),
    )
  );
}

/** Feature-local collection mechanics; a refresh atomically rebuilds the loaded extent. */
export class IndicatorLifecyclePaging<T> {
  private state: LifecyclePagingState<T> = {
    items: [],
    phase: "initial_loading",
    nextCursor: null,
    hasMore: false,
    failure: null,
    restartRequired: false,
  };
  private request: AbortController | null = null;
  private generation = 0;
  private pageCount = 1;
  private failedCursor: string | null = null;
  private failedRefresh = false;
  private readonly cursors = new Set<string>();
  private readonly listeners = new Set<() => void>();
  constructor(
    private readonly read: (
      cursor: string | null,
      signal: AbortSignal,
    ) => Promise<WorkbookPortResult<LifecyclePage<T>>>,
    private readonly identity: (item: T) => string,
    private readonly inOrder?: (previous: T, next: T) => boolean,
  ) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(change: Partial<LifecyclePagingState<T>>) {
    this.state = Object.freeze({ ...this.state, ...change });
    for (const listener of this.listeners) listener();
  }
  dispose() {
    this.generation++;
    this.request?.abort();
    this.request = null;
  }
  async refresh() {
    this.dispose();
    return this.load(null, true);
  }
  async restart() {
    this.pageCount = 1;
    return this.refresh();
  }
  async more() {
    if (!this.request && this.state.hasMore && !this.state.failure)
      return this.load(this.state.nextCursor, false);
    return false;
  }
  async retry() {
    if (this.state.restartRequired) return this.restart();
    return this.load(this.failedCursor, this.failedRefresh);
  }
  async load(cursor: string | null = null, refresh = false): Promise<boolean> {
    if (this.request) return false;
    const controller = new AbortController(),
      generation = ++this.generation;
    this.request = controller;
    this.failedCursor = cursor;
    this.failedRefresh = refresh;
    this.publish({
      phase: cursor
        ? "loading_more"
        : this.state.items.length
          ? "refreshing"
          : "initial_loading",
      failure: null,
      restartRequired: false,
    });
    const items = cursor === null ? [] : [...this.state.items];
    const seen = new Map(items.map((item) => [this.identity(item), item]));
    const visited = cursor === null ? new Set<string>() : new Set(this.cursors);
    let next = cursor,
      pages = 0;
    try {
      do {
        const result = await boundedRead(
          (signal) => this.read(next, signal),
          controller.signal,
        );
        if (
          controller.signal.aborted ||
          generation !== this.generation ||
          result.kind === "aborted"
        )
          return false;
        if (result.kind === "rejected") {
          this.publish({
            phase: "failed",
            failure: result.failure,
            restartRequired:
              result.failure.publicCode === "invalid_pagination_request" ||
              result.failure.kind === "invalid_contract",
          });
          return false;
        }
        const page = result.value;
        if (
          (page.hasMore &&
            (!page.nextCursor ||
              visited.has(page.nextCursor) ||
              page.nextCursor === next)) ||
          (!page.hasMore && page.nextCursor !== null)
        )
          throw new Error("invalid_page");
        for (const item of page.items) {
          const id = this.identity(item),
            prior = seen.get(id);
          if (prior) {
            if (!sameCollectionValue(prior, item))
              throw new Error("invalid_page");
          } else {
            const previous = items.at(-1);
            if (previous && this.inOrder && !this.inOrder(previous, item))
              throw new Error("invalid_page");
            seen.set(id, item);
            items.push(item);
          }
        }
        if (next) visited.add(next);
        next = page.nextCursor;
        pages++;
        if (!refresh || pages >= this.pageCount || !page.hasMore) break;
      } while (!controller.signal.aborted);
      if (generation !== this.generation || controller.signal.aborted)
        return false;
      if (cursor !== null) this.pageCount++;
      else this.pageCount = pages;
      this.cursors.clear();
      for (const token of visited) this.cursors.add(token);
      this.publish({
        items,
        phase: "ready",
        nextCursor: next,
        hasMore: next !== null,
      });
      return true;
    } catch (error) {
      if (generation !== this.generation || controller.signal.aborted)
        return false;
      const malformed =
        error instanceof Error && error.message === "invalid_page";
      this.publish({
        phase: "failed",
        failure: {
          kind: malformed ? "invalid_contract" : "retryable",
          message: malformed
            ? "The continuation could not be verified. Restart the collection; the displayed results are incomplete."
            : "The read did not finish. Retry to load current results.",
        },
        restartRequired: malformed,
      });
      return false;
    } finally {
      if (this.request === controller) this.request = null;
    }
  }
}
