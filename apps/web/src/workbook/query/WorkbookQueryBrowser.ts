import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { savedViewJSONEqual } from "../models/workbookSavedViews";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "./WorkbookQueryRow";
import type {
  WorkbookCanonicalQuery,
  WorkbookProducingQueryRequest,
  WorkbookViewQueryAccepted,
  WorkbookViewQueryPort,
  WorkbookViewQueryResult,
} from "./WorkbookViewQueryPort";
import { canonicalAuthoredQuery } from "./workbookQueryMetadata";

export const workbookBrowsingBounds = Object.freeze({
  pageLimit: 100,
  windowPages: 3,
  checkpoints: 20,
  recoveryAttempts: 2,
});
export type WorkbookBrowseAction =
  | "more"
  | "earlier"
  | "restart"
  | "reconcile"
  | "retry";
type Page = WorkbookViewQueryAccepted;
type Checkpoint = {
  request: WorkbookProducingQueryRequest;
  canonical: WorkbookCanonicalQuery;
};
type Destination = {
  kind: WorkbookBrowseAction | "replace" | "resume";
  start: Checkpoint;
  pageCount: number;
};
type Proposal = {
  generation: number;
  pages: readonly Page[];
  checkpoints: readonly Checkpoint[];
  earlierEvicted: boolean;
  authored: WorkbookQueryState;
  destination: Destination;
};

type WorkbookBrowsingSnapshot = {
  readonly accepted: WorkbookViewQueryAccepted | null;
  readonly authored: WorkbookQueryState | null;
  readonly canonicalQuery: WorkbookCanonicalQuery | null;
  readonly requested: WorkbookQueryState | null;
  readonly pending: Destination["kind"] | null;
  readonly failure: WorkbookOperationFailure | null;
  readonly hasEarlier: boolean;
  readonly canLoadMore: boolean;
  readonly earlierEvicted: boolean;
  readonly pageCount: number;
};

const initialSnapshot = (): WorkbookBrowsingSnapshot => ({
  accepted: null,
  authored: null,
  canonicalQuery: null,
  requested: null,
  pending: null,
  failure: null,
  hasEarlier: false,
  canLoadMore: false,
  earlierEvicted: false,
  pageCount: 0,
});
const contractFailure = (
  message = "Workbook query metadata could not be verified.",
): WorkbookOperationFailure => ({ kind: "invalid_contract", message });

/** One surface's read lifetime. Domain projection acknowledges a staged result explicitly. */
export class WorkbookQueryBrowser {
  private snapshot = initialSnapshot();
  private pages: readonly Page[] = [];
  private checkpoints: readonly Checkpoint[] = [];
  private resume: Checkpoint | null = null;
  private resumeAuthored: WorkbookQueryState | null = null;
  private generation = 0;
  private authorityGeneration = 0;
  private controller: AbortController | null = null;
  private proposal: {
    value: WorkbookViewQueryAccepted;
    plan: Proposal;
  } | null = null;
  private requestedAction: WorkbookBrowseAction | null = null;
  private failedDestination: Destination | null = null;
  private activationPending = false;
  private anchorRecordId: string | null = null;
  private recoveryDepth = 0;
  private chainInvalidated = false;
  private reconciliationEpoch = 0;
  private reconciliation: Promise<void> | null = null;
  private pendingReconciliation: (() => Promise<void>) | null = null;
  recoveryAttemptsUsed() {
    return this.recoveryDepth;
  }
  private listeners = new Set<() => void>();

  constructor(
    private readonly port: WorkbookViewQueryPort,
    readonly viewSchemaId: string,
  ) {}
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<WorkbookBrowsingSnapshot>) {
    this.snapshot = { ...this.snapshot, ...patch };
    for (const listener of this.listeners) listener();
  }
  /** Keep one pending live invalidation while the current captured read finishes. */
  reconcile(read: () => Promise<void>): Promise<void> {
    this.pendingReconciliation = read;
    if (this.reconciliation !== null) return this.reconciliation;
    const epoch = this.reconciliationEpoch;
    const run = async () => {
      while (this.pendingReconciliation && epoch === this.reconciliationEpoch) {
        if (this.snapshot.pending !== null) {
          await new Promise<void>((resolve) => {
            const unsubscribe = this.subscribe(() => {
              if (
                this.snapshot.pending !== null &&
                epoch === this.reconciliationEpoch
              )
                return;
              unsubscribe();
              resolve();
            });
          });
        }
        if (epoch !== this.reconciliationEpoch) return;
        const next = this.pendingReconciliation;
        this.pendingReconciliation = null;
        if (next) await next();
      }
    };
    // Capture the promise before a synchronously completing read can re-enter.
    const pending = Promise.resolve()
      .then(run)
      .finally(() => {
        if (this.reconciliation === pending) this.reconciliation = null;
      });
    this.reconciliation = pending;
    return pending;
  }
  presentationQuery(requested: WorkbookQueryState): WorkbookQueryState {
    return this.snapshot.canonicalQuery && this.snapshot.authored
      ? canonicalAuthoredQuery(
          this.snapshot.authored,
          this.snapshot.canonicalQuery,
        )
      : requested;
  }
  canonicalIntent(requested: WorkbookQueryState): WorkbookQueryState {
    return this.snapshot.canonicalQuery &&
      savedViewJSONEqual(requested, this.snapshot.authored)
      ? canonicalAuthoredQuery(requested, this.snapshot.canonicalQuery)
      : requested;
  }
  hasUnapplied(requested: WorkbookQueryState) {
    return (
      this.snapshot.authored !== null &&
      !savedViewJSONEqual(requested, this.snapshot.authored)
    );
  }
  /** New committed evidence updates admitted members; it cannot insert membership. */
  observeRows(observed: readonly WorkbookQueryRow[]) {
    const accepted = this.snapshot.accepted;
    if (!accepted) return;
    const byId = new Map(observed.map((row) => [row.record_id, row]));
    const newer = (row: WorkbookQueryRow) => {
      const next = byId.get(row.record_id);
      return next && next.row_version > row.row_version ? next : row;
    };
    const rows = accepted.rows.map(newer);
    if (!rows.some((row, index) => row !== accepted.rows[index])) return;
    this.pages = this.pages.map((page) => ({
      ...page,
      rows: page.rows.map(newer),
    }));
    this.publish({ accepted: { ...accepted, rows } });
  }
  /** Guard before React has a chance to paint disabled controls. */
  async activate(action: WorkbookBrowseAction, read: () => Promise<void>) {
    if (this.activationPending || this.snapshot.pending !== null) return;
    if (action === "more" && !this.snapshot.canLoadMore) return;
    if (action === "earlier" && this.checkpoints.length === 0) return;
    this.activationPending = true;
    this.requestedAction = action;
    try {
      await read();
    } finally {
      this.activationPending = false;
      this.requestedAction = null;
    }
  }
  private cancel() {
    this.generation += 1;
    this.controller?.abort();
    this.controller = null;
    this.proposal = null;
  }
  /** Passive rows may leave; authoring and receipts have different owners. */
  rememberAnchor(recordId: string | null) {
    if (recordId !== null) this.anchorRecordId = recordId;
  }
  detach(anchorRecordId: string | null = this.anchorRecordId) {
    this.reconciliationEpoch += 1;
    this.pendingReconciliation = null;
    this.reconciliation = null;
    const page =
      this.pages.find((candidate) =>
        candidate.rows.some((row) => row.record_id === anchorRecordId),
      ) ?? this.pages[0];
    if (page) {
      const preceding = this.pages
        .slice(0, this.pages.indexOf(page))
        .map((candidate) => ({
          request: candidate.producingRequest,
          canonical: candidate.canonicalQuery,
        }));
      const checkpointCount = this.checkpoints.length + preceding.length;
      this.checkpoints = [...this.checkpoints, ...preceding].slice(
        -workbookBrowsingBounds.checkpoints,
      );
      if (checkpointCount > workbookBrowsingBounds.checkpoints)
        this.publish({ earlierEvicted: true });
      this.resume = {
        request: page.producingRequest,
        canonical: page.canonicalQuery,
      };
      this.resumeAuthored = this.snapshot.authored;
    }
    this.cancel();
    this.pages = [];
    this.failedDestination = null;
    this.publish({
      accepted: null,
      pending: null,
      failure: null,
      pageCount: 0,
      hasEarlier: this.checkpoints.length > 0,
      canLoadMore: false,
    });
  }
  invalidate() {
    this.reconciliationEpoch += 1;
    this.pendingReconciliation = null;
    this.reconciliation = null;
    this.cancel();
    this.authorityGeneration += 1;
    this.pages = [];
    this.checkpoints = [];
    this.resume = null;
    this.resumeAuthored = null;
    this.failedDestination = null;
    this.anchorRecordId = null;
    this.chainInvalidated = false;
    this.snapshot = initialSnapshot();
    this.publish({});
  }
  private destination(queryState: WorkbookQueryState): Destination {
    const first: Checkpoint = {
      request: {
        queryState: structuredClone(queryState),
        limit: workbookBrowsingBounds.pageLimit,
      },
      canonical: { filters: [], sort: [] },
    };
    if (this.requestedAction === "restart" || this.chainInvalidated)
      return { kind: "restart", start: first, pageCount: 1 };
    if (
      this.snapshot.authored &&
      !savedViewJSONEqual(queryState, this.snapshot.authored)
    )
      return { kind: "replace", start: first, pageCount: 1 };
    if (this.requestedAction === "retry" && this.failedDestination)
      return this.failedDestination;
    const accepted = this.snapshot.accepted;
    if (this.requestedAction === "more" && accepted?.paging.nextCursor)
      return {
        kind: "more",
        pageCount: 1,
        start: {
          request: {
            queryState: canonicalAuthoredQuery(
              queryState,
              accepted.canonicalQuery,
            ),
            limit: accepted.paging.limit,
            cursorToken: accepted.paging.nextCursor,
          },
          canonical: accepted.canonicalQuery,
        },
      };
    const earlier = this.checkpoints.at(-1);
    if (this.requestedAction === "earlier" && earlier)
      return { kind: "earlier", start: earlier, pageCount: 1 };
    const page = this.pages[0];
    if (page)
      return {
        kind: "reconcile",
        start: {
          request: page.producingRequest,
          canonical: page.canonicalQuery,
        },
        pageCount: this.pages.length,
      };
    if (this.resume && savedViewJSONEqual(queryState, this.resumeAuthored))
      return { kind: "resume", start: this.resume, pageCount: 1 };
    return { kind: "replace", start: first, pageCount: 1 };
  }
  async query(
    input: Parameters<WorkbookViewQueryPort["query"]>[0],
    options: { readonly recoveryDepth?: number } = {},
  ): Promise<WorkbookViewQueryResult> {
    if (input.contract.viewSchemaId !== this.viewSchemaId)
      return { kind: "rejected", failure: contractFailure() };
    this.cancel();
    const generation = this.generation;
    const controller = new AbortController();
    this.controller = controller;
    const abort = () => controller.abort();
    input.signal.addEventListener("abort", abort, { once: true });
    if (input.signal.aborted) controller.abort();
    let destination = this.destination(input.queryState);
    this.recoveryDepth = options.recoveryDepth ?? 0;
    if (this.recoveryDepth > 0 && this.failedDestination)
      destination = this.failedDestination;
    if (destination.kind === "replace" || destination.kind === "restart") {
      this.reconciliationEpoch += 1;
      this.pendingReconciliation = null;
      this.reconciliation = null;
    }
    if (destination.kind === "restart") {
      this.chainInvalidated = true;
      this.checkpoints = [];
      this.resume = null;
      this.resumeAuthored = null;
      this.publish({
        hasEarlier: false,
        canLoadMore: false,
        earlierEvicted: false,
      });
    }
    this.publish({
      pending: destination.kind,
      failure: null,
      requested: structuredClone(input.queryState),
    });
    let restarted = false;
    try {
      for (;;) {
        const result = await this.readPages(
          input.contract,
          destination,
          controller.signal,
          generation,
        );
        if (
          controller.signal.aborted ||
          generation !== this.generation ||
          result.kind === "aborted"
        )
          return { kind: "aborted" };
        if (result.kind === "rejected") {
          if (
            result.failure.kind === "stale_target" &&
            this.recoveryDepth < workbookBrowsingBounds.recoveryAttempts
          ) {
            this.recoveryDepth += 1;
            continue;
          }
          if (
            !restarted &&
            this.recoveryDepth < workbookBrowsingBounds.recoveryAttempts &&
            destination.start.request.cursorToken !== undefined &&
            this.isPaginationFailure(result.failure)
          ) {
            restarted = true;
            this.chainInvalidated = true;
            this.recoveryDepth += 1;
            this.checkpoints = [];
            this.resume = null;
            destination = {
              kind: "restart",
              pageCount: 1,
              start: {
                request: {
                  queryState: structuredClone(input.queryState),
                  limit: workbookBrowsingBounds.pageLimit,
                },
                canonical: { filters: [], sort: [] },
              },
            };
            this.publish({
              pending: "restart",
              hasEarlier: false,
              canLoadMore: false,
              earlierEvicted: false,
            });
            continue;
          }
          this.failedDestination = destination;
          this.publish({ pending: null, failure: result.failure });
          return result;
        }
        const incoming = result.pages;
        let pages =
          destination.kind === "more" ? [...this.pages, ...incoming] : incoming;
        let checkpoints =
          destination.kind === "earlier"
            ? this.checkpoints.slice(0, -1)
            : [...this.checkpoints];
        let earlierEvicted = this.snapshot.earlierEvicted;
        if (destination.kind === "replace" || destination.kind === "restart") {
          checkpoints = [];
          earlierEvicted = false;
        }
        while (pages.length > workbookBrowsingBounds.windowPages) {
          const evicted = pages[0];
          if (!evicted) break;
          pages = pages.slice(1);
          checkpoints.push({
            request: evicted.producingRequest,
            canonical: evicted.canonicalQuery,
          });
        }
        if (checkpoints.length > workbookBrowsingBounds.checkpoints) {
          earlierEvicted = true;
          checkpoints = checkpoints.slice(-workbookBrowsingBounds.checkpoints);
        }
        const first = pages[0],
          last = pages.at(-1);
        if (!first || !last)
          return { kind: "rejected", failure: contractFailure() };
        // A later occurrence is a live placement observation, not a duplicate-record error.
        const rows = new Map<string, WorkbookQueryRow>();
        for (const page of pages)
          for (const row of page.rows) {
            const previous = rows.get(row.record_id);
            rows.delete(row.record_id);
            rows.set(
              row.record_id,
              previous && previous.row_version > row.row_version
                ? previous
                : row,
            );
          }
        const value: WorkbookViewQueryAccepted = {
          ...first,
          rows: [...rows.values()],
          paging: last.paging,
        };
        this.proposal = {
          value,
          plan: {
            generation,
            pages,
            checkpoints,
            earlierEvicted,
            authored: structuredClone(input.queryState),
            destination,
          },
        };
        return { kind: "accepted", value };
      }
    } finally {
      input.signal.removeEventListener("abort", abort);
      if (controller.signal.aborted && generation === this.generation)
        this.publish({ pending: null });
    }
  }
  private async readPages(
    contract: ViewContract,
    destination: Destination,
    signal: AbortSignal,
    generation: number,
  ): Promise<
    | { kind: "accepted"; pages: Page[] }
    | Exclude<WorkbookViewQueryResult, { kind: "accepted" }>
  > {
    let request = destination.start.request;
    let canonical =
      request.cursorToken === undefined
        ? undefined
        : destination.start.canonical;
    const pages: Page[] = [];
    for (let index = 0; index < destination.pageCount; index++) {
      const result = await this.port.query({
        contract,
        queryState: request.queryState,
        limit: request.limit,
        ...(request.cursorToken === undefined
          ? {}
          : { cursorToken: request.cursorToken }),
        ...(canonical === undefined
          ? {}
          : { expectedCanonicalQuery: canonical }),
        identity: {
          authorityGeneration: this.authorityGeneration,
          requestGeneration: generation,
          surfaceIdentity: this.viewSchemaId,
        },
        signal,
      });
      if (signal.aborted || generation !== this.generation)
        return { kind: "aborted" };
      if (result.kind !== "accepted") return result;
      const page = result.value;
      if (
        page.viewSchemaId !== this.viewSchemaId ||
        page.rows.length > workbookBrowsingBounds.pageLimit ||
        !page.canonicalQuery ||
        !page.paging ||
        !page.producingRequest ||
        page.paging.limit !== request.limit ||
        page.producingRequest.limit !== request.limit ||
        page.producingRequest.cursorToken !== request.cursorToken ||
        !savedViewJSONEqual(
          page.producingRequest.queryState,
          request.queryState,
        ) ||
        (page.producingRequest.identity !== undefined &&
          !savedViewJSONEqual(page.producingRequest.identity, {
            authorityGeneration: this.authorityGeneration,
            requestGeneration: generation,
            surfaceIdentity: this.viewSchemaId,
          })) ||
        (this.snapshot.accepted !== null &&
          page.incidentId !== this.snapshot.accepted.incidentId) ||
        (pages[0] !== undefined && page.incidentId !== pages[0].incidentId) ||
        (canonical && !savedViewJSONEqual(page.canonicalQuery, canonical))
      )
        return { kind: "rejected", failure: contractFailure() };
      if (
        page.rows.some(
          (row) =>
            row.row_version <
            (this.snapshot.accepted?.rows.find(
              (accepted) => accepted.record_id === row.record_id,
            )?.row_version ?? 0),
        )
      )
        return {
          kind: "rejected",
          failure: {
            kind: "stale_target",
            message:
              "The query is older than an accepted change. Refresh current rows.",
          },
        };
      pages.push(page);
      if (!page.paging.hasMore) break;
      if (
        pages.some(
          (old) => old.producingRequest.cursorToken === page.paging.nextCursor,
        )
      )
        return {
          kind: "rejected",
          failure: contractFailure(
            "Workbook continuation did not advance. Refresh current rows.",
          ),
        };
      canonical = page.canonicalQuery;
      request = {
        queryState: canonicalAuthoredQuery(request.queryState, canonical),
        limit: page.paging.limit,
        ...(page.paging.nextCursor === null
          ? {}
          : { cursorToken: page.paging.nextCursor }),
      };
    }
    return { kind: "accepted", pages };
  }
  accept(value: WorkbookViewQueryAccepted) {
    const candidate = this.proposal;
    if (
      !candidate ||
      candidate.value !== value ||
      candidate.plan.generation !== this.generation ||
      this.controller?.signal.aborted
    )
      return false;
    this.pages = candidate.plan.pages;
    this.chainInvalidated = false;
    this.checkpoints = candidate.plan.checkpoints;
    this.resume = null;
    this.resumeAuthored = null;
    this.failedDestination = null;
    this.proposal = null;
    this.publish({
      accepted: value,
      canonicalQuery: value.canonicalQuery,
      authored: candidate.plan.authored,
      pageCount: this.pages.length,
      hasEarlier: this.checkpoints.length > 0,
      canLoadMore: value.paging.hasMore,
      earlierEvicted: candidate.plan.earlierEvicted,
      pending: null,
      failure: null,
    });
    return true;
  }
  reject(value: WorkbookViewQueryAccepted, failure: WorkbookOperationFailure) {
    if (this.proposal?.value !== value) return;
    this.failedDestination = this.proposal.plan.destination;
    this.proposal = null;
    this.publish({ pending: null, failure });
  }
  private isPaginationFailure(failure: WorkbookOperationFailure) {
    return (
      failure.publicCode === "invalid_view_query" &&
      [
        "invalid_cursor_token",
        "cursor_query_mismatch",
        "cursor_snapshot_unavailable",
        "invalid_limit",
      ].includes(failure.publicReason ?? "")
    );
  }
}
