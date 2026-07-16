import type { WorkbookSurface } from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  applyFilterDraft,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  removeFilterField,
  replaceWorkbookSort,
  updateGroupBy,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import type { WorkbookSheetRef } from "../models/workbookStartup";
import {
  workbookContractForViewSchemaId,
  workbookQuerySurfaceSlot,
} from "../models/workbookSurfaceQueryRuntime";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";

type WorkbookQueryEntry = {
  readonly filterDraft: FilterDraft;
  readonly queryState: WorkbookQueryState;
};

const defaultEntryByViewSchemaId = new Map<string, WorkbookQueryEntry>();

export type WorkbookQueryStateSetter = Dispatch<
  SetStateAction<WorkbookQueryState>
>;

export type WorkbookActiveQueryControls = {
  readonly contract: ViewContract;
  readonly filterDraft: FilterDraft;
  readonly onApplyFilter: (draft: FilterDraft) => void;
  readonly onClearAll: () => void;
  readonly onFilterDraftChange: Dispatch<SetStateAction<FilterDraft>>;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly queryState: WorkbookQueryState;
  readonly surface: WorkbookSurface;
};

function newQueryEntry(viewSchemaId: string): WorkbookQueryEntry {
  return {
    filterDraft: defaultFilterDraft(
      workbookContractForViewSchemaId(viewSchemaId),
    ),
    queryState: emptyWorkbookQueryState(),
  };
}

function defaultQueryEntry(viewSchemaId: string): WorkbookQueryEntry {
  const existing = defaultEntryByViewSchemaId.get(viewSchemaId);
  if (existing !== undefined) {
    return existing;
  }
  const created = newQueryEntry(viewSchemaId);
  defaultEntryByViewSchemaId.set(viewSchemaId, created);
  return created;
}

function resolveStateAction<T>(current: T, action: SetStateAction<T>): T {
  return typeof action === "function"
    ? (action as (current: T) => T)(current)
    : action;
}

function clearAppliedFilterDraft(current: FilterDraft): FilterDraft {
  return {
    ...current,
    booleanValue: "",
    value: "",
  };
}

export function useWorkbookQueryController({
  startupSheetRef,
  surface,
}: {
  readonly startupSheetRef: WorkbookSheetRef;
  readonly surface: string;
}) {
  const [entries, setEntries] = useState<
    Readonly<Record<string, WorkbookQueryEntry>>
  >(() => ({}));

  const entryFor = useCallback(
    (viewSchemaId: string) =>
      entries[viewSchemaId] ?? defaultQueryEntry(viewSchemaId),
    [entries],
  );

  const updateEntry = useCallback(
    (
      viewSchemaId: string,
      update: (current: WorkbookQueryEntry) => WorkbookQueryEntry,
    ) => {
      setEntries((current) => {
        const previous =
          current[viewSchemaId] ?? defaultQueryEntry(viewSchemaId);
        return { ...current, [viewSchemaId]: update(previous) };
      });
    },
    [],
  );

  const setQueryStateForSurface = useCallback(
    (viewSchemaId: string, action: SetStateAction<WorkbookQueryState>) => {
      updateEntry(viewSchemaId, (current) => ({
        ...current,
        queryState: resolveStateAction(current.queryState, action),
      }));
    },
    [updateEntry],
  );

  const setFilterDraftForSurface = useCallback(
    (viewSchemaId: string, action: SetStateAction<FilterDraft>) => {
      updateEntry(viewSchemaId, (current) => ({
        ...current,
        filterDraft: resolveStateAction(current.filterDraft, action),
      }));
    },
    [updateEntry],
  );

  const resetSurfaceQuery = useCallback((viewSchemaId: string) => {
    setEntries((current) => ({
      ...current,
      [viewSchemaId]: newQueryEntry(viewSchemaId),
    }));
  }, []);

  const applyQueryStateForSurface = useCallback(
    (viewSchemaId: string, queryState: WorkbookQueryState) => {
      updateEntry(viewSchemaId, () => ({
        filterDraft: defaultFilterDraft(
          workbookContractForViewSchemaId(viewSchemaId),
        ),
        queryState,
      }));
    },
    [updateEntry],
  );

  const currentQueryStateForSurface = useCallback(
    (viewSchemaId: string) => entryFor(viewSchemaId).queryState,
    [entryFor],
  );

  const makeQuerySetter = useCallback(
    (viewSchemaId: string): WorkbookQueryStateSetter =>
      (action) =>
        setQueryStateForSurface(viewSchemaId, action),
    [setQueryStateForSurface],
  );

  const setTimelineQueryState = useMemo(
    () => makeQuerySetter(timelineViewSchemaId),
    [makeQuerySetter],
  );
  const setHostQueryState = useMemo(
    () => makeQuerySetter(hostsViewSchemaId),
    [makeQuerySetter],
  );
  const setIdentityQueryState = useMemo(
    () => makeQuerySetter(identitiesViewSchemaId),
    [makeQuerySetter],
  );
  const setAssessmentQueryState = useMemo(
    () => makeQuerySetter(assessmentsViewSchemaId),
    [makeQuerySetter],
  );
  const setGenericQueryState = useMemo(
    () => makeQuerySetter(surface),
    [makeQuerySetter, surface],
  );

  useEffect(() => {
    if (
      startupSheetRef.kind === "view_schema" &&
      workbookQuerySurfaceSlot(surface) === "generic"
    ) {
      resetSurfaceQuery(surface);
    }
  }, [resetSurfaceQuery, startupSheetRef.kind, surface]);

  const activeContract = useMemo(
    () => workbookContractForViewSchemaId(surface),
    [surface],
  );
  const activeEntry = entryFor(surface);
  const activeQueryControls = useMemo<WorkbookActiveQueryControls>(() => {
    const setActiveQueryState = makeQuerySetter(surface);
    const setActiveFilterDraft = (action: SetStateAction<FilterDraft>) =>
      setFilterDraftForSurface(surface, action);
    return {
      contract: activeContract,
      filterDraft: activeEntry.filterDraft,
      onApplyFilter: (draft) => {
        setActiveQueryState((current) => applyFilterDraft(current, draft));
        setActiveFilterDraft(clearAppliedFilterDraft);
      },
      onClearAll: () => resetSurfaceQuery(surface),
      onFilterDraftChange: setActiveFilterDraft,
      onGroupByChange: (groupBy) => {
        setActiveQueryState((current) =>
          updateGroupBy(activeContract, current, groupBy),
        );
      },
      onRemoveFilter: (fieldKey) => {
        setActiveQueryState((current) => removeFilterField(current, fieldKey));
      },
      onSortChange: (sort) => {
        setActiveQueryState((current) =>
          replaceWorkbookSort(activeContract, current, sort),
        );
      },
      queryState: activeEntry.queryState,
      surface: surface as WorkbookSurface,
    };
  }, [
    activeContract,
    activeEntry,
    makeQuerySetter,
    resetSurfaceQuery,
    setFilterDraftForSurface,
    surface,
  ]);

  return {
    commands: {
      applyQueryStateForSurface,
      currentQueryStateForSurface,
      setAssessmentQueryState,
      setGenericQueryState,
      setHostQueryState,
      setIdentityQueryState,
      setTimelineQueryState,
    },
    snapshot: {
      activeContract,
      activeQueryControls,
      assessmentQueryState: entryFor(assessmentsViewSchemaId).queryState,
      genericQueryState: activeEntry.queryState,
      hostQueryState: entryFor(hostsViewSchemaId).queryState,
      identityQueryState: entryFor(identitiesViewSchemaId).queryState,
      timelineQueryState: entryFor(timelineViewSchemaId).queryState,
    },
  };
}
