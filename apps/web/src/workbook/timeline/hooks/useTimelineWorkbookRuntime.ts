import { requireViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useState,
} from "react";
import {
  applyFilterDraft,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  replaceWorkbookSort,
  updateGroupBy,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";

const timelineRuntimeContract = requireViewContract(timelineViewSchemaId);

type TimelineWorkbookRuntimeSaveState = "Syncing" | "Saved" | "Conflict";

type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;

export type TimelineWorkbookRuntimeInput = {
  readonly controlledFilterDraft?: FilterDraft | undefined;
  readonly controlledQueryState?: WorkbookQueryState | undefined;
  readonly onFilterDraftChange?: FilterDraftSetter | undefined;
  readonly onQueryStateChange?: WorkbookQueryStateSetter | undefined;
};

function clearAppliedTimelineFilterDraft(current: FilterDraft): FilterDraft {
  return {
    ...current,
    booleanValue: "",
    value: "",
  };
}

function applyTimelineFilterDraftToQuery(
  setQueryState: WorkbookQueryStateSetter,
  setFilterDraft: FilterDraftSetter,
  draft: FilterDraft,
): void {
  setQueryState((current) => applyFilterDraft(current, draft));
  setFilterDraft(clearAppliedTimelineFilterDraft);
}

export function useTimelineWorkbookRuntime({
  controlledFilterDraft,
  controlledQueryState,
  onFilterDraftChange,
  onQueryStateChange,
}: TimelineWorkbookRuntimeInput) {
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [refreshError, setRefreshError] = useState<string | null>(null);
  const [saveState, setSaveState] =
    useState<TimelineWorkbookRuntimeSaveState>("Saved");
  const [saveStateSecondaryMessage, setSaveStateSecondaryMessage] = useState<
    string | null
  >(null);
  const [uncontrolledQueryState, setUncontrolledQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [uncontrolledFilterDraft, setUncontrolledFilterDraft] =
    useState<FilterDraft>(() => defaultFilterDraft(timelineRuntimeContract));

  const queryState = controlledQueryState ?? uncontrolledQueryState;
  const setQueryState = onQueryStateChange ?? setUncontrolledQueryState;
  const filterDraft = controlledFilterDraft ?? uncontrolledFilterDraft;
  const setFilterDraft = onFilterDraftChange ?? setUncontrolledFilterDraft;

  const applyQueryFilter = useCallback(
    (draft: FilterDraft = filterDraft) => {
      applyTimelineFilterDraftToQuery(setQueryState, setFilterDraft, draft);
    },
    [filterDraft, setFilterDraft, setQueryState],
  );

  const handleQueryGroupByChange = useCallback(
    (groupBy: string | null) => {
      setQueryState((current) =>
        updateGroupBy(timelineRuntimeContract, current, groupBy),
      );
    },
    [setQueryState],
  );

  const handleQuerySortChange = useCallback(
    (sort: WorkbookQueryState["sort"]) => {
      setQueryState((current) =>
        replaceWorkbookSort(timelineRuntimeContract, current, sort),
      );
    },
    [setQueryState],
  );

  return {
    lifecycle: {
      isInitialLoading,
      isRefreshing,
      loadError,
      refreshError,
      saveState,
      saveStateSecondaryMessage,
      setIsInitialLoading,
      setIsRefreshing,
      setLoadError,
      setRefreshError,
      setSaveState,
      setSaveStateSecondaryMessage,
    },
    query: {
      applyQueryFilter,
      filterDraft,
      handleQueryGroupByChange,
      handleQuerySortChange,
      queryState,
      setFilterDraft,
      setQueryState,
    },
  };
}
