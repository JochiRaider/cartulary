import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookResolvedLayoutState } from "../layout/workbookColumnLayout";
import {
  type WorkbookChromeMode,
  workbookQueryChipCapacity,
} from "../layout/workbookResponsiveLayout";
import {
  type FilterDraft,
  filterChipLabel,
  filterInputMode,
  type WorkbookQueryState,
  type WorkbookSortEntry,
} from "./workbookQuery";

export const workbookOrderedSortLimit = 8;

export type WorkbookGridQueryCommand =
  | { readonly kind: "sort_add"; readonly fieldKey: string }
  | {
      readonly kind: "sort_set_direction";
      readonly direction: WorkbookSortEntry["direction"];
      readonly fieldKey: string;
    }
  | {
      readonly kind: "sort_move";
      readonly direction: "earlier" | "later";
      readonly fieldKey: string;
    }
  | { readonly kind: "sort_remove"; readonly fieldKey: string }
  | { readonly kind: "group_set"; readonly fieldKey: string | null }
  | { readonly kind: "filter_remove"; readonly fieldKey: string }
  | { readonly kind: "query_clear" }
  | {
      readonly kind: "column_set_hidden";
      readonly fieldKey: string;
      readonly hidden: boolean;
    }
  | {
      readonly kind: "column_move";
      readonly direction: "earlier" | "later";
      readonly fieldKey: string;
    }
  | { readonly kind: "columns_reset" };

export type WorkbookSortCommand = Extract<
  WorkbookGridQueryCommand,
  {
    readonly kind:
      | "sort_add"
      | "sort_set_direction"
      | "sort_move"
      | "sort_remove";
  }
>;

export type WorkbookQueryChip = {
  readonly command: WorkbookGridQueryCommand;
  readonly key: string;
  readonly label: string;
  readonly testId?: string | undefined;
};

export type WorkbookSortControlEntry = {
  readonly direction: WorkbookSortEntry["direction"];
  readonly fieldKey: string;
  readonly label: string;
  readonly priority: number;
};

export type WorkbookColumnControlEntry = {
  readonly fieldKey: string;
  readonly hidden: boolean;
  readonly label: string;
  readonly position: number;
};

export type WorkbookGridQueryControlProjection = {
  readonly activeGroupLabel: string;
  readonly activeSortLabel: string | null;
  readonly chips: readonly WorkbookQueryChip[];
  readonly columns: readonly WorkbookColumnControlEntry[];
  readonly groupOptions: readonly {
    readonly fieldKey: string;
    readonly label: string;
  }[];
  readonly hiddenChips: readonly WorkbookQueryChip[];
  readonly sortEntries: readonly WorkbookSortControlEntry[];
  readonly sortableFields: readonly {
    readonly fieldKey: string;
    readonly label: string;
  }[];
  readonly visibleChipCapacity: number;
  readonly visibleChips: readonly WorkbookQueryChip[];
};

export type WorkbookGridControlPanel = "columns" | "filters" | "sort";

export type WorkbookGridSurfaceTransientState = {
  readonly filterDraft: FilterDraft;
  readonly openPanel: WorkbookGridControlPanel | null;
};

export type WorkbookGridControlsTransientState = {
  readonly activeSurface: string;
  readonly surfaces: ReadonlyMap<string, WorkbookGridSurfaceTransientState>;
};

export type WorkbookGridControlsTransientEvent =
  | {
      readonly type: "activate_surface";
      readonly filterDraft: FilterDraft;
      readonly surface: string;
    }
  | {
      readonly type: "sync_filter_draft";
      readonly filterDraft: FilterDraft;
      readonly surface: string;
    }
  | {
      readonly type: "toggle_panel";
      readonly panel: WorkbookGridControlPanel;
      readonly surface: string;
    }
  | { readonly type: "close_panel"; readonly surface: string }
  | {
      readonly type: "change_filter_draft";
      readonly filterDraft: FilterDraft;
      readonly surface: string;
    }
  | {
      readonly type: "complete_filter";
      readonly filterDraft: FilterDraft;
      readonly surface: string;
    };

export function projectWorkbookGridQueryControls({
  chromeMode,
  contract,
  filterChipTestId,
  layoutState,
  queryState,
}: {
  readonly chromeMode: WorkbookChromeMode;
  readonly contract: ViewContract;
  readonly filterChipTestId: (fieldKey: string) => string;
  readonly layoutState: WorkbookResolvedLayoutState;
  readonly queryState: WorkbookQueryState;
}): WorkbookGridQueryControlProjection {
  const chips = projectQueryChips(contract, queryState, filterChipTestId);
  const visibleChipCapacity = workbookQueryChipCapacity(chromeMode);
  return {
    activeGroupLabel: fieldLabel(contract, queryState.groupBy) ?? "None",
    activeSortLabel: fieldLabel(contract, queryState.sort[0]?.fieldKey ?? null),
    chips,
    columns: projectColumns(contract, layoutState),
    groupOptions: contract.groupingFields.map((fieldKey) => ({
      fieldKey,
      label: fieldLabel(contract, fieldKey) ?? fieldKey,
    })),
    hiddenChips: chips.slice(visibleChipCapacity),
    sortEntries: queryState.sort.map((entry, index) => ({
      ...entry,
      label: fieldLabel(contract, entry.fieldKey) ?? entry.fieldKey,
      priority: index + 1,
    })),
    sortableFields: contract.fields.flatMap((field) =>
      contract.sortableFieldMap[field.fieldKey]
        ? [{ fieldKey: field.fieldKey, label: field.label }]
        : [],
    ),
    visibleChipCapacity,
    visibleChips: chips.slice(0, visibleChipCapacity),
  };
}

export function applyWorkbookSortCommand(
  contract: ViewContract,
  sort: readonly WorkbookSortEntry[],
  command: WorkbookSortCommand,
): readonly WorkbookSortEntry[] {
  if (!contract.sortableFieldMap[command.fieldKey]) return sort;
  switch (command.kind) {
    case "sort_add":
      return addSortField(sort, command.fieldKey);
    case "sort_set_direction":
      return setSortDirection(sort, command.fieldKey, command.direction);
    case "sort_move":
      return moveSortField(sort, command.fieldKey, command.direction);
    case "sort_remove":
      return removeSortField(sort, command.fieldKey);
  }
}

export function parseWorkbookBooleanDraftValue(
  value: string,
): FilterDraft["booleanValue"] | null {
  return value === "" || value === "false" || value === "true" ? value : null;
}

export function parseDeclaredFieldKey(
  value: string,
  declaredFieldKeys: readonly string[],
): string | null {
  return declaredFieldKeys.includes(value) ? value : null;
}

export function parseDeclaredGroupField(
  value: string,
  declaredFieldKeys: readonly string[],
):
  | { readonly kind: "none" }
  | { readonly kind: "field"; readonly fieldKey: string }
  | null {
  if (value === "") return { kind: "none" };
  const fieldKey = parseDeclaredFieldKey(value, declaredFieldKeys);
  return fieldKey === null ? null : { kind: "field", fieldKey };
}

export function validateFilterDraft(
  contract: ViewContract,
  draft: FilterDraft,
):
  | { readonly kind: "valid" }
  | { readonly kind: "invalid"; readonly message: string } {
  if (parseDeclaredFieldKey(draft.fieldKey, contract.filterFields) === null) {
    return { kind: "invalid", message: "Select a supported filter field." };
  }
  const missing =
    filterInputMode(draft.fieldKey) === "boolean"
      ? draft.booleanValue === ""
      : draft.value.trim() === "";
  return missing
    ? { kind: "invalid", message: "Enter a value before applying this filter." }
    : { kind: "valid" };
}

export function createWorkbookGridControlsTransientState(
  surface: string,
  filterDraft: FilterDraft,
  defaultFilterPopoverOpen: boolean,
): WorkbookGridControlsTransientState {
  return {
    activeSurface: surface,
    surfaces: new Map([
      [
        surface,
        {
          filterDraft,
          openPanel: defaultFilterPopoverOpen ? "filters" : null,
        },
      ],
    ]),
  };
}

export function workbookGridSurfaceTransientState(
  state: WorkbookGridControlsTransientState,
  surface: string,
  filterDraft: FilterDraft,
): WorkbookGridSurfaceTransientState {
  return state.surfaces.get(surface) ?? { filterDraft, openPanel: null };
}

export function reduceWorkbookGridControlsTransientState(
  state: WorkbookGridControlsTransientState,
  event: WorkbookGridControlsTransientEvent,
): WorkbookGridControlsTransientState {
  switch (event.type) {
    case "activate_surface":
      return activateSurface(state, event);
    case "sync_filter_draft":
      return updateSurface(state, event.surface, (current) =>
        current.openPanel === "filters"
          ? current
          : { ...current, filterDraft: event.filterDraft },
      );
    case "toggle_panel":
      return updateSurface(state, event.surface, (current) => ({
        ...current,
        openPanel: current.openPanel === event.panel ? null : event.panel,
      }));
    case "close_panel":
      return updateSurface(state, event.surface, (current) =>
        current.openPanel === null ? current : { ...current, openPanel: null },
      );
    case "change_filter_draft":
      return updateSurface(state, event.surface, (current) => ({
        ...current,
        filterDraft: event.filterDraft,
      }));
    case "complete_filter":
      return updateSurface(state, event.surface, (current) => ({
        ...current,
        filterDraft: event.filterDraft,
        openPanel: null,
      }));
  }
}

function projectQueryChips(
  contract: ViewContract,
  queryState: WorkbookQueryState,
  filterChipTestId: (fieldKey: string) => string,
): readonly WorkbookQueryChip[] {
  const groupChip: readonly WorkbookQueryChip[] =
    queryState.groupBy === null
      ? []
      : [
          {
            command: { kind: "group_set", fieldKey: null },
            key: `group:${queryState.groupBy}`,
            label: `Group: ${fieldLabel(contract, queryState.groupBy) ?? queryState.groupBy}`,
          },
        ];
  const sortChips = queryState.sort.map(
    (entry): WorkbookQueryChip => ({
      command: { kind: "sort_remove", fieldKey: entry.fieldKey },
      key: `sort:${entry.fieldKey}`,
      label: `Sort: ${fieldLabel(contract, entry.fieldKey) ?? entry.fieldKey} ${entry.direction}`,
    }),
  );
  const filterChips = queryState.filters.map(
    (filter): WorkbookQueryChip => ({
      command: { kind: "filter_remove", fieldKey: filter.fieldKey },
      key: `filter:${filter.fieldKey}`,
      label: filterChipLabel(contract, filter),
      testId: filterChipTestId(filter.fieldKey),
    }),
  );
  return [...groupChip, ...sortChips, ...filterChips];
}

function projectColumns(
  contract: ViewContract,
  layoutState: WorkbookResolvedLayoutState,
): readonly WorkbookColumnControlEntry[] {
  const hidden = new Set(layoutState.hiddenFieldKeys);
  return layoutState.columnOrder.flatMap((fieldKey, position) => {
    const field = contract.fieldMap[fieldKey];
    return field === undefined
      ? []
      : [
          {
            fieldKey,
            hidden: hidden.has(fieldKey),
            label: field.label,
            position,
          },
        ];
  });
}

function addSortField(
  sort: readonly WorkbookSortEntry[],
  fieldKey: string,
): readonly WorkbookSortEntry[] {
  if (
    sort.length >= workbookOrderedSortLimit ||
    sort.some((entry) => entry.fieldKey === fieldKey)
  ) {
    return sort;
  }
  return [...sort, { direction: "asc", fieldKey }];
}

function setSortDirection(
  sort: readonly WorkbookSortEntry[],
  fieldKey: string,
  direction: WorkbookSortEntry["direction"],
): readonly WorkbookSortEntry[] {
  const index = sort.findIndex((entry) => entry.fieldKey === fieldKey);
  const current = sort[index];
  if (current === undefined || current.direction === direction) return sort;
  return sort.map((entry, entryIndex) =>
    entryIndex === index ? { ...entry, direction } : entry,
  );
}

function moveSortField(
  sort: readonly WorkbookSortEntry[],
  fieldKey: string,
  direction: "earlier" | "later",
): readonly WorkbookSortEntry[] {
  const index = sort.findIndex((entry) => entry.fieldKey === fieldKey);
  const nextIndex = direction === "earlier" ? index - 1 : index + 1;
  if (index < 0 || nextIndex < 0 || nextIndex >= sort.length) return sort;
  const next = [...sort];
  const current = next[index];
  const adjacent = next[nextIndex];
  if (current === undefined || adjacent === undefined) return sort;
  next[index] = adjacent;
  next[nextIndex] = current;
  return next;
}

function removeSortField(
  sort: readonly WorkbookSortEntry[],
  fieldKey: string,
): readonly WorkbookSortEntry[] {
  const index = sort.findIndex((entry) => entry.fieldKey === fieldKey);
  return index < 0 ? sort : [...sort.slice(0, index), ...sort.slice(index + 1)];
}

function activateSurface(
  state: WorkbookGridControlsTransientState,
  event: Extract<
    WorkbookGridControlsTransientEvent,
    { readonly type: "activate_surface" }
  >,
): WorkbookGridControlsTransientState {
  if (state.activeSurface === event.surface) return state;
  const surfaces = new Map(state.surfaces);
  const previous = surfaces.get(state.activeSurface);
  if (previous !== undefined && previous.openPanel !== null) {
    surfaces.set(state.activeSurface, { ...previous, openPanel: null });
  }
  const current = surfaces.get(event.surface);
  surfaces.set(
    event.surface,
    current ?? {
      filterDraft: event.filterDraft,
      openPanel: null,
    },
  );
  return { activeSurface: event.surface, surfaces };
}

function updateSurface(
  state: WorkbookGridControlsTransientState,
  surface: string,
  update: (
    current: WorkbookGridSurfaceTransientState,
  ) => WorkbookGridSurfaceTransientState,
): WorkbookGridControlsTransientState {
  const current = state.surfaces.get(surface);
  if (current === undefined) return state;
  const next = update(current);
  if (next === current) return state;
  const surfaces = new Map(state.surfaces);
  surfaces.set(surface, next);
  return { ...state, surfaces };
}

function fieldLabel(
  contract: ViewContract,
  fieldKey: string | null,
): string | null {
  if (fieldKey === null) return null;
  return contract.fieldMap[fieldKey]?.label ?? fieldKey;
}
