import type {
  NetworkFlowDiagnostic,
  NetworkFlowFilter,
  NetworkFlowRejectedRowsQueryContinuation,
  NetworkFlowRejectedRowsQueryRequest,
  NetworkFlowRow,
  NetworkFlowSort,
  NetworkFlowTableQueryContinuation,
  NetworkFlowTableQueryRequest,
} from "../services/networkFlowContractAdapter";

export type NetworkFlowTimeWindow = {
  readonly startUTC: string | null;
  readonly endUTC: string | null;
};

export type NetworkFlowAcceptedQuery = {
  readonly filters: readonly NetworkFlowFilter[];
  readonly sort: readonly NetworkFlowSort[];
  readonly timeWindow: NetworkFlowTimeWindow | null;
};

export type NetworkFlowRejectedQuery = {
  readonly errorCodes: readonly string[];
  readonly fieldKeys: readonly NonNullable<
    NetworkFlowRejectedRowsQueryRequest["field_keys"]
  >[number][];
  readonly sourceRowRange: {
    readonly gte: number | null;
    readonly lte: number | null;
  } | null;
};

export type NetworkFlowAcceptedInitialRequest = Omit<
  NetworkFlowTableQueryRequest,
  "limit"
> & { readonly limit?: never };

export type NetworkFlowRejectedInitialRequest = Omit<
  NetworkFlowRejectedRowsQueryRequest,
  "limit"
> & { readonly limit?: never };

export type NetworkFlowAcceptedPageRequest =
  | NetworkFlowAcceptedInitialRequest
  | NetworkFlowTableQueryContinuation;

export type NetworkFlowRejectedPageRequest =
  | NetworkFlowRejectedInitialRequest
  | NetworkFlowRejectedRowsQueryContinuation;

export const emptyNetworkFlowAcceptedQuery: NetworkFlowAcceptedQuery = {
  filters: [],
  sort: [],
  timeWindow: null,
};

export const emptyNetworkFlowRejectedQuery: NetworkFlowRejectedQuery = {
  errorCodes: [],
  fieldKeys: [],
  sourceRowRange: null,
};

export function acceptedInitialRequest(
  query: NetworkFlowAcceptedQuery,
): NetworkFlowAcceptedInitialRequest {
  const filters = compileAcceptedFilters(query);
  return {
    schema_id: "cartulary.network_flow.table_query_request.v1",
    ...(filters.length === 0 ? {} : { filters }),
    ...(query.sort.length === 0 ? {} : { sort: [...query.sort] }),
  };
}

export function acceptedContinuationRequest(
  cursorToken: string,
): NetworkFlowTableQueryContinuation {
  return {
    schema_id: "cartulary.network_flow.table_query_continuation.v1",
    cursor_token: cursorToken,
  };
}

export function rejectedInitialRequest(
  query: NetworkFlowRejectedQuery,
): NetworkFlowRejectedInitialRequest {
  return {
    schema_id: "cartulary.network_flow.rejected_rows_query_request.v1",
    ...(query.errorCodes.length === 0
      ? {}
      : { error_codes: [...query.errorCodes] }),
    ...(query.fieldKeys.length === 0
      ? {}
      : { field_keys: [...query.fieldKeys] }),
    ...(query.sourceRowRange === null
      ? {}
      : {
          source_row_range: {
            gte: query.sourceRowRange.gte,
            lte: query.sourceRowRange.lte,
          },
        }),
  };
}

export function rejectedContinuationRequest(
  cursorToken: string,
): NetworkFlowRejectedRowsQueryContinuation {
  return {
    schema_id: "cartulary.network_flow.rejected_rows_query_continuation.v1",
    cursor_token: cursorToken,
  };
}

export function compileAcceptedFilters(
  query: NetworkFlowAcceptedQuery,
): NetworkFlowFilter[] {
  const filters = [...query.filters];
  if (query.timeWindow !== null && query.timeWindow.startUTC !== null) {
    filters.push({
      field_key: "network_flow.flow_end_utc",
      op: "range",
      value: { gte: query.timeWindow.startUTC, lt: null },
    });
  }
  if (query.timeWindow !== null && query.timeWindow.endUTC !== null) {
    filters.push({
      field_key: "network_flow.flow_start_utc",
      op: "range",
      value: { gte: null, lt: query.timeWindow.endUTC },
    });
  }
  return filters;
}

export function reconcileNetworkFlowRows(
  previous: readonly NetworkFlowRow[],
  incoming: readonly NetworkFlowRow[],
): NetworkFlowRow[] {
  return reconcileByOwnerID(
    previous,
    incoming,
    (row) => row.network_flow_row_id,
  );
}

export function reconcileNetworkFlowDiagnostics(
  previous: readonly NetworkFlowDiagnostic[],
  incoming: readonly NetworkFlowDiagnostic[],
): NetworkFlowDiagnostic[] {
  return reconcileByOwnerID(
    previous,
    incoming,
    (diagnostic) => diagnostic.diagnostic_id,
  );
}

function reconcileByOwnerID<T>(
  previous: readonly T[],
  incoming: readonly T[],
  ownerID: (value: T) => string,
): T[] {
  const previousByID = new Map(
    previous.map((value) => [ownerID(value), value] as const),
  );
  return incoming.map((value) => {
    const prior = previousByID.get(ownerID(value));
    return prior !== undefined && contractValuesEqual(prior, value)
      ? prior
      : value;
  });
}

function contractValuesEqual(left: unknown, right: unknown): boolean {
  if (Object.is(left, right)) {
    return true;
  }
  if (Array.isArray(left) && Array.isArray(right)) {
    return (
      left.length === right.length &&
      left.every((value, index) => contractValuesEqual(value, right[index]))
    );
  }
  if (
    left === null ||
    right === null ||
    typeof left !== "object" ||
    typeof right !== "object" ||
    Array.isArray(left) ||
    Array.isArray(right)
  ) {
    return false;
  }
  const leftRecord = left as Record<string, unknown>;
  const rightRecord = right as Record<string, unknown>;
  const leftKeys = Object.keys(leftRecord);
  const rightKeys = Object.keys(rightRecord);
  return (
    leftKeys.length === rightKeys.length &&
    leftKeys.every(
      (key) =>
        Object.hasOwn(rightRecord, key) &&
        contractValuesEqual(leftRecord[key], rightRecord[key]),
    )
  );
}
