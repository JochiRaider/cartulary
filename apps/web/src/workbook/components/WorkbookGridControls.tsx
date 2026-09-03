import {
  gridFilterChipTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { useEffect, useMemo, useReducer } from "react";
import type { WorkbookResolvedLayoutState } from "../layout/workbookColumnLayout";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import {
  applyWorkbookSortCommand,
  createWorkbookGridControlsTransientState,
  projectWorkbookGridQueryControls,
  reduceWorkbookGridControlsTransientState,
  type WorkbookGridQueryCommand,
  workbookGridSurfaceTransientState,
} from "../models/workbookGridQueryControls";
import type { FilterDraft, WorkbookQueryState } from "../models/workbookQuery";
import { WorkbookActiveQueryChips } from "./WorkbookActiveQueryChips";
import { WorkbookColumnsControl } from "./WorkbookColumnsControl";
import { WorkbookFiltersControl } from "./WorkbookFiltersControl";
import { WorkbookGroupControl } from "./WorkbookGroupControl";
import { WorkbookSortControl } from "./WorkbookSortControl";

type WorkbookGridControlsProps = {
  readonly chromeMode?: WorkbookChromeMode | undefined;
  readonly contract: ViewContract;
  readonly defaultFilterPopoverOpen?: boolean | undefined;
  readonly filterDraft: FilterDraft;
  readonly layoutState: WorkbookResolvedLayoutState;
  readonly onApplyFilter: (draft: FilterDraft) => void;
  readonly onClearAll?: (() => void) | undefined;
  readonly onFilterDraftChange: (draft: FilterDraft) => void;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onColumnHiddenChange: (fieldKey: string, hidden: boolean) => void;
  readonly onColumnMove: (
    fieldKey: string,
    direction: "earlier" | "later",
  ) => void;
  readonly onResetColumns: () => void;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly queryState: WorkbookQueryState;
  readonly surface: string;
};

export function WorkbookGridControls({
  chromeMode = "base",
  contract,
  defaultFilterPopoverOpen = false,
  filterDraft,
  layoutState,
  onApplyFilter,
  onClearAll,
  onFilterDraftChange,
  onGroupByChange,
  onColumnHiddenChange,
  onColumnMove,
  onResetColumns,
  onRemoveFilter,
  onSortChange,
  queryState,
  surface,
}: WorkbookGridControlsProps) {
  const [transientState, dispatch] = useReducer(
    reduceWorkbookGridControlsTransientState,
    createWorkbookGridControlsTransientState(
      surface,
      filterDraft,
      defaultFilterPopoverOpen,
    ),
  );
  const surfaceState = workbookGridSurfaceTransientState(
    transientState,
    surface,
    filterDraft,
  );
  const projection = useMemo(
    () =>
      projectWorkbookGridQueryControls({
        chromeMode,
        contract,
        filterChipTestId: (fieldKey) => gridFilterChipTestId(surface, fieldKey),
        layoutState,
        queryState,
      }),
    [chromeMode, contract, layoutState, queryState, surface],
  );
  const commandPorts: WorkbookGridCommandPorts = {
    contract,
    onClearAll,
    onColumnHiddenChange,
    onColumnMove,
    onGroupByChange,
    onRemoveFilter,
    onResetColumns,
    onSortChange,
    queryState,
  };
  const onCommand = (command: WorkbookGridQueryCommand) => {
    executeWorkbookGridQueryCommand(command, commandPorts);
  };
  const closePanel = () => {
    dispatch({ type: "close_panel", surface });
  };
  const closeFilterPanel = () => {
    dispatch({ type: "change_filter_draft", surface, filterDraft });
    closePanel();
  };

  useEffect(() => {
    dispatch({
      type: "activate_surface",
      filterDraft,
      surface,
    });
  }, [filterDraft, surface]);

  useEffect(() => {
    dispatch({ type: "sync_filter_draft", filterDraft, surface });
  }, [filterDraft, surface]);

  return (
    <fieldset
      aria-label="Workbook query controls"
      data-hidden-query-chip-count={projection.hiddenChips.length}
      data-query-chip-capacity={projection.visibleChipCapacity}
      data-testid={workbookViewBarQueryControlsTestId(surface)}
      style={queryControlsStyleFor(
        chromeMode,
        projection.hiddenChips.length,
        projection.visibleChipCapacity,
      )}
    >
      <WorkbookSortControl
        constrained={chromeMode !== "base" || projection.hiddenChips.length > 0}
        isOpen={surfaceState.openPanel === "sort"}
        onClose={closePanel}
        onCommand={onCommand}
        onToggle={() => {
          dispatch({ type: "toggle_panel", panel: "sort", surface });
        }}
        projection={projection}
        surface={surface}
      />
      <WorkbookGroupControl
        compact={chromeMode !== "base"}
        constrained={chromeMode !== "base" || projection.hiddenChips.length > 0}
        onCommand={onCommand}
        projection={projection}
        selectedFieldKey={queryState.groupBy}
        surface={surface}
      />
      <WorkbookFiltersControl
        contract={contract}
        draft={surfaceState.filterDraft}
        filterCount={queryState.filters.length}
        isOpen={surfaceState.openPanel === "filters"}
        onApply={(draft) => {
          onFilterDraftChange(draft);
          onApplyFilter(draft);
          dispatch({
            type: "complete_filter",
            filterDraft: clearAppliedDraftValue(draft),
            surface,
          });
        }}
        onChangeDraft={(draft) => {
          dispatch({
            type: "change_filter_draft",
            filterDraft: draft,
            surface,
          });
        }}
        onClose={closeFilterPanel}
        onCommand={onCommand}
        onToggle={() => {
          if (surfaceState.openPanel !== "filters") {
            dispatch({ type: "change_filter_draft", surface, filterDraft });
          }
          dispatch({ type: "toggle_panel", panel: "filters", surface });
        }}
        projection={projection}
        surface={surface}
      />
      <WorkbookColumnsControl
        isOpen={surfaceState.openPanel === "columns"}
        onClose={closePanel}
        onCommand={onCommand}
        onToggle={() => {
          dispatch({ type: "toggle_panel", panel: "columns", surface });
        }}
        projection={projection}
        surface={surface}
      />
      <WorkbookActiveQueryChips
        condensed={chromeMode !== "base"}
        onCommand={onCommand}
        projection={projection}
      />
    </fieldset>
  );
}

type WorkbookGridCommandPorts = Pick<
  WorkbookGridControlsProps,
  | "contract"
  | "onClearAll"
  | "onColumnHiddenChange"
  | "onColumnMove"
  | "onGroupByChange"
  | "onRemoveFilter"
  | "onResetColumns"
  | "onSortChange"
  | "queryState"
>;

function executeWorkbookGridQueryCommand(
  command: WorkbookGridQueryCommand,
  ports: WorkbookGridCommandPorts,
) {
  switch (command.kind) {
    case "sort_add":
    case "sort_set_direction":
    case "sort_move":
    case "sort_remove": {
      const next = applyWorkbookSortCommand(
        ports.contract,
        ports.queryState.sort,
        command,
      );
      if (next !== ports.queryState.sort) ports.onSortChange(next);
      return;
    }
    case "group_set":
      ports.onGroupByChange(command.fieldKey);
      return;
    case "filter_remove":
      ports.onRemoveFilter(command.fieldKey);
      return;
    case "query_clear":
      ports.onClearAll?.();
      return;
    case "column_set_hidden":
      ports.onColumnHiddenChange(command.fieldKey, command.hidden);
      return;
    case "column_move":
      ports.onColumnMove(command.fieldKey, command.direction);
      return;
    case "columns_reset":
      ports.onResetColumns();
      return;
  }
}

function clearAppliedDraftValue(current: FilterDraft): FilterDraft {
  return { ...current, booleanValue: "", value: "" };
}

function queryControlsStyleFor(
  chromeMode: WorkbookChromeMode,
  hiddenChipCount: number,
  visibleChipCapacity: number,
) {
  return {
    ...queryControlsStyle,
    ...(chromeMode === "base" && hiddenChipCount > 0
      ? constrainedBaseQueryControlsStyle
      : null),
    ...(chromeMode === "base" ? null : condensedQueryControlsStyle),
    ...(visibleChipCapacity === 0 ? compactQueryControlsStyle : null),
  };
}

const queryControlsStyle = {
  display: "grid",
  gridTemplateColumns:
    "max-content max-content minmax(0, max-content) max-content minmax(5.5rem, 1fr)",
  alignItems: "center",
  gap: "0.35rem",
  inlineSize: "100%",
  maxInlineSize: "100%",
  boxSizing: "border-box" as const,
  border: 0,
  margin: 0,
  minWidth: 0,
  minInlineSize: 0,
  padding: 0,
  flex: "1 1 0",
  overflow: "visible",
};
const constrainedBaseQueryControlsStyle = {
  gridTemplateColumns: "5.5rem 8rem max-content max-content minmax(0, 1fr)",
};
const condensedQueryControlsStyle = {
  columnGap: "0.25rem",
  gridTemplateColumns: "3.75rem 5.75rem max-content max-content minmax(0, 1fr)",
};
const compactQueryControlsStyle = {
  columnGap: "0.05rem",
  gridTemplateColumns: "3.75rem 4.5rem max-content max-content 0",
};
