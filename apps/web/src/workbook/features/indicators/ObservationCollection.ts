import { boundedRead } from "../../../services/asyncObservation";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";
import type { ObservationPage } from "./observationOperation";

type ObservationCollectionState<T> = Readonly<{
  request: number;
  hasAccepted: boolean;
  items: readonly T[];
  phase: "initial_loading" | "ready" | "loading_more" | "refreshing" | "failed";
  hasMore: boolean;
  nextCursor: string | null;
  failure: WorkbookOperationFailure | null;
  restartRequired: boolean;
}>;
/** Read ownership is separate from both drafts and mutation retention. */
export class ObservationCollection<T> {
  private state: ObservationCollectionState<T> = {
    request: 0,
    hasAccepted: false,
    items: [],
    phase: "initial_loading",
    hasMore: false,
    nextCursor: null,
    failure: null,
    restartRequired: false,
  };
  private readonly listeners = new Set<() => void>();
  private controller: AbortController | null = null;
  private generation = 0;
  private extent = 1;
  private visited = new Set<string>();
  private failed: { cursor: string | null; refresh: boolean } = {
    cursor: null,
    refresh: false,
  };
  constructor(
    private readonly read: (
      cursor: string | null,
      signal: AbortSignal,
    ) => Promise<WorkbookPortResult<ObservationPage<T>>>,
    private readonly identity: (item: T) => string,
    private readonly version: (item: T) => number,
    private readonly ordered?: (a: T, b: T) => boolean,
  ) {}
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.state;
  private publish(patch: Partial<ObservationCollectionState<T>>) {
    this.state = Object.freeze({ ...this.state, ...patch });
    for (const listener of this.listeners) listener();
  }
  dispose() {
    this.generation++;
    this.controller?.abort();
    this.controller = null;
  }
  refresh() {
    this.dispose();
    return this.load(null, true);
  }
  restart() {
    this.extent = 1;
    return this.refresh();
  }
  more() {
    return this.state.hasMore && this.state.phase === "ready"
      ? this.load(this.state.nextCursor)
      : Promise.resolve(false);
  }
  retry() {
    return this.state.restartRequired
      ? this.restart()
      : this.load(this.failed.cursor, this.failed.refresh);
  }
  async load(cursor: string | null = null, refresh = false): Promise<boolean> {
    if (this.controller) return false;
    const controller = new AbortController(),
      generation = ++this.generation;
    this.controller = controller;
    this.failed = { cursor, refresh };
    this.publish({
      request: generation,
      phase: cursor
        ? "loading_more"
        : this.state.hasAccepted
          ? "refreshing"
          : "initial_loading",
      failure: null,
      restartRequired: false,
    });
    const items = cursor ? [...this.state.items] : [],
      indices = new Map(
        items.map((item, index) => [this.identity(item), index]),
      );
    const priorVersions = new Map(
      this.state.items.map((item) => [this.identity(item), this.version(item)]),
    );
    const visited = cursor ? new Set(this.visited) : new Set<string>();
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
              result.failure.kind === "invalid_contract" ||
              result.failure.publicCode === "invalid_pagination_request",
          });
          return false;
        }
        const page = result.value;
        if (
          (page.hasMore &&
            (!page.nextCursor ||
              page.nextCursor === next ||
              visited.has(page.nextCursor))) ||
          (!page.hasMore && page.nextCursor !== null)
        )
          throw new Error("invalid_page");
        let added = 0;
        for (const item of page.items) {
          const id = this.identity(item),
            priorIndex = indices.get(id);
          if (priorIndex !== undefined) {
            const prior = items[priorIndex];
            if (prior && this.version(item) > this.version(prior))
              items[priorIndex] = item;
            continue;
          }
          if (this.version(item) < (priorVersions.get(id) ?? 0))
            throw new Error("invalid_page");
          const previous = items.at(-1);
          if (previous && this.ordered && !this.ordered(previous, item))
            throw new Error("invalid_page");
          indices.set(id, items.length);
          items.push(item);
          added++;
        }
        if (page.hasMore && added === 0) throw new Error("invalid_page");
        if (next) visited.add(next);
        next = page.nextCursor;
        pages++;
        if (!refresh || pages >= this.extent || !page.hasMore) break;
      } while (!controller.signal.aborted);
      if (controller.signal.aborted || generation !== this.generation)
        return false;
      this.extent = cursor ? this.extent + 1 : pages;
      this.visited = visited;
      this.publish({
        hasAccepted: true,
        items,
        phase: "ready",
        hasMore: next !== null,
        nextCursor: next,
      });
      return true;
    } catch (error) {
      if (controller.signal.aborted || generation !== this.generation)
        return false;
      const invalid =
        error instanceof Error && error.message === "invalid_page";
      this.publish({
        phase: "failed",
        restartRequired: invalid,
        failure: {
          kind: invalid ? "invalid_contract" : "retryable",
          message: invalid
            ? "The continuation could not be verified. Restart this collection."
            : "The collection could not be loaded. Retry this read.",
        },
      });
      return false;
    } finally {
      if (this.controller === controller) this.controller = null;
    }
  }
}
