import {
  workbookAddRowButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { Filter, Plus, Rows3, Search, SlidersHorizontal } from "lucide-react";
import type { CSSProperties, ReactNode } from "react";
import type { FilterDraft, WorkbookQueryState } from "../models/workbookQuery";
import { WorkbookGridControls } from "./WorkbookGridControls";

type WorkbookSheetToolbarProps = {
  readonly addRowDisabled?: boolean | undefined;
  readonly addRowLabel?: string | undefined;
  readonly contract: ViewContract;
  readonly filterDraft: FilterDraft;
  readonly leading?: ReactNode | undefined;
  readonly onAddRow?: (() => void) | undefined;
  readonly onApplyFilter: () => void;
  readonly onClearAll?: (() => void) | undefined;
  readonly onFilterDraftChange: (draft: FilterDraft) => void;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onInspectorToggle?: (() => void) | undefined;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly queryState: WorkbookQueryState;
  readonly showQueryControls?: boolean | undefined;
  readonly showSurfaceStatus?: boolean | undefined;
  readonly surface: string;
};

export function WorkbookSheetToolbar({
  addRowDisabled = false,
  addRowLabel = "Add row",
  contract,
  filterDraft,
  leading,
  onAddRow,
  onApplyFilter,
  onClearAll,
  onFilterDraftChange,
  onGroupByChange,
  onInspectorToggle,
  onRemoveFilter,
  queryState,
  showQueryControls = true,
  showSurfaceStatus = true,
  surface,
}: WorkbookSheetToolbarProps) {
  const activeControlCount =
    queryState.sort.length +
    queryState.filters.length +
    (queryState.groupBy === null ? 0 : 1);
  return (
    <div style={showQueryControls ? toolbarStyle : savedViewToolbarStyle}>
      <div style={leftRailStyle}>
        {leading}
        {showSurfaceStatus ? (
          <>
            <span style={toolbarDividerStyle} />
            <span style={toolbarStatusStyle}>
              <Rows3 aria-hidden="true" size={16} />
              {contract.title}
            </span>
            <span style={toolbarStatusStyle}>
              <Filter aria-hidden="true" size={16} />
              {activeControlCount}
            </span>
          </>
        ) : null}
      </div>
      {showQueryControls ? (
        <WorkbookGridControls
          contract={contract}
          filterDraft={filterDraft}
          onApplyFilter={onApplyFilter}
          onClearAll={onClearAll}
          onFilterDraftChange={onFilterDraftChange}
          onGroupByChange={onGroupByChange}
          onRemoveFilter={onRemoveFilter}
          queryState={queryState}
          surface={surface}
        />
      ) : null}
      <div style={rightRailStyle}>
        {onInspectorToggle ? (
          <button
            aria-label="Open inspector"
            data-testid={workbookInspectorToggleTestId(surface)}
            style={toolbarButtonStyle}
            type="button"
            onClick={onInspectorToggle}
          >
            <SlidersHorizontal aria-hidden="true" size={16} />
            Inspector
          </button>
        ) : null}
        {onAddRow ? (
          <button
            data-testid={workbookAddRowButtonTestId(surface)}
            disabled={addRowDisabled}
            style={primaryToolbarButtonStyle}
            type="button"
            onClick={onAddRow}
          >
            <Plus aria-hidden="true" size={16} />
            {addRowLabel}
          </button>
        ) : null}
      </div>
    </div>
  );
}

type WorkbookViewBarQueryControlsProps = {
  readonly contract: ViewContract;
  readonly filterDraft: FilterDraft;
  readonly onApplyFilter: () => void;
  readonly onClearAll?: (() => void) | undefined;
  readonly onFilterDraftChange: (draft: FilterDraft) => void;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly queryState: WorkbookQueryState;
  readonly surface: string;
};

export function WorkbookViewBarQueryControls({
  contract,
  filterDraft,
  onApplyFilter,
  onClearAll,
  onFilterDraftChange,
  onGroupByChange,
  onRemoveFilter,
  queryState,
  surface,
}: WorkbookViewBarQueryControlsProps) {
  const activeControlCount =
    queryState.sort.length +
    queryState.filters.length +
    (queryState.groupBy === null ? 0 : 1);

  return (
    <div style={viewBarQueryControlsStyle}>
      <span style={toolbarStatusStyle}>
        <Filter aria-hidden="true" size={16} />
        {activeControlCount}
      </span>
      <WorkbookGridControls
        contract={contract}
        filterDraft={filterDraft}
        onApplyFilter={onApplyFilter}
        onClearAll={onClearAll}
        onFilterDraftChange={onFilterDraftChange}
        onGroupByChange={onGroupByChange}
        onRemoveFilter={onRemoveFilter}
        queryState={queryState}
        surface={surface}
      />
    </div>
  );
}

export const compactInputStyle = {
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  font: "inherit",
  minHeight: "1.75rem",
  padding: "0.22rem 0.45rem",
} satisfies CSSProperties;

export const toolbarButtonStyle = {
  ...compactInputStyle,
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.35rem",
  cursor: "pointer",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

export const primaryToolbarButtonStyle = {
  ...toolbarButtonStyle,
  borderColor: "var(--ct-colors-accent-active)",
  background: "var(--ct-colors-accent)",
  color: "var(--ct-colors-on-accent)",
  fontWeight: 700,
} satisfies CSSProperties;

const toolbarStyle = {
  display: "grid",
  gridTemplateColumns: "auto minmax(0, 1fr) auto",
  alignItems: "center",
  gap: "0.5rem",
  minHeight: "var(--ct-layout-viewBarHeight)",
  minWidth: 0,
  padding: "0 var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

const savedViewToolbarStyle = {
  ...toolbarStyle,
  gridTemplateColumns: "minmax(0, 1fr) auto",
} satisfies CSSProperties;

const leftRailStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.45rem",
  minWidth: 0,
} satisfies CSSProperties;

const rightRailStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "end",
  gap: "0.45rem",
  minWidth: 0,
} satisfies CSSProperties;

const viewBarQueryControlsStyle = {
  display: "inline-flex",
  alignItems: "center",
  flexWrap: "wrap",
  gap: "0.35rem",
  flex: "1 1 auto",
  minWidth: 0,
} satisfies CSSProperties;

const toolbarStatusStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.3rem",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const toolbarDividerStyle = {
  display: "inline-block",
  inlineSize: 1,
  blockSize: "1.25rem",
  background: "var(--ct-colors-hairline)",
} satisfies CSSProperties;

export function WorkbookToolbarSearchLabel({
  children,
}: {
  readonly children: ReactNode;
}) {
  return (
    <span style={toolbarStatusStyle}>
      <Search aria-hidden="true" size={16} />
      {children}
    </span>
  );
}
