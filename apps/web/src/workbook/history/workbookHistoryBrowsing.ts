import type { RecordHistoryItem } from "../adapters/workbookHistoryResponse";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import {
  historyItemContentEqual,
  normalizeRecordHistoryData,
} from "./workbookHistoryItem";
import {
  type HistoryPage,
  type HistoryPageProvenance,
  type HistoryPageRequest,
  type HistoryReadScope,
  validHistoryPaging,
} from "./workbookHistoryPage";

export type HistoryReadKind = "initial" | "continuation" | "refresh";
type HistoryBrowseRequest = HistoryPageProvenance & {
  readonly kind: HistoryReadKind;
};
type HistoryAcceptedChain = {
  readonly data: HistoryPage;
  readonly provenance: ReadonlyMap<string, HistoryPageProvenance>;
  readonly pages: readonly HistoryPageProvenance[];
};
export type HistoryBrowsingState = {
  readonly maxRetainedPages?: number;
  readonly scope: HistoryReadScope;
  readonly recordId: string;
  readonly viewSchemaId: string;
  readonly generation: number;
  readonly chainId: number;
  readonly accepted: HistoryAcceptedChain | null;
  readonly pending: HistoryBrowseRequest | null;
  readonly failure: {
    readonly request: HistoryBrowseRequest;
    readonly error: WorkbookOperationFailure;
    readonly restart: boolean;
  } | null;
  readonly chainValid: boolean;
  readonly cursors: readonly string[];
};

export function initialHistoryBrowsing(
  scope: HistoryReadScope,
  recordId: string,
  viewSchemaId: string,
  maxRetainedPages?: number,
): HistoryBrowsingState {
  return {
    ...(maxRetainedPages === undefined ? {} : { maxRetainedPages }),
    scope,
    recordId,
    viewSchemaId,
    generation: 0,
    chainId: 0,
    accepted: null,
    pending: null,
    failure: null,
    chainValid: false,
    cursors: [],
  };
}

export function beginHistoryRead(
  state: HistoryBrowsingState,
  kind: HistoryReadKind,
  retry = false,
): HistoryBrowsingState {
  if (state.pending && (kind !== "refresh" || state.pending.kind === "refresh"))
    return state;
  if (
    kind === "continuation" &&
    (!state.chainValid ||
      !state.accepted?.data.paging.has_more ||
      state.failure?.restart ||
      (state.failure && !retry))
  )
    return state;
  const request: HistoryPageRequest =
    retry && state.failure
      ? state.failure.request.request
      : kind === "continuation"
        ? { cursorToken: state.accepted?.data.paging.next_cursor as string }
        : {};
  const chainId = kind === "continuation" ? state.chainId : state.chainId + 1;
  const pending: HistoryBrowseRequest = {
    scope: state.scope,
    recordId: state.recordId,
    viewSchemaId: state.viewSchemaId,
    chainId,
    effectiveLimit:
      kind === "continuation"
        ? (state.accepted?.data.paging.limit ?? null)
        : null,
    generation: state.generation + 1,
    kind,
    request,
  };
  return {
    ...state,
    chainId,
    generation: pending.generation,
    pending,
    failure: null,
    chainValid: kind === "continuation" && state.chainValid,
    cursors: kind === "continuation" ? state.cursors : [],
  };
}

export function rejectHistoryRead(
  state: HistoryBrowsingState,
  request: HistoryBrowseRequest,
  error: WorkbookOperationFailure,
): HistoryBrowsingState {
  if (state.pending !== request) return state;
  const concealed =
    error.kind === "authorization_lost" ||
    error.kind === "authentication_required" ||
    error.publicCode === "record_not_found" ||
    error.publicCode === "incident_not_found";
  const restart =
    error.publicCode === "invalid_pagination_request" ||
    error.kind === "invalid_contract";
  return {
    ...state,
    pending: null,
    accepted: concealed ? null : state.accepted,
    chainValid: !concealed && !restart && state.chainValid,
    failure: { request, error, restart },
    cursors: concealed ? [] : state.cursors,
  };
}

const invalidPage: WorkbookOperationFailure = {
  kind: "invalid_contract",
  message:
    "History continuation could not be verified. Start fresh history to continue.",
};

export function acceptHistoryPage(
  state: HistoryBrowsingState,
  request: HistoryBrowseRequest,
  page: HistoryPage,
  current: { readonly rowVersion: number; readonly deleted: boolean },
): HistoryBrowsingState {
  if (state.pending !== request) return state;
  if (
    page.record_id !== state.recordId ||
    page.incident_id !== state.scope.incidentId ||
    !validHistoryPaging(page.paging) ||
    normalizeRecordHistoryData(page) === null
  )
    return rejectHistoryRead(state, request, invalidPage);
  const previous = request.kind === "continuation" ? state.accepted : null;
  // A deployment can preserve logical identities while changing their public
  // representation. Retire this disposable read chain before comparing items;
  // authoring, captured requests and receipts belong to separate owners.
  if (
    previous &&
    previous.data.representation_generation !== page.representation_generation
  ) {
    return rejectHistoryRead(
      { ...state, accepted: null, chainValid: false, cursors: [] },
      request,
      {
        kind: "invalid_contract",
        message: "History was updated. Start fresh history to continue.",
      },
    );
  }
  if (previous && previous.data.paging.limit !== page.paging.limit)
    return rejectHistoryRead(state, request, invalidPage);
  const cursor = page.paging.next_cursor;
  if (
    cursor !== null &&
    (cursor === request.request.cursorToken || state.cursors.includes(cursor))
  )
    return rejectHistoryRead(state, request, invalidPage);
  const items = [...(previous?.data.items ?? [])];
  const positions = new Map(
    items.map((item, index) => [item.history_item_ref, index]),
  );
  const provenance = new Map(previous?.provenance);
  const origin: HistoryPageProvenance = {
    ...request,
    effectiveLimit: page.paging.limit,
  };
  let appended = 0;
  for (const item of page.items) {
    const index = positions.get(item.history_item_ref);
    if (index === undefined) {
      positions.set(item.history_item_ref, items.length);
      items.push(item);
      appended += 1;
    } else {
      const old = items[index] as RecordHistoryItem;
      if (!historyItemContentEqual(old, item))
        return rejectHistoryRead(state, request, invalidPage);
      if (page.row_version >= (previous?.data.row_version ?? 0))
        items[index] = item;
    }
    provenance.set(item.history_item_ref, origin);
  }
  if (previous && page.items.length > 0 && appended === 0)
    return rejectHistoryRead(state, request, invalidPage);
  const prior = state.accepted?.data;
  const latest =
    prior && prior.row_version > current.rowVersion
      ? { rowVersion: prior.row_version, deleted: prior.deleted }
      : current;
  const pages = [...(previous?.pages ?? []), origin];
  const retainedPages =
    state.maxRetainedPages === undefined
      ? pages
      : pages.slice(-state.maxRetainedPages);
  if (state.maxRetainedPages !== undefined) {
    const retained = new Set(retainedPages);
    for (const [id, page] of provenance)
      if (!retained.has(page)) provenance.delete(id);
  }
  const data: HistoryPage = {
    ...page,
    items:
      state.maxRetainedPages === undefined
        ? items
        : items.filter((item) => provenance.has(item.history_item_ref)),
    row_version: Math.max(page.row_version, latest.rowVersion),
    deleted:
      page.row_version >= latest.rowVersion ? page.deleted : latest.deleted,
  };
  return {
    ...state,
    pending: null,
    failure: null,
    chainValid: true,
    cursors: (request.request.cursorToken
      ? [...state.cursors, request.request.cursorToken]
      : state.cursors
    ).slice(state.maxRetainedPages === undefined ? 0 : -state.maxRetainedPages),
    accepted: {
      data,
      provenance,
      pages: retainedPages,
    },
  };
}

/** Reconcile authoritative record observations without replacing event provenance. */
export function observeHistoryRecord(
  state: HistoryBrowsingState,
  observation: {
    readonly recordId: string;
    readonly rowVersion: number;
    readonly deleted: boolean;
  },
): HistoryBrowsingState {
  const accepted = state.accepted;
  if (
    state.recordId !== observation.recordId ||
    !accepted ||
    observation.rowVersion <= accepted.data.row_version
  )
    return state;
  return {
    ...state,
    accepted: {
      ...accepted,
      data: {
        ...accepted.data,
        row_version: observation.rowVersion,
        deleted: observation.deleted,
      },
    },
  };
}
