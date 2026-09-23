import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
} from "react";
import {
  applyFilterDraft,
  clearFilterDraftValue,
  defaultFilterDraft,
  type FilterDraft,
  removeFilterField,
  replaceWorkbookSort,
  updateGroupBy,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { useWorkbookQueryPresentation } from "../query/WorkbookQueryBrowsingContext";
import {
  useWorkbookQueryState,
  type WorkbookQueryStateSetter,
} from "../view-state/useWorkbookQueryState";

type WorkbookActiveQueryControls = {
  readonly contract: ViewContract;
  readonly filterDraft: FilterDraft;
  readonly onApplyFilter: (draft: FilterDraft) => void;
  readonly onClearFilters: () => void;
  readonly onFilterDraftChange: Dispatch<SetStateAction<FilterDraft>>;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly queryState: WorkbookQueryState;
  readonly surface: string;
};

export function useWorkbookQueryController({
  surface,
}: {
  readonly surface: string;
}) {
  const browsing = useWorkbookQueryPresentation();
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
    setFilterDraftForSurface,
    setQueryStateForSurface,
    updateEntry,
  } = useWorkbookQueryState(viewSchemaIds);

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
    (viewSchemaId: string) => {
      const requested = entryFor(viewSchemaId).queryState;
      return (
        browsing.find(viewSchemaId)?.canonicalIntent(requested) ?? requested
      );
    },
    [entryFor, browsing],
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

  const activeContract = useMemo(
    () => workbookContractForViewSchemaId(surface),
    [surface],
  );
  const activeEntry = entryFor(surface);
  useEffect(
    () =>
      browsing.bindRevert(surface, () => {
        const accepted = browsing.find(surface)?.getSnapshot().authored;
        if (accepted) setQueryStateForSurface(surface, accepted);
      }),
    [browsing, surface, setQueryStateForSurface],
  );
  const presentedQuery =
    browsing.find(surface)?.presentationQuery(activeEntry.queryState) ??
    activeEntry.queryState;
  const activeQueryControls = useMemo<WorkbookActiveQueryControls>(() => {
    const setActiveQueryState = makeQuerySetter(surface);
    const setActiveFilterDraft = (action: SetStateAction<FilterDraft>) =>
      setFilterDraftForSurface(surface, action);
    return {
      contract: activeContract,
      filterDraft: activeEntry.filterDraft,
      onApplyFilter: (draft) => {
        setActiveQueryState((current) => applyFilterDraft(current, draft));
        setActiveFilterDraft(clearFilterDraftValue);
      },
      onClearFilters: () => {
        setActiveQueryState((current) =>
          current.filters.length === 0 ? current : { ...current, filters: [] },
        );
        setActiveFilterDraft(defaultFilterDraft(activeContract));
      },
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
      queryState: presentedQuery,
      surface,
    };
  }, [
    activeContract,
    activeEntry,
    makeQuerySetter,
    setFilterDraftForSurface,
    surface,
    presentedQuery,
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
