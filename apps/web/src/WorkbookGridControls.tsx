import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";

import type { ViewContract } from "@cartulary/view-contracts";
import type { ChangeEvent } from "react";

import {
  type FilterDraft,
  filterChipLabel,
  filterInputMode,
  type WorkbookQueryState,
} from "./workbookQuery";

type WorkbookGridControlsProps = {
  readonly contract: ViewContract;
  readonly filterDraft: FilterDraft;
  readonly onApplyFilter: () => void;
  readonly onClearAll?: (() => void) | undefined;
  readonly onFilterDraftChange: (draft: FilterDraft) => void;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly queryState: WorkbookQueryState;
  readonly surface: WorkbookSurface;
};

export function WorkbookGridControls({
  contract,
  filterDraft,
  onApplyFilter,
  onClearAll,
  onFilterDraftChange,
  onGroupByChange,
  onRemoveFilter,
  queryState,
  surface,
}: WorkbookGridControlsProps) {
  const inputMode = filterInputMode(filterDraft.fieldKey);

  return (
    <div style={toolbarStyle}>
      <div style={controlRowStyle}>
        <button style={modeButtonStyle} type="button">
          Sort {queryState.sort.length > 0 ? queryState.sort.length : ""}
        </button>

        <label style={inlineLabelStyle}>
          Group
          <select
            data-testid={gridGroupingSelectTestId(surface)}
            style={selectStyle}
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

        <span style={filterClusterStyle}>
          <label style={inlineLabelStyle}>
            Filters
            <select
              data-testid={gridFilterFieldTestId(surface)}
              style={selectStyle}
              value={filterDraft.fieldKey}
              onChange={(event) => {
                onFilterDraftChange({
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
            <select
              aria-label="Filter value"
              data-testid={gridFilterValueTestId(surface)}
              style={selectStyle}
              value={filterDraft.booleanValue}
              onChange={(event) => {
                onFilterDraftChange({
                  ...filterDraft,
                  booleanValue: event.target
                    .value as FilterDraft["booleanValue"],
                });
              }}
            >
              <option value="">Value</option>
              <option value="true">true</option>
              <option value="false">false</option>
            </select>
          ) : (
            <input
              aria-label="Filter value"
              data-testid={gridFilterValueTestId(surface)}
              placeholder={placeholderForMode(inputMode)}
              style={inputStyle}
              type="text"
              value={filterDraft.value}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onFilterDraftChange({
                  ...filterDraft,
                  value: event.target.value,
                });
              }}
            />
          )}

          <button
            data-testid={gridFilterApplyTestId(surface)}
            style={actionButtonStyle}
            type="button"
            onClick={onApplyFilter}
          >
            Apply
          </button>
        </span>

        {queryState.filters.map((filter) => (
          <button
            key={filter.fieldKey}
            data-testid={gridFilterChipTestId(surface, filter.fieldKey)}
            style={chipButtonStyle}
            type="button"
            onClick={() => {
              onRemoveFilter(filter.fieldKey);
            }}
          >
            {filterChipLabel(contract, filter)} ×
          </button>
        ))}

        {queryState.filters.length > 0 ||
        queryState.sort.length > 0 ||
        queryState.groupBy !== null ? (
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

const toolbarStyle = {
  display: "flex",
  alignItems: "center",
  flex: "1 1 auto",
  minWidth: 0,
  overflowX: "auto" as const,
};

const controlRowStyle = {
  display: "flex",
  gap: "0.45rem",
  flexWrap: "nowrap" as const,
  alignItems: "center",
  minWidth: 0,
};

const inlineLabelStyle = {
  display: "inline-flex",
  gap: "0.35rem",
  alignItems: "center",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.86rem",
};

const inputStyle = {
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.42rem 0.55rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
  boxSizing: "border-box" as const,
  minWidth: "8rem",
  maxWidth: "10rem",
};

const selectStyle = {
  ...inputStyle,
  minWidth: "9rem",
};

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.42rem 0.6rem",
  font: "inherit",
  cursor: "pointer",
  height: "fit-content",
};

const modeButtonStyle = {
  ...actionButtonStyle,
  whiteSpace: "nowrap" as const,
};

const chipButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};

const clearButtonStyle = {
  ...actionButtonStyle,
  borderColor: "transparent",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
};

const filterClusterStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  minWidth: 0,
};
