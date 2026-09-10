import { observeAsyncOperation } from "../../services/asyncObservation";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";
import {
  buildRecordRollbackTargetFromHistoryAction,
  historyItemContentEqual,
  historyTargetEqual,
  type WorkbookRecordHistoryPendingAction,
} from "./workbookHistoryItem";
import {
  type HistoryPage,
  type HistoryPageProvenance,
  type HistoryPageRequest,
  type HistoryReadScope,
  type RecordHistoryItem,
  sameHistoryReadScope,
  validHistoryPaging,
} from "./workbookHistoryPage";

export type HistoryLookupState = {
  readonly phase:
    | "idle"
    | "checking"
    | "paused"
    | "failed"
    | "restart_required"
    | "unavailable"
    | "changed"
    | "matched"
    | "cancelled";
  readonly pagesChecked: number;
  readonly page: HistoryPage | null;
  readonly provenance: HistoryPageProvenance | undefined;
  readonly failure: WorkbookOperationFailure | null;
};
const changed: WorkbookOperationFailure = {
  kind: "stale_target",
  message:
    "The record changed. Review current history and confirm the action again.",
};
const unavailable: WorkbookOperationFailure = {
  kind: "stale_target",
  message: "This action is no longer available in current history.",
};

/** A disposable read search. It never allocates or sends a mutation. */
export class HistoryActionLookup {
  private request: HistoryPageRequest = {};
  private hint = false;
  private cursors = new Set<string>();
  private identities = new Map<string, RecordHistoryItem>();
  private effectiveLimit: number | null = null;
  private generation = 0;
  private observation: { cancel: () => void } | null = null;
  private value: HistoryLookupState = {
    phase: "idle",
    pagesChecked: 0,
    page: null,
    provenance: undefined,
    failure: null,
  };
  constructor(
    private readonly options: {
      readonly scope: HistoryReadScope;
      readonly recordId: string;
      readonly viewSchemaId: string;
      readonly pending: WorkbookRecordHistoryPendingAction;
      readonly provenance?: HistoryPageProvenance;
      readonly expectedVersion?: number;
      readonly currentScope: () => HistoryReadScope | null;
      readonly latestVersion: () => number;
      readonly read: (
        request: HistoryPageRequest,
        signal: AbortSignal,
      ) => Promise<WorkbookOperationOutcome<HistoryPage>>;
    },
  ) {
    const origin = options.provenance;
    if (
      origin &&
      origin.recordId === options.recordId &&
      origin.viewSchemaId === options.viewSchemaId &&
      sameHistoryReadScope(origin.scope, options.scope)
    ) {
      this.request = origin.request;
      this.hint = origin.request.cursorToken !== undefined;
    }
  }
  get snapshot() {
    return this.value;
  }
  cancel() {
    this.generation += 1;
    this.observation?.cancel();
    this.value = { ...this.value, phase: "cancelled" };
  }
  interrupted() {
    this.cancel();
    this.value = {
      ...this.value,
      phase: "failed",
      failure: {
        kind: "retryable",
        message: "Checking history timed out. Retry checking to continue.",
      },
    };
  }
  restart() {
    this.cancel();
    this.request = {};
    this.hint = false;
    this.cursors.clear();
    this.identities.clear();
    this.effectiveLimit = null;
    this.value = {
      phase: "idle",
      pagesChecked: 0,
      page: null,
      provenance: undefined,
      failure: null,
    };
  }
  async run(signal?: AbortSignal): Promise<HistoryLookupState> {
    if (
      [
        "checking",
        "matched",
        "unavailable",
        "changed",
        "cancelled",
        "restart_required",
      ].includes(this.value.phase)
    )
      return this.value;
    const generation = ++this.generation;
    this.value = { ...this.value, phase: "checking", failure: null };
    const observation = observeAsyncOperation(async (requestSignal) => {
      for (let count = 0; count < 3; count += 1) {
        if (
          requestSignal.aborted ||
          !sameHistoryReadScope(this.options.scope, this.options.currentScope())
        )
          return;
        const request = this.request;
        const outcome = await this.options.read(request, requestSignal);
        if (
          requestSignal.aborted ||
          generation !== this.generation ||
          !sameHistoryReadScope(this.options.scope, this.options.currentScope())
        )
          return;
        if (outcome.kind === "rejected") {
          this.value = {
            ...this.value,
            phase:
              outcome.failure.publicCode === "invalid_pagination_request" ||
              outcome.failure.kind === "invalid_contract"
                ? "restart_required"
                : "failed",
            failure: outcome.failure,
          };
          return;
        }
        const page = outcome.value;
        this.value = {
          ...this.value,
          page,
          pagesChecked: this.value.pagesChecked + 1,
          provenance: {
            scope: this.options.scope,
            recordId: this.options.recordId,
            viewSchemaId: this.options.viewSchemaId,
            chainId: 0,
            effectiveLimit: page.paging.limit,
            generation,
            request,
          },
        };
        const ids = new Set(page.items.map((item) => item.history_item_ref));
        const next = page.paging.next_cursor;
        if (
          !validHistoryPaging(page.paging) ||
          ids.size !== page.items.length ||
          page.items.some((item) => {
            const previous = this.identities.get(item.history_item_ref);
            return (
              previous !== undefined && !historyItemContentEqual(previous, item)
            );
          }) ||
          (this.effectiveLimit !== null &&
            page.paging.limit !== this.effectiveLimit) ||
          (next !== null &&
            (next === request.cursorToken || this.cursors.has(next)))
        ) {
          this.value = {
            ...this.value,
            phase: "restart_required",
            failure: {
              kind: "invalid_contract",
              message:
                "History continuation could not be verified. Start checking again.",
            },
          };
          return;
        }
        this.effectiveLimit = page.paging.limit;
        if (
          page.row_version < this.options.latestVersion() ||
          (this.options.expectedVersion !== undefined &&
            page.row_version !== this.options.expectedVersion)
        ) {
          this.value = { ...this.value, phase: "changed", failure: changed };
          return;
        }
        const pending = this.options.pending;
        if (pending.kind === "destructive") {
          const legal = page.deleted === (pending.operation === "restore");
          this.value = {
            ...this.value,
            phase: legal ? "matched" : "unavailable",
            failure: legal ? null : unavailable,
          };
          return;
        }
        const item = page.items.find(
          (item) => item.history_item_ref === pending.historyItemRef,
        );
        if (item) {
          const target = buildRecordRollbackTargetFromHistoryAction(
            item,
            pending.action,
          );
          const legal =
            !page.deleted &&
            item.reversible &&
            target !== null &&
            historyTargetEqual(target, pending.target);
          this.value = {
            ...this.value,
            phase: legal ? "matched" : "unavailable",
            failure: legal ? null : unavailable,
          };
          return;
        }
        if (this.hint) {
          // Absence from an old page says nothing about the full live chain.
          this.hint = false;
          this.request = {};
          this.effectiveLimit = null;
          continue;
        }
        if (
          page.items.length > 0 &&
          [...ids].every((id) => this.identities.has(id))
        ) {
          this.value = {
            ...this.value,
            phase: "restart_required",
            failure: {
              kind: "invalid_contract",
              message: "History lookup is not advancing. Start checking again.",
            },
          };
          return;
        }
        if (!page.paging.has_more) {
          this.value = {
            ...this.value,
            phase: "unavailable",
            failure: unavailable,
          };
          return;
        }
        for (const item of page.items)
          this.identities.set(item.history_item_ref, item);
        if (request.cursorToken) this.cursors.add(request.cursorToken);
        this.request = { cursorToken: next as string };
      }
      this.value = { ...this.value, phase: "paused" };
    });
    this.observation = observation;
    const cancel = () => this.cancel();
    signal?.addEventListener("abort", cancel, { once: true });
    if (signal?.aborted) cancel();
    const result = await observation.result;
    signal?.removeEventListener("abort", cancel);
    if (generation === this.generation && this.value.phase === "checking") {
      this.value = {
        ...this.value,
        phase: "failed",
        failure: {
          kind: "retryable",
          message:
            result.kind === "timeout"
              ? "Checking history timed out. Retry checking to continue."
              : "History could not be checked. Retry checking to continue.",
        },
      };
    }
    this.observation = null;
    return this.value;
  }
}
