import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookResolvedLayoutState } from "../layout/workbookColumnLayout";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import type { WorkbookFilter, WorkbookQueryState } from "./workbookQuery";
import type { SavedViewResource } from "./workbookSavedViews";

export const workbookViewBarControlOrder = [
  "saved_view",
  "sort",
  "group",
  "filters",
  "columns",
  "applied_query",
  "inspector",
  "create",
] as const;

export type WorkbookViewBarControlId =
  (typeof workbookViewBarControlOrder)[number];

export type WorkbookViewBarPanelId =
  | "columns"
  | "filters"
  | "group"
  | "saved_view"
  | "sort";

export type WorkbookQueryEntryIdentity =
  | { readonly fieldKey: string; readonly kind: "filter" }
  | { readonly fieldKey: string; readonly kind: "group" }
  | { readonly fieldKey: string; readonly kind: "sort" };

export type WorkbookQueryEntryDescriptor = {
  readonly accessibleName: string;
  readonly detail: string;
  readonly editorPanel: Extract<
    WorkbookViewBarPanelId,
    "filters" | "group" | "sort"
  >;
  readonly identity: WorkbookQueryEntryIdentity;
  readonly key: string;
  readonly token: string;
};

export type WorkbookViewBarAvailableActions = {
  readonly createSavedView: boolean;
  readonly deleteSavedView: boolean;
  readonly duplicateSavedView: boolean;
  readonly resetSavedView: boolean;
  readonly setDefault: boolean;
  readonly setHome: boolean;
  readonly updateSavedView: boolean;
};

export type WorkbookSavedViewPresentationAction =
  | "create"
  | "delete"
  | "duplicate"
  | "reset"
  | "set_default"
  | "set_home"
  | "update";

export type WorkbookSavedViewActionGroup = {
  readonly actions: readonly WorkbookSavedViewPresentationAction[];
  readonly id: "create" | "delete" | "duplicate" | "selected" | "startup";
};

export type WorkbookViewBarTransientState = {
  readonly activeEntryKey: string | null;
  readonly openPanel: WorkbookViewBarPanelId | null;
  readonly subjectKey: string;
};

export type WorkbookViewBarWorkingSetInput = {
  readonly availableActions: WorkbookViewBarAvailableActions;
  readonly chromeMode: WorkbookChromeMode;
  readonly contract: ViewContract;
  readonly createAvailable: boolean;
  readonly incidentId: string;
  readonly inspectorAvailable: boolean;
  readonly isSavedViewDirty: boolean;
  readonly layoutState: WorkbookResolvedLayoutState;
  readonly queryState: WorkbookQueryState;
  readonly savedViewResourceKind:
    | "invalid_selection"
    | "loading"
    | "ready"
    | "unavailable";
  readonly selectedSavedView: SavedViewResource | null;
  readonly transient: WorkbookViewBarTransientState;
  readonly viewSchemaId: string;
};

export type WorkbookViewBarWorkingSet = {
  readonly availableActions: WorkbookViewBarAvailableActions;
  readonly columns: readonly {
    readonly fieldKey: string;
    readonly hidden: boolean;
    readonly label: string;
    readonly position: number;
  }[];
  readonly controlOrder: readonly WorkbookViewBarControlId[];
  readonly createAvailable: boolean;
  readonly filterEntries: readonly WorkbookQueryEntryDescriptor[];
  readonly filterEditor: {
    readonly fieldKey: string | null;
    readonly mode: "add" | "edit";
  };
  readonly focusTargets: {
    readonly entryKeys: readonly string[];
    readonly invokingEntryKey: string | null;
    readonly ownerControlByKind: Readonly<
      Record<WorkbookQueryEntryIdentity["kind"], WorkbookViewBarControlId>
    >;
  };
  readonly groupEntry: WorkbookQueryEntryDescriptor | null;
  readonly groupOptions: readonly {
    readonly fieldKey: string;
    readonly label: string;
  }[];
  readonly hiddenQueryEntries: readonly WorkbookQueryEntryDescriptor[];
  readonly inspectorAvailable: boolean;
  readonly isSavedViewDirty: boolean;
  readonly savedView: {
    readonly accessibleName: string;
    readonly actionGroups: readonly WorkbookSavedViewActionGroup[];
    readonly displayName: string;
    readonly mutable: boolean;
    readonly primaryAction: WorkbookSavedViewPresentationAction | null;
    readonly resourceKind: WorkbookViewBarWorkingSetInput["savedViewResourceKind"];
    readonly savedViewId: string | null;
    readonly scope: SavedViewResource["scope"] | null;
    readonly version: number | null;
  };
  readonly sortEntries: readonly (WorkbookQueryEntryDescriptor & {
    readonly direction: "asc" | "desc";
    readonly priority: number;
  })[];
  readonly subjectKey: string;
  readonly transient: WorkbookViewBarTransientState;
  readonly unusedSortFields: readonly {
    readonly fieldKey: string;
    readonly label: string;
  }[];
  readonly visibleQueryEntries: readonly WorkbookQueryEntryDescriptor[];
  readonly visibleQueryEntryCapacity: number;
};

export function workbookViewBarSubjectKey({
  incidentId,
  selectedSavedView,
  viewSchemaId,
}: Pick<
  WorkbookViewBarWorkingSetInput,
  "incidentId" | "selectedSavedView" | "viewSchemaId"
>): string {
  return selectedSavedView === null
    ? `${incidentId}:${viewSchemaId}:base`
    : `${incidentId}:${viewSchemaId}:saved_view:${selectedSavedView.saved_view_id}:${selectedSavedView.saved_view_version}`;
}

export function projectWorkbookViewBarWorkingSet(
  input: WorkbookViewBarWorkingSetInput,
): WorkbookViewBarWorkingSet {
  const subjectKey = workbookViewBarSubjectKey(input);
  const queryEntries = projectWorkbookQueryEntries(
    input.contract,
    input.queryState,
  );
  const visibleQueryEntryCapacity = queryEntryCapacity(input.chromeMode);
  const selectedSortFields = new Set(
    input.queryState.sort.map((entry) => entry.fieldKey),
  );
  const hiddenColumns = new Set(input.layoutState.hiddenFieldKeys);
  const sortEntries = queryEntries.filter(
    (entry): entry is WorkbookViewBarWorkingSet["sortEntries"][number] =>
      entry.identity.kind === "sort" && "priority" in entry,
  );
  const activeEntry = queryEntries.find(
    (entry) => entry.key === input.transient.activeEntryKey,
  );
  const actionGroups = projectSavedViewActionGroups(input.availableActions);
  const primaryAction = input.availableActions.updateSavedView
    ? "update"
    : input.availableActions.createSavedView
      ? "create"
      : null;
  const displayName = input.selectedSavedView?.display_name ?? "Unsaved view";

  return {
    availableActions: input.availableActions,
    columns: input.layoutState.columnOrder.flatMap((fieldKey, position) => {
      const field = input.contract.fieldMap[fieldKey];
      return field === undefined
        ? []
        : [
            {
              fieldKey,
              hidden: hiddenColumns.has(fieldKey),
              label: field.label,
              position,
            },
          ];
    }),
    controlOrder: workbookViewBarControlOrder,
    createAvailable: input.createAvailable,
    filterEntries: queryEntries.filter(
      (entry) => entry.identity.kind === "filter",
    ),
    filterEditor: {
      fieldKey:
        input.transient.openPanel === "filters" &&
        activeEntry?.identity.kind === "filter"
          ? activeEntry.identity.fieldKey
          : null,
      mode:
        input.transient.openPanel === "filters" &&
        activeEntry?.identity.kind === "filter"
          ? "edit"
          : "add",
    },
    focusTargets: {
      entryKeys: queryEntries.map((entry) => entry.key),
      invokingEntryKey: activeEntry?.key ?? null,
      ownerControlByKind: {
        filter: "filters",
        group: "group",
        sort: "sort",
      },
    },
    groupEntry:
      queryEntries.find((entry) => entry.identity.kind === "group") ?? null,
    groupOptions: input.contract.groupingFields.map((fieldKey) => ({
      fieldKey,
      label: fieldLabel(input.contract, fieldKey),
    })),
    hiddenQueryEntries: queryEntries.slice(visibleQueryEntryCapacity),
    inspectorAvailable: input.inspectorAvailable,
    isSavedViewDirty: input.isSavedViewDirty,
    savedView: {
      accessibleName: `${displayName}${input.isSavedViewDirty ? ", Modified" : ""}`,
      actionGroups,
      displayName,
      mutable:
        input.selectedSavedView !== null &&
        input.availableActions.updateSavedView,
      primaryAction,
      resourceKind: input.savedViewResourceKind,
      savedViewId: input.selectedSavedView?.saved_view_id ?? null,
      scope: input.selectedSavedView?.scope ?? null,
      version: input.selectedSavedView?.saved_view_version ?? null,
    },
    sortEntries,
    subjectKey,
    transient:
      input.transient.subjectKey === subjectKey
        ? input.transient
        : { activeEntryKey: null, openPanel: null, subjectKey },
    unusedSortFields: input.contract.fields.flatMap((field) =>
      input.contract.sortableFieldMap[field.fieldKey] &&
      !selectedSortFields.has(field.fieldKey)
        ? [{ fieldKey: field.fieldKey, label: field.label }]
        : [],
    ),
    visibleQueryEntries: queryEntries.slice(0, visibleQueryEntryCapacity),
    visibleQueryEntryCapacity,
  };
}

function projectSavedViewActionGroups(
  actions: WorkbookViewBarAvailableActions,
): readonly WorkbookSavedViewActionGroup[] {
  return [
    {
      actions: actions.createSavedView ? (["create"] as const) : [],
      id: "create" as const,
    },
    {
      actions: [
        ...(actions.updateSavedView ? (["update"] as const) : []),
        ...(actions.resetSavedView ? (["reset"] as const) : []),
      ],
      id: "selected" as const,
    },
    {
      actions: actions.duplicateSavedView ? (["duplicate"] as const) : [],
      id: "duplicate" as const,
    },
    {
      actions: [
        ...(actions.setHome ? (["set_home"] as const) : []),
        ...(actions.setDefault ? (["set_default"] as const) : []),
      ],
      id: "startup" as const,
    },
    {
      actions: actions.deleteSavedView ? (["delete"] as const) : [],
      id: "delete" as const,
    },
  ].filter((group) => group.actions.length > 0);
}

export function projectWorkbookQueryEntries(
  contract: ViewContract,
  queryState: WorkbookQueryState,
): readonly (
  | WorkbookQueryEntryDescriptor
  | (WorkbookQueryEntryDescriptor & {
      readonly direction: "asc" | "desc";
      readonly priority: number;
    })
)[] {
  const entries: (
    | WorkbookQueryEntryDescriptor
    | (WorkbookQueryEntryDescriptor & {
        readonly direction: "asc" | "desc";
        readonly priority: number;
      })
  )[] = [];
  if (queryState.groupBy !== null) {
    const label = fieldLabel(contract, queryState.groupBy);
    entries.push({
      accessibleName: `Group by ${label}`,
      detail: label,
      editorPanel: "group",
      identity: { fieldKey: queryState.groupBy, kind: "group" },
      key: `group:${queryState.groupBy}`,
      token: "G",
    });
  }
  queryState.sort.forEach((sort, index) => {
    const priority = index + 1;
    const label = fieldLabel(contract, sort.fieldKey);
    const directionLabel =
      sort.direction === "asc" ? "ascending" : "descending";
    entries.push({
      accessibleName: `Sort ${priority}, ${label}, ${directionLabel}`,
      detail: `${label} ${sort.direction === "asc" ? "↑" : "↓"}`,
      direction: sort.direction,
      editorPanel: "sort",
      identity: { fieldKey: sort.fieldKey, kind: "sort" },
      key: `sort:${sort.fieldKey}`,
      priority,
      token: `S${priority}`,
    });
  });
  queryState.filters.forEach((filter, index) => {
    const position = index + 1;
    const label = fieldLabel(contract, filter.fieldKey);
    const description = describeFilter(filter);
    entries.push({
      accessibleName: `Filter ${position}, ${label}, ${description}`,
      detail: `${label} ${description}`,
      editorPanel: "filters",
      identity: { fieldKey: filter.fieldKey, kind: "filter" },
      key: `filter:${filter.fieldKey}`,
      token: `F${position}`,
    });
  });
  return entries;
}

export function queryEntryCapacity(chromeMode: WorkbookChromeMode): number {
  if (chromeMode === "base") return 3;
  if (chromeMode === "narrow_desktop") return 2;
  return 0;
}

export function describeFilter(filter: WorkbookFilter): string {
  const value = describeFilterArgument(filter);
  const operator =
    {
      contains_all: "contains all",
      contains_any: "contains any",
      eq: "equals",
      full_text: "contains tokens",
      prefix: "starts with",
      range: "is in range",
    }[filter.op] ?? filter.op;
  return value === "" ? operator : `${operator} ${value}`;
}

function describeFilterArgument(filter: WorkbookFilter): string {
  if (Array.isArray(filter.arg.values)) return filter.arg.values.join(", ");
  if (filter.arg.value === null) return "empty";
  if (
    typeof filter.arg.value === "string" ||
    typeof filter.arg.value === "number" ||
    typeof filter.arg.value === "boolean"
  ) {
    return String(filter.arg.value);
  }
  if (typeof filter.arg.query === "string") return filter.arg.query;
  return (["gt", "gte", "lt", "lte"] as const)
    .flatMap((bound) =>
      typeof filter.arg[bound] === "string" ||
      typeof filter.arg[bound] === "number"
        ? [`${bound} ${String(filter.arg[bound])}`]
        : [],
    )
    .join(" ");
}

function fieldLabel(contract: ViewContract, fieldKey: string): string {
  return contract.fieldMap[fieldKey]?.label ?? fieldKey;
}
