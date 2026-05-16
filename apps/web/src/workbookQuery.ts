import {
  resolveHeaderSortFieldKey,
  type ViewContract,
} from "@cartulary/view-contracts";

export type WorkbookFilter = {
  readonly arg: Record<string, unknown>;
  readonly fieldKey: string;
  readonly op: string;
};

export type WorkbookSortEntry = {
  readonly direction: "asc" | "desc";
  readonly fieldKey: string;
};

export type WorkbookQueryState = {
  readonly filters: readonly WorkbookFilter[];
  readonly groupBy: string | null;
  readonly sort: readonly WorkbookSortEntry[];
};

export type FilterDraft = {
  readonly booleanValue: "" | "false" | "true";
  readonly fieldKey: string;
  readonly value: string;
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
  return {
    booleanValue: "",
    fieldKey: fieldKey ?? "",
    value: "",
  };
}

export function toggleSortField(
  contract: ViewContract,
  state: WorkbookQueryState,
  fieldKey: string,
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
      sort: [{ fieldKey: sortableFieldKey, direction: "asc" }],
    };
  }
  if (existing.direction === "asc") {
    return {
      ...state,
      sort: [{ fieldKey: sortableFieldKey, direction: "desc" }],
    };
  }
  return {
    ...state,
    sort: state.sort.filter((entry) => entry.fieldKey !== sortableFieldKey),
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
): Record<string, unknown> {
  const request: Record<string, unknown> = {};
  const sort = normalizeSortForRequest(contract, state);
  if (sort.length > 0) {
    request.sort = sort.map((entry) => ({
      direction: entry.direction,
      field_key: entry.fieldKey,
    }));
  }
  const filters = state.filters.filter(
    (filter) => contract.filterableFieldMap[filter.fieldKey],
  );
  if (filters.length > 0) {
    request.filters = filters.map((filter) => ({
      arg: filter.arg,
      field_key: filter.fieldKey,
      op: filter.op,
    }));
  }
  if (state.groupBy && contract.groupableFieldMap[state.groupBy]) {
    request.group_by = state.groupBy;
  }
  return request;
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
  if (
    fieldKey === "timeline.occurred_day" ||
    fieldKey === "timeline.recorded_day"
  ) {
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
  const sort = state.sort.filter(
    (entry) => contract.sortableFieldMap[entry.fieldKey],
  );
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

function buildFilterFromDraft(draft: FilterDraft): WorkbookFilter | null {
  const mode = filterInputMode(draft.fieldKey);
  if (draft.fieldKey === "") {
    return null;
  }
  if (mode === "boolean") {
    if (draft.booleanValue === "") {
      return null;
    }
    return {
      fieldKey: draft.fieldKey,
      op: "eq",
      arg: {
        value: draft.booleanValue === "true",
      },
    };
  }

  const trimmed = draft.value.trim();
  if (trimmed === "") {
    return null;
  }

  if (draft.fieldKey === "note.full_text") {
    return {
      fieldKey: draft.fieldKey,
      op: "full_text",
      arg: { query: trimmed },
    };
  }

  if (mode === "tagset") {
    const values = trimmed
      .split(/[\n,]/u)
      .map((value) => value.trim())
      .filter((value) => value !== "");
    if (values.length < 1) {
      return null;
    }
    return {
      fieldKey: draft.fieldKey,
      op: "contains_any",
      arg: { values },
    };
  }

  if (mode === "date" || mode === "timestamp") {
    const range = parseRange(trimmed);
    if (range !== null) {
      return {
        fieldKey: draft.fieldKey,
        op: "range",
        arg: range,
      };
    }
  }

  return {
    fieldKey: draft.fieldKey,
    op: "eq",
    arg: { value: trimmed },
  };
}

function parseRange(value: string): Record<string, string> | null {
  const [rawStart, rawEnd, ...rest] = value.split("..");
  if (rest.length > 0 || rawStart === undefined || rawEnd === undefined) {
    return null;
  }
  const start = rawStart.trim();
  const end = rawEnd.trim();
  if (start === "" && end === "") {
    return null;
  }
  const range: Record<string, string> = {};
  if (start !== "") {
    range.gte = start;
  }
  if (end !== "") {
    range.lte = end;
  }
  return Object.keys(range).length < 1 ? null : range;
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
