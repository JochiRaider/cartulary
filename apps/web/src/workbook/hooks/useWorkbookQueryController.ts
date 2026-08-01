import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import {
  applyFilterDraft,
  defaultFilterDraft,
  type FilterDraft,
  removeFilterField,
  replaceWorkbookSort,
  updateGroupBy,
  type WorkbookQueryState,
} from "../models/workbookQuery";
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
import {
  useWorkbookQueryState,
  type WorkbookQueryStateSetter,
} from "../view-state/useWorkbookQueryState";

type WorkbookActiveQueryControls = {
  readonly contract: ViewContract;
  readonly filterDraft: FilterDraft;
  readonly onApplyFilter: (draft: FilterDraft) => void;
  readonly onClearAll: () => void;
  readonly onFilterDraftChange: Dispatch<SetStateAction<FilterDraft>>;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly queryState: WorkbookQueryState;
  readonly surface: string;
};

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
  readonly startupSheetRef: SheetRef;
  readonly surface: string;
}) {
  const viewSchemaIds = useMemo(
    () => [
      assessmentsViewSchemaId,
      hostsViewSchemaId,
      identitiesViewSchemaId,
      surface,
      timelineViewSchemaId,
    ],
    [surface],
  );
  const {
    entryFor,
    resetEntry,
    setFilterDraftForSurface,
    setQueryStateForSurface,
    updateEntry,
  } = useWorkbookQueryState(viewSchemaIds);

  const resetSurfaceQuery = useCallback(
    (viewSchemaId: string) => {
      resetEntry(viewSchemaId);
    },
    [resetEntry],
  );

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
      surface,
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
