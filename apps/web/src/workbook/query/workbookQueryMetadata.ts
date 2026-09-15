import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { savedViewJSONEqual } from "../models/workbookSavedViews";
import type {
  WorkbookCanonicalQuery,
  WorkbookQueryPaging,
} from "./WorkbookViewQueryPort";

const object = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === "object" && !Array.isArray(value);

function validArgument(op: string, arg: Record<string, unknown>): boolean {
  const keys = Object.keys(arg);
  const scalar = (value: unknown) =>
    typeof value === "string" ||
    typeof value === "boolean" ||
    (typeof value === "number" && Number.isFinite(value));
  const values = () =>
    Array.isArray(arg.values) &&
    arg.values.length > 0 &&
    arg.values.every(scalar);
  switch (op) {
    case "eq":
      return (
        keys.length === 1 &&
        ((keys[0] === "value" && (arg.value === null || scalar(arg.value))) ||
          (keys[0] === "values" && values()))
      );
    case "contains_any":
    case "contains_all":
      return keys.length === 1 && keys[0] === "values" && values();
    case "prefix":
      return (
        keys.length === 1 &&
        keys[0] === "value" &&
        typeof arg.value === "string" &&
        arg.value.length > 0
      );
    case "full_text":
      return (
        keys.length === 1 &&
        keys[0] === "query" &&
        typeof arg.query === "string" &&
        arg.query.length > 0
      );
    case "range":
      return (
        keys.length > 0 &&
        keys.every(
          (key) =>
            ["gt", "gte", "lt", "lte"].includes(key) &&
            typeof arg[key] === "string",
        ) &&
        !("gt" in arg && "gte" in arg) &&
        !("lt" in arg && "lte" in arg)
      );
    default:
      return false;
  }
}

/** Validates structure and request correlation; text normalization belongs to the server. */
export function readWorkbookQueryMetadata(
  contract: ViewContract,
  meta: unknown,
  requested: WorkbookQueryState,
  limit: number,
  cursor: string | undefined,
  expected?: WorkbookCanonicalQuery,
): { canonicalQuery: WorkbookCanonicalQuery; paging: WorkbookQueryPaging } {
  if (!object(meta) || !object(meta.query) || !object(meta.paging))
    throw new Error("Missing query metadata");
  const query = meta.query,
    paging = meta.paging;
  if (
    !Array.isArray(query.filters) ||
    !Array.isArray(query.sort) ||
    paging.limit !== limit ||
    typeof paging.has_more !== "boolean" ||
    (paging.has_more
      ? typeof paging.next_cursor !== "string" ||
        paging.next_cursor.length === 0 ||
        paging.next_cursor === cursor
      : paging.next_cursor !== null) ||
    (query.group_by !== undefined &&
      (typeof query.group_by !== "string" ||
        !contract.groupableFieldMap[query.group_by])) ||
    (query.group_by ?? null) !== requested.groupBy
  )
    throw new Error("Invalid query metadata");
  const filters = query.filters.map((filter) => {
    if (
      !object(filter) ||
      typeof filter.field_key !== "string" ||
      typeof filter.op !== "string" ||
      !object(filter.arg) ||
      !validArgument(filter.op, filter.arg) ||
      !contract.fieldMap[filter.field_key]?.filterOps.includes(filter.op) ||
      !requested.filters.some(
        (entry) =>
          entry.fieldKey === filter.field_key && entry.op === filter.op,
      )
    )
      throw new Error("Mismatched query filter");
    return {
      fieldKey: filter.field_key,
      op: filter.op,
      arg: structuredClone(filter.arg),
    } as WorkbookQueryState["filters"][number];
  });
  if (
    filters.length !== requested.filters.length ||
    filters.some(
      (filter, index) =>
        index > 0 && (filters[index - 1]?.fieldKey ?? "") >= filter.fieldKey,
    )
  )
    throw new Error("Mismatched filters");
  const sort = query.sort.map((entry) => {
    if (
      !object(entry) ||
      typeof entry.field_key !== "string" ||
      (entry.field_key !== "record_id" &&
        !contract.sortableFieldMap[entry.field_key]) ||
      (entry.direction !== "asc" && entry.direction !== "desc")
    )
      throw new Error("Invalid effective sort");
    return { fieldKey: entry.field_key, direction: entry.direction };
  });
  const effectiveSort = [...requested.sort];
  for (const entry of [
    ...contract.defaultSort,
    { fieldKey: "record_id", direction: "asc" as const },
  ]) {
    if (
      !effectiveSort.some((candidate) => candidate.fieldKey === entry.fieldKey)
    )
      effectiveSort.push(entry);
  }
  if (!savedViewJSONEqual(effectiveSort, sort))
    throw new Error("Mismatched effective sort");
  const canonicalQuery: WorkbookCanonicalQuery = {
    filters,
    sort: sort as WorkbookQueryState["sort"],
    ...(typeof query.group_by === "string" ? { groupBy: query.group_by } : {}),
  };
  if (expected && !savedViewJSONEqual(expected, canonicalQuery))
    throw new Error("Continuation query mismatch");
  return {
    canonicalQuery,
    paging: {
      limit,
      hasMore: paging.has_more,
      nextCursor: paging.next_cursor as string | null,
    },
  };
}

/** Preserve author sort intent; effective default tails never become overrides. */
export function canonicalAuthoredQuery(
  requested: WorkbookQueryState,
  canonical: WorkbookCanonicalQuery,
): WorkbookQueryState {
  return {
    filters: canonical.filters,
    sort: requested.sort,
    groupBy: canonical.groupBy ?? null,
  };
}
