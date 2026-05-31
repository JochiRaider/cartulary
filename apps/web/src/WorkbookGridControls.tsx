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
  onFilterDraftChange,
  onGroupByChange,
  onRemoveFilter,
  queryState,
  surface,
}: WorkbookGridControlsProps) {
  const inputMode = filterInputMode(filterDraft.fieldKey);

  return (
    <div style={toolbarStyle}>
      <div style={chipRowStyle}>
        {queryState.filters.length > 0 ? (
          queryState.filters.map((filter) => (
            <button
              key={filter.fieldKey}
              data-testid={gridFilterChipTestId(surface, filter.fieldKey)}
              style={chipButtonStyle}
              type="button"
              onClick={() => {
                onRemoveFilter(filter.fieldKey);
              }}
            >
              {filterChipLabel(contract, filter)}
            </button>
          ))
        ) : (
          <span style={hintStyle}>No active filters</span>
        )}
      </div>

      <div style={controlRowStyle}>
        <label style={labelStyle}>
          Filter
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
          <label style={labelStyle}>
            Value
            <select
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
              <option value="">Select value</option>
              <option value="true">true</option>
              <option value="false">false</option>
            </select>
          </label>
        ) : (
          <label style={labelStyle}>
            Value
            <input
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
          </label>
        )}

        <button
          data-testid={gridFilterApplyTestId(surface)}
          style={actionButtonStyle}
          type="button"
          onClick={onApplyFilter}
        >
          Apply filter
        </button>

        <label style={labelStyle}>
          Group by
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
            <option value="">No grouping</option>
            {contract.groupingFields.map((fieldKey) => (
              <option key={fieldKey} value={fieldKey}>
                {contract.fieldMap[fieldKey]?.label ?? fieldKey}
              </option>
            ))}
          </select>
        </label>
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
  display: "grid",
  gap: "0.75rem",
  marginBottom: "0.75rem",
};

const chipRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};

const controlRowStyle = {
  display: "flex",
  gap: "0.75rem",
  flexWrap: "wrap" as const,
  alignItems: "end",
};

const labelStyle = {
  display: "grid",
  gap: "0.35rem",
  minWidth: "12rem",
  color: "var(--ct-colors-ink-muted)",
};

const inputStyle = {
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "var(--ct-component-text-input-padding)",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};

const selectStyle = {
  ...inputStyle,
};

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "var(--ct-component-button-secondary-padding)",
  font: "inherit",
  cursor: "pointer",
  height: "fit-content",
};

const chipButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};

const hintStyle = {
  color: "var(--ct-colors-ink-muted)",
};
