import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowFilter,
  NetworkFlowGraphQueryRequest,
  NetworkFlowRejectedRowsQueryContinuation,
  NetworkFlowRejectedRowsQueryRequest,
  NetworkFlowRow,
  NetworkFlowSort,
  NetworkFlowTable,
  NetworkFlowTableQueryContinuation,
  NetworkFlowTableQueryRequest,
  NetworkFlowTableScope,
} from "../services/networkFlowContractAdapter";
import {
  decodeNetworkFlowFilter,
  decodeNetworkFlowGraphRequest,
  networkFlowQueryMetadata,
} from "../services/networkFlowContractAdapter";
import {
  compareQueryScalars,
  parseQueryScalar,
  parseQueryTimestamp,
} from "./networkFlowQueryValues";

export type GraphQuerySettings = {
  readonly scopeMode: "active_table" | "selected_tables" | "all_active_tables";
  readonly selectedTableIds: readonly string[];
  readonly aggregation: NetworkFlowGraphQueryRequest["aggregation"];
};
export const defaultGraphQuerySettings: GraphQuerySettings = {
  scopeMode: "active_table",
  selectedTableIds: [],
  aggregation: { mode: "default_flow_edge_v1", include_example_row_refs: true },
};
export function graphQueryIdentity(settings: GraphQuerySettings): string {
  return JSON.stringify([
    settings.scopeMode,
    settings.scopeMode === "selected_tables"
      ? [...settings.selectedTableIds].sort()
      : [],
    settings.aggregation,
  ]);
}
export function validateGraphDraft(
  settings: GraphQuerySettings,
  query: NetworkFlowAcceptedQuery,
  activeTableId: string | null,
  tables: readonly NetworkFlowTable[],
): QueryIssue[] {
  const issues: QueryIssue[] = [];
  if (activeTableId === null || tables.length === 0)
    issues.push({
      path: "scope",
      message: "Select an active table before querying.",
    });
  if (settings.scopeMode === "selected_tables") {
    if (settings.selectedTableIds.length === 0)
      issues.push({ path: "scope", message: "Select at least one table." });
    if (
      new Set(settings.selectedTableIds).size !==
      settings.selectedTableIds.length
    )
      issues.push({
        path: "scope",
        message: "Remove duplicate table selections.",
      });
    if (
      settings.selectedTableIds.some(
        (id) => !tables.some((table) => table.network_flow_table_id === id),
      )
    )
      issues.push({
        path: "scope",
        message:
          "A selected table is unavailable. Remove that selection explicitly before applying.",
      });
  }
  if (
    settings.aggregation.mode === "time_bucket_v1" &&
    (!query.timeWindow?.startUTC || !query.timeWindow.endUTC)
  )
    issues.push({
      path: "startUTC",
      message: "Time-bucketed graphs require both UTC range bounds.",
    });
  return issues;
}
export function graphInitialRequest(options: {
  readonly filters: readonly NetworkFlowFilter[];
  readonly tableScope: NetworkFlowTableScope;
  readonly aggregation: NetworkFlowGraphQueryRequest["aggregation"];
  readonly timeRange: NonNullable<
    NetworkFlowGraphQueryRequest["time_range"]
  > | null;
}): NetworkFlowGraphQueryRequest {
  const timestamp = (value: string | null | undefined) => {
    if (value == null) return null;
    const parsed = parseQueryTimestamp(value);
    if (!parsed) throw new Error("network_flow_invalid_graph_time_range");
    return parsed.canonical;
  };
  const start_utc = timestamp(options.timeRange?.start_utc);
  const end_utc = timestamp(options.timeRange?.end_utc);
  const admit = (value: unknown): NetworkFlowGraphQueryRequest => {
    const decoded = decodeNetworkFlowGraphRequest(value);
    if (!decoded.ok) throw new Error("network_flow_invalid_graph_query");
    return decoded.value;
  };
  const common = {
    schema_id: "cartulary.network_flow.graph_query_request.v2" as const,
    table_scope: options.tableScope,
    ...(options.filters.length === 0 ? {} : { filters: [...options.filters] }),
  };
  if (options.aggregation.mode === "time_bucket_v1") {
    if (start_utc === null || end_utc === null)
      throw new Error("network_flow_complete_graph_time_range_required");
    return admit({
      ...common,
      aggregation: options.aggregation,
      time_range: { start_utc, end_utc },
    });
  }
  return admit({
    ...common,
    aggregation: options.aggregation,
    ...(options.timeRange === null
      ? {}
      : { time_range: { start_utc, end_utc } }),
  });
}

export type QueryField = NetworkFlowFilter["field_key"];
export type QueryOperator = NetworkFlowFilter["op"];
export type QueryBasicSlot =
  | "endpoint"
  | "protocol"
  | "bytesMinimum"
  | "packetsMinimum";
export type QueryInput =
  | { readonly kind: "scalar"; readonly text: string }
  | { readonly kind: "list"; readonly members: readonly string[] }
  | { readonly kind: "range"; readonly lower: string; readonly upper: string }
  | { readonly kind: "null" };
export type QueryPredicateDraft = {
  readonly id: string;
  readonly field: QueryField;
  readonly op: QueryOperator;
  readonly input: QueryInput;
  readonly slot: QueryBasicSlot | "advanced";
  readonly original?: NetworkFlowFilter;
};
export type NetworkFlowAcceptedDraft = {
  readonly predicates: readonly QueryPredicateDraft[];
  readonly startUTC: string;
  readonly endUTC: string;
  readonly sort: readonly NetworkFlowSort[];
};
export type NetworkFlowRejectedDraft = {
  readonly errorCodes: readonly string[];
  readonly fieldKeys: readonly string[];
  readonly lower: string;
  readonly upper: string;
};
export type QueryIssue = { readonly path: string; readonly message: string };
export type QueryCompilation<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly issues: readonly QueryIssue[] };

export const queryFieldMetadata = networkFlowQueryMetadata.fields;
export function queryField(field: string) {
  return queryFieldMetadata.find((entry) => entry.field_key === field);
}
export function queryOperator(value: string): QueryOperator | null {
  return (
    queryFieldMetadata
      .flatMap((entry) => [...entry.operators])
      .find((op) => op === value) ?? null
  );
}
export function emptyPredicateInput(op: QueryOperator): QueryInput {
  if (op === "is_null" || op === "not_null") return { kind: "null" };
  if (op === "in") return { kind: "list", members: [""] };
  if (op === "range") return { kind: "range", lower: "", upper: "" };
  return { kind: "scalar", text: "" };
}
function inputFromFilter(filter: NetworkFlowFilter): QueryInput {
  if (filter.op === "is_null" || filter.op === "not_null")
    return { kind: "null" };
  if (filter.op === "in")
    return { kind: "list", members: filter.value.map(String) };
  if (filter.op === "range") {
    const range = filter.value;
    return {
      kind: "range",
      lower: range.gte == null ? "" : String(range.gte),
      upper:
        "lt" in range
          ? String(range.lt ?? "")
          : "lte" in range
            ? String(range.lte ?? "")
            : "",
    };
  }
  return { kind: "scalar", text: String(filter.value) };
}
export function reconstructAcceptedQuery(
  query: NetworkFlowAcceptedQuery,
): NetworkFlowAcceptedDraft {
  const occupied = new Set<QueryBasicSlot>();
  return {
    predicates: query.filters.map((filter, index) => {
      let slot: QueryPredicateDraft["slot"] = "advanced";
      if (
        filter.field_key === "network_flow.endpoint_ip" &&
        (filter.op === "eq" || filter.op === "cidr_contains")
      )
        slot = "endpoint";
      if (filter.field_key === "network_flow.ip_protocol" && filter.op === "eq")
        slot = "protocol";
      if (
        filter.op === "range" &&
        filter.value.gte != null &&
        "lte" in filter.value &&
        filter.value.lte == null
      ) {
        if (filter.field_key === "network_flow.bytes_count")
          slot = "bytesMinimum";
        if (filter.field_key === "network_flow.packets_count")
          slot = "packetsMinimum";
      }
      // Omitted upper bounds have exactly the same meaning as explicit null.
      if (
        filter.op === "range" &&
        filter.value.gte != null &&
        !("lte" in filter.value)
      ) {
        if (filter.field_key === "network_flow.bytes_count")
          slot = "bytesMinimum";
        if (filter.field_key === "network_flow.packets_count")
          slot = "packetsMinimum";
      }
      if (slot !== "advanced") {
        if (occupied.has(slot)) slot = "advanced";
        else occupied.add(slot);
      }
      return {
        id: "predicate-" + index,
        field: filter.field_key,
        op: filter.op,
        input: inputFromFilter(filter),
        slot,
        original: filter,
      };
    }),
    startUTC: query.timeWindow?.startUTC ?? "",
    endUTC: query.timeWindow?.endUTC ?? "",
    sort: query.sort,
  };
}
export function reconstructRejectedQuery(
  query: NetworkFlowRejectedQuery,
): NetworkFlowRejectedDraft {
  return {
    errorCodes: query.errorCodes,
    fieldKeys: query.fieldKeys,
    lower: String(query.sourceRowRange?.gte ?? ""),
    upper: String(query.sourceRowRange?.lte ?? ""),
  };
}
export function compilePredicate(
  draft: QueryPredicateDraft,
): QueryCompilation<NetworkFlowFilter> {
  const metadata = queryField(draft.field);
  const issues: QueryIssue[] = [];
  const fail = (part: string, message: string) =>
    issues.push({ path: draft.id + "." + part, message });
  if (!metadata?.operators.some((op) => op === draft.op))
    return {
      ok: false,
      issues: [
        {
          path: draft.id + ".op",
          message: "This operator is not supported for this field.",
        },
      ],
    };
  const parse = (text: string, part: string) => {
    const result = parseQueryScalar(
      metadata.kind,
      text,
      draft.op === "cidr_contains",
    );
    if (!result.ok) {
      fail(part, result.message);
      return null;
    }
    return result.scalar;
  };
  let value: unknown;
  if (draft.op === "is_null" || draft.op === "not_null") {
    if (draft.input.kind !== "null")
      fail("value", "Null tests must not carry a value.");
  } else if (draft.op === "in" && draft.input.kind === "list") {
    if (
      draft.input.members.length === 0 ||
      draft.input.members.length > networkFlowQueryMetadata.listMaximum
    )
      fail("value", "Use between 1 and 256 list members.");
    const seen = new Set<string>();
    value = draft.input.members.map((member, i) => {
      // Retained empty text remains explicit; a new blank list member is incomplete.
      const retainedEmpty =
        metadata.kind === "text" &&
        draft.original?.op === "in" &&
        draft.original.field_key === draft.field &&
        draft.original.value[i] === "";
      if (member === "" && !retainedEmpty) {
        fail(
          "member-" + i,
          "Enter a value or explicitly remove this list member.",
        );
        return null;
      }
      const scalar = parse(member, "member-" + i);
      if (scalar && seen.has(scalar.canonical))
        fail(
          "member-" + i,
          "This value duplicates another list member after normalization.",
        );
      if (scalar) seen.add(scalar.canonical);
      return scalar?.value;
    });
  } else if (draft.op === "range" && draft.input.kind === "range") {
    const lower =
      draft.input.lower === "" ? null : parse(draft.input.lower, "lower");
    const upper =
      draft.input.upper === "" ? null : parse(draft.input.upper, "upper");
    if (draft.input.lower === "" && draft.input.upper === "")
      fail("lower", "Enter at least one range bound.");
    if (lower && upper) {
      const order = compareQueryScalars(lower, upper);
      if (order > 0 || (metadata.kind === "timestamp" && order === 0))
        fail(
          "upper",
          metadata.kind === "timestamp"
            ? "The exclusive upper bound must be later than the lower bound."
            : "The upper bound must be at least the lower bound.",
        );
    }
    value =
      metadata.kind === "timestamp"
        ? { gte: lower?.value ?? null, lt: upper?.value ?? null }
        : { gte: lower?.value ?? null, lte: upper?.value ?? null };
  } else if (
    draft.input.kind === "scalar" &&
    draft.op !== "in" &&
    draft.op !== "range"
  ) {
    if (
      (draft.op === "prefix" || draft.op === "contains") &&
      draft.input.text === ""
    )
      fail("value", "Enter a nonempty text value.");
    value = parse(draft.input.text, "value")?.value;
  } else fail("value", "Choose the value control for this operator.");
  if (issues.length > 0) return { ok: false, issues };
  const decoded = decodeNetworkFlowFilter({
    field_key: draft.field,
    op: draft.op,
    ...(draft.op === "is_null" || draft.op === "not_null" ? {} : { value }),
  });
  if (!decoded.ok)
    return {
      ok: false,
      issues: [
        {
          path: draft.id + ".value",
          message: "The value does not satisfy this field's query contract.",
        },
      ],
    };
  const original = draft.original;
  return {
    ok: true,
    value:
      original?.field_key === draft.field &&
      original.op === draft.op &&
      JSON.stringify(inputFromFilter(original)) === JSON.stringify(draft.input)
        ? original
        : decoded.value,
  };
}
export function canonicalFilterKey(filter: NetworkFlowFilter): string {
  const metadata = queryField(filter.field_key);
  if (!metadata) throw new Error("Unknown query field.");
  const canonical = (value: string | number) => {
    const result = parseQueryScalar(
      metadata.kind,
      String(value),
      filter.op === "cidr_contains",
    );
    if (!result.ok) throw new Error(result.message);
    return result.scalar;
  };
  if (filter.op === "is_null" || filter.op === "not_null")
    return JSON.stringify([filter.field_key, filter.op]);
  let value: unknown;
  if (filter.op === "in")
    value = filter.value
      .map(canonical)
      .sort(compareQueryScalars)
      .map((v) => v.canonical);
  else if (filter.op === "range") {
    const range = filter.value;
    const upper = "lt" in range ? range.lt : "lte" in range ? range.lte : null;
    value = [
      range.gte == null ? null : canonical(range.gte).canonical,
      upper == null ? null : canonical(upper).canonical,
    ];
  } else value = canonical(filter.value).canonical;
  return JSON.stringify([filter.field_key, filter.op, value]);
}
export function acceptedQueryIdentity(query: NetworkFlowAcceptedQuery): string {
  return JSON.stringify([
    query.filters.map(canonicalFilterKey).sort(),
    query.sort,
    query.timeWindow === null
      ? null
      : [
          query.timeWindow.startUTC === null
            ? null
            : parseQueryTimestamp(query.timeWindow.startUTC)?.canonical,
          query.timeWindow.endUTC === null
            ? null
            : parseQueryTimestamp(query.timeWindow.endUTC)?.canonical,
        ],
  ]);
}
export function compileAcceptedDraft(
  draft: NetworkFlowAcceptedDraft,
  surface: "rows" | "graph" = "rows",
): QueryCompilation<NetworkFlowAcceptedQuery> {
  const issues: QueryIssue[] = [];
  const filters: NetworkFlowFilter[] = [];
  const paths: string[] = [];
  for (const entry of draft.predicates) {
    const result = compilePredicate(entry);
    if (!result.ok) issues.push(...result.issues);
    else {
      filters.push(result.value);
      paths.push(entry.id + ".value");
    }
  }
  const start =
    draft.startUTC === "" ? null : parseQueryTimestamp(draft.startUTC);
  const end = draft.endUTC === "" ? null : parseQueryTimestamp(draft.endUTC);
  if (draft.startUTC !== "" && start === null)
    issues.push({
      path: "startUTC",
      message:
        "Enter a real UTC timestamp ending in Z with at most six fractional digits.",
    });
  if (draft.endUTC !== "" && end === null)
    issues.push({
      path: "endUTC",
      message:
        "Enter a real UTC timestamp ending in Z with at most six fractional digits.",
    });
  if (start && end && start.microseconds >= end.microseconds)
    issues.push({
      path: "endUTC",
      message: "The end must be later than the start.",
    });
  if (issues.length > 0) return { ok: false, issues };
  const query: NetworkFlowAcceptedQuery = {
    filters,
    sort: draft.sort,
    timeWindow:
      draft.startUTC === "" && draft.endUTC === ""
        ? null
        : { startUTC: draft.startUTC || null, endUTC: draft.endUTC || null },
  };
  const wireFilters =
    surface === "rows" ? compileAcceptedFilters(query) : filters;
  if (surface === "rows") {
    if (start) paths.push("startUTC");
    if (end) paths.push("endUTC");
  }
  if (wireFilters.length > networkFlowQueryMetadata.filterMaximum)
    issues.push({
      path: paths[networkFlowQueryMetadata.filterMaximum] ?? "filters",
      message:
        "Use at most " +
        networkFlowQueryMetadata.filterMaximum +
        " predicates, including row time-window bounds.",
    });
  const seen = new Set<string>();
  wireFilters.forEach((filter, index) => {
    const key = canonicalFilterKey(filter);
    if (seen.has(key))
      issues.push({
        path: paths[index] ?? "filters",
        message:
          "This predicate duplicates another filter after normalization. Remove one explicitly.",
      });
    seen.add(key);
  });
  return issues.length ? { ok: false, issues } : { ok: true, value: query };
}
export function compileRejectedDraft(
  draft: NetworkFlowRejectedDraft,
): QueryCompilation<NetworkFlowRejectedQuery> {
  const issues: QueryIssue[] = [];
  const tokenList = <T extends string>(
    values: readonly string[],
    allowed: readonly T[],
    path: string,
  ): T[] => {
    const seen = new Set<string>();
    if (values.length > networkFlowQueryMetadata.diagnosticListMaximum)
      issues.push({ path, message: "Use at most 64 values." });
    return values.flatMap((value, index) => {
      const token = allowed.find((candidate) => candidate === value);
      if (token === undefined)
        issues.push({
          path: path + ".member-" + index,
          message:
            "Select a registered diagnostic token or remove this member.",
        });
      if (seen.has(value))
        issues.push({
          path: path + ".member-" + index,
          message: "Remove the duplicate member.",
        });
      seen.add(value);
      return token === undefined ? [] : [token];
    });
  };
  const errorCodes = tokenList(
    draft.errorCodes,
    networkFlowQueryMetadata.diagnosticErrors,
    "errorCodes",
  );
  const fieldKeys = tokenList(
    draft.fieldKeys,
    networkFlowQueryMetadata.diagnosticFields,
    "fieldKeys",
  );
  const range = (text: string, path: string): number | null => {
    if (text === "") return null;
    const parsed = parseQueryScalar("positive_integer", text);
    if (!parsed.ok) {
      issues.push({ path, message: parsed.message });
      return null;
    }
    return typeof parsed.scalar.value === "number" ? parsed.scalar.value : null;
  };
  const gte = range(draft.lower, "lower");
  const lte = range(draft.upper, "upper");
  if (gte !== null && lte !== null && gte > lte)
    issues.push({
      path: "upper",
      message: "The last source row must be at least the first.",
    });
  return issues.length
    ? { ok: false, issues }
    : {
        ok: true,
        value: {
          errorCodes,
          fieldKeys,
          sourceRowRange: gte === null && lte === null ? null : { gte, lte },
        },
      };
}
export function predicateSummary(filter: NetworkFlowFilter): string {
  const label = filter.field_key
    .replace("network_flow.", "")
    .replaceAll("_", " ");
  if (filter.op === "is_null") return label + " is null";
  if (filter.op === "not_null") return label + " is not null";
  if (filter.op === "range") {
    const range = filter.value;
    const upper = "lt" in range ? range.lt : "lte" in range ? range.lte : null;
    return (
      label +
      " " +
      [
        range.gte == null ? null : "≥ " + range.gte,
        upper == null ? null : ("lt" in range ? "< " : "≤ ") + upper,
      ]
        .filter((v) => v !== null)
        .join(" and ")
    );
  }
  return label + " " + filter.op + " " + JSON.stringify(filter.value);
}

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
  readonly errorCodes: readonly NonNullable<
    NetworkFlowRejectedRowsQueryRequest["error_codes"]
  >[number][];
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
  const range = query.sourceRowRange;
  const sourceRowRange =
    range === null
      ? null
      : range.gte !== null
        ? { gte: range.gte, lte: range.lte }
        : range.lte !== null
          ? { gte: null, lte: range.lte }
          : null;
  if (range !== null && sourceRowRange === null)
    throw new Error("A source row range needs a bound.");
  return {
    schema_id: "cartulary.network_flow.rejected_rows_query_request.v1",
    ...(query.errorCodes.length === 0
      ? {}
      : { error_codes: [...query.errorCodes] }),
    ...(query.fieldKeys.length === 0
      ? {}
      : { field_keys: [...query.fieldKeys] }),
    ...(sourceRowRange === null
      ? {}
      : {
          source_row_range: sourceRowRange,
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

export function reconcileNetworkFlowContributors(
  previous: readonly NetworkFlowContributor[],
  incoming: readonly NetworkFlowContributor[],
): NetworkFlowContributor[] {
  return reconcileByOwnerID(
    previous,
    incoming,
    (contributor) => contributor.row_ref.network_flow_row_id,
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
