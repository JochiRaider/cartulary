import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookResolvedLayoutState } from "../layout/workbookColumnLayout";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import {
  buildFilterFromDraft,
  type FilterDraft,
  type WorkbookQueryState,
  type WorkbookSortEntry,
} from "./workbookQuery";
import {
  projectWorkbookQueryEntries,
  queryEntryCapacity,
  type WorkbookQueryEntryDescriptor,
} from "./workbookViewBarWorkingSet";

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
  | { readonly kind: "filters_clear" }
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

export type WorkbookQueryChip = WorkbookQueryEntryDescriptor & {
  readonly label: string;
  readonly removeCommand: WorkbookGridQueryCommand;
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
  readonly unusedSortableFields: readonly {
    readonly fieldKey: string;
    readonly label: string;
  }[];
  readonly visibleChipCapacity: number;
  readonly visibleChips: readonly WorkbookQueryChip[];
};

export type WorkbookGridControlPanel = "columns" | "filters" | "group" | "sort";

export type WorkbookGridSurfaceTransientState = {
  readonly activeEntryKey: string | null;
  readonly editingFilterFieldKey: string | null;
  readonly filterDraft: FilterDraft;
  readonly openPanel: WorkbookGridControlPanel | null;
  readonly rovingEntryKey: string | null;
};

export type WorkbookGridControlsTransientState = {
  readonly subjectKey: string;
  readonly value: WorkbookGridSurfaceTransientState;
};

export type WorkbookGridControlsTransientEvent =
  | {
      readonly type: "activate_subject";
      readonly filterDraft: FilterDraft;
      readonly subjectKey: string;
    }
  | {
      readonly type: "sync_filter_draft";
      readonly filterDraft: FilterDraft;
      readonly subjectKey: string;
    }
  | {
      readonly type: "toggle_panel";
      readonly panel: WorkbookGridControlPanel;
      readonly subjectKey: string;
    }
  | { readonly type: "close_panel"; readonly subjectKey: string }
  | {
      readonly type: "change_filter_draft";
      readonly filterDraft: FilterDraft;
      readonly subjectKey: string;
    }
  | {
      readonly type: "complete_filter";
      readonly filterDraft: FilterDraft;
      readonly subjectKey: string;
    }
  | {
      readonly activeEntryKey: string;
      readonly filterDraft: FilterDraft;
      readonly fieldKey: string;
      readonly subjectKey: string;
      readonly type: "edit_filter";
    }
  | {
      readonly activeEntryKey: string;
      readonly panel: "group" | "sort";
      readonly subjectKey: string;
      readonly type: "edit_query_entry";
    }
  | {
      readonly activeEntryKey: string | null;
      readonly subjectKey: string;
      readonly type: "set_active_entry";
    }
  | {
      readonly entryKey: string | null;
      readonly subjectKey: string;
      readonly type: "set_roving_entry";
    };

export function projectWorkbookGridQueryControls({
  chromeMode,
  contract,
  layoutState,
  queryState,
}: {
  readonly chromeMode: WorkbookChromeMode;
  readonly contract: ViewContract;
  readonly layoutState: WorkbookResolvedLayoutState;
  readonly queryState: WorkbookQueryState;
}): WorkbookGridQueryControlProjection {
  const chips = projectQueryChips(contract, queryState);
  const visibleChipCapacity = queryEntryCapacity(chromeMode);
  const selectedSortFields = new Set(
    queryState.sort.map((entry) => entry.fieldKey),
  );
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
    unusedSortableFields: contract.fields.flatMap((field) =>
      contract.sortableFieldMap[field.fieldKey] &&
      !selectedSortFields.has(field.fieldKey)
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
): "" | "false" | "true" | null {
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
  if (!contract.fieldMap[draft.fieldKey]?.filterOps.includes(draft.op)) {
    return {
      kind: "invalid",
      message: "Select a supported operator for this field.",
    };
  }
  return buildFilterFromDraft(draft) === null
    ? { kind: "invalid", message: "Enter a value before applying this filter." }
    : { kind: "valid" };
}

export function createWorkbookGridControlsTransientState(
  subjectKey: string,
  filterDraft: FilterDraft,
  defaultFilterPopoverOpen: boolean,
): WorkbookGridControlsTransientState {
  return {
    subjectKey,
    value: {
      activeEntryKey: null,
      editingFilterFieldKey: null,
      filterDraft,
      openPanel: defaultFilterPopoverOpen ? "filters" : null,
      rovingEntryKey: null,
    },
  };
}

export function workbookGridSurfaceTransientState(
  state: WorkbookGridControlsTransientState,
  subjectKey: string,
  filterDraft: FilterDraft,
): WorkbookGridSurfaceTransientState {
  return state.subjectKey === subjectKey
    ? state.value
    : {
        activeEntryKey: null,
        editingFilterFieldKey: null,
        filterDraft,
        openPanel: null,
        rovingEntryKey: null,
      };
}

export function reduceWorkbookGridControlsTransientState(
  state: WorkbookGridControlsTransientState,
  event: WorkbookGridControlsTransientEvent,
): WorkbookGridControlsTransientState {
  switch (event.type) {
    case "activate_subject":
      return state.subjectKey === event.subjectKey
        ? state
        : createWorkbookGridControlsTransientState(
            event.subjectKey,
            event.filterDraft,
            false,
          );
    case "sync_filter_draft":
      return updateTransient(state, event.subjectKey, (current) =>
        current.openPanel === "filters"
          ? current
          : { ...current, filterDraft: event.filterDraft },
      );
    case "toggle_panel":
      return updateTransient(state, event.subjectKey, (current) => ({
        ...current,
        activeEntryKey: null,
        editingFilterFieldKey: null,
        openPanel: current.openPanel === event.panel ? null : event.panel,
      }));
    case "close_panel":
      return updateTransient(state, event.subjectKey, (current) =>
        current.openPanel === null && current.activeEntryKey === null
          ? current
          : {
              ...current,
              activeEntryKey: null,
              editingFilterFieldKey: null,
              openPanel: null,
            },
      );
    case "change_filter_draft":
      return updateTransient(state, event.subjectKey, (current) => ({
        ...current,
        filterDraft: event.filterDraft,
      }));
    case "complete_filter":
      return updateTransient(state, event.subjectKey, (current) => ({
        ...current,
        activeEntryKey: null,
        editingFilterFieldKey: null,
        filterDraft: event.filterDraft,
        openPanel: null,
      }));
    case "edit_filter":
      return updateTransient(state, event.subjectKey, (current) => ({
        ...current,
        activeEntryKey: event.activeEntryKey,
        editingFilterFieldKey: event.fieldKey,
        filterDraft: event.filterDraft,
        openPanel: "filters",
      }));
    case "edit_query_entry":
      return updateTransient(state, event.subjectKey, (current) => ({
        ...current,
        activeEntryKey: event.activeEntryKey,
        editingFilterFieldKey: null,
        openPanel: event.panel,
      }));
    case "set_active_entry":
      return updateTransient(state, event.subjectKey, (current) => ({
        ...current,
        activeEntryKey: event.activeEntryKey,
      }));
    case "set_roving_entry":
      return updateTransient(state, event.subjectKey, (current) => ({
        ...current,
        rovingEntryKey: event.entryKey,
      }));
  }
}

function projectQueryChips(
  contract: ViewContract,
  queryState: WorkbookQueryState,
): readonly WorkbookQueryChip[] {
  return projectWorkbookQueryEntries(contract, queryState).map((entry) => ({
    ...entry,
    label: `${entry.token} ${entry.detail}`,
    removeCommand:
      entry.identity.kind === "group"
        ? { kind: "group_set", fieldKey: null }
        : entry.identity.kind === "sort"
          ? { kind: "sort_remove", fieldKey: entry.identity.fieldKey }
          : { kind: "filter_remove", fieldKey: entry.identity.fieldKey },
  }));
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

function updateTransient(
  state: WorkbookGridControlsTransientState,
  subjectKey: string,
  update: (
    current: WorkbookGridSurfaceTransientState,
  ) => WorkbookGridSurfaceTransientState,
): WorkbookGridControlsTransientState {
  if (state.subjectKey !== subjectKey) return state;
  const next = update(state.value);
  return next === state.value ? state : { ...state, value: next };
}

function fieldLabel(
  contract: ViewContract,
  fieldKey: string | null,
): string | null {
  if (fieldKey === null) return null;
  return contract.fieldMap[fieldKey]?.label ?? fieldKey;
}
