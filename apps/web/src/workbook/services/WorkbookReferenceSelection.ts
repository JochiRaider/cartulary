import type { ReferenceFieldContract } from "@cartulary/view-contracts";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import {
  type WorkbookReference,
  type WorkbookReferencePage,
  type WorkbookReferenceReadPort,
  type WorkbookReferenceRequest,
  workbookReferenceKey,
} from "../ports/WorkbookReferenceReadPort";

type Checkpoint = Readonly<{
  request: WorkbookReferenceRequest;
  pageNumber: number;
}>;
type Pending = Readonly<{
  checkpoint: Checkpoint;
  history: readonly Checkpoint[];
  kind: "initial" | "continuation";
}>;
type WorkbookReferenceSelectionSnapshot = Readonly<{
  source: string;
  queryState: WorkbookQueryState;
  page: WorkbookReferencePage | null;
  pageNumber: number;
  previousCount: number;
  selected: readonly WorkbookReference[];
  loading: boolean;
  concealed: boolean;
  failure: Readonly<{
    phase: "initial" | "continuation";
    detail: WorkbookOperationFailure;
  }> | null;
  selectionError: string | null;
}>;

/** Picker lifetime only: one page, ten earlier requests, no parent draft or writes. */
export class WorkbookReferenceSelection {
  readonly #listeners = new Set<() => void>();
  readonly #field: ReferenceFieldContract;
  readonly #reader: WorkbookReferenceReadPort;
  readonly #onAuthorityFailure: (failure: WorkbookOperationFailure) => void;
  readonly #sourceRecordId: string | undefined;
  #snapshot: WorkbookReferenceSelectionSnapshot;
  #history: readonly Checkpoint[] = [];
  #accepted: Checkpoint | null = null;
  #pending: Pending | null = null;
  #controller: AbortController | null = null;
  #generation = 0;
  #disposed = false;

  constructor(options: {
    readonly field: ReferenceFieldContract;
    readonly reader: WorkbookReferenceReadPort;
    readonly selected: readonly WorkbookReference[];
    readonly sourceRecordId?: string | undefined;
    readonly onAuthorityFailure: (failure: WorkbookOperationFailure) => void;
  }) {
    this.#field = options.field;
    this.#reader = options.reader;
    this.#sourceRecordId = options.sourceRecordId;
    this.#onAuthorityFailure = options.onAuthorityFailure;
    const selected = [
      ...new Map(
        options.selected.map((item) => [
          workbookReferenceKey(item),
          {
            ...item,
            presentation:
              item.presentation === "unresolved"
                ? ("unresolved" as const)
                : ("retained" as const),
          },
        ]),
      ).values(),
    ];
    this.#snapshot = {
      source:
        this.sources().find((source) => source === selected[0]?.viewSchemaId) ??
        this.sources()[0] ??
        "",
      queryState: emptyWorkbookQueryState(),
      page: null,
      pageNumber: 0,
      previousCount: 0,
      selected: selected.slice(0, this.limit),
      loading: false,
      concealed: false,
      failure: null,
      selectionError:
        selected.length > this.limit
          ? `Choose at most ${this.limit} references for this edit.`
          : null,
    };
  }
  get limit(): number {
    return this.#field.kind === "direct" ? 1 : 64;
  }
  sources(): readonly string[] {
    return this.#field.identityKind === "incident_member"
      ? ["incident_members"]
      : this.#field.targetViewSchemaIds;
  }
  readonly getSnapshot = () => this.#snapshot;
  readonly subscribe = (listener: () => void) => {
    this.#listeners.add(listener);
    return () => {
      this.#listeners.delete(listener);
    };
  };
  #publish(next: WorkbookReferenceSelectionSnapshot) {
    if (this.#disposed) return;
    this.#snapshot = next;
    for (const listener of this.#listeners) listener();
  }
  replace(
    source: string,
    queryState: WorkbookQueryState = emptyWorkbookQueryState(),
  ): Promise<void> {
    if (
      this.#disposed ||
      this.#snapshot.concealed ||
      !this.sources().includes(source)
    )
      return Promise.resolve();
    this.#history = [];
    this.#accepted = null;
    this.#publish({
      ...this.#snapshot,
      source,
      queryState,
      page: null,
      pageNumber: 0,
      previousCount: 0,
      failure: null,
    });
    return this.#load({
      checkpoint: {
        request: {
          identityKind: this.#field.identityKind,
          viewSchemaId: source,
          queryState,
        },
        pageNumber: 1,
      },
      history: [],
      kind: "initial",
    });
  }
  first(): Promise<void> {
    return this.replace(this.#snapshot.source, this.#snapshot.queryState);
  }
  next(): Promise<void> {
    const { page } = this.#snapshot;
    if (
      this.#snapshot.loading ||
      !page?.paging.hasMore ||
      !page.paging.nextCursor ||
      !this.#accepted
    )
      return Promise.resolve();
    return this.#load({
      checkpoint: {
        request: {
          ...this.#accepted.request,
          cursorToken: page.paging.nextCursor,
          ...(page.canonicalQuery
            ? { expectedCanonicalQuery: page.canonicalQuery }
            : {}),
        },
        pageNumber: this.#accepted.pageNumber + 1,
      },
      history: [...this.#history, this.#accepted].slice(-10),
      kind: "continuation",
    });
  }
  previous(): Promise<void> {
    const checkpoint = this.#history.at(-1);
    if (this.#snapshot.loading || !checkpoint) return Promise.resolve();
    return this.#load({
      checkpoint,
      history: this.#history.slice(0, -1),
      kind: "continuation",
    });
  }
  retry(): Promise<void> {
    return this.#pending && !this.#snapshot.loading
      ? this.#load(this.#pending)
      : Promise.resolve();
  }
  async #load(pending: Pending): Promise<void> {
    if (this.#disposed || this.#snapshot.concealed) return;
    this.#controller?.abort();
    const controller = new AbortController();
    this.#controller = controller;
    const generation = ++this.#generation;
    this.#pending = pending;
    this.#publish({ ...this.#snapshot, loading: true, failure: null });
    const result = await this.#reader.page(
      pending.checkpoint.request,
      controller.signal,
    );
    if (
      this.#disposed ||
      controller.signal.aborted ||
      generation !== this.#generation
    )
      return;
    if (result.kind === "aborted") {
      this.#publish({ ...this.#snapshot, loading: false });
      return;
    }
    if (result.kind === "rejected") {
      const concealed =
        workbookFailureLifecycle(result.failure).kind ===
        "authority_unavailable";
      if (concealed) {
        this.#history = [];
        this.#accepted = null;
        this.#pending = null;
      }
      this.#publish({
        ...this.#snapshot,
        loading: false,
        concealed,
        ...(concealed ? { page: null, selected: [], previousCount: 0 } : {}),
        failure: { phase: pending.kind, detail: result.failure },
      });
      if (concealed) this.#onAuthorityFailure(result.failure);
      return;
    }
    const page = result.value;
    const cursor = page.paging.nextCursor;
    if (
      page.candidates.length > 100 ||
      page.paging.limit !== 100 ||
      page.paging.hasMore !== (cursor !== null) ||
      (cursor !== null &&
        (!cursor ||
          cursor === pending.checkpoint.request.cursorToken ||
          pending.history.some((item) => item.request.cursorToken === cursor)))
    ) {
      this.#publish({
        ...this.#snapshot,
        loading: false,
        failure: {
          phase: pending.kind,
          detail: {
            kind: "invalid_contract",
            message: "Reference continuation is invalid. Restart discovery.",
          },
        },
      });
      return;
    }
    this.#accepted = pending.checkpoint;
    this.#history = pending.history;
    const candidates = page.candidates.filter(
      (item) =>
        !this.#field.excludeSource || item.identity.id !== this.#sourceRecordId,
    );
    const observed = new Map(
      candidates.map((item) => [workbookReferenceKey(item), item]),
    );
    this.#publish({
      ...this.#snapshot,
      loading: false,
      page: { ...page, candidates },
      pageNumber: pending.checkpoint.pageNumber,
      previousCount: this.#history.length,
      failure: null,
      selected: this.#snapshot.selected.map(
        (item) => observed.get(workbookReferenceKey(item)) ?? item,
      ),
    });
  }
  /** Only current-page selection changes; off-page choices retain their order. */
  selectPage(keys: readonly string[]): void {
    if (this.#disposed || this.#snapshot.concealed || !this.#snapshot.page)
      return;
    const candidates = this.#snapshot.page.candidates;
    const pageKeys = new Set(candidates.map(workbookReferenceKey));
    const wanted = new Set(keys);
    const selected =
      this.limit === 1
        ? []
        : this.#snapshot.selected.filter(
            (item) =>
              !pageKeys.has(workbookReferenceKey(item)) ||
              wanted.has(workbookReferenceKey(item)),
          );
    const existing = new Set(selected.map(workbookReferenceKey));
    for (const candidate of candidates) {
      const key = workbookReferenceKey(candidate);
      if (wanted.has(key) && !existing.has(key)) {
        selected.push(candidate);
        existing.add(key);
      }
    }
    if (selected.length > this.limit) {
      this.#publish({
        ...this.#snapshot,
        selectionError: `Choose at most ${this.limit} references for this edit.`,
      });
      return;
    }
    this.#publish({ ...this.#snapshot, selected, selectionError: null });
  }
  remove(key: string): void {
    this.#publish({
      ...this.#snapshot,
      selected: this.#snapshot.selected.filter(
        (item) => workbookReferenceKey(item) !== key,
      ),
      selectionError: null,
    });
  }
  dispose(): void {
    this.#controller?.abort();
    this.#generation += 1;
    this.#history = [];
    this.#accepted = null;
    this.#pending = null;
    this.#publish({
      ...this.#snapshot,
      page: null,
      selected: [],
      previousCount: 0,
      loading: false,
      concealed: true,
    });
    this.#disposed = true;
    this.#listeners.clear();
  }
}
