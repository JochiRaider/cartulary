import type {
  HistoryPaging,
  RecordHistoryData,
} from "../adapters/workbookHistoryResponse";

export type {
  HistoryPaging,
  RecordHistoryData,
  RecordHistoryItem,
} from "../adapters/workbookHistoryResponse";
export type HistoryPage = RecordHistoryData & {
  readonly paging: HistoryPaging;
};
export type HistoryPageRequest = {
  readonly limit?: number;
  readonly cursorToken?: string;
};
export type HistoryReadScope = {
  readonly epoch: number;
  readonly incidentId: string;
  readonly actorId: string;
  readonly sessionIdentity: string | undefined;
};
export type HistoryPageProvenance = {
  readonly scope: HistoryReadScope;
  readonly recordId: string;
  readonly viewSchemaId: string;
  readonly chainId: number;
  readonly effectiveLimit: number | null;
  readonly generation: number;
  readonly request: HistoryPageRequest;
};
export function sameHistoryReadScope(
  a: HistoryReadScope,
  b: HistoryReadScope | null,
): boolean {
  return (
    b !== null &&
    a.epoch === b.epoch &&
    a.incidentId === b.incidentId &&
    a.actorId === b.actorId &&
    a.sessionIdentity === b.sessionIdentity
  );
}
export function validHistoryPaging(paging: HistoryPaging): boolean {
  return (
    Number.isSafeInteger(paging.limit) &&
    paging.limit >= 1 &&
    paging.limit <= 500 &&
    (paging.has_more
      ? typeof paging.next_cursor === "string" && paging.next_cursor.length > 0
      : paging.next_cursor === null)
  );
}
