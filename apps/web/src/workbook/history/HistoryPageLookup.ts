import { observeAsyncOperation } from "../../services/asyncObservation";
import type { RecordHistoryItem } from "../adapters/workbookHistoryResponse";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";
import { historyItemContentEqual } from "./workbookHistoryItem";
import {
  type HistoryPage,
  type HistoryPageProvenance,
  type HistoryPageRequest,
  type HistoryReadScope,
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
  message: "The record changed. Review current history before continuing.",
};
/** A disposable read search. It never allocates or sends a mutation. */
export class HistoryPageLookup {
  private request: HistoryPageRequest = {};
  private hint = false;
  private cursors = new Set<string>();
  private identities = new Map<string, RecordHistoryItem>();
  private identityPages: string[][] = [];
  private effectiveLimit: number | null = null;
  private representationGeneration: string | null = null;
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
      readonly evaluate: (page: HistoryPage) => {
        readonly phase: "matched" | "unavailable";
        readonly failure?: WorkbookOperationFailure;
      } | null;
      readonly unavailable: WorkbookOperationFailure;
      readonly maxRetainedPages?: number;
      /** Navigation publishes accepted pages through onPage, without a second payload slot. */
      readonly retainResultPage?: boolean;
      readonly onPage?: (
        page: HistoryPage,
        provenance: HistoryPageProvenance,
      ) => void;
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
    this.identityPages = [];
    this.effectiveLimit = null;
    this.representationGeneration = null;
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
        if (
          this.representationGeneration !== null &&
          this.representationGeneration !== page.representation_generation
        ) {
          this.identities.clear();
          this.value = {
            ...this.value,
            page: null,
            phase: "restart_required",
            failure: {
              kind: "invalid_contract",
              message: "History was updated. Start checking again.",
            },
          };
          return;
        }
        this.representationGeneration = page.representation_generation;
        this.value = {
          ...this.value,
          page: this.options.retainResultPage === false ? null : page,
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
          page.record_id !== this.options.recordId ||
          page.incident_id !== this.options.scope.incidentId ||
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
        this.options.onPage?.(
          page,
          this.value.provenance as HistoryPageProvenance,
        );
        const result = this.options.evaluate(page);
        if (result) {
          this.value = {
            ...this.value,
            phase: result.phase,
            failure: result.failure ?? null,
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
            failure: this.options.unavailable,
          };
          return;
        }
        for (const item of page.items)
          this.identities.set(item.history_item_ref, item);
        this.identityPages.push([...ids]);
        const bound = this.options.maxRetainedPages;
        if (bound !== undefined) {
          while (this.identityPages.length > bound) this.identityPages.shift();
          const retainedIds = new Set(this.identityPages.flat());
          for (const id of this.identities.keys())
            if (!retainedIds.has(id)) this.identities.delete(id);
        }
        if (request.cursorToken) this.cursors.add(request.cursorToken);
        if (bound !== undefined)
          while (this.cursors.size > bound) {
            const first = this.cursors.values().next().value;
            if (first !== undefined) this.cursors.delete(first);
          }
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
