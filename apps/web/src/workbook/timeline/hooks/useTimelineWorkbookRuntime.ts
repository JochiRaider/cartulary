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
  toggleSortField,
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

  const handleQuerySortToggle = useCallback(
    (fieldKey: string) => {
      setQueryState((current) =>
        toggleSortField(timelineRuntimeContract, current, fieldKey),
      );
    },
    [setQueryState],
  );

  return {
    lifecycle: {
      isInitialLoading,
      loadError,
      refreshError,
      saveState,
      saveStateSecondaryMessage,
      setIsInitialLoading,
      setLoadError,
      setRefreshError,
      setSaveState,
      setSaveStateSecondaryMessage,
    },
    query: {
      applyQueryFilter,
      filterDraft,
      handleQueryGroupByChange,
      handleQuerySortToggle,
      queryState,
      setFilterDraft,
      setQueryState,
    },
  };
}
