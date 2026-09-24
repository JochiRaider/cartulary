import { requireViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useReducer,
} from "react";
import {
  type FilterDraft,
  replaceWorkbookSort,
  updateGroupBy,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  initialWorkbookLifecycleState,
  reduceWorkbookLifecycle,
  type WorkbookOperationFeedback,
} from "../../runtime/workbookLifecycleModel";

const timelineRuntimeContract = requireViewContract(timelineViewSchemaId);

type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;

type TimelineWorkbookRuntimeInput = {
  readonly filterDraft: FilterDraft;
  readonly queryState: WorkbookQueryState;
  readonly setFilterDraft: FilterDraftSetter;
  readonly setQueryState: WorkbookQueryStateSetter;
};

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
  const setOperationError = useCallback(
    (value: WorkbookOperationFeedback | null) =>
      dispatchLifecycle({ type: "operation_error", value }),
    [],
  );
  const setMutationError = useCallback(
    (message: string | null) =>
      setOperationError(
        message === null ? null : { family: "mutation", message },
      ),
    [setOperationError],
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
      setOperationError,
      setMutationError,
    },
    query: {
      filterDraft,
      handleQueryGroupByChange,
      handleQuerySortChange,
      queryState,
      setFilterDraft,
      setQueryState,
    },
  };
}
