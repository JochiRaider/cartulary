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
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import { SlidersHorizontal } from "lucide-react";
import {
  type ChangeEvent,
  type CSSProperties,
  type KeyboardEvent as ReactKeyboardEvent,
  type RefObject,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { WorkbookResolvedLayoutState } from "../models/workbookLayout";
import {
  cycleWorkbookSortField,
  type FilterDraft,
  filterChipLabel,
  filterInputMode,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import {
  type WorkbookChromeMode,
  workbookQueryChipCapacity,
} from "../models/workbookResponsiveLayout";
import { visuallyHiddenStyle } from "../utils/workbookStyles";

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
  readonly surface: WorkbookSurface;
};

type QueryChip = {
  readonly key: string;
  readonly label: string;
  readonly remove: () => void;
  readonly testId?: string | undefined;
};

function focusFirstMenuItem(menuId: string) {
  window.requestAnimationFrame(() => {
    document
      .getElementById(menuId)
      ?.querySelector<HTMLElement>('[role^="menuitem"]:not(:disabled)')
      ?.focus({ preventScroll: true });
  });
}

function handleMenuKeyboard(
  event: ReactKeyboardEvent<HTMLDivElement>,
  triggerRef: RefObject<HTMLButtonElement | null>,
  close: () => void,
) {
  if (event.key === "Escape") {
    event.preventDefault();
    event.stopPropagation();
    close();
    triggerRef.current?.focus({ preventScroll: true });
    return;
  }
  if (
    event.key !== "ArrowDown" &&
    event.key !== "ArrowUp" &&
    event.key !== "Home" &&
    event.key !== "End"
  ) {
    return;
  }
  const items = Array.from(
    event.currentTarget.querySelectorAll<HTMLElement>(
      '[role^="menuitem"]:not(:disabled)',
    ),
  );
  if (items.length === 0) return;
  const activeIndex =
    document.activeElement instanceof HTMLElement
      ? items.indexOf(document.activeElement)
      : -1;
  let nextIndex = 0;
  if (event.key === "End") {
    nextIndex = items.length - 1;
  } else if (event.key === "ArrowUp") {
    nextIndex = activeIndex <= 0 ? items.length - 1 : activeIndex - 1;
  } else if (event.key === "ArrowDown") {
    nextIndex =
      activeIndex < 0 || activeIndex === items.length - 1 ? 0 : activeIndex + 1;
  }
  event.preventDefault();
  items[nextIndex]?.focus({ preventScroll: true });
}

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
  const [isSortMenuOpen, setIsSortMenuOpen] = useState(false);
  const [isColumnsMenuOpen, setIsColumnsMenuOpen] = useState(false);
  const [isFilterPopoverOpen, setIsFilterPopoverOpen] = useState(
    defaultFilterPopoverOpen,
  );
  const [draft, setDraft] = useState(filterDraft);
  const mountedSurfaceRef = useRef(surface);
  const sortMenuTriggerRef = useRef<HTMLButtonElement>(null);
  const columnsMenuTriggerRef = useRef<HTMLButtonElement>(null);
  const filterPopoverTriggerRef = useRef<HTMLButtonElement>(null);
  const inputMode = filterInputMode(draft.fieldKey);
  const activeSort = queryState.sort[0] ?? null;
  const hiddenFieldKeys = useMemo(
    () => new Set(layoutState.hiddenFieldKeys),
    [layoutState.hiddenFieldKeys],
  );
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
  const groupChipLabel =
    queryState.groupBy === null
      ? null
      : `Group: ${
          contract.fieldMap[queryState.groupBy]?.label ?? queryState.groupBy
        }`;
  const draftValueMissing =
    inputMode === "boolean"
      ? draft.booleanValue === ""
      : draft.value.trim() === "";
  const activeQueryChips: QueryChip[] = [
    ...(queryState.groupBy && groupChipLabel
      ? [
          {
            key: `group:${queryState.groupBy}`,
            label: groupChipLabel,
            remove: () => onGroupByChange(null),
          },
        ]
      : []),
    ...queryState.sort.map((sort) => ({
      key: `sort:${sort.fieldKey}`,
      label: `Sort: ${sortLabel(contract, sort.fieldKey)} ${sort.direction}`,
      remove: () => {
        onSortChange(
          queryState.sort.filter((entry) => entry.fieldKey !== sort.fieldKey),
        );
      },
    })),
    ...queryState.filters.map((filter) => ({
      key: `filter:${filter.fieldKey}`,
      label: filterChipLabel(contract, filter),
      remove: () => onRemoveFilter(filter.fieldKey),
      testId: gridFilterChipTestId(surface, filter.fieldKey),
    })),
  ];
  const visibleChipCapacity = workbookQueryChipCapacity(chromeMode);
  const visibleQueryChips = activeQueryChips.slice(0, visibleChipCapacity);
  const hiddenQueryChips = activeQueryChips.slice(visibleChipCapacity);

  useEffect(() => {
    if (!isFilterPopoverOpen) {
      setDraft(filterDraft);
    }
  }, [filterDraft, isFilterPopoverOpen]);

  useEffect(() => {
    if (mountedSurfaceRef.current === surface) return;
    mountedSurfaceRef.current = surface;
    setDraft(filterDraft);
    setIsColumnsMenuOpen(false);
    setIsFilterPopoverOpen(false);
    setIsSortMenuOpen(false);
  }, [filterDraft, surface]);

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
    <fieldset
      aria-label="Workbook query controls"
      data-hidden-query-chip-count={hiddenQueryChips.length}
      data-query-chip-capacity={visibleChipCapacity}
      data-testid={workbookViewBarQueryControlsTestId(surface)}
      style={{
        ...queryControlsStyle,
        ...(visibleChipCapacity === 0 ? compactQueryControlsStyle : null),
      }}
      onKeyDown={(event) => {
        if (event.key !== "Escape" || event.defaultPrevented) return;
        if (isFilterPopoverOpen) {
          event.preventDefault();
          event.stopPropagation();
          closeFilterPopover();
          filterPopoverTriggerRef.current?.focus({ preventScroll: true });
          return;
        }
        if (isColumnsMenuOpen) {
          event.preventDefault();
          event.stopPropagation();
          setIsColumnsMenuOpen(false);
          columnsMenuTriggerRef.current?.focus({ preventScroll: true });
          return;
        }
        if (isSortMenuOpen) {
          event.preventDefault();
          event.stopPropagation();
          setIsSortMenuOpen(false);
          sortMenuTriggerRef.current?.focus({ preventScroll: true });
        }
      }}
    >
      <div style={menuFrameStyle}>
        <button
          aria-controls={
            isSortMenuOpen ? workbookSortMenuTestId(surface) : undefined
          }
          aria-expanded={isSortMenuOpen}
          aria-haspopup="menu"
          data-testid={workbookSortMenuTriggerTestId(surface)}
          ref={sortMenuTriggerRef}
          style={controlButtonStyle}
          type="button"
          onClick={() => {
            setIsSortMenuOpen((current) => !current);
          }}
          onKeyDown={(event) => {
            if (event.key !== "ArrowDown") return;
            event.preventDefault();
            setIsSortMenuOpen(true);
            focusFirstMenuItem(workbookSortMenuTestId(surface));
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
            onKeyDown={(event) => {
              handleMenuKeyboard(event, sortMenuTriggerRef, () =>
                setIsSortMenuOpen(false),
              );
            }}
          >
            {visibleSortFields.map((fieldKey) => {
              const priority = queryState.sort.findIndex(
                (entry) => entry.fieldKey === fieldKey,
              );
              const selectedSort =
                priority < 0 ? undefined : queryState.sort[priority];
              const isSelected = selectedSort !== undefined;
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
                    onSortChange(
                      cycleWorkbookSortField(
                        contract,
                        queryState,
                        fieldKey,
                        true,
                      ).sort,
                    );
                  }}
                >
                  {sortLabel(contract, fieldKey)}
                  {isSelected
                    ? ` ${priority + 1} ${selectedSort?.direction}`
                    : ""}
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
          aria-label={
            hiddenQueryChips.length > 0
              ? `Filters, ${queryState.filters.length} active filters, ${hiddenQueryChips.length} active query chips hidden`
              : `Filters, ${queryState.filters.length} active filters`
          }
          aria-controls={
            isFilterPopoverOpen
              ? workbookFilterPopoverTestId(surface)
              : undefined
          }
          aria-expanded={isFilterPopoverOpen}
          aria-haspopup="dialog"
          data-testid={workbookFilterPopoverTriggerTestId(surface)}
          ref={filterPopoverTriggerRef}
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
                event.stopPropagation();
                closeFilterPopover();
                filterPopoverTriggerRef.current?.focus({
                  preventScroll: true,
                });
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

            {activeQueryChips.length > 0 ? (
              <section
                aria-label="Active query overflow"
                style={queryListStyle}
              >
                <strong>Active query</strong>
                {activeQueryChips.map((chip, index) => (
                  <button
                    key={chip.key}
                    aria-label={`Remove ${chip.label}${
                      index >= visibleChipCapacity
                        ? ", hidden from the view bar"
                        : ""
                    }`}
                    style={queryListButtonStyle}
                    type="button"
                    onClick={chip.remove}
                  >
                    {chip.label}
                  </button>
                ))}
                {activeQueryChipCount > 1 ? (
                  <button
                    style={clearButtonStyle}
                    type="button"
                    onClick={() => onClearAll?.()}
                  >
                    Clear all
                  </button>
                ) : null}
              </section>
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

      <div style={menuFrameStyle}>
        <button
          aria-expanded={isColumnsMenuOpen}
          aria-haspopup="menu"
          ref={columnsMenuTriggerRef}
          style={controlButtonStyle}
          type="button"
          onClick={() => {
            setIsColumnsMenuOpen((current) => !current);
          }}
          onKeyDown={(event) => {
            if (event.key !== "ArrowDown") return;
            event.preventDefault();
            setIsColumnsMenuOpen(true);
            focusFirstMenuItem(`${surface}-column-menu`);
          }}
        >
          Columns
        </button>
        {isColumnsMenuOpen ? (
          <div
            aria-label="Column controls"
            id={`${surface}-column-menu`}
            role="menu"
            style={menuStyle}
            onKeyDown={(event) => {
              handleMenuKeyboard(event, columnsMenuTriggerRef, () =>
                setIsColumnsMenuOpen(false),
              );
            }}
          >
            {layoutState.columnOrder.map((fieldKey, index) => {
              const field = contract.fieldMap[fieldKey];
              if (field === undefined) return null;
              const hidden = hiddenFieldKeys.has(fieldKey);
              return (
                <div key={fieldKey} role="none" style={columnMenuRowStyle}>
                  <button
                    aria-checked={!hidden}
                    role="menuitemcheckbox"
                    style={menuItemStyle}
                    type="button"
                    onClick={() => {
                      onColumnHiddenChange(fieldKey, !hidden);
                    }}
                  >
                    {field.label}
                  </button>
                  <button
                    aria-label={`Move ${field.label} earlier`}
                    disabled={index === 0}
                    role="menuitem"
                    type="button"
                    onClick={() => {
                      onColumnMove(fieldKey, "earlier");
                    }}
                  >
                    ↑
                  </button>
                  <button
                    aria-label={`Move ${field.label} later`}
                    disabled={index === layoutState.columnOrder.length - 1}
                    role="menuitem"
                    type="button"
                    onClick={() => {
                      onColumnMove(fieldKey, "later");
                    }}
                  >
                    ↓
                  </button>
                </div>
              );
            })}
            <button
              role="menuitem"
              style={menuItemStyle}
              type="button"
              onClick={onResetColumns}
            >
              Reset columns
            </button>
          </div>
        ) : null}
      </div>

      <div
        aria-label="Active query chips"
        hidden={visibleChipCapacity === 0}
        role="toolbar"
        style={chipRailStyle}
      >
        <span style={visuallyHiddenStyle}>Active query chips</span>
        {visibleQueryChips.map((chip) => (
          <button
            key={chip.key}
            data-testid={chip.testId}
            style={chipButtonStyle}
            title={chip.label}
            type="button"
            onClick={chip.remove}
          >
            <span style={chipLabelStyle}>{chip.label}</span>
          </button>
        ))}
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
    </fieldset>
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
    "max-content max-content minmax(0, max-content) max-content minmax(5.5rem, 1fr)",
  alignItems: "center",
  gap: "0.35rem",
  inlineSize: "auto",
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

const compactQueryControlsStyle = {
  gridTemplateColumns:
    "max-content minmax(0, max-content) max-content max-content 0",
};

const menuFrameStyle = {
  position: "relative" as const,
  display: "inline-flex",
  flex: "0 0 auto",
};

const columnMenuRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(10rem, 1fr) 2rem 2rem",
  alignItems: "center",
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

const queryListStyle = {
  display: "grid",
  gap: "0.25rem",
  borderBlockStart: "var(--ct-border-hairline)",
  paddingBlockStart: "0.45rem",
} satisfies CSSProperties;

const queryListButtonStyle = {
  ...menuItemStyle,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
} satisfies CSSProperties;

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
  justifyContent: "flex-start",
  flex: "1 1 0",
  minInlineSize: 0,
  maxInlineSize: "7rem",
  overflow: "hidden",
};

const chipLabelStyle = {
  display: "block",
  minInlineSize: 0,
  maxInlineSize: "100%",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const clearButtonStyle = {
  ...controlButtonStyle,
  flex: "0 0 auto",
  borderColor: "transparent",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
};
