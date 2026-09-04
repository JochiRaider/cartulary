import {
  resolveHeaderSortFieldKey,
  type ViewContract,
} from "@cartulary/view-contracts";
import type { WorkbookProtocolQueryViewRequest } from "../adapters/workbookProtocolTypes";

export type WorkbookFilter = {
  readonly arg: Record<string, unknown>;
  readonly fieldKey: string;
  readonly op: WorkbookFilterOperator;
};

export type WorkbookFilterOperator =
  | "contains_all"
  | "contains_any"
  | "eq"
  | "full_text"
  | "prefix"
  | "range";

export type WorkbookSortEntry = {
  readonly direction: "asc" | "desc";
  readonly fieldKey: string;
};

export type WorkbookQueryState = {
  readonly filters: readonly WorkbookFilter[];
  readonly groupBy: string | null;
  readonly sort: readonly WorkbookSortEntry[];
};

export type WorkbookSavedViewFilterJson = {
  readonly arg: Record<string, unknown>;
  readonly field_key: string;
  readonly op: WorkbookFilterOperator;
};

export type WorkbookSavedViewQueryJson = {
  readonly filters: readonly WorkbookSavedViewFilterJson[];
  readonly group_by?: string;
  readonly sort: readonly WorkbookSavedViewSortJson[];
};

export type WorkbookSavedViewSortJson = {
  readonly direction: "asc" | "desc";
  readonly field_key: string;
};

export type WorkbookLayoutColumnWidth = {
  readonly fieldKey: string;
  readonly widthPx: number;
};

export type WorkbookLayoutState = {
  readonly columnOrder?: readonly string[] | undefined;
  readonly columnWidths?:
    | Readonly<Record<string, number>>
    | readonly WorkbookLayoutColumnWidth[]
    | undefined;
  readonly hiddenFieldKeys?: readonly string[] | undefined;
};

export type WorkbookSavedViewLayoutJson = {
  readonly column_order: readonly string[];
  readonly column_widths: readonly {
    readonly field_key: string;
    readonly width_px: number;
  }[];
  readonly hidden_field_keys: readonly string[];
  readonly layout_schema_id: "cartulary.layout.v1";
};

export type FilterDraft =
  | {
      readonly booleanValue: "" | "false" | "true";
      readonly fieldKey: string;
      readonly op: "eq";
      readonly operandKind: "null" | "value" | "values";
      readonly value: string;
      readonly valueType: "boolean" | "number" | "string";
      readonly values: string;
    }
  | {
      readonly fieldKey: string;
      readonly lowerKind: "gt" | "gte";
      readonly lowerValue: string;
      readonly op: "range";
      readonly upperKind: "lt" | "lte";
      readonly upperValue: string;
    }
  | {
      readonly fieldKey: string;
      readonly op: "contains_all" | "contains_any";
      readonly values: string;
    }
  | {
      readonly fieldKey: string;
      readonly op: "prefix";
      readonly value: string;
    }
  | {
      readonly fieldKey: string;
      readonly op: "full_text";
      readonly query: string;
    };

export type FilterInputMode =
  | "boolean"
  | "date"
  | "tagset"
  | "text"
  | "timestamp";

export function emptyWorkbookQueryState(): WorkbookQueryState {
  return {
    filters: [],
    groupBy: null,
    sort: [],
  };
}

export function defaultFilterDraft(contract: ViewContract): FilterDraft {
  const [fieldKey] = contract.filterFields;
  return filterDraftForField(contract, fieldKey ?? "");
}

export function filterDraftForField(
  contract: ViewContract,
  fieldKey: string,
  requestedOperator?: WorkbookFilterOperator,
): FilterDraft {
  const field = contract.fieldMap[fieldKey];
  const allowed = field?.filterOps.filter(isWorkbookFilterOperator) ?? [];
  const op =
    requestedOperator !== undefined && allowed.includes(requestedOperator)
      ? requestedOperator
      : (allowed[0] ?? "eq");
  switch (op) {
    case "range":
      return {
        fieldKey,
        lowerKind: "gte",
        lowerValue: "",
        op,
        upperKind: "lte",
        upperValue: "",
      };
    case "contains_all":
    case "contains_any":
      return { fieldKey, op, values: "" };
    case "prefix":
      return { fieldKey, op, value: "" };
    case "full_text":
      return { fieldKey, op, query: "" };
    case "eq":
      return {
        booleanValue: "",
        fieldKey,
        op,
        operandKind: "value",
        value: "",
        valueType:
          filterInputMode(fieldKey) === "boolean" ? "boolean" : "string",
        values: "",
      };
  }
}

export function filterDraftFromFilter(filter: WorkbookFilter): FilterDraft {
  switch (filter.op) {
    case "range":
      return {
        fieldKey: filter.fieldKey,
        lowerKind: "gt" in filter.arg ? "gt" : "gte",
        lowerValue: String(filter.arg.gt ?? filter.arg.gte ?? ""),
        op: "range",
        upperKind: "lt" in filter.arg ? "lt" : "lte",
        upperValue: String(filter.arg.lt ?? filter.arg.lte ?? ""),
      };
    case "contains_all":
    case "contains_any":
      return {
        fieldKey: filter.fieldKey,
        op: filter.op,
        values: Array.isArray(filter.arg.values)
          ? filter.arg.values.map(String).join(", ")
          : "",
      };
    case "prefix":
      return {
        fieldKey: filter.fieldKey,
        op: "prefix",
        value: typeof filter.arg.value === "string" ? filter.arg.value : "",
      };
    case "full_text":
      return {
        fieldKey: filter.fieldKey,
        op: "full_text",
        query: typeof filter.arg.query === "string" ? filter.arg.query : "",
      };
    case "eq": {
      const values = Array.isArray(filter.arg.values)
        ? filter.arg.values.map(String).join(", ")
        : "";
      const rawValue = filter.arg.value;
      return {
        booleanValue:
          typeof rawValue === "boolean"
            ? (String(rawValue) as "false" | "true")
            : "",
        fieldKey: filter.fieldKey,
        op: "eq",
        operandKind: Array.isArray(filter.arg.values)
          ? "values"
          : rawValue === null
            ? "null"
            : "value",
        value:
          typeof rawValue === "string" || typeof rawValue === "number"
            ? String(rawValue)
            : "",
        valueType:
          typeof rawValue === "boolean"
            ? "boolean"
            : typeof rawValue === "number"
              ? "number"
              : "string",
        values,
      };
    }
  }
}

export function clearFilterDraftValue(draft: FilterDraft): FilterDraft {
  switch (draft.op) {
    case "eq":
      return { ...draft, booleanValue: "", value: "", values: "" };
    case "range":
      return { ...draft, lowerValue: "", upperValue: "" };
    case "contains_all":
    case "contains_any":
      return { ...draft, values: "" };
    case "prefix":
      return { ...draft, value: "" };
    case "full_text":
      return { ...draft, query: "" };
  }
}

export function isWorkbookFilterOperator(
  value: string,
): value is WorkbookFilterOperator {
  return (
    value === "eq" ||
    value === "range" ||
    value === "contains_any" ||
    value === "contains_all" ||
    value === "prefix" ||
    value === "full_text"
  );
}

export function toggleSortField(
  contract: ViewContract,
  state: WorkbookQueryState,
  fieldKey: string,
): WorkbookQueryState {
  return cycleWorkbookSortField(contract, state, fieldKey, false);
}

export function cycleWorkbookSortField(
  contract: ViewContract,
  state: WorkbookQueryState,
  fieldKey: string,
  additive: boolean,
): WorkbookQueryState {
  const sortableFieldKey = resolveHeaderSortFieldKey(contract, fieldKey);
  if (!sortableFieldKey || !contract.sortableFieldMap[sortableFieldKey]) {
    return state;
  }

  const existing = state.sort.find(
    (entry) => entry.fieldKey === sortableFieldKey,
  );
  if (!existing) {
    return {
      ...state,
      sort: additive
        ? [
            ...state.sort,
            { fieldKey: sortableFieldKey, direction: "asc" as const },
          ].slice(0, 8)
        : [{ fieldKey: sortableFieldKey, direction: "asc" }],
    };
  }
  if (existing.direction === "asc") {
    return {
      ...state,
      sort: additive
        ? state.sort.map((entry) =>
            entry.fieldKey === sortableFieldKey
              ? { ...entry, direction: "desc" as const }
              : entry,
          )
        : [{ fieldKey: sortableFieldKey, direction: "desc" }],
    };
  }
  return {
    ...state,
    sort: additive
      ? state.sort.filter((entry) => entry.fieldKey !== sortableFieldKey)
      : [],
  };
}

export function replaceWorkbookSort(
  contract: ViewContract,
  state: WorkbookQueryState,
  sort: readonly WorkbookSortEntry[],
): WorkbookQueryState {
  return {
    ...state,
    sort: normalizeUserSortForPersistence(contract, { ...state, sort }).slice(
      0,
      8,
    ),
  };
}

export function updateGroupBy(
  contract: ViewContract,
  state: WorkbookQueryState,
  groupBy: string | null,
): WorkbookQueryState {
  if (groupBy === null || groupBy === "") {
    return {
      ...state,
      groupBy: null,
    };
  }
  if (!contract.groupableFieldMap[groupBy]) {
    return state;
  }
  return {
    ...state,
    groupBy,
  };
}

export function applyFilterDraft(
  state: WorkbookQueryState,
  draft: FilterDraft,
): WorkbookQueryState {
  const nextFilter = buildFilterFromDraft(draft);
  if (nextFilter === null) {
    return state;
  }
  return {
    ...state,
    filters: [
      ...state.filters.filter(
        (filter) => filter.fieldKey !== nextFilter.fieldKey,
      ),
      nextFilter,
    ].sort((left, right) => left.fieldKey.localeCompare(right.fieldKey)),
  };
}

export function removeFilterField(
  state: WorkbookQueryState,
  fieldKey: string,
): WorkbookQueryState {
  return {
    ...state,
    filters: state.filters.filter((filter) => filter.fieldKey !== fieldKey),
  };
}

export function buildQueryRequest(
  contract: ViewContract,
  state: WorkbookQueryState,
): WorkbookProtocolQueryViewRequest {
  const request: WorkbookProtocolQueryViewRequest = {};
  const sort = normalizeSortForRequest(contract, state);
  if (sort.length > 0) {
    request.sort = sort.map((entry) => ({
      direction: entry.direction,
      field_key: entry.fieldKey,
    }));
  }
  const filters = normalizeFiltersForWire(contract, state);
  if (filters.length > 0) {
    request.filters = [...filters];
  }
  if (state.groupBy && contract.groupableFieldMap[state.groupBy]) {
    request.group_by = state.groupBy;
  }
  return request;
}

export function buildSavedViewQueryJson(
  contract: ViewContract,
  state: WorkbookQueryState,
): WorkbookSavedViewQueryJson {
  const queryJson: WorkbookSavedViewQueryJson = {
    filters: normalizeFiltersForWire(contract, state),
    sort: normalizeUserSortForPersistence(contract, state).map((entry) => ({
      direction: entry.direction,
      field_key: entry.fieldKey,
    })),
  };
  if (state.groupBy && contract.groupableFieldMap[state.groupBy]) {
    return {
      ...queryJson,
      group_by: state.groupBy,
    };
  }
  return queryJson;
}

export function buildSavedViewLayoutJson(
  contract: ViewContract,
  state: WorkbookLayoutState = {},
): WorkbookSavedViewLayoutJson {
  const fieldKeys = contract.fields.map((field) => field.fieldKey);
  const allowed = new Set(fieldKeys);
  const columnOrder = canonicalColumnOrder(fieldKeys, state.columnOrder);
  return {
    layout_schema_id: "cartulary.layout.v1",
    column_order: columnOrder,
    hidden_field_keys: canonicalHiddenFieldKeys(
      allowed,
      state.hiddenFieldKeys ?? contract.defaultHiddenFields,
    ),
    column_widths: canonicalColumnWidths(allowed, state.columnWidths),
  };
}

export function workbookQueryStateFromSavedViewQueryJson(
  contract: ViewContract,
  value: unknown,
): WorkbookQueryState {
  if (!isObjectRecord(value)) {
    return emptyWorkbookQueryState();
  }
  const state: WorkbookQueryState = {
    filters: savedViewFiltersFromQueryJson(contract, value.filters),
    groupBy:
      typeof value.group_by === "string" &&
      contract.groupableFieldMap[value.group_by]
        ? value.group_by
        : null,
    sort: savedViewSortFromQueryJson(contract, value.sort),
  };
  return {
    ...state,
    filters: normalizeFiltersForWire(contract, state).map((filter) => ({
      arg: { ...filter.arg },
      fieldKey: filter.field_key,
      op: filter.op,
    })),
    sort: normalizeUserSortForPersistence(contract, state),
  };
}

export function workbookLayoutStateFromSavedViewLayoutJson(
  contract: ViewContract,
  value: unknown,
): WorkbookLayoutState {
  if (!isObjectRecord(value)) {
    return {};
  }
  return {
    columnOrder: Array.isArray(value.column_order)
      ? value.column_order.filter(
          (fieldKey): fieldKey is string => typeof fieldKey === "string",
        )
      : [],
    columnWidths: Array.isArray(value.column_widths)
      ? value.column_widths.filter(isObjectRecord).map((entry) => ({
          fieldKey: typeof entry.field_key === "string" ? entry.field_key : "",
          widthPx:
            typeof entry.width_px === "number"
              ? Math.trunc(entry.width_px)
              : Number.NaN,
        }))
      : [],
    hiddenFieldKeys: Array.isArray(value.hidden_field_keys)
      ? value.hidden_field_keys.filter(
          (fieldKey): fieldKey is string => typeof fieldKey === "string",
        )
      : contract.defaultHiddenFields,
  };
}

export function filterChipLabel(
  contract: ViewContract,
  filter: WorkbookFilter,
): string {
  const field = contract.fieldMap[filter.fieldKey];
  const label = field?.label ?? filter.fieldKey;
  return `${label}: ${stringifyFilterValue(filter)}`;
}

export function filterInputMode(fieldKey: string): FilterInputMode {
  if (
    fieldKey === "timeline.has_evidence" ||
    fieldKey === "timeline.has_unresolved_mentions"
  ) {
    return "boolean";
  }
  if (fieldKey === "timeline.date_entered_sort_day") {
    return "date";
  }
  if (
    fieldKey === "indicator.first_observed_at" ||
    fieldKey === "indicator.last_observed_at"
  ) {
    return "timestamp";
  }
  if (fieldKey === "timeline.tags") {
    return "tagset";
  }
  return "text";
}

function normalizeSortForRequest(
  contract: ViewContract,
  state: WorkbookQueryState,
): readonly WorkbookSortEntry[] {
  const sort = normalizeUserSortForPersistence(contract, state);
  if (!state.groupBy) {
    return sort;
  }
  if (sort.some((entry) => entry.fieldKey === state.groupBy)) {
    return sort;
  }
  if (!contract.groupableFieldMap[state.groupBy]) {
    return sort;
  }
  return [{ fieldKey: state.groupBy, direction: "asc" }, ...sort];
}

function savedViewSortFromQueryJson(
  contract: ViewContract,
  value: unknown,
): readonly WorkbookSortEntry[] {
  if (!Array.isArray(value)) {
    return [];
  }
  const seen = new Set<string>();
  const sort: WorkbookSortEntry[] = [];
  for (const entry of value) {
    if (!isObjectRecord(entry)) {
      continue;
    }
    if (
      typeof entry.field_key !== "string" ||
      !contract.sortableFieldMap[entry.field_key] ||
      seen.has(entry.field_key) ||
      (entry.direction !== "asc" && entry.direction !== "desc")
    ) {
      continue;
    }
    seen.add(entry.field_key);
    sort.push({
      direction: entry.direction,
      fieldKey: entry.field_key,
    });
    if (sort.length === 8) {
      break;
    }
  }
  return sort;
}

function savedViewFiltersFromQueryJson(
  contract: ViewContract,
  value: unknown,
): readonly WorkbookFilter[] {
  if (!Array.isArray(value)) {
    return [];
  }
  const filters = new Map<string, WorkbookFilter>();
  for (const entry of value) {
    if (!isObjectRecord(entry) || !isObjectRecord(entry.arg)) {
      continue;
    }
    if (
      typeof entry.field_key !== "string" ||
      typeof entry.op !== "string" ||
      !isWorkbookFilterOperator(entry.op) ||
      !contract.filterableFieldMap[entry.field_key]
    ) {
      continue;
    }
    const field = contract.fieldMap[entry.field_key];
    if (!field?.filterOps.includes(entry.op)) {
      continue;
    }
    const candidate: WorkbookFilter = {
      arg: { ...entry.arg },
      fieldKey: entry.field_key,
      op: entry.op,
    };
    if (normalizeFilterArg(candidate) === null) {
      continue;
    }
    filters.set(entry.field_key, candidate);
  }
  return [...filters.values()].sort((left, right) =>
    left.fieldKey.localeCompare(right.fieldKey),
  );
}

function isObjectRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function normalizeUserSortForPersistence(
  contract: ViewContract,
  state: WorkbookQueryState,
): readonly WorkbookSortEntry[] {
  const seen = new Set<string>();
  const sort: WorkbookSortEntry[] = [];
  for (const entry of state.sort) {
    if (
      !contract.sortableFieldMap[entry.fieldKey] ||
      seen.has(entry.fieldKey)
    ) {
      continue;
    }
    seen.add(entry.fieldKey);
    sort.push(entry);
    if (sort.length === 8) {
      break;
    }
  }
  return sort;
}

function normalizeFiltersForWire(
  contract: ViewContract,
  state: WorkbookQueryState,
): readonly WorkbookSavedViewFilterJson[] {
  const filters = new Map<string, WorkbookSavedViewFilterJson>();
  for (const filter of state.filters) {
    const field = contract.fieldMap[filter.fieldKey];
    if (
      !field ||
      !contract.filterableFieldMap[filter.fieldKey] ||
      !field.filterOps.includes(filter.op)
    ) {
      continue;
    }
    const normalized = normalizeFilterArg(filter);
    if (normalized === null) {
      continue;
    }
    filters.set(filter.fieldKey, {
      arg: normalized,
      field_key: filter.fieldKey,
      op: filter.op,
    });
  }
  return [...filters.values()].sort((left, right) =>
    left.field_key.localeCompare(right.field_key),
  );
}

function normalizeFilterArg(
  filter: WorkbookFilter,
): Record<string, unknown> | null {
  if (
    (filter.op === "contains_any" || filter.op === "contains_all") &&
    Array.isArray(filter.arg.values)
  ) {
    const values = canonicalStringValues(filter.arg.values);
    return values.length > 0 ? { values } : null;
  }
  return filter.arg;
}

function canonicalStringValues(values: readonly unknown[]): readonly string[] {
  return [
    ...new Set(
      values
        .filter((value): value is string => typeof value === "string")
        .map((value) => value.trim())
        .filter((value) => value !== ""),
    ),
  ].sort((left, right) => left.localeCompare(right));
}

function canonicalColumnOrder(
  fieldKeys: readonly string[],
  requested: readonly string[] = [],
): readonly string[] {
  const allowed = new Set(fieldKeys);
  const seen = new Set<string>();
  const ordered: string[] = [];
  for (const fieldKey of requested) {
    if (!allowed.has(fieldKey) || seen.has(fieldKey)) {
      continue;
    }
    seen.add(fieldKey);
    ordered.push(fieldKey);
  }
  for (const fieldKey of fieldKeys) {
    if (seen.has(fieldKey)) {
      continue;
    }
    ordered.push(fieldKey);
  }
  return ordered;
}

function canonicalHiddenFieldKeys(
  allowed: ReadonlySet<string>,
  values: readonly string[],
): readonly string[] {
  return [...new Set(values.filter((fieldKey) => allowed.has(fieldKey)))].sort(
    (left, right) => left.localeCompare(right),
  );
}

function canonicalColumnWidths(
  allowed: ReadonlySet<string>,
  values: WorkbookLayoutState["columnWidths"],
): WorkbookSavedViewLayoutJson["column_widths"] {
  const widths = new Map<string, number>();
  const entries = Array.isArray(values)
    ? values.map((value) => [value.fieldKey, value.widthPx] as const)
    : Object.entries(values ?? {});
  for (const [fieldKey, widthPx] of entries) {
    if (
      !allowed.has(fieldKey) ||
      !Number.isSafeInteger(widthPx) ||
      widthPx < 40 ||
      widthPx > 4096
    ) {
      continue;
    }
    widths.set(fieldKey, widthPx);
  }
  return [...widths.entries()]
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([fieldKey, widthPx]) => ({
      field_key: fieldKey,
      width_px: widthPx,
    }));
}

export function buildFilterFromDraft(
  draft: FilterDraft,
): WorkbookFilter | null {
  if (draft.fieldKey === "") {
    return null;
  }
  switch (draft.op) {
    case "eq": {
      if (draft.operandKind === "null") {
        return { arg: { value: null }, fieldKey: draft.fieldKey, op: "eq" };
      }
      if (draft.operandKind === "values") {
        const values = canonicalStringValues(draft.values.split(/[\n,]/u));
        return values.length === 0
          ? null
          : { arg: { values }, fieldKey: draft.fieldKey, op: "eq" };
      }
      if (draft.valueType === "boolean") {
        return draft.booleanValue === ""
          ? null
          : {
              arg: { value: draft.booleanValue === "true" },
              fieldKey: draft.fieldKey,
              op: "eq",
            };
      }
      const value = draft.value.trim();
      if (value === "") return null;
      if (draft.valueType === "number") {
        const numberValue = Number(value);
        return Number.isFinite(numberValue)
          ? { arg: { value: numberValue }, fieldKey: draft.fieldKey, op: "eq" }
          : null;
      }
      return { arg: { value }, fieldKey: draft.fieldKey, op: "eq" };
    }
    case "range": {
      const lower = draft.lowerValue.trim();
      const upper = draft.upperValue.trim();
      if (lower === "" && upper === "") return null;
      return {
        arg: {
          ...(lower === "" ? {} : { [draft.lowerKind]: lower }),
          ...(upper === "" ? {} : { [draft.upperKind]: upper }),
        },
        fieldKey: draft.fieldKey,
        op: "range",
      };
    }
    case "contains_all":
    case "contains_any": {
      const values = canonicalStringValues(draft.values.split(/[\n,]/u));
      return values.length === 0
        ? null
        : { arg: { values }, fieldKey: draft.fieldKey, op: draft.op };
    }
    case "prefix": {
      const value = draft.value.trim();
      return value === ""
        ? null
        : { arg: { value }, fieldKey: draft.fieldKey, op: "prefix" };
    }
    case "full_text": {
      const query = draft.query.trim();
      return query === ""
        ? null
        : { arg: { query }, fieldKey: draft.fieldKey, op: "full_text" };
    }
  }
}

function stringifyFilterValue(filter: WorkbookFilter): string {
  if (Array.isArray(filter.arg.values)) {
    return filter.arg.values.join(", ");
  }
  if (typeof filter.arg.value === "boolean") {
    return filter.arg.value ? "true" : "false";
  }
  if (typeof filter.arg.value === "string") {
    return filter.arg.value;
  }
  const rangeValues = ["gte", "gt", "lte", "lt"]
    .map((key) =>
      typeof filter.arg[key] === "string"
        ? `${key} ${String(filter.arg[key])}`
        : null,
    )
    .filter((value): value is string => value !== null);
  return rangeValues.join(" ");
}
