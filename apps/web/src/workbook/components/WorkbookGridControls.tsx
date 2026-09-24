import { workbookViewBarQueryControlsTestId } from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { useEffect, useMemo, useReducer, useRef } from "react";
import type {
  WorkbookColumnSizingControls,
  WorkbookFrozenColumnControls,
} from "../layout/WorkbookColumnLayoutController";
import { WorkbookQuerySummarySlot } from "../layout/WorkbookQuerySummarySlot";
import type { WorkbookResolvedLayoutState } from "../layout/workbookColumnLayout";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import {
  applyWorkbookSortCommand,
  createWorkbookGridControlsTransientState,
  projectWorkbookGridQueryControls,
  reduceWorkbookGridControlsTransientState,
  type WorkbookGridQueryCommand,
  type WorkbookGridQueryControlProjection,
  workbookGridSurfaceTransientState,
} from "../models/workbookGridQueryControls";
import {
  clearFilterDraftValue,
  type FilterDraft,
  filterDraftFromFilter,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { WorkbookActiveQueryChips } from "./WorkbookActiveQueryChips";
import { WorkbookColumnsControl } from "./WorkbookColumnsControl";
import { WorkbookFiltersControl } from "./WorkbookFiltersControl";
import { WorkbookGroupControl } from "./WorkbookGroupControl";
import { WorkbookSortControl } from "./WorkbookSortControl";

export type WorkbookGridControlsProps = {
  readonly chromeMode?: WorkbookChromeMode | undefined;
  readonly contract: ViewContract;
  readonly defaultFilterPopoverOpen?: boolean | undefined;
  readonly filterDraft: FilterDraft;
  readonly layoutState: WorkbookResolvedLayoutState;
  readonly onApplyFilter: (draft: FilterDraft) => void;
  readonly onClearFilters?: (() => void) | undefined;
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
  readonly requestedSort?: WorkbookQueryState["sort"] | undefined;
  readonly subjectKey?: string | undefined;
  readonly surface: string;
  readonly sizing: WorkbookColumnSizingControls;
  readonly freezing: WorkbookFrozenColumnControls;
};

export function WorkbookGridControls({
  chromeMode = "base",
  contract,
  defaultFilterPopoverOpen = false,
  filterDraft,
  layoutState,
  onApplyFilter,
  onClearFilters,
  onFilterDraftChange,
  onGroupByChange,
  onColumnHiddenChange,
  onColumnMove,
  onResetColumns,
  onRemoveFilter,
  onSortChange,
  queryState,
  requestedSort,
  subjectKey: suppliedSubjectKey,
  surface,
  sizing,
  freezing,
}: WorkbookGridControlsProps) {
  const subjectKey = suppliedSubjectKey ?? surface;
  const queryEntryRefs = useRef(new Map<string, HTMLButtonElement>());
  const queryEntryReturnFocusRef = useRef<HTMLElement | null>(null);
  const sortTriggerRef = useRef<HTMLButtonElement>(null);
  const groupTriggerRef = useRef<HTMLSelectElement>(null);
  const filterTriggerRef = useRef<HTMLButtonElement>(null);
  const [transientState, dispatch] = useReducer(
    reduceWorkbookGridControlsTransientState,
    createWorkbookGridControlsTransientState(
      subjectKey,
      filterDraft,
      defaultFilterPopoverOpen,
    ),
  );
  const surfaceState = workbookGridSurfaceTransientState(
    transientState,
    subjectKey,
    filterDraft,
  );
  const projection = useMemo(
    () =>
      projectWorkbookGridQueryControls({
        chromeMode,
        contract,
        layoutState,
        queryState,
      }),
    [chromeMode, contract, layoutState, queryState],
  );
  const editorSort = requestedSort ?? queryState.sort;
  const sortEditorProjection = useMemo(
    () =>
      editorSort === queryState.sort
        ? projection
        : projectWorkbookGridQueryControls({
            chromeMode,
            contract,
            layoutState,
            queryState: { ...queryState, sort: editorSort },
          }),
    [chromeMode, contract, editorSort, layoutState, projection, queryState],
  );
  const sortUnapplied =
    editorSort.length !== queryState.sort.length ||
    editorSort.some(
      (entry, index) =>
        entry.fieldKey !== queryState.sort[index]?.fieldKey ||
        entry.direction !== queryState.sort[index]?.direction,
    );
  const commandPorts: WorkbookGridCommandPorts = {
    contract,
    onClearFilters,
    freezing,
    onColumnHiddenChange,
    onColumnMove,
    onGroupByChange,
    onRemoveFilter,
    onResetColumns,
    onSortChange,
    queryState,
    requestedSort: editorSort,
  };
  const onCommand = (command: WorkbookGridQueryCommand) => {
    executeWorkbookGridQueryCommand(command, commandPorts);
  };
  const closePanel = () => {
    dispatch({ type: "close_panel", subjectKey });
  };
  const closeFilterPanel = () => {
    dispatch({ type: "change_filter_draft", subjectKey, filterDraft });
    closePanel();
  };

  useEffect(() => {
    dispatch({
      type: "activate_subject",
      filterDraft,
      subjectKey,
    });
  }, [filterDraft, subjectKey]);

  useEffect(() => {
    dispatch({ type: "sync_filter_draft", filterDraft, subjectKey });
  }, [filterDraft, subjectKey]);

  const activateQueryChip = (
    chip: WorkbookGridQueryControlProjection["chips"][number],
    invokingElement?: HTMLButtonElement,
  ) => {
    queryEntryReturnFocusRef.current = invokingElement ?? null;
    if (chip.identity.kind === "filter") {
      const filter = queryState.filters.find(
        (candidate) => candidate.fieldKey === chip.identity.fieldKey,
      );
      if (filter === undefined) return;
      dispatch({
        type: "edit_filter",
        activeEntryKey: chip.key,
        fieldKey: filter.fieldKey,
        filterDraft: filterDraftFromFilter(filter),
        subjectKey,
      });
      return;
    }
    dispatch({
      type: "edit_query_entry",
      activeEntryKey: chip.key,
      panel: chip.identity.kind,
      subjectKey,
    });
  };

  return (
    <fieldset
      aria-label="Workbook query controls"
      data-hidden-query-chip-count={projection.hiddenChips.length}
      data-query-chip-capacity={projection.visibleChipCapacity}
      data-testid={workbookViewBarQueryControlsTestId(surface)}
      style={queryControlsStyleFor(chromeMode)}
    >
      <WorkbookSortControl
        constrained={chromeMode !== "base" || projection.hiddenChips.length > 0}
        editorProjection={sortEditorProjection}
        isOpen={surfaceState.openPanel === "sort"}
        onClose={closePanel}
        onCommand={onCommand}
        onToggle={() => {
          queryEntryReturnFocusRef.current = null;
          dispatch({ type: "toggle_panel", panel: "sort", subjectKey });
        }}
        projection={projection}
        returnFocusRef={queryEntryReturnFocusRef}
        requestedFieldKey={
          surfaceState.openPanel === "sort" &&
          surfaceState.activeEntryKey?.startsWith("sort:")
            ? surfaceState.activeEntryKey.slice("sort:".length)
            : null
        }
        surface={surface}
        sortUnapplied={sortUnapplied}
        triggerRef={sortTriggerRef}
      />
      <WorkbookGroupControl
        isOpen={surfaceState.openPanel === "group"}
        onClose={closePanel}
        onCommand={onCommand}
        onToggle={() => {
          queryEntryReturnFocusRef.current = null;
          dispatch({ type: "toggle_panel", panel: "group", subjectKey });
        }}
        projection={projection}
        returnFocusRef={queryEntryReturnFocusRef}
        selectedFieldKey={queryState.groupBy}
        surface={surface}
        triggerRef={groupTriggerRef}
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
            filterDraft: clearFilterDraftValue(draft),
            subjectKey,
          });
        }}
        onChangeDraft={(draft) => {
          dispatch({
            type: "change_filter_draft",
            filterDraft: draft,
            subjectKey,
          });
        }}
        onClose={closeFilterPanel}
        onCommand={onCommand}
        editingFieldKey={surfaceState.editingFilterFieldKey}
        onEditFilter={(fieldKey) => {
          const filter = queryState.filters.find(
            (candidate) => candidate.fieldKey === fieldKey,
          );
          if (filter === undefined) return;
          dispatch({
            type: "edit_filter",
            activeEntryKey: `filter:${fieldKey}`,
            fieldKey,
            filterDraft: filterDraftFromFilter(filter),
            subjectKey,
          });
        }}
        onEditQueryEntry={activateQueryChip}
        onToggle={() => {
          queryEntryReturnFocusRef.current = null;
          if (surfaceState.openPanel !== "filters") {
            dispatch({
              type: "change_filter_draft",
              subjectKey,
              filterDraft,
            });
          }
          dispatch({ type: "toggle_panel", panel: "filters", subjectKey });
        }}
        projection={projection}
        returnFocusRef={queryEntryReturnFocusRef}
        surface={surface}
        triggerRef={filterTriggerRef}
      />
      <WorkbookColumnsControl
        sizing={sizing}
        freezing={freezing}
        isOpen={surfaceState.openPanel === "columns"}
        onClose={closePanel}
        onCommand={onCommand}
        onToggle={() => {
          queryEntryReturnFocusRef.current = null;
          dispatch({ type: "toggle_panel", panel: "columns", subjectKey });
        }}
        projection={projection}
        surface={surface}
      />
      <WorkbookQuerySummarySlot>
        <WorkbookActiveQueryChips
          activeKey={surfaceState.rovingEntryKey}
          condensed={chromeMode !== "base"}
          entryRefs={queryEntryRefs}
          onActivate={activateQueryChip}
          onCommand={onCommand}
          onFallbackFocus={(chip) => {
            const trigger =
              projection.visibleChipCapacity === 0
                ? filterTriggerRef.current
                : chip.identity.kind === "sort"
                  ? sortTriggerRef.current
                  : chip.identity.kind === "group"
                    ? groupTriggerRef.current
                    : filterTriggerRef.current;
            if (trigger?.isConnected) trigger.focus({ preventScroll: true });
          }}
          onRovingEntryChange={(entryKey) => {
            dispatch({ type: "set_roving_entry", entryKey, subjectKey });
          }}
          projection={projection}
          surface={surface}
        />
      </WorkbookQuerySummarySlot>
    </fieldset>
  );
}

type WorkbookGridCommandPorts = Pick<
  WorkbookGridControlsProps,
  | "contract"
  | "onClearFilters"
  | "freezing"
  | "onColumnHiddenChange"
  | "onColumnMove"
  | "onGroupByChange"
  | "onRemoveFilter"
  | "onResetColumns"
  | "onSortChange"
  | "queryState"
  | "requestedSort"
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
      const currentSort = ports.requestedSort ?? ports.queryState.sort;
      const next = applyWorkbookSortCommand(
        ports.contract,
        currentSort,
        command,
      );
      if (next !== currentSort) ports.onSortChange(next);
      return;
    }
    case "group_set":
      ports.onGroupByChange(command.fieldKey);
      return;
    case "filter_remove":
      ports.onRemoveFilter(command.fieldKey);
      return;
    case "filters_clear":
      ports.onClearFilters?.();
      return;
    case "column_set_hidden":
      ports.onColumnHiddenChange(command.fieldKey, command.hidden);
      return;
    case "column_move":
      ports.onColumnMove(command.fieldKey, command.direction);
      return;
    case "columns_freeze":
      ports.freezing.onBoundaryChange(command.fieldKey);
      return;
    case "columns_reset":
      ports.onResetColumns();
      return;
  }
}

function queryControlsStyleFor(chromeMode: WorkbookChromeMode) {
  return {
    display: "flex",
    alignItems: "center",
    gap:
      chromeMode === "compact_desktop"
        ? "var(--ct-spacing-xxs)"
        : "var(--ct-spacing-xs)",
    boxSizing: "border-box" as const,
    border: 0,
    margin: 0,
    minWidth: 0,
    padding: 0,
    overflow: "visible",
  };
}
