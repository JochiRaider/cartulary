import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  type WorkbookSurface,
} from "@cartulary/test-utils";

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
  color: "rgb(52 79 72)",
};

const inputStyle = {
  borderRadius: "0.75rem",
  border: "1px solid rgb(192 205 198)",
  background: "rgb(255 255 255)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "inherit",
};

const selectStyle = {
  ...inputStyle,
};

const actionButtonStyle = {
  borderRadius: "999px",
  border: "1px solid rgb(129 165 154)",
  background: "rgb(234 244 239)",
  color: "rgb(34 74 63)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
  height: "fit-content",
};

const chipButtonStyle = {
  ...actionButtonStyle,
  background: "rgb(247 249 247)",
};

const hintStyle = {
  color: "rgb(87 112 104)",
};
