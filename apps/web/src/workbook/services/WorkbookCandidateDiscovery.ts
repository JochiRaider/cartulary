import { boundedRead } from "../../services/asyncObservation";
import type { WorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type {
  WorkbookCandidate,
  WorkbookCandidatePage,
  WorkbookCandidateReader,
} from "../ports/WorkbookCandidateReadPort";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import type { WorkbookCanonicalQuery } from "../query/WorkbookViewQueryPort";

export const workbookCandidatePageLimit = 100;
export const workbookCandidateCheckpointLimit = 10;
type Checkpoint = Readonly<{
  cursor: string | null;
  pageNumber: number;
  canonicalQuery?: WorkbookCanonicalQuery;
}>;
type Destination = Readonly<{
  checkpoint: Checkpoint;
  direction: "first" | "next" | "previous" | "refresh";
}>;
export type WorkbookCandidateDiscoverySnapshot<T extends WorkbookCandidate> =
  Readonly<{
    scope: string;
    page: WorkbookCandidatePage<T> | null;
    pageNumber: number;
    previousCount: number;
    pending: boolean;
    concealed: boolean;
    unavailable: boolean;
    failure: WorkbookOperationFailure | null;
    failedRead: Destination["direction"] | null;
  }>;

/** One accepted observation; navigation never owns selection, authoring or writes. */
export class WorkbookCandidateDiscovery<T extends WorkbookCandidate> {
  private state: WorkbookCandidateDiscoverySnapshot<T>;
  private readonly listeners = new Set<() => void>();
  private history: readonly Checkpoint[] = [];
  private accepted: Checkpoint | null = null;
  private destination: Destination | null = null;
  private controller: AbortController | null = null;
  private generation = 0;
  private disposed = false;
  constructor(
    private readonly options: {
      readonly scope: string;
      readonly queryState: WorkbookQueryState;
      readonly read: WorkbookCandidateReader<T>;
      readonly isCurrent: () => boolean;
      readonly onAuthorityFailure: (failure: WorkbookOperationFailure) => void;
    },
  ) {
    this.state = {
      scope: options.scope,
      page: null,
      pageNumber: 0,
      previousCount: 0,
      pending: false,
      concealed: false,
      unavailable: false,
      failure: null,
      failedRead: null,
    };
  }
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
  start = () => {
    if (!this.options.isCurrent()) return Promise.resolve();
    this.disposed = false;
    return this.first();
  };
  private publish(patch: Partial<WorkbookCandidateDiscoverySnapshot<T>>) {
    if (this.disposed || !this.options.isCurrent()) return;
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  first = () =>
    this.load({
      checkpoint: { cursor: null, pageNumber: 1 },
      direction: "first",
    });
  next = () => {
    const cursor = this.state.page?.nextCursor;
    if (!this.accepted || !cursor || this.state.unavailable)
      return Promise.resolve();
    return this.load({
      checkpoint: {
        cursor,
        pageNumber: this.accepted.pageNumber + 1,
        ...(this.state.page?.canonicalQuery
          ? { canonicalQuery: this.state.page.canonicalQuery }
          : {}),
      },
      direction: "next",
    });
  };
  previous = () => {
    const checkpoint = this.history.at(-1);
    return checkpoint && !this.state.unavailable
      ? this.load({ checkpoint, direction: "previous" })
      : Promise.resolve();
  };
  refresh = () =>
    this.accepted
      ? this.load({ checkpoint: this.accepted, direction: "refresh" })
      : this.first();
  retry = () =>
    this.destination ? this.load(this.destination) : Promise.resolve();
  private async load(destination: Destination): Promise<void> {
    if (
      this.disposed ||
      this.state.pending ||
      this.state.concealed ||
      !this.options.isCurrent()
    )
      return;
    const controller = new AbortController();
    const generation = ++this.generation;
    this.controller = controller;
    this.destination = destination;
    this.publish({ pending: true, failure: null, failedRead: null });
    let settled = false;
    const current = () =>
      !settled &&
      !this.disposed &&
      !controller.signal.aborted &&
      generation === this.generation &&
      this.options.isCurrent();
    let authorityReported = false;
    try {
      const result = await boundedRead(
        (signal) =>
          this.options.read({
            queryState: this.options.queryState,
            cursor: destination.checkpoint.cursor,
            signal,
            isCurrent: current,
            onAuthorityFailure: (failure) => {
              if (!current() || authorityReported) return;
              authorityReported = true;
              this.options.onAuthorityFailure(failure);
            },
            ...(destination.checkpoint.canonicalQuery
              ? {
                  expectedCanonicalQuery: destination.checkpoint.canonicalQuery,
                }
              : {}),
          }),
        controller.signal,
      );
      if (!current()) return;
      if (result.kind === "aborted") {
        this.publish({ pending: false });
        return;
      }
      if (result.kind === "rejected") {
        this.fail(result.failure, destination, authorityReported);
        return;
      }
      const page = result.value;
      const cursor = page.nextCursor;
      if (
        page.candidates.length > workbookCandidatePageLimit ||
        new Set(page.candidates.map((item) => item.recordId)).size !==
          page.candidates.length ||
        page.candidates.some(
          (item) => !item.recordId || typeof item.displayText !== "string",
        ) ||
        page.hasMore !== (cursor !== null) ||
        (cursor !== null &&
          (!cursor ||
            cursor === destination.checkpoint.cursor ||
            (destination.direction === "next" &&
              this.history.some((item) => item.cursor === cursor))))
      ) {
        this.fail(
          {
            kind: "invalid_contract",
            message: "Candidate paging changed. Restart from First.",
          },
          destination,
        );
        return;
      }
      if (destination.direction === "first") this.history = [];
      else if (destination.direction === "next" && this.accepted)
        this.history = [...this.history, this.accepted].slice(
          -workbookCandidateCheckpointLimit,
        );
      else if (destination.direction === "previous")
        this.history = this.history.slice(0, -1);
      this.accepted = {
        ...destination.checkpoint,
        ...(page.canonicalQuery ? { canonicalQuery: page.canonicalQuery } : {}),
      };
      this.destination = null;
      this.publish({
        page,
        pageNumber: this.accepted.pageNumber,
        previousCount: this.history.length,
        pending: false,
        failure: null,
        failedRead: null,
        unavailable: false,
      });
    } catch {
      if (current())
        this.fail(
          {
            kind: "retryable",
            message: "Candidates could not be loaded. Retry this read.",
          },
          destination,
        );
    } finally {
      settled = true;
      if (this.controller === controller) this.controller = null;
    }
  }
  private fail(
    failure: WorkbookOperationFailure,
    destination: Destination,
    authorityReported = false,
  ) {
    const lifecycle = workbookFailureLifecycle(failure);
    const concealed = lifecycle.kind === "authority_unavailable";
    const unavailable =
      lifecycle.kind === "record_unavailable" ||
      failure.publicCode === "view_schema_not_found";
    if (concealed || unavailable) {
      this.history = [];
      this.accepted = null;
      if (concealed) this.destination = null;
    }
    this.publish({
      failure,
      failedRead: destination.direction,
      pending: false,
      concealed,
      unavailable,
      ...(concealed || unavailable
        ? { page: null, pageNumber: 0, previousCount: 0 }
        : {}),
    });
    if (concealed && !authorityReported)
      this.options.onAuthorityFailure(failure);
  }
  dispose = () => {
    this.disposed = true;
    this.generation++;
    this.controller?.abort();
    this.controller = null;
    this.history = [];
    this.accepted = null;
    this.destination = null;
    this.state = {
      ...this.state,
      page: null,
      pending: false,
      previousCount: 0,
    };
    this.listeners.clear();
  };
}
