import { requireViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useReducer,
} from "react";
import {
  applyFilterDraft,
  type FilterDraft,
  replaceWorkbookSort,
  updateGroupBy,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  initialWorkbookLifecycleState,
  reduceWorkbookLifecycle,
} from "../../runtime/workbookLifecycleModel";

const timelineRuntimeContract = requireViewContract(timelineViewSchemaId);

type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;

export type TimelineWorkbookRuntimeInput = {
  readonly filterDraft: FilterDraft;
  readonly queryState: WorkbookQueryState;
  readonly setFilterDraft: FilterDraftSetter;
  readonly setQueryState: WorkbookQueryStateSetter;
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
  filterDraft,
  queryState,
  setFilterDraft,
  setQueryState,
}: TimelineWorkbookRuntimeInput) {
  const [lifecycle, dispatchLifecycle] = useReducer(
    reduceWorkbookLifecycle,
    initialWorkbookLifecycleState,
  );
  const setIsInitialLoading = useCallback(
    (value: boolean) => dispatchLifecycle({ type: "initial_loading", value }),
    [],
  );
  const setIsRefreshing = useCallback(
    (value: boolean) => dispatchLifecycle({ type: "refreshing", value }),
    [],
  );
  const setLoadError = useCallback(
    (value: string | null) => dispatchLifecycle({ type: "load_error", value }),
    [],
  );
  const setRefreshError = useCallback(
    (value: string | null) =>
      dispatchLifecycle({ type: "refresh_error", value }),
    [],
  );
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
      ...lifecycle,
      setIsInitialLoading,
      setIsRefreshing,
      setLoadError,
      setRefreshError,
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
