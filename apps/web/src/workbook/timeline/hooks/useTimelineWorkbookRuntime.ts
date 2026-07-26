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

const timelineRuntimeContract = requireViewContract(timelineViewSchemaId);

export type TimelineWorkbookRuntimeSaveState = "Syncing" | "Saved" | "Conflict";

type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;

export type TimelineWorkbookRuntimeInput = {
  readonly filterDraft: FilterDraft;
  readonly queryState: WorkbookQueryState;
  readonly setFilterDraft: FilterDraftSetter;
  readonly setQueryState: WorkbookQueryStateSetter;
};

export type TimelineWorkbookLifecycleState = {
  readonly isInitialLoading: boolean;
  readonly isRefreshing: boolean;
  readonly loadError: string | null;
  readonly refreshError: string | null;
  readonly saveState: TimelineWorkbookRuntimeSaveState;
  readonly saveStateSecondaryMessage: string | null;
};

export type TimelineWorkbookLifecycleAction =
  | { readonly type: "initial_loading"; readonly value: boolean }
  | { readonly type: "refreshing"; readonly value: boolean }
  | { readonly type: "load_error"; readonly value: string | null }
  | { readonly type: "refresh_error"; readonly value: string | null }
  | {
      readonly type: "save_state";
      readonly value: TimelineWorkbookRuntimeSaveState;
    }
  | {
      readonly type: "save_state_secondary_message";
      readonly value: string | null;
    };

export const initialTimelineWorkbookLifecycleState: TimelineWorkbookLifecycleState =
  {
    isInitialLoading: true,
    isRefreshing: false,
    loadError: null,
    refreshError: null,
    saveState: "Saved",
    saveStateSecondaryMessage: null,
  };

export function reduceTimelineWorkbookLifecycle(
  state: TimelineWorkbookLifecycleState,
  action: TimelineWorkbookLifecycleAction,
): TimelineWorkbookLifecycleState {
  switch (action.type) {
    case "initial_loading":
      return { ...state, isInitialLoading: action.value };
    case "refreshing":
      return { ...state, isRefreshing: action.value };
    case "load_error":
      return { ...state, loadError: action.value };
    case "refresh_error":
      return { ...state, refreshError: action.value };
    case "save_state":
      return { ...state, saveState: action.value };
    case "save_state_secondary_message":
      return { ...state, saveStateSecondaryMessage: action.value };
  }
}

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
    reduceTimelineWorkbookLifecycle,
    initialTimelineWorkbookLifecycleState,
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
  const setSaveState = useCallback(
    (value: TimelineWorkbookRuntimeSaveState) =>
      dispatchLifecycle({ type: "save_state", value }),
    [],
  );
  const setSaveStateSecondaryMessage = useCallback(
    (value: string | null) =>
      dispatchLifecycle({
        type: "save_state_secondary_message",
        value,
      }),
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
