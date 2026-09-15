import { requireViewContract } from "@cartulary/view-contracts";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
  workbookQueryStateFromSavedViewQueryJson,
} from "../workbook/models/workbookQuery";
import type { WorkbookViewQueryAccepted } from "../workbook/query/WorkbookViewQueryPort";

/** Complete first-page metadata for fixtures; malformed-envelope tests supply raw metadata. */
export function acceptedQueryMetadata(
  viewSchemaId: string,
  queryState: WorkbookQueryState = emptyWorkbookQueryState(),
): Pick<
  WorkbookViewQueryAccepted,
  "canonicalQuery" | "paging" | "producingRequest"
> {
  const sort = [...queryState.sort];
  for (const entry of requireViewContract(viewSchemaId).defaultSort) {
    if (!sort.some((candidate) => candidate.fieldKey === entry.fieldKey))
      sort.push(entry);
  }
  return {
    canonicalQuery: {
      filters: queryState.filters,
      sort,
      ...(queryState.groupBy ? { groupBy: queryState.groupBy } : {}),
    },
    paging: { limit: 100, hasMore: false, nextCursor: null },
    producingRequest: { queryState, limit: 100 },
  };
}

export function workbookQueryMeta(
  viewSchemaId: string,
  queryState: WorkbookQueryState = emptyWorkbookQueryState(),
) {
  const { canonicalQuery } = acceptedQueryMetadata(viewSchemaId, queryState);
  return {
    request_id: "req-query",
    paging: { limit: 100, has_more: false, next_cursor: null },
    query: {
      filters: canonicalQuery.filters.map((filter) => ({
        field_key: filter.fieldKey,
        op: filter.op,
        arg: filter.arg,
      })),
      sort: canonicalQuery.sort.map((sort) => ({
        field_key: sort.fieldKey,
        direction: sort.direction,
      })),
      ...(canonicalQuery.groupBy ? { group_by: canonicalQuery.groupBy } : {}),
    },
  };
}

/** Transport for interaction fixtures that author rows separately from query controls.
 * Captures the producing body before a deferred reply; raw adapter-contract tests
 * deliberately use their own transport instead.
 */
export function withWorkbookQueryFixtureMetadata(fetchFixture: unknown) {
  if (typeof fetchFixture !== "function")
    throw new Error("Expected callable interaction fixture");
  return async (
    input: RequestInfo | URL,
    init?: RequestInit,
  ): Promise<Response> => {
    const isQuery = /\/views\/[^/]+\/query(?:\?.*)?$/.test(String(input));
    const body = isQuery ? JSON.parse(String(init?.body ?? "{}")) : null;
    const response = await fetchFixture(input, init);
    if (!(response instanceof Response))
      throw new Error("Interaction fixture must return a Response");
    if (!isQuery || !response.ok) return response;
    const envelope = await response.clone().json();
    const schema = envelope.data?.view_schema_id;
    if (typeof schema !== "string" || !Array.isArray(envelope.data?.rows))
      return response;
    return new Response(
      JSON.stringify({
        ...envelope,
        meta: workbookQueryMeta(
          schema,
          workbookQueryStateFromSavedViewQueryJson(
            requireViewContract(schema),
            body,
          ),
        ),
      }),
      { status: response.status, headers: response.headers },
    );
  };
}
