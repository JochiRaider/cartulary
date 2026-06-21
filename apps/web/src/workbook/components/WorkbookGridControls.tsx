import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  type WorkbookSurface,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
  workbookTopBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { SlidersHorizontal } from "lucide-react";
import {
  type ChangeEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  useEffect,
  useMemo,
  useState,
} from "react";

import {
  type FilterDraft,
  filterChipLabel,
  filterInputMode,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { visuallyHiddenStyle } from "../utils/workbookStyles";

type WorkbookGridControlsProps = {
  readonly contract: ViewContract;
  readonly defaultFilterPopoverOpen?: boolean | undefined;
  readonly filterDraft: FilterDraft;
  readonly onApplyFilter: (draft: FilterDraft) => void;
  readonly onClearAll?: (() => void) | undefined;
  readonly onFilterDraftChange: (draft: FilterDraft) => void;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly onToggleSort: (fieldKey: string) => void;
  readonly queryState: WorkbookQueryState;
  readonly surface: WorkbookSurface;
};

export function WorkbookGridControls({
  contract,
  defaultFilterPopoverOpen = false,
  filterDraft,
  onApplyFilter,
  onClearAll,
  onFilterDraftChange,
  onGroupByChange,
  onRemoveFilter,
  onToggleSort,
  queryState,
  surface,
}: WorkbookGridControlsProps) {
  const [isSortMenuOpen, setIsSortMenuOpen] = useState(false);
  const [isFilterPopoverOpen, setIsFilterPopoverOpen] = useState(
    defaultFilterPopoverOpen,
  );
  const [draft, setDraft] = useState(filterDraft);
  const inputMode = filterInputMode(draft.fieldKey);
  const activeSort = queryState.sort[0] ?? null;
  const visibleSortFields = useMemo(
    () =>
      contract.fields
        .map((field) => field.fieldKey)
        .filter((fieldKey) => contract.sortableFieldMap[fieldKey]),
    [contract],
  );
  const hasActiveQuery =
    queryState.filters.length > 0 ||
    queryState.sort.length > 0 ||
    queryState.groupBy !== null;
  const activeQueryChipCount =
    queryState.filters.length +
    queryState.sort.length +
    (queryState.groupBy === null ? 0 : 1);
  const draftValueMissing =
    inputMode === "boolean"
      ? draft.booleanValue === ""
      : draft.value.trim() === "";

  useEffect(() => {
    if (!isFilterPopoverOpen) {
      setDraft(filterDraft);
    }
  }, [filterDraft, isFilterPopoverOpen]);

  const closeFilterPopover = () => {
    setDraft(filterDraft);
    setIsFilterPopoverOpen(false);
  };

  const commitFilterDraft = () => {
    onFilterDraftChange(draft);
    onApplyFilter(draft);
    setDraft(clearAppliedDraftValue(draft));
    setIsFilterPopoverOpen(false);
  };

  return (
    <div
      data-testid={workbookTopBarQueryControlsTestId(surface)}
      style={queryControlsStyle}
    >
      <div style={menuFrameStyle}>
        <button
          aria-controls={
            isSortMenuOpen ? workbookSortMenuTestId(surface) : undefined
          }
          aria-expanded={isSortMenuOpen}
          aria-haspopup="menu"
          data-testid={workbookSortMenuTriggerTestId(surface)}
          style={controlButtonStyle}
          type="button"
          onClick={() => {
            setIsSortMenuOpen((current) => !current);
          }}
        >
          Sort
          {activeSort ? `: ${sortLabel(contract, activeSort.fieldKey)}` : ""}
        </button>
        {isSortMenuOpen ? (
          <div
            data-testid={workbookSortMenuTestId(surface)}
            id={workbookSortMenuTestId(surface)}
            role="menu"
            style={menuStyle}
          >
            {visibleSortFields.map((fieldKey) => {
              const isSelected = activeSort?.fieldKey === fieldKey;
              return (
                <button
                  key={fieldKey}
                  aria-checked={isSelected}
                  data-testid={workbookSortOptionTestId(surface, fieldKey)}
                  role="menuitemradio"
                  style={{
                    ...menuItemStyle,
                    ...(isSelected ? menuItemSelectedStyle : null),
                  }}
                  type="button"
                  onClick={() => {
                    onToggleSort(fieldKey);
                    setIsSortMenuOpen(false);
                  }}
                >
                  {sortLabel(contract, fieldKey)}
                  {isSelected ? ` ${activeSort.direction}` : ""}
                </button>
              );
            })}
          </div>
        ) : null}
      </div>

      <label style={inlineLabelStyle}>
        Group:
        <select
          aria-label="Group rows"
          data-testid={gridGroupingSelectTestId(surface)}
          style={groupSelectStyle}
          value={queryState.groupBy ?? ""}
          onChange={(event) => {
            onGroupByChange(
              event.target.value === "" ? null : event.target.value,
            );
          }}
        >
          <option value="">None</option>
          {contract.groupingFields.map((fieldKey) => (
            <option key={fieldKey} value={fieldKey}>
              {contract.fieldMap[fieldKey]?.label ?? fieldKey}
            </option>
          ))}
        </select>
      </label>

      <div style={menuFrameStyle}>
        <button
          aria-controls={
            isFilterPopoverOpen
              ? workbookFilterPopoverTestId(surface)
              : undefined
          }
          aria-expanded={isFilterPopoverOpen}
          aria-haspopup="dialog"
          data-testid={workbookFilterPopoverTriggerTestId(surface)}
          style={controlButtonStyle}
          type="button"
          onClick={() => {
            setDraft(filterDraft);
            setIsFilterPopoverOpen((current) => !current);
          }}
        >
          <SlidersHorizontal aria-hidden="true" size={15} />
          Filters
          {queryState.filters.length > 0 ? ` ${queryState.filters.length}` : ""}
        </button>
        {isFilterPopoverOpen ? (
          <div
            aria-label="Draft filters"
            data-testid={workbookFilterPopoverTestId(surface)}
            id={workbookFilterPopoverTestId(surface)}
            role="dialog"
            style={filterPopoverStyle}
            onKeyDown={(event: ReactKeyboardEvent<HTMLDivElement>) => {
              if (event.key === "Escape") {
                event.preventDefault();
                closeFilterPopover();
              }
            }}
          >
            <label style={stackedLabelStyle}>
              Field
              <select
                data-testid={gridFilterFieldTestId(surface)}
                style={selectStyle}
                value={draft.fieldKey}
                onChange={(event) => {
                  setDraft({
                    booleanValue: "",
                    fieldKey: event.target.value,
                    value: "",
                  });
                }}
              >
                {contract.filterFields.map((fieldKey) => (
                  <option key={fieldKey} value={fieldKey}>
                    {contract.fieldMap[fieldKey]?.label ?? fieldKey}
                  </option>
                ))}
              </select>
            </label>

            {inputMode === "boolean" ? (
              <label style={stackedLabelStyle}>
                Value
                <select
                  data-testid={gridFilterValueTestId(surface)}
                  style={selectStyle}
                  value={draft.booleanValue}
                  onChange={(event) => {
                    setDraft({
                      ...draft,
                      booleanValue: event.target
                        .value as FilterDraft["booleanValue"],
                    });
                  }}
                >
                  <option value="">Select value</option>
                  <option value="true">true</option>
                  <option value="false">false</option>
                </select>
              </label>
            ) : (
              <label style={stackedLabelStyle}>
                Value
                <input
                  data-testid={gridFilterValueTestId(surface)}
                  placeholder={placeholderForMode(inputMode)}
                  style={inputStyle}
                  type="text"
                  value={draft.value}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => {
                    setDraft({
                      ...draft,
                      value: event.target.value,
                    });
                  }}
                />
              </label>
            )}

            {draftValueMissing ? (
              <p role="status" style={filterValidationStyle}>
                Enter a value before applying this filter.
              </p>
            ) : null}

            <div style={popoverActionsStyle}>
              <button
                style={secondaryButtonStyle}
                type="button"
                onClick={closeFilterPopover}
              >
                Cancel
              </button>
              <button
                data-testid={gridFilterApplyTestId(surface)}
                disabled={draftValueMissing}
                style={primaryButtonStyle}
                type="button"
                onClick={commitFilterDraft}
              >
                Apply
              </button>
            </div>
          </div>
        ) : null}
      </div>

      <div aria-label="Active query chips" role="toolbar" style={chipRailStyle}>
        <span style={visuallyHiddenStyle}>Active query chips</span>
        {queryState.groupBy ? (
          <button
            style={chipButtonStyle}
            title={`Group: ${
              contract.fieldMap[queryState.groupBy]?.label ?? queryState.groupBy
            }`}
            type="button"
            onClick={() => {
              onGroupByChange(null);
            }}
          >
            Group:{" "}
            {contract.fieldMap[queryState.groupBy]?.label ?? queryState.groupBy}
          </button>
        ) : null}
        {queryState.sort.map((sort) => {
          const label = `Sort: ${sortLabel(contract, sort.fieldKey)} ${
            sort.direction
          }`;
          return (
            <button
              key={sort.fieldKey}
              style={chipButtonStyle}
              title={label}
              type="button"
              onClick={() => {
                onToggleSort(sort.fieldKey);
              }}
            >
              {label}
            </button>
          );
        })}
        {queryState.filters.map((filter) => {
          const label = filterChipLabel(contract, filter);
          return (
            <button
              key={filter.fieldKey}
              data-testid={gridFilterChipTestId(surface, filter.fieldKey)}
              style={chipButtonStyle}
              title={label}
              type="button"
              onClick={() => {
                onRemoveFilter(filter.fieldKey);
              }}
            >
              {label}
            </button>
          );
        })}
        {hasActiveQuery && activeQueryChipCount > 1 ? (
          <button
            style={clearButtonStyle}
            type="button"
            onClick={() => {
              onClearAll?.();
            }}
          >
            Clear all
          </button>
        ) : null}
      </div>
    </div>
  );
}

function clearAppliedDraftValue(current: FilterDraft): FilterDraft {
  return {
    ...current,
    booleanValue: "",
    value: "",
  };
}

function placeholderForMode(mode: ReturnType<typeof filterInputMode>) {
  switch (mode) {
    case "date":
      return "YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD";
    case "tagset":
      return "comma-separated values";
    case "timestamp":
      return "RFC3339 or start..end";
    default:
      return "Value";
  }
}

function sortLabel(contract: ViewContract, fieldKey: string) {
  return contract.fieldMap[fieldKey]?.label ?? fieldKey;
}

const queryControlsStyle = {
  display: "grid",
  gridTemplateColumns:
    "max-content minmax(0, max-content) max-content minmax(5.5rem, 1fr)",
  alignItems: "center",
  gap: "0.35rem",
  inlineSize: "100%",
  maxInlineSize: "100%",
  boxSizing: "border-box" as const,
  minWidth: 0,
  minInlineSize: 0,
  flex: "1 1 auto",
  overflow: "hidden",
};

const menuFrameStyle = {
  position: "relative" as const,
  display: "inline-flex",
  flex: "0 0 auto",
};

const controlButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.3rem",
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  padding: "0.28rem 0.5rem",
  font: "inherit",
  cursor: "pointer",
  minBlockSize: "1.8rem",
  whiteSpace: "nowrap" as const,
};

const inputStyle = {
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.35rem 0.5rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
  boxSizing: "border-box" as const,
  minInlineSize: "12rem",
  minBlockSize: "1.9rem",
};

const selectStyle = {
  ...inputStyle,
  minInlineSize: "9rem",
};

const groupSelectStyle = {
  ...selectStyle,
  minInlineSize: "5.75rem",
  maxInlineSize: "7rem",
};

const inlineLabelStyle = {
  display: "inline-flex",
  gap: "0.25rem",
  alignItems: "center",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
  whiteSpace: "nowrap" as const,
  minWidth: 0,
};

const stackedLabelStyle = {
  display: "grid",
  gap: "0.3rem",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
};

const menuStyle = {
  position: "absolute" as const,
  zIndex: 20,
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineStart: 0,
  display: "grid",
  gap: "0.2rem",
  inlineSize: "min(18rem, 80vw)",
  maxBlockSize: "18rem",
  overflowY: "auto" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  padding: "0.45rem",
};

const menuItemStyle = {
  border: 0,
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
  font: "inherit",
  padding: "0.45rem 0.5rem",
  textAlign: "left" as const,
};

const menuItemSelectedStyle = {
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};

const filterPopoverStyle = {
  ...menuStyle,
  inlineSize: "min(24rem, 88vw)",
  gap: "0.55rem",
};

const filterValidationStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  fontSize: "0.8rem",
  fontWeight: 700,
};

const popoverActionsStyle = {
  display: "flex",
  justifyContent: "end",
  gap: "0.4rem",
};

const secondaryButtonStyle = {
  ...controlButtonStyle,
  background: "transparent",
};

const primaryButtonStyle = {
  ...controlButtonStyle,
  borderColor: "var(--ct-colors-accent-active)",
  background: "var(--ct-colors-accent)",
  color: "var(--ct-colors-on-accent)",
  fontWeight: 700,
};

const chipRailStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.3rem",
  border: 0,
  padding: 0,
  margin: 0,
  inlineSize: "100%",
  maxInlineSize: "100%",
  minWidth: 0,
  minInlineSize: 0,
  flex: "1 1 auto",
  overflowX: "hidden" as const,
  overflowY: "hidden" as const,
};

const chipButtonStyle = {
  ...controlButtonStyle,
  background: "var(--ct-colors-surface-3)",
  flex: "1 1 0",
  minInlineSize: 0,
  maxInlineSize: "7rem",
  overflow: "hidden",
  textOverflow: "ellipsis",
};

const clearButtonStyle = {
  ...controlButtonStyle,
  flex: "0 0 auto",
  borderColor: "transparent",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
};
